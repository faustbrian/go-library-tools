# Capability generation 0 native implementation checkpoint

Local committed implementation after parent-authorized exact-closure independent review. No integrated-unit, hosted CI, publication, downstream migration, deployment or production claim. No push.

Frozen complete source closure SHA256: 33477022febae1d13643e816e650e20640583a998f36ce203c532992d2ef92d2. Fifteen-path dirty closure SHA256: f4ea776da81951f49cd9828b894fc8ad20021cfbc61812dcf7a9025f2cdc3878. Algorithm: sorted NUL-delimited paths, shasum -a 256 per file, then shasum -a 256 of that manifest.

- Repository: github.com/faustbrian/go-capability.
- Sole workspace: /Users/brian/Developer/.worktrees/golib-go-capability-identity-platform-worker-g0.
- Branch: feature/identity-platform-capability-g0.
- Base: b0b21d15c9653decb0ff4a4fbffbe71c77d8d8fc. Sole worker commit/current HEAD: 4cfaa62b96a875a683f7e035f60f06cd2d8bca37. Tree: 226969a5a3b81f4625200b9c2ec7c564fc90685e.
- Assignment commit: fdb194246358d89a01d549485a7a6e15d9cf9aa4; exact assignment SHA256 b709498b85f2454e1da514e11ef4606f7f20cfb0490d4016601551147f096b63.
- Goal SHA256: 0c38f1416ef88d1020f2775db80f3bbce01f17c7350ae66cf09b24051b1988cc.
- Historical API and required-extension pins preserved: 3ad716bb7799ba0d0045b7456d2dabfe486503b9050b4d984dead16d69da7e14 and fa22b30449283df23ee2e4dca6f60531a54dd49b02d866aa2e63c55db9dde2af.

## Observable implementation

Additive Purpose, Reference, opaque aliasing Proof and package-owned Service constructors; literal purpose/audience matching, reference-domain digest, caller-configured limits, signature/time/revocation validation without implicit consumption, synchronous terminal transfer with full exposed-capacity cleanup and redacted diagnostics. Legacy Issue/Verify wire, timestamp/caveat and explicit consumption behavior remains unchanged. No postgres/valkey implementation or test edits.

Fifteen owned paths: identity_contracts.go; identity_contracts_test.go; identity_contracts_internal_test.go; identity_example_test.go; errors.go; docs/identity-contracts.md; docs/README.md; CHANGELOG.md; api/baseline.txt; scripts/check-clean-consumer.sh; scripts/check-fuzz.sh; scripts/check-interoperability.py; .golib/scripts/check-module.sh; .golib/scripts/gate-input-digest.sh; .golib/scripts/repository-check.sh.

Mechanical root corrections preserve gate thresholds: root package admitted to exact-source fingerprints; worktree .git pointer excluded from sibling-source scan; conformance/benchmark calls use existing package makefile dispatcher. Disposable digest controls prove root production, root observer and consumed-descendant sensitivity plus unrelated adapter exclusion.

## Requirement coverage

| Original QA slice | Native behavioral boundary |
|---|---|
| CAP-Q01 | Old-baseline API compatibility before regeneration; new external clean consumer and constructor/error tests |
| CAP-Q02 | PurposeExactBounds and PurposeReferenceAndIngressBounds; independent Python reference digest oracle |
| CAP-Q03 | ServiceNeverConsumesAndHonorsLimits; configured 16KiB default, above-4096 valid ingress and zero/error outputs |
| CAP-Q04 | TestIdentityServiceRejectsStructurallyValidForgedSignature: canonical modified signature parses through public ParseProof, then matching active Service rejects ErrInvalidSignature with completely zero Grant; genuine signed control succeeds |
| CAP-Q05 | ProofTransferInvalidatesAliases and TransferFailureCustody; shared lifetime and retained entire-capacity erasure |
| CAP-Q06 | Nil callback/context and pre-cancel tests; preserved custody and no callback execution |
| CAP-Q07 | Callback error/panic, custom classifier panic and legacy panic(nil) subprocess; redacted errors, cleanup and no retries |
| CAP-Q08 | Joined barrier-driven concurrent transfer and Verify snapshot ordering; context methods checked outside custody lock |
| CAP-Q09 | Real signer/resolver, exact purpose/audience, time and revocation checks through Service |
| CAP-Q10 | Repeated Verify leaves explicit store first consumption ordinal 1; unchanged legacy replay/unknown-outcome suite |
| CAP-Q11 | Unchanged pinned golden/wire/timestamp/caveat and caphttp suite; original API comparison |
| CAP-Q12 | ProofRepresentations value/pointer/container formatting, JSON and structured logging non-disclosure |
| CAP-Q13 | Bounded deterministic ingress and registered FuzzIdentityProofIngress |
| CAP-Q14 | Compiled ExampleNewService, guide and native exhaustive module gates |

