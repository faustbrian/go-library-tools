# Native administrative lockout workflow checkpoint

Source base/HEAD: `885b9bf6a99cd4aac2d9f8089d25e4b27e73d09f`.
Review scope: entire dirty native contributor/workflow batch, eight tracked
and six new files under `risk/postgres`, including shared coordinator,
authorization owner, core risk policy and existing unselected variants.
Original dirty-scope digest:
`803a5a9e1805dd224098ff69664ee35ed3edd1d93a0c02ad5aac37774deaadcf`.
This is a patch identity, not a complete gate-input fingerprint.

The two completed reports below are transcribed from the prior turn's retained
native role context; they are not new parent-authored independent reviews.
Every role used isolated read-only `gpt-6.1-sol`, medium, `fork_turns="none"`.
The mapper and bounded QA input remain in `risk-lockout-state-review.md`.

## Regression reviewer: /root/lockout_workflow_regression_885

Reviewed base/HEAD: `885b9bf6a99cd4aac2d9f8089d25e4b27e73d09f` and the
complete dirty native batch, including all six new files and causal slice.
Disposition: **APPROVE for a bounded local development checkpoint**.

No concrete observable defect found. Tests exercise start-deny/end-allow,
original intervals, stale CAS, actor-epoch loss, rollback, original snapshots,
denied disclosure, changed intent, concurrent CAS, lost acknowledgement without
duplicate effects, closed-field tampering and typed-nil dependencies.

Remaining nonblocking behavioral coverage: permit expiry during lock acquisition
(`lockout_execution.go:84`); native start, counter-digest rotation and native end
(`lockout_execution.go:143`); selected missing/mismatched immutable transition
history failing closed (`lockout_admission.go:36`). Existing rotation admission
and signing-key rotation tests are separate partitions, not substitutes.

Inspected supplied race PASS16.544s and vet exit0. Reviewer executed no tests,
builds or services. Full coverage, mutation, provider, clean-consumer, release,
login and platform completion are outside this approval.

## Go compatibility: /root/lockout_workflow_compat_885

Initial disposition: **APPROVE for a local development checkpoint only**.
An adoption observation identified the new direct authorization dependency:
the former `a74c24b38463` pin lacks the consumed PostgreSQL authority APIs.
Pre-existing companion identity `b22`/`6b3` clean-consumer gaps were separately
classified; they were not introduced by the native workflow.

After manifest alignment the same isolated principal inspected the updated
complete scope. Final disposition: **APPROVE for a local development checkpoint
only**, with the new authorization-pin observation resolved at the immutable
source-contract boundary.

`risk/postgres/go.mod:8` now pins
`v1.1.1-0.20261005011349-60ed4b3229d8`, matching kernel/core risk.
Immutable commit `60ed4b3229d83dabc40b3e81235ee7a69a95c4b1` supplies
AuthorityScope, AuthorityStore, Get and generic RequireAuthority, Go 1.27 and
pgx 5.10. The changelog records adoption. No new compatibility finding;
diff whitespace passed. Reviewer executed no tests, builds or services.

No remote availability, checksum, clean-published-consumer, full API baseline,
catalog, provider or release verification is claimed. Exact coverage/mutation
and canonical risk Service remain unfinished.

## Complete-final-diff correctness review

## Native review: APPROVE — local coherent batch only

**Principal:** `/root/lockout_native_final_885`, isolated read-only reviewer; single-reviewer structured pass using `$requesting-code-review` and `$testing-coverage`. No sibling reports, writes, tests, builds, commits or remote operations.

**Boundary:** `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`, `feature/identity-platform`. Base and HEAD: `885b9bf6a99cd4aac2d9f8089d25e4b27e73d09f`, plus all eight tracked modifications and six untracked lockout files. Independently recomputed scope digest matches `803a5a9e1805dd224098ff69664ee35ed3edd1d93a0c02ad5aac37774deaadcf`.

**No findings.** No unresolved correctness defect within the stated local-batch contract.

Source-backed closure:
- Current policy admission precedes Work; the original signed preparation time remains fixed. Actor epoch is rechecked under the authorization owner’s retained share lock. Freshness is checked again after root/restriction lock acquisition.
- Coordinator tenant exclusion, retained-key history, logical-root locking and exact-version CAS fence concurrent transitions. Active start conflicts; end preserves the original interval.
- Domain state, immutable version evidence, canonical audit and lifecycle effects use the sole coordinator transaction. Selected admission honors an end only with matching immutable transition evidence.
- Replay/recovery resolves original immutable state rather than latest state. Changed duration or closed subject fields cannot disclose an original success; retained signing-key rotation preserves original-result recovery.
- Nil/unselected profiles retain interval-only behavior and acquire no new migration dependency. Typed-nil authorization is rejected.
- Counter dimensions reject unsupported extras before preparation; raw network/session inputs are not stored in transition history. SQL lookups use bounded keyed identities and indexed unique keys.

