# Supply-chain ownership

Each standalone repository owns its source, dependency graph, imported assets,
update configuration and releases. Tools owns shared verification policy, not
an umbrella updater or release. Existing authenticated Security evidence remains
usable while its source, dependencies, fixtures and tool inputs are unchanged;
mutable advisory results require a dated refresh.

## Inventory and dependency decisions

Use each owned module's `go.mod` and `go.sum` together with a clean selected
build list (`GOWORK=off go list -m -json all`). Include test, tool, build and
fixture modules, not only packages imported by production. `modules.json` owns
module classification; intentional non-workspace consumers and fixtures still
own their dependency inputs. Workflows and local actions own action identities;
Tools' pinned scanner constants, embedded documentation `package.json` and
`package-lock.json`, mutation assets and their upstream lock records own the
verification-tool graph. Record inventories at the source or release boundary
where they will be consumed; do not manufacture ordinary-PR SBOMs.

A new direct dependency requires a recorded purpose, exact version, compatible
license and notices, maintenance evidence, security history with review date,
and replacement cost. Prefer a maintainable standard-library implementation.
Existing dependency evaluations and license/provenance records remain their
owners: do not copy them into a second mandatory manifest. Optional integrations
stay in their owning adapter or test module, not the core dependency graph.

For mutable maintenance review, use `go list -m -u -retracted -json all` with
public proxy/checksum authentication and retain the selected versions, available
updates, deprecations, retractions, errors and observation date. A failed or
partial lookup is unverified, not healthy. A newer version alone is not a
vulnerability or an instruction to upgrade. Check upstream ownership, archive
status, release history and advisories when evaluating a replacement or stale
dependency. Review intentional retention with an owner and revisit condition.
Run the pinned vulnerability checker against the actual selected graph and
preserve failures; do not infer advisory freshness from an old green commit.

## Updates and authenticated inputs

Keep exactly one repository-owned Dependabot configuration and no competing
Renovate configuration. Check its directories against actual dependency
manifests, including owned non-workspace modules and embedded tooling. Root
Actions updates do not cover nested Go, npm or Composer manifests. Keep
compatible directory globs; do not require a fleet migration merely to rewrite
equivalent scopes. Selection-owned composition dependencies must update their
selection and module inputs together rather than accepting a bot's partial
direct-version edit.

Tools' retained `release/compatibility-consumer` is catalog-generated historical
publication evidence. Its bot updates only indirect requirements; direct
requirements remain owned by the published compatibility set and generator.
Current scenario selection belongs to `go-library-assurance`. Do not rewrite
the historical consumer into a moving current selection to satisfy an updater.

Use public Go proxy and SumDB authentication, `GOWORK=off`, task-owned caches,
`go mod verify` and clean module resolution. A file proxy with SumDB disabled is
only an owned pre-publication rehearsal, not published-consumer proof. Never
publish local replacements or fixture modules. Pins and upstream locks remain
immutable; review dependency/action changes for provenance and least privilege.
Untrusted PR code must not acquire secrets or privileged publication identity.

## Imported assets and licenses

Imported specifications, schemas, corpora, timezone data, generated code and
other upstream bytes require their source URL, immutable revision, checksum,
license/notices, modifications and update procedure. Existing manifest plus
license and transformation documentation may jointly supply this information.
An upstream update must preserve applicable cases or explicitly review their
removal and its conformance consequence. Keep Apache NOTICE and attribution;
the repository license does not relicense imported assets or nested modules.

## Release assets

See the [release process](release.md) for the different binary and source-module
boundaries. Generate applicable SBOM/provenance once for the reviewed release
source, retain them with the published assets, and verify actual downloads.
Never overwrite old evidence or claim a local preparation is a publication.

The [compromise response](ecosystem/security/vulnerability-management.md#supply-chain-compromise-response)
owns credential revocation, affected versions, dependency/action replacement
and history-preserving communication. Exceptions require a named approver,
scope, evidence, mitigation and expiry; unreviewed or expired exceptions do not
support a passing release or ecosystem verdict.
