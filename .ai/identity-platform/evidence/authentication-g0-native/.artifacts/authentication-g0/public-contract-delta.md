# Exact authentication catalog delta

Generation 0, owner github.com/faustbrian/go-authentication root; no coordinator
files edited. Compared against the current authentication external_contracts
projection in fragments/public_contracts_meta.json. Original required types,
interfaces, methods and errors remain present with their exact declarations.
Existing Principal/Authenticator/credential/result/error APIs are unchanged.
This records implemented declarations, not verification or catalog approval.

The complete Go-shaped declarations for the following additive shared-issuer
types, constants, interface and methods are in shared-contract-proposal.md;
their source authority is authority_contracts.go and authority.go:

## New types versus the current catalog

```go
type AuthorityOwnerKind string
// Closed constants: AuthorityOwnerUnspecified="", AuthorityOwnerUser="user",
// AuthorityOwnerOrganization="organization", AuthorityOwnerWorkload="workload".
type AuthorityOwner struct { Kind AuthorityOwnerKind; ID string; Version uint64 }
type AuthorityGrant struct { ID string; Version uint64; ExpiresAt time.Time }
type AccessTokenClaims struct {
    Issuer string; Audience string; Subject string; ClientID string
    IssuedAt int64; ExpiresAt int64; TokenID string; Purpose Purpose
    Environment string; Method string; Assurance Assurance; Owner AuthorityOwner
    Actor *AuthorityOwner; Tenant string; TenantVersion uint64; ClientVersion uint64
    SigningVersion uint64; MembershipID string; MembershipVersion uint64
    RootCredential AuthorityGrant; RootGrant AuthorityGrant
    Parents []AuthorityGrant; Scopes []string
}
type ValidityDimensionKind string
// Closed constants: ValidityRootCredential="root-credential",
// ValidityRootGrant="root-grant", ValidityParentGrant="parent-grant",
// ValidityOwner="owner", ValidityActor="actor", ValidityTenant="tenant",
// ValidityMembership="membership", ValidityClient="client", ValiditySigning="signing".
type ValidityDimension struct { Kind ValidityDimensionKind; ID string; Version uint64 }
type DimensionValidity struct { Dimension ValidityDimension; Active bool; VerifiedAt time.Time; ValidUntil time.Time }
type ValidityRequest struct { Claims AccessTokenClaims }
type SignedValidityEvidence struct { Issuer string; Environment string; TokenID string; Tenant string; Dimensions []DimensionValidity; KeyID string; Signature []byte }
type SignedAuthenticationRecord struct { Issuer string; Environment string; State SignInState; Permanent bool; Proof Proof; PermanentSubject string; CredentialID string; CredentialVersion uint64; KeyID string; Signature []byte }
type AuthorityKey struct { ID string; PublicKey ed25519.PublicKey }
type AuthorityConfig struct {
    Issuer string; Environment string; Audience string; Algorithm string; Profile string
    AccessKeys []AuthorityKey; EvidenceKeys []AuthorityKey; Source AuthoritySource
    Clock clock.Clock; ReadTimeout time.Duration; ClockUncertainty time.Duration
    MaximumDelegationDepth int; RequiredScopes []string
}
type AuthorityAuthenticator struct { /* all fields private */ }
type ResourcePrincipal struct { /* all fields private */ }
```

Actual JWT JSON tags, in AccessTokenClaims declaration order:
iss, aud, sub, client_id, iat, exp, jti, purpose, environment, method, assurance,
owner, actor,omitempty, tenant, tenant_version, client_version, signing_version,
membership_id,omitempty, membership_version,omitempty, root_credential,
root_grant, parents, scopes. AuthorityOwner tags: kind,id,version.
AuthorityGrant tags: id,version,expires_at. Header tags: alg,typ,kid.
Exact case-sensitive lowercase ASCII member admission applies recursively to
these JWT objects; duplicates/unknown/case-folded/Unicode aliases are rejected.
The typed signed-evidence SigningBytes encodings are separate: they retain their
documented canonical Go field names and domain separation, not JWT member tags.

## New interface

```go
type AuthoritySource interface {
    CurrentValidity(context.Context, ValidityRequest) (SignedValidityEvidence, error)
    PermanentAuthentication(context.Context, ProofID) (SignedAuthenticationRecord, error)
}
```

## New methods/functions versus the current catalog

```go
func NewAuthorityAuthenticator(AuthorityConfig) (*AuthorityAuthenticator, error)
func (*AuthorityAuthenticator) Authenticate(context.Context, Credential) (Result, error)
func (*AuthorityAuthenticator) ValidateAccessToken(context.Context, string) (ResourcePrincipal, error)
func (*AuthorityAuthenticator) ValidatePermanentAccountProof(PermanentAccountProof, ProofUse) error
func (ResourcePrincipal) Bindings() AccessTokenClaims
func (ResourcePrincipal) ValidUntil() time.Time
func (ResourcePrincipal) Validate(time.Time) error
func (SignedValidityEvidence) SigningBytes() ([]byte, error)
func (SignedAuthenticationRecord) SigningBytes() ([]byte, error)
func (PermanentAccountProof) Bindings() PermanentAccountProofSpec
func (Assurance) AtLeast(Assurance) bool
func (ReauthenticationOrRecoveryProof) Validate(ReauthenticationUse, bool) error
func (RecoveryPath) Valid() bool
func (SignInState) Valid() bool
```

