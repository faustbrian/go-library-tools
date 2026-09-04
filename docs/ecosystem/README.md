# Golib Ecosystem

This directory is the versioned public entry point for the independently
released Golib libraries. It describes shared consumer expectations without
creating an umbrella module, application runtime, service container, or
lockstep release train.

- [Design language](design-language.md): construction, ownership, lifecycle,
  errors, adapters, compatibility, and explicit composition.

The reviewed repository revisions and their adopted tooling identities are
declared in the closed
[`release/cohesion-sources.json`](../../release/cohesion-sources.json) lock and
validated by
[`schema/cohesion-sources.schema.json`](../../schema/cohesion-sources.schema.json).
The ecosystem catalogs are derived only from the resulting explicitly listed,
SHA-256-bound repository engineering projections described by
[`schema/cohesion-inputs.schema.json`](../../schema/cohesion-inputs.schema.json).
Projection paths are safe POSIX paths confined to the manifest directory.
A checksum-verified released `golib` binary generates
`catalog-consumer.{json,md}` and `catalog-engineering.{json,md}`. The consumer
view contains releasable libraries and adapters; the engineering view retains
fixtures, harnesses, examples, benchmarks, and internal tools. Run the matching
`cohesion aggregate check` command to byte-verify a checked-in catalog set.
Aggregation accepts at most 256 repositories, 256 MiB of projection input, and
4,096 modules; each rendered artifact is capped at 512 MiB.

Each tooling release publishes the source lock, standalone aggregation input
manifest, deterministic projection bundle, and all four catalogs as assets
covered by checksums and attestations. The release pipeline independently
regenerates every locked projection and catalog, and binds exact bundle
membership back to the source lock before the write-capable publication job
runs.

The development copy on a default branch is not an immutable compatibility
contract. Published documents identify their design-language version, exact
`go-library-tools` tag, and content digest.
