package gates

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/faustbrian/go-library-tools/internal/config"
	"github.com/faustbrian/go-library-tools/internal/inventory"
)

func TestFormattingAndSafetyReportMalformedOrUnreadableTrees(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "bad.go"), []byte("package ["), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := checkFormatting(root); err == nil || !strings.Contains(err.Error(), "format ") {
		t.Fatalf("checkFormatting() error = %v", err)
	}
	if err := checkSafety(root); err == nil || !strings.Contains(err.Error(), "parse ") {
		t.Fatalf("checkSafety() error = %v", err)
	}
	if err := checkFormatting(filepath.Join(root, "missing")); err == nil {
		t.Fatal("checkFormatting() missing root error = nil")
	}
}

func TestCoverageReportsOwnedFileAndExecutionFailures(t *testing.T) {
	failure := errors.New("injected failure")
	module := inventory.Module{Directory: ".", Packages: []inventory.Package{{ImportPath: "example", CoverageRequired: true}}}
	tests := []struct {
		name     string
		files    coverageFileSystem
		executor Executor
		want     string
	}{
		{"create", &fakeCoverageFiles{createErr: failure}, executorFunction(func(context.Context, Command) error { return nil }), "create coverage profile"},
		{"close", &fakeCoverageFiles{file: &fakeNamedFile{name: "profile", closeErr: failure}}, executorFunction(func(context.Context, Command) error { return nil }), "close coverage profile"},
		{"execute", &fakeCoverageFiles{file: &fakeNamedFile{name: "profile"}}, executorFunction(func(context.Context, Command) error { return failure }), "injected failure"},
		{"open", &fakeCoverageFiles{file: &fakeNamedFile{name: "profile"}, openErr: failure}, executorFunction(func(context.Context, Command) error { return nil }), "open coverage profile"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			runner := Runner{Executor: test.executor, coverageFiles: test.files}
			err := runner.runCoverage(context.Background(), io.Discard, t.TempDir(), module)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("runCoverage() error = %v", err)
			}
		})
	}
}

func TestCoverageUsesTaskWorkspace(t *testing.T) {
	files := &fakeCoverageFiles{file: &fakeNamedFile{name: "profile"}}
	executor := workspaceExecutor{directory: "/task", run: func(context.Context, Command) error { return nil }}
	runner := Runner{Executor: executor, coverageFiles: files}
	module := inventory.Module{Directory: ".", Packages: []inventory.Package{{ImportPath: "example", CoverageRequired: true}}}
	if err := runner.runCoverage(context.Background(), io.Discard, t.TempDir(), module); err != nil {
		t.Fatalf("runCoverage() error = %v", err)
	}
	if files.directory != "/task" {
		t.Fatalf("coverage directory = %q", files.directory)
	}
}

func TestDocumentationSpellingUsesPinnedTaskOwnedInstall(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "cspell.json"), []byte("{}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	task, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	commands := make([]Command, 0, 2)
	executor := workspaceExecutor{directory: task, run: func(_ context.Context, command Command) error {
		commands = append(commands, command)
		return nil
	}}
	runner := Runner{Executor: executor}
	if err := runner.runDocumentationSpelling(context.Background(), root); err != nil {
		t.Fatalf("runDocumentationSpelling() error = %v", err)
	}
	if len(commands) != 2 || commands[0].Name != "npm" || !strings.HasSuffix(commands[1].Name, filepath.Join("node_modules", ".bin", "cspell")) {
		t.Fatalf("commands = %#v", commands)
	}
	if commands[0].Dir != filepath.Join(task, "documentation", "spelling") ||
		!strings.HasPrefix(commands[0].Env["NPM_CONFIG_CACHE"], task+string(filepath.Separator)) {
		t.Fatalf("npm command is not task-owned: %#v", commands[0])
	}
	for _, name := range []string{"package.json", "package-lock.json"} {
		if info, err := os.Stat(filepath.Join(commands[0].Dir, name)); err != nil || !info.Mode().IsRegular() {
			t.Fatalf("asset %s = %v, %v", name, info, err)
		}
	}
}

func TestDocumentationSpellingRequiresTaskWorkspace(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "cspell.json"), []byte("{}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	runner := Runner{Executor: executorFunction(func(context.Context, Command) error { return nil })}
	if err := runner.runDocumentationSpelling(context.Background(), root); err == nil || !strings.Contains(err.Error(), "task-owned workspace") {
		t.Fatalf("runDocumentationSpelling() error = %v", err)
	}
}

