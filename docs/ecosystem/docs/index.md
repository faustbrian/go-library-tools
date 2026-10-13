# Documentation map

[Ecosystem entry point](../README.md)

Start with a task, not a list of dependencies. Import only the independently
versioned modules your application uses. The nine families describe semantic
ownership, not import layers or a shared release train.

## Choose your route

| Audience | Start here | Continue with |
| --- | --- | --- |
| Evaluate one library | [Package catalog](packages.md) | Its README, adoption guide and limitations |
| Build an HTTP or RPC API | [Protocol selection](api-protocols.md) | [Recommended stacks](recommended-stacks.md) and [recipes](recipes/index.md) |
| Build a worker, ingester or scheduler | [Choose packages](choosing-packages.md) | [Recipes](recipes/index.md) and [operations](operations/index.md) |
| Migrate a Laravel/PHP application | [Migration](migration/index.md) | Coexistence, payload contracts and rollback |
| Operate Kubernetes services | [Operations](operations/index.md) | Dependency readiness, drain, migrations and incident handling |
| Contribute a library or adapter | [Contributor route](contributing.md) | Owning repository policy and focused verification |
| Review security or architecture | [Ownership map](integration-map.md) | [Architecture](architecture.md), [limitations](limitations.md) and [security](../security/README.md) |

## Complete map

- [Packages](packages.md), [choosing packages](choosing-packages.md),
  [protocols](api-protocols.md), [stacks](recommended-stacks.md).
- [Architecture](architecture.md), [integration map](integration-map.md),
  [glossary](glossary.md), [design language 1.0](../design-language.md).
- [Runnable recipes](recipes/index.md), [migration](migration/index.md),
  [operations](operations/index.md), [contribution](contributing.md).
- [Comparisons](comparisons/index.md), [benchmark interpretation](comparisons/performance.md),
  [limitations](limitations.md), [status](status.md), [versioning](versioning.md).
- [Consumer catalog](../catalog-consumer.md), [engineering inventory](../catalog-engineering.md),
  [compatibility set](../compatibility-sets.md), [security](../security/README.md).
- [Catalog publication](catalog-publication.md), [documentation validation](validation.md) and
  [tooling documentation](../../README.md).

Frozen catalog and recipe facts describe the published compatibility cohort.
Package-current changes are supported only by their own releases and evidence;
they do not silently join that cohort. See [versioning](versioning.md).

Next: [choose packages](choosing-packages.md) or [run a recipe](recipes/index.md).
