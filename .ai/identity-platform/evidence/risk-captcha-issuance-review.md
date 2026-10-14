# CAPTCHA core issuance development checkpoint

## Native local delivery audit

**APPROVE — local committed development checkpoint only.**

Auditor: `/root/captcha_issuance_delivery_final`, read-only. Source is clean on `feature/identity-platform`, HEAD `80fa92cb3afbe41fac409e83c47d37a846aac281`, parent `2b3b51def584feb62559a2f24d9852881d7b8634`. Exactly five reviewed paths changed; independently recomputed digest matches `84ab404c907e103b861a62ed9b2c3d0efe470e864377c3a1fa561eb90d80fd0e`.

Inspected retained native reports:
- Mapper `/root/captcha_issuance_map_2b3` and QA `/root/captcha_issuance_qa_2b3`: source `2b3b51d`, control `4c5113c`; no blocking contradiction for bounded core slice.
- First regression `/root/captcha_issuance_regression_2b3`: `REQUEST_CHANGES`, digest `510fb868…`; callback-completion timestamps extended validity.
- Corrected regression `/root/captcha_issuance_regression_final`, compatibility `/root/captcha_issuance_compat_2b3`, correctness `/root/captcha_issuance_correctness_final`: `APPROVE`, corrected digest `84ab404c…`, no unresolved material findings. Correctness report discloses incidental exposure to prior conclusions after its independent sweep.

All 126 recorded source-input hashes match committed source. Recomputed native runtime fingerprint matches `5e1d5a96b12657216c72bb6f9ea4b1127440d4e66011d45e7f1ece2b17015e6d`. Recorded race tests, vet and bounded fuzz passed. Execution remains **`2b3b51d+dirty`, reused at `80fa92c`**, not a rerun.

States: **local implemented; source committed; push not evidenced; CI not verified; deployment not evidenced; production not verified**. Control evidence and updated ledger remain uncommitted.

Provider truth/native union, canonical Service/audit composition, restart reconstruction, PostgreSQL atomic issuance/tombstones, protected consumption, exact coverage/mutation and release gates remain pending. Counts remain **5/67 integrated, 1/67 verified**; no unit, release or platform completion certified.

## Final correctness review — native report

## APPROVE — local development scope only

**Principal:** `/root/captcha_issuance_correctness_final`, independent single-reviewer structured pass; no children, writes, Go commands, services, or external mutations.

**Repository:** `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`  
**Branch:** `feature/identity-platform`  
**Base/head:** `2b3b51def584feb62559a2f24d9852881d7b8634`, plus all five supplied dirty/untracked files.  
**Independently confirmed scope digest:** `84ab404c907e103b861a62ed9b2c3d0efe470e864377c3a1fa561eb90d80fd0e`.

**No findings.** Disposition derives from my complete current-source sweep, caller/sibling traces, test-oracle inspection, and native execution evidence.

Source-backed closure:
- Configuration, flow/profile matching, supported-field availability, score boundaries and exact bindings reject before dispatch.
- Original provider/challenge expiry and core invocation horizon constrain materialization; rounding cannot extend validity (`captcha_issuance.go:455–465,552–567`).
- Token commitment and retained integrity keys bind command recovery independently of rotating replay aliases (`:232–251,347–374`).
- Unknown/late executor outcomes disclose no evidence; unexpected committed candidates require exact primary reconciliation (`:470–495`).
- Recovery cannot verify or execute; expired original records remain historical results, not renewed authority (`:499–535`).
- Bounded input/key/list counts, immutable copies, deferred cancellation, and absence of locks/goroutines cover this core’s resource/concurrency ownership.

Falsification challenges included completion-based timestamp renewal, mismatched historical token/flow, exact expiry materialization, and a different concurrent committed candidate. Source guards reject or reconcile these states without renewing or disclosing unsupported authority.

