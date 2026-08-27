# Compatibility Rehearsals

Compatibility rehearsals compare copied legacy tooling with `golib` against
content-identical package source. A successful new run is not sufficient by
itself: the comparison records selected modules, effective gates, exit states,
coverage denominators, mutation inventories, advisory behavior, and retained
package-specific operations.

## `go-clock`

`go-clock` is the root-only, pure-library rehearsal. The migrated copy retained
the same production and test source while replacing copied `.golib` tooling
with `.golib.yaml`, a thin Makefile, and the reusable workflow caller.

The legacy aggregate command could not complete under Go 1.27 because its
copied Staticcheck 0.7.0 binary could not decode the compiler export data. This
is a legacy toolchain incompatibility rather than a package failure. The shared
tooling used its centrally pinned compatible Staticcheck version and passed.

The rehearsal also established that the package-specific stress and leak tests
were not part of the legacy aggregate command. They are represented as typed
test operations in `.golib.yaml`, so the migrated aggregate contract is
strictly stronger without changing package behavior.

Legacy mutation checkpoints were not imported because their content identity
was not approved for migration. No approval was fabricated. Fresh campaigns
killed 39 of 39 root-package mutants and 155 of 155 `manual` package mutants.
The migrated repository then passed its complete shared contract, including
exact coverage, race, mutation, fuzz, documentation, API, benchmark, static
analysis, security, license, and SBOM gates.

The rehearsal changed no Go source or public API and left no dependency on the
archived monorepo or copied `.golib` implementation.

## Required Coverage

Additional rehearsals must cover a large package tree, genuine multi-module
layout, service-backed tests, broker-backed tests, specification conformance,
external runtimes, nested release units, and large or zero-mutant inventories
before the first public release. Each difference must be resolved or recorded
as an intentional compatibility decision before migration.
