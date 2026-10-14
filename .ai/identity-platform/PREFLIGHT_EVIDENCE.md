# Identity Platform Preflight Evidence

The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT",
"SHOULD", "SHOULD NOT", "RECOMMENDED", "NOT RECOMMENDED", "MAY", and
"OPTIONAL" in this document are to be interpreted as described in BCP 14
[RFC2119] [RFC8174] when, and only when, they appear in all capitals, as
shown here.

This coordinator-owned file is the durable preflight and resource registry for
one identity-platform execution. The coordinator MUST replace every `pending`
value, add rows as requirements are discovered, and commit the completed
record on the integration branch before the first worker assignment. It MUST
NOT contain credentials, tokens, provider payloads, secret-shaped values, or
unsanitized logs.

## Execution identity

| Field | Value |
| --- | --- |
| Recorded control execution baseline | 1b295c7cd17d4633442fe87c34d8e7505da16628 |
| Identity-platform base tree object | ba10d3d317bba9345720326ffb5046aa29fcd98d |
| Identity-platform base tree digest | sha256:170052193da133e496ea7b283772701c67541d6b1788dee19f112c5a26993592 |
| Integration branch | `feature/identity-platform` |
| Integration worktree | /Users/brian/Developer/.worktrees/golib-go-library-tools-identity-platform |
| Task-owned worktree parent | /Users/brian/Developer/.worktrees |
| Preflight input revision before the record commit | 1b295c7cd17d4633442fe87c34d8e7505da16628 |
| Preflight recorded at (RFC3339) | 2026-08-30T14:36:39Z |

The committed control execution baseline and preflight input MUST be existing
full commit hashes on the coordinator branch. The baseline is the finalized
coordinator-program checkpoint immediately before durable preflight; it is not
the control repository's immutable `main` revision. Package repository
`main` revisions remain recorded separately below.
The recorded identity-platform tree object and SHA-256 digest MUST be derived
from the exact committed base. The integration branch MUST be registered, its
registered worktree MUST have the same branch and HEAD, and that HEAD MUST
descend from the exact base and contain the preflight input. The integration worktree and
task-owned parent MUST be absolute, distinct from the repository root and the
home directory itself, registered by Git, and resolved before any destructive
cleanup.
Every post-commit execution check MUST use
`validate.rb --execution --clean-integration` plus the required previous
fixture quartet after the initial snapshot; a dirty, unreadable, detached,
or branch-HEAD-divergent integration worktree blocks spawn, merge, gates,
restart recovery, and final acceptance. Pre-commit proposed snapshots use
explicit fixture paths without clean mode.

## Repository integration identities

| Repository identity | Object format | Main base revision | Main base tree | Preflight integration baseline revision | Preflight integration baseline tree | Foundation | Integration branch | Integration worktree | Recorded at |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| github.com/faustbrian/go-library-tools | git-sha1 | 7290f888410e88b4b1025b040a7eb3ccfaed9e5f | git-sha1:c2f81d57347d2e4114ef9026a54d6bf452d89a8d | 1b295c7cd17d4633442fe87c34d8e7505da16628 | git-sha1:33976fe45ac932dc0f45d73abc7aa54402d9cd3f | existing | feature/identity-platform | /Users/brian/Developer/.worktrees/golib-go-library-tools-identity-platform | 2026-08-30T14:36:39Z |
| github.com/faustbrian/go-authentication | git-sha1 | 9926ce00a89c4f3187b12b4b25e3268158e138e6 | git-sha1:0779f23ed79e694b7e0298c832b090d9a1da3595 | 9926ce00a89c4f3187b12b4b25e3268158e138e6 | git-sha1:0779f23ed79e694b7e0298c832b090d9a1da3595 | existing | feature/identity-platform | /Users/brian/Developer/.worktrees/golib-go-authentication-identity-platform | 2026-08-30T14:36:39Z |
| github.com/faustbrian/go-authorization | git-sha1 | e08db3af02adde4ad67094d678febc7b12302d85 | git-sha1:a2e6f2c6d10301f244c54949669f2fa7cbe976e0 | e08db3af02adde4ad67094d678febc7b12302d85 | git-sha1:a2e6f2c6d10301f244c54949669f2fa7cbe976e0 | existing | feature/identity-platform | /Users/brian/Developer/.worktrees/golib-go-authorization-identity-platform | 2026-08-30T14:36:39Z |
| github.com/faustbrian/go-capability | git-sha1 | b0b21d15c9653decb0ff4a4fbffbe71c77d8d8fc | git-sha1:57fc01d61ab921c0dead07a585235a461011c537 | b0b21d15c9653decb0ff4a4fbffbe71c77d8d8fc | git-sha1:57fc01d61ab921c0dead07a585235a461011c537 | existing | feature/identity-platform | /Users/brian/Developer/.worktrees/golib-go-capability-identity-platform | 2026-08-30T14:36:39Z |
| github.com/faustbrian/go-identifier | git-sha1 | 23a8818faef585a87e3e31965d6af51bd7900478 | git-sha1:d944dffdd7e48f784d29fd93acb911f98b6375bd | 23a8818faef585a87e3e31965d6af51bd7900478 | git-sha1:d944dffdd7e48f784d29fd93acb911f98b6375bd | existing | feature/identity-platform | /Users/brian/Developer/.worktrees/golib-go-identifier-identity-platform | 2026-08-30T14:36:39Z |
| github.com/faustbrian/go-identity | git-sha1 | 8101066eb1607fb2a678a75ec808fee5d2bfa01f | git-sha1:62e4c9f34e23957f4347eb0a128373ed01750554 | 8101066eb1607fb2a678a75ec808fee5d2bfa01f | git-sha1:62e4c9f34e23957f4347eb0a128373ed01750554 | local-bootstrap | feature/identity-platform | /Users/brian/Developer/.worktrees/golib-go-identity-identity-platform | 2026-08-30T14:36:39Z |
| github.com/faustbrian/go-oauth-server | git-sha1 | 8101066eb1607fb2a678a75ec808fee5d2bfa01f | git-sha1:62e4c9f34e23957f4347eb0a128373ed01750554 | 8101066eb1607fb2a678a75ec808fee5d2bfa01f | git-sha1:62e4c9f34e23957f4347eb0a128373ed01750554 | local-bootstrap | feature/identity-platform | /Users/brian/Developer/.worktrees/golib-go-oauth-server-identity-platform | 2026-08-30T14:36:39Z |
| github.com/faustbrian/go-organization | git-sha1 | 8101066eb1607fb2a678a75ec808fee5d2bfa01f | git-sha1:62e4c9f34e23957f4347eb0a128373ed01750554 | 8101066eb1607fb2a678a75ec808fee5d2bfa01f | git-sha1:62e4c9f34e23957f4347eb0a128373ed01750554 | local-bootstrap | feature/identity-platform | /Users/brian/Developer/.worktrees/golib-go-organization-identity-platform | 2026-08-30T14:36:39Z |
| github.com/faustbrian/go-passkey | git-sha1 | 8101066eb1607fb2a678a75ec808fee5d2bfa01f | git-sha1:62e4c9f34e23957f4347eb0a128373ed01750554 | 8101066eb1607fb2a678a75ec808fee5d2bfa01f | git-sha1:62e4c9f34e23957f4347eb0a128373ed01750554 | local-bootstrap | feature/identity-platform | /Users/brian/Developer/.worktrees/golib-go-passkey-identity-platform | 2026-08-30T14:36:39Z |
| github.com/faustbrian/go-password | git-sha1 | 082f25b85d7975049b0629e1b57226384469af3b | git-sha1:7cd5697a11cbe8cf6ec96ce4b3498da0aeee7ac5 | 082f25b85d7975049b0629e1b57226384469af3b | git-sha1:7cd5697a11cbe8cf6ec96ce4b3498da0aeee7ac5 | existing | feature/identity-platform | /Users/brian/Developer/.worktrees/golib-go-password-identity-platform | 2026-08-30T14:36:39Z |
| github.com/faustbrian/go-scim | git-sha1 | 8101066eb1607fb2a678a75ec808fee5d2bfa01f | git-sha1:62e4c9f34e23957f4347eb0a128373ed01750554 | 8101066eb1607fb2a678a75ec808fee5d2bfa01f | git-sha1:62e4c9f34e23957f4347eb0a128373ed01750554 | local-bootstrap | feature/identity-platform | /Users/brian/Developer/.worktrees/golib-go-scim-identity-platform | 2026-08-30T14:36:39Z |
| github.com/faustbrian/go-sso | git-sha1 | 8101066eb1607fb2a678a75ec808fee5d2bfa01f | git-sha1:62e4c9f34e23957f4347eb0a128373ed01750554 | 8101066eb1607fb2a678a75ec808fee5d2bfa01f | git-sha1:62e4c9f34e23957f4347eb0a128373ed01750554 | local-bootstrap | feature/identity-platform | /Users/brian/Developer/.worktrees/golib-go-sso-identity-platform | 2026-08-30T14:36:39Z |
| github.com/faustbrian/go-webauthn | git-sha1 | 8101066eb1607fb2a678a75ec808fee5d2bfa01f | git-sha1:62e4c9f34e23957f4347eb0a128373ed01750554 | 8101066eb1607fb2a678a75ec808fee5d2bfa01f | git-sha1:62e4c9f34e23957f4347eb0a128373ed01750554 | local-bootstrap | feature/identity-platform | /Users/brian/Developer/.worktrees/golib-go-webauthn-identity-platform | 2026-08-30T14:36:39Z |

This append-only table records the immutable execution identity of the control
repository and every mapped package repository. Its repository set MUST equal
the exact topology closure: `github.com/faustbrian/go-library-tools` followed
by every standalone package repository in bytewise identity order. Each main
base, preflight integration baseline, tree, foundation, branch, and worktree is resolved
only in the named repository.
Cross-repository ancestry is forbidden. `Object format` is `git-sha1` or
`git-sha256`; both tree fields include that same prefix. Every integration
baseline MUST descend from its own recorded `main` base, and every worktree
MUST be the exact registered `feature/identity-platform` worktree for that
repository. Current integration `HEAD` MAY advance only as a descendant of
this immutable preflight baseline. The control baseline MUST equal the
recorded control execution baseline. Rows are recorded before any package
assignment and remain the lower-bound authority for repository-qualified
final gates, blocked-state projection, and no-push scope.
`Foundation` is `existing` for a repository that already existed when the
program began and `local-bootstrap` only for a missing repository created
under the narrow five-file bootstrap exception. Only `local-bootstrap` rows
MUST have a root `main` commit containing exactly `.gitattributes`,
`.gitignore`, `AGENTS.md`, `CLAUDE.md`, and `LICENSE`; classifying an existing
repository as a bootstrap does not authorize rewriting it.

## Platform evidence trust

| Field | Value |
| --- | --- |
| Trust document path | `.ai/identity-platform/PLATFORM_EVIDENCE_TRUST.json` |
| Trust document commit | 1b295c7cd17d4633442fe87c34d8e7505da16628 |
| Trust document digest | sha256:ac722b46ef7b323bb0fa7fa273d1f4a7a65bb32d34721f5ad815a43e23b42a19 |
| Platform thread ID | 019ff56e-5561-7f23-8cbe-b325ec164521 |
| Classification | available |
| Blocker or attestation | capture-policy-bound |

The capture policy MUST exist with the exact recorded bytes on the control
execution baseline. The coordinator MUST NOT create or replace it after that
baseline.
`Classification` is `available` only when the validator verifies the committed
policy and its required Git-custody controls. It means the bounded coordinator
capture lane is usable; it MUST NOT mean platform authenticity, runtime
realization, or delivery was independently attested.

Repository rows, coordinator captures, worker text, deterministic IDs, and
hashes are custody evidence only. They MUST NOT be represented as platform
signatures, immutable runtime attestations, or proof of delivery.

`PLATFORM_EVIDENCE_TRUST.json` retains its historical filename but is a capture
policy, not a cryptographic trust store. It is canonical compact JSON with
ordered fields `schema_version`, `schema`, `capture_source`, `limitations`, and
`required_controls`. Its schema is
`identity-platform.evidence-capture-policy.v1`; it records that the
collaboration API exposes tool-visible identities, statuses, and messages but
does not export signatures or immutable delivery receipts. The file pins the
required append-only first-commit custody, exact-byte digests, isolated worker
branch/worktree ownership, coordinator re-verification, independent package
worker, and explicit user semantic-authority controls.

Every tool-visible capture binding is
`<path>@<first-commit-containing-exact-bytes>@sha256:<canonical-envelope-digest>`.
The path is beneath `.ai/identity-platform/evidence/platform-events/`. An
envelope is canonical compact JSON with ordered fields `schema_version`,
`schema`, `capture_source`, `thread_id`, `event_id`, `event_type`, `role`,
`actor_id`, `task_id`, `agent_id`, `turn_ordinal`, `recorded_at`, `content`,
`content_length`, `content_sha256`, and `claims`.
`schema` is `identity-platform.tool-visible-capture.v1`; content is exact UTF-8 text;
its length counts bytes. Claims are a canonical object containing every
capture-specific identity named below. The event ID is globally unique within
the execution. The coordinator transcribes only tool-visible output and commits
the exact bytes immediately. The artifact proves committed custody and
ordering, not platform authenticity, and MUST NOT be called signed evidence.

Coordinator execution captures use event type
`coordinator-gate-capture` for the producer and
`coordinator-verifier-capture` for the independent verifier. Their `content`
is the exact canonical receipt or verifier-attestation bytes, excluding the
companion event binding itself to avoid a digest fixed point. Claims are an
ordered canonical object containing `execution_receipt_path`,
`execution_receipt_sha256`, `execution_identity`, `tested_revision`,
`input_root`, `capture_authority`, `producer_argv`, `verifier_argv`,
`canonical_workdir`, both producer and verifier executable realpaths and
SHA-256 digests, the version and environment-probe argv/stdout/stderr, producer
and verifier stdout/stderr/exit fields, byte lengths and digests,
`captured_output_sha256`, `raw_capture_sha256`, `output_artifact_path`,
`artifact_payload_binding`, `verifier_attestation_sha256`, `started_at`,
`completed_at`, and `scope`.
Coverage/mutation receipts additionally bind `package_discovery_argv`,
`mutant_discovery_argv`, `package_manifest_sha256`, and
`mutant_manifest_sha256`. `scope` contains ordered keys `kind`, `unit`,
`artifact_id`, `profile_id`, and `claim_ids`; inapplicable nullable values are
JSON `null`, and `claim_ids` is bytewise sorted and unique. `kind` is exactly
`local`, `external`, `acceptance`, or `final-gate`. Producer and verifier
captures MUST use distinct capture IDs and bind the distinct coordinator-owned
process results. They are not platform signatures.

Because receipt schemas are closed and self-embedding a capture creates
a fixed point, bindings live in the companion table below. The table is keyed
by the immutable execution identity and exact receipt path/digest. A row is
valid only when both committed captures reproduce every receipt execution field,
the producer and verifier executable/argv/workdir/environment/input-root,
stdout/stderr/exit/timestamps, the independent verifier-attestation digest,
and the complete scope. A missing, replayed, or mismatched capture
leaves that gate unverified even when its receipt JSON says `passed`.
The receipt payload intentionally omits its own `execution_identity`,
`execution_receipt_path`, `execution_receipt_sha256`, and the fixed
`artifact_payload_binding`; the validator derives the first three claims from
the companion row and exact receipt binding and requires the last to equal
`record.artifact_hashes[path,sha256]` before verifying the capture. All
other claims come from the receipt bytes.

## Platform gate capture bindings

| Execution identity | Scope | Receipt binding | Producer capture event binding | Verifier capture event binding | Recorded at |
| --- | --- | --- | --- | --- | --- |

`Scope` is canonical compact JSON with ordered keys `kind`, `unit`,
`artifact_id`, `profile_id`, and `claim_ids`. The row is append-only; its two
event bindings and receipt binding are immutable for the execution identity.

