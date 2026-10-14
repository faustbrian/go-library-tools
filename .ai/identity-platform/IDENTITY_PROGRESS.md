# Identity Platform Progress

Last updated: **2026-10-14 03:01 UTC / 2026-10-14 06:01 Europe/Helsinki**.

## Maintenance contract

The primary implementation agent MUST maintain this file at material delivery,
verification, review or scope changes, and at least every six hours while actively
working. Next scheduled freshness boundary: **2026-10-14 09:01 UTC**. This is an
active-session obligation, not a background automation or an idle-session promise.
Carry the last update and next boundary through continuation/context handoff.

Keep the complete 67-unit scope. Update actual consumer outcomes, per-unit
qualification, current slice, blockers and next action. Do not count helpers,
commits or procedural artifacts as completed units. Preserve distinctions between
partial implementation, integrated units, fully verified units and publication.
Do not rerun unchanged checks just to refresh this document. Reconcile it against
live source and the execution ledger; the document itself is not proof of delivery.

## Summary

- Whole-platform outcome progress is not measured by a weighted percentage;
  the historical 5–10% estimate is retired as a dependable delivery measure.
  Integrated and qualified fractions below retain their original meaning.
- Active goal: bounded execution repair plus clean reviewed checkpoint, not
  resumption of the full-platform goal. IMPLEMENTATION_GOAL.md is drafted with
  39 fixed consumer milestones; it remains inactive until transition handoff.
- Reviewed consumer checkpoint (02:28 UTC): real native pending secondary-OTP
  enrollment, encrypted queue/captured delivery and confirmation, password
  challenge, selected HTTP send and delivered-code HTTP verification, session/
  CSRF delivery, public session read, credential-free replay and factor-version
  invalidation PASS36.047s on Go1.27.2/race. This selected email profile is now
  reviewed and committed, not full MFA qualification or the complete M16 family.
  Native relocated acceptance PASS37.043s; HTTP-MFA race PASS11.034s;
  reference race PASS7.605s. Cross-owner tests were moved downstream to
  reference/integration/password, avoiding reverse module imports; the upstream
  password adapter still compiles integration-tagged tests (PASS0.307s, no tests
  run). New fixture errors in request envelope, transferred opaque-proof aliases
  and missing atomic quota migration were corrected without production changes.
- Final selected rollback, both lost-acknowledgement sides and contention pass.
  Earlier failures exposed test fault-profile leakage; corrected fixtures do
  not weaken production leases. Exact attribution of every earlier unknown
  result remains unproven and is preserved in the ledger/native review.
  Graph/bootstrap, candidate external build, affected defaults/direct consumers,
  fuzz and vet pass within their stated boundaries. The candidate build uses
  temporary source replacements, not published-release compatibility evidence.
- Runtime checkpoint `55d2e6a` is integrated and pushed to main at `ef377dd`;
  both complete-diff and bounded integration reviews APPROVE. Existing main
  policy/automation is retained; structural planning and required secret checks
  pass locally. Hosted Required failed on historical secret-detector matches,
  not runtime tests; narrow shared Tools classification/release authority is
  pending. Control main is locally committed at362f94d, but its push was rejected
  because exact-head Required is expected. A CI-only carrier/PR needs authority
  under the current no-PR goal. Remaining transition work is that protected-main
  delivery and final cleanup. No whole-platform completion
  or released/product qualification is claimed; the revised goal stays inactive.
- Recorded integrated units: **5/67 (7.5%)**, all primitive extensions.
- Recorded fully verified units: **2/67 (3.0%)**.
- Fully completed product units: **0/61**; usable slices span unfinished units.
- Scope: all **67 units, 19 journey categories, 10 shared-issuer cases**.

The checkpoints below are historical. The current transition checkpoint above
supersedes their next-action and live-resource statements, not their results.
- Current transport checkpoint (00:37 UTC): independently selected native
  password-origin `POST /v1/mfa/otp-verify` handler, closed wire schema and
  reference selection are written. Browser/Origin/callback/source quotas remain
  guarded; no caller factor/recipient/remember override or TOTP fallback is
  accepted. Fresh matching password/OTP session custody alone delivers cookie/
  CSRF; replay, conflict, unknown, cancellation and expired final write withhold
  credentials. Native quota-source real-PG journey PASS37.379s proves original
  subject/child/signed interval, foreign-input denial, conflicting historical
  code denial, counted-denial/consumed-child replay and unchanged lifetime after
  factor-epoch changes. A canceled-result implementation defect was reproduced
  and fixed; explicit unselected variant admission and recovery-only graph
  omission were also reproduced and corrected. Final affected HTTP-MFA/reference
  race PASS9.532s/PASS7.467s; bounded hostile OTP JSON fuzz PASS6.272s.
  This is transport/source acceptance, NOT the complete selected native
  enrollment/send/delivery/HTTP-login journey. Next: OTP sending/enrollment
  transport and real selected journey, clean-consumer test placement, final
  coherent-slice review/local commit. Current source remains uncommitted
  at1928f8e; all unit counts/full scope unchanged. Owned loopback PostgreSQL
  container/volume, synthetic credentials, supplier workspace and Go caches
  removed; user services/images/data preserved. No full-repository rerun.
- Current composition checkpoint (00:20 UTC): `PasswordSigninConfiguration.OTP`
  and `FeatureSet.PasswordOTP` now assemble actual native primary/child/browser/
  session/OTP/delivery owners, enforce matching browser keys and selected method,
  retain graph admission and initialize the actual OTP issuer before readiness.
  Real public ComposeStores→Providers→Workers→Features→migration Apply→Initialize
  bootstrap/retry passes (PASS2.554s). Independent native suffixes preserve the
  shared MFA1–7/optional8–12 allocations; later TOTP selection plans no replay
  and preserves OTP session constraints. Storage predecessors enable no TOTP or
  recovery API. Affected reference default race PASS7.556s; selection/key/history
  checks PASS1.406s. This proves construction/bootstrap, NOT selected enrollment
  or delivered-code login, HTTP or a delivered/reviewed slice. Next: selected
  OTP enrollment/login/HTTP journey and clean-consumer test placement, final
  affected verification and complete-slice review/local commit. Source remains
  uncommitted at1928f8e; all67/19/10 scope and unit counts are unchanged. Owned
  PG/TLS-Valkey containers, PG volume, credentials/certificates, supplier workspace
  and Go caches are removed; user services/images/data are preserved.
