package mutation_test

import (
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/faustbrian/go-library-tools/v2/internal/mutation"
)

func TestEquivalentInventoryRejectsOrdinaryInputFailuresWithoutPartialPolicy(t *testing.T) {
	valid, err := mutation.ParseEquivalentInventory(strings.NewReader(reviewedEquivalent))
	if err != nil {
		t.Fatal(err)
	}
	duplicate := valid
	duplicate.Packages = append(duplicate.Packages, duplicate.Packages[0])
	duplicateJSON, err := json.Marshal(duplicate)
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name  string
		input io.Reader
	}{
		{"reader failure", equivalentFailedReader{}},
		{"valid bytes followed by reader failure", io.MultiReader(strings.NewReader(reviewedEquivalent), equivalentFailedReader{})},
		{"malformed JSON", strings.NewReader(`{"schema_version":}`)},
		{"wrong schema", strings.NewReader(`{"schema_version":2,"packages":[]}`)},
		{"missing packages", strings.NewReader(`{"schema_version":1}`)},
		{"duplicate package", strings.NewReader(string(duplicateJSON))},
		{"malformed identity", strings.NewReader(strings.Replace(reviewedEquivalent, strings.Repeat("a", 64), "bad", 1))},
	} {
		t.Run(test.name, func(t *testing.T) {
			got, err := mutation.ParseEquivalentInventory(test.input)
			if !errors.Is(err, mutation.ErrInvalid) || got.SchemaVersion != 0 || got.Packages != nil {
				t.Fatalf("ParseEquivalentInventory() = %#v, %v", got, err)
			}
		})
	}
}

func TestEquivalentReviewLookupAndReturnedCoordinatesAreIndependent(t *testing.T) {
	value, err := mutation.ParseEquivalentInventory(strings.NewReader(reviewedEquivalent))
	if err != nil {
		t.Fatal(err)
	}
	source, verifier := strings.Repeat("a", 64), strings.Repeat("b", 64)
	for _, identity := range [][2]string{{"other", "."}, {".", "other"}} {
		review, err := value.Review(identity[0], identity[1], source, "v0.6.0", verifier)
		if review != nil || err != nil {
			t.Fatalf("unmatched Review() = %#v, %v", review, err)
		}
	}
	selected, err := value.Review(".", ".", source, "v0.6.0", verifier)
	if err != nil || selected == nil || len(selected.Mutations) != 1 {
		t.Fatalf("selected Review() = %#v, %v", selected, err)
	}
	selected.Mutations[0].Line = 99
	again, err := value.Review(".", ".", source, "v0.6.0", verifier)
	if err != nil || again == nil || len(again.Mutations) != 1 || again.Mutations[0].Line != 3 || value.Packages[0].Mutations[0].Line != 3 {
		t.Fatalf("caller changed stored review: %#v, %v", again, err)
	}
	value.Packages[0].SourceDigest = "bad"
	if review, err := value.Review(".", ".", source, "v0.6.0", verifier); review != nil || !errors.Is(err, mutation.ErrInvalid) {
		t.Fatalf("malformed caller-owned review = %#v, %v", review, err)
	}
}

func TestEquivalentInventoryIndependentFieldAdmission(t *testing.T) {
	for _, test := range []struct {
		name   string
		change func(*mutation.EquivalentReview)
	}{
		{"module directory", func(review *mutation.EquivalentReview) { review.ModuleDirectory = "../module" }},
		{"package directory", func(review *mutation.EquivalentReview) { review.PackageDirectory = "../package" }},
		{"non-Go filename", func(review *mutation.EquivalentReview) { review.Mutations[0].FileName = "source.txt" }},
		{"escaping filename", func(review *mutation.EquivalentReview) { review.Mutations[0].FileName = "../source.go" }},
		{"empty type", func(review *mutation.EquivalentReview) { review.Mutations[0].Type = "" }},
		{"multiline type", func(review *mutation.EquivalentReview) { review.Mutations[0].Type = "A\nB" }},
		{"zero line", func(review *mutation.EquivalentReview) { review.Mutations[0].Line = 0 }},
		{"zero column", func(review *mutation.EquivalentReview) { review.Mutations[0].Column = 0 }},
		{"short domain", func(review *mutation.EquivalentReview) { review.Mutations[0].ContractDomain = strings.Repeat("d", 19) }},
		{"short reason", func(review *mutation.EquivalentReview) { review.Mutations[0].Reason = strings.Repeat("r", 39) }},
	} {
		t.Run(test.name, func(t *testing.T) {
			value, err := mutation.ParseEquivalentInventory(strings.NewReader(reviewedEquivalent))
			if err != nil {
				t.Fatal(err)
			}
			test.change(&value.Packages[0])
			data, err := json.Marshal(value)
			if err != nil {
				t.Fatal(err)
			}
			got, err := mutation.ParseEquivalentInventory(strings.NewReader(string(data)))
			if !errors.Is(err, mutation.ErrInvalid) || got.SchemaVersion != 0 || got.Packages != nil {
				t.Fatalf("independent field refusal = %#v, %v; want no partial policy and ErrInvalid", got, err)
			}
		})
	}
	value, err := mutation.ParseEquivalentInventory(strings.NewReader(reviewedEquivalent))
	if err != nil {
		t.Fatal(err)
	}
	value.Packages[0].Mutations[0].ContractDomain = strings.Repeat("d", 20)
	value.Packages[0].Mutations[0].Reason = strings.Repeat("r", 40)
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	got, err := mutation.ParseEquivalentInventory(strings.NewReader(string(data)))
	if err != nil || len(got.Packages) != 1 || len(got.Packages[0].Mutations) != 1 {
		t.Fatalf("exact domain/reason minimum = %#v, %v; want complete admitted policy", got, err)
	}
}

func TestEquivalentReviewFindsLaterPackageAndEnforcesItsBinding(t *testing.T) {
	value, err := mutation.ParseEquivalentInventory(strings.NewReader(reviewedEquivalent))
	if err != nil {
		t.Fatal(err)
	}
	matching := value.Packages[0]
	unrelated := matching
	unrelated.PackageDirectory = "unrelated"
	value.Packages = []mutation.EquivalentReview{unrelated, matching}
	source, verifier := strings.Repeat("a", 64), strings.Repeat("b", 64)
	got, err := value.Review(".", ".", source, "v0.6.0", verifier)
	if err != nil || got == nil || got.PackageDirectory != "." || len(got.Mutations) != 1 || got.Mutations[0].Line != 3 {
		t.Fatalf("later matching package = %#v, %v; want its complete review", got, err)
	}
	if got, err := value.Review(".", ".", strings.Repeat("c", 64), "v0.6.0", verifier); got != nil || !errors.Is(err, mutation.ErrInvalid) {
		t.Fatalf("later stale source = %#v, %v; want no review and ErrInvalid", got, err)
	}
	if got, err := value.Review(".", ".", source, "v0.6.0", strings.Repeat("c", 64)); got != nil || !errors.Is(err, mutation.ErrInvalid) {
		t.Fatalf("later stale verifier = %#v, %v; want no review and ErrInvalid", got, err)
	}
}

type equivalentFailedReader struct{}

func (equivalentFailedReader) Read([]byte) (int, error) {
	return 0, errors.New("ordinary read failure")
}
