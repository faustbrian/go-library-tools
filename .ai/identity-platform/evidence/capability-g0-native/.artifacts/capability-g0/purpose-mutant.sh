#!/usr/bin/env bash
set -euo pipefail
task_source="$(mktemp -d /tmp/capability-g0-mutant.XXXXXX)"
cleanup(){ chmod -R u+w "${task_source}"; find "${task_source}" -depth -delete; }
trap cleanup EXIT HUP INT TERM
cp ./*.go go.mod go.sum "${task_source}/"
cp -R testdata caphttp memory postgres valkey "${task_source}/"
perl -0pi -e "s/(if index != 0 && \(c == '\.' \|\| c == '_' \|\| c == '-'\) \{\n\s*)continue/\${1}break/" "${task_source}/identity_contracts.go"
shasum -a 256 "${task_source}/identity_contracts.go"
cd "${task_source}"
go test . -run '^TestIdentityPurposeExactBounds$' -count=1
