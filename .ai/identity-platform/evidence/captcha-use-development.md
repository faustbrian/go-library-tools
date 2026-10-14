# CAPTCHA protected-use development

Primary classification: public-contract; secondary behavioral, security,
persistence, concurrency and replay. Source base d1df7dee0a5627a8d6bb568be165fd3fce5f65a8.
This work advances the complete consuming chain, starting with authenticated
core use custody. That prerequisite is not durable reservation/consumption.
Current owner: primary agent; all delegated roles are read-only.

## Coordinator consuming-plan development

Next development slice: typed CAPTCHA reservation/rebind through the same
coordinator journal transaction, preserving exact signed plan and command;
reject incomplete/foreign participants, callback escape, failed SQL/open rows,
panic/cancellation, wrong predecessor and copied handles after return. Real
primary tests must prove command/evidence ownership transfers or rolls back
together. Domain execution and noncommit release stay closed until their native
typed contracts are wired; no time-based release is introduced.

Source d10dc0b plus consuming-reservation development. Missing interface/handle
REDaf7f29. All `TestCaptchaUse` focused tests PASS0.235s (02b616).
Real isolated PG18.6 `go test . -tags integration -run
'^TestPostgresIntegration$/captcha_use_reservation' -count=1` PASS0.955s
(8b8e19). Initial PG fixture failed pure native validation because it supplied
codec fields rather than the selected executable fixture schema; corrected
fixture, no weakened production admission. Loopback task-owned cluster
identity-captcha-phases-5osxbp, image
sha256:5a5a84b19854a9ffaa54082c166ff4ec27473a361e496e5ea167f298f2da9722,
anonymous volume f0496d798e575538956fcaa00ad6397240f8b9a1e78614adc1acbe9044f01fa0.
Scratch /private/tmp/identity-captcha-phases.5osXbp contains exact primitive
archives only for this bounded batch; command caches removed by traps.
Native domain checks/finalize/release/recovery and the actual risk-store
implementation remain pending; independent SQL fixture is not cryptographic
or production protected-workflow proof.

Current twelve-path source manifest
972b49f0a654ff6668a5b7b3655bba46ecff3bad3a7eb9a9cc8df113cef8289c.
Owned counterfactual omitting the coordinator consuming reservation callback
caused actual PG test failures (bb74ad): pending command survived while evidence
stayed issued; error/panic/open-row/poison modes received ownership because
their native callbacks were bypassed. This proves the integration assertions
detect absent wiring, not full mutation efficacy. Final current affected
`go test . -race -tags integration -run
'^TestCaptcha|^TestPostgresIntegration$/(captcha_use_reservation|captcha_replay_reservation|risk_reservation)$'
-count=1` PASS8.394s and `go vet -tags integration .` exit0 (8d87c1).
Mutant tree/workfiles and caches removed by trap. Complete final review assigned to
/root/captcha_use_reservation_final_d10dc0b against the full twelve-path diff.
Owned PostgreSQL cluster and its exact anonymous volume were removed after
verification; shared pre-existing image retained. Scratch remains only through
the bounded source review/local commit. No active Go process/cache remains.

Source base576dc30471f52b86b2f318c70202f66577d4277c plus eight coordinator paths.
Ordered manifestf1f35e43e204a7b588e01614d9af21ec7129bc1ae5e4377b6255988dcf2955d6.
Missing typed plan/field RED343696; focused `go test . -run '^TestCaptchaUsePlan'
-count=1` GREEN0.221s. Final complete coordinator `go test . -race -count=1`
PASS2.269s and `go vet .` exit0 (23068). Historical full-unit run2.045s precedes
final guard/mixed-role assertions and is not represented as the final result.
Disposable counterfactual removing policy-agreement and admission guards caused
actual failures: shared risk owner admitted conflicting policy versions;
structural custody enabled incomplete consuming execution. This is narrow
defect detection, not full mutation efficacy. Counterfactual tree/workfiles and
all command caches removed by trap. Exact original primitive archives in
/private/tmp/identity-captcha-plan.Rnvclu support development checks, not clean
consumer/release source pins. Original review rejected the kernel bypass below;
current nine-path final review APPROVE and source commit d10dc0b close that defect.
Owned scratch removed after commit. Native reserve/apply/finalize
and release/recovery are still required before executable admission is opened.

### Native consuming-plan review and rejection regression

**REQUEST_CHANGES**

Native reviewer: `/root/captcha_use_plan_final_576dc30`
Repository: `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`
Branch: `feature/identity-platform`
Base/HEAD: `576dc30471f52b86b2f318c70202f66577d4277c`
Independently confirmed ordered eight-path manifest: `f1f35e43e204a7b588e01614d9af21ec7129bc1ae5e4377b6255988dcf2955d6`.

#### Blocker

**Kernel preparation silently drops CAPTCHA custody instead of rejecting it.** At `postgres/coordinator.go:188`, the non-feature rejection excludes `CaptchaEvidence`. The snapshot constructed at line 224 also omits it. Consequently, adding even `&CaptchaEvidencePlan{}` to the otherwise valid kernel fixture in `postgres/coordinator_test.go:16` still allows `PrepareCommandPlan()` to succeed with ordinary executable admission. `KernelPlanDigest()` cannot catch this because it receives the already-stripped snapshot. A caller retaining the ordinary prepared digest and adding the declaration can likewise pass `ReserveCommand()`'s preparation and digest comparison.

This violates the batch's explicit fail-closed preparation/reservation contract and discards a security-sensitive declared obligation. Add `plan.CaptchaEvidence != nil` to the non-feature rejection before snapshotting, and cover both kernel preparation and reservation with otherwise-valid kernel input.

**Falsification:** Traced the valid creation fixture through every dominating guard, snapshot, kernel digest/admission, and reservation digest comparison. None examines the original CAPTCHA field; feature admission's new guard is unreachable on this path. Runtime reproduction was not executed under the read-only/no-Go assignment.

#### Remaining review results

- Complete literal sweep covered all eight files.
- Feature encoding conserves all nine custody fields; policy disagreement and duplicate risk ownership reject.
- Decoder counts are bounded; canonical re-encoding rejects duplicate/missing fields, alternate order, foreign profiles, and trailing bytes.
- Delivery/issuance wrappers retain the declaration and cannot bypass feature admission.
- Absent declarations preserve existing encoding branches and historical profiles.
- Existing regression covers feature rejection, not the kernel counterexample above.
- `git diff --check`: passed.

Parent-supplied race/vet and targeted counterfactual results are acknowledged, not independently executed. This review does not certify native consumption, real-primary behavior, clean-consumer compatibility, release gates, or unit completion.

Primary reproduced both empty and structurally valid custody through public kernel
preparation/reservation: `go test . -run
'^TestCaptchaUsePlanKernelPreparationAndReservationRejectCustody$' -count=1`
FAIL0.229s, terminalc4deca. Preparation succeeded and reservation crossed to
coordinator-configuration rejection rather than command-plan rejection.
Added the missing early guard. Current complete nine-path manifest
4aed6348d16f0052683f902e6ef79140233c9bc8fc5068461d9716599eed84eb;
all consuming-plan tests PASS0.225s. Final complete `go test . -race -count=1`
PASS1.950s and `go vet .` exit0 (terminalc183f0, session38296); disposable cache
removed by trap. New final review APPROVE below. No protected consumption
or package completion claim.

### Native complete consuming-plan re-review

**APPROVE — No findings.**

Native reviewer: `/root/captcha_use_plan_final_576dc30`
Repository/branch: `golib-go-identity-identity-platform`, `feature/identity-platform`
Base/HEAD: `576dc30471f52b86b2f318c70202f66577d4277c` plus complete nine-path dirty diff.
Independently confirmed manifest: `4aed6348d16f0052683f902e6ef79140233c9bc8fc5068461d9716599eed84eb`.

Complete-diff re-review confirms:
- Kernel preparation now rejects CAPTCHA declarations before snapshotting; preparation and reservation regressions require the specific `command_plan` rejection, not a later configuration failure.
- Feature custody remains deterministic, bounded, independently versioned, and conserved alongside phone evidence or counters under one exact risk owner.
- Historical decoder canonicality and absent-profile compatibility remain intact.
- Feature, kernel, authenticated, wrapper, and recovery admission cannot turn structural custody into executable consuming authority.

Verification: independently inspected `git diff --check` passed. Parent supplied focused CAPTCHA tests passing, complete coordinator `go test . -race -count=1` passing in 1.950s, and `go vet .` exit 0; reviewer executed no Go commands.

Approval covers this nonterminal structural development batch only. Native protected consumption, real-primary consuming/recovery behavior, clean-consumer compatibility, exact coverage/mutation and remaining release gates are not certified; unit/platform completion remains unproven.

## Native consuming storage development

### Native consuming reservation review

**Native review: REQUEST_CHANGES**

Principal: `/root/captcha_use_reservation_final_d10dc0b`
Repository: `golib-go-identity`, selected worktree, `feature/identity-platform`
Base/HEAD: `d10dc0b4a9c4a8cec3eaddfb2ef023fb65e29e54` plus all 12 supplied dirty paths. Verified ordered manifest: `972b49f0a654ff6668a5b7b3655bba46ecff3bad3a7eb9a9cc8df113cef8289c`.

#### Blocker: combined counter denial rolls back its required durable result

Anchor: `postgres/coordinator.go:356–357`.

An admitted plan can declare both `CommandCounters` and `CaptchaEvidence`, with matching policy versions and one contributor implementing both interfaces. When counters return `CommandCounterDenied`, `reserveCommandCounters()` persists the charge/audit and terminalizes the command to `aborted` (`command_counter.go:259–263`). The new unconditional `reserveCaptchaUse()` then calls `beginJournal()` (`captcha_use_reservation.go:177`), whose pending-state guard rejects this already-aborted command. `failedReservation()` consequently rolls back the charge, denied audit and aborted receipt, returning a zero result and conflict instead of the documented durable `NotCommitted`/`ForbiddenError` result. Repeated denied requests likewise lose their recorded attempt.

