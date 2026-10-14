# Package assignment: primitive/authentication-identity-contracts

You own exactly pkg/authentication in the standalone repository
github.com/faustbrian/go-authentication at
/Users/brian/Developer/golib/go-authentication.
Worktree: /Users/brian/Developer/.worktrees/golib-go-authentication-identity-platform-worker-g0
Branch: feature/identity-platform-authentication-g0
Recorded main/develop base: 9926ce00a89c4f3187b12b4b25e3268158e138e6
Package integration baseline: 9926ce00a89c4f3187b12b4b25e3268158e138e6
Control repository/worktree: /Users/brian/Developer/.worktrees/golib-go-library-tools-identity-platform
Control requirements baseline: 38d1517aa142155e157987ec25332da570e9aa1d
Control assignment checkpoint: the commit containing this assignment; its exact
SHA is supplied in the coordinator's spawn directive after the commit exists.
Assignment generation: 0
Goal: /Users/brian/Developer/.worktrees/golib-go-library-tools-identity-platform/.ai/identity-platform/goals/primitive-authentication-identity-contracts.md
Goal SHA-256: aec6a4d2dfbe75e501f5f83b0e4a12651152c72e1daf9b7f4662f3f2913fa151
Writable module directory: . in the assigned authentication worktree
Reserved nested roots: authotel/, jwt/, oidc/ (independently released modules)
Verified prerequisites: none; the goal has Requires: None. The source-derived
current API matches sha256:57ba45a7bb9441a3d936c5073e3e59460d214e68d8aeb3ded8f32bc329b385d0.
Independent mapping and bounded QA input: the native change_mapper
/root/authentication_contract_map and qa_test_planner
/root/authentication_contract_qa reports, retained in the coordinator's agent
conversation. Their bounded implementation inputs are summarized below; this
summary does not replace the native reports or certify implementation.

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
- `leak=not_applicable` — primitive/authentication-identity-contracts owns no leak boundary; composed consumers own that evidence
- `benchmark=not_applicable` — primitive/authentication-identity-contracts owns no benchmark boundary; composed consumers own that evidence
- `infrastructure=not_applicable` — primitive/authentication-identity-contracts owns no infrastructure boundary; composed consumers own that evidence
- `provider_interoperability=not_applicable` — primitive/authentication-identity-contracts owns no provider interoperability boundary; composed consumers own that evidence

## Independent implementation inputs

Mapping: retain existing credential-to-principal validation and all current
consumers. The missing trusted authority construction is implementation work,
not an inconsistency: arbitrary public Composite bindings cannot themselves
confer minting authority. Add the genuinely package-owned trusted composition
required by the goal, with its private capability. Shared-issuer validity types
are authorized implementation/catalog reconciliation, not a second approval
program. Report any exact additions needed to the coordinator before using them.

QA: the original defined-string-versus-opaque contradiction is corrected in
the committed catalog and independently regression-reviewed. Runtime opacity
is still unproven. Implement the bounded proving set:

1. External clean-consumer compilation, non-constructibility and non-disclosure
   for opaque codes and authority proofs.
2. Exact constructor boundaries, closed enums/unions and bounded hostile fuzz.
3. Proof identity/tenant/subject/purpose/audience/issuance/expiry bindings,
   including the 24-hour lifetime cap and every mismatch/error classification.
4. Reauthentication exact session/version/action and explicit recovery allowance.
5. Reject external Authenticator/issuer implementations lacking private minting
   authority; the factory must not authenticate. Prove valid trusted issuance.
6. Independent signed access-token fixtures for lineage and separate owner,
   actor, tenant and caller; deny ID-token substitution, wrong environment,
   missing roots, cycles and excessive delegation depth where this unit owns
   resource-validation behavior.
7. Freshness starts at authoritative read start, expires within 20 seconds,
   uses the earliest contributing expiry, cannot be renewed by cache copying,
   and fails closed on stale evidence/outage. Lost events and warm caches must
   still satisfy the selected 30-second revocation guarantee.
8. Existing API/credential/principal compatibility and race-safe validation.

The parent supplies fresh independent regression/correctness/Go-compatibility
roles after implementation. Before a local implementation commit, request
those reviews at a coherent checkpoint; do not self-attest mandatory review.
No legacy global pre-start validator, exact-byte approval capture or release
handshake is required. All product/security/package acceptance remains required.
