# Goal: Deliver the Complete Golib Identity Platform Through Finite Consumer Milestones

## Objective

Implement and verify the complete composable Go 1.27 backend identity platform
competitive with Better Auth at pinned baseline
`b8077b74ef9a80a7757220b72834349bd8de05c0`. Deliver supported public libraries,
real durable adapters, opt-in HTTP composition, an executable reference backend,
clean-consumer examples and behavioral acceptance evidence.

Complete all 67 existing units (six primitive extensions and 61 product units),
all nineteen journey categories and ten shared-issuer acceptance cases. A usable
baseline, a milestone or a primitive is not whole-platform completion.
This is a backend library platform, not an administrative UI.

## Execution authority and activation

This is the single canonical implementation prompt. On explicit user activation,
continue the existing implementation checkpoints; do not restart the project.
Until activated, this file is a prepared goal, not authority to resume work.
The bounded repair/checkpoint goal must finish first.

This user-approved execution model replaces coordinator-only execution,
package-at-a-time scheduling, implementation workers, routine separate mapper/
QA/regression/Go/delivery stages, per-helper review/commit cycles and historical
global preflight/assignment/capture-release ceremonies. One persistent primary
agent owns the causal slice across necessary independent modules. Use isolated
GPT-6.1 Sol medium subagents only for a concrete bounded read-only investigation
or independent complete-final-diff review. Never self-attest independent review.

Detailed product goals and catalogs remain binding for behavior and ownership.
Historical execution files do not add procedural gates. Do not weaken security,
discard work, remove final qualification gates, fabricate evidence or violate
higher-priority instructions. A concrete product contradiction must be resolved
against actual user authority, not by making stale validators pass.

## Workspaces, preservation and continuity

Use the existing control worktree:
`/Users/brian/Developer/.worktrees/golib-go-library-tools-identity-platform`,
on branch `main` after the transition checkpoint; its `.ai/identity-platform` owns active
specifications and the compact execution ledger. Use REPOSITORY_TOPOLOGY.md
to resolve logical `pkg/...` labels into standalone repositories/nested modules.
The collection `/Users/brian/Developer/golib` is not a Git repository or Go workspace.
Keep repository-qualified commits; never compare ancestry across repositories.

Preserve existing worktrees, dirty work, assignment history, decisions and valid
evidence. Inspect live state once on resumption; verify repository/branch before
writes and confirm previous writers have stopped. Follow required branch/worktree
skills only when setup is necessary and authorized. Read the selected complete
package goals on first entry and relevant contract rows, not all catalogs before
each edit. After compaction, recover from the existing current-milestone note and
changed source. Do not restart discovery, preflight or previous delivered work.

All current public, nested and test modules target Go 1.27 with at least
`go 1.27.0`. Historical results retain their original toolchain identity.
Work on each affected Golib repository's `main` branch and deliver coherent
reviewed commits by ordinary push to its established origin/main. This is the
user-authorized delivery target, not permission to push unrelated work. Preserve
concurrent changes and verify current remote main before adoption/delivery.
Finish pending work in its existing worktree before integrating its reviewed
checkpoint into main; verify and push that integration before follow-up edits.
Do not create routine feature branches. Publication/tags/releases/deployment,
rebase/reset/amend/force-push, hook bypass and discarding work remain unauthorized.

## Milestone contract: freeze, implement, prove, close

Work on one complete consumer outcome, across its necessary owners, at a time.
Before coding, make one short entry in EXECUTION_LEDGER.md with:
- milestone ID, selected profile and supported public entry point;
- exact observable consumer outcome and explicit excluded variants;
- actual owners, necessary consumed contracts and their committed source pins;
- security invariants, finite success/denial/recovery acceptance scenarios;
- exact affected test/check commands and expected outcomes;
- public-path integration/example proof and selected HTTP/reference surface.

Freeze this entry after one bounded investigation pass. Keep it in the existing
ledger, not a new plan, schema, approval chain or receipt. A constructor, storage
contributor, helper or isolated passing test is not a delivery milestone.

