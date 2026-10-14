# Native independent risk counter review

## First report

**REQUEST_CHANGES — local development batch**

Principal: `/root/risk_counter_final`  
Base/HEAD: `2be900da6ba888b37f9dc296e266d039e6a22e2b`  
Reviewed staged tree: `ddc794df7ec13112694df66e9a7e725aec372905`

### Blocker: complete committed results are discarded when accompanied by an error

[risk/counter.go:425](/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform/risk/counter.go:425)

An owned adapter can return an attributable, complete `CounterCommitted` result together with a cancellation or cleanup error after mutation. `Mutate()` immediately returns `CounterUnknown` because `readErr != nil`, before retaining the result’s explicit commit state. This erases a known commit and incorrectly routes subsequent handling toward ambiguity/reconciliation.

The store contract distinguishes errors **without** complete direct results, and cancellation must invalidate usable facts without erasing known mutation state.

**Reproduced:** an isolated copy of the exact staged tree returned `CounterCommitted` plus `context.Canceled`. The focused regression failed with `outcome=3` (`CounterUnknown`), while the signal correctly remained unavailable. Dominating guards accepted a valid, owned, fresh command; the result constructor accepted its complete snapshot. No guard disproved the path.

**Remediation:** validate attribution and preserve complete declared outcome before handling the accompanying error; return unavailable facts when an error exists. Add committed-result-plus-error cancellation/timeout regressions while retaining unknown behavior for absent, invalid, or foreign results.

### Complete reviewed boundary

All five staged files and every diff hunk were inspected, including unchanged policy admission, catalog, assessment validation, evaluator timing, test fixtures, examples, and repository consumer searches. No pre-existing production consumer dispatches this new API.

Exact phase/dimension admission, bounded keyed scope, independent ownership, complete snapshots, original observation expiry, threshold boundaries, protected-phase rejection, foreign-result handling, and no automatic retry otherwise have source-backed closure.

Supplied race/tests, vet, and fuzz results support the bounded batch, but do not cover the reproduced result-plus-error partition. Full-unit coverage/mutation, adapters, journal recovery, protected transactions, canonical service/catalog integration, and release obligations remain pending—not additional blockers to this **local batch** claim.

No project files were changed. Disposable reproduction source and Go cache were removed.

## Final report

**APPROVE — local development batch. No findings.**

Principal: `/root/risk_counter_final`  
Base/HEAD: `2be900da6ba888b37f9dc296e266d039e6a22e2b`  
Reviewed staged tree: `7f4917b15be696c0d145234fdfb3da689f147554`

Complete review covered all five changed files, unchanged policy/catalog/evaluator boundaries, fixtures, examples, and direct-consumer searches—not only the previous finding.

**Previous blocker resolved:** [counter.go:425](/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform/risk/counter.go:425) validates attribution before preserving the explicit outcome; accompanying errors prevent usable facts without erasing committed or proven-not-committed state. Missing/foreign results remain unknown. Independently executed the staged tree’s `TestCounterCompleteResultWithErrorPreservesOutcomeWithoutUsableFacts`: **passed**, including all outcome/error partitions.

Source-backed closure retained for exact phase/dimension admission, bounded domain-separated keys, immutable projections/results, complete post-state, original observation/window freshness, threshold boundaries, cancellation/timeout conservation, redacted errors, protected-phase rejection, and no automatic retry. Consumer searches found only tests/examples/documentation for this new path.

Verification supplied for this tree: `go test -race ./... -count=2 -coverprofile` **passed**, 91.0% statements; `go vet ./...` **passed**. Earlier fuzz evidence applies only to unchanged preparation—not changed normalization.

This approves the **local development batch**, not risk-unit completion, workflow adoption, or release. Exact coverage/mutation, official adapters, rotation/epoch continuity, journal recovery, protected transaction participation, canonical service/catalog integration, and live integration remain pending.

No project edits or external mutations. Disposable test copy and Go cache removed.
