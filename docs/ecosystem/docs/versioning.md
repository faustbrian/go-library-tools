# Install, upgrade and release independent modules

[Documentation map](index.md)

There is no `golib` umbrella application module. Consumers import only the
modules they need. The [published compatibility set](../compatibility-sets.md)
`golib-compat-v1-20260912.2` binds exact versions, source revisions, Go 1.27.0,
platforms, compositions, external versions, caveats, upgrade and rollback
guidance. It is immutable historical evidence, not a synchronized release train
or automatic statement of latest supported versions.

## Install one module

In your application's initialized Go module, select an exact module/version
from the set or the owning repository's public release. For example, the
frozen set contains this standalone time seam:

```sh
go get github.com/faustbrian/go-clock@v1.1.0
```

Use Go 1.27.0 for the documented cohort. Do not use a workspace or local
`replace` as proof that published modules resolve. Keep authenticated public
proxy/SumDB resolution enabled. Follow the chosen package's API and README;
the tooling CLI is not needed to import a library.

## Nested modules and major paths

A releasable nested adapter has its own `go.mod`, requirements and tag prefix.
For example, a nested `adapters/target` module is tagged
`adapters/target/vX.Y.Z` in its owning repository. Major v2+ modules retain Go's
required `/v2`, `/v3`, etc. module/import suffix even though work stays on main;
there are no versioned source branches or ecosystem source directories.
Confirm the actual tag prefix in the owning `modules.json`, not by guessing
from a directory. Fixtures may be intentionally non-releasable.

## Frozen cohort versus package-current adoption

The generated catalog is a snapshot of the set, not a mutable latest-release
directory. A newer major may change constructors, persistence names, wire
defaults or lifetime. Read its changelog and migration guide, then run the
application's affected contracts with public module resolution. Do not attach
the frozen set's compatibility label to that new selection without a newly
published set and attributable composition evidence. The maintained tooling
consumer may contain explicit released replacements/additions; it does not
rewrite membership of the frozen cohort.

Upgrade one owned dependency boundary at a time. Preserve old message readers
and database readers while new writers are introduced. Roll back code only
when it can still read persisted data/payloads; restoring `go.mod` cannot undo
an irreversible schema or external side effect. Use [migration](migration/index.md)
and [operations](operations/index.md) to prepare the actual rollback boundary.

## Maintainer publication

Each repository owns its release, changelog, exact-source CI and public
consumer proof. Read its CONTRIBUTING/release instructions and the
[tooling release process](../../release.md). Preserve Unreleased entries;
verify major path/tag alignment, checksums, SBOM and provenance where required.
Documentation-only edits do not require a module release. A tagged portal
snapshot is documentation publication, not a new compatibility cohort.

Next: [packages](packages.md), [compatibility set](../compatibility-sets.md),
[contributor route](contributing.md).
