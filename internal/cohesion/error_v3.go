package cohesion

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

type schemaV3ErrorCode string

var schemaV3ClosedErrorCodes = []schemaV3ErrorCode{
	"canonical-encoding-required",
	"json-bom", "json-duplicate-key", "json-integer-overflow", "json-invalid-utf8", "json-lone-surrogate", "json-negative-zero", "json-noncharacter", "json-noninteger-number", "json-trailing-value",
	"limit-allocation-capacity", "limit-array-elements", "limit-artifact-bytes", "limit-depth", "limit-object-members", "limit-string-bytes",
	"schema-constant", "schema-enum", "schema-pattern", "schema-range", "schema-required-member", "schema-type", "schema-union", "schema-unknown-member",
	"semantic-canonical-order", "semantic-count-mismatch", "semantic-cross-field", "semantic-digest-mismatch", "semantic-duplicate", "semantic-external-identity", "semantic-path", "semantic-preimage", "semantic-release-identity", "semantic-unsupported",
}

var schemaV3KnownErrorCodes = func() map[schemaV3ErrorCode]struct{} {
	result := make(map[schemaV3ErrorCode]struct{}, len(schemaV3ClosedErrorCodes))
	for _, code := range schemaV3ClosedErrorCodes {
		result[code] = struct{}{}
	}
	return result
}()

type schemaV3Failure struct {
	Code       schemaV3ErrorCode
	ByteOffset int
	Occurrence uint64
	Pointer    string
	Detail     string
}

func (failure schemaV3Failure) Error() string {
	if failure.Detail == "" {
		return string(failure.Code)
	}
	return fmt.Sprintf("%s: %s", failure.Code, failure.Detail)
}

func schemaV3ErrorCodes() []schemaV3ErrorCode {
	return slices.Clone(schemaV3ClosedErrorCodes)
}

func newSchemaV3Failure(code schemaV3ErrorCode, byteOffset int, occurrence uint64, pointer string) (schemaV3Failure, error) {
	if _, ok := schemaV3KnownErrorCodes[code]; !ok {
		return schemaV3Failure{}, fmt.Errorf("unknown schema-v3 error code %q", code)
	}
	if byteOffset < 0 {
		return schemaV3Failure{}, errors.New("schema-v3 failure byte offset must be nonnegative")
	}
	return schemaV3Failure{Code: code, ByteOffset: byteOffset, Occurrence: occurrence, Pointer: pointer}, nil
}

func selectSchemaV3Failure(candidates []schemaV3Failure) (schemaV3Failure, bool) {
	if len(candidates) == 0 {
		return schemaV3Failure{}, false
	}
	best := candidates[0]
	for _, candidate := range candidates[1:] {
		if compareSchemaV3Failures(candidate, best) < 0 {
			best = candidate
		}
	}
	return best, true
}

func compareSchemaV3Failures(left, right schemaV3Failure) int {
	leftStage, leftRank := schemaV3FailureRank(left.Code)
	rightStage, rightRank := schemaV3FailureRank(right.Code)
	if leftStage != rightStage {
		return leftStage - rightStage
	}
	switch leftStage {
	case 3:
		if left.ByteOffset != right.ByteOffset {
			return left.ByteOffset - right.ByteOffset
		}
	case 4:
		if left.Occurrence < right.Occurrence {
			return -1
		}
		if left.Occurrence > right.Occurrence {
			return 1
		}
	case 5, 6:
		if compared := strings.Compare(decodeRFC6901Pointer(left.Pointer), decodeRFC6901Pointer(right.Pointer)); compared != 0 {
			return compared
		}
	}
	if leftRank != rightRank {
		return leftRank - rightRank
	}
	if compared := strings.Compare(left.Pointer, right.Pointer); compared != 0 {
		return compared
	}
	if left.ByteOffset != right.ByteOffset {
		return left.ByteOffset - right.ByteOffset
	}
	return strings.Compare(left.Detail, right.Detail)
}

func schemaV3FailureRank(code schemaV3ErrorCode) (int, int) {
	orders := [][]schemaV3ErrorCode{
		{"limit-artifact-bytes"},
		{"json-bom", "json-invalid-utf8"},
		{"json-lone-surrogate", "json-noncharacter", "json-negative-zero", "json-noninteger-number", "json-integer-overflow", "json-duplicate-key", "json-trailing-value"},
		{"limit-allocation-capacity", "limit-string-bytes", "limit-object-members", "limit-array-elements", "limit-depth"},
		{"schema-required-member", "schema-unknown-member", "schema-type", "schema-constant", "schema-enum", "schema-pattern", "schema-range", "schema-union"},
		{"semantic-path", "semantic-count-mismatch", "semantic-duplicate", "semantic-canonical-order", "semantic-cross-field", "semantic-preimage", "semantic-digest-mismatch", "semantic-release-identity", "semantic-external-identity", "semantic-unsupported"},
		{"canonical-encoding-required"},
	}
	for stage, order := range orders {
		if rank := slices.Index(order, code); rank >= 0 {
			return stage + 1, rank
		}
	}
	return len(orders) + 1, 0
}

func decodeRFC6901Pointer(pointer string) string {
	return strings.ReplaceAll(strings.ReplaceAll(pointer, "~1", "/"), "~0", "~")
}
