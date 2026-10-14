# Native CAPTCHA preparation development batch

## Public executor committed checkpoint d1df7de

Owned scratch `/private/tmp/identity-captcha-key-custody.sbAEQe` removed after
the source and control commits; all current-boundary disposable PG containers,
volumes and caches removed. No permanent replacements introduced.

### Native delivery report

**Role/principal:** `delivery_auditor`, `/root/captcha_public_delivery_d1df7de`
**Disposition:** **APPROVE — local committed checkpoint only. No findings.**

- Source clean on `feature/identity-platform`, HEAD `d1df7dee0a5627a8d6bb568be165fd3fce5f65a8`, parent `7d365d68b552ab0cb81273fd2cd50edcfd33e6b0`.
- Commit contains exactly the seven specified `risk/postgres` paths. Independently reproduced ordered manifest digest `3fb6c22c6920b2f9664e503d224015273f45def52f9888e92f6d1002cc55e091`; committed content matches final reviewed snapshot.
- Control clean at `6593fe416e40d7983413a202c124c3b0f23c0898`. Inspected retained native original/final regression, compatibility and correctness reports directly. Fresh dispositions are APPROVE; prior authorization-oracle, restart and augmented-attribute findings are retained and closed. Final constructor correction and actual 128/129 boundary regression are present.
- Recorded RED20747, current focused PostgreSQL/race PASS4.464s and vet exit0 (22198), plus unchanged executor-consumer PASS1.813s remain supplied execution evidence—not independently rerun. Commit metadata does not change their matching source inputs.
- Ledger retains **5/67 integrated, 1/67 verified**, active full goal and unfinished login/shared issuer. CAPTCHA production-effect composition, protected consumption, full coverage/mutation/provider/source-pin/clean-consumer and release gates remain incomplete.

**Delivery boundary:** locally committed; no pushed, hosted-CI, deployed or production-verified state established. No full-unit, release or platform completion certified.

Read-only audit; no writes, Go execution, services, remotes or children.

## Constructor final-context regression on source7d365d6

Current ordered seven-path source manifest:
`3fb6c22c6920b2f9664e503d224015273f45def52f9888e92f6d1002cc55e091`.

Independent correctness principal `/root/captcha_public_correctness_7d365d6`
returned REQUEST_CHANGES against seven-path manifest
`a899c95dfb297069e29ec986de2167c76aeba8fef77826a1c7fe8307d94dee15`:
initial caller validation accepted 122 unrelated attributes, then seven trusted
attributes made an unusable 129-attribute caller rejected by coordinator policy.
This is a constructor contract defect, not an authority leak. Native report
retention and fresh complete-diff review remain required.

Actual isolated PG18.6 reproduction, terminal20747:
`go test . -tags integration -run
'^TestIssuedEvidenceSameWorkIntegration$/native-captcha$/public-executor$/pre-auth-anonymous$'
-count=1` failed at captcha_executor_integration_test.go:132:
`constructor accepted 129 final policy attributes` (exit1).
The same regression first proves an accepted final128 caller can really commit
public issuance. Minimal fix revalidates the final augmented caller before common
executor construction. Terminal22198: `go test . -race -tags integration -run
'^TestCaptcha|^TestExecutor|^TestIssuedEvidenceSameWorkIntegration$/native-captcha'
-count=1` PASS4.464s; `go vet -tags integration .` exit0. This includes the
exact128 successful issuance and final129 constructor rejection. Go1.27.1,
same frozen temporary primitive wiring and pinned PG18.6 image as prior runs;
owned green container, volume and cache removed by its trap.
The reproduction's owned container, volume and cache were removed by its trap.

### Current native correctness review

- **Role/principal:** independent `correctness_reviewer`, `/root/captcha_public_correctness_7d365d6`.
- **Disposition:** **APPROVE** for the scoped local issuance/recovery development batch.
- **Boundary:** `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`, branch `feature/identity-platform`; base/HEAD `7d365d68b552ab0cb81273fd2cd50edcfd33e6b0` plus all seven supplied dirty/untracked paths.
- **Snapshot identity:** ordered seven-path SHA-256 manifest `3fb6c22c6920b2f9664e503d224015273f45def52f9888e92f6d1002cc55e091`, independently matched before and after review.
- **Method:** complete fresh literal-diff sweep, then caller, authorization, coordinator, persistence, encryption, recovery, sibling and test-oracle tracing. Applied the previously loaded `requesting-code-review` contract and applicable boundary/testing references. No writes, execution of Go, service access, remote operations or children.

#### Findings and previous finding closure

**No unresolved findings in this scoped batch.**

The prior final-attribute-count finding is independently closed:

- `captcha_executor.go:71` now evaluates `boundedExecutorCallerShape()` **after** trusted attribute injection. Final 129 attributes fail through the existing `newExecutor()` invalid-policy gate.
- The integration regression constructs distinct 121/122 unrelated attributes, independently yielding final 128/129. It asserts rejection at 129 and exercises actual public issuer/executor/coordinator issuance at 128, rather than merely inspecting construction.
- Parent supplied actual red execution **RED20747**, failing “constructor accepted 129 final policy attributes”, followed by current focused green execution.
- README changes remove the stale statements that public issuance/recovery remains unimplemented while retaining protected-use and release exclusions.

#### Complete affected-boundary dispositions

- **Composition/security:** immutable exact flow, tenant/target/action and administrative actor checks; cloned caller state; authoritative overwrite and final bounded validation. Challenge/operation/session mixing cannot reach reservation because `query()` requires whole-flow equality.
- **Current authorization:** admission and original disclosure use separate current-policy decisions. Denied, wrong-scope, retired, malformed and unavailable observations do not produce a bearer.
- **Persistence/recovery:** command mapping authenticates signed custody; native participant validates selected profile and sealed ownership. Evidence, required audit/lifecycle and committed receipt remain one shared Work. Unknown/rollback paths do not expose candidates.
- **Original response:** indexed tenant/command lookup, bound metadata/AAD, retained intent/key match and core bearer integrity authenticate the historical response. Restart/expiry do not renew it; consumption authority remains separate.
- **Concurrency/re-entry:** existing global command claim, generation fence, replay winner/alias uniqueness and participant locks remain the owner. Changed command, changed intent, live pending and terminal history cannot manufacture a successor reference.
- **Scale/resources:** ≤16 binding candidates/aliases, indexed single-row recovery, ≤1024-byte envelope, bounded contexts and coordinator cleanup; no new unbounded loop, private transaction or goroutine.
- **Compatibility:** old `NewExecutor()` preserves its exact operation/resource gate and shared resource/registry validation. No changed persisted schema or dependency manifest.
- **Tests:** real internal execution with independent literal policy expectations, durable effect counts, real commit-ack fault, audit rollback, replay, restart, current denial, forged caller attributes and final attribute limits. Synthetic verification/effect adapters prove this selected integration boundary, not official provider or production-effect interoperability.

#### Verification and limits

Parent-supplied current command:

`go test . -race -tags integration -run '^TestCaptcha|^TestExecutor|^TestIssuedEvidenceSameWorkIntegration$/native-captcha' -count=1`

**PASS, 4.464s**, terminal **22198**; integration vet **exit 0**. Prior old-executor integration race **PASS, 1.813s** exercises the unchanged sibling consumer. Trusted-prefix mutant failure remains narrow defect-detection evidence. Independently ran read-only `git diff --check`: **PASS**.

Approval is **not** full package/platform completion, release readiness, protected consumption, production CAPTCHA-effect composition, provider interoperability, exact full coverage/mutation, source-pin, clean-consumer, publication or deployment certification.

#### Historical native report — verbatim

## Native correctness review

- **Role/principal:** independent `correctness_reviewer`, `/root/captcha_public_correctness_7d365d6`.
- **Disposition:** **REQUEST_CHANGES**.
- **Boundary:** source worktree `golib-go-identity-identity-platform`, branch `feature/identity-platform`; base/HEAD `7d365d68b552ab0cb81273fd2cd50edcfd33e6b0` plus the complete seven supplied dirty/untracked paths.
- **Snapshot identity:** ordered seven-path manifest SHA-256 `a899c95dfb297069e29ec986de2167c76aeba8fef77826a1c7fe8307d94dee15`, independently matched before and after inspection.
- **Method:** read-only literal-diff sweep, then causal/security/state/oracle inspection using `requesting-code-review`, its applicable boundary references, `testing-coverage`, and `testing-integration`. No children, writes, Go execution, services, or remote operations.

### Finding: validate the final augmented caller before accepting composition

**P2 — `risk/postgres/captcha_executor.go:68–71`.**

`NewCaptchaExecutor()` validates caller shape **before** adding seven trusted request attributes, then passes unconditional `true` to `newExecutor()`. A caller with 122 distinct, otherwise valid unrelated request attributes passes the constructor's ≤128 bound, but trusted augmentation produces 129. Every subsequent admission/recovery authorization rejects this retained caller in `postgres/query.go:287–288,358–360`, before the authorization service runs. Construction therefore succeeds with an unusable executor; issuance returns denial despite otherwise matching scope and an allowing policy.

**Falsification:** traced cloning, attribute replacement, coordinator construction, and authorization normalization. Distinct unrelated names survive cloning and all seven trusted keys are additions; the coordinator does not remove request attributes before `boundedQueryCaller()`. Its cleanup concerns resource attributes only. Thus 122 + 7 deterministically exceeds 128. No runtime execution was authorized.

**Smallest remediation:** revalidate `boundedExecutorCallerShape()` after trusted augmentation and pass that result to `newExecutor()`. Add a constructor boundary regression distinguishing final 128 from final 129 attributes.

