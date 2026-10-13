# Choose packages by the problem

[Documentation map](index.md)

Begin with ordinary Go and add a module only for a contract the application
needs. Names below identify owners; their exact installable paths, optional
modules and cohort versions are in the [catalog](packages.md).

| Primary family | Problem and likely owner | Alternative or exception |
| --- | --- | --- |
| Foundations | Explicit time seams (`go-clock`), configuration (`go-config`), identifiers, bounded synchronization | Use `time`, explicit structs and standard synchronization when sufficient. A clock does not own calendars or scheduling. |
| Service edge | Process lifecycle (`go-service`), routing (`go-router`), HTTP policy (`go-http-middleware`), authentication and authorization | A small `net/http` process need not adopt a service runtime. Credential validation and permission decisions are separate concerns. |
| Protocols and descriptions | JSON-RPC/OpenRPC, JSON:API, JSON Schema, OpenAPI, wire formats | Select the consumer interaction first; an API description does not generate or secure a running server by itself. |
| Persistence and durability | PostgreSQL (`go-postgres`), migrations (`go-migrations`), cache, queue, lease, idempotency, transactional outbox, workflow | Direct pgx or broker clients are appropriate when their backend contracts are exactly what the application wants. A cache is not a durable source of truth. |
| Resilience | HTTP client policy, retry, rate limiting, circuit breaking, admission/backpressure | Do not install every policy by default. Retry only operations whose side-effect semantics permit repetition. |
| Observability | Logging (`go-log`), telemetry (`go-telemetry`), correlation | Plain `log/slog` may suffice. Applications own providers, exporters and cardinality; optional adapters do not create them. |
| Integration and data movement | Filesystem (`go-filesystem`), tabular data (`go-tabular`), Kafka, CloudEvents, schema registry, ingestion | Filesystem access, decoding and application validation are distinct. Broker acceptance is not completed business processing. |
| Domain utilities | Calendars, dates/timezones, geography, localization, money, measurements, search | Prefer a domain-specific owner over extending a generic foundation. Check precision, locale, backend and specification caveats. |
| Tooling | Repository validation and release integrity (`go-library-tools`) | Tooling is not an application runtime; reference fixtures are not deployable service products. |

## Decisions that overlap

- **Queue versus RabbitMQ-specific owners:** use `go-queue` when the documented
  delivery/worker abstraction matches the use case. RabbitMQ queues and streams
  have distinct settlement, retention and replay models; select the dedicated
  owner when those semantics are required rather than treating backend names
  as interchangeable.
- **Scheduler versus lease versus workflow:** a scheduler decides when to
  attempt work; a lease arbitrates ownership for a bounded interval; a durable
  workflow records application progress/compensation. A singleton schedule
  does not guarantee exactly-once effects.
- **Idempotency versus outbox:** idempotency deduplicates one named application
  operation; outbox couples durable intent to the business transaction. Neither
  removes receiver deduplication or unknown external outcomes.
- **Authentication versus authorization:** identify the caller and credential
  validity first, then evaluate application permissions on the resolved tenant
  and resource. No shared container selects that order for you.
- **Service versus router versus middleware:** lifecycle owns process startup,
  drain and cleanup; routing chooses handlers; policy wraps requests. Each can
  be used independently.

Before adopting any stack, read [limitations](limitations.md), including planned
identity modules, released facades and package-current major migrations.

**Verkle Tree is experimental and not production-ready:** do not select it to
protect production integrity or confidentiality. Its stable API and historical
active/stable catalog labels do not qualify its cryptography. Read the
[specific security limitation and revisit condition](limitations.md) before
choosing it for an isolated research experiment.

The [design language](../design-language.md) is the authority for construction,
adapter naming and lifecycle terms; this page does not invent a new API.

Next: [protocol choice](api-protocols.md), [stacks](recommended-stacks.md),
[fair comparisons](comparisons/index.md).
