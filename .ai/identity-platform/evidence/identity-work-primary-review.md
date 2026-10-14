# Native restricted Work execution review

Principal: `/root/identity_work_review`.
Invocation: default, GPT-6.1 Sol, medium, fork_turns=none; read-only.
Base: `5e81f864e323cf743a8913f395da93d15d48a581`, plus nine Work batch files.

## Original native disposition

**REQUEST_CHANGES — development batch**

**Blocker:** work.go:249 and work.go:284 do not enforce the Query caller’s cancellation/deadline while delivering buffered rows. Concrete trigger: obtain rows from `Query(callCtx, "SELECT 1")`, cancel `callCtx`, then call `Next()` and `Scan()`. `Next()` accepts the prefetched row, and `Scan()` populates the destination successfully because both check only the longer-lived Work context. This violates the documented earlier-caller-deadline contract and bypasses cancellation poisoning. Check the cursor context before delivering or scanning data, classify/poison before cleanup cancellation, and add cancellation-after-Query and cancellation-between-Next-and-Scan regressions.

**Falsification:** traced every dominating guard and pinned pgx v5.10.0. Query prefetches the row; `Work.usable()` inspects only `state.ctx`; pgx `baseRows.Scan()` decodes buffered values without checking context. Driver network cancellation cannot prevent these already-buffered reads. This is source-confirmed; runtime reproduction was not performed under the read-only/no-execution assignment.

Complete nine-file diff reviewed against `5e81f864e323cf743a8913f395da93d15d48a581`; all supplied hashes match. No additional findings. Scope/lease fencing, restricted SQL and scalar encoding, budgets, redaction, alias invalidation, rollback ownership, tests and documentation were traced.

Supplied integration/race, vet, diff-check and fuzz results pass; they do not cover this cancellation partition. Public Coordinator issuance, typed participant phases, audit/effect staging, exact 100% coverage, mutation and release verification remain pending—not batch-completion claims.

## Primary reproduction and correction

Both new real PostgreSQL integration cases failed before the fix: prefetched
row delivery after caller cancellation; buffered Scan after cancellation
between Next and Scan. After checking the cursor context before/after driver
delivery and decoding, both pass with CancelledError, no cancelled Scan
destination mutation, and poisoned subsequent Work execution.
Focused integration PASS 0.414s; complete integration/race count=2 PASS 2.537s;
integration vet and diff whitespace PASS. Exact statement coverage 771/889,
not 100%. Committed source: `f38c1e798cb99e6c04945487d57ca2e33c669581`.

## Updated native final disposition

**APPROVE — development-only batch. No findings.**

Complete nine-file final diff reviewed against `5e81f864e323cf743a8913f395da93d15d48a581`; all frozen hashes match.

The cancellation blocker is resolved: cursor operations check both lifetimes before/after delivery and scanning, preserving classified failure and Work poisoning. Both new regressions reject buffered-data admission after cancellation.

Full closure covers scope/lease fencing, exact extended-protocol execution, scalar encoding, result budgets, redaction, transaction ownership, alias invalidation, cleanup, tests and documentation.

Parent-supplied final integration/race and vet results pass; independently inspected `git diff --check` passes. No code or services executed by this reviewer.

Approval permits the reviewed local development commit only. Coordinator issuance, typed participant phases, audit/effect staging, exact 100% coverage, mutation and release gates remain pending; this is not complete command authority or release readiness.
