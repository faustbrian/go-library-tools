# Native risk custody development reports

Source base: f87c5148578c2fcb111fc71a418a750a72fc2afa.
Reviewed source tree: efbb5db8a4f878554f90aba0518a492170de06a0.
This checkpoint implements mapping, response custody and a migration source;
the original durable journal QA contract remains unfinished implementation work.
No package implementation worker was assigned. Primary /root owns all writes.

## Mapper — /root/risk_journal_map

**The RecordID boundary resolves the earlier exact-intent gap; no new coordinator API is required. One concrete incompatibility exists in the suggested encoding.**

At unchanged identity head `f87c5148578c2fcb111fc71a418a750a72fc2afa`:

- `postgres/codec.go:173` persists RecordID in original command scope.
- `postgres/query.go:204` compares exact RecordID **before** pending/terminal classification. `ObserveCommandAdmission` uses this same path.
- `postgres/recovery.go:80` repeats scoped `commandReceipt` checks during locked primary recovery; `postgres/recovery.go:108` uses retained original fingerprint and frozen plan.
- `risk/issuance.go:671` already binds CommandID, actor, tenant, subject, operation, canonical number, pre-auth transaction, attempt, policy, phase, purpose, all Subject dimensions and configured signal/carrier versions into the keyed pre-evaluation IntentDigest. Actual evaluated facts remain separately bound by signed ProofDigest.
- Required record creation's **0→1** transition is accepted by `encodeFeaturePlan` and `FeatureReceipt`.

**Encoding incompatibility:** `rei1.<version>.<digest>` passes general RecordID syntax but fails `postgres/work_audit.go:381`, which permits only ASCII letters/digits/hyphen/underscore, maximum 64 bytes. Feature admission uses RecordID as the audit target.

`rei1_<version>_<rawurl32digest>` meets that existing boundary: at most 59 bytes, permitted characters.

I withdraw the earlier missing-API conclusion for this record-target design. Remaining checks are implementation obligations: exhaustive signed-field validation must enforce the derived RecordID, retained-key lookup must stay bounded, native results must hide root record metadata, and committed replay still needs the encrypted original journal reference. This does not establish protected evidence consumption or full package completion.

## QA — /root/risk_journal_qa

## Native QA planning report

- Role: `qa_test_planner`; read-only, independent context.
- Reviewed base/head: `f87c5148578c2fcb111fc71a418a750a72fc2afa`.
- Repository: `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`.
- Scope: first durable **RiskEvidence issuance slice**, not complete `risk/postgres`.
- Skill: fully read `$qa-test-planner`.
- Sources: assigned goal, common requirements, transaction ownership/issuance rules, `risk/issuance.go`, coordinator admission and feature-execution contracts.
- No files changed; no tests executed.

### Observable acceptance contract

Only a successful owned, signed evaluation may produce evidence. The adapter must persist its exact scoped binding and envelope-encrypted original response inside the existing coordinator domain Work, together with required audit/outbox effects and committed command result. No reference escapes until commitment is proven.

Before providers run, the adapter must authorize and compare retained exact intent. Matching committed replay returns the original reference and timestamps, without evaluation or renewal. Unknown stays reference-free until the same command's primary-authoritative outcome resolves. Changed intent, phase, phone, actor or relevant configuration fails closed.

`ObserveCommandAdmission` alone proves neither retained intent equality nor protected result recovery. A metadata match is not replay authority.

### Smallest sufficient behavioral matrix

Use public issuer/executor composition against real PostgreSQL, real coordinator Work and actual migrations. Test controls may supply deterministic provider signals and create failures; they must not replace durable authority.

