# Kubernetes operations: PostgreSQL, Valkey, Kafka, OpenSearch and telemetry

[Documentation index](../index.md) · [Ecosystem entry](../../README.md)

This is an operator ownership and rollout guide, not a bundled deployment,
Helm chart or proof that all five backends have been tested together. Select
only the facilities your operation needs. Applications own configuration,
resource acquisition, wiring, policy order, health and cleanup; platform teams
own backend provisioning, credentials, storage, backups and cluster policy.

Use design-language **1.0** and
[`golib-compat-v1-20260912.2`](../../compatibility-sets.md) when selecting the frozen
cohort. Package-current releases, newer majors and their stronger backend tests
have their own versioned contracts; they do not retroactively qualify a frozen
recipe, a managed backend, or your production cluster. Record backend versions
and the actual application composition in the deployment's own runbook.

## Assign ownership before provisioning

| Facility | Package boundary | Deployment owner and operational contract |
| --- | --- | --- |
| Process, probes and drain | [Service Kubernetes guide](https://github.com/faustbrian/go-service/blob/v1.1.0/docs/kubernetes.md) | Application owns role-specific commands and bounded lifecycle; Kubernetes supervises/restarts processes. Keep the management port private. |
| PostgreSQL | [Postgres Kubernetes guide](https://github.com/faustbrian/go-postgres/blob/v1.1.0/docs/kubernetes.md) and [migration integration](https://github.com/faustbrian/go-postgres/blob/v1.1.0/docs/migrations.md) | Platform owns server, backup/failover and connection budget; application owns SQL, transactions, pool and use-case readiness; deployment Job owns reviewed DDL. |
| Valkey cache and durable jobs | [Cache operations](https://github.com/faustbrian/go-cache/blob/v1.0.0/docs/operations.md), [Queue dead letters](https://github.com/faustbrian/go-queue/blob/v1.1.0/docs/dead-letter-operations.md) | Client acquisition, TLS, key/codec namespace and failure policy are explicit. Cache eviction is normal for reconstructible data; queue/lease durability and recovery are separate contracts. Do not treat every backend as a supported cluster/failover topology. |
| Queue administration and scaling | [Queue control-plane operations](https://github.com/faustbrian/go-queue-control-plane/blob/v1.1.0/docs/operations.md) and [Kubernetes boundary](https://github.com/faustbrian/go-queue-control-plane/blob/v1.1.0/docs/kubernetes.md) | Queue owns delivery/settlement; the control plane owns authenticated, authorized administrative commands/audit. HPA/KEDA or another platform controller owns scaling. A scale adapter is not a workload operator. |
| Kafka | [Kafka operations](https://github.com/faustbrian/go-kafka/blob/v1.1.0/docs/operations.md) | Platform owns brokers/topics, ACLs, retention and replication policy. Application owns producer/consumer identities, group ownership, schemas, commit timing, retries and replay authorization. |
| OpenSearch projection | [Search schema migrations](https://github.com/faustbrian/go-search/blob/v1.0.0/docs/schema-migrations.md), [OpenSearch adapter operations](https://github.com/faustbrian/go-search/blob/adapters/opensearch/v1.0.0/adapters/opensearch/docs/operations.md) | Platform owns cluster/storage; application owns source-of-truth projection, bounded bulk outcomes, index generations, aliases, checkpoints, write fencing and reconciliation. Search is not the transactional source of truth. |
| Telemetry | [Telemetry Kubernetes guide](https://github.com/faustbrian/go-telemetry/blob/v1.2.0/docs/kubernetes.md) and [shutdown](https://github.com/faustbrian/go-telemetry/blob/v1.2.0/docs/graceful-shutdown.md) | Application owns providers/instrumentation; platform owns Collector transport, storage, access, retention and alerts. Stable traces/metrics do not imply a stable Golib log exporter. |

The frozen Kafka/CloudEvents and Track/Location projection compositions do not
prove live broker/OpenSearch interoperability. Package fixture results prove
only their declared server/tool versions and exercised cases, not the complete
operator topology or application business behavior.

## Startup, probes and shutdown

Construct dependencies in application order: validated configuration, telemetry,
backend clients/pools, stores and policies, then handlers or workers. Transfer
ownership only where the selected adapter contract says so; close borrowed
clients in their application owner. One-shot migration commands should acquire
only migration dependencies, not start the serving stack.

Use a startup probe for construction, readiness for safe new-work admission,
and liveness for local process health. Do not restart every pod because a
backend or Collector is unavailable. A Collector outage should not withdraw
business readiness. Match this distinction to the selected service and backend
probe contracts; see [Kubernetes probe semantics](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/).

On termination, withdraw readiness and stop local admission, drain requests and
jobs, preserve delivery settlement/unknown outcomes, then close dependencies in
reverse ownership order. Flush telemetry after business components stop.
Budget `terminationGracePeriodSeconds` for endpoint-removal delay, composed
shutdown deadlines and measured margin. Kubernetes can forcibly terminate work
when grace expires; a pod or node loss may skip drain entirely. Durable jobs
need safe redelivery and idempotency, not a reliance on cleanup always running.
See [Pod termination](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#pod-termination).

## Configuration, credentials and refresh

Load explicit environment/file sources through
[Config](https://github.com/faustbrian/go-config/blob/v1.1.0/docs/operations.md),
validate a complete candidate before acquisition and never print plaintext
configuration, DSNs, tokens or payloads. Config does not watch or hot-reload.
Infisical delivery/authentication/rotation belongs to the deployment's selected
external integration; Golib core supplies no native Infisical SDK adapter.

Prefer narrow Secret access, verified TLS and least-privilege application,
migration and administrative identities. Mounted Kubernetes Secret volumes
update eventually, and `subPath` mounts do not receive automatic updates; this
is not proof that an already-running client reloaded credentials.
[Secret update behavior](https://kubernetes.io/docs/concepts/configuration/secret/#using-secrets-as-files-from-a-pod)
requires an explicit restart or a bounded application refresh that reconstructs
and validates dependencies, atomically transfers use and drains the old owner.
Do not assume every backend supports live credential rotation.

## Deployment and data movement

1. **Prepare immutable inputs.** Select application image digest, exact direct
   module versions, backend topology and schema/payload/index generations.
   Verify restore access, capacity and the prior compatible image.
2. **Expand PostgreSQL first.** Review plan/status and run one controlled
   migration Job with finite lock, statement and process deadlines, a narrow
   DDL role and an explicit retry policy. Deploy serving code only after success.
   Do not run migrations in every replica; an advisory lock is not deployment
   orchestration. Follow the selected
   [migration operations](https://github.com/faustbrian/go-migrations/blob/v1.1.0/docs/operations.md).
3. **Deploy compatible readers, then writers.** Keep old schema and payload
   formats understood during rolling overlap. Allocate connection headroom for
   `MaxConns × maximum replicas`, including surge, plus administration/migrations.
   Set CPU/memory requests from representative work, including queues, buffers,
   goroutines and cold-start peaks; illustrative manifests are not sizing proof.
4. **Canary effects and pressure.** Verify authorized requests, committed data,
   cache invalidation, job redelivery/settlement and any broker-to-projection
   path actually used. Bound concurrency, backlog, payload sizes and retries
   across maximum replicas, not just within one process. Retries, rate limits,
   circuit breakers and backpressure are selected policies, not automatic defaults.
5. **Move projections with a fence.** Create a new index generation, checkpoint
   rebuild, reconcile writes and verify the full target before a fenced alias
   cutover. Retain the old generation and a rollback path; do not infer success
   from counts, sampling or task submission alone. Use the selected version's
   exact durable coordination contract, not a process-local mutex.
6. **Contract later.** Remove old columns, payloads or generations only after
   the rollback window and supported old binaries/consumers are gone. Cleanup
   and historical replay remain separately authorized operations.

For the package-current migration ledger change, read
[Migrations v3.0.1 operations](https://github.com/faustbrian/go-migrations/blob/v3.0.1/docs/operations.md)
before upgrade. Its `public.migrations` ledger must not be confused with the
cohort's `public.go_schema_migrations` or Laravel's existing history.

## Failure and rollback runbook

| Symptom | First decision and preserved boundary |
| --- | --- |
| PostgreSQL timeout or migration error | Stop the rollout; inspect database and ledger before retrying an uncertain commit or dirty migration. Restore schema and ledger from the same consistent backup if restore is required. |
| Valkey unavailable or cache corruption | Backend errors are not misses. Use the named use-case fail-open/closed policy; isolate key/codec generations. Do not use stale values for authorization, balances or revocation. |
| Queue failures or backlog | Bound or stop redrive, inspect lineage/idempotency and advertised management capabilities. Unsupported metrics are not zero and command admission is not successful execution. |
| Kafka lag or delivery ambiguity | Inspect real group/topic/ISR state and retained offsets; keep payloads private. Coordinate replay with business idempotency. Resetting live group offsets is not a rollback plan. |
| OpenSearch partial/unknown writes or cutover | Classify individual outcomes and reconcile against authoritative data. Resolve alias/task state before retry; rollback requires the same fence, reconciliation and compatible resolver state. |
| Collector/exporter failure | Keep exporter queues finite; observe drops/rejections through independent health/logs. Do not turn telemetry loss into a backend restart storm. |

Roll back one application dependency/traffic boundary to its prior image and
versions only while schema, payload and index semantics remain compatible.
A Kubernetes rollout undo does not reverse durable writes, recover evicted
queue data, restore database history or undo an alias mutation. Verify the
original operation after rollback, not only pod readiness.

Monitor dependency errors and acquisition waits, cache freshness/invalidation,
queue wait/failure/settlement, Kafka lag/commit/rebalance, projection drift,
partial shards and exporter drops. Use bounded labels; identifiers, tenants,
raw payloads and queries are not metric labels by default. Measure equivalent
workloads before making performance/capacity claims and maintain restore drills
for the actual backend topology.

## Next steps

Use [recommended compositions](../recommended-stacks.md) for supported recipe
boundaries, [compatibility](../../compatibility-sets.md) for versions and
[Laravel/PHP migration](../migration/laravel-php.md) for coexistence and ledger
handover. Backend provisioning manifests and production qualification stay in
the application's deployment repository.