Implement the promised full path: supported public API, actual persistence,
authorization/current authority, security, errors/recovery, required lifecycle/
session effects and selected composition. Consumers may supply configuration,
credentials, templates, bounded policy callbacks and email/SMS senders, not
missing domain stores, orchestration, handlers or migrations.

Only a changed requirement or concrete code/test/review defect in that promised
behavior can reopen scope. Record that reason in the same note. Additional
variants, generalization, unrelated refactoring and polish remain later work.
Unselected/unsupported variants expose no capabilities and fail closed; never
silently substitute a weaker factor or authority.

When frozen acceptance and affected checks pass and one independent complete-
final-diff review has no unresolved material findings, commit coherent explicit
paths with affected changelogs and close the milestone immediately. Do not add
neighboring behavior before committing. Concrete post-review fixes require
affected reruns and complete-final-diff review of the resulting tree, not fresh
planning or repetition of unchanged predecessor reviews.

Do not start another independent milestone until closure unless a concrete
external dependency prevents continuation; record the gap without claiming
delivery. A prerequisite owner change is part of the same consumer outcome.

## First deliverable: adopt the existing operational baseline

Start with the existing implemented account/password/session behavior. Do not
rewrite it or rediscover its security decisions. Close these milestones separately:
- B1: runnable selected reference backend with real PostgreSQL bootstrap,
  registration -> captured verification delivery -> verification -> signin ->
  authenticated session read -> logout. Provide reproducible local run commands
  and a supported clean-consumer example.
- B2: reset request -> captured delivery -> completion -> old-password denial ->
  new-password signin -> required session invalidation, with relevant failure/
  replay/rollback and current recovery-factor authority.
- B3: fresh reauthentication -> password change -> required session effects,
  with wrong/stale-proof denial.

Reuse valid committed workflow evidence and connect missing composition only.
A controlled sender can capture delivery at the declared external sender seam;
fake domain stores or invented authority cannot prove the journey. No UI,
deployment or application integration is required. Demonstrate these baseline
outcomes before expanding into another capability family. Do not combine their
remaining gaps into another indefinitely growing slice.

## Fixed consumer roadmap and dependency order

The following IDs are the fixed unweighted delivery denominator: 39 milestones,
including B1-B3. Each closes a consumer outcome, not an entire release unit.
Detailed operation/profile catalogs define its full eventual capability scope;
freeze one supported profile before coding. Preserve already delivered evidence
and assess its applicability once rather than restarting that milestone.

