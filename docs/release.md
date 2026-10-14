# Release Process

Releases require a clean reviewed final diff, successful exact-source remote
CI, documentation, lint, tests, security and license checks. Select affected API,
integration and direct-consumer qualification for the actual changed contracts;
verify the published version through a clean public consumer. A normal library
release does not require a new Tools release or ecosystem-catalog publication.

Coverage, mutation, race, fuzz, performance and full rehearsal work require a
named changed risk or relevant release boundary. Reuse evidence for unchanged
source, dependencies, tools and material environment inputs; do not repeat full
campaigns merely because a tag, prose or progress record changed. Known concrete
defects block affected delivery until resolved under the security/release policy.
Unresolved viable survivors require repair or an explicit behavioral-risk
disposition; moving a campaign to a schedule is not that disposition. Missing selected evidence,
timeouts and infrastructure failures are not passing results. Scheduled work
does not waive a known defect or an untested critical boundary.

The root module explicitly selects coverage evidence mode: it still runs the
complete instrumented module tests and requires valid, nonzero execution for
every expected production package. Counts do not certify behavioral adequacy.
The release review owns the assessment of reachable behavior and defensive
guards that cannot be independently reached through admitted inputs; tests
must cover concrete scanner metadata, failure and redaction contracts. Other
modules retain exact coverage when selected unless their own policy explicitly
selects evidence mode. Selected mutation campaigns retain complete viable-mutant
accounting and narrow reviewed equivalent/invalid dispositions; this policy
does not lower their acceptance threshold. See [verification](verification.md).

Release automation builds Linux and macOS archives for amd64 and arm64,
generates checksums, SBOMs, provenance, and a release manifest, and publishes
from an immutable tag. Consumers verify artifacts before execution.

The release workflow runs only for exact semantic-version tags. It verifies
that the tag resolves to the checked-out commit and that the exact commit
already has a successful latest applicable `CI` main push run. The tagged commit
must belong to the fetched immutable main history, but later main commits do not
invalidate it. Tags outside main, unavailable ancestry and non-green or wrong-source CI
are refused. Native source tags require actual successful Required, Quality,
CodeQL runtime-build and pinned analyze execution, not merely green jobs with
skipped scanner steps. An authenticated merged PR must identify the tag and
target main. A distinct PR head is accepted only with Git tree equality to the
tag. The release runs one pinned, five-minute dependency review with that PR's
authenticated base SHA and the exact tagged SHA, retaining low-severity,
license and vulnerability refusal without comments or write permissions.
Missing ref outputs fail closed. Historical PR CI associations are not reused:
GitHub may clear them after merge, and they do not prove the release comparison.

Native releases retain valid human risk dispositions and historical scanner
records as policy inputs, without fresh eleven-row transcription or a
records-only commit. They run the existing security validator, one full secret
scan and release metadata check using one private cache. Scanner execution does
not promise zero CodeQL findings; reviewed severity, expiry and unresolved-risk
policy remain required. Existing records-only tags retain their supported
sole-parent interlock; see [security](security.md).

It then builds and verifies current source while `.golib.yaml`
remains pinned to a stable published
bootstrap release. This avoids both requiring an unpublished binary and
checksum set to bootstrap their own release and repeating
content-identical mutation campaigns during publication. The workflow
revalidates release metadata, cross-compiles
a `CGO_ENABLED=0` binary for each supported platform, embeds the tag as the
binary identity, packages the binary with the product license and
`THIRD_PARTY_NOTICES`, produces SPDX JSON
SBOMs, and attests the artifacts.

The Ubuntu build lane retrieves the Syft v1.54.0 Linux amd64 archive and checks
its reviewed SHA-256 before extracting or executing it. Version checks and
task-owned caches retain the existing SPDX JSON outputs; no mutable upstream
installer is executed. Updating this executable requires review of the release
asset identity and checksum, not just changing a version label.

