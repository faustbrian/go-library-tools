package gates

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/faustbrian/go-library-tools/v2/internal/inventory"
)

func TestStaticcheckUsesIsolatedBuildForBothGateContracts(t *testing.T) {
	for _, local := range []bool{false, true} {
		for _, directory := range []string{".", "nested"} {
			t.Run(directory+"/local="+boolString(local), func(t *testing.T) {
				root, task := t.TempDir(), t.TempDir()
				target := filepath.Join(root, directory)
				if err := os.MkdirAll(target, 0o700); err != nil {
					t.Fatal(err)
				}
				manifest := []byte("module example.com/target\n\ngo 1.27.0\n")
				if err := os.WriteFile(filepath.Join(target, "go.mod"), manifest, 0o600); err != nil {
					t.Fatal(err)
				}
				var build, analysis string
				executor := workspaceExecutor{directory: task, run: func(ctx context.Context, command Command) error {
					if slices.Contains(command.Args, "honnef.co/go/tools/cmd/staticcheck@v0.8.1") {
						return errors.New("stock importer cannot decode export version 5")
					}
					if len(command.Args) > 0 && command.Args[0] == "build" {
						build = command.Dir
						if !strings.HasPrefix(build, task+string(filepath.Separator)) || command.Env["GOWORK"] != "off" || command.Env["GOFLAGS"] != "" {
							t.Fatal("tool build escaped its owned workspace")
						}
						data, err := os.ReadFile(filepath.Join(build, "go.mod"))
						if err != nil || !bytes.Contains(data, []byte("honnef.co/go/tools v0.8.1")) || !bytes.Contains(data, []byte("golang.org/x/tools v0.50.0")) {
							t.Fatalf("compatible isolated graph missing: %v", err)
						}
						return nil
					}
					if filepath.Base(command.Name) == "staticcheck" {
						if _, overridden := command.Env["GOFLAGS"]; overridden {
							t.Fatal("tool-build flags leaked into target analysis")
						}
						analysis = command.Dir
						if !slices.Equal(command.Args, []string{"./..."}) || command.Env["GOWORK"] != "off" || ctx != t.Context() {
							t.Fatal("analysis lost target scope, workspace isolation or context")
						}
					}
					return nil
				}}
				runner := Runner{Root: root, Executor: executor, Catalog: inventory.Inventory{Modules: []inventory.Module{{Directory: directory, ModulePath: "example.com/target", Gates: map[string]bool{"lint": true}}}}}
				var err error
				if local {
					err = runner.Local(t.Context(), []string{directory})
				} else {
					err = runner.Check(t.Context(), []string{directory})
				}
				if err != nil || build == "" || analysis != target {
					t.Fatalf("Staticcheck contract: build=%q analysis=%q error=%v", build, analysis, err)
				}
				data, err := os.ReadFile(filepath.Join(target, "go.mod"))
				if err != nil || !bytes.Equal(data, manifest) {
					t.Fatal("application graph changed")
				}
				entries, err := os.ReadDir(task)
				if err != nil || len(entries) != 0 {
					t.Fatal("tool artifacts remained after analysis")
				}
			})
		}
	}
}

func TestStaticcheckPropagatesBuildAndAnalysisFailure(t *testing.T) {
	for _, phase := range []string{"build", "analysis", "cancel-build", "cancel-analysis"} {
		t.Run(phase, func(t *testing.T) {
			task := t.TempDir()
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			failure := errors.New("independent phase failure")
			var stages []string
			executor := workspaceExecutor{directory: task, run: func(commandContext context.Context, command Command) error {
				stage := "analysis"
				if len(command.Args) > 0 && command.Args[0] == "build" {
					stage = "build"
				}
				stages = append(stages, stage)
				if phase == "cancel-"+stage {
					cancel()
					return commandContext.Err()
				}
				if phase == stage {
					return failure
				}
				return nil
			}}
			err := (Runner{Executor: executor}).goTool(ctx, io.Discard, "nested", "staticcheck", t.TempDir(), "honnef.co/go/tools/cmd/staticcheck@v0.8.1", "./...")
			want := failure
			if strings.HasPrefix(phase, "cancel-") {
				want = context.Canceled
			}
			if !errors.Is(err, want) || !strings.Contains(err.Error(), "nested staticcheck") {
				t.Fatalf("phase failure not preserved: %v", err)
			}
			if strings.HasSuffix(phase, "build") && !slices.Equal(stages, []string{"build"}) {
				t.Fatalf("analysis ran after failed build: %v", stages)
			}
			entries, err := os.ReadDir(task)
			if err != nil || len(entries) != 0 {
				t.Fatal("owned artifacts leaked on failure")
			}
		})
	}
}

