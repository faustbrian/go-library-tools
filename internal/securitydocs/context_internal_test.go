package securitydocs

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/faustbrian/go-library-tools/v2/internal/repositoryfile"
)

func TestSecurityValidationStopsAtCooperativeReadCheckpoint(t *testing.T) {
	for _, canceled := range []bool{false, true} {
		ctx, cancel := context.WithCancel(context.Background())
		calls := 0
		err := validateAtContext(ctx, t.TempDir(), time.Now(), func(_ *os.Root, path string) ([]byte, error) {
			calls++
			if path == "risk-register.json" {
				if canceled {
					cancel()
				}
				return []byte(`{"schema_version":1,"risks":[]}`), nil
			}
			return []byte(`{"schema_version":1,"modules":[]}`), nil
		})
		cancel()
		if canceled {
			if !errors.Is(err, context.Canceled) || calls != 1 {
				t.Fatalf("canceled checkpoint = %v, reads%d", err, calls)
			}
		} else if err != nil || calls != 2 {
			t.Fatalf("ordinary checkpoint = %v, reads%d", err, calls)
		}
	}
}

func TestSecurityDocumentRejectsDirectoryAndAcceptsRegularControl(t *testing.T) {
	directory := t.TempDir()
	if err := os.Mkdir(filepath.Join(directory, "directory"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "ordinary"), []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	root, err := os.OpenRoot(directory)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	if data, err := readSecurityDocument(root, "directory"); data != nil || !errors.Is(err, repositoryfile.ErrNotRegular) {
		t.Fatalf("directory = %v", err)
	}
	if data, err := readSecurityDocument(root, "ordinary"); err != nil || string(data) != "{}" {
		t.Fatalf("ordinary = %q, %v", data, err)
	}
}
