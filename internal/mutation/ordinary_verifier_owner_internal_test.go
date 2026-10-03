package mutation

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOrdinaryVerifierOwnerBindsEveryDirectProductionFile(t *testing.T) {
	root := t.TempDir()
	directory := filepath.Join(root, "nested", "adapters", "wire")
	if err := os.MkdirAll(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	first := []byte("package wire\nconst First = 1\n")
	second := []byte("package wire\nconst Second = 2\n")
	for _, file := range []struct {
		name string
		data []byte
	}{
		{"z.go", second}, // Creation order must not change the sorted manifest.
		{"a.go", first},
		{"ordinary_test.go", []byte("package wire_test\n")},
		{"README.md", []byte("Ordinary adapter documentation.\n")},
	} {
		if err := os.WriteFile(filepath.Join(directory, file.name), file.data, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	manifest := fmt.Sprintf("%x  nested/adapters/wire/a.go\n%x  nested/adapters/wire/z.go\n", sha256.Sum256(first), sha256.Sum256(second))
	expected := fmt.Sprintf("%x", sha256.Sum256([]byte(manifest)))
	actual, err := SourceDigest(root, "nested", "adapters/wire")
	if err != nil || actual != expected {
		t.Fatalf("direct source identity = %q, %v; want independently calculated %q", actual, err, expected)
	}
	if err := legacyPackagePathsUnchanged(root, "nested", "adapters/wire", "example/v2", expected); err != nil {
		t.Fatalf("unchanged multi-file declaration proof rejected: %v", err)
	}

	// A changed ordinary production value keeps the declaration valid but
	// invalidates the retained source identity, unlike ignored test/docs files.
	if err := os.WriteFile(filepath.Join(directory, "z.go"), []byte("package wire\nconst Second = 3\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	err = legacyPackagePathsUnchanged(root, "nested", "adapters/wire", "example/v2", expected)
	if !errors.Is(err, ErrInputChanged) || !strings.Contains(err.Error(), "declaration proof does not match mutation source") {
		t.Fatalf("changed direct production bytes admitted: %v", err)
	}
}

func TestOrdinaryVerifierOwnerRejectsAliasDespiteMatchingSourceIdentity(t *testing.T) {
	root := t.TempDir()
	directory := filepath.Join(root, "adapters", "wire")
	if err := os.MkdirAll(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "source.go"), []byte("package adapter\nconst Value = 1\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	digest, err := SourceDigest(root, ".", "adapters/wire")
	if err != nil {
		t.Fatal(err)
	}
	err = legacyPackagePathsUnchanged(root, ".", "adapters/wire", "example/v2", digest)
	if !errors.Is(err, ErrInputChanged) || !strings.Contains(err.Error(), "historical verifier selected a different package") {
		t.Fatalf("matching digest bypassed direct package-path compatibility: %v", err)
	}
}
