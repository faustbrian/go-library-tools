# Identity PostgreSQL listing: native independent review

Reviewer: `/root/identity_postgres_listing_final`, default GPT 6.1 Sol medium,
isolated `fork_turns="none"`, read-only. Local development batch, not release.

## First native report

**Internal disposition: APPROVE for this local development batch, with one non-blocking test finding.**

**Boundary:** `feature/identity-platform`, base/committed HEAD `93bd9948719c7a39705462af624c6a8772146d5e`, plus all seven tracked-file patches and four untracked listing files. Single independent read-only review; no tests, runtime access, writes or remote mutations.

**Non-blocking:** `postgres/list_integration_test.go:97` does not exercise oversized materialization. `list-main` already retains one snapshot, and the test configures `MaxSnapshots: 1`; `List():107` therefore rejects quota before evaluating `MaxUsers`. The test would pass with broken overflow protection. Use a fresh tenant containing two matching users with `MaxUsers: 1`, then assert neither snapshot metadata nor rows remain. Falsification: traced preceding fixture/state and every dominating guard; the existing snapshot is unexpired and no prior test removes it.

**Source-backed closure:** Transactional bounded materialization preserves structural ID/status/full-width version data; continuations bind tenant, canonical statuses and exact expiry. Admission serializes quota checks; continuation share-locks protect against pruning. Schema/index limits bound selection, retention and cleanup. Migration 7 is additive and leaves migrations 1–6 unchanged. Public service composition preserves collection/row authorization, redaction and denied-row advancement. No production defect confirmed.

**Verification:** Supplied adapter integration/race, core race, adapter vet and whitespace checks pass; not independently rerun. The 20,000-row fixture proves the selected indexed query plan, not measured execution throughput.

**Still pending:** identifier-filter lifecycle, attribute projections, remaining commands, privacy cascades, exact coverage/mutation and applicable release/composition gates. This approval does **not** certify package completion, release readiness, hosted CI or deployment.

## Revised native final report

**Internal disposition: APPROVE for the revised local development batch. No findings.**

**Reviewed boundary:** `feature/identity-platform`, exact base/committed HEAD `93bd9948719c7a39705462af624c6a8772146d5e`, plus the complete current dirty patch: seven tracked files and four untracked listing files. Read-only single-reviewer pass; no writes, execution, runtime access or remote mutations.

**Previous finding resolved:** `postgres/list_integration_test.go:98` now uses a fresh tenant with two matching users. `MaxUsers: 1` reaches overflow materialization without quota interference; independent database assertions verify zero retained headers and rows. Retrying with `MaxUsers: 2` and unchanged snapshot quota returns both ordered rows terminally, rejecting an off-by-one limit or leaked quota. Rechecking preceding fixtures and guards did not reveal another masking path.

**Complete-diff closure:** Bounded transactional materialization, stable tenant/filter/expiry continuations, quota serialization, pruning locks/cascades, redacted cancellation/errors, additive migration 7, public authorization/redaction composition and documentation remain consistent with the development contract. No production-code changes followed the earlier review.

**Supplied verification:** adapter `go test -race -tags=integration -count=2 .` PASS (5.181s); adapter `go vet -tags=integration .` and `git diff --check` PASS. Core race evidence remains applicable to unchanged inputs. Disposable fixture databases reported zero. These results were supplied, not independently rerun.

Identifier-filter lifecycle, attributes, remaining commands, privacy cascades and package coverage/mutation/release gates remain pending. Approval is limited to this development batch—not package completion, release readiness or deployment.