- Previous workflow checkpoint (23:58 UTC): actual password → encrypted secondary
  OTP delivery → verification → session/CSRF issuance now includes direct and
  selected public session reads, ordinary-SFA isolation, foreign scope/wrong-code
  denial, exactly-once shared budget, replay/recovery and live factor-epoch denial
  (real PostgreSQL/race PASS34.749s). Session-write rollback and lost commit
  acknowledgements before/after commit pass without credential replay. The
  competing-request defect is now corrected: exactly one fresh credential
  winner, code/browser-free recovery of the losing unattached preparation and
  stable credential-free terminal replay/recovery pass (PASS35.928s). Recovery
  retains original charged attempts and uses the existing declaration plus
  claim/journal exclusion and generation fence, never another journal or release
  of the winning command's OTP/child. Focused historical observation/replay races
  pass (PASS1.390s). All six affected default package race regressions and final
  five-case real PostgreSQL completion batch pass (PASS167.596s). Base-schema
  real-SQL query compatibility without optional preparation migrations also
  passes (PASS1.926s). Owned PostgreSQL container/volume/credential, supplier
  workspace and Go caches are removed; user services/images/data are preserved.
  Selected graph/HTTP, clean-consumer test placement,
  final affected verification and one complete-slice review/commit remain next.
  Source stays uncommitted at1928f8e. No unit promotion or publication.
  Older checkpoints below are historical, not the current failure state.
- Latest positive workflow (23:13 UTC): actual password signin → OTP-only child
  → native encrypted delivery → captured code → public verification → committed
  session and CSRF now passes real PostgreSQL/race (PASS64.403s). Native OTP
  reservation loads/CAS now use the authenticated challenge subject, not the
  anonymous command's empty UserID; release keeps the consumed MFA primary in
  its separate read-only arm. Original claim/attempt/generation/scope guards
  remain. Public current-session reads, denial/rollback/replay/recovery, HTTP,
  final probe removal, affected regression and coherent-slice review/commit
  remain; this is a positive-path checkpoint, not a delivered slice or qualified
  unit. Source uncommitted at1928f8e. Current bounded cluster/workspace is live.
  Counts/full scope unchanged.
- Current partial work (23:02 UTC): the public secondary-OTP completion service,
  closed OTP session record schemas/forward migration and live OTP factor reader
  are written. Real delivered-code login still FAILS: after correcting an OTP
  counter-scope exclusion and duplicate reservation of an already prepared
  charge, native counted admission, final attachment and domain Begin succeed,
  but ApplyFeature fails and the result remains Unknown with no token. The
  pending journal retains original preparation/output/final plan and one charge.
  No usable OTP login/session or additional qualified unit is claimed. Next:
  locate the domain-apply failure, prove actual session issuance/public reads,
  denial/rollback/replay/recovery and selected graph/HTTP. Temporary stage probes
  and public error-stage wrappers were removed. Final focused coordinator,
  reference, MFA-PG and session-PG race regressions pass; full affected regression,
  coherent-slice review/commit and release qualification remain. Source remains
  uncommitted at1928f8e. Owned test cluster/volume/credential/workspace/caches
  were removed; no publication. Counts and full scope unchanged.
- Latest partial work (22:15 UTC): the existing command journal now retains
  encrypted native OTP attempt custody and original counter time/key metadata
  before session generation. Real PostgreSQL/race proves restoration after
  failed generation and restart/key rotation, conflict on replacement metadata,
  unchanged pending history, and no raw network persistence (PASS1.475s).
  Original opaque intent reconstructs the remaining mapped fields; native OTP
  and counter validation still apply. OTP completion risk preparation preserves
  code/browser custody and existing limits; its selected forward migration
  preserves populated windows/charges/snapshots while keeping browser-flow
  scope exclusive to sending (real PG/race PASS1.779s). Final affected default
  races pass. These are transaction/counter prerequisites, NOT actual public
  OTP verification or session issuance. Next: consume these boundaries in the
  public password/OTP verification service, selected session storage/live
  factor validation and atomic-session acceptance, then graph/HTTP. Source
  remains uncommitted at1928f8e; coherent-slice review/commit and release gates
  remain. No unit promotion. Both bounded test clusters/volumes/credentials,
  supplier workspaces and Go caches removed; user resources preserved.
- Latest partial work (21:32 UTC): native continuation now attributes its exact
  original committed OTP send/factor; extended actual password→OTP delivery
  race journey passes (PASS64.171s). Native consuming attempt restoration,
  original-primary OTP session metadata/issuer, closed mapping and a same-Work
  session effect are written with focused denial/replay/unknown/cancellation
  checks; affected default races pass. Reference original-source decryption and
  native recorded completion/session-effect positive paths still lack public
  verification/real SQL acceptance. This is NOT a usable OTP verification or
  authenticated-session workflow. Next: complete public verify, original
  attempt/risk custody, selected session storage/live validation and real atomic
  rollback/contention/unknown acceptance, then graph/HTTP. Source remains
  uncommitted at1928f8e; final full-slice review and clean-consumer qualification
  remain, with no unit promotion. Owned cluster/volume/credential/workspace/
  caches removed; user services/images/data preserved. Counts/full scope unchanged.
- Latest partial workflow (21:01 UTC): standalone public password → actual
  OTP-only child → confirmed factor/recipient → encrypted queue → delivery
  worker now passes real PostgreSQL/race (PASS64.263s). Enrollment/login share
  native resend limits; cooldown denial, unchanged historical denial, foreign
  browser/factor/channel/callback, injected rollback, restart/replay and proof/
  cookie-free recovery are covered. Sending consumes neither child nor shared
  completion attempts, changes no factor/primary and issues no session.
  Original remember policy/freshness budget are preserved; shorter OTP TTL
  caps the original child lifetime without restarting issuance. Affected
  reference/OTP-PG/password-PG/MFA-PG default race checks pass. Native OTP
  verification → atomic session issuance and opt-in graph/HTTP remain next.
  Source stays uncommitted at1928f8e; final coherent-slice review and release/
  clean-consumer qualification remain. No unit promotion or publication.
  Owned sender test cluster/volume, fixture credential, suppliers/workspace and
  caches removed; user services/images/data preserved. Counts/full scope unchanged.
