# Shared Issuer First-Release Profile

The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT",
"SHOULD", "SHOULD NOT", "RECOMMENDED", "NOT RECOMMENDED", "MAY", and
"OPTIONAL" in this document are to be interpreted as described in BCP 14
[RFC2119] [RFC8174] when, and only when, they appear in all capitals, as
shown here.

## Authority and scope

This is the 2026-09-30 user-authorized amendment to the existing identity
goals, not a second identity implementation or a replacement full-program goal.
The user directed: "then amend them and give me a msg i can send to the agent
that already started working on them so he knows what changed" after selecting
Golib-owned protocols and identifying delegation, revocation, legacy-key and
first-release dependency gaps. The requirements below are part of the affected
existing goals and MUST NOT be postponed to a follow-up after their APIs ship.

The profile ID is `shared-issuer-v1`. It is a reusable library profile; Shipit
is its first consumer, not a hard-coded dependency. Golib MUST own the OAuth
protocol implementation. A hosted identity provider, vendor-owned authorization
server engine or commercially controlled identity SDK MUST NOT be required.
Standard cryptography and narrowly reviewed foundational libraries remain
permitted; implementing cryptographic algorithms is not part of this goal.

For this selected profile, this file takes precedence over older reference
defaults, three-grant token request shapes, session-only exchange exclusions,
and all-feature bootstrap requirements. Other profiles retain their existing
requirements. The full 61-product-unit / 67-schedulable-unit completion contract
is unchanged. A first-release milestone MUST NOT be called `PROGRAM-COMPLETE`.
Existing goal and public-contract identifiers remain stable. Typed contracts
and acceptance catalogs MUST incorporate this profile before an affected
implementation is claimed verified; older generated shapes MUST NOT be used
to omit these requirements. Catalog reconciliation is implementation work,
not a new user approval ceremony or a recursive evidence program.

## Ownership and consumer boundary

| Capability | Existing owner | Required reusable boundary |
| --- | --- | --- |
| Canonical user identity, status, external ID mapping | `identity`, `identity/postgres` | Preserve canonical IDs and versions; deny disabled or deleted identities |
| Organizations, membership and tenant authority | `organization`, `organization/postgres` | Typed owner versus human actor; current membership and permission resolution |
| Legacy and new API keys | `identity/apikey`, `identity/apikey/postgres` | Digest-only import, verification profiles, rotation and root credential lifecycle |
| OAuth clients, grants, delegated issuance | `oauth-server`, `oauth-server/postgres` | One issuer; workload authentication separate from tenant delegation |
| JWKS and discovery | `oauth-server/oidc` with core signing ownership | Public verification material without requiring social login or enterprise SSO |
| Resource validation and freshness | `oauth-server`, `primitive/authentication-identity-contracts` | Local signature validation plus bounded authoritative validity evidence |
| Optional cache projections | Existing Valkey adapters | Never an independent authority; original source freshness cannot be reset |
| Selected HTTP composition | `identity/http` | Register only selected, implemented features; no mandatory all-feature runtime |

The deployed application, source-system importer, login bridge, billing,
cutover, domain resource authorization, issuer hostname and operational secrets
belong to the consuming service, not to these libraries. For the first consumer,
those are Mono/API responsibilities. Golib MUST NOT acquire an API database,
Laravel model, billing dependency, Shipit header name or deployment platform.
The independent service will be the issuer for API, Mono applications and MCP;
MCP tooling, entitlement policy and client qualification remain consumer work.

## Typed directory and organization-owned keys

Owner identity MUST be a validated tagged user-or-organization value, bound to
its canonical ID, tenant and authoritative version. Credential owner, optional
authenticated human actor, selected tenant/account and calling workload MUST
remain distinct. Financial payer is not an authentication claim owner.

Canonical identity and organization IDs MUST support preservation of existing
validated identifiers, including canonical ULIDs, through injected identifier
policy and external-source mappings. An import MUST NOT silently generate a new
identity for an existing owner or infer ownership from an ID string alone.
Source versions, tombstones and duplicate/conflicting mappings MUST have
idempotent, typed outcomes. Importing status or membership data MUST NOT bypass
authorization or restore revoked credentials.

Organization-key creation/management MUST authorize the human actor's current
membership and required permission. Verification of an organization-owned key
authenticates the credential/integration principal, not its creator as a human.
Removing that member MUST deny their management and human delegation without
automatically revoking unrelated organization-owned keys. Organization archive,
deletion, suspension or a reduced authority ceiling MUST remove the affected
key authority. Reactivation MUST NOT resurrect terminally revoked keys.

## Restricted exchange and credential lineage

