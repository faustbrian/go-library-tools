# Compatibility Rehearsals

Compatibility rehearsals compare copied legacy tooling with `golib` against
content-identical package source. A successful new run is not sufficient by
itself: the comparison records selected modules, effective gates, exit states,
coverage denominators, mutation inventories, advisory behavior, and retained
package-specific operations. Service-backed representatives additionally prove
that the same service requirements are selected and that no task-owned
container, network, volume, or fixture lock survives either run.

Preparation hashes tracked consumer content without rewriting module files or
dependency sums and fails if the repository state changes. Shared runs install
the consumer's declared Go version once. Building `golib` may select the
tooling module's required compiler through Go's automatic toolchain mechanism,
but consumer gates continue under the declared consumer version.

Some internal `v1.0.0` tags were intentionally replaced before public adoption,
leaving historical checksums in the representative repositories. Rehearsals
therefore prepare task-owned alternate module and sum files, refresh only those
internal dependency sums, and select the matching alternate file through a Go
wrapper. Tracked `go.mod`, `go.sum`, package source, fixtures, and evidence stay
unchanged. The wrapper resolves canonical module identities, so disposable
repository snapshots and nested modules select the correct alternate file.
The root alternate module is also exported explicitly for child processes that
resolve the Go executable before applying command-specific environments.
Third-party child modules stop resolution at their own nearest `go.mod`, so an
inherited rehearsal module file cannot replace their dependency graph.

The rehearsal wrapper injects current standalone dependency checksums into a
per-process module-file copy and removes that copy on success, failure, or
cancellation. Versioned external tools run without consumer module flags.
Tracked source, source-comparable sums, and copied legacy tooling remain
unchanged, including when child commands overlap. Legacy tooling's explicit Go
entrypoint is directed through the same wrapper rather than bypassing it, and
GitHub Actions receives the wrapper through its dedicated executable-path
channel. When copied tooling builds a task-local module proxy, the wrapper
materializes that archive's checksum in the per-process sum before a readonly
nested-module command starts. The historical tracked checksum remains intact,
while the command verifies the exact task-local archive it consumes.

Configured fuzz selectors are anchored regular expressions. This matches the
copied contract's exact target discovery and prevents a selector such as
`FuzzDecodeJSON` from also executing `FuzzDecodeJSONBatch`.

Versioned analyzers are compiled without consumer module flags, then executed
against a disposable tracked-source snapshot containing the selected module's
refreshed module and sum files. This is required because Go places
`$GOROOT/bin` ahead of wrapper paths for subprocesses started by `go run`, and
`go/packages` performs module-disabled discovery calls that cannot accept a
`-modfile` flag. The snapshot is task-owned, process-specific, and removed as
soon as the analyzer exits; the representative checkout remains unchanged.

Mutation campaigns use module-identity-scoped workspaces. Verifier source,
coverage profiles, reports, and caches therefore cannot collide when one
repository verifies multiple independently releasable modules.

Shared rehearsals install the representative repository's declared Go version
once. The source CLI build may use Go's automatic toolchain selection for its
own module, but every consumer gate continues under the representative version.

Mutation campaigns serialize package tests while retaining parallel mutant
workers. This prevents test-level scheduler contention from deciding short
deadline assertions without reducing the mutant inventory or gate strictness.

Shared runs temporarily hide copied `.golib` tooling only while validating the
standalone repository contract, then restore it before exercising the package
gate set. Shared fixture files are excluded through task-owned Git metadata.
This keeps production code, tests, evidence inputs, and package behavior
content-identical during parity; each real consumer migration removes copied
tooling and updates tooling-coupled package checks in its own reviewed batch.
Release-decision checks hide copied tooling again so standalone validation sees
the final repository shape; current representative tags then exercise the same
tag-collision decision as the copied release contract without rerunning gates.

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
