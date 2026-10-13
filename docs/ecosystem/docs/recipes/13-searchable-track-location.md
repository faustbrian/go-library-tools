# Searchable Track/Location projection

[Recipe index](index.md) · [Documentation map](../index.md)

This route consumes the published [`golib-compat-v1-20260912.2`](../../compatibility-sets.md)
recipe `13-searchable-track-location`. The [complete source](https://github.com/faustbrian/go-search/tree/a93ac8817a9a172ffe552f3a4c1694bb74bd7e9b/adapters/opensearch) contains imports, configuration,
explicit wiring and executable assertions. It is a reference, not a production
runtime dependency. Required owner boundaries: **Search and adapters/opensearch**. Optional adapters
are selected explicitly; exact versions and dependencies are in the source
module and the frozen [catalog](../packages.md), not a second version table.

## Construction and work lifecycle

Application constructs the projection event, selected client/transport, credentials, tenant/query policy and bounds.

Write the representative projection through the public adapter, then query and interpret typed results through the same boundary.

## Failure, ownership and cleanup

Application owns projection/source-of-truth reconciliation and client cleanup. Preserve transport and per-item/partial outcomes rather than claiming a successful batch from one success.

## Execute the reference

Follow the [clean checkout and version-rebase procedure](index.md#run-a-reference)
with these exact inputs:

| Input | Value |
| --- | --- |
| Repository | `https://github.com/faustbrian/go-search.git` |
| Source | `a93ac8817a9a172ffe552f3a4c1694bb74bd7e9b` |
| Module directory | `adapters/opensearch` |

From that prepared module directory:

```sh
go test -mod=readonly -run '^(TestTrackLocationProjectionRecipe|ExampleClient_trackLocationProjection)$' -count=1 .
```

## Evidence and deployment boundary

Bounded HTTP transport seam, not live OpenSearch interoperability or Track/Location business behavior.

For Kubernetes, retain the [operator-owned setup and shutdown](../operations/kubernetes.md)
instead of deploying this fixture as a product. Keep actual backend versions,
application payloads and recovery policy in the deployment's own runbook.

Next: [supported stacks](../recommended-stacks.md), [operations](../operations/index.md),
[limitations](../limitations.md).
