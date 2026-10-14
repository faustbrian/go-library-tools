# Identity Platform Entry Point

## Execute in this order

1. Activate [REPAIR_AND_CHECKPOINT_GOAL.md](REPAIR_AND_CHECKPOINT_GOAL.md) first.
   It fixes the five execution issues, reconciles companions and completes the
   pending OTP implementation into a reviewed checkpoint committed and pushed
   to each affected repository's main branch, as directed by the user.
2. At its handoff, inspect the reported commits, verification and remaining
   qualification. The transition goal stops without starting another feature.
3. Explicitly activate [IMPLEMENTATION_GOAL.md](IMPLEMENTATION_GOAL.md).
   It is the full self-contained implementation prompt: one primary implementer,
   finite consumer milestones, operational baseline first, and all 67 units.
4. Complete the remaining milestones and qualify stable units incrementally.
   Only full requirement-level proof supports whole-platform completion.

Writing these files is not activation. ORCHESTRATOR_GOAL.md and WORKER_PROMPT.md
are retired execution entrypoints. Historical preflight/assignment/capture-release
records preserve evidence, not additional current start gates.

## Current execution and product sources

IMPLEMENTATION_GOAL.md is the single execution authority after user activation.
Use main branches and valid checkpoints; finish existing dirty worktree work
before main integration rather than discarding or reconstructing it.
Code belongs to standalone repositories resolved by
[REPOSITORY_TOPOLOGY.md](REPOSITORY_TOPOLOGY.md), not the coordination collection.

Product behavior remains defined by the full package goals and:
- [SHARED_ISSUER_PROFILE.md](SHARED_ISSUER_PROFILE.md);
- [PROGRAM.md](PROGRAM.md) and [COMMON_REQUIREMENTS.md](COMMON_REQUIREMENTS.md);
- [INVENTORY.md](INVENTORY.md), [DEPENDENCIES.md](DEPENDENCIES.md) and GOAL_MANIFEST.json;
- END_STATE.md, END_STATE_ACCEPTANCE.json and ACCEPTANCE_ARTIFACTS.json;
- BETTER_AUTH_PARITY.md, PARITY_DISPOSITIONS.json and the pinned upstream source;
- PUBLIC_CONTRACTS.json, API_OPERATIONS.md and OPERATION_SEMANTICS.json;
- protocol, transaction, lifecycle, security-event and configuration catalogs.

Read selected goals and owned/consumed rows on first entry, not every artifact
before each edit. Shared-issuer requirements are part of existing goals.
Full scope remains six primitive extensions, 61 product units, nineteen journey
categories and ten shared-issuer cases. HIBP/CAPTCHA/device authorization remain;
billing, SIWE and dedicated MCP/agent packages do not. This is a backend, not a UI.

## Progress and evidence

[IDENTITY_PROGRESS.md](IDENTITY_PROGRESS.md) records consumer outcomes, milestone
counts and outstanding qualification. [EXECUTION_LEDGER.md](EXECUTION_LEDGER.md)
retains the current frozen milestone note, source pins and attributable results.
Preserve prior native evidence and assignments. Unweighted milestone delivery,
unit integration and unit qualification are separate measures.

Qualification retains the exact package coverage, mutation and other mandatory
release gates. Local working behavior or a reviewed commit is not publication,
hosted CI, release, deployment or production proof.

The pinned Better Auth object database is needed for final parity/source proof;
reuse the existing checkout identified by BETTER_AUTH_REPOSITORY. Its availability
is not a global development start ceremony. Do not delete user-owned source
checkouts or acquire a new one merely to repeat completed preflight.