## Final gate bindings

| Repository identity | Gate ID | Worktree | Resolved argv | Evidence record binding | Runner receipt binding | Producer capture event binding | Verifier capture event binding | Tested revision | Tested tree | Repository input root | Program input root | Recorded at |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |

This append-only table is the sole custody index for the exact repository/gate
closure: the four control-repository gates and the three package-repository
gates for every mapped package repository. The `(Repository identity, Gate
ID)` pair MUST be unique and exactly match that closure. `Worktree`, tested
revision and tested tree MUST identify a descendant of that repository's
immutable preflight integration baseline. All gate rows for one repository
MUST bind the same tested revision, tested tree, and repository input root;
every row MUST share the same aggregate program input root. Before terminal
validation this table MAY be empty, but a nonempty partial closure is invalid.
Once all 67 units are `verified`, the complete exact closure is REQUIRED.
At that terminal phase each package tested revision MUST have no file delta to
its current integration `HEAD`; the control repository MAY differ only by the
closed final-input bookkeeping paths accepted by the final-report contract.
An untested behavior-affecting descendant invalidates the final-gate closure.
`Resolved argv` is
canonical compact JSON containing the exact absolute executable and argument
array; placeholders, shell strings, and unresolved base revisions are
forbidden. The executable MUST exist and be executable, and its exact realpath
MUST equal the current coordinator-owned `which make`, `which codex`, or
`RbConfig.ruby` identity, as applicable; a task-owned same-basename executable
is forbidden. Its complete argument array MUST equal this closed mapping:
`affected-release-gates` is `make ci-changed BASE=<repository-recorded-base>`;
`complete-repository-gate` is `make ci`; `final-complete-diff-review` is
`ruby <absolute-final_review_gate.rb>`;
`inventory-validation` is `make inventory`;
`pinned-upstream-validation` is `ruby
<absolute-generate_upstream_leaves.rb> --check
<absolute-BETTER_AUTH_REPOSITORY-realpath>`;
`program-final-input-acceptance` is `ruby
<absolute-identity-platform-validate.rb> --execution --clean-integration`;
and `structural-validation` is `ruby
<absolute-identity-platform-validate.rb>`.
The repository-owned review gate invokes the installed `codex review` with the
exact recorded base and validator-owned prompt, then writes its captured output
to the runner-owned raw-output path. The review prompt requires canonical
compact JSON with ordered keys `critical` and `important`. Its independent
verifier MUST be the repository-owned `final_review_verifier.rb`; it accepts
only exact `{"critical":[],"important":[]}` bytes and emits a digest-bound
zero-count attestation. Process exit alone never proves a clean review. Extra flags,
wrappers, shell launchers, aliases, and a command assigned to another gate are
forbidden. The affected-release command MUST contain the exact recorded
committed `main` base of its named package repository. Evidence records live at
`.ai/identity-platform/evidence/final-gates/<repository-slug>/<gate-id>.json`; runner-v4 receipts
live at `.ai/identity-platform/evidence/final-gates/executions/<execution-identity>.json`.
Both bindings use `<path>@<first-commit-containing-exact-bytes>@sha256:<digest>`.
Each final-gate evidence record is canonical compact JSON with ordered fields
`schema_version`, `schema`, `record_repository`, `tested_repository`, `gate_id`,
`outcome`, `resolved_argv`, `tested_revision`, `tested_tree`,
`gate_execution_revision`, `gate_execution_tree`, `repository_input_root`, `program_input_root`,
`execution_identity`, `receipt_path`, `receipt_sha256`, and `record_digest`.
`schema_version` is `2`, `schema` is `identity-platform.final-gate.v2`, and
`outcome` is `passed`. `record_digest` is the SHA-256 of canonical compact JSON
over the preceding fields with `record_digest` omitted. The record's repository
identities, gate ID, argv, revisions, trees, input roots, execution identity, and receipt path/digest MUST
equal this row and the committed capture exactly.
The producer and verifier event bindings MUST equal the matching verified
`Platform gate capture bindings` row whose scope kind is `final-gate`, unit and
profile are JSON `null`, `artifact_id` is `<repository-identity>:<gate-id>`, and
`claim_ids` is the one-element array containing that same repository-qualified
ID. Tested repository state and both input roots MUST equal the canonical
report and receipt; all rows MUST be current
before `PROGRAM-COMPLETE`.

`BETTER_AUTH_REPOSITORY` MUST name an existing absolute local Git repository
whose `origin` is exactly `https://github.com/better-auth/better-auth.git` and
which contains the pinned commit, SHA-1 object format, and every source object
declared by `UPSTREAM_SURFACE.json`. The resolved realpath MUST be the final
argument captured by the upstream gate. A missing checkout, missing pinned
object, different remote, different object format, or source-object mismatch
fails the gate; checked-in `UPSTREAM_SURFACE.json` and `UPSTREAM_LEAVES.json`
alone are not evidence. The repository may be a normal clone or local mirror
and the gate performs no network access. It MUST be supplied before
orchestration as a persistent user-owned prerequisite, MUST remain available
through terminal validation, and MUST NOT be entered in the task-owned resource
registry or deleted by the coordinator.

## Blocked audit binding

| Persistent goal ID | Platform event binding | Threshold state | Execution identity | Recorded at |
| --- | --- | --- | --- | --- |

This append-only row is required only for `PROGRAM-BLOCKED`. Its coordinator
capture type is `persistent-goal-blocked-audit`. Its exact claims bind the current thread, persistent goal ID,
threshold state, consecutive blocked-audit count, execution identity, exact
current repository snapshots and aggregate program input root, remaining
blockers, repository-qualified reconciliations, and exhausted independent
frontier. `Threshold state` is canonical compact JSON with ordered fields
`required_count`, `observed_count`, `history_event_ids`,
`repository_snapshots`, `program_input_root`, `blockers`, `reconciliations`,
and `exhausted_frontier`. Legacy scalar `final_input_revision` and `input_root`
fields are forbidden. `repository_snapshots` MUST be the exact topology closure
in the same order as `Repository integration identities`; every closed row has
exactly the ordered fields `repository_identity`, `object_format`, `revision`,
and `tree`. Its revision is the last behavior-affecting input revision for that
named repository, its tree is the exact tree at that revision, and both object
identities use that repository's recorded object format. Omitted, duplicated,
substituted, or reordered repositories are invalid.

`blockers` is sorted by the canonical compact JSON bytes of each closed object
and every object has exactly the ordered fields `repository_identity`, `id`,
`category`, `unit`, `evidence`, and `required_user_action`. Its repository
identity owns the unfinished unit; `id` is the safe blocker ID without the
inventory `blocker:` prefix; `category` is one of `credential`,
`external-infrastructure`, `product-decision`, or `user-authority`; `unit` is
the exact unfinished unit; and the evidence and required action are nonempty
newline-free strings. The committed blocker objects,
not later final-report prose, are the authority for every
reported blocker detail. Their `(repository_identity, id, unit)` tuples MUST equal the complete
unfinished inventory blocker set exactly. The execution identity is SHA-256 over
the canonical compact JSON object whose ordered fields are
`repository_snapshots` and `program_input_root`; a digest over scalar revisions
or one repository is invalid. `program_input_root` MUST equal the independently
recomputed aggregate input root for the exact snapshot set.

`reconciliations` MUST contain one closed row for every repository in the same
exact order. Each row has exactly the ordered fields `repository_identity`,
`assignment_ids`, `resource_ids`, `assignments_root`, and `resources_root`.
The ID arrays are sorted and unique. Assignment and resource roots are SHA-256
over their repository-owned canonical JSON projections; assignment rows contain
every assigned ledger row's unit, generation, inventory status and owner, task,
assignment, worker and integration commits, gate revision and fingerprint,
external disposition, and transition, while resource rows contain every
resource's ID, type, owner, state, target, and cleanup evidence. Projection rows
are sorted by unit or resource ID and object keys are byte-sorted before
hashing. Every root MUST equal its independently recomputed current projection;
matching IDs alone are insufficient.

Each snapshot revision MUST be an ancestor of the current clean `HEAD` within
its named repository only. The terminal-audit capture MUST bind the complete
snapshot array byte-for-byte; its control-repository commit ancestry is checked
only against the control-repository snapshot. Cross-repository ancestry is
forbidden. Only the validated
append-only ledger, inventory, preflight, and evidence/report bookkeeping paths
permitted by the final-input contract MAY change in a repository after its
snapshot revision. The state is the validator's
projection of committed repository facts, not platform-authenticated state.
Missing or unverifiable committed capture evidence means the blocked threshold
is unproved and stopping is forbidden.
`PROGRAM-BLOCKED` is terminal only from a successfully inspected, clean
repository state after every permitted bookkeeping/evidence change has been
committed and every task-owned resource and assignment has the exact reconciled
disposition reported here. Uncommitted semantic or orchestration state makes
the blocked predicate false; it is work to reconcile, not a durable blocker.

Final-report claims derived from the validator's committed cleanup,
parity-disposition, or aggregate predicate inputs use the exact binding
`program:<claim-id>@<execution-identity>@<terminal-history-root>`.
The validator independently recomputes every repository snapshot and the
aggregate behavior input root from tracked inputs plus the five attested
environment identities, then derives the execution identity from that exact
repository-qualified payload. It separately computes
the terminal-history root from the committed execution ledger, inventory,
preflight history, and evidence tree, excluding only the final report itself to
avoid circular self-digestion. It accepts the binding only for a claim derived
true from those complete committed histories. Bare `program:<claim-id>`,
a binding to another input or history root, or substitution for a final-gate
receipt is invalid.

Authorization summaries, every authorization boundary, and
`program-complete.7` MUST instead reference the committed
`final-gate:structural-validation` evidence record. That record's independently
captured runner receipt, record and receipt commits, exact SHA-256 digests,
tested revision, gate-execution revision, and input root are the evidence that
the validator accepted the complete authorization, transition, and assignment
topology histories. A `program:` projection for any of these claims is
synthetic self-certification and is invalid.

Every `history_event_ids` entry names exactly one committed coordinator capture
of a tool-visible blocked observation in the current thread. The IDs are in
observation order and unique; their capture ordinals MUST
be consecutive, MUST precede the terminal blocked-audit event, and their
timestamps MUST NOT follow it. Each observation's claims bind the same
`persistent_goal_id`, execution identity, and canonical `blocked_state`; that
state is the threshold object with `required_count`, `observed_count`, and
`history_event_ids` removed. Its content is the canonical compact JSON for that
exact state. `observed_count` MUST equal the number of resolved observation
captures and `required_count` MUST be positive. A count without those exact
committed captures is not threshold evidence. This proves repository
persistence, not a platform-wide blocked-turn audit.

## No-push coordinator assertion

| Repository-qualified Git-observable scope | Coordinator event binding | Outcome | Recorded at |
| --- | --- | --- | --- |

The coordinator MUST still not push. It records at most one explicit assertion
covering the repository-qualified intervals in its scope, but the row is not a
command audit and is not completion evidence. The scope is canonical compact
JSON with the sole ordered field `repositories`. That array MUST equal the
exact topology closure in `Repository integration identities` order. Every
closed repository row has exactly the ordered fields `repository_identity`,
`program_base_revision`, `final_input_revision`, `remote_names`, `refspecs`,
`observed_local_branch`, and `observed_local_head`. Legacy scalar program-base,
final-input, or input-root fields outside the repository rows are forbidden.
For each named repository, `program_base_revision` MUST equal its recorded main
base and `final_input_revision` MUST equal its matching terminal repository
snapshot revision. The validator MUST prove their ancestry only inside that
repository. Remote/ref arrays are nonempty, sorted, and unique;
remotes equal that repository's locally configured remotes, refspecs are
exactly `["refs/*"]`, and the observed branch/head equal its local Git state.
The final input revision MUST be an ancestor of the observed head in the same
repository. Omitted, duplicated, substituted, reordered, or scope-mismatched
repository rows are invalid. `Outcome` is exactly
`coordinator-asserts-no-push`. The event is coordinator-authored and MUST use
the same assertion text and complete bounded repository array. It MUST expose
`assertion_verified=false` and the literal limitation
`no-complete-command-audit-or-remote-non-delivery-proof`. Neither this row nor
its zero/absence claim may satisfy a verified terminal requirement. It does not
prove complete command history, remote state, network non-delivery, or actions
by another actor; the final report MUST surface that limitation unchanged.

## Worker assignment authorizations

| Unit | Generation | Control repository | Package repository | Package integration baseline | Package baseline tree | Control assignment commit | Control assignment tree | Assignment goal path | Rendered prompt | Prompt digest | Model | Reasoning | Fork turns | Subagents | Package scope | Reserved descendants | Goal digest | Authorized by | Status |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |

Each proposed worker spawn MUST have exactly one durable `authorized` row
before spawn. This row is pre-spawn authorization, not proof of the runtime
visible through the collaboration tool. The coordinator MUST commit the complete rendered prompt bytes at
the attributable repository path and bind their exact SHA-256 digest, unit,
generation, control and package repository identities, package integration
baseline and tree, control assignment commit and tree, immutable assignment-time
goal path, `gpt-5.6-sol`, `medium`,
`none` fork turns, `false` subagents, exact canonical package scope, complete
reserved-descendant set, pinned goal-body digest, and `coordinator`
authorization. Pending or missing authorization blocks spawn. The first commit
that adds the exact row and rendered prompt is the immutable assignment
authorization checkpoint. Returned-scope validation MUST validate the baseline-
to-assignment-authorization-to-runtime-capture-to-release-row envelope
separately and MUST validate only package integration-baseline-to-tip history as
worker-authored package work. Control checkpoints MUST NOT be merged into the
package repository or used in package ancestry checks.

## Worker runtime attestations

The historical heading is retained for table compatibility; rows are
coordinator captures, not cryptographic runtime attestations.

| Unit | Generation | Control repository | Package repository | Worker task | Agent ID | Model | Reasoning | Fork turns | Subagents | Platform event binding | Recorded at |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |

Immediately after the exact readiness receipt from the spawn turn, the
coordinator MUST append exactly one row from the tool-visible spawn result.
`Worker task` MUST be returned by the collaboration tool. `Agent ID` is the
coordinator's stable assignment-local identifier when the tool does not expose
one; it MUST NOT be described as immutable platform identity.
`Model`, `Reasoning`, `Fork turns`, and `Subagents` MUST equal the requested
values `gpt-5.6-sol`, `medium`, `none`, and `false`. `Platform event binding`
MUST bind one canonical coordinator capture of that result, including its
thread ID, worker task, agent ID, settings, timestamp, and exact content digest.
The coordinator MUST commit this row in the control repository before
authorizing package work. It MUST NOT merge that control commit into the
package worker branch or test the two repositories for shared ancestry. Worker self-report is not a
substitute. A missing, duplicate, identity-drifted, or settings-drifted runtime
row blocks release, worker-return acceptance, and integration. The worker MUST
remain readiness-only until the committed control capture and release row are
repository-qualified and the later activation directive is sent.

## Worker release handshakes

| Unit | Generation | Control repository | Package repository | Worker task | Readiness event binding | Control assignment authorization checkpoint | Control runtime-capture commit | Release event binding | Released at |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |

The first worker turn is readiness-only. `Readiness event binding` MUST bind a
coordinator capture of the tool-visible `worker-message` from the exact worker in
the recorded thread whose role is `assistant` and whose complete UTF-8 content
is exactly `READY-AND-WAITING <unit> g<generation>
package=<package-repository-identity>@<package-head> control=<control-repository-identity>@<assignment-authorization-checkpoint>`
with no prefix, suffix, or newline.
Before recording the runtime capture, the coordinator MUST validate that
capture against the base-pinned capture policy. After committing it in the
control repository, the coordinator sends the exact non-authorizing `PREPARE-RELEASE` directive
defined by `WORKER_PROMPT.md`. `Release event binding` captures the coordinator
send invocation/result for those exact bytes, same thread, worker, runtime capture, and
authorization checkpoint. The coordinator then commits the row in the control
repository and sends `ACTIVATE` naming that repository-qualified row commit. No worker-
authored package commit may precede that activation. An early worker mutation,
missing capture, duplicate capture ID, wrong actor/thread, preparation before runtime, or
identity/commit mismatch permanently
invalidates that assignment generation and requires safe abandonment; a later
row cannot retroactively authorize the early work.

