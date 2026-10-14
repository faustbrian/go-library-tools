# Proposed additive shared-issuer public contract

Owner: github.com/faustbrian/go-authentication root module only. Generation 0.
Declarations below are implemented in the current uncommitted checkpoint. Existing APIs remain unchanged.
All non-secret structs are caller-owned mutable projections, validated and
defensively copied before retention. Zero-valued configuration, token bindings,
records, resource principals and authority proofs are invalid. Source response
values are created by the explicit source result methods before callers supply
them; opaque principal/proof results are created only by package-owned methods.

```go
type AuthorityOwnerKind string // closed: unspecified, user, organization, workload
const AuthorityOwnerUnspecified AuthorityOwnerKind = ""
const AuthorityOwnerUser AuthorityOwnerKind = "user"
const AuthorityOwnerOrganization AuthorityOwnerKind = "organization"
const AuthorityOwnerWorkload AuthorityOwnerKind = "workload"
type AuthorityOwner struct { Kind AuthorityOwnerKind; ID string; Version uint64 }
type AuthorityGrant struct { ID string; Version uint64; ExpiresAt time.Time }
type AccessTokenClaims struct {
    Issuer string // iss
    Audience string // aud: this profile permits exactly one string audience
    Subject string // sub
    ClientID string // client_id; calling workload, never owner/actor
    IssuedAt int64 // iat, UTC Unix seconds
    ExpiresAt int64 // exp, exclusive UTC Unix seconds
    TokenID string // jti
    Purpose Purpose // purpose; exactly access-token, never ID-token
    Environment string
    Method string
    Assurance Assurance
    Owner AuthorityOwner
    Actor *AuthorityOwner // absent or verified user, never inferred
    Tenant string
    TenantVersion uint64
    ClientVersion uint64
    SigningVersion uint64
    MembershipID string // required with Actor; otherwise empty
    MembershipVersion uint64 // required with Actor; otherwise zero
    RootCredential AuthorityGrant
    RootGrant AuthorityGrant
    Parents []AuthorityGrant // root-to-immediate-parent ordered lineage
    Scopes []string
}
type ValidityDimensionKind string // closed values below; unknown/empty invalid
const ValidityRootCredential ValidityDimensionKind = "root-credential"
const ValidityRootGrant ValidityDimensionKind = "root-grant"
const ValidityParentGrant ValidityDimensionKind = "parent-grant"
const ValidityOwner ValidityDimensionKind = "owner"
const ValidityActor ValidityDimensionKind = "actor"
const ValidityTenant ValidityDimensionKind = "tenant"
const ValidityMembership ValidityDimensionKind = "membership"
const ValidityClient ValidityDimensionKind = "client"
const ValiditySigning ValidityDimensionKind = "signing"
type ValidityDimension struct { Kind ValidityDimensionKind; ID string; Version uint64 }
type DimensionValidity struct {
    Dimension ValidityDimension
    Active bool
    VerifiedAt time.Time // original authoritative read START, never response/cache time
    ValidUntil time.Time // exclusive expiry <= VerifiedAt + 20 seconds
}
type ValidityRequest struct { Claims AccessTokenClaims }
type SignedValidityEvidence struct {
    Issuer string
    Environment string
    TokenID string
    Tenant string
    Dimensions []DimensionValidity
    KeyID string
    Signature []byte
}
type SignedAuthenticationRecord struct {
    Issuer string
    Environment string
    State SignInState // exactly SignInStateAuthenticated
    Permanent bool // exactly true; anonymous authentication cannot mint
    Proof Proof // ORIGINAL proof produced by successful permanent authentication
    PermanentSubject string
    CredentialID string
    CredentialVersion uint64
    KeyID string
    Signature []byte
}
type AuthoritySource interface {
    CurrentValidity(context.Context, ValidityRequest) (SignedValidityEvidence, error)
    PermanentAuthentication(context.Context, ProofID) (SignedAuthenticationRecord, error)
}
type AuthorityKey struct { ID string; PublicKey ed25519.PublicKey }
type AuthorityConfig struct {
    Issuer string
    Environment string
    Audience string
    Algorithm string // explicitly exactly EdDSA
    Profile string // explicitly exactly shared-issuer-v1
    AccessKeys []AuthorityKey
    EvidenceKeys []AuthorityKey
    Source AuthoritySource
    Clock clock.Clock // existing go-clock v1.0.0 capability, not a new clock authority
    ReadTimeout time.Duration
    ClockUncertainty time.Duration
    MaximumDelegationDepth int
    RequiredScopes []string
}
type AuthorityAuthenticator struct { /* private immutable composition/minting capability */ }
func NewAuthorityAuthenticator(AuthorityConfig) (*AuthorityAuthenticator, error)
func (*AuthorityAuthenticator) Authenticate(context.Context, Credential) (Result, error)
func (*AuthorityAuthenticator) ValidateAccessToken(context.Context, string) (ResourcePrincipal, error)
type ResourcePrincipal struct { /* private copied typed bindings + absolute original deadline */ }
func (ResourcePrincipal) Bindings() AccessTokenClaims
func (ResourcePrincipal) ValidUntil() time.Time
func (ResourcePrincipal) Validate(time.Time) error
func (SignedValidityEvidence) SigningBytes() ([]byte, error)
func (SignedAuthenticationRecord) SigningBytes() ([]byte, error)
type PermanentAccountProof struct { /* private copied proof + permanent subject + private capability */ }
type PermanentAccountProofSpec struct { Proof Proof; PermanentSubject string }
type PermanentAccountProofIssuer interface {
    IssuePermanentAccountProof(context.Context, PermanentAccountProofSpec) (PermanentAccountProof, error)
}
func PermanentAccountProofIssuerFromAuthenticator(Authenticator) (PermanentAccountProofIssuer, error)
func (PermanentAccountProof) Bindings() PermanentAccountProofSpec
func (*AuthorityAuthenticator) ValidatePermanentAccountProof(PermanentAccountProof, ProofUse) error
```

