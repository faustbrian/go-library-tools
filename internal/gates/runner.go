// Package gates orchestrates repository checks from canonical module policy.
package gates

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/faustbrian/go-library-tools/v2/internal/config"
	"github.com/faustbrian/go-library-tools/v2/internal/coverage"
	"github.com/faustbrian/go-library-tools/v2/internal/docscheck"
	"github.com/faustbrian/go-library-tools/v2/internal/inventory"
	"github.com/faustbrian/go-library-tools/v2/internal/repositoryfile"
	"github.com/faustbrian/go-library-tools/v2/internal/services"
	"golang.org/x/mod/module"
)

const maximumMakefileSize = 4 << 20

const (
	maximumSecuritySourceFiles = 100_000
	maximumSecuritySourceSize  = 4 << 20
)

var (
	securityRuleList = regexp.MustCompile(`^G[0-9]{3}(?:\s*,\s*G[0-9]{3})*$`)
)

const (
	gitleaksPolicy = `title = "golib centrally owned secret scanning"
[extend]
useDefault = true

[[allowlists]]
description = "The immutable CI tooling checkout is verified separately."
paths = ['''^\.golib-tooling(?:/|$)''']

[[allowlists]]
description = "Pinned apidiff versions in exact compatibility rehearsal Makefiles are tool identities."
condition = "AND"
targetRules = ["generic-api-key"]
regexTarget = "secret"
regexes = ['''^v0\.0\.0-[0-9]{14}-[0-9a-f]{12}$''']
paths = [
  '''^internal/gates/api\.go$''',
  '''^rehearsals/go-(?:authorization|openapi)/verification/package\.mk$''',
]

[[allowlists]]
description = "Historical APIDIFF_VERSION pseudo-version in the retired legacy tool-version file."
condition = "AND"
targetRules = ["generic-api-key"]
regexTarget = "secret"
regexes = ['''^v0\.0\.0-[0-9]{14}-[0-9a-f]{12}$''']
paths = ['''^\.golib/versions\.env$''']

[[allowlists]]
description = "Exact synthetic Stripe token used by hostile inventory identity tests."
condition = "AND"
targetRules = ["stripe-access-token"]
regexTarget = "secret"
regexes = ['''^sk_test_0123456789abcdefghijklmnopqrstuv$''']
paths = ['''^internal/inventory/inventory_test\.go$''']

[[allowlists]]
description = "Exact public signing fixture in the shared HTTP-signature differential corpus."
condition = "AND"
targetRules = ["generic-api-key"]
regexTarget = "secret"
regexes = ['''^01234567(?:)89abcdef(?:)01234567(?:)89abcdef(?:)01234567(?:)89abcdef(?:)01234567(?:)89abcdef$''']
paths = ['''^differential/shared-corpus/corpus_test\.go$''']

[[allowlists]]
description = "Exact immutable Confluent decision digests recorded in the provider changelog."
condition = "AND"
targetRules = ["confluent-secret-key"]
regexTarget = "secret"
regexes = [
  '''^67b4c198(?:)5e70a7ae(?:)a45d754e(?:)14be8468(?:)83a37ccd(?:)073d76e5(?:)99d71900(?:)4af0ea37$''',
  '''^4c9ab0b7(?:)2db6bcd6(?:)a6f90cd8(?:)e638e7f2(?:)80708c95(?:)18128b59(?:)33a9a904(?:)ad072ff7$''',
  '''^c92530ae(?:)87091474(?:)8c82e238(?:)b72ff1d3(?:)c60c2ec0(?:)360150b0(?:)1863750e(?:)6dad0ac1$''',
]
paths = ['''^providers/confluent/CHANGELOG\.md$''']
`
	analysisSecurityPolicy = `version: 1
rules:
  security/no-unsafe:
    status: blocking
    promotion:
      version: 1.0.0
      evidence: ecosystem security policy prohibits unsafe, cgo, and go:linkname bypasses
`
)

const (
	golangCILintVersion = "v2.13.1"
	staticcheckVersion  = "v0.8.1"
	nilAwayVersion      = "v0.0.0-20260720194628-9fd1b8d7bac8"
	govulncheckVersion  = "v1.6.0"
	gosecVersion        = "v2.29.0"
	goAnalysisVersion   = "v1.0.0"
	gitleaksVersion     = "v8.30.1"
	goLicensesVersion   = "v2.0.1"
	cycloneDXVersion    = "v1.10.0"
)

