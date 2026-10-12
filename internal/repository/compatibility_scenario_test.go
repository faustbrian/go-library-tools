package repository_test

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"syscall"
	"testing"
	"time"

	"go.yaml.in/yaml/v3"
)

func TestCompatibilityConsumerSelectionRetainsSharedAndUnknownQualification(t *testing.T) {
	script := readProjectFile(t, ".github/scripts/select-compatibility-consumer.py")
	for _, test := range []struct {
		name, changed, alsoChanged, mode string
		full, none, refuse               bool
	}{
		{name: "unrelated root metadata", changed: "modules.json", none: true},
		{name: "one independent file", changed: "release/compatibility-consumer/wsdl2_composition_test.go"},
		{name: "consumer dependency", changed: "release/compatibility-consumer/go.mod", full: true},
		{name: "checksum input", changed: "release/compatibility-consumer/go.sum", full: true},
		{name: "toolchain", changed: ".go-version", full: true},
		{name: "root dependency", changed: "go.mod", none: true},
		{name: "root checksum", changed: "go.sum", none: true},
		{name: "root dependency and scenario", changed: "go.mod", alsoChanged: "release/compatibility-consumer/wsdl2_composition_test.go"},
		{name: "root checksum and consumer dependency", changed: "go.sum", alsoChanged: "release/compatibility-consumer/go.mod", full: true},
		{name: "root dependency and publication", changed: "go.mod", mode: "publication", full: true},
		{name: "selector", changed: ".github/scripts/select-compatibility-consumer.py", full: true},
		{name: "generator", changed: "tools/compatibility/new.py", full: true},
		{name: "publication", mode: "publication", full: true},
		{name: "deleted scenario", changed: "release/compatibility-consumer/wsdl2_composition_test.go", mode: "delete", full: true},
		{name: "renamed scenario", changed: "release/compatibility-consumer/wsdl2_composition_test.go", mode: "rename", full: true},
		{name: "unknown consumer input", changed: "release/compatibility-consumer/input.json", full: true},
		{name: "new helper", changed: "release/compatibility-consumer/helper.go", full: true},
		{name: "nested source", changed: "release/compatibility-consumer/nested/helper.go", refuse: true},
		{name: "invalid immutable base", mode: "invalid", refuse: true},
		{name: "missing immutable base", mode: "missing", refuse: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			for _, name := range []string{"go.mod", "go.sum", "wsdl2_composition_test.go", "openapi2_composition_test.go", "consumer_test.go"} {
				writeFile(t, root, "release/compatibility-consumer/"+name, "initial\n")
			}
			if test.changed != "" && test.mode != "rename" {
				writeFile(t, root, test.changed, "initial\n")
			}
			runGit(t, root, "init", "-q")
			runGit(t, root, "add", "--", ".")
			commit := func() {
				runGit(t, root, "-c", "core.hooksPath=/dev/null", "-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "-c", "commit.gpgsign=false", "commit", "-qm", "fixture")
			}
			commit()
			base := strings.TrimSpace(runGit(t, root, "rev-parse", "HEAD"))
			if test.changed != "" {
				switch test.mode {
				case "delete":
					if err := os.Remove(filepath.Join(root, test.changed)); err != nil {
						t.Fatal(err)
					}
				case "rename":
					if err := os.Rename(filepath.Join(root, test.changed), filepath.Join(root, "release/compatibility-consumer/renamed_test.go")); err != nil {
						t.Fatal(err)
					}
				default:
					writeFile(t, root, test.changed, "changed\n")
				}
				if test.alsoChanged != "" {
					writeFile(t, root, test.alsoChanged, "changed\n")
				}
				runGit(t, root, "add", "--", ".")
				commit()
			}
			if test.mode == "invalid" {
				base = "main"
			}
			if test.mode == "missing" {
				base = strings.Repeat("a", 40)
			}
			publication := "false"
			if test.mode == "publication" {
				publication = "true"
			}
			outputPath := filepath.Join(root, "output")
			command := exec.CommandContext(t.Context(), "python3", "-c", script)
			command.Dir = root
			command.Env = append(os.Environ(), "BASE_SHA="+base, "PUBLICATION_REQUIRED="+publication, "GITHUB_OUTPUT="+outputPath)
			data, err := command.CombinedOutput()
			if (err != nil) != test.refuse {
				t.Fatalf("selection refusal=%v want %t: %s", err, test.refuse, data)
			}
			output, _ := os.ReadFile(outputPath)
			if test.refuse {
				if len(output) != 0 {
					t.Fatalf("invalid scope published output: %s", output)
				}
				return
			}
			parts := strings.Split(strings.TrimSpace(string(output)), "\n")
			if len(parts) != 2 {
				t.Fatalf("missing selection: %q", output)
			}
			var files []string
			if err := json.Unmarshal([]byte(strings.TrimPrefix(parts[1], "files=")), &files); err != nil {
				t.Fatal(err)
			}
			if parts[0] != fmt.Sprintf("required=%t", len(files) > 0) {
				t.Fatalf("required output disagrees with selected files: %q", output)
			}
			switch {
			case test.full:
				paths, err := filepath.Glob(filepath.Join(root, "release/compatibility-consumer/*.go"))
				if err != nil {
					t.Fatal(err)
				}
				var expected []string
				for _, path := range paths {
					expected = append(expected, filepath.Base(path))
				}
				slices.Sort(expected)
				if !slices.Equal(files, expected) {
					t.Fatalf("shared/unknown input selected %v, want every source %v", files, expected)
				}
			case test.none:
				if len(files) != 0 {
					t.Fatalf("unrelated root inputs rebuilt: %v", files)
				}
			case len(files) != 1 || files[0] != "wsdl2_composition_test.go":
				t.Fatalf("scenario inputs: %v", files)
			}
		})
	}
}

