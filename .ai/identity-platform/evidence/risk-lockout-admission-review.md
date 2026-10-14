# Native independent reports

All principals below were isolated, read-only `gpt-6.1-sol`, medium,
`fork_turns="none"`. The selected implementation is an intermediate current
admission/schema boundary; original mutation/service requirements remain open.

## Independent change_mapper report

**Principal:** `/root/risk_service_map_b0e746a`
**Mode:** read-only source inspection; no writes, tests, builds, services or children.
**Identity base/head:** `b0e746ad5fd75f575a605337cf0e21e325a1c069`
**Control head:** `952ec6e97cb57dfb012f3727c7db59d6b1b56efc`

### Missing production boundary

The audited non-issuing Evaluator still cannot provide production risk permission: there is no canonical `Service`, durable lockout owner/current lockout check, or complete decision-command recovery composition.

The smallest meaningful advancement is **native PostgreSQL durable lockout state and current admission enforcement**, followed by canonical Service composition over existing machinery. A Service wrapper over Evaluator alone would not close that boundary.

### Contract and execution map

| Boundary | Existing implementation | Missing behavior |
|---|---|---|
| Policy and source orchestration | `risk/policy.go:163`, `risk/evaluation.go:86`, `:224`; exact catalog, mandatory sources, fixed precedence, bounded freshness | Complete service permission and construction-time durable ownership validation |
| Ephemeral counters | `risk/counter.go`; Evaluator freezes one CounterCoordinator and excludes duplicate velocity Source | Native Valkey adapter absent from assigned worktree; PG journal is not an ephemeral mutation engine |
| Durable velocity recovery | `risk/postgres/velocity_journal.go:45`, `:139`; `velocity_results.go:28` | These existing records must remain the journal owner; do not create a second service replay journal |
| Protected counter ownership | `risk/postgres/counter_admission.go:17`; `postgres/command_counter_admission.go:11` | Existing admission lane explicitly anticipates selected lockout/override checks, but native implementation checks fixed-counter allowance only |
| Evidence issuance/recovery | `risk/issuance.go:516`, `:602`; `risk/postgres/executor.go:120`, `:167` | Canonical unified Service; full phone/CAPTCHA consumption and complete workflows |
| Decision audit | `risk/decision_audit.go:28`; original canonical record/outcome retained | Successful loss handoff is expressly not durable reconciliation proof |
| Current domain authority | Feature owner and sole identity Coordinator | Risk must not replace credential/user/factor/organization checks or authorization |

### Authoritative requirements

- `REFERENCE_CONFIGURATION.md:142`: PostgreSQL owns lockout/journal; Valkey owns ephemeral velocity; Valkey unknown is unavailable.
- `REFERENCE_CONFIGURATION.md:144`: authentication throttles before credential/provider work; protected phases reserve under their owning command.
- `goals/identity-risk.md:117–131`: rotation continuity, account-lockout DoS resistance, safe recovery, no enumeration, sole mutation ownership.
- `fragments/public_contracts_risk_delivery.json:8054–8116`: canonical Service construction rejects overlapping authority/missing journal; `Evaluate` and `Recover` are its operations.
- `fragments/public_contracts_risk_delivery.json:2706–2861`: canonical Evaluate binds command/idempotency, trusted attempt/action/scope and issuance phase; replay returns the original outcome without provider reevaluation.
- `goals/identity-risk-postgres.md:84–91`, `:156–161`: durable lockout/CAS overrides, database time, retention and rotation continuity.

### Smallest coherent next step

1. Add explicit bounded immutable lockout policy/state contracts, keeping lockout authority distinct from ephemeral velocity.
2. Implement native PostgreSQL versioned lockout transitions and same-command recovery through the **existing identity Coordinator**, canonical audit and lifecycle participants.
3. Enforce current lockout/override state through the existing native current-admission lane, checking before domain work and before commitment without recharging counters or renewing evidence.
4. Preserve lockouts across retained-key rotation and Valkey loss/reset. Deny unavailable current authority.
5. Compose canonical Service afterward from these owners, Evaluator, existing counter journal, audit recorder and evidence executor; preserve the supported development Evaluator API.

**Important constraint:** the selected reference configuration supplies velocity limits, not automatic lockout durations/triggers. Do not infer account lockout from the authentication limit or permit identifier-only unauthenticated attempts to disable known accounts. Safe recovery and narrow authorized overrides need explicit policy semantics.

### Callers and variants

Current source consumers are tests/examples and PhoneEvidenceIssuer/native adapter seams; no production password/login Service caller exists in this worktree. Intentionally different variants must remain distinct:

- source-only versus counted Evaluator;
- unaudited versus required-audit Evaluator;
- non-issuing assessment versus phone evidence issuance;
- ephemeral velocity versus command-owned PostgreSQL counters;
- historical evidence recovery versus current permission.

`risk/valkey` is absent in the assigned identity checkout. Its goals/catalog and PG recovery interfaces were inspected, but native Valkey behavior was not available to inspect.

### Observable acceptance boundary

A useful next checkpoint proves: a durable active lockout denies despite clean/reset ephemeral state; concurrent mutation/current checks cannot commit past a newly applicable lockout; expiry uses primary time; rotation preserves continuity; unknown commits recover the same transition without duplicate state/audit; recovery remains available without account enumeration. This checkpoint advances login prerequisites but does **not** establish usable login or full risk-unit completion.

## Bounded QA artifact

**Principal:** `/root/risk_service_qa_b0e746a`
**Role:** fresh independent read-only QA planner; `$qa-test-planner` read completely.
**Source base/head:** `b0e746ad5fd75f575a605337cf0e21e325a1c069`.
**Control inspected:** live head `4c2d60cdfa29f35f6d3393e3ebcc4586c25de931`, differing from supplied `952ec6e9…`; parent notified.
**Disposition:** implementation can proceed. No blocking contradiction identified. No writes, tests, builds, services or children.

### Contract and present boundary

The next advancement must establish actual current durable-lockout enforcement and non-overlapping state ownership through the canonical risk service—not rename the existing evaluator as complete.

Canonical `Service` owns `Evaluate` and recovery; PostgreSQL owns durable lockouts/journal, Valkey only ephemeral windows. Invalid bindings reject before state/provider access. Lockout loss, stale snapshots, ambiguous writes and Valkey reset cannot produce allowance. Phone issuance conserves exact phase/purpose/binding and native command recovery.

Current source explicitly lacks lockout/override policy and canonical Service. Existing native protected-counter admission checks current windows but says it grants neither domain authorization nor a complete risk verdict. Native PostgreSQL velocity journal/recovery exists; no `risk/valkey` source was found in this checkout. Lease/replay acquisition is not healthy ephemeral authority.

### Smallest meaningful partitions

| Concrete risk | Observable assertion | Proving boundary |
|---|---|---|
| Clean signals bypass durable lockout | Persist active lockout, supply otherwise clean facts: public evaluation denies, returns no evidence, and protected mutation remains unchanged. Unlocked control permits the configured result. | Public service + real PG |
| Old positive assessment bypasses newly imposed lockout | Evaluate before lockout; impose lockout before protected apply/issuance. Current authoritative check prevents the effect, without refreshing original signal/evidence times. | Real shared-coordinator transaction |
| Expiry uses process clock | Test immediately before, exactly at and after selected lockout expiry; primary database time determines state. A skewed caller clock cannot shorten it. | Real PG |
| Account-existence/DoS leakage | Equivalent known/unknown-identifier attempts expose the same safe response/retry shape and do not create an unconditional account-wide lockout from unauthenticated hints. Configured safe recovery remains available without bypassing its own proof requirements. | Public service/workflow boundary |
| Duplicate state owners | Missing durable owner, competing durable owners, durable state assigned to ephemeral owner, policy/version mismatch and invalid owner custody fail construction; valid disjoint ownership works. No callback/IO executes during rejection. | Constructor tests; runtime assertions for valid composition |
| Rotation resets restriction | Create restriction under retained v1, evaluate/mutate under v2+v1: restriction and investigation linkage survive. Missing required historical custody fails closed, never creates a fresh allowance. Tenant/operation/dimension changes cannot reuse the state. | Public API + real PG |
| Retry double-mutates or ambiguous write allows | Exact command replay returns original outcome without another transition; changed binding conflicts. Proven noncommit leaves state unchanged; commit persists post-state; lost acknowledgement returns unavailable/unknown and recovery reconciles the original command. | Native PG commit/abort/ambiguity fixtures |
| Concurrent start/end/override loses restriction | Concurrent commands have attributable outcomes, one authorized transition winner where required, and no lost restriction. Stale expected version/generation cannot clear or shorten a newer restriction. | Real PG contention + race |
| Override becomes generic bypass | Wrong role, missing reason, expired scope, foreign tenant/subject/operation or stale expected version cannot override. Valid selected override is bounded, audited and cannot widen unrelated authority. | Public service + PG |
| Canonical service bypasses admitted evaluation | Invalid canonical action/phase/binding denies before source/state access. Non-issuing result contains only safe decision metadata; phone phases preserve separate purpose/binding and return no reference on denied/failed/unknown. Recovery never reevaluates providers or counters. | Public service; existing issuer/native PG suites |
| Audit delivery misreports transition | Lockout start/end/override/reconciliation produces the canonical action and truthful outcome. Failed/unknown writes never claim successful transition; errors/events contain no raw network, identifier, token or digest. | Persisted state + captured canonical audit record |
| Ephemeral loss clears durable protection | Missing/reset epoch or unavailable ephemeral owner cannot clear/shorten the PG lockout; credential/session issuance fails closed. Acquire/replay lease alone cannot qualify as healthy. | Service + PG; real Valkey when implemented |

