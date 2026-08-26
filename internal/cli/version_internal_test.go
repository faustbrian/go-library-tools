package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/faustbrian/go-library-tools/internal/buildinfo"
)

func TestExecuteRejectsWrongReleasedVersion(t *testing.T) {
	original := buildinfo.Version
	buildinfo.Version = "v1.0.1"
	t.Cleanup(func() { buildinfo.Version = original })

	var stdout, stderr bytes.Buffer
	code := Execute([]string{"inventory"}, internalFixture(t), &stdout, &stderr)
	if code != 1 || !strings.Contains(stderr.String(), "does not match required v1.0.0") {
		t.Fatalf("Execute() code = %d, stderr = %q", code, stderr.String())
	}
}
