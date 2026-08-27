# Changelog

All notable changes to this project are documented in this file.

## Unreleased

### Changed

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
