# Selected library assurance

Ordinary `library-ci.yml` remains bounded local verification. A daily ordinary
schedule without a committed change range runs deterministic tests and saved
fuzz seeds, not active exploration. The separate `library-assurance.yml` caller
has no pull-request or push trigger and does not replace Required, advisory
refresh, owner full operations, or release qualification.

## Pilot ownership and cadence

The reusable guard admits the 50 mapped wholly in-process library owners so
rollout does not require another central Tools edit for each small caller batch.
Admission is eligibility, not a schedule or qualification claim: only the three
pilots below currently have delivered and hosted-qualified callers. Each new
caller needs an owner-selected cadence, exact producer pin and relevant native
qualification before it counts as adopted. The 33 service/tag/executable or
nested-contract owners retain their full/manual routes. Low-risk owners may
retain seeds/advisory plus manual exploration rather than automatic campaigns.
Repeated names in distinct packages, including Geo, Retry, Validation, Verkle
and Wire, are separate declared identities. Eligibility still does not certify
their native targets or schedule a caller.

| Owner | Weekly UTC slot | Selected risk and native package |
| --- | --- | --- |
| Clock | Monday 03:41 | Manual clock lifecycle, duration and callback cancellation/work limits; `./manual` |
| Concurrency Limit | Wednesday 03:41 | Finite bounded limit histories and single permit outcomes; `.` |
| HTTP Client | Friday 03:41 | URL/header/authentication validation, redirect credential boundaries, retry/error classification; `.` |

Each slot rotates one current declared package/selector pair, sorted by package
then selector and indexed by UTC week. Unchanged source does not suppress a fresh stochastic
sample. Three, two and seven targets respectively retain their original full
operations and budgets; this lane neither replaces nor certifies those full
campaigns. HTTP tests retain their native loopback-only/no-public-network
contract and leak checks. No broker, cloud, Docker or credential certification
is implied by these in-process pilots.

Manual diagnostics run only on main and require an exact declared identity.
A bare selector such as `^FuzzLifecycleSequences$` remains valid only when it
names one declared package. Use `./jsonwire::^FuzzDecode$` (or `.::^FuzzDecode$`
for root) to distinguish the same selector in different packages. The combined
input is resolved against declarations, never forwarded to Go. Empty, unknown,
malformed or ambiguous selectors fail; repeated declarations of the same pair,
including conflicting race settings, also fail. Pull-request/push events and
non-main refs fail before a campaign. Manual diagnostics are not release or
full-contract substitutes.

## Authority and bounds

The immutable reusable-workflow source owns `tools/selected-assurance.py`.
It installs the repository-declared public CLI through the existing checksum
and attestation-verifying setup action, without a source build. The installer
uses only the ephemeral read-only GitHub token; no owner secrets or persisted
checkout credentials are supplied. Native commands receive a sanitized
environment without tokens or service credentials.

The runner consumes existing `golib config show --json` and `golib inventory
--json` authority. It derives one anchored native target and its singular
package from the selected module's typed fuzz operation; no target catalog or
scenario JSON is copied. The default module is root. Unknown module/package
ownership, disabled fuzz contracts, absent native target discovery, service/test-tag/executable-runtime
requirements and unsupported legacy steps refuse this in-process route. Such
owners retain their relevant full/manual assurance until a separately selected
service-capable route is qualified. A new owner is not scheduled merely because
it has a fuzz declaration: its risk and cadence must be selected first.
Informational Go dependency entries in `external_runtime_dependencies` are
retained; executable prerequisites come from the typed `runtimes` policy.
An explicitly declared fuzz step's race setting is retained; otherwise this
lane does not manufacture a race campaign. Native module resolution is read-only.
Production package admission retains the inventory contract. A policy-declared
test-only package, such as Barcode's root encoder fuzz harness, may be absent
from production inventory only when bounded filesystem discovery finds regular
native test files and no production Go files. This fallback refuses missing/empty
directories, symbolic module/package/source paths and nested-module crossings; native discovery
must still resolve exactly one selected target. No production package records or
runtime qualification are manufactured for this exception.

The job has a ten-minute deadline, including checkouts, installation, campaign,
artifact handling and cleanup. Installation is capped at three minutes. One
private cache session bounds identity/policy discovery, instrumented target
discovery and campaign to five minutes combined. Active exploration is 30
seconds with two workers, `GOMAXPROCS=2` and `GOFLAGS=-p=1`; hosted Linux children
inherit two-CPU affinity. Local platforms without affinity retain process,
worker and compiler admission but do not prove the hosted CPU isolation.
Aggregate owned-process resident memory is sampled during every native
command; exceeding 2 GiB fails and terminates that process group. Combined
native output is capped at 1 MiB. Timeout, cancellation, test failure, output,
memory, discovery and cleanup failures remain failures; there is no retry.

Failures identify owner, immutable source, stage and selector. Native payloads
are not copied into logs. A newly emitted native failure corpus may be uploaded
for seven days, bounded to 32 files/2 MiB with native content-addressed names.
Missing corpus is not passing evidence. Build/module/temp caches and owned
process groups are cleaned on success and failure; the workflow always removes
the exact run/attempt cache, installer checkout and diagnostic slots, including
after an interrupted runner. Artifact retention preserves a native reproducer,
not a second report or checkpoint engine.

## Verification boundary

Workflow execution fixtures double only public CLI discovery and slow native
Go execution, and exercise actual selection/refusal/failure/cleanup behavior.
They do not establish native fuzz target behavior or hosted scheduling. Selected
native pilot runs and complete independent review qualify the changed boundary;
an actual immutable Tools commit must replace every caller's preparation
placeholder before those callers can be delivered. No new CLI or library
release is required.
