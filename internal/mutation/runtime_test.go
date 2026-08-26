package mutation_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/faustbrian/go-library-tools/internal/mutation"
)

func TestParseRuntimeIdentityAcceptsExactGoEnvironment(t *testing.T) {
	identity, err := mutation.ParseRuntimeIdentity(strings.NewReader(`{"GOVERSION":"go1.27.0","GOOS":"linux","GOARCH":"amd64","CGO_ENABLED":"0"}`))
	if err != nil || identity.GoVersion != "go1.27.0" || identity.GOARCH != "amd64" {
		t.Fatalf("ParseRuntimeIdentity() = %#v, %v", identity, err)
	}
}

func TestParseRuntimeIdentityRejectsUntrustedOutput(t *testing.T) {
	for _, input := range []string{
		`{}`,
		`{"GOVERSION":"go1.27.0","GOOS":"linux","GOARCH":"amd64","CGO_ENABLED":"0","SECRET":"value"}`,
		`{"GOVERSION":"go1.27.0","GOOS":"linux","GOARCH":"amd64","CGO_ENABLED":"0"}{}`,
		`{`,
		strings.Repeat("x", 16<<10+1),
	} {
		if _, err := mutation.ParseRuntimeIdentity(strings.NewReader(input)); !errors.Is(err, mutation.ErrInvalid) {
			t.Fatalf("ParseRuntimeIdentity(%q) error = %v", input[:min(len(input), 32)], err)
		}
	}
}

func TestParseRuntimeIdentityRequiresEveryFieldIndependently(t *testing.T) {
	fields := []string{"GOVERSION", "GOOS", "GOARCH", "CGO_ENABLED"}
	for _, missing := range fields {
		t.Run(missing, func(t *testing.T) {
			values := map[string]string{"GOVERSION": "go1.27.0", "GOOS": "linux", "GOARCH": "amd64", "CGO_ENABLED": "0"}
			values[missing] = ""
			input := `{"GOVERSION":"` + values["GOVERSION"] + `","GOOS":"` + values["GOOS"] + `","GOARCH":"` + values["GOARCH"] + `","CGO_ENABLED":"` + values["CGO_ENABLED"] + `"}`
			if _, err := mutation.ParseRuntimeIdentity(strings.NewReader(input)); !errors.Is(err, mutation.ErrInvalid) {
				t.Fatalf("ParseRuntimeIdentity() error = %v", err)
			}
		})
	}
}

func TestParseRuntimeIdentityAcceptsExactSizeLimit(t *testing.T) {
	input := []byte(`{"GOVERSION":"go1.27.0","GOOS":"linux","GOARCH":"amd64","CGO_ENABLED":"0"}`)
	input = append(input, []byte(strings.Repeat(" ", (16<<10)-len(input)))...)
	if _, err := mutation.ParseRuntimeIdentity(strings.NewReader(string(input))); err != nil {
		t.Fatalf("ParseRuntimeIdentity() error = %v", err)
	}
}

func TestParseRuntimeIdentityRequiresFinalByteAtSizeLimit(t *testing.T) {
	prefix := `{"GOVERSION":"go1.27.0","GOOS":"linux","GOARCH":"amd64","CGO_ENABLED":"0"`
	input := prefix + strings.Repeat(" ", (16<<10)-len(prefix)-1) + `}`
	if _, err := mutation.ParseRuntimeIdentity(strings.NewReader(input)); err != nil {
		t.Fatal(err)
	}
}
