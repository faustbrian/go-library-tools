# Goal: Centralize Go Library Tooling And Migrate Every Repository

## Role

Act as the principal engineer responsible for designing, implementing,
hardening, releasing, and adopting the shared tooling used by every maintained
Go library under `/Users/brian/Developer/golib`.

This is an end-to-end implementation and migration goal. Do not stop after
creating a CLI, proving one repository, drafting a migration plan, or replacing
only the obvious shell wrappers. The goal is complete only when the shared
tooling is production-ready, every applicable standalone library uses it, and
the duplicated `.golib` implementations have been removed.

## Context

The standalone libraries currently duplicate a private build and verification
system under `.golib`. A representative repository contains more than fifty
tooling files for module orchestration, coverage, mutation testing, API
compatibility, fuzzing, documentation, CodeQL builds, service fixtures,
evidence handling, and release checks. GitHub Actions and Makefiles invoke
those files directly.

This duplication is already divergent and will become harder to update safely.
The replacement must centralize behavior without hiding package policy,
weakening gates, discarding valid evidence, creating a second source of truth,
or requiring synchronized copy-and-paste changes across dozens of repositories.

At goal creation time, `/Users/brian/Developer/golib` contains approximately:

- 85 Git repositories including this tooling repository;
- 82 repositories with a `.golib` implementation;
- 29 genuine multi-module repositories using `go.work`;
- package-specific mutation evidence, API baselines, interoperability fixtures,
  service requirements, and release metadata that must remain attributable to
  the package source they verify.

Treat current filesystem and remote state as authoritative. Re-inventory before
implementation and again before final completion. Do not hard-code the counts
above as permanent truth.

## Objective

Create and release `github.com/faustbrian/go-library-tools` as the single
maintained implementation of the Go-library repository contract, then migrate
every applicable repository under `/Users/brian/Developer/golib` to it.

The final state must provide:

1. a versioned `golib` CLI implemented primarily in Go;
2. reusable, immutable-reference GitHub Actions workflows and setup actions;
3. one strict repository configuration contract with a published schema;
4. repository-owned source-specific verification evidence and fixtures;
5. consistent local and CI behavior;
6. no copied `.golib/scripts`, `.golib/package.mk`, tool-version files, or
   duplicated workflow implementations in consumer repositories;
7. no fallback to legacy `.golib` behavior after a repository is migrated;
8. an independently testable and releasable tooling repository;
9. complete migration of all active library repositories without modifying
   library runtime behavior or public APIs.

## Non-Goals

- Do not create a runtime framework used by library consumers.
- Do not merge the standalone library repositories back into a monorepo.
- Do not centralize package source, tests, fixtures, API baselines, or mutation
  evidence whose meaning is tied to one repository revision or content digest.
- Do not weaken 100% meaningful production coverage or 100% mutation-kill
  requirements.
- Do not make CI green by skipping packages, swallowing failures, lowering
  thresholds, classifying production packages as test support, or adding broad
  exclusions.
- Do not retain arbitrary package-owned shell hooks as an escape hatch. A hook
  mechanism that lets every repository recreate its own build system defeats
  this goal.
- Do not retag or rewrite existing library releases. Tooling-only migration does
  not change library runtime versions or public API compatibility.
- Do not overwrite unrelated or concurrent work. In particular, inspect dirty
  repositories and preserve specialist-owned changes exactly.

## Required Final Repository Model

### `go-library-tools`

Use a cohesive structure such as:

```text
go-library-tools/
├── cmd/golib/
├── internal/
│   ├── config/
│   ├── inventory/
│   ├── gates/
│   ├── coverage/
│   ├── mutation/
│   ├── api/
│   ├── documentation/
│   ├── services/
│   ├── evidence/
│   ├── release/
│   └── repository/
├── schema/
│   └── golib.schema.json
├── .github/
│   ├── actions/setup-golib/
│   └── workflows/library-ci.yml
├── docs/
├── testdata/
├── go.mod
├── Makefile
└── README.md
```

The exact internal package split may change when evidence supports a better
design. Public CLI behavior, configuration, workflow inputs, evidence formats,
and exit semantics must be deliberate and documented.

### Consumer repositories

The intended consumer shape is:

