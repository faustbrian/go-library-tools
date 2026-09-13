module github.com/faustbrian/go-library-tools/release/compatibility-consumer

go 1.27.0

require (
	github.com/faustbrian/go-adaptive-throttle v1.0.0
	github.com/faustbrian/go-analysis v1.1.0
	github.com/faustbrian/go-api-query v1.1.0
	github.com/faustbrian/go-audit v1.0.0
	github.com/faustbrian/go-audit/postgres v1.0.0
	github.com/faustbrian/go-authentication v1.2.0
	github.com/faustbrian/go-authentication/adapters/otel v1.0.0
	github.com/faustbrian/go-authentication/jwt v1.1.0
	github.com/faustbrian/go-authentication/oidc v1.0.0
	github.com/faustbrian/go-authorization v1.1.0
	github.com/faustbrian/go-barcode v1.0.0
	github.com/faustbrian/go-bulkhead v1.0.0
	github.com/faustbrian/go-cache v1.0.0
	github.com/faustbrian/go-calendar v1.1.0
	github.com/faustbrian/go-capability v1.1.0
	github.com/faustbrian/go-circuit-breaker v1.0.0
	github.com/faustbrian/go-cli v1.0.0
	github.com/faustbrian/go-clock v1.1.0
	github.com/faustbrian/go-cloudevents v1.1.0
	github.com/faustbrian/go-cloudevents/adapters/audit v1.0.0
	github.com/faustbrian/go-cloudevents/adapters/correlation v1.0.0
	github.com/faustbrian/go-cloudevents/adapters/event-sourcing v1.0.0
	github.com/faustbrian/go-cloudevents/adapters/jsonschema v1.0.0
	github.com/faustbrian/go-cloudevents/adapters/kafka v1.0.0
	github.com/faustbrian/go-cloudevents/adapters/outbox v1.0.0
	github.com/faustbrian/go-cloudevents/adapters/queue v1.0.0
	github.com/faustbrian/go-cloudevents/adapters/rabbitstream v1.0.0
	github.com/faustbrian/go-cloudevents/adapters/schema-registry v1.0.0
	github.com/faustbrian/go-cloudevents/adapters/telemetry v1.0.0
	github.com/faustbrian/go-cloudevents/adapters/tenancy v1.0.0
	github.com/faustbrian/go-cloudevents/adapters/workflow v1.0.0
	github.com/faustbrian/go-concurrency-limit v1.0.0
	github.com/faustbrian/go-config v1.1.0
	github.com/faustbrian/go-config/adapters/awssecretsmanager v1.0.0
	github.com/faustbrian/go-correlation v1.1.0
	github.com/faustbrian/go-ecma-regexp v1.0.0
	github.com/faustbrian/go-event-sourcing v1.0.0
	github.com/faustbrian/go-event-sourcing/adapters/kafka v1.0.2
	github.com/faustbrian/go-event-sourcing/adapters/otel v1.0.2
	github.com/faustbrian/go-event-sourcing/adapters/outbox v1.0.0
	github.com/faustbrian/go-event-sourcing/adapters/queue v1.0.0
	github.com/faustbrian/go-event-sourcing/postgres v1.0.0
	github.com/faustbrian/go-external-sort v1.0.0
	github.com/faustbrian/go-fault-injection v1.0.0
	github.com/faustbrian/go-feature-flags v1.0.0
	github.com/faustbrian/go-filesystem v1.1.0
	github.com/faustbrian/go-geo v1.1.0
	github.com/faustbrian/go-hedge v1.0.0
	github.com/faustbrian/go-http-client v1.1.0
	github.com/faustbrian/go-http-middleware v1.0.0
	github.com/faustbrian/go-http-signature v1.0.0
	github.com/faustbrian/go-idempotency v1.1.0
	github.com/faustbrian/go-identifier v1.0.0
	github.com/faustbrian/go-international v1.1.0
	github.com/faustbrian/go-json-schema v1.0.0
	github.com/faustbrian/go-jsonapi v1.0.0
	github.com/faustbrian/go-jsonrpc v1.0.0
	github.com/faustbrian/go-kafka v1.1.0
	github.com/faustbrian/go-kafka/adapters/mskiam v1.1.0
	github.com/faustbrian/go-kafka/adapters/otel v1.0.0
	github.com/faustbrian/go-kafka/adapters/service v1.0.0
	github.com/faustbrian/go-keyphrase v1.0.0
	github.com/faustbrian/go-knapsack v1.0.0
	github.com/faustbrian/go-knapsack/objective/money v1.0.0
	github.com/faustbrian/go-lease v1.1.0
	github.com/faustbrian/go-localized v1.1.0
	github.com/faustbrian/go-log v1.0.0
	github.com/faustbrian/go-math v1.1.0
	github.com/faustbrian/go-measurement v1.1.0
	github.com/faustbrian/go-merkle-patricia-trie v1.0.0
	github.com/faustbrian/go-merkle-tree v1.0.0
	github.com/faustbrian/go-migrations v1.1.0
	github.com/faustbrian/go-money v1.0.0
	github.com/faustbrian/go-openapi v1.0.0
	github.com/faustbrian/go-opening-hours v1.1.0
	github.com/faustbrian/go-openrpc v1.0.0
	github.com/faustbrian/go-password v1.1.0
	github.com/faustbrian/go-postgres v1.1.0
	github.com/faustbrian/go-prompts v1.1.0
	github.com/faustbrian/go-queue v1.1.1
	github.com/faustbrian/go-queue-control-plane v1.1.0
	github.com/faustbrian/go-queue/adapters/rabbitmq v1.0.0
	github.com/faustbrian/go-queue/adapters/service v1.0.1
	github.com/faustbrian/go-rabbitmq-queues v1.1.0
	github.com/faustbrian/go-rabbitmq-streams v1.1.1
	github.com/faustbrian/go-rabbitmq-streams/adapters/otel v1.0.0
	github.com/faustbrian/go-rabbitmq-streams/adapters/rabbitmq v1.0.1
	github.com/faustbrian/go-rate-limit v1.1.0
	github.com/faustbrian/go-resilience v1.0.0
	github.com/faustbrian/go-retry v1.1.0
	github.com/faustbrian/go-router v1.0.0
	github.com/faustbrian/go-rule-engine v1.0.0
	github.com/faustbrian/go-rule-engine/adapters/math v1.0.0
	github.com/faustbrian/go-rule-engine/adapters/measurement v1.0.0
	github.com/faustbrian/go-rule-engine/adapters/temporal v1.0.0
	github.com/faustbrian/go-scheduler v1.1.0
	github.com/faustbrian/go-schema-registry v1.0.0
	github.com/faustbrian/go-schema-registry/providers/confluent v1.0.0
	github.com/faustbrian/go-schema-registry/providers/glue v1.0.0
	github.com/faustbrian/go-search v1.0.0
	github.com/faustbrian/go-search/adapters/opensearch v1.0.0
	github.com/faustbrian/go-secret-envelope v1.0.0
	github.com/faustbrian/go-semaphore v1.0.0
	github.com/faustbrian/go-sequencer v1.1.0
	github.com/faustbrian/go-service v1.1.0
	github.com/faustbrian/go-settings v1.1.0
	github.com/faustbrian/go-state-machine v1.0.0
	github.com/faustbrian/go-tabular v1.0.0
	github.com/faustbrian/go-telemetry v1.2.0
	github.com/faustbrian/go-temporal v1.1.0
	github.com/faustbrian/go-tenancy v1.1.0
	github.com/faustbrian/go-transactional-outbox v1.0.0
	github.com/faustbrian/go-transactional-outbox/adapters/kafka v1.0.0
	github.com/faustbrian/go-transactional-outbox/adapters/otel v1.0.0
	github.com/faustbrian/go-transactional-outbox/adapters/queue v1.0.0
	github.com/faustbrian/go-transactional-outbox/adapters/rabbitstream v1.0.1
	github.com/faustbrian/go-validation v1.1.0
	github.com/faustbrian/go-verkle-tree v1.0.0
	github.com/faustbrian/go-webhook v1.0.0
	github.com/faustbrian/go-wire v1.0.0
	github.com/faustbrian/go-workflow v1.0.0
	github.com/faustbrian/go-wsdl v1.0.0
	github.com/faustbrian/go-xsd v1.0.0
)

