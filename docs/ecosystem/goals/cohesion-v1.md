# Golib Cohesion Goal Contract

Goal ID: `golib-cohesion-v1`

This contract defines the consumer-visible outcome for making Golib a cohesive
ecosystem of independently adoptable Go libraries. It governs the Cohesion
program without turning the ecosystem into a framework or a lockstep release
train.

The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT",
"SHOULD", "SHOULD NOT", "RECOMMENDED", "NOT RECOMMENDED", "MAY", and
"OPTIONAL" in this document are to be interpreted as described in BCP 14
[RFC2119] [RFC8174] when, and only when, they appear in all capitals, as shown
here.

[RFC2119]: https://www.rfc-editor.org/rfc/rfc2119
[RFC8174]: https://www.rfc-editor.org/rfc/rfc8174

## Mission

Golib MUST provide one recognizable design language, predictable package
selection and interoperability, and explicit adoption paths while preserving
the API shape and independent release ownership appropriate to each domain.

A consumer who learns one Golib package SHOULD recognize the construction,
configuration, ownership, lifecycle, failure, cancellation, security,
observability, documentation, and integration conventions used by the others.
Equivalent concepts MUST follow the same convention unless domain or
compatibility evidence establishes a justified exception.

## Actors And Scope

This contract assigns obligations to these actors:

- The **Cohesion coordinator** owns the ecosystem inventory, frozen decisions,
  cross-repository ordering, aggregate verification, and residual-exception
  register.
- The **tooling owner** owns shared schemas, deterministic projections,
  validation, publication controls, and public contract identities in
  `go-library-tools`.
- Each **standalone repository owner** owns that repository's public API,
  metadata, documentation, evidence, migration safety, and independent release.
- A **composition owner** owns each supported multi-package recipe, including
  dependency direction, initialization order, failure behavior, and shutdown.
- **Downstream goal owners** consume frozen Cohesion outputs for documentation,
  compatibility, hardening, operational assurance, and release without
  silently changing this contract.

The coordinator and tooling MUST operate through explicit standalone
repository inputs. They MUST NOT treat sibling repositories as one source tree,
Go workspace, module, release unit, or mutable aggregate checkout.

## Consumer Outcome

The completed ecosystem MUST provide:

- a reviewed consumer-facing design language;
- stable package ownership and dependency-direction rules;
- predictable construction, configuration, validation, lifecycle, error,
  cancellation, resource, security, and observability semantics;
- target-oriented adapters with isolated optional dependencies;
- reviewed package families and problem-oriented selection guidance;
- separate consumer and engineering catalogs;
- runnable and tested public multi-package compositions;
- explicit known-good compatibility sets for independent releases;
- independently useful package documentation linked to the ecosystem index;
  and
- local and authoritative-CI enforcement that detects semantic drift without
  requiring superficial API sameness.

The consumer catalog MUST contain only installable libraries and adapters and
MUST answer what to install and why. The engineering catalog MUST retain the
libraries, tools, fixtures, examples, benchmarks, and interoperability evidence
needed to explain what exists and how it is verified.

## Non-Goals And Framework Prohibitions

Cohesion MUST NOT create or require:

- an umbrella module or dependency that imports the ecosystem;
- a mandatory Golib runtime, application object, kernel, or bootstrap;
- a service container, locator, facade system, or global registry;
- runtime discovery, hidden initialization, or mutable global configuration;
- a universal configuration object for unrelated domains;
- a generic contracts package without one semantic owner;
- a universal error type that erases domain failure information;
- a mandatory logger, telemetry backend, router, queue, database, transport,
  or resilience stack;
- framework inheritance, model binding, application magic, or implicit
  lifecycle ownership;
- identical APIs where domains require different ownership or outcome models;
  or
- synchronized module releases merely to publish an ecosystem view.

Documentation MUST NOT excuse a needlessly inconsistent API. An API rewrite
MUST NOT be used to solve a discoverability-only problem. Released public paths
MUST NOT be broken without an additive migration path, reverse-consumer audit,
repository-owned compatibility evidence, and the authorized removal boundary.

