package gates

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFormattingAndSafetyReportMalformedOrUnreadableTrees(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "bad.go"), []byte("package ["), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := checkFormatting(root); err == nil || !strings.Contains(err.Error(), "format ") {
		t.Fatalf("checkFormatting() error = %v", err)
	}
	if err := checkSafety(root); err == nil || !strings.Contains(err.Error(), "parse ") {
		t.Fatalf("checkSafety() error = %v", err)
	}
	if err := checkFormatting(filepath.Join(root, "missing")); err == nil {
		t.Fatal("checkFormatting() missing root error = nil")
	}
}

func TestSafetyIgnoresTestOnlyUnsafeImports(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "unsafe_test.go"), []byte("package example\n\nimport \"unsafe\"\n\nvar _ = unsafe.Sizeof(0)\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := checkSafety(root); err != nil {
		t.Fatalf("checkSafety() error = %v", err)
	}
}

func TestWalkModuleFilesSkipsVendorAndGitDirectories(t *testing.T) {
	root := t.TempDir()
	for _, directory := range []string{"vendor", ".git"} {
		path := filepath.Join(root, directory)
		if err := os.Mkdir(path, 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(path, "bad.go"), []byte("package ["), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := checkFormatting(root); err != nil {
		t.Fatalf("checkFormatting() error = %v", err)
	}
}

func TestFormattingReportsUnreadableSource(t *testing.T) {
	root := t.TempDir()
	if err := os.Symlink("missing", filepath.Join(root, "missing.go")); err != nil {
		t.Fatal(err)
	}
	if err := checkFormatting(root); err == nil {
		t.Fatal("checkFormatting() error = nil")
	}
}

func TestWalkModuleFilesReportsNestedModuleInspectionFailure(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "nested")
	if err := os.Mkdir(nested, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("go.mod", filepath.Join(nested, "go.mod")); err != nil {
		t.Fatal(err)
	}
	if err := checkFormatting(root); err == nil {
		t.Fatal("checkFormatting() error = nil")
	}
}
