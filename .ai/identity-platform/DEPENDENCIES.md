# Identity Platform Dependencies

The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT",
"SHOULD", "SHOULD NOT", "RECOMMENDED", "NOT RECOMMENDED", "MAY", and
"OPTIONAL" in this document are to be interpreted as described in BCP 14
[RFC2119] [RFC8174] when, and only when, they appear in all capitals, as
shown here.

## Status model

`proposed` means scoped but not authorized. `ready` means all start gates
are satisfied and one agent may claim the unit. `in-progress` has one owner.
`implemented-unverified` has integrated implementation but incomplete or stale
release evidence. A committed, integrated, focused-contract-tested prerequisite
at `implemented-unverified` MAY satisfy development readiness under
ORCHESTRATOR_GOAL.md; only `verified` satisfies a dependant's final verification.
`blocked` is an
inventory execution state with an explicit safe blocker identifier; it is not
the coordinator agent's system goal status.

```text
proposed -> ready -> in-progress -> implemented-unverified -> verified
         <- ready (prerequisite invalidation; generation unchanged)
                  \-> blocked -> in-progress (same paused assignment)
                  \-> blocked -> ready (assignment abandoned, generation + 1)
                  \-> ready (abandoned assignment, generation + 1)
implemented-unverified -> in-progress (same owner/branch repair)
implemented-unverified -> blocked -> in-progress (same assignment repair)
verified -> implemented-unverified (changed complete inputs)
```

A unit MUST NOT become `ready` unless all `Requires` units meet the amended
development-readiness rule. Status alone is insufficient: pin the integrated
revision and consumed contract evidence in the existing assignment.
An unassigned `ready` unit MUST return to `proposed`, without changing its
generation or empty assignment/evidence fields, in the same coordinator commit
that invalidates a consumed prerequisite contract. It becomes `ready` again only after every
start gate is current.
An initial `proposed` row MUST have generation `0`. A post-initial `proposed`
row retains the current generation, which remains `0` before any assignment and
may be greater than `0` after an assignment was abandoned. History validation
MUST prove its only incoming edge was `ready -> proposed` prerequisite
invalidation; no other status may return to `proposed`.
It MUST return to `implemented-unverified` when changed complete inputs
invalidate its evidence.
An `in-progress` unit MAY return to `ready` only after the coordinator proves
no unintegrated package work is discarded and increments its generation. A
`blocked` unit that retains its ledger assignment MAY return directly to
`in-progress` under that exact task, branch, worktree, and generation after its
blocker is resolved. For conflict or stale-baseline recovery, that transition
MUST be the coordinator resume/authorization commit. Its recovery row MUST pin
the commit's first-parent pre-resume integration baseline and the worker's clean
checkpoint; before any package edit the worker MUST merge that exact resume
commit and verify both ancestry relationships. A `blocked` unit whose
assignment is abandoned MUST return to `ready`, increment generation, clear
active assignment fields, and receive a new assignment before work resumes.
Repair of an integrated unit retains its
recorded assignment identity and uses `implemented-unverified -> in-progress`;
a replacement owner requires a new generation. If a prerequisite loses
contract validity, every transitive dependant whose complete input fingerprint
changes MUST be demoted, every unassigned `ready` dependant MUST return to
`proposed`, and every active dependant on that stale baseline MUST transition
to inventory `blocked` and pause until refreshed.

The development-first amendment also supersedes the legacy cross-repository
resume-commit merge and custody procedures above. Refresh package ancestry
only with package-repository commits; never merge a control commit into a
package. An unrelated pending release gate MUST NOT block development.

For `proposed`, `ready`, `implemented-unverified`, and `verified`, the inventory
owner/blocker MUST be `—`. For `in-progress`, it MUST equal the recorded worker
task. For `blocked`, it MUST be a whitespace-free `blocker:<safe-id>` and the
ledger assignment fields determine whether the same assignment is retained.
Using inventory `blocked` does not call `update_goal`, does not count as a
system blocked turn, and does not authorize the coordinator to stop while other
safe work remains. The coordinator MAY report its own system goal as blocked
only under repository blocked-audit rules and the stop conditions in
`ORCHESTRATOR_GOAL.md`.

## Dependency DAG

`SHARED_ISSUER_PROFILE.md` defines the focused runtime milestone. Base HTTP
requires identity and authorization contracts, not every optional feature.
OAuth PostgreSQL does not require device flow to construct its base adapter.
Selected-feature implementations and all-feature reference acceptance still
require their actual feature contracts. Removed bootstrap edges do not remove
behavior, waive proof or mark unfinished units verified.

