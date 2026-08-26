package repositoryfile_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/faustbrian/go-library-tools/internal/repositoryfile"
)

func TestReadReturnsBoundedRegularFile(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "nested", "file.txt"), "content")
	got, err := repositoryfile.Read(root, "nested/file.txt", 7)
	if err != nil || string(got) != "content" {
		t.Fatalf("Read() = %q, %v", got, err)
	}
}

func TestReadRejectsUnsafeAndInvalidFiles(t *testing.T) {
	root := t.TempDir()
	outside := filepath.Join(t.TempDir(), "outside")
	write(t, outside, "secret")
	write(t, filepath.Join(root, "large"), "12345")
	if err := os.Symlink(outside, filepath.Join(root, "link")); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, "directory"), 0o700); err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(root, "component"), "file")

	tests := []struct {
		name string
		path string
		max  int64
		is   error
	}{
		{"absolute", outside, 100, repositoryfile.ErrUnsafePath},
		{"parent", "../outside", 100, repositoryfile.ErrUnsafePath},
		{"symlink", "link", 100, repositoryfile.ErrUnsafePath},
		{"directory", "directory", 100, repositoryfile.ErrNotRegular},
		{"file parent", "component/child", 100, repositoryfile.ErrNotRegular},
		{"oversized", "large", 4, repositoryfile.ErrTooLarge},
		{"invalid limit", "large", 0, repositoryfile.ErrTooLarge},
		{"missing", "missing", 100, os.ErrNotExist},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := repositoryfile.Read(root, test.path, test.max)
			if !errors.Is(err, test.is) {
				t.Fatalf("Read() error = %v, want %v", err, test.is)
			}
		})
	}
}

func TestReadRejectsSymlinkedParent(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	write(t, filepath.Join(outside, "file"), "secret")
	if err := os.Symlink(outside, filepath.Join(root, "nested")); err != nil {
		t.Fatal(err)
	}
	_, err := repositoryfile.Read(root, "nested/file", 100)
	if !errors.Is(err, repositoryfile.ErrUnsafePath) {
		t.Fatalf("Read() error = %v", err)
	}
}

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(content) == "" {
		t.Fatal("test fixture must not be empty")
	}
}
