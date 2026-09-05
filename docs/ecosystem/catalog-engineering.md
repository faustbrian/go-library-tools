# Golib Engineering Catalog

Design language `1.0` (`v1.5.3`); tooling `v1.5.3`.

## foundations

- `github.com/faustbrian/go-clock` (public library): Provide explicit wall-time and elapsed-time seams, owned timers and tickers, cancelable sleep, deterministic manual time, and synchronization-aware clock test helpers.

- `github.com/faustbrian/go-config` (public library): Load explicit, layered configuration sources into immutable typed snapshots with deterministic precedence, validation, safe provenance, and redacted secrets.

- `github.com/faustbrian/go-config/adapters/awssecretsmanager` (adapter): Load one bounded JSON configuration document from AWS Secrets Manager as an explicit sensitive configuration source.

- `github.com/faustbrian/go-correlation` (public library): Provide immutable correlation, request, causation, and external identifiers with explicit trust and disclosure across transport boundaries.

- `github.com/faustbrian/go-identifier` (public library): Provide immutable UUID, ULID, TypeID, KSUID, NanoID, typed domain identifier, and deterministic slug values with explicit generation, validation, encoding, ordering, and leakage contracts.

- `github.com/faustbrian/go-international` (public library): Provide immutable offline international reference values, strict parsing, explicit normalization, and versioned dataset provenance.

- `github.com/faustbrian/go-localized` (public library): Provide immutable localized text values, exact locale lookup, deterministic matching and fallback, and focused configuration, encoding, HTTP, persistence, validation, and wire integrations.

- `github.com/faustbrian/go-tenancy` (public library): Provide validated tenant identity, explicit tenant, system, and unscoped work, and fail-closed propagation and enforcement seams.

- `github.com/faustbrian/go-validation` (public library): Provide deterministic typed application validation, bounded diagnostics, explicit presence semantics, reusable rules, and transport-neutral report composition.

## service-edge

- `github.com/faustbrian/go-api-query` (public library): Compile declared API query capabilities into immutable, bounded, storage-neutral plans and adapt reviewed plans to explicit transport, validation, and persistence boundaries.

- `github.com/faustbrian/go-authentication` (public library): Turn Basic credentials, opaque bearer tokens, and API keys into immutable authenticated principals.

- `github.com/faustbrian/go-authentication/authotel` (public library): Adapt bounded authentication observations to OpenTelemetry traces and metrics.

- `github.com/faustbrian/go-authentication/jwt` (public library): Validate signed compact JWTs and operate bounded static or remote JWK sets at an authentication boundary.

- `github.com/faustbrian/go-authentication/oidc` (public library): Discover OpenID Providers and validate signed OpenID Connect ID tokens at the authentication trust boundary.

- `github.com/faustbrian/go-authorization` (public library): Provide typed ACL, RBAC, and ABAC policy evaluation, immutable revisioned snapshots, bounded policy compilation, fail-closed transport integration, and explicit persistence and invalidation adapters.

- `github.com/faustbrian/go-capability` (public library): Issue and verify narrowly scoped, tamper-evident, expiring capabilities and signed URLs with explicit replay, revocation, and storage boundaries.

- `github.com/faustbrian/go-http-middleware` (public library): Provide explicit, bounded server-side net/http middleware and deterministic chain composition without hidden registration or defaults.

- `github.com/faustbrian/go-identity` (public library): Plan a storage-neutral identity core for users, accounts, identifiers, credential references, verification state, account status, policy boundaries, and domain events.

- `github.com/faustbrian/go-oauth-server` (public library): Plan a storage-neutral OAuth 2.1 authorization-server core for clients, authorization, consent, grants, token issuance, refresh rotation, revocation, introspection, metadata, and signing keys.

- `github.com/faustbrian/go-organization` (public library): Plan a storage-neutral organization core for organizations, memberships, invitations, teams, role assignments, ownership transfer, domain claims, and lifecycle policy.

