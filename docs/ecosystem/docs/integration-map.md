# Integration and ownership map

[Documentation map](index.md)

This is a map of semantic owners, not permission for every package to import
every other package. Exact optional module paths and dependency declarations
remain in the [catalog](packages.md). Applications wire the arrows.

| Concern | Owner | Application or adapter boundary |
| --- | --- | --- |
| Process lifecycle and probes | Service | Configuration, admission policy and dependency readiness remain application-owned. |
| Routing and HTTP policy | HTTP router, HTTP middleware | Accept handlers and explicit order; do not initialize database/broker infrastructure. |
| Credentials and permissions | Authentication, Authorization | Caller supplies keys/providers, tenant/resource context and permission policy. |
| Protocol parsing and descriptions | JSON-RPC, JSON:API, JSON Schema, OpenAPI, OpenRPC | Use-case execution and authorization are not supplied by a document parser. |
| PostgreSQL | Postgres | Explicit pool acquisition; callers own transactions, schema and pool lifetime. |
| Schema migration | Migrations | Application controls deployment jobs, history ownership and irreversible-change gates. |
| Cache | Cache and selected backend adapter | Cache invalidation and authoritative data remain application policy. |
| Queue and broker work | Queue, RabbitMQ Queues/Streams, Kafka | Acknowledge after business success; caller owns replay and poison-message policy. |
| Lease, idempotency and outbox | Their separate semantic owners | Lease expiry, logical operation identity and transaction atomicity are not interchangeable. |
| Scheduling and durable workflow | Scheduler, Workflow | Schedule triggers attempts; application owns effects and compensation semantics. |
| Logging, correlation and telemetry | Log, Correlation, Telemetry | Supply slog handlers/providers explicitly; caller owns flush, shutdown and labels. |
| Configuration and secrets | Config and explicitly selected provider | Application acquires/refreshes secrets and owns credentials; planned Secret Store is not installable. |
| Filesystems and tabular formats | Filesystem, Tabular | Storage acquisition differs from decoding, limits and application validation. |
| Wire formats, time and domain values | Wire, Clock, Calendar, Temporal, Geography, Localization and cataloged domain owners | Domain-specific guarantees remain separate; no global clock or locale registry. |
| HTTP client and policies | HTTP Client plus Retry/Rate Limit/Circuit Breaker owners | Caller defines budgets, side-effect-safe retry, ordering and cleanup. |
| Webhook delivery and signatures | Webhook, HTTP Signature | Application owns receivers, deduplication, signing keys and replay windows. |

An adapter to `otel` targets OpenTelemetry; an adapter to `telemetry` targets
Golib's telemetry contract. Neither initializes a global provider. Service
adapters target `adapters/service`; caller-owned pools and clients are closed
only by their documented owner. Backend-specific persistence may remain
domain-owned when that backend is part of the installed capability; do not
move it solely for visual naming consistency.

Next: [architecture](architecture.md), [stacks](recommended-stacks.md),
[limitations](limitations.md).
