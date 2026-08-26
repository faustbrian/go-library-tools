# Release Process

The first public release is `v1.0.0`. A release requires a clean final diff,
exact coverage and mutation results, race and fuzz checks, static analysis,
security and license checks, documentation validation, representative consumer
parity, and successful remote CI.

Release automation builds Linux and macOS archives for amd64 and arm64,
generates checksums, SBOMs, provenance, and a release manifest, and publishes
from an immutable tag. Consumers verify artifacts before execution.

Later releases follow semantic versioning. Never replace release artifacts or
move tags; publish a new patch release for corrections.

Run `golib release check` before preparing a tag and `golib release dry-run`
against the final source. Releasable modules must use stable versions, canonical
tag prefixes, and all mandatory gates.
