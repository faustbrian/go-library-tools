package coverage_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/faustbrian/go-library-tools/internal/coverage"
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
		{"invalid statements", "mode: atomic\npackage/file.go:1.1,2.1 nope 1\n", []string{"package"}, "line 2"},
		{"invalid count", "mode: atomic\npackage/file.go:1.1,2.1 1 nope\n", []string{"package"}, "line 2"},
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
