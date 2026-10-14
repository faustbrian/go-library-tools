# Capability: constructive service and lookup metadata

The integrated family checkpoint is
`c43af8abe83f33119ce2b13ec978e556408344f3`, with reviewed worker source tree
`226969a5a3b81f4625200b9c2ec7c564fc90685e`. Source declarations were
inspected directly. No package source, execution state or assignment changes
are included in this coordinator reconciliation.

The external catalog now exposes the source-exact ServiceConfig and NewService
construction boundary. It requires explicit limits, a trusted fast local clock,
concurrently safe collaborators when shared and nonnegative skew; absent
revocation evidence is permitted only under explicit application policy.
The Reference declaration retains its original exported fields. Its prose now
states that this metadata is mutable and unauthoritative: NewReference checks a
nonzero supplied digest, not its preimage or capability existence. Service.Issue
derives the domain-separated digest from its canonical token. This removes an
unsupported immutability/authentication claim, not the digest requirement.

The complete canonical required-contract digest is
`sha256:56dcf10ab5f1212e1b5a114d5c13ee3d24a933f1e7fc19c461c1919a6137d4fd`.
The updated complete goal-body pin is
`3b11c998df7b07492605efdfa246bb0fc3002ac4acae3ac52e9f7bc422e23295`.
Historical current-API start pins and every prior source-review and native
execution identity remain unchanged. Public-contract generation/check and all
goal-byte pins pass locally. No expensive package gate was rerun merely for a
catalog change. Coordinator integrated verification, affected reverse-consumer
closure and delivery audit remain outstanding; capability remains
implemented-unverified, not release-ready or a completed runtime milestone.