The collaboration API provides no delivery receipt for `PREPARE-RELEASE` or `ACTIVATE`; the
validator therefore does not claim to prove delivery. The enforceable release
boundary is the package branch remaining at the recorded baseline until the
exact control-repository row exists, plus the worker's instruction to wait for
the matching activation. Coordinator acceptance MUST resolve the control row
and package commit independently and reject identity, revision, tree, or digest
drift; the remaining send/receive ordering is an explicit tool limitation.

## Worker return attestations

| Unit | Generation | Control repository | Package repository | Worker task | Return event binding | Report binding | Package worker tip | Ordered package worker commits | Recorded at |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |

The coordinator MUST record exactly one return row before accepting or
integrating worker output. `Return event binding` binds a coordinator capture
of the tool-visible `worker-message` from the assigned assistant in the recorded thread.
Its exact content bytes MUST equal the committed worker report named by
`Report binding`; the report binding is
`<repository-path>@<first-commit-containing-exact-bytes>@sha256:<digest>`.
The capture claims MUST bind the unit, generation, worker task, tool-visible agent ID,
assignment authorization checkpoint, runtime event ID, release event ID,
report binding, exact worker tip, and complete ordered worker-commit list.
The list is oldest to newest, contains exactly the worker-authored commits
reachable from the package baseline through the tip, and contains no
coordinator custody commit. Cross-repository ancestry is forbidden. The return
capture ordinal MUST follow the PREPARE
send result. The worker asserts activation was observed; the API cannot prove
delivery. Missing, replayed, stale, or mismatched return capture
blocks acceptance and integration. This coordinator-authored capture is
custody evidence only; acceptance also requires independent diff review and
authoritative checks at the exact returned committed worker tip.

## Semantic authorization requests

| Approval ID | Proposal manifest binding | Affected closure | Status | User message event binding | Blocker or action | Recorded at |
| --- | --- | --- | --- | --- | --- | --- |

This append-only request lifecycle is `awaiting-user -> authorized -> applied`
or `awaiting-user -> superseded`. `Affected closure` is canonical compact JSON
with ordered keys `units`, `contract_ids`, `claim_ids`, `artifact_ids`, and
`source_paths`; every value is a bytewise sorted unique string array. Source
paths and contract IDs MUST equal the proposal manifest closure, and units,
claims, and artifacts MUST equal the transitive consumers derived from the
authoritative catalogs. Until an executable catalog-specific derivation proves
a narrower closure, `units` MUST conservatively contain all 67 program units;
the coordinator MUST NOT self-certify a smaller set. An
`awaiting-user` row uses `—` for the user event and
`blocker:platform-user-authority` until a valid event is available. While
awaiting authority, no affected unit may advance assignment, integration, or
verification state and no affected authoritative source may change; independent
lanes continue. Missing platform trust or an unverifiable user event never
falls back to coordinator-authored text and remains awaiting user authority.

## User semantic authorizations

| Approval ID | Proposal manifest binding | Change classes | Previous semantic root | Proposed semantic root | User message event binding | Status | Recorded at |
| --- | --- | --- | --- | --- | --- | --- | --- |

This append-only table is the only execution-time authority for semantic
changes classified by `PROGRAM.md`. The coordinator MUST present the exact
canonical statement below and the user MUST send it back byte-for-byte as a
standalone line; conversational agreement or approval of a summary is not
authorization:

`AUTHORIZE identity-platform <approval-id> proposal=<repository-path>@<sha256-manifest-digest> previous=<sha256-root> proposed=<sha256-root> classes=<sorted-comma-separated-classes>`

`Approval ID` MUST be `user:<approval-id>` using the exact safe ID in that user
statement; `<approval-id>` MUST match `[a-zA-Z0-9._-]+`, and each root token
MUST be `sha256:<64-lowercase-hex>`. `Change classes` MUST be one or
more sorted values from `scope`, `ownership`, `dependency`, `public-api`,
`operation-semantics`, `parity`, `protocol`, `acceptance`, `configuration`,
`security`, `transaction`, `lifecycle`, or `goal`.

Before asking the user, the coordinator MUST commit one canonical proposal
manifest at
`.ai/identity-platform/evidence/semantic-proposals/<approval-id>.json` and every
full proposed source blob it names beneath
`.ai/identity-platform/evidence/semantic-proposals/<approval-id>/`. The manifest
uses schema `identity-platform.semantic-proposal.v1`, canonical UTF-8 JSON with
no insignificant whitespace or trailing newline, and ordered fields
`schema_version`, `schema`, `approval_id`, `change_classes`,
`previous_semantic_root`, `proposed_semantic_root`, and `entries`.
`schema_version` MUST be `1`; `approval_id` MUST be `user:<approval-id>`;
classes MUST be bytewise sorted and unique. Entries MUST be sorted and unique
by `source_path` and use ordered fields `source_path`, `contract_ids`,
`previous_content_sha256`, `proposed_blob_path`, and
`proposed_content_sha256`. Contract IDs MUST be exhaustive, bytewise sorted and
unique, using an empty array only when the complete file rather than named rows
is the semantic authority. Each source path MUST be a semantic source governed
by `PROGRAM.md`; each proposal blob path MUST remain inside that approval's
proposal directory and contain the complete exact replacement bytes for its
source, not a patch or summary.

At the proposal-manifest commit, every source MUST still equal its recorded
previous digest and every committed proposal blob MUST equal its proposed
digest. `previous_semantic_root` and `proposed_semantic_root` are SHA-256 roots
over canonical JSON projections with ordered fields `change_classes` and
`entries`; each projected entry uses ordered fields `source_path`,
`contract_ids`, and `content_sha256`, with the last field set to the
corresponding previous or proposed content digest. The line's proposal path,
manifest digest, roots, classes, and approval ID MUST reproduce the committed
manifest exactly. `Proposal manifest binding` records
`<path>@<first-commit-containing-exact-bytes>@sha256:<manifest-digest>`; that
commit MUST precede the authorized row. `User message event binding` MUST bind
the committed capture carrying the exact
standalone UTF-8 user line including no trailing newline. `Status` MUST be `authorized`,
`applied`, or `superseded`; an `applied` or `superseded` row requires an exact
earlier committed `authorized` row for the same approval and roots.

`User message event binding` MUST identify one committed coordinator capture
of the tool-visible `conversation-message` whose thread equals the recorded thread, role is
`user`, timestamp follows the exact proposal-manifest commit, and complete
UTF-8 content equals the standalone line byte-for-byte with no prefix, suffix,
or newline. The capture ID MUST be unique across the execution and its payload
digest and byte length MUST verify under the base-pinned capture policy. The
same binding is retained byte-for-byte in terminal rows. This establishes the
exact user-authority bytes used by the coordinator, not cryptographic platform
authenticity.

The coordinator MUST NOT create an `authorized` row unless the user explicitly
provided that exact user message binding the proposal manifest. Coordinator judgment, worker
findings, source control authorship, a validator repair, or a mechanically
generated digest is not authorization. The authorized row MUST be committed
while the previous semantic root is current; the applied row is committed only
after every authoritative source equals its exact committed proposed blob and
the proposed root is current. Any affected source or contract ID absent from
the user-bound manifest, any different proposed byte, or any unlisted semantic
byte change is unauthorized and blocks assignment, integration, verification,
and final acceptance for its affected closure. If the platform cannot supply a
verifiable user-message event, the request remains `awaiting-user`; authority is
unavailable rather than inferred.

## Tool and environment lanes

| Requirement/profile | Required by units or claims | Classification | Version/environment identity | Evidence path or blocking claim |
| --- | --- | --- | --- | --- |
| Go toolchain and repository gate tools | all units | available | 10 exact binary/source identities bound by sha256:3b7b33a4ee41b35a398403f12057b818b1033d7d4ae9fff965e2c286b2fb4cba | .ai/identity-platform/evidence/preflight/tool-environment-inventory.json |
| PostgreSQL profile | PostgreSQL adapters and reference journeys | available | postgres:18.4-alpine@sha256:9a8afca54e7861fd90fab5fdf4c42477a6b1cb7d293595148e674e0a3181de15 | .ai/identity-platform/evidence/preflight/tool-environment-inventory.json |
| Valkey profile | Valkey adapters and reference journeys | available | valkey/valkey:9.1.0-alpine@sha256:c9b77919daeba2c02ad954d0c844cc4e7142069d177b89c5fd771f405daf9e02 | .ai/identity-platform/evidence/preflight/tool-environment-inventory.json |
| race, fuzz, leak, and mutation tooling | all applicable unit gates | available | Go race/fuzz/leak and gremlins identities bound by sha256:3b7b33a4ee41b35a398403f12057b818b1033d7d4ae9fff965e2c286b2fb4cba | .ai/identity-platform/evidence/preflight/tool-environment-inventory.json |
| browser and interoperability harnesses | browser journeys and provider/protocol claims | unavailable | browser=unavailable; interoperability=not-yet-needed | blocker:browser-harness |

`Classification` MUST be exactly `available`, `unavailable`, or
`not-yet-needed`. Each unavailable row MUST name every acceptance claim it can
block. Availability is preflight information only and MUST NOT be reused as a
passing interoperability result.

## External evidence lanes

