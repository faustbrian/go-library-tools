# Golib Consumer Catalog

Design language `1.0` (`v1.5.3`); tooling `v1.5.3`.

## foundations

- `github.com/faustbrian/go-clock`: Provide explicit wall-time and elapsed-time seams, owned timers and tickers, cancelable sleep, deterministic manual time, and synchronization-aware clock test helpers.

- `github.com/faustbrian/go-config`: Load explicit, layered configuration sources into immutable typed snapshots with deterministic precedence, validation, safe provenance, and redacted secrets.

- `github.com/faustbrian/go-config/adapters/awssecretsmanager`: Load one bounded JSON configuration document from AWS Secrets Manager as an explicit sensitive configuration source.

- `github.com/faustbrian/go-correlation`: Provide immutable correlation, request, causation, and external identifiers with explicit trust and disclosure across transport boundaries.

- `github.com/faustbrian/go-identifier`: Provide immutable UUID, ULID, TypeID, KSUID, NanoID, typed domain identifier, and deterministic slug values with explicit generation, validation, encoding, ordering, and leakage contracts.

- `github.com/faustbrian/go-international`: Provide immutable offline international reference values, strict parsing, explicit normalization, and versioned dataset provenance.

- `github.com/faustbrian/go-localized`: Provide immutable localized text values, exact locale lookup, deterministic matching and fallback, and focused configuration, encoding, HTTP, persistence, validation, and wire integrations.

- `github.com/faustbrian/go-tenancy`: Provide validated tenant identity, explicit tenant, system, and unscoped work, and fail-closed propagation and enforcement seams.

- `github.com/faustbrian/go-validation`: Provide deterministic typed application validation, bounded diagnostics, explicit presence semantics, reusable rules, and transport-neutral report composition.

## service-edge

- `github.com/faustbrian/go-api-query`: Compile declared API query capabilities into immutable, bounded, storage-neutral plans and adapt reviewed plans to explicit transport, validation, and persistence boundaries.

- `github.com/faustbrian/go-authentication`: Turn Basic credentials, opaque bearer tokens, and API keys into immutable authenticated principals.

- `github.com/faustbrian/go-authentication/authotel`: Adapt bounded authentication observations to OpenTelemetry traces and metrics.

- `github.com/faustbrian/go-authentication/jwt`: Validate signed compact JWTs and operate bounded static or remote JWK sets at an authentication boundary.

- `github.com/faustbrian/go-authentication/oidc`: Discover OpenID Providers and validate signed OpenID Connect ID tokens at the authentication trust boundary.

- `github.com/faustbrian/go-authorization`: Provide typed ACL, RBAC, and ABAC policy evaluation, immutable revisioned snapshots, bounded policy compilation, fail-closed transport integration, and explicit persistence and invalidation adapters.

- `github.com/faustbrian/go-capability`: Issue and verify narrowly scoped, tamper-evident, expiring capabilities and signed URLs with explicit replay, revocation, and storage boundaries.

- `github.com/faustbrian/go-http-middleware`: Provide explicit, bounded server-side net/http middleware and deterministic chain composition without hidden registration or defaults.

- `github.com/faustbrian/go-password`: Provide bounded Argon2id and bcrypt password hashing, verification, parsing, admission, and login-time upgrade primitives.

- `github.com/faustbrian/go-router`: Provide explicit startup-time HTTP route composition, immutable compiled dispatch, safe URL generation, and deterministic route introspection over ordinary net/http handlers.

- `github.com/faustbrian/go-service`: Coordinate explicit service construction, ordered lifecycle, supervision, health probes, HTTP serving, maintenance state, and caller-owned integration hooks without imposing an application framework.

## protocols-and-descriptions

- `github.com/faustbrian/go-cloudevents`: Provide a transport-independent CloudEvents 1.0 envelope with bounded JSON, HTTP, and Kafka mappings.

- `github.com/faustbrian/go-cloudevents/adapters/golib`: Bridge CloudEvents explicitly to selected Golib event, transport, workflow, metadata, audit, telemetry, and schema contracts.

- `github.com/faustbrian/go-http-signature`: Provide bounded HTTP Message Signatures, digest fields, structured-field interpretation, explicit signing and verification policy, key resolution, replay consumption, and HTTP adapters.

