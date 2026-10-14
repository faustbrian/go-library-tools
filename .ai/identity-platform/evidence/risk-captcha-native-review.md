# CAPTCHA native custody development checkpoint

## Native exact-head delivery

Native principal: `/root/captcha_native_delivery_final`.

**APPROVE — local development checkpoint only.**

Exact committed head `bb4eb68d4484aa090f372a5fd2f3050e77abaf4f` is clean; its six paths match native reviewed digest `508338395192ab6ae3e41c8641900b271483f32ecff564aaaf5978c63e240be9`.

Inspected native mapper/QA and final compatibility/regression/correctness reports. Earlier findings are retained and superseded; no unresolved native finding remains.

Runtime artifact:
- All 129 source inputs and 651 temporary dependency inputs match; workspace hashes match.
- Fingerprint independently recomputes to `4caa887dc01b9bf81161d578b0b9f18cfe39d50fef567e0db141f77fb8e74604`.
- Race, vet and both bounded fuzz runs have terminal exit `0`.
- Original execution remains `80fa92cb…` plus dirty delta—not falsely attributed to the new commit.
- Disposable build cache absent; scratch remains for primary cleanup after this audit.

Control head remains `256f4725…`; two new evidence files are uncommitted. No push, release, durable PostgreSQL/protected-use/provider verification, permanent-pin compatibility, exact coverage/mutation or completed-unit claim is supported. Counts remain **5/67 integrated, 1/67 verified**.

## Native final correctness

**APPROVE — no findings for this development batch.**

- Principal: `/root/captcha_native_correctness_final` (`gpt-6.1-sol`, medium; independent read-only pass).
- Source: `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`, `feature/identity-platform`.
- Base/head: `80fa92cb3afbe41fac409e83c47d37a846aac281` plus the complete six-file dirty delta.
- Verified delta digest: `508338395192ab6ae3e41c8641900b271483f32ecff564aaaf5978c63e240be9`.
- Control head: `256f4725fb7c2dd17787e9b6b143e8ebc66fe8e5`.

**Source-backed closure**
- The extracted ordinary intent fields preserve the original field order, score sentinel, section counts, token commitment and MAC domain. Existing bearer signing and lookup derivation remain unchanged.
- Native signature authentication covers complete flow, command, original aliases, configuration and intent commitments, proof, nonce, hostname and original interval. The nested digest-lane commitment rejects a substituted digest key even when signing key/version match.
- Restoration bounds variable-width metadata before MAC construction, rejects malformed flows/aliases/time intervals, copies alias custody and reconstructs the same candidate at the same primary time.
- Issuer restoration checks policy admission, current configuration commitment and retained original key versions; new retained keys do not regenerate original aliases.
- Restored intents cannot enter `Issue()`. Recovery preserves Missing/Denied/Failed/Unknown/Committed classifications without provider verification or replacement issuance; noncommitted results reveal no evidence.
- Tests observe exact original candidate conservation, valid-field tampering, signed malformed input rejection, digest-only key mismatch, rotation, expired historical recovery, copy isolation, concurrent reads and external public construction. Doubles are correctly identified as core-boundary evidence, not database proof.
- Adversarial challenges closed: mixed original/tampered metadata fails signature verification; changed digest key fails authentication; expired restoration cannot materialize new evidence; Missing recovery cannot re-enter issuance.

**Verification**
The native runtime artifact’s complete `identity.inputs` map matches current files: **zero mismatches**. It records exit 0 for:
- `go test -race . -count=1` — 1.710s.
- `go vet .`.
- `FuzzCaptchaNativeProjection` — 3.283s.
- `FuzzCaptchaNormalizedBindingAndScore` — 3.658s.

No Go commands, writes, external operations or children were performed in this review. Supplied RED/GREEN session references were not independently replayed.

**Claim boundary**
Approval supports the bounded local development batch only. PostgreSQL command mapping/atomic issuance, protected consumption, provider interoperability, permanent dependency pins, exact coverage/mutation and remaining release/platform gates are unverified and remain required before broader readiness or completion claims.

## Native final regression

**APPROVE — native regression coverage audit**

