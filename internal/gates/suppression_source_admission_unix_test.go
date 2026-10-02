//go:build darwin || linux

package gates

import (
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

// A FIFO is created and metadata-inspected only. An inert open collaborator
// ensures the unsafe pre-fix state cannot open, read, or write the FIFO.
func TestSuppressionSourceRefusesNonregularMetadataBeforeOpen(t *testing.T) {
	rootPath := t.TempDir()
	path := filepath.Join(rootPath, "source.go")
	if err := syscall.Mkfifo(path, 0o600); err != nil {
		t.Fatal(err)
	}
	root, err := os.OpenRoot(rootPath)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	entries, err := os.ReadDir(rootPath)
	if err != nil || len(entries) != 1 {
		t.Fatalf("inspect owned FIFO metadata: entries=%d error=%v", len(entries), err)
	}
	info, err := entries[0].Info()
	if err != nil || info.Mode()&os.ModeNamedPipe == 0 {
		t.Fatalf("fixture is not a real FIFO: %v", err)
	}
	opens := 0
	content, err := readSecuritySuppressionSource(root, "source.go", entries[0], 8, securitySuppressionSourceFiles{
		info: os.DirEntry.Info,
		open: func(*os.Root, string) (*os.File, error) {
			opens++
			return nil, errors.New("inert open must not be dispatched")
		},
		close: (*os.File).Close,
	})
	if err == nil || err.Error() != "security suppression source is not a regular file: source.go" || content != nil || opens != 0 {
		t.Fatalf("nonregular source reached open: error=%v content=%q opens=%d", err, content, opens)
	}
	if after, err := os.Lstat(path); err != nil || !os.SameFile(info, after) || after.Mode()&os.ModeNamedPipe == 0 {
		t.Fatal("source refusal changed FIFO identity or type")
	}
}
