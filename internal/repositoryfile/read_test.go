package repositoryfile_test

import (
	"errors"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/faustbrian/go-library-tools/v2/internal/repositoryfile"
)

func TestReadReturnsBoundedRegularFile(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "nested", "file.txt"), "content")
	got, err := repositoryfile.Read(root, "nested/file.txt", 7)
	if err != nil || string(got) != "content" {
		t.Fatalf("Read() = %q, %v", got, err)
	}
}

func TestReadZeroLimitAdmitsOnlyEmptyRegularFiles(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "empty"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := repositoryfile.Read(root, "empty", 0)
	if err != nil || len(got) != 0 {
		t.Fatalf("Read(empty, zero) = %q, %v; want empty success", got, err)
	}
	write(t, filepath.Join(root, "one"), "a")
	for _, test := range []struct {
		path  string
		cause error
	}{
		{"one", repositoryfile.ErrTooLarge},
		{"missing", os.ErrNotExist},
		{".", repositoryfile.ErrNotRegular},
	} {
		t.Run(test.path, func(t *testing.T) {
			got, err := repositoryfile.Read(root, test.path, 0)
			if got != nil || !errors.Is(err, test.cause) {
				t.Fatalf("Read(%q, zero) = %q, %v; want nil and %v", test.path, got, err, test.cause)
			}
		})
	}
}

func TestReadRejectsUnsupportedLimitsBeforeAcquisition(t *testing.T) {
	root := t.TempDir()
	for _, limit := range []int64{-1, math.MaxInt64} {
		got, err := repositoryfile.Read(root, "missing", limit)
		if got != nil || !errors.Is(err, repositoryfile.ErrTooLarge) || errors.Is(err, os.ErrNotExist) {
			t.Fatalf("Read(missing, %d) = %q, %v; want pre-acquisition refusal", limit, got, err)
		}
	}
}

func TestReadRejectsOverflowingLookAheadLimit(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "file"), "abc")
	got, err := repositoryfile.Read(root, "file", math.MaxInt64)
	if got != nil || !errors.Is(err, repositoryfile.ErrTooLarge) {
		t.Fatalf("Read(MaxInt64) = %q, %v; want nil bytes and ErrTooLarge", got, err)
	}
	for _, limit := range []int64{math.MaxInt64 - 1, 3} {
		got, err := repositoryfile.Read(root, "file", limit)
		if err != nil || string(got) != "abc" {
			t.Fatalf("Read(%d) = %q, %v; want exact fixture bytes", limit, got, err)
		}
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

func TestValidateDirectoryRejectsUnsafeAndNonDirectoryPaths(t *testing.T) {
	root := t.TempDir()
	directory := filepath.Join(root, "directory")
	if err := os.Mkdir(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(root, "file"), "content")
	if err := os.Symlink(directory, filepath.Join(root, "link")); err != nil {
		t.Fatal(err)
	}
	if err := repositoryfile.ValidateDirectory(root, "directory"); err != nil {
		t.Fatalf("ValidateDirectory() error = %v", err)
	}
	for name, path := range map[string]string{"file": "file", "symlink": "link"} {
		t.Run(name, func(t *testing.T) {
			if err := repositoryfile.ValidateDirectory(root, path); err == nil {
				t.Fatal("ValidateDirectory() error = nil")
			}
		})
	}
}

func TestValidateRegularFileRejectsUnsafeAndNonRegularPaths(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "nested", "file"), "content")
	outside := t.TempDir()
	write(t, filepath.Join(outside, "file"), "outside")
	if err := os.Symlink(outside, filepath.Join(root, "link")); err != nil {
		t.Fatal(err)
	}
	if err := repositoryfile.ValidateRegularFile(root, "nested/file"); err != nil {
		t.Fatalf("ValidateRegularFile() error = %v", err)
	}
	for name, path := range map[string]string{"directory": "nested", "symlinked parent": "link/file"} {
		t.Run(name, func(t *testing.T) {
			if err := repositoryfile.ValidateRegularFile(root, path); err == nil {
				t.Fatal("ValidateRegularFile() error = nil")
			}
		})
	}
}

func TestInspectRegularFileReturnsValidatedMetadata(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "nested", "file")
	write(t, path, "abc")
	want, err := os.Lstat(path)
	if err != nil {
		t.Fatal(err)
	}
	info, err := repositoryfile.InspectRegularFile(root, "nested/file")
	if err != nil || info == nil {
		t.Fatalf("InspectRegularFile() = %v, %v; want admitted metadata", info, err)
	}
	if info.Name() != "file" || info.Size() != 3 || !info.Mode().IsRegular() || !os.SameFile(info, want) {
		t.Fatalf("InspectRegularFile() metadata = %v; want the regular three-byte fixture", info)
	}
	if err := repositoryfile.ValidateRegularFile(root, "nested/file"); err != nil {
		t.Fatalf("ValidateRegularFile() changed admitted-file result: %v", err)
	}
}

func TestInspectRegularFilePreservesValidationFailures(t *testing.T) {
	root := t.TempDir()
	for _, test := range []struct {
		name string
		path string
		is   error
	}{
		{"missing file", "missing", os.ErrNotExist},
		{"directory", ".", repositoryfile.ErrNotRegular},
	} {
		t.Run(test.name, func(t *testing.T) {
			info, err := repositoryfile.InspectRegularFile(root, test.path)
			if info != nil || !errors.Is(err, test.is) {
				t.Fatalf("InspectRegularFile(%q) = %v, %v; want nil metadata and %v", test.path, info, err, test.is)
			}
			validationErr := repositoryfile.ValidateRegularFile(root, test.path)
			if !errors.Is(validationErr, test.is) || validationErr.Error() != err.Error() {
				t.Fatalf("ValidateRegularFile(%q) = %v; want unchanged inspection diagnostic %v", test.path, validationErr, err)
			}
			if test.path == "missing" {
				var pathErr *os.PathError
				if !errors.As(err, &pathErr) || pathErr.Op == "" || pathErr.Path != filepath.Join(root, test.path) || !strings.HasPrefix(err.Error(), "inspect missing: ") {
					t.Fatalf("missing inspection error = %v; want wrapped fixture lstat failure", err)
				}
			} else if err.Error() != "repository file is not regular: ." {
				t.Fatalf("directory diagnostic = %q; want existing regular-file refusal", err.Error())
			}
		})
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