func TestCompatibilityConsumerCancellationStopsChildAndCleansResources(t *testing.T) {
	for _, admission := range []bool{false, true} {
		t.Run(fmt.Sprintf("admission=%t", admission), func(t *testing.T) {
			testCompatibilityConsumerCancellation(t, admission)
		})
	}
}

const compatibilityConsumerPIDPublication = `    ready=root/'child.pid.tmp'
    ready.write_text(str(os.getpid()))
    ready.replace(root/'child.pid')
`

// Exercise the cancellation fixture's readiness boundary without delivering
// signals or starting a long-lived child. File existence must imply a complete
// PID, including the interval between opening and committing the payload.
func TestCompatibilityConsumerPIDReadinessRequiresCommittedPayload(t *testing.T) {
	root := t.TempDir()
	script := `import os,pathlib
root=pathlib.Path(os.environ['PROBE_ROOT'])
original=pathlib.Path.write_text
def observe(path,data,*args,**kwargs):
    if path.name in ('child.pid','child.pid.tmp'):
        with path.open('w') as stream:
            if (root/'child.pid').exists():
                raise AssertionError('child.pid became visible before its PID payload was committed')
            return stream.write(data)
    return original(path,data,*args,**kwargs)
pathlib.Path.write_text=observe
if True:
` + compatibilityConsumerPIDPublication
	command := exec.CommandContext(t.Context(), "python3", "-c", script)
	command.Env = append(os.Environ(), "PROBE_ROOT="+root)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("PID readiness publication failed: %v: %s", err, output)
	}
	data, err := os.ReadFile(filepath.Join(root, "child.pid"))
	if err != nil {
		t.Fatal(err)
	}
	var pid int
	if _, err := fmt.Sscanf(string(data), "%d", &pid); err != nil || pid <= 0 {
		t.Fatalf("published readiness did not contain a complete PID: %q / %v", data, err)
	}
	if _, err := os.Stat(filepath.Join(root, "child.pid.tmp")); !os.IsNotExist(err) {
		t.Fatalf("staging file remains: %v", err)
	}
}