- Latest partial work (20:16 UTC): native anonymous MFA OTP-send counters now
  enforce original browser/account budgets, preserve actual cookie custody and
  empty journal UserID, and obtain account/browser scope only from the exact
  registered MFA owner. Real PostgreSQL/race PASS1.724s proves six same-browser
  sends share a budget (five effects, sixth denied), replay does not debit again,
  forged subject leaves no charge and populated state survives the forward
  upgrade. This uses an explicit feature/authentication ownership fixture, not
  a public password-to-OTP-send journey. OTP's code/bearer-free input codec and
  original generation/delivery now restore after restart without new recipient,
  configuration or deadlines. Affected kernel/OTP-PG/risk-core/native-risk races
  pass. Public reference sender, actual protected primary/factor fences,
  completion/session and HTTP remain next; no new unit qualifies. All changes
  remain uncommitted at1928f8e and full-slice review is outstanding. Owned cluster,
  volume, synthetic password, scratch/workspace/suppliers/caches removed.
  Consolidated status19:38UTC; next20:38UTC while active.
- Latest partial work (19:42 UTC): risk core now prepares and validates the
  closed anonymous MFA OTP-send profile with the actual account and original
  browser-flow budget, not a fabricated session or a fresh budget per child.
  Existing protected limits, keyed storage and original evidence expiry are
  reused. Substitution/expiry/canonical-network cases and the existing core
  regression pass under race testing (PASS1.364s). Native adapter scope/storage
  and public sender wiring remain unfinished; this is a prerequisite, not a
  working login journey. All source is uncommitted and final slice review is
  outstanding. Counts unchanged. Owned test scratch/caches removed; no database
  created. Consolidated status19:38UTC; next20:38UTC while active.
- Latest partial work (19:23 UTC): native OTP login generation/restoration now
  uses the actual consumed-primary browser arm and original MFA child/recipient
  scope, with no authenticated-session dependency. Another parent, browser,
  generation, recipient, authority, policy or deadline cannot adopt retained
  output. The existing code generator, encrypted delivery and send-family owners
  are reused; the new profile remains distinct from ordinary OTP signin and
  authenticated enrollment. Complete policy/key/store/template configuration
  is bound; a missing policy-identity binding was reproduced and fixed.
  The new native OTP factor reader resolves confirmed owned factor/verified
  recipient metadata only under the actual live OTP-enabled child; selected
  real-PG password→child→factor-source race PASS2.094s, including foreign factor
  denial. Native OTP-PG/MFA-PG/session-PG races pass and bounded hostile login-
  record fuzz PASS11.322s. These are component/source boundaries, not a usable
  public login send/session journey. Public sending/queued persistence,
  completion/session, HTTP/reference and full-slice review/commit remain. Source
  stays uncommitted at1928f8e; no new unit qualifies. Last consolidated status
  19:13UTC; next20:13UTC while active. Changed shared-source native enrollment
  regression remains passing (seven real-PG race cases, PASS20.655s). Owned
  bounded test container, volume, caches/suppliers/workspace removed; unrelated
  processes, images and data preserved.
- Latest partial workflow (18:46 UTC): actual public password login now issues
  an explicitly OTP-only signed continuation for confirmed OTP authority and
  its exact current verified recipient. Unrelated TOTP, missing OTP storage and
  stale/unverified recipients cannot produce that child or consume its browser.
  Existing TOTP default, hash upgrade, rollback/unknown and known-empty SFA
  remain covered; selected real-PG race PASS2.090s. Native OTP consuming custody
  now retains the original consumed primary read-only, never re-reserves,
  releases or finalizes it. Actual cookie substitution/expiry denies.
  Coordinator joined-budget real-PG race PASS1.505s proves both provisional
  debits/audit atomicity, rollback and exactly-once concurrent replay using
  explicit SQL ownership fixtures, not a full native factor-verification journey.
  Shared native MFA helpers reuse original continuation/CAS. Final affected
  kernel/MFA-PG/OTP-PG/password-PG/session-PG races pass. Existing seven native
  enrollment-confirmation race cases remain passing after these changes
  (PASS21.399s). Next: actual OTP login send/delivery/completion participant,
  original-password session effects and selected HTTP/reference. No usable OTP
  login/session or new completed unit is claimed; source remains uncommitted
  at1928f8e. Full-slice independent review/commit remains. Consolidated status
  18:13UTC; next19:13UTC while active. Owned test container, volume, caches,
  supplier archives and scratch removed; unrelated processes/data preserved.
- Latest partial workflow (18:13 UTC): OTP enrollment/delivery/confirmation
  now also passes competing-command and lost-COMMIT-before/after-durability
  acceptance with real PostgreSQL. Exactly one contender activates the factor;
  code-free recovery preserves commit/rollback and replay never redebits.
  A genuine stranded-first-attachment conflict was reproduced and fixed through
  the existing journal, without weakening uncertain outcomes. Seven selected
  real-adapter race cases PASS21.057s; affected kernel/OTP-PG/password-PG default
  races pass. Current source is uncommitted at1928f8e. OTP login/session and
  HTTP remain disconnected; next is explicit password OTP method selection,
  consumed-primary custody and shared-budget atomic completion. No new unit
  qualifies. Owned disposable test resources remain live for this bounded run.
  Last consolidated status18:13UTC; next19:13UTC while active.
- Latest partial workflow (17:40 UTC): native pending OTP enrollment → keyed
  send → actual encrypted delivery worker → code-confirmed factor/recovery
  activation now passes real PostgreSQL/race testing. Confirmation restores the
  actual original send and original enable/session/CSRF; it retains one native
  consuming attempt in the existing preparation journal. Wrong-code denial and
  exact retry debit once. Correct-code activation atomically finalizes OTP,
  activates factor/recovery authority and persists audit/lifecycle, with no
  login session. Restart/replay and code/proof-free recovery pass; changed code,
  foreign tenant/factor/session/CSRF and unsupported trusted-device requests deny.
  Activation rollback leaves factor/recovery inactive and releases, not finalizes,
  OTP; replay cannot redebit or reactivate. Post-preparation recipient revocation
  and account blocking deny before OTP comparison/debit. Final four real-adapter
  race journeys PASS12.245s. Affected default race suites pass; existing TOTP
  URI/confirmation/ambiguity journey PASS6.460s was captured before the final
  OTP-only constructor-key/cancellation hardening. No complete-unit qualification.
  Source remains uncommitted at1928f8e. Contention/ambiguous OTP-confirmation
  acceptance, explicit password-primary OTP login/shared budget/session, selected
  HTTP/reference, catalogs and full-slice review/commit remain. Last consolidated
  status16:50UTC; next17:50UTC while active. Owned PostgreSQL container/volume,
  workspace, supplier archives and caches removed; existing images/data preserved.