The last four helpers implement the already-required closed enum/union policy.
The existing PermanentAccountProofIssuer retains EXACTLY its single existing
IssuePermanentAccountProof(context.Context, PermanentAccountProofSpec)
(PermanentAccountProof,error) method. The existing factory's signature remains
PermanentAccountProofIssuerFromAuthenticator(Authenticator)
(PermanentAccountProofIssuer,error). No standalone permanent-proof Validate
method is introduced: expected-instance receiver validation is the trust oracle.

## Errors and ownership/zero-value semantics

No error delta versus the catalog: the four proof sentinels already listed are
implemented; no new shared-issuer error is added. NewAuthorityAuthenticator and
the issuer factory reject invalid/zero configurations with ErrInvalidConfiguration.
Token format/closed-schema failures classify ErrCredentialsInvalid; signature,
profile/lineage or current-validity rejection classifies ErrCredentialsRejected;
source outage/cancellation/clock rollback/late read classifies
ErrAuthenticationUnavailable. SigningBytes malformed input is ErrInvalidProof.
ResourcePrincipal.Validate uses ErrInvalidProof/ErrExpiredProof. Permanent-proof
validation/issuance preserves exact proof invalid/expired/purpose/audience errors;
bad signed original records/capability identity are ErrInvalidProof.

Config is constructible caller input, not authority evidence. Constructor requires
explicit EdDSA/shared-issuer-v1, required issuer/environment/audience, nonnil
concurrency-safe Source and existing go-clock Clock, 1..32 unique 32-byte keys
per set, 0..128 unique bounded required scopes, depth 1..4 and positive timeout
plus uncertainty totaling at most ten seconds. It copies retained values and
creates a private per-instance positive-size capability; it reads/authenticates
nothing. Zero/nil AuthorityAuthenticator is invalid. Applications own Source,
Clock, signing-key custody, source/cache lifetime and shutdown; no background
work or Close/resource owner is introduced. Reads derive bounded contexts.

Non-secret structured inputs are caller-owned mutable projections; zero values
are invalid at their consuming validation boundary. Required IDs are bounded,
versions positive and enum/dimension kinds closed. Root grant/credential zero
expiry is deliberately legitimate non-expiring source authority; children remain
at most five minutes and below every finite contributing expiry. Opaque
ResourcePrincipal/PermanentAccountProof zero is invalid; validated methods alone
produce them. Returned Bindings copies never confer independent authority;
opaque values are immutable/concurrency-safe and own no lifecycle resource.

## Selected resource boundary and trust invariants

AuthorityAuthenticator is **customer resource validation only**. Workload-only
owners are denied here, and tenant/current-validity tenant binding is required.
AuthorityOwnerWorkload remains a closed recognized vocabulary value; this denial
does not declare OAuth machine/client-credentials flows unsupported globally.
Machine-only resource policy belongs to its owning consumer/protocol boundary.
Owner (user/organization), optional user Actor, selected Tenant and ClientID
caller stay separate; Subject equals Owner.ID. Actor requires membership ID and
version. Exact required validity dimensions: root credential, root grant, each
parent grant, owner, tenant, client and signing key; actor/membership additionally
when actor exists. Exact IDs/versions and token/tenant/issuer/environment bind
every signed response. Each original read-start-based dimension expires within
20 seconds; earliest original expiry minus uncertainty wins. Copies cannot renew
freshness, lost events cannot extend it, stale/outage evidence fails closed.

PermanentAuthentication retrieves the ORIGINAL successful record by ProofID,
not requester bindings. State must be authenticated and Permanent must be true:
non-anonymous status is an authoritative signed property, not inferred from an
identifier spelling. All original Proof bindings/timestamps and permanent subject
must exactly match, with permanent-account-upgrade purpose and configured audience.
Signed account existence alone never mints. Only the package-owned configured
authority's issuer carries the private capability. The expected same configured
AuthorityAuthenticator.ValidatePermanentAccountProof rechecks capability identity
and complete current ProofUse; another composition with caller-selected keys is
not accepted. Bindings-only validation MUST NOT replace this receiver oracle.

## Deliberate decisions

No unresolved new contract decision remains. The bounded selected EdDSA profile,
source-owned lifecycle, existing clock composition, optional non-expiring roots,
customer-only resource boundary and expected-instance proof oracle follow prior
coordinator directions. Algorithm expansion/general JWT engines, machine-resource
authorization, provider deployment and signing-key lifecycle implementation are
outside this root primitive scope. Catalog/digest reconciliation remains solely
coordinator-owned; source is not modified to match stale draft rows.
