package evidence_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/faustbrian/go-library-tools/internal/evidence"
)

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) { return 0, errors.New("read failed") }

func TestDigestIsDeterministicAndBoundarySafe(t *testing.T) {
	first := evidence.Digest("verifier:v1", map[string][]byte{"b": []byte("c"), "a": []byte("bc")})
	second := evidence.Digest("verifier:v1", map[string][]byte{"a": []byte("bc"), "b": []byte("c")})
	ambiguous := evidence.Digest("verifier:v1", map[string][]byte{"ab": []byte("c")})
	if first != second || first == ambiguous || !strings.HasPrefix(first, "sha256:") {
		t.Fatalf("digests = %q, %q, %q", first, second, ambiguous)
	}
}

func TestMarshalParseAndStoreRoundTrip(t *testing.T) {
	record := validRecord()
	data, err := evidence.Marshal(record)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	got, err := evidence.Parse(bytes.NewReader(data))
	if err != nil || got.Repository != record.Repository {
		t.Fatalf("Parse() = %#v, %v", got, err)
	}
	root := t.TempDir()
	path, reused, err := evidence.Store(root, record)
	if err != nil || reused {
		t.Fatalf("Store() = %q, %v, %v", path, reused, err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
	secondPath, reused, err := evidence.Store(root, record)
	if err != nil || !reused || secondPath != path {
		t.Fatalf("Store() reuse = %q, %v, %v", secondPath, reused, err)
	}
	record.CompletedAt = record.CompletedAt.Add(time.Hour)
	if _, reused, err := evidence.Store(root, record); err != nil || !reused {
		t.Fatalf("Store() semantic reuse = %v, %v", reused, err)
	}
}

func TestStoreRejectsConflictingEvidence(t *testing.T) {
	root := t.TempDir()
	record := validRecord()
	if _, _, err := evidence.Store(root, record); err != nil {
		t.Fatal(err)
	}
	record.Result = "failed"
	if _, _, err := evidence.Store(root, record); !errors.Is(err, evidence.ErrConflict) {
		t.Fatalf("Store() error = %v", err)
	}
}

func TestParseRejectsInvalidRecords(t *testing.T) {
	valid, err := evidence.Marshal(validRecord())
	if err != nil {
		t.Fatal(err)
	}
	tests := map[string][]byte{
		"invalid":   []byte("{}"),
		"unknown":   bytes.Replace(valid, []byte(`"gate"`), []byte(`"unknown":true,"gate"`), 1),
		"multiple":  append(append([]byte(nil), valid...), []byte("{}")...),
		"trailing":  append(append([]byte(nil), valid...), '['),
		"oversized": append(append([]byte(nil), valid...), bytes.Repeat([]byte(" "), evidence.MaximumSize)...),
	}
	for name, input := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := evidence.Parse(bytes.NewReader(input)); !errors.Is(err, evidence.ErrInvalid) {
				t.Fatalf("Parse() error = %v", err)
			}
		})
	}
}

func TestParseReportsReaderFailure(t *testing.T) {
	if _, err := evidence.Parse(failingReader{}); !errors.Is(err, evidence.ErrInvalid) {
		t.Fatalf("Parse() error = %v", err)
	}
}

func TestValidateRejectsEveryInvalidField(t *testing.T) {
	tests := map[string]func(*evidence.Record){
		"schema":      func(record *evidence.Record) { record.SchemaVersion = 2 },
		"repository":  func(record *evidence.Record) { record.Repository = "" },
		"module":      func(record *evidence.Record) { record.Module = "" },
		"gate":        func(record *evidence.Record) { record.Gate = "Bad/Gate" },
		"input":       func(record *evidence.Record) { record.InputDigest = "bad" },
		"verifier":    func(record *evidence.Record) { record.VerifierDigest = "bad" },
		"report":      func(record *evidence.Record) { record.ReportDigest = "bad" },
		"result":      func(record *evidence.Record) { record.Result = "unknown" },
		"completed":   func(record *evidence.Record) { record.CompletedAt = time.Time{} },
		"environment": func(record *evidence.Record) { record.Environment[" "] = "value" },
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			record := validRecord()
			mutate(&record)
			if err := record.Validate(); !errors.Is(err, evidence.ErrInvalid) {
				t.Fatalf("Validate() error = %v", err)
			}
		})
	}
}

func TestMarshalRejectsInvalidRecord(t *testing.T) {
	record := validRecord()
	record.Gate = "invalid/gate"
	if _, err := evidence.Marshal(record); !errors.Is(err, evidence.ErrInvalid) {
		t.Fatalf("Marshal() error = %v", err)
	}
}

func TestStoreReportsFilesystemFailures(t *testing.T) {
	record := validRecord()
	file := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(file, []byte("occupied"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := evidence.Store(file, record); err == nil {
		t.Fatal("Store() error = nil")
	}
	record.InputDigest = "bad"
	if _, _, err := evidence.Store(t.TempDir(), record); !errors.Is(err, evidence.ErrInvalid) {
		t.Fatalf("Store() validation error = %v", err)
	}
}

func TestStoreRejectsUnsafeEvidenceRoots(t *testing.T) {
	record := validRecord()
	if _, _, err := evidence.Store("relative", record); !errors.Is(err, evidence.ErrInvalid) {
		t.Fatalf("Store() relative error = %v", err)
	}
	parent := t.TempDir()
	target := filepath.Join(parent, "target")
	if err := os.Mkdir(target, 0o700); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(parent, "link")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	if _, _, err := evidence.Store(link, record); !errors.Is(err, evidence.ErrInvalid) {
		t.Fatalf("Store() symlink error = %v", err)
	}
	root := filepath.Join(parent, "evidence")
	if err := os.Mkdir(root, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(root, "by-input")); err != nil {
		t.Fatal(err)
	}
	if _, _, err := evidence.Store(root, record); !errors.Is(err, evidence.ErrInvalid) {
		t.Fatalf("Store() nested symlink error = %v", err)
	}
}

func validRecord() evidence.Record {
	digest := "sha256:" + strings.Repeat("a", 64)
	return evidence.Record{
		SchemaVersion: evidence.SchemaVersion, Repository: "github.com/faustbrian/example",
		Module: ".", Gate: "coverage", InputDigest: digest,
		VerifierDigest: digest, Result: "passed", ReportDigest: digest,
		Environment: map[string]string{"go": "1.27.0"}, CompletedAt: time.Unix(1, 0).UTC(),
	}
}
