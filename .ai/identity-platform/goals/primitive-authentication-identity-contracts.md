# Goal: primitive/authentication-identity-contracts

The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT",
"SHOULD", "SHOULD NOT", "RECOMMENDED", "NOT RECOMMENDED", "MAY", and
"OPTIONAL" in this document are to be interpreted as described in BCP 14
[RFC2119] [RFC8174] when, and only when, they appear in all capitals, as
shown here.

## Shared-issuer amendment: Shared-issuer resource verification

[`SHARED_ISSUER_PROFILE.md`](../SHARED_ISSUER_PROFILE.md) is binding for this
unit. Supply typed principal and current-validity adapter contracts for SI-03
through SI-06. Validate JWT signature/purpose/issuer/environment/audience
locally and bind exact lineage, owner, actor, tenant and caller. Cached
original evidence may avoid per-request lookup but MUST NOT extend the
twenty-second source-age bound; stale/unavailable evidence fails closed.

## Execution metadata

- Unit: `primitive/authentication-identity-contracts`
- Canonical module: `pkg/authentication`
- Canonical goal after scheduling: `pkg/authentication/.ai/GOAL_IDENTITY_CONTRACTS.md`
- Requires: None
- Exact identity-platform consumers: `identity/anonymous`, `identity/email`, `identity/magiclink`, `identity/mfa`, `identity/otp`, `identity/password`, `identity/phone`, `identity/username`, `passkey`, `webauthn`
- Unlocks after verification: `identity/anonymous`, `identity/email`, `identity/magiclink`, `identity/mfa`, `identity/otp`, `identity/password`, `identity/phone`, `identity/username`, `passkey`, `webauthn`

## Start gate and pinned contract

The primary agent MUST follow IMPLEMENTATION_GOAL.md and COMMON_REQUIREMENTS.md.
Record the frozen consumer milestone and its necessary owners in the existing
ledger. Development requires committed, integrated, focused behavioral and
compatibility proof of the contracts actually consumed, not whole-unit release
verification or a generated worker assignment. Final unit verification retains
all product requirements, applicable release gates and verified prerequisites.

The primary agent MUST confirm that the current `pkg/authentication` API still has digest
`sha256:57ba45a7bb9441a3d936c5073e3e59460d214e68d8aeb3ded8f32bc329b385d0`.
The completed public contract MUST have required-contract digest
`sha256:7530318a6e41279264a7098d6d13adf6a10c0fa6c561d61cb8893698b6f50e53`
under the canonical digest algorithm in
`.ai/identity-platform/fragments/public_contracts_meta.json`. A current-API digest mismatch MUST
stop execution for a fresh contract reconciliation; it MUST NOT be silently
re-pinned.

## Objective and exact additions

Extend the existing authentication primitive only enough to supply the typed,
bounded proof vocabulary required by identity-platform consumers. The module
MUST add these exact public symbols and declarations:

- `Assurance`: `enum{AssuranceUnspecified,AssuranceSingleFactor,AssuranceMultiFactor,AssurancePhishingResistant}`.
- `OneTimeCode`: opaque immutable value containing 4..128 ASCII symbols in
  unexported storage and constructible only through `NewOneTimeCode`.
- `PermanentAccountProof`: opaque immutable authority-issued value with
  unexported proof and permanent-subject bindings.
- `PermanentAccountProofSpec`: `struct{Proof Proof; PermanentSubject string}`
  accepted only by the authority-gated issuer boundary.
- `PermanentAccountProofIssuer`: interface with exactly
  `IssuePermanentAccountProof(context.Context, PermanentAccountProofSpec) (PermanentAccountProof, error)`.
- `Proof`: `struct{ID ProofID; Tenant string; Subject string; Method string; Assurance Assurance; Purpose Purpose; Audience string; IssuedAt time.Time; ExpiresAt time.Time}`.
- `ProofID`: defined string containing 22..128 canonical ASCII characters.
- `Purpose`: defined string containing 1..128 canonical ASCII characters.
- `ReauthenticationProof`: `struct{Proof Proof; SessionID string; SessionVersion uint64; Action string}`.
- `ReauthenticationOrRecoveryProof`: the closed tagged union selected by `ProofKind`, with exactly `Reauthentication ReauthenticationProof` and `Recovery Proof` arms.
- `RecoveryCode`: opaque immutable value containing 8..256 canonical ASCII
  symbols in unexported storage and constructible only through
  `NewRecoveryCode`.
