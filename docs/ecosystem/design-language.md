# Golib Design Language

Design-language version: **1.0**

This document defines the shared consumer experience for independently
released Golib modules. A package remains responsible for its own domain API,
release, documentation, and compatibility. The common language makes packages
predictable to combine; it does not create a framework.

The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT",
"SHOULD", "SHOULD NOT", "RECOMMENDED", "NOT RECOMMENDED", "MAY", and
"OPTIONAL" in this document are to be interpreted as described in BCP 14
[RFC2119] [RFC8174] when, and only when, they appear in all capitals, as shown
here.

[RFC2119]: https://www.rfc-editor.org/rfc/rfc2119
[RFC8174]: https://www.rfc-editor.org/rfc/rfc8174

## What Golib Is

Golib is a collection of explicit, independently adoptable Go libraries.
Applications select packages, construct values, connect adapters, and own
lifecycle in ordinary Go code.

Golib MUST NOT require an umbrella runtime, application object, service
container, locator, facade, global registry, mandatory bootstrap, synchronized
release, or hidden initialization. A compatibility set is evidence that exact
versions worked together, not a dependency bundle.

## Standard-Library-First Boundaries

Public APIs MUST use standard Go types when those types express the complete
contract: `context.Context`, `error`, `time.Time`, `time.Duration`,
`io.Reader`, `io.Writer`, `fs.FS`, `http.Handler`, `http.RoundTripper`, and
`*slog.Logger`. A package MUST NOT wrap a standard type only to brand it.

Interfaces belong at the consuming boundary and SHOULD be minimal. Constructors
SHOULD return concrete types unless consumers or independently released
adapters need substitution. Reflection, generators, registration, and package
initialization MUST remain explicit optional mechanisms with visible failure
boundaries.

## Ownership And Dependency Direction

One semantic owner defines each domain contract. Adapters translate that
contract to a target; applications compose owners explicitly. An adapter MUST
NOT silently take ownership of application retry, transactions, brokers,
exporters, telemetry providers, loggers, or service lifecycle.

Core packages SHOULD depend on standard interfaces or narrowly owned domain
contracts instead of optional implementations. A new cloud SDK, database
driver, broker client, telemetry implementation, native dependency, or external
service client MUST live in an independently releasable nested module unless
the installed core capability necessarily requires it.

A direct persistence package MAY remain under its domain owner when persistence
is part of the installed core capability, its driver is already required by
the root module, and it owns semantic storage behavior. The catalog identifies
that case as `domain-owned`; other backends are adapters.

## Construction

Choose construction by observable behavior, not visual uniformity.

| Shape | Consumer expectation |
| --- | --- |
| Plain function | Stateless, deterministic work with explicit inputs and no hidden I/O. |
| `New(Config)` | Complete validation before work, no retained caller-mutable aliases, `(T, error)` when invariants may fail. |
| `New(Options)` | One coherent named settings value with the same validation and ownership rules as `Config`. |
| Functional options | Genuinely optional, orthogonal settings with deterministic ordering and an explicit duplicate/conflict rule. |
| Builder then `Compile` or `Build` | Single-owner registration followed by complete validation and an immutable concurrent runtime. |
| `Open`, `Connect`, `Load`, or `Init` | Explicit external I/O or acquisition with `context.Context`, bounded work, returned ownership, and cleanup. |
| `Must...` | An explicit panic contract for tests, generated constants, or process startup; never the only production API. |

Packages MUST NOT rename `Config` to `Options`, introduce functional options,
or add a builder only to resemble another package. Equivalent concepts SHOULD
use the same shape when no domain or compatibility constraint distinguishes
them.

## Configuration, Defaults, And Validation

Configuration is explicit and caller-owned. Constructors MUST reject invalid
configuration before starting background work or external I/O unless the
constructor explicitly performs acquisition.

Safe defaults MUST be visible through a documented zero-value contract,
`DefaultConfig`, a named profile, or the constructor. An API MUST distinguish
absent, zero, empty, disabled, and defaulted states when their behavior differs.
Functional options MUST define duplicate and conflict behavior.

Maps, slices, byte buffers, callbacks, and mutable collaborators MUST document
whether they are copied, borrowed, or transferred. Ordinary libraries MUST NOT
discover configuration from environment variables or files. Packages whose
domain is configuration loading MAY do so through explicit bounded operations.
Validation errors MUST identify safe field paths or option names without
exposing secret values.

