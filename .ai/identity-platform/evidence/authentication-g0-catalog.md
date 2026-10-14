# Authentication: shared-issuer typed contract reconciliation

This coordinator-only reconciliation implements the user-authorized
`shared-issuer-v1` amendment. It does not change package code, assignments,
execution state or the historical authentication current-API start pin.

The integrated source authority is `github.com/faustbrian/go-authentication`
at `a5eb0e543f53610ceaece35373986cf0a4fa5891`, tree
`fe4949e07e5e62cbed73458111e456719f1532e4`. Its canonical public-source
identity is `sha256:4b2e9ac6983a72251f4398be99b8eb2c952cd00cbc57a3119ab2340284745ac2`.
The independently reviewed complete source manifest is retained with the
native worker evidence; source declarations and public-contract delta were
inspected directly, not inferred from a green gate.

The authentication external contract retains every original type, interface,
method and error. It adds fourteen shared-issuer types, the two-method
AuthoritySource and fourteen public constructor/validation/projection/helpers.
The complete canonical required-contract digest is
`sha256:15c5ac8392d5b7abc241b0b09fc069786f648c3a6d5c3daa966e4c97443e2e38`.
The original subset digest `f5ed6241...` is preserved by prior committed
evidence; it is not sufficient to certify the amended contract. Goal pins
were updated from the complete bytes without changing the start API baseline.

Generated public contracts now select the integrated authentication source.
Their password source-lock refresh is mechanical: it selects the already
verified integrated password API `f8d503b8...`, not a new password contract.
No other external owner or product operation/type contract was changed.

These typed declarations describe customer-resource verification only, not
an authorization-server engine, machine-resource policy or composed cache
proof. The primitive's original signed-evidence deadlines, explicit source
custody, bounded read/clock budget and expected-instance permanent-proof
oracle remain mandatory. No bindings-only authority or cache freshness reset
is introduced.

Public-contract generation/check and all 67 goal-body pins were checked.
Protocol and composed acceptance catalog reconciliation remains outstanding,
as do coordinator integrated and affected reverse-consumer gates. Inventory
therefore remains implemented-unverified; this document certifies no unit,
runtime milestone, publication, hosted CI or full-program completion.
