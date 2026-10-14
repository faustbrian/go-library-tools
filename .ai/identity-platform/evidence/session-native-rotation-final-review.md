# Native session rotation final review

Reviewer: `/root/native_rotation_final_704a1a4`, default `gpt-6.1-sol`,
medium reasoning, isolated read-only context, no child agents.
Local source checkpoint: `171573076ce928e96cadfb423be85bc06e00191a`.

## Native final report (verbatim)

**APPROVE — No findings** for this ordinary-rotation development batch.

The previous blocker is resolved: retained source and successor records are compared with the original request before returning metadata. The substitution regression preserves Committed, returns an error and exposes no Session.

Reviewed base/head: `704a1a40bc878be9eebda07abff159a3e0498596`; sixteen-file fingerprint: `5e8dc947dccca3bcd64b53f08823c00e14095f7fc2ec86251f3be1e990cacc4b`.

`git diff --check` passed. Supplied current race, PostgreSQL integration and vet results passed; not independently rerun.

This is not whole-unit release approval: mutation, runtime-role privileges, consumer/reference, family-replay revocation, remaining profiles and complete release gates remain pending.

## Original finding (verbatim)

**REQUEST_CHANGES** — one blocker.

[rotation_execution.go:184](/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform/session/postgres/rotation_execution.go:184): Recovery does not bind retained records to the original request. A changed `successor_record` lookup digest plus its recomputed `successor_digest` passes `scanRotation()` while the original `intent_digest` remains unchanged. `RecoverRotation()` then returns the substituted successor without error. Core recovery rejects that successor but consequently reports Unknown despite the known committed receipt.

Falsification: `NewRotateCommand()` permits another valid successor digest; the journal validates command scope, not this ancillary payload; migration constraints permit both updates. Existing corruption tests modify only hash columns and therefore miss this internally consistent substitution.

Compare retained **source and successor** canonical records against the exact original request before returning metadata. Add an integration regression preserving Committed plus an error and zero Session under this substitution.

Reviewed base/head: `704a1a40bc878be9eebda07abff159a3e0498596`; complete sixteen-file fingerprint: `c133a9419a1a8ad918214df84c6ee6583fc68c21f81205a87f381ac9aa2ab598`.

Read-only source review; supplied race/integration/vet results inspected, not rerun. Whole-unit release, mutation, runtime-role privilege, consumer/reference, family-replay revocation and remaining-profile gates remain unverified and outside this development-batch approval.
