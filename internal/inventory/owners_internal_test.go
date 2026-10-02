package inventory

import (
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

func TestCompileModuleSchemasFailsClosed(t *testing.T) {
	for _, test := range []struct {
		name, identity, source string
	}{
		{"syntax", modulesV3SchemaIdentity, `{`},
		{"resource", `%`, `{}`},
		{"compile", modulesV3SchemaIdentity, `{"type":1}`},
	} {
		t.Run(test.name, func(t *testing.T) {
			defer func() {
				failure := recover()
				err, ok := failure.(error)
				if !ok {
					t.Fatalf("compilation must panic with its original error, got %T", failure)
				}
				switch test.name {
				case "syntax":
					if !errors.Is(err, io.ErrUnexpectedEOF) {
						t.Fatalf("schema syntax cause = %v", err)
					}
				case "resource":
					var cause *jsonschema.ParseURLError
					if !errors.As(err, &cause) {
						t.Fatalf("schema resource cause = %v", err)
					}
				case "compile":
					var cause *jsonschema.SchemaValidationError
					if !errors.As(err, &cause) {
						t.Fatalf("schema compilation cause = %v", err)
					}
				}
			}()
			compileModuleSchemas(map[string]string{test.identity: test.source}, test.identity)
		})
	}
	compiled := compileModuleSchemas(map[string]string{modulesV3SchemaIdentity: `{"type":"object"}`}, modulesV3SchemaIdentity)
	if err := compiled.Validate(map[string]any{}); err != nil {
		t.Fatal(err)
	}
	if err := compiled.Validate(nil); err == nil {
		t.Fatal("compiled schema accepted null instead of an object")
	}
}

func TestProjectSchemaV3InventoryOwnsDocumentAdmission(t *testing.T) {
	for _, test := range []struct {
		name, source, want string
	}{
		{"syntax", `{`, "unexpected end of JSON input"},
		{"modules", `{"modules":null}`, "schema-v3 inventory modules must be an array"},
		{"module", `{"modules":[null]}`, "schema-v3 inventory module must be an object"},
	} {
		t.Run(test.name, func(t *testing.T) {
			encoded, err := projectSchemaV3Inventory([]byte(test.source))
			if err == nil || err.Error() != test.want || encoded != nil {
				t.Fatalf("projection = %s, %v", encoded, err)
			}
		})
	}
	source := `{"modules":[{"provenance":{"source":"local"},"goal_status":"queued","goal_files":[],"goal_evidence":[],"cohesion":{"delivery":{},"scope":"owned"},"directory":"."}],"schema_version":3}`
	want := `{"modules":[{"cohesion":{"scope":"owned"},"directory":"."}],"schema_version":3}`
	encoded, err := projectSchemaV3Inventory([]byte(source))
	if err != nil || string(encoded) != want {
		t.Fatalf("projection = %s, %v", encoded, err)
	}
}

func TestValidateSchemaV3ManifestOwnsParserAndSchemaErrors(t *testing.T) {
	err := validateSchemaV3Manifest([]byte(`{`))
	if !errors.Is(err, io.ErrUnexpectedEOF) || !strings.HasPrefix(err.Error(), "load module manifest: ") {
		t.Fatalf("incomplete parser input = %v", err)
	}
	err = validateSchemaV3Manifest([]byte(`!`))
	var syntax *json.SyntaxError
	if !errors.As(err, &syntax) || !strings.HasPrefix(err.Error(), "load module manifest: ") {
		t.Fatalf("parser error = %v", err)
	}
	err = validateSchemaV3Manifest([]byte(`{}`))
	var validation *jsonschema.ValidationError
	if !errors.As(err, &validation) || !strings.HasPrefix(err.Error(), "load module manifest: schema v3: ") {
		t.Fatalf("schema error = %v", err)
	}
}
