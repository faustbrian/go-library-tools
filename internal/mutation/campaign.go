package mutation

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/faustbrian/go-library-tools/internal/evidence"
)

// CampaignPolicy contains the canonical module and package policy required by
// mutation execution. Service lifecycle remains owned by the caller.
type CampaignPolicy struct {
	Repository        string
	ModuleDirectory   string
	ModulePath        string
	GoVersion         string
	Packages          []string
	TestTags          []string
	BuildTags         []string
	RequiredServices  []string
	ServiceIdentities map[string]string
	OwnedModules      []OwnedModule
	Workers           int
}

// RuntimeIdentity contains only non-secret execution metadata.
type RuntimeIdentity struct {
	GoVersion  string `json:"GOVERSION"`
	GOOS       string `json:"GOOS"`
	GOARCH     string `json:"GOARCH"`
	CGOEnabled string `json:"CGO_ENABLED"`
}

// Campaign executes or reuses package-granular mutation evidence.
type Campaign struct {
	Root            string
	EvidenceRoot    string
	MutationRoot    string
	Workspace       string
	Policy          CampaignPolicy
	ZeroReviews     ZeroInventory
	Environment     map[string]string
	RuntimeIdentity RuntimeIdentity
	Process         Process
	Output          io.Writer
	Now             func() time.Time
	directoryFiles  campaignDirectoryFileSystem
}

type campaignDirectoryFileSystem interface {
	MkdirAll(string, os.FileMode) error
	Lstat(string) (os.FileInfo, error)
}

type operatingCampaignDirectories struct{}

func (operatingCampaignDirectories) MkdirAll(path string, mode os.FileMode) error {
	return os.MkdirAll(path, mode)
}
func (operatingCampaignDirectories) Lstat(path string) (os.FileInfo, error) {
	return os.Lstat(path)
}

// Run executes missing package campaigns and persists each result immediately.
func (campaign Campaign) Run(ctx context.Context) error {
	if err := campaign.validate(); err != nil {
		return err
	}
	if err := campaign.prepareDirectories(); err != nil {
		return err
	}
	output := campaign.Output
	if output == nil {
		output = io.Discard
	}
	packages := append([]string(nil), campaign.Policy.Packages...)
	sort.Strings(packages)
	state := campaignState{}
	for _, packageDirectory := range packages {
		if err := campaign.runPackage(ctx, output, packageDirectory, &state); err != nil {
			return err
		}
	}
	return nil
}