- Role: `/root/captcha_native_regression_final`, fresh read-only blind audit.
- Reviewed HEAD/base: `80fa92cb3afbe41fac409e83c47d37a846aac281`, plus complete six-file working delta.
- Verified scope digest: `508338395192ab6ae3e41c8641900b271483f32ecff564aaaf5978c63e240be9`.
- Findings: no concrete behavioral regression or material coverage defect found.

Tests exercise tokenless restoration and exact candidate equality; closed flow variants; exclusive validity; independently authenticated projection fields and digest-key mismatch; malformed signed input; redaction; defensive copies/concurrent restoration; retained-key rotation; recovery-only outcomes without provider verification, reissuance, renewal, or noncommitted disclosure. The fuzz oracle compares exact original bytes independently of production validation.

Runtime artifact now contains four terminal successful runs, with all six source hashes matching current files:
- `go test -race . -count=1`: pass, 1.710s.
- `go vet .`: exit 0.
- Native projection fuzz, 3s: pass, 3.283s.
- Normalized binding/score fuzz, 3s: pass, 3.658s.

Unverified boundaries: durable PostgreSQL/shared-Work atomicity, provider truth/integration, protected one-use consumption, exact coverage/mutation and other release gates. This approves the bounded native regression slice, not package or platform completion.

## Native final compatibility

Principal: `/root/captcha_native_compat_current` — independent read-only Go compatibility audit.

**APPROVE** for the current complete six-file addition against `80fa92cb3afbe41fac409e83c47d37a846aac281`. Recomputed digest:
`508338395192ab6ae3e41c8641900b271483f32ecff564aaaf5978c63e240be9`.

No unresolved material compatibility finding:
- Native signature now authenticates both key lanes: digest-key MAC over every prepared field is included in the signing-key MAC. Digest-only key substitution rejects; its focused regression tests that independent substitution.
- Existing intent-MAC field ordering/framing/domain, bearer signature, reference syntax and lookup derivation remain unchanged.
- Public construction remains additive. Restoration preserves original candidate identity, aliases and exclusive expiry; restored intents remain recovery-only.
- The changed unreleased native signature introduces no demonstrated released-protocol compatibility obligation.

No Go commands or writes performed. Runtime execution, permanent dependency/clean-consumer compatibility, authoritative rotation history, durable coordinator integration and release/platform readiness remain unverified. This confirmation supersedes my earlier digest-bound report.

Primary reproduction/resolution of digest-lane finding: session87072 focused
`TestCaptchaNativeRejectsDigestOnlyKeyMismatch` failed with "digest-only wrong
key restored native custody" before the correction. The prepared signature now
authenticates a separately domain-separated HMAC over the same projection under
the digest key, as well as the existing signing-key HMAC. Session3900 complete
focused native/public-native tests passed0.221s. No existing bearer/intent MAC
format changed. Corrected six-file digest:
`508338395192ab6ae3e41c8641900b271483f32ecff564aaaf5978c63e240be9`.
The final runtime rerun and fresh independent reports determine delivery;
earlier approvals do not certify this correction.

## Native regression after flow consolidation

Native principal: `/root/captcha_native_regression_current` (`regression_reviewer`); independent read-only blind pass.

Reviewed source: `80fa92cb3afbe41fac409e83c47d37a846aac281` plus six-file delta, independently verified digest `7b5adc6d324b113ee2d63b6cf86a224f659ab50c237c8228621366cbace379a8`.

**Disposition: REQUEST_CHANGES**

**P2 — Digest-only wrong key restores successfully and changes the candidate.** [captcha_native.go](/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform/risk/captcha_native.go:75) authenticates only `k.signing`; it never authenticates `k.digest`. Yet [captcha_issuance.go](/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform/risk/captcha_issuance.go:355) uses `k.digest` to construct `EvidenceDigest`.

Smallest witness using `captchaNativeFixture`:
```go
c, p, original := captchaNativeFixture(t, captchaFlowInput())
wrong := c.Keys[0]
wrong.digest[0] ^= 1 // Same version and signing key.
restored, err := wrong.RestoreCaptchaIssue(p)
```
`err` is nil by source inspection: every checked value and the signing MAC are unchanged. `restored.Materialize(original.IssuedAt)` succeeds but recomputes `EvidenceDigest` under the wrong digest key, violating exact-candidate restoration and the documented wrong-key rejection. The public equivalent constructs version `1`, signing bytes `0x41 × 32`, and different digest bytes with `NewEvidenceKey`.