**Execution evidence inspected, not rerun:** all 126 recorded repository-input hashes match. Native artifact records passing:
- `go test -race . -count=1`
- `go vet .`
- `go test . -run '^$' -fuzz '^FuzzCaptchaNormalizedBindingAndScore$' -fuzztime=3s -parallel=2`

Retained reproduction entries record missing-API RED/GREEN and invocation-horizon regression RED/GREEN.

**Claim boundary:** approval covers normalized core admission, provisional signed custody and executor-result conservation. Provider truth/native union, canonical Service and failure/audit attribution, native command restoration, PostgreSQL atomic single-winner/tombstones, protected consumption, complete coverage/mutation and release/composition gates remain unverified. This does not approve unit/platform completion or release readiness.

**Disclosure:** after completing the blind source/diff/causal sweep, an overbroad evidence search exposed prior review conclusion lines. They were not used to determine this disposition; this is ordinary local review, not a blinded replay.

## Local committed checkpoint

Source commit `80fa92cb3afbe41fac409e83c47d37a846aac281` contains the exact
reviewed five-file scope without post-review edits. Runtime execution remains
`2b3b51def584feb62559a2f24d9852881d7b8634+dirty`, not relabeled as a rerun.
Its complete source-input hashes match the committed checkpoint. Intent
Commits Core and message-line validation passed; no project profile applies,
effective impact unspecified. Local only, not pushed/published/deployed.

## Corrected regression review — native report

## Native regression report

- **Principal:** `/root/captcha_issuance_regression_final` — independent read-only `regression_reviewer`.
- **Reviewed boundary:** five-path CAPTCHA core issuance delta over `2b3b51def584feb62559a2f24d9852881d7b8634`.
- **Independently verified digest:** `84ab404c907e103b861a62ed9b2c3d0efe470e864377c3a1fa561eb90d80fd0e`.
- **Disposition:** **APPROVE** for this bounded development slice.
- **Findings:** no unresolved material regression findings.

**Previous finding resolved:** admission now caps expiry at the core-owned verification invocation plus configured lifetime, preserving earlier provider/challenge expiry. `TestCaptchaCallbackCannotMoveInvocationValidityHorizon` delays verification and supplies completion-based timestamps, then asserts the original invocation horizon. The retained native evidence records its expected RED failure and GREEN result.

Tests exercise actual issuance results, rejection before dispatch, score boundaries, original expiry, record corruption, changed-token/flow recovery, retained-key recovery, concurrent-winner reconciliation, unknown-result nondisclosure, callback budgets and custody redaction. The fuzz oracle independently states the hostname/action/score acceptance predicate.

**Execution evidence inspected, not rerun:** runtime fingerprint `5e1d5a96b12657216c72bb6f9ea4b1127440d4e66011d45e7f1ece2b17015e6d`; all 126 recorded source-input hashes match. Race tests, vet and bounded normalized-binding/score fuzzing passed.

**Unverified boundaries:** provider capability truth/native evidence union, canonical Service and audit composition, restart-native reconstruction, PostgreSQL atomic issuance/replay tombstones, protected reserve/apply/finalize, exact coverage/mutation and remaining release gates. Executor doubles establish core outcome conservation, not durable commit or single-winner behavior.

## Corrected compatibility review — native report

Principal/role: `/root/captcha_issuance_compat_2b3`, read-only Go compatibility specialist  
Disposition: **APPROVE** — complete corrected local development batch.

- Reviewed base/head: `2b3b51def584feb62559a2f24d9852881d7b8634`
- Independently confirmed five-file digest: `84ab404c907e103b861a62ed9b2c3d0efe470e864377c3a1fa561eb90d80fd0e`
- Findings: none material.

The complete API remains additive and externally constructible through public constructors and executor methods. Existing signatures, consumers, module paths and dependency pins remain unchanged; only standard-library production imports are added. Zero/foreign custody, copied configuration, bounded contexts, redaction, committed-only disclosure and retained-key recovery remain consistent.

The correction caps expiry at the core-owned verification invocation plus lifetime while preserving earlier provider/challenge expiry. It changes no exported signature and introduces no consumer compatibility regression.