// Import migrates explicitly approved legacy checkpoints to current
// content-addressed evidence without depending on their Git revisions.
func (campaign Campaign) Import(ctx context.Context, checkpoints []Checkpoint, ledger MigrationLedger) error {
	if err := campaign.validate(); err != nil {
		return err
	}
	if len(checkpoints) == 0 {
		return fmt.Errorf("%w: mutation checkpoint selection is empty", ErrInvalid)
	}
	if err := ledger.validate(); err != nil {
		return err
	}
	if err := campaign.prepareDirectories(); err != nil {
		return err
	}
	allowed := make(map[string]struct{}, len(campaign.Policy.Packages))
	for _, packageDirectory := range campaign.Policy.Packages {
		allowed[packageDirectory] = struct{}{}
	}
	ordered := append([]Checkpoint(nil), checkpoints...)
	slices.SortFunc(ordered, func(left, right Checkpoint) int {
		return strings.Compare(left.Module+"\x00"+left.Package, right.Module+"\x00"+right.Package)
	})
	seen := make(map[string]struct{}, len(ordered))
	output := campaign.Output
	if output == nil {
		output = io.Discard
	}
	var changedInputs error
	for _, checkpoint := range ordered {
		identity := checkpoint.Module + "\x00" + checkpoint.Package
		if checkpoint.Module != campaign.Policy.ModuleDirectory {
			return fmt.Errorf("%w: checkpoint module %s does not match %s", ErrInvalid, checkpoint.Module, campaign.Policy.ModuleDirectory)
		}
		if _, exists := allowed[checkpoint.Package]; !exists {
			return fmt.Errorf("%w: checkpoint package %s is not mutation-enabled", ErrInvalid, checkpoint.Package)
		}
		if _, duplicate := seen[identity]; duplicate {
			return fmt.Errorf("%w: duplicate checkpoint identity %s %s", ErrInvalid, checkpoint.Module, checkpoint.Package)
		}
		seen[identity] = struct{}{}
		review, currentInput, legacyInput, err := campaign.packageInputs(ctx, checkpoint.Package)
		if err != nil {
			return err
		}
		if checkpoint.Mutants == 0 && review == nil {
			return fmt.Errorf("%w: zero-mutant package %s lacks an exact review", ErrInvalid, packageTarget(checkpoint.Package))
		}
		if err := ledger.approveTransition(
			checkpoint,
			strings.TrimPrefix(currentInput, "sha256:"),
			strings.TrimPrefix(legacyInput, "sha256:"),
			LegacyVerifierDigest(),
		); err != nil {
			if errors.Is(err, ErrInputChanged) {
				changedInputs = errors.Join(changedInputs, fmt.Errorf("%s %s: %w", checkpoint.Module, checkpoint.Package, err))
				_, _ = fmt.Fprintf(output, "[%s] %s skipped stale mutation checkpoint\n",
					campaign.Policy.ModuleDirectory, packageTarget(checkpoint.Package))
				continue
			}
			return fmt.Errorf("approve checkpoint %s %s: %w", checkpoint.Module, checkpoint.Package, err)
		}
		_, _, stored, err := StoreReport(campaign.MutationRoot, currentInput, checkpoint.Report)
		if err != nil {
			return err
		}
		if stored.Digest != checkpoint.ReportDigest || stored.Mutants != checkpoint.Mutants {
			return fmt.Errorf("%w: checkpoint report identity changed during import", ErrInvalid)
		}
		now := campaign.Now
		if now == nil {
			now = time.Now
		}
		record := evidence.Record{
			SchemaVersion: evidence.SchemaVersion, Repository: campaign.Policy.Repository,
			Module: campaign.Policy.ModuleDirectory, Package: checkpoint.Package, Gate: "mutation",
			InputDigest: currentInput, VerifierDigest: SemanticVerifierDigest(), Result: "passed",
			ReportDigest: stored.Digest, CompletedAt: now().UTC(),
			Environment: map[string]string{
				"GOVERSION": campaign.RuntimeIdentity.GoVersion, "GOOS": campaign.RuntimeIdentity.GOOS,
				"GOARCH": campaign.RuntimeIdentity.GOARCH, "CGO_ENABLED": campaign.RuntimeIdentity.CGOEnabled,
				"evidence_origin": "approved_legacy_checkpoint",
			},
		}
		if _, _, err := evidence.Store(campaign.EvidenceRoot, record); err != nil {
			return err
		}
		_, _ = fmt.Fprintf(output, "[%s] %s imported approved mutation evidence (%d mutants)\n",
			campaign.Policy.ModuleDirectory, packageTarget(checkpoint.Package), stored.Mutants)
	}
	return changedInputs
}

func (campaign Campaign) prepareDirectories() error {
	directoryFiles := campaign.directoryFiles
	if directoryFiles == nil {
		directoryFiles = operatingCampaignDirectories{}
	}
	for _, root := range []struct{ name, directory string }{
		{"evidence", campaign.EvidenceRoot},
		{"mutation", campaign.MutationRoot},
		{"workspace", campaign.Workspace},
	} {
		if err := ensureCampaignDirectory(directoryFiles, root.directory); err != nil {
			return fmt.Errorf("prepare %s root: %w", root.name, err)
		}
	}
	return nil
}

type campaignState struct {
	tool            Tool
	toolBuilt       bool
	coverageProfile string
	coverageElapsed string
}

