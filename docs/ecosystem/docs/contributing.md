# Contribute to one Golib library

[Documentation index](index.md) · [Ecosystem entry](../README.md)

A normal library contribution needs the owning standalone repository, not a
Golib monorepo, sibling checkouts, ecosystem ledger or new Tools release.
Choose the owner in the [consumer catalog](../catalog-consumer.md) or
[engineering catalog](../catalog-engineering.md), then read that repository's
`CONTRIBUTING.md`, `AGENTS.md`, `go.mod`, Makefile and CI at the revision you
intend to change. Nested adapters are separate modules within their owner's
repository; test them from their declared module directories.

## Pick a bounded change

Report a bug with a minimal reproduction, expected versus actual outcome,
module/import path and exact version, Go/platform versions and backend/version
where relevant. Remove credentials, private payloads and production identifiers.
Use the owner's support guide for usage questions and its `SECURITY.md` private
reporting route for vulnerabilities; do not publish exploit details in an
ordinary issue.

Before editing, identify the observable contract, affected callers/modules,
compatibility boundary, named risk and smallest useful checks. Preserve
unrelated work. Stateless functions, validated construction, explicit I/O and
caller-owned lifecycle follow [design-language 1.0](../design-language.md); adapters
normally live at `<owner>/adapters/<target>`. Intentional domain-specific API
shapes remain valid. Propose an API/ownership decision explicitly rather than
normalizing it accidentally in documentation or a convenience wrapper.

The published [compatibility cohort](../compatibility-sets.md) is frozen adoption
guidance, not the latest source of every package. Contribute against the owning
repository's supported branch/version, retain required `/vN` module/import
suffixes and account for its existing consumers. Do not rewrite the cohort,
source lock or historical receipts to make a new revision appear included.

## Reproduce and verify the affected contract

For a simple module, `go test ./...` runs its deterministic tests and saved fuzz
corpus; it does not exercise separately nested modules or every service/tagged
integration. Use the module's pinned Go 1.27 toolchain and real native targets.
This bounded session shares only task-owned caches and removes them on exit:

```sh
(
  set -eu
  golib_task_root=$(mktemp -d "${TMPDIR:-/tmp}/golib-contributor.XXXXXX")
  golib_task_root=$(cd "$golib_task_root" && pwd -P)
  trap 'chmod -R u+w "$golib_task_root" 2>/dev/null || true; find "$golib_task_root" -depth -delete' EXIT
  mkdir -p "$golib_task_root/cache" "$golib_task_root/mod" "$golib_task_root/tmp"
  export GOCACHE="$golib_task_root/cache" GOMODCACHE="$golib_task_root/mod"
  export GOTMPDIR="$golib_task_root/tmp" GOWORK=off
  go test ./...
  go vet ./...
)
```

A behavioral fix demonstrates the failure before the fix and the corrected
outcome afterward. Test authorization, failure, cancellation, settlement and
cleanup where they are actually affected, not only a success path or mocked
call sequence. Public APIs need affected direct-consumer/API proof; persistence
needs its real database boundary; material concurrency/lifetime changes need
focused race/leak proof. Changed hostile inputs add bounded active fuzzing using
the owner's actual target. An incomplete or timed-out check is not a pass.

Run applicable formatting, static/security and changed-content secret checks.
Read the repository's native `CONTRIBUTING.md` and Makefile before selecting
`make check` or `make ci`: their meanings and costs differ between libraries.
Mutation is optional local or selected scheduled assurance, never a required PR
campaign. Do not broaden ordinary work into an unrelated ecosystem, benchmark,
release rehearsal or all-backend campaign. A performance/conformance claim needs
its actual reproducible workload or pinned official fixture/independent oracle.

When contributing to **go-library-tools itself**, its verified native entry
points are:

```sh
make check-packages PACKAGES=./internal/inventory
make local-check
make source-docs-check
```

The package selection above is for inventory changes, not every Tools change.
Those Makefile targets manage their own disposable Go resources. Tools'
process-control, descendant-cleanup and Docker-impacting tests stay on their
specified hosted jobs; its full `make local-ci` is not a safe universal local
substitute. Follow the versioned [Tools contribution guide](https://github.com/faustbrian/go-library-tools/blob/12d2a3d5425b20f9482e26eef60c957569145cbe/CONTRIBUTING.md)
and [development guide](https://github.com/faustbrian/go-library-tools/blob/12d2a3d5425b20f9482e26eef60c957569145cbe/docs/development.md).

## Submit useful review context

Keep the diff and explicit staged paths coherent. Describe the resulting
behavior and why it matters, affected API/imports, meaningful regression
coverage and any adoption or rollback action. Update affected docs and the
owner's changelog for material consumer changes; prose-only changes do not
force a module release. Maintain backward compatibility or propose a separately
versioned major boundary with its migration path.

Follow the owner's contribution/commit/PR rules and inspect the entire final
diff. Meaningful public-API, lifecycle, security, persistence, concurrency or
release work gets one independent complete-final-diff review. Fix concrete
findings and rerun affected checks; clerical edits do not restart the project.
No routine signed receipts, separate role factories or recursive approvals are
needed to submit a library fix.

Use immutable pinned identities for Actions and external tools, retain
least privilege and untrusted-PR/release separation, and review new dependencies
for maintenance, license, vulnerabilities and provenance. Do not add mutable
pins, arbitrary execution hooks, permanent local replacements, machine paths,
bypass flags or exclusions that hide failures. Submission is not permission to
publish a release, merge a PR or rewrite history.

## Maintainer and release boundary

Work and tags remain on main unless a branch is explicitly requested; major
versions use Go's import/module suffix and Git tags, not version-specific source
directories. Maintainers use the owner's protected-main delivery process and
required exact-source CI. A committed/pushed change is not CI-qualified or
released. A required public release additionally needs its published tags and
an actual clean public consumer with no workspace/local replacement. See the
[release guide](../../release.md) and
[versioning guide](versioning.md).

Specification-backed behavior also follows the owner's versioned decision
register. Reuse unchanged immutable proof when it still applies, but never turn
partial, stale or failed evidence into release readiness. Required aggregation
must reject applicable failed, cancelled, missing or skipped work.

## Next steps

Find an owner in [choosing packages](choosing-packages.md), inspect its
[package detail](../catalog-consumer.md), and use the [security contract](../security/README.md)
for shared vulnerability and trust boundaries. For deployment implications,
continue to the [operator route](operations/kubernetes.md).
