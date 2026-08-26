package repository

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckLegacyReportsInspectionFailure(t *testing.T) {
	failure := errors.New("injected failure")
	err := checkLegacy("/repo", func(string) (os.FileInfo, error) { return nil, failure })
	if err == nil || !strings.Contains(err.Error(), "inspect legacy") {
		t.Fatalf("checkLegacy() error = %v", err)
	}
}

func TestCheckWorkspaceAcceptsOneNestedModule(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.work"), []byte("go 1.27.0\nuse ./nested\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := checkWorkspace(root, "1.27.0", []string{"nested"}); err != nil {
		t.Fatal(err)
	}
}