## Current native verification

`bash .artifacts/capability-g0/run.sh authentication-final-gates make ci`: exit 0. Exact command, Go1.26.6 environment and pre-execution complete source hashes retained in authentication-final-gates.inputs (SHA256 4cc84b87da5e680f1deb65c40ae6e96a5b5bf119682265191ad467ec5b4ce5da); full native stdout/stderr in authentication-final-gates.log. Fresh root mutation input fingerprint: 8f11ed0d203b44a4ad4b53c614493ed43f93397c2810504c10910ee56eb03f8e. Four adapter/package mutation checkpoints are reused only after complete input equality.

Coverage: root 658/658; caphttp 88/88; memory 84/84; postgres 92/92; valkey 38/38. Mutation: root 432/432; caphttp 44/44; memory 37/37; postgres 51/51; valkey 38/38. Every viable mutant killed, exact efficacy/coverage 100%, no survivors, uncovered, timeouts, nonviable or skipped. Native per-package checkpoints under .artifacts/mutation-checkpoints and complete schema-3 aggregate .artifacts/mutation.json.

Inventory, format/tidy, vet, race, coverage, lint, staticcheck, vulnerability/secrets/licenses, SBOM, four bounded fuzz targets, docs/examples/API, conformance/official fixtures, Python interoperability and native benchmark contract passed. Explicit clean consumer passed in clean-consumer.log. Joined custody race stress (100 iterations) and legacy panicnil red/green artifacts retained.

Historical exhaustive run full-final-inputs-2 passed, followed by a documentation-only ASCII C0/DEL clarification and passing docs-final gate; frozen-mutation-inputs proved every complete package mutation fingerprint unchanged by prose. The fresh native regression review then requested CAP-R1/CAP-R2 test strengthening. Only identity_contracts_test.go changed; genuine production bytes were restored exactly after isolated proving counterfactuals. Current review-final-gates freshly executes changed root mutation inputs (73c8ca3dda444ec6a65657ee3faaf78ddace6472d4d6e003cec2e7bff5e80345) and reuses four content-identical package mutation checkpoints.

CAP-R1: genuinely signed correct-audience matching control, independent valid wrong/prefix/case purposes, exact ErrUnauthorized and completely zero Grant. Omitting only the purpose condition causes all four cases to fail (review-purpose-counterfactual-red native exit 1).

CAP-R2: explicit Clock/IAT/future NBF are distinct, with preactivation ErrNotYetValid and completely zero Grant; activated signed grant preserves both timestamps. Mixed zero/nonzero cases default only the zero field. Unconditional Clock overwrite causes all three cases to fail (review-timestamp-counterfactual-red native exit 1). Restored real implementation passes both new tests under race (review-regressions-green native exit 0). Production identity_contracts.go SHA256 remains 6ed5904496b10e3bc33df789a2f0219ba2dda1b6b5073299ed9306bc3cdbb452. These are implementation responses, not self-attested independent review disposition.