Refreshed runtime evidence fingerprint: `5e1d5a96b12657216c72bb6f9ea4b1127440d4e66011d45e7f1ece2b17015e6d`. All recorded source-input hashes match. Recorded race tests, vet and targeted fuzzing passed; I ran no Go or service commands.

Not certified: canonical Service/provider/native PostgreSQL composition, durable one-use/restart recovery, public API baseline and clean-consumer/reverse-dependant release checks, exact coverage/mutation, or remaining release/security gates. Tokenless `Recover` requires an already-held intent; post-restart reconstruction is not proven.

## First regression review — retained native report

## Native regression report

- **Principal:** `/root/captcha_issuance_regression_2b3` — independent `regression_reviewer`, read-only.
- **Base/head:** `2b3b51def584feb62559a2f24d9852881d7b8634`, plus frozen dirty scope.
- **Scope:** `risk/captcha_issuance.go`, `risk/captcha_issuance_test.go`, `risk/captcha_issuance_example_test.go`, `risk/README.md`, `risk/CHANGELOG.md`.
- **Verified scope digest:** `510fb868c732224079afcb0ed6f177e28d6e92223a4ec948aa3bc5cb806f84f3`.
- **Disposition:** **REQUEST_CHANGES**.

### Finding: callback-completion timestamps can extend the issuance validity horizon

**Location:** `risk/captcha_issuance.go:443–455`, `:552–555`.

The core records `verifiedAt` before invoking verification, but admission only checks `ObservedAt <= callback completion` and `ExpiresAt - ObservedAt <= Lifetime`. Consequently, a verification callback taking one second can return `ObservedAt = completion` and `ExpiresAt = completion + Lifetime`; the executor can commit evidence expiring one second beyond the original verification-start horizon.

This misses the documented original-read/nonrenewal contract. Existing `TestCaptchaScoreThresholdAndOriginalValidity` advances callback time but deliberately preserves its initial observation, so it cannot detect completion-based renewal. The “future-observation” rejection only tests an observation later than completion.

**Smallest remedy:** constrain the provisional expiry by a core-owned verification-invocation horizon, while preserving any earlier provider or challenge expiry. This need not reject an adapter’s naturally later network-read timestamp. Add a deterministic regression with callback delay and completion-based timestamps; assert that committed expiry cannot exceed invocation time plus configured lifetime.

This conclusion is source-derived; no Go commands or reproducer execution were authorized for this audit.

### Evidence assessed

Runtime artifact fingerprint: `048beb974956e0cc7cf597a6d3157dc69d7e4757b574261ffd0f53e8f479c3c8`. Its recorded hashes match the five reviewed files.

Recorded successful executions:

- `go test -race . -count=1`
- `go vet .`
- `go test . -run '^$' -fuzz '^FuzzCaptchaNormalizedBindingAndScore$' -fuzztime=3s -parallel=2`

The binding/score fuzz oracle is independent of production helpers. Integrity/materialization tests legitimately exercise core custody, but their executor declarations do not establish database commit or single-winner behavior.

### Claim boundaries

Only local normalized admission, provisional candidate integrity and declared executor-result conservation were audited. Provider capability truth/native unions, canonical Service composition, audit attribution, native command restoration, PostgreSQL atomicity/replay tombstones, protected consumption, exact coverage/mutation and release gates remain unverified.

## First compatibility review — retained native report

Principal: `/root/captcha_issuance_compat_2b3` — Go compatibility specialist  
Disposition: **APPROVE** for the scoped local development batch.

Reviewed:
- Base/head: `2b3b51def584feb62559a2f24d9852881d7b8634`
- Five-file digest independently matched: `510fb868c732224079afcb0ed6f177e28d6e92223a4ec948aa3bc5cb806f84f3`