### Remaining causal review

No additional concrete defects established:

- **Trusted composition:** exact flow equality, tenant/target/action binding, admin actor identity, caller cloning and authoritative attribute overwrite traced.
- **Disclosure:** independently current recovery authorization precedes the private primary lookup; retained binding, original intent, metadata/AAD and bearer authentication fence output.
- **Atomicity/retry:** shared coordinator owns reservation, domain Work, required audit/lifecycle and receipt; rollback/unknown paths disclose no provisional reference. Global command and replay-alias uniqueness remain the concurrency owner.
- **Historical response:** expired recovery preserves original times/reference and performs no provider verification; later evidence state does not itself confer protected-use authority.
- **Scale/resources:** indexed one-row original lookup, ≤16 retained binding candidates and replay aliases, ≤1024-byte encrypted response, bounded contexts; no new unbounded I/O loop or goroutine.
- **Sibling compatibility:** old `NewExecutor()` retains its exact action/resource gate; shared constructor extraction otherwise preserves registry/resource validation.
- **Oracles:** public integration fixtures reach the real issuer/executor/coordinator/repository. Literal policy expectations, independent SQL effect counts, actual lost-ack fault, restart restoration, current denial and provider-call assertions are meaningful scoped observations.

### Evidence and claim limits

Parent-supplied focused integration race run: **PASS, 2.319s**; integration vet: **PASS**. Parent-supplied old-executor integration race: **PASS, 1.813s**. Reported missing-public-API compile red and trusted-attribute-prefix mutant failure support their respective narrow claims; I did not rerun them.

This report does **not** certify full package/release completion, production CAPTCHA effects, protected consumption, official provider interoperability, exact coverage/full mutation, source-pin or clean-consumer gates. Those remain outside this development-batch review.

### Native implemented-coverage report after augmented-bound correction

- **Principal/role:** `/root/captcha_public_regression_7d365d6`, `regression_reviewer`
- **Disposition:** **APPROVE**
- **Reviewed identity:** `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`, `feature/identity-platform`, base/HEAD `7d365d68b552ab0cb81273fd2cd50edcfd33e6b0`, complete current seven specified dirty/untracked paths.
- **Scope:** public CAPTCHA issuance and independently authorized original encrypted-response recovery development slice.
- **Method:** `$testing-coverage`, read-only complete-diff/source/test inspection against the selected goals and retained bounded QA requirements; no execution, writes, services, children or remote operations.

**No unresolved material coverage defect found.** The constructor now revalidates the augmented caller context. The boundary regression independently proves rejection of 129 final attributes and exercises actual public issuance at the accepted 128-attribute limit, preventing constructor-only false confidence.

The complete inspected slice also protects trusted attribute overwrite/copying, separate current admission and recovery authorization, six flow variants, foreign-context rejection before reservation/provider work, SQL-confirmed atomic success and audit rollback, actual lost acknowledgement without bearer disclosure, exact replay, and original-response recovery through fresh repository/executor/coordinator instances without provider verification or expiry renewal. Existing `NewExecutor` caller restrictions remain structurally unchanged.

**Parent-supplied execution evidence:** augmented-boundary regression RED (`20747`); current focused race/PostgreSQL suite **PASS 4.464s**, integration vet **exit 0** (`22198`). Existing executor consumer integration **PASS 1.813s** remains applicable to unchanged inputs. The earlier injected-prefix mutant failed actual public issuance, demonstrating detection of the trusted-policy-input defect. I did not independently execute these checks.

This approval does not certify CAPTCHA protected consumption, production-effect interoperability, complete lifecycle/retirement, exact coverage/mutation, provider/source-pin/clean-consumer gates, release readiness or platform completion.

### Fresh native Go compatibility report after augmented-bound correction

- **Principal:** `/root/captcha_public_compat_7d365d6`, independent read-only reviewer.
- **Snapshot:** `feature/identity-platform`, base/HEAD `7d365d68b552ab0cb81273fd2cd50edcfd33e6b0`, plus all seven current specified dirty/untracked paths.
- **Disposition:** **APPROVE** for local-development CAPTCHA public execution/recovery.
- **Findings:** None. Earlier reports remain historical; this disposition covers the latest constructor, tests and documentation.

#### Compatibility conclusions

- `captcha_executor.go:64-71` now validates the **augmented** immutable caller context, preventing construction of an executor whose added trusted attributes exceed the coordinator's supported bound.
- `captcha_executor_integration_test.go:121-169` covers rejection at 129 final attributes and actual committed issuance at 128, rather than constructor-only acceptance.
- Existing `NewExecutor` signature, operation/resource admission, registry requirements and supported production-adapter consumers remain unchanged (`executor.go:84-95,288-293`).
- Exact flow/tenant/target/administrative-actor isolation, separate current admission/recovery authorization, encrypted original-response recovery and retained signed configuration remain intact.
- README updates consistently identify public issuance/recovery as implemented while retaining unfinished protected-use/lifecycle/release boundaries.
- Go 1.27 and module dependencies remain unchanged; no provider dependency, permanent replacement or cross-module ownership expansion was introduced.

#### Verification boundary

Parent supplied regression **RED**, terminal `20747`; current focused native CAPTCHA race **PASS, 4.464s**, integration vet **exit 0**, terminal `22198`. Existing Executor production-effect integration **PASS, 1.813s** remains supplied unchanged evidence. I did not independently execute these checks.

**Adoption limits:** No full-package/release certification, published-source-pin or clean-consumer verification, CAPTCHA production-effect composition, real provider integration, protected consumption, complete retirement/cleanup, deployment or production-effect claim.

Review was read-only: no writes, Go execution, services, remote operations, children or other reviewer reports.

## Public executor development in progress after7d365d6

Corrected current seven-path source manifest SHA256
`a899c95dfb297069e29ec986de2167c76aeba8fef77826a1c7fe8307d94dee15`.
New literal independently observing policy denies missing/forged trusted flow
attributes and validates separate admission/recovery markers; caller map changes
after construction cannot change them. Actual fresh repository/executor/coordinator
restart reconstructs only retained signed command, with provider verification
forbidden. Absent-command current admission denial creates no claim or reference.
Current same focused race command PASS2.319s / integration vet exit0 (50742).
Existing `NewExecutor` integration consumer selector
`'^TestExecutor|^TestIssuedEvidenceSameWorkIntegration$/official-executor'`,
race PASS1.813s (69886), including owned audit/outbox effects and lost ack.
Owned source-copy mutation of injected attribute prefix failed specifically
public pre-auth-anonymous issuance at captcha_executor_integration_test.go:152
(25302): the policy oracle detects the real security-input defect. Fixture
copy and all four current-turn PG containers/volumes and test caches removed.
Native regression/compatibility reports retained below; final correctness pending.

### Original public native regression report

- **Role/principal:** `regression_reviewer`, `/root/captcha_public_regression_7d365d6`
- **Disposition:** **REQUEST_CHANGES**
- **Reviewed state:** source base/HEAD `7d365d68b552ab0cb81273fd2cd50edcfd33e6b0` plus the seven specified dirty/untracked `risk/postgres` paths.
- **Boundary:** public CAPTCHA issuance and authorized encrypted-original recovery development slice only.
- **Method:** `$testing-coverage`; read-only source/test/requirements inspection. No writes, execution, services, remote operations, children, or other review conclusions.

#### Material finding

**P2 — Trusted authorization attributes have no behavioral oracle.**
`risk/postgres/captcha_executor_integration_test.go:76–77` supplies the same always-allow `issuanceQueryPolicy` for admission and recovery. The later deny policy checks unconditional rejection, not the trusted flow delivered to either policy.

Consequently, removing or corrupting the `identity.captcha.*` attribute injection in `captcha_executor.go:65–67` would not fail these tests. Those attributes are the new public mechanism allowing current policies to distinguish protected operation, pre-auth transaction, session, administrator, challenge and request. Constructor target checks and intent equality do not prove this policy input.

**Required correction:** use independently observing admission/recovery policies that assert the exact trusted attributes, tenant/resource namespace and actor for the existing flow variants. Seed conflicting caller attributes and mutate caller attributes after construction to prove trusted overwrite and ownership. Include current admission denial of an absent command with zero provider calls and no reservation.

This is a concrete assertion gap in the implemented security contract, not a request to elaborate the planning artifact. The counterfactual above is source-inspection reasoning; no mutation was executed.

#### Verification-claim limitation

`captcha_executor_integration_test.go:169–182` reconstructs the issuer but retains `profile.Executor = capture`, whose delegate is the original executor. It proves issuer restoration and expired-history conservation, **not executor/coordinator/repository restart**.

Before claiming public executor restart recovery, rebuild the repository/custody and `NewCaptchaExecutor` from retained configuration, restore only the captured prepared command, and compare the full historical response with provider verification forbidden. Otherwise narrow the restart claim explicitly.

#### Positive evidence

The inspected assertions cover all six distinct flow variants, SQL-confirmed evidence/audit/lifecycle atomicity, audit rollback without bearer disclosure, actual lost-ack commitment with unknown response, exact replay without provider reevaluation, foreign-challenge denial before reservation, historical response equality and independent current recovery denial.

The shared constructor extraction preserves the old `NewExecutor` action/resource/subject and bounded-caller checks structurally. Existing old-executor integration scenarios remain available, but the supplied current focused selector does **not** execute its `official-executor-*` integration subtests.

#### Execution evidence and unfinished boundaries

