package cli

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/faustbrian/go-library-tools/internal/gates"
)

func TestExecuteReportsExecutorCreationAndCleanupFailures(t *testing.T) {
	root := internalFixture(t)
	failure := errors.New("injected failure")
	tests := []struct {
		name    string
		factory executorFactory
	}{
		{"creation", func(string, io.Writer, io.Writer) (gates.Executor, func() error, error) {
			return nil, nil, failure
		}},
		{"cleanup", func(string, io.Writer, io.Writer) (gates.Executor, func() error, error) {
			return successfulExecutor{}, func() error { return failure }, nil
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := execute([]string{"check"}, root, &stdout, &stderr, test.factory)
			if code != 1 || !strings.Contains(stderr.String(), failure.Error()) {
				t.Fatalf("execute() code = %d, stderr = %q", code, stderr.String())
			}
		})
	}
}

func TestExecuteContextPropagatesCancellationToCommands(t *testing.T) {
	root := internalFixture(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	factory := func(string, io.Writer, io.Writer) (gates.Executor, func() error, error) {
		return cliExecutorFunction(func(commandContext context.Context, _ gates.Command) error {
			return commandContext.Err()
		}), func() error { return nil }, nil
	}
	var stdout, stderr bytes.Buffer
	if code := executeContext(ctx, []string{"check"}, root, &stdout, &stderr, factory); code != 1 || !strings.Contains(stderr.String(), context.Canceled.Error()) {
		t.Fatalf("executeContext() = %d, %q", code, stderr.String())
	}
}

func TestExecuteRoutesWorkflowChecks(t *testing.T) {
	root := internalFixture(t)
	var command gates.Command
	factory := func(string, io.Writer, io.Writer) (gates.Executor, func() error, error) {
		return cliExecutorFunction(func(_ context.Context, value gates.Command) error {
			command = value
			return nil
		}), func() error { return nil }, nil
	}
	var stdout, stderr bytes.Buffer
	if code := execute([]string{"workflows", "check"}, root, &stdout, &stderr, factory); code != 0 || stderr.Len() != 0 {
		t.Fatalf("execute(workflows check) = %d, %q", code, stderr.String())
	}
	if command.Name != "go" || command.Dir != root || !strings.Contains(strings.Join(command.Args, " "), "actionlint@v1.7.12") {
		t.Fatalf("workflow command = %#v", command)
	}
	for _, args := range [][]string{{"workflows"}, {"workflows", "lint"}, {"workflows", "check", "extra"}} {
		stderr.Reset()
		if code := execute(args, root, &stdout, &stderr, factory); code != 2 || !strings.Contains(stderr.String(), "usage: golib workflows") {
			t.Fatalf("execute(%v) = %d, %q", args, code, stderr.String())
		}
	}
}

func TestExecuteRoutesAPICheckAndUpdate(t *testing.T) {
	root := internalFixture(t)
	if err := os.Mkdir(filepath.Join(root, "api"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "api", "baseline.txt"), []byte("baseline"), 0o600); err != nil {
		t.Fatal(err)
	}
	manifest := `{"schema_version":1,"repository":"example","go_version":"1.27.0","modules":[{"directory":".","module_path":"example","go_version":"1.27.0","kind":"public","releasable":true,"gates":{"api_compatibility":true},"packages":[]}]}`
	if err := os.WriteFile(filepath.Join(root, "modules.json"), []byte(manifest), 0o600); err != nil {
		t.Fatal(err)
	}
	factory := func(string, io.Writer, io.Writer) (gates.Executor, func() error, error) {
		return apiExecutor{}, func() error { return nil }, nil
	}
	tests := []struct {
		args       []string
		wantOutput string
		wantBase   string
	}{
		{args: []string{"api", "check"}, wantOutput: "API compatibility passed", wantBase: "baseline"},
		{args: []string{"api", "update", "--module", "."}, wantOutput: "API baseline updated", wantBase: "snapshot"},
	}
	for _, test := range tests {
		if err := os.WriteFile(filepath.Join(root, "api", "baseline.txt"), []byte("baseline"), 0o600); err != nil {
			t.Fatal(err)
		}
		var stdout, stderr bytes.Buffer
		if code := execute(test.args, root, &stdout, &stderr, factory); code != 0 || stderr.Len() != 0 || !strings.Contains(stdout.String(), test.wantOutput) {
			t.Fatalf("execute(%v) = %d, %q, %q", test.args, code, stdout.String(), stderr.String())
		}
		baseline, err := os.ReadFile(filepath.Join(root, "api", "baseline.txt"))
		if err != nil || string(baseline) != test.wantBase {
			t.Fatalf("execute(%v) baseline = %q, %v", test.args, baseline, err)
		}
	}
}

func TestExecuteRejectsInvalidAPIArguments(t *testing.T) {
	for _, args := range [][]string{{"api"}, {"api", "remove"}, {"api", "check", "--all", "extra"}} {
		var stdout, stderr bytes.Buffer
		if code := execute(args, internalFixture(t), &stdout, &stderr, nil); code != 2 || !strings.Contains(stderr.String(), "usage: golib api") {
			t.Fatalf("execute(%v) = %d, %q", args, code, stderr.String())
		}
	}
}

func TestExecuteRoutesStandaloneCoverage(t *testing.T) {
	root := internalFixture(t)
	manifest := `{"schema_version":1,"repository":"example","go_version":"1.27.0","modules":[{"directory":".","module_path":"example","go_version":"1.27.0","kind":"public","releasable":true,"gates":{"coverage":true},"packages":[{"module_directory":".","directory":".","name":"example","import_path":"example","kind":"public","production":true,"executable":true,"coverage_required":true,"build_required":true,"build_tags":[]}]}]}`
	if err := os.WriteFile(filepath.Join(root, "modules.json"), []byte(manifest), 0o600); err != nil {
		t.Fatal(err)
	}
	packages := `{"schema_version":1,"repository":"example","packages":[{"module_directory":".","directory":".","name":"example","import_path":"example","kind":"public","production":true,"executable":true,"coverage_required":true,"build_required":true,"build_tags":[]}]}`
	if err := os.WriteFile(filepath.Join(root, "packages.json"), []byte(packages), 0o600); err != nil {
		t.Fatal(err)
	}
	factory := func(string, io.Writer, io.Writer) (gates.Executor, func() error, error) {
		return coverageExecutor{}, func() error { return nil }, nil
	}
	var stdout, stderr bytes.Buffer
	if code := execute([]string{"coverage", "--module", "."}, root, &stdout, &stderr, factory); code != 0 || stderr.Len() != 0 {
		t.Fatalf("execute() = %d, %q", code, stderr.String())
	}
	if code := execute([]string{"coverage", "--bad"}, root, &stdout, &stderr, factory); code != 2 {
		t.Fatalf("execute() invalid coverage = %d", code)
	}
}

func TestExecuteRoutesStandaloneMutation(t *testing.T) {
	root := internalFixture(t)
	factory := func(string, io.Writer, io.Writer) (gates.Executor, func() error, error) {
		return coverageExecutor{}, func() error { return nil }, nil
	}
	var stdout, stderr bytes.Buffer
	if code := execute([]string{"mutation"}, root, &stdout, &stderr, factory); code != 0 || stderr.Len() != 0 {
		t.Fatalf("execute() default mutation = %d, %q", code, stderr.String())
	}
	if code := execute([]string{"mutation", "--module", "."}, root, &stdout, &stderr, factory); code != 0 || stderr.Len() != 0 {
		t.Fatalf("execute() mutation = %d, %q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "not applicable") {
		t.Fatalf("mutation output = %q", stdout.String())
	}
	if code := execute([]string{"mutation", "--bad"}, root, &stdout, &stderr, factory); code != 2 {
		t.Fatalf("execute() invalid mutation = %d", code)
	}
	if code := execute([]string{"mutation", "import"}, root, &stdout, &stderr, factory); code != 2 {
		t.Fatalf("execute() invalid mutation import = %d", code)
	}
	manifest := `{"schema_version":1,"repository":"example","go_version":"1.27.0","modules":[{"directory":".","module_path":"example","go_version":"1.27.0","kind":"public","releasable":true,"gates":{"mutation":true},"packages":[]}]}`
	if err := os.WriteFile(filepath.Join(root, "modules.json"), []byte(manifest), 0o600); err != nil {
		t.Fatal(err)
	}
	stderr.Reset()
	if code := execute([]string{"mutation", "import", "--module", ".", "--archive", "legacy.zip", "--ledger", "ledger.json"}, root, &stdout, &stderr, factory); code != 1 || !strings.Contains(stderr.String(), "checkpoint archive") {
		t.Fatalf("execute() mutation import = %d, %q", code, stderr.String())
	}
}

func TestMutationImportArguments(t *testing.T) {
	options, err := mutationImportArguments([]string{"--ledger", "ledger.json", "--module", ".", "--archive", "legacy.zip"})
	if err != nil || options.module != "." || options.archive != "legacy.zip" || options.ledger != "ledger.json" {
		t.Fatalf("mutationImportArguments() = %#v, %v", options, err)
	}
	for _, args := range [][]string{
		nil,
		{"--module", ".", "--archive", "legacy.zip", "--unknown", "ledger.json"},
		{"--module", ".", "--module", ".", "--ledger", "ledger.json"},
		{"--module", ".", "--archive", "", "--ledger", "ledger.json"},
		{"--module", ".", "--archive", "legacy.zip", "--archive", "other.zip"},
	} {
		if _, err := mutationImportArguments(args); err == nil {
			t.Fatalf("mutationImportArguments(%v) error = nil", args)
		}
	}
}

func TestExecuteRoutesDocumentationCheck(t *testing.T) {
	root := internalFixture(t)
	manifest := `{"schema_version":1,"repository":"example","go_version":"1.27.0","modules":[{"directory":".","module_path":"example","go_version":"1.27.0","kind":"public","releasable":true,"gates":{},"packages":[]}]}`
	if err := os.WriteFile(filepath.Join(root, "modules.json"), []byte(manifest), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("# Example\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "cspell.json"), []byte("{}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	factory := func(string, io.Writer, io.Writer) (gates.Executor, func() error, error) {
		return cliWorkspaceExecutor{directory: t.TempDir()}, func() error { return nil }, nil
	}
	var stdout, stderr bytes.Buffer
	if code := execute([]string{"docs", "check"}, root, &stdout, &stderr, factory); code != 0 || stderr.Len() != 0 || !strings.Contains(stdout.String(), "docs: not applicable") {
		t.Fatalf("execute() = %d, %q, %q", code, stdout.String(), stderr.String())
	}
	for _, args := range [][]string{{"docs"}, {"docs", "update"}, {"docs", "check", "--bad"}} {
		stderr.Reset()
		if code := execute(args, root, &stdout, &stderr, factory); code != 2 || !strings.Contains(stderr.String(), "usage: golib docs") {
			t.Fatalf("execute(%v) = %d, %q", args, code, stderr.String())
		}
	}
}

func TestExecuteRoutesReleaseChecks(t *testing.T) {
	root := internalFixture(t)
	manifest := `{"schema_version":1,"repository":"example","go_version":"1.27.0","modules":[{"directory":".","module_path":"example","go_version":"1.27.0","kind":"public","releasable":true,"version":"1.0.0","tag_prefix":"v","gates":{"coverage":true,"documentation":true,"lint":true,"mutation":true,"race":true,"security":true,"tests":true},"packages":[]}]}`
	if err := os.WriteFile(filepath.Join(root, "modules.json"), []byte(manifest), 0o600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if code := execute([]string{"release", "check"}, root, &stdout, &stderr, nil); code != 0 || stderr.Len() != 0 || !strings.Contains(stdout.String(), "release contract passed") {
		t.Fatalf("execute(release check) = %d, %q, %q", code, stdout.String(), stderr.String())
	}
	factory := func(string, io.Writer, io.Writer) (gates.Executor, func() error, error) {
		return successfulExecutor{}, func() error { return nil }, nil
	}
	stderr.Reset()
	if code := execute([]string{"release", "dry-run"}, root, &stdout, &stderr, factory); code != 1 || stderr.Len() == 0 {
		t.Fatalf("execute(release dry-run) = %d, %q", code, stderr.String())
	}
	for _, args := range [][]string{{"release"}, {"release", "publish"}, {"release", "check", "extra"}} {
		stderr.Reset()
		if code := execute(args, root, &stdout, &stderr, nil); code != 2 || !strings.Contains(stderr.String(), "usage: golib release") {
			t.Fatalf("execute(%v) = %d, %q", args, code, stderr.String())
		}
	}
}

func TestExecuteReportsReleaseContractFailures(t *testing.T) {
	root := internalFixture(t)
	var stdout, stderr bytes.Buffer
	if code := execute([]string{"release", "check"}, root, &stdout, &stderr, nil); code != 1 || !strings.Contains(stderr.String(), "stable version") {
		t.Fatalf("execute(release invalid) = %d, %q", code, stderr.String())
	}
	manifest := `{"schema_version":1,"repository":"example","go_version":"1.27.0","modules":[{"directory":".","module_path":"example","go_version":"1.27.0","kind":"public","releasable":true,"version":"1.0.0","tag_prefix":"v","gates":{"coverage":true,"documentation":true,"lint":true,"mutation":true,"race":true,"security":true,"tests":true},"packages":[]}]}`
	if err := os.WriteFile(filepath.Join(root, "modules.json"), []byte(manifest), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".go-version"), []byte("1.26.0\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	stderr.Reset()
	if code := execute([]string{"release", "check"}, root, &stdout, &stderr, nil); code != 1 || !strings.Contains(stderr.String(), ".go-version") {
		t.Fatalf("execute(release repository) = %d, %q", code, stderr.String())
	}
}

type successfulExecutor struct{}

func (successfulExecutor) Run(context.Context, gates.Command) error { return nil }

type cliExecutorFunction func(context.Context, gates.Command) error

func (function cliExecutorFunction) Run(ctx context.Context, command gates.Command) error {
	return function(ctx, command)
}

type cliWorkspaceExecutor struct {
	directory string
}

func (cliWorkspaceExecutor) Run(context.Context, gates.Command) error { return nil }
func (executor cliWorkspaceExecutor) TemporaryDirectory() string      { return executor.directory }

type apiExecutor struct{}

func (apiExecutor) Run(_ context.Context, command gates.Command) error {
	for index, argument := range command.Args {
		if argument == "-w" {
			return os.WriteFile(command.Args[index+1], []byte("snapshot"), 0o600)
		}
	}
	return nil
}

type coverageExecutor struct{}

func (coverageExecutor) Run(_ context.Context, command gates.Command) error {
	for _, argument := range command.Args {
		if profile, found := strings.CutPrefix(argument, "-coverprofile="); found {
			return os.WriteFile(profile, []byte("mode: atomic\nexample/file.go:1.1,2.1 1 1\n"), 0o600)
		}
	}
	return nil
}

func internalFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	files := map[string]string{
		".golib.yaml":   "schema_version: 1\ntool_version: v1.0.0\n",
		".go-version":   "1.27.0\n",
		"go.mod":        "module example\n\ngo 1.27.0\n",
		"example.go":    "package example\n",
		"modules.json":  `{"schema_version":1,"repository":"example","go_version":"1.27.0","modules":[{"directory":".","module_path":"example","go_version":"1.27.0","kind":"public","releasable":true,"gates":{},"packages":[]}]}`,
		"packages.json": `{"schema_version":1,"repository":"example","packages":[]}`,
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(root, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return root
}
