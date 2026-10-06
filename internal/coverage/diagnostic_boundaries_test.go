package coverage_test

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/faustbrian/go-library-tools/v2/internal/coverage"
)

func TestVerifyDiagnosticLocationsRequireRelativeNames(t *testing.T) {
	for _, packagePath := range []string{"/private", "example:a", `example\a`} {
		t.Run(packagePath, func(t *testing.T) {
			profile := "mode: atomic\n" + packagePath + "/file.go:1.1,2.1 1 0\n"
			report, err := coverage.Verify(strings.NewReader(profile), []string{packagePath})
			want := "uncovered production blocks:\n  (1 additional blocks omitted)\n"
			if err == nil || report != want {
				t.Fatalf("diagnostic = %q, %v; want %q", report, err, want)
			}
		})
	}
}

func TestVerifyDiagnosticSelectionAccountsForEveryBlock(t *testing.T) {
	var profile, want strings.Builder
	profile.WriteString("mode: atomic\n")
	want.WriteString("uncovered production blocks:\n")
	for index := 0; index < 1024; index++ {
		fmt.Fprintf(&profile, "example/file-%04d.go:1.1,2.1 1 0\n", index)
		if index < 512 {
			fmt.Fprintf(&want, "  file-%04d.go:1.1,2.1\n", index)
		}
	}
	want.WriteString("  (512 additional blocks omitted)\n")
	report, err := coverage.Verify(strings.NewReader(profile.String()), []string{"example"})
	if err == nil || report != want.String() {
		t.Fatal("capped diagnostic selection lost sorted locations or total omission accounting")
	}
}

func TestVerifyDiagnosticLocationInclusiveBoundary(t *testing.T) {
	for _, size := range []int{159, 160, 161} {
		t.Run(strconv.Itoa(size), func(t *testing.T) {
			location := strings.Repeat("a", size-11) + ".go:1.1,2.1"
			profile := "mode: atomic\nexample/" + location + " 1 0\n"
			report, err := coverage.Verify(strings.NewReader(profile), []string{"example"})
			want := "uncovered production blocks:\n"
			if size <= 160 {
				want += "  " + location + "\n"
			} else {
				want += "  (1 additional blocks omitted)\n"
			}
			if err == nil || report != want {
				t.Fatalf("diagnostic = %q, %v; want %q", report, err, want)
			}
		})
	}
}

func TestVerifyAggregateDiagnosticLabelAdmission(t *testing.T) {
	cases := []struct {
		name, label, display string
	}{
		{"ordinary", "example/a", "example/a"},
		{"absolute", "/private", ""},
		{"unclean", "private/../example", ""},
		{"colon", "example:a", ""},
		{"backslash", `example\a`, ""},
		{"raw inclusive", strings.Repeat("a", 160), strings.Repeat("a", 160)},
		{"raw oversized", strings.Repeat("a", 161), ""},
		{"escaped below", strings.Repeat("a", 155) + "\x01", strings.Repeat("a", 155) + `\x01`},
		{"escaped inclusive", strings.Repeat("a", 156) + "\x01", strings.Repeat("a", 156) + `\x01`},
		{"escaped oversized", strings.Repeat("a", 157) + "\x01", ""},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			report, err := coverage.Verify(strings.NewReader("mode: atomic\n"), []string{test.label, "zzzz"})
			want := "incomplete production coverage:\n"
			if test.display != "" {
				want += "  " + test.display + " missing executable coverage evidence\n"
			}
			want += "  zzzz missing executable coverage evidence\n"
			if test.display == "" {
				want += "  (1 additional packages omitted)\n"
			}
			if err == nil || report != want {
				t.Fatalf("diagnostic = %q, %v; want %q", report, err, want)
			}
		})
	}
}
