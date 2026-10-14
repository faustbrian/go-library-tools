# Independent identity admin update development review

Principal: `/root/identity_update_review`
Model: `gpt-6.1-sol`; reasoning: `medium`; isolated context: `fork_turns="none"`.
Permission boundary: read-only. No package implementation delegated.
Base/HEAD: `f23f016a47963e81e8af9922aa5462cc866cfc3d` plus eight-file dirty batch.

## Initial native report

**REQUEST_CHANGES — one authorization race.**

Reviewed independently in one read-only complete pass: base/HEAD `f23f016a47963e81e8af9922aa5462cc866cfc3d` plus the specified eight-file dirty snapshot. No writes or external mutations.

**Blocker:** `update.go:208` binds only the caller's `ExpectedVersion`, not the version used for authorization.

- **Trigger:** A new command requests expected version 3 while the authoritative read returns active/version 2. Policy allows application only to active users. Between authorization and `Execute()`, another command bans the user and advances it to version 3.
- **First divergence:** Lines 201–203 authorize active/version 2, but line 208 constructs a command fenced against caller version 3.
- **Impact:** The documented durable-owner fence accepts banned/version 3 and applies a write that current-state policy would deny. The existing owner demonstrates that path at `update_test.go:52`.
- **Falsification/counterevidence:** Terminal-state guards reject deleted/anonymized users, but banned/suspended updates are intentionally supported. The existing race test covers deletion with the original expected version; it does not exclude this future-version witness. Neither request admission nor command construction rejects expected versions greater than the observed version.
- **Remediation:** Reject future expected versions before execution, or carry and fence the authorization-observed version for new mutations after replay lookup. Preserve successful same-command replay. Add the active→banned/future-version regression with a status-dependent authorization policy.

No other findings. `git diff --check` passed. Supplied race/vet/fuzz results were not rerun. PostgreSQL, keyed transport, full composition, exact coverage and mutation certification remain pending—not approved by this development-batch review.

## Primary regression evidence

`TestUserUpdateAdminFutureVersionCannotAuthorizeAgainstOlderState` reproduced
the finding before the fix: Committed, version 4, one audit, nil error. The
service now rejects future expected versions before any authoritative attempt;
older expected versions still reach replay-before-CAS. The complete race suite
and vet passed after the fix.

## Final native report

**APPROVE — No findings.**

Independent, read-only complete-final-diff review of all eight files against `f23f016a47963e81e8af9922aa5462cc866cfc3d`, including shared creation/recovery, field-policy and public-consumer paths.

The authorization race is resolved: future expected versions cannot reach `Execute()`, while older-version replay remains admitted. The new regression exercises the reported witness and asserts no mutation or authoritative outcome claim.

`git diff --check` passed. Supplied red/green, race-suite and vet results were not independently rerun.

Approval covers this development batch only; PostgreSQL, keyed admission, full composition/catalogs and final release gates remain pending.

Reviewed eight-file checksum-manifest SHA-256: `70f688e09d5e694954c754937866c3612a4ae5c6129494d17fb334b1b14d9a6f`.

## Execution reconciliation

Committed unchanged at `efd3d8615bc0ffc9e3a1527b7428c35bea7fad15`.
`go test -race -count=1 -timeout=90s -covermode=atomic` passed the complete
root suite; exact statement coverage 979/1026, not release-complete.
`go vet ./...` and `git diff --check` passed. Prior admission fuzz evidence
remains applicable to the unchanged constructor. No SQL or release claim.
