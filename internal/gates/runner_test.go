package gates_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/faustbrian/go-library-tools/internal/config"
	"github.com/faustbrian/go-library-tools/internal/gates"
	"github.com/faustbrian/go-library-tools/internal/inventory"
)

func TestCheckRunsStandardGatesInDeterministicOrder(t *testing.T) {
	root := fixture(t)
	write(t, filepath.Join(root, "README.md"), "# Example\n")
	executor := &recordingExecutor{}
	var output bytes.Buffer
	runner := gates.Runner{Root: root, Catalog: inventory.Inventory{Modules: []inventory.Module{{
		Directory: ".",
		Gates: map[string]bool{
			"lint": true, "tests": true, "race": true, "documentation": true,
		},
	}}}, Executor: executor, Output: &output, DocumentationSpelling: successfulSpelling}

	if err := runner.Check(context.Background(), []string{"."}); err != nil {
		t.Fatalf("Check() error = %v", err)
	}

	want := []string{
		"go mod tidy -diff",
		"go vet ./...",
		"go test ./... -count=1 -timeout=20m",
		"go test -race ./... -count=1 -timeout=20m",
		"go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2 run --allow-parallel-runners --timeout=10m ./...",
		"go run honnef.co/go/tools/cmd/staticcheck@v0.7.0 ./...",
		"go run go.uber.org/nilaway/cmd/nilaway@v0.0.0-20260720194628-9fd1b8d7bac8 -include-pkgs= ./...",
	}
	if !reflect.DeepEqual(executor.commands, want) {
		t.Fatalf("commands = %#v, want %#v", executor.commands, want)
	}
	for _, gate := range []string{"format-check", "tidy-check", "safety", "vet", "test", "race", "docs"} {
		if !strings.Contains(output.String(), "[.] "+gate+"\n") {
			t.Errorf("output does not include gate %q: %s", gate, output.String())
		}
	}
}

func TestCheckRunsTypedOperationsWithoutShellInterpretation(t *testing.T) {
	root := fixture(t)
	write(t, filepath.Join(root, "README.md"), "# Example\n")
	executor := &recordingExecutor{}
	runner := gates.Runner{
		Root:    root,
		Catalog: inventory.Inventory{Modules: []inventory.Module{{Directory: ".", TestTags: []string{"integration"}, Gates: map[string]bool{"documentation": true}}}},
		Policy: config.Config{Operations: []config.Operation{
			{Module: ".", Gate: "fuzz", Steps: []config.Step{
				{Type: "go-test", Packages: []string{"./..."}, Fuzz: "FuzzSpec", Budget: "100x", Count: 1, Timeout: "1m"},
			}},
			{Module: ".", Gate: "conformance", Steps: []config.Step{
				{Type: "go-test", Packages: []string{"./..."}, Run: "^TestSpec$", Count: 1, Timeout: "1m"},
			}},
			{Module: ".", Gate: "docs", Steps: []config.Step{
				{Type: "go-test", Packages: []string{"./..."}, Run: "^TestDocs$", Count: 1, Timeout: "1m"},
			}},
		}},
		Executor:              executor,
		DocumentationSpelling: successfulSpelling,
	}
	if err := runner.Check(context.Background(), []string{"."}); err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	wantSuffix := "go test -tags=integration ./... -count=1 -timeout=1m -run=^TestSpec$"
	if executor.commands[len(executor.commands)-1] != wantSuffix {
		t.Fatalf("operation commands = %#v", executor.commands)
	}
}