## Six-Phase Acceptance Contract

### Phase 1: Inventory And Decisions

The coordinator MUST remeasure every releasable module and nested adapter,
including public APIs, construction, configuration, ownership, lifecycle,
errors, cancellation, resources, observability, dependencies, consumers,
releases, documentation, supported environments, and migration risk.

The phase MUST:

1. classify every module by consumer responsibility and primary family;
2. distinguish consumer-facing libraries from engineering-only artifacts;
3. inventory current consumers, versions, import paths, names, and public
   documentation;
4. freeze construction, lifecycle, error, adapter, dependency, documentation,
   and package-selection decisions;
5. classify every divergence as justified, compatibility-bound, temporary,
   stale, unsafe, or unresolved; and
6. record every justified exception and unresolved maintainer decision.

No public rename or API normalization MAY begin before the applicable inventory
and decision are reviewed.

### Phase 2: Design-Language Foundation

The tooling owner MUST publish the consumer design language, closed catalog
metadata, deterministic repository projections, separate consumer and
engineering views, and a locally runnable cohesion check.

The foundation MUST:

1. define the nine reviewed package families and controlled secondary
   capabilities without creating import layers;
2. record responsibility, non-goals, construction and lifecycle styles,
   ownership, dependencies, adapters, companions, environments, documentation,
   and independent delivery dimensions;
3. preserve explicit standalone repository scope and reject implicit sibling
   discovery;
4. generate deterministic JSON as the canonical projection and deterministic
   Markdown from that projection;
5. bind published artifacts to immutable tooling and design-language identities
   and content digests; and
6. expose repository-local validation through the repository's existing
   checksum-pinned tooling and authoritative CI.

The initial foundation MUST NOT claim later compatibility-set, recipe,
semantic-source, cross-repository, or terminal delivery evidence that it has
not verified.

The nine family slugs, in catalog order, are `foundations`, `service-edge`,
`protocols-and-descriptions`, `persistence-and-durability`, `resilience`,
`observability`, `integration-and-data-movement`, `domain-utilities`, and
`tooling`. Adding a family, secondary capability, or another controlled enum
value requires a reviewed schema-version change; free-form aliases are invalid.

### Phase 3: Pre-v1 Remediation And Attribution

Each repository owner MUST resolve the reviewed Cohesion requirements that
apply to its modules and attribute implementation, hardening, and release state
without conflating those dimensions.

The phase MUST:

1. normalize inconsistent pre-v1 names and equivalent public concepts only
   where consumer, release, and domain evidence authorizes the change;
2. repair stale Go versions, badges, links, module paths, package identifiers,
   examples, and documentation entry points;
3. add missing nested-module documentation and ecosystem navigation;
4. preserve released behavior through additive compatibility and migration
   paths where required;
5. bind implementation and hardening terminal claims to immutable,
   dimension-specific evidence over their complete applicable inputs;
6. retain legacy manifest meaning during the approved schema transition; and
7. update API baselines, dependency manifests, changelogs, and migration notes
   atomically with each affected public change.

Evidence for one repository, module, goal identity, or delivery dimension MUST
NOT authorize another.

### Phase 4: Composition And Compatibility

The coordinator and composition owners MUST define supported package stacks and
prove them through public APIs and clean module resolution.

The phase MUST:

1. define required and optional modules, adapter ownership, dependency
   direction, construction order, middleware or policy order, transaction and
   acknowledgement boundaries, failure recovery, and shutdown order;
2. promote existing integration evidence into public executable recipes and
   add missing reference compositions;
3. verify service, worker, queue, database, event, workflow, search, ingestion,
   and observability compositions where applicable;
4. publish immutable known-good compatibility sets containing exact module,
   toolchain, platform, service, protocol, evidence, and release identities;
5. prove Track, Postal, and Location adoption against this same public design
   language; and
6. keep each compatibility set advisory and independently versioned rather
   than enforcing a hidden dependency bundle or lockstep release.