| Safe profile ID | Consuming units | Exact acceptance claim IDs | Classification | Credential source metadata | Evidence path or blocker | Evidence record commit | Evidence digest or blocker |
| --- | --- | --- | --- | --- | --- | --- | --- |
| captcha-captchafox | `identity/risk/captcha/captchafox` | `identity.admin.user-password-set`<br>`identity.anonymous.signin`<br>`identity.magic-link.request`<br>`identity.oauth.onetap-callback`<br>`identity.oauth.signin-start`<br>`identity.oauth.signin-token`<br>`identity.otp.send`<br>`identity.otp.signin`<br>`identity.passkey.signin-options`<br>`identity.passkey.signin-verify`<br>`identity.password.change`<br>`identity.password.reset-complete`<br>`identity.password.reset-request`<br>`identity.password.set`<br>`identity.password.signin`<br>`identity.password.signup`<br>`identity.phone.password-reset-complete`<br>`identity.phone.password-reset-request`<br>`identity.phone.password-signin`<br>`identity.phone.send-verification`<br>`identity.phone.signin`<br>`identity.sso.discover`<br>`identity.sso.signin-start`<br>`identity.username.signin`<br>`identity.username.signup`<br>`journey.abuse-controls.v1` | not-yet-needed | none | not-yet-needed:captcha-captchafox | — | not-yet-needed:captcha-captchafox |
| captcha-hcaptcha | `identity/risk/captcha/hcaptcha` | `identity.admin.user-password-set`<br>`identity.anonymous.signin`<br>`identity.magic-link.request`<br>`identity.oauth.onetap-callback`<br>`identity.oauth.signin-start`<br>`identity.oauth.signin-token`<br>`identity.otp.send`<br>`identity.otp.signin`<br>`identity.passkey.signin-options`<br>`identity.passkey.signin-verify`<br>`identity.password.change`<br>`identity.password.reset-complete`<br>`identity.password.reset-request`<br>`identity.password.set`<br>`identity.password.signin`<br>`identity.password.signup`<br>`identity.phone.password-reset-complete`<br>`identity.phone.password-reset-request`<br>`identity.phone.password-signin`<br>`identity.phone.send-verification`<br>`identity.phone.signin`<br>`identity.sso.discover`<br>`identity.sso.signin-start`<br>`identity.username.signin`<br>`identity.username.signup`<br>`journey.abuse-controls.v1` | not-yet-needed | none | not-yet-needed:captcha-hcaptcha | — | not-yet-needed:captcha-hcaptcha |
| captcha-recaptcha | `identity/risk/captcha/recaptcha` | `identity.admin.user-password-set`<br>`identity.anonymous.signin`<br>`identity.magic-link.request`<br>`identity.oauth.onetap-callback`<br>`identity.oauth.signin-start`<br>`identity.oauth.signin-token`<br>`identity.otp.send`<br>`identity.otp.signin`<br>`identity.passkey.signin-options`<br>`identity.passkey.signin-verify`<br>`identity.password.change`<br>`identity.password.reset-complete`<br>`identity.password.reset-request`<br>`identity.password.set`<br>`identity.password.signin`<br>`identity.password.signup`<br>`identity.phone.password-reset-complete`<br>`identity.phone.password-reset-request`<br>`identity.phone.password-signin`<br>`identity.phone.send-verification`<br>`identity.phone.signin`<br>`identity.sso.discover`<br>`identity.sso.signin-start`<br>`identity.username.signin`<br>`identity.username.signup`<br>`journey.abuse-controls.v1` | not-yet-needed | none | not-yet-needed:captcha-recaptcha | — | not-yet-needed:captcha-recaptcha |
| captcha-turnstile | `identity/risk/captcha/turnstile` | `identity.admin.user-password-set`<br>`identity.anonymous.signin`<br>`identity.magic-link.request`<br>`identity.oauth.onetap-callback`<br>`identity.oauth.signin-start`<br>`identity.oauth.signin-token`<br>`identity.otp.send`<br>`identity.otp.signin`<br>`identity.passkey.signin-options`<br>`identity.passkey.signin-verify`<br>`identity.password.change`<br>`identity.password.reset-complete`<br>`identity.password.reset-request`<br>`identity.password.set`<br>`identity.password.signin`<br>`identity.password.signup`<br>`identity.phone.password-reset-complete`<br>`identity.phone.password-reset-request`<br>`identity.phone.password-signin`<br>`identity.phone.send-verification`<br>`identity.phone.signin`<br>`identity.sso.discover`<br>`identity.sso.signin-start`<br>`identity.username.signin`<br>`identity.username.signup`<br>`journey.abuse-controls.v1` | not-yet-needed | none | not-yet-needed:captcha-turnstile | — | not-yet-needed:captcha-turnstile |
| hibp-pwned-passwords | `identity/risk/hibp` | `journey.email-password-username.v1` | not-yet-needed | none | not-yet-needed:hibp-pwned-passwords | — | not-yet-needed:hibp-pwned-passwords |
| protocol-openid-conformance-suite | `identity/oauth`<br>`oauth-server`<br>`oauth-server/oidc`<br>`sso/oidc` | `identity.oauth-server.authorize`<br>`identity.oauth-server.client-create`<br>`identity.oauth-server.client-delete`<br>`identity.oauth-server.client-get`<br>`identity.oauth-server.client-get-public`<br>`identity.oauth-server.client-get-public-prelogin`<br>`identity.oauth-server.client-list`<br>`identity.oauth-server.client-rotate-secret`<br>`identity.oauth-server.client-update`<br>`identity.oauth-server.consent-delete`<br>`identity.oauth-server.consent-get`<br>`identity.oauth-server.consent-list`<br>`identity.oauth-server.consent-update`<br>`identity.oauth-server.continue`<br>`identity.oauth-server.device-approve`<br>`identity.oauth-server.device-authorize`<br>`identity.oauth-server.device-deny`<br>`identity.oauth-server.device-inspect`<br>`identity.oauth-server.device-token`<br>`identity.oauth-server.discovery-oauth`<br>`identity.oauth-server.discovery-oidc`<br>`identity.oauth-server.dynamic-register`<br>`identity.oauth-server.end-session`<br>`identity.oauth-server.introspect`<br>`identity.oauth-server.jwks`<br>`identity.oauth-server.protected-resource-metadata`<br>`identity.oauth-server.resource-verify`<br>`identity.oauth-server.revoke`<br>`identity.oauth-server.session-token`<br>`identity.oauth-server.token`<br>`identity.oauth-server.userinfo`<br>`journey.oauth-oidc-provider.v1` | not-yet-needed | none | not-yet-needed:protocol-openid-conformance-suite | — | not-yet-needed:protocol-openid-conformance-suite |
| protocol-simplesamlphp | `sso/saml` | `identity.sso.break-glass.consume`<br>`identity.sso.break-glass.issue`<br>`identity.sso.directory-sync-apply`<br>`identity.sso.directory-sync-cancel`<br>`identity.sso.directory-sync-start`<br>`identity.sso.directory-sync-status`<br>`identity.sso.discover`<br>`identity.sso.domain-challenge`<br>`identity.sso.domain-verify`<br>`identity.sso.enforcement.update`<br>`identity.sso.oauth-callback`<br>`identity.sso.oidc-callback`<br>`identity.sso.oidc-logout`<br>`identity.sso.oidc-logout-complete`<br>`identity.sso.provider.credentials-rotate`<br>`identity.sso.provider.delete`<br>`identity.sso.provider.disable`<br>`identity.sso.provider.enable`<br>`identity.sso.provider.get`<br>`identity.sso.provider.list`<br>`identity.sso.provider.register-oauth`<br>`identity.sso.provider.register-oidc`<br>`identity.sso.provider.register-saml`<br>`identity.sso.provider.update`<br>`identity.sso.saml-acs`<br>`identity.sso.saml-idp-init`<br>`identity.sso.saml-logout-start`<br>`identity.sso.saml-metadata`<br>`identity.sso.saml-slo`<br>`identity.sso.saml-start`<br>`identity.sso.signin-start`<br>`journey.enterprise-federation.v1` | not-yet-needed | none | not-yet-needed:protocol-simplesamlphp | — | not-yet-needed:protocol-simplesamlphp |
| protocol-unboundid-scim2-sdk | `scim`<br>`scim/organization`<br>`scim/postgres` | `journey.scim.v1` | not-yet-needed | none | not-yet-needed:protocol-unboundid-scim2-sdk | — | not-yet-needed:protocol-unboundid-scim2-sdk |
| protocol-web-platform-tests | `passkey`<br>`webauthn` | `journey.mfa-webauthn.v1` | not-yet-needed | none | not-yet-needed:protocol-web-platform-tests | — | not-yet-needed:protocol-web-platform-tests |
| provider-apple | `identity/oauth/providers` | `identity.account.access-token`<br>`identity.account.link-start`<br>`identity.account.link-token`<br>`identity.account.list`<br>`identity.account.provider-info`<br>`identity.account.refresh-token`<br>`identity.account.unlink`<br>`identity.oauth.callback`<br>`identity.oauth.callback-form-post`<br>`identity.oauth.logout-complete`<br>`identity.oauth.logout-start`<br>`identity.oauth.onetap-callback`<br>`identity.oauth.onetap-start`<br>`identity.oauth.popup-complete`<br>`identity.oauth.provider-apple-client-secret-sign`<br>`identity.oauth.provider-list`<br>`identity.oauth.proxy-forward`<br>`identity.oauth.signin-start`<br>`identity.oauth.signin-token`<br>`journey.social-oauth.v1` | not-yet-needed | none | not-yet-needed:provider-apple | — | not-yet-needed:provider-apple |
| provider-atlassian | `identity/oauth/providers` | `identity.account.access-token`<br>`identity.account.link-start`<br>`identity.account.link-token`<br>`identity.account.list`<br>`identity.account.provider-info`<br>`identity.account.refresh-token`<br>`identity.account.unlink`<br>`identity.oauth.callback`<br>`identity.oauth.callback-form-post`<br>`identity.oauth.logout-complete`<br>`identity.oauth.logout-start`<br>`identity.oauth.onetap-callback`<br>`identity.oauth.onetap-start`<br>`identity.oauth.popup-complete`<br>`identity.oauth.provider-apple-client-secret-sign`<br>`identity.oauth.provider-list`<br>`identity.oauth.proxy-forward`<br>`identity.oauth.signin-start`<br>`identity.oauth.signin-token`<br>`journey.social-oauth.v1` | not-yet-needed | none | not-yet-needed:provider-atlassian | — | not-yet-needed:provider-atlassian |
| provider-auth0 | `identity/oauth/providers` | `identity.account.access-token`<br>`identity.account.link-start`<br>`identity.account.link-token`<br>`identity.account.list`<br>`identity.account.provider-info`<br>`identity.account.refresh-token`<br>`identity.account.unlink`<br>`identity.oauth.callback`<br>`identity.oauth.callback-form-post`<br>`identity.oauth.logout-complete`<br>`identity.oauth.logout-start`<br>`identity.oauth.onetap-callback`<br>`identity.oauth.onetap-start`<br>`identity.oauth.popup-complete`<br>`identity.oauth.provider-apple-client-secret-sign`<br>`identity.oauth.provider-list`<br>`identity.oauth.proxy-forward`<br>`identity.oauth.signin-start`<br>`identity.oauth.signin-token`<br>`journey.social-oauth.v1` | not-yet-needed | none | not-yet-needed:provider-auth0 | — | not-yet-needed:provider-auth0 |
| provider-cognito | `identity/oauth/providers` | `identity.account.access-token`<br>`identity.account.link-start`<br>`identity.account.link-token`<br>`identity.account.list`<br>`identity.account.provider-info`<br>`identity.account.refresh-token`<br>`identity.account.unlink`<br>`identity.oauth.callback`<br>`identity.oauth.callback-form-post`<br>`identity.oauth.logout-complete`<br>`identity.oauth.logout-start`<br>`identity.oauth.onetap-callback`<br>`identity.oauth.onetap-start`<br>`identity.oauth.popup-complete`<br>`identity.oauth.provider-apple-client-secret-sign`<br>`identity.oauth.provider-list`<br>`identity.oauth.proxy-forward`<br>`identity.oauth.signin-start`<br>`identity.oauth.signin-token`<br>`journey.social-oauth.v1` | not-yet-needed | none | not-yet-needed:provider-cognito | — | not-yet-needed:provider-cognito |
| provider-discord | `identity/oauth/providers` | `identity.account.access-token`<br>`identity.account.link-start`<br>`identity.account.link-token`<br>`identity.account.list`<br>`identity.account.provider-info`<br>`identity.account.refresh-token`<br>`identity.account.unlink`<br>`identity.oauth.callback`<br>`identity.oauth.callback-form-post`<br>`identity.oauth.logout-complete`<br>`identity.oauth.logout-start`<br>`identity.oauth.onetap-callback`<br>`identity.oauth.onetap-start`<br>`identity.oauth.popup-complete`<br>`identity.oauth.provider-apple-client-secret-sign`<br>`identity.oauth.provider-list`<br>`identity.oauth.proxy-forward`<br>`identity.oauth.signin-start`<br>`identity.oauth.signin-token`<br>`journey.social-oauth.v1` | not-yet-needed | none | not-yet-needed:provider-discord | — | not-yet-needed:provider-discord |
| provider-dropbox | `identity/oauth/providers` | `identity.account.access-token`<br>`identity.account.link-start`<br>`identity.account.link-token`<br>`identity.account.list`<br>`identity.account.provider-info`<br>`identity.account.refresh-token`<br>`identity.account.unlink`<br>`identity.oauth.callback`<br>`identity.oauth.callback-form-post`<br>`identity.oauth.logout-complete`<br>`identity.oauth.logout-start`<br>`identity.oauth.onetap-callback`<br>`identity.oauth.onetap-start`<br>`identity.oauth.popup-complete`<br>`identity.oauth.provider-apple-client-secret-sign`<br>`identity.oauth.provider-list`<br>`identity.oauth.proxy-forward`<br>`identity.oauth.signin-start`<br>`identity.oauth.signin-token`<br>`journey.social-oauth.v1` | not-yet-needed | none | not-yet-needed:provider-dropbox | — | not-yet-needed:provider-dropbox |
| provider-facebook | `identity/oauth/providers` | `identity.account.access-token`<br>`identity.account.link-start`<br>`identity.account.link-token`<br>`identity.account.list`<br>`identity.account.provider-info`<br>`identity.account.refresh-token`<br>`identity.account.unlink`<br>`identity.oauth.callback`<br>`identity.oauth.callback-form-post`<br>`identity.oauth.logout-complete`<br>`identity.oauth.logout-start`<br>`identity.oauth.onetap-callback`<br>`identity.oauth.onetap-start`<br>`identity.oauth.popup-complete`<br>`identity.oauth.provider-apple-client-secret-sign`<br>`identity.oauth.provider-list`<br>`identity.oauth.proxy-forward`<br>`identity.oauth.signin-start`<br>`identity.oauth.signin-token`<br>`journey.social-oauth.v1` | not-yet-needed | none | not-yet-needed:provider-facebook | — | not-yet-needed:provider-facebook |
| provider-figma | `identity/oauth/providers` | `identity.account.access-token`<br>`identity.account.link-start`<br>`identity.account.link-token`<br>`identity.account.list`<br>`identity.account.provider-info`<br>`identity.account.refresh-token`<br>`identity.account.unlink`<br>`identity.oauth.callback`<br>`identity.oauth.callback-form-post`<br>`identity.oauth.logout-complete`<br>`identity.oauth.logout-start`<br>`identity.oauth.onetap-callback`<br>`identity.oauth.onetap-start`<br>`identity.oauth.popup-complete`<br>`identity.oauth.provider-apple-client-secret-sign`<br>`identity.oauth.provider-list`<br>`identity.oauth.proxy-forward`<br>`identity.oauth.signin-start`<br>`identity.oauth.signin-token`<br>`journey.social-oauth.v1` | not-yet-needed | none | not-yet-needed:provider-figma | — | not-yet-needed:provider-figma |
| provider-github | `identity/oauth/providers` | `identity.account.access-token`<br>`identity.account.link-start`<br>`identity.account.link-token`<br>`identity.account.list`<br>`identity.account.provider-info`<br>`identity.account.refresh-token`<br>`identity.account.unlink`<br>`identity.oauth.callback`<br>`identity.oauth.callback-form-post`<br>`identity.oauth.logout-complete`<br>`identity.oauth.logout-start`<br>`identity.oauth.onetap-callback`<br>`identity.oauth.onetap-start`<br>`identity.oauth.popup-complete`<br>`identity.oauth.provider-apple-client-secret-sign`<br>`identity.oauth.provider-list`<br>`identity.oauth.proxy-forward`<br>`identity.oauth.signin-start`<br>`identity.oauth.signin-token`<br>`journey.social-oauth.v1` | not-yet-needed | none | not-yet-needed:provider-github | — | not-yet-needed:provider-github |
| provider-gitlab | `identity/oauth/providers` | `identity.account.access-token`<br>`identity.account.link-start`<br>`identity.account.link-token`<br>`identity.account.list`<br>`identity.account.provider-info`<br>`identity.account.refresh-token`<br>`identity.account.unlink`<br>`identity.oauth.callback`<br>`identity.oauth.callback-form-post`<br>`identity.oauth.logout-complete`<br>`identity.oauth.logout-start`<br>`identity.oauth.onetap-callback`<br>`identity.oauth.onetap-start`<br>`identity.oauth.popup-complete`<br>`identity.oauth.provider-apple-client-secret-sign`<br>`identity.oauth.provider-list`<br>`identity.oauth.proxy-forward`<br>`identity.oauth.signin-start`<br>`identity.oauth.signin-token`<br>`journey.social-oauth.v1` | not-yet-needed | none | not-yet-needed:provider-gitlab | — | not-yet-needed:provider-gitlab |
| provider-google | `identity/oauth/providers` | `identity.account.access-token`<br>`identity.account.link-start`<br>`identity.account.link-token`<br>`identity.account.list`<br>`identity.account.provider-info`<br>`identity.account.refresh-token`<br>`identity.account.unlink`<br>`identity.oauth.callback`<br>`identity.oauth.callback-form-post`<br>`identity.oauth.logout-complete`<br>`identity.oauth.logout-start`<br>`identity.oauth.onetap-callback`<br>`identity.oauth.onetap-start`<br>`identity.oauth.popup-complete`<br>`identity.oauth.provider-apple-client-secret-sign`<br>`identity.oauth.provider-list`<br>`identity.oauth.proxy-forward`<br>`identity.oauth.signin-start`<br>`identity.oauth.signin-token`<br>`journey.social-oauth.v1` | not-yet-needed | none | not-yet-needed:provider-google | — | not-yet-needed:provider-google |
| provider-gumroad | `identity/oauth/providers` | `identity.account.access-token`<br>`identity.account.link-start`<br>`identity.account.link-token`<br>`identity.account.list`<br>`identity.account.provider-info`<br>`identity.account.refresh-token`<br>`identity.account.unlink`<br>`identity.oauth.callback`<br>`identity.oauth.callback-form-post`<br>`identity.oauth.logout-complete`<br>`identity.oauth.logout-start`<br>`identity.oauth.onetap-callback`<br>`identity.oauth.onetap-start`<br>`identity.oauth.popup-complete`<br>`identity.oauth.provider-apple-client-secret-sign`<br>`identity.oauth.provider-list`<br>`identity.oauth.proxy-forward`<br>`identity.oauth.signin-start`<br>`identity.oauth.signin-token`<br>`journey.social-oauth.v1` | not-yet-needed | none | not-yet-needed:provider-gumroad | — | not-yet-needed:provider-gumroad |
| provider-hubspot | `identity/oauth/providers` | `identity.account.access-token`<br>`identity.account.link-start`<br>`identity.account.link-token`<br>`identity.account.list`<br>`identity.account.provider-info`<br>`identity.account.refresh-token`<br>`identity.account.unlink`<br>`identity.oauth.callback`<br>`identity.oauth.callback-form-post`<br>`identity.oauth.logout-complete`<br>`identity.oauth.logout-start`<br>`identity.oauth.onetap-callback`<br>`identity.oauth.onetap-start`<br>`identity.oauth.popup-complete`<br>`identity.oauth.provider-apple-client-secret-sign`<br>`identity.oauth.provider-list`<br>`identity.oauth.proxy-forward`<br>`identity.oauth.signin-start`<br>`identity.oauth.signin-token`<br>`journey.social-oauth.v1` | not-yet-needed | none | not-yet-needed:provider-hubspot | — | not-yet-needed:provider-hubspot |
| provider-huggingface | `identity/oauth/providers` | `identity.account.access-token`<br>`identity.account.link-start`<br>`identity.account.link-token`<br>`identity.account.list`<br>`identity.account.provider-info`<br>`identity.account.refresh-token`<br>`identity.account.unlink`<br>`identity.oauth.callback`<br>`identity.oauth.callback-form-post`<br>`identity.oauth.logout-complete`<br>`identity.oauth.logout-start`<br>`identity.oauth.onetap-callback`<br>`identity.oauth.onetap-start`<br>`identity.oauth.popup-complete`<br>`identity.oauth.provider-apple-client-secret-sign`<br>`identity.oauth.provider-list`<br>`identity.oauth.proxy-forward`<br>`identity.oauth.signin-start`<br>`identity.oauth.signin-token`<br>`journey.social-oauth.v1` | not-yet-needed | none | not-yet-needed:provider-huggingface | — | not-yet-needed:provider-huggingface |
| provider-kakao | `identity/oauth/providers` | `identity.account.access-token`<br>`identity.account.link-start`<br>`identity.account.link-token`<br>`identity.account.list`<br>`identity.account.provider-info`<br>`identity.account.refresh-token`<br>`identity.account.unlink`<br>`identity.oauth.callback`<br>`identity.oauth.callback-form-post`<br>`identity.oauth.logout-complete`<br>`identity.oauth.logout-start`<br>`identity.oauth.onetap-callback`<br>`identity.oauth.onetap-start`<br>`identity.oauth.popup-complete`<br>`identity.oauth.provider-apple-client-secret-sign`<br>`identity.oauth.provider-list`<br>`identity.oauth.proxy-forward`<br>`identity.oauth.signin-start`<br>`identity.oauth.signin-token`<br>`journey.social-oauth.v1` | not-yet-needed | none | not-yet-needed:provider-kakao | — | not-yet-needed:provider-kakao |
| provider-keycloak | `identity/oauth/providers` | `identity.account.access-token`<br>`identity.account.link-start`<br>`identity.account.link-token`<br>`identity.account.list`<br>`identity.account.provider-info`<br>`identity.account.refresh-token`<br>`identity.account.unlink`<br>`identity.oauth.callback`<br>`identity.oauth.callback-form-post`<br>`identity.oauth.logout-complete`<br>`identity.oauth.logout-start`<br>`identity.oauth.onetap-callback`<br>`identity.oauth.onetap-start`<br>`identity.oauth.popup-complete`<br>`identity.oauth.provider-apple-client-secret-sign`<br>`identity.oauth.provider-list`<br>`identity.oauth.proxy-forward`<br>`identity.oauth.signin-start`<br>`identity.oauth.signin-token`<br>`journey.social-oauth.v1` | not-yet-needed | none | not-yet-needed:provider-keycloak | — | not-yet-needed:provider-keycloak |
| provider-kick | `identity/oauth/providers` | `identity.account.access-token`<br>`identity.account.link-start`<br>`identity.account.link-token`<br>`identity.account.list`<br>`identity.account.provider-info`<br>`identity.account.refresh-token`<br>`identity.account.unlink`<br>`identity.oauth.callback`<br>`identity.oauth.callback-form-post`<br>`identity.oauth.logout-complete`<br>`identity.oauth.logout-start`<br>`identity.oauth.onetap-callback`<br>`identity.oauth.onetap-start`<br>`identity.oauth.popup-complete`<br>`identity.oauth.provider-apple-client-secret-sign`<br>`identity.oauth.provider-list`<br>`identity.oauth.proxy-forward`<br>`identity.oauth.signin-start`<br>`identity.oauth.signin-token`<br>`journey.social-oauth.v1` | not-yet-needed | none | not-yet-needed:provider-kick | — | not-yet-needed:provider-kick |
| provider-line | `identity/oauth/providers` | `identity.account.access-token`<br>`identity.account.link-start`<br>`identity.account.link-token`<br>`identity.account.list`<br>`identity.account.provider-info`<br>`identity.account.refresh-token`<br>`identity.account.unlink`<br>`identity.oauth.callback`<br>`identity.oauth.callback-form-post`<br>`identity.oauth.logout-complete`<br>`identity.oauth.logout-start`<br>`identity.oauth.onetap-callback`<br>`identity.oauth.onetap-start`<br>`identity.oauth.popup-complete`<br>`identity.oauth.provider-apple-client-secret-sign`<br>`identity.oauth.provider-list`<br>`identity.oauth.proxy-forward`<br>`identity.oauth.signin-start`<br>`identity.oauth.signin-token`<br>`journey.social-oauth.v1` | not-yet-needed | none | not-yet-needed:provider-line | — | not-yet-needed:provider-line |
| provider-linear | `identity/oauth/providers` | `identity.account.access-token`<br>`identity.account.link-start`<br>`identity.account.link-token`<br>`identity.account.list`<br>`identity.account.provider-info`<br>`identity.account.refresh-token`<br>`identity.account.unlink`<br>`identity.oauth.callback`<br>`identity.oauth.callback-form-post`<br>`identity.oauth.logout-complete`<br>`identity.oauth.logout-start`<br>`identity.oauth.onetap-callback`<br>`identity.oauth.onetap-start`<br>`identity.oauth.popup-complete`<br>`identity.oauth.provider-apple-client-secret-sign`<br>`identity.oauth.provider-list`<br>`identity.oauth.proxy-forward`<br>`identity.oauth.signin-start`<br>`identity.oauth.signin-token`<br>`journey.social-oauth.v1` | not-yet-needed | none | not-yet-needed:provider-linear | — | not-yet-needed:provider-linear |
| provider-linkedin | `identity/oauth/providers` | `identity.account.access-token`<br>`identity.account.link-start`<br>`identity.account.link-token`<br>`identity.account.list`<br>`identity.account.provider-info`<br>`identity.account.refresh-token`<br>`identity.account.unlink`<br>`identity.oauth.callback`<br>`identity.oauth.callback-form-post`<br>`identity.oauth.logout-complete`<br>`identity.oauth.logout-start`<br>`identity.oauth.onetap-callback`<br>`identity.oauth.onetap-start`<br>`identity.oauth.popup-complete`<br>`identity.oauth.provider-apple-client-secret-sign`<br>`identity.oauth.provider-list`<br>`identity.oauth.proxy-forward`<br>`identity.oauth.signin-start`<br>`identity.oauth.signin-token`<br>`journey.social-oauth.v1` | not-yet-needed | none | not-yet-needed:provider-linkedin | — | not-yet-needed:provider-linkedin |
| provider-microsoft | `identity/oauth/providers` | `identity.account.access-token`<br>`identity.account.link-start`<br>`identity.account.link-token`<br>`identity.account.list`<br>`identity.account.provider-info`<br>`identity.account.refresh-token`<br>`identity.account.unlink`<br>`identity.oauth.callback`<br>`identity.oauth.callback-form-post`<br>`identity.oauth.logout-complete`<br>`identity.oauth.logout-start`<br>`identity.oauth.onetap-callback`<br>`identity.oauth.onetap-start`<br>`identity.oauth.popup-complete`<br>`identity.oauth.provider-apple-client-secret-sign`<br>`identity.oauth.provider-list`<br>`identity.oauth.proxy-forward`<br>`identity.oauth.signin-start`<br>`identity.oauth.signin-token`<br>`journey.social-oauth.v1` | not-yet-needed | none | not-yet-needed:provider-microsoft | — | not-yet-needed:provider-microsoft |
| provider-microsoft-entra-id | `identity/oauth/providers` | `identity.account.access-token`<br>`identity.account.link-start`<br>`identity.account.link-token`<br>`identity.account.list`<br>`identity.account.provider-info`<br>`identity.account.refresh-token`<br>`identity.account.unlink`<br>`identity.oauth.callback`<br>`identity.oauth.callback-form-post`<br>`identity.oauth.logout-complete`<br>`identity.oauth.logout-start`<br>`identity.oauth.onetap-callback`<br>`identity.oauth.onetap-start`<br>`identity.oauth.popup-complete`<br>`identity.oauth.provider-apple-client-secret-sign`<br>`identity.oauth.provider-list`<br>`identity.oauth.proxy-forward`<br>`identity.oauth.signin-start`<br>`identity.oauth.signin-token`<br>`journey.social-oauth.v1` | not-yet-needed | none | not-yet-needed:provider-microsoft-entra-id | — | not-yet-needed:provider-microsoft-entra-id |
| provider-naver | `identity/oauth/providers` | `identity.account.access-token`<br>`identity.account.link-start`<br>`identity.account.link-token`<br>`identity.account.list`<br>`identity.account.provider-info`<br>`identity.account.refresh-token`<br>`identity.account.unlink`<br>`identity.oauth.callback`<br>`identity.oauth.callback-form-post`<br>`identity.oauth.logout-complete`<br>`identity.oauth.logout-start`<br>`identity.oauth.onetap-callback`<br>`identity.oauth.onetap-start`<br>`identity.oauth.popup-complete`<br>`identity.oauth.provider-apple-client-secret-sign`<br>`identity.oauth.provider-list`<br>`identity.oauth.proxy-forward`<br>`identity.oauth.signin-start`<br>`identity.oauth.signin-token`<br>`journey.social-oauth.v1` | not-yet-needed | none | not-yet-needed:provider-naver | — | not-yet-needed:provider-naver |
| provider-notion | `identity/oauth/providers` | `identity.account.access-token`<br>`identity.account.link-start`<br>`identity.account.link-token`<br>`identity.account.list`<br>`identity.account.provider-info`<br>`identity.account.refresh-token`<br>`identity.account.unlink`<br>`identity.oauth.callback`<br>`identity.oauth.callback-form-post`<br>`identity.oauth.logout-complete`<br>`identity.oauth.logout-start`<br>`identity.oauth.onetap-callback`<br>`identity.oauth.onetap-start`<br>`identity.oauth.popup-complete`<br>`identity.oauth.provider-apple-client-secret-sign`<br>`identity.oauth.provider-list`<br>`identity.oauth.proxy-forward`<br>`identity.oauth.signin-start`<br>`identity.oauth.signin-token`<br>`journey.social-oauth.v1` | not-yet-needed | none | not-yet-needed:provider-notion | — | not-yet-needed:provider-notion |
| provider-okta | `identity/oauth/providers` | `identity.account.access-token`<br>`identity.account.link-start`<br>`identity.account.link-token`<br>`identity.account.list`<br>`identity.account.provider-info`<br>`identity.account.refresh-token`<br>`identity.account.unlink`<br>`identity.oauth.callback`<br>`identity.oauth.callback-form-post`<br>`identity.oauth.logout-complete`<br>`identity.oauth.logout-start`<br>`identity.oauth.onetap-callback`<br>`identity.oauth.onetap-start`<br>`identity.oauth.popup-complete`<br>`identity.oauth.provider-apple-client-secret-sign`<br>`identity.oauth.provider-list`<br>`identity.oauth.proxy-forward`<br>`identity.oauth.signin-start`<br>`identity.oauth.signin-token`<br>`journey.social-oauth.v1` | not-yet-needed | none | not-yet-needed:provider-okta | — | not-yet-needed:provider-okta |
| provider-patreon | `identity/oauth/providers` | `identity.account.access-token`<br>`identity.account.link-start`<br>`identity.account.link-token`<br>`identity.account.list`<br>`identity.account.provider-info`<br>`identity.account.refresh-token`<br>`identity.account.unlink`<br>`identity.oauth.callback`<br>`identity.oauth.callback-form-post`<br>`identity.oauth.logout-complete`<br>`identity.oauth.logout-start`<br>`identity.oauth.onetap-callback`<br>`identity.oauth.onetap-start`<br>`identity.oauth.popup-complete`<br>`identity.oauth.provider-apple-client-secret-sign`<br>`identity.oauth.provider-list`<br>`identity.oauth.proxy-forward`<br>`identity.oauth.signin-start`<br>`identity.oauth.signin-token`<br>`journey.social-oauth.v1` | not-yet-needed | none | not-yet-needed:provider-patreon | — | not-yet-needed:provider-patreon |
| provider-paybin | `identity/oauth/providers` | `identity.account.access-token`<br>`identity.account.link-start`<br>`identity.account.link-token`<br>`identity.account.list`<br>`identity.account.provider-info`<br>`identity.account.refresh-token`<br>`identity.account.unlink`<br>`identity.oauth.callback`<br>`identity.oauth.callback-form-post`<br>`identity.oauth.logout-complete`<br>`identity.oauth.logout-start`<br>`identity.oauth.onetap-callback`<br>`identity.oauth.onetap-start`<br>`identity.oauth.popup-complete`<br>`identity.oauth.provider-apple-client-secret-sign`<br>`identity.oauth.provider-list`<br>`identity.oauth.proxy-forward`<br>`identity.oauth.signin-start`<br>`identity.oauth.signin-token`<br>`journey.social-oauth.v1` | not-yet-needed | none | not-yet-needed:provider-paybin | — | not-yet-needed:provider-paybin |
| provider-paypal | `identity/oauth/providers` | `identity.account.access-token`<br>`identity.account.link-start`<br>`identity.account.link-token`<br>`identity.account.list`<br>`identity.account.provider-info`<br>`identity.account.refresh-token`<br>`identity.account.unlink`<br>`identity.oauth.callback`<br>`identity.oauth.callback-form-post`<br>`identity.oauth.logout-complete`<br>`identity.oauth.logout-start`<br>`identity.oauth.onetap-callback`<br>`identity.oauth.onetap-start`<br>`identity.oauth.popup-complete`<br>`identity.oauth.provider-apple-client-secret-sign`<br>`identity.oauth.provider-list`<br>`identity.oauth.proxy-forward`<br>`identity.oauth.signin-start`<br>`identity.oauth.signin-token`<br>`journey.social-oauth.v1` | not-yet-needed | none | not-yet-needed:provider-paypal | — | not-yet-needed:provider-paypal |
| provider-polar | `identity/oauth/providers` | `identity.account.access-token`<br>`identity.account.link-start`<br>`identity.account.link-token`<br>`identity.account.list`<br>`identity.account.provider-info`<br>`identity.account.refresh-token`<br>`identity.account.unlink`<br>`identity.oauth.callback`<br>`identity.oauth.callback-form-post`<br>`identity.oauth.logout-complete`<br>`identity.oauth.logout-start`<br>`identity.oauth.onetap-callback`<br>`identity.oauth.onetap-start`<br>`identity.oauth.popup-complete`<br>`identity.oauth.provider-apple-client-secret-sign`<br>`identity.oauth.provider-list`<br>`identity.oauth.proxy-forward`<br>`identity.oauth.signin-start`<br>`identity.oauth.signin-token`<br>`journey.social-oauth.v1` | not-yet-needed | none | not-yet-needed:provider-polar | — | not-yet-needed:provider-polar |
| provider-railway | `identity/oauth/providers` | `identity.account.access-token`<br>`identity.account.link-start`<br>`identity.account.link-token`<br>`identity.account.list`<br>`identity.account.provider-info`<br>`identity.account.refresh-token`<br>`identity.account.unlink`<br>`identity.oauth.callback`<br>`identity.oauth.callback-form-post`<br>`identity.oauth.logout-complete`<br>`identity.oauth.logout-start`<br>`identity.oauth.onetap-callback`<br>`identity.oauth.onetap-start`<br>`identity.oauth.popup-complete`<br>`identity.oauth.provider-apple-client-secret-sign`<br>`identity.oauth.provider-list`<br>`identity.oauth.proxy-forward`<br>`identity.oauth.signin-start`<br>`identity.oauth.signin-token`<br>`journey.social-oauth.v1` | not-yet-needed | none | not-yet-needed:provider-railway | — | not-yet-needed:provider-railway |
| provider-reddit | `identity/oauth/providers` | `identity.account.access-token`<br>`identity.account.link-start`<br>`identity.account.link-token`<br>`identity.account.list`<br>`identity.account.provider-info`<br>`identity.account.refresh-token`<br>`identity.account.unlink`<br>`identity.oauth.callback`<br>`identity.oauth.callback-form-post`<br>`identity.oauth.logout-complete`<br>`identity.oauth.logout-start`<br>`identity.oauth.onetap-callback`<br>`identity.oauth.onetap-start`<br>`identity.oauth.popup-complete`<br>`identity.oauth.provider-apple-client-secret-sign`<br>`identity.oauth.provider-list`<br>`identity.oauth.proxy-forward`<br>`identity.oauth.signin-start`<br>`identity.oauth.signin-token`<br>`journey.social-oauth.v1` | not-yet-needed | none | not-yet-needed:provider-reddit | — | not-yet-needed:provider-reddit |
| provider-roblox | `identity/oauth/providers` | `identity.account.access-token`<br>`identity.account.link-start`<br>`identity.account.link-token`<br>`identity.account.list`<br>`identity.account.provider-info`<br>`identity.account.refresh-token`<br>`identity.account.unlink`<br>`identity.oauth.callback`<br>`identity.oauth.callback-form-post`<br>`identity.oauth.logout-complete`<br>`identity.oauth.logout-start`<br>`identity.oauth.onetap-callback`<br>`identity.oauth.onetap-start`<br>`identity.oauth.popup-complete`<br>`identity.oauth.provider-apple-client-secret-sign`<br>`identity.oauth.provider-list`<br>`identity.oauth.proxy-forward`<br>`identity.oauth.signin-start`<br>`identity.oauth.signin-token`<br>`journey.social-oauth.v1` | not-yet-needed | none | not-yet-needed:provider-roblox | — | not-yet-needed:provider-roblox |
| provider-salesforce | `identity/oauth/providers` | `identity.account.access-token`<br>`identity.account.link-start`<br>`identity.account.link-token`<br>`identity.account.list`<br>`identity.account.provider-info`<br>`identity.account.refresh-token`<br>`identity.account.unlink`<br>`identity.oauth.callback`<br>`identity.oauth.callback-form-post`<br>`identity.oauth.logout-complete`<br>`identity.oauth.logout-start`<br>`identity.oauth.onetap-callback`<br>`identity.oauth.onetap-start`<br>`identity.oauth.popup-complete`<br>`identity.oauth.provider-apple-client-secret-sign`<br>`identity.oauth.provider-list`<br>`identity.oauth.proxy-forward`<br>`identity.oauth.signin-start`<br>`identity.oauth.signin-token`<br>`journey.social-oauth.v1` | not-yet-needed | none | not-yet-needed:provider-salesforce | — | not-yet-needed:provider-salesforce |
| provider-slack | `identity/oauth/providers` | `identity.account.access-token`<br>`identity.account.link-start`<br>`identity.account.link-token`<br>`identity.account.list`<br>`identity.account.provider-info`<br>`identity.account.refresh-token`<br>`identity.account.unlink`<br>`identity.oauth.callback`<br>`identity.oauth.callback-form-post`<br>`identity.oauth.logout-complete`<br>`identity.oauth.logout-start`<br>`identity.oauth.onetap-callback`<br>`identity.oauth.onetap-start`<br>`identity.oauth.popup-complete`<br>`identity.oauth.provider-apple-client-secret-sign`<br>`identity.oauth.provider-list`<br>`identity.oauth.proxy-forward`<br>`identity.oauth.signin-start`<br>`identity.oauth.signin-token`<br>`journey.social-oauth.v1` | not-yet-needed | none | not-yet-needed:provider-slack | — | not-yet-needed:provider-slack |
| provider-spotify | `identity/oauth/providers` | `identity.account.access-token`<br>`identity.account.link-start`<br>`identity.account.link-token`<br>`identity.account.list`<br>`identity.account.provider-info`<br>`identity.account.refresh-token`<br>`identity.account.unlink`<br>`identity.oauth.callback`<br>`identity.oauth.callback-form-post`<br>`identity.oauth.logout-complete`<br>`identity.oauth.logout-start`<br>`identity.oauth.onetap-callback`<br>`identity.oauth.onetap-start`<br>`identity.oauth.popup-complete`<br>`identity.oauth.provider-apple-client-secret-sign`<br>`identity.oauth.provider-list`<br>`identity.oauth.proxy-forward`<br>`identity.oauth.signin-start`<br>`identity.oauth.signin-token`<br>`journey.social-oauth.v1` | not-yet-needed | none | not-yet-needed:provider-spotify | — | not-yet-needed:provider-spotify |
| provider-tiktok | `identity/oauth/providers` | `identity.account.access-token`<br>`identity.account.link-start`<br>`identity.account.link-token`<br>`identity.account.list`<br>`identity.account.provider-info`<br>`identity.account.refresh-token`<br>`identity.account.unlink`<br>`identity.oauth.callback`<br>`identity.oauth.callback-form-post`<br>`identity.oauth.logout-complete`<br>`identity.oauth.logout-start`<br>`identity.oauth.onetap-callback`<br>`identity.oauth.onetap-start`<br>`identity.oauth.popup-complete`<br>`identity.oauth.provider-apple-client-secret-sign`<br>`identity.oauth.provider-list`<br>`identity.oauth.proxy-forward`<br>`identity.oauth.signin-start`<br>`identity.oauth.signin-token`<br>`journey.social-oauth.v1` | not-yet-needed | none | not-yet-needed:provider-tiktok | — | not-yet-needed:provider-tiktok |
| provider-twitch | `identity/oauth/providers` | `identity.account.access-token`<br>`identity.account.link-start`<br>`identity.account.link-token`<br>`identity.account.list`<br>`identity.account.provider-info`<br>`identity.account.refresh-token`<br>`identity.account.unlink`<br>`identity.oauth.callback`<br>`identity.oauth.callback-form-post`<br>`identity.oauth.logout-complete`<br>`identity.oauth.logout-start`<br>`identity.oauth.onetap-callback`<br>`identity.oauth.onetap-start`<br>`identity.oauth.popup-complete`<br>`identity.oauth.provider-apple-client-secret-sign`<br>`identity.oauth.provider-list`<br>`identity.oauth.proxy-forward`<br>`identity.oauth.signin-start`<br>`identity.oauth.signin-token`<br>`journey.social-oauth.v1` | not-yet-needed | none | not-yet-needed:provider-twitch | — | not-yet-needed:provider-twitch |
| provider-twitter | `identity/oauth/providers` | `identity.account.access-token`<br>`identity.account.link-start`<br>`identity.account.link-token`<br>`identity.account.list`<br>`identity.account.provider-info`<br>`identity.account.refresh-token`<br>`identity.account.unlink`<br>`identity.oauth.callback`<br>`identity.oauth.callback-form-post`<br>`identity.oauth.logout-complete`<br>`identity.oauth.logout-start`<br>`identity.oauth.onetap-callback`<br>`identity.oauth.onetap-start`<br>`identity.oauth.popup-complete`<br>`identity.oauth.provider-apple-client-secret-sign`<br>`identity.oauth.provider-list`<br>`identity.oauth.proxy-forward`<br>`identity.oauth.signin-start`<br>`identity.oauth.signin-token`<br>`journey.social-oauth.v1` | not-yet-needed | none | not-yet-needed:provider-twitter | — | not-yet-needed:provider-twitter |
| provider-vercel | `identity/oauth/providers` | `identity.account.access-token`<br>`identity.account.link-start`<br>`identity.account.link-token`<br>`identity.account.list`<br>`identity.account.provider-info`<br>`identity.account.refresh-token`<br>`identity.account.unlink`<br>`identity.oauth.callback`<br>`identity.oauth.callback-form-post`<br>`identity.oauth.logout-complete`<br>`identity.oauth.logout-start`<br>`identity.oauth.onetap-callback`<br>`identity.oauth.onetap-start`<br>`identity.oauth.popup-complete`<br>`identity.oauth.provider-apple-client-secret-sign`<br>`identity.oauth.provider-list`<br>`identity.oauth.proxy-forward`<br>`identity.oauth.signin-start`<br>`identity.oauth.signin-token`<br>`journey.social-oauth.v1` | not-yet-needed | none | not-yet-needed:provider-vercel | — | not-yet-needed:provider-vercel |
| provider-vk | `identity/oauth/providers` | `identity.account.access-token`<br>`identity.account.link-start`<br>`identity.account.link-token`<br>`identity.account.list`<br>`identity.account.provider-info`<br>`identity.account.refresh-token`<br>`identity.account.unlink`<br>`identity.oauth.callback`<br>`identity.oauth.callback-form-post`<br>`identity.oauth.logout-complete`<br>`identity.oauth.logout-start`<br>`identity.oauth.onetap-callback`<br>`identity.oauth.onetap-start`<br>`identity.oauth.popup-complete`<br>`identity.oauth.provider-apple-client-secret-sign`<br>`identity.oauth.provider-list`<br>`identity.oauth.proxy-forward`<br>`identity.oauth.signin-start`<br>`identity.oauth.signin-token`<br>`journey.social-oauth.v1` | not-yet-needed | none | not-yet-needed:provider-vk | — | not-yet-needed:provider-vk |
| provider-wechat | `identity/oauth/providers` | `identity.account.access-token`<br>`identity.account.link-start`<br>`identity.account.link-token`<br>`identity.account.list`<br>`identity.account.provider-info`<br>`identity.account.refresh-token`<br>`identity.account.unlink`<br>`identity.oauth.callback`<br>`identity.oauth.callback-form-post`<br>`identity.oauth.logout-complete`<br>`identity.oauth.logout-start`<br>`identity.oauth.onetap-callback`<br>`identity.oauth.onetap-start`<br>`identity.oauth.popup-complete`<br>`identity.oauth.provider-apple-client-secret-sign`<br>`identity.oauth.provider-list`<br>`identity.oauth.proxy-forward`<br>`identity.oauth.signin-start`<br>`identity.oauth.signin-token`<br>`journey.social-oauth.v1` | not-yet-needed | none | not-yet-needed:provider-wechat | — | not-yet-needed:provider-wechat |
| provider-yandex | `identity/oauth/providers` | `identity.account.access-token`<br>`identity.account.link-start`<br>`identity.account.link-token`<br>`identity.account.list`<br>`identity.account.provider-info`<br>`identity.account.refresh-token`<br>`identity.account.unlink`<br>`identity.oauth.callback`<br>`identity.oauth.callback-form-post`<br>`identity.oauth.logout-complete`<br>`identity.oauth.logout-start`<br>`identity.oauth.onetap-callback`<br>`identity.oauth.onetap-start`<br>`identity.oauth.popup-complete`<br>`identity.oauth.provider-apple-client-secret-sign`<br>`identity.oauth.provider-list`<br>`identity.oauth.proxy-forward`<br>`identity.oauth.signin-start`<br>`identity.oauth.signin-token`<br>`journey.social-oauth.v1` | not-yet-needed | none | not-yet-needed:provider-yandex | — | not-yet-needed:provider-yandex |
| provider-zoom | `identity/oauth/providers` | `identity.account.access-token`<br>`identity.account.link-start`<br>`identity.account.link-token`<br>`identity.account.list`<br>`identity.account.provider-info`<br>`identity.account.refresh-token`<br>`identity.account.unlink`<br>`identity.oauth.callback`<br>`identity.oauth.callback-form-post`<br>`identity.oauth.logout-complete`<br>`identity.oauth.logout-start`<br>`identity.oauth.onetap-callback`<br>`identity.oauth.onetap-start`<br>`identity.oauth.popup-complete`<br>`identity.oauth.provider-apple-client-secret-sign`<br>`identity.oauth.provider-list`<br>`identity.oauth.proxy-forward`<br>`identity.oauth.signin-start`<br>`identity.oauth.signin-token`<br>`journey.social-oauth.v1` | not-yet-needed | none | not-yet-needed:provider-zoom | — | not-yet-needed:provider-zoom |

