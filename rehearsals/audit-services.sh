#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 1 ]]; then
    printf 'usage: %s <status-output>\n' "$0" >&2
    exit 2
fi

output="$1"
status=0
containers="$(docker ps -a --format '{{.Names}}' | grep -E '^(golib-|codex-rabbitstream-)' || true)"
networks="$(docker network ls --format '{{.Name}}' | grep -E '^(golib-|codex-rabbitstream-)' || true)"
volumes="$(docker volume ls --format '{{.Name}}' | grep -E '^(golib-|codex-rabbitstream-)' || true)"
locks=""
if [[ -e "${TMPDIR:-/tmp}/golib-rabbitstream-fixture.lock" ]]; then
    locks="${TMPDIR:-/tmp}/golib-rabbitstream-fixture.lock"
fi

for resource in "${containers}" "${networks}" "${volumes}" "${locks}"; do
    if [[ -n "${resource}" ]]; then
        status=1
    fi
done
printf '%s\n' "${status}" >"${output}"

if [[ "${status}" -ne 0 ]]; then
    printf 'service cleanup left task-owned resources\n' >&2
    [[ -z "${containers}" ]] || printf 'containers:\n%s\n' "${containers}" >&2
    [[ -z "${networks}" ]] || printf 'networks:\n%s\n' "${networks}" >&2
    [[ -z "${volumes}" ]] || printf 'volumes:\n%s\n' "${volumes}" >&2
    [[ -z "${locks}" ]] || printf 'locks:\n%s\n' "${locks}" >&2
fi

exit "${status}"