Parent-supplied evidence: focused race/integration tests **PASS 2.457s**; integration-tagged vet **exit 0**, terminal `81695`; missing-public-API compile RED `96202`. I did not independently execute these checks.

Public protected consumption, production CAPTCHA audit/outbox interoperability, complete coverage/mutation, provider truth, source pins, clean-consumer and package/platform/release completion remain unverified and outside this disposition.

### Fresh public native regression report

- **Role/principal:** `regression_reviewer`, `/root/captcha_public_regression_7d365d6`
- **Disposition:** **APPROVE**
- **Reviewed state:** base/HEAD `7d365d68b552ab0cb81273fd2cd50edcfd33e6b0`, complete current seven-path dirty/untracked slice.
- **Boundary:** public CAPTCHA issuance and independently authorized encrypted-original recovery—not package/platform/release completion.
- **Method:** `$testing-coverage`; read-only inspection, no execution or mutations.

**Previous finding resolved.** The new `captchaFlowPolicy` independently checks caller/tenant/target namespace, all seven trusted flow attributes and distinct admission/recovery markers. Conflicting supplied attributes and postconstruction map mutation now exercise trusted overwrite and copied ownership. Absent-command admission denial forbids provider work and independently verifies no command claim.

**Previous restart limitation resolved.** Recovery now uses a newly constructed repository, executor and coordinator, restores prepared command custody, forbids provider verification and compares the complete expired historical response with the original.

No unresolved material regression or assertion defect found within this bounded slice. The shared `NewExecutor` extraction preserves its existing admission contract; its invalid-composition and production-effect integration scenarios were additionally executed according to parent-supplied evidence.

**Reported execution evidence:** focused CAPTCHA race/PostgreSQL tests **PASS 2.319s**, integration vet **exit 0** (`50742`); existing executor integration regressions **PASS 1.813s** (`69886`). The attribute-prefix mutant failed at public issuance as intended (`25302`, test exit 1), supporting the corrected oracle’s defect detection. These execution outcomes are parent-supplied, not independently rerun.

Public protected consumption, CAPTCHA production-effect interoperability, full coverage/mutation, provider/source-pin/clean-consumer and release gates remain unfinished.

### Original public native Go compatibility report

- **Principal:** `/root/captcha_public_compat_7d365d6`, independent read-only Go compatibility reviewer.
- **Reviewed identity:** `feature/identity-platform`, base/HEAD `7d365d68b552ab0cb81273fd2cd50edcfd33e6b0`, plus the seven specified dirty/untracked paths. Worktree state remained unchanged.
- **Boundary:** Local-development public CAPTCHA execution and original-response recovery; complete diff, consumed risk/kernel/coordinator contracts, existing `NewExecutor` callers, module metadata, and authoritative goal/transaction requirements.
- **Disposition:** **APPROVE** for this bounded batch.
- **Findings:** No concrete Go compatibility defect found.

#### Supporting observations

- `executor.go:84-95,288-293`: Existing `NewExecutor` signature and risk-issuance operation/resource restrictions remain intact; extraction into `newExecutor` does not broaden its accepted caller shape.
- `captcha_executor.go:50-75,175-185`: CAPTCHA is additive and explicitly selected. Constructor checks exact operation, tenant, target namespace and administrative actor; requests must match the complete frozen flow, including session, pre-auth, challenge and request fingerprint.
- `captcha_executor.go:82-103,188-207,232-270`: Admission and independently current-authorized recovery use existing coordinator ownership. Unknown outcomes disclose no bearer. Committed recovery authenticates historical binding, matching intent/key version and encrypted original response.
- Existing `restoreCaptcha` and core `RestoreCaptchaIssue` retain signed configuration, original aliases and original validity; this batch does not introduce regeneration or provider reevaluation.
- `risk/postgres/go.mod` is unchanged, retaining the existing Go 1.27/module boundary. No new dependency, replace directive or provider dependency was introduced.

#### Evidence and limits

Parent supplied focused native unit/integration/race **PASS (2.457s)** and integration vet **exit 0**, session `81695`. I performed source inspection only and did not independently execute those checks.

This approval does **not** establish full-package/release completion, pinned published dependency compatibility, clean-consumer adoption, production audit/outbox composition, provider integration, protected consumption, lifecycle/retirement completion, deployment or production behavior. No external mutation, file write, service access or Go execution occurred.

### Fresh public native Go compatibility report

**Principal:** `/root/captcha_public_compat_7d365d6` — independent, read-only compatibility reviewer.

**Reviewed identity:** `feature/identity-platform`; exact base/HEAD `7d365d68b552ab0cb81273fd2cd50edcfd33e6b0`, plus the complete current seven-path dirty/untracked batch. Original report remains retained; this report covers the subsequently changed tests.

**Scope:** Local-development public CAPTCHA issuance/recovery, consumed risk/kernel/coordinator APIs, existing Executor compatibility and module boundaries.

**Disposition: APPROVE. No concrete compatibility findings.**

#### Current evidence

- `executor.go:84-95,288-293`: Existing constructor signature, operation/resource restrictions and registry validation remain preserved.
- `captcha_executor.go:50-75,175-185`: Additive opt-in construction retains exact tenant, target namespace, administrative actor and complete immutable flow matching.
- `captcha_executor.go:82-103,188-207,232-270`: Separate current admission/recovery authorization, fail-closed ambiguity and authenticated original-response disclosure remain intact.
- `captcha_executor_integration_test.go:35-60,105-120`: Separate observing policy oracles independently check literal expected caller/flow attributes and admission/recovery markers; conflicting attributes and subsequent caller-map mutation cannot silently satisfy the oracle.
- `captcha_executor_integration_test.go:207-238`: Restart now constructs a fresh repository/executor/coordinator, forbids provider verification and compares the recovered historical result without expiry renewal.
- `captcha_executor_integration_test.go:259-286`: Current denial on an absent command asserts no disclosure or reservation; provider access would fail the test.

Module metadata and Go 1.27 baseline remain unchanged. No new dependency or permanent replacement was introduced.

**Execution evidence supplied by parent:** focused native race **PASS, 2.319s**, integration vet **exit 0**, terminal `50742`; existing Executor invalid-composition/production-effect integration race **PASS, 1.813s**, terminal `69886`. Not independently executed.

**Unverified boundaries:** Full-package/release gates, published source pins, clean-consumer adoption, CAPTCHA production-effect composition, provider integration, protected consumption, complete lifecycle/retirement, deployment and production effects.

No files changed, Go execution, services, remote operations or children.


Source base/head7d365d6 plus seven risk/postgres paths. Primary public-contract;
secondary security, persistence, concurrency and replay/ambiguity. The retained
whole-issuance mapper/QA below covers this implementation; no planning restart.
Missing NewCaptchaExecutor/CaptchaExecutorConfig RED96202. Constructor, actual
public issuer, same-Work command/audit/lifecycle/evidence, original response,
six flow variants, exact replay, foreign flow admission, independent current
recovery denial, restart without token, lost commit acknowledgement and audit
rollback tested against real isolated PG18.6. Effects are SQL seams, not the
production audit/outbox composition. Initial race GREEN1.829s/vet0 (12959).
Current widened focused command `go test . -race -tags integration -run
'^TestCaptcha|^TestExecutor|^TestIssuedEvidenceSameWorkIntegration$/native-captcha'
-count=1` PASS2.457s (81695); `go vet -tags integration .` exit0, same terminal
session. No execution cache remains.
Go1.27.1; exact temporary primitive wiring same owned directory recorded below.
Protected use/lifecycle/retirement, production effects, provider and full
coverage/mutation/clean-consumer release gates open; independent reviews pending.

## Native participant committed checkpoint

### Native delivery audit

**Role/principal:** delivery_auditor, `/root/captcha_native_delivery_7d365d6`
**Disposition:** APPROVE — local committed CAPTCHA checkpoint only. No delivery findings.

- Source is clean on `feature/identity-platform`, HEAD `7d365d68b552ab0cb81273fd2cd50edcfd33e6b0`, parent `9caf047417b0dddbdebee03581ad9c27a296b999`.
- Commit contains exactly the 20 reviewed paths. Ordered file manifest reproduces `c632d037eeebab5b554a62b1e3092916b78e2408b49e93351fc87e80f1350efe`.
- Inspected native original/fresh regression and compatibility reports plus complete final-diff correctness report directly. Earlier findings are retained and resolved; fresh dispositions are APPROVE against matching source content.
- Control advanced during audit to clean committed HEAD `7f21eaf481c8147df30a7f588d080e70240df69c`, parent `63fb790d947be764432eb7216be0bb21edb7133b`. Exactly the two identified documentation paths were committed.
- Focused race/real-SQL/vet results are recorded execution evidence, not independently rerun or full release verification. No push, hosted CI, deployment, provider verification, package/platform completion, public executor/consumption, full coverage/mutation, source-pin or clean-consumer claim is supported.
- Parent-owned scratch `/private/tmp/identity-captcha-key-custody.sbAEQe` remains present and actively needed; its later cleanup remains parent responsibility. No cluster-removal assertion was independently verified.

Read-only audit; no writes, Go execution, services, remote operations or children.

### Committed input identity

Source `7d365d68b552ab0cb81273fd2cd50edcfd33e6b0`, parent9caf047.
Exactly20 reviewed paths committed unchanged, source clean. Review manifest
SHA256 `c632d037eeebab5b554a62b1e3092916b78e2408b49e93351fc87e80f1350efe`
reproduced before staging. Original tracked patch SHA256
`efabc0e4c7cd056a7415c73f3b1f18718e8cadf951629d675ac225318ae33e5a`.
Final risk/postgres focused race+real SQL PASS2.105s and integration vet exit0
(89724), same command recorded below. Core1.164s/root1.289s unchanged.
Additional RED: populated scope literal marker corruption64800; missing
classification helper4951. GREEN proves exact literal scope preservation and
authentic/corrupt/foreign-binding/absent row classification with real SQL rows.
ACK Core message and line validation passed; effective impact unspecified.
No profile or release decision invented. No push or broader completion claim.

