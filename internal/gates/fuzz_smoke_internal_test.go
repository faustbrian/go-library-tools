package gates

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/faustbrian/go-library-tools/v2/internal/config"
	"github.com/faustbrian/go-library-tools/v2/internal/inventory"
)

func TestFuzzSmokeSelectsRuntimeAndTestDependencyInputs(t *testing.T) {
	root := t.TempDir()
	module := inventory.Module{Directory: ".", ModulePath: "example"}
	runner := Runner{Root: root, Catalog: inventory.Inventory{Modules: []inventory.Module{
		module, {Directory: "adapter", ModulePath: "example/adapter"},
	}}}
	graph := map[string]fuzzPackage{
		"example/parser": {ImportPath: "example/parser", Dir: filepath.Join(root, "parser"),
			Imports: []string{"example/helper", "example/adapter"}, TestImports: []string{"example/testhelper"},
			XTestImports: []string{"example/externaltesthelper"}, EmbedFiles: []string{"fixture.json"},
			EmbedPatterns: []string{"fixtures/*.json"}},
		"example/helper":             {ImportPath: "example/helper", Dir: filepath.Join(root, "helper"), Imports: []string{"example/parser"}, TestImports: []string{"unused/dependency/tests"}, TestEmbedFiles: []string{"test-only.json"}, TestEmbedPatterns: []string{"testing/*.json"}},
		"example/testhelper":         {ImportPath: "example/testhelper", Dir: filepath.Join(root, "testhelper")},
		"example/externaltesthelper": {ImportPath: "example/externaltesthelper", Dir: filepath.Join(root, "externaltesthelper")},
		"example/adapter": {ImportPath: "example/adapter", Dir: filepath.Join(root, "adapter"), Module: &struct {
			Path string `json:"Path"`
		}{Path: "example/adapter"}},
	}
	for _, scenario := range []struct {
		name, input string
		want        bool
	}{
		{"production dependency", "helper/value.go", true},
		{"internal test dependency", "testhelper/value.go", true},
		{"dependency own tests", "helper/value_test.go", false},
		{"dependency test helper own tests", "testhelper/value_test.go", false},
		{"dependency test-only embed", "helper/test-only.json", false},
		{"deleted dependency test-only embed", "helper/testing/deleted.json", false},
		{"target tests", "parser/parser_test.go", true},
		{"assembly input", "helper/value.s", true},
		{"cgo input", "helper/value.c", true},
		{"header input", "helper/value.h", true},
		{"native object input", "helper/value.syso", true},
		{"external test dependency", "externaltesthelper/value.go", true},
		{"testdata", "helper/testdata/saved-crash", true},
		{"embedded file", "parser/fixture.json", true},
		{"deleted embedded file", "parser/fixtures/deleted.json", true},
		{"newline source", "helper/value\n.go", true},
		{"unrelated source", "other/value.go", false},
		{"unrelated documentation", "README.md", false},
		{"path prefix collision", "helper-other/value.go", false},
		{"changed module dependencies", "go.mod", true},
		{"changed owned dependency module", "adapter/go.mod", true},
		{"changed configured targets", ".golib.yaml", true},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			got, err := runner.fuzzTargetAffected(t.Context(), module, "example/parser", graph, []string{scenario.input})
			if err != nil || got != scenario.want {
				t.Fatalf("affected = %v, error = %v, want %v", got, err, scenario.want)
			}
		})
	}
	delete(graph, "example/helper")
	for _, change := range []string{"README.md", ".golib.yaml"} {
		if _, err := runner.fuzzTargetAffected(t.Context(), module, "example/parser", graph, []string{change}); err == nil {
			t.Fatal("missing dependency was silently omitted")
		}
	}
}

func TestFuzzSmokeCommandBoundsAndFixedDiagnostics(t *testing.T) {
	for _, scenario := range []struct {
		name, output string
		failure      error
		wantError    bool
	}{
		{"exact bound", "12345678", nil, false},
		{"overflow", "123456789", nil, true},
		{"native failure", "private-marker", errors.New("private-marker"), true},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			runner := Runner{Executor: executorFunction(func(_ context.Context, command Command) error {
				if !command.boundedScanner || command.Stdout == nil || command.Stderr == nil {
					t.Fatal("command output/process boundary is not bounded")
				}
				_, _ = io.WriteString(command.Stdout, scenario.output)
				return scenario.failure
			})}
			output, err := runner.fuzzCommand(t.Context(), Command{Name: "go"}, 8)
			if (err != nil) != scenario.wantError || (err != nil && strings.Contains(err.Error(), "private-marker")) {
				t.Fatalf("bounded command output = %q, error = %v", output, err)
			}
		})
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	runner := Runner{Executor: executorFunction(func(context.Context, Command) error {
		t.Fatal("cancelled command executed")
		return nil
	})}
	if _, err := runner.fuzzCommand(ctx, Command{}, 8); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation lost: %v", err)
	}
}

func TestFuzzSmokeRejectsIncompletePackageDiscovery(t *testing.T) {
	root := t.TempDir()
	for _, data := range []string{"", "{}", "{", "[]",
		`{"ImportPath":"example/parser","Dir":".","Error":null}`,
		`{"ImportPath":"example/parser","Dir":"/tmp","Error":{"Err":"private-marker"}}`,
		`{"ImportPath":"example/parser","Dir":"/tmp","DepsErrors":[{}]}`,
		`{"ImportPath":"example/parser","Dir":"/tmp"}{"ImportPath":"example/parser","Dir":"/tmp"}`,
	} {
		runner := Runner{Root: root, Executor: executorFunction(func(_ context.Context, command Command) error {
			_, _ = io.WriteString(command.Stdout, data)
			return nil
		})}
		if _, err := runner.fuzzPackageGraph(t.Context(), inventory.Module{Directory: "."}, []string{"./parser"}); err == nil || strings.Contains(err.Error(), "private-marker") {
			t.Fatalf("package graph accepted invalid evidence or exposed native diagnostics: %v", err)
		}
	}
}

