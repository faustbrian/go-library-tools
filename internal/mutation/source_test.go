package mutation_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/faustbrian/go-library-tools/v2/internal/mutation"
)

func TestSourceDigestRejectsBareParentDirectoryEscape(t *testing.T) {
	parent := t.TempDir()
	root := filepath.Join(parent, "root")
	sibling := filepath.Join(parent, "sibling")
	nested := filepath.Join(root, "nested")
	for _, directory := range []string{root, sibling, nested} {
		if err := os.MkdirAll(directory, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	for _, directory := range []string{parent, root, sibling, nested} {
		if err := os.WriteFile(filepath.Join(directory, "source.go"), []byte("package example\n"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	for _, test := range []struct {
		name, module, pkg string
	}{
		{"module parent selects sibling", "..", "sibling"},
		{"package parent selects parent", ".", ".."},
	} {
		t.Run(test.name, func(t *testing.T) {
			digest, err := mutation.SourceDigest(root, test.module, test.pkg)
			if digest != "" || !errors.Is(err, mutation.ErrInvalid) {
				t.Fatalf("escaping SourceDigest(%q, %q) = %q, %v; want no identity and ErrInvalid", test.module, test.pkg, digest, err)
			}
		})
	}
	for _, directories := range [][2]string{{".", "."}, {".", "nested"}, {"nested", "."}} {
		digest, err := mutation.SourceDigest(root, directories[0], directories[1])
		if err != nil || len(digest) != 64 {
			t.Fatalf("ordinary SourceDigest(%v) = %q, %v", directories, digest, err)
		}
	}
}

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
