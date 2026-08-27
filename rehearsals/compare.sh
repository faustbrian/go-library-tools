#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 2 ]]; then
    printf 'usage: %s <artifact-root> <repository-manifest>\n' "$0" >&2
    exit 2
fi

artifacts="$1"
manifest="$2"
while IFS= read -r repository; do
    legacy="${artifacts}/parity-legacy-${repository}/legacy-summary"
    shared="${artifacts}/parity-shared-${repository}/shared-summary"
    [[ -s "${legacy}" && -s "${shared}" ]]
    jq -e --slurp '
      length == 2 and
      .[0].schema_version == 1 and
      .[1].schema_version == 1 and
      .[0].mode == "legacy" and
      .[1].mode == "shared" and
      .[0].content_digest == .[1].content_digest and
      .[0].check_status == 0 and
      .[1].check_status == 0 and
      .[0].service_cleanup_status == 0 and
      .[1].service_cleanup_status == 0 and
      .[0].selected_modules == .[1].selected_modules and
      .[0].effective_gates == .[1].effective_gates and
      .[0].required_services == .[1].required_services and
      .[0].release_units == .[1].release_units and
      .[0].coverage == .[1].coverage and
      .[0].mutation == .[1].mutation and
      .[0].nilaway_advisories == .[1].nilaway_advisories and
      .[0].release_decision == .[1].release_decision
    ' "${legacy}" "${shared}" >/dev/null
    printf '%s: compatible\n' "${repository}"
done < <(jq -r '.repositories[].name' "${manifest}")
