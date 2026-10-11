package repository_test

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"go.yaml.in/yaml/v3"
)

func mutationAssuranceStep(t *testing.T, name string) workflowStep {
	t.Helper()
	var workflow workflowDocument
	if err := yaml.Unmarshal([]byte(readProjectFile(t, ".github/workflows/mutation-assurance.yml")), &workflow); err != nil {
		t.Fatal(err)
	}
	for _, step := range workflow.Jobs["mutation"].Steps {
		if step.Name == name {
			return step
		}
	}
	t.Fatalf("missing mutation assurance step %s", name)
	return workflowStep{}
}

func writeFile(t *testing.T, root, relative, content string) {
	t.Helper()
	path := filepath.Join(root, relative)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestMutationAssuranceConfinesAndBoundsNativeOutput(t *testing.T) {
	script := readProjectFile(t, ".github/scripts/check-mutation-output.py")
	for _, test := range []struct {
		name    string
		prepare func(*testing.T, string, string)
		fail    bool
	}{
		{name: "missing output"},
		{name: "flat native JSON", prepare: func(t *testing.T, root, outside string) {
			for _, path := range []string{".verification/by-input/mutation/input.json", ".verification/mutation/reports/report.json", ".verification/mutation/failed-reports/failure.json"} {
				writeFile(t, root, path, "{}\n")
			}
		}},
		{name: "non JSON", fail: true, prepare: func(t *testing.T, root, outside string) {
			writeFile(t, root, ".verification/mutation/reports/native.log", "not an admitted projection")
		}},
		{name: "nested output", fail: true, prepare: func(t *testing.T, root, outside string) {
			writeFile(t, root, ".verification/mutation/reports/nested/output.json", "{}")
		}},
		{name: "linked output", fail: true, prepare: func(t *testing.T, root, outside string) {
			path := filepath.Join(root, ".verification/mutation/reports")
			if err := os.MkdirAll(path, 0o700); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(filepath.Join(outside, "sentinel"), filepath.Join(path, "linked.json")); err != nil {
				t.Fatal(err)
			}
		}},
		{name: "linked ancestor", fail: true, prepare: func(t *testing.T, root, outside string) {
			writeFile(t, outside, "mutation/reports/report.json", "{}")
			if err := os.Symlink(outside, filepath.Join(root, ".verification")); err != nil {
				t.Fatal(err)
			}
		}},
		{name: "too many files", fail: true, prepare: func(t *testing.T, root, outside string) {
			for i := range 2049 {
				writeFile(t, root, fmt.Sprintf(".verification/mutation/reports/%04d.json", i), "{}")
			}
		}},
		{name: "oversized logical output", fail: true, prepare: func(t *testing.T, root, outside string) {
			writeFile(t, root, ".verification/mutation/reports/report.json", "{}")
			if err := os.Truncate(filepath.Join(root, ".verification/mutation/reports/report.json"), 32*1024*1024+1); err != nil {
				t.Fatal(err)
			}
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			root, outside := t.TempDir(), t.TempDir()
			writeFile(t, outside, "sentinel", "outside content stays private")
			if test.prepare != nil {
				test.prepare(t, root, outside)
			}
			command := exec.CommandContext(t.Context(), "python3", "-c", script)
			command.Dir = root
			data, err := command.CombinedOutput()
			if (err != nil) != test.fail {
				t.Fatalf("output admission = %v, output %q", err, data)
			}
			if strings.Contains(string(data), "outside content stays private") {
				t.Fatal("output admission disclosed outside bytes")
			}
			content, err := os.ReadFile(filepath.Join(outside, "sentinel"))
			if err != nil || string(content) != "outside content stays private" {
				t.Fatal("output admission changed outside content")
			}
		})
	}
}

func TestMutationAssuranceRefusesCacheReplacementOfCommittedEvidence(t *testing.T) {
	script := mutationAssuranceStep(t, "Validate restored native output").Run
	checker := readProjectFile(t, ".github/scripts/check-mutation-output.py")
	for _, changed := range []bool{false, true} {
		t.Run(fmt.Sprintf("changed=%t", changed), func(t *testing.T) {
			root := t.TempDir()
			writeFile(t, root, ".github/scripts/check-mutation-output.py", checker)
			path := ".verification/by-input/mutation/committed.json"
			writeFile(t, root, path, "{\"source\":\"committed\"}\n")
			gitEnv := append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_NOSYSTEM=1")
			for _, args := range [][]string{{"init", "--quiet"}, {"add", "--", path}} {
				command := exec.CommandContext(t.Context(), "git", args...)
				command.Dir, command.Env = root, gitEnv
				if data, err := command.CombinedOutput(); err != nil {
					t.Fatalf("fixture Git: %v, %q", err, data)
				}
			}
			writeFile(t, root, ".verification/by-input/mutation/additional.json", "{}\n")
			if changed {
				writeFile(t, root, path, "{\"source\":\"cache replacement\"}\n")
			}
			command := exec.CommandContext(t.Context(), "bash", "-c", script)
			command.Dir, command.Env = root, gitEnv
			data, err := command.CombinedOutput()
			if (err != nil) != changed {
				t.Fatalf("restored evidence admission = %v, output %q", err, data)
			}
		})
	}
}

func TestMutationAssuranceSelectsOneMainRiskAndRejectsUnsafeEvents(t *testing.T) {
	script := mutationAssuranceStep(t, "Select one owned risk").Run
	for _, test := range []struct {
		name, event, ref, selection, want string
		fail                              bool
	}{
		{name: "weekly configuration", event: "schedule", ref: "refs/heads/main", want: "internal/config"},
		{name: "explicit inventory", event: "workflow_dispatch", ref: "refs/heads/main", selection: "internal/inventory", want: "internal/inventory"},
		{name: "explicit confined reads", event: "workflow_dispatch", ref: "refs/heads/main", selection: "internal/repositoryfile", want: "internal/repositoryfile"},
		{name: "missing selection", event: "workflow_dispatch", ref: "refs/heads/main", fail: true},
		{name: "unknown package", event: "workflow_dispatch", ref: "refs/heads/main", selection: "internal/gates", fail: true},
		{name: "shell-shaped input", event: "workflow_dispatch", ref: "refs/heads/main", selection: "internal/config; false", fail: true},
		{name: "pull request", event: "pull_request", ref: "refs/heads/main", selection: "internal/config", fail: true},
		{name: "main push", event: "push", ref: "refs/heads/main", selection: "internal/config", fail: true},
		{name: "feature dispatch", event: "workflow_dispatch", ref: "refs/heads/feature/example", selection: "internal/config", fail: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			output := filepath.Join(t.TempDir(), "output")
			command := exec.CommandContext(t.Context(), "bash", "-c", script)
			command.Env = append(os.Environ(), "EVENT_NAME="+test.event, "SOURCE_REF="+test.ref, "SELECTED_PACKAGE="+test.selection, "GITHUB_OUTPUT="+output)
			data, err := command.CombinedOutput()
			if (err != nil) != test.fail {
				t.Fatalf("selection = %v, output %q", err, data)
			}
			selected, readErr := os.ReadFile(output)
			if test.fail {
				if !os.IsNotExist(readErr) {
					t.Fatalf("invalid selection published an output: %q, %v", selected, readErr)
				}
			} else if readErr != nil || string(selected) != "package="+test.want+"\n" {
				t.Fatalf("selected package = %q, %v", selected, readErr)
			}
		})
	}
}

