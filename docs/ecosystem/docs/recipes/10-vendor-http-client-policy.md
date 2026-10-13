# Vendor HTTP client policy stack

[Recipe index](index.md) · [Documentation map](../index.md)

This route consumes the published [`golib-compat-v1-20260912.2`](../../compatibility-sets.md)
recipe `10-vendor-http-client-policy`. The [complete source](https://github.com/faustbrian/go-http-client/tree/f400c235e3c94e16d3e38a43932501bb045017e5/examples/vendorcomposition) contains imports, configuration,
explicit wiring and executable assertions. It is a reference, not a production
runtime dependency. Required owner boundaries: **HTTP Client, Retry, Rate Limit, Circuit Breaker, Cache, Telemetry and explicit bulkhead**. Optional adapters
are selected explicitly; exact versions and dependencies are in the source
module and the frozen [catalog](../packages.md), not a second version table.

## Construction and work lifecycle

Configure one total operation budget and caller-owned transport/providers. Validate replayability and policy limits before client construction.

Cache and logical rate admission surround breaker and bounded retry; bulkhead admission applies per physical attempt. Observation and transport wrappers preserve those counting boundaries.

## Failure, ownership and cleanup

Caller owns the client and response body. Close bodies and idle transports under their actual owner. Do not infer failed side effects from connection loss; retry only eligible operations.

## Execute the reference

Follow the [clean checkout and version-rebase procedure](index.md#run-a-reference)
with these exact inputs:

| Input | Value |
| --- | --- |
| Repository | `https://github.com/faustbrian/go-http-client.git` |
| Source | `f400c235e3c94e16d3e38a43932501bb045017e5` |
| Module directory | `.` |

From that prepared module directory:

```sh
go test -mod=readonly -count=1 ./examples/vendorcomposition
```

## Evidence and deployment boundary

TLS httptest contract and policy ordering; not a measured production latency recommendation.

For Kubernetes, retain the [operator-owned setup and shutdown](../operations/kubernetes.md)
instead of deploying this fixture as a product. Keep actual backend versions,
application payloads and recovery policy in the deployment's own runbook.

Next: [supported stacks](../recommended-stacks.md), [operations](../operations/index.md),
[limitations](../limitations.md).
