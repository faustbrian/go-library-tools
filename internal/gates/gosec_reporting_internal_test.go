package gates

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestGosecFailureReportsOnlyValidatedMetadata(t *testing.T) {
	for _, test := range []struct {
		name    string
		loading bool
		finding bool
		want    string
	}{
		{"findings", false, true, "gosec-findings findings=1 loading_packages=0 loading_errors=0"},
		{"loading", true, false, "gosec-loading-failure findings=0 loading_packages=1 loading_errors=1"},
		{"mixed", true, true, "gosec-loading-failure findings=1 loading_packages=1 loading_errors=1"},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			report := ordinaryGosecReport(root, test.finding, test.loading)
			failure := errors.New("inert process failure")
			var announced strings.Builder
			var captured *sourceInventoryOutput
			runner := Runner{Executor: executorFunction(func(_ context.Context, command Command) error {
				captured, _ = command.Stdout.(*sourceInventoryOutput)
				if !slices.Contains(command.Args, "-fmt=json") {
					t.Error("Gosec report format is not explicit")
				}
				_, _ = command.Stdout.Write(report[:len(report)/2])
				_, _ = command.Stdout.Write(report[len(report)/2:])
				_, _ = io.WriteString(command.Stderr, "private diagnostic\nexit status 1\n")
				return failure
			})}
			err := runner.securityTool(t.Context(), &announced, ".", "gosec", root,
				"github.com/securego/gosec/v2/cmd/gosec@"+gosecVersion,
				"-nosec-require-rules", "-nosec-require-justification", "./")
			if !errors.Is(err, failure) || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("native failure or fixed classification missing: %v", err)
			}
			if captured == nil || captured.data.Len() != 0 {
				t.Fatal("private report retained after classification")
			}
			text := announced.String() + err.Error()
			for _, private := range []string{"private", root, "ordinary.go", "loading.go", "exit status 1"} {
				if strings.Contains(text, private) {
					t.Fatal("scanner-controlled text escaped")
				}
			}
			if test.finding {
				want := fmt.Sprintf("gosec-location %x G115 12 4", sha256.Sum256([]byte("ordinary.go")))
				if !strings.Contains(text, want) {
					t.Fatal("finding identity and coordinates missing")
				}
			}
		})
	}
}

func ordinaryGosecReport(root string, finding, loading bool) []byte {
	issues := []any{}
	loadErrors := map[string]any{}
	if finding {
		issues = append(issues, map[string]any{
			"rule_id": "G115", "file": filepath.Join(root, "ordinary.go"), "line": "12", "column": "4",
			"details": "private diagnostic", "code": "private source", "nosec": false, "suppressions": nil,
		})
	}
	if loading {
		loadErrors[filepath.Join(root, "loading.go")] = []any{map[string]any{"line": 3, "column": 1, "error": "private loading diagnostic"}}
	}
	data, _ := json.Marshal(map[string]any{
		"Golang errors": loadErrors, "Issues": issues,
		"Stats": map[string]int{"files": 1, "lines": 20, "nosec": 0, "found": len(issues)}, "GosecVersion": "v2.29.0",
	})
	return data
}

