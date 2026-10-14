# Core CAPTCHA replay preparation checkpoint

Source base/HEAD `f2b893075237fb213704a3222a837247b437bf72`, branch
`feature/identity-platform`, five-file dirty replay batch. Original scope digest
`fb3ec9cd0c2e3f2517db7ab26f9aa00f4297a6162385785f2d221cae490e1272`:
sorted path + NUL + SHA256(file bytes) + LF, then SHA256 of concatenation.
Native runtime results are in `risk-captcha-replay-runtime.json` with original
pre-run source/dependency/workspace identity; not full-unit or release proof.
All roles below are isolated read-only GPT-6.1 Sol medium, fork_turns=none.

## Native mapper report

**Native change_mapper report**

- Principal: `/root/captcha_replay_map_f2b`, independent read-only inspection.
- Source: `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`.
- Exact base/head inspected: `f2b893075237fb213704a3222a837247b437bf72`.
- No writes, tests, builds, or runtime access performed.

### Required core semantics

1. `identity/risk` alone derives replay identity from the bounded raw token and immutable trusted provider configuration. Caller/provider-supplied fingerprints cannot substitute.
2. Exact HMAC input, per `TRANSACTION_CONTRACT.md:445–460`:

   `ASCII("identity-captcha-replay-v1") || NUL || u32be(provider length) || provider UTF-8 || u32be(site length) || site UTF-8 || u32be(API/profile length) || API/profile UTF-8 || u32be(configuration-version length) || configuration-version UTF-8 || u64be(raw-token length) || raw-token bytes`.

   **Do not encode tenant, operation, flow, or key version into this prescribed byte sequence.** Tenant is the outer replay-uniqueness scope; operation/flow are separately bound trusted issuance intent.

3. Replay identity remains identical across different logical commands consuming the same token within the same provider configuration. Adding command/action/flow to the replay MAC would defeat the required single-winner boundary.
4. Newest configured digest key derives new issuance identity; retained key versions must remain available for old fingerprints/tombstones and unknown-command recovery. Rotation cannot silently reopen old-token admission.
5. Operation identity is closed canonical identity-operation membership; flow context is closed tagged trusted pre-auth, authenticated subject/session, or administrative actor context, retaining exact subject/anonymous-flow binding.
6. Token bound is `1..8192` secret UTF-8 bytes, further limited by immutable `MaxTokenBytes`; bounds precede hashing/allocation. Site configuration is bounded, with configuration version `1..64` ASCII bytes.
7. Automatic formatting/serialization, errors, response results, audit events, and journal results must not disclose token or replay digest. A bounded adapter-only projection is distinct from a caller-visible result.
8. Derivation grants **no** provider-verification, issuance, durability, reservation, or consumption authority. PostgreSQL later owns unique winner, unknown reconciliation, retained tombstones, and one-use issuance.

### Existing boundaries and consumers

- `risk/policy.go`: `Action`, `Policy.Admit`, canonical generated operation catalog, trusted `Context`; no CAPTCHA flow/proof contract.
- `risk/evaluation.go`: generic configured source evaluation; existing `"captcha"` appearances are generic test/example signal names, not provider verification or core replay implementation.
- `risk/counter.go`: existing copied versioned HMAC-key/redacted-projection precedent. **Its MAC encoding is different**: length-prefixed domain/fields; it cannot be reused unchanged.
- `risk/issuance.go`, `use.go`, `use_bridge.go`: phone-specific immutable issuance/use contracts and retained-key bindings; useful ownership precedent, not CAPTCHA authority.
- `risk/hibp`: downstream module already imports risk to normalize provider facts. Preserve this adapter→core direction.
- `risk/postgres`: downstream module imports risk; comments explicitly state CAPTCHA remains unimplemented.
- No existing `risk.Service`, `FlowContext`, CAPTCHA module, CAPTCHA replay derivation, or CAPTCHA durable contributor at inspected head.

### Safe dependency seam

Keep core replay/provenance/binding contracts in `risk`; future CAPTCHA and PostgreSQL modules consume these public contracts. Core must not import `risk/captcha` or `risk/postgres`. The core-created value should preserve immutable trusted scope and retained-version projections without accepting reconstructed digest strings as proof of core derivation.

### Contract tensions

