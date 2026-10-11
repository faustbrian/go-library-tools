# Reusable Workflows

## Tools-only aggregate milestones

Tools' own CI runs `make milestone-check` only when a compatibility set is
published, an already-published set changes, or a manual run selects
`ecosystem_milestone: true`. Ordinary
main pushes and scheduled runs retain their applicable local checks without
automatically repeating the aggregate catalog campaign. Changed consumer inputs
still select their separate compatibility-consumer qualification.
PR/main classification retains the existing generator's offline structure and
generated-artifact checks. `python3 tools/compatibility/generate.py check --offline`
does not refresh remote tags or public proxy identities; full publication and
explicit milestone qualification retain those checks.

The final Required check demands success from every applicable job, including
PR dependency review and explicitly selected milestones, rehearsals, and
diagnostics. Skipped applicable work is a failure, not an assurance pass.

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

## Tools-only selected mutation assurance

`mutation-assurance.yml` runs one configuration-admission package weekly on
main. A main-only manual dispatch can instead select `internal/config`,
`internal/inventory` or `internal/repositoryfile`. These cover configuration
admission, manifest admission and confined repository reads respectively.
There is no PR/push trigger and no dependency from Required. The owner is
`faustbrian`; inspect the selected-contract failure rather than treating every
survivor as a production bug. This is partial assurance, not a full-module or
release qualification.

The job has a 30-minute ceiling; campaign execution has the CLI's 20-minute
budget and one worker. Module baselines and native integration tests still run.
Build, capability, campaign and output-admission failures fail the job. Native
checkpoint/report caches are populated only by successful trusted-main runs;
native input validation still determines whether a record is reusable. Restore
must not change committed evidence. Only flat JSON output is retained, bounded
to 2,048 files and 32 MiB, with no symlinked ancestors or entries. Cache paths
exclude binaries and equivalence-policy files. Native diagnostics are retained
as artifacts for 14 days; the cache provider controls cache eviction. Private
build/module caches are removed after the command, including failures. Hosted
timing and useful campaign completion must be measured before widening scope.

## Consumer workflows

Ordinary PR and main runtime changes select their committed base-to-HEAD module
owners through `golib inventory --json --base-sha <commit>`. The longest nested
module owns a changed path; declared relationships and actual owned `go.mod`
requirements/local replacements include repository-local reverse consumers.
Shared fixture modules, shared/unknown paths and an empty diff select all
modules. Fixture consumers can be scripts rather than Go import edges. Missing or invalid
Git input fails closed. Schedules, manual checks and explicit release rehearsals
retain their whole-repository or explicit release selection.

The quality job passes all selected directories to one
`golib check --local --base-sha <commit> --module <directory>
[--module <directory> ...]` invocation for PR/main events,
sharing its bounded executor session. These new capabilities require adopting
a qualified capable release or reviewed source bootstrap; older workflow pins
are unchanged. Committed-change inventory is a CI scope selector, not proof of
uncommitted or untracked local work; contributors should select modules explicitly
or use `--all` for that work. Metadata-only documentation checks remain separate.
Ordinary runtime quality has a 30-minute job ceiling, including setup; only an
explicitly selected release rehearsal retains the six-hour ceiling. Exhausting a
budget fails the applicable check and cannot satisfy Required. These ceilings are
failure limits, not measured feedback-time targets; pilot latency must still be
recorded separately.
Runtime and metadata PR/main routes require changed-secret capability; older
fuzz-capable tooling is rejected before checks, without a full-scan fallback.

