//go:build darwin

package cohesion

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDigestResolutionTreeFailsClosedOnExtendedAttributes(t *testing.T) {
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "file"), []byte("value"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err = DigestResolutionTree(root, ResolutionTreeLimits{MaximumFiles: 1, MaximumDirectories: 0, MaximumFileBytes: 5, MaximumTotalBytes: 5})
	if err == nil {
		t.Fatal("DigestResolutionTree(extended attributes) error = nil")
	}
}
