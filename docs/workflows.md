# Reusable Workflows

Consumer CI uses the reusable workflow from this repository at an immutable
commit SHA. The nearby comment records the corresponding release. The setup
action installs the exact `.golib.yaml` tool version, verifies checksums, and
does not execute downloaded shell installers.

Consumer workflows retain least-privileged permissions, explicit concurrency,
module matrices, attributable evidence artifacts, scheduled checks, CodeQL,
release dry-runs, and one stable final required job.

Tool version, binary checksum, setup-action SHA, and reusable-workflow SHA are
updated together through reviewable pull requests. Existing consumers do not
change behavior when a new tooling release is published.
