# Selected restricted exchange: protocol catalog reconciliation

The user-approved shared-issuer profile requires restricted RFC 8693 exchange
as existing goal scope, distinct from the private session-JWT extension.
The old source pin incorrectly called all RFC 8693 exchange unsupported.

`PROTOCOL_CONFORMANCE_MANIFEST.json` now records a profile decision against
the already pinned RFC source. `PROTOCOL_BASELINES.md` distinguishes the
selected standard grant, local lineage/identity/non-expansion/freshness rules,
and the separate private extension. Source checksum, other pins, consumers,
assignments, execution state and the 67-unit scope are unchanged.

The existing acceptance generator changed only the derived supply-chain
source-profile closure from unsupported to passed-required evidence. This
schema value is a required future result, not a runtime pass. The current
acceptance catalog check passed: 67 program units, 61 product units,
329 operations and 70 artifacts. No new receipt owner or gate was added.

The ten focused SI-01 through SI-10 runtime cases remain authoritative in
`SHARED_ISSUER_PROFILE.md`. Their detailed typed/protocol/runtime acceptance
reconciliation belongs to the affected implementation batches and cannot be
substituted by this source-profile correction. No OAuth implementation,
exchange interoperability or shared-issuer milestone is certified here.
