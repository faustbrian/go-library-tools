# Incremental Laravel/PHP migration

[Documentation index](../index.md) · [Ecosystem entry](../../README.md)

Migrate one owned operation, endpoint, worker or schedule at a time. Golib
provides independent packages, not a Laravel replacement framework. Keep PHP
and Go running together until the observable response, persistence, failure
and operator contracts of that slice have been compared.

## Choose a version boundary first

This route uses design-language **1.0** and the immutable
[`golib-compat-v1-20260912.2` cohort](../../compatibility-sets.md) for package
selection. Follow its exact module versions rather than upgrading every package
at once. The linked v1 package docs describe those releases; newer package
majors can have different imports, types and operational requirements. A newer
release or the maintained current consumer is not membership in that frozen
cohort. Record each intentional replacement and prove your application's direct
composition before adoption.

## Translate responsibilities, not facade syntax

| Laravel/PHP concern | Go owner and starting point | Behavior to retain explicitly |
| --- | --- | --- |
| Container, providers, routes and middleware | Application composition root; [service](https://github.com/faustbrian/go-service/blob/v1.1.0/docs/migration.md), [router](https://github.com/faustbrian/go-router/blob/v1.0.0/docs/migration.md), [middleware](https://github.com/faustbrian/go-http-middleware/blob/v1.0.0/docs/migration.md) | Construction, dependency wiring, middleware order, startup and reverse cleanup; no automatic container or facade initialization. |
| Eloquent and query builder | Application repositories/SQL; [Postgres](https://github.com/faustbrian/go-postgres/blob/v1.1.0/docs/migration.md) and [API Query](https://github.com/faustbrian/go-api-query/blob/v1.1.0/docs/migration-laravel-cline.md) | Tenant/global scopes, authorization, casts/nulls, eager loading, pagination, ordering and transaction boundaries. Postgres is a pgx pool/lifecycle layer, not an ORM. sqlc is an optional independent query generator, not a Golib runtime. |
| Horizon and jobs | [Queue](https://github.com/faustbrian/go-queue/blob/v1.1.0/docs/horizon-migration.md), [queue control plane](https://github.com/faustbrian/go-queue-control-plane/blob/v1.1.0/docs/horizon-migration.md), Kubernetes and telemetry | Worker delivery/settlement, administrative authorization/audit, process supervision and dashboards have separate owners. No automatic PHP job decoding or Horizon Redis-schema compatibility. |
| Cache facade | [Cache](https://github.com/faustbrian/go-cache/blob/v1.0.0/docs/migration.md) | Namespace, key/codec version, TTL, invalidation and failure policy. Sharing a Valkey server does not make serialized values compatible. |
| Scheduler and singleton work | [Scheduler](https://github.com/faustbrian/go-scheduler/blob/v1.1.0/docs/laravel-migration.md) and [Lease](https://github.com/faustbrian/go-lease/blob/v1.1.0/docs/laravel-migration.md) | Stable task names, time zones, overlap/leader ownership, missed runs, fencing and idempotent business effects. |
| Filesystem facade | [Filesystem](https://github.com/faustbrian/go-filesystem/blob/v1.1.0/docs/README.md) | Path confinement, ownership, streaming limits, errors and the selected storage backend. |
| HTTP client and retry helpers | [HTTP Client](https://github.com/faustbrian/go-http-client/blob/v1.1.0/docs/migration.md) | Deadlines, authentication, retry eligibility and total budgets; ambiguous non-idempotent operations require reconciliation. |
| Logging and tracing | [Log](https://github.com/faustbrian/go-log/blob/v1.0.0/docs/migration.md), [Telemetry](https://github.com/faustbrian/go-telemetry/blob/v1.2.0/docs/README.md) | Redaction, correlation, sampling and cardinality. Telemetry does not supply a stable log exporter/provider. |
| Validation, authentication and authorization | [Validation](https://github.com/faustbrian/go-validation/blob/v1.1.0/docs/migration.md), [Authentication](https://github.com/faustbrian/go-authentication/blob/v1.2.0/docs/migration.md), [Authorization](https://github.com/faustbrian/go-authorization/blob/v1.1.0/docs/operations.md) | Request/response failures, identity and tenant boundaries, policy order and deny-by-default behavior; package adoption does not migrate sessions or policies. |
| Carbon and dates | [Calendar](https://github.com/faustbrian/go-calendar/blob/v1.1.0/docs/carbon-migration.md) | Time zone, DST folds/gaps, calendar arithmetic and wire-format semantics. |
| Configuration and secrets | [Config](https://github.com/faustbrian/go-config/blob/v1.1.0/docs/migration.md) | Explicit source order, typed validation and immutable candidates. Platform-delivered environment/files are supported; no native Infisical SDK adapter or automatic hot reload is implied. |

Keep Laravel-owned behavior that has no selected Go equivalent in PHP, or give
it an explicit application/platform owner. Do not infer parity from a similarly
named helper. Compare the source application's versioned
[Horizon](https://laravel.com/framework/docs/12.x/horizon) and
[queue contracts](https://laravel.com/framework/docs/12.x/queues), including its
configured attempts, timeouts and supervisor behavior; those links describe
Laravel 12, not every Laravel installation.

## A reversible sequence

1. **Characterize one operation.** Capture authorized and denied responses,
   tenant filtering, validation, null/time/decimal handling, database writes,
   retryable failures and unknown outcomes. Record its callers, dependencies,
   traffic switch and rollback trigger.
2. **Keep one writer.** Begin with a read-only endpoint or an isolated queue.
   Sharing PostgreSQL is possible only with compatible SQL/schema semantics and
   a named write owner. Do not introduce unsynchronized PHP/Go dual writes.
3. **Compare before switching.** Use a production-like disposable database and
   bounded payload corpus. Shadow only read-only behavior; do not replay real
   side effects to compare implementations. Version deliberate differences.
4. **Canary the boundary.** Route a bounded traffic partition or new jobs to
   Go, keeping PHP available. Use an explicit payload schema, stable business
   identity and idempotency policy. Horizon and Go queues keep separate ownership
   until a converter and their delivery/settlement semantics are proven.
5. **Transfer ownership.** Stop the old producer/scheduler, drain accepted work,
   reconcile pending and failed records, then enable the new owner. Retain old
   schemas/readers and failure records for the rollback/audit window. Administrative
   command admission is not proof that a worker enforced it.
6. **Retire only after the window.** Remove old code, queues and columns only
   when no supported caller or old binary needs them. Treat replay, purge and
   historical repair as separate authorized operations.

Traffic rollback is viable while PHP can still understand the current schema
and payloads. Stop new Go admission, drain/reconcile its accepted work, restore
the previous traffic/producer owner and confirm that operation end to end.
Do not send new-format jobs to old handlers, reset offsets blindly, or assume
routing back undoes committed database or external effects.

## Adopt the existing schema without replaying Laravel history

Keep schema-change ownership singular even while both runtimes serve traffic.
Inventory grants, ownership, RLS and other security properties in addition to
schema objects. Use expand, compatible code, bounded backfill, ownership switch,
then contract; a code rollback stops being safe once contract removes old
requirements. Laravel's [migration documentation](https://laravel.com/framework/docs/12.x/migrations)
is a separate history contract, not a format the Go runner imports.

There are two materially different adoption paths:

- **Frozen cohort: Migrations v1.1.0.** Its
  [baseline runbook](https://github.com/faustbrian/go-migrations/blob/v1.1.0/docs/laravel-baseline.md)
  records a reviewed baseline in `public.go_schema_migrations` without reading,
  renaming or modifying Laravel's `migrations` history.
- **Newer release: Migrations /v3 v3.0.1.** Its
  [baseline runbook](https://github.com/faustbrian/go-migrations/blob/v3.0.1/docs/laravel-baseline.md)
  owns `public.migrations`. Quiesce schema changes, manually relocate Laravel
  history and configure any remaining Laravel tooling for the relocated table
  before fingerprint review. V3 does not perform that relocation. Follow its
  [ledger-upgrade and recovery guidance](https://github.com/faustbrian/go-migrations/blob/v3.0.1/docs/operations.md)
  before starting a v3 runner; never merge PHP and Go histories.

For either path, preserve a consistent pre-adoption backup, review the selected
catalog fingerprint and its omissions, choose a baseline below future Go
migrations, inspect plan/status and run baseline once in a dedicated deployment
Job. Fingerprinting is not a full schema/security audit. A baseline is not
executable migration SQL and cannot be crossed by ordinary rollback. On an
uncertain commit or dirty migration, inspect ledger and database effects before
retrying. A code rollback is not a database restore.

The newer [Postgres /v2 migration guide](https://github.com/faustbrian/go-postgres/blob/v2.0.0/docs/migration.md)
also requires explicit application-owned DSN resolution. Do not transplant v1
examples into a newer major and assume their construction or types still match.

## Next steps

Choose modules in the [consumer catalog](../../catalog-consumer.md), inspect
[compatibility and rollback](../../compatibility-sets.md), then assign infrastructure
owners using the [Kubernetes operator route](../operations/kubernetes.md).
