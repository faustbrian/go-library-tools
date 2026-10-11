package gates_test

import (
	"bytes"
	"context"
	"errors"
	"go/format"
	"io"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/faustbrian/go-library-tools/v2/internal/config"
	"github.com/faustbrian/go-library-tools/v2/internal/gates"
	"github.com/faustbrian/go-library-tools/v2/internal/inventory"
)

type localTestExecutor struct {
	commands        []gates.Command
	names           string
	failure         error
	deadline        bool
	executionOutput string
}

func (executor *localTestExecutor) Run(ctx context.Context, command gates.Command) error {
	executor.commands = append(executor.commands, command)
	if command.Name != "go" || len(command.Args) == 0 || command.Args[0] != "test" {
		return nil
	}
	_, executor.deadline = ctx.Deadline()
	if executor.failure != nil {
		return executor.failure
	}
	for _, argument := range command.Args {
		if strings.HasPrefix(argument, "-list=") {
			_, err := io.WriteString(command.Stdout, executor.names)
			return err
		}
	}
	if command.Stdout != nil && executor.executionOutput != "" {
		_, _ = io.WriteString(command.Stdout, executor.executionOutput)
	}
	return nil
}

func TestLocalRiskExecutionRefusesExcessOutput(t *testing.T) {
	executor := &localTestExecutor{names: "TestContract\n", executionOutput: strings.Repeat("x", (4<<20)+1)}
	runner := gates.Runner{
		Root: fixture(t), Catalog: inventory.Inventory{Modules: []inventory.Module{{Directory: "."}}},
		Policy: config.Config{Operations: []config.Operation{{Module: ".", Gate: "local-test", Steps: []config.Step{{Type: "go-test", Packages: []string{"."}, Run: "^TestContract$", Count: 1, Timeout: "1m"}}}}}, Executor: executor,
	}
	if err := runner.Local(t.Context(), []string{"."}); err == nil || !strings.Contains(err.Error(), "output budget") {
		t.Fatalf("excessive execution output: %v", err)
	}
}

func TestLocalRiskSelectorCannotEscapeTestScope(t *testing.T) {
	executor := &localTestExecutor{names: "TestContract\nFuzzOther\n"}
	runner := gates.Runner{
		Root: fixture(t), Catalog: inventory.Inventory{Modules: []inventory.Module{{Directory: "."}}},
		Policy: config.Config{Operations: []config.Operation{{Module: ".", Gate: "local-test", Steps: []config.Step{{Type: "go-test", Packages: []string{"."}, Run: "^TestContract$|^FuzzOther$", Count: 1, Timeout: "1m"}}}}}, Executor: executor,
	}
	if err := runner.Local(t.Context(), []string{"."}); err == nil {
		t.Fatal("escaping selector was admitted")
	}
	for _, command := range executor.commands {
		if command.Name == "go" && slices.Contains(command.Args, "test") {
			t.Fatal("escaping selector reached Go execution")
		}
	}
}

func TestLocalRunsOnlyDeclaredFocusedRiskTests(t *testing.T) {
	for _, scenario := range []struct {
		name, names string
		failure     error
		wantFailure bool
	}{
		{"matched", "TestContract\nok example 0.01s\n", nil, false},
		{"missing", "ok example 0.01s\n", nil, true},
		{"other test", "TestOtherContract\n", nil, true},
		{"failed discovery", "", errors.New("command failed"), true},
		{"cancelled", "", context.Canceled, true},
		{"deadline", "", context.DeadlineExceeded, true},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			root := fixture(t)
			executor := &localTestExecutor{names: scenario.names, failure: scenario.failure}
			runner := gates.Runner{
				Root: root, Catalog: inventory.Inventory{Modules: []inventory.Module{{Directory: "."}, {Directory: "nested"}}},
				Policy: config.Config{Operations: []config.Operation{
					{Module: ".", Gate: "local-test", Steps: []config.Step{{Type: "go-test", Packages: []string{"."}, Run: "^TestContract$", Count: 1, Timeout: "1m"}}},
					{Module: ".", Gate: "test", Steps: []config.Step{{Type: "make", Makefile: "verification/package.mk", Target: "stress", Timeout: "20m"}}},
					{Module: "nested", Gate: "local-test", Steps: []config.Step{{Type: "go-test", Packages: []string{"."}, Run: "^TestNested$", Count: 1, Timeout: "1m"}}},
				}}, Executor: executor,
			}
			err := runner.Local(t.Context(), []string{"."})
			if (err != nil) != scenario.wantFailure {
				t.Fatalf("risk check error = %v", err)
			}
			var testCommands []gates.Command
			for _, command := range executor.commands {
				if command.Name == "make" || command.Dir != root {
					t.Fatal("unselected stress or nested operation ran")
				}
				if len(command.Args) > 0 && command.Args[0] == "test" {
					testCommands = append(testCommands, command)
				}
			}
			if len(testCommands) == 0 || !executor.deadline {
				t.Fatal("declared focused risk test was skipped or unbounded")
			}
			if !scenario.wantFailure && (len(testCommands) != 2 || !slices.Contains(testCommands[1].Args, "-run=^TestContract$") || !slices.Contains(testCommands[1].Args, "-count=1")) {
				t.Fatalf("wrong focused execution: %#v", testCommands)
			}
		})
	}
}