Existing wrong-key coverage changes both keys ([captcha_native_test.go](/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform/risk/captcha_native_test.go:230)), so it misses this defect. Add a digest-only key mismatch rejection oracle and ensure restoration authenticates both key identities. Parent reproduction remains necessary; I did not execute Go commands.

Other inspected oracles meaningfully cover exact-byte projection authentication, independent copied custody, exclusive validity, recovery-only outcomes, retained-key rotation, and no provider/issuance retry.

Persisted runtime evidence now contains exit `0` for `go test -race . -count=1`, `go vet .`, and both three-second CAPTCHA fuzz runs. All six reviewed file hashes match that evidence. These passing checks do not exercise the finding above.

No PostgreSQL atomicity, durable commit, protected consumption, provider truth, public release readiness, or complete-platform claim was audited or established.

## Native compatibility after flow consolidation

Principal: `/root/captcha_native_compat_current` — independent read-only Go compatibility audit.

**Disposition: APPROVE** for this bounded native-restoration addition; no material API or compatibility defect found.

Reviewed source HEAD/base `80fa92cb3afbe41fac409e83c47d37a846aac281` plus the six supplied dirty paths. Recomputed digest matches:
`7b5adc6d324b113ee2d63b6cf86a224f659ab50c237c8228621366cbace379a8`.

Control HEAD confirmed `256f4725fb7c2dd17787e9b6b143e8ebc66fe8e5`; authoritative core-risk goal resolves to `.ai/identity-platform/goals/identity-risk.md`.

Findings:
- Public additions are additive; existing issuer/executor signatures and bearer syntax remain unchanged.
- Extracting `captchaIntentFields` preserves the original ordered fields, section counts, token commitment, MAC domain and key lane. Existing command-intent identity remains compatible.
- Signed restoration preserves complete flow, original replay aliases, evidence-key version, intent/configuration commitments, proof, nonce, hostname and exclusive interval.
- Issuer restoration admits retained historical keys and additional newer keys without rewriting custody; configuration/policy changes and removed referenced versions reject.
- Restored intents cannot call `Issue`; `Recover` remains resolve-only, including Missing and expired historical committed results.
- The external-package fixture demonstrates public construction without private access. No non-test consumer of these CAPTCHA types exists outside the issuer/native files in the inspected repository.
- README and changelog correctly distinguish provisional integrity from durable commitment and protected-use authority.

**Unverified boundaries:** no tests/builds executed by this audit; permanent dependency resolution, clean consumers, API-baseline tooling, release gates, provider interoperability, authoritative rotation history, PostgreSQL atomic issuance/reservation/finalization, and complete risk/platform composition remain unproven. The temporary capability baseline is not release evidence.

## Native correctness report before consolidation

Principal: `/root/captcha_native_correctness_80fa`  
Mode: single-reviewer structured, read-only complete-diff pass.  
Base/head: `80fa92cb3afbe41fac409e83c47d37a846aac281` plus supplied six-file dirty scope.  
Verified scope digest: `5fff7a3f89b06067e8a1c2b70b662dfc0494bd3ebb2845bd2bcc7eed161a20fd`.

**Disposition: REQUEST_CHANGES** — one non-blocking source finding remains; repository policy requires confirmed findings resolved before delivery. No material native-custody defect found.

**Non-blocking — `risk/captcha_native.go:76`:** The second `NewFlowContext()` error branch is unreachable. The dominating `validPreparedCaptcha()` call already constructs the same scalar-only flow successfully at lines 163–165; no intervening operation changes it. Restoration therefore validates/allocates the same flow twice and adds an unexercisable return branch. Consolidate flow validation/construction into one result consumed by restoration, retaining rejection before MAC-input construction.

Falsification: traced the early guard, signature calculation, and both constructor calls; their identical copied input and deterministic constructor disprove a reachable second-call failure.

Publication body:
> Non-blocking: `validPreparedCaptcha()` already requires `NewFlowContext()` to succeed for these exact copied fields, so this second error branch cannot execute and the flow is constructed twice. Consolidate validation and construction so restoration reuses the validated flow without introducing an unreachable rejection branch.

