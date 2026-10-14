#!/usr/bin/env bash
set -euo pipefail
for package in . caphttp memory postgres valkey; do
    checkpoint="${package}"
    [[ "${package}" != . ]] || checkpoint=root
    actual="$(bash .golib/scripts/gate-input-digest.sh mutation . "${package}")"
    stored="$(jq -r '.gate_input_digest' ".artifacts/mutation-checkpoints/${checkpoint}.json")"
    [[ "${actual}" == "${stored}" ]]
    printf '%s current complete mutation inputs match executed checkpoint: %s\n' "${package}" "${actual}"
done