- `github.com/faustbrian/go-passkey` (public library): Plan storage-neutral passkey policy for identity-facing enrolment, discoverable credentials, passwordless signup and sign-in, backup state, credential management, and passkey-first flows.

- `github.com/faustbrian/go-password` (public library): Provide bounded Argon2id and bcrypt password hashing, verification, parsing, admission, and login-time upgrade primitives.

- `github.com/faustbrian/go-router` (public library): Provide explicit startup-time HTTP route composition, immutable compiled dispatch, safe URL generation, and deterministic route introspection over ordinary net/http handlers.

- `github.com/faustbrian/go-service` (public library): Coordinate explicit service construction, ordered lifecycle, supervision, health probes, HTTP serving, maintenance state, and caller-owned integration hooks without imposing an application framework.

- `github.com/faustbrian/go-sso` (public library): Plan a storage-neutral enterprise SSO core for provider lifecycle, verified-domain routing, discovery, mapping and JIT provisioning, organization membership, enforcement, break-glass recovery, enterprise token custody, and directory synchronization.

## protocols-and-descriptions

- `github.com/faustbrian/go-cloudevents` (public library): Provide a transport-independent CloudEvents 1.0 envelope with bounded JSON, HTTP, and Kafka mappings.

- `github.com/faustbrian/go-cloudevents/adapters/golib` (adapter): Bridge CloudEvents explicitly to selected Golib event, transport, workflow, metadata, audit, telemetry, and schema contracts.

- `github.com/faustbrian/go-http-signature` (public library): Provide bounded HTTP Message Signatures, digest fields, structured-field interpretation, explicit signing and verification policy, key resolution, replay consumption, and HTTP adapters.

- `github.com/faustbrian/go-json-schema` (public library): Compile and validate exact-number, multi-dialect JSON Schema documents with explicit resource loading and bounded execution.

- `github.com/faustbrian/go-jsonapi` (public library): Provide strict framework-neutral JSON:API encoding, decoding, validation, negotiation, query handling, Atomic Operations, and Cursor Pagination primitives.

- `github.com/faustbrian/go-jsonrpc` (public library): Provide bounded transport-neutral JSON-RPC 2.0 envelopes, dispatch, clients, middleware, hooks, and an optional HTTP binding.

- `github.com/faustbrian/go-openapi` (public library): Provide immutable, lossless, and resource-bounded Swagger 2.0 and OpenAPI 3.0 through 3.2 document modeling, parsing, validation, reference resolution, composition, conversion, compatibility diffing, and serialization.

- `github.com/faustbrian/go-openrpc` (public library): Provide bounded OpenRPC 1.3.x and 1.4.x document modeling, parsing, canonical serialization, validation, reference resolution, runtime expressions, discovery, composition, compatibility diffing, and JSON-RPC integration.

- `github.com/faustbrian/go-schema-registry` (public library): Provide provider-neutral schema registration, resolution, compatibility, canonicalization, portable fingerprints, bounded caching, offline bundles, and explicit wire-codec composition.

- `github.com/faustbrian/go-schema-registry/providers/confluent` (public library): Adapt the provider-neutral schema-registry contract to bounded Confluent-compatible REST, subject/version, compatibility, deletion, and version-0 wire semantics.

- `github.com/faustbrian/go-schema-registry/providers/glue` (public library): Adapt the provider-neutral schema-registry contract to bounded AWS Glue registry identity, lifecycle, service, and uncompressed header-version-3 wire semantics.

- `github.com/faustbrian/go-scim` (public library): Plan SCIM 2.0 discovery, Users and Groups, search, filtering, sorting, pagination, PATCH, Bulk, ETags, typed protocol errors, and organization/provider-owned connection lifecycle.

- `github.com/faustbrian/go-webauthn` (public library): Plan WebAuthn registration and authentication ceremonies, relying-party and origin policy, bounded protocol parsing, attestation and assertion verification, authenticator metadata, counters, and security-key profiles.

