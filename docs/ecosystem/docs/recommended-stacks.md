# Recommended package compositions

[Documentation map](index.md)

Choose a composition for its actual operation, not a default collection of
packages. The fourteen routes below are bound to the published
[compatibility cohort](../compatibility-sets.md). Each recipe identifies the
exact source, required owners, initialization, optional adapter selection,
request/job order, failure/unknown outcomes, shutdown and execution boundary.
The source module owns exact import/version declarations; the portal does not
maintain a second drifting version matrix.

| Operation | Required semantic owners | Processing order and boundary |
| --- | --- | --- |
| [Minimal HTTP service](recipes/01-minimal-http-service.md) | Service, healthhttp and serverhttp | Serve a business request and readiness request; withdraw readiness during drain before joining both servers. |
| [Internal JSON-RPC service](recipes/02-internal-json-rpc-service.md) | Service, Router, JSON-RPC, Config and selected identity/policy/observation modules | Outer observation encloses signature/digest, capability/tenant, authentication, authorization, bounded decoding/validation, use case and audit boundaries. Request IDs do not confer idempotency. |
| [External JSON:API service](recipes/03-external-jsonapi-service.md) | HTTP Middleware, Service, HTTP Client, Log, Authentication, Authorization and JSON:API | Resolve identity/permissions, validate the resource contract, dispatch application logic and map its outcome without leaking transport models into domain code. |
| [Authenticated and authorized service](recipes/04-authenticated-authorized-service.md) | Service, Authentication, Authorization, Tenancy, HTTP Signature, Capability and Audit | Integrity and replay policy, capability/tenant resolution, authentication and authorization precede protected use-case dispatch. Bounded validation and audit remain explicit. |
| [Queue producer and worker](recipes/05-queue-producer-worker.md) | Queue, its service adapter, Postgres, Migrations, Idempotency and Transactional Outbox | Business/idempotency/outbox commit makes work publishable. Publish to Valkey; process before acknowledging. Replacement workers reclaim unacknowledged work. |
| [Ingester and processor](recipes/06-ingester-processor.md) | Service, Correlation, Queue and adapters/service | Accepted handoff preserves correlation; drain rejects late intake before admitted work finishes. Acknowledge after successful application processing. |
| [PostgreSQL transaction, migrations, idempotency and outbox](recipes/07-postgres-idempotency-outbox.md) | Postgres, Migrations, Idempotency, Transactional Outbox and Queue | Business state, idempotency completion and outbox intent share one caller-visible transaction. Publish only committed intent; receiver settles after processing. |
| [Scheduled singleton work](recipes/08-scheduled-singleton.md) | Scheduler, Lease backends and explicit dispatch/queue owner | Two competing runners admit one fenced dispatch. The schedule triggers an attempt; the lease is not exactly-once business-effect proof. |
| [Kafka event flow with schema and CloudEvents](recipes/09-kafka-schema-cloudevents.md) | CloudEvents target adapters, Kafka, Schema Registry and JSON Schema | Invalid events never publish. Decode/validate, perform application handling, then settle the exact topic/partition/offset; handler failure does not commit. |
| [Vendor HTTP client policy stack](recipes/10-vendor-http-client-policy.md) | HTTP Client, Retry, Rate Limit, Circuit Breaker, Cache, Telemetry and explicit bulkhead | Cache and logical rate admission surround breaker and bounded retry; bulkhead admission applies per physical attempt. Observation and transport wrappers preserve those counting boundaries. |
| [Filesystem and tabular ingestion](recipes/11-filesystem-tabular-ingestion.md) | Filesystem and Tabular | Require the header, deliver rows to the application callback and check cancellation before reads and buffered delivery. No automatic acknowledgement or retry. |
| [Durable workflow with compensation](recipes/12-durable-workflow-compensation.md) | Workflow and its PostgreSQL persistence module | Persist progress and failures; restart through the public durable state machine. Compensation has its own application semantics and may require operator resolution. |
| [Searchable Track/Location projection](recipes/13-searchable-track-location.md) | Search and adapters/opensearch | Write the representative projection through the public adapter, then query and interpret typed results through the same boundary. |
| [Track, Postal and Location service roles](recipes/14-track-postal-location-roles.md) | Service and explicit role-specific dependencies in the adoption fixture | API/RPC, workers, schedulers, migrations and activation retain separate plans. Unselected roles do not initialize unrelated facilities; application router/domain behavior remains outside the fixture. |

## Optional choices and common ownership

Observability, configuration providers and backend adapters are optional unless
required by the chosen source module. Construct caller-owned infrastructure
before consumers; validate policy and resource bounds before intake; register
handlers/workers explicitly. Applications own secrets, pool/transport/provider
lifetimes, policy ordering, cancellation, drain and reverse cleanup. No recipe
installs a global registry, service container or universal resilience stack.

A known failure, partial success and unknown external outcome require different
recovery. Publish committed outbox intent; acknowledge only after the stated
application processing boundary. Close consumers before the transport/pool they
borrow and flush observation resources last within a bounded shutdown budget.
Follow each recipe rather than substituting one generic `Close` sequence.

Mixed protocols, webhook sender/receiver stacks, SFTP/object-storage/R2 variants,
Infisical delivery and a combined observable Kubernetes deployment are additional
architectural/documentation targets, not new verified cohort compositions.
Label their unproved runtime boundaries and qualify them separately. OpenAPI
is description tooling: the JSON:API recipe is not generated-handler evidence.

Next: [recipe index](recipes/index.md), [ownership map](integration-map.md),
[operations](operations/index.md), [limitations](limitations.md).
