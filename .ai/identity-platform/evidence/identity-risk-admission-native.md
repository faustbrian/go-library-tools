# Native command-admission development checkpoint

Base f2d177ba117e2c963e7460596b57c7a01075eda6; committed locally
f87c5148578c2fcb111fc71a418a750a72fc2afa. Complete eight-file tree
87d0aba87cdec81cb2af63419fda63ccca5c545c. No push or release.
All independent roles used default gpt-6.1-sol, medium, fork_turns none,
read-only. Parent owns code and execution; no implementation worker.

## Results

- Missing admission public API compile RED observed before production changes.
- Focused admission/query/feature observation GREEN0.223s.
- Unit `go test -race -count=2 ./...`: PASS1.708s.
- Final PostgreSQL18.6 `go test -tags=integration -race -count=2
  -coverprofile=<task>/integration.coverage ./...`: PASS8.259s86.5%.
- Final `go vet -tags=integration ./...`: PASS.
- Staged whitespace: PASS; review/commit tree equality confirmed.
- Go1.27.1 darwin/arm64; fourteen exact declared owned dependency Git archives
  selected through a disposable external workspace, not current main. No
  permanent replace or dependency manifest change.
- PostgreSQL test distribution: official Homebrew18.6 bottle SHA256
  31b40a10b68e4e69350a26f7033ea5c68c3360512ae61393ba1d801d7770c480;
  task-local loader/path relocation and ad-hoc signing only. No shared package
  installation or Docker service start. Disposable fixture databases remaining0.
- Exact coverage/mutation, native risk issuer composition and unit/release
  acceptance remain pending. Metadata observation is not durable issuance.

## change_mapper native report — /root/risk_admission_map

**Role:** read-only `change_mapper`
**Workspace:** `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`
**Reviewed base/head:** `f2d177ba117e2c963e7460596b57c7a01075eda6`
**Disposition:** mapping complete; no edits or execution gates run.

### Observable contract and paths

- **Before provider evaluation:** `PhoneEvidenceIssuer.Evaluate` constructs immutable `EvidenceIntent`, invokes `EvidenceExecutor.Resolve`, and evaluates providers/counters only for `EvidenceMissing`. Unknown, denied, failed, or committed outcomes do not evaluate again. `Recover` only resolves, including for missing commands. References: `risk/issuance.go:143`, `:179`, `:387`, `:510`, `:524`, `:535`, `:584`.
- **Intent identity:** retained evidence-key versions each produce an `EvidenceIntentBinding`; the digest binds command, actor, tenant, subject, recovery operation, canonical phone/pre-auth/attempt, policy/phase/purpose, subject dimensions, and startup configuration. No provisional issue artifact exists at this point. References: `risk/issuance.go:165`, `:179`, `:671`.
- **Existing metadata lookup:** `QueryCommand → queryCommandReceipt → authorizeCommandQuery + commandRead → commandReceipt`. Query policy is opt-in, current, cloned/bounded, revision-checked, actor/tenant/action/resource-bound, and performed before SQL without a transaction. A valid denied request still performs the same primary read. References: `postgres/query.go:119`, `:155`, `:175`, `:273`; `postgres/coordinator.go:76`.
- **Enumeration fence:** missing rows, retired claims, wrong tenant/purpose, and wrong historical binding return zero plus `ForbiddenError`; no ordinary caller can distinguish absence. Pending remains `InProgress`, even after lease expiry. Incomplete/corrupt matching journals return unavailable. References: `postgres/query.go:193`, `:197`, `:204`, `:207`, `:223`.
- **Recovery is different:** `RecoverCommand` first authorizes/observes, then may reconcile pending commands. Private recovery locks claims/journals, can advance generation, invoke frozen participant recovery, persist failure audit, and abort with unanimous noncommit proof. It is therefore not an absence-only admission primitive. References: `postgres/recovery.go:16`, `:30`, `:48`, `:83`, `:128`.
- **Mutation/replay authority:** `ReserveCommand` owns reservation. `lockCommandClaim` compares immutable scope and recreates canonical intent using the stored fingerprint-key version; retired, mismatched, missing-key, or incomplete claims fail closed. References: `postgres/coordinator.go:200`, `:218`; `postgres/claims.go:74`, `:82`, `:88`, `:91`, `:97`, `:103`.

### Public APIs and intentionally different variants