func TestGosecUnusableReportsRemainToolFailures(t *testing.T) {
	root := t.TempDir()
	valid := string(ordinaryGosecReport(root, true, false))
	for name, report := range map[string]string{
		"empty": "", "truncated": valid[:len(valid)-1], "trailing": valid + "{}",
		"duplicate":                 strings.Replace(valid, `"found":1`, `"found":1,"found":0`, 1),
		"inconsistent":              strings.Replace(valid, `"found":1`, `"found":0`, 1),
		"no-failure-results":        string(ordinaryGosecReport(root, false, false)),
		"invalid-rule":              strings.Replace(valid, `"G115"`, `"private"`, 1),
		"unknown-rule":              strings.Replace(valid, `"G115"`, `"G999"`, 1),
		"case-duplicate":            strings.Replace(valid, `"found":1`, `"found":1,"Found":0`, 1),
		"invalid-coordinate":        strings.Replace(valid, `"12"`, `"private"`, 1),
		"outside-module":            strings.Replace(valid, filepath.Join(root, "ordinary.go"), filepath.Join(filepath.Dir(root), "ordinary.go"), 1),
		"missing-stats":             strings.Replace(valid, `"Stats"`, `"private"`, 1),
		"suppressed":                strings.Replace(valid, `"nosec":false`, `"nosec":true`, 1),
		"missing-suppression-state": strings.Replace(valid, `"suppressions"`, `"private"`, 1),
	} {
		t.Run(name, func(t *testing.T) {
			if got := gosecFailureMetadata([]byte(report), root); got != "gosec-tool-or-report-failure" {
				t.Fatal("unusable report classified as completed findings")
			}
		})
	}
}

func TestGosecLocationsAreBoundedWhileCountsRemainComplete(t *testing.T) {
	root := t.TempDir()
	var report map[string]any
	if err := json.Unmarshal(ordinaryGosecReport(root, true, false), &report); err != nil {
		t.Fatal(err)
	}
	issue := report["Issues"].([]any)[0]
	issues := make([]any, 33)
	for index := range issues {
		issues[index] = issue
	}
	report["Issues"] = issues
	report["Stats"].(map[string]any)["found"] = 33
	data, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	metadata := gosecFailureMetadata(data, root)
	if !strings.Contains(metadata, "gosec-findings findings=33") || strings.Count(metadata, "gosec-location ") != 32 {
		t.Fatal("private report projection lost complete count or exceeded location bound")
	}
	issue.(map[string]any)["line"] = "12-14"
	data, _ = json.Marshal(report)
	if !strings.Contains(gosecFailureMetadata(data, root), " G115 12-14 4") {
		t.Fatal("official multi-line coordinate was not retained")
	}
}

func TestGosecFailurePreservesProcessBoundaries(t *testing.T) {
	for _, test := range []struct {
		name           string
		terminal       bool
		cancel         bool
		overflow       bool
		stderrOverflow bool
		success        bool
		otherTool      bool
		wantClass      bool
	}{
		{name: "terminal", terminal: true, wantClass: true},
		{name: "nonterminal"}, {name: "cancelled", terminal: true, cancel: true},
		{name: "overflow", terminal: true, overflow: true},
		{name: "stderr-overflow", terminal: true, stderrOverflow: true},
		{name: "success", terminal: true, success: true},
		{name: "other-tool", terminal: true, otherTool: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			failure := errors.New("inert process failure")
			runner := Runner{Executor: executorFunction(func(_ context.Context, command Command) error {
				_, _ = command.Stdout.Write(ordinaryGosecReport(root, true, false))
				if test.overflow {
					_, _ = io.WriteString(command.Stdout, strings.Repeat("x", maximumSecurityProcessOutput))
				}
				if test.terminal {
					_, _ = io.WriteString(command.Stderr, "exit status 1\n")
				}
				if test.stderrOverflow {
					_, _ = io.WriteString(command.Stderr, strings.Repeat("x", maximumSecurityProcessOutput))
				}
				if test.cancel {
					cancel()
				}
				if test.success {
					return nil
				}
				return failure
			})}
			tool := "github.com/securego/gosec/v2/cmd/gosec@" + gosecVersion
			if test.otherTool {
				tool = "example.invalid/ordinary@v1.0.0"
			}
			err := runner.securityTool(ctx, io.Discard, ".", "gosec", root, tool, "./")
			if test.success {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			if !errors.Is(err, failure) {
				t.Fatal("native error identity lost")
			}
			classified := strings.Contains(err.Error(), "gosec-findings")
			if classified != test.wantClass {
				t.Fatal("completion classified across a process boundary")
			}
			if !test.terminal && !strings.Contains(err.Error(), "gosec-tool-or-report-failure") {
				t.Fatal("missing tool failure class")
			}
		})
	}
}
