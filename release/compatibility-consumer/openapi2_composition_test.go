package compatibilityconsumer_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	canonical "github.com/faustbrian/go-json-schema/v2"
	openapi "github.com/faustbrian/go-openapi/v2"
	schema "github.com/faustbrian/go-openapi/v2/jsonschema"
	"github.com/faustbrian/go-openapi/v2/parse"
	"github.com/faustbrian/go-openapi/v2/security"
)

func TestPublicOpenAPI2Composition(t *testing.T) {
	const source = `{"openapi":"3.1.0","info":{"title":"public","version":"1"},"paths":{},"x-value":-0.0e+00}`
	value, err := parse.JSON(t.Context(), strings.NewReader(source), parse.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	document, err := openapi.Decode(value)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(document.Raw())
	if err != nil || string(encoded) != source {
		t.Fatalf("semantic roundtrip: %s, %v", encoded, err)
	}
	limits := parse.DefaultLimits()
	limits.MaxBytes = 2
	if _, err := parse.JSON(t.Context(), strings.NewReader("[0]"), limits); !errors.Is(err, parse.ErrLimitExceeded) {
		t.Fatalf("bytes: %v", err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, err := parse.JSON(ctx, strings.NewReader("null"), parse.DefaultLimits()); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancel: %v", err)
	}
	compiler, err := schema.NewCompiler(schema.DialectOAS31)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := parse.JSON(t.Context(), strings.NewReader(`{"type":"string","pattern":"^\\p{Script=Greek}+$"}`), parse.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	var compiled *canonical.Schema
	compiled, err = compiler.Compile(t.Context(), raw)
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		value string
		valid bool
	}{{`"αβ"`, true}, {`"AB"`, false}} {
		result, err := compiled.Validate(t.Context(), []byte(test.value))
		if err != nil || result.Valid != test.valid {
			t.Fatalf("Unicode validation: %v, %v", result, err)
		}
	}
}

func TestPublicOpenAPI2AuthorizationAdmission(t *testing.T) {
	requirements, err := parse.JSON(t.Context(), strings.NewReader(`[{"OAuth":["read"]}]`), parse.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		credentials security.Credentials
		valid       bool
	}{
		{security.Credentials{"OAuth": {"read"}}, true}, {security.Credentials{"OAuth": {"write"}}, false},
	} {
		ok, err := security.Satisfied(requirements, test.credentials, security.DefaultLimits())
		if err != nil || ok != test.valid {
			t.Fatalf("authorization: %t, %v", ok, err)
		}
	}
	label := strings.Repeat("private", (1<<20)/7+1)
	ok, err := security.Satisfied(requirements, security.Credentials{"OAuth": {"read", label}}, security.DefaultLimits())
	if ok || !errors.Is(err, security.ErrLimitExceeded) || strings.Contains(err.Error(), "private") {
		t.Fatalf("credential admission: %t, %v", ok, err)
	}
	malformed, err := parse.JSON(t.Context(), strings.NewReader(`[{},42]`), parse.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	if ok, err := security.Satisfied(malformed, nil, security.DefaultLimits()); ok || !errors.Is(err, security.ErrInvalidRequirements) {
		t.Fatalf("malformed alternative: %t, %v", ok, err)
	}
}