func TestCheckRequiresExactProductionCoverage(t *testing.T) {
	root := fixture(t)
	executor := &recordingExecutor{coverageProfile: "mode: atomic\ngithub.com/acme/example/file.go:1.1,2.1 1 1\n"}
	var output bytes.Buffer
	runner := gates.Runner{Root: root, Catalog: inventory.Inventory{Modules: []inventory.Module{{
		Directory: ".",
		TestTags:  []string{"integration"},
		Gates:     map[string]bool{"coverage": true},
		Packages:  []inventory.Package{{ImportPath: "github.com/acme/example", CoverageRequired: true}},
	}}}, Executor: executor, Output: &output}
	if err := runner.Check(context.Background(), []string{"."}); err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if !strings.Contains(output.String(), "all production packages have exact 100% statement coverage") {
		t.Fatalf("coverage output = %q", output.String())
	}

	executor.coverageProfile = "mode: atomic\ngithub.com/acme/example/file.go:1.1,2.1 1 0\n"
	if err := runner.Check(context.Background(), []string{"."}); err == nil {
		t.Fatal("Check() uncovered error = nil")
	}
}

func TestCoverageRunsStandaloneAndReportsSkippedModules(t *testing.T) {
	root := fixture(t)
	executor := &recordingExecutor{coverageProfile: "mode: atomic\nexample/file.go:1.1,2.1 1 1\n"}
	var output bytes.Buffer
	runner := gates.Runner{Root: root, Catalog: inventory.Inventory{Modules: []inventory.Module{
		{Directory: ".", Gates: map[string]bool{"coverage": true}, Packages: []inventory.Package{{ImportPath: "example", CoverageRequired: true}}},
		{Directory: "nested", Gates: map[string]bool{}},
	}}, Executor: executor, Output: &output}
	if err := runner.Coverage(context.Background(), []string{"nested", "."}); err != nil {
		t.Fatalf("Coverage() error = %v", err)
	}
	if !strings.Contains(output.String(), "exact 100%") || !strings.Contains(output.String(), "not applicable") {
		t.Fatalf("Coverage() output = %q", output.String())
	}
	if err := runner.Coverage(context.Background(), []string{"missing"}); err == nil {
		t.Fatal("Coverage() unknown module error = nil")
	}
	runner.Output = nil
	if err := runner.Coverage(context.Background(), []string{"nested"}); err != nil {
		t.Fatalf("Coverage() discarded output error = %v", err)
	}
	executor.failure = errors.New("failed")
	executor.failureAt = len(executor.commands)
	if err := runner.Coverage(context.Background(), []string{"."}); !errors.Is(err, executor.failure) {
		t.Fatalf("Coverage() execution error = %v", err)
	}
}

func TestDocsRunsNativeAndTypedChecks(t *testing.T) {
	root := fixture(t)
	write(t, filepath.Join(root, "README.md"), "# Root\n")
	executor := &recordingExecutor{}
	var output bytes.Buffer
	runner := gates.Runner{
		Root: root,
		Catalog: inventory.Inventory{Modules: []inventory.Module{
			{Directory: ".", Gates: map[string]bool{"documentation": true}},
			{Directory: "nested", Gates: map[string]bool{}},
		}},
		Policy:                config.Config{Operations: []config.Operation{{Module: ".", Gate: "docs", Steps: []config.Step{{Type: "go-test", Packages: []string{"."}, Run: "^TestDocs$", Count: 1, Timeout: "1m"}}}}},
		Executor:              executor,
		Output:                &output,
		DocumentationSpelling: successfulSpelling,
	}
	if err := runner.Docs(context.Background(), []string{"nested", "."}); err != nil {
		t.Fatalf("Docs() error = %v", err)
	}
	if !strings.Contains(output.String(), "[.] docs\n") || !strings.Contains(output.String(), "[nested] docs: not applicable") ||
		!strings.Contains(strings.Join(executor.commands, "\n"), "-run=^TestDocs$") {
		t.Fatalf("Docs() output = %q, commands = %#v", output.String(), executor.commands)
	}
	runner.Policy.Operations = nil
	if err := runner.Docs(context.Background(), []string{"."}); err != nil {
		t.Fatalf("Docs(native only) error = %v", err)
	}
	runner.Output = nil
	if err := runner.Docs(context.Background(), []string{"nested"}); err != nil {
		t.Fatalf("Docs(discarded output) error = %v", err)
	}
	if err := runner.Docs(context.Background(), []string{"missing"}); err == nil {
		t.Fatal("Docs(unknown module) error = nil")
	}
}

