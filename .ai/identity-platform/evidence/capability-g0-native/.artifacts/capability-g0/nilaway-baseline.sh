#!/usr/bin/env bash
set -euo pipefail
fixture="$(mktemp -d /tmp/capability-g0-nilaway-baseline.XXXXXX)"
cleanup() {
    chmod -R u+w "${fixture}"
    find "${fixture}" -depth -delete
}
trap cleanup EXIT HUP INT TERM
git archive b0b21d15c9653decb0ff4a4fbffbe71c77d8d8fc | tar -x -C "${fixture}"
GOWORK=off go -C "${fixture}" run go.uber.org/nilaway/cmd/nilaway@v0.0.0-20260720194628-9fd1b8d7bac8 -include-pkgs=github.com/faustbrian/go-capability ./...
