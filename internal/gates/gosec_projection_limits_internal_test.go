package gates

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

func projectionReport(t *testing.T, root string, finding, loading bool) map[string]any {
	t.Helper()
	var report map[string]any
	if err := json.Unmarshal(ordinaryGosecReport(root, finding, loading), &report); err != nil {
		t.Fatal(err)
	}
	return report
}

func projectReport(t *testing.T, report map[string]any, root string) string {
	t.Helper()
	data, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	return gosecFailureMetadata(data, root)
}

func TestGosecProjectionStatisticBoundaries(t *testing.T) {
	for _, field := range []string{"files", "lines", "nosec"} {
		for _, test := range []struct {
			name  string
			value any
			valid bool
		}{
			{"missing", nil, false}, {"negative", -1, false},
			{"zero", 0, true}, {"exact maximum", 100_000_000, true},
			{"excess maximum", 100_000_001, false},
		} {
			t.Run(field+"/"+test.name, func(t *testing.T) {
				root := t.TempDir()
				report := projectionReport(t, root, false, true)
				report["Stats"].(map[string]any)[field] = test.value
				want := "gosec-tool-or-report-failure"
				if test.valid {
					want = "gosec-loading-failure findings=0 loading_packages=1 loading_errors=1"
				}
				if got := projectReport(t, report, root); got != want {
					t.Fatalf("statistic projection = %q, want %q", got, want)
				}
			})
		}
	}
}

func TestGosecProjectionLoadingEntryBoundaries(t *testing.T) {
	for _, field := range []string{"line", "column"} {
		for _, value := range []int{-1, 0, 100_000_000, 100_000_001} {
			t.Run(fmt.Sprintf("%s/%d", field, value), func(t *testing.T) {
				root := t.TempDir()
				report := projectionReport(t, root, false, true)
				entries := report["Golang errors"].(map[string]any)[filepath.Join(root, "loading.go")].([]any)
				entries[0].(map[string]any)[field] = value
				want := "gosec-tool-or-report-failure"
				if value == 0 || value == 100_000_000 {
					want = "gosec-loading-failure findings=0 loading_packages=1 loading_errors=1"
				}
				if got := projectReport(t, report, root); got != want {
					t.Fatalf("loading entry projection = %q, want %q", got, want)
				}
			})
		}
	}
	for _, field := range []string{"Golang errors", "Issues", "Stats"} {
		t.Run("missing "+field, func(t *testing.T) {
			root := t.TempDir()
			report := projectionReport(t, root, true, true)
			delete(report, field)
			if got := projectReport(t, report, root); got != "gosec-tool-or-report-failure" {
				t.Fatalf("incomplete report projection = %q", got)
			}
		})
	}
}

func TestGosecProjectionLoadingCountBoundaries(t *testing.T) {
	for _, test := range []struct {
		name              string
		packages, entries int
		valid             bool
	}{
		{"empty entries", 1, 0, false},
		{"exact package count", 4096, 1, true},
		{"excess package count", 4097, 1, false},
		{"exact entry count", 1, 4096, true},
		{"excess entry count", 1, 4097, false},
		{"exact aggregate count", 2, 2048, true},
		{"excess aggregate count", 2, 2049, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			report := projectionReport(t, root, false, true)
			loading := map[string]any{}
			for pkg := range test.packages {
				entries := make([]any, test.entries)
				for entry := range entries {
					entries[entry] = map[string]any{"line": 0, "column": 0, "error": "private loading marker"}
				}
				loading[fmt.Sprintf("private-package-%d", pkg)] = entries
			}
			report["Golang errors"] = loading
			want := "gosec-tool-or-report-failure"
			if test.valid {
				want = fmt.Sprintf("gosec-loading-failure findings=0 loading_packages=%d loading_errors=%d", test.packages, test.packages*test.entries)
			}
			if got := projectReport(t, report, root); got != want || strings.Contains(got, "private") {
				t.Fatalf("loading count projection = %q, want %q without private text", got, want)
			}
		})
	}
}

func TestGosecCoordinateGrammarBoundaries(t *testing.T) {
	for _, test := range []struct {
		value       string
		line, valid bool
	}{
		{"1", true, true}, {"9", true, true}, {"100000000", true, true},
		{"0", true, false}, {"100000001", true, false},
		{"1-1", true, true}, {"1-100000000", true, true},
		{"2-1", true, false}, {"1-2-3", true, false},
		{"1-2", false, false}, {"100000000", false, true},
		{"", true, false}, {"/", true, false}, {":", true, false},
		{"1a", true, false}, {"+1", true, false}, {"-1", true, false}, {"0000000001", true, false},
	} {
		t.Run(fmt.Sprintf("%s/line=%t", test.value, test.line), func(t *testing.T) {
			if got := gosecCoordinate(test.value, test.line); got != test.valid {
				t.Fatalf("coordinate admission = %t, want %t", got, test.valid)
			}
		})
	}
}
