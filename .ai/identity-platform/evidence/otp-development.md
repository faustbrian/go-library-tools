# OTP development evidence

## Original Send input checkpoint — 2026-10-07

Source `5ff5eb8ddf0c10e8b00111b97f79d57a6c0e28e4`, base `0f74387`:
private code-free request framing and deterministic command mapping preserve
original session/pre-auth, policy, authority and times. Generated digests and
submitted bearers are absent from the record; possession remains separate.
Legacy issuance byte oracle is independent of the shared continuation helper.
No new public API, storage, journal, migration or permanent replacement added.

Native missing-API RED003c23/5dfed4; focused PASSa8ad5e; full adapter
`go test -race -count=1 .` PASS480360(1.332s); `go vet .` PASS1caa22;
`go test -run '^$' -fuzz '^FuzzSendRequestRecord$' -fuzztime=3s -parallel=2 .`
PASSb9b577(53,168 executions). Removing original-bearer digest comparisons
in a disposable overlay fails the intended foreign-token assertion9304e6.
Regression `/root/send_request_regression_0f74387` and complete final review
`/root/send_request_final_0f74387` APPROVE the exact ten-file batch.
Local selected-module workspace only; full clean-consumer/release/coverage/
mutation and actual public Send/PostgreSQL composition remain unproven.
No unit promotion:5/67 integrated,1/67 verified. No push.

## Prepared-owner invocation custody, source base 8ce205f

The kernel retains the existing authenticated preparation bytes and original
output generation in its reservation; exact final owner access returns owned
copies, not shared request state or a new wire profile. Native missing-API RED
c5f686 precedes final focused PostgreSQL/race PASS df6049 (3.329s), full kernel
unit PASS 37cf5c (0.342s), and final integration-tagged vet PASS 502326.
The real coordinator tests cover 17,000-byte output, independent concurrent
commands, terminal replay, foreign-owner denial and restart/takeover provenance.
They use a trusted test participant, not complete native public OTP Send.
Go1.27, disposable GOCACHE, pinned isolated PostgreSQL18.4 and temporary local
GOWORK overrides; published source pins/clean-consumer evidence remain pending.
Native regression/Go compatibility /root/prepared_output_regression_8ce205f
and complete final review /root/prepared_output_final_8ce205f APPROVE all six
unchanged source hashes. Local source commit0f74387f01b1f9e0cb9c75060e0e30a22cec647a
is clean. A copied overlay removing return cloning fails the intended runtime
assertion (`returned copy mutated retained custody`, d15f81, exit1); this is
focused defect detection, not full mutation coverage. Full Send risk,
recipient and original-request custody plus exact coverage/mutation/release
gates remain unfinished. Counts unchanged5/67 integrated,1/67 verified.

## Joined original Send-output restoration, 3ebd6d2

Source base8316d8af9fd74669a11b1bf3f38b87d52a3cfb38, local commit
3ebd6d204b0c513567365e96deaf67664f3cd036, exactly six otp/postgres files.
Existing four core OTP dirty files are excluded and preserved.
Fresh isolated gpt-6.1-sol medium native reports:
- /root/otp_send_output_regression_8316d8a: regression and Go compatibility
  APPROVE; native parent results inspected, no independent execution.
- /root/otp_send_output_final_8316d8a: complete six-file final-diff APPROVE,
  no findings. Both bind unchanged final file SHA256 identities.

Native parent transcript is the existing session path below. Final full
`go test -race -count=1 .` session91019 chunk1bf416 PASS1.359s;
`GOMAXPROCS=2 go test -run '^$' -fuzz '^FuzzSendOutputField$' -fuzztime=100x .`
chunk7cf6cd terminal exit0 PASS0.277s,100 executions. Copied overlay removing
the trailing-frame check, session55047 chunk54e595, exits1 with the expected
`malformed output returned usable custody` regression. Initial missing-API
RED1922d2 remains native evidence. Go1.27.1, temporary integration workspace,
disposable GOCACHE exit cleanup; no service created. Dependency pin/checksums
come from the immutable delivery8316d8a Git object, native e62de6.
No reusable release fingerprint archived. Public Send, SQL consuming/recovery
composition, exact coverage/mutation, clean-consumer and release gates remain
pending. Counts unchanged5/67 integrated,1/67 verified; no push/unit promotion.

Canonical native transcript:
`/Users/brian/.codex/sessions/2026/10/01/rollout-2026-10-01T04-46-29-01a0f524-42c6-7471-acc1-387f9f5ef3af.jsonl`.

## Delivery original-output restoration development, 8316d8a

Source base fa2afb2013c2cb069ba763b8b1dc7942443eed68; local source commit
8316d8af9fd74669a11b1bf3f38b87d52a3cfb38 contains exactly six delivery/postgres
files. Four dirty core OTP paths are excluded and preserved. Public OTP Send
is still unfinished; no unit promotion, release certification or push.

Native reports remain in the canonical parent transcript below, not a new
receipt framework. Fresh isolated read-only agents used gpt-6.1-sol, medium:
- /root/send_composition_map_fa2afb2: Send/delivery restart ownership mapping.
- /root/send_composition_qa_fa2afb2: one bounded public-composition QA artifact.
- /root/delivery_checkpoint_regression_fa2afb2: updated native APPROVE for
  regression and additive Go compatibility; expired restore-to-staging coverage
  remains advisory. No tests/services/writes by this reviewer.
- /root/delivery_checkpoint_final_fa2afb2: native APPROVE, no findings, complete
  six-file final diff; native parent executions inspected, not independently
  rerun. Local development only; full Send and release gates remain pending.

Both final native reports bind these reviewed SHA-256 values:
```
0cd5c0c0be378ca66111ae60a9d813b8a937196075e144cf01a0455821c0a674 preparation_checkpoint.go
549d13b916b0934297408ce93047f4ec535a7ba5c8a240c367ae4ccf9ab54e67 preparation_checkpoint_internal_test.go
09354952ba12b17191895454b59e1448789a47789a7bea5f6133b2efb1a3b048 preparation_checkpoint_test.go
17fc3070fa4ccd164234de6ffd41aee756d7626ea9c8f394bb976c8c61df2c26 contributor.go
5228156a0cce267619be47bfe8dbb562a8e9f87ddcc372c853b20930f53e7313 README.md
e827c1ed254e99341841d73c1b8a8d4d35bc7dddd62cdb092641786fab8c73c1 CHANGELOG.md
```

Native execution records: 9cf6d4 missing API compilation RED; 6ef036 copied
disposable generation-check mutant runtime RED (changed original generation
accepted); fe6147 session81591 terminal exit0, full delivery
`go test -race -count=1 .` PASS1.301s; 5c54e7 session83154 terminal exit0,
`GOMAXPROCS=2 go test -run '^$' -fuzz '^FuzzPreparationCheckpoint$' -fuzztime=100x .`
PASS0.301s,100 executions. Fuzz preceded later test-only partitions/docs; final
race covers final tests and production. All Go commands used Go1.27.1,
disposable GOCACHE with exit cleanup and temporary integration-only workspace.
No reusable complete-input release fingerprint was archived. Exact coverage,
complete mutation, PostgreSQL/public Send composition and release gates remain
pending. No database service was created for this pure restoration batch.

Canonical native parent transcript:
`/Users/brian/.codex/sessions/2026/10/01/rollout-2026-10-01T04-46-29-01a0f524-42c6-7471-acc1-387f9f5ef3af.jsonl`.

## Private journal preparation development, fa2afb2

Sender: /root/preparation_delivery_fa2afb2
Native final report:

**APPROVE — reviewed local committed private-journal development checkpoint only. No concrete delivery findings.**

- Verified `feature/identity-platform`, HEAD `fa2afb2013c2cb069ba763b8b1dc7942443eed68`, direct parent/base `52a2df5c5f49128f6630d91c9ff8d2b017c23814`.
- Exactly six PostgreSQL paths committed; all six committed SHA-256 hashes match the native final review. No PostgreSQL drift or staged work.
- Four dirty OTP paths remain preserved: `otp/CHANGELOG.md`, `otp/go.mod`, `otp/send_preparation.go`, `otp/send_preparation_test.go`. Control worktree has only the two specified documentation edits.
- Inspected native regression, compatibility and complete-diff final messages directly in the parent transcript; each approves this bounded snapshot. Retained reports accurately reproduce those messages. Correctness reviewer disclosed post-blind-review exposure to a sibling summary; approval does not establish absolute context isolation.
- Native execution `058479`, exit 0, confirms affected integration/race PASS **2.952s** and full PostgreSQL unit PASS **0.340s**. Not independently executed; no rerun for commit identity alone.
- Scratch absent; matching Docker container and volume absent. Native cleanup `6b9852` confirms removal.

**State boundaries:** local and committed verified; push not performed per retained parent evidence, and no local remote-tracking ref contains this commit—remote publication not independently queried. CI, deployment and production behavior are not verified.

Public Send, authenticated decoding/binding, final-plan attachment, composed recovery, complete-input fingerprints, exact coverage/mutation, clean-consumer and release gates remain unfinished or unverified. This is **not** package/platform completion, PR readiness or release certification.

Source baseline 52a2df5c5f49128f6630d91c9ff8d2b017c23814; local commit
fa2afb2013c2cb069ba763b8b1dc7942443eed68, six PostgreSQL files only. Existing
four dirty core OTP files remain outside this batch. Public Send is not implemented.
Fresh mapper otp_public_send_map_6c695b5 and bounded QA
send_preparation_qa_52a2df5 identify original-journal pre-generation ownership,
durable original output and no-regeneration after an ambiguous generation as
the consumed requirement. Native reports remain in the parent session transcript.

Fresh native executions: 0f901c missing API RED; copied disposable mutant
353260 RED for missing-output regeneration; copied pre-lock-time mutant b6b0b5
RED for generation using stale time across an observed journal lock wait.
Final 058479, session97362 exit0:
`go test -race -tags=integration -count=1 -run '^TestJournalPreparationIntegration$|^TestPostgresIntegration/(global_command_claims|journal_leases|terminal_journal|begin_ownership|coordinator_reservation)$|^TestPreparation' .`
PASS2.952s; `go test -count=1 .` PASS0.340s. Legacy lease/terminal helpers also
execute after the optional migration. Disposable Go1.27.1 workspace and PG18.6
only. Execution input fingerprints were not archived for release reuse; these
fresh development passes are not release certification. Owned cluster
identity-send-journal-zsvvhs and scratch /tmp/identity-send-journal.zsVvHS removed.

Native parent transcript:
`/Users/brian/.codex/sessions/2026/10/01/rollout-2026-10-01T04-46-29-01a0f524-42c6-7471-acc1-387f9f5ef3af.jsonl`.

Sender: /root/preparation_regression_52a2df5
Native final report:

**APPROVE — final-snapshot replacement report.**

Role: `regression_reviewer`. Six selected `postgres` files, branch `feature/identity-platform`, working-tree changes over base/HEAD `52a2df5c5f49128f6630d91c9ff8d2b017c23814`; dirty OTP files excluded. Independently confirmed integration-test SHA-256 `6f1da1306300539cef56b8153e1d63e296d233f15121581cc280ded6732441e3`.

Inspected both added legacy helper bodies and their execution after preparation migration. They meaningfully cover unchanged legacy ownership, rollback, heartbeat, takeover, terminal commit/abort, immutable receipt replay, and stale/expired-owner denial on the upgraded schema. No unresolved findings; prior post-lock-time advisory remains resolved.

Parent-reported final `058479`: affected integration race PASS `2.952s`, full PostgreSQL unit regression PASS `0.340s`, session `97362` exit `0`. Not independently executed.

Approval remains bounded to private journal preparation development. Exact release coverage/mutation gates, authenticated decoding, public Send, coordinator orchestration, and final-plan attachment are not certified.

Sender: /root/preparation_go_52a2df5
Native final report:

**Independent Go compatibility reconciliation: APPROVE; no findings.**

Reviewed unchanged HEAD/base `52a2df5c5f49128f6630d91c9ff8d2b017c23814` plus the complete six-file delta. Updated integration-test SHA-256: `6f1da1306300539cef56b8153e1d63e296d233f15121581cc280ded6732441e3`; other five reviewed file hashes unchanged.

The two added calls execute existing lease and terminal suites **after** installing the optional migration. Their assertions cover ordinary insertion, heartbeat, takeover, rollback, terminal commit/abort, immutable replay, and stale/expired-owner denial on the upgraded schema. Parent-reported integration-race PASS `2.952s` therefore addresses the prior upgraded-schema legacy journal limitation; full PostgreSQL unit PASS `0.340s` remains parent evidence. I inspected source but did not execute tests.

Additive API, unchanged migration history/dependencies, and explicit Go `1.27.0` compatibility conclusions stand. Clean external-consumer compilation, full public Send composition, authenticated decoding, final-plan attachment, and release gates remain unverified or unfinished.

Sender: /root/preparation_final_52a2df5
Native final report:

**APPROVE — bounded local development only. No findings.**

Independent correctness review of `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`, branch `feature/identity-platform`, base/HEAD `52a2df5c5f49128f6630d91c9ff8d2b017c23814` plus the complete six-file dirty delta. Four pre-existing dirty OTP paths are excluded.

Source-backed conclusions:
- Global claim locking preserves original intent and serializes competing reservations.
- Committed generation markers prohibit regeneration without retained output; rollback does not fabricate generation history.
- Start/retention checks fence owner and generation and read primary time after locking.
- Ciphertext is bounded, immutable, and retains its original generation across takeover.
- Optional migration preserves historical migration bytes and leaves ordinary rows unenrolled. Upgraded-schema legacy lease and terminal tests are included.
- Existing public query/recovery paths reject missing execution plans; they cannot convert these preparation rows into execution authority. Missing journals and retired claims fail closed rather than recreate authority.

**Verification:** inspected native execution records, not independently rerun:
- `0f901c`: missing preparation API RED.
- `353260`: removed no-regeneration fence mutant RED.
- `b6b0b5`: pre-lock clock mutant RED.
- Final `058479`, exit 0: affected integration/race command PASS **2.952s**; `go test -count=1 .` PASS **0.340s**.

Final SHA-256:
```text
f279bf62ea30745e24095945abc8434626f71ac2483e8cd27f334ee5e6a35ec0  postgres/preparation_journal.go
84f8c6538cb7387c82bbb211e55f3aa44ad93d9d2c2e6012449d4960a94f50b4  postgres/preparation_schema.go
8a76f060eeb56cea1eefae2979b7fd5785ad7bd30f937ea2c97f4b7b6518bfb6  postgres/preparation_journal_test.go
6f1da1306300539cef56b8153e1d63e296d233f15121581cc280ded6732441e3  postgres/preparation_journal_integration_test.go
c43a89b5efaf11d2aedea099b6760e88471e3a6ba17d17c1e3972a8bb7341901  postgres/README.md
84919749ba11f11406d795b77dac9faac4a73870caaa327f64c7ca482e10ba16  postgres/CHANGELOG.md
```

Not certified: public Send, authenticated decoding/binding, final-plan attachment, composed delivery/recovery/expiry erasure, exhaustive coverage/mutation, release gates, privileges, deployment, or production behavior.

Provenance caveat: a log search exposed a nested sibling summary **after** the blind source review; it was not used as review evidence. Final verification was extracted directly from native result records.

## Committed development state audit, 52a2df5

Sender: /root/request_key_delivery_52a2df5
Native final report:

**APPROVE — local committed DEVELOPMENT checkpoint only.**

- Verified branch `feature/identity-platform`, ancestry `924c6c5 → 8969d59 → 52a2df5`, seven root and ten PostgreSQL committed paths, unchanged reviewed paths, and four preserved uncommitted OTP preparation paths.
- Inspected native regression, compatibility and replacement complete-final-diff approvals. Four supplied root hashes match committed blobs; no separate final PostgreSQL hash snapshot exists. Existing seven-field configuration and additive opt-in constructor remain intact.
- Race/integration/unit and legacy-consumer passes are parent execution evidence, not independently rerun. Cluster removal is parent-reported; `/tmp/identity-request-pg.tlI7nP` still requires cleanup after handoff.

No publication, CI, release, deployment or unit promotion certified. Complete-input evidence fingerprints, exact coverage/mutation, clean consumers, downstream OTP/HTTP behavior and release gates remain pending; counts stay **5/67 integrated, 1/67 verified**.

## Replacement PG final and compatibility reviews, 8969d59

Sender: /root/request_key_pg_final_8969d59
Native final report:

**APPROVE — fresh local-development checkpoint. No findings.**

The earlier approval is withdrawn: it missed nested diagnostic disclosure of the new lookup key. This disposition replaces it after complete re-review of all ten pending PostgreSQL files against HEAD `8969d596e5c8d6d790377bce44099c08338108ee`; `otp/*` remains excluded.

`NewRequestKeyConfig()` now captures a copied key in a private closure, leaving no reflectively traversable key bytes or public projection. The nested-wrapper regression exercises ordinary and unsupported formatting verbs; parent supplied its expected RED result before the fix. Constructor consumers, current authorization, namespace pinning, immutable storage, retained-key replay, cancellation, ambiguous commits, migrations and fixture changes remain coherent.

Current verification supplied by parent:
- Legacy seven-field configuration consumer: **PASS**.
- Focused PostgreSQL race/integration command: **PASS**, 19.046s.
- Full PostgreSQL unit tests: **PASS**, 0.342s.
- Reviewer’s `git diff --check -- postgres`: **PASS**.

No writes, execution, services or delegation performed. This approves only the correlation development batch—not OTP Send, duplicate-generation prevention, HTTP enforcement, exact coverage/mutation, published clean-consumer compatibility, release readiness or package completion.

Sender: /root/request_key_go_924c6c5
Native final report:

**Native Go compatibility disposition: APPROVE — local development checkpoint.**

Reviewed `8969d596e5c8d6d790377bce44099c08338108ee` plus the current pending PostgreSQL slice.

- `NewRequestKeyConfig` now owns copied lookup-key custody through a private closure (`request_key.go:43`). All inspected callers use the updated constructor. Removing the earlier uncommitted exported key field does not break an established public API.
- Existing seven-field `CoordinatorConfig` and `NewCoordinator` remain unchanged; correlation remains explicitly opt-in.
- PostgreSQL still pins the committed root APIs at `v0.0.0-20261007092517-8969d596e5c8`; no new dependency edge or cycle is introduced.

No actual compatibility defect found. PostgreSQL remains uncommitted; published availability, replacement-free clean-consumer interoperability and release gates remain pending. Reported test/legacy-consumer passes were not rerun; no writes or execution performed.

## Final PG key-custody regression reconciliation, 8969d59

Sender: /root/request_key_regression_924c6c5
Native final report:

**APPROVE — local-development regression reconciliation.**

Reviewed HEAD `8969d596e5c8d6d790377bce44099c08338108ee` plus the complete pending PostgreSQL slice; OTP excluded.

The lookup-key disclosure finding is resolved: `NewRequestKeyConfig` copies the key into private closure-backed custody, removing reflectively reachable bytes. The new nested-diagnostics regression covers private value/pointer wrappers and slices across supported and unsupported formatting verbs. Constructor validation, configuration copying, authorization, rotation and correlation tests remain coherent. No residual material regression found.

Parent-reported final evidence: legacy seven-field consumer PASS; PostgreSQL integration race PASS (19.046s); full PostgreSQL unit tests PASS (0.342s). Not rerun here.

Release, exact coverage/mutation, clean-consumer and downstream OTP/HTTP gates remain pending.

## PG key-custody regression finding after initial final review, 8969d59

The earlier PostgreSQL final approval below is superseded for the subsequently
edited tree; commit was withheld. Runtime RED41dd86 reproduced the finding.

Sender: /root/request_key_regression_924c6c5
Native final report:

**REQUEST_CHANGES — local-development regression audit.**

Reviewed HEAD `8969d596e5c8d6d790377bce44099c08338108ee` plus the ten pending PostgreSQL files; OTP excluded.

Constructor compatibility is preserved. Tests cover default-disabled behavior, configuration copying and invalid configuration rejection.

**Confidentiality gap:** `postgres/request_key.go:28–35` retains the secret lookup key as a reflectively reachable `[32]byte`. Formatting a private wrapper containing `RequestKeyConfig` bypasses its `Format` method—the same mechanism established by the root custody regression. `postgres/request_key_test.go:139` checks only direct formatting, leaving nested configuration diagnostics uncovered. Reproduce with synthetic bytes and protect this new secret custody boundary before committing.

Parent-reported root race, PostgreSQL integration race and PostgreSQL unit checks passed; not rerun here. Release, mutation, clean-consumer and downstream OTP/HTTP gates remain pending.

## Request-key PostgreSQL native final and compatibility reports, 8969d59

Sender: /root/request_key_pg_final_8969d59
Native final report:

**APPROVE — local-development checkpoint only. No findings.**

Reviewed all ten selected PostgreSQL files, including complete untracked source/tests, against committed HEAD `8969d596e5c8d6d790377bce44099c08338108ee` on `feature/identity-platform`. Unrelated `otp/*` changes excluded. Single-reviewer structured pass; no writes or code execution.

Source-backed conclusions:
- Separate opt-in constructor preserves existing `CoordinatorConfig` and default-disabled behavior.
- Current authorization precedes SQL; tenant/actor/method/route scope and independently keyed intent remain distinct.
- Unique immutable mappings conserve the winner across concurrent retries, retained-key rotation, ambiguous commits and aged/retired commands.
- Changed intent, lookup-key replacement and missing retained keys fail closed without admission.
- Root pseudo-version matches the immutable commit timestamp/hash. Inspected canonical `x/mod v0.40.0` archive/checksum derivation; supplied hashes match `go.sum`.
- CAPTCHA fixture rename consistently separates its integration-test ownership.

Verification supplied by parent:
- Seven-field unkeyed `CoordinatorConfig` consumer: **PASS**.
- `go test -race -tags=integration -count=1 -run '^TestRequestKey|^TestPostgresIntegration$' .`: **PASS**, 18.992s.
- `go test -count=1 .`: **PASS**, 0.332s.
- Reviewer’s `git diff --check -- postgres`: **PASS**.

This does **not** certify public OTP Send, duplicate-generation prevention, HTTP enforcement, published clean-consumer compatibility, exact coverage/mutation, release readiness or whole-package completion. Those boundaries remain pending.

Sender: /root/request_key_go_924c6c5
Native final report:

**Native Go compatibility disposition: APPROVE — local development checkpoint only.**

Reviewed committed root `8969d596e5c8d6d790377bce44099c08338108ee` and the complete pending PostgreSQL slice; OTP excluded.

- Earlier unkeyed-literal incompatibility is **resolved**: `CoordinatorConfig` retains its original seven fields, and `NewCoordinator` is unchanged. Opt-in correlation uses additive `NewCoordinatorWithRequestKeys` (`coordinator.go:126`). Inspected the actual seven-value legacy consumer; its reported successful execution was not rerun here.
- Root custody is compatible: newly introduced `IdempotencyKey` remains comparable, but equality expressly identifies the constructed owner—not byte content (`idempotency_key.go:12`). PostgreSQL correlates through exact-byte keyed digests, so independently constructed equal-byte retries remain interoperable. Request-intent projections remain copied.
- PostgreSQL now pins the committed root API revision at `go.mod:11`. Pseudo-version timestamp matches the immutable commit. Inspected canonical VCS checksum derivation and matching recorded revision; hashes were not independently recomputed.
- No new module, dependency cycle, permanent replacement or forced downstream pin is introduced. Existing consumers can retain their current APIs; consumers adopting request correlation need the upcoming PostgreSQL revision.

No actual source/behavior compatibility defect remains in scope. PostgreSQL is still uncommitted; remote availability, replacement-free clean-consumer interoperability and full release gates remain unproven. No writes, builds, tests or services executed.

## Root request custody final native reviews, 924c6c5 pending batch

Sender: /root/request_key_root_final_review_924c6c5
Native final report:

**APPROVE — local development checkpoint only. No findings.**

Reviewed all seven selected files against base/committed HEAD `924c6c5f62adacc2778d75e9425c34a8feb95a94` on `feature/identity-platform`. Current production/test hashes match the supplied final snapshot.

The diagnostic-custody findings are resolved: private closure storage prevents recursive formatting—including unsupported-verb fallback—from reaching keys or intent fields. Expanded regressions cover private value, pointer, slice, array, and map wrappers. Bounds, immutable ownership, HMAC framing, scalar preservation, redacted errors, and existing FeatureCommand behavior remain coherent.

Verification: parent supplied current `go test -race -count=1 .` **PASS (1.523s, exit 0)** and expected RED results for both disclosure regressions. Reviewer’s `git diff --check` passed; no code execution or mutations performed.

Coverage, mutation, API baselines, clean-consumer, and release gates remain pending. PostgreSQL/OTP changes are excluded from this approval; no dependent-package or release-readiness claim.

Sender: /root/request_key_regression_924c6c5
Native final report:

**APPROVE — bounded regression reconciliation, local development only.**

Reviewed current root custody correction over HEAD `924c6c5f62adacc2778d75e9425c34a8feb95a94`; PostgreSQL context retained, OTP excluded. The earlier confidentiality approval was incomplete: nested private wrappers could disclose values despite `Format`.

Closure-backed custody removes reflectively reachable raw values. Both new diagnostic tests exercise private value/pointer wrappers, slices, arrays and maps across ordinary and unsupported verbs, checking plaintext and hexadecimal disclosure. Existing digest, owned-projection and concurrent-read assertions remain intact. No unresolved defect found.

Parent-reported final root `go test -race -count=1 .`: PASS, 1.523s; not rerun here.

Release, exact coverage/mutation, clean-consumer, HTTP enforcement and working OTP Send gates remain pending.

## Request-key native regression and compatibility reports, 924c6c5 dirty batch

Sender: /root/request_key_regression_924c6c5
Native final report:

**APPROVE — scoped regression audit.**

Reviewed branch `feature/identity-platform`, base/committed HEAD `924c6c5f62adacc2778d75e9425c34a8feb95a94`, plus all 15 named pending retry-key files, including untracked files. Separate `otp/*` preparation changes excluded.

No material behavioral or regression-test defects found. Coverage exercises independent HMAC/CBOR framing, immutable input custody, exact lookup scope, matching/conflicting SQL races, retained-key rotation, current authorization, ambiguous commits, aged/retired mappings, and cancellation while waiting on database uniqueness.

Execution evidence supplied by parent, not rerun:
- Root `go test -race -count=1 .`: PASS.
- PostgreSQL `go test -race -tags=integration -count=1 -run '^TestRequestKey|^TestPostgresIntegration$' .`: PASS.

**Unverified boundaries:** working OTP Send and duplicate-generation prevention (RK-11), HTTP header-source enforcement, full scalar-oracle breadth, prolonged stress/leak behavior, exact coverage/mutation and release/clean-consumer gates. This disposition supports only the local correlation development batch—not package/platform completion or release readiness.

Sender: /root/request_key_go_924c6c5
Native final report:

**Go compatibility disposition: APPROVE for the scoped local development checkpoint.** Reviewed base/head `924c6c5f62adacc2778d75e9425c34a8feb95a94` plus pending root/PostgreSQL retry-key changes. No confirmed incompatibility in existing consumers.

- Root additions are isolated public types/functions; `feature_command.go` validation extraction preserves existing semantics.
- Existing in-tree `CoordinatorConfig` consumers use keyed literals; default-disabled behavior remains unchanged.
- **Compatibility caveat:** [coordinator.go:42](/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform/postgres/coordinator.go:42) adds an eighth exported config field. Existing seven-value **unkeyed** literals would stop compiling. Do not claim universal source compatibility; preserving those callers would require a separate opt-in constructor/config API.
- **Pending dependency boundary:** [postgres/go.mod:11](/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform/postgres/go.mod:11) still pins `19cdb33a1e8e`, which lacks both new root API files. Commit the root API/changelog batch first, then update PostgreSQL’s root requirement/checksum to that immutable prerequisite revision before committing the dependent adapter. Preserve separate unfinished OTP changes.
- No new module/dependency edge or cycle is introduced. Only PostgreSQL consumes the new root APIs presently; other nested modules need no pin change solely for these additions.
- Published/clean-consumer interoperability, release gates and package completion remain unproven. Reported workspace race results are parent-provided local evidence, not independent execution here. No builds, tests, services or writes performed.

## Native issuance delivery audit and next public Send mapping/QA, 924c6c5

Message Type: FINAL_ANSWER
Task name: /root
Sender: /root/otp_public_send_map_6c695b5
Payload:
## Read-only change mapping — NEXT public OTP Send

**Source:** `go-identity` worktree `feature/identity-platform`, committed HEAD `6c695b5703b6a345cd61ed34ce3e377a56ce063c`, plus seven pending `otp/postgres` files inspected as provisional context. No writes, code execution, services, or delegation. This is mapping, not correctness approval or executable prerequisite proof.

**Conclusion:** Public usable `otp.Service.Send` is absent. The smallest next coherent behavior boundary is the platform-owned public Send operation and its official adapter execution/recovery path—not another exposed low-level issuance constructor.

### Present seams and owners

| Boundary | Current source and status |
|---|---|
| Code policy/generation | Committed `otp/profile.go:62`, `otp/generation.go:30`: validated purpose, entropy, TTL, resend, sends, attempts and replacement; bounded unbiased generation; borrowed bytes cleared; callback failures redacted. Generation is expressly not issuance. |
| Scoped keyed digest | Committed `otp/digest.go:73,199`, `otp/scope_digest.go:47`: complete tenant/purpose/subject/channel/destination/challenge/target/policy/generation/authorization scope; retained key-version lookup, fail-closed retirement. |
| Original authorization custody | Committed `otp/authorization.go:21,84,213`, `otp/session_authorization.go`: explicit disjoint pre-auth/session union and original records; remember policy cannot be inferred, renewed, or upgraded. Structural custody alone is not live authorization. |
| Official issuance mapping | Committed `otp/postgres/issuance_custody.go:35,43`: code-free authenticated encrypted command custody; actual cookie/token remains separate. Mapping once is required; remapping produces different intent. |
| Durable family/challenge ownership | Committed `otp/postgres/issuance.go:17`, `send_window.go:42`: database-owned windows, family serialization, challenge insert, replacement, bounded predecessor locks, retained replay fingerprint. Reserved predecessors block replacement rather than being silently revoked. |
| Atomic execution | **Pending**, `otp/postgres/issuance_execution.go:16,84`: private contributor joins primary directory/user/credential/authorization, original flow, lockout/family, challenge, encrypted delivery, lifecycle and audit through the existing Coordinator. It explicitly rejects RiskEvidence, CommandCounters and CaptchaEvidence plans. It is not a ready public consumed contract. |
| Encrypted delivery | Committed `delivery/postgres/contributor.go:148,253`: `Prepare` encrypts bounded rendered recipient custody before transaction; `Stage` borrows original Work. It does **not** resolve or authorize recipients. Worker/provider retries belong to delivery, not OTP. |
| Primary command admission | Committed `postgres/admission.go:53`: authorized actual absence or safe historical metadata before expensive work; absence is not rollback proof after submission. Existing Coordinator/query/recovery owns journal authority. |
| Bounded recipient state | Committed `postgres/identifiers.go:39`: complete tenant/user snapshot capped at 128, rejecting truncation. `delivery/postgres/recipient.go:29,81` has owned canonical-recipient selection and native version fencing, but that path is private to delivery enqueue and cannot be assumed available to OTP Send. |

### Missing public behavior

No `Service`, `SendRequest`, or `Send` implementation exists under `otp`. The public operation catalog requires `Send(context.Context, SendRequest) (SendResult, error)` with tenant, stable command ID, channel, canonical subject reference, closed purpose, and idempotency key (`PUBLIC_CONTRACTS.json:7900,89785`).

Missing integrated behavior includes:

- public construction/admission and stable operation-specific errors/results;
- server-owned subject/destination resolution, including absent-account enumeration-safe handling;
- current original pre-auth/session possession and authority admission **before** generation, plus native rechecks after lock waits;
- public-key admission bound to request intent, and exact replay/recovery lookup **before** risk evaluation, recipient rendering or randomness;
- generation once with the **same** borrowed code used for digest and rendered encrypted delivery;
- durable safe Send result reconstruction for committed/replayed/unknown commands;
- risk, identifier/IP/device admission and declared delivery-rate policy integration;
- public direct-consumer composition without custom domain stores/workflows.

The private issuance phase currently validates delivery purpose, subject and expiry, but does not establish that the supplied rendered message contains the same code as the supplied challenge digest. That guarantee belongs to the missing owned Send orchestration.

### Risk / counter / CAPTCHA ownership

**Correction to my intermediate message:** the exact source enum is `BeforeCredential`; the normative meaning is before credential **or provider** work.

- `risk/catalog_generated.go:201` assigns `identity.otp.send` to **Authentication**.
- `risk/policy.go:198` fixes tenant + canonical network + device + canonical identifier digest, `BeforeCredential`, throttle, unknown-deny.
- `risk.CounterCoordinator` + `CounterStore` owns nontransactional velocity mutation and same-fingerprint replay (`risk/counter.go:243`). PostgreSQL owns the journal/recovery; Valkey is subordinate ephemeral authority.
- `risk.CommandCounterPolicy.Prepare` explicitly accepts only CredentialMutation/Privileged `ReserveWithCommand` profiles (`risk/command_counter.go:80`). It is **not** the existing Send authentication counter seam.
- `risk.Evaluator.Evaluate` is non-issuing; public decision fields cannot become durable evidence (`risk/evaluation.go:206,224`). Configuring no counters/audit is documented development mode and cannot prove full Send admission.
- Phone evidence issuance is phone-recovery-specific (`risk/issuance.go:415`); it is not a generic OTP Send evidence service.
- CAPTCHA issue/use and native reservation/application/finalization already belong to `risk.CaptchaEvidenceIssuer`, `risk/postgres` and the shared Coordinator. OTP must consume those owned seams when its selected policy requires CAPTCHA, not accept caller-selected “risk allowed” data or create a duplicate engine.
- `API_OPERATIONS.md:238,256` separately requires `rate.delivery`: trusted IP 10/hour, destination 5/hour, subject when present 10/hour, shared store, fail closed with enumeration-safe delivery envelope. OTP family sends/hour and Authentication velocity are different policies; neither substitutes for this contract.

This pass inspected risk core/PostgreSQL ownership but did **not** inspect the separate Valkey or generic rate-limit repository/revision. Their consumed adapter readiness therefore remains unproved here, not categorically unavailable.

### Callers, consumers and variants

The real source has primitive/adaptor use, not a usable public Send caller. Intended public consumers are direct backend composition and opt-in `POST /v1/otp/send`; owning email/password/phone/MFA workflows consume purpose-bound OTP authority. Send itself must not implement those protected domain transitions or issue a signin session.

Required variants remain distinct:

- signin and pre-auth MFA preserve original remember policy;
- authenticated-session purposes retain full session, credential, authority and original authentication timeline;
- verification/reset/change/MFA/step-up cannot cross purpose/target/channel;
- absent-account delivery must preserve the same safe public shape;
- retain versus replace is explicit; neither renews previous expiry;
- delivery accepted/queued is not delivered, verified, or authenticated;
- cancellation before dispatch differs from cancellation/unknown after dispatch; unknown replay must not regenerate, re-evaluate, replace or infer rollback.

Identifier change/removal, anonymization/deletion and global compromise explicitly include OTP as a lifecycle consumer (`LIFECYCLE_CONSUMERS.md:29–30,41–46`). Current Send must fence identifier/current authority, while complete lifecycle cleanup remains an affected later obligation rather than being manufactured as a second workflow.

### Concrete specification reconciliation points

1. `REFERENCE_CONFIGURATION.md:266` calls predecessor challenge replacement **superseded**, whereas the selected OTP transaction contract permits `issued → revoked` and current source uses `revoked`. This is a naming/catalog contradiction, not justification to weaken atomic replacement or delivery semantics.
2. Public catalog `DestinationInput` bounds allow up to 4096 bytes; committed `otp.NewDestinationInput` caps at 320 (`otp/digest.go:46`). Current actual supported bound should stay explicit; do not claim the catalog-wide larger bound is implemented.
3. Catalog request fields alone do not carry full authorization/remember/risk custody. Relevant trusted public composition must supply that context without nullable-field inference or treating a subject reference as authentication.

### Development disposition

Committed HEAD contains substantial usable **primitives**, not a complete Send prerequisite. Pending private atomic issuance must first receive its own review/commit/proof as requested. Afterwards, public Send can build on that committed contract, but full usable Send cannot honestly be claimed until owned recipient resolution, idempotent replay/recovery, authentication/risk/rate admission and same-code delivery are integrated and exercised through its public API. No extra receipt machinery or release-wide rerun follows merely from this mapping.

Message Type: FINAL_ANSWER
Task name: /root
Sender: /root/otp_issuance_execution_delivery_924c6c5
Payload:
**Disposition: APPROVE — bounded local committed development claim only.**

- Verified clean source worktree, branch `feature/identity-platform`, HEAD `924c6c5f62adacc2778d75e9425c34a8feb95a94`, and exactly seven committed OTP paths against `6c695b5703b6a345cd61ed34ce3e377a56ce063c`. `git diff --check` passes.
- Inspected retained native regression, compatibility and complete-final reports. All seven committed blob hashes match final/compatibility approvals; the four production/test hashes also match the earlier regression approval. Later README/dependency metadata does not acquire regression approval retroactively.
- Dependency pins match the actual prerequisite commit and timestamp. No unresolved delivery finding for this private issuance/scalar-history batch.
- Real-PG integration/race `39.651s` and updated-pin focused test `0.283s` remain **parent-reported**, not independently executed or authenticated here. Scratch absence independently verified; cluster/cache cleanup remains parent-reported.

Delivery states: **local implemented and committed**; pushed, hosted-CI verified, released/published, deployed and production-verified **not established**. No branch upstream or matching remote-tracking feature branch exists locally.

Public Send/complete OTP/platform acceptance, direct session/recovery partitions, exact coverage/mutation, published clean-consumer composition and applicable supply-chain/release gates remain pending. No readiness or product-completion claim; no rerun warranted solely because reviewed bytes became a commit.
Message Type: FINAL_ANSWER
Task name: /root
Sender: /root/otp_public_send_qa_6c695b5
Payload:
## Native QA report

- Principal: `/root/otp_public_send_qa_6c695b5`, fresh isolated read-only pass.
- Skill: read the complete `$qa-test-planner` contract.
- Repository: `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`.
- Branch/head inspected: `feature/identity-platform`, `6c695b5703b6a345cd61ed34ce3e377a56ce063c`, plus the seven disclosed dirty private issuance files.
- Scope: the next **public OTP Send owning workflow**, not approval of private issuance as a public API.
- No writes, tests, services, implementation delegation, historical evidence prose, or other reviewers’ outputs.

### Contract and source anchors

Control repository `.ai/identity-platform`:

- `goals/identity-otp.md`: generation, complete scoped keyed digest, identifier/IP/device admission, enumeration resistance, explicit authorization union, remember preservation.
- `goals/identity-otp-postgres.md`: authoritative issuance/send-limit/replacement storage; database time; stable-command recovery.
- `API_OPERATIONS.md:359`; `OPERATION_SEMANTICS.json:3518`: keyed Send, purpose-dependent flow authorization, enumeration-safe delivery/challenge state.
- `REFERENCE_CONFIGURATION.md:259–266`: reference code policy, send window/cooldown, atomic replacement.
- `TRANSACTION_CONTRACT.md:658–665,733–753`: issue before delivery, reserved predecessor exclusion, atomic journal/outbox effects and authenticated encrypted bearer delivery.
- `SHARED_CONTRACT_APPLICABILITY.json:1261–1368`: selected OTP transactions, configuration, cascades and security events.
- `SECURITY_EVENTS.md:145`; `LIFECYCLE_CASCADES.md` exact event catalog.

Implementation repository:

- `otp/generation.go:12–82`: bounded rejection sampling, synchronous borrowed code, cleanup and error normalization.
- `otp/profile.go:78–144`: construction bounds and replacement policy.
- `otp/authorization.go:163` onward: immutable original flow, purpose restrictions and unchanged remember policy.
- `otp/postgres/issuance_custody.go:20–158`: digest-only command custody; mapping is expressly not generation/admission.
- `otp/postgres/issuance_execution.go:16–22,46–176`: private-only phase, primary authority/flow rechecks, atomic persistence plus encrypted delivery.
- `otp/postgres/send_window.go:15–70`: stable family identity, exact rolling window and cooldown.
- `delivery/postgres/recipient.go:70–109,144` onward: bounded primary canonical recipient selection and transaction-time recipient fencing.
- `risk/evaluation.go:128–150,224`; `risk/counter.go:22–30,310`: server-bound risk dimensions and assessment.
- Existing `TestNativeIssuanceCoordinatorAtomicity`, `TestNativeChallengeIssuanceStorage`, `TestNativeIssuanceReplacementAndRollback`, and `TestNativeSendWindowStorage` exercise the private infrastructure. They do **not** prove public Send ownership.

## One bounded implementation coverage artifact

Use external-package public-consumer tests backed by real PostgreSQL and existing owned adapters. Sender capture is permitted; consumer-written stores/domain orchestration are not.

| Required scenario | Observable assertions / independent oracle | Realistic defect caught |
|---|---|---|
| **1. Public success and same code** | Invoke public Send without a consumer-generated code, challenge, digest or rendered payload. Read primary challenge, original command result and encrypted effect. Through the owned delivery worker, capture the recipient and message in memory; extract its code and independently verify against the stored keyed digest using the exact original scope. Assert one challenge, one delivery effect, one success audit/lifecycle transition, stable journal identity, pending delivery—not provider success. | Generated digest and delivered code differ; public API merely exposes private orchestration; duplicate generation/effects; false delivery-success claims. |
| **2. Flow/purpose/remember variants** | Cover pre-auth signin with persistent and nonpersistent remember choices, pre-auth MFA continuation, and authenticated-session non-signin use. Inspect retained complete flow and unchanged remember durations/choice. Reject legacy ID-only context, wrong bearer, wrong discriminant, cross-tenant/purpose/subject/target and changed original records. Purpose/channel parameter cases must make mismatches fail against the actual captured code. | Inferred authentication, fallback persistence upgrade, renewed flow, cross-purpose code authority. |
| **3. Owned risk before secret/provider work** | Configure real owned risk policy/counter adapters and server sources. Vary canonical identifier, network and device independently until each dimension rejects. Unknown/unavailable/deny produces no issued challenge or deliverable effect. Caller-forged decision/profile/subject facts cannot bypass admission. Assert canonical dimension aliases share allowance. Verify source input binds `identity.otp.send` and the original logical request. | Identifier-only limits, forged Allow, request-selected policy, IP/device omission, generation/provider work before denial. |
| **4. Bounded resolution and non-enumeration** | Known eligible recipient, absent recipient, wrong tenant, suspended user, removed/changed identifier and purpose-ineligible pending identifier. Compare public result **shape and allowed fields**, not random values; absent/ineligible must not disclose account existence through challenge/effect handles, status or error distinction. Assert canonical address/channel actually sent; no delivery to request-substituted address. Resolver overflow/cancellation fails boundedly without arbitrary first-match selection. | Enumeration via supposedly safe result, wrong recipient, unbounded lookup, unsupported pending-recipient admission. |
| **5. Exact durable resend limits** | Real database-time cases immediately before/at/after cooldown and rolling one-hour cutoff; sends at limit versus next send. Retry with new pre-auth flow, challenge ID, target and rotated digest key must not refill the same family. Assert primary send history, rejected issuance absence and predecessor state. Preserve existing exact deterministic window tests rather than duplicate all pure partitions. | Off-by-one time comparisons; flow/key rotation resets allowance; rejected sends consume a slot unintentionally. |
| **6. Replacement, retention and races** | Replace: predecessor remains valid before commit, becomes unusable after successor commits; rollback preserves it. Retain: earlier expiry/budget remain unchanged. Reserved predecessor blocks replacement until authoritative recovery. Race two public sends and send-versus-reserve with barriers: persisted family history never exceeds limits; no partially replaced state or second reservation winner. | Premature invalidation; successor without delivery; reserved credential destruction; resend/verify race. |
| **7. Stable replay and conflict** | Replay same request identity after success with generation/resolution/risk sources unavailable or changed: original result/effect identity, no new code, no extra counter charge or sender submission. Changed tenant/purpose/destination/flow/remember/policy intent under the same identity conflicts without mutation. Scope-authorized result access remains required; replay does not confer renewed current authority. | “Idempotency” implemented by rerunning generation/risk/recipient selection; changed intent silently accepted. |
| **8. Ambiguous commit and recovery** | Inject dropped commit acknowledgement using the existing owned coordinator test seam. Send reports Unknown, not clean failure/success; recover the original command and recover committed or proven-not-committed outcome. During uncertainty, another Send/recovery must not regenerate, re-evaluate risk, reconstruct a bearer, replace predecessors or enqueue another effect. Payload erasure/expiry alone cannot prove rollback; retained challenge/tombstone prevents negative classification. | Lost-ACK duplicates; timeout treated as rollback; absent ciphertext interpreted as noncommit. |
| **9. Authority races and full rollback** | After initial resolution, change directory/user/identifier/credential/authorization/session/pre-auth authority before application; separately hold family/predecessor lock until flow or credential freshness expires. Public operation denies and rolls back challenge, family append, replacement, encrypted effect, success audit/lifecycle and successful journal result. Inject delivery/audit failure after provisional challenge mutation and assert the same atomic outcome. Distinguish an authoritative aborted journal from “no journal at all.” | Snapshot authorization mistaken for commit authority; late lock waits extend validity; partial commit. |
| **10. Cancellation and secret confinement** | Cancellation before admission, after preparation and during lock wait; renderer/source/provider error containing recognizable sentinel secret. Assert stable redacted public errors and no leaked sentinel/code/address/bearer in formatted values, serializations, audit/lifecycle/journal payloads. Capture/decrypt only inside test memory; primary challenge has digest only and delivery is authenticated ciphertext. Existing generation tests remain the oracle for bounded rejection sampling and cleanup. | Secret-bearing wrapped errors, plaintext persistence, late cancellation success, weaker modulo generator. |

### Smallest sufficient verification set

1. Existing focused `otp` generation/profile/digest/authorization tests plus new public contract tests.
2. Focused real-PostgreSQL public Send integration cases above in `otp/postgres`; reuse existing native issuance/window/race tests as lower-bound evidence.
3. Direct public consumer compilation/example proving composition needs no consumer workflow/store.
4. Targeted race execution for new Send/resend/recovery concurrency tests, not an unrelated repository matrix.
5. Only affected public API, documentation and module contracts; pending full release coverage/mutation and other release gates remain explicit obligations, not pre-write approvals.

All executed Go checks require task-owned disposable `GOCACHE` and cleanup. **No checks executed in this QA pass.**

### Red-test strategy

Start with the external public-consumer success test: public Send must own generation, resolution, admission and atomic issue/enqueue. Missing public API compilation proves only missing API; obtain an executable failing behavior assertion once the contract surface exists, **before implementing successful orchestration**. Its expected failure is no committed, delivery-verifiable challenge—not a mock call count.

Then introduce negative/replay/unknown/race cases before their corresponding implementation branches. Use known-bad perturbations only where needed to prove an existing test would reject independent generation, skipped authority fencing or duplicate effects. Do not weaken assertions to match private fixtures.

### Material reconciliation required

The inspected private phase requires audit `"otp.send"` and lifecycle `"otp.sent.v1"` in `issuance_execution.go:46`. The selected security catalog registers `"identity.otp.request"`/`"identity.otp.verify"`; the exact lifecycle catalog has no OTP event row. Public Send must reconcile the authoritative names/catalog before completion; inheriting fixture names is not a valid oracle. This is implementation work, **not a reason to block the first write**.

The catalogs establish enumeration-safe outcomes, but do not prescribe the exact public opaque result shape or the no-recipient effect representation. Choose one bounded representation and test its indistinguishability/no-provider-send behavior. Do not turn the selected representation into a second command journal.

### Optional/deferred coverage

Provider-specific interoperability, broad cross-purpose consuming workflows, cleanup/migration backup-restore, all-factor lifecycle composition, long fuzz/stress campaigns and exhaustive release mutation belong to their affected later acceptance/release boundary. This plan requires targeted purpose/authority cases now, not the complete 67-unit matrix.

**Disposition:** no contradiction makes the requested public Send behavior invalid; implementation may begin. The private phase remains private until the owning workflow and required behavioral assertions are satisfied.


## Private native issuance final and compatibility native reviews, 6c695b5 to 924c6c5

Message Type: FINAL_ANSWER
Task name: /root
Sender: /root/otp_issuance_execution_compat_6c695b5
Payload:
APPROVE for this intermediate development compatibility batch. No material compatibility findings.

Reviewed base `6c695b5703b6a345cd61ed34ce3e377a56ce063c` plus the seven specified working-tree files.

- No exported API additions/removals; issuance execution and construction remain private.
- Dependency edges remain acyclic: OTP PostgreSQL → delivery PostgreSQL → kernel PostgreSQL. Neither added dependency points back to OTP.
- Go minimum remains `1.27.0`; no permanent `replace`.
- Both new dependency pins identify actual committed source at `6c695b5`, whose commit timestamp matches `20261007080438`. Inspected dependency source is unchanged from that commit.
- Scalar SQL preserves the existing array storage, ordinal history, row lock, version CAS, rolling window and cooldown. Kernel `work_values.go` admits the actual scalar arguments and nullable timestamp destinations; no coordinator codec expansion.
- Existing storage integration tests cover rolling-history pruning, live-prefix preservation, concurrency and rollback. New integration tests compose real encrypted delivery and check original challenge compatibility with existing verification.

Unverified boundaries: I executed no code. The reported integration/race result was supplied by the parent, not independently rerun. Published dependency resolution, standalone clean-consumer builds, release gates and public Send/risk/recipient/same-code workflows remain unproven; approval makes no completion or release claim.

SHA-256:
```text
CHANGELOG.md                            4c0ded1039421a58ee3e85bbff55de8457fdf0feb48737fc96bf3530569e62ee
README.md                               4712c2a4212361b5f52be1b2e725a66929a2e29a1cb7c8f34c7f6b50fca7053a
go.mod                                  654db3ff5e700cd5bbb42fe5fb053a41cfc5373d208baf35928a1191a60c369d
send_window.go                          f13febf79ceaea48578eeab245cfe1e5d507de404a0d7b8c49c8c6d9560f3c86
issuance_execution.go                   ece7e93116661635a9884b4140733ea509dfc5553918a8d0d530d9aadfa2f90a
issuance_execution_test.go              c55894349a74eca72699364f7e54b33caa238be3b69ec10d0c4f6850ad5f6a87
issuance_execution_integration_test.go  86ee80e95505e1b1351fb9c03ed7c0325f720c5a0714e9c7a2e28b90a952d932
```
All paths are under `otp/postgres/`.
Message Type: FINAL_ANSWER
Task name: /root
Sender: /root/otp_issuance_execution_final_6c695b5
Payload:
**Internal disposition: APPROVE — scoped local development batch. No findings.**

Reviewed `feature/identity-platform`, base/head `6c695b5703b6a345cd61ed34ce3e377a56ce063c`, plus all seven dirty/untracked files. Single-reviewer read-only pass; no execution, services, edits, or delegation.

Source closure:
- Exact feature-instance invocation and authenticated original custody dominate writes. User status/security epochs, directory, optional credential and authorization checks use existing primary owners.
- Session/pre-auth variants remain disjoint. Flow/challenge validity is rechecked after family and predecessor waits; errors propagate into coordinator rollback.
- Scalar history preserves ordinal order, empty-history representation, ten-entry bound, cooldown and version CAS. Parameters remain inside Work’s closed codec and 64-parameter bound.
- Challenge, replacement, allowance, encrypted delivery, audit and lifecycle share coordinator transaction ownership. Same-command replay remains journal-owned.
- Recovery challenges: a revoked/expired retained challenge still prevents a negative vote; missing delivery payload cannot override that evidence. Schema prohibits deleting challenge custody.
- Prepared delivery mismatch/foreign-instance challenges fail at validation or exact staging ownership. No incomplete public Send constructor or caller-selected risk admission is exposed.
- Dependency pseudo-versions match the committed prerequisite revision and timestamp.

Supplied runtime evidence supports this development slice; I did not independently execute it. The additional updated-pin focused test passed. Broader exact coverage/mutation, clean published consumer, supply-chain/release and complete public Send/login boundaries remain unproven—not approved by this report. New integration cases exercise pre-auth user/anonymous execution; authenticated-session issuance execution and recovery fault injection are not independently demonstrated by those new cases.

Reviewed SHA-256 values, unchanged at final inspection:

```text
otp/postgres/CHANGELOG.md
4c0ded1039421a58ee3e85bbff55de8457fdf0feb48737fc96bf3530569e62ee
otp/postgres/README.md
4712c2a4212361b5f52be1b2e725a66929a2e29a1cb7c8f34c7f6b50fca7053a
otp/postgres/go.mod
654db3ff5e700cd5bbb42fe5fb053a41cfc5373d208baf35928a1191a60c369d
otp/postgres/send_window.go
f13febf79ceaea48578eeab245cfe1e5d507de404a0d7b8c49c8c6d9560f3c86
otp/postgres/issuance_execution.go
ece7e93116661635a9884b4140733ea509dfc5553918a8d0d530d9aadfa2f90a
otp/postgres/issuance_execution_test.go
c55894349a74eca72699364f7e54b33caa238be3b69ec10d0c4f6850ad5f6a87
otp/postgres/issuance_execution_integration_test.go
86ee80e95505e1b1351fb9c03ed7c0325f720c5a0714e9c7a2e28b90a952d932
```


## Prerequisite exact-head native delivery audit, 6c695b5

Message Type: FINAL_ANSWER
Task name: /root
Sender: /root/otp_delivery_prereq_delivery_6c695b5
Payload:
**Disposition: APPROVE — bounded local committed prerequisite claim only.**

- Verified branch `feature/identity-platform`, HEAD `6c695b5703b6a345cd61ed34ce3e377a56ce063c`, and exactly 21 committed paths against `d613644`.
- Committed-blob scope digest matches both native post-fix approvals: `f55d985e20c846e98784c42e0eec5591807799303f0e9df21fb49c593c2539b1`. Accepted native regression and complete final-diff reports; the earlier `REQUEST_CHANGES` receipt defect is addressed by the post-fix reviewed snapshot. Older compatibility approval supports its unchanged scope, not independent approval of the later receipt change.
- Prerequisite directories match HEAD; committed diff passes whitespace validation. Seven OTP consumer files remain dirty and excluded; both dependency pins now reference `6c695b5703b6`.
- No branch upstream or matching remote branch exists. No pushed, CI, released, production, clean-worktree, unit-verification, or complete-product claim is supported.
- Execution passes and resource cleanup remain **parent-supplied evidence**, not independently authenticated here. Retained scratch directory exists and remains needed for consumer work.

No unresolved delivery finding for this narrow claim. Consumer approval, exact coverage/mutation and affected release gates remain pending; no rerun is warranted solely because the reviewed snapshot became a commit.


## Prerequisite post-fix native final review, d613644 to 6c695b5

Message Type: FINAL_ANSWER
Task name: /root
Sender: /root/otp_delivery_prereq_final_receipt_d613644
Payload:
**Internal disposition: APPROVE — intermediate local development batch only.**

Reviewed `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`, branch `feature/identity-platform`, base/head `d613644a3af2ce42ae4c6f3a8e1d998fa2380e83`, plus all **21 specified dirty files**, including the untracked anonymous-management integration test.

Scope SHA-256: `f55d985e20c846e98784c42e0eec5591807799303f0e9df21fb49c593c2539b1`, calculated over lexically sorted repository-relative `path NUL file-bytes NUL` records.

**Findings:** no unresolved code defects or intermediate-commit readiness gaps.

Source-backed closure and falsification:
- `postgres/user_admission.go:36–50` requires the exact live native participant through `FeatureScope`; foreign/nil owners and live failures poison Work. Returned invocation fails before poisoning. Its unchanged primary helper at `:53–99` uses tenant/user-scoped `FOR SHARE`, compares all three full-width original epochs and active status, and excludes aggregate version and credential fabrication. Challenged stale identifier epoch, blocked/absent user, open cursor and retained lock; the expanded unit/integration tests independently assert denial, rollback, lock retention and zero credential references.
- `delivery/postgres/custody.go:101–134` preserves the old user-profile map unchanged and introduces a distinct anonymous profile binding exact flow, tenant, purpose and original command. `binding.go:38–130` bounds framing and requires canonical reconstruction. Challenged same-text user/flow substitution, changed flow/tenant, malformed field set and missing retained key; tests reject substitution or fail closed.
- `binding.go:150–159` is the shared user-only namespace guard. Status, cancellation, receipt, reconciliation and reconciliation recovery call it before disclosing results or performing their protected effect. In particular, `receipt_public.go:191–230` authenticates signed evidence and policy first, then rejects anonymous namespace **before** provider/configuration comparison or command reservation. The config-1/config-2 test independently asserts identical NotFound, empty result, unchanged command count and custody.
- `schema.go:22–34,40–60` retains bounded AAD/ciphertext, unique effect identity and immutable binding. Existing persistence verifies custody before SQL and inserts intent/outbox through the same Work. No new migration, transaction owner or unbounded I/O loop.
- `attempt.go:226–269` authenticates recovered custody locally, commits its fenced lease, then decrypts outside the transaction. The new public-worker test checks original recipient/body/expiry, freely obtainable row lock during Sender, terminal erasure and no second send. Existing worker lease, audit, retry and cleanup consumers retain the namespace-capable decoder.
- Both module changelogs and documentation describe the additive behavior and explicitly require reader/worker upgrade before anonymous writes; older-reader rollback requires draining/disabling anonymous delivery.

**Verification boundary:** execution was prohibited for this reviewer. I inspected test sources and used the supplied RED/GREEN, affected race/integration, full adapter/kernel and parser-fuzz results; I did not independently rerun them.

OTP dirty files were direct-consumer context only, **excluded from approval**. Its old kernel pin still needs the subsequent actual prerequisite pin/integration step. Exact coverage, mutation, clean published-consumer, remaining release gates and complete-product verification remain pending; this approval certifies neither standalone OTP readiness nor PR/release readiness.


## Prerequisite post-fix native regression, d613644

Message Type: FINAL_ANSWER
Task name: /root
Sender: /root/otp_delivery_prereq_regression_receipt_d613644
Payload:
**Internal disposition: APPROVE for the scoped development batch.** No concrete regression or ineffective new oracle found.

- Role: fresh read-only regression coverage auditor using `testing-coverage`.
- Base/head: `d613644a3af2ce42ae4c6f3a8e1d998fa2380e83`, plus the 21 dirty files under `delivery/postgres` and `postgres`; OTP changes excluded except consumer inspection.
- Scope SHA256: `f55d985e20c846e98784c42e0eec5591807799303f0e9df21fb49c593c2539b1` (sorted path + NUL + current file bytes + NUL).

**Behavioral assertions**
- Native feature admission tests cover exact versus foreign/nil participant, expired invocation, poisoned live failures, each independent security epoch, inactive/missing users, aggregate-only version changes, full-width epochs, cancellation, open cursors, transaction rollback and retained share locks. They assert outcomes and durable effects without fabricating credential references.
- Anonymous custody tests authenticate namespace and exact flow, reject malformed profiles and user/flow substitution, restore original plaintext, and seed the bounded canonical parser fuzz target.
- Real-PG namespace-collision tests cover status/versioned status, cancellation, receipt persistence, reconciliation, recovery and discovery. Signed public receipt callbacks use both matching configuration 1 and valid mismatched configuration 2, requiring empty `NotFound`, no command reservation and unchanged state/custody/receipt count.
- Restart integration uses a fresh codec and public worker, asserts original recipient/body/expiry, no held transaction/row lock during sender execution, one send, terminal erasure and no resend.
- Public receipt source performs authentication and policy before namespace inspection; namespace rejection precedes provider/configuration comparison and reservation. Existing public receipt tests retain forged-signature, denied-policy and no-reservation assertions.

**Falsification**
Removing the public receipt namespace guard lets valid configuration 2 reach `Forbidden`; the new oracle rejects that result. Parent-supplied RED `8e9f01` records this exact failure. Removing the other namespace guards would expose state or reach mutation/provider logic instead of the asserted `NotFound`.

**Limits**
No tests, services or mutation tooling executed by this auditor. Parent-supplied GREEN `553928`, earlier full integration/race runs and fuzz results are execution evidence supplied by the parent, not independently rerun here. Exact release coverage/mutation, published clean-consumer and complete platform acceptance remain unproven. User-profile preservation is supported by unchanged user-profile construction in the diff and existing round-trip/composition assertions, not a new historical ciphertext fixture.


## Prerequisite final review and private OTP execution regression, d613644

Message Type: FINAL_ANSWER
Task name: /root
Sender: /root/otp_issuance_execution_regression_d613644
Payload:
**Disposition: APPROVE — bounded regression-coverage audit; no concrete reachable defect found.**

Reviewed `feature/identity-platform`, base/head `d613644a3af2ce42ae4c6f3a8e1d998fa2380e83`, plus the six selected dirty OTP files. This approves the private transaction phase, not public Send, package verification, or release readiness.

Meaningful behavioral oracles:
- `issuance_execution_integration_test.go:158–228` uses real coordinator/PostgreSQL execution, independently changes primary authority epochs/status/flow, induces actual delivery/audit failures, and asserts challenge/send/effect/lifecycle rollback, stable replay, loadable digest challenge and unconsumed pre-auth state.
- `issuance_execution_test.go:75` rejects altered frozen delivery declarations and unowned Work.
- Existing `issuance_races_integration_test.go:17` asserts replacement/retention, insertion-failure rollback, unresolved reservation preservation, collisions and one resend winner.
- Existing send-window tests exercise cooldown/hour boundaries, history preservation, primary SQL guards and concurrency. These meaningfully cover the scalar-history rewrite; they are not source-text or mock-choreography assertions.

**Evidence limits:** I inspected source/tests only and did not execute anything. Parent-reported integration/race passes and red/green reproduction were supplied, not independently verified from raw output. No exact coverage/mutation or clean-published-consumer claim is established.

Nonblocking remaining partitions for subsequent public Send verification: direct issuance-session execution, post-predecessor-lock time expiry, and issuance-specific recovery votes with retained challenge tombstones. Existing lower/shared-boundary tests and fail-closed source support this private batch, but must not be represented as direct execution evidence for those new paths.

Reviewed source hashes:
- `issuance_execution.go`: `ece7e93116661635a9884b4140733ea509dfc5553918a8d0d530d9aadfa2f90a`
- `issuance_execution_test.go`: `c55894349a74eca72699364f7e54b33caa238be3b69ec10d0c4f6850ad5f6a87`
- `issuance_execution_integration_test.go`: `86ee80e95505e1b1351fb9c03ed7c0325f720c5a0714e9c7a2e28b90a952d932`
- `send_window.go`: `f13febf79ceaea48578eeab245cfe1e5d507de404a0d7b8c49c8c6d9560f3c86`
Message Type: FINAL_ANSWER
Task name: /root
Sender: /root/otp_delivery_prereq_final_d613644
Payload:
## REQUEST_CHANGES

Reviewed `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`, branch `feature/identity-platform`, `BASE=HEAD=d613644a3af2ce42ae4c6f3a8e1d998fa2380e83`, plus the exact 20-file dirty snapshot below. Single read-only reviewer; no execution, services, writes, or delegation. Dirty `otp/postgres` was consumer context only.

### Blocker: public receipt admission still crosses the anonymous namespace

[receipt_public.go:215](/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform/delivery/postgres/receipt_public.go:215)

`RecordReceipt()` still selects an intent's provider/configuration by `subject_id` text without checking its retained namespace. With an anonymous flow named `user-a`, a valid user-only receipt binding targeting `UserID="user-a"` reaches that row. A valid signed callback with another retained configuration returns provider-dependent `ErrForbidden`, whereas an absent target returns `ErrNotFound`; matching configuration proceeds into coordinator reservation before the new private guard rejects the anonymous effect.

**First divergence:** the public target lookup omits `binding_aad`/`decodeUserBinding()`, unlike the updated reconciliation admission paths. The later private rejection prevents receipt mutation, but does not prevent provider discovery or command admission.

**Falsification attempted:** traced `ValidateFor()`, signed-callback authentication, actor/tenant/intent matching, and current policy. These validate the user-side request and callback, not the stored intent's namespace. The private guard is reached only after reservation and its `ErrNotFound` becomes `identity.ConflictError` in [receipt_native.go:108](/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform/delivery/postgres/receipt_native.go:108); it therefore does not disprove the public-boundary failure.

**Fix:** check the retained user binding immediately after public target selection and before provider/configuration comparison or reservation. Add a real public `RecordReceipt()` regression using valid signed callbacks for equal user/flow text, with matching and mismatched retained provider configurations; assert uniform `ErrNotFound`, empty result, unchanged custody/status/receipt state, and no new command admission. The new [anonymous management test:103](/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform/delivery/postgres/anonymous_management_integration_test.go:103) exercises the private receipt store, not this public path.

### Other reviewed boundaries

- `Work.RequireActiveUserAuthority()` uses exact live participant ownership, original command user/tenant, existing primary-row share locking and full-width security epochs. Aggregate versions remain separate; live failures poison Work.
- Anonymous custody has an explicit authenticated profile/flow field. Canonical parsing and local envelope verification reject changed flow, tenant, or namespace. Existing user-profile construction remains unchanged.
- Cancellation, status, private receipt/reconciliation effects, and public reconciliation/recovery now reject equal-text anonymous targets before exposing results or applying effects.
- Persisted-worker reconstruction authenticates original custody; decryption/sending follows committed leasing outside the transaction. The added restart test independently checks original recipient/body/expiry, released locks, single delivery, and terminal erasure.
- Parsing/storage limits and immutable retained-binding enforcement remain compatible with the additional field.

Supplied focused/full integration-race and fuzz results were considered execution evidence, not rerun. No additional source-confirmed findings. Public receipt isolation remains unproved and source-confirmed defective. Release, coverage/mutation completeness, clean-consumer/publication, and actual OTP dependency pins remain pending outside this development-batch disposition.

### Frozen SHA-256 snapshot

```text
fbb2e0d909c2aad1fec9e23a61d656e477bf98e5258b387a9035e6658da3e482  delivery/postgres/CHANGELOG.md
88e4d00e234c16eab921138c57c70a1cda3530618ed96d55c6d1bb4318ce1c42  delivery/postgres/README.md
97f01b20fc71267b9ec71728022e31a5a29b67c6acbcc7f3dd5871cab2176172  delivery/postgres/anonymous_management_integration_test.go
53434abd1e41c650a2fa4a97867f9e01e916f4de331eec25864efbe32a270486  delivery/postgres/binding.go
0545c4a470c48d6a82b5688943c6bd1c8ffdd565130fac30f3a84b7e32e1f02f  delivery/postgres/binding_test.go
23f63b66ad5d569c68d269559bba99e9c9d89f747b32cd70e363cbe6de43f068  delivery/postgres/cancel.go
a949c688c3b49ef5b47abf71ad0ef395863edfd3ff629a5a108167b1a40bb9bd  delivery/postgres/cancel_integration_test.go
350f36c65f5b15af2da7fd70e108a040021052462a8be00fc4046b22ace937ee  delivery/postgres/composition_integration_test.go
aa3aa1fcd0fcf0f68ee746fbacebf81a7c65cbf9ffa646c011924e47c32eecbd  delivery/postgres/contributor_test.go
d8d5bcae6c24409a52fdf581a1861ee9559ee9d7fda9d803d5290ffa39bb207b  delivery/postgres/custody.go
3c1aad272127efaaef01a593de5e072575baad863a823a8561b91327539657a1  delivery/postgres/receipt.go
f2374020dc8094c7d08379c975784980090ccf7ffc35f98ebc9e7a525c19cf31  delivery/postgres/reconcile.go
f537d3a173da38939559c428b9d499b6a690f7233f054784ce8fca4a85c158f0  delivery/postgres/reconcile_public.go
be0addc7f8e7f64c42e401d722326cbe9230896b87065dcb15d24be6496897aa  delivery/postgres/reconcile_recovery.go
fbffb73f46ddd5a3cada8139ed9d67a69de248db3fd66fb0cb9bbfb70ce89cb8  delivery/postgres/status.go
a8beceefb00671e3b99c2f4e52187c8dacb186ac0607109fbd9d2359cb7b97e5  postgres/CHANGELOG.md
5b64945bdcd358d4f8444eab7af544078853cc4fd8c9f084a69b4415190f1832  postgres/README.md
913a23903dd6481cc595f715bdf99fa3795105064601857a3144d2783068a0c9  postgres/otp_user_admission_integration_test.go
9517ad9b5f593b3b27b6307a81b7004f25f10f8224a05429bd58940072d5ac66  postgres/otp_user_admission_test.go
9c9e41aeaa2a46c64c9e99ae194e67475073c2a9644d09752f48ea5ed025f929  postgres/user_admission.go
```


## Kernel and anonymous delivery prerequisite native regression and compatibility, d613644

Message Type: FINAL_ANSWER
Task name: /root
Sender: /root/otp_delivery_prereq_regression_d613644
Payload:
## APPROVE — scoped prerequisite regression audit

**Principal:** `/root/otp_delivery_prereq_regression_d613644`, fresh read-only `regression_reviewer`; used `testing-coverage`. No edits, execution, services, delegation, or publication.

**Repository:** `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`  
**Branch:** `feature/identity-platform`  
**Base/HEAD:** `d613644a3af2ce42ae4c6f3a8e1d998fa2380e83`, plus exactly the 18 prerequisite files hashed below.

No concrete regression defect found within live feature user-epoch admission, explicit anonymous delivery custody, worker consumption, and user-only management isolation.

- Admission tests extend all three independently stale epochs, inactive/missing users, aggregate-only changes, full-width values, retained share locks, lock deadlines, cursor rejection, foreign owners, and ignored-denial poisoning to the feature invocation. Successful admission asserts durable effects without credential references.
- Public anonymous preparation exercises the actual contributor boundary and rejects missing/mixed/malformed subjects, invalid link mode, and mismatched intent target.
- Custody tests restore the original complete binding and plaintext and reject changed flow, namespace, tenant, and malformed profiles. The anonymous seed joins the canonical parser fuzz oracle.
- Real PostgreSQL tests deny equal-text anonymous/user crossover through status, versioned status, cancellation, receipt, reconciliation effect/discovery/recovery, while asserting unchanged ciphertext/state and absent receipts.
- A fresh codec and public worker assert the exact original recipient/body/expiry, sender invocation outside transactions and locks, one send, no repeated claim, and terminal ciphertext erasure.
- The dirty OTP consumer—not part of this batch—exercises the helper through the real coordinator, anonymous commit/delivery rollback, user authority denials, persisted challenge/effect/lifecycle consistency, and stable replay.

**Evidence boundaries:** parent-reported race/integration passes `7464f0`, `e4e308`, and `260db2`, and earlier RED results, were not independently executed or authenticated here. Existing user AAD construction is structurally unchanged; no dedicated frozen pre-change AAD golden fixture was added. Anonymous-specific coordinator recovery and extended fuzz execution remain outside the supplied execution evidence; existing recovery reconstructs and compares original AAD, with no observed defect. This approves neither full OTP Send nor package/release/platform completion, clean published dependency resolution, or remaining release gates. Later documentation changes require their own final snapshot review.

```text
fbb2e0d909c2aad1fec9e23a61d656e477bf98e5258b387a9035e6658da3e482  delivery/postgres/CHANGELOG.md
53434abd1e41c650a2fa4a97867f9e01e916f4de331eec25864efbe32a270486  delivery/postgres/binding.go
0545c4a470c48d6a82b5688943c6bd1c8ffdd565130fac30f3a84b7e32e1f02f  delivery/postgres/binding_test.go
23f63b66ad5d569c68d269559bba99e9c9d89f747b32cd70e363cbe6de43f068  delivery/postgres/cancel.go
a949c688c3b49ef5b47abf71ad0ef395863edfd3ff629a5a108167b1a40bb9bd  delivery/postgres/cancel_integration_test.go
350f36c65f5b15af2da7fd70e108a040021052462a8be00fc4046b22ace937ee  delivery/postgres/composition_integration_test.go
aa3aa1fcd0fcf0f68ee746fbacebf81a7c65cbf9ffa646c011924e47c32eecbd  delivery/postgres/contributor_test.go
d8d5bcae6c24409a52fdf581a1861ee9559ee9d7fda9d803d5290ffa39bb207b  delivery/postgres/custody.go
3c1aad272127efaaef01a593de5e072575baad863a823a8561b91327539657a1  delivery/postgres/receipt.go
f2374020dc8094c7d08379c975784980090ccf7ffc35f98ebc9e7a525c19cf31  delivery/postgres/reconcile.go
f537d3a173da38939559c428b9d499b6a690f7233f054784ce8fca4a85c158f0  delivery/postgres/reconcile_public.go
be0addc7f8e7f64c42e401d722326cbe9230896b87065dcb15d24be6496897aa  delivery/postgres/reconcile_recovery.go
fbffb73f46ddd5a3cada8139ed9d67a69de248db3fd66fb0cb9bbfb70ce89cb8  delivery/postgres/status.go
97f01b20fc71267b9ec71728022e31a5a29b67c6acbcc7f3dd5871cab2176172  delivery/postgres/anonymous_management_integration_test.go
a8beceefb00671e3b99c2f4e52187c8dacb186ac0607109fbd9d2359cb7b97e5  postgres/CHANGELOG.md
913a23903dd6481cc595f715bdf99fa3795105064601857a3144d2783068a0c9  postgres/otp_user_admission_integration_test.go
9517ad9b5f593b3b27b6307a81b7004f25f10f8224a05429bd58940072d5ac66  postgres/otp_user_admission_test.go
9c9e41aeaa2a46c64c9e99ae194e67475073c2a9644d09752f48ea5ed025f929  postgres/user_admission.go
```
Message Type: FINAL_ANSWER
Task name: /root
Sender: /root/otp_delivery_prereq_compat_d613644
Payload:
**APPROVE — bounded local Go compatibility prerequisite.**

Principal: `/root/otp_delivery_prereq_compat_d613644`. Reviewed base/HEAD `d613644a3af2ce42ae4c6f3a8e1d998fa2380e83`, complete dirty delivery adapter batch and kernel admission batch, including newly added README changes. No concrete compatibility defect found.

- `postgres/user_admission.go:35` adds a concrete `Work` method without expanding a contributor interface or changing existing OTP-phase methods. Exact live participant admission uses existing `FeatureScope`; credential-specific admission remains unchanged.
- `delivery/postgres/custody.go:101` preserves the original user field set, values, profile and canonical encoding. Encryption framing, custody/logical HMAC domains and payload encoding are unchanged; retained user ciphertext/AAD/logical-digest compatibility follows from unchanged byte inputs.
- Anonymous custody selects a distinct 29-field profile with authenticated `anonymous_flow`, empty user and matching subject. `binding.go:64` accepts only the explicit profiles, then reconstructs and compares the complete canonical field set. Older readers cannot consume new anonymous records; deploy updated readers before writing them. Existing user records remain readable.
- User-only management checks immutable binding namespace before returning results or mutating effects. Existing public enqueue already requires `SubjectID == UserID`; equal user/flow text cannot borrow management authority.
- Public-worker restart consumption is exercised in the new integration source using retained original payload, transaction-free sender execution and terminal ciphertext erasure.
- No new module dependency, cycle, permanent override or language/library construct incompatible with declared Go `1.27.0` identified.

**Limits:** Source-only audit; no tests/builds/services executed. Parent-supplied integration/race passes used disposable workspace overrides and do not establish published-pin or standalone-consumer compatibility. Coverage, mutation, release gates and full OTP completion remain unproven. Dirty `otp/postgres` implementation is excluded from approval. Broad evidence search exposed older sibling-report excerpts; disposition is based on the inspected current source, not their conclusions.

Reviewed SHA-256:
```text
fbb2e0d909c2aad1fec9e23a61d656e477bf98e5258b387a9035e6658da3e482 delivery/postgres/CHANGELOG.md
88e4d00e234c16eab921138c57c70a1cda3530618ed96d55c6d1bb4318ce1c42 delivery/postgres/README.md
97f01b20fc71267b9ec71728022e31a5a29b67c6acbcc7f3dd5871cab2176172 delivery/postgres/anonymous_management_integration_test.go
53434abd1e41c650a2fa4a97867f9e01e916f4de331eec25864efbe32a270486 delivery/postgres/binding.go
0545c4a470c48d6a82b5688943c6bd1c8ffdd565130fac30f3a84b7e32e1f02f delivery/postgres/binding_test.go
23f63b66ad5d569c68d269559bba99e9c9d89f747b32cd70e363cbe6de43f068 delivery/postgres/cancel.go
a949c688c3b49ef5b47abf71ad0ef395863edfd3ff629a5a108167b1a40bb9bd delivery/postgres/cancel_integration_test.go
350f36c65f5b15af2da7fd70e108a040021052462a8be00fc4046b22ace937ee delivery/postgres/composition_integration_test.go
aa3aa1fcd0fcf0f68ee746fbacebf81a7c65cbf9ffa646c011924e47c32eecbd delivery/postgres/contributor_test.go
d8d5bcae6c24409a52fdf581a1861ee9559ee9d7fda9d803d5290ffa39bb207b delivery/postgres/custody.go
3c1aad272127efaaef01a593de5e072575baad863a823a8561b91327539657a1 delivery/postgres/receipt.go
f2374020dc8094c7d08379c975784980090ccf7ffc35f98ebc9e7a525c19cf31 delivery/postgres/reconcile.go
f537d3a173da38939559c428b9d499b6a690f7233f054784ce8fca4a85c158f0 delivery/postgres/reconcile_public.go
be0addc7f8e7f64c42e401d722326cbe9230896b87065dcb15d24be6496897aa delivery/postgres/reconcile_recovery.go
fbffb73f46ddd5a3cada8139ed9d67a69de248db3fd66fb0cb9bbfb70ce89cb8 delivery/postgres/status.go
a8beceefb00671e3b99c2f4e52187c8dacb186ac0607109fbd9d2359cb7b97e5 postgres/CHANGELOG.md
5b64945bdcd358d4f8444eab7af544078853cc4fd8c9f084a69b4415190f1832 postgres/README.md
913a23903dd6481cc595f715bdf99fa3795105064601857a3144d2783068a0c9 postgres/otp_user_admission_integration_test.go
9517ad9b5f593b3b27b6307a81b7004f25f10f8224a05429bd58940072d5ac66 postgres/otp_user_admission_test.go
9c9e41aeaa2a46c64c9e99ae194e67475073c2a9644d09752f48ea5ed025f929 postgres/user_admission.go
```


## Anonymous delivery prerequisite native mapping and bounded QA, d613644

Message Type: FINAL_ANSWER
Task name: /root
Sender: /root/otp_anonymous_delivery_map_d613644
Payload:
Native mapper report — read-only, baseline `d613644a3af2ce42ae4c6f3a8e1d998fa2380e83`, current dirty tree inspected. No writes, tests, services, or delegates.

**Confirmed first divergence:** `Contributor.Prepare` → `validateInputs` → `validateInputShape` → `custodyContext` (`delivery/postgres/custody.go:106`) requires nonempty `Command.UserID`. This rejects legitimate explicit anonymous-feature bindings before encryption. Kernel `CommandBinding` documents mutually exclusive `UserID`/`AnonymousFlowID`; `ValidateFor` enforces that union according to `FeatureOperation.SubjectRequirement` (`command.go:62`, `feature_operation.go:83`). Anonymous flow is a target, never actor/authentication authority.

| Ownership seam | Relevant contract / affected consumers |
|---|---|
| Custody preparation | `custodyContext` owns authenticated scope and logical fingerprint input. Currently encodes `user` but entirely omits `AnonymousFlowID`. Merely relaxing nonempty-user validation would lose flow binding and permit cross-flow substitution. |
| Durable binding parser | `binding.go:38` requires exactly 28 fields, v1 profile, then exact canonical reconstruction. Anonymous metadata requires coordinated encoder/decoder support; unknown/missing fields must still fail closed. |
| Plan validation / commit | `persistence.go:76–144` reconstructs the same custody context for `ValidateDeliveryPlan`, envelope verification, and installed AAD. `Stage` already compares the complete `CommandBinding`; no alternate transaction required. |
| Recovery | `persistence.go:275–335` reconstructs retained original-binding AAD and compares it byte-for-byte. Scope encoding must agree with installation and preserve anonymous target identity. |
| Worker / retained reads | `lease.go`, `enqueue.go`, and `worker_audit.go` consume `decodeBinding`; restore/open also verifies custody. Parser support must survive real lease, send, retry, replay, and audit paths—not just Prepare. |
| OTP native issuance | Dirty `otp/postgres/issuance_execution.go:68–75` explicitly requires delivery `SubjectID` to equal the selected user or anonymous flow and expiry to equal the challenge expiry. It stages ciphertext through the original Work. This is the consumed anonymous requirement. |

**Compatibility implications**
- Preserve existing user v1 AAD bytes and logical digests. Changing every user encoding would invalidate retained ciphertext/recovery evidence even though user semantics are unchanged.
- Anonymous encoding must distinguish namespace and authenticate the exact flow. Reject empty-both, nonempty-both, unsafe flow IDs, invalid link modes, and malformed profile/field combinations.
- SQL `delivery_intents.subject_id` is a bounded `identity.directory_id`, not a user foreign key (`schema.go:14`). Existing storage can hold a canonical flow target; no schema migration is evidenced as necessary.
- Deterministic effect ID already uses deployment/tenant/command/class/index, not user (`persistence.go:61`). Its allocation algorithm need not change.
- Kernel `DeliveryIntentPlan` accepts a bounded generic `SubjectID`; kernel feature encoding already retains `anonymous_flow_id`. No public kernel contract expansion is required by this divergence.

**Intentionally user-only variants**
Public `EnqueueService` remains user-bound: `admission.go:323` validates `EnqueueOperation`; recipient resolution/no-op digest use `UserID`, and `enqueue_execution.go:108` authorizes an `identity.user` resource. Status, cancellation, reconciliation, and their management recovery also require user scope and constrain SQL by `subject_id = UserID`. Do not relax those surfaces as an incidental consequence of native custody support. Worker processing has its own service-account authorization and is not a user management endpoint.

**Minimal affected slice:** shared private custody union validation/context plus durable decoder; regression coverage across Prepare/seal/open, persistence/recovery, and leased worker decode, with existing user fixtures unchanged. OTP issuance owns the flow/intent matching invariant; public enqueue and management APIs remain closed user-only variants. Parent must verify the concrete implementation and executed behavior; this map is source evidence only.
Message Type: FINAL_ANSWER
Task name: /root
Sender: /root/otp_anonymous_delivery_qa_d613644
Payload:
**Role:** fresh read-only `qa_test_planner`; complete `$qa-test-planner` read.  
**Reviewed boundary:** `d613644a3af2ce42ae4c6f3a8e1d998fa2380e83` plus current OTP issuance/helper dirty files. No edits, Go execution, services, or subagents.

**Contract and finding:** anonymous OTP issuance must encrypt its original rendered code, atomically persist challenge and delivery intent, and restore the exact immutable anonymous flow for worker execution/recovery. `Prepare` currently fails in `custodyContext` because `UserID` is mandatory (`custody.go:106`). The shared persisted decoder accepts exactly the 28-field user-only v1 profile and never reconstructs `AnonymousFlowID` (`binding.go:43–67`). Relaxing only user validation is insufficient: anonymous identity must enter authenticated custody and survive restoration. Preparation itself remains neither authentication nor recipient authorization.

### Smallest sufficient coverage

| Boundary | Observable assertion / defect caught |
|---|---|
| Public preparation, red first | Real contributor with empty `UserID`, valid positive `AnonymousFlowID`, genuine OTP binding and matching anonymous intent returns a usable frozen plan; currently must fail at the actual owner boundary. Do not fabricate a placeholder user. |
| Subject-shape partitions | User only and anonymous only succeed; neither, both, malformed/oversized flow fail with no usable preparation. Preserve bounded plaintext/envelope checks and cancellation/expiry behavior using existing coverage rather than duplicating the matrix. |
| Authenticated anonymous custody | Seal → canonical persisted AAD → fresh codec restore → open preserves full original binding and exact recipient/rendered code. Changing flow, subject namespace, tenant, command, or intent subject cannot authenticate/open the original envelope. Reusing unchanged ciphertext under another flow must fail even when effect identity is unchanged. |
| Installed user compatibility | Existing user v1 AAD bytes round-trip unchanged and authenticate using retained keys after the change. Unknown profiles, missing/extra fields, malformed flow, mixed user/flow, and noncanonical framing fail closed. Add one valid anonymous seed to the existing parser fuzz test. |
| Original coordinator atomicity | Genuine anonymous OTP issuance commits one challenge, family counter transition, journal/audit/lifecycle record, delivery intent and delivery outbox row together. Induced delivery persistence/outbox failure leaves none of those changes. Replay returns original stable result/effect without generating or persisting another credential. |
| Restart and worker | Load committed anonymous AAD/ciphertext with a fresh codec/worker; sender receives the original destination and exact original code after lease commit, outside its transaction. Tampered flow cannot reach sender. Existing terminal completion erases sensitive ciphertext and preserves recorded outcome. |
| Recovery | Original anonymous command/flow reconciles committed issuance without regeneration or another send. Wrong flow/user namespace cannot recover that command. Existing unknown/noncommitted distinctions remain observable. |
| User-only management isolation | Existing public enqueue, status, cancellation, receipt and reconciliation entry points continue requiring their original user authority. A user whose ID text equals an anonymous flow cannot read or mutate the anonymous intent. Empty user or added anonymous flow must not become authorization for those management surfaces. |

**Execution recommendation:** first the focused public anonymous `Prepare` regression; then delivery custody/binding/persistence unit tests, existing parser fuzz coverage, and one bounded real-PostgreSQL anonymous issuance→restart→worker/recovery composition with rollback and namespace-isolation subcases. Reuse existing user management and worker regression suites for unchanged behavior. Execute affected repository gates separately as required; all Go commands need task-owned disposable `GOCACHE`.

**Disposition:** implementation input, not a blocking approval program. No contradictory requirement or unsafe requested action identified. Tests and integration behavior remain **not run** in this planning pass.

## Native issuance custody exact-head delivery, d613644

Message Type: FINAL_ANSWER
Task name: /root
Sender: /root/otp_issue_custody_delivery_d613644
Payload:
## APPROVE — locally committed issuance custody prerequisite only

Principal: `/root/otp_issue_custody_delivery_d613644`; read-only audit, no writes, tests, publication or delegation.

- **Exact head:** `d613644a3af2ce42ae4c6f3a8e1d998fa2380e83`, parent `3bb3e55e1a6c4764e84fe1f6dac81b245c6b75d5`.
- **Source:** clean `feature/identity-platform`; exactly nine committed paths; `git diff HEAD^ HEAD --check` passed.
- **Independent reviews:** inspected native corrected regression, compatibility and complete-final-diff reports. All approve the bounded custody/possession prerequisite. Every committed blob’s SHA-256 matches all three corrected reports.
- **Execution:** authenticated raw parent-session outputs `4ac9a4` and terminal `f6ede4`, plus underlying chained command: focused named-key regression passed **0.214s**, full non-integration adapter race passed **1.300s**, vet succeeded, and `FuzzIssuanceRecord` passed **6.253s / 102,191 executions**; combined terminal exit **0**. Local source composition used disposable workspace dependency overrides, not clean standalone resolution.
- **Cleanup:** all three specified scratch/message paths absent.

**Delivery states:** local and committed verified; no publication evidenced and no remote-tracking branch contains this head. Hosted CI, deployment and production verification are not established.

**Findings:** none against the scoped local-commit claim. Executable Send, current-primary admission, native issuance/encrypted-delivery composition, remote baseline availability, standalone consumer resolution, exact coverage/mutation and remaining release gates remain pending. This is not OTP, platform or release completion.

## Native issuance custody blind final review, base 3bb3e55

Message Type: FINAL_ANSWER
Task name: /root
Sender: /root/otp_issue_custody_final
Payload:
**APPROVE — No findings within the scoped development batch.**

Role: independent complete-final-diff correctness reviewer, single read-only session using `requesting-code-review`, `testing-coverage`, and `testing-integration`.

Reviewed repository: `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`, branch `feature/identity-platform`. Base and committed HEAD: `3bb3e55e1a6c4764e84fe1f6dac81b245c6b75d5`, plus the complete nine-file dirty patch below.

Source-backed closure:
- `issuance_custody.go:52-141`: only initial digest-only challenges enter issuance custody; exact scope/descriptor and original authority are retained. Separate root constructors preserve bearer custody without manufacturing submitted-code custody.
- `issuance_custody.go:155-223`: restoration rejects altered binding, duplicate/unknown OTP fields, malformed ciphertext, missing retained keys, and changed key material. The encrypted keyed scope commitment prevents same-version code-key reinterpretation.
- `issuance_custody.go:230-302`: local possession derives the digest from the separately submitted bearer under the original named key; wrong, absent, and swapped variants deny. Session expectations are cross-checked through the existing session owner.
- `issuance_record.go`: bounded closed serialization restores complete challenge policy, authorization variant, original timelines, and remember choice without code/bearer fields. Truncation, trailing data, invalid tags, and unsupported scalar states fail closed.
- `command.go` and `authority_custody.go`: existing attempt AAD bytes and accepted attempt record families remain unchanged; issuance uses separate domain/profile and record validation.

Falsification challenges: traced metadata-only commands, cross-variant bearer substitution, changed same-version code-key bytes, retained-key rotation, and challenge/binding mismatch. Each denies or preserves original custody as required. Neither private restoration nor possession checking performs SQL, admission, issuance, delivery, or renewal.

Verification: supplied focused red/green results, `go test . -race -count=1` (1.300s), `go vet ./...` (exit 0), and bounded `FuzzIssuanceRecord` (102,191 executions, terminal 0) support this local slice. I did not rerun tests or access services. `git diff --check` passed.

**Not a release-readiness verdict:** public Send composition, current-primary admission, real PostgreSQL issuance composition, standalone/remote clean-consumer resolution, coverage/mutation, and remaining release gates are unproved or pending. The committed root constructor baseline `93d9eaf592ae` was inspected locally.

Frozen SHA-256 values (`otp/postgres/`):
```text
16c682cbac400ca8028abacacf8a389987877f345f731869f012b528e5bf718d authority_custody.go
00511e25427db5aed5406556cbf342900f70f1e8843ca113e28766867f53fe93 command.go
e8022cf57697cf79632638ec11a8ac762f3f9ea2989e051fac992a2d2aedefb7 issuance_custody.go
7d63cbb0323224bd093d186e53cb5b37697c9748ed6457a64a846fce5bee9a70 issuance_record.go
0b251c2d6acb8a00a70e595ce331a5a27529df40e8bcd736393a30893ed6e53f issuance_custody_test.go
04e67badb1661dcfda6a22b17d7637fad8289a08ff300d58d23b5d76d2c9d36e issuance_record_test.go
1a299c6ca8d1ef81d86ad2fc44de3953b3e49f39c5ac991d924f48a45b45998c README.md
3e99e84970d8c7e6e2e7ba19c274fe6bced51c9e20402f95d27f3d3b5234e0bf CHANGELOG.md
9c1f870a67786e6ea16922f5f8da42b67f9c753f08e45db72b8c44e11ba02993 go.mod
```




## Native issuance custody corrected regression and compatibility, base 3bb3e55

Message Type: FINAL_ANSWER
Task name: /root
Sender: /root/otp_issue_custody_corrected_compatibility
Payload:
APPROVE — `go_compatibility`, read-only source audit.

Reviewed repository `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`, branch `feature/identity-platform`, base/HEAD `3bb3e55e1a6c4764e84fe1f6dac81b245c6b75d5`, plus exactly nine dirty paths below. No source-backed compatibility defect found in the bounded local issuance-custody/possession prerequisite.

- Public additions are two methods on existing `AttemptCodec`; existing API signatures remain unchanged.
- Existing attempt AAD retains its exact domain, profile and framing order. Existing authority decoding retains its previous accepted record prefixes; issuance restoration uses a separate strict prefix and AAD profile.
- Issuance framing preserves original digest/version, initial policy, remember choice, complete session/pre-auth projection and timestamps without code/bearer fields. Retained wrapping-key identity and authenticated scope commitment reject changed wrapping/code material; possession derives against the original commitment using its named retained session/pre-auth key.
- Root requirement `v0.0.0-20261007061102-93d9eaf592ae` matches actual local commit `93d9eaf592aebc6c069c488c3a6f5aff9b8ad5c1`, timestamp `2026-10-07T06:11:02Z`. That commit contains both additive code-free bearer constructors consumed here.
- Go directives remain `1.27.0`. No new dependency edge, adapter-to-core inversion or permanent replacement appears in the patch.

**Unverified boundaries:** remote availability of the pinned baseline, clean standalone consumer resolution, published release, complete release gates and executable native Send/delivery composition. Approval does not certify those. No builds, tests, services, writes or delegation performed; parent execution results were not substituted for source review.

SHA-256, paths relative to repository:

```text
16c682cbac400ca8028abacacf8a389987877f345f731869f012b528e5bf718d  otp/postgres/authority_custody.go
00511e25427db5aed5406556cbf342900f70f1e8843ca113e28766867f53fe93  otp/postgres/command.go
e8022cf57697cf79632638ec11a8ac762f3f9ea2989e051fac992a2d2aedefb7  otp/postgres/issuance_custody.go
7d63cbb0323224bd093d186e53cb5b37697c9748ed6457a64a846fce5bee9a70  otp/postgres/issuance_record.go
0b251c2d6acb8a00a70e595ce331a5a27529df40e8bcd736393a30893ed6e53f  otp/postgres/issuance_custody_test.go
04e67badb1661dcfda6a22b17d7637fad8289a08ff300d58d23b5d76d2c9d36e  otp/postgres/issuance_record_test.go
1a299c6ca8d1ef81d86ad2fc44de3953b3e49f39c5ac991d924f48a45b45998c  otp/postgres/README.md
3e99e84970d8c7e6e2e7ba19c274fe6bced51c9e20402f95d27f3d3b5234e0bf  otp/postgres/CHANGELOG.md
9c1f870a67786e6ea16922f5f8da42b67f9c753f08e45db72b8c44e11ba02993  otp/postgres/go.mod
```

Message Type: FINAL_ANSWER
Task name: /root
Sender: /root/otp_issue_custody_corrected_regression
Payload:
**APPROVE — scoped implemented regression audit.**

- Role/principal: fresh read-only `regression_reviewer`, `/root/otp_issue_custody_corrected_regression`.
- Repository: `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`.
- Branch: `feature/identity-platform`.
- Base/head: `3bb3e55e1a6c4764e84fe1f6dac81b245c6b75d5`, plus the exact nine dirty files below.
- No writes, tests, services, reruns or delegation.

No concrete regression finding in this custody tranche:
- Restart tests compare original digest, policy, remember parameters, original authorization record and timestamps, not merely successful decoding.
- Wrong, absent and cross-variant bearer tests exercise the private possession boundary for both original flow variants (`issuance_custody_test.go:176–226`).
- The original keyed scope commitment is encrypted at mapping and compared against the exact retained code-key version at restoration (`issuance_custody.go:109–114,209–220`). The previously accepted same-version changed-key case now has a meaningful denial assertion (`issuance_custody_test.go:237,251–252`).
- Separate issuance AAD, closed vocabulary and record-prefix validation prevent attempt-profile adoption (`issuance_custody.go:144–145,169–181`; `authority_custody.go:156–163`).
- Parser tests exercise both variants, every truncation, trailing bytes and size limits; successful fuzz decoding must re-encode identically.

**Claim boundaries:** supplied race/vet/fuzz results were not rerun or independently authenticated. No live Send, primary admission, issuance persistence, delivery composition or release claim is approved. Named bearer-key retirement/rotation lacks a new issuance-specific test in this tranche; exact named lookup and commitment comparison are present in both possession branches, with no observed defect. Complete-diff correctness and delivery audits remain separate gates.

Exact SHA-256 values, unchanged on final reinspection:

```text
16c682cbac400ca8028abacacf8a389987877f345f731869f012b528e5bf718d  otp/postgres/authority_custody.go
00511e25427db5aed5406556cbf342900f70f1e8843ca113e28766867f53fe93  otp/postgres/command.go
e8022cf57697cf79632638ec11a8ac762f3f9ea2989e051fac992a2d2aedefb7  otp/postgres/issuance_custody.go
7d63cbb0323224bd093d186e53cb5b37697c9748ed6457a64a846fce5bee9a70  otp/postgres/issuance_record.go
0b251c2d6acb8a00a70e595ce331a5a27529df40e8bcd736393a30893ed6e53f  otp/postgres/issuance_custody_test.go
04e67badb1661dcfda6a22b17d7637fad8289a08ff300d58d23b5d76d2c9d36e  otp/postgres/issuance_record_test.go
1a299c6ca8d1ef81d86ad2fc44de3953b3e49f39c5ac991d924f48a45b45998c  otp/postgres/README.md
3e99e84970d8c7e6e2e7ba19c274fe6bced51c9e20402f95d27f3d3b5234e0bf  otp/postgres/CHANGELOG.md
9c1f870a67786e6ea16922f5f8da42b67f9c753f08e45db72b8c44e11ba02993  otp/postgres/go.mod
```


## Native issuance custody first regression and compatibility, base 3bb3e55

Message Type: FINAL_ANSWER
Task name: /root
Sender: /root/otp_issue_custody_regression
Payload:
## Native regression review — APPROVE

- **Role/principal:** read-only `regression_reviewer`, `/root/otp_issue_custody_regression`.
- **Repository:** `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`.
- **Branch:** `feature/identity-platform`.
- **Base/HEAD:** `3bb3e55e1a6c4764e84fe1f6dac81b245c6b75d5`, plus the eight-file dirty snapshot below.
- **Disposition scope:** restart-safe code-free issuance custody and local original-bearer possession prerequisite only.
- No files changed, tests rerun, services accessed, or delegates used.

### Findings

No material regression defect found within the assigned boundary.

The implemented tests exercise observable custody outcomes:

- `issuance_custody_test.go:67–143` independently recreates the codec, restores both authorization variants, and compares original digest, policy, authority, continuation scope, remember policy, counters and times. It also checks absent submitted-code custody, disjoint bearer variants, tamper rejection and rejection by the existing attempt protocol.
- `issuance_custody_test.go:145–174` checks invalid scope/profile inputs, zero output on rejection, and cancellation.
- `issuance_custody_test.go:176–227` requires the original separately supplied bearer and rejects wrong, absent and cross-variant bearer custody for both flows.
- `issuance_custody_test.go:229–280` exercises retained wrapping/code-key identity, retirement, rotation without substitution, changed digest version and unknown reserved fields.
- `issuance_record_test.go:53–111` checks complete bounded record round-trips, every truncation, trailing/oversized data, all six purposes, policy and remember preservation. The fuzz oracle requires byte-exact canonical re-encoding.

The affected existing attempt boundary remains disjoint: `authority_custody.go:123–167` retains its original allowed record profiles unless explicitly decoding issuance, and `command.go:370–393` preserves the original attempt AAD domain/profile while giving issuance a separate domain/profile.

### Verification and pending boundaries

Parent-supplied execution evidence reports focused green regressions, full `otp/postgres` race testing, vet, and bounded record fuzzing passing. I reviewed the source/assertions but did not independently rerun those commands.

Real PostgreSQL tests were unchanged and not rerun. This approval does **not** establish live authority admission, native Send execution, transactional challenge/delivery composition, resend concurrency, unknown-commit recovery, or complete OTP/release readiness. Exact coverage, mutation and remaining release gates are not established by this review.

### Reviewed SHA-256 identities

```text
16c682cbac400ca8028abacacf8a389987877f345f731869f012b528e5bf718d  otp/postgres/authority_custody.go
00511e25427db5aed5406556cbf342900f70f1e8843ca113e28766867f53fe93  otp/postgres/command.go
5eb09414e42c2d8ab549cb20e8733e26b2499dca62f28cbbd7526177e8cf2861  otp/postgres/issuance_custody.go
7d63cbb0323224bd093d186e53cb5b37697c9748ed6457a64a846fce5bee9a70  otp/postgres/issuance_record.go
a5222dccf264e70ce8c9d6b83c9c27d6306722ce47c9c48bbfd5dd3cd3d02292  otp/postgres/issuance_custody_test.go
04e67badb1661dcfda6a22b17d7637fad8289a08ff300d58d23b5d76d2c9d36e  otp/postgres/issuance_record_test.go
1d1cdb1f8f4b0a57f85204f4dd5e801916ee999b95bb3a739ee9e8cc86afdd11  otp/postgres/README.md
de5b66d1b98bd4fdac325ad866f9e7916ebf2a53b114a25794df44556967b21c  otp/postgres/CHANGELOG.md
```

Message Type: FINAL_ANSWER
Task name: /root
Sender: /root/otp_issue_custody_compatibility
Payload:
**APPROVE — bounded source compatibility review.**

Principal: `/root/otp_issue_custody_compatibility`  
Repository: `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`  
Branch: `feature/identity-platform`  
Base/HEAD: `3bb3e55e1a6c4764e84fe1f6dac81b245c6b75d5`, plus the eight hashed dirty files below.

No compatibility defect found within native code-free custody and local bearer-possession scope:

- `command.go:372–394` preserves the existing attempt AAD domain, profile, validation and complete byte ordering.
- `authority_custody.go:123–164` preserves existing attempt record acceptance through the `issuance=false` path. Issuance accepts its separate discriminator; encrypted AAD domains are also disjoint.
- New public methods are additive. Existing codec/configuration/contributor shapes remain unchanged. Newly used imports introduce no new module dependency or ownership edge.
- Source uses APIs and language constructs compatible with declared Go `1.27.0`.
- Issuance preserves original digest, complete authorization record, policy, remember choice and times; submitted bearer custody remains separate. Local possession uses the retained original session-key version and commitment.

**Bounded gaps—not release approval:**

1. Clean-consumer compatibility is currently unavailable from declared pins. `otp/postgres/go.mod:8` pins root revision `84f8cd2317d7`; that tree lacks `feature_flow_bearer.go`. New calls at `issuance_custody.go:126–128` require constructors introduced by commit `93d9eaf592aebc6c069c488c3a6f5aff9b8ad5c1`. Owner-source composition does not prove published-consumer compilation.
2. Exact retained **code-key bytes** are a trusted configuration prerequisite, not authenticated by issuance restoration. `issuance_custody.go:93,210` discards `ScopeDigest` output and checks only availability of the named version. Same-version code-key byte substitution is not rejected here. Wrapping-key and bearer/session-key substitution are checked cryptographically. Do not claim that local restoration detects all named-code-key substitutions.
3. No builds, tests, services, release gates or runtime verification executed, as instructed. No claim about native Send registration, current primary authority or release readiness.

SHA256, rechecked at unchanged HEAD:

```text
16c682cbac400ca8028abacacf8a389987877f345f731869f012b528e5bf718d  otp/postgres/authority_custody.go
00511e25427db5aed5406556cbf342900f70f1e8843ca113e28766867f53fe93  otp/postgres/command.go
5eb09414e42c2d8ab549cb20e8733e26b2499dca62f28cbbd7526177e8cf2861  otp/postgres/issuance_custody.go
7d63cbb0323224bd093d186e53cb5b37697c9748ed6457a64a846fce5bee9a70  otp/postgres/issuance_record.go
a5222dccf264e70ce8c9d6b83c9c27d6306722ce47c9c48bbfd5dd3cd3d02292  otp/postgres/issuance_custody_test.go
04e67badb1661dcfda6a22b17d7637fad8289a08ff300d58d23b5d76d2c9d36e  otp/postgres/issuance_record_test.go
1d1cdb1f8f4b0a57f85204f4dd5e801916ee999b95bb3a739ee9e8cc86afdd11  otp/postgres/README.md
de5b66d1b98bd4fdac325ad866f9e7916ebf2a53b114a25794df44556967b21c  otp/postgres/CHANGELOG.md
```


## Native issuance storage exact-head delivery audit, source 3bb3e55

Message Type: FINAL_ANSWER
Task name: /root
Sender: /root/otp_issuance_storage_delivery
Payload:
## APPROVE — locally committed issuance-storage prerequisite only

Principal: `/root/otp_issuance_storage_delivery`; read-only audit, no writes, reruns, pushes or service mutations.

- **Exact head:** `3bb3e55e1a6c4764e84fe1f6dac81b245c6b75d5`, parent `93d9eaf592aebc6c069c488c3a6f5aff9b8ad5c1`.
- **Workspace:** clean `feature/identity-platform`; complete delta contains exactly the seven specified files; `git diff --check` passes.
- **Independent evidence:** inspected native mapping/QA principals and their original `eafc154` boundary, plus corrected regression, Go compatibility and final-review reports. All three final audits approve only storage prerequisites against `93d9eaf` plus frozen files.
- **Identity reconciliation:** all seven committed-blob SHA-256 hashes equal current working-tree bytes, all three final reports, and native execution fingerprint output `f4a77b`.
- **Execution evidence:** inspected native terminal output `5c6783`, exit 0, and underlying completed command: `go test . -tags integration -race -count=1` passed **25.625s**, followed by `go vet -tags integration ./...`, combined exit **0**. No reruns performed.
- **Cleanup:** `/tmp/identity-otp-issuance-storage.1PreTf` absent; named container absent and exact container ID returns “No such container.” No image removed by this audit.

**Delivery states:** local and committed verified; not pushed under supplied task authority, with no remote-tracking branch containing this head; hosted CI not established; deployment and production verification not established.

Public Send, native encrypted-delivery composition, full OTP/platform completion, and package/release gates remain unproven. This audit approves none of those broader claims.


## Native issuance storage blind final review, base 93d9eaf

Message Type: FINAL_ANSWER
Task name: /root
Sender: /root/otp_issuance_storage_final
Payload:
## APPROVE — native issuance-storage prerequisite only

No findings. Single-reviewer structured pass; no delegation or source changes.

**Reviewed boundary:** `feature/identity-platform`, base/HEAD `93d9eaf592aebc6c069c488c3a6f5aff9b8ad5c1`, plus exactly the seven frozen dirty files below.

### Source-backed closure and falsification

- **Atomic proposal:** `issuance.go:60–99` appends history, revokes replace-policy predecessors, then inserts the challenge through caller-owned handles. Failed insert, duplicate challenge, duplicate command and explicit rollback cases inspect both predecessor custody and send/challenge counts independently in `issuance_races_integration_test.go`. The helper does not own or commit another transaction.
- **Canonical custody:** `issuance.go:83–94` selects the fingerprint from the sole kernel claim, requires its pending command, and verifies the exact appended family version/time. Challenge identities remain protected by the existing schema uniqueness and lifecycle constraints.
- **Replacement versus unresolved ownership:** `issuance.go:51–55` rejects reserved predecessors before mutation. Challenge CAS and the existing lifecycle guard also prevent overwriting a concurrent reservation. The expired-reserved challenge falsification remains excluded from replacement; its owning generation and terminal state are preserved.
- **History horizon:** `issuance.go:117–180` separates indexed live-issued and reserved candidate lanes, bounding each before final sorting/locking. Eleven expired retained tombstones do not exhaust live capacity; unresolved reservations remain visible after expiry. Excess live candidates fail closed.
- **Migration/legacy compatibility:** `issuance_migration.go:46–72` preserves populated NULL-family custody, installs a tenant-qualified foreign key and immutable family trigger, and explicitly rejects later unbound inserts. The immutability test attempts an otherwise valid counted transition, so the existing lifecycle guard cannot falsely supply the oracle. Unchanged explicit-column readers retain their original representation.
- **Caller boundaries:** Original-flow/current-authority checks, shared-abuse exclusion, command replay and post-wait authority rechecks remain explicit caller prerequisites, not behavior this private storage helper claims to provide. Native production Send composition remains absent and documented.

### Verification boundary

Supplied final results, not independently rerun:
- `go test . -tags integration -race -count=1`: passed, **25.625s**.
- `go vet -tags integration ./...`: exit **0**.
- Supplied missing-source and horizon/expired-legacy red runs support the corresponding regressions.
- Read-only `git diff --check`: passed.

This approval does **not** establish public Send, encrypted delivery/audit/journal composition, complete login, package/release readiness, exact coverage/mutation, production migration lock duration, provider interoperability or runtime deployment. Those remain pending under the active goal.

### Frozen SHA-256 hashes

```text
df49f37d8c1ba6d31e78e5351cda4b59a501128a2f93febb9f501b502c30e3bc  otp/postgres/CHANGELOG.md
dba3cedaff2131b4fd82ff271257bcbc1f01b3e99c2aa548f233e7e541d0f15a  otp/postgres/README.md
e66f126576d9c9ae5c4c1d24c5f723e7fcf665636c76dfa7be6647444568d14a  otp/postgres/issuance.go
c0633226dd389806e7e82fd465604b455197a5762b0ba9796cd0a8a56ffa9328  otp/postgres/issuance_integration_test.go
d58d556333b95aaa3d508d9e8711eec635ea297dc1647b91d1dfce81e382037a  otp/postgres/issuance_migration.go
7a28a99affc8726140b856e20ddbda97f1bd5441deb507f4d6b6c5271244e866  otp/postgres/issuance_migration_test.go
f6753e4cb5ba26d5b2e0ce1bfc5a324d78b00d6dc489cb8ea2c5669f9ea9a97b  otp/postgres/issuance_races_integration_test.go
```


## Native issuance storage corrected regression and compatibility, base 93d9eaf

Message Type: FINAL_ANSWER
Task name: /root
Sender: /root/otp_issuance_corrected_compatibility
Payload:
**APPROVE — bounded Go compatibility review.**

Role: `go_compatibility` (`/root/otp_issuance_corrected_compatibility`).
Reviewed branch: `feature/identity-platform`.
Base/current HEAD: `93d9eaf592aebc6c069c488c3a6f5aff9b8ad5c1`, plus the seven-file dirty tranche.

No concrete compatibility findings.

- `IssuanceMigrationSource` is additive and conforms to the existing `migrations.Source` contract. Existing exported APIs, module requirements, Go `1.27.0` baseline, and dependency graph are unchanged.
- Installation is separately opted in, forward-only and transaction-default, requiring existing send-family storage. Existing rows retain NULL family bindings; no custody, counters, policy, keys or original times are rewritten.
- Existing challenge loading and consuming/recovery paths use explicit columns and do not require family bindings. The new update guard permits unchanged NULL bindings while prohibiting attachment, detachment or rewriting.
- Rejecting subsequent unbound inserts is an intentional, documented consequence of opting into this migration—not a default change to existing migration sources.
- Native issuance remains private storage work. This review does **not** certify public Send, coordinator admission, encrypted delivery composition, full module readiness or release completion.

Read-only source review; no builds or tests rerun and no parent test assurance used.

SHA256, paths relative to reviewed repository:

```text
df49f37d8c1ba6d31e78e5351cda4b59a501128a2f93febb9f501b502c30e3bc  otp/postgres/CHANGELOG.md
dba3cedaff2131b4fd82ff271257bcbc1f01b3e99c2aa548f233e7e541d0f15a  otp/postgres/README.md
e66f126576d9c9ae5c4c1d24c5f723e7fcf665636c76dfa7be6647444568d14a  otp/postgres/issuance.go
c0633226dd389806e7e82fd465604b455197a5762b0ba9796cd0a8a56ffa9328  otp/postgres/issuance_integration_test.go
d58d556333b95aaa3d508d9e8711eec635ea297dc1647b91d1dfce81e382037a  otp/postgres/issuance_migration.go
7a28a99affc8726140b856e20ddbda97f1bd5441deb507f4d6b6c5271244e866  otp/postgres/issuance_migration_test.go
f6753e4cb5ba26d5b2e0ce1bfc5a324d78b00d6dc489cb8ea2c5669f9ea9a97b  otp/postgres/issuance_races_integration_test.go
```

Message Type: FINAL_ANSWER
Task name: /root
Sender: /root/otp_issuance_corrected_regression
Payload:
## Native regression audit: APPROVE

- Principal: `/root/otp_issuance_corrected_regression`, independent read-only regression reviewer.
- Repository: `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`.
- Branch: `feature/identity-platform`.
- Base/current committed HEAD: `93d9eaf592aebc6c069c488c3a6f5aff9b8ad5c1`.
- Reviewed identity: that HEAD plus exactly the seven dirty paths and hashes below, unchanged across initial/final inspection.
- Scope: native issuance **storage prerequisite**, not public Send or release completion.
- No writes, services accessed, or test/build/analyzer commands run. I inspected authoritative goals, the permitted original mapping/QA section, complete new source/tests/docs, and affected existing challenge/send-window contracts. No peer report or ledger content was consulted.

### Source-backed closure

No unresolved concrete regression found.

- Expired unreserved `issued` predecessors no longer enter the family candidate set. The retained-horizon regression seeds **11 expired issued snapshots**, successfully issues the successor, and independently checks those tombstones remain `issued`, version 1, and expired.
- Reserved predecessors remain candidates regardless of expiry. Both live and expired reserved tests require replacement denial, preserve send count and challenge count, and check expired reservation custody remains unresolved.
- Live unbound legacy issuance blocks new issuance; expired unreserved legacy does not. Separate indexed legacy issued/reserved lanes implement that distinction.
- Migration adds nullable binding without rewriting existing custody, rejects subsequent unbound INSERTs, and prevents binding rewrite during an otherwise legal counted attempt transition. The existing reader is exercised after migration.
- The stored issue fingerprint is selected directly from the original kernel claim, not supplied by the binding. Integration asserts its exact expected bytes.
- Replacement and send-history changes remain in the caller transaction. The actual INSERT-failure test injects a database constraint failure **after** provisional family append/revocation; final assertions require the predecessor and allowance to remain unchanged with no successor.
- The two-connection resend test requires exactly one success, accepts only the intended denial for the loser, and checks final predecessor state, committed challenge count and send-history count.
- Collision, duplicate issuance-command identity, expired draft, explicit rollback, replace and retain outcomes have persisted-state assertions rather than callback choreography.

### Bounded evidence gaps

These do not invalidate this storage tranche:

- Migration preservation is checked through the existing reader and selected original state/version—not a complete before/after populated snapshot comparison.
- Expired **legacy reserved** exclusion lacks its own integration partition, although the SQL explicitly includes legacy reservations independently of expiry.
- The new private writer has no production caller yet. Tests do not prove original-flow/shared-abuse lock ownership, post-wait authority rechecks, consuming-attempt/resend races, encrypted delivery/audit/journal composition, unknown-commit recovery, or public enumeration safety.
- Supplied passing execution summaries were not independently rerun or certified by this source audit. Exact coverage/mutation and other release gates remain outside this approval.

### Reviewed SHA-256

```text
df49f37d8c1ba6d31e78e5351cda4b59a501128a2f93febb9f501b502c30e3bc  otp/postgres/CHANGELOG.md
dba3cedaff2131b4fd82ff271257bcbc1f01b3e99c2aa548f233e7e541d0f15a  otp/postgres/README.md
e66f126576d9c9ae5c4c1d24c5f723e7fcf665636c76dfa7be6647444568d14a  otp/postgres/issuance.go
c0633226dd389806e7e82fd465604b455197a5762b0ba9796cd0a8a56ffa9328  otp/postgres/issuance_integration_test.go
d58d556333b95aaa3d508d9e8711eec635ea297dc1647b91d1dfce81e382037a  otp/postgres/issuance_migration.go
7a28a99affc8726140b856e20ddbda97f1bd5441deb507f4d6b6c5271244e866  otp/postgres/issuance_migration_test.go
f6753e4cb5ba26d5b2e0ce1bfc5a324d78b00d6dc489cb8ea2c5669f9ea9a97b  otp/postgres/issuance_races_integration_test.go
```


## Native issuance storage first regression and compatibility, base 93d9eaf

Message Type: FINAL_ANSWER
Task name: /root
Sender: /root/otp_issuance_storage_regression_93d9eaf
Payload:
## Native regression audit: REQUEST_CHANGES

- Principal: `/root/otp_issuance_storage_regression_93d9eaf`; fresh read-only context.
- Skills: `$testing-coverage` and `$verification-before-completion`, read completely.
- Repository: `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`.
- Branch: `feature/identity-platform`.
- Base/HEAD: `93d9eaf592aebc6c069c488c3a6f5aff9b8ad5c1`, plus the seven listed dirty files.
- No writes, Go commands, services, Git mutations, or delegates.

### Concrete defect: expired retained rows eventually block issuance permanently

[issuance.go:137](/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform/otp/postgres/issuance.go:137) selects every `issued` predecessor, regardless of expiry, and [line 150](/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform/otp/postgres/issuance.go:150) rejects the eleventh selected row.

Retain-policy issuance preserves predecessor states. Over successive rolling windows, expired challenges therefore accumulate as `issued`, even though their send-history entries correctly age out. Eleven retained issuances can accumulate; the twelfth then returns `ErrUnavailable` until someone terminalizes previous rows. The writer neither performs that terminalization nor documents an enforced cleanup prerequisite. A rolling-hour quota has inadvertently become a lifetime challenge-count limit.

The current `retain` test creates only two challenges. The existing hour-pruning test exercises send history without issuance predecessors, so neither catches this interaction.

**Smallest regression:** populate eleven valid family-bound, expired retain-policy challenges with legally aged send history, then attempt a current issuance. Assert success, exactly one history increment/new challenge, and unchanged old custody. Separately retain an expired `reserved` predecessor and assert replacement still denies: expiry must not discard unresolved reservation authority.

The correction should keep predecessor work bounded without counting expired nonreserved custody as live issuance authority. General cleanup implementation is not required to demonstrate this bounded storage invariant.

### Existing assertions support

- Committed history, challenge insertion and predecessor transitions.
- Replace versus retain, explicit rollback, and failure at the actual challenge insertion.
- Reserved replacement denial.
- Duplicate command/challenge rejection without partial mutation.
- Expired new-draft denial.
- Concurrent resend winner count and final persisted counts/state.
- Populated legacy fail-closed behavior without inferred family.
- Exact stored code/scope digests and original timestamps.
- Family immutability during an otherwise valid attempt update.
- Migration input validation and cancellation.

### Remaining evidence boundaries

The supplied RED, storage PASS, focused race matrix, full adapter integration/race and vet results were **not independently executed or certified** in this audit.

Existing consumption/recovery after issuance-migration installation, old unbound-writer rejection, and exact full issuance-fingerprint equality are not directly asserted by the added tests. These are additional confidence limitations, not separate findings requiring another QA plan.

Public Send, original-flow/current-risk admission, encrypted delivery/audit/journal composition, unknown issuance recovery, cleanup and complete release gates remain pending outside this tranche.

### SHA-256 snapshot

```text
15403291ef19364a2e89949f558cfc0366270cfc4f764128b7d9c3e04586bfaa  otp/postgres/CHANGELOG.md
0097675253e391af64c576e9fb78ed5b8dca3ca1b9d1aedb6666b53b55e0ac46  otp/postgres/README.md
5686509996e3130e247bb9cbcf778440e4412e143d0844ebfc4fcd6f4f2b9492  otp/postgres/issuance.go
dfc507a71d8c8bb48b05bd3359324eec2cdeedfc1271e53991dae2b1c67543b4  otp/postgres/issuance_integration_test.go
f3701142403104109527b60622031b950424946809b67bf1a91511973d7fbb72  otp/postgres/issuance_migration.go
7a28a99affc8726140b856e20ddbda97f1bd5441deb507f4d6b6c5271244e866  otp/postgres/issuance_migration_test.go
2de484040f2bbe79ff5a70167ce66daa229e8cb9fc0a80bb732e532e2f95feb3  otp/postgres/issuance_races_integration_test.go
```

Message Type: FINAL_ANSWER
Task name: /root
Sender: /root/otp_issuance_storage_compat_93d9eaf
Payload:
**APPROVE — bounded source/API compatibility audit.**

Reviewed `93d9eaf592aebc6c069c488c3a6f5aff9b8ad5c1` plus exactly seven dirty files, including untracked files; checked relevant send-window, challenge, consuming/checking and recovery contracts and same-repository ancestry.

- Only public addition is opt-in `IssuanceMigrationSource`; existing Go APIs, Go 1.27 baseline, module paths and dependency pins are unchanged.
- Migration preserves existing row custody and nullable legacy family bindings. Existing explicit-column consumers and lifecycle/recovery updates remain compatible.
- Older unbound INSERTs deliberately fail after opt-in installation; README/changelog disclose this operational compatibility boundary.
- Private writer does not introduce another transaction, journal or replay owner. Command fingerprint derives from the kernel claim; replacement cannot revoke reserved predecessors.

**Checkpoint limitation:** `issuance.go:138–154` counts all `issued`/`reserved` predecessors, including expired-but-not-terminalized rows, against its ten-row bound. Retain-policy abandoned rows can eventually deny further issuance until terminalization. This is not an existing-API regression; cleanup/full Send remain explicitly incomplete. Do not extend approval to complete rolling-hour issuance availability.

Verification: source inspection and `git diff --check` passed. No Go commands, services or mutations performed. Supplied integration/race/vet outcomes were not independently verified. No clean-consumer, interoperability, release or full-platform completion claim.

SHA-256, paths relative to worktree:

```text
15403291ef19364a2e89949f558cfc0366270cfc4f764128b7d9c3e04586bfaa  otp/postgres/CHANGELOG.md
0097675253e391af64c576e9fb78ed5b8dca3ca1b9d1aedb6666b53b55e0ac46  otp/postgres/README.md
5686509996e3130e247bb9cbcf778440e4412e143d0844ebfc4fcd6f4f2b9492  otp/postgres/issuance.go
dfc507a71d8c8bb48b05bd3359324eec2cdeedfc1271e53991dae2b1c67543b4  otp/postgres/issuance_integration_test.go
f3701142403104109527b60622031b950424946809b67bf1a91511973d7fbb72  otp/postgres/issuance_migration.go
7a28a99affc8726140b856e20ddbda97f1bd5441deb507f4d6b6c5271244e866  otp/postgres/issuance_migration_test.go
2de484040f2bbe79ff5a70167ce66daa229e8cb9fc0a80bb732e532e2f95feb3  otp/postgres/issuance_races_integration_test.go
```

## Code-free original-bearer exact-head delivery audit, commit 93d9eaf

Message Type: FINAL_ANSWER
Task name: /root
Sender: /root/otp_issue_bearer_delivery_93d9eaf
Payload:
**APPROVE — local committed bearer-custody prerequisite only.**

- Exact head `93d9eaf592aebc6c069c488c3a6f5aff9b8ad5c1`, parent `ff9fd86bd7e3415c148b0da2aa1e9f9bad5a984b`, branch `feature/identity-platform`.
- All ten committed blob SHA-256 values match native final, regression and compatibility snapshots; committed delta is ten files, 337 additions, four deletions.
- Source checkout clean; committed diff whitespace check passed; `/tmp/identity-otp-send-native.ncp6zM` absent.
- Inspected native mapper, bounded QA, regression, compatibility and blind final reports with their distinct principals and reviewed base/snapshot. No unresolved reported findings or snapshot discrepancies.
- Ledger correctly retains **5/67 integrated, 1/67 verified**; no unit promotion.

No Go execution or independent test-run certification performed. OTP Send, primary authentication, SQL integration, clean-consumer composition, exact coverage/mutation, release readiness and platform completion remain unverified. Local commit only; no push, release, deployment or production certification.

## Code-free original-bearer complete blind final review, base ff9fd86

Message Type: FINAL_ANSWER
Task name: /root
Sender: /root/otp_issue_bearer_final_ff9fd86
Payload:
## APPROVE — narrow development prerequisite

No findings.

Reviewed `feature/identity-platform` at base/HEAD `ff9fd86bd7e3415c148b0da2aa1e9f9bad5a984b` plus all ten dirty files below. Single-reviewer structured, read-only pass; no Go execution, services, file writes, Git mutations, or peer delegation.

### Source-backed closure

- `feature_flow_bearer.go:36–56` bounds transport before decoding, requires strict canonical unpadded base64url representing 32–64 bytes, clones retained bytes, freezes structural scope, and populates exactly one bearer slot without code/proof.
- `feature_command.go:119–123`, both changed borrow accessors, and `feature_code.go:38–46` preserve disjoint custody, deny mixed/kernel commands, and leave `OneTimeCode()` absent. The consumed immutable authentication source confirms `BearerCredential` is a private string with value-only methods.
- `Feature()` projections contain only cloned metadata; reconstruction cannot restore custody. Existing `Command.Format()` and `MarshalJSON()` keep diagnostics redacted and serialization forbidden.
- `postgres/codec.go:119–152` binds the original bearer under the unchanged `pre-auth-cookie` or `session-token` named-secret domain. Only private keyed digests enter canonical intent; safe fields exclude those digests and bearer bytes. `postgres/claims.go:97–113` recomputes against the stored retained key from original command custody.
- `postgres/otp_attempt.go:213` still requires actual code custody before OTP preparation; reservation delegates through that guard. Native `otp/postgres/command.go:361–364` independently rejects restoration without submitted code.
- Existing code-bearing constructors are unchanged. Their borrowers retain their former behavior for every publicly constructible old command.

Bounded falsification challenges covered trailing-bit/newline canonicalization, identical bytes under opposite custody variants, mixed custody, projection reconstruction, changed-bearer intent, retained-key recomputation, and code-free entry into OTP comparison/reservation. Source guards preserve the stated invariants.

Tests use deterministic fixtures and independent literal HMAC framing. Supplied red/race/vet/fuzz outcomes were **not independently executed or certified**. `git diff --check` passed during this review.

This approval does **not** establish OTP Send, primary authentication, SQL integration, published dependency compatibility, clean-consumer composition, exact coverage/mutation gates, release readiness, or platform completion. Those boundaries remain pending; the authoritative goal explicitly permits this bounded development prerequisite.

### Frozen SHA-256 snapshot

```text
ceef4af8ddd03d5c1eb0ceb7fd9ad65d92f1fc2718e047444a578b1fce13664c  CHANGELOG.md
8f42dbd77fa5bdaebf87c763594814937c4ee70450de275404fa58cb3ef2b228  README.md
9163416d54e8c3db6625b2f7df207d5c09fe7f294422c2fed5dac24b0df462ce  feature_command_internal_test.go
799493a730613d6b7794305724f7de08d1ce54e763708df67e74bedc70a3c301  feature_preauth_cookie.go
6bed18323b81087ef5087c43da645ebf6e70faac6328002e21aecc51e0d3a891  feature_session_token.go
8f95b8522e00cfbd5af03726c2b8652cfa8370a585467d8fe2f39665da53e381  feature_flow_bearer.go
6cd0af64703342de2a339d5bc7ceb930f1bc5d28ad89cddcc9832138b58ee89d  feature_flow_bearer_test.go
d90c883371b64310c0001b76a71e65f272ab325e3844e270ea2858140670f7f7  postgres/CHANGELOG.md
28998cade3af2daf6344cf6fa799b7724e40987f5754bb5bfeada31448d94412  postgres/README.md
1038704265e0e12f7a2562b3aaa43d456835d7edc9ce68dd9d768193b164b3c0  postgres/codec_flow_bearer_test.go
```

## Code-free original-bearer native regression and compatibility, base ff9fd86

Message Type: FINAL_ANSWER
Task name: /root
Sender: /root/otp_issue_bearer_regression_ff9fd86
Payload:
## Native regression review: APPROVE

**Principal:** `/root/otp_issue_bearer_regression_ff9fd86`, fresh read-only regression reviewer using `$testing-coverage`.

**Reviewed snapshot:** `feature/identity-platform`, HEAD/base `ff9fd86bd7e3415c148b0da2aa1e9f9bad5a984b` plus the ten dirty files below.

**Findings:** No concrete observable regression or blocking coverage defect identified in this bounded code-free bearer-custody slice.

- Public tests distinguish cookie-only and session-only custody, retain exact submitted bytes, and reject fabricated OTP custody or opposite bearer variants.
- Bounds, canonical encoding, redacted typed failures, immutable scalar ownership, concurrent reads, formatting, serialization rejection and projection reconstruction are meaningfully asserted.
- Internal tests reject mixed bearer and kernel/feature custody.
- Kernel tests independently derive the existing named HMAC domains, distinguish changed bearer and context, preserve safe scalar projections, exclude raw bearer/code entries, and recompute original intent using retained original custody.
- Existing code-bearing constructors are unchanged. Relaxed bearer accessors retain the closed-union check. OTP comparison and reservation preparation still require `OneTimeCode()`; new tests exercise both denial routes.

**Execution evidence limits:** No Go commands, services, writes or delegation performed. Parent supplied expected compile-red results and passing focused/full root/kernel race, vet, direct OTP checks and bounded fuzz results; these were not independently rerun or inspected as native execution logs. Approval is source/test-quality review, not independent certification of those executions.

**Unverified boundaries:** Native OTP sending, actual bearer possession/current primary authority, SQL integration, durable replay/unknown outcomes, release coverage/mutation gates and unit/platform completion. None is claimed by this custody addition.

### Snapshot SHA-256

```text
ceef4af8ddd03d5c1eb0ceb7fd9ad65d92f1fc2718e047444a578b1fce13664c  CHANGELOG.md
8f42dbd77fa5bdaebf87c763594814937c4ee70450de275404fa58cb3ef2b228  README.md
9163416d54e8c3db6625b2f7df207d5c09fe7f294422c2fed5dac24b0df462ce  feature_command_internal_test.go
799493a730613d6b7794305724f7de08d1ce54e763708df67e74bedc70a3c301  feature_preauth_cookie.go
6bed18323b81087ef5087c43da645ebf6e70faac6328002e21aecc51e0d3a891  feature_session_token.go
8f95b8522e00cfbd5af03726c2b8652cfa8370a585467d8fe2f39665da53e381  feature_flow_bearer.go
6cd0af64703342de2a339d5bc7ceb930f1bc5d28ad89cddcc9832138b58ee89d  feature_flow_bearer_test.go
d90c883371b64310c0001b76a71e65f272ab325e3844e270ea2858140670f7f7  postgres/CHANGELOG.md
28998cade3af2daf6344cf6fa799b7724e40987f5754bb5bfeada31448d94412  postgres/README.md
1038704265e0e12f7a2562b3aaa43d456835d7edc9ce68dd9d768193b164b3c0  postgres/codec_flow_bearer_test.go
```



Message Type: FINAL_ANSWER
Task name: /root
Sender: /root/otp_issue_bearer_compat_ff9fd86
Payload:
## Go compatibility disposition: APPROVE

**Scope:** development compatibility of the complete ten-file dirty change over `ff9fd86bd7e3415c148b0da2aa1e9f9bad5a984b`, branch `feature/identity-platform`. No concrete compatibility regression found.

- **Additive API:** `feature_flow_bearer.go:19,32` adds constructors without removing or changing existing signatures or exported types. Bounded canonical transport matches the existing code-bearing constructors.
- **Accessor compatibility:** `feature_preauth_cookie.go:49` and `feature_session_token.go:49` admit the new code-free custody while retaining existing code-bearing results. Zero, ordinary, projected, opposite-variant and mixed commands remain denied through `Feature()` (`feature_command.go:121`).
- **Canonical intent:** existing constructors and PostgreSQL codec are unchanged. Old code-bearing commands retain their original fields and keyed framing (`postgres/codec.go:119–152`). New code-free commands omit `one_time_code` while binding bearer intent into the existing distinct domains.
- **OTP consumers:** comparison admission independently requires `OneTimeCode()` (`postgres/otp_attempt.go:213`); reservation delegates through that check (`postgres/otp_reservation.go:148`). Native restoration independently rejects absent code (`otp/postgres/command.go:361`). Borrowing a bearer therefore does not admit an OTP comparison or consumption.
- **Module boundaries:** root, kernel, OTP and OTP/PostgreSQL manifests retain `go 1.27.0`. No module/package paths, dependency pins, replacements or manifests change in this batch. New APIs remain in the already registered root package.

**Evidence limits:** I performed read-only source inspection, no Go execution. Parent-supplied race/vet/fuzz results support scoped development verification, not clean published consumption, release gates, PostgreSQL integration, current-primary admission or native OTP Send. Nested modules still pin historical root revisions; using these new constructors from those modules requires an updated source pin or explicitly evidenced development overlay before any composition claim. This review does not certify publication or release completeness.

### Reviewed file SHA-256

```text
ceef4af8ddd03d5c1eb0ceb7fd9ad65d92f1fc2718e047444a578b1fce13664c  CHANGELOG.md
8f42dbd77fa5bdaebf87c763594814937c4ee70450de275404fa58cb3ef2b228  README.md
9163416d54e8c3db6625b2f7df207d5c09fe7f294422c2fed5dac24b0df462ce  feature_command_internal_test.go
799493a730613d6b7794305724f7de08d1ce54e763708df67e74bedc70a3c301  feature_preauth_cookie.go
6bed18323b81087ef5087c43da645ebf6e70faac6328002e21aecc51e0d3a891  feature_session_token.go
8f95b8522e00cfbd5af03726c2b8652cfa8370a585467d8fe2f39665da53e381  feature_flow_bearer.go
6cd0af64703342de2a339d5bc7ceb930f1bc5d28ad89cddcc9832138b58ee89d  feature_flow_bearer_test.go
d90c883371b64310c0001b76a71e65f272ab325e3844e270ea2858140670f7f7  postgres/CHANGELOG.md
28998cade3af2daf6344cf6fa799b7724e40987f5754bb5bfeada31448d94412  postgres/README.md
1038704265e0e12f7a2562b3aaa43d456835d7edc9ce68dd9d768193b164b3c0  postgres/codec_flow_bearer_test.go
```

## Code-free original-bearer native mapping and bounded QA, base ff9fd86

Message Type: FINAL_ANSWER
Task name: /root
Sender: /root/otp_issue_bearer_mapping_ff9fd86
Payload:
## Native change-mapper report

**Identity:** fresh read-only mapper  
**Source:** `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`  
**Reviewed head/base:** `ff9fd86bd7e3415c148b0da2aa1e9f9bad5a984b`, clean  
**Execution:** source inspection only; no writes, Go commands, services, delegation, or peer reports.

### Concrete missing public seam

Root cannot construct a feature command carrying **only the actual submitted pre-auth cookie or session token**:

- `NewFeatureOneTimeCodePreAuthCommand` and `NewFeatureOneTimeCodeSessionCommand` require a valid submitted OTP.
- `PreAuthCookie()` and `SessionToken()` additionally require `feature.code != nil`.
- Ordinary `NewFeatureCommand` carries scalar metadata only; its contract expressly forbids raw bearers.
- Generic coordinator preparation rejects every command exposing `OneTimeCode()` (`postgres/coordinator.go:182`).

Therefore OTP issuance cannot use existing custody APIs without fabricating an OTP or violating scalar-secret policy. This is a confirmed prerequisite gap, not evidence that the native issuance workflow is complete.

### Contract and caller map

| Boundary | Existing contract and affected consumers |
|---|---|
| Root custody | `feature_command.go` stores separate code, pre-auth-cookie, and session-token pointers. `Feature()` excludes kernel/feature mixing and simultaneous bearer contexts; projections copy scalar inputs and expose no secrets. Constructors freeze canonical unpadded base64url, 32–64-byte bearers, validating syntax only. |
| Diagnostics | `command.go` redacts all formatting and denies JSON serialization. Borrowed bearer copies remain secret-bearing; callers must discard them after bounded use. |
| Canonical command intent | `postgres/codec.go:119–155` independently binds code, cookie, and session token into keyed intent. Cookie and session use distinct domains and payload members; neither collides with `feature.<name>`. Complete binding scope includes tenant, user or anonymous flow, actor/kind, operation, policy/schema versions, correlation, optional record/link mode, and feature descriptor. |
| Scalar mapping | `postgres/codec_feature.go` binds every scalar’s name, explicit kind, and value through conservative keyed digests. Structural ordering is canonical; scalar projections cannot restore secret custody. |
| Replay/key rotation | `postgres/claims.go` persists fingerprints and safe scope, not raw canonical commands. Replay recomputes original command intent under the stored retained fingerprint key using `claimRequest.source`. Changing only bearer bytes must remain changed intent. |
| Generic issuance-capable execution | `FeatureExecutionContributor.ValidateFeatureCommand` and `ApplyFeature` receive the opaque original `Command`. `ApplyFeature` compares its complete canonical intent with reserved intent before invocation. `Work.FeatureScope` restricts native invocation to the exact registered comparable instance and active callback. Recovery scope does not reconstruct original secrets. |
| Final-authentication variant | `postgres/authenticated_command.go` is explicitly restricted to `identity.session.issue`, validates the configured opaque authenticated proof, and independently binds complete original proof metadata. It is not a substitute for OTP issuance cookie/session possession. |
| OTP attempt mapping | `AttemptCodec.MapPreAuth[WithAuthority]` and `MapSession[WithAuthority]` freeze complete original continuation plus submitted code; bearer remains outside scalar and encrypted attempt custody. Existing public signatures and semantics are compatible consumers to preserve. |
| Native possession validation | Pre-auth derives `PreAuthDigest` under the original cookie key version; session derives its lookup digest under the original lookup key version. Both compare commitments using constant-time comparison and reject missing retained keys or incompatible contexts. Mapping itself grants no live admission. |
| Current native authority | Existing attempt paths recheck directory, user, credential, authorization and original session/pre-auth primary state. Session policy/freshness caps preserve original authentication time; optional organization/factor/passkey authority currently fails closed. |
| Borrowed OTP handles | Counted-attempt and reservation handles lend original opaque `Command` only to the exact active owner. Apply/finalize handles instead lend code-free `CommandSpec` and separate cookie/token custody with exact-instance, generation and callback-lifetime checks. Release/recovery deliberately have no bearer access. |

### Original-context variants and security expectations

`otp.FlowAuthorization` already provides an explicit disjoint tagged union:

- Legacy ID-only authorization is structural and cannot substitute for complete original context.
- Bound pre-auth authorization retains the full original transaction and cookie commitment.
- Bound authenticated-session authorization retains the complete verified session, lookup commitment, authority vector and original authentication timeline.
- `NewContinuation` binds the discriminator and full original context; signin requires pre-auth and preserves original remember policy. Authenticated-session signin is forbidden.

Any additive root seam must preserve these observable requirements:

1. Exactly one bearer context, never inferred from user presence, remember policy, nullable fields, or equal bearer bytes.
2. No fabricated submitted code for issuance.
3. No bearer in `FeatureCommandSpec`, encrypted persisted envelopes, receipts, recovery projections, errors, or diagnostics.
4. Syntax-valid possession custody is **not** proof: native owners derive and compare actual possession under original retained keys and check current primary authority.
5. Wrong bearer, changed context, changed original metadata, and unavailable retained key fail closed; retries cannot adopt fresh records or renew authentication.
6. Existing code-bearing constructors, fingerprint framing, supported attempt consumers, and code-free recovery remain compatible.

### Remaining concrete boundaries

- The inspected production tree has no native `identity.otp.send` feature executor; existing send-family policy/storage helpers are not complete issue/send orchestration.
- The new root custody capability alone will not prove send authorization, current-state admission, digest-only challenge issuance, resend atomicity, delivery, or unknown-commit recovery.
- No additional blocker was identified for an additive root-only custody seam.

Message Type: FINAL_ANSWER
Task name: /root
Sender: /root/otp_issue_bearer_qa_ff9fd86
Payload:
## Native QA plan

**Role:** `qa_test_planner`, fresh read-only bounded pass using `$qa-test-planner`.  
**Reviewed source:** `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`, clean branch `feature/identity-platform`, base/head `ff9fd86bd7e3415c148b0da2aa1e9f9bad5a984b`.  
**Disposition:** No contradictory requirement, unsafe requested action, or design defect blocks first write. No writes, Go commands, services, delegation, or peer reports used.

### Established contract and present gap

`feature_preauth_cookie.go` and `feature_session_token.go` expose original bearer custody only through constructors requiring a nonzero submitted OTP code. Their accessors also explicitly require `feature.code != nil`. This cannot represent native OTP sending before the user submits any code.

Existing invariants to preserve:

- Exactly one bearer variant: pre-auth cookie or session token.
- Canonical unpadded base64url encoding of 32–64 decoded bytes.
- Custody is syntax-validated input, not possession proof.
- `Feature()` exposes copied scalar metadata without code/bearer custody.
- Command formatting is redacted; JSON serialization rejects.
- `postgres/codec.go` binds bearer intent with separate keyed `pre_auth_cookie` / `session_token` domains, distinct from `one_time_code`.
- Current attempt mappers remain code-bearing; original flow/session owners independently compare bearer commitments and authoritative original scopes.
- Recovery and durable projections must not carry raw bearers or fabricated codes.

### Minimal observable regression coverage

| Boundary / risk | Required observable assertion |
|---|---|
| Public no-code pre-auth custody | Valid spec plus canonical cookie succeeds; exact cookie is borrowable; `OneTimeCode()` and `SessionToken()` both return false. |
| Public no-code session custody | Valid spec plus canonical token succeeds; exact token is borrowable; `OneTimeCode()` and `PreAuthCookie()` both return false. |
| Closed union / accidental disclosure | Zero, ordinary feature, kernel, mixed kernel-feature and mixed bearer commands disclose no secret. Corrupted mixed commands also deny `Feature()` and every secret accessor. |
| Canonical bounds | Accept 32 and 64 decoded bytes; reject 31/65, empty, padding, leading/trailing whitespace, newline, invalid alphabet and noncanonical trailing bits. Failure returns zero command and redacted typed invalid-input error. One oversized input exercises the pre-decode bound. |
| Projection and immutable ownership | Mutating constructor inputs or returned binding/field/byte slices cannot alter command intent; reconstructing `NewFeatureCommand(projected)` yields no bearer/code custody. Concurrent reads retain exact bearer and metadata. |
| Diagnostics | All existing formatting verbs exclude the synthetic bearer; JSON rejects both no-code variants. Invalid-input errors do not include bearer content. |
| Kernel canonical intent | Same metadata with a different bearer changes canonical intent; identical bearer bytes in different variants produce different intent; plain projection differs from secret-bearing command. Literal independent HMAC oracle verifies each existing named secret domain. Raw bearer is absent from canonical and safe fields. No `one_time_code` entry appears for the new no-code variants. |
| Replay / key rotation | Recomputing retained original custody under the original fingerprint key reproduces canonical bytes; rotating the key changes private canonical commitment without altering safe scalar projection. |
| Existing consumer compatibility | Old code-only, code+pre-auth and code+session constructors still reject zero code, retain the exact code, retain only their selected bearer, and produce unchanged canonical payload bytes. |
| Authority boundary | No-code custody must not make code-consuming attempt paths admissible, fabricate an attempt/proof/reservation, or silently enable an unimplemented generic execution path. Existing applicable deny cases remain failures. |

### Smallest sufficient checks

1. **Root public unit tests:** new no-code custody cases, bounds table, immutability/redaction, compatibility; retain existing `Feature*` regressions.
2. **Root internal unit tests:** mixed-variant fail-closed cases, including no-code mixed bearers.
3. **Kernel codec unit tests:** no-code canonical intent, independent HMAC framing, no raw secret/no code entry, retained-key replay, and old payload byte compatibility.
4. **Affected narrow race run:** concurrent immutable custody/projection reads.
5. **Bounded fuzz:** canonical bearer acceptance/rejection for both new variants; reuse existing corpus classes, including 32/64-byte limits and malformed trailing bits.
6. **Existing affected OTP admission/attempt tests:** ensure relaxing custody accessors does not bypass code-required consuming behavior.

All Go executions must use automatically cleaned task-owned `GOCACHE`; none were executed in this planning pass.

### Explicit claim limits

Root/codec tests can prove transport custody and canonical binding, **not** current possession, primary-state validity, original-scope admission, send-counter enforcement, delivery, SQL persistence, or unknown-outcome recovery. The future native send owner must separately prove correct versus wrong bearer, variant mismatch, original scope/version/key mismatch, expired/revoked authority, and unchanged original scope across retry/recovery. Those are consumer implementation assertions, not blockers for safely adding the missing root custody contract.

## Send-family storage exact-head delivery, ff9fd86

Message Type: FINAL_ANSWER
Task name: /root
Sender: /root/otp_send_window_delivery_ff9fd86
Payload:
**APPROVE — bounded local committed storage checkpoint only.**

- Verified branch `feature/identity-platform`, clean worktree, HEAD `ff9fd86bd7e3415c148b0da2aa1e9f9bad5a984b`, parent `eafc154d7a466dc864613c3128d3dad1f697bf70`.
- Inspected the complete seven-path committed delta. All seven current file SHA-256 hashes exactly match native final-correctness, regression and compatibility reports; no post-review code change detected. `git diff HEAD^ HEAD --check` passed.
- Inspected native reports from `otp_issue_mapping_eafc154`, `otp_issue_qa_eafc154`, `otp_send_window_regression_eafc154`, `otp_send_window_compat_eafc154` and `otp_send_window_final_eafc154`. Mapping/QA identify broader issuance work; independent final audits approve only this storage tranche.
- Parent-supplied final focused PostgreSQL integration/race result: **PASS, 2.307s**; vet: **exit 0**. Not independently rerun. Earlier full-adapter **20.600s** result preceded final tightening and does **not** establish a current full-adapter pass.
- Run directory, commit-message directory and named Docker container are absent.

**Delivery state:** locally committed; push, CI, deployment and production verification are not claimed or audited. Public Send, complete OTP integration, exact coverage/mutation and release gates remain pending. This approval does not certify unit/release completeness.

Applied `verification-before-completion` to constrain the delivery claim to inspected current artifacts.

## Send-family storage blind final correctness, base eafc154

Message Type: FINAL_ANSWER
Task name: /root
Sender: /root/otp_send_window_final_eafc154
Payload:
## Native review: APPROVE

No findings.

**Boundary:** `feature/identity-platform`, repository `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`; complete seven-path dirty slice against `eafc154d7a466dc864613c3128d3dad1f697bf70`. HEAD and scoped file hashes remained unchanged during review. Single-reviewer structured pass; no writes, execution of Go code, service access, external mutation, or subdelegation.

### Source-backed disposition

- **Identity and isolation:** `send_window.go:19` validates the complete bounded scope before deriving a separately domain-separated HMAC. Tenant, purpose, subject, channel and destination distinguish families; challenge, target, authorization, generation and policy changes do not reset them. The database selector and primary key preserve tenant/family identity.
- **Window and numerical boundaries:** `send_window.go:42` rejects invalid/future/unordered history, preserves input ownership, excludes sends exactly one hour old, admits exact resend equality, and applies the current validated cap without discarding live history. Full-width versions are parsed canonically and exhaustion rejects further increments.
- **Persistence and concurrency:** initialization uses enforced uniqueness; loading obtains `FOR UPDATE`; persistence combines the locked proposal with an exact version CAS and database-time fences. Helpers neither open nor commit transactions. Coordinator replay, original-flow admission, shared-abuse locking and atomic challenge/audit/delivery composition remain explicit caller obligations—not falsely implemented public behavior.
- **Schema and migration:** `send_migration.go:47` adds new objects after the existing lockout namespace without rewriting earlier migration SQL or challenge state. Guards reject invalid histories, identity changes, live-prefix omission, future/backward append time and ordinary row deletion. Maximum retained history is ten instants. No destructive down migration or implicit adoption of existing objects is introduced.
- **Falsification:** attempted same-family reissue, stale-version update, deletion/reset, omission of an existing live send, and exact cutoff/cooldown counterexamples. Source guards and independent test expectations reject the unsafe outcomes while retaining valid pruning and appends.
- **Tests and documentation:** deterministic unit boundaries and owned real-PostgreSQL fixtures exercise helper behavior, rollback, retained family identity, key mismatch, guard failures and competing connections. Documentation accurately limits the checkpoint to private storage.

### Evidence limits

I inspected source, assertions, fixtures, contracts and `git diff --check` directly; the latter passed. Supplied execution evidence reports final focused PostgreSQL 18.3 integration/race tests passed in **2.307s** and `go vet ./...` exited **0**. I did not independently execute those checks.

Approval covers this incremental development batch only. Public Send, challenge replacement, encrypted-delivery/coordinator composition, ambiguous issuance recovery, complete coverage/mutation, interoperability, release gates and production behavior remain unverified or unfinished; this is not unit/release completion.

### SHA-256

```text
ec0a3073b3a1b6459debd6e04c590a8b371c0864bd0e3aa0ba208e2656fb677e  otp/postgres/CHANGELOG.md
89d3a48d5694d58c16185321716a5f012aab18cd6411f1c22771341d7fbbd64f  otp/postgres/README.md
06535d83d7a08a4603db526339f5a8a377c25f415794b3f50d9d94e79306ad7e  otp/postgres/send_migration.go
dd7cee590002187bce7ef509469e5c4d6e569abf76ff665e7ba72366cd3370b8  otp/postgres/send_migration_test.go
a037301aafde4190ca74c67ef82a0771ec9b34e414d9cd75acb168646eedd6fb  otp/postgres/send_window.go
fc9a2c68af84acd89b99d28364c630efd631b139b3e9828bc2863ff334014109  otp/postgres/send_window_test.go
6f1ec7c35b2a2403b46a17bf990a4896991ab2cec231f413ef4a3fd9e891fdfa  otp/postgres/send_window_integration_test.go
```

## Send-family storage native regression and compatibility, base eafc154

Message Type: FINAL_ANSWER
Task name: /root
Sender: /root/otp_send_window_regression_eafc154
Payload:
## Native regression audit: APPROVE — bounded storage tranche

**Principal:** `/root/otp_send_window_regression_eafc154`, fresh read-only regression reviewer.  
**Boundary:** `feature/identity-platform` in `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`, base/HEAD `eafc154d7a466dc864613c3128d3dad1f697bf70`, plus the seven dirty paths below.

No concrete implemented-behavior defect identified.

### Assertions support the storage contract

- Family scope tests distinguish tenant/purpose/subject/channel/destination and preserve identity across challenge, target, authorization, generation and policy changes. Source excludes code-key identity and uses a separate HMAC domain with the existing pinned-key check.
- Unit tests exercise cooldown and hour-window equality and adjacent microseconds, cap denial, immutable/non-aliasing histories, corrupt timestamps/order/spacing, excessive history and SQL-precision rejection.
- Real PostgreSQL test sources assert persisted version/history, rollback preservation, two-connection cooldown exclusion, changed-key denial, family reuse, live-prefix preservation and expired-history pruning.
- Installed trigger tests reject deletion, history erasure, omitted live sends, identity changes, null instants and future additions, then independently inspect unchanged stored state.
- Forward migration is installed after existing populated challenge storage in integration fixtures; migration input tests cover unsafe schema/version, nil context and cancellation.

### Explicit evidence limits

- Parent reports compile-red/unit-green, final focused PostgreSQL integration/race **PASS 2.307s**, and vet exit 0. I inspected test assertions and current source but did not rerun Go or access services; those execution results remain parent-reported.
- Storage cancellation during lock wait, driver query/scan/exec failures, stale-CAS rejection, version exhaustion and policy-cap changes lack dedicated new assertions. Source fails closed on those paths; they remain targeted coverage gaps, not identified defects or grounds for a theoretical approval program.
- Concurrent test asserts two independent connection outcomes and persisted state, but does not deterministically force a particular lock-wait schedule.
- No full Send, challenge/replacement/delivery/journal atomic composition, original-flow admission, risk integration, clean-consumer, exact coverage/mutation or release qualification is certified.

Applied `testing-coverage` and `verification-before-completion` to constrain this approval to the implemented private family/window primitives.

`git diff --check` passed for tracked changes. No writes, Go/service commands or delegates.

### Reviewed SHA-256

```text
ec0a3073b3a1b6459debd6e04c590a8b371c0864bd0e3aa0ba208e2656fb677e  otp/postgres/CHANGELOG.md
89d3a48d5694d58c16185321716a5f012aab18cd6411f1c22771341d7fbbd64f  otp/postgres/README.md
06535d83d7a08a4603db526339f5a8a377c25f415794b3f50d9d94e79306ad7e  otp/postgres/send_migration.go
dd7cee590002187bce7ef509469e5c4d6e569abf76ff665e7ba72366cd3370b8  otp/postgres/send_migration_test.go
a037301aafde4190ca74c67ef82a0771ec9b34e414d9cd75acb168646eedd6fb  otp/postgres/send_window.go
6f1ec7c35b2a2403b46a17bf990a4896991ab2cec231f413ef4a3fd9e891fdfa  otp/postgres/send_window_integration_test.go
fc9a2c68af84acd89b99d28364c630efd631b139b3e9828bc2863ff334014109  otp/postgres/send_window_test.go
```

Message Type: FINAL_ANSWER
Task name: /root
Sender: /root/otp_send_window_compat_eafc154
Payload:
**APPROVE — source-level compatibility of this storage checkpoint only.**

Role: independent read-only Go compatibility specialist (`/root/otp_send_window_compat_eafc154`). Reviewed workspace `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`, branch `feature/identity-platform`, base/current HEAD `eafc154d7a466dc864613c3128d3dad1f697bf70`, plus the complete seven-path dirty change.

No concrete compatibility findings:

- `SendFamilyMigrationSource` is additive; existing consuming/check APIs, module manifests, dependency pins and applied migration sources remain unchanged.
- New forward-only migration requires existing lockout storage explicitly, uses separately allocated versions and does not rewrite populated challenge state.
- Private helpers introduce no coordinator substitute, transaction owner, public Send API or consumer-implemented orchestration requirement.
- Bounded, domain-separated family identity deliberately excludes challenge/target/generation/policy/authorization identities; stable abuse-key verification prevents silent key-driven quota reset.
- Family locking and caller-owned post-lock database-time sampling fit the existing transaction boundary. CAS/version handling and redacted errors follow existing conventions.
- No new dependency, permanent `replace`, or Go-target change; module still targets Go 1.27.

**Claim limits:** This is not OTP unit completion, clean published-consumer proof, release approval or ecosystem completion. Native Send authorization, digest insertion, replacement, risk, encrypted delivery and existing-journal orchestration remain future owned implementation. Complete exact coverage/mutation and other applicable release gates, published dependency/source pins, API baseline and clean-consumer proof remain pending. I ran no Go commands or services; supplied parent race/vet outcomes were not independently executed or used as compatibility certification.

Final inspected SHA-256 identities:
```text
otp/postgres/send_migration.go
06535d83d7a08a4603db526339f5a8a377c25f415794b3f50d9d94e79306ad7e
otp/postgres/send_migration_test.go
dd7cee590002187bce7ef509469e5c4d6e569abf76ff665e7ba72366cd3370b8
otp/postgres/send_window.go
a037301aafde4190ca74c67ef82a0771ec9b34e414d9cd75acb168646eedd6fb
otp/postgres/send_window_test.go
fc9a2c68af84acd89b99d28364c630efd631b139b3e9828bc2863ff334014109
otp/postgres/send_window_integration_test.go
6f1ec7c35b2a2403b46a17bf990a4896991ab2cec231f413ef4a3fd9e891fdfa
otp/postgres/README.md
89d3a48d5694d58c16185321716a5f012aab18cd6411f1c22771341d7fbbd64f
otp/postgres/CHANGELOG.md
ec0a3073b3a1b6459debd6e04c590a8b371c0864bd0e3aa0ba208e2656fb677e
```

## Metered-check native mapping and bounded QA, base bbdea02

Message Type: FINAL_ANSWER
Task name: /root
Sender: /root/otp_metered_check_mapping_bbdea02
Payload:
## Native change-map report

**Principal:** fresh read-only `change_mapper`, isolated context.  
**Boundary:** `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`, clean observed HEAD/base `bbdea0265ae7855290d0148abcac7ec3eef59fcd`. No files changed, tests run, or sibling review/ledger evidence consulted.

### Confirmed gap

The core already models successful metered non-consuming checks, but no native durable entry point commits them.

- `otp/challenge.go:195` — `Verifier.EvaluateAttempt` implements separately enabled/capped checks sharing the total comparison budget; correct checks propose `AttemptChecked`, never a reservation.
- `otp/postgres/contributor.go:116` — `Contributor.original` rejects every mode except `OTPConsume`.
- `otp/postgres/contributor.go:238` — denial execution returns `OTPAttemptUnchanged` for `AttemptChecked`; coordinator rollback would discard its counters.
- `postgres/otp_attempt.go:224` — coordinator supports committed denials only; unchanged becomes conflict/rollback.
- `postgres/otp_recovery.go:23` — attempt recovery accepts only pending or aborted/denied history, not a successful check.
- `otp/postgres/README.md:85` explicitly marks successful metered checks pending.

### Contract and ownership map

| Owner | Existing contract/path | Required preservation |
|---|---|---|
| OTP core | `AttemptIssuer.PrepareForCommand`; signed `CommandAttemptRecord`; `CheckPolicySpec`; `EvaluateAttempt` | Server-issued attempt identity; mode/code/full-context fingerprint; separate check cap **and** total budget; correct final check exhausts without wrong-code lockout |
| Kernel PostgreSQL | `PrepareOTPAttemptPlan` → locked command claim/journal → bounded `OTPAttemptWork` → audits/terminal receipt/commit | Sole transaction and replay authority; exact frozen plan; same-attempt replay before comparison; unknown commit classification |
| OTP PostgreSQL | `MapPreAuthWithAuthority` / `MapSessionWithAuthority` → `Contributor.original` → primary authority/flow → shared lockouts → challenge | Authenticate original complete context plus separately submitted bearer; challenge counter/state CAS; no private transactions |
| Session PostgreSQL | `RequirePreAuth` / `RequireSession` via `requireNativeFlow(..., nil)` | Read-only flow admission for checks; **do not call** `reserveNativeFlow`, finalization, issuance, release, or rebind |
| Consuming protocol | `PrepareOTPReservationPlan` / `ReserveOTPCommand` → `ReserveOTP` → `OTPApplyWork.CheckOTP` / finalize | Remains consuming-only. Existing `CheckOTP` means protected reservation validation, **not** user non-consuming check |

`otp/postgres/challenge.go:54` locks the exact full-scope challenge; `:151` already persists `AttemptChecked` proposals with counters, version, terminal state and a final database-time guard. Existing storage includes `checks_enabled`, `check_limit`, `used_attempts`, `failed_attempts`, and `checks`; no additional challenge-counter model is needed.

### Smallest coherent implementation boundary

1. Add a bounded native **metered-check coordinator phase**, distinct from consuming reservation and denial-only behavior.
2. Authenticate check-mode custody through the same original pre-auth/session and primary authority paths.
3. Reuse ordered shared-lockout/challenge locking and core policy evaluation; persist admitted correct **and incorrect** check proposals.
4. Atomically persist an explicit stable checked/denied receipt with counters and required audits in the existing command journal.
5. Provide exact command/attempt recovery for that terminal receipt without comparison or mutation.
6. Preserve all consuming-mode admission exclusions; successful check must expose no proof, reservation, apply handle, session effect, or consuming allowance.

Result persistence must be designed explicitly: `command.go:167` and `feature_receipt.go:17` expose structural receipts, while current denial recovery recognizes only denied terminal history. A transient boolean returned independently of the atomic journal would not satisfy stable replay/recovery.

### Concrete risks and missing executable behaviors

- Correct checks must charge `Checks` **and** `UsedAttempts`; otherwise successful-code probing or checks followed by consume weaken brute-force limits.
- Wrong check must charge once and trigger shared exhaustion lockout where required; correct final check exhausts but must not create wrong-code lockout.
- Disabled/capped/stale/terminal checks must not silently enter successful-check handling.
- Original bearer validation alone is insufficient: current locked primary authority, original flow, optional credential and post-wait time checks remain required.
- Mixed wrong/correct/check/consume execution must retain the same command→authority→flow→lockout→challenge ordering.
- Ambiguous commit must recover the same command and attempt; remapping changes encrypted intent and must conflict.
- Successful-check replay after exhaustion/expiry must return retained history without reevaluation.
- Existing production callers are consuming/denial paths; direct check operation/HTTP/service composition remains separate unfinished surface, not proven by this native slice.

### Requirements interpretation

No blocking contradiction identified. `TRANSACTION_CONTRACT.md:665` permits bounded non-consuming comparison while prohibiting reservation/consumption. `OPERATION_SEMANTICS.json:3425` labels HTTP check `GET`/read and forbids client idempotency keys; this does not erase mandatory internal metering or server-issued attempt identity. HTTP integration must preserve that distinction.

**Consumed prerequisites:** authentication opaque code/bearer contracts; identity command/feature/journal/query authorization; directory/user/credential authorization admissions; session original-flow contracts and primary checks; OTP retained digest/attempt/envelope keys; installed challenge and shared-lockout migrations. Their current release-gate readiness was not audited in this mapping.



Message Type: FINAL_ANSWER
Task name: /root
Sender: /root/otp_metered_check_qa_bbdea02
Payload:
## Native bounded QA artifact

**Role:** fresh read-only QA planner, `/root/otp_metered_check_qa_bbdea02`; `$qa-test-planner` fully read and applied.  
**Source:** `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`, `feature/identity-platform`.  
**Base/head inspected:** `bbdea0265ae7855290d0148abcac7ec3eef59fcd`.  
**Authority:** raw identity-platform goal, selected OTP/OTP-PostgreSQL goals, transaction contract, existing typed `OTPCheck`/`AttemptModeCheck` contracts and original-custody admission paths. No writes or checks executed.

### Observable contract

A new logical check must authenticate the original server-prepared submission and current native authority, then persist its bounded comparison charge and stable result atomically. Exact replay/recovery does not compare or charge again. Check success is strictly non-consuming: no reservation, proof, protected mutation, pre-auth consumption, session issuance/invalidation, or lifetime renewal.

Existing policy defines `UsedAttempts` as the shared comparison budget and `Checks` as an additional check-specific budget; both correct and incorrect checks count. Wrong checks additionally increment `FailedAttempts`. Successful exhaustion of the total budget produces checked/exhausted without wrong-code lockout.

### Smallest sufficient behavioral partitions

| Scenario / realistic defect | Observable assertions |
|---|---|
| **Correct check, both original pre-auth and authenticated-session variants:** success accidentally rolls back, reserves, or mutates authentication state | Durable result agrees with returned result; `UsedAttempts +1`, `Checks +1`, unchanged `FailedAttempts`, version advances once; no reservation ownership/proof; pre-auth/session/family records and authentication timeline unchanged; no domain/outbox/session effects. |
| **Wrong check versus wrong consume:** a second entry point bypasses accounting | Each logical admitted wrong submission increments total and failed counts exactly once; only check increments `Checks`; stable denial and required audit persist with those changes. |
| **Check-specific limit:** below, final allowance, subsequent attempt; disabled checks | Final permitted check counts once; next/disabled check cannot become successful or change comparison counters. Consuming allowance remains governed by the shared total budget rather than incorrectly adopting the check-only limit. |
| **Shared total limit:** penultimate/final correct check; final wrong check; later consume | Correct final check succeeds non-consumingly and leaves challenge exhausted with no wrong-exhaustion lockout; subsequent consume cannot reserve. Wrong final check exhausts and extends the exact existing shared abuse exclusions. |
| **Exact replay and fresh-process recovery:** retry charges twice or stale keys remap intent | Original command/AttemptID/keyed fingerprint recovers the same persisted result; challenge counters/version and audit count unchanged. Fresh codec/config instance with retained keys succeeds without code/bearer reconstruction or renewed timing. |
| **Changed-input reuse:** ID reuse swaps code, mode, tenant, purpose, challenge/version, authorization context, remember policy, bearer or authority snapshot | Same-command mismatch conflicts/fails closed with no counter/result/audit overwrite and no mutation of another challenge. Check-to-consume reuse cannot promote a checked result into authority. |
| **Current authority/custody denial:** stored declaration substitutes for possession or live admission | Wrong/absent bearer, incomplete legacy custody, wrong variant, revoked/expired original flow, session/family or directory/user/credential/authorization epoch mismatch deny; no successful receipt, reservation, or charged foreign challenge. Cover unsupported optional session authority and both configured installation variants. |
| **Post-lock-wait expiry:** pre-wait time authorizes an expired submission | Hold the actual challenge lock, allow original flow or credential deadline to pass, then release: no checked success. Separately cross challenge expiry and assert the defined expired result/state without comparison-budget renewal. Cover exclusive equality in deterministic policy tests and real database elapsed-time behavior. |
| **Concurrent exact duplicates:** command locking still permits duplicate accounting | Multiple identical executions return the same terminal observation, with one committed comparison charge and one required result/audit set. |
| **Concurrent new checks / mixed check-wrong-consume:** independent paths overspend or create multiple winners | Barrier-start native submissions through real PostgreSQL; persisted counters never exceed either cap, versions match admitted changes, at most one consuming reservation, no check reservation. Include stale expected-version competitors: conflicts are acceptable and must not count as admissions. |
| **Brute-force/reissue/purpose switching:** check becomes a parallel budget or resets shared exclusion | Mix wrong checks and consumes to exhaust the shared cap; reissue/change purpose/flow retains subject/destination lockout. Other tenant/destination remains isolated. Denied attempts cannot obtain success by switching to the check entry point. |
| **Rollback / failed audit / cancellation / bounded-work failure:** partial counters or orphan success | From independent connections assert no persisted counter/version/result/audit subset after rollback; returned outcome is not success. Work handles become unusable after callback return and all rows/transactions terminate. |
| **Commit ambiguity, both actual commit and actual rollback:** uncertainty rewritten as clean failure or retry compares twice | Inject lost commit acknowledgement/disconnect; initial result is unknown, never guessed success/not-committed. Recovery of the exact command identifies committed checked/denied result or authoritative noncommit; counters/audits remain exactly once. Pending uncertainty cannot allocate a replacement submission. |
| **Secret handling / bounds / compatibility:** receipt becomes bearer or public API breaks | Errors, formats, serialized result/recovery metadata and persisted records expose no raw code/bearer/private scope; oversized/malformed inputs fail before costly work. Existing denial and consuming consumers retain behavior; checked output cannot enter reservation/finalization APIs. |

### Verification selection

1. **Focused red first:** native successful check through the coordinator and actual OTP contributor, asserting committed result plus persisted counters and absence of reservation/authentication effects. Current `RecordOTPAttemptDenial` explicitly returns unchanged for `AttemptChecked`; a pure `EvaluateAttempt` test cannot prove this missing durable behavior.
2. **Deterministic policy/unit tests:** cap boundaries, disabled checks, correct-versus-wrong final exhaustion, mode/fingerprint mismatch, result redaction and callback lifetime. Reuse existing policy tests where they already prove the invariant.
3. **Real PostgreSQL integration:** success/replay/restart, both custody variants, changed intent, rollback/audit failure, post-wait expiry, same-ID concurrency and mixed submissions. Extend native composition fixtures; coordinator-only synthetic-counter fixtures are supplemental, not native OTP proof.
4. **Concurrent boundary:** targeted race run and bounded stress of mixed submissions plus exact duplicate replay; assert persisted outcomes, not merely goroutine completion.
5. **Ambiguity boundary:** real-driver commit-acknowledgement fault/disconnect combined with credential-free exact recovery; mocked transaction return values alone are insufficient.
6. **Affected consumer/public-contract checks:** root identity, coordinator PostgreSQL, OTP and OTP PostgreSQL consumers; source locks and clean-consumer composition must exercise the changed bridge. Preserve existing consuming, denial, session admission and lockout regressions.
7. **Release boundary:** affected repository gates, exact statement coverage and viable-mutant killing remain required before final verification. Target mutants that omit a charge, skip check cap/current admission/post-wait expiry, treat checked as reserved, replay comparison, or split result/audit persistence.

Suggested execution families, from the respective module directories:

- `go test ./...`
- `go test -tags=integration ./...`
- `go test -race -tags=integration -run '<new check tests and affected concurrent native tests>' ./...`

Use task-owned disposable `GOCACHE` for every run and remove it afterward. Integration evidence must show a configured real PostgreSQL DSN, not skipped tests; OTP fixtures use `IDENTITY_OTP_POSTGRES_TEST_DSN`. Parent should retain the exact new test names and actual commands/outcomes, rather than treating these suggested families as executed evidence.

**Disposition:** No contradictory requirement, unsafe requested action, or reproducible design defect identified. Existing provisional checked decisions and durable unchanged behavior establish the missing behavior, not a reason to block implementation. Successful-result vocabulary is an implementation contract decision to make explicitly; it must preserve the transaction owner and must not falsely describe a non-consuming check as a protected workflow commit.


## Local source checkpoint bbdea0265ae7855290d0148abcac7ec3eef59fcd

Message Type: FINAL_ANSWER
Task name: /root
Sender: /root/otp_native_session_delivery_bbdea02
Payload:
## APPROVE — local development checkpoint only

**Native principal:** `/root/otp_native_session_delivery_bbdea02`  
**Exact source:** `feature/identity-platform`, HEAD `bbdea0265ae7855290d0148abcac7ec3eef59fcd`, parent `bcad0046199bb7934c762f77725cefcf7efdbd7d`.

- Source checkout clean; commit contains exactly 14 files, 966 additions/58 deletions. `git diff HEAD^ HEAD --check` passed; coherent change/rationale commit shape.
- Inspected retained native mapper/QA covering continuous authenticated-session OTP work, native regression/compatibility reports, and accepted blind final review from `/root/otp_native_session_blind_final_bcad004`.
- **All 14 committed blob SHA-256 values match the accepted blind review.** Earlier peer-exposed `otp_native_session_final_bcad004` was excluded from final-gate authority.
- Regression’s two older anchors belong to `consuming_execution.go` and `contributor.go` before subsequent comment-only corrections. Preserve its original reviewed boundary; do not represent those anchors as final-byte evidence. Parent confirms runtime/test code unchanged, and accepted blind review covers final committed bytes.
- Retained parent evidence reports both affected modules’ full integration/race suites and vet passing. No Go checks rerun by this auditor.
- Owned temporary directory and named PostgreSQL container are absent. Pre-existing image preservation remains parent-reported; no cleanup performed here.

**Boundary:** committed locally only; no push, hosted CI, release, or production claim. Counts remain **5/67 integrated, 1/67 verified**, with no unit promotion. Complete login/shared issuer, successful metered checks, clean consumers/source pins, exact coverage/mutation and full release gates remain pending.

Applied `$verification-before-completion` to constrain approval to directly inspected delivery state and retained evidence.


## Accepted blind native final review

Message Type: FINAL_ANSWER
Task name: /root
Sender: /root/otp_native_session_blind_final_bcad004
Payload:
## Disposition: APPROVE — bounded local development batch

**Native principal:** `/root/otp_native_session_blind_final_bcad004`, independent read-only correctness reviewer.

**Reviewed boundary:** `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`, branch `feature/identity-platform`, base/HEAD `bcad0046199bb7934c762f77725cefcf7efdbd7d`, plus all 14 dirty tracked/untracked files listed below. No files modified, code executed, or other review conclusions consulted.

No concrete correctness findings in the complete final diff.

Source-backed closure:
- Native session execution remains explicit opt-in; existing `ContributorConfig` shape and pre-auth constructor remain compatible.
- Separate token possession, exact retained key, original authority keys/epochs, current primary credential/directory/authorization and complete session/family agreement precede OTP effects.
- Session/family share locks precede challenge exclusion. Post-wait primary time rechecks credential expiry, session lifetime and original freshness; takeover cannot replace custody or renew deadlines.
- Protected checks/finalization retain exact OTP ownership and generation. Sessions are neither reserved nor consumed.
- Rollback/restart recovery requires authenticated retained custody and exact reserved predecessor evidence; coordinator unanimity owns noncommit release. Missing or divergent evidence remains unresolved rather than granting release.
- Changed and affected test oracles cover independent persisted state, counters, wrong-token denial, stale authority, deadline equality/waits, generation takeover, rollback and fresh-instance recovery. Domain/audit/lifecycle implementations remain deliberate SQL test seams, not proof of complete owning-workflow interoperability.

**Verification boundary:** Parent supplied passing `go test ./... -tags=integration -race -count=1` and `go vet ./...` for both affected modules, plus focused native red/green evidence. I did not independently rerun them. My `git diff --check -- otp/postgres session/postgres` passed. Coverage, mutation, clean-consumer/source-pin, actual complete workflow and release gates remain pending; this approval does **not** establish unit/platform verification, release readiness or full login completion.

### Final-file SHA-256

```text
07503d71031754de9bf0f81b09f50aa3da6797ba8db9b51c4a0a7309f2ea1b1b  otp/postgres/CHANGELOG.md
59096cfa812db7004d3dc74fa1b8d0c6157cb33e863c58d0f013e65861bbb8fe  otp/postgres/README.md
1a6296852eb98b62d3969651232734a29605c256980cb689747c8175bc2453b3  otp/postgres/consuming.go
fc2e76f6da000beaaab5f4cbacdc15950ec5f75a1b76a8cd6f0ba1326ef9ee1b  otp/postgres/consuming_execution.go
55b18970670f479f893c8a6ed466ebfe82b5eff10d88b2c80e791a442e245568  otp/postgres/contributor.go
01a3cc559184120570ddc5edb38f21d25008525b16aee8aee5ca72ab8f4ccd9b  otp/postgres/contributor_integration_test.go
3aae93d75c400d57e3fd411bb3c72ec1d3bb81e6d65b9bf51d1d3945cfc1ec01  otp/postgres/native_flow.go
2aae84cc841ad40e112599d78681cc6ec01e3cbf3a058deacc63f0778c0af309  otp/postgres/session_admission.go
e8ac9f3ed1ab73e5a1878166a1f64be7eaed71f8642b9d9a8a0f6b6750df2bf5  otp/postgres/session_admission_integration_test.go
3136548e463e09fe1744622b0f5a2b38b856ef968d9b5954d7f2987c64dd4941  otp/postgres/session_admission_test.go
1d73b6fae75b19583bcd007660a8331f47f3e0b496cda06a4aa861d0e81b10d7  session/postgres/CHANGELOG.md
cdb629bc58b60aaa0fdd979f98233e5e07ba334b9fed7216e209fcb38161d1fe  session/postgres/README.md
ff6454b437c6979b557d5f572138d8788671fb3c851fbf2006f965f98bb1d11e  session/postgres/original_admission.go
b1dd2d742444ee74f0b0c9046ae566d864dcfdb2096ef671ba180a118b39bd7b  session/postgres/original_admission_test.go
```


## Source review not accepted as blind gate

This report disclosed sibling-report exposure; retain its native record but
DO NOT use it to satisfy independent final correctness review. Replacement
otp_native_session_blind_final_bcad004 is running with role evidence excluded.

Message Type: FINAL_ANSWER
Task name: /root
Sender: /root/otp_native_session_final_bcad004
Payload:
## Source review: APPROVE — bounded local development only

**Principal:** `/root/otp_native_session_final_bcad004`  
**Repository:** `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`  
**Branch/base/head:** `feature/identity-platform`, `bcad0046199bb7934c762f77725cefcf7efdbd7d`  
**Boundary:** all 14 current tracked/untracked files under `otp/postgres` and `session/postgres`, including final comment corrections.

No concrete correctness findings.

Source-backed conclusions:
- `native_flow.go:36` authenticates disjoint original custody and separately submitted bearer; `session_admission.go:83` compares session-owner-derived expectations rather than trusting arbitrary authority projections.
- `session/postgres/original_admission.go:33` delegates to the existing owner derivation, preserving tenant, actor, audience, credential, scope-key and optional-dimension checks.
- `consuming.go:90` rechecks current credential expiry and original session lifetime/freshness after downstream waits. Exclusive freshness equality and future originals deny.
- `consuming_execution.go:76` retains exact command/envelope/reservation matching during protected execution; authenticated sessions receive no reservation, finalization or release mutation.
- `consuming_execution.go:136` and `:182` use sealed noncommit phases and exact retained OTP predecessor evidence. Revocation does not prevent proven rollback; missing/divergent evidence remains unknown.
- Existing `ContributorConfig` and pre-auth construction remain compatible; native session execution is explicit opt-in.

**Verification:** independently inspected complete diff/new files and `git diff --check` PASS. Parent-supplied full integration/race tests and vet results were not rerun. Coverage, mutation, source-lock/clean-consumer, complete login and release gates remain pending.

**Independence limitation:** after the causal source inspection, opening the supplied evidence note exposed newly prepended current sibling regression/compatibility reports. Those conclusions were not used as authority, but this report **does not satisfy a fully blind independent final-review gate**. Parent has acknowledged and arranged a replacement.

### Final reviewed SHA-256 values
```text
07503d71031754de9bf0f81b09f50aa3da6797ba8db9b51c4a0a7309f2ea1b1b  otp/postgres/CHANGELOG.md
59096cfa812db7004d3dc74fa1b8d0c6157cb33e863c58d0f013e65861bbb8fe  otp/postgres/README.md
1a6296852eb98b62d3969651232734a29605c256980cb689747c8175bc2453b3  otp/postgres/consuming.go
fc2e76f6da000beaaab5f4cbacdc15950ec5f75a1b76a8cd6f0ba1326ef9ee1b  otp/postgres/consuming_execution.go
55b18970670f479f893c8a6ed466ebfe82b5eff10d88b2c80e791a442e245568  otp/postgres/contributor.go
01a3cc559184120570ddc5edb38f21d25008525b16aee8aee5ca72ab8f4ccd9b  otp/postgres/contributor_integration_test.go
3aae93d75c400d57e3fd411bb3c72ec1d3bb81e6d65b9bf51d1d3945cfc1ec01  otp/postgres/native_flow.go
2aae84cc841ad40e112599d78681cc6ec01e3cbf3a058deacc63f0778c0af309  otp/postgres/session_admission.go
e8ac9f3ed1ab73e5a1878166a1f64be7eaed71f8642b9d9a8a0f6b6750df2bf5  otp/postgres/session_admission_integration_test.go
3136548e463e09fe1744622b0f5a2b38b856ef968d9b5954d7f2987c64dd4941  otp/postgres/session_admission_test.go
1d73b6fae75b19583bcd007660a8331f47f3e0b496cda06a4aa861d0e81b10d7  session/postgres/CHANGELOG.md
cdb629bc58b60aaa0fdd979f98233e5e07ba334b9fed7216e209fcb38161d1fe  session/postgres/README.md
ff6454b437c6979b557d5f572138d8788671fb3c851fbf2006f965f98bb1d11e  session/postgres/original_admission.go
b1dd2d742444ee74f0b0c9046ae566d864dcfdb2096ef671ba180a118b39bd7b  session/postgres/original_admission_test.go
```


## Native session postimplementation reports

Message Type: FINAL_ANSWER
Task name: /root
Sender: /root/otp_native_session_regression_bcad004
Payload:
## Native regression audit: APPROVE

**Principal:** `/root/otp_native_session_regression_bcad004`  
**Repository:** `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`  
**Branch/base/head:** `feature/identity-platform`, `bcad0046199bb7934c762f77725cefcf7efdbd7d`  
**Reviewed boundary:** complete 14-file tracked/untracked dirty delta, including the final comment corrections.

No concrete regression or blocking behavioral coverage gap found for this bounded local-development tranche.

- Explicit session opt-in preserves pre-auth construction and disjoint flow handling.
- Submitted bearer authentication, exact original authority derivation and primary session/family comparison remain separate.
- Post-lock checks preserve original freshness, session lifetime and credential expiry.
- Real PostgreSQL tests assert single-charge denial/replay, generation takeover, atomic protected consumption and unchanged session state.
- Rollback and restart recovery use exact retained OTP custody; revoked sessions do not strand proven noncommit, and unreadable custody remains reserved/unknown.

Applied `testing-coverage` to assess observable assertions and realistic defects, not planning-artifact completeness. No files changed or Go/service checks executed.

Parent-supplied results: both full integration race suites and both vet runs passed. Coverage, mutation, clean-consumer/source locks, complete login and release verification remain pending; this approval does not promote either unit to verified.

**Snapshot anchors:**
```text
otp/postgres/consuming.go 1a6296852eb98b62d3969651232734a29605c256980cb689747c8175bc2453b3
otp/postgres/consuming_execution.go 765392af23eed9e409f724eed11e11d1205edd028b3baec95f9eca1f285a364c
otp/postgres/contributor.go 31c43b8e4b68bec32c3e2d04ae72426251103e9749773b1a39c3ce4d41069f29
otp/postgres/native_flow.go 3aae93d75c400d57e3fd411bb3c72ec1d3bb81e6d65b9bf51d1d3945cfc1ec01
otp/postgres/session_admission.go 2aae84cc841ad40e112599d78681cc6ec01e3cbf3a058deacc63f0778c0af309
otp/postgres/session_admission_integration_test.go e8ac9f3ed1ab73e5a1878166a1f64be7eaed71f8642b9d9a8a0f6b6750df2bf5
otp/postgres/session_admission_test.go 3136548e463e09fe1744622b0f5a2b38b856ef968d9b5954d7f2987c64dd4941
session/postgres/original_admission.go ff6454b437c6979b557d5f572138d8788671fb3c851fbf2006f965f98bb1d11e
session/postgres/original_admission_test.go b1dd2d742444ee74f0b0c9046ae566d864dcfdb2096ef671ba180a118b39bd7b
```

Message Type: FINAL_ANSWER
Task name: /root
Sender: /root/otp_native_session_compatibility_bcad004
Payload:
**APPROVE — limited local-development compatibility.**

Reviewer: `/root/otp_native_session_compatibility_bcad004`. Reviewed base/head `bcad0046199bb7934c762f77725cefcf7efdbd7d`, plus every current tracked and untracked change under `otp/postgres` and `session/postgres`.

- No blocking API or storage compatibility defects found.
- `ContributorConfig` remains unchanged; `NewContributor` retains pre-auth requirements. Authenticated sessions require additive explicit configuration.
- Public session-owned derivation preserves original authority expectations; existing codecs, migrations and module boundaries remain unchanged.
- Retained keys are copied; configuration and errors remain redacted.
- Reported stale public comments were corrected and inspected.
- `git diff --check` passed. No Go commands executed by this reviewer.

Clean-consumer dependency pins, release gates and full unit/platform verification remain pending; this approval does not certify them.




## Native authenticated-session execution, base bcad004

Current dirty implementation uses explicit session configuration without
changing ContributorConfig or pre-auth construction. Original owner
expectations are derived without IO; native phases authenticate possession,
check current primary owners and recheck original temporal bounds after waits.
Sessions are not one-time effects. OTP predecessor evidence controls sealed
release/recovery; ambiguity retains reservations.

Focused native session regression/race PASS39c36b (5.897s). Full affected
`go test ./... -tags=integration -race -count=1` passes in OTP PostgreSQL
(c0b67d,19.060s) and session PostgreSQL (1c5778,14.662s). Subsequent
`go vet ./...` in both modules exits0 (c0b67d,5988d2). Composed disposable
workspace, Go1.27.1 and task-owned PostgreSQL18.3; not clean-consumer or
release proof. Independent postimplementation review pending; no unit promotion.

## Local custody checkpoint bcad0046199bb7934c762f77725cefcf7efdbd7d

Source commit preserves all 24 reviewed blobs; source clean/local only.
All three post-implementation native reviews and exact-head audit APPROVE for
local development. Focused results retained below; no full release/unit claim.
All task-owned Go/cache/message resources removed after their bounded use.

Message Type: FINAL_ANSWER
Task name: /root
Sender: /root/otp_session_custody_delivery_bcad004
Payload:
## Native delivery audit

**Principal:** `/root/otp_session_custody_delivery_bcad004`  
**Disposition: APPROVE — LOCAL DEVELOPMENT checkpoint only.**

- Source worktree is clean on `feature/identity-platform`, HEAD `bcad0046199bb7934c762f77725cefcf7efdbd7d`, parent `5e7206b6478fcfc48607f871400a75b68e5f017b`.
- Exactly 24 committed paths; every committed blob matches the native final correctness report’s SHA-256 entry.
- Inspected retained native mapper, QA, regression, compatibility and correctness reports: distinct named principals; implementation reviews cover the complete 24-file dirty boundary atop the exact parent. Regression, compatibility and correctness dispositions are APPROVE for local development only.
- Parent-reported four-module race/vet successes remain attributed to the parent; this auditor executed no Go or service checks.
- Local commit established. No local remote-tracking ref or tag contains it. Push, hosted CI, release and deployment are not established or claimed.
- Native authenticated-session SQL execution remains disabled. Clean-consumer source pins, exact coverage/mutation and release gates remain pending.
- Ledger counts remain **5/67 integrated, 1/67 verified**; no OTP-unit promotion or full OTP/platform readiness claim.

### Retained native final correctness report

Message Type: FINAL_ANSWER
Task name: /root
Sender: /root/otp_session_custody_final_5e7206b
Payload:
## Native independent correctness report

**Role:** independent read-only correctness reviewer `/root/otp_session_custody_final_5e7206b`  
**Disposition:** **APPROVE — LOCAL DEVELOPMENT tranche only. No findings.**

**Boundary:** `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`, branch `feature/identity-platform`, base/HEAD `5e7206b6478fcfc48607f871400a75b68e5f017b`, plus the complete 24-file dirty diff. Initial and final file hashes matched. Single-reviewer structured pass using `$requesting-code-review`, applicable coverage guidance and `$verification-before-completion`; no writes, Go execution, services, external mutations or delegation.

### Source-backed closure

- **Original session custody:** `otp/session_authorization.go:18` rejects zero/anonymous records. `otp/authorization.go:170` rejects authenticated-session signin and nonzero continuation remember policy; `:194` enforces exact tenant/subject. `otp/session_authorization.go:42` commits every applicable field of the session-owned record, complete authority vector and optional presence, lookup digest/version, resolved remember policy and original authentication timeline. Verified metadata requires zero `AnonymousAt` under unchanged `session/record.go:94`.
- **Restart representation:** `otp/postgres/command_record.go:30` selects the additive v4 frame; `otp/postgres/session_record.go:54` reconstructs through the session owner’s constructors. Malformed lengths, invalid flags/times, mixed selected IDs, truncation and trailing bytes fail closed. Existing v1/v2 encoding remains unchanged; the authority envelope’s inner-frame acceptance adds v4 without weakening signature or AAD verification in `otp/postgres/command.go:275`.
- **Separate submitted bearer:** `feature_session_token.go:21` bounds/canonicalizes transport syntax and retains opaque custody separately from scalar metadata and the OTP code. `feature_command.go:122` rejects internally mixed bearer custody. Public constructors cannot manufacture both profiles.
- **Digest and replay identity:** `otp/postgres/session_token.go:122` selects only the original lookup-key version; `:130` checks the derived tenant-bound digest and its version, with no fallback or original-record replacement. `postgres/codec.go:141` binds the token under a distinct command-secret domain. Unchanged `postgres/claims.go:97` recomputes replay intent from retained original custody using the recorded key.
- **Protected-phase lifetime:** `postgres/otp_execution.go:109` delegates to the established phase, coordinator, generation and exact-owner guards before borrowing the bearer. Returned, foreign, zero, absent and pre-auth handles deny. Release/recovery have no bearer accessor; scalar and safe collision/recovery projections do not carry the token/code.
- **Adversarial challenges:** wrong bearer with unchanged metadata, missing/replaced lookup key, changed plan, anonymous or foreign-subject original, mixed bearer custody, and copied returned phase handles all reach explicit denial guards. Generic execution still rejects code-bearing commands at `postgres/coordinator.go:182`. The native contributor still requires complete pre-auth custody/cookie at `otp/postgres/contributor.go:106`; this tranche does not accidentally enable authenticated-session SQL execution.
- **Tests:** inspected all changed test assertions and their fixtures. Literal HMAC framing independently checks the fingerprint domain; original-record comparisons check restoration; field mutations check commitment differentiation; negative cases check scope/profile/key/lifetime guards. No added goroutines, callbacks, SQL operations, migrations or durable-state transitions require additional execution claims in this bounded tranche.

### Verification boundary

`git diff --check` passed. The parent supplies final `go test . -race -count=1` and `go vet .` exit-zero results for root, `postgres`, `otp` and `otp/postgres`; **I did not independently execute or certify those runs**.

This approval permits the coherent reviewed **local development batch**, not package verification, clean standalone adoption, release or platform completion. Native authenticated-session primary admission, consuming/takeover/finalization/release/recovery execution, successful metered checks, native SQL journeys, clean-consumer dependency pins, exact coverage/mutation and applicable release gates remain pending. Old binaries cannot consume the new opt-in frame; mixed-version release compatibility remains a later release obligation.

### Reviewed changed-file SHA-256

```text
6665d4b15f8cfc11317ec8b50899fa226e3611f746020bd16b5aad1db288117e  CHANGELOG.md
22b3e546450f0690fc3608b44fa3856b184ee813434571f81025148dfba8e23a  README.md
12856eb4eda3f215979c5d9240af19a34d3f7588f5415de8579792803d793605  feature_command.go
15aa710e567e4f8b139d2f03d3ab63ca0b0e93acc150bcfc7836e638215c7f3a  feature_command_internal_test.go
5318e77acfbc30ea367e1f5545cb17eccdf0efc16ea721a3de46b194d9751b5b  feature_session_token.go
7ce28a87c8366c95b8a5931d692c7f9c464bcc218b5ca86c8b23cc29c6c9a13b  feature_session_token_test.go
e64ae8024a69a8038d9a255bca053753e8774239dbc776cc99918c433f850652  otp/CHANGELOG.md
7850fb9836269fe6767c4f32b4a229482b049efc29c5f7e8ec527d219ccad0fd  otp/README.md
7b41c2e23d66f03d2449c6a5b829c551401f68899a9aca0d2f5e8adbd911241c  otp/authorization.go
7c8bd0a0fd96fbb81c45baa17f2559e868efbd635ec8f1a49bbf2bb318169e75  otp/session_authorization.go
d926a500d10546470ca4b30a4fe42742f2bf17be9e52e86125eb3942902d2cf2  otp/session_authorization_test.go
94b9be049dbb8056ae708e4126d83be3d8e3dea993750d5103b7809a9877f384  otp/postgres/CHANGELOG.md
5de031554a863104290cd809d03b745bd524e94f55299e2a8fb03704fe34c19f  otp/postgres/README.md
fc47afb9fdf070ac65d02f5dacc9ba266ecb5870485361f9a8c2cf533320bd33  otp/postgres/authority_custody.go
bcc3e0cdb35ddfdcbe6b98143772ac95b8ba92eeec6312f7c7c521912524e109  otp/postgres/command_record.go
aa44369c3adf572380be3eaa70696c262d40f227dd3f82b4554093f477122d5b  otp/postgres/session_record.go
333f85722f0b84ab571d963d6866e127bf90426357de0d488eef01d153d482b4  otp/postgres/session_record_test.go
6ce12afd0ac2627fbd33c4b6030376f14a7cc126f284f731ffe092aaa7f5b222  otp/postgres/session_token.go
f67e45ca13c8e6e13266e6b7ad71169363ddb6dff07e2eb4c6207907790e2bfa  otp/postgres/session_token_test.go
ea1e0eecf5b7df560f1e78a32e9ce7c8d04c74d18a7567d7bc678c7cfdf8a526  postgres/CHANGELOG.md
359498803392b9819ce7e33b2dd6f7433b92a2d0c8d2c84222ac359fa44e3930  postgres/codec.go
29dfdc09a4730edccedbd7c9d7817880dbe80ba55fa52d2631b14a72908e5130  postgres/codec_session_token_test.go
b00d7f4712fa16ecb048371ebdf23b8165a3dc4f3a8ccb972e2dc4e522b05d4f  postgres/otp_execution.go
c0e4330486deda899e1ce49d0005c78dec7a96df81e3f35050aa26942f674f86  postgres/otp_phase_session_test.go
```

## Complete authenticated-session custody tranche — 2026-10-07

Complete original session record, v4 encrypted restart custody, separately submitted
session bearer, retained-key command fingerprint and protected-phase accessors are
implemented locally across four modules. Native authenticated-session SQL execution
remains disabled pending primary authority and transactional consuming phases.
Final `go test . -race -count=1` and `go vet .` exit0: root47157a 1.534s,
postgres63fe69 3.119s, otp687b73 1.230s, otp/postgres e4db68 1.168s.
Bound-session record fuzz PASS604aee 4.259s, 33389 executions. Composed sources
are not clean-consumer/release evidence. Regression and compatibility native
reports follow; final correctness review pending. No source commit/unit promotion.

### Retained native report: `/root/otp_session_mapping_5e7206b`

Message Type: FINAL_ANSWER
Task name: /root
Sender: /root/otp_session_mapping_5e7206b
Payload:
## Native follow-up mapping

**Role:** read-only `change_mapper`. **Baseline:** committed `5e7206b6478fcfc48607f871400a75b68e5f017b`. Inspected root/kernel contracts from that commit; current dirty OTP additions were not reviewed for correctness. No writes or execution.

### Exact custody path

1. **Root command:** `featureCommandState` holds structural spec, opaque OTP code and a separate pre-auth bearer (`feature_command.go:76`). `Feature()` returns cloned scalar metadata only. `OneTimeCode()` and `PreAuthCookie()` require the exclusive valid feature branch.
2. **Fingerprint:** `postgres/codec.go:119–142` binds code and pre-auth bearer independently into canonical command payload using dedicated keyed domains. Raw canonical payload is not persisted; durable claims retain keyed fingerprint, key version and safe scope.
3. **Replay:** `postgres/claims.go:91` re-encodes the original opaque command using the **stored fingerprint-key version**, not current issuance key. A new session-token custody branch must participate in that same regeneration.
4. **Counted phases:** denial and reservation lend the exact native command through owner-fenced `NativeCommand`. They can obtain the submitted code and whichever distinct bearer variant the command retains.
5. **Protected phases:** `OTPApplyWork`/`OTPFinalizeWork` currently expose only code-free `CommandSpec` plus exact-owner `PreAuthCookie` (`postgres/otp_execution.go:87–107,141,190`). A session-token accessor must use the same owner, phase, generation, callback-lifetime and Work checks; it must not expose the native command or OTP code.
6. **Recovery/release:** `RecoveryWork.OTPBinding` and `OTPReleaseWork.Binding` expose frozen command/plan/fingerprint/generation correlation, not submitted code or bearer. Native restart recovery reconstructs encrypted original custody from challenge storage. Preserve that separation.

### Compatibility constraints

- Distinguish session-token and pre-auth-cookie custody with separate opaque members/accessors and mutually exclusive construction. Do not relabel pre-auth bearer custody as an authenticated session.
- Root cannot import `identity/session`: session already imports root, creating a cycle. Existing `authentication.BearerCredential` can carry opaque submitted bytes without creating another token owner.
- Keep existing constructor/accessor behavior and pre-auth fingerprint bytes unchanged. Add a distinct payload key and secret-digest domain only when session custody exists; old commands must not gain a zero/default field.
- Session-token substitution must change command intent even when encrypted continuation, OTP code and session ID match. Retained-key replay must bind the exact submitted token again.
- Reject mixed bearer variants and absent required OTP custody. All feature projections, safe collision fields, formatting, receipts and recovery outputs remain raw-secret-free.
- Current secret bytes must never enter encrypted continuation persistence; only the original session’s existing token digest belongs there.

### Existing session owner APIs

- `session.ParseToken`: validates canonical bounded 32–64-byte token encoding.
- `session.NewDigestKey`: owns retained, versioned lookup-key material.
- `Token.Digest`: computes existing tenant/version/domain-bound session lookup digest. Recompute using the **original session’s `Lookup.KeyVersion`**, then compare against its original digest; do not substitute current key or accept a caller digest.
- `sessionpostgres.RequireSession`: compares complete frozen original session/family under primary locks and checks database-time expiry after lock acquisition.
- `SessionService.Get`/`GetResult.ValidateAt`: available when authenticated-reader provenance is required. Structural session reconstruction alone is not authentication.

These APIs avoid another credential owner, token format or replay store. Current `RequireSession` does not establish other authority dimensions or operation permission; those remain existing primary-owner/policy checks. No additional plan approval is needed from this report.

### Retained native report: `/root/otp_session_custody_compatibility_5e7206b`

Message Type: FINAL_ANSWER
Task name: /root
Sender: /root/otp_session_custody_compatibility_5e7206b
Payload:
Role: `go_compatibility`  
Principal: `/root/otp_session_custody_compatibility_5e7206b`  
Repository: `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`  
Base/current HEAD: `5e7206b6478fcfc48607f871400a75b68e5f017b`  
Reviewed state: complete tracked and untracked dirty tranche; read-only inspection, no Go/service execution.  
Disposition: **APPROVE for this local development tranche**; no concrete new compatibility regression found.

- Public constructors/accessors are additive. Existing constructor signatures and scalar projections remain unchanged. The new private `Feature()` denial affects internally forged mixed bearer custody, not previously constructible public commands.
- Existing ID-only authorization v1, pre-auth v2 and command fingerprints retain their encoding paths. Complete sessions deliberately select distinct authorization v3/private-record v4 profiles; session token intent uses a separate keyed domain.
- New readers retain v1/v2 and authority-envelope v3 restoration. Old readers reject v4 directly or nested inside v3, rather than silently downgrading session custody. Deploying v4 writers therefore requires upgraded readers; this is opt-in and documented.
- Session decoding applies bounded field reads, canonical UTC timestamp checks, verified-only authentication, owner constructors, exact family/session versions, optional-presence validation and trailing-byte denial. No truncating field conversion or newly introduced dependency cycle found.
- Root retains bearer syntax/custody only; session owns lookup policy; OTP depends on session; adapters compose those public contracts. No permanent `replace` directives introduced.

**Deferred release gaps, not new tranche defects:** pinned module versions do not include the newly consumed APIs, so composed-workspace results cannot establish clean-consumer compatibility. Native authenticated-session SQL admission/consumption, full OTP behavior and release gates remain unfinished. This approval does not establish standalone module buildability, readiness, release completion or platform completion.

`git diff --check` passed. Parent-reported race/vet results were not independently executed or certified here.

### Retained native report: `/root/otp_session_custody_regression_5e7206b`

Message Type: FINAL_ANSWER
Task name: /root
Sender: /root/otp_session_custody_regression_5e7206b
Payload:
## Native regression report

**Identity:** independent read-only reviewer `/root/otp_session_custody_regression_5e7206b`.  
**Reviewed source:** selected identity worktree, HEAD/base `5e7206b6478fcfc48607f871400a75b68e5f017b` plus complete 24-file tracked/untracked development delta. Individual file hashes captured in tool output.  
**Method:** `$testing-coverage`, source/test inspection; no writes, Go execution, services or delegation.  
**Disposition: APPROVE** for this bounded custody development tranche. No blocking observable regression identified.

### Implemented coverage

- Root tests protect separate session/pre-auth bearer custody, canonical transport bounds, zero-code/scope rejection, redaction and scalar reconstruction denial.
- Core tests protect original session/timeline retention, session-versus-legacy framing, tenant/subject isolation, anonymous/signin rejection and commitment divergence across authentication, lookup and authority changes.
- Codec round trips through a fresh instance compare the complete session, remember policy and authority vector, including all optional dimensions.
- Private-record tests reject every truncated prefix and trailing bytes; the new fuzz oracle checks canonical reconstruction.
- Mapping tests reject wrong/absent/pre-auth credentials, missing/replaced retained keys, mismatched plans and canceled contexts.
- Coordinator tests protect domain-separated secret fingerprints, retained-key replay and exact live-owner borrowing during check/finalize.

Parent-supplied red/green records and final four-package race/vet successes support these focused contracts. They were not independently executed here.

### Nonblocking coverage limitations

- `MapSessionWithAuthority` is not exercised directly: authority wrapping and session-token mapping are tested separately.
- Optional authority presence is tested, but optional scope/version substitutions and every remember-policy field are not individually varied in the new commitment test.
- New hostile-record coverage principally establishes truncation, trailing-byte rejection and fuzz canonicality; deterministic unknown/mixed-profile and noncanonical-time cases would strengthen the advertised parser contract.
- Existing `FuzzAttemptRecord` still assumes every accepted record is at most 4096 bytes, despite the shared decoder accepting bound profiles up to 8192. This predates the new session profile, but its oracle should become profile-aware before claiming comprehensive shared-parser fuzz completion.

These are bounded follow-up coverage opportunities, not contradictory requirements or demonstrated unsafe custody behavior.

### Remaining claim limits

Native authenticated-session SQL admission/consumption, primary-authority comparisons, post-wait expiry/freshness, exactly-once accounting, atomic protected effects and recovery remain unfinished implementation. No real PostgreSQL evidence establishes them here.

Clean-consumer dependency pins, exact coverage/mutation, complete release gates, OTP-unit verification and platform completion remain unproven. This report does not certify those boundaries or publication/deployment.


## Authenticated-session custody development — 2026-10-07

Source baseline5e7206b. Focused core missing API REDec5868 after disposable
workspace setup was corrected (initial334bac was unpublished-pin setup failure,
not behavior evidence). New complete original session constructor/getter and
v3 continuation binding GREEN29c0f1. Final core `go test ./... -race -count=1`
PASSd1bb8f 1.232s after final source/test edits. Legacy framing and pre-auth
regressions remain covered by the full core suite. Native session PostgreSQL
custody/admission/consumption and independent post-implementation reviews are
not yet implemented/run; no commit or unit promotion. Disposable root
/tmp/identity-otp-session.oC0qTK removed after results.

### Retained native report: `/root/otp_session_qa_5e7206b`

Message Type: FINAL_ANSWER
Task name: /root
Sender: /root/otp_session_qa_5e7206b
Payload:
## Native QA planning report

**Role:** independent read-only `qa_test_planner`, `/root/otp_session_qa_5e7206b`  
**Source:** `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`  
**Base/head:** `5e7206b6478fcfc48607f871400a75b68e5f017b`  
**Method:** `$qa-test-planner` read fully; one source-inspection pass. No writes, Go execution, services, or delegation.

### Established contract and current gap

Authenticated-session authorization must remain disjoint from pre-auth authorization. A structural session ID or reconstructed `session.Session` proves neither possession nor current authority. Consuming execution must preserve the original session, original authentication time, complete authority vector and exact submission identity; compare them against active primary state; and finalize OTP only with the protected domain commit.

Current source supports legacy authenticated-session **check custody** (`TestAuthenticatedSessionCheckCustody`), but native `Contributor.original` requires consuming mode and authenticates a complete pre-auth flow. Contributor documentation explicitly declares authenticated-session execution unavailable. Existing native integration coverage exercises pre-auth denial, reservation, takeover, rollback and recovery—not authenticated-session admission/consumption.

`session/postgres.RequireSession` is an existing usable session-owned primary admission boundary. It compares the frozen complete record and active family, locks both, and checks database-time expiry after acquiring locks. Its explicit contract does **not** prove bearer possession, complete external authority or sensitive-operation freshness.

### Smallest sufficient coverage artifact

| Risk / realistic defect | Proving boundary and observable assertion |
|---|---|
| Historical compatibility breaks when complete session custody is added | Core/codec tests retain existing v1 authenticated-session and pre-auth literal framing, existing v2/v3 profiles, and legacy round trips. Legacy ID-only input must not acquire complete-session authority or become natively admitted merely because decoding still succeeds. |
| Tagged union accidentally mixes contexts or allows session signin | Core tests reject zero, unknown/mixed tags, session-ID mismatch with the bound record, tenant/subject mismatch, and session-authenticated signin. Preserve valid pre-auth signin/MFA remember semantics. |
| Copies/restart replace original session or authority | Codec round trip through a fresh instance must retain exact original record, lookup digest, authentication time, complete authority vector and selected scope. Change one field at a time—including each optional dimension’s presence—and require binding/fingerprint divergence or rejection. Caller mutation must not alter retained custody. |
| Knowing a session ID or constructing a valid record bypasses possession | Native primary integration: valid record with absent, malformed, wrong-session or wrong-key credential denies before OTP accounting/reservation. Assert unchanged challenge counters/state and no protected domain/audit/outbox success effects. Possession must authenticate through the session-owned token contract. |
| Primary authority refresh silently replaces stale original expectations | Real PostgreSQL table-driven cases: revoke session/family; advance session/family or directory/user/identifier/credential/authorization authority; change applicable organization/factor/passkey authority. Each denies against the original custody, even if a newly read replacement record would pass. Assert no charge, reservation or domain success. |
| Anonymous metadata accepted as authenticated | Structurally valid anonymous session must be rejected by authenticated-session admission, with no OTP mutation. A verified discriminator alone must not replace authenticating proof/credential validation. |
| Lock waits extend expiry or authentication freshness | Deterministic blocked-lock integration cases cross absolute expiry, idle expiry and required sensitive-operation freshness while waiting on downstream challenge locks. Test immediately before and exactly at boundaries; equality denies. Recheck before protected finalization as well as initial admission. Preserve original authentication time—no retry/reconstruction renewal. |
| New authority dispatch breaks exactly-once denial accounting | Real coordinator + PostgreSQL: wrong code commits one increment and stable aborted result; same command/attempt/fingerprint replay increments zero additional times; changed identity/fingerprint conflicts unchanged. Correct submission through denial path remains unchanged, not consumed. Wrong scope/context never charges another challenge. |
| Successful reservation or rollback consumes outside domain commit | Real native reserve/apply/finalize: correct code reserves once; success commits OTP terminal state, protected mutation, required audit/outbox and command result together. Inject a failure after a provisional prefix and assert the entire protected transaction rolls back. Retryable rollback retains the exact reservation; proven noncommit terminally releases it. |
| Race, takeover or recovery repeats comparison or revives authority | Two different commands for one challenge yield one winner. Same-command replay and generation takeover preserve original custody, reserved time and counters; stale generation cannot apply/finalize. Fresh contributor/codec recovery returns the stable result. Ambiguous commit leaves unresolved custody reserved; timeout, expiry or revoked live session alone must not imply noncommit or release. |
| Durable secret leakage or backward key behavior changes | Codec/persistence tests assert no raw code or raw session credential in command fields, replay/custody rows or result records; bounded parsing; redacted formatting/errors; serialization refusal. Active retained wrapping/digest keys restore historical custody; retired/unknown keys fail closed without revealing credential or code. |

**Reduction:** reuse existing pre-auth cases and session-owner record/admission tests rather than duplicate their full matrices. New integration cases should prove authenticated-session dispatch and its materially different possession, complete-authority and freshness checks. Exercise shared accounting/recovery scenarios once through each authorization variant where dispatch can change the outcome. No browser/provider suite is needed to prove this library boundary.

### Focused verification set

1. Core OTP authorization/continuation and attempt-fingerprint tests.
2. OTP PostgreSQL codec/custody tests, including historical literals and hostile-record regressions.
3. Native coordinator PostgreSQL integration for authenticated-session admission, exactly-once denial, consumption, post-wait expiry/freshness, rollback, takeover and recovery.
4. Existing pre-auth native composition regressions and affected session-owner admission tests.
5. A targeted race run for new concurrent native cases and changed custody paths; direct-consumer compilation for additive public APIs.

Use the repository’s existing integration harness and task-owned disposable caches. Full coverage/mutation and remaining release gates remain completion obligations; they are not additional pre-write approval machinery.

### Disposition

**No planning blocker identified.** The missing authenticated-session native path is implementation work, not a contradictory requirement.

Important implementation constraints: `RequireSession` alone is insufficient; current source has no complete authenticated-session custody, and recovery/noncommit release must not depend on renewed live admission. Sensitive-operation freshness is distinct from general session lifetime freshness; use the owning operation’s explicit requirement rather than inventing a stricter numeric threshold.

**Verification status:** source inspected only; no tests executed and no runtime behavior certified.

### Retained native report: `/root/otp_session_mapping_5e7206b`

Message Type: FINAL_ANSWER
Task name: /root
Sender: /root/otp_session_mapping_5e7206b
Payload:
## Native change-mapping report

**Role:** fresh independent, read-only `change_mapper`.  
**Source:** `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`, branch `feature/identity-platform`.  
**Exact base/head:** `5e7206b6478fcfc48607f871400a75b68e5f017b`; inspected worktree clean.  
**Scope:** original authenticated-session OTP custody and PostgreSQL admission/consumption; preserve existing pre-auth consumption and released APIs. Read governing platform attachment, both OTP goals completely, session goal and relevant typed/transaction contracts. No writes, execution, services or delegation.

### Confirmed gap

Authenticated OTP authorization currently retains **only session ID** (`otp/authorization.go:37–55`). Native `Contributor.original` deliberately rejects every profile except consuming, fully bound pre-auth submissions (`otp/postgres/contributor.go:118`). Consequently, authenticated-session consumption is presently unavailable—not an unsafe permissive implementation.

Reading the current session by ID and adopting it would not resolve this gap: it would substitute current versions, digest, family, authority or lifetime for independently retained original custody.

### Ownership and causal slice

| Boundary | Existing producer/consumer and required preservation |
|---|---|
| Original session | `session.Session` contains exact family/session versions, token digest, complete authority vector, authentication/credential metadata, resolved remember policy and original timestamps (`session/record.go:56`). `NewSession` reconstructs structure; it does **not** authenticate. |
| Authentication provenance | `SessionService.Get` produces opaque `GetResult`; `ValidateAt` rejects fabricated/substituted public fields and enforces the original exclusive validity deadline (`session/service.go:160,189`). A stored `Session` or `Authenticated` flag cannot substitute for this provenance when a flow claims authenticated caller admission. |
| OTP continuation | `NewContinuation` derives the authorization commitment, forbids authenticated-session signin and accepts no remember policy for authenticated non-signin purposes. Its current v1 profile commits only session ID; bound pre-auth uses complete-record v2 (`otp/authorization.go:163`). |
| Submission custody | `AttemptIssuer.PrepareForCommand` → sealed `CommandAttemptRecord` → codec encryption. Original scope feeds keyed code/submission fingerprints and record MAC. Mapping/restoration never generates replacement attempts or grants authority. |
| Native durable participant | Coordinator owns command journal, denial result, reservation generations, protected transaction and recovery classification. Native OTP contributor owns challenge/counters/retained encrypted original custody; session owner owns session/family state. No second replay owner is needed. |
| Protected effect | `CheckOTP`/`FinalizeOTP` recheck original bindings without comparing or charging code again. OTP terminalization, owning mutation, required audit/outbox and command result share the existing coordinator commit. |

Production callers of continuation/mapping currently remain internal custody functions; executable consuming composition is exercised by tests. No production HTTP/login composition was found in this causal slice.

### Required authenticated-session admission

- Preserve the complete original session independently before reservation; bind tenant, subject and selected session ID to continuation and command.
- Derive owner scope keys from that original session—not request-provided keys or current replacement records.
- Compare original directory, user, identifier, credential and authorization epochs through existing primary owners; compare session/family identity, versions and complete record through `sessionpostgres.RequireSession` (`session/postgres/admission.go:49`).
- `RequireSession` checks **only session-owned state**. It does not establish global/user/credential/policy authority or a policy Allow decision.
- Existing `sessionAdmissionBindings` supplies the shared derivation boundary (`session/postgres/identifier_admission.go:188`), but is private and supports a narrower explicit method/kind profile. It rejects organization, factor and passkey bindings. Those variants must remain rejected unless their actual owners are checked.
- Original bounded authentication admission must not be extended by continuation, retry or takeover. Session expiry and any operation-required freshness use original timestamps/policy.

### Cancellation, waits and recovery seams

- Caller/Work deadlines must remain effective; current mapping checks cancellation before/after fixed-local crypto. Pure restoration uses a finite local timeout and creates no authority lease.
- Lock order currently precedes lockout/challenge exclusion with directory/user/credential/authorization and flow admission. Authenticated-session admission must retain coherent ordering.
- Re-sample primary time **after downstream waits**, not merely after session admission. Recheck original session lifetime and applicable credential expiry immediately before durable counter/reservation/finalization changes.
- Takeover preserves exact original envelope, counters, times and authority; only reservation generation changes.
- Restart recovery/release already restore code/cookie-free original custody from the challenge. Authenticated-session recovery must recognize its own retained variant; it cannot require a pre-auth row or synthesize an original session from current state.
- Expiry/revocation is denial evidence, not noncommit proof. Missing/divergent custody remains unknown; terminal release requires the coordinator’s existing proven-noncommit authority.

### Compatibility and intentional variants

- Preserve ID-only constructors and exact legacy v1 framing.
- Preserve bound-pre-auth v2 framing, authority-envelope v3, existing MAC/fingerprint domains, size limits and retained-key behavior.
- A complete authenticated-session profile requires explicit new framing; old readers reject opt-in new records, while new readers retain old decoding. Do not reinterpret old ID-only records as complete custody.
- Pre-auth remains single-use with separately submitted cookie and session-owned reserve/rebind/finalize/release.
- Authenticated sessions are **not** consumed or turned into pre-auth reservations.
- Successful metered non-consuming check remains a separate pending profile: independently bounded counter/result behavior, no reservation or proof. Adding authenticated consumption must not accidentally enable it.
- Both module changelogs and affected API/source contracts are consumers of this additive public boundary. Current dependency pins predate portions of inspected source; local replacement compilation would not prove clean-consumer compatibility.

**Disposition:** no contradictory requirement prevents this bounded implementation. The concrete missing contract is complete original authenticated-session custody plus its native phase/recovery dispatch. This report does not approve full OTP, login, release or platform completion.

### Retained native report: `/root/otp_consuming_delivery_5e7206b`

## Native delivery audit: APPROVE

**Role:** independent read-only delivery auditor `/root/otp_consuming_delivery_5e7206b`. No Go/service execution, writes, delegation or external mutation.

**Verified boundary:** local development commit `5e7206b6478fcfc48607f871400a75b68e5f017b`, parent `3c4f196e82e4c248d8594870c3d074bac416f65f`, on `feature/identity-platform` in the selected identity worktree.

- Source checkout is clean. Commit contains exactly the eleven intended OTP adapter files.
- Independently calculated SHA-256 identities of all eleven committed blobs match the complete native correctness review, including its final README amendment.
- Inspected retained native change-mapper, QA, initial `REQUEST_CHANGES` regression, reconciled `APPROVE` regression, Go compatibility and complete correctness reports. Their baseline/snapshot scopes and role identities are explicit; the final correctness disposition is `APPROVE`.
- Original transcript confirms reservation RED `88766f`, protected-execution RED `4e431c`, final adapter integration/race PASS `fecd1c` (12.973s), subsequent exit-zero completion `ed0efd`, and owned PostgreSQL cleanup `27e632`.
- The governing attachment authorizes reviewed local development commits with explicitly pending release gates; it prohibits push/publication/deployment.

**Delivery state:** locally committed only. No pushed, hosted-CI, deployed or production-verified claim is supported.

**Remaining boundaries:** SQL domain/audit/outbox seams are not real email/login interoperability. Public legacy missing-custody recovery is unproved; the test establishes private restoration denial only. Authenticated-session and successful metered-check profiles, source pins, clean consumers, exact coverage, mutation, full profiles and release gates remain pending. Counts remain **5/67 integrated, 1/67 verified**.

Control ledger/evidence changes remain uncommitted. The review workspace roots are absent; message-only `/tmp/identity-otp-commit.q2mTkA` remains for the forthcoming control commit and requires cleanup afterward. These do not invalidate the scoped source-commit approval or establish whole-goal completion.

## Native pre-auth consuming development batch — 2026-10-07

Source baseline `3c4f196e82e4c248d8594870c3d074bac416f65f`; eleven-file dirty
snapshot. Native role reports below are exact parent-transcript text, not
implementation-agent attestations. Final race integration PASS `fecd1c`, 12.973s;
subsequent integration vet exited 0 `ed0efd`. No package/release promotion.

### Retained native report: `/root/otp_consuming_qa_3c4f196`

Message Type: FINAL_ANSWER
Task name: /root
Sender: /root/otp_consuming_qa_3c4f196
Payload:
## Native QA report

**Role:** fresh read-only `qa_test_planner`; `$qa-test-planner` read completely.  
**Source:** `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`, clean HEAD/base `3c4f196e82e4c248d8594870c3d074bac416f65f`.  
**Authority:** complete supplied platform goal, both active OTP goals, transaction contract.  
**Disposition:** no contradictory requirement or unsafe requested behavior blocks implementation.

### Existing evidence boundaries

- `otp/postgres/Contributor` currently supports consuming **pre-auth counted-denial only**. Correct submissions return unchanged/rollback. Native reservation, protected execution and consuming recovery are explicitly unavailable.
- Native `ownership_integration_test.go` proves private SQL reservation/finalization/rebind/release effects, not coordinator admission or noncommit authority.
- Coordinator reservation/execution integration tests use test SQL OTP participants. They exercise coordinator behavior but do not establish native OTP composition.
- Native denial tests already exercise replay, exhaustion, original pre-auth admission, directory/user/identifier/credential/authorization epochs and lock-wait expiry. Preserve these behaviors.

### Bounded coverage artifact

| Risk / realistic defect | Smallest proving test and observable assertion |
|---|---|
| Correct code accepted without full native admission, or charged twice | Real PostgreSQL composition with production OTP contributor and coordinator: reserve correct submission; assert issued → reserved, exactly one used attempt, unchanged failed/check counters, exact original command/attempt/fingerprint linkage. Same-command replay leaves every field unchanged. |
| Cross-scope attempt spends another challenge | Change tenant, purpose, subject/channel, challenge, attempt or fingerprint independently; assert redacted denial/conflict and unchanged challenge/counters/domain state. Retain existing wrong-code denial replay and exhaustion cases. |
| Concurrent winners or mixed correct/incorrect lost updates | Independent connections submit two correct commands and mixed correct/incorrect commands against one challenge. Assert one reserved winner, serialized bounded counters, stable loser classification, no duplicate domain/session effects. |
| Partial takeover or stale generation remains authoritative | Expire command lease; exact successor takes over OTP **and original pre-auth/other enlisted one-time participants** atomically. Corrupt one predecessor generation or make a participant terminal; assert complete rollback, including command generation. Successful takeover preserves original timestamps, expiry, authority epochs, remember policy and attempt count; stale owner cannot apply/finalize. |
| Current checks adopt replacement authority or reuse pre-lock validity | After reserve, change relevant directory/user/identifier/credential/authorization epoch, disable authority, or expire OTP/pre-auth while blocked behind a lock. Protected apply/finalize must deny; domain/session/outbox success effects remain absent. Include a credential-free user and a credential-bound user. |
| Finalization and owning mutation split commits | Successful protected command: finalized OTP, finalized original flow, domain/session effect, command result and required audit/outbox commit together. Inject owning mutation/finalizer/audit/outbox failure; assert no successful prefix persists. Retryable rollback retains reserved custody when that route is selected. |
| Timeout/expiry mistaken for noncommit, or released code resurrects | Unknown commit/process loss leaves pending command and reserved participants. Expiry alone cannot release. Proven rollback/unanimous primary noncommit permits only terminal released plus stable aborted result; retries require a new OTP. Release failure retains unknown custody rather than claiming clean abort. |
| Restart recovery requires original code/cookie or silently refreshes authority | Reconstruct coordinator and native contributor from durable configuration, discarding original command/code/cookie objects. Recover through authorized public recovery with original persisted binding. Assert stable committed replay or proven-noncommit terminal release; missing/divergent participant evidence remains unknown. No additional code comparison, counter charge, workflow/session issuance, renewed expiry, refreshed original epochs or upgraded remember policy. |
| New phase access leaks or outlives custody | Focused unit tests: zero/foreign/returned handles deny; phase confusion denies; recovery/release cannot obtain submitted code/cookie; diagnostics/serialization remain redacted. Retained-key recovery succeeds; missing required original key fails closed. |

**First red regression:** extend the existing native contributor composition to reserve a correct consuming pre-auth command through the coordinator and complete protected finalization. Assert persisted OTP/flow/domain/result state, not interface registration. At this head the native reservation contract is missing; that should be the explicit expected failure.

### Supported variants and full scope

- This batch may implement consuming **pre-auth** continuation without claiming authenticated-session or successful non-consuming check support.
- Add explicit denial regressions for those currently unavailable native profiles; do not reinterpret them as implemented. Core tagged authorization variants remain disjoint.
- Successful check must ultimately have separately bounded attempt policy and no reservation/proof. Authenticated-session consuming workflows must ultimately preserve session-owned original authority/freshness. Both remain product obligations, not unrelated first-write gates.
- Full issuance/resend, purpose profiles, keyed digest rotation, encrypted delivery, scanner-safe wrappers, cleanup/tombstones, migrations and owning workflow/session issuance remain required for OTP-unit completion. This reservation batch does not prove them.

### Smallest verification set

1. Focused new native consuming integration cases above plus existing native denial composition.
2. Existing native ownership/capacity/pre-auth lock-expiry tests and coordinator OTP reservation/execution/recovery/phase-custody tests.
3. Affected OTP core unit tests and native adapter/core direct-consumer compatibility checks.
4. Race execution for changed concurrent native composition; controlled disconnect/restart tests for the ambiguity claim.

Exact coverage/mutation, broader fuzz, cleanup/migration, supply-chain and other affected release gates remain necessary **before final unit verification**, not a new prewrite approval program.

**Execution status:** read-only source/test inspection only; no tests, Go commands, services, resources or writes. No current pass claim.

---

### Retained native report: `/root/otp_consuming_map_3c4f196`

Message Type: FINAL_ANSWER
Task name: /root
Sender: /root/otp_consuming_map_3c4f196
Payload:
## Native change-mapper report

**Role:** fresh read-only `change_mapper`  
**Reviewed base/head:** `3c4f196e82e4c248d8594870c3d074bac416f65f`  
**Workspace:** `feature/identity-platform`, clean worktree at `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`.  
Source inspection only; no writes, Go execution, services, resources, or delegation.

### Existing execution boundary

The native `otp/postgres.Contributor` currently implements counted denial only. Correct submissions return `OTPAttemptUnchanged`; consuming reservation, protected execution, and conclusive participant recovery are explicitly unavailable (`otp/postgres/contributor.go:38`, `:199`, `:224`).

The coordinator already supplies the consuming phase infrastructure:

- `PrepareOTPReservationPlan` / `ReserveOTPCommand`: exact consuming owner, command replay, coordinated reservation or predecessor-generation transfer (`postgres/otp_reservation.go:144`, `:170`, `:181`).
- `Begin`: requires the full `OTPExecutionContributor`, not reservation alone (`postgres/domain.go:47`).
- `ApplyFeature`: invokes OTP current checks before domain execution (`postgres/feature_execution.go:242`).
- `Commit`: repeats current checks and invokes OTP finalization in the protected transaction (`postgres/domain.go:326`).
- Proven rollback: invokes OTP release in the existing abort-terminalization transaction (`postgres/domain.go:474`).
- Restart reconciliation: reconstructs the stored plan, collects every contributor’s primary noncommit evidence, then invokes OTP release and terminalizes the journal (`postgres/recovery.go:156`, `:237`, `:283`).

Ordinary `PrepareCommandPlan` remains closed to OTP/opaque-code submissions (`postgres/coordinator.go:175`). Enabling the native contributor does not justify opening that generic route.

### Persisted custody channels

| Existing channel | Retained information | Missing original custody |
|---|---|---|
| `identity.commands.command_plan` | Frozen contributors, feature schema, audits, lifecycle, OTP’s eight scalar attempt fields | Feature command fields, encrypted OTP envelope, complete original flow and authority |
| Global claim / journal scope | Exact command binding, feature operation, fingerprint and key version, generation and outcome | Full OTP continuation and original authority |
| OTP challenge row | Issuance linkage, keyed scope/code digests, immutable policy/times, counters, consuming command/fingerprint/attempt/generation | Recoverable original preauth projection and authority expectations |
| Session preauth row | Canonical original flow record, cookie commitment, intent, remember policy, immutable timeline/version, consuming command/fingerprint/generation | Independent proof that a newly read record was the original OTP command’s selected record |

Anchors: `postgres/plan_feature.go:127`, `postgres/plan_otp.go:37`, `postgres/recovery.go:307`, `otp/postgres/migrations.go:39`, `session/postgres/preauth_migration.go:39`.

**Central consumed gap:** the original code-free OTP envelope exists only in `FeatureCommandSpec.Fields["otp.attempt-custody"]`, not the persisted plan. That envelope already contains the full original flow and authority; restart recovery currently cannot obtain it through existing sealed handles.

`RecoveryWork.OTPBinding` and `OTPReleaseWork.Binding` expose correlation only. Release gets neither original `CommandSpec` nor cookie (`postgres/recovery_participant.go:184`, `postgres/otp_execution.go:231`, `:307`). Reading the current session record by ID cannot substitute for independently retained original command custody.

### Already available original projection

The envelope formats distinguish:

- v1: legacy flow/session ID projection.
- v2: full bound original preauth transaction.
- v3: original authority expectations wrapping the attempt projection.

The retained projection is **code-free**: attempt ID/fingerprint, scoped continuation, mode/version, remember policy, full original preauth record, signature, and original directory/user/authorization/credential epochs. Raw code and cookie remain separately borrowed opaque command inputs (`otp/postgres/command_record.go:15`, `otp/postgres/preauth_record.go:14`, `otp/postgres/authority_custody.go:22`).

Its authenticated restoration depends on exact command binding, operation, digest-key version, retained envelope/attempt/code keys, and core signature (`otp/postgres/command.go:275`). Persisting this existing authenticated original projection through the existing plan or participant storage would close the custody transport gap without another journal or process map.

### Minimal consuming gaps and required ordering

1. **Durably retain and recover original code-free custody.** Bind it to the frozen OTP plan, command fingerprint and consuming owner. Make it available to proof-bound rollback/recovery release without submitted code/cookie. Missing, corrupt, foreign, or divergent custody must fail closed.
2. **Native reserve:** restore and validate original command/authority; enforce current directory and applicable user, credential and authorization epochs; authenticate actual cookie against original commitment; lock original issued preauth flow; enforce ordered shared lockouts; lock challenge; resample primary time; compare once for reservation; persist OTP and preauth reservation in the coordinator’s single transaction.
3. **Native takeover:** validate original linkage, CAS both OTP and preauth exact predecessor generations atomically, preserve counters and original times, and perform no new comparison/charge. Existing operations: `rebindReservation` (`otp/postgres/ownership.go:134`) and `sessionpg.RebindPreAuth` (`session/postgres/preauth_reserved.go:102`).
4. **Current check/finalize:** use code-free original envelope plus actual borrowed cookie, exact reserved ownership, current original authority and downstream-wait expiry checks. Finalize OTP and preauth in the protected mutation/result transaction. Existing `OTPApplyWork` / `OTPFinalizeWork` expose `CommandSpec`, `PreAuthCookie`, authority helpers and bounded SQL.
5. **Recovery evidence and release:** inspect exact predecessor ownership and retained original projection under primary exclusion. Only conclusive retained negative evidence may return `RecoveryNotCommitted`; missing/divergent/terminal evidence remains unknown. Release both OTP and original flow only via live `OTPReleaseWork` after proven rollback or unanimous noncommit. Expired credentials/flows do not prevent authorized terminal release; they never authorize it.

Existing OTP persistence helpers already implement reservation, exact ownership loading, takeover, terminal CAS and expiry-independent release (`otp/postgres/ownership.go:54`, `:89`, `:147`). Session equivalents already exist, including proof-bound `ReleasePreAuth` (`session/postgres/preauth.go:118`, `:127`, `session/postgres/preauth_release.go:28`).

### Compatibility and variant traps

- `OTPAttemptPlan` and reservation bindings are currently comparable. Production code uses exact equality in custody validation, challenge ownership and denial recovery (`otp/postgres/command.go:358`, `preauth_admission.go:87`, `ownership.go:55`, `postgres/otp_recovery.go:34`). Adding slice/map custody breaks that contract; pointer identity would not provide content equality.
- OTP encoding is an exact eight-field deterministic CBOR map; feature decoding reconstructs and canonically validates historical profiles. Preserve existing bytes and denial recovery for legacy plans while making incomplete historical custody inadmissible for newly activated consuming recovery (`postgres/plan_otp.go:92`, `postgres/plan_feature_decode.go:130`).
- Challenge trigger forbids deletion, immutable-field mutation and terminal reactivation. Storage changes require forward migration rather than rewriting installed migration history (`otp/postgres/migrations.go:151`). Session terminal release separately requires `PreAuthReleaseMigrationSource`.
- `Contributor.original` restores raw-code custody and requires actual cookie. It cannot be reused unchanged by code-free protected handles or cookie-free recovery (`otp/postgres/contributor.go:109`).
- Authenticated-session authorization and successful non-consuming checks remain separately required contracts. Current contributor intentionally rejects both; activation must not reinterpret ID-only session custody as bound preauth or let a successful check produce consuming proof (`otp/postgres/contributor.go:100`).
- Full module completion still includes issue/resend, authenticated-session admission, successful-check durable results, owning workflow/session effects and retention/cleanup; the preauth consuming activation is a bounded prerequisite slice, not their implementation.

**Original-custody retention authority:** journal/plan or OTP participant persistence must retain the authenticated original projection through unresolved reservations and required terminal recovery. Session primary storage remains its state owner; retained projection is an immutable expectation, not a second flow authority. Rollback uses the unchanged generation; recovery release uses the exact predecessor and coordinator successor. Unknown outcomes retain reserved custody until authoritative reconciliation.

---

### Retained native report: `/root/otp_consuming_map_3c4f196`

Message Type: FINAL_ANSWER
Task name: /root
Sender: /root/otp_consuming_map_3c4f196
Payload:
**Confirmed cause:** `ReserveOTPCommand` rejects `identity.session.issue` through `authenticatedOperation(binding)` before transaction creation or native OTP invocation (`postgres/otp_reservation.go:170`, `postgres/authenticated_command.go:74`). Successful preparation does not override that execution guard.

**No production change is needed for this failure.** Keep the guard: session issuance requires its separate authenticated-proof admission/current-check protocol (`postgres/authenticated_command.go:55`, `:85`).

Using an `identity.email.verify` owned-feature fixture can exercise OTP reservation and protected domain atomicity without bypassing session issuance. Bind the fixture’s operation, original preauth action, OTP purpose, and continuation consistently; an email-verification test should not retain signin-purpose/remember semantics.

That fixture proves the coordinator/native OTP consuming boundary, not production email workflow or OTP signin/session issuance. The eventual signin bridge remains separate owning-workflow implementation.

---

### Retained native report: `/root/otp_consuming_regression_3c4f196`

Message Type: FINAL_ANSWER
Task name: /root
Sender: /root/otp_consuming_regression_3c4f196
Payload:
## Native regression review: REQUEST_CHANGES

**Reviewed:** baseline/current committed HEAD `3c4f196e82e4c248d8594870c3d074bac416f65f`, plus the complete dirty batch:
- `otp/postgres/{contributor,ownership,preauth_admission,contributor_integration_test}.go`
- Untracked `otp/postgres/{consuming,consuming_execution,consuming_migration}.go`
- Adapter README and CHANGELOG.

Read the authoritative platform goal, OTP/core-adapter goals, and related kernel reservation/execution/authority and session pre-auth contracts/tests. This was read-only; supplied test results were not independently rerun.

**No confirmed runtime defect found. Two concrete batch-level evidence gaps remain:**

### 1. Native consuming admission lacks negative post-wait/post-reservation coverage

**Anchors:** `contributor_integration_test.go:246`, `consuming.go:90`, `consuming_execution.go:80`.

All six native modes use valid anonymous authority. Existing stale-user, credential-expiry and flow-lock-expiry modes reach the **denial** path after the native branch returns. Consequently, they do not exercise the newly introduced `consumingTime` or protected-phase `checkConsuming` rejection paths.

**Counterfactual:** removing the new post-lock flow/credential validity checks, or omitting protected-phase original-authority admission, would leave the six native tests passing. A correct code could then reserve or finalize after original eligibility expired or changed.

**Minimal followup:** add focused real-PostgreSQL native cases proving:
- Reservation waits on downstream challenge exclusion until the original flow expires, then leaves both OTP/pre-auth issued and counters unchanged.
- After reservation, changing the original user/directory security epoch prevents protected application/finalization and produces no committed domain/outbox/success audit.
- A credential-bound protected operation waiting downstream past credential expiry is rejected using fresh database time.

Reuse the existing lock-wait observation and authority fixtures; no expanded matrix or new machinery is required.

### 2. The new migration’s populated-state and immutable-custody contract is untested

**Anchors:** `contributor_integration_test.go:252`, `consuming_migration.go:47`.

The integration test installs `ConsumingMigrationSource` before inserting any challenges. Existing legacy ownership tests do not install this extension. Replay assertions demonstrate that application code preserves the envelope; they do not demonstrate the migration trigger rejects replacement/erasure.

**Counterfactual:** a migration incompatible with populated legacy reservations, or a weakened/absent immutable-custody trigger, could pass all current native tests.

**Minimal followup:** one populated-state migration test:
- Seed an issued challenge and a legacy reserved challenge before installation; apply the extension and assert original state/counters/ownership remain unchanged.
- Verify a legacy reservation without original custody stays fail-closed under native recovery without being released.
- Create native custody and assert SQL replacement/erasure is rejected after reservation and terminalization.

## Evidence boundaries

The supplied focused native pass and full adapter integration/race pass support happy-path reservation, replay/takeover, consumption, rollback, restart recovery and divergent-predecessor unknown retention. Existing kernel tests already cover returned/foreign phase handles and generic failure-prefix transaction rollback; those do not need duplicate test machinery.

Domain/audit/outbox are explicitly SQL seams, not platform email/login interoperability proof. Pending release gates and unfinished OTP profiles are outside this batch disposition and are **not** the reason for `REQUEST_CHANGES`.

---

### Retained native report: `/root/otp_consuming_regression_3c4f196`

Message Type: FINAL_ANSWER
Task name: /root
Sender: /root/otp_consuming_regression_3c4f196
Payload:
## Native regression reconciliation: APPROVE

**Reviewed head:** `3c4f196e82e4c248d8594870c3d074bac416f65f`, plus the complete current dirty batch previously listed and new:
- `otp/postgres/consuming_migration_integration_test.go`
- `otp/postgres/consuming_wait_integration_test.go`

**Both prior findings are resolved for this batch.**

- Native reservation now observes an actual downstream challenge lock, crosses original flow expiry, and asserts neither reservation nor counter charge survives.
- Protected native application now rejects a changed original directory epoch and expired credential eligibility after downstream waiting. Assertions distinguish rejection plus proven release from committed domain/outbox effects. The longer credential-test lease appropriately isolates credential expiry from ownership loss.
- Populated-state migration tests compare all original row fields before/after installation for issued and legacy-reserved rows and require new custody columns to remain null.
- SQLSTATE `23514` assertions exercise replacement and erasure of original custody while reserved and finalized, directly testing the database invariant.
- Wait helper cancellation, exclusion release and worker joining are bounded and owned; I found no new test-oracle defect requiring changes.

**Claim boundaries:** public command-free/code-free/cookie-free restart recovery remains established by the separate native restart tests. The legacy missing-custody assertion exercises only private restoration failure; its partially constructed contributor does not certify public legacy recovery or isolate every possible rejection reason. Do not broaden that claim.

No unresolved concrete regression finding remains. Supplied focused passes were not independently rerun. Final full adapter integration/race and vet results remain pending and must be assessed separately; this approval does not certify release gates, complete OTP profiles, or platform email/login interoperability.

---

### Retained native report: `/root/otp_consuming_compatibility_3c4f196`

Message Type: FINAL_ANSWER
Task name: /root
Sender: /root/otp_consuming_compatibility_3c4f196
Payload:
**APPROVE — bounded local-development compatibility snapshot.**

Reviewed baseline `3c4f196e82e4c248d8594870c3d074bac416f65f` plus all 11 dirty `otp/postgres` files identified in the assignment; HEAD and scope remained unchanged at final inspection. Read-only source audit; no commands executed Go code or contacted services.

- Public changes are additive: `ConsumingMigrationSource` and five contributor methods implement the existing coordinator-owned reservation/execution interfaces. Existing constructor/configuration signatures and contributor descriptor remain unchanged.
- The canonical eight-field `OTPAttemptPlan` and CBOR encoding/reader are unchanged. Original v1/v2/v3 custody implementations are unchanged; retained custody restores through existing authenticated decoding rather than inventing historical authority.
- Existing OTP/lockout migration sources are unchanged, preserving their SQL/checksum inputs. The new separately versioned forward migration leaves historical rows intact with nullable custody, authenticates new originals, and prevents replacement or erasure after reservation.
- Deployment prerequisites are documented: install OTP consuming and session pre-auth release extensions before enabling consumption; deploy v3-capable readers before authority-aware writers. Historical missing custody fails closed.
- New calls remain adapter → kernel/core/session adapter; no reverse dependency, new engine, permanent `replace`, module declaration change, or Go toolchain downgrade was introduced. Affected modules retain Go `1.27.0`.
- Existing denial callers retain the separate denial entry point: correct codes still roll back there. Reservation/protected execution is explicitly selected through existing coordinator APIs.

**No new material compatibility finding.** Minor documentation inconsistency: README’s authority-custody paragraph still says “remaining phases … are pending,” although its opening section now documents implemented consuming phases; this does not alter the documented incomplete-login boundary.

**Unverified boundaries:** clean external consumption/source-pin repair, publication/SemVer, full release gates, exact coverage/mutation, provider interoperability, authenticated-session and metered-check profiles, and complete login/platform acceptance. Existing source-pin/clean-consumer debt is not repaired by disposable `GOWORK` checks and is not a new regression from this batch. Parent-reported integration race passed; pending standalone vet is **not certified as passed** by this audit.

---

### Retained native report: `/root/otp_consuming_final_review_3c4f196`

Message Type: FINAL_ANSWER
Task name: /root
Sender: /root/otp_consuming_final_review_3c4f196
Payload:
## Native correctness review

**Disposition: APPROVE — scoped local development batch. No findings.**

**Boundary:** `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`, branch `feature/identity-platform`, committed base/head `3c4f196e82e4c248d8594870c3d074bac416f65f` plus the complete eleven-file dirty snapshot below. This includes the final README amendment.

Used `$requesting-code-review`, including its independent, adaptive, bounded-closure and advanced-boundary contracts; applied `$review-postgresql-routines`, `$testing-integration` and `$testing-coverage` directly. Single-reviewer structured pass; no delegation, writes, services, Go execution or external mutation.

### Source-backed closure

- **Reservation and takeover:** traced native coordinator dispatch through original command/code/cookie authentication, primary directory/user/credential/authorization checks, pre-auth exclusion, ordered shared lockouts and challenge exclusion. Initial reservation charges once and attaches original custody atomically. Takeover requires exact predecessor ownership, retains original envelope/timestamps and transfers both reservations within the existing coordinator transaction.
- **Protected execution:** traced `ApplyFeature()` and `Commit()` into native checks/finalization. They authenticate separately borrowed cookie custody, retain exact original authority, recheck time-dependent eligibility after downstream waits and finalize OTP/pre-auth alongside provisional domain, audit, lifecycle and journal effects.
- **Recovery and release:** retained encrypted original custody is independently authenticated and its derived plan must equal the journal plan. Recovery compares, rather than adopts, current pre-auth state; missing or divergent predecessor evidence cannot authorize release. The existing sealed noncommit phase owns release and checks exact pre-auth/OTP predecessor custody.
- **Migration:** new forward migration leaves historical rows nullable without invented authority. Its total all-null/all-present constraint bounds custody and metadata; its trigger permits initial attachment only during `issued → reserved`, rejecting replacement/removal thereafter. Existing challenge lifecycle, FK, counter and terminal-retention enforcement remains intact. Installation has explicit lock/statement timeouts and no destructive rollback.
- **Adversarial challenges:** traced competing commands against the same pre-auth/challenge, stale takeover ownership, credential/pre-auth expiry during challenge waits, divergent recovery generation, nullable historical custody and original-envelope mutation. Exclusion/CAS, fresh expiry checks, authenticated retained evidence or fail-closed handling prevents the claimed unsafe outcomes.
- **Oracles:** new tests observe persisted counters, generations, immutable envelopes/timestamps, final states and independent SQL domain/audit/lifecycle outcomes. Failure tests inspect rollback/release conservation. Restart recovery discards command/reservation custody before invoking recovery. The legacy nullable-custody test proves only private restoration denial.

### Verification and claim limits

Supplied evidence: reservation RED `88766f`; protected-execution RED `4e431c`; `go test -race -tags integration -count=1 .` PASS `fecd1c` (12.973s); subsequent `go vet -tags integration .` exit 0 `ed0efd`. Reviewer independently ran `git diff --check`: passed. I did not rerun Go or PostgreSQL.

Approval supports this reviewed **local development commit**, not package verification, release readiness, public login, real email/session workflow completion or audit/outbox interoperability. Source-pin/clean-consumer, exact coverage/mutation, authenticated-session, metered-check, complete-login and release gates remain pending. PostgreSQL deployment privileges, production-scale migration duration, replication and production execution are not certified.

### Reviewed file SHA-256 identities

```text
0ddef16c6ba25f696afc1b58b46dc5ddf6d48870c4562a8039ebe5ca0bf3bc12  otp/postgres/CHANGELOG.md
71705949d46da3e4c21cbda78a1509c6af9ac028640ce29ffde4ebe38335b2dd  otp/postgres/README.md
f381877e33ef25f2b37208890e6d3524f3f1b0317869b6cd27f7da93dcf01776  otp/postgres/contributor.go
1bdf630f078f590ccff76f722cb60b09283826231aaf81839d6729593fb1e97a  otp/postgres/contributor_integration_test.go
d6aa59c1fa9a8912658effaaee04e1d6a6ee4e906f213b40159832b842b15af5  otp/postgres/ownership.go
77407fa51e8dbd6dd1269a220e9f95ac06485ca2936a65da022e359666286d43  otp/postgres/preauth_admission.go
cb486568db7dabb3398879f13fa8a3f0e2dd5bfea16a9ddbddefc7bc169dfd7e  otp/postgres/consuming.go
8dd3882cd0fa53442c402f7131a9325ec0e42a2616c2061d3d04e3e4432b53b2  otp/postgres/consuming_execution.go
b4bfbd90e59954833aa6f1a8bfbc0ce01c4edc2738e6d9fb2ddd0ea768f3a47b  otp/postgres/consuming_migration.go
7edb8459c17e1b8dc0cc4bd3ccbad8babd27e37d8b81d9fb3938c564def01744  otp/postgres/consuming_migration_integration_test.go
b41c9c71d2ecd7dd3beba239612932305dfc6877abe4770ba7db779a8c20e2a7  otp/postgres/consuming_wait_integration_test.go
```

## Native counted-denial checkpoint — 2026-10-07

### Native exact-head delivery audit

**Native delivery auditor:** `/root/otp_contributor_delivery_3c4f196`
**Disposition:** `APPROVE` — local counted-denial development checkpoint only.

- Clean source at `3c4f196e82e4c248d8594870c3d074bac416f65f`, parent `b523b3feb87a0428def6dcd75a692decb1a8079d`.
- All five committed file hashes match native final regression, compatibility and complete correctness approvals.
- Authenticated native mapping/QA reports inspected. Incorrect QA extraction was corrected and now matches the original OTP planner report.
- Native execution outputs inspected: expiry RED `d6207b`; focused GREEN `81385b`; full adapter integration/race `9ae2fb` passed 9.393s; standalone vet `fc135b` exited 0. Cleanup outputs `bb0ae7` and `4bd8b8` succeeded.
- Locally committed; no push, hosted CI, deployment or production-verification claim.

Successful OTP/check/reservation/recovery/issuance/login, clean consumers/source locks, exact coverage/mutation and release gates remain pending. Counts remain **5/67 integrated, 1/67 verified**. No Go/service execution, writes, remote operations or delegates performed by this auditor.

Source `3c4f196e82e4c248d8594870c3d074bac416f65f`, parent
`b523b3feb87a0428def6dcd75a692decb1a8079d`; five-file reviewed batch.
Native mapping/QA reports follow below. Original missing-API RED581253.
Real coordinator tests use production OTP/cookie/current-authority/counter code;
audit persistence and the unexecuted feature/lifecycle remain documented seams.
User-bound positive controls and independent user/security/authorization and
credential drift were added after regression review identified that anonymous
fixtures alone could not detect omitted native dispatch.

Final review found flow/credential expiry crossing downstream challenge waits.
Corrected reproducer REDd6207b showed used1/failed1/version2/audits1 for both
expiry cases when zero changes were required. Production now retains original
admitted flow and rechecks temporal validity after waits; optional credential
eligibility is rechecked through its existing owner on already-held share locks.
Focused race GREEN81385b 3.487s; final full adapter
`go test -race -tags integration -count=1 .` PASS9ae2fb 9.393s;
standalone `go vet -tags integration .` exit0 fc135b;
final diff/hash check c67019. Owned PostgreSQL containers auto-removed699491 /
bb0ae7, caches/workspace removed4bd8b8. Native final reports below.
Full OTP/check/reservation/recovery/issuance/login, source locks, clean consumers,
coverage/mutation and remaining release gates are pending. No unit promotion.

### Native regression final report

**Native regression reviewer:** `/root/otp_contributor_regression_b523`
**Fresh disposition:** `APPROVE` — no unresolved findings. This supersedes approval of the earlier, expiry-defective snapshot.

The final tests exercise actual challenge-lock waits crossing both flow and credential expiry, then assert no charges, audit, lockout extension or aborted receipt. Production rechecks the original admitted flow and exact credential at post-wait primary time before comparison. Existing user/credential drift coverage remains effective; cancellation and goroutine joining cover failure cleanup.

**Reviewed base/HEAD:** `b523b3feb87a0428def6dcd75a692decb1a8079d`

```text
contributor.go                  433d11c1ab1a3991dce5af396db378a81fab7b53a86d978b3b2f2c9e4eefa275
contributor_test.go             3a7ce40b632ab768d49842b7a3a9214e8febd0ae488ee8ebe0e8fe09dfcbbcb5
contributor_integration_test.go ea8abc9250482f580643ae8719dd0c756a77c9aee3dcbe16d4ae2b0655d9353b
README.md                      6f5952c506a8d27f2089af299e525ffd96f4dbea3e93d11eead63a08cc04b0a6
CHANGELOG.md                   81f0766ea2b1d60a1e5ed6bc2ff94ab5d338062cc45d7bfd780e8f3ab8bb6942
```

Read-only audit; focused RED/GREEN and vet results were supplied by the parent, not independently executed here. Final full race result remains pending. Approval covers counted-denial development only, not successful OTP workflows, clean-consumer compatibility or release verification.

### Native complete final review

**Native role/principal:** independent complete-final-diff correctness reviewer, `/root/otp_contributor_final_b523`
**Disposition:** `APPROVE` for this local counted-denial development batch. No unresolved findings.

Re-reviewed all five dirty files against base/current committed HEAD `b523b3feb87a0428def6dcd75a692decb1a8079d`, including complete production code, tests, documentation, and affected custody/admission/coordinator/persistence/recovery contracts—not merely the temporal amendment.

The prior temporal finding is resolved: the admitted original flow is retained and checked against post-lock primary time before comparison; optional credential eligibility is rechecked through its existing owner using the same frozen expectations and already-held share locks. No replacement authority, expiry renewal, new transaction owner, or lock-order inversion is introduced.

The two new real-coordinator regressions reach an observed challenge lock wait after initial admission, release after primary expiry, and require unchanged counters/version, challenge state, lockouts, audit rows, and denial journal. Worker cancellation/join and transaction cleanup are bounded. Supplied RED `d6207b` establishes detection of the prior defect; GREEN `81385b` and final adapter integration/race PASS `9ae2fb` establish the corrected cases under the supplied execution environment.

Complete-pass closure found no additional defect in configuration ownership/redaction, authenticated original custody, primary authority admission, scoped lock ordering, counter/exhaustion persistence, rollback, terminal replay, or unknown recovery. Correct codes still roll back without authentication; documentation preserves that limitation.

**Reviewed SHA-256 hashes, unchanged at close:**
```text
433d11c1ab1a3991dce5af396db378a81fab7b53a86d978b3b2f2c9e4eefa275 contributor.go
3a7ce40b632ab768d49842b7a3a9214e8febd0ae488ee8ebe0e8fe09dfcbbcb5 contributor_test.go
ea8abc9250482f580643ae8719dd0c756a77c9aee3dcbe16d4ae2b0655d9353b contributor_integration_test.go
6f5952c506a8d27f2089af299e525ffd96f4dbea3e93d11eead63a08cc04b0a6 README.md
81f0766ea2b1d60a1e5ed6bc2ff94ab5d338062cc45d7bfd780e8f3ab8bb6942 CHANGELOG.md
```

Verification considered: supplied final `go test -race -tags integration -count=1 .` PASS 9.393s and `go vet -tags integration .` exit 0; independently inspected `git diff --check` passed. No Go/service execution, edits, or external mutations performed by this reviewer.

Approval does **not** certify successful consuming/check workflows, reservation/finalization, historical participant recovery, issuance/login, real audit/outbox interoperability, exact coverage/mutation, source-lock/clean-consumer, release qualification, hosted CI, publication, or deployment. Those remain pending under the authoritative full goal.

### Native Go compatibility final report

**Native identity:** `/root/otp_contributor_compat_b523`
**Disposition:** **APPROVE** for current local-development compatibility; no concrete compatibility findings.

**Reviewed base/head:** `b523b3feb87a0428def6dcd75a692decb1a8079d`
**Five-file SHA-256 snapshot:**
- `contributor.go`: `433d11c1ab1a3991dce5af396db378a81fab7b53a86d978b3b2f2c9e4eefa275`
- `contributor_test.go`: `3a7ce40b632ab768d49842b7a3a9214e8febd0ae488ee8ebe0e8fe09dfcbbcb5`
- `contributor_integration_test.go`: `ea8abc9250482f580643ae8719dd0c756a77c9aee3dcbe16d4ae2b0655d9353b`
- `README.md`: `6f5952c506a8d27f2089af299e525ffd96f4dbea3e93d11eead63a08cc04b0a6`
- `CHANGELOG.md`: `81f0766ea2b1d60a1e5ed6bc2ff94ab5d338062cc45d7bfd780e8f3ab8bb6942`

The expiry correction reuses existing session `PreAuthTransaction.Check` and exact-owner kernel credential admission contracts. It introduces no public signature, interface, dependency or Go-version change. Credential rereading uses the same already share-locked rows and original expectations; it does not adopt fresh authority or create another transaction owner. Lock-expiry regressions observe actual challenge-lock waiting and provide cancellation/join cleanup.

**Preserved limitations:** stale source pins, clean-consumer qualification and release gates remain pending; workspace-local proof is not release evidence. Successful OTP, reservation/finalization, historical recovery, login and unit completion remain unverified. This refresh inspected source only; supplied red/green/vet summaries and the running aggregate race result were not independently inspected.

## Native contributor pre-write reports — 2026-10-07

### Mapping: /root/otp_native_contributor_mapping_b523b3f

**Read-only change_mapper report — exact head `b523b3feb87a0428def6dcd75a692decb1a8079d`.** No writes, tests, services, or subagents.

### Existing ownership and callbacks

- Coordinator already owns claim/journal replay, transactions, generation fencing, audits, and ambiguity. The production OTP adapter presently supplies **custody and private persistence helpers, not a registered native contributor**.
- Required native interfaces:
  - `ValidateOTPAttemptCommand(identity.Command, OTPAttemptPlan) error`
  - `RecordOTPAttemptDenial(context.Context, OTPAttemptWork) (OTPAttemptDisposition,error)`
  - `ReserveOTP/RebindOTP(context.Context, OTPReservationWork) error`
  - `CheckOTP(context.Context, OTPApplyWork) error`
  - `FinalizeOTP(context.Context, OTPFinalizeWork) error`
  - `ReleaseOTP(context.Context, OTPReleaseWork) error`
  - `RecoverParticipant(context.Context, RecoveryWork) (RecoveryStatus,error)`
- Exact descriptor is `{Owner:"identity/otp/postgres", ContractID:"identity.otp.store", Version:1}`. Work custody is tied to the **exact comparable registered instance**, callback lifetime, plan, command fingerprint and generation. (`postgres/otp_attempt.go:34`, `otp_reservation.go:29`, `otp_execution.go:23`, `recovery_participant.go:34`, `plan_otp.go:59`)

### Custody and authority

- `AttemptCodec` fixes local envelope/keyring crypto and core issuers/verifier; configuration copies retained keys. `restoreAttempt` authenticates command-bound encrypted record, exact declared digest version and plan, then binds opaque code. Raw code is excluded from persisted ciphertext. (`otp/postgres/command.go:66,88,154,275,339`)
- `MapWithAuthority` / `MapPreAuthWithAuthority` freeze original directory and, for user-bearing commands, authorization/user/identifier/credential epochs. Passwordless users still require all three security epochs; credential reference is optional. Anonymous commands prohibit user dimensions. These originals must never be replaced with newly observed epochs. (`authority_custody.go:26,44,75`)
- Phase handles already expose owner-controlled `RequireActiveDirectoryAuthority`, `RequireActiveUserAuthority`, `RequireActiveUserAuthorization`, and optional `RequireActiveCredential`; failures latch into coordinator work. These are current primary owner checks, not independent permission decisions. (`postgres/otp_authority.go:72–257`)
- Pre-auth original record includes cookie digest, intent, remember policy, version, issuance and original expiry. Actual submitted bearer is separate ephemeral command custody; `authenticatePreAuth` derives its digest using the **record’s retained cookie-key version** and compares to the frozen record. Protected apply/finalize expose only `CommandSpec` plus borrowed `PreAuthCookie`, not code. (`preauth_admission.go:58,125`; `postgres/otp_execution.go:131,141,183,190`)
- Session owner already provides `RequirePreAuth`, `ReservePreAuth`, `RequireReservedPreAuth`, `RebindPreAuth`, `FinalizePreAuth`, proof-bound `ReleasePreAuth`. These preserve original timeline/policy; expiration never supplies release authority. (`session/postgres/preauth.go:105,118,127`; `preauth_reserved.go:79,102`; `preauth_release.go:28`)

### Concrete behavior and lock ordering

A coherent native callback must authenticate frozen custody, check directory → effective-user/security → credential if present → user authorization → session/pre-auth authority, then shared lockouts, then challenge; compare only in counted initial attempt/reserve. Existing owners must remain synchronization authorities.

- Shared lockout identities exclude challenge, purpose, flow and code-key version; destination precedes subject. Existing rows must be present, retained and locked, never reset/recreated to bypass budget. Primary time is sampled **after waits**. (`otp/postgres/lockout.go:61,134`)
- `loadChallenge` locks exact keyed scope and requires committed issuance. `Verifier.EvaluateAttempt` proposes shared-total counters; correct checks consume check allowance and total allowance, wrong exhaustion proposes shared lockout. `persistCount` alone does not persist replay/audit/result. (`challenge.go:42,169`; `otp/challenge.go:200`)
- Denial must atomically persist count/expiry/lockout proposal with coordinator’s aborted receipt and audits; matching replay precedes comparison. Reservation must atomically reserve pre-auth and OTP in the same existing coordinator transaction. Takeover transfers both exact predecessor generations without comparison or allowance charge.
- `persistReservation`, `loadReservation`, `rebindReservation`, `terminalReservation` already enforce complete command/attempt fingerprint linkage and full-width version fences. Reserve/rebind retain terminal version capacity. Finalize rechecks current authority and expiry; release requires sealed noncommit invocation and must work after expiry/revocation. (`ownership.go:54,89,134,147`)
- Recovery must conservatively establish retained exact reserved predecessor evidence; missing/divergent/finalized evidence remains unknown. Coordinator requires unanimous participant noncommit before invoking release and terminalizing journal. (`postgres/recovery.go:160–294`)

### Actual consumed-API gaps

1. **Restart recovery/release lacks original flow custody.** `RecoveryWork.OTPBinding` and `OTPReleaseWork.Binding` expose only OTP projection/fingerprint/generations. Neither exposes encrypted original feature custody or pre-auth ID/record. `admittedKernelPlan` stores feature registration, not feature-command fields; canonical commands are explicitly not persisted. Thus simply adding a `CommandSpec` accessor cannot recover absent information. Frozen recovery linkage must be retained through the existing participant/plan authority—never inferred from a fresh primary record or a process map. (`postgres/recovery_participant.go:184`; `otp_execution.go:307`; `plan.go:181`; `claims.go:24`)
2. **Authenticated-session variant lacks frozen complete session custody.** OTP authorization currently retains only session ID, whereas `sessionpostgres.RequireSession` requires the exact original complete `session.Session`, including family/digest/version/lifetime. ID lookup followed by adopting current record would refresh authority. Safely support an explicit bound original-session profile, or reject this variant pending that contract. (`otp/authorization.go:39,51,88`; `session/postgres/admission.go:49`)
3. **Correct non-consuming check has no successful durable coordinator route.** Current denial disposition is only `Unchanged` or `Denied`; unchanged rolls back and returns conflict, reservation rejects `OTPCheck`. Wiring `AttemptChecked` through denial would lose or misclassify a metered success. It needs an explicit coordinator-owned counted-check result path, not a direct store API. (`postgres/otp_attempt.go:19,225`; `otp_reservation.go:141`)

**Safe independent slice:** native exact-instance contributor plus the existing locked helpers can activate consuming pre-auth/user variants once recovery original-flow custody is made durable in the existing protocol. Do not claim authenticated-session or successful-check support by borrowing current records or treating denial plumbing as success. No additional journal, transaction owner, callback-injected authority, or evidence gate is needed.

### QA: /root/otp_native_contributor_qa_b523b3f

Message Type: FINAL_ANSWER
Task name: /root
Sender: /root/otp_native_contributor_qa_b523b3f
Payload:
## Native QA planning report

**Role:** read-only QA planner using `$qa-test-planner`.
**Reviewed base/head:** `b523b3feb87a0428def6dcd75a692decb1a8079d`.
**Scope:** activate the production OTP contributor through the existing coordinator protocol. No writes, Go commands, services, or subagents used.

**First-write blockers:** none identified. Existing private helpers and sealed coordinator phase capabilities support the required composition. Helper-only tests and coordinator fixture contributors do **not** prove that production composition works.

### Bounded coverage artifact

Add real PostgreSQL tests that construct the actual production contributor, enlist it in the coordinator registry, submit mapped opaque commands, and inspect authoritative rows and protected effects.

| Risk / realistic defect | Observable assertions |
|---|---|
| Production contributor never reaches existing phases | Correct submission through coordinator produces one reserved challenge; domain commit atomically produces finalized OTP, committed command, owning mutation and required audit/outbox effects. No fixture OTP implementation substitutes for production code. |
| Scope or custody substitution borrows another challenge’s budget | Table-driven mutations of tenant, purpose, subject/channel, workflow target, expected version, digest-key version, attempt ID/fingerprint, and authorization variant deny with unchanged unrelated counters/state. Original sealed authority is required; legacy mapping without that authority cannot silently gain it. |
| Current reads refresh stale original authority | After mapping, bump each original directory/user authorization or security epoch, or revoke/block the applicable credential/user. Admission and protected-use checks deny; no counter charge or protected effect. Include anonymous and user-bound admission, and actual pre-auth cookie possession. |
| Wrong-code charging rolls back, duplicates, or leaks into reserve | Wrong consuming submission commits exactly one failed/used increment and stable aborted command result; exact replay changes nothing. Changed identity/fingerprint conflicts unchanged. Correct consuming submission never goes through counted-denial terminalization. Cancellation or required-audit failure leaves the denial transaction’s entire delta absent. |
| Non-consuming checks grant usable authority | Correct check advances only its bounded counters and returns no reservation/proof or protected mutation. Check-limit and shared attempt-budget boundaries deny at exhaustion. |
| Brute-force bypass through resend or purpose changes | At budget minus one and exactly budget, persist exhaustion and required shared lockout atomically. Same subject/destination across challenge replacement and relevant purpose changes remains locked; unrelated scopes remain eligible. Exact lockout expiry and expiry after row-lock wait use fresh database time. |
| Competing attempts produce multiple winners | Run two correct commands and a correct/wrong pair concurrently against the same challenge. At most one reservation exists; counters never exceed the issued budget; losing commands have no protected effects. Do not require an arbitrary scheduling winner. |
| Takeover partially transfers authority | Expire coordinator ownership, then replay the original command. Exact predecessor generation transfers all enlisted participants atomically, without another code charge or rewritten original expectations. A mismatched/terminal participant rolls back every transfer. Stale-generation apply/finalize denies. |
| Finalize skips fresh fences or escapes domain atomicity | Change purpose/version/flow authority or expire challenge between reserve and protected commit, including lock wait. Protected effects do not commit. Inject a downstream audit/outbox failure after OTP finalization SQL; no finalized OTP or committed domain/result survives. |
| Recovery creates fresh authority or releases ambiguity | Committed outcome returns stable classification without comparison, charge, or renewed authority. Known noncommit terminally releases with exact original linkage; released OTP cannot reserve again. Unknown/missing/conflicting evidence remains pending/reserved. Lease loss, challenge expiry, or lost response alone never authorizes release. Recovery requires no fresh code or bearer and cannot substitute current epochs for originals. |

### Smallest sufficient verification set

1. **Focused red:** one real-PG production-composition test currently unable to complete native reserve → protected commit. Assert durable outcomes, not registration or interface satisfaction.
2. **Green composition suite:** grouped tests for admission, denial/check accounting, reserve/rebind concurrency, protected commit rollback, and release/recovery above.
3. **Affected existing suites:** reuse native challenge, lockout, ownership, pre-auth admission and authority tests; retain coordinator OTP phase suites as complementary boundary evidence.
4. **Targeted race/concurrency:** execute the production-composition concurrent cases under race detection. Use bounded barriers and database lock coordination rather than sleeps.
5. **Local compatibility:** affected OTP, OTP/Postgres and coordinator packages plus direct consumers using their existing repository checks. Every Go execution requires disposable `GOCACHE` and cleanup.

Deterministic pure custody/profile validation belongs in unit tests; transaction ownership, current authority, atomicity and ambiguity require real PostgreSQL. Existing helper fixtures may seed issued state, but must not replace the production contributor being activated.

### Claim boundaries

This batch can establish **native OTP contributor activation**, not full OTP completion. Issue/resend/delivery orchestration, cleanup/key-retirement tombstones, complete workflow/session composition, migration variants and all applicable exact coverage, mutation, fuzz, interoperability and release gates remain within the original full OTP scope. No execution evidence was produced in this planning pass.

## Original OTP authority custody — 2026-10-07

Source checkpoint `918965f0a05adc9e73fa8f7474091ff586150abc`, parent
`2f9b18eec1756acd6fc2552b64eabfa781d3e0ef`; single primary writer.
Pre-write mapping/QA: native `otp_native_mapping` / `otp_native_qa` reports
in the parent session, identifying original authority custody and drift oracles.
Raw executions remain in the parent native transcript: RED `8c7c09`, `1d8441`,
`a206d1`; complete adapter race PASS `fcf7f0` (1.246s); bounded parser fuzz
PASS `6772d6` (345,462 executions); standalone integration-tag vet exit 0
`265c84` supersedes masked shell status `93d1e9`. Reviewed hashes match staged
source (`e6c7c3`); normal local commit `700466`, clean source `09cb32`.
ACK Core and message lines passed (`2272bb`, `283b27`); no project profile.
No PostgreSQL execution for this code-free parser/custody change. Existing
source-lock/clean-consumer debt, native participant activation, live primary
phase drift denial, full release gates and complete OTP remain pending.
Owned caches removed (`b6fba7`, `09cb32`); no services were created.
Counts unchanged: 5/67 integrated, 1/67 verified. Goal active; no push.

### Native report: /root/otp_authority_regression

Message Type: FINAL_ANSWER
Task name: /root
Sender: /root/otp_authority_regression
Payload:
Native role: `regression_reviewer`
Disposition: **APPROVE — final bounded custody checkpoint**

Re-audited all seven changed files against base/HEAD `2f9b18eec1756acd6fc2552b64eabfa781d3e0ef`.

No concrete regression found. Passwordless user commands now retain all three independent security epochs without requiring a credential reference; each missing epoch is rejected. Anonymous commands reject user dimensions. Existing custody, parser, restart, legacy, and pre-auth flow/cookie coverage remains applicable.

Reviewed authority source SHA-256: `ed14ea3837a3a5101d241cea020c56fd8122b2ac21404895c0d95e5a74a831d8`; test SHA-256: `09cb8ef717bfaf72f2cbdba0566d3feccaa0012f5a31b4a57379a543800c424b`.

Executed no Go commands; race/vet results were parent-supplied. Native activation, live primary drift denial, end-to-end workflows, and full release gates remain unverified. This supersedes my earlier snapshot review and does not establish unit completion.

### Native report: /root/otp_authority_compatibility

Message Type: FINAL_ANSWER
Task name: /root
Sender: /root/otp_authority_compatibility
Payload:
## Native Go compatibility report

**Role:** `go_compatibility`
**Disposition:** `APPROVE` for the local intermediate development checkpoint only.
**Reviewed repository:** `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`
**Branch/base/HEAD:** `feature/identity-platform`, `2f9b18eec1756acd6fc2552b64eabfa781d3e0ef`
**Reviewed change:** complete seven-file dirty snapshot, including both untracked authority-custody files.

### Findings

No unresolved compatibility defect found in this snapshot.

- `Map` and `MapPreAuth` retain their signatures and unchanged v1/v2 encoding paths. New authority-aware methods are opt-in additions.
- v3 embeds the original v1/v2 projection without redefining its core binding or MAC. Restoration authenticates the envelope, validates authority framing, verifies the original issuer MAC and reconstructs the original plan before returning custody.
- User commands require independent positive user, identifier and credential-security epochs even without an existing credential reference. Passwordless round-trip and missing-epoch rejection are represented in the final tests. Anonymous commands reject all user dimensions.
- Restoration retains exact named wrapping keys, fingerprint-key versions and digest-key versions. No current-key fallback or invented authority for legacy records was introduced.
- Actual ownership remains acyclic: OTP PostgreSQL consumes kernel, OTP, session and authorization contracts; authorization does not import identity. `go-authorization/postgres` belongs to the authorization root module, so the new direct root requirement is correct.
- The authorization pseudo-version resolves locally to `60ed4b3229d83dabc40b3e81235ee7a69a95c4b1`; its commit time matches `20261005011349`. Its committed `go.mod` contains no identity dependency or permanent replacement.

### Remaining boundaries

This is not clean-consumer, release or OTP-unit completion evidence. Existing root/session source-pin debt remains; local workspace substitutions do not resolve it. Native contributor activation, live phase-by-phase authority checks and end-to-end login remain pending and are explicitly documented.

New v3 records are not readable by pre-v3 binaries; the additive API profile requires updated readers. Existing v1/v2 records remain readable by the new implementation.

No Go execution, mutation, external verification or subagent delegation was performed. Parent-reported race/vet/fuzz results are not independently attested here.

### Snapshot identity

SHA-256:
```text
CHANGELOG.md              7d7b2bf9bc335fca2641748c750bfad3cdbd4cb886d0975d36847ce04cf07f51
README.md                 7282e35d98294dab39f6f62c4845d1b4a54cfe15985f4554f3ed66e406868815
command.go                1e5c6d98fda4f7dcde8e45f465aa4e0478e5c011a0ede13225f4437aa8ca1527
go.mod                    c3f89a1cf258163e640de1b543effdd69ca4dbcd9eac2fc846c9143e6528e340
preauth_cookie.go         539ac1f7c07df9dfa0570609878e1c299ba211d58a5edfc43721c54bca7867d2
authority_custody.go      ed14ea3837a3a5101d241cea020c56fd8122b2ac21404895c0d95e5a74a831d8
authority_custody_test.go  09cb8ef717bfaf72f2cbdba0566d3feccaa0012f5a31b4a57379a543800c424b
```



### Native report: /root/otp_authority_v3_review_20261007

Message Type: FINAL_ANSWER
Task name: /root
Sender: /root/otp_authority_v3_review_20261007
Payload:
**Disposition: APPROVE — custody-only development checkpoint. No findings.**

**Native reviewer:** `/root/otp_authority_v3_review_20261007`, independent read-only single-reviewer structured pass using `requesting-code-review` and `testing-coverage` guidance.

**Boundary:** `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`, branch `feature/identity-platform`, base/current committed HEAD `2f9b18eec1756acd6fc2552b64eabfa781d3e0ef` plus the complete seven-file dirty patch below. Final README rollout amendment was inspected separately; source and test identities remained unchanged.

### Source-backed closure

- Authority validation requires original deployment/tenant epochs. User commands independently require effective-user authorization and all three user/identifier/credential security epochs, including passwordless commands without a credential reference. Anonymous commands reject every user authority dimension.
- Optional credential references retain exact tenant, user, ID, kind/owner pairing and full-width version through canonical `CredentialRef.Validate()`.
- V3 custody embeds the unchanged V1/V2 attempt encoding. Restoration authenticates the envelope, validates the bounded authority projection, decodes the original attempt, verifies its original issuer MAC and validates its retained digest plan before returning either projection.
- Original code remains opaque root custody; cookie custody remains separate. The pre-auth consumer restores the original flow and compares the submitted cookie under its exact retained key. No fresh authority reads, SQL changes, reservations or permission decisions were introduced.
- Bounded challenges traced foreign tenant/effective-user substitution, missing passwordless epochs, malformed authority tags, truncation/trailing data, caller mutation, legacy records lacking authority, and altered ciphertext. Guards reject these without returning usable custody.
- Old mapping remains unchanged. New readers retain old profiles; old binaries reject opt-in V3. The final README correctly requires reader-first rollout.
- Tests use independently specified authority values and denial outcomes. Restart, exact plan/cookie/flow conservation, passwordless epochs, malformed profiles and original-value retention exercise the changed boundary. Existing restoration tests continue through the refactored shared path.

### Verification boundary

Native parent transcript confirms missing-API RED results `8c7c09`, `1d8441`, `a206d1`; final `go test -race -count=1 .` PASS `fcf7f0`; bounded `FuzzAuthorityRecord` PASS `6772d6` with 345,462 executions. Independently inspected `git diff --check`: PASS.

`93d1e9` contains the integration-tagged vet invocation without reported diagnostics, but subsequent shell commands mask its individual exit status. **Do not describe that artifact as independently established vet exit 0.** This does not block this intermediate checkpoint; mandatory full-release verification remains pending.

This approval does **not** establish native contributor activation, live authority drift enforcement, PostgreSQL execution, clean-consumer/source-pin readiness, complete OTP unit verification, publication or deployment. Existing release debt remains unwaived. No files, services or external state were changed by this reviewer.

### Final seven-file SHA-256 identities

```text
7d7b2bf9bc335fca2641748c750bfad3cdbd4cb886d0975d36847ce04cf07f51  otp/postgres/CHANGELOG.md
5445bcc8de5deffbfa76548953d529ad7159f573e3504a286d80efb6d886de7d  otp/postgres/README.md
1e5c6d98fda4f7dcde8e45f465aa4e0478e5c011a0ede13225f4437aa8ca1527  otp/postgres/command.go
c3f89a1cf258163e640de1b543effdd69ca4dbcd9eac2fc846c9143e6528e340  otp/postgres/go.mod
539ac1f7c07df9dfa0570609878e1c299ba211d58a5edfc43721c54bca7867d2  otp/postgres/preauth_cookie.go
ed14ea3837a3a5101d241cea020c56fd8122b2ac21404895c0d95e5a74a831d8  otp/postgres/authority_custody.go
09cb8ef717bfaf72f2cbdba0566d3feccaa0012f5a31b4a57379a543800c424b  otp/postgres/authority_custody_test.go
```


## Owned pre-auth preparation and cookie issuance — native independent review

Principal `/root/preauth_cookie_issuer_final_review`, gpt-6.1-sol medium,
isolated read-only context. Native report below. Reviewed hashes matched3aa384;
unchanged snapshot locally committed as c53bc05072105afb4f15d6d07abac2d5822b85b1.
Original executions: missing issuer REDf1599e; focused core PASSad4cfa0.261s;
core/adapter race PASS093c9b1.527s/1.298s and integration vet exit0;
PostgreSQL18.3 focused composed journey PASSe378331.725s, full native matrix
PASSe41f9a9.557s. Container absent0461b4; owned caches removed0f2a3f.
No execution is relabelled as a rerun at the later commit.

APPROVE — development batch only. No findings.

Reviewed `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`, branch `feature/identity-platform`, exact base/HEAD `3dd2ae21e94df71f897e3cebafe180dd083fd995` plus the complete frozen eight-file dirty snapshot, including both untracked files.

Source-backed closure:
- `preauth_issuer.go`: preparation privately owns original entropy/time/commitment; copies share one atomic attempt. Complete receipt binding, operation/version, error, cancellation, replay and remaining-lifetime checks dominate credential delivery.
- `preauth_issue_execution.go`, coordinator reservation and native effect: durable idempotency and atomic flow/evidence/audit/outbox remain under the existing coordinator. Terminal observations cannot gain fresh-execution eligibility.
- Recovery clears `Flow` and never returns credentials or asserts reconstructed historical payload. The smallest mixed-payload challenge therefore cannot turn caller-provided metadata into recovered flow evidence.
- Core operation descriptor matches the base adapter descriptor exactly; mapping, persisted schema and migration bytes remain unchanged.
- Tests independently pin cookie scope, fractional TTL, both remember choices, matching bearer commitment, reveal-once behavior, copied-preparation concurrency, failure suppression and outcome-only recovery. Native composition checks persisted original bytes and conserved durable effects.

Independently read the integration invocation/workspace and confirmed the selected integration root reaches the new composed journey. Parent-provided native results report focused core, core/adapter race, vet, focused PostgreSQL journey and full `TestNativeFamilyRevocationIntegration` passing. These were not independently executed here. Independently executed `git diff --check`: passed.

Exact SHA-256:
```text
session/preauth_issuer.go
498534084b6492adcfb29da04a6e80dfc1bdea5ec0588d13a12832868ed3ff12
session/preauth_issuer_test.go
7961be85baf3701b5e45041d3981fa78b586d6e110461ba2c6a2bb82125e635f
session/CHANGELOG.md
8a0c3b70bb238efb5503171657ca1bc23d2b46187dc0421b081ec763187a5ad5
session/README.md
2b05a288e41b9e06661e9566e2ee1ce65e2f3a24e16aac9eeff4a6490b70debc
session/postgres/CHANGELOG.md
ef74ff8b6bfc2d4bec5c5912691f334dadca9b5319c9e08dfc9f6f45d8e030fa
session/postgres/README.md
46633c020c73b3a9f08ed553fc6416e4e1d17bf670ebbe074b8b04bddc0fc30b
session/postgres/preauth_issue_mapping.go
076dadff0b67e7811ba86553018f603ab313ee82ad57c26adf2804ee372fc92f
session/postgres/preauth_issue_execution_integration_test.go
4ddd35782ba09202f012d0738bc8d63b2f1f9a8e5eaeaa8c7e8faf23b66e61c7
```

Limits: no package-verification, release, clean-consumer, aggregate coverage/mutation, hosted CI, HTTP/CSRF route, completed OTP continuation or usable-login certification. No writes, runtime commands, commits or delegation performed.

## Pre-auth replay-delivery distinction — native independent review

Principal `/root/preauth_replay_final_review`, gpt-6.1-sol medium, isolated
read-only context. Native report below; unchanged reviewed snapshot committed
locally as3dd2ae21e94df71f897e3cebafe180dd083fd995 (278472).
Execution results retain original tool identities: compilation REDc71b5f;
core/adapter race PASSdcfc81 (1.473s/1.280s) and integration vet exit0;
focused PostgreSQL18.3 PASS074c3b1.734s and full native matrix
PASS3131289.804s. No rerun is claimed at the later commit. Task-owned
PostgreSQL container absent26b3e5; caches removed107811.

**APPROVE — development batch only. No findings.**

Reviewed base/HEAD `c4fac93cf873a3e45e9a17b5a69aeaa03f43cd1b` plus the complete frozen nine-file dirty diff on `feature/identity-platform`.

Source-backed conclusions:
- `preauth_issue_execution.go:160–206` defaults to historical metadata and clears `Replayed` only after successful acquisition of a nonzero reservation. Invalid composition/context, denial, conflict and pending observation cannot claim custody.
- `postgres/coordinator.go:280–420` and `postgres/journal.go:34–104` establish exclusive execution and historical terminal replay. Challenge: replaying a committed command after payload retirement still returns zero reservation, so it cannot become fresh execution.
- Existing receipt-only methods/interfaces remain compatible; feature, audit, lifecycle, transaction closure and recovery owners are unchanged. No second journal or migration changes.
- Integration assertions independently require exactly one fresh committed result among eight calls (`preauth_issue_execution_integration_test.go:247–277`), historical/restarted replay, changed-intent conflict and pending denial. New result formatting/exports redact or deny custody.

Verification outputs were **parent-provided**, not independently executed: compilation RED; core/adapter race PASS; integration vet PASS; focused PostgreSQL PASS; full native PostgreSQL matrix PASS. Inspected the execution script and source-lock workspace. Coverage, mutation, clean-consumer and release gates remain deferred; this does **not** certify either package verified or cookie delivery implemented.

Final rechecked SHA-256 hashes:
```text
session/preauth_issue.go
fdede449773c11923306b927c2f08cbe6787143735878484fa7000da8e23d8c0
session/preauth_issue_test.go
8e16741b4e7822dddcb8a261fb1287f743cde315774269c4c6c9948081cfbac7
session/README.md
eee1671f9103011711c76feb47fa4f6cd1744c5c38d20448d19b068f95a56d41
session/CHANGELOG.md
a9bac1db00763f9fb5d1b79ee3cd95483eebf291f2b85cd2408141c954540e2e
session/postgres/preauth_issue_execution.go
f216868a9556a4eaa821ec2b3ce5297d1b64c56d744b68747935c0872e772ef4
session/postgres/preauth_issue_execution_test.go
398d3d6222df61f0cee629f01493e9184d9f511cb385e1fab1813012c940f948
session/postgres/preauth_issue_execution_integration_test.go
ef9e3fb17e56e748d41d8aa2c31dfb80ac7f2528e91c82c0e16a60fa1682862f
session/postgres/README.md
503076d86d93bae451f5de4df235a7b36e75e8f20f9e76ee5fde812d1302205a
session/postgres/CHANGELOG.md
b5f5ca170a5273954a188300204b34b49c4fda6de688e46a6391406c3c431d46
```

## Owned durable pre-auth initiation — native independent review

Principal `/root/preauth_owned_initiation_review`, gpt-6.1-sol medium, isolated
read-only context. Native report retained below with hard-break whitespace
normalized. All reviewed input
hashes matched3ac249; unchanged snapshot committed locally as
c4fac93cf873a3e45e9a17b5a69aeaa03f43cd1b. Original executions are not relabelled
as reruns at that revision. Missing executor REDf947a0; core contract RED31609f;
retired flow-ID resurrection RED569364 was fixed by retained tenant/flow uniqueness.
Earlier integration oracle correction preserved required failure audits rather
than incorrectly requiring zero audit rows after proven domain rollback.

## APPROVE — development checkpoint only

**Reviewer:** `/root/preauth_owned_initiation_review`
**Boundary:** `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`, branch `feature/identity-platform`; base/head `a1f2eb5d50a4555f7963e7836c9e623e6e999789`, plus the complete frozen 15-file dirty snapshot.

No concrete blocking or non-blocking findings. No unresolved review questions.

### Reviewed closure

- **Configured scope and ownership:** `preauth_issue_execution.go:102` constructs the service actor and separate anonymous-flow namespace. `preauth_issue_effect.go:23` and the existing closed mapping enforce exact configured intent and complete immutable original state.
- **Atomicity and expiry:** `preauth_issue_effect.go:44` stages original flow and retained evidence through the coordinator’s sealed `Work`. Existing `RequirePreAuth()` samples primary time after locking; kernel `ApplyFeature()` and `Commit()` own audit, lifecycle and receipt persistence.
- **Replay and retirement:** existing claim/fingerprint and journal fences return historical receipts without applying another flow. `preauth_issue_migration.go:39` adds the permanent tenant/flow uniqueness fence; a different command cannot resurrect a retired flow, and transaction rollback removes its attempted insertion.
- **Recovery:** `preauth_issue_execution.go:200` routes exact original anonymous scope through current authorization and the sole kernel recovery owner. Recovery neither reserves a missing command nor derives noncommit from live-payload absence. Native recovery instead checks retained issuance evidence.
- **Tests and compatibility:** inspected every new test, changed integration registration, fixture provenance, existing mapping/storage tests and kernel recovery contracts. Public additions are storage-neutral and additive; docs explicitly exclude entropy preparation, cookie delivery, OTP continuation and complete browser login.

### Verification boundary

Execution evidence was supplied by the parent, **not executed by this reviewer**:

- `go test -race -count=1 github.com/faustbrian/go-identity/session .` — core **PASS 1.497s**, adapter **PASS 1.488s** (`c00382`).
- `go test -tags=integration -race -count=1 -run '^TestNativeFamilyRevocationIntegration$/authenticated_issuance/owned_executor_with_real_audit_and_lifecycle/owned_pre-auth_initiation' .` — PostgreSQL 18.3 **PASS 1.696s** (`662328`).
- `go vet -tags=integration github.com/faustbrian/go-identity/session .` — **exit 0** (`22d3ae`).
- Reviewer independently inspected `git diff --check` — **exit 0**.

Go executions used the supplied source-locked `GOWORK` and task-owned disposable caches. Exact coverage/mutation, clean-consumer/publication, provider/full-release gates remain uncertified. This approval does **not** certify either package verified, release readiness, browser login, or completion of the 67-unit objective.

### SHA-256 of reviewed files

```text
b421e8ac95447109641cf012f2ddb7271fe00e75a555e5aa5f8856e2104e0edd  session/CHANGELOG.md
bb8d95de9c85babf7f34389135439cd450f8d7a7d0d2cb9ad3c4d1706c6cab3b  session/README.md
6d942788b481487eae13bf3100efa89f47166e00519dca1c72222714d73a3a4c  session/postgres/CHANGELOG.md
ebe2f20d2dca3a6aea47269f94adaca85543493b2abf22d59cde0a5e57e769d1  session/postgres/README.md
41efb6643b0070214416db5c6081faedc8cb2b2b057f7559319990be5c5f76ee  session/postgres/issue_integration_test.go
e4b2eade989f74f1f53e15a9ba2e66d8cbe8fd79d6825a05cad029b16a9f80f8  session/postgres/repository.go
4558a10b934199afc52d2241c7e6622830c0208af8f56dfaa56c183e26b71795  session/postgres/revocation.go
10ddd19e32ad5af206b84b0aa83e1b6821cb3007d8e2f99161bb817905e0ed61  session/postgres/revocation_integration_test.go
2ed3ceb352332d623c1a17d7875b7c1c9a1de6bb266042d5440289d9b637d2f8  session/preauth_issue.go
24f8c766993bd7e028edc31c74a0698fc5be3b9120099bf866f5b108296f30ce  session/preauth_issue_test.go
8503eb5431ec4cb1a562d05216fc6e55b9411e48e3c0f5c06f67de8c3a3e6f41  session/postgres/preauth_issue_effect.go
01c244537e4c361415b3f70745e0b90bd3169de490d3115736b832b9f4f9cf16  session/postgres/preauth_issue_execution.go
c8f05f42c127c4330704ed0292c0277f180f4bdba2e9ff47bc9df5935aa14ac5  session/postgres/preauth_issue_execution_test.go
1b2543ab58f65ee2947501987c1165f5680620d6def6e187d31d358db0575e03  session/postgres/preauth_issue_execution_integration_test.go
1f3533c7b411495cd06a29e9b4c4ee265e33c731a8d8e46a9807704a1f625b92  session/postgres/preauth_issue_migration.go
```

## Pre-auth initiation command mapping — native independent review

Principal `/root/preauth_initiation_mapping_review`, gpt-6.1-sol medium,
isolated read-only context. Native report retained verbatim below. Reviewed
unchanged inputs committed locally as a1f2eb5d50a4555f7963e7836c9e623e6e999789;
original execution results are not relabelled as reruns at that revision.

**APPROVE — development checkpoint only. No findings.**

Boundary: `feature/identity-platform`, base/head `2adb5ec5a0fc3f63ee862ebea2c748422037fb81`, plus the complete four-file dirty snapshot. Single-reviewer, read-only independent pass.

Source closure:
- `preauth_issue_mapping.go:20–52` retains every native record field and binds tenant, anonymous-flow target, record and policy independently of actor.
- `preauth_issue_mapping.go:60–130` rejects malformed/missing/unknown scalars, narrowing overflow, noncanonical timelines and substituted record scope.
- Existing feature construction copies and sorts scalars; existing coordinator codec fingerprints each scalar. No operation registration, journal, public API, migration or execution authority was added.
- Tests exercise canonical-record preservation, every missing field, unknown fields, scope substitutions and malformed scalar partitions.

Verification: supplied RED58928b/RED6b3287, adapter race PASS (1.292s, tool47363a), vet/diffcheck exit 0 (d52f1c); not independently rerun. Exact coverage, mutation, SQL integration, public initiation/recovery and release readiness remain uncertified.

SHA-256, independently checked:
```text
session/postgres/preauth_issue_mapping.go
a4ad0ef1512b920f9550eabba98247b210b909443494807ba2fdf47afbac6092
session/postgres/preauth_issue_mapping_test.go
acf9a6bf976e75d55b7fd084796e344914d51c65d9fdae2d503e40528b7831c1
session/postgres/README.md
f40d047a592d7e802ca7ec02228c869550670906f176cc3b0ac551a536841748
session/postgres/CHANGELOG.md
0d4bc28b97f4a11aca0d59607b075afccf787331a9dfa54cbe5d1bf44af43e05
```

## Session PostgreSQL terminal pre-auth release — final native report

Principal `/root/preauth_release_final_review`, gpt-6.1-sol medium, isolated
read-only context. Refreshed native report verbatim; nine hashes matchedf6510a,
source committed unchanged at2adb5ec5a0fc3f63ee862ebea2c748422037fb81.

## DEVELOPMENT disposition: APPROVE

No findings on the refreshed complete nine-file diff. This report supersedes my earlier approval, which missed caller-cancellation classification.

**Reviewed boundary:** `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`, `feature/identity-platform`, base/HEAD `2a9a25f0265c43377332465bcd7a25ad0b6eb277` plus the supplied dirty snapshot. All nine hashes match; changed hashes:
- `preauth_release.go`: `1007aedc1cf7f8e1332f6e439e7317e186e7ebace46d1870dcda0e504fdd9d00`
- `preauth_release_test.go`: `29453fa9788c36bbc91ea526782653368a6a13e1989e3307c0510676ebaa058c`

**Closure:** `preauth_release.go:32–34` now classifies an already-canceled or deadline-expired caller before kernel binding. The new regression asserts `CancelledError`, independently of authority denial. Exact-owner sealed authority, predecessor/successor fencing, original-record exclusion, terminal CAS/replay denial, post-lock live expiry and forward populated-state migration remain intact across the complete final diff.

**Verification:** supplied cancellation RED `2f88dc`; final unit race PASS 1.292s and integration vet exit 0 `9d41c2`. Unchanged SQL/migration integration/race evidence remains PASS 4.283s `bdccb0`. Reviewer `git diff --check` passed; no Go/service execution or writes performed.

**Claim limit:** development approval only. Positive public sealed participant composition, owned issuance/recovery, exact coverage/mutation, publication/clean-consumer and remaining release gates are not certified.

## Session PostgreSQL terminal pre-auth release — initial native report

Principal `/root/preauth_release_final_review`, gpt-6.1-sol medium, isolated
read-only context. Original nine-file report below is retained verbatim;
subsequent caller-cancellation correction invalidated final-snapshot approval.

## DEVELOPMENT disposition: APPROVE

No findings. Independent read-only, single-reviewer complete-final-diff pass using `$requesting-code-review`, with coverage/integration guidance.

**Boundary:** `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`, branch `feature/identity-platform`, base/HEAD `2a9a25f0265c43377332465bcd7a25ad0b6eb277` plus the supplied nine-file dirty snapshot. All nine SHA-256 values match the assignment; production hashes rechecked before reporting:
- `preauth.go`: `6b628cda1ae9502896fb20acbc1b56938175d7c12b74e31c393ca38e3f30aba8`
- `preauth_release.go`: `c10939c61f72e7ad4075fe4b3291104c014c02dbfe8e564f2395512d867179b4`
- `preauth_release_migration.go`: `c4bcec505554a4a490cf3fbaa38d9880b07c13ff7aaef1e0b59994812335bdf2`

**Source-backed closure:**
- `preauth_release.go:28–37` requires the opaque live kernel handle and exact registered owner; `postgres/otp_execution.go:42–82,278–306` fences callback lifetime, ownership, proof phase, rollback generation and immediate recovery successor.
- `preauth_release.go:48–56` locks the exact original record/custody before a one-row terminal CAS. Changed fingerprint, stale/skipped generation, missing row and terminal replay deny. UPDATE preserves original record, cookie and timeline.
- `preauth.go:167–233` retains post-lock database-time sampling and exclusive expiry for every live-admission caller. Release deliberately shares only record exclusion, so expiry permits no reactivation.
- `preauth_release_migration.go:40–47` adds `released` through forward transactional constraint replacement; original migration remains unchanged. Old issued/reserved/consumed semantics remain valid.
- Integration assertions exercise expired terminal release, both legal generations, rejection, rollback, committed replay denial, unchanged payload and populated-state preservation. Smallest adversarial challenges—changed custody, skipped successor and terminal rerelease—remain denied.

**Verification:** supplied RED `e1709e`, untagged race PASS `594b46`, PostgreSQL 18.3 integration/race PASS and integration vet exit 0 `bdccb0` accepted only within their stated boundaries; not independently rerun. Reviewer `git diff --check` passed.

**Explicit exclusions:** positive release tests execute private SQL, not a real usable public OTP participant journey. Owned participant wiring, issuance/recovery composition, exact coverage/mutation, dependency publication, clean-consumer and other release gates remain pending. This approval is a development gate, not package verification, release readiness or full-goal completion.

## Session PostgreSQL reserved pre-auth review

Principal `/root/preauth_reserved_final_review`, gpt-6.1-sol medium, isolated
read-only context. Native report verbatim; six hashes matchedeb88b8 and source
committed unchanged at2a9a25f0265c43377332465bcd7a25ad0b6eb277.

**APPROVE — development batch only. No findings.**

Native reviewer: `/root/preauth_reserved_final_review`; single read-only reviewer, no delegation or execution of Go/service commands.

**Reviewed boundary:** repository `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`, branch `feature/identity-platform`; base = HEAD `9248f5cc8e4337ea62b1c2c597faf8581bfe0183`, plus exactly these six dirty files under `session/postgres`:

| File | SHA-256 |
|---|---|
| `preauth_reserved.go` | `aa0b2d3f69ebafc1c2184071c7fbe2bf78acfca180876fbc53530157bfb9c37a` |
| `preauth_read_test.go` | `b9aa86573c13c3c77bf0891688d1153f2af0e9c83dc0b144bdce7a2aa4204580` |
| `preauth_test.go` | `864f746018b3397d24a5785c02293a0c93637f9965959e3ac209bde4d6929631` |
| `preauth_integration_test.go` | `86b505fa9f4a5cdfe26930cadc0e352fefeee8774376244ef057a62083537d70` |
| `README.md` | `2cdb565495912b7e516519fa666be6c2b9793f7248a51644c72e34ca1f19b7ce` |
| `CHANGELOG.md` | `36b549b80b85ba5046dd6ff680e9cace9b5e552b4d19eac88c39dc364fce1079` |

HEAD and hashes reconfirmed after review; `git diff --check` passed.

**Source-backed closure:**
- `preauth_reserved.go:24–70`: structural loading scopes tenant/ID, validates canonical payload against physical indexes, preserves original values, closes cursors, and classifies empty-cursor cancellation correctly. It deliberately grants no active-flow authority.
- `preauth_reserved.go:79–88`, `preauth.go:169–220`: reserved admission reuses complete original-record and exact-owner matching, retains a share lock, then checks exclusive expiry against primary time after lock acquisition.
- `preauth_reserved.go:102–116`: transfer requires a positive exact predecessor and immediate successor; the locked CAS changes only generation. Challenged stale predecessor, mismatched fingerprint, terminal state, overflow and rollback cases all deny or preserve prior ownership as required.
- `preauth_integration_test.go:177–296`: real-transaction assertions cover original-record conservation, foreign tenant denial, transfer, stale finalization, terminal denial, rollback, lock exclusion and indexed corruption. Existing shared-lock helper regression covers expiry after lock wait.
- Repository-wide new-API searches found definitions and tests only: no public OTP activation or additional runtime authority was introduced.

**Verification provenance:** parent reports cancellation RED `e82017`, final unit race PASS 1.302s, PostgreSQL 18.3 integration/race PASS 2.011s, and integration-tagged vet exit 0. I reviewed their behavioral oracles; I did not rerun these commands.

No unresolved development finding. Clean-consumer/publication proof, exact coverage/mutation, remaining release gates and complete usable workflow composition remain pending. This disposition does **not** certify release readiness or full-unit completion.

## Session PostgreSQL pre-auth final review

Principal `/root/preauth_pg_final_review`, gpt-6.1-sol medium, isolated read-only
context. Native report verbatim; eight hashes matched1b3703, committed9248f5c.

**APPROVE — native development checkpoint. No findings.**

Reviewed the complete eight-file dirty diff in `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`, branch `feature/identity-platform`, against base/head `771638a6f6d416b578f3f44949274b814373b28a`.

Source traces cover canonical binding/codec, indexed scope and versions, migration history, row locking, post-lock primary time, reservation custody, finalization/rollback, error redaction, cursor cleanup, and compatibility with `WorkRows` and sealed OTP phase interfaces. Falsification challenges—altered original binding, wrong reservation owner, competing reservations, expired-after-lock flow, and rollback followed by retry—remain denied or transactional under the reviewed implementation.

Supplied native evidence:
- Focused `TestPreAuth`: PASS.
- Existing `TestSessionPostgresIntegration`, real PostgreSQL 18.3, race: PASS.
- Complete untagged package race: PASS.
- Integration vet: PASS.
- `FuzzPreAuthStoredRecord`, bounded three-second budget: PASS, 222 executions.

Tests were not independently executed during this read-only review. Development uses an explicit disposable source-locked workspace; the committed module still pins the older session dependency.

**Not release approval:** public initiation/receipt recovery, reservation rebind/release, actual OTP/workflow registration, clean-consumer dependency integration, exact coverage/mutation and remaining package/release gates remain pending.

Verified SHA-256 boundary, all paths beneath `session/postgres/`:

```text
preauth.go
2cd6c237670818cdba64acf5394990ebe6badf228cc245fb95789d5665794f9c
preauth_record.go
95c4498836b959d27faa058cdce3977e9a47e6da0b18d4c2625f4e93d7cf3912
preauth_migration.go
1c77ebbeddf870c5c14b22e1ac1a001e2466650a7df64adb2c60c75841983dd8
preauth_test.go
80b01eacbd1aeb7fbe6c764e17be4b9ae463ee2048ca5342814d55a7e35e8dc7
preauth_integration_test.go
278374494744b90f630e61fec656dd5fff86b96847db4fb2228c9b2f5266859b
repository_integration_test.go
696e2e87f73de377d0b22b3d8ce52094099c7c5468bde52484067f9ae5d97bdf
README.md
b45cfd84ff9bc4140cac95f4db9c3100f030f194cdf6248d17093e9164f816ba
CHANGELOG.md
d3579fd26f94de88e4ce7e87e02f1ddcf1486bafb2a01548d4e116a3772cc4cd
```

## Session pre-auth final review

Principal `/root/preauth_binding_final_review`, gpt-6.1-sol medium, isolated
read-only context. Native report verbatim; hashes matchedd480af, committed771638a.

**APPROVE — No findings.**

Complete four-file dirty diff reviewed against base/head `f0800e94c702b9b58b933d5a1f73030eff8c6023`, branch `feature/identity-platform`.

Both previous findings are resolved: UTC bounds are validated before storage/framing, with public reconstruction/comparison regressions; the unreachable token-decoding failure branch is removed. Complete binding, key custody, expiry, immutability, URL and consumer-contract traces remain consistent.

`git diff --check`: PASS. Supplied race/vet/fuzz results were not independently rerun. Approval covers this local development batch only; durable reservation/finalization, HTTP composition, exact coverage, mutation and release gates remain pending.

```text
preauth.go      883fd145968740fa86da3d3816e1930e3e27e2fd4cc0e1d504418abadb74ca8d
preauth_test.go ecf678c36d4b39fcf06d6df0d2330892458fb0aca6523960f2e851b9cdabace2
README.md       e2f5d65477a5f59a99fadd8c586d66df3f390663aa4d7125ab9bd2415ab946bf
CHANGELOG.md    bdfcbfde701732e6255f480ab622853eccb8ca3f1a5e25793f420a50481fd7c5
```

## Session pre-auth initial review

Principal `/root/preauth_binding_final_review`, gpt-6.1-sol medium, isolated
read-only context. Native report verbatim:

**REQUEST_CHANGES** — complete four-file dirty diff on `feature/identity-platform`, base/head `f0800e94c702b9b58b933d5a1f73030eff8c6023`. Single-reviewer read-only pass; no project edits.

1. **Non-blocking:** [preauth.go:89](/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform/session/preauth.go:89) normalizes timestamps **after** validation. A year-1 local timestamp with offset `+01:00` passes `validTime()`, but becomes UTC year 0. The accepted record’s `Bindings()` then fails reconstruction and `Check()` validation. Standard-library reproduction confirmed `before=true,true ttl=5m0s afterUTC=false,false`. Validate normalized UTC timestamps before acceptance; add the round-trip regression.

2. **Non-blocking:** [preauth.go:182](/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform/session/preauth.go:182) adds an unreachable decoding-failure return. `NewToken()` and `ParseToken()` are the only `tokenState.wire` writers; both guarantee canonical 32–64-byte decoded values, and the state is immutable. Remove the redundant failure branch, consistent with `Token.Digest()`, rather than manufacture an invalid private state to satisfy coverage.

Repository policy requires confirmed findings resolved before committing; neither finding warrants a release-safety claim.

Remaining traces closed: exact intent equality, complete domain-separated framing, retained-key version matching, owned projections, exclusive expiry, bounded URLs, redacted errors/automatic serialization, and explicit separation from durable one-use authority. Supplied race/vet/fuzz results were not independently rerun. Disposable reproduction resources were removed. Coverage, mutation, durable composition and release gates remain pending.

Final SHA-256:
```text
preauth.go      4413436762d6b5715aa9fc272ac294dc102e53b38aa06ef2eaa740ed9473d046
preauth_test.go e290f24fc8e9a485e1aa0d802b245e7c0a008da7bf3fe6a93300fa1f67d961ae
README.md       e2f5d65477a5f59a99fadd8c586d66df3f390663aa4d7125ab9bd2415ab946bf
CHANGELOG.md    bdfcbfde701732e6255f480ab622853eccb8ca3f1a5e25793f420a50481fd7c5
```

## Kernel OTP authority final review

Principal `/root/otp_authority_final_review`, gpt-6.1-sol medium, isolated
read-only context. Native report verbatim; exact hashes matched50f087.

**APPROVE — No findings.**

Complete five-file development diff reviewed against base/HEAD `9721db43b2f1831c47f37fe2d1e0f9bb040dd77b`.

The prior blocker is resolved: every directory/credential helper error now poisons live Work before cursor cleanup can erase the failure. Authorization already enforces equivalent poisoning. Phase/lifetime validation precedes poisoning, so returned handles cannot affect later invocations. Release remains ungated.

New direct and real-primary cursor regressions exercise both affected dimensions across all four phases. Supplied final focused integration race, complete untagged unit race and integration vet results pass; independently inspected `git diff --check` passes. Tests were not independently rerun.

Approval covers this development batch only. Full native OTP session/pre-auth composition, exact coverage, mutation and affected release gates remain pending.

Reviewed SHA-256:
```text
postgres/otp_authority.go
469a4724e0ca5b64358476b7b16837891889bd2d48c4c8fafe824f05a43e6fd7
postgres/otp_authority_test.go
bf358cf06eb58d1799093f105b1fe40afd65ed4f9b1af8a74840bbb51ed96f07
postgres/otp_authority_integration_test.go
982caef608d9983bdf3089184a8aac672af355d16da5342e2dacca387cd724eb
postgres/README.md
2e001fccbd7b0d7995ba6a03eab51651e6ed3cd6c78c5862589bfc7199f0c1e0
postgres/CHANGELOG.md
70b047265ff2f4e3b70387d4a430c46d5bb3975be829a9c401f19b1d427a7089
```

## Kernel OTP authority initial review

Principal `/root/otp_authority_final_review`; initial five-file dirty snapshot
against identity9721db4. Native report verbatim; finding reproduced RED479a5d
before the fix, then focused GREEN6f58c9. Final review follows separately.

**REQUEST_CHANGES — one blocker.**

At [otp_authority.go:67](/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform/postgres/otp_authority.go:67), directory and credential methods return underlying helper errors without ensuring live Work is poisoned; this affects all four phases.

A live exact owner can open a cursor, call either admission check with valid expectations, ignore its error, close the cursor and return success. [Work.Query():133](/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform/postgres/work.go:133) rejects an already-active cursor without `Work.fail()`. The authority SQL therefore never runs, but subsequent execution remains permitted. Coordinator checks inspect only remaining cursors and Work usability, so closing the cursor defeats those checks.

Falsification: inspected attempt/reservation/apply/finalize callback enforcement; none remembers this discarded error. Existing identifier/privacy admission wrappers explicitly poison equivalent helper failures. User-authorization already poisons returned errors.

**Required:** poison every underlying admission failure after validating live phase/owner, preserving harmless returned-handle rejection. Add active-cursor regressions for directory and credential across all four phases.

Complete five-file development diff inspected against base/HEAD `9721db43b2f1831c47f37fe2d1e0f9bb040dd77b`; no other findings. Supplied focused integration race, unit race and vet results pass, but omit this case. No tests independently executed. Full native OTP composition, exact coverage, mutation and release verification remain pending—not blockers to development solely because pending.

Reviewed SHA-256:
```text
postgres/otp_authority.go
8fc4ca5a2500ac02f4daa1e67ef4bcb45b2859f01fc8faf8db68a2528fafe003
postgres/otp_authority_test.go
549ee216a37fb0c364630a18f781333c5df6955903afbc1a45c634c7ec6047d1
postgres/otp_authority_integration_test.go
340c3c7620ccc54b17ba900349f0ffc82a83cb79a6fcb0f875211c1ea0514bcd
postgres/README.md
2e001fccbd7b0d7995ba6a03eab51651e6ed3cd6c78c5862589bfc7199f0c1e0
postgres/CHANGELOG.md
70b047265ff2f4e3b70387d4a430c46d5bb3975be829a9c401f19b1d427a7089
```

## Native primary counter persistence review

Principal `/root/otp_primary_count_final_review`; reviewed hashes matched at
source commit2f4927f8c3692f3b76b1dd76b1ab3bc4fc5f86c7. Native report verbatim:

**APPROVE — development checkpoint only. No findings.**

Single-reviewer, isolated read-only complete-diff pass; no files changed and no tests independently rerun.

Reviewed branch `feature/identity-platform`, base/HEAD `64d44818d90265691d1874d9250537a619f6f55c`, plus all four dirty files:

| File | SHA-256 |
|---|---|
| `otp/postgres/challenge.go` | `8332c307741b12db513a780383a8705efd0135ed7bc227fd1aa7574f776a64ac` |
| `otp/postgres/challenge_integration_test.go` | `7f80164d59df9ead3eade9fa17edd85c8c98e9d1eaff8e2fbcd0eab97b7d95f8` |
| `otp/postgres/README.md` | `99a6a286cee9822ac0bb22408bccbd47e72727548948b3d207bba1419442a67f` |
| `otp/postgres/CHANGELOG.md` | `16b539557eb3fc5524d75f5b24551ca4235b19675b0331ceb45ded8c2811e828` |

Source traces closed:
- Authenticated original submission → retained keyed complete-scope lookup → committed issuance → exact row lock → validated frozen core policy/counters/full-width version.
- Core wrong/check/expiry proposals → issued-state/version CAS → SQL lifecycle guard → exclusive database-time expiry fence.
- Rollback leaves counters unchanged; stale concurrent submission cannot obtain another charge; correct consuming proposals are rejected.
- Rows close on every acquired path. Concurrent test cancellation, transaction rollback and worker join precede connection cleanup.

Bounded adversarial challenges included foreign destination, alternate retained key, maximum `uint64` version, correct final check versus wrong exhaustion, and expiry between evaluation and persistence. No reachable violation found within this private-helper contract.

Trusted parent-run evidence: expected missing-helper RED; PostgreSQL 18.3 integration/race PASS, 83.1% coverage; integration vet and diff checks passed. These results were not independently executed.

**Not certified:** public counted workflow, current flow/shared lockout authority, replay/results/audit, consuming participant/recovery, exact coverage/mutation, full adapter, clean-consumer/reference composition or release readiness. Those remain pending as required by the execution goal.

## Native PostgreSQL challenge storage final review

Native final report below retained verbatim. Reviewed source hashes matched at
local commite223a15 (d01964); source clean. Local ACK Core and line validation
passed with unspecified impact. No unit promotion and no push.

```markdown
## Internal disposition: APPROVE

**No findings.** Complete final dirty diff reviewed against base/head `84f8cd2317d789a535210942d35f393f79009c38` in `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`, branch `feature/identity-platform`: both manifests and all seven `otp/postgres` files. Final production, integration-test and README hashes match the supplied snapshot.

Previous findings are closed:
- `migrations.go:75` preserves empty, non-null pre-account subjects without changing immutable scope custody.
- `migrations_integration_test.go:34–40` registers administrative connection cleanup before subsequent fallible setup; cleanup ordering closes child connections, removes the database, then closes administration.
- `migrations.go:192–193` accommodates unchanged-generation rollback release and exact-successor recovery release while preserving finalized/revoked ownership. Skipped generations remain rejected.
- New PostgreSQL oracles independently observe preserved subject, rejected skipped successor, retained exact successor and absence of a partially installed namespace.

The earlier REQUEST_CHANGES report remains historical evidence; this disposition supersedes it for this snapshot.

**Verification:** supplied final `go test -race -tags integration -coverprofile=<owned temp> .` PASS **1.394s**, **14/15 statements**; supplied vet PASS; independently observed `git diff --check` PASS. Source review only; no Go execution or external mutation.

Approval is for the coherent **local storage/migration development checkpoint**, not unit completion or release readiness. Issuance, attempts/replay, authority, participant execution/recovery, cleanup, exact coverage, mutation and release/composition gates remain pending as documented.
```

## Native PostgreSQL challenge storage review, initial pass

Native report retained verbatim as decoded JSON below. The pre-account storage rejection was
reproduced81a88f/aae7fd before repair; subsequent one-CAS recovery release
REDf63610 was repaired as well. Final native module SQL/race d12e73/cc7a00
PASS1.394s (14/15 Go statements), vet/diffaaf753 PASS. All full adapter/release
gates remain pending. This checkpoint is storage, not a new integrated unit.

```json
"## Internal disposition: REQUEST_CHANGES\n\n**Reviewed boundary:** `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`, `feature/identity-platform`, base/head `84f8cd2317d789a535210942d35f393f79009c38`, plus the complete dirty patch in `modules.json`, `packages.json`, and all seven `otp/postgres` files. Production/integration-test SHA-256 values match the supplied snapshot. Single-reviewer read-only pass using requesting-code-review, PostgreSQL-routine, coverage and integration guidance; no execution or mutation.\n\n### Findings\n\n1. **Blocker — pre-account challenge storage contradicts the consumed OTP contract.**  \n   Anchor: `otp/postgres/migrations.go:75`. The `subject_id` constraint requires at least one byte, but `otp/digest.go:63–65,103–107` explicitly supports an empty subject for pre-account challenges; `otp/command_attempt.go:23–25,44` requires that empty subject for anonymous commands. The supported anonymous-command example at `otp/command_attempt_test.go:173–198` confirms reachability. Persisting that exact valid continuation fails at the new storage constraint, preventing anonymous/pre-account OTP custody. Permit the empty, non-null subject representation while retaining width/control-character defenses, and add a real PostgreSQL case preserving it. **Falsification:** the journal FK contains no subject restriction; replacing the empty subject with an invented identifier would instead violate `PrepareForCommand()`’s exact subject binding. This migration is uncommitted, so repair it directly before publication.\n\n2. **Non-blocking — setup failures leak the administrative connection.**  \n   Anchor: `otp/postgres/migrations_integration_test.go:30–43`. After successful connection, entropy failure or `CREATE DATABASE` rejection invokes `t.Fatal()` before connection cleanup is registered. The connection remains open until process teardown, contrary to task-owned resource cleanup requirements. Register administrative connection cleanup immediately after connection succeeds; register database removal only after creation succeeds. **Falsification:** the already-registered context cancellation does not call `admin.Close()`; successful setup’s later cleanup cannot run on these earlier exits.\n\n### Other dispositions\n\n- Wrong prechecks legitimately increment both failed-attempt and check counters; this agrees with `otp/challenge.go`, so simultaneous increments are **not** a defect.\n- Recovery’s direct successor-generation release projection is rejected by `migrations.go:192`, whereas unchanged-generation rollback release is accepted. However, the schema permits an atomic reserved-generation rebind followed by terminal release in the same coordinator transaction. Thus it accommodates recovery at this storage-only boundary; I do **not** classify the absent one-statement projection as a blocker. Native participant implementation must preserve the actual `OTPReleaseWork` predecessor/successor contract.\n- Null completeness, immutable policy/digest/expiry, full-width numeric custody, FK retention, terminal non-resurrection, identifier quoting, atomic forward-only migration and manifests/docs show no additional findings.\n\n**Verification boundary:** supplied race/integration PASS, vet PASS, diff-check PASS and registration PASS; independently inspected source only, not rerun. Coverage **93.3%**, mutation, full participant behavior, composition and release gates remain pending—not storage-checkpoint commit prerequisites under the authoritative development goal. No unit-completion or release-readiness approval."
```

## Independent consuming reservation review

Native report below retained verbatim. Parent independently matched supplied
hashes93fceb; unchanged reviewed snapshot committed161eb5b (89bce9), clean91e67b.
ACK Coree28c88 and lineeb6468 validation passed; no profile, impact unspecified.

**APPROVE — No findings.**

Reviewer: `/root/otp_reservation_final_review`, isolated single-reviewer structured pass using `$requesting-code-review` and `$testing-coverage`; read-only, no delegation or runtime diagnostics.

**Boundary:** complete dirty diff, including all three untracked OTP reservation files, against base/HEAD `ebae68e64c30622ed88f61a4e0a66685026c7719`, branch `feature/identity-platform`, selected identity worktree. Final source snapshot remained unchanged.

**Source-backed closure:**
- Preparation requires consuming mode, opaque native custody, exact frozen registration and a reservation-capable owner; generic/authenticated preparation still rejects OTP.
- Global claim, journal and frozen-plan matching precede OTP invocation. Live/terminal replay exits before reserve/rebind; changed intent conflicts.
- OTP runs last in the shared reservation transaction and skips counted-denial terminalization. Callback errors, panic, poisoned SQL and unfinished rows return through whole-transaction rollback.
- Takeover carries the exact predecessor generation; missing/mismatched native state cannot commit a partial takeover.
- Native binding requires the registered instance. Aliases lose authority on return/panic; command custody is cleared; existing context, SQL and result budgets remain controlling.
- Commit-attempt failures return Unknown with no ownership. `Begin()` and recovery release remain closed for OTP.

**Falsification checks:** traced foreign owners, nonconsuming plans, stale retained handles, changed fingerprints/plans, expired ownership, non-prior generations and failed callback prefixes through their dominating guards and terminal outcomes. No source-reachable violation found in this batch.

**Verification:** reviewer’s `git diff --check` passed. Parent-reported integration, race and vet results were supplied context, not independently executed by this reviewer. SQL fixtures test coordinator reservation ownership, not production OTP cryptography. Full consuming workflow, production adapter, current/finalize/recovery release and package/release gates remain pending; this approval covers only the development batch.

**Reviewed-path SHA256 snapshot:**
```text
cfb3253ae732643c21dcd314e8c956b8b83f250d7a0893bce2de12fe535087d5 postgres/CHANGELOG.md
b01b13f85cc4f63c7a39c2ca582a1e2c1990c71ade40af471c100870754ebc25 postgres/README.md
595f7b6cd151ec6bdeeb9c34de81ebb75254e6eddb016c8b85fb2a31fca2c8f7 postgres/coordinator.go
1361c672c7aa62418bafb2c72bbd4a193cb1756dd42357938136ad89bdbcad4e postgres/integration_test.go
b6de9039e950d1286d8743ca298708736f2e1cd6894ad00df83e7238d0bd1be9 postgres/otp_attempt.go
b9b1e7f8f59b4aae7a255660df5fcc4d4824c4a5964efdb0393aa7093503eb47 postgres/otp_reservation.go
05586944e3e33a45dbf4c90f378f0c9fff2eee50538e5af93bba635761a05b83 postgres/otp_reservation_test.go
a67300ca126724f0d33756b9589e4ae1978914a3b4ec79f0e952e90d2f7ce3b9 postgres/otp_reservation_integration_test.go
```

## Consuming reservation in flight

Baselineebae68e. Focused RED2eac52 proves missing typed reservation APIs;
GREENa79fca PASS0.270s covers native owner/mode admission, copied custody and
zero/serialization denial. Generic Begin and consuming recovery remain closed.
Real PostgreSQL focused race907f58 PASS1.721s via `go test .
-tags=integration -run '^TestPostgresIntegration$/^OTP_reservation$'
-count=1 -race`: six same-command concurrent retries, live replay, exact
prior-generation rebind without repeating reserve, foreign prior and expired
challenge refusal, error/panic/rows/poison rollback. Native participant remains
a SQL fixture, not production OTP comparison/policy. Disposable scratch/cache
identity-otp-reserve.c5onFg and tmpfs PostgreSQL identity-otp-reserve-c5onfg
are live for this bounded run. No review, commit or unit promotion yet.
Extended SQL fixturef24654 PASS1.540s adds competing-command winner preservation
and deferred commit failure classified Unknown. Intermediatef047b1 failed
because the competing-command fixture retained the first command's audit
binding; corrected to its own exact scope without production changes.
Current extended PostgreSQL reservation/OTP integration31113c PASS9.325s:
`go test . -tags=integration -run '^TestPostgresIntegration$/
^(OTP.*|coordinator_reservation|capability_reservation|risk_reservation|
captcha_use_reservation)$' -count=1 -race` (single uninterrupted regex).
Whole unit race53c42e PASS2.005s; vet6f2279 exited0; diff checkdff4eb PASS.
Current source frozen for one independent complete-diff development review.
Reviewer /root/otp_reservation_final_review, gpt-6.1-sol medium fork none,
source-only, no disposition yet. Task-owned container stopped/auto-removed;
scratch/cache root removed and absence verified17bbeb. Existing image preserved.

## Independent denial development review and commit

Native reviewer /root/otp_denial_final_review, gpt-6.1-sol medium fork none.
Report retained verbatim below. Parent independently matched every supplied
snapshot hash f818a0; committed snapshot unchanged as ebae68e (ffa4f4), source
clean76ec2b. ACK Core and line validation97436b PASS. Earlier line failures
c17e00/e7bcd0 were message-only; normal hook77bd1d rejected the attempted commit
and no source changed. No profile applies; effective impact unspecified.

**APPROVE — No findings.**

Reviewed the complete dirty patch, including all seven untracked files, against base/HEAD `70725ee40c10a759fb444468dad2e3f502f8deac` in `feature/identity-platform` at `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`. Single-reviewer, read-only structured pass using `requesting-code-review` and `testing-coverage`.

Source closure:
- Command identity, retained-key fingerprint and frozen attempt projection jointly reject changed-code/changed-attempt replay before owner invocation.
- Owner SQL, required denial audits and aborted journal terminalization share one transaction; callback errors/panics, unchanged decisions and incomplete rows roll back. Commit-attempt failure remains Unknown.
- Challenge comparison belongs to the exact registered native owner. Returned callback handles lose SQL and custody access.
- Recovery applies current authorization and complete historical validation before checking every original attempt field. Pending—including expired pending—provides no takeover or release authority.
- Canonical re-encoding rejects alternate, duplicate, truncated and trailing encodings; nil OTP preserves historical profiles. Generic admission/domain execution remain closed.

Verification: `git diff --check` passed. Go diagnostics were not rerun; supplied unit-race, focused PostgreSQL, vet and fuzz results remain parent-reported evidence, not independently executed results.

This approval covers the **development batch only**. Native production OTP cryptography/challenge policy, consuming reservation/current/finalize/release, full release gates, exact coverage/mutation and platform completion remain unverified/pending.

Reviewed snapshot SHA-256:
```text
ae19fd910e4a640f3e730e0b34f79486a0e111e2f0601efbd547c3a202fcb96e postgres/CHANGELOG.md
2ce2a652421d308cb948bd684bac35a8b04fccfd81d9a5027289e2c6dd48e074 postgres/README.md
662f24572bfa5d2720f2a8e69909458df9a5541ae0b069ff442cf5e6047da622 postgres/coordinator.go
2f8b72d6f8ebebc4a6d16fc793ea4492116e9ab214333191c16d41e505258a51 postgres/domain.go
5bde22f00e6f8fb2122515e128cd47f594e889b4476fe7c3aff643a22f52b0de postgres/feature_execution.go
aaaae69fe29190b16d95e5822f90a420989c2d1766beff6d8e37e045497b7806 postgres/integration_test.go
5e57242990ee4394ed0b4035b1bcc63d44c672cd32c6b44d2057bfb016ad43ab postgres/plan.go
bcfa961acda50285719a8fd400885e0665497a073e3c61aecfd60287242d64ad postgres/plan_feature.go
e756e62cd4b5354dce98b7640feafb9c19abfd2a3eefacf1b6e6853242ff3560 postgres/plan_feature_decode.go
0711ba119adafab374496394629cbe198565d9928ef12637f40b260843755ea7 postgres/query.go
9032d1c9aedeac3f5df25bbc657dd2235a6c39aa0eafdad54ac2009ba6ab7fe8 postgres/recovery.go
ffd6eadd9996ecb1f3358ddb540f2be3278bd7e0902d5ec9bbcdb9e91801a3f0 postgres/otp_attempt.go
bd06176e03332dfa210c1a4b7f515dc5cbe8e3cd970594e65715763b11c969d1 postgres/otp_attempt_integration_test.go
8c9f856f9f7b73c0f0126f6ffc0c8ad066e5ba2fe7c6c8b7bcf7beed0666a859 postgres/otp_attempt_test.go
e01ab02b8f997b1c174a705eb6d413a644b80903fc6f0ef1d0848d73113556ef postgres/otp_recovery.go
f17daab1b229024343e0618386e6562e5a8ca9111c2dc881d10e06c7e69285d3 postgres/otp_recovery_test.go
e93bac7fc27f807868263fa26898fb224e8a5e003ac615ee38b3ed5d399c334a postgres/plan_otp.go
b18e8b9e1109bf943734a4c1a6bf3466e0294b9a68baa8448bfcac208849deb9 postgres/plan_otp_test.go
```

## In-flight OTP counted-denial transaction

Continuation: internal incomplete Work regression230707 reproduced a nil-plan
panic; explicit plan guard fixes it. Combined OTP/CAPTCHA/counter history tests
cover canonical roundtrip, every truncation and trailing data; the counter
fixture initially selected a mismatching policy version8af960 and was corrected
to the existing CAPTCHA policy. Whole PostgreSQL unit race4090a9 PASS1.978s;
vet705d2d exited0 after these edits. New workspace setup7c3fc7 initially missed
the exact root pseudo-version mapping; added only in the owned go.work, not in
the releasable module. README/changelog describe the narrow denial protocol and
remaining consuming phases. Parser fuzz and final independent review pending.
Parser fuzz628d5f PASS43980 executions3.350s; recovery adverse-state test
a8324c PASS0.220s: pending never proves rollback, non-denial terminal profile
does not satisfy denial recovery, invalid original attempt performs no IO.
Current isolated reviewer /root/otp_denial_final_review, gpt-6.1-sol medium
fork none, inspecting complete dirty batch read-only. No disposition yet.
Final whole unit race after adverse-state test af4183 PASS2.053s.
Owned continuation scratch/cache identity-otp-final.fgQYQy removed and absence
verified. Reviewer confirms source-only inspection, no diagnostic resources.

Baseline70725ee; dirty implementation, not independently reviewed or committed.
Plan RED0cbeb7 / GREEN387eb7 and typed preparation REDdcf649 /
GREENdbd889 establish frozen attempt intent and owner custody. Integration
fixture compilation1ce300 failed on an audit enum; corrected without changing
production behavior. Real PostgreSQL focused race lane abaeec passed1.503s:
`go test . -tags=integration -run
'^TestPostgresIntegration$/^OTP_attempt_denial$' -count=1 -race`.
It proves coordinator-owned charge/audit/aborted-journal atomicity, six concurrent
same-command retries, changed-input rejection, retained-key replay, and rollback
on native failure/panic, retained rows, audit failure and deferred commit failure.
The SQL native participant is a fixture, not the production OTP adapter: actual
challenge comparison, scope/policy enforcement and consuming phases remain
unimplemented. Deferred commit failure returns Unknown, not rollback proof.
Task-owned PostgreSQL container identity-otp-attempt-kipbo6 uses a pre-existing
image, loopback55473 and tmpfs only. Scratch/cache root
/private/tmp/identity-otp-reservation.KIPbo6 remains owned for this bounded run.
Original-attempt code-free recovery REDf4839b proves missing API; implementation
focused OTP/query GREEN9cfe07 passed0.292s. Extended primary recovery integration
41e4aa failed because the fixture omitted constructor QueryPolicy; enabling the
existing public configuration fixed the fixture. GREENfa6619 passed1.584s for
the same focused PostgreSQL race command, now proving code-free original-attempt
recovery and changed-attempt denial without native re-comparison. Full release
gates and unit promotion remain pending.
Whole PostgreSQL unit lane `go test ./... -race -count=1`6bfd79 PASS1.903s;
`go vet ./...`df7bd0 exited0; `git diff --check`f54b9c exited0.
No independent current-batch review, commit, release coverage or mutation claim.
Bounded-run container stopped and automatically removed; exact scratch/cache
root removed and absence verified. Pre-existing PostgreSQL image preserved.

## PostgreSQL opaque command intent, development batch

Baseline19cdb33 consumes committed root19cdb33 and authentication0f9f7ea.
RED28f2f5 proved missing secret intent and unsafe generic preparation; focused
GREENf18b94 passed0.256s after separate keyed payload and central denial.
Intermediate c2f444 was an unpublished-module workspace setup failure, not
behavioral RED; exact new root pin was then explicitly mapped in owned go.work.
Whole PostgreSQL unit lane `go test ./... -race -coverprofile=<owned profile>
-count=1` c58932 passed2.495s, reported51.1%; `go vet ./...`73c4e6 exited0.
Integration build-tag lane, database behavior, complete coverage/mutation and
release/composed OTP remain pending; no database resources created for this
pre-IO codec/admission change. Exact independent HMAC framing, absent code
bytes, changed code, unchanged safe scope, rotation and retained-key replay
have assertions. No one-time execution was enabled. Exact raw e5413c is
3425/6699, codec191/201 and coordinator84/214. Independent review APPROVE;
reviewed snapshot committed unchanged locally as70725ee (9fd4e0), clean source.
ACK Core/line validation603447 passed. Native report follows.

Native reviewer /root/postgres_code_intent_review, gpt-6.1-sol medium fork none:

**APPROVE — No findings.**

Reviewed base/HEAD `19cdb33a1e8ead25c5f977f8658ccbebb360a2fa` plus all six dirty files, including the untracked test, on `feature/identity-platform`.

- `codec.go:119–127` uses auth-owned exact `CommandDigest()` framing, adds only its digest to private canonical intent, preserves absent-code encoding and safe scope, and retains immutable source for stored-key recomputation.
- `coordinator.go:179` closes ordinary and authenticated preparation/reservation before feature admission or I/O. Substituting code custody into previously admitted plain execution also fails canonical equality in `ApplyFeature()`.
- Tests independently frame the expected HMAC and distinguish absent, changed, rotated and retained-key intent. No additional journal or transaction authority is introduced.
- Independently inspected coverage: package **3425/6699**, codec **191/201**, coordinator **84/214**; `git diff --check` passed. Reported RED/GREEN, race lane and vet execution remain parent-reported, not independently rerun.

Snapshot SHA-256, independently observed and unchanged across inspection:
```text
codec.go           9ba222193be2016c74175d9665fdc9a664b1a38c978e307f02a9001494ae35dd
coordinator.go     d28e5b173809419af9947ede356549732d442c907ebb96f3334651f70deb08bf
codec_code_test.go a06bc70cff1d92ae5ec5e613639ffbbbab851a1e7862d48f8e312cd6d3e4a8b3
go.mod             2b95850945c5cbb40f03144de7fbd689868dac6cc91cd71ab02b2a9f0932ab46
README.md          988790611ef43074b7ab55db86f0677946c4887758241c49c64be902a0accbb6
CHANGELOG.md       38a236a5be4e8bea08585f757d23312710651ac4502f932306763f50bd9e9321
```

Approval covers this local codec/fail-closed batch only. Native counted OTP phases, database/composed execution, exact coverage, mutation, release gates and clean remote-consumer resolution remain unverified; this does not certify complete OTP or adapter delivery.

## Native root custody final review, verbatim

## Disposition: APPROVE

No findings within the bounded root-custody prerequisite.

**Reviewer:** `/root/identity_code_custody_review`; isolated, single-reviewer, read-only pass using `$requesting-code-review` and applicable `$testing-coverage` guidance. No writes, tests, commits, external mutations or delegation performed.

### Exact boundary

Repository: `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`
Branch: `feature/identity-platform`
Base and HEAD: `15ba2a98ff3fb87c4288c2adeff38e1b7c2db2af`

Reviewed every tracked dirty hunk and both untracked files: `feature_code.go`, `feature_code_test.go`, `feature_command.go`, `feature_command_internal_test.go`, `go.mod`, `README.md`, `CHANGELOG.md`.

Consumed authentication repository independently confirmed clean at `0f9f7eac1f1dcb2dd8ac91218888df294a4e9bd0`. Reviewed its opaque representation, constructor and both digest profiles.

### Covered causal traces

- **Construction and denial:** `feature_code.go:21–30` delegates complete structural validation/copying to `NewFeatureCommand()`, rejects zero opaque code through the authentication-owned bounded digest operation, and returns an empty command on either failure. The public validation key produces no retained verifier or authority.
- **Secret custody:** `feature_command.go:76–78` retains code outside `FeatureCommandSpec` and `FeatureValue`. Authentication’s private closure representation supplies immutable exact bytes without a raw accessor. Root imports authentication, not OTP.
- **Projection isolation:** `feature_command.go:119–148` returns fresh structural copies, including field slices and byte values. Source challenge: mutate original binding, field bytes and returned projection after construction; the retained command remains unchanged. Existing structural tests independently exercise byte copying.
- **Variant discrimination:** `feature_code.go:38–45` denies ordinary, zero and kernel commands, then delegates mixed-union exclusion to `Feature()`. The internal mixed fixture exercises that denial. Existing ordinary constructors leave custody nil and retain their previous projections.
- **Diagnostics:** unchanged `command.go:153–160` universally redacts formatting and rejects JSON before encoding retained fields. The secret is also absent from the public feature projection by construction.
- **Concurrency and bounds:** after successful construction, neither state nor opaque code is mutated. Structural limits remain 128 fields/256 KiB; authentication limits codes to 4–128 visible ASCII bytes. No external I/O, goroutines, durable writes or caller-injectable callbacks are introduced.
- **Test oracles:** custody digest comparison rejects code substitution/loss, not independently incorrect cryptographic framing; upstream authentication tests independently assert literal transaction framing. Root tests separately assert malformed/zero denial, projection isolation, redaction and concurrent borrowing.

### Verification boundary

Independently inspected the raw coverage profile: **2,908/3,199** root statements, **12/12** in `feature_code.go`, **57/57** in `feature_command.go`. `git diff --check` passed.

Execution results remain **parent-reported native evidence**, not reviewer reruns: focused RED `4ac998`; GREEN `64c9a8`; root race/coverage `fbe8c1`; vet `a60ede`; OTP consumer race `dd07e7`; shared structural fuzz `733b49`. Inspected the disposable runner/workspace configuration binding authentication to the exact local source.

### Explicit limitations

Approval permits the **local prerequisite batch**, not durable OTP execution, release readiness or platform verification.

Existing PostgreSQL encoding still processes only `Feature()` metadata (`postgres/codec.go:109–115`, `postgres/codec_feature.go:9`). It does **not** fingerprint the new opaque submission. Commands differing only in code therefore remain indistinguishable to that old codec. The separately scoped adapter batch must bind custody and reject unsupported one-time profiles before admission; this review does not certify old coordinator support.

Whole-root exact coverage, mutation, publication/clean-consumer resolution and remaining release/composition gates remain pending.

### Independently observed SHA-256 snapshot

```text
feature_code.go
9201055b621eb39231c60c0e70e953815743217c18537d0e7895ddb96a316a9c
feature_code_test.go
ff0231ee334d833f766ca18c75d8ec3a102998463008daadeb88ad1e3883b02c
feature_command.go
2dca4e0ffded1424a7a2117e6f66d6054fde59a8ed6557ca4c98369bcbe0eac5
feature_command_internal_test.go
927202eaa9ec14ac0d439b9f6ce0ae862de2bfc945726bc14aadcfea26d0694d
go.mod
ab76c79dd429d9b10187dce31a73e932862317dae4b709c2dab28ff62e4f4b21
README.md
e6303acf432faea91fa31da014271c5fdda25b746a2f422e9ef9b0402bf19dc0
CHANGELOG.md
4f45cffdb5e8b1d7d45c6c40c544fc1a861898ffa7a7c036317eb89a50242dec
```

Hashes were unchanged on the final reinspection.

Reviewed snapshot committed unchanged locally as19cdb33. ACK Core and line
validation passed2dbf90; source clean43b831. No release/unit promotion.

## Root typed code custody, development batch

Baseline15ba2a9 consumes committed authentication0f9f7ea. Focused RED4ac998
failed on missing custody APIs; GREEN64c9a8 passed. Whole root race fbe8c1
passed1.653s, exact raw coverage279fc9:2908/3199 overall, feature_code12/12
and feature_command57/57. Whole-root release coverage remains incomplete.
Vet a60ede exited0; direct OTP consumer race dd07e7 passed1.401s. Commands:
`go test ./... -race -coverprofile=<owned profile> -count=1`, `go vet ./...`,
and `go test ./... -race -count=1` in OTP. Task-owned workspace/cache:
/private/tmp/identity-otp-custody.n0Ecyf; explicit auth/authorization/capability
source mappings, no releasable replacements. No database resources created.
Opaque code survives immutable command copies, scalar projection remains
code-free, malformed/mixed custody denies and command serialization rejects.
Independent root final review approved above. PostgreSQL encoding is now
committed above; typed OTP
phases, native adapter and release/mutation/composition remain unfinished.
Existing shared structural-boundary fuzz733b49 passed152078 executions4.355s:
`go test . -run '^$' -fuzz '^FuzzFeatureCommand$' -fuzztime=3s -parallel=2`.
This checks structural constructor behavior, not a new opaque-secret parser.

## Opaque coordinator code custody prerequisite

Authentication baseline303520b. Focused REDe90183 failed for missing
OneTimeCode.CommandDigest; GREEN767cf4 passed0.272s. Exact framing follows
TRANSACTION_CONTRACT.md's existing identity-command-secret-v1 byte profile,
not a changed protocol or replacement challenge-storage digest. Final root
module/tests/examples racec5acef passed (root2.367s); exact raw174c76 covers
root671/671 and each of eight root-module subpackages at100% without rounding.
Owned bounded-run scratch /private/tmp/identity-auth-command.gctI6L supplies
Go1.27.1 caches; cleanup required at closure. Native coordinator dispatch,
OTP PostgreSQL and full release/mutation gates remain pending.

Vet9fa57b exited0. Fuzz26118d PASS155174 executions4.263s against new
CommandDigest hostile UTF-8/path/code inputs and an independent framing oracle.
Final five-path source hashesa53e91 remain the review snapshot. Catalog adds
only the primitive method required by the pre-existing transaction profile;
required-contract digest49520c is7530318a6e41279264a7098d6d13adf6a10c0fa6c561d61cb8893698b6f50e53.

Committed authentication0f9f7eac1f1dcb2dd8ac91218888df294a4e9bd0 (b6fe0e),
ACK Core/line validation passed, clean sourcefec2dc. Goal pin78696c updated
in GOAL_MANIFEST. Aggregate generation513003 failed: core Work.StageDeliveryFor
references undeclared DeliveryContributor and DeliveryEnvelope; generator wrote
no aggregate output. PUBLIC_CONTRACTS refresh remains pending, not certified.
Direct OTP consumer with new auth0f9f7ea and unchanged root15ba2a9/session,
authorization60ed4b3/capability2e30f8b PASSrace9625771.273s. First workspace
attemptfdc927 was a setup failure, corrected by mapping exact declared historical
root/session versions from native2d3eae; no permanent replacements or module edits.

Native independent report, /root/auth_command_secret_review_303,
gpt-6.1-sol medium fork none:

> APPROVE — No findings.
>
> **Reviewed boundary:** `feature/identity-platform`, base/HEAD `303520b8d5af937e4e38ec8cb3ee8ff6f9813230`, plus the complete five-file dirty batch. All five SHA-256 hashes match the supplied snapshot. Single-reviewer isolated structured pass; no tests, writes, mutations, or delegation.
>
> **Source-backed closure:**
> - `command_digest.go:30–47` matches `TRANSACTION_CONTRACT.md` and the updated primitive requirement exactly: domain without NUL, uint32 big-endian UTF-8 path byte length, exact path, uint64 big-endian secret byte length, exact code.
> - `command_digest.go:31–32` rejects zero code/key, empty/oversized/invalid-UTF-8 paths with zero digest and the stable redacted sentinel. Constructor custody in `proof.go:43–64` bounds code bytes and prevents external access or construction.
> - Bounded adversarial challenges: 4096-byte multibyte paths preserve byte lengths; a NUL suffix remains distinct; composed/decomposed Unicode remains distinct; zero opaque values cannot reach the secret closure. No normalization or input retention occurs.
> - Temporary mutable code bytes clear on return; cryptographic/immutable-storage erasure limitations are documented. Calls share no mutable state, perform no I/O, and grant no authentication, consumption, or transaction authority.
> - Existing `KeyedDigest` implementation and framing are unchanged. Coordinator reference `postgres/codec.go:238–254` uses equivalent framing, while schema-bound validation remains explicitly consumer-owned.
> - Tests use literal framing and separately assembled HMAC oracles, cover independent key/code/path/domain changes, invalid partitions, size endpoints, concurrency, and hostile fuzz inputs. Documentation, example, and changelog accurately preserve the prerequisite-only boundary.
>
> **Verification boundary:** Independently inspected `/private/tmp/identity-auth-command.gctI6L/coverage.out`: new method **16/16**, root package **671/671**, all eight subpackages exact **100%**. Race-suite, vet, fuzz, and red/green execution outcomes are supplied parent evidence, not independently rerun here.
>
> This approves the coherent local development batch, **not** final package/platform verification. Aggregate release/mutation gates, generated control-catalog refresh, native coordinator dispatch, and OTP/PostgreSQL integration remain outside this review’s completion claim.

## Command-bound submission development

Baseline08eaae7. Native focused REDac5341 exited1 for missing PrepareForCommand
and AttemptCommand.CommandBinding methods. One bounded red/green/verification
run owns /private/tmp/identity-otp-command.Axe2HK build/module caches and explicit
local version mappings; cleanup is required at run closure. This extends native
command correlation, not coordinator authority or durable dispatch.

Native focused GREEN676dbe passed. Expanded independent literal HMAC framing,
command-field separation, foreign-scope rejection, explicit anonymous namespace
and immutable recovery tests pass in whole OTP race46797e (1.293s).
Raw coverage6747ef is exactly344/344 statements; vet854c3a exited0.
Legacy exact fingerprint oracle remains unchanged and passing. Complete final
five-path independent review follows; durable coordinator/OTP PostgreSQL and
release/mutation gates remain pending. No whole-unit promotion.

Committed source15ba2a98ff3fb87c4288c2adeff38e1b7c2db2af (nativecc57eb), clean
source worktree. ACK Core and line validationfd3bc9 passed; no profile selected.
All five source hashes6747ef stayed unchanged through review and commit.

Native independent report, /root/otp_command_binding_review_08e,
gpt-6.1-sol medium fork none:

> **APPROVE — No findings.**
>
> Reviewed the complete five-file dirty diff against `08eaae7c6f836196730d3b3d279718e0fcc07978`; all supplied file hashes remained unchanged.
>
> Source traces confirm:
> - Root validation and exact tenant/subject/required-target matching precede preparation.
> - The structural digest covers every command-binding and operation-descriptor field; the outer fingerprint remains keyed and domain-separated.
> - Recovery preserves command and attempt identities without code.
> - Projections retain immutable scalar state; legacy `Prepare` preserves its exact encoding.
> - New API consumers are tests only; no execution authority or durable side effects are introduced.
>
> Inspected independent literal fingerprint oracles, foreign-scope rejection, namespace separation, recovery, and legacy tests. Directly inspected coverage: **344/344 statements**. Parent reports final race tests and vet passed; I did not rerun them. `git diff --check` passed.
>
> Approval covers this development batch only. Durable dispatch, native operation-purpose/provenance enforcement, coordinator fingerprints, full-unit release gates, and deployed behavior remain unverified and outside this batch.

## Locked challenge policy development

Baseline8657b13. Focused `go test . -run '^TestChallenge' -count=1` first hit
unpublished local module resolution (setup failure, not RED). Task-owned
version mappings corrected the local composition without changing go.mod.
Rerun7d4d80 exited1 for absent ChallengeSpec/NewAttemptPolicy/NewChallenge/
EvaluateAttempt symbols: expected missing-behavior RED. No durable attempt,
reservation or authentication claim follows from this pure policy batch.
Focused GREEN9e5f35 passed0.229s. Whole OTP/tests/examples race3935b2 passed
1.237s with100.0% reported statement coverage; exact block accounting follows.
Immutable snapshots and evaluation are pure adapter inputs: all primary locks,
authorization/provenance, replay/CAS/reservation/finalization and composed
workflow evidence remain unfinished. No whole-unit promotion.
Native exact raw coverage2b1a76 is310/310 statements; native vet413e0e exited0.
Added direct matching-maximum-version overflow regression and cross-purpose
denial assertion after the first green; final affected verification follows.
Final OTP/tests/examples race310b47 PASS1.285s, raw6d4828 exact310/310;
FuzzChallengeSnapshotBounds5ee697 PASS166014 executions3.299s. All four source
hashes unchanged through final diffcheck. Isolated complete-diff review pending.
Initial native reviewer /root/otp_challenge_review_865 returned APPROVE for the
four-path snapshot above (not final evidence after later edits). Primary then
found the existing CodeProfile.Attempts contract requires TOTAL admitted
verification, not merely wrong submissions. New runtime regressionea1755 failed
for successful checks bypassing that budget. Policy now retains UsedAttempts
separately from FailedAttempts/Checks: correct final Check exhausts without
wrong-code lockout, correct final Consume may reserve. Reserved/finalized/
released snapshots require a counted correct consuming attempt. Focused final
and complete final-diff re-review are required before commit; prior review does
not approve these edits. This is an original policy acceptance correction, not
new release/evidence machinery.
Corrected focusedf82ca7 passed0.228s. Corrected whole OTP/tests/examples race
d00fe4 passed1.332s,100.0% reported coverage. Final source hashes9c61ec bind
the corrected four-path snapshot; exact accounting, vet/fuzz and native final
review follow before commit.
Corrected exact rawdd51fb313/313; vetca185e exit0; fuzze3ea49 PASS162280
executions3.287s. Final reviewer identified one stale wrong-code-vs-total-budget
API comment, corrected with no behavior changes; finalracef00f58 passed1.146s.
Four-path final hashes340b27:
```text
9c3060f37723cb67d6e0392d8286593022558fdc09c5e6ca2e5d6edcc07da5a2 otp/challenge.go
e1d44fb054be4e329869164a218090afae32a654a2dac1d2cab81b35f90c52b5 otp/challenge_test.go
d6d028b92b1d023bfd277c059326a79bc9a90bca5b0031c393f8770ac0e4df45 otp/README.md
6689514739ead5f549f2b24f109b5963b1c1c7981d9dcec5ccb29d15625a0f23 otp/CHANGELOG.md
```
Initial approval was explicitly withdrawn by the native reviewer for its missed
original total-budget requirement. Final native disposition remains required.
Final native reviewer approval below applies to exactly hashes340b27; normal
explicit staging matched those bytes, secret-shaped scan clear, Core/message-line
validation passed. Committed locally08eaae7c6f836196730d3b3d279718e0fcc07978;
source worktree inspected clean. Final comment-adjusted vet831836 exited0,
fuzz1a0680 passed170102 executions3.196s; raw809c4b exact313/313. No push,
publication, unit promotion or composed OTP claim.

## Native challenge policy final review, verbatim

**APPROVE — final development snapshot only. No findings.**

Native reviewer `/root/otp_challenge_review_865`; read-only complete-final-diff review using `$requesting-code-review` and `$testing-coverage`, without delegation, edits, code execution or external mutation.

**Boundary:** same worktree and branch `feature/identity-platform`; base/committed HEAD `8657b130fd481835f73c19f0e4691f58aa68944f`; complete four-path dirty snapshot, including untracked production/tests. Independently verified final hashes:
- `challenge.go`: `9c3060f37723cb67d6e0392d8286593022558fdc09c5e6ca2e5d6edcc07da5a2`
- `challenge_test.go`: `e1d44fb054be4e329869164a218090afae32a654a2dac1d2cab81b35f90c52b5`
- README: `d6d028b92b1d023bfd277c059326a79bc9a90bca5b0031c393f8770ac0e4df45`
- CHANGELOG: `6689514739ead5f549f2b24f109b5963b1c1c7981d9dcec5ccb29d15625a0f23`

**Correction:** the previous approval is withdrawn for overlooking the existing profile’s **total verification-attempt** requirement. The superseded wrong-only snapshot did not satisfy that contract.

**Final causal/oracle closure:**
- `challenge.go:112–120` validates bounded total/failed/check counters, remaining Issued allowance, exhausted totals and consuming-state counter prerequisites.
- `challenge.go:235–250` counts every admitted comparison once, including successful Check and Consume. Correct final Check becomes Exhausted without wrong-code lockout; correct final Consume becomes Reserved.
- The two-successful-check sequence independently asserts total accounting and subsequent Consume denial. The smaller check-cap fixture preserves one final consuming allowance. Wrong Check/Consume sequencing separately proves failure accounting and exhaustion lockout.
- Complete encoded scope, authorization/remember binding, exact versions, unavailable-key denial, exclusive expiry, immutable projections and terminal/reserved denial remain intact.
- The stale exhaustion comment identified during this pass is corrected; final API comments, README and changelog consistently describe total accounting.
- Symbol fan-out still finds only new test consumers, no production caller or adapter. No existing API/dependency changes or new persistence/retry authority.

**Verification:** not independently executed. Parent-reported regression RED `ea1755`, focused GREEN `f82ca7`, final whole OTP/tests/examples race PASS `f00f58` (1.146s), vet PASS `ca185e`, fuzz PASS `e3ea49` (162,280 executions). Independently inspected raw statement accounting: **313/313**; final `git diff --check` passed.

**Unfinished boundaries:** current primary authorization/provenance; durable replay, CAS, accounting and replacement lockout; reservation/finalization and ambiguous-commit recovery; owning workflow/session effects and service normalization; mutation, release, remote clean-consumer and composed journeys. Approval supports this coherent local development batch—not unit verification or release/platform completion.

Source baseline0d4d7bb; unit identity/otp owner/root. Initial standard-library
profile/generation batch consumes no unresolved dependencies. Missing public
API RED0ce285 (undefined CodeProfileSpec/Purpose); first complete package GREEN
7c2de7,0.188s. Durable issuance/attempt/store/delivery/session workflows and
primitive code verification boundary remain unimplemented. No unit promoted.

Bounded acceptance: complete closed-purpose policy, minimum million-code space,
unique visible alphabet, unbiased rejected tail, finite draws, cancellation,
ephemeral callback custody cleared on success/error/panic, error redaction and
concurrent public generation. Added exhaustive 256-byte bucket and draw tests,
hostile profile fuzz seeds and a compiling external-package example. Final
focused checks and isolated complete-final-diff review pending.

Complete initial profile/generator race PASSc47aa2,1.216s; statement coverage
98.3% (not a release pass). Ephemeral reader/draw/cancellation tests and external
example included. Purpose retention now copied alongside the alphabet; final
verification must be refreshed for that changed input. No unrelated package gate
rerun or unit promotion.

Final race/coverage PASSaa0bb1,1.268s: `go test . -race -count=1
-coverprofile=<owned temporary profile>`; exact profile must be inspected before
claiming unrounded coverage. All accepted byte buckets have equal multiplicity;
rejected tails exhaust exactly192 draws without callback. Redundant pre-loop
cancellation check removed; the per-draw check rejects pre-cancelled calls before
any entropy read. Exact OTP module/catalog projection PASS398919. Fuzz/vet live
handles36326/2204; no completion claim until terminal results are captured.

Fuzz terminal PASS93f1b2,3.292s,533745 executions: `go test . -run '^$'
-fuzz '^FuzzCodeProfile$' -fuzztime=3s -parallel=2`. Vet terminal PASSf899b7
exit0: `go vet .`. Native raw coverage checked626c87: exactly58/58 statements,
not rounded or aggregated. This does not certify mutation/release/workflow
completion. Independent complete-final-diff review/root/otp_generation_review_0d4
running against source0d4d7bb plus nine exact new/changed paths. No source changes
during review; all owned command caches removed on terminal exit.

## Native final independent review, verbatim

**APPROVE — No findings.**

Native principal: `/root/otp_generation_review_0d4`. Single independent, read-only reviewer using `$requesting-code-review` and test-oracle guidance; no writes, test execution, delegation, or external submission.

**Boundary:** `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`, branch `feature/identity-platform`, base=HEAD=`0d4d7bb80e541a7e882a3038be92d20b6a68bf23`, plus the complete nine-path dirty scope. Final hashes remained unchanged:

| Path | SHA-256 |
|---|---|
| modules.json | `98f0f32646264209f673ed779348451837d7b415296a050a65ef347330eaeb84` |
| packages.json | `444008bcc4a71c796d51d14a27bb2adaa6385e155fed1c683f1a53b2089680ee` |
| otp/go.mod | `654ac354d6724108a9bd51e42ddf2c102ef00a15b52541c524708db6874bd91a` |
| otp/profile.go | `0ae9b0ce25f5ecc17c33eda0b560c5da56ac22e48c91ea4e2be8a9f627e659bf` |
| otp/generation.go | `9ab51d54183a5f78d4b9c658f7daec085e8958dbdb7000b5e2ee094f473d48f8` |
| otp/generation_test.go | `b715e80e5eba23a7bb56a9bf4b9a07b875fed54fee5e0d3022c55a6381368233` |
| otp/example_test.go | `b8ff05d52645e81623ff8f45efab3dd5514a19ddb187dace01e55db835386f7d` |
| otp/README.md | `d42cf54a27d7baab9abbbbec5005c803d8bdf01f9b2f5561515edbf0d0feee88` |
| otp/CHANGELOG.md | `e271ff8904cfbcf2ef4bd85239646c8ab275fa442dc0170c06bf74c4a2a0274d` |

**Closed causal boundaries**
- Constructor validates all policy inputs before bounded cloning/allocation. Code-space arithmetic cannot overflow before its threshold exit.
- Rejection sampling admits equally sized source buckets; exhausted entropy never invokes the consumer. Maximum reads are 4,096, code storage 128 bytes.
- Success, callback failure, panic, cancellation, and entropy failure preserve buffer cleanup; returned callback errors are redacted.
- Profile copies retain only immutable policy; concurrent calls own separate buffers and contexts.
- Exact-symbol searches found no pre-existing or external repository callers. Module/package registration matches the standalone topology; no dependencies or replacements were introduced.
- Documentation and changelog explicitly deny durable issuance or authentication authority.

**Verification:** parent-supplied RED `0ce285`; race PASS `aa0bb1`; raw coverage **58/58** `626c87`; fuzz PASS `93f1b2` (533,745 executions); vet PASS `f899b7`; manifest projection PASS `398919`. Reviewer independently observed `git diff --check` PASS.

**Retained limitations:** fuzz checks accepted-profile projection and generation, but does not independently classify every invalid generated profile. Deterministic bounds tests supply the current rejection oracle. Approval covers this development batch only—not complete OTP, mutation/release verification, durable issue/attempt/digest/storage, delivery, or session workflows.

Reviewed unchanged snapshot committed locally as6ebfac9. No runtime result is
relabelled as rerun at the new Git revision. All nine complete inputs match the
reviewed hashes above. Counts remain5/67 integrated and1/67 verified. Next work:
complete scoped keyed-digest/attempt identity boundary, then authoritative durable
issue/reserve/attempt/finalize/recovery and owning workflows.

## Opaque code verification seam

Authentication source3467aab has no raw-code accessor or usable hashing seam;
OTP must not copy its private representation. Added pure OneTimeCode/RecoveryCode
KeyedDigest boundary, with type-separated HMAC-SHA256 and length-delimited exact
scope/code; no raw bytes leave authentication and no authority is minted.
Server-owned workflow encoding/key versions remain OTP/MFA responsibility.
Missing methods REDaef87c; focused external-consumer race GREEN9e1e08,1.226s.
Initial shared module cache checksum mismatch was isolated to that existing
cache: owned empty GOMODCACHE downloaded public proxy bytes matching unchanged
committed go.sum, without checksum substitution or shared-cache mutation.
Full root/direct-consumer checks and final independent review remain pending.

Authentication root and all eight existing direct consumer packages race
PASSfbad8d (root4.147s); `go test ./... -race -count=1`, vet PASS9c7dba exit0:
`go vet ./...`. These exercise existing formatting/non-constructibility and
credential/instrumentation/HTTP consumers; no released API was removed or changed.
Required typed contract adds only the two methods, digest78bfcc8a. Scope framing
is explicit in the existing primitive goal and both canonical catalogs; start
API baseline pin is preserved. Final hostile-input fuzz checks/review pending.

Final root race/coverage PASS7366b0,4.164s (`go test . -race -count=1
-coverprofile=<owned profile>`); fuzz PASS927795,3.322s/165534 executions:
`go test . -run '^$' -fuzz '^FuzzOpaqueCodeKeyedDigest$' -fuzztime=3s -parallel=2`.
Only test additions followed the all-consumer result; unchanged direct-consumer
inputs remain reusable. Catalog's stale Go1.26.6 declarations now agree with
all four existing Go1.27.0 module manifests; no module dependency changed.

Raw native coverage inspected95ac5e: exactly655/655 root statements. Affected
typed-contract projection and source/goal pins PASS412680. Full release and
mutation gates, durable OTP attempts/consumption and composed login remain
pending; no unit promotion.

## Native opaque-code final review, verbatim

**APPROVE — No findings.**

Principal: `/root/opaque_code_digest_review_346`; isolated single-reviewer pass using `requesting-code-review`, with test-oracle and verification guidance. No code execution, writes, delegation, or external mutations.

Reviewed both `feature/identity-platform` dirty snapshots:
- Authentication base/head: `3467aab01d0c1ecb6501e14f371c8e34f6739fb2`.
- Control base/head: `6ef2943f398851182c76f198faf7b4760d829550`.

Every scoped hunk and both new files were inspected. Constructor ownership guarantees bounded immutable code storage; the new methods preserve non-disclosure, closed type domains, injective versioned framing, bounded work, redacted errors, and call-local state. The literal HMAC test oracles independently check both domains. Invalid bounds, zero values, changed key/scope/code, mutable-input isolation, concurrency, and hostile-input fuzz partitions are covered. Existing consumers do not acquire new behavior; OTP generation remains a separate prerequisite rather than completed digest/attempt orchestration.

Adversarial challenge: equal code/key/binding across the two opaque types cannot share their framed message; embedded NUL or appended binding bytes remain distinct through explicit lengths. Maximum constructor code lengths fit the uint16 field.

`git diff --check` passed for both scoped tracked diffs. Runtime results remain **parent-supplied native evidence**, not independently executed or re-certified by this reviewer.

### Reviewed SHA-256 fingerprints

Authentication paths:
```text
7098d48c540318abc54e2cd5889dcac6519a980f79bbac37f25e625ed1aedc85  code_digest.go
551d7672e482086a6182f5816ba6e8e32ca1bc01b0c547c9c9cb479f2a58c0a3  code_digest_test.go
6e4d5e9a277491885e378c21fb97917a4c56b777e04f447c42c8087632348b09  CHANGELOG.md
432ed70dd3c538d970fb175275cb67eb64605b249c4cd6f6e36ecd5a47404ccd  docs/identity-proofs.md
ea93d40614a3544f23fbec3ff1428c0cad20460068a9d8f7b29c20cf29ac447c  modules.json
ae6cb3b691b5b26571282af018b523a7be7c588fe51258fdda9955a70b4adbae  README.md
```

Control paths beneath `.ai/identity-platform/`:
```text
142c23be9b95c7c694a987c8dff0b80e116cd5cde76150cdd8dd7e0cb8608db8  GOAL_MANIFEST.json
bf8809585fe3313d23660200376bc7aed0d5a2141989a205f58c4453121993c6  PUBLIC_CONTRACTS.json
569859d47129478d6cde209eaa3253987496e70926d0d6690a3fbceeddd6f84a  fragments/public_contracts_meta.json
c0a9512e4c54f762df8c66932d920f36a10d3b321b3d22c30a51fda95376451c  goals/primitive-authentication-identity-contracts.md
```

Approval covers this coherent **development batch only**. Mutation, full release gates, complete OTP/MFA scope encoding and key rotation, durable attempts/consumption/recovery, composition, and full platform verification remain outstanding—not waived or certified.

Unchanged reviewed authentication snapshot committed locally as303520b. Runtime
evidence retains its original execution inputs/revision; it was not rerun merely
for this commit. No checksum or dependency substitutions. Full units unchanged.

## OTP scoped keyed digest implementation

Source6ebfac9 plus six OTP paths now implements immutable complete scope,
bounded protected destination input and retained-key constant-time comparison.
Consumes committed authentication303520b, exact pseudo-version pin, through
task-owned explicit workspace wiring only; no permanent module replacements.
Initial proxy404 for the unpublished pin was a setup error, not RED. Explicit
task-owned workspace version mapping resolved it without changing production
dependency bytes. Intended missing-api RED31f671; focused GREENc9c3350.220s.
Final all OTP tests/example race PASS5927c7,1.222s, exact125/125 native raw
statements c6af3a. Vet PASSc2e47e exit0; fuzz PASS99af70,3.273s363486 executions:
`go test . -run '^$' -fuzz '^FuzzDigestScopeIsolation$' -fuzztime=3s -parallel=2`.
Final race command `go test ./... -race -count=1 -coverprofile=<owned profile>`.
All commands use per-command disposable caches; no database or shared service.
Independent full six-path review/root/otp_digest_review_6eb in progress.
Matching remains non-authoritative; durable counted attempts, replay, scope/
authorization construction, issue/reserve/finalize/recover and login unfinished.
Full unit/release/mutation/clean remote-consumer gates remain pending.

## Native scoped-digest final review, verbatim

**APPROVE — development batch only. No findings.**

Reviewer: `/root/otp_digest_review_6eb`; isolated single-reviewer pass using `$requesting-code-review` and applicable `$testing-coverage` guidance. No delegation, file changes, Go execution, or external mutations.

### Reviewed boundary

Repository: `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`
Branch: `feature/identity-platform`
Base/committed head: `6ebfac9497e92b91f4784b9fce0754f8ed24b6d0`
Reviewed head: that commit plus the complete seven-path dirty snapshot below.

Consumed authentication source is clean at `303520b8d5af937e4e38ec8cb3ee8ff6f9813230`.

### Source-backed conclusions

- Exact length-delimited framing binds all scope fields and retained key version; bounded inputs remain below authentication’s 4096-byte binding limit.
- Digest derivation uses the opaque authentication code contract and HMAC-SHA-256; comparison uses fixed-size constant-time equality.
- Wrong code, changed scope, and retired keys fail closed. Changing the current version does not replace the recorded verification key.
- Destination custody is explicitly raw and bounded; canonicalization and authorization remain workflow responsibilities. Formatting redacts and JSON/text serialization denies sensitive values.
- Construction copies caller key storage; subsequent caller-map mutation cannot affect comparison. Operations introduce no I/O, callbacks, goroutines, or shared mutation.
- Existing profile/generation APIs remain unchanged. New symbols have no existing production callers; only the new tests/example consume them.
- The independent literal framing oracle exercises the consumed authentication implementation rather than calling the production scope encoder.
- Documentation and registration accurately describe a primitive development checkpoint, not completed OTP authentication or consumption.

Bounded falsification challenges—scope-field substitution, ambiguous concatenation, caller key mutation, retained-key rotation, and removed-key lookup—resolve safely through the reviewed guards/framing.

### Verification boundaries

Parent-supplied results: intended missing-API RED; focused GREEN; OTP tests/examples/race PASS; exact coverage `125/125`; fuzz PASS; vet PASS; affected registration structural PASS. I inspected the tests and source but did **not independently execute** those checks. My read-only `git diff --check` passed.

Focused untested risk: same-version key-custody replacement lacks an explicit regression case; source fails comparison under a changed key, and the public contract requires an explicit migration decision.

Mutation, release gates, clean remote-consumer qualification, and composed durable OTP workflows remain pending. This approval supports a coherent local development commit only; it does not establish unit verification, release readiness, or platform completion.

### Frozen SHA-256 snapshot

```text
otp/digest.go              91b970f642d1fcce7cc02111923ae298269b023c00a4dcb72432939782fa10cb
otp/digest_test.go         21600c21cfaf97455b7f0aed6058de7845d5429c48d9c1f1da06ba7d63fb485e
otp/digest_example_test.go b20df356b9063705c0b7413cf902c0721f41212bcd46ffa7bce161e8b7899b85
otp/go.mod                5175fe29703a9fc5d4c1949acbd768f44037f6e2efe7c0fe0c8c5393e1ccd30b
otp/README.md             fac6aa0161321a7c9d7f1c15358b77d042d44d2c565d625c1692995ac25c8027
otp/CHANGELOG.md          5bb5944fdc348f0ea45711dc705ccb5720bfb7bdaf4460ad249930f217d05933
modules.json              cc3b70efa760b262c50d4ede0cebd40162178982ae9812ef1885c6bbe7f500c2
```

Reviewed unchanged snapshot committed locally asaa3c66f. All seven reviewed
hashes match; runtime results retain original execution input/revision, not a
claim of rerun after commit. No whole-unit promotion. Owned scratch/workspace/
module cache removed after recording results; no database resources created.

## Explicit authorization continuation and server-owned attempt preparation

Current sourceaa3c66f plus ten OTP/module paths consumes existing root IDs and
session RememberPolicy through public APIs. Both variants/no variant/unknown
discriminants fail; authenticated-session signin fails; pre-auth signin/MFA
preserves exact resolved remember limits with no defaults or renewed validity.
Attempt preparation accepts no caller ID/fingerprint, draws random128 identity,
derives complete keyed submission binding and retains code-free recovery identity.
Wrong-code decrement, reservation, current authorization and durable recovery
remain the next PostgreSQL/owning-workflow implementation, not this preparation.
Bounded plan is recorded in existing ledger; no implementation worker.

Missing authorization API RED8b26e2, focused GREENf48434,0.248s. Missing attempt
API REDf8f5b4, focused GREENdbd73b,0.243s. Final whole OTP/tests/examples race
PASS26fe40,1.275s, exact250/250 raw statements9de756. Vet PASS2f2a9e exit0.
Fuzz PASSa18556,3.288s203015 executions: `go test . -run '^$' -fuzz
'^FuzzFlowAuthorization$' -fuzztime=3s -parallel=2`. Race command:
`go test ./... -race -count=1 -coverprofile=<owned profile>`. Keys/scope/mode/
version conservation, independent HMAC/framing, entropy/cancel cleanup and
concurrent preparation/redaction have native behavioral assertions.
Explicit task-owned workspace composition uses root/sessionaa3c66f,
authentication303520b, authorization60ed4b3 and capability2e30f8b. Current
dependency worktrees inspected clean; no API changes, no permanent replace,
no checksum substitutions or shared build/module cache. No database resources.
Final isolated complete-diff review APPROVE, retained below; no whole-unit
promotion or release claim. Mutation/release/clean remote-consumer and composed
login remain pending. The unchanged reviewed snapshot was committed locally as
8657b130fd481835f73c19f0e4691f58aa68944f; source worktree inspected clean.
Task-owned /private/tmp/identity-otp-attempt.939USI and its isolated module cache
were removed after capturing native results; no Go processes or per-run build
caches remained. No database/container resources were created.

## Native attempt-preparation final review, verbatim

**APPROVE — development batch only. No findings.**

Native reviewer: `/root/otp_attempt_review_aa3`; single read-only reviewer using `$requesting-code-review`, without delegation, code execution, writes, or external mutations.

Reviewed `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`, branch `feature/identity-platform`, base/current committed HEAD `aa3c66f5c5aa41d7a2fab11b301974db72aed52a`, plus the complete ten-path dirty snapshot. Independently computed all ten SHA-256 hashes; each exactly matches the parent’s supplied hash.

Source-backed closure:
- Authorization discriminator and exactly-one member validation are explicit; signin rejects authenticated-session context.
- Session remember policy is retained without re-resolution, precision loss, or renewal. Non-signin/non-preauth-MFA purposes reject nonzero policy.
- Preparation accepts no selected ID/fingerprint; fingerprints bind complete scope, authorization/policy, mode, code, expected version, and key version.
- Command copies preserve opaque code custody; recovery projects only tenant/ID/fingerprint/key version. Formatting redacts and serialization rejects.
- Invalid/cancelled/entropy-failed preparation produces no command. Temporary fingerprint, framing, and entropy buffers are cleared.
- No new production callers exist; additions are consumed only by new tests/example. Existing root/session contracts remain unchanged.
- Independent literal HMAC framing, union-validation fuzz oracle, remember-field perturbations, cancellation/entropy cleanup, and concurrent ID tests meaningfully exercise the preparation contract.

Falsification challenges included mixed union members, authenticated signin, policy smuggling, nanosecond-distinct policies, repeated identical submissions, changed code/mode/version/key, and cancellation during entropy acquisition; source guards preserve the declared invariants.

Verification execution was **not independently repeated**. Parent-reported race/tests/example, exact coverage, fuzz, and vet results support this development-only disposition.

Remaining unproved boundaries: actual current authorization and original-policy provenance; durable denial accounting, reservation/replay fencing and ambiguity reconciliation; owning workflow/session effects; mutation, full release, remote clean-consumer, and composed login evidence. These remain explicitly unfinished, not approved as implemented or release-ready.

Reviewed exact source SHA-256s:
```text
44b8ba9f3239fffab1f6de5ba9728703dcfeaec972aea50aa3dd92fd1a1589dc otp/authorization.go
54b37bf796c95b61bfd8e39f56aa97c7b3b278791c04f277558ee9f20b1ab503 otp/authorization_test.go
f4d896a80ca10d6e5fa7f9b1065a2e9140f6efc9613f70192447b03fa9abbf8e otp/attempt.go
b009894351b77a4d311da6c87fae916cef678bf85ffe779ab51a0ca4e0046f33 otp/attempt_test.go
2385ab6387df70a967c0a781dac431e169fcf631fa00c84f5803ffa683fa85a3 otp/attempt_entropy_test.go
4fc7e0687355986909b58fbfb0a86c560c48e72076cb75b663f75fa9ba3264fc otp/attempt_example_test.go
29b1c12076da3db671ae8531956fbe0b2e49397108f0c45732244963a0f1abc2 otp/go.mod
457c1eb75dc73800b221cecc38d66f13a035a1cc760c0a1073ac1c7a23c988c8 otp/README.md
9090e21d7107f30cdcec7d4ee2f115d5811789de74d4daa0b06a35f07572109b otp/CHANGELOG.md
b1b63314855af58b53e48cac1b7c9534130c3c90bab7b7f2e85e318947de6104 modules.json
```
# OTP execution committed checkpoint

Unchanged reviewed snapshot committed5c1e4babacfbc4483326fe9be9f79ad34b74a004
(nativec9f0da); clean source. Binary diff and untracked hashes matchedb381bd.
ACK Core+line74adb2 PASS; effective impact unspecified. No push/unit promotion.

## Native independent report: /root/otp_execution_final_review

gpt-6.1-sol medium fork none, read-only default principal. Native final:

**APPROVE — No findings.** Scoped approval for this reviewed development batch only.

**Boundary:** `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`, branch `feature/identity-platform`; base and committed HEAD `161eb5bdcc32b16ee40bd9f3c56359fd743ce691`. Reviewed all nine modified files and three untracked OTP execution files, including tests and documentation.

**Snapshot SHA-256**
- Tracked binary diff: `40f43df7d9b7ec360826af89765dbae58b6f210e160de99f44b21eabb1980538`
- Complete 12-file content manifest: `d6e1e2f36ae94b485b7d56d2026da4a4ce07ed5ab0c244080cbdad841845ee6d`
- `otp_execution.go`: `f909a5873ec1dfc255cf26d8c0557593533515f1a9e285307b0247e9748b7339`
- `otp_execution_test.go`: `8bebf26012509683082b6b48fb55b4e9d59d40ee72677fa1df909edbfcad2c4a`
- `otp_execution_integration_test.go`: `e3d33c7682396506f96b98b02205f82c49de93b0252f2e3da4d0e3c9b5ddc1d6`

**Causal conclusions**
- Exact-owner, phase, lifetime and recovery-generation guards deny invalid custody (`otp_execution.go:43–83`). Code-free projections use the defensively copied `Command.Feature()` contract.
- Current checks precede protected mutation and repeat before terminal finalization. OTP finalization shares the domain/effect/receipt transaction (`feature_execution.go:242`, `domain.go:326–360`).
- Rollback release requires successful domain rollback and guarded terminalization; recovery release follows unanimous primary noncommit evidence (`domain.go:414–475`, `recovery.go:235–284`).
- Falsification challenges—expiry between apply and commit, failure after provisional finalization, divergent recovery generation, failed release and ambiguous commit—retain rollback or reserved custody rather than exposing partial success. Independent SQL assertions observe challenge, journal, domain, audit and outbox state (`otp_execution_integration_test.go:298–317`).
- Existing CAPTCHA/risk/capability sibling protocols and reservation/rebind/denial paths retain their established ordering and closure ownership.

**Verification:** Parent-reported native results: `test . -count=1 -race` PASS (2.154s); `vet .` PASS; affected PostgreSQL integration selector with `-race` PASS (9.759s); `git diff --check` PASS. No tests or mutations performed by this reviewer.

Production OTP storage/cryptography, authenticated final-login composition, exact coverage, mutation and remaining release gates are **not verified**. This approval does not establish unit completion, release readiness, publication or deployment.

## Native execution evidence

Baseline161eb5bdcc32b16ee40bd9f3c56359fd743ce691; dirty native execution delta.
Missing work/contributor APIs RED5499e7; focused custody GREEN1ae1cf0.225s.
Recovery OTPBinding REDcbd746; whole unit GREENdec01a0.346s.
Final affected SQL invocation241f01/685dc3:
`go test . -tags=integration -run '^TestPostgresIntegration$/^(OTP.*|coordinator_reservation|capability_reservation|risk_reservation|captcha_use_reservation)$' -count=1 -race`
PASS9.759s, owned PostgreSQL18.3 image80630f8; Go1.27.1. Fixture table
collision8ee794 fixed with disjoint OTP tables. Native owner fixture compares
only during reservation; phases check frozen generation/expiry and preserve
rollback/unknown custody. Not production OTP cryptography or final-login proof.
Final whole unit `go test . -count=1 -race` bb3850/85a6f3 PASS2.154s.
`go vet .` and `git diff --check`736114 PASS. All Go commands used disposable
workspace/caches Zv9165 mapping exact consumed source revisions, unchanged
published dependency declarations. Owned PG stopped0a8139 and absence7f21ee;
scratch/caches removed3a0f17. Image and unrelated services preserved.
Review is pending; no unit promotion or release/composed workflow claim.
# OTP native command custody development evidence

Reviewed snapshot committed unchanged84f8cd2317d789a535210942d35f393f79009c38
(45f9f2); clean source. Snapshot hashes independently matcheda4d119. ACK Core
and line validationa4d119 PASS, effective impact unspecified. Task-owned
cache/workspace removedf4d7c3; no service created, no push or unit promotion.

## Native independent report: /root/otp_command_record_final_review

gpt-6.1-sol medium fork none, read-only default principal. Native final:

**APPROVE — No findings.**

Single-reviewer read-only pass using `$requesting-code-review` and applicable boundary/coverage guidance. No writes, tests, Git mutations, delegation, prior reports or ledger narratives were used.

**Reviewed boundary:** `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`, branch `feature/identity-platform`, base/committed HEAD `5c1e4babacfbc4483326fe9be9f79ad34b74a004`, plus the complete nine-path dirty delta. Tracked diff SHA-256 independently confirmed: `3f4ace0e0d097f7ac2c6f85deda892f8d077e90c3a1cb2bbb01c81da6c0a1de7`.

### Source-backed closure

- `SealForCommand()` rejects zero/unbound attempts and authenticates original code custody before signing. `recordSignature()` binds the original random ID, complete authorization-bound continuation, submitted-code fingerprint, issuer version, mode, expected version and complete structural command/operation digest.
- `RestoreForCommand()` verifies record custody before checking the original submission fingerprint. It restores original bindings directly: no entropy, ID generation, clock access, expiry renewal or policy regeneration. Wrong submission, changed record and unavailable issuer version fail closed.
- `ScopeDigest()` uses domain-separated HMAC over the complete bounded scope and explicit retained key version. Historical lookup does not silently switch to the current key; absent versions deny.
- Refactoring `fingerprint()` preserves both existing preparation profiles byte-for-byte, including redacted error behavior. Existing literal independent profile tests cover the legacy and command-bound variants.
- Root `CommandBinding`/`FeatureOperation` validation bounds and freezes scalar inputs; all relevant fields remain represented by `attemptCommandDigest()`. Continuation construction binds the closed authorization union and exact session remember policy. Authentication’s opaque `OneTimeCode.KeyedDigest()` preserves original code bytes without exposing them.
- Public restoration values and retained verifier keys remain immutable; new operations mutate only local buffers. Automatic formatting is redacted and JSON/text export rejects. Persistence encoding, replay enforcement, current authorization and counted primary transactions are explicitly assigned to the later native adapter.
- Tests exercise original recovery identity, independent MAC/lookup framing, field tampering, wrong/retired keys, missing code custody, wrong submissions, expiry non-renewal, concurrent restoration and forged-projection fuzzing. Their fixture-to-outcome traces support the stated core contract, not durable single-winner enforcement.

**Adversarial falsification:** swapping an otherwise valid server attempt ID fails the record MAC; substituting a freshly constructed continuation with changed generation fails authentication; substituting a structurally valid foreign command remains covered by the command digest; restoring after expiry leaves `EvaluateAttempt()`’s expiry denial unchanged. The repeated ignored `recordScope()` error in restoration cannot become a new failure through caller mutation: the record is passed by value and its embedded continuation is opaque and immutable.

### Verification boundary

Parent-supplied native results: focused OTP race tests passed with exact **438/438 statements**, vet and diff checks passed, and final bounded fuzzing passed **146,141 executions**. I did not independently execute those commands. Release mutation, PostgreSQL adapter/storage, current-authority/replay integration, workflow/provider composition and complete unit/release verification remain pending; this approval covers the reviewed development batch only.

### Reviewed dirty snapshot SHA-256

```text
otp/attempt.go
02a7af8f8d5eda9bbebe29f97a5bd5000d89636cc8425356c584236e52bce58d
otp/command_record.go
98e85f3d83e2f03ea03577b5125db1ef2b08198141df0ad02efac85cb8f1be8b
otp/scope_digest.go
ce94d93c4e3da44bcd3e47755794db994f6ad3a679b1497dc8b664bef4fa610c
otp/command_record_test.go
bfc007ede8f9cdc6f30c373b49b6e6932bfdd2c633462ce7333068d68f382fac
otp/command_record_internal_test.go
d06506ca84f0b91f38b39b177515092a55c8d7c6c0060650a70b824a993777a5
otp/command_record_example_test.go
ceda709395b684a7ecffc942e5e9338a4fe89725b2d51079c02257232d05dad5
otp/scope_digest_test.go
7f17e11f96fc565b760aa1cbd1876fc4297e77b62fc0b721e626122deb47dbb4
otp/README.md
3849339ad904c6767c84d8b3761e9524626eb7887fa4ff3ae46b8633587f9b58
otp/CHANGELOG.md
eb805322b9eb77dd9f84088c105e28780f47f7e72f55dff240e406e7e0a793d4
```

## Native execution evidence

Base5c1e4babacfbc4483326fe9be9f79ad34b74a004, dirty OTP core only.
RED47a6c0 missing SealForCommand/RestoreForCommand; GREEN2bb9fc0.259s.
RED933af2 missing retained-version ScopeDigest. Final whole module111135:
`go test . -count=1 -race -coverprofile=<owned scratch>/coverage.out`
PASS1.284s, raw438/438 statements (5b3fa3). `go vet .`+diff-check5b3fa3 PASS.
Final fuzze488ae/884726 `go test . -run '^$' -fuzz '^FuzzCommandRecordRestore$' -fuzztime=3s -parallel=2`
PASS3.353s,146141exec, no failure. Earlier fuzze0312f passed148581exec.
Native wrapper uses Go1.27.1, task-owned cache/modcache/workspace e8jfxJ;
exact committed root/session/authentication/authorization/capability sources
replace only temporary workspace resolution, not releasable dependency pins.
Coverage raw profile sha256f6a595e0c731b24ffad2c8490e3d7846acef6c9316a3d6cb806f047e4ca70f51.
Independent literal HMAC framing oracles, changed original record/code/keys,
scope separation, fixed recovery identity, original counted policy/expiry,
concurrency, private diagnostics and lost internal custody covered. These are
adapter-neutral prerequisites, not native PostgreSQL storage/replay/issuance,
current authorization or adoptable login proof. Mutation/release/composition
remain pending; no unit promotion. Independent final review running.
# Native OTP codec review — final report

Principal `/root/otp_codec_final_review`, final reviewed source committed as
64d44818d90265691d1874d9250537a619f6f55c. All eight reviewed hashes matched.

**APPROVE — reviewed development checkpoint. No findings.**

The prior blocker is resolved: restoration derives the expected plan from the unique digest-key version authenticated by ciphertext/AAD, rejecting coupled version/scope-digest substitution. The added regression directly exercises that failure.

Reviewed the complete current dirty diff against base/HEAD `e223a15eb6ea79bd227271829c471e081a652385`, including core/coordinator consumers and local envelope/keyring contracts. No writes or additional agents.

Parent-supplied evidence: regression RED before fix; final `go test -race -coverprofile` PASS, 90.3%; `go vet` and diff checks PASS. Not independently rerun. Exact coverage, mutation, clean-consumer and complete adapter/workflow release gates remain pending; approval does not establish package readiness.

Reviewed SHA-256 hashes:

```text
modules.json                         673d885d885289975ce4f96c48ef3929d65030d92eca83261a84943ce675aa36
otp/postgres/go.mod                  54744c9dffb77822c0e7932c288cd7b737524b3bc371b19c75b5a66ec172c648
otp/postgres/README.md               f31d8015737fbadc5b03b2d721a321a28f461733ee6ec1d21266b7019c83cdff
otp/postgres/CHANGELOG.md            ba271a1b8184b37784fa04942ab6c9aec63231174c0e74783fbbef912910bf0b
otp/postgres/command.go              30a46a20ddfbb248c99e03cf29eb6063cb5b1ec321982b686f185834a6ef719d
otp/postgres/command_record.go       ba3eb8b68341a4705dde7eae71ee3a193df80f6ba7a58583490799f73e17b11d
otp/postgres/command_test.go         6e8c430ade25a4d9e62ec4725211515f39bebde0cea273dc1d710f3ce15244f6
otp/postgres/command_fuzz_test.go    22f21e5fec542812a8ecc97d6913724d3fb70cba911db0861bfb078987845465
```

# Native OTP codec review — original report

Principal `/root/otp_codec_final_review`; base/head e223a15eb6ea79bd227271829c471e081a652385 plus dirty patch.

**REQUEST_CHANGES** — development checkpoint review.

**Blocker:** [command.go:321](/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform/otp/postgres/command.go:321) reconstructs the expected plan using `plan.DigestKeyVersion`, rather than the version authenticated by command custody. A command mapped under version 3 accepts a substituted plan containing version 4 **and** its corresponding scope digest. `restoreRecord()` authenticates version 3 but discards it; plan equality then validates the substitution against itself. This violates exact retained-version restoration, though it does not currently expose an online verification endpoint.

Preserve the authenticated version through restoration, require plan equality against that version, and add a regression changing both fields together. Falsification: ciphertext/AAD, issuer MAC and submitted code remain unchanged; none of their guards compare the substituted plan version with the authenticated version. Existing tests change those fields individually, missing this coupled case. Finding is source-proven; no independent runtime reproduction was run.

Reviewed complete dirty patch at base/HEAD `e223a15eb6ea79bd227271829c471e081a652385`, including core OTP, coordinator contracts, local envelope/keyring, decoder, tests, documentation and dependencies. No additional findings.

Supplied race, fuzz, vet and diff results were considered, not independently rerun. Exact coverage, mutation, clean-consumer and complete native adapter/workflow release gates remain pending; this is not package-readiness approval.

Parent reconciliation: coupled-version regression RED883174, exact authenticated
version restored; whole native unit race58296b PASS1.267s90.3%, vet/diff PASS.
Final complete-diff re-review pending. No package promotion.
# Native OTP ownership review — original report

Principal `/root/otp_ownership_final_review`, model gpt-6.1-sol, effort medium,
isolated fork none. Reviewed base/head2f4927f8c3692f3b76b1dd76b1ab3bc4fc5f86c7
plus six-file dirty snapshot, subsequently committed unchanged at
f00a6b6c98f38228b78612bc03fc5fcfc1ef7af0. Parent matched all six hashes ec3e7a.

APPROVE — development batch only. No findings.

Reviewed `2f4927f8c3692f3b76b1dd76b1ab3bc4fc5f86c7` plus all six dirty files using `$requesting-code-review`; single-reviewer structured pass, no writes or test reruns.

Source-backed checks:
- `ownership.go:55–72`: authenticated original command/plan custody, one allowance increment, exact issued-version CAS.
- `ownership.go:99–158`: exact reservation identity and predecessor generation; takeover preserves counters/times; release accepts only unchanged owner or exact successor.
- `ownership.go:172–177`: terminalization shares the supplied transaction and live writes independently fence database expiry.
- Adversarial challenges: stale predecessor, skipped release successor, expired sampled validity, and terminal-version exhaustion are rejected; no expiry-derived release or exported release/transaction shortcut exists.

Supplied PostgreSQL integration race run and vet passed; both terminal-capacity regressions have supplied red/green evidence. Independently ran `git diff --check`: passed.

Not certified: public native dispatch, current authority/lockout, command replay/audit, issuance, authoritative recovery composition, provider workflows, complete coverage/mutation, or release readiness.

SHA-256, paths beneath `otp/postgres/`:
```text
challenge.go                           b8a10aaafad129ea9d73231b3ed8fcbf34450c5d771156e17aaf9d3e92066619
ownership.go                           d62fe394488b43720618e3e8706cbd0c97e605577a2de49e2115610d7b8b6635
ownership_integration_test.go          8646408eda8e328a40ea20314a3b55c28587a2c69708670c35e934b9d61346be
ownership_capacity_integration_test.go 794a047324ca512ddb51e881420cb196f618d2932611378e22da5a2675671357
README.md                              2d411884d9ce03eb83464fbd3e1b290c81a650c2427df8aaa043b3d6a267cfa5
CHANGELOG.md                           2c91f63f64006ada6d3a2b85b13a201ceb4b08a874cb08b50df6d02b007b5711
```
# Native OTP submitted-cookie custody review — original report

## Native independent review

**Disposition: APPROVE — development checkpoint only.** No concrete code defects or unresolved checkpoint findings.

**Reviewer:** `/root/otp_cookie_custody_final_review`
**Repository:** `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`
**Branch:** `feature/identity-platform`
**Base/current committed HEAD:** `e5e9138b1d0e7b6af6ae4a5f6bd1aeb07b1d8535`
**Reviewed head:** that commit plus the complete 14-file dirty batch. File identities were captured and rechecked unchanged before concluding.

### Source-backed conclusions

- Root cookie custody accepts only canonical unpadded base64url encoding of 32–64 bytes; length checks precede decoding. Padding, whitespace, noncanonical trailing bits, and undersized/oversized decoded values deny with zero command.
- Cookie/code custody remains outside `Feature()` projections. Existing `Command.Format()` and JSON denial protect command diagnostics. The retained bearer is an immutable defensive string copy.
- `PreAuthCookie()` denies zero, ordinary, code-only, and mixed commands. Every publicly constructible cookie-bearing command also retains OTP custody, so the existing generic `OneTimeCode()` admission rejection covers this new variant.
- `prepareClaim()` binds the actual submitted canonical cookie with existing keyed `identity-command-secret-v1` framing and distinct `pre-auth-cookie` path. Safe fields remain unchanged. `lockCommandClaim()` recomputes from original custody under the stored retained key, preventing changed-cookie replay from matching the prior intent.
- Legacy no-cookie/code-only encoding remains unchanged: the new payload member is conditional on cookie custody.
- `MapPreAuth()` requires complete bound pre-auth continuation, validates token syntax, and preserves existing `Map()` issuer, binding, encryption, plan, and code validation. Nil/cancelled contexts, nil codec, malformed credentials, and legacy ID-only attempts return zero outputs.
- The encrypted attempt encoder receives only the authenticated attempt record, not the submitted bearer. Restoration verifies the original record and plan without reconstructing absent credentials.
- The bounded adversarial challenge—same metadata and code with a different syntactically valid cookie—produces different canonical intent without changing safe fields. Such a cookie remains submitted intent, not authenticated possession; documentation correctly assigns comparison and admission to forthcoming native phases.
- All three affected module changelogs and READMEs distinguish custody/fingerprinting from admission and usable OTP login.

### Test and evidence assessment

The literal HMAC oracle is independent of the production framing helper. New assertions cover cookie retention/redaction, projection non-reconstruction, malformed syntax, concurrent reads, changed-cookie intent, retained-key reconstruction, generic admission denial, complete-flow mapping, restart restoration, legacy rejection, cancellation, and nil codec behavior.

Supplied executions were assessed against those source paths, **not independently rerun**:

- Four-package `go test -race -count=1`: PASS (`a3e6bf`).
- Four-package vet: PASS (`feed4c`).
- Final targeted kernel race test with literal cookie-HMAC oracle: PASS (`17ffa3`).
- Bounded `FuzzFeaturePreAuthCookieCustody`: PASS (`5b86e1`).
- `git diff --check`: PASS (`cb858d`).
- Supplied missing-API RED results: `3e202c`, `3147e4`.

No SQL, schema, migration, store, or native phase implementation changed. PostgreSQL execution was not run for this custody/serialization batch. Release coverage/mutation, aggregate gates, clean-consumer release proof, native participant activation, complete primary admission, and usable OTP login remain unproven; this approval does **not** certify those outcomes.

### Exact reviewed file identities

SHA-256:

```text
feature_command.go
5a7bcf06c38b46d937b0f862acd543b942c8149d16e2f89dc23cb53cd539e504
feature_command_internal_test.go
5f5085ce3ae623e1235eb814947db1fd4896dc2a324f13c7ab465284c61fe557
feature_preauth_cookie.go
1f03e7f0c7bc48fa3b9b59fe20918ebb85f4aa4c672e337c65b2abc8beb52571
feature_preauth_cookie_test.go
44cbe619409e9f702b1045f6ca2bc5e2634a263df0d88d1bf80cd3533ec36de8
postgres/codec.go
24f43b47182737a2578446499bf5c81b02282e13511a94f1f2d6d6efd2e41933
postgres/codec_preauth_cookie_test.go
f9b56c1802446111ba93f380aa80bc7f7c0bae9e307f34fece398f78471098aa
otp/postgres/preauth_cookie.go
a4c1d09eed84d8dea9d032b3bdf48e3a222ade80abdc0a29c06843459219aae0
otp/postgres/preauth_cookie_test.go
5d3e53811605edca223a8c79bd6b32f288c3ac740d285e9bb63b84a41314534f
README.md
cd26e453235d9c2d29d005bbbaa46996fb2e0c0fecb648536be9440f8dcf2b80
CHANGELOG.md
25b8b32116a603caff69acacd672ae3ed2e0b737672362d148a09804da78b086
postgres/README.md
41a5d23794db4ee9f45e382fb732d7170839ce85ceb0b5ffde2e7bdfa3610260
postgres/CHANGELOG.md
ea37e354faff9a5d1a7fb88a6867eaac6f37d85108ea9c1fc008b32f629d715a
otp/postgres/README.md
db3d13ee65c2f1a2c52951be502da1e31fac8f919931f7ab8a67cbc45ecbb198
otp/postgres/CHANGELOG.md
6e628cebcba32474870448c76123e19d77e82b5c2291f85a2ab7c0478d74084a
```

Parent reconciliation: exact reviewed fourteen-file snapshot committed locally
at7c1bce3b5ba2cb26f1c63c4bdcedbde5e996d717; source clean45b8b4. No push,
unit promotion, release claim or database access. Bounded-run caches removed.

# Native OTP complete pre-auth custody review — original report

## Native independent review

**Disposition: APPROVE — development checkpoint only. No findings.**

Reviewer: `/root/otp_preauth_custody_final_review`, isolated single-reviewer read-only pass using `$requesting-code-review` and its coverage-oracle guidance. No writes, Go execution, service access or delegation.

Repository: `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`
Branch: `feature/identity-platform`
Reviewed base and unchanged HEAD: `c53bc05072105afb4f15d6d07abac2d5822b85b1`
Reviewed head state: complete dirty ten-file batch; hashes reconfirmed after inspection.

### Source-backed conclusions

- Full original session projection survives immutable continuation custody and v2 decoding: intent, nonce, callback/no-redirect, original timeline, remember policy, named cookie commitment and record version.
- Continuation binding feeds the existing keyed submitted-code fingerprint and independent record MAC. Changing the retained nonce invalidates authenticated custody; the other original fields independently change the authorization digest.
- Tenant and signin/pre-auth-MFA remember substitution deny. Bound anonymous commands reject a different flow namespace.
- Legacy constructors, authorization v1 framing, attempt-record v1 framing and their bounds remain unchanged; existing literal golden tests exercise those contracts.
- Restart restoration authenticates encrypted command AAD and the original core MAC before use. Counted restoration still requires the original opaque submitted code; persisted custody cannot reconstruct it.
- Decoder bounds, scalar validation, canonical UTC timestamps, truncation and trailing-field rejection fail closed with redacted errors.
- Maximum valid bound plaintext is conservatively **6,170 bytes**; fixed local binary envelope overhead produces at most **6,404 bytes**, below the unchanged 8,192-byte wire/feature-field limits. The suspected envelope-size overflow is disproved by binary—not base64—framing.
- Native adapter restoration currently has no activated production participant caller. Coordinator phase contracts and canonical session primary admission remain their existing owners; this batch neither grants admission nor adds transaction/state ownership.

### Verification boundary

Parent-supplied evidence accepted as execution evidence, not independently rerun: focused RED for missing API; final core/adapter race tests, integration-tag vet and bound-record fuzz passed. Reviewer independently confirmed `git diff --check` and inspected the pinned clean secret-envelope source at `22fc4a2d925dfad2023437c9ef4ed862d501d4d0`.

**Not certified:** native phase activation, working OTP login, live PostgreSQL phase behavior, exact coverage/mutation, clean published-consumer composition or release completion. Those remain pending under the authoritative execution goal, which permits reviewed development checkpoints.

### Exact reviewed file identities

```text
2601fe803b4ab2e12ab4d4001a97030f7f7fff4ee31f6dc3f0dab03b1be2fed7 otp/authorization.go
e06dcdb2a883d0e4175accf707a545b1bfd97a568bfe474ec95bb2f987719ca0 otp/command_attempt.go
bdbbf1830d614b99a4580217b4f752fa5c10b13967ed60499f7081cdb04b66a3 otp/README.md
3ab1d42460db2ed29e717d6f41b1ce650a09bbd10028809a5366e02466c1d05b otp/CHANGELOG.md
1e05d3745f60f33bbc3c1d061a1eab983f9eb9e87e1623f0c43c288fc8784d71 otp/preauth_authorization_test.go
cb11c46deaea1ba87be02feb032b343b8572ee7e711fe7ed4741bbf93ecbc930 otp/postgres/command_record.go
db3f470ee38f36793afd41f3b7c51dbd5e5dcf9504a1f9862dca779e207ca4ba otp/postgres/preauth_record.go
ece09fe9dcfe015725691c89a3ef1cf1ba2c348ad6b03d7e5ac915de9a85856b otp/postgres/preauth_record_test.go
09ce6f6005a5ebf473e8b61009011c830bf12cbfc5bd4dfb50e0bfc6acd21c5d otp/postgres/README.md
0a2408def51001dcefd6b02b74f860198e3ee39f7d4bde9d5413b80f0c066a82 otp/postgres/CHANGELOG.md
```

Parent reconciliation: unchanged reviewed ten-file snapshot committed locally
at e5e9138b1d0e7b6af6ae4a5f6bd1aeb07b1d8535; clean source134447.
Final native race213989, vet exit0, fuzz34218 PASS; no unit promotion or push.

# Native OTP shared lockout review — original report

Principal `/root/otp_lockout_final_review`, model gpt-6.1-sol, effort medium,
isolated fork none. Reviewed base/headf00a6b6c98f38228b78612bc03fc5fcfc1ef7af0
plus six-file dirty snapshot, subsequently committed unchanged at
9721db43b2f1831c47f37fe2d1e0f9bb040dd77b. Parent matched all six hashes cef323.

**APPROVE — development checkpoint only. No findings.**

Reviewed base/HEAD `f00a6b6c98f38228b78612bc03fc5fcfc1ef7af0` plus all six frozen dirty files, using a single independent read-only review.

Source-backed invariant challenges:
- Reissue/purpose/policy/generation changes cannot alter shared identities; subject changes retain destination exclusion and destination changes retain subject exclusion (`lockout.go:61–83`). Tenant/channel separation and pre-account destination-only behavior are explicit.
- Changed or absent deployment binding fails before row initialization/loading (`lockout.go:86–124`); singleton uniqueness and immutable-binding trigger prevent ordinary DML reset (`lockout_migration.go:48–50,77–82`).
- Missing rows deny; existing rows lock destination then subject, and database time is sampled only after both lock waits (`lockout.go:152–207`). The corrected committed-issuance contention fixture exercises this ordering (`lockout_integration_test.go:230–307`).
- Shorter extensions cannot reduce expiry: `GREATEST`, exact version CAS, uint64 overflow denial, and SQL monotonicity guards preserve stored exclusion (`lockout.go:216–235`; `lockout_migration.go:61–76`). A failure after the first projection remains inside the borrowing transaction; no helper commits, retries, or invokes callbacks.

The supplied final `go test -tags=integration -race -count=1 .` result passes against PostgreSQL 18.3; integration vet/diff checks pass. I did not rerun execution.

Public dispatch, issuance/bootstrap/key migration, current-flow authority, coordinator composition, denial/count/audit/result atomic integration, full coverage/mutation, release/provider/composed-workflow gates remain unfinished or unverified. This approval does **not** certify package readiness.

SHA-256:
```text
ae966c5d233af6415db6a59012d4a197fcf9bf5fd7993da64d4137548e85fee2  otp/postgres/lockout.go
ca03b544dbcc7f6aa68984cf5ce0e6c2592fe911ce051cd3a4bb1ca4b8c16d1e  otp/postgres/lockout_migration.go
4d5c3eb3dfc9cc309f145c4dc03947a71fc129d4cfed4dacfcb76334f1804f32  otp/postgres/lockout_test.go
a4e2b77dd77ba97202c5675a00f2d07ae16a6230eef66af9d4affbf614d5e849  otp/postgres/lockout_integration_test.go
616dffc1ff973031fc01f0e417d527b2cb7b28fcbdb0c51fc304ac13a6eadf99  otp/postgres/README.md
c9c8e74ca70dcc75cf401c252199d1aba6d1f6eda0a3f9352815bf3f255bf2c3  otp/postgres/CHANGELOG.md
```

## Native pre-auth admission — 2026-10-07

Local source commit: `2f9b18eec1756acd6fc2552b64eabfa781d3e0ef`.
Every role below used a distinct read-only default principal, gpt-6.1-sol,
medium reasoning and fork none. No implementation workers.

### Native mapping report — /root/otp_admission_mapping

Native `change_mapper` report — head/base `7c1bce3b5ba2cb26f1c63c4bdcedbde5e996d717`; read-only, clean checkout.

**Existing contracts**
- `otp/postgres.MapPreAuth` requires complete bound pre-auth continuation and canonical actual bearer; legacy ID-only continuation denies. Mapping explicitly grants no cookie verification or primary admission.
- Encrypted attempt custody preserves complete original flow, cookie commitment/version and issued record version. `restoreRecord` authenticates envelope, original attempt seal, retained named attempt/code keys and exact plan.
- Root command keeps actual cookie separate from scalar feature metadata. Coordinator fingerprint separately binds submitted cookie; changed cookie changes command intent.
- `Token.PreAuthDigest` already implements dedicated named-key HMAC over complete original flow plus actual token. `PreAuthTransaction.Check` supports independently computed digest comparison and exclusive expiry.
- Session SQL owner already locks and compares complete canonical record, original timeline/digest/version, primary time after lock, and reserved command/fingerprint/generation. Reserve/finalize/rebind/release remain distinct transitions; proven-noncommit release intentionally does not enforce live expiry.

**Concrete admission gaps**
1. No production caller recomputes `PreAuthDigest` from `Command.PreAuthCookie()`. Existing session admission calls `record.Check(now, s.Binding, s.Cookie)`—the original projection’s own digest—so it proves row agreement/expiry, not submitted-cookie possession.
2. `AttemptCodecConfig` has wrapping, attempt and OTP-code keys, but no retained dedicated pre-auth keyring. Exact original cookie key selection and retired-key denial currently have no native OTP owner.
3. Denial/reserve/rebind handles expose original `NativeCommand`; apply/finalize expose only code-free `CommandSpec`, which omits bearer custody. Any required apply/finalize recomputation needs narrowly scoped separate-cookie access, not general code exposure.
4. No native OTP contributor implementation is present in `otp/postgres`; phase orchestration contracts exist in `identity/postgres`, but automatic current-flow admission across those phases is not wired.

**Variants to preserve**
- Authenticated-session continuation and complete pre-auth continuation are disjoint.
- Legacy ID-only continuation remains structurally available but cannot satisfy cookie-bound admission.
- Original commands without cookies remain supported by existing constructors; their use as pre-auth OTP admission must deny.
- Issued admission differs from reserved exact-owner admission; takeover requires predecessor generation, not current generation.
- Historical recovery is code/cookie-free and returns the original receipt; it must not recompare, remint or renew.

**Required observable boundary**
Actual separately submitted bearer → exact retained key named by authenticated original flow → independently derived commitment → locked original primary record/state/version → primary exclusive expiry → exact reservation owner where applicable. Missing/wrong cookie, absent/retired key, changed original primary fields or ownership must deny before attempt charging or protected effects.

Goals require database-owned expiry, enumeration-safe cross-scope denial, no unauthorized attempt decrement, generation-fenced reservation and ambiguity recovery. Current storage helpers cover much of primary fencing, but do not prove bearer admission. No tests or gates executed; this report is mapping, not correctness or completion certification.

### Native QA report — /root/otp_admission_qa

**QA role:** `qa_test_planner`; one bounded read-only pass using `$qa-test-planner`.
**Reviewed base/head:** `7c1bce3b5ba2cb26f1c63c4bdcedbde5e996d717`. No edits or execution.

### Acceptance assertions

| Partition | Observable assertion / realistic defect caught |
|---|---|
| Correct canonical submitted cookie, original flow, retained named key, issued primary row | Admission succeeds; all persisted rows and counters remain unchanged. Prevents an implementation that denies every input. |
| Wrong canonical cookie | Uniform denial despite otherwise valid original custody and primary record. Catches trusting persisted digest instead of submitted bearer. |
| Missing cookie or reconstructed metadata-only command | Denial; persisted fields cannot recreate bearer authority. Catches accepting custody without actual submission. |
| Original key retained alongside another newer key | Original cookie succeeds using its named version. Same version with different material denies. Catches current-key substitution/version-only checking. |
| Named original key retired while another key remains | Denial, without fallback or renewal. |
| Stale original custody with coherently changed primary record | Denial. Change one representative authenticated intent field, such as nonce or origin, in a valid replacement record and matching physical indexes; keep original submitted command. Catches reading current primary data and treating it as authenticated original intent. |
| Primary row absent, terminal, or wrong issued version | Denial; ID alone is not current flow authority. |
| Expired primary flow, including expiry reached during lock wait | Denial using database time after locking. Catches pre-lock/application-clock expiry checks. |
| Reserved row and exact command/fingerprint/generation | Reserved admission succeeds; issued-only admission denies. |
| Reserved row with different command, fingerprint, or generation | Each denies independently. Catches incomplete reservation fencing. |
| Repeated successful/denied admission | No OTP `used_attempts`, challenge state/version, preauth state/version/ownership, attempt-denial records, command claims, or command journal changes attributable to admission. |

All errors must remain redacted. Corrupt physical index/payload agreement may retain the existing `Unavailable` classification; it must never admit.

### Smallest sufficient verification

1. A focused real-PostgreSQL OTP admission table test, using the production mapping/restoration path and newly owned admission boundary—not only direct `session.RequirePreAuth` calls.
2. One PostgreSQL lock-wait expiry test through that new boundary.
3. Reuse the existing `TestSessionPostgresIntegration` preauth cases for lower-level lock, owner-transfer, primary drift, and expiry behavior.
4. Rerun existing mapping/original-record tests and focused OTP admission tests with `-race`.

The existing `otpDatabase` harness requires `IDENTITY_OTP_POSTGRES_TEST_DSN` pointing to loopback `identity_test`; it creates and drops a unique task-owned database. Install session preauth migrations in that database alongside the existing OTP fixtures. Give every Go invocation a disposable `GOCACHE`, cleaned on every exit.

Observe the wrong-cookie regression fail before the production fix. Snapshot persisted consumer state before and after admission inside the same transaction; commit successful read-only admission and verify unchanged durable state.

No contradictory requirement or reproducible design defect blocks the first write. Exhaustive combinations, full coordinator delivery/recovery journeys, and additional fuzz campaigns are not necessary for this bounded admission slice. Package/release gates remain separate mandatory delivery requirements.

### Native initial regression report — /root/otp_admission_regression

**Native disposition: APPROVE — bounded admission checkpoint only.**

**Reviewed:** `7c1bce3b5ba2cb26f1c63c4bdcedbde5e996d717` plus the complete six-file dirty patch in `otp/postgres`: admission implementation, unit/integration tests, dependency, README and changelog.

**Findings:** No concrete regression or material coverage defect identified.

- Authentication derives the commitment from the actual submitted bearer, the exact retained cookie-key version and encrypted/authenticated original flow. It does not substitute caller plan fields or fresh primary intent.
- Issued/reserved admission delegates to session-owned primary locks, exact original record/version checks and database-time expiry after locking.
- Reserved admission checks complete command/plan correlation, nonzero fingerprint, positive current generation and absence of predecessor transfer authority.
- Assertions exercise absent/wrong cookies, missing/changed keys, altered plans, changed intent/version, foreign reservation custody, terminal/missing state, unchanged durable flow/journal state and expiry reached during an observed lock wait.
- Existing session tests additionally exercise reserved expiry, takeover exclusion and stale-generation rejection.

**Evidence boundary:** I inspected sources and assertions; no commands executing tests were run by this reviewer. Parent-reported integration race and vet results remain parent execution evidence.

**Remaining boundary:** Production contributor activation, current-authority checks, challenge/lockout ordering, protected effects and recovery are intentionally outstanding. This disposition does not certify full OTP completion or release readiness.

### Native reconciled regression report — /root/otp_admission_regression

**Native disposition: APPROVE — reconciled current six-file checkpoint.**

Reviewed head remains `7c1bce3b5ba2cb26f1c63c4bdcedbde5e996d717` plus the complete dirty patch.

The revised stale-primary fixture recomputes the cookie commitment using the same actual bearer, dedicated named key and changed nonce binding. It now represents a coherent replacement intent, so rejection meaningfully detects fresh-primary-intent substitution rather than merely inconsistent persisted custody.

No concrete regression or material coverage defect found. Production implementation and other assertions remain unchanged.

No tests run by this reviewer; the parent's targeted race rerun remains separate execution evidence. Production OTP contributor activation and full-goal completion remain outstanding.

### Native compatibility report — /root/otp_admission_compatibility

**Native go_compatibility report**

Reviewed `7c1bce3b5ba2cb26f1c63c4bdcedbde5e996d717` plus the six-file dirty admission patch.

**Disposition: APPROVE for the bounded local checkpoint only.**

- New `session/postgres` pseudo-version matches the commit’s exact UTC timestamp and revision.
- New dependency is acyclic; consumed APIs are exported. Admission helpers remain private, with no new public consumer store or permanent replacement.
- Go 1.27 declarations agree.
- **Source-lock/clean-consumer gap remains:** `otp/postgres` selects `session@84f8cd2317d7`, which lacks `session/preauth.go`. The newly pinned session adapter itself selects an older session revision. Disposable workspace replacements conceal this existing manifest debt; the adapter pin alone does not establish reproducible compilation.
- No release readiness, published availability, clean-consumer success, or complete source-lock claim is approved. No tests executed during this read-only audit.

### Native final correctness report — /root/otp_admission_final_review

**APPROVE — No findings.**

Role: independent read-only correctness reviewer `/root/otp_admission_final_review`.

Boundary: base/head `7c1bce3b5ba2cb26f1c63c4bdcedbde5e996d717` plus the complete six-file dirty checkpoint in `otp/postgres`: `go.mod`, admission implementation, both admission tests, README and changelog. Reviewed production hash `6378a6fc…`, integration-test hash `ff41e350…`; parent confirms all six hashes unchanged.

Source tracing confirms independently submitted cookie verification under the exact retained key and authenticated original intent, followed by locked primary state/version, exact reservation ownership and post-lock expiry checks. Coherently replaced primary intent, wrong cookie, retired key and foreign custody deny. Admission introduces no writes, counter charging, renewal or journal authority.

Verification supplied by parent: final `go test -race -tags integration -run '^TestNativePreAuth' -count=1 github.com/faustbrian/go-identity/otp/postgres` passed (2.620s); integration vet and diff check passed. Earlier complete adapter integration and affected race checks passed.

Approval covers this **local intermediate checkpoint only**. Production contributor activation, current authority, complete phase composition, pinned clean-consumer compatibility and remaining release gates are not certified. No tests or mutations performed by this reviewer.

Final six-file SHA-256 values, matched immediately before source commit:
```text
c7c6c5700bc2475980c9aa534a25495050dc88659195038f94bd07717a64b240  otp/postgres/go.mod
6378a6fc0e89f93221e4942f1164c7a6ba42a8b1278b3b9f695d782ee4368a74  otp/postgres/preauth_admission.go
149831e9c308b11970380a9004761aeda9f7ad2157b55c97c4b67a002843df68  otp/postgres/preauth_admission_test.go
ff41e350909f846487bae70bd80c203cdc98c99296f166a7bd328c2a09ed16b6  otp/postgres/preauth_admission_integration_test.go
6923f13b4126f1a72c963af42f9d472ce4587deb9b1af14aa183725c43c1f4a0  otp/postgres/README.md
f0eafbf9c66a8f2c1a6745d637d8219e5909cbaa1eb078e0dfed3f7a801f1559  otp/postgres/CHANGELOG.md
```

Execution: initial compile RED7672e2 (missing native admission APIs); complete
OTP PostgreSQL integration race PASSa297b4 5.401s; coherent replacement oracle
final native race PASS8fa4fe 2.620s; existing session integration race PASS87508a
4.196s; OTP core and adapter race PASS176c24; final integration vet/diff PASS09766c.
Disposable PostgreSQL container e935dc8e84a31e780270f87879f00af3fdf3e7f4d1284bb4f56f218b3ec48e59
used pre-existing immutable image5a5a84b19854; owned tmpfs data, loopback-only
random port and unique per-test databases. Owned container stopped/auto-removed
a98f48; pre-existing images/services untouched. All Go commands used owned caches.
No release/clean-consumer, coverage/mutation, login or unit-completion claim.

### Native delivery report — /root/otp_admission_delivery

**Native delivery audit: APPROVE — local intermediate checkpoint only.**

Reviewed retained native mapping, QA, initial/reconciled regression, compatibility and final correctness reports. Their base/dirty-scope boundaries reconcile to committed `2f9b18eec1756acd6fc2552b64eabfa781d3e0ef`, parent `7c1bce3b5ba2cb26f1c63c4bdcedbde5e996d717`. All six recorded SHA-256 values match committed blobs; six files, 599 additions; source checkout clean.

| State | Supported claim |
|---|---|
| Local | Private admission checkpoint present; focused race/vet passes recorded, not rerun by auditor. |
| Committed | Exact reviewed content committed as `2f9b18e`. |
| Pushed | Not delivered; no containing local remote-tracking ref. Remote state not independently refreshed. |
| CI | No current-head hosted CI certification. |
| Deployed | No deployment claim or evidence. |
| Production | No production verification or full OTP completion claim. |

**Outstanding, nonblocking for this bounded claim:** source-lock/clean-consumer debt is confirmed: pinned `session@84f8cd2317d7` lacks `session/preauth.go`. Private admission has no production caller; contributor activation, current-authority enforcement, complete phase composition and release gates remain pending. Coverage/mutation and release readiness are not certified.

Cleanup is parent-recorded: owned caches removed and PostgreSQL container stopped/auto-removed; auditor performed no runtime inspection.

## Credential-free user admission and protected cookie access — b523b3f

### Native delivery report — /root/otp_user_admission_delivery_b523b3f

**APPROVE — focused local prerequisite checkpoint only.**

Exact head: `b523b3feb87a0428def6dcd75a692decb1a8079d`; parent: `918965f0a05adc9e73fa8f7474091ff586150abc`.

Independently inspected native mapping, QA, regression, compatibility and final-review reports. All nine committed blob hashes match the final reviewer’s snapshot; source checkout is clean. Retained focused race/integration/vet passes remain applicable; no Go or services rerun.

- **Local/committed:** reviewed checkpoint present and committed.
- **Pushed:** not claimed; no containing remote-tracking ref. Remote not refreshed.
- **CI/deployed/production:** not certified.
- **Cleanup:** native `9c12fa` succeeded; owned build/modules absent; recorded Docker inspection confirms owned container absent.

Attachment explicitly permits coherent reviewed development commits with pending gates. Coverage/mutation, source locks, native composition, login, aggregate/release verification and full unit completion remain pending—not waived.


Current prerequisite development, not a unit-completion or native-workflow
claim. User-only primary locks compare original three independent security
epochs without a credential reference. Existing credential checks unchanged.
Protected phases lend only the separately submitted original bearer to their
exact live owner. Actual dispatch invalidates returned access.
Source committed b523b3feb87a0428def6dcd75a692decb1a8079d from base918965f.
Nine files,530 additions/1 deletion; reviewed hashes unchanged. Source clean.
Normal hooks and Intent Core validation passed; no push.
Missing APIs RED13efdc/afc089; complete kernel PostgreSQL race PASS98e0d5
2.336s; real PostgreSQL authority integration race PASS186563 2.196s;
standalone integration vet exit0 a62e46. Coverage/mutation, source locks,
native contributor composition, login and release gates remain pending.

### Native complete-final-diff report — /root/otp_user_admission_final_review_918965

APPROVE — No findings.

Reviewed exact base/current HEAD `918965f0a05adc9e73fa8f7474091ff586150abc` plus the complete nine-file working-tree delta. Single-reviewer structured pass using `$requesting-code-review`, required references, testing-coverage and testing-integration.

Source-backed closure:
- Command-bound tenant/user lookup is primary-key scoped; active status and all three independent epochs are compared under a retained `FOR SHARE` lock without credential creation.
- Live malformed, stale, canceled, foreign-owner and cursor failures poison Work; returned handles cannot poison later invocations.
- Existing credential eligibility/expiry admission remains unchanged.
- Protected cookie access requires exact live ownership and phase, returns only the submitted bearer, and expires after actual dispatch. Recovery/release expose no cookie accessor.
- Adversarial challenges: changed credential epoch despite matching aggregate version denies; a retained protected handle after callback return yields zero bearer without poisoning later Work.

Inspected native results: missing-API reds `13efdc`/`afc089`; race suite `98e0d5` passed; real PostgreSQL integration `186563` passed; standalone integration vet `a62e46` passed; diff validation `e6b5c1` passed. I did not execute Go or services.

Approval is limited to this focused development checkpoint. Native OTP/login composition, exact coverage/mutation, aggregate and remaining release gates are not certified or waived.

SHA-256 snapshot:
```text
postgres/user_admission.go fa76d80b07bb9fcbe1328d3727d924bdfc7156dc8f7bb4895efed657a84babd0
postgres/otp_authority.go 435583bf5b0cdfe9752c27208e0ff85d3ee5a015b20a499a8c0b41824652403f
postgres/otp_execution.go 57cb8b5bf3277a8587afdbfdbf151150a89ad8b76ebfea986e8c313167505611
postgres/otp_phase_cookie_test.go 9d1f090d7ef5548cb4f458e48b941f79d3868b70f884a833e68111523d9df8f2
postgres/otp_user_admission_test.go 00a192b0bc1c26bba36a72a8afbac1678e9703e430156c753a940dcdd48c7b0e
postgres/otp_user_admission_integration_test.go 47af66611f7f419c4935efb4e2be6aa325c9e45d78da1d92532e4367bb1359a8
postgres/otp_authority_integration_test.go 01399c87ffda7d8b2a738c04820818cc22536516218c57ea9c454284d49c1d23
postgres/README.md 31d1152442f822f4fdcbd832def365db059fb0bcc21871ed8932d397edf08371
postgres/CHANGELOG.md ba4f9a8329c2c8510995c216422667a314f2b3fe0c948db133b729a33a8a9f24
```

### Native report — /root/otp_native_activation_qa_918965f

## Bounded QA artifact

**Role:** fresh read-only `qa_test_planner`; `$qa-test-planner` read completely.
**Reviewed repository:** `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`
**Exact base/head:** `918965f0a05adc9e73fa8f7474091ff586150abc`
**Disposition:** no contradictory requirement or reproducible design defect blocking this prerequisite’s implementation. Additional cases below are implementation work, not approval dependencies.

### Contract and current boundary

- Existing OTP attempt, reservation/rebind, apply and finalize handles expose exact-owner directory, credential and authorization checks. Credential admission requires a credential-reference row; this does not prove passwordless user security eligibility.
- `CredentialAdmission` already distinguishes user, identifier and credential security epochs from aggregate `User.Version`.
- The command already retains independently submitted pre-auth cookie and opaque OTP code separately. Protected OTP handles expose structural `CommandSpec`, not the original command’s cookie.
- Apply/finalize must admit current primary security against **original bound expectations** and preserve cookie custody without granting OTP-code access, authentication, renewal or transaction ownership.
- Directory, user-security, authorization, cookie/pre-auth provenance and OTP one-time authority remain separate checks. This prerequisite is not full OTP native-contributor, reference-composition or release completion.

### Smallest sufficient behavioral coverage

| Boundary | Observable assertions | Realistic defect detected |
|---|---|---|
| Real PostgreSQL, credential-free user | With no credential-reference row, exact original security epochs and active user admit; primary persisted effect can commit. No credential row is created. | Accidental credential JOIN, invented reference or password-only eligibility. |
| Original epochs | Freeze distinct positive user/identifier/credential epochs; independently change each primary epoch after preparation. Each denies, returns no usable clock/authority, and prevents persisted effects. Change only aggregate `User.Version`: security-epoch admission remains valid unless a separate target-version contract requires denial. | Adopting current epochs, substituting aggregate version, comparing the wrong epoch. |
| Primary status and scope | Suspended/deleted/absent user, foreign tenant/user, missing authority and malformed/zero expected epochs deny without mutation or identity disclosure. Repeat denial after blocked→active with a changed security epoch; original command still denies. | Status omission, cross-tenant admission, resurrection or stale intent refreshing. |
| Lock behavior | Hold primary user lifecycle lock in another transaction. Bounded admission times out; ignored failure cannot commit. Successful admission retains exclusion until its owning transaction ends; concurrent blocking/epoch update cannot slip between admission and finalization. | Unlocked read, unbounded wait, premature lock release. |
| Cursor poisoning | During each live relevant OTP phase, leave `WorkRows` open, invoke security admission, ignore error, close rows, then attempt further effect/commit. Transaction remains poisoned and effects absent. | Cursor cleanup erases security denial. |
| Handle custody | Zero, nil/typed-nil/noncomparable/foreign owner, wrong phase, copied-after-return, canceled/nil context and expired Work reject without SQL or panic. A failed live invocation poisons its Work; an already-returned handle cannot poison a later valid phase. | Descriptor equality replaces instance custody; retained handle survives callback; late denial damages later Work. |
| Apply/finalize cookie conservation | Original submitted cookie A reaches only the exact live protected owner in both phases; valid different cookie B remains B, not a declared commitment or reconstructed token. Code-only/ordinary commands return absent cookie. Command mutation/copying cannot change frozen cookie. | Dropped or substituted cookie, treating caller digest as proof. |
| Secret isolation | Cookie borrowing exposes no OTP code or full original command; structural projection, plan, receipt, recovery metadata, serialization, formatting and errors contain neither cookie token nor code. Access after callback rejects/returns zero. | Full-command accessor creates secret disclosure or durable bearer custody. |
| Protected dispatch integration | Exercise actual coordinator callback dispatch, not only manually assembled handles. Exact admission/cookie checks succeed through apply/finalize; failed or deliberately ignored admission rolls back participant finalization, protected effects and success result together. | Helper works in fixtures but is unavailable/mis-scoped in production; denial commits prefixes. |

Use paired tables for phase × custody-invalid class; use separate cases for each independent epoch because substituting one for another is a security defect. Existing credential/directory/authorization behavior should remain covered by its current suites rather than duplicating their complete matrices.

### Verification recommendation

1. Observe focused new admission/cookie regressions fail for the expected missing behavior.
2. Run affected identity command/cookie and PostgreSQL OTP handle tests.
3. Run real PostgreSQL admission and production-dispatch integration cases above, including lock and rollback assertions.
4. Run affected race tests for callback lifetime/copies and concurrent primary lifecycle updates.
5. Run the repository’s narrow affected release gates separately before any unit-verification claim.

Every Go invocation requires task-owned disposable `GOCACHE` and cleanup. **No Go command, service operation, file edit or external mutation was performed in this planning pass.** Full OTP contributor, reference journeys, mutation/coverage and release evidence remain outside this prerequisite proof.

### Native report — /root/otp_native_activation_map_918965f

## Native change-mapper report

**Scope:** read-only, bounded pass; no writes, tests, services, or subagents.
**Reviewed repository/base/head:** `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`, `feature/identity-platform`, `918965f0a05adc9e73fa8f7474091ff586150abc`.

**Authoritative requirements:** current attachment; control-repository `goals/identity-otp.md`, `goals/identity-otp-postgres.md`, and `TRANSACTION_CONTRACT.md` durable OTP protocol. Historical prerequisites do not replace the attachment’s focused consumed-contract development allowance.

### 1. Credential-free primary user security admission

**Existing ownership**
- `identity.users` owns status and independent `authority_version`, `identifier_version`, and `credential_version`.
- Kernel PostgreSQL owns their admission, synchronization and transaction lifetime.
- `postgres/status_owner.go` changes status/user authority; `identifier_mutation.go` changes identifier authority; credential lifecycle owns credential references and credential authority.
- Directory global/tenant epochs have `DirectoryAuthorityAdmission`; effective-user authorization has its existing authorization-owner admission. Neither replaces user security state.

**Current public/read boundaries**
- `Repository.GetStatus` (`postgres/status_read.go`) reads status plus user authority.
- `GetIdentifierState` (`postgres/identifiers.go`) reads identifier authority.
- `GetCredentialAuthority` (`postgres/credential_authority.go`) requires an actual credential reference; it cannot capture credential authority for a user with no reference.
- `CredentialAdmission` (`postgres/credential_admission.go`) binds the exact reference plus all three security epochs. Its primary query joins `identity.credential_references`, share-locks user/reference, checks reference eligibility and exclusive credential expiry after locking.
- Four native handles expose `RequireActiveCredential`: attempt, reservation/rebind, apply and finalize (`postgres/otp_authority.go`).

**Actual missing contract**
A kernel-owned admission boundary for the exact command-bound tenant/user that compares **status and all three original positive full-width security epochs without requiring or fabricating a credential reference**. Passwordless user commands require it; anonymous commands must not acquire user dimensions.

`otp/postgres/AttemptAuthority` already freezes those epochs and permits no credential reference for user-bound commands (`authority_custody.go`). Its v3 envelope authenticates and preserves the values. Mapping is explicitly not current admission. Thus custody exists; its credential-free primary enforcement is missing.

If native activation must itself capture this snapshot, the current repository readers also lack a single credential-independent projection containing all three epochs. Do not infer credential authority from an empty reference or `User.Version`, and do not refresh frozen expectations during retry/rebind.

**Compatibility requirement**
Keep existing credential admission unchanged in meaning: exact reference identity/version/kind/owner, enabled state, expiry and all three epochs remain required for workflows using credentials. A user-only admission must not weaken or silently replace that check.

### 2. Separate submitted pre-auth cookie through protected phases

**Existing custody**
- `FeatureCommand` holds code and pre-auth bearer separately from scalar fields (`feature_command.go`).
- `MapPreAuth` / `MapPreAuthWithAuthority` freeze the actual submitted canonical bearer separately from the authenticated attempt envelope (`otp/postgres/preauth_cookie.go`).
- Coordinator fingerprinting includes its domain-separated keyed commitment (`postgres/codec.go:129`); changed cookie means changed command intent.
- Original pre-auth binding carries tenant, action, origin, callback, nonce, remember policy, policy ID/version and original issued/expiry instants.
- `authenticatePreAuth` derives the commitment under the exact original retained cookie-key version and compares it without inspecting OTP code (`preauth_admission.go`).
- Session PostgreSQL owns issued/reserved/consumed pre-auth state and the original record, cookie digest and command/fingerprint/generation checks. Native OTP must borrow its sole coordinator transaction.

**Actual missing contract**
- Attempt and reservation/rebind can obtain `NativeCommand` from their exact-owner handles.
- Apply/finalize expose only **code-free `CommandSpec`**, primary query/time, binding and finalization SQL (`postgres/otp_execution.go`).
- They internally retain the original `Command`, but cannot obtain its separately submitted bearer.
- Consequently the current `authenticatePreAuth`/`admitPreAuth` bridge cannot perform actual-cookie admission through protected apply/finalize without exposing the full code-bearing command.

The minimum consumed contract is **exact-owner, live-phase access to only the separately submitted pre-auth bearer**, alongside existing immutable `CommandSpec` and reservation binding. It must not expose OTP code, introduce full `NativeCommand` access in protected phases, or infer possession from stored digest fields.

### 3. Handle lifetime, denial and recovery invariants

- Exact ownership is the registered comparable contributor instance; foreign/nil owners deny.
- Sealed handles inherit coordinator-owned context, deadline, generation, fingerprint, registry and row/SQL budgets. Copies share callback lifetime; serial use and closed rows are mandatory.
- Attempt allows only initial ownership; reservation permits exact predecessor generation transfer; apply/finalize require current generation and zero predecessor.
- Dispatch invalidates callback state and clears retained command afterward. Returned handles must deny **without poisoning a later invocation**.
- A failure on a **live** admission invocation—including nil/cancelled context, malformed/foreign owner or outstanding-cursor rejection—must poison Work even if ignored. Closing the cursor cannot restore authority. Existing `otpAuthorityFailure`/`authorityScope` establish this pattern.
- Original expectations remain authenticated frozen intent. Database time is a check sample, never a renewed lease.
- Current `admitPreAuth` selects issued admission with nil custody and reserved admission with exact live current-generation custody. It deliberately does not authorize takeover or release.
- Rebind retains original cookie/flow/epochs and uses exact predecessor transfer; it cannot compare or charge the OTP again.
- Release remains possible after revocation/expiry when coordinator proves noncommit. It must not require renewed live authentication authority.
- Recovery returns historical redacted classification under current query authorization, not a replacement submission, new epoch snapshot or pre-auth authority.

### 4. Required bounded evidence for these prerequisites

1. Credential-free active user with exact three epochs succeeds through each live native phase; blocked/missing/foreign user and each stale epoch deny and poison.
2. Full-width epochs decode exactly; malformed/corrupt primary values fail closed; aggregate `User.Version` is not substituted.
3. User share locking excludes concurrent lifecycle changes through transaction completion; cancelled lock wait denies.
4. Existing credential-based admission retains disabled/expired/reference/version checks.
5. Apply/finalize bearer access rejects zero/foreign/returned handles, preserves original submitted bearer and exposes no OTP code.
6. Actual wrong cookie, retired named key, altered original flow, expired flow and mismatched reserved generation deny before protected effects.
7. Ignored live admission denial cannot commit; retained handles cannot poison later callbacks.
8. Retry/takeover preserve original flow, cookie and epochs; recovery/release do not renew authority or repeat OTP comparison.

**Disposition:** two concrete native prerequisite seams are missing. This report does not certify complete OTP activation, product completion or executed verification.

### Native report — /root/otp_user_admission_compat_918965

APPROVE — Go compatibility review of HEAD `918965f0a05adc9e73fa8f7474091ff586150abc` plus the complete nine-file working change.

- Additive public methods/type; existing contributor interfaces and credential-specific admission remain unchanged.
- Passwordless admission compares all three security epochs under the primary user lock, independently of aggregate `User.Version`.
- Cookie access remains exact-owner and live-phase scoped; apply/finalize expose neither OTP code nor full command. Release/recovery gain no accessor.
- Go `1.27.0`, module pins, checksums and schema migrations are unchanged.

No concrete compatibility findings. Source inspection only; no Go commands executed. This is not platform-completion or release verification.

### Native report — /root/otp_user_admission_regression_918965

**APPROVE** — regression audit of `918965f0a05adc9e73fa8f7474091ff586150abc` plus the complete nine-file working change.

No concrete regression or blocking coverage gap found for this prerequisite:
- Four phases exercise credential-free success, each security-epoch drift, inactive/absent users, aggregate-version independence, full-width epochs, retained locks, cancellation, and ignored-denial poisoning.
- Cookie tests exercise actual coordinator dispatch, exact owner, submitted-cookie identity, missing custody, and returned-handle denial.
- Existing credential admission remains unchanged.

Execution results were supplied by the parent; I independently inspected source and tests without running commands that execute Go or services. This approval does not certify native contributor composition, complete login, or release gates.
## Native metered-check regression and compatibility reports

Message Type: FINAL_ANSWER
Task name: /root
Sender: /root/otp_metered_regression_bbdea02
Payload:
## Native regression audit — APPROVE

**Principal:** `/root/otp_metered_regression_bbdea02`, fresh read-only regression reviewer.  
**Boundary:** `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`, branch `feature/identity-platform`; base/current HEAD `bbdea0265ae7855290d0148abcac7ec3eef59fcd` plus the complete dirty 12-path metered-check change below. No writes, Go/service execution, or delegation.

**Findings:** No concrete behavioral defect identified in this bounded implementation.

Decisive implementation/test anchors:

- `otp/postgres/contributor.go:179-272`: original submission custody and current directory/user/credential/authorization plus native pre-auth/session admission precede comparison. Downstream lock waits resample database time and recheck credential/flow validity. Both check and consume use the same challenge/lockout ownership.
- `otp/postgres/metered_check.go:34-50`: native checks require exact record-free operation/schema, check mode, original fields, zero workflow versions and the required audit; unrelated participant effects are rejected. Generic feature application always denies.
- `postgres/otp_attempt.go:283-334`: exact journal replay precedes comparison; pending takeover does not trigger another comparison; all commit-attempt failures remain Unknown.
- `postgres/otp_attempt.go:341-436`: counting, required audit and terminal receipt share the coordinator transaction. Checked success produces no reservation, domain application, lifecycle delivery or session operation.
- `postgres/otp_recovery.go:52-78`: credential-free historical recovery requires authorized query scope and the exact retained OTP attempt plan; it grants no new authentication authority.
- `otp/postgres/contributor_integration_test.go:442-464`: real native pre-auth correct/wrong checks assert counters, replay, one audit and absence of flow consumption/domain/lifecycle effects.
- `otp/postgres/session_admission_integration_test.go:190-341`: native session partitions cover correct/wrong checks, check-only cap, concurrent first duplicates/terminal replay, final correct/wrong exhaustion, disabled checks, audit rollback and fresh-instance credential-free recovery. Final correct exhaustion produces no lockout; final wrong exhaustion extends both shared lockouts.
- Shared `otp/challenge.go:200-259` and `otp/postgres/challenge.go:169-210` retain the total/check budgets, wrong-failure accounting, checked-versus-reserved distinction and bounded issued-row update.

**Evidence scope:** Parent-reported full OTP-adapter integration/race PASS and nine native metered session partitions PASS support this audit; I inspected the test implementation but did not independently execute those commands. Existing denial/consuming regression evidence covers the shared custody, primary-authority and post-lock-wait admission path.

**Still unverified:** actual lost-successful-commit acknowledgement recovery for this metered operation; distinct-command concurrent mixed check/consume traffic; metered-specific post-lock-wait expiry scenarios; release gates, exact coverage/mutation, clean-consumer/source-lock composition and full-unit acceptance. The seeded final-budget fixture proves accounting at that state, not an end-to-end mixed-submission journey. These boundaries prevent a unit-completion or release-readiness claim; they do not reveal a defect in this bounded slice.

### Reviewed path SHA-256

```text
eeeff1802c045053941a9abe0b98052e165ac9da280ad87b8701f8257761f514  otp/postgres/CHANGELOG.md
445d3088465c69ea17920d02027296dde14ba8a7dae30da9068174f0bc73d6c1  otp/postgres/README.md
204f8c46bacc8966db4d1e753ee5710be0b7c6e1577a91544c767502fa27f52d  otp/postgres/contributor.go
330a6b1fe37c95ee475dc3a24d8b04daa4a97aa834b157aa492e2d33b4b028ca  otp/postgres/contributor_integration_test.go
b84312c8b03d3bc064cd3f55cbfe73c188d915e40574bb98e4cd46c17d3ab293  otp/postgres/session_admission_integration_test.go
2877330f8f24e56033fb79b6e138b9a4669a896e6471c8e87fa2ea87c758fb8e  postgres/CHANGELOG.md
9147d8b292b8bc5f8fb9a929cb862fff84b7fe389e730f43c54653761038ce92  postgres/README.md
5d1fc940520a6be44e353236bc52a2422461492c4dd602d657247b56546669f6  postgres/otp_attempt.go
af6b09c5461240196e537285f8ac59267265019acb11fd2bea4175e903f0666d  postgres/otp_recovery.go
173eb047ba8d89c4af8d2e174694540d92d2b3f5c5887d54e610a01ab22c87cd  otp/postgres/metered_check.go
101512efc6a157eb92faef8fb3dcc23528f3f934f3d2b0c2ac8c0ce6ef77b9e4  otp/postgres/metered_check_test.go
aef4b8a495e61ecdd727f44914c38de56a6d2860505edf8ce0b22ab69226eb6f  postgres/otp_metered_check_test.go
```

Message Type: FINAL_ANSWER
Task name: /root
Sender: /root/otp_metered_compat_bbdea02
Payload:
## Native compatibility report

- **Role/principal:** fresh read-only `go_compatibility`, `/root/otp_metered_compat_bbdea02`.
- **Reviewed boundary:** complete dirty metered-check change, including all three untracked files, plus OTP core verifier, registry admission, feature execution, consuming reservation, recovery and module dependency consumers.
- **Base/current HEAD:** `bbdea0265ae7855290d0148abcac7ec3eef59fcd`, branch `feature/identity-platform`.
- **Disposition:** **APPROVE for this development source compatibility slice.** No material compatibility defect found. This is **not** clean-consumer, publication, release or platform-readiness approval.

### Compatibility conclusions

- Existing contributor interfaces and method signatures remain unchanged. `OTPMeteredCheckContributor` is additive; denial-only implementations do not automatically acquire it. Existing disposition values remain unchanged; `OTPAttemptChecked` is appended.
- Native `Contributor` acquires `FeatureExecutionContributor` through explicit methods, declaring two bounded record-free schemas owned by `identity/otp`. Registry owner validation matches `identity/otp/postgres`; no new registration machinery or dependency cycle is introduced.
- The coordinator requires the exact native owner, operation, schema, record policy and check mode. Native validation additionally closes command fields, audit/lifecycle declarations and unrelated participant effects.
- Denial and consuming paths retain their separation. Denial rejects check mode and cannot accept `OTPAttemptChecked`; consuming preparation still requires `OTPConsume`. Native `ApplyFeature` denies generic application.
- Metered success means durable counters/audit/journal history, not authentication, reservation, protected mutation or session effects. Replay precedes comparison; recovery binds the original complete attempt plan, retained journal classification and current query authorization without borrowing current challenge state.
- No new language construct or standard-library API incompatible with declared Go 1.27 was identified.

### Nonblocking documentation finding

`otp/postgres/contributor.go` still describes successful metered checks as under development and `ValidateOTPAttemptCommand` as consuming-only/check-denying. Those comments contradict the new implemented exact-check profile. Update them to describe both modes and preserve the explicit phase separation.

### Remaining gates

- **Pinned clean-consumer compatibility is unresolved.** `otp/postgres/go.mod` still pins `go-identity/postgres` at `84f8cd2317d7`; inspection of that revision shows none of the newly required `OTPMeteredCheckContributor`, `OTPAttemptChecked` or coordinator check APIs. Current source requires the updated kernel source, so disposable `GOWORK` results cannot prove published-module composition. Publish/pin the compatible kernel revision and rerun the clean consumer at the release boundary.
- Source locks, released API baseline qualification, publication and applicable release gates were not established by this review.
- No Go commands or services were executed. Parent-reported full OTP PostgreSQL integration/race PASS (18.403s) and subsequent vet exit 0 are acknowledged as parent evidence, not independently executed evidence.

### Reviewed affected-file SHA256

```text
eeeff1802c045053941a9abe0b98052e165ac9da280ad87b8701f8257761f514  otp/postgres/CHANGELOG.md
445d3088465c69ea17920d02027296dde14ba8a7dae30da9068174f0bc73d6c1  otp/postgres/README.md
204f8c46bacc8966db4d1e753ee5710be0b7c6e1577a91544c767502fa27f52d  otp/postgres/contributor.go
330a6b1fe37c95ee475dc3a24d8b04daa4a97aa834b157aa492e2d33b4b028ca  otp/postgres/contributor_integration_test.go
b84312c8b03d3bc064cd3f55cbfe73c188d915e40574bb98e4cd46c17d3ab293  otp/postgres/session_admission_integration_test.go
173eb047ba8d89c4af8d2e174694540d92d2b3f5c5887d54e610a01ab22c87cd  otp/postgres/metered_check.go
101512efc6a157eb92faef8fb3dcc23528f3f934f3d2b0c2ac8c0ce6ef77b9e4  otp/postgres/metered_check_test.go
2877330f8f24e56033fb79b6e138b9a4669a896e6471c8e87fa2ea87c758fb8e  postgres/CHANGELOG.md
9147d8b292b8bc5f8fb9a929cb862fff84b7fe389e730f43c54653761038ce92  postgres/README.md
5d1fc940520a6be44e353236bc52a2422461492c4dd602d657247b56546669f6  postgres/otp_attempt.go
af6b09c5461240196e537285f8ac59267265019acb11fd2bea4175e903f0666d  postgres/otp_recovery.go
aef4b8a495e61ecdd727f44914c38de56a6d2860505edf8ce0b22ab69226eb6f  postgres/otp_metered_check_test.go
```

Message Type: FINAL_ANSWER
Task name: /root
Sender: /root/otp_metered_regression_bbdea02
Payload:
## Refreshed native regression audit — APPROVE

**Principal:** `/root/otp_metered_regression_bbdea02`, read-only regression reviewer.  
**Reviewed boundary:** complete 12-path dirty metered-check change in `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`, branch `feature/identity-platform`, base/current HEAD `bbdea0265ae7855290d0148abcac7ec3eef59fcd`.

The two changed files contain only the reported additional comment corrections. Executable behavior and all ten other reviewed paths remain unchanged. Corrections accurately describe native metered-check support, exact schema admission and counted-phase handles. Prior behavioral findings, decisive anchors and unverified boundaries remain applicable; no unresolved concrete regression finding.

Parent additionally reports OTP PostgreSQL `go vet ./...` exit 0 after full integration/race PASS 18.403s. No Go execution or QA restart performed here. This is bounded regression approval, not unit completion or release readiness.

### Refreshed complete boundary SHA-256

```text
eeeff1802c045053941a9abe0b98052e165ac9da280ad87b8701f8257761f514  otp/postgres/CHANGELOG.md
445d3088465c69ea17920d02027296dde14ba8a7dae30da9068174f0bc73d6c1  otp/postgres/README.md
93ec1e0c8ecd4ec2a235afc751fb033e7f42a74a002f073733b0e3680f870377  otp/postgres/contributor.go
330a6b1fe37c95ee475dc3a24d8b04daa4a97aa834b157aa492e2d33b4b028ca  otp/postgres/contributor_integration_test.go
b84312c8b03d3bc064cd3f55cbfe73c188d915e40574bb98e4cd46c17d3ab293  otp/postgres/session_admission_integration_test.go
2877330f8f24e56033fb79b6e138b9a4669a896e6471c8e87fa2ea87c758fb8e  postgres/CHANGELOG.md
9147d8b292b8bc5f8fb9a929cb862fff84b7fe389e730f43c54653761038ce92  postgres/README.md
8c27ed2a1952c2a6921ee6bcfbc732a5f8ea749e6c8df85fdea7a46c2c57883d  postgres/otp_attempt.go
af6b09c5461240196e537285f8ac59267265019acb11fd2bea4175e903f0666d  postgres/otp_recovery.go
173eb047ba8d89c4af8d2e174694540d92d2b3f5c5887d54e610a01ab22c87cd  otp/postgres/metered_check.go
101512efc6a157eb92faef8fb3dcc23528f3f934f3d2b0c2ac8c0ce6ef77b9e4  otp/postgres/metered_check_test.go
aef4b8a495e61ecdd727f44914c38de56a6d2860505edf8ce0b22ab69226eb6f  postgres/otp_metered_check_test.go
```
## Refreshed native metered compatibility

Message Type: FINAL_ANSWER
Task name: /root
Sender: /root/otp_metered_compat_bbdea02
Payload:
## Refreshed native compatibility report

**Principal:** `/root/otp_metered_compat_bbdea02`, read-only `go_compatibility`.

**Reviewed boundary:** complete metered-check dirty change, all 12 affected paths including three untracked files, and previously inspected registry/phase/recovery/dependency consumers. Base/current HEAD remains `bbdea0265ae7855290d0148abcac7ec3eef59fcd`.

**Disposition: APPROVE for the development source compatibility slice.** No unresolved source compatibility findings.

The documentation finding is **closed**: contributor comments now accurately describe metered checks, consuming-or-exact-native-check validation and no authentication proof; shared phase comments no longer describe counted work as denial-only. The two changed hashes reflect comment-only corrections; the other ten affected files retain their previously reviewed hashes. No executable or test changes were identified.

Existing interfaces and enum values remain compatible; optional check acquisition, exact native schema ownership, separate denial/consume execution, journal replay and authorized historical recovery conclusions remain unchanged.

**Release limitation remains:** `otp/postgres/go.mod` pins kernel revision `84f8cd2317d7`, which lacks the added check interface/constants/APIs. Compatible publication/source pinning and pinned clean-consumer proof remain outstanding. Source-lock qualification, API release baseline and remaining release gates are not certified. No Go execution occurred in this audit; parent-reported integration/race and vet passes remain parent evidence, not independent execution. No platform-readiness claim.

### Current affected-file SHA256

```text
eeeff1802c045053941a9abe0b98052e165ac9da280ad87b8701f8257761f514  otp/postgres/CHANGELOG.md
445d3088465c69ea17920d02027296dde14ba8a7dae30da9068174f0bc73d6c1  otp/postgres/README.md
93ec1e0c8ecd4ec2a235afc751fb033e7f42a74a002f073733b0e3680f870377  otp/postgres/contributor.go
330a6b1fe37c95ee475dc3a24d8b04daa4a97aa834b157aa492e2d33b4b028ca  otp/postgres/contributor_integration_test.go
b84312c8b03d3bc064cd3f55cbfe73c188d915e40574bb98e4cd46c17d3ab293  otp/postgres/session_admission_integration_test.go
173eb047ba8d89c4af8d2e174694540d92d2b3f5c5887d54e610a01ab22c87cd  otp/postgres/metered_check.go
101512efc6a157eb92faef8fb3dcc23528f3f934f3d2b0c2ac8c0ce6ef77b9e4  otp/postgres/metered_check_test.go
2877330f8f24e56033fb79b6e138b9a4669a896e6471c8e87fa2ea87c758fb8e  postgres/CHANGELOG.md
9147d8b292b8bc5f8fb9a929cb862fff84b7fe389e730f43c54653761038ce92  postgres/README.md
8c27ed2a1952c2a6921ee6bcfbc732a5f8ea749e6c8df85fdea7a46c2c57883d  postgres/otp_attempt.go
af6b09c5461240196e537285f8ac59267265019acb11fd2bea4175e903f0666d  postgres/otp_recovery.go
aef4b8a495e61ecdd727f44914c38de56a6d2860505edf8ce0b22ab69226eb6f  postgres/otp_metered_check_test.go
```
## Accepted blind native metered final review

Message Type: FINAL_ANSWER
Task name: /root
Sender: /root/otp_metered_blind_final_bbdea02
Payload:
**APPROVE for this local development commit. No findings.**

Reviewer identity: `/root/otp_metered_blind_final_bbdea02`, fresh isolated read-only reviewer. Applied `requesting-code-review`, its complete independent contract and relevant boundary/topology references, `testing-coverage`, and `verification-before-completion`. No writes, Go/service execution, delegation, sibling-report inspection, or external mutation.

### Reviewed boundary

Repository: `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`  
Branch: `feature/identity-platform`  
Base/current HEAD: `bbdea0265ae7855290d0148abcac7ec3eef59fcd`  
Scope: complete nine tracked dirty-file diffs plus three untracked files, including code, tests, both READMEs and changelogs. Also traced original custody, consuming reservation/execution, feature registry/admission, command claim/journal/replay, authorized recovery, session/pre-auth admission, primary authorization, counter persistence, schema guards and shared lockouts.

Raw platform authority and both selected OTP goals were read. Assessment is for the coherent metered development tranche, **not** finished OTP units, login, platform acceptance or release readiness.

### Decisive closure

- **Authority and phase separation:** `otp/postgres/metered_check.go:34` validates exact record-free check declarations, original custody and effect-free plans. `postgres/otp_attempt.go:255` requires the additive capability and exact native owner. Existing consuming preparation still rejects `OTPCheck` (`postgres/otp_reservation.go:154`); native `ApplyFeature()` always denies.
- **Metering and exhaustion:** `otp/postgres/contributor.go:180` retains original bearer and primary-authority admission before shared locks/challenge comparison. Core `otp/challenge.go:230` charges separate check and shared total allowances; persistence CAS/schema guards retain bounded monotonic counters. Challenged correct final exhaustion: it becomes exhausted without extending lockouts; wrong exhaustion extends them.
- **Atomicity and replay:** `postgres/otp_attempt.go:294` locks claim/journal and matches frozen intent before callback. Terminal replay bypasses comparison. Counter changes, lockout extension, required audit and terminal receipt share the sole transaction; audit failure rolls back every provisional effect. Challenged simultaneous identical first submissions: global-claim exclusion makes the later submissions terminal replays.
- **Recovery and secrecy:** `postgres/otp_recovery.go:52` binds original scalar attempt and current query authorization to historical state without bearer/code/comparison. Shared `commandReceipt()` enforces tenant, actor, target, operation and stored-plan/result agreement. Changed attempts deny; retained results do not refresh authority.
- **Oracles:** New native integration cases independently observe physical counters, challenge/session/pre-auth states, audit cardinality, zero domain/lifecycle effects, retained results and lockout state. Existing denial/consuming paths remain on the shared implementation with separate mode guards.

`git diff --check` passed. Runtime results are **parent-supplied**, not independently rerun: OTP PostgreSQL full integration/race and vet; nine metered-session cases; pre-auth correct/wrong; kernel full race; top-level OTP integration/race and vet; nested kernel OTP denial/reservation/execution integration/race.

Exact coverage/mutation, published source pins/clean consumers, real-driver metered commit ambiguity, and full OTP/login/platform/release acceptance remain unverified and are not certified by this approval.

### Frozen file SHA-256

```text
eeeff1802c045053941a9abe0b98052e165ac9da280ad87b8701f8257761f514  otp/postgres/CHANGELOG.md
445d3088465c69ea17920d02027296dde14ba8a7dae30da9068174f0bc73d6c1  otp/postgres/README.md
93ec1e0c8ecd4ec2a235afc751fb033e7f42a74a002f073733b0e3680f870377  otp/postgres/contributor.go
330a6b1fe37c95ee475dc3a24d8b04daa4a97aa834b157aa492e2d33b4b028ca  otp/postgres/contributor_integration_test.go
b84312c8b03d3bc064cd3f55cbfe73c188d915e40574bb98e4cd46c17d3ab293  otp/postgres/session_admission_integration_test.go
2877330f8f24e56033fb79b6e138b9a4669a896e6471c8e87fa2ea87c758fb8e  postgres/CHANGELOG.md
9147d8b292b8bc5f8fb9a929cb862fff84b7fe389e730f43c54653761038ce92  postgres/README.md
8c27ed2a1952c2a6921ee6bcfbc732a5f8ea749e6c8df85fdea7a46c2c57883d  postgres/otp_attempt.go
af6b09c5461240196e537285f8ac59267265019acb11fd2bea4175e903f0666d  postgres/otp_recovery.go
173eb047ba8d89c4af8d2e174694540d92d2b3f5c5887d54e610a01ab22c87cd  otp/postgres/metered_check.go
101512efc6a157eb92faef8fb3dcc23528f3f934f3d2b0c2ac8c0ce6ef77b9e4  otp/postgres/metered_check_test.go
aef4b8a495e61ecdd727f44914c38de56a6d2860505edf8ce0b22ab69226eb6f  postgres/otp_metered_check_test.go
```
## Native exact-head metered delivery audit

Message Type: FINAL_ANSWER
Task name: /root
Sender: /root/otp_metered_delivery_eafc154
Payload:
**APPROVE — local committed DEVELOPMENT state only.**

Native delivery auditor: `/root/otp_metered_delivery_eafc154`.

- Exact head `eafc154d7a466dc864613c3128d3dad1f697bf70`; parent `bbdea0265ae7855290d0148abcac7ec3eef59fcd`; clean `feature/identity-platform`.
- **12/12 committed blob SHA256 matches** accepted native blind reviewer `/root/otp_metered_blind_final_bbdea02`.
- Inspected native mapper/QA, original and refreshed regression/compatibility, and accepted correctness reports. Refreshed audits and final review approve this bounded slice; comment finding closed.
- Commit scope and both additive changelog entries verified; `git diff --check` passes.
- Both named temporary directories and exact container are absent; existing image untouched.
- Runtime verification remains parent-executed: full OTP PostgreSQL integration/race PASS **18.403s**, vet exit **0**, nine metered partitions PASS **2.668s**. No Go execution by this auditor.

**Not full-unit readiness:** source-pin/clean-consumer composition, exact coverage/mutation and release gates, real-driver metered commit ambiguity, mixed submissions/post-wait partitions, and full login/shared-issuer acceptance remain pending. No push, CI, release, deployment or production certification. Counts remain **5/67 integrated, 1/67 verified**; full goal incomplete.
## Native issuance/resend mapping and bounded QA

Message Type: FINAL_ANSWER
Task name: /root
Sender: /root/otp_issue_qa_eafc154
Payload:
## Native QA plan

- Principal: `/root/otp_issue_qa_eafc154`, fresh read-only QA planner.
- Skill: `$qa-test-planner`, read completely and applied in one bounded pass.
- Repository: `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`.
- Branch: `feature/identity-platform`.
- Exact inspected base/head: `eafc154d7a466dc864613c3128d3dad1f697bf70`.
- Scope: native OTP issuance/resend and its affected attempt/delivery boundaries.
- No files changed, Go commands executed, services accessed, or delegates used.

### Existing proving boundaries

`otp` already supplies validated profiles, bounded unbiased generation, continuation/remember binding, and versioned keyed code/scope digests. Existing tests cover sampling, borrowed-buffer erasure, cancellation, key versions and scope differentiation.

`otp/postgres` already owns challenge custody and native attempted verification/reservation. Its existing integration fixtures insert issued challenges directly and mark issuance commands committed. Those fixtures do **not** prove production issuance, resend limits, or OTP-plus-delivery atomicity.

`delivery/postgres.Contributor.Prepare`/`Stage` supplies encrypted effects inside the original coordinator `Work`. The standalone `Queue.Enqueue` path executes its own delivery command; successful sequential OTP storage followed by queue enqueue would not prove the required single-commit contract.

## Smallest sufficient behavioral coverage

| Boundary and scenario | Independent observable assertion | Realistic defect detected |
|---|---|---|
| **Core unit:** issuance preparation for all six purposes, valid minimum/maximum profiles, invalid adjacent bounds, zero/nil inputs, malformed/oversized destination and identifiers | Valid preparation preserves exact declared purpose/profile; rejected preparation produces no code-bearing output or invoked downstream work | Silent defaults, purpose coercion, or oversized input reaching allocation/crypto |
| **Core unit:** generated code enters both scoped digest and protected delivery preparation | Independently calculate expected HMAC from a captured request-memory code and complete scope; digest and sender content agree; borrowed bytes are erased after success, error and panic | Different code delivered from stored digest; plaintext retained; digest missing a scope field |
| **Native composition:** pre-auth signin/MFA and authenticated-session verification/change/step-up | Stored challenge validates only under the original flow; swapped tenant, subject, destination, target, purpose, authorization variant or authority/policy version denies without modifying the original row | Freshly reconstructed authority accepted in place of original continuation |
| **Remember binding:** both persistent and nonpersistent signin policies, including resend | Original choice and every lifetime parameter survive resend unchanged; changed choice or lifetime conflicts/denies | Resend upgrades persistence or extends the selected session lifetime |
| **PostgreSQL limits:** resend interval −1 µs, exact equality, +1 µs; send count `N−1`, `N`, `N+1`; hour-window boundary | Successful new send advances one family counter; denied sends insert neither challenge nor effect; exact command replay advances nothing | Off-by-one admission, application-clock enforcement, replay charged twice |
| **Limit-scope attacks:** change challenge ID, authorization transaction, digest-key version or workflow target while tenant/purpose/subject/destination remain the same | None creates a new family allowance; independently scoped families remain unaffected | Full challenge scope mistakenly used as the send-limit key |
| **Concurrent resends:** two distinct commands at one remaining allowance | Exactly one new committed challenge/effect and one counter increment; loser has stable enumeration-safe outcome | Check-then-insert race exceeds interval/hour cap |
| **Replace policy:** predecessor issued; successor succeeds, rolls back, or has unknown commit | Until proven commit predecessor remains valid; after commit predecessor is revoked and successor issued; rollback preserves predecessor, counter and existing effect | Early revocation, split enqueue, two valid replace-policy challenges |
| **Retain policy:** two issued challenges and subsequent checks/consumption | Earlier challenge retains original expiry and counters; shared send and exhaustion lockout protections persist | Resend renews earlier expiry or bypasses family lockout |
| **Verify/resend race:** correct consume versus replacement; wrong final attempt versus resend; separately metered check versus replacement | Results match one serial ordering; a reserved predecessor cannot be revoked/replaced; exhausted authority cannot be bypassed; counters never exceed bounds | Resend races the existing participant or releases an unknown reservation |
| **Store/effect/journal atomicity:** fail challenge insert, delivery staging/persistence, required audit, or terminalization | Challenge, family count, predecessor transition, encrypted effect and committed command are all present together or none of the proposed changes commit | Partial challenge issuance, orphan delivery, missing required audit |
| **Replay/conflict:** same command and exact input; same command with changed destination/profile/authorization/template/provider/remember | Exact replay returns the same safe challenge/effect identity without regeneration or enqueue; changed input conflicts with unchanged persisted state | Replay creates a fresh secret/effect or fingerprint omits behavior-bearing fields |
| **Unknown outcome:** disconnect/cancel before commit versus commit acknowledgement loss; restart and primary recovery | Ambiguous commit returns `Unknown`, no plaintext/new code; recovery finds original challenge/effect or authoritative non-commit; no replacement command or duplicate counter increment | Timeout guessed as rollback, re-generation on unknown, recovery using only challenge existence |
| **Risk admission:** each selected identifier/IP/device dimension, exhaustion and counter-store failure | Denial/unavailability leaves no new issuance effect; public denial envelope does not disclose account/challenge existence; retries respect existing durable admission | Missing bucket, fail-open outage, enumeration through error/result differences |
| **Delayed delivery:** expired planned effect, duplicated delivery, bounced/permanent rejection, unknown provider outcome, replacement while predecessor effect is already submitted | Expired/superseded planned effects are not sent and ciphertext is erased; delayed old code cannot verify under replacement policy; unknown delivery does not imply issuance rollback or authorize unsafe resubmit | Stale code becomes authoritative, expiry renewed, ambiguous provider effect sent again |
| **Secret custody:** query challenge, journal, counters, audit, outbox and delivery tables after issuance and terminal delivery | OTP state contains only keyed versioned digests; bearer exists only as authenticated delivery ciphertext; AAD binds original tenant/purpose/subject/command/effect/template; terminal ciphertext disappears atomically | Recoverable OTP stored outside delivery custody, missing AAD binding, redaction leak |
| **Cancellation/resource limits:** precancelled request, cancellation after family-lock wait, expired original flow while waiting, invalid numeric/version capacity | No stale positive authorization survives the wait; outcome distinguishes proven rollback from ambiguity; connections/rows and owned resources close | Using pre-lock authority/time, integer wrap, leaked wait/connection, cancellation rewritten as success |

### Test-level and verification selection

1. Add the first red regression at the **native issuance API** demonstrating a committed issuance whose generated code can subsequently be checked through the existing native path. Assert database state and one encrypted delivery effect; direct SQL challenge fixture insertion cannot establish this acceptance criterion.
2. Use deterministic **unit tests** for preparation, bounds, exact binding, code/digest agreement and redaction. Reuse existing generation/digest tests rather than expanding sampling coverage ceremonially.
3. Use a bounded **real PostgreSQL composition suite** for counters, replacement/retention, replay, failure atomicity and recovery. Extend existing task-owned OTP database support and existing delivery participant composition.
4. Use deterministic two-connection barriers for resend/attempt races; assert final persisted states rather than callback counts or sleeps. Run those affected paths with `-race` at the concurrency gate.
5. Exercise delayed delivery through the existing owned delivery worker using a bounded capturing sender. This proves owned payload/lifecycle behavior, **not** real provider interoperability.
6. Run affected `otp`, `otp/postgres`, and any changed kernel/delivery package checks; include direct consumers if their contracts change. Every Go command needs a disposable `GOCACHE` with cleanup on all exits. Full coverage/mutation and other mandatory release gates remain required before final verification, not extra pre-write approval stages.

### Decisions to preserve

- Reference configuration uses “superseded” for predecessor challenges, while the transaction contract and existing SQL use the closed terminal state `revoked`. Test the **observable supersession behavior** and explicitly document its `revoked` representation; do not invent an additional state solely to match wording.
- A family send counter must remain stable across challenge IDs and code-key rotation. Complete per-challenge scope digests alone cannot prove that invariant.
- The reference replacement rule does not generally permit superseding a submitted/unknown predecessor **delivery effect**. Old delivery may arrive, but its code must lose authority after the replacement commit.

**Disposition:** no contradictory requirement or reproducible design defect blocks implementation. Production native issuance/resend coverage is missing work, not a planning rejection. No current behavioral pass or release-readiness claim is made.

Message Type: FINAL_ANSWER
Task name: /root
Sender: /root/otp_issue_mapping_eafc154
Payload:
**Native mapping — `change_mapper`, read-only.** Source branch `feature/identity-platform`, base/head `eafc154d7a466dc864613c3128d3dad1f697bf70`. No files changed, services accessed, or Go checks run.

### Owned root preparation checkpoint — source 8ce205f

Native REDs: `cf468b` missing API; `2caacc` value-copy redaction;
`ccf016` renderer-pointer-slot aliasing. Final bounded root module race and
vet completed with exit0 at `f1fb2c`, session7441 (1.240s).
Native regression/Go compatibility `/root/otp_preparer_regression_e64c4a3`
and isolated complete final review `/root/otp_preparer_final_blind_e64c4a3`
APPROVE; all seven committed blobs match their final SHA256. Previous final
principal `/root/otp_preparer_final_e64c4a3` disclosed incidental exposure to
a sibling report and is not used as independent approval. No production
patch was needed for that observation; native test outputs were narrowly
structurally filtered for the replacement review. Public native Send,
adapter/risk/recipient composition and all release gates remain pending.

### Prepared-feature development checkpoint — source e64c4a3

Native parent evidence: RED `6cfcff`; affected race/unit `b3b9e9`
(3.355s/0.334s); final root-key-rotation restart race `e4112c` (1.848s);
intentional execution-binding removal regression `f8ab40` detected.
Independent native reports `/root/prepared_feature_regression_3ebd6d2` and
`/root/prepared_feature_final_3ebd6d2`: APPROVE, exact nine committed blobs
match their final hashes. Real OTP/public Send and release certification are
not established. Fresh next-slice mapper `/root/otp_stable_send_owner_map`
and QA `/root/otp_stable_send_owner_qa` identify stable-owner identity,
invocation-local custody, inaccessible core preparation, complete original
request binding and still-needed risk/recipient composition. No new workflow
or receipt owner is introduced.

### Available consumed boundaries

- **Kernel owns transactions, command identity, replay and commit recovery.** `Coordinator.PrepareCommandPlan` → `ReserveCommand` → `Begin` → `ApplyFeature` → `Commit`/`Abort`; historical observation uses `RecoverFeatureCommand`. Native adapters receive bounded `Work.Query`, `Exec`, `DatabaseTime` and authority admissions. `Commit` already persists delivery, audit, lifecycle and the command receipt in the single domain commit; acknowledgement failure returns `Unknown`. Sources: `postgres/feature_execution.go:27,174`, `postgres/domain.go:47,239,267,359`, `postgres/recovery.go:63`, `postgres/work.go:102,129,185`.
- **OTP core already generates and binds codes.** `CodeProfile.Generate` lends unbiased bounded bytes synchronously and clears its buffer; it explicitly does **not** issue/store. `Verifier.Digest` and `ScopeDigest` provide versioned keyed code/scope digests. `NewContinuation` binds the exact destination, purpose, challenge, target, policy/generation and original closed flow context; signin/pre-auth MFA preserve the session-owned remember policy. Sources: `otp/generation.go:9,27`, `otp/profile.go:44,63`, `otp/digest.go:90,266`, `otp/scope_digest.go`, `otp/authorization.go:153,168`.
- **Existing OTP PG participant is consumption/check authority, not issuance.** `Contributor` descriptor is `identity.otp.store`; it implements counted denial, metered checks, reservation, protected checking/finalization, release and recovery. Its declared feature variants are only user/anonymous-flow `identity.otp.check`; generic `ApplyFeature` intentionally denies. Sources: `otp/postgres/contributor.go:34,120,191`, `otp/postgres/metered_check.go:23,58`, `otp/postgres/consuming_execution.go:20,46`, `otp/postgres/native_flow.go`.
- **Delivery can already join OTP’s original command atomically.** `delivery/postgres.Contributor.Prepare` encrypts bounded rendered intent before domain work; `.Plan()` freezes metadata; `.Stage()` stages ciphertext into the exact owning `Work`; `PersistDelivery` and `RecoverParticipant` execute under the existing coordinator. Calling public Queue/`EnqueueService.Enqueue` separately would create a different delivery command and would not prove OTP store+delivery atomicity. Sources: `delivery/postgres/contributor.go:127,148,253,286,296`, `delivery/postgres/enqueue_execution.go:67`, `postgres/work_delivery.go:12,62`.
- **Session original-flow authority is available.** Native OTP admission already uses complete original pre-auth/session custody, separate cookie/token possession, current primary authority and after-wait expiry/freshness checks. Session offers real pre-auth issuance and primary `RequirePreAuth`/`RequireSession` contracts; issuance should require the original flow without consuming it merely to send. Sources: `session/postgres/preauth_issue_execution.go:159`, `otp/postgres/native_flow.go:72,85`, `otp/postgres/session_admission.go:83,103`.
- **Risk has a significant variant distinction.** `identity.otp.send` is the **Authentication / BeforeCredential** profile with tenant/network/device/identifier dimensions. Existing `PrepareCommandCounters` / `ReserveCommandCounters` supports only **CredentialMutation / Privileged / ReserveWithCommand**. It cannot be reused for OTP sends unchanged. Use the existing authentication evaluation/velocity contracts, not a fabricated protected-counter profile. Sources: `risk/catalog_generated.go:201`, `risk/policy.go:195`, `risk/command_counter.go:80`, `risk/postgres/counter_command.go:22,108`, `risk/evaluation.go:224`.

These are implemented source contracts with existing focused/composition test files—not freshly executed proof or release verification. Relevant existing tests include `delivery/postgres/composition_integration_test.go:32`, `otp/postgres/contributor_integration_test.go`, `otp/postgres/session_admission_integration_test.go`, and `session/postgres/preauth_issue_execution_integration_test.go`.

### Missing issuance boundary

1. **No public OTP `Service.Send`, native send command/codec, executable send declaration, or owned send executor.** Catalog authority specifies `Send(context.Context, SendRequest) (SendResult, error)`, keyed operation identity, enumeration-safe delivery state and optional challenge ID; implemented core currently supplies generation, digests, continuation and attempts. Control sources: `PUBLIC_CONTRACTS.json:7899–7930,89782–89870`, `API_OPERATIONS.md:359`.
2. **No native challenge insertion or send-family/window enforcement.** Migration stores immutable resend/max-send/replacement policy but explicitly says issuance/send limits are absent. Existing consumption tests seed issuance journal/challenge rows directly; those fixtures are not the missing issuance path. Sources: `otp/postgres/migrations.go:19,69`, `otp/postgres/contributor_integration_test.go:387–391`, `otp/postgres/session_admission_integration_test.go:144–152`.
3. **No atomic resend interval/hour-cap/replacement family state.** The authoritative owner must serialize family sends and challenge replacement against existing challenge consumption. Replace may revoke only `issued`; `reserved` cannot be revoked/replaced without owning-command recovery. Existing shared destination/subject lockouts must be initialized through the adapter’s private `initializeLockouts`, then preserve its established ordering. Sources: `otp/postgres/lockout.go:58,108,134`, control `TRANSACTION_CONTRACT.md:658–664`.
4. **Unknown issuance recovery remains unimplemented.** Persist issuance command/fingerprint, family decision and protected challenge+delivery together; recover by the original coordinator command, not regeneration, a new command, or delivery payload presence. Existing challenge readers already require the issuance journal to be committed. Source: `otp/postgres/challenge.go:73–77`.
5. **Bounded cleanup remains missing.** Initial migration intentionally retains tombstones and has no deletion API. Cleanup must expire untouched `issued` rows in indexed DB-time batches, exclude unresolved `reserved`, and retain terminal keyed digest/key-version evidence until referenced keys retire. Sources: `otp/postgres/migrations.go:69–76`, control `TRANSACTION_CONTRACT.md:726–733`.
6. **Public Send still needs owning-workflow resolution.** Canonicalization/identifier uniqueness belong to email/phone workflows; OTP’s raw `DestinationInput` is not such authority. Full enumeration safety must cover absent accounts and no-op equivalent encrypted delivery, not only redact adapter errors. No complete native email/password/phone/MFA owning callbacks were established by this mapping.

### Next real package write

Implement **native issuance/resend in `otp/postgres`**, using the existing kernel **`FeatureExecutionContributor` path for `identity.otp.send`**, with the minimal OTP-owned public Send/issuance input/result contract required by that path. Keep challenge/family windows, original issuance custody and recovery owned by this adapter; stage the existing delivery contributor into the same `Work`. The existing `identity.otp.store` participant remains the sole consuming-state authority—do not introduce an application-written Store, second journal, new coordinator, or generalized issuance framework.

This path can produce a real public-send → stored digest + encrypted delivery → stable result/recovery composition using already available contracts. It need not wait for downstream reference-server or all-package release evidence. It must not be called complete login or full OTP completion until owning callbacks, risk composition and cleanup are actually implemented and exercised.
