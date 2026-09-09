# Assurance Simplification

Go Library Tools separates evidence by the failure it is intended to detect.

## Retained

- Fast local and pull-request checks cover compilation, tests, static analysis,
  repository configuration, manifests, Cohesion metadata, and workflows.
- Focused package checks support routine internal changes.
- The complete all-enabled gate remains available for releases and changes
  whose material risks require mutation, race, fuzz, benchmark, conformance, or
  external-service evidence.
- Public release assets, final ecosystem source locks, and published
  compatibility receipts retain immutable identities at their trust boundary.
- Module-manifest schemas v1 and v2 remain supported. Schema v3 is an optional,
  repository-owned metadata contract and does not trigger a fleet migration.

## Simplified

- Main and scheduled milestones add consumer and compatibility validation
  without automatically running every expensive gate.
- The compatibility-set draft remains explicitly non-installable until its
  listed versions and scenarios are verified through public consumers.
- Evidence is reused by immutable input identity without retaining temporary
  workspaces or routine execution transcripts.

## Retired

The unpublished recursive authorization, review, freeze, control-schema,
forward-oracle, schema-provenance, and resolved-reference-graph chains were
removed. Their primary failure mode was inconsistency among locally generated
evidence about other locally generated evidence; focused schema tests and one
complete-diff review cover the maintained contracts directly.

The schema-v3 fleet migrator and host capability probe were also removed.
Neither had a current consumer after schema v3 became optional and
repository-local.
