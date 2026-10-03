package evidence_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/faustbrian/go-library-tools/v2/internal/evidence"
)

func TestOrdinaryEvidenceInspectAdmitsRealDirectoriesOnly(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "target", "evidence"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("target", filepath.Join(root, "alias")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("missing", filepath.Join(root, "dangling")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "file"), []byte("ordinary"), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"alias/evidence", "dangling/evidence", "file/evidence", "file"} {
		t.Run(path, func(t *testing.T) {
			got, err := evidence.Inspect(root, path, "ordinary", nil)
			if !errors.Is(err, evidence.ErrInvalid) || got != nil || err.Error() != "invalid evidence: evidence root is not a real directory" {
				t.Fatalf("directory admission = %#v, %v; want categorical rejection", got, err)
			}
		})
	}
	for _, path := range []string{"target/evidence", "target/missing", "missing/evidence"} {
		got, err := evidence.Inspect(root, path, "ordinary", nil)
		if err != nil || got == nil || len(got) != 0 {
			t.Fatalf("real or missing directory %s = %#v, %v", path, got, err)
		}
	}
}