func testCompatibilityConsumerCancellation(t *testing.T, admission bool) {
	t.Helper()
	root := t.TempDir()
	for _, name := range []string{"go.mod", "go.sum", "wsdl2_composition_test.go"} {
		writeFile(t, root, "release/compatibility-consumer/"+name, "fixture\n")
	}
	writeFile(t, root, "runner.sh", readProjectFile(t, ".github/scripts/run-compatibility-consumer.sh"))
	writeExecutable(t, filepath.Join(root, "bin", "go"), `#!/usr/bin/env python3
import os,pathlib,signal,sys,time
root=pathlib.Path(os.environ['PROBE_ROOT'])
arguments=sys.argv[1:]
if arguments==['env','GOVERSION']: print('go1.27.2')
elif arguments==['env','GOOS']: print('linux')
elif arguments==['env','GOARCH']: print('amd64')
elif arguments==['mod','edit','-json']: print('{"Replace":null}')
elif arguments==['mod','tidy']: pass
elif arguments[0]=='test':
    def stop(signum, frame):
        (root/'child-stopped').write_text('terminated')
        sys.exit(143)
    signal.signal(signal.SIGTERM,stop)
`+compatibilityConsumerPIDPublication+`
    time.sleep(30)
else: sys.exit(64)
`)
	command := exec.CommandContext(t.Context(), "bash", "runner.sh")
	if admission {
		// DEBUG executes the actual runner's boundary before PID registration.
		// Waiting for the double proves launch happened before delivering TERM.
		command = exec.CommandContext(t.Context(), "bash", "-T", "-c", `
trap 'if [[ "$BASH_SUBSHELL" == 0 && "$BASH_COMMAND" == "child=\$!" && "${2:-}" == test ]]; then
  while [[ ! -f "$PROBE_ROOT/child.pid" ]]; do sleep .01; done
  kill -TERM "$$"
fi' DEBUG
source runner.sh
`)
	}
	command.Dir = root
	command.Env = append(os.Environ(), "PATH="+filepath.Join(root, "bin")+string(os.PathListSeparator)+os.Getenv("PATH"), "RUNNER_TEMP="+root, "SCENARIO_FILES=[\"wsdl2_composition_test.go\"]", "EXPECTED_GOOS=linux", "EXPECTED_GOARCH=amd64", "PROBE_ROOT="+root)
	if err := command.Start(); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- command.Wait() }()
	defer func() { _ = command.Process.Kill() }()
	deadline := time.Now().Add(5 * time.Second)
	for {
		if _, err := os.Stat(filepath.Join(root, "child.pid")); err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("selected command did not start")
		}
		time.Sleep(10 * time.Millisecond)
	}
	data, err := os.ReadFile(filepath.Join(root, "child.pid"))
	if err != nil {
		t.Fatal(err)
	}
	var pid int
	if _, err := fmt.Sscanf(string(data), "%d", &pid); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = syscall.Kill(pid, syscall.SIGKILL) }()
	if !admission {
		if err := command.Process.Signal(syscall.SIGTERM); err != nil {
			t.Fatal(err)
		}
	}
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("cancellation reported success")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("cancelled runner did not stop promptly")
	}
	if _, err := os.Stat(filepath.Join(root, "child-stopped")); err != nil {
		t.Fatal("cancellation left the selected child running")
	}
	leftovers, err := filepath.Glob(filepath.Join(root, "golib-compatibility-consumer.*"))
	if err != nil || len(leftovers) != 0 {
		t.Fatalf("cancellation retained consumer resources: %v %v", leftovers, err)
	}
}

