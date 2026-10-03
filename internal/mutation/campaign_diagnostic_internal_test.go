package mutation

import (
	"strings"
	"testing"
)

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