func TestDocsAndCheckReportNativeDocumentationFailure(t *testing.T) {
	root := fixture(t)
	write(t, filepath.Join(root, "README.md"), "[missing](missing.md)\n")
	module := inventory.Module{Directory: ".", Gates: map[string]bool{"documentation": true}}
	runner := gates.Runner{Root: root, Catalog: inventory.Inventory{Modules: []inventory.Module{module}}, Executor: &recordingExecutor{}}
	if err := runner.Docs(context.Background(), []string{"."}); err == nil || !strings.Contains(err.Error(), "broken local link") {
		t.Fatalf("Docs() error = %v", err)
	}
	if err := runner.Check(context.Background(), []string{"."}); err == nil || !strings.Contains(err.Error(), "broken local link") {
		t.Fatalf("Check() error = %v", err)
	}
}

func TestAPICheckAndUpdateUsePinnedTool(t *testing.T) {
	root := fixture(t)
	if err := os.MkdirAll(filepath.Join(root, "api"), 0o700); err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(root, "api", "baseline.txt"), "old baseline")
	executor := &recordingExecutor{apiSnapshot: "new baseline"}
	var output bytes.Buffer
	runner := gates.Runner{Root: root, Catalog: inventory.Inventory{Modules: []inventory.Module{{
		Directory: ".", ModulePath: "example", Gates: map[string]bool{"api_compatibility": true},
	}}}, Executor: executor, Output: &output}
	if err := runner.API(context.Background(), []string{"."}, false); err != nil {
		t.Fatalf("API(check) error = %v", err)
	}
	if !strings.Contains(output.String(), "API compatibility passed") ||
		!strings.Contains(strings.Join(executor.commands, "\n"), "apidiff@v0.0.0-20260718201538-764159d718ef") {
		t.Fatalf("API check output/commands = %q, %#v", output.String(), executor.commands)
	}
	output.Reset()
	if err := runner.API(context.Background(), []string{"."}, true); err != nil {
		t.Fatalf("API(update) error = %v", err)
	}
	updated, err := os.ReadFile(filepath.Join(root, "api", "baseline.txt"))
	if err != nil || string(updated) != "new baseline" {
		t.Fatalf("updated baseline = %q, %v", updated, err)
	}
}

