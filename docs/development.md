# Development

Use the Go version in `.go-version`. Every Go command must receive disposable
`GOCACHE`, `GOMODCACHE`, and `GOTMPDIR` directories and remove them afterward.

Run focused package tests and static analysis while editing with
`make check-packages PACKAGES='./internal/example ./internal/other'`.
`make local-check` validates repository-owned configuration, inventory,
cohesion, and workflow invariants; `make local-ci` adds the complete Go test and
vet, formatting, tidy, safety, lint, local-link, and applicable API checks used
by ordinary pull requests. `make milestone-check` runs only the additional
consumer and compatibility checks after that immutable revision's local check;
`make ci` composes both locally. `make check` retains the complete all-enabled repository contract
for releases or changes whose identified risks require it. Expensive mutation,
race, fuzz, benchmark, and external-service checks are selected for those risks
rather than imposed on every milestone.

Docker fixture behavior is unit-tested behind process boundaries; use
designated CI or an explicit manual environment for real service rehearsals.

Keep CLI, configuration, workflow, evidence, and fixture contracts backward
compatible within a major release. Update tests, documentation, schema, and
changelog together.