Other reviewed boundaries:
- Signed projection covers every flow dimension, original aliases, commitments, proof, nonce, hostname and validity.
- Copies isolate alias slices; restored intents cannot enter `Issue()`.
- Historical recovery retains exact candidate identity and exclusive expiry.
- Existing non-native intent commitment bytes remain unchanged by the helper extraction.
- Direct key restoration deliberately grants provisional custody only; issuer restoration adds current configuration/policy and retained-version checks.

Evidence: supplied final runtime artifact records passing race, vet and both fuzz targets; independently matched all **129 source-input hashes** against the reviewed checkout. No Go commands executed in this pass.

Remaining unverified boundaries: permanent dependency-pin composition, PostgreSQL command mapping/atomicity/single-winner issuance, protected one-use consumption, providers/interoperability, complete risk Service, exact coverage/mutation and remaining release gates. This is not unit, release or platform readiness.

Primary resolution: consolidation returns the validated FlowContext to restoration;
the redundant issuer structural pass was also removed. Focused native tests passed
in session33052. Current source digest and rerun runtime replace the prior scope
for the forthcoming complete final-diff review; the native report above is retained.

Primary classification: behavioral — tokenless native restart recovery.
Secondary: public API, security, retry/replay, native persistence boundary,
and immutable concurrency custody. Core only; no durable commit claim.
Source base `80fa92cb3afbe41fac409e83c47d37a846aac281`.
Control baseline `256f4725fb7c2dd17787e9b6b143e8ebc66fe8e5`.
All independent roles: default GPT 6.1 Sol, medium, fork none, read-only.

Baseline session49947 original CAPTCHA issuance test passed0.221s.
First native restart test session21366 failed before production with undefined
PreparedCaptchaEvidence, CaptchaEvidenceIssue.Prepared and
EvidenceKey.RestoreCaptchaIssue. Session30380 same native restart test
passed0.211s after implementation. Expanded CAPTCHA tests session7326
passed0.199s; complete risk tests session82294 passed0.288s. Added partitions
are not individually claimed as test-first observations. Final native runtime
inputs/results are checkpointed separately in risk-captcha-native-runtime.json.
The task-owned test workspace retains the previous explicit capability
development baseline2985d8e; permanent dependency reconciliation remains pending.
No permanent replacement, package worker, new ledger or formal receipt.

## Native mapper

Native principal: `/root/captcha_native_map_80fa`.

Native report: `change_mapper`, read-only, no writes/tests/services/children.

Reviewed source HEAD `80fa92cb3afbe41fac409e83c47d37a846aac281` (clean); control HEAD `256f4725fb7c2dd17787e9b6b143e8ebc66fe8e5`.

### Actual missing boundary

CAPTCHA provisional issuance and intent are process-local pointer custody. `CaptchaEvidenceIssue` exposes only `Intent` and `Materialize`; there is no signed native preparation/restoration contract. `CaptchaEvidenceIssuer.Recover` requires an intent owned by that issuer. Existing restart/rotation tests recreate it with `Prepare(request, rawToken)`, so they **do not prove tokenless restart recovery**.

`risk/postgres` currently has no CAPTCHA command mapper, executor, schema, replay-alias constraint, or typed consumption contributor. Its repository explicitly documents CAPTCHA as unsupported.

### Existing causal slice and owners

- Core: trusted `FlowContext` → core-derived replay aliases → `CaptchaEvidenceIntent` → `Resolve` before verification → admitted normalized provider facts → provisional issue → `Execute` → authenticated committed record/result.
- `Recover` resolves only; Missing, Failed, Denied, Unknown disclose no evidence and never verify/resubmit.
- Materialization conserves original nonce/proof/exclusive expiry and derives bearer/digest using the original evidence key and authoritative database time.
- Replay identity owns token uniqueness independently of command/flow. All retained aliases identify one token; tenant namespaces durable uniqueness.
- Provider modules own bounded verification facts, not issuance decisions or durable transitions.
- Risk PostgreSQL owns CAPTCHA issue/reconcile and subsequent reserve/apply/finalize through the **existing identity coordinator and Work**, not another ledger.
- Downstream protected workflows include password, username, magiclink, OTP, phone, anonymous, passkey, OAuth and SSO operations in `API_OPERATIONS.md`. CAPTCHA verification is middleware/once; its internal command identity must not become a new public idempotency transport.