func (campaign Campaign) runPackage(ctx context.Context, output io.Writer, packageDirectory string, state *campaignState) error {
	review, input, err := campaign.packageInput(ctx, packageDirectory)
	if err != nil {
		return err
	}
	reused, result, err := Reuse(campaign.EvidenceRoot, campaign.MutationRoot, campaign.Policy.Repository, campaign.Policy.ModuleDirectory, packageDirectory, input)
	if err != nil {
		return err
	}
	target := packageTarget(packageDirectory)
	if reused {
		_, _ = fmt.Fprintf(output, "[%s] %s reused content-identical mutation evidence (%d mutants)\n", campaign.Policy.ModuleDirectory, target, result.Mutants)
		return nil
	}
	if err := campaign.prepareExecution(ctx, state); err != nil {
		return err
	}
	reportPath := filepath.Join(campaign.Workspace, "report-"+packageSlug(packageDirectory)+".json")
	if err := removeStaleReport(reportPath); err != nil {
		return err
	}
	arguments, err := Arguments(target, reportPath, strings.Join(campaign.Policy.TestTags, ","), false, campaign.Policy.Workers)
	if err != nil {
		return err
	}
	environment := campaign.commandEnvironment()
	environment["GOLIB_GREMLINS_COVERAGE_PROFILE"] = state.coverageProfile
	environment["GOLIB_GREMLINS_COVERAGE_ELAPSED"] = state.coverageElapsed
	environment["GOCACHE"] = filepath.Join(campaign.Workspace, "mutation-cache", packageSlug(packageDirectory))
	if err := os.MkdirAll(environment["GOCACHE"], 0o700); err != nil {
		return fmt.Errorf("create package mutation cache: %w", err)
	}
	directory := filepath.Join(campaign.Root, filepath.FromSlash(campaign.Policy.ModuleDirectory))
	if err := campaign.Process(ctx, state.tool.Path, arguments, directory, environment, output, output); err != nil {
		return fmt.Errorf("mutation tool failed for %s %s: %w", campaign.Policy.ModuleDirectory, target, err)
	}
	report, err := os.ReadFile(reportPath)
	if errors.Is(err, os.ErrNotExist) {
		if review != nil {
			report = []byte("{\"files\":[]}\n")
			err = nil
		}
	}
	if err != nil {
		return fmt.Errorf("read mutation report for %s: %w", target, err)
	}
	validated, err := ValidateReport(bytes.NewReader(report))
	if err != nil {
		return err
	}
	if validated.Mutants == 0 && review == nil {
		return fmt.Errorf("%w: zero-mutant package %s lacks an exact review", ErrInvalid, target)
	}
	_, currentInput, err := campaign.packageInput(ctx, packageDirectory)
	if err != nil {
		return err
	}
	if currentInput != input {
		return fmt.Errorf("%w: mutation inputs changed while running %s", ErrInvalid, target)
	}
	_, _, stored, err := StoreReport(campaign.MutationRoot, input, report)
	if err != nil {
		return err
	}
	now := campaign.Now
	if now == nil {
		now = time.Now
	}
	recordEnvironment := map[string]string{
		"GOVERSION":   campaign.RuntimeIdentity.GoVersion,
		"GOOS":        campaign.RuntimeIdentity.GOOS,
		"GOARCH":      campaign.RuntimeIdentity.GOARCH,
		"CGO_ENABLED": campaign.RuntimeIdentity.CGOEnabled,
	}
	recordEnvironment["gremlins_binary_sha256"] = state.tool.Digest
	record := evidence.Record{
		SchemaVersion: evidence.SchemaVersion, Repository: campaign.Policy.Repository,
		Module: campaign.Policy.ModuleDirectory, Package: packageDirectory, Gate: "mutation",
		InputDigest: input, VerifierDigest: SemanticVerifierDigest(), Result: "passed",
		ReportDigest: stored.Digest, Environment: recordEnvironment, CompletedAt: now().UTC(),
	}
	if _, _, err := evidence.Store(campaign.EvidenceRoot, record); err != nil {
		return err
	}
	if stored.Mutants == 0 {
		_, _ = fmt.Fprintf(output, "[%s] %s has an exact zero-viable-mutant review\n", campaign.Policy.ModuleDirectory, target)
	} else {
		_, _ = fmt.Fprintf(output, "[%s] %s killed %d/%d viable mutants\n", campaign.Policy.ModuleDirectory, target, stored.Mutants, stored.Mutants)
	}
	return nil
}

