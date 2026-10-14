# Goal: Repair Identity Execution and Deliver a Clean Main-Branch Checkpoint

The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT",
"SHOULD", "SHOULD NOT", "RECOMMENDED", "NOT RECOMMENDED", "MAY", and
"OPTIONAL" in this document are to be interpreted as described in BCP 14
[RFC2119] [RFC8174] when, and only when, they appear in all capitals, as
shown here.

## Objective and terminal boundary

User delivery direction (2026-10-14): commit and ordinarily push the work to
each affected Golib repository's main branch, with main-branch execution
thereafter. Finish and review existing dirty work first, retain concurrent
remote main changes, integrate only in-scope work, verify resulting main heads
and ordinary pushes, then stop before the revised goal activates. No force-push,
tags/releases, deployment or disposal of work is authorized.

Fix the five execution problems below, write the revised full identity-platform
implementation goal, reconcile its accompanying execution instructions, and
finish and commit the existing pending identity implementation into a clean,
reviewed checkpoint delivered to main. Then stop for revised-goal activation.
Do not continue into another platform feature or automatically activate that goal.

This is a finite transition task, not a new preflight, architecture redesign,
platform-wide qualification exercise, or implementation of all 67 units.
Creating or reading this prompt does not itself activate it. When the user
activates this transition goal, it authorizes its bounded documentation repair
and completion of the existing pending implementation; it does not activate
the subsequent full-platform goal.

## Inputs, workspaces and preservation

Use the existing control worktree, branch `feature/identity-platform`:
`/Users/brian/Developer/.worktrees/golib-go-library-tools-identity-platform`.
Its `.ai/identity-platform` directory is the active planning snapshot.

Use the existing runtime worktree, branch `feature/identity-platform`:
`/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`.
At prompt creation it has 220 modified/untracked paths above runtime commit
`1928f8ef5118cdd6980846551b5390a4146bae5b`. The control baseline is
`88dc950f2f44722f7dd6de5a513d46ae2174f0d4`, with existing uncommitted planning
and execution records. These are starting observations, not immutable live state.
Inspect each repository's actual head, branch, index, dirty files and live writers
once before taking ownership. Preserve all existing code, assignments, history,
valid evidence and unrelated work. Do not restart previously delivered slices.

Read the full existing implementation prompt once:
`/Users/brian/.codex/attachments/20198e15-da75-4bd2-b31f-c9294389a6b6/pasted-text-1.txt`.
Read the relevant current companion instructions and pending causal slice.
Recover decisions and test results from the existing ledger; do not reconstruct
every past turn. Resolve logical package labels through REPOSITORY_TOPOLOGY.md.
The collection `/Users/brian/Developer/golib` is not a repository or Go workspace.
Use the required workspace skills and verify repository/branch before writes.

All modules and current checks target Go 1.27. One persistent primary agent owns
all writes. Use GPT-6.1 Sol medium, isolated context (`fork_turns="none"`), for
bounded read-only investigation and independent final review only. Do not spawn
implementation workers or fixed custom roles with different model settings.
Retain native independent findings and dispositions; do not self-attest review.

## Exactly five execution repairs

Encode these in the revised goal and active companions, not merely in a report.
These change execution sequencing and reporting, not required product behavior.

### 1. Freeze complete consumer milestones and prevent scope creep

Replace undefined "vertical slice" scheduling with finite consumer milestones.
Before coding each milestone, record one short entry in EXECUTION_LEDGER.md:
selected profile; supported public entry point; exact consumer outcome; necessary
owners and consumed contracts; explicit exclusions; security invariants; finite
acceptance scenarios; and exact affected verification commands.

Freeze that entry after one bounded source/requirements pass. A helper, storage
contributor or constructor is not a milestone. A complete enrollment workflow
and a complete signin workflow can be separate milestones; splitting their
private helpers into deliveries cannot. Prerequisites are implemented as part
of the promised causal path, not as independent excursions.

