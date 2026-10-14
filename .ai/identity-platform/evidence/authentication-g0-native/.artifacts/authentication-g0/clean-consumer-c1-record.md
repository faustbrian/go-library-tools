# Current C1-repaired clean consumer

Go1.26.6 darwin/arm64, execution revision
9926ce00a89c4f3187b12b4b25e3268158e138e6, dirty owned source identified by
final-dirty-manifest.sha256 SHA11722560f78341a6c0301173e3f9a6dd1dc86d6c19fd377ba6de49015dea86b5.
Exact snapshot selection: git ls-files --cached --others --exclude-standard.
Disposable Git index supports actual build-local-proxy.sh proxy v1.0.0 .;
separate clean consumer uses the root ZIP without replace directives. Public
clock dependency resolves with ordinary public proxy and checksum security;
only exact unpublished root namespace has local ZIP GONOSUMDB. GOWORK=off;
task-owned build/module/temp caches, snapshot, index, consumer and proxy cleaned.

go mod tidy and go test ./... -count=1 both passed. Native output:
clean-consumer-c1.log. Unchanged consumer inputs: consumer-input/.

SHA256 inputs:
- root ZIP: 60f3b94f733cf9708d3cd037bf765c32eec8a13948dad751a72675621a531322
- consumer go.mod: f2e881eb9837c5eb0920bfa6af9886eda1b8b507d2a12676b30897737e9bb2fe
- consumer go.sum: 4451ba3ca7955fbd15a0451b31fc2a203c9efed259e53d551799dcccc8003edd

Local consumer compatibility only, not publication, hosted CI, release
provenance, reserved-module completion or downstream identity integration.
