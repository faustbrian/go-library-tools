# Changelog

All notable changes to this project are documented in this file.

## Unreleased

### Added

- Added the versioned Golib ecosystem entry point and consumer-facing design
  language for explicit construction, ownership, lifecycle, failure,
  observability, adapter naming, compatibility, and application composition.
- Added schema-v2 cohesion metadata, strict machine-readable validation,
  deterministic consumer and engineering repository catalogs, and a local/CI
  `make cohesion` gate.

## 1.2.0 - 2026-08-31

### Added

- Added truthful restricted-source monitoring for licensed normative
  publications whose exact edition can be identified but whose content cannot
  be publicly retrieved or hashed, while retaining bounded denial-status probes
  and content-pinned public change monitoring.
- Separated official-fixture and provider interoperability evidence from
  maintained-peer differential results so each claim retains its actual proof
  boundary.
- Added a shared specification-decision contract that discovers
  specification-backed modules; validates strict decision, conformance,
  provenance, documentation, and change-control records; blocks unresolved
  release policy; and detects changed authoritative errata or release feeds in
  reusable CI. Conformance records distinguish authoritative sources from
  optional evidence lanes, record unassessed differential work honestly, and
  preserve superseded evidence without requiring retired artifacts to remain.

### Fixed

- Canonicalized repository roots before source discovery and derived current
  and legacy mutation identities from one captured snapshot so macOS `/tmp`
  aliases and other repository-root symlinks cannot retarget evidence between
  validation passes.
- Reused an existing legacy mutation report digest only when it matches the
  uniquely approved checkpoint, preserving immutable content-addressed records
  while continuing to reject every other evidence conflict.
- Accepted the complete BCP 14 requirement vocabulary in specification
  decisions and enforced normative scope for `REQUIRED` decisions.

## 1.1.0 - 2026-08-31

### Added

- Added content-identical legacy performance rehearsals with raw startup,
  inventory, checkpoint-reuse, module-scaling, peak-RSS, artifact-size, cache,
  and service-lifecycle measurements.
- Added a bounded `golib services cycle` command for fixture readiness and
  immediate cleanup without detached lifecycle state.

### Fixed

- Accepted non-decreasing, fully killed mutation inventories as strengthened
  compatibility while continuing to reject missing packages or weaker counts.
- Allowed explicit reuse of successful compatibility artifacts so comparison
  fixes do not rerun completed mutation campaigns.

## 1.0.14 - 2026-08-30

### Fixed

- Preserved production-specific exact-coverage denominators while allowing the
  complete module test set to contribute coverage in one run when integration
  or external tests exercise a package.

## 1.0.13 - 2026-08-29

### Fixed

- Exposed task-owned RabbitMQ Streams network and volume identities so
  package-owned rolling-upgrade tests can recreate generated services.

## 1.0.12 - 2026-08-29

### Fixed

- Ran RabbitMQ Streams healthchecks as the broker account so diagnostics
  cannot create an Erlang cookie that the broker cannot read.

## 1.0.11 - 2026-08-29

### Fixed

- Prevented RabbitMQ image data from being copied into task-owned broker
  volumes so generated Erlang cookies retain readable ownership.

## 1.0.10 - 2026-08-29

### Fixed

- Included bounded, credential-redacted logs from the failed RabbitMQ Streams
  container when Docker Compose cannot start the topology.

## 1.0.9 - 2026-08-29

### Fixed

- Retained the bounded tail of Docker Compose diagnostics so image-pull
  progress cannot hide the final service startup error.

## 1.0.8 - 2026-08-29

### Fixed

- Included bounded, credential-redacted Docker Compose diagnostics when a
  RabbitMQ Streams topology cannot start, while retaining exact cleanup of all
  task-owned resources.

## 1.0.7 - 2026-08-29

### Fixed

- Scoped each exact-coverage run to one production package and its tests so
  unrelated test binaries cannot pollute its profile or denominator.
- Excluded only the immutable `.golib-tooling` CI checkout from repository
  secret scans while retaining checks of all other tracked and untracked files.
- Prevented fenced and inline code examples from being interpreted as local
  Markdown links by the documentation checker.

## 1.0.6 - 2026-08-29

### Fixed

- Ignored nested fixture symlink entries when deriving mutation input identity,
  without following their targets or allowing symlinked data roots.

## 1.0.5 - 2026-08-29

### Fixed

- Allowed fixture modules that are not released to use non-repository module
  paths while retaining namespace enforcement for published modules.
- Excluded Go `testdata`, hidden, underscore-prefixed, vendored, and nested
  module directories from production formatting and safety scans.

## 1.0.4 - 2026-08-29

### Fixed

- Preserved the original checkpoint runtime metadata when importing mutation
  evidence so content-identical proof is reusable across CI platforms.
