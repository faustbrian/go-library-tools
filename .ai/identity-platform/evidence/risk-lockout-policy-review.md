# Native lockout-policy reports

Scope: intermediate immutable lockout preparation, not durable mutation or full
Service. Existing mapper and bounded QA: risk-lockout-admission-review.md.
Reports below initially reviewed the pre-correction four-path tree. Subsequent
recovery-route edits invalidate their approval for the final tree; retained
history does not supply current review approval.

## Regression

## Native regression audit — APPROVE

**Principal:** `/root/lockout_policy_regression_bc9281c`
**Reviewed:** identity base/head `bc9281c98b28d481f786c170ebaa718ad242f24d`, plus complete dirty `risk/lockout.go`, `risk/lockout_test.go`, `risk/README.md`, `risk/CHANGELOG.md`. Scope: intermediate immutable administrative lockout-intent preparation only.

No concrete regression found.

- Tests exercise immutable projections, exact request/policy ownership, actor/scope/CAS/duration rejection, original exclusive TTL and invalid primary-time boundaries.
- Preparation accepts only operation-specific credential-mutation subject targets; authentication/recovery, identifier hints, privileged operations and generic factor bypass flags reject.
- Existing protected-counter tests cover retained-key rotation, operation isolation, canonical network handling and separation from ephemeral authority.
- Formatting redacts linkage; JSON/text serialization rejects. Documentation accurately separates preparation from authorization and durable commitment.

**Evidence inspected, not rerun:** all 115 recorded inputs match current files for both `go test -race . -count=1` (**PASS, 1.325s**) and `go vet .` (**exit 0**).

**Boundary:** approval does not establish native mutation, authorization/CAS enforcement, audit/recovery, canonical Service, exact coverage/mutation, release readiness or unit/platform completion. Same-coordinator PostgreSQL mutation remains the required next stage.

## Compatibility

Role: `go_compatibility`. Disposition: **APPROVE** for the intermediate immutable preparation slice; no compatibility findings.

Reviewed identity HEAD/base `bc9281c98b28d481f786c170ebaa718ad242f24d` plus exactly four dirty paths: `risk/lockout.go`, `risk/lockout_test.go`, `risk/README.md`, `risk/CHANGELOG.md`; control HEAD `6c6b6a8c0cf47c99aa6064d745979122963bb300`.

- Additive APIs preserve existing counter signatures, implementations, consumers and dependency graph; new imports are standard-library only and compatible with declared Go 1.27.
- Zero policy/command cannot admit use; immutable ownership is retained and returned counter/target slices are independently owned.
- Core performs no IO, authorization or durable mutation. PostgreSQL remains sole durable owner; no reverse import, new coordinator or service authority.
- Preparation excludes authentication/recovery, privileged and broad targets; original command TTL and exact policy/request binding remain enforced.

Current source hashes match the inspected race artifact: `lockout.go` `bfdaaddd…5204ffb`; `lockout_test.go` `b03fdf28…1998a5e7`. Artifact records race PASS with Go 1.27.1, development composition—not published-consumer proof. No commands executed beyond read-only inspection.

Catalog reconciliation, full release gates and existing PostgreSQL historical-pin incompatibility remain outside this approval. No native lockout implementation, unit completion or release-readiness claim.

Principal: `/root/lockout_policy_compat_bc9281c`.

## Initial complete correctness report

## Native correctness review

**Principal:** `/root/lockout_policy_correctness_bc9281c`
**Internal disposition: REQUEST_CHANGES**

**Reviewed boundary:** identity `feature/identity-platform`, base/head `bc9281c98b28d481f786c170ebaa718ad242f24d`, plus complete dirty contents of `risk/lockout.go`, `risk/lockout_test.go`, `risk/README.md`, and `risk/CHANGELOG.md`. Control head `6c6b6a8c0cf47c99aa6064d745979122963bb300`. Single-reviewer structured pass using `$requesting-code-review` and applicable coverage guidance; no children, writes, tests, builds, services, or remote operations.

### Finding

**Blocker — `risk/lockout.go:108–110`: recovery exclusion is not enforced.**

`Prepare()` gates solely on `CredentialMutation`, but the authoritative operation matrix and `risk/catalog_generated.go:102` classify `identity.mfa.recovery-use` as `CredentialMutation`. Using the existing valid request and changing only `Scope.Operation` to that operation passes every guard and returns a subject restriction intent. `Validate()` subsequently accepts it. This contradicts the explicit recovery exclusion in `Prepare()` documentation and `risk/README.md:24–28`.

**Falsification:** traced `Policy.Admit()`, `CommandCounterPolicy.Prepare()`, and `counterDimensions()`. No dominating recovery guard exists; the same valid user/session/network dimensions satisfy this profile. Native authorization does not disprove the defect: the preparation API itself promises to reject this target, independently of subsequent mutation.

**Remediation:** enforce recovery-safe operation eligibility independently of velocity profile classification, including the applicable recovery/proof-support operations; add rejection regressions exercising those canonical operations. Preserve the existing counter operation matrix.