No material compatibility findings:
- Additive exported API; existing consumers, module paths, dependency pins and signatures remain unchanged.
- External consumers can construct configuration, implement the executor, obtain opaque intents and reconstruct persisted references using public constructors.
- Zero/foreign custody is rejected; configuration slices and threshold values are copied. Context, token borrowing, concurrency obligations, expiry and unknown-result semantics are explicit.
- Added production imports are standard-library only; no cross-module dependency or cycle is introduced.
- Canonical Service/contributor/provider contracts remain unfinished and are explicitly distinguished from this development composition.

Verification inspected, not executed: runtime artifact records passing risk race tests, vet and targeted fuzzing; every recorded source-input checksum matches the reviewed checkout.

Release boundaries remain unverified: canonical provider/native PostgreSQL integration, restart recovery, complete public-contract/API-baseline alignment, clean-consumer and reverse-dependant gates, exact coverage/mutation, and remaining release/security/supply-chain gates. `Recover` without a token currently requires an already-held intent; this report does not certify post-restart reconstruction.

## Reviewer finding reproduction

Session6435: `go test . -run '^TestCaptchaCallbackCannotMoveInvocationValidityHorizon$' -count=1`
failed with `provider completion renewed original invocation validity`.
Session52298: same command after invocation-horizon correction passed0.202s.
The final corrected scope digest is
`84ab404c907e103b861a62ed9b2c3d0efe470e864377c3a1fa561eb90d80fd0e`.
This correction preserves earlier provider/challenge expiry without rejecting
a legitimate later inner network-read timestamp. Updated final runtime results
remain in the adjacent native runtime artifact.

Primary classification: behavioral. Secondary: public-contract, security,
external-verification boundary, retry/replay and provisional native custody.
Source base: `2b3b51def584feb62559a2f24d9852881d7b8634`.
Control baseline: `4c5113c5e06901569fc2e17ba3a88c114b25a4a9`.
All independent principals use default GPT 6.1 Sol, medium, fork none,
read-only scope. No package worker or formal receipt workflow.

First focused test failed before production: session24961,
`go test . -run '^TestCaptchaIssuanceReturnsOnlyCommittedOriginalEvidence$' -count=1`,
exit1, undefined CaptchaEvidenceIntent/Execution/Issue APIs. First green:
session30231, same focused command, PASS0.204s. Additional tests subsequently
expand coverage; they are not individually claimed as test-first observations.
Expanded sessions39371/83959 PASS0.192s/0.208s. Terminal native final results
and complete pre-run identities are in `risk-captcha-issuance-runtime.json`.

## Native change mapper: /root/captcha_issuance_map_2b3

## change_mapper report

**Source head:** `2b3b51def584feb62559a2f24d9852881d7b8634`  
**Control head:** `4c5113c5e06901569fc2e17ba3a88c114b25a4a9`  
Read-only inspection; no writes, Go execution, or delegation.

### Current causal slice

- `risk/captcha_replay.go` already owns token bounds, exact prescribed HMAC bytes, closed flow-context admission, retained-key aliases, and opaque preparing-policy/flow provenance. `Prepare` accepts only an admitted protected operation whose selected profile requires `captcha`.
- `risk/policy.go` owns operation admission and decision precedence. `identity.risk.captcha-verify` is presently **DefaultDeny**; issuing evaluation must use the protected operation bound in `FlowContext`, not invent a CAPTCHA-operation exemption.
- `risk/evaluation.go` reads configured mandatory sources and returns non-issuing decisions. Its `Source.Read` seam carries generic `Context`, not CAPTCHA challenge/profile/flow provenance. Generic `Signal` currently cannot retain the required attributable CAPTCHA provider facts.
- `risk/issuance.go` supplies the closest reusable native custody pattern: opaque intent and provisional issue; same-command resolution before source work; signed prepared native projection; database-time materialization; integrity-checked committed execution; reference disclosure only after committed outcome.
- No CAPTCHA provider module, normalized CAPTCHA fact admission, issuer, durable CAPTCHA contributor, or consuming reservation implementation exists at this head. Existing generic CAPTCHA tests use trusted test sources; they do not prove provider-bound issuance.
- Eventual consumers are protected-operation workflows and HTTP challenge composition. Their protected mutation must later enlist the durable contributor for reserve/apply/finalize; possession of the issued reference alone must not grant permission.