| Case | Concrete defect caught | Observable assertions |
|---|---|---|
| **Successful issuance, both phases** | Unsigned/caller-fabricated issuance; crossed purposes; exposing provisional evidence | Initiation/completion produce separate correctly bound rows. Returned committed reference authenticates through core reconstruction; `MaxUses=1`. Journal, command result and required audit/outbox records agree. No reference is available before committed outcome. |
| **Database time and original expiry** | App-clock timestamps or renewed source freshness | Stored issuance equals authoritative Work database UTC microsecond time. Expiry equals original signed expiry. Database time before preparation or exactly at/after exclusive expiry rejects issuance; rollback leaves no issued row. Test immediately-before boundary with a controlled, signed preparation window, not timing sleeps. |
| **Atomic failure** | Private transaction or partial commit | Fail after journal write and at audit/outbox/result staging. Issuance transaction leaves no issued evidence, success effects or committed result. Any durable failed-command bookkeeping remains safely classified. No reference returned. |
| **Pre-provider retained-intent gate** | Treating metadata-only admission as equality/absence | After a first command, replay with one changed dimension at a time: tenant, subject, actor, phase/purpose, phone binding, attempt/pre-auth/network binding and relevant configuration/policy version. Every mismatch returns safe failure/unknown with zero evidence and leaves stored state unchanged. Provider/counter work must be observable as forbidden on these paths. |
| **Committed replay and rotation** | Re-evaluation, replacement bearer, expiry extension, forgotten old intent key | Retry same command through a newly constructed adapter/issuer after advancing clock and adding a new retained key. Return byte-identical reference and original times, including historical expired evidence. Database still contains one evidence row and one committed effect set. Configure providers to fail if reached; replay nevertheless succeeds. |
| **Concurrent issuance/replay** | Duplicate journal rows, results or externally visible winners | Race two exact retries and two conflicting requests under one command ID using separate connections. Exact retries converge on one committed reference; mismatched request gets no reference and cannot overwrite original. Assert durable cardinality and binding, not merely callback counts. |
| **Unknown commit and primary reconciliation** | Guessing rollback from disconnect, minting new evidence or using replica absence | Exercise a real commit whose acknowledgement is lost, and a proved rollback alternative. First ambiguous response contains zero evidence. Same-command primary reconciliation recovers the exact committed reference when present, never reruns providers or generates another row; unresolved outcome remains unknown despite elapsed lease/expiry. Primary read failure cannot authorize fresh evaluation. |
| **Envelope, corruption, redaction and migration** | Plaintext response persistence, wrong-AAD recovery, silent security-state reset | Search relevant journal/result/effect columns for fixture bearer/phone/provider secrets; ciphertext must not equal plaintext. Swapping ciphertext between scope/command records or corrupting it fails closed without leaking diagnostic material. Install, populated upgrade and migration retry preserve pending intent and committed encrypted response, original expiry and key versions; no duplicate issuance or allowance reset. |

### Supporting narrow checks

- Pure adapter validation/encoding tests: invalid zero values, oversized inputs, unsupported configuration, signed-projection tampering and current policy/version mismatch.
- Redaction assertions: formatted values and returned errors never contain reference, phone/network/provider input, ciphertext-decryption diagnostics or keyed material.
- External-consumer compilation: a valid issuer/executor/contributor composition is constructible without private imports; invalid zero values reject.
- Structural ownership inspection: no private transaction, command ledger, coordinator, owner takeover or protected-consumption seam.
- Focused race run over issuance concurrency tests; focused module gate and changed shared-coordinator consumers if that public seam changes. Every Go invocation uses disposable `GOCACHE`.

### Deliberate scope exclusions

Reservation, consumption/apply/finalize, release, takeover, durable counters/lockout, cleanup/retention and production-scale query plans remain future full-unit obligations. Issuance tests must not claim those behaviors, `risk/postgres` verification, or end-to-end phone recovery.

### Planning disposition

One bounded plan complete. No contradictory requirement identified. The existing metadata-only admission limitation must be resolved at the authoritative coordinator boundary; a private intent ledger or inferred replay would violate the contract. No additional planning cycle or approval artifact is proposed.

## Regression — /root/risk_codec_regression

Initial native report:

**REQUEST_CHANGES — implemented security-boundary coverage gaps; no confirmed production-code defect.**

Reviewed `f87c5148578c2fcb111fc71a418a750a72fc2afa` plus current dirty `risk/postgres`, manifests, core issuance, and transaction requirements.

