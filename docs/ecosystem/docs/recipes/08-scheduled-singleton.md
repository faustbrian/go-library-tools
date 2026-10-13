# Scheduled singleton work

[Recipe index](index.md) · [Documentation map](../index.md)

This route consumes the published [`golib-compat-v1-20260912.2`](../../compatibility-sets.md)
recipe `08-scheduled-singleton`. The [complete source](https://github.com/faustbrian/go-scheduler/tree/84492d44c8937f6a5dbe3c1d260f988c51799e14/examples/distributed-singleton) contains imports, configuration,
explicit wiring and executable assertions. It is a reference, not a production
runtime dependency. Required owner boundaries: **Scheduler, Lease backends and explicit dispatch/queue owner**. Optional adapters
are selected explicitly; exact versions and dependencies are in the source
module and the frozen [catalog](../packages.md), not a second version table.

## Construction and work lifecycle

Validate schedule/queue and acquire selected PostgreSQL or Valkey lease store; prepare the backend schema where required.

Two competing runners admit one fenced dispatch. The schedule triggers an attempt; the lease is not exactly-once business-effect proof.

## Failure, ownership and cleanup

Non-owner never dispatches. Cancellation stops admission; drain accepted work and join runners before closing the store. Application owns missed-run and retry/idempotency policy.

## Execute the reference

Follow the [clean checkout and version-rebase procedure](index.md#run-a-reference)
with these exact inputs:

| Input | Value |
| --- | --- |
| Repository | `https://github.com/faustbrian/go-scheduler.git` |
| Source | `84492d44c8937f6a5dbe3c1d260f988c51799e14` |
| Module directory | `.` |

From that prepared module directory:

```sh
go test -mod=readonly -run '^(TestRunSingletonDispatchesOnceAndDrainsBothReplicas|TestRunWithStoreClosesCallerResourceAfterFailure)$' -count=1 ./examples/distributed-singleton
```

The unit selectors alone do not cover the external singleton claim. Also run
`examples/distributed-singleton` once per selected PostgreSQL/Valkey backend
using the source configuration contract (`SCHEDULER_LEASE_BACKEND`) and
PostgreSQL schema preparation. Backend URLs/credentials must be supplied by the
controlled environment; do not put real values in this guide.

## Evidence and deployment boundary

Unit selectors plus separate PostgreSQL 18.4 and Valkey 9.1.0 executions; Valkey noeviction durability assumptions are explicit.

For Kubernetes, retain the [operator-owned setup and shutdown](../operations/kubernetes.md)
instead of deploying this fixture as a product. Keep actual backend versions,
application payloads and recovery policy in the deployment's own runbook.

Next: [supported stacks](../recommended-stacks.md), [operations](../operations/index.md),
[limitations](../limitations.md).
