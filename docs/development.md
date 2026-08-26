# Development

Use the Go version in `.go-version`. Every Go command must receive disposable
`GOCACHE`, `GOMODCACHE`, and `GOTMPDIR` directories and remove them afterward.

Run focused package tests while editing. `make check` runs the repository
contract and `make ci` runs the complete local CI-equivalent contract. Docker
fixture behavior is unit-tested behind process boundaries; use designated CI
or an explicit manual environment for real service rehearsals.

Keep CLI, configuration, workflow, evidence, and fixture contracts backward
compatible within a major release. Update tests, documentation, schema, and
changelog together.