- Registry defines `captcha.ReplayFingerprint` as an underlying string although authoritative derivation belongs to risk. Importing that type into core would invert the existing dependency direction; accepting its freely constructible string as authority violates the no-substitution requirement. A risk-owned opaque provenance value with downstream projection avoids both defects.
- Registry CAPTCHA constructor/Service describes coordinating durable issuance, while CAPTCHA goal describes a stateless verifier and transaction contract assigns issuance decision to core. This requires reconciliation when implementing the composed Service, **not a blocker for the bounded core prerequisite**.
- Tenant’s reference “scope” versus exact HMAC encoding is not contradictory: tenant belongs to durable namespace, not prescribed MAC bytes.

### Concrete implementation risks

Wrong framing; token truncation/normalization; caller-selected provider scope; operation/flow confusion; digest identity changing with logical request; retained versions omitted after rotation; mutable startup key/config aliases; diagnostic/serialization leakage; and presenting this foundation as completed CAPTCHA issuance or full risk Service.

## Native QA report

- **Principal:** `/root/captcha_replay_qa_f2b`, read-only QA planner.
- **Skill:** `$qa-test-planner`, one bounded pass.
- **Source:** `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`.
- **Inspected base/head:** `f2b893075237fb213704a3222a837247b437bf72`.
- **Authority:** risk/CAPTCHA/PostgreSQL goals; reference configuration; transaction protocol §§420–488; CAPTCHA lifecycle catalog; relevant public-contract rows; supplied platform goal.
- **Disposition:** **No blocking contradiction or unsafe requested behavior found.** No files changed or tests executed.

### Bounded coverage artifact

| Risk / realistic defect | Smallest proving test and observable assertion |
|---|---|
| Wrong algorithm, framing, ordering, encoding, or domain | Independently pinned known-answer vector for ASCII domain + NUL, four uint32-BE length-prefixed UTF-8 fields in provider/site/API-profile/configuration order, uint64-BE raw-token length + unchanged bytes, HMAC-SHA-256. Assert exact 32-byte digest and canonical 43-character unpadded base64url native projection. Expected bytes must not be computed through production helpers. |
| Concatenation collision or unintended token normalization | Unit cases shifting bytes between adjacent scope fields must produce distinct fingerprints; token whitespace/case/byte changes must remain distinct. Include non-ASCII UTF-8 to catch character-count rather than byte-count framing. |
| Command/action binding accidentally defeats replay detection | Same token/key/provider/site/profile/configuration produces the **same** fingerprint across distinct command IDs and valid protected operations/flow bindings. Changing any encoded scope field changes the digest. Tenant changes the outer native replay scope, **not** HMAC bytes; exact provider scope and command-independent fingerprint remain observable. |
| Rotation opens a second replay window | With old key retained and a new active key, the same request still exposes the original old-version lookup identity plus the new identity, with correct version associations. Reordering retained keys must not lose or duplicate identities. Reject zero, duplicate, missing/invalid keys and invalid version/configuration inputs. Native projections must be sufficient for later persistence to find old tombstones; this test does not claim durable denial. |
| Hostile inputs trigger excessive allocation or permissive acceptance | Constructor/derivation table for zero values, missing required scope, unknown provider/profile, malformed UTF-8, configured token limit at limit−1/limit/limit+1, empty token, and supported ceiling 8192/8193 bytes. Assert typed redacted rejection and no valid native proof on failure. Test key minimum immediately below/exactly at/above 32 bytes where supported. Do not trim or silently substitute configuration. |
| Closed flow proof is interchangeable or mutable | One valid pre-auth, authenticated subject/session, and administrator case; zero/mixed/missing discriminators reject. Alter operation, tenant, exact subject/anonymous flow, selected context, challenge, request fingerprint or configuration: original trusted proof must not validate under the changed binding. Copy/mutate caller-owned configuration inputs after construction and returned snapshots after retrieval; established identities and validation remain unchanged. |
| Secret leakage or false authority | Formatting (`%v`, `%+v`, `%#v`, `%s`, `%q`), errors and supported serializers must not reveal sentinel raw token, key or reusable fingerprint. Only explicit restricted native projection may contain required digest/version data; safe results must not. Zero/constructed public metadata must not confer issuance, reservation or use authority. |

**Focused fuzz:** One scope/token admission/framing target, seeded with each boundary above and adjacent-field ambiguity cases. Assert no panic, bounded acceptance, deterministic identities for accepted inputs, unchanged raw-byte semantics, and redacted failure. Cap input before expensive work; retain every discovered failure as a deterministic regression.

**Focused race:** Concurrent reads/derivations from one constructed immutable configuration and its copies; assert deterministic native identities and no corruption. Concurrent mutation of caller-owned startup inputs is not a supported contract unless explicitly documented. No leak/stress suite is necessary for a purely synchronous core with no goroutines or callbacks.

