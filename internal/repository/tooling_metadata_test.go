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

func toolingStep(t *testing.T, job, name string) workflowStep {
	t.Helper()
	var workflow workflowDocument
	if err := yaml.Unmarshal([]byte(readProjectFile(t, ".github/workflows/ci.yml")), &workflow); err != nil {
		t.Fatal(err)
	}
	for _, step := range workflow.Jobs[job].Steps {
		if step.Name == name {
			return step
		}
	}
	t.Fatalf("missing Tools workflow step %s / %s", job, name)
	return workflowStep{}
}

func TestToolingMetadataWorkflowRunsNoRuntimeContract(t *testing.T) {
	step := toolingStep(t, "quality", "Run fast local CI")
	for _, scenario := range []struct {
		name, runtime, result, commands string
		code                            int
	}{
		{"metadata", "false", "0", "metadata-ci\n", 0},
		{"runtime", "true", "0", "local-ci\n", 0},
		{"metadata failure", "false", "19", "metadata-ci\n", 19},
		{"runtime failure", "true", "19", "local-ci\n", 19},
		{"missing classification", "", "0", "", 1},
		{"invalid classification", "unknown", "0", "", 1},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			task := t.TempDir()
			log := filepath.Join(task, "commands")
			writeExecutable(t, filepath.Join(task, "make"), `#!/usr/bin/env bash
set -euo pipefail
printf '%s\n' "$*" >>"$MAKE_LOG"
exit "$MAKE_RESULT"
`)
			command := exec.CommandContext(t.Context(), "bash", "-e", "-c", step.Run)
			command.Env = append(os.Environ(), "PATH="+task+string(os.PathListSeparator)+os.Getenv("PATH"),
				"GOLIB_RUNTIME_REQUIRED="+scenario.runtime, "MAKE_RESULT="+scenario.result, "MAKE_LOG="+log)
			output, err := command.CombinedOutput()
			code := 0
			if err != nil {
				var failure *exec.ExitError
				if !errors.As(err, &failure) {
					t.Fatal(err)
				}
				code = failure.ExitCode()
			}
			if code != scenario.code {
				t.Fatalf("routing code = %d, want %d: %s", code, scenario.code, output)
			}
			data, err := os.ReadFile(log)
			if err != nil && !os.IsNotExist(err) {
				t.Fatal(err)
			}
			if string(data) != scenario.commands {
				t.Fatalf("workflow executed %q, want %q", data, scenario.commands)
			}
		})
	}
}