- `NewFeatureCommand` freezes typed bounded fields; it conveys neither authorization nor execution readiness. Feature fields are keyed individually by the coordinator codec, and complete command intent contributes to the command fingerprint. References: `feature_command.go:59`, `:77`, `:85`, `:115`; `postgres/codec_feature.go:9`.
- `PrepareCommandPlan` is pure preparation, not observation or authority. Native feature preparation validates the registered owner/schema and complete typed command before reservation. References: `postgres/coordinator.go:124`; `postgres/plan_feature.go:103`.
- `CommandRecord` deliberately exposes only safe historical binding/outcome/retry and optional feature descriptor, not replay/native payloads or fingerprints. Reference: `postgres/query.go:70`.
- `CreationUnitOfWork.Recover` is a trusted bridge behind core authorization/redaction, restricted to its admitted kernel operations; it cannot serve the risk feature executor. Reference: `postgres/creation_owner.go:226`.
- No non-test production callers of public `QueryCommand` or `RecoverCommand` exist in this checkout beyond their definitions. Native risk/postgres executor implementation is not present here.

### Genuine gaps and smallest causal scope

1. **Authorized absence observation is absent.** `QueryCommand` intentionally collapses `pgx.ErrNoRows` into forbidden. A trusted native adapter currently cannot prove genuine command-ID absence through a public root-owner API without violating that contract or querying root-private tables.
2. **Root read does not include retained fingerprint-key version.** `commandRead` fetches the fingerprint but not `fingerprint_key_version`; direct exact-command replay matching cannot be performed from its current observation alone. Reference: `postgres/query.go:108`.
3. **Pre-provider intent differs from the full issuance command.** Resolve receives pre-evaluation `EvidenceIntent`; subsequent issuance adds proof digest, nonce, signature, and preparation/expiry times. It cannot reconstruct the complete issued command fingerprint before providers run. References: `risk/issuance.go:233`, `:246`, `:623`. An absence observation must not be represented as complete issuance-command matching.
4. **Retained intent must still be checked by its authority owner.** Existing `CommandRecord`/`RecoveryQuery` cannot compare the risk intent digest. Matching historical scope alone is insufficient to replay evidence for changed phone/pre-auth/attempt/configuration. The native risk adapter’s authoritative retained intent comparison must remain mandatory before returning a recovered result.

**Smallest causal boundary:** an additive root-coordinator trusted observation seam, reusing current bounded actor/resource authorization and primary command-ID observation, while preserving `QueryCommand`’s exact enumeration contract. Only actual primary `ErrNoRows` after current authorization can represent absence; retired/wrong-scope/incomplete/corrupt/read-failed observations must not. The risk executor retains ownership of evidence-intent matching and encrypted native-response recovery through its own tables. Absence supplies no reservation, lease, generation, mutation, or noncommit proof.

Existing integration assertions explicitly protect equal decision/read shape and durable nonmutation for missing/wrong-scope/denied `QueryCommand`: `postgres/query_integration_test.go:83`, `:104`. Those are affected compatibility consumers even if ordinary lookup’s external behavior remains unchanged.

## qa_test_planner native report — /root/risk_admission_qa

**Role:** `qa_test_planner`
**Reviewed base/head:** `f2d177ba117e2c963e7460596b57c7a01075eda6`
**Workspace:** `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`
**Skill:** `$qa-test-planner`, read completely.
**Disposition:** Requirements permit implementation; no contradictory requirement or reproducible design defect identified. Read-only planning; no files changed or tests executed.

### Observable contract

Only an authorized, bounded primary observation proving the command ID absent permits native risk provider/counter evaluation. Denial, wrong scope, intent collision, retired claim, incomplete history, corruption, missing retained keys, cancellation, and dependency failure never constitute absence.

The admission observation must compare complete immutable command intent—not merely `RecoveryQuery` scope—and preserve safe historical outcome/error classification without disclosing command replay payloads. It must not create claims, reserve participants, acquire ownership, renew leases, or invoke contributors. `QueryCommand` must retain its existing missing/wrong-scope/denied indistinguishability.

### Smallest sufficient behavior partitions