**Verification sequence:** Observe a focused replay-across-command regression red; implement; rerun focused unit/fuzz-seed tests and `go test -race` for `risk`; reconcile affected API/direct-consumer checks. Use task-owned disposable `GOCACHE`. Exact package coverage/mutation and affected repository gates remain delivery requirements, not permission to postpone the first production write.

**Explicitly unproved here:** PostgreSQL unique-winner races, tombstone retention/erasure, terminal transitions, primary reconciliation, provider transport/interoperability and lifecycle delivery. Those require later persistence/adapter tests; core hash equality alone cannot establish one-use issuance.

## Native regression review

Native regression review: **APPROVE** for this bounded local replay-core batch.

- Principal: `/root/captcha_replay_regression_f2b`
- Base/head: `f2b893075237fb213704a3222a837247b437bf72`
- Reviewed scope: five dirty files; independently confirmed digest `fb3ec9cd0c2e3f2517db7ab26f9aa00f4297a6162385785f2d221cae490e1272`.
- Authority: corrected control worktree goals, transaction contract and reference configuration.

No concrete behavioral defect found. Tests meaningfully exercise pinned independent HMAC vectors, UTF-8 byte framing, raw-token preservation, command/action/tenant independence, exact flow substitution rejection, retained-key aliases, custody isolation, redaction, bounds and concurrent derivation. Production preparation retains no raw token and introduces no provider or durable issuance authority.

Runtime evidence inspected in the authoritative control worktree: risk-package race test passed, vet exited zero, focused fuzz passed with 5,572 executions. Its three CAPTCHA source/test hashes match reviewed files. No tests executed by this reviewer.

Nonblocking coverage improvements:
- Exercise subject/anonymous-flow, pre-auth transaction, authenticated session and administrator actor substitution through `Validate`.
- Add retained-key count/order boundaries and canonical base64url trailing-bit rejection.
- Strengthen fuzz rejection assertions to independently distinguish invalid UTF-8 and malformed size, rather than accepting every error.

These are narrow missing partitions, not demonstrated defects or additional checkpoint gates. Full coverage, mutation, durable uniqueness/tombstones, provider integration, clean consumers and canonical Service remain unverified; this approval does not certify platform completion or release readiness.

## Native Go compatibility review

**APPROVE — local coherent replay-preparation batch only.**

- Principal: `/root/captcha_replay_compat_f2b` (`go_compatibility`, read-only).
- Base/head: `f2b893075237fb213704a3222a837247b437bf72`.
- Reviewed complete five-file dirty scope; independently matched digest `fb3ec9cd0c2e3f2517db7ab26f9aa00f4297a6162385785f2d221cae490e1272`.
- Authority: full goal attachment; control risk/CAPTCHA/PostgreSQL goals; transaction contract lines 445–460.
- Concrete compatibility findings: **none**.

Public additions are additive; existing declarations and consumer interfaces remain unchanged. Opaque custody types reject invalid zero values and preserve constructor/preparing-policy provenance. Caller-owned input/key slices and returned projections cannot mutate retained custody through the public API.

Imports remain standard-library-only; module manifests are unchanged. Existing PostgreSQL/HIBP consumers retain their dependency direction toward core risk, with no reverse adapter import or cycle introduced. MAC framing agrees with the authoritative contract; tenant/flow/action remain outside MAC bytes and separately bound.

Both root and risk retain `go 1.27.0`. Added APIs use facilities available at that baseline. Inspected runtime evidence was executed with **Go 1.27.1**, not exactly 1.27.0; it records passing core race, vet and bounded fuzz runs. I executed no tests/builds.

Approval does **not** certify catalog/API-baseline updates, exact-baseline execution, published clean-consumer compatibility, reverse-dependent/release gates, durable one-use behavior, provider verification, canonical Service, or complete platform delivery; those remain pending.

## Native complete-final-diff correctness review

## Native review result: APPROVE

**No findings in the coherent local development batch.**

- **Principal:** `/root/captcha_replay_correctness_f2b`; isolated single-reviewer pass, no sibling reports inspected.
- **Repository/branch:** `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`, `feature/identity-platform`.
- **Base/head:** `f2b893075237fb213704a3222a837247b437bf72`, plus complete dirty scope: `risk/captcha_replay.go`, `risk/captcha_replay_test.go`, `risk/captcha_replay_example_test.go`, `risk/README.md`, `risk/CHANGELOG.md`.
- **Snapshot digest:** independently matched `fb3ec9cd0c2e3f2517db7ab26f9aa00f4297a6162385785f2d221cae490e1272`.

### Closed causal boundaries

