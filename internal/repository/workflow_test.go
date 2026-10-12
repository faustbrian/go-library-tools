package repository_test

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/faustbrian/go-library-tools/v2/internal/config"
	"go.yaml.in/yaml/v3"
)

var remoteAction = regexp.MustCompile(`(?m)^\s*- uses: ([^./][^@\s]+)@([^\s#]+)`)
var immutableAction = regexp.MustCompile(`^[0-9a-f]{40}$`)

func TestReusableWorkflowCleanupRestoresTemporaryDirectory(t *testing.T) {
	var workflow workflowDocument
	if err := yaml.Unmarshal([]byte(readProjectFile(t, ".github/workflows/library-ci.yml")), &workflow); err != nil {
		t.Fatal(err)
	}
	node, err := exec.LookPath("node")
	if err != nil {
		t.Fatal(err)
	}
	for _, job := range []string{"prepare", "repository-contract", "quality", "codeql"} {
		for _, present := range []bool{true, false} {
			t.Run(job+"/present="+strconv.FormatBool(present), func(t *testing.T) {
				var script string
				for _, step := range workflow.Jobs[job].Steps {
					if step.Name == "Remove source bootstrap workspace" {
						script = step.Run
					}
				}
				if script == "" {
					t.Fatal("missing cleanup")
				}
				runner := t.TempDir()
				install := filepath.Join(runner, "golib-source-install.fixture")
				if present {
					for _, directory := range []string{"bin", "lint", "xdg", "npm", "tmp"} {
						path := filepath.Join(install, directory)
						if err := os.MkdirAll(path, 0o700); err != nil {
							t.Fatal(err)
						}
						if err := os.WriteFile(filepath.Join(path, "owned"), []byte("fixture"), 0o600); err != nil {
							t.Fatal(err)
						}
					}
				}
				environment := filepath.Join(runner, "github-env")
				if err := os.WriteFile(environment, nil, 0o600); err != nil {
					t.Fatal(err)
				}
				command := exec.CommandContext(t.Context(), "bash", "-c", script)
				command.Env = append(os.Environ(), "RUNNER_TEMP="+runner,
					"GOLIB_SOURCE_INSTALL_ROOT="+install, "GITHUB_ENV="+environment)
				if combined, err := command.CombinedOutput(); err != nil {
					t.Fatalf("cleanup: %v\n%s", err, combined)
				}
				if _, err := os.Stat(install); !os.IsNotExist(err) {
					t.Fatalf("installation remains: %v", err)
				}
				data, err := os.ReadFile(environment)
				if err != nil {
					t.Fatal(err)
				}
				post := exec.CommandContext(t.Context(), node, "-e", "process.stdout.write(require('node:fs').realpathSync(require('node:os').tmpdir()))")
				post.Env = append(os.Environ(), "TMPDIR="+filepath.Join(install, "tmp"))
				if updates := strings.TrimSuffix(string(data), "\n"); updates != "" {
					post.Env = append(post.Env, strings.Split(updates, "\n")...)
				}
				expected, err := filepath.EvalSymlinks(runner)
				if err != nil {
					t.Fatal(err)
				}
				combined, err := post.CombinedOutput()
				if err != nil || string(combined) != expected {
					t.Fatalf("post temporary directory: %v, output=%q", err, combined)
				}
			})
		}
	}
}

func TestReusableWorkflowPersistsPublicDependencyAuthority(t *testing.T) {
	var workflow workflowDocument
	if err := yaml.Unmarshal([]byte(readProjectFile(t, ".github/workflows/library-ci.yml")), &workflow); err != nil {
		t.Fatal(err)
	}
	for _, job := range []string{"prepare", "repository-contract", "quality", "codeql"} {
		t.Run(job, func(t *testing.T) {
			var script string
			for _, step := range workflow.Jobs[job].Steps {
				if step.Name == "Select public dependency authority" {
					script = step.Run
				}
			}
			if script == "" {
				t.Fatal("no persistent public dependency selection")
			}
			output := filepath.Join(t.TempDir(), "github-env")
			command := exec.CommandContext(t.Context(), "bash", "-c", script)
			command.Env = append(os.Environ(), "GITHUB_ENV="+output,
				"GOPROXY=file:///unavailable,direct", "GOSUMDB=off",
				"GOPRIVATE=*", "GONOPROXY=*", "GONOSUMDB=*")
			if combined, err := command.CombinedOutput(); err != nil {
				t.Fatalf("public dependency selection: %v\n%s", err, combined)
			}
			contents, err := os.ReadFile(output)
			if err != nil {
				t.Fatal(err)
			}
			want := "GOPROXY=https://proxy.golang.org\nGOSUMDB=sum.golang.org\nGOPRIVATE=\nGONOPROXY=\nGONOSUMDB=\n"
			if string(contents) != want {
				t.Fatalf("persisted public authority = %q, want %q", contents, want)
			}
		})
	}
}

func TestRepositoryWorkflowsPinRemoteActions(t *testing.T) {
	root := projectRoot(t)
	entries, err := os.ReadDir(filepath.Join(root, ".github", "workflows"))
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".yml" {
			continue
		}
		path := filepath.Join(root, ".github", "workflows", entry.Name())
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			t.Fatal(readErr)
		}
		matches := remoteAction.FindAllStringSubmatch(string(data), -1)
		if len(matches) == 0 {
			t.Fatalf("%s has no pinned remote actions", entry.Name())
		}
		for _, match := range matches {
			if !immutableAction.MatchString(match[2]) {
				t.Errorf("%s action %s uses mutable ref %s", entry.Name(), match[1], match[2])
			}
		}
	}
}

func TestReusableWorkflowPreservesConsumerContract(t *testing.T) {
	content := readProjectFile(t, ".github/workflows/library-ci.yml")
	for _, required := range []string{
		"workflow_call:",
		"tooling_sha:",
		"security-events: write",
		"golib repository check",
		"golib workflows check",
		"golib specification check",
		"golib specification check --online",
		"arguments=(check --local)",
		"github/codeql-action/init@",
		"github/codeql-action/analyze@",
		"name: Required",
		"if: always()",
		".verification",
	} {
		if !strings.Contains(content, required) {
			t.Errorf("reusable workflow lacks %q", required)
		}
	}
	if strings.Contains(content, "curl |") || strings.Contains(content, "@main") {
		t.Fatal("reusable workflow contains an unsafe bootstrap reference")
	}
	if strings.Contains(content, "packages: read") {
		t.Fatal("reusable workflow requests package access that consumer callers do not grant")
	}
}

func TestReusableWorkflowKeepsOnlineSpecificationMonitoringOutOfOrdinaryPullRequests(t *testing.T) {
	var workflow workflowDocument
	if err := yaml.Unmarshal([]byte(readProjectFile(t, ".github/workflows/library-ci.yml")), &workflow); err != nil {
		t.Fatal(err)
	}
	offline := 0
	online := 0
	for _, step := range workflow.Jobs["repository-contract"].Steps {
		switch strings.TrimSpace(step.Run) {
		case "golib specification check":
			if step.If != "" {
				t.Fatalf("offline specification check condition = %q, want unconditional", step.If)
			}
			offline++
		case "golib specification check --online":
			if step.If != "github.event_name == 'schedule'" {
				t.Fatalf("online specification check condition = %q, want scheduled monitoring only", step.If)
			}
			online++
		}
	}
	if offline != 1 || online != 1 {
		t.Fatalf("specification steps = offline %d, online %d; want one of each", offline, online)
	}
}

func TestReusableWorkflowBuildsReleaseModuleMatrix(t *testing.T) {
	var workflow workflowDocument
	if err := yaml.Unmarshal([]byte(readProjectFile(t, ".github/workflows/library-ci.yml")), &workflow); err != nil {
		t.Fatal(err)
	}
	input, exists := workflow.On.WorkflowCall.Inputs["release_module"]
	if !exists || input.Type != "string" || input.Default != "" {
		t.Fatalf("release_module workflow input = %#v, present = %v", input, exists)
	}
	var matrixScript string
	for _, step := range workflow.Jobs["prepare"].Steps {
		if step.ID == "modules" {
			if step.Env["RELEASE_MODULE"] != "${{ inputs.release_module }}" || step.Env["RELEASE_DRY_RUN"] != "${{ inputs.release_dry_run }}" {
				t.Fatalf("module-matrix environment = %#v", step.Env)
			}
			matrixScript = step.Run
			break
		}
	}
	if matrixScript == "" {
		t.Fatal("reusable workflow has no executable module-matrix step")
	}

	bin := t.TempDir()
	golib := filepath.Join(bin, "golib")
	stub := `#!/bin/sh
case "$*" in
  '--help')
    if [ "$LEGACY_RELEASE" = "true" ]; then
      printf '%s\n' 'golib release dry-run'
    else
      printf '%s\n' 'golib release dry-run [--all|--module <directory>]'
    fi
    ;;
  'inventory --json') printf '%s\n' "$INVENTORY" ;;
  'config show --json') printf '%s\n' '{"runtimes":{}}' ;;
  *) exit 64 ;;
esac
`
	if err := os.WriteFile(golib, []byte(stub), 0o700); err != nil {
		t.Fatal(err)
	}
	inventory := `{"modules":[{"directory":".","releasable":true},{"directory":"nested","releasable":true},{"directory":"fixture","releasable":false}]}`
	tests := []struct {
		name      string
		dryRun    string
		module    string
		legacy    string
		want      []string
		wantError string
	}{
		{name: "default", dryRun: "false", want: []string{".", "nested", "fixture"}},
		{name: "all releasable", dryRun: "true", want: []string{".", "nested"}},
		{name: "selected releasable", dryRun: "true", module: "nested", want: []string{"nested"}},
		{name: "legacy all releasable runs once", dryRun: "true", legacy: "true", want: []string{"."}},
		{name: "legacy selected module widens and runs once", dryRun: "true", module: "nested", legacy: "true", want: []string{"."}},
		{name: "unknown selection", dryRun: "true", module: "missing", wantError: "release module is unknown or not releasable: missing"},
		{name: "non-releasable selection", dryRun: "true", module: "fixture", wantError: "release module is unknown or not releasable: fixture"},
		{name: "legacy unknown selection", dryRun: "true", module: "missing", legacy: "true", wantError: "release module is unknown or not releasable: missing"},
		{name: "legacy non-releasable selection", dryRun: "true", module: "fixture", legacy: "true", wantError: "release module is unknown or not releasable: fixture"},
		{name: "selection outside rehearsal", dryRun: "false", module: "nested", wantError: "release_module requires release_dry_run: true"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			output := filepath.Join(t.TempDir(), "github-output")
			command := exec.CommandContext(t.Context(), "bash", "-c", matrixScript)
			command.Env = append(os.Environ(),
				"PATH="+bin+string(os.PathListSeparator)+os.Getenv("PATH"),
				"INVENTORY="+inventory,
				"GITHUB_OUTPUT="+output,
				"RELEASE_DRY_RUN="+test.dryRun,
				"RELEASE_MODULE="+test.module,
				"LEGACY_RELEASE="+test.legacy,
			)
			combined, err := command.CombinedOutput()
			if test.wantError != "" {
				if err == nil || !strings.Contains(string(combined), test.wantError) {
					t.Fatalf("matrix step error = %v, output = %q", err, combined)
				}
				return
			}
			if err != nil {
				t.Fatalf("matrix step error = %v, output = %q", err, combined)
			}
			contents, err := os.ReadFile(output)
			if err != nil {
				t.Fatal(err)
			}
			line, _, _ := strings.Cut(string(contents), "\n")
			encoded, found := strings.CutPrefix(line, "matrix=")
			if !found {
				t.Fatalf("matrix output = %q", contents)
			}
			var entries []struct {
				Directory string `json:"directory"`
			}
			if err := json.Unmarshal([]byte(encoded), &entries); err != nil {
				t.Fatal(err)
			}
			got := make([]string, 0, len(entries))
			for _, entry := range entries {
				got = append(got, entry.Directory)
			}
			if strings.Join(got, ",") != strings.Join(test.want, ",") {
				t.Fatalf("matrix directories = %v, want %v", got, test.want)
			}
		})
	}
}