- **Cipher substitution coverage:** `cipher_test.go:40–48` checks only expiry and tenant changes. It does not prove rejection of same-tenant cross-command/subject/record swaps or altered intent, proof, lookup digest, key version, and original issue time. These are implemented authenticated inputs; losing their binding could survive the present tests. Add bounded negative substitutions, including initiation/completion response swapping.
- **Decoder malformed-field coverage:** `command_test.go:74–93` mutates values but not valid coordinator scalars with wrong native kinds or digest widths. Add wrong-kind cases and 31/33-byte digest/signature cases to exercise the implemented closed decoder.

Observed evidence is parent-reported race `-count=2` PASS, 82.0% coverage, vet PASS; latest manifest-pruned rerun pending. I ran no Go commands. Exact 100% coverage and release gates remain unproven.

Durable execution, atomicity, replay, one-use transitions, counters, retention, and real PostgreSQL remain outside this batch; their absence is not an additional finding here.

Final native report after additional negative cases:

**APPROVE — bounded regression audit.**

Reviewed current dirty `risk/postgres` and manifests on `f87c5148578c2fcb111fc71a418a750a72fc2afa`.

Both findings are resolved:
- Cipher tests now reject record-metadata substitutions, exact-command scope swaps, and cross-purpose binding.
- Decoder tests exercise wrong kinds for all eleven fields and invalid widths for all four binary fields.

No unresolved concrete regression finding in this batch. Latest race/coverage run remains pending; I performed source inspection only. Durable execution, PostgreSQL behavior, exact coverage, and release completion are not certified.

Primary subsequently observed terminal final current-source race/count2
PASS1.266s87.2%, vet PASS and staged diff check PASS. This updates execution
evidence only, not the independent source report's original wording.

## Go compatibility — /root/risk_codec_go

Role: independent `go_compatibility` auditor
Disposition: **APPROVE**, limited to the development-only staged batch.

Reviewed base/head `f87c5148578c2fcb111fc71a418a750a72fc2afa`, staged tree `efbb5db8a4f878554f90aba0518a492170de06a0`, all 11 staged files.

No concrete API, module-cycle, version, source-boundary, or ownership defect found:
- Existing root/risk APIs remain unchanged; new module consumes their public contracts.
- Mapper freezes signed projections using root command ownership; decoder delegates authentication to core.
- Migration implements the pinned migrations interface; its command foreign key matches existing kernel authority.
- Envelope custody uses the pinned implementation and copied local keyring.
- Module path, Go version, manifests and independent tag prefix agree; no permanent replacements exist.
- Archived root feature and risk issuance sources match inspected current sources.
- Staged whitespace validation passed.

No Go commands or services executed. Temporary workspace substitutions—including the older transitive identity requirement redirected to the newer archive—do **not** establish clean published-consumer compatibility. Checksums, API baseline, release gates and durable execution remain pending; this approval does not certify unit completion or release readiness.

## Correctness — /root/risk_codec_final

**APPROVE — bounded development batch. No findings.**

**Native reviewer:** `/root/risk_codec_final`, independent read-only single-reviewer pass using requesting-code-review, testing-coverage, and verification-before-completion. No delegation, code execution, service access, edits, or publication.

**Reviewed boundary:** repository `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`, branch `feature/identity-platform`, base/HEAD `f87c5148578c2fcb111fc71a418a750a72fc2afa`, complete 11-file staged tree `efbb5db8a4f878554f90aba0518a492170de06a0`. Every staged hunk inspected; no unstaged diff observed.