```text
go-queue/
├── .golib.yaml
├── .verification/
│   ├── mutation/
│   └── package-specific evidence or fixtures
├── .github/workflows/ci.yml
├── api/baseline.txt
├── modules.json
├── packages.json
├── Makefile
└── ... library source and documentation
```

Only keep repository-owned files that contain package policy, source-specific
evidence, or package-specific fixtures. The implementation that interprets
those files belongs in `go-library-tools`.

Do not move files merely to match the example. Preserve established durable
locations such as `api/baseline.txt` when they already communicate intent.
Rename `.golib` evidence to `.verification` only when the new location is
clearer and every reference is migrated atomically.

## Canonical Configuration

Define a strict, versioned configuration format, expected to be `.golib.yaml`,
and publish both its JSON Schema and human documentation.

The configuration must contain only repository-specific policy and a tooling
compatibility pin. It must not duplicate facts already canonical in
`modules.json`, `packages.json`, `.go-version`, `go.mod`, or `go.work`.

At minimum, define ownership for:

- configuration schema version;
- required `golib` tool version or compatible release line;
- canonical module/package manifests;
- enabled gates and explicit justified exceptions;
- required external services and fixture identifiers;
- build and test tags;
- conformance and interoperability commands represented as typed operations;
- documentation paths and intentional exclusions;
- API baseline locations;
- mutation evidence locations and zero-mutant declarations;
- benchmark and release policies;
- repository-specific runtime prerequisites such as Node, Deno, Java, zsh,
  PostgreSQL, Valkey, RabbitMQ, Kafka, NATS, NSQ, or Keycloak.

Requirements:

- Unknown fields must fail validation.
- Invalid combinations must fail with actionable field paths.
- Defaults must be explicit, stable, documented, and testable.
- Configuration upgrades must have deterministic migrations.
- The CLI must reject unsupported future schema versions.
- One setting must have one owner. Do not permit contradictory values across
  `.golib.yaml`, `modules.json`, workflow inputs, and environment variables.
- Environment overrides must be narrowly defined, documented, and visible in
  diagnostic output. They must not silently weaken gates.
- Secrets and credentials must never be accepted as committed configuration.

## CLI Contract

Implement a discoverable `golib` CLI with stable commands equivalent to:

```text
golib check
golib check --all
golib check --module <directory>
golib repository check
golib config validate
golib inventory
golib coverage
golib mutation
golib mutation --module <directory>
golib api check
golib api update
golib docs check
golib services start <fixture>
golib services stop <fixture>
golib release check
golib release dry-run
golib evidence inspect
```

Command naming may be refined for consistency, but every current capability
must have a clear replacement. Provide `--help`, machine-readable output where
CI needs it, stable exit codes, cancellation, timeouts, signal handling, and
redacted diagnostics.

The CLI must:

- operate from the repository root or resolve it deterministically;
- support root-only and genuine multi-module repositories;
- use task-owned disposable `GOCACHE`, `GOMODCACHE`, temporary directories,
  service resources, and generated outputs where appropriate;
- clean task-owned resources on success, failure, cancellation, and signals;
- never remove shared or user-owned resources;
- avoid hidden global state and process-wide mutable configuration;
- make every selected module and gate visible before execution;
- preserve deterministic ordering in output and artifacts;
- distinguish skipped, unavailable, passed, failed, and advisory gates;
- fail closed when required tools, evidence, fixtures, or configuration are
  missing;
- provide concise human output and structured JSON output without changing
  semantics;
- avoid shelling out where a reliable Go implementation is practical;
- isolate unavoidable external processes behind tested interfaces;
- never require the archived monorepo or a sibling checkout.

## Gate Parity

Inventory every distinct `.golib` implementation before designing the final
command graph. Do not assume one representative repository contains every
variant.

The shared tool must preserve or strengthen all applicable existing gates:

- formatting and tidy checks;
- Go safety policy;
- `go vet`;
- unit and integration tests;
- race detection;
- exact production-package coverage;
- meaningful mutation testing with exact kill requirements;
- fuzz targets and bounded fuzz smoke checks;
- golangci-lint;
- staticcheck;
- NilAway as advisory unless policy is deliberately changed;
- govulncheck;
- CodeQL build coverage;
- Gitleaks and repository secret checks;
- dependency, license, and SBOM checks;
- documentation spelling and link checks;
- API compatibility baselines;
- conformance and interoperability suites;
- benchmarks and performance budgets;
- repository/manifests consistency;
- release dry-runs and clean-consumer verification;
- package-specific service-backed tests.

