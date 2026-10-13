# Golib: independent Go libraries

[![Tooling CI](https://github.com/faustbrian/go-library-tools/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/faustbrian/go-library-tools/actions/workflows/ci.yml)
[![Tooling release](https://img.shields.io/github/v/release/faustbrian/go-library-tools?sort=semver)](https://github.com/faustbrian/go-library-tools/releases)
[![Canonical documentation](https://img.shields.io/badge/docs-Markdown-blue)](docs/index.md)
[![Security policy](https://img.shields.io/badge/security-policy-blue)](security/README.md)

Golib is a collection of independently released, standard-library-first Go
libraries for explicitly composed services and domain work. Choose only the
modules your application needs; applications own configuration, dependency
wiring, policy order, resource acquisition and reverse-order cleanup. There is
no umbrella module, service container, hidden initialization, global registry
or synchronized release train.

## Start with your task

| Problem | Start here |
| --- | --- |
| Evaluate and install one package | [Catalog](docs/packages.md) and [selection guide](docs/choosing-packages.md) |
| Build an API or RPC service | [Protocol decisions](docs/api-protocols.md), [stacks](docs/recommended-stacks.md), [recipes](docs/recipes/index.md) |
| Build a worker, ingester or scheduler | [Role/processing recipes](docs/recipes/index.md) |
| Move incrementally from Laravel/PHP | [Migration route](docs/migration/laravel-php.md) |
| Deploy and operate Kubernetes services | [Operator route](docs/operations/kubernetes.md) |
| Extend a library or adapter | [Contributor route](docs/contributing.md) |
| Review security, architecture or tradeoffs | [Ownership](docs/integration-map.md), [security](security/README.md), [comparisons](docs/comparisons/index.md) |

The [complete documentation map](docs/index.md) connects all audiences and
related next steps. Canonical Markdown is usable directly on GitHub; package
API docs and pkg.go.dev links remain with the owning module.

## Packages by problem

| Primary family | Responsibility and selection |
| --- | --- |
| Foundations | Explicit time, configuration, identifiers and foundational contracts |
| Service edge | Process lifecycle, routing, HTTP policy, authentication and authorization |
| Protocols and descriptions | Resource/operation protocols, schema and API descriptions |
| Persistence and durability | PostgreSQL, migrations, cache, queue, lease, idempotency, outbox and workflow |
| Resilience | Bounded retry, admission/rate policy and circuit breaking |
| Observability | Logging, telemetry and correlation with caller-owned providers |
| Integration and data movement | Filesystems, tabular data, broker/event movement and ingestion |
| Domain utilities | Time/calendar/geography/localization, numerical values and other domain owners |
| Tooling | Repository validation, release integrity and engineering fixtures |

The [consumer catalog](catalog-consumer.md) supplies each installable module's
purpose, status, path and version; the [selection guide](docs/choosing-packages.md)
explains alternatives and deliberate non-goals. Families are ownership groups,
not import layers. The [engineering inventory](catalog-engineering.md) also
contains planned work and non-production fixtures; it is not an install list.

## Five-minute minimal service path

With Go 1.27.0 installed, follow the [minimal HTTP recipe](docs/recipes/01-minimal-http-service.md).
Its real loopback server exercises a business request, readiness withdrawal and
bounded shutdown. The complete public source owns configuration, imports and
wiring; do not deploy the fixture as a product. The
[recipe index](docs/recipes/index.md#run-a-reference) explains clean checkout,
exact compatible versions and task-owned resource cleanup. For application code,
continue to the selected Service README/adoption guide from the catalog.

## Install and version independently

In an initialized application module, install only a chosen published module;
for example, the frozen cohort's standalone clock seam:

```sh
go get github.com/faustbrian/go-clock@v1.1.0
```

Follow its [versioned README](https://github.com/faustbrian/go-clock/blob/v1.1.0/README.md)
and [API reference](https://pkg.go.dev/github.com/faustbrian/go-clock@v1.1.0).
Nested adapters have separate module requirements and repository-relative tag
prefixes. Major versions retain Go's `/vN` import/module suffix; development and
Git tags stay on main, not version-specific source branches. See
[installation and release policy](docs/versioning.md).

## Stability, compatibility and limitations

Design-language **1.0** and published **`golib-compat-v1-20260912.2`** describe
exact public versions on Go **1.27.0**, their platform/backend assumptions and
fourteen bounded composition results. They are a frozen cohort, not the latest
version of every module. Package-current releases and majors have their own
migration and compatibility contracts. [Status](docs/status.md),
[compatibility](compatibility-sets.md) and [limitations](docs/limitations.md)
distinguish active, experimental, planned, deprecated and unavailable work.
Planned Identity-family roots and Secret Store have no public install route.
No badge here certifies every library, backend or production deployment: CI and
release badges refer to this tooling repository; the other badges link documents.

## Continue

[Architecture](docs/architecture.md) · [Integration map](docs/integration-map.md) ·
[Recommended stacks](docs/recommended-stacks.md) · [Recipes](docs/recipes/index.md) ·
[Migration](docs/migration/index.md) · [Operations](docs/operations/index.md) ·
[Tradeoffs](docs/comparisons/index.md) · [Benchmarks](docs/comparisons/performance.md) ·
[Contribution](docs/contributing.md) · [Validation](docs/validation.md) ·
[Glossary](docs/glossary.md)

Maintainers: [catalog publication and maintained consumer](docs/catalog-publication.md),
[release integrity](../release.md), [tooling documentation](../README.md).