- `RecoveryPath`: `enum{RecoveryPathUnspecified,RecoveryPathEmail,RecoveryPathPhone,RecoveryPathPasskey,RecoveryPathAdministrator}`.
- `SignInState`: `enum{SignInStateUnspecified,SignInStateAuthenticated,SignInStateChallengeRequired,SignInStateRejected}`.

The support contract MUST include `ProofKind`, `ProofSpec`, `ProofUse`, and
`ReauthenticationUse` exactly as declared by the pinned required contract. It
MUST expose `NewOneTimeCode`, `NewProof`, `NewPurpose`, `NewRecoveryCode`,
`Proof.Validate`, and `ReauthenticationProof.Validate` with the exact
signatures in `.ai/identity-platform/fragments/public_contracts_meta.json`. It MUST expose
`ErrExpiredProof`, `ErrInvalidProof`, `ErrProofAudienceMismatch`, and
`ErrProofPurposeMismatch` in addition to retaining the pinned authentication
errors. It MUST NOT introduce an opaque payload, generic claims map, alternate
proof authority, or identity-specific service facade.

## Behavioral and security requirements

- Successful assurance MUST reject `AssuranceUnspecified`; ordering MUST be
  compared through a package helper rather than numeric conversion.
- `OneTimeCode` and `RecoveryCode` MUST NOT implement string, text, JSON,
  logging, or formatting disclosure. Their concrete representation and secret
  bytes MUST be inaccessible outside the package; no accessor, conversion,
  marshal method, exported field, or error may reveal them. Callers MUST be
  able to release the opaque value immediately after terminal verification.
- Opaque codes MUST expose `KeyedDigest([32]byte, []byte) ([32]byte, error)`
  for verification without a raw-code accessor. The key MUST be a trusted,
  independently generated versioned digest key; zero key/code and bindings
  outside 1..4096 bytes MUST fail with zero digest and ErrInvalidProof. Use
  HMAC-SHA-256 with the exact type-specific `authentication-one-time-code-digest-v1`
  or `authentication-recovery-code-digest-v1` domain, NUL, big-endian uint32
  binding length, exact binding, big-endian uint16 code length and exact code.
  Workflow-owned bindings MUST include complete tenant/purpose/subject/channel/
  challenge/key-version scope injectively. No input is retained and temporary
  mutable code copies MUST be cleared; immutable Go secret erasure is not promised.
  The digest is sensitive correlation data, not bearer or proof. Comparison,
  attempts, expiry, rotation and atomic consumption remain consumer-owned.
- `OneTimeCode.CommandDigest([32]byte, string) ([32]byte, error)` MUST supply
  the existing TRANSACTION_CONTRACT.md command-secret profile without a raw
  code accessor: ASCII identity-command-secret-v1 (no NUL), big-endian uint32
  UTF-8 field-path byte length, exact path, big-endian uint64 secret byte length,
  exact code. Nonzero trusted fingerprint key, nonzero code and 1..4096-byte
  valid UTF-8 path are mandatory; invalid input returns zero and ErrInvalidProof.
  No normalization or input retention is permitted. The coordinator owns schema
  bounds, outer deterministic encoding, key version/retention and durable replay;
  this primitive MUST NOT authenticate, consume or grant transaction authority.
  Existing KeyedDigest framing MUST remain unchanged.
- `Proof` construction MUST require every identifier, bound each to at most
  256 UTF-8 bytes, require `IssuedAt < ExpiresAt`, and cap lifetime at 24 hours.
- Proof validation MUST bind the exact tenant, subject, purpose, and audience
  and MUST fail closed on expiry or mismatch.
- `PermanentAccountProof` MUST bind a non-anonymous permanent subject, tenant,
  audience, purpose, proof identity, issuance, and expiry. Only an
  authentication-authority implementation that has authoritatively validated
  the underlying permanent account may mint it. The package MUST NOT expose a
  struct literal or public constructor that accepts caller-supplied subject or
  identifier data as sufficient authority, and identifier existence alone MUST
  NOT produce the proof. Consumers MAY inspect only bounded non-secret binding
  projections needed for exact validation; they MUST NOT recover or replace
  the authority evidence.
- `PermanentAccountProofIssuer` MUST be constructed only by the package's
  trusted authentication-authority composition with an unexported minting
  capability. Its method MUST independently validate the supplied `Proof`,
  require the exact permanent-account-upgrade purpose, require subject equality
  with `PermanentSubject`, reject anonymous subjects, and copy all bindings.
  Merely implementing the public interface MUST NOT grant minting authority;
  consumers receive the issuer instance by explicit trusted composition and
  cannot instantiate an equivalent issuer from public values.
