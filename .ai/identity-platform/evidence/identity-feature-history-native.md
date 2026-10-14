# Feature history native reports and focused outputs

Repository: /Users/brian/Developer/.worktrees/golib-go-identity-identity-platform
Base/head: 42e93db0d80748ab500ceded0d75eab86bea2e25 plus three dirty paths.
Source manifest:
8a7d0af5466ca298b29847b38dd93e5f3f3877709a3c42c7797f01fff30cf5bd postgres/plan_feature_decode.go
16850694d963d2a9b3137456b597e2ae3f7f24b9b0ffb1b423694ea65a048876 postgres/plan_feature_decode_test.go
9a8b0524838ee1344e45979fd0bbe1a8c8c3d03801dedfd0169b679c804ad6cf postgres/CHANGELOG.md

Primary-executed native output, captured from exec session20778 (exit0):
```
ok  	github.com/faustbrian/go-identity/postgres	2.006s	coverage: 52.5% of statements
```
Command: go test ./postgres/... -race -count=2 -coverprofile=<owned>/cover.out
Environment: Go1.27.1; task-owned GOCACHE/GOMODCACHE/GOTMPDIR/GOWORK at
/private/tmp/identity-feature-history.Rk74MJ; exact consumed root7ca793f plus
authorizationa74c24b38463/capability174b128867a4/migrationsa461b18b6b24/
postgres61261695bcc7 archives. Generated coverage available there until cleanup.
RED: undefined decodeFeaturePlan; follow-up RED oversized scope allocated158
before rejection. Focused corrected GREEN0.258s. Full release proof not claimed.
Final vet exit0. Native exec session91178 fuzz output (exit0):
```
fuzz: elapsed: 0s, gathering baseline coverage: 0/5 completed
fuzz: elapsed: 0s, gathering baseline coverage: 5/5 completed, now fuzzing with 2 workers
fuzz: elapsed: 3s, execs: 425 (142/sec), new interesting: 1 (total: 6)
fuzz: elapsed: 6s, execs: 425 (0/sec), new interesting: 1 (total: 6)
fuzz: elapsed: 6s, execs: 425 (0/sec), new interesting: 1 (total: 6)
PASS
ok  	github.com/faustbrian/go-identity/postgres	6.291s
```
Command: go test ./postgres/... -run '^$' -fuzz '^FuzzDecodeFeaturePlan$'
-fuzztime=5s -parallel=2. Exact statement counts from generated coverage:
adapter1757/3345; new decoder95/95. Bounded fuzz is not exhaustive proof.
Coverage-only correction added paired MaxUint64->0 wraparound and duplicate
audit contributor contract-ID mutations. Production/changelog unchanged.
Native final unit exec session30570 exit0:
```
ok  	github.com/faustbrian/go-identity/postgres	2.068s	coverage: 52.6% of statements
```
Same race/count2/cover command and owned environment; preceding fuzz proves
unchanged production input, not execution of the two later deterministic tests.

## Native corrected regression report: /root/feature_history_regression

**APPROVE — bounded implemented-coverage audit.** The previous finding is closed; no unresolved material coverage defect remains.

Reviewed `feature/identity-platform`, base/head `42e93db0d80748ab500ceded0d75eab86bea2e25`, plus the complete three-file dirty slice matching the current manifest.

- **Covered contract:** canonical reconstruction, frozen fields/order/history, binding/fingerprint checks, bounded hostile input, audit cardinality/order, owned copies, record variants, lifecycle limits, and authority denial.
- **Corrected assertions:** paired `MaxUint64 → 0` lifecycle history and duplicate audit contributor contract IDs now require zero plan plus `UnavailableError`. They detect removal of the respective semantic guards. Tests: postgres/plan_feature_decode_test.go:95.
- **Misleading tests:** the earlier masking cases are supplemented by decisive assertions; no remaining material misleading proof identified.
- **Test boundary:** focused decoder/coordinator unit tests suffice. No SQL or browser coverage is needed for this private, unconnected decoder.
- **Evidence inspected:** final `go test ./postgres/... -race -count=2 -coverprofile=…` passed; decoder coverage is **95/95 statements**. Vet and bounded fuzz passed against unchanged production inputs.
- **Remaining evidence gaps:** package coverage is **52.6%**; full release coverage, mutation, API/catalog gates, and complete final-diff correctness review remain pending. This approval does not certify release readiness.

No writes, Go execution, external changes, or new planning artifacts were performed.

The preceding REQUEST_CHANGES report is retained in the native session transcript:
its concrete missing wraparound partition was corrected, independently reread,
and approved above. It was a test-oracle finding, not a production-code defect.

## Native final correctness report: /root/feature_history_final

**APPROVE — no findings.** This approves the local development slice, not release readiness.