- Applied the checksum-verified bootstrap module proxy to both quality and
  CodeQL builds in consumer repositories.

## 1.0.3 - 2026-08-28

### Fixed

- Delegated formatting checks to the active repository Go toolchain so the
  released CLI cannot impose formatting rules from the Go version used to
  build the CLI itself.
- Applied approved mutation checkpoint imports consistently through both
  `golib mutation` and the full `golib check` contract.

## 1.0.2 - 2026-08-28

### Fixed

- Removed an unnecessary package-read permission from the reusable CodeQL job
  so least-privileged consumer workflows can start successfully.

## 1.0.1 - 2026-08-28

### Changed

- Scoped mutation identities to sibling modules actually observed by each
  package and added deterministic migration of previously approved v1 input
  identities, so unrelated module inventory changes do not trigger campaigns.
- Added repository-declared mutation checkpoint imports that materialize
  approved source-specific evidence before verification, including in
  service-backed CI where local imports are intentionally avoided.
- Kept release verification bootstrapped by the previously published tooling
  pin instead of requiring an unpublished release to verify itself.
- Required a successful full CI run for the exact tagged commit instead of
  repeating the repository contract and mutation campaigns during publishing.

## 1.0.0 - 2026-08-28

### Changed

- Ran RabbitMQ Streams readiness checks with the broker's task-owned Erlang
  cookie identity and kept nested-module license audits scoped to the owning
  repository namespace.
- Generated each rehearsal module's exact tidy checksum set so transitive
  owned-module checksums are present without retaining stale graph entries,
  while preserving module-file drift for the strict tidy gate and resolving
  nested module dependencies from their own source directories.
- Added an opt-in CI release rehearsal that executes the complete pre-tag
  dry-run after ordinary quality checks and feeds its outcome into the stable
  required job.
- Advanced the Knapsack and OpenAPI compatibility rehearsals to their finalized
  standalone tooling repairs.
- Kept the isolated-module environment aligned with task-owned execution module
  files across nested legacy-tool invocations.
- Updated the central `golangci-lint` pin to `v2.13.1` so Linux Go 1.27 checks
  use its compatible Staticcheck tool graph without disabling analyzers.
- Isolated intentionally replaced internal module checksums in task-owned
  rehearsal module files and propagated those identities through nested and
  legacy child Go processes without changing representative repository content,
  while deferring locally proxied modules to their task-owned archive identity.
- Isolated checksum-pinned verifier construction from consumer module flags so
  external tooling always resolves against its own module graph.
- Injected refreshed standalone dependency checksums into per-process execution
  module-file copies, leaving source-comparable sums unchanged under overlapping
  commands.
- Routed copied legacy tooling's explicit Go entrypoint through the rehearsal
  wrapper and exported that wrapper through the GitHub Actions path channel so
  checksum isolation applies consistently.
- Materialized task-local proxy checksums in per-process rehearsal sums before
  readonly nested-module commands, without changing tracked dependency sums.
- Anchored every configured rehearsal fuzz target so one target cannot
  accidentally select another target with the same prefix.
- Ran versioned analyzer binaries against task-owned source snapshots carrying
  refreshed module sums, so analyzer subprocesses cannot bypass rehearsal
  dependency isolation.
- Enforced module mode in the rehearsal Go wrapper so analyzer subprocesses
  cannot disable module-aware package discovery for nested modules.
- Kept repository-wide Markdown, spelling, and link validation on the root
  module while nested documentation gates run bounded example checks.
- Scoped mutation verifier builds and campaign artifacts by module identity to
  prevent independently verified modules from sharing mutable workspace paths.
- Published the tooling repository's verification tree after every CI outcome
  so newly generated content-addressed evidence is retained immediately.
- Kept shared compatibility gates on each representative repository's declared
  Go version while allowing the source CLI build to select its required toolchain.
- Serialized package tests inside mutation campaigns so deadline behavior is
  not determined by parallel test scheduling.
- Canonicalized local task workspace paths so macOS temporary-directory
  symlinks cannot change repository path identity during verification.

### Added

- Strict, bounded `.golib.yaml`, module-manifest, and package-manifest loading.
- Deterministic `config validate`, `inventory`, and initial `check` commands.
- Task-owned Go build, module, binary, and temporary caches for gate execution.
- Formatting, module-tidiness, unsafe-code, vet, test, and race gates.
- Strict typed operations for repository-specific conformance, documentation,
  API, fuzz, interoperability, and benchmark behavior without shell parsing.
- Exact production-package coverage enforcement and pinned analyzer, nil-safety,
  vulnerability, secret, and license tooling.
- Versioned content-addressed evidence records with history-independent identity,
  symlink rejection, atomic publication, and semantic concurrent reuse.