| ID | Consumer outcome and principal owning units |
| --- | --- |
| B1 | Runnable registration/verification/signin/session/logout baseline: identity, identity/postgres, identity/password, identity/password/postgres, identity/session, identity/session/postgres, identity/delivery, identity/delivery/postgres, identity/http, identity/reference; required primitive contracts |
| B2 | Reset and required session invalidation using the baseline owners plus identity/risk/hibp |
| B3 | Password change and fresh reauthentication through baseline owners |
| M04 | Email add/change/verify/list/get/remove with session/lifecycle policy: identity/email |
| M05 | Username availability/signin/rename: identity/username, stable identifier contracts |
| M06 | Delivered magic-link signin and session: identity/magiclink |
| M07 | Purpose-bound email OTP request/delivery/completion: identity/otp, identity/otp/postgres |
| M08 | Verified phone enrollment/change, OTP/password signin and reset: identity/phone, OTP/password owners |
| M09 | Anonymous session, self-deletion and collision-safe permanent upgrade: identity/anonymous, identity/anonymous/postgres |
| M10 | Confirmed TOTP enrollment and TOTP signin: identity/mfa, identity/mfa/postgres |
| M11 | Confirmed secondary-OTP enrollment, delivered-code signin and session: MFA/OTP/password/session owners |
| M12 | Factor listing/update/removal/disable/replacement with fresh proof: MFA owners |
| M13 | Recovery-code reveal/use/regeneration and independently approved administrative recovery: MFA owners |
| M14 | Trusted-device issuance/use/revocation and step-up: MFA/session owners |
| M15 | WebAuthn register/assert ceremony and durable credential validation: webauthn, webauthn/postgres |
| M16 | Passkey-first registration, discoverable signin and credential management: passkey, passkey/postgres |
| M17 | Social OAuth authorization-code/S256 callback to session: identity/oauth, identity/oauth/postgres, identity/oauth/providers |
| M18 | Social account linking/unlinking and stored provider-token refresh: OAuth owners |
| M19 | Verified One Tap to identity/session: identity/oauth/onetap |
| M20 | Bounded OAuth proxy callback to the intended application: identity/oauth/proxy |
| M21 | User/organization API-key issue/import/verify/rotate/revoke: identity/apikey, identity/apikey/postgres, identity/apikey/valkey |
| M22 | Organization create/manage/join/invite/remove and authority lifecycle: organization, organization/postgres |
| M23 | Teams and dynamic roles/permissions enforcement: organization, authorization contracts |
| M24 | Bans/session administration and bounded audited impersonation: identity, session, identity/impersonation, identity/impersonation/postgres |
| M25 | Verified domain/provider registration and federation policy: sso, sso/domain-verification, sso/postgres |
| M26 | Enterprise OIDC/OAuth federation, mapping/JIT/linking and session: sso/oidc, sso/oauth2 |
| M27 | SAML federation including replay/lifecycle policy: sso/saml |
| M28 | Authorized SCIM discovery/User CRUD/filter/PATCH/pagination: scim, scim/postgres |
| M29 | SCIM Groups, organization reconciliation and deprovisioning effects: scim/organization and organization owners |
| M30 | OAuth client/consent, authorization-code/S256 grants and refresh rotation: oauth-server, oauth-server/postgres |
| M31 | OIDC discovery/JWKS/ID-token/userinfo consumers: oauth-server/oidc |
| M32 | Workload client-credentials issuance, introspection and revocation: OAuth server owners |
| M33 | Restricted token exchange with lineage/non-expanding authority and bounded cache validity: OAuth server, authentication, API-key and organization owners; SI-01 through SI-10 |
| M34 | Device authorization to issuance/expiry/denial: oauth-server/device, optional native PostgreSQL composition |
| M35 | Action-bound risk/velocity/admission and every selected CAPTCHA provider: identity/risk, identity/risk/postgres, identity/risk/valkey, identity/risk/captcha, identity/risk/captcha/recaptcha, identity/risk/captcha/turnstile, identity/risk/captcha/hcaptcha, identity/risk/captcha/captchafox; HIBP remains required when setting passwords |
| M36 | Locale selection and localized public messages/delivery templates: identity/i18n |
| M37 | Privacy export/deletion/anonymization/legal retention and audit investigation: identity and applicable lifecycle owners |
| M38 | Session refresh/rotation, account switching/transfer/enrichment, devices and consistent remember-me, including selected cache projection: identity/session, identity/session/postgres, identity/session/valkey |
| M39 | Complete opt-in reference composition, OpenAPI/readiness, typed extensions/operations and reusable consumer contract suites: identity/reference, identity/http, identity/identitytest; all remaining cross-cutting acceptance |

Roadmap order is a dependency-aware priority, not a reason to implement security
prerequisites late. Required risk, HIBP, CAPTCHA, factor/organization authority,
delivery, primitive or session behavior must exist before a milestone relying
on it is accepted; later roadmap rows retain broader feature completeness.
M15 must precede security-key variants of M10-M14/M16; do not require unrelated
WebAuthn completion to develop a proven TOTP-only or OTP-only profile. M22 and
required authority precede organization-owned variants of M21. Shared-issuer
requirements apply from the first affected API, not only at M33.

If a row contains materially separate supported profiles, freeze successive
complete profile deliveries inside the same milestone; do not change the
denominator or call its full row delivered until its assigned behavior is proven.
The roadmap is not permission to omit catalog operations or arbitrary scope.

## Prerequisites without readiness bureaucracy

A consumed contract unlocks development when implemented, locally committed,
integrated and proven by focused behavior and direct-consumer compatibility,
even if its owning unit is implemented-unverified. Record its owner revision,
consumed behavior, relevant native result and pending release obligations.
A stub, interface or compilation alone is not contract readiness.