Source-backed closure:
- **Closed mapping/authentication:** `command.go:29–42` maps all signed projection members and rejects foreign command, tenant, subject, actor, policy version, and caller-selected record targets. `command.go:61–133` checks exact descriptor, cardinality, names, scalar kinds, digest widths, configured policy, and derived record identity before retained-key authentication. Upstream `feature_command.go:85–112` rejects duplicate names and freezes owned values; `risk/issuance.go:290–323` validates and authenticates complete prepared input. Neither structural mapping nor authenticated restoration grants durable authority.
- **Custody/scope:** `cipher.go:25–58` restricts wrapping configuration to a copied concrete local keyring. `cipher.go:61–114` binds the complete command binding, every retained digest, key version, and original timestamps. Secret-envelope archive `22fc4a2d925dfad2023437c9ef4ed862d501d4d0` confirms authenticated envelope/data-key handling, copied key ownership, cancellation admission, and concurrent-safe local providers. The source-derived maximum envelope is 437 bytes, within the 1,024-byte codec/schema ceiling.
- **Schema/ownership:** `migrations.go:27–84` returns one forward-only atomic installation, rejects unsafe schemas/version allocations, and defines scoped uniqueness and fixed-width custody constraints. Its command FK matches the existing key in `postgres/claims_schema.go:20–27`. New dedicated schema creation cannot overwrite existing managed or unmanaged objects. Existing reservation, replay, retention, and coordinator phases are not claimed implemented.
- **Oracles/consumer boundary:** tests obtain signed issues through real public core evaluation; fixtures replace only provider/counter boundaries. Tampering, wrong kinds/widths, purpose separation, ownership copying, ciphertext corruption, scope changes, freshness preservation, retained wrapping-key rotation, and cancellation assertions exercise the changed boundary. Exact-symbol fan-out found no additional production consumers or writers. Documentation and manifests accurately label the module experimental.

**Verification:** independently inspected coverage artifact and clean staged diff check. Parent-reported current-source `go test -race -count=2` passed in 1.266s with 87.2% coverage; `go vet` passed. I did not independently rerun Go commands.

**Claim limit:** approval supports the coherent local development commit only. Durable executor/shared-Work issuance, primary recovery, one-use transitions, counters/CAPTCHA, retention, real PostgreSQL execution, published dependency resolution/checksums, clean-consumer proof, exact coverage/mutation, and release gates remain pending. This is not unit verification, release readiness, or platform completion.

## Local checkpoint

Identity commit 2585d19c9bb7b5eafac88c7413629723a375fc60 has the exact reviewed
tree efbb5db8a4f878554f90aba0518a492170de06a0. No push or release performed.
Native independent roles used gpt-6.1-sol, medium, fork_turns none.

## Same-Work issuance checkpoint — f83ec70

Primary implementation: /root. Source baseline
2585d19c9bb7b5eafac88c7413629723a375fc60; final source commit
f83ec70e80e7ed983f7dcc24263b3b30bb8f5e70; exact reviewed tree
036c93a4fc6c1aa40c1eb88b453ff61485a8104c. Local only; no push/release.
Original single journal QA above remains unchanged. Classification: behavioral;
secondary public-contract/security/persistence/concurrency/retry/migration,
because the native participant writes encrypted durable evidence and participates
in atomic command execution and locked outcome reconciliation.

Final commands on current source: `go test -race -tags integration -count=2
-coverprofile=<task-owned>/coverage.out .` PASS1.699s85.7%;
`go vet -tags integration .` PASS; staged `git diff --check` PASS. Go1.27.1,
darwin/arm64, exact declared owned Git archives, isolated PostgreSQL18.6.
Constructor and migration absent-symbol compile REDs preceded source additions.
Mapping privacy RED and diagnostic redaction RED preceded corrections.
These establish the bounded native participant, not official executor,
authorized reference replay, complete module coverage or release verification.

### Mapper — /root/risk_work_mapping

Native report (base/head2585d19), read-only, no writes/execution/services or new
QA plan: no concrete requirement contradiction or missing coordinator API for
this issuance slice. Issuer resolves before counters/providers; only Missing
permits fresh evaluation. The shared feature path supports exactly four
participants, protected coordinator-owned audit/lifecycle/result and a native
0-to-1 receipt. Native Apply must authenticate signed projection, verify exact
active instance, materialize at database time and insert encrypted custody.
Commit uncertainty discloses nothing. Admission and recovery require current
authenticated caller composition, never manufacturing a principal from Actor.
Historical response decryption must use returned historical binding/AAD.
Recovery authorizes first, locks claim/journal and invokes frozen contributors;
only unanimous negative evidence permits abort. Positive/mismatched native row
never reconstructs success; expiry and independent pool absence prove nothing.
Both phone phases are distinct private purposes. Consumption/counters/CAPTCHA/
retention remain outside this slice. Original bounded QA remains applicable.

