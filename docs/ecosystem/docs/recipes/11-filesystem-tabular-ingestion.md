# Filesystem and tabular ingestion

[Recipe index](index.md) · [Documentation map](../index.md)

This route consumes the published [`golib-compat-v1-20260912.2`](../../compatibility-sets.md)
recipe `11-filesystem-tabular-ingestion`. The [complete source](https://github.com/faustbrian/go-filesystem/tree/098f47cae3fd01b67355062a76aae7d6e29306cd/integration/tabular-ingestion) contains imports, configuration,
explicit wiring and executable assertions. It is a reference, not a production
runtime dependency. Required owner boundaries: **Filesystem and Tabular**. Optional adapters
are selected explicitly; exact versions and dependencies are in the source
module and the frozen [catalog](../packages.md), not a second version table.

## Construction and work lifecycle

Validate total-object/record/field/row bounds, open with caller context, then wrap the stream and construct the bounded parser.

Require the header, deliver rows to the application callback and check cancellation before reads and buffered delivery. No automatic acknowledgement or retry.

## Failure, ownership and cleanup

Stop parser/callback delivery before stream close on success or any failure; retain open/parse/callback/close error causes. Close the owning adapter/session separately after all ingestions.

## Execute the reference

Follow the [clean checkout and version-rebase procedure](index.md#run-a-reference)
with these exact inputs:

| Input | Value |
| --- | --- |
| Repository | `https://github.com/faustbrian/go-filesystem.git` |
| Source | `098f47cae3fd01b67355062a76aae7d6e29306cd` |
| Module directory | `integration/tabular-ingestion` |

From that prepared module directory:

```sh
go test -mod=readonly -run '^(TestIngestCSV|ExampleIngestCSV)' -count=1 .
```

## Evidence and deployment boundary

In-memory filesystem. SFTP/object-storage/R2/other remote variants require their own backend evidence.

For Kubernetes, retain the [operator-owned setup and shutdown](../operations/kubernetes.md)
instead of deploying this fixture as a product. Keep actual backend versions,
application payloads and recovery policy in the deployment's own runbook.

Next: [supported stacks](../recommended-stacks.md), [operations](../operations/index.md),
[limitations](../limitations.md).
