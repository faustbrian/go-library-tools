# Engineering Policy

The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT",
"SHOULD", "SHOULD NOT", "RECOMMENDED", "NOT RECOMMENDED", "MAY", and
"OPTIONAL" are interpreted as described in BCP 14 when capitalized.

## Contract

- This file is the canonical repository policy. `CLAUDE.md` MUST point here.
- Shared behavior belongs in typed Go code. Repository-specific facts belong
  in strict configuration, manifests, fixtures, or evidence.
- Arbitrary package-owned shell hooks, hidden global state, and compatibility
  fallbacks to copied `.golib` implementations are forbidden.
- Public CLI, schema, workflow, and evidence changes MUST be documented and
  represented in `CHANGELOG.md`.

## Resources

- Every Go command MUST use a task-owned disposable `GOCACHE`, `GOMODCACHE`,
  and `GOTMPDIR`; all task resources MUST be removed after use.
- Containers, networks, volumes, ports, credentials, and temporary files MUST
  be uniquely task-owned, bounded, and cleaned on success and failure.
- Cleanup MUST resolve exact owned resources and MUST NOT use broad Docker,
  process, cache, or filesystem cleanup.
- Agents MUST NOT run Docker-impacting, process-control, descendant-cleanup,
  or comparable system-risk tests locally.

## Design And Security

- Prefer standard library behavior, explicit composition, immutable data, and
  narrow interfaces over reflection, registration magic, or service locators.
- Untrusted paths, YAML, JSON, archives, process output, and environment input
  MUST be bounded and validated before allocation, execution, or persistence.
- External commands MUST use explicit argument arrays without shell
  interpretation. Secrets MUST NOT enter arguments, logs, errors, artifacts,
  fixtures, or committed configuration.
- Concurrent work MUST have explicit ownership, cancellation, synchronization,
  and cleanup. Locks MUST NOT cover callbacks, network IO, or unbounded work.

## Verification

- Behavioral work MUST establish an observable contract and meaningful test
  before implementation. Tests MUST cover outcomes, failure paths, hostile
  input, cancellation, concurrency, cleanup, and compatibility as applicable.
- Routine changes MUST use the least expensive current evidence that directly
  covers their material risk. Documentation and metadata changes require
  structural checks; internal behavior requires focused package tests and
  applicable static analysis; public, security, persistence, or concurrency
  changes require their directly affected integrations and consumers.
- Fast local and pull-request checks MUST remain separate from aggregate
  milestone, scheduled, and release checks. A routine pull request MUST NOT be
  blocked by unrelated mutation, fuzz, race, benchmark, external-service,
  release-rehearsal, or ecosystem-wide work.
- Exact per-production-package coverage and complete viable-mutant accounting
  are available aggregate controls when an identified risk or release boundary
  requires them. Equivalent or unreachable mutants require narrow reviewed
  machine-readable records, never percentage exceptions.
- Parsers and untrusted boundaries MUST be fuzzed, concurrent code MUST pass
  race tests, and performance claims require equivalent reproducible
  benchmarks only when those checks exercise a material risk of the change.
- NilAway is advisory; all other required quality, security, documentation,
  compatibility, and release gates fail closed.
- Completion claims require fresh evidence for affected behavior and a final
  review of the complete change.

## Evidence

- Evidence identity MUST include every behavior-affecting source, test,
  configuration, fixture, service, verifier, and tool input.
- Git commits, branches, tags, and history shape MAY be metadata but MUST NOT
  invalidate content-identical evidence.
- Completed evidence MUST be persisted atomically as soon as it is received.
  Unrelated changes MUST NOT trigger broad reruns.
- Missing, partial, corrupt, forged, ambiguous, or mismatched evidence MUST be
  rejected.

## Git And Delivery

- Preserve unrelated and concurrent work. Stage explicit paths only.
- Commit coherent verified batches with conventional messages. Do not amend,
  force-push, rewrite history, or bypass hooks.
- Do not push, publish, tag, or change repository settings without authority.
- The first public release is `v1.0.0`. It requires complete local gates,
  review, remote CI, checksums, SBOM, and provenance.