### Exact next-behavior requirements

1. **Provider facts are not decisions.** Admission must distinguish solved/rejected/unavailable/malformed/throttled/cancelled/ambiguous verification and unsupported capability/binding failure. An adapter cannot select allow, deny, throttle, step-up, or fail behavior.
2. **Trusted configuration controls binding.** Provider, exact API/version/tier, site, configuration version, allowed canonical hostnames/origins, expected provider action, score capability/policy, and disclosure policy cannot be request-selectable. Missing required or unsupported configured bindings fail construction.
3. **Absence is not fabricated success.** Retain hostname/origin/action availability and match status, original challenge time, score availability/value, recognized reason classification, and transport outcome. hCaptcha `publisher_pro` and CaptchaFox `free` cannot fabricate action or score; reference profiles expose no origin.
4. **Closed bounded native evidence.** Exactly one discriminator-matching reCAPTCHA/Turnstile/hCaptcha/CaptchaFox member; no arbitrary maps, response bodies, or provider error text. Inputs and collection/string sizes must be bounded before allocation or hashing; scores must be finite and within `[0,1]`.
5. **Core owns issuance policy.** Protected ambiguity denies; low-risk ambiguity yields step-up only with independently configured/current factor authority, otherwise deny. Below selected score threshold denies. A step-up/deny/throttle or uncertain result cannot mint solved CAPTCHA evidence.
6. **Opaque replay preparation stays authoritative.** The new path must accept and validate the existing opaque `CaptchaReplayIdentity` against its original policy and exact `FlowContext`, never reconstruct custody from copied fingerprints or adapter output.
7. **Resolve before verification.** Same-command/same-binding committed replay returns its original safe result without provider calls or new issuance. Unknown resolves that same command; it cannot trigger a replacement command/reference or provider retry.
8. **Durable execution is the sole issuance authority.** Core provisional proof is not issued evidence. The injected native executor must journal evidence plus the command result atomically through shared Work. Failed proves no row; Unknown returns no reference.
9. **Freshness cannot renew.** Reference profile is two minutes and one use. Database issuance must respect original preparation/provider validity and reject backward/late time. Recovery preserves original issuance/expiry.
10. **Safe disclosure only.** Public result contains opaque reference, provider, original issued/expiry timestamps and one-use metadata. Raw token, IP, native facts, scores, `cdata`, proof/replay digests, signatures and journal IDs cannot escape through results, formatting, serialization, errors or audit.

### Reuse versus distinct CAPTCHA behavior

**Reuse phone custody:** immutable startup composition; independently keyed integrity/linkage custody; bound command intent; redacted opaque native values; resolve/execute/recover separation; bounded total context; signed prepared projection; database-time materialization; integrity-checked recovery; no core store, private ledger, provider retry, or concrete adapter dependency.

**Do not reuse phone-specific authorization:** phone issuing phases/purposes, carrier negativity, canonical number, phone reset operation restriction, and requirement for a full phone evaluation/counter snapshot do not define CAPTCHA verification. CAPTCHA binds exact pre-auth/authenticated/admin flow, challenge, request fingerprint, provider/site/profile/configuration and replay aliases. Its evidence enables later protected evaluation/consumption rather than authenticating a phone-reset phase.

**Additional durable obligation:** token reuse has one winner across every retained replay-key alias in the exact tenant/provider/site/profile/configuration namespace. Different command collisions deny without explaining the matched field. Tombstones survive payload erasure; unresolved commands never release by time; rotation cannot restore admissibility.

### Smallest coherent core batch

