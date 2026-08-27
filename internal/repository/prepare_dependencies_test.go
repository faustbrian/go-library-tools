package repository_test

import (
	"bufio"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestDependencyWrapperDefersLocallyProxiedChecksums(t *testing.T) {
	root := t.TempDir()
	task := filepath.Join(root, "task")
	fakeBin := filepath.Join(root, "bin")
	isolated := filepath.Join(root, "isolated")
	localProxy := filepath.Join(root, "proxy")
	capture := filepath.Join(root, "captured.sum")
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
	exit 0
fi
printf 'unexpected fake go invocation: %s\n' "$*" >&2
exit 1
`)

	environmentFile := filepath.Join(root, "environment")
	command := exec.Command("bash", dependencyPreparationScript(t), task, environmentFile)
	command.Dir = root
	command.Env = append(os.Environ(), "PATH="+fakeBin+":"+os.Getenv("PATH"))
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("prepare dependencies: %v\n%s", err, output)
	}

	activeMod := filepath.Join(isolated, "state", "isolated.mod")
	activeSum := strings.TrimSuffix(activeMod, ".mod") + ".sum"
	writeRehearsalFile(t, activeMod, "module github.com/faustbrian/go-example\n")
	writeRehearsalFile(t, activeSum, strings.Join([]string{
		"github.com/faustbrian/go-local v1.0.0 h1:historical-local",
		"github.com/faustbrian/go-remote v1.0.0 h1:historical-remote",
	}, "\n")+"\n")
	writeRehearsalFile(t, filepath.Join(localProxy, "github.com", "faustbrian", "go-local", "@v", "v1.0.0.zip"), "local archive")

	wrapper := exec.Command(filepath.Join(task, "bin", "go"), "rehearsal-command")
	wrapper.Dir = root
	wrapper.Env = append(os.Environ(), readEnvironment(t, environmentFile)...)
	wrapper.Env = append(wrapper.Env,
		"GOFLAGS=-modfile="+activeMod,
		"GOLIB_ISOLATED_MODFILES_DIRECTORY="+isolated,
		"GOLIB_LOCAL_PROXY="+localProxy,
		"REHEARSAL_CAPTURE="+capture,
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
		t.Fatalf("historical checksums were not restored:\n%s", after)
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
	command := exec.Command("bash", dependencyPreparationScript(t), task, filepath.Join(root, "environment"))
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
