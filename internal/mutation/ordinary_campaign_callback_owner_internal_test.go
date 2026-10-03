package mutation

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func TestOrdinaryRunPackageCallbackAdmission(t *testing.T) {
	for _, mode := range []string{"stale source", "caller cancellation", "changed selection"} {
		t.Run(mode, func(t *testing.T) {
			campaign, _ := campaignFixture(t)
			if err := campaign.prepareDirectories(); err != nil {
				t.Fatal(err)
			}
			source, err := SourceDigest(campaign.Root, ".", ".")
			if err != nil {
				t.Fatal(err)
			}
			campaign.EquivalentReviews = EquivalentInventory{SchemaVersion: 1, Packages: []EquivalentReview{{
				ModuleDirectory: ".", PackageDirectory: ".", SourceDigest: source,
				GremlinsVersion: GremlinsVersion, GremlinsVerifierSHA256: LegacyVerifierDigest(),
				Mutations: []EquivalentMutation{{FileName: "source.go", Type: "A", Line: 3, Column: 1,
					ContractDomain: "Observable integer result for admitted nonnegative values",
					Reason:         "Both boundary forms return the same maximum integer at equality."}},
			}}}
			if mode == "stale source" {
				first := "0"
				if source[:1] == first {
					first = "1"
				}
				campaign.EquivalentReviews.Packages[0].SourceDigest = first + source[1:]
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			state := campaignState{toolBuilt: true, tool: Tool{Path: "ordinary-result-callback"}, coverageProfile: "already-owned-profile", coverageElapsed: "1s"}
			listingCalls, resultCalls := 0, 0
			campaign.Process = func(_ context.Context, name string, args []string, directory string, _ map[string]string, stdout, _ io.Writer) error {
				switch {
				case name == "go" && len(args) > 0 && args[0] == "list":
					listingCalls++
					_, writeErr := fmt.Fprintf(stdout, `{"Dir":%q,"ImportPath":"example","GoFiles":["source.go"],"Module":{"Path":"example","Main":true,"GoVersion":"1.27.0"}}`, directory)
					return writeErr
				case name == state.tool.Path:
					resultCalls++
					if mode == "caller cancellation" {
						cancel()
						return nil // No report exists: cancellation must precede its read.
					}
					const report = `{"files":[{"file_name":"source.go","mutations":[{"type":"A","status":"LIVED","line":3,"column":1}]}],"mutants_killed":0,"mutants_lived":1,"mutants_not_covered":0,"mutants_not_viable":0,"mutants_total":1,"mutations_coverage":100,"test_efficacy":0}`
					for index, argument := range args {
						if argument == "--output" && index+1 < len(args) {
							if err := os.WriteFile(args[index+1], []byte(report), 0o600); err != nil {
								return err
							}
							campaign.EquivalentReviews.Packages[0].Mutations[0].ContractDomain = "Observable integer result for admitted positive values"
							return nil
						}
					}
					t.Fatal("result callback lacks its owned report destination")
				default:
					t.Fatal("unexpected callback; no executable or shared coverage is permitted")
				}
				return nil
			}
			var output bytes.Buffer
			err = campaign.runPackage(ctx, &output, ".", &state)
			wantError := "invalid mutation evidence: equivalent-mutant source or verifier identity changed"
			wantListing, wantResult := 1, 0
			wantSentinel := ErrInvalid
			switch mode {
			case "caller cancellation":
				wantError, wantResult, wantSentinel = "context canceled", 1, context.Canceled
			case "changed selection":
				wantError = "invalid mutation evidence: equivalent-mutant selection changed while running ."
				wantListing, wantResult = 2, 1
			}
			if !errors.Is(err, wantSentinel) || err.Error() != wantError {
				t.Fatalf("admission = %v; want %s", err, wantError)
			}
			if listingCalls != wantListing || resultCalls != wantResult || output.Len() != 0 {
				t.Fatalf("callbacks list=%d result=%d, output=%q; want %d/%d and no publication", listingCalls, resultCalls, output.String(), wantListing, wantResult)
			}
			if err := filepath.WalkDir(campaign.EvidenceRoot, func(_ string, entry fs.DirEntry, walkErr error) error {
				if walkErr != nil {
					return walkErr
				}
				if !entry.IsDir() {
					t.Error("rejected admission published report or evidence")
				}
				return nil
			}); err != nil {
				t.Fatal(err)
			}
		})
	}
}