Falsification: reread structural validation, admission, all dominating coordinator guards and journal semantics. No guard excludes CAPTCHA plus counters; `validateCaptchaEvidencePlan()` expressly accepts matching counter policy versions. Counter denial neither exits orchestration nor changes the in-memory `journalOwned` disposition. `beginJournal()` explicitly rejects every non-pending persisted state. This is source-confirmed; reviewer executed no Go.

Preserve the existing counter-denial commit and bypass subsequent consuming reservation phases on denial. Add a real PostgreSQL combined-profile regression asserting durable charge/audit/aborted receipt, unchanged issued CAPTCHA evidence, no CAPTCHA callback, and matching replay without another charge.

#### Non-blocking: public admission documentation is stale

`postgres/plan.go`, `CommandPlan.CaptchaEvidence` documentation still says executable admission remains closed until all reservation/current-check/finalization/recovery phases exist. This batch opens typed preparation and reservation while keeping domain execution/recovery closed. Update that field's documentation to match the precise distinction already described in README.

#### Verification and claim limits

Parent-supplied current results: focused CAPTCHA tests passed; affected actual-PostgreSQL race/integration command passed in `8.394s`; `go vet -tags integration .` passed; controlled callback omission failed for expected reasons. These scenarios do not exercise combined counter denial. Reviewer's `git diff --check` passed.

Complete supplied diff and relevant preparation, authenticated entry, historical encoding/replay, transaction lifetime, sibling participant and recovery paths were inspected read-only. No additional confirmed defects found. Native adapter cryptographic authentication, domain consumption, proof-bound release and package/release completion remain explicitly unimplemented or unverified; this report does not certify them.

Primary reproduced the combined-role blocker in actual PG18.6:
`go test . -tags integration -run
'^TestPostgresIntegration$/captcha_use_reservation/combined_counter_denial'
-count=1` FAIL0.536s (d8b1dd): Conflict/zero outcome instead of durable denial.
New regression observes the protected result, denied audit, recorded charge,
aborted journal, untouched issued evidence, zero CAPTCHA callback and
non-recharging terminal replay. Coordinator now bypasses the consuming phase
after counter denial; public field documentation corrected. Current full
twelve-path manifest fdce3caecbb82c8c5ced8380e869367c55a9a3a0fac39925ef9d0ac05ff37b5a.
Final focused actual PG PASS0.495s, affected CAPTCHA/phone/replay reservation
integration race PASS12.713s and integration vet0 (8ce117), exact unchanged
manifest above. Disposable command cache removed by trap. Complete final
re-review remains pending.
Second owned loopback cluster uses volume
e9e1388d91644e1cfcd09a2736ec63bdcb44c7b11b4ed47c0462d25047dc2083.

### Native complete consuming-reservation re-review

**Native disposition: APPROVE — No findings.**

Principal: `/root/captcha_use_reservation_final_d10dc0b`
Base/HEAD: `d10dc0b4a9c4a8cec3eaddfb2ef023fb65e29e54`, `feature/identity-platform`, plus the complete supplied 12-path dirty diff. Revalidated ordered manifest: `fdce3caecbb82c8c5ced8380e869367c55a9a3a0fac39925ef9d0ac05ff37b5a`.

Complete final-diff re-review covered typed admission/validation, immutable ownership, authenticated callers, historical encoding/replay, reserve/rebind transaction boundaries, handle expiry/redaction/budgets, sibling participant interactions, and deliberately closed domain/recovery paths.

Both original findings are resolved:
- Counter denial bypasses CAPTCHA reservation and preserves its durable charge, denied audit and aborted receipt. Remaining participant calls are structurally no-ops for this admitted combination.
- Public `CommandPlan.CaptchaEvidence` documentation now distinguishes supported preparation/reservation from unsupported execution/release.

Parent-supplied final evidence: actual-PG combined denial regression passed `0.495s` after demonstrated RED; affected CAPTCHA/phone/replay race-integration command passed `12.713s`; `go vet -tags integration .` passed. Reviewer's `git diff --check` passed. Reviewer executed no Go or mutations.

Approval covers this coherent local prerequisite batch only. Native risk-adapter implementation/authentication, consuming current checks/finalization, proof-bound release, clean release-consumer composition and full package/release verification remain outside this approval and pending.

Source locally committed bc3f51e, source worktree clean. Both task-owned PG
clusters/anonymous volumes and the mutant fixture/workfiles removed; shared
image retained. Exact primitive archive scratch removed after local commits.
Runtime evidence conserved from the exact reviewed tree, not rerun for metadata.
Full counts unchanged:5/67 integrated,1/67 verified. No push/release claim.

Base62222cb, six risk/postgres paths including three new migration/test files.
Ordered manifest77d7c77e17f6ebf43c4d16a14b901111d0120dbb785ec6bf0fb9ab8881df4008.
Missing API RED3157c1 (CaptchaUseMigrationSource undefined). Temporary module
wiring setup first failed for historical unpublished migrations pseudo-version;
that is setup failure, not RED. Exact original primitive archives are supplied
by task-owned go.work, with no tracked dependency changes or checksum weakening.
Focused `go test . -tags integration -run
'^TestCaptchaUseMigrationAllocation$|^TestCaptchaMigrationSQL$' -count=1`
PASS0.362s (9559d1), real isolated PG18.6 image
sha256:5a5a84b19854a9ffaa54082c166ff4ec27473a361e496e5ea167f298f2da9722.
The test observes whole-upgrade rollback for old unknown reserved history,
exact original row conservation, partial ownership denial, existing command
and purpose FK, full-width generation, exclusive expiry and terminal tombstone
custody after envelope erasure. This is not protected consumption, transition
authorization or authoritative noncommit proof. Affected `go test . -race
-tags integration -run '^TestCaptcha|^TestIssuedEvidenceSameWorkIntegration$/native-captcha'
-count=1` PASS6.015s; `go vet -tags integration .` exit0, terminal64032.
Independent final review APPROVE below. Source locally committed as
576dc30471f52b86b2f318c70202f66577d4277c; no rerun for metadata-only history.
Fixture /private/tmp/identity-captcha-consumption.V3CxKy and its explicitly owned
loopback cluster are retained only through this bounded review/testing batch;
each command cache is removed on exit. No package promotion or release claim.

### Native consuming-schema final review

Principal: `/root/captcha_use_schema_final_62222cb` — independent read-only single-reviewer pass using `$requesting-code-review` and `testing-coverage`.

**APPROVE. No findings for this schema-prerequisite development batch.**

Independently matched branch `feature/identity-platform`, base/HEAD `62222cb3f60a38f91fc9ff8325c44c5b593ab5de`, all six dirty paths, and ordered manifest `77d7c77e17f6ebf43c4d16a14b901111d0120dbb785ec6bf0fb9ab8881df4008`.

Source-backed closure:
- Additive migration preserves original issuance columns and replay relations; existing production issuance/recovery use explicit columns.
- Composite consumer FK retains same-tenant, exact-purpose journal linkage. Explicit nonnull checks prevent partial ownership bypass through SQL NULL semantics.
- Generation supports positive integral uint64 range; reservation/finalization enforce original exclusive expiry.
- Unowned legacy consuming states fail constraint installation; populated-upgrade tests independently assert rollback and original-row conservation.
- Terminal response erasure retains consuming fields and journal references. Native CAS, cryptographic authentication and legal transitions remain explicitly outside this schema’s authority.
- Challenged partial linkage, foreign journal scope, expiry equality and erased terminal response against actual guards and test oracles.

Verification supplied by parent: missing-symbol RED; PG18.6 focused upgrade/allocation PASS; affected CAPTCHA/native issuance integration race PASS `6.015s`; integration vet exit 0. Independently ran `git diff --check`: PASS. No Go commands or service access performed by reviewer.

Approval is not native consumption, provider interoperability, clean-consumer, release-gate or unit-completion evidence.

Corrected complete four-path manifest:
`5be1bcc8281b3c7df1303389f2361fcf52daba0796854f2f1d5deb7262d9781f`.
New tests reject valid-distinct signed original, valid foreign current tier,
nonmicro/nonUTC primary time, exact expiry and zero custody through both recheck
forms, and prove retained-key addition conserves the original signed claim.
Second-clock cancellation returns context.Canceled and zero use.
Owned source-copy counterfactual bypassing original/current/primary checks and
the second cancellation guard failed both new test boundaries as expected
(29830: invalid native recheck returned claim, foreign current custody returned
claim, mid-precheck cancellation returned custody). Current full-risk
`go test . -race -count=1` PASS1.417s and `go vet .` exit0, same terminal.
Owned mutation fixture/workfile and all command caches removed. This is narrow
defect-detection proof, not exact full mutation efficacy. Final correctness
APPROVE below covers this corrected manifest. Source locally committed as
62222cb3f60a38f91fc9ff8325c44c5b593ab5de; no runtime rerun for commit metadata.

Historical first-review four-path ordered manifest (`risk/captcha_use.go`, `captcha_use_test.go`,
`README.md`, `CHANGELOG.md`):
`9c57e21c11c97894fd8e2e96fbbb4b8983fd5800b77a79fc1a09f876b3d9e1c5`.
Callback-free native custody API RED85537 same focused selector failed solely
for missing CaptchaEvidenceCustody.RecheckUse. Final whole-risk
`go test . -race -count=1` PASS1.362s; `go vet .` exit0 (35753).
All command caches removed. Independent regression/compatibility underway;
complete final-diff correctness still pending. Consuming coordinator/adapter
phases remain implementation work; no package status promotion.

