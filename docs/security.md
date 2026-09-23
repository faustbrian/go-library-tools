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
version. SBOM stdout and stderr have independent 16 MiB limits and are never
rendered in diagnostics. This gate proves that the current module graph can
produce a valid SBOM; release workflows separately publish attestable SBOM
artifacts.

The versioned [ecosystem security contract](ecosystem/security/README.md)
defines the shared threat model, vulnerability-management process, risk
ownership, scanner-result matrix, and per-revision release verdict.
Resolved risks that permit a passing verdict bind the exact raw matrix bytes as
`sha256:<digest>:security-matrix.json`. A release tag may add only those risk
and matrix records on top of the scanned source commit; the workflow verifies
that the matrix revision is the tag commit's sole parent and that no executable,
source, dependency, manifest, or workflow content changed in between.

Security-enabled modules run pinned govulncheck, standalone gosec, a centrally
owned go-analysis security policy, full-history gitleaks with a centrally owned
configuration, a current-tree scan that includes generated artifacts, license
checks, and SBOM generation during ordinary local/PR checks as well as the full
gate. Repository-owned gitleaks suppressions are not loaded, and source-level
native gosec suppressions must name exact rules and include a reason, while
`nolint:gosec` explanations must appear on the same comment line. A consumer
cannot silently widen the shared policy. NilAway remains advisory.
Scanner-controlled stdout and stderr are suppressed and independently limited
to 4 MiB. Exceeding either limit terminates the complete scanner process tree
and returns only an owned overflow class while preserving the process failure.
Complete-tree bounded execution is supported on Darwin and Linux. On other
platforms, bounded scanner commands fail before process start; ordinary
unbounded command execution remains portable.

Module manifests reject identity fields before schema diagnostics can expose
their values. Directories must be canonical relative paths using lowercase
ASCII letters, digits, dots, hyphens, and slashes. Module paths use the same
bounded character set and root modules must match the repository identity or
its semantic-major suffix. Rejected identities never enter inventory snapshots
or cohesion output.

GitHub Actions workflows are checked with a pinned Actionlint release plus
owned checks for privileged triggers, broad permissions, persisted checkout
credentials, mutable remote action references, and non-digest container images, and
pull requests run GitHub's dependency review action from an immutable commit.
The final required job accepts dependency review as skipped only for events
that are not pull requests.

The tool executes repository tests and approved external analyzers. A repository
maintainer must therefore treat gate execution as code execution and must not
run untrusted pull-request code with secrets or write-capable credentials.

Bootstrap proxy archives are size-bounded during download and validated for
entry count, expanded bytes, path confinement, entry type, and unsafe modes
after digest verification and before extraction. Rejections expose only stable
error classes and never include archive-controlled entry names.