- Earlier partial workflow (16:30 UTC): native pending OTP enrollment → keyed MFA send →
  encrypted durable queue → actual delivery-worker callback now passes real
  PostgreSQL/race testing. The send reuses original enable/session/CSRF, protected
  risk and the existing coordinator; OTP retains generation/digest/shared abuse
  and delivery custody. Rollback leaves no challenge/message/factor-version
  advance. Wrong factor/channel/account/session/CSRF deny before generation;
  restart/replay/proof-free recovery preserve the original command/challenge and
  do not generate or deliver again. Changed same-key intent conflicts. Actual
  queue acceptance is not confirmation or an authenticated session.
  Observed UTC-representation and borrowed-synchronizer-reveal failures were
  fixed with regressions, without renewing original time or weakening CSRF.
  Configuration-substitution regression now binds actual OTP/session stores,
  retained keys, freshness and policy. Final selected native OTP plus existing
  TOTP URI/confirmation/ambiguity race journey PASS6.649s; MFA/MFA-PG/OTP-PG/
  password-PG default race suites pass. Source remains uncommitted at1928f8e;
  full-slice review, confirmation, OTP MFA login/session and HTTP/reference remain.
  No unit promotion. Owned PostgreSQL container/volume, caches, scratch and
  supplier archives removed; existing images/data preserved. Last consolidated
  status15:52UTC; next16:52UTC while active.
- Delivery: local feature branches/worktrees only; no platform publication or deployment.
- Runtime: `feature/identity-platform` in the isolated go-identity worktree;
  latest committed checkpoint `1928f8ef5118cdd6980846551b5390a4146bae5b`;
  runtime was clean at that reviewed checkpoint; current secondary-OTP slice
  has uncommitted source/test changes and is not yet delivered.
- Current slice: explicit secondary OTP, not a TOTP/security-key fallback.
  Bounded/redacted public MFA send/verify inputs and OTP's distinct MFA custody
  arm are implemented. Actual consuming-attempt preparation, sealed restoration
  and scope-substitution denial now pass a focused regression that first failed
  on the original unsupported MFA command. Attempt and digest-only issuance
  records preserve original consumed-primary custody in new closed profiles,
  without exposing an ordinary pending pre-auth flow or full session. Existing
  profiles remain supported. MFA/OTP/OTP-PG default race suites pass locally.
  Native pending enrollment/send/confirmation now have real public PostgreSQL
  evidence; no OTP login/session or HTTP/reference journey has run yet. Explicit login policy,
  authoritative consumed-primary admission, shared-budget atomic completion,
  delivery and HTTP/reference wiring remain. No unit promotion or new usable
  workflow claim. Independent final review/commit follow the coherent slice.
- Current OTP storage/enrollment prerequisites advanced: forward shared factor
  references preserve populated TOTP/recovery rows and original TOTP inserts.
  OTP records distinguish pending/confirmed/revoked state, bind owned recipient
  IDs/versions and erase recipient references on revocation. Real PostgreSQL
  storage checks pass. A newly added deletion case exposed orphaned legacy
  recovery state; method-bound and generated per-method foreign references now
  reject deletion for both TOTP and OTP, with an observed red/green regression.
  Core enable requires explicit OTP identifier/channel selection. Authenticated
  pending-enrollment send/confirmation and partial-login child inputs are
  disjoint; bounded code-free original preparation metadata is implemented.
  Core/default adapter race checks pass; these are NOT native service/journey
  completion. All changes remain uncommitted, with full-slice review outstanding.
  Local test cluster, volume, supplier snapshots and disposable caches removed.
  Next: actual delivery/confirmation/session, then selected HTTP/reference.
  Last full status14:30UTC; next15:30UTC while active.
- Native OTP pending enrollment is now connected through `OTPEnrollmentService`.
  It reuses the existing primary-proof, request-key, protected-risk, encrypted
  preparation and transaction owners, without generating a TOTP secret. Real
  PostgreSQL public-path tests prove verified channel/owner selection, authority
  changes after preparation, rollback without primary consumption, atomic pending
  factor and inactive recovery-code creation, fresh reveal-once output, original
  retry and proof/code-free restart recovery. A counter-namespace substitution
  regression failed and is fixed without changing historical TOTP profiles.
  Final selected native OTP plus existing TOTP URI/confirmation/ambiguity journey
  race run PASS6.108s; affected MFA-PG default race PASS1.283s. Pending enrollment
  is not code confirmation or a usable OTP login/session. No OTP HTTP/reference,
  delivery or full-slice final review/commit yet. Source remains uncommitted at
  base1928f8e; counts and full67/19/10 scope are unchanged. All task-owned test
  cluster/volume, supplier archives, scratch and caches are removed.
- Previously delivered slice: authenticated safe MFA factor listing and remaining recovery-code
  counts. Core/native metadata reader and guarded public reference operation
  implemented. Four real-adapter race journeys prove known-empty then enrolled
  TOTP/count10, foreign-tenant denial, corrupt-set and blocked-account denial,
  actual code consumption count10→9, and post-read authority-loss withholding.
  HTTP handler and independently selected feature/closed OpenAPI wire now pass
  focused and whole-module race checks; bounded hostile-envelope fuzz passed.
  Reference HTTP is now connected: real journey PASS17.941s and standalone
  without HTTP-session routes PASS16.043s. Selected-dependency/shutdown tests,
  affected declarations, wire catalog and adoption docs are reconciled. Twenty-
  nine runtime paths received independent GPT-6.1 Sol medium final-review
  APPROVE with no findings. Reviewed base b782f99 and snapshot digest
  `985225784989c9aa74c91e2a81a5a2a0af76f874abefda1a65439679e02c8ee3`;
  locally committed at 9e124de. No full-module
  qualification claim.
- Previously delivered slice: factor label management. New core request/action bounds,
  exact factor-update/disable step-up purpose admission and proof projection
  protections are implemented. Native rename is connected through public reference
  composition and its real-adapter race journey passes: rollback, fresh commit,
  source-session revocation, one-time durable effect and proof-free recovery/replay.
  Opt-in management HTTP/reference and the HTTP TOTP proof producer are now
  connected. The real HTTP step-up → rename journey passes with PostgreSQL and
  TLS Valkey, including rollback, commit, session invalidation and native
  recovery/replay. Affected HTTP, HTTP-MFA and reference race suites pass;
  bounded factor-update and step-up hostile-envelope fuzz runs pass. Null-purpose
  downgrade and malformed-command observation regressions are covered. The
  native and HTTP competing-command cases each produce one winner. Native and
  HTTP lost-acknowledgement cases preserve uncertainty and recover the committed
  original outcome without another mutation or metadata. Real denial tests
  cover second-live-session substitution, foreign tenant, stale factor version,
  altered proof time, changed factor authority and blocked account with protected
  state unchanged. HTTP empty-label clearing passes against a populated label.
  Affected typed request/result/semantics, HTTP catalog and adoption docs are
  reconciled; no complete catalog-generator qualification is claimed. Final
  affected HTTP/HTTP-MFA/reference races pass. Independent GPT-6.1 Sol medium
  final review APPROVE after its sole README caveat was corrected; reviewed
  61-path snapshot `84c8f4ed28b01a7ac2111786f868663d4d7c05568831ab79217ba518b81e8a6c`.
  Locally committed at c282521; runtime worktree was clean at that checkpoint,
  nothing pushed.
