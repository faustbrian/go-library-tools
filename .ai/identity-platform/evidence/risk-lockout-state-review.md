# Native lockout transition storage reports

The native roles below are isolated, read-only `gpt-6.1-sol`, medium,
`fork_turns="none"`. This checkpoint is forward-only populated storage, not
authorized native start/end execution, ended-state admission or unit completion.
The mapper and QA cover the continuing native mutation batch; retain this input
for that implementation rather than restarting planning for each helper.

## Native change mapper

Principal: `/root/lockout_native_map_4398`

## Read-only change map

**Principal:** `/root/lockout_native_map_4398`
**Reviewed identity base/head:** `4398e2dde31d82dac7005bcbc96610aba0645dcb`
**Scope:** native authorized start/end/CAS transitions through the existing PostgreSQL Coordinator. No writes, tests, builds, or runtime access.

No contradictory product requirement found. Existing policy and admission are explicitly intermediate, not a completed mutation API.

### Existing owners and extension points

| Boundary | Current owner | Required integration |
|---|---|---|
| Immutable restriction intent | `risk/lockout.go` | Preserve pure `LockoutPolicy.Prepare/Validate`, exact actor/operation/subject/version/duration, redaction, and recovery exclusions. Preparation conveys no permission. |
| Native immutable command | `risk/postgres/command.go` provides issuance precedent | Add closed lockout schemas and bounded decoding; bind frozen keyed intent into the existing feature command fingerprint. Target operation and administrative command identity remain distinct. |
| Feature registration/execution | `postgres/feature_registry.go`, `feature_execution.go` | Extend the existing risk Repository’s operation declaration, validation, apply and recovery dispatch. Use its exact registered identity and Coordinator Work; no new transaction or journal owner. |
| Durable state | `risk/postgres/lockout_migrations.go` | Mutate the logical keyed counter root, retaining rotation continuity. Root/alias establishment belongs to the native adapter, never consumer SQL. |
| Current admission | `risk/postgres/counter_admission.go`, `lockout_admission.go` | Preserve current checks before feature apply and commitment. Logical counter is locked before restriction; primary time is sampled after locks. Mutation must follow this lock order. |
| Replay/recovery | Existing Coordinator command journal, feature query/recovery, registered recovery participants | Matching committed replay returns original result without transition replay; unknown returns no success and reconciles the same command. Changed intent conflicts. |
| Authorization | Trusted executor caller, query/admission policies, native current-authority checks | Actor string and successful command observation are not transition authorization. Mutation requires separately authenticated, exact-scope current administrative authority. |
| Audit/lifecycle | Coordinator stages effects; registered audit/outbox contributors persist them | Use canonical start/end actions and lifecycle events atomically with state/result. Do not stage competing adapter-owned effects or add a journal. |

### Concrete schema interaction

The installed schema has only `version`, `policy_version`, `started_at`, and `expires_at`, with `expires_at > started_at`. Admission evaluates that original interval.

**Ending without deletion needs an explicit durable representation and corresponding admission change.** An additive migration for ended/state metadata preserves the original interval and CAS history. Rewriting installed migration bytes or deleting the row would violate retained-history expectations. Setting expiry equal to start violates the existing constraint; merely returning an end receipt leaves admission denying.

A later start must compare the current durable version—including ended or naturally expired state—not treat inactivity as nonexistent history. Exhausted versions must fail without wrapping.

### Authorization and recovery boundaries

- The target restriction is **operation-specific subject state**, not a tenant/network/session restriction.
- `LockoutPolicy` admits only credential-mutation operations and explicitly excludes factor-proof/admin recovery routes. Authentication/password recovery cannot silently become targets.
- Start/end administrative operations must not be confused with the target operation’s risk profile or introduced as aliases in its closed operation matrix.
- End must remain callable by properly authorized administration while its target is locked. Reusing target-current-admission as a prerequisite for the administrative end could prevent recovery.
- Apply must compare the exact expected version under the same native locks used by admission, validate bounded primary-time duration, and return attributable before/after versions.
- Fresh mutation authorization and historical recovery authorization are distinct. Historical committed replay must not require renewing preparation TTL or replaying current mutation checks.
- Existing issuance-only `Executor` and Repository profiles must retain their supported behavior; enabling lockout mutation must be explicit and reject missing policy/authority dependencies.

### Canonical effects and consumers

Security catalog specifies:

- `identity.risk.start_lockout`
- `identity.risk.end_lockout`
- retention class `protected`

