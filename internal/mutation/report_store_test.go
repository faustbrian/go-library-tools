package mutation_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/faustbrian/go-library-tools/internal/mutation"
)

func TestStoreAndLoadReportByInputIdentity(t *testing.T) {
	root := filepath.Join(t.TempDir(), "mutation")
	input := "sha256:" + strings.Repeat("a", 64)
	report := []byte(`{"files":[{"file_name":"source.go","mutations":[{"type":"ARITHMETIC_BASE","status":"KILLED","line":2,"column":3}]}]}`)
	path, reused, result, err := mutation.StoreReport(root, input, report)
	if err != nil || reused || result.Mutants != 1 {
		t.Fatalf("StoreReport() = %q, %v, %#v, %v", path, reused, result, err)
	}
	secondPath, reused, secondResult, err := mutation.StoreReport(root, input, report)
	if err != nil || !reused || secondPath != path || secondResult.Digest != result.Digest {
		t.Fatalf("StoreReport(reuse) = %q, %v, %#v, %v", secondPath, reused, secondResult, err)
	}
	data, loaded, err := mutation.LoadReport(root, input)
	if err != nil || string(data) != string(report) || loaded.Digest != result.Digest {
		t.Fatalf("LoadReport() = %q, %#v, %v", data, loaded, err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
}

func TestStoreReportRejectsConflictingContent(t *testing.T) {
	root := filepath.Join(t.TempDir(), "mutation")
	input := "sha256:" + strings.Repeat("b", 64)
	first := []byte(`{"files":[{"file_name":"a.go","mutations":[{"type":"A","status":"KILLED","line":1,"column":1}]}]}`)
	second := []byte(`{"files":[{"file_name":"b.go","mutations":[{"type":"B","status":"KILLED","line":1,"column":1}]}]}`)
	if _, _, _, err := mutation.StoreReport(root, input, first); err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := mutation.StoreReport(root, input, second); !errors.Is(err, mutation.ErrInvalid) {
		t.Fatalf("StoreReport(conflict) error = %v", err)
	}
}
