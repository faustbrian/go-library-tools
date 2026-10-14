# Account link native independent review

Reviewer: `/root/identity_link_review`; default GPT-6.1 Sol, medium,
isolated `fork_turns="none"`, read-only. Native reports follow verbatim.

## Initial report

**Internal disposition: REQUEST_CHANGES — bounded development batch.**

**Blocker:** [link_command.go:372](/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform/link_command.go:372): `RecoverAccountLink()` never supplies the requested account ID to current authorization. `admit()` removes caller-supplied `identity.*` attributes and restores only recovery/read-admission/mode; after reading the receipt, recovery checks account equality but never authorizes its account scope. Consequently, a current policy allowing this actor/user while denying `identity.account.id == "account-b"` can permit recovery and disclose that forbidden account’s committed outcome. A policy positively requiring the allowed account instead cannot recover legitimate outcomes. Bind the exact requested account ID into trusted recovery authorization before receipt access, retaining the receipt equality check; test both account-specific denial and successful permitted recovery.

**Falsification:** Traced all dominating actor, tenant, user, mode, operation and receipt guards. None supplies account identity to authorization; `finish()` applies user redaction only, so neither the receipt equality check nor output handling disproves the bypass. Source-confirmed; no Go reproduction executed under this read-only assignment.

**Reviewed boundary:** Single-reviewer structured complete-diff pass, including all eleven dirty files and full untracked contents, unchanged command variants, account/identifier validators, authorization/projection consumers and upstream immutable capability semantics. Base/head `978702f3fda12e98f6f5ba733ee61a597e7dfff9`; branch `feature/identity-platform`; snapshot SHA-256 independently matched:
`b7fa89910a632203a512cae532968a88471faedfbb4bae818d7498ec51b8150e`.

**Verification:** Independently inspected `git diff --check`: passed. Parent-supplied Go 1.27.1 race tests, vet and ten-second fuzz run passed; coverage is **2070/2277 statements**, not 100%. No Go commands or mutations performed by this reviewer.

**Remaining boundaries:** SQL uniqueness/recovery and proof/risk/token-vault atomicity, real provider/session ceremonies, HTTP/Service/catalog/cascade composition, exact coverage/mutation and release certification remain explicitly unfinished—not additional predevelopment blockers. Existing replay tests prove the non-durable seam only.

## Final report

**No findings. Internal disposition: APPROVE for the bounded development batch.**

Complete twelve-file dirty diff and unchanged causal consumers reviewed at base/head `978702f3fda12e98f6f5ba733ee61a597e7dfff9`, branch `feature/identity-platform`. Independently confirmed snapshot:
`d1e66f04ad692e9e0ba755c018217b13c0da47f1fcc6c08e5b36a0a979200ff5`.

The prior recovery finding is resolved: link and unlink authorize the exact typed account before receipt reads, replace spoofed reserved metadata, and retain original receipt-binding checks. Negative/positive recovery tests have independent account-policy and read-count assertions.

`git diff --check` independently passed. Parent reports Go 1.27.1 race tests and vet passed, with recovery regressions demonstrated red then green; unchanged digest fuzz evidence remains applicable. No Go execution or mutations performed here.

This approves a local development commit, **not full-unit verification or release**: coverage remains **2076/2279 statements**; exact coverage/mutation/release gates, durable SQL/proof/risk/token-vault atomicity, real ceremonies and public composition remain pending.

Source commit after review: `eea8bb7736378bbad6727e341f07f1903b8db771`.