- `github.com/faustbrian/go-json-schema`: Compile and validate exact-number, multi-dialect JSON Schema documents with explicit resource loading and bounded execution.

- `github.com/faustbrian/go-jsonapi`: Provide strict framework-neutral JSON:API encoding, decoding, validation, negotiation, query handling, Atomic Operations, and Cursor Pagination primitives.

- `github.com/faustbrian/go-jsonrpc`: Provide bounded transport-neutral JSON-RPC 2.0 envelopes, dispatch, clients, middleware, hooks, and an optional HTTP binding.

- `github.com/faustbrian/go-openapi`: Provide immutable, lossless, and resource-bounded Swagger 2.0 and OpenAPI 3.0 through 3.2 document modeling, parsing, validation, reference resolution, composition, conversion, compatibility diffing, and serialization.

- `github.com/faustbrian/go-openrpc`: Provide bounded OpenRPC 1.3.x and 1.4.x document modeling, parsing, canonical serialization, validation, reference resolution, runtime expressions, discovery, composition, compatibility diffing, and JSON-RPC integration.

- `github.com/faustbrian/go-schema-registry`: Provide provider-neutral schema registration, resolution, compatibility, canonicalization, portable fingerprints, bounded caching, offline bundles, and explicit wire-codec composition.

- `github.com/faustbrian/go-schema-registry/providers/confluent`: Adapt the provider-neutral schema-registry contract to bounded Confluent-compatible REST, subject/version, compatibility, deletion, and version-0 wire semantics.

- `github.com/faustbrian/go-schema-registry/providers/glue`: Adapt the provider-neutral schema-registry contract to bounded AWS Glue registry identity, lifecycle, service, and uncompressed header-version-3 wire semantics.

- `github.com/faustbrian/go-webhook`: Provide protocol-independent webhook signing, exact-byte verification, replay protection, bounded delivery, and explicit adapters for durable and observable integrations.

- `github.com/faustbrian/go-wsdl`: Provide bounded WSDL 1.1 and WSDL 2.0 document modeling, parsing, validation, injected resolution, immutable compilation, composition, code-generation models, and semantic diffing.

- `github.com/faustbrian/go-xsd`: Parse, compile, validate, serialize, and build bounded XML Schema 1.0 documents without implicit external I/O.

## persistence-and-durability

- `github.com/faustbrian/go-audit`: Provide immutable audit records, explicit delivery and redaction policy, integrity verification, bounded query and export contracts, and retention planning without coupling consumers to a storage backend.

- `github.com/faustbrian/go-audit/postgres`: Persist audit records in append-only PostgreSQL storage with idempotent insertion, bounded query and export behavior, caller-owned transaction staging, and legal-hold-aware retention.

- `github.com/faustbrian/go-cache`: Provide typed cache semantics, bounded cache-aside loading, explicit key and codec policy, and backend-neutral ownership contracts.

- `github.com/faustbrian/go-event-sourcing`: Provide event-store and dispatcher contracts, aggregate repositories, codecs, upcasting, projections, snapshots, process managers, and deterministic test support.

- `github.com/faustbrian/go-event-sourcing/adapters/gokafka`: Adapt event-sourcing deliveries to Kafka records, synchronous dispatch, consumer handling, explicit poison and retry disposition, and dead-letter publication.

- `github.com/faustbrian/go-event-sourcing/adapters/gotelemetry`: Instrument event-sourcing dispatch, storage, serialization, snapshots, projections, process managers, and Kafka context propagation with caller-supplied OpenTelemetry components.

- `github.com/faustbrian/go-event-sourcing/adapters/outbox`: Atomically stage event rows and transactional-outbox envelopes through a savepoint in an existing caller-owned PostgreSQL transaction.

- `github.com/faustbrian/go-event-sourcing/adapters/queue`: Adapt complete event-sourcing deliveries to bounded queue envelopes, synchronous enqueue dispatch, and explicit live or replay task handling.

- `github.com/faustbrian/go-event-sourcing/postgres`: Provide PostgreSQL event, snapshot, projection, checkpoint, migration, and caller-owned transaction-writer implementations for event sourcing.

- `github.com/faustbrian/go-feature-flags`: Provide deterministic tenant-safe feature evaluation, versioned management, durable providers, bounded caching, and explicit fleet refresh.