- `captcha_replay.go:54–89`: bounded, closed flow construction; canonical operation membership and canonical 32-byte request fingerprint; distinct pre-auth/authenticated/administrator requirements. Structural admission explicitly does not authenticate or authorize.
- `captcha_replay.go:130–145,182–208`: selected immutable policy, closed providers, bounded token/key/configuration inputs, exact prescribed HMAC framing, unchanged token bytes, newest-first retained aliases.
- `captcha_replay.go:213–228`: preparing-policy identity plus complete flow equality protects provenance; independently owned projections cannot mutate custody.
- `captcha_replay.go:232–317`: formatting redaction and rejecting JSON/text serialization match existing custody conventions.
- Dependency direction remains core-to-existing public contracts; no provider or persistence adapter import, new durable writer, or changed existing risk-evaluation consumer. Base exact-symbol search found no prior `CaptchaReplay`/`FlowContext` APIs.

**Falsification attempts:** traced adjacent-field collisions (`ab/c` versus `a/bc`), changed tenant/action/request/challenge, foreign equivalent policy, rotated key aliases, zero values, token-limit successor, and caller mutation of keys/projections. Guards, framing and separate provenance reject substitution or preserve the intended replay identity. None exposed a reachable defect.

### Test quality and evidence

Independent literal HMAC vectors, UTF-8 byte-length framing, scope changes, raw-token distinctions, rotation continuity, configured bounds, flow rejection, ownership mutation, redaction and concurrent derivation exercise observable outcomes without mocked production collaborators. External-package example exercises public composition.

Inspected runtime evidence reports:
- `go test -race . -count=1`: **PASS**, 1.266s.
- `go vet .`: **PASS**.
- `go test . -run '^$' -fuzz '^FuzzCaptchaReplayBoundedRawToken$' -fuzztime=3s -parallel=2`: **PASS**, 5,572 executions.

All 34 recorded risk runtime input hashes match current source. No tests/builds were rerun during this read-only review. Initial missing-API red evidence does not prove every subsequently added partition was individually TDD-tested.

**Approval is limited to this local batch.** Exact coverage/mutation, catalog/API baseline, clean published consumer, release gates, provider interoperability, durable cross-alias single-winner issuance, canonical Service and full 67-unit completion remain unverified; this report certifies none of them.

## Native runtime provenance

Missing API RED observed before production: `go test . -run
^TestCaptchaReplayPinnedEncodingAndCommandIndependence$ -count=1`, session84138,
undefined FlowContextInput/NewCaptchaReplayKey/NewCaptchaReplayPolicy etc.
First GREEN0.192s session8794; expanded focused tests PASS0.216s session30046.
Further partition tests followed initial implementation and are not individually
claimed TDD. Independent OpenSSL ASCII/UTF-8 HMAC vectors are pinned in tests.
Final fingerprinted core race PASS1.266s, vet0, bounded fuzz PASS4.213s.
Durable replay prevention and actual CAPTCHA Service remain unimplemented.

## Native local delivery audit

**APPROVE: bounded local committed replay-preparation claim only.**

- Native auditor principal: `/root/captcha_replay_delivery_2b3`; read-only, no tests/builds, writes, runtime or remote operations.
- Current source head: `2b3b51def584feb62559a2f24d9852881d7b8634`; parent `f2b893075237fb213704a3222a837247b437bf72`; branch `feature/identity-platform`; working tree clean.
- Commit contains exactly the five reviewed files. Their bytes match digest `fb3ec9cd0c2e3f2517db7ab26f9aa00f4297a6162385785f2d221cae490e1272`.
- Inspected retained native mapper, QA, regression, Go compatibility and complete-final-diff correctness reports. Their independent principals and original reviewed revision/scopes are explicit; regression, compatibility and correctness dispositions are **APPROVE**, limited to this batch.
- All recorded source and immutable dependency file hashes match current inputs; retained workspace hash also matches. Focused race, vet and bounded fuzz results therefore remain applicable without rerunning. Their execution revision remains **`f2b8930…` plus dirty files**, not commit `2b3b51d…`.
- No concrete evidence mismatch found.

**Delivery boundaries:** locally implemented and committed; push/publication not established by this local audit, and no current-head remote ref is present locally. CI, deployment and production verification are unverified.

**Remaining gates:** exact coverage/mutation, catalog/API baseline, clean published consumers, affected release/reverse-dependent checks and exact-baseline execution remain unverified. Provider verification, durable cross-alias single-winner/tombstone behavior, canonical Service and complete 67-unit platform delivery are not proven by this batch.