### Original native compatibility report

**Role/principal:** Independent Go compatibility reviewer `/root/captcha_native_compat_9caf`
**Disposition:** `REQUEST_CHANGES`
**Boundary:** `feature/identity-platform`, base/committed HEAD `9caf047417b0dddbdebee03581ad9c27a296b999`, plus all current tracked/untracked changes in `postgres`, `risk`, and `risk/postgres`. Read-only source review; no execution or external mutations.

#### Finding: migration rewrites valid original scope metadata

**P2 — captcha_scope_migrations.go:69.** `Load()` injects the quoted custody projection before replacing every `@schema@` occurrence. Consequently, valid configuration such as `Site: "site-@schema@"` becomes `site-"risk_store"` inside the generated SQL literal. Core validation permits this value (captcha_replay.go:131). Existing aliases retain the original site, so the backfill join cannot locate their supplied commitment; the populated upgrade fails its non-null constraint despite correct original custody. API-profile/configuration identifiers containing the marker have the same exposure.

**Remediation:** Replace schema placeholders before inserting custody values, or use non-recursive simultaneous substitution. Add a populated-history regression asserting exact marker-containing scope preservation.

**Falsification:** Rechecked core validation and SQL quoting: neither rejects nor escapes `@schema@`; the second whole-string substitution necessarily modifies the inserted literal. This conclusion is source-derived, not runtime reproduction.

#### Remaining assessment

No other source-confirmed compatibility defect found in historical signed custody/intent/response preservation, original alias conservation, retained-key addition, sealed generation linkage, or optional feature isolation. Go 1.27 module boundaries remain structurally unchanged.

Tests/builds were not run. Public executor/protected consumption, complete rotation/retirement, coverage/mutation/provider verification, dependency-pin/source-lock integration, and clean-consumer release gates are **not certified** by this report.

### Fresh native Go compatibility report

**Role/principal:** Independent Go compatibility reviewer `/root/captcha_native_compat_9caf`
**Disposition:** `APPROVE`
**Reviewed boundary:** Base/committed HEAD `9caf047417b0dddbdebee03581ad9c27a296b999`, branch `feature/identity-platform`, complete current 20-path tracked/untracked batch in `postgres`, `risk`, and `risk/postgres`.

**No findings.**

**Prior report retained; finding resolved:** The earlier `REQUEST_CHANGES` correctly identified scope-marker corruption. captcha_scope_migrations.go:68 now substitutes schema before injecting custody literals. The populated migration fixture retains `fixture-@schema@-site`, directly exercising the original failure. The previous finding is closed, not withdrawn as incorrect.

The complete current source preserves historical signed configuration bytes, intent/response identity, original aliases and expiry; retained-key additions remain distinct from replacement/retirement. Scope serialization and exact-generation transfer remain coordinator-owned. Recovery classification stays unexported and downstream of sealed primary ownership; persisted authentic/corrupt/foreign-binding tests now assert the classification directly.

**Evidence boundary:** No commands executing Go or services were run by this reviewer. Parent reports current focused race/actual-SQL pass and integration-tag vet pass; these are supplied execution evidence, not independently rerun results. Public executor/protected consumption, complete retirement policy, full release coverage/mutation/provider gates, dependency-pin/source-lock integration and clean-consumer release verification remain uncertified. Approval applies to this local development batch, not release/readiness or full-unit completion.

### Original native regression report

- **Role/principal:** `regression_reviewer`, `/root/captcha_native_regression_9caf`
- **Reviewed state:** base/HEAD `9caf047417b0dddbdebee03581ad9c27a296b999`, plus all 20 dirty tracked/untracked paths in the assigned identity worktree.
- **Boundary:** native reservation/rebind, domain issuance, participant recovery, sealed ownership, scope-key migration and restricted configuration/key commitments.
- **Disposition:** **REQUEST_CHANGES** — one concrete recovery-test oracle gap; no demonstrated production defect.
- **Execution:** read-only inspection; no writes, test execution, services or child agents. Focused execution results are parent-reported.

#### Required finding

**P2 — Assert the participant’s authentic-positive recovery classification separately from aggregate uncertainty.**

At `risk/postgres/captcha_execution_integration_test.go:383–431`, authentic-positive, corrupt-response and foreign-generation histories all assert the same aggregate `InProgress` outcome. Because the authentic-positive fixture deliberately lacks other participant effects, that aggregate result remains correct even if `recoverCaptcha` always returns `RecoveryUnknown` for present evidence.

Consequently, replacing the authenticated-positive return at `risk/postgres/captcha_execution.go:172` with `RecoveryUnknown` would satisfy the current observable assertions. The newly implemented positive participant classification is not proven.

Add a narrowly scoped observation through the real coordinator invocation that asserts:
- authenticated exact-owner evidence → `RecoveryCommitted`;
- corrupted ciphertext or foreign generation → `RecoveryUnknown`;
- exact-owner absence → `RecoveryNotCommitted`.

This does **not** require implementing public disclosure, restart recovery or commit-ambiguity execution.

#### Supported boundaries

Existing tests meaningfully prove SQL atomicity/rollback, retained replay ownership, non-newest alias contention, disjoint-profile serialization, secret-version continuity, exact-generation takeover, original alias/expiry conservation, populated migration rollback and restricted ownership/projection handling.

#### Unverified claims

Public issuer/disclosure/restart/commit-ambiguity executor, protected consumption, production audit/outbox interoperability, positive participant classification above, full coverage/mutation, provider and clean-consumer/release gates remain unverified.

### Accepted bounded correction followup

Yes. Extract the existing retained-row authentication/classification into an unexported deterministic helper, called by `recoverCaptcha` after sealed ownership validation and primary reads. Exercise it with actual persisted SQL rows: authentic → `RecoveryCommitted`; corrupt/mismatched → `RecoveryUnknown`; absent with authenticated owner → `RecoveryNotCommitted`.

Keep ownership admission in `recoverCaptcha`; the helper must grant no authority, perform no SQL, or become another recovery mechanism. Existing coordinator tests then prove the unsafe false-negative classification cannot terminalize authentic evidence.

No legitimate forwarding observer exists: `FeatureScope` requires the exact registered principal. My earlier observer suggestion was unsuitable. `postgres/recovery.go` also confirms committed and unknown currently produce identical aggregate outcomes.

**Original disposition retained: REQUEST_CHANGES**, narrowed to classification-oracle coverage—not a production defect or a requirement for public recovery implementation. The helper plus real-coordinator assertions is proportionate.

I inspected both updated test diffs; the marker-containing site adds meaningful populated-migration coverage. Its execution result remains pending.

### Fresh native regression report

- **Role/principal:** `regression_reviewer`, `/root/captcha_native_regression_9caf`
- **Reviewed state:** `feature/identity-platform`, exact base/HEAD `9caf047417b0dddbdebee03581ad9c27a296b999`, including all 20 current dirty tracked/untracked paths.
- **Boundary:** native CAPTCHA reservation/rebind, domain issuance, participant recovery, sealed ownership, scope-key migration and restricted core commitments.
- **Disposition:** **APPROVE for this local development checkpoint.**
- **Execution:** independent read-only inspection; no writes, tests, services or child agents.

**Previous finding resolved.** `classifyCaptchaRecovery` now preserves the existing production classification without introducing SQL or another authority. Tests read actual persisted rows and independently assert authentic → committed, corrupt/foreign binding → unknown, and owner-confirmed absence → not committed. The real coordinator assertions remain and prove unresolved histories cannot terminalize or transfer ownership.

The complete current audit found no additional concrete behavioral defects or blocking regression-oracle gaps. Existing tests cover atomic rollback, retained aliases, disjoint/non-newest-key contention, secret-version continuity, exact-generation takeover, original expiry conservation, populated migration failure atomicity and literal-marker preservation.

**Verification considered:** parent-reported current risk/postgres focused integration/race PASS `2.105s`, integration vet exit `0`; unchanged core/coordinator focused race/vet results retained.

**Not certified:** public issuance/disclosure/restart/commit-ambiguity executor, protected consumption, production audit/outbox interoperability, provider truth, complete coverage/mutation, clean-consumer or release gates. This approval supersedes my earlier `REQUEST_CHANGES` for the resolved classification-oracle gap only.

### Independent correctness review

- **Role/principal:** correctness reviewer, `/root/captcha_native_correctness_9caf`.
- **Method:** single-reviewer read-only pass using `$requesting-code-review`, its independent review, bounded-closure and advanced-boundary references, and `$testing-coverage`.
- **Disposition:** **APPROVE for this local development batch.**
- **Findings:** No findings.

#### Exact reviewed boundary

Repository: `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`
Branch: `feature/identity-platform`
Base and committed HEAD: `9caf047417b0dddbdebee03581ad9c27a296b999`
Comparison: that commit plus all 20 current dirty paths:

