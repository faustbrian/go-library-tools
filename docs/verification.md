# Verification

The standard contract validates formatting, module tidiness, unsafe imports,
vet, tests, races, exact package coverage, mutation evidence, linting,
vulnerabilities, secrets, licenses, fuzzing, documentation, API compatibility,
conformance, interoperability, and benchmarks when enabled by the manifests.
Formatting and safety scans follow Go package discovery boundaries: they skip
`testdata`, hidden, underscore-prefixed, vendored, and nested module
directories. Repository namespace checks apply to releasable modules; fixture
modules that are not released may use external example identities, but their
manifest and `go.mod` identities must still agree.

Secret scanning combines the repository-owned `.gitleaks.toml` policy with one
tool-owned exclusion for the root `.golib-tooling` checkout created by the
reusable workflow. The merged configuration is task-owned and removed after
the scan. Every other tracked and untracked repository path remains in scope.

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

Coverage is evaluated per production package and must be exactly 100%. The
runner instruments the exact production-package set once while executing the
complete module test set. This keeps denominators production-specific while
allowing integration and external-package tests to contribute only when they
actually exercise a package.
Mutation reports must account for every viable mutant and kill 100%; equivalent
or unreachable cases require narrow reviewed records.

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