### Reusable existing boundary

Phone issuance already provides:
`EvidenceIssue.Prepared` → signed `PreparedEvidence` → `EvidenceKey.RestoreIssue` / issuer restoration → `risk/postgres.MapIssue` → closed native decoder → shared-Work `Materialize(DatabaseTime)`.

The key-level restoration avoids issuer/executor construction cycles. Authentication restores provisional custody only; current policy, authorization, Work scope and durable commitment remain independently owned checks. Phone `PreparedEvidenceUse` similarly distinguishes native integrity claims from current-use authority.

### Smallest complete consumed contract

For this **risk-core bridge slice**, expose authenticated, bounded native custody that permits an official adapter to:

1. Persist/reconstruct original command and exact trusted flow, provider/site/profile/configuration, policy version, replay aliases, original intent commitment and evidence key version.
2. Preserve proof digest, hostname, nonce, preparation time and original exclusive expiry.
3. Restore provisional issue and recovery intent without token, provider callback, clock read, configuration-derived renewal or regenerated randomness.
4. Reproduce the identical record at identical database time; authenticate expired historical records while rejecting fresh materialization at expiry.
5. Preserve Missing/Denied/Failed/Unknown/Committed outcomes and exact committed originals across process restart and retained-key rotation.
6. Reject altered scope, alias lists, commitments, keys, times and signatures before native execution.

The projection must authenticate its entire variable-width alias/flow structure, remain explicitly redacted, copy caller-owned slices, and reject oversized/malformed inputs before MAC work. It must not grant authority merely by reconstructing opaque custody.

**Important rotation interaction:** current intent digests deliberately exclude replay-alias additions by using per-evidence-key token commitments. Preserve that continuity; changing aliases must not accidentally change an existing command’s authenticated intent or authorize omitted unresolved aliases.

### Contradictions / claim limits

No contradictory requirement blocks the core bridge. Provider-neutral `VerifyRequest` eventually requires the core-derived fingerprint, whereas the current core callback lacks that projection; this is a later verifier-consumer seam, not a reason to manufacture durable provider state.

A signed restoration bridge alone cannot establish completed native CAPTCHA issuance or one-use consumption. Those claims still require the official PostgreSQL consumer, durable single-winner aliases/tombstones, authorized recovery disclosure and shared-Work transitions. Missing unrelated provider/release gates do not block this scoped first write.

## Native compatibility — original scope report

Native principal: `/root/captcha_native_compat_80fa`.

Role: independent Go compatibility specialist
Reviewed HEAD/base: `80fa92cb3afbe41fac409e83c47d37a846aac281` plus six-file working delta
Verified scope digest: `b19404f120919e40218b954cbc27a58e9f5f2de57515836945d1fc16456ccbc7`
Disposition: **APPROVE**, limited to this native recovery delta. No material compatibility findings.

- Public changes are additive: signed projection, `Prepared`, and key/issuer restoration methods. Existing interface signatures and ordinary intent paths remain unchanged.
- Extracted intent-field construction preserves original ordered MAC fields, domain, section counts, and final independently keyed token commitment. Replay and bearer framing are unchanged.
- Projection aliases are copied on export/restoration. Private native state is immutable after construction; no new dependency, goroutine, store, transaction authority, or module cycle is introduced.
- Restored custody retains original key version, aliases, proof, nonce and exclusive expiry. Same primary time reproduces the original candidate. Historical committed recovery authenticates the original record without renewing present-use validity. Restored intents cannot `Issue`.
- Issuer restoration checks original configuration commitment, policy version, scope and retained replay-key versions; added newer keys do not rewrite historical custody.
- Source uses APIs compatible with the existing Go 1.27 baseline. Public external-package test demonstrates construction without an issuer/executor cycle.

Unverified release boundaries: no Go commands executed by this reviewer. Supplied runtime artifact records passing race, vet and bounded fuzz runs against temporary capability baseline `2985d8ea…`; it does **not** prove permanent pinned dependency compatibility. Permanent-pin reconciliation, clean consumer/API baseline/reverse-dependant checks, exact coverage/mutation and other release gates remain pending. Official provider interoperability, PostgreSQL command mapping, atomic issuance and durable protected consumption are not established by this delta.