Reviewed `feature/identity-platform` at `42e93db0d80748ab500ceded0d75eab86bea2e25` plus the complete three-file dirty slice. Final hashes match the supplied manifest.

- Bounded parsing, exact binding/fingerprint matching, canonical re-encoding, zero-plan failures, and owned copies are conserved.
- Historical descriptors remain independent of today’s registry.
- No production caller invokes the decoder; `ReserveCommand()` still rejects feature plans before SQL (`postgres/coordinator.go:219`).
- Tests include independent pinned bytes, truncation/malformed inputs, audit bounds/order, lifecycle overflow, scope mismatch, ownership and execution denial.

**Evidence inspected:** final race/count-two test output passed; decoder coverage **95/95 statements**; vet and bounded fuzz passed against unchanged production. `git diff --check` passed. No Go commands, writes, runtime access or external mutations performed by this reviewer.

**Remaining gates:** module coverage is **52.6%**; full coverage, mutation, API/catalog and applicable release gates remain pending. Public feature query, recovery and execution remain excluded.

Reviewed content committed unchanged as e12a6b4cb19482ca1d7289c3ccc05d05dafbbf12.
Local clean identity worktree; not pushed. Intent Core message validation passed
with effective impact unspecified. This is a development checkpoint only.

## Native QA final report: /root/feature_history_qa

Artifact: `identity-postgres-feature-history-coverage`, produced by `/root/feature_history_qa`. Its immutable content is the canonical sorted JSON emitted in my preceding tool result.

Repository: `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`
Base/head: `42e93db0d80748ab500ceded0d75eab86bea2e25`

**No implementation-blocking contradiction or unsafe requested behavior identified. This is coverage planning, not implementation approval.**

### Observable contract

Reconstruct exact canonical `identity-feature-plan-v1` declarations from bounded historical bytes and the supplied binding/nonzero fingerprint. Preserve feature identity, contributor order/versions, all audit fields, lifecycle versions and SHA-256 of the original bytes. Invalid input returns a zero plan and redacted `UnavailableError`.

Reconstruction supplies **no current authorization, registry ownership, reservation or execution authority**.

### Smallest sufficient coverage

| ID | Proving boundary | Decisive assertion / realistic defect |
|---|---|---|
| FH-01 | Unit codec contract | Independently authored frozen byte fixture reconstructs explicit expected fields and independent SHA-256. Detects mutually consistent encoder/decoder omissions. Cover required and omitted records. |
| FH-02 | Unit codec contract | Preserve four ordered contributors, historical versions, and all fields of 1/64 audits. Detects sorting, first-audit-only decoding and replacement with current registry values. |
| FH-03 | Unit hostile input | Changing each binding field or supplied fingerprint yields zero plus `UnavailableError`. Detects foreign-command substitution and omitted scope checks. |
| FH-04 | Unit + bounded fuzz | Reject every truncation, trailing data, duplicate/unknown keys, nonminimal/indefinite CBOR, invalid UTF-8, excessive lengths/counts and oversized input without panic or partial output. Successful fuzz decoding conserves canonical bytes and digest. |
| FH-05 | Unit semantic boundaries | Exercise audit counts 0/1/64/65, contributor completeness, numeric narrowing, audit budgets/order and lifecycle overflow. Required-record maximum increment succeeds; overflow fails. Omitted records require empty target and 0/0 versions. Missing lifecycle and any capability-bearing declaration fail. |
| FH-06 | Coordinator unit boundary | Decoded/prepared feature declarations still cannot reserve or enter kernel admission; zero ownership/result and rejection before SQL. Mutating source bytes cannot mutate decoded declarations. |
| FH-07 | Existing regressions | Preserve kernel audit-only/lifecycle/capability historical decoding, feature preparation/digest and command-codec behavior. |

**Execution set:** focused red reproduction, these deterministic assertions, existing affected regressions, then a bounded decoder fuzz run. Apply affected module/static/API/coverage gates at their appropriate development/release boundary.

### Scope reductions and remaining boundaries

- One representative feature owner plus mismatches suffices for this shared grammar; no full future-feature/provider matrix.
- Independently mutate constraints; combine only count/allocation, narrowing and record/lifecycle interactions.
- SQL, runtime services and provider tests cannot add proof for this pure decoder batch.
- Public `QueryCommand` disclosure, feature `RecoverCommand`, native phases, migrations and execution enablement remain explicitly excluded.
- Historical support assumes the currently evidenced feature-plan-v1 grammar; no additional legacy feature profile was found.

Source inspection confirms the existing historical decoder is kernel-only. **No writes, Go commands, tests or fuzz execution performed by this planner.** Parent-reported RED is not represented as independently executed evidence.

Receipt artifact SHA-256: 50b7c6abcec0137e8315c5e59baf45d5daafaeedcfc7bd9005e5090cdc5852cd
