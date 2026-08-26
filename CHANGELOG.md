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
