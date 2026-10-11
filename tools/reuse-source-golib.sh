#!/usr/bin/env bash
set -euo pipefail

case "${1:-}" in
  prepare)
    [[ "${TOOLING_SHA}" =~ ^[0-9a-f]{40}$ && "${BINARY_SHA256}" =~ ^[0-9a-f]{64}$ && "${ARTIFACT_ID}" =~ ^[1-9][0-9]{0,19}$ ]] || { echo 'invalid same-run source identity' >&2; exit 1; }
    [[ "${RUNNER_OS}/${RUNNER_ARCH}" == Linux/X64 ]] || { echo 'source artifact requires the prepare platform' >&2; exit 1; }
    source_root="${GITHUB_WORKSPACE}/.golib-tooling"
    [[ ! -L "${source_root}" && "$(git -C "${source_root}" rev-parse --show-toplevel)" == "${source_root}" && "$(git -C "${source_root}" rev-parse HEAD)" == "${TOOLING_SHA}" ]] || { echo 'tooling checkout identity mismatch' >&2; exit 1; }
    case "$(git -C "${source_root}" remote get-url origin)" in
      https://github.com/faustbrian/go-library-tools|https://github.com/faustbrian/go-library-tools.git) ;;
      *) echo 'source artifact requires the fixed tooling repository' >&2; exit 1 ;;
    esac
    install_root="$(mktemp -d "${RUNNER_TEMP}/golib-source-install.XXXXXX")"
    echo "GOLIB_SOURCE_INSTALL_ROOT=${install_root}" >>"${GITHUB_ENV}"
    mkdir -m 700 "${install_root}/bin"
    ;;
  install)
    root="${GOLIB_SOURCE_INSTALL_ROOT}"
    case "${root}" in
      "${RUNNER_TEMP}"/golib-source-install.*/*) echo 'unexpected source installation' >&2; exit 1 ;;
      "${RUNNER_TEMP}"/golib-source-install.*) ;;
      *) echo 'unexpected source installation' >&2; exit 1 ;;
    esac
    [[ ! -L "${root}" && -d "${root}/bin" && ! -L "${root}/bin" && -f "${root}/bin/golib" && ! -L "${root}/bin/golib" ]] || { echo 'unexpected source artifact payload' >&2; exit 1; }
    [[ "$(find "${root}/bin" -mindepth 1 -maxdepth 1 | wc -l | tr -d ' ')" == 1 ]] || { echo 'unexpected source artifact entries' >&2; exit 1; }
    size="$(wc -c <"${root}/bin/golib" | tr -d ' ')"
    [[ "${size}" -gt 0 && "${size}" -le 67108864 ]] || { echo 'source artifact exceeds binary budget' >&2; exit 1; }
    [[ "$(shasum -a 256 "${root}/bin/golib" | cut -d ' ' -f 1)" == "${BINARY_SHA256}" ]] || { echo 'source artifact checksum mismatch' >&2; exit 1; }
    chmod 700 "${root}/bin/golib"
    test "$("${root}/bin/golib" --version)" = dev
    mkdir -m 700 "${root}/lint" "${root}/xdg" "${root}/npm" "${root}/tmp"
    {
      echo "GOLANGCI_LINT_CACHE=${root}/lint"
      echo "XDG_CACHE_HOME=${root}/xdg"
      echo "NPM_CONFIG_CACHE=${root}/npm"
      echo "TMPDIR=${root}/tmp"
      echo 'GOENV=off'
      echo 'GOWORK=off'
      echo 'GOTOOLCHAIN=local'
      echo 'GOMAXPROCS=2'
      echo 'GOFLAGS=-p=1'
    } >>"${GITHUB_ENV}"
    echo "${root}/bin" >>"${GITHUB_PATH}"
    ;;
  *) echo 'source installation requires prepare or install' >&2; exit 1 ;;
esac