Create a capability matrix mapping every legacy script and every divergent
variant to its new CLI command or typed configuration. A legacy capability may
be removed only when it is proven obsolete, redundant, or invalid and the
decision is documented.

## Coverage And Mutation Requirements

The tooling repository itself must meet the same standards it enforces:

- exact 100% statement coverage for production packages;
- meaningful tests that prove behavior, not line execution;
- 100% killed non-equivalent mutants for production code;
- no denominator manipulation, broad exclusions, generated junk, or tests that
  merely mirror implementation;
- boundary, failure-path, hostile-input, cancellation, concurrency, and
  resource-cleanup tests;
- fuzzing for configuration parsing, manifests, evidence formats, archives,
  path handling, subprocess output, and untrusted repository content;
- race tests for concurrent orchestration and evidence access;
- deterministic golden fixtures for CLI output and configuration diagnostics;
- integration tests using disposable fixture repositories;
- tests proving paths cannot escape the repository or task-owned temporary
  roots;
- tests proving secrets are redacted from output and artifacts.

Equivalent or unreachable mutants must be reviewed individually and represented
through a strict, minimal, machine-readable policy. They must not become a
percentage-based escape hatch.

## Evidence Model

Verification evidence belongs to the repository whose source it proves, while
the evidence implementation and schemas belong to `go-library-tools`.

Design a documented, versioned evidence format with these properties:

- content-addressed package and gate input identity;
- tool and verifier semantic identity;
- report digest and execution environment metadata where material;
- atomic writes and immediate persistence when evidence is produced;
- no dependence on Git commit hashes, branch names, repository history, or tag
  identity when exact source content is unchanged;
- history rewrites and force pushes do not invalidate content-identical proof;
- code, tests, configuration, tool semantics, or relevant fixture changes do
  invalidate affected proof;
- one package changing does not invalidate unrelated packages;
- evidence reuse is explicit, inspectable, and fail-closed;
- corrupted, partial, stale, forged, or mismatched evidence is rejected;
- concurrent runs cannot overwrite or cross-contaminate each other;
- CI artifacts remain attributable to one repository, module, package, gate,
  and input identity;
- no global cache is required for correctness.

Preserve valid existing mutation checkpoints by exact content and verifier
identity. Do not rerun every package merely because history changed or tooling
was relocated. Conversely, do not migrate evidence when semantic equivalence
cannot be proven.

## Service Fixtures

Centralize generic lifecycle implementations for supported fixture types, but
keep package-specific topology, schemas, certificates, compatibility payloads,
and interoperability data in the consuming repository.

Required characteristics:

- unique task-owned names, ports, networks, volumes, and temporary paths;
- parallel-safe execution in one working tree and across repositories;
- deterministic readiness probes;
- explicit environment export to the child gate only;
- cleanup after success, failure, cancellation, and interruption;
- no broad Docker cleanup or interaction with user-owned containers;
- support for the service types actually used by the repositories;
- typed fixture configuration instead of arbitrary scripts wherever possible;
- a narrowly reviewed extension boundary only when a generic fixture cannot
  represent a real package requirement.

Do not run process-control, descendant-cleanup, Docker-impacting, or other
system-risk tests locally during agent execution. Such behavior must be tested
through isolated unit boundaries or CI environments designed for it.

## Reusable GitHub Actions

Provide a reusable workflow such as:

```yaml
jobs:
  ci:
    uses: faustbrian/go-library-tools/.github/workflows/library-ci.yml@<immutable-sha>
    with:
      config: .golib.yaml
```

The exact caller may include required permissions and secrets, but it should be
small, uniform, and package-neutral.

Requirements:

- Consumers must pin reusable workflows and setup actions to immutable commit
  SHAs, with a nearby release comment.
- The installed CLI must be a released, immutable artifact whose version is
  declared by repository configuration.
- Verify downloaded binary checksums and release provenance before execution.
- Do not use `curl | sh`, mutable branches, floating tags, or unverified
  binaries.
- Support Linux amd64 and arm64 at minimum; add macOS amd64 and arm64 for local
  release artifacts.
