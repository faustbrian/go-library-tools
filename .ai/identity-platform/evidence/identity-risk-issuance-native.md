# Native identity/risk issuance review

## /root/risk_issuance_final — first complete pass

**REQUEST_CHANGES**

Reviewed base/HEAD `54029e84c1920568de20079159d8285b9c4fb38b` and complete staged tree `f8d507ddc115e43af10b85d46a12deee7674267d`. Single-reviewer read-only pass; no delegation or project edits.

1. **Blocker — bound reconstructed inputs before cryptographic work.** [issuance.go:279](/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform/risk/issuance.go:279) calls `preparedEvidenceSignature()` before validating any `EvidenceIntentBinding` strings. A reconstructed projection with valid timestamps, policy/key versions and nonzero proof reaches hashing with arbitrarily large command, tenant, subject, actor, purpose, phase or operation strings. Invalid signatures eventually fail, but only after unbounded synchronous work, violating the selected goal’s explicit pre-hashing input limits. Validate fixed command syntax, bounded identifiers and exact phase/operation/purpose combinations before MAC computation. Add oversized-binding regression coverage.
   - Falsification: checked every dominating guard; none bounds these strings. Disposable witness also confirmed a correctly signed structurally invalid projection is accepted. This is **not** a signature-forgery claim.

2. **Blocker — reject typed-nil native executors at construction.** [issuance.go:401](/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform/risk/issuance.go:401) checks only `c.Executor == nil`. A typed-nil implementation passes construction and panics when `Evaluate()` invokes `Resolve()`, instead of rejecting incomplete startup composition with `ErrInvalidPolicy`. Apply the typed-nil rejection already used by `NewCounterCoordinator()` and cover it with a constructor regression.
   - Falsification: disposable witness supplied `(*evidenceExecutor)(nil)`; construction succeeded and evaluation panicked.

Remaining inspected paths preserve phase-derived binding, owned full evaluation, earliest original expiry, microsecond interval contraction, primary-first recovery, redacted noncommit outcomes, retained-key verification and exact historical replay. No additional findings.

**Verification:** supplied race tests, vet and staged whitespace checks passed; supplied parser fuzz passed with 3,050,403 executions. Independent disposable witnesses passed. Temporary repository and cache were removed.

**Unverified boundary:** real PostgreSQL durability/concurrency, complete canonical Service, consumption, provider interoperability, exact coverage/mutation and release gates remain pending; this review does not certify them.

## /root/risk_issuance_final — complete final re-review

**APPROVE — No findings.**

Reviewed the complete five-file final diff and affected unchanged causal slice against base `54029e84c1920568de20079159d8285b9c4fb38b`; exact staged tree `2b849dac4b791ed60ffd60eb3cd6981a42797cde`.

Both previous blockers are closed: reconstructed bindings are structurally bounded before MAC verification, and typed-nil executors fail startup admission. Key-only command restoration preserves the original signed interval and artifact without an issuer-construction cycle or durable authority.

Supplied exact-tree verification passed: race tests twice, vet, whitespace checks, and parser fuzz with 2,908,897 executions. Coverage is 87.2%, not the required final 100%.

Approval covers this **local risk-core development batch only**. Real PostgreSQL durability/concurrency, full canonical Service, reservation/consumption, provider interoperability, exact coverage/mutation and remaining release gates are unverified. No commit, publication, deployment or production claim is established. No reviewer-owned resources remain.