Initial core API RED29560: `go test . -run
'^TestCaptchaUsePreservesOriginalAndSeparatelyBindsConsumer$' -count=1`
failed for missing CaptchaEvidenceUseRequest, issuer.PrecheckUse/RecheckUse and
EvidenceKey.RestoreCaptchaUse. GREEN25474 same selector PASS0.219s.
The test authenticates local original native custody and consumer identity only;
it does not claim journal existence or consumption. Complete risk module
`go test . -race -count=1` PASS2.019s and `go vet .` exit0 (92428), before adding
the callback-free startup-custody consumer regression. This tests six closed
flow variants, altered signed native fields, key lanes, owned alias copies,
exclusive interval, zero values, cancellation and redaction. Native PG consuming
plan/handles/migration/finalization remain unimplemented, not proven by core tests.

Temporary workspace `/private/tmp/identity-captcha-use.rk4kSs` selects current
root/risk, immutable capability2985d8ea8700bba3c5560dd34c700f0e66598d13 and
authorization60ed4b3229d83dabc40b3e81235ee7a69a95c4b1 archives, audit/clock v1.0.0
archives. No permanent replaces. Initial setup attempts failed on unpublished
root revision and existing clock checksum drift; those are not behavioral RED
and are not clean-consumer proof. Owned per-command caches removed by traps.

## Original native implemented-coverage report

**Native principal:** `/root/captcha_use_regression_d1df7de`; read-only `regression_reviewer`, `$testing-coverage`.

**Reviewed:** base/HEAD `d1df7dee0a5627a8d6bb568be165fd3fce5f65a8` plus complete four-path change; verified manifest `9c57e21c11c97894fd8e2e96fbbb4b8983fd5800b77a79fc1a09f876b3d9e1c5`.

**Disposition: REQUEST_CHANGES — regression coverage gap, not a demonstrated production defect.**

1. **P2: Native recheck rejection boundary is untested** (`risk/captcha_use_test.go:50–62`). Both issuer and callback-free custody `RecheckUse` receive only valid original custody/configuration/time. Existing lower-level restoration tests cannot detect bypassing those checks in the new bridge.

   **Concrete counterfactual:** replace issuer recheck with retained-key `RestoreCaptchaUse(p)` followed by `claim.Validate(now)`, ignoring supplied `original` and current verifier configuration. Every new test still passes, but signed custody from a foreign current configuration becomes admissible. Dropping primary UTC-microsecond validation likewise survives this suite.

   Add focused rejection cases through both public recheck forms: changed valid verifier configuration with retained evidence key; distinct valid signed original; nonprimary time; exact expiry; zero custody. Require redacted error and zero returned claim. Add retained-key rotation success so rejecting every rotated profile cannot satisfy the negatives.

2. **P2: Cancellation during local precheck is unprotected** (`risk/captcha_use_test.go:179–182`). Only already-cancelled context is tested. Removing the second `ctx.Err()` check survives all new assertions although cancellation triggered by the configured clock would return successful custody. Add a clock-triggered cancellation case asserting `context.Canceled` and zero use.

**Covered:** six closed flow variants, separate consumer actor, issuance/consumer command separation, signed-field tampering, both key lanes, owned alias projections, exclusive original expiry, selected zero inputs, JSON rejection and bearer-format redaction.

**Limits:** no Go execution, writes, services, remotes or children. Recorded parent race/vet passes are not independently executed here; exact coverage/mutation and clean-consumer evidence are not established. This review addresses local authenticated custody only—not durable issuance, reservation, finalization, protected mutation or full consuming-flow completion.

## Original native Go compatibility report

Native principal: `/root/captcha_use_compat_d1df7de` — read-only Go compatibility review.

**Disposition: APPROVE** for the bounded additive local-custody slice; no material compatibility findings.

Reviewed exact HEAD/base `d1df7dee0a5627a8d6bb568be165fd3fce5f65a8` plus all four dirty paths:
- `risk/captcha_use.go`
- `risk/captcha_use_test.go`
- `risk/README.md`
- `risk/CHANGELOG.md`

Ordered SHA-256 manifest independently matched `9c57e21c11c97894fd8e2e96fbbb4b8983fd5800b77a79fc1a09f876b3d9e1c5`; unchanged at final inspection.

Compatibility conclusions:
- New exported types and methods are additive. Existing phone-use and CAPTCHA issuance APIs, signatures, signed issuance format, and native consumers remain unchanged.
- Original flow/configuration, replay aliases, retained-key authentication and exclusive expiry are preserved. Consumer command identity remains distinct; administrator actor substitution fails closed.
- `Prepared()` and restoration own copied alias slices; caller mutation cannot alter authenticated custody.
- Native `CaptchaEvidenceCustody.RecheckUse` requires no executable issuer, dummy executor, provider, clock or store cycle. Its call path performs local authentication only.
- New production imports are exclusively standard library; no PostgreSQL dependency enters core.
- Documentation explicitly denies stand-alone protected-use authority.

**Adoption limits:** No current production consumer uses the new CAPTCHA-use bridge. Durable reservation, generation/fingerprint/target-version binding, locked primary-time apply, atomic finalization and unknown-outcome recovery remain unimplemented for protected CAPTCHA consumption. This approval does not certify those transitions, package completion, clean-consumer compatibility, source pins or release readiness.

No Go commands, service access, writes or remote operations performed. Parent-provided race/vet results were not independently executed.

## Native contract map

Read-only mapping at identity HEAD/base `d1df7dee0a5627a8d6bb568be165fd3fce5f65a8`, clean worktree; control HEAD `68c1fd1ab3e92e328993243a2a6de9d46c198483`. No writes, Go execution, services, remotes, or children.

**Missing boundary: CAPTCHA issuance exists; durable protected consumption does not.** `risk/postgres/repository.go` explicitly says CAPTCHA protected use is unsupported. `captcha_issued_evidence` permits consumption state names but lacks consuming command/fingerprint/generation custody. No native `CaptchaEvidenceUse` precheck, signed consuming projection, coordinator consuming plan, or reserve/apply/finalize/release participant exists.

### Ownership and existing seams

| Boundary | Existing owner and paths | Constraint |
|---|---|---|
| Provider verification and normalization | `risk/captcha_issuance.go`; provider-neutral/nested verifier goals | Provider facts are not authorization; no durable mutation in verifier. |
| Exact issuance intent and proof custody | `risk/captcha_replay.go`, `captcha_issuance.go`, `captcha_native.go`, `captcha_custody.go` | Core derives token aliases; signed restoration authenticates historical input, not current permission or durability. |
| Replay winner reservation | `postgres/captcha_reservation.go`, `captcha_ownership.go`; `risk/postgres/captcha_reservation.go` | **Issuance-only** `CaptchaReplayReservation`; cannot be reused as consumption reservation. Winners/aliases persist independently of domain rollback. |
| Issuance mutation/recovery | `risk/postgres/captcha_command.go`, `captcha_execution.go`, `captcha_executor.go`, `captcha_cipher.go` | Evidence insertion and safe issuance result use existing command UOW. Recovery is by original command/fingerprint, not fresh provider verification. |
| Sole transaction/journal owner | `postgres/coordinator.go`, `feature_execution.go`, `domain.go`, `recovery.go` | Reservation precedes domain Work; current checks before mutation and before terminalization; native effects, audit/outbox, result commit together. |
| Existing analogous consuming protocol | `risk/use.go`, `use_bridge.go`; `postgres/plan_risk.go`, `risk_reservation.go`, `risk_execution.go`, `risk_release.go`; `risk/postgres/use_command.go`, `reservation.go`, `release.go` | Phone RiskEvidence is distinct; preserve its exact purpose/scope. Do not reinterpret its plan as CAPTCHA or introduce another transaction owner. |

### Authoritative consuming requirements

`TRANSACTION_CONTRACT.md:420–496` requires:

- Existing durable `issued` evidence; local proof alone never authorizes mutation.
- Consuming fingerprint covers opaque reference and exact action/request/subject/version binding.
- Reservation locks the existing evidence in the **same `tx.uow.reserve` transaction** as command and all declared one-time participants.
- Apply rechecks exact ownership, primary expiry, trusted flow, action/request fingerprint, provider/site/configuration and current risk-policy version inside domain Work.
- Finalize CASes `reserved → finalized` in the protected mutation's commit with authority/session changes, audit/outbox and result.
- Same-command replay returns the recorded protected result without provider work or another consumption.
- Another command gets indistinguishable replay denial for missing, invalid, terminal or differently owned evidence.
- Unknown commit retains reservation; timeouts, lease expiry and evidence expiry prove neither noncommit nor release.
- Primary authoritative all-participant recovery alone can release to terminal `released`; retry then needs new evidence.
- Generation takeover must CAS-rebind **every** declared participant, never transfer CAPTCHA separately.
- Terminal payload erasure preserves scoped keyed-digest/state tombstones and referenced-key continuity; unresolved reservations are never time-pruned.

### Variants that must remain distinct

`FlowContextInput`/`NewFlowContext` already enforce:

- All flows: tenant, canonical operation, challenge, canonical request fingerprint, exactly one subject/anonymous flow.
- Pre-auth: pre-auth transaction, no session/admin actor; supports known subject **or** anonymous flow.
- Authenticated: subject plus exact session, no pre-auth/admin actor.
- Administrator: independently authenticated actor, optional independently verified session, no pre-auth; subject/anonymous target remains separate from actor.
- Structural flow validation is not authentication. Protected feature/native authority checks must independently resolve current tenant/user/session/admin permission.
- Core time/precheck is provisional; database time must enforce exclusive original expiry during reserve, apply and final CAS. Recovery authenticates historical custody without requiring now-valid evidence or minting freshness.

