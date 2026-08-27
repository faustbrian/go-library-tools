#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 3 ]]; then
    printf 'usage: %s <task-root> <environment-output> <path-output>\n' "$0" >&2
    exit 2
fi

task_root="$1"
environment_output="$2"
path_output="$3"
source_root="$(pwd -P)"
real_go="$(command -v go)"
module_map="${task_root}/modules.tsv"
wrapper_directory="${task_root}/bin"
source_archive="${task_root}/source.tar"
root_alternate_mod=''

mkdir -p "${task_root}/cache" "${task_root}/mod" "${task_root}/tmp" \
    "${task_root}/modules" "${task_root}/execution" "${wrapper_directory}"
: >"${module_map}"
git archive --format=tar HEAD >"${source_archive}"

module_directories=()
while IFS= read -r directory; do
    module_directories+=("${directory}")
done < <(jq -r '.modules[].directory' modules.json)
for index in "${!module_directories[@]}"; do
    directory="${module_directories[${index}]}"
    module_root="${source_root}"
    if [[ "${directory}" != "." ]]; then
        module_root="${source_root}/${directory}"
    fi
    if [[ ! -f "${module_root}/go.mod" ]]; then
        printf 'module %s has no go.mod\n' "${directory}" >&2
        exit 1
    fi

    alternate_mod="${task_root}/modules/${index}.mod"
    alternate_sum="${task_root}/modules/${index}.sum"
    cp "${module_root}/go.mod" "${alternate_mod}"
    if [[ -f "${module_root}/go.sum" ]]; then
        awk '$1 !~ /^github\.com\/faustbrian\/go-/' \
            "${module_root}/go.sum" >"${alternate_sum}"
    else
        : >"${alternate_sum}"
    fi

    module_metadata="$(
        GOWORK=off GOCACHE="${task_root}/cache" GOMODCACHE="${task_root}/mod" \
            GOTMPDIR="${task_root}/tmp" "${real_go}" mod edit \
            -json -modfile="${alternate_mod}"
    )"
    module_path="$(jq -er '.Module.Path' <<<"${module_metadata}")"
    dependencies=()
    while IFS= read -r dependency; do
        dependencies+=("${dependency}")
    done < <(
        jq -r '
            .Require[]? |
            select(.Version != null and (.Path | startswith("github.com/faustbrian/go-"))) |
            "\(.Path)@\(.Version)"
        ' <<<"${module_metadata}"
    )
    if [[ "${#dependencies[@]}" -gt 0 ]]; then
        for dependency in "${dependencies[@]}"; do
            GOWORK=off GOFLAGS="-modfile=${alternate_mod}" \
                GOCACHE="${task_root}/cache" GOMODCACHE="${task_root}/mod" \
                GOTMPDIR="${task_root}/tmp" "${real_go}" mod download "${dependency}"
        done
    fi
    printf '%s\t%s\t%s\n' "${module_path}" "${alternate_mod}" "${directory}" >>"${module_map}"
    if [[ "${directory}" == "." ]]; then
        root_alternate_mod="${alternate_mod}"
    fi
done

if [[ -z "${root_alternate_mod}" ]]; then
    printf 'module inventory has no root module\n' >&2
    exit 1
fi

cat >"${wrapper_directory}/go" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

real_go="${GOLIB_REHEARSAL_REAL_GO:?}"
module_map="${GOLIB_REHEARSAL_MODULE_MAP:?}"
execution_directory="${GOLIB_REHEARSAL_EXECUTION_DIRECTORY:?}"
directory="$(pwd -P)"
alternate_mod=''

while true; do
    if [[ -f "${directory}/go.mod" ]]; then
        module_path="$(awk '$1 == "module" { print $2; exit }' "${directory}/go.mod")"
        alternate_mod="$(awk -F '\t' -v module="${module_path}" '$1 == module { print $2; exit }' "${module_map}")"
        break
    fi
    [[ "${directory}" == / ]] && break
    directory="${directory%/*}"
    [[ -n "${directory}" ]] || directory=/
done

command_arguments=("$@")
versioned_tool=0
if [[ "${command_arguments[0]:-}" == run ]]; then
    for argument in "${command_arguments[@]:1}"; do
        if [[ "${argument}" == *@* ]]; then
            versioned_tool=1
            break
        fi
    done
fi
explicit_modfile=''
for argument in "${command_arguments[@]}"; do
    case "${argument}" in
        -modfile=*) explicit_modfile="${argument#-modfile=}" ;;
    esac
done

