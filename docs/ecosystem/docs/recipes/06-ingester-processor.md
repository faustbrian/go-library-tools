# Ingester and processor

[Recipe index](index.md) · [Documentation map](../index.md)

This route consumes the published [`golib-compat-v1-20260912.2`](../../compatibility-sets.md)
recipe `06-ingester-processor`. The [complete source](https://github.com/faustbrian/go-service/tree/61535b800e87ae68451a7134810cedfe0cd985f6/integration/reference-recipes) contains imports, configuration,
explicit wiring and executable assertions. It is a reference, not a production
runtime dependency. Required owner boundaries: **Service, Correlation, Queue and adapters/service**. Optional adapters
are selected explicitly; exact versions and dependencies are in the source
module and the frozen [catalog](../packages.md), not a second version table.

## Construction and work lifecycle

Validate bounds, construct caller-owned in-memory transport and correlated handoff, then start the processor/ingester roles.

Accepted handoff preserves correlation; drain rejects late intake before admitted work finishes. Acknowledge after successful application processing.

## Failure, ownership and cleanup

Drain and shutdown differ: first let accepted work complete, then join the processor before transport release. Decode success is not acknowledgement.

## Execute the reference

Follow the [clean checkout and version-rebase procedure](index.md#run-a-reference)
with these exact inputs:

| Input | Value |
| --- | --- |
| Repository | `https://github.com/faustbrian/go-service.git` |
| Source | `61535b800e87ae68451a7134810cedfe0cd985f6` |
| Module directory | `integration/reference-recipes` |

From that prepared module directory:

```sh
go test -mod=readonly -run '^TestIngesterProcessorRecipeHandsOffAcknowledgesDrainsAndShutsDown$' -count=1 .
```

## Evidence and deployment boundary

Bounded in-memory transport, not broker durability. No remote storage interoperability is implied.

For Kubernetes, retain the [operator-owned setup and shutdown](../operations/kubernetes.md)
instead of deploying this fixture as a product. Keep actual backend versions,
application payloads and recovery policy in the deployment's own runbook.

Next: [supported stacks](../recommended-stacks.md), [operations](../operations/index.md),
[limitations](../limitations.md).
