# Golib Consumer Catalog

Design language `1.0` (`v1.8.3`); tooling `v1.8.3`.

## foundations

- [github.com/faustbrian/go-clock](https://pkg.go.dev/github.com/faustbrian/go-clock@v1.1.0) — **active**, `v1.1.0`: Provide explicit wall-time and elapsed-time seams, owned timers and tickers, cancelable sleep, deterministic manual time, and synchronization-aware clock test helpers. ([README](https://github.com/faustbrian/go-clock/blob/v1.1.0/README.md))

- [github.com/faustbrian/go-config](https://pkg.go.dev/github.com/faustbrian/go-config@v1.1.0) — **active**, `v1.1.0`: Load explicit, layered configuration sources into immutable typed snapshots with deterministic precedence, validation, safe provenance, and redacted secrets. ([README](https://github.com/faustbrian/go-config/blob/v1.1.0/README.md))

- [github.com/faustbrian/go-config/adapters/awssecretsmanager](https://pkg.go.dev/github.com/faustbrian/go-config/adapters/awssecretsmanager@v1.0.0) — **active**, `v1.0.0`: Load one bounded JSON configuration document from AWS Secrets Manager as an explicit sensitive configuration source. ([README](https://github.com/faustbrian/go-config/blob/adapters/awssecretsmanager/v1.0.0/adapters/awssecretsmanager/README.md))

- [github.com/faustbrian/go-correlation](https://pkg.go.dev/github.com/faustbrian/go-correlation@v1.1.0) — **active**, `v1.1.0`: Provide immutable correlation, request, causation, and external identifiers with explicit trust and disclosure across transport boundaries. ([README](https://github.com/faustbrian/go-correlation/blob/v1.1.0/README.md))

- [github.com/faustbrian/go-identifier](https://pkg.go.dev/github.com/faustbrian/go-identifier@v1.0.0) — **active**, `v1.0.0`: Provide immutable UUID, ULID, TypeID, KSUID, NanoID, typed domain identifier, and deterministic slug values with explicit generation, validation, encoding, ordering, and leakage contracts. ([README](https://github.com/faustbrian/go-identifier/blob/v1.0.0/README.md))

- [github.com/faustbrian/go-international](https://pkg.go.dev/github.com/faustbrian/go-international@v1.1.0) — **active**, `v1.1.0`: Provide immutable offline international reference values, strict parsing, explicit normalization, and versioned dataset provenance. ([README](https://github.com/faustbrian/go-international/blob/v1.1.0/README.md))

- [github.com/faustbrian/go-localized](https://pkg.go.dev/github.com/faustbrian/go-localized@v1.1.0) — **active**, `v1.1.0`: Provide immutable localized text values, exact locale lookup, deterministic matching and fallback, and focused configuration, encoding, HTTP, persistence, validation, and wire integrations. ([README](https://github.com/faustbrian/go-localized/blob/v1.1.0/README.md))

- [github.com/faustbrian/go-tenancy](https://pkg.go.dev/github.com/faustbrian/go-tenancy@v1.1.0) — **active**, `v1.1.0`: Provide validated tenant identity, explicit tenant, system, and unscoped work, and fail-closed propagation and enforcement seams. ([README](https://github.com/faustbrian/go-tenancy/blob/v1.1.0/README.md))

- [github.com/faustbrian/go-validation](https://pkg.go.dev/github.com/faustbrian/go-validation@v1.1.0) — **active**, `v1.1.0`: Provide deterministic typed application validation, bounded diagnostics, explicit presence semantics, reusable rules, and transport-neutral report composition. ([README](https://github.com/faustbrian/go-validation/blob/v1.1.0/README.md))

## service-edge

- [github.com/faustbrian/go-api-query](https://pkg.go.dev/github.com/faustbrian/go-api-query@v1.1.0) — **active**, `v1.1.0`: Compile declared API query capabilities into immutable, bounded, storage-neutral plans and adapt reviewed plans to explicit transport, validation, and persistence boundaries. ([README](https://github.com/faustbrian/go-api-query/blob/v1.1.0/README.md))

- [github.com/faustbrian/go-authentication](https://pkg.go.dev/github.com/faustbrian/go-authentication@v1.2.0) — **active**, `v1.2.0`: Turn Basic credentials, opaque bearer tokens, and API keys into immutable authenticated principals. ([README](https://github.com/faustbrian/go-authentication/blob/v1.2.0/README.md))

- [github.com/faustbrian/go-authentication/adapters/otel](https://pkg.go.dev/github.com/faustbrian/go-authentication/adapters/otel@v1.0.0) — **active**, `v1.0.0`: Adapt bounded authentication observations to OpenTelemetry traces and metrics. ([README](https://github.com/faustbrian/go-authentication/blob/adapters/otel/v1.0.0/adapters/otel/README.md))

- [github.com/faustbrian/go-authentication/authotel](https://pkg.go.dev/github.com/faustbrian/go-authentication/authotel@v1.1.0) — **deprecated**, `v1.1.0`: Adapt bounded authentication observations to OpenTelemetry traces and metrics. ([README](https://github.com/faustbrian/go-authentication/blob/authotel/v1.1.0/authotel/README.md)) ([migration](https://github.com/faustbrian/go-authentication/blob/authotel/v1.1.0/authotel/docs/README.md))

- [github.com/faustbrian/go-authentication/jwt](https://pkg.go.dev/github.com/faustbrian/go-authentication/jwt@v1.1.0) — **active**, `v1.1.0`: Validate signed compact JWTs and operate bounded static or remote JWK sets at an authentication boundary. ([README](https://github.com/faustbrian/go-authentication/blob/jwt/v1.1.0/jwt/README.md))

- [github.com/faustbrian/go-authentication/oidc](https://pkg.go.dev/github.com/faustbrian/go-authentication/oidc@v1.0.0) — **active**, `v1.0.0`: Discover OpenID Providers and validate signed OpenID Connect ID tokens at the authentication trust boundary. ([README](https://github.com/faustbrian/go-authentication/blob/oidc/v1.0.0/oidc/README.md))

- [github.com/faustbrian/go-authorization](https://pkg.go.dev/github.com/faustbrian/go-authorization@v1.1.0) — **active**, `v1.1.0`: Provide typed ACL, RBAC, and ABAC policy evaluation, immutable revisioned snapshots, bounded policy compilation, fail-closed transport integration, and explicit persistence and invalidation adapters. ([README](https://github.com/faustbrian/go-authorization/blob/v1.1.0/README.md))

- [github.com/faustbrian/go-capability](https://pkg.go.dev/github.com/faustbrian/go-capability@v1.1.0) — **active**, `v1.1.0`: Issue and verify narrowly scoped, tamper-evident, expiring capabilities and signed URLs with explicit replay, revocation, and storage boundaries. ([README](https://github.com/faustbrian/go-capability/blob/v1.1.0/README.md))

- [github.com/faustbrian/go-http-middleware](https://pkg.go.dev/github.com/faustbrian/go-http-middleware@v1.0.0) — **active**, `v1.0.0`: Provide explicit, bounded server-side net/http middleware and deterministic chain composition without hidden registration or defaults. ([README](https://github.com/faustbrian/go-http-middleware/blob/v1.0.0/README.md))

- [github.com/faustbrian/go-password](https://pkg.go.dev/github.com/faustbrian/go-password@v1.1.0) — **active**, `v1.1.0`: Provide bounded Argon2id and bcrypt password hashing, verification, parsing, admission, and login-time upgrade primitives. ([README](https://github.com/faustbrian/go-password/blob/v1.1.0/README.md))

- [github.com/faustbrian/go-router](https://pkg.go.dev/github.com/faustbrian/go-router@v1.0.0) — **active**, `v1.0.0`: Provide explicit startup-time HTTP route composition, immutable compiled dispatch, safe URL generation, and deterministic route introspection over ordinary net/http handlers. ([README](https://github.com/faustbrian/go-router/blob/v1.0.0/README.md))

- [github.com/faustbrian/go-service](https://pkg.go.dev/github.com/faustbrian/go-service@v1.1.0) — **active**, `v1.1.0`: Coordinate explicit service construction, ordered lifecycle, supervision, health probes, HTTP serving, maintenance state, and caller-owned integration hooks without imposing an application framework. ([README](https://github.com/faustbrian/go-service/blob/v1.1.0/README.md))

## protocols-and-descriptions

- [github.com/faustbrian/go-cloudevents](https://pkg.go.dev/github.com/faustbrian/go-cloudevents@v1.1.0) — **active**, `v1.1.0`: Provide a transport-independent CloudEvents 1.0 envelope with bounded JSON, HTTP, and Kafka mappings. ([README](https://github.com/faustbrian/go-cloudevents/blob/v1.1.0/README.md))

- [github.com/faustbrian/go-cloudevents/adapters/audit](https://pkg.go.dev/github.com/faustbrian/go-cloudevents/adapters/audit@v1.0.0) — **active**, `v1.0.0`: Provide an explicit target-oriented CloudEvents boundary for Golib audit metadata. ([README](https://github.com/faustbrian/go-cloudevents/blob/adapters/audit/v1.0.0/adapters/audit/README.md))

- [github.com/faustbrian/go-cloudevents/adapters/correlation](https://pkg.go.dev/github.com/faustbrian/go-cloudevents/adapters/correlation@v1.0.0) — **active**, `v1.0.0`: Provide an explicit target-oriented CloudEvents boundary for Golib correlation identifiers. ([README](https://github.com/faustbrian/go-cloudevents/blob/adapters/correlation/v1.0.0/adapters/correlation/README.md))

- [github.com/faustbrian/go-cloudevents/adapters/event-sourcing](https://pkg.go.dev/github.com/faustbrian/go-cloudevents/adapters/event-sourcing@v1.0.0) — **active**, `v1.0.0`: Provide an explicit target-oriented CloudEvents boundary for Golib event-sourcing messages. ([README](https://github.com/faustbrian/go-cloudevents/blob/adapters/event-sourcing/v1.0.0/adapters/event-sourcing/README.md))

- [github.com/faustbrian/go-cloudevents/adapters/golib](https://pkg.go.dev/github.com/faustbrian/go-cloudevents/adapters/golib@v1.1.0) — **deprecated**, `v1.1.0`: Preserve v1 compatibility for the former broad CloudEvents bridge while callers migrate to independently versioned target adapters. ([README](https://github.com/faustbrian/go-cloudevents/blob/adapters/golib/v1.1.0/adapters/golib/README.md)) ([migration](https://github.com/faustbrian/go-cloudevents/blob/adapters/golib/v1.1.0/adapters/golib/docs/reference.md))

- [github.com/faustbrian/go-cloudevents/adapters/jsonschema](https://pkg.go.dev/github.com/faustbrian/go-cloudevents/adapters/jsonschema@v1.0.0) — **active**, `v1.0.0`: Provide an explicit target-oriented CloudEvents boundary for Golib compiled JSON Schema validation. ([README](https://github.com/faustbrian/go-cloudevents/blob/adapters/jsonschema/v1.0.0/adapters/jsonschema/README.md))

- [github.com/faustbrian/go-cloudevents/adapters/kafka](https://pkg.go.dev/github.com/faustbrian/go-cloudevents/adapters/kafka@v1.0.0) — **active**, `v1.0.0`: Provide an explicit target-oriented CloudEvents boundary for Golib Kafka records. ([README](https://github.com/faustbrian/go-cloudevents/blob/adapters/kafka/v1.0.0/adapters/kafka/README.md))

- [github.com/faustbrian/go-cloudevents/adapters/outbox](https://pkg.go.dev/github.com/faustbrian/go-cloudevents/adapters/outbox@v1.0.0) — **active**, `v1.0.0`: Provide an explicit target-oriented CloudEvents boundary for Golib transactional outbox envelopes. ([README](https://github.com/faustbrian/go-cloudevents/blob/adapters/outbox/v1.0.0/adapters/outbox/README.md))

- [github.com/faustbrian/go-cloudevents/adapters/queue](https://pkg.go.dev/github.com/faustbrian/go-cloudevents/adapters/queue@v1.0.0) — **active**, `v1.0.0`: Provide an explicit target-oriented CloudEvents boundary for Golib queue jobs. ([README](https://github.com/faustbrian/go-cloudevents/blob/adapters/queue/v1.0.0/adapters/queue/README.md))

- [github.com/faustbrian/go-cloudevents/adapters/rabbitstream](https://pkg.go.dev/github.com/faustbrian/go-cloudevents/adapters/rabbitstream@v1.0.0) — **active**, `v1.0.0`: Provide an explicit target-oriented CloudEvents boundary for Golib RabbitMQ Streams messages. ([README](https://github.com/faustbrian/go-cloudevents/blob/adapters/rabbitstream/v1.0.0/adapters/rabbitstream/README.md))

- [github.com/faustbrian/go-cloudevents/adapters/schema-registry](https://pkg.go.dev/github.com/faustbrian/go-cloudevents/adapters/schema-registry@v1.0.0) — **active**, `v1.0.0`: Provide an explicit target-oriented CloudEvents boundary for Golib schema-registry JSON Schema resolution. ([README](https://github.com/faustbrian/go-cloudevents/blob/adapters/schema-registry/v1.0.0/adapters/schema-registry/README.md))

- [github.com/faustbrian/go-cloudevents/adapters/telemetry](https://pkg.go.dev/github.com/faustbrian/go-cloudevents/adapters/telemetry@v1.0.0) — **active**, `v1.0.0`: Provide an explicit target-oriented CloudEvents boundary for Golib telemetry trace context. ([README](https://github.com/faustbrian/go-cloudevents/blob/adapters/telemetry/v1.0.0/adapters/telemetry/README.md))

- [github.com/faustbrian/go-cloudevents/adapters/tenancy](https://pkg.go.dev/github.com/faustbrian/go-cloudevents/adapters/tenancy@v1.0.0) — **active**, `v1.0.0`: Provide an explicit target-oriented CloudEvents boundary for Golib tenant routing identity. ([README](https://github.com/faustbrian/go-cloudevents/blob/adapters/tenancy/v1.0.0/adapters/tenancy/README.md))

- [github.com/faustbrian/go-cloudevents/adapters/workflow](https://pkg.go.dev/github.com/faustbrian/go-cloudevents/adapters/workflow@v1.0.0) — **active**, `v1.0.0`: Provide an explicit target-oriented CloudEvents boundary for Golib durable workflow history. ([README](https://github.com/faustbrian/go-cloudevents/blob/adapters/workflow/v1.0.0/adapters/workflow/README.md))

- [github.com/faustbrian/go-http-signature](https://pkg.go.dev/github.com/faustbrian/go-http-signature@v1.0.0) — **active**, `v1.0.0`: Provide bounded HTTP Message Signatures, digest fields, structured-field interpretation, explicit signing and verification policy, key resolution, replay consumption, and HTTP adapters. ([README](https://github.com/faustbrian/go-http-signature/blob/v1.0.0/README.md))

- [github.com/faustbrian/go-json-schema](https://pkg.go.dev/github.com/faustbrian/go-json-schema@v1.0.0) — **active**, `v1.0.0`: Compile and validate exact-number, multi-dialect JSON Schema documents with explicit resource loading and bounded execution. ([README](https://github.com/faustbrian/go-json-schema/blob/v1.0.0/README.md))

- [github.com/faustbrian/go-jsonapi](https://pkg.go.dev/github.com/faustbrian/go-jsonapi@v1.0.0) — **active**, `v1.0.0`: Provide strict framework-neutral JSON:API encoding, decoding, validation, negotiation, query handling, Atomic Operations, and Cursor Pagination primitives. ([README](https://github.com/faustbrian/go-jsonapi/blob/v1.0.0/README.md))

- [github.com/faustbrian/go-jsonrpc](https://pkg.go.dev/github.com/faustbrian/go-jsonrpc@v1.0.0) — **active**, `v1.0.0`: Provide bounded transport-neutral JSON-RPC 2.0 envelopes, dispatch, clients, middleware, hooks, and an optional HTTP binding. ([README](https://github.com/faustbrian/go-jsonrpc/blob/v1.0.0/README.md))

- [github.com/faustbrian/go-openapi](https://pkg.go.dev/github.com/faustbrian/go-openapi@v1.0.0) — **active**, `v1.0.0`: Provide immutable, lossless, and resource-bounded Swagger 2.0 and OpenAPI 3.0 through 3.2 document modeling, parsing, validation, reference resolution, composition, conversion, compatibility diffing, and serialization. ([README](https://github.com/faustbrian/go-openapi/blob/v1.0.0/README.md))

- [github.com/faustbrian/go-openrpc](https://pkg.go.dev/github.com/faustbrian/go-openrpc@v1.0.0) — **active**, `v1.0.0`: Provide bounded OpenRPC 1.3.x and 1.4.x document modeling, parsing, canonical serialization, validation, reference resolution, runtime expressions, discovery, composition, compatibility diffing, and JSON-RPC integration. ([README](https://github.com/faustbrian/go-openrpc/blob/v1.0.0/README.md))

- [github.com/faustbrian/go-schema-registry](https://pkg.go.dev/github.com/faustbrian/go-schema-registry@v1.0.0) — **active**, `v1.0.0`: Provide provider-neutral schema registration, resolution, compatibility, canonicalization, portable fingerprints, bounded caching, offline bundles, and explicit wire-codec composition. ([README](https://github.com/faustbrian/go-schema-registry/blob/v1.0.0/README.md))

- [github.com/faustbrian/go-schema-registry/providers/confluent](https://pkg.go.dev/github.com/faustbrian/go-schema-registry/providers/confluent@v1.0.0) — **active**, `v1.0.0`: Adapt the provider-neutral schema-registry contract to bounded Confluent-compatible REST, subject/version, compatibility, deletion, and version-0 wire semantics. ([README](https://github.com/faustbrian/go-schema-registry/blob/providers/confluent/v1.0.0/providers/confluent/README.md))

- [github.com/faustbrian/go-schema-registry/providers/glue](https://pkg.go.dev/github.com/faustbrian/go-schema-registry/providers/glue@v1.0.0) — **active**, `v1.0.0`: Adapt the provider-neutral schema-registry contract to bounded AWS Glue registry identity, lifecycle, service, and uncompressed header-version-3 wire semantics. ([README](https://github.com/faustbrian/go-schema-registry/blob/providers/glue/v1.0.0/providers/glue/README.md))

- [github.com/faustbrian/go-webhook](https://pkg.go.dev/github.com/faustbrian/go-webhook@v1.0.0) — **active**, `v1.0.0`: Provide protocol-independent webhook signing, exact-byte verification, replay protection, bounded delivery, and explicit adapters for durable and observable integrations. ([README](https://github.com/faustbrian/go-webhook/blob/v1.0.0/README.md))

- [github.com/faustbrian/go-wsdl](https://pkg.go.dev/github.com/faustbrian/go-wsdl@v1.0.0) — **active**, `v1.0.0`: Provide bounded WSDL 1.1 and WSDL 2.0 document modeling, parsing, validation, injected resolution, immutable compilation, composition, code-generation models, and semantic diffing. ([README](https://github.com/faustbrian/go-wsdl/blob/v1.0.0/README.md))

- [github.com/faustbrian/go-xsd](https://pkg.go.dev/github.com/faustbrian/go-xsd@v1.0.0) — **active**, `v1.0.0`: Parse, compile, validate, serialize, and build bounded XML Schema 1.0 documents without implicit external I/O. ([README](https://github.com/faustbrian/go-xsd/blob/v1.0.0/README.md))

## persistence-and-durability

- [github.com/faustbrian/go-audit](https://pkg.go.dev/github.com/faustbrian/go-audit@v1.0.0) — **active**, `v1.0.0`: Provide immutable audit records, explicit delivery and redaction policy, integrity verification, bounded query and export contracts, and retention planning without coupling consumers to a storage backend. ([README](https://github.com/faustbrian/go-audit/blob/v1.0.0/README.md))

- [github.com/faustbrian/go-audit/postgres](https://pkg.go.dev/github.com/faustbrian/go-audit/postgres@v1.0.0) — **active**, `v1.0.0`: Persist audit records in append-only PostgreSQL storage with idempotent insertion, bounded query and export behavior, caller-owned transaction staging, and legal-hold-aware retention. ([README](https://github.com/faustbrian/go-audit/blob/postgres/v1.0.0/postgres/README.md))

- [github.com/faustbrian/go-cache](https://pkg.go.dev/github.com/faustbrian/go-cache@v1.0.0) — **active**, `v1.0.0`: Provide typed cache semantics, bounded cache-aside loading, explicit key and codec policy, and backend-neutral ownership contracts. ([README](https://github.com/faustbrian/go-cache/blob/v1.0.0/README.md))

- [github.com/faustbrian/go-event-sourcing](https://pkg.go.dev/github.com/faustbrian/go-event-sourcing@v1.0.0) — **active**, `v1.0.0`: Provide event-store and dispatcher contracts, aggregate repositories, codecs, upcasting, projections, snapshots, process managers, and deterministic test support. ([README](https://github.com/faustbrian/go-event-sourcing/blob/v1.0.0/README.md))

- [github.com/faustbrian/go-event-sourcing/adapters/gokafka](https://pkg.go.dev/github.com/faustbrian/go-event-sourcing/adapters/gokafka@v1.0.3) — **deprecated**, `v1.0.3`: Adapt event-sourcing deliveries to Kafka records, synchronous dispatch, consumer handling, explicit poison and retry disposition, and dead-letter publication. ([README](https://github.com/faustbrian/go-event-sourcing/blob/adapters/gokafka/v1.0.3/adapters/gokafka/README.md)) ([migration](https://github.com/faustbrian/go-event-sourcing/blob/adapters/gokafka/v1.0.3/adapters/gokafka/docs/reference.md))

- [github.com/faustbrian/go-event-sourcing/adapters/gotelemetry](https://pkg.go.dev/github.com/faustbrian/go-event-sourcing/adapters/gotelemetry@v1.0.3) — **deprecated**, `v1.0.3`: Instrument event-sourcing dispatch, storage, serialization, snapshots, projections, process managers, and Kafka context propagation with caller-supplied OpenTelemetry components. ([README](https://github.com/faustbrian/go-event-sourcing/blob/adapters/gotelemetry/v1.0.3/adapters/gotelemetry/README.md)) ([migration](https://github.com/faustbrian/go-event-sourcing/blob/adapters/gotelemetry/v1.0.3/adapters/gotelemetry/docs/reference.md))

- [github.com/faustbrian/go-event-sourcing/adapters/kafka](https://pkg.go.dev/github.com/faustbrian/go-event-sourcing/adapters/kafka@v1.0.2) — **active**, `v1.0.2`: Adapt event-sourcing deliveries to Kafka records, synchronous dispatch, consumer handling, explicit poison and retry disposition, and dead-letter publication. ([README](https://github.com/faustbrian/go-event-sourcing/blob/adapters/kafka/v1.0.2/adapters/kafka/README.md))

- [github.com/faustbrian/go-event-sourcing/adapters/otel](https://pkg.go.dev/github.com/faustbrian/go-event-sourcing/adapters/otel@v1.0.2) — **active**, `v1.0.2`: Instrument event-sourcing dispatch, storage, serialization, snapshots, projections, process managers, and Kafka context propagation with caller-supplied OpenTelemetry components. ([README](https://github.com/faustbrian/go-event-sourcing/blob/adapters/otel/v1.0.2/adapters/otel/README.md))

- [github.com/faustbrian/go-event-sourcing/adapters/outbox](https://pkg.go.dev/github.com/faustbrian/go-event-sourcing/adapters/outbox@v1.0.0) — **active**, `v1.0.0`: Atomically stage event rows and transactional-outbox envelopes through a savepoint in an existing caller-owned PostgreSQL transaction. ([README](https://github.com/faustbrian/go-event-sourcing/blob/adapters/outbox/v1.0.0/adapters/outbox/README.md))

- [github.com/faustbrian/go-event-sourcing/adapters/queue](https://pkg.go.dev/github.com/faustbrian/go-event-sourcing/adapters/queue@v1.0.0) — **active**, `v1.0.0`: Adapt complete event-sourcing deliveries to bounded queue envelopes, synchronous enqueue dispatch, and explicit live or replay task handling. ([README](https://github.com/faustbrian/go-event-sourcing/blob/adapters/queue/v1.0.0/adapters/queue/README.md))

- [github.com/faustbrian/go-event-sourcing/postgres](https://pkg.go.dev/github.com/faustbrian/go-event-sourcing/postgres@v1.0.0) — **active**, `v1.0.0`: Provide PostgreSQL event, snapshot, projection, checkpoint, migration, and caller-owned transaction-writer implementations for event sourcing. ([README](https://github.com/faustbrian/go-event-sourcing/blob/postgres/v1.0.0/postgres/README.md))

- [github.com/faustbrian/go-feature-flags](https://pkg.go.dev/github.com/faustbrian/go-feature-flags@v1.0.0) — **active**, `v1.0.0`: Provide deterministic tenant-safe feature evaluation, versioned management, durable providers, bounded caching, and explicit fleet refresh. ([README](https://github.com/faustbrian/go-feature-flags/blob/v1.0.0/README.md))

- [github.com/faustbrian/go-idempotency](https://pkg.go.dev/github.com/faustbrian/go-idempotency@v1.1.0) — **active**, `v1.1.0`: Provide durable operation ownership, fencing, bounded result replay, canonical fingerprints, and explicit transport and workload adapters. ([README](https://github.com/faustbrian/go-idempotency/blob/v1.1.0/README.md))

- [github.com/faustbrian/go-lease](https://pkg.go.dev/github.com/faustbrian/go-lease@v1.1.0) — **active**, `v1.1.0`: Provide backend-time distributed leases, managed renewal, monotonically increasing fencing tokens, and explicit worker and service coordination over memory, PostgreSQL, and Valkey backends. ([README](https://github.com/faustbrian/go-lease/blob/v1.1.0/README.md))

- [github.com/faustbrian/go-migrations](https://pkg.go.dev/github.com/faustbrian/go-migrations@v1.1.0) — **active**, `v1.1.0`: Provide engine-neutral migration identity, validation, deterministic planning, execution coordination, recovery, and a PostgreSQL ledger and backend. ([README](https://github.com/faustbrian/go-migrations/blob/v1.1.0/README.md))

- [github.com/faustbrian/go-postgres](https://pkg.go.dev/github.com/faustbrian/go-postgres@v1.1.0) — **active**, `v1.1.0`: Provide finite pgx pool configuration and lifecycle, bounded transaction cleanup, SQLSTATE classification, safe observations, and real PostgreSQL test support. ([README](https://github.com/faustbrian/go-postgres/blob/v1.1.0/README.md))

- [github.com/faustbrian/go-queue](https://pkg.go.dev/github.com/faustbrian/go-queue@v1.1.0) — **active**, `v1.1.0`: Provide backend-neutral bounded worker coordination plus explicit in-memory, Redis, Valkey, NATS, and NSQ queue implementations with observable delivery, settlement, failure, and lifecycle semantics. ([README](https://github.com/faustbrian/go-queue/blob/v1.1.0/README.md))

- [github.com/faustbrian/go-queue-control-plane](https://pkg.go.dev/github.com/faustbrian/go-queue-control-plane@v1.1.0) — **active**, `v1.1.0`: Provide an authenticated administrative control plane for durable queue commands, desired state, audit history, fleet visibility, and optional Kubernetes workload scaling. ([README](https://github.com/faustbrian/go-queue-control-plane/blob/v1.1.0/README.md))

- [github.com/faustbrian/go-queue/adapters/rabbitmq](https://pkg.go.dev/github.com/faustbrian/go-queue/adapters/rabbitmq@v1.0.0) — **active**, `v1.0.0`: Adapt the backend-neutral go-queue worker contract to RabbitMQ AMQP 0-9-1 publishing, consumption, recovery, settlement, retry, dead-letter, and shutdown policy owned by go-rabbitmq-queues. ([README](https://github.com/faustbrian/go-queue/blob/adapters/rabbitmq/v1.0.0/adapters/rabbitmq/README.md))

- [github.com/faustbrian/go-queue/adapters/service](https://pkg.go.dev/github.com/faustbrian/go-queue/adapters/service@v1.0.0) — **active**, `v1.0.0`: Integrate caller-selected queue producers and workers with go-service startup, readiness, supervision, admission closure, drain, and shutdown while preserving backend-owned delivery semantics. ([README](https://github.com/faustbrian/go-queue/blob/adapters/service/v1.0.0/adapters/service/README.md))

- [github.com/faustbrian/go-queue/queueservice](https://pkg.go.dev/github.com/faustbrian/go-queue/queueservice@v1.0.1) — **deprecated**, `v1.0.1`: Preserve the released queueservice import path while delegating every public type and operation to adapters/service. ([README](https://github.com/faustbrian/go-queue/blob/queueservice/v1.0.1/queueservice/README.md)) ([migration](https://github.com/faustbrian/go-queue/blob/queueservice/v1.0.1/docs/migration.md))

- [github.com/faustbrian/go-queue/rabbitmq](https://pkg.go.dev/github.com/faustbrian/go-queue/rabbitmq@v1.0.1) — **deprecated**, `v1.0.1`: Preserve the released RabbitMQ import path while delegating every public type and operation to adapters/rabbitmq. ([README](https://github.com/faustbrian/go-queue/blob/rabbitmq/v1.0.1/rabbitmq/README.md)) ([migration](https://github.com/faustbrian/go-queue/blob/rabbitmq/v1.0.1/rabbitmq/README.md))

- [github.com/faustbrian/go-scheduler](https://pkg.go.dev/github.com/faustbrian/go-scheduler@v1.1.0) — **active**, `v1.1.0`: Provide code-defined recurring schedules, immutable compilation, bounded execution, fenced multi-replica coordination, explicit dispatch, and observable lifecycle integration. ([README](https://github.com/faustbrian/go-scheduler/blob/v1.1.0/README.md))

- [github.com/faustbrian/go-sequencer](https://pkg.go.dev/github.com/faustbrian/go-sequencer@v1.1.0) — **active**, `v1.1.0`: Provide dependency-ordered durable operation execution with immutable plans, fenced attempts, explicit retry and unknown-outcome policy, bounded fleet lifecycle, persistent stores, and selectable integration adapters. ([README](https://github.com/faustbrian/go-sequencer/blob/v1.1.0/README.md))

- [github.com/faustbrian/go-settings](https://pkg.go.dev/github.com/faustbrian/go-settings@v1.1.0) — **active**, `v1.1.0`: Provide typed runtime-mutable settings, explicit precedence, immutable snapshots, optimistic writes, audit history, schema evolution, and optional persistence and cache adapters. ([README](https://github.com/faustbrian/go-settings/blob/v1.1.0/README.md))

- [github.com/faustbrian/go-transactional-outbox](https://pkg.go.dev/github.com/faustbrian/go-transactional-outbox@v1.0.0) — **active**, `v1.0.0`: Own PostgreSQL transactional outbox persistence, durable claims and state transitions, and caller-bounded at-least-once relay execution. ([README](https://github.com/faustbrian/go-transactional-outbox/blob/v1.0.0/README.md))

- [github.com/faustbrian/go-transactional-outbox/adapters/gokafka](https://pkg.go.dev/github.com/faustbrian/go-transactional-outbox/adapters/gokafka@v1.0.0) — **deprecated**, `v1.0.0`: Preserve the released Kafka adapter contract while consumers migrate to the target-oriented adapters/kafka module. ([README](https://github.com/faustbrian/go-transactional-outbox/blob/adapters/gokafka/v1.0.0/adapters/gokafka/README.md)) ([migration](https://github.com/faustbrian/go-transactional-outbox/blob/adapters/gokafka/v1.0.0/adapters/gokafka/docs/reference.md))

- [github.com/faustbrian/go-transactional-outbox/adapters/gorabbitstream](https://pkg.go.dev/github.com/faustbrian/go-transactional-outbox/adapters/gorabbitstream@v1.0.2) — **deprecated**, `v1.0.2`: Preserve the released RabbitMQ Streams adapter contract while consumers migrate to the target-oriented adapters/rabbitstream module. ([README](https://github.com/faustbrian/go-transactional-outbox/blob/adapters/gorabbitstream/v1.0.2/adapters/gorabbitstream/README.md)) ([migration](https://github.com/faustbrian/go-transactional-outbox/blob/adapters/gorabbitstream/v1.0.2/adapters/gorabbitstream/README.md))

- [github.com/faustbrian/go-transactional-outbox/adapters/kafka](https://pkg.go.dev/github.com/faustbrian/go-transactional-outbox/adapters/kafka@v1.0.0) — **active**, `v1.0.0`: Map one durable outbox envelope to one confirmed first-party Kafka record without acquiring or owning the producer. ([README](https://github.com/faustbrian/go-transactional-outbox/blob/adapters/kafka/v1.0.0/adapters/kafka/README.md))

- [github.com/faustbrian/go-transactional-outbox/adapters/otel](https://pkg.go.dev/github.com/faustbrian/go-transactional-outbox/adapters/otel@v1.0.0) — **active**, `v1.0.0`: Add bounded outbox semantic spans, metrics, propagation, observations, and publisher instrumentation without owning telemetry infrastructure. ([README](https://github.com/faustbrian/go-transactional-outbox/blob/adapters/otel/v1.0.0/adapters/otel/README.md))

- [github.com/faustbrian/go-transactional-outbox/adapters/queue](https://pkg.go.dev/github.com/faustbrian/go-transactional-outbox/adapters/queue@v1.0.0) — **active**, `v1.0.0`: Map one durable outbox envelope to one bounded deterministic first-party queue task while preserving acceptance ambiguity. ([README](https://github.com/faustbrian/go-transactional-outbox/blob/adapters/queue/v1.0.0/adapters/queue/README.md))

- [github.com/faustbrian/go-transactional-outbox/adapters/rabbitstream](https://pkg.go.dev/github.com/faustbrian/go-transactional-outbox/adapters/rabbitstream@v1.0.1) — **active**, `v1.0.1`: Map one durable outbox envelope to one confirmed RabbitMQ Stream or Super Stream message without owning transport lifecycle. ([README](https://github.com/faustbrian/go-transactional-outbox/blob/adapters/rabbitstream/v1.0.1/adapters/rabbitstream/README.md))

- [github.com/faustbrian/go-workflow](https://pkg.go.dev/github.com/faustbrian/go-workflow@v1.0.0) — **active**, `v1.0.0`: Provide immutable workflow definitions and history, deterministic orchestration decisions, bounded durable work processing, explicit recovery semantics, and a PostgreSQL persistence adapter. ([README](https://github.com/faustbrian/go-workflow/blob/v1.0.0/README.md))

## resilience

- [github.com/faustbrian/go-adaptive-throttle](https://pkg.go.dev/github.com/faustbrian/go-adaptive-throttle@v1.0.0) — **active**, `v1.0.0`: Provide process-local rolling overload history and probabilistic admission that sheds a bounded share of work while preserving probe flow. ([README](https://github.com/faustbrian/go-adaptive-throttle/blob/v1.0.0/README.md))

- [github.com/faustbrian/go-bulkhead](https://pkg.go.dev/github.com/faustbrian/go-bulkhead@v1.0.0) — **active**, `v1.0.0`: Provide process-local fixed-capacity resource isolation with weighted permits, bounded FIFO waiting, explicit partitions, and graceful drain. ([README](https://github.com/faustbrian/go-bulkhead/blob/v1.0.0/README.md))

- [github.com/faustbrian/go-circuit-breaker](https://pkg.go.dev/github.com/faustbrian/go-circuit-breaker@v1.0.0) — **active**, `v1.0.0`: Provide protocol-neutral, bounded circuit-breaker state, dependency-health admission, rolling outcome windows, and explicit permit and observation lifecycles. ([README](https://github.com/faustbrian/go-circuit-breaker/blob/v1.0.0/README.md))

- [github.com/faustbrian/go-concurrency-limit](https://pkg.go.dev/github.com/faustbrian/go-concurrency-limit@v1.0.0) — **active**, `v1.0.0`: Provide bounded, process-local adaptive in-flight admission that learns a safe concurrency limit from explicit execution outcomes. ([README](https://github.com/faustbrian/go-concurrency-limit/blob/v1.0.0/README.md))

- [github.com/faustbrian/go-fault-injection](https://pkg.go.dev/github.com/faustbrian/go-fault-injection@v1.0.0) — **active**, `v1.0.0`: Provide deterministic, bounded fault schedules, failure wrappers, and a fail-closed runtime for tests and explicitly authorized controlled experiments. ([README](https://github.com/faustbrian/go-fault-injection/blob/v1.0.0/README.md))

- [github.com/faustbrian/go-hedge](https://pkg.go.dev/github.com/faustbrian/go-hedge@v1.0.0) — **active**, `v1.0.0`: Provide finite delayed duplicate attempts for explicitly replay-safe work under caller-owned deadlines and bounded amplification. ([README](https://github.com/faustbrian/go-hedge/blob/v1.0.0/README.md))

- [github.com/faustbrian/go-rate-limit](https://pkg.go.dev/github.com/faustbrian/go-rate-limit@v1.1.0) — **active**, `v1.1.0`: Provide transport-neutral inbound admission policies, bounded memory and distributed backends, concurrency leases, and explicit HTTP, RPC, queue, principal, logging, and telemetry integrations. ([README](https://github.com/faustbrian/go-rate-limit/blob/v1.1.0/README.md))

- [github.com/faustbrian/go-resilience](https://pkg.go.dev/github.com/faustbrian/go-resilience@v1.0.0) — **active**, `v1.0.0`: Provide deterministic generic policy composition, typed outcomes, caller-owned total deadlines, bounded observation, and shared process-local retry and hedge work budgets. ([README](https://github.com/faustbrian/go-resilience/blob/v1.0.0/README.md))

- [github.com/faustbrian/go-retry](https://pkg.go.dev/github.com/faustbrian/go-retry@v1.1.0) — **active**, `v1.1.0`: Provide bounded retry execution, explicit failure classification, finite time and work budgets, deterministic backoff strategies, and focused transport and observability adapters. ([README](https://github.com/faustbrian/go-retry/blob/v1.1.0/README.md))

- [github.com/faustbrian/go-semaphore](https://pkg.go.dev/github.com/faustbrian/go-semaphore@v1.0.0) — **active**, `v1.0.0`: Provide a process-local FIFO weighted semaphore with bounded waiting, exactly-once permits, deterministic shutdown, and bounded observation events. ([README](https://github.com/faustbrian/go-semaphore/blob/v1.0.0/README.md))

## observability

- [github.com/faustbrian/go-log](https://pkg.go.dev/github.com/faustbrian/go-log@v1.0.0) — **active**, `v1.0.0`: Provide composable log/slog handlers for routing, redaction, sampling, bounded asynchronous delivery, capture, local rotation, and OpenTelemetry correlation. ([README](https://github.com/faustbrian/go-log/blob/v1.0.0/README.md))

- [github.com/faustbrian/go-telemetry](https://pkg.go.dev/github.com/faustbrian/go-telemetry@v1.2.0) — **active**, `v1.2.0`: Provide a vendor-neutral OpenTelemetry runtime, OTLP exporters, propagation, sampling, bounded instrumentation, and explicit provider lifecycle for Go services. ([README](https://github.com/faustbrian/go-telemetry/blob/v1.2.0/README.md))

## integration-and-data-movement

- [github.com/faustbrian/go-external-sort](https://pkg.go.dev/github.com/faustbrian/go-external-sort@v1.0.0) — **active**, `v1.0.0`: Provide bounded external sorting of fixed-width opaque records using authenticated encrypted temporary storage. ([README](https://github.com/faustbrian/go-external-sort/blob/v1.0.0/README.md))

- [github.com/faustbrian/go-filesystem](https://pkg.go.dev/github.com/faustbrian/go-filesystem@v1.1.0) — **active**, `v1.1.0`: Provide capability-oriented streaming filesystem contracts, backend-specific adapters, composable decorators, and conformance helpers. ([README](https://github.com/faustbrian/go-filesystem/blob/v1.1.0/README.md))

- [github.com/faustbrian/go-http-client](https://pkg.go.dev/github.com/faustbrian/go-http-client@v1.1.0) — **active**, `v1.1.0`: Provide typed outbound HTTP policy with finite transport defaults, immutable request specifications, deterministic middleware, and explicit response ownership. ([README](https://github.com/faustbrian/go-http-client/blob/v1.1.0/README.md))

- [github.com/faustbrian/go-kafka](https://pkg.go.dev/github.com/faustbrian/go-kafka@v1.1.0) — **active**, `v1.1.0`: Provide bounded first-party Apache Kafka producer, consumer, inspection, replay, and transaction policy over franz-go. ([README](https://github.com/faustbrian/go-kafka/blob/v1.1.0/README.md))

- [github.com/faustbrian/go-kafka/adapters/gotelemetry](https://pkg.go.dev/github.com/faustbrian/go-kafka/adapters/gotelemetry@v1.0.1) — **deprecated**, `v1.0.1`: Preserve the released OpenTelemetry adapter path and behavior while delegating to the canonical adapters/otel successor. ([README](https://github.com/faustbrian/go-kafka/blob/adapters/gotelemetry/v1.0.1/adapters/gotelemetry/README.md)) ([migration](https://github.com/faustbrian/go-kafka/blob/adapters/gotelemetry/v1.0.1/adapters/gotelemetry/docs/README.md))

- [github.com/faustbrian/go-kafka/adapters/mskiam](https://pkg.go.dev/github.com/faustbrian/go-kafka/adapters/mskiam@v1.1.0) — **active**, `v1.1.0`: Provide Kafka SASL/OAUTHBEARER authentication configuration through AWS MSK IAM credentials and signing. ([README](https://github.com/faustbrian/go-kafka/blob/adapters/mskiam/v1.1.0/adapters/mskiam/README.md))

- [github.com/faustbrian/go-kafka/adapters/otel](https://pkg.go.dev/github.com/faustbrian/go-kafka/adapters/otel@v1.0.0) — **active**, `v1.0.0`: Translate Kafka observation and propagation contracts into OpenTelemetry spans and metrics. ([README](https://github.com/faustbrian/go-kafka/blob/adapters/otel/v1.0.0/adapters/otel/README.md))

- [github.com/faustbrian/go-kafka/adapters/service](https://pkg.go.dev/github.com/faustbrian/go-kafka/adapters/service@v1.0.0) — **active**, `v1.0.0`: Bridge go-kafka consumer and producer resources into the explicit go-service lifecycle. ([README](https://github.com/faustbrian/go-kafka/blob/adapters/service/v1.0.0/adapters/service/README.md))

- [github.com/faustbrian/go-kafka/kafkaservice](https://pkg.go.dev/github.com/faustbrian/go-kafka/kafkaservice@v1.0.1) — **deprecated**, `v1.0.1`: Preserve the released service-integration path and behavior while delegating to the canonical adapters/service successor. ([README](https://github.com/faustbrian/go-kafka/blob/kafkaservice/v1.0.1/kafkaservice/README.md)) ([migration](https://github.com/faustbrian/go-kafka/blob/kafkaservice/v1.0.1/kafkaservice/docs/README.md))

- [github.com/faustbrian/go-rabbitmq-queues](https://pkg.go.dev/github.com/faustbrian/go-rabbitmq-queues@v1.1.0) — **active**, `v1.1.0`: Provide bounded RabbitMQ-native AMQP 0-9-1 classic and quorum queue publishing, consumption, settlement, recovery, topology verification, health, and observation policy. ([README](https://github.com/faustbrian/go-rabbitmq-queues/blob/v1.1.0/README.md))

- [github.com/faustbrian/go-rabbitmq-streams](https://pkg.go.dev/github.com/faustbrian/go-rabbitmq-streams@v1.1.1) — **active**, `v1.1.1`: Provide vendor-neutral bounded policy for RabbitMQ Streams messages, publishing, consumption, replay, inspection, failures, lifecycle, and observations. ([README](https://github.com/faustbrian/go-rabbitmq-streams/blob/v1.1.1/README.md))

- [github.com/faustbrian/go-rabbitmq-streams/adapters/otel](https://pkg.go.dev/github.com/faustbrian/go-rabbitmq-streams/adapters/otel@v1.0.0) — **active**, `v1.0.0`: Translate bounded RabbitMQ Streams observations into caller-owned OpenTelemetry metrics and propagate W3C Trace Context through message headers. ([README](https://github.com/faustbrian/go-rabbitmq-streams/blob/adapters/otel/v1.0.0/adapters/otel/README.md))

- [github.com/faustbrian/go-rabbitmq-streams/adapters/rabbitmq](https://pkg.go.dev/github.com/faustbrian/go-rabbitmq-streams/adapters/rabbitmq@v1.0.1) — **active**, `v1.0.1`: Adapt the RabbitMQ-supported Go Streams client to bounded rabbitstream policy while owning protocol connections, sessions, cursors, recovery, and wire conversion. ([README](https://github.com/faustbrian/go-rabbitmq-streams/blob/adapters/rabbitmq/v1.0.1/adapters/rabbitmq/README.md))

- [github.com/faustbrian/go-rabbitmq-streams/otel](https://pkg.go.dev/github.com/faustbrian/go-rabbitmq-streams/otel@v1.0.1) — **deprecated**, `v1.0.1`: Preserve the released OpenTelemetry adapter import path, public type identities, instrumentation scope, errors, and behavior while delegating to adapters/otel. ([README](https://github.com/faustbrian/go-rabbitmq-streams/blob/otel/v1.0.1/otel/README.md)) ([migration](https://github.com/faustbrian/go-rabbitmq-streams/blob/otel/v1.0.1/otel/README.md))

- [github.com/faustbrian/go-rabbitmq-streams/rabbitmq](https://pkg.go.dev/github.com/faustbrian/go-rabbitmq-streams/rabbitmq@v1.0.2) — **deprecated**, `v1.0.2`: Preserve the released RabbitMQ Streams adapter import path, public type identities, errors, and behavior while delegating to adapters/rabbitmq. ([README](https://github.com/faustbrian/go-rabbitmq-streams/blob/rabbitmq/v1.0.2/rabbitmq/README.md)) ([migration](https://github.com/faustbrian/go-rabbitmq-streams/blob/rabbitmq/v1.0.2/rabbitmq/README.md))

- [github.com/faustbrian/go-search](https://pkg.go.dev/github.com/faustbrian/go-search@v1.0.0) — **active**, `v1.0.0`: Provide backend-neutral contracts for bounded document indexing, typed querying, cursor pagination, schema migration, projections, and reconciliation while treating application data as authoritative. ([README](https://github.com/faustbrian/go-search/blob/v1.0.0/README.md))

- [github.com/faustbrian/go-search/adapters/opensearch](https://pkg.go.dev/github.com/faustbrian/go-search/adapters/opensearch@v1.0.0) — **active**, `v1.0.0`: Translate the backend-neutral search contract to a bounded OpenSearch client with explicit transport, trust, lifecycle, resilience, and observability policy. ([README](https://github.com/faustbrian/go-search/blob/adapters/opensearch/v1.0.0/adapters/opensearch/README.md))

- [github.com/faustbrian/go-secret-envelope](https://pkg.go.dev/github.com/faustbrian/go-secret-envelope@v1.0.0) — **active**, `v1.0.0`: Provide bounded authenticated secret envelopes with explicit key-provider boundaries, immutable encryption context, and versioned persistence bytes. ([README](https://github.com/faustbrian/go-secret-envelope/blob/v1.0.0/README.md))

- [github.com/faustbrian/go-tabular](https://pkg.go.dev/github.com/faustbrian/go-tabular@v1.0.0) — **active**, `v1.0.0`: Provide explicit, bounded ingestion for delimited, fixed-width, XLS, XLSX, and ZIP-backed tabular sources. ([README](https://github.com/faustbrian/go-tabular/blob/v1.0.0/README.md))

- [github.com/faustbrian/go-wire](https://pkg.go.dev/github.com/faustbrian/go-wire@v1.0.0) — **active**, `v1.0.0`: Provide explicit, bounded JSON, XML, SOAP, YAML, TOML, MessagePack, CBOR, and BSON encoding and decoding boundaries. ([README](https://github.com/faustbrian/go-wire/blob/v1.0.0/README.md))

## domain-utilities

- [github.com/faustbrian/go-barcode](https://pkg.go.dev/github.com/faustbrian/go-barcode@v1.0.0) — **active**, `v1.0.0`: Provide immutable logical barcode symbols, strict validation, encoding, decoding, rendering, and standards-conformance evidence. ([README](https://github.com/faustbrian/go-barcode/blob/v1.0.0/README.md))

- [github.com/faustbrian/go-calendar](https://pkg.go.dev/github.com/faustbrian/go-calendar@v1.1.0) — **active**, `v1.1.0`: Provide immutable civil dates, Gregorian arithmetic, typed periods, explicit DST conversion, and bounded business calendars. ([README](https://github.com/faustbrian/go-calendar/blob/v1.1.0/README.md))

- [github.com/faustbrian/go-ecma-regexp](https://pkg.go.dev/github.com/faustbrian/go-ecma-regexp@v1.0.0) — **active**, `v1.0.0`: Provide bounded ECMAScript regular-expression parsing, compilation, and matching, including the JSON Schema pattern profile. ([README](https://github.com/faustbrian/go-ecma-regexp/blob/v1.0.0/README.md))

- [github.com/faustbrian/go-geo](https://pkg.go.dev/github.com/faustbrian/go-geo@v1.1.0) — **active**, `v1.1.0`: Provide immutable geospatial values, bounded geometry and geodesy operations, interoperable codecs, PostGIS mapping, and deterministic test helpers. ([README](https://github.com/faustbrian/go-geo/blob/v1.1.0/README.md))

- [github.com/faustbrian/go-keyphrase](https://pkg.go.dev/github.com/faustbrian/go-keyphrase@v1.0.0) — **active**, `v1.0.0`: Provide unbiased bounded password and passphrase generation, BIP-39 mnemonic interoperability, immutable word lists, and explicit secret-handling boundaries. ([README](https://github.com/faustbrian/go-keyphrase/blob/v1.0.0/README.md))

- [github.com/faustbrian/go-knapsack](https://pkg.go.dev/github.com/faustbrian/go-knapsack@v1.0.0) — **active**, `v1.0.0`: Provide deterministic, bounded offline orthogonal packing, exact objectives, extension constraints, canonical plans, and independent verification. ([README](https://github.com/faustbrian/go-knapsack/blob/v1.0.0/README.md))

- [github.com/faustbrian/go-knapsack/objective/gomoney](https://pkg.go.dev/github.com/faustbrian/go-knapsack/objective/gomoney@v1.1.0) — **deprecated**, `v1.1.0`: Preserve the released exact-money objective API while delegating behavior to the canonical target-oriented module. ([README](https://github.com/faustbrian/go-knapsack/blob/objective/gomoney/v1.1.0/objective/gomoney/README.md)) ([migration](https://github.com/faustbrian/go-knapsack/blob/objective/gomoney/v1.1.0/objective/gomoney/README.md))

- [github.com/faustbrian/go-knapsack/objective/money](https://pkg.go.dev/github.com/faustbrian/go-knapsack/objective/money@v1.0.0) — **active**, `v1.0.0`: Adapt exact go-money values into deterministic Knapsack container-cost objective evaluation through the canonical target-oriented path. ([README](https://github.com/faustbrian/go-knapsack/blob/objective/money/v1.0.0/objective/money/README.md))

- [github.com/faustbrian/go-math](https://pkg.go.dev/github.com/faustbrian/go-math@v1.1.0) — **active**, `v1.1.0`: Provide immutable arbitrary-precision integer, rational, decimal, and binary-float values with explicit precision, rounding, limits, conditions, and deterministic encodings. ([README](https://github.com/faustbrian/go-math/blob/v1.1.0/README.md))

- [github.com/faustbrian/go-measurement](https://pkg.go.dev/github.com/faustbrian/go-measurement@v1.1.0) — **active**, `v1.1.0`: Provide immutable, exact, unit-safe quantities, dimensions, conversions, logistics formulas, and bounded wire encodings. ([README](https://github.com/faustbrian/go-measurement/blob/v1.1.0/README.md))

- [github.com/faustbrian/go-merkle-patricia-trie](https://pkg.go.dev/github.com/faustbrian/go-merkle-patricia-trie@v1.0.0) — **active**, `v1.0.0`: Provide bounded immutable Ethereum modified Merkle Patricia tries, roots, proofs, storage integration, retention, pruning, and recovery. ([README](https://github.com/faustbrian/go-merkle-patricia-trie/blob/v1.0.0/README.md))

- [github.com/faustbrian/go-merkle-tree](https://pkg.go.dev/github.com/faustbrian/go-merkle-tree@v1.0.0) — **active**, `v1.0.0`: Provide storage-independent ordered Merkle trees, canonical and RFC 9162 profiles, immutable snapshots, and bounded inclusion, multi-inclusion, and consistency proofs. ([README](https://github.com/faustbrian/go-merkle-tree/blob/v1.0.0/README.md))

- [github.com/faustbrian/go-money](https://pkg.go.dev/github.com/faustbrian/go-money@v1.0.0) — **active**, `v1.0.0`: Provide immutable exact monetary values, explicit precision and rounding contexts, bounded arithmetic, allocation, tax, discount, conversion, formatting, and persistence encodings. ([README](https://github.com/faustbrian/go-money/blob/v1.0.0/README.md))

- [github.com/faustbrian/go-opening-hours](https://pkg.go.dev/github.com/faustbrian/go-opening-hours@v1.1.0) — **active**, `v1.1.0`: Model immutable recurring opening hours, dated exceptions, timezone-aware availability, and bounded schedule composition. ([README](https://github.com/faustbrian/go-opening-hours/blob/v1.1.0/README.md))

- [github.com/faustbrian/go-rule-engine](https://pkg.go.dev/github.com/faustbrian/go-rule-engine@v1.0.0) — **active**, `v1.0.0`: Provide deterministic typed rule construction, bounded compilation into immutable plans, explicit fact evaluation, canonical JSON, and redacted diagnostics without hidden I/O. ([README](https://github.com/faustbrian/go-rule-engine/blob/v1.0.0/README.md))

- [github.com/faustbrian/go-rule-engine/adapters/math](https://pkg.go.dev/github.com/faustbrian/go-rule-engine/adapters/math@v1.0.0) — **active**, `v1.0.0`: Bridge immutable go-math decimals into rule-engine tagged values and deterministic equality and ordering operators without coupling the core module to decimal arithmetic. ([README](https://github.com/faustbrian/go-rule-engine/blob/adapters/math/v1.0.0/adapters/math/README.md))

- [github.com/faustbrian/go-rule-engine/adapters/measurement](https://pkg.go.dev/github.com/faustbrian/go-rule-engine/adapters/measurement@v1.0.0) — **active**, `v1.0.0`: Bridge immutable go-measurement quantities into tagged rule-engine values and dimension-safe deterministic comparison operators. ([README](https://github.com/faustbrian/go-rule-engine/blob/adapters/measurement/v1.0.0/adapters/measurement/README.md))

- [github.com/faustbrian/go-rule-engine/adapters/temporal](https://pkg.go.dev/github.com/faustbrian/go-rule-engine/adapters/temporal@v1.0.0) — **active**, `v1.0.0`: Bridge exact go-temporal instants and periods into tagged rule-engine values and deterministic relation operators. ([README](https://github.com/faustbrian/go-rule-engine/blob/adapters/temporal/v1.0.0/adapters/temporal/README.md))

- [github.com/faustbrian/go-state-machine](https://pkg.go.dev/github.com/faustbrian/go-state-machine@v1.0.0) — **active**, `v1.0.0`: Provide deterministic typed state-machine compilation, transition selection, inert effect planning, replay and evolution, with optional explicit execution, persistence, outbox delivery, and diagram rendering. ([README](https://github.com/faustbrian/go-state-machine/blob/v1.0.0/README.md))

- [github.com/faustbrian/go-temporal](https://pkg.go.dev/github.com/faustbrian/go-temporal@v1.1.0) — **active**, `v1.1.0`: Provide immutable temporal algebra, explicit interval bounds and relations, normalized sets, fixed durations, daily intervals, strict notation, and versioned encodings. ([README](https://github.com/faustbrian/go-temporal/blob/v1.1.0/README.md))

- [github.com/faustbrian/go-verkle-tree](https://pkg.go.dev/github.com/faustbrian/go-verkle-tree@v1.0.0) — **active**, `v1.0.0`: Provide bounded immutable authenticated key/value trees, roots, proofs, witnesses, stateless updates, and caller-owned storage protocols for the package-owned Bandersnatch IPA profile. ([README](https://github.com/faustbrian/go-verkle-tree/blob/v1.0.0/README.md))

## tooling

- [github.com/faustbrian/go-analysis](https://pkg.go.dev/github.com/faustbrian/go-analysis@v1.1.0) — **active**, `v1.1.0`: Provide deterministic go/analysis policies, governed rule metadata, and bounded JSON and SARIF reports for Go repositories. ([README](https://github.com/faustbrian/go-analysis/blob/v1.1.0/README.md))

- [github.com/faustbrian/go-cli](https://pkg.go.dev/github.com/faustbrian/go-cli@v1.0.0) — **active**, `v1.0.0`: Provide explicit typed command construction, immutable compilation, deterministic parsing, lifecycle middleware, bounded output, and stable process-facing results. ([README](https://github.com/faustbrian/go-cli/blob/v1.0.0/README.md))

- [github.com/faustbrian/go-prompts](https://pkg.go.dev/github.com/faustbrian/go-prompts@v1.1.0) — **active**, `v1.1.0`: Provide typed interactive prompts, deterministic non-interactive parsing, semantic rendering, caller-driven presentation, and an explicit target-oriented terminal adapter. ([README](https://github.com/faustbrian/go-prompts/blob/v1.1.0/README.md))
