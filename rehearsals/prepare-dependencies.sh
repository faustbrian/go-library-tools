#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 2 ]]; then
    printf 'usage: %s <task-root> <environment-output>\n' "$0" >&2
    exit 2
fi

task_root="$1"
environment_output="$2"
source_root="$(pwd -P)"
real_go="$(command -v go)"
module_map="${task_root}/modules.tsv"
wrapper_directory="${task_root}/bin"
root_alternate_mod=''

mkdir -p "${task_root}/cache" "${task_root}/mod" "${task_root}/tmp" \
    "${task_root}/modules" "${wrapper_directory}"
: >"${module_map}"

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
    for dependency in "${dependencies[@]}"; do
        GOWORK=off GOFLAGS="-modfile=${alternate_mod}" \
            GOCACHE="${task_root}/cache" GOMODCACHE="${task_root}/mod" \
            GOTMPDIR="${task_root}/tmp" "${real_go}" mod download "${dependency}"
    done
    printf '%s\t%s\n' "${module_path}" "${alternate_mod}" >>"${module_map}"
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
directory="$(pwd -P)"
alternate_mod=''

while true; do
    if [[ -f "${directory}/go.mod" ]]; then
        module_path="$(awk '$1 == "module" { print $2; exit }' "${directory}/go.mod")"
        alternate_mod="$(awk -F '\t' -v module="${module_path}" '$1 == module { print $2; exit }' "${module_map}")"
        [[ -z "${alternate_mod}" ]] || break
    fi
    [[ "${directory}" == / ]] && break
    directory="${directory%/*}"
    [[ -n "${directory}" ]] || directory=/
done

explicit_modfile=false
for argument in "$@"; do
    [[ "${argument}" != -modfile=* ]] || explicit_modfile=true
done

active_modfile=''
for flag in ${GOFLAGS:-}; do
    case "${flag}" in
        -modfile=*) active_modfile="${flag#-modfile=}" ;;
    esac
done

if [[ "${explicit_modfile}" == false && -n "${alternate_mod}" ]]; then
    if [[ -z "${active_modfile}" ]]; then
        export GOFLAGS="${GOFLAGS:+${GOFLAGS} }-modfile=${alternate_mod}"
    elif [[ "${active_modfile}" != "${alternate_mod}" ]]; then
        if awk -F '\t' -v modfile="${active_modfile}" \
            '$2 == modfile { found = 1 } END { exit !found }' "${module_map}"; then
            updated_flags=''
            for flag in ${GOFLAGS:-}; do
                case "${flag}" in
                    -modfile=*) flag="-modfile=${alternate_mod}" ;;
                esac
                updated_flags="${updated_flags:+${updated_flags} }${flag}"
            done
            export GOFLAGS="${updated_flags}"
        else
            active_sum="${active_modfile%.mod}.sum"
            alternate_sum="${alternate_mod%.mod}.sum"
            temporary_sum="${active_sum}.golib.$$"
            {
                if [[ -f "${active_sum}" ]]; then
                    awk '$1 !~ /^github\.com\/faustbrian\/go-/' "${active_sum}"
                fi
                awk '$1 ~ /^github\.com\/faustbrian\/go-/' "${alternate_sum}"
            } >"${temporary_sum}"
            mv "${temporary_sum}" "${active_sum}"
        fi
    fi
elif [[ "${explicit_modfile}" == false && -n "${active_modfile}" ]] &&
    awk -F '\t' -v modfile="${active_modfile}" \
        '$2 == modfile { found = 1 } END { exit !found }' "${module_map}"; then
    updated_flags=''
    for flag in ${GOFLAGS:-}; do
        case "${flag}" in
            -modfile=*) continue ;;
        esac
        updated_flags="${updated_flags:+${updated_flags} }${flag}"
    done
    export GOFLAGS="${updated_flags}"
fi

exec "${real_go}" "$@"
EOF
chmod 0755 "${wrapper_directory}/go"

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
    printf 'GOFLAGS=%s%s-modfile=%s\n' \
        "${base_flags}" "${base_flags:+ }" "${root_alternate_mod}"
    printf 'PATH=%s:%s\n' "${wrapper_directory}" "${PATH}"
} >>"${environment_output}"
