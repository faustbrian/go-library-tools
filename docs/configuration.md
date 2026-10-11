# Configuration

`.golib.yaml` contains repository-specific policy and an exact released tool
version. Its schema is published at [`schema/golib.schema.json`](../schema/golib.schema.json).
Unknown fields, future schema versions, duplicate operations, unsafe paths,
and arbitrary executables are rejected.

```yaml
schema_version: 1
tool_version: v1.0.1
tool_checksums_sha256: 0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef
manifest:
  modules: modules.json
  packages: packages.json
evidence:
  root: .verification
mutation:
  root: .verification/mutation
  imports:
    - module: .
      archive: .verification/mutation/bootstrap/root.zip
      ledger: .verification/mutation/migration-ledger.json
api:
  baselines:
    - module: .
      mode: apidiff
      path: api/v1.export
runtimes:
  deno: 2.9.4
  zsh: "5.9"
```

Defaults match the example paths. `modules.json` owns module identity, Go
version, lifecycle, gates, tags, services, and release metadata.
`packages.json` owns package classification. Configuration references those
facts rather than duplicating them.

With capable tooling, a module may omit its `packages` field: `packages.json`
is the single classification owner, and the verifier expands its records into
the module's inventory and generated catalogs. No filesystem inference or new
classification defaults are involved: raw records supplying omitted mirrors
must satisfy the existing required package definition before expansion. This
optional form works with module schemas 1–3 and requires no schema migration.
Existing explicit mirrors
remain supported and must match every canonical record exactly; `[]` and `null`
do not mean omission. Unknown owners and duplicate or conflicting records remain
errors. JSON inventory/catalog output still includes complete package arrays,
and manifest snapshots retain the original input bytes. Older published tools
require the mirrors, so remove them only after adopting a qualified release
with this capability; unchanged repositories need no migration or release.

Each `goal_evidence` entry in `modules.json` is a structured record with the
goal file, requirements digest, implementation artifacts, verification gates,
and implementation status. Unknown record fields are rejected. This metadata
documents goal traceability but does not select or waive verification gates.

Module manifests may use schema v1, v2, or v3. Schema v3 is optional and adds a
fixed schema identity while limiting its module and cohesion fields to facts
owned by the repository. Goal status, review state, provenance chains, and
delivery receipts remain outside the repository manifest. Existing v1 and v2
manifests continue to work without migration.

Consumer repositories pin both `tool_version` and the lowercase SHA-256 digest
of that release's `checksums.txt` asset. The setup action verifies this digest
before trusting the archive checksum selected for the current platform. The
tooling repository may omit the checksum only while bootstrapping and testing
the release that will produce it; released consumers must include it.

`api.baselines` assigns at most one API policy to each module. `apidiff` mode
tracks incompatible exported API changes against a repository-relative
baseline. `go-doc` mode compares an exact normalized documentation snapshot
for an explicit package list and is intended for repositories whose existing
contract includes exported documentation. The API update command writes either
format atomically. Checksum-only and Git-history scripts migrate to explicit
baseline files; they do not remain executable hooks.

For a named concurrency, lifecycle or integration risk, declare a separate
`local-test` operation rather than adding the complete stress campaign to PRs:

```yaml
operations:
  - module: .
    gate: local-test
    steps:
      - type: go-test
        packages: [.]  # One actual package, never ./...
        run: '^TestCancellationReleasesPermit$'  # Use your owned test's name.
        race: true
        timeout: 90s
```

`golib check --local` and full checks both run these declared tests. Each step
requires one explicit non-recursive package, an anchored top-level `Test`
selector, one execution, and an explicit deadline. Discovery and execution share
that deadline; a selector matching no tests fails. Put alternatives after the
mandatory `Test` prefix, for example `^Test(Commit|Rollback)$`; anchors must cover
the complete expression. At most eight steps are
allowed per module, and local checks admit at most three minutes of combined
step deadlines across selected modules. Shorten or split a genuinely broader
campaign into scheduled assurance; do not silently omit its required regression.
Existing module test tags and services still apply. Make, fuzz, benchmark and
stress-repeat steps are not admitted on this route. Native command failures,
timeouts and cancellation remain failures, not evidence of a passing risk check.
Execution uses the existing owned process-group cancellation and caps each output
stream at four MiB; overflowing output fails the check rather than growing logs
without bound. Package selectors may not contain any Go `...` wildcard.

