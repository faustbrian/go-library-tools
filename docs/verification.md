# Verification

Verification is proportional to the change. Ordinary development uses focused
package tests and applicable static analysis, then the fast repository-local
checks. Pull requests run `make local-ci`, including formatting, module
tidiness, safety, tests, vet, configured lint/static analysis, local Markdown
links, and applicable API compatibility. Main and scheduled runs do not
automatically select an aggregate milestone. `make milestone-check` supplies
consumer/compatibility validation only when that boundary is selected; shared
ordinary quality uses the selected affected modules in one local session.

Each verifier invocation owns private Go caches, scratch and GolangCI-Lint's
result cache, overriding an ambient lint-cache directory. Adjacent commands in
that invocation reuse the cache; cleanup removes it on success or failure.
The executor disables external Go/lint cache backends and nested session opt-ins
after command overrides so they cannot escape that ownership boundary.
Opting into a bounded local cache session borrows Go compilation/module caches
only for explicitly classified native repository commands. Pinned verifiers keep
separate invocation-owned caches; unknown scripts cannot borrow the session and
are not given verifier cache identities. This is cache isolation, not a sandbox
for arbitrary same-user code. Lint results remain invocation-owned and are never
written to the user-wide cache or retained across independent invocations.

Reusable CodeQL PR analysis compiles the same prepared changed owners and owned
reverse consumers with their canonical build tags and a 30-minute job bound.
Invalid, empty, duplicated or unknown selections fail before compilation; build
failures remain failures. The manual builds ignore ambient Go workspaces and use
one disposable compilation session. Main, scheduled and explicit assurance
analysis still compiles the complete inventory: partial PR analysis does not
replace repository-wide security coverage. Query suites and mandatory-result
aggregation are unchanged. This routing reduces unrelated PR compilation, not
the security requirements of the selected source and its consumers.

Tools' own CI applies the same documentation/runtime distinction on both PRs
and main pushes. Prose, license and notice changes use `make metadata-ci` with
`GOLIB_BASE_SHA` set to the immutable comparison commit: repository/security
structure, local documentation and changed-content secrets still fail closed.
Only regular, non-executable prose is metadata; executable or symlinked Markdown,
unknown legal-document suffixes and names such as `LICENSE.go` require runtime
checks in both the Tools and reusable workflows.
Unknown, source, workflow and policy changes retain `make local-ci` and CodeQL;
scheduled/manual analysis retains full runtime coverage. An empty diff is not
metadata, and missing comparison evidence fails rather than hiding applicable
work. The metadata CodeQL job validates applicability but does not compile or
query unchanged runtime source; Required still rejects a failed or cancelled job.

Mutation, broad fuzzing, race, benchmark, external-service, release-rehearsal,
and ecosystem-wide checks do not block a routine change unless they directly
exercise its material risk. `make check-packages PACKAGES=...` is the focused
developer entry point, `make local-check` validates repository-owned metadata,
and `make check` remains the explicit complete all-enabled repository gate for
release work or a risk-selected aggregate run.
Cold mutation-verifier construction uses the hosted `verifierintegration`
suite, outside ordinary package tests. Tools CI selects its finite-execution
checks when mutation code, dependencies or the toolchain change, and for
scheduled/manual qualification without a comparison baseline. Selection does
not run a mutation campaign; the selected integration failures remain fatal.
Changed-input fuzz smoke permits two minutes per target for discovery, cold
instrumented compilation and execution, within a three-minute total smoke
deadline. It still runs a two-second active campaign with a 15-second test
timeout and two workers. Earlier caller deadlines take precedence; exhausted
budgets, missing or ambiguous targets and native failures remain failures.
This allowance avoids rejecting a healthy first fuzz run solely because ordinary
test caches do not contain instrumented dependencies. It does not extend the
active campaign or select unrelated targets.

Tools' own GolangCI-Lint configuration retains Staticcheck's unique `QF*`
quick-fix rules plus `SA9003` and `ST1023`, which the standalone CLI excludes
by default. The local and full verifier gates still run standalone Staticcheck
for its current correctness, simplification, style and unused-code checks.
The combined stack avoids repeating the overlapping analysis without dropping
either tool's unique rules. Other linters and the broader configured
govet rules remain unchanged. `golangci-lint run` alone is not the combined
stack: use `make local-ci` for normal verification. This repository-local
configuration must not be copied to a pipeline without standalone Staticcheck
or without inspecting its own rule settings and exclusions.
Formatting and safety scans follow Go package discovery boundaries: they skip
`testdata`, hidden, underscore-prefixed, vendored, and nested module
directories. Repository namespace checks apply to releasable modules; fixture
modules that are not released may use external example identities, but their
manifest and `go.mod` identities must still agree.

Secret scanning uses a centrally generated policy and does not load the
repository-owned `.gitleaks.toml`. The generated configuration is task-owned
and removed after the scan. Its narrow conjunction allowlists cover only
centrally justified tool identities and exact synthetic security-test fixtures;
all other tracked and untracked repository paths remain in scope.
Security-enabled checks reject a repository-owned `.gitleaksignore`, and both
scan modes ignore inline `gitleaks:allow` directives.

Security-record validation registers its embedded schemas under an absolute
in-memory resource identity. Validating an absolute document directory does
not require an unrelated ambient working directory to remain available;
schema registration and compilation do not fetch that logical URI. Published
schemas, built-in draft resolution, document admission and cooperative
cancellation retain their existing behavior.

