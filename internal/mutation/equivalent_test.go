package mutation_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/faustbrian/go-library-tools/v2/internal/mutation"
)

const reviewedEquivalent = `{"schema_version":1,"packages":[{"module_directory":".","package_directory":".","source_digest":"` +
	"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" + `","gremlins_version":"v0.6.0","gremlins_verifier_sha256":"` +
	"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb" + `","mutations":[{"file_name":"source.go","type":"A","line":3,"column":1,"contract_domain":"Observable integer result for admitted nonnegative values","reason":"Both boundary forms return the same maximum integer at equality."}]}]}`

const livedEquivalentReport = `{"files":[{"file_name":"source.go","mutations":[{"type":"A","status":"LIVED","line":3,"column":1}]}],"mutants_killed":0,"mutants_lived":1,"mutants_not_covered":0,"mutants_not_viable":0,"mutants_total":1,"mutations_coverage":100,"test_efficacy":0}`

func TestReviewedEquivalentRetainsRawStatusAndFailsClosed(t *testing.T) {
	inventory, err := mutation.ParseEquivalentInventory(strings.NewReader(reviewedEquivalent))
	if err != nil {
		t.Fatal(err)
	}
	review, err := inventory.Review(".", ".", strings.Repeat("a", 64), "v0.6.0", strings.Repeat("b", 64))
	if err != nil || review == nil {
		t.Fatalf("Review() = %#v, %v", review, err)
	}
	if _, err := mutation.ValidateReport(strings.NewReader(livedEquivalentReport)); !errors.Is(err, mutation.ErrInvalid) {
		t.Fatalf("default report error = %v", err)
	}
	result, err := mutation.ValidateReportWithReview(strings.NewReader(livedEquivalentReport), review)
	if err != nil || result.Mutants != 1 || result.Killed != 0 || result.Equivalent != 1 {
		t.Fatalf("reviewed result = %#v, %v", result, err)
	}
	mixed := `{"files":[{"file_name":"source.go","mutations":[{"type":"A","status":"LIVED","line":3,"column":1},{"type":"B","status":"KILLED","line":4,"column":1}]}],"mutants_killed":1,"mutants_lived":1,"mutants_not_covered":0,"mutants_not_viable":0,"mutants_total":2,"mutations_coverage":100,"test_efficacy":50}`
	result, err = mutation.ValidateReportWithReview(strings.NewReader(mixed), review)
	if err != nil || result.Mutants != 2 || result.Killed != 1 || result.Equivalent != 1 {
		t.Fatalf("mixed native result = %#v, %v", result, err)
	}
	for name, data := range map[string]string{
		"unlisted":         strings.Replace(livedEquivalentReport, `"line":3`, `"line":4`, 1),
		"timeout":          strings.Replace(livedEquivalentReport, `"LIVED"`, `"TIMED_OUT"`, 1),
		"uncovered":        strings.Replace(livedEquivalentReport, `"LIVED"`, `"NOT_COVERED"`, 1),
		"not viable":       strings.Replace(livedEquivalentReport, `"LIVED"`, `"NOT_VIABLE"`, 1),
		"wrong counters":   strings.Replace(livedEquivalentReport, `"mutants_lived":1`, `"mutants_lived":0`, 1),
		"missing counters": `{"files":[{"file_name":"source.go","mutations":[{"type":"A","status":"LIVED","line":3,"column":1}]}]}`,
		"obsolete review":  strings.Replace(livedEquivalentReport, `"LIVED"`, `"KILLED"`, 1),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := mutation.ValidateReportWithReview(strings.NewReader(data), review); !errors.Is(err, mutation.ErrInvalid) {
				t.Fatalf("ValidateReportWithReview() error = %v", err)
			}
		})
	}
}

func TestEquivalentInventoryRejectsStaleAndAmbiguousReview(t *testing.T) {
	inventory, err := mutation.ParseEquivalentInventory(strings.NewReader(reviewedEquivalent))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := inventory.Review(".", ".", strings.Repeat("c", 64), "v0.6.0", strings.Repeat("b", 64)); !errors.Is(err, mutation.ErrInvalid) {
		t.Fatalf("stale source error = %v", err)
	}
	if _, err := inventory.Review(".", ".", strings.Repeat("a", 64), "v0.6.0", strings.Repeat("c", 64)); !errors.Is(err, mutation.ErrInvalid) {
		t.Fatalf("stale verifier error = %v", err)
	}
	if _, err := mutation.ParseEquivalentInventory(strings.NewReader(strings.Replace(reviewedEquivalent, `"mutations":[`, `"mutations":[{"file_name":"source.go","type":"A","line":3,"column":1,"contract_domain":"Observable integer result for admitted nonnegative values","reason":"Both boundary forms return the same maximum integer at equality."},`, 1))); !errors.Is(err, mutation.ErrInvalid) {
		t.Fatalf("duplicate review error = %v", err)
	}
	if _, err := mutation.ParseEquivalentInventory(strings.NewReader(strings.Replace(reviewedEquivalent, `"contract_domain":"Observable integer result for admitted nonnegative values"`, `"contract_domain":"all"`, 1))); !errors.Is(err, mutation.ErrInvalid) {
		t.Fatalf("unbounded contract domain error = %v", err)
	}
}
