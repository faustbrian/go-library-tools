package docscheck

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckAcceptsNavigableDocumentation(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "README.md"), "# Project\n\n[Guide](docs/guide.md \"Guide\") [Encoded](<docs/space%20name.md>) [Site](HTTPS://example.com) [Mail](mailto:test@example.com) [Here](#project) [Empty]()\n")
	write(t, filepath.Join(root, "SECURITY.md"), "# Security\n")
	write(t, filepath.Join(root, "docs", "guide.md"), "# Guide\n\n[Root](../README.md) [Encoded](space%20name.md)\n")
	write(t, filepath.Join(root, "docs", "space name.md"), "# Encoded\n")
	write(t, filepath.Join(root, "docs", "ignored.txt"), "ignored trailing whitespace \n")
	write(t, filepath.Join(root, "unrelated", "ignored.md"), "ignored trailing whitespace \n")
	if err := Check(root); err != nil {
		t.Fatalf("Check() error = %v", err)
	}
}

func TestCheckAcceptsExactDocumentBounds(t *testing.T) {
	root := basic(t)
	write(t, filepath.Join(root, "README.md"), strings.Repeat("x", maximumDocumentSize))
	for index := 1; index < maximumDocuments; index++ {
		write(t, filepath.Join(root, "docs", fmt.Sprintf("%04d.md", index)), "# Document\n")
	}
	if err := Check(root); err != nil {
		t.Fatalf("Check() error = %v", err)
	}
}

func TestCheckRejectsInvalidRootsAndDocuments(t *testing.T) {
	tests := []struct {
		name string
		make func(*testing.T) string
		want string
	}{
		{"relative", func(*testing.T) string { return "." }, "absolute"},
		{"missing README", func(t *testing.T) string { return t.TempDir() }, "README"},
		{"README directory", func(t *testing.T) string {
			root := t.TempDir()
			mustMkdir(t, filepath.Join(root, "README.md"))
			return root
		}, "regular file"},
		{"root unavailable", func(t *testing.T) string { return filepath.Join(t.TempDir(), "missing") }, "resolve documentation root"},
		{"docs symlink", func(t *testing.T) string {
			root := basic(t)
			if err := os.Symlink(root, filepath.Join(root, "docs")); err != nil {
				t.Fatal(err)
			}
			return root
		}, "symlink"},
		{"trailing whitespace", func(t *testing.T) string {
			root := basic(t)
			write(t, filepath.Join(root, "README.md"), "# Readme \n")
			return root
		}, "trailing whitespace"},
		{"oversized", func(t *testing.T) string {
			root := basic(t)
			write(t, filepath.Join(root, "README.md"), strings.Repeat("x", maximumDocumentSize+1))
			return root
		}, "size limit"},
		{"unreadable", func(t *testing.T) string {
			root := basic(t)
			path := filepath.Join(root, "SECRET.md")
			write(t, path, "# Secret\n")
			if err := os.Chmod(path, 0); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = os.Chmod(path, 0o600) })
			return root
		}, "read documentation"},
		{"unreadable directory", func(t *testing.T) string {
			root := basic(t)
			path := filepath.Join(root, "docs", "blocked")
			mustMkdir(t, path)
			if err := os.Chmod(path, 0); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = os.Chmod(path, 0o700) })
			return root
		}, "walk documentation"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := Check(test.make(t))
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Check() error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestCheckRejectsUnsafeAndBrokenLinks(t *testing.T) {
	for _, test := range []struct{ name, target, want string }{
		{"invalid URL", "%zz", "invalid link"},
		{"unclosed angle", "<missing.md", "not closed"},
		{"unsupported scheme", "file:secret", "unsupported link scheme"},
		{"network path", "//example.com/path", "repository-relative"},
		{"absolute", "/tmp/file", "repository-relative"},
		{"escape", "../outside.md", "escapes repository"},
		{"exact parent escape", "..", "escapes repository"},
		{"broken", "missing.md", "broken local link"},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := basic(t)
			write(t, filepath.Join(root, "README.md"), "[target]("+test.target+")\n")
			err := Check(root)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Check() error = %v, want %q", err, test.want)
			}
		})
	}

	t.Run("symlink target", func(t *testing.T) {
		root := basic(t)
		write(t, filepath.Join(root, "target.txt"), "target\n")
		if err := os.Symlink(filepath.Join(root, "target.txt"), filepath.Join(root, "link.txt")); err != nil {
			t.Fatal(err)
		}
		write(t, filepath.Join(root, "README.md"), "[target](link.txt)\n")
		if err := Check(root); err == nil || !strings.Contains(err.Error(), "targets symlink") {
			t.Fatalf("Check() error = %v", err)
		}
	})

	t.Run("checks every link and line number", func(t *testing.T) {
		root := basic(t)
		write(t, filepath.Join(root, "README.md"), "# Readme\n[valid](#readme) [broken](missing.md)\n")
		if err := Check(root); err == nil || !strings.Contains(err.Error(), "README.md:2") || !strings.Contains(err.Error(), "broken local link") {
			t.Fatalf("Check() error = %v", err)
		}
	})
}

func TestCheckBoundsDocumentCount(t *testing.T) {
	root := basic(t)
	for index := range maximumDocuments {
		write(t, filepath.Join(root, "docs", fmt.Sprintf("%04d.md", index)), "# Document\n")
	}
	if err := Check(root); err == nil || !strings.Contains(err.Error(), "file count") {
		t.Fatalf("Check() error = %v", err)
	}
}

func basic(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	write(t, filepath.Join(root, "README.md"), "# Readme\n")
	return root
}

func write(t *testing.T, path, content string) {
	t.Helper()
	mustMkdir(t, filepath.Dir(path))
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func mustMkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o700); err != nil {
		t.Fatal(err)
	}
}