### Phase 5: Documentation Handoff

The coordinator MUST hand the frozen design language, taxonomy, package
selection, compositions, compatibility sets, remediation results, and residual
exceptions to the documentation goal.

The handoff MUST ensure that:

1. every implemented releasable module and adapter has an independently useful
   entry point with purpose, ownership, installation, quick start, package map,
   lifecycle, errors, security, compatibility, support, and ecosystem links as
   applicable;
2. planned modules remain visibly planned and do not claim installable or
   released behavior;
3. consumer guidance does not expose fixtures, harnesses, or internal tools as
   installable packages;
4. generated facts and reviewed recommendations remain distinguishable; and
5. documentation work does not reopen settled API decisions without a new
   reviewed contract decision.

### Phase 6: Final Verification

The coordinator MUST verify the complete final source set, not merely collect
repository claims.

Final verification MUST:

1. run cohesion validation and every affected module and reverse-dependent
   gate against exact authoritative revisions;
2. recompute applicable-input fingerprints and validate every selected
   implementation and hardening receipt;
3. compile and execute supported compositions through clean module resolution
   without workspace replacements;
4. verify documentation links, examples, package names, paths, Go versions,
   catalogs, compatibility sets, and target-consumer walkthroughs;
5. independently validate catalog source completeness, projection identity,
   deterministic rendering, and current bytes;
6. verify that no framework runtime, global registry, umbrella dependency,
   hidden initialization, or dependency cycle was introduced; and
7. record every remaining exception, caveat, and release blocker.

Completion requires documented walkthroughs for:

- a user adopting one standalone value or protocol package;
- a developer choosing among overlapping packages;
- a team creating a new HTTP or JSON-RPC service;
- a team creating a worker, ingester, processor, scheduler, or workflow;
- a Laravel or PHP team mapping familiar concerns without framework magic;
- an operator deploying PostgreSQL, Valkey, Kafka, OpenSearch, and telemetry;
- an open-source contributor adding a package or adapter; and
- a maintainer releasing an independently versioned compatible package set.

Each walkthrough MUST identify every point where the user must guess a package,
name, constructor, default, ownership rule, lifecycle method, error category,
integration order, compatible version, or next documentation page. Every
unjustified guess is an unresolved Cohesion defect.

## Completion Criteria

The Cohesion goal is complete only when:

- Golib has one reviewed consumer-facing design language;
- every releasable module is classified by family, responsibility, non-goals,
  construction style, lifecycle style, companions, and maturity;
- equivalent API concepts follow the same convention or have a documented
  justified exception;
- adapter and service-integration names follow one target-oriented scheme;
- every releasable module and nested adapter has a compliant entry-point
  README and versioned ecosystem navigation;
- stale Go-version claims, package-local workflow badges, standalone paths,
  and misleading status claims are eliminated;
- supported multi-package stacks have executable recipes and explicit
  initialization, ownership, failure, and shutdown contracts;
- independently released modules have published known-good compatibility
  sets;
- local and authoritative-CI cohesion validation detects objective drift;
- Track, Postal, and Location composition uses the same public conventions;
- the documentation goal can execute without inventing or silently changing
  API design decisions;
- no unjustified consumer guess identified by the walkthroughs remains; and
- Cohesion was achieved without creating a framework, umbrella dependency,
  service container, global runtime, or hidden magic.

## Evidence And Completion Semantics

Implementation, hardening, and release are independent delivery dimensions.
Implementation covers accepted remediation on the authoritative `main` branch.
Hardening covers all applicable final-input module, risk, and reverse-consumer
gates. Release reports whether a Cohesion-bearing version, tag, artifacts, and
clean external resolution have been verified.

Every delivery dimension uses these exact states and meanings:

| State | Meaning |
| --- | --- |
| `not-started` | The dimension applies, but scoped work or evidence has not begun. |
| `in-progress` | Work or evidence exists, but terminal proof is incomplete. |
| `blocked` | A matching blocker decision identifies an actual impasse. |
| `not-applicable` | A matching audited decision proves the dimension imposes no requirement. |
| `verified` | Exact immutable evidence proves module, goal, dimension, and final input. |