| Risk / realistic defect | Observable assertion | Proving boundary |
|---|---|---|
| Forbidden mistaken for absence | Authorized genuinely missing command returns explicit absence; missing with denied or mismatched caller returns zero result plus denial, never absence. `QueryCommand` still returns zero plus Forbidden for missing, wrong scope, and denial. | Public admission unit tests; existing public query tests |
| Historical authorization reused | Independently vary actor ID/kind, tenant, user resource, operation/action, resource type, caller policy revision, decision revision/outcome. Only exact current Allow admits; denied existing and absent commands return no history or absence. Caller-injected command attributes cannot override owned values. | Public boundary with captured authorization inputs |
| Scope-only replay accepts changed intent | Start from one valid feature command. Independently alter immutable feature schema/owner and each accepted command argument while retaining identical safe scope. Exact command recovers history; altered intent cannot return absence or successful historical recovery. Include kernel command intent as the existing codec supports both variants. | Public admission, production command constructors and codec |
| Rotation compares against newest key | Write using fingerprint key v1; observe exact command with current v2 and retained v1. Historical result remains identical. Changed intent remains rejected. Removing v1 fails unavailable, never absence. Include secret-containing kernel intent to detect failure to regenerate canonical secret digests under v1. | Unit plus real PG |
| Missing journal interpreted as missing command | Claim exists but journal missing; retired claim; malformed fingerprint/key version/scope/plan; mismatched plan feature; malformed or inconsistent terminal result/class. Every case fails closed without absence or payload. | Unit corruption seam; real PG valid stored bytes |
| Replay changes historical outcomes | Cover committed, live pending, expired pending, and each existing terminal class: denied, conflict, cancelled, deadline-exceeded, dependency-failure. Preserve binding, feature schema, terminal error, and bounded retry semantics. Expired pending remains pending; no takeover. Historical result remains readable after current user status/composition changes where existing query contract allows it. | Public admission unit table; real PG committed/pending/aborted fixtures |
| Late dependency success bypasses timeout | Already-cancelled caller, policy/read outage, panic, zero clock, policy finishing after deadline, read finishing after deadline. No absence/history escapes after deadline; errors remain typed/redacted. Test an earlier caller deadline. | Unit with deterministic cancellation, no sleep dependence |
| Observation grants mutation authority | Before/after snapshots of claims, journal ownership/generation/lease/result, and any enlisted participant rows are identical for missing, matching, collision, pending, denial, and repeated/concurrent observations. Missing remains row count zero. | Real PG |
| Protected receipt accidentally exposed | Query/admission result and formatting contain only safe metadata; sentinel feature-result/credential payload never appears. Returned feature metadata is independently owned: modifying one result cannot alter the next observation. | Unit and real PG fixture |
| Risk evaluation occurs without proven absence | Resolve denied/unavailable/colliding/pending/committed outcomes cannot advance counter state or obtain provider-derived evidence. Only explicit authorized absence permits evaluation. `Recover` never evaluates, including absence. | Risk issuer consumer tests; real native adapter when available |

Counter/provider instrumentation is supplementary; consumer assertions must also establish zero evidence, stable durable/counter state, and the expected unavailable/denied/historical result.

### Minimal execution set

1. **Red-green:** New public-admission regression showing authorized true absence is distinguishable from denial while existing `QueryCommand` remains indistinguishable. Add the same-scope/different-intent and retained-v1 rotation regressions before the production change.
2. **Unit regression:** New admission table plus existing query, feature-observation, claims/codec, and risk issuance tests. Run affected `postgres` and `risk` modules under race detection.
3. **Real PostgreSQL:** Extend existing production commit/abort/feature fixtures to exercise actual primary bytes, retained-key replay, true missing command, collision, pending/expired pending, denial, concurrency, and unchanged durable snapshots. No need to repeat unrelated domain recovery matrices.
4. **Repository gates:** Run changed module contract and reverse-dependant gates required by repository policy after the final affected changes. Keep exact coverage/mutation evidence separate from focused regression evidence.

Every Go invocation requires task-owned disposable `GOCACHE`, removed after its bounded run. Existing integration tests require an explicitly authorized task-owned loopback PostgreSQL database; missing infrastructure is an unverified boundary, not a pass.

### Scope limits

This artifact does not require a new reservation/recovery mechanism, broad end-to-end suite, or complete PostgreSQL isolation/deadlock/retention matrix for a read-only admission change. Actual evidence issuance, bearer recovery, and protected consumption remain later adapter obligations; metadata admission alone cannot prove those outcomes.

### Parent reconciliation of bounded plan

