package cohesion

import (
	"errors"
	"testing"
)

func TestValidateAndNormalizeV3ExposesStableProductionResultAndDiagnostic(t *testing.T) {
	valid := []byte(`{"diagnostics":[{"code":"failure","message":"failed","path":""}],"schema_id":"urn:golib:cohesion:diagnostic:v1","schema_version":1,"status":"failed"}`)
	got, err := ValidateAndNormalizeV3(diagnosticV3SchemaIdentity, valid)
	if err != nil {
		t.Fatal(err)
	}
	if string(got.JSON) != string(valid) || got.SHA256 == "" {
		t.Fatalf("ValidateAndNormalizeV3() = %#v", got)
	}

	_, err = ValidateAndNormalizeV3(diagnosticV3SchemaIdentity, []byte(`{}`))
	diagnostic, ok := DiagnosticFromV3Error(err)
	if !ok || diagnostic.Code != "schema-required-member" || diagnostic.Path == "" {
		t.Fatalf("DiagnosticFromV3Error() = %#v, %v", diagnostic, ok)
	}
}

func TestValidateAndNormalizeV3FileReportsReadFailures(t *testing.T) {
	t.Parallel()

	if _, err := ValidateAndNormalizeV3File(diagnosticV3SchemaIdentity, t.TempDir()+"/missing.json"); err == nil {
		t.Fatal("ValidateAndNormalizeV3File() missing path error = nil")
	}
}

func TestDiagnosticFromV3ErrorRejectsUnrelatedErrors(t *testing.T) {
	t.Parallel()

	if diagnostic, ok := DiagnosticFromV3Error(errors.New("unrelated")); ok || diagnostic != (DiagnosticV3{}) {
		t.Fatalf("DiagnosticFromV3Error(unrelated) = %#v, %v", diagnostic, ok)
	}
}