func (campaign Campaign) packageInput(ctx context.Context, packageDirectory string) (*ZeroReview, string, error) {
	review, current, _, err := campaign.packageInputs(ctx, packageDirectory)
	return review, current, err
}

func (campaign Campaign) packageInputs(ctx context.Context, packageDirectory string) (*ZeroReview, string, string, error) {
	source, err := SourceDigest(campaign.Root, campaign.Policy.ModuleDirectory, packageDirectory)
	if err != nil {
		return nil, "", "", err
	}
	review, _ := campaign.ZeroReviews.Review(campaign.Policy.ModuleDirectory, packageDirectory, source, GremlinsVersion, LegacyVerifierDigest())
	policy := InputPolicy{
		ModuleDirectory: campaign.Policy.ModuleDirectory, PackageDirectory: packageDirectory,
		ModulePath: campaign.Policy.ModulePath, GoVersion: campaign.Policy.GoVersion,
		TestTags: campaign.Policy.TestTags, BuildTags: campaign.Policy.BuildTags,
		RequiredServices: campaign.Policy.RequiredServices, ServiceIdentities: campaign.Policy.ServiceIdentities,
		OwnedModules: campaign.Policy.OwnedModules,
	}
	var listing boundedBuffer
	listing.maximum = maximumListSize
	arguments := []string{"list", "-deps", "-test", "-json"}
	arguments = appendTagArgument(arguments, campaign.Policy.TestTags)
	arguments = append(arguments, packageTarget(packageDirectory))
	directory := filepath.Join(campaign.Root, filepath.FromSlash(campaign.Policy.ModuleDirectory))
	if err := campaign.Process(ctx, "go", arguments, directory, campaign.commandEnvironment(), &listing, io.Discard); err != nil {
		return nil, "", "", fmt.Errorf("list mutation input for %s: %w", packageDirectory, err)
	}
	current, err := InputDigest(campaign.Root, policy, strings.NewReader(listing.String()), review)
	if err != nil {
		return nil, "", "", err
	}
	legacy, err := legacyInputDigestV1(campaign.Root, policy, strings.NewReader(listing.String()), review)
	return review, current, legacy, err
}

func (campaign Campaign) prepareExecution(ctx context.Context, state *campaignState) error {
	if !state.toolBuilt {
		tool, err := BuildVerifier(ctx, filepath.Join(campaign.Workspace, "verifier"), campaign.Process)
		if err != nil {
			return err
		}
		state.tool = tool
		state.toolBuilt = true
	}
	if state.coverageProfile != "" {
		return nil
	}
	state.coverageProfile = filepath.Join(campaign.Workspace, "integration.coverage")
	arguments := []string{"test", "-count=1", "-timeout=20m", "-cover", "-coverpkg=./...", "-coverprofile=" + state.coverageProfile}
	arguments = appendTagArgument(arguments, campaign.Policy.TestTags)
	arguments = append(arguments, "./...")
	started := time.Now()
	directory := filepath.Join(campaign.Root, filepath.FromSlash(campaign.Policy.ModuleDirectory))
	if err := campaign.Process(ctx, "go", arguments, directory, campaign.commandEnvironment(), io.Discard, io.Discard); err != nil {
		return fmt.Errorf("build shared mutation coverage: %w", err)
	}
	elapsed := max(time.Since(started).Round(time.Second), time.Second)
	state.coverageElapsed = elapsed.String()
	return nil
}

func appendTagArgument(arguments, tags []string) []string {
	if len(tags) == 0 {
		return arguments
	}
	return append(arguments, "-tags="+strings.Join(tags, ","))
}

