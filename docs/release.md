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
current source while `.golib.yaml` remains pinned to the previous published
release. This avoids both requiring an unpublished binary and checksum set to
bootstrap their own release and repeating content-identical mutation campaigns
during publication. The workflow revalidates release metadata, cross-compiles
a `CGO_ENABLED=0` binary for each supported platform, embeds the tag as the
binary identity, packages the binary with the license, produces SPDX JSON
SBOMs, and attests the artifacts. The publish job creates a source- and
artifact-bound release manifest,
`checksums.txt`, attestations for both, and the GitHub release only after every
platform succeeds.

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