- `github.com/faustbrian/go-webhook` (public library): Provide protocol-independent webhook signing, exact-byte verification, replay protection, bounded delivery, and explicit adapters for durable and observable integrations.

- `github.com/faustbrian/go-wsdl` (public library): Provide bounded WSDL 1.1 and WSDL 2.0 document modeling, parsing, validation, injected resolution, immutable compilation, composition, code-generation models, and semantic diffing.

- `github.com/faustbrian/go-xsd` (public library): Parse, compile, validate, serialize, and build bounded XML Schema 1.0 documents without implicit external I/O.

## persistence-and-durability

- `github.com/faustbrian/go-audit` (public library): Provide immutable audit records, explicit delivery and redaction policy, integrity verification, bounded query and export contracts, and retention planning without coupling consumers to a storage backend.

- `github.com/faustbrian/go-audit/postgres` (public library): Persist audit records in append-only PostgreSQL storage with idempotent insertion, bounded query and export behavior, caller-owned transaction staging, and legal-hold-aware retention.

- `github.com/faustbrian/go-cache` (public library): Provide typed cache semantics, bounded cache-aside loading, explicit key and codec policy, and backend-neutral ownership contracts.

- `github.com/faustbrian/go-event-sourcing` (public library): Provide event-store and dispatcher contracts, aggregate repositories, codecs, upcasting, projections, snapshots, process managers, and deterministic test support.

- `github.com/faustbrian/go-event-sourcing/adapters/gokafka` (adapter): Adapt event-sourcing deliveries to Kafka records, synchronous dispatch, consumer handling, explicit poison and retry disposition, and dead-letter publication.

- `github.com/faustbrian/go-event-sourcing/adapters/gotelemetry` (adapter): Instrument event-sourcing dispatch, storage, serialization, snapshots, projections, process managers, and Kafka context propagation with caller-supplied OpenTelemetry components.

- `github.com/faustbrian/go-event-sourcing/adapters/outbox` (adapter): Atomically stage event rows and transactional-outbox envelopes through a savepoint in an existing caller-owned PostgreSQL transaction.

- `github.com/faustbrian/go-event-sourcing/adapters/queue` (adapter): Adapt complete event-sourcing deliveries to bounded queue envelopes, synchronous enqueue dispatch, and explicit live or replay task handling.

- `github.com/faustbrian/go-event-sourcing/postgres` (public library): Provide PostgreSQL event, snapshot, projection, checkpoint, migration, and caller-owned transaction-writer implementations for event sourcing.

- `github.com/faustbrian/go-feature-flags` (public library): Provide deterministic tenant-safe feature evaluation, versioned management, durable providers, bounded caching, and explicit fleet refresh.

- `github.com/faustbrian/go-idempotency` (public library): Provide durable operation ownership, fencing, bounded result replay, canonical fingerprints, and explicit transport and workload adapters.

- `github.com/faustbrian/go-lease` (public library): Provide backend-time distributed leases, managed renewal, monotonically increasing fencing tokens, and explicit worker and service coordination over memory, PostgreSQL, and Valkey backends.

- `github.com/faustbrian/go-migrations` (public library): Provide engine-neutral migration identity, validation, deterministic planning, execution coordination, recovery, and a PostgreSQL ledger and backend.

- `github.com/faustbrian/go-postgres` (public library): Provide finite pgx pool configuration and lifecycle, bounded transaction cleanup, SQLSTATE classification, safe observations, and real PostgreSQL test support.

- `github.com/faustbrian/go-queue` (public library): Provide backend-neutral bounded worker coordination plus explicit in-memory, Redis, Valkey, NATS, and NSQ queue implementations with observable delivery, settlement, failure, and lifecycle semantics.

- `github.com/faustbrian/go-queue-control-plane` (public library): Provide an authenticated administrative control plane for durable queue commands, desired state, audit history, fleet visibility, and optional Kubernetes workload scaling.