## Native compatibility — final scope confirmation

Independent Go compatibility confirmation: **APPROVE**, no material findings.

HEAD/base remains `80fa92cb3afbe41fac409e83c47d37a846aac281`; verified six-file digest is `5fff7a3f89b06067e8a1c2b70b662dfc0494bd3ebb2845bd2bcc7eed161a20fd`.

Corrected godoc accurately includes authenticated native restoration and recovery-only custody. Original compatibility assessment and all unverified release boundaries remain unchanged. No Go commands executed; the pending runtime rerun is not certified by this confirmation.

## Native regression

Native regression report — **APPROVE** for bounded core development.

- Principal: `/root/captcha_native_regression_80fa`, fresh read-only regression reviewer.
- Source base/head: `80fa92cb3afbe41fac409e83c47d37a846aac281` plus six-file working delta.
- Control baseline: `256f4725fb7c2dd17787e9b6b143e8ebc66fe8e5`.
- Independently verified digest: `5fff7a3f89b06067e8a1c2b70b662dfc0494bd3ebb2845bd2bcc7eed161a20fd`.
- No concrete behavioral or regression-test defect found.

Tests exercise token-free reconstruction, exact original candidate equality, all closed flow variants, authentication of every projection field, exclusive expiry, retained-key historical recovery, no reverify/resubmit, immutable copies, redaction and external public construction. The original record and exact-byte fuzz expectations provide meaningful outcome oracles.

Final persisted runtime inputs match the four changed executable/test files. Recorded results: `go test -race . -count=1` passed; `go vet .` passed; both three-second native-projection and normalized-binding fuzz runs passed.

Unproved boundaries remain explicit: real PostgreSQL command mapping/atomic single-winner issuance, protected reserve/apply/finalize, authorization and rotation-history enforcement, provider interoperability, permanent dependency reconciliation, canonical composition, exact release coverage/mutation and full risk/CAPTCHA completion. This disposition certifies none of those.

## Native QA

## Native QA report

- Principal: `/root/captcha_native_qa_80fa`, read-only `qa_test_planner`.
- Source head: `80fa92cb3afbe41fac409e83c47d37a846aac281`; clean checkout verified.
- Requirements head: `256f4725fb7c2dd17787e9b6b143e8ebc66fe8e5`; verified.
- Skill: completely read `/Users/brian/.dotfiles/.codex/skills/qa-test-planner/SKILL.md`.
- Read: complete core-risk/CAPTCHA goals and `COMMON_REQUIREMENTS.md`; shared-issuer amendment; transaction foundation and CAPTCHA protocol; relevant reference/applicability/public-contract catalog; CAPTCHA issuance/replay and phone-native restoration/mapping sources and tests.
- No writes, Go commands, services, or child agents.

**Disposition: APPROVE proceeding with implementation.** No contradictory requirement or unsafe requested behavior identified. This approves the bounded coverage plan, not implementation correctness or release readiness.

### Established gap

Current CAPTCHA restart examples reconstruct intent by calling `Prepare` with the original raw token. `CaptchaEvidenceIssue` has no phone-equivalent authenticated native projection/restoration path. Existing `Recover` accepts only intent belonging to the same issuer instance. Consequently, existing tests do not prove token-free process-restart recovery.

### One bounded coverage artifact

