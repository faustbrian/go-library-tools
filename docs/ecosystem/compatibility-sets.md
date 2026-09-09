# Golib Compatibility Sets

Generated from `compatibility-sets.json`; edit the JSON source only.

## `golib-core-2026-09`

**Status:** unreleased; **non-installable**. Observed `2026-09-08T00:00:00Z`.

### Modules

- `github.com/faustbrian/go-authentication@v1.0.0` (source `d932f8e875c928daf8b63328ad4b2b8cb6b0c9fe`)
- `github.com/faustbrian/go-service@v1.0.0` (source `62fc1c1340c679c468d72e984efbaf41c01bcd2e`)
- `github.com/faustbrian/go-tenancy@v1.0.0` (source `1c474f0a84ae3cc2c3988bb002d57d41cda3ae73`)

### Planned/unverified scenarios

- compose authenticated service startup with caller-owned lifecycle and tenant propagation
- exercise shutdown and cancellation ordering without a framework bootstrap
- resolve modules from clean source identities without replace directives

**Go:** `1.26.6` on darwin/arm64, linux/amd64, linux/arm64.

### Caveats

- Non-installable until each module is published at the listed version and clean external-consumer verification passes.
- No external service, database, or production deployment acceptance is implied.
- The content fingerprint covers this set definition and must be recomputed before publication.

### Evidence

- Content fingerprint: `sha256:51635ec7e8cf236760d33a9cf3a8eb115e729187545c0cb14c88aa567292d3fd`
- Observation: `Local source commits verified; release tags unavailable for some modules; cross-module consumer execution remains pending and unverified.`