The coordinator MUST create one row per required external provider or
independent implementation. Consumer units and stable claim IDs MUST be exact,
sorted, and unique. Credential source metadata records presence and a
safe source identifier only. It MUST NOT contain the credential, its hash, a
prefix, or a provider response. A unit with any row here MUST use the lane's
attributable JSON evidence-record path before becoming `verified`; `not-needed`,
`available`, and blockers are not acceptance evidence. The digest MUST bind the
exact committed record bytes. `Evidence record commit` MUST be an existing
commit that contains those exact bytes at the recorded repository path and is
an ancestor of the coordinator commit that marks a consumer verified.
Working-tree-only, untracked, modified-after-commit, or path-exists-only
evidence is invalid. Schema version 2 records MUST contain an RFC3339 time no
earlier than preflight, exact profile, claim and unit attribution, and a passing
unit result with distinct `tested_revision`, `gate_execution_revision`, and
`revalidation_revision` fields, a complete input manifest and canonical input root,
tool/environment identity, and nonempty sanitized artifact SHA-256 hashes. The
revalidation field MUST be null when the tested and gate revisions are equal.
When they differ, it MUST equal the gate execution revision, and validation
MUST recompute the same complete input root from both committed trees before
reusing the original result.
Each external record's artifact hashes MUST bind its separately committed
coordinator execution receipt by path and byte digest. That receipt MUST
contain the same immutable direct-execution, tool-binary, environment,
stdout/stderr, raw-capture, verifier, and final-artifact fields required for a
local gate. The artifact-specific verifier MUST be independent of the consuming
package worker. Protocol and security claims MUST additionally bind the
selected independent suite or implementation, or a separately owned verifier
over raw observations when the protocol manifest selects no external suite.
Producer self-reports, mocks, package-only checks, and unbound provider logs are
not passing external evidence.

