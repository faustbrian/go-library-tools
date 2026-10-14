# Capability: integrated coordinator checks

The integrated dependency is `c43af8abe83f33119ce2b13ec978e556408344f3`,
tree `226969a5a3b81f4625200b9c2ec7c564fc90685e`. Package source remained
clean throughout. No source or gate configuration was changed for these runs.

The coordinator directly executed `/usr/bin/make ci`: exit 0, from
2026-10-01T10:47:06Z to 10:51:56Z. `native-result.json` and `native.log`
retain the actual launch, timestamps, result and merged stream identity.
`input-manifest.json` was captured before execution, not reconstructed as a
fictional earlier receipt. Its root is
`sha256:c8a916e642ba0e2c6d31bab7e4323d96faf0df52f674ea16e8e829d364af5a2f`.
The existing canonical reader independently rebuilt that manifest and all
environment preimage hashes matched `environment-inputs.json`.

The later combined closure uses the same canonical reader, with the focused
external and reverse-consumer pre-execution identities added to the retained
native preimages. `complete-input-manifest.json` and
`complete-environment-inputs.json` bind the recorded results for future
content-sensitive reuse. Their root is
`sha256:d776965fbca77d3ecb43466e25d98734209d7cf14dbf7e3c3b8895d42bdbf33a`.
This combined manifest was assembled after the separate runs; it is not
falsely attributed to the beginning of the original aggregate execution.

All five production packages have exact coverage: 658/658, 88/88, 84/84,
92/92 and 38/38 statements. The native mutation gate revalidated and reused
content-identical checkpoints: 602 killed, none survived. Original execution
revisions remain in the reports; this was not a newly executed 602-mutant
campaign. Native race, fuzz, documentation, API, conformance, Python HMAC
interoperability, benchmarks and configured supply-chain checks passed.
NilAway retains advisory baseline diagnostics; this is not a clean NilAway
claim. The SBOM main-module version qualification remains.

The original aggregate's live vulnerability feed and service image tags were
not fully bound by its pre-execution record. Only those external boundaries
and the separately required clean consumer received focused execution:

- `scripts/check-clean-consumer.sh`: exit 0, old and constructive new public
  APIs exercised in the existing task-owned temporary-replace consumer.
  This is not replacement-free released consumption.
- Official pinned govulncheck v1.6.0, supported `-db file://...` mode:
  exit 0, 172 native messages, zero findings. The official database ZIP is
  the unchanged same-day snapshot already retained at
  `../password-g0-narrow/vulndb.zip`, SHA-256
  `faa8c45538d259ae1fd3c34ae20632ee738d1190836c84ac95d5f2bc604d85bc`.
  Scanner executable/build identity and database identity were captured
  before scanning. No live database was consulted during this focused scan.
- `go test -count=1 -race -tags=integration ./postgres ./valkey`: exit 0.
  The fresh task-owned fixtures were launched using exact immutable local
  image IDs captured before execution; each running image was checked against
  that ID. Inputs include registry digests, architecture and OS. The source's
  mutable image tags are not presented as immutable evidence for the earlier
  aggregate run.
- The existing reverse consumer, `go-service/integration/reference-http`,
  was snapshotted from committed revision
  `b0a6f412d2f5aa3c3fb067222f96dc4242b8034f`. Only its disposable manifest
  replaced capability with the unchanged integrated source; other dependencies
  retained their selected versions. `go test -race -count=1 -timeout=90s
  ./...` passed. `reverse/` retains the original Git archive, temporary
  manifests, pre-execution identities, distinct stdout/stderr and result.
  No source-repository files or refs were changed. This proves the existing
  consumer's request/lifecycle and tampered-capability regressions against
  this integration, not future identity or released-platform adoption.

Raw artifacts were copied without changing bytes and source/destination hashes
matched. All task-owned caches, downloaded tools, extracted databases,
temporary consumer sources and fixture containers/anonymous volumes were
removed. Pre-existing shared images were retained.

The default raw-evidence secret scan exited 1: 60 sourcegraph-token matches
were independently classified as official public go.googlesource.com Go/net
revision URLs, and one generic-key match was the Valkey package's mutation
input SHA-256. The redacted native report is preserved. Neither scanner rules
nor raw artifacts were changed, and that nonzero result is not called a pass.

Native independent `/root/capability_delivery_g0` inspected the four original
QA/regression/correctness/Go reports directly from their producing principals,
the committed integrated source and native coordinator evidence. Its original
disposition is APPROVE for local integrated verification and DAG eligibility.
Committed inventory/ledger disposition remains required before this unit is
recorded verified. No publication,
hosted CI, deployment, shared-issuer runtime milestone or full 67-unit
completion is claimed. No push was performed.
