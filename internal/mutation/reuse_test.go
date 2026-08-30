package mutation_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/faustbrian/go-library-tools/internal/evidence"
	"github.com/faustbrian/go-library-tools/internal/mutation"
)

func TestReuseValidatesExactPackageEvidence(t *testing.T) {
	repositoryRoot := t.TempDir()
	evidenceRoot := filepath.Join(repositoryRoot, ".verification")
	mutationRoot := filepath.Join(evidenceRoot, "mutation")
	if err := os.MkdirAll(mutationRoot, 0o700); err != nil {
		t.Fatal(err)
	}
	input := "sha256:" + strings.Repeat("a", 64)
	report := []byte(`{"files":[{"file_name":"source.go","mutations":[{"type":"A","status":"KILLED","line":1,"column":1}]}]}`)
	_, _, result, err := mutation.StoreReport(mutationRoot, input, report)
	if err != nil {
		t.Fatal(err)
	}
	record := evidence.Record{
		SchemaVersion: evidence.SchemaVersion, Repository: "example", Module: ".", Package: "adapter",
		Gate: "mutation", InputDigest: input, VerifierDigest: mutation.SemanticVerifierDigest(),
		Result: "passed", ReportDigest: result.Digest, CompletedAt: time.Unix(1, 0).UTC(),
	}
	if _, _, err := evidence.Store(evidenceRoot, record); err != nil {
		t.Fatal(err)
	}
	reused, loaded, err := mutation.Reuse(evidenceRoot, mutationRoot, "example", ".", "adapter", input)
	if err != nil || !reused || loaded.Mutants != 1 {
		t.Fatalf("Reuse() = %v, %#v, %v", reused, loaded, err)
	}
}

func TestReuseAcceptsTheExactLegacyReportDigest(t *testing.T) {
	root := t.TempDir()
	evidenceRoot := filepath.Join(root, ".verification")
	mutationRoot := filepath.Join(evidenceRoot, "mutation")
	if err := os.MkdirAll(mutationRoot, 0o700); err != nil {
		t.Fatal(err)
	}
	input := "sha256:" + strings.Repeat("9", 64)
	report := []byte(`{"elapsed_time":1,"files":[]}`)
	if _, _, _, err := mutation.StoreReport(mutationRoot, input, report); err != nil {
		t.Fatal(err)
	}
	record := evidence.Record{
		SchemaVersion: evidence.SchemaVersion, Repository: "example", Module: ".", Package: ".",
		Gate: "mutation", InputDigest: input, VerifierDigest: mutation.SemanticVerifierDigest(), Result: "passed",
		ReportDigest: "sha256:e8091387ea1e4cc0805a2b33c6f6897feaa858a7b1d411233762dbc2a9f7c844",
		CompletedAt:  time.Unix(1, 0).UTC(),
	}
	if _, _, err := evidence.Store(evidenceRoot, record); err != nil {
		t.Fatal(err)
	}
	reused, result, err := mutation.Reuse(evidenceRoot, mutationRoot, "example", ".", ".", input)
	if err != nil || !reused || result.Mutants != 0 {
		t.Fatalf("Reuse(legacy digest) = %v, %#v, %v", reused, result, err)
	}
}

func TestReuseTreatsOnlyAbsentEvidenceAsMiss(t *testing.T) {
	root := t.TempDir()
	evidenceRoot := filepath.Join(root, ".verification")
	if err := os.Mkdir(evidenceRoot, 0o700); err != nil {
		t.Fatal(err)
	}
	digest := "sha256:" + strings.Repeat("b", 64)
	reused, _, err := mutation.Reuse(evidenceRoot, filepath.Join(evidenceRoot, "mutation"), "example", ".", ".", digest)
	if err != nil || reused {
		t.Fatalf("Reuse(missing) = %v, %v", reused, err)
	}
}

func TestReuseRejectsMismatchedEvidence(t *testing.T) {
	mutations := map[string]func(*evidence.Record){
		"repository": func(record *evidence.Record) { record.Repository = "other" },
		"module":     func(record *evidence.Record) { record.Module = "other" },
		"package":    func(record *evidence.Record) { record.Package = "other" },
		"result":     func(record *evidence.Record) { record.Result = "failed" },
		"verifier":   func(record *evidence.Record) { record.VerifierDigest = "sha256:" + strings.Repeat("f", 64) },
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			evidenceRoot := filepath.Join(root, ".verification")
			mutationRoot := filepath.Join(evidenceRoot, "mutation")
			if err := os.MkdirAll(mutationRoot, 0o700); err != nil {
				t.Fatal(err)
			}
			digest := "sha256:" + strings.Repeat("c", 64)
			_, _, result, err := mutation.StoreReport(mutationRoot, digest, []byte(`{"files":[]}`))
			if err != nil {
				t.Fatal(err)
			}
			record := evidence.Record{
				SchemaVersion: evidence.SchemaVersion, Repository: "example", Module: ".", Package: ".", Gate: "mutation",
				InputDigest: digest, VerifierDigest: mutation.SemanticVerifierDigest(), Result: "passed",
				ReportDigest: result.Digest, CompletedAt: time.Unix(1, 0).UTC(),
			}
			mutate(&record)
			if _, _, err := evidence.Store(evidenceRoot, record); err != nil {
				t.Fatal(err)
			}
			if _, _, err := mutation.Reuse(evidenceRoot, mutationRoot, "example", ".", ".", digest); !errors.Is(err, mutation.ErrInvalid) {
				t.Fatalf("Reuse(mismatch) error = %v", err)
			}
		})
	}
}

func TestReuseRejectsCorruptAndIncompleteEvidence(t *testing.T) {
	digest := "sha256:" + strings.Repeat("e", 64)
	t.Run("corrupt record", func(t *testing.T) {
		root := t.TempDir()
		evidenceRoot := filepath.Join(root, ".verification")
		directory := filepath.Join(evidenceRoot, "by-input", "mutation")
		if err := os.MkdirAll(directory, 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(directory, strings.Repeat("e", 64)+".json"), []byte(`{}`), 0o600); err != nil {
			t.Fatal(err)
		}
		if _, _, err := mutation.Reuse(evidenceRoot, filepath.Join(evidenceRoot, "mutation"), "example", ".", ".", digest); err == nil {
			t.Fatal("Reuse(corrupt record) error = nil")
		}
	})
	for _, test := range []struct {
		name        string
		storeReport bool
	}{
		{"missing report", false},
		{"mismatched report", true},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			evidenceRoot := filepath.Join(root, ".verification")
			mutationRoot := filepath.Join(evidenceRoot, "mutation")
			if err := os.MkdirAll(mutationRoot, 0o700); err != nil {
				t.Fatal(err)
			}
			reportDigest := "sha256:" + strings.Repeat("f", 64)
			if test.storeReport {
				_, _, _, err := mutation.StoreReport(mutationRoot, digest, []byte(`{"files":[]}`))
				if err != nil {
					t.Fatal(err)
				}
			}
			record := evidence.Record{
				SchemaVersion: evidence.SchemaVersion, Repository: "example", Module: ".", Package: ".",
				Gate: "mutation", InputDigest: digest, VerifierDigest: mutation.SemanticVerifierDigest(),
				Result: "passed", ReportDigest: reportDigest, CompletedAt: time.Unix(1, 0).UTC(),
			}
			if _, _, err := evidence.Store(evidenceRoot, record); err != nil {
				t.Fatal(err)
			}
			if _, _, err := mutation.Reuse(evidenceRoot, mutationRoot, "example", ".", ".", digest); err == nil {
				t.Fatalf("Reuse(%s) error = nil", test.name)
			}
		})
	}
}
