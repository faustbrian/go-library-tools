# Consumer Inventory And Upgrades

[`consumers.json`](../consumers.json) is the maintained source of truth for
repositories governed by this tooling project. Its schema is published at
[`schema/consumers.schema.json`](../schema/consumers.schema.json). Entries are
sorted, unique, and classified as:

- `active`: receives the shared tooling contract and future upgrade proposals;
- `deferred`: intentionally excluded with a required reason;
- `tooling`: the single repository that owns the CLI and reusable workflows.

Run `golib consumers validate` for a concise inventory summary or add `--json`
for automation. Validation rejects unknown fields, malformed repository or
branch names, duplicates, unsorted entries, unsupported classifications, and
missing classification reasons.

## Initial Adoption

The first migration of a repository is deliberate and repository-specific.
Follow the [migration guide](migration.md), preserve package-owned fixtures and
evidence, prove parity, and review the complete change before removing copied
`.golib` implementations. The rollout workflow does not perform first-time
adoption because that would hide repository-specific migration decisions.

## Released Tool Upgrades

After adoption, the `Propose consumer upgrades` workflow coordinates the three
immutable identities that define one tooling release:

- the exact semantic version of the `golib` binary;
- the SHA-256 digest of the release `checksums.txt` asset;
- the commit SHA used by the reusable workflow and setup action.

Dispatch the workflow with a JSON cohort of one to ten active repository names.
`dry_run` defaults to `true`; this validates the inventory and prints the exact
planned changes without creating branches or pull requests. Apply mode updates
only `.golib.yaml` and `.github/workflows/ci.yml`, validates the result, and
opens one pull request per changed repository. It never force-pushes or mutates
an existing rollout branch.

Apply mode requires a fine-grained `GOLIB_ROLLOUT_TOKEN` repository secret with
read/write contents and pull-request access to the selected consumer
repositories. GitHub's repository-scoped workflow token is insufficient for
cross-repository writes. Keep cohorts small, inspect failures independently,
and merge only after each consumer's own required CI succeeds.
