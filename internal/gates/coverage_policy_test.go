package gates_test

import (
	"bytes"
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/faustbrian/go-library-tools/v2/internal/config"
	"github.com/faustbrian/go-library-tools/v2/internal/gates"
	"github.com/faustbrian/go-library-tools/v2/internal/inventory"
)

func TestCoverageEvidenceRetainsCountsAndExecutionContract(t *testing.T) {
	for _, aggregate := range []bool{false, true} {
		t.Run(map[bool]string{false: "standalone", true: "check"}[aggregate], func(t *testing.T) {
			executor := &recordingExecutor{task: t.TempDir(), coverageProfile: "mode: atomic\nexample/file.go:1.1,2.1 2 1\nexample/file.go:3.1,4.1 1 0\n"}
			var output bytes.Buffer
			runner := gates.Runner{Root: fixture(t), Policy: config.Config{Coverage: config.Coverage{Modules: []config.CoverageModule{{Module: ".", Mode: "evidence"}}}}, Catalog: inventory.Inventory{Modules: []inventory.Module{{
				Directory: ".", TestTags: []string{"integration"}, Gates: map[string]bool{"coverage": true},
				Packages: []inventory.Package{{ImportPath: "example", CoverageRequired: true}},
			}}}, Executor: executor, Output: &output}
			var err error
			if aggregate {
				err = runner.Check(context.Background(), []string{"."})
			} else {
				err = runner.Coverage(context.Background(), []string{"."})
			}
			if err != nil {
				t.Fatalf("evidence coverage: %v", err)
			}
			if !strings.Contains(output.String(), "example 2/3 statements\n") || !strings.Contains(output.String(), "coverage evidence collected") || strings.Contains(output.String(), "exact 100%") {
				t.Fatalf("misleading or missing evidence output: %q", output.String())
			}
			coverageCommands := 0
			for _, command := range executor.commands {
				if strings.Contains(command, " -coverprofile=") {
					coverageCommands++
					if !strings.Contains(command, "-tags=integration ./... -count=1 -timeout=20m -covermode=atomic -coverpkg=example -coverprofile=") {
						t.Fatalf("changed instrumentation: %q", command)
					}
				}
			}
			if coverageCommands != 1 {
				t.Fatalf("coverage runs = %d", coverageCommands)
			}
			for _, profile := range executor.coveragePaths {
				if _, err := os.Stat(profile); !errors.Is(err, os.ErrNotExist) {
					t.Fatalf("profile retained: %v", err)
				}
			}
		})
	}
}

func TestCoverageEvidenceDoesNotChangeOtherModuleAcceptance(t *testing.T) {
	executor := &recordingExecutor{task: t.TempDir(), coverageProfiles: map[string]string{
		"example":        "mode: atomic\nexample/file.go:1.1,2.1 1 1\nexample/file.go:3.1,4.1 1 0\n",
		"example/nested": "mode: atomic\nexample/nested/file.go:1.1,2.1 1 1\nexample/nested/file.go:3.1,4.1 1 0\n",
	}}
	var output bytes.Buffer
	runner := gates.Runner{Root: fixture(t), Policy: config.Config{Coverage: config.Coverage{Modules: []config.CoverageModule{{Module: ".", Mode: "evidence"}}}}, Catalog: inventory.Inventory{Modules: []inventory.Module{
		{Directory: ".", Gates: map[string]bool{"coverage": true}, Packages: []inventory.Package{{ImportPath: "example", CoverageRequired: true}}},
		{Directory: "nested", Gates: map[string]bool{"coverage": true}, Packages: []inventory.Package{{ImportPath: "example/nested", CoverageRequired: true}}},
	}}, Executor: executor, Output: &output}
	if err := runner.Coverage(context.Background(), []string{".", "nested"}); err == nil || !strings.Contains(err.Error(), "example/nested is below exact") {
		t.Fatalf("undeclared nested module lost exact enforcement: %v", err)
	}
	if !strings.Contains(output.String(), "example 1/2 statements") || !strings.Contains(output.String(), "coverage evidence collected") || strings.Contains(output.String(), "all production packages have exact") {
		t.Fatalf("mixed module output: %q", output.String())
	}
	for _, profile := range executor.coveragePaths {
		if _, err := os.Stat(profile); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("mixed failure retained profile: %v", err)
		}
	}
}