### Consumers and current executable limits

The authoritative CAPTCHA consumer inventory is `API_OPERATIONS.md:195`: password signup/signin/change/set/reset, admin password set, username signup/signin, magic-link request, OTP send/signin, phone verification/signin/password signin/reset, anonymous signin, passkey signin options/verify, social OAuth starts/token/One Tap, and SSO discover/start.

**That catalog is not proof these consumers currently execute.** In this identity worktree, phone/password/username/anonymous implementations are absent as directories. Existing native phone consumption mapping supports only `identity.phone.password-reset-request.native-v1`; completion is explicitly rejected by `postgres/plan_feature.go:41`. Production risk adapter has reserve/rebind/release but **no** `CheckRiskEvidence` or `FinalizeRiskEvidence`; `postgres/risk_execution_integration_test.go` labels its SQL owner a coordinator-atomicity fixture, not production risk policy or phone recovery.

Therefore a CAPTCHA adapter plus fixture mutation can advance and prove the consuming infrastructure contract, but cannot certify supported signup/signin/reset end-to-end workflows or full risk unit completion.

### Smallest coherent advancement

Add one typed CAPTCHA consuming boundary through the existing core/native adapter/coordinator ownership chain:

1. Core-authenticated immutable use request/precheck and signed native projection preserving original full flow/configuration/proof/expiry and separate consumer identity.
2. An admitted CAPTCHA consumption declaration in the existing command plan/fingerprint/encoding; select the exact risk store instance alongside existing participants.
3. Invocation-limited reservation/rebind, apply, finalization and authoritative recovery/release through existing Coordinator and Work phases.
4. Forward-only risk-owned migration adding exact consuming custody and legal-state consistency to existing CAPTCHA evidence storage; preserve issuance winners, aliases, configuration/key continuity and historical recovery.
5. Complete atomic consuming proof against the declared protected-domain participant, with production-consumer limitations stated accurately.

No second journal, orchestration service, retry engine or transaction owner is needed.

### Concrete contract tension / prerequisites

- `PUBLIC_CONTRACTS.json:37430–37461` names `risk.CaptchaEvidenceContributor` with `identitypostgres.Work` signatures. Literal placement in core `risk` would invert/cycle the current dependency boundary. Preserve the selected durable semantics via consumer-owned adapter interfaces; do not introduce the catalog's PostgreSQL import into core risk.
- Existing `CaptchaReplayReservation` validation forbids `RiskEvidence`, counters, capability, capability issuance and delivery (`postgres/captcha_reservation.go`). Consumption must be distinct; changing issuance into a generic consuming flag breaks its closed role.
- Existing feature plan contributor counts and owner recognition only admit extra risk participants for phone evidence or counters (`postgres/plan_feature.go`). A consuming declaration must update structural admission, registry identity, persisted plan decoding/recovery and domain terminalization together.
- Phone completion remains a separate multi-participant prerequisite requiring capability/OTP phases; CAPTCHA progress must not silently enable it.
- A real supported protected-consumer claim requires that consumer's production implementation and current-authority checks, not only a synthetic SQL fixture.
- Forward migration, original issuance replay after consumption/tombstoning, current-policy/key continuity, populated-state recovery and non-enumerating failure equivalence are part of this advancement's causal slice.

Native principal: `/root/captcha_use_map_d1df7de`, read-only change_mapper.

## Native QA plan — one bounded pass

**Role:** `qa_test_planner`, read-only; `$qa-test-planner` read completely.
**Source head:** `go-identity` `d1df7dee0a5627a8d6bb568be165fd3fce5f65a8`.
**Control head:** `go-library-tools` `68c1fd1ab3e92e328993243a2a6de9d46c198483`.
**Disposition:** no requirement contradiction or unsafe design established; implementation may proceed. No files changed, tests executed, services accessed, or children spawned.

Authoritative scope: supplied goal attachment; complete risk, risk/PostgreSQL, and CAPTCHA goals; `TRANSACTION_CONTRACT.md:420-497`; CAPTCHA reference-policy rows; public-contract fragment; acceptance catalog. This is **behavioral, security-sensitive, persistence-sensitive, concurrency-sensitive, replay-sensitive public-contract work**.

### Actual missing behavior

Current CAPTCHA implementation and tests establish **issuance**, encrypted original-response custody, durable provider-token replay winners, and issuance-owner recovery. `CaptchaReplayReservation` is explicitly restricted to `identity.risk.captcha-issue`; it does not reserve an issued artifact for a protected action.

`captcha_issued_evidence` already names consumption states but presently has no consuming-command reservation linkage. The existing `RiskEvidencePlan`/reservation/use APIs cover **phone RiskEvidence**, not CAPTCHA. Existing CAPTCHA integration success therefore cannot establish protected-use atomicity.

### Smallest sufficient behavioral matrix

| Risk / realistic defect | Proving boundary and required observable assertion |
|---|---|
| Signature/precheck accepted without durable issuance | External core API negatives plus real coordinator/PostgreSQL: fabricated or missing row, uncommitted issuance journal, zero proof, altered reference, wrong key/configuration deny; protected state, successful audit/outbox, and terminal result remain absent. No evidence row is created by reservation. |
| Flow substitution / confused deputy | Exercise pre-auth anonymous, pre-auth subject, authenticated subject/session, administrator subject, administrator anonymous, and administrator-with-session. For each legitimate variant, commit a protected transition. Change tenant, action, request fingerprint, challenge, subject/flow, pre-auth transaction, session, actor, provider/site/profile/configuration, or policy version independently; assert denial before protected mutation. Invalid tag combinations deny structurally. |
| Consumer proof does not bind target | Change consuming command ID/fingerprint and expected target/authority versions after preparation. Core validation and adapter admission must reject; no mutation/effects. Issuance ID must remain distinct from consuming command ID. |
| Two consuming commands win | Real PostgreSQL barrier-controlled race against one issued artifact: exactly one durable consuming owner and one protected transition; losing command sees the same redacted denial as missing/terminal/foreign evidence. Same-command replay returns the recorded result, without another mutation, artifact, provider call, audit, or outbox record. |
| Expiry checked before lock wait | Hold the relevant row lock until original expiry. Reserve/apply must deny after acquiring the lock using primary time. Deterministic core boundary cases: just before expiry succeeds; exact expiry and immediately after deny; copies/rechecks preserve original times and cannot renew validity. |
| Partial or stale generation takeover | Real coordinator takeover with CAPTCHA plus another declared one-use participant: exact same command/fingerprint transfers all from predecessor to successor atomically. Missing, terminal, wrong-owner/fingerprint, nonprior generation, or generation overflow rolls back the **entire** takeover. Stale handles cannot apply/finalize. |
| Precheck substitutes for current authority | Revoke/block the subject, revoke/replace the session or administrator authority, disable selected configuration, or change policy/target versions between preparation and apply. Domain-time authorization must deny without protected mutation. Assertions should check the authoritative boundary, not merely retained prepared attributes. |
| Finalization detached from domain/effects | Successful real PostgreSQL transaction: evidence finalized together with target mutation, authority/session changes applicable to the selected operation, audit/outbox, and command result. Inject domain, audit, lifecycle/outbox, finalization, cancellation, and serialization/deadlock failures: no partial protected success. A fixture must perform a real persisted domain transition; the existing reservation-only phone fixture explicitly cannot prove this. |
| Rollback or elapsed time releases unknown ownership | Proven live-owner rollback may retry only under retained exact ownership; terminal release requires authoritative noncommit proof and thereafter demands new evidence. Lost acknowledgement/disconnect returns Unknown and no new authority. Lease expiry, evidence expiry, cancellation, cleanup, and a different command cannot release/reassign an unresolved reservation. |
| Recovery issues another bearer or reexecutes | Restart recovery using retained command custody and the real primary journal. Committed returns the **original protected result**, with unchanged mutation/effect counts and no provider invocation/new CAPTCHA artifact. Proven noncommit permits the specified release/retry path; missing/corrupt/contradictory participant evidence stays Unknown. Recovery requires current disclosure authority and never treats ciphertext/signature alone as permission. |
| Forward migration or diagnostics weakens custody | Populated forward migration preserves existing issued CAPTCHA rows, original expiry/envelope, replay winners/aliases/key references. Forbidden transitions and malformed linkage fail. Formatting/serialization/errors reveal neither bearer, token, proof/signature/digest, nor private flow linkage. |

### Focused red-green entry

First regression should issue CAPTCHA through the existing real owned issuance path, then attempt a protected command whose success requires consuming that artifact. Assert the domain mutation and terminal consumption commit together, and that a second command cannot consume it.

The initial failure must be the missing protected-use contract/participation—not missing infrastructure, fixture setup, or an unrelated assertion. Add deterministic core negatives before expanding the PostgreSQL matrix.

### Verification scope and reduction

1. Core external-consumer construction/validation tests, exact time boundaries, altered proofs, redaction, and concurrent immutable-copy use.
2. Coordinator handle/plan tests: declaration completeness, exact participant identity, generation/lifetime, cancellation, and no access after callback return.
3. One owned real-PostgreSQL composition suite covering the matrix, including contention, lock-wait expiry, atomic rollback, takeover, primary recovery, and populated migration.
4. Focused race run for changed modules and the existing CAPTCHA issuance/replay suite to protect compatibility.
5. Required affected module/reverse-dependant release gates before any verified-package claim; this plan does not waive exact coverage/mutation or other applicable gates.

Provider protocols are unchanged: reuse their existing evidence rather than multiplying consumption cases across four providers. Full 25-operation/four-provider reference acceptance remains downstream; a real adapter/coordinator fixture proves this bounded slice, **not complete consumer journeys**. All Go verification must use task-owned disposable `GOCACHE` with cleanup.

