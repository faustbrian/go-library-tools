package evidence_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/faustbrian/go-library-tools/internal/evidence"
)

func TestInspectValidatesAndSortsContentAddressedRecords(t *testing.T) {
	root := t.TempDir()
	first := inspectRecord("b", "coverage", "a")
	second := inspectRecord("a", "mutation", "b")
	for index, record := range []evidence.Record{first, second} {
		storeRoot := filepath.Join(root, ".verification", string(rune('a'+index)), "evidence")
		if err := os.MkdirAll(filepath.Dir(storeRoot), 0o700); err != nil {
			t.Fatal(err)
		}
		if _, _, err := evidence.Store(storeRoot, record); err != nil {
			t.Fatal(err)
		}
	}
	writeInspect(t, filepath.Join(root, ".verification", "notes.txt"), "ignored")
	writeInspect(t, filepath.Join(root, ".verification", "metadata.json"), "{}")
	records, err := evidence.Inspect(root, ".verification", "example", []string{"a", "b"})
	if err != nil || len(records) != 2 || records[0].Module != "a" || records[1].Module != "b" {
		t.Fatalf("Inspect() = %#v, %v", records, err)
	}
}

func TestInspectAcceptsMissingEvidenceRoot(t *testing.T) {
	records, err := evidence.Inspect(t.TempDir(), ".verification", "example", []string{"."})
	if err != nil || len(records) != 0 {
		t.Fatalf("Inspect() = %#v, %v", records, err)
	}
}

func TestInspectRejectsUnsafeRootsAndTrees(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "target")
	if err := os.Mkdir(target, 0o700); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "link")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	if _, err := evidence.Inspect(root, "link", "example", nil); !errors.Is(err, evidence.ErrInvalid) {
		t.Fatalf("Inspect() root symlink error = %v", err)
	}
	verification := filepath.Join(root, ".verification")
	if err := os.Mkdir(verification, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(verification, "nested")); err != nil {
		t.Fatal(err)
	}
	if _, err := evidence.Inspect(root, ".verification", "example", nil); !errors.Is(err, evidence.ErrInvalid) {
		t.Fatalf("Inspect() tree symlink error = %v", err)
	}
}

func TestInspectRejectsMalformedOrMismatchedRecords(t *testing.T) {
	tests := map[string]func(*testing.T, string){
		"malformed path": func(t *testing.T, root string) {
			writeInspect(t, filepath.Join(root, ".verification", "by-input", "coverage", "extra", "record.json"), "{}")
		},
		"invalid record": func(t *testing.T, root string) {
			writeInspect(t, filepath.Join(root, ".verification", "by-input", "coverage", strings.Repeat("a", 64)+".json"), "{}")
		},
		"path mismatch": func(t *testing.T, root string) {
			storeInspectRecord(t, root, inspectRecord(".", "coverage", "a"), "mutation", "a")
		},
		"repository mismatch": func(t *testing.T, root string) {
			record := inspectRecord(".", "coverage", "a")
			record.Repository = "other"
			storeInspectRecord(t, root, record, "coverage", "a")
		},
		"unknown module": func(t *testing.T, root string) {
			storeInspectRecord(t, root, inspectRecord("unknown", "coverage", "a"), "coverage", "a")
		},
	}
	for name, setup := range tests {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			setup(t, root)
			if _, err := evidence.Inspect(root, ".verification", "example", []string{"."}); !errors.Is(err, evidence.ErrInvalid) {
				t.Fatalf("Inspect() error = %v", err)
			}
		})
	}
}

func TestInspectRejectsDuplicateIdentity(t *testing.T) {
	root := t.TempDir()
	record := inspectRecord(".", "coverage", "a")
	storeInspectRecord(t, root, record, "coverage", "a")
	data, err := evidence.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	name := strings.TrimPrefix(record.InputDigest, "sha256:") + ".json"
	writeInspect(t, filepath.Join(root, ".verification", "other", "by-input", "coverage", name), string(data))
	if _, err := evidence.Inspect(root, ".verification", "example", []string{"."}); !errors.Is(err, evidence.ErrInvalid) {
		t.Fatalf("Inspect() error = %v", err)
	}
}

func inspectRecord(module, gate, digestCharacter string) evidence.Record {
	digest := "sha256:" + strings.Repeat(digestCharacter, 64)
	return evidence.Record{SchemaVersion: 1, Repository: "example", Module: module, Gate: gate,
		InputDigest: digest, VerifierDigest: digest, Result: "passed", ReportDigest: digest,
		CompletedAt: time.Unix(1, 0).UTC()}
}

func storeInspectRecord(t *testing.T, root string, record evidence.Record, gate, digestCharacter string) {
	t.Helper()
	data, err := evidence.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	writeInspect(t, filepath.Join(root, ".verification", "by-input", gate, strings.Repeat(digestCharacter, 64)+".json"), string(data))
}

func writeInspect(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}
