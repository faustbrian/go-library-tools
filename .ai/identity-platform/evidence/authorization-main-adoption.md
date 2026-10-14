# Authorization main-baseline adoption

The user selected **Adopt current main and verify backward compatibility**
in the conversation on 2026-10-01. This is ordinary conversation authority
under ORCHESTRATOR_GOAL.md, not an exact-byte approval ceremony.

- Repository: `github.com/faustbrian/go-authorization`.
- Approved committed main: `ce41a61000844dd3c06484a7799af336ad1b2892`.
- Historical integration: `e08db3af02adde4ad67094d678febc7b12302d85`.
- Merge base: `8eba7c321d847a99f7829837c2f6709f242249f0`.
- New branch: `feature/identity-platform-authorization-g0`.
- New worktree:
  `/Users/brian/Developer/.worktrees/golib-go-authorization-identity-platform-worker-g0`.

The new worktree is clean at the exact approved main revision. It has no
implementation owner or active assignment yet; no package implementation has
started there. The existing integration worktree and all unrelated assignments
are preserved. Historical preflight source rows remain historical evidence,
not claims that the old integration descends from current main.

Native read-only mapper `/root/authorization_contract_map_v1` found additive
adapters and instrumentation APIs, with unchanged policy-engine sources. Its
three-way inspection found tooling conflicts in `.golib.yaml` and
`.github/workflows/ci.yml`, but no production-source conflict. That report is
source evidence, not runtime backward-compatibility verification.

The coordinator independently recomputed both source pins through the existing
`IdentityPlatformPublicContractRepository::Catalog#contract` implementation:

- Historical API: `5b35ec9e45205d530b211e8b7e50426466468613ee632195bf76ee17c698c5b8`.
- Approved API: `73b296651509c8f8314a42c696c6adddc8548f1166e46e09bf895395683b2019`.

The canonical projection includes root production Go files, the owning
`go.mod`, and committed API baseline files with their historical logical labels;
it is not an independently invented API extraction algorithm. The goal's
current-API pin now reflects the approved revision. Its required additive
contract and all backward-compatibility requirements remain unchanged.

Before assignment, reconcile the source-derived current API baseline against
the approved main, retain the additive primitive requirements, resolve stale
generated contract declarations, and obtain the bounded QA input against this
source revision. Before verification, independently prove existing consumers
and behavior remain compatible and resolve integration/tooling conflicts
without weakening mandatory checks. Full-program scope remains 67 units.