Implement **core-owned admission of bounded normalized facts into opaque, exact-flow/replay-bound provenance, plus issuance orchestration/custody through an injected native executor**, returning only committed safe metadata. Include same-command recovery that cannot verify or issue. Keep concrete verification, provider protocols, PostgreSQL execution, and reserve/apply/finalize implementation outside this batch.

This is a partial core capability, not canonical `Service`, provider interoperability, durable single-winner proof, or platform completion.

### Catalog tensions and disposition

- `PUBLIC_CONTRACTS.json` places `Service.Verify` in `identity/risk/captcha`, with constructor taking UnitOfWork and contributor; the goal calls that module a stateless verifier and forbids it owning durable roles. A facade may delegate core issuance without owning those roles, but a verifier that directly decides or privately persists would contradict the goal.
- Catalog contributor signatures mention `identitypostgres.Work`, while core ownership forbids importing concrete persistence adapters. Existing native executor separation is consistent with the ownership requirement; literal implementation of those signatures in core needs contract reconciliation before claiming catalog fidelity.
- Catalog `CaptchaEvidenceIssue` exposes string `ReplayFingerprint`; normative text prohibits caller/adapter authority over that string. Such fields are only valid **restricted native projections**, never public issuance inputs. Existing opaque replay preparation resolves that distinction.
- Reference replay “scope” lists tenant, but the transaction contract’s exact MAC omits tenant and places it in durable uniqueness. Existing implementation correctly follows those prescribed bytes.
- Canonical `captcha-verify` remains DefaultDeny. The new core seam must continue binding the protected operation rather than treating the catalog operation as its own allowed risk action.

**No contradiction blocks the bounded core batch**, provided it remains explicitly native/development composition and does not claim canonical facade/catalog or durable/provider completion.

## Native QA planner: /root/captcha_issuance_qa_2b3

## Native QA report

- Principal: `/root/captcha_issuance_qa_2b3` — `qa_test_planner`, read-only.
- Skill: `/Users/brian/.dotfiles/.codex/skills/qa-test-planner/SKILL.md`, read completely.
- Source base/head: `2b3b51def584feb62559a2f24d9852881d7b8634`.
- Control head: `4c5113c5e06901569fc2e17ba3a88c114b25a4a9`.
- Scope: CAPTCHA normalized-fact admission and core issuance custody/decision using existing opaque replay preparation and injected durable execution. Provider protocols/network, PostgreSQL implementation, and protected-action consumption are outside this batch.
- Disposition: **No blocking contradiction or unsafe requirement identified.** Proceed with the bounded coverage below.

### Contract established

Read the full core-risk and CAPTCHA goals, common requirements and shared-issuer amendment; CAPTCHA transaction protocol, reference-profile/configuration rows, and relevant public-contract rows. Traced existing `captcha_replay.go`, policy/evaluation, phone-issuance execution seams, and their relevant tests.

Current replay preparation already derives the prescribed MAC, retains exact immutable flow/policy provenance, and exposes copied native bindings. It neither decides issuance nor establishes durable authority. Existing generic CAPTCHA signals cannot preserve typed provider failure classification: `validSignal` rejects nonzero `Failure` outside HIBP. The new slice must therefore prove actual CAPTCHA normalization through its owned consumer boundary, not merely add separately tested facts.

### Bounded coverage plan

