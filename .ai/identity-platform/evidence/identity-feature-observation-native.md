# Historical feature command observation development checkpoint

Source baseline: identity repository85148fc7aa32aa7da82a80645ae2002ce78db156.
Reviewed/committed source:984b8abccaad061f98406d3cccef88456c0780f0.
Development only; no unit completion, release, push, or native feature issuance.

## Parent execution evidence

Go1.27.1 darwin/arm64; disposable PostgreSQL18.6 image
sha256:5a5a84b19854a9ffaa54082c166ff4ec27473a361e496e5ea167f298f2da9722.
Loopback task-owned cluster, tmpfs data, no existing service/database access.
Workspace consumed exact committed authorization a74c24b3846323e0c1627c6fccbb79ce25248392,
capability174b128867a4, migrations a461b18b6b246f3b6f868259bfd1583c438b0a8a,
database61261695bcc7db86becda200859edb54d00fd498 and audit bc0c96e196f1bd6159839f58038778be247d7f10.
No permanent module replaces; all local archives/workspace/caches were task-owned.

- Missing public descriptor and scope parser produced the expected compiler RED.
- Valid1->2 receipt substitution against frozen0->1 produced the behavioral RED:
  `substituted historical feature was disclosed`.
- Final `go test -race -tags=integration ./postgres/... -count=2 -coverprofile=<owned>/coverage.out`:
  PASS7.977s, exact2930/3375 statements (86.8%). Full100% coverage remains pending.
- Final `go vet -tags=integration ./postgres/...`: PASS.
- Final `go test ./postgres/... -run '^$' -fuzz '^FuzzObservedCommandScope$' -fuzztime=3s -parallel=2`:
  PASS4.327s29719exec.
- Committed audit/postgres direct consumer `go test -race ./... -count=1`: PASS1.250s.
- `git diff --check`: PASS. Intent Core/72-column commit validation: PASS.
- Fixture databases0 before teardown. Task Docker container/network removed;
  existing image preserved. All disposable archives/caches removed at closure.

Parent tests used private journal seeding only to represent old/new-binary
historical receipts; they do not prove a real session issuance transaction.
No provider/runtime milestone, mutation, API/catalog, clean-consumer release,
supply-chain or complete platform gate is claimed passed.

## Initial native independent report

Reviewer: /root/feature_observation_final, default GPT6.1 Sol medium,
fork_turns=none, read-only. Base/HEAD85148fc7aa32aa7da82a80645ae2002ce78db156.
Disposition REQUEST_CHANGES; reviewed all eight changed files.

> P2 — Match committed feature receipt versions to the frozen lifecycle.
> Anchor: postgres/query.go:241–249.
> A canonical committed receipt can describe a different record transition
> from its frozen plan and still be disclosed as committed.
> Keep the exact binding, descriptor, fingerprint, and lifecycle0→1, but
> replace the receipt versions with BeforeVersion1, AfterVersion2.
> validFeatureResult and decodeJournalResult accept the valid successor;
> commandReceipt never compares plan.Lifecycle with result.Feature.
> Both observation and recovery therefore accept inconsistent history.
> Required correction: match both committed versions to frozen lifecycle;
> add a valid-successor substitution regression.

Initial production hashes:
query.go e73b512b5ef0e30ad3524654fb437c9a68015c0708249594ee2389f8087da835
recovery.go b8f64138e8c2884dacdb96c4332fcb47ac671253d8c9a1c43ea582a8f6cb1ee5
scope.go 31533f7a243d73c94a3fd2e2e58ff9d08e68e52db9d0df6f7252b7df5c4fadde

## Native independent corrected-diff report

Disposition: APPROVE — No findings for this local developmental batch.

- Reviewer: /root/feature_observation_final, independent read-only reviewer.
- Repository: identity implementation worktree, branch feature/identity-platform.
- Exact inspected base/HEAD:85148fc7aa32aa7da82a80645ae2002ce78db156.
- Reviewed all eight changed files, including both untracked feature-observation
  tests, and affected authorization, canonical codecs, receipts, frozen plans,
  and recovery callers.

The previous lifecycle-substitution finding is resolved: committed feature
receipts now match both versions against the immutable plan before disclosure.
Receipt decoding guarantees a nonnil matching feature receipt before that
comparison. Unit and integration regressions now exercise the structurally valid
1→2 substitution against frozen0→1.

Current authorization, exact immutable scope, descriptor ownership, bounded
canonical decoding, redacted failures, kernel-only legacy decoding, and
unsupported-feature recovery fences remain intact. Expired feature recovery
cannot reach generation takeover or fabricate terminal results.

Verification boundary: git diff --check passed. No Go commands or tests executed
by this reviewer; execution evidence remains the parent's responsibility.
Approval does not certify release gates, native issuance, publication, or
67-unit completion.

Current production-file SHA256:

query.go 782efa0243b2c7a646053344de1d01cf76f33f07ccc712478a5942a6aea785ee
recovery.go b8f64138e8c2884dacdb96c4332fcb47ac671253d8c9a1c43ea582a8f6cb1ee5
scope.go 31533f7a243d73c94a3fd2e2e58ff9d08e68e52db9d0df6f7252b7df5c4fadde
