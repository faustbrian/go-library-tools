package docscheck

import _ "embed"

var (
	//go:embed assets/package.json
	spellingPackage []byte
	//go:embed assets/package-lock.json
	spellingLock []byte
)

// SpellingAssets returns independent copies of the exact npm dependency graph
// used by the documentation spelling gate.
func SpellingAssets() ([]byte, []byte) {
	return append([]byte(nil), spellingPackage...), append([]byte(nil), spellingLock...)
}
