# Track, Postal and Location service roles

[Recipe index](index.md) · [Documentation map](../index.md)

This route consumes the published [`golib-compat-v1-20260912.2`](../../compatibility-sets.md)
recipe `14-track-postal-location-roles`. The [complete source](https://github.com/faustbrian/go-service/tree/61535b800e87ae68451a7134810cedfe0cd985f6/integration/adoption) contains imports, configuration,
explicit wiring and executable assertions. It is a reference, not a production
runtime dependency. Required owner boundaries: **Service and explicit role-specific dependencies in the adoption fixture**. Optional adapters
are selected explicitly; exact versions and dependencies are in the source
module and the frozen [catalog](../packages.md), not a second version table.

## Construction and work lifecycle

Select only the role needed; construct configuration, migrations, queue, schedule, PostgreSQL, Kafka, telemetry and resilience as applicable.

API/RPC, workers, schedulers, migrations and activation retain separate plans. Unselected roles do not initialize unrelated facilities; application router/domain behavior remains outside the fixture.

## Failure, ownership and cleanup

Readiness, deadlines, correlation, policy closure, drain and reverse shutdown remain caller-visible. No shared registry or business bootstrap is introduced.

## Execute the reference

Follow the [clean checkout and version-rebase procedure](index.md#run-a-reference)
with these exact inputs:

| Input | Value |
| --- | --- |
| Repository | `https://github.com/faustbrian/go-service.git` |
| Source | `61535b800e87ae68451a7134810cedfe0cd985f6` |
| Module directory | `integration/adoption` |

From that prepared module directory:

```sh
go test -mod=readonly -count=1 ./...
```

## Evidence and deployment boundary

Loopback/in-memory public construction fixtures; not the real applications business behavior or production deployment.

For Kubernetes, retain the [operator-owned setup and shutdown](../operations/kubernetes.md)
instead of deploying this fixture as a product. Keep actual backend versions,
application payloads and recovery policy in the deployment's own runbook.

Next: [supported stacks](../recommended-stacks.md), [operations](../operations/index.md),
[limitations](../limitations.md).