func TestDocumentationSpellingReportsConfigurationAndWorkspaceFailures(t *testing.T) {
	t.Run("configuration", func(t *testing.T) {
		runner := Runner{Executor: workspaceExecutor{directory: t.TempDir(), run: func(context.Context, Command) error { return nil }}}
		if err := runner.runDocumentationSpelling(context.Background(), t.TempDir()); err == nil || !strings.Contains(err.Error(), "read cspell configuration") {
			t.Fatalf("runDocumentationSpelling() error = %v", err)
		}
	})

	t.Run("tool root", func(t *testing.T) {
		root := spellingRepository(t)
		workspace := filepath.Join(t.TempDir(), "workspace")
		if err := os.WriteFile(workspace, []byte("occupied"), 0o600); err != nil {
			t.Fatal(err)
		}
		runner := Runner{Executor: workspaceExecutor{directory: workspace, run: func(context.Context, Command) error { return nil }}}
		if err := runner.runDocumentationSpelling(context.Background(), root); err == nil || !strings.Contains(err.Error(), "create spelling tool root") {
			t.Fatalf("runDocumentationSpelling() error = %v", err)
		}
	})

	t.Run("asset", func(t *testing.T) {
		root := spellingRepository(t)
		workspace := t.TempDir()
		packagePath := filepath.Join(workspace, "documentation", "spelling", "package.json")
		if err := os.MkdirAll(packagePath, 0o700); err != nil {
			t.Fatal(err)
		}
		runner := Runner{Executor: workspaceExecutor{directory: workspace, run: func(context.Context, Command) error { return nil }}}
		if err := runner.runDocumentationSpelling(context.Background(), root); err == nil || !strings.Contains(err.Error(), "write spelling tool package.json") {
			t.Fatalf("runDocumentationSpelling() error = %v", err)
		}
	})
}

func TestDocumentationSpellingReportsToolFailures(t *testing.T) {
	failure := errors.New("injected failure")
	tests := []struct {
		name      string
		failAt    int
		want      string
		wantCalls int
	}{
		{name: "install", failAt: 1, want: "install pinned spelling tool", wantCalls: 1},
		{name: "spellcheck", failAt: 2, want: "check documentation spelling", wantCalls: 2},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			calls := 0
			executor := workspaceExecutor{directory: t.TempDir(), run: func(context.Context, Command) error {
				calls++
				if calls == test.failAt {
					return failure
				}
				return nil
			}}
			runner := Runner{Executor: executor}
			if err := runner.runDocumentationSpelling(context.Background(), spellingRepository(t)); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("runDocumentationSpelling() error = %v", err)
			}
			if calls != test.wantCalls {
				t.Fatalf("tool calls = %d, want %d", calls, test.wantCalls)
			}
		})
	}
}

func spellingRepository(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "cspell.json"), []byte("{}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	return root
}

func TestCheckDocumentationReportsNativeAndSpellingFailures(t *testing.T) {
	failure := errors.New("injected failure")
	runner := Runner{DocumentationSpelling: func(context.Context, string) error { return failure }}
	if err := runner.checkDocumentation(context.Background(), t.TempDir(), inventory.Module{}); err == nil || errors.Is(err, failure) {
		t.Fatalf("checkDocumentation() native error = %v", err)
	}
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("# Example\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := runner.checkDocumentation(context.Background(), root, inventory.Module{}); !errors.Is(err, failure) {
		t.Fatalf("checkDocumentation() spelling error = %v", err)
	}
}

func TestCheckDocumentationUsesPinnedSpellingByDefault(t *testing.T) {
	root := spellingRepository(t)
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("# Example\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	runner := Runner{Executor: workspaceExecutor{directory: t.TempDir(), run: func(context.Context, Command) error { return nil }}}
	if err := runner.checkDocumentation(context.Background(), root, inventory.Module{}); err != nil {
		t.Fatalf("checkDocumentation() error = %v", err)
	}
}

func TestRunOperationReportsTimeoutCommandAndExecutionFailures(t *testing.T) {
	failure := errors.New("injected failure")
	tests := []struct {
		name     string
		step     config.Step
		executor Executor
		want     string
	}{
		{"timeout", config.Step{Type: "go-test", Timeout: "forever"}, executorFunction(func(context.Context, Command) error { return nil }), "timeout"},
		{"execution", config.Step{Type: "go-test", Packages: []string{"."}, Count: 1, Timeout: "1m"}, executorFunction(func(context.Context, Command) error { return failure }), "injected failure"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			runner := Runner{Executor: test.executor}
			err := runner.runOperation(context.Background(), t.TempDir(), inventory.Module{Directory: "."}, config.Operation{Gate: "docs", Steps: []config.Step{test.step}})
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("runOperation() error = %v", err)
			}
		})
	}
}

