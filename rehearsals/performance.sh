#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 8 ]]; then
    printf 'usage: %s <legacy|shared> <source-root> <golib> <repository> <revision> <content-sha256> <core|service> <tooling-revision>\n' "$0" >&2
    exit 2
fi

implementation="$1"
source_root="$(cd "$2" && pwd -P)"
golib="$3"
repository="$4"
revision="$5"
content_sha256="$6"
profile="$7"
tooling_revision="$8"
[[ "${implementation}" == legacy || "${implementation}" == shared ]]
[[ "${profile}" == core || "${profile}" == service ]]
[[ "${revision}" =~ ^[0-9a-f]{40}$ ]]
[[ "${content_sha256}" =~ ^[0-9a-f]{64}$ ]]
[[ "${tooling_revision}" =~ ^[0-9a-f]{40}$ ]]
[[ "${GITHUB_RUN_ID:-}" =~ ^[1-9][0-9]*$ ]]
[[ "${GITHUB_RUN_ATTEMPT:-}" =~ ^[1-9][0-9]*$ ]]
[[ "${GITHUB_REPOSITORY:-}" == */* ]]

task="$(mktemp -d "${RUNNER_TEMP:-${TMPDIR:-/tmp}}/golib-performance.XXXXXX")"
samples="${task}/samples.jsonl"
runtime="${task}/runtime"
copied_tooling="${task}/copied-tooling"
mkdir -p "${runtime}"
: >"${samples}"

cleanup() {
    if [[ -d "${copied_tooling}" && ! -e "${source_root}/.golib" ]]; then
        mv "${copied_tooling}" "${source_root}/.golib"
    fi
    chmod -R u+w "${task}" 2>/dev/null || true
    find "${task}" -depth -delete 2>/dev/null || true
}
trap cleanup EXIT HUP INT TERM

if [[ "${implementation}" == shared ]]; then
    mv "${source_root}/.golib" "${copied_tooling}"
fi

record_command() {
    local metric="$1"
    local expected="$2"
    local repetitions="$3"
    shift 3
    local index status started finished wall_ns peak_rss stats command_output
    local reuse_count
    for index in $(seq 1 "${repetitions}"); do
        stats="${task}/time-${implementation}-${metric}-${index}.txt"
        command_output="${task}/output-${implementation}-${metric}-${index}.log"
        started="$(date +%s%N)"
        set +e
        (
            cd "${source_root}"
            TMPDIR="${runtime}" /usr/bin/time --quiet -f '%M' -o "${stats}" \
                "$@" >"${command_output}" 2>&1
        )
        status=$?
        set -e
        finished="$(date +%s%N)"
        if [[ "${status}" -ne "${expected}" ]]; then
            printf '%s %s sample %d exited %d, want %d\n' \
                "${implementation}" "${metric}" "${index}" "${status}" "${expected}" >&2
            exit 1
        fi
        wall_ns="$((finished - started))"
        peak_rss="$(tail -1 "${stats}")"
        reuse_count=null
        if [[ "${metric}" == checkpoint-reuse ]]; then
            reuse_count="$(grep -Fc 'reused content-identical mutation evidence' "${command_output}" || true)"
            if [[ "${reuse_count}" -ne "${mutation_package_count}" ]]; then
                printf '%s %s sample %d reused %d packages, want %d\n' \
                    "${implementation}" "${metric}" "${index}" \
                    "${reuse_count}" "${mutation_package_count}" >&2
                exit 1
            fi
        fi
        jq -cn \
            --arg implementation "${implementation}" \
            --arg metric "${metric}" \
            --argjson sample "${index}" \
            --argjson wall_ns "${wall_ns}" \
            --argjson peak_rss_kib "${peak_rss}" \
            --argjson exit_code "${status}" \
            --argjson reuse_count "${reuse_count}" \
            '{implementation:$implementation,metric:$metric,sample:$sample,wall_ns:$wall_ns,peak_rss_kib:$peak_rss_kib,exit_code:$exit_code,reuse_count:$reuse_count}' \
            >>"${samples}"
    done
}

run_api_module() {
    local module="$1"
    if [[ "${implementation}" == legacy ]]; then
        "${source_root}/.golib/scripts/with-disposable-go-cache.sh" \
            "${source_root}/.golib/scripts/check-module.sh" "${module}" api \
            >/dev/null 2>&1
    else
        "${golib}" api check --module "${module}" >/dev/null 2>&1
    fi
}

record_module_batch() {
    local metric="$1"
    local concurrent="$2"
    local repetition started finished status module pid
    local -a modules=() pids=()
    while IFS= read -r module; do
        modules+=("${module}")
    done < <(jq -r '.modules[].directory' "${source_root}/modules.json")
    for repetition in $(seq 1 3); do
        started="$(date +%s%N)"
        status=0
        if [[ "${concurrent}" == true ]]; then
            pids=()
            for module in "${modules[@]}"; do
                (cd "${source_root}" && TMPDIR="${runtime}" run_api_module "${module}") &
                pids+=("$!")
            done
            for pid in "${pids[@]}"; do
                wait "${pid}" || status=$?
            done
        else
            for module in "${modules[@]}"; do
                (cd "${source_root}" && TMPDIR="${runtime}" run_api_module "${module}") || status=$?
            done
        fi
        finished="$(date +%s%N)"
        [[ "${status}" -eq 0 ]]
        jq -cn \
            --arg implementation "${implementation}" \
            --arg metric "${metric}" \
            --argjson sample "${repetition}" \
            --argjson wall_ns "$((finished - started))" \
            --argjson module_count "${#modules[@]}" \
            '{implementation:$implementation,metric:$metric,sample:$sample,wall_ns:$wall_ns,peak_rss_kib:null,exit_code:0,module_count:$module_count}' \
            >>"${samples}"
    done
}

if [[ "${implementation}" == legacy ]]; then
    startup=("${source_root}/.golib/scripts/run-modules.sh")
    inventory=("${source_root}/.golib/scripts/repository-check.sh")
else
    startup=("${golib}")
    inventory=("${golib}" inventory)
fi
mutation_package_count=null
if [[ "${profile}" == core ]]; then
    mutation_package_count="$(jq '[.packages[] | select(.module_directory == "." and .production == true)] | length' "${source_root}/packages.json")"
    [[ "${mutation_package_count}" -gt 0 ]]
fi
record_command startup-diagnostic 2 25 "${startup[@]}"
record_command repository-inventory 0 10 "${inventory[@]}"

if [[ "${profile}" == core ]]; then
    if [[ "${implementation}" == legacy ]]; then
        checkpoint=(
            "${source_root}/.golib/scripts/with-disposable-go-cache.sh"
            "${source_root}/.golib/scripts/run-modules.sh" mutation --modules .
        )
    else
        checkpoint=("${golib}" mutation --module .)
    fi
    record_command checkpoint-warmup 0 1 "${checkpoint[@]}"
    record_command checkpoint-reuse 0 3 "${checkpoint[@]}"
    record_module_batch module-scaling-sequential false
    record_module_batch module-scaling-concurrent true
else
    if [[ "${implementation}" == legacy ]]; then
        # shellcheck disable=SC2016 # The child shell expands its positional root.
        service_command=(
            bash -c 'set -euo pipefail
root=$1
task=$(mktemp -d "${TMPDIR:-/tmp}/golib-service-cycle.XXXXXX")
cleanup() {
    "$root/.golib/scripts/stop-services.sh" "$task/state" >/dev/null 2>&1 || true
    find "$task" -depth -delete 2>/dev/null || true
}
trap cleanup EXIT HUP INT TERM
"$root/.golib/scripts/start-services.sh" . "$task/environment" "$task/state"
"$root/.golib/scripts/stop-services.sh" "$task/state"
: >"$task/state"' bash "${source_root}"
        )
    else
        service_command=("${golib}" services cycle --module .)
    fi
    record_command service-warmup 0 1 "${service_command[@]}"
    record_command service-lifecycle 0 5 "${service_command[@]}"
fi

if [[ "${implementation}" == legacy ]]; then
    artifact_size_bytes="$(du -sb "${source_root}/.golib" | awk '{print $1}')"
else
    artifact_size_bytes="$(stat -c '%s' "${golib}")"
fi
residue_count="$(find "${runtime}" -mindepth 1 -maxdepth 1 | wc -l | tr -d ' ')"
residue_bytes="$(du -sb "${runtime}" | awk '{print $1}')"

jq -s \
    --arg repository "${repository}" \
    --arg revision "${revision}" \
    --arg content_sha256 "${content_sha256}" \
    --arg tooling_revision "${tooling_revision}" \
    --arg implementation "${implementation}" \
    --arg profile "${profile}" \
    --arg runner_image "${ImageOS:-ubuntu-24.04}" \
    --arg runner_arch "${RUNNER_ARCH:-unknown}" \
    --arg go_version "$(go env GOVERSION)" \
    --arg workflow_repository "${GITHUB_REPOSITORY}" \
    --argjson workflow_run_id "${GITHUB_RUN_ID}" \
    --argjson workflow_run_attempt "${GITHUB_RUN_ATTEMPT}" \
    --arg golib_sha256 "$(sha256sum "${golib}" | awk '{print $1}')" \
    --argjson cpu_count "$(getconf _NPROCESSORS_ONLN)" \
    --argjson artifact_size_bytes "${artifact_size_bytes}" \
    --argjson mutation_package_count "${mutation_package_count}" \
    --argjson residue_count "${residue_count}" \
    --argjson residue_bytes "${residue_bytes}" \
    '{
        schema_version: 1,
        repository: $repository,
        revision: $revision,
        content_sha256: $content_sha256,
        tooling_revision: $tooling_revision,
        implementation: $implementation,
        profile: $profile,
        environment: {runner_image:$runner_image,runner_arch:$runner_arch,go_version:$go_version,cpu_count:$cpu_count},
        workflow: {repository:$workflow_repository,run_id:$workflow_run_id,attempt:$workflow_run_attempt},
        golib_sha256: $golib_sha256,
        artifact_size_bytes: $artifact_size_bytes,
        mutation_package_count: $mutation_package_count,
        isolated_cache_residue: {entries:$residue_count,bytes:$residue_bytes},
        samples: .
    }' "${samples}"
