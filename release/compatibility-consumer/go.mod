module github.com/faustbrian/go-library-tools/release/compatibility-consumer

go 1.27.0

require (
	github.com/faustbrian/go-adaptive-throttle v1.0.1
	github.com/faustbrian/go-analysis v1.2.0
	github.com/faustbrian/go-api-query v1.1.0
	github.com/faustbrian/go-api-query/v3 v3.0.0
	github.com/faustbrian/go-api-query/v4 v4.0.0
	github.com/faustbrian/go-audit v1.0.0
	github.com/faustbrian/go-audit/postgres v1.0.0
	github.com/faustbrian/go-audit/v2 v2.0.0
	github.com/faustbrian/go-authentication v1.2.0
	github.com/faustbrian/go-authentication/adapters/otel v1.0.0
	github.com/faustbrian/go-authentication/jwt v1.1.0
	github.com/faustbrian/go-authentication/oidc v1.0.0
	github.com/faustbrian/go-authentication/v2 v2.0.0
	github.com/faustbrian/go-authorization v1.1.0
	github.com/faustbrian/go-authorization/v3 v3.0.0
	github.com/faustbrian/go-barcode v1.0.0
	github.com/faustbrian/go-barcode/v2 v2.0.0
	github.com/faustbrian/go-bulkhead v1.0.0
	github.com/faustbrian/go-cache v1.0.0
	github.com/faustbrian/go-cache/v2 v2.0.0
	github.com/faustbrian/go-calendar v1.1.0
	github.com/faustbrian/go-calendar/v2 v2.0.0
	github.com/faustbrian/go-capability v1.1.0
	github.com/faustbrian/go-capability/v2 v2.0.0
	github.com/faustbrian/go-circuit-breaker v1.0.0
	github.com/faustbrian/go-cli v1.0.0
	github.com/faustbrian/go-clock v1.2.0
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
	github.com/faustbrian/go-cloudevents/adapters/workflow/v2 v2.0.0
	github.com/faustbrian/go-concurrency-limit v1.0.0
	github.com/faustbrian/go-config/adapters/awssecretsmanager/v2 v2.0.0
	github.com/faustbrian/go-config/v2 v2.0.0
	github.com/faustbrian/go-correlation v1.1.1
	github.com/faustbrian/go-ecma-regexp v1.0.0
	github.com/faustbrian/go-event-sourcing/adapters/kafka/v2 v2.0.0
	github.com/faustbrian/go-event-sourcing/adapters/otel/v2 v2.0.0
	github.com/faustbrian/go-event-sourcing/adapters/outbox/v2 v2.0.0
	github.com/faustbrian/go-event-sourcing/adapters/outbox/v3 v3.0.0
	github.com/faustbrian/go-event-sourcing/adapters/queue/v2 v2.0.0
	github.com/faustbrian/go-event-sourcing/postgres/v2 v2.0.2
	github.com/faustbrian/go-event-sourcing/v2 v2.0.1
	github.com/faustbrian/go-external-sort v1.0.0
	github.com/faustbrian/go-fault-injection/v2 v2.0.0
	github.com/faustbrian/go-feature-flags/v2 v2.0.0
	github.com/faustbrian/go-filesystem/v2 v2.0.0
	github.com/faustbrian/go-geo v1.1.2
	github.com/faustbrian/go-hedge v1.0.2
	github.com/faustbrian/go-http-client v1.1.0
	github.com/faustbrian/go-http-client/v2 v2.0.0
	github.com/faustbrian/go-http-middleware/v2 v2.0.0
	github.com/faustbrian/go-http-signature v1.0.0
	github.com/faustbrian/go-idempotency v1.2.0
	github.com/faustbrian/go-idempotency/v2 v2.0.1
	github.com/faustbrian/go-identifier v1.0.0
	github.com/faustbrian/go-identifier/v2 v2.0.0
	github.com/faustbrian/go-international/v3 v3.0.0
	github.com/faustbrian/go-international/v4 v4.0.0
	github.com/faustbrian/go-json-schema v1.0.0
	github.com/faustbrian/go-json-schema/v2 v2.0.0
	github.com/faustbrian/go-jsonapi v1.0.0
	github.com/faustbrian/go-jsonapi/v2 v2.0.0
	github.com/faustbrian/go-jsonrpc v1.0.0
	github.com/faustbrian/go-kafka v1.1.0
	github.com/faustbrian/go-kafka/adapters/mskiam v1.1.0
	github.com/faustbrian/go-kafka/adapters/otel v1.0.0
	github.com/faustbrian/go-kafka/adapters/service v1.0.0
	github.com/faustbrian/go-keyphrase v1.0.0
	github.com/faustbrian/go-knapsack/objective/money/v3 v3.0.0
	github.com/faustbrian/go-knapsack/v2 v2.0.0
	github.com/faustbrian/go-lease v1.1.0
	github.com/faustbrian/go-localized v1.1.0
	github.com/faustbrian/go-localized/v3 v3.0.0
	github.com/faustbrian/go-localized/v5 v5.0.0
	github.com/faustbrian/go-log/v2 v2.0.0
	github.com/faustbrian/go-math v1.1.2
	github.com/faustbrian/go-measurement/v2 v2.0.1
	github.com/faustbrian/go-measurement/v3 v3.0.0
	github.com/faustbrian/go-merkle-patricia-trie v1.0.0
	github.com/faustbrian/go-merkle-patricia-trie/v2 v2.0.0
	github.com/faustbrian/go-merkle-tree v1.0.0
	github.com/faustbrian/go-migrations v1.1.0
	github.com/faustbrian/go-migrations/v2 v2.0.0
	github.com/faustbrian/go-money/v2 v2.0.0
	github.com/faustbrian/go-openapi v1.0.0
	github.com/faustbrian/go-openapi/v2 v2.0.0
	github.com/faustbrian/go-opening-hours v1.1.0
	github.com/faustbrian/go-opening-hours/v2 v2.0.1
	github.com/faustbrian/go-opening-hours/v3 v3.0.0
	github.com/faustbrian/go-opening-hours/v4 v4.0.0
	github.com/faustbrian/go-openrpc/v2 v2.0.0
	github.com/faustbrian/go-password/v2 v2.0.0
	github.com/faustbrian/go-postgres/v2 v2.0.0
	github.com/faustbrian/go-prompts v1.1.2
	github.com/faustbrian/go-queue v1.1.3
	github.com/faustbrian/go-queue-control-plane v1.1.0
	github.com/faustbrian/go-queue-control-plane/v2 v2.0.1
	github.com/faustbrian/go-queue-control-plane/v3 v3.0.0
	github.com/faustbrian/go-queue/adapters/rabbitmq v1.0.0
	github.com/faustbrian/go-queue/adapters/service v1.0.1
	github.com/faustbrian/go-rabbitmq-queues v1.1.0
	github.com/faustbrian/go-rabbitmq-streams v1.1.1
	github.com/faustbrian/go-rabbitmq-streams/adapters/otel v1.0.0
	github.com/faustbrian/go-rabbitmq-streams/adapters/rabbitmq v1.0.1
	github.com/faustbrian/go-rate-limit v1.1.0
	github.com/faustbrian/go-rate-limit/v2 v2.0.0
	github.com/faustbrian/go-resilience v1.0.0
	github.com/faustbrian/go-retry v1.1.0
	github.com/faustbrian/go-router/v2 v2.0.0
	github.com/faustbrian/go-rule-engine v1.0.0
	github.com/faustbrian/go-rule-engine/adapters/math v1.0.0
	github.com/faustbrian/go-rule-engine/adapters/measurement/v2 v2.0.0
	github.com/faustbrian/go-rule-engine/adapters/temporal v1.0.0
	github.com/faustbrian/go-scheduler v1.1.0
	github.com/faustbrian/go-schema-registry v1.0.0
	github.com/faustbrian/go-schema-registry/providers/confluent v1.0.0
	github.com/faustbrian/go-schema-registry/providers/confluent/v2 v2.0.0
	github.com/faustbrian/go-schema-registry/providers/glue v1.0.0
	github.com/faustbrian/go-schema-registry/providers/glue/v2 v2.0.0
	github.com/faustbrian/go-schema-registry/v2 v2.0.0
	github.com/faustbrian/go-search v1.0.0
	github.com/faustbrian/go-search/adapters/opensearch v1.0.0
	github.com/faustbrian/go-secret-envelope/v2 v2.0.0
	github.com/faustbrian/go-semaphore/v2 v2.0.0
	github.com/faustbrian/go-sequencer v1.1.0
	github.com/faustbrian/go-sequencer/v2 v2.0.0
	github.com/faustbrian/go-service v1.1.2
	github.com/faustbrian/go-settings/v2 v2.0.0
	github.com/faustbrian/go-state-machine/v2 v2.0.0
	github.com/faustbrian/go-tabular v1.0.0
	github.com/faustbrian/go-tabular/v2 v2.0.0
	github.com/faustbrian/go-telemetry v1.2.0
	github.com/faustbrian/go-temporal v1.1.0
	github.com/faustbrian/go-temporal/v2 v2.0.0
	github.com/faustbrian/go-tenancy/v2 v2.0.0
	github.com/faustbrian/go-transactional-outbox v1.0.0
	github.com/faustbrian/go-transactional-outbox/adapters/kafka v1.0.0
	github.com/faustbrian/go-transactional-outbox/adapters/kafka/v2 v2.0.0
	github.com/faustbrian/go-transactional-outbox/adapters/otel v1.0.0
	github.com/faustbrian/go-transactional-outbox/adapters/otel/v2 v2.0.0
	github.com/faustbrian/go-transactional-outbox/adapters/queue v1.0.0
	github.com/faustbrian/go-transactional-outbox/adapters/queue/v2 v2.0.0
	github.com/faustbrian/go-transactional-outbox/adapters/rabbitstream v1.0.1
	github.com/faustbrian/go-transactional-outbox/adapters/rabbitstream/v2 v2.0.0
	github.com/faustbrian/go-transactional-outbox/v2 v2.0.0
	github.com/faustbrian/go-validation/v2 v2.0.0
	github.com/faustbrian/go-verkle-tree v1.0.0
	github.com/faustbrian/go-webhook v1.0.0
	github.com/faustbrian/go-webhook/v3 v3.0.0
	github.com/faustbrian/go-wire v1.0.1
	github.com/faustbrian/go-wire/v3 v3.0.0
	github.com/faustbrian/go-workflow v1.0.0
	github.com/faustbrian/go-workflow/v2 v2.0.0
	github.com/faustbrian/go-wsdl v1.0.0
	github.com/faustbrian/go-wsdl/v2 v2.0.0
	github.com/faustbrian/go-xsd v1.0.0
)