- `github.com/faustbrian/go-queue/queueservice` (public library): Integrate caller-selected queue producers and workers with go-service startup, readiness, supervision, admission closure, drain, and shutdown while preserving backend-owned delivery semantics.

- `github.com/faustbrian/go-queue/rabbitmq` (public library): Adapt the backend-neutral go-queue worker contract to RabbitMQ AMQP 0-9-1 publishing, consumption, recovery, settlement, retry, dead-letter, and shutdown policy owned by go-rabbitmq-queues.

- `github.com/faustbrian/go-scheduler` (public library): Provide code-defined recurring schedules, immutable compilation, bounded execution, fenced multi-replica coordination, explicit dispatch, and observable lifecycle integration.

- `github.com/faustbrian/go-sequencer` (public library): Provide dependency-ordered durable operation execution with immutable plans, fenced attempts, explicit retry and unknown-outcome policy, bounded fleet lifecycle, persistent stores, and selectable integration adapters.

- `github.com/faustbrian/go-settings` (public library): Provide typed runtime-mutable settings, explicit precedence, immutable snapshots, optimistic writes, audit history, schema evolution, and optional persistence and cache adapters.

- `github.com/faustbrian/go-transactional-outbox` (public library): Own PostgreSQL transactional outbox persistence, durable claims and state transitions, and caller-bounded at-least-once relay execution.

- `github.com/faustbrian/go-transactional-outbox/adapters/gokafka` (adapter): Map one durable outbox envelope to one confirmed first-party Kafka record without acquiring or owning the producer.

- `github.com/faustbrian/go-transactional-outbox/adapters/gorabbitstream` (adapter): Map one durable outbox envelope to one confirmed RabbitMQ Stream or Super Stream message without owning transport lifecycle.

- `github.com/faustbrian/go-transactional-outbox/adapters/otel` (adapter): Add bounded outbox semantic spans, metrics, propagation, observations, and publisher instrumentation without owning telemetry infrastructure.

- `github.com/faustbrian/go-transactional-outbox/adapters/queue` (adapter): Map one durable outbox envelope to one bounded deterministic first-party queue task while preserving acceptance ambiguity.

- `github.com/faustbrian/go-workflow` (public library): Provide immutable workflow definitions and history, deterministic orchestration decisions, bounded durable work processing, explicit recovery semantics, and a PostgreSQL persistence adapter.

## resilience

- `github.com/faustbrian/go-adaptive-throttle` (public library): Provide process-local rolling overload history and probabilistic admission that sheds a bounded share of work while preserving probe flow.

- `github.com/faustbrian/go-bulkhead` (public library): Provide process-local fixed-capacity resource isolation with weighted permits, bounded FIFO waiting, explicit partitions, and graceful drain.

- `github.com/faustbrian/go-circuit-breaker` (public library): Provide protocol-neutral, bounded circuit-breaker state, dependency-health admission, rolling outcome windows, and explicit permit and observation lifecycles.

- `github.com/faustbrian/go-concurrency-limit` (public library): Provide bounded, process-local adaptive in-flight admission that learns a safe concurrency limit from explicit execution outcomes.

- `github.com/faustbrian/go-fault-injection` (public library): Provide deterministic, bounded fault schedules, failure wrappers, and a fail-closed runtime for tests and explicitly authorized controlled experiments.

- `github.com/faustbrian/go-hedge` (public library): Provide finite delayed duplicate attempts for explicitly replay-safe work under caller-owned deadlines and bounded amplification.

- `github.com/faustbrian/go-rate-limit` (public library): Provide transport-neutral inbound admission policies, bounded memory and distributed backends, concurrency leases, and explicit HTTP, RPC, queue, principal, logging, and telemetry integrations.

- `github.com/faustbrian/go-resilience` (public library): Provide deterministic generic policy composition, typed outcomes, caller-owned total deadlines, bounded observation, and shared process-local retry and hedge work budgets.