## Context, Cancellation, Deadlines, And Time

An operation that may block on I/O, waiting, external callbacks, resource
admission, or long-running work MUST accept `context.Context` as its first
parameter. Pure value operations SHOULD NOT accept context solely for
uniformity. Request contexts MUST NOT be retained after their operation.

Packages MUST NOT detach work from caller cancellation unless a separately
owned lifecycle is explicit. Retries, hedges, queues, worker attempts, and
deadlines MUST share one observable total-work budget when they compose. A
timeout MUST NOT claim to stop arbitrary code that ignores cancellation.

Deterministic packages SHOULD accept an explicit clock or time seam when wall
or elapsed time affects behavior. A result whose external side effect is
unknown after cancellation or transport failure MUST remain distinguishable
from known failure.

## Resource And Lifecycle Vocabulary

Packages expose only the lifecycle operations they own:

- `Run(ctx)` performs one caller-bounded long-running execution and returns its
  terminal result.
- `Start(ctx)` transfers ownership after successful startup and requires a
  documented stop or join operation.
- `Drain(ctx)` stops new intake and waits for accepted work to reach its stated
  drain boundary.
- `Shutdown(ctx)` is repeatable, safe for concurrent calls, bounded by context,
  and performs the object's complete ordered shutdown.
- `Close()` performs immediate synchronous release compatible with
  `io.Closer`. Context-aware cleanup uses another name.
- `Wait(ctx)` waits for an already-owned operation and does not acquire hidden
  ownership.

Every resource-owning API MUST document acquisition, ownership transfer,
partial-start rollback, cleanup responsibility and order, repeated and
concurrent cleanup, goroutine and timer lifetime, connection and transaction
lifetime, buffer ownership, deadline behavior, and use after shutdown.

## Errors And Outcomes

Golib uses ordinary Go errors. There is no ecosystem-wide exception base.

Stable categories MUST support `errors.Is`; structured details SHOULD support
`errors.As`. Wrapping MUST preserve the original cause when disclosure is safe.
Error strings MUST be bounded and secret-safe and MUST NOT be the public
classification contract.

Domains SHOULD distinguish invalid configuration, local rejection, conflict,
cancellation, deadline, unavailable dependency, permanent failure, partial
success, retryable failure, and unknown outcome when those states exist.
Transient does not automatically mean safe to retry. Backend-specific errors
MUST NOT cross a semantic abstraction unless the adapter contract explicitly
allows them. Aggregate errors MUST preserve item identity and relevant outcomes
within documented bounds.

## Concurrency, Callbacks, And Aliasing

Public documentation MUST state whether a value is immutable, single-owner, or
safe for concurrent use. Every goroutine, channel, timer, iterator, observer,
callback, and worker MUST have one visible owner and bounded lifetime.

Callbacks MUST document synchronization, re-entry, blocking, cancellation,
panic, and retention behavior. Packages MUST expose bounds and backpressure
instead of silently creating unbounded goroutines, queues, buffers, histories,
or cardinality. Test helpers SHOULD provide deterministic fake time, shutdown,
and concurrency seams without changing production global state.

## Security And Sensitive Data

Secure behavior MUST be the default when a generally useful safe default
exists. Proxy trust, credential placement, debug disclosure, weak algorithms,
unbounded input, and permissive fallback require explicit opt-in.

Secrets, credentials, payloads, tenant identifiers, and high-cardinality values
MUST have a documented treatment across errors, logs, traces, metrics,
fixtures, and snapshots. Validation and diagnostics SHOULD describe the failing
field or operation without echoing its value. Hostile-input parsing and
allocation MUST use explicit limits.

## Observability

Core packages SHOULD expose bounded domain observations or hooks instead of
importing telemetry implementations. Logging integrations SHOULD use
`*slog.Logger`. Direct OpenTelemetry adapters use the `otel` target; adapters
to Golib's telemetry contract use `telemetry`.

Packages MUST document who creates, owns, flushes, and shuts down providers,
exporters, loggers, and observers. Optional observability is expressed by not
constructing an adapter; an explicitly constructed adapter SHOULD reject
missing required providers rather than silently falling back to globals.
Metric and trace names SHOULD use a documented Golib namespace and stable
attribute vocabulary. Applications remain responsible for provider setup; a
package MUST NOT initialize global telemetry to obtain naming consistency.