The optional `race` field adds Go's race build to typed test steps; omission or
false preserves existing behavior. It is not accepted for Make, fuzz or benchmark
steps. Existing `test` operations and full race/stress campaigns remain unchanged
and explicit. Older tooling does not understand these additions; adopt them only
with capable qualified tooling. Repositories without them need no migration.

`runtimes` declares exact non-Go executables used by repository tests. Deno is
installed at the declared semantic version. The current zsh fixture supports
exactly `5.9`; unsupported versions fail configuration validation rather than
silently selecting a different shell. Node is owned by the documentation tool
chain and therefore is not repeated as repository policy.

`mutation.max_workers` optionally lowers the native campaign worker ceiling.
Omission, null, or zero retains automatic admission: four workers normally and one
for modules using Testcontainers. Values one through four cap that policy;
they never increase the service-module ceiling. Negative values and values
above four are rejected. Use one for process-heavy suites whose independent
full-suite executions would otherwise compete for the same machine resources.
This changes scheduling only, not phase deadlines, discovery, test selection,
required mutation efficacy, or equivalent-mutant accounting. Older tool
releases do not understand this optional field; add it only when using a tool
version that supports it. Existing configurations need no migration.

`mutation.imports` assigns at most one approved legacy checkpoint archive and
migration ledger to a module. The files remain repository-owned under
`.verification`; before mutation verification, `golib` validates and imports
them into the current content-addressed evidence store. A previously approved
v1 identity may transition to the v2 identity only when both are derived from
the same observed source and the v1 identity is already approved. Unrelated
sibling modules are omitted from v2 package identity. Missing or ambiguous
verifier approval fails closed. A checkpoint whose package input genuinely
changed is skipped, while approved packages are imported and normal mutation
verification runs only for packages without current evidence.

An optional `.verification/mutation/equivalent-inventory.json` classifies
specific native `LIVED` mutants only when an independent review has established
equivalence under a named observable contract. Each package entry binds its
module and package directories, `source_digest`, `gremlins_version`, and
`gremlins_verifier_sha256`; each mutation names its exact `file_name`, `type`,
`line`, `column`, `contract_domain`, and detailed `reason`. Use schema version
`1` with a `packages` array. The file is absent in repositories with no such
review. A selected package with stale source or verifier identity fails closed.
Changing explanatory prose does not change native campaign identity; every
reused raw report is checked against the exact selection captured at campaign
start. Inventory edits and withdrawals apply on the next invocation, not to
a campaign already running.

The root CLI's record covers only its repeated compressed-byte bound after
ordinary stable regular-file admission at the same byte limit. Production
streaming enforcement remains intact. Changing or virtual files, arbitrary
readers and the archive validator's own limit mutations are outside that
disposition. Native statuses and failed historical runs remain unchanged;
the record does not establish a successful matching-source release rehearsal.

Typed operations may invoke bounded `go test` runs or one named target from a
repository-owned Makefile. A `test` operation runs after the module's standard
test command and preserves package-specific stress, leak, or lifecycle checks.
A `make` step is reserved for package-owned verification that cannot be
expressed as a direct Go test, such as generated-source comparison or an
external-runtime conformance suite:

```yaml
operations:
  - module: .
    gate: conformance
    steps:
      - type: make
        makefile: verification/package.mk
        target: generated
        timeout: 10m
```

The Makefile path is repository-relative, symlinks are rejected, the file is
read with a fixed size limit, and its exact validated bytes are provided to
Make through standard input. Configuration cannot select an executable, add
command-line flags, interpolate shell text, or provide secrets. Enabling fuzz,
benchmark, or conformance in a module manifest requires a matching typed
operation. Declaring an interoperability tool has the same requirement.
Missing operations fail while loading the repository rather than silently
skipping an enabled gate.