`golib release package --binary PATH --license LICENSE --output ARCHIVE`
creates a new deterministic three-member archive without overwriting existing
output. `THIRD_PARTY_NOTICES` retains exact reviewed license, NOTICE, patent
grant and copyright attribution bytes for the compiled Go dependency graph
and the pinned Go runtime. The embedded
[`binary license mapping`](../internal/releasecheck/licenses.json) binds module
versions and Go module checksums to source URLs and file SHA-256 values. It does
not distribute notices for unused go.mod entries or infer licensing from names.
New/replaced dependencies, changed versions/checksums, missing/changed material,
unreviewed runtime versions and CGO/experimental builds fail closed. Review the
actual upstream license obligations and refresh this corpus when those inputs
change; module source/license acquisition retains normal Go checksum validation.

`golib release archive-check --file ARCHIVE` independently checks safe archive
structure and compares its product license and notices byte-for-byte with the
compiled binary's reviewed mapping without executing it. Every platform build
uses this check before attestation; catalog preparation repeats it after
attestation verification. Extraction scripts must accept
the additional notice member and retain it when redistributing the binary.
This new contract does not retroactively qualify or replace older published
two-member archives; corrections require a new release.

Normal tooling releases publish eight payloads: four platform archives and four
matching SBOMs. A read-only publication-preparation job creates the source- and
artifact-bound `release-manifest.json` and computes `checksums.txt` over those
payloads and the manifest. A separate read-only job rejects missing, extra, or
mismatched assets before the complete set is attested. The write-capable
publisher receives only that static set, verifies every attestation, and
creates the immutable GitHub release. It never generates or modifies a release
asset while holding publication authority.

Catalog and source-lock publication is a separate ecosystem milestone. Before
creating the selected tag, refresh and review
[`release/cohesion-sources.json`](../release/cohesion-sources.json), then set
the repository variable `GOLIB_CATALOG_MILESTONE_TAG` to that exact tag. The
exact-tag comparison prevents a stale variable from expanding later tooling
releases. A selected milestone adds ten payloads: the source lock, residual
register, compatibility-set JSON and Markdown, input manifest, deterministic
projection bundle, and both JSON and Markdown catalog views.

For a selected milestone, read-only jobs check out every locked revision
without persisted credentials, submodules, or LFS. They execute no consumer
builds or workflows and run only the tagged candidate's typed source verifier
and engineering projection command. Projection jobs run twice and compare
bytes. A separate read-only job constructs the digest-bound input manifest,
generates both catalog views twice, and creates the projection bundle. A second
bounded read-only matrix independently regenerates every projection from the
same locked commit and byte-compares it with the bundle. A final read-only job
binds the bundle's member set, repository order, paths, and digests back to the
source lock and independently regenerates all four catalogs. Only after those
checks pass are the ten milestone payloads included in the same manifest,
checksums, attestation, and immutable publication path as the tooling assets.

Later releases follow semantic versioning. Never replace release artifacts or
move tags; publish a new patch release for corrections.

## Source-module publication

Source modules are not Tools binary distributions. Their consumer artifact is
the public module zip, authenticated by the Go proxy and checksum database;
do not create platform archives for a library that ships no command. Each
module's maintainer owns publication, with explicit approval for tagging and
publishing. `release check` and `release dry-run` are prerequisites, not asset
publication commands.

From a clean reviewed source with successful applicable exact-source CI, prepare
one module-scoped dependency inventory and applicable SBOM. Record the canonical
module/version, tag prefix, source commit/tree, selected graph, generator
identities, checksums and preparation environment. Retain the assets at that
release boundary. Provenance must describe the actual builder: locally prepared
and maintainer-signed evidence must not claim a hosted or independently
attested build. Sign the checksums or attest the distributed assets using the
maintainer's approved identity; retain signer identity and verification guidance.

After authorized publication, verify the remote tag's exact commit and module
directory, download the actual assets and verify their membership, digests and
signature/attestation. Consume the published canonical module and version in a
clean temporary module with `GOWORK=off`, public proxy/SumDB authentication and
no replacements. Exercise its real public contract, not merely a cached or
file-proxy lookup. For distributed commands, use the clean reproducibility
comparison appropriate to their declared build environment. A module-only
release does not claim binary reproducibility.

