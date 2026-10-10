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
The historical public-tool-identity exception for root `Makefile` and exact
`.golib/package.mk` is restricted to `generic-api-key` findings on complete
`APIDIFF_VERSION` assignments using `:=` or `?=` with the public version form
`v0.0.0-<14 digits>-<12 lowercase hex>` on complete lines.
Only spaces/tabs around the operator and after the version, plus the scanner's
optional single leading newline, are admitted. The line target binds the
assignment identity, not merely the version-shaped value: other keys, paths,
rules, operators or surrounding content remain scanned. No file or commit is
allowlisted wholesale, and the existing `.golib/versions.env` exception is
unchanged. Owner: library maintainer; review when historical tool identity or
pinned scanner line-framing semantics changes. Full-history and current-tree
scanning, private finding metadata and suppression accounting are unchanged.
Gosec receives explicit directories from `go list -json=Dir ./...` with
`GOWORK=off`, not the scanner's recursive filesystem pattern or a production
package manifest. This includes newly added and test-support packages selected
by Go; ignored fixture trees and nested modules are not selected. Intentional
fixture qualification remains a separate, dependency-complete check owned by
the relevant repository. Discovery uses private, independently bounded 4 MiB
streams and the same cancellation owner as scanners. Empty, failed, malformed,
duplicate or out-of-module selections fail closed. At most 4096 directories,
4096 bytes per argument and 128 KiB of package arguments are admitted. No
scanner rules or source suppression requirements change.
The Authentication metadata exceptions require the generic-api-key detector,
the exact root `CHANGELOG.md` or `jwt/specification/README.md` path, and an
exact enumerated line. The 16 deduplicated AUTH decision-ID/SHA-256 lines are
registered in both immutable decision histories at Authentication commits
`0062d0a8c92135e6020522db36d361f7ca9e3b5d` and
`2931094d5e905565e52f0be9cada6b009a147939`. The single JWT-DEC-004 table row
at the first commit names public RFC authorities and real executable tests;
its comma-separated test names are not a credential assignment. Only the
scanner's optional single leading newline is accepted. Other detectors,
paths, changed metadata and adjacent assignments remain eligible for scanning.
Owner: library maintainer; review when the public source facts or pinned
scanner framing changes. These exceptions do not waive any historical gate
failure or claim a completed Authentication qualification.

The centrally owned generic-api-key exception for JSONAPI decision metadata
matches only 23 exact decision-ID/SHA-256 lines in root `CHANGELOG.md`.
The optional single leading newline accommodates Gitleaks v8.30.1's line
framing; no arbitrary whitespace or surrounding content is admitted.
Each checksum was independently recomputed from the decision register at
`4d52a36a5b76853597bcf85506f3c4ba2b8a1262` or
`ad9dda2d359e7474b25abe59909818cc68686bf9` and matched its decision history.
These public checksums are not credentials. The shared exception is content
bound, not repository or commit bound: an identical public line is recognized
elsewhere, but another ID, checksum, path, surrounding text, rule or credential
assignment remains scanned. Owner: library maintainer; review condition:
changed decision provenance, scanner matching semantics or exception scope.
The OpenAPI metadata exception likewise requires the generic-api-key detector,
root `CHANGELOG.md`, and one of 14 exact public decision-ID/SHA-256 lines.
Their checksums were independently recomputed from decision registers at
`09771cf1e03bbdb684cca77785cbc7ee1ee2e9a1` and
`92dcffd42c0dbd6d27c526e5cf33ea7f9b8591f1` and matched their history.
Only the scanner's optional leading newline is admitted. Changed pairs,
paths, surrounding text, credential assignments and other detectors remain
findings. Owner and review conditions are the same as for JSONAPI metadata;
this classification does not waive other findings in OpenAPI history.

Scanner-controlled stdout and stderr are suppressed and independently limited
to 4 MiB. Gitleaks uses a distinct findings exit status; only the exact final
Go-wrapper status line is recognized, without retaining scanner text. Completed
findings produce the fixed `secret-findings` diagnostic and still fail the gate.
An owned report template emits only SHA-256 file and rule identities, positive
line numbers and Git commit identities. At most 32 validated, 256-byte locations
are included as `secret-location <file-sha256> <rule-sha256> <line> <commit>`;
current-tree findings use `-` for the commit. Additional locations are explicitly
omitted. Filenames, rule text, finding payloads and reporter metadata are never
rendered. Invalid or incomplete location metadata is suppressed in full.
Other failures remain unclassified; cancellation or overflow cannot qualify as
completed findings. Exceeding either limit terminates the scanner's original process group
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

### Immutable public Password and GitHub metadata

The same exact-detector/path/line policy recognizes the independently generated
Webhook v1 test-key and canonical-message records, including the scanner's
decoded canonical-message lines, at
`testdata/vectors/v1.json`. Its Python standard-library generator at
`525998a52d79a0ddc8827719a9f759e418f98674` matches the current generator, and
the checked fixture reproduces exactly. These are public test inputs, not
application signing credentials. The allowance does not cover other keys or
future vector versions.

Two exact Verkle source-checksum records in `specification/sources.json`
were recomputed from public `crate-crypto/go-eth-kzg` revision
`01d14404df5f295f5afb3cf6ca8839382a7243b7`: `api.go` and
`api_eip7594.go`. Only those complete immutable lines are recognized, not
arbitrary hashes, manifests or source directories. This scanner disposition
does not approve the candidate cryptographic backend or remove its documented
experimental production blockers.

The central policy recognizes only complete public lines at their exact paths
and the `generic-api-key` detector. Password decision records were independently
recomputed from the decision registers at commits
`481f9392d0abeed878e7379cb0732b8a4521ceae` and
`828d92419da8ef8bfe558e0fab21b4913a8a0f6b`. Its historical
`tools/versions.env` allowance covers one exact APIDIFF module identity from
`c79001812c419c28b2445cc23ae1bd3985b7b6c7`, not arbitrary version-shaped
assignments or that entire file.

The GitHub `temp_clone_token` example comes from the public `descriptions-next`
artifact at `417c4fb368fc6a7162ce5f3eeeddce1a9a217747`. The independently
downloaded artifact matches the historical OpenAPI fixture and source manifest
SHA-256 `9d85f3a842c0215768f30f83ac7d1595430236fc51ce9c84e344b991a9f6b3da`.
Only its exact example lines under
`specification/independent/github-rest-api/api.github.com.2022-11-28.json`
are recognized; no token class or upstream directory is exempted.

Pinned scanner regressions exercise both Git history and the current tree.
Altered values, different paths, and adjacent or same-line credential
assignments must remain findings. The central policy owner must review any
new immutable record before extending these allowances; public provenance
does not automatically make a future token or changed artifact acceptable.

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
