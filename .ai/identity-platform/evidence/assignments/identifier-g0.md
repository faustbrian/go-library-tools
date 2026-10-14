# Package assignment: primitive/identifier-identity-contracts

You own exactly pkg/identifier in standalone repository
github.com/faustbrian/go-identifier at /Users/brian/Developer/golib/go-identifier.
Worktree: /Users/brian/Developer/.worktrees/golib-go-identifier-identity-platform-worker-g0
Branch: feature/identity-platform-identifier-g0
Recorded main/develop base: 23a8818faef585a87e3e31965d6af51bd7900478
Package integration baseline: 23a8818faef585a87e3e31965d6af51bd7900478
Control repository/worktree: /Users/brian/Developer/.worktrees/golib-go-library-tools-identity-platform
Control requirements baseline: f96595933386f62447b8b1dafdb8da9315151a2c
Control assignment checkpoint: the commit containing this assignment; the
coordinator supplies its exact SHA in the spawn directive after it exists.
Assignment generation: 0
Goal: /Users/brian/Developer/.worktrees/golib-go-library-tools-identity-platform/.ai/identity-platform/goals/primitive-identifier-identity-contracts.md
Goal SHA-256: 22fe87ce22b4101e88f220f7731f4d4f1ce8362976c93214fea95e75c4274d93
Writable module directory: . in the assigned identifier worktree
Reserved nested roots: none (only one go.mod exists in this repository)
Verified prerequisites: none; Requires: None. Start API digest is
sha256:99bd12b8c660306f4d4ab99ffcc9d4f8df171e09f01c62cfc65fde004bcb536c.
Independent mapping and bounded QA input: native change_mapper report
/root/identifier_contract_map and qa_test_planner report
/root/identifier_contract_qa, retained in the coordinator agent conversation.
Their implementation inputs are summarized below, not substituted for native
reports or presented as implementation certification.

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
- `leak=not_applicable` — primitive/identifier-identity-contracts owns no leak boundary; composed consumers own that evidence
- `benchmark=not_applicable` — primitive/identifier-identity-contracts owns no benchmark boundary; composed consumers own that evidence
- `infrastructure=not_applicable` — primitive/identifier-identity-contracts owns no infrastructure boundary; composed consumers own that evidence
- `provider_interoperability=not_applicable` — primitive/identifier-identity-contracts owns no provider interoperability boundary; composed consumers own that evidence

## Independent implementation inputs

Mapping: preserve existing generic ID and UUID, ULID, TypeID, KSUID, NanoID,
and slug variants. Do not move email/username policy, directory imports,
storage collisions or auth authority into this primitive. Future catalog
consumers use canonical profiles/records, References and raw bounded Input;
no existing runtime consumer requires redesign. Validate direct defined-string
Input casts and opaque zeros at every processing boundary. Transformation
order, canonical ASCII grammar, selected Unicode/IDNA versions and zero-value
Record/Equal behavior are documented, tested implementation decisions within
this existing goal, not a new approval program.

Native single-pass QA artifact: identifier-identity-contracts-qa-single-pass,
principal /root/identifier_contract_qa; no blocking findings. Implement:

- ID-Q01: exact UTF-8/byte bounds, no premature normalization, cast/zero bypass.
- ID-Q02: closed profile fields/modes, positive version, immutable owned Spec.
- ID-Q03: distinguishing None/NFC/NFKC, full casefold, trim and selected IDNA
  vectors, interactions/order and fixed points; pinned official corpora and
  separately maintained independent oracle for specification claims.
- ID-Q04: final-output byte bounds after expansion/contraction; empty trim.
- ID-Q05: equality binds ID/version/bytes; real JSON complete record roundtrip;
  reject absent/mismatched profile and forged non-fixed-point records.
- ID-Q06: exact reference kind/display bounds, canonical validity, identity
  distinct from display and from authorization.
- ID-Q07: synthetic sentinels absent from owned errors/fmt/slog/nested reference
  diagnostics; Record remains the explicit sensitive complete wire projection.
- ID-Q08: external positive and negative compiler probes for public construction
  versus forged private-field/nonzero opaque values; zero validation paths.
- ID-Q09: visible deterministic within-profile collisions; separate identity
  across versions and mixed-version restore without silent migration.
- ID-Q10: race/stress immutable reads and caller Spec/Record mutation isolation.
- ID-Q11: bounded hostile Unicode/IDNA/input/restore fuzz and golden oracles.
- ID-Q12: unchanged historical generic/concrete-family/codec/SQL/API behavior,
  clean consumers, exact coverage, mutation, docs/examples and affected gates.

Recorded historical baseline lacks api/baseline.txt; the committed goal explains
the source-artifact reconciliation. Use genuine historical API extraction or
the proven byte-equivalent c42902d6478497518cb48406bdd8a4122912f497 baseline
artifact for additive compatibility, not a newly generated baseline as proof
that incompatible changes never happened. Current local main may be a newer
major line; do not merge it or change module identity/pins.

Request parent independent regression/correctness/Go review at the full coherent
tested checkpoint before a local implementation commit. Report first focused
red and first production write promptly, then keep implementing the full goal.
