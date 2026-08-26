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
