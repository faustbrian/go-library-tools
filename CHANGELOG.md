# Changelog

All notable changes to this project are documented in this file.

## Unreleased

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