Mapper proved full future issue intent is unavailable before provider evaluation.
The QA intent/key-rotation assertions belong to native retained-intent replay and
existing ReserveCommand, not this root metadata/absence API. They remain native
adapter obligations, not waived or falsely credited. Core risk inputs did not
change, so unchanged risk runtime evidence was reused rather than rerun. Root
historical metadata intentionally needs neither current registry nor retained
fingerprint-key possession; it grants no native replay or mutation authority.

## regression_reviewer first native report — /root/risk_admission_regression

**Role:** regression_reviewer
**Reviewed:** base/HEAD `f2d177ba117e2c963e7460596b57c7a01075eda6`, complete current dirty admission change, shared query/coordinator paths, and `risk/issuance.go` consumer contract.
**Disposition:** APPROVE. No material findings.

- Admission tests exercise authorized absence, denial/current-policy fences, cancellation, corrupted/existing claims, expired pending history, terminal classifications, and owned feature copies.
- Primary integration tests assert unchanged persisted state for absence, denial, orphan journals, pending/committed history, wrong scope, and concurrent metadata reads.
- `postgres/admission.go:84` returns absence only after authorization and an actual missing global claim. Existing history delegates to the established receipt parser; no native replay or reservation authority is introduced.
- `postgres/query.go:271` preserves ordinary lookup semantics through the shared authorization helper. `risk/issuance.go:524` still requires the future native `Resolve` adapter to match retained intent before reporting Missing.

**Nonblocking coverage gap:** constructor-level isolation with *only* `QueryPolicy` and *only* `AdmissionPolicy` is not directly exercised; unit tests alter private policy fields, while integration enables both (`postgres/coordinator.go:92`, `postgres/admission_test.go:65`).

Execution results supplied by the parent were considered, not rerun. This approves the development regression checkpoint, not complete release gates or the future native risk adapter.

## regression_reviewer current native report — /root/risk_admission_regression

**Role:** regression_reviewer
**Reviewed:** base/HEAD `f2d177ba117e2c963e7460596b57c7a01075eda6`, complete current dirty admission scope, including updated integration coverage.
**Disposition:** APPROVE. No unresolved findings.

The previous coverage gap is resolved: `postgres/admission_integration_test.go:54` now constructs QueryPolicy-only and AdmissionPolicy-only coordinators through public `NewCoordinator`, verifies each enabled/disabled API independently, and checks unchanged primary database state.

Production behavior and original acceptance remain unchanged. The current integration/race rerun is pending; this report does not claim its result or release readiness.

## correctness_reviewer native report — /root/risk_admission_final

**Native role:** `correctness_reviewer`
**Internal disposition:** `APPROVE` — local development checkpoint only.

**Boundary:** base/HEAD `f2d177ba117e2c963e7460596b57c7a01075eda6`; complete eight-file staged tree `87d0aba87cdec81cb2af63419fda63ccca5c545c`. Independently confirmed index/tree equality, no unstaged delta, and clean staged whitespace.

**No findings.**

Source-backed closure:
- `postgres/coordinator.go:90–105`: admission and ordinary query are independently opt-in; policies are validated and copied.
- `postgres/admission.go:67–92`: bounded current authorization precedes one primary read; denied, cancelled, failed, orphaned, corrupt, retired, and mismatched observations cannot become absence.
- `postgres/query.go:110–116` and `postgres/claims_schema.go:8–26`: global command-ID lookup, primary-key uniqueness and scoped journal foreign key preserve true global absence semantics.
- `postgres/query.go:198–268`: admission shares exact historical scope, plan, receipt and terminal-classification fences; expired pending remains pending. The shared authorization extraction preserves ordinary query behavior.
- `postgres/coordinator.go:231–293`, `postgres/claims.go:82–111`, and `risk/issuance.go:389–397,524–539`: metadata observation grants neither reservation nor replay authority. Retained pre-evaluation intent remains adapter-owned; complete immutable command fencing remains in `ReserveCommand()`.
- Reviewed both new test files, constructor isolation, original query regressions, real-primary no-mutation assertions, and concurrent historical observation.

Verification execution is parent-supplied: focused RED/GREEN, unit race, PostgreSQL 18.6 integration/race, and final integration vet passed. I ran only read-only structural checks; created no disposable resources.

Exact coverage/mutation, native risk issuer composition, complete-unit acceptance and release gates remain pending. This approval does **not** establish whole-unit verification, PR readiness, publication or deployment.

Report content is retained; Markdown hard-break trailing spaces were normalized.
