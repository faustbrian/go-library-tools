# Security Model

Threats include path traversal and symlinks, malicious YAML or JSON, command
injection, environment poisoning, unbounded process output, forged evidence,
cache contamination, service cleanup races, secret disclosure, mutable action
references, compromised release archives, and privileged pull-request code.

Controls include bounded strict decoders, repository-contained path reads,
argument-array execution, disposable task workspaces, content-addressed
evidence, digest-pinned images, exact resource ownership, redacted diagnostics,
least-privileged workflows, immutable action pins, checksum verification,
SBOMs, and provenance.

For every module with the `security` gate enabled, `golib check` generates a
deterministic CycloneDX 1.6 library SBOM with the pinned `cyclonedx-gomod`
version. The CLI captures output in memory with a hard size limit, rejects
empty or malformed output, and verifies the CycloneDX format and specification
version. This gate proves that the current module graph can produce a valid
SBOM; release workflows separately publish attestable SBOM artifacts.

GitHub Actions workflows are checked with a pinned Actionlint release, and
pull requests run GitHub's dependency review action from an immutable commit.
The final required job accepts dependency review as skipped only for events
that are not pull requests.

The tool executes repository tests and approved external analyzers. A repository
maintainer must therefore treat gate execution as code execution and must not
run untrusted pull-request code with secrets or write-capable credentials.
