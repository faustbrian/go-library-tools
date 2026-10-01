package coverage_test

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/faustbrian/go-library-tools/v2/internal/coverage"
)

func TestVerifyRequiresExactCoverageForEveryExpectedPackage(t *testing.T) {
	profile := `mode: atomic
github.com/acme/example/one/file.go:1.1,2.1 2 1
github.com/acme/example/one/file.go:3.1,4.1 1 2
github.com/acme/example/two/file.go:1.1,2.1 3 1
`
	report, err := coverage.Verify(strings.NewReader(profile), []string{
		"github.com/acme/example/two", "github.com/acme/example/one",
	})
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	want := "github.com/acme/example/one 3/3 statements\ngithub.com/acme/example/two 3/3 statements\n"
	if report != want {
		t.Fatalf("Verify() report = %q", report)
	}
}

func TestVerifyMergesDuplicateBlocksFromMultipleTestBinaries(t *testing.T) {
	profile := `mode: atomic
github.com/acme/example/one/file.go:1.1,2.1 2 1
github.com/acme/example/two/file.go:1.1,2.1 3 0
github.com/acme/example/one/file.go:1.1,2.1 2 0
github.com/acme/example/two/file.go:1.1,2.1 3 1
`
	report, err := coverage.Verify(strings.NewReader(profile), []string{
		"github.com/acme/example/one", "github.com/acme/example/two",
	})
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	want := "github.com/acme/example/one 2/2 statements\ngithub.com/acme/example/two 3/3 statements\n"
	if report != want {
		t.Fatalf("Verify() report = %q", report)
	}
}

func TestVerifyRejectsIncompleteMissingAndMalformedProfiles(t *testing.T) {
	tests := []struct {
		name     string
		profile  string
		expected []string
		want     string
	}{
		{"missing mode", "package/file.go:1.1,2.1 1 1\n", []string{"package"}, "mode"},
		{"malformed block", "mode: atomic\nbad\n", []string{"package"}, "line 2"},
		{"missing location", "mode: atomic\nbad 1 1\n", []string{"package"}, "line 2"},
		{"zero location separator", "mode: atomic\n:1.1,2.1 1 1\n", []string{"package"}, "line 2"},
		{"invalid statements", "mode: atomic\npackage/file.go:1.1,2.1 nope 1\n", []string{"package"}, "line 2"},
		{"zero statements", "mode: atomic\npackage/file.go:1.1,2.1 0 1\n", []string{"package"}, "missing executable"},
		{"invalid count", "mode: atomic\npackage/file.go:1.1,2.1 1 nope\n", []string{"package"}, "line 2"},
		{"inconsistent duplicate", "mode: atomic\npackage/file.go:1.1,2.1 1 1\npackage/file.go:1.1,2.1 2 1\n", []string{"package"}, "inconsistent duplicate"},
		{"uncovered", "mode: atomic\npackage/file.go:1.1,2.1 1 0\n", []string{"package"}, "below exact"},
		{"missing package", "mode: atomic\nother/file.go:1.1,2.1 1 1\n", []string{"package"}, "missing executable"},
		{"empty expected", "mode: atomic\npackage/file.go:1.1,2.1 1 1\n", nil, "expected packages"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := coverage.Verify(strings.NewReader(test.profile), test.expected)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Verify() error = %v", err)
			}
		})
	}
}

func TestVerifyReportsReaderFailure(t *testing.T) {
	reader := &failingReader{}
	_, err := coverage.Verify(reader, []string{"package"})
	if err == nil || !strings.Contains(err.Error(), "read coverage profile") {
		t.Fatalf("Verify() error = %v", err)
	}
}

func TestVerifyRejectsStatementTotalOverflow(t *testing.T) {
	maximum := int(^uint(0) >> 1)
	profile := "mode: atomic\npackage/one.go:1.1,2.1 " + strconv.Itoa(maximum) +
		" 1\npackage/two.go:1.1,2.1 1 1\n"
	_, err := coverage.Verify(strings.NewReader(profile), []string{"package"})
	if err == nil || !strings.Contains(err.Error(), "overflows") {
		t.Fatalf("Verify() error = %v", err)
	}
}

func TestVerifyAcceptsMaximumStatementTotal(t *testing.T) {
	maximum := int(^uint(0) >> 1)
	profile := "mode: atomic\npackage/file.go:1.1,2.1 " + strconv.Itoa(maximum) + " 1\n"
	report, err := coverage.Verify(strings.NewReader(profile), []string{"package"})
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	want := "package " + strconv.Itoa(maximum) + "/" + strconv.Itoa(maximum) + " statements\n"
	if report != want {
		t.Fatalf("Verify() report = %q", report)
	}
}

func TestVerifyAcceptsCoverageLinesBeyondScannerDefault(t *testing.T) {
	packagePath := strings.Repeat("segment", 700)
	profile := "mode: atomic\n" + packagePath + "/file.go:1.1,2.1 1 1\n"
	report, err := coverage.Verify(strings.NewReader(profile), []string{packagePath})
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if report != packagePath+" 1/1 statements\n" {
		t.Fatalf("Verify() report length = %d", len(report))
	}
}

type failingReader struct {
	done bool
}

