# Package assignment: primitive/capability-identity-contracts

You own exactly pkg/capability in standalone repository
github.com/faustbrian/go-capability at /Users/brian/Developer/golib/go-capability.
Worktree: /Users/brian/Developer/.worktrees/golib-go-capability-identity-platform-worker-g0
Branch: feature/identity-platform-capability-g0
Recorded main/develop base: b0b21d15c9653decb0ff4a4fbffbe71c77d8d8fc
Package integration baseline: b0b21d15c9653decb0ff4a4fbffbe71c77d8d8fc
Control repository/worktree: /Users/brian/Developer/.worktrees/golib-go-library-tools-identity-platform
Control requirements baseline: 18432e4151c7786ce1812c838fe9f00fc5ecac9d
Control assignment checkpoint: the commit containing this assignment; the
coordinator supplies its exact SHA in the spawn directive after it exists.
Assignment generation: 0
Goal: /Users/brian/Developer/.worktrees/golib-go-library-tools-identity-platform/.ai/identity-platform/goals/primitive-capability-identity-contracts.md
Goal SHA-256: 0c38f1416ef88d1020f2775db80f3bbce01f17c7350ae66cf09b24051b1988cc
Writable module directory: . in the assigned capability worktree
Reserved roots: postgres/ and valkey/; do not modify adapter source or tests.
Verified prerequisites: none; Requires: None. Current API digest is
sha256:3ad716bb7799ba0d0045b7456d2dabfe486503b9050b4d984dead16d69da7e14.
Required contract digest:
sha256:fa22b30449283df23ee2e4dca6f60531a54dd49b02d866aa2e63c55db9dde2af.
Independent mapping and bounded QA input: native read-only reports
/root/capability_contract_map_v1 and /root/capability_contract_qa_v1, retained
in the coordinator conversation. QA requirements-closure commitment:
644b2850954e6431ab6c2cf40d1c3de48ff4ca0684fcbabfaa964a7075c3d696.
The original QA input pin preceded the timestamp metadata reconciliation;
that correction preserves runtime requirements, not a new planning pass.

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
Keep code, tests, fixtures, module files, docs, examples and CHANGELOG.md inside
the writable projection, excluding reserved units. Do not modify the planning
goal, Requires edges, coordinator catalogs or another worktree. Report needed
contract/dependency corrections to the coordinator and continue independent
work only where safe. Root manifests/tool scripts may change only when necessary
to register or verify this core extension without weakening existing gates.

Use $test-driven-development: establish observable outcomes and show the
focused regression failing for the expected reason before production changes.
Implement the full requested behavior, not a stub or easier substitute.
Run focused checks and all applicable package-required gates with task-owned
GOCACHE, GOMODCACHE and GOTMPDIR; clean them on success, failure and interruption.
Preserve exact native outputs and complete input identities. Exact coverage,
mutation, race/fuzz, API compatibility, independent interoperability,
documentation and supply-chain requirements remain mandatory where applicable.
Do not label mocks as provider/protocol interoperability or report unavailable
results as passing. The pinned target toolchain is Go 1.26.6.

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
- `benchmark=not_applicable` — primitive/capability-identity-contracts owns no benchmark boundary; composed consumers own that evidence
- `infrastructure=required`
- `provider_interoperability=not_applicable` — primitive/capability-identity-contracts owns no provider interoperability boundary; composed consumers own that evidence

## Independent implementation inputs

Preserve the existing caveat map/wire encoding, timestamp skew (not-before may
precede issuance), expiry after both timestamps and lifetime from their minimum.
Do not tighten existing Issue, Parse or Verify. PostgreSQL enlisted-transaction
contribution belongs to the ordered adapter unit; legacy Grant.Consume remains.

Before the first production write, supply the smallest package-owned concrete
Service construction proposal using the existing signer/resolver/clock/limits
boundaries. The interface alone cannot prove Issue/Verify behavior. Supply exact
purpose grammar and Reference digest domain framing/rejection policy with
independent vectors. These are implementation decisions within the existing
goal, not a new approval ceremony; coordinator reconciliation precedes verified.

The single bounded native QA pass supplies these proving slices:
CAP-Q01 old/new external APIs and construction; CAP-Q02 purpose/reference bounds;
CAP-Q03 configured token limits and zero ingress; CAP-Q04 structural-only parsing;
CAP-Q05 alias invalidation and borrow cleanup; CAP-Q06 pre-cancel/invalid arguments;
CAP-Q07 redacted error/panic cleanup without retry; CAP-Q08 barrier-driven single
transfer and verification snapshot ordering; CAP-Q09 actual Service signature,
purpose/audience, time and revocation checks; CAP-Q10 no implicit consumption and
unchanged explicit replay/unknown outcomes; CAP-Q11 pinned wire/caveat/timestamp
and HTTP compatibility; CAP-Q12 value/pointer/container fmt, JSON/text and logging
non-disclosure; CAP-Q13 bounded deterministic/fuzz ingress; CAP-Q14 compilable
custody/nonconsumption examples and native scoped gates.

Independent oracles include canonical checked-in bytes, Python golden tokens,
official crypto vectors, fixed clocks, independent malformed inputs, retained
borrow observations and real memory-store ordinals. Include complete allocated
capacity in cleanup guarantees; passing race or mutation totals alone does not
prove erasure. Never promise prevention of malicious separately allocated copies
or runtime-wide memory erasure.

Preserve historical API artifacts before additive regeneration for actual
old-to-new comparison. Hosted publication and downstream composition remain
unverified without actual evidence. Request parent fresh regression/correctness/
Go reviews at the coherent tested checkpoint before local implementation commit.
Report the first expected red and first production write promptly.