func TestFuzzSmokeRefusesMissingAmbiguousAndFailedTargets(t *testing.T) {
	module := inventory.Module{Directory: "."}
	step := config.Step{Type: "go-test", Packages: []string{"./parser"}, Fuzz: "^FuzzParse$", Budget: "10m", Timeout: "20m", Count: 100}
	for _, scenario := range []struct {
		name, listed string
		failure      bool
		wantError    bool
	}{
		{"one target", "FuzzParse\n", false, false},
		{"no target", "ok example/parser\n", false, true},
		{"duplicate targets", "FuzzParse\nFuzzParse\n", false, true},
		{"failure", "FuzzParse\n", true, true},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			active := 0
			runner := Runner{Root: t.TempDir(), Executor: executorFunction(func(ctx context.Context, command Command) error {
				deadline, exists := ctx.Deadline()
				if !exists || time.Until(deadline) > time.Minute || time.Until(deadline) < 30*time.Second {
					t.Fatal("target wall budget is unbounded or cannot admit the observed cold compilation")
				}
				if slices.Contains(command.Args, "-list=^FuzzParse$") {
					_, _ = io.WriteString(command.Stdout, scenario.listed)
					return nil
				}
				active++
				for _, expected := range []string{"-fuzztime=2s", "-timeout=15s", "-count=1", "-parallel=2", "-p=1"} {
					if !slices.Contains(command.Args, expected) {
						t.Errorf("missing smoke control %s", expected)
					}
				}
				if scenario.failure {
					return errors.New("private-marker")
				}
				return nil
			})}
			err := runner.fuzzSmokeStep(t.Context(), module, step)
			if (err != nil) != scenario.wantError || (err != nil && strings.Contains(err.Error(), "private-marker")) {
				t.Fatalf("target outcome = %v", err)
			}
			if (scenario.listed != "FuzzParse\n" && active != 0) || active > 1 {
				t.Fatalf("active target invocations = %d", active)
			}
		})
	}
}

func TestFuzzSmokeNeverExtendsEarlierParentDeadline(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()
	parentDeadline, _ := ctx.Deadline()
	executions := 0
	runner := Runner{Root: t.TempDir(), Executor: executorFunction(func(commandContext context.Context, command Command) error {
		deadline, exists := commandContext.Deadline()
		if !exists || !deadline.Equal(parentDeadline) {
			t.Fatal("target extended the total/parent deadline")
		}
		if slices.Contains(command.Args, "-list=^FuzzParse$") {
			_, _ = io.WriteString(command.Stdout, "FuzzParse\n")
			return nil
		}
		executions++
		return context.DeadlineExceeded
	})}
	step := config.Step{Type: "go-test", Packages: []string{"."}, Fuzz: "^FuzzParse$"}
	err := runner.fuzzSmokeStep(ctx, inventory.Module{Directory: "."}, step)
	if !errors.Is(err, context.DeadlineExceeded) || executions != 1 {
		t.Fatalf("expired target outcome = %v, executions = %d", err, executions)
	}
	cancel()
	if err := runner.fuzzSmokeStep(ctx, inventory.Module{Directory: "."}, step); !errors.Is(err, context.Canceled) || executions != 1 {
		t.Fatalf("cancelled target outcome = %v, executions = %d", err, executions)
	}
}

func TestFuzzSmokeTotalTargetBudgetFailsBeforeExecution(t *testing.T) {
	root := localCoverageFixture(t, "# Fixture\n")
	module := inventory.Module{Directory: ".", ModulePath: "example"}
	operation := config.Operation{Module: ".", Gate: "fuzz"}
	for i := range fuzzSmokeMaximumTargets + 1 {
		operation.Steps = append(operation.Steps, config.Step{Type: "go-test", Packages: []string{"."}, Fuzz: "Fuzz" + strings.Repeat("X", i+1)})
	}
	runner := Runner{Root: root, Catalog: inventory.Inventory{Modules: []inventory.Module{module}},
		Policy: config.Config{Operations: []config.Operation{operation}}, Output: &bytes.Buffer{}}
	runner.Executor = workspaceExecutor{directory: t.TempDir(), run: func(ctx context.Context, command Command) error {
		if command.Name == "git" && slices.Contains(command.Args, "rev-parse") {
			_, _ = io.WriteString(command.Stdout, strings.Repeat("b", 40)+"\n")
		}
		if command.Name == "git" && slices.Contains(command.Args, "diff") {
			_, _ = io.WriteString(command.Stdout, "example.go\x00")
		}
		if command.Name == "go" && slices.Contains(command.Args, "list") {
			deadline, ok := ctx.Deadline()
			if !ok || time.Until(deadline) > fuzzSmokeTotalBudget {
				t.Fatal("missing total smoke budget")
			}
			return json.NewEncoder(command.Stdout).Encode(fuzzPackage{ImportPath: "example", Dir: root})
		}
		for _, argument := range command.Args {
			if strings.HasPrefix(argument, "-fuzz=") {
				t.Fatal("oversized target plan executed partially")
			}
		}
		return nil
	}}
	if err := runner.LocalChanges(t.Context(), []string{"."}, strings.Repeat("a", 40)); err == nil || !strings.Contains(err.Error(), "target limit") {
		t.Fatalf("oversized smoke plan = %v", err)
	}
}
