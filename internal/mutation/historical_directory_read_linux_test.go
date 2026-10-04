//go:build linux

package mutation

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Hosted Linux must enforce directory read permissions; privileged execution
// that bypasses the refusal fails this assertion rather than skipping it.
func TestHistoricalPackageProofPreservesDirectoryReadRefusal(t *testing.T) {
	root := t.TempDir()
	directory := filepath.Join(root, "ordinary")
	if err := os.Mkdir(directory, 0o700); err != nil {
		t.Fatal("create task-owned source directory")
	}
	if err := os.WriteFile(filepath.Join(directory, "source.go"), []byte("package ordinary\n"), 0o600); err != nil {
		t.Fatal("create ordinary production source")
	}
	source, err := SourceDigest(root, ".", "ordinary")
	if err != nil {
		t.Fatal("identify ordinary production source")
	}
	if err := legacyPackagePathsUnchanged(root, ".", "ordinary", "example", source); err != nil {
		t.Fatal("readable ordinary source must admit historical package proof")
	}

	// Registered after TempDir cleanup, so LIFO restores access before removal,
	// including when changing permissions or any later assertion fails.
	t.Cleanup(func() {
		if err := os.Chmod(directory, 0o700); err != nil {
			t.Error("restore task-owned source directory permissions")
		}
	})
	if err := os.Chmod(directory, 0o111); err != nil {
		t.Fatal("remove task-owned source directory read permission")
	}

	err = legacyPackagePathsUnchanged(root, ".", "ordinary", "example", source)
	if !errors.Is(err, ErrInputChanged) || !errors.Is(err, os.ErrPermission) {
		t.Fatal("historical proof must refuse unreadable source and preserve its permission cause")
	}
	if !strings.HasPrefix(err.Error(), ErrInputChanged.Error()+": inspect historical package source: ") {
		t.Fatal("directory read refusal must retain the historical source category")
	}
}