- `github.com/faustbrian/go-retry` (public library): Provide bounded retry execution, explicit failure classification, finite time and work budgets, deterministic backoff strategies, and focused transport and observability adapters.

- `github.com/faustbrian/go-semaphore` (public library): Provide a process-local FIFO weighted semaphore with bounded waiting, exactly-once permits, deterministic shutdown, and bounded observation events.

## observability

- `github.com/faustbrian/go-log` (public library): Provide composable log/slog handlers for routing, redaction, sampling, bounded asynchronous delivery, capture, local rotation, and OpenTelemetry correlation.

- `github.com/faustbrian/go-telemetry` (public library): Provide a vendor-neutral OpenTelemetry runtime, OTLP exporters, propagation, sampling, bounded instrumentation, and explicit provider lifecycle for Go services.

## integration-and-data-movement

- `github.com/faustbrian/go-external-sort` (public library): Provide bounded external sorting of fixed-width opaque records using authenticated encrypted temporary storage.

- `github.com/faustbrian/go-filesystem` (public library): Provide capability-oriented streaming filesystem contracts, backend-specific adapters, composable decorators, and conformance helpers.

- `github.com/faustbrian/go-http-client` (public library): Provide typed outbound HTTP policy with finite transport defaults, immutable request specifications, deterministic middleware, and explicit response ownership.

- `github.com/faustbrian/go-kafka` (public library): Provide bounded first-party Apache Kafka producer, consumer, inspection, replay, and transaction policy over franz-go.

- `github.com/faustbrian/go-kafka/adapters/gotelemetry` (adapter): Translate Kafka observation and propagation contracts into OpenTelemetry spans and metrics.

- `github.com/faustbrian/go-kafka/adapters/mskiam` (adapter): Provide Kafka SASL/OAUTHBEARER authentication configuration through AWS MSK IAM credentials and signing.

- `github.com/faustbrian/go-kafka/kafkaservice` (public library): Bridge go-kafka consumer and producer resources into the explicit go-service lifecycle.

- `github.com/faustbrian/go-rabbitmq-queues` (public library): Provide bounded RabbitMQ-native AMQP 0-9-1 classic and quorum queue publishing, consumption, settlement, recovery, topology verification, health, and observation policy.

- `github.com/faustbrian/go-rabbitmq-streams` (public library): Provide vendor-neutral bounded policy for RabbitMQ Streams messages, publishing, consumption, replay, inspection, failures, lifecycle, and observations.

- `github.com/faustbrian/go-rabbitmq-streams/otel` (public library): Translate bounded RabbitMQ Streams observations into caller-owned OpenTelemetry metrics and propagate W3C Trace Context through message headers.

- `github.com/faustbrian/go-rabbitmq-streams/rabbitmq` (public library): Adapt the RabbitMQ-supported Go Streams client to bounded rabbitstream policy while owning protocol connections, sessions, cursors, recovery, and wire conversion.

- `github.com/faustbrian/go-search` (public library): Provide backend-neutral contracts for bounded document indexing, typed querying, cursor pagination, schema migration, projections, and reconciliation while treating application data as authoritative.

- `github.com/faustbrian/go-search/adapters/opensearch` (adapter): Translate the backend-neutral search contract to a bounded OpenSearch client with explicit transport, trust, lifecycle, resilience, and observability policy.

- `github.com/faustbrian/go-secret-envelope` (public library): Provide bounded authenticated secret envelopes with explicit key-provider boundaries, immutable encryption context, and versioned persistence bytes.

- `github.com/faustbrian/go-tabular` (public library): Provide explicit, bounded ingestion for delimited, fixed-width, XLS, XLSX, and ZIP-backed tabular sources.

- `github.com/faustbrian/go-wire` (public library): Provide explicit, bounded JSON, XML, SOAP, YAML, TOML, MessagePack, CBOR, and BSON encoding and decoding boundaries.

## domain-utilities

