package cohesion

import "errors"

// NormalizedV3 is the canonical, schema-validated representation of a
// trust-bearing Cohesion document.
type NormalizedV3 struct {
	JSON   []byte
	SHA256 string
}

// DiagnosticV3 is the stable, secret-safe failure surface consumed by the V2
// command adapter. It intentionally carries no resolution-map paths.
type DiagnosticV3 struct {
	Code    string
	Path    string
	Message string
}

// ValidateAndNormalizeV3 executes the production strict-parser, schema, static
// semantic, and canonicalization stages for one registered schema identity.
func ValidateAndNormalizeV3(schemaIdentity string, input []byte) (NormalizedV3, error) {
	validated, err := validateAndNormalizeSchemaV3(schemaIdentity, input)
	if err != nil {
		return NormalizedV3{}, err
	}
	return NormalizedV3{JSON: validated.Normalized, SHA256: validated.NormalizedSHA256}, nil
}

// ValidateAndNormalizeV3File retains a no-follow file identity while reading
// a bounded trust-bearing artifact and then executes the production pipeline.
func ValidateAndNormalizeV3File(schemaIdentity, path string) (NormalizedV3, error) {
	input, err := readResolutionFile(path, int64(schemaV3ArtifactByteLimit(schemaIdentity)))
	if err != nil {
		return NormalizedV3{}, err
	}
	return ValidateAndNormalizeV3(schemaIdentity, input)
}

// DiagnosticFromV3Error extracts a closed diagnostic without leaking the
// implementation detail attached to an error.
func DiagnosticFromV3Error(err error) (DiagnosticV3, bool) {
	var failure schemaV3Failure
	if !errors.As(err, &failure) {
		return DiagnosticV3{}, false
	}
	return DiagnosticV3{Code: string(failure.Code), Path: failure.Pointer, Message: string(failure.Code)}, true
}