Native principal: `/root/captcha_use_qa_d1df7de`.

## Native complete final-diff correctness report

## Native correctness review

**Principal:** `/root/captcha_use_correctness_d1df7de`
**Role:** independent read-only correctness reviewer
**Disposition:** **APPROVE — bounded local CAPTCHA consuming-custody development batch only.**
**Findings:** no unresolved defects in the reviewed contract. No GitHub review submitted.

### Exact reviewed boundary

- Repository: `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`
- Branch: `feature/identity-platform`
- Base/HEAD: `d1df7dee0a5627a8d6bb568be165fd3fce5f65a8`
- Complete dirty scope: `risk/captcha_use.go`, `risk/captcha_use_test.go`, `risk/README.md`, `risk/CHANGELOG.md`
- Independently recomputed ordered-path manifest: `5be1bcc8281b3c7df1303389f2361fcf52daba0796854f2f1d5deb7262d9781f`
- Final identity check preserved the same four-path scope and digest.

Read `$requesting-code-review`, its independent-review, bounded-closure and advanced-boundary references, and `$testing-coverage`. Completed the blind literal sweep and source-contract analysis before reading native mapper/QA sections. No other reviewers’ conclusions used. No writes, Go execution, services, remotes or children.

### Source-backed closure

- **Original integrity and presented-reference linkage:** `captcha_use.go:63–75` restores original issuance under the selected current configuration, requires value-equal trusted flow, authenticates the presented reference against the supplied record, and routes record validation through `validCaptchaRecord()`. Its native branch rematerializes the exact original candidate; record field substitution cannot survive that equality.
- **Separate consuming identity:** `captcha_use.go:79–92` rejects the issuance command as consumer, validates bounded canonical consumer identity, preserves every flow dimension and requires administrator consumer actor equality. This remains structural custody, not independently verified authentication or permission.
- **Signed custody:** `captcha_use.go:112–124,172–175` authenticates both original and consuming signatures, binds independently configured integrity/linkage key lanes, and rejects malformed intervals and consumer identity. The original signature transitively covers issuance flow/configuration/aliases. Key-level restoration intentionally admits authentic expired history; `Validate()` cannot grant current validity at expiry.
- **Current configuration and primary time:** `captcha_use.go:132–152` rejects non-UTC/nonmicrosecond primary timestamps, requires the supplied original to authenticate under current immutable custody and match the consuming projection, then authenticates the consuming claim and unchanged interval. The callback-free form delegates exactly to that path without clock/provider/executor invocation.
- **Temporal behavior:** `captcha_use.go:58–74,165–169` checks initial and post-clock cancellation, denies backward/over-budget sampling and pre-issuance sampling, and preserves exclusive original expiry. Copies, restoration and rechecks cannot refresh either check time or expiry.
- **Ownership/concurrency:** `Prepared()` and restoration clone the only slice-bearing projection through `clonePreparedCaptcha()`. Private retained flow/use state is immutable; no goroutine, lock, network, SQL or shared mutation is introduced.
- **Compatibility:** new public declarations/methods are additive. Existing issuer restoration, native issuance and phone-evidence use remain unchanged. Exact-symbol searches found no production CAPTCHA-use consumer yet; the phone sibling intentionally owns different scope and signatures.
- **Redaction/adoption:** formatting is constant-redacted and JSON/text serialization fails closed. README and changelog explicitly prohibit treating this bridge as protected-use authority.

**Bounded adversarial challenges:** mixed original/consuming custody, modified consumer/key/digest/times/aliases, foreign current configuration and nonprimary timestamps were traced through their dominating guards. None bypasses authentication without changing signed bytes or failing validation. Restored expired custody remains historical input, not a renewed permission.

### Test-oracle assessment

The new assertions independently require preserved issuance/consumer identities and times, rejection of tampering and foreign custody, owned alias copies, exclusive expiry, zero-state denial and cancellation returning zero custody. Six valid flow variants use fixed trusted metadata/time; fixture randomness does not control expected admission partitions.

The corrected native-recheck test specifically observes zero claim plus `ErrInvalidEvidenceExecution` for a valid-distinct signed original, valid foreign tier, nonprimary time and exact expiry, through both issuer and callback-free custody forms. Thus it rejects merely calling retained-key restoration/interval validation while ignoring current configuration or supplied original.

**Supplied execution evidence, not rerun by this reviewer:** missing-API RED29560/RED85537; focused GREEN0.219s; current complete-risk `go test . -race -count=1` PASS1.417s and `go vet .` exit0 at terminal29830. Supplied owned counterfactual evidence reports expected original/configuration/primary-time and second-cancellation failures. This supports narrow defect detection, not full mutation efficacy.

### Claim limits

Approval permits this coherent **nonterminal local prerequisite batch**, consistent with the operative goal’s development-boundary rules. It does **not** certify durable CAPTCHA protected use, release readiness, verified package status or completed production journeys.

Coordinator consuming plans/handles, forward migration, authoritative reservation/apply/finalization/recovery, and actual protected consumers remain unimplemented. Exact coverage/mutation, release gates, dependency pins, clean-consumer and provider/reference acceptance remain pending; they must be satisfied at their applicable completion boundary.

## Native consuming-command custody, base bc3f51e

Bounded plan: retain all six trusted flow variants, exact signed issuance/use
custody and complete protected target/version fields; reject consumer/actor,
tenant/subject/action, reserved-field, signature and scalar-plan substitution.
Current selected verifier configuration must authenticate without callbacks;
risk policy and authorization policy are independently versioned. No durable
reservation or protected mutation is claimed by this batch.

Native results on the dirty four-path snapshot:
- Expected missing API RED c4f172 (`MapCaptchaUse`, restore helper and plan helper).
  Setup-only unpublished dependency errors were resolved with exact task-owned
  archives. The new fixture's replay keys were corrected to newest-first.
- `go test . -run '^TestCaptchaUseNative|^TestCaptchaNative|^TestCaptchaRepository'
  -count=1` PASS0.284s, dd0ebb.
- `go test . -race -count=1` PASS1.293s, b12dd0.
- `go vet .` exit0, fc777d.
- All commands ran in risk/postgres with owned disposable GOCACHE and exact
  temporary go.work/primitive archives; command caches removed on completion.
  No tracked dependency changes, private transaction or provider calls.

Four-path reviewed input identity (SHA-256, risk/postgres relative):
- captcha_use_command.go: b7c57d01401c843d6ec15e2c445d894b18b55f43ce35415745491bd2cbcabd39
- captcha_use_command_test.go: c3296598e10f3057287badd2240c3a4f5a77051f04629f810f787e972b28040f
- README.md: 69bdc55dcd8b62894f176b23d255ff9b1e7207af03fda0787c761258d0bd605e
- CHANGELOG.md: f4824f1b26798c0639173bc049393c3a62e7d5f591a6ab28627e649836ffd94d

Independent full-final-diff review pending from
`/root/captcha_use_native_review_bc3`, gpt-6.1-sol medium, isolated context.
Full package coverage/mutation, release/composition/provider checks remain
pending. Counts remain 5/67 integrated and 1/67 verified.

### Native independent first report

## REQUEST_CHANGES

Reviewed `feature/identity-platform` at `bc3f51eb4f13c06f597e894deda79c2f022ac734` plus all four dirty paths; their SHA-256 hashes match the supplied snapshot. Single-reviewer read-only pass using `$requesting-code-review` and applicable boundary/test guidance.

### Blocker — valid nanosecond prechecks cannot map

**Anchor:** `risk/postgres/captcha_use_command.go:37`

`PrecheckUse` signs `CheckedAt: now.UTC()` without microsecond truncation (`risk/captcha_use.go:74`), and core restoration explicitly accepts nanosecond precision. Mapping that instant as `FeatureTime` causes `NewFeatureCommand` to reject it whenever `Nanosecond()%1000 != 0` (`feature_command.go:167`). Consequently, ordinary clock samples can produce valid signed custody that this new public mapper cannot consume.

**Concrete witness:** otherwise valid custody checked at `record.IssuedAt + 1µs + 1ns`; precheck succeeds, but mapping returns zero outputs and `ErrInvalidEvidenceExecution`.

**Remediation:** preserve exact signed nanosecond precision in the closed native codec, such as canonical bounded RFC3339Nano text. Do not truncate or re-sign the instant. Add a public producer-to-mapper-to-restorer regression with a non-microsecond clock sample.

**Falsification:** traced all producer, mapper and constructor guards; none requires a microsecond clock. Constructor rejection occurs before its UTC cloning. Truncation cannot rescue restoration because the use signature authenticates the original RFC3339Nano instant.

### Remaining review conclusions

- Six flow variants, binding conservation, protected fields, namespace closure, exact scalar-plan comparison, dual signatures and selected-profile authentication otherwise trace consistently.
- Immutable ownership follows core projection copies and `NewFeatureCommand` deep copies. Restoration invokes no clocks, callbacks or SQL.
- Tests independently exercise custody equality, field preservation, substitution denial, selected-profile rejection and exclusive expiry. Their uniformly microsecond clocks conceal the precision defect.
- Coordinator reservation interface remains unsatisfied by this repository; domain execution and unresolved consuming recovery remain explicitly closed. This patch does not grant durable consumption authority.

**Supplied verification:** focused tests PASS; `go test . -race -count=1` PASS (1.293s); `go vet .` PASS. No reviewer execution or writes.

**Unverified/pending:** nanosecond regression, durable consuming reservation/finalization/recovery, real consuming PostgreSQL integration, exact release coverage/mutation, provider interoperability and full module/composition/release completion.

### Reconciliation

