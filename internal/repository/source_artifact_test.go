package repository_test

import (
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"go.yaml.in/yaml/v3"
)

func sourceActionSteps(t *testing.T, path string) []workflowStep {
	t.Helper()
	var action struct {
		Runs struct {
			Steps []workflowStep `yaml:"steps"`
		} `yaml:"runs"`
	}
	if err := yaml.Unmarshal([]byte(readProjectFile(t, path)), &action); err != nil {
		t.Fatal(err)
	}
	return action.Runs.Steps
}

func runSourceScript(t *testing.T, script, directory string, environment []string) {
	t.Helper()
	command := exec.CommandContext(t.Context(), "bash", "-c", script)
	command.Dir = directory
	command.Env = append(os.Environ(), environment...)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("source setup: %v: %s", err, output)
	}
}

func sourceEnvironment(t *testing.T, path string) []string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return strings.Split(strings.TrimSpace(string(data)), "\n")
}

func TestReusableSourceBootstrapBuildsOnceAcrossJobs(t *testing.T) {
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	var workflow workflowDocument
	if err := yaml.Unmarshal([]byte(readProjectFile(t, ".github/workflows/library-ci.yml")), &workflow); err != nil {
		t.Fatal(err)
	}
	fixture, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	bin := filepath.Join(fixture, "bin")
	if err := os.Mkdir(bin, 0o700); err != nil {
		t.Fatal(err)
	}
	log := filepath.Join(fixture, "builds")
	writeExecutable(t, filepath.Join(bin, "go"), `#!/usr/bin/env bash
set -euo pipefail
test "$1" = build
test -d "$GOCACHE" && test -d "$GOMODCACHE" && test -d "$GOTMPDIR"
printf 'build\n' >> "$SOURCE_BUILD_LOG"
while [[ "$1" != -o ]]; do shift; done
printf '#!/bin/sh\nprintf "dev\\n"\n' > "$2"
chmod 700 "$2"
`)
	var payload []byte
	var digest string
	for _, job := range []string{"prepare", "repository-contract", "quality", "codeql"} {
		workspace, runner := filepath.Join(fixture, job), filepath.Join(fixture, job+"-runner")
		source := filepath.Join(workspace, ".golib-tooling")
		if err := os.MkdirAll(source, 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.Mkdir(runner, 0o700); err != nil {
			t.Fatal(err)
		}
		runGit(t, source, "init", "-q")
		writeRehearsalFile(t, filepath.Join(source, "README.md"), "fixture\n")
		writeRehearsalFile(t, filepath.Join(source, "tools/reuse-source-golib.sh"), readProjectFile(t, "tools/reuse-source-golib.sh"))
		runGit(t, source, "add", "README.md", "tools/reuse-source-golib.sh")
		runGit(t, source, "-c", "core.hooksPath=/dev/null", "-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "-c", "commit.gpgsign=false", "commit", "-qm", "fixture")
		runGit(t, source, "remote", "add", "origin", "https://github.com/faustbrian/go-library-tools.git")
		sha := strings.TrimSpace(runGit(t, source, "rev-parse", "HEAD"))
		envFile, pathFile, outputFile := filepath.Join(runner, "env"), filepath.Join(runner, "path"), filepath.Join(runner, "output")
		for _, file := range []string{envFile, pathFile, outputFile} {
			writeRehearsalFile(t, file, "")
		}
		environment := []string{"PATH=" + bin + string(os.PathListSeparator) + os.Getenv("PATH"), "SOURCE_BUILD_LOG=" + log, "GITHUB_WORKSPACE=" + workspace, "RUNNER_TEMP=" + runner, "GITHUB_ENV=" + envFile, "GITHUB_PATH=" + pathFile, "GITHUB_OUTPUT=" + outputFile, "TOOLING_SHA=" + sha, "ARTIFACT_ID=123", "BINARY_SHA256=" + digest, "RUNNER_OS=Linux", "RUNNER_ARCH=X64"}
		var setup workflowStep
		for _, step := range workflow.Jobs[job].Steps {
			if step.If == "inputs.source_bootstrap == true" && (strings.Contains(step.Uses, "setup-") || step.Name == "Prepare same-run source installation") {
				setup = step
				break
			}
		}
		if setup.Uses == "" && setup.Run == "" {
			t.Fatalf("no source setup for %s", job)
		}
		if strings.HasPrefix(setup.Uses, "faustbrian/go-library-tools/.github/actions/setup-golib@") {
			for _, step := range sourceActionSteps(t, ".github/actions/setup-golib/action.yml") {
				if step.Name == "Build exact source golib" {
					runSourceScript(t, step.Run, workspace, environment)
				}
			}
		} else {
			if setup.Name != "Prepare same-run source installation" || job == "prepare" {
				t.Fatalf("unexpected source setup: %s", setup.Name)
			}
			// Resolve the actual workflow-to-action bindings, not test-owned values.
			outputs := map[string]string{
				"${{ inputs.tooling_sha }}":                         sha,
				"${{ needs.prepare.outputs.source_artifact_id }}":   "123",
				"${{ needs.prepare.outputs.source_binary_sha256 }}": digest,
			}
			for _, step := range workflow.Jobs[job].Steps {
				if step.If != "inputs.source_bootstrap == true" {
					continue
				}
				for key, value := range step.Env {
					environment = append(environment, key+"="+outputs[value])
				}
				if step.Run != "" {
					runSourceScript(t, step.Run, workspace, environment)
					environment = append(environment, sourceEnvironment(t, envFile)...)
				} else if strings.HasPrefix(step.Uses, "actions/download-artifact@") {
					// The only external seam: the service transfers the prepare payload.
					if step.With["artifact-ids"] != "${{ needs.prepare.outputs.source_artifact_id }}" || step.With["path"] != "${{ env.GOLIB_SOURCE_INSTALL_ROOT }}/bin" || step.With["merge-multiple"] != "true" || step.With["digest-mismatch"] != "error" || step.With["github-token"] != "" || step.With["repository"] != "" || step.With["run-id"] != "" || step.With["name"] != "" || step.With["pattern"] != "" {
						t.Fatal("transport does not select only prepare's immutable same-run artifact")
					}
					for _, assignment := range sourceEnvironment(t, envFile) {
						if root, found := strings.CutPrefix(assignment, "GOLIB_SOURCE_INSTALL_ROOT="); found {
							if err := os.WriteFile(filepath.Join(root, "bin", "golib"), payload, 0o600); err != nil {
								t.Fatal(err)
							}
						}
					}
				}
			}
		}
		environment = append(environment, sourceEnvironment(t, envFile)...)
		var install string
		for _, assignment := range sourceEnvironment(t, envFile) {
			if root, found := strings.CutPrefix(assignment, "GOLIB_SOURCE_INSTALL_ROOT="); found {
				install = root
			}
		}
		if job == "prepare" {
			payload, err = os.ReadFile(filepath.Join(install, "bin", "golib"))
			if err != nil {
				t.Fatal(err)
			}
			digest = fmt.Sprintf("%x", sha256.Sum256(payload))
			for _, step := range workflow.Jobs[job].Steps {
				if step.Name == "Bind source verifier binary" {
					runSourceScript(t, step.Run, workspace, environment)
					data, err := os.ReadFile(outputFile)
					if err != nil || !strings.Contains(string(data), "sha256="+digest+"\n") {
						t.Fatal("producer did not bind the transferred binary")
					}
				}
			}
		}
		runSourceScript(t, reusableStep(t, job, "", "Remove source bootstrap workspace").Run, workspace, environment)
		if _, err := os.Stat(install); !os.IsNotExist(err) {
			t.Fatal("job installation remains")
		}
	}
	data, err := os.ReadFile(log)
	if err != nil || string(data) != "build\n" {
		t.Fatalf("source verifier rebuilt across four jobs: %q (%v)", data, err)
	}
}

func TestSourceArtifactRefusesInvalidPayloadAndCleansJobResources(t *testing.T) {
	scriptSource := readProjectFile(t, "tools/reuse-source-golib.sh")
	installScript, prepareScript := "set -- install\n"+scriptSource, "set -- prepare\n"+scriptSource
	for _, scenario := range []string{"valid", "checksum mismatch", "wrong version", "symlink", "extra file", "oversized", "missing", "wrong platform", "invalid artifact", "invalid checksum"} {
		t.Run(scenario, func(t *testing.T) {
			runner := t.TempDir()
			root := filepath.Join(runner, "golib-source-install.fixture")
			if err := os.MkdirAll(filepath.Join(root, "bin"), 0o700); err != nil {
				t.Fatal(err)
			}
			binary := []byte("#!/bin/sh\nprintf 'dev\\n'\n")
			if scenario == "wrong version" {
				binary = []byte("#!/bin/sh\nprintf 'v1.0.0\\n'\n")
			}
			path := filepath.Join(root, "bin", "golib")
			if err := os.WriteFile(path, binary, 0o600); err != nil {
				t.Fatal(err)
			}
			digest := fmt.Sprintf("%x", sha256.Sum256(binary))
			switch scenario {
			case "checksum mismatch":
				digest = strings.Repeat("0", 64)
			case "symlink":
				if err := os.Rename(path, filepath.Join(runner, "other")); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(filepath.Join(runner, "other"), path); err != nil {
					t.Fatal(err)
				}
			case "extra file":
				writeRehearsalFile(t, filepath.Join(root, "bin", ".unexpected"), "extra")
			case "oversized":
				if err := os.Truncate(path, 64*1024*1024+1); err != nil {
					t.Fatal(err)
				}
			case "missing":
				if err := os.Remove(path); err != nil {
					t.Fatal(err)
				}
			}
			envFile, pathFile := filepath.Join(runner, "env"), filepath.Join(runner, "path")
			writeRehearsalFile(t, envFile, "")
			writeRehearsalFile(t, pathFile, "")
			environment := []string{"RUNNER_TEMP=" + runner, "GOLIB_SOURCE_INSTALL_ROOT=" + root, "GITHUB_ENV=" + envFile, "GITHUB_PATH=" + pathFile, "BINARY_SHA256=" + digest, "ARTIFACT_ID=123", "TOOLING_SHA=" + strings.Repeat("a", 40), "RUNNER_OS=Linux", "RUNNER_ARCH=X64"}
			script := installScript
			switch scenario {
			case "wrong platform":
				environment = append(environment, "RUNNER_ARCH=ARM64")
				script = prepareScript
			case "invalid artifact":
				environment = append(environment, "ARTIFACT_ID=")
				script = prepareScript
			case "invalid checksum":
				environment = append(environment, "BINARY_SHA256=not-a-checksum")
				script = prepareScript
			}
			command := exec.CommandContext(t.Context(), "bash", "-c", script)
			command.Env = append(os.Environ(), environment...)
			output, err := command.CombinedOutput()
			if (err == nil) != (scenario == "valid") {
				t.Fatalf("source artifact outcome: %v: %s", err, output)
			}
			if scenario == "valid" {
				command = exec.CommandContext(t.Context(), path, "--version")
				if output, err := command.CombinedOutput(); err != nil || string(output) != "dev\n" {
					t.Fatalf("verified binary is not usable: %v, %s", err, output)
				}
			} else if data, err := os.ReadFile(pathFile); err != nil || len(data) != 0 {
				t.Fatal("invalid binary was admitted to PATH")
			}
			runSourceScript(t, reusableStep(t, "quality", "", "Remove source bootstrap workspace").Run, runner, environment)
			if _, err := os.Stat(root); !os.IsNotExist(err) {
				t.Fatal("failed or successful installation resources remain")
			}
		})
	}
}
