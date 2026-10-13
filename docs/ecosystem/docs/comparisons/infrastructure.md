# Infrastructure layers versus direct clients

[Comparisons](index.md) · [Documentation map](../index.md)

| Concern | Direct client or alternative | What the owned layer must justify |
| --- | --- | --- |
| PostgreSQL | [pgx](https://github.com/jackc/pgx), standard database/sql, [sqlc](https://docs.sqlc.dev/) generated queries | Golib Postgres is not an Eloquent replacement. Compare explicit pool/query/transaction/observation seams with direct pgx use; sqlc owns generated query typing, not process lifecycle. |
| Migrations | [golang-migrate](https://github.com/golang-migrate/migrate), [Goose](https://github.com/pressly/goose) | Compare source formats, history/locking, dirty/partial failure and rollback semantics. A shared SQL backend does not make history interchangeable. |
| Queue | Broker-native clients, [Asynq](https://github.com/hibiken/asynq) | Compare delivery identity, retry scheduling, acknowledgement, reclamation, dead letters and backend durability. Asynq's Redis task model is not arbitrary RabbitMQ/Kafka interoperability. |
| Cache | [valkey-go](https://github.com/valkey-io/valkey-go) or another backend-native client | Compare key/value/TTL semantics, atomic operations and failure policy. A generic cache abstraction cannot silently supply unsupported backend guarantees. |
| HTTP client | [net/http](https://pkg.go.dev/net/http), direct transport wrappers | Golib HTTP Client/policies are useful only for the budgets, errors, replayability, telemetry and cleanup your application requires. Plain http.Client may suffice. |

Owned layers can stabilize a consumed semantic interface or isolate optional
backends. They also introduce policy choices and an extra upgrade boundary.
Avoid a wrapper that obscures transactions, backend error semantics or resource
ownership without delivering a needed contract. Compare realistic behavior,
not a no-op wrapper against a fully configured direct client.

Queue/control-plane management and broker settlement are adjacent capabilities.
Cache, idempotency, lease and outbox have different consistency promises. A
successful publish can still precede a failed acknowledgement; an ambiguous
database commit remains unknown until application reconciliation. Examine
those outcomes before choosing abstraction ergonomics.

Current Golib majors can change owned metadata tables and constructors. Follow
[migration](../migration/index.md) and [versioning](../versioning.md) instead of
combining frozen recipe code with latest dependencies. These comparisons are
adoption considerations, not a benchmark or certification of interoperability.

Next: [stacks](../recommended-stacks.md), [operations](../operations/index.md),
[performance interpretation](performance.md).