Lifecycle catalog specifies:

- `risk.lockout_started.v1`
- `risk.lockout_ended.v1`

Scheduled expiry uses `identity.risk.end_lockout`, succeeded/completed. Admission itself currently only observes natural expiry; that read must not claim or emit a durable transition.

Public source callers of `LockoutPolicy` are presently tests; native mutation/executor callers do not exist. Existing protected feature callers use `PrepareCommandCounters` and the shared Coordinator. Preserve issuance, risk-evidence use, protected counters, original-response recovery, and counter-only development profiles as distinct variants. Valkey is not the lockout owner and cannot clear or supersede this state.

### Bounded proving contract

The next implementation should prove authorized start and end, stale-version rejection, one winner under contention, retained-root rotation continuity, exclusive primary-time expiry, original-result replay, changed-binding conflict, rollback without state/effect, and lost-acknowledgement recovery without duplicate transition/audit. Include a protected feature denied by start and subsequently permitted by authorized end.

This establishes native lockout mutation—not canonical risk Service, usable login, or full risk-unit completion.

## Native QA planner

Principal: `/root/lockout_native_qa_4398`

## Bounded QA input

**Role:** read-only QA planner using `$qa-test-planner`
**Repository:** `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`
**Reviewed base/head:** `4398e2dde31d82dac7005bcbc96610aba0645dcb`
**Disposition:** no contradictory requirement or reproducible design defect blocks implementation. No writes or tests executed.

### Observable contract

Native start/end commands must preserve the existing pure `risk.LockoutPolicy` intent, require current administrative authorization, mutate only the exact tenant/operation/keyed subject root with CAS, and commit restriction state, canonical audit/lifecycle effects, and command result through the existing Coordinator.

Preparation is not permission. Lockouts are operation-specific restrictions, not risk decisions, account-wide bans, or recovery overrides. Expiry uses primary database time; retained key rotation and counter-window reset cannot bypass restrictions. End must preserve CAS history rather than silently reset first-start authority.

Current `lockout_admission_integration_test.go` deliberately seeds restrictions through SQL. It proves admission behavior and fencing, **not native start/end authorization, audit, idempotency, or recovery**.

### Smallest meaningful additions

| Partition | Observable assertion | Realistic defect detected |
|---|---|---|
| Authorized first start → protected operation → authorized end | Start commits one restriction/version and one matching canonical event; protected mutation rolls back while active; exact-version end commits and subsequent protected mutation succeeds without deleting CAS history. | Commands mutate the wrong root, admission ignores native writes, end remains ineffective, or deletion permits stale restart. |
| Existing-version start/end; stale, missing and overflow boundaries | Exact expected version advances once according to the selected transition contract. Missing/stale versions fail without state/effect mutation. Existing maximum-version behavior fails closed rather than wrapping. No unselected duration/extension semantics are inferred. | Unconditional upsert, version reset, uint overflow, stale end shortening a newer restriction. |
| Untrusted/mismatched scope and authorization loss | Actor/tenant/target/operation mismatch rejects; an actor denied or revoked after preparation cannot commit. Assert unchanged restriction plus no successful audit/lifecycle effect. | Pure preparation becomes authority, confused deputy, target/actor substitution, stale authorization. |
| Matching replay versus changed intent | Same command/fingerprint replays the original committed result with unchanged version/timestamps/effect counts. Changed mode, duration, expected version, actor or target under the same command cannot mutate. | Retry extends duration, duplicates effects, or command-ID reuse changes intent. |
| Concurrent CAS and admission fence | Two independent connections using one expected version yield exactly one committed transition. Race native start against an already admitted protected transaction; ordering is serialized and the next protected command observes the committed restriction. | Lost update, double winner, lock-order reversal, restriction appearing between admission and domain commit. |
| Proven rollback and commit ambiguity | Inject an effect failure before commit: no restriction transition or success effect survives. Lose commit acknowledgement: return unknown without reapplying; authorized primary recovery resolves the same command and exposes one committed transition/event. | Private transaction escape, split state/effects, unproved retry, duplicate transition following disconnect. |
| Primary-time and retained-key continuity | At the exclusive expiry boundary the restriction no longer denies; before it does. Native start followed by retained-key rotation/window reset still denies. End under a rotated key resolves the same historical logical root. | Application-clock expiry, inclusive-boundary error, rotation creating a fresh unrestricted root. |
| Populated forward migration | Install previous immutable history, seed active/expired restrictions, roots/aliases and existing commands/evidence, then apply new history twice. Compare original rows/intervals/versions and recover/decrypt historical records; execute native commands afterward. | Changed historical migration bytes, silent allowance reset, lost linkage, migration that works only on an empty database. |

