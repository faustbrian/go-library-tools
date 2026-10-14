# Common Identity Goal Requirements

The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT",
"SHOULD", "SHOULD NOT", "RECOMMENDED", "NOT RECOMMENDED", "MAY", and
"OPTIONAL" in this document are to be interpreted as described in BCP 14
[RFC2119] [RFC8174] when, and only when, they appear in all capitals, as
shown here.

## 2026-09-30 shared-issuer amendment

[`SHARED_ISSUER_PROFILE.md`](SHARED_ISSUER_PROFILE.md) is the user-authorized
addition to the existing goals. Read it before older contract/reference shapes
for an affected unit. It selects owned OAuth delegation, bounded revocation,
legacy-key compatibility and a focused runtime milestone. Its selected-profile
requirements override older defaults only as explicitly stated there; other
profiles and the full 67-unit completion predicate remain unchanged. Preserve
active ownership, existing implementation and unrelated verification evidence.

These requirements apply to every manifest-registered goal at its planning
path. They MUST be executed under
`/Users/brian/Developer/golib/.ai/TOPOLOGY.md` and
`REPOSITORY_TOPOLOGY.md`. Historical `pkg/...` metadata is a logical label and
MUST be resolved through that map before any filesystem, Git, module, or import
operation.

The shared-issuer profile is additionally binding on identity, organization,
API-key, OAuth, resource-verification and HTTP units and applicable adapters.
It is a canonical profile requirement, not a discretionary extra selector in
the existing five-selector applicability manifest. Its SI-01 through SI-10
acceptance cases supplement the existing artifacts.

The following coordinator specifications are binding. Applicability MUST NOT be
left to worker discretion. `SHARED_CONTRACT_APPLICABILITY.json` is the sole
canonical per-unit selector manifest. Every inventory unit MUST have exactly
the five selectors `transaction`, `lifecycle`, `lifecycle_consumers`,
`configuration`, and `security_events`. Each selector is a bytewise-sorted,
duplicate-free array of exact catalog IDs; the sole no-applicability value is
`["none"]`, and `none` MUST NOT coexist with another value. Goal files MUST NOT
duplicate these selectors because such copies can drift when a goal moves.

The primary agent MUST inspect the selected IDs and their canonical source rows
before implementing an affected milestone. The existing renderer may help locate
them; rendering/verbatim worker assignments are not development start gates.
Selector completeness and canonical encoding remain final product-catalog
requirements, not a new approval or receipt workflow.

Selecting any `tx.*` ID MUST include `tx.foundation`; selecting any
`lifecycle.*` ID MUST include `lifecycle.foundation`. Configuration IDs use the
exact `ref.<template-path>` form defined by `REFERENCE_CONFIGURATION.md`.
`lifecycle_consumers` values are exact cascade row IDs from
`LIFECYCLE_CONSUMERS.md`; every owning or consuming inventory unit MUST select
the row, and its `lifecycle` selector MUST select the same cascade.

- `TRANSACTION_CONTRACT.md` for every PostgreSQL mutation, cross-module
  transition, one-time capability, reveal-once result, and unknown commit;
- `LIFECYCLE_CASCADES.md` for every identity, credential, session,
  organization, authorization, provider, client, and destructive lifecycle;
- `LIFECYCLE_CONSUMERS.md` for every unit that owns or consumes a cascade;
- `REFERENCE_CONFIGURATION.md`, the canonical machine-value authority, for
  every configurable package and for `identity/reference`; and
- `SECURITY_EVENTS.md` for every exact audit action emitted by a unit.

A package goal MUST NOT be interpreted to permit a weaker or different local
contract. If an existing primitive cannot implement a required shared
contract, the primary agent MUST implement the required upstream behavior as
part of the same milestone or identify its concrete unavailable authority.
It MUST NOT add a private substitute or undocumented application glue.

## Milestone execution and ownership

IMPLEMENTATION_GOAL.md owns execution sequencing. The primary agent owns package
implementation and control-state writes across the necessary mapped repositories,
without worker assignment ceremonies or package-at-a-time walls. It MUST record
one frozen consumer milestone in EXECUTION_LEDGER.md, preserve historical
assignments and inspect complete selected goals and applicable contract rows.

Development requires committed, integrated, focused behavioral/compatibility proof
of the contracts actually consumed. Requires names ownership dependencies, not
whole-unit release barriers for unrelated profiles. A stub or compile-only result
MUST NOT unlock a security dependency. Final verified promotion still requires
verified prerequisites and every applicable complete unit gate.

Product scope/ownership/semantics remain governed by actual user authority under
PROGRAM.md. Existing user authorization MUST NOT require exact-byte replay.
Goals stay at their planning paths; historical canonical locations MUST NOT
cause file relocation or monorepo creation.

The agent MUST implement each owned/consumed parity, configuration, transaction
and lifecycle requirement. Every selected row needs its public owner and
behavioral proof before the relevant final acceptance. Core zero values remain
unambiguous; identity/reference owns selected deployment defaults.
New mandatory dependencies or ownership transfers outside authorized scope need
a material decision; necessary in-scope prerequisite work is not a new approval.

## Module and API contract

- The module MUST live at the repository and repository-relative directory
  resolved by `REPOSITORY_TOPOLOGY.md`, use the mapped
  `github.com/faustbrian/go-*` module path, and be declared in that repository's
  local module/package catalog.
- Public contracts for operations that own or perform those concerns MUST expose
  ownership, cancellation, resource limits, concurrency, retry and timeout
  policy, stable errors, and lifecycle.