func TestToolingMetadataClassifierUsesActualGitAndFailsClosed(t *testing.T) {
	for _, job := range []string{"quality", "codeql", "reusable"} {
		var step workflowStep
		if job == "reusable" {
			step = reusableStep(t, "prepare", "scope", "")
		} else {
			step = toolingStep(t, job, "Select proportional assurance tier")
		}
		for _, scenario := range []struct {
			name, event, path, base, runtime string
			remove, pass                     bool
		}{
			{"docs PR", "pull_request", "README.md", "current", "false", false, true},
			{"docs main", "push", "README.md", "current", "false", false, true},
			{"notice main", "push", "NOTICE", "current", "false", false, true},
			{"license text", "push", "LICENSE.txt", "current", "false", false, true},
			{"LICENSE Go source", "pull_request", "LICENSE.go", "current", "true", false, true},
			{"NOTICE Go source", "push", "NOTICE.go", "current", "true", false, true},
			{"license shell source", "pull_request", "LICENSE.sh", "current", "true", false, true},
			{"notice Python source", "push", "NOTICE.py", "current", "true", false, true},
			{"unknown legal suffix", "pull_request", "NOTICE.data", "current", "true", false, true},
			{"executable Markdown", "push", "README.md", "current", "true", false, true},
			{"symlink Markdown", "pull_request", "notes.md", "current", "true", false, true},
			{"source PR", "pull_request", "source.go", "current", "true", false, true},
			{"source main", "push", "source.go", "current", "true", false, true},
			{"workflow trust", "pull_request", ".github/workflows/ci.yml", "current", "true", false, true},
			{"metadata policy", "pull_request", "policy.json", "current", "true", false, true},
			{"executable example", "pull_request", "examples/example.go", "current", "true", false, true},
			{"deleted source", "pull_request", "source.go", "current", "true", true, true},
			{"empty diff", "push", "", "current", "true", false, true},
			{"schedule", "schedule", "README.md", "", "true", false, true},
			{"manual", "workflow_dispatch", "README.md", "", "true", false, true},
			{"missing base", "pull_request", "README.md", "", "", false, false},
			{"malformed base", "push", "README.md", "bad", "", false, false},
			{"unavailable base", "pull_request", "README.md", strings.Repeat("a", 40), "", false, false},
			{"unsupported event", "unknown", "README.md", "current", "", false, false},
		} {
			t.Run(job+"/"+scenario.name, func(t *testing.T) {
				root, runner := t.TempDir(), t.TempDir()
				runGit(t, root, "init", "-q")
				runGit(t, root, "config", "user.name", "Fixture")
				runGit(t, root, "config", "user.email", "fixture@example.invalid")
				write(t, filepath.Join(root, "README.md"), "# Fixture\n")
				write(t, filepath.Join(root, "source.go"), "package fixture\n")
				runGit(t, root, "add", "README.md", "source.go")
				runGit(t, root, "commit", "-qm", "baseline")
				base := scenario.base
				if base == "current" {
					base = strings.TrimSpace(runGit(t, root, "rev-parse", "HEAD"))
				}
				if scenario.path != "" {
					if scenario.remove {
						if err := os.Remove(filepath.Join(root, scenario.path)); err != nil {
							t.Fatal(err)
						}
					} else if scenario.name == "symlink Markdown" {
						if err := os.Symlink("README.md", filepath.Join(root, scenario.path)); err != nil {
							t.Fatal(err)
						}
					} else {
						write(t, filepath.Join(root, scenario.path), "changed fixture\n")
						if scenario.name == "executable Markdown" {
							if err := os.Chmod(filepath.Join(root, scenario.path), 0o700); err != nil {
								t.Fatal(err)
							}
						}
					}
					runGit(t, root, "add", "--", scenario.path)
					runGit(t, root, "commit", "-qm", "change")
				}
				output := filepath.Join(t.TempDir(), "output")
				command := exec.CommandContext(t.Context(), "bash", "-c", step.Run)
				command.Dir = root
				command.Env = append(os.Environ(), "EVENT_NAME="+scenario.event, "BASE_SHA="+base,
					"GITHUB_OUTPUT="+output, "RUNNER_TEMP="+runner, "RELEASE_DRY_RUN=false")
				combined, err := command.CombinedOutput()
				if (err == nil) != scenario.pass {
					t.Fatalf("classification pass = %t, want %t: %s", err == nil, scenario.pass, combined)
				}
				data, err := os.ReadFile(output)
				if err != nil && !os.IsNotExist(err) {
					t.Fatal(err)
				}
				want := ""
				if scenario.pass {
					want = "runtime=" + scenario.runtime + "\n"
				}
				if string(data) != want {
					t.Fatalf("classification = %q, want %q", data, want)
				}
				remaining, err := os.ReadDir(runner)
				if err != nil || len(remaining) != 0 {
					t.Fatalf("classifier cleanup: entries=%v error=%v", remaining, err)
				}
			})
		}
	}
}

func TestToolingMetadataCodeQLRetainsApplicableAnalysis(t *testing.T) {
	var workflow workflowDocument
	if err := yaml.Unmarshal([]byte(readProjectFile(t, ".github/workflows/ci.yml")), &workflow); err != nil {
		t.Fatal(err)
	}
	scope := toolingStep(t, "quality", "Select proportional assurance tier")
	analysisScope := toolingStep(t, "codeql", "Select proportional assurance tier")
	if scope.Run != analysisScope.Run || scope.ID != "scope" || analysisScope.ID != "scope" {
		t.Fatal("quality and analysis do not share the same classifier")
	}
	// This is structural proof of job applicability, not a claim that CodeQL
	// extraction/queries have executed; actual classification is tested above.
	for _, step := range workflow.Jobs["codeql"].Steps {
		if step.Uses != "" && !strings.HasPrefix(step.Uses, "actions/checkout@") || step.Name == "Build runtime source" {
			if step.If != "steps.scope.outputs.runtime == 'true'" {
				t.Fatalf("analysis step %s/%s lacks applicable scope: %q", step.Name, step.Uses, step.If)
			}
		}
	}
}