func TestAPIRejectsMissingEmptyIncompatibleAndFailedSnapshots(t *testing.T) {
	root := fixture(t)
	module := inventory.Module{Directory: ".", ModulePath: "example", Gates: map[string]bool{"api_compatibility": true}}
	runner := gates.Runner{Root: root, Catalog: inventory.Inventory{Modules: []inventory.Module{module}}, Executor: &recordingExecutor{apiSnapshot: "snapshot"}}
	if err := runner.API(context.Background(), []string{"."}, false); err == nil || !strings.Contains(err.Error(), "missing API baseline") {
		t.Fatalf("API() missing baseline error = %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "api"), 0o700); err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(root, "api", "baseline.txt"), "baseline")
	runner.Executor = &recordingExecutor{}
	if err := runner.API(context.Background(), []string{"."}, false); err == nil || !strings.Contains(err.Error(), "is empty") {
		t.Fatalf("API() empty snapshot error = %v", err)
	}
	failure := errors.New("tool failed")
	runner.Executor = &recordingExecutor{apiSnapshot: "snapshot", apiReport: "breaking change", failureAt: 1, failure: failure}
	if err := runner.API(context.Background(), []string{"."}, false); err == nil || !strings.Contains(err.Error(), "breaking change") {
		t.Fatalf("API() incompatible error = %v", err)
	}
	runner.Executor = &recordingExecutor{apiSnapshot: "snapshot", failureAt: 1, failure: failure}
	if err := runner.API(context.Background(), []string{"."}, false); !errors.Is(err, failure) {
		t.Fatalf("API() tool error = %v", err)
	}
	runner.Executor = &recordingExecutor{failureAt: 0, failure: failure}
	if err := runner.API(context.Background(), []string{"."}, false); !errors.Is(err, failure) {
		t.Fatalf("API() snapshot error = %v", err)
	}
}

func TestAPIRejectsSymlinkBaselineAndDirectory(t *testing.T) {
	root := fixture(t)
	target := filepath.Join(root, "target")
	write(t, target, "outside")
	if err := os.Mkdir(filepath.Join(root, "api"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(root, "api", "baseline.txt")); err != nil {
		t.Fatal(err)
	}
	module := inventory.Module{Directory: ".", ModulePath: "example", Gates: map[string]bool{"api_compatibility": true}}
	runner := gates.Runner{Root: root, Catalog: inventory.Inventory{Modules: []inventory.Module{module}}, Executor: &recordingExecutor{apiSnapshot: "snapshot"}}
	if err := runner.API(context.Background(), []string{"."}, false); err == nil {
		t.Fatal("API() symlink baseline error = nil")
	}
	second := fixture(t)
	targetDirectory := filepath.Join(second, "target")
	if err := os.Mkdir(targetDirectory, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(targetDirectory, filepath.Join(second, "api")); err != nil {
		t.Fatal(err)
	}
	runner.Root = second
	if err := runner.API(context.Background(), []string{"."}, true); err == nil {
		t.Fatal("API() symlink directory error = nil")
	}
}

func TestAPISkipsDisabledModulesAndRejectsUnknownSelection(t *testing.T) {
	executor := &recordingExecutor{}
	var output bytes.Buffer
	runner := gates.Runner{Root: fixture(t), Catalog: inventory.Inventory{Modules: []inventory.Module{{Directory: "."}}}, Executor: executor, Output: &output}
	if err := runner.API(context.Background(), []string{"."}, false); err != nil || len(executor.commands) != 0 {
		t.Fatalf("API() skip = %#v, %v", executor.commands, err)
	}
	if !strings.Contains(output.String(), "not applicable") {
		t.Fatalf("API() skip output = %q", output.String())
	}
	if err := runner.API(context.Background(), []string{"missing"}, false); err == nil {
		t.Fatal("API() unknown module error = nil")
	}
}

func TestCheckRunsSecurityTools(t *testing.T) {
	root := fixture(t)
	executor := &recordingExecutor{}
	runner := gates.Runner{Root: root, Catalog: inventory.Inventory{Modules: []inventory.Module{{
		Directory: ".", ModulePath: "github.com/acme/example", Gates: map[string]bool{"security": true},
	}}}, Executor: executor}
	if err := runner.Check(context.Background(), []string{"."}); err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	want := []string{
		"go mod tidy -diff",
		"go run golang.org/x/vuln/cmd/govulncheck@v1.6.0 ./...",
		"go run github.com/zricethezav/gitleaks/v8@v8.30.1 dir . --config " + filepath.Join(root, ".gitleaks.toml") + " --no-banner --redact",
		"go run github.com/google/go-licenses/v2@v2.0.1 check ./... --ignore github.com/acme/example",
	}
	if !reflect.DeepEqual(executor.commands, want) {
		t.Fatalf("commands = %#v", executor.commands)
	}
}

func TestCheckTreatsNilAwayAsAdvisory(t *testing.T) {
	executor := &recordingExecutor{failureAt: 4, failure: errors.New("finding")}
	var output bytes.Buffer
	runner := gates.Runner{Root: fixture(t), Catalog: inventory.Inventory{Modules: []inventory.Module{{Directory: ".", Gates: map[string]bool{"lint": true}}}}, Executor: executor, Output: &output}
	if err := runner.Check(context.Background(), []string{"."}); err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if !strings.Contains(output.String(), "NilAway advisory") {
		t.Fatalf("output = %q", output.String())
	}
}

func TestCheckStopsAtAnalyzerAndSecurityFailures(t *testing.T) {
	tests := []struct {
		name      string
		gates     map[string]bool
		failureAt int
	}{
		{"lint", map[string]bool{"lint": true}, 2},
		{"staticcheck", map[string]bool{"lint": true}, 3},
		{"vulnerability", map[string]bool{"security": true}, 1},
		{"secrets", map[string]bool{"security": true}, 2},
		{"licenses", map[string]bool{"security": true}, 3},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			failure := errors.New("failed")
			executor := &recordingExecutor{failureAt: test.failureAt, failure: failure}
			runner := gates.Runner{Root: fixture(t), Catalog: inventory.Inventory{Modules: []inventory.Module{{Directory: ".", Gates: test.gates}}}, Executor: executor}
			if err := runner.Check(context.Background(), []string{"."}); !errors.Is(err, failure) {
				t.Fatalf("Check() error = %v", err)
			}
		})
	}
}

func TestCheckStopsAtCoverageAndTypedOperationFailures(t *testing.T) {
	failure := errors.New("failed")
	tests := []struct {
		name   string
		module inventory.Module
		policy config.Config
	}{
		{"coverage", inventory.Module{Directory: ".", Gates: map[string]bool{"coverage": true}, Packages: []inventory.Package{{ImportPath: "example", CoverageRequired: true}}}, config.Config{}},
		{"fuzz operation", inventory.Module{Directory: "."}, config.Config{Operations: []config.Operation{{Module: ".", Gate: "fuzz", Steps: []config.Step{{Type: "go-test", Packages: []string{"."}, Count: 1, Timeout: "1m"}}}}}},
		{"operation", inventory.Module{Directory: "."}, config.Config{Operations: []config.Operation{{Module: ".", Gate: "conformance", Steps: []config.Step{{Type: "go-test", Packages: []string{"."}, Count: 1, Timeout: "1m"}}}}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			executor := &recordingExecutor{failureAt: 1, failure: failure}
			runner := gates.Runner{Root: fixture(t), Catalog: inventory.Inventory{Modules: []inventory.Module{test.module}}, Policy: test.policy, Executor: executor}
			if err := runner.Check(context.Background(), []string{"."}); !errors.Is(err, failure) {
				t.Fatalf("Check() error = %v", err)
			}
		})
	}
}

func TestCheckAppliesModuleTestTagsAndSkipsDisabledGates(t *testing.T) {
	root := fixture(t)
	executor := &recordingExecutor{}
	runner := gates.Runner{Root: root, Catalog: inventory.Inventory{Modules: []inventory.Module{{
		Directory: "nested", TestTags: []string{"integration", "postgres"},
		Gates: map[string]bool{"tests": true},
	}}}, Executor: executor}

	if err := runner.Check(context.Background(), []string{"nested"}); err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	want := []string{
		"go mod tidy -diff",
		"go test -tags=integration,postgres ./... -count=1 -timeout=20m",
	}
	if !reflect.DeepEqual(executor.commands, want) {
		t.Fatalf("commands = %#v, want %#v", executor.commands, want)
	}
}

func TestCheckRejectsUnknownSelectionAndStopsAtFailure(t *testing.T) {
	root := fixture(t)
	failure := errors.New("failed")
	executor := &recordingExecutor{failureAt: 1, failure: failure}
	runner := gates.Runner{Root: root, Catalog: inventory.Inventory{Modules: []inventory.Module{{Directory: ".", Gates: map[string]bool{"tests": true}}}}, Executor: executor}

	if err := runner.Check(context.Background(), []string{"missing"}); err == nil || !strings.Contains(err.Error(), "unknown module") {
		t.Fatalf("Check() unknown error = %v", err)
	}
	if err := runner.Check(context.Background(), []string{"."}); !errors.Is(err, failure) {
		t.Fatalf("Check() failure = %v", err)
	}
	if len(executor.commands) != 2 {
		t.Fatalf("executed %d commands after failure", len(executor.commands))
	}
}

func TestCheckStopsAtExternalGateFailures(t *testing.T) {
	tests := []struct {
		name      string
		gates     map[string]bool
		failureAt int
	}{
		{"tidy", map[string]bool{}, 0},
		{"vet", map[string]bool{"lint": true}, 1},
		{"race", map[string]bool{"tests": true, "race": true}, 2},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			failure := errors.New("failed")
			executor := &recordingExecutor{failureAt: test.failureAt, failure: failure}
			runner := gates.Runner{Root: fixture(t), Catalog: inventory.Inventory{Modules: []inventory.Module{{Directory: ".", Gates: test.gates}}}, Executor: executor}
			if err := runner.Check(context.Background(), []string{"."}); !errors.Is(err, failure) {
				t.Fatalf("Check() error = %v", err)
			}
		})
	}
}