- The exact factory MUST be
  `PermanentAccountProofIssuerFromAuthenticator(Authenticator) (PermanentAccountProofIssuer, error)`.
  It MUST return an issuer only when the supplied authenticator is a
  package-owned trusted authority carrying the package's unexported minting
  capability; arbitrary external implementations of `Authenticator`,
  `PermanentAccountProofIssuer`, or both MUST be rejected with
  `ErrInvalidConfiguration`. The factory MUST retain no caller credential and
  MUST NOT authenticate as a side effect.
- `ReauthenticationProof` MUST require a positive session version and exact
  session, action, purpose, subject, tenant, and audience matches.
- The recovery union MUST populate exactly the arm selected by `ProofKind` and
  MUST reject recovery where the consuming operation does not explicitly allow it.
- Constructors MUST reject invalid zero values and MUST copy every retained
  caller-owned value.
- Errors MUST be stable, typed, redacted, and distinguish invalid, expired,
  audience-mismatched, and purpose-mismatched proofs.
- Public documentation MUST define zero values, lifetime, ownership,
  concurrency safety, cleanup, and redaction for every addition.

## Compatibility, migration, and evidence

### Shared-issuer additive contract reconciliation

The complete required contract includes the exact shared-issuer declarations
in `fragments/public_contracts_meta.json`: `AccessTokenClaims`,
`AuthorityAuthenticator`, `AuthorityConfig`, `AuthorityGrant`, `AuthorityKey`,
`AuthorityOwner`, `AuthorityOwnerKind`, `DimensionValidity`,
`ResourcePrincipal`, `SignedAuthenticationRecord`, `SignedValidityEvidence`,
`ValidityDimension`, `ValidityDimensionKind`, `ValidityRequest`, and the
two-method `AuthoritySource`. Public construction and validation methods,
defensive `Bindings`, original `ValidUntil`, signed evidence `SigningBytes`
and closed-enum/union helpers MUST match that catalog. The historical
current-API start pin remains unchanged; this required-contract digest extends
the original proof vocabulary with the already-authorized shared-issuer scope.

Customer-resource validation MUST keep owner, optional human actor, selected
tenant and client/caller distinct, reject workload-only customer authority,
preserve exact root/parent versions and fail closed on stale signed evidence.
Every original validity interval is at most twenty seconds from source-read
start, and read-timeout plus clock uncertainty totals at most ten seconds.
Sources and caches MUST preserve original timestamps. The expected configured
authenticator instance, not a returned binding projection or another
caller-configured authority, validates its permanent-proof minting capability.

This primitive's signed-fixture and source-contract checks contribute to SI-03
through SI-06; they MUST NOT be reported as composed OAuth issuance,
PostgreSQL lineage, real cache-adapter or full thirty-second revocation proof.
Those affected package and end-state acceptance obligations remain unchanged.

The change MUST be additive for the pinned current authentication API and MUST
NOT alter existing `Principal`, `Authenticator`, credential, result, or error
semantics. Existing consumers MUST continue to compile unchanged. Consumers
MUST migrate only from locally duplicated proof/code types to these canonical
types; no wire or persisted representation may be rewritten without an
explicit versioned migration and compatibility test.

Focused tests MUST prove constructor bounds, closed enums/unions, exact binding,
expiry, mismatch errors, compile-time non-constructibility of opaque values,
secret non-disclosure across conversion/marshal/format/error surfaces,
authority-only permanent-account-proof issuance, defensive copying, and
race-safe concurrent validation. Applicable fuzz tests MUST cover hostile code
and proof inputs. The worker MUST run the module's exact coverage, mutation, race, fuzz,
API-baseline, clean-consumer, documentation, example, inventory, supply-chain,
and changed reverse-dependant gates. `pkg/authentication/CHANGELOG.md` MUST
describe the additive public contract and migration impact. All exported
symbols MUST have complete Go documentation. This unit MUST remain unverified
until the required digest is reproduced and every applicable gate passes; a
skipped, stale, warning-only, or missing result MUST NOT count as verification.

## Scope boundary

The worker MUST NOT redesign authentication flows, identity ownership,
authorization, storage, sessions, delivery, or protocol packages. Changes MUST
be limited to this exact primitive extension, its tests, API baseline,
documentation, examples, changelog, and mechanically required manifests.
