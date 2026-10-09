# Reusable Workflows

## Tools-only native CLI diagnostic

Tools' own `ci.yml` accepts the optional manual input `native_cli_diagnostic`.
It selects one retained `internal/cli/cli.go:138:25` `INVERT_BITWISE` mutation
through the pinned engine's normal AST discovery and executor. The diagnostic
uses the existing shared-coverage baseline and phase-budget calculation,
preserves native command/cancellation behavior, observes bounded scalar phase
results, and checks exact source rollback in a disposable copy. It runs only
on hosted runners; ordinary tests skip its opt-in subprocess path.

This is an investigation result, not a campaign, historical-cache reproduction,
equivalence waiver or release pass. A timeout remains a timeout. The manually
selected job must succeed for that workflow's Required check to succeed; no
consumer workflow or production verifier asset is changed by this diagnostic.

## Consumer workflows

Consumer CI calls `library-ci.yml` at an immutable commit SHA. The same SHA is
passed as `tooling_sha` for the isolated tooling checkout. By default the setup action
is pinned to `e8c0c51cc517dfede5475858bb417f0e7a21f39d`; the bootstrap
action is pinned to `712163686bea3c3ad465acdca026b4b7bf5f0b0b`.
By default, `tooling_sha` does not select those executable actions. A nearby
comment records the corresponding tooling release.

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
check mode and otherwise uses the older all-module check form, so updating an
immutable workflow pin does not force a simultaneous tool-manifest migration.
Release rehearsals similarly omit module selectors when the installed tool
predates module-scoped release commands.

The setup action also offers an explicit development-source option for a
reviewed prerequisite. An individually pinned action call may set
`source_bootstrap: 'true'` and supply `tooling_sha`; it must first check out
`faustbrian/go-library-tools` at that exact lowercase 40-character commit into
the fixed `.golib-tooling` directory. No arbitrary repository or source path
is accepted. The action verifies the checkout identity and requires a pristine
checkout, including tracked, untracked and ignored files. Assume-unchanged and
skip-worktree index entries are refused because they can conceal changed inputs.
Git inspection failures also refuse the build without rendering file paths.
The caller owns a stable checkout during validation and building; this is not
an atomic snapshot against concurrent filesystem changes. The action uses
read-only module resolution and private public-proxy/SumDB Go caches, and
requires the unmodified `dev` version.
It does not represent a checksum-verified or attested release, and
`.golib.yaml` keeps its existing public declarations unchanged.

Build caches are removed on success and failure. On success the installed
binary and ancillary caches remain job-owned at `GOLIB_SOURCE_INSTALL_ROOT`.
The caller must remove that exact directory in an always-cleanup step, after
checking it lies beneath `RUNNER_TEMP` with the `golib-source-install.` prefix.
The development action does not select or waive any consumer gate. The action's
default false route retains the existing release checksum, attestation and
exact-version verification.

Reusable callers may explicitly set `source_bootstrap: true`, default false.
All four setup sites then use the immutable remote setup action at
`e8c0c51cc517dfede5475858bb417f0e7a21f39d`, passing the same `tooling_sha`
used for their fixed tooling checkout. Documentation/license-only pull requests
use the ordinary metadata assurance tier, including in source-bootstrap mode.
Metadata-only source-bootstrap jobs still scan repository history and the
current tree through `golib secrets check`, using the existing bounded scan
policy. Older source tooling without this command fails closed; adopt the
workflow and supporting tooling source together.
Changed source-bootstrap tooling pins still require runtime assurance, even
when the reusable workflow and tooling pins match. Runtime-selected source
jobs run full ordinary `golib check --all`, never the help-detected local
shortcut. Pushes, schedules, manual runs and release rehearsals retain runtime
assurance. Every source-mode
job removes the exact owned installation and ancillary caches in an
always-cleanup step, including when later gates fail. Setup failures clean
build resources inside the action; cleanup tolerates an installation that was
never created or was already removed on failure.

