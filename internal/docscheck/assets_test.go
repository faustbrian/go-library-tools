package docscheck_test

import (
	"bytes"
	"testing"

	"github.com/faustbrian/go-library-tools/internal/docscheck"
)

func TestSpellingAssetsArePinnedAndSelfContained(t *testing.T) {
	packageJSON, packageLock := docscheck.SpellingAssets()
	if !bytes.Contains(packageJSON, []byte(`"cspell": "10.0.0"`)) {
		t.Fatalf("package.json does not pin cspell: %s", packageJSON)
	}
	if !bytes.Contains(packageLock, []byte(`"lockfileVersion": 3`)) ||
		!bytes.Contains(packageLock, []byte(`"cspell": "10.0.0"`)) {
		t.Fatal("package-lock.json does not bind the pinned spelling tool")
	}
	if len(packageJSON) == 0 {
		t.Fatal("package.json asset is empty")
	}
	packageJSON[0] ^= 1
	second, _ := docscheck.SpellingAssets()
	if bytes.Equal(packageJSON, second) {
		t.Fatal("SpellingAssets returned shared mutable storage")
	}
}
