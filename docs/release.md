# Release Process

The first public release is `v1.0.0`. A release requires a clean final diff,
exact coverage and mutation results, race and fuzz checks, static analysis,
security and license checks, documentation validation, representative consumer
parity, and successful remote CI.

Release automation builds Linux and macOS archives for amd64 and arm64,
generates checksums, SBOMs, provenance, and a release manifest, and publishes
from an immutable tag. Consumers verify artifacts before execution.

The release workflow runs only for exact semantic-version tags. It verifies
that the tag resolves to the checked-out commit and that the exact commit
already has a successful full `CI` push run. It then builds and verifies the
current source while `.golib.yaml` remains pinned to a stable published
bootstrap release. This avoids both requiring an unpublished binary and
checksum set to bootstrap their own release and repeating
content-identical mutation campaigns during publication. The workflow
revalidates release metadata, cross-compiles
a `CGO_ENABLED=0` binary for each supported platform, embeds the tag as the
binary identity, packages the binary with the license, produces SPDX JSON
SBOMs, and attests the artifacts.

Catalog publication begins only after the release candidate exists. A closed
[`release/cohesion-sources.json`](../release/cohesion-sources.json) lock binds
each consumer to an exact commit and adopted tool/checksum identity; the one
tooling entry resolves to the tagged source commit. Read-only jobs check out
those revisions without persisted credentials, submodules, or LFS, execute no
consumer builds or workflows, and run only the tagged candidate's typed source
verifier and engineering projection command. Projection jobs run twice and
compare bytes. A separate read-only job constructs the digest-bound input
manifest, generates both catalog views twice, and creates a deterministic
projection bundle. A second bounded read-only matrix independently regenerates
every projection from the same locked commit and byte-compares it with the
bundle.
A final read-only job safely unpacks that bundle, binds its exact member set,
repository order, paths, and digests back to the source lock, independently
regenerates the four catalogs, and compares every byte.

Only after binary and catalog verification succeeds does a read-only
publication-preparation job create a source- and artifact-bound
`release-manifest.json` and compute `checksums.txt` over every binary, SBOM,
source lock, input manifest, projection bundle, catalog, and the release
manifest. A separate read-only job rejects missing, extra, or mismatched
publication assets before the complete set is attested. The write-capable
publisher receives only that static set, verifies every attestation, and
creates the immutable GitHub release. It never generates or modifies a release
asset while holding publication authority.

Later releases follow semantic versioning. Never replace release artifacts or
move tags; publish a new patch release for corrections.

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

Run `golib release check` before preparing a tag and `golib release dry-run`
against the final source. The dry-run must happen before creating the tag: it
rejects existing tag identities and verifies each module through a task-owned
local proxy before running all gates. Releasable modules must use stable
versions, canonical tag prefixes, and all mandatory gates.

For a release candidate on `main`, dispatch the `CI` workflow with
`release_rehearsal` enabled. The ordinary quality and CodeQL jobs must pass,
then the release-rehearsal job runs the complete dry-run in GitHub Actions.
This keeps process-lifecycle and cleanup tests in their isolated CI boundary
instead of running them on a maintainer workstation. Do not create the tag
until the workflow's stable `Required` job passes.

Consumers can verify an archive with:

```bash
grep '  golib_1.0.0_linux_amd64.tar.gz$' checksums.txt | sha256sum --check -
gh attestation verify golib_1.0.0_linux_amd64.tar.gz --repo faustbrian/go-library-tools
```
