# Operations

[Documentation map](../index.md) · [Ecosystem entry point](../../README.md)

The [Kubernetes operator route](kubernetes.md) combines process/probes/drain,
PostgreSQL, Valkey, queues, Kafka, OpenSearch, telemetry, controlled migration
Jobs, configuration/secret ownership, overload, upgrade and incident boundaries.
It is a runbook, not an all-five-backend deployment or production certificate.

## Before deployment

- Select only the modules/backends your operation uses and record exact
  application and backend versions. The frozen
  [compatibility set](../../compatibility-sets.md) is bounded composition
  evidence, not proof of your infrastructure topology.
- Assign pool, broker, transport, provider, secret and migration owners. Keep
  migration credentials separate from serving credentials and management ports
  private. Validate configuration before acquisition and never log secrets.
- Define total deadlines, admission/pool bounds, retry eligibility, settlement,
  idempotency and unknown-outcome reconciliation before adding resilience.
- Measure resource requests/limits and shutdown budgets on equivalent workloads;
  see [capacity interpretation](../comparisons/performance.md).

## Local development and recovery

Use the [runnable recipe](../recipes/index.md) for the narrow contract being
developed. Backend fixtures are task-owned and disposable; use explicit image
identities and remove their exact containers/volumes/caches after evidence
capture. External-service and destructive recovery scripts belong only in
their documented controlled environment, not against shared developer or
production resources. Loopback tests do not demonstrate managed failover.

For an incident, identify the first failed application boundary and distinguish
known failure from unknown external effects. Stop unsafe new admission if
necessary, preserve failure/audit records, and use the owning replay/recovery
guide. Reclaiming transport work, retrying a logical operation, rebuilding a
projection and restoring a backup are different actions. Confirm the original
operation afterward, not only pod readiness. The
[limitations register](../limitations.md) keeps unproved recovery claims visible.

Next: [Kubernetes runbook](kubernetes.md), [migration](../migration/index.md),
[supported stacks](../recommended-stacks.md).