**Publication comment:**
Blocker: `Prepare()` still accepts recovery restrictions because `CredentialMutation` includes `identity.mfa.recovery-use`. Changing only the operation in the existing valid request produces an accepted recovery-targeted intent, contradicting the documented recovery exclusion. Enforce recovery-safe eligibility separately from counter classification and add rejection coverage for the applicable recovery operations.

### Completed closure

- Policy ownership and exact request retention reject foreign custody and altered scope, actor, subject, version, or duration.
- Positive microsecond-exact ceilings, start/end partitions, version overflow, UTC precision, date bounds, and exclusive original preparation TTL are coherently enforced.
- Projections clone retained aliases; target selection cannot broaden into tenant/network/session restrictions.
- Formatting redacts custody; JSON/text serialization rejects; errors remain generic.
- Existing counter APIs and intentionally separate ephemeral/protected variants are unchanged. No production lockout consumer is introduced.
- This pure preparation batch introduces no durable writes, callbacks, goroutines, external I/O, or transaction/recovery owner.

### Evidence and limits

Both supplied final artifacts have **115 matching current inputs**, with no hash mismatches: `go test -race . -count=1` passed in **1.325s**; `go vet .` exited **0**. These results do not exercise the missing recovery rejection.

Durable mutations/CAS enforcement, administrator authorization, overrides, canonical audit, same-command recovery, Service composition, clean published consumers, and full-unit release verification remain separate unfinished obligations. They are not newly imposed gates for this intermediate batch.

## Amended compatibility

Role: `go_compatibility`. **APPROVE** for the amended intermediate preparation slice; no compatibility findings. Previous report is superseded for the earlier dirty tree.

Reviewed base/HEAD `bc9281c98b28d481f786c170ebaa718ad242f24d` plus complete current four-path delta: `risk/lockout.go`, `risk/lockout_test.go`, `risk/README.md`, `risk/CHANGELOG.md`.

The eight recovery/proof exclusions affect only newly added lockout preparation. Existing canonical operation classifications, counter policies, public APIs, consumers and dependency ownership remain unchanged. Additive signatures, zero-value rejection, immutable custody, independently owned projections and Go 1.27 compatibility remain sound. No native mutation, authorization or coordinator duplication was introduced.

Reviewed source hashes:
- `lockout.go`: `08fc83d4f25e071c6526af8f6f5eaa900a512d50d11a96a0cfec23d7ccb504c3`
- `lockout_test.go`: `baf0f2c2b30fee706fb25459450ecc06c10dcb56034d6d7f1d94f8636b4fa6ba`
- `README.md`: `a1a19eb801f5b3ceb9c83286d3480a26804d63778759f72098054cd9a47e090e`
- `CHANGELOG.md`: `45b004366ee242749fadf60c13632f681a2561142baa52b58fc9e655366ae22b`

Inspected race/vet artifacts match current Go source hashes and report race PASS `1.269s`, vet exit `0`, Go `1.27.1`. These establish development-composition results only.

Catalog reconciliation, full release gates, published-consumer compatibility and existing PostgreSQL historical-pin incompatibility remain unverified or unresolved. This disposition does not certify native lockout implementation, Service/unit completion or release readiness.

Principal: `/root/lockout_policy_compat_bc9281c`.

## Final regression

## Native regression audit — APPROVE

**Principal:** `/root/lockout_recovery_regression_bc9281c`
**Role:** fresh independent `regression_reviewer`, read-only; no writes, tests, builds, services or children.
**Reviewed identity base/head:** `bc9281c98b28d481f786c170ebaa718ad242f24d`, plus all four dirty paths: `risk/lockout.go`, `risk/lockout_test.go`, `risk/README.md`, `risk/CHANGELOG.md`.
**Control head:** `6c6b6a8c0cf47c99aa6064d745979122963bb300`.

No concrete regression or blocking coverage defect found in the bounded immutable lockout-preparation contract.

- The new regression exercises public `Prepare` with otherwise-valid requests for all eight canonical factor-proof/recovery/admin-support operations. Removing their exclusion would admit those requests; an error and empty target are required.
- Canonical counter classifications remain unchanged. Authentication, privileged and unknown operations reject independently; ordinary password-change preparation remains the positive control.
- Existing tests cover immutable projection ownership, foreign-policy rejection, changed request bindings, Start/End version/duration rules, exclusive preparation expiry, invalid primary time, tenant isolation and sensitive-output rejection.
- Counterfactual inspection found no route around the exclusion through `EndLockout`, caller factor flags or forged requests: both modes share operation admission, and opaque commands retain exact policy/request custody.
- All **115 recorded input hashes** match current source for each final artifact:
  - `go test . -run '^TestLockoutPreservesFactorAndAdministrativeRecoveryRoutes$' -count=1`: exit 0, **0.194s**.
  - `go test -race . -count=1`: exit 0, **1.269s**.
  - `go vet .`: exit 0.