- `github.com/faustbrian/go-barcode` (public library): Provide immutable logical barcode symbols, strict validation, encoding, decoding, rendering, and standards-conformance evidence.

- `github.com/faustbrian/go-calendar` (public library): Provide immutable civil dates, Gregorian arithmetic, typed periods, explicit DST conversion, and bounded business calendars.

- `github.com/faustbrian/go-ecma-regexp` (public library): Provide bounded ECMAScript regular-expression parsing, compilation, and matching, including the JSON Schema pattern profile.

- `github.com/faustbrian/go-geo` (public library): Provide immutable geospatial values, bounded geometry and geodesy operations, interoperable codecs, PostGIS mapping, and deterministic test helpers.

- `github.com/faustbrian/go-keyphrase` (public library): Provide unbiased bounded password and passphrase generation, BIP-39 mnemonic interoperability, immutable word lists, and explicit secret-handling boundaries.

- `github.com/faustbrian/go-knapsack` (public library): Provide deterministic, bounded offline orthogonal packing, exact objectives, extension constraints, canonical plans, and independent verification.

- `github.com/faustbrian/go-knapsack/objective/gomoney` (adapter): Adapt exact go-money values into deterministic Knapsack container-cost objective evaluation.

- `github.com/faustbrian/go-math` (public library): Provide immutable arbitrary-precision integer, rational, decimal, and binary-float values with explicit precision, rounding, limits, conditions, and deterministic encodings.

- `github.com/faustbrian/go-measurement` (public library): Provide immutable, exact, unit-safe quantities, dimensions, conversions, logistics formulas, and bounded wire encodings.

- `github.com/faustbrian/go-merkle-patricia-trie` (public library): Provide bounded immutable Ethereum modified Merkle Patricia tries, roots, proofs, storage integration, retention, pruning, and recovery.

- `github.com/faustbrian/go-merkle-tree` (public library): Provide storage-independent ordered Merkle trees, canonical and RFC 9162 profiles, immutable snapshots, and bounded inclusion, multi-inclusion, and consistency proofs.

- `github.com/faustbrian/go-money` (public library): Provide immutable exact monetary values, explicit precision and rounding contexts, bounded arithmetic, allocation, tax, discount, conversion, formatting, and persistence encodings.

- `github.com/faustbrian/go-opening-hours` (public library): Model immutable recurring opening hours, dated exceptions, timezone-aware availability, and bounded schedule composition.

- `github.com/faustbrian/go-rule-engine` (public library): Provide deterministic typed rule construction, bounded compilation into immutable plans, explicit fact evaluation, canonical JSON, and redacted diagnostics without hidden I/O.

- `github.com/faustbrian/go-rule-engine/adapters/math` (adapter): Bridge immutable go-math decimals into rule-engine tagged values and deterministic equality and ordering operators without coupling the core module to decimal arithmetic.

- `github.com/faustbrian/go-rule-engine/adapters/measurement` (adapter): Bridge immutable go-measurement quantities into tagged rule-engine values and dimension-safe deterministic comparison operators.

- `github.com/faustbrian/go-rule-engine/adapters/temporal` (adapter): Bridge exact go-temporal instants and periods into tagged rule-engine values and deterministic relation operators.

- `github.com/faustbrian/go-state-machine` (public library): Provide deterministic typed state-machine compilation, transition selection, inert effect planning, replay and evolution, with optional explicit execution, persistence, outbox delivery, and diagram rendering.

- `github.com/faustbrian/go-temporal` (public library): Provide immutable temporal algebra, explicit interval bounds and relations, normalized sets, fixed durations, daily intervals, strict notation, and versioned encodings.

- `github.com/faustbrian/go-verkle-tree` (public library): Provide bounded immutable authenticated key/value trees, roots, proofs, witnesses, stateless updates, and caller-owned storage protocols for the package-owned Bandersnatch IPA profile.

## tooling

