# Package assignment: primitive/password-secret-contracts

You own exactly pkg/password in standalone repository
github.com/faustbrian/go-password at /Users/brian/Developer/golib/go-password.
Worktree: /Users/brian/Developer/.worktrees/golib-go-password-identity-platform-worker-g0
Branch: feature/identity-platform-password-g0
Recorded main/develop base: 082f25b85d7975049b0629e1b57226384469af3b
Package integration baseline: 082f25b85d7975049b0629e1b57226384469af3b
Control repository/worktree: /Users/brian/Developer/.worktrees/golib-go-library-tools-identity-platform
Control requirements baseline: 81f631b56678310da17bee970de0ea112b925d8d
Control assignment checkpoint: the commit containing this assignment; the
coordinator supplies its exact SHA in the spawn directive after it exists.
Assignment generation: 0
Goal: /Users/brian/Developer/.worktrees/golib-go-library-tools-identity-platform/.ai/identity-platform/goals/primitive-password-secret-contracts.md
Goal SHA-256: c76c360e0ccf57a0475547448dddcda4d5edaad57f843eac4b36f6d55d0ea4d9
Writable module directory: . in the assigned password worktree
Reserved nested roots: none (only one go.mod exists)
Verified prerequisites: none; Requires: None. Start API digest is
sha256:12edda3d3097088256d5657037779ca2a3d9ef5cb4c6eab82515e65ea1e91f95.
Required contract digest:
sha256:225402cb5159d79d5acb6debacab7ae4666ed5567432cbbd6a7fb6f884d4df91.
Independent mapping and bounded QA input: native read-only reports
/root/password_contract_map_v1 and /root/password_contract_qa_v1, retained in
the coordinator conversation. The QA artifact digest is
7304986f2b98b4b3eabe7a484aac230b8151aed8e139bfe1212c067afc757ce8.
Inputs are summarized below, not implementation certification.

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
Run focused checks and all applicable package-required gates with task-owned
GOCACHE, GOMODCACHE and GOTMPDIR; clean them on success, failure and interruption.
Preserve exact native outputs and complete input identities. Exact coverage,
mutation, race/fuzz, API compatibility, independent interoperability,
documentation and supply-chain requirements remain mandatory where applicable.
Do not label mocks as provider/protocol interoperability or report unavailable
results as passing.

Use $commit-with-intent for coherent verified local commits. Stage only explicit
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

## Exact shared-contract applicability

### transaction
- `none`

### lifecycle
- `none`

### lifecycle consumers
- `none`

### configuration
- `none`

### security events
- `none`

## Exact verification applicability

- `race=required`
- `fuzz=required`
- `hostile=required`
- `leak=required`
- `benchmark=not_applicable` — primitive/password-secret-contracts owns no benchmark boundary; composed consumers own that evidence
- `infrastructure=not_applicable` — primitive/password-secret-contracts owns no infrastructure boundary; composed consumers own that evidence
- `provider_interoperability=not_applicable` — primitive/password-secret-contracts owns no provider interoperability boundary; composed consumers own that evidence

## Independent implementation inputs

The recorded v1 line is authorized. Current main is a materially different v2;
do not merge it, change module identity, or drop historical APIs. New Parameters
arm closure must not retroactively tighten accepted legacy NewPolicy shapes.
Existing Service, adapters, encodings, admission, cancellation, dummy work and
CAS behavior remain unchanged. HIBP acceptance/provider policy belongs to its
downstream owners; Secret must support their exact borrowed-byte composition.

One bounded native QA pass found no blocker. Its nine proving slices are:
1. Closed parameter arms; intrinsic boundaries, including memory 8*p at p=255;
   use an independent predicate and never hash hostile constructor maxima.
2. Separate caller policy limits, encoded-length budget, above-default valid
   parameters and tiny real Argon2id/bcrypt hash/verify composition.
3. Secret 0/1/max-1/max/max+1 byte bounds, input copying and binary fidelity.
4. Borrowed alias cleanup, irreversible Destroy, zero/copy/container ownership,
   callback mutation and sequential-use semantics; no resurrection.
5. Nil callback, error/panic cleanup/redaction, bounded reentrant Len/Destroy/Use
   without holding locks across callbacks; preserve useful classified errors.
6. Joined barrier-driven race/stress/leak tests across copied owners; at most
   one unsafe borrower, no exposure after completed destruction.
7. Value/pointer/copy fmt, JSON/text and log/slog non-disclosure, including
   decimal-byte, hexadecimal and base64 representations of synthetic material.
8. External old/new consumers plus unchanged service/adapter/cancellation,
   monotonic upgrades, CAS and independent pinned PHP fixture interoperability.
9. Bounded constructor fuzz with independent oracles, no giant crypto work,
   input retention or secret diagnostics; wire targets into native gate selection.

Document safe exact zero/copy/reuse/panic/reentrancy/concurrency behavior as
implementation decisions, not new approval dependencies. Do not promise
preventing malicious callback copies or runtime-wide memory erasure.
Preserve the old baseline for actual old-to-new compatibility before updating
the additive API artifact. Hosted publication/provenance remains unverified
without its actual release evidence; never substitute local logs for it.

Request parent fresh regression/correctness/Go review at the full coherent
tested checkpoint before a local implementation commit. Report the first
focused red and first production write promptly, then implement the full goal.
