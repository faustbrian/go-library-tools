#!/usr/bin/env bash
set -euo pipefail
current_log="${1:-.artifacts/capability-g0/full-final-inputs-2.log}"
diff -u \
    <(awk '/nilaway/{record=1;next} /NilAway advisory exit status/{record=0} record && !/^go: downloading/ {print}' "${current_log}" | sed 's@/Users/brian/Developer/.worktrees/golib-go-capability-identity-platform-worker-g0/@ROOT/@g; s@golib-go-capability-identity-platform-worker-g0/@ROOT/@g') \
    <(sed '/^go: downloading/d; /^exit status 3$/d' .artifacts/capability-g0/nilaway-baseline.log | sed -E 's@/private/tmp/capability-g0-nilaway-baseline\.[^/]+/@ROOT/@g; s@capability-g0-nilaway-baseline\.[^/]+/@ROOT/@g')
printf 'Pinned analyzer diagnostics match pre-change baseline after root-directory normalization.\n'