Follow actual ownership/dependency edges. Do not invent private substitute
engines or duplicate transaction, proof, replay, event or authority systems.
Commit/review prerequisite changes before adoption when their cross-repository
contract requires a committed pin. A prerequisite review does not certify later
composition. Final unit verification still requires verified prerequisites and
current composition proof; unrelated release evidence is not a development wall.

## Fixed verification boundaries

During coding: use TDD for new behavior; run the focused affected tests and real
integration boundary where transactional/cross-owner behavior is the contract.
Continue existing tests and retain original red evidence; no ceremonial restart.

Before each consumer milestone commit: run its frozen affected regression,
public-path integration, selected clean-consumer compatibility and relevant
security/concurrency/migration checks. Prove account/tenant isolation,
authorization/current factor policy, enumeration resistance, secret redaction/
custody, origin/CSRF/cookies, atomic consumption/effects, revocation and relevant
denial/rollback/replay/contention/ambiguous outcomes. Race/stress/leak checks
apply to changed concurrency; bounded fuzz applies to changed hostile parsing.
No unsafe behavior can be labeled usable pending later hardening.

Obtain one independent complete-final-diff review of the coherent slice against
requirements, tests and affected consumers. Resolve material findings; a merely
more elaborate plan or speculative variant is not a new acceptance gate.
Retain the reviewer's native identity, exact source snapshot and disposition.

When a complete unit stabilizes, and before marking it verified: run all its
mandatory final gates. Preserve exact per-production-package 100% statement
coverage, every viable mutant killed and independently reviewed equivalent/
invalid dispositions with stable ID, evidence, reviewer, date and expiry.
Preserve applicable official fixtures/interoperability, fuzz/race/stress/leak,
benchmarks, vulnerability/secret/license, SBOM/provenance, documentation,
clean-consumer and affected/aggregate release obligations. Close qualification
incrementally as units stabilize; do not defer all debt to final platform review.
Never change correct algorithms to hide equivalent mutants or lower thresholds.

## Evidence and disposable resources

Reuse unchanged evidence only when complete behavior-affecting input identity
is proven, retaining its original execution revision/result. New HEAD, metadata
or unrelated documentation changes do not invalidate identical runtime inputs.
Changed source, tests, fixtures, dependencies, scripts, tools, service config or
environment invalidate exactly their affected checks. Before expensive reruns,
name the changed input or unproved acceptance they address; otherwise reuse.
Do not repeat the full package matrix after each helper or compaction.

Use existing native checkpoints and compact reporting. Do not create another
fingerprint framework, signature authority, replay journal, receipt, validator,
approval flow, background timer or evidence product. Persist results promptly.
Inspect live handles after interruption; resume live runs, not duplicates.

Every Go run uses task-owned disposable caches/fixtures; a bounded red/green/
regression run may share its own cache/supplier archives. Never use shared global
build caches. Clean exact owned containers/volumes/credentials/scratch after the
bounded run, including failures/interruption. Preserve user resources and carry
applicable runtime authorization; availability alone is not access authority.

## Non-negotiable engineering boundaries

Preserve supported released APIs and consumers. Do not downgrade dependencies,
replace existing engines with private substitutes, invent cryptography or
silently change stored identity normalization. Use fixed provenance-tracked
Unicode tables for stable versioned normalization while supporting Go 1.27.

Keep authentication credential-to-principal validation separate from identity
workflows and authorization. Use explicit typed composition, narrow consumer-
owned interfaces, bounded context/resources and observable ownership/errors.
Capability Verify does not automatically consume: protected transactions own
atomic one-time use. Preserve opaque proof/secret custody and compatible APIs.

Use real PostgreSQL authority with transactional reservation/finalization,
idempotency, rollback and unknown-outcome recovery. Valkey is an explicit
projection/ephemeral store, not independent durable authority. Preserve tenant
isolation, enumeration resistance, origin/CSRF/cookie security, secret redaction,
reveal-once semantics, lifecycle cascades and fresh credential-specific
reauthentication. OTP/capability/risk consumption and required session effects
are atomic with the protected mutation. Fail closed on unavailable authority;
never pass a test by bypassing the security boundary it should prove.