- Every public type whose zero value is invalid MUST be constructible in a
  clean external consumer. The contract MUST identify exactly one of: a public
  constructor or parser with bounded validation and stable errors; an
  authority-issued result method that returns the value before any caller must
  supply it; or an exact existing external type with that construction
  contract. An unexported-field value MUST NOT appear in a caller-created
  request, configuration, contributor input, persistence reconstruction, or
  interface argument unless a reachable public constructor/parser produces it.
  Goal review and clean-consumer evidence MUST compile one valid construction
  path and every invalid-zero rejection path for each such type. Placeholder
  prose that a constructor, transition, or query may return a value is not an
  API contract and MUST block scheduling until the exact symbol is named.
- Core modules MUST define consumer-oriented interfaces. Persistence,
  protocols, transports, and providers MUST remain in named adapters.
- Cross-module dependencies MUST remain acyclic. Permanent `replace`
  directives and private sibling imports are forbidden.
- Every external operation MUST accept or derive a bounded
  `context.Context`. Callbacks MUST NOT run under locks or open transactions
  unless the public contract explicitly transfers that ownership.
- Public interfaces shared by two or more program modules MUST have one named
  upstream owner and a compile-time conformance test in every implementation.
  A downstream worker MUST NOT invent a parallel transaction, event, version,
  capability, locale, cookie, token, or authorization contract.

## Identity and authorization invariants

- Every subject, credential, session, organization, client, and provider MUST
  have a stable opaque identifier. User input MUST NOT become a storage key
  without canonicalization and collision analysis.
- Enumeration-sensitive operations MUST return indistinguishable public
  outcomes while preserving redacted internal reason codes.
- Authentication MUST NOT imply authorization. Every privileged operation MUST
  require an explicit decision and emit an audit record.
- Tenant and organization scope MUST be explicit and fail closed. Cross-scope
  reads, writes, linking, replay, and cache collisions MUST be tested.
- Secrets, credentials, tokens, codes, cookies, recovery material, provider
  tokens, challenges, and PII MUST NOT appear in errors, logs, traces, metric
  labels, fixtures, snapshots, or evidence.
- Bearer-equivalent values MUST be irreversibly digested when lookup permits.
  Recoverable secrets MUST use `secret-envelope` with authenticated context
  and a rotation policy.
- Security state changes MUST define atomicity, idempotency, replay,
  revocation propagation, audit semantics, and ambiguous-commit recovery.
- Security-sensitive versions and invalidations MUST use the dimensions and
  monotonic semantics in `LIFECYCLE_CASCADES.md`. A cache or stateless token
  MUST bind every dimension capable of invalidating the authority it carries.

## Persistence and migration contract

- Core modules MUST remain storage-neutral. Durable adapters own schemas,
  transactions, constraints, indexes, migrations, cleanup, and reconciliation.
- Migrations MUST be forward-only, interruption-safe, and safe with concurrent
  old/new binaries. Race-sensitive invariants MUST be database-enforced.
- Stores MUST specify consistency, pagination, locking, clock ownership,
  duplicates, retention, deletion, and not-committed/committed/unknown outcomes.
- Deletion and anonymization MUST account for audit, outbox, revocation,
  legal-hold, and referential-integrity boundaries without claiming deletion
  outside owned stores.
- PostgreSQL adapters participating in a shared mutation MUST use the common
  unit-of-work and command-result protocol in `TRANSACTION_CONTRACT.md`.
  Opening a private transaction inside an enlisted operation is forbidden.
- One-time capabilities MUST use the reservation/finalization/query protocol
  in `TRANSACTION_CONTRACT.md`. Consuming a capability in one commit and
  applying its domain transition in another without a durable recovery state
  is forbidden.

## Hostile boundaries and lifecycle

- Parsers and handlers MUST impose limits before allocation, conversion,
  recursion, decompression, or cryptographic work and MUST include fuzz targets
  with deterministic regression seeds.
- Concurrent code MUST document synchronization, cancellation, shutdown,
  channel closure, and goroutine lifetime and pass race, stress, and leak checks.
- Provider proof MUST cover timeouts, cancellation, partial and malformed
  responses, throttling, retries, duplicate callbacks, clock skew, key
  rotation, outage recovery, and redaction. Fakes are not interoperability proof.

## Required verification and release evidence

Each goal MUST produce:

1. Behavioral tests for success, denial, invalid transitions, replay,
   idempotency, cancellation, cleanup, and scope isolation.
2. Exact 100% statement coverage and exact 100% mutation efficacy and mutant
   coverage for every viable mutant.
3. Race, fuzz, hostile-input, leak, and benchmark evidence where applicable.
4. Pinned official fixtures and independent interoperability evidence for
   specification claims; untested profiles MUST be explicit.
5. Public API baseline, clean-consumer proof, dependency and license review,
   vulnerability scanning, secret scanning, SBOM and provenance gates, and
   manifests.
6. README quick start, API/adoption/security/operations/tradeoff/FAQ docs,
   compilable examples, and an Unreleased `CHANGELOG.md` entry.
7. The narrow module gate and reverse-dependant gates whose complete input
   fingerprints changed, using task-owned disposable Go caches.

A unit may move to `implemented-unverified` when implementation and docs are
complete. It may move to `verified` only after every goal requirement and
per-unit verification gate passes against its current exhaustive inputs. Gaps,
warning substitutions,
unavailable provider proof, or stale fingerprints block `verified`.

The primary agent records status from executed native results, never from a
helper or compilation count. A unit-gate result is not the later full-program
acceptance result. Before final acceptance, revalidate or validly reuse each
affected gate under complete-input fingerprint rules. Independent final review
remains required; evidence machinery and historical worker ceremonies MUST NOT
substitute for consumer proof.
