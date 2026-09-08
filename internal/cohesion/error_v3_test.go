//nolint:copyloopvar,modernize // parallel subtests intentionally capture immutable cases
package cohesion

import (
	"slices"
	"testing"
)

func TestSchemaV3FailureErrorFormatsCodeAndDetail(t *testing.T) {
	if got := (schemaV3Failure{Code: "schema-type"}).Error(); got != "schema-type" {
		t.Fatalf("Error() without detail = %q", got)
	}
	if got := (schemaV3Failure{Code: "schema-type", Detail: "expected object"}).Error(); got != "schema-type: expected object" {
		t.Fatalf("Error() with detail = %q", got)
	}
}

func TestSchemaV3ErrorCodesAreClosedAndOrdered(t *testing.T) {
	t.Parallel()

	want := []schemaV3ErrorCode{
		"canonical-encoding-required",
		"json-bom", "json-duplicate-key", "json-integer-overflow", "json-invalid-utf8", "json-lone-surrogate", "json-negative-zero", "json-noncharacter", "json-noninteger-number", "json-trailing-value",
		"limit-allocation-capacity", "limit-array-elements", "limit-artifact-bytes", "limit-depth", "limit-object-members", "limit-string-bytes",
		"schema-constant", "schema-enum", "schema-pattern", "schema-range", "schema-required-member", "schema-type", "schema-union", "schema-unknown-member",
		"semantic-canonical-order", "semantic-count-mismatch", "semantic-cross-field", "semantic-digest-mismatch", "semantic-duplicate", "semantic-external-identity", "semantic-path", "semantic-preimage", "semantic-release-identity", "semantic-unsupported",
	}
	if !slices.Equal(schemaV3ErrorCodes(), want) {
		t.Fatalf("schema-v3 error codes = %q, want closed set %q", schemaV3ErrorCodes(), want)
	}
	if _, err := newSchemaV3Failure("invented", 0, 0, ""); err == nil {
		t.Fatal("unknown schema-v3 error code accepted")
	}
	if _, err := newSchemaV3Failure("schema-type", -1, 0, ""); err == nil {
		t.Fatal("negative schema-v3 byte offset accepted")
	}
	if failure, err := newSchemaV3Failure("schema-type", 3, 7, "/value"); err != nil || failure.ByteOffset != 3 || failure.Occurrence != 7 || failure.Pointer != "/value" {
		t.Fatalf("valid schema-v3 failure = (%+v, %v)", failure, err)
	}
}

func TestSelectSchemaV3FailureUsesFrozenPrecedence(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		candidates []schemaV3Failure
		want       schemaV3ErrorCode
	}{
		{"stage", []schemaV3Failure{{Code: "schema-type", Pointer: "/a"}, {Code: "json-bom"}, {Code: "limit-artifact-bytes"}}, "limit-artifact-bytes"},
		{"encoding", []schemaV3Failure{{Code: "json-invalid-utf8"}, {Code: "json-bom"}}, "json-bom"},
		{"scanner offset", []schemaV3Failure{{Code: "json-duplicate-key", ByteOffset: 9}, {Code: "json-trailing-value", ByteOffset: 4}}, "json-trailing-value"},
		{"scanner tie", []schemaV3Failure{{Code: "json-duplicate-key", ByteOffset: 4}, {Code: "json-lone-surrogate", ByteOffset: 4}}, "json-lone-surrogate"},
		{"capacity occurrence", []schemaV3Failure{{Code: "limit-string-bytes", Occurrence: 5}, {Code: "limit-depth", Occurrence: 2}}, "limit-depth"},
		{"capacity tie", []schemaV3Failure{{Code: "limit-depth", Occurrence: 2}, {Code: "limit-allocation-capacity", Occurrence: 2}}, "limit-allocation-capacity"},
		{"schema pointer", []schemaV3Failure{{Code: "schema-required-member", Pointer: "/z"}, {Code: "schema-union", Pointer: "/a"}}, "schema-union"},
		{"schema pointer decoded bytes", []schemaV3Failure{{Code: "schema-type", Pointer: "/~0"}, {Code: "schema-enum", Pointer: "/~1"}}, "schema-enum"},
		{"schema tie", []schemaV3Failure{{Code: "schema-union", Pointer: "/a"}, {Code: "schema-required-member", Pointer: "/a"}}, "schema-required-member"},
		{"semantic tie", []schemaV3Failure{{Code: "semantic-unsupported", Pointer: "/a"}, {Code: "semantic-path", Pointer: "/a"}}, "semantic-path"},
		{"canonical last", []schemaV3Failure{{Code: "canonical-encoding-required"}, {Code: "semantic-unsupported", Pointer: "/z"}}, "semantic-unsupported"},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got, ok := selectSchemaV3Failure(test.candidates)
			if !ok || got.Code != test.want {
				t.Fatalf("selectSchemaV3Failure() = (%+v, %v), want code %q", got, ok, test.want)
			}
		})
	}
}

func TestCompareSchemaV3FailuresUsesDecodedPointersAndUnknownFallback(t *testing.T) {
	t.Parallel()

	if got := compareSchemaV3Failures(
		schemaV3Failure{Code: "semantic-path", Pointer: "/~1"},
		schemaV3Failure{Code: "semantic-path", Pointer: "/~0"},
	); got >= 0 {
		t.Fatalf("decoded semantic pointer comparison = %d, want negative", got)
	}
	if stage, rank := schemaV3FailureRank("invented"); stage != 8 || rank != 0 {
		t.Fatalf("unknown schema-v3 rank = (%d, %d), want (8, 0)", stage, rank)
	}
}
