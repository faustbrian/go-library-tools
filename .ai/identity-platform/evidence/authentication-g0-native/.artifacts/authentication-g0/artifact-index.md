# Authentication generation 0 artifact index

This index distinguishes current native evidence from historical failures and
earlier snapshots. It is not an independent approval or coordinator verified
record. Frozen owned source manifest: final-dirty-manifest.sha256,
SHA11722560f78341a6c0301173e3f9a6dd1dc86d6c19fd377ba6de49015dea86b5.
Execution revision before local implementation commit:
9926ce00a89c4f3187b12b4b25e3268158e138e6. Go1.26.6 darwin/arm64.

## Current native evidence

`.artifacts/evidence/<gate>.json` and corresponding `.log` are the current
native records for format-check, tidy-check, safety, vet, test, race, coverage,
fuzz, api, conformance, interoperability, lint, staticcheck, vulnerability,
secrets, licenses, sbom, docs, benchmark and mutation. Each record retains
complete input/completed digests, original execution revision, environment,
exit status and native log checksum. NilAway is advisory with seven existing
baseline findings; it is not a clean analyzer pass. SBOM retains its main-module
version warning, not release provenance. `format` is historical invalidated,
not the current passing `format-check` result.

Exact coverage: root561/561, apikey93/93, authhttp243/243, authlog5/5,
basic52/52, bearer83/83. Nine hostile fuzz targets executed. Current mutation
actually reran all six packages:456+52+155+2+28+45=738/738 viable killed,
no lived/uncovered/nonviable/skipped/timeouts. `.artifacts/mutation.json` and
`.artifacts/mutation-checkpoints/*.json` retain complete native reports.
Mutation wrapper input/completed digest:
aeceda62a069047544cfe5ac17bfd06a5698fc7110dcc27fa25aba85a2186ea0.
Native log SHA131c22f992b96543025bef18f20b800819e918d631503c61c80edf988c7f413a;
aggregate JSON SHA96953c07f109125e6b73bbd612f1077639da850141b325f4b6135b98a23a2915.

Current clean consumer: clean-consumer-c1.log and clean-consumer-c1-record.md.
The retained `consumer-input/go.mod` is the **pre-tidy input**, SHA
438d053ec9616ae73fd45aab8c431e3dd49980cc44c815c2b3327c9cda114b69;
consumer_test.go SHAfd14e7cb46084b730a2fb5112c1a81279df4f0028e32e6a5498fa19bbf294321.
The record's go.mod SHAf2e881eb9837c5eb0920bfa6af9886eda1b8b507d2a12676b30897737e9bb2fe
is the deterministic **post-tidy result**, not the retained pre-tidy bytes.
Post-tidy go.sum SHA4451ba3ca7955fbd15a0451b31fc2a203c9efed259e53d551799dcccc8003edd;
root ZIP SHA60f3b94f733cf9708d3cd037bf765c32eec8a13948dad751a72675621a531322.
All disposable consumer resources were removed; raw outputs are unchanged.

Current inventory: c1-inventory.log, correct root `make inventory`, exit0.
Old-to-new API: original-api-root.txt and api-compatibility-record.md preserve
original committed baseline plus actual incompatible-only apidiff exit0;
the current api gate alone compares the regenerated additive baseline.

## Decisive red/green and gate-boundary records

- json-profile-red.log/json-profile-green.log: signed public token alias bug.
- c1-controls-red.log/c1-controls-green.log/c1-controls-boundary.md: public
  C1 identity admission defect, shared consumer inventory and repaired outcomes.
- mutation-input-correction.md/mutation-directory-controls.json/
  mutation-root-directory.patch: old omission and repaired root/descendant
  positives plus unrelated-path negative controls.
- mutation-portable-identity.json/mutation-upstream-checksum.json: portable
  verifier identity includes original authenticated upstream checksums, exact
  semantic patches, enabled operators and invocation policy; no exemptions.
- dispatch-failure-probe.log: actual dispatcher propagates deliberate target
  failure, not a passing conformance run.
- secret-negative-ordinary.log/secret-negative-same-path.log: actual scanner
  still rejects ordinary-source JWT and unrelated secret rules in fixture path.

## Historical evidence and claim boundaries

Every `.artifacts/evidence/by-input/` record retains its original result and
input identity. Failed lint/conformance/secrets, invalidated format/race/secrets,
and earlier advisory findings are historical, not current passes. Earlier
mutation checkpoints with incomplete root fingerprints and prior742 campaigns
must not certify this738 result. Earlier clean-consumer records and logs
describe their distinct ZIP snapshots; they are not the current C1 snapshot.
c1-inventory-wrong-dispatch.log retains the missing-target failure from selecting
the package Makefile for the repository inventory target; correct root native
inventory passes separately. No failed evidence was rewritten as passing.

These are local package checks only. Hosted CI, public publication/release
provenance, reserved nested modules, reverse-consumer composition, coordinator
catalog reconciliation/integration and unit-verified inventory are not certified.