The record MUST NOT embed its own containing commit; this table and the ledger bind
that later commit. Every verified unit result MUST match its recorded gate
execution revision and gate fingerprint, while the integration checkpoint
remains a separate worker-merge identity.

## Existing primitive contracts

| Primitive | Consuming units | Registered module/package | API input fingerprint | Gate fingerprint and result | Evidence path |
| --- | --- | --- | --- | --- | --- |
| audit | `identity`<br>`identity/anonymous`<br>`identity/anonymous/postgres`<br>`identity/apikey`<br>`identity/apikey/postgres`<br>`identity/apikey/valkey`<br>`identity/delivery`<br>`identity/delivery/postgres`<br>`identity/email`<br>`identity/http`<br>`identity/i18n`<br>`identity/impersonation`<br>`identity/impersonation/postgres`<br>`identity/magiclink`<br>`identity/mfa`<br>`identity/mfa/postgres`<br>`identity/oauth`<br>`identity/oauth/onetap`<br>`identity/oauth/postgres`<br>`identity/oauth/proxy`<br>`identity/otp`<br>`identity/otp/postgres`<br>`identity/password`<br>`identity/password/postgres`<br>`identity/phone`<br>`identity/postgres`<br>`identity/risk`<br>`identity/risk/captcha`<br>`identity/risk/captcha/captchafox`<br>`identity/risk/captcha/hcaptcha`<br>`identity/risk/hibp`<br>`identity/risk/postgres`<br>`identity/risk/valkey`<br>`identity/session`<br>`identity/session/postgres`<br>`identity/session/valkey`<br>`identity/username`<br>`oauth-server`<br>`oauth-server/device`<br>`oauth-server/oidc`<br>`oauth-server/postgres`<br>`organization`<br>`organization/postgres`<br>`passkey`<br>`passkey/postgres`<br>`scim`<br>`scim/organization`<br>`scim/postgres`<br>`sso`<br>`sso/domain-verification`<br>`sso/oauth2`<br>`sso/oidc`<br>`sso/postgres`<br>`sso/saml`<br>`webauthn`<br>`webauthn/postgres` | pkg/audit | sha256:bd282090f0dd70ebda8e2167e481c8f8d220ab35d230199ef6b92d433f63438f | sha256:a21e63c323903e4a87937eaea83d0a148b644a89a991874eb936c489f233c5af stale | .ai/identity-platform/evidence/preflight/primitive-contracts.json |
| audit/postgres | `identity/reference` | pkg/audit/postgres | sha256:88c70dc4db8402201da65ff7e2cd7b86f5bc58bbfcc149c9d6a4368dd04d3920 | sha256:88876fa0ef6f562770da8c531714972aa7ef22585d5e617d497dcae2c162b50f stale | .ai/identity-platform/evidence/preflight/primitive-contracts.json |
| authentication | `identity`<br>`identity/apikey`<br>`identity/http`<br>`identity/session`<br>`oauth-server`<br>`scim`<br>`sso`<br>`webauthn` | pkg/authentication | sha256:09bc976cc1b154d1da946ee3dee76e5d4e64989752aa4328c6677740c5dc8b2e | sha256:36df7624c200bc4a30f9281d830e3d8723164bd77c2585e1db495b479e3cc54a stale | .ai/identity-platform/evidence/preflight/primitive-contracts.json |
| authentication/jwt | `identity/oauth`<br>`identity/oauth/onetap`<br>`oauth-server/oidc`<br>`sso/oidc` | pkg/authentication/jwt | sha256:3d5bc93fe63f08288b45a716e5d83247f22cae59c0a3e6bca6364a52b4297473 | sha256:bd47ad5a0cc29d679b8a76d686389612b1deff49e51a3af6c6175c630d9b6acc stale | .ai/identity-platform/evidence/preflight/primitive-contracts.json |
| authentication/oidc | `identity/oauth`<br>`identity/oauth/onetap`<br>`identity/oauth/providers`<br>`sso/oidc` | pkg/authentication/oidc | sha256:93c32df4dfd877af5301792612a66449e2c9054ead0a4081aac8461234e7946f | sha256:c70ad7b010e2f2b9a895c1e83c7d38283d614b00d3f20540eae0e792bbed543d stale | .ai/identity-platform/evidence/preflight/primitive-contracts.json |
| authorization | `identity`<br>`identity/apikey`<br>`identity/http`<br>`identity/impersonation`<br>`identity/session`<br>`identity/username`<br>`oauth-server`<br>`organization`<br>`scim`<br>`scim/organization`<br>`sso` | pkg/authorization | sha256:0dd1daf3000bc69b55edbbb909ec2ccbf657fcf103550a54625ec8021b90973b | sha256:249534eabc303bc5a44a9584f4ba299008f37fa4e604fbcbebc36caa5467b50e stale | .ai/identity-platform/evidence/preflight/primitive-contracts.json |
| authorization/postgres | `identity/reference` | pkg/authorization/postgres | sha256:67e42228b4f6f220cead264d5ee4372894ca25d9a6552af87b7d6e0ff6dbb16d | sha256:e8a7819015bb64f95a566f86419c4033d0505d9d6ff94a6c09536df2f78c565a stale | .ai/identity-platform/evidence/preflight/primitive-contracts.json |
| authorization/valkey | `identity/reference` | pkg/authorization/valkey | sha256:f87e4c9a977400fa09ff920ffc2e5d998d3bb494ef6d152b81db37293e443902 | sha256:fe11b663f9078a4fa90a541730af4cbc29a81fb8f78b59cbaafea8d9d2ecc0dd stale | .ai/identity-platform/evidence/preflight/primitive-contracts.json |
| cache | `identity/risk/valkey`<br>`identity/session/valkey` | pkg/cache | sha256:107a99ca8127620196c00a712c317b4852b581066c4d66141a9bcb0107e8fe1e | sha256:906123941baa2c77c93b01bea2285829ad9a26ebbcb8a7afca8004b776acc3d3 stale | .ai/identity-platform/evidence/preflight/primitive-contracts.json |
| capability | `identity`<br>`identity/email`<br>`identity/impersonation`<br>`identity/magiclink`<br>`identity/mfa`<br>`identity/oauth`<br>`identity/oauth/onetap`<br>`identity/oauth/proxy`<br>`identity/otp`<br>`identity/password`<br>`identity/session`<br>`oauth-server`<br>`oauth-server/device`<br>`organization`<br>`sso`<br>`sso/oauth2`<br>`sso/saml`<br>`webauthn` | pkg/capability | sha256:100fc7265785b2f4ce87359dc5befff502cf20dc29e955035807cbe01e6803d2 | sha256:0700ffb01506934f039892a213728216c9654b2528440672a54b97baa02b7c01 stale | .ai/identity-platform/evidence/preflight/primitive-contracts.json |
| capability/postgres | `identity/mfa/postgres`<br>`identity/oauth/postgres`<br>`identity/reference`<br>`oauth-server/postgres`<br>`sso/postgres` | pkg/capability/postgres | sha256:7f7537dbacd91fcedd96607a8fb4440b2af6c41362a5473235cd3d3fd0309c14 | sha256:e8c04327e4b46f488ccaa4b159f00b4f33a8e712c0a49ae708591133ed5526ed stale | .ai/identity-platform/evidence/preflight/primitive-contracts.json |
| capability/valkey | `identity/reference` | pkg/capability/valkey | sha256:a6c97b7f74ca05c37de828989a4dd154e8ba4ad47a23fbf4f0400afd6db4526f | sha256:89f22e3f1f703e473a62764d69fa5f151036514184259c6907520c581542cc76 stale | .ai/identity-platform/evidence/preflight/primitive-contracts.json |
| http-client | `identity/oauth`<br>`identity/oauth/providers`<br>`identity/oauth/proxy`<br>`identity/risk/captcha`<br>`identity/risk/captcha/captchafox`<br>`identity/risk/captcha/hcaptcha`<br>`identity/risk/captcha/recaptcha`<br>`identity/risk/captcha/turnstile`<br>`identity/risk/hibp`<br>`oauth-server`<br>`sso/domain-verification`<br>`sso/oauth2`<br>`sso/oidc`<br>`sso/saml` | pkg/http-client | sha256:5a12aa09aa8e7ee9e9a7ffff4cb7c1c5eef616a52724681e54d9f1d7b3f38ed7 | sha256:5bc302a10c7b4317da828d74646cead5924eceaf036064b4fc26af0879ad7e19 stale | .ai/identity-platform/evidence/preflight/primitive-contracts.json |
| idempotency/postgres | `identity/reference` | pkg/idempotency/postgres | sha256:96a5030074cb750831e55209b6a556097d2472fbd8abd750fd00b5b119b4a8da | sha256:ece300f91ff85bba3091c46f688d48e17d881ed4f4df549c6b0106034be7475a stale | .ai/identity-platform/evidence/preflight/primitive-contracts.json |
| idempotency/valkey | `identity/reference` | pkg/idempotency/valkey | sha256:6cd958a3b8b173daaaa71c59fb5b88cf2c2667eef65eaebea8ea84591ff333ac | sha256:2126130932aa4889deb7062d5ef15504db62f20135f89ca90e4eba498cd2076d stale | .ai/identity-platform/evidence/preflight/primitive-contracts.json |
| identifier | `identity`<br>`identity/anonymous`<br>`identity/email`<br>`identity/i18n`<br>`identity/identitytest`<br>`identity/phone`<br>`identity/risk`<br>`identity/session`<br>`identity/session/valkey`<br>`identity/username`<br>`oauth-server/oidc`<br>`organization`<br>`passkey`<br>`scim`<br>`webauthn` | pkg/identifier | sha256:220a6517f2df023ac27b2120c7335fa4a4344ad7b8a4eda8caef2f8a38ef0d0b | sha256:1e532ad6920a3e0c6fc7375ba167dbe4977f42646478db6727a6054e017b20a5 stale | .ai/identity-platform/evidence/preflight/primitive-contracts.json |
| migrations | `identity/anonymous/postgres`<br>`identity/apikey/postgres`<br>`identity/delivery/postgres`<br>`identity/identitytest`<br>`identity/impersonation/postgres`<br>`identity/mfa/postgres`<br>`identity/oauth/postgres`<br>`identity/otp/postgres`<br>`identity/password/postgres`<br>`identity/postgres`<br>`identity/reference`<br>`identity/risk/postgres`<br>`identity/session/postgres`<br>`oauth-server/postgres`<br>`organization/postgres`<br>`passkey/postgres`<br>`scim/postgres`<br>`sso/postgres`<br>`webauthn/postgres` | pkg/migrations | sha256:5d418cac9d44752d5e160d10dce9bda6e11e1fceff6862cecdc7681976829a43 | sha256:5683585d6df08d5db3815aec19a142aef3b4360fe068e27cb497fc3393a27979 stale | .ai/identity-platform/evidence/preflight/primitive-contracts.json |
| openapi | `identity/http`<br>`scim` | pkg/openapi | sha256:392c7241a9db7ca29fe2e44e33b044bb18b6355165d511712649e5787494fa98 | sha256:d480c7617507fd6efc1d21bc641c92a2364ca83a4a6add68c0f68815661872f1 stale | .ai/identity-platform/evidence/preflight/primitive-contracts.json |
| outbox | `identity/anonymous/postgres`<br>`identity/apikey/postgres`<br>`identity/delivery`<br>`identity/delivery/postgres`<br>`identity/impersonation/postgres`<br>`identity/mfa/postgres`<br>`identity/oauth/postgres`<br>`identity/otp/postgres`<br>`identity/password/postgres`<br>`identity/postgres`<br>`identity/reference`<br>`oauth-server/postgres`<br>`organization/postgres`<br>`passkey/postgres`<br>`scim/postgres`<br>`sso/postgres`<br>`webauthn/postgres` | pkg/outbox | sha256:7a176925cbb3df0e2dfc7f4b431ad8fa054c9be116c27bb2f4758dbcf8887a1f | sha256:9ace8721024ea2b0b1e0b0b604dd2996110d54447bdd450b96a6dd112cdf1a8b stale | .ai/identity-platform/evidence/preflight/primitive-contracts.json |
| outbox/postgres | `identity/reference` | pkg/outbox/postgres | sha256:f863c145687a29b481ed72f9e0518ef37e7eb841340b87424bc03ce0a5cfca46 | sha256:4849d2224c93e51a3f1493d0fe90afd142e8c6dc301ba54acacf5a42d41f00d9 stale | .ai/identity-platform/evidence/preflight/primitive-contracts.json |
| password | `identity/mfa`<br>`identity/otp`<br>`identity/password`<br>`identity/password/postgres`<br>`identity/risk/hibp` | pkg/password | sha256:aa00de2760bdad7d0f73e84e2b2c806c8798bd0adf132e03c801d8d885b6acb5 | sha256:684f98d260ac0548ec3a73da2129d827ebc99dcc6c13948b4549ea57940c640c stale | .ai/identity-platform/evidence/preflight/primitive-contracts.json |
| postgres | `identity/anonymous/postgres`<br>`identity/apikey/postgres`<br>`identity/delivery/postgres`<br>`identity/identitytest`<br>`identity/impersonation/postgres`<br>`identity/mfa/postgres`<br>`identity/oauth/postgres`<br>`identity/otp/postgres`<br>`identity/password/postgres`<br>`identity/postgres`<br>`identity/reference`<br>`identity/risk/postgres`<br>`identity/session/postgres`<br>`oauth-server/postgres`<br>`organization/postgres`<br>`passkey/postgres`<br>`scim/postgres`<br>`sso/postgres`<br>`webauthn/postgres` | pkg/postgres | sha256:a7c86b16cd98395fa1f54f90ff74b367d1a160323b5bc81f59c1f80ea025ecd0 | sha256:907d3d290999771e7ea850de44578409bdebbae68dcea74489ef4f59a17b907b stale | .ai/identity-platform/evidence/preflight/primitive-contracts.json |
| rate-limit | `identity/apikey`<br>`identity/http`<br>`identity/magiclink`<br>`identity/mfa`<br>`identity/oauth/onetap`<br>`identity/oauth/proxy`<br>`identity/otp`<br>`identity/password`<br>`identity/risk`<br>`identity/risk/valkey`<br>`identity/username`<br>`oauth-server`<br>`oauth-server/device`<br>`sso/domain-verification` | pkg/rate-limit | sha256:e0f17895a8956161f02e4069f6eb2dda8e6676d10edd5e4473095783c45b266b | sha256:245ec1448dd77e0c172c8cc20ad632a18fd059b917ac49aba55505378f52e7ee stale | .ai/identity-platform/evidence/preflight/primitive-contracts.json |
| rate-limit/postgres | `identity/reference` | pkg/rate-limit/postgres | sha256:15da1069f325a4ebd53329fe8ce9f99c9e6f2ef03f36c5852ae409b8e73e8e78 | sha256:9a3e5c83cd27b233f2517b8d81d0157602065c3c5909983ee0f4f14afa707737 stale | .ai/identity-platform/evidence/preflight/primitive-contracts.json |
| rate-limit/valkey | `identity/reference` | pkg/rate-limit/valkey | sha256:ac0005f4b95ef06346906a8ad0d029a4b2d98906af7093039d7697f08e8ecfac | sha256:fca2c646a9106a688a380f3694df6af282e2aa98b2d6267e0490662311911abc stale | .ai/identity-platform/evidence/preflight/primitive-contracts.json |
| secret-envelope | `identity/apikey`<br>`identity/delivery/postgres`<br>`identity/mfa`<br>`identity/mfa/postgres`<br>`identity/oauth`<br>`identity/oauth/postgres`<br>`identity/oauth/providers`<br>`identity/oauth/proxy`<br>`identity/reference`<br>`identity/risk/captcha/recaptcha`<br>`identity/risk/captcha/turnstile`<br>`identity/session`<br>`oauth-server`<br>`oauth-server/postgres`<br>`sso`<br>`sso/oauth2`<br>`sso/oidc`<br>`sso/postgres`<br>`sso/saml`<br>`webauthn` | pkg/secret-envelope | sha256:e12450a509776bc0ccbdcd9b67f0f72c21af8598be4d79792fbbb8ab189ca36d | sha256:33ffd2fc8f598d60ad5e6e17d28fbec1a8e3b27663c27e58bc69252f2667c898 stale | .ai/identity-platform/evidence/preflight/primitive-contracts.json |
| telemetry | `identity/apikey/valkey`<br>`identity/delivery`<br>`identity/http`<br>`identity/oauth/providers`<br>`identity/reference`<br>`identity/risk`<br>`identity/risk/captcha`<br>`identity/risk/captcha/captchafox`<br>`identity/risk/captcha/hcaptcha`<br>`identity/risk/captcha/recaptcha`<br>`identity/risk/captcha/turnstile`<br>`identity/risk/hibp`<br>`sso/domain-verification` | pkg/telemetry | sha256:4c896df6527c49d70188d1c6bb35b9de78fa79836d13138513f6f747bcd18191 | sha256:8ec6080f20f6e744f51d3e793c20a2c876e876a4e854d19222bf7f36c038f591 stale | .ai/identity-platform/evidence/preflight/primitive-contracts.json |
| tenancy | `identity`<br>`organization` | pkg/tenancy | sha256:c1dfc1c5b539b37437c364f91ddb60481bfd9c6190c5ae6989991aa1a683d08d | sha256:4192c67bbeac5dd31546336c2330b63c83c41af7bab403692f350b5cbe054f74 stale | .ai/identity-platform/evidence/preflight/primitive-contracts.json |
| workflow | `identity/delivery`<br>`identity/delivery/postgres`<br>`identity/email`<br>`identity/magiclink`<br>`identity/password`<br>`identity/phone`<br>`identity/reference`<br>`sso` | pkg/workflow | sha256:a582c81f4ac39ca83b28fe890dd77313c91df25788c176e039334e8d5399340f | sha256:c3f6d4ab90e4214acb153c4f643b2eb84cb045019c2fbf6cdd8e73353f5000c3 stale | .ai/identity-platform/evidence/preflight/primitive-contracts.json |
| workflow/postgres | `identity/reference` | pkg/workflow/postgres | sha256:ccdc844d648ea8dcec1f2080c1d4e846fa26353bf10ff1d10a71ac589185aa35 | sha256:0effdbf80dffbb2919bec529413ecb7e0731002a582874e1049f3e941b150a49 stale | .ai/identity-platform/evidence/preflight/primitive-contracts.json |