PR/main checks add short active fuzz smoke only for configured typed targets
affected through their current Go production/test/embed dependencies. They retain
saved fuzz regressions in ordinary tests and do not start mutation campaigns.
Smoke has a three-minute total ceiling, a 20-second per-target wall deadline and
two seconds of fuzz work per target; invalid selection or an exceeded budget fails
honestly. Scheduled/manual local checks do not implicitly select this smoke or a
full campaign. See [CLI](cli.md) for limits and evidence boundaries.

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
The workflow requires the installed verifier to support bounded `check --local`
and PR/main changed-input `--base-sha` smoke.
An incapable verifier fails with an adoption diagnostic rather than falling
back to a full campaign. Keep an existing workflow pin until adopting a
qualified capable release or explicitly reviewed immutable source bootstrap.
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
Prepare builds once with the immutable remote setup action at
`e8c0c51cc517dfede5475858bb417f0e7a21f39d`, passing the same `tooling_sha`
used for its fixed tooling checkout. Before any library-controlled execution,
it binds the binary checksum and uploads only that binary, capped at 64 MiB,
with one-day retention. Dependent jobs use `tools/reuse-source-golib.sh`
from their exact tooling checkout and the pinned download action to reuse that
immutable artifact ID from
the same workflow run. There is no cross-run lookup, download token or cached
PR binary in privileged release workflows. Missing identities, unexpected
payloads, checksum mismatch or failed `dev` validation fail without rebuilding
or falling back to another run. Each job retains its own installation and
scratch cleanup. Adopt the workflow and supporting tooling together; this
source-reuse script is not present in older tooling. Actual transfer latency and
hosted cancellation cleanup need hosted qualification, not just local fixtures.
Documentation/license-only pull requests
use the ordinary metadata assurance tier, including in source-bootstrap mode.
Metadata-only PR/main jobs use `golib secrets check --base-sha <commit>` to
scan all new commits plus changed current files, including documentation.
Schedule/manual integrity checks retain full `golib secrets check` coverage.
Both routes retain bounded scan policy in installed and source-bootstrap modes.
Older tooling without the applicable command fails closed;
adopt the workflow and supporting tooling together.
Exactly paired reusable-workflow and tooling-pin changes use metadata assurance
in both installation modes, on pull requests and corresponding main pushes.
Runtime-selected ordinary jobs pass the affected module matrix to one
`golib check --local` invocation; bootstrapping does not select a campaign.
Schedules and manual runs retain their applicable
runtime route; release rehearsals remain explicitly dispatched. Every
source-mode job removes the exact owned installation and ancillary caches in an
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
a bounded selected module inventory, runs one quality invocation for that
selection with module-specific gate attribution, uploads one aggregate
`evidence-selected-modules` artifact containing repository-owned `.verification`
evidence, and runs CodeQL. Repository secrets run once; all selected module
security scans finish before repository tests execute. The caller's final `required` job converts the reusable-call
result into the stable `Required` check used by branch protection.
Set `release_dry_run: true` only for an explicit non-PR release rehearsal; a PR
attempt is rejected before it can select campaign work. A full rehearsal is
selected for a named risk or release boundary, not every tag; selected failures
still block affected publication. See the [release policy](release.md). This first
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

Pull requests and corresponding main pushes that change only Markdown,
`LICENSE`, or `NOTICE` run repository structural validation without launching
module-quality or CodeQL jobs. An
otherwise lightweight pull request may also update only the canonical
reusable-workflow SHA and matching `tooling_sha`; structural validation still
checks that caller, including when the replaced release comment has a legacy
descriptive suffix. Source, configuration, substantive or mismatched workflow
changes, structured metadata, scheduled, and explicit release-rehearsal
events retain the complete applicable runtime path.
Pull-request runs review dependency changes before the final required job.
Workflow syntax and expression validation are available locally through
`golib workflows check`, run as part of the repository `make ci` contract, and
execute in the reusable workflow's hosted repository-contract job.

Metadata routing also runs `golib docs check --local`: bounded repository-local
Markdown structure and link checks without network access or runtime commands.
This capability is required when adopting this workflow revision. The existing
full documentation command and runtime example checks remain separate.

The final required job rejects missing or invalid classifications, failed or
cancelled mandatory results, and applicable checks that were skipped. A job
deliberately excluded by the selected route is not a passing runtime check.

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
