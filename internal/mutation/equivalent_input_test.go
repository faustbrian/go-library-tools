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

type equivalentFailedReader struct{}

func (equivalentFailedReader) Read([]byte) (int, error) {
	return 0, errors.New("ordinary read failure")
}
