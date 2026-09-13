# Golib Cohesion v1

Status: current public ecosystem contract
Effective: 2026-09-10

This document supersedes earlier unpublished Cohesion v1 delivery contracts,
including schema-v3 fleet rollout, fixed-review counts, recursive provenance,
and per-repository ecosystem-state requirements.

## Purpose

Golib is a set of independently released Go modules that should feel coherent
when selected together. Cohesion covers package discovery, construction,
ownership, lifecycle, errors, adapter placement, module compatibility, and
public composition guidance. It does not create an umbrella module, application
framework, service container, global registry, or lockstep release train.

## Public entry points

- [Design language](../design-language.md) defines shared API and lifecycle
  vocabulary.
- [Consumer catalog](../catalog-consumer.md) lists installable public modules
  and their package-selection guidance.
- [Engineering catalog](../catalog-engineering.md) retains the complete
  implementation inventory.
- [Compatibility sets](../compatibility-sets.md) publish exact known-good
  module versions and focused composition receipts.
- [Residual register](../../../release/cohesion-residuals.json) records only
  remaining material exceptions and follow-ups.

## Module boundaries

Each public package family remains owned by one standalone repository. Nested
modules are used only where independent versioning or dependency isolation has
a consumer benefit. Adapters make the integration target explicit and do not
hide network clients, databases, brokers, telemetry, goroutines, or shutdown
ownership.

Catalog lifecycle has direct consumer meaning:

- **active** modules are recommended for new use;
- **deprecated** modules remain compatible during a documented migration;
- **planned** modules are non-installable design inventory; and
- **retired** modules are excluded from current selection.

## Compatibility sets

A published set is a central recommendation, not dependency management.
It binds every selected module to an exact public version and source revision,
records the supported Go and platform matrix, and cites focused composition
receipts. Consumers may adopt only the modules they use and may select other
compatible versions.

Published set identifiers are immutable. A changed module selection receives a
new identifier. Status, navigation, or receipt wording may reuse existing
behavioral evidence when the selected module/version/source tuples are
unchanged.

## Observable composition

Supported recipes exercise these boundaries once per selected set:

1. minimal HTTP service;
2. internal JSON-RPC service;
3. external JSON:API service;
4. authenticated and authorized service;
5. queue producer and worker;
6. ingester and processor;
7. PostgreSQL idempotency and outbox;
8. scheduled singleton;
9. Kafka, schema, and CloudEvents adapters;
10. vendor HTTP client policy;
11. filesystem and tabular ingestion;
12. durable workflow compensation;
13. searchable track and location projection; and
14. track, postal, location, and role adoption.

A receipt states the exact source, selected public dependencies, platform, and
boundary actually exercised. It does not imply production deployment or a live
external service that the recipe did not use.

## Proportional assurance

Verification follows the risk of the change:

- documentation and metadata receive structural validation;
- internal behavior receives focused and affected-module checks;
- public API, lifecycle, security, persistence, and concurrency receive direct
  behavioral and consumer evidence; and
- public releases and ecosystem milestones receive the relevant clean-consumer,
  composition, aggregate, and artifact checks once.

One complete independent review is the default for meaningful public-contract
or milestone changes. Additional reviews are used only for a distinct named
risk. Mutable plans, prose, ordinary CI logs, and evidence already bound to an
immutable run do not require recursive hashes or repeated reviews.

## Evolution

Tooling and manifest changes remain backward-compatible with supported
repository versions. Optional fields do not force a fleet migration or module
release. Ecosystem-owned facts are derived centrally from repository-owned
manifests and public releases instead of being copied back into every
repository.

Known compatibility or migration gaps stay visible in the residual register
until their documented removal condition is met. Historical contracts remain
available through Git history and published release artifacts but do not
authorize or block current work.