Reuse already verified signed assets and public consumers for untouched
releases. Correct evidence through new assets with explicit supersession or a
new release as appropriate, never by moving a tag or replacing historical
bytes. A missing or mismatched asset, signature, source identity or actual
public-consumer result remains an unverified release boundary. See
[supply-chain ownership](supply-chain.md) for inventories and imported assets.

Consumer upgrades use `golib upgrade plan` before `golib upgrade apply`. The
command rejects incomplete, duplicated, mutable, or mismatched pins and changes
only `.golib.yaml` and the thin CI caller. It restores the configuration if the
workflow replacement fails. Machine-readable output supports reviewable
per-repository pull requests and bounded cohorts without coupling normal CI to
a live sibling-repository inventory.

The canonical [`consumers.json`](../consumers.json) manifest controls rollout
eligibility. Validate it with `golib consumers validate`, then dispatch the
upgrade workflow in dry-run mode before apply mode. Apply mode requires the
fine-grained cross-repository `GOLIB_ROLLOUT_TOKEN` described in the
[consumer rollout guide](consumers.md). Initial adoption remains a manual,
parity-reviewed migration rather than an automated upgrade.

Run `golib release check` before preparing a tag. It validates metadata and
required gate configuration; it does not execute tests or certify release
readiness. Releasable modules use stable versions, canonical tag prefixes and
mandatory documentation, lint, security and test gates. Coverage, mutation and
race configuration alone is not a universal release prerequisite.

`golib release dry-run` verifies security and candidate packaging, not complete
runtime qualification. It retains suppression admission, current/history secret
scans, vulnerability/security/license checks and SBOM generation for selected
security-enabled modules. It rejects existing tags, constructs task-owned proxy
archives and proves candidate module resolution. It does not start services or
repeat full tests, race, coverage, mutation, lint or documentation campaigns.
Its local candidate proxy does not replace clean consumption of the actual
published public version.
Qualify affected contracts through focused tests and applicable exact-source CI.
Run `golib check --module DIR` separately only when a named risk requires a full
all-enabled campaign; selected failures still block the applicable qualification.
Keep broad mutation assurance in the selected main/scheduled workflow rather
than making every tooling patch wait for it. Unrelated library development uses
its existing supported tooling pin and must not wait for a Tools release unless
it actually depends on that tooling correction.
In independently versioned repositories, pass `--module DIR` to confine metadata,
security, tag collision and proxy-consumption work to the exact release owner.
Repository structure and specification validation remain global prerequisites;
omitting the selector or passing `--all` selects every releasable module.

When a hosted packaging rehearsal is selected for a candidate on `main`, dispatch
the `CI` workflow with `release_rehearsal` enabled. Set `release_module` to the exact independently
versioned module directory, or leave it blank for the repository-wide dry-run.
The ordinary quality and CodeQL jobs must pass, and the release-rehearsal job
runs the selected packaging/security dry-run with a 30-minute ceiling.
Reusable release CI retains ordinary module tests instead of replacing them
with packaging checks. Process-lifecycle and cleanup tests remain in their
isolated hosted Quality boundary, not on a maintainer workstation.
The selected rehearsal must pass before
tagging; a skipped or failed selected job is not qualification. Releases always
retain their applicable exact-source required CI and current-source security
evidence interlock, plus checksums, SBOMs, attestations and immutable publication.
The identity/metadata verification job has a 30-minute ceiling and does not run
a mutation campaign; budget exhaustion fails publication.

Reusable-workflow consumers must expose a `release_module` dispatch input and
forward it to `library-ci.yml` before relying on a root-only hosted rehearsal.
Updating only the tooling pin does not add or wire that caller input.

Consumers can verify an archive with:

```bash
grep '  golib_1.0.0_linux_amd64.tar.gz$' checksums.txt | sha256sum --check -
gh attestation verify golib_1.0.0_linux_amd64.tar.gz --repo faustbrian/go-library-tools
```
