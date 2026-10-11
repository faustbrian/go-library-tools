package config_test

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/faustbrian/go-library-tools/v2/internal/config"
)

func TestLocalTestBudgetPreservesAdmissionBoundaries(t *testing.T) {
	for _, scenario := range []struct {
		name     string
		gate     string
		steps    int
		selector string
		invalid  bool
	}{
		{"non-local operation", "test", 1, "^TestContract$", true},
		{"empty operation", "local-test", 0, "^TestContract$", true},
		{"one step", "local-test", 1, "^TestContract$", false},
		{"eight steps", "local-test", 8, "^TestContract$", false},
		{"nine steps", "local-test", 9, "^TestContract$", true},
		{"255-byte selector", "local-test", 1, "^Test" + strings.Repeat("A", 249) + "$", false},
		{"256-byte selector", "local-test", 1, "^Test" + strings.Repeat("A", 250) + "$", false},
		{"257-byte selector", "local-test", 1, "^Test" + strings.Repeat("A", 251) + "$", true},
		{"literal terminal dollar after repetition", "local-test", 1, `^Test.*\$`, true},
		{"literal terminal dollar after group", "local-test", 1, `^Test(Alpha|Beta)\$`, true},
		{"escaped dollar with end anchor", "local-test", 1, `^Test(Alpha|Beta)\$$`, false},
		{"escaped dollar before repetition and end anchor", "local-test", 1, `^Test\$.*$`, false},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			operation := config.Operation{Module: ".", Gate: scenario.gate}
			var policy strings.Builder
			fmt.Fprintf(&policy, "schema_version: 1\ntool_version: v2.0.0\noperations:\n  - module: .\n    gate: %s\n    steps:\n", scenario.gate)
			for range scenario.steps {
				operation.Steps = append(operation.Steps, config.Step{Type: "go-test", Packages: []string{"."}, Run: scenario.selector, Count: 1, Timeout: "1s"})
				fmt.Fprintf(&policy, "      - type: go-test\n        packages: [.]\n        run: '%s'\n        count: 1\n        timeout: 1s\n", scenario.selector)
			}
			budget, err := operation.LocalTestBudget()
			if (err != nil) != scenario.invalid {
				t.Fatalf("direct local-test admission: budget=%v, error=%v", budget, err)
			}
			if scenario.invalid && budget != 0 || !scenario.invalid && budget != time.Duration(scenario.steps)*time.Second {
				t.Fatalf("direct local-test budget=%v", budget)
			}
			// Load admits other campaign gates deliberately. Only local-test
			// operations must agree with this method's local admission contract.
			if scenario.gate != "local-test" {
				return
			}
			root := t.TempDir()
			write(t, filepath.Join(root, ".golib.yaml"), policy.String())
			loaded, err := config.Load(root)
			if (err != nil) != scenario.invalid {
				t.Fatalf("loaded local-test admission: error=%v", err)
			}
			if err == nil && (len(loaded.Operations) != 1 || len(loaded.Operations[0].Steps) != scenario.steps || loaded.Operations[0].Steps[0].Run != scenario.selector) {
				t.Fatal("loaded local-test operation changed")
			}
		})
	}
}

func TestLocalTestBudgetAccumulatesDeadlinesWithoutOverflowAdmission(t *testing.T) {
	for _, scenario := range []struct {
		name     string
		timeouts []string
		want     time.Duration
		invalid  bool
	}{
		{"one exact ceiling", []string{"3m"}, 3 * time.Minute, false},
		{"combined exact ceiling", []string{"2m", "1m"}, 3 * time.Minute, false},
		{"positive remaining budget exceeded", []string{"2m", "61s"}, 0, true},
		{"no remaining budget", []string{"2m", "1m", "1ns"}, 0, true},
		{"negative remaining budget cannot admit another step", []string{"181s", "1ns"}, 0, true},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			operation := config.Operation{Module: ".", Gate: "local-test"}
			var policy strings.Builder
			policy.WriteString("schema_version: 1\ntool_version: v2.0.0\noperations:\n  - module: .\n    gate: local-test\n    steps:\n")
			for _, timeout := range scenario.timeouts {
				operation.Steps = append(operation.Steps, config.Step{Type: "go-test", Packages: []string{"."}, Run: "^TestContract$", Count: 1, Timeout: timeout})
				fmt.Fprintf(&policy, "      - type: go-test\n        packages: [.]\n        run: '^TestContract$'\n        count: 1\n        timeout: %s\n", timeout)
			}
			budget, err := operation.LocalTestBudget()
			if (err != nil) != scenario.invalid || budget != scenario.want {
				t.Fatalf("cumulative budget=%v, error=%v, want=%v refusal=%v", budget, err, scenario.want, scenario.invalid)
			}
			root := t.TempDir()
			write(t, filepath.Join(root, ".golib.yaml"), policy.String())
			if _, err := config.Load(root); (err != nil) != scenario.invalid {
				t.Fatalf("loaded cumulative budget admission=%v", err)
			}
		})
	}
}
