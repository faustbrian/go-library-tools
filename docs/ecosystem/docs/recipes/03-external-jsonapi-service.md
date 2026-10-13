# External JSON:API service

[Recipe index](index.md) · [Documentation map](../index.md)

This route consumes the published [`golib-compat-v1-20260912.2`](../../compatibility-sets.md)
recipe `03-external-jsonapi-service`. The [complete source](https://github.com/faustbrian/go-http-middleware/tree/93e41da128a0a7696e89769c2fd688a74f552678/integration/siblings) contains imports, configuration,
explicit wiring and executable assertions. It is a reference, not a production
runtime dependency. Required owner boundaries: **HTTP Middleware, Service, HTTP Client, Log, Authentication, Authorization and JSON:API**. Optional adapters
are selected explicitly; exact versions and dependencies are in the source
module and the frozen [catalog](../packages.md), not a second version table.

## Construction and work lifecycle

Application constructs dependencies and registers the external representation and explicit policy order.

Resolve identity/permissions, validate the resource contract, dispatch application logic and map its outcome without leaking transport models into domain code.

## Failure, ownership and cleanup

Application owns client/response bodies, handlers and service cleanup; no protocol library acquires database or broker ownership.

## Execute the reference

Follow the [clean checkout and version-rebase procedure](index.md#run-a-reference)
with these exact inputs:

| Input | Value |
| --- | --- |
| Repository | `https://github.com/faustbrian/go-http-middleware.git` |
| Source | `93e41da128a0a7696e89769c2fd688a74f552678` |
| Module directory | `integration/siblings` |

From that prepared module directory:

```sh
go test -mod=readonly -run '^TestServiceComposesHTTPClientLoggingAuthenticationAuthorizationAndJSONAPI$' -count=1 .
```

## Evidence and deployment boundary

httptest JSON:API composition. The separate handwritten OpenAPI-named handler is not generated-contract evidence.

For Kubernetes, retain the [operator-owned setup and shutdown](../operations/kubernetes.md)
instead of deploying this fixture as a product. Keep actual backend versions,
application payloads and recovery policy in the deployment's own runbook.

Next: [supported stacks](../recommended-stacks.md), [operations](../operations/index.md),
[limitations](../limitations.md).
