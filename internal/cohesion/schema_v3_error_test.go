package cohesion

import (
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/santhosh-tekuri/jsonschema/v6/kind"
)

func TestReduceSchemaV3ValidationErrorUsesTypedKindsAndPointers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  *jsonschema.ValidationError
		code schemaV3ErrorCode
		path string
	}{
		{"escaped required", &jsonschema.ValidationError{InstanceLocation: []string{"root"}, ErrorKind: &kind.Required{Missing: []string{"~x", "a/b"}}}, "schema-required-member", "/root/a~1b"},
		{"unknown member", &jsonschema.ValidationError{InstanceLocation: []string{"root"}, ErrorKind: &kind.AdditionalProperties{Properties: []string{"z", "a"}}}, "schema-unknown-member", "/root/a"},
		{"type", &jsonschema.ValidationError{InstanceLocation: []string{"value"}, ErrorKind: &kind.Type{}}, "schema-type", "/value"},
		{"constant", &jsonschema.ValidationError{ErrorKind: &kind.Const{}}, "schema-constant", ""},
		{"enum", &jsonschema.ValidationError{ErrorKind: &kind.Enum{}}, "schema-enum", ""},
		{"pattern", &jsonschema.ValidationError{ErrorKind: &kind.Pattern{}}, "schema-pattern", ""},
		{"range", &jsonschema.ValidationError{ErrorKind: &kind.MinItems{}}, "schema-range", ""},
		{"union", &jsonschema.ValidationError{ErrorKind: &kind.OneOf{}}, "schema-union", ""},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			failure, err := reduceSchemaV3ValidationError(test.err)
			if err != nil {
				t.Fatal(err)
			}
			if failure.Code != test.code || failure.Pointer != test.path {
				t.Fatalf("failure = %+v, want code %q pointer %q", failure, test.code, test.path)
			}
		})
	}
}

func TestReduceSchemaV3ValidationErrorSelectsPointerBeforeCode(t *testing.T) {
	t.Parallel()

	err := &jsonschema.ValidationError{ErrorKind: &kind.Schema{}, Causes: []*jsonschema.ValidationError{
		{InstanceLocation: []string{"z"}, ErrorKind: &kind.Type{}},
		{InstanceLocation: []string{"a"}, ErrorKind: &kind.OneOf{}},
	}}
	failure, reduceErr := reduceSchemaV3ValidationError(err)
	if reduceErr != nil {
		t.Fatal(reduceErr)
	}
	if failure.Code != "schema-union" || failure.Pointer != "/a" {
		t.Fatalf("failure = %+v, want schema-union at /a", failure)
	}
}