- `github.com/faustbrian/go-idempotency`: Provide durable operation ownership, fencing, bounded result replay, canonical fingerprints, and explicit transport and workload adapters.

- `github.com/faustbrian/go-lease`: Provide backend-time distributed leases, managed renewal, monotonically increasing fencing tokens, and explicit worker and service coordination over memory, PostgreSQL, and Valkey backends.

- `github.com/faustbrian/go-migrations`: Provide engine-neutral migration identity, validation, deterministic planning, execution coordination, recovery, and a PostgreSQL ledger and backend.

- `github.com/faustbrian/go-postgres`: Provide finite pgx pool configuration and lifecycle, bounded transaction cleanup, SQLSTATE classification, safe observations, and real PostgreSQL test support.

- `github.com/faustbrian/go-queue`: Provide backend-neutral bounded worker coordination plus explicit in-memory, Redis, Valkey, NATS, and NSQ queue implementations with observable delivery, settlement, failure, and lifecycle semantics.

- `github.com/faustbrian/go-queue-control-plane`: Provide an authenticated administrative control plane for durable queue commands, desired state, audit history, fleet visibility, and optional Kubernetes workload scaling.

- `github.com/faustbrian/go-queue/queueservice`: Integrate caller-selected queue producers and workers with go-service startup, readiness, supervision, admission closure, drain, and shutdown while preserving backend-owned delivery semantics.

- `github.com/faustbrian/go-queue/rabbitmq`: Adapt the backend-neutral go-queue worker contract to RabbitMQ AMQP 0-9-1 publishing, consumption, recovery, settlement, retry, dead-letter, and shutdown policy owned by go-rabbitmq-queues.

- `github.com/faustbrian/go-scheduler`: Provide code-defined recurring schedules, immutable compilation, bounded execution, fenced multi-replica coordination, explicit dispatch, and observable lifecycle integration.

- `github.com/faustbrian/go-sequencer`: Provide dependency-ordered durable operation execution with immutable plans, fenced attempts, explicit retry and unknown-outcome policy, bounded fleet lifecycle, persistent stores, and selectable integration adapters.

- `github.com/faustbrian/go-settings`: Provide typed runtime-mutable settings, explicit precedence, immutable snapshots, optimistic writes, audit history, schema evolution, and optional persistence and cache adapters.

- `github.com/faustbrian/go-transactional-outbox`: Own PostgreSQL transactional outbox persistence, durable claims and state transitions, and caller-bounded at-least-once relay execution.

- `github.com/faustbrian/go-transactional-outbox/adapters/gokafka`: Map one durable outbox envelope to one confirmed first-party Kafka record without acquiring or owning the producer.

- `github.com/faustbrian/go-transactional-outbox/adapters/gorabbitstream`: Map one durable outbox envelope to one confirmed RabbitMQ Stream or Super Stream message without owning transport lifecycle.

- `github.com/faustbrian/go-transactional-outbox/adapters/otel`: Add bounded outbox semantic spans, metrics, propagation, observations, and publisher instrumentation without owning telemetry infrastructure.

- `github.com/faustbrian/go-transactional-outbox/adapters/queue`: Map one durable outbox envelope to one bounded deterministic first-party queue task while preserving acceptance ambiguity.

- `github.com/faustbrian/go-workflow`: Provide immutable workflow definitions and history, deterministic orchestration decisions, bounded durable work processing, explicit recovery semantics, and a PostgreSQL persistence adapter.

## resilience

- `github.com/faustbrian/go-adaptive-throttle`: Provide process-local rolling overload history and probabilistic admission that sheds a bounded share of work while preserving probe flow.

- `github.com/faustbrian/go-bulkhead`: Provide process-local fixed-capacity resource isolation with weighted permits, bounded FIFO waiting, explicit partitions, and graceful drain.

- `github.com/faustbrian/go-circuit-breaker`: Provide protocol-neutral, bounded circuit-breaker state, dependency-health admission, rolling outcome windows, and explicit permit and observation lifecycles.

- `github.com/faustbrian/go-concurrency-limit`: Provide bounded, process-local adaptive in-flight admission that learns a safe concurrency limit from explicit execution outcomes.

- `github.com/faustbrian/go-fault-injection`: Provide deterministic, bounded fault schedules, failure wrappers, and a fail-closed runtime for tests and explicitly authorized controlled experiments.