func TestReusableWorkflowSelectsRuntimeWorkFromChangedPaths(t *testing.T) {
	var workflow workflowDocument
	if err := yaml.Unmarshal([]byte(readProjectFile(t, ".github/workflows/library-ci.yml")), &workflow); err != nil {
		t.Fatal(err)
	}
	prepare := workflow.Jobs["prepare"]
	var scopeScript string
	for _, step := range prepare.Steps {
		if step.ID == "scope" {
			scopeScript = step.Run
			break
		}
	}
	if scopeScript == "" {
		t.Fatal("reusable workflow has no executable change-scope selector")
	}
	if workflow.Jobs["quality"].If != "needs.prepare.outputs.runtime == 'true'" {
		t.Fatalf("quality condition = %q", workflow.Jobs["quality"].If)
	}
	if workflow.Jobs["codeql"].If != "needs.prepare.outputs.runtime == 'true'" {
		t.Fatalf("CodeQL condition = %q", workflow.Jobs["codeql"].If)
	}

	repository := t.TempDir()
	run := func(t *testing.T, event, dryRun, changedPath string, rename bool, pinMode string, sourceBootstrap bool) string {
		t.Helper()
		if combined, err := exec.CommandContext(t.Context(), "git", "init", "-q", repository).CombinedOutput(); err != nil {
			t.Fatalf("git init: %v\n%s", err, combined)
		}
		command := exec.CommandContext(t.Context(), "git", "config", "user.email", "fixture@example.com")
		command.Dir = repository
		if combined, err := command.CombinedOutput(); err != nil {
			t.Fatalf("git config: %v\n%s", err, combined)
		}
		command = exec.CommandContext(t.Context(), "git", "config", "user.name", "Fixture")
		command.Dir = repository
		if combined, err := command.CombinedOutput(); err != nil {
			t.Fatalf("git config: %v\n%s", err, combined)
		}
		baseline := filepath.Join(repository, "baseline.go")
		if err := os.WriteFile(baseline, []byte("package fixture\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		if changedPath == ".github/workflows/ci.yml" {
			workflow := filepath.Join(repository, changedPath)
			if err := os.MkdirAll(filepath.Dir(workflow), 0o700); err != nil {
				t.Fatal(err)
			}
			comment := " # v1.4.0"
			if strings.HasPrefix(pinMode, "legacy-comment") {
				comment = " # v1.4.0 specification monitoring"
			}
			content := "jobs:\n  ci:\n    uses: faustbrian/go-library-tools/.github/workflows/library-ci.yml@" + strings.Repeat("a", 40) + comment + "\n    with:\n      tooling_sha: " + strings.Repeat("a", 40) + "\n"
			if err := os.WriteFile(workflow, []byte(content), 0o600); err != nil {
				t.Fatal(err)
			}
		}
		addArguments := []string{"add", "baseline.go"}
		if changedPath == ".github/workflows/ci.yml" {
			addArguments = append(addArguments, changedPath)
		}
		command = exec.CommandContext(t.Context(), "git", addArguments...)
		command.Dir = repository
		if combined, err := command.CombinedOutput(); err != nil {
			t.Fatalf("git add: %v\n%s", err, combined)
		}
		command = exec.CommandContext(t.Context(), "git", "-c", "commit.gpgsign=false", "commit", "-q", "-m", "baseline")
		command.Dir = repository
		if combined, err := command.CombinedOutput(); err != nil {
			t.Fatalf("git commit: %v\n%s", err, combined)
		}
		command = exec.CommandContext(t.Context(), "git", "rev-parse", "HEAD")
		command.Dir = repository
		baseBytes, err := command.Output()
		if err != nil {
			t.Fatal(err)
		}
		base := strings.TrimSpace(string(baseBytes))
		path := filepath.Join(repository, changedPath)
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatal(err)
		}
		if rename {
			command = exec.CommandContext(t.Context(), "git", "mv", "baseline.go", changedPath)
			command.Dir = repository
			if combined, err := command.CombinedOutput(); err != nil {
				t.Fatalf("git mv: %v\n%s", err, combined)
			}
		} else {
			content := "changed\n"
			if pinMode != "" {
				usesSHA := strings.Repeat("b", 40)
				inputSHA := usesSHA
				switch pinMode {
				case "tooling-only":
					usesSHA = strings.Repeat("a", 40)
				case "mismatched":
					inputSHA = strings.Repeat("c", 40)
				case "legacy-comment-mismatched":
					inputSHA = strings.Repeat("c", 40)
				}
				content = "jobs:\n  ci:\n    uses: faustbrian/go-library-tools/.github/workflows/library-ci.yml@" + usesSHA + " # v1.7.1\n    with:\n      tooling_sha: " + inputSHA + "\n"
				if err := os.WriteFile(filepath.Join(repository, "AGENTS.md"), []byte("policy update\n"), 0o600); err != nil {
					t.Fatal(err)
				}
			} else if changedPath == ".github/workflows/ci.yml" {
				content = "permissions: write-all\njobs:\n  ci:\n    uses: faustbrian/go-library-tools/.github/workflows/library-ci.yml@" + strings.Repeat("a", 40) + " # v1.4.0\n    with:\n      tooling_sha: " + strings.Repeat("a", 40) + "\n"
			}
			if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
				t.Fatal(err)
			}
		}
		changedArguments := []string{"add", changedPath}
		if pinMode != "" {
			changedArguments = append(changedArguments, "AGENTS.md")
		}
		command = exec.CommandContext(t.Context(), "git", changedArguments...)
		command.Dir = repository
		if combined, err := command.CombinedOutput(); err != nil {
			t.Fatalf("git add: %v\n%s", err, combined)
		}
		command = exec.CommandContext(t.Context(), "git", "-c", "commit.gpgsign=false", "commit", "-q", "-m", "change")
		command.Dir = repository
		if combined, err := command.CombinedOutput(); err != nil {
			t.Fatalf("git commit: %v\n%s", err, combined)
		}
		output := filepath.Join(t.TempDir(), "github-output")
		command = exec.CommandContext(t.Context(), "bash", "-c", scopeScript)
		command.Dir = repository
		command.Env = append(os.Environ(),
			"BASE_SHA="+base,
			"EVENT_NAME="+event,
			"GITHUB_OUTPUT="+output,
			"RELEASE_DRY_RUN="+dryRun,
			"SOURCE_BOOTSTRAP="+strconv.FormatBool(sourceBootstrap),
		)
		if combined, err := command.CombinedOutput(); err != nil {
			t.Fatalf("scope selector: %v\n%s", err, combined)
		}
		contents, err := os.ReadFile(output)
		if err != nil {
			t.Fatal(err)
		}
		return strings.TrimSpace(strings.TrimPrefix(string(contents), "runtime="))
	}

	for _, test := range []struct {
		name, event, dryRun, path, want string
		rename                          bool
		pinMode                         string
		sourceBootstrap                 bool
	}{
		{name: "policy only", event: "pull_request", dryRun: "false", path: "AGENTS.md", want: "false"},
		{name: "documentation only", event: "pull_request", dryRun: "false", path: "docs/usage.md", want: "false"},
		{name: "policy and reusable workflow pin only", event: "pull_request", dryRun: "false", path: ".github/workflows/ci.yml", want: "false", pinMode: "matched"},
		{name: "policy and legacy-commented reusable workflow pin only", event: "pull_request", dryRun: "false", path: ".github/workflows/ci.yml", want: "false", pinMode: "legacy-comment"},
		{name: "one-sided reusable workflow pin", event: "pull_request", dryRun: "false", path: ".github/workflows/ci.yml", want: "true", pinMode: "tooling-only"},
		{name: "mismatched reusable workflow pins", event: "pull_request", dryRun: "false", path: ".github/workflows/ci.yml", want: "true", pinMode: "mismatched"},
		{name: "legacy-commented mismatched reusable workflow pins", event: "pull_request", dryRun: "false", path: ".github/workflows/ci.yml", want: "true", pinMode: "legacy-comment-mismatched"},
		{name: "substantive workflow change", event: "pull_request", dryRun: "false", path: ".github/workflows/ci.yml", want: "true"},
		{name: "structured documentation metadata", event: "pull_request", dryRun: "false", path: "docs/ecosystem/compatibility-sets.json", want: "true"},
		{name: "source renamed to documentation", event: "pull_request", dryRun: "false", path: "docs/baseline.md", want: "true", rename: true},
		{name: "source", event: "pull_request", dryRun: "false", path: "internal/example/example.go", want: "true"},
		{name: "release rehearsal", event: "workflow_dispatch", dryRun: "true", path: "AGENTS.md", want: "true"},
		{name: "main push", event: "push", dryRun: "false", path: "AGENTS.md", want: "false"},
		{name: "main source push", event: "push", dryRun: "false", path: "internal/example/example.go", want: "true"},
		{name: "main matched pins", event: "push", dryRun: "false", path: ".github/workflows/ci.yml", want: "false", pinMode: "matched"},
		{name: "source bootstrap documentation", event: "pull_request", dryRun: "false", path: "docs/usage.md", want: "false", sourceBootstrap: true},
		{name: "source bootstrap license", event: "pull_request", dryRun: "false", path: "LICENSE", want: "false", sourceBootstrap: true},
		{name: "source bootstrap source", event: "pull_request", dryRun: "false", path: "internal/example/example.go", want: "true", sourceBootstrap: true},
		{name: "source bootstrap structured metadata", event: "pull_request", dryRun: "false", path: "docs/config.json", want: "true", sourceBootstrap: true},
		{name: "source bootstrap renamed source", event: "pull_request", dryRun: "false", path: "docs/baseline.md", want: "true", rename: true, sourceBootstrap: true},
		{name: "source bootstrap release rehearsal", event: "workflow_dispatch", dryRun: "true", path: "docs/usage.md", want: "true", sourceBootstrap: true},
		{name: "source bootstrap push", event: "push", dryRun: "false", path: "docs/usage.md", want: "false", sourceBootstrap: true},
		{name: "source bootstrap scheduled", event: "schedule", dryRun: "false", path: "docs/usage.md", want: "true", sourceBootstrap: true},
		{name: "source bootstrap manual", event: "workflow_dispatch", dryRun: "false", path: "docs/usage.md", want: "true", sourceBootstrap: true},
		{name: "source bootstrap matched pins", event: "pull_request", dryRun: "false", path: ".github/workflows/ci.yml", want: "false", pinMode: "matched", sourceBootstrap: true},
		{name: "source bootstrap main matched pins", event: "push", dryRun: "false", path: ".github/workflows/ci.yml", want: "false", pinMode: "matched", sourceBootstrap: true},
		{name: "newline source path", event: "pull_request", dryRun: "false", path: "internal/example/new\nsource.go", want: "true"},
	} {
		t.Run(test.name, func(t *testing.T) {
			repository = t.TempDir()
			if got := run(t, test.event, test.dryRun, test.path, test.rename, test.pinMode, test.sourceBootstrap); got != test.want {
				t.Fatalf("runtime = %q, want %q", got, test.want)
			}
		})
	}
}

