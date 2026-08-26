package mutation_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/faustbrian/go-library-tools/internal/mutation"
)

func TestSourceDigestMatchesLegacyContentIdentity(t *testing.T) {
	root := t.TempDir()
	directory := filepath.Join(root, "module", "adapter")
	if err := os.MkdirAll(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	for name, content := range map[string]string{
		"b.go":      "package adapter\nconst B = 2\n",
		"a.go":      "package adapter\nconst A = 1\n",
		"a_test.go": "package adapter\n",
		"notes.txt": "ignored\n",
	} {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	digest, err := mutation.SourceDigest(root, "module", "adapter")
	if err != nil {
		t.Fatal(err)
	}
	if digest != "eae0326a65fb3ecefd819ff6e3a7d8ae67f1dff53cf21f47b23b9347482db3e9" {
		t.Fatalf("SourceDigest() = %q", digest)
	}
}
