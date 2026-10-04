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

Owned Markdown validation admits only regular files, at most 4 MiB each and
4096 documents per selected documentation tree, before reading or parsing.
Directory enumeration uses batches of at most 128 entries, a shared allowance
of 100,000 entries (including ignored files and directories), and maximum
directory depth 128 from the tree root at depth zero. Excluded directories are
counted but not descended into; admitted documents retain deterministic order.
Document reads reuse the repository-file admission owner, recheck opened-file
identity and type, and retain a one-byte bounded lookahead. Caller cancellation
is checked during enumeration, reads, line/link validation and around bounded
Markdown parsing. Repository/documentation roots retain their existing
canonicalization policy, and below-root symlinks remain refused. The caller
must supply a trusted, stable filesystem: these checks are not an atomic
snapshot and cannot preempt an in-flight filesystem operation or parser call.

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
configuration, a current-tree scan that includes existing generated artifacts, license
checks, and SBOM generation during ordinary local/PR checks as well as the full
gate. Repository-owned `.gitleaksignore` files are rejected before execution,
and inline `gitleaks:allow` directives are ignored by both scans. Source-level
native gosec suppressions must name exact rules and include a reason, while
`nolint:gosec` explanations must appear on the same comment line. A consumer
cannot silently widen the shared policy. NilAway remains advisory.
Scanner-controlled stdout and stderr are suppressed and independently limited
to 4 MiB. Exceeding either limit terminates the scanner's original process group
and returns only an owned overflow class while preserving the process failure.
Original-process-group bounded execution is supported on Darwin and Linux. On other
platforms, bounded scanner commands fail before process start; ordinary
unbounded command execution remains portable.

After a bounded scanner's wait completes, its cancellation monitor is joined
before the executor returns. Caller cancellation or deadline expiry observed by
the terminal result policy is preserved alongside any child failure; classification
does not depend on which monitor branch ran. This does not expand the original
process-group termination boundary or promise interruption of arbitrary IO.

Pinned scanners and Git are trusted executable collaborators processing hostile
data. A process which detaches from its original group is outside that group's
termination boundary. Repository tests do not use the scanner group mechanism.
Golib is not a sandbox for hostile executables: arbitrary untrusted PR code
requires disposable OS, VM, container, or ephemeral hosted-runner isolation,
without secrets or write-capable credentials.

History and current-source snapshots are acquired and scanned once per check
invocation, before any repository tests or release consumer builds execute.
The reusable workflow uses one invocation across selected modules, retaining
module-attributed analyzers, licenses, and SBOMs. Standalone `--module` checks
still perform both repository scans. Missing or failed source scans fail closed.
All selected modules' vulnerability, gosec, owned-analysis, license, and SBOM
scans also finish before repository execution; runtime service scopes remain
module-owned and start afterward. Repository manifests are capped at 64 modules.
Before creating a snapshot, history is bounded to 10,000 refs, 100,000 objects
(including unreachable objects),
and 4 GiB of total uncompressed object bytes. Bundle output has an independent
4 GiB hard write cap. Current-tree preflight and copying count all entries,
including directories and links, up to 100,000 entries and 4 GiB; symlinks are
not followed or copied. Directory enumeration is incremental and traversal depth
is capped at 128. Cancellation is checked during traversal and copying.

Module manifests reject identity fields before schema diagnostics can expose
their values. Directories must be canonical relative paths using lowercase
ASCII letters, digits, dots, hyphens, and slashes. Module paths use the same
bounded character set and root modules must match the repository identity or
its semantic-major suffix. Rejected identities never enter inventory snapshots
or cohesion output.

GitHub Actions workflows are checked with a pinned Actionlint release plus
owned checks for privileged triggers, broad permissions, persisted checkout
credentials, mutable remote action references, and non-digest container images.
Repository-contained local actions are recursively inspected, including nested
composites and Docker action image references, with cycle, depth (32), descriptor
count (512), per-file (4 MiB), and combined-byte (32 MiB) limits. Traversal and
symlink paths fail closed. Docker action images must be immutable digests;
local Dockerfile builds are not accepted by this immutability gate. In addition,
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

Evidence `Parse` read, decode and trailing-data failures expose only static
categories, never reader causes, field names or parser-controlled value text.
`ErrInvalid`, strict unknown-field/trailing validation and zero failure records
are retained. `Load` and `evidence inspect` propagate these categorical parse
errors; unknown-module failures omit the rejected record's module value.
Successful explicit inventories still contain record fields and must be handled
according to their data sensitivity. This control does not
claim all filesystem path diagnostics or explicitly inspected records are
automatically redacted.

The mutation bootstrap/report strict parser and archive-admission owners also
return static rejection categories without reader/parser causes, JSON keys,
archive entry names, rejected report filenames or semantic identities. This
includes strict JSON consumers and reader-error categories for zero/equivalent
inventories. These changes
retain rejection, sentinel classification and accepted complete accounting;
they do not mask the separate admitted mutation-coordinate diagnostics or
successful checkpoint/report fields. Caller filesystem paths and other parser
owners outside this scope are not covered by this default-text guarantee.

Mutation source digests and historical package-path proofs enumerate directory
entries in fixed batches, capped at 100,000 entries before retaining excess
metadata. They admit regular source files against the existing 16 MiB per-file
and 256 MiB aggregate bounds before bounded reads; per-file admission precedes
aggregate admission. Admitted entries are sorted and digest bytes are unchanged.
Regularity, confinement and opened-file identity checks remain fail closed.
`SourceDigest` validates every module and package directory component below the
trusted caller-selected root before enumeration, rejecting symlink and
non-directory ancestors with no digest. Metadata failures retain their causes.
These operations assume a trusted, stable caller filesystem, not an atomic or
race-resistant snapshot. `SourceDigest` remains non-context-aware; this resource
control does not add cancellation or arbitrary filesystem-operation deadlines.

Public evidence inspection validates every existing component below the caller's
repository root before walking or returning an empty missing-root result.
Symlink and non-directory components are rejected with a categorical
`ErrInvalid`; a genuinely missing directory still produces an empty inventory.
This assumes a trusted, stable repository root and filesystem: component
validation and subsequent walking are not atomic or race-resistant traversal.

The shared OpenSearch readiness and RabbitMQ fixture-control HTTP client imposes
a five-second request deadline, preserving earlier caller deadlines and caller
cancellation. Its isolated transport disables keep-alive reuse and rejects
redirects. This bounds cooperative HTTP operations, not arbitrary injected
transports or process/container lifecycle work.
