#!/usr/bin/env bash
set -euo pipefail
fixture="$(mktemp -d /tmp/capability-g0-fingerprint.XXXXXX)"
cleanup(){ chmod -R u+w "${fixture}"; find "${fixture}" -depth -delete; }
trap cleanup EXIT HUP INT TERM
git ls-files -co --exclude-standard | tar -T - -cf - | tar -xf - -C "${fixture}"
git -C "${fixture}" init -q
cd "${fixture}"
digest(){ bash .golib/scripts/gate-input-digest.sh mutation . "$1"; }
root_before="$(digest .)"
memory_before="$(digest memory)"
printf 'baseline root=%s memory=%s\n' "${root_before}" "${memory_before}"
perl -pi -e 's/const tokenPrefix = "cap1"/const tokenPrefix = "cap1" \/\/ source discriminator/' token.go
root_source="$(digest .)"
memory_source="$(digest memory)"
[[ "${root_source}" != "${root_before}" && "${memory_source}" != "${memory_before}" ]]
printf 'root source changes root and dependent memory digests\n'
perl -pi -e 's/ \/\/ source discriminator//' token.go
perl -pi -e 's/^package capability_test$/package capability_test \/\/ test discriminator/' identity_contracts_test.go
[[ "$(digest .)" != "${root_before}" && "$(digest memory)" == "${memory_before}" ]]
printf 'root observer test changes root, not unrelated memory observer digest\n'
perl -pi -e 's/ \/\/ test discriminator//' identity_contracts_test.go
perl -pi -e 's/^package memory$/package memory \/\/ descendant discriminator/' memory/consumption.go
[[ "$(digest memory)" != "${memory_before}" && "$(digest .)" != "${root_before}" ]]
printf 'consumed descendant source changes memory and root observer digests\n'
perl -pi -e 's/ \/\/ descendant discriminator//' memory/consumption.go
perl -pi -e 's/^package valkey$/package valkey \/\/ unrelated discriminator/' valkey/store.go
[[ "$(digest .)" == "${root_before}" && "$(digest memory)" == "${memory_before}" ]]
printf 'unrelated adapter source leaves root and memory digests unchanged\n'
