package mutation

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/faustbrian/go-library-tools/v2/internal/repositoryfile"
)

func TestHistoricalPackageProofRejectsOrdinaryAdmissionFailures(t *testing.T) {
	for _, test := range []struct {
		name, content, want string
		directoryFile       bool
		cause               error
	}{
		{"nonregular source", "", "read historical package declaration", true, repositoryfile.ErrNotRegular},
		{"invalid package clause", "package\n", "historical package declaration cannot be proven", false, nil},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			path := filepath.Join(root, "source.go")
			if test.directoryFile {
				if err := os.Mkdir(path, 0o700); err != nil {
					t.Fatal(err)
				}
			} else if err := os.WriteFile(path, []byte(test.content), 0o600); err != nil {
				t.Fatal(err)
			}
			err := legacyPackagePathsUnchanged(root, ".", ".", "example", strings.Repeat("a", 64))
			if !errors.Is(err, ErrInputChanged) || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("package proof unexpectedly admitted invalid source: %v", err)
			}
			if test.cause != nil && !errors.Is(err, test.cause) {
				t.Fatalf("underlying filesystem classification lost: %v", err)
			}
		})
	}
	t.Run("missing package directory", func(t *testing.T) {
		err := legacyPackagePathsUnchanged(t.TempDir(), ".", "missing", "example", strings.Repeat("a", 64))
		if !errors.Is(err, ErrInputChanged) || !errors.Is(err, os.ErrNotExist) || !strings.Contains(err.Error(), "inspect historical package directory") {
			t.Fatalf("missing directory admission = %v", err)
		}
	})
	t.Run("no production source", func(t *testing.T) {
		root := t.TempDir()
		if err := os.WriteFile(filepath.Join(root, "ordinary_test.go"), []byte("package example\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		err := legacyPackagePathsUnchanged(root, ".", ".", "example", strings.Repeat("a", 64))
		if !errors.Is(err, ErrInputChanged) || !strings.Contains(err.Error(), "no provable production files") {
			t.Fatalf("test-only source admitted: %v", err)
		}
	})
}

func TestHistoricalPackageProofRejectsWrongExpectedDigestAfterValidControl(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "source.go"), []byte("package example\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	digest, err := SourceDigest(root, ".", ".")
	if err != nil {
		t.Fatal(err)
	}
	if err := legacyPackagePathsUnchanged(root, ".", ".", "example", digest); err != nil {
		t.Fatalf("valid source rejected: %v", err)
	}
	first := "0"
	if digest[:1] == first {
		first = "1"
	}
	err = legacyPackagePathsUnchanged(root, ".", ".", "example", first+digest[1:])
	if !errors.Is(err, ErrInputChanged) || !strings.Contains(err.Error(), "declaration proof does not match mutation source") {
		t.Fatalf("wrong expected digest admitted: %v", err)
	}
}
