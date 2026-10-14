# Goal: primitive/identifier-identity-contracts

The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT",
"SHOULD", "SHOULD NOT", "RECOMMENDED", "NOT RECOMMENDED", "MAY", and
"OPTIONAL" in this document are to be interpreted as described in BCP 14
[RFC2119] [RFC8174] when, and only when, they appear in all capitals, as
shown here.

## Execution metadata

- Unit: `primitive/identifier-identity-contracts`
- Canonical module: `pkg/identifier`
- Canonical goal after scheduling: `pkg/identifier/.ai/GOAL_IDENTITY_CONTRACTS.md`
- Requires: None
- Exact identity-platform consumers: `identity/email`, `identity/magiclink`, `identity/otp`, `identity/password`, `identity/risk/captcha`, `identity/username`, `passkey`, `webauthn`
- Unlocks after verification: `identity/email`, `identity/magiclink`, `identity/otp`, `identity/password`, `identity/username`, `passkey`, `webauthn`

## Start gate and pinned contract

The primary agent MUST follow IMPLEMENTATION_GOAL.md and COMMON_REQUIREMENTS.md.
Record the frozen consumer milestone and its necessary owners in the existing
ledger. Development requires committed, integrated, focused behavioral and
compatibility proof of the contracts actually consumed, not whole-unit release
verification or a generated worker assignment. Final unit verification retains
all product requirements, applicable release gates and verified prerequisites.

The current API
MUST match
`sha256:99bd12b8c660306f4d4ab99ffcc9d4f8df171e09f01c62cfc65fde004bcb536c`.
The completed required contract MUST reproduce
`sha256:29aa06d6671d8a64a2baed5457767877e6a0a07e7105f8d5c61d447f3614a614`
under the canonical digest algorithm in
`.ai/identity-platform/fragments/public_contracts_meta.json`. A current-API mismatch MUST block the
goal for reconciliation and MUST NOT be silently re-pinned.

Baseline reconciliation: the recorded source revision
`23a8818faef585a87e3e31965d6af51bd7900478` omits `api/baseline.txt` present
at historical `c42902d6478497518cb48406bdd8a4122912f497`. The committed
`identifier.go` and `go.mod` are byte-identical; no public symbol or runtime
behavior changed. The explicit start pin now binds the recorded revision's
actual digest inputs instead of an absent baseline artifact.

## Objective and exact additions

Extend the existing identifier primitive with the exact bounded input,
normalization, canonical, and reference vocabulary required by identity
consumers. The module MUST add:

- `Canonical`: opaque immutable value containing 1..4096 valid UTF-8 bytes plus
  the exact normalization profile ID and positive profile version, returned
  only by `Normalize`.
- `Input`: defined string containing 1..4096 valid UTF-8 bytes.
- `NormalizationProfileID`: defined string containing 1..128 canonical ASCII
  characters.
- `NormalizationProfileSpec`: `struct{ID NormalizationProfileID; Version uint32; Unicode UnicodeNormalization; CaseFold bool; TrimSpace bool; IDNA bool; MaximumBytes uint32}`.
- `NormalizationProfile`: opaque immutable value constructed only from
  `NormalizationProfileSpec`.
- `CanonicalRecord`: `struct{Value string; ProfileID NormalizationProfileID; ProfileVersion uint32}` used only as the complete versioned persistence/wire projection.
- `Reference`: `struct{Kind string; Canonical Canonical; Display string}`.
- `UnicodeNormalization`: `enum{UnicodeNormalizationUnspecified,UnicodeNormalizationNone,UnicodeNormalizationNFC,UnicodeNormalizationNFKC}`.
- `NewInput(string) (Input, error)`.
- `NewNormalizationProfile(NormalizationProfileSpec) (NormalizationProfile, error)`.
- `NewReference(string, Canonical, string) (Reference, error)`.
- `Normalize(Input, NormalizationProfile) (Canonical, error)`.
- `RestoreCanonical(CanonicalRecord, NormalizationProfile) (Canonical, error)`.
- `Canonical.Record() CanonicalRecord`.
- `Canonical.Equal(Canonical) bool`.