func (reader *failingReader) Read(target []byte) (int, error) {
	if !reader.done {
		reader.done = true
		copy(target, "mode: atomic\n")
		return len("mode: atomic\n"), nil
	}
	return 0, errors.New("injected failure")
}

func TestVerifyReportsUncoveredProductionBlocks(t *testing.T) {
	profile := `mode: atomic
github.com/acme/example/one/z.go:10.1,12.1 2 0
github.com/acme/example/other/unrelated.go:1.1,2.1 1 0
github.com/acme/example/one/covered.go:1.1,2.1 1 1
github.com/acme/example/one/a.go:3.1,4.1 1 0
`
	report, err := coverage.Verify(strings.NewReader(profile), []string{"github.com/acme/example/one"})
	if err == nil || err.Error() != "github.com/acme/example/one is below exact 100% coverage" {
		t.Fatalf("Verify() error = %v", err)
	}
	want := "uncovered production blocks:\n  a.go:3.1,4.1\n  z.go:10.1,12.1\n"
	if report != want {
		t.Fatalf("Verify() report = %q, want %q", report, want)
	}
}

func TestVerifyDiagnosticsRespectMergedCountsAndPackageOrder(t *testing.T) {
	profile := `mode: atomic
example/z/other.go:1.1,2.1 1 0
example/a/covered.go:1.1,2.1 1 0
example/a/missing.go:3.1,4.1 2 0
example/a/missing.go:3.1,4.1 2 0
example/a/covered.go:1.1,2.1 1 1
example/a/empty.go:5.1,6.1 0 0
`
	want := "uncovered production blocks:\n  missing.go:3.1,4.1\n"
	lines := strings.Split(strings.TrimSuffix(profile, "\n"), "\n")
	slices.Reverse(lines[1:])
	reversed := strings.Join(lines, "\n") + "\n"
	for _, input := range []string{profile, reversed} {
		report, err := coverage.Verify(strings.NewReader(input), []string{"example/z", "example/a"})
		if err == nil || err.Error() != "example/a is below exact 100% coverage" || report != want {
			t.Fatalf("Verify() = %q, %v", report, err)
		}
	}
}

func TestVerifyReportsCurrentReleaseCoverageGaps(t *testing.T) {
	var profile strings.Builder
	profile.WriteString("mode: atomic\n")
	for index := 107; index >= 0; index-- {
		fmt.Fprintf(&profile, "example/file-%03d.go:1.1,2.1 1 0\n", index)
	}
	report, err := coverage.Verify(strings.NewReader(profile.String()), []string{"example"})
	if err == nil || err.Error() != "example is below exact 100% coverage" ||
		strings.Count(report, "file-") != 108 || strings.Contains(report, "omitted") ||
		!strings.Contains(report, "file-000.go:1.1,2.1") || !strings.Contains(report, "file-107.go:1.1,2.1") {
		t.Fatalf("complete bounded release diagnostics = %q, %v", report, err)
	}
}

func TestVerifyBoundsUncoveredBlockDiagnostics(t *testing.T) {
	var profile strings.Builder
	profile.WriteString("mode: atomic\n")
	for index := 512; index >= 0; index-- {
		fmt.Fprintf(&profile, "example/file-%03d.go:1.1,2.1 1 0\n", index)
	}
	fmt.Fprintf(&profile, "example/%s.go:1.1,2.1 1 0\n", strings.Repeat("long", 100))
	report, err := coverage.Verify(strings.NewReader(profile.String()), []string{"example"})
	if err == nil || strings.Count(report, "file-") != 512 || len(report) > 85000 ||
		!strings.Contains(report, "file-000.go:1.1,2.1") || !strings.Contains(report, "file-511.go:1.1,2.1") ||
		strings.Contains(report, "file-512.go") || strings.Contains(report, "longlong") ||
		!strings.Contains(report, "(2 additional blocks omitted)") {
		t.Fatalf("bounded diagnostics = %q, %v", report, err)
	}
}

func TestVerifyDoesNotDiscloseUnsafeProfileLocations(t *testing.T) {
	profile := "mode: atomic\n" +
		"PRIVATE_SENTINEL/../example/a/hidden.go:1.1,2.1 1 0\n" +
		"/private/ABSOLUTE_SENTINEL/hidden.go:1.1,2.1 1 0\n" +
		"example/a/SECRET_SENTINEL:1.1,2.1 1 0\n" +
		"example/a/bad.go:SECRET_COORDINATE 1 0\n" +
		"example/a/control\x1b.go:1.1,2.1 1 0\n"
	report, err := coverage.Verify(strings.NewReader(profile), []string{"example/a"})
	if err == nil || err.Error() != "example/a is below exact 100% coverage" ||
		strings.Contains(report, "SENTINEL") || strings.Contains(report, "SECRET_COORDINATE") ||
		strings.Contains(report, "\x1b") || !strings.Contains(report, `control\x1b.go:1.1,2.1`) ||
		!strings.Contains(report, "(3 additional blocks omitted)") {
		t.Fatalf("safe diagnostics = %q, %v", report, err)
	}
	report, err = coverage.Verify(strings.NewReader("mode: atomic\nSECRET_SENTINEL invalid raw profile\n"), []string{"example/a"})
	if err == nil || err.Error() != "invalid coverage profile line 2" || report != "" {
		t.Fatalf("malformed diagnostics = %q, %v", report, err)
	}
}