Only a changed requirement or a concrete code/test/review defect in the promised
behavior can reopen the frozen scope. Record the reason in the same entry.
Adjacent variants, refactoring, generalization and polish remain later work.
Unsupported/unselected variants stay disabled and fail closed, never silently
downgrade security. Do not start another independent milestone before closure,
except when a concrete external dependency prevents continuation.

### 2. Require an operational baseline and measure delivered outcomes

The revised implementation goal starts with a runnable selected account/password/
session reference backend, using the existing implementation rather than
rebuilding it. Close these baseline milestones separately, in order:
- B1: real PostgreSQL bootstrap, registration and verification through captured
  delivery, signin, authenticated session read and logout; supported clean-consumer
  example and reproducible local run commands.
- B2: reset request -> captured delivery -> completion -> old-password rejection
  -> new-password signin -> required session revocation.
- B3: fresh reauthentication -> password change -> required session effects,
  including wrong/stale-proof denial.
Each has its own frozen acceptance and reviewed commit. Do not combine all
baseline gaps into another open-ended slice. No UI, deployment or application-
specific integration is required.

Make that baseline demonstrable before expanding into another capability family.
Existing committed behavior should be adopted and connected, not treated as
missing merely because its owning unit has outstanding release qualification.

Use a fixed, explicit milestone list for the full retained scope. Record delivered
consumer outcomes separately from unit integration and release qualification.
Preserve 67 units, 19 journey categories and ten shared-issuer cases; map them
to milestones without reclassifying them to inflate progress. Milestone-count
fractions are an unweighted delivery measure, not an overall effort percentage.
Overall percentages must state their basis and uncertainty. Do not repeat the
old 5–10% estimate as a measured result or invent new historical completion.

### 3. Separate finite slice proof from stabilization/release qualification

Keep three explicit check boundaries: focused red/green during implementation;
affected regression and consumer/security acceptance before milestone commit;
all required whole-unit/release gates when a unit stabilizes and before it is
marked verified. The selected milestone's check set is recorded before coding.
Add checks only for a concrete changed boundary or evidenced defect, not because
an adjacent reviewer artifact could be more elaborate.

Before usability is claimed, prove authorization, tenant/account isolation,
current factor authority, enumeration resistance, secret custody, origin/CSRF,
atomic one-use and protected effects, revocation, relevant rollback/replay,
competing commands and ambiguous-commit recovery. Use real durable adapters.
Run race/stress, parser fuzz and migration/compatibility checks where affected.
Do not call an unsafe workflow an MVP or postpone its security prerequisites.

Exact per-production-package 100% statement coverage, all viable mutants killed,
reviewed equivalent/invalid dispositions, applicable interoperability, benchmarks,
supply-chain, clean-consumer and aggregate release checks remain mandatory at
qualification. Do not lower them. Do not require full platform qualification to
commit a focused, proven development milestone. Close qualification as complete
units stabilize; do not leave every release obligation until the final audit.

Reuse unchanged evidence only with proven complete input identity and its
original execution identity. Rerun affected gates and reverse consumers, not
the whole repository after every helper, new SHA or documentation update.
Use the existing evidence tooling; do not build another cache/receipt framework.

### 4. Enforce closure and a delivery-based stop-loss

When the frozen consumer acceptance and affected checks pass and an independent
complete-final-diff review has no unresolved material findings, create the
coherent local commit and close the milestone. Do not add neighboring behavior
before committing. Post-review fixes require affected reruns and review of the
changed complete final diff, not new planning or repetition of unchanged proof.

Use existing six-hour active-work progress checkpoints. If two consecutive
checkpoints deliver no accepted consumer outcome, identify the exact obstacle
and change the implementation strategy or split an oversized milestone into
smaller complete consumer outcomes. Private-helper progress does not defeat
this trigger. This is a strategy trigger, not a deadline for unsafe acceptance.
Do not produce a new planning framework, restart preflight, or rename a failed
evidence cycle. Ask the user only when the choice materially changes scope,
security, authority or strategy and cannot be resolved from available evidence.

### 5. Remove contradictory accompanying execution authorities