Terminal states require these exact evidence kinds:

| Dimension state | Required evidence kind |
| --- | --- |
| implementation `verified` | `source-acceptance` for implementation |
| hardening `verified` | `gate-attestation` for hardening |
| release `verified` | `release-attestation` for release |
| any dimension `blocked` | `blocked-decision` for that dimension |
| any dimension `not-applicable` | `not-applicable-decision` for that dimension |

The evidence-kind set is closed to `source-acceptance`, `gate-attestation`,
`release-attestation`, `blocked-decision`, and `not-applicable-decision`.
An evidence kind for one row MUST NOT satisfy another row. A `not-started` or
`in-progress` dimension is non-terminal and MUST NOT cite terminal evidence as
if it had completed.

Each evidence reference MUST bind one tracked repository-relative receipt path,
the receipt file's SHA-256, and exactly one entry ID. A terminal evidence
receipt MUST bind the goal ID and requirements digest, the exact repository and
module, exactly one dimension, source revision, evidence kind, artifact
identities and digests, verifier identity and checksum, exact Go toolchain,
gate-policy identity and digest, exact reverse-consumer revisions, and a
deterministic complete applicable-input fingerprint. It MUST also bind the
sorted per-file applicable-input manifest, its digest, and the reason each
input category is included. Evidence paths MUST be bounded, tracked,
repository-relative regular files and MUST NOT be symlinks. Declared and actual
digests MUST match.

Every `source-acceptance`, `blocked-decision`, and `not-applicable-decision`
receipt MUST bind a coordinator-authorized immutable decision or review record,
its digest, issuer and reviewer identities, and an accepted outcome. Final
aggregation MUST resolve and cross-check that record independently of the
receipt and projection.

The verifier and gate policy MUST use immutable identities and digests supplied
by the Cohesion coordinator through the canonical source lock and aggregate
inputs, independently of a receipt or projection being verified.
The dimension fingerprint MUST normalize the module manifest once, retaining
every non-delivery field, `cohesion.integration_roles`, the goal ID, and the
requirements digest while omitting only the derived goal status, the three
dimension statuses, and delivery evidence references. Receipt files MUST also
be excluded so no status or receipt can validate or invalidate itself.

A current terminal claim MUST be rejected when any applicable source, test,
dependency, build, toolchain, policy, API, specification, documentation, or
reverse-consumer input changes. A receipt MAY reference an ancestor revision
only when recomputation proves that the complete applicable-input fingerprint
is unchanged. Status fields and receipt files MUST NOT validate or invalidate
the evidence they report.

The implementation, hardening, and release dimension-state enum is exactly
`not-started`, `in-progress`, `blocked`, `not-applicable`, and `verified`.
Each schema-v3 module's goal-status enum is exactly `not-started`,
`in-progress`, `blocked`, `complete`, and `not-applicable`. Its goal object MUST
be non-null and match the canonical goal ID and requirements digest. The module
goal status MUST satisfy exactly one of these relations:

- `complete` if and only if implementation and hardening are each `verified`
  or `not-applicable`, and at least one is `verified`;
- `not-applicable` if and only if implementation and hardening are both
  `not-applicable` with matching decisions;
- `blocked` if and only if the goal is not complete and implementation or
  hardening is `blocked` with a matching decision;
- `not-started` if and only if the goal is neither complete nor blocked,
  implementation and hardening are each `not-started` or `not-applicable`, and
  at least one is `not-started`; and
- `in-progress` if and only if the goal is not complete, blocked, or
  not-started and at least one of implementation or hardening is `in-progress`
  or `verified`.

Release state MUST NOT affect this module goal-status derivation. In
particular, a blocked release MUST NOT make the module goal `blocked`.
Ecosystem-wide completion remains subject to every phase and completion
criterion in this contract; a terminal module goal does not establish it.