- Built-in exported API compatibility checks and atomic baseline updates using
  the pinned `apidiff` verifier and task-owned snapshots.
- Standalone repository validation for module identity, Go versions, workspace
  membership, committed replacements, and complete legacy-tool removal.
- Deterministic evidence inspection that validates attribution, content paths,
  duplicate identities, bounded records, and symlink-free evidence trees.
- Standalone exact-coverage execution with deterministic module selection and
  explicit not-applicable reporting.
- Strict legacy mutation-checkpoint and zero-mutant review validation with
  bounded archive expansion, duplicate rejection, and complete-kill proofs.
- Strict semantic-migration ledger validation that preserves approved mutation
  evidence without making new records depend on mutable Git history.
- Embedded, checksum-pinned Gremlins verifier assets whose derived identity
  exactly reproduces the verifier used by approved legacy campaigns.
- A canonical repository-owned mutation evidence root in `.golib.yaml`, with
  strict path validation and a published schema default.
- Package-granular mutation input identities derived from the observed Go
  dependency graph, exact source and fixtures, semantic policy, and verifier.
- Standalone strict Gremlins report validation for runtime campaigns and
  imported checkpoints, including bounded input and aggregate reconciliation.
- Native standalone mutation gate execution with isolated local-module
  replacements, exact evidence reuse, and immediate package-level persistence.
- Parallel-safe generic service leases for PostgreSQL, Valkey, Redis, NATS,
  NSQ, and RabbitMQ with exact cleanup and runtime image identities.
- Module-scoped fixture environments across standard, coverage, and mutation
  gates, including cleanup on failure and service-bound mutation evidence.
- Digest-pinned OpenSearch fixtures loaded from strict module-owned image
  policies without evaluating repository content as shell input.
- Parallel-safe RabbitMQ Streams standalone fixtures with dynamic loopback
  ports, task-owned credentials and networks, and pinned Toxiproxy control.
- Parallel-safe RabbitMQ Streams cluster, rolling-upgrade, authorization, and
  mutual-TLS topology with task-owned Compose files, volumes, and certificates.
- The executable entry point is covered through an injectable, side-effect-free
  boundary so the tooling repository can enforce coverage uniformly.
- Standalone repository, contributor, security, support, architecture,
  configuration, verification, fixture, workflow, migration, and release
  documentation for the shared tooling contract.
- Exact released-binary compatibility checks with an explicit source-build
  identity for non-circular tooling development.
- Canonical self-hosted repository configuration, module and package
  manifests, analyzer policy, secret scanning, and dependency notices.
- Exact coverage now merges duplicate source blocks emitted by multiple Go
  test binaries before evaluating each production package.
- Inventory validation now rejects duplicated or divergent package and module
  identities across the two canonical manifests.
- Native bounded Markdown validation with deterministic local-link checks and
  a dedicated `golib docs check` command.
- Stable release metadata validation and full releasable-module dry runs through
  the `golib release` command family.
- A reusable consumer CI workflow with immutable tooling identity, module
  matrices, attributable evidence, CodeQL, and one stable required check.
- A checksum- and provenance-verifying setup action plus attested Linux and
  macOS amd64/arm64 release automation with SPDX SBOMs.
- Strict Deno and zsh runtime policy consumed by package-neutral reusable CI.
- Task-owned documentation spelling with an embedded, exact CSpell dependency
  graph and repository-owned dictionaries.
- Checksum-pinned external-link verification with bounded hostile-archive
  handling and no consumer-owned installation scripts.
- Bounded self-hosted fuzz campaigns for every untrusted tooling input family.
- Self-hosted exact mutation verification with durable content-addressed
  package reports and a repository-owned zero-mutant inventory.
- A strict maintained consumer inventory with schema validation, CLI reporting,
  and bounded dry-run-first pull-request automation for coordinated immutable
  tooling upgrades.
- CI-only compatibility rehearsals for pinned representative libraries,
  including content, module, gate, service lifecycle, cleanup, coverage,
  mutation, advisory, and release-decision parity.
- Release dry-runs now reject tag collisions and verify task-owned module proxy
  archives through clean module resolution before executing release gates.

### Fixed

- Compatibility rehearsals now preserve consumer source byte-for-byte and use
  separate Go toolchains for building the CLI and exercising each consumer.
- Module manifests now accept the structured goal-evidence records emitted by
  current standalone repositories while retaining strict unknown-field checks.
- Rejected mutation checkpoint migrations now report the exact replacement
  input digest required for a reviewed approval ledger.
- Mutation input identity now excludes the synthetic test executable emitted
  by `go list`, which otherwise attributed generated cache source to a package.
- Embedded zero-context verifier patches are applied explicitly as
  zero-context diffs instead of depending on Git's default patch context.
