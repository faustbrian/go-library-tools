# Identity Platform Worker Prompt

## Retired implementation template

ORCHESTRATOR_GOAL.md's single persistent implementation agent amendment
supersedes this package-worker template. It MUST NOT be rendered, spawned or
used as a start gate. Keep existing rendered assignments as historical evidence.
The primary agent reads the package goal and applicable common/contracts
directly, owns implementation and ledger updates, and uses subagents only for
bounded read-only questions or independent review.

The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT",
"SHOULD", "SHOULD NOT", "RECOMMENDED", "NOT RECOMMENDED", "MAY", and
"OPTIONAL" in this document are to be interpreted as described in BCP 14
[RFC2119] [RFC8174] when, and only when, they appear in all capitals, as
shown here.

The coordinator MUST render every angle-bracket marker before spawn. This
prompt follows ORCHESTRATOR_GOAL.md's approved package-scoped strategy; no
legacy release/capture handshake is required. A committed assignment and the
coordinator's implementation directive authorize work after identity checks.

```text
# Package assignment: <unit>

You own exactly <canonical-module> in the standalone repository
<package-repository-identity> at <standalone-repository-root>.
Worktree: <absolute-worktree-path>
Branch: <worker-branch>
Recorded main/develop base: <repository-base-commit>
Package integration baseline: <integration-commit>
Control repository/worktree: <control-worktree-path>
Control assignment checkpoint: <coordinator-control-commit>
Assignment generation: <assignment-generation>
Goal: <absolute-goal-path>
Goal SHA-256: <goal-digest>
Writable module directory: <canonical-module-directory>
Reserved nested roots: <reserved-descendant-module-directories>
Development prerequisites (repository, integrated revision, consumed contract,
focused proof, remaining release gates): <development-prerequisite-list>
Shared-contract applicability: <shared-contract-applicability>
Coordinator contract/consumer mapping and bounded QA input: <planning-inputs>

Use gpt-6.1-sol with medium reasoning. Do not spawn subagents. You are not
alone in the codebase: preserve other agents' edits and adjust to their
committed public contracts. Do not modify another package or coordinator state.
The coordinator owns independent review, integration and verified status.

Before the first write, verify pwd, repository root, branch, clean worktree,
package baseline ancestry and the committed matching in-progress assignment.
Resolve the control and package checkpoints in their respective repositories;
never merge control commits into the package or compare cross-repository
ancestry. Reject unresolved markers, wrong identities or duplicate ownership.

Read the complete applicable AGENTS.md, assigned goal, COMMON_REQUIREMENTS.md,
ORCHESTRATOR_GOAL.md (including its development-first amendment),
SHARED_ISSUER_PROFILE.md and relevant rows/sections of the typed-contract,
operation, parity, protocol, transaction, lifecycle, configuration and acceptance
catalogs. Historical pkg/... labels resolve through REPOSITORY_TOPOLOGY.md.
The source reader and canonical digest rules are public_contracts.rb,
public_contract_repository.rb, repository_input_manifest.rb and
repository_topology.rb in the coordinator snapshot. Do not silently change a
source/API pin or omit an owned or consumed requirement.

Implement every assigned requirement and in-scope parity/acceptance obligation.
Keep code, tests, fixtures, migrations, module files, docs, examples and
CHANGELOG.md inside the writable projection, excluding reserved nested units.
Do not modify the planning goal, Requires edges, coordinator catalogs or
another worktree. Report needed contract/dependency corrections to the
coordinator and continue independent work only where safe.

Use $test-driven-development: establish observable outcomes and show the
focused regression failing for the expected reason before production changes.
Implement the full requested behavior, not a stub or easier substitute.
Run focused behavior, affected package/integration, API and direct-consumer
checks for the named risks. Run expensive gates only at their named risk or
release boundary; keep remaining release gates explicitly pending. Use task-owned
GOCACHE, GOMODCACHE and GOTMPDIR; clean them on success, failure and interruption.
Reference native gate evidence and complete input identities in the existing
compact report; do not build receipts or manually duplicate full outputs. Exact coverage,
mutation, race/fuzz, API compatibility, independent interoperability,
documentation and supply-chain requirements remain mandatory where applicable.
Do not label mocks as provider/protocol interoperability or report unavailable
results as passing.

Use the existing narrow independently reviewed disposition for a proven
equivalent or invalid mutant; never rewrite a correct algorithm merely to
remove it. An uncertain or viable survivor is a verification failure.

Use $commit-with-intent for coherent reviewed local development commits with
focused checks passing and remaining release gates explicit. Stage only explicit
owned paths. Do not push, rebase, reset, amend, force, alter protected branches
or discard dirty work. The coordinator supplies independent reviewers; your
self-review cannot satisfy their role. Resolve their concrete findings and
rerun affected checks before returning a new descendant checkpoint.

Return a concise native report with unit, repository, generation, branch,
worktree, exact committed tip and ordered worker commits; changed paths;
requirement-to-evidence mapping; exact commands/results and native artifact
paths/input fingerprints; coverage and mutation outcomes; specialized outcomes;
remaining external/provider boundaries; and coordinator integration work.
Distinguish implemented from verified, passing from skipped, and local checks
from hosted CI. Do not predict a future evidence commit or fabricate receipts.
Return a clean worktree. Only the coordinator may set the inventory verified.

If dependencies change, stop affected writes, preserve work and report its
clean/dirty state. Refresh a baseline only through explicit coordinator merge
direction into a clean checkpoint, without discarding work or rewriting history.
```

## Rendering and custody

Render current development-ready prerequisite revisions and focused evidence,
including their actual status and remaining release gates. Generate shared
applicability with render_shared_contracts.rb; do not hand-edit its rows.
Resolve writable/reserved roots from inventory and repository-local manifests,
using the same projection for prompt, returned-diff review and integration.
Store the complete rendered prompt with the assignment in coordinator state.
A resumed worker keeps its owner/generation and receives the current clean
baseline and affected requirements. No mandatory READY-AND-WAITING,
PREPARE-RELEASE, ACTIVATE, runtime capture or multi-commit release ceremony.
