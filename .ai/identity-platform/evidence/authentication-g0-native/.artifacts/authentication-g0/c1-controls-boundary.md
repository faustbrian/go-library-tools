# Unicode no-control repair

Fresh regression g1 identified a documented-contract defect, not a signature
bypass: valid UTF-8 C1 U+0080..U+009F passed identifier's C0+DEL-only predicate.
Public NewProof red independently exercises every C1 code point in tenant,
subject, method and audience; DEL is already denied. Printable accented/CJK
Unicode and U+00A0 remain admitted unchanged. Invalid cases require the exact
ErrInvalidProof classification and zero returned proof. Native outputs:
c1-controls-red.log (expected exit 1), c1-controls-green.log (exit 0), pinned
Go1.26.6, GOWORK=off and disposable build/module/temp caches cleaned on exit.

Shared identifier consumers: Proof.valid/NewProof and Proof.Validate use;
AuthorityConfig issuer/environment/audience; access claims subject/client/tenant/
method, owner/actor, membership and root/parent grants; signed validity envelope
and every dimension identity; signed original authentication record including
permanent subject/credential; permanent-proof issuance permanent subject.
No existing legacy credential/principal API uses this new internal helper.

Minimal production repair uses standard unicode.IsControl, rejecting Unicode
Cc (C0, DEL, C1) while retaining original valid UTF-8 and 1..256 byte bounds.
No normalization, identifier rewrite, new error, exported API, algorithm or
authority policy is introduced. Existing fuzz target now independently checks
explicit numeric control ranges against public construction, with C1 and
printable Unicode seeds. Prior gate results remain historical after this edit;
current affected-input gates and new full independent reviews are required.
