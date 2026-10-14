# Native independent counted-evaluation review

## First report

**REQUEST_CHANGES — one low-impact concrete finding requiring closure under repository policy.**

Principal: `/root/risk_counted_final`; independent single-reviewer structured pass using `$requesting-code-review` and test-oracle guidance.

Boundary: repository `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`, branch `feature/identity-platform`; base/HEAD `62dbbfe1e9fca0e595483bd75880219b4a18acf5`, staged tree `7ccc0137e242a7195c2119efa175e3ea69a13c99`. All four staged paths reviewed; no project changes or external mutations.

**Non-blocking: counted throttle metadata overstates the final remaining window.** At `risk/evaluation.go:330`, final assessment consumes the unchanged velocity signal produced before authentication providers. `CounterCoordinator.Mutate()` computes `RetryAfter` against its earlier clock sample, while `Policy.Assess()` copies that duration without accounting for elapsed provider work. The reference operation matrix requires `retry_after=window_remaining`. Thus an exceeded authentication counter with 4m59s remaining at mutation and 500ms of subsequent provider work returns 4m59s instead of 4m58.5s, delaying retries beyond the counter window. Recalculate counter-derived retry metadata at the final decision sample using the original exceeded counters’ window expiries, without renewing observation/expiry or changing commit state; add a focused elapsed-provider regression.

Falsification: traced admission, exact authentication phase, valid positive counter snapshot, provider timeout and final freshness guards; a 500ms provider advancement remains inside all bounds. A disposable copied-module witness failed exactly: `returned=4m59s final window remaining=4m58.5s`. It used a task-owned cache; copied module and cache were removed.

Other causal boundaries closed without findings: frozen sole ownership, exact subject/profile admission, protected-phase rejection, source-only compatibility, required-signal conservation, post-state thresholds, unavailable/unknown/error classification, no automatic retry, cancellation/expiry commit retention, final freshness, bounded calls, concurrency separation and redaction. Existing consumers are tests/examples; no complete service consumer was introduced.

Supplied unchanged-tree checks: focused/full module tests, `go test -race -count=2 -coverprofile=… ./...`, and `go vet ./...` passed. Coverage is 491/543 statements, not exact 100%. Coverage/mutation, durable issuance/journal/recovery, shared-Work phases and real-adapter interoperability remain named release boundaries—not additional blockers for this local development checkpoint.

## Final report

**APPROVE — No findings.**

Principal: `/root/risk_counted_final`; independent single-reviewer structured complete-diff pass using `$requesting-code-review`.

Boundary: `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`, branch `feature/identity-platform`; base/HEAD `62dbbfe1e9fca0e595483bd75880219b4a18acf5`, staged tree `51ef501c353da5e634f1f2d046bb1ad904f98898`. All five staged paths reviewed: `risk/evaluation.go`, `risk/counter.go`, `risk/counted_evaluation_test.go`, `risk/README.md`, `risk/CHANGELOG.md`. No project edits or external mutations.

The prior retry-delay finding is **fixed and withdrawn**. Final assessment now recomputes delay from original exceeded counters’ window expiries, excludes clean counters, preserves configured clipping and leaves the privately retained mutation outcome unchanged. The shared helper preserves existing coordinator behavior at mutation time. The regression’s independent original-expiry oracle rejects elapsed-time overstatement, clean-counter contamination and incorrect clipping.

Complete causal review also closes admission and trusted subject scope, sole frozen ownership, canonical counter phases, source-only compatibility and consumers, post-state threshold decisions, mandatory-signal conservation, error/unknown/noncommit handling, cancellation and partial-success retention, bounded freshness, concurrency, resource ownership and redaction without findings.

Supplied current-tree verification: focused `TestCounted|TestCounter` tests **PASS** after expected focused **RED**; `go test -race -count=2 -coverprofile=… ./...` **PASS**; `go vet ./...` **PASS**; `git diff --cached --check` **PASS**. Coverage: **497/549 statements (90.5%)**, not exact 100%.

Approval is limited to this **local reviewed development checkpoint**. Exact coverage/mutation, durable issuance/journal/recovery, protected shared-Work participation and real-adapter interoperability remain pending release boundaries; no full-unit completion, release readiness, publication, deployment or production claim.