The coordinator MUST derive each primitive's consuming units as the exact,
sorted, unique union of every goal's `Consumes existing primitives` section.
A fingerprint MUST cover the complete behavior-affecting inputs required by
repository evidence policy. A `pass` result permits the normal unit lifecycle.
A `failed`, `blocked`, or `stale` result permits a consumer to remain only
`proposed` with owner `—`, before assignment, or `blocked` as a paused retained
assignment with the exact owner claim
`blocker:primitive-<primitive-with-slashes-replaced-by-hyphens>`. Such a
consumer MUST NOT be `ready`, `in-progress`, `implemented-unverified`, or
`verified`, and MUST NOT receive new work or assignment until the primitive
result is `pass`. Such a unit is explicitly excluded from dependency-frontier
promotion even when every unit in its inventory `Requires` list is `verified`;
primitive start gates are additional to, and not represented by, DAG edges.
The sole migration waypoint is the initial generation-0, unassigned `ready`
row that predates the first durable primitive observation: the pure initial
preflight-population commit MAY record that observation while retaining the
row, but the next coordinator state transition MUST reconcile the exact
consumer to `proposed` before any assignment or other state transition.

## Task-owned resource registry

| Resource ID | Type | Owning unit/task | Exact path or safe external ID | State | Cleanup trigger | Last reconciled at | Cleanup evidence or attestation |
| --- | --- | --- | --- | --- | --- | --- | --- |
| integration-go-library-tools | worktree | coordinator | /Users/brian/Developer/.worktrees/golib-go-library-tools-identity-platform | active | program-complete | 2026-08-30T14:36:39Z | not-yet-needed:integration-go-library-tools |
| integration-go-authentication | worktree | repository-integration/go-authentication | /Users/brian/Developer/.worktrees/golib-go-authentication-identity-platform | active | program-complete | 2026-08-30T14:36:39Z | not-yet-needed:integration-go-authentication |
| integration-go-authorization | worktree | repository-integration/go-authorization | /Users/brian/Developer/.worktrees/golib-go-authorization-identity-platform | active | program-complete | 2026-08-30T14:36:39Z | not-yet-needed:integration-go-authorization |
| integration-go-capability | worktree | repository-integration/go-capability | /Users/brian/Developer/.worktrees/golib-go-capability-identity-platform | active | program-complete | 2026-08-30T14:36:39Z | not-yet-needed:integration-go-capability |
| integration-go-identifier | worktree | repository-integration/go-identifier | /Users/brian/Developer/.worktrees/golib-go-identifier-identity-platform | active | program-complete | 2026-08-30T14:36:39Z | not-yet-needed:integration-go-identifier |
| integration-go-identity | worktree | repository-integration/go-identity | /Users/brian/Developer/.worktrees/golib-go-identity-identity-platform | active | program-complete | 2026-08-30T14:36:39Z | not-yet-needed:integration-go-identity |
| integration-go-oauth-server | worktree | repository-integration/go-oauth-server | /Users/brian/Developer/.worktrees/golib-go-oauth-server-identity-platform | active | program-complete | 2026-08-30T14:36:39Z | not-yet-needed:integration-go-oauth-server |
| integration-go-organization | worktree | repository-integration/go-organization | /Users/brian/Developer/.worktrees/golib-go-organization-identity-platform | active | program-complete | 2026-08-30T14:36:39Z | not-yet-needed:integration-go-organization |
| integration-go-passkey | worktree | repository-integration/go-passkey | /Users/brian/Developer/.worktrees/golib-go-passkey-identity-platform | active | program-complete | 2026-08-30T14:36:39Z | not-yet-needed:integration-go-passkey |
| integration-go-password | worktree | repository-integration/go-password | /Users/brian/Developer/.worktrees/golib-go-password-identity-platform | active | program-complete | 2026-08-30T14:36:39Z | not-yet-needed:integration-go-password |
| integration-go-scim | worktree | repository-integration/go-scim | /Users/brian/Developer/.worktrees/golib-go-scim-identity-platform | active | program-complete | 2026-08-30T14:36:39Z | not-yet-needed:integration-go-scim |
| integration-go-sso | worktree | repository-integration/go-sso | /Users/brian/Developer/.worktrees/golib-go-sso-identity-platform | active | program-complete | 2026-08-30T14:36:39Z | not-yet-needed:integration-go-sso |
| integration-go-webauthn | worktree | repository-integration/go-webauthn | /Users/brian/Developer/.worktrees/golib-go-webauthn-identity-platform | active | program-complete | 2026-08-30T14:36:39Z | not-yet-needed:integration-go-webauthn |
| worker-authentication-g0 | worktree | primitive/authentication-identity-contracts | /Users/brian/Developer/.worktrees/golib-go-authentication-identity-platform-worker-g0 | active | unit-verified-or-safe-abandonment | 2026-10-01T04:27:13Z | active-assignment-0707282e6549b77305ee085c9b6b604eed5434a8 |
| worker-identifier-g0 | worktree | primitive/identifier-identity-contracts | /Users/brian/Developer/.worktrees/golib-go-identifier-identity-platform-worker-g0 | active | unit-verified-or-safe-abandonment | 2026-10-01T05:07:00Z | active-assignment-aedda67bee50ee67d9040778d68756bb88767e90 |
| worker-password-g0 | worktree | primitive/password-secret-contracts | /Users/brian/Developer/.worktrees/golib-go-password-identity-platform-worker-g0 | active | unit-verified-or-safe-abandonment | 2026-10-01T06:16:21Z | active-assignment-57bf66f8125bb684e4fb08e5ee9bcd98ff07d38b |
| worker-capability-g0 | worktree | primitive/capability-identity-contracts | /Users/brian/Developer/.worktrees/golib-go-capability-identity-platform-worker-g0 | active | unit-verified-or-safe-abandonment | 2026-10-01T06:16:21Z | committed-assignment-checkpoint-supplied-on-spawn |
| worker-authorization-g0 | worktree | primitive/authorization-identity-contracts | /Users/brian/Developer/.worktrees/golib-go-authorization-identity-platform-worker-g0 | active | unit-verified-or-safe-abandonment | 2026-10-01T07:07:37Z | evidence/authorization-main-adoption.md; prepared-clean-at-ce41a61000844dd3c06484a7799af336ad1b2892-not-assigned |

