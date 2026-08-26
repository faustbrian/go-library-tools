package buildinfo

import (
	"strings"
	"testing"
)

func TestValidateRequired(t *testing.T) {
	original := Version
	t.Cleanup(func() { Version = original })

	for _, test := range []struct {
		version  string
		required string
		want     string
	}{
		{"dev", "v1.0.0", ""},
		{"v1.0.0", "v1.0.0", ""},
		{"invalid", "v1.0.0", "malformed"},
		{"v1.0.1", "v1.0.0", "does not match"},
	} {
		Version = test.version
		err := ValidateRequired(test.required)
		if test.want == "" && err != nil {
			t.Fatalf("ValidateRequired(%q) error = %v", test.required, err)
		}
		if test.want != "" && (err == nil || !strings.Contains(err.Error(), test.want)) {
			t.Fatalf("ValidateRequired(%q) error = %v, want %q", test.required, err, test.want)
		}
	}
}