- `github.com/faustbrian/go-hedge`: Provide finite delayed duplicate attempts for explicitly replay-safe work under caller-owned deadlines and bounded amplification.

- `github.com/faustbrian/go-rate-limit`: Provide transport-neutral inbound admission policies, bounded memory and distributed backends, concurrency leases, and explicit HTTP, RPC, queue, principal, logging, and telemetry integrations.

- `github.com/faustbrian/go-resilience`: Provide deterministic generic policy composition, typed outcomes, caller-owned total deadlines, bounded observation, and shared process-local retry and hedge work budgets.

- `github.com/faustbrian/go-retry`: Provide bounded retry execution, explicit failure classification, finite time and work budgets, deterministic backoff strategies, and focused transport and observability adapters.

- `github.com/faustbrian/go-semaphore`: Provide a process-local FIFO weighted semaphore with bounded waiting, exactly-once permits, deterministic shutdown, and bounded observation events.

## observability

- `github.com/faustbrian/go-log`: Provide composable log/slog handlers for routing, redaction, sampling, bounded asynchronous delivery, capture, local rotation, and OpenTelemetry correlation.

- `github.com/faustbrian/go-telemetry`: Provide a vendor-neutral OpenTelemetry runtime, OTLP exporters, propagation, sampling, bounded instrumentation, and explicit provider lifecycle for Go services.

## integration-and-data-movement

- `github.com/faustbrian/go-external-sort`: Provide bounded external sorting of fixed-width opaque records using authenticated encrypted temporary storage.

- `github.com/faustbrian/go-filesystem`: Provide capability-oriented streaming filesystem contracts, backend-specific adapters, composable decorators, and conformance helpers.

- `github.com/faustbrian/go-http-client`: Provide typed outbound HTTP policy with finite transport defaults, immutable request specifications, deterministic middleware, and explicit response ownership.

- `github.com/faustbrian/go-kafka`: Provide bounded first-party Apache Kafka producer, consumer, inspection, replay, and transaction policy over franz-go.

- `github.com/faustbrian/go-kafka/adapters/gotelemetry`: Translate Kafka observation and propagation contracts into OpenTelemetry spans and metrics.

- `github.com/faustbrian/go-kafka/adapters/mskiam`: Provide Kafka SASL/OAUTHBEARER authentication configuration through AWS MSK IAM credentials and signing.

- `github.com/faustbrian/go-kafka/kafkaservice`: Bridge go-kafka consumer and producer resources into the explicit go-service lifecycle.

- `github.com/faustbrian/go-rabbitmq-queues`: Provide bounded RabbitMQ-native AMQP 0-9-1 classic and quorum queue publishing, consumption, settlement, recovery, topology verification, health, and observation policy.

- `github.com/faustbrian/go-rabbitmq-streams`: Provide vendor-neutral bounded policy for RabbitMQ Streams messages, publishing, consumption, replay, inspection, failures, lifecycle, and observations.

- `github.com/faustbrian/go-rabbitmq-streams/otel`: Translate bounded RabbitMQ Streams observations into caller-owned OpenTelemetry metrics and propagate W3C Trace Context through message headers.

- `github.com/faustbrian/go-rabbitmq-streams/rabbitmq`: Adapt the RabbitMQ-supported Go Streams client to bounded rabbitstream policy while owning protocol connections, sessions, cursors, recovery, and wire conversion.

- `github.com/faustbrian/go-search`: Provide backend-neutral contracts for bounded document indexing, typed querying, cursor pagination, schema migration, projections, and reconciliation while treating application data as authoritative.

- `github.com/faustbrian/go-search/adapters/opensearch`: Translate the backend-neutral search contract to a bounded OpenSearch client with explicit transport, trust, lifecycle, resilience, and observability policy.

- `github.com/faustbrian/go-secret-envelope`: Provide bounded authenticated secret envelopes with explicit key-provider boundaries, immutable encryption context, and versioned persistence bytes.

- `github.com/faustbrian/go-tabular`: Provide explicit, bounded ingestion for delimited, fixed-width, XLS, XLSX, and ZIP-backed tabular sources.

- `github.com/faustbrian/go-wire`: Provide explicit, bounded JSON, XML, SOAP, YAML, TOML, MessagePack, CBOR, and BSON encoding and decoding boundaries.

