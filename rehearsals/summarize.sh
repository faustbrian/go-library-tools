#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 8 ]]; then
    printf 'usage: %s <mode> <log> <content-digest> <check-status> <release-log> <release-status> <service-cleanup-status> <output>\n' "$0" >&2
    exit 2
fi

mode="$1"
log="$2"
content="$3"
check_status="$4"
release_log="$5"
release_status="$6"
service_cleanup_status="$7"
output="$8"
[[ "${mode}" == legacy || "${mode}" == shared ]]
[[ -s "${log}" && -s "${content}" && -s "${check_status}" && -s "${release_log}" && -s "${release_status}" && -s "${service_cleanup_status}" ]]
[[ -s modules.json ]]

directory="$(dirname "${output}")"
mkdir -p "${directory}"
coverage="${directory}/${mode}-coverage.tsv"
mutation="${directory}/${mode}-mutation.tsv"
modules="${directory}/${mode}-modules.tsv"
gates="${directory}/${mode}-gates.tsv"
services="${directory}/${mode}-services.tsv"
releases="${directory}/${mode}-releases.tsv"
advisories="${directory}/${mode}-advisories.tsv"

perl -ne '
    if (/^(\S+)\s+(\d+)\/(\d+) statements$/) {
        print "$1\t$2\t$3\n";
    }
' "${log}" | LC_ALL=C sort -u >"${coverage}"
if [[ "${mode}" == legacy ]]; then
    while IFS= read -r report; do
        jq -r '
          .module as $module |
          .packages[] |
          [
            $module,
            (if .package == "." then "." else "./" + .package end),
            ([.report.files[].mutations[]?] | length)
          ] | @tsv
        ' "${report}"
    done < <(find .artifacts -type f -name mutation.json | LC_ALL=C sort)
else
    perl -ne '
        if (/^\[([^]]+)\]\s+(\S+) killed (\d+)\/\d+ viable mutants$/) {
            print "$1\t$2\t$3\n";
        } elsif (/^\[([^]]+)\]\s+(\S+) reused content-identical mutation evidence \((\d+) mutants\)$/) {
            print "$1\t$2\t$3\n";
        } elsif (/^\[([^]]+)\]\s+(\S+) has an exact zero-viable-mutant review$/) {
            print "$1\t$2\t0\n";
        }
    ' "${log}"
fi | LC_ALL=C sort -u >"${mutation}"
[[ -s "${coverage}" && -s "${mutation}" ]]
jq -r '.modules[] | [.directory, .module_path] | @tsv' modules.json | LC_ALL=C sort -u >"${modules}"
jq -r '.modules[] as $module | $module.gates | to_entries[] | select(.value == true) | [$module.directory, .key] | @tsv' modules.json | LC_ALL=C sort -u >"${gates}"
jq -r '.modules[] as $module | $module.required_services[] | [$module.directory, .] | @tsv' modules.json | LC_ALL=C sort -u >"${services}"
jq -r '.modules[] | select(.releasable == true) | [.directory, .tag_prefix] | @tsv' modules.json | LC_ALL=C sort -u >"${releases}"
while IFS= read -r module; do
    status=0
    if [[ "${mode}" == legacy ]]; then
        line="$(grep -F "[${module}] NilAway advisory exit status: " "${log}")"
        [[ "$(printf '%s\n' "${line}" | wc -l | tr -d ' ')" == 1 ]]
        status="${line##*: }"
        [[ "${status}" =~ ^[0-9]+$ ]]
        if [[ "${status}" -ne 0 ]]; then
            status=1
        fi
    elif grep -Fq "[${module}] NilAway advisory:" "${log}"; then
        status=1
    fi
    printf '%s\t%s\n' "${module}" "${status}"
done < <(jq -r '.modules[] | select(.gates.lint == true) | .directory' modules.json | LC_ALL=C sort -u) >"${advisories}"
[[ -s "${modules}" && -s "${gates}" && -s "${releases}" ]]

release_decision=success
if [[ "$(cat "${release_status}")" -ne 0 ]]; then
    release_decision=failure
    if grep -Fq 'release tag already exists:' "${release_log}"; then
        release_decision=tag-collision
    fi
fi

jq -n \
    --arg mode "${mode}" \
    --arg content_digest "$(cat "${content}")" \
    --arg check_status "$(cat "${check_status}")" \
    --arg release_status "$(cat "${release_status}")" \
    --arg release_decision "${release_decision}" \
    --arg service_cleanup_status "$(cat "${service_cleanup_status}")" \
    --rawfile coverage "${coverage}" \
    --rawfile mutation "${mutation}" \
    --rawfile modules "${modules}" \
    --rawfile gates "${gates}" \
    --rawfile services "${services}" \
    --rawfile releases "${releases}" \
    --rawfile advisories "${advisories}" \
    '{
      schema_version: 1,
      mode: $mode,
      content_digest: $content_digest,
      check_status: ($check_status | tonumber),
      release_status: ($release_status | tonumber),
      release_decision: $release_decision,
      service_cleanup_status: ($service_cleanup_status | tonumber),
      nilaway_advisories: ($advisories | split("\n") | map(select(length > 0))),
      selected_modules: ($modules | split("\n") | map(select(length > 0))),
      effective_gates: ($gates | split("\n") | map(select(length > 0))),
      required_services: ($services | split("\n") | map(select(length > 0))),
      release_units: ($releases | split("\n") | map(select(length > 0))),
      coverage: ($coverage | split("\n") | map(select(length > 0))),
      mutation: ($mutation | split("\n") | map(select(length > 0)))
    }' >"${output}"