`oauth-server` MUST support a policy-enabled restricted
[RFC 8693 token exchange](https://www.rfc-editor.org/rfc/rfc8693.html) at the
existing token endpoint, alongside authorization code with S256 PKCE, refresh
rotation and client credentials. This is distinct from the existing private
`urn:golib:params:oauth:grant-type:session-jwt` extension. The private extension
MUST NOT be presented as standard token exchange or used as its substitute.

The selected grant identifier is
`urn:ietf:params:oauth:grant-type:token-exchange`. Parsing MUST discriminate the
grant before processing credentials, reject duplicate/conflicting parameters,
and require `subject_token` and `subject_token_type`. The initial profile
accepts only issuer-owned OAuth access tokens and the explicit private API-key
type `urn:golib:params:oauth:token-type:api-key`. It issues an OAuth access token
with `issued_token_type=urn:ietf:params:oauth:token-type:access_token`; raw API
keys MUST NOT be described as standard OAuth access tokens. An `actor_token`
MUST have its matching type and independently verified authority; the initial
profile MAY reject actor-token inputs entirely rather than approximate them.

The authenticated OAuth client identifies the calling workload. Client
authentication alone MUST NOT authorize arbitrary user, organization or tenant
selection. No free-form subject or human actor override is permitted. A
selected tenant/account is a typed policy input checked against the source
credential's authority and, for human delegation, current membership. Exactly
one configured target resource/audience is selected for each exchange;
conflicting target selectors or an unregistered target MUST fail closed.

The result MUST intersect the immutable source grant, current credential and
owner authority, applicable human membership, selected tenant, caller allowance
and target capability. It MUST NOT add scopes, change the owner, fabricate a
human actor or broaden audience authority. An organization-key source does not
require a human merely to authenticate the integration. Workload-only client
credentials remain machine authority, not customer delegation.

Every derived token MUST preserve issuer, environment, root credential/grant
ID and version, parent token/grant linkage, typed owner, selected tenant,
verified human actor when present, authenticated client/caller, exact audience,
scope and expiry. A bounded delegation chain MUST reject cycles, missing roots
and excessive depth. The default maximum depth is four exchanges; configuration
MAY lower it. Child expiry MUST NOT exceed parent expiry or five minutes.
Access JWTs MUST carry `iss`, `aud`, `sub`, `client_id`, `iat`, `exp`, `jti`
and explicit access-token purpose, plus the profile bindings above. ID tokens,
wrong-environment tokens and workload tokens without delegated tenant authority
MUST NOT be accepted as customer access credentials.

Root revocation, refresh-family replay revocation, owner blocking, relevant
membership removal and explicit parent revocation MUST invalidate affected
descendants within the freshness bound below. PostgreSQL owns durable lineage,
version/status indexes and reconciliation. Ambiguous issuance MUST reconcile
before retry; it MUST NOT create another independent usable delegation grant.
Derived grants MUST NOT create billable API-key records or rotate source keys.
Resource services receive scoped tokens, never the source raw API key.

## Bounded revocation and local verification

The selected profile MUST deny new authorization decisions no later than
30 seconds after authoritative credential/grant revocation, owner/account
blocking or relevant membership removal, including all positive cache paths.
JWT lifetime alone and best-effort invalidation messages do not prove this.

A resource verifier MUST check signature, algorithm, issuer, environment,
audience, purpose, expiry and scope locally, then require current validity for
every applicable root/parent grant, owner, tenant, human membership, client and
signing-compromise dimension. Typed authoritative validity evidence MUST bind
those exact dimensions and versions, its source `verified_at`, and absolute
`valid_until`. Authenticated transport or an issuer-authenticated evidence
envelope MUST prevent the caller from fabricating those values.

Positive evidence MUST expire at most 20 seconds after the authoritative read
begins, not after its response arrives or a cache entry is inserted. When
several sources contribute, the earliest expiry wins. Every intermediary MUST
preserve the original timestamps; token exchange, copying, cache promotion and
retries MUST NOT restart the clock. The profile reserves at most ten seconds
in total for bounded clock uncertainty and read/decision races. Deployment
configuration MUST account for that budget explicitly and reject combinations
that cannot meet 30 seconds; an unbounded read or clock assumption is invalid.

A warm verifier MUST NOT require a legacy API lookup or a central lookup per
request. It MAY reuse still-valid authoritative evidence, but stale, missing,
unavailable or mismatched evidence MUST trigger a bounded authority refresh or
deny. Refresh failure MUST fail closed; no stale-while-revalidate authorization
is allowed. Events MAY accelerate invalidation but MUST NOT be required for
the bound. A self-contained JWT or cookie without this evidence cannot claim
the selected revocation guarantee. The existing five-minute reference cache
policy remains available only to other explicitly selected profiles.

## Legacy API-key verification and import

The API-key core and PostgreSQL adapter MUST support a named verify-only legacy
profile for existing bare secrets whose stored verifier is an unkeyed SHA-256
digest. The existing HMAC-SHA-256-with-pepper profile remains the default for
new issuance; legacy verification MUST NOT weaken the new-key generator or
enable new issuance under the legacy configuration.

An authenticated, authorized import command MUST accept only digest and
metadata: stable key ID, typed owner and tenant binding, source ID/version,
configuration/algorithm revision, permission ceiling, expiry, status/revocation
and replacement lineage. It MUST preserve these values without the original
secret. Repeated identical imports are idempotent; conflicting same-ID imports,
newer tombstones and attempts to resurrect revoked material MUST be rejected
or reconciled as typed outcomes. A bare legacy key MUST be lookup-verifiable
without inventing a prefix absent from the historical record. Configuration
selection is server-owned; attacker-controlled text cannot choose a weaker
algorithm or trigger unbounded profile/digest scans.

SHA-256 compatibility fixtures MUST prove existing keys authenticate unchanged,
wrong secrets fail, user/org owners remain distinct, imported restrictions are
not widened, and rotation creates a successor under the current issuance
profile. Imports MUST NOT count as new paid-key issuance. Source-system
encrypted reveal copies and their custody/replacement are consumer migration
work; the library MUST NOT reconstruct, store or re-reveal legacy plaintext.

## First-release dependency slice

The reusable runtime slice is identity/directory, sessions needed by the login
handoff, organizations, API keys, OAuth core, public JWKS/discovery, resource
verification, HTTP composition and their PostgreSQL support. Existing
authentication, authorization, capability, audit, outbox, risk/rate-limit,
identifier and secret-envelope primitives supply their applicable contracts.
`identity/delivery` remains an organization contract prerequisite; selected
policy MAY disable invitations. Valkey is optional, not a second authority.

`oauth-server/postgres` MUST be constructible and releasable for code, refresh,
client-credentials and restricted-exchange flows without device flow. Device
storage implementation MUST be isolated behind an optional device adapter
subpath or nested module, not imported by the base adapter. Device integration
and its full-program proof remain required when that feature is selected.

`identity/http` MUST expose typed opt-in feature registration and reject
missing dependencies for a selected feature. Its base MUST NOT require every
password, passkey, social OAuth, enterprise SSO, SCIM or device implementation
to build or start. Disabled features MUST register no routes, publish no
metadata and claim no capability. All existing in-scope features still need
their eventual composed HTTP proof; `identity/reference` remains the full
all-feature composition, not a first-release prerequisite. Concrete consumer
handlers and deployment are not moved into Golib by this milestone.

Satisfying this runtime milestone MUST NOT mark unfinished units `verified`.
An optional integration may remain outstanding in its existing unit until its
normal complete-goal proof exists. Removing artificial bootstrap edges does
not remove feature behavior or full-program acceptance. MCP-specific DCR/CIMD
client qualification MUST be resolved in the consuming service; unsupported
registration mechanisms MUST be rejected and omitted from metadata, not
advertised merely because basic OAuth works.

## Focused acceptance cases

These are additions to existing package tests and composed contract fixtures,
not a new test harness, receipt system or review-count requirement.

| ID | Observable assertion | Owners |
| --- | --- | --- |
| SI-01 | Validated canonical IDs survive directory import; duplicate conflicts, stale updates and tombstone resurrection are denied | identity, organization and PostgreSQL adapters |
| SI-02 | Removing a key creator denies their management/delegation but leaves an otherwise valid organization key usable; owner block denies it | organization, API-key and OAuth core |
| SI-03 | Workload credentials alone cannot select an arbitrary tenant; exchange narrows scopes/audience and preserves owner/actor/caller separation | OAuth core and resource verifier |
| SI-04 | Root or parent revocation denies descendants; cycles, excess depth, wrong environment, missing lineage and ID-token substitution fail | OAuth core, PostgreSQL and verifier |
| SI-05 | With invalidation events lost and all caches warm, revocation/block/member removal is denied within 30 seconds; cache copying cannot renew evidence | verifier and applicable cache adapters |
| SI-06 | Authority outage after freshness expires denies; warm valid evidence avoids per-request central lookups; read delay and clock budgets cannot extend acceptance | verifier and validity source |
| SI-07 | Existing bare SHA-256 keys verify without rotation; import retries preserve IDs and restrictions without plaintext, new issuance or duplicate billing events | API-key core and PostgreSQL |
| SI-08 | Fresh-state and populated-state PostgreSQL migrations preserve delegation lineage, revoked tombstones and reveal-once/reconciliation semantics under retries | PostgreSQL adapters |
| SI-09 | A clean consumer builds and starts the selected HTTP/OAuth/PostgreSQL composition without device, password, social OAuth, SSO or SCIM implementations | HTTP, OAuth PostgreSQL and composition owners |
| SI-10 | Standard exchange and the private session extension have distinct parsing/metadata; unsupported token types, actors, registration and grant fields fail closed | OAuth core, OIDC and HTTP |

## Adoption by the active agent

Preserve existing implementation and ownership. Synchronize the amended planning
inputs into the active coordinator control snapshot in `go-library-tools` using
the existing ownership boundary; do not overwrite its runtime state or copy
package implementation back into this coordination directory. Read this file
with the amended goals before continuing an affected public API, schema or
verifier. Reconcile
the affected typed-contract, protocol and acceptance catalog rows as part of
those owned implementation batches, and assess existing code against SI-01
through SI-10. Do not restart unrelated goals or rerun unrelated expensive
checks. Evidence for unchanged behavior remains usable; prior evidence does
not certify the new profile. Report the focused milestone separately from the
unchanged full-program completion predicate.