func TestReleaseModuleSelectorFlowsThroughHostedReleasePaths(t *testing.T) {
	var reusable workflowDocument
	if err := yaml.Unmarshal([]byte(readProjectFile(t, ".github/workflows/library-ci.yml")), &reusable); err != nil {
		t.Fatal(err)
	}
	assertReleaseStep := func(job, name, expectedIf, expectedInput, expectedCommand string) string {
		t.Helper()
		for _, step := range reusable.Jobs[job].Steps {
			if step.Name == name {
				if step.If != expectedIf || step.Env["RELEASE_MODULE"] != expectedInput || !strings.Contains(step.Run, expectedCommand) {
					t.Fatalf("%s/%s does not preserve release routing: if=%q env=%#v run=%q", job, name, step.If, step.Env, step.Run)
				}
				return step.Run
			}
		}
		t.Fatalf("%s has no %q step", job, name)
		return ""
	}
	releaseCheckScript := assertReleaseStep("repository-contract", "Validate release contract", "inputs.release_dry_run == true", "${{ inputs.release_module }}", `golib release check`)
	qualityRehearsalScript := assertReleaseStep("quality", "Run release rehearsal", "inputs.release_dry_run == true", "${{ inputs.release_module }}", `golib release dry-run`)
	rehearsalOwners := make([]string, 0, 1)
	for job, definition := range reusable.Jobs {
		for _, step := range definition.Steps {
			if step.Name == "Run release rehearsal" {
				rehearsalOwners = append(rehearsalOwners, job)
			}
		}
	}
	if len(rehearsalOwners) != 1 || rehearsalOwners[0] != "quality" {
		t.Fatalf("release rehearsal owners = %v, want one quality owner", rehearsalOwners)
	}
	ordinaryContract := false
	for _, step := range reusable.Jobs["quality"].Steps {
		if step.Name == "Run repository module contracts" {
			ordinaryContract = true
			if step.If != "inputs.release_dry_run != true" {
				t.Fatalf("ordinary module contract guard = %q", step.If)
			}
		}
	}
	if !ordinaryContract {
		t.Fatal("quality job has no ordinary module contract step")
	}

	golibBin := t.TempDir()
	golibInvocations := filepath.Join(t.TempDir(), "golib-invocations")
	golibStub := `#!/bin/sh
if [ "$1" = "--help" ]; then
	if [ "$LEGACY" = "true" ]; then
		printf '%s\n' 'golib release check'
		printf '%s\n' 'golib release dry-run'
	else
		printf '%s\n' 'golib release check [--all|--module <directory>]'
		printf '%s\n' 'golib release dry-run [--all|--module <directory>]'
	fi
	exit 0
fi
printf '%s\n' "$*" >>"$INVOCATIONS"
`
	if err := os.WriteFile(filepath.Join(golibBin, "golib"), []byte(golibStub), 0o700); err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name, module, legacy, want string
	}{
		{name: "blank", want: "release check"},
		{name: "selected current tool", module: "nested", want: "release check --module nested"},
		{name: "selected v1.4 tool", module: "nested", legacy: "true", want: "release check"},
	} {
		t.Run("reusable release check "+test.name, func(t *testing.T) {
			if err := os.WriteFile(golibInvocations, nil, 0o600); err != nil {
				t.Fatal(err)
			}
			command := exec.CommandContext(t.Context(), "bash", "-c", releaseCheckScript)
			command.Env = append(os.Environ(),
				"PATH="+golibBin+string(os.PathListSeparator)+os.Getenv("PATH"),
				"INVOCATIONS="+golibInvocations,
				"RELEASE_MODULE="+test.module,
				"LEGACY="+test.legacy,
			)
			if combined, err := command.CombinedOutput(); err != nil {
				t.Fatalf("release check error = %v, output = %q", err, combined)
			}
			contents, err := os.ReadFile(golibInvocations)
			if err != nil {
				t.Fatal(err)
			}
			if strings.TrimSpace(string(contents)) != test.want {
				t.Fatalf("golib invocation = %q, want %q", contents, test.want)
			}
		})
	}
	for _, test := range []struct {
		name, legacy, want string
	}{
		{name: "current tool", want: "release dry-run --module nested"},
		{name: "v1.4 tool", legacy: "true", want: "release dry-run"},
	} {
		t.Run("quality release rehearsal "+test.name, func(t *testing.T) {
			if err := os.WriteFile(golibInvocations, nil, 0o600); err != nil {
				t.Fatal(err)
			}
			qualityCommand := exec.CommandContext(t.Context(), "bash", "-c", qualityRehearsalScript)
			qualityCommand.Env = append(os.Environ(),
				"PATH="+golibBin+string(os.PathListSeparator)+os.Getenv("PATH"),
				"INVOCATIONS="+golibInvocations,
				"RELEASE_MODULE=nested",
				"LEGACY="+test.legacy,
			)
			if combined, err := qualityCommand.CombinedOutput(); err != nil {
				t.Fatalf("quality release rehearsal error = %v, output = %q", err, combined)
			}
			qualityInvocation, err := os.ReadFile(golibInvocations)
			if err != nil {
				t.Fatal(err)
			}
			if got := strings.TrimSpace(string(qualityInvocation)); got != test.want {
				t.Fatalf("quality golib invocation = %q, want %q", got, test.want)
			}
		})
	}

	var caller workflowDocument
	if err := yaml.Unmarshal([]byte(readProjectFile(t, ".github/workflows/ci.yml")), &caller); err != nil {
		t.Fatal(err)
	}
	input, exists := caller.On.WorkflowDispatch.Inputs["release_module"]
	if !exists || input.Type != "string" || input.Default != "" {
		t.Fatalf("tooling release_module dispatch input = %#v, present = %v", input, exists)
	}
	rehearsal := caller.Jobs["release-rehearsal"]
	if rehearsal.If != "github.event_name == 'workflow_dispatch' && (inputs.release_rehearsal || inputs.release_module != '')" {
		t.Fatalf("tooling release rehearsal condition = %q", rehearsal.If)
	}
	var validationScript, rehearsalScript string
	for _, step := range rehearsal.Steps {
		switch step.Name {
		case "Validate release inputs":
			if step.Env["RELEASE_MODULE"] != "${{ inputs.release_module }}" || step.Env["RELEASE_REHEARSAL"] != "${{ inputs.release_rehearsal }}" {
				t.Fatalf("tooling input validation environment = %#v", step.Env)
			}
			validationScript = step.Run
		case "Run pre-tag release dry-run":
			if step.Env["RELEASE_MODULE"] != "${{ inputs.release_module }}" || !strings.Contains(step.Run, `release dry-run --module "${RELEASE_MODULE}"`) {
				t.Fatalf("tooling rehearsal does not propagate the release selector: env=%#v run=%q", step.Env, step.Run)
			}
			rehearsalScript = step.Run
		}
	}
	if validationScript == "" || rehearsalScript == "" {
		t.Fatalf("tooling workflow release scripts missing: validation=%v rehearsal=%v", validationScript != "", rehearsalScript != "")
	}

	for _, test := range []struct {
		name      string
		rehearsal string
		module    string
		wantError string
	}{
		{name: "blank", rehearsal: "true"},
		{name: "selected", rehearsal: "true", module: "nested"},
		{name: "selector without rehearsal", rehearsal: "false", module: "nested", wantError: "release_module requires release_rehearsal: true"},
	} {
		t.Run("input "+test.name, func(t *testing.T) {
			command := exec.CommandContext(t.Context(), "bash", "-c", validationScript)
			command.Env = append(os.Environ(), "RELEASE_REHEARSAL="+test.rehearsal, "RELEASE_MODULE="+test.module)
			combined, err := command.CombinedOutput()
			if test.wantError == "" && err != nil {
				t.Fatalf("input validation error = %v, output = %q", err, combined)
			}
			if test.wantError != "" && (err == nil || !strings.Contains(string(combined), test.wantError)) {
				t.Fatalf("input validation error = %v, output = %q", err, combined)
			}
		})
	}

	bin := t.TempDir()
	invocations := filepath.Join(t.TempDir(), "go-invocations")
	stub := "#!/bin/sh\nprintf '%s\\n' \"$*\" >>\"$INVOCATIONS\"\n"
	if err := os.WriteFile(filepath.Join(bin, "go"), []byte(stub), 0o700); err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name   string
		module string
		want   string
	}{
		{name: "blank", want: "run ./cmd/golib release dry-run"},
		{name: "selected", module: "nested", want: "run ./cmd/golib release dry-run --module nested"},
	} {
		t.Run("rehearsal "+test.name, func(t *testing.T) {
			if err := os.WriteFile(invocations, nil, 0o600); err != nil {
				t.Fatal(err)
			}
			command := exec.CommandContext(t.Context(), "bash", "-c", rehearsalScript)
			command.Env = append(os.Environ(),
				"PATH="+bin+string(os.PathListSeparator)+os.Getenv("PATH"),
				"INVOCATIONS="+invocations,
				"RELEASE_MODULE="+test.module,
				"RUNNER_TEMP="+t.TempDir(),
			)
			if combined, err := command.CombinedOutput(); err != nil {
				t.Fatalf("release rehearsal error = %v, output = %q", err, combined)
			}
			contents, err := os.ReadFile(invocations)
			if err != nil {
				t.Fatal(err)
			}
			if strings.TrimSpace(string(contents)) != test.want {
				t.Fatalf("go invocation = %q, want %q", contents, test.want)
			}
		})
	}
}

func TestReusableWorkflowAdaptsModuleCheckToInstalledToolCapability(t *testing.T) {
	var reusable workflowDocument
	if err := yaml.Unmarshal([]byte(readProjectFile(t, ".github/workflows/library-ci.yml")), &reusable); err != nil {
		t.Fatal(err)
	}

	var contract workflowStep
	for _, step := range reusable.Jobs["quality"].Steps {
		if step.Name == "Run repository module contracts" {
			contract = step
			break
		}
	}
	if contract.Run == "" {
		t.Fatal("quality job has no module contract step")
	}

	golibBin := t.TempDir()
	invocations := filepath.Join(t.TempDir(), "golib-invocations")
	stub := `#!/bin/sh
if [ "$1" = "--help" ]; then
	if [ "$LEGACY" = "true" ]; then
		printf '%s\n' 'golib check [--all|--module <directory>]'
	else
		printf '%s\n' 'golib check [--local] [--all|--module <directory>]'
		printf '%s\n' 'golib check --local --module <directory> [--module <directory> ...]'
	fi
	exit 0
fi
printf '%s\n' "$*" >>"$INVOCATIONS"
exit "${GOLIB_ROUTING_FIXTURE_STATUS:-0}"
`
	if err := os.WriteFile(filepath.Join(golibBin, "golib"), []byte(stub), 0o700); err != nil {
		t.Fatal(err)
	}

	for _, test := range []struct {
		name            string
		legacy          string
		want            string
		sourceBootstrap bool
		wantError       bool
	}{
		{name: "v1.4 tool", legacy: "true", wantError: true},
		{name: "v1.5 tool", legacy: "true", wantError: true},
		{name: "local capable tool", legacy: "false", want: "check --local --module ."},
		{name: "source legacy tool", legacy: "true", wantError: true, sourceBootstrap: true},
		{name: "source local capable tool", legacy: "false", want: "check --local --module .", sourceBootstrap: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			if err := os.WriteFile(invocations, nil, 0o600); err != nil {
				t.Fatal(err)
			}
			command := exec.CommandContext(t.Context(), "bash", "-euo", "pipefail", "-c", contract.Run)
			command.Env = append(os.Environ(),
				"PATH="+golibBin+string(os.PathListSeparator)+os.Getenv("PATH"),
				"INVOCATIONS="+invocations,
				"LEGACY="+test.legacy,
				"MODULE_DIRECTORY=nested",
				`MODULE_MATRIX=[{"directory":"."}]`,
				"SOURCE_BOOTSTRAP="+strconv.FormatBool(test.sourceBootstrap),
			)
			combined, err := command.CombinedOutput()
			if test.wantError {
				if err == nil || !strings.Contains(string(combined), "bounded check --local") {
					t.Fatalf("unsupported verifier should fail without a campaign: error = %v, output = %q", err, combined)
				}
			} else if err != nil {
				t.Fatalf("module contract error = %v, output = %q", err, combined)
			}
			contents, err := os.ReadFile(invocations)
			if err != nil {
				t.Fatal(err)
			}
			if got := strings.TrimSpace(string(contents)); got != test.want {
				t.Fatalf("golib invocation = %q, want %q", got, test.want)
			}
			if test.wantError {
				return
			}
			failed := exec.CommandContext(t.Context(), "bash", "-euo", "pipefail", "-c", contract.Run)
			failed.Env = append(append([]string(nil), command.Env...), "GOLIB_ROUTING_FIXTURE_STATUS=7")
			var failure *exec.ExitError
			if !errors.As(failed.Run(), &failure) || failure.ExitCode() != 7 {
				t.Fatal("module contract did not preserve failed gate exit status")
			}
		})
	}
}

type workflowDocument struct {
	On struct {
		WorkflowCall struct {
			Inputs map[string]workflowInput `yaml:"inputs"`
		} `yaml:"workflow_call"`
		WorkflowDispatch struct {
			Inputs map[string]workflowInput `yaml:"inputs"`
		} `yaml:"workflow_dispatch"`
	} `yaml:"on"`
	Jobs map[string]workflowJob `yaml:"jobs"`
}

type workflowInput struct {
	Type    string `yaml:"type"`
	Default any    `yaml:"default"`
}

type workflowJob struct {
	If    string         `yaml:"if"`
	Steps []workflowStep `yaml:"steps"`
}

type workflowStep struct {
	Uses string            `yaml:"uses"`
	ID   string            `yaml:"id"`
	Name string            `yaml:"name"`
	If   string            `yaml:"if"`
	Env  map[string]string `yaml:"env"`
	With map[string]string `yaml:"with"`
	Run  string            `yaml:"run"`
}

func TestReusableWorkflowConfiguresBootstrapProxyForEveryGoBuild(t *testing.T) {
	content := readProjectFile(t, ".github/workflows/library-ci.yml")
	for _, required := range []string{
		"uses: faustbrian/go-library-tools/.github/actions/setup-bootstrap-proxy@",
		"bootstrap_url: ${{ vars.GOLIB_BOOTSTRAP_PROXY_URL }}",
		"bootstrap_sha256: ${{ vars.GOLIB_BOOTSTRAP_PROXY_SHA256 }}",
	} {
		if count := strings.Count(content, required); count != 2 {
			t.Errorf("reusable workflow %q count = %d, want quality and CodeQL", required, count)
		}
	}
}

func TestBootstrapProxyActionVerifiesArchiveBeforeExport(t *testing.T) {
	content := readProjectFile(t, ".github/actions/setup-bootstrap-proxy/setup.sh")
	checksum := strings.Index(content, "sha256sum --check")
	extraction := strings.Index(content, "tar --extract")
	export := strings.Index(content, "GOPROXY=https://proxy.golang.org,file://")
	if checksum < 0 || extraction < 0 || export < 0 || checksum > extraction || extraction > export {
		t.Fatal("bootstrap proxy action must verify before extraction and export")
	}
	for _, required := range []string{"golib archive validate --file", "GONOSUMDB=github.com/faustbrian/go-*"} {
		if !strings.Contains(content, required) {
			t.Errorf("bootstrap proxy action lacks %q", required)
		}
	}
}