Confirmed core-clock to FeatureTime precision divergence; public regression
`TestCaptchaUseNativePreservesSignedNanoseconds` failed exactly at mapping
(d9a00f). Preserve the authenticated instant as canonical bounded RFC3339Nano
text; database-issued/expiry timestamps retain microsecond semantics. No
resigning or timestamp normalization. Final parser regressions exercise wrong
kind, empty/overlong/noncanonical and changed signed instants. Full final
review requested again from the same native principal.

Final revised Go paths:
- captcha_use_command.go: e96e51573f7d546f675dce5af7381139eb519cd0a716c03a0911e3b03f1faac3
- captcha_use_command_test.go: 184f1b322b566e311008f5cf8545673000f8114f4a2ba6c02ce69449332a3d87
Documentation paths unchanged.

Final revised snapshot `go test . -race -count=1` PASS1.296s exit0 (e68198);
`go vet .` exit0 (c8c0cb). Original failure is included in the passing suite.
Disposable command caches removed after both commands.

### Native independent final report

## APPROVE — No findings.

**Boundary:** complete four-path dirty diff against `bc3f51eb4f13c06f597e894deda79c2f022ac734` on `feature/identity-platform`. All supplied final hashes verified. Read-only single-reviewer pass; no writes or additional agents.

**Prior blocker resolved:** `captcha_use_command.go:40` preserves signed `CheckedAt` as RFC3339Nano text; lines 90–97 require bounded canonical UTC encoding before authentication. No truncation, re-signing or validity renewal occurs.

**Oracle/falsification:** the new public producer→mapper→restorer test preserves `2001ns` exactly and distinguishes the preceding nanosecond from the signed instant. Wrong-kind, empty, overlong, offset-equivalent and substituted timestamps reach the closed codec and deny. Supplied RED `d9a00f` establishes detection of the original defect.

Complete causal review also closes protected binding/field conservation, six flow variants, namespace/schema closure, immutable ownership, dual signatures, exact scalar-plan equality and selected immutable profile checks. Restoration performs no callbacks, clock reads or SQL. Reservation/domain/recovery authority remains fail-closed.

**Final supplied verification:** `go test . -race -count=1` PASS (1.296s, `e68198`); `go vet .` PASS (`c8c0cb`); reviewer `git diff --check` PASS.

Approval covers this bounded local construction/authentication batch only. Durable consuming reservation/finalization/recovery, real consuming PostgreSQL integration, release coverage/mutation, provider interoperability and complete module/platform verification remain pending.

Source committed as fc4b5c6; reviewed four-path bytes unchanged, source worktree
clean. Local development only, no push or full-unit completion claim.

## Production consuming reservation, base fc4b5c6

Bounded plan: use the existing sealed coordinator participant/transaction, lock
the indexed keyed issued row, require committed original issuance and compare
authenticated native facts. Bind exactly one consumer/fingerprint/generation;
rebind only the same predecessor without changing expiry/reservation time.
Deny tampered/terminal ownership and roll back command/claim prefixes. Test real
contention, matching replay, exact takeover, divergent linkage and expiry after
an observed primary row-lock wait. No domain execution or recovery release is
enabled by this batch.

Results so far:
- Missing native methods/conformance RED36992b.
- Real PostgreSQL 18.6 production issuance -> consuming reservation, two-command
  contention, matching replay, takeover and divergent-generation rollback
  PASS0.545s (12e52e).
- Primary lock-wait expiry plus native ownership suite PASS3.583s (b38852).
  Short-lived verifier fixture initially violated its own declared lifetime;
  corrected normalized observation expiry, not production acceptance.
- Disposable cluster `identity-captcha-reserve-hefxex`, image
  sha256:5a5a84b19854a9ffaa54082c166ff4ec27473a361e496e5ea167f298f2da9722.
  Command caches removed after each bounded invocation.
- Initial final race failed before behavioral execution: the disposable cluster
  became unavailable (37ae29, exit1, connection timeouts). Docker subsequently
  returned; restarting only the exact owned container restored PostgreSQL18.6
  on loopback50467. No shared daemon/service was restarted by this task.
- Unchanged eight-path final inputs: affected `go test . -race -tags integration
  -run '^TestCaptcha|^TestIssuedEvidenceSameWorkIntegration$/native-captcha'
  -count=1` PASS6.455s (9bb790, exit0). This includes tampered linkage and terminal
  ownership rejection plus observed lock-wait expiry. `go vet -tags integration
  .` exit0 (0b7acb). No aggregate repository/release claim.

### Native independent complete-diff report, initial disposition

**REQUEST_CHANGES — verification gap only.** No concrete code defects found.

**Reviewed boundary:** `fc4b5c6b602f904c5f0193305e2dc97d2e24475e` plus all eight supplied dirty paths; independently verified every supplied SHA-256. Single-reviewer read-only pass using `$requesting-code-review` and `$testing-coverage`; no writes, execution, delegation, or external mutations.

**Source-backed conclusions**
- `captcha_use_reservation.go:45–68`: sealed exact-instance ownership, full immutable command/plan validation, positive generation, and authenticated signed original/consumer linkage.
- `:70–100`: indexed tenant/key/digest lookup requires committed original issuance; locks the evidence row and independently compares original identity, issuance interval, signed intent/proof/hostname, and reconstructed digest.
- `:101–131`: authoritative time sampled after exclusion; current selected configuration and exclusive expiry rechecked, including expiry in the update predicate.
- `:117–136`: takeover requires exact reserved predecessor, command, fingerprint, binding, signature, and nonterminal state; updates generation only. First reservation cannot adopt foreign or terminal custody.
- Coordinator reservation owns the sole transaction and rolls back claim/journal/participant prefixes on denial; rows close before callback return.
- `postgres/domain.go:50` and `postgres/recovery.go:174` keep domain execution and recovery closed. Reservation cannot grant protected-use permission or release unresolved evidence.

**Oracle/falsification conclusions:** Real coordinator/SQL assertions independently detect altered original facts, competing consuming commands, divergent fingerprints/bindings/signatures/generations, terminal resurrection, and command-prefix advancement on failure. The lock-wait test observes actual exclusion before crossing primary expiry and checks both rollback and unchanged evidence state. Attempted counterexamples—provisional signed issuance without committed original authority, mixed consumer linkage, stale predecessor, and expiry after waiting—are rejected by dominating guards/CAS.

**Remaining readiness gap:** Affected CAPTCHA integration `-race` has not passed: latest supplied execution failed before behavioral tests because the disposable cluster timed out. Supplied non-race PostgreSQL 18.6 suite passed in `3.583s`; `go vet -tags integration .` passed. Rerun only the affected race gate after restoring the cluster; unchanged reviewed inputs permit disposition update without another full source review.

Approval scope would remain this local consuming-reservation checkpoint only. Current checks, finalization, release/recovery, protected consumers, exact coverage/mutation and remaining release/composition gates are unfinished—not module or platform completion.

### Native independent final disposition

Principal `/root/captcha_native_reserve_review_fc4`, default gpt-6.1-sol medium,
isolated context, read-only; complete eight-path diff against fc4b5c6.

**APPROVE. No findings.**

The previously reported race readiness gap is closed: supplied `go test . -race -tags integration -run '^TestCaptcha|^TestIssuedEvidenceSameWorkIntegration$/native-captcha' -count=1` passed, exit `0`, package time `6.455s` (`9bb790`). Integration vet remains passed (`0b7acb`).

Independently rechecked HEAD `fc4b5c6b602f904c5f0193305e2dc97d2e24475e` and all eight reviewed file hashes; unchanged. Prior complete-diff source, oracle, and falsification conclusions remain valid.

Approval covers only this local native consuming-reservation/takeover checkpoint. Domain current checks, finalization, release/recovery, protected consumers, and remaining module/platform release gates remain unfinished.

Reviewed eight-path identity (also revalidated unchanged after race):

```text
5d4d2e299192038806be5f9ed4c814a8296f556cc332046828451df2063d3c08 captcha_use_reservation.go
85970c373f13f3b6a7501da8cc83d24d179e2a7433c2d5bbaf599a4b2733b435 captcha_use_reservation_test.go
be6d8e89ba716d4a42cb63294a52ecffbb862c0cce2ccc5fa9e6dc63a3a44ff1 captcha_use_reservation_integration_test.go
f23e33df5681a5c530c930b795580ef58bbaa93093d8b4807a83e2cb8f0e6281 captcha_execution_integration_test.go
eda262a8b4bb190efc2acf6fe573d4d84ba6a25c76c07e55f0a27c0e14fcb835 issuance_integration_test.go
459c4feb5fdf2bc9f7187891802aedf318ab862cde559f3ecc11aca62ba16f98 repository.go
20996970b72db4c65089ff96aa05e2d0d5e325ec66dfd9164709c88c7e72a41d README.md
3bd73b2597eefba6bbdbe41334398ae8bf928b7b66c2e56af798fd63356f49ff CHANGELOG.md
```

Source committed9749928; reviewed bytes unchanged, clean worktree. Local
development checkpoint only; no unit promotion or push.
Exact owned container6272b68c and volumec8d53e5e removed after evidence capture;
shared pre-existing image retained. Per-command caches removed by check.rb.

## Domain consuming protocol, base9749928

Bounded plan: extend existing sealed Work phases for current read checks and
terminal finalization; invoke before protected mutation and again after effects.
Admit Begin only with the complete current/finalize/proof-release native phase
set. Add rollback/recovery release under the existing coordinator's primary
proof and exclusions, never expiry-based release. Real PostgreSQL test-owner
oracles must prove mutation/evidence/audit/outbox/receipt commit or rollback,
unknown retention, exact recovery and invocation invalidation. Production native
store integration remains a separate obligation until its phase set is complete.