**Falsification challenges:** traced stale actor epoch, active restart, concurrent same-version commands, altered recovered intent, later transitions, missing selected history, retained-key rotation and legacy nil-custody restrictions. Dominating guards/transaction ownership prevent the suspected bypasses.

**Test-quality boundary:** integration assertions independently observe persisted versions, original intervals, counter charges, rollback, competing winners, historical snapshots and single durable effects after lost acknowledgement. Audit/outbox implementations are same-Work SQL test seams, not production-adapter interoperability proof.

**Verification provenance:** supplied native results report full-module `go test -race -tags=integration . -count=1` PASS, `go vet -tags=integration .` exit 0, and subsequent focused lockout integration PASS. I did not rerun them.

**Not unit/release readiness:** canonical risk Service/CAPTCHA, coverage/mutation, full production-adapter composition, companion dependency pins, clean published consumers and other package release gates remain pending. No complete gate-input fingerprint exists; this review does not certify expensive evidence reuse, publication or deployment.

### Native addendum — APPROVE retained

Inspected both fresh runtime artifacts. All recorded source, dependency-archive and workspace file hashes match current inputs; no mismatches.

- `go test -race -tags=integration . -count=1`: PASS **13.115s**, exit 0.
- `go vet -tags=integration .`: exit 0, empty output.

Artifacts record execution revision `885b9bf6a99cd4aac2d9f8089d25e4b27e73d09f` plus dirty inputs, Go 1.27.1 environment, pinned PostgreSQL image and isolated PostgreSQL 18.6 profile.

This replaces the earlier statement that input identities were unavailable. It supports this composed native-adapter development verification only—not release readiness, clean-consumer compatibility or expensive-gate reuse. Original review boundaries remain unchanged.

## Local source checkpoint

Committed as `f2b893075237fb213704a3222a837247b437bf72`. The staged
fourteen-file scope matched the independent review digest before commit.
Source worktree clean; no push. Unit counts remain unchanged.

## Native delivery audit

**APPROVE — bounded local committed-batch claim only.**

- **Audit principal:** `/root/lockout_native_delivery_f2b`; read-only.
- **Exact head:** `f2b893075237fb213704a3222a837247b437bf72`, branch `feature/identity-platform`; source worktree clean.
- All **14 committed files** match current bytes and independent final-review digest `803a5a9e1805dd224098ff69664ee35ed3edd1d93a0c02ad5aac37774deaadcf`.
- Native correctness principal `/root/lockout_native_final_885` reviewed the complete dirty batch at base `885b9bf…`; retained verbatim report/addendum approve this local batch. Mapper/QA reports cover the continuing mutation scope. Regression/compatibility transcriptions remain carried-forward provenance, **not fresh native byte-captured reports**.
- Both runtime artifacts match **376 source inputs, 2,280 archived dependency inputs and two workspace files**, with valid recorded fingerprints and no mismatches. Recorded race integration PASS **13.115s**; vet exit **0**. Execution revision remains `885b9bf…` plus dirty inputs; these were not rerun at committed head.
- **Delivery:** local and committed confirmed; no local remote-tracking ref contains head. No publication, CI, deployment or production evidence inspected; remote state was not queried. Do not infer remote absence from local refs.

**Remaining:** canonical risk Service/CAPTCHA, outstanding behavioral partitions, exact coverage/mutation, production-adapter composition, companion dependency compatibility and clean published consumers, applicable release/provider gates, usable login and all-67-unit acceptance. No unit-verification, release-readiness or platform-completion claim is supported.

## Earlier native runtime progression

Retained from prior turn context, not a new fingerprinted run:
- Missing owned workflow API RED: session15722; first GREEN1.248s session62826.
- Dropped subject fact replay RED: session17342; corrected GREEN1.333s30734.
- Omitted caller epoch RED: session45109; owned owner lookup GREEN1.386s9938.
- Typed-nil authorization RED: session64856; corrected final native suite.
- Concurrency/key rotation: PASS1.435s98718; lost acknowledgement/focused
  native suite PASS2.054s86886; full integration/race PASS16.544s/vet0 session4637.
- Final dependency alignment only: focused integration PASS1.573s46505.
Initial SQL contributor tests followed execution code and are not claimed TDD.
Those sessions lack complete-input reuse identity. The fresh fingerprinted
artifacts above are the final native adapter verification for this checkpoint.