- Keep GitHub token permissions least-privileged per job.
- Retain a stable final required job suitable for branch protection.
- Preserve module matrices, release dry-run mode, CodeQL, evidence upload on
  failure, scheduled checks, and explicit concurrency behavior.
- Central workflow updates must be rolled out through automated pull requests
  that update the immutable workflow SHA and tool version deliberately.
- A tooling release must not silently change consumers pinned to an earlier
  release.
- The tooling repository must avoid circular bootstrap. Define and test how its
  own CI builds and verifies the current CLI before a prior release exists and
  how it dogfoods stable releases afterward.

## Distribution And Supply-Chain Security

Publish `golib` as a normal open-source Go project and signed release artifact.

Implement:

- reproducible release builds where practical;
- checksums for every binary archive;
- provenance/attestation through GitHub Actions;
- SBOM generation;
- pinned GitHub Actions by immutable SHA;
- dependency review, Dependabot/Renovate policy, CodeQL, Gitleaks,
  govulncheck, and license checks;
- release manifests that bind source, artifacts, checksums, and provenance;
- documented verification instructions for local and CI installation;
- no committed secrets, credentials, or machine-specific paths.

The first public tooling release must be `v1.0.0`, not a pre-release. Do not
publish it until the implementation, hardening, documentation, compatibility
tests, and migration rehearsal are green. Later fixes follow Semantic
Versioning.

## Documentation

Create concise, human-oriented documentation:

- a lean README explaining purpose, installation, a minimal example, guarantees,
  documentation links, compatibility, support, and license;
- `docs/README.md` as the navigation entry point;
- architecture and trust boundaries;
- complete CLI reference;
- complete configuration reference and schema;
- local-development guide;
- reusable-workflow guide;
- migration guide from copied `.golib` tooling;
- gate behavior and failure semantics;
- evidence and mutation-checkpoint model;
- service-fixture lifecycle and safety;
- release and upgrade process;
- troubleshooting and FAQ;
- security policy, threat model, and supported versions;
- contributor guide and code of conduct;
- examples for root-only, multi-module, service-backed, conformance-heavy, and
  mutation-heavy repositories.

Avoid internal execution language such as “goal,” “hardening prompt,” “release
claim,” “proof matrix,” or “adoption and tradeoffs” in user-facing docs. Write
for maintainers and users of the tool.

## Repository Setup

Initialize `go-library-tools` as a complete standalone OSS repository with:

- current stable Go version selected from authoritative current evidence;
- `go.mod`, `go.sum`, `.go-version`, and explicit tool pins;
- MIT `LICENSE` and accurate third-party notices;
- `README.md`, `docs/README.md`, `CHANGELOG.md`, `SECURITY.md`, `SUPPORT.md`,
  `CONTRIBUTING.md`, `CODE_OF_CONDUCT.md`, and `DEPRECATION.md`;
- `AGENTS.md` and `CLAUDE.md` with the same strict engineering, disposable
  resource, evidence, changelog, coverage, mutation, fuzzing, race, security,
  and commit discipline used by the libraries;
- one cohesive CI workflow plus reusable consumer workflow and setup action;
- repository description and relevant GitHub topics;
- branch protection/ruleset recommendations documented and applied when
  authorized;
- release automation for `v1.0.0` and later versions;
- no generated or transient artifacts committed without a durable consumer.

## Migration Inventory

Before implementation, inventory every repository under
`/Users/brian/Developer/golib` and classify:

- active public library;
- active internal/reference module;
- empty or intentionally deferred repository;
- dirty repository with concurrent work;
- root-only or multi-module;
- current `.golib` fingerprint and divergence family;
- required services and external runtimes;
- package-specific scripts or fixtures;
- API and mutation evidence;
- workflow jobs and required checks;
- current Makefile entry points;
- release state and module tag prefixes.

Do not commit a transient migration audit to final library repositories. Keep a
working migration matrix in task-owned temporary storage while executing and
put only durable design decisions or reusable migration guidance in
`go-library-tools` documentation.

`go-secret-store` has historically been intentionally deferred and may still
be empty. Reconfirm its state. An empty repository cannot consume the tool; list
it explicitly as not applicable rather than pretending it was migrated.