The documentation gate requires a regular root README, bounds document count
and size, rejects trailing whitespace and symlinks, and verifies local Markdown
targets without making network requests. Link-shaped text inside fenced,
indented, or inline code is treated as code rather than navigation. It then
runs CSpell `10.0.0` from an
embedded lockfile in the task-owned workspace. Consumer repositories retain
only their word policy in `cspell.json`, not shared npm manifests or lockfiles.
The immutable ecosystem catalog Markdown mirrors are excluded from this
repository-level spelling pass because their bytes must remain identical to the
attested release assets. Source documentation remains spell-checked by its
owning repository; catalog publication independently verifies the generated
mirror bytes.
External links are checked by checksum-pinned Lychee `0.24.2` binaries for
Linux and macOS on amd64 and arm64. Archives are bounded and treated as
untrusted input; the verifier reads only the expected executable after checking
the complete archive digest and entry structure.
Nested modules do not repeat repository-wide Markdown checks. Their enabled
documentation gate runs the configured typed operation or a bounded Go example
test when no module-specific operation is declared.

When selected for an aggregate milestone or release, coverage is evaluated per
production package and defaults to exact 100% enforcement. The
runner instruments the exact production-package set once while executing the
complete module test set. This keeps denominators production-specific while
allowing integration and external-package tests to contribute only when they
actually exercise a package.

An explicit repository policy may instead collect statement evidence for a
named module:

```yaml
coverage:
  modules:
    - module: .
      mode: evidence
```

The optional `coverage.modules` entries accept only `exact` or `evidence`, must
name distinct canonical module directories, and require the module's coverage
gate to be enabled. Omission keeps exact enforcement, including for other
modules in the same repository. Evidence mode runs the identical instrumented
module tests and reports attributed covered/total counts; every expected package
must have executable statements and at least one executed statement. Missing,
malformed, inconsistent, unreadable, or overflowing profiles still fail closed.
Evidence collection does not certify adequate behavioral-risk coverage or
release readiness. Those decisions require the repository's acceptance contract,
meaningful behavior tests, review, and all other applicable gates. Adopt the
policy deliberately with tooling that supports it; older strict decoders reject
the new optional field. Neither mode skips collection or changes other gates.

Exact-coverage failures retain the first sorted failing-package error and report all
incomplete expected packages from the same profile. Source-block filenames and
ranges share a global limit of 512 records; package labels and locations are
limited to 160 escaped ASCII bytes, and the complete failure report to 85,000
bytes. Unsafe, oversized, and excess locations or package records are counted
as omitted. Single-package reports retain their existing shape. Profiles remain
task-owned and are removed on success and failure.

When selected for an aggregate milestone or release, mutation reports must
account for every viable mutant: each must be killed or covered by a narrow
reviewed equivalence classification. The optional equivalent inventory accepts only
the exact source-, verifier-, and coordinate-bound `LIVED` classifications it
names. Native statuses and counters are retained, not rewritten into kills;
survivors lacking review, uncovered, timed-out, not-viable, missing, or stale
results still fail. A reviewed equivalence may be limited to a stated public
contract domain rather than byte-for-byte internal behavior. Failed campaigns
emit bounded coordinate diagnostics without publishing passing evidence.

For optional local or owner-selected scheduled work, `golib mutation --module
DIR --package DIR` selects one catalogued production package and one worker.
Its 20-minute deadline includes service setup, input discovery, tool preparation,
module baselines and campaign execution; service cleanup keeps its independent
30-second bound. Existing content-identical package checkpoints and reviewed
equivalence classifications remain reusable, including after a broader run.
No whole-module legacy import is performed. Unknown, disabled or excluded
packages fail rather than yielding a successful not-applicable result. Survivor,
infrastructure and budget failures remain failures; this partial scope does not
certify unselected packages or release readiness. The native backend still uses
module baselines and integration tests; selecting a package bounds the campaign
scope and wall time but does not make every mutant a cheap unit-only run.
Admitted failed-report projections persist in the mutation evidence directory's
`failed-reports` sibling, outside disposable campaign workspaces. They retain
native coordinates, statuses and numeric aggregate field presence, but omit
module, elapsed-time and statistics metadata. Filenames must identify a direct
production Go source in the selected package; mutation types and statuses must
match the pinned producer. Projection content identity does not identify raw
native bytes or certify complete accounting: native totals can omit timed-out
mutants. Failed projections never satisfy evidence reuse or a passing gate.
Missing or refused reports emit only a static capture diagnostic and preserve
the original campaign failure. Cancellation and task cleanup remain unchanged.

Evidence is keyed by complete behavior-affecting content and verifier identity,
not Git history. It is persisted atomically when available. Code, tests,
fixtures, relevant configuration, tools, or service identity invalidate only
the affected evidence. Mutation fixture walks ignore nested symlink entries
without following their targets; fixture roots and explicitly listed source
files must remain real repository-contained paths. Mutation package identity
includes only sibling modules that the package actually observes. Approved v1
checkpoint inputs transition to this narrower v2 identity from the same source
listing without rerunning a campaign. A stale package checkpoint does not
discard approved sibling checkpoints; that package alone proceeds through
current mutation verification. Missing or ambiguous verifier identity fails
closed.
