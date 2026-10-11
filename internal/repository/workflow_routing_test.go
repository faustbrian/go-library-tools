package repository_test

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"go.yaml.in/yaml/v3"
)

func TestReusableAffectedModulesUseOneBoundedCheckSession(t *testing.T) {
	step := reusableStep(t, "quality", "", "Run repository module contracts")
	for _, scenario := range []struct {
		name, matrix, capable string
		want                  int
		arguments             string
	}{
		{"selected consumers", `[{"directory":"adapter"},{"directory":"consumer"}]`, "true", 0, "check --local --module adapter --module consumer\n"},
		{"empty selection", `[]`, "true", 1, ""},
		{"invalid JSON", `{`, "true", 1, ""},
		{"invalid directory", `[{"directory":"adapter\nother"}]`, "true", 1, ""},
		{"missing capability", `[{"directory":"adapter"}]`, "false", 1, ""},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			task := t.TempDir()
			log := filepath.Join(task, "commands")
			writeExecutable(t, filepath.Join(task, "golib"), `#!/usr/bin/env bash
set -euo pipefail
if test "$1" = --help; then
  printf '%s\n' 'golib check [--local] [--all|--module <directory>]'
  if test "$CAPABLE" = true; then printf '%s\n' 'golib check --local --module <directory> [--module <directory> ...]'; fi
else
  printf '%s\n' "$*" >>"$CHECK_LOG"
  exit "${CHECK_STATUS:-0}"
fi
`)
			run := func(status string) (int, string) {
				command := exec.CommandContext(t.Context(), "bash", "-c", step.Run)
				command.Env = append(os.Environ(), "PATH="+task+string(os.PathListSeparator)+os.Getenv("PATH"), "CAPABLE="+scenario.capable, "MODULE_MATRIX="+scenario.matrix, "CHECK_LOG="+log, "CHECK_STATUS="+status)
				output, err := command.CombinedOutput()
				if err == nil {
					return 0, string(output)
				}
				var failure *exec.ExitError
				if !errors.As(err, &failure) {
					t.Fatal(err)
				}
				return failure.ExitCode(), string(output)
			}
			if code, output := run("0"); code != scenario.want {
				t.Fatalf("check code = %d, want %d: %s", code, scenario.want, output)
			}
			data, err := os.ReadFile(log)
			if err != nil && !os.IsNotExist(err) {
				t.Fatal(err)
			}
			if string(data) != scenario.arguments {
				t.Fatalf("check invocations = %q, want %q", data, scenario.arguments)
			}
			if scenario.want == 0 {
				if code, _ := run("23"); code != 23 {
					t.Fatalf("lost check failure: %d", code)
				}
			}
		})
	}
}

func TestReusableModuleMatrixUsesAffectedInventoryForRuntimePushes(t *testing.T) {
	step := reusableStep(t, "prepare", "modules", "")
	task := t.TempDir()
	log, output := filepath.Join(task, "commands"), filepath.Join(task, "output")
	writeExecutable(t, filepath.Join(task, "golib"), `#!/usr/bin/env bash
set -euo pipefail
printf '%s\n' "$*" >>"$MATRIX_LOG"
case "$*" in
  '--help') printf '%s\n' 'golib inventory --json --base-sha <commit>' 'golib release dry-run [--all|--module <directory>]' ;;
  'inventory --json') printf '%s\n' '{"modules":[{"directory":"."},{"directory":"adapter"},{"directory":"other"}]}' ;;
  "inventory --json --base-sha $BASE_SHA") printf '%s\n' '{"modules":[{"directory":"adapter"}]}' ;;
  'config show --json') printf '%s\n' '{"runtimes":{}}' ;;
  *) exit 64 ;;
esac
`)
	for _, event := range []string{"push", "pull_request"} {
		command := exec.CommandContext(t.Context(), "bash", "-c", step.Run)
		command.Env = append(os.Environ(), "PATH="+task+string(os.PathListSeparator)+os.Getenv("PATH"), "MATRIX_LOG="+log, "GITHUB_OUTPUT="+output, "EVENT_NAME="+event, "RUNTIME=true", "RELEASE_DRY_RUN=false", "RELEASE_MODULE=", "BASE_SHA="+strings.Repeat("a", 40))
		if data, err := command.CombinedOutput(); err != nil {
			t.Fatalf("matrix failed: %v, %s", err, data)
		}
	}
	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), `"directory":"other"`) || !strings.Contains(string(data), `"directory":"adapter"`) {
		t.Fatalf("matrix includes unrelated modules: %s", data)
	}
	data, err = os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(data), "inventory --json --base-sha") != 2 {
		t.Fatalf("push/PR inventory not narrowed: %s", data)
	}
}

func reusableStep(t *testing.T, job, id, name string) workflowStep {
	t.Helper()
	var workflow workflowDocument
	if err := yaml.Unmarshal([]byte(readProjectFile(t, ".github/workflows/library-ci.yml")), &workflow); err != nil {
		t.Fatal(err)
	}
	for _, step := range workflow.Jobs[job].Steps {
		if (id != "" && step.ID == id) || (name != "" && step.Name == name) {
			return step
		}
	}
	t.Fatalf("workflow step missing: %s / %s / %s", job, id, name)
	return workflowStep{}
}

