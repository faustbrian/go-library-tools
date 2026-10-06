package coverage_test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/faustbrian/go-library-tools/v2/internal/coverage"
)

func TestEvidenceUsesProductionCountsWithoutExactClaim(t *testing.T) {
	profile := "mode: atomic\nexample/a/one.go:1.1,2.1 2 0\nexample/a/one.go:1.1,2.1 2 1\nexample/a/two.go:3.1,4.1 3 0\nexample/b/file.go:1.1,2.1 1 1\nother/file.go:1.1,2.1 9 0\n"
	report, err := coverage.VerifyWithMode(strings.NewReader(profile), []string{"example/b", "example/a", "example/a"}, "evidence")
	if err != nil || report != "example/a 2/5 statements\nexample/b 1/1 statements\n" {
		t.Fatalf("evidence counts = %q, %v", report, err)
	}
	exact, err := coverage.VerifyWithMode(strings.NewReader(profile), []string{"example/b", "example/a"}, "exact")
	legacy, legacyErr := coverage.Verify(strings.NewReader(profile), []string{"example/b", "example/a"})
	if err == nil || legacyErr == nil || err.Error() != legacyErr.Error() || exact != legacy {
		t.Fatalf("explicit exact changed legacy output/errors: %q, %v vs %q, %v", exact, err, legacy, legacyErr)
	}
	full := "mode: atomic\nexample/file.go:1.1,2.1 2 1\n"
	for _, mode := range []string{"exact", "evidence"} {
		report, err := coverage.VerifyWithMode(strings.NewReader(full), []string{"example"}, mode)
		if err != nil || report != "example 2/2 statements\n" {
			t.Fatalf("full %s counts = %q, %v", mode, report, err)
		}
	}
}

func TestEvidenceRejectsMissingUnexecutedAndInvalidProfiles(t *testing.T) {
	maximum := strconv.Itoa(int(^uint(0) >> 1))
	for _, test := range []struct{ name, profile, want string }{
		{"missing header", "example/file.go:1.1,2.1 1 1\n", "mode header"},
		{"unsupported header", "mode: invented\nexample/file.go:1.1,2.1 1 1\n", "mode header"},
		{"malformed coordinates", "mode: atomic\nexample/file.go:not-coordinates 1 1\nexample/file.go:3.1,4.1 1 0\n", "line 2"},
		{"missing coordinates", "mode: atomic\nexample/file.go: 1 1\n", "line 2"},
		{"incomplete coordinates", "mode: atomic\nexample/file.go:1.1,2 1 1\n", "line 2"},
		{"missing package", "mode: atomic\nother/file.go:1.1,2.1 1 1\n", "missing executable"},
		{"empty profile", "mode: atomic\n", "missing executable"},
		{"zero denominator", "mode: atomic\nexample/file.go:1.1,2.1 0 1\n", "missing executable"},
		{"unexecuted", "mode: atomic\nexample/file.go:1.1,2.1 2 0\n", "no executed"},
		{"zero statement execution", "mode: atomic\nexample/file.go:1.1,2.1 2 0\nexample/file.go:3.1,4.1 0 1\n", "no executed"},
		{"malformed", "mode: atomic\nbad\n", "line 2"},
		{"negative statements", "mode: atomic\nexample/file.go:1.1,2.1 -1 1\n", "statement count"},
		{"invalid execution", "mode: atomic\nexample/file.go:1.1,2.1 1 -1\n", "execution count"},
		{"inconsistent duplicate", "mode: atomic\nexample/file.go:1.1,2.1 1 1\nexample/file.go:1.1,2.1 2 0\n", "inconsistent duplicate"},
		{"overflow", "mode: atomic\nexample/file.go:1.1,2.1 " + maximum + " 1\nexample/file.go:3.1,4.1 1 0\n", "overflows"},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := coverage.VerifyWithMode(strings.NewReader(test.profile), []string{"example"}, "evidence")
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("evidence error = %v", err)
			}
		})
	}
	if _, err := coverage.VerifyWithMode(&failingReader{}, []string{"example"}, "evidence"); err == nil || !strings.Contains(err.Error(), "read coverage profile") {
		t.Fatalf("reader failure = %v", err)
	}
	if _, err := coverage.VerifyWithMode(strings.NewReader("mode: atomic\n"), nil, "evidence"); err == nil {
		t.Fatal("empty expected inventory accepted")
	}
	for _, mode := range []string{"", "skip", "Evidence"} {
		if _, err := coverage.VerifyWithMode(strings.NewReader("mode: atomic\nexample/file.go:1.1,2.1 1 1\n"), []string{"example"}, mode); err == nil {
			t.Fatalf("invalid acceptance %q admitted", mode)
		}
	}
}
