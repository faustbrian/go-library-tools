# Queue producer and worker

[Recipe index](index.md) · [Documentation map](../index.md)

This route consumes the published [`golib-compat-v1-20260912.2`](../../compatibility-sets.md)
recipe `05-queue-producer-worker`. The [complete source](https://github.com/faustbrian/go-service/tree/61535b800e87ae68451a7134810cedfe0cd985f6/integration/reference-durability) contains imports, configuration,
explicit wiring and executable assertions. It is a reference, not a production
runtime dependency. Required owner boundaries: **Queue, its service adapter, Postgres, Migrations, Idempotency and Transactional Outbox**. Optional adapters
are selected explicitly; exact versions and dependencies are in the source
module and the frozen [catalog](../packages.md), not a second version table.

## Construction and work lifecycle

Caller supplies endpoints; acquire PostgreSQL/Valkey resources and migrate before constructing the durable producer and worker.

Business/idempotency/outbox commit makes work publishable. Publish to Valkey; process before acknowledging. Replacement workers reclaim unacknowledged work.

## Failure, ownership and cleanup

Stop new intake, drain admitted handlers, acknowledge only completed processing, join workers and close clients/pools in reverse order. Rollback emits nothing; ambiguous commit/settlement needs reconciliation.

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
```

The scripts provision/stop task-owned containers; the recovery supplement
terminates/replaces task-owned processes and dependencies. Execute only in the
owner-approved controlled/hosted environment after reading the complete script
and resource contract, never against shared services.

## Evidence and deployment boundary

PostgreSQL 18.4 and Valkey 9.1.0 controlled external fixtures; no managed failover or production claim.

For Kubernetes, retain the [operator-owned setup and shutdown](../operations/kubernetes.md)
instead of deploying this fixture as a product. Keep actual backend versions,
application payloads and recovery policy in the deployment's own runbook.

Next: [supported stacks](../recommended-stacks.md), [operations](../operations/index.md),
[limitations](../limitations.md).