**Policy uncertainty, not a blocker:** the inspected reference matrix defines velocity thresholds/windows but no numeric lockout duration/algorithm. Do not infer lockout duration from those values. Tests should exercise explicitly selected bounded policy and document the decision, rather than invent reference defaults.

**Catalog reconciliation:** generated typed shapes differ from existing public APIs (for example Decision representation and request fields). The complete goal explicitly requires preserving supported APIs and reconciling stale generated shapes. Verify canonical semantics and clean typed consumer composition; mechanically replacing existing APIs is not required by this QA plan.

### Smallest sufficient verification set

1. **Focused RED/GREEN:** public evaluation of an active durable lockout with otherwise-clean mandatory signals; independently add stale-assessment/new-lockout and retained-key rotation regressions.
2. **Core focused tests:** ownership rejection, binding admission, lockout/failure/override partitions, cancellation, safe result/error/audit projection; affected module under race detection.
3. **Real PostgreSQL:** active/expired state, exact boundary, contention, CAS denial, replay/mismatch, commit ambiguity and same-command recovery; assert durable state and protected effect, not just callback order.
4. **Direct consumer composition:** canonical Service evaluates both non-issuing and supported phone paths without exposing native persistence records or requiring consumer-written domain stores.
5. **Affected repository gates:** changed module/direct-consumer gates after final changes. Exact coverage/mutation and remaining provider/Valkey/reference release gates remain explicit; they are not new pre-write approval requirements.

### Existing proof to reuse

Reuse unchanged evaluator tests for canonical admission, mandatory-signal precedence, original freshness, cancellation and provider redaction; counted-evaluation tests for exact phases/post-state and final retry-after; decision-audit tests for acknowledgement/loss attribution; native issuance/use tests for exact replay, purpose separation and command ambiguity; protected-counter integration tests for current-window and retained-alias admission; velocity journal/recovery tests for epoch/lease fencing.

Reuse execution evidence only when its complete input identity is proven unchanged. These tests do **not** already prove durable lockout enforcement, complete canonical Service, healthy Valkey recovery or full-unit completion.

## Native regression audit — APPROVE

Role: `regression_reviewer`; read-only inspection. Reviewed base/head `b0e746ad5fd75f575a605337cf0e21e325a1c069` plus all nine dirty `risk/postgres` paths.

No concrete regression found in the bounded current-lockout admission contract.

- Behavioral red genuinely detected active restriction bypass before implementation.
- Native tests assert domain-effect denial and rollback, not merely returned errors. Controls cover expiry, foreign scope, older policy, retained-key rotation, window reset, and missing schema.
- Separate-connection tests exercise insertion fencing through the immediate foreign key and existing-row update fencing through admission’s share lock; subsequent commands detect newly committed restrictions.
- Existing counter-only sibling behavior remains exercised by the complete adapter integration suite.
- Current source/test hashes match both final artifacts: `go test -race -tags=integration . -count=1` passed in **12.283s**; `go vet -tags=integration .` exited **0**. No commands were rerun.

Claim boundaries: exact exclusive expiry is directly tested at the predicate boundary; native tests distinguish active versus expired primary-time intervals. Authorized mutation, overrides, safe-recovery policy, canonical Service, full-unit release gates, and repaired/published dependency composition remain unproven and explicitly pending. This approval applies only to the selected admission/schema batch.

Principal: `/root/risk_lockout_regression_b0e746a`.

## Go compatibility

**Role:** independent Go compatibility auditor
**Disposition:** APPROVE — bounded development batch only.

**Reviewed:** base/head `b0e746ad5fd75f575a605337cf0e21e325a1c069` plus all nine dirty paths; control `4c2d60cdfa29f35f6d3393e3ebcc4586c25de931`.

