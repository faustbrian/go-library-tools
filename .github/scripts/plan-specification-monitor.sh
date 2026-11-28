#!/usr/bin/env bash
set -euo pipefail

# The assurance repository is a composition harness, not a module owner.
repositories="$(
  gh api --paginate --slurp "users/${OWNER}/repos?type=owner&per_page=100" |
    jq -cer '
      [.[][] | select(.name | test("^go-[a-z0-9]+(-[a-z0-9]+)*$")) |
       select(.archived == false and .fork == false and .private == false) |
       select(.name != "go-library-assurance")] |
      if length == 0 or length > 256 then error("unexpected public library count") else . end |
      if any(.[]; .default_branch != "main") then error("public libraries must use main") else . end |
      sort_by(.name)[] | .name
    '
)"

selected='[]'
while IFS= read -r name; do
  repository="${OWNER}/${name}"
  revision="$(gh api "repos/${repository}/git/ref/heads/main" --jq .object.sha)"
  [[ "${revision}" =~ ^[0-9a-f]{40}$ ]] || { echo "Invalid main revision for ${name}" >&2; exit 1; }
  declared="$(
    gh api -H 'Accept: application/vnd.github.raw+json' "repos/${repository}/contents/modules.json?ref=${revision}" |
      jq -c '
        if (.modules | type) != "array" then error("missing module array") else .modules end |
        if any(.[]; .specifications != null and (.specifications | type) != "array")
        then error("invalid specification declaration")
        else any(.[]; (.specifications // [] | length) > 0)
        end
      '
  )" || { echo "Cannot read module declarations for ${name}" >&2; exit 1; }
  case "${declared}" in
    true)
      selected="$(jq -c --arg name "${name}" --arg revision "${revision}" '. + [{name: $name, revision: $revision}]' <<<"${selected}")"
      ;;
    false) ;;
    *) echo "Missing module manifest for ${name}" >&2; exit 1 ;;
  esac
done <<<"${repositories}"

[[ "$(jq length <<<"${selected}")" -gt 0 ]] || { echo 'No specification owners found' >&2; exit 1; }
printf 'repositories=%s\n' "${selected}" >>"${GITHUB_OUTPUT}"
