#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 1 ]]; then
    printf 'usage: %s <content-digest-output>\n' "$0" >&2
    exit 2
fi

output="$1"
before="$(git status --porcelain=v1 --untracked-files=all)"

git ls-files -z | while IFS= read -r -d '' file; do
    case "${file}" in
        .golib/*|.github/workflows/*|Makefile) continue ;;
    esac
    [[ -f "${file}" && ! -L "${file}" ]] || continue
    printf '%s\0' "${file}"
    sha256sum "${file}" | awk '{printf "%s%c", $1, 0}'
done | sha256sum | awk '{print $1}' >"${output}"

after="$(git status --porcelain=v1 --untracked-files=all)"
if [[ "${before}" != "${after}" ]]; then
    printf 'parity preparation modified repository content\n' >&2
    exit 1
fi
