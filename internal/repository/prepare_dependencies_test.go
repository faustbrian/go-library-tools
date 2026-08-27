package repository_test

import (
	"bufio"
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestDependencyWrapperUsesParallelSafeExecutionModfiles(t *testing.T) {
	root := t.TempDir()
	task := filepath.Join(root, "task")
	fakeBin := filepath.Join(root, "bin")
	isolated := filepath.Join(root, "isolated")
	localProxy := filepath.Join(root, "proxy")
	capture := filepath.Join(root, "captured.sum")
	toolCapture := filepath.Join(root, "tool.flags")
	writeRehearsalFile(t, filepath.Join(root, "modules.json"), `{"modules":[{"directory":"."}]}`)
	writeRehearsalFile(t, filepath.Join(root, "go.mod"), `module github.com/faustbrian/go-example

go 1.26.6

require (
	github.com/faustbrian/go-local v1.0.0
	github.com/faustbrian/go-remote v1.0.0
)
`)
	writeRehearsalFile(t, filepath.Join(root, "go.sum"), strings.Join([]string{
		"github.com/faustbrian/go-local v1.0.0 h1:historical-local",
		"github.com/faustbrian/go-remote v1.0.0 h1:historical-remote",
	}, "\n")+"\n")
	writeExecutable(t, filepath.Join(fakeBin, "go"), `#!/usr/bin/env bash
set -euo pipefail
if [[ "${1:-}" == mod && "${2:-}" == edit ]]; then
	cat <<'JSON'
{"Module":{"Path":"github.com/faustbrian/go-example"},"Require":[{"Path":"github.com/faustbrian/go-local","Version":"v1.0.0"},{"Path":"github.com/faustbrian/go-remote","Version":"v1.0.0"}]}
JSON
	exit 0
fi
modfile=''
for flag in ${GOFLAGS:-}; do
	case "${flag}" in
		-modfile=*) modfile="${flag#-modfile=}" ;;
	esac
done
sumfile="${modfile%.mod}.sum"
if [[ "${1:-}" == mod && "${2:-}" == download ]]; then
	case "${3:-}" in
		github.com/faustbrian/go-local@v1.0.0)
			printf '%s\n' 'github.com/faustbrian/go-local v1.0.0 h1:current-local' >>"${sumfile}"
			;;
		github.com/faustbrian/go-remote@v1.0.0)
			printf '%s\n' 'github.com/faustbrian/go-remote v1.0.0 h1:current-remote' >>"${sumfile}"
			;;
	esac
	exit 0
fi
if [[ "${1:-}" == rehearsal-command ]]; then
	cp "${sumfile}" "${REHEARSAL_CAPTURE}"
	printf '%s\n' 'github.com/faustbrian/go-local v1.0.0 h1:local-proxy' >>"${sumfile}"
	if [[ "${REHEARSAL_FAILURE:-0}" == 1 ]]; then
		exit 23
	fi
	exit 0
fi
if [[ "${1:-}" == rehearsal-signal ]]; then
	cp "${sumfile}" "${REHEARSAL_CAPTURE}"
	printf '%s\n' 'github.com/faustbrian/go-local v1.0.0 h1:local-proxy' >>"${sumfile}"
	kill -TERM "${PPID}"
	exit 0
fi
if [[ "${1:-}" == rehearsal-block ]]; then
	printf '%s' "${modfile}" >"${REHEARSAL_BLOCK_MODFILE}"
	printf '%s\n' 'github.com/faustbrian/go-local v1.0.0 h1:local-proxy' >>"${sumfile}"
	: >"${REHEARSAL_BLOCK_READY}"
	while [[ ! -f "${REHEARSAL_BLOCK_RELEASE}" ]]; do
		sleep 0.01
	done
	exit 0
fi
if [[ "${1:-}" == run && "${2:-}" == *@* ]]; then
	printf '%s' "${GOFLAGS:-}" >"${REHEARSAL_TOOL_CAPTURE}"
	exit 0
fi
printf 'unexpected fake go invocation: %s\n' "$*" >&2
exit 1
`)

	environmentFile := filepath.Join(root, "environment")
	pathFile := filepath.Join(root, "path")
	command := exec.CommandContext(t.Context(), "bash", dependencyPreparationScript(t), task, environmentFile, pathFile)
	command.Dir = root
	command.Env = append(os.Environ(), "PATH="+fakeBin+":"+os.Getenv("PATH"))
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("prepare dependencies: %v\n%s", err, output)
	}
	environment := readEnvironment(t, environmentFile)
	wrapperPath := filepath.Join(task, "bin", "go")
	if !slices.Contains(environment, "GOLIB_REAL_GO="+wrapperPath) {
		t.Fatalf("legacy Go entrypoint does not select task-owned wrapper: %v", environment)
	}
	if path := strings.TrimSpace(readRehearsalFile(t, pathFile)); path != filepath.Join(task, "bin") {
		t.Fatalf("Go wrapper path = %q, want task-owned wrapper directory", path)
	}

	activeMod := filepath.Join(isolated, "state", "isolated.mod")
	activeSum := strings.TrimSuffix(activeMod, ".mod") + ".sum"
	writeRehearsalFile(t, activeMod, "module github.com/faustbrian/go-example\n")
	writeRehearsalFile(t, activeSum, strings.Join([]string{
		"github.com/faustbrian/go-local v1.0.0 h1:historical-local",
		"github.com/faustbrian/go-remote v1.0.0 h1:historical-remote",
	}, "\n")+"\n")
	writeRehearsalFile(t, filepath.Join(localProxy, "github.com", "faustbrian", "go-local", "@v", "v1.0.0.zip"), "local archive")

	wrapper := exec.CommandContext(t.Context(), wrapperPath, "rehearsal-command")
	wrapper.Dir = root
	wrapper.Env = append(os.Environ(), environment...)
	wrapper.Env = append(wrapper.Env,
		"GOFLAGS=-modfile="+activeMod,
		"GOLIB_ISOLATED_MODFILES_DIRECTORY="+isolated,
		"GOLIB_LOCAL_PROXY="+localProxy,
		"REHEARSAL_CAPTURE="+capture,
		"REHEARSAL_TOOL_CAPTURE="+toolCapture,
	)
	if output, err := wrapper.CombinedOutput(); err != nil {
		t.Fatalf("run dependency wrapper: %v\n%s", err, output)
	}

	during := readRehearsalFile(t, capture)
	if strings.Contains(during, "github.com/faustbrian/go-local ") {
		t.Fatalf("locally proxied checksum was preloaded:\n%s", during)
	}
	if !strings.Contains(during, "github.com/faustbrian/go-remote v1.0.0 h1:current-remote") {
		t.Fatalf("remote checksum was not refreshed:\n%s", during)
	}
	after := readRehearsalFile(t, activeSum)
	if !strings.Contains(after, "h1:historical-local") || !strings.Contains(after, "h1:historical-remote") || strings.Contains(after, "h1:local-proxy") {
		t.Fatalf("source-comparable checksums were not restored:\n%s", after)
	}

	wrapper.Env = append(wrapper.Env, "REHEARSAL_FAILURE=1")
	if output, err := wrapper.CombinedOutput(); err == nil {
		t.Fatalf("failing dependency command unexpectedly passed:\n%s", output)
	}
	afterFailure := readRehearsalFile(t, activeSum)
	if afterFailure != after {
		t.Fatalf("failed dependency command changed restored checksums:\n%s", afterFailure)
	}

	interrupted := exec.CommandContext(t.Context(), filepath.Join(task, "bin", "go"), "rehearsal-signal")
	interrupted.Dir = root
	interrupted.Env = wrapper.Env
	if output, err := interrupted.CombinedOutput(); err == nil {
		t.Fatalf("interrupted dependency command unexpectedly passed:\n%s", output)
	} else if exit := new(exec.ExitError); !errors.As(err, &exit) || exit.ExitCode() != 143 {
		t.Fatalf("interrupted dependency command error = %v, want exit 143\n%s", err, output)
	}
	if afterSignal := readRehearsalFile(t, activeSum); afterSignal != after {
		t.Fatalf("interrupted dependency command changed restored checksums:\n%s", afterSignal)
	}

	firstReady := filepath.Join(root, "first.ready")
	firstRelease := filepath.Join(root, "first.release")
	firstModfile := filepath.Join(root, "first.modfile")
	first := exec.CommandContext(t.Context(), filepath.Join(task, "bin", "go"), "rehearsal-block")
	first.Dir = root
	first.Env = slices.Clone(wrapper.Env)
	first.Env = append(first.Env, "REHEARSAL_BLOCK_READY="+firstReady, "REHEARSAL_BLOCK_RELEASE="+firstRelease, "REHEARSAL_BLOCK_MODFILE="+firstModfile)
	var firstOutput bytes.Buffer
	first.Stdout = &firstOutput
	first.Stderr = &firstOutput
	if err := first.Start(); err != nil {
		t.Fatal(err)
	}
	waitForRehearsalFile(t, firstReady)
	if used := readRehearsalFile(t, firstModfile); !strings.HasPrefix(used, filepath.Join(task, "execution")+string(filepath.Separator)) {
		t.Fatalf("overlapping dependency command modfile = %q, want task-owned execution path", used)
	}

	secondReady := filepath.Join(root, "second.ready")
	secondRelease := filepath.Join(root, "second.release")
	secondModfile := filepath.Join(root, "second.modfile")
	second := exec.CommandContext(t.Context(), filepath.Join(task, "bin", "go"), "rehearsal-block")
	second.Dir = root
	second.Env = slices.Clone(wrapper.Env)
	second.Env = append(second.Env, "REHEARSAL_BLOCK_READY="+secondReady, "REHEARSAL_BLOCK_RELEASE="+secondRelease, "REHEARSAL_BLOCK_MODFILE="+secondModfile)
	var secondOutput bytes.Buffer
	second.Stdout = &secondOutput
	second.Stderr = &secondOutput
	if err := second.Start(); err != nil {
		t.Fatal(err)
	}
	waitForRehearsalFile(t, secondReady)
	writeRehearsalFile(t, firstRelease, "release")
	if err := first.Wait(); err != nil {
		t.Fatalf("first overlapping dependency command: %v\n%s", err, firstOutput.Bytes())
	}
	writeRehearsalFile(t, secondRelease, "release")
	if err := second.Wait(); err != nil {
		t.Fatalf("second overlapping dependency command: %v\n%s", err, secondOutput.Bytes())
	}
	if afterOverlap := readRehearsalFile(t, activeSum); afterOverlap != after {
		t.Fatalf("overlapping dependency commands changed source-comparable checksums:\n%s", afterOverlap)
	}

	tool := exec.CommandContext(t.Context(), filepath.Join(task, "bin", "go"), "run", "example.com/tool@v1.0.0")
	tool.Dir = root
	tool.Env = wrapper.Env
	if output, err := tool.CombinedOutput(); err != nil {
		t.Fatalf("run external tool: %v\n%s", err, output)
	}
	if flags := readRehearsalFile(t, toolCapture); flags != "" {
		t.Fatalf("external tool inherited consumer GOFLAGS %q", flags)
	}
}