`go-rabbitmq-queues` may contain concurrent specialist work. Migrate its
repository tooling without staging, rewriting, or coupling to those source
changes. If tooling validation cannot be separated from active behavior changes,
record the exact bounded dependency and finish every independent migration step.

## Migration Strategy

### Phase 1: Inventory and behavioral contract

1. Hash and classify every distinct `.golib` file across all repositories.
2. Identify common behavior, repository-specific policy, genuine fixtures,
   obsolete code, and accidental divergence.
3. Build a legacy-to-new capability matrix.
4. Capture characterization fixtures for important legacy behavior before
   replacing it.
5. Define the canonical configuration, evidence, CLI, and workflow contracts.

### Phase 2: Implement and harden `golib`

1. Implement configuration and inventory validation first.
2. Port gates in coherent vertical slices with tests.
3. Implement content-addressed evidence and mutation checkpoint migration.
4. Implement service fixture orchestration behind safe interfaces.
5. Implement release and clean-consumer verification.
6. Implement the reusable workflow and setup action.
7. Add documentation, security review, benchmarks, and release artifacts.
8. Prove exact coverage and mutation requirements.

### Phase 3: Parity rehearsal

Select representative repositories covering:

- a simple root-only pure library;
- a large root module with many packages;
- a genuine multi-module repository;
- PostgreSQL and Valkey service-backed tests;
- RabbitMQ, Kafka, or other broker-backed tests;
- specification/conformance suites;
- external runtimes such as Node, Deno, Java, or zsh;
- nested independently releasable modules;
- large mutation inventories and zero-mutant packages.

Run legacy and new tooling against content-identical source in isolated
environments. Compare selected modules, commands, exit states, coverage
denominators, mutation inventories, skipped/advisory semantics, generated
evidence, service lifecycle, and release decisions. Differences require an
explicit documented resolution; “the new command passed” is not parity proof.

### Phase 4: Release tooling `v1.0.0`

Release only after all required checks and representative parity rehearsals
pass. Record immutable workflow/action SHAs and artifact checksums used by
consumer migrations.

### Phase 5: Migrate every repository

For each applicable repository:

1. Verify repository root, branch, status, and concurrent work.
2. Add strict `.golib.yaml` with no duplicated manifest facts.
3. Move only source-specific evidence and package fixtures to durable,
   human-named locations.
4. Replace the Makefile with thin calls to the pinned `golib` CLI.
5. Replace CI with the pinned reusable workflow.
6. Update Dependabot/Renovate configuration for tooling updates.
7. Remove copied `.golib` implementations completely.
8. Remove obsolete workflow code, package-manager tooling, lockfiles, scripts,
   and tool pins that existed only for copied tooling.
9. Preserve package-specific fixtures and external-runtime dependencies that
   remain genuinely necessary.
10. Update changelog and contributor documentation.
11. Run repository contract, documentation, focused gate parity, and the full
    required CI-equivalent command through the new CLI.
12. Commit the repository as one coherent migration batch.

Migrate in bounded cohorts. Do not modify all repositories first and verify at
the end. A cohort must be green and reviewed before the next cohort begins.

### Phase 6: Remove compatibility paths

After every applicable repository is migrated:

- remove temporary shadow-mode adapters and legacy command aliases;
- prove no Makefile, workflow, documentation, or script references `.golib`;
- prove no consumer repository contains copied shared-tool implementation;
- prove all workflow references use the intended immutable tooling release;
- prove package-specific evidence and fixtures remain reachable;
- run a final ecosystem-wide audit from fresh checkouts where practical.

## Update Automation

Create a supported mechanism that proposes tooling upgrades across all
repositories without directly rewriting them in place.

It must:

- discover repositories from GitHub or an explicit maintained inventory;
- update tool version, checksum, and immutable workflow/action SHA together;
- validate configuration migration before opening a change;
- never lower policy or rewrite package configuration unrelated to the upgrade;
- produce reviewable per-repository pull requests or bounded cohorts;
- report repositories that cannot migrate automatically;
- support dry-run output;
- never force-push or bypass branch protection.

Do not make normal consumer CI depend on a live inventory of sibling
repositories.

## Performance

Benchmark the new tool against the current script implementation using
content-identical representative repositories.

Measure:

- startup latency;
- repository inventory latency;
- no-op/checkpoint-reuse latency;
- module scheduling overhead;
- peak RSS;
- artifact size;
- isolated cache behavior;
- service startup overhead;
- concurrent module execution scaling.