```mermaid
flowchart TD
  authentication[primitive/authentication-identity-contracts]
  authorization[primitive/authorization-identity-contracts]
  capability[primitive/capability-identity-contracts]
  capability_pg[primitive/capability-postgres-identity-contracts]
  identifier[primitive/identifier-identity-contracts]
  primitive_password[primitive/password-secret-contracts]
  session[identity/session]
  delivery[identity/delivery]
  delivery_pg[identity/delivery/postgres]
  risk[identity/risk]
  captcha[identity/risk/captcha]
  recaptcha[identity/risk/captcha/recaptcha]
  turnstile[identity/risk/captcha/turnstile]
  hcaptcha[identity/risk/captcha/hcaptcha]
  captchafox[identity/risk/captcha/captchafox]
  hibp[identity/risk/hibp]
  email[identity/email]
  password[identity/password]
  password_pg[identity/password/postgres]
  username[identity/username]
  magiclink[identity/magiclink]
  otp[identity/otp]
  otp_pg[identity/otp/postgres]
  phone[identity/phone]
  anonymous[identity/anonymous]
  anonymous_pg[identity/anonymous/postgres]
  mfa[identity/mfa]
  mfa_pg[identity/mfa/postgres]
  webauthn_pg[webauthn/postgres]
  oauth[identity/oauth]
  oauth_pg[identity/oauth/postgres]
  oauth_providers[identity/oauth/providers]
  oauth_onetap[identity/oauth/onetap]
  oauth_proxy[identity/oauth/proxy]
  apikey[identity/apikey]
  apikey_pg[identity/apikey/postgres]
  apikey_valkey[identity/apikey/valkey]
  impersonation[identity/impersonation]
  impersonation_pg[identity/impersonation/postgres]
  authentication --> anonymous
  authentication --> email
  authentication --> magiclink
  authentication --> mfa
  authentication --> otp
  authentication --> password
  authentication --> phone
  authentication --> username
  authentication --> passkey
  authentication --> webauthn
  authorization --> identity
  authorization --> apikey
  authorization --> identity_http
  authorization --> impersonation
  authorization --> reference
  authorization --> session
  authorization --> organization
  capability --> identity
  capability --> email
  capability --> magiclink
  capability --> oauth
  capability --> oauth_onetap
  capability --> oauth_pg
  capability --> oauth_proxy
  capability --> otp
  capability --> password
  capability --> phone
  capability --> session
  capability --> organization
  capability --> sso
  capability --> sso_saml
  capability --> sso_pg
  capability --> oauth_server_oidc
  capability --> oauth_server_device
  capability --> oauth_server
  capability --> webauthn
  capability --> capability_pg
  capability_pg --> mfa_pg
  capability_pg --> oauth_pg
  capability_pg --> sso_pg
  capability_pg --> oauth_server_pg
  capability_pg --> reference
  identifier --> email
  identifier --> magiclink
  identifier --> otp
  identifier --> password
  identifier --> username
  identifier --> passkey
  identifier --> webauthn
  primitive_password --> otp
  primitive_password --> password
  primitive_password --> phone
  primitive_password --> username
  identity --> identity_pg[identity/postgres]
  identity_pg --> capability_pg
  identity --> session
  session --> session_pg[identity/session/postgres]
  identity_pg --> session_pg
  session --> session_valkey[identity/session/valkey]
  identity --> risk
  identity_pg --> risk_pg[identity/risk/postgres]
  risk --> risk_pg[identity/risk/postgres]
  risk --> risk_valkey[identity/risk/valkey]
  risk --> captcha
  captcha --> recaptcha
  captcha --> turnstile
  captcha --> hcaptcha
  captcha --> captchafox
  risk --> hibp
  identity --> email
  delivery --> email
  otp --> email
  delivery --> delivery_pg
  identity_pg --> delivery_pg
  identity --> password
  session --> password
  risk --> password
  delivery --> password
  otp --> password
  password --> password_pg
  identity_pg --> password_pg
  identity --> username
  password --> username
  email --> magiclink
  identity --> magiclink
  session --> magiclink
  risk --> magiclink
  identity --> otp
  delivery --> otp
  session --> otp
  risk --> otp
  otp --> otp_pg
  identity_pg --> otp_pg
  otp --> phone
  password --> phone
  identity --> phone
  delivery --> phone
  risk --> phone
  identity --> anonymous
  session --> anonymous
  risk --> anonymous
  anonymous --> anonymous_pg
  identity_pg --> anonymous_pg
  session --> anonymous_pg
  otp --> mfa
  identity --> mfa
  identity --> webauthn
  session --> mfa
  risk --> mfa
  webauthn --> mfa
  mfa --> mfa_pg
  identity_pg --> mfa_pg
  webauthn_pg --> mfa_pg
  webauthn --> webauthn_pg
  identity_pg --> webauthn_pg
  webauthn --> passkey
  identity --> passkey
  session --> passkey
  risk --> passkey
  passkey --> passkey_pg[passkey/postgres]
  identity_pg --> passkey_pg
  webauthn_pg --> passkey_pg
  identity --> oauth
  session --> oauth
  risk --> oauth
  oauth --> oauth_pg
  identity_pg --> oauth_pg
  oauth --> oauth_providers
  oauth --> oauth_onetap
  oauth_providers --> oauth_onetap
  session --> oauth_onetap
  oauth --> oauth_proxy
  session --> oauth_proxy
  identity --> apikey
  organization --> apikey
  apikey --> apikey_pg
  identity_pg --> apikey_pg
  organization_pg --> apikey_pg
  apikey --> apikey_valkey
  identity --> impersonation
  session --> impersonation
  risk --> impersonation
  impersonation --> impersonation_pg
  identity_pg --> impersonation_pg
  session_pg --> impersonation_pg
  identity --> organization
  session --> organization
  delivery --> organization
  organization --> organization_pg[organization/postgres]
  identity_pg --> organization_pg
  organization --> sso
  identity --> sso
  session --> sso
  risk --> sso
  sso --> sso_oidc[sso/oidc]
  sso --> sso_oauth2[sso/oauth2]
  sso --> sso_saml[sso/saml]
  sso --> sso_pg[sso/postgres]
  organization_pg --> sso_pg
  identity_pg --> sso_pg
  sso --> domain_verify[sso/domain-verification]
  organization --> domain_verify
  organization --> scim
  identity --> scim
  scim --> scim_pg[scim/postgres]
  organization_pg --> scim_pg
  identity_pg --> scim_pg
  scim --> scim_org[scim/organization]
  organization --> scim_org
  identity --> scim_org
  sso --> scim_org
  scim_org --> scim_pg
  identity --> oauth_server[oauth-server]
  session --> oauth_server
  risk --> oauth_server
  oauth_server --> oauth_server_oidc[oauth-server/oidc]
  oauth_server --> oauth_server_device[oauth-server/device]
  oauth_server --> oauth_server_pg[oauth-server/postgres]
  identity_pg --> oauth_server_pg
  identity --> i18n[identity/i18n]
  identity --> identity_http[identity/http]
  identity_http --> reference[identity/reference]
  identity --> reference
  anonymous --> reference
  apikey --> reference
  delivery --> reference
  email --> reference
  i18n --> reference
  impersonation --> reference
  magiclink --> reference
  mfa --> reference
  oauth --> reference
  oauth_onetap --> reference
  oauth_providers --> reference
  oauth_proxy --> reference
  otp --> reference
  password --> reference
  phone --> reference
  risk --> reference
  captcha --> reference
  session --> reference
  username --> reference
  oauth_server --> reference
  oauth_server_device --> reference
  oauth_server_oidc --> reference
  organization --> reference
  passkey --> reference
  scim --> reference
  scim_org --> reference
  sso --> reference
  sso_oauth2 --> reference
  sso_oidc --> reference
  sso_saml --> reference
  webauthn --> reference
  delivery_pg --> reference
  anonymous_pg --> reference
  domain_verify --> reference
  identity_pg --> reference
  session_pg --> reference
  session_valkey --> reference
  risk_pg --> reference
  risk_valkey --> reference
  recaptcha --> reference
  turnstile --> reference
  hcaptcha --> reference
  captchafox --> reference
  hibp --> reference
  password_pg --> reference
  otp_pg --> reference
  mfa_pg --> reference
  webauthn_pg --> reference
  passkey_pg --> reference
  oauth_pg --> reference
  apikey_pg --> reference
  apikey_valkey --> reference
  impersonation_pg --> reference
  organization_pg --> reference
  sso_pg --> reference
  scim_pg --> reference
  oauth_server_pg --> reference
  reference --> identitytest[identity/identitytest]
```

The authoritative inventory currently resolves to 67 schedulable units and
241 direct edges across nine zero-based waves. Wave populations are
`0:6, 1:1, 2:6, 3:16, 4:21, 5:14, 6:1, 7:1, 8:1`. The 61-unit canonical
identity-platform scope is unchanged; the sixth prerequisite unit is the
ordered PostgreSQL capability extension.

The `Requires` field in `INVENTORY.md` is authoritative. The diagram is
explanatory and MUST change in the same commit when an edge changes.