### Test boundary and scope reduction

Use **real PostgreSQL plus the real Coordinator and native risk participant** for the rows above. SQL is appropriate for migration fixtures, persisted-state inspection, and hostile legacy rows; SQL-written restrictions alone do not prove mutation ownership.

Reuse existing admission coverage for foreign tenant/operation, older policy, missing schema, future-start fail-closed behavior and existing lock fencing. Extend its fixture with native commands instead of reproducing the entire matrix.

Keep malformed duration/mode/recovery-operation and exact-policy custody partitions at the existing core/unit boundary. Verify native decoding rejects forged projections, but do not recreate all core parsing combinations in PostgreSQL.

Canonical events must use `identity.risk.start_lockout` / `identity.risk.end_lockout`, protected retention, matching command/correlation, attributable actor/subject/policy, and genuine mutation changes—not `BuildRiskDecision`'s non-mutating `identity.risk.decide` record. Assert redacted durable payloads and replay effect counts.

### Execution input

Observe one focused native regression fail before production implementation; then run the focused native lockout tests, existing admission/counter fixture and populated-upgrade fixture. Existing owning entry points are:

- `TestIssuedEvidenceSameWorkIntegration`
- `TestRiskReservationPopulatedUpgrade`

They require the explicitly disposable loopback `IDENTITY_RISK_POSTGRES_TEST_DSN` profile. Every Go invocation needs a task-owned disposable `GOCACHE` with failure-path cleanup.

This artifact identifies focused implementation evidence only; it does not certify aggregate, release, race, mutation, interoperability or clean-consumer gates.

## Native regression reviewer

Principal: `/root/lockout_state_regression_4398`

**APPROVE — bounded populated-storage upgrade.**

Role: read-only `regression_reviewer`. Reviewed repository `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`, branch `feature/identity-platform`, base/HEAD `4398e2dde31d82dac7005bcbc96610aba0645dcb`, plus the six-file dirty batch. No writes or test execution performed.

No material regression findings:
- Migration history equality protects installed migration bytes; invalid allocation, cancelled/nil context and returned-history ownership are exercised.
- The real PostgreSQL fixture snapshots every original lockout field across populated active/expired restrictions, checks null legacy custody and unchanged command count, and verifies rejected updates leave retained intervals unchanged.
- Complete same-tenant start/end linkage succeeds; partial custody, noncanonical identifier, nonexistent command and operation/state mismatch fail.
- Entry is reachable through `TestIssuedEvidenceSameWorkIntegration` → protected counters → durable-lockout admission → populated-state upgrade.
- Existing admission selects named original columns; the additive schema does not change its interval-only behavior.

Persisted evidence records `go test -race -tags=integration . -count=1`, exit 0, **12.406s**, with PostgreSQL 18.6. Current hashes of all four changed Go files match its input map:
`85b1aac0…`, `390ab947…`, `e1249024…`, `701a00d6…`.

Bounded untested partitions: actual existing foreign-tenant command linkage, duplicate native record IDs, invalid end-time bounds, and legacy admission explicitly repeated after this upgrade. These are additional confidence opportunities, not observed defects blocking this storage-only batch.

This approval does **not** establish native authorized start/end/CAS, lifecycle/audit/recovery, ended-state admission, full-unit completion, or release-gate completion.

## Native Go compatibility reviewer

Principal: `/root/lockout_state_compat_4398`

**APPROVE — Go compatibility audit**

- Role: `go_compatibility`; head/base `4398e2dde31d82dac7005bcbc96610aba0645dcb` plus the six-file dirty batch.
- No concrete compatibility or populated-upgrade defect found. The new source is additive, preserves prior migration history, retains legacy intervals/versions/root keys, and introduces no dependency or Go-version change.
- Existing admission reads named legacy columns; its interval-only behavior remains unchanged and explicitly documented.
- Reviewed integration evidence records passing race-enabled PostgreSQL tests on Go 1.27.1; the four changed Go-file hashes match current files. No tests or services executed by this reviewer.

**Pending:** native authorized start/end/CAS, ended-state admission, lifecycle/recovery, published clean-consumer compatibility, and complete unit/release gates. This approval covers the forward-storage batch, not full-unit completion.

## Initial native correctness review — rejected tree

Principal: `/root/lockout_state_final_4398`