func TestBootstrapProxyActionStopsBeforeExtractionWhenArchiveValidationFails(t *testing.T) {
	root := projectRoot(t)
	task := t.TempDir()
	bin := filepath.Join(task, "bin")
	if err := os.Mkdir(bin, 0o700); err != nil {
		t.Fatal(err)
	}
	log := filepath.Join(task, "commands.log")
	writeStub := func(name, body string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(bin, name), []byte("#!/usr/bin/env bash\nset -eu\n"+body), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	writeStub("curl", `echo curl >>"${COMMAND_LOG}"
while [[ $# -gt 0 ]]; do
  if [[ "$1" == "--output" ]]; then shift; : >"$1"; fi
  shift
done
`)
	writeStub("sha256sum", `echo checksum >>"${COMMAND_LOG}"
cat >/dev/null
`)
	writeStub("golib", `if [[ "${1:-}" == "--help" ]]; then
  echo 'golib archive validate --file <path>'
  exit 0
fi
echo golib >>"${COMMAND_LOG}"
exit 42
`)
	writeStub("tar", `echo tar >>"${COMMAND_LOG}"
`)
	githubEnv := filepath.Join(task, "github-env")
	if err := os.WriteFile(githubEnv, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	command := exec.CommandContext(t.Context(), "bash", filepath.Join(root, ".github", "actions", "setup-bootstrap-proxy", "setup.sh"))
	environment := append(os.Environ(),
		"PATH="+bin+":"+os.Getenv("PATH"),
		"COMMAND_LOG="+log,
		"RUNNER_TEMP="+task,
		"GITHUB_ENV="+githubEnv,
		"GITHUB_ACTION_PATH="+filepath.Join(root, ".github", "actions", "setup-bootstrap-proxy"),
		"BOOTSTRAP_URL=https://example.invalid/proxy.tar.gz",
		"BOOTSTRAP_SHA256="+strings.Repeat("a", 64),
	)
	command.Env = environment
	legacyCommand := exec.CommandContext(t.Context(), "bash", filepath.Join(root, ".github", "actions", "setup-bootstrap-proxy", "setup.sh"))
	legacyCommand.Env = environment
	if err := legacyCommand.Run(); err == nil {
		t.Fatal("bootstrap setup succeeded after archive validation failure")
	}
	commands, err := os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	if string(commands) != "curl\nchecksum\ngolib\n" || strings.Contains(string(commands), "tar") {
		t.Fatalf("bootstrap command order = %q", commands)
	}
	exported, err := os.ReadFile(githubEnv)
	if err != nil || len(exported) != 0 {
		t.Fatalf("bootstrap exported environment after failure: %q, %v", exported, err)
	}

	writeStub("golib", `if [[ "${1:-}" == "--help" ]]; then
  echo 'legacy golib help'
  exit 0
fi
exit 99
`)
	writeStub("go", `[[ "${1:-}" == "-C" ]]
[[ "${3:-}" == "run" ]]
[[ "${4:-}" == "./cmd/golib" ]]
echo go >>"${COMMAND_LOG}"
exit 42
`)
	if err := os.WriteFile(log, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := command.Run(); err == nil {
		t.Fatal("legacy fallback succeeded after pinned-source validation failure")
	}
	commands, err = os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	if string(commands) != "curl\nchecksum\ngo\n" || strings.Contains(string(commands), "tar") {
		t.Fatalf("legacy bootstrap command order = %q", commands)
	}
}

func TestReusableWorkflowInstallsGolibBeforeEveryArchiveValidation(t *testing.T) {
	var workflow workflowDocument
	if err := yaml.Unmarshal([]byte(readProjectFile(t, ".github/workflows/library-ci.yml")), &workflow); err != nil {
		t.Fatal(err)
	}
	for _, job := range []string{"quality", "codeql"} {
		t.Run(job, func(t *testing.T) {
			bootstrap := -1
			var setups []int
			for index, step := range workflow.Jobs[job].Steps {
				if strings.HasPrefix(step.Uses, "faustbrian/go-library-tools/.github/actions/setup-golib@") || step.Name == "Validate and install source verifier" {
					setups = append(setups, index)
				}
				if strings.HasPrefix(step.Uses, "faustbrian/go-library-tools/.github/actions/setup-bootstrap-proxy@") {
					bootstrap = index
				}
			}
			if bootstrap < 0 || len(setups) != 2 {
				t.Fatal("job must define both golib setup modes and bootstrap archive validation")
			}
			for _, setup := range setups {
				if setup >= bootstrap {
					t.Fatal("both golib setup modes must precede bootstrap archive validation")
				}
			}
		})
	}
}

func TestToolingWorkflowSeparatesFastPullRequestAndAggregateMilestoneChecks(t *testing.T) {
	content := readProjectFile(t, ".github/workflows/ci.yml")
	qualityStart := strings.Index(content, "  quality:\n")
	milestoneStart := strings.Index(content, "  milestone:\n")
	codeQLStart := strings.Index(content, "  codeql:\n")
	if qualityStart < 0 || milestoneStart < 0 || codeQLStart < 0 || qualityStart >= milestoneStart || milestoneStart >= codeQLStart {
		t.Fatal("tooling workflow does not define ordered quality and milestone jobs")
	}
	quality := content[qualityStart:milestoneStart]
	milestone := content[milestoneStart:codeQLStart]
	for _, required := range []string{"make local-ci", "Run mutation verifier behavior when affected", "-tags=verifierintegration"} {
		if !strings.Contains(quality, required) {
			t.Errorf("pull-request quality lacks %q", required)
		}
	}
	for _, forbidden := range []string{"make check", "golib check --all", "Upload verification evidence"} {
		if strings.Contains(quality, forbidden) {
			t.Errorf("pull-request quality includes aggregate work %q", forbidden)
		}
	}
	for _, required := range []string{
		"always() && needs.quality.result == 'success'",
		"needs.compatibility-consumer-scope.outputs.publication_required == 'true'",
		"needs: [quality, compatibility-consumer-scope]",
		"make milestone-check",
	} {
		if !strings.Contains(milestone, required) {
			t.Errorf("milestone job lacks %q", required)
		}
	}
	if strings.Contains(milestone, "make local-ci") {
		t.Fatal("milestone job repeats the successful local contract")
	}
	makefile := readProjectFile(t, "Makefile")
	if !strings.Contains(makefile, "milestone-check: consumers compatibility") {
		t.Fatal("Makefile does not keep bounded local and aggregate checks separate")
	}
	for _, retired := range []string{"forward-oracle", "tools/provenance", "Upload verification evidence"} {
		if strings.Contains(content, retired) {
			t.Errorf("tooling workflow retains retired routine evidence machinery %q", retired)
		}
	}
}

func TestCompatibilityConsumerMatrixPinsRuntimeToolchain(t *testing.T) {
	content := readProjectFile(t, ".github/workflows/ci.yml")
	runner := readProjectFile(t, ".github/scripts/run-compatibility-consumer.sh")
	version := strings.TrimSpace(readProjectFile(t, ".go-version"))
	for _, required := range []string{
		"go-version-file: .go-version",
		"run: bash .github/scripts/run-compatibility-consumer.sh",
		"EXPECTED_GOOS: ${{ matrix.goos }}",
		"EXPECTED_GOARCH: ${{ matrix.goarch }}",
	} {
		if !strings.Contains(content, required) {
			t.Errorf("compatibility consumer matrix lacks %q", required)
		}
	}
	for _, required := range []string{
		`test "$(go env GOVERSION)" = go` + version,
		`test "$(go env GOOS)/$(go env GOARCH)" = "${EXPECTED_GOOS}/${EXPECTED_GOARCH}"`,
	} {
		if !strings.Contains(runner, required) {
			t.Errorf("compatibility consumer runner lacks %q", required)
		}
	}
}

func TestParityWorkflowDoesNotModifyRepresentativeSource(t *testing.T) {
	content := readProjectFile(t, ".github/workflows/parity-rehearsal.yml")
	for _, forbidden := range []string{
		"Normalize legacy analyzer compatibility",
		"sed -i 's#golib",
		"sed -i 's/^STATICCHECK_VERSION=",
		".golib/scripts/build-golib-gremlins.sh",
	} {
		if strings.Contains(content, forbidden) {
			t.Errorf("parity workflow modifies representative source through %q", forbidden)
		}
	}
}

func TestSharedParityUsesPatchedGoVersionForConsumerGates(t *testing.T) {
	content := readProjectFile(t, ".github/workflows/parity-rehearsal.yml")
	sharedStart := strings.Index(content, "  shared:\n")
	if sharedStart < 0 {
		t.Fatal("parity workflow has no shared job")
	}
	sharedRemainder := content[sharedStart:]
	shared, _, found := strings.Cut(sharedRemainder, "\n  performance:\n")
	if !found {
		t.Fatal("parity workflow has no performance job after shared parity")
	}
	if count := strings.Count(shared, "uses: actions/setup-go@"); count != 1 {
		t.Fatalf("shared parity setup-go steps = %d, want 1", count)
	}
	setup := strings.Index(shared, "go-version-file: tooling/rehearsals/.go-version")
	build := strings.Index(shared, "name: Build source CLI")
	if setup < 0 || build < 0 || setup > build {
		t.Fatal("shared parity must select the patched consumer Go version before building the source CLI")
	}
	if !strings.Contains(shared[build:], "GOTOOLCHAIN=auto") {
		t.Fatal("source CLI build does not opt into its required automatic Go toolchain")
	}
}

func TestParityWorkflowUsesActionsPathChannelForGoWrapper(t *testing.T) {
	content := readProjectFile(t, ".github/workflows/parity-rehearsal.yml")
	if count := strings.Count(content, `"${GITHUB_ENV}" "${GITHUB_PATH}"`); count != 3 {
		t.Fatalf("parity wrapper path exports = %d, want legacy, shared, and performance", count)
	}
}

func TestPerformanceRehearsalPublishesComparableRawMeasurements(t *testing.T) {
	workflow := readProjectFile(t, ".github/workflows/parity-rehearsal.yml")
	for _, required := range []string{
		"name: Performance / ${{ matrix.name }}",
		"rehearsals/performance.sh",
		"performance-${{ matrix.artifact }}",
		"rehearsals/performance-compare.sh",
		"performance-results.json",
		"performance-services.status",
		"performance_source_run_id",
		"github-token: ${{ github.token }}",
		"Resolve measurement source",
		`/repos/${GITHUB_REPOSITORY}/actions/runs/${REQUESTED_RUN_ID}`,
		"run-id: ${{ steps.source.outputs.run_id }}",
		"rehearsals/repositories.json",
		`["core", "service"]`,
	} {
		if !strings.Contains(workflow, required) {
			t.Errorf("performance workflow lacks %q", required)
		}
	}

	harness := readProjectFile(t, "rehearsals/performance.sh")
	for _, required := range []string{
		"startup-diagnostic",
		"repository-inventory",
		"checkpoint-reuse",
		"module-scaling-sequential",
		"module-scaling-concurrent",
		"peak_rss_kib",
		"artifact_size_bytes",
		"isolated_cache_residue",
		"service-lifecycle",
		"mutation_package_count",
		"reused content-identical mutation evidence",
		"tooling_revision",
		"golib_sha256",
	} {
		if !strings.Contains(harness, required) {
			t.Errorf("performance harness lacks %q", required)
		}
	}

	documentation := readProjectFile(t, "docs/performance.md")
	for _, required := range []string{
		"content-identical",
		"Raw Results",
		"Runner variance",
		"No-op and checkpoint reuse",
		"Concurrent module scaling",
	} {
		if !strings.Contains(documentation, required) {
			t.Errorf("performance documentation lacks %q", required)
		}
	}
}

func TestParityWorkflowCanReuseCompletedContractArtifacts(t *testing.T) {
	workflow := readProjectFile(t, ".github/workflows/parity-rehearsal.yml")
	for _, required := range []string{
		"compatibility_source_run_id",
		"Resolve compatibility source",
		`/repos/${GITHUB_REPOSITORY}/actions/runs/${REQUESTED_RUN_ID}`,
		`/repos/${GITHUB_REPOSITORY}/actions/runs/${REQUESTED_RUN_ID}/jobs?per_page=100`,
		`git fetch --no-tags --depth=1 origin "${source_sha}"`,
		`git diff --name-only "${source_sha}" "${GITHUB_SHA}" --`,
		`grep -Ev '(^rehearsals/compare\.sh$|_test\.go$)'`,
		"run-id: ${{ steps.contract-source.outputs.run_id }}",
	} {
		if !strings.Contains(workflow, required) {
			t.Errorf("compatibility workflow lacks %q", required)
		}
	}
	if count := strings.Count(workflow, "if: inputs.compatibility_source_run_id == ''"); count != 2 {
		t.Fatalf("compatibility artifact-reuse guards = %d, want legacy and shared", count)
	}
}

func TestParityRehearsalReusesRepresentativeMutationEvidence(t *testing.T) {
	t.Parallel()

	workflow := readProjectFile(t, ".github/workflows/parity-rehearsal.yml")
	sharedStart := strings.Index(workflow, "  shared:\n")
	performanceStart := strings.Index(workflow, "  performance:\n")
	performanceReportStart := strings.Index(workflow, "  performance-report:\n")
	if sharedStart < 0 || performanceStart < 0 || performanceReportStart < 0 ||
		sharedStart >= performanceStart || performanceStart >= performanceReportStart {
		t.Fatal("parity workflow job order is invalid")
	}
	shared := workflow[sharedStart:performanceStart]
	performance := workflow[performanceStart:performanceReportStart]
	for _, required := range []string{
		"Restore mutation checkpoints",
		"restore-ci-mutation-evidence.sh",
		"if: inputs.performance_source_run_id == ''",
	} {
		if !strings.Contains(workflow, required) {
			t.Errorf("parity workflow lacks %q", required)
		}
	}
	for name, job := range map[string]string{"shared": shared, "performance": performance} {
		for _, required := range []string{
			"cp .golib/mutation-bootstrap/*.zip",
			"cp .golib/mutation-history-migrations.json",
		} {
			if !strings.Contains(job, required) {
				t.Errorf("%s parity job lacks %q", name, required)
			}
		}
	}

	wantImports := map[string]int{
		"go-authorization":        1,
		"go-cloudevents":          2,
		"go-knapsack":             2,
		"go-openapi":              1,
		"go-transactional-outbox": 5,
	}
	for repository, want := range wantImports {
		configuration, err := config.Load(filepath.Join(projectRoot(t), "rehearsals", repository))
		if err != nil {
			t.Errorf("load %s rehearsal configuration: %v", repository, err)
			continue
		}
		if got := len(configuration.Mutation.Imports); got != want {
			t.Errorf("%s mutation imports = %d, want %d", repository, got, want)
		}
	}
}

func TestPerformanceRehearsalExcludesInjectedPolicyFromSourceChecks(t *testing.T) {
	workflow := readProjectFile(t, ".github/workflows/parity-rehearsal.yml")
	performanceStart := strings.Index(workflow, "  performance:\n")
	if performanceStart < 0 {
		t.Fatal("parity workflow has no performance job")
	}
	performance, _, found := strings.Cut(workflow[performanceStart:], "\n  performance-report:\n")
	if !found {
		t.Fatal("parity workflow has no performance report after performance jobs")
	}
	for _, required := range []string{
		`printf '%s\n' '/.golib.yaml' '/.verification/'`,
		`>>.git/info/exclude`,
	} {
		if !strings.Contains(performance, required) {
			t.Errorf("performance policy setup lacks %q", required)
		}
	}
}

func TestRehearsalFuzzTargetsAreExact(t *testing.T) {
	fixtures, err := filepath.Glob(filepath.Join(projectRoot(t), "rehearsals", "go-*", ".golib.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	for _, fixture := range fixtures {
		content, readErr := os.ReadFile(fixture)
		if readErr != nil {
			t.Fatal(readErr)
		}
		for line := range strings.SplitSeq(string(content), "\n") {
			selector := strings.TrimSpace(line)
			if !strings.HasPrefix(selector, "fuzz:") {
				continue
			}
			value := strings.TrimSpace(strings.TrimPrefix(selector, "fuzz:"))
			if !strings.HasPrefix(value, "'^Fuzz") || !strings.HasSuffix(value, "$'") {
				t.Errorf("%s has non-exact fuzz selector %q", filepath.Base(filepath.Dir(fixture)), value)
			}
		}
	}
}

func TestSetupActionVerifiesReleasedArtifactBeforeExtraction(t *testing.T) {
	content := readProjectFile(t, ".github/actions/setup-golib/action.yml")
	checksumSet := strings.Contains(content, "tool_checksums_sha256:")
	checksumSetVerification := strings.Index(content, "checksums.txt\" | sha256sum --check")
	checksum := strings.LastIndex(content, "sha256sum --check")
	attestation := strings.Index(content, "gh attestation verify")
	extraction := strings.Index(content, "tar --extract")
	if !checksumSet || checksumSetVerification < 0 || checksumSetVerification > checksum ||
		checksum < 0 || attestation < 0 || extraction < 0 || checksum > extraction || attestation > extraction {
		t.Fatal("setup action must verify checksum and provenance before extraction")
	}
	for _, platform := range []string{"linux_amd64", "linux_arm64", "darwin_amd64", "darwin_arm64"} {
		if !strings.Contains(content, platform) {
			t.Errorf("setup action lacks %s", platform)
		}
	}
}

func TestReleaseWorkflowBuildsAndAttestsEverySupportedPlatform(t *testing.T) {
	content := readProjectFile(t, ".github/workflows/release.yml")
	for _, required := range []string{
		"{goos: linux, goarch: amd64}",
		"{goos: linux, goarch: arm64}",
		"{goos: darwin, goarch: amd64}",
		"{goos: darwin, goarch: arm64}",
		"checksums.txt",
		"release-manifest.json",
		`"${task}/syft" scan dir:dist/stage`,
		`--output "spdx-json=dist/golib_${GOOS}_${GOARCH}.spdx.json"`,
		"actions/attest@",
		"subject-path: dist/*",
		"name: release-${{ matrix.goos }}-${{ matrix.goarch }}",
		"path: dist/*",
		"actions: read",
		"actions/workflows/ci.yml/runs",
		"gh release create",
	} {
		if !strings.Contains(content, required) {
			t.Errorf("release workflow lacks %q", required)
		}
	}
	if !strings.Contains(content, `declared="$(jq -er`) ||
		!strings.Contains(content, `[[ "${declared}" != "${GITHUB_REF_NAME}" ]]`) {
		t.Fatal("release workflow must bind the tag to the manifest version without bootstrapping an unpublished binary")
	}
	if strings.Contains(content, "go run ./cmd/golib check --all") {
		t.Fatal("release workflow repeats the exact-commit repository contract")
	}
}

func TestReleaseWorkflowRefusesUnverifiedSBOMExecutableAndCleansAcquisition(t *testing.T) {
	var workflow workflowDocument
	if err := yaml.Unmarshal([]byte(readProjectFile(t, ".github/workflows/release.yml")), &workflow); err != nil {
		t.Fatal(err)
	}
	var script string
	for _, step := range workflow.Jobs["build"].Steps {
		if step.Name == "Generate SBOM with digest-pinned Syft" {
			script = step.Run
		}
	}
	checksum, extract, execute := strings.Index(script, "sha256sum --check --strict -"), strings.Index(script, "tar --extract"), strings.Index(script, `"${task}/syft" version`)
	if checksum < 0 || extract < 0 || execute < 0 || checksum > extract || extract > execute {
		t.Fatal("release SBOM executable must be verified before extraction and execution")
	}
	for _, required := range []string{
		"https://github.com/anchore/syft/releases/download/v1.54.0/syft_1.54.0_linux_amd64.tar.gz",
		"54a87372498168b2d033e876fd41fa4e8035b872699e525a57046e1f2f09c860",
		"--proto '=https' --tlsv1.2", "set -euo pipefail",
	} {
		if !strings.Contains(script, required) {
			t.Fatalf("release SBOM acquisition lacks %q", required)
		}
	}
	if strings.Contains(script, "install.sh") {
		t.Fatal("release SBOM acquisition must not run an upstream installer")
	}
	if _, err := exec.LookPath("sha256sum"); err != nil {
		t.Fatal(err)
	}
	for _, failure := range []string{"download", "digest"} {
		t.Run(failure, func(t *testing.T) {
			task := t.TempDir()
			bin := filepath.Join(task, "bin")
			runner := filepath.Join(task, "runner")
			for _, directory := range []string{bin, runner} {
				if err := os.Mkdir(directory, 0o700); err != nil {
					t.Fatal(err)
				}
			}
			log := filepath.Join(task, "commands.log")
			stubs := map[string]string{
				"curl": `echo download >>"${COMMAND_LOG}"
if [[ "${ACQUISITION_FAILURE}" == download ]]; then exit 42; fi
while [[ $# -gt 0 ]]; do
  if [[ "$1" == --output ]]; then shift; printf 'corrupt executable archive' >"$1"; fi
  shift
done
`,
				"tar": `echo extraction >>"${COMMAND_LOG}"
exit 43
`,
			}
			for name, body := range stubs {
				if err := os.WriteFile(filepath.Join(bin, name), []byte("#!/usr/bin/env bash\nset -eu\n"+body), 0o700); err != nil {
					t.Fatal(err)
				}
			}
			command := exec.CommandContext(t.Context(), "bash", "-c", script)
			command.Env = append(os.Environ(), "PATH="+bin+":"+os.Getenv("PATH"), "COMMAND_LOG="+log,
				"RUNNER_TEMP="+runner, "ACQUISITION_FAILURE="+failure, "GOOS=linux", "GOARCH=amd64")
			output, runErr := command.CombinedOutput()
			if runErr == nil {
				t.Fatal("release SBOM generation accepted failed executable acquisition")
			}
			if failure == "digest" && !strings.Contains(string(output), "FAILED") {
				t.Fatalf("corrupt acquisition failed outside checksum verification: %v\n%s", runErr, output)
			}
			commands, err := os.ReadFile(log)
			if err != nil || string(commands) != "download\n" {
				t.Fatalf("unverified executable reached extraction: %q, %v", commands, err)
			}
			children, err := os.ReadDir(runner)
			if err != nil || len(children) != 0 {
				t.Fatalf("failed SBOM acquisition left staging artifacts: %v", err)
			}
		})
	}
}

func TestReleaseWorkflowVerifiesRealTagAgainstCurrentMainAndSecurityEvidence(t *testing.T) {
	var workflow workflowDocument
	if err := yaml.Unmarshal([]byte(readProjectFile(t, ".github/workflows/release.yml")), &workflow); err != nil {
		t.Fatal(err)
	}
	var verifyScript string
	for _, step := range workflow.Jobs["verify"].Steps {
		if step.Name == "Validate release identity and contract" {
			verifyScript = step.Run
		}
	}
	if verifyScript == "" {
		t.Fatal("release workflow lacks executable verification boundary")
	}
	for _, test := range []struct {
		name        string
		annotated   bool
		ready       bool
		advanceMain bool
		outsideMain bool
		failedCI    bool
		wrongCIHead bool
		wantSuccess bool
		wantGoCalls int
	}{
		{name: "exact annotated evidence tag", annotated: true, ready: true, wantSuccess: true, wantGoCalls: 2},
		{name: "exact lightweight evidence tag", ready: true, wantSuccess: true, wantGoCalls: 2},
		{name: "tagged ancestor of main", annotated: true, ready: true, advanceMain: true, wantSuccess: true, wantGoCalls: 2},
		{name: "tag outside main", annotated: true, ready: true, outsideMain: true},
		{name: "failed exact source CI", annotated: true, ready: true, failedCI: true},
		{name: "green wrong source CI", annotated: true, ready: true, wrongCIHead: true},
		{name: "current security interlock", annotated: true, wantGoCalls: 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			work := filepath.Join(t.TempDir(), "work")
			origin := filepath.Join(t.TempDir(), "origin.git")
			runGit(t, "", "init", "--bare", origin)
			runGit(t, "", "init", "--initial-branch=main", work)
			runGit(t, work, "config", "user.name", "Release Test")
			runGit(t, work, "config", "user.email", "release@example.invalid")
			runGit(t, work, "remote", "add", "origin", origin)

			for _, directory := range []string{"docs/ecosystem/security", ".github/workflows"} {
				if err := os.MkdirAll(filepath.Join(work, directory), 0o700); err != nil {
					t.Fatal(err)
				}
			}
			if err := os.WriteFile(filepath.Join(work, "modules.json"), []byte(`{"modules":[{"directory":".","module_path":"github.com/faustbrian/go-library-tools/v2","version":"2.0.0","tag_prefix":"v"}]}`), 0o600); err != nil {
				t.Fatal(err)
			}
			for name, data := range map[string]string{
				"risk-register.json":   `{"schema_version":1,"risks":[]}`,
				"security-matrix.json": `{"schema_version":1,"modules":[]}`,
			} {
				if err := os.WriteFile(filepath.Join(work, "docs", "ecosystem", "security", name), []byte(data), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			if err := os.WriteFile(filepath.Join(work, ".github", "workflows", "release.yml"), []byte("name: fixture\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			runGit(t, work, "add", "modules.json", "docs", ".github")
			runGit(t, work, "commit", "-m", "source")
			sourceSHA := strings.TrimSpace(runGit(t, work, "rev-parse", "HEAD"))
			if test.ready {
				risks, matrix := releaseSecurityDocuments(t, sourceSHA)
				if err := os.WriteFile(filepath.Join(work, "docs", "ecosystem", "security", "risk-register.json"), risks, 0o600); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(work, "docs", "ecosystem", "security", "security-matrix.json"), matrix, 0o600); err != nil {
					t.Fatal(err)
				}
			} else {
				for _, name := range []string{"risk-register.json", "security-matrix.json"} {
					data, err := os.ReadFile(filepath.Join(projectRoot(t), "docs", "ecosystem", "security", name))
					if err != nil {
						t.Fatal(err)
					}
					if err := os.WriteFile(filepath.Join(work, "docs", "ecosystem", "security", name), data, 0o600); err != nil {
						t.Fatal(err)
					}
				}
			}
			runGit(t, work, "add", "docs/ecosystem/security/risk-register.json", "docs/ecosystem/security/security-matrix.json")
			runGit(t, work, "commit", "-m", "security evidence")
			sha := strings.TrimSpace(runGit(t, work, "rev-parse", "HEAD"))
			if test.annotated {
				runGit(t, work, "tag", "-a", "v2.0.0", "-m", "release")
			} else {
				runGit(t, work, "tag", "v2.0.0")
			}
			main := "main"
			if test.outsideMain {
				main = sourceSHA + ":refs/heads/main"
			}
			runGit(t, work, "push", "origin", main, "refs/tags/v2.0.0")
			if test.advanceMain {
				if err := os.WriteFile(filepath.Join(work, "later"), []byte("later\n"), 0o600); err != nil {
					t.Fatal(err)
				}
				runGit(t, work, "add", "later")
				runGit(t, work, "commit", "-m", "advance main")
				runGit(t, work, "push", "origin", "main")
				remoteMain := strings.Fields(runGit(t, "", "ls-remote", origin, "refs/heads/main"))[0]
				if remoteMain == sha {
					t.Fatal("fixture did not advance remote main")
				}
			}
			runGit(t, work, "checkout", "--detach", "v2.0.0")

			bin := filepath.Join(t.TempDir(), "bin")
			if err := os.Mkdir(bin, 0o700); err != nil {
				t.Fatal(err)
			}
			calls := filepath.Join(t.TempDir(), "go-calls")
			ciSHA, conclusion := sha, "success"
			if test.wrongCIHead {
				ciSHA = sourceSHA
			}
			if test.failedCI {
				conclusion = "failure"
			}
			releaseWriteExecutable(t, filepath.Join(bin, "gh"), fmt.Sprintf("printf '%%s' '%s'\n", fmt.Sprintf(`{"workflow_runs":[{"head_sha":%q,"event":"push","status":"completed","conclusion":%q}]}`, ciSHA, conclusion)))
			releaseWriteExecutable(t, filepath.Join(bin, "go"), "printf '%s\\n' \"$*\" >>\"${GO_CALLS}\"\n")
			command := exec.CommandContext(t.Context(), "bash", "-euo", "pipefail", "-c", verifyScript)
			command.Dir = work
			command.Env = append(os.Environ(), "PATH="+bin+string(os.PathListSeparator)+os.Getenv("PATH"), "GITHUB_REF_NAME=v2.0.0", "GITHUB_SHA="+sha, "GITHUB_REPOSITORY=faustbrian/go-library-tools", "GITHUB_TOKEN=test", "GITHUB_OUTPUT="+filepath.Join(t.TempDir(), "github-output"), "RUNNER_TEMP="+t.TempDir(), "GO_CALLS="+calls)
			output, err := command.CombinedOutput()
			if (err == nil) != test.wantSuccess {
				t.Fatalf("verify success = %t, want %t: %v, %s", err == nil, test.wantSuccess, err, output)
			}
			data, readErr := os.ReadFile(calls)
			if readErr != nil && !os.IsNotExist(readErr) {
				t.Fatal(readErr)
			}
			if got := strings.Count(strings.TrimSpace(string(data)), "\n") + boolInt(len(strings.TrimSpace(string(data))) > 0); got != test.wantGoCalls {
				t.Fatalf("go calls = %d, want %d: %q; verify output: %s", got, test.wantGoCalls, data, output)
			}
		})
	}
}

// This exercises the hosted verification shell, not a transcription of its
// policy. Git identity/ancestry are real; only GitHub and Go are external seams.
func TestReleaseWorkflowUsesExactNativeEvidenceWithoutRecordsOnlyCommit(t *testing.T) {
	const cacheQualification = "test -p=1 -tags=verifierintegration ./internal/gates -run ^TestTaskVerifier(SessionReusesPinnedToolCompilation|SessionsRemainDistinctAndCanceledCallsCleanScratch)$ -count=1 -timeout=4m -json"
	var workflow workflowDocument
	if err := yaml.Unmarshal([]byte(readProjectFile(t, ".github/workflows/release.yml")), &workflow); err != nil {
		t.Fatal(err)
	}
	var script string
	for _, step := range workflow.Jobs["verify"].Steps {
		if step.Name == "Validate release identity and contract" {
			script = step.Run
		}
	}
	if script == "" {
		t.Fatal("release verification boundary missing")
	}
	for _, test := range []struct {
		name, defect string
		wantSuccess  bool
	}{
		{name: "exact native source with historical records", wantSuccess: true},
		{name: "squash merge retains same-tree source proof", defect: "distinct-pr-head", wantSuccess: true},
		{name: "live merged PR has empty historical run associations", defect: "empty-pr-association", wantSuccess: true},
		{name: "merged PR source differs from tagged tree", defect: "different-pr-tree"},
		{name: "tag remains ancestor of current remote main", defect: "advance-main", wantSuccess: true},
		{name: "tag outside remote main", defect: "outside-main"},
		{name: "manifest version differs from tag", defect: "wrong-version"},
		{name: "workflow source differs from tag", defect: "wrong-source"},
		{name: "green CI belongs to another head", defect: "wrong-ci-head"},
		{name: "CI belongs to another branch", defect: "wrong-ci-branch"},
		{name: "CI run failed", defect: "failed-ci"},
		{name: "new failed CI cannot reuse an older green run", defect: "newer-failed-ci"},
		{name: "Required missing", defect: "missing-required"},
		{name: "Required failed", defect: "failed-required"},
		{name: "Quality missing", defect: "missing-quality"},
		{name: "Quality execution skipped", defect: "skipped-quality"},
		{name: "Quality execution failed", defect: "failed-quality"},
		{name: "CodeQL build missing", defect: "missing-build"},
		{name: "CodeQL build skipped in green job", defect: "skipped-build"},
		{name: "CodeQL build failed", defect: "failed-build"},
		{name: "CodeQL analyze missing", defect: "missing-analyze"},
		{name: "CodeQL analyze skipped in green job", defect: "skipped-analyze"},
		{name: "CodeQL analyze failed", defect: "failed-analyze"},
		{name: "CodeQL analyze uses another action", defect: "wrong-analyze"},
		{name: "merged PR absent", defect: "missing-pr"},
		{name: "PR not merged", defect: "unmerged-pr"},
		{name: "PR merge belongs to another tag head", defect: "wrong-pr-merge"},
		{name: "PR targets another branch", defect: "wrong-pr-base"},
		{name: "authenticated PR base SHA malformed", defect: "invalid-pr-base-sha"},
		{name: "authenticated PR base SHA missing", defect: "missing-pr-base-sha"},
		{name: "full release secret scan fails", defect: "secrets-fail"},
		{name: "existing security validator refuses disposition", defect: "security-fail"},
		{name: "cache qualification command fails", defect: "cache-command-fail"},
		{name: "cache compilation pass missing", defect: "cache-missing-first"},
		{name: "cache isolation pass missing", defect: "cache-missing-second"},
		{name: "cache qualification result failed", defect: "cache-failed-result"},
		{name: "cache qualification result skipped", defect: "cache-skipped-result"},
		{name: "cache qualification test identity wrong", defect: "cache-wrong-test"},
		{name: "cache compilation pass duplicated", defect: "cache-duplicate-first"},
		{name: "cache isolation pass duplicated", defect: "cache-duplicate-second"},
		{name: "cache qualification JSON malformed", defect: "cache-malformed-result"},
	} {
		t.Run(test.name, func(t *testing.T) {
			work := filepath.Join(t.TempDir(), "work")
			origin := filepath.Join(t.TempDir(), "origin.git")
			runGit(t, "", "init", "--bare", origin)
			runGit(t, "", "init", "--initial-branch=main", work)
			runGit(t, work, "config", "user.name", "Release Test")
			runGit(t, work, "config", "user.email", "release@example.invalid")
			runGit(t, work, "remote", "add", "origin", origin)
			write := func(path, data string) {
				t.Helper()
				path = filepath.Join(work, path)
				if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			write("modules.json", `{"modules":[{"directory":".","module_path":"github.com/faustbrian/go-library-tools/v2","version":"2.0.0","tag_prefix":"v"}]}`)
			write("runtime.go", "package fixture\nconst Version = 1\n")
			runGit(t, work, "add", "modules.json", "runtime.go")
			runGit(t, work, "commit", "-m", "historical source")
			historical := strings.TrimSpace(runGit(t, work, "rev-parse", "HEAD"))
			risks, matrix := releaseSecurityDocuments(t, historical)
			write("docs/ecosystem/security/risk-register.json", string(risks))
			write("docs/ecosystem/security/security-matrix.json", string(matrix))
			runGit(t, work, "add", "docs")
			runGit(t, work, "commit", "-m", "historical security disposition")
			prHead := strings.TrimSpace(runGit(t, work, "rev-parse", "HEAD"))
			// No fresh scanner rows and no records-only release commit: the tag
			// identifies executable/version preparation, with native runtime CI.
			write("runtime.go", "package fixture\nconst Version = 2\n")
			write("modules.json", `{"modules":[{"directory":".","module_path":"github.com/faustbrian/go-library-tools/v2","version":"2.1.0","tag_prefix":"v"}]}`)
			runGit(t, work, "add", "runtime.go", "modules.json")
			runGit(t, work, "commit", "-m", "prepare current source release")
			sha := strings.TrimSpace(runGit(t, work, "rev-parse", "HEAD"))
			runGit(t, work, "tag", "-a", "v2.1.0", "-m", "release")
			main := "main"
			if test.defect == "outside-main" {
				main = historical + ":refs/heads/main"
			}
			runGit(t, work, "push", "origin", main, "refs/tags/v2.1.0")
			if test.defect == "advance-main" {
				write("later", "later\n")
				runGit(t, work, "add", "later")
				runGit(t, work, "commit", "-m", "advance main")
				runGit(t, work, "push", "origin", "main")
			}
			runGit(t, work, "checkout", "--detach", "v2.1.0")
			if test.defect == "distinct-pr-head" {
				tree := strings.TrimSpace(runGit(t, work, "rev-parse", "HEAD^{tree}"))
				prHead = strings.TrimSpace(runGit(t, work, "commit-tree", tree, "-p", historical, "-m", "reviewed same-tree PR source"))
				runGit(t, work, "push", "origin", prHead+":refs/heads/feature/release")
			} else if test.defect != "different-pr-tree" {
				prHead = sha
			}
			fixtures := releaseNativeEvidence(t, sha, prHead, historical, test.defect)
			bin := filepath.Join(t.TempDir(), "bin")
			if err := os.Mkdir(bin, 0o700); err != nil {
				t.Fatal(err)
			}
			releaseWriteExecutable(t, filepath.Join(bin, "gh"), `
[[ "${GH_TOKEN:-}" == test ]] || { echo 'unauthenticated fixture request' >&2; exit 90; }
endpoint=''
for argument in "$@"; do
  case "${argument}" in /repos/*) endpoint="${argument}" ;; esac
done
case "${endpoint}" in
  */actions/workflows/ci.yml/runs*)
    if [[ "${endpoint}" == *event=pull_request* ]]; then file=pr-runs; else file=push-runs; fi ;;
  */actions/runs/100/jobs*) file=push-jobs ;;
  */actions/runs/200/jobs*) file=pr-jobs ;;
  */commits/*/pulls*) file=pulls ;;
  */pulls/7) file=pull ;;
  *) echo "unexpected gh request: ${endpoint}" >&2; exit 91 ;;
esac
cat "${GH_FIXTURES}/${file}.json"
`)
			releaseWriteExecutable(t, filepath.Join(bin, "go"), `
case "${GOCACHE:-}|${GOMODCACHE:-}|${GOTMPDIR:-}" in
  "${RUNNER_TEMP}"/golib-release-check.*/cache\|"${RUNNER_TEMP}"/golib-release-check.*/mod\|"${RUNNER_TEMP}"/golib-release-check.*/tmp) ;;
  *) echo 'unowned Go fixture cache' >&2; exit 93 ;;
esac
printf '%s\n' "$*" >>"${GO_CALLS}"
printf '%s|%s|%s\n' "$GOCACHE" "$GOMODCACHE" "$GOTMPDIR" >>"${GO_CACHE_CALLS}"
case "$*" in
  'run ./cmd/golib security validate --directory docs/ecosystem/security')
    [[ "${DEFECT}" != security-fail ]] || exit 42 ;;
  'run ./cmd/golib secrets check')
    [[ "${DEFECT}" != secrets-fail ]] || exit 42 ;;
  'run ./cmd/golib release check') ;;
  '`+cacheQualification+`')
    [[ "${GOMAXPROCS:-}" == 2 && "${GOFLAGS:-}" == -p=1 ]] || { echo 'unbounded cache fixture command' >&2; exit 94; }
    [[ "${DEFECT}" != cache-command-fail ]] || exit 42
    if [[ "${DEFECT}" == cache-malformed-result ]]; then printf '{malformed\n'; exit 0; fi
    emit_result() {
      printf '{"Action":"%s","Package":"github.com/faustbrian/go-library-tools/v2/internal/gates","Test":"%s"}\n' "$1" "$2"
    }
    first=TestTaskVerifierSessionReusesPinnedToolCompilation
    second=TestTaskVerifierSessionsRemainDistinctAndCanceledCallsCleanScratch
    [[ "${DEFECT}" == cache-missing-first ]] || emit_result pass "$first"
    case "${DEFECT}" in
      cache-missing-second) ;;
      cache-failed-result) emit_result fail "$second" ;;
      cache-skipped-result) emit_result skip "$second" ;;
      cache-wrong-test) emit_result pass TestAnotherCacheContract ;;
      *) emit_result pass "$second" ;;
    esac
    if [[ "${DEFECT}" == cache-duplicate-first ]]; then emit_result pass "$first"; fi
    if [[ "${DEFECT}" == cache-duplicate-second ]]; then emit_result pass "$second"; fi
    ;;
  *) echo 'unexpected Go fixture command' >&2; exit 92 ;;
esac
`)
			calls := filepath.Join(t.TempDir(), "go-calls")
			cacheCalls := filepath.Join(t.TempDir(), "cache-calls")
			githubSHA, tag := sha, "v2.1.0"
			if test.defect == "wrong-source" {
				githubSHA = historical
			}
			if test.defect == "wrong-version" {
				tag = "v2.0.0"
			}
			runner := t.TempDir()
			githubOutput := filepath.Join(t.TempDir(), "github-output")
			command := exec.CommandContext(t.Context(), "bash", "-euo", "pipefail", "-c", script)
			command.Dir = work
			command.Env = append(os.Environ(), "PATH="+bin+string(os.PathListSeparator)+os.Getenv("PATH"),
				"GITHUB_REF_NAME="+tag, "GITHUB_SHA="+githubSHA, "GITHUB_REPOSITORY=faustbrian/go-library-tools",
				"GITHUB_TOKEN=test", "GITHUB_OUTPUT="+githubOutput, "RUNNER_TEMP="+runner, "GO_CALLS="+calls, "GO_CACHE_CALLS="+cacheCalls, "GH_FIXTURES="+fixtures, "DEFECT="+test.defect)
			output, err := command.CombinedOutput()
			if strings.Contains(string(output), "unexpected gh request") || strings.Contains(string(output), "unexpected Go fixture command") || strings.Contains(string(output), "unauthenticated fixture request") || strings.Contains(string(output), "unowned Go fixture cache") || strings.Contains(string(output), "unbounded cache fixture command") {
				t.Fatalf("external fixture contract mismatch: %s", output)
			}
			if (err == nil) != test.wantSuccess {
				t.Fatalf("native verify success = %t, want %t: %v, %s", err == nil, test.wantSuccess, err, output)
			}
			if test.wantSuccess {
				refs, err := os.ReadFile(githubOutput)
				if err != nil || string(refs) != "native=true\nbase_ref="+historical+"\nhead_ref="+sha+"\n" {
					t.Fatalf("dependency review refs must come from authenticated PR base and exact tagged head: %q, %v", refs, err)
				}
				data, err := os.ReadFile(calls)
				if err != nil {
					t.Fatal(err)
				}
				for _, required := range []string{"run ./cmd/golib security validate --directory docs/ecosystem/security", "run ./cmd/golib secrets check", "run ./cmd/golib release check"} {
					if strings.Count(string(data), required+"\n") != 1 {
						t.Fatalf("native release must validate disposition, scan full secrets once and retain release checks: %q", data)
					}
				}
				if strings.Count(string(data), cacheQualification+"\n") != 1 || len(strings.Split(strings.TrimSpace(string(data)), "\n")) != 4 {
					t.Fatalf("native release did not execute exactly one bounded cache qualification: %q", data)
				}
			}
			if strings.HasPrefix(test.defect, "cache-") {
				data, err := os.ReadFile(calls)
				if err != nil || strings.Count(string(data), cacheQualification+"\n") != 1 || strings.Contains(string(data), "run ./cmd/golib release check") {
					t.Fatalf("cache refusal did not stop publication qualification: %q / %v", data, err)
				}
				if refs, err := os.ReadFile(githubOutput); err == nil || !os.IsNotExist(err) || len(refs) != 0 {
					t.Fatalf("cache refusal published dependency review refs: %q / %v", refs, err)
				}
			}
			if data, err := os.ReadFile(cacheCalls); err == nil {
				identities := strings.Split(strings.TrimSpace(string(data)), "\n")
				for _, identity := range identities {
					if identity != identities[0] {
						t.Fatal("release Go commands did not reuse one owned cache session")
					}
				}
			} else if !os.IsNotExist(err) {
				t.Fatal(err)
			}
			for path, before := range map[string][]byte{"risk-register.json": risks, "security-matrix.json": matrix} {
				after, err := os.ReadFile(filepath.Join(work, "docs/ecosystem/security", path))
				if err != nil || string(after) != string(before) {
					t.Fatalf("native verification rewrote historical evidence %s: %v", path, err)
				}
			}
			remaining, err := os.ReadDir(runner)
			if err != nil || len(remaining) != 0 {
				t.Fatalf("release private workspace remains: %v, %v", remaining, err)
			}
		})
	}
}

func TestReleaseWorkflowRequiresAuthenticatedDependencyReviewRefs(t *testing.T) {
	var workflow workflowDocument
	if err := yaml.Unmarshal([]byte(readProjectFile(t, ".github/workflows/release.yml")), &workflow); err != nil {
		t.Fatal(err)
	}
	var action, guard workflowStep
	for _, step := range workflow.Jobs["verify"].Steps {
		if strings.HasPrefix(step.Uses, "actions/dependency-review-action@") {
			if action.Uses != "" {
				t.Fatal("multiple release dependency reviews")
			}
			action = step
		}
		if step.Name == "Validate dependency review inputs" {
			guard = step
		}
	}
	if action.Uses != "actions/dependency-review-action@a1d282b36b6f3519aa1f3fc636f609c47dddb294" || guard.Run == "" {
		t.Fatal("native release lacks pinned dependency review with an executable fail-closed ref guard")
	}
	for key, want := range map[string]string{
		"base-ref": "${{ steps.identity.outputs.base_ref }}", "head-ref": "${{ steps.identity.outputs.head_ref }}",
		"fail-on-severity": "low", "license-check": "true", "vulnerability-check": "true",
		"warn-only": "false", "comment-summary-in-pr": "never",
	} {
		if action.With[key] != want {
			t.Fatalf("dependency review %s = %q, want %q", key, action.With[key], want)
		}
	}
	if action.If != "steps.identity.outputs.native == 'true'" || guard.If != "steps.identity.outputs.native != 'false'" {
		t.Fatal("missing native identity outputs must not silently skip dependency review")
	}
	for key, want := range map[string]string{
		"NATIVE": "${{ steps.identity.outputs.native }}", "BASE_REF": "${{ steps.identity.outputs.base_ref }}", "HEAD_REF": "${{ steps.identity.outputs.head_ref }}",
	} {
		if guard.Env[key] != want {
			t.Fatalf("ref guard %s does not consume authenticated identity outputs", key)
		}
	}
	for _, test := range []struct {
		name, native, base, head string
		wantSuccess              bool
	}{
		{name: "authenticated immutable refs", native: "true", base: strings.Repeat("a", 40), head: strings.Repeat("b", 40), wantSuccess: true},
		{name: "missing identity outputs"},
		{name: "missing base", native: "true", head: strings.Repeat("b", 40)},
		{name: "missing head", native: "true", base: strings.Repeat("a", 40)},
		{name: "symbolic base", native: "true", base: "main", head: strings.Repeat("b", 40)},
		{name: "symbolic head", native: "true", base: strings.Repeat("a", 40), head: "v2.1.0"},
		{name: "unknown mode", native: "unknown", base: strings.Repeat("a", 40), head: strings.Repeat("b", 40)},
	} {
		t.Run(test.name, func(t *testing.T) {
			command := exec.CommandContext(t.Context(), "bash", "-euo", "pipefail", "-c", guard.Run)
			command.Env = append(os.Environ(), "NATIVE="+test.native, "BASE_REF="+test.base, "HEAD_REF="+test.head)
			output, err := command.CombinedOutput()
			if (err == nil) != test.wantSuccess {
				t.Fatalf("dependency review ref admission = %t, want %t: %v, %s", err == nil, test.wantSuccess, err, output)
			}
		})
	}
}

func releaseNativeEvidence(t *testing.T, sha, prHead, other, defect string) string {
	t.Helper()
	step := func(name string) map[string]any {
		return map[string]any{"name": name, "number": 1, "status": "completed", "conclusion": "success", "started_at": "2099-09-01T00:00:00Z", "completed_at": "2099-09-01T00:01:00Z"}
	}
	job := func(name string, steps ...map[string]any) map[string]any {
		return map[string]any{"id": 10, "run_id": 100, "head_sha": sha, "name": name, "status": "completed", "conclusion": "success", "steps": steps}
	}
	required := job("Required", step("Run set -euo pipefail"))
	qualityStep := step("Run fast local CI")
	quality := job("Quality", qualityStep)
	build := step("Build runtime source")
	analyze := step("Run github/codeql-action/analyze@2892aa5e19bbd11bc0cff5427e3b750a04d9e3c2")
	codeql := job("CodeQL", build, analyze)
	dependencyStep := step("Run actions/dependency-review-action@a1d282b36b6f3519aa1f3fc636f609c47dddb294")
	dependency := job("Dependency review", dependencyStep)
	dependency["run_id"], dependency["head_sha"] = 200, prHead
	run := func(id int, head, event string) map[string]any {
		return map[string]any{"id": id, "head_sha": head, "head_branch": "main", "event": event, "status": "completed", "conclusion": "success", "path": ".github/workflows/ci.yml", "pull_requests": []any{}}
	}
	pushRun := run(100, sha, "push")
	prRun := run(200, prHead, "pull_request")
	prRun["head_branch"] = "feature/release"
	prRun["pull_requests"] = []any{map[string]any{"number": 7, "head": map[string]any{"sha": prHead}, "base": map[string]any{"ref": "main"}}}
	pullBase := map[string]any{"ref": "main", "sha": other, "repo": map[string]any{"full_name": "faustbrian/go-library-tools"}}
	pull := map[string]any{"number": 7, "state": "closed", "merged": true, "merged_at": "2099-09-01T00:00:00Z", "merge_commit_sha": sha,
		"head": map[string]any{"sha": prHead, "repo": map[string]any{"full_name": "faustbrian/go-library-tools"}},
		"base": pullBase}
	jobs := []map[string]any{required, quality, codeql}
	pulls := []map[string]any{pull}
	prRuns := []any{prRun}
	pushRuns := []any{pushRun}
	switch defect {
	case "wrong-ci-head":
		pushRun["head_sha"] = other
	case "wrong-ci-branch":
		pushRun["head_branch"] = "feature/unmerged"
	case "failed-ci":
		pushRun["conclusion"] = "failure"
	case "newer-failed-ci":
		newer := run(101, sha, "push")
		newer["conclusion"] = "failure"
		pushRuns = []any{newer, pushRun}
	case "missing-required":
		jobs = jobs[1:]
	case "failed-required":
		required["conclusion"] = "failure"
	case "missing-quality":
		jobs = []map[string]any{required, codeql}
	case "skipped-quality":
		qualityStep["conclusion"] = "skipped"
	case "failed-quality":
		qualityStep["conclusion"] = "failure"
	case "missing-build":
		codeql["steps"] = []map[string]any{analyze}
	case "skipped-build":
		build["conclusion"] = "skipped"
	case "failed-build":
		build["conclusion"] = "failure"
	case "missing-analyze":
		codeql["steps"] = []map[string]any{build}
	case "skipped-analyze":
		analyze["conclusion"] = "skipped"
	case "failed-analyze":
		analyze["conclusion"] = "failure"
	case "wrong-analyze":
		analyze["name"] = "Run github/codeql-action/analyze@" + strings.Repeat("b", 40)
	case "missing-pr":
		pulls = []map[string]any{}
	case "unmerged-pr":
		pull["merged"], pull["merged_at"] = false, nil
	case "wrong-pr-merge":
		pull["merge_commit_sha"] = other
	case "wrong-pr-base":
		pullBase["ref"] = "develop"
	case "invalid-pr-base-sha":
		pullBase["sha"] = "refs/heads/main"
	case "missing-pr-base-sha":
		delete(pullBase, "sha")
	case "empty-pr-association":
		prRun["pull_requests"] = []any{}
	case "wrong-dependency-head":
		prRun["head_sha"] = other
	case "wrong-dependency-pr":
		prRun["pull_requests"] = []any{map[string]any{"number": 8}}
	case "missing-dependency-run":
		prRuns = []any{}
	case "failed-dependency-job":
		dependency["conclusion"] = "failure"
	case "newer-failed-dependency":
		newer := run(201, prHead, "pull_request")
		newer["conclusion"], newer["pull_requests"] = "failure", prRun["pull_requests"]
		prRuns = []any{newer, prRun}
	case "missing-dependency":
		dependency["steps"] = []map[string]any{}
	case "skipped-dependency":
		dependencyStep["conclusion"] = "skipped"
	case "failed-dependency":
		dependencyStep["conclusion"] = "failure"
	}
	directory := t.TempDir()
	for name, value := range map[string]any{
		"push-runs": map[string]any{"workflow_runs": pushRuns}, "push-jobs": map[string]any{"jobs": jobs},
		"pr-runs": map[string]any{"workflow_runs": prRuns}, "pr-jobs": map[string]any{"jobs": []any{dependency}},
		"pulls": pulls, "pull": pull,
	} {
		data, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(directory, name+".json"), data, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return directory
}

func releaseSecurityDocuments(t *testing.T, revision string) ([]byte, []byte) {
	t.Helper()
	names := []string{"codeql", "dependency-review", "go-vet", "gosec", "govulncheck", "license", "owned-analysis", "secret-current-tree", "secret-history", "staticcheck", "workflow-analysis"}
	scanners := make([]map[string]any, 0, len(names))
	for _, name := range names {
		scanners = append(scanners, map[string]any{
			"name": name, "tool_version": "test", "command": "test", "status": "passed",
			"completed_at": "2099-09-01T00:00:00Z", "result": "passed", "result_sha256": strings.Repeat("a", 64),
		})
	}
	matrix, err := json.Marshal(map[string]any{"schema_version": 1, "modules": []any{map[string]any{
		"module": "github.com/faustbrian/go-library-tools/v2", "revision": revision, "scanners": scanners,
		"release_verdict": map[string]any{"status": "pass", "owner": "maintainers", "decided_at": "2099-09-01T00:00:00Z", "rationale": "verified", "residual_risks": []string{"SEC-FLEET-001"}},
	}}})
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(matrix)
	risks := []byte(fmt.Sprintf(`{"schema_version":1,"risks":[{"id":"SEC-FLEET-001","module":"github.com/faustbrian/go-library-tools/v2","severity":"medium","status":"accepted","owner":"maintainers","rationale":"bounded adoption","mitigation":"explicit rows","review_condition":"on release","evidence":"sha256:%x:security-matrix.json","expires_at":"2099-10-01T00:00:00Z"}]}`, digest))
	return risks, matrix
}

func runGit(t *testing.T, directory string, args ...string) string {
	t.Helper()
	command := exec.CommandContext(t.Context(), "git", args...)
	command.Dir = directory
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v: %s", args, err, output)
	}
	return string(output)
}

func releaseWriteExecutable(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte("#!/usr/bin/env bash\nset -euo pipefail\n"+body), 0o700); err != nil {
		t.Fatal(err)
	}
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func TestReleaseWorkflowPublishesCatalogsOnlyForExplicitMilestones(t *testing.T) {
	content := readProjectFile(t, ".github/workflows/release.yml")
	for _, required := range []string{
		"  prepare-catalog:\n",
		"  project-catalog:\n",
		"  assemble-catalog:\n",
		"  verify-projection:\n",
		"  verify-catalog:\n",
		"  prepare-publication:\n",
		"  verify-publication:\n",
		"  attest-publication:\n",
		"max-parallel: 8",
		"persist-credentials: false",
		"submodules: false",
		"lfs: false",
		"gh attestation verify",
		"golib cohesion sources check",
		"golib cohesion sources verify",
		"golib cohesion catalog engineering --json",
		"golib cohesion aggregate generate",
		"golib cohesion aggregate check",
		"cmp --silent",
		"cohesion-sources.json",
		"cohesion-residuals.json",
		"cohesion-inputs.json",
		"cohesion-projections.tar.gz",
		"compatibility-sets.json",
		"compatibility-sets.md",
		"catalog-consumer.json",
		"catalog-consumer.md",
		"catalog-engineering.json",
		"catalog-engineering.md",
		"gzip -n",
		"needs: [build, verify-catalog]",
		"needs: prepare-publication",
		"needs: verify-publication",
		"needs: attest-publication",
		"if: vars.GOLIB_CATALOG_MILESTONE_TAG == github.ref_name",
		"PUBLISH_CATALOGS: ${{ vars.GOLIB_CATALOG_MILESTONE_TAG == github.ref_name }}",
	} {
		if !strings.Contains(content, required) {
			t.Errorf("release workflow lacks catalog publication contract %q", required)
		}
	}
	verifyProjectionStart := strings.Index(content, "  verify-projection:\n")
	verifyCatalogStart := strings.Index(content, "  verify-catalog:\n")
	if verifyProjectionStart < 0 || verifyCatalogStart < 0 || verifyProjectionStart >= verifyCatalogStart {
		t.Fatal("release workflow must independently verify projections before catalog verification")
	}
	verifyProjection := content[verifyProjectionStart:verifyCatalogStart]
	for _, required := range []string{
		"needs: [prepare-catalog, assemble-catalog]",
		"matrix: ${{ fromJSON(needs.prepare-catalog.outputs.matrix) }}",
		"ref: ${{ matrix.commit }}",
		"golib cohesion sources verify",
		"golib cohesion catalog engineering --json",
		"cmp --silent",
		"cohesion-projections.tar.gz",
	} {
		if !strings.Contains(verifyProjection, required) {
			t.Errorf("projection verifier lacks %q", required)
		}
	}
	verifyCatalog, _, found := strings.Cut(content[verifyCatalogStart:], "\n  prepare-publication:\n")
	if !found || !strings.Contains(verifyCatalog, "needs: [prepare-catalog, assemble-catalog, verify-projection]") {
		t.Fatal("catalog verification must wait for independent projection verification")
	}
	for _, required := range []string{
		"expected-members",
		"source_repository=",
		"input_repository=",
		"input_projection=",
		"sha256sum \"${projection}\"",
	} {
		if !strings.Contains(verifyCatalog, required) {
			t.Errorf("catalog verifier lacks source-lock binding %q", required)
		}
	}
	prepareStart := strings.Index(content, "  prepare-publication:\n")
	verifyStart := strings.Index(content, "  verify-publication:\n")
	attestStart := strings.Index(content, "  attest-publication:\n")
	publishStart := strings.Index(content, "  publish:\n")
	if prepareStart < 0 || verifyStart < 0 || attestStart < 0 || publishStart < 0 ||
		prepareStart >= verifyStart || verifyStart >= attestStart || attestStart >= publishStart {
		t.Fatal("release workflow must prepare, verify, attest, then publish immutable assets")
	}
	prepare := content[prepareStart:verifyStart]
	for _, required := range []string{
		"needs.verify-catalog.result == 'skipped'",
		"if: needs.verify-catalog.result == 'success'",
		"release-manifest.json",
		"checksums.txt",
		"name: release-publication",
	} {
		if !strings.Contains(prepare, required) {
			t.Errorf("publication preparation lacks %q", required)
		}
	}
	verify := content[verifyStart:attestStart]
	for _, required := range []string{
		"needs: prepare-publication",
		"PUBLISH_CATALOGS: ${{ vars.GOLIB_CATALOG_MILESTONE_TAG == github.ref_name }}",
		"--argjson expected_count",
		"name: release-publication",
		"sha256sum --check checksums.txt",
		"release-manifest.json",
		"checksums.txt",
	} {
		if !strings.Contains(verify, required) {
			t.Errorf("publication verification lacks %q", required)
		}
	}
	attest := content[attestStart:publishStart]
	if !strings.Contains(attest, "needs: verify-publication") ||
		!strings.Contains(attest, "actions/attest@") ||
		!strings.Contains(attest, "name: release-publication") {
		t.Error("attestation must consume the independently verified publication set")
	}
	publish := content[publishStart:]
	for _, required := range []string{
		"needs: attest-publication",
		"name: release-publication",
		"gh attestation verify",
		"gh release create",
	} {
		if !strings.Contains(publish, required) {
			t.Errorf("static publisher lacks %q", required)
		}
	}
	_, releaseCreate, found := strings.Cut(publish, "gh release create")
	if !found {
		t.Fatal("static publisher lacks release creation command")
	}
	if !strings.Contains(releaseCreate, `--repo "${GITHUB_REPOSITORY}"`) {
		t.Error("static publisher must bind release creation to the workflow repository")
	}
	for _, forbidden := range []string{
		"golib cohesion",
		"go run",
		"make ",
		"--clobber",
		">dist/release-manifest.json",
		">dist/checksums.txt",
		"actions/attest@",
	} {
		if strings.Contains(publish, forbidden) {
			t.Errorf("write-capable publish job contains forbidden execution %q", forbidden)
		}
	}
}

func TestReleaseWorkflowContinuesLeanPublicationAfterSkippedCatalogJobs(t *testing.T) {
	var workflow workflowDocument
	if err := yaml.Unmarshal([]byte(readProjectFile(t, ".github/workflows/release.yml")), &workflow); err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"prepare-publication": "always() && needs.build.result == 'success' && ((vars.GOLIB_CATALOG_MILESTONE_TAG != github.ref_name && needs.verify-catalog.result == 'skipped') || (vars.GOLIB_CATALOG_MILESTONE_TAG == github.ref_name && needs.verify-catalog.result == 'success'))",
		"verify-publication":  "always() && needs.prepare-publication.result == 'success'",
		"attest-publication":  "always() && needs.verify-publication.result == 'success'",
		"publish":             "always() && needs.attest-publication.result == 'success'",
	}
	for job, condition := range want {
		if got := workflow.Jobs[job].If; got != condition {
			t.Errorf("%s condition = %q, want %q", job, got, condition)
		}
	}
}

func TestConsumerUpgradeWorkflowIsBoundedAndReviewable(t *testing.T) {
	content := readProjectFile(t, ".github/workflows/update-consumers.yml")
	for _, required := range []string{
		"workflow_dispatch:",
		"version:",
		"workflow_sha:",
		"checksums_sha256:",
		"cohort:",
		"dry_run:",
		"golib consumers validate --json",
		"cohort must contain 1-10 repositories",
		"git fetch --no-tags --depth=1 origin",
		"gh release download",
		"actual_checksums_sha256",
		"ref: ${{ inputs.workflow_sha }}",
		"max-parallel: 5",
		"if: inputs.dry_run == false",
		"GOLIB_ROLLOUT_TOKEN",
		"git add .golib.yaml .github/workflows/ci.yml",
		"git ls-files --others --exclude-standard",
		"gh pr create",
	} {
		if !strings.Contains(content, required) {
			t.Errorf("consumer upgrade workflow lacks %q", required)
		}
	}
	for _, forbidden := range []string{"--force", "git add --all", "git add -A", "@main", "cancel-in-progress: true"} {
		if strings.Contains(content, forbidden) {
			t.Errorf("consumer upgrade workflow contains forbidden %q", forbidden)
		}
	}
	if regexp.MustCompile(`(?m)^\s*git add \.\s*$`).MatchString(content) {
		t.Error("consumer upgrade workflow contains wholesale staging")
	}
}

func projectRoot(t *testing.T) string {
	t.Helper()
	_, source, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("locate repository test source")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(source), "..", ".."))
}

func readProjectFile(t *testing.T, relative string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(projectRoot(t), filepath.FromSlash(relative)))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
