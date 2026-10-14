# Local clean consumer

Go1.26.6 darwin/arm64; execution revision 9926ce00a89c4f3187b12b4b25e3268158e138e6
with the current dirty root implementation copied to a task-owned archived snapshot.
Snapshot files were precisely `git ls-files --cached --others --exclude-standard`;
reserved modules were included unchanged only because the existing owned proxy
builder enumerates the module catalog. Root source was consumed as a module ZIP,
not a sibling replacement. No replace directives were present in the consumer.

The actual repository `.golib/scripts/build-local-proxy.sh <proxy> v1.0.0 .`
built the root snapshot. A separate `example.invalid/authentication-clean-consumer`
module required github.com/faustbrian/go-authentication v1.0.0 and compiled and
exercised NewProof/Proof.Validate, NewOneTimeCode and the existing BearerCredential
API. `go mod tidy` then `go test ./... -count=1` both exited 0.
Native compiler output is clean-consumer.log.

Only the unpublished local root ZIP used the local proxy/checksum exemption;
go-clock v1.0.0 resolved normally through https://proxy.golang.org and sumdb.
GOWORK=off, GOTOOLCHAIN=go1.26.6; disposable GOCACHE/GOMODCACHE/GOTMPDIR.

SHA-256 inputs:

- root module ZIP: 47ddd8fbcf7506665c4952bb30c4b9ed7751d390230f8ea3d4899ae84dd77408
- consumer go.mod: f2e881eb9837c5eb0920bfa6af9886eda1b8b507d2a12676b30897737e9bb2fe
- consumer go.sum: 2d8a25d60820d81783cb0ae5f64a270081ffa3f88c1530e6340e77884551b916

All consumer resources, archived snapshot, module ZIP and caches were removed
on exit. This proves local clean-consumer compatibility, not publication,
hosted CI, released-module resolution or release provenance.
