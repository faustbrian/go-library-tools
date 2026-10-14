# Final local clean consumer

Go1.26.6 darwin/arm64, execution revision 9926ce00a89c4f3187b12b4b25e3268158e138e6,
current dirty source snapshot after redundant-guard simplification and direct parent
lineage assertions. Snapshot input: git ls-files --cached --others --exclude-standard.
The snapshot alone received a disposable Git index for the actual repository
.golib/scripts/build-local-proxy.sh <proxy> v1.0.0 . invocation.

Separate example.invalid/authentication-clean-consumer module, no replace directives;
go mod tidy and go test ./... -count=1 exited 0. Native output:
clean-consumer-final.log. Exact consumer inputs retained in consumer-input/.
The unpublished root ZIP used file proxy and exact root namespace GONOSUMDB;
go-clock resolved normally via public proxy and sumdb. No checksum bypass for
public dependencies. GOWORK=off, task-owned GOCACHE/GOMODCACHE/GOTMPDIR.

SHA256:
- root ZIP: 51f50ece1c7616f38dab9bc0424d289b999cd136a0a32cef6a4e77a28a4e8bcc
- consumer go.mod: f2e881eb9837c5eb0920bfa6af9886eda1b8b507d2a12676b30897737e9bb2fe
- consumer go.sum: 1a8e12892288d367f79defd4002d2f780207b5fe9fab0d175176cf9f38b8794b

All disposable snapshot, proxy, consumer and caches cleaned on exit. Local
consumer compatibility only; publication, hosted CI and release provenance remain
unverified external boundaries.