func removeStaleReport(path string) error {
	err := os.Remove(path)
	switch {
	case err == nil:
		return nil
	case errors.Is(err, os.ErrNotExist):
		return nil
	default:
		return fmt.Errorf("remove stale mutation report: %w", err)
	}
}

func (campaign Campaign) commandEnvironment() map[string]string {
	environment := cloneStrings(campaign.Environment)
	flags := strings.Fields(environment["GOFLAGS"])
	flags = slices.DeleteFunc(flags, func(flag string) bool {
		return strings.HasPrefix(flag, "-parallel=")
	})
	environment["GOFLAGS"] = strings.Join(append(flags, "-parallel=1"), " ")
	environment["GOWORK"] = "off"
	return environment
}

func (campaign Campaign) validate() error {
	policy := campaign.Policy
	identityParts := []string{ //nolint:prealloc // Fixed identity fields keep mutation semantics observable.
		policy.Repository, policy.ModulePath, policy.GoVersion,
		campaign.RuntimeIdentity.GoVersion, campaign.RuntimeIdentity.GOOS,
		campaign.RuntimeIdentity.GOARCH, campaign.RuntimeIdentity.CGOEnabled,
	}
	identityParts = append(identityParts, policy.TestTags...)
	identityParts = append(identityParts, policy.BuildTags...)
	identityText := strings.Join(identityParts, "")
	valid := [16]bool{
		filepath.IsAbs(campaign.Root), filepath.IsAbs(campaign.EvidenceRoot), filepath.IsAbs(campaign.MutationRoot), filepath.IsAbs(campaign.Workspace),
		campaign.Process != nil, policy.Repository != "", policy.ModulePath != "", policy.GoVersion != "", policy.Workers >= 1, policy.Workers <= 64,
		campaign.RuntimeIdentity.GoVersion != "", campaign.RuntimeIdentity.GOOS != "", campaign.RuntimeIdentity.GOARCH != "", campaign.RuntimeIdentity.CGOEnabled != "",
		!strings.ContainsAny(identityText, "\x00\r\n"), validRelative(policy.ModuleDirectory) && len(policy.Packages) > 0,
	}
	if valid != [16]bool{true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true} {
		return fmt.Errorf("%w: mutation campaign configuration is malformed", ErrInvalid)
	}
	for key := range campaign.Environment {
		if key == "" || strings.ContainsAny(key, "=\x00\r\n") {
			return fmt.Errorf("%w: mutation command environment key is malformed", ErrInvalid)
		}
	}
	if !isWithin(campaign.Root, campaign.EvidenceRoot) || !isWithin(campaign.Root, campaign.MutationRoot) {
		return fmt.Errorf("%w: mutation evidence roots must remain in the repository", ErrInvalid)
	}
	seen := make(map[string]struct{}, len(policy.Packages))
	for _, pkg := range policy.Packages {
		if !validRelative(pkg) {
			return fmt.Errorf("%w: mutation package path is malformed", ErrInvalid)
		}
		if _, exists := seen[pkg]; exists {
			return fmt.Errorf("%w: duplicate mutation package %s", ErrInvalid, pkg)
		}
		seen[pkg] = struct{}{}
	}
	return nil
}

func packageTarget(directory string) string {
	if directory == "." {
		return "."
	}
	return "./" + directory
}

func packageSlug(directory string) string {
	if directory == "." {
		return "root"
	}
	digest := sha256.Sum256([]byte(directory))
	return hex.EncodeToString(digest[:])
}

func cloneStrings(values map[string]string) map[string]string {
	result := make(map[string]string, len(values))
	maps.Copy(result, values)
	return result
}

func ensureCampaignDirectory(files campaignDirectoryFileSystem, path string) error {
	if err := files.MkdirAll(path, 0o700); err != nil {
		return err
	}
	info, err := files.Lstat(path)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%w: path is not a real directory", ErrInvalid)
	}
	return nil
}