## Complete package scope: all 67 units

Implement full APIs, errors/cancellation, persistence, security, tests, examples,
documentation and changelogs. Nested adapters remain independent modules/units.

### Six primitive extensions

1. `primitive/authentication-identity-contracts`: additive credential/principal,
   opaque proof, assurance and bounded shared-issuer validity contracts.
2. `primitive/authorization-identity-contracts`: existing engine integration and
   identity/organization policy contracts without a duplicate policy service.
3. `primitive/capability-identity-contracts`: opaque purpose-bound proofs and
   one-time workflows while preserving existing capability APIs.
4. `primitive/capability-postgres-identity-contracts`: transactional durable
   reservation/consumption and protected-mutation recovery.
5. `primitive/identifier-identity-contracts`: fixed upstream Unicode tables,
   stable versioned normalization and preserved existing identifiers.
6. `primitive/password-secret-contracts`: owned opaque secrets and intrinsic
   parameter validation distinct from existing policy limits.

### Identity kernel, sessions and delivery: 9 units

- `identity`: users/accounts/identifiers/credentials/verifications, lifecycle,
  repositories, typed extensions, hooks and transactions.
- `identity/postgres`: complete durable kernel and migrations.
- `identity/session`: issuance, cookies/bearer, persistence, rotation, freshness,
  expiry/revocation, devices, account switching, transfer/enrichment and
  consistent remember-me policy across login methods.
- `identity/session/postgres` and `identity/session/valkey`.
- `identity/delivery` and `identity/delivery/postgres`: email/SMS intents,
  templates, bounded sender callbacks, idempotency, retries and recovery.
- `identity/i18n`: locale selection and localized messages/templates.
- `identity/http`: reusable typed opt-in standard-library transport.

### Risk and account-authentication workflows: 17 units

- `identity/risk`, `identity/risk/postgres`, `identity/risk/valkey`: action-bound
  abuse/rate/velocity policy, typed one-use evidence and storage.
- `identity/risk/captcha`, plus nested `recaptcha`, `turnstile`, `hcaptcha` and
  `captchafox`: bounded verification and explicit provider-failure policy.
- `identity/risk/hibp`: privacy-preserving breached-password checking, bounded
  requests and availability policy. HIBP remains required.
- `identity/password` and `identity/password/postgres`: registration/signin,
  verification, reset/change, upgrades and reauthentication.
- `identity/username`: availability/signin/rename with stable normalization.
- `identity/email`: verification and address add/change/list/get/remove policy.
- `identity/magiclink`: purpose/expiry-bound replay-safe passwordless login.
- `identity/otp` and `identity/otp/postgres`: purpose-bound challenges, bounded
  attempts, atomic reservation/finalization and unknown-outcome recovery.
- `identity/phone`: verified phone, OTP/password signin, change/reset, E.164
  normalization and factor/risk-safe recovery.

### Anonymous identity, MFA and passkeys: 8 units

- `identity/anonymous` and `identity/anonymous/postgres`: anonymous sessions,
  self-deletion and collision-safe permanent-account upgrade.
- `identity/mfa` and `identity/mfa/postgres`: TOTP/OTP/recovery codes/trusted
  devices, step-up, enrollment/removal and safe user/administrative recovery.
- `webauthn` and `webauthn/postgres`: protocol ceremonies, validation,
  attestation/assertion policy and durable credentials.
- `passkey` and `passkey/postgres`: identity/session orchestration, passkey-first
  registration, discoverable/usernameless signin and credential management.
  Do not duplicate the WebAuthn protocol engine.

### Social OAuth, API keys and impersonation: 10 units

- `identity/oauth`, `identity/oauth/postgres`, `identity/oauth/providers`,
  `identity/oauth/onetap`, `identity/oauth/proxy`: state/S256 PKCE/callbacks,
  exchange, provider config, token custody/refresh, linking, One Tap and proxy.
- `identity/apikey`, `identity/apikey/postgres`, `identity/apikey/valkey`:
  user/organization ownership, issuance/reveal-once/lookup/verification,
  scopes/expiry/rotation/revocation, legacy digest import and safe projections.
