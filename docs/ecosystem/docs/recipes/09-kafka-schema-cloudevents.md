# Kafka event flow with schema and CloudEvents

[Recipe index](index.md) · [Documentation map](../index.md)

This route consumes the published [`golib-compat-v1-20260912.2`](../../compatibility-sets.md)
recipe `09-kafka-schema-cloudevents`. The [complete source](https://github.com/faustbrian/go-cloudevents/tree/644650726f142427df506c9be14a3f9c0576007d/integration/target-adapters) contains imports, configuration,
explicit wiring and executable assertions. It is a reference, not a production
runtime dependency. Required owner boundaries: **CloudEvents target adapters, Kafka, Schema Registry and JSON Schema**. Optional adapters
are selected explicitly; exact versions and dependencies are in the source
module and the frozen [catalog](../packages.md), not a second version table.

## Construction and work lifecycle

Caller supplies Kafka/schema seams and schema policy; compile/validate before publication and construct producer/consumer explicitly.

Invalid events never publish. Decode/validate, perform application handling, then settle the exact topic/partition/offset; handler failure does not commit.

## Failure, ownership and cleanup

Commit failure remains unknown rather than successful settlement. Shut down consumer before producer; still clean producer when consumer cleanup fails and preserve both errors.

## Execute the reference

Follow the [clean checkout and version-rebase procedure](index.md#run-a-reference)
with these exact inputs:

| Input | Value |
| --- | --- |
| Repository | `https://github.com/faustbrian/go-cloudevents.git` |
| Source | `644650726f142427df506c9be14a3f9c0576007d` |
| Module directory | `integration/target-adapters` |

From that prepared module directory:

```sh
go test -mod=readonly -run '^(Example_kafkaSchemaCloudEventFlow|TestKafkaSchemaCloudEventRecipeRejectsInvalidEventBeforePublish|TestKafkaSchemaCloudEventRecipeDoesNotAcknowledgeHandlerFailure|TestKafkaSchemaCloudEventRecipePreservesCommitFailureWithoutClaimingSettlement|TestKafkaSchemaCloudEventRecipeJoinsOperationAndOrderedShutdownFailures)$' -count=1 .
```

## Evidence and deployment boundary

Bounded application-owned Kafka/schema seams; no broker-runtime interoperability claim.

For Kubernetes, retain the [operator-owned setup and shutdown](../operations/kubernetes.md)
instead of deploying this fixture as a product. Keep actual backend versions,
application payloads and recovery policy in the deployment's own runbook.

Next: [supported stacks](../recommended-stacks.md), [operations](../operations/index.md),
[limitations](../limitations.md).
