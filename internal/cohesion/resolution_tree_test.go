//go:build linux

package cohesion

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDigestResolutionTreeIsDeterministicAndRejectsAliases(t *testing.T) {
	root := t.TempDir()
	resolved, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}
	root = resolved
	if err := os.Mkdir(filepath.Join(root, "empty"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, "bin"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "bin", "go"), []byte("binary"), 0o755); err != nil {
		t.Fatal(err)
	}

	limits := ResolutionTreeLimits{MaximumFiles: 10, MaximumDirectories: 10, MaximumFileBytes: 10, MaximumTotalBytes: 20}
	first, err := DigestResolutionTree(root, limits)
	if err != nil {
		t.Fatal(err)
	}
	second, err := DigestResolutionTree(root, limits)
	if err != nil || second != first || first == "" {
		t.Fatalf("second digest = %q, first %q, error = %v", second, first, err)
	}
	if err := os.Symlink("bin/go", filepath.Join(root, "alias")); err != nil {
		t.Fatal(err)
	}
	if _, err := DigestResolutionTree(root, limits); err == nil {
		t.Fatal("DigestResolutionTree(symlink) error = nil")
	}
}