func TestMutationAssuranceExecutesSelectedCampaignAndCleansFailures(t *testing.T) {
	script := mutationAssuranceStep(t, "Run bounded package campaign").Run
	for _, test := range []struct {
		name, buildStatus, campaignStatus, capable string
		want                                       int
		campaign                                   bool
		cleanupFailure                             bool
	}{
		{"success", "0", "0", "true", 0, true, false},
		{"build failure", "19", "0", "true", 19, false, false},
		{"missing capability", "0", "0", "false", 1, false, false},
		{"viable survivor", "0", "23", "true", 23, true, false},
		{"budget exhaustion", "0", "124", "true", 124, true, false},
		{"cleanup failure", "0", "0", "true", 1, true, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			task := t.TempDir()
			log, summary := filepath.Join(task, "commands"), filepath.Join(task, "summary")
			cli := filepath.Join(task, "cli-fixture")
			writeExecutable(t, cli, `#!/usr/bin/env bash
set -euo pipefail
if [[ "$1" == --help ]]; then
  if [[ "$CAPABLE" == true ]]; then
    printf '%s\n' 'golib mutation --module <directory> --package <directory>'
  fi
else
  printf '%s\n' "$*" >>"$CAMPAIGN_LOG"
  exit "$CAMPAIGN_STATUS"
fi
`)
			writeExecutable(t, filepath.Join(task, "go"), `#!/usr/bin/env bash
set -euo pipefail
[[ "$GOCACHE" == "$RUNNER_TEMP"/golib-mutation-assurance.*/cache ]]
[[ "$GOMODCACHE" == "$RUNNER_TEMP"/golib-mutation-assurance.*/mod ]]
[[ "$GOTMPDIR" == "$RUNNER_TEMP"/golib-mutation-assurance.*/tmp ]]
if [[ "$BUILD_STATUS" != 0 ]]; then exit "$BUILD_STATUS"; fi
while [[ $# -gt 0 ]]; do
  if [[ "$1" == -o ]]; then cp "$CLI_FIXTURE" "$2"; exit 0; fi
  shift
done
exit 64
`)
			if test.cleanupFailure {
				writeExecutable(t, filepath.Join(task, "chmod"), "#!/usr/bin/env bash\nexit 1\n")
			}
			command := exec.CommandContext(t.Context(), "bash", "-c", script)
			command.Dir = task
			command.Env = append(os.Environ(), "PATH="+task+string(os.PathListSeparator)+os.Getenv("PATH"), "RUNNER_TEMP="+task, "GITHUB_STEP_SUMMARY="+summary,
				"SELECTED_PACKAGE=internal/config", "CAMPAIGN_LOG="+log, "CLI_FIXTURE="+cli, "CAPABLE="+test.capable,
				"BUILD_STATUS="+test.buildStatus, "CAMPAIGN_STATUS="+test.campaignStatus)
			data, err := command.CombinedOutput()
			code := 0
			if err != nil {
				var failure *exec.ExitError
				if !errors.As(err, &failure) {
					t.Fatal(err)
				}
				code = failure.ExitCode()
			}
			if code != test.want {
				t.Fatalf("campaign exit = %d, want %d: %q", code, test.want, data)
			}
			commands, readErr := os.ReadFile(log)
			if test.campaign {
				if readErr != nil || string(commands) != "mutation --module . --package internal/config\n" {
					t.Fatalf("campaign invocation = %q, %v", commands, readErr)
				}
			} else if !os.IsNotExist(readErr) {
				t.Fatalf("failed prerequisite invoked campaign: %q, %v", commands, readErr)
			}
			leftovers, err := filepath.Glob(filepath.Join(task, "golib-mutation-assurance.*"))
			if err != nil || len(leftovers) != 0 {
				t.Fatalf("owned compile/cache workspace retained: %v, %v", leftovers, err)
			}
			result, err := os.ReadFile(summary)
			summaryCode := test.want
			if test.cleanupFailure {
				summaryCode = 0
			}
			if err != nil || !strings.Contains(string(result), "Owner: faustbrian") || !strings.Contains(string(result), "exit "+strconv.Itoa(summaryCode)) {
				t.Fatalf("attributable result summary = %q, %v", result, err)
			}
		})
	}
}
