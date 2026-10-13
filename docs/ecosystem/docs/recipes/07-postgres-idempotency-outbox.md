# PostgreSQL transaction, migrations, idempotency and outbox

[Recipe index](index.md) · [Documentation map](../index.md)

This route consumes the published [`golib-compat-v1-20260912.2`](../../compatibility-sets.md)
recipe `07-postgres-idempotency-outbox`. The [complete source](https://github.com/faustbrian/go-service/tree/61535b800e87ae68451a7134810cedfe0cd985f6/integration/reference-durability) contains imports, configuration,
explicit wiring and executable assertions. It is a reference, not a production
runtime dependency. Required owner boundaries: **Postgres, Migrations, Idempotency, Transactional Outbox and Queue**. Optional adapters
are selected explicitly; exact versions and dependencies are in the source
module and the frozen [catalog](../packages.md), not a second version table.

## Construction and work lifecycle

Use separate migration ownership; acquire pool and Valkey clients explicitly, complete migrations before application work.

Business state, idempotency completion and outbox intent share one caller-visible transaction. Publish only committed intent; receiver settles after processing.

## Failure, ownership and cleanup

Rollback publishes nothing. Replay returns prior results without duplicate business/outbox effects. Ambiguous PostgreSQL commit is unknown and requires idempotent reconciliation, not blind retry. Drain/join before backend cleanup.

## Execute the reference

Follow the [clean checkout and version-rebase procedure](index.md#run-a-reference)
with these exact inputs:

| Input | Value |
| --- | --- |
| Repository | `https://github.com/faustbrian/go-service.git` |
| Source | `61535b800e87ae68451a7134810cedfe0cd985f6` |
| Module directory | `integration/reference-durability` |

From that prepared module directory:

```sh
./check-durability.sh
./check-recovery.sh
```

The scripts provision/stop task-owned containers; the recovery supplement
terminates/replaces task-owned processes and dependencies. Execute only in the
owner-approved controlled/hosted environment after reading the complete script
and resource contract, never against shared services.

## Evidence and deployment boundary

Controlled PostgreSQL/Valkey durability and process-loss supplement. No backup/restore, cross-region or ambiguous-commit resolution proof.

For Kubernetes, retain the [operator-owned setup and shutdown](../operations/kubernetes.md)
instead of deploying this fixture as a product. Keep actual backend versions,
application payloads and recovery policy in the deployment's own runbook.

Next: [supported stacks](../recommended-stacks.md), [operations](../operations/index.md),
[limitations](../limitations.md).
