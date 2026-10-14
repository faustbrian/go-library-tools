# Goal: primitive/capability-identity-contracts

The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT",
"SHOULD", "SHOULD NOT", "RECOMMENDED", "NOT RECOMMENDED", "MAY", and
"OPTIONAL" in this document are to be interpreted as described in BCP 14
[RFC2119] [RFC8174] when, and only when, they appear in all capitals, as
shown here.

## Execution metadata

- Unit: `primitive/capability-identity-contracts`
- Canonical module: `pkg/capability`
- Canonical goal after scheduling: `pkg/capability/.ai/GOAL_IDENTITY_CONTRACTS.md`
- Requires: None
- Exact identity-platform consumers: `identity`, `identity/email`, `identity/magiclink`, `identity/oauth`, `identity/oauth/onetap`, `identity/oauth/postgres`, `identity/oauth/proxy`, `identity/otp`, `identity/password`, `identity/phone`, `identity/session`, `oauth-server`, `oauth-server/device`, `oauth-server/oidc`, `organization`, `sso`, `sso/postgres`, `sso/saml`, `webauthn`
- Unlocks after verification: `identity`, `identity/email`, `identity/magiclink`, `identity/oauth`, `identity/oauth/onetap`, `identity/oauth/postgres`, `identity/oauth/proxy`, `identity/otp`, `identity/password`, `identity/phone`, `identity/session`, `oauth-server`, `oauth-server/device`, `oauth-server/oidc`, `organization`, `primitive/capability-postgres-identity-contracts`, `sso`, `sso/postgres`, `sso/saml`, `webauthn`

## Start gate and pinned contract

The primary agent MUST follow IMPLEMENTATION_GOAL.md and COMMON_REQUIREMENTS.md.
Record the frozen consumer milestone and its necessary owners in the existing
ledger. Development requires committed, integrated, focused behavioral and
compatibility proof of the contracts actually consumed, not whole-unit release
verification or a generated worker assignment. Final unit verification retains
all product requirements, applicable release gates and verified prerequisites.

The current
`pkg/capability` API MUST match
`sha256:3ad716bb7799ba0d0045b7456d2dabfe486503b9050b4d984dead16d69da7e14`,
and its completed required contract MUST reproduce
`sha256:56dcf10ab5f1212e1b5a114d5c13ee3d24a933f1e7fc19c461c1919a6137d4fd`.
All digests MUST use the canonical algorithm in
`.ai/identity-platform/fragments/public_contracts_meta.json`; mismatch blocks
execution and MUST NOT be silently re-pinned.

## Objective and exact additions

Extend only the capability core with the identity-platform capability boundary:

- `Proof`: opaque secret value with unexported canonical capability-token bytes;
  bounds MUST use the configured `Limits.MaxTokenBytes`, including the existing
  16 KiB default, rather than introducing a smaller incompatible token limit.
- `Purpose`: defined string containing 1..128 canonical ASCII characters.
- `Reference`: `struct{CapabilityID string; Purpose Purpose; Digest [32]byte; ExpiresAt time.Time}`.
- `Service`: interface with exactly `Issue(context.Context, Payload) (Proof, Reference, error)` and `Verify(context.Context, Proof, Purpose, string) (Grant, error)`.
- `ServiceConfig`: `struct{Signer Signer; Resolver Resolver; Clock func() time.Time; Limits Limits; Skew time.Duration; Revocations RevocationChecker}`.
- `NewService(ServiceConfig) (Service, error)` constructs the explicit pinned
  policy holder; limits MUST be explicit, negative skew is invalid, and an
  absent revocation source requires an explicit application policy permitting it.

The core support contract MUST retain the actual pinned `Consumption`,
`ConsumptionResult`, `Payload`, and `RevocationQuery`
declarations and exact `ConsumptionStore`, `Resolver`, and `RevocationChecker`
interfaces. `Payload.Caveats` MUST remain `map[string]string`; no new
`CaveatEntry` wire representation is required or permitted by this extension.
It MUST expose `NewPurpose` and `NewReference` with the pinned
signatures while retaining the pinned grant, issue, parse, verify methods and
existing error set. Opaque proof ingress and transfer MUST expose exactly
`ParseProof(string, Limits) (Proof, error)` and
`Proof.Transfer(context.Context, func(context.Context, []byte) error) error`,
with the additive redacted `ErrProofTransferFailed` sentinel. PostgreSQL behavior belongs exclusively to the ordered
`primitive/capability-postgres-identity-contracts` unit.

Consumer admission additionally requires
`Proof.ValidateBounds(minimum, maximum int) error`: inclusive positive canonical
byte bounds, `ErrInvalidConfiguration` for invalid bounds and `ErrInvalidToken`
for zero, transferred or out-of-range proofs. It MUST inspect custody under
the existing synchronization owner without copying, disclosing or consuming
token bytes. Success is an observation, not authentication or future custody.
This additive method reconciles the narrower email-confirmation admission
contract without changing configured parser limits. Historical API digests
above do not include this addition; reconcile the consumed source baseline
before final unit verification rather than treating them as current evidence.

