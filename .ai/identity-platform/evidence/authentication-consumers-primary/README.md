# Candidate-bound authentication consumer verification

Authentication integration `1c479773ed26604d9cd7cb6ffb471423458abb61` restores
the released HTTP and logging owners, legacy facades and separate Begin seam.
Unchanged service compatibility, service reference HTTP and queue authz/server
all passed `go test -race -count=1 -timeout=180s` against disposable candidate
snapshots. `results` contains native source identities, complete input manifests,
commands, toolchain/environment metadata, outputs and checkpointed results.
The existing attempt-1 runner was adapted to these three affected consumers;
acquisition errors now retain complete stderr rather than a truncated prefix.

Consumer source and original required versions were preserved. Temporary local
module replacements are explicitly recorded and do not prove published,
replacement-free consumption. Root and optional release gates, hosted CI,
publication and deployment remain pending. Disposable snapshots and caches were
removed by the runner after all checkpointed results were captured.

The raw stdin secret scan of the nested checksum commit reported seven
generic-api-key matches, all public Go module checksum lines (six changed
authentication archive entries and one unchanged oauth2 context line), not
credentials. Its redacted native report is retained as `admission-scan.json`;
no raw-stdin scanner pass is claimed. Native path-aware directory scans of
authotel, jwt and oidc passed without changing scanner configuration.

The consumer evidence directory scan reported three generic-api-key matches at
`inputs.json` line 7. Each is the public authentication integration commit in
`candidate_revisions`, not a credential. The redacted native report is retained
as `control-scan.json`; no directory-scanner pass is claimed for these records.
