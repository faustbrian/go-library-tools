# Authenticated and authorized service

[Recipe index](index.md) · [Documentation map](../index.md)

This route consumes the published [`golib-compat-v1-20260912.2`](../../compatibility-sets.md)
recipe `04-authenticated-authorized-service`. The [complete source](https://github.com/faustbrian/go-service/tree/61535b800e87ae68451a7134810cedfe0cd985f6/integration/reference-http) contains imports, configuration,
explicit wiring and executable assertions. It is a reference, not a production
runtime dependency. Required owner boundaries: **Service, Authentication, Authorization, Tenancy, HTTP Signature, Capability and Audit**. Optional adapters
are selected explicitly; exact versions and dependencies are in the source
module and the frozen [catalog](../packages.md), not a second version table.

## Construction and work lifecycle

Validate credentials/key material and permission policy before handler registration.

Integrity and replay policy, capability/tenant resolution, authentication and authorization precede protected use-case dispatch. Bounded validation and audit remain explicit.

## Failure, ownership and cleanup

Fail closed on missing/invalid identity or policy rejection. Application owns key/provider/audit lifetimes; bounded cancellation closes the service without implying external rollback.

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
go test -mod=readonly -run '^TestReferenceHTTPServiceLifecycleAndRequestContract$' -count=1 .
```

## Evidence and deployment boundary

Loopback and in-memory collaborators; not an installable Identity-platform service.

For Kubernetes, retain the [operator-owned setup and shutdown](../operations/kubernetes.md)
instead of deploying this fixture as a product. Keep actual backend versions,
application payloads and recovery policy in the deployment's own runbook.

Next: [supported stacks](../recommended-stacks.md), [operations](../operations/index.md),
[limitations](../limitations.md).