## Adapter And Module Names

New optional integrations MUST use `<owner>/adapters/<target>` unless the domain
has a reviewed consumer term such as `providers` or `formats`. Target names
MUST omit `go` and owner branding. Examples include `kafka`, `queue`, `outbox`,
`otel`, `postgres`, `valkey`, `http`, `service`, and `money`.

Service lifecycle integrations MUST converge on `adapters/service`. AWS
adapters MUST identify the actual service or protocol, such as
`awssecretsmanager` or `mskiam`. Directory, module path, package identifier,
documentation title, and catalog label MUST be intentionally related, though a
package identifier MAY remain source-qualified to prevent import collisions.

Released inconsistent paths are compatibility constraints. A replacement MUST
be additive, migrate owned consumers, and retain the old path for the longer of
180 days and two published stable minor releases after the replacement is
public. Time alone cannot expire the interval without releases. Removal also
requires clean external-consumer evidence and an authorized next-major release.
Deprecated paths MUST continue to receive security and correctness fixes during
the interval and MUST NOT be silently retagged or rewritten.

Before each migration, the owner MUST inventory public tags, owned consumers,
observed external usage, documentation, the affected dependency closure, and
release ordering. Public code search is supplementary; no result does not prove
that external consumers are absent.

## Package Families And Selection

Every releasable module belongs to one primary family: Foundations, Service
edge, Protocols and descriptions, Persistence and durability, Resilience,
Observability, Integration and data movement, Domain utilities, or Tooling.
Families organize discovery and ownership; they do not create import layers.

Secondary capabilities use only this reviewed vocabulary:

- `administration-and-control`;
- `configuration`;
- `cryptography-and-secrets`;
- `data-encoding`;
- `database-and-storage`;
- `distributed-coordination`;
- `eventing-and-messaging`;
- `http-and-service-edge`;
- `identity-and-access`;
- `observability`;
- `protocols-and-schemas`;
- `resilience-and-admission`;
- `scheduling-and-orchestration`;
- `testing-and-conformance`; and
- `transport-and-networking`.

Backend, service, protocol, specification, and platform names belong in their
dedicated catalog fields instead of becoming one-off capability synonyms. A
new capability term requires a reviewed schema change.

A coherent public suite MAY omit a root package when one module owns several
peer entry packages. Its catalog record uses no default package identifier and
MUST identify the primary entry packages, a package-selection map, and an
executable quick start. `go-analysis` and `go-barcode` use this reviewed shape;
root-package absence alone is not an API defect.

The reviewed family boundaries classify RabbitMQ Queues, secret envelopes, and
wire formats as Integration and data movement; state machines as Domain
utilities; and settings and feature flags as Persistence and durability.
Tenancy remains Foundations. Owner-specific adapters inherit their owner's
primary family and record their target as a secondary capability unless a
reviewed consumer-selection reason requires another primary family.

Consumer selection follows responsibility:

- `config` loads startup configuration; `settings` owns runtime-mutable,
  versioned, persisted values.
- `state-machine` models deterministic in-process transitions; `workflow` owns
  durable history, activities, retry, compensation, timers, and recovery.
- `scheduler` owns time-based admission; `sequencer` owns dependency-ordered
  operations, attempts, and reconciliation.
- `semaphore` owns fixed weighted permits; `concurrency-limit` owns adaptive
  in-flight admission; `bulkhead` owns isolated capacity and bounded queues.
- `http-client` owns HTTP transport concerns; resilience modules own reusable
  retry, limit, breaker, bulkhead, and hedge policies. A composition selects one
  owner for each policy.
- `openapi` and `openrpc` each own their protocol document and runtime
  expression semantics. Similar schema, reference, and diff concepts do not
  create a shared runtime dependency without a proven semantic owner.
- `validation` owns application value rules; schema and protocol modules own
  serialized document validation.
- `queue` owns the portable worker abstraction; `rabbitmq-queues` owns AMQP
  0-9-1 queue semantics; `rabbitmq-streams` owns append-log, offset, and replay
  semantics.
- `search` owns indexes and queries; an application datastore remains
  authoritative; outbox and projection consumers move committed intent.