The subsequent native regression identified CAP-Q05/Q07's unpoisoned spare-capacity oracle gap. Public ParseProof/Transfer measurement of the existing canonical golden proof exposes len 367/cap 384/spare 17 (public-golden-borrow-capacity native exit 0; integer sizes only). TestIdentityProofTransferErasesPoisonedSpareCapacity fills every actually exposed tail byte with 0xa5 on success/error/panic, retains the full allocation alias, then asserts zeroed allocation and terminal shared custody. Isolated length-only clearing fails all three cases at retained allocation byte 367 (review-tail-counterfactual-red exit 1). Restored full-capacity clearing passes all three under race (review-tail-green exit 0); genuine production bytes remain unchanged. Current tail-final-gates freshly executes changed root mutation inputs 962efddfd6c4f9be636ae84067f44a1421a86489291b9e9d632b2fc4208cded7 and reuses four content-identical package mutation checkpoints. Tail gate diagnostics again byte-match the native pre-change NilAway baseline after directory-only normalization.

CAP-R4 attribution correction: the earlier CAP-Q04 row overstated proof at the new Service boundary. Prior signature-tamper tests exercised legacy Verify only, not public ParseProof followed by actual Service rejection. That prior claim is withdrawn. The new named public test now independently signs a genuine active success control, flips a decoded signature byte and canonically reencodes it, proves ParseProof accepts its structure, and requires matching-purpose/audience Service rejection with ErrInvalidSignature and a completely zero Grant. Replacing authenticated Service verification with structural Parse→Grant promotion fails the new test (review-authentication-counterfactual-red exit 1). Restored genuine Service passes this and all strengthened purpose/timestamp/tail tests together under race (review-authentication-green exit 0). Production bytes are unchanged; the fresh authentication-final-gates aggregate run passed exit 0. Current advisory diagnostics again byte-match the native pre-change NilAway baseline after directory-only normalization.

NilAway is advisory, not clean: native exit 3. Exact pre-change b0b21 archive with the same pinned analyzer/Go emits identical diagnostics after directory-only normalization (nilaway-compare.sh exit 0); no new diagnostics. Baseline log SHA256 17a13ca33bd203382266f3b78c0c86e2b11fa3fa2103aa6b3a30cfdc2fec3c34. Its go-run wrapper exit is 1, with native analyzer exit 3 explicitly retained. SBOM emitted a main-module Git-version lookup warning; version/provenance attribution is not certified.

All Go invocations used disposable GOCACHE/GOMODCACHE/GOTMPDIR; caches and baseline fixtures removed on exits. Native module gates created and removed uniquely owned PostgreSQL/Valkey service fixtures; no existing user service accessed. Evidence remains locally retained for parent review custody. The task-generated mutation executable was removed after its SHA256 was independently matched to the five native checkpoints; native verifier/binary identities remain in those checkpoints.

## Known failed evidence retained

First missing-feature red preceded first production write, reported natively to parent. Further decisive red/green evidence: invalid purpose suffix survivor, caller Context.Err under lock, legacy panic(nil), redacted panic cancellation classifier. Failed prior inventory .git-pointer scan, lint/test diagnostics, mutation survivor/timeouts, and conformance dispatch are retained rather than represented as passes; full-final-inputs-2 is the current successful exhaustive run.

## Parent-owned remaining boundaries

Parent confirmed fresh regression, complete-final-diff correctness and Go compatibility approvals for this exact source closure and authorized the sole commit. Post-commit complete source closure remains byte-identical to the reviewed and executed closure; no expensive gate was repeated for commit metadata alone. Staged secret scan, 72-column message validator, Intent Commits Core ACK validation and normal Git hooks passed; effective release impact is unspecified, with no applicable repository ACK profile found. Native ignored evidence was not staged. Counterfactual mutation patches were native apply_patch tool calls in the worker conversation; no standalone executed patch/diff files were retained. Red .inputs/.log/.exit files establish exact native commands, temporary source identities and failures, not reconstructable mutation preimages.

Construction declarations catalog reconciliation without changing required extension digest, coordinator integration and integrated unit verification, reserved adapter future reconciliation, actual downstream service/reference composition and migration, hosted CI, replacement-free release consumer, publication/provenance and delivery-state audit remain parent-owned/unverified here.
