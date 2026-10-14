# Native independent outbox recovery review

Agent: `/root/identity_outbox_recovery_final`, default GPT6.1 Sol medium,
isolated `fork_turns="none"`, read-only. No implementation worker.
Reviewed dirty batch committed unchanged as
`b956e68880ab4fcee4a0e14947db0f38dc2e2b4a`.

## Native report

**APPROVE — DEVELOPMENT batch only. No findings.**

**Reviewed boundary:** `feature/identity-platform` in `/Users/brian/Developer/.worktrees/golib-go-transactional-outbox-identity-platform`, base/HEAD `3e1ab9a8c7c327538b6daee014683c45d25f606c`, plus all five dirty tracked files and untracked `postgres/identity/recovery.go` and `recovery_test.go`. Consumed coordinator source inspected at `93bd9948719c7a39705462af624c6a8772146d5e`. Single-reviewer, read-only pass; no edits, service access, executions of reviewed code, or subagents.

**Source-backed conclusions:**
- `recovery.go:62–68` reads only retained tenant/command bindings. The existing unique `(tenant_id, command_id, event_index)` index supports the ordered two-row prefix. Zero rows → negative; exactly one expected index → historical effect; wrong or additional indices → unknown.
- Persistence claims the immutable binding in the same Work transaction as the payload. Existing migration triggers reject binding update/delete/truncate; there is no payload foreign key. Generic pruning and relay updates affect payload/delivery state, not recovery evidence.
- Frozen scope and lifecycle declaration come from invocation-bound `RecoveryWork`. The public method acquires no private transaction, invokes no hooks/network, retains no authority, and closes its cursor on every exit.
- Coordinator admission validates the frozen plan. Recovery obtains the command claim/journal locks and tenant exclusion before participant reads. `postgres/recovery.go:157–164` refuses terminalization for every status except `RecoveryNotCommitted`; positive evidence cannot manufacture a receipt.
- Existing constructor, descriptor, and persistence APIs remain unchanged; recovery capability is additive and its unpublished dependency pin is explicit.

**Falsification checks:** Reading only `event_index` initially appeared insufficient for payload integrity, but the authoritative contract explicitly limits positive evidence to historical participant effects. `LIMIT 2` cannot hide conflicting cardinality: any second matching row forces unknown. Negative-read races are excluded for supported coordinator writers by claim and tenant locking; append-only retention prevents later deletion from invalidating positive evidence.

**Verification:** Parent-supplied native `go test -race -tags=integration -count=2 ./postgres/identity` passed against disposable PostgreSQL 18.6; `go vet -tags=integration ./postgres/identity` passed; fixture database/role counts were `0|0`. I inspected the disposable workspace mapping and independently obtained a clean `git diff --check`. Tests use independent literal status expectations and real binding/pruning/privilege outcomes.

**Remaining boundaries:** Valid public coordinator/audit/outbox `RecoveryWork` composition is the next batch, not proven here. Exact coverage, mutation, release/clean-consumer gates, later domain/cascade recovery, publication, deployment, and production behavior remain unverified. This approval supports the reviewed local development batch, not release readiness.
