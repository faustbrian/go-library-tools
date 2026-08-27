# CLI Reference

`golib --help` lists the stable command surface. Commands locate the repository
by walking upward to `.golib.yaml`; failure to find it exits with status 1.
Invalid command syntax exits with status 2. Failed validation or gates exit
with status 1. Success exits with status 0.

Core commands:

- `golib --version` prints `dev` for source builds or the immutable release.
- `golib config validate` validates configuration and both manifests.
- `golib config show --json` emits normalized policy for workflow orchestration.
- `golib inventory [--json]` prints the canonical module inventory.
- `golib repository check` validates standalone repository structure.
- `golib workflows check` validates GitHub Actions workflows with the centrally
  pinned Actionlint release.
- `golib check [--all|--module DIR]` runs the enabled contract.
- `golib coverage [--module DIR]` verifies exact package coverage.
- `golib mutation [--module DIR]` verifies or executes mutation evidence.
- `golib mutation import --module DIR --archive PATH --ledger PATH` validates
  an approved legacy checkpoint archive and persists current content-addressed
  evidence without retaining Git identity.
- `golib api check|update [--module DIR]` checks or deliberately updates API
  baselines.
- `golib evidence inspect [--json]` validates attributable evidence records.
- `golib docs check [--module DIR]` validates bounded Markdown navigation and
  optional typed documentation tests.
- `golib release check` validates stable versions, tag prefixes, mandatory
  gates, and repository structure. `golib release dry-run` then executes the
  complete contract for every releasable module.

Selection and output ordering are deterministic. Required gates fail closed;
NilAway remains advisory and visible.

Released binaries must exactly match `.golib.yaml`. The explicit `dev` identity
is accepted only for deliberate source builds and prevents circular bootstrap
while developing this repository.
