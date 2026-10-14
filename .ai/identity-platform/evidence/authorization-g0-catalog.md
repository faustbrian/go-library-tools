# Authorization: approved-main administrative contract reconciliation

The user approved adopting committed main
`ce41a61000844dd3c06484a7799af336ad1b2892` while preserving historical
consumer compatibility. Its current API baseline is
`73b296651509c8f8314a42c696c6adddc8548f1166e46e09bf895395683b2019`.
This coordinator-only change preserves assignments and execution state.

The authoritative goal already requires canonical authorization Actor and
Target subjects, seven read-error variants, nine mutation-error variants,
opaque operation keys and no dependency on future identity packages.
The metadata and the two external administrative operation projections now
match those requirements and the independently reviewed source. Constructors,
validators, defensive-copy helpers and operation-specific classifiers are
explicit. Deny carries no mutation projection/version; Allow requires the exact
target and a strictly newer committed version. Empty role/statement collections
may clear assignments. The consuming authority still owns scope, persistence,
atomic replay, audit and reconciliation; primitive validation proves none of
those durable behaviors. Permission-read HTTP error projections cannot report
conflict or unknown commit.

The complete canonical required-contract digest is
`sha256:33f0787b321be7f21ced591008930a457b86e84c3dc6fd6f99c2e3678cb3d0ba`.
The goal-byte pin is
`5174b0efa932aceb33b2535d46eb90de037a39258a52365e49bd3cc39a556c86`.
The former required digest described superseded declarations and is not a
completion authority for the corrected contract.

Native read-only principals `/root/authorization_regression_g1`,
`/root/authorization_correctness_g0` and `/root/authorization_go_g0`
each returned APPROVE for the frozen 23-path package-local source closure
`780b6ed40addf4bba78cb796c073ef087f8df86c6f56ee9b317535d864557405`
at base `ce41a61000844dd3c06484a7799af336ad1b2892`. Their complete native
reports remain in the coordinator conversation; this summary does not replace
them or extend their dispositions to integrated-unit readiness. They observed
the stale coordinator metadata independently. NilAway remains advisory, not
clean; terminal output is partially truncated. Replacement-content assertions
could be stronger, but no current source defect was established.

Existing public-contract generation/check and all goal-byte pins pass locally.
The generated current authorization API still selects the old family integration
until the approved-main package checkpoint is integrated; that historical API
selection is not presented as the completed extension. Family integration,
historical compatibility closure, affected reverse consumers and coordinator
unit verification remain outstanding. This change certifies no verified unit,
hosted CI, release, publication or full-platform completion.
