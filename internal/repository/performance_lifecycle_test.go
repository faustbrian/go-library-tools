package repository_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestCopiedToolingLifecycleScopesHiddenState(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	harness := filepath.Join(root, "lifecycle.sh")
	content := `#!/usr/bin/env bash
set -euo pipefail

source "$1"
root="$2"
source_root="${root}/source"
copied_tooling="${root}/copied"
mkdir -p "${source_root}/.golib"
printf 'original\n' >"${source_root}/.golib/marker"

inventory() {
    [[ ! -e "${source_root}/.golib" ]]
}
consumer_gate() {
    [[ "$(cat "${source_root}/.golib/marker")" == original ]]
    [[ ! -e "${copied_tooling}" ]]
}

hide_copied_tooling shared "${source_root}" "${copied_tooling}"
inventory
restore_copied_tooling shared "${source_root}" "${copied_tooling}"
consumer_gate

failure_source="${root}/failure-source"
failure_copy="${root}/failure-copy"
mkdir -p "${failure_source}/.golib"
printf 'failure\n' >"${failure_source}/.golib/marker"
status=0
(
    cleanup() {
        restore_copied_tooling shared "${failure_source}" "${failure_copy}"
    }
    trap cleanup EXIT
    hide_copied_tooling shared "${failure_source}" "${failure_copy}"
    [[ ! -e "${failure_source}/.golib" ]]
    exit 23
) || status=$?
[[ "${status}" -eq 23 ]]
[[ "$(cat "${failure_source}/.golib/marker")" == failure ]]
[[ ! -e "${failure_copy}" ]]

hide_copied_tooling legacy "${source_root}" "${copied_tooling}"
consumer_gate
restore_copied_tooling legacy "${source_root}" "${copied_tooling}"
consumer_gate

signal_source="${root}/signal-source"
signal_copy="${root}/signal-copy"
mkdir -p "${signal_source}/.golib"
printf 'signal\n' >"${signal_source}/.golib/marker"
status=0
(
    cleanup() {
        restore_copied_tooling shared "${signal_source}" "${signal_copy}"
    }
    trap cleanup EXIT
    trap 'cleanup; trap - EXIT TERM; exit 143' TERM
    hide_copied_tooling shared "${signal_source}" "${signal_copy}"
    kill -TERM "${BASHPID}"
) || status=$?
[[ "${status}" -ne 0 ]]
[[ "$(cat "${signal_source}/.golib/marker")" == signal ]]
[[ ! -e "${signal_copy}" ]]
`
	if err := os.WriteFile(harness, []byte(content), 0o700); err != nil {
		t.Fatal(err)
	}
	command := exec.CommandContext(
		t.Context(), "bash", harness,
		filepath.Join(projectRoot(t), "rehearsals", "copied-tooling-lifecycle.sh"), root,
	)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("copied tooling lifecycle: %v\n%s", err, output)
	}
}