**REQUEST_CHANGES** — local intermediate storage batch.

Principal: `/root/lockout_state_final_4398`. Base/head: `4398e2dde31d82dac7005bcbc96610aba0645dcb` plus complete six-file dirty patch, including untracked files. Single read-only reviewer; no writes, executions, services or child agents.

**Blocker:** `risk/postgres/lockout_state_migrations.go:57` permits `record_id='AAAAAAAAAAAAAAAAAAAAAA'`, the all-zero ID explicitly rejected by `identity.ParseSubjectID()`. With an existing same-tenant start command and 32-byte digest, every other constraint accepts this state, leaving purported canonical native linkage that the identity API cannot reconstruct. Exclude the zero spelling and add a real-PostgreSQL rejection regression with valid command linkage.

Falsification: traced all transition predicates and the composite FK; neither validates record-ID entropy. Existing negative fixtures use missing/foreign commands or partial custody, so their failure cannot prove rejection of a fully linked zero record ID.

Remaining scoped paths inspected: forward history preservation, nullable legacy state, finite end times, start/end consistency, command tenant/operation custody, uniqueness, existing interval-only admission and populated-upgrade oracles. No additional findings.

Supplied evidence: `go test -race -tags=integration . -count=1` passed, 12.406s; `go vet -tags=integration .` passed. Their four affected Go-file hashes match this reviewed snapshot. No independent execution performed.

Reviewed SHA-256:

```text
lockout_state_migrations.go              85b1aac0e20859e6a2c788571861bca7de6a72ea25c9020f56bf5a1cc7c8878e
lockout_state_migrations_test.go         390ab9479edb0e9ea0b72d28534415c08dd20bd131ee97cb367ee5616173e080
lockout_state_integration_test.go        e124902421a22bf8b0876a6178eb63cb861acb35282221406da05141dde00aee
lockout_admission_integration_test.go    701a00d6e64eb22ef8d7047b3756f4f14a37520536773e48d88b1ca404b78562
README.md                               6f2d1e85551bfcbe214a5a3708025bf8483f354862601efca4e87082d3c8a33c
CHANGELOG.md                            94f53de38b2b470654a8a2f86530a3456098934ece89a1c9bf29f9fe7d9a5be9
```

Native authorized transitions, CAS/audit/recovery, ended-state admission, clean consumer/reference composition and full unit/release gates remain pending; this review does not certify those boundaries or full-goal completion.

### Primary reproduction (not independent closure)

Added valid existing same-tenant start linkage with the zero record spelling.
Real PG focused regression FAILED at `lockout_state_integration_test.go`:
`zero subject ID accepted as native aggregate identity` (1.031s total).
The subsequent schema correction excludes that exact zero spelling; older
race/vet and review reports above remain bound to their original tree.

## Final native regression reviewer

Principal: `/root/lockout_state_regression_final_4398`

**APPROVE — bounded populated-schema upgrade coverage.**

Identity: `regression_reviewer`, `/root/lockout_state_regression_final_4398`. Reviewed base/head `4398e2dde31d82dac7005bcbc96610aba0645dcb` plus the complete six-file dirty batch.

No concrete regression defects found. Tests cover immutable installed history, caller-owned history, invalid allocation/context, populated active/expired state preservation, absent fabricated custody, rejected partial/inconsistent linkage, and valid native start/end storage without interval rewriting. The zero-ID regression now uses a valid same-tenant start command, isolating the schema/API mismatch.

All four affected Go-file hashes match `risk-lockout-state-final-race.json`, recording `go test -race -tags=integration . -count=1` PASS (`13.120s`). I inspected evidence; executed no tests or services.

Approval covers storage readiness only. Native authorization, start/end/CAS execution, audit/lifecycle/recovery, ended-state admission, and release gates remain unfinished; this is not package-completion approval.

## Amended native Go compatibility reviewer

Principal: `/root/lockout_state_compat_4398`

**APPROVE — amended `go_compatibility` disposition**

Reviewed the complete current six-file batch on `feature/identity-platform`, base/head `4398e2dde31d82dac7005bcbc96610aba0645dcb` plus dirty changes. No unresolved compatibility findings.

The zero-ID exclusion now matches the existing nonzero `SubjectID` contract. Legacy rows remain nullable and unaffected; previous migration bytes, interval-only admission, module paths, dependencies and Go 1.27 target remain unchanged.

Current SHA-256 hashes:

```text
lockout_state_migrations.go          31779afa82493d0ea27ff67154e0834f64ca30e614d0876fd016cd6372e087c4
lockout_state_migrations_test.go     390ab9479edb0e9ea0b72d28534415c08dd20bd131ee97cb367ee5616173e080
lockout_state_integration_test.go    69487fd6e38efcd7625d5766fa7e11b522ef1835aec8bb04cdffabaf9cb75447
lockout_admission_integration_test.go 701a00d6e64eb22ef8d7047b3756f4f14a37520536773e48d88b1ca404b78562
README.md                           6f2d1e85551bfcbe214a5a3708025bf8483f354862601efca4e87082d3c8a33c
CHANGELOG.md                        94f53de38b2b470654a8a2f86530a3456098934ece89a1c9bf29f9fe7d9a5be9
```

Final evidence records `go test -race -tags=integration . -count=1` passing in 13.120s; changed Go-file hashes match. Reviewer executed no tests, services or writes.

Approval remains storage-batch-only. Native transitions/ended-state admission, lifecycle/recovery, published clean-consumer proof and complete release gates remain pending.

## Final native correctness review and finding closure

Principal: `/root/lockout_state_final_4398`

**APPROVE — No findings for the complete six-file intermediate storage batch.**

Principal: `/root/lockout_state_final_4398`. Repository: `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`, branch `feature/identity-platform`. Base/head: `4398e2dde31d82dac7005bcbc96610aba0645dcb` plus the six-file dirty patch, including untracked files.

The zero-ID finding is closed: the schema explicitly excludes `AAAAAAAAAAAAAAAAAAAAAA`. The regression supplies an existing same-tenant start command and complete digest, preventing FK or partial-custody failures from masking acceptance. The subsequent positive linkage fixture verifies that valid IDs remain accepted.

Fresh complete-diff inspection covered migration history, legacy/populated preservation, ID encoding, end-time consistency, command custody, uniqueness, admission consumers, test oracles and documentation. Existing interval-only admission is intentionally unchanged and explicitly documented.

Retained results match the reviewed Go-file hashes:
- `go test -race -tags=integration . -count=1`: passed, 13.120s.
- `go vet -tags=integration .`: passed.
- `git diff --check`: passed during this review.

Reviewed SHA-256, paths relative to `risk/postgres/`:

```text
lockout_state_migrations.go           31779afa82493d0ea27ff67154e0834f64ca30e614d0876fd016cd6372e087c4
lockout_state_migrations_test.go      390ab9479edb0e9ea0b72d28534415c08dd20bd131ee97cb367ee5616173e080
lockout_state_integration_test.go     69487fd6e38efcd7625d5766fa7e11b522ef1835aec8bb04cdffabaf9cb75447
lockout_admission_integration_test.go 701a00d6e64eb22ef8d7047b3756f4f14a37520536773e48d88b1ca404b78562
README.md                            6f2d1e85551bfcbe214a5a3708025bf8483f354862601efca4e87082d3c8a33c
CHANGELOG.md                         94f53de38b2b470654a8a2f86530a3456098934ece89a1c9bf29f9fe7d9a5be9
```

No writes, tests, builds, services or children were started by this reviewer. Native authorized transitions, CAS/audit/recovery, ended-state admission, clean composition and full unit/release gates remain pending. This approval does not certify those boundaries or full-goal completion.

## Native local delivery audit

**APPROVE — local intermediate storage checkpoint only.**

Principal: `/root/lockout_state_delivery_885b9bf`.

- Source HEAD `885b9bf6a99cd4aac2d9f8089d25e4b27e73d09f`, parent `4398e2dde31d82dac7005bcbc96610aba0645dcb`; six-file commit, clean source checkout.
- Inspected native initial `REQUEST_CHANGES`, final regression, amended compatibility, and complete-six-file correctness `APPROVE`. Final hashes match committed files; zero-ID finding is closed.
- Final race and vet evidence retain original `4398e2d` dirty execution; all **369 inputs match** committed HEAD. Race passed `13.120s`; vet exit `0`. No gates rerun.
- Control HEAD `cf4e3481`; ledger and five evidence files remain **uncommitted**.
- Owned container and named temporary directory are absent. Shared image was not modified by this audit.
- Local commit only; no pushed, hosted-CI, deployed, or production-verified delivery established.

Counts remain **5/67 integrated, 1/67 verified**; goal active. Native authorized transitions/CAS, ended-state admission, audit/lifecycle/recovery, canonical Service, and full release gates remain pending.
