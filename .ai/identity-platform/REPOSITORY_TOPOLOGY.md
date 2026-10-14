# Identity Platform Standalone Repository Map

The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT",
"SHOULD", "SHOULD NOT", "RECOMMENDED", "NOT RECOMMENDED", "MAY", and
"OPTIONAL" in this document are to be interpreted as described in BCP 14
[RFC2119] [RFC8174] when, and only when, they appear in all capitals, as
shown here.

This file translates the immutable logical `pkg/...` labels used by the
identity-platform planning artifacts into current standalone repository
locations and module paths. It is an execution map, not a product-scope change.

## Mapping Algorithm

For a logical module `pkg/<family>/<remainder>`:

1. Resolve `<family>` using the table below.
2. Use the listed repository as the only Git repository and worktree root.
3. Use `<remainder>` as the repository-relative nested-module directory. An
   empty remainder means the repository root.
4. Use the listed module prefix plus `/<remainder>` as the canonical Go module
   path.
5. Keep the assigned goal in
   `/Users/brian/Developer/golib/.ai/identity-platform/goals`; do not copy or
   move it into a package repository.

| Logical family | Standalone repository | Module prefix |
| --- | --- | --- |
| `authentication` | `/Users/brian/Developer/golib/go-authentication` | `github.com/faustbrian/go-authentication` |
| `authorization` | `/Users/brian/Developer/golib/go-authorization` | `github.com/faustbrian/go-authorization` |
| `capability` | `/Users/brian/Developer/golib/go-capability` | `github.com/faustbrian/go-capability` |
| `identifier` | `/Users/brian/Developer/golib/go-identifier` | `github.com/faustbrian/go-identifier` |
| `password` | `/Users/brian/Developer/golib/go-password` | `github.com/faustbrian/go-password` |
| `identity` | `/Users/brian/Developer/golib/go-identity` | `github.com/faustbrian/go-identity` |
| `webauthn` | `/Users/brian/Developer/golib/go-webauthn` | `github.com/faustbrian/go-webauthn` |
| `passkey` | `/Users/brian/Developer/golib/go-passkey` | `github.com/faustbrian/go-passkey` |
| `organization` | `/Users/brian/Developer/golib/go-organization` | `github.com/faustbrian/go-organization` |
| `sso` | `/Users/brian/Developer/golib/go-sso` | `github.com/faustbrian/go-sso` |
| `scim` | `/Users/brian/Developer/golib/go-scim` | `github.com/faustbrian/go-scim` |
| `oauth-server` | `/Users/brian/Developer/golib/go-oauth-server` | `github.com/faustbrian/go-oauth-server` |

## Repository Lifecycle

### Optional capability identity composition

The existing logical capability PostgreSQL unit also owns the additive
`go-capability/postgres/identity` module, with public module path
`github.com/faustbrian/go-capability/postgres/identity` and tag prefix
`postgres/identity/v`. This is an optional composition adapter, not a new product
unit. The existing `go-capability/postgres` package and ConsumptionStore remain
in the root module with their legacy API and schema unchanged. Do not move or
replace them to satisfy a literal historical nested-module label.

The optional module depends on identity/postgres through public sealed handles;
the root capability module does not depend on this optional module. Importing
identity/postgres into the root module would create a cycle because identity
already consumes capability. The optional pattern preserves independent releases
and the existing 67-unit scope. Full participant lifecycle and compatibility
catalog reconciliation remain obligations of the same capability PostgreSQL unit.

Existing repositories MUST be reused. During program preflight, the coordinator
MAY create every missing mapped standalone repository needed by the closed
repository-identity table, no later than that family's first schedulable unit.
A new local repository MUST contain exactly one non-behavioral `main` root
commit limited to `AGENTS.md`, `CLAUDE.md`, `LICENSE`, `.gitignore`, and
`.gitattributes` before its integration worktree is created. It MUST NOT
fetch, push, claim that its configured canonical origin exists remotely, create
a parent Git repository, create a shared `pkg/` tree, or create one repository
per nested adapter.

Each repository owns its branches, worktrees, commits, CI, changelog, module
catalog, tags, and releases. Cross-repository prerequisites MUST be consumed
through released versions or explicit task-owned local integration wiring that
is removed before release. Permanent sibling `replace` directives are
forbidden.

## Coordinator Control Repository

The versioned execution copy of this orchestration program and its mutable
ledger MUST live in
`/Users/brian/Developer/golib/go-library-tools/.ai/identity-platform` on main,
or its selected main worktree. The copy in the parent `.ai/` directory is the
launch source and coordination index only; the parent directory MUST NOT be
initialized as a repository to make its state committable.

Use the existing control execution snapshot and preserve its history. Do not
copy/reset the program or rerun global orchestration validators on resumption.
Planning/state commits belong to the control repository; production commits
belong to each mapped repository. The primary agent is the sole writer.
Repository-qualified source pins and input identity bind cross-repository
consumption; shared Git ancestry and worker assignments are not required.

Pending family work MUST be finished and reviewed in its existing worktree,
then delivered to that repository's main branch without discarding history.
Subsequent implementation runs on main. New isolation follows workspace policy,
not an automatic one-worktree-per-unit ceremony. Independent modules retain
their ownership/dependency boundaries even when a consumer milestone spans them.
Historical orchestration scripts/schemas are not implementation start gates.
Never create a parent monorepo to satisfy legacy ancestry assumptions.

## Legacy Label Rules

The phrases `canonical module`, `canonical package directory`, `canonical goal
path`, and `pkg/...` in pinned goal bodies MUST be interpreted through this
map. They MUST NOT be interpreted literally as paths beneath
`/Users/brian/Developer/golib`.

The historical destination fields in `GOAL_MANIFEST.json` remain digest and
provenance metadata. They MUST NOT cause goal relocation or filesystem
creation. The planning goal path is the authoritative current goal path.
