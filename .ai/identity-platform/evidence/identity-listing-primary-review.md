# Identity listing development batch

Package commit: `github.com/faustbrian/go-identity@b3bce20b3682e3608c3bc668624afa7ea5fb5f6f`.
Partial core development; unit remains in-progress, not verified/released.

## Native independent review

Principal `/root/identity_listing_review`, GPT 6.1 Sol medium, isolated
`fork_turns=none`; single read-only complete-final-diff reviewer using
`$requesting-code-review`. No reviewer writes, Go executions or external mutations.
Base/current HEAD before commit: `71a0ab11c1063dbe5a6625bb4cdf20c6355b85d7`.
All nine files: query.go, list.go, list_cursor.go, list_test.go,
list_cursor_test.go, list_example_test.go, README.md, CHANGELOG.md, docs/listing.md.

Initial native disposition: REQUEST_CHANGES — development batch only.

> Blocker: list_cursor.go:66 generates continuations exceeding its own 1024-byte
> admission limit. A valid 128-byte user ID containing only & passes structural
> validation and the permitted injected canonical validator. json.Marshal()
> expands the last ID alone to 768 bytes; encryption/signature/base64 overhead
> makes the token exceed 1105 bytes even before other claim fields. The first
> page succeeds, but continuation fails before storage. Falsification: neither
> canonical validation nor page validation excludes that supported ID; sealing
> has no size guard. Use bounded canonical encoding and add a full
> listing/continuation regression with maximally escaped IDs.
>
> Non-blocking: list.go:184 does not fully reject overflowing clock samples as
> documented. Go's UnixNano() wraps outside its supported range; year 2600
> produces positive 1434155126290448384, passing this guard and producing a
> repository expiry in 2015 while authorization receives year 2600. Explicitly
> validate the representable interval before conversion, or narrow the
> documentation's rejection claim. Source-confirmed against installed Go
> implementation; no reviewer runtime execution.

Final native disposition after complete revised-diff review:

> APPROVE — revised development batch only. No findings.
>
> Complete nine-file review against base/current HEAD
> 71a0ab11c1063dbe5a6625bb4cdf20c6355b85d7, read-only, single independent reviewer.
>
> Both prior findings are resolved: supported maximum IDs produce usable
> continuations, and unrepresentable clock samples fail closed. CR/LF alternate
> cursor encodings are rejected before cryptographic work. Existing get-only
> construction and UserGet() remain compatible.
>
> Supplied current-tree verification: race tests pass (1.488s); vet and diff
> check pass; revised cursor fuzz passes (10.288s, 1358637 executions). Exact root
> coverage is 382/383, not 100%; final post-generation cancellation return
> remains unexecuted.
>
> This approves the local development batch, not module verification, release
> readiness or complete identity functionality. Exact coverage, mutation,
> durable adapter retention, catalog reconciliation and clean-consumer/composed
> acceptance gates remain pending.

## Primary execution

Go go1.26.6, task-owned disposable build/module/temp storage. Explicit temporary
workspace uses authorization candidate a74c24b3846323e0c1627c6fccbb79ce25248392;
there is no tracked replacement or published dependency availability claim.

```text
go test -run '^TestUserListAuthorizesAndContinuesStableSnapshot$' .
FAIL before implementation: missing listing API
go test -run '^TestUserListSnapshotExpiryBoundsInFlightWork$' .
FAIL before bounded expiry context: failed listing returned partial output
PASS after correction
go test -run '^TestUserListCursorRoundTripsMaximumCanonicalIDs$' .
FAIL before encoding correction: valid maximum canonical ID produced unusable cursor
go test -run '^TestUserListCursorRejectsAlternativeBase64Encodings$' .
FAIL before canonical base64 guard: failed listing returned partial output
go test -run '^TestUserListRejectsPositiveWrappedUnixTimestamp$' .
FAIL before representability guard: failed listing returned partial output
go test -run '^TestUserList(CursorRoundTripsMaximumCanonicalIDs|CursorRejectsAlternativeBase64Encodings|RejectsPositiveWrappedUnixTimestamp)$' .
PASS 0.187s after corrections
go test -race -count=1 -timeout=90s -covermode=atomic -coverprofile=<task>/list.cover ./...
PASS 1.488s; exact382/383 root statements, compiled example included
go vet ./...
exit0
go test -run '^$' -fuzz '^FuzzUserListCursorAdmission$' -fuzztime=10s -parallel=2 .
PASS 10.288s; 1358637 executions on revised parser
```

Earlier fuzz 497362 executions and coverage376/377 are historical, not proof of
the revised parser. One final cancellation return at list.go:288–290 remains
unexecuted; it is not waived or rounded away. Tests prove the retained-snapshot
storage seam against an in-memory fixture, not durable PostgreSQL retention.
Root inventory is unchanged/readable. Whitespace, selected-file secret-shaped
scan, message lines and Intent Core commit validation passed. No project ACK
profile applies. Module creation/writes, full PolicySet/attributes, durable
adapters, exact release coverage/mutation, catalogs and composed parity remain
required. No push, publication, release or deployment occurred.
