# Identity accounts independent review

Principal: `/root/identity_accounts_review`
Model: GPT-6.1 Sol; medium; isolated fork_turns=none; read-only.

## Native first report

**REQUEST_CHANGES — development batch**

Reviewed single-reviewer, read-only snapshot: `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`, branch `feature/identity-platform`, base/HEAD `d7dc708d8ab0d8ca44cc253d838888c8a651b2d3`, complete seven-file dirty/untracked scope. Final digest remained `a10b81e747c22268dda670ca4a4f7fb3a1a88eefa9b8b57bd0942847b37eef9e`.

**Blocker:** accounts.go:330 accepts `authorization.NotApplicable` as a valid denial. It becomes `ForbiddenError`, so the per-row loop silently omits that account and returns successful empty or partial output. This contradicts the batch contract that invalid/unavailable policy aborts the whole result. Upstream `NewDecisionService` explicitly rejects this outcome; the public constructor also accepts custom `authorization.Service` implementations, including the direct test implementation.

Falsification: admission and current-user checks can return valid `Allow`; only the row decision needs `{Outcome: NotApplicable, Revision: current}`. No dominating guard or automatic wrapper rejects that zero outcome. The existing invalid-outcome test covers `99`, not `NotApplicable`.

Remediation: admit only explicit `Allow` and `Deny`; classify every other outcome as unavailable. Add a regression asserting no successful empty/partial result when a row returns `NotApplicable`, including after a previously allowed row.

The remaining inspected tenant/subject fences, projection copying/redaction, bounded reads and pure final-access partitions produced no additional findings. `git diff --check` passed. No tests or mutating gates were run by this reviewer; supplied race/vet/example results were not independently rerun.

Transactional link/unlink, locked eligibility, fresh proofs, cleanup, pagination, released dependency composition and release certification remain unproved and explicitly outstanding; this report does not certify full identity completion.

## Native final report

**APPROVE — reviewed development batch. No findings.**

Single-reviewer read-only complete-final-diff re-review covered all seven scoped files and their unchanged causal contracts in `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`, branch `feature/identity-platform`, base/HEAD `d7dc708d8ab0d8ca44cc253d838888c8a651b2d3`.

Reviewed snapshot digest: `89c948ac65bde863aa9b4670ce46b04e796385b87d0ef578906a9b404379c55e`; unchanged at final inspection.

The previous finding is resolved: `AccountService.allow()` now accepts only current-revision explicit `Allow`/`Deny`. The new regression reaches `NotApplicable` after an allowed row and requires `UnavailableError` with no partial output. Tenant/current-subject fencing, metadata bounds, projection/redaction, final-access exclusions and Go `1.27.0` directive remain consistent with this batch’s requirements.

Verification: directly observed `git diff --check` pass. Parent reports post-fix Go1.27.1 root race tests, external example and vet passed; coverage remains 1592/1721 statements. I did not run Go commands. Account fuzz is pending.

This approval is for the bounded development batch—not release or full identity completion. Transactional/locked link-unlink, fresh proofs, cleanup, pagination, SQL, replacement-free released dependency composition, full coverage/mutation and remaining release gates are still outstanding.

## Primary subsequent observation

The unchanged reviewed snapshot's account fuzz rerun subsequently passed:
`go test -run '^$' -fuzz '^FuzzAccountValidation$' -fuzztime=10s -parallel=2 .`
on Go 1.27.1, 437922 executions. This is the primary implementer's direct
observation, not an independently executed reviewer result.
