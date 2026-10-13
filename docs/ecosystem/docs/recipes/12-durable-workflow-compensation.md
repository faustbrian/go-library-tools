# Durable workflow with compensation

[Recipe index](index.md) · [Documentation map](../index.md)

This route consumes the published [`golib-compat-v1-20260912.2`](../../compatibility-sets.md)
recipe `12-durable-workflow-compensation`. The [complete source](https://github.com/faustbrian/go-workflow/tree/7066c265edb25a30984de304859110a641812051) contains imports, configuration,
explicit wiring and executable assertions. It is a reference, not a production
runtime dependency. Required owner boundaries: **Workflow and its PostgreSQL persistence module**. Optional adapters
are selected explicitly; exact versions and dependencies are in the source
module and the frozen [catalog](../packages.md), not a second version table.

## Construction and work lifecycle

Application defines activities and compensations, supplies store and worker policy, and acquires persistence explicitly.

Persist progress and failures; restart through the public durable state machine. Compensation has its own application semantics and may require operator resolution.

## Failure, ownership and cleanup

Retry only eligible activity outcomes; ambiguous effects are not erased by a compensation label. Stop admission, drain/join workers and close persistence last.

## Execute the reference

Follow the [clean checkout and version-rebase procedure](index.md#run-a-reference)
with these exact inputs:

| Input | Value |
| --- | --- |
| Repository | `https://github.com/faustbrian/go-workflow.git` |
| Source | `7066c265edb25a30984de304859110a641812051` |
| Module directory | `.` |

From that prepared module directory:

```sh
go test -mod=readonly -run '^Example_durableCompensationRecovery$' -count=1 .
go test -tags=integration -mod=readonly -run '^TestPostgreSQLDurableCompensationRecipeSurvivesRestartAndDrainsWorker$' -count=1 ./postgres
```

The PostgreSQL selector needs a task-owned database bound through
`WORKFLOW_POSTGRES_URL`; the pure example alone does not demonstrate restart
durability. Read the versioned persistence fixture preparation before execution.

## Evidence and deployment boundary

Pure example plus PostgreSQL 18.4 restart/compensation execution; not universal reversal of external effects.

For Kubernetes, retain the [operator-owned setup and shutdown](../operations/kubernetes.md)
instead of deploying this fixture as a product. Keep actual backend versions,
application payloads and recovery policy in the deployment's own runbook.

Next: [supported stacks](../recommended-stacks.md), [operations](../operations/index.md),
[limitations](../limitations.md).
