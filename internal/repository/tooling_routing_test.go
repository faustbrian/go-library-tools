package repository_test

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"go.yaml.in/yaml/v3"
)

func TestToolingScopeRejectsInvalidCatalogsAndQualifiesPublishedChanges(t *testing.T) {
	var workflow workflowDocument
	if err := yaml.Unmarshal([]byte(readProjectFile(t, ".github/workflows/ci.yml")), &workflow); err != nil {
		t.Fatal(err)
	}
	var script string
	for _, step := range workflow.Jobs["compatibility-consumer-scope"].Steps {
		if step.ID == "scope" {
			script = step.Run
		}
	}
	if script == "" {
		t.Fatal("tooling scope has no script")
	}
	for _, scenario := range []struct {
		name, invalid   string
		publishedChange bool
		changedPath     string
	}{
		{"unchanged publication", "", false, ""},
		{"changed published content", "", true, ""},
		{"Tools manifest only", "", false, "modules.json"},
		{"consumer dependency", "", false, "release/compatibility-consumer/go.mod"},
		{"consumer checksums", "", false, "release/compatibility-consumer/go.sum"},
		{"consumer source", "", false, "release/compatibility-consumer/consumer_test.go"},
		{"consumer toolchain", "", false, ".go-version"},
		{"consumer workflow", "", false, ".github/workflows/ci.yml"},
		{"malformed JSON", "{", false, ""},
		{"missing sets", `{}`, false, ""},
		{"empty sets", `{"sets":[]}`, false, ""},
		{"wrong sets type", `{"sets":{}}`, false, ""},
		{"invalid draft identity", `{"sets":[{"publication_status":"unreleased","set_id":"invalid"}]}`, false, ""},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			root := t.TempDir()
			write(t, filepath.Join(root, "tools/compatibility/generate.py"), readProjectFile(t, "tools/compatibility/generate.py"))
			write(t, filepath.Join(root, "docs/ecosystem/design-language.md"), "# Fixture\n")
			catalog := map[string]any{"modules": []any{map[string]any{
				"repository": "github.com/faustbrian/go-parent", "directory": "adapters/foo",
				"module_path": "github.com/faustbrian/go-parent/adapters/foo", "version": "1.2.3", "tag_prefix": "adapters/foo/v",
			}}}
			encodedCatalog, err := json.MarshalIndent(catalog, "", "  ")
			if err != nil {
				t.Fatal(err)
			}
			for _, name := range []string{"catalog-consumer.json", "catalog-engineering.json"} {
				write(t, filepath.Join(root, "docs/ecosystem", name), string(encodedCatalog)+"\n")
			}
			item := map[string]any{
				"set_id": "golib-compat-v1-20260909.1", "publication_status": "published", "installable": true,
				"go":        map[string]any{"version": "1.27.2", "os_arch": []string{"linux/amd64"}},
				"modules":   []any{map[string]any{"module_path": "github.com/faustbrian/go-parent/adapters/foo", "version": "v1.2.3", "source_revision": strings.Repeat("a", 40)}},
				"scenarios": []string{"fixture"}, "observed_at": "2026-09-09T00:00:00Z",
				"evidence": map[string]any{"observation": "fixture public resolution"}, "caveats": []string{"fixture"},
				"source_references": []string{"docs/ecosystem/design-language.md"},
			}
			writeSet := func() {
				t.Helper()
				evidence := item["evidence"].(map[string]any)
				delete(evidence, "content_sha256")
				canonical, err := json.Marshal(item)
				if err != nil {
					t.Fatal(err)
				}
				evidence["content_sha256"] = fmt.Sprintf("sha256:%x", sha256.Sum256(canonical))
				encoded, err := json.Marshal(map[string]any{"format": "golib-compatibility-sets-v1", "sets": []any{item}})
				if err != nil {
					t.Fatal(err)
				}
				write(t, filepath.Join(root, "docs/ecosystem/compatibility-sets.json"), string(encoded)+"\n")
			}
			writeSet()
			if scenario.changedPath != "" {
				initial := "fixture input\n"
				if scenario.changedPath == "modules.json" {
					initial = `{"modules":[{"directory":".","version":"2.0.0"}]}` + "\n"
				}
				write(t, filepath.Join(root, scenario.changedPath), initial)
			}
			runGit(t, root, "init", "-q")
			runGit(t, root, "add", "tools/compatibility/generate.py", "docs/ecosystem")
			if scenario.changedPath != "" {
				runGit(t, root, "add", scenario.changedPath)
			}
			runGit(t, root, "-c", "core.hooksPath=/dev/null", "-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "-c", "commit.gpgsign=false", "commit", "-qm", "fixture")
			base := strings.TrimSpace(runGit(t, root, "rev-parse", "HEAD"))
			if scenario.invalid != "" {
				write(t, filepath.Join(root, "docs/ecosystem/compatibility-sets.json"), scenario.invalid)
			} else if scenario.publishedChange {
				item["caveats"] = []string{"changed fixture"}
				writeSet()
			}
			if scenario.changedPath != "" {
				updated := "changed fixture input\n"
				if scenario.changedPath == "modules.json" {
					updated = `{"modules":[{"directory":".","version":"2.1.0"}]}` + "\n"
				}
				write(t, filepath.Join(root, scenario.changedPath), updated)
				runGit(t, root, "add", scenario.changedPath)
				runGit(t, root, "-c", "core.hooksPath=/dev/null", "-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "-c", "commit.gpgsign=false", "commit", "-qm", "changed input")
			}
			resultPath := filepath.Join(root, "github-output")
			command := exec.CommandContext(t.Context(), "bash", "-e", "-c", script)
			command.Dir = root
			command.Env = append(os.Environ(), "BASE_SHA="+base, "RUNNER_TEMP="+root, "GITHUB_OUTPUT="+resultPath, "PYTHONDONTWRITEBYTECODE=1")
			output, err := command.CombinedOutput()
			if scenario.invalid != "" {
				if err == nil {
					t.Fatal("invalid compatibility catalog was classified successfully")
				}
				return
			}
			if err != nil {
				t.Fatalf("scope failed: %v: %s", err, output)
			}
			result, err := os.ReadFile(resultPath)
			if err != nil {
				t.Fatal(err)
			}
			consumerRequired := scenario.publishedChange || (scenario.changedPath != "" && scenario.changedPath != "modules.json")
			want := fmt.Sprintf("publication_required=%t\nrequired=%t\n", scenario.publishedChange, consumerRequired)
			if string(result) != want {
				t.Fatalf("scope = %q, want %q", result, want)
			}
		})
	}
}