Produce one canonical full implementation prompt at:
`.ai/identity-platform/IMPLEMENTATION_GOAL.md` in the control worktree.
It must be directly usable as the user's goal: include objective, all existing
67-unit inventory, 19 journey categories, ten shared-issuer requirements,
exclusions, execution rules, milestone order, acceptance, qualification and
terminal completion predicate. It must not just say "read another goal file."
Detailed product catalogs remain the source for exact operation/profile rows.

Make README.md the activation/order index and ORCHESTRATOR_GOAL.md a concise
retired-entry redirect, not a competing full execution prompt. Reconcile active
procedural clauses in COMMON_REQUIREMENTS.md, PROGRAM.md, REPOSITORY_TOPOLOGY.md
and the control AGENTS.md. Read selected package goals and applicable package
policies; replace obsolete worker/assignment, package-only scheduling and
whole-prerequisite-verification start clauses where they contradict the new
consumer-milestone model. Preserve dependency ownership: a consumed contract
requires committed, integrated, focused behavioral/compatibility proof; a stub
or compile result is not sufficient. Whole-unit verification remains required
for final verified promotion, not for every development start.

Do one targeted search for conflicting active rules in this bounded document
set and the package-goal start/verification sections. Change substantive active
conflicts, not every historical mention of "worker". Mark historical evidence
and retired execution documents non-authoritative for procedure. Retain their
records unchanged. Do not append yet another precedence amendment over obsolete
active rules. Product requirements in mixed documents remain binding.

Update IDENTITY_PROGRESS.md and the existing EXECUTION_LEDGER.md with the current
checkpoint, milestone progress and pending qualification; preserve old results
and assignment identities. Update existing goal pins mechanically only where
goal bodies actually changed. Do not regenerate unrelated catalogs, change
operation IDs, repair every historical validator, or weaken product semantics
to make old orchestration validators pass. Actual user direction is authority
for these execution corrections; do not demand exact-byte approval replays.

## Preserve the complete product target

The revised goal retains Better Auth baseline
`b8077b74ef9a80a7757220b72834349bd8de05c0`, all 67 units (six primitive extensions,
61 product units), all nineteen journeys and shared-issuer SI-01 through SI-10.
Retain HIBP, CAPTCHA adapters, device authorization, full backend administration,
organizations, MFA/passkeys, federation, SCIM and owned OAuth/OIDC protocols.
No identity administration UI or vendor-owned authentication engine is introduced.
Billing, SIWE, speculative integrations and dedicated MCP/agent packages remain
excluded. Login bridging, source-system migration, deployment and application/MCP
integration remain consumer/Mono/API responsibilities.

Preserve restricted exchange and non-expanding lineage/owner/actor/tenant/workload
authority; denial within 30 seconds including cached paths; original authoritative
evidence lifetime of at most 20 seconds without freshness reset; digest-only
legacy bare-SHA-256 key import; organization-owned keys surviving creator removal;
opt-in HTTP and no mandatory device dependency for base OAuth PostgreSQL.
Preserve stable IDs, opaque proof/secret custody, normalization provenance,
compatible primitive APIs, and transaction-owned atomic capability consumption.

## Finish the existing pending implementation — before handoff

After the five execution repairs are drafted, apply their bounded work rules
to close the current pending secondary-OTP work. Do not discard, stash away,
commit unfinished code merely to empty the worktree, or restart the slice.

Inspect the complete tracked and untracked diff once and record its finite
consumer promise. Starting scope is the existing native secondary-OTP enrollment,
encrypted delivery/confirmation, password-origin OTP challenge, delivered-code
completion, session/CSRF issuance, session reads and revocation, with selected
reference composition and the existing opt-in sending/verification HTTP paths.
Close required wiring and acceptance gaps for that promise; do not expand into
new OTP channels/providers, phone/passwordless products, trusted devices,
WebAuthn, more TOTP variants or a redesign of transaction/evidence machinery.

Complete these concrete outstanding obligations:

1. Exercise one reproducible selected consumer journey with real PostgreSQL:
   pending enrollment -> native encrypted queue -> bounded sender captures code
   -> public confirmation activates factor -> password signin yields OTP-only
   continuation -> selected HTTP send -> actual delivery -> selected HTTP verify
   -> session/CSRF -> supported session read -> revocation/factor change denies.
   Enrollment/confirmation may use their supported public API; add transport
   only when required by the already-promised selected public surface. No test
   may substitute a fake domain store, transaction owner or session issuer.
2. Exercise the affected denial, replay, rollback, competing-command and
   unknown-outcome paths and selected-feature omission. Use prior matching
   native results rather than rerunning unchanged predecessors. Preserve
   credential-free replay/recovery and no factor-method fallback.
3. Correct workspace-only integration-test placement/imports so public modules
   and the selected clean consumer build without permanent sibling replaces or
   cyclic dependencies. Update module/package manifests only when necessary.
4. Finish current API docs, adoption/example/run instructions and every affected
   changelog, including the latest OTP-send work. Remove temporary debug probes
   and owned test artifacts, not valid existing implementation.
5. Run the frozen affected checks on the final inputs, obtain an isolated
   independent review of the complete coherent runtime delta and its affected
   contracts/consumers, resolve concrete findings, then commit explicit paths
   through commit-with-intent. Review the whole causal slice, not only the last
   few files. Coherent dependency-order commits are allowed; every original
   pending path must be accounted for, not hidden by a partial staging pass.

This is a reviewed local development checkpoint, not a claim that every touched
unit or the platform is fully release-qualified. Record outstanding mandatory
qualification honestly. Complete all work needed for the promised consumer
behavior and compatibility; defer only nonessential polish or full qualification
that the approved sequencing places at the whole-unit/release boundary.

Commit related pending control contracts/catalog corrections and execution
records in the control repository after checking them against the actual runtime
checkpoint. Preserve unchanged evidence. No unrelated architectural/catalog
cleanup is required. Use task-owned disposable test resources, retain their
bounded-run checkpoints and clean them on success/failure. Prior test clusters
were removed; do not assume those services remain live.

## Bounded order, acceptance and stop condition

Execute in this order:
1. One bounded inspection; draft the five repairs and the full revised goal.
2. Finish, verify, independently review and commit the existing runtime work;
   integrate current main changes and ordinarily push the verified integration.
3. Finalize companion instructions and truthful progress against that checkpoint;
   run bounded document/link/manifest checks, commit them and deliver only the
   identity snapshot to current main without restoring unrelated old tooling.
4. Verify final heads and status of touched repositories; report and stop.

The documentation repair gets one edit batch and one bounded consistency pass;
fix concrete contradictions found in that pass without inventing new machinery.
Do not run a full historical preflight/validator suite for a wording correction.
Runtime verification follows the affected promise, not an arbitrary one-pass
limit; repeat only changed or failed checks. If the same concrete defect survives
two correction/review cycles, stop that strategy and make the required material
strategy decision rather than issuing another renamed patch loop.

Done means all five repairs are encoded; IMPLEMENTATION_GOAL.md is a complete
ready-to-use prompt; active companions agree; the promised pending OTP workflow
is proven and independently reviewed; all task-owned pending runtime/control
changes are committed and main pushes verified; final Git status contains no task-owned dirty or
staged paths; and disposable resources are cleaned. Unrelated pre-existing dirty
work must be preserved and disclosed, not erased to manufacture a "clean" report.

At handoff, report the exact absolute path/link to IMPLEMENTATION_GOAL.md, changed
companion files, repository-qualified checkpoint commits, supported consumer
behavior, exact focused verification outcomes, pending qualification and any
unrelated dirty paths. State explicitly that the revised implementation goal
has NOT been activated. Then stop; do not implement its operational baseline
or another feature until the user activates it.

Only the bounded main integrations and ordinary pushes above are authorized.
No PR, tag, release, deploy, rebase, reset, amend, force-push, hook bypass or
deletion of user work is authorized. Do not
fabricate evidence or treat a skipped/stale/failed mandatory check as a pass.
Unavailable checks or genuinely non-convergent design require a concrete bounded
handoff, not a false completion or an endless approval/verification program.
