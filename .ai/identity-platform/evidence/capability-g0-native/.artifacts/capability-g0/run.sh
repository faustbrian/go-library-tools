#!/usr/bin/env bash
set -euo pipefail
name="$1"
shift
root="$(git rev-parse --show-toplevel)"
artifact="${root}/.artifacts/capability-g0"
task_cache="$(mktemp -d /tmp/capability-g0-go.XXXXXX)"
cleanup() {
    chmod -R u+w "${task_cache}"
    find "${task_cache}" -depth -delete
}
trap cleanup EXIT HUP INT TERM
mkdir -p "${task_cache}/build" "${task_cache}/mod" "${task_cache}/tmp"
export GOCACHE="${task_cache}/build" GOMODCACHE="${task_cache}/mod" GOTMPDIR="${task_cache}/tmp" GOTOOLCHAIN=go1.26.6
{
    git rev-parse HEAD
    git ls-files -co --exclude-standard -z | sort -z | xargs -0 shasum -a 256
    printf 'command:'
    printf ' %q' "$@"
    printf '\n'
    go version
} >"${artifact}/${name}.inputs"
set +e
"$@" 2>&1 | tee "${artifact}/${name}.log"
status=${PIPESTATUS[0]}
set -e
printf '%s\n' "${status}" >"${artifact}/${name}.exit"
exit "${status}"