- `github.com/faustbrian/go-analysis` (public library): Provide deterministic go/analysis policies, governed rule metadata, and bounded JSON and SARIF reports for Go repositories.

- `github.com/faustbrian/go-cli` (public library): Provide explicit typed command construction, immutable compilation, deterministic parsing, lifecycle middleware, bounded output, and stable process-facing results.

- `github.com/faustbrian/go-library-tools` (public tool): Validate and execute the shared contract for independently released Golib repositories.

- `github.com/faustbrian/go-prompts` (public library): Provide typed interactive prompts, deterministic non-interactive parsing, semantic rendering, caller-driven presentation, and an explicit terminal adapter.

## unclassified-internal

- `example.com/analysis-coverage` (fixture): See the module README and goal files.

- `github.com/faustbrian/go-authorization/integration/contracts` (interoperability harness): See the module README and goal files.

- `github.com/faustbrian/go-bulkhead/benchmarks/comparison` (benchmark harness): This non-releasable benchmark module isolates comparison dependencies from the public bulkhead module. It measures capacity-one admitted and saturated paths for the local bulkhead, direct `x/sync/semaphore`, Failsafe-Go, and Fortify.

- `github.com/faustbrian/go-bulkhead/integration/resilience` (interoperability harness): This non-releasable integration module proves the application-owned resilience ordering through the public bulkhead, retry, and circuit-breaker contracts. Local bulkhead admission failure remains a permanent retry outcome and occurs before circuit-breaker admission, so it neither amplifies attempts nor records a downstream failure.

- `github.com/faustbrian/go-circuit-breaker/integration/consumers` (interoperability harness): See the module README and goal files.

- `github.com/faustbrian/go-cli/benchmarks` (benchmark harness): See the module README and goal files.

- `github.com/faustbrian/go-concurrency-limit/benchmarks/comparison` (benchmark harness): This non-releasable module isolates comparison dependencies from the public `concurrency-limit` module. It compares bounded Gradient2 update and permit paths across pinned local and external implementations.

- `github.com/faustbrian/go-concurrency-limit/integration/resilience` (interoperability harness): This non-releasable integration module proves application-owned composition through the public adaptive limiter, retry, and hedge contracts. Local admission rejection remains a permanent retry outcome, and every hedge attempt must acquire its own limiter permit before invoking downstream work.

- `github.com/faustbrian/go-correlation/integration/siblings` (interoperability harness): See the module README and goal files.

- `github.com/faustbrian/go-event-sourcing/benchmarks/competitors` (benchmark harness): This non-releasable module isolates comparison dependencies from the event-sourcing core. It compares equivalent observable work and keeps correctness checks separate from timing.

- `github.com/faustbrian/go-fault-injection/benchmarks/comparison` (benchmark harness): This non-releasable module isolates comparison dependencies from the core fault-injection module. All candidates return a caller-visible error. The fault-injection and goresilience cases prevent the wrapped operation from running; the direct double is the minimum equivalent test outcome.

- `github.com/faustbrian/go-fault-injection/integration/resilience` (interoperability harness): This non-releasable module proves that `fault-injection` can drive deterministic campaigns through the public retry and circuit-breaker contracts. The resilience modules do not import fault-injection in production; the dependency direction exists only in this integration module.

- `github.com/faustbrian/go-http-middleware/integration/siblings` (interoperability harness): Non-releasable interoperability harness proving middleware composition with the broader HTTP service stack.

- `github.com/faustbrian/go-http-signature/benchmarks/comparison` (benchmark harness): This non-releasable module isolates the comparison dependency from the public HTTP Message Signatures module. It compares an equivalent HMAC-SHA256 request operation with the maintained `github.com/yaronf/httpsign` implementation at commit `de382d35c1add89cc09b9355161d61471fb7f632` and `github.com/dadrus/httpsig` at commit `0f24bf7dd9b76727af985d9a6f7ce87207a18387`.