```text
postgres/CHANGELOG.md
postgres/captcha_reservation_test.go
postgres/captcha_ownership.go
postgres/captcha_ownership_test.go
risk/CHANGELOG.md
risk/README.md
risk/captcha_custody.go
risk/captcha_custody_test.go
risk/captcha_native.go
risk/postgres/CHANGELOG.md
risk/postgres/README.md
risk/postgres/captcha_command_test.go
risk/postgres/captcha_migrations_integration_test.go
risk/postgres/issuance_integration_test.go
risk/postgres/repository.go
risk/postgres/captcha_execution.go
risk/postgres/captcha_execution_integration_test.go
risk/postgres/captcha_repository_test.go
risk/postgres/captcha_reservation.go
risk/postgres/captcha_scope_migrations.go
```

Tracked binary diff SHA-256: `efabc0e4c7cd056a7415c73f3b1f18718e8cadf951629d675ac225318ae33e5a`.

Ordered `shasum -a 256` manifest of those 20 files, itself SHA-256: `c632d037eeebab5b554a62b1e3092916b78e2408b49e93351fc87e80f1350efe`.

#### Source-backed boundary closure

- **Ownership:** `postgres/captcha_ownership.go`, together with `postgres/feature_scope.go`, rejects zero, returned, foreign and undeclared handles. The recovery projection names the actual predecessor, rather than demanding a transfer that recovery has not performed.
- **Reservation and rotation:** `risk/postgres/captcha_reservation.go` serializes the exact tenant/provider scope before reading retained key identities. Retained-version omission, same-version secret substitution, competing disjoint configurations and alternate-command alias acquisition fail closed. Rebind conserves the original alias subset and transfers exactly the predecessor generation.
- **Domain atomicity:** `risk/postgres/captcha_execution.go` rechecks the primary winner and every original alias inside shared Work, samples database time after lock waits and inserts encrypted evidence without a private transaction. Domain rollback leaves reservation custody intact.
- **Recovery:** `postgres/recovery.go` excludes competing writers before invoking the participant. Missing/foreign ownership or corrupt encrypted responses remain unknown. Exact owner plus primary evidence absence supplies negative participant evidence; inconsistent positive evidence cannot terminalize the command as noncommitted.
- **Compatibility and migration:** `risk/captcha_native.go` retains historical signed bytes and binds new configuration commitments to original version-to-secret identities. The forward migration preserves existing aliases and enforces referenced-key retention; missing populated-history custody fails atomically. Migration input freezing is explicitly documented.
- **Oracle and scaling:** The changed integration assertions inspect independent primary row counts, generations, retained aliases, ciphertext and command state. Runtime retention lookup is capped at 17 rows; alias work is bounded by 16 aliases and selected profiles by 32. The migration backfill remains whole-history work, not a demonstrated production-scale rollout.

#### Falsification challenges

1. **Changed secret under the same retained version:** new signed preparation can authenticate against its selected profile, but persisted scope commitment comparison still rejects reservation.
2. **Old pending command after a newer key is added:** authenticated retained custody covers the enlarged scope history while rebind retains only the old command’s aliases and expiry.
3. **Recovery generation advance without participant transfer:** the participant compares the persisted winner against `PreviousGeneration`; missing or foreign generations remain unknown.
4. **Partial participant effect beside a pending command:** authentic encrypted evidence yields positive participant evidence, which the coordinator refuses to reinterpret as unanimous noncommit.

These traces did not expose a changed-contract defect.

#### Verification and claim limits

Supplied focused race tests and vet results were accepted as development evidence; I did not execute tests, builds, services or mutations.

This approval does **not** certify package completion, release readiness, provider interoperability, public executor/disclosure/restart workflows, commit ambiguity, protected consumption, production-shaped migration/query plans, full coverage/mutation gates, dependency pins or clean-consumer gates. Those remain explicitly unfinished goal work.

## Native participant key continuity and primary recovery checkpoint

Execution revision9caf047; all current dirty coordinator/risk/risk-postgres
source included. Go1.27.1, disposable PG18.6 pinned image
`sha256:5a5a84b19854a9ffaa54082c166ff4ec27473a361e496e5ea167f298f2da9722`.
Temporary immutable primitive wiring is development evidence, not clean consumer.

Focused RED observations: missing ReplayKeyCustody API (8685), changed replay
secret restored current signed custody (47470), actual same-version replacement
secret admitted another SQL winner (66206), original owner takeover after key
addition denied (84014), exact absent participant recovery incorrectly unavailable
(12384). Each regression now passes at the same observable boundary.

Final focused commands (race) and native terminal outcomes:
- risk: `go test . -race -run '^TestCaptcha' -count=1`, PASS1.164s;
  `go vet .`, exit0 (38162).
- coordinator/postgres: same focused command, PASS1.289s; `go vet .`, exit0
  (11675).
- risk/postgres: `go test . -race -tags integration -run
  '^TestCaptcha|^TestIssuedEvidenceSameWorkIntegration$/native-captcha$' -count=1`,
  PASS3.858s; `go vet -tags integration .`, exit0 (47565).

Actual SQL additionally checks eight disjoint-key races, non-newest-key overlap,
version-to-secret continuity, populated migration/incomplete backfill rollback,
retained-key additions conserving original aliases/expiry, and primary recovery
of absent versus authentic/corrupt/foreign-generation retained participant state.
Effects use existing test seams, not production audit/outbox interoperability.
Public issuance executor, disclosure/restart/commit-ambiguity recovery, protected
consumption, provider truth and full release gates remain unfinished. No unit
promotion or completeness claim. Independent native reviews pending.

## Native schema review checkpoint

Source base625df1c; committed `9caf047417b0dddbdebee03581ad9c27a296b999`,
exactly five risk/postgres schema/test/documentation paths. Execution observations
are retained below under next-native-schema work; no runtime result relabelled
as having executed after commit. New native implementation begins after9caf047.

### Native schema regression report

**APPROVE — regression_reviewer**

- **Native principal:** `/root/captcha_schema_regression_625d`.
- **Reviewed boundary:** source base/head `625df1c62f9cab80df015f8066f4e8e024e260e4` plus complete dirty `risk/postgres/{CHANGELOG.md,README.md,captcha_migrations.go,captcha_migrations_test.go,captcha_migrations_integration_test.go}`.
- **Method:** applied `$testing-coverage`; read-only source, tests and authoritative CAPTCHA requirements inspection. No writes, execution, services or children.
- **Findings:** no concrete regression-test defect requiring changes.

The SQL oracle directly checks same-scope alias collision between two valid winners, retention of two key versions for one alias, tenant isolation, rejection of dual-target evidence, anonymous evidence insertion, continued collision after evidence deletion, referenced-winner deletion rejection and preservation of both retained key versions. Unit checks cover independent allocation, forward-only/default-transaction metadata, defensive ownership, invalid input and cancellation.

**Claim limits:** populated migration coverage preserves a pre-existing kernel command, not populated prior risk evidence or actual migration-engine history. Positive user evidence insertion, neither-target rejection and other provider/site/profile/configuration scope partitions are not exercised. These limit broader verification claims; inspection found no corresponding defect in this additive schema slice.

Parent-supplied focused unit, PostgreSQL, race and vet results were considered but not independently rerun. Native executor, protected consumption, clean-consumer composition and complete release/coverage/mutation gates remain uncertified.

### Native schema compatibility report

**Native principal:** `go_compatibility`
**Disposition:** `APPROVE` — scoped schema checkpoint; no concrete compatibility findings.

Reviewed base/head `625df1c62f9cab80df015f8066f4e8e024e260e4` plus all five dirty `risk/postgres` paths.

- Additive `CaptchaMigrationSource` implements the existing migrations interface without changing existing migration histories, module manifests, dependency pins, or Go 1.27 baseline.
- Independently selected forward migration requires only kernel and existing risk namespace; phone/counter/lockout upgrades are not bootstrap dependencies.
- Replay alias uniqueness excludes key version; its separate version relation preserves multiple retained versions sharing a fingerprint.
- Evidence deletion leaves replay ownership and referenced replay-key versions intact.
- Production `Repository.FeatureOperations` remains unchanged and does not select CAPTCHA execution. Documentation accurately excludes runtime issuance, protected consumption, cleanup, and retirement.

**Unproven boundaries:** released-pin/clean-consumer compilation, migration-runner integration/history allocation, native executor/recovery, protected-use interoperability, coverage/mutation and remaining release gates. Parent-reported focused PostgreSQL/race results were not independently executed or certified here.

Reviewed source SHA-256: `captcha_migrations.go` = `594b6b50aefc19639ecc795e6874636e0cb834a464438714cf1234328d2ab213`.

No writes, tests, services, or child agents used.

Native agent identity for the preceding report: `/root/captcha_schema_compat_625d`.

### Native schema complete-diff correctness report

**APPROVE — bounded additive schema checkpoint only. No findings.**

Native principal: `/root/captcha_schema_correctness_625d`, independent read-only correctness reviewer. Applied `$requesting-code-review`, required causal/boundary references, `$testing-coverage`, `$testing-integration`, and `$verification-before-completion`. No writes, tests, services, external operations, or children.

**Exact boundary:** `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`, branch `feature/identity-platform`, base/head `625df1c62f9cab80df015f8066f4e8e024e260e4`, plus the complete five-path dirty change:
`risk/postgres/{CHANGELOG.md,README.md,captcha_migrations.go,captcha_migrations_test.go,captcha_migrations_integration_test.go}`.

