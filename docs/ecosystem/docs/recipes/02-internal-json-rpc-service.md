# Internal JSON-RPC service

[Recipe index](index.md) · [Documentation map](../index.md)

This route consumes the published [`golib-compat-v1-20260912.2`](../../compatibility-sets.md)
recipe `02-internal-json-rpc-service`. The [complete source](https://github.com/faustbrian/go-service/tree/61535b800e87ae68451a7134810cedfe0cd985f6/integration/reference-http) contains imports, configuration,
explicit wiring and executable assertions. It is a reference, not a production
runtime dependency. Required owner boundaries: **Service, Router, JSON-RPC, Config and selected identity/policy/observation modules**. Optional adapters
are selected explicitly; exact versions and dependencies are in the source
module and the frozen [catalog](../packages.md), not a second version table.

## Construction and work lifecycle

Validate caller-owned configuration and keys; register handlers and explicit policies before transferring listener ownership.

Outer observation encloses signature/digest, capability/tenant, authentication, authorization, bounded decoding/validation, use case and audit boundaries. Request IDs do not confer idempotency.

## Failure, ownership and cleanup

Invalid construction fails before service start. Unsigned, denied and invalid requests are distinguishable. Cancellation and restart use fresh listeners; preserve unknown external outcomes.

## Execute the reference

Follow the [clean checkout and version-rebase procedure](index.md#run-a-reference)
with these exact inputs:

| Input | Value |
| --- | --- |
| Repository | `https://github.com/faustbrian/go-service.git` |
| Source | `61535b800e87ae68451a7134810cedfe0cd985f6` |
| Module directory | `integration/reference-http` |

From that prepared module directory:

```sh
go test -mod=readonly -run '^(TestReferenceHTTPServiceLifecycleAndRequestContract|TestReferenceHTTPRejectsInvalidConstructionAndPreparation|TestReferenceHTTPServiceCanRestartWithFreshListeners)$' -count=1 .
```

## Evidence and deployment boundary

Loopback and in-memory audit/telemetry; no production or managed-service claim.

For Kubernetes, retain the [operator-owned setup and shutdown](../operations/kubernetes.md)
instead of deploying this fixture as a product. Keep actual backend versions,
application payloads and recovery policy in the deployment's own runbook.

Next: [supported stacks](../recommended-stacks.md), [operations](../operations/index.md),
[limitations](../limitations.md).
