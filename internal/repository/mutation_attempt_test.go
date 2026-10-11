package repository_test

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"go.yaml.in/yaml/v3"
)

func TestMutationAssuranceScheduledAttemptDecision(t *testing.T) {
	var workflow workflowDocument
	if err := yaml.Unmarshal([]byte(readProjectFile(t, ".github/workflows/mutation-assurance.yml")), &workflow); err != nil {
		t.Fatal(err)
	}
	guard := "printf 'run=true\\n' >>\"$GITHUB_OUTPUT\"" // Existing workflow has no attempt guard.
	for _, step := range workflow.Jobs["mutation"].Steps {
		if step.Name == "Decide scheduled attempt" {
			guard = step.Run
		}
	}
	for _, test := range []struct {
		name, event, changed, priorResult, marker, history string
		setupFailure, apiFailure, missingGit               bool
		wantRun, wantFailure                               bool
	}{
		{name: "unchanged successful attempt", wantRun: false},
		{name: "unchanged exhausted attempt", priorResult: "failure", wantRun: false},
		{name: "prose does not restart", changed: "README.md", wantRun: false},
		{name: "historical output does not restart", changed: ".verification/mutation/reports/old.json", wantRun: false},
		{name: "source changes", changed: "internal/config/config.go", wantRun: true},
		{name: "tests change", changed: "internal/config/config_test.go", wantRun: true},
		{name: "fixture changes", changed: "internal/config/testdata/README.md", wantRun: true},
		{name: "dependencies change", changed: "go.sum", wantRun: true},
		{name: "verifier changes", changed: "internal/mutation/assets/verifier.go", wantRun: true},
		{name: "workflow changes", changed: ".github/workflows/mutation-assurance.yml", wantRun: true},
		{name: "runner changes", marker: "package=internal/config image=ubuntu24/changed arch=X64 go=go1.27.2", wantRun: true},
		{name: "other package", marker: "package=internal/inventory image=ubuntu24/20261004.1 arch=X64 go=go1.27.2", wantRun: true},
		{name: "setup failure was not a campaign", setupFailure: true, wantRun: true},
		{name: "manual retry", event: "workflow_dispatch", apiFailure: true, wantRun: true},
		{name: "history unavailable", apiFailure: true, wantFailure: true},
		{name: "malformed history", history: "{}", wantFailure: true},
		{name: "missing immutable Git", missingGit: true, wantFailure: true},
		{name: "ambiguous runtime identity", marker: "unrecognized", wantFailure: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			gitEnv := append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_NOSYSTEM=1")
			runGit := func(args ...string) string {
				command := exec.CommandContext(t.Context(), "git", args...)
				command.Dir, command.Env = root, gitEnv
				data, err := command.CombinedOutput()
				if err != nil {
					t.Fatalf("fixture Git: %v, %q", err, data)
				}
				return strings.TrimSpace(string(data))
			}
			writeFile(t, root, "internal/config/config.go", "package config\n")
			writeFile(t, root, ".go-version", "1.27.2\n")
			runGit("init", "--quiet")
			runGit("add", "--", "internal/config/config.go", ".go-version")
			runGit("-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "-c", "commit.gpgsign=false", "commit", "--quiet", "-m", "baseline")
			prior := runGit("rev-parse", "HEAD")
			if test.changed != "" {
				writeFile(t, root, test.changed, "changed execution input\n")
				runGit("add", "--", test.changed)
				runGit("-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "-c", "commit.gpgsign=false", "commit", "--quiet", "-m", "changed")
			}
			head := runGit("rev-parse", "HEAD")
			if test.missingGit {
				prior = strings.Repeat("a", 40)
			}
			history, _ := json.Marshal(map[string]any{"total_count": 1, "workflow_runs": []any{map[string]any{
				"id": 73, "run_attempt": 1, "head_sha": prior, "head_branch": "main", "event": "schedule", "status": "completed", "conclusion": "failure",
			}}})
			if test.history != "" {
				history = []byte(test.history)
			}
			writeFile(t, root, "history.json", string(history))
			result := test.priorResult
			if result == "" {
				result = "success"
			}
			if test.setupFailure {
				result = "skipped"
			}
			jobs, _ := json.Marshal(map[string]any{"total_count": 1, "jobs": []any{map[string]any{
				"id": 71, "run_id": 73, "head_sha": prior, "name": "Selected mutation assurance", "status": "completed", "conclusion": "failure",
				"steps": []any{map[string]any{"name": "Run bounded package campaign", "status": "completed", "conclusion": result, "started_at": "2026-10-11T15:05:08Z"}},
			}}})
			writeFile(t, root, "jobs.json", string(jobs))
			marker := test.marker
			if marker == "" {
				marker = "package=internal/config image=ubuntu24/20261004.1 arch=X64 go=go1.27.2"
			}
			writeFile(t, root, "logs.txt", "2026-10-11T15:05:08Z golib-mutation-inputs "+marker+"\n")
			writeExecutable(t, filepath.Join(root, "bin", "gh"), `#!/usr/bin/env bash
set -euo pipefail
[[ "$API_FAILURE" != true ]] || exit 1
case "$*" in
  *'/jobs?filter=latest&per_page=100') cat jobs.json ;;
  *'/jobs/71/logs') cat logs.txt ;;
  *'/workflows/mutation-assurance.yml/runs?per_page=100') cat history.json ;;
  *) exit 64 ;;
esac
`)
			if data, err := os.ReadFile(filepath.Join(projectRoot(t), ".github/scripts/select-mutation-attempt.py")); err == nil {
				writeFile(t, root, ".github/scripts/select-mutation-attempt.py", string(data))
			}
			output, summary := filepath.Join(root, "output"), filepath.Join(root, "summary")
			event := test.event
			if event == "" {
				event = "schedule"
			}
			command := exec.CommandContext(t.Context(), "bash", "-c", guard)
			command.Dir = root
			command.Env = append(gitEnv, "PATH="+filepath.Join(root, "bin")+string(os.PathListSeparator)+os.Getenv("PATH"),
				"EVENT_NAME="+event, "SELECTED_PACKAGE=internal/config", "GITHUB_SHA="+head,
				"GITHUB_REPOSITORY=faustbrian/go-library-tools", "GITHUB_RUN_ID=74", "GITHUB_OUTPUT="+output, "GITHUB_STEP_SUMMARY="+summary,
				"ImageOS=ubuntu24", "ImageVersion=20261004.1", "RUNNER_ARCH=X64", fmt.Sprintf("API_FAILURE=%t", test.apiFailure))
			data, err := command.CombinedOutput()
			if (err != nil) != test.wantFailure {
				t.Fatalf("attempt decision failure = %v, want %t: %q", err, test.wantFailure, data)
			}
			decision, _ := os.ReadFile(output)
			if test.wantFailure {
				if strings.Contains(string(decision), "run=") {
					t.Fatalf("invalid history published a decision: %q", decision)
				}
			} else if !strings.Contains(string(decision), fmt.Sprintf("run=%t\n", test.wantRun)) {
				t.Fatalf("scheduled execution decision = %q, want run=%t", decision, test.wantRun)
			}
			if !test.wantRun && !test.wantFailure {
				summaryData, _ := os.ReadFile(summary)
				if !strings.Contains(string(summaryData), "/actions/runs/73") || !strings.Contains(string(summaryData), "not passing assurance") {
					t.Fatalf("prior actual result/link lost: %q", summaryData)
				}
			}
		})
	}
}

func TestMutationAssuranceRetainsOnlyFreshNativeOutput(t *testing.T) {
	script := readProjectFile(t, ".github/scripts/check-mutation-output.py")
	for _, mode := range []string{"unchanged", "new", "updated", "unsafe"} {
		t.Run(mode, func(t *testing.T) {
			root := t.TempDir()
			old := ".verification/mutation/reports/old.json"
			writeFile(t, root, old, "{\"result\":\"historical\"}\n")
			baseline, fresh := filepath.Join(root, "before.json"), filepath.Join(root, "fresh")
			if err := os.Mkdir(fresh, 0o700); err != nil {
				t.Fatal(err)
			}
			run := func(arguments ...string) ([]byte, error) {
				command := exec.CommandContext(t.Context(), "python3", append([]string{"-c", script}, arguments...)...)
				command.Dir = root
				return command.CombinedOutput()
			}
			if data, err := run("--snapshot", baseline); err != nil {
				t.Fatalf("snapshot: %v, %q", err, data)
			}
			switch mode {
			case "new":
				writeFile(t, root, ".verification/mutation/failed-reports/new.json", "{\"result\":\"failed\"}\n")
			case "updated":
				writeFile(t, root, old, "{\"result\":\"fresh\"}\n")
			case "unsafe":
				if err := os.Symlink(baseline, filepath.Join(root, ".verification/mutation/reports/linked.json")); err != nil {
					t.Fatal(err)
				}
			}
			data, err := run("--fresh", baseline, fresh)
			if mode == "unsafe" {
				if err == nil || strings.Contains(string(data), "fresh=") {
					t.Fatalf("unsafe native output admitted: %v, %q", err, data)
				}
				return
			}
			want := "fresh=true\n"
			if mode == "unchanged" {
				want = "fresh=false\n"
			}
			if err != nil || string(data) != want {
				t.Fatalf("fresh native output = %q, %v; want %q", data, err, want)
			}
			retained, _ := os.ReadFile(filepath.Join(fresh, old))
			if mode != "updated" && len(retained) != 0 {
				t.Fatalf("historical artifact copied: %q", retained)
			}
			if mode == "updated" && string(retained) != "{\"result\":\"fresh\"}\n" {
				t.Fatalf("fresh replacement missing: %q", retained)
			}
			if mode == "new" {
				newData, readErr := os.ReadFile(filepath.Join(fresh, ".verification/mutation/failed-reports/new.json"))
				if readErr != nil || string(newData) != "{\"result\":\"failed\"}\n" {
					t.Fatalf("fresh failed diagnostic missing: %q, %v", newData, readErr)
				}
			}
		})
	}
}
