# Service Fixtures

<!-- cspell:words RABBITSTREAM -->

The gate runner owns generic PostgreSQL, Valkey, Redis, NATS, NSQ, RabbitMQ,
RabbitMQ Streams, and OpenSearch fixtures used by current libraries.

Every lease uses unique task-owned containers, networks, volumes, ports,
credentials, and files. Readiness is bounded and deterministic. Environment is
exposed only to the selected module gate, and cleanup removes only exact owned
resources on success or failure.

Fixtures are intentionally not exposed as detached `services start` and
`services stop` commands. `golib check` starts the selected module's declared
fixtures in-process, passes their environment only to that gate, and closes the
lease on success, failure, or process cancellation. This prevents stale state
files and orphaned manual leases from becoming another lifecycle protocol.

RabbitMQ Streams supports standalone, three-node cluster, rolling-upgrade,
authorization, Toxiproxy, and mutual-TLS scenarios. OpenSearch images come from
a strict module-owned digest lock. Package-specific schemas and payloads remain
in the consumer repository.

RabbitMQ Streams gates that recreate cluster members receive the generated
Compose path and project through `RABBITSTREAM_CLUSTER_COMPOSE` and
`RABBITSTREAM_CLUSTER_PROJECT`. The same gate environment exposes the exact
task-owned `RABBITSTREAM_NETWORK`, `RABBITSTREAM_VOLUME_RABBIT1`,
`RABBITSTREAM_VOLUME_RABBIT2`, `RABBITSTREAM_VOLUME_RABBIT3`,
`RABBITSTREAM_VOLUME_TLS_CERTS`, and `RABBITSTREAM_VOLUME_TLS_DATA` identities.
Consumers must pass those values back to the generated Compose file rather than
constructing names or using shared resources.