The mapper's original digest-derived RecordID pre-provider comparison is
superseded by the source-backed privacy correction below. Opaque command IDs
identify the aggregate, not exact private intent. Future committed resolution
must compare the private retained digest; prior pending/failed commands never
permit provider reevaluation.

### Regression — /root/risk_native_regression

Final native report:

**APPROVE — final staged tree `036c93a4fc6c1aa40c1eb88b453ff61485a8104c`.**

The corrected cipher oracle tests genuine cross-purpose substitution using
independently signed private metadata, rather than assuming purpose changes the
opaque command ID. The added `NewEvidenceExecution` assertion independently
rejects the original bearer for the other purpose. No production cipher change
or weakened security assertion found.

Parent-supplied final integration race/count-two PASS, vet PASS and staged-diff
check PASS assessed; no Go commands or services executed by this reviewer.

Prior executor, contention, ambiguity, exact-boundary and release-gate limits
remain. No whole-unit completion claim.

Earlier native reconciliation additionally approved real same-Work atomicity,
actual persisted lifecycle subject/type, additive migration history preservation
and unchanged legacy encrypted rows. It explicitly did not certify provider
suppression, authorized private-intent comparison or original-response replay.

### Go compatibility — /root/risk_native_go

Initial native disposition REQUEST_CHANGES: risk/postgres manifest omitted the
new integration test tag and four direct owned dependencies. Corrected before
final review; all nine owned direct dependencies and integration lane registered.

Final native report:

APPROVE — complete current dirty scope reviewed against HEAD
`2585d19c9bb7b5eafac88c7413629723a375fc60`, including the untracked issuance-ID
migration.

No unresolved Go compatibility or ownership findings. New aggregate IDs use
canonical opaque command IDs; signed private intent remains intact. The additive
migration preserves original installation SQL, legacy rows and ciphertext/AAD.

Compatibility boundary: the current native decoder rejects legacy digest-target
commands; retained storage is not historical response-resolution support. This
is acceptable for the stated experimental, unreleased durable profile and is
documented.

No tests or services executed by this reviewer. Approval covers this development
batch only; clean-consumer, whole-unit and release verification remain pending.

### Correctness — /root/risk_native_final

Initial native disposition REQUEST_CHANGES, reviewed tree
01d2b6ea9a9cd1eb4c04084796c72c3ad7495b70: MapIssue encoded IntentDigest in
RecordID, and coordinator lifecycle staging copied it to SubjectID despite the
public privacy promise. Original outbox seam did not observe SubjectID. Checked
mapping/staging/guards; none transformed it. This concrete finding prompted the
opaque-ID correction and actual staged-subject test, not a privacy exception.

Final native report:

**APPROVE — No findings for this intermediate development batch.**

Reviewed the complete 12-file staged diff against base
`2585d19c9bb7b5eafac88c7413629723a375fc60`, exact tree
`036c93a4fc6c1aa40c1eb88b453ff61485a8104c`, with no unstaged identity changes.
Control lifecycle specification remains at the previously reviewed hash.

The prior blocker is resolved: new record identities are opaque command IDs;
actual staged lifecycle subjects are persisted and asserted against that ID.
Private intent digests remain authenticated command/journal data. The forward
migration preserves original installation bytes, legacy rows and encryption AAD;
populated upgrade/retry coverage checks retained bytes and decryption.

Complete sweeps covered signed-field admission, invocation identity,
schema/serialization bounds, same-Work atomicity, failure prefixes,
uniqueness/replay, locked recovery, redaction, identity conservation, migration
compatibility and test oracles.

Supplied final-source verification: integration race/count-2 **PASS**, vet
**PASS**, staged diff check **PASS**; coverage **85.7%**. Reviewer performed
read-only source inspection, not execution or publication.

Approval permits a coherent local development commit only. Official executor
and authorized original-bearer resolution/replay, consumption, counters/CAPTCHA,
retention, production adapter interoperability, migration-scale evidence, exact
coverage/mutation and clean-consumer/release gates remain pending. No whole-unit,
release, deployment or production completion is certified.
