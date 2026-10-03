# Golib Ecosystem

This directory is the versioned public entry point for the independently
released Golib libraries. It describes shared consumer expectations without
creating an umbrella module, application runtime, service container, or
lockstep release train.

- [Design language](design-language.md): construction, ownership, lifecycle,
  errors, adapters, compatibility, and explicit composition.
- [Cohesion v1 contract](goals/cohesion-v1.md): current module, compatibility,
  composition, assurance, and evolution boundaries.
- [Consumer catalog](catalog-consumer.md): active and deprecated public modules
  organized by capability and intended adoption boundary.
- [Engineering catalog](catalog-engineering.md): complete repository-owned
  module and package inventory for maintainers.
- [Compatibility set and receipts](compatibility-sets.md): exact known-good
  public versions, covered compositions, exclusions, and receipt sources.
- [Security contract](security/README.md): shared threat model, vulnerability
  management, risk ownership, scanner records, and release verdicts.
- [Residual register](../../release/cohesion-residuals.json): remaining
  material exceptions, migrations, and removal conditions.
- [Release process](../release.md): milestone validation, native consumer
  matrix, immutable assets, and rollback boundary.

At an ecosystem milestone, the reviewed repository revisions and adopted
tooling identities are declared in
[`release/cohesion-sources.json`](../../release/cohesion-sources.json) and
validated by
[`schema/cohesion-sources.schema.json`](../../schema/cohesion-sources.schema.json).
The matching standalone input manifest lists only SHA-256-bound repository
engineering projections described by
[`schema/cohesion-inputs.schema.json`](../../schema/cohesion-inputs.schema.json).
Projection paths are safe POSIX paths confined to the manifest directory. A
checksum-verified released `golib` binary generates deterministic consumer and
engineering catalogs; the consumer view contains releasable libraries and
adapters, while the engineering view also retains fixtures, harnesses,
examples, benchmarks, and internal tools. The matching `cohesion aggregate
check` command byte-verifies the checked-in set. Aggregation accepts at most
256 repositories, 256 MiB of projection input, and 4,096 modules; each rendered
artifact is capped at 512 MiB.

Only a selected ecosystem-milestone tooling release publishes ten
checksum- and attestation-covered assets: the source lock, residual register,
compatibility-set JSON and Markdown, standalone input manifest, deterministic
projection bundle, and four catalogs. That milestone pipeline
independently regenerates locked projections and catalogs and binds bundle
membership to the source lock before publication. Ordinary tooling and library
releases do not refresh or publish these aggregate artifacts.

Compatibility-set drafts use `draft-<YYYYMMDD>.<sequence>` identifiers and are
non-installable. Published sets use
`golib-compat-v<schema>-<YYYYMMDD>.<sequence>`, are installable, and bind every
module to its owning repository's exact remote release tag and public Go proxy
version. Set records can also carry structured recipe,
external-version, upgrade, rollback, and exclusion details. Published
identifiers are immutable; a later recommendation receives a new identifier.
The maintained `release/compatibility-consumer` may separately declare
`selection.json`: a published base set and explicit released-module
`replacements` or `additions`, each with its exact tag source revision and
primary imports. Replacements substitute a selected module path or version;
additions import a distinct major path alongside the base module, preserving
its v1 coverage. The generator checks each remote tag and public Go proxy
version and the resulting consumer files without modifying the published base
set or projecting new `known_good_compatibility_sets` membership. This
current consumer is not evidence that the changed versions belong to the frozen
verified cohort; its blank imports establish composition-only compilation, not
runtime security certification. Query, International and Localized select
their published v3.0.0 identities together, while Validation retains v2.0.0.
Money selects published `/v2` v2.0.0 together with canonical Knapsack
`objective/money/v3` v3.0.0; root Knapsack and Measurement retain `/v2`.
The separate ordinary composition test checks exact Money v2 totals and copied
cost-map ownership through canonical objective v3, not runtime security
certification. Deprecated facade selections and frozen cohort records are
unchanged.
Log selects published `/v2` v2.0.0. Its ordinary composition assertion checks
default message replacement and attribute omission, then explicit `TrustedNew`
preservation of a fixed application event and field; it is not a broader
security certification.
OpenRPC selects published `/v2` v2.0.0. Its ordinary composition assertion
builds a single method, validates it with explicit method and diagnostic
bounds, and checks exact canonical output; it does not certify all protocol
or security boundaries.
Other selections and the frozen cohort's release identities remain unchanged.
Without this optional selection, older fixtures retain the existing candidate or
published-set selection behavior. Remove the selection explicitly before
generating a new cohort candidate; candidate writes must not silently overwrite
a maintained current-release consumer. Middleware, feature-flags, measurement,
filesystem, fault-injection, router, password, semaphore and Knapsack select
their published `/v2` modules, along with the optional measurement rule adapter.
Barcode, Rate Limit, Sequencer and Idempotency
add their released `/v2` modules while retaining v1 imports. Neither form
rewrites the base set's v1 versions or historical evidence.
Matching module entries in the generated catalogs list the published set under
`known_good_compatibility_sets`; unreleased drafts are never projected there.
Candidate fingerprints withdrawn before the first publication are not public
identities; `golib-compat-v1-20260910.1` becomes immutable only with its v1.8.0
milestone release.

## Adopt, upgrade, or roll back

Select only the modules an application directly uses from the published set,
apply those exact versions in the application's `go.mod`, and run its directly
affected compositions with `GOWORK=off` and public proxy/SumDB resolution. The
set is guidance, not an umbrella module or mandatory fleet upgrade.

Upgrade one owned dependency boundary at a time. To roll back, restore the
application's prior direct versions for that boundary and rerun the same
focused compositions; unrelated module versions do not need to change.

After the final aggregate consumer catalog contains the complete active public
roster, generate one unreleased candidate and its clean-consumer module with:

```sh
make compatibility-candidate \
  COMPATIBILITY_SET_ID=draft-YYYYMMDD.1 \
  COMPATIBILITY_OBSERVED_AT=YYYY-MM-DDTHH:MM:SSZ \
  COMPATIBILITY_MODULE_COUNT=<active-module-count> \
  COMPATIBILITY_VERSION_OVERRIDES='<module>@<public-version> ...'
```

Generation selects only active, releasable catalog modules, sorts them by
module path, applies any deliberate public-version overrides, resolves every
selected version through its exact remote tag and the public Go proxy, and
writes `compatibility-sets.{json,md}` together with
`release/compatibility-consumer/{go.mod,consumer_test.go}`. A missing release,
deprecated or planned module, stale catalog version, or roster-count mismatch
fails before a candidate is written. Do not run the write command until the
catalog, public releases, candidate metadata, and receipt sources are final.

Composition receipts run from task-owned disposable source checkouts. Rebase
only dependencies already required by that checkout, reject any local replace,
then tidy and execute with public resolution:

```sh
make compatibility-rebase \
  COMPATIBILITY_SET_ID=draft-YYYYMMDD.1 \
  COMPATIBILITY_GO_MOD=/task/checkout/go.mod
GOWORK=off go mod tidy
GOWORK=off go test -mod=readonly ./...
```

`compatibility-rebase` reads the selected candidate, updates matching existing
requirements to its versions, and leaves unrelated dependencies unchanged. It
does not edit a maintained repository unless that repository is itself the
explicit task-owned checkout.

The development copy on a default branch is not an immutable compatibility
contract. Published documents identify their design-language version, exact
`go-library-tools` tag, and content digest.