Source-backed closure:
- Independently allocated forward migration creates only new tables/indexes; existing histories and rows are unchanged. Existing namespace/kernel are the only schema dependencies.
- Alias uniqueness uses the exact prescribed scope and excludes key version. Separate version relations conserve multiple retained versions sharing one fingerprint without admitting another winner.
- Composite foreign keys preserve tenant/provider/site/profile/configuration ownership. Kernel linkage inherits nonzero command identity and prevents premature command retirement.
- Evidence has exclusive user/anonymous targets, bounded payloads/digests/versions, finite timestamps and maximum two-minute validity.
- Evidence deletion cannot cascade into replay ownership or referenced versions. Referenced winner deletion is rejected.
- Index widths remain bounded by canonical core limits; parent-key and retirement indexes support the introduced relationships.
- Existing mapping, restoration and encryption contracts agree with storage bounds; no runtime consumer or selection is introduced.

Bounded falsification challenges covered competing commands sharing an alias, retained versions sharing one fingerprint, foreign-tenant reuse, dual-target insertion, and evidence deletion followed by replay. Constraints and independent SQL assertions establish the intended outcomes.

**Verification:** reviewer `git diff --check` passed; head, dirty scope and five checksums remained unchanged. Parent-reported missing-API RED, focused unit PASS, PostgreSQL 18.6 SQL PASS, final migration unit/SQL race PASS1.710s and corrected vet exit0 were assessed against their actual test oracles, not independently rerun.

**Not certified:** migration-runner recovery, native executor/reservation/commit reconciliation, protected consumption, cleanup/key-retirement execution, provider truth, clean-consumer composition, exact coverage/mutation, or release/platform completion. These remain later goal work; this approval permits only the reviewed local development checkpoint.

## Native participant work in progress after9caf047

Uncommitted source: selected immutable profile authentication, real all-alias
reservation/rebind, sealed domain/recovery fingerprint+generation linkage,
same-Work encrypted evidence and participant reconciliation. Original mapper/QA
still applies. Native Config/restore/reservation missing APIs RED85014;
missing domain owner API RED11826; selected schemas RED48933. Pure native profile
GREEN0.237s33568. Domain ownership GREEN0.294s89316 after correcting a zero-sized
test principal (Go permits coalesced addresses; production owner is nonzero).
First real SQL run exposed a test oracle using string `succeeded` instead of
audit's canonical numeric outcome1; corrected from upstream canonical DTO and
existing native SQL oracle. Actual native SQL GREEN0.552s21610; exact-generation
takeover/stale denial GREEN0.881s29322. Final focused coordinator CAPTCHA race
PASS1.282s31460/vet0; focused risk/PostgreSQL CAPTCHA unit+real SQL race
PASS2.225s46491/integration vet0. Go1.27.1, disposable caches and temporary exact
primitive source wiring. No clean-consumer/release/native executor claim.

Required NEXT before committing native selection: close replay-key omission and
disjoint-key concurrency bypass (all referenced versions must remain represented
under serialized scope custody), prove multikey/non-newest contention and
historical participant ambiguity, implement current-authorized public executor
and original-response recovery. Do not certify the current native batch or reuse
the initial schema review as its correctness review. Existing committed schema
history must remain byte-identical; any scope/key custody schema correction needs
an additive forward migration, not rewriting9caf047. Protected consumption,
cleanup/retirement, canonical service/providers and full release obligations stay
open. Current tests prove bounded positive/rollback/takeover paths, not rotation
safety or complete CAPTCHA.

## Callback-free startup custody checkpoint

### Next native schema work in progress

Source625df1c plus five risk/postgres paths: README/CHANGELOG and new
`captcha_migrations{,_test,_integration_test}.go`. Primary classification:
persistence; secondary public-contract/security/replay/migration. Existing
whole CAPTCHA mapper/QA applies. Missing migration API compile RED56021;
`go test . -run '^TestCaptchaMigration' -count=1` GREEN0.277s38309.
Actual PostgreSQL18.6 schema test GREEN0.561s15115; final
`go test . -race -tags integration -run '^TestCaptchaMigration' -count=1`
PASS1.710s28421. A following incorrectly ordered vet invocation failed as an
import-path syntax error; corrected `go vet -tags integration .` exit0 95815.
Go1.27.1; task-owned disposable caches, same temporary explicit primitive source
wiring as preparation. Real SQL proves populated command preservation, global
alias winner across retained versions, tenant separation, exclusive target
namespace, alias survival after evidence deletion and referenced key retention.
This does not prove native reservation/executor or protected-use behavior.
Task-labelled container90725611 and its volume removed. New code remains
uncommitted pending independent reviews; not a readiness or unit-completion claim.

Source base `032f0ded3be487a62effa5859ebea05daf43d6af`; local commit
`625df1c62f9cab80df015f8066f4e8e024e260e4`, exactly five risk paths.
Missing API RED session19641; focused custody GREEN0.427s session88904.
Final complete risk `go test . -race -count=1` PASS1.389s session16472;
`go vet .` exit0 session53840. Go1.27, explicit temporary immutable primitive
source wiring, not clean-consumer proof. Both execution scratch directories
removed. Original tracked-patch SHA256
`eaf90037853e5f64e5e555f17c9d0ffba53cba880a01d6252793cb252f7c77ca`;
new custody source SHA256
`31791991fc1d75b9bc60d4171960f49005d8568c53b1bfbbbedc5a51a642a600`;
new custody tests SHA256
`caaa73e604775a69c63916f12cb1f59379e42bc6f16094b07f0fce0f23fbe2cc`.

### Native custody regression report

**Role:** `/root/captcha_custody_regression_032f`, fresh read-only `regression_reviewer`.

**Reviewed boundary:** base/head `032f0ded3be487a62effa5859ebea05daf43d6af`, plus all five dirty paths:
- `risk/CHANGELOG.md`
- `risk/README.md`
- `risk/captcha_issuance.go`
- `risk/captcha_custody.go`
- `risk/captcha_custody_test.go`

**Disposition: APPROVE.** No concrete regression or material test defect found in this bounded custody change.

The tests meaningfully establish callback-free construction, absence of retained runtime resources, typed-nil executor rejection by the executable issuer, unchanged static validation, owned configuration copies, complete order-independent evidence-key matching, original signed candidate conservation, exclusive expiry, selected-profile rejection, and serialization/diagnostic redaction.

Existing native restoration tests cover the delegated implementation's signature tampering, retained-key rotation, missing historical keys, immutable projections, concurrent copies, changed configuration, and recovery without reverification or reissuance. The constructor extraction leaves executable runtime requirements intact.

Execution evidence is parent-reported: missing-API RED, focused GREEN, and whole-risk race PASS. I ran no tests or services. This approval does not establish clean-consumer/release gates, provider truth, production SQL/executor behavior, durable commitment, or protected-use authority; those are outside this batch.

### Native custody compatibility report

**APPROVE — `go_compatibility`, principal `/root/captcha_custody_compat_032f`.**

Reviewed base/head `032f0ded3be487a62effa5859ebea05daf43d6af` plus all five dirty files: `risk/{CHANGELOG.md,README.md,captcha_issuance.go,captcha_custody.go,captcha_custody_test.go}`.

- Additive public API; existing executable issuer still requires clock, verifier, executor and rejects typed-nil executors. Shared static validation preserves existing bounds.
- Custody drops runtime callbacks before construction, owns copied configuration and introduces no mutable registration cycle.
- Restoration delegates to unchanged native authentication: original evidence key, policy/configuration commitment, retained replay versions, aliases, nonce and exclusive expiry remain preserved.
- Complete retained evidence-key comparison is order-independent; constructor uniqueness plus equal cardinality prevents duplicates concealing missing keys.
- No module/dependency changes; Go `1.27.0` baseline remains unchanged.

No concrete compatibility defects found. Downstream adapter consumption, clean-consumer/source-pin adoption and release gates remain unproven; this approves the local core prerequisite only. No commands executing code, tests, services or writes performed by this reviewer.

### Native custody complete-diff correctness report

**APPROVE — bounded local development batch only. No findings.**

Native principal: `/root/captcha_custody_correctness_032f`; fresh independent read-only correctness reviewer. Applied `$requesting-code-review`, its causal/boundary references, and `$testing-coverage`. No writes, tests, services, external mutations, or child agents.

**Reviewed boundary:** `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`, branch `feature/identity-platform`, exact base/head `032f0ded3be487a62effa5859ebea05daf43d6af`, plus all five dirty paths:
- `risk/CHANGELOG.md`
- `risk/README.md`
- `risk/captcha_issuance.go`
- `risk/captcha_custody.go`
- `risk/captcha_custody_test.go`

Inspected complete diff, constructor callers, native restoration, materialization/recovery consumers, fixture inputs, and current risk/CAPTCHA goals.

**Source-backed closure:**
- Executable construction retains its original runtime requirements, including typed-nil executor rejection. Custody clears runtime resources before shared validation and retains no callback, clock, or executor.
- Shared construction still copies mutable slices and score policy; replay/policy handles already own immutable state.
- Exact retained-key comparison rejects foreign/missing/duplicate keys: constructor-enforced uniqueness plus equal cardinality and membership establishes set equality, independently of order.
- Restoration delegates to the existing authenticated profile/policy/key path. Original signed aliases, key version, nonce, candidate and exclusive expiry remain unchanged; retained newer keys cannot rewrite them.
- Restoration grants no commitment or protected-use authority. Native intents remain issuance-ineligible; runtime recovery still requires the executable issuer's ownership.
- Formatting redacts custody; JSON/text serialization fails closed. Comparisons are bounded by the retained-key limit, with no I/O or shared-state mutation.

Bounded falsification challenges included duplicate-for-missing key substitution, typed-nil runtime wiring, changed verifier profiles, caller mutation after construction, and materialization exactly at expiry; source paths deny or preserve the required invariant.

**Verification:** `git diff --check` passed. Parent-reported missing-API RED, focused custody PASS0.427s, whole-risk race PASS1.389s, and vet exit0 were assessed against their actual oracles; not independently rerun. Five file checksums remained unchanged through review.

