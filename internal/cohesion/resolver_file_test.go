package cohesion

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadResolutionFileRejectsSymlinksAndBoundsBytes(t *testing.T) {
	root := t.TempDir()
	resolvedRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}
	root = resolvedRoot
	realPath := filepath.Join(root, "real.json")
	if err := os.WriteFile(realPath, []byte("1234"), 0o600); err != nil {
		t.Fatal(err)
	}
	linkPath := filepath.Join(root, "link.json")
	if err := os.Symlink(realPath, linkPath); err != nil {
		t.Fatal(err)
	}

	if _, err := readResolutionFile(realPath, 4); err != nil {
		t.Fatalf("readResolutionFile(real) error = %v", err)
	}
	if _, err := readResolutionFile(realPath, 3); err == nil {
		t.Fatal("readResolutionFile(over limit) error = nil")
	}
	if _, err := readResolutionFile(linkPath, 4); err == nil {
		t.Fatal("readResolutionFile(symlink) error = nil")
	}

	linkDirectory := filepath.Join(root, "linked-directory")
	if err := os.Symlink(root, linkDirectory); err != nil {
		t.Fatal(err)
	}
	if _, err := readResolutionFile(filepath.Join(linkDirectory, "real.json"), 4); err == nil {
		t.Fatal("readResolutionFile(symlink component) error = nil")
	}
}