- Missing typed handles RED1bc092; focused zero/serialization checks PASS0.208s
  (db4a82). New complete-owner public Begin REDaefae4; after wiring, independent
  primary effect/ownership suite PASS0.637s (2a9418). Initial recovery fixture
  incorrectly expected nil on the existing dependency-failure terminal receipt;
  traced successful noncommit proof and corrected the oracle to typed
  Unavailable + NotCommitted, not production behavior.
- Expanded17-case domain suite PASS1.252s (db0242): expiry after provisional
  mutation, denied initial/final checks, panic, open rows, ignored SQL failure,
  finalization failures, failed release retaining reserved/pending, unanimous
  recovery and divergent/unknown evidence denial. Final affected CAPTCHA unit
  and real PostgreSQL integration race PASS2.490s (2af20b), integration vet exit0
  (af1a5e). No native store protected consumption or release-gate claim.
- Reviewer /root/captcha_domain_review_974 running read-only complete11-path
  diff against source97499282a509997560a207ea7ec1354b8891f903; no disposition yet.
- Final unchanged coordinator whole-unit `go test . -race -count=1`
  PASS2.644s exit0 (faaae5). Affected interaction `go test . -race -tags
  integration -run '^TestPostgresIntegration$/(captcha_use_reservation|
  risk_reservation|capability_reservation|feature_execution)$' -count=1`
  (regex without prose newline) PASS11.571s exit0 (91e0f6). Same owned PG18.6,
  primitive archives and disposable per-command caches; no full repository gate.

### Native independent domain final report

## Native independent review

**APPROVE. No findings.**

Principal: `/root/captcha_domain_review_974`; isolated, read-only, single-reviewer pass using `$requesting-code-review` and `$testing-coverage`. No writes, project execution, delegation, or external mutations.

**Boundary:** `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`, branch `feature/identity-platform`; base/HEAD `97499282a509997560a207ea7ec1354b8891f903` plus all eleven dirty paths below. Hashes independently captured and rechecked unchanged.

### Source and falsification conclusions

- Complete phase ownership is required before `Begin` or recovery; the production reservation-only risk repository remains rejected.
- Apply/current-check/finalize handles enforce exact registered owner, frozen plan/binding, current generation, invocation lifetime, redaction, and shared Work poisoning.
- Current checks precede protected mutation and repeat after provisional domain/audit/outbox effects; finalization and command result share the sole domain commit.
- Abort releases only after successful rollback and reacquired current command/tenant exclusion. Missing audit evidence, rollback failure, lease loss, release failure, or ambiguous closure retains unresolved custody.
- Recovery reads the frozen plan under primary command/tenant exclusion and requires **every** participant’s noncommit proof before predecessor-bound terminal release. Unknown or divergent evidence rolls back takeover and retains reserved/pending state.
- Adversarial source traces—foreign owner, returned alias, stale generation, final-check expiry, successful finalization followed by failure, and recovery with divergent custody—do not bypass the dominating lifetime, transaction, proof, and generation guards.

### Oracle and execution conclusions

The SQL fixture observes independent persisted evidence, command state, domain mutation, audit, and outbox outcomes. Its assertions reject premature consumption, leaked provisional effects, release without unanimous proof, and revival of returned handles. It does not establish provider truth or production-store consuming behavior.

Parent-supplied final execution evidence:
- `go test . -race -count=1`: passed `2.644s`, exit `0` (`faaae5`).
- `go test . -race -tags integration -run '^TestPostgresIntegration$/(captcha_use_reservation|risk_reservation|capability_reservation|feature_execution)$' -count=1`: passed `11.571s`, exit `0` (`91e0f6`).
- `go test . -race -tags integration -run '^TestCaptcha|^TestPostgresIntegration$/captcha_use_reservation' -count=1`: passed `2.490s`, exit `0` (`2af20b`).
- `go vet -tags integration .`: exit `0` (`af1a5e`).

Approval covers this **local coordinator development checkpoint**, not production native-store consumption, protected-workflow/provider composition, module release verification, or platform completion. Exact coverage/mutation and remaining release gates are not claimed.

### Reviewed SHA-256 snapshot

```text
dd6e077bdf112a9f8fd1a158fd0638cf5b98cedc2f62df99f9f740f233afd891 postgres/captcha_use_execution.go
172ba465c49f69fca91c244acf30c3e8f93042380e79e257fc13bc1c7907b346 postgres/captcha_use_execution_test.go
1eb7771287c591a99bd081a6e66c750e5a7530ca06e861fcbaf4ad257a7502cd postgres/captcha_use_execution_integration_test.go
452569855191d23cb15832fa09984b56ed82cc57464343608665745cdc6951a3 postgres/captcha_use_reservation_integration_test.go
90b4edaf82d6f1b7ce3aa8de810eadd0c0a504c0ffcfe25390425b00189d5567 postgres/domain.go
37650bd79266aa0fbc4af903b46aef3835fc6953767a2c51f19ff356672e1831 postgres/feature_execution.go
2eee4573e9f3944cffaf5d221cd193a835271d14014ca27e43cb60e69feb47cf postgres/recovery.go
ae3f57364bec3f36c2a474544873781ea2b892dd2d479b56eda0758797137d28 postgres/recovery_participant.go
72bcee799a9b93d71d4bcdcb34036aec2328601c71a47d46ff016729565d5d94 postgres/work.go
921a7c901a26b72b404d295bba51b87864718906c11496be7f4fc94af6dd9b8e postgres/README.md
178828f5b17dc1eea21525dd1bf5fc949a4d1377f60b3417e0628633eddfb537 postgres/CHANGELOG.md
```

Source committed54f956f; reviewed11-path bytes unchanged, clean worktree. Owned
container961a16d5 and volume25c4142a removed after final evidence capture; shared
pre-existing image retained, command caches removed. Local checkpoint only.

## Native consuming execution, base54f956f

Bounded plan: production Check/Finalize authenticate complete immutable signed
custody and selected current profiles, lock exact committed-original/reserved
row, compare original facts, sample primary time after exclusion and CAS exact
current owner to finalized with protected effects. Recovery/release instead
use frozen validated ownership plus primary noncommit proof, preserve original
interval, remain possible after expiry/configuration removal, and never renew
or make evidence reusable. Shared-owner recovery must consider CAPTCHA alongside
every selected counter/phone role, not return early on only one role. Real native
producer/store/coordinator tests observe protected fixture effects, audit/outbox,
terminal rows and journal generations; fixtures are not real password workflows.

Missing native methods reproduced REDa7b4d6 (compile contract); first complete
native phase integration PASS8c41d7,6.691s: success, protected failure rollback,
expiry after provisional mutation, original-hostname tamper, unanimous recovery,
expired recovery, unknown retention and divergent predecessor retention. Real
PG18.6 owned loopback57898; SQL feature/audit/outbox seams, not actual password
workflow or production effect-adapter interoperability. Subsequent checks must
use the formatted final bytes; no release/unit promotion is claimed.

Final-tree native unit/race PASS2884be1.314s; root regression/race PASSc73f58
1.490s. Expanded integration first exposed invalid test data (zero signature
violated the existing nonzero custody CHECK), not a production failure; use a
different valid nonzero signature for the corrupt-owner recovery case.

Final current affected native PG18.6 integration/race PASS46aa96,21.418s:
`go test . -tags integration -run '^TestIssuedEvidenceSameWorkIntegration$'
-race -count=1`, via disposable-cache wrapper and owned DSN. Includes existing
phone/counter/lockout/CAPTCHA issuance and new consuming success, rollback,
expiry/tamper, exact recovery, expiry/config-removal recovery, unknown/divergent/
corrupt/terminal retention. Counts remain unchanged; full release gates pending.

Final integration vet PASSb8c7fb exit0. Source final seven-path snapshot is in
9c178a; isolated read-only gpt-6.1-sol medium complete-diff review now pending.

### Native independent review, verbatim

## APPROVE — local development checkpoint

Principal: `/root/captcha_native_consumption_review_54f`, isolated single-reviewer pass using `$requesting-code-review` and coverage guidance. No findings.

Reviewed base/HEAD: `54f956fe3b60613c795c57de76a4926c5d82a67a`, plus all seven dirty paths, including both untracked files. Snapshot remained unchanged during review.

### Source-backed conclusions

- Current checks authenticate complete admitted signed custody and selected configuration, lock exact reserved ownership, compare original facts, then sample primary time. Finalization repeats validation and independently checks exclusive expiry in its terminal CAS: `captcha_use_execution.go:87–149`, `:31–44`.
- Coordinator checks precede protected mutation and repeat before finalization; evidence, protected effects, audit/outbox and result remain in one transaction: `postgres/feature_execution.go:230`, `postgres/domain.go:310`.
- Release requires sealed coordinator rollback/noncommit authority and exact predecessor ownership; expired or removed current profiles do not become renewed use authority: `captcha_use_execution.go:53–75`, `postgres/captcha_use_execution.go:43–80`, `postgres/recovery.go:233–279`.
- Shared-owner recovery separately proves selected CAPTCHA, counter and phone-risk roles: `repository.go:297–313`.
- Original payload alteration blocks current use. Terminal cleanup concerns exact consumer ownership, not renewed original validity or issuer-response recovery. Under the supplied clarification, omitted payload validation during cleanup is intentional, not a defect.

### Verification and remaining boundaries

`git diff --check`: passed. Independently inspected tests and causal consumers; did not execute Go or provider operations. Parent-supplied race/integration results are not independently rerun evidence.

Narrow untested risk: the new native fixture does not select CAPTCHA plus counter/phone-risk roles together and corrupt one sibling role; the shared-owner guard is source-proven but lacks a combined regression here. Native current-check lock-wait expiry is likewise not directly exercised by the new fixture.

Coverage, mutation, provider interoperability and complete release gates remain pending; approval does not certify unit completion or release readiness.