- Latest delivered slice: protected sole-TOTP removal with explicitly selected
  emailed-password-reset recovery, locally committed at cb481b6. Native and
  opt-in HTTP journeys prove removal → actual reset request/delivery/completion
  → old-password denial → replacement-password login. Removal erases factor
  secrets, retires dependent recovery codes and revokes two genuine sessions.
  Final eight real-adapter race journeys PASS139.010s cover recovery authority
  loss, lock serialization, rollback, valid competing removals, lost-acknowledgement
  recovery and original-outcome replay. An ambiguous password-reference defect
  was observed red and fixed; disabled duplicate references also deny safely.
  Five affected module race suites and bounded hostile-envelope fuzz pass.
  Independent regression, Go compatibility and complete-final-diff reports
  APPROVE the bounded slice; delivery audit verifies the local committed
  checkpoint and exact reviewed 38-path snapshot
  `007dc9219b7ebfdf3d3a901fee9d8f10888ed41dbab6e97a628429804e806b79`.
  The auditor corrected its initial report-provenance wording; its inspected
  native reports do not certify full release qualification. All task-owned
  test containers, PostgreSQL volume, scratch and caches are removed.
  Nothing pushed or deployed; no full unit promotion. Last full status06:53UTC;
  next07:53UTC while active.
- Latest slice: explicit keyed MFA disable, native and opt-in HTTP/reference,
  locally committed at ef6011f after complete-final-diff review APPROVE.
  One/all targets retain distinct intent and derive optimistic generation from
  original signed disable-purpose proof, not private caller SQL. Real-adapter
  native/HTTP success, two-session revocation, genuine pending-sibling erasure,
  protected rollback, contention, lost-acknowledgement recovery, changed-authority
  denials and actual emailed-reset/new-password login pass. Seven affected module
  races and bounded hostile-envelope fuzz pass; source/rendered typed fields and
  selected HTTP catalog are reconciled. Go compatibility review APPROVE for the
  local additive contract scope; independent regression review APPROVE. The
  original final-review finding was reproduced red and fixed: one-target now
  preserves and reports genuine unexpired pending siblings from the protected
  transaction; all-target erases them and returns empty. Native/HTTP pending
  and all-target integration PASS58.237s; remaining eight disable scenarios
  PASS134.382s; changed modules/direct consumers and bounded parser fuzz pass.
  Complete-final-diff re-review APPROVE on55-path snapshot
  `92893b2887ce07653142b673f80b112f9fe5bb1bdfebede1525d72e1a574d00b`,
  independently reproduced from the actual committed head. Native delivery
  audit APPROVE independently verifies committed content, clean runtime and
  cleanup, and directly retrieves native QA/mapping/Go/regression/final reports.
  Their differing reviewed boundaries remain recorded, not re-certified as one
  whole-platform release. Scope is sole active TOTP plus bounded pending
  siblings and selected email recovery, not general multi-factor/device/admin.
  All task-owned test containers, anonymous PostgreSQL volume, scratch, suppliers
  and caches have been removed; existing images/data preserved. Nothing pushed.
  Last consolidated status08:02UTC; next09:02UTC while active.
- Latest committed slice: public TOTP replacement, locally committed at
  `1928f8ef5118cdd6980846551b5390a4146bae5b`. Native and opt-in HTTP consumers
  can initiate a distinct pending successor, reveal its URI once, confirm the
  new code, retire old factor/codes, revoke subject sessions and sign in with
  the successor. Old-factor and old-code authentication remain usable while
  pending; successor codes cannot authenticate before confirmation.
- Real-adapter checks cover delayed genuine authority refresh without renewing
  original ceilings; native/HTTP begin and confirmation rollback, competing
  commands and lost-acknowledgement recovery; current authority/expiry/outage
  denials and original same-user-session substitution. Actual old-factor login
  advances replay state without cancelling the pending ceremony, then both
  sessions are revoked on confirmation. Recovery-code activation/retirement
  is proven through actual password/recovery HTTP.
- Independent review exposed and corrected discarded native plaintext custody,
  MFA-origin continuation schema rejection and unfinishable penultimate-version
  initiation. Each has actual red/green evidence. Both genuine password/TOTP and
  password/recovery source-session ceremonies now pass. The highest finishable
  version also completes through HTTP without overflow; the next version denies
  before proof consumption or successor creation. Exact schema/current-factor
  fences and original one-step rename/removal compatibility are preserved.
- Final affected six-module/direct-consumer race results pass, with unchanged
  HTTP evidence reused; bounded parser fuzz and full changed-file secret scan
  pass. Original failed multi-case setup runs remain recorded as failed, with
  isolated passing corrections, not relabelled as green batches. Native
  regression, Go compatibility and complete-final-diff reviews APPROVE the
  final45-path snapshot
  `3a80058353e1056a01105028d7d2caaf8d8dd4ea1438ded96860c9fa244a6fa5`.
  Native delivery audit APPROVE directly verifies reports, actual terminal
  outcomes, clean committed content and cleanup. No independent execution is
  attributed to source-only reviewers.
- All owned local test services, PostgreSQL volume, supplier snapshots, scratch
  and caches are removed; existing images/data are preserved. Nothing pushed
  or deployed. Supporting replacement transaction/type rows are aligned;
  aggregate catalog generation still fails on recorded pre-existing issues.
  Full module/release qualification is not claimed. Next: remaining MFA
  capabilities and the rest of the unchanged full platform scope.
  Last consolidated status12:31UTC; next13:31UTC while active.
- Remaining: other factor methods/devices, general multi-factor replacement,
  administrative recovery and full MFA/release qualification remain required.
- No external blocker. Full module/release qualification remains outstanding.

The flat unit count does not imply no recent code delivery. Registration, reset,
password changes, magic-link login and MFA workflows have advanced within large
unfinished units. Conversely, these slices do not establish complete package parity.
The inventory's coarse statuses underreport some partial source implementation;
do not interpret a historical `proposed` row as proof of zero code.

