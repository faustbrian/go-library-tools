package coverage

import (
	"strings"
	"testing"
)

func TestIncompleteDiagnosticsAcceptExactRemainingBytes(t *testing.T) {
	blocks := map[string]block{
		"example/a.go:1.1,2.1": {packagePath: "example", statements: 1},
	}
	want := "uncovered production blocks:\n  a.go:1.1,2.1\n"
	report := incompleteReport(blocks, []packageFailure{{packagePath: "example"}}, diagnosticLimits{blocks: 1, bytes: 128 + len(want)})
	if report != want {
		t.Fatalf("exact-fit diagnostic = %q, want %q", report, want)
	}
}

func TestIncompleteDiagnosticsShareOneGlobalAllowance(t *testing.T) {
	blocks := map[string]block{
		"example/a/a.go:1.1,2.1":     {packagePath: "example/a", statements: 1},
		"example/b/b.go:1.1,2.1":     {packagePath: "example/b", statements: 1},
		"example/other/c.go:1.1,2.1": {packagePath: "example/other", statements: 1},
	}
	failures := []packageFailure{{packagePath: "example/a"}, {packagePath: "example/b"}}
	for _, limit := range []int{2, 1} {
		report := incompleteReport(blocks, failures, diagnosticLimits{blocks: limit, bytes: 1024})
		if !strings.Contains(report, "example/a is below exact 100%") || !strings.Contains(report, "example/b is below exact 100%") || strings.Contains(report, "example/other") {
			t.Fatal("aggregate diagnostics lost a package or included unexpected evidence")
		}
		if limit == 2 {
			if strings.Count(report, ".go:") != 2 || strings.Contains(report, "omitted") {
				t.Fatal("inclusive global source-record allowance changed diagnostics")
			}
		} else if strings.Count(report, ".go:") != 1 || !strings.Contains(report, "a.go:1.1,2.1") || strings.Contains(report, "b.go:") || !strings.Contains(report, "(1 additional blocks omitted)") {
			t.Fatal("source-record allowance multiplied across failed packages")
		}
	}
	report := incompleteReport(blocks, failures, diagnosticLimits{blocks: 2, bytes: 200})
	if len(report) > 200 || !strings.Contains(report, "additional blocks omitted") || !strings.Contains(report, "additional packages omitted") {
		t.Fatal("aggregate byte allowance failed to bound package and location output")
	}
}

func TestIncompleteDiagnosticsOmitOversizedPackageLabels(t *testing.T) {
	longPackage := "example/" + strings.Repeat("a", 161)
	blocks := map[string]block{
		"example/a/a.go:1.1,2.1":      {packagePath: "example/a", statements: 1},
		longPackage + "/b.go:1.1,2.1": {packagePath: longPackage, statements: 1},
	}
	report := incompleteReport(blocks, []packageFailure{{packagePath: "example/a"}, {packagePath: longPackage}}, diagnosticLimits{blocks: 2, bytes: 1024})
	if !strings.Contains(report, "example/a is below exact 100%") || strings.Contains(report, longPackage) || strings.Contains(report, "b.go:") ||
		!strings.Contains(report, "(1 additional blocks omitted)") || !strings.Contains(report, "(1 additional packages omitted)") {
		t.Fatal("oversized package label escaped admission or lost omission accounting")
	}
}
