package mutation

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

type faultySourceFiles struct {
	base      sourceFileSystem
	operation string
}

func (files faultySourceFiles) Lstat(path string) (os.FileInfo, error) {
	if files.operation == "lstat" {
		return nil, errors.New("lstat failed")
	}
	return files.base.Lstat(path)
}

func (files faultySourceFiles) ReadDir(path string) ([]os.DirEntry, error) {
	if files.operation == "read-dir" {
		return nil, errors.New("read directory failed")
	}
	return files.base.ReadDir(path)
}

func (files faultySourceFiles) ReadFile(path string) ([]byte, error) {
	if files.operation == "read-file" {
		return nil, errors.New("read file failed")
	}
	return files.base.ReadFile(path)
}

func (files faultySourceFiles) Rel(base, target string) (string, error) {
	if files.operation == "relative" {
		return "", errors.New("relative failed")
	}
	return files.base.Rel(base, target)
}

func TestSourceDigestRejectsInvalidAndUnsafeInputs(t *testing.T) {
	root := t.TempDir()
	if _, err := SourceDigest("relative", ".", "."); !errors.Is(err, ErrInvalid) {
		t.Fatalf("SourceDigest(relative root) error = %v", err)
	}
	if _, err := SourceDigest(root, "../outside", "."); !errors.Is(err, ErrInvalid) {
		t.Fatalf("SourceDigest(escaping module) error = %v", err)
	}
	file := filepath.Join(root, "file")
	if err := os.WriteFile(file, []byte("file"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := SourceDigest(root, ".", "file"); !errors.Is(err, ErrInvalid) {
		t.Fatalf("SourceDigest(file package) error = %v", err)
	}
	if _, err := SourceDigest(root, ".", "missing"); err == nil {
		t.Fatal("SourceDigest(missing package) error = nil")
	}
}

func TestSourceDigestRejectsEmptyAndSymlinkedPackages(t *testing.T) {
	root := t.TempDir()
	if _, err := SourceDigest(root, ".", "."); !errors.Is(err, ErrInvalid) {
		t.Fatalf("SourceDigest(empty package) error = %v", err)
	}
	if err := os.Symlink("missing.go", filepath.Join(root, "linked.go")); err != nil {
		t.Fatal(err)
	}
	if _, err := SourceDigest(root, ".", "."); !errors.Is(err, ErrInvalid) {
		t.Fatalf("SourceDigest(symlink) error = %v", err)
	}
}

func TestSourceDigestReportsFileSystemFailures(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "source.go"), []byte("package source\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, operation := range []string{"lstat", "read-dir", "read-file", "relative"} {
		t.Run(operation, func(t *testing.T) {
			files := faultySourceFiles{base: operatingSourceFiles{}, operation: operation}
			if _, err := sourceDigest(files, root, ".", "."); err == nil {
				t.Fatalf("sourceDigest(%s) error = nil", operation)
			}
		})
	}
}
