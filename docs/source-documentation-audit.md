# Source documentation audit

## Scope and acceptance

This Tier A source-comment pass owns baseline
`0e8ff6fee42e88858069a6f6ae6f1080d95c5eae` and does not change release,
workflow, dependency or policy inputs. Reproduce the inventory with
`git ls-files '*.go' '*go.mod'` and `golib source-docs inventory --json`:
323 tracked Go files, 86 source and 237 test files, in 23 directories. The
root CLI module is the product; `release/compatibility-consumer` is a generated
cross-library fixture, not another implementation or newly qualified consumer.
Seven generated files are excluded from hand editing. Baseline objective
findings are 357 exported declaration/member gaps and two package-entry gaps.

| Contract slice | Audit boundary and named risk |
| --- | --- |
| Configuration/inventory/cohesion | Strict bounded decoding, immutable source locks, package/module classification, deterministic reports and fail-closed applicable aggregation. |
| Repository/file admission | Lexical containment versus descriptor/root semantics, symlink and identity checks, stable-filesystem assumptions and cooperative I/O cancellation. |
| Process/gates/services | Command/environment/output ownership, completion versus cancellation races, platform termination capabilities, scanner admission ordering and cleanup. |
| Mutation/evidence/coverage | Execution versus checkpoint identity, source/tool/report provenance, atomic publication and rollback windows, diagnostic privacy and partial-result refusal. |
| Release/upgrade/CLI | Security-before-rehearsal, local versus public consumption, immutable publication inputs, per-file replacement versus coordinated transaction, output/error responsibility. |
| Documentation/specification/source audit | Bounded traversal and parser input, generated ownership, pinned source/decision history, manual semantic review versus objective structural enforcement. |
| Tests/variants/fixtures | Hostile-input and lifecycle seams, fuzz input limits/invariants, scanner/verifier/release tags, Darwin/Linux versus unsupported-platform behavior, preserved embedded executable fixtures. |

Missing declaration/member documentation is an objective repair list, not a
prose score. Non-obvious process lifetime, mutation reuse and publication claims
require source/test tracing before comments are written. No claim is inferred
from passing counts or historical goal status. Package entries must distinguish
borrowed caller resources from owned command/session state and avoid blanket
cancellation, privacy or public-consumer guarantees.

## Native validation and generation

Use formatting, all-tracked AST/token/literal/directive preservation, rendered
package/API inspection, current source/local documentation and applicable
specification checks, pinned changed-content secrets and one ordinary complete
final-diff review. No runtime contracts or example inputs are changed; no
mutation, service, process-control, descendant cleanup or release rehearsal is
selected locally. Hosted-only system-risk tests remain hosted-only.

Tools root has API compatibility disabled and no configured full-prose snapshot.
Its internal Go-doc render is useful review, not a published importable API or
new release qualification. Preserve all module/version, dependency, schema,
fixture, workflow, template-string and historical authority bytes.

Local schema producers are `internal/cohesion/schema_generate_main.go`,
`internal/inventory/schema_generate_main.go` and
`internal/securitydocs/cmd/genschemas/main.go`. Existing package go-generate
directives select these tracked producers; generated headers and emitted bytes
are guarded by native generator tests. The compatibility fixture producer is
`tools/compatibility/generate.py`, with offline source regeneration documented
in CONTRIBUTING.md. Do not hand-edit generated outputs or execute generators
while auditing prose. Generic existing schema headers name the actual
`go generate` mechanism; changing emitted header strings is a producer-behavior
change, separate from this executable-preserving pass.

The coordinator owns release publication, fleet adoption and exact-source CI.
This source pass does not qualify a local proxy rehearsal as public consumption,
refresh frozen evidence, or claim an unchanged system-risk campaign passed.

## Traced constraints and preserved boundaries

- Legacy Checkpoint preserves ExecutionRevision for exact ledger matching;
  emitted content-addressed evidence deliberately omits Git identity.
- Evidence Store equivalence excludes CompletedAt, not all semantic metadata;
  immutable conflict refusal is distinct from authenticating claimed outcomes.
- Campaign validates constructed Workers as 1–64, verifies input/source again
  before publication and keeps earlier successful packages after later refusal.
  Native source identities include complete Go bytes, including prose and inactive
  variants; this audit does not alter that existing cache authority.
- Ordinary process execution and bounded original-process-group cancellation
  intentionally differ. Cancellation racing successful exit can still fail;
  detached descendants, blocked caller streams and indivisible I/O are not
  magically preempted. Runtime/verifier caches retain separate trust owners.
- Workflow/local-action security and traversal admission precede actionlint.
  Scanner current/history snapshots, byte/cardinality budgets and diagnostic
  projection remain intact; raw unrelated tool errors are not privacy-safe.
- ReleaseDryRun performs security admission/scanning before a local file-proxy
  rehearsal. GOSUMDB-off checkout resolution is not a public released consumer.
- Pin Apply uses individually atomic replacements with attempted rollback;
  rollback failure is joined and crash-atomic cross-file publication is absent.
- Lease.Close handles tracked containers/volumes/networks once and caches the
  first result. A canceled cleanup is not retried implicitly; fixture files and
  credential erasure remain caller-owned.

All tracked directories are classified in the inventory: CLI and generator
commands; archive/build identity; cohesion/configuration/inventory/consumer
owners; coverage/evidence; gates/services/mutation; repository/file admission;
release/security/specification validators; source comments and upgrades. Tagged
scannerintegration, verifierintegration and releaseintegration fixtures, Linux
historical permission behavior and Darwin/Linux process admission are included
in source review without executing hosted-only system-risk work. Embedded Go
fixture text, shell patches, JSON schemas/license assets, npm pins and generated
consumer imports remain executable/data authority, not prose rewritten here.

Marker occurrences in Go are the source-checker's regex/policy and negative
fixture strings, not outstanding maintained debt; their bytes are preserved.
Real fuzz declarations describe their actual bounds, including targets without
an explicit harness cap. Function-like text inside executable fixtures is not
counted as a separate Go declaration or manually edited.
