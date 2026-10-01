package mutation

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/faustbrian/go-library-tools/v2/internal/evidence"
)

func TestCampaignExecutesPersistsAndReusesPackageEvidence(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "source.go"), []byte("package example\n\nfunc Value() int { return 1 }\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, "adapter"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "adapter", "adapter.go"), []byte("package adapter\n\nfunc Value() int { return 2 }\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	source := t.TempDir()
	if err := os.WriteFile(filepath.Join(source, "go.mod"), []byte("module verifier\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	process := &campaignProcess{root: root, verifierSource: source}
	process.requireTags = true
	process.requireSerial = true
	var output bytes.Buffer
	campaign := Campaign{
		Root: root, EvidenceRoot: filepath.Join(root, ".verification"),
		MutationRoot: filepath.Join(root, ".verification", "mutation"), Workspace: filepath.Join(root, ".task"),
		Policy: CampaignPolicy{
			Repository: "example", ModuleDirectory: ".", ModulePath: "example", GoVersion: "1.27.0",
			Packages: []string{"adapter", "."}, TestTags: []string{"integration"},
			ServiceIdentities: map[string]string{}, Workers: 2,
		},
		Environment: map[string]string{
			"GOFLAGS": "-tags=integration -parallel=9", "SECRET_SERVICE_PASSWORD": "must-not-persist",
		},
		RuntimeIdentity: RuntimeIdentity{GoVersion: "go1.27.0", GOOS: "linux", GOARCH: "amd64", CGOEnabled: "0"},
		Process:         process.run,
		Output:          &output, Now: func() time.Time { return time.Unix(10, 0) },
	}
	if err := campaign.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if process.mutations != 2 || !strings.Contains(output.String(), "killed 1/1") {
		t.Fatalf("first campaign mutations/output = %d, %q", process.mutations, output.String())
	}
	if len(process.coverageElapsed) != 2 {
		t.Fatalf("mutation phase budgets = %v", process.coverageElapsed)
	}
	for _, budget := range process.coverageElapsed {
		parsed, err := time.ParseDuration(budget)
		if err != nil || parsed < time.Minute {
			t.Fatalf("mutation phase budget = %q, %v", budget, err)
		}
	}
	records, err := evidence.Inspect(root, ".verification", "example", []string{"."})
	if err != nil || len(records) != 2 {
		t.Fatalf("Inspect() = %#v, %v", records, err)
	}
	for _, record := range records {
		if _, leaked := record.Environment["SECRET_SERVICE_PASSWORD"]; leaked {
			t.Fatal("command secret persisted in mutation evidence")
		}
		if record.Environment["GOOS"] != "linux" || record.Environment["GOARCH"] != "amd64" {
			t.Fatalf("runtime identity = %#v", record.Environment)
		}
	}
	output.Reset()
	if err := campaign.Run(context.Background()); err != nil {
		t.Fatalf("Run(reuse) error = %v", err)
	}
	if process.mutations != 2 || strings.Count(output.String(), "reused content-identical") != 2 {
		t.Fatalf("reused campaign mutations/output = %d, %q", process.mutations, output.String())
	}
}

func TestCampaignAcceptsOnlyExactReviewedEfficacyExit(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "source.go"), []byte("package example\n\nfunc Value() int { return 1 }\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	verifier := t.TempDir()
	if err := os.WriteFile(filepath.Join(verifier, "go.mod"), []byte("module verifier\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	sourceDigest, err := SourceDigest(root, ".", ".")
	if err != nil {
		t.Fatal(err)
	}
	selected := `{"schema_version":1,"packages":[{"module_directory":".","package_directory":".","source_digest":"` + sourceDigest + `","gremlins_version":"v0.6.0","gremlins_verifier_sha256":"` + LegacyVerifierDigest() + `","mutations":[{"file_name":"source.go","type":"A","line":3,"column":1,"contract_domain":"Observable integer result for admitted nonnegative values","reason":"Both boundary forms return the same maximum integer at equality."}]}]}`
	inventory, err := ParseEquivalentInventory(strings.NewReader(selected))
	if err != nil {
		t.Fatal(err)
	}
	report := `{"files":[{"file_name":"source.go","mutations":[{"type":"A","status":"LIVED","line":3,"column":1}]}],"mutants_killed":0,"mutants_lived":1,"mutants_not_covered":0,"mutants_not_viable":0,"mutants_total":1,"mutations_coverage":100,"test_efficacy":0}`
	process := &campaignProcess{root: root, verifierSource: verifier, report: &report, fail: "efficacy"}
	var output bytes.Buffer
	campaign := Campaign{
		Root: root, EvidenceRoot: filepath.Join(root, ".verification"),
		MutationRoot: filepath.Join(root, ".verification", "mutation"), Workspace: filepath.Join(root, ".task"),
		Policy:            CampaignPolicy{Repository: "example", ModuleDirectory: ".", ModulePath: "example", GoVersion: "1.27.0", Packages: []string{"."}, ServiceIdentities: map[string]string{}, Workers: 1},
		EquivalentReviews: inventory,
		Environment:       map[string]string{}, RuntimeIdentity: RuntimeIdentity{GoVersion: "go1.27.0", GOOS: "linux", GOARCH: "amd64", CGOEnabled: "0"},
		Process: process.run, Output: &output,
	}
	if err := campaign.Run(context.Background()); err != nil {
		t.Fatalf("Run(reviewed) error = %v", err)
	}
	if !strings.Contains(output.String(), "0 killed, 1 reviewed equivalent") {
		t.Fatalf("Run(reviewed) output = %q", output.String())
	}
	_, input, err := campaign.packageInput(context.Background(), ".")
	if err != nil {
		t.Fatal(err)
	}
	parsedReview, err := inventory.Review(".", ".", sourceDigest, GremlinsVersion, LegacyVerifierDigest())
	if err != nil {
		t.Fatal(err)
	}
	stored, result, err := LoadReportWithReview(campaign.MutationRoot, input, parsedReview)
	if err != nil || string(stored) != report || result.Killed != 0 || result.Equivalent != 1 {
		t.Fatalf("stored native report = %q, %#v, %v", stored, result, err)
	}
	if _, _, err := LoadReport(campaign.MutationRoot, input); !errors.Is(err, ErrInvalid) {
		t.Fatalf("unreviewed report lookup error = %v", err)
	}
	output.Reset()
	if err := campaign.Run(context.Background()); err != nil || process.mutations != 1 || !strings.Contains(output.String(), "reused content-identical") {
		t.Fatalf("Run(reuse) = %v, mutations = %d, output = %q", err, process.mutations, output.String())
	}
	updated, err := ParseEquivalentInventory(strings.NewReader(strings.Replace(selected, "Both boundary forms", "The two boundary forms", 1)))
	if err != nil {
		t.Fatal(err)
	}
	campaign.EquivalentReviews = updated
	if err := campaign.Run(context.Background()); err != nil || process.mutations != 1 {
		t.Fatalf("Run(reason edit) = %v, mutations = %d", err, process.mutations)
	}
	campaign.EquivalentReviews = EquivalentInventory{}
	if err := campaign.Run(context.Background()); !errors.Is(err, ErrInvalid) {
		t.Fatalf("Run(withdrawn review) error = %v", err)
	}
	campaign.EquivalentReviews = inventory
	campaign.EvidenceRoot = filepath.Join(root, ".other-evidence")
	campaign.MutationRoot = filepath.Join(root, ".other-mutation")
	campaign.Workspace = filepath.Join(root, ".other-task")
	process.fail = "other-exit"
	output.Reset()
	if err := campaign.Run(context.Background()); err == nil || !strings.Contains(err.Error(), "mutation tool failed") {
		t.Fatalf("Run(non-efficacy exit) error = %v", err)
	}
	if !strings.Contains(output.String(), `"LIVED" "source.go" "A" 3:1`) {
		t.Fatalf("Run(non-efficacy exit) diagnostic = %q", output.String())
	}
	if _, err := os.Stat(filepath.Join(campaign.MutationRoot, "reports")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("failed native report was published: %v", err)
	}
}

func TestCampaignRejectsSourceChangeBeforeEvidenceReuse(t *testing.T) {
	root := t.TempDir()
	sourcePath := filepath.Join(root, "source.go")
	sourceA := []byte("package example\n\nfunc Value() int { return 1 }\n")
	sourceB := []byte("package example\n\nfunc Value() int { return 2 }\n")
	if err := os.WriteFile(sourcePath, sourceB, 0o600); err != nil {
		t.Fatal(err)
	}
	verifier := t.TempDir()
	if err := os.WriteFile(filepath.Join(verifier, "go.mod"), []byte("module verifier\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	makeReview := func() EquivalentInventory {
		t.Helper()
		digest, err := SourceDigest(root, ".", ".")
		if err != nil {
			t.Fatal(err)
		}
		return EquivalentInventory{SchemaVersion: 1, Packages: []EquivalentReview{{
			ModuleDirectory: ".", PackageDirectory: ".", SourceDigest: digest,
			GremlinsVersion: GremlinsVersion, GremlinsVerifierSHA256: LegacyVerifierDigest(),
			Mutations: []EquivalentMutation{{FileName: "source.go", Type: "A", Line: 3, Column: 1,
				ContractDomain: "Observable integer result for admitted nonnegative values",
				Reason:         "Both boundary forms return the same maximum integer at equality."}},
		}}}
	}
	report := `{"files":[{"file_name":"source.go","mutations":[{"type":"A","status":"LIVED","line":3,"column":1}]}],"mutants_killed":0,"mutants_lived":1,"mutants_not_covered":0,"mutants_not_viable":0,"mutants_total":1,"mutations_coverage":100,"test_efficacy":0}`
	process := &campaignProcess{root: root, verifierSource: verifier, report: &report, fail: "efficacy"}
	var output bytes.Buffer
	campaign := Campaign{
		Root: root, EvidenceRoot: filepath.Join(root, ".verification"),
		MutationRoot: filepath.Join(root, ".verification", "mutation"), Workspace: filepath.Join(root, ".task"),
		Policy: CampaignPolicy{Repository: "example", ModuleDirectory: ".", ModulePath: "example",
			GoVersion: "1.27.0", Packages: []string{"."}, ServiceIdentities: map[string]string{}, Workers: 1},
		EquivalentReviews: makeReview(), Environment: map[string]string{},
		RuntimeIdentity: RuntimeIdentity{GoVersion: "go1.27.0", GOOS: "linux", GOARCH: "amd64", CGOEnabled: "0"},
		Process:         process.run, Output: &output,
	}
	if err := campaign.Run(context.Background()); err != nil {
		t.Fatalf("Run(source B) error = %v", err)
	}
	if err := os.WriteFile(sourcePath, sourceA, 0o600); err != nil {
		t.Fatal(err)
	}
	campaign.EquivalentReviews = makeReview()
	process.afterList = func() error { return os.WriteFile(sourcePath, sourceB, 0o600) }
	output.Reset()
	if err := campaign.Run(context.Background()); !errors.Is(err, ErrInvalid) {
		t.Fatalf("Run(source A to B during go list) error = %v, output = %q", err, output.String())
	}
	if process.mutations != 1 || strings.Contains(output.String(), "reused content-identical") {
		t.Fatalf("stale review reused prior B report: mutations = %d, output = %q", process.mutations, output.String())
	}
}

func TestMutationPhaseTimeoutIncludesMeasuredColdCompileAndMinimum(t *testing.T) {
	tests := []struct {
		name     string
		baseline time.Duration
		want     time.Duration
	}{
		{name: "minimum", baseline: time.Second, want: time.Minute},
		{name: "rounds up", baseline: 45*time.Second + time.Nanosecond, want: 61 * time.Second},
		{name: "measured baseline", baseline: 2 * time.Minute, want: 2*time.Minute + 15*time.Second},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := mutationPhaseTimeout(test.baseline); got != test.want {
				t.Fatalf("mutationPhaseTimeout(%s) = %s, want %s", test.baseline, got, test.want)
			}
		})
	}
}

type campaignProcess struct {
	root            string
	verifierSource  string
	mutations       int
	listCalls       int
	fail            string
	report          *string
	skipReport      bool
	mutationOutput  string
	mutateSource    bool
	afterMutation   func() error
	afterList       func() error
	requireTags     bool
	requireSerial   bool
	coverageElapsed []string
}

type mutationExitCode int

func (code mutationExitCode) Error() string { return "mutation efficacy threshold" }
func (code mutationExitCode) ExitCode() int { return int(code) }

func (process *campaignProcess) run(_ context.Context, name string, args []string, _ string, environment map[string]string, stdout, _ io.Writer) error {
	packageCommand := name == "go" && len(args) > 0 && (args[0] == "list" || args[0] == "test") || strings.HasSuffix(name, "golib-gremlins")
	if process.requireSerial && packageCommand {
		parallelFlags := 0
		for flag := range strings.FieldsSeq(environment["GOFLAGS"]) {
			if strings.HasPrefix(flag, "-parallel=") {
				parallelFlags++
				if flag != "-parallel=1" {
					return errors.New("mutation package tests retain a conflicting parallel limit")
				}
			}
		}
		if parallelFlags != 1 {
			return errors.New("mutation package tests are not serialized")
		}
	}
	switch {
	case name == "go" && len(args) > 1 && args[0] == "list":
		if containsArgument(args, "-tags=") {
			return errors.New("empty list tags")
		}
		if process.requireTags && !containsArgument(args, "-tags=integration") {
			return errors.New("missing list tags")
		}
		process.listCalls++
		if process.fail == "list" || process.fail == "second-list" && process.listCalls > 1 {
			return errors.New("list failed")
		}
		directory := process.root
		importPath := "example"
		goFiles := []string{"source.go"}
		if args[len(args)-1] == "./adapter" {
			directory = filepath.Join(process.root, "adapter")
			importPath = "example/adapter"
			goFiles = []string{"adapter.go"}
		}
		listing := listedPackage{
			Dir: directory, ImportPath: importPath, GoFiles: goFiles,
			Module: &listedModule{Path: "example", Main: true, GoVersion: "1.27.0"},
		}
		if err := json.NewEncoder(stdout).Encode(listing); err != nil {
			return err
		}
		if process.afterList != nil {
			return process.afterList()
		}
		return nil
	case name == "go" && len(args) > 1 && args[0] == "mod":
		if process.fail == "download" {
			return errors.New("download failed")
		}
		_, err := io.WriteString(stdout, validDownloadMetadata(process.verifierSource))
		return err
	case name == "git":
		return nil
	case name == "go" && args[0] == "build":
		if process.fail == "build" {
			return errors.New("build failed")
		}
		return os.WriteFile(args[4], []byte("verifier"), 0o700)
	case name == "go" && args[0] == "test":
		if process.requireTags && !containsArgument(args, "-tags=integration") {
			return errors.New("missing test tags")
		}
		if process.fail == "coverage" {
			return errors.New("coverage failed")
		}
		for _, argument := range args {
			if argument == "-tags=" {
				return errors.New("empty tags argument")
			}
			if profile, found := strings.CutPrefix(argument, "-coverprofile="); found {
				return os.WriteFile(profile, []byte("mode: set\n"), 0o600)
			}
		}
	case strings.HasSuffix(name, "golib-gremlins"):
		process.mutations++
		process.coverageElapsed = append(process.coverageElapsed, environment["GOLIB_GREMLINS_COVERAGE_ELAPSED"])
		if process.fail == "mutation" {
			return errors.New("mutation failed")
		}
		if environment["GOLIB_GREMLINS_COVERAGE_PROFILE"] == "" || environment["GOCACHE"] == "" {
			return errors.New("missing isolated mutation environment")
		}
		if process.mutationOutput != "" {
			if _, err := io.WriteString(stdout, process.mutationOutput); err != nil {
				return err
			}
		}
		if !process.skipReport {
			value := `{"files":[{"file_name":"source.go","mutations":[{"type":"A","status":"KILLED","line":3,"column":1}]}]}`
			if process.report != nil {
				value = *process.report
			}
			for index, argument := range args {
				if argument == "--output" {
					if err := os.WriteFile(args[index+1], []byte(value), 0o600); err != nil {
						return err
					}
				}
			}
		}
		if process.mutateSource {
			if err := os.WriteFile(filepath.Join(process.root, "source.go"), []byte("package example\n\nfunc Value() int { return 9 }\n"), 0o600); err != nil {
				return err
			}
		}
		if process.afterMutation != nil {
			if err := process.afterMutation(); err != nil {
				return err
			}
		}
		if process.fail == "efficacy" {
			return mutationExitCode(10)
		}
		if process.fail == "other-exit" {
			return mutationExitCode(11)
		}
		return nil
	}
	return nil
}

func containsArgument(arguments []string, expected string) bool {
	return slices.Contains(arguments, expected)
}