No material compatibility finding:
- `EnforceLockouts` preserves counter-only behavior when false and rejects selection without native counters.
- `LockoutMigrationSource` matches the consumed migrations/v3 API, appends forward-only atomic history, and preserves existing migration bytes.
- Admission uses the existing native contributor/Work contracts. Coordinator checks occur before domain effects and again before commit; recovery remains prerequisite evidence, not lockout mutation or domain-commit proof.
- Logical keyed roots preserve selected restrictions across retained-key rotation and window reset. No new dependency, permanent replacement, ownership authority, or adapter cycle is introduced.

**Evidence inspected:** final `go test -race -tags=integration . -count=1` and `go vet -tags=integration .` both exit 0 on Go 1.27.1. Recorded risk/core and risk/postgres input hashes match current files.

**Boundary:** these results use exact b0 identity/root and postgres snapshots, authentication `3467aab01d0c`, capability `2985d8ea8700`, and the other pinned development dependencies. The unchanged historical `go.mod` postgres `b22` pin remains previously compile-incompatible; this batch does **not** repair or prove a published clean consumer.

Authorized lockout mutation, overrides, safe recovery-route policy, full risk-service composition, API/release gates, and unit completion remain unproven. Adding a public Config field also does not preserve external unkeyed struct literals; inspected consumers use keyed construction.

Principal: `/root/risk_lockout_compat_b0e746a`.

## Complete final-diff correctness review

Internal disposition: **APPROVE** for the bounded intermediate lockout-admission batch. No findings.

- Reviewed complete dirty diff and all four new files on `feature/identity-platform`, base/head `b0e746ad5fd75f575a605337cf0e21e325a1c069`; authoritative control revision `4c2d60cdfa29f35f6d3393e3ebcc4586c25de931`.
- Single-reviewer structured pass, read-only; no children, writes, tests, services or external mutations.
- Traced keyed-root continuity, tenant/operation/dimension isolation, immutable intervals, missing/malformed authority, cursor cleanup, cancellation, coordinator pre-mutation/pre-commit checks, additive migration and counter-only compatibility.
- Falsification: attempted concurrent insertion into an absent restriction and reactivation of an expired restriction. Counter `FOR UPDATE` plus immediate FK fencing and restriction `FOR SHARE` preserve admission through commit/abort; supplied native integration evidence exercises both competing writes and subsequent denial.
- Supplied evidence: active regression failed before admission enforcement; expanded lockout integration, complete adapter integration race suite and integration vet passed. All 84 recorded final-race source inputs match current files. I did not rerun commands.

Limits: approval does not certify complete risk service, authorized lockout mutations/overrides, lifecycle/recovery, release gates, production readiness, or historical adapter dependency pins. Evidence applies only to the explicit current root/postgres plus authentication `3467aab01d0c` and capability `2985d8ea8700` composition.

Principal: `/root/risk_lockout_correctness_b0e746a`.

## Delivery audit

**APPROVE — bounded local development checkpoint only.**

- Identity is clean at `bc9281c98b28d481f786c170ebaa718ad242f24d`; its complete nine-file commit matches the native reviewed admission/schema boundary.
- All **84** recorded inputs match current files for expanded integration GREEN, final integration/race PASS (**12.283s**) and integration vet exit **0**. Execution remains attributed to `b0e746a` plus dirty inputs; the identical committed inputs support reuse without a new-SHA rerun.
- Native reports inspected directly: mapper `/root/risk_service_map_b0e746a`; QA `/root/risk_service_qa_b0e746a` (proceed); regression `/root/risk_lockout_regression_b0e746a`, compatibility `/root/risk_lockout_compat_b0e746a`, correctness `/root/risk_lockout_correctness_b0e746a` (**APPROVE** each). Post-implementation reviews cover `b0e746a` plus all nine dirty paths, now committed unchanged. Mapper’s older control head and QA’s corrected control head are explicitly retained.
- Control remains at `4c2d60c` with exactly the four specified dirty paths; those evidence/ledger changes are **not committed yet**. The goal permits local commits and prohibits push/publication/deployment.
- Scratch directory is absent; Docker confirms the named owned container is absent.

**Delivery states:** local implemented; source committed/control pending; pushed not evidenced and prohibited; CI not verified; deployed not evidenced; production not verified.

Authorized mutations/overrides, canonical Service, full release/coverage/mutation gates and historical `b22` dependency repair remain unproven. Counts remain **5/67 integrated, 1/67 verified**; neither unit nor platform completion is established.

Principal: `/root/risk_lockout_delivery_bc9281c`.