## Consumer-visible workflows

Historical committed workflows below retain their original delivery status;
they are not automatically whole roadmap rows or fully qualified packages.

| Workflow | Current bounded state |
| --- | --- |
| Password registration → emailed verification → confirmation | Locally committed; focused integration evidence |
| Password signin → session cookie → session read/logout | Locally committed; focused integration evidence |
| Reset request → durable delivery → completion → password/session invalidation | Locally committed; focused integration evidence |
| Fresh password verification → protected password change | Locally committed; focused integration evidence |
| Magic-link signup/login, including MFA continuation | Reviewed coherent slice locally committed |
| TOTP enrollment → provisioning → confirmation → password-first TOTP login | Reviewed coherent slice locally committed |
| Password-first recovery-code login → atomic code consumption → session/logout | Reviewed coherent slice locally committed at d082ef6 |
| Fresh TOTP step-up → replace codes → old-code denial → session revocation → new-code login | Independently approved and locally committed at b782f99; real-adapter denial checks passed |
| Safe MFA factor inventory and remaining recovery-code counts | Native and actual reference HTTP workflows tested, including standalone selection; independent review APPROVE, locally committed at 9e124de |
| Fresh TOTP proof → factor rename/clear label → session invalidation | Native/HTTP journeys, rollback, one-winner contention, lost-ack recovery and authority/substitution denials pass; independent review APPROVE, locally committed at c282521 |
| Fresh TOTP proof → remove sole factor → selected email reset → replacement-password login | Native/HTTP real-adapter recovery, denial, rollback, contention and lost-ack journeys pass; independent reviews APPROVE, locally committed at cb481b6; full module qualification remains |
| Fresh disable-purpose TOTP proof → explicit one/all disable → session revocation → email-reset recovery | Native/HTTP journeys pass, including accurate pending-sibling conservation; final re-review/local delivery audit APPROVE, locally committed at ef6011f; full module qualification outstanding |
| Fresh TOTP proof → pending replacement → reveal-once URI → confirmation → retirement/revocation → new-factor login | Native and selected HTTP journeys pass, including genuine MFA source sessions, old-factor/code usability, authority denials, rollback/competition/recovery and version bounds; independent reviews/audit APPROVE, locally committed at 1928f8e; full module qualification remains |

These are backend public APIs and opt-in HTTP handlers, not UI. They do not
certify every requirement of their containing units, hosted CI or production.

## Fixed consumer milestones

The canonical goal has 39 unweighted milestone rows. None is promoted merely
from a helper, a historical slice count or a compilation result. Existing proven
profiles remain progress and are not restarted; full-row delivery requires its
assigned profile/operation scope and review/commit closure. Initial row status
below is deliberately conservative, not a new whole-platform percentage.

| ID | Consumer outcome and principal owning units | Current status |
| --- | --- | --- |
| B1 | Runnable registration/verification/signin/session/logout baseline: identity, identity/postgres, identity/password, identity/password/postgres, identity/session, identity/session/postgres, identity/delivery, identity/delivery/postgres, identity/http, identity/reference; required primitive contracts Existing committed account/login work; runnable baseline closure pending |
| B2 | Reset and required session invalidation using the baseline owners plus identity/risk/hibp Committed selected reset workflow; baseline consumer closure pending |
| B3 | Password change and fresh reauthentication through baseline owners Committed selected verify/change workflow; baseline closure pending |
| M04 | Email add/change/verify/list/get/remove with session/lifecycle policy: identity/email Not delivered; reconcile existing proof on entry |
| M05 | Username availability/signin/rename: identity/username, stable identifier contracts Not delivered; reconcile existing proof on entry |
| M06 | Delivered magic-link signin and session: identity/magiclink Committed selected magic-link workflow; full profile qualification pending |
| M07 | Purpose-bound email OTP request/delivery/completion: identity/otp, identity/otp/postgres Native/selected OTP workflow exists; full operation/profile closure pending |
| M08 | Verified phone enrollment/change, OTP/password signin and reset: identity/phone, OTP/password owners Not delivered; reconcile existing proof on entry |
| M09 | Anonymous session, self-deletion and collision-safe permanent upgrade: identity/anonymous, identity/anonymous/postgres Not delivered; reconcile existing proof on entry |
| M10 | Confirmed TOTP enrollment and TOTP signin: identity/mfa, identity/mfa/postgres Committed selected TOTP enrollment/login; full profile closure pending |
| M11 | Confirmed secondary-OTP enrollment, delivered-code signin and session: MFA/OTP/password/session owners Email OTP HTTP journey proven; final regression/review/commit pending |
| M12 | Factor listing/update/removal/disable/replacement with fresh proof: MFA owners Committed TOTP management/removal/disable/replacement; full profile closure pending |
| M13 | Recovery-code reveal/use/regeneration and independently approved administrative recovery: MFA owners Committed code login/regeneration; administrative recovery/profile closure pending |
| M14 | Trusted-device issuance/use/revocation and step-up: MFA/session owners Not delivered; reconcile existing proof on entry |
| M15 | WebAuthn register/assert ceremony and durable credential validation: webauthn, webauthn/postgres Not delivered; reconcile existing proof on entry |
| M16 | Passkey-first registration, discoverable signin and credential management: passkey, passkey/postgres Not delivered; reconcile existing proof on entry |
| M17 | Social OAuth authorization-code/S256 callback to session: identity/oauth, identity/oauth/postgres, identity/oauth/providers Not delivered; reconcile existing proof on entry |
| M18 | Social account linking/unlinking and stored provider-token refresh: OAuth owners Not delivered; reconcile existing proof on entry |
| M19 | Verified One Tap to identity/session: identity/oauth/onetap Not delivered; reconcile existing proof on entry |
| M20 | Bounded OAuth proxy callback to the intended application: identity/oauth/proxy Not delivered; reconcile existing proof on entry |
| M21 | User/organization API-key issue/import/verify/rotate/revoke: identity/apikey, identity/apikey/postgres, identity/apikey/valkey Not delivered; reconcile existing proof on entry |
| M22 | Organization create/manage/join/invite/remove and authority lifecycle: organization, organization/postgres Not delivered; reconcile existing proof on entry |
| M23 | Teams and dynamic roles/permissions enforcement: organization, authorization contracts Not delivered; reconcile existing proof on entry |
| M24 | Bans/session administration and bounded audited impersonation: identity, session, identity/impersonation, identity/impersonation/postgres Not delivered; reconcile existing proof on entry |
| M25 | Verified domain/provider registration and federation policy: sso, sso/domain-verification, sso/postgres Not delivered; reconcile existing proof on entry |
| M26 | Enterprise OIDC/OAuth federation, mapping/JIT/linking and session: sso/oidc, sso/oauth2 Not delivered; reconcile existing proof on entry |
| M27 | SAML federation including replay/lifecycle policy: sso/saml Not delivered; reconcile existing proof on entry |
| M28 | Authorized SCIM discovery/User CRUD/filter/PATCH/pagination: scim, scim/postgres Not delivered; reconcile existing proof on entry |
| M29 | SCIM Groups, organization reconciliation and deprovisioning effects: scim/organization and organization owners Not delivered; reconcile existing proof on entry |
| M30 | OAuth client/consent, authorization-code/S256 grants and refresh rotation: oauth-server, oauth-server/postgres Not delivered; reconcile existing proof on entry |
| M31 | OIDC discovery/JWKS/ID-token/userinfo consumers: oauth-server/oidc Not delivered; reconcile existing proof on entry |
| M32 | Workload client-credentials issuance, introspection and revocation: OAuth server owners Not delivered; reconcile existing proof on entry |
| M33 | Restricted token exchange with lineage/non-expanding authority and bounded cache validity: OAuth server, authentication, API-key and organization owners; SI-01 through SI-10 Not delivered; reconcile existing proof on entry |
| M34 | Device authorization to issuance/expiry/denial: oauth-server/device, optional native PostgreSQL composition Not delivered; reconcile existing proof on entry |
| M35 | Action-bound risk/velocity/admission and every selected CAPTCHA provider: identity/risk, identity/risk/postgres, identity/risk/valkey, identity/risk/captcha, identity/risk/captcha/recaptcha, identity/risk/captcha/turnstile, identity/risk/captcha/hcaptcha, identity/risk/captcha/captchafox; HIBP remains required when setting passwords Risk/native stores and HIBP implemented in part; CAPTCHA/full profile closure pending |
| M36 | Locale selection and localized public messages/delivery templates: identity/i18n Not delivered; reconcile existing proof on entry |
| M37 | Privacy export/deletion/anonymization/legal retention and audit investigation: identity and applicable lifecycle owners Not delivered; reconcile existing proof on entry |
| M38 | Session refresh/rotation, account switching/transfer/enrichment, devices and consistent remember-me, including selected cache projection: identity/session, identity/session/postgres, identity/session/valkey Selected session issuance/read/revocation work; remaining session lifecycle pending |
| M39 | Complete opt-in reference composition, OpenAPI/readiness, typed extensions/operations and reusable consumer contract suites: identity/reference, identity/http, identity/identitytest; all remaining cross-cutting acceptance Incremental selected composition; complete reference/typed acceptance pending |