## Trusted issuance, not identifier existence

The factory accepts only the concrete package-owned composition with its
private capability, rejects arbitrary Authenticator/issuer implementations
and performs no authentication. Issuance independently validates the supplied
Proof and exact permanent-account-upgrade purpose, asks the source ONLY for
its proof ID, and cryptographically verifies the signed ORIGINAL successful
permanent-authentication record. It requires exact complete Proof equality,
PermanentSubject equality and non-anonymous subject before issuing the private
result. A signed account-active/existence result for caller-created bindings
is neither the record contract nor sufficient authority. Source adapters must
retrieve the previously issued authentication record, not sign requester data.
Source callbacks cannot fabricate the record without an evidence signing key.
Public callers cannot manufacture the package capability or private result.
The expected configured AuthorityAuthenticator instance validates its own private
minting capability identity. Bindings is projection only, never a trust oracle;
another independently configured authority cannot validate that instance's proof.

## Local profile and hostile limits

The root composition accepts only bounded compact EdDSA access JWTs for the
explicit shared-issuer-v1 profile. It is not a general JWT/JWK engine and does
not import or edit the reserved jwt module. At most 32 keys per trust set,
16 KiB compact tokens, 16 validity dimensions, 128 scopes, four parents;
identifiers are at most 256 UTF-8 bytes and scope/key/version fields validated.
Limits are checked before decoding/allocation/crypto. Duplicate JSON fields,
unknown algorithm, ID-token purpose, wrong issuer/environment/audience,
missing roots, cycles, over-depth lineage and widened parent expiry deny.
MaximumDelegationDepth is explicitly 1..4, child lifetime at most five minutes,
and expiry cannot exceed any finite contributing lineage expiry. Root credentials
and grants may explicitly have zero expiry without imposing a new expiration
requirement; child access lifetime remains bounded. Workload-only
owners have no customer tenant authority. Owner, optional actor, tenant and
client remain distinct; the claim subject must equal the typed owner ID.

## Freshness and resources

Evidence signs exact claim identity and exact applicable dimension IDs/versions.
Every contributing dimension is active and valid; earliest original deadline
wins. Every deadline is <= original read start + 20 seconds; cache copying,
response delay, retries and promotion cannot replace timestamps. ClockUncertainty
and ReadTimeout are positive, total <=10 seconds; their explicit validated
budget preserves the selected <=30-second decision guarantee. There is no
background work, goroutine, or event requirement. The source owns any optional
cache and may return warm signed original evidence to avoid central per-request
lookups. A bounded context is derived for each source read. Expired/missing/
mismatched evidence and refresh outage deny; no stale-while-revalidate authority.
No caller callback runs under a lock. Config/key/scopes and retained results
are copied; sources and clocks must support concurrent calls. The application
owns source lifecycle, signing-key custody and independent resource authorization.

SigningBytes is a bounded, versioned, domain-separated canonical encoding of
the typed response excluding Signature. It permits source signing but never
confers minting authority or reveals a secret. Errors remain the pinned proof
and authentication sentinels and redact operational causes at the public edge.
No runtime provider access or deployment is requested.