Documentation-only evidence MUST NOT be presented as proof of runtime,
compatibility, hardening, or release behavior. Expensive evidence MAY be reused
only when its complete applicable-input fingerprint is unchanged.

## Release Reporting Independence

Release state is reporting-only for Cohesion. A module's unreleased Cohesion
delta or recorded release blocker MUST NOT by itself prevent Cohesion
completion when implementation and hardening meet the completion rule.

A verified release state MUST bind the peeled source revision, Cohesion-bearing
module version and its exact tag derived through the manifest-declared module
identity and tag prefix, release-manifest and checksum-set digests, published
artifact digests, and exact clean-consumer input and result. A version or tag
name alone is insufficient. An older stable release MUST NOT prove a later
Cohesion change.

Every truthful non-terminal release state and every blocked release state MUST
remain visible in the residual/release-blocker register and be handed to the
ordered downstream release work. A later release receipt and catalog refresh
MAY report delivery without reopening or invalidating an already completed
Cohesion goal.

Each such register entry MUST identify the repository and module, the release
dimension, its exact current state, any blocking condition, the owner, and the
next action or removal condition.

## Residual Exceptions

Completion MUST include one explicit residual-exception register. Every entry
MUST identify its repository and module scope, consumer-visible difference,
classification, evidence, owner, review decision, and removal condition or
reason it is permanent.

A temporary exception MUST have an actionable expiry or migration condition.
A domain-specific permanent exception MUST explain the ownership, protocol,
failure, lifecycle, resource, or compatibility semantics that make uniformity
incorrect. A package MUST NOT differ merely because it was implemented by a
different contributor, agent, repository, or release cohort.

## Tool And Schema Delivery Handoff

The approved tool and schema delivery MUST preserve historical schema
identities and legacy decoded meaning while introducing the versioned Cohesion
delivery contract. Mixed-schema previews MUST remain non-terminal and MUST NOT
be published as final Cohesion catalogs.

The delivery MUST include regression evidence that:

- every new source-lock, delivery-evidence, manifest, catalog, and aggregation
  parser or other untrusted boundary is fuzzed through its schema, decoding,
  and semantic validation path; and
- final aggregation independently cross-checks the canonical source lock and
  input manifest for exact count, order, repository identities, immutable
  revisions, projection paths, and projection digests, while accepting only an
  exact projection bundle of referenced regular files with no duplicate or
  unreferenced members.

The final aggregate MUST recompute source-backed evidence from clean read-only
checkouts with the independently locked verifier and gate policy. It MUST NOT
establish truth by comparing values supplied by the same projection or by
trusting an internally consistent manifest and bundle without the independent
source-lock binding.

A final Cohesion catalog MUST be rejected unless every releasable module has
terminal implementation and hardening states. Release MAY remain non-terminal
or blocked and MUST remain visible under the release-reporting rules.

## Immutable Identity, Versioning, And Digests

The stable goal ID for this contract version is `golib-cohesion-v1`. The
requirements digest is the lowercase SHA-256 of the exact canonical bytes of
this file. Released tooling that validates this goal MUST embed both the exact
goal ID and requirements digest and MUST reject another ID or digest.

The canonical contract MUST receive an accepted independent review with no
unresolved findings before its source tag and digest are frozen. The freeze
record MUST bind the exact source commit, immutable semantic-version tag,
contract path, and requirements digest. The tag MUST peel to that source commit,
and the recorded requirements digest MUST equal the SHA-256 of the contract at
that path in the tagged commit. Public catalogs and evidence MUST cite that
immutable identity; a default-branch path alone is not a compatibility
contract.

After the freeze, these canonical bytes MUST NOT be rewritten under the same
goal ID. Any semantic change MUST publish a new versioned contract and stable
goal ID, preserve the prior contract and digest, define migration and
compatibility treatment, and receive a new independent review and freeze.

Generated catalogs, compatibility sets, receipts, and release reports MUST be
deterministic and digest-bound. Publishing a new artifact MUST NOT mutate,
reuse, or silently reinterpret an earlier immutable identifier.
