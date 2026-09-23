package inventory

import "testing"

func TestValidateModuleIdentitiesRejectsMalformedEnvelopeSafely(t *testing.T) {
	t.Parallel()

	err := validateModuleIdentities("", []byte(`{"schema_version":1,"modules":{"hostile-secret":"value"}}`))
	if err == nil || err.Error() != "invalid module manifest" {
		t.Fatalf("validateModuleIdentities() error = %v, want sanitized invalid module manifest", err)
	}
}