func TestCompatibilityConsumerStagesOnlySelectedSourcesAndPropagatesFailure(t *testing.T) {
	var workflow workflowDocument
	if err := yaml.Unmarshal([]byte(readProjectFile(t, ".github/workflows/ci.yml")), &workflow); err != nil {
		t.Fatal(err)
	}
	var script string
	for _, step := range workflow.Jobs["compatibility-consumer"].Steps {
		if step.Name == "Resolve and test the public compatibility consumer" {
			script = step.Run
		}
	}
	if script == "" {
		t.Fatal("missing consumer execution")
	}
	for _, test := range []struct {
		name, files                    string
		missing, failTest, wantFailure bool
	}{
		{name: "one scenario", files: `["wsdl2_composition_test.go"]`},
		{name: "two scenarios", files: `["wsdl2_composition_test.go","openapi2_composition_test.go"]`},
		{name: "missing selected source", files: `["wsdl2_composition_test.go"]`, missing: true, wantFailure: true},
		{name: "invalid selection", files: `["../outside.go"]`, wantFailure: true},
		{name: "empty selection", files: `[]`, wantFailure: true},
		{name: "malformed selection", files: `{`, wantFailure: true},
		{name: "selected test fails", files: `["wsdl2_composition_test.go"]`, failTest: true, wantFailure: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			for _, path := range []string{"go.mod", "go.sum", "consumer_test.go", "composition_test.go", "workflow_composition_test.go", "wire3_composition_test.go", "qcp3_composition_test.go", "postgres2_composition_test.go", "state_machine2_composition_test.go", "wsdl2_composition_test.go", "openapi2_composition_test.go", "outbox2_composition_test.go", "tabular_xlsx_composition_test.go"} {
				if test.missing && path == "wsdl2_composition_test.go" {
					continue
				}
				writeFile(t, root, "release/compatibility-consumer/"+path, "fixture\n")
			}
			if data, err := os.ReadFile(filepath.Join(projectRoot(t), ".github/scripts/run-compatibility-consumer.sh")); err == nil {
				writeFile(t, root, ".github/scripts/run-compatibility-consumer.sh", string(data))
			}
			log := filepath.Join(root, "executed")
			writeExecutable(t, filepath.Join(root, "bin", "go"), `#!/usr/bin/env bash
set -euo pipefail
case "$*" in
  'env GOVERSION') echo go1.27.2 ;;
  'env GOOS') echo linux ;;
  'env GOARCH') echo amd64 ;;
  'mod edit -json') echo '{"Replace":null}' ;;
  'mod tidy')
    python3 - <<'PY'
import glob,json,os
assert sorted(glob.glob('*_test.go')) == sorted(json.loads(os.environ['SCENARIO_FILES'])), 'unrelated source compiled'
assert os.path.isfile('go.sum'), 'checksum input lost'
PY
    ;;
  test*) printf '%s\n' "$*" >>"$EXECUTED"; [[ "$FAIL_TEST" == false ]] ;;
  *) exit 64 ;;
esac
`)
			command := exec.CommandContext(t.Context(), "bash", "-e", "-c", strings.NewReplacer("${{ matrix.goos }}", "linux", "${{ matrix.goarch }}", "amd64").Replace(script))
			command.Dir = root
			command.Env = append(os.Environ(), "PATH="+filepath.Join(root, "bin")+string(os.PathListSeparator)+os.Getenv("PATH"), "RUNNER_TEMP="+root, "SCENARIO_FILES="+test.files, "EXPECTED_GOOS=linux", "EXPECTED_GOARCH=amd64", "EXECUTED="+log, "FAIL_TEST="+map[bool]string{true: "true", false: "false"}[test.failTest])
			output, err := command.CombinedOutput()
			if (err != nil) != test.wantFailure {
				t.Fatalf("consumer failure=%v, want %t: %s", err, test.wantFailure, output)
			}
			if !test.wantFailure {
				if _, err := os.Stat(log); err != nil {
					t.Fatal("selected tests did not execute")
				}
			} else if !test.failTest {
				if _, err := os.Stat(log); !os.IsNotExist(err) {
					t.Fatal("invalid selection reached test execution")
				}
			} else if _, err := os.Stat(log); err != nil {
				t.Fatal("test failure did not originate from selected tests")
			}
			leftovers, err := filepath.Glob(filepath.Join(root, "golib-compatibility-consumer.*"))
			if err != nil || len(leftovers) != 0 {
				t.Fatalf("consumer resources retained: %v %v", leftovers, err)
			}
		})
	}
}