// Command is one external process invocation without shell interpretation.
type Command struct {
	Name   string
	Args   []string
	Dir    string
	Env    map[string]string
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

const maximumSecurityProcessOutput = 4 << 20

const gitleaksIgnoreLogMarker = "found .gitleaksignore file"

var errRepositoryGitleaksIgnore = errors.New("security-policy: repository-owned .gitleaksignore is not permitted for security-enabled checks")

type boundedProcessOutput struct {
	mutex          sync.Mutex
	limit          int
	written        int
	overflow       bool
	onLimit        func()
	forbidden      []byte
	forbiddenTail  []byte
	forbiddenFound bool
}

func (output *boundedProcessOutput) Write(value []byte) (int, error) {
	output.mutex.Lock()
	if len(output.forbidden) > 0 {
		window := make([]byte, 0, len(output.forbiddenTail)+len(value))
		window = append(window, output.forbiddenTail...)
		window = append(window, value...)
		if bytes.Contains(window, output.forbidden) {
			output.forbiddenFound = true
		}
		keep := min(len(output.forbidden)-1, len(window))
		output.forbiddenTail = append(output.forbiddenTail[:0], window[len(window)-keep:]...)
	}
	if len(value) > output.limit-output.written {
		output.written = output.limit
		first := !output.overflow
		output.overflow = true
		onLimit := output.onLimit
		output.mutex.Unlock()
		if first && onLimit != nil {
			onLimit()
		}
		return len(value), nil
	}
	output.written += len(value)
	output.mutex.Unlock()
	return len(value), nil
}

func (output *boundedProcessOutput) setOverflowCallback(callback func()) {
	output.mutex.Lock()
	output.onLimit = callback
	alreadyOverflowed := output.overflow
	output.mutex.Unlock()
	if alreadyOverflowed && callback != nil {
		callback()
	}
}

func (output *boundedProcessOutput) didOverflow() bool {
	output.mutex.Lock()
	defer output.mutex.Unlock()
	return output.overflow
}

func (output *boundedProcessOutput) foundForbiddenOutput() bool {
	output.mutex.Lock()
	defer output.mutex.Unlock()
	return output.forbiddenFound
}

// Executor runs one external command.
type Executor interface {
	Run(context.Context, Command) error
}

type taskWorkspace interface {
	TemporaryDirectory() string
}

// Runner executes gates for modules in one validated repository.
type Runner struct {
	Root              string
	Catalog           inventory.Inventory
	Policy            config.Config
	Executor          Executor
	Output            io.Writer
	coverageFiles     coverageFileSystem
	secretConfigFiles secretConfigFileSystem
	apiFiles          apiFileSystem
	apiReadBaseline   func(string, string, int64) ([]byte, error)
	releaseFiles      releaseFileSystem
	releaseArchive    func(io.Writer, module.Version, string) error
	mutationFiles     mutationFileSystem
	mutationCampaign  mutationCampaignRunner
	mutationImport    mutationImportRunner
	startServices     serviceStarter
	serviceHTTPProbe  services.HTTPProbe
	serviceIdentities map[string]string
	// DocumentationSpelling is an isolated test boundary. Production callers
	// leave it nil and use the pinned task-owned implementation.
	DocumentationSpelling func(context.Context, string) error
	// DocumentationLinks is an isolated test boundary. Production callers
	// leave it nil and use the checksum-pinned task-owned implementation.
	DocumentationLinks   func(context.Context, string) error
	documentationRelease func(string, string) (docscheck.LycheeRelease, error)
	documentationExtract func(string, docscheck.LycheeRelease) ([]byte, error)
}

type namedWriteCloser interface {
	io.WriteCloser
	Name() string
}

type coverageFileSystem interface {
	CreateTemp(string) (namedWriteCloser, error)
	Open(string) (io.ReadCloser, error)
	Remove(string) error
}

type operatingCoverageFiles struct{}

func (operatingCoverageFiles) CreateTemp(directory string) (namedWriteCloser, error) {
	return os.CreateTemp(directory, "golib-coverage-*.out")
}

func (operatingCoverageFiles) Open(path string) (io.ReadCloser, error) {
	// #nosec G304 -- callers open the exact task-owned coverage profile path created by this gate
	return os.Open(path)
}

func (operatingCoverageFiles) Remove(path string) error {
	return os.Remove(path)
}

type secretConfigFileSystem interface {
	CreateTemp(string, string) (namedWriteCloser, error)
	Remove(string) error
}

type operatingSecretConfigFiles struct{}

func (operatingSecretConfigFiles) CreateTemp(directory, pattern string) (namedWriteCloser, error) {
	return os.CreateTemp(directory, pattern)
}

func (operatingSecretConfigFiles) Remove(path string) error {
	return os.Remove(path)
}

// Check runs the standard contract for each explicitly selected module.
func (runner Runner) Check(ctx context.Context, selection []string) error {
	modules, err := runner.selectModules(selection)
	if err != nil {
		return err
	}
	output := runner.Output
	if output == nil {
		output = io.Discard
	}
	if err := runner.preflightSecurityPolicies(output, modules); err != nil {
		return err
	}
	for _, module := range modules {
		if err := runner.withModuleServices(ctx, module, func(scoped Runner) error {
			return scoped.checkModule(ctx, output, module)
		}); err != nil {
			return err
		}
	}
	return nil
}

// Local runs the bounded repository-owned checks used for ordinary pull
// requests. Expensive evidence gates remain explicit Check operations selected
// for a material risk or release milestone.
func (runner Runner) Local(ctx context.Context, selection []string) error {
	modules, err := runner.selectModules(selection)
	if err != nil {
		return err
	}
	output := runner.Output
	if output == nil {
		output = io.Discard
	}
	if err := runner.preflightSecurityPolicies(output, modules); err != nil {
		return err
	}
	for _, module := range modules {
		if err := runner.withModuleServices(ctx, module, func(scoped Runner) error {
			return scoped.checkModuleLocal(ctx, output, module)
		}); err != nil {
			return err
		}
	}
	return nil
}

// Coverage runs only exact production-package coverage for selected modules.
func (runner Runner) Coverage(ctx context.Context, selection []string) error {
	modules, err := runner.selectModules(selection)
	if err != nil {
		return err
	}
	output := runner.Output
	if output == nil {
		output = io.Discard
	}
	for _, module := range modules {
		if !module.Gates["coverage"] {
			_, _ = fmt.Fprintf(output, "[%s] coverage: not applicable\n", module.Directory)
			continue
		}
		if err := runner.withModuleServices(ctx, module, func(scoped Runner) error {
			directory := filepath.Join(scoped.Root, module.Directory)
			return announce(output, module.Directory, "coverage", func() error {
				return scoped.runCoverage(ctx, output, directory, module)
			})
		}); err != nil {
			return err
		}
	}
	return nil
}

// Docs validates Markdown navigation and any additional typed documentation
// operation for selected modules.
func (runner Runner) Docs(ctx context.Context, selection []string) error {
	modules, err := runner.selectModules(selection)
	if err != nil {
		return err
	}
	output := runner.Output
	if output == nil {
		output = io.Discard
	}
	for _, module := range modules {
		if !module.Gates["documentation"] {
			_, _ = fmt.Fprintf(output, "[%s] docs: not applicable\n", module.Directory)
			continue
		}
		directory := filepath.Join(runner.Root, module.Directory)
		if err := announce(output, module.Directory, "docs", func() error {
			return runner.checkDocumentation(ctx, directory, module)
		}); err != nil {
			return err
		}
	}
	return nil
}

func (runner Runner) selectModules(selection []string) ([]inventory.Module, error) {
	available := make(map[string]inventory.Module, len(runner.Catalog.Modules))
	for _, module := range runner.Catalog.Modules {
		available[module.Directory] = module
	}
	unique := make(map[string]struct{}, len(selection))
	for _, directory := range selection {
		if _, ok := available[directory]; !ok {
			return nil, fmt.Errorf("unknown module: %s", directory)
		}
		unique[directory] = struct{}{}
	}
	directories := make([]string, 0, len(unique))
	for directory := range unique {
		directories = append(directories, directory)
	}
	sort.Strings(directories)
	modules := make([]inventory.Module, 0, len(directories))
	for _, directory := range directories {
		modules = append(modules, available[directory])
	}
	return modules, nil
}

func (runner Runner) checkModule(ctx context.Context, output io.Writer, module inventory.Module) error {
	directory := filepath.Join(runner.Root, module.Directory)
	if err := announce(output, module.Directory, "format-check", func() error {
		return runner.checkFormatting(ctx, directory)
	}); err != nil {
		return err
	}
	if err := runner.command(ctx, output, module.Directory, "tidy-check", directory, "mod", "tidy", "-diff"); err != nil {
		return err
	}
	if err := announce(output, module.Directory, "safety", func() error {
		return checkSafety(directory)
	}); err != nil {
		return err
	}
	if module.Gates["lint"] {
		if err := runner.command(ctx, output, module.Directory, "vet", directory, "vet", "./..."); err != nil {
			return err
		}
	}
	if module.Gates["tests"] {
		args := testArguments(module.TestTags, false)
		if err := runner.command(ctx, output, module.Directory, "test", directory, args...); err != nil {
			return err
		}
		if operation, exists := runner.operation(module.Directory, "test"); exists {
			if err := runner.runOperation(ctx, directory, module, operation); err != nil {
				return err
			}
		}
	}
	if module.Gates["race"] {
		args := testArguments(module.TestTags, true)
		if err := runner.command(ctx, output, module.Directory, "race", directory, args...); err != nil {
			return err
		}
	}
	if module.Gates["coverage"] {
		if err := announce(output, module.Directory, "coverage", func() error {
			return runner.runCoverage(ctx, output, directory, module)
		}); err != nil {
			return err
		}
	}
	if module.Gates["mutation"] {
		if err := runner.verifyMutation(ctx, output, module); err != nil {
			return err
		}
	}
	if module.Gates["lint"] {
		if err := runner.goTool(ctx, output, module.Directory, "lint", directory,
			"github.com/golangci/golangci-lint/v2/cmd/golangci-lint@"+golangCILintVersion,
			"run", "--allow-parallel-runners", "--timeout=10m", "./..."); err != nil {
			return err
		}
		if err := runner.goTool(ctx, output, module.Directory, "staticcheck", directory,
			"honnef.co/go/tools/cmd/staticcheck@"+staticcheckVersion, "./..."); err != nil {
			return err
		}
	}
	if module.Gates["security"] {
		if err := runner.runSecurity(ctx, output, directory, module); err != nil {
			return err
		}
	}
	if operation, exists := runner.operation(module.Directory, "fuzz"); exists {
		if err := announce(output, module.Directory, "fuzz", func() error {
			return runner.runOperation(ctx, directory, module, operation)
		}); err != nil {
			return err
		}
	}
	if module.Gates["documentation"] {
		if err := announce(output, module.Directory, "docs", func() error {
			return runner.checkDocumentation(ctx, directory, module)
		}); err != nil {
			return err
		}
	}
	if module.Gates["api_compatibility"] {
		if operation, exists := runner.operation(module.Directory, "api"); exists {
			if err := announce(output, module.Directory, "api", func() error {
				return runner.runOperation(ctx, directory, module, operation)
			}); err != nil {
				return err
			}
		} else if err := announce(output, module.Directory, "api", func() error {
			return runner.apiModule(ctx, output, module, false)
		}); err != nil {
			return err
		}
	}
	if module.Gates["lint"] {
		_, _ = fmt.Fprintf(output, "[%s] nilaway\n", module.Directory)
		err := runner.Executor.Run(ctx, Command{
			Name: "go", Dir: directory, Env: map[string]string{"GOWORK": "off"},
			Args: []string{"run", "go.uber.org/nilaway/cmd/nilaway@" + nilAwayVersion,
				"-include-pkgs=" + module.ModulePath, "./..."},
		})
		if err != nil {
			_, _ = fmt.Fprintf(output, "[%s] NilAway advisory: %v\n", module.Directory, err)
		}
	}
	for _, gate := range []string{"conformance", "interoperability", "benchmark"} {
		operation, exists := runner.operation(module.Directory, gate)
		if !exists {
			continue
		}
		if err := announce(output, module.Directory, gate, func() error {
			return runner.runOperation(ctx, directory, module, operation)
		}); err != nil {
			return err
		}
	}
	return nil
}

func (runner Runner) checkModuleLocal(ctx context.Context, output io.Writer, module inventory.Module) error {
	directory := filepath.Join(runner.Root, module.Directory)
	if err := announce(output, module.Directory, "format-check", func() error {
		return runner.checkFormatting(ctx, directory)
	}); err != nil {
		return err
	}
	if err := runner.command(ctx, output, module.Directory, "tidy-check", directory, "mod", "tidy", "-diff"); err != nil {
		return err
	}
	if err := announce(output, module.Directory, "safety", func() error {
		return checkSafety(directory)
	}); err != nil {
		return err
	}
	if module.Gates["lint"] {
		if err := runner.command(ctx, output, module.Directory, "vet", directory, "vet", "./..."); err != nil {
			return err
		}
	}
	if module.Gates["tests"] {
		if err := runner.command(ctx, output, module.Directory, "test", directory, testArguments(module.TestTags, false)...); err != nil {
			return err
		}
	}
	if module.Gates["lint"] {
		if err := runner.goTool(ctx, output, module.Directory, "lint", directory,
			"github.com/golangci/golangci-lint/v2/cmd/golangci-lint@"+golangCILintVersion,
			"run", "--allow-parallel-runners", "--timeout=10m", "./..."); err != nil {
			return err
		}
		if err := runner.goTool(ctx, output, module.Directory, "staticcheck", directory,
			"honnef.co/go/tools/cmd/staticcheck@"+staticcheckVersion, "./..."); err != nil {
			return err
		}
	}
	if module.Gates["security"] {
		if err := runner.runSecurity(ctx, output, directory, module); err != nil {
			return err
		}
	}
	if module.Gates["documentation"] {
		if err := announce(output, module.Directory, "docs-local", func() error {
			return docscheck.CheckWithin(runner.Root, directory)
		}); err != nil {
			return err
		}
	}
	if module.Gates["api_compatibility"] {
		if err := announce(output, module.Directory, "api", func() error {
			return runner.apiModule(ctx, output, module, false)
		}); err != nil {
			return err
		}
	}
	return nil
}

func (runner Runner) createGitleaksConfig() (string, func() error, error) {
	return runner.createOwnedPolicy("gitleaks", "gitleaks-config-*.toml", gitleaksPolicy)
}

func (runner Runner) createAnalysisConfig() (string, func() error, error) {
	return runner.createOwnedPolicy("analysis", "analysis-security-*.yaml", analysisSecurityPolicy)
}

func (runner Runner) createOwnedPolicy(name, pattern, policy string) (string, func() error, error) {
	workspace, ok := runner.Executor.(taskWorkspace)
	if !ok || !filepath.IsAbs(workspace.TemporaryDirectory()) {
		return "", nil, errors.New("security policy requires an absolute task-owned temporary directory")
	}
	files := runner.secretConfigFiles
	if files == nil {
		files = operatingSecretConfigFiles{}
	}
	temporary, err := files.CreateTemp(workspace.TemporaryDirectory(), pattern)
	if err != nil {
		return "", nil, fmt.Errorf("create temporary %s config: %w", name, err)
	}
	path := temporary.Name()
	cleanup := func() error { return files.Remove(path) }
	if _, err := io.WriteString(temporary, policy); err != nil {
		return "", nil, errors.Join(fmt.Errorf("write temporary %s config: %w", name, err), temporary.Close(), cleanup())
	}
	if err := temporary.Close(); err != nil {
		return "", nil, errors.Join(fmt.Errorf("close temporary %s config: %w", name, err), cleanup())
	}
	return path, cleanup, nil
}

func (runner Runner) checkSecurity(ctx context.Context, output io.Writer, directory string, module inventory.Module) error {
	if err := runner.checkSecurityPolicy(output, directory, module.Directory); err != nil {
		return err
	}
	return runner.runSecurity(ctx, output, directory, module)
}

func (runner Runner) checkSecurityPolicy(output io.Writer, directory, module string) error {
	return announce(output, module, "security-suppressions", func() error {
		return checkSecuritySuppressions(directory)
	})
}

func (runner Runner) preflightSecurityPolicies(output io.Writer, modules []inventory.Module) error {
	securityEnabled := slices.ContainsFunc(modules, func(module inventory.Module) bool {
		return module.Gates["security"]
	})
	if securityEnabled {
		if err := rejectRepositoryGitleaksIgnore(runner.Root); err != nil {
			return err
		}
	}
	for _, module := range modules {
		if !module.Gates["security"] {
			continue
		}
		directory := filepath.Join(runner.Root, module.Directory)
		if err := runner.checkSecurityPolicy(output, directory, module.Directory); err != nil {
			return err
		}
	}
	return nil
}

func (runner Runner) runSecurity(ctx context.Context, output io.Writer, directory string, module inventory.Module) error {
	if err := runner.securityTool(ctx, output, module.Directory, "vulnerability", directory,
		"golang.org/x/vuln/cmd/govulncheck@"+govulncheckVersion, "./..."); err != nil {
		return err
	}
	if err := runner.securityTool(ctx, output, module.Directory, "gosec", directory,
		"github.com/securego/gosec/v2/cmd/gosec@"+gosecVersion,
		"-nosec-require-rules", "-nosec-require-justification", "./..."); err != nil {
		return err
	}
	analysisPath, cleanupAnalysis, err := runner.createAnalysisConfig()
	if err != nil {
		return err
	}
	if err := runner.securityTool(ctx, output, module.Directory, "owned-security-analysis", directory,
		"github.com/faustbrian/go-analysis/cmd/golib-analysis@"+goAnalysisVersion,
		"check", "-config", analysisPath, "-root", directory, "./..."); err != nil {
		return errors.Join(err, cleanupAnalysis())
	}
	if err := cleanupAnalysis(); err != nil {
		return fmt.Errorf("remove temporary analysis config: %w", err)
	}
	configPath, cleanupSecrets, err := runner.createGitleaksConfig()
	if err != nil {
		return err
	}
	if err := rejectRepositoryGitleaksIgnore(runner.Root); err != nil {
		return errors.Join(err, cleanupSecrets())
	}
	if err := runner.gitleaksTool(ctx, output, module.Directory, "secrets-history", runner.Root,
		"github.com/zricethezav/gitleaks/v8@"+gitleaksVersion,
		"git", ".", "--config", configPath, "--log-opts=--all", "--ignore-gitleaks-allow", "--log-level", "debug", "--no-banner", "--redact"); err != nil {
		return errors.Join(err, cleanupSecrets())
	}
	if err := rejectRepositoryGitleaksIgnore(runner.Root); err != nil {
		return errors.Join(err, cleanupSecrets())
	}
	if err := runner.gitleaksTool(ctx, output, module.Directory, "secrets-current-tree", runner.Root,
		"github.com/zricethezav/gitleaks/v8@"+gitleaksVersion,
		"dir", ".", "--config", configPath, "--ignore-gitleaks-allow", "--log-level", "debug", "--no-banner", "--redact"); err != nil {
		return errors.Join(err, cleanupSecrets())
	}
	if err := rejectRepositoryGitleaksIgnore(runner.Root); err != nil {
		return errors.Join(err, cleanupSecrets())
	}
	if err := cleanupSecrets(); err != nil {
		return fmt.Errorf("remove temporary gitleaks config: %w", err)
	}
	licenseOwner := module.ModulePath
	if repository := strings.TrimSuffix(runner.Catalog.Repository, "/"); repository != "" &&
		(module.ModulePath == repository || strings.HasPrefix(module.ModulePath, repository+"/")) {
		licenseOwner = repository
	}
	if err := runner.securityTool(ctx, output, module.Directory, "licenses", directory,
		"github.com/google/go-licenses/v2@"+goLicensesVersion,
		"check", "./...", "--ignore", licenseOwner); err != nil {
		return err
	}
	return announce(output, module.Directory, "SBOM", func() error { return runner.runSBOM(ctx, directory) })
}

func rejectRepositoryGitleaksIgnore(root string) error {
	info, err := os.Lstat(filepath.Join(root, ".gitleaksignore"))
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return errors.New("inspect repository-owned .gitleaksignore")
	}
	if info.IsDir() {
		return nil
	}
	return errRepositoryGitleaksIgnore
}