- `github.com/faustbrian/go-http-signature/differential/shared-corpus` (interoperability harness): This standalone module executes one checked-in corpus through this module and maintained independent Go implementations. It does not run peer self-tests.

- `github.com/faustbrian/go-idempotency/compatibility/ecosystem` (interoperability harness): Compile and exercise the published idempotency, logging, migrations, outbox, queue, telemetry, and webhook contracts as one independent consumer.

- `github.com/faustbrian/go-json-schema/benchmarks/comparison` (benchmark harness): See the module README and goal files.

- `github.com/faustbrian/go-jsonapi/interoperability` (interoperability harness): This non-releasable harness compares overlapping JSON:API decisions with pinned DataDog/jsonapi v0.13.0 behavior while keeping the peer outside the public module.

- `github.com/faustbrian/go-jsonrpc/interoperability` (interoperability harness): This non-releasable differential harness compares the JSON-RPC decision contract with pinned creachadair/jrpc2 v1.3.5 behavior while keeping the peer dependency outside the public module.

- `github.com/faustbrian/go-kafka/benchmarks/clients` (benchmark harness): This non-releasable module isolates comparison clients and container tooling from the Kafka policy module. It records end-to-end broker measurements only where the candidates can provide the same observable contract. It is not a source of production support claims or a substitute for fault evidence.

- `github.com/faustbrian/go-knapsack/integration/references` (interoperability harness): See the module README and goal files.

- `github.com/faustbrian/go-openapi/interoperability` (interoperability harness): This non-releasable harness compares OpenAPI parsing, modeling, validation, and round-trip behavior with pinned maintained peers while keeping competitor dependencies outside the public module.

- `github.com/faustbrian/go-postgres/examples/migrations` (example): This executable combines `postgres`, pgx's `database/sql` bridge, and `migrations`. It plans and applies embedded migrations once as a dedicated deployment job:

- `github.com/faustbrian/go-prompts/benchmarks/comparison` (benchmark harness): See the module README and goal files.

- `github.com/faustbrian/go-rule-engine/benchmarks/competitors` (benchmark harness): See the module README and goal files.

- `github.com/faustbrian/go-secret-store` (planned boundary): Reserve an unclassified planning boundary without claiming an installable package, runtime contract, implementation, or release.

- `github.com/faustbrian/go-service/benchmarks/platform` (benchmark harness): This non-releasable benchmark module compares equivalent request behavior across:

- `github.com/faustbrian/go-service/compatibility` (interoperability harness): See the module README and goal files.

- `github.com/faustbrian/go-service/integration/adoption` (interoperability harness): This integration module validates that Track, Postal, and Location can replace their generic bootstrap with the public `service` construction model while keeping application dependencies and business behavior explicit.

- `github.com/faustbrian/go-service/integration/reference-durability` (interoperability harness): This maintained non-production module exercises Golib's PostgreSQL and Valkey durability stack through public APIs. It is assurance infrastructure, not a deployable product or an application dependency.

- `github.com/faustbrian/go-service/integration/reference-external` (interoperability harness): This non-production module proves that Golib's public outbound-integration APIs compose without a private framework layer. It covers bounded HTTP calls, standalone resilience controls, signed webhook delivery, filesystem storage, and authenticated secret envelopes.

- `github.com/faustbrian/go-service/integration/reference-http` (interoperability harness): This non-production integration module proves that the recommended Golib HTTP stack composes through public APIs without a private framework layer. It is an executable assurance fixture, not a deployable product or a runtime dependency for application services.

- `github.com/faustbrian/go-service/integration/reference-platform` (interoperability harness): This non-production module verifies the public `service` process model in disposable Linux containers. Its platform harness builds and runs both `linux/amd64` and `linux/arm64` images with `CGO_ENABLED=0`, a non-root user, a read-only root filesystem, a bounded writable temporary filesystem, dropped capabilities, process and descriptor limits, health probes, DNS, private TLS trust, and graceful `SIGTERM` handling.
