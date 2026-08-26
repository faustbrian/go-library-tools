package mutation

import (
	"errors"
	"testing"
)

func TestParseRuntimeIdentityReportsReadFailure(t *testing.T) {
	if _, err := ParseRuntimeIdentity(failingReader{}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("ParseRuntimeIdentity() error = %v", err)
	}
}
