# Capability timestamp compatibility reconciliation

The authorized extension preserves the existing capability API and stored wire
behavior. Fresh independent mapping and QA found that the derived Payload
semantic row incorrectly required IssuedAt <= NotBefore. The committed goal
instead explicitly preserves the pinned timestamp contract, and the user
approved keeping capability's existing API unchanged.

At the pinned package revision b0b21d15c9653decb0ff4a4fbffbe71c77d8d8fc,
payload.go validates non-negative timestamps, expiry after both issuance and
not-before, and maximum lifetime from their minimum. docs/protocol.md explicitly
permits not-before preceding issuance for planned skew; the ordinary successful
fixture exercises that ordering. This reconciliation changes no package code,
wire format, permission, lifetime limit, or extension scope.

Only the stale metadata sentence and its reproducibly derived required-contract
digest change. The current API digest remains unchanged. The goal body changes
only its derived required digest; manifest and generated catalog follow that
input. The focused coordinator regression first failed on the old semantic row
and then passes on the preserved timestamp contract. This is metadata proof,
not package runtime certification or a gate waiver.