func TestCheckRejectsUnformattedAndUnsafeSources(t *testing.T) {
	tests := map[string]string{
		"unformatted":   "package example\nfunc Value( )int{return 1}\n",
		"unsafe import": "package example\n\nimport \"unsafe\"\n\nvar _ = unsafe.Sizeof(0)\n",
		"cgo":           "package example\n\nimport \"C\"\n",
		"linkname":      "package example\n\n//go:linkname local target\nfunc local()\n",
	}
	for name, source := range tests {
		t.Run(name, func(t *testing.T) {
			root := fixture(t)
			write(t, filepath.Join(root, "unsafe.go"), source)
			runner := gates.Runner{Root: root, Catalog: inventory.Inventory{Modules: []inventory.Module{{Directory: "."}}}, Executor: &recordingExecutor{}}
			if err := runner.Check(context.Background(), []string{"."}); err == nil {
				t.Fatal("Check() error = nil")
			}
		})
	}
}

type recordingExecutor struct {
	commands        []string
	failureAt       int
	failure         error
	coverageProfile string
	apiSnapshot     string
	apiReport       string
}

func successfulSpelling(context.Context, string) error { return nil }

func (executor *recordingExecutor) Run(_ context.Context, command gates.Command) error {
	executor.commands = append(executor.commands, strings.Join(append([]string{command.Name}, command.Args...), " "))
	for _, argument := range command.Args {
		if strings.HasPrefix(argument, "-coverprofile=") {
			if err := os.WriteFile(strings.TrimPrefix(argument, "-coverprofile="), []byte(executor.coverageProfile), 0o600); err != nil {
				return err
			}
		}
	}
	for index, argument := range command.Args {
		if argument == "-w" && index+1 < len(command.Args) {
			if err := os.WriteFile(command.Args[index+1], []byte(executor.apiSnapshot), 0o600); err != nil {
				return err
			}
		}
		if argument == "-incompatible" && command.Stdout != nil {
			_, _ = io.WriteString(command.Stdout, executor.apiReport)
		}
	}
	if executor.failure != nil && len(executor.commands)-1 == executor.failureAt {
		return executor.failure
	}
	return nil
}

func fixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	write(t, filepath.Join(root, "go.mod"), "module example\n\ngo 1.27.0\n")
	write(t, filepath.Join(root, "example.go"), "package example\n\nfunc Value() int { return 1 }\n")
	if err := os.MkdirAll(filepath.Join(root, "nested"), 0o700); err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(root, "nested", "go.mod"), "module example/nested\n\ngo 1.27.0\n")
	write(t, filepath.Join(root, "nested", "example.go"), "package nested\n\nfunc Value() int { return 1 }\n")
	return root
}

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}