func TestReusableMetadataDocumentationFailsClosedWithoutCampaigns(t *testing.T) {
	step := reusableStep(t, "repository-contract", "", "Validate metadata documentation locally")
	if step.If != "needs.prepare.outputs.runtime == 'false'" {
		t.Fatalf("metadata documentation condition = %q", step.If)
	}
	for _, scenario := range []struct {
		name, helpStatus, docsStatus, capable string
		wantCode                              int
	}{
		{"bounded docs", "0", "0", "true", 0},
		{"broken documentation", "0", "19", "true", 19},
		{"help failure", "17", "0", "true", 17},
		{"old verifier", "0", "0", "false", 1},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			task := t.TempDir()
			log := filepath.Join(task, "commands")
			writeExecutable(t, filepath.Join(task, "golib"), `#!/usr/bin/env bash
set -euo pipefail
printf '%s\n' "$*" >>"$DOCS_LOG"
case "$*" in
  --help)
    if test "$DOCS_CAPABLE" = true; then printf '%s\n' 'golib docs check --local'; fi
    exit "$DOCS_HELP_STATUS" ;;
  'docs check --local') exit "$DOCS_STATUS" ;;
  *) exit 64 ;;
esac
`)
			command := exec.CommandContext(t.Context(), "bash", "-c", step.Run)
			command.Env = append(os.Environ(), "PATH="+task+string(os.PathListSeparator)+os.Getenv("PATH"),
				"DOCS_LOG="+log, "DOCS_CAPABLE="+scenario.capable,
				"DOCS_HELP_STATUS="+scenario.helpStatus, "DOCS_STATUS="+scenario.docsStatus)
			output, err := command.CombinedOutput()
			code := 0
			if err != nil {
				if failure, ok := err.(*exec.ExitError); ok {
					code = failure.ExitCode()
				} else {
					t.Fatal(err)
				}
			}
			if code != scenario.wantCode {
				t.Fatalf("documentation exit = %d, want %d: %s", code, scenario.wantCode, output)
			}
			commands, err := os.ReadFile(log)
			if err != nil {
				t.Fatal(err)
			}
			want := "--help\n"
			if scenario.capable == "true" && scenario.helpStatus == "0" {
				want += "docs check --local\n"
			}
			if string(commands) != want {
				t.Fatalf("documentation commands = %q, want %q", commands, want)
			}
		})
	}
}

func TestReusableScopeRefusesMissingGitEvidenceAndPRRehearsal(t *testing.T) {
	script := reusableStep(t, "prepare", "scope", "").Run
	for _, test := range []struct{ name, event, dryRun, base string }{
		{"missing PR base", "pull_request", "false", ""},
		{"missing main base", "push", "false", ""},
		{"unavailable commit", "pull_request", "false", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
		{"PR cannot select mutation through rehearsal", "pull_request", "true", ""},
	} {
		t.Run(test.name, func(t *testing.T) {
			task := t.TempDir()
			command := exec.CommandContext(t.Context(), "bash", "-c", script)
			command.Dir = task
			command.Env = append(os.Environ(), "EVENT_NAME="+test.event,
				"RELEASE_DRY_RUN="+test.dryRun, "BASE_SHA="+test.base,
				"GITHUB_OUTPUT="+filepath.Join(task, "output"))
			if output, err := command.CombinedOutput(); err == nil {
				t.Fatalf("unsafe classification passed: %s", output)
			}
		})
	}
}

func TestReusableRequiredGateRejectsMissingAndSkippedMandatoryResults(t *testing.T) {
	script := reusableStep(t, "required", "", "Require every repository result").Run
	for _, test := range []struct {
		name      string
		overrides map[string]string
		pass      bool
	}{
		{"runtime PR", nil, true},
		{"docs PR", map[string]string{"RUNTIME": "false", "QUALITY_RESULT": "skipped", "CODEQL_RESULT": "skipped"}, true},
		{"docs main", map[string]string{"EVENT_NAME": "push", "RUNTIME": "false", "QUALITY_RESULT": "skipped", "CODEQL_RESULT": "skipped", "DEPENDENCY_REVIEW_RESULT": "skipped"}, true},
		{"missing classification", map[string]string{"RUNTIME": "", "QUALITY_RESULT": "skipped", "CODEQL_RESULT": "skipped"}, false},
		{"invalid classification", map[string]string{"RUNTIME": "unknown", "QUALITY_RESULT": "skipped", "CODEQL_RESULT": "skipped"}, false},
		{"skipped PR dependency review", map[string]string{"DEPENDENCY_REVIEW_RESULT": "skipped"}, false},
		{"cancelled classifier", map[string]string{"PREPARE_RESULT": "cancelled"}, false},
		{"skipped contract", map[string]string{"CONTRACT_RESULT": "skipped"}, false},
		{"failed tests", map[string]string{"QUALITY_RESULT": "failure"}, false},
		{"cancelled analysis", map[string]string{"CODEQL_RESULT": "cancelled"}, false},
		{"missing analysis", map[string]string{"CODEQL_RESULT": ""}, false},
		{"applicable quality skipped", map[string]string{"QUALITY_RESULT": "skipped"}, false},
		{"inapplicable quality failed", map[string]string{"RUNTIME": "false", "QUALITY_RESULT": "failure", "CODEQL_RESULT": "skipped"}, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			environment := map[string]string{
				"EVENT_NAME": "pull_request", "RUNTIME": "true",
				"DEPENDENCY_REVIEW_RESULT": "success", "PREPARE_RESULT": "success",
				"CONTRACT_RESULT": "success", "QUALITY_RESULT": "success", "CODEQL_RESULT": "success",
			}
			for key, value := range test.overrides {
				environment[key] = value
			}
			command := exec.CommandContext(t.Context(), "bash", "-c", script)
			command.Env = os.Environ()
			for key, value := range environment {
				command.Env = append(command.Env, key+"="+value)
			}
			if output, err := command.CombinedOutput(); (err == nil) != test.pass {
				t.Fatalf("required gate pass = %t, want %t: %s", err == nil, test.pass, output)
			}
		})
	}
}