### Dirty SHA-256 snapshot

```text
captcha_use_execution.go
a1645381138490b0e9c19c71308f525ac293852e37acd869738dc326c72b0db9
captcha_use_execution_test.go
ae16144ee523a2c14b2063be17dd5a38adfe21685685deed278b26fbf85a9773
captcha_use_reservation_integration_test.go
eef30af210162a2e9ff28a761dac5975f25ce1093949e35d2f523424371e8d30
captcha_execution_integration_test.go
a03cfd0982fa18de6b0654f4689c47f3ebc24fb192027d5e9271798ea6be1032
repository.go
445178615bfda9428fef8f50a432ff57a154e9d4de068b0487da9dde22a2a479
README.md
9f1d067e6d641954037bf2b674e2d726d29c9ad087edc6ac881e480ad4ece935
CHANGELOG.md
cd5e2d7b0981eb40b46c0cfd35933e085c26f2dd7c016c0ecc001b3fb8e9324f
```

All paths above are under `risk/postgres/`. No files changed or disposable resources created.

Source committeda04ba44; reviewed seven-path bytes unchanged, source clean.
Owned cluster2c0632ed/volumef7a4b9ab removed after captured checks; shared image
retained. Task caches are removed by each wrapper invocation. Local development
checkpoint, not release readiness; actual protected feature/effect composition
and the review's two narrow runtime coverage gaps remain subsequent work.

## Public CAPTCHA production-effect composition, basea04ba44

Bounded plan: public issuer/executor through real owned audit/outbox migrations
and contributors, all six subject/flow variants, authorized original recovery,
lost-commit acknowledgement and write-then-fail effect rollback. Assert durable
record/binding cardinality and token-free projection, not seam call order.
Provider verification remains a bounded synthetic external seam; this proves
owned adapter composition, not remote provider interoperability or login.

Initial public production-effect integration PASS5c1c44,0.588s. All six flow
variants, lost acknowledgement, audit and outbox write-then-fail cases pass.
Existing implementation was already compatible; this is new integration proof,
not retroactive TDD. Early failures were fixture signature/type/query corrections
(concrete contributor return type and outbox binding uses effect_id), not product
defects. A missing-effect counterfactual is required before accepting the oracle.

Oracle RED3d8bf5: task-owned archived real audit contributor falsely returned
success without persisting, and production cardinality assertion failed
"issuance effects were not atomic". Mutation removed immediately; no source
repository was mutated. This proves the composed test rejects missing native
audit persistence rather than merely counting seam callbacks.

Restored audit fixture equals upstream21ad3d11d3b2:postgres/identity.go exactly
(d0fa696e7235c20c1f7d517eda4398dafdd38c054feb8ed9b81648e3d5ce6de1).
Final public production-effect race PASSb99709,1.904s, including payload bearer/
token redaction and exact immutable binding cardinality. Full affected existing
composition regression and final independent review remain pending.

Final whole affected real PG18.6 integration/race PASSab31ee,24.387s:
`go test . -tags integration -run '^TestIssuedEvidenceSameWorkIntegration$'
-race -count=1`. Covers original phone/counter/lockout/CAPTCHA fixture paths and
real public production effects together, with exactly owned role/database cleanup.
No new runtime algorithms were needed; unit/runtime inputs outside selected
integration-tag test files are unchanged and prior unit results remain reusable.

### First independent production-effect review, verbatim

## Native independent review

**Principal:** `/root/captcha_production_effect_review_a04`  
**Disposition:** **REQUEST_CHANGES**

**Boundary:** `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`, branch `feature/identity-platform`, base/HEAD `a04ba44f48e7856da264d37699c944cd06725f3b`, plus all five dirty files, including the untracked file. Single-reviewer, read-only pass; no writes, execution, external mutations, or delegation.

### Finding

**Blocker — ineffective payload-redaction oracle:** [captcha_executor_integration_test.go:256](/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform/risk/postgres/captcha_executor_integration_test.go:256)

`row_to_json(r)` and `row_to_json(m)` encode `audit.records.canonical_record` and `outbox_messages.payload` as bytea text, normally hexadecimal. Searching those row representations for the plaintext provider token or opaque bearer therefore misses a secret embedded in the actual serialized payload. Such a regression can pass the newly claimed redaction proof.

Both production migrations declare these payload columns `bytea`; the existing sibling production-effects test already decodes them with `convert_from(...,'UTF8')`. Decode both payloads before checking, while retaining inspection of other persisted fields. Verify that deliberate token/bearer inclusion fails the corrected oracle.

**Falsification:** traced the complete query, scanning, and comparison path; no decoding occurs between bytea serialization and `strings.Contains()`. The current typed production producers exclude these secrets, so this finding concerns the acceptance oracle—not an observed production disclosure.

### Remaining review conclusions

No additional findings across the complete diff and affected chains:

- All six trusted-flow variants exercise the real public issuer/executor and actual production contributors.
- Coordinator ordering keeps evidence, success audit, lifecycle message/binding, and receipt inside one transaction; failure wrappers return errors after real participant writes.
- Lost-ack injection consumes the actual successful PostgreSQL commit acknowledgement; recovery and restart traverse persisted custody and current authorization without provider reevaluation.
- Shared migration installation retains transactional role acquisition and enclosing owned-database/role cleanup.
- Runtime APIs, schemas, original SQL-seam tests, and existing composition contracts remain unchanged.

**Verification:** `git diff --check` passed. Supplied focused/full integration race and vet results were considered reported evidence, not independently rerun. Live-provider interoperability, complete protected-login journeys, and unit-release gates remain outside this local checkpoint.

### Reviewed SHA-256s

```text
risk/postgres/CHANGELOG.md
b09ff1e2cb5cff50dd13a9432a87671c5e2dcbcc35057792ac81983e63df3f12

risk/postgres/README.md
9de25dcab7677f43c750af9978b0b338ade1c0f7b6356deeec7c0d8be3ccdf81

risk/postgres/captcha_executor_integration_test.go
4b101435470b1558695935998fb50bbbf378cae988d5b662e63952be881283d2

risk/postgres/issuance_integration_test.go
314ff8bd3948975eb9462504e11f6ddcb52e32e08e5a953d28e0607fc4e06654

risk/postgres/production_effects_integration_test.go
d25be681cf476b9d7c6ff690446464420236a08566af258e4404788cf8e56d29
```

### Redaction oracle repair

Reproduced review finding with task-owned real outbox producer emitting a valid
JSON payload containing the exact synthetic provider input. Old oracle passed
85fc97,0.481s despite that leak. Both audited payload columns now explicitly
decode UTF8 in addition to complete row metadata; same controlled leak must fail.
This was a test defect, not observed production disclosure. No new QA-plan chain.

Corrected oracle RED63063c,0.659s rejects the same exact producer leak with
"production projection disclosed bearer or provider input". Task-owned mutation
removed immediately; restored native producer checksum compared with exact
upstream before final integration checks.

Final corrected whole affected PG18.6 integration/race PASSaf8c1a,23.282s;
integration vet PASSc5e9d9 exit0. Restored upstream outbox/persistence bytes match
fc1e6e8232f77c4a82a51fb78f0a5281f23a9b0f1fcce81e1e2c34c7f3bcb76d.
Final five-path snapshot0b3ece differs only in the corrected oracle file
(012fd923ab7fc356514c4c648039f13ebfaab25460df89d908090b763fe86128).

### Final independent production-effect review, verbatim

**APPROVE — No findings.**

**Native principal:** `/root/captcha_production_effect_review_a04`

**Reviewed boundary:** complete five-file dirty diff, including untracked content, against base/HEAD `a04ba44f48e7856da264d37699c944cd06725f3b` in `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`, branch `feature/identity-platform`. Read-only; no mutations, test execution, or additional agents.

The prior finding is resolved: the redaction oracle now checks decoded `canonical_record` and outbox `payload` alongside persisted row fields. ASCII synthetic tokens and opaque references remain searchable after JSON encoding. The supplied real-producer fault reproduction establishes that the old oracle falsely passed and the corrected oracle rejected the same leak. Archived audit/outbox producer hashes match the restored originals.

The complete diff retains meaningful assertions for six trusted-flow variants, current admission/disclosure authorization, exact replay and fresh-instance historical recovery without provider reevaluation, real commit-ack loss, production write-prefix rollback, and exact evidence/effect/binding cardinality. Shared production migrations and owned database/role cleanup remain intact. Public runtime APIs, schemas, and existing consumers are unchanged.

**Verification:** independently observed `git diff --check` PASS. Parent-reported final `TestIssuedEvidenceSameWorkIntegration` race PASS **23.282s** and integration vet exit **0**; not rerun during this read-only review.

**Remaining boundary:** approval covers this local development batch—not live-provider interoperability, complete protected-login journeys, unit-release gates, publication, or deployment.

### Final SHA-256s

```text
risk/postgres/CHANGELOG.md
b09ff1e2cb5cff50dd13a9432a87671c5e2dcbcc35057792ac81983e63df3f12

risk/postgres/README.md
9de25dcab7677f43c750af9978b0b338ade1c0f7b6356deeec7c0d8be3ccdf81

risk/postgres/captcha_executor_integration_test.go
012fd923ab7fc356514c4c648039f13ebfaab25460df89d908090b763fe86128

risk/postgres/issuance_integration_test.go
314ff8bd3948975eb9462504e11f6ddcb52e32e08e5a953d28e0607fc4e06654

risk/postgres/production_effects_integration_test.go
d25be681cf476b9d7c6ff690446464420236a08566af258e4404788cf8e56d29
```

Reviewed source committed locally as0d4d7bb; unit counts unchanged. No runtime
provider/login/release completion claim is made by this integration checkpoint.
