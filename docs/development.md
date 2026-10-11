# Development

Use the Go version in `.go-version`. Every Go command must receive disposable
`GOCACHE`, `GOMODCACHE`, and `GOTMPDIR` directories and remove them afterward.

Run focused package tests and static analysis while editing with
`make check-packages PACKAGES='./internal/example ./internal/other'`.
Its test and vet commands share one disposable cache session. `make local-check`
and `make local-ci` each build the CLI once and reuse it for adjacent checks;
their build-session directories are removed on success or failure. Each CLI
invocation retains its existing executor-owned isolation. No shared global
cache or persistent local binary is introduced.

`make local-check` validates repository-owned configuration, inventory,
cohesion, and workflow invariants; `make local-ci` adds the complete Go test and
vet, formatting, tidy, safety, lint, local-link, and applicable API checks used
by ordinary pull requests. `make milestone-check` runs only the additional
consumer and compatibility checks after that immutable revision's local check;
`make ci` composes both locally. `make check` retains the complete all-enabled repository contract
for releases or changes whose identified risks require it. Expensive mutation,
race, fuzz, benchmark, and external-service checks are selected for those risks
rather than imposed on every milestone.

Ordinary local checks retain security and license validation but skip throwaway
SBOM generation. Full generator validation and retained release SBOM artifacts
remain at their explicit boundaries.

Release validation requires the ordinary tests, documentation, lint, and
security gates. Coverage, race, and mutation remain available when the changed
behavior's risk classification selects them; disabling one of those gates for
a module does not by itself prevent an otherwise verified release.

Docker fixture behavior is unit-tested behind process boundaries; use
designated CI or an explicit manual environment for real service rehearsals.

The reusable library workflow requires a verifier with `check --local`. A
legacy verifier without that capability fails with an adoption diagnostic;
it never falls back to `check --all` and a mutation campaign. Keep existing
workflow pins until adopting a qualified capable release or explicitly select
`source_bootstrap: true` with an immutable reviewed Tools commit. Bootstrap
does not change the assurance tier.

Documentation-only and exactly paired reusable-workflow/tooling SHA changes
use the metadata route on both pull requests and main pushes. Source removal,
renames, unmatched pins, and substantive workflow changes retain runtime
checks. Missing comparison commits fail classification rather than produce
green. Release rehearsals must be explicitly dispatched, not selected by a PR.

Keep CLI, configuration, workflow, evidence, and fixture contracts backward
compatible within a major release. Update tests, documentation, schema, and
changelog together.