func TestToolingRequiredGateHonorsApplicableJobs(t *testing.T) {
	var workflow workflowDocument
	if err := yaml.Unmarshal([]byte(readProjectFile(t, ".github/workflows/ci.yml")), &workflow); err != nil {
		t.Fatal(err)
	}
	var script string
	for _, step := range workflow.Jobs["required"].Steps {
		if step.Run != "" {
			script = step.Run
		}
	}
	if script == "" {
		t.Fatal("tooling required job has no script")
	}
	for _, scenario := range []struct {
		name      string
		overrides map[string]string
		pass      bool
	}{
		{"ordinary PR", nil, true},
		{"ordinary main", map[string]string{"EVENT_NAME": "push", "DEPENDENCY_REVIEW_RESULT": "skipped"}, true},
		{"publication", map[string]string{"PUBLICATION_REQUIRED": "true", "COMPATIBILITY_CONSUMER_REQUIRED": "true", "COMPATIBILITY_CONSUMER_RESULT": "success", "MILESTONE_RESULT": "success"}, true},
		{"consumer update", map[string]string{"COMPATIBILITY_CONSUMER_REQUIRED": "true", "COMPATIBILITY_CONSUMER_RESULT": "success"}, true},
		{"scheduled", map[string]string{"EVENT_NAME": "schedule", "DEPENDENCY_REVIEW_RESULT": "skipped", "COMPATIBILITY_SCOPE_RESULT": "skipped", "COMPATIBILITY_CONSUMER_REQUIRED": "", "PUBLICATION_REQUIRED": ""}, true},
		{"manual milestone", map[string]string{"EVENT_NAME": "workflow_dispatch", "DEPENDENCY_REVIEW_RESULT": "skipped", "COMPATIBILITY_SCOPE_RESULT": "skipped", "COMPATIBILITY_CONSUMER_REQUIRED": "", "PUBLICATION_REQUIRED": "", "MILESTONE_SELECTED": "true", "MILESTONE_RESULT": "success"}, true},
		{"PR dependency review skipped", map[string]string{"DEPENDENCY_REVIEW_RESULT": "skipped"}, false},
		{"classifier skipped", map[string]string{"COMPATIBILITY_SCOPE_RESULT": "skipped"}, false},
		{"consumer classification missing", map[string]string{"COMPATIBILITY_CONSUMER_REQUIRED": ""}, false},
		{"publication classification invalid", map[string]string{"PUBLICATION_REQUIRED": "unknown"}, false},
		{"applicable consumer skipped", map[string]string{"COMPATIBILITY_CONSUMER_REQUIRED": "true"}, false},
		{"publication milestone skipped", map[string]string{"PUBLICATION_REQUIRED": "true", "COMPATIBILITY_CONSUMER_REQUIRED": "true", "COMPATIBILITY_CONSUMER_RESULT": "success"}, false},
		{"manual milestone skipped", map[string]string{"EVENT_NAME": "workflow_dispatch", "DEPENDENCY_REVIEW_RESULT": "skipped", "COMPATIBILITY_SCOPE_RESULT": "skipped", "MILESTONE_SELECTED": "true"}, false},
		{"unexpected ordinary milestone", map[string]string{"MILESTONE_RESULT": "success"}, false},
		{"quality failed", map[string]string{"QUALITY_RESULT": "failure"}, false},
		{"analysis cancelled", map[string]string{"CODEQL_RESULT": "cancelled"}, false},
		{"selected rehearsal skipped", map[string]string{"RELEASE_REHEARSAL_SELECTED": "true"}, false},
		{"selected rehearsal succeeded", map[string]string{"EVENT_NAME": "workflow_dispatch", "DEPENDENCY_REVIEW_RESULT": "skipped", "COMPATIBILITY_SCOPE_RESULT": "skipped", "RELEASE_REHEARSAL_SELECTED": "true", "RELEASE_REHEARSAL_RESULT": "success"}, true},
		{"selected diagnostic skipped", map[string]string{"NATIVE_CLI_DIAGNOSTIC_SELECTED": "true"}, false},
		{"selected diagnostic succeeded", map[string]string{"EVENT_NAME": "workflow_dispatch", "DEPENDENCY_REVIEW_RESULT": "skipped", "COMPATIBILITY_SCOPE_RESULT": "skipped", "NATIVE_CLI_DIAGNOSTIC_SELECTED": "true", "NATIVE_CLI_DIAGNOSTIC_RESULT": "success"}, true},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			environment := map[string]string{
				"EVENT_NAME": "pull_request", "DEPENDENCY_REVIEW_RESULT": "success", "QUALITY_RESULT": "success",
				"COMPATIBILITY_SCOPE_RESULT": "success", "COMPATIBILITY_CONSUMER_REQUIRED": "false", "PUBLICATION_REQUIRED": "false",
				"COMPATIBILITY_CONSUMER_RESULT": "skipped", "MILESTONE_SELECTED": "false", "MILESTONE_RESULT": "skipped",
				"CODEQL_RESULT": "success", "RELEASE_REHEARSAL_SELECTED": "false", "RELEASE_REHEARSAL_RESULT": "skipped",
				"NATIVE_CLI_DIAGNOSTIC_SELECTED": "false", "NATIVE_CLI_DIAGNOSTIC_RESULT": "skipped",
			}
			for key, value := range scenario.overrides {
				environment[key] = value
			}
			// Match the hosted default Bash errexit contract, not an interactive shell.
			command := exec.CommandContext(t.Context(), "bash", "-e", "-c", script)
			command.Env = os.Environ()
			for key, value := range environment {
				command.Env = append(command.Env, key+"="+value)
			}
			if output, err := command.CombinedOutput(); (err == nil) != scenario.pass {
				t.Fatalf("required pass = %t, want %t: %s", err == nil, scenario.pass, output)
			}
		})
	}
}