**Not certified:** native SQL/executor integration, durable replay ownership or commitment, protected consumption, clean-consumer composition, exact coverage/mutation, provider interoperability, remaining release gates, or complete CAPTCHA/platform delivery. These remain explicit later goal obligations, not defects in this additive prerequisite.

### Native custody exact-head delivery report

**APPROVE — bounded local committed core custody checkpoint only. No findings.**

Native principal: `/root/captcha_custody_delivery_625d` (`delivery_auditor`).
Exact head: `625df1c62f9cab80df015f8066f4e8e024e260e4`.
Verified direct parent: `032f0ded3be487a62effa5859ebea05daf43d6af`.

- Read the retained native regression, compatibility, complete-diff correctness, mapper and QA reports directly. Custody reviewers independently report `APPROVE`.
- Commit contains exactly the five reviewed risk paths. Reproduced the complete three-path patch hash `eaf90037853e5f64e5e555f17c9d0ffba53cba880a01d6252793cb252f7c77ca`, custody source hash `31791991fc1d75b9bc60d4171960f49005d8568c53b1bfbbbedc5a51a642a600`, and custody test hash `caaa73e604775a69c63916f12cb1f59379e42bc6f16094b07f0fce0f23fbe2cc`. All match retained reviewed inputs.
- Worktree was clean at audit; no subsequent adapter edits observed. Committed diff check passed.
- Retained race `PASS1.389s` and vet exit `0` are original base-plus-dirty execution evidence, committed unchanged—not new executions.
- Both custody execution scratch directories are absent. Next-adapter archive directory `/private/tmp/identity-captcha-store.rOayDm` exists and is not a leftover custody execution cache.

**States:** local and committed verified; push not established, no remote operations performed; CI, deployed and production unverified. No clean-consumer, full-unit readiness/completion, coverage/mutation or release certification. All broader goal obligations remain active.

Source base: `28fc29aaeee1d4222d6f40808322fb9b322d0c86`.
Committed source: `032f0ded3be487a62effa5859ebea05daf43d6af`.
Branch: `feature/identity-platform`, source worktree
`/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`.
Reviews inspected the complete 18-path dirty boundary before commit.
Sorted committed 18-path file-checksum output SHA256:
`a79ce1f7815c617f7c12afd066f573655093431fd63ad7bd80976cb7c9b295ae`.

## Execution observations

- Native mapping missing API RED, then all six flow variants and strict custody
  negatives GREEN. Missing `sealCaptcha`/`openCaptcha` RED (34423 and refreshed
  e0d2f1); exact-original and cross-scope cipher negatives GREEN0.276s.
- Missing coordinator declaration/handle APIs RED (b3bd04); declaration,
  historical codec, participant and generation/lifetime tests GREEN0.221s.
- PostgreSQL independent SQL fixture proves reservation ownership persists
  across domain rollback, denies another command, rebinds exact generations,
  and rolls back callback errors/panic/open rows/poison/transaction closure.
  Focused integration PASS0.382s. This is coordinator-boundary proof, not the
  unfinished production risk participant.
- `postgres`: `go test . -race -tags integration -run
  '^TestPostgresIntegration$/^(captcha_replay_reservation|feature_observation|risk_reservation)$'
  -count=1` PASS8.433s (session72838). Disposable PostgreSQL18.6 image
  `postgres@sha256:5a5a84b19854a9ffaa54082c166ff4ec27473a361e496e5ea167f298f2da9722`.
  Container and its owned volume removed.
- Final unchanged source/test tree: `postgres` `go test . -race -count=1`
  PASS1.819s and `go vet .` exit0 (98694); `risk/postgres` same commands
  PASS1.284s and vet exit0 (98767). Initial alternate disposable workspace
  layout failed offline module resolution; original explicit-source wiring
  restored and passed. No permanent manifests/pins changed.
- Go1.27 target. Tests used owned disposable caches and immutable archived
  primitive sources. Both scratch directories were removed. No clean-consumer,
  exact coverage/mutation, provider truth, aggregate release or full-unit claim.

## Native regression report

The fresh pre-write mapper `/root/captcha_pg_map_28fc` and QA planner
`/root/captcha_pg_qa_28fc` inspected source28fc29a/control4af530e, read-only,
default `gpt-6.1-sol` medium with isolated context. The mapper's followup
corrected its preliminary claim that no shared primitive was missing:

> **No: head `28fc29aaeee1d4222d6f40808322fb9b322d0c86` exposes no public native-feature reservation phase for CAPTCHA issuance aliases.** My earlier “no missing shared primitive” conclusion was too broad; this reservation-before-domain requirement exposes a genuine missing shared boundary.
>
> Source evidence:
>
> - `postgres/coordinator.go:293–364`: the reservation transaction locks the command claim, reserves its journal, freezes the plan, then invokes only `reserveCommandCounters`, `reserveRisk` and `reserveCapability` before committing.
> - `postgres/feature_execution.go:14–31`: the native feature contract offers pure `ValidateFeatureCommand` and domain-transaction `ApplyFeature`; no feature reserve/rebind method.
> - `postgres/plan.go:24–55`: `CommandPlan` has no CAPTCHA-issuance or generic feature-reservation declaration.
> - `postgres/plan_risk.go:11–61`: `RiskEvidencePlan` means consumption of existing signed evidence and admits only the exact phone-reset-initiation schema.
> - `postgres/risk_reservation.go:14–29`: its reservation contract expressly locks an **already-issued** row and forbids insertion of missing evidence.
> - `postgres/command_counter.go:49–63`: the counter contract increments/records a protected attempt and reuses its original snapshot; it is not an issuance-replay ownership contract.
>
> Thus domain-only alias insertion cannot establish durable unresolved ownership: rollback removes those aliases while the independently committed command remains pending. Reusing phone-risk/counter contracts would misrepresent their frozen semantics.
>
> **First missing boundary:** a coordinator-owned, explicitly declared native CAPTCHA issuance reservation/rebind phase, invoked inside the existing command reservation transaction, with invocation-bound access to the exact frozen command, authoritative fingerprint and current/prior owner generations. It must persist/revalidate the original alias winner independently of later domain rollback, and preserve that ownership through pending reconciliation and takeover. No private transaction is warranted.

Native QA artifact:

> - **Principal:** `/root/captcha_pg_qa_28fc` — fresh read-only `qa_test_planner`; `$qa-test-planner` read fully.
> - **Source base/head:** `28fc29aaeee1d4222d6f40808322fb9b322d0c86`.
> - **Control head:** `4af530ee0058c1302d9df111ea911a474773e758`.
> - **Disposition:** **Implementation may proceed.** No contradictory requirement, unsafe requested action, or reproducible consumed-contract defect identified.
> - **Scope:** durable CAPTCHA issuance and original-result recovery only. This does not certify protected consumption, canonical risk service, provider adapters, or complete package verification.
> - **Execution:** read-only inspection; no writes, tests, services, or child agents.
>
> ### Observable contract
>
> A successful response exposes exactly the originally committed opaque CAPTCHA reference, provider, original database-issued time, original exclusive expiry, and `MaxUses=1`. Evidence, command result, audit, and lifecycle persistence succeed or roll back together through the existing coordinator.
>
> Unknown commitment exposes no reference and preserves the original command/aliases for primary reconciliation. Recovery never verifies the provider again, manufactures another candidate, renews expiry, or treats signature validity as commitment/current authorization.
>
> ### Smallest sufficient focused verification
>
> | Boundary / concrete risk | Observable assertion |
> |---|---|
> | **Native mapping and strict restoration:** omitted, duplicated, unknown, wrong-kind, oversized, altered or cross-schema fields bypass authentication | Round-trip preserves every signed scalar and original alias ordering. Individually altered structurally valid fields fail with stable redacted errors and zero custody. Duplicate/extra/missing fields and invalid widths fail before durable work. |
> | **Closed flow variants:** anonymous target becomes fake user; target confused with caller | Exercise pre-auth subject, pre-auth anonymous, authenticated subject/session, administrator subject and anonymous target, with and without permitted administrator session. Persist/query/audit exact target namespace. Anonymous `UserID` stays empty; authenticated real service actor remains distinct. Cross-namespace lookalikes, actor substitutions and wrong-session/pre-auth/challenge/request bindings deny without disclosure. |
> | **Real PostgreSQL success and rollback:** reference escapes before atomic commitment | Public issuer success corresponds to one committed evidence row, original encrypted response, command result, protected audit and lifecycle records. Inject failure after evidence staging and at audit/lifecycle persistence; prove absence of all committed effects and zero exposed reference. Include one real owned audit/outbox composition, not only effect seams. |
> | **Multikey replay winner:** rotation, alias expansion or concurrent commands produce duplicate authority | Race independent PostgreSQL connections with overlapping retained key sets and differently ordered startup configuration, including overlap on a non-newest key. Exactly one command owns the token; losers expose no reference. All retained aliases resolve to that winner. Concurrent additions preserve previous aliases/tombstones rather than replacing them. Different commands with the same token remain denied after original expiry. |
> | **Unknown commitment and cancellation:** timeout interpreted as rollback; pending history permits reverification | Cover committed-but-acknowledgement-lost and proven rollback separately, using an actual PostgreSQL failure boundary or existing commit-ack seam with SQL-confirmed outcome. Unknown returns no reference; same-command primary recovery returns original commitment or proved failure. Pending/inconsistent participant evidence remains unknown. Cancellation before submission, during lock wait, and after possible commit neither leaks a candidate nor authorizes fresh issuance. |
> | **Restart and historical result:** new keys or time produce replacement bearer | Discard original issuer/executor instances and recover using restored native custody with retained keys. Byte-identical reference, issue time, expiry and alias set return; provider callback must fail immediately if invoked. Recover both before and after expiry. New materialization at exactly expiry is rejected; historical committed recovery remains nonrenewing. |
> | **Current authorization:** historical authorization discloses a sensitive original response | Admission and recovery policies are independently exercised. Current denial/revision mismatch, foreign tenant/target/caller, unauthorized service actor, and scope substitution disclose no reference even for committed history. Authorized anonymous recovery requests the anonymous-flow resource, never a user alias. |
> | **Ciphertext/privacy and populated migration:** transplant, corrupt custody or migration silently resets replay protection | Swap ciphertext between valid commands/tenants/flows and alter retained metadata; decryption fails closed. Missing retained key and corrupt/oversized envelope expose no plaintext or fresh reference. Migrate populated prior identity/risk state; existing commands and custody remain recoverable, old replay winners remain blocked, and new issuance works. Inspect persisted command/result/audit/lifecycle values for absence of token, raw provider facts, scores, remote IP, `cdata`, and plaintext bearer. |
>
> ### Independent oracle and reduction
>
> Use public issuer responses plus independent SQL reads of primary committed state, ownership, aliases and side effects—not adapter return values alone. Compare recovery against captured original bytes/times. Use the existing pinned replay-framing vector for digest derivation, rather than deriving expected values through the implementation under test.
>
> Table-driven unit negatives can cover all closed metadata fields and all four provider discriminators. Real PostgreSQL flow success/recovery must cover each distinct subject/actor namespace; contention and ambiguity need not multiply across providers because provider networking is outside this slice. Existing core native/replay tests are relevant regression inputs, not substitutes for adapter SQL/coordinator proof.
>
> Remaining full-goal obligations—reserve/apply/finalize, cleanup/erasure lifecycle, official provider behavior and complete release gates—remain unverified and must not be claimed by this slice.