- Core packages expose bounded domain observations; direct `otel` adapters
  translate them; `telemetry` owns providers, exporters, propagation, flush,
  and shutdown.
- Identity account, session, authentication, authorization, and capability
  packages are Service edge. SCIM and WebAuthn wire contracts are Protocols and
  descriptions. Providers and persistence remain adapters of those owners.
- `localized` owns reusable localized text and fallback. A planned identity
  `i18n` package may own only identity-specific message keys and catalog
  mapping, and composes `localized` rather than duplicating it.
- `identity/reference` is a non-releasable engineering composition and evidence
  harness, not a consumer runtime or umbrella module.

The consumer catalog contains only installable libraries and adapters. The
engineering inventory additionally contains fixtures, examples, benchmarks,
interoperability harnesses, and internal tools.

## Test Helpers And Deterministic Seams

Test packages MAY provide fakes, fixtures, clocks, identifiers, recorders, and
bounded synchronization helpers. They MUST preserve the production contract,
avoid process-global mutation, expose deterministic ownership, and remain
clearly separated from runtime dependencies. A fake MUST NOT silently provide
stronger ordering, cancellation, or consistency than the real boundary unless
the difference is explicit.

## Compatibility, Deprecation, And Versioning

Each module follows independent semantic versioning. A repository containing
multiple modules MUST NOT imply synchronized releases. Public import paths,
package identifiers, lifecycle signatures, wire behavior, and documented
semantics remain compatibility constraints after v1.

Deprecations MUST identify the replacement and migration. Compatibility wrappers
MUST preserve observable semantics; when a wrapper would weaken or misstate the
contract, the old path remains an explicit exception until a safe major-version
migration exists.

Published compatibility sets record exact module versions, Go and platform
matrices, covered compositions, external services or protocols, fingerprints,
observation time, caveats, upgrade and rollback notes, and source identities.
Consumers MAY select other compatible versions.

Public set identifiers use
`golib-compat-v<schema>-<YYYYMMDD>.<sequence>`. Each public set records the
design-language version, exact tooling/documentation tag, content digests,
clean-consumer evidence, publication time, upgrade and rollback guidance, and
source and release-manifest references. An identifier is immutable and MUST NOT
be reused or mutated.

Before public module releases, a set MAY use
`draft-<YYYYMMDD>.<sequence>` with exact unreleased content identities. A draft
set is non-installable and MUST NOT be presented as a public known-good release
set.

## Explicit Application Composition

Applications compose packages in ordinary Go:

1. load configuration through the package that owns loading;
2. validate and construct caller-owned infrastructure clients;
3. construct domain packages and adapters with explicit dependencies;
4. compile immutable definitions before concurrent use;
5. start only the runtimes whose lifecycle the application transfers;
6. pass request or job contexts through every blocking operation;
7. drain intake before shutting down dependencies in reverse ownership order;
   and
8. flush or close only resources the application or object explicitly owns.

A composition MUST NOT apply retry, transactions, telemetry, acknowledgement,
or shutdown twice through hidden layers. Reference recipes use only public APIs
and remain examples or interoperability harnesses, never mandatory bootstrap
code.

## Documentation Contract

Every releasable module and independently releasable adapter MUST provide an
entry README that states or links to its purpose, non-goals, maturity, Go
version, installation path, quick start, package map, selection guidance,
construction, defaults, validation, ownership, concurrency, errors,
cancellation, shutdown, integrations, adapters, companions, security and
sensitive-data handling, compatibility, migration, performance, operations,
testing helpers, support, and changelog where applicable. It MUST also identify
supported platforms, backends, protocols, and specifications.

The entry point MUST link to API reference, executable examples, FAQ,
troubleshooting, license, security-reporting instructions, support, and
changelog material. A package MAY omit an irrelevant topic rather than publish
an empty section, but the omission MUST NOT hide an applicable consumer risk or
ownership boundary.

Each README MUST link back to a versioned ecosystem index and family guidance.
Planned modules MUST be visibly planned and MUST NOT present installation or
released behavior. Internal harnesses belong only in the engineering inventory.

## Version Binding

Published copies of this document record design-language version `1.0`, an
exact immutable `go-library-tools` tag, and a content digest. Default-branch
content is the latest development view and MUST NOT be cited as an immutable
compatibility contract.