The module MUST expose `ErrInvalid`, `ErrNormalization`, and `ErrUnsupported`
with the exact pinned semantics. It MUST NOT add a generic identifier payload,
an untyped normalization option map, a hidden global profile, authentication
meaning, authorization meaning, or identity-record ownership.

## Behavioral and security requirements

- `NewInput` MUST validate UTF-8 and byte length without normalizing before a
  profile is selected.
- Input values MUST be treated as enumeration-sensitive and MUST NOT appear in
  logs, traces, metrics, errors, or fixtures.
- `NormalizationProfile` MUST explicitly select Unicode behavior and MUST
  require a stable non-zero ID, a positive version, and `MaximumBytes` within
  1..4096. Its fields MUST be unexported and immutable after construction.
- `UnicodeNormalizationUnspecified` MUST be invalid.
- `Normalize` MUST be deterministic for one immutable profile and MUST reject
  unsupported or over-limit results.
- A profile change MUST require a migration and MUST NOT silently reinterpret
  an existing canonical value.
- `Canonical` MUST retain its exact profile ID and version in its opaque value;
  callers MUST NOT construct it from a string or mutate its canonical bytes or
  profile binding. Equality MUST compare profile ID, profile version, and
  canonical bytes and MUST NOT compare canonical bytes across different
  profiles or present them as a display value. Any controlled text/persistence
  projection MUST preserve the profile ID and version alongside the bytes and
  MUST reject an absent, unknown, or mismatched profile on restore.
  `RestoreCanonical` MUST require the exact matching immutable profile and MUST
  re-normalize `Value` to prove it is a fixed point before returning a value;
  it MUST NOT accept caller-declared canonical bytes on profile identity alone.
- `Reference.Kind` MUST contain 1..64 canonical ASCII bytes.
- `Reference.Canonical` MUST be non-zero; `Display` MUST be optional valid
  UTF-8 bounded to 4096 bytes.
- `Reference` MUST carry identifier identity only and MUST NOT constitute
  authentication or authorization evidence.
- Constructors and normalization MUST copy or immutably own retained values.
- Errors MUST be stable, typed, bounded, redacted, and distinguish invalid,
  normalization-failed, and unsupported cases.
- Public documentation MUST define profile identity, zero values, ordering,
  ownership, concurrency, privacy, migration, and error semantics.

## Compatibility, migration, and evidence

The extension MUST be additive to the pinned current identifier API. Existing
identifier kinds and normalization behavior MUST remain source- and
behavior-compatible. Consumers MUST migrate from local input, canonical,
profile, and reference types to the canonical additions. Stored canonical
identifiers MUST retain their profile/version identity; a profile or algorithm
change MUST use a versioned, collision-aware, rollback-capable migration with
mixed-version compatibility evidence.

Focused tests MUST prove UTF-8 and byte bounds, deterministic normalization,
every Unicode mode, case/space/IDNA flags, profile immutability, compile-time
non-constructibility, exact profile-ID/version equality and persistence,
unknown-profile rejection, collision handling, reference bounds, redaction,
defensive ownership, and concurrent use.
Applicable fuzz tests MUST exercise hostile Unicode, IDNA, length, and invalid
encoding inputs. The worker MUST run exact coverage, mutation, race, fuzz,
API-baseline, clean-consumer, documentation, example, inventory, supply-chain,
and changed reverse-dependant gates. `pkg/identifier/CHANGELOG.md` MUST describe
the additive API and migration impact. Every exported addition MUST have useful
Go documentation. This unit MUST remain unverified until the required digest
and every applicable gate pass; skipped, stale, warning-only, or missing
evidence MUST NOT count as a pass.

## Scope boundary

The worker MUST NOT redesign identifier kinds, identity records, authentication,
authorization, delivery, storage, or protocol packages. Changes MUST be
limited to the exact extension, its tests, API baseline, documentation,
examples, changelog, and mechanically required manifests.
