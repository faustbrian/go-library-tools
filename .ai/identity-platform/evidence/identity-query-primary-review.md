# Identity user-query development batch

Package commit: `github.com/faustbrian/go-identity@71a0ab11c1063dbe5a6625bb4cdf20c6355b85d7`.
Partial core development only; no unit/release completion claim.

## Native independent review

Principal `/root/identity_query_review`, GPT 6.1 Sol medium, `fork_turns=none`,
read-only complete-final-diff review using `$requesting-code-review`.
Base/HEAD before commit: `4a609aaf2a91d1256f74419aac02d34a59cae66a`, plus
`query.go`, `query_test.go`, `go.mod`, `modules.json`, `README.md`,
`CHANGELOG.md`, `docs/queries.md`.

Initial disposition: REQUEST_CHANGES. Native finding:

> Blocker: query.go:96–101 omits the registered reason required by
> REFERENCE_CONFIGURATION.md:126 and goals/identity.md. A redactor returning
> the four visible fields and matching policy identity/version discloses the
> user without any reason (query.go:211–214). Add a typed reason, bounded
> registered-reason configuration, and rejection tests for missing/unregistered
> reasons. Falsification: the generated contract omits this member, but the
> normative policy contract explicitly requires it; no dominating guard
> validates a reason.

Final native disposition after complete revised-diff review:

> APPROVE — No findings.
>
> Complete revised review covers all seven dirty files over base/HEAD
> 4a609aaf2a91d1256f74419aac02d34a59cae66a, including upstream authorization at
> a74c24b3846323e0c1627c6fccbb79ce25248392. The registered-reason finding is
> resolved; missing and unregistered reasons deny disclosure, and configuration
> owns its copied registration slice.
>
> Fresh supplied verification: race tests passed, exact coverage 193/193
> statements, vet passed. Independently checked git diff --check passed. Prior
> fuzz evidence applies to unchanged attribute admission.
>
> Approval is limited to this development batch, not module release, published
> clean-consumer availability, or full identity-platform completion. No writes,
> Go executions, or external mutations performed.

## Primary execution

Go `go1.26.6`; all build/module/temp storage task-owned and disposable.
Explicit disposable workspace composes the committed authorization candidate;
no tracked replacement or published availability claim.

```text
go test -run '^TestUserQueryAuthorizesExactTenantAndTargetBeforeRead$' .
FAIL before implementation: missing query API (previous continuation)
go test -run '^TestUserQueryRejectsMissingRedactionReason$' .
FAIL before correction: redaction without a registered reason disclosed user data
go test -run '^TestUserQuery' .
PASS 0.196s after correction
go test -race -count=1 -timeout=90s -covermode=atomic -coverprofile=<task>/query.cover ./...
PASS 1.308s; exact 193/193 root statements
go vet ./...
exit0
go test -run '^$' -fuzz '^FuzzUserQueryAttributeAdmission$' -fuzztime=10s -parallel=2 .
PASS 10.161s; 1251860 executions before reason correction
```

Fuzz evidence exercises the unchanged request attribute-admission path, not the
new reason-registration boundary. Deterministic tests cover that boundary.
Module dependency metadata, whitespace, selected-file secret-shaped scan and
Intent Core commit validation passed. No ACK project profile applies here.
Full-unit mutation, durable persistence, interoperability, replacement-free
published consumers, catalog reconciliation and release gates remain pending.

Reviewed file SHA-256 (boundary traceability, not a complete release reuse key):

```text
4888220323fceef5e095d9bf89d2be67d040db5205b1cfe59d2ef5392604db1f query.go
588fbf7823b7388f5cf5c3263f37ee3dbedd676da368aced45a93178d90bdb5e query_test.go
a0632a7aa89a912a3e9379e116e9e31cf3073207541ac6aaac4f9406fe93587c go.mod
6993a99f5a0b0c5f8e82caddc2e61b3d51a404bdae9f182e818d2b6fbacd367d modules.json
e567175cb4936db4b4ddf867f2c7a481fce4e7d77baae48ec4d649b0532a906d README.md
58b322424f31703d2bbc892386767bcece67b0356f72a52a4f2d248719c81794 CHANGELOG.md
a1e024d6c91547d0e353b0a0ab08ae0e55c72f72f14519050735fd97ad773d3d docs/queries.md
```