func checkSecuritySuppressions(root string) error {
	sourceRoot, err := os.OpenRoot(root)
	if err != nil {
		return err
	}
	defer sourceRoot.Close()
	files := 0
	return filepath.WalkDir(root, func(filePath string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if filePath != root && (entry.Name() == ".git" || entry.Name() == ".golib-tooling" || entry.Name() == "vendor") {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(entry.Name()) != ".go" {
			return nil
		}
		files++
		if files > maximumSecuritySourceFiles {
			return errors.New("security suppression source file limit exceeded")
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if info.Size() > maximumSecuritySourceSize {
			return fmt.Errorf("security suppression source exceeds size limit: %s", entry.Name())
		}
		relative, err := filepath.Rel(root, filePath)
		if err != nil {
			return err
		}
		file, err := sourceRoot.Open(relative)
		if err != nil {
			return err
		}
		content, readErr := io.ReadAll(io.LimitReader(file, maximumSecuritySourceSize+1))
		closeErr := file.Close()
		if readErr != nil || closeErr != nil {
			return errors.Join(readErr, closeErr)
		}
		if len(content) > maximumSecuritySourceSize {
			return fmt.Errorf("security suppression source exceeds size limit: %s", entry.Name())
		}
		fileSet := token.NewFileSet()
		parsed, err := parser.ParseFile(fileSet, filePath, content, parser.ParseComments)
		if err != nil {
			return err
		}
		for _, group := range parsed.Comments {
			for _, suppression := range nativeNosecGroupDirectives(group, fileSet) {
				directive, lineNumber := suppression.arguments, suppression.line
				if err := validateNativeSecurityDirective(directive); err != nil {
					return fmt.Errorf("%s:%d: %w", filepath.ToSlash(filePath), lineNumber, err)
				}
			}
			for _, comment := range group.List {
				lineNumber := fileSet.Position(comment.Slash).Line
				if directive, found := gosecDisableDirective(comment.Text); found {
					if err := validateNativeSecurityDirective(directive); err != nil {
						return fmt.Errorf("%s:%d: %w", filepath.ToSlash(filePath), lineNumber, err)
					}
				}
				if reason, found := nolintGosecDirective(comment.Text); found && strings.TrimSpace(reason) == "" {
					return fmt.Errorf("%s:%d: nolint:gosec requires an inline reason", filepath.ToSlash(filePath), lineNumber)
				}
			}
		}
		return nil
	})
}

func nativeSecurityDirective(line string, allowGosecDisable bool) (string, bool) {
	body := strings.TrimSpace(line)
	if allowGosecDisable {
		if directive, found := gosecDisableDirective(body); found {
			return directive, true
		}
	}
	if trimmed, found := strings.CutPrefix(body, "//"); found {
		body = strings.TrimSpace(trimmed)
	}
	const nosec = "#nosec"
	if trimmed, found := strings.CutPrefix(body, nosec); found {
		return strings.TrimSpace(trimmed), true
	}
	return "", false
}

type nativeSecuritySuppression struct {
	arguments string
	line      int
}

func nativeNosecGroupDirectives(group *ast.CommentGroup, fileSet *token.FileSet) []nativeSecuritySuppression {
	var suppressions []nativeSecuritySuppression
	for _, comment := range group.List {
		body := comment.Text
		baseLine := fileSet.Position(comment.Slash).Line
		if strings.HasPrefix(body, "//") {
			if directive, found := nativeSecurityDirective(body, false); found {
				suppressions = append(suppressions, nativeSecuritySuppression{arguments: directive, line: baseLine})
			}
			continue
		}
		if !strings.HasPrefix(body, "/*") {
			continue
		}
		body = strings.TrimPrefix(body, "/*")
		body = strings.TrimSuffix(body, "*/")
		for offset, line := range strings.Split(body, "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "//") {
				continue
			}
			if directive, found := nativeSecurityDirective(line, false); found {
				suppressions = append(suppressions, nativeSecuritySuppression{arguments: directive, line: baseLine + offset})
			}
		}
	}
	return suppressions
}

func nativeNosecGroupDirective(group *ast.CommentGroup, fileSet *token.FileSet) (string, int, bool) {
	suppressions := nativeNosecGroupDirectives(group, fileSet)
	if len(suppressions) == 0 {
		return "", 0, false
	}
	return suppressions[0].arguments, suppressions[0].line, true
}

func gosecDisableDirective(comment string) (string, bool) {
	body := strings.TrimSpace(comment)
	const directive = "//gosec:disable"
	if body == directive || strings.HasPrefix(body, directive+" ") {
		return strings.TrimSpace(strings.TrimPrefix(body, directive)), true
	}
	return "", false
}

func nolintGosecDirective(comment string) (string, bool) {
	body := strings.TrimLeft(comment, "/ ")
	if !strings.HasPrefix(body, "nolint:") {
		return "", false
	}
	directive, reason, _ := strings.Cut(body, "//")
	for linter := range strings.SplitSeq(strings.TrimPrefix(directive, "nolint:"), ",") {
		if strings.EqualFold(strings.TrimSpace(linter), "gosec") {
			return reason, true
		}
	}
	return "", false
}

func validateNativeSecurityDirective(arguments string) error {
	rules, reason, found := strings.Cut(arguments, "--")
	rules = strings.TrimSpace(rules)
	if !securityRuleList.MatchString(rules) {
		fields := strings.Fields(arguments)
		if !found && len(fields) > 0 && securityRuleList.MatchString(fields[0]) {
			return errors.New("security suppression requires a reason after --")
		}
		return errors.New("security suppression requires exact rule IDs")
	}
	if !found || strings.TrimSpace(strings.TrimLeft(reason, "-")) == "" {
		return errors.New("security suppression requires a reason after --")
	}
	return nil
}

func (runner Runner) goTool(ctx context.Context, output io.Writer, module, gate, directory, tool string, args ...string) error {
	arguments := append([]string{"run", tool}, args...)
	return runner.command(ctx, output, module, gate, directory, arguments...)
}

func (runner Runner) securityTool(ctx context.Context, output io.Writer, module, gate, directory, tool string, args ...string) error {
	return runner.securityToolWithForbiddenOutput(ctx, output, module, gate, directory, tool, nil, args...)
}

func (runner Runner) gitleaksTool(ctx context.Context, output io.Writer, module, gate, directory, tool string, args ...string) error {
	return runner.securityToolWithForbiddenOutput(
		ctx, output, module, gate, directory, tool, []byte(gitleaksIgnoreLogMarker), args...,
	)
}

func (runner Runner) securityToolWithForbiddenOutput(
	ctx context.Context,
	output io.Writer,
	module, gate, directory, tool string,
	forbidden []byte,
	args ...string,
) error {
	arguments := append([]string{"run", tool}, args...)
	return announce(output, module, gate, func() error {
		stdout := &boundedProcessOutput{limit: maximumSecurityProcessOutput, forbidden: forbidden}
		stderr := &boundedProcessOutput{limit: maximumSecurityProcessOutput, forbidden: forbidden}
		err := runner.Executor.Run(ctx, Command{
			Name: "go", Args: arguments, Dir: directory, Env: map[string]string{"GOWORK": "off"},
			Stdout: stdout, Stderr: stderr,
		})
		var overflow error
		if stdout.didOverflow() || stderr.didOverflow() {
			overflow = fmt.Errorf("security scanner output exceeded %d bytes", maximumSecurityProcessOutput)
		}
		var policyViolation error
		if stdout.foundForbiddenOutput() || stderr.foundForbiddenOutput() {
			policyViolation = errRepositoryGitleaksIgnore
		}
		if err != nil || overflow != nil || policyViolation != nil {
			return fmt.Errorf("%s %s: %w", module, gate, errors.Join(policyViolation, overflow, err))
		}
		return nil
	})
}

func (runner Runner) runCoverage(ctx context.Context, output io.Writer, directory string, module inventory.Module) error {
	targets := make([]string, 0, len(module.Packages))
	for _, packagePolicy := range module.Packages {
		if packagePolicy.CoverageRequired {
			targets = append(targets, packagePolicy.ImportPath)
		}
	}
	slices.Sort(targets)
	if len(targets) == 0 {
		return errors.New("coverage gate has no coverage-required packages")
	}
	files := runner.coverageFiles
	if files == nil {
		files = operatingCoverageFiles{}
	}
	temporaryDirectory := ""
	if workspace, ok := runner.Executor.(taskWorkspace); ok {
		temporaryDirectory = workspace.TemporaryDirectory()
	}
	profile, err := files.CreateTemp(temporaryDirectory)
	if err != nil {
		return fmt.Errorf("create coverage profile: %w", err)
	}
	profilePath := profile.Name()
	if err := profile.Close(); err != nil {
		_ = files.Remove(profilePath)
		return fmt.Errorf("close coverage profile: %w", err)
	}
	defer func() { _ = files.Remove(profilePath) }()
	args := []string{"test"}
	if len(module.TestTags) > 0 {
		args = append(args, "-tags="+strings.Join(module.TestTags, ","))
	}
	args = append(args, "./...", "-count=1", "-timeout=20m", "-covermode=atomic", "-coverpkg="+strings.Join(targets, ","), "-coverprofile="+profilePath)
	if err := runner.Executor.Run(ctx, Command{Name: "go", Args: args, Dir: directory, Env: map[string]string{"GOWORK": "off"}}); err != nil {
		return err
	}
	opened, err := files.Open(profilePath)
	if err != nil {
		return fmt.Errorf("open coverage profile: %w", err)
	}
	defer opened.Close()
	report, err := coverage.Verify(opened, targets)
	if err != nil {
		return err
	}
	_, _ = io.WriteString(output, report)
	_, _ = io.WriteString(output, "all production packages have exact 100% statement coverage\n")
	return nil
}

func (runner Runner) operation(module, gate string) (config.Operation, bool) {
	for _, operation := range runner.Policy.Operations {
		if operation.Module == module && operation.Gate == gate {
			return operation, true
		}
	}
	return config.Operation{}, false
}

func (runner Runner) runOperation(ctx context.Context, directory string, module inventory.Module, operation config.Operation) error {
	for index, step := range operation.Steps {
		timeout, err := time.ParseDuration(step.Timeout)
		if err != nil {
			return fmt.Errorf("%s %s step %d timeout: %w", module.Directory, operation.Gate, index, err)
		}
		stepContext, cancel := context.WithTimeout(ctx, timeout)
		command, err := operationCommand(directory, module, step)
		if err == nil {
			err = runner.Executor.Run(stepContext, command)
		}
		cancel()
		if err != nil {
			return fmt.Errorf("%s %s step %d: %w", module.Directory, operation.Gate, index, err)
		}
	}
	return nil
}

func operationCommand(directory string, module inventory.Module, step config.Step) (Command, error) {
	switch step.Type {
	case "go-test":
		args := []string{"test"}
		if len(module.TestTags) > 0 {
			args = append(args, "-tags="+strings.Join(module.TestTags, ","))
		}
		args = append(args, step.Packages...)
		args = append(args, fmt.Sprintf("-count=%d", step.Count), "-timeout="+step.Timeout)
		if step.Run != "" {
			args = append(args, "-run="+step.Run)
		}
		if step.Benchmark != "" {
			args = append(args, "-run=^$", "-bench="+step.Benchmark, "-benchmem", "-benchtime="+step.Budget)
		}
		if step.Fuzz != "" {
			args = append(args, "-run=^$", "-fuzz="+step.Fuzz, "-fuzztime="+step.Budget)
		}
		return Command{Name: "go", Args: args, Dir: directory, Env: map[string]string{"GOWORK": "off"}}, nil
	case "make":
		makefile, err := repositoryfile.Read(directory, step.Makefile, maximumMakefileSize)
		if err != nil {
			return Command{}, fmt.Errorf("read makefile: %w", err)
		}
		return Command{
			Name:  "make",
			Args:  []string{"--no-print-directory", "-f", "-", step.Target},
			Dir:   directory,
			Env:   map[string]string{"GOWORK": "off"},
			Stdin: bytes.NewReader(makefile),
		}, nil
	default:
		return Command{}, fmt.Errorf("unsupported operation type: %s", step.Type)
	}
}

func testArguments(tags []string, race bool) []string {
	args := []string{"test"}
	if race {
		args = append(args, "-race")
	}
	if len(tags) > 0 {
		args = append(args, "-tags="+strings.Join(tags, ","))
	}
	return append(args, "./...", "-count=1", "-timeout=20m")
}

func (runner Runner) command(ctx context.Context, output io.Writer, module, gate, directory string, args ...string) error {
	return announce(output, module, gate, func() error {
		if err := runner.Executor.Run(ctx, Command{Name: "go", Args: args, Dir: directory, Env: map[string]string{"GOWORK": "off"}}); err != nil {
			return fmt.Errorf("%s %s: %w", module, gate, err)
		}
		return nil
	})
}

func announce(output io.Writer, module, gate string, operation func() error) error {
	_, _ = fmt.Fprintf(output, "[%s] %s\n", module, gate)
	return operation()
}

func (runner Runner) checkFormatting(ctx context.Context, root string) error {
	files := make([]string, 0)
	if err := walkModuleFiles(root, func(path string, _ fs.DirEntry) error {
		relative := filepath.ToSlash(strings.TrimPrefix(path, root+string(filepath.Separator)))
		if strings.ContainsAny(relative, "\r\n") {
			return fmt.Errorf("go source path contains a line break: %q", relative)
		}
		files = append(files, relative)
		return nil
	}); err != nil {
		return err
	}
	if len(files) == 0 {
		return nil
	}
	sort.Strings(files)
	var output boundedBuffer
	if err := runner.Executor.Run(ctx, Command{
		Name: "gofmt", Args: append([]string{"-l", "--"}, files...), Dir: root,
		Stdout: &output,
	}); err != nil {
		return fmt.Errorf("run gofmt: %w", err)
	}
	if output.overflow {
		return errors.New("gofmt output exceeds limit")
	}
	unformatted := strings.TrimSpace(output.String())
	if unformatted == "" {
		return nil
	}
	return fmt.Errorf("unformatted Go file: %s", strings.Split(unformatted, "\n")[0])
}

func checkSafety(root string) error {
	return walkModuleFiles(root, func(path string, _ fs.DirEntry) error {
		if strings.HasSuffix(path, "_test.go") {
			return nil
		}
		parsed, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly|parser.ParseComments)
		if err != nil {
			return fmt.Errorf("parse %s: %w", path, err)
		}
		for _, imported := range parsed.Imports {
			if imported.Path.Value == `"unsafe"` || imported.Path.Value == `"C"` {
				return fmt.Errorf("forbidden production import %s in %s", imported.Path.Value, path)
			}
		}
		for _, group := range parsed.Comments {
			for _, comment := range group.List {
				if strings.HasPrefix(comment.Text, "//go:linkname") {
					return fmt.Errorf("forbidden go:linkname directive in %s", path)
				}
			}
		}
		return nil
	})
}

func walkModuleFiles(root string, visit func(string, fs.DirEntry) error) error {
	return filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if path != root {
				name := entry.Name()
				if name == "testdata" || name == "vendor" || strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_") {
					return filepath.SkipDir
				}
				if _, statErr := os.Stat(filepath.Join(path, "go.mod")); statErr == nil {
					return filepath.SkipDir
				} else if !os.IsNotExist(statErr) {
					return statErr
				}
			}
			return nil
		}
		if filepath.Ext(path) != ".go" {
			return nil
		}
		return visit(path, entry)
	})
}