require (
	github.com/aws/aws-msk-iam-sasl-signer-go v1.0.4 // indirect
	github.com/aws/aws-sdk-go-v2 v1.43.4 // indirect
	github.com/aws/aws-sdk-go-v2/config v1.32.35 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.19.34 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.18.35 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.4.35 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.7.35 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.4.36 // indirect
	github.com/aws/aws-sdk-go-v2/service/glue v1.152.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.15 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.13.35 // indirect
	github.com/aws/aws-sdk-go-v2/service/secretsmanager v1.44.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/signin v1.5.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.33.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.38.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.45.4 // indirect
	github.com/aws/smithy-go v1.27.7 // indirect
	github.com/bits-and-blooms/bitset v1.24.4 // indirect
	github.com/cenkalti/backoff/v5 v5.0.3 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/consensys/gnark-crypto v0.20.1 // indirect
	github.com/coreos/go-oidc/v3 v3.20.0 // indirect
	github.com/crate-crypto/go-ipa v0.0.0-20240223125850-b1e8a79f509c // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.4.1 // indirect
	github.com/deszhou/jcs v1.0.0 // indirect
	github.com/dlclark/regexp2 v1.12.0 // indirect
	github.com/dlclark/regexp2/v2 v2.5.1 // indirect
	github.com/dunglas/httpsfv v1.1.0 // indirect
	github.com/ericlevine/zxinggo v0.1.0 // indirect
	github.com/go-jose/go-jose/v4 v4.1.4 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/goccy/go-json v0.10.6 // indirect
	github.com/golang/snappy v1.0.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.29.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.10.0 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/jpillora/backoff v1.0.0 // indirect
	github.com/klauspost/compress v1.19.1 // indirect
	github.com/lestrrat-go/blackmagic v1.0.4 // indirect
	github.com/lestrrat-go/dsig v1.2.1 // indirect
	github.com/lestrrat-go/dsig-secp256k1 v1.0.0 // indirect
	github.com/lestrrat-go/httpcc v1.0.1 // indirect
	github.com/lestrrat-go/httprc/v3 v3.0.5 // indirect
	github.com/lestrrat-go/jwx/v3 v3.1.1 // indirect
	github.com/lestrrat-go/option/v2 v2.0.0 // indirect
	github.com/makiuchi-d/gozxing v0.1.1 // indirect
	github.com/oklog/ulid/v2 v2.1.1 // indirect
	github.com/opensearch-project/opensearch-go/v4 v4.7.3 // indirect
	github.com/peterstace/simplefeatures v0.59.0 // indirect
	github.com/pierrec/lz4 v2.6.1+incompatible // indirect
	github.com/pierrec/lz4/v4 v4.1.26 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/rabbitmq/amqp091-go v1.14.0 // indirect
	github.com/rabbitmq/rabbitmq-stream-go-client v1.8.3 // indirect
	github.com/richardlehane/mscfb v1.0.7 // indirect
	github.com/richardlehane/msoleps v1.0.6 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	github.com/santhosh-tekuri/jsonschema/v6 v6.0.2 // indirect
	github.com/segmentio/asm v1.2.1 // indirect
	github.com/spaolacci/murmur3 v1.1.0 // indirect
	github.com/tiendc/go-deepcopy v1.7.2 // indirect
	github.com/twmb/franz-go v1.21.5 // indirect
	github.com/twmb/franz-go/pkg/kadm v1.18.0 // indirect
	github.com/twmb/franz-go/pkg/kmsg v1.13.1 // indirect
	github.com/unixdj/qr v0.8.2 // indirect
	github.com/valyala/fastjson v1.6.10 // indirect
	github.com/xuri/efp v0.0.1 // indirect
	github.com/xuri/excelize/v2 v2.11.0 // indirect
	github.com/xuri/nfp v0.0.2-0.20250530014748-2ddeb826f9a9 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/otel v1.44.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v1.44.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp v1.44.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.44.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.44.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.44.0 // indirect
	go.opentelemetry.io/otel/metric v1.44.0 // indirect
	go.opentelemetry.io/otel/sdk v1.44.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.44.0 // indirect
	go.opentelemetry.io/otel/trace v1.44.0 // indirect
	go.opentelemetry.io/proto/otlp v1.10.0 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	golang.org/x/crypto v0.56.0 // indirect
	golang.org/x/net v0.58.0 // indirect
	golang.org/x/oauth2 v0.36.0 // indirect
	golang.org/x/sync v0.22.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/text v0.41.0 // indirect
	golang.org/x/tools v0.48.0 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20260526163538-3dc84a4a5aaa // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260526163538-3dc84a4a5aaa // indirect
	google.golang.org/grpc v1.83.2 // indirect
	google.golang.org/protobuf v1.36.12-0.20260120151049-f2248ac996af // indirect
)
