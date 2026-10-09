package mutation

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCaptureFailedDiagnosticRejectsUnsupportedContent(t *testing.T) {
	for _, test := range []struct {
		name   string
		change func(*report)
		raw    string
	}{
		{name: "unknown type", change: func(value *report) { value.Files[0].Mutations[0].Type = "UNRECOGNIZED" }},
		{name: "unknown status", change: func(value *report) { value.Files[0].Mutations[0].Status = "UNRECOGNIZED" }},
		{name: "unknown source", change: func(value *report) { value.Files[0].FileName = "other.go" }},
		{name: "test source", change: func(value *report) { value.Files[0].FileName = "source_test.go" }},
		{name: "absolute source", change: func(value *report) { value.Files[0].FileName = "/source.go" }},
		{name: "parent source", change: func(value *report) { value.Files[0].FileName = "../source.go" }},
		{name: "invalid coordinate", change: func(value *report) { value.Files[0].Mutations[0].Line = 0 }},
		{name: "zero column", change: func(value *report) { value.Files[0].Mutations[0].Column = 0 }},
		{name: "non-Go source", change: func(value *report) { value.Files[0].FileName = "ordinary.txt" }},
		{name: "absent files", change: func(value *report) { value.Files = nil }},
		{name: "malformed", raw: `{"files":`},
		{name: "duplicate key", raw: `{"files":[],"files":[]}`},
	} {
		t.Run(test.name, func(t *testing.T) {
			campaign, _ := campaignFixture(t)
			if err := campaign.prepareDirectories(); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(campaign.Root, "ordinary.txt"), []byte("ordinary"), 0o600); err != nil {
				t.Fatal(err)
			}
			value := report{Files: []reportFile{{FileName: "source.go", Mutations: []mutation{{Type: "INVERT_LOGICAL", Status: "LIVED", Line: 3, Column: 1}}}}}
			if test.change != nil {
				test.change(&value)
			}
			data, err := json.Marshal(value)
			if err != nil {
				t.Fatal(err)
			}
			if test.raw != "" {
				data = []byte(test.raw)
			}
			reportPath := filepath.Join(campaign.Workspace, "report.json")
			if err := os.WriteFile(reportPath, data, 0o600); err != nil {
				t.Fatal(err)
			}
			path, err := campaign.captureFailedDiagnostic(".", "sha256:"+strings.Repeat("a", 64), reportPath)
			if path != "" || !errors.Is(err, ErrInvalid) {
				t.Fatalf("capture() = %q, %v", path, err)
			}
			if _, err := os.Stat(filepath.Join(campaign.MutationRoot, "failed-reports")); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("refused diagnostic published: %v", err)
			}
		})
	}
}

func TestCaptureFailedDiagnosticRejectsMalformedInputIdentity(t *testing.T) {
	for _, input := range []string{strings.Repeat("a", 64), "sha256:bad"} {
		campaign, _ := campaignFixture(t)
		if err := campaign.prepareDirectories(); err != nil {
			t.Fatal(err)
		}
		reportPath := filepath.Join(campaign.Workspace, "report.json")
		data := `{"files":[{"file_name":"source.go","mutations":[{"type":"INVERT_LOGICAL","status":"LIVED","line":3,"column":1}]}]}`
		if err := os.WriteFile(reportPath, []byte(data), 0o600); err != nil {
			t.Fatal(err)
		}
		path, err := campaign.captureFailedDiagnostic(".", input, reportPath)
		if path != "" || !errors.Is(err, ErrInvalid) {
			t.Fatalf("malformed input capture = %q, %v; want refusal", path, err)
		}
		if _, err := os.Stat(filepath.Join(campaign.MutationRoot, "failed-reports")); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("malformed identity published a diagnostic: %v", err)
		}
	}
}

