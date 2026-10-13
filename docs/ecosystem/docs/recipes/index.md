# Runnable composition recipes

[Documentation map](../index.md) · [Ecosystem entry point](../../README.md)

These fourteen routes consume the exact published
[`golib-compat-v1-20260912.2`](../../compatibility-sets.md) bindings. Complete
program/test sources stay in their standalone owners; this portal does not
copy private APIs, create an umbrella module or replace executable examples
with pseudocode. Recipe fixtures are not installable application products.

| Composition | Complete route | Boundary |
| --- | --- | --- |
| Minimal HTTP service | [01-minimal-http-service](01-minimal-http-service.md) | Loopback HTTP only; no deployed network, container or load proof. |
| Internal JSON-RPC service | [02-internal-json-rpc-service](02-internal-json-rpc-service.md) | Loopback and in-memory audit/telemetry; no production or managed-service claim. |
| External JSON:API service | [03-external-jsonapi-service](03-external-jsonapi-service.md) | httptest JSON:API composition. The separate handwritten OpenAPI-named handler is not generated-contract evidence. |
| Authenticated and authorized service | [04-authenticated-authorized-service](04-authenticated-authorized-service.md) | Loopback and in-memory collaborators; not an installable Identity-platform service. |
| Queue producer and worker | [05-queue-producer-worker](05-queue-producer-worker.md) | PostgreSQL 18.4 and Valkey 9.1.0 controlled external fixtures; no managed failover or production claim. |
| Ingester and processor | [06-ingester-processor](06-ingester-processor.md) | Bounded in-memory transport, not broker durability. No remote storage interoperability is implied. |
| PostgreSQL transaction, migrations, idempotency and outbox | [07-postgres-idempotency-outbox](07-postgres-idempotency-outbox.md) | Controlled PostgreSQL/Valkey durability and process-loss supplement. No backup/restore, cross-region or ambiguous-commit resolution proof. |
| Scheduled singleton work | [08-scheduled-singleton](08-scheduled-singleton.md) | Unit selectors plus separate PostgreSQL 18.4 and Valkey 9.1.0 executions; Valkey noeviction durability assumptions are explicit. |
| Kafka event flow with schema and CloudEvents | [09-kafka-schema-cloudevents](09-kafka-schema-cloudevents.md) | Bounded application-owned Kafka/schema seams; no broker-runtime interoperability claim. |
| Vendor HTTP client policy stack | [10-vendor-http-client-policy](10-vendor-http-client-policy.md) | TLS httptest contract and policy ordering; not a measured production latency recommendation. |
| Filesystem and tabular ingestion | [11-filesystem-tabular-ingestion](11-filesystem-tabular-ingestion.md) | In-memory filesystem. SFTP/object-storage/R2/other remote variants require their own backend evidence. |
| Durable workflow with compensation | [12-durable-workflow-compensation](12-durable-workflow-compensation.md) | Pure example plus PostgreSQL 18.4 restart/compensation execution; not universal reversal of external effects. |
| Searchable Track/Location projection | [13-searchable-track-location](13-searchable-track-location.md) | Bounded HTTP transport seam, not live OpenSearch interoperability or Track/Location business behavior. |
| Track, Postal and Location service roles | [14-track-postal-location-roles](14-track-postal-location-roles.md) | Loopback/in-memory public construction fixtures; not the real applications business behavior or production deployment. |

## Run a reference

Use Go 1.27.0 and a fresh disposable checkout. Each recipe supplies an exact
repository, source revision, module directory and command. Select those inputs;
do not run a frozen example against latest dependencies without migration.

1. Clone the named repository into a task-owned directory and check out the
   listed immutable source revision. Read its configuration and fixture README.
2. Separately clone `go-library-tools` at public tag `v1.8.4` for the frozen
   rebase helper. From the recipe module, invoke its
   `tools/compatibility/generate.py rebase` command with
   `--set-id golib-compat-v1-20260912.2` and `--go-mod` pointing to that disposable
   module's `go.mod`. This updates only already-required cohort modules; it
   does not mutate the published set or a maintained application checkout.
3. Set `GOWORK=off`, task-owned `GOCACHE`, `GOMODCACHE`, `GOTMPDIR`, and retain
   authenticated public proxy/SumDB resolution. Reject any `replace` directive.
   Tidy this disposable module after rebase, then run the recipe command with
   read-only module resolution. A blank import/compile check is not a runtime
   recipe result; a skipped external selector is not a passing recipe.
4. Supply only the fixtures required by that scenario. PostgreSQL/Valkey image
   identities and assumptions belong to the versioned source/compatibility set.
   System-risk scripts run only in their documented controlled/hosted boundary.
5. Capture the exact scenario outcome, platform and backend version where
   material, then remove the exact task-owned checkout/caches/fixtures, including
   after failures. Never clean shared developer or production resources.

The recipes expose contexts, cancellation, errors, budgets and reverse cleanup
in real public-API code. The final documentation milestone executes all
supported routes; unchanged immutable Cohesion evidence can be reused only
where it still proves the same claim, with its original identity and result.
Neither these commands nor a passing fixture certifies a production deployment.

Next: [recommended stacks](../recommended-stacks.md), [operator route](../operations/kubernetes.md),
[limitations](../limitations.md).
