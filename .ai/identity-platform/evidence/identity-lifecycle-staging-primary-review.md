# Native independent lifecycle-staging review

Principal: `/root/identity_lifecycle_staging_review`; default GPT6.1 Sol medium,
fork_turns=none, read-only. Primary retained implementation ownership.

## First complete-diff disposition

**REQUEST_CHANGES — local-development review**

Reviewed HEAD `7eced9f1cd1a84531ecf8391d9f8a2e6bb2bc685` plus the complete
six-file frozen dirty patch, using one independent read-only structured pass.

1. **Blocker: unregistered account lifecycle names.** `work_lifecycle.go:159`
   admits `account.linked.v1` and `account.unlinked.v1`, which do not exist in the
   authoritative exact catalog. `LIFECYCLE_CASCADES.md:80` declares
   `social_provider.linked.v1`/`social_provider.unlinked.v1`; row 116 assigns
   account link/token/unlink authority to this adapter. Valid account bindings
   therefore stage catalog-invalid notifications, and the new tests assert
   those invented names. Use the authoritative applicable names, or reject the
   unsupported profile. **Falsification:** traced binding validation, audit
   aliases and every catalog occurrence; no dominating guard rejects these
   account commands and no newer authority declares `account.*`.

2. **Blocker: legitimate account relinking cannot declare lifecycle staging.**
   `work_lifecycle.go:68` treats every `identity.account.link-token` as
   zero-version creation. The consumed core contract explicitly retains account
   tombstone versions after unlink (`account_link.go:22–25`) and permits fresh
   relinking; its existing regression exercises account version `4 → 5`
   (`account_link_test.go:70–77`). Such a valid declaration is rejected before
   admission by `creating != (BeforeVersion == 0)`. Permit nonzero checked-
   successor account-link transitions and cover both first-link and relink
   staging. **Falsification:** inspected core command/receipt validation and
   the fresh-relink test; neither requires zero, and no separate relink operation
   bypasses this guard.

Other inspected boundaries: plan/digest ownership, exact outbox-phase admission,
scalar redaction, audit target/version agreement, deterministic identity,
duplicate poisoning and Work lifetime. No additional source-confirmed defects.

Parent-supplied race tests, vet and whitespace checks passed; not rerun here.
Durable persistence, Coordinator execution, cascade closure, exact coverage,
mutation and end-state gates remain unfinished—not package verification or
release readiness.

## Updated complete final-diff disposition

**APPROVE — local-development batch. No findings.**

Reviewed the updated complete six-file patch against HEAD/base
`7eced9f1cd1a84531ecf8391d9f8a2e6bb2bc685`. Both prior findings are resolved:
account notifications use catalog names, and first-link/relink versions preserve
the core contract with checked successors.

Parent-supplied `go test -race -count=1 -timeout=120s ./...`, `go vet ./...` and
`git diff --check` passed; not rerun by this read-only reviewer.

Approval covers staging, admission, snapshots and tests only. Persistence,
Coordinator invocation, cascade closure, exact coverage, mutation and final
release gates remain pending; this is not package verification or release
readiness.

Final unchanged reviewed source committed at
`677584cab9c08e7aa15b8e0fcc3abd06035ebb1a`.
