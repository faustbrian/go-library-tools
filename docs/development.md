# Development

Use the Go version in `.go-version`. Every Go command must receive disposable
`GOCACHE`, `GOMODCACHE`, and `GOTMPDIR` directories and remove them afterward.

Run focused package tests and static analysis while editing with
`make check-packages PACKAGES='./internal/inventory'` (select the actual affected
packages, not this example indiscriminately).
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

`golib docs check --local` validates repository Markdown structure and local
links with the existing bounded checker, without runtime commands, network
access, or spelling-tool installation. Metadata CI uses it before secret
scanning. Existing `golib docs check` retains full configured documentation
checks; executable examples need their relevant compile/tests.

Ordinary local checks retain security and license validation but skip throwaway
SBOM generation. Full generator validation and retained release SBOM artifacts
remain at their explicit boundaries.

Release validation requires the ordinary tests, documentation, lint, and
security gates. Coverage, race, and mutation remain available when the changed
behavior's risk classification selects them; disabling one of those gates for
a module does not by itself prevent an otherwise verified release.

Tools' complete test set includes process-control and descendant-cleanup cases.
Agents must not run those or Docker-impacting rehearsals locally; select safe
package/test cases and use the designated hosted jobs for those boundaries.
This restriction is specific to Tools' test effects, not an instruction to
replace every ordinary library test run with hosted CI.

## Contribute to one standalone library

Work in the owning repository/module with its declared Go toolchain. A normal
simple module's starting point is `go test ./...`: it includes deterministic
fuzz corpus regressions, not active fuzzing or mutation. No sibling checkout,
ecosystem ledger, agent protocol, or new Tools release is required to submit a
normal source fix. Follow any actual adapter/service requirements documented by
that repository, and test affected nested modules separately.

For an isolated bounded local session, these commands share private caches and
remove them afterward:

```bash
(
  set -eu
  golib_task_root=$(mktemp -d "${TMPDIR:-/tmp}/golib-contributor.XXXXXX")
  trap 'chmod -R u+w "$golib_task_root" 2>/dev/null || true; find "$golib_task_root" -depth -delete' EXIT
  mkdir -p "$golib_task_root/cache" "$golib_task_root/mod" "$golib_task_root/tmp"
  export GOCACHE="$golib_task_root/cache" GOMODCACHE="$golib_task_root/mod"
  export GOTMPDIR="$golib_task_root/tmp"
  go test ./...
  go vet ./...
)
```

For changed parser/untrusted-input behavior, replace or add the selected test
command inside that session with a short active fuzz smoke. In this Tools
repository the bounded manifest target, for example, is:

```bash
go test ./internal/inventory -run '^$' -fuzz '^FuzzManifest$' -fuzztime=10s -parallel=2
```

Choose the owning library's actual package and target; this example does not
exercise another parser. Save any minimized failing input as a deterministic
regression and fix it; do not call a short smoke exhaustive fuzz assurance.
Use focused `go test -race` and actual integration/direct-consumer checks when
concurrency, persistence, public API or external-boundary changes require them.

When the matching `golib` tool is already available, explicit local commands are:

```bash
golib check --local --module .
golib mutation --module .
golib release check --module .
golib release dry-run --module .
```

These are different boundaries, not a required four-command checklist. The first
runs ordinary configured checks. Mutation is an optional developer-local or
selected scheduled campaign and can be expensive; it is not routine PR work.
The last two apply only to a release decision/rehearsal and retain full selected
release obligations. Use a nested module's declared directory instead of `.`
where applicable. These `golib` commands manage their own disposable executors;
source builds still need the private Go session above. Tools' system-risk tests
remain hosted even when a command would otherwise select them.

Do not edit unrelated manifests, published compatibility sets or SBOM records
solely because source changed. Update facts genuinely affected by the change;
generate retained release artifacts at their actual publication boundary.

The reusable library workflow requires a verifier with `check --local` and
changed-input `--base-sha` fuzz smoke for PR/main checks. A
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