func TestLocalRefusesCombinedRiskBudgetBeforeExecuting(t *testing.T) {
	root := fixture(t)
	executor := &localTestExecutor{names: "TestContract\n"}
	runner := gates.Runner{
		Root: root, Catalog: inventory.Inventory{Modules: []inventory.Module{{Directory: "."}, {Directory: "nested"}}},
		Policy: config.Config{Operations: []config.Operation{
			{Module: ".", Gate: "local-test", Steps: []config.Step{{Type: "go-test", Packages: []string{"."}, Run: "^TestContract$", Count: 1, Timeout: "2m"}}},
			{Module: "nested", Gate: "local-test", Steps: []config.Step{{Type: "go-test", Packages: []string{"."}, Run: "^TestContract$", Count: 1, Timeout: "2m"}}},
		}}, Executor: executor,
	}
	if err := runner.Local(t.Context(), []string{".", "nested"}); err == nil || !strings.Contains(err.Error(), "local-test budget") || len(executor.commands) != 0 {
		t.Fatalf("combined budget was not refused before execution: %v, %#v", err, executor.commands)
	}
}

func TestCheckRetainsDeclaredFocusedRiskTests(t *testing.T) {
	for _, names := range []string{"TestContract\n", "TestOther\n"} {
		t.Run(strings.TrimSpace(names), func(t *testing.T) {
			executor := &localTestExecutor{names: names}
			runner := gates.Runner{
				Root: fixture(t), Catalog: inventory.Inventory{Modules: []inventory.Module{{Directory: "."}}},
				Policy: config.Config{Operations: []config.Operation{{Module: ".", Gate: "local-test", Steps: []config.Step{{Type: "go-test", Packages: []string{"."}, Run: "^TestContract$", Count: 1, Timeout: "1m"}}}}}, Executor: executor,
			}
			err := runner.Check(t.Context(), []string{"."})
			if (err != nil) != (names != "TestContract\n") || !executor.deadline {
				t.Fatalf("full risk check: %v, deadline=%v", err, executor.deadline)
			}
			count := 0
			for _, command := range executor.commands {
				if command.Name == "go" && slices.Contains(command.Args, "-run=^TestContract$") {
					count++
				}
			}
			want := 2
			if names != "TestContract\n" {
				want = 1
			}
			if count != want {
				t.Fatalf("full check risk commands=%d, want=%d", count, want)
			}
		})
	}
}

func TestLocalNativeRiskTestsUseRaceBuildAndRejectEmptySelection(t *testing.T) {
	root := fixture(t)
	for name, source := range map[string]string{
		"enabled_test.go":  "//go:build race\n\npackage example\n\nconst raceBuild = true\n",
		"disabled_test.go": "//go:build !race\n\npackage example\n\nconst raceBuild = false\n",
		"contract_test.go": "package example\n\nimport \"testing\"\n\nfunc TestContract(t *testing.T) { if !raceBuild { t.Fatal(\"named contract requires race build\") } }\n\nfunc TestUnrelated(t *testing.T) { t.Fatal(\"unrelated test must not run\") }\n",
	} {
		formatted, err := format.Source([]byte(source))
		if err != nil {
			t.Fatal(err)
		}
		write(t, filepath.Join(root, name), string(formatted))
	}
	var output bytes.Buffer
	executor, cleanup, err := gates.NewProcessExecutor(root, &output, &output)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := cleanup(); err != nil {
			t.Error(err)
		}
	})
	for _, scenario := range []struct {
		name, race, run string
		wantFailure     bool
	}{
		{"race build", "true", "^TestContract$", false},
		{"actual contract failure", "false", "^TestContract$", true},
		{"empty selection", "true", "^TestAbsent$", true},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			write(t, filepath.Join(root, ".golib.yaml"), "schema_version: 1\ntool_version: v2.0.0\noperations:\n  - module: .\n    gate: local-test\n    steps:\n      - type: go-test\n        packages: [.]\n        run: '"+scenario.run+"'\n        race: "+scenario.race+"\n        timeout: 3m\n")
			policy, err := config.Load(root)
			if err != nil {
				t.Fatal(err)
			}
			runner := gates.Runner{Root: root, Catalog: inventory.Inventory{Modules: []inventory.Module{{Directory: "."}}}, Policy: policy, Executor: executor}
			ctx, cancel := context.WithTimeout(t.Context(), 3*time.Minute)
			defer cancel()
			err = runner.Local(ctx, []string{"."})
			if (err != nil) != scenario.wantFailure {
				t.Fatalf("native risk check: %v; %s", err, output.String())
			}
			if scenario.name == "empty selection" && !strings.Contains(err.Error(), "no matching tests") {
				t.Fatalf("empty test selection failed for a different reason: %v", err)
			}
		})
	}
}