func waitForRehearsalFile(t *testing.T, path string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for {
		if _, err := os.Stat(path); err == nil {
			return
		} else if !errors.Is(err, os.ErrNotExist) {
			t.Fatal(err)
		}
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for %s", filepath.Base(path))
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestDependencyPreparationAcceptsNoOwnedDependencies(t *testing.T) {
	root := t.TempDir()
	task := filepath.Join(root, "task")
	fakeBin := filepath.Join(root, "bin")
	writeRehearsalFile(t, filepath.Join(root, "modules.json"), `{"modules":[{"directory":"."}]}`)
	writeRehearsalFile(t, filepath.Join(root, "go.mod"), "module github.com/faustbrian/go-example\n\ngo 1.26.6\n")
	writeExecutable(t, filepath.Join(fakeBin, "go"), `#!/usr/bin/env bash
set -euo pipefail
if [[ "${1:-}" == mod && "${2:-}" == edit ]]; then
	printf '%s\n' '{"Module":{"Path":"github.com/faustbrian/go-example"}}'
	exit 0
fi
printf 'unexpected fake go invocation: %s\n' "$*" >&2
exit 1
`)
	command := exec.CommandContext(t.Context(), "bash", dependencyPreparationScript(t), task,
		filepath.Join(root, "environment"), filepath.Join(root, "path"))
	command.Dir = root
	command.Env = append(os.Environ(), "PATH="+fakeBin+":"+os.Getenv("PATH"))
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("prepare dependencies: %v\n%s", err, output)
	}
	if info, err := os.Stat(filepath.Join(task, "bin", "go")); err != nil || info.Mode()&0o100 == 0 {
		t.Fatalf("generated wrapper = %v, %v", info, err)
	}
}

func dependencyPreparationScript(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve test source")
	}
	return filepath.Join(filepath.Dir(file), "..", "..", "rehearsals", "prepare-dependencies.sh")
}

func readEnvironment(t *testing.T, path string) []string {
	t.Helper()
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	var values []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		values = append(values, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	return values
}

func writeExecutable(t *testing.T, path, content string) {
	t.Helper()
	writeRehearsalFile(t, path, content)
	if err := os.Chmod(path, 0o700); err != nil {
		t.Fatal(err)
	}
}

func writeRehearsalFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func readRehearsalFile(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(content)
}