active_modfile="${explicit_modfile}"
if [[ -z "${active_modfile}" ]]; then
    for flag in ${GOFLAGS:-}; do
        case "${flag}" in
            -modfile=*) active_modfile="${flag#-modfile=}" ;;
        esac
    done
fi

execution_modfile=''
execution_sum=''
tool_source=''
local_dependencies=()
cleanup_execution_modfile() {
    for path in "${execution_modfile}" "${execution_sum}"; do
        [[ -n "${path}" ]] || continue
        if [[ -e "${path}" || -L "${path}" ]]; then
            find "${path}" -delete
        fi
    done
    execution_modfile=''
    execution_sum=''
    if [[ -n "${tool_source}" && -d "${tool_source}" ]]; then
        find "${tool_source}" -depth -delete
    fi
    tool_source=''
}
handle_signal() {
    local signal="$1"
    trap - EXIT HUP INT TERM
    cleanup_execution_modfile
    exit "$((128 + signal))"
}
if [[ -n "${alternate_mod}" ]]; then
    if [[ -z "${active_modfile}" ]]; then
        export GOFLAGS="${GOFLAGS:+${GOFLAGS} }-modfile=${alternate_mod}"
    elif [[ "${active_modfile}" != "${alternate_mod}" ]]; then
        if awk -F '\t' -v modfile="${active_modfile}" \
            '$2 == modfile { found = 1 } END { exit !found }' "${module_map}"; then
            if [[ -n "${explicit_modfile}" ]]; then
                for index in "${!command_arguments[@]}"; do
                    case "${command_arguments[${index}]}" in
                        -modfile=*) command_arguments[${index}]="-modfile=${alternate_mod}" ;;
                    esac
                done
            else
                updated_flags=''
                for flag in ${GOFLAGS:-}; do
                    case "${flag}" in
                        -modfile=*) flag="-modfile=${alternate_mod}" ;;
                    esac
                    updated_flags="${updated_flags:+${updated_flags} }${flag}"
                done
                export GOFLAGS="${updated_flags}"
            fi
        else
            active_sum="${active_modfile%.mod}.sum"
            alternate_sum="${alternate_mod%.mod}.sum"
            execution_modfile="${execution_directory}/$(basename "${active_modfile%.mod}")-$$.mod"
            execution_sum="${execution_modfile%.mod}.sum"
            trap cleanup_execution_modfile EXIT
            trap 'handle_signal 1' HUP
            trap 'handle_signal 2' INT
            trap 'handle_signal 15' TERM
            cp "${active_modfile}" "${execution_modfile}"
            while read -r module version _; do
                [[ "${version}" != */go.mod ]] || continue
                if [[ -n "${GOLIB_LOCAL_PROXY:-}" &&
                    -f "${GOLIB_LOCAL_PROXY}/${module}/@v/${version}.zip" ]]; then
                    local_dependencies+=("${module}@${version}")
                fi
            done < <(
                awk '$1 ~ /^github\.com\/faustbrian\/go-/' \
                    "${alternate_sum}"
            )
            {
                if [[ -f "${active_sum}" ]]; then
                    awk '$1 !~ /^github\.com\/faustbrian\/go-/' "${active_sum}"
                fi
                while IFS= read -r checksum; do
                    module="${checksum%% *}"
                    version_and_sum="${checksum#* }"
                    version="${version_and_sum%% *}"
                    archive_version="${version%/go.mod}"
                    if [[ -n "${GOLIB_LOCAL_PROXY:-}" &&
                        -f "${GOLIB_LOCAL_PROXY}/${module}/@v/${archive_version}.zip" ]]; then
                        continue
                    fi
                    printf '%s\n' "${checksum}"
                done < <(
                    awk '$1 ~ /^github\.com\/faustbrian\/go-/' \
                        "${alternate_sum}"
                )
            } | LC_ALL=C sort -u >"${execution_sum}"
            for dependency in "${local_dependencies[@]}"; do
                GOWORK=off GOFLAGS="-modfile=${execution_modfile}" \
                    "${real_go}" mod download "${dependency}"
            done
            if [[ -n "${explicit_modfile}" ]]; then
                for index in "${!command_arguments[@]}"; do
                    case "${command_arguments[${index}]}" in
                        -modfile=*) command_arguments[${index}]="-modfile=${execution_modfile}" ;;
                    esac
                done
            else
                updated_flags=''
                for flag in ${GOFLAGS:-}; do
                    case "${flag}" in
                        -modfile=*) flag="-modfile=${execution_modfile}" ;;
                    esac
                    updated_flags="${updated_flags:+${updated_flags} }${flag}"
                done
                export GOFLAGS="${updated_flags}"
            fi
        fi
    fi