**Claim boundary:** approval covers preparation and its executed development-composition tests only. Current administrator authorization, durable CAS/mutation/audit/recovery, Service composition, published clean-consumer compatibility, release gates and risk-unit completion remain unproven and explicitly unfinished.

## Complete final correctness

## Native complete final-diff re-review

**Principal:** `/root/lockout_policy_correctness_bc9281c`
**Internal disposition: APPROVE — bounded preparation batch only.**

**No findings.** The prior recovery-exclusion blocker is resolved.

**Reviewed:** identity base/head `bc9281c98b28d481f786c170ebaa718ad242f24d` plus the complete final four-path dirty diff: `risk/lockout.go`, `risk/lockout_test.go`, `risk/README.md`, and `risk/CHANGELOG.md`. Control head `6c6b6a8c0cf47c99aa6064d745979122963bb300`. Read-only single-reviewer pass; no children, writes, execution of tests/builds, services, or remote mutations.

### Closure

- `Prepare()` now rejects eight canonical recovery/factor-proof/administrative-recovery operations independently of their unchanged `CredentialMutation` classifications.
- Falsification: traced the formerly admitted `identity.mfa.recovery-use` request through the new gate. It now returns an error and zero command before keyed-target preparation. The same guard applies to Start and End; changing the request during validation cannot bypass immutable original-request ownership.
- The eight rejection cases correspond to the selected factor/recovery boundaries. Authentication-class recovery and passkey-signin operations remain excluded by existing class admission.
- New tests exercise public preparation using independently explicit operation IDs and assert both rejection and absence of a restriction target. Ordinary password-change preparation remains the positive control.
- Rechecked the complete batch’s policy ownership, scope/CAS retention, duration/time/overflow boundaries, exclusive original TTL, owned projections, redaction, serialization rejection, errors, and unchanged counter variants.
- No durable mutation, callback, external I/O, goroutine, transaction owner, or production consumer is introduced.

### Current evidence

All **115 inputs match** each supplied final artifact:

- `go test . -run '^TestLockoutPreservesFactorAndAdministrativeRecoveryRoutes$' -count=1`: PASS, **0.194s**.
- `go test -race . -count=1`: PASS, **1.269s**.
- `go vet .`: exit **0**.
- Read-only `git diff --check`: PASS.

Final reviewed source hashes: `lockout.go` `08fc83d4…`; `lockout_test.go` `baf0f2c2…`; `README.md` `a1a19eb8…`; `CHANGELOG.md` `45b00436…`.

Approval does **not** establish native mutations, durable CAS/authorization, overrides, canonical audit/recovery, Service composition, clean published consumers, production readiness, release verification, or completion of either risk unit or the full platform. Those original obligations remain unfinished.

## Delivery

**Delivery disposition: APPROVE — local intermediate preparation checkpoint only.**

- **Local:** clean identity worktree, `feature/identity-platform`, HEAD `4398e2dde31d82dac7005bcbc96610aba0645dcb`.
- **Committed:** exactly four requested paths committed over `bc9281c98b28d481f786c170ebaa718ad242f24d`. Control HEAD remains `6c6b6a8c0cf47c99aa6064d745979122963bb300`; ledger/reports/artifacts remain dirty or untracked, **not committed**.
- **Pushed:** no push recorded; no local remote-tracking ref contains this commit. Remote publication was not independently queried.
- **CI:** not verified for this head.
- **Deployed:** not deployed under this batch.
- **Production:** not verified.

**Native review inspection:** final regression `/root/lockout_recovery_regression_bc9281c`, compatibility `/root/lockout_policy_compat_bc9281c`, and complete correctness `/root/lockout_policy_correctness_bc9281c` each record **APPROVE** for the complete final four-path preparation delta. They reviewed `bc9281c` plus dirty contents, not the later commit identity. All four recorded final hashes exactly match current files **and committed blobs**; the identity checkout is clean. The initial recovery-exclusion blocker remains preserved with its final closure.

**Verification evidence inspected, not rerun:** all 115 recorded input hashes match current source for each artifact. Original execution revision remains `bc9281c`, accurately preserving precommit execution:
- focused recovery regression: exit 0, `0.194s`;
- risk race suite: exit 0, `1.269s`;
- risk vet: exit 0.

These artifacts establish recorded **development-composition** results, not release/published-consumer proof; removed dependency archives were not independently rehashed here.

**Findings:** none against this bounded checkpoint. Scratch `/private/tmp/identity-lockout-policy.wtzdQB` is absent. No database/container work is recorded for this batch; this audit performed no mutations, tests, builds, or service access.

The full **67-unit goal remains active**, with **5/67 integrated and 1/67 verified unchanged**. Native authorized lockout mutations/CAS overrides, audit/recovery, canonical Service, historical dependency repair, release/coverage/mutation gates, and login/shared-issuer completion remain unfinished. The goal attachment explicitly prohibits push/publication/deployment and permits local committed delivery only.

Principal: `/root/lockout_policy_delivery_4398e2d`.
