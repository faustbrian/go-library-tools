#!/usr/bin/env bash
set -euo pipefail

if [[ -z "${BOOTSTRAP_URL}" ]]; then
  [[ -z "${BOOTSTRAP_SHA256}" ]]
  exit 0
fi
[[ "${BOOTSTRAP_URL}" == https://* ]]
[[ "${BOOTSTRAP_SHA256}" =~ ^[0-9a-f]{64}$ ]]
task="$(mktemp -d "${RUNNER_TEMP}/golib-bootstrap-proxy.XXXXXX")"
trap 'chmod -R u+w "${task}" 2>/dev/null || true; find "${task}" -depth -delete' EXIT
archive="${task}/proxy.tar.gz"
curl --fail --silent --show-error --location \
  --retry 5 --retry-delay 2 --retry-all-errors \
  --max-filesize 268435456 \
  "${BOOTSTRAP_URL}" --output "${archive}"
printf '%s  %s\n' "${BOOTSTRAP_SHA256}" "${archive}" | sha256sum --check
if golib --help | grep -Fq 'golib archive validate --file <path>'; then
  golib archive validate --file "${archive}"
else
  mkdir -p "${task}/go-build" "${task}/go-mod" "${task}/go-tmp"
  GOCACHE="${task}/go-build" GOMODCACHE="${task}/go-mod" GOTMPDIR="${task}/go-tmp" GOWORK=off \
    go -C "${GITHUB_ACTION_PATH}/../../.." run ./cmd/golib archive validate --file "${archive}"
fi
proxy="$(mktemp -d "${RUNNER_TEMP}/golib-bootstrap-proxy-data.XXXXXX")"
tar --extract --gzip --no-same-owner --no-same-permissions \
  --file "${archive}" --directory "${proxy}"
echo "GOPROXY=https://proxy.golang.org,file://${proxy},direct" >>"${GITHUB_ENV}"
echo 'GONOSUMDB=github.com/faustbrian/go-*' >>"${GITHUB_ENV}"
