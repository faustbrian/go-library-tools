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