elif [[ -n "${active_modfile}" ]] &&
    awk -F '\t' -v modfile="${active_modfile}" \
        '$2 == modfile { found = 1 } END { exit !found }' "${module_map}"; then
    if [[ -n "${explicit_modfile}" ]]; then
        filtered_arguments=()
        for argument in "${command_arguments[@]}"; do
            case "${argument}" in
                -modfile=*) continue ;;
            esac
            filtered_arguments+=("${argument}")
        done
        command_arguments=("${filtered_arguments[@]}")
    else
        updated_flags=''
        for flag in ${GOFLAGS:-}; do
            case "${flag}" in
                -modfile=*) continue ;;
            esac
            updated_flags="${updated_flags:+${updated_flags} }${flag}"
        done
        export GOFLAGS="${updated_flags}"
    fi
fi

if [[ "${versioned_tool}" -eq 1 ]]; then
    effective_modfile=''
    for flag in ${GOFLAGS:-}; do
        case "${flag}" in
            -modfile=*) effective_modfile="${flag#-modfile=}" ;;
        esac
    done
    module_directory="$(awk -F '\t' -v module="${module_path}" '$1 == module { print $3; exit }' "${module_map}")"
    [[ -n "${effective_modfile}" && -n "${module_directory}" ]]
    tool_source="${execution_directory}/source-$$"
    trap cleanup_execution_modfile EXIT
    trap 'handle_signal 1' HUP
    trap 'handle_signal 2' INT
    trap 'handle_signal 15' TERM
    mkdir -p "${tool_source}"
    tar -xf "${GOLIB_REHEARSAL_SOURCE_ARCHIVE:?}" -C "${tool_source}"
    tool_directory="${tool_source}"
    if [[ "${module_directory}" != . ]]; then
        tool_directory="${tool_source}/${module_directory}"
    fi
    cp "${effective_modfile}" "${tool_directory}/go.mod"
    effective_sum="${effective_modfile%.mod}.sum"
    : >"${tool_directory}/go.sum"
    if [[ -f "${effective_sum}" ]]; then
        cp "${effective_sum}" "${tool_directory}/go.sum"
    fi
    export GOLIB_REHEARSAL_CONSUMER_GOFLAGS=''
    export GOLIB_REHEARSAL_TOOL_DIRECTORY="${tool_directory}"
    export GOFLAGS=''
    command_arguments=(
        run "-exec=${GOLIB_REHEARSAL_TOOL_RUNNER:?}"
        "${command_arguments[@]:1}"
    )
fi

if [[ -n "${execution_modfile}" || -n "${tool_source}" ]]; then
    status=0
    "${real_go}" "${command_arguments[@]}" || status=$?
    cleanup_execution_modfile
    trap - EXIT HUP INT TERM
    exit "${status}"
fi

exec "${real_go}" "${command_arguments[@]}"
EOF
cat >"${wrapper_directory}/run-versioned-tool" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

export GOFLAGS="${GOLIB_REHEARSAL_CONSUMER_GOFLAGS:-}"
export GO111MODULE=on
export GOWORK=off
unset GOLIB_REHEARSAL_CONSUMER_GOFLAGS
cd "${GOLIB_REHEARSAL_TOOL_DIRECTORY:?}"
exec "$@"
EOF
chmod 0755 "${wrapper_directory}/go"
chmod 0755 "${wrapper_directory}/run-versioned-tool"

base_flags=''
for flag in ${GOFLAGS:-}; do
    case "${flag}" in
        -modfile=*) ;;
        *) base_flags="${base_flags:+${base_flags} }${flag}" ;;
    esac
done
{
    printf 'GOLIB_REHEARSAL_REAL_GO=%s\n' "${real_go}"
    printf 'GOLIB_REHEARSAL_MODULE_MAP=%s\n' "${module_map}"
    printf 'GOLIB_REHEARSAL_EXECUTION_DIRECTORY=%s\n' "${task_root}/execution"
    printf 'GOLIB_REHEARSAL_TOOL_RUNNER=%s\n' \
        "${wrapper_directory}/run-versioned-tool"
    printf 'GOLIB_REHEARSAL_SOURCE_ARCHIVE=%s\n' "${source_archive}"
    printf 'GOLIB_REAL_GO=%s\n' "${wrapper_directory}/go"
    printf 'GOFLAGS=%s%s-modfile=%s\n' \
        "${base_flags}" "${base_flags:+ }" "${root_alternate_mod}"
    printf 'PATH=%s:%s\n' "${wrapper_directory}" "${PATH}"
} >>"${environment_output}"
printf '%s\n' "${wrapper_directory}" >>"${path_output}"