## domain-utilities

- `github.com/faustbrian/go-barcode`: Provide immutable logical barcode symbols, strict validation, encoding, decoding, rendering, and standards-conformance evidence.

- `github.com/faustbrian/go-calendar`: Provide immutable civil dates, Gregorian arithmetic, typed periods, explicit DST conversion, and bounded business calendars.

- `github.com/faustbrian/go-ecma-regexp`: Provide bounded ECMAScript regular-expression parsing, compilation, and matching, including the JSON Schema pattern profile.

- `github.com/faustbrian/go-geo`: Provide immutable geospatial values, bounded geometry and geodesy operations, interoperable codecs, PostGIS mapping, and deterministic test helpers.

- `github.com/faustbrian/go-keyphrase`: Provide unbiased bounded password and passphrase generation, BIP-39 mnemonic interoperability, immutable word lists, and explicit secret-handling boundaries.

- `github.com/faustbrian/go-knapsack`: Provide deterministic, bounded offline orthogonal packing, exact objectives, extension constraints, canonical plans, and independent verification.

- `github.com/faustbrian/go-knapsack/objective/gomoney`: Adapt exact go-money values into deterministic Knapsack container-cost objective evaluation.

- `github.com/faustbrian/go-math`: Provide immutable arbitrary-precision integer, rational, decimal, and binary-float values with explicit precision, rounding, limits, conditions, and deterministic encodings.

- `github.com/faustbrian/go-measurement`: Provide immutable, exact, unit-safe quantities, dimensions, conversions, logistics formulas, and bounded wire encodings.

- `github.com/faustbrian/go-merkle-patricia-trie`: Provide bounded immutable Ethereum modified Merkle Patricia tries, roots, proofs, storage integration, retention, pruning, and recovery.

- `github.com/faustbrian/go-merkle-tree`: Provide storage-independent ordered Merkle trees, canonical and RFC 9162 profiles, immutable snapshots, and bounded inclusion, multi-inclusion, and consistency proofs.

- `github.com/faustbrian/go-money`: Provide immutable exact monetary values, explicit precision and rounding contexts, bounded arithmetic, allocation, tax, discount, conversion, formatting, and persistence encodings.

- `github.com/faustbrian/go-opening-hours`: Model immutable recurring opening hours, dated exceptions, timezone-aware availability, and bounded schedule composition.

- `github.com/faustbrian/go-rule-engine`: Provide deterministic typed rule construction, bounded compilation into immutable plans, explicit fact evaluation, canonical JSON, and redacted diagnostics without hidden I/O.

- `github.com/faustbrian/go-rule-engine/adapters/math`: Bridge immutable go-math decimals into rule-engine tagged values and deterministic equality and ordering operators without coupling the core module to decimal arithmetic.

- `github.com/faustbrian/go-rule-engine/adapters/measurement`: Bridge immutable go-measurement quantities into tagged rule-engine values and dimension-safe deterministic comparison operators.

- `github.com/faustbrian/go-rule-engine/adapters/temporal`: Bridge exact go-temporal instants and periods into tagged rule-engine values and deterministic relation operators.

- `github.com/faustbrian/go-state-machine`: Provide deterministic typed state-machine compilation, transition selection, inert effect planning, replay and evolution, with optional explicit execution, persistence, outbox delivery, and diagram rendering.

- `github.com/faustbrian/go-temporal`: Provide immutable temporal algebra, explicit interval bounds and relations, normalized sets, fixed durations, daily intervals, strict notation, and versioned encodings.

- `github.com/faustbrian/go-verkle-tree`: Provide bounded immutable authenticated key/value trees, roots, proofs, witnesses, stateless updates, and caller-owned storage protocols for the package-owned Bandersnatch IPA profile.

## tooling

- `github.com/faustbrian/go-analysis`: Provide deterministic go/analysis policies, governed rule metadata, and bounded JSON and SARIF reports for Go repositories.

- `github.com/faustbrian/go-cli`: Provide explicit typed command construction, immutable compilation, deterministic parsing, lifecycle middleware, bounded output, and stable process-facing results.

- `github.com/faustbrian/go-prompts`: Provide typed interactive prompts, deterministic non-interactive parsing, semantic rendering, caller-driven presentation, and an explicit terminal adapter.
