# Performance Rehearsals

Performance rehearsals compare copied legacy tooling and `golib` against
content-identical representative source. They are diagnostics, not marketing
benchmarks: contract parity remains the prerequisite, and no result permits a
gate, fixture, evidence check, or cleanup guarantee to be removed.

## Methodology

The workflow checks out each representative at one immutable revision, hashes
the tracked source before applying the shared-tool policy, and runs each
implementation in an isolated Ubuntu 24.04 job. Every Go subprocess uses
task-owned disposable caches. Generic service fixtures use unique task-owned
resources and are audited after every job. Every raw report binds the tooling
commit, built CLI digest, workflow repository, run ID, and run attempt.

The core profile uses `go-knapsack`, which has a root module and two nested
modules. The service profile uses `go-authorization`, which requires PostgreSQL
and Valkey. The compatibility rehearsal separately covers the complete gate
contract for all five representatives.

Measurements are repeated and retained as individual samples rather than
collapsed into one favorable number:

- **Startup diagnostic:** invokes each command surface without arguments and
  requires the same usage-error exit state. This measures process and argument
  parsing startup, not repository work.
- **Repository inventory:** compares the legacy `inventory` Makefile target's
  underlying repository validator with `golib inventory`. Both load the same
  canonical manifest; their output formats intentionally differ.
- **No-op and checkpoint reuse:** warms the root mutation checkpoint and then
  repeats the same mutation command three times. The measured repetitions must
  report content-identical reuse for every production package; a missing,
  newly executed, or partially reused package fails the rehearsal.
- **Concurrent module scaling:** runs the same API check for every Knapsack
  module first sequentially and then concurrently. The commands, source, and
  selected modules are identical; only scheduling changes.
- **Service lifecycle:** starts, waits for, and closes the authorization
  repository's declared PostgreSQL and Valkey fixtures. Legacy uses its copied
  lifecycle scripts; `golib services cycle` performs the same bounded lifecycle
  without detached state.
- **Peak RSS:** GNU `time` records maximum resident set size for every directly
  measured command sample.
- **Artifact size:** reports copied `.golib` bytes for legacy and the built
  `golib` binary bytes for shared tooling. These are distinct deployment
  artifacts and are reported without claiming byte-for-byte equivalence.
- **Isolated cache behavior:** the job records any entries left in its dedicated
  runtime directory. A non-zero residue fails result validation.

## Interpretation

Runner variance is material because GitHub-hosted jobs do not share a physical
machine. Compare medians and distributions, not one sample, and do not infer a
runtime regression from small differences. Startup and inventory results do
not predict full contract duration, which is dominated by repository tests and
external tools. Service results include readiness and cleanup because those are
part of the lifecycle guarantee.

Concurrent module scaling is bounded by module count and runner CPU. A faster
concurrent result proves parallel-safe orchestration for the measured modules;
it does not imply linear scaling for service-heavy or CPU-saturated packages.

## Raw Results

Each workflow run uploads four unmodified implementation reports and one
validated `performance-results.json` aggregate. Accepted release results are
retained under [performance-results](performance-results/README.md) with the
tooling version, source revisions, workflow run, and complete sample arrays.
The aggregation script rejects missing profiles, mismatched source identities,
failed commands, partial checkpoint reuse, malformed sample counts, and leaked
task-owned runtime or service resources.
