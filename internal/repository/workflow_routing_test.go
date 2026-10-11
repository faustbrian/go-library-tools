package repository_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"go.yaml.in/yaml/v3"
)

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
