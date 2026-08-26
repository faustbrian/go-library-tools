package config

import "testing"

func TestStepValidationAcceptsDefaultCount(t *testing.T) {
	step := Step{Type: "go-test", Packages: []string{"."}, Timeout: "1m"}
	if err := step.validate(); err != nil {
		t.Fatalf("validate() error = %v", err)
	}
}