## All 67 units

States: **Verified** = full recorded unit qualification; **Integrated** = accepted
primitive implementation with qualification outstanding; **Partial** = usable or
consumed behavior exists but full scope/qualification is unfinished; **Remaining**
= no completed unit acceptance, not necessarily no source code.

### Primitive foundations — 6

| # | Unit | Required outcome | State |
| --- | --- | --- | --- |
| 1 | primitive/authentication-identity-contracts | Credential/principal contracts, opaque proofs, assurance, bounded validity | Integrated |
| 2 | primitive/authorization-identity-contracts | Existing policy engine and identity/organization contracts | Integrated |
| 3 | primitive/capability-identity-contracts | Purpose-bound proofs and compatible transactional one-use | Verified |
| 4 | primitive/capability-postgres-identity-contracts | Durable reservation, consumption and mutation recovery | Partial |
| 5 | primitive/identifier-identity-contracts | Stable normalization with fixed Unicode tables | Integrated |
| 6 | primitive/password-secret-contracts | Owned secrets and compatible parameter validation | Verified |

### Identity, sessions and delivery — 9

| # | Unit | Required outcome | State |
| --- | --- | --- | --- |
| 7 | identity | Users/accounts/identifiers/credentials, lifecycle, hooks, transactions | Partial |
| 8 | identity/postgres | Durable kernel, migrations and atomic lifecycle | Partial |
| 9 | identity/session | Issuance/freshness/rotation/revocation, devices, switching/transfer | Partial |
| 10 | identity/session/postgres | Durable authoritative sessions and lifecycle | Partial |
| 11 | identity/session/valkey | Bounded safe session projections/cache | Partial |
| 12 | identity/delivery | Email/SMS intents, templates, senders and retry policy | Partial |
| 13 | identity/delivery/postgres | Durable encrypted delivery custody and recovery | Partial |
| 14 | identity/i18n | Locale selection and localized messages/templates | Remaining |
| 15 | identity/http | Typed secure independently opt-in transport | Partial |

### Risk and account authentication — 17

| # | Unit | Required outcome | State |
| --- | --- | --- | --- |
| 16 | identity/risk | Action-bound abuse/rate/velocity and one-use evidence | Partial |
| 17 | identity/risk/postgres | Durable counters/reservations and atomic enforcement | Partial |
| 18 | identity/risk/valkey | Distributed ephemeral risk controls | Partial |
| 19 | identity/risk/captcha | Bounded provider-neutral verification and failure policy | Remaining |
| 20 | identity/risk/captcha/recaptcha | Google verification adapter | Remaining |
| 21 | identity/risk/captcha/turnstile | Cloudflare verification adapter | Remaining |
| 22 | identity/risk/captcha/hcaptcha | hCaptcha verification adapter | Remaining |
| 23 | identity/risk/captcha/captchafox | CaptchaFox verification adapter | Remaining |
| 24 | identity/risk/hibp | Privacy-preserving breached-password screening | Partial |
| 25 | identity/password | Registration/login/verification/reset/change/upgrades/reauthentication | Partial |
| 26 | identity/password/postgres | Durable password workflows and recovery | Partial |
| 27 | identity/username | Availability/signin and stable rename | Remaining |
| 28 | identity/email | Verification and address add/change/list/get/remove | Partial |
| 29 | identity/magiclink | Replay-safe passwordless signup/login | Partial |
| 30 | identity/otp | Purpose-bound challenges/attempts and atomic use | Partial |
| 31 | identity/otp/postgres | Durable OTP reservation/finalization/recovery | Partial |
| 32 | identity/phone | E.164, OTP/password signin and safe change/reset | Remaining |

### Anonymous identity, MFA and passkeys — 8