The tool should materially reduce orchestration overhead and duplicated files,
but correctness, isolation, and transparent gate behavior take precedence over
microbenchmark wins. Publish methodology and raw reproducible results; do not
construct unequal comparisons.

## Security Review

Threat-model at least:

- untrusted repository files and prompt-like text;
- malicious paths, symlinks, archives, and traversal;
- command and argument injection;
- environment poisoning;
- secret leakage through logs, artifacts, process arguments, or fixtures;
- compromised release artifacts or mutable workflow references;
- forged or stale verification evidence;
- cache poisoning and cross-repository contamination;
- race conditions in concurrent evidence writes;
- unsafe Docker or process cleanup;
- pull-request code executing with privileged tokens;
- artifact upload of credentials or private data;
- denial of service through unbounded manifests, output, files, or subprocesses.

Resolve all confirmed high and critical findings. Document residual risks and
trust boundaries without calling them “release claims.”

## Commit And Delivery Discipline

- Commit coherent, verified batches regularly in `go-library-tools` and each
  migrated repository.
- Stage explicit paths only.
- Maintain every affected changelog.
- Do not amend commits or rewrite unrelated history.
- Do not overwrite dirty files or silently absorb another agent's changes.
- Do not force-push.
- Obtain any approval required by repository rules immediately before pushing
  protected branches.
- Do not publish `v1.0.0` or change branch protection until the required local
  authority and remote checks are satisfied.
- Report local, committed, pushed, CI-verified, released, and migrated states
  separately.

## Completion Evidence

Completion requires current evidence for every item below:

### Tooling repository

- `go-library-tools` is a complete OSS repository, not only a plan.
- All production code has exact meaningful 100% coverage.
- All non-equivalent production mutants are killed.
- Race, fuzz, lint, static analysis, vulnerability, CodeQL, secret, license,
  SBOM, documentation, API, integration, and release checks pass.
- Configuration schema, CLI, reusable workflow, setup action, evidence format,
  service orchestration, and release process are documented and tested.
- `v1.0.0` artifacts, checksums, provenance, and immutable GitHub references
  exist when release authority has been granted.

### Consumer repositories

- Every active applicable repository is inventoried and migrated.
- Every empty/deferred repository is explicitly accounted for.
- No repository-owned runtime/public behavior changed as part of tooling
  migration.
- No migrated repository contains `.golib` or copied shared tooling.
- Every migrated repository has one strict configuration file and a thin
  Makefile.
- Every migrated repository uses the pinned reusable workflow.
- Every repository-specific fixture and verification artifact has a clear,
  durable owner and location.
- All local CI-equivalent gates pass through the new CLI.
- Remote CI passes for every pushed migration when push authority is available.
- Required branch-protection job names remain stable or are deliberately
  migrated.
- Changelogs and contributor documentation describe the new commands.

### Ecosystem audit

Run and retain a concise final report proving:

- repository count and classification;
- migration status per repository;
- zero live `.golib` implementation trees;
- zero legacy `.golib` command references;
- zero mutable workflow or action references;
- one canonical tool/config schema version per repository;
- no root-only `go.work` files introduced;
- no shared/global caches required by agent or CI execution;
- no leaked disposable services, caches, artifacts, or temporary files;
- no uncommitted migration changes except explicitly identified concurrent
  third-party work;
- no unresolved parity, security, documentation, or verification findings.

Do not declare completion from absence searches alone. Confirm the new commands
and workflows actually exercise each repository's required gates.

## Final Cleanup

Once implementation and migration are fully proven:

- remove this completed `.ai/GOAL.md` from the release tree; Git history retains
  it;
- remove transient migration matrices and rehearsal artifacts;
- retain only durable user, contributor, security, operation, and architecture
  documentation;
- leave every repository clean except explicitly preserved concurrent work.

## Final Report

Report concisely:

- tooling release and immutable references;
- repositories migrated, excluded, or blocked with exact reasons;
- legacy file count and disk duplication removed;
- verification and CI outcomes;
- any preserved concurrent work;
- any genuinely remaining user decision or external blocker.

Do not describe the project as complete if any applicable repository still
executes copied `.golib` tooling or if the shared tool has not passed its own
strict release gates.
