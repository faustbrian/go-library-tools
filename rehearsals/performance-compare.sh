#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 2 ]]; then
    printf 'usage: %s <results-directory> <output>\n' "$0" >&2
    exit 2
fi

results="$1"
output="$2"
reports=()
while IFS= read -r report; do
    reports+=("${report}")
done < <(find "${results}" -type f -name 'performance.json' | LC_ALL=C sort)
[[ "${#reports[@]}" -eq 4 ]]
for report in "${reports[@]}"; do
    cleanup_status="$(dirname "${report}")/performance-services.status"
    [[ -s "${cleanup_status}" ]]
    [[ "$(cat "${cleanup_status}")" == 0 ]]
done

jq -s -e '
  . as $reports |
  (
    ($reports | length == 4) and
    all($reports[];
        .schema_version == 1 and
        (.revision | test("^[0-9a-f]{40}$")) and
        (.content_sha256 | test("^[0-9a-f]{64}$")) and
        (.tooling_revision | test("^[0-9a-f]{40}$")) and
        (.golib_sha256 | test("^[0-9a-f]{64}$")) and
        (.workflow.repository | test("^[^/]+/[^/]+$")) and
        .workflow.run_id > 0 and
        .workflow.attempt > 0 and
        (.implementation == "legacy" or .implementation == "shared") and
        (.profile == "core" or .profile == "service") and
        .artifact_size_bytes > 0 and
        .isolated_cache_residue.entries == 0 and
        .isolated_cache_residue.bytes == 0 and
        all(.samples[];
            .wall_ns > 0 and
            ((.metric | startswith("module-scaling-")) or
              (.peak_rss_kib != null and .peak_rss_kib > 0))) and
        ([.samples[] | select(.metric == "startup-diagnostic")] | length) == 25 and
        all(.samples[] | select(.metric == "startup-diagnostic"); .exit_code == 2) and
        ([.samples[] | select(.metric == "repository-inventory")] | length) == 10 and
        all(.samples[] | select(.metric != "startup-diagnostic"); .exit_code == 0)
    ) and
    ([.[] | .tooling_revision] | unique | length) == 1 and
    ([.[] | .golib_sha256] | unique | length) == 1 and
    ([.[] | [.workflow.repository,.workflow.run_id,.workflow.attempt]] | unique | length) == 1 and
    ([.[] | .profile] | unique | sort) == ["core","service"] and
    ([.[] | [.repository,.revision,.content_sha256,.profile] | join("\u0000")] | unique | length) == 2 and
    all(group_by(.repository + "\u0000" + .profile)[];
        length == 2 and
        ([.[].implementation] | sort) == ["legacy","shared"] and
        ([.[].revision] | unique | length) == 1 and
        ([.[].content_sha256] | unique | length) == 1
    ) and
    all(.[] | select(.profile == "core");
        . as $report |
        .mutation_package_count > 0 and
        ([.samples[] | select(.metric == "checkpoint-warmup")] | length) == 1 and
        ([.samples[] | select(.metric == "checkpoint-reuse")] | length) == 3 and
        all(.samples[] | select(.metric == "checkpoint-reuse");
            .reuse_count == $report.mutation_package_count) and
        ([.samples[] | select(.metric == "module-scaling-sequential")] | length) == 3 and
        ([.samples[] | select(.metric == "module-scaling-concurrent")] | length) == 3 and
        all(.samples[] | select(.metric | startswith("module-scaling-"));
            .module_count > 1)
    ) and
    all(.[] | select(.profile == "service");
        .mutation_package_count == null and
        ([.samples[] | select(.metric == "service-warmup")] | length) == 1 and
        ([.samples[] | select(.metric == "service-lifecycle")] | length) == 5
    )
  ) as $valid |
  if $valid then $reports else error("performance result contract failed") end
' "${reports[@]}" >"${output}"