- `identity/impersonation` and `identity/impersonation/postgres`: explicitly
  authorized, bounded, auditable impersonation with owner/actor separation.

### Organizations, federation and provisioning: 11 units

- `organization` and `organization/postgres`: lifecycle, membership,
  invitations, teams, roles/permissions and organization-owned authority.
- `sso`, `sso/domain-verification`, `sso/oidc`, `sso/oauth2`, `sso/saml`,
  `sso/postgres`: enterprise federation, providers/domain ownership, mappings,
  JIT provisioning/linking and protocol replay/lifecycle policy. Federation
  remains distinct from social-provider login.
- `scim`, `scim/postgres`, `scim/organization`: SCIM 2.0 discovery, Users/Groups,
  CRUD/filtering/PATCH/pagination/authorization and directory reconciliation.

### Authorization server and reference composition: 6 units

- `oauth-server`, `oauth-server/oidc`, `oauth-server/device`,
  `oauth-server/postgres`: Golib-owned OAuth/OIDC protocol, clients/consent,
  code/PKCE, refresh rotation, client credentials, restricted token exchange,
  signing/JWKS/discovery, introspection/revocation and device authorization.
  Device flow is in scope, not a mandatory base PostgreSQL bootstrap dependency.
- `identity/reference`: executable complete server, concrete adapters, OpenAPI,
  health/readiness and all-feature journeys. Integrate incrementally.
- `identity/identitytest`: reusable contract suites and integration support.

Administrative management, bans, session control, factor recovery, audit
investigation and privacy workflows belong to the specified public identity
and feature contracts. This is a backend platform, not an administrative UI.

## Required shared-issuer behavior

Implement in the affected existing packages, not a later retrofit:

1. Own OAuth protocol in Golib, not a hosted provider, vendor-owned auth engine
   or commercial identity SDK. Reviewed foundational/crypto libraries are allowed.
2. Policy-enabled restricted RFC 8693 exchange remains distinct from private
   session-JWT exchange. Reject duplicate/conflicting fields, unsupported token
   types, actor overrides and unregistered target audiences.
3. Separate credential owner, verified human actor, tenant/account and workload.
   Client authentication cannot select arbitrary customer identity/tenant.
   Intersect source/current authority; never expand scopes/audiences/permissions.
4. Preserve root credential/grant ID/version, parent, issuer/environment,
   owner/actor/tenant/client/audience/scope across delegation. Reject missing
   roots, cycles and excessive depth. Default maximum is four exchanges;
   configuration may lower it. Child expiry cannot exceed parent or five minutes.
5. Guarantee denial within 30 seconds of revocation, blocking or relevant
   membership removal, including warm caches and lost invalidation events.
   Positive evidence expires within 20 seconds of authoritative read beginning.
   Copies/promotion/retries/exchange cannot renew freshness; earliest contributing
   expiry wins. Fit clock/read budgets to the remaining bound. Expired evidence
   plus failed bounded refresh denies; no stale authorization fallback.
6. Authorized digest-only import and unchanged verification of existing bare
   SHA-256 API keys: no plaintext, invented prefixes or forced rotation.
   Preserve stable IDs, ownership/tenant/restrictions/versions/tombstones.
   Idempotent import cannot resurrect revocation. New issuance remains modern
   HMAC-with-pepper; rotation creates a modern successor.
7. Organization-owned keys survive creator membership removal when otherwise
   valid. Removal denies that human's management/delegation. Owner suspension
   or key revocation denies affected authority; restore never revives revoked keys.
8. Preserve validated canonical IDs including ULIDs through injected identifier
   policy/source mappings. Reject conflicts and stale/tombstone resurrection.
9. HTTP features are opt-in: disabled features expose no routes, metadata or
   advertised capability. Selected features require actual dependencies.
   Base OAuth PostgreSQL builds without device; base HTTP/OAuth does not require
   password, social OAuth, enterprise SSO or SCIM implementations.
10. Prove all ten shared-issuer cases: cache/outage timing, lineage denial,
    legacy import, populated migrations, clean selected composition and distinct
    standard/private exchange behavior.

