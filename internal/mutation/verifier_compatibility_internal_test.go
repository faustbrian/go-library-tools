package mutation

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/faustbrian/go-library-tools/v2/internal/evidence"
)

func TestHistoricalVerifierImportPreservesProvenanceOnlyWhenPackagePathUnchanged(t *testing.T) {
	for _, verifier := range []string{
		"9a9499ff68a8dfd49a0be7995590297a8ee563a1aa226bfd8b9361dc53058108",
		"5eba124f842305aecef71fc439aea0f1056429b253fad3b2ed47468f614595a3",
	} {
		for _, declaration := range []string{"adapter", "temporalwire"} {
			t.Run(verifier[:8]+"/"+declaration, func(t *testing.T) {
				campaign, _ := campaignFixture(t)
				campaign.Policy.Packages = []string{"adapter"}
				if err := os.Mkdir(filepath.Join(campaign.Root, "adapter"), 0o700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(campaign.Root, "adapter", "adapter.go"), []byte("package "+declaration+"\nfunc Value() int { return 1 }\n"), 0o600); err != nil {
					t.Fatal(err)
				}
				inputs, err := campaign.packageInputsForVerifiers(context.Background(), "adapter", LegacyVerifierDigest(), verifier)
				if err != nil {
					t.Fatal(err)
				}
				current := inputs[LegacyVerifierDigest()].current
				checkpoint, ledger := approvedImportFixture(t, inputs[verifier].current)
				checkpoint.Package, checkpoint.VerifierDigest = "adapter", verifier
				ledger.VerifierMigrationReview.GremlinsVerifierSHA256 = verifier
				ledger.VerifierMigrations[0].Package, ledger.VerifierMigrations[0].GremlinsVerifierSHA256 = "adapter", verifier
				ledger.Entries[0].Package, ledger.Entries[0].GremlinsVerifierSHA256 = "adapter", verifier
				err = campaign.Import(context.Background(), []Checkpoint{checkpoint}, ledger)
				if declaration == "temporalwire" {
					if !errors.Is(err, ErrInputChanged) {
						t.Fatalf("affected package Import() = %v, want fresh-execution requirement", err)
					}
					if _, err := evidence.Load(campaign.EvidenceRoot, "mutation", current); !errors.Is(err, os.ErrNotExist) {
						t.Fatalf("affected historical report became current evidence: %v", err)
					}
					return
				}
				if err != nil {
					t.Fatal(err)
				}
				record, err := evidence.Load(campaign.EvidenceRoot, "mutation", current)
				if err != nil {
					t.Fatal(err)
				}
				if record.VerifierDigest != "sha256:"+verifier {
					t.Fatalf("historical execution relabeled as %s", record.VerifierDigest)
				}
				if reused, _, err := Reuse(campaign.EvidenceRoot, campaign.MutationRoot, "example", ".", "adapter", current); err != nil || !reused {
					t.Fatalf("unchanged package Reuse() = %v, %v", reused, err)
				}
			})
		}
	}
}

func TestHistoricalVerifierReuseRequiresExactApplicabilityRecord(t *testing.T) {
	input := "sha256:" + strings.Repeat("a", 64)
	for name, change := range map[string]func(*evidence.Record){
		"missing proof": func(record *evidence.Record) { delete(record.Environment, legacyPackagePathCompatibility) },
		"foreign input": func(record *evidence.Record) {
			record.Environment[legacyPackagePathCompatibility] = "sha256:" + strings.Repeat("b", 64)
		},
		"wrong origin":     func(record *evidence.Record) { record.Environment["evidence_origin"] = "native" },
		"unknown verifier": func(record *evidence.Record) { record.VerifierDigest = "sha256:" + strings.Repeat("f", 64) },
	} {
		t.Run(name, func(t *testing.T) {
			record := evidence.Record{InputDigest: input, VerifierDigest: "sha256:" + publishedVerifierDigestV2,
				Environment: map[string]string{"evidence_origin": "approved_legacy_checkpoint", legacyPackagePathCompatibility: input}}
			if !applicableVerifierRecord(record) {
				t.Fatal("exact applicable historical record was rejected")
			}
			change(&record)
			if applicableVerifierRecord(record) {
				t.Fatal("historical verifier accepted without exact applicability proof")
			}
		})
	}
}

func TestHistoricalPackagePathProofUsesLegacyResolutionForEveryProductionFile(t *testing.T) {
	for _, test := range []struct {
		name, moduleDirectory, packageDirectory, declaration string
		wantCompatible                                       bool
	}{
		{"root alias", ".", ".", "different", true},
		{"matching directory", ".", "adapters/wire", "wire", true},
		{"matching suffix", ".", "adapters/wire", "ire", true},
		{"wire alias", ".", "adapters/wire", "temporalwire", false},
		{"ancestor name", ".", "adapters/wire", "adapters", false},
		{"nested module root", "nested", ".", "temporalwire", true},
		{"nested module alias", "nested", "wire", "temporalwire", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			directory := filepath.Join(root, test.moduleDirectory, test.packageDirectory)
			if err := os.MkdirAll(directory, 0o700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(directory, "source.go"), []byte("package "+test.declaration+"\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			source, err := SourceDigest(root, test.moduleDirectory, test.packageDirectory)
			if err != nil {
				t.Fatal(err)
			}
			err = legacyPackagePathsUnchanged(root, test.moduleDirectory, test.packageDirectory, "example/v2", source)
			if (err == nil) != test.wantCompatible {
				t.Fatalf("package compatibility = %v, want compatible=%v", err, test.wantCompatible)
			}
			if test.wantCompatible && test.packageDirectory != "." {
				if err := os.WriteFile(filepath.Join(directory, "other.go"), []byte("//go:build alternate\n\npackage different\n"), 0o600); err != nil {
					t.Fatal(err)
				}
				if err := legacyPackagePathsUnchanged(root, test.moduleDirectory, test.packageDirectory, "example/v2", source); !errors.Is(err, ErrInputChanged) {
					t.Fatalf("different declaration outside checkpoint report = %v", err)
				}
			}
		})
	}
}

func TestHistoricalPackagePathProofRejectsChangedSource(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "source.go")
	if err := os.WriteFile(path, []byte("package example\nfunc Value() int { return 1 }\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	source, err := SourceDigest(root, ".", ".")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("package example\nfunc Value() int { return 2 }\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := legacyPackagePathsUnchanged(root, ".", ".", "example", source); !errors.Is(err, ErrInputChanged) {
		t.Fatalf("changed source compatibility proof = %v", err)
	}
}