func TestEquivalentSelectionEqualUsesOwnerIdentity(t *testing.T) {
	base := ordinaryEquivalentSelection()
	if !equivalentSelectionEqual(nil, nil) {
		t.Fatal("absent selections differ")
	}
	if equivalentSelectionEqual(nil, &base) || equivalentSelectionEqual(&base, nil) {
		t.Fatal("present selection equals absent selection")
	}
	if !equivalentSelectionEqual(&base, &base) {
		t.Fatal("selection differs from itself")
	}
	for _, test := range []struct {
		name   string
		change func(*EquivalentReview)
	}{
		{"module", func(review *EquivalentReview) { review.ModuleDirectory = "nested" }},
		{"package", func(review *EquivalentReview) { review.PackageDirectory = "adapter" }},
		{"source", func(review *EquivalentReview) { review.SourceDigest = strings.Repeat("c", 64) }},
		{"version", func(review *EquivalentReview) { review.GremlinsVersion = "v0.7.0" }},
		{"verifier", func(review *EquivalentReview) { review.GremlinsVerifierSHA256 = strings.Repeat("d", 64) }},
	} {
		t.Run(test.name, func(t *testing.T) {
			changed := base
			test.change(&changed)
			if equivalentSelectionEqual(&base, &changed) || equivalentSelectionEqual(&changed, &base) {
				t.Fatal("different owner identities compare equal")
			}
		})
	}
}

func TestEquivalentSelectionEqualComparesCoordinateSelection(t *testing.T) {
	base := ordinaryEquivalentSelection()
	reordered := base
	reordered.Mutations = []EquivalentMutation{base.Mutations[1], base.Mutations[0]}
	if !equivalentSelectionEqual(&base, &reordered) || !equivalentSelectionEqual(&reordered, &base) {
		t.Fatal("ordering changed the selected coordinates")
	}
	reordered.Mutations[0].Reason = "A revised explanation does not change the selected coordinate or contract."
	if !equivalentSelectionEqual(&base, &reordered) {
		t.Fatal("explanation changed the selected coordinates")
	}
	for _, test := range []struct {
		name   string
		change func(*EquivalentReview)
	}{
		{"file", func(review *EquivalentReview) { review.Mutations[0].FileName = "other.go" }},
		{"type", func(review *EquivalentReview) { review.Mutations[0].Type = "comparison" }},
		{"line", func(review *EquivalentReview) { review.Mutations[0].Line++ }},
		{"column", func(review *EquivalentReview) { review.Mutations[0].Column++ }},
		{"contract", func(review *EquivalentReview) {
			review.Mutations[0].ContractDomain = "A different observable integer result contract"
		}},
		{"removed", func(review *EquivalentReview) { review.Mutations = review.Mutations[:1] }},
		{"added", func(review *EquivalentReview) {
			additional := review.Mutations[0]
			additional.Line = 9
			review.Mutations = append(review.Mutations, additional)
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			changed := base
			changed.Mutations = append([]EquivalentMutation(nil), base.Mutations...)
			test.change(&changed)
			if equivalentSelectionEqual(&base, &changed) || equivalentSelectionEqual(&changed, &base) {
				t.Fatal("different coordinate selections compare equal")
			}
		})
	}
}

func ordinaryEquivalentSelection() EquivalentReview {
	return EquivalentReview{
		ModuleDirectory: ".", PackageDirectory: ".",
		SourceDigest: strings.Repeat("a", 64), GremlinsVersion: "v0.6.0",
		GremlinsVerifierSHA256: strings.Repeat("b", 64),
		Mutations: []EquivalentMutation{
			{FileName: "value.go", Type: "arithmetic", Line: 3, Column: 1,
				ContractDomain: "Observable integer result for admitted values",
				Reason:         "The reviewed coordinate preserves the same integer result for admitted values."},
			{FileName: "value.go", Type: "comparison", Line: 5, Column: 2,
				ContractDomain: "Observable ordering for admitted integer values",
				Reason:         "The reviewed comparison preserves the same ordering for admitted integer values."},
		},
	}
}