| # | Unit | Required outcome | State |
| --- | --- | --- | --- |
| 33 | identity/anonymous | Anonymous sessions/deletion/permanent-account upgrade | Remaining |
| 34 | identity/anonymous/postgres | Durable collision-safe anonymous lifecycle | Remaining |
| 35 | identity/mfa | TOTP/OTP/recovery/trusted devices/step-up/factor recovery | Partial |
| 36 | identity/mfa/postgres | Durable factor authority/replay fences/atomic effects | Partial |
| 37 | webauthn | Registration/assertion protocol and attestation policy | Remaining |
| 38 | webauthn/postgres | Durable credentials and ceremonies | Remaining |
| 39 | passkey | Passkey-first registration/usernameless signin/management | Remaining |
| 40 | passkey/postgres | Durable identity/session orchestration around WebAuthn | Remaining |

### Social login, API keys and impersonation — 10

| # | Unit | Required outcome | State |
| --- | --- | --- | --- |
| 41 | identity/oauth | Social OAuth state/PKCE/callbacks/refresh/linking | Remaining |
| 42 | identity/oauth/postgres | Durable OAuth state and provider-token custody | Remaining |
| 43 | identity/oauth/providers | Supported provider catalog/configuration | Remaining |
| 44 | identity/oauth/onetap | Secure One Tap orchestration | Remaining |
| 45 | identity/oauth/proxy | Bound secure OAuth proxy flows | Remaining |
| 46 | identity/apikey | User/org keys, scopes, rotation/revocation and legacy import | Remaining |
| 47 | identity/apikey/postgres | Durable key authority and digest-only import | Remaining |
| 48 | identity/apikey/valkey | Safe cached verification and bounded revocation | Remaining |
| 49 | identity/impersonation | Authorized bounded auditable owner/actor separation | Remaining |
| 50 | identity/impersonation/postgres | Durable impersonation lifecycle and sessions | Remaining |

### Organizations, federation and provisioning — 11

| # | Unit | Required outcome | State |
| --- | --- | --- | --- |
| 51 | organization | Organizations/memberships/invitations/teams/roles | Remaining |
| 52 | organization/postgres | Durable organization lifecycle and authority | Remaining |
| 53 | sso | Enterprise federation/mappings/JIT provisioning | Remaining |
| 54 | sso/domain-verification | Domain ownership and federation selection | Remaining |
| 55 | sso/oidc | Enterprise OIDC federation | Remaining |
| 56 | sso/oauth2 | Enterprise OAuth federation | Remaining |
| 57 | sso/saml | SAML federation and replay/lifecycle policy | Remaining |
| 58 | sso/postgres | Durable federation connections/mappings | Remaining |
| 59 | scim | Discovery, Users/Groups, CRUD/filter/PATCH | Remaining |
| 60 | scim/postgres | Durable provisioning/reconciliation | Remaining |
| 61 | scim/organization | Directory-to-organization lifecycle integration | Remaining |

### Authorization server and complete composition — 6

| # | Unit | Required outcome | State |
| --- | --- | --- | --- |
| 62 | oauth-server | Owned OAuth: clients/consent/PKCE/grants/restricted exchange | Remaining |
| 63 | oauth-server/oidc | OIDC issuance/signing/discovery/JWKS | Remaining |
| 64 | oauth-server/device | Device authorization | Remaining |
| 65 | oauth-server/postgres | Durable clients/grants/rotation/introspection/revocation | Remaining |
| 66 | identity/reference | Complete executable composition/OpenAPI/migrations/journeys | Partial |
| 67 | identity/identitytest | Reusable public contract/integration suites | Remaining |

## Remaining acceptance beyond package implementation

All nineteen journey categories require complete proof: identity lifecycle;
email/password/username; sessions; passwordless/phone/anonymous; MFA/WebAuthn;
social OAuth; API keys; administration; organizations; enterprise federation;
SCIM; OAuth/OIDC provider; abuse/CAPTCHA/HIBP; localization; developer adoption;
extensions/access policy; typed modules/operations; privacy export; audit investigation.
Working subsets above are not full-category completion.

All ten shared-issuer cases in [SHARED_ISSUER_PROFILE.md](SHARED_ISSUER_PROFILE.md)
remain required, including restricted delegation/root lineage, identity separation,
non-expanding authority, bounded revocation/freshness through caches and outages,
legacy SHA-256 digest import, creator-independent organization keys, canonical IDs,
populated migrations, opt-in composition and distinct standard/private exchange.
No complete shared-issuer milestone is claimed.

Before unit verification: exact production statement coverage, viable-mutant killing
with reviewed exceptions, applicable race/stress/leak/fuzz/interoperability and
benchmark evidence; vulnerability/secret/license/SBOM/provenance/clean-consumer
checks; useful APIs/examples/adoption docs/migrations/changelogs. Final affected
and aggregate obligations remain. Narrow slice tests do not replace these gates.

## Source authority

- Active goal: `/Users/brian/.codex/attachments/20198e15-da75-4bd2-b31f-c9294389a6b6/pasted-text-1.txt`.
- [Inventory](INVENTORY.md), [execution ledger](EXECUTION_LEDGER.md),
  [end-state acceptance](END_STATE.md), [shared issuer](SHARED_ISSUER_PROFILE.md).
- Runtime source: `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`.
- Latest independent review: `/root/replacement_final_review`, base ef6011f,
  final45-path snapshot3a80058353e1056a01105028d7d2caaf8d8dd4ea1438ded96860c9fa244a6fa5;
  APPROVE after genuine MFA-origin and near-exhaustion corrections. Local commit
  1928f8e retains that exact content. `/root/replacement_delivery_audit` directly
  inspected native role reports and terminal outcomes and verified the clean
  committed snapshot plus removal of owned services/caches. No full unit promotion.
- Prior regeneration review: `/root/recovery_regeneration_final_review`, base d082ef6;
  original REQUEST_CHANGES solely for denial evidence. Added second-genuine-live-session
  substitution and monotonic factor-epoch denial plus coherent foreign-tenant graph
  denial, asserting unchanged protected state. Real-adapter race PASS16.771s; final
  reference race PASS7.515s. No production changes; these tests are not retroactive TDD.
  Final re-review APPROVE at 83-path snapshot
  `995cabb0507fbd4c4b4ddd157cc9a9ef1b2a2666d22f687ef8bed8fd2baee2db`. Catalog-dependent
  HTTP/HTTP-MFA/reference race checks passed after the frozen catalog reconciliation;
  no passing aggregate catalog-generation or release qualification claim.
