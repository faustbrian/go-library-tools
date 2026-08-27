#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 1 ]]; then
    printf 'usage: %s <content-digest-output>\n' "$0" >&2
    exit 2
fi

output="$1"
task="$(mktemp -d "${RUNNER_TEMP:-${TMPDIR:-/tmp}}/golib-parity-prepare.XXXXXX")"
cleanup() {
    chmod -R u+w "${task}" 2>/dev/null || true
    find "${task}" -depth -delete 2>/dev/null || true
}
trap cleanup EXIT HUP INT TERM
mkdir -p "${task}/cache" "${task}/mod" "${task}/tmp"

while IFS= read -r module_file; do
    module_directory="${module_file%/go.mod}"
    if [[ "${module_directory}" == "${module_file}" ]]; then
        module_directory='.'
    fi
    : >"${module_directory}/go.sum"
    (
        cd "${module_directory}"
        GOCACHE="${task}/cache" GOMODCACHE="${task}/mod" \
            GOTMPDIR="${task}/tmp" GOWORK=off go mod tidy
    )
done < <(find . -type f -name go.mod -not -path './.golib/*' | LC_ALL=C sort)

git ls-files -z | while IFS= read -r -d '' file; do
    case "${file}" in
        .golib/*|.github/workflows/*|Makefile) continue ;;
    esac
    [[ -f "${file}" && ! -L "${file}" ]] || continue
    printf '%s\0' "${file}"
    sha256sum "${file}" | awk '{printf "%s%c", $1, 0}'
done | sha256sum | awk '{print $1}' >"${output}"
