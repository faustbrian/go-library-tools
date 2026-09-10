# Reusable Workflows

Consumer CI calls `library-ci.yml` at an immutable commit SHA. The same SHA is
passed as `tooling_sha`, allowing the workflow to check out its setup action
without a mutable reference. A nearby comment records the corresponding
tooling release.

```yaml
name: CI

on:
  pull_request:
  push:
    branches: [main]
  schedule:
    - cron: '17 3 * * *'
  workflow_dispatch:
    inputs:
      release_dry_run:
        type: boolean
        default: false
      release_module:
        type: string
        default: ''

permissions:
  contents: read
  security-events: write

jobs:
  ci:
    uses: faustbrian/go-library-tools/.github/workflows/library-ci.yml@0123456789abcdef0123456789abcdef01234567 # v1.0.0
    with:
      tooling_sha: 0123456789abcdef0123456789abcdef01234567
      release_dry_run: ${{ inputs.release_dry_run || false }}
      release_module: ${{ inputs.release_module || '' }}

  required:
    name: Required
    if: always()
    needs: ci
    runs-on: ubuntu-24.04
    timeout-minutes: 5
    steps:
      - name: Require shared workflow
        env:
          CI_RESULT: ${{ needs.ci.result }}
        run: test "${CI_RESULT}" = success
```

The setup action reads the exact `tool_version` from `.golib.yaml`, selects the
Linux or macOS amd64/arm64 archive, downloads it with GitHub CLI, verifies its
SHA-256 checksum and GitHub artifact attestation, and only then extracts the
binary. It never evaluates a downloaded installer.
The workflow detects whether that installed version supports the bounded local
check mode and otherwise uses the older module-check form, so updating an
immutable workflow pin does not force a simultaneous tool-manifest migration.
Release rehearsals similarly omit module selectors when the installed tool
predates module-scoped release commands.

Repositories whose initial dependency graph cannot be reconstructed from the
public Go proxy may define both `GOLIB_BOOTSTRAP_PROXY_URL` and
`GOLIB_BOOTSTRAP_PROXY_SHA256` as repository variables. The URL must identify
an immutable HTTPS archive containing a file-based Go module proxy, and the
checksum must be its lowercase SHA-256 digest. The reusable workflow verifies
the archive before extraction and exposes it to both quality and CodeQL builds.
Defining only one variable, using a mutable or non-HTTPS URL, or providing an
invalid checksum fails closed. These variables bootstrap historical module
identity only; they do not replace normal dependency resolution or permit a
consumer to weaken any gate. Published module versions resolve from the public
Go proxy first. The file proxy is consulted only when the public proxy reports
that a version is unavailable, so bootstrap archives cannot shadow a published
module with different bytes.

The reusable workflow keeps consumer policy in repository manifests. It builds
a module matrix from `golib inventory --json`, runs one isolated module
contract per matrix entry, uploads repository-owned `.verification` evidence,
and runs CodeQL. The caller's final `required` job converts the reusable-call
result into the stable `Required` check used by branch protection.
Set `release_dry_run: true` only for an explicit release rehearsal; this first
validates the stable release contract and then runs the complete release
dry-run for every releasable module. Set `release_module` to an exact module
directory to limit the release matrix and module-scoped release checks to that
independently versioned module when the installed tool supports release
selectors. Older tools safely widen that request to one whole-repository
release check and dry-run rather than repeating the full rehearsal per module.
Whole-repository structure, offline specification validation, and CodeQL checks
still run. Mutable online authority and errata monitoring is separate from
pull-request and release feedback and runs only on the caller's scheduled
monitoring event. A blank selector preserves the all-module rehearsal,
and a non-blank selector without `release_dry_run: true` fails closed. Existing
consumer callers must expose and forward `release_module` before they can
dispatch an exact-module hosted rehearsal; a tooling-pin upgrade alone does not
wire the caller input.

Consumer workflows retain least-privileged permissions, explicit concurrency,
module matrices, attributable evidence artifacts, scheduled checks, CodeQL,
release dry-runs, and one stable final required job.

Pull requests that change only Markdown, `LICENSE`, or `NOTICE` run repository
structural validation without launching module-quality or CodeQL jobs. An
otherwise lightweight pull request may also update only the canonical
reusable-workflow SHA and matching `tooling_sha`; structural validation still
checks that caller, including when the replaced release comment has a legacy
descriptive suffix. Source, configuration, substantive or mismatched workflow
changes, structured metadata, push, scheduled, and explicit release-rehearsal
events retain the complete applicable runtime path.
Pull-request runs review dependency changes before the final required job.
Workflow syntax and expression validation are available locally through
`golib workflows check`, run as part of the repository `make ci` contract, and
execute in the reusable workflow's hosted repository-contract job.

Tool version, binary checksum, setup-action SHA, and reusable-workflow SHA are
updated together through reviewable pull requests. Existing consumers do not
change behavior when a new tooling release is published.

The manually dispatched `Propose consumer upgrades` workflow performs those
updates in bounded cohorts from the validated
[consumer inventory](consumers.md). It defaults to a read-only dry run, limits
each cohort to ten active repositories and five concurrent jobs, and opens one
pull request per changed consumer. Apply mode requires the separately managed
fine-grained `GOLIB_ROLLOUT_TOKEN`; the repository-scoped workflow token cannot
write to sibling repositories. The workflow changes only `.golib.yaml` and the
thin CI caller and never force-pushes an existing rollout branch.

This repository bootstraps its own CI from source so the first release does not
depend on itself. Consumer repositories always use released binaries.

Tooling tags publish only the four platform archives, their SBOMs, the release
manifest, and checksums by default. Catalog and source-lock assets are added
only at a defined ecosystem milestone: refresh and review the source lock, then
set the repository variable `GOLIB_CATALOG_MILESTONE_TAG` to the exact new tag
before pushing it. The exact-tag comparison prevents a stale variable from
expanding later tooling releases. Previously published release assets remain
immutable.
