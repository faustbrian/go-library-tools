# Compatibility Rehearsals

Compatibility rehearsals compare copied legacy tooling with `golib` against
content-identical package source. A successful new run is not sufficient by
itself: the comparison records selected modules, effective gates, exit states,
coverage denominators, mutation inventories, advisory behavior, and retained
package-specific operations. Service-backed representatives additionally prove
that the same service requirements are selected and that no task-owned
container, network, volume, or fixture lock survives either run.

Preparation hashes tracked consumer content without rewriting module files or
dependency sums and fails if the repository state changes. Shared runs build
`golib` with the tooling repository's Go version, then switch to the consumer's
declared Go version before executing its contract.

Some internal `v1.0.0` tags were intentionally replaced before public adoption,
leaving historical checksums in the representative repositories. Rehearsals
therefore prepare task-owned alternate module and sum files, refresh only those
internal dependency sums, and select the matching alternate file through a Go
wrapper. Tracked `go.mod`, `go.sum`, package source, fixtures, and evidence stay
unchanged. The wrapper resolves canonical module identities, so disposable
repository snapshots and nested modules select the correct alternate file.
The root alternate module is also exported explicitly for child processes that
resolve the Go executable before applying command-specific environments.

The remote representative rehearsal normalizes the copied legacy Staticcheck
pin from `v0.7.0` to `v0.8.1` before execution. Staticcheck `v0.7.0` cannot read
Go 1.27 compiler export data, so leaving that obsolete tooling dependency in
place would stop the legacy contract before later gates and would compare tool
incompatibility rather than repository behavior. Only copied tooling is
changed; the content digest proves the package source and repository-owned
fixtures remain identical.

The copied legacy isolated-Go shim also retains a monorepo-era checksum filter.
The rehearsal updates that copied filter from the retired `golib/` module path
to standalone `go-*` modules so legacy and shared checks consume the same
intentionally replaced dependency identities.

Shared lint execution pins `golangci-lint` `v2.13.1`. That release bundles
Staticcheck `v0.8.0`, avoiding the Go 1.27 analyzer panic in the older
`golangci-lint` `v2.12.2` tool graph while preserving the complete lint policy.

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