Native command intent additionally requires the additive
`Proof.CommandDigest(key [32]byte, fieldPath string, maximumBytes int) ([32]byte, error)`.
It MUST bind exact original canonical token bytes, including signature, using
the existing identity-command-secret-v1 HMAC-SHA-256 framing without transfer,
authentication, clocks, consumption or a cached result. Key MUST be nonzero;
path MUST be exact valid UTF-8 of 1..4096 bytes; maximum MUST be positive and
consumer-selected without narrowing configured parser profiles. Invalid
configuration returns zero/ErrInvalidConfiguration; zero, transferred or
oversized custody returns zero/ErrInvalidToken. Shared lifetime remains:
a synchronized snapshot acquired before Transfer MAY finish, later calls deny,
and every private snapshot MUST be cleared. The digest is sensitive intent
correlation, never authority. Historical API digests omit this consumed addition
and MUST be reconciled before final verification, not silently asserted current.

## Behavioral and security requirements

- `Proof` MUST remain secret and MUST NOT implement string, text, JSON,
  logging, or formatting disclosure.
- `ParseProof` MUST use existing `Parse` with configured limits, reject lengths
  before copying/decoding, and return a zero proof on failure. Structural
  validity MUST NOT be treated as cryptographic authority or current validity.
- Copies MUST share one private secret lifetime. `Transfer` MUST invalidate all
  aliases before invoking one synchronous callback outside locks; zero or
  transferred proofs fail with `ErrInvalidToken`, and nil context or callback
  fails with `ErrInvalidConfiguration`. Pre-cancellation MUST prevent callback
  invocation and preserve ownership. Detached borrowed bytes MUST be cleared
  after success, callback error, or panic. Errors/panics MUST become redacted
  `ErrProofTransferFailed`, retaining only safe cancellation classification.
- Transfer MUST NOT consume a grant or imply successful transport. Failed
  transfers MUST NOT retry automatically. A verification snapshot acquired
  before transfer MAY finish; later verification MUST reject the invalidated
  proof. Transport-owned explicit copies remain the caller's secret-custody
  responsibility; the package MUST NOT promise prevention of malicious copying
  or runtime-wide erasure. Non-disclosing formatting/logging hooks MAY be used
  to prevent reflection-based disclosure of private byte storage.
- `Purpose` MUST name one exact operation and MUST reject prefix, wildcard,
  hierarchy, and cross-purpose substitution.
- `Reference` MUST require a bounded capability ID, exact purpose,
  domain-separated SHA-256 digest, and expiry; it MUST grant no authority alone.
  Its exported metadata is mutable and reconstructed values MUST be validated.
  `NewReference` validates the nonzero caller-supplied digest but cannot prove
  its preimage or capability existence; `Service.Issue` derives the actual
  domain-separated token digest.
- `Service.Issue` MUST use pinned signer, limits, and clock configuration.
- `Service.Verify` MUST bind exact purpose and audience and MUST check signature,
  time, and revocation before returning a grant. It MUST NOT automatically
  consume the grant or imply that a protected operation has committed.
- The protected operation's transaction owner MUST atomically consume or
  reserve/finalize a one-time grant with its protected domain transaction;
  verification alone MUST NOT authorize a committed one-time effect. Preserve
  existing explicit `Grant.Consume` and `ConsumptionStore` behavior. The
  PostgreSQL adapter owns transaction participation in its separate unit.
- Every verification, revocation, or consumption error MUST fail closed.
- Capability audiences and caveats MUST retain their bounded canonical wire
  representation as defined by `Limits`, including bytewise-sorted audiences
  and caveat keys, unique audiences, and the existing caveat map semantics.
- Capability timestamps MUST preserve the pinned UTC-second ordering and
  maximum-lifetime contract.
- Public errors MUST remain stable, typed, redacted, and usable for
  reconciliation without exposing proof material.
- Documentation MUST state secret ownership, zero values, concurrency, expiry,
  replay, revocation, consumption, and unknown-outcome semantics.

## Compatibility, migration, and evidence

The change MUST be additive to the pinned current API. Existing issue, parse,
verify, grant, revocation, and consumption behavior MUST remain compatible.
Consumers MUST migrate from local proof, purpose, reference, and service
facades to these canonical contracts. Stored capability data MUST NOT be
reinterpreted; representation changes require an explicit versioned migration
with rollback and mixed-version tests.

Focused tests MUST prove construction bounds, exact purpose/audience binding,
secret non-disclosure, expiry, revocation, replay, consumption, and concurrent
use. Applicable fuzz and hostile-input tests MUST exercise token parsing. The
worker MUST run exact coverage, mutation, race, fuzz, leak, API-baseline,
clean-consumer, documentation, example, inventory, supply-chain, and changed
reverse-dependant gates for the core package. `pkg/capability/CHANGELOG.md` MUST
describe the additive contracts and migration impact. This unit MUST remain
unverified until its required digest and every applicable gate pass.

## Scope boundary

Changes MUST be limited to the core extension, tests, API baseline,
documentation, examples, changelog, and mechanically required manifests. The
worker MUST NOT modify PostgreSQL or another adapter, redesign token formats or
cryptography, or add identity-specific authority.
