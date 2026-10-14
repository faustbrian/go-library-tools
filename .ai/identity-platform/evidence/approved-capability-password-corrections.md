# Approved capability and password corrections

On 2026-10-01 the user answered "Apply those corrections (recommended)" to:

> Two goals contain actual contradictions. May I resolve them by keeping
> capability's existing API unchanged, making its new Proof opaque, keeping
> one-time use atomic with the protected transaction rather than automatically
> consuming in Verify, and validating password parameters in two stages
> (intrinsic bounds, then the existing policy limits)? These preserve
> compatibility and security; they change the contradictory wording, not the
> intended feature scope.

This approval applies to the existing capability caveat map/wire contract,
opaque proof bounds using configured token limits, explicit atomic one-time
transaction ownership, and intrinsic parameter versus caller policy validation.
Existing primitive API pins, package ownership, feature scope, and applicable
verification gates remain unchanged. No package implementation is certified.

Independent final review identified required opaque ingress/egress reachability
under COMMON_REQUIREMENTS.md. The read-only package-owner proposal from
`/root/capability_opaque_contract_proposal` supplies `ParseProof` and a controlled
one-shot `Transfer` with redacted `ErrProofTransferFailed`. These are callable
support for the approved opacity and existing one-time-transfer lifetime, not
a new authentication, replay, transport-adapter, or product authority.
