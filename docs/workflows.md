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

permissions:
  contents: read
  security-events: write

jobs:
  ci:
    uses: faustbrian/go-library-tools/.github/workflows/library-ci.yml@0123456789abcdef0123456789abcdef01234567 # v1.0.0
    with:
      tooling_sha: 0123456789abcdef0123456789abcdef01234567
      release_dry_run: ${{ inputs.release_dry_run || false }}
```

The setup action reads the exact `tool_version` from `.golib.yaml`, selects the
Linux or macOS amd64/arm64 archive, downloads it with GitHub CLI, verifies its
SHA-256 checksum and GitHub artifact attestation, and only then extracts the
binary. It never evaluates a downloaded installer.

The reusable workflow keeps consumer policy in repository manifests. It builds
a module matrix from `golib inventory --json`, runs one isolated module
contract per matrix entry, uploads repository-owned `.verification` evidence,
runs CodeQL, and exposes one stable `Required` job for branch protection.
Set `release_dry_run: true` only for an explicit release rehearsal; this first
validates the stable release contract and then checks every releasable module.

Consumer workflows retain least-privileged permissions, explicit concurrency,
module matrices, attributable evidence artifacts, scheduled checks, CodeQL,
release dry-runs, and one stable final required job.

Pull-request runs review dependency changes before the final required job.
Workflow syntax and expression validation are available locally through
`golib workflows check` and run as part of the repository `make ci` contract.

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