Cleanup restores `TMPDIR` to the hosted job's stable `RUNNER_TEMP` before
removing the installation. Post-job actions can therefore resolve their
temporary directory after every private source-bootstrap cache and binary
has been deleted. This teardown does not preserve a pre-bootstrap custom
temporary-directory selection.

Default release installation and routing remain unchanged. Repository,
specification, dependency review, CodeQL, artifact retention, release selectors
and Required remain in force. Source bootstrap supports reviewed development
prerequisites; it is not a released-tool attestation or a consumer release pass.

Callers whose external dependency graph is publicly available may independently
set `public_dependencies: true`, default false. All four reusable jobs then
persist `GOPROXY=https://proxy.golang.org`, `GOSUMDB=sum.golang.org`, and empty
`GOPRIVATE`, `GONOPROXY` and `GONOSUMDB` before dependency-consuming gates.
Quality and CodeQL skip legacy bootstrap archive restoration in this mode.
Unavailable public dependencies fail without file-proxy or direct fallback;
checksum verification remains enabled for owned modules. Gate-owned
same-repository composition retains its separate source-qualified boundary.
This input does not select development tooling or change any required gate.
Omitting it preserves historical bootstrap routing, regardless of
`source_bootstrap` selection.

Repositories whose initial dependency graph cannot be reconstructed from the
public Go proxy may define both `GOLIB_BOOTSTRAP_PROXY_URL` and
`GOLIB_BOOTSTRAP_PROXY_SHA256` as repository variables. The URL must identify
an immutable HTTPS archive containing a file-based Go module proxy, and the
checksum must be its lowercase SHA-256 digest. The reusable workflow verifies
the archive before extraction and exposes it to both quality and CodeQL builds.
The pinned source fallback accepts bounded ordinary zero-record tar padding;
an installed archive-capable `golib` remains the validator and is not upgraded
by this action pin.
Defining only one variable, using a mutable or non-HTTPS URL, or providing an
invalid checksum fails closed. These variables bootstrap historical module
identity only; they do not replace normal dependency resolution or permit a
consumer to weaken any gate. Published module versions resolve from the public
Go proxy first. The file proxy is consulted only when the public proxy reports
that a version is unavailable, so bootstrap archives cannot shadow a published
module with different bytes.

The reusable workflow keeps consumer policy in repository manifests. It builds
a bounded module inventory from `golib inventory --json`, runs one all-module
quality invocation with module-specific gate attribution, uploads one aggregate
`evidence-selected-modules` artifact containing repository-owned `.verification`
evidence, and runs CodeQL. Repository secrets run once; all selected module
security scans finish before repository tests execute. The caller's final `required` job converts the reusable-call
result into the stable `Required` check used by branch protection.
Set `release_dry_run: true` only for an explicit release rehearsal; this first
validates the stable release contract and then runs the complete release
dry-run for every releasable module. Set `release_module` to an exact module
directory to limit module-scoped release checks to that
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
bounded module selection, attributable aggregate evidence, scheduled checks, CodeQL,
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

Own tooling builds and CI select Go 1.27.2 from `.go-version`; the module
language floor remains Go 1.27.0. Runtime or CI changes select the existing
native compatibility-consumer matrix without republishing the historical
Go 1.27.0 compatibility cohort or relabeling its observations.

This repository bootstraps its own CI from source so the first release does not
depend on itself. Consumer workflows use released binaries by default;
the optional source-bootstrap route explicitly selects reviewed development source.

Tooling tags publish only the four platform archives, their SBOMs, the release
manifest, and checksums by default. Catalog, compatibility-set, residual, and
source-lock assets are added only at a defined ecosystem milestone: refresh
and review the source lock, then set the repository variable
`GOLIB_CATALOG_MILESTONE_TAG` to the exact new tag before pushing it. The
exact-tag comparison prevents a stale variable from expanding later tooling
releases. Previously published release assets remain immutable.
