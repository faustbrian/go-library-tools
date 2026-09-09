//nolint:perfsprint // frozen validation diagnostics are contractual
package cohesion

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/santhosh-tekuri/jsonschema/v6/kind"
)

func reduceSchemaV3ValidationError(validationError *jsonschema.ValidationError) (schemaV3Failure, error) {
	if validationError == nil {
		return schemaV3Failure{}, fmt.Errorf("schema-v3 validation error is required")
	}
	candidates := make([]schemaV3Failure, 0)
	if err := collectSchemaV3ValidationFailures(validationError, &candidates); err != nil {
		return schemaV3Failure{}, err
	}
	if selected, ok := selectSchemaV3Failure(candidates); ok {
		return selected, nil
	}
	return schemaV3Failure{}, fmt.Errorf("schema-v3 validation error contains no reducible failure")
}

func collectSchemaV3ValidationFailures(validationError *jsonschema.ValidationError, candidates *[]schemaV3Failure) error {
	pointer := encodeRFC6901Pointer(validationError.InstanceLocation)
	appendFailure := func(code schemaV3ErrorCode, path string) {
		*candidates = append(*candidates, schemaV3Failure{Code: code, Pointer: path})
	}
	switch typed := validationError.ErrorKind.(type) {
	case *kind.Schema, *kind.Group, *kind.Reference, *kind.AllOf:
	case *kind.Required:
		for _, missing := range typed.Missing {
			appendFailure("schema-required-member", appendRFC6901Pointer(pointer, missing))
		}
	case *kind.Dependency:
		for _, missing := range typed.Missing {
			appendFailure("schema-required-member", appendRFC6901Pointer(pointer, missing))
		}
	case *kind.DependentRequired:
		for _, missing := range typed.Missing {
			appendFailure("schema-required-member", appendRFC6901Pointer(pointer, missing))
		}
	case *kind.AdditionalProperties:
		for _, property := range typed.Properties {
			appendFailure("schema-unknown-member", appendRFC6901Pointer(pointer, property))
		}
	case *kind.Type, *kind.InvalidJsonValue:
		appendFailure("schema-type", pointer)
	case *kind.Const:
		if reflect.TypeOf(typed.Got) != reflect.TypeOf(typed.Want) {
			appendFailure("schema-type", pointer)
		} else {
			appendFailure("schema-constant", pointer)
		}
	case *kind.Enum:
		appendFailure("schema-enum", pointer)
	case *kind.Pattern, *kind.PropertyNames, *kind.Format:
		appendFailure("schema-pattern", pointer)
	case *kind.MinProperties, *kind.MaxProperties, *kind.MinItems, *kind.MaxItems,
		*kind.AdditionalItems, *kind.MinContains, *kind.MaxContains, *kind.MinLength,
		*kind.MaxLength, *kind.Minimum, *kind.Maximum, *kind.ExclusiveMinimum,
		*kind.ExclusiveMaximum, *kind.MultipleOf:
		appendFailure("schema-range", pointer)
	case *kind.UniqueItems:
		appendFailure("semantic-duplicate", appendRFC6901Pointer(pointer, strconv.Itoa(typed.Duplicates[1])))
	case *kind.OneOf, *kind.AnyOf, *kind.Not, *kind.FalseSchema, *kind.Contains,
		*kind.ContentEncoding, *kind.ContentMediaType, *kind.ContentSchema:
		appendFailure("schema-union", pointer)
	default:
		return fmt.Errorf("unmapped schema-v3 validation kind %T", validationError.ErrorKind)
	}
	for _, cause := range validationError.Causes {
		if err := collectSchemaV3ValidationFailures(cause, candidates); err != nil {
			return err
		}
	}
	return nil
}

func encodeRFC6901Pointer(tokens []string) string {
	var pointer strings.Builder
	for _, token := range tokens {
		pointer.WriteByte('/')
		pointer.WriteString(strings.ReplaceAll(strings.ReplaceAll(token, "~", "~0"), "/", "~1"))
	}
	return pointer.String()
}

func appendRFC6901Pointer(pointer, token string) string {
	return pointer + "/" + strings.ReplaceAll(strings.ReplaceAll(token, "~", "~0"), "/", "~1")
}