func boolString(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

func TestStaticcheckRequiresOwnedWorkspaceAndRespectsCancellation(t *testing.T) {
	for _, task := range []string{"", "relative", filepath.Join(t.TempDir(), "missing")} {
		called := false
		executor := workspaceExecutor{directory: task, run: func(context.Context, Command) error { called = true; return nil }}
		err := (Runner{Executor: executor}).goTool(t.Context(), io.Discard, ".", "staticcheck", t.TempDir(), "honnef.co/go/tools/cmd/staticcheck@v0.8.1", "./...")
		if err == nil || called {
			t.Fatalf("invalid workspace: error=%v called=%v", err, called)
		}
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	called := false
	executor := workspaceExecutor{directory: t.TempDir(), run: func(context.Context, Command) error { called = true; return nil }}
	if err := (Runner{Executor: executor}).goTool(ctx, io.Discard, ".", "staticcheck", t.TempDir(), "honnef.co/go/tools/cmd/staticcheck@v0.8.1", "./..."); !errors.Is(err, context.Canceled) || called {
		t.Fatalf("already cancelled invocation: error=%v called=%v", err, called)
	}
}

func TestStaticcheckDisabledDoesNotBuildOrAnalyze(t *testing.T) {
	for _, local := range []bool{false, true} {
		executor := workspaceExecutor{directory: t.TempDir(), run: func(_ context.Context, command Command) error {
			if filepath.Base(command.Name) == "staticcheck" || slices.Contains(command.Args, "honnef.co/go/tools/cmd/staticcheck") {
				t.Fatal("disabled analyzer invoked")
			}
			return nil
		}}
		runner := Runner{Root: t.TempDir(), Executor: executor, Catalog: inventory.Inventory{Modules: []inventory.Module{{Directory: ".", Gates: map[string]bool{"lint": false}}}}}
		var err error
		if local {
			err = runner.Local(t.Context(), []string{"."})
		} else {
			err = runner.Check(t.Context(), []string{"."})
		}
		if err != nil {
			t.Fatal(err)
		}
	}
}

// Opt-in because this contract intentionally downloads and builds real tools.
func TestStaticcheckOwnedGraphIntegration(t *testing.T) {
	if os.Getenv("GOLIB_STATICCHECK_INTEGRATION") != "1" {
		t.Skip("real tool integration not requested")
	}
	t.Setenv("GOFLAGS", "-tags=control")
	root := t.TempDir()
	target := filepath.Join(root, "nested")
	if err := os.MkdirAll(filepath.Join(target, "child"), 0o700); err != nil {
		t.Fatal(err)
	}
	write := func(path, value string) {
		t.Helper()
		if err := os.WriteFile(path, []byte(value), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	write(filepath.Join(root, "go.mod"), "module example.com/unselected\n\ngo 1.27.0\n")
	write(filepath.Join(root, "bad.go"), "package unselected\nimport \"strings\"\nfunc Bad() { strings.TrimSpace(\" x \" ) }\n")
	manifest := "module example.com/selected\n\ngo 1.27.0\n"
	write(filepath.Join(target, "go.mod"), manifest)
	write(filepath.Join(target, "child", "clean.go"), "// Package child supplies an analysis fixture.\npackage child\nimport \"sync/atomic\"\nfunc Value() int64 { var v atomic.Int64; return v.Load() }\n")
	var output bytes.Buffer
	executor, cleanup, err := NewProcessExecutor(root, &output, &output)
	if err != nil {
		t.Fatal(err)
	}
	workspace, ok := executor.(taskWorkspace)
	if !ok {
		t.Fatal("process executor lost its task workspace")
	}
	task := workspace.TemporaryDirectory()
	defer func() {
		if err := cleanup(); err != nil {
			t.Error(err)
		}
		if _, err := os.Stat(task); !errors.Is(err, os.ErrNotExist) {
			t.Error("task survived cleanup")
		}
	}()
	run := func(args ...string) error {
		output.Reset()
		return (Runner{Executor: executor}).goTool(t.Context(), io.Discard, "nested", "staticcheck", target, "honnef.co/go/tools/cmd/staticcheck@v0.8.1", args...)
	}
	graph := func() string {
		t.Helper()
		output.Reset()
		if err := executor.Run(t.Context(), Command{Name: "go", Args: []string{"list", "-m", "all"}, Dir: target, Env: map[string]string{"GOWORK": "off"}}); err != nil {
			t.Fatal(err)
		}
		return output.String()
	}
	beforeGraph := graph()
	if err := run("./..."); err != nil {
		t.Fatalf("clean selected target: %v\n%s", err, output.String())
	}
	if err := run("-debug.version"); err != nil {
		t.Fatal(err)
	}
	for _, identity := range []string{"honnef.co/go/tools", "v0.8.1", "golang.org/x/tools", "v0.50.0"} {
		if !strings.Contains(output.String(), identity) {
			t.Fatalf("missing build identity %s: %s", identity, output.String())
		}
	}
	if err := run("-list-checks"); err != nil {
		t.Fatal(err)
	}
	catalogue := output.String()
	output.Reset()
	if err := executor.Run(t.Context(), Command{Name: "go", Args: []string{"run", "honnef.co/go/tools/cmd/staticcheck@v0.8.1", "-list-checks"}, Dir: target, Env: map[string]string{"GOWORK": "off"}}); err != nil {
		t.Fatal(err)
	}
	// Go's module download notices can precede the upstream catalogue.
	stock := output.String()
	start := strings.Index(stock, "S1000 ")
	if start < 0 || catalogue != stock[start:] {
		t.Fatalf("released check catalogue changed\nowned: %s\nstock: %s", catalogue, stock)
	}
	write(filepath.Join(target, "child", "control_test.go"), "//go:build control\n\npackage child\nimport (\"strings\"; \"testing\")\nfunc TestControl(t *testing.T) { strings.TrimSpace(\" x \" ) }\n")
	if err := run("./..."); err == nil || !strings.Contains(output.String(), "SA4017") || !strings.Contains(output.String(), "control_test.go") {
		t.Fatalf("tagged nested test finding lost: %v\n%s", err, output.String())
	}
	write(filepath.Join(target, "staticcheck.conf"), "checks = [\"all\", \"-SA4017\"]\n")
	if err := run("./..."); err != nil {
		t.Fatalf("target configuration ignored: %v\n%s", err, output.String())
	}
	if afterGraph := graph(); afterGraph != beforeGraph {
		t.Fatalf("application graph changed: %q -> %q", beforeGraph, afterGraph)
	}
	data, err := os.ReadFile(filepath.Join(target, "go.mod"))
	if err != nil || string(data) != manifest {
		t.Fatal("application manifest changed")
	}
	if _, err := os.Stat(filepath.Join(target, "go.sum")); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("tool graph wrote application go.sum")
	}
	repository, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	output.Reset()
	if err := (Runner{Executor: executor}).goTool(t.Context(), io.Discard, ".", "staticcheck", repository, "honnef.co/go/tools/cmd/staticcheck@v0.8.1", "./..."); err != nil {
		t.Fatalf("repository analysis: %v\n%s", err, output.String())
	}
	entries, err := os.ReadDir(task)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "staticcheck-") {
			t.Fatal("owned build remained")
		}
	}
}