The reports below approve only the completed preparation/coordinator slice;
the QA's remaining production adapter cases stay implementation work.

**Native role:** `regression_reviewer`
**Disposition:** `APPROVE` for this development batch only.

**Reviewed boundary:** `feature/identity-platform` at base/HEAD
`28fc29aaeee1d4222d6f40808322fb9b322d0c86`, including all 18 current
tracked/untracked dirty paths and their affected core/coordinator contracts.
Read-only inspection; no tests or services executed by this reviewer.

No concrete behavioral defects or blocking regression-proof gaps found within
the assigned boundary.

- Mapping/restoration tests exercise all six flow variants, preserve signed
  custody, and reject altered fields, field kinds, missing/unknown fields,
  foreign keys/policy and substituted administrator actors.
- Encryption tests prove exact original-response recovery and rejection after
  metadata, tenant, actor or flow-scope alteration; CAPTCHA uses a separate
  envelope profile.
- Reservation tests cover explicit mandatory admission, frozen-plan
  persistence, distinct anonymous/user namespaces, exact participant identity,
  predecessor generations and expired handle denial.
- Real-PostgreSQL fixture tests exercise independently committed alias custody
  surviving domain rollback, denial of another command, exact-generation
  takeover, and atomic rollback after callback errors, panic, leaked rows,
  poisoned SQL or attempted transaction closure.
- Unselected legacy plans retain their existing encoding paths; the added
  phase changes encoding only when explicitly selected.

Parent-reported focused unit/race and PostgreSQL integration runs passed.
Those results support this bounded review, not release certification.

**Not certified:** production native CAPTCHA SQL store/executor,
original-result reconciliation, protected consumption, full CAPTCHA completion,
or aggregate release gates.

Native principal: `/root/captcha_reservation_regression_28fc`.

## Native compatibility report

**APPROVE — bounded development batch only.**

Principal: `/root/captcha_reservation_compat_28fc` (fresh, read-only
compatibility review).

Reviewed: `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`,
`feature/identity-platform`, base/head
`28fc29aaeee1d4222d6f40808322fb9b322d0c86`, plus all current
tracked/untracked CAPTCHA mapping, ciphertext, reservation, tests and
documentation changes.

No concrete compatibility defects found:

- New public mapping and reservation contracts are additive; existing
  contributor interfaces remain unchanged.
- Unselected historical CBOR plan profiles retain their existing encoding
  path. Existing pinned historical fixture coverage remains present.
- User and anonymous-flow schemas remain distinct; mapping binds signed
  target scope and administrator actor without inventing a user.
- CAPTCHA response ciphertext has a separate authenticated profile from
  phone-risk evidence.
- Independent module boundaries remain intact, with no new dependency edges,
  replacements or manifest changes. Existing modules retain Go `1.27.0`.
- Production `Repository.FeatureOperations` does not select CAPTCHA, so this
  preparation batch does not silently enable incomplete durable execution.

**Pending release boundaries:** unchanged dependency pins do not establish
clean-consumer compatibility with these new APIs. Parent-reported
temporary-source-wired tests establish selected local behavior only. Native
CAPTCHA SQL/execution/recovery, protected consumption, release gates and
whole-unit completion remain unproven and explicitly unfinished.

No tests, services, writes or external mutations performed by this reviewer.

## Native complete-diff correctness report

**Internal disposition: APPROVE — this local development batch only. No findings.**

Principal: `/root/captcha_reservation_correctness_28fc`, fresh read-only
correctness reviewer. Applied `$requesting-code-review`, its required
causal/boundary references, and `$testing-coverage`; no writes, tests, services,
children, or external mutations.

Reviewed repository:
`/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`, branch
`feature/identity-platform`. Exact base/head:
`28fc29aaeee1d4222d6f40808322fb9b322d0c86`, plus **all 18 dirty paths**,
including seven untracked CAPTCHA source/test files. Complete literal diff and
affected causal slice inspected.

Boundary fingerprints:
- Tracked patch SHA-256:
  `d15546a1c1d157e05aef67b77e7087018a2d17d98cc617428a6b4cbccb7f1bc0`
- Ordered 18-path `shasum -a 256` output SHA-256:
  `db009f46de37554198495a4bd204c28dcb8c587e2045e3e0d6c56000a582c3ee`

Source-backed closure:
- `risk/postgres/captcha_command.go:31,75`: mapping conserves signed original
  custody; restoration rejects missing, altered, extra, wrong-kind,
  wrong-width, wrong-policy/key, and foreign-scope input. User and
  anonymous-flow namespaces remain distinct; administrator actor identity
  remains signed.
- `risk/postgres/captcha_cipher.go:17,41,68`: separate encrypted-response
  profile authenticates exact binding and retained metadata; bounded
  payload/time/hostname handling and plaintext clearing match the preparation
  contract.
- `postgres/captcha_reservation.go:47,102,140`: only explicit exact CAPTCHA
  schemas admit reservation; callback handles bind the registered participant,
  retained authoritative fingerprint, and exact predecessor generation.
  Return/panic invalidates copies; failures and outstanding rows deny
  reservation.
- `postgres/coordinator.go:351` and existing journal/Work ownership: alias
  callback runs inside the original reservation transaction before its commit;
  domain rollback cannot erase that committed reservation. Changed-input reuse
  is fenced by claim/frozen-plan identity.
- `postgres/plan_feature.go:123` and `postgres/plan_feature_decode.go:97`:
  declaration survives admission, snapshots, canonical serialization, and
  historical decoding; unrelated profiles remain unchanged when unselected.

`git diff --check` passed. Supplied RED/GREEN, whole-unit race, and focused
PostgreSQL integration results were assessed against their test oracles, not
independently rerun.

**Not certified:** production native CAPTCHA SQL storage/executor,
original-result recovery, protected consumption, clean-consumer composition,
exact coverage/mutation and remaining release gates, full-unit completion, or
release readiness. The independent SQL fixture proves the coordinator boundary,
not the unfinished production risk participant.

## Native exact-head delivery report

**APPROVE — bounded local committed preparation/reservation batch only.**

Native principal: `/root/captcha_reservation_delivery_032f`
Source head: `032f0ded3be487a62effa5859ebea05daf43d6af`
Base: `28fc29aaeee1d4222d6f40808322fb9b322d0c86`
Control head: `4af530ee0058c1302d9df111ea911a474773e758`

- Source worktree is clean; commit contains exactly the reviewed 18 paths.
- Inspected native regression, compatibility and complete-diff correctness
  reports, each internally `APPROVE`. Reconstructed tracked-patch fingerprint
  `d15546…` and tracked-then-added 18-path fingerprint `db009f…` match the
  correctness report exactly. Sorted committed fingerprint `a79ce1…` also
  matches. No post-review source/test divergence.
- Legacy unselected encoding paths remain unchanged; pinned historical fixture
  coverage remains present. Production risk `FeatureOperations` does not select
  CAPTCHA.
- Supplied unit/race/vet and selected PostgreSQL integration results support
  this batch. The independent SQL participant proves coordinator reservation
  persistence, competing-owner denial, exact-generation takeover and callback
  rollback—not production native CAPTCHA storage/execution.
- Both specified scratch directories are absent. Specified container and
  task-labelled containers/volumes are absent.
- Control worktree contains only the execution-ledger modification and new
  native-review evidence file.

**Delivery states:** local and committed verified; push reported not performed,
remote state not independently queried; CI, deployment and production
unverified. No full CAPTCHA, full-unit, clean-consumer, coverage/mutation or
release-completion certification. No tests or services started by this auditor.