type executorFunction func(context.Context, Command) error

func (function executorFunction) Run(ctx context.Context, command Command) error {
	return function(ctx, command)
}

type fakeCoverageFiles struct {
	file      namedWriteCloser
	createErr error
	openErr   error
	directory string
}

func (fake *fakeCoverageFiles) CreateTemp(directory string) (namedWriteCloser, error) {
	fake.directory = directory
	return fake.file, fake.createErr
}

func (fake *fakeCoverageFiles) Open(string) (io.ReadCloser, error) {
	return io.NopCloser(bytes.NewBufferString("mode: atomic\nexample/file.go:1.1,2.1 1 1\n")), fake.openErr
}

func (*fakeCoverageFiles) Remove(string) error { return nil }

type fakeNamedFile struct {
	name     string
	closeErr error
}

func (*fakeNamedFile) Write(data []byte) (int, error) { return len(data), nil }
func (file *fakeNamedFile) Close() error              { return file.closeErr }
func (file *fakeNamedFile) Name() string              { return file.name }

func TestOperationCommandSupportsSelectorsAndRejectsUnknownTypes(t *testing.T) {
	root := t.TempDir()
	module := inventory.Module{TestTags: []string{"integration"}}
	for name, test := range map[string]struct {
		step config.Step
		want string
	}{
		"benchmark": {
			step: config.Step{Type: "go-test", Packages: []string{"."}, Benchmark: ".", Budget: "250ms", Count: 1, Timeout: "1m"},
			want: "go test -tags=integration . -count=1 -timeout=1m -run=^$ -bench=. -benchmem -benchtime=250ms",
		},
		"fuzz": {
			step: config.Step{Type: "go-test", Packages: []string{"."}, Fuzz: "FuzzInput", Budget: "100x", Count: 1, Timeout: "1m"},
			want: "go test -tags=integration . -count=1 -timeout=1m -run=^$ -fuzz=FuzzInput -fuzztime=100x",
		},
	} {
		t.Run(name, func(t *testing.T) {
			command, err := operationCommand(root, module, test.step)
			if err != nil || command.Name != "go" {
				t.Fatalf("operation command = %#v, %v", command, err)
			}
			if got := strings.Join(append([]string{command.Name}, command.Args...), " "); got != test.want {
				t.Fatalf("operation command = %q, want %q", got, test.want)
			}
		})
	}
	if _, err := operationCommand(root, module, config.Step{Type: "unsupported"}); err == nil {
		t.Fatal("operationCommand() error = nil")
	}
}

func TestSafetyIgnoresTestOnlyUnsafeImports(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "unsafe_test.go"), []byte("package example\n\nimport \"unsafe\"\n\nvar _ = unsafe.Sizeof(0)\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := checkSafety(root); err != nil {
		t.Fatalf("checkSafety() error = %v", err)
	}
}

func TestWalkModuleFilesSkipsVendorAndGitDirectories(t *testing.T) {
	root := t.TempDir()
	for _, directory := range []string{"vendor", ".git"} {
		path := filepath.Join(root, directory)
		if err := os.Mkdir(path, 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(path, "bad.go"), []byte("package ["), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := checkFormatting(root); err != nil {
		t.Fatalf("checkFormatting() error = %v", err)
	}
}

func TestFormattingReportsUnreadableSource(t *testing.T) {
	root := t.TempDir()
	if err := os.Symlink("missing", filepath.Join(root, "missing.go")); err != nil {
		t.Fatal(err)
	}
	if err := checkFormatting(root); err == nil {
		t.Fatal("checkFormatting() error = nil")
	}
}

func TestWalkModuleFilesReportsNestedModuleInspectionFailure(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "nested")
	if err := os.Mkdir(nested, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("go.mod", filepath.Join(nested, "go.mod")); err != nil {
		t.Fatal(err)
	}
	if err := checkFormatting(root); err == nil {
		t.Fatal("checkFormatting() error = nil")
	}
}
