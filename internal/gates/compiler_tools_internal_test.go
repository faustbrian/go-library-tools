package gates

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/faustbrian/go-library-tools/v2/internal/inventory"
)

// CC-04: exercise the production scanner owner, not an independently built
// tool, so a compatible test-only graph cannot hide broken gate wiring.
func TestGosecCompatibleGraphIntegration(t *testing.T) {
	if os.Getenv("GOLIB_COMPILER_TOOLS_INTEGRATION") != "1" {
		t.Skip("opt-in real analyzer control")
	}
	root := t.TempDir()
	manifest := "module example.invalid/compiler-control\n\ngo 1.27.0\n"
	for name, data := range map[string]string{
		"go.mod":                    manifest,
		"control.go":                "package control\nimport \"sync/atomic\"\nfunc Value() int64 { var v atomic.Int64; return v.Load() }\n",
		"support/support.go":        "package support\nfunc Value() int { return 2 }\n",
		"testdata/broken/broken.go": "package broken\nimport _ \"example.invalid/missing\"\n",
	} {
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	executor, cleanup, err := NewProcessExecutor(root, io.Discard, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := cleanup(); err != nil {
			t.Error(err)
		}
	})
	var output strings.Builder
	runner := Runner{Root: root, Executor: executor, Output: &output}
	packages, err := runner.gosecPackages(t.Context(), root)
	if err != nil {
		t.Fatal(err)
	}
	args := append([]string{"-nosec-require-rules", "-nosec-require-justification"}, packages...)
	tool := "github.com/securego/gosec/v2/cmd/gosec@" + gosecVersion
	if err := runner.securityTool(t.Context(), &output, ".", "gosec", root, tool, args...); err != nil {
		t.Fatalf("clean source must load under patched compiler: %v", err)
	}
	source := "package control\nimport \"os\"\nfunc Read(path string) ([]byte,error) { return os.ReadFile(path) }\n"
	if err := os.WriteFile(filepath.Join(root, "control.go"), []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	output.Reset()
	err = runner.securityTool(t.Context(), &output, ".", "gosec", root, tool, args...)
	var exit *exec.ExitError
	if err == nil || !errors.As(err, &exit) || !strings.Contains(err.Error(), "gosec-findings") || !strings.Contains(err.Error(), "G304") {
		t.Fatalf("real G304 finding must fail with validated metadata and typed exit: %v", err)
	}
	if strings.Contains(output.String()+err.Error(), "ReadFile(path)") || strings.Contains(output.String()+err.Error(), "control.go") {
		t.Fatal("private scanner source or path escaped")
	}
	if err := os.WriteFile(filepath.Join(root, "control.go"), []byte("package control\nimport _ \"example.invalid/missing\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	err = runner.securityTool(context.Background(), &output, ".", "gosec", root, tool, args...)
	if err == nil || strings.Contains(err.Error(), "gosec-findings") {
		t.Fatalf("unresolved imports must not qualify as completed findings: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil || string(data) != manifest {
		t.Fatal("scanner changed application manifest")
	}
	if _, err := os.Stat(filepath.Join(root, "go.sum")); !os.IsNotExist(err) {
		t.Fatal("scanner created application dependency sums")
	}
}

// CC-05: the generated blocking policy must survive a tool-only importer fix.
func TestOwnedAnalysisCompatibleGraphIntegration(t *testing.T) {
	if os.Getenv("GOLIB_COMPILER_TOOLS_INTEGRATION") != "1" {
		t.Skip("opt-in real analyzer control")
	}
	root := t.TempDir()
	for name, data := range map[string]string{
		"go.mod":     "module example.invalid/analysis-control\n\ngo 1.27.0\n",
		"control.go": "package control\nimport \"sync/atomic\"\nfunc Value() int64 {var v atomic.Int64;return v.Load()}\n",
	} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(data), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	executor, cleanup, err := NewProcessExecutor(root, io.Discard, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := cleanup(); err != nil {
			t.Error(err)
		}
	})
	var privateReport strings.Builder
	workspace, ok := executor.(taskWorkspace)
	if !ok {
		t.Fatal("executor lost task workspace")
	}
	observer := workspaceExecutor{directory: workspace.TemporaryDirectory(), run: func(ctx context.Context, command Command) error {
		if filepath.Base(command.Name) == "golib-analysis" {
			command.Stdout = io.MultiWriter(command.Stdout, &privateReport)
			command.Stderr = io.MultiWriter(command.Stderr, &privateReport)
		}
		return executor.Run(ctx, command)
	}}
	runner := Runner{Root: root, Executor: observer}
	policy, remove, err := runner.createAnalysisConfig()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := remove(); err != nil {
			t.Error(err)
		}
	})
	tool := "github.com/faustbrian/go-analysis/cmd/golib-analysis@" + goAnalysisVersion
	args := []string{"check", "-config", policy, "-root", root, "./..."}
	if err := runner.securityTool(t.Context(), io.Discard, ".", "owned-security-analysis", root, tool, args...); err != nil {
		t.Fatalf("clean analysis must load: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "control.go"), []byte("package control\nimport \"unsafe\"\nfunc Size(v int) uintptr {return unsafe.Sizeof(v)}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	var output strings.Builder
	err = runner.securityTool(t.Context(), &output, ".", "owned-security-analysis", root, tool, args...)
	var exit *exec.ExitError
	if err == nil || !errors.As(err, &exit) || !strings.Contains(privateReport.String(), "security/no-unsafe") {
		t.Fatalf("blocking unsafe policy did not fail: %v", err)
	}
	if strings.Contains(output.String()+err.Error(), "Sizeof(v)") {
		t.Fatal("scanner-controlled source escaped")
	}
}

// CC-02/CC-09: one common lifecycle contract for all rebuilt tools. The
// executor is inert; these are not local process-control tests.
func TestCompilerToolsPreserveIsolationFailureAndPrivacy(t *testing.T) {
	tools := make([]compilerTool, 0, 3)
	tools = append(tools, staticcheckCompilerTool)
	for _, identity := range []string{"github.com/securego/gosec/v2/cmd/gosec@" + gosecVersion, "github.com/faustbrian/go-analysis/cmd/golib-analysis@" + goAnalysisVersion} {
		tool, ok := securityCompilerTool(identity)
		if !ok {
			t.Fatal("missing compiler tool")
		}
		tools = append(tools, tool)
	}
	for _, tool := range tools {
		for _, phase := range []string{"success", "build", "analysis", "cancel-build", "cancel-analysis", "stdout-overflow", "stderr-overflow"} {
			t.Run(tool.name+"/"+phase, func(t *testing.T) {
				task, target := t.TempDir(), t.TempDir()
				ctx, cancel := context.WithCancel(t.Context())
				defer cancel()
				failure := errors.New("independent execution failure")
				var stages []string
				executor := workspaceExecutor{directory: task, run: func(commandContext context.Context, command Command) error {
					stages = append(stages, "build")
					if command.Name != "go" || command.Dir == target || !strings.HasPrefix(command.Dir, task+string(filepath.Separator)) || command.Env["GOWORK"] != "off" || command.Env["GOFLAGS"] != "" || !command.boundedScanner {
						t.Fatal("tool construction lost isolation or bounded output")
					}
					manifest, err := os.ReadFile(filepath.Join(command.Dir, "go.mod"))
					if err != nil || !strings.Contains(string(manifest), tool.module+" "+tool.version) || !strings.Contains(string(manifest), "golang.org/x/tools v0.50.0") {
						t.Fatal("wrong effective build inputs")
					}
					_, _ = io.WriteString(command.Stdout, "private build stdout")
					_, _ = io.WriteString(command.Stderr, "private build stderr")
					switch phase {
					case "build":
						return failure
					case "cancel-build":
						cancel()
						return nil // a successful return must not erase cancellation
					case "stdout-overflow":
						_, _ = io.WriteString(command.Stdout, strings.Repeat("x", maximumSecurityProcessOutput))
					case "stderr-overflow":
						_, _ = io.WriteString(command.Stderr, strings.Repeat("x", maximumSecurityProcessOutput))
					}
					return commandContext.Err()
				}}
				runner := Runner{Executor: executor}
				err := runner.withCompilerTool(ctx, tool, func(binary string) error {
					stages = append(stages, "analysis")
					if filepath.Base(binary) != tool.name || !strings.HasPrefix(binary, task+string(filepath.Separator)) {
						t.Fatal("analysis escaped owned binary")
					}
					if phase == "analysis" {
						return failure
					}
					if phase == "cancel-analysis" {
						cancel()
						return ctx.Err()
					}
					return nil
				})
				switch phase {
				case "success":
					if err != nil {
						t.Fatal(err)
					}
				case "build", "analysis":
					if !errors.Is(err, failure) {
						t.Fatal("underlying failure discarded")
					}
				case "cancel-build", "cancel-analysis":
					if !errors.Is(err, context.Canceled) {
						t.Fatal("cancellation discarded")
					}
				default:
					if err == nil || !strings.Contains(err.Error(), "output exceeded") {
						t.Fatal("build overflow accepted")
					}
				}
				if phase != "success" && phase != "analysis" && phase != "cancel-analysis" && len(stages) != 1 {
					t.Fatal("analysis ran after failed build")
				}
				if err != nil && strings.Contains(err.Error(), "private build") {
					t.Fatal("build diagnostics escaped")
				}
				entries, readErr := os.ReadDir(task)
				if readErr != nil || len(entries) != 0 {
					t.Fatal("owned tool artifacts leaked")
				}
			})
		}
	}
}

type compilerScannerExit struct {
	error
	code int
}

func (exit compilerScannerExit) ExitCode() int { return exit.code }
func (exit compilerScannerExit) Unwrap() error { return exit.error }

func TestGosecDirectExitRequiresTypedCompletedStatus(t *testing.T) {
	for _, test := range []struct {
		name   string
		code   int
		marker bool
		want   bool
	}{
		{name: "direct finding", code: 1, want: true},
		{name: "unrelated exit with wrapper text", code: 2, marker: true},
		{name: "untyped wrapper text", code: 0, marker: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			failure := errors.New("native scanner failure")
			var output strings.Builder
			executor := workspaceExecutor{directory: t.TempDir(), run: func(_ context.Context, command Command) error {
				if command.Name == "go" {
					return nil
				}
				_, _ = command.Stdout.Write(ordinaryGosecReport(root, true, false))
				if test.marker {
					_, _ = io.WriteString(command.Stderr, "exit status 1\n")
				}
				if test.code == 0 {
					return failure
				}
				return compilerScannerExit{error: failure, code: test.code}
			}}
			err := (Runner{Executor: executor}).securityTool(t.Context(), &output, ".", "gosec", root, "github.com/securego/gosec/v2/cmd/gosec@"+gosecVersion, "./")
			if !errors.Is(err, failure) || strings.Contains(err.Error(), "gosec-findings") != test.want {
				t.Fatalf("incorrect direct status classification: %v", err)
			}
		})
	}
}

// CC-10: real analyzers through the actual full/local owners on a small safe
// repository, with no tests, services, mutations or process-interruption probes.
func TestCompilerToolsProductionPipelineIntegration(t *testing.T) {
	if os.Getenv("GOLIB_COMPILER_TOOLS_INTEGRATION") != "1" {
		t.Skip("opt-in real pipeline control")
	}
	root := t.TempDir()
	for name, data := range map[string]string{
		"go.mod":     "module example.invalid/compiler-control\n\ngo 1.27.0\n",
		"control.go": "// Package control provides a compiler compatibility fixture.\npackage control\n\nimport \"sync/atomic\"\n\n// Value returns an atomic value.\nfunc Value() int64 {\n\tvar value atomic.Int64\n\treturn value.Load()\n}\n",
		"LICENSE":    "MIT License\n\nCopyright (c) 2026 Fixture Authors\n\nPermission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the Software), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:\n\nThe above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.\n\nTHE SOFTWARE IS PROVIDED AS IS, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.\n",
	} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(data), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	executor, cleanup, err := NewProcessExecutor(root, io.Discard, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := cleanup(); err != nil {
			t.Error(err)
		}
	})
	var output strings.Builder
	observed := map[string]int{}
	workspace, ok := executor.(taskWorkspace)
	if !ok {
		t.Fatal("executor lost task workspace")
	}
	observer := workspaceExecutor{directory: workspace.TemporaryDirectory(), run: func(ctx context.Context, command Command) error {
		if name := filepath.Base(command.Name); name == "gosec" || name == "golib-analysis" || name == "staticcheck" {
			var identity strings.Builder
			if err := executor.Run(ctx, Command{Name: "go", Args: []string{"version", "-m", command.Name}, Dir: command.Dir, Stdout: &identity, Stderr: io.Discard}); err != nil {
				return err
			}
			if !strings.Contains(identity.String(), "golang.org/x/tools\tv0.50.0") {
				return errors.New("real compiler tool importer identity differs")
			}
		}
		err := executor.Run(ctx, command)
		if err != nil {
			return err
		}
		for _, name := range []string{"gosec", "golib-analysis", "staticcheck"} {
			if filepath.Base(command.Name) == name {
				observed[name]++
			}
		}
		for _, identity := range []string{"golang.org/x/vuln/cmd/govulncheck@", "github.com/google/go-licenses/v2@", "github.com/CycloneDX/cyclonedx-gomod/cmd/cyclonedx-gomod@", "github.com/golangci/golangci-lint/v2/cmd/golangci-lint@", "golang.org/x/exp/cmd/apidiff@", "go.uber.org/nilaway/cmd/nilaway@", "github.com/zricethezav/gitleaks/v8@"} {
			if len(command.Args) > 1 && strings.HasPrefix(command.Args[1], identity) {
				observed[identity]++
			}
		}
		return nil
	}}
	runner := Runner{Root: root, Output: &output, Executor: observer, Catalog: inventory.Inventory{GoVersion: "1.27.2", Repository: "example.invalid/compiler-control", Modules: []inventory.Module{{Directory: ".", ModulePath: "example.invalid/compiler-control", Gates: map[string]bool{"lint": true, "security": true, "api_compatibility": true}}}}}
	if err := runner.API(t.Context(), []string{"."}, true); err != nil {
		t.Fatalf("real API baseline generation: %v", err)
	}
	for _, args := range [][]string{{"init", "-q"}, {"add", "go.mod", "control.go", "LICENSE", "api/baseline.txt"}, {"-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "commit", "-qm", "test: add compiler fixture"}} {
		if err := executor.Run(t.Context(), Command{Name: "git", Args: args, Dir: root}); err != nil {
			t.Fatal(err)
		}
	}
	if err := runner.Local(t.Context(), []string{"."}); err != nil {
		t.Fatalf("real local pipeline: %v\n%s", err, output.String())
	}
	nilaway := "go.uber.org/nilaway/cmd/nilaway@"
	if observed[nilaway] != 0 {
		t.Fatal("Local ran full-only advisory")
	}
	if err := runner.Check(t.Context(), []string{"."}); err != nil {
		t.Fatalf("real full pipeline: %v\n%s", err, output.String())
	}
	for _, name := range []string{"gosec", "golib-analysis", "staticcheck", "golang.org/x/vuln/cmd/govulncheck@", "github.com/google/go-licenses/v2@", "github.com/CycloneDX/cyclonedx-gomod/cmd/cyclonedx-gomod@", "github.com/golangci/golangci-lint/v2/cmd/golangci-lint@", "golang.org/x/exp/cmd/apidiff@", "github.com/zricethezav/gitleaks/v8@", nilaway} {
		if observed[name] == 0 {
			t.Fatalf("real pipeline did not complete %s", name)
		}
	}
	t.Log("production full/local/security/API paths completed all selected real analyzers")
}

func TestCompilerToolsRepositorySecurityIntegration(t *testing.T) {
	if os.Getenv("GOLIB_COMPILER_REPOSITORY_INTEGRATION") != "1" {
		t.Skip("opt-in current repository security scan")
	}
	directory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	root := filepath.Clean(filepath.Join(directory, "..", ".."))
	executor, cleanup, err := NewProcessExecutor(root, io.Discard, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := cleanup(); err != nil {
			t.Error(err)
		}
	})
	runner := Runner{Root: root, Executor: executor, Catalog: inventory.Inventory{Repository: "github.com/faustbrian/go-library-tools"}}
	if err := runner.runSecurity(t.Context(), io.Discard, root, inventory.Module{Directory: ".", ModulePath: "github.com/faustbrian/go-library-tools/v2"}); err != nil {
		t.Fatalf("current repository security pipeline: %v", err)
	}
}
