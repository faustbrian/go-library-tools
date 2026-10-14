# Native PostgreSQL session structural-read review

Principal /root/session_postgres_read_final; default GPT6.1 Sol medium,
fork_turns none. Parent captured native final reports verbatim below.

## Initial complete-diff report

REQUEST_CHANGES — local development batch only.

**Blocker:** `session/postgres/record.go:76` checks `Lookup` object keys but not the exact digest-array representation. Decoding into `[32]byte` zero-fills short arrays, discards extra elements and accepts `null` scalar elements as zero. Consequently, malformed stored records can reconstruct successfully and match the indexed digest. Require exactly 32 non-null integer bytes, with regression cases for short, long and null-containing arrays.

Complete pass covered every new adapter file and both manifest changes. No other concrete production finding.

Untested risk: the duplicate-digest integration case repeats the same session primary key, so primary-key rejection masks whether digest uniqueness is enforced; the SQL uniqueness constraint is correct.

Supplied race/integration and vet results support this structural-read batch. Coverage, mutation, fuzz execution, API/catalog, full-kernel composition and release obligations remain pending; no authentication or unit-completion approval.

## Refreshed complete final-diff report

APPROVE — local structural-read development batch only. No findings.

Complete boundary: base/head `f48bd027d6a6b7afe39874be69503ab6af0e1741` plus all ten new `session/postgres` files and both manifest changes. Reviewed production hashes match the final verified tree.

The malformed-digest finding is resolved: exactly 32 non-null bytes are required, with short/long/null regressions. The uniqueness test now proves rejection by the digest constraint independently of the session primary key.

Final supplied verification:
- Race/integration, twice: PASS; 137/155 statements.
- Vet: PASS.
- Bounded fuzz: PASS; 1,462 executions.
- Diff check and disposable-resource cleanup: PASS.

Full coverage, mutation, API/catalog, clean-consumer/kernel composition and release gates remain pending. Issuance, current-authority enforcement and lifecycle writes remain unimplemented; this approves neither authentication safety nor unit/release completion.

## Parent execution checkpoint

Reviewed tree committed85148fc7aa32aa7da82a80645ae2002ce78db156; clean/not pushed.
`go test -race -tags=integration ./session/postgres/... -count=2 -coverprofile=<owned>/cover.out`
PASS1.469s137/155 statements; `go vet -tags=integration ./session/postgres/...` PASS.
`go test ./session/postgres/... -run '^$' -fuzz '^FuzzStoredRecord$' -fuzztime=3s -parallel=2`
PASS4.274s1462exec. Tests used PostgreSQL18.6, image
sha256:5a5a84b19854a9ffaa54082c166ff4ec27473a361e496e5ea167f298f2da9722.
Fixture databases0; task container/network removed, pre-existing image unchanged.

Production SHA256:
- migrations.go:89f025ae00ce77dfb68006450d6d16fa4641ae149daa0fd404b8714d276075af
- record.go:cd3137faf0dd8dc9e079b518f1c1992757bcbeb48c7c5708ea41ba2e32ab0d07
- repository.go:a3371f5ff7bf0562136766b82b70b5de37d5042a90c90c2503b1021496cd8dd8

Parent fixture supplied only the minimal identity.users relationship. This
does not prove full kernel/reference composition or authorize credential use.