require (
	github.com/BurntSushi/toml v1.6.0 // indirect
	github.com/appleboy/com v1.2.2 // indirect
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
	github.com/bufbuild/protocompile v0.14.1 // indirect
	github.com/cenkalti/backoff/v5 v5.0.3 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/consensys/gnark-crypto v0.20.1 // indirect
	github.com/coreos/go-oidc/v3 v3.20.0 // indirect
	github.com/crate-crypto/go-ipa v0.0.0-20240223125850-b1e8a79f509c // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.4.1 // indirect
	github.com/deszhou/jcs v1.0.0 // indirect
	github.com/dlclark/regexp2 v1.12.0 // indirect
	github.com/dlclark/regexp2/v2 v2.8.2 // indirect
	github.com/dunglas/httpsfv v1.1.0 // indirect
	github.com/ericlevine/zxinggo v0.1.0 // indirect
	github.com/faustbrian/go-event-sourcing v1.0.0 // indirect
	github.com/faustbrian/go-international v1.1.0 // indirect
	github.com/faustbrian/go-postgres v1.1.0 // indirect
	github.com/faustbrian/go-telemetry/v2 v2.0.0 // indirect
	github.com/faustbrian/go-tenancy v1.1.0 // indirect
	github.com/faustbrian/go-validation v1.1.0 // indirect
	github.com/felixge/httpsnoop v1.1.0 // indirect
	github.com/fxamacker/cbor/v2 v2.9.4 // indirect
	github.com/go-jose/go-jose/v4 v4.1.4 // indirect
	github.com/go-logr/logr v1.4.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/goccy/go-json v0.10.6 // indirect
	github.com/golang/snappy v1.0.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.29.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.11.0 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/jpillora/backoff v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.20.0 // indirect
	github.com/lestrrat-go/blackmagic v1.0.4 // indirect
	github.com/lestrrat-go/dsig v1.2.1 // indirect
	github.com/lestrrat-go/dsig-secp256k1 v1.0.0 // indirect
	github.com/lestrrat-go/httpcc v1.0.1 // indirect
	github.com/lestrrat-go/httprc/v3 v3.0.5 // indirect
	github.com/lestrrat-go/jwx/v3 v3.1.1 // indirect
	github.com/lestrrat-go/option/v2 v2.0.0 // indirect
	github.com/linkedin/goavro/v2 v2.15.0 // indirect
	github.com/makiuchi-d/gozxing v0.1.1 // indirect
	github.com/mfridman/interpolate v0.0.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.3-0.20250322232337-35a7c28c31ee // indirect
	github.com/nats-io/nats.go v1.52.0 // indirect
	github.com/nats-io/nkeys v0.4.15 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/nsqio/go-nsq v1.1.0 // indirect
	github.com/nyaruka/phonenumbers v1.8.1 // indirect
	github.com/oklog/ulid/v2 v2.1.1 // indirect
	github.com/opensearch-project/opensearch-go/v4 v4.7.3 // indirect
	github.com/peterstace/simplefeatures v0.59.0 // indirect
	github.com/pierrec/lz4 v2.6.1+incompatible // indirect
	github.com/pierrec/lz4/v4 v4.1.26 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pressly/goose/v3 v3.27.1 // indirect
	github.com/rabbitmq/amqp091-go v1.14.0 // indirect
	github.com/rabbitmq/rabbitmq-stream-go-client v1.8.3 // indirect
	github.com/redis/go-redis/v9 v9.21.0 // indirect
	github.com/richardlehane/mscfb v1.0.8 // indirect
	github.com/richardlehane/msoleps v1.0.6 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	github.com/santhosh-tekuri/jsonschema/v6 v6.0.2 // indirect
	github.com/segmentio/asm v1.2.1 // indirect
	github.com/segmentio/ksuid v1.0.4 // indirect
	github.com/sethvargo/go-retry v0.3.0 // indirect
	github.com/spaolacci/murmur3 v1.1.0 // indirect
	github.com/tiendc/go-deepcopy v1.7.2 // indirect
	github.com/twmb/franz-go v1.21.5 // indirect
	github.com/twmb/franz-go/pkg/kadm v1.18.0 // indirect
	github.com/twmb/franz-go/pkg/kmsg v1.13.1 // indirect
	github.com/unixdj/qr v0.8.2 // indirect
	github.com/valkey-io/valkey-go v1.0.78 // indirect
	github.com/valyala/fastjson v1.6.10 // indirect
	github.com/vmihailenco/msgpack/v5 v5.4.1 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	github.com/xuri/efp v0.0.2 // indirect
	github.com/xuri/excelize/v2 v2.11.1-0.20260930021559-01a9ff32fb3c // indirect
	github.com/xuri/nfp v0.0.2-0.20250530014748-2ddeb826f9a9 // indirect
	go.mongodb.org/mongo-driver/v2 v2.9.1 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/otel v1.46.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v1.45.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp v1.45.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.45.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.45.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.45.0 // indirect
	go.opentelemetry.io/otel/metric v1.46.0 // indirect
	go.opentelemetry.io/otel/sdk v1.46.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.46.0 // indirect
	go.opentelemetry.io/otel/trace v1.46.0 // indirect
	go.opentelemetry.io/proto/otlp v1.11.0 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.uber.org/mock v0.6.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.yaml.in/yaml/v2 v2.4.3 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	go.yaml.in/yaml/v4 v4.0.0-rc.6 // indirect
	golang.org/x/crypto v0.57.0 // indirect
	golang.org/x/net v0.59.0 // indirect
	golang.org/x/oauth2 v0.37.0 // indirect
	golang.org/x/sync v0.23.0 // indirect
	golang.org/x/sys v0.48.0 // indirect
	golang.org/x/term v0.46.0 // indirect
	golang.org/x/text v0.42.0 // indirect
	golang.org/x/tools v0.49.0 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20260803160001-6ac0973c030d // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260803160001-6ac0973c030d // indirect
	google.golang.org/grpc v1.83.2 // indirect
	google.golang.org/protobuf v1.36.12-0.20260120151049-f2248ac996af // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	k8s.io/api v0.36.2 // indirect
	k8s.io/apimachinery v0.36.2 // indirect
	k8s.io/klog/v2 v2.140.0 // indirect
	k8s.io/kube-openapi v0.0.0-20260317180543-43fb72c5454a // indirect
	k8s.io/utils v0.0.0-20260210185600-b8788abfbbc2 // indirect
	sigs.k8s.io/json v0.0.0-20250730193827-2d320260d730 // indirect
	sigs.k8s.io/randfill v1.0.0 // indirect
	sigs.k8s.io/structured-merge-diff/v6 v6.3.2 // indirect
)