| Risk / realistic defect | Observable assertion |
|---|---|
| Forged, zero, foreign-policy or wrong-flow replay preparation substitutes public bindings | Admission rejects before provider-fact evaluation or durable issue submission. Change each flow binding independently, including operation, tenant, subject/anonymous flow, session/pre-auth/actor, challenge and request fingerprint. Mutating returned bindings cannot alter accepted provenance. |
| Provider facts are structurally accepted despite foreign profile/site or contradictory availability | Admit one valid fact set for each closed provider profile. Reject wrong provider/API/version/tier/site, wrong discriminator/member, unsupported fabricated action/origin/score, and unavailable fields containing values or asserted matches. Distinguish unavailable from present-empty. Preserve bounded failure/transport classification. |
| Provider success or match flags directly become authorization | Through the core decision/issuance entry point, solved and exact bound facts may issue; rejected or mismatched facts do not. Wrong hostname/action remains rejected even if supplied `Matched=true`. Facts remain facts, not a caller-selected decision. |
| Incorrect score or ambiguity policy | Score-bearing profile: selected threshold minus one representable step, exactly threshold, and plus one step produce distinct expected results. Reject NaN, infinities and values outside `[0,1]`. Unsupported score does not acquire a fabricated zero. Protected-operation unavailable/ambiguous facts deny; low-risk ambiguity yields step-up only with independently configured factor authority, otherwise deny. None returns evidence. |
| Delayed callbacks renew original validity | Freeze a trusted clock. Assert original observation/challenge validity survives admission, decision and submission unchanged. Just-before expiry may proceed; exact expiry, later expiry, future observation, backward clock and total deadline exhaustion cannot issue. Public committed expiry must not extend the original configured interval. |
| Provisional or corrupt execution becomes a bearer | Boundary executor double declares Missing/Denied/Failed/Unknown/Committed. Only validated Committed exposes opaque reference and safe provider/time/one-use metadata. Every other outcome has zero evidence. Reject foreign intent, noncommitted outcome with a record, altered reference/proof/binding/key/timestamps, and a Missing response where execution requires a terminal result. |
| Replay or unknown-outcome retry creates a replacement artifact | Same-command resolution returns the exact original result/reference/times, including historical expiry, without reevaluating facts or submitting another issue. Unknown, executor error, late acknowledgement and cancellation after dispatch expose no bearer. Recovery only resolves the original command; Missing recovery cannot submit or reevaluate. A retained winner returned after a race must be resolved as the original result, not accepted as unexplained fresh output. |
| Secret/native facts escape custody or mutable aliases change decisions | Formatting, JSON/text serialization, error chains and public result projection must not expose token, replay digest, flow identifiers, native record/proof, score, provider payload, reason text or bearer except explicit authorized reference reveal. Mutate input slices/configuration and returned projections after admission; owned state and decisions remain unchanged. |

### Smallest sufficient verification

1. Add focused deterministic public-boundary tests for the rows above; observe the expected red before production changes.
2. Use an injected executor boundary double that retains submitted intent/result and returns declared durable outcomes. Assert returned result and retained bounded projection, not call counts alone. Counts may additionally prove prohibited resubmission.
3. Run focused CAPTCHA admission/issuance tests together with existing CAPTCHA replay and affected policy/evaluator tests.
4. Run those tests under `-race`; include concurrent immutable reads/submissions and callback deadline checks. A fake may prove core outcome conservation, **not** database single-winner atomicity.
5. Add a bounded fuzz target for admitted fact/profile combinations, seeded with discriminator mismatches, contradictory optional fields, malformed lengths and nonfinite scores.
6. Compile one external-consumer example constructing every required opaque input through its public authority path; verify zero-value rejection.
7. Run affected module/reverse-dependant gates required by repository policy, reusing only fingerprint-proven unchanged evidence. Every Go execution requires a disposable task-owned `GOCACHE` and cleanup.

**Unverified boundaries:** actual provider truth/interoperability, PostgreSQL atomic issue/command commit, rotation-alias uniqueness/tombstones, and reserve/apply/finalize consumption remain separate required work. This scoped batch must not claim those outcomes or complete CAPTCHA/platform verification.

## Primary scope reconciliation

The core now implements normalized projection admission and provisional issuance
through the executor boundary. Provider-native union/capability truth, preserved
denial attribution into the generic risk/audit path, canonical Service/catalog
reconciliation, signed native command restoration and actual contributor/state
ownership remain implementation work. Current boundary tests cannot certify
those absent owners. Low-risk step-up remains the existing full risk policy's
independent-factor decision, never solved CAPTCHA issuance. Full risk coverage,
mutation, release/consumer/provider gates are still pending. No new unit is
marked integrated or verified; full 67-unit and reference goals remain active.