Login bridging, billing, source-system migration, deployment, application
integration and MCP-specific qualification remain Mono/API or consumer work.
Billing, SIWE, MCP/agent-specific packages and speculative integrations are
excluded. HIBP, CAPTCHA and device authorization are not excluded.

## End-state acceptance and source authority

The reference server composes public APIs, standard-library `net/http`, real
PostgreSQL and selected Valkey/provider adapters with OpenAPI 3.1.1. Consumers
supply configuration, credentials, templates, explicit bounded policy callbacks
and email/SMS senders, not missing domain stores, handlers, migrations or workflows.

Prove all nineteen journey categories: identity lifecycle; email/password/
username; sessions; passwordless/phone/anonymous; MFA/WebAuthn/passkeys; social
OAuth; API keys; administration; organizations; enterprise federation; SCIM;
OAuth/OIDC provider; abuse/CAPTCHA/HIBP; localization; developer adoption;
typed extension/access policy; typed modules/operations; privacy export;
audit investigation.

Preserve sensitive-change reauthentication, session transfer/remember-me,
safe email removal, phone password/reset, recovery-code reveal/regeneration,
independently approved administrative factor recovery, deletion/anonymization
cascades and legal retention. Prove transport security, recovery, privacy,
operability, provider interoperability and clean-consumer composition.

Detailed package goals and typed operation/protocol/transaction/lifecycle/
configuration/parity/acceptance catalogs remain product requirements, not
authority to restart orchestration. `INVENTORY.md`, `GOAL_MANIFEST.json`,
`SHARED_ISSUER_PROFILE.md`, `END_STATE.md`, `END_STATE_ACCEPTANCE.json` and
`BETTER_AUTH_PARITY.md` retain their complete obligations. Correct contradictory
catalog shapes against current product authority; do not weaken behavior to
make stale generators pass. Do not reread the entire catalog before each edit.

## Progress, delivery stop-loss and full completion

Maintain one concise current-milestone note in EXECUTION_LEDGER.md and keep
IDENTITY_PROGRESS.md current at material checkpoints and at least every six
hours of active work. Preserve old results, assignment identities and pending
qualification. Do not create additional reporting frameworks or timers.

Report delivered consumer milestones out of 39 separately from integrated and
fully verified units out of 67, product units out of 61 and journey/shared-issuer
acceptance. Fractions are unweighted counts, not an estimate of work effort.
Any overall percentage must state basis and confidence; do not keep repeating
the historical rough 5–10% as a measured result. Do not count helpers, hashes,
test counts, commits or procedural activity as acceptance progress.
Use the required status skill for requested and hourly consolidated updates.

At material checkpoints, state what consumers can actually do, what remains
disconnected, the next operation and pending qualification. Distinguish local,
committed, published, hosted CI, deployed and production evidence. Do not invent
a completion ETA or relabel old evidence as fresh execution.

If two consecutive six-hour active-work checkpoints close no accepted consumer
outcome, diagnose the concrete obstacle and change strategy or split oversized
scope into smaller complete consumer outcomes. Helper activity does not defeat
this trigger. Keep security intact; the trigger is not permission to rush or
claim incomplete acceptance. Use the existing note, not another planning cycle.
Two correction/review cycles failing to resolve the same concrete defect require
a materially different strategy; do not rename and repeat the mechanism.
Ask only for genuinely unavailable material scope/security/strategy choices or
external authority, not clerical approval or exact-byte replays.

Full completion requires all 67 units integrated and verified, all pinned parity
obligations, nineteen journeys and ten shared-issuer cases proven, complete
public/reference/consumer composition, applicable final gates and independent
reviews, documented APIs/examples/migrations/changelogs, and resource cleanup.
No skipped/stale/failed/unavailable/warning-only mandatory result is a pass.
Milestones or local commits do not imply release, hosted CI or production readiness.

Continue across turns toward that full authorized outcome. Stop on actual
completion, explicit user pause/stop, or a genuine repeated impasse under the
goal's blocked audit. Do not mark blocked for procedural friction while safe
in-scope implementation remains.
