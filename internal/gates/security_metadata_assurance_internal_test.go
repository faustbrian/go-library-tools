package gates

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGosecMetadataRejectsInvalidStatisticsAndLoading(t *testing.T) {
	for _, test := range []struct {
		name   string
		change func(map[string]any)
	}{
		{"missing statistic", func(report map[string]any) { delete(report["Stats"].(map[string]int), "lines") }},
		{"negative statistic", func(report map[string]any) { report["Stats"].(map[string]int)["files"] = -1 }},
		{"excessive statistic", func(report map[string]any) { report["Stats"].(map[string]int)["nosec"] = 100000001 }},
		{"invalid issues type", func(report map[string]any) { report["Issues"] = "private diagnostic" }},
		{"invalid loading type", func(report map[string]any) { report["Golang errors"] = "private diagnostic" }},
		{"empty loading entry", func(report map[string]any) { report["Golang errors"] = map[string]any{"private.go": []any{}} }},
		{"excessive loading entries", func(report map[string]any) { report["Golang errors"] = map[string]any{"private.go": make([]any, 4097)} }},
		{"missing loading message", func(report map[string]any) {
			report["Golang errors"] = map[string]any{"private.go": []any{map[string]any{"line": 1, "column": 1}}}
		}},
		{"negative loading line", func(report map[string]any) {
			report["Golang errors"] = map[string]any{"private.go": []any{map[string]any{"line": -1, "column": 1, "error": "private diagnostic"}}}
		}},
		{"excessive loading column", func(report map[string]any) {
			report["Golang errors"] = map[string]any{"private.go": []any{map[string]any{"line": 1, "column": 100000001, "error": "private diagnostic"}}}
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			report, _ := ordinaryGosecFindings(root, 1)
			test.change(report)
			data, err := json.Marshal(report)
			if err != nil {
				t.Fatal(err)
			}
			if got := gosecFailureMetadata(data, root); got != "gosec-tool-or-report-failure" {
				t.Fatal("invalid report did not retain fixed, private tool-failure classification")
			}
		})
	}
}

func TestGosecMetadataProjectsRelativePathsAndRejectsInvalidCoordinates(t *testing.T) {
	root := t.TempDir()
	report, issues := ordinaryGosecFindings(root, 1)
	issues[0]["file"] = "ordinary.go"
	data, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	want := fmt.Sprintf("gosec-findings findings=1 loading_packages=0 loading_errors=0\ngosec-location %x G115 1 4", sha256.Sum256([]byte("ordinary.go")))
	if got := gosecFailureMetadata(data, root); got != want {
		t.Fatal("relative finding identity projection changed")
	}
	for _, test := range []struct{ name, field, value string }{
		{"outside relative path", "file", "../outside.go"},
		{"empty coordinate", "line", ""},
		{"extra line range", "line", "1-2-3"},
		{"column range", "column", "1-2"},
	} {
		t.Run(test.name, func(t *testing.T) {
			candidate, candidateIssues := ordinaryGosecFindings(root, 1)
			candidateIssues[0][test.field] = test.value
			data, err := json.Marshal(candidate)
			if err != nil {
				t.Fatal(err)
			}
			if gosecFailureMetadata(data, root) != "gosec-tool-or-report-failure" {
				t.Fatal("invalid relative identity or coordinate admitted")
			}
		})
	}
}

func TestGosecMetadataNestingAdmission(t *testing.T) {
	// Finite inert JSON values exercise the private parser only, not a scanner.
	if !uniqueGosecJSON([]byte(strings.Repeat("[", 64) + "0" + strings.Repeat("]", 64))) {
		t.Fatal("inclusive nesting boundary refused")
	}
	if uniqueGosecJSON([]byte(strings.Repeat("[", 65) + "0" + strings.Repeat("]", 65))) {
		t.Fatal("one-over nesting boundary admitted")
	}
	if uniqueGosecJSON([]byte("[")) {
		t.Fatal("incomplete nested value admitted")
	}
}

func TestSecretMetadataInvalidationDoesNotRetainLaterLocations(t *testing.T) {
	var metadata secretMetadata
	metadata.write([]byte("private diagnostic\n"))
	location := strings.Repeat("a", 64) + " " + strings.Repeat("b", 64) + " 1 -\n"
	metadata.write([]byte(location))
	if !metadata.invalid || metadata.length != 0 || len(metadata.locations) != 0 || metadata.summary() != "" {
		t.Fatal("invalid metadata retained or disclosed a later candidate")
	}
	for _, value := range metadata.line {
		if value != 0 {
			t.Fatal("invalid metadata retained input bytes")
		}
	}
}

func TestSecretMetadataTemplateFailurePreventsExecution(t *testing.T) {
	calls := 0
	runner := Runner{Executor: workspaceExecutor{directory: filepath.Join(t.TempDir(), "missing"), run: func(context.Context, Command) error { calls++; return nil }}}
	err := runner.securityTool(t.Context(), io.Discard, ".", "secrets-history", ".", "github.com/zricethezav/gitleaks/v8@"+gitleaksVersion, "git", ".")
	if err == nil || !errors.Is(err, os.ErrNotExist) || calls != 0 {
		t.Fatal("template creation failure did not stop scanner dispatch")
	}
}
