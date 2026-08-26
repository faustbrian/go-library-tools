# Service Fixtures

The gate runner owns generic PostgreSQL, Valkey, Redis, NATS, NSQ, RabbitMQ,
RabbitMQ Streams, and OpenSearch fixtures used by current libraries.

Every lease uses unique task-owned containers, networks, volumes, ports,
credentials, and files. Readiness is bounded and deterministic. Environment is
exposed only to the selected module gate, and cleanup removes only exact owned
resources on success or failure.

RabbitMQ Streams supports standalone, three-node cluster, rolling-upgrade,
authorization, Toxiproxy, and mutual-TLS scenarios. OpenSearch images come from
a strict module-owned digest lock. Package-specific schemas and payloads remain
in the consumer repository.
