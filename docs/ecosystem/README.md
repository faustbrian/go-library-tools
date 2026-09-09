# Golib Ecosystem

This directory is the versioned public entry point for the independently
released Golib libraries. It describes shared consumer expectations without
creating an umbrella module, application runtime, service container, or
lockstep release train.

- [Design language](design-language.md): construction, ownership, lifecycle,
  errors, adapters, compatibility, and explicit composition.
- [Cohesion goal contract](goals/cohesion-v1.md): the versioned ecosystem
  outcome, delivery phases, evidence semantics, and completion rules.

At an ecosystem milestone, the reviewed repository revisions and adopted
tooling identities are declared in
[`release/cohesion-sources.json`](../../release/cohesion-sources.json) and
validated by
[`schema/cohesion-sources.schema.json`](../../schema/cohesion-sources.schema.json).
The matching standalone input manifest lists only SHA-256-bound repository
engineering projections described by
[`schema/cohesion-inputs.schema.json`](../../schema/cohesion-inputs.schema.json).
Projection paths are safe POSIX paths confined to the manifest directory. A
checksum-verified released `golib` binary generates deterministic consumer and
engineering catalogs; the consumer view contains releasable libraries and
adapters, while the engineering view also retains fixtures, harnesses,
examples, benchmarks, and internal tools. The matching `cohesion aggregate
check` command byte-verifies the checked-in set. Aggregation accepts at most
256 repositories, 256 MiB of projection input, and 4,096 modules; each rendered
artifact is capped at 512 MiB.

Only a selected ecosystem-milestone tooling release publishes the source lock,
standalone input manifest, deterministic projection bundle, and four catalogs
as checksum- and attestation-covered assets. That milestone pipeline
independently regenerates locked projections and catalogs and binds bundle
membership to the source lock before publication. Ordinary tooling and library
releases do not refresh or publish these aggregate artifacts.

Compatibility-set drafts use `draft-<YYYYMMDD>.<sequence>` identifiers and are
non-installable. Published sets use
`golib-compat-v<schema>-<YYYYMMDD>.<sequence>`, are installable, and bind every
module to its owning repository's exact remote release tag and public Go proxy
version. Set records can also carry structured recipe,
external-version, upgrade, rollback, and exclusion details. Published
identifiers are immutable; a later recommendation receives a new identifier.

The development copy on a default branch is not an immutable compatibility
contract. Published documents identify their design-language version, exact
`go-library-tools` tag, and content digest.
