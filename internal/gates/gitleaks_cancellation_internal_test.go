package gates

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func TestGitleaksEntryAdmissionCancellation(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"first", "second"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte("a"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	metadataCalls, visits := 0, 0
	err := inspectGitleaksCurrentTreeWithMetadata(ctx, root, securitySourceLimits{entries: 2, bytes: 2}, func(string, fs.DirEntry) error {
		visits++
		cancel()
		return nil
	}, func(_ string, entry fs.DirEntry) (fs.FileInfo, error) {
		metadataCalls++
		return entry.Info()
	})
	if !errors.Is(err, context.Canceled) || errors.Unwrap(err) != nil || metadataCalls != 1 || visits != 1 {
		t.Fatalf("inspection = %v, metadata=%d visits=%d; want original cancellation before second entry", err, metadataCalls, visits)
	}
}