| Concrete risk | Required observable assertion | Proving boundary / independent oracle |
|---|---|---|
| Restart renews or changes provisional authority | Obtain issue from real core `Issue`; retain native projection, discard original issuer/issue, reconstruct with fresh startup custody and no token. At identical primary time, reconstructed `Materialize` equals the original record **in every field**, including reference, nonce-derived identity, intent/proof/lookup digest, hostname and times. | Core public bridge; original pre-restart record is the equality oracle. Verification callback must fail if invoked during restoration/recovery. |
| Reconstruction drops flow dimensions | Exercise pre-auth subject, pre-auth anonymous flow, authenticated subject/session and administrator actor contexts. Restored bindings preserve exact command, tenant, operation, subject/anonymous flow, pre-auth/session/actor, challenge and request fingerprint. | Table-driven real core preparation and restoration; original trusted flow and binding are independent expected values. |
| Unsigned or partially signed native fields grant authority | Mutate each behavior-bearing field independently: command, every flow dimension, provider/site/profile/configuration, policy version, key version, token commitment, replay alias, proof, nonce, hostname, prepared time and expiry. Each mutation is rejected with zero usable issue. | Signed projection tamper tests; mutations must remain structurally admissible where possible so rejection proves authentication rather than fixture invalidity. |
| Native restoration accepts malformed inputs before bounded validation | Zero, malformed command/flow, unsupported operation/provider, invalid digest width/encoding, duplicate or invalid alias versions, excessive collection/string sizes and invalid time precision/interval are rejected deterministically. Wrong key and absent retained key reject. | Constructor/decoder boundary; stable redacted rejection and zero result. Include signed-but-invalid cases where the contract exposes a way to create them. |
| Restoration extends exclusive validity | Materialization rejects before original `PreparedAt`; accepts at `PreparedAt` and expiry minus one microsecond; rejects at exact expiry and after. Round-trip preserves original earlier provider/challenge expiry, not configured lifetime from restart. | Fixed clock and exact microsecond inputs; previously recorded original expiry is oracle. |
| Rotation loses historical custody or silently substitutes newest key | Add newest evidence and replay keys while retaining originals. Old native custody still authenticates and yields exact old candidate/result; a newly prepared issuance uses newest versions. Removing the referenced evidence key rejects restoration. No new replay alias changes the original command commitment. | Fresh configuration instances; old record/bindings are oracle. Retention legality remains adapter responsibility. |
| Historical replay is treated as fresh issuance | Reconstructed custody can authenticate an expired **committed historical result**, returning the exact original public result. It cannot materialize new evidence after original expiry. Missing/Denied/Failed/Unknown recovery returns zero evidence and never verifies or executes anew. | Core execution/result boundary with deterministic coordinator fixture; original committed result and zero-evidence classifications are oracles. |
| Native projection aliasing changes frozen authority | Mutate returned slices/nested replay metadata and any exposed byte storage; subsequent original projection, restoration and materialization remain unchanged. Concurrent immutable projection/restoration reads remain deterministic. | Copy-isolation test plus focused race run. |
| Custody export leaks replay or bearer material | Formatting is redacted; automatic JSON/text export rejects for each new native custody type. Errors contain neither sentinel token nor flow/profile/digest values. Projection contains no raw token/provider payload/score/reason text/remote address. | Runtime formatting/serialization/error assertions and typed-field inspection; no source-text test substituted for behavior. |
| Core bridge gets mistaken for durable commit | Restored provisional custody alone never produces a committed public result; only an attested executor result can. A malformed committed record is rejected; all noncommitted outcomes require zero record. | Existing core executor seam; positive and negative public results. Real PostgreSQL atomicity is **not** established by this fixture. |

### Minimal sufficient verification

1. **Red reproduction:** add a token-free restart round-trip through the intended public bridge; observe its expected failure before production implementation.
2. **Focused new tests:** native round-trip, field-authentication/structural bounds, exact validity boundaries, retained-key rotation/historical results, aliasing/redaction and all closed flow variants.
3. **Regression suite:** all existing `TestCaptcha*` tests, including replay framing, token substitution, concurrent winner reconciliation, callback validity horizon, cancellation and zero-disclosure recovery; phone-native round-trip tests if shared key/restoration behavior changes.
4. **Focused race:** new bridge tests plus existing CAPTCHA concurrency/custody tests.
5. **Clean external consumer:** construct/restorе valid native custody using public APIs only; zero custody rejects. No consumer-written store or concrete PostgreSQL dependency in core.
6. **Affected existing repository gates:** execute or validly reuse only gates whose complete inputs changed. This plan adds no new harness or receipt gate.

**Scope boundary:** these checks prove the core native custody/recovery contract. They do not prove PostgreSQL replay-alias uniqueness, single-coordinator atomicity, primary reconciliation, protected-use reservation/finalization, provider interoperability, or full risk/CAPTCHA completion. Those require their owning adapter/composition evidence when implemented; they do not block this bounded first write.
