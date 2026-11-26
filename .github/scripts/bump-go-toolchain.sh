#!/usr/bin/env bash
# Move a repository's pinned Go toolchain to the newest stable patch of its
# current minor release, or to an explicit patch of that minor release.
#
# Only the repository toolchain pin changes: .go-version, the top-level
# modules.json go_version and a go.work go directive equal to the old pin.
# go.mod language minimums and workflow files never change; workflows that
# pin a Go version deliberately (fixed benchmark refs) keep their own pin, and
# the Actions token cannot push workflow edits.
# The new version is printed when files changed; nothing is printed when the
# repository is already current.
set -euo pipefail

root="${1:?usage: bump-go-toolchain.sh <repository-root> [version]}"
target="${2:-}"
manifest_url='https://raw.githubusercontent.com/actions/go-versions/main/versions-manifest.json'
cd "${root}"

current="$(tr -d '[:space:]' <.go-version)"
if [[ ! "${current}" =~ ^([0-9]+)\.([0-9]+)\.([0-9]+)$ ]]; then
  echo "unsupported .go-version: ${current}" >&2
  exit 1
fi
minor="${BASH_REMATCH[1]}.${BASH_REMATCH[2]}"
current_patch="${BASH_REMATCH[3]}"

# setup-go installs from this manifest, so a patch is selected only once the
# runner can actually download it.
if [[ -z "${target}" ]]; then
  target="$(curl -fsSL --retry 3 --max-time 60 "${manifest_url}" | jq -er --arg minor "${minor}" '
    [.[]
      | select(.stable == true)
      | .version as $version
      | select($version | test("^[0-9]+\\.[0-9]+\\.[0-9]+$"))
      | select(($version | split(".") | .[0:2] | join(".")) == $minor)
      | select(any(.files[]; .platform == "linux" and .arch == "x64"))
      | $version]
    | if length == 0 then error("no stable Go patch for " + $minor)
      else max_by(split(".") | map(tonumber)) end')"
fi
if [[ ! "${target}" =~ ^([0-9]+)\.([0-9]+)\.([0-9]+)$ || "${BASH_REMATCH[1]}.${BASH_REMATCH[2]}" != "${minor}" ]]; then
  echo "target ${target} is not a Go ${minor} patch" >&2
  exit 1
fi
if (( BASH_REMATCH[3] <= current_patch )); then
  exit 0
fi

export OLD_PATTERN="${current//./\\.}" NEW_VERSION="${target}"

printf '%s\n' "${target}" >.go-version

if [[ -f modules.json ]]; then
  if [[ "$(jq -er '.go_version' modules.json)" != "${current}" ]]; then
    echo "modules.json go_version differs from .go-version ${current}" >&2
    exit 1
  fi
  perl -0pi -e 's/^(  "go_version": ")$ENV{OLD_PATTERN}(",?)$/${1}$ENV{NEW_VERSION}${2}/m
    or die "modules.json top-level go_version line not found\n"' modules.json
  if [[ "$(jq -er '.go_version' modules.json)" != "${target}" ]]; then
    echo 'modules.json go_version was not updated' >&2
    exit 1
  fi
fi

if [[ -f go.work ]]; then
  perl -pi -e 's/^go $ENV{OLD_PATTERN}$/go $ENV{NEW_VERSION}/' go.work
fi

printf '%s\n' "${target}"