Every worktree, disposable cache, temporary directory, process, container,
image, volume, database payload, browser artifact, and provider fixture created
for the program MUST be registered before use. The integration worktree is the
only bootstrap exception: it MUST be registered immediately after creation and
workspace verification, before any operation other than completing and
committing this preflight record. `State` MUST be `active`,
`retained-for-recovery`, `removal-pending-after-final-commit`, or `removed`.
Only the integration worktree MAY use
`removal-pending-after-final-commit`, under the two-phase cleanup in
`ORCHESTRATOR_GOAL.md`; its `Cleanup trigger` MUST be exactly
`user-authorized-post-report-removal`, expressing why it remains and who may
authorize physical removal. `Type` MUST be `worktree`, `go-cache`,
`temporary-directory`, `browser-artifact`, `process`, `container`, `image`,
`volume`, `database-payload`, or `provider-fixture`. Live local paths MUST
exist, removed local paths MUST be absent, and worktrees MUST additionally
match Git registration. Removed local resources MUST name an attributable
repository evidence path. External resources MUST name either an attributable
repository evidence path or the exact sanitized
`attestation:<state>:<safe-external-id>` for their current state. Evidence MUST
be captured into its durable record before the disposable source resource is
removed. Removed entries remain as sanitized audit records. A retained entry
MUST name an exact cleanup trigger; the coordinator MUST reconcile this table
after interruption and before final or blocked reporting. Once every inventory
unit is `verified`, every resource MUST be `removed` except the exact integration
worktree, which MAY remain `removal-pending-after-final-commit` until the final
record commit is created.

## Acceptance evidence bindings

| Artifact ID | Evidence path | Evidence blob digest | Evidence record commit | Bound at |
| --- | --- | --- | --- | --- |

At final acceptance this table MUST contain exactly one row for every
`END_STATE_ACCEPTANCE.json` artifact. The path and artifact ID MUST match the
catalog, the digest MUST bind the exact blob read from the named existing
commit, and that commit MUST descend from the payload's tested revision and be
an ancestor of the binding/finalization commit. The payload MUST NOT contain or
predict its own containing commit. Rows are append-only; replacement evidence
requires a new attributable artifact generation rather than rewriting a prior
binding.

Local package and reverse-dependant gate records use
`identity-platform.local-gate.v2`. Each record MUST contain the exact tested
revision, current gate execution revision, nullable revalidation revision,
exhaustive sorted behavior-input manifest, canonical input root, canonical
record path, commands, outcome, tool/environment identity, artifact hashes, and
record digest. The coordinator captures the clean committed integration `HEAD`
immediately before execution. On reuse, that current HEAD is both gate execution
and revalidation revision while tested revision remains the original execution;
the complete manifest and root MUST validate identically at both revisions.
The later append-only `EXECUTION_LEDGER.md` binding supplies the record commit
and exact canonical-byte digest without self-reference.

Every such record's artifact hashes MUST include a canonical coordinator
execution-receipt path and exact receipt-byte digest. The receipt MUST be
created from a direct coordinator-run subprocess capture and bind the absolute
executable path, executable version, executable-byte SHA-256, exact argument
vector, working directory, sorted redacted environment identity, start and
completion times, exit status, exact stdout/stderr bytes with lengths and
digests, raw-capture path and digest, artifact-specific verifier command and
capture, and final artifact path and digest. The coordinator MUST commit the
receipt and record before adding either ledger binding. A receipt supplied or
finalized by a package worker, producer, or final artifact is invalid.

Affected-package and reverse-dependant manifests MUST come from independent
repository-native discovery at the tested revision and be receipt-bound. For a
coverage or mutation gate, separate discovery receipts MUST capture the exact
gate with `ACCEPTANCE_DISCOVERY=affected-packages` and
`ACCEPTANCE_DISCOVERY=viable-mutants`; the mutation result receipt MUST also
bind the mutation tool's native machine-readable output. Worker-authored or
producer-normalized identity lists MUST NOT replace these captures.

Only `INVENTORY.md`, `EXECUTION_LEDGER.md`, this file, and files under
`.ai/identity-platform/evidence/` are explicitly non-behavioral execution or
evidence bookkeeping excluded from the behavior-input manifest. Their commit
ancestry and evidence bindings remain provenance. The authoritative current
`Requires` closure still selects module roots and `DEPENDENCIES.md` remains a
behavior input. All other selected
identity-platform files remain behavior inputs.

## Goal digest revisions

| Revision ID | Unit | Previous goal digest | Current goal digest | Status | Authorized by | Recorded at |
| --- | --- | --- | --- | --- | --- | --- |

Rows are append-only. Before changing any `GOAL_MANIFEST.json` digest, the
coordinator MUST append an `authorized` row while the previous digest is still
current. That row MUST name the exact matching `user:<safe-approval-id>` from a
prior committed `User semantic authorizations` row; `coordinator` is forbidden
in `Authorized by`. The later manifest-changing commit MUST append the matching
`applied` row with identical revision ID, unit, old/new digests, and user
approval; an abandoned change MUST instead append `superseded`. A terminal row
without its preceding user authorization is invalid. Revision IDs use
`goal:<unit-with-slashes-replaced-by-hyphens>:g<positive-sequence>` and each
digest is the exact `sha256:<hex>` identity of the goal body.
The first execution snapshot MAY omit prior fixtures only while every ledger
row is `initial` and every append-only execution-history table is empty. Every
later execution validation MUST receive the previous committed inventory,
ledger, execution snapshot, and goal manifest together through the four
`--previous-*-fixture` options. This lets the validator prove append-only goal
authorization and recovery lifecycles even when a candidate claims not to
change them.

## Conflict-recovery baselines

| Recovery epoch | Unit | Generation | Integration baseline | Authorization input root | Worker checkpoint | Authorization checkpoint | Conflict evidence path | Result worker commit | Result integration checkpoint | Status | Recorded at |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |

Rows are append-only. `Recovery epoch` MUST be
`recovery:<unit-with-slashes-replaced-by-hyphens>:g<generation>:e<positive-sequence>`;
the sequence starts at one and increases by exactly one for each later recovery
of the same assignment generation. `Status` MUST be `authorized`, `effective`,
`superseded`, or `completed`. The `authorized` row MUST be committed with the
retained assignment's `blocked -> in-progress` transition, name the exact
integration baseline and worker checkpoint, record the repaired unit's
canonical complete-input root at that baseline, use `pending` for Authorization
checkpoint, and use `—` for both result commits. Its commit's first parent MUST
be the integration baseline. An immediate recovery-finalization `effective` row MUST bind
that authorization commit as Authorization checkpoint, copy every other
identity byte, and still use `—` for result commits. Only the committed
effective row authorizes the recovery merge and package edit.

An advanced integration `HEAD` does not by itself supersede an effective row.
Before retry, the coordinator MUST recompute the canonical complete-input
manifest and root for the repaired unit at the recorded baseline and advanced
`HEAD`. Byte-identical roots retain the effective authority; unequal roots
require `superseded` and the next epoch. Path comparison, ancestry, or a claim
that unrelated work advanced is insufficient. Successful integration appends a
`completed` row in a later commit and MUST bind the exact descendant result
worker commit and exact non-fast-forward result integration checkpoint. A
`superseded` row MUST retain `—` result commits. Every `effective`, `superseded`,
or `completed` row MUST be preceded by the matching `authorized` row for the
same epoch and exact commit identity in an earlier committed preflight
snapshot; authorization, effectuation, and a terminal row MUST be separate
commits. A later row cannot create or inherit authorization. Completion is terminal only for that epoch, not for
the assignment generation: if the retained assignment later re-enters
`blocked` because a repaired tip conflicts again, the coordinator MUST append
the next epoch and MAY authorize the new clean checkpoint only when the prior
checkpoint is its ancestor on the exact worker branch. The table contains only
sanitized evidence paths, not conflict contents or secrets.

## Integrated-repair authorizations

| Repair epoch | Unit | Generation | Integration baseline | Integration checkpoint | Worker checkpoint | Goal path | Goal digest | Rendered repair prompt | Prompt digest | Reserved descendants | Result worker commit | Result integration checkpoint | Status | Recorded at |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |

Rows are append-only. A repair epoch MUST be
`repair:<unit-with-slashes-replaced-by-hyphens>:g<generation>:e<positive-sequence>`
and sequence independently within an assignment generation. An `authorized`
row MUST be committed with the exact `implemented-unverified -> in-progress`
transition, and its commit's first parent MUST equal `Integration baseline`.
It MUST bind the already integrated checkpoint, the clean worker-branch
checkpoint, the canonical current goal path, exact rendered repair-prompt bytes
and digest, the exact current goal-body digest, and the complete reserved nested-root set. The worker MUST merge
that exact authorization commit before editing. The coordinator MUST validate
the authorization-checkpoint-to-repaired-tip package-only range, integrate the
repair with a non-fast-forward merge, record the new integration checkpoint,
and append a matching `completed` terminal row in a later commit. An
authorization or `superseded` row MUST use `—` for both result fields. A
`completed` row MUST bind the exact repaired worker commit and resulting
non-fast-forward integration checkpoint, and the same commit MUST record those
values in the unit ledger. A terminal row MUST otherwise have an identical
epoch identity and MUST be preceded by its committed authorization. A changed baseline, checkpoint, goal, prompt, reservation set,
or worker checkpoint requires `superseded` followed by the next epoch; restart
recovery MUST reconstruct the active epoch from this table and Git ancestry.

## Execution-strategy convergence assessment — 2026-10-01

The pre-assignment orchestration repair mechanism is `non_convergent` under
the parent engineering policy's Goodhart and Convergence rules. Repeated
catalog, custody and validator repairs have not established a running package
implementation or any completed identity-platform unit. Implementation remains
0/67, with no assigned package worker. Additional repair of this mechanism is
stopped; existing edits, source pins, assignments and evidence are preserved.

The user subsequently authorized the proposed replacement on 2026-10-01 and
also authorized correcting the secret-type catalog to the existing opaque
security requirements. ORCHESTRATOR_GOAL.md and WORKER_PROMPT.md now define
package-scoped start checks and direct committed assignments. This records
actual conversation direction, not a platform-signed or exact-byte capture.
No package worker has been assigned or activated by this record.

The recommended replacement is package-scoped start checks and explicit
isolated-worker assignments: prove each selected package's source baseline,
goal, ownership and verified dependency closure before its worker starts.
Retain the full 67-unit scope, shared-issuer requirements, substantive
independent engineering roles, package tests and release gates, attributable
committed results, and final cross-package integration acceptance. Do not
represent unrelated catalog failures as package correctness, and do not claim
program completion before the complete end state is verified. The replacement
is now adopted; this assessment is still not a passing package gate or a
program-completion result. Legacy tables are preserved for provenance, not
used to reintroduce the removed pre-start ceremony.
