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
	loaded, err := evidence.Load(root, record.Gate, record.InputDigest)
	if err != nil || loaded.ReportDigest != record.ReportDigest {
		t.Fatalf("Load() = %#v, %v", loaded, err)
	}
}

func TestLoadRejectsInvalidAndMismatchedEvidence(t *testing.T) {
	digest := "sha256:" + strings.Repeat("a", 64)
	if _, err := evidence.Load("relative", "test", digest); !errors.Is(err, evidence.ErrInvalid) {
		t.Fatalf("Load(relative) error = %v", err)
	}
	if _, err := evidence.Load(t.TempDir(), "bad/gate", digest); !errors.Is(err, evidence.ErrInvalid) {
		t.Fatalf("Load(gate) error = %v", err)
	}
	if _, err := evidence.Load(t.TempDir(), "test", "bad"); !errors.Is(err, evidence.ErrInvalid) {
		t.Fatalf("Load(digest) error = %v", err)
	}
	root := t.TempDir()
	path := filepath.Join(root, "by-input", "test")
	if err := os.MkdirAll(path, 0o700); err != nil {
		t.Fatal(err)
	}
	if _, err := evidence.Load(root, "test", digest); err == nil {
		t.Fatal("Load(missing) error = nil")
	}
	file := filepath.Join(path, strings.Repeat("a", 64)+".json")
	if err := os.Mkdir(file, 0o700); err != nil {
		t.Fatal(err)
	}
	if _, err := evidence.Load(root, "test", digest); !errors.Is(err, evidence.ErrInvalid) {
		t.Fatalf("Load(directory) error = %v", err)
	} else if !strings.Contains(err.Error(), "not a regular file") {
		t.Fatalf("Load(directory) error = %v", err)
	}
}

func TestStoreRejectsConflictingEvidence(t *testing.T) {
	tests := map[string]func(*evidence.Record){
		"repository":  func(record *evidence.Record) { record.Repository = "github.com/faustbrian/other" },
		"module":      func(record *evidence.Record) { record.Module = "nested" },
		"package":     func(record *evidence.Record) { record.Package = "nested" },
		"verifier":    func(record *evidence.Record) { record.VerifierDigest = digest("b") },
		"result":      func(record *evidence.Record) { record.Result = "failed" },
		"report":      func(record *evidence.Record) { record.ReportDigest = digest("b") },
		"environment": func(record *evidence.Record) { record.Environment["go"] = "1.27.1" },
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			record := validRecord()
			if _, _, err := evidence.Store(root, record); err != nil {
				t.Fatal(err)
			}
			mutate(&record)
			if _, _, err := evidence.Store(root, record); !errors.Is(err, evidence.ErrConflict) {
				t.Fatalf("Store() error = %v", err)
			}
		})
	}
}

func TestParseRejectsInvalidRecords(t *testing.T) {
	valid, err := evidence.Marshal(validRecord())
	if err != nil {
		t.Fatal(err)
	}
	tests := map[string]struct {
		input []byte
		want  string
	}{
		"invalid":   {input: []byte("{}"), want: "schema_version"},
		"unknown":   {input: bytes.Replace(valid, []byte(`"gate"`), []byte(`"unknown":true,"gate"`), 1), want: "unknown field"},
		"multiple":  {input: append(append([]byte(nil), valid...), []byte("{}")...), want: "multiple JSON values"},
		"trailing":  {input: append(append([]byte(nil), valid...), '['), want: "trailing data"},
		"oversized": {input: append(append([]byte(nil), valid...), bytes.Repeat([]byte(" "), evidence.MaximumSize)...), want: "exceeds"},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := evidence.Parse(bytes.NewReader(test.input)); !errors.Is(err, evidence.ErrInvalid) || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Parse() error = %v", err)
			}
		})
	}
}

func TestParseAcceptsExactSizeLimit(t *testing.T) {
	valid, err := evidence.Marshal(validRecord())
	if err != nil {
		t.Fatal(err)
	}
	input := append([]byte(nil), valid...)
	input = append(input, bytes.Repeat([]byte(" "), evidence.MaximumSize-len(valid))...)
	if _, err := evidence.Parse(bytes.NewReader(input)); err != nil {
		t.Fatalf("Parse() error = %v", err)
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
	digest := digest("a")
	return evidence.Record{
		SchemaVersion: evidence.SchemaVersion, Repository: "github.com/faustbrian/example",
		Module: ".", Gate: "coverage", InputDigest: digest,
		VerifierDigest: digest, Result: "passed", ReportDigest: digest,
		Environment: map[string]string{"go": "1.27.0"}, CompletedAt: time.Unix(1, 0).UTC(),
	}
}

func digest(character string) string { return "sha256:" + strings.Repeat(character, 64) }
