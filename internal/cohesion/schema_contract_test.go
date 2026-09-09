package cohesion

import (
	"bytes"
	"encoding/json"
	"errors"
	"maps"
	"os"
	"strings"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

func TestValidateModulesSchemaAcceptsNormativeDocumentAndRejectsInvalidInput(t *testing.T) {
	valid := []byte(`{
  "schema_version": 2,
  "repository": "github.com/faustbrian/example",
  "go_version": "1.27.0",
  "modules": [{
    "directory": ".", "module_path": "github.com/faustbrian/example", "go_version": "1.27.0",
    "kind": "fixture", "purpose": "", "lifecycle": "", "releasable": false, "version": "", "tag_prefix": "",
    "gates": {}, "test_tags": [], "build_tags": [], "required_services": [], "external_runtime_dependencies": [],
    "interoperability_tools": [], "conformance_corpora": [], "specifications": [], "owned_dependencies": [],
    "reverse_owned_dependencies": [], "packages": [], "family": "", "family_label": "", "family_description": "",
    "family_order": 0, "goal_status": "", "goal_files": [], "goal_evidence": [], "provenance": []
  }]
}`)
	if err := validateModulesSchema(valid); err != nil {
		t.Fatalf("validateModulesSchema(valid) error = %v", err)
	}
	for _, invalid := range [][]byte{[]byte(`{`), []byte(`{"schema_version":2}`)} {
		if err := validateModulesSchema(invalid); err == nil {
			t.Fatalf("validateModulesSchema(%s) error = nil", invalid)
		}
	}
}

func TestEmbeddedCohesionSchemasAreCurrent(t *testing.T) {
	for _, test := range []struct {
		path     string
		embedded string
	}{
		{path: "../../schema/modules.schema.json", embedded: modulesSchemaJSON},
		{path: "../../schema/cohesion-catalog.schema.json", embedded: catalogSchemaJSON},
		{path: "../../schema/cohesion-inputs.schema.json", embedded: inputsSchemaJSON},
		{path: "../../schema/cohesion-sources.schema.json", embedded: sourcesSchemaJSON},
	} {
		current, err := os.ReadFile(test.path)
		if err != nil {
			t.Fatal(err)
		}
		if string(current) != test.embedded {
			t.Fatalf("embedded schema for %s is stale; run go generate ./internal/cohesion", test.path)
		}
	}
}

func TestSourcesSchemaAcceptsReviewedLockAndRejectsMalformedEntries(t *testing.T) {
	if err := validateSourcesSchema([]byte("{")); err == nil {
		t.Fatal("validateSourcesSchema(malformed JSON) error = nil")
	}
	data, err := os.ReadFile("../../release/cohesion-sources.json")
	if err != nil {
		t.Fatal(err)
	}
	var valid map[string]any
	if err := json.Unmarshal(data, &valid); err != nil {
		t.Fatal(err)
	}
	if err := validateSourcesSchema(data); err != nil {
		t.Fatalf("reviewed source lock: %v", err)
	}

	for name, mutate := range map[string]func(map[string]any){
		"unknown field":         func(value map[string]any) { value["unexpected"] = true },
		"zero repository count": func(value map[string]any) { value["repository_count"] = 0 },
		"empty repositories":    func(value map[string]any) { value["repositories"] = []any{} },
		"leading-zero version": func(value map[string]any) {
			first := jsonObjectValue(jsonArrayValue(value["repositories"])[0])
			jsonObjectValue(first["tooling"])["version"] = "v01.4.0"
		},
		"malformed commit": func(value map[string]any) {
			first := jsonObjectValue(jsonArrayValue(value["repositories"])[0])
			jsonObjectValue(first["source"])["commit"] = "main"
		},
		"uppercase commit": func(value map[string]any) {
			first := jsonObjectValue(jsonArrayValue(value["repositories"])[0])
			jsonObjectValue(first["source"])["commit"] = strings.Repeat("A", 40)
		},
		"uppercase checksum": func(value map[string]any) {
			first := jsonObjectValue(jsonArrayValue(value["repositories"])[0])
			jsonObjectValue(first["tooling"])["checksums_sha256"] = strings.Repeat("B", 64)
		},
		"missing consumer tooling": func(value map[string]any) {
			first := jsonObjectValue(jsonArrayValue(value["repositories"])[0])
			delete(first, "tooling")
		},
		"tooling on release source": func(value map[string]any) {
			for _, raw := range jsonArrayValue(value["repositories"]) {
				repository := jsonObjectValue(raw)
				if jsonObjectValue(repository["source"])["kind"] == "release-source" {
					repository["tooling"] = map[string]any{"version": "v1.5.0", "checksums_sha256": strings.Repeat("b", 64)}
					return
				}
			}
		},
		"wrong release-source repository": func(value map[string]any) {
			for _, raw := range jsonArrayValue(value["repositories"]) {
				repository := jsonObjectValue(raw)
				if jsonObjectValue(repository["source"])["kind"] == "release-source" {
					repository["repository"] = "github.com/faustbrian/go-other"
					return
				}
			}
		},
	} {
		t.Run(name, func(t *testing.T) {
			candidate := cloneJSONMap(t, valid)
			mutate(candidate)
			candidateData, err := json.Marshal(candidate)
			if err != nil {
				t.Fatal(err)
			}
			if err := validateSourcesSchema(candidateData); err == nil {
				t.Fatal("validateSourcesSchema() error = nil")
			}
		})
	}
}

func jsonArrayValue(value any) []any {
	result, ok := value.([]any)
	if !ok {
		return nil
	}
	return result
}

func jsonObjectValue(value any) map[string]any {
	result, ok := value.(map[string]any)
	if !ok {
		return map[string]any{}
	}
	return result
}

func TestMustPanicsOnSchemaInitializationFailure(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("must() did not panic")
		}
	}()
	must(0, errors.New("failure"))
}

func TestModulesSchemaRejectsRuntimeForbiddenCohesionValues(t *testing.T) {
	compiled := compileSchemaFile(t, "../../schema/modules.schema.json", "modules-contract.json")
	data, err := os.ReadFile("../../modules.json")
	if err != nil {
		t.Fatal(err)
	}
	var valid map[string]any
	if err := json.Unmarshal(data, &valid); err != nil {
		t.Fatal(err)
	}
	if err := compiled.Validate(valid); err != nil {
		t.Fatalf("current modules manifest: %v", err)
	}
	for _, mutate := range []func(map[string]any){
		func(value map[string]any) {
			metadata := firstModuleCohesion(t, value)
			metadata["lifecycle_status"] = "planned"
			metadata["maturity"] = "stable"
		},
		func(value map[string]any) {
			objectField(t, firstModuleCohesion(t, value), "ownership")["mutable_inputs"] = []any{"copy", "none"}
		},
		func(value map[string]any) { firstModuleCohesion(t, value)["supported_backends"] = []any{"Postgres"} },
		func(value map[string]any) { firstModuleCohesion(t, value)["supported_protocols"] = []any{"HTTP/1"} },
		func(value map[string]any) { firstModuleCohesion(t, value)["responsibility"] = "   " },
		func(value map[string]any) { firstModuleCohesion(t, value)["non_goals"] = []any{"\t"} },
		func(value map[string]any) {
			firstModuleCohesion(t, value)["package_selection"] = map[string]any{"github.com/faustbrian/go-library-tools/cmd/golib": " "}
		},
		func(value map[string]any) { firstModuleCohesion(t, value)["optional_owned_dependencies"] = []any{"\n"} },
	} {
		candidate := cloneJSONMap(t, valid)
		mutate(candidate)
		if err := compiled.Validate(candidate); err == nil {
			t.Fatalf("modules schema accepted runtime-forbidden metadata %#v", firstModuleCohesion(t, candidate))
		}
	}
}

func TestCatalogSchemaRejectsForbiddenEnvelopeCombinations(t *testing.T) {
	compiled := compileCatalogSchema(t)
	valid := map[string]any{
		"schema_version": 1, "view": "consumer", "scope": "repository", "repository": "github.com/faustbrian/example",
		"design_language": map[string]any{"version": "1.0", "sha256": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "source_identity": "unpublished"},
		"tooling":         map[string]any{"version": "dev", "publication_status": "unpublished"}, "modules": []any{},
	}
	if err := compiled.Validate(valid); err != nil {
		t.Fatalf("valid repository envelope: %v", err)
	}
	published := cloneCatalogEnvelope(t, valid)
	objectField(t, published, "design_language")["source_identity"] = "v1.2.3"
	published["tooling"] = map[string]any{"version": "v1.2.3", "publication_status": "published"}
	if err := compiled.Validate(published); err != nil {
		t.Fatalf("valid published envelope: %v", err)
	}
	for _, version := range []string{"v01.2.3", "v1.02.3", "v1.2.03"} {
		for _, field := range []string{"source", "tooling"} {
			candidate := cloneCatalogEnvelope(t, published)
			if field == "source" {
				objectField(t, candidate, "design_language")["source_identity"] = version
			} else {
				objectField(t, candidate, "tooling")["version"] = version
			}
			if err := compiled.Validate(candidate); err == nil {
				t.Fatalf("catalog schema accepted leading-zero %s version %q", field, version)
			}
		}
	}
	for _, mutate := range []func(map[string]any){
		func(value map[string]any) { value["repository"] = nil },
		func(value map[string]any) { value["scope"] = "ecosystem" },
		func(value map[string]any) { objectField(t, value, "tooling")["version"] = "v1.2.3" },
		func(value map[string]any) {
			objectField(t, value, "design_language")["source_identity"] = "v1.2.3"
			value["tooling"] = map[string]any{"version": "dev", "publication_status": "published"}
		},
	} {
		candidate := cloneCatalogEnvelope(t, valid)
		mutate(candidate)
		if err := compiled.Validate(candidate); err == nil {
			t.Fatalf("catalog schema accepted forbidden envelope %#v", candidate)
		}
	}
	engineering := cloneCatalogEnvelope(t, valid)
	engineering["view"] = "engineering"
	engineering["modules"] = []any{engineeringModuleWithoutCohesion()}
	if err := compiled.Validate(engineering); err == nil {
		t.Fatal("catalog schema accepted a releasable engineering module without cohesion metadata")
	}
	planned := engineeringModuleWithoutCohesion()
	metadataJSON, err := json.Marshal(validModule().Cohesion)
	if err != nil {
		t.Fatal(err)
	}
	var metadata map[string]any
	if err := json.Unmarshal(metadataJSON, &metadata); err != nil {
		t.Fatal(err)
	}
	metadata["lifecycle_status"] = "planned"
	metadata["maturity"] = "preview"
	objectField(t, metadata, "documentation")["readme"] = nil
	planned["cohesion"] = metadata
	engineering["modules"] = []any{planned}
	if err := compiled.Validate(engineering); err == nil {
		t.Fatal("catalog schema accepted a planned module without its README entry point")
	}
}

func TestAggregationInputsSchemaAcceptsOnlySafeManifestRelativeProjectionPaths(t *testing.T) {
	compiled := compileSchemaFile(t, "../../schema/cohesion-inputs.schema.json", "cohesion-inputs.json")
	manifest := map[string]any{
		"schema_version": 1,
		"design_language": map[string]any{
			"version": "1.0",
			"sha256":  "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		},
		"repositories": []any{map[string]any{
			"repository": "github.com/faustbrian/go-example",
			"projection": ".ai/cohesion/repository.json",
			"sha256":     "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		}},
	}
	if err := compiled.Validate(manifest); err != nil {
		t.Fatalf("valid aggregation manifest: %v", err)
	}
	for _, path := range []string{"/outside.json", "../outside.json", "projections/../outside.json", `.\\outside.json`, ".", "projections//repository.json"} {
		candidate := cloneJSONMap(t, manifest)
		repositories, ok := candidate["repositories"].([]any)
		if !ok || len(repositories) == 0 {
			t.Fatal("cloned aggregation manifest has no repositories")
		}
		candidateRepository, ok := repositories[0].(map[string]any)
		if !ok {
			t.Fatal("cloned aggregation manifest repository is not an object")
		}
		candidateRepository["projection"] = path
		if err := compiled.Validate(candidate); err == nil {
			t.Fatalf("aggregation inputs schema accepted unsafe projection %q", path)
		}
	}
}

func TestCheckSchemaRejectsInternallyContradictoryReports(t *testing.T) {
	compiled := compileSchemaFile(t, "../../schema/cohesion-check.schema.json", "check.json")
	valid := map[string]any{
		"schema_version": 1, "repository": "github.com/faustbrian/example", "manifest_schema_version": 2, "valid": true,
		"summary": map[string]any{"total_modules": 1, "releasable_modules": 1, "classified_modules": 1, "error_count": 0}, "diagnostics": []any{},
	}
	if err := compiled.Validate(valid); err != nil {
		t.Fatalf("valid check report: %v", err)
	}
	for _, candidate := range []map[string]any{
		{"schema_version": 1, "repository": nil, "manifest_schema_version": nil, "valid": false, "summary": map[string]any{"total_modules": nil, "releasable_modules": 1, "classified_modules": nil, "error_count": 1}, "diagnostics": []any{map[string]any{"code": "invalid-manifest", "path": "", "message": "invalid"}}},
		{"schema_version": 1, "repository": "github.com/faustbrian/example", "manifest_schema_version": 2, "valid": true, "summary": map[string]any{"total_modules": 1, "releasable_modules": 1, "classified_modules": 1, "error_count": 1}, "diagnostics": []any{map[string]any{"code": "invalid-value", "path": "/modules/0", "message": "invalid"}}},
		{"schema_version": 1, "repository": "github.com/faustbrian/example", "manifest_schema_version": 2, "valid": false, "summary": map[string]any{"total_modules": 1, "releasable_modules": 1, "classified_modules": 1, "error_count": 0}, "diagnostics": []any{}},
	} {
		if err := compiled.Validate(candidate); err == nil {
			t.Fatalf("check schema accepted contradictory report %#v", candidate)
		}
	}
}

func compileCatalogSchema(t *testing.T) *jsonschema.Schema {
	t.Helper()
	compiler := jsonschema.NewCompiler()
	modules, err := jsonschema.UnmarshalJSON(bytes.NewReader([]byte(modulesSchemaJSON)))
	if err != nil {
		t.Fatal(err)
	}
	if err := compiler.AddResource(modulesSchemaIdentity, modules); err != nil {
		t.Fatal(err)
	}
	catalogData, err := os.ReadFile("../../schema/cohesion-catalog.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := jsonschema.UnmarshalJSON(bytes.NewReader(catalogData))
	if err != nil {
		t.Fatal(err)
	}
	identity := "https://github.com/faustbrian/go-library-tools/schema/cohesion-catalog.schema.json"
	if err := compiler.AddResource(identity, catalog); err != nil {
		t.Fatal(err)
	}
	compiled, err := compiler.Compile(identity)
	if err != nil {
		t.Fatal(err)
	}
	return compiled
}

func compileSchemaFile(t *testing.T, path, identity string) *jsonschema.Schema {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	document, err := jsonschema.UnmarshalJSON(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource(identity, document); err != nil {
		t.Fatal(err)
	}
	compiled, err := compiler.Compile(identity)
	if err != nil {
		t.Fatal(err)
	}
	return compiled
}

func cloneCatalogEnvelope(t *testing.T, source map[string]any) map[string]any {
	t.Helper()
	clone := maps.Clone(source)
	for _, key := range []string{"design_language", "tooling"} {
		clone[key] = maps.Clone(objectField(t, source, key))
	}
	return clone
}

func objectField(t *testing.T, source map[string]any, key string) map[string]any {
	t.Helper()
	object, ok := source[key].(map[string]any)
	if !ok {
		t.Fatalf("%s is not an object", key)
	}
	return object
}

func firstModuleCohesion(t *testing.T, document map[string]any) map[string]any {
	t.Helper()
	modules, ok := document["modules"].([]any)
	if !ok || len(modules) == 0 {
		t.Fatal("modules is not a non-empty array")
	}
	module, ok := modules[0].(map[string]any)
	if !ok {
		t.Fatal("first module is not an object")
	}
	return objectField(t, module, "cohesion")
}

func cloneJSONMap(t *testing.T, source map[string]any) map[string]any {
	t.Helper()
	data, err := json.Marshal(source)
	if err != nil {
		t.Fatal(err)
	}
	var clone map[string]any
	if err := json.Unmarshal(data, &clone); err != nil {
		t.Fatal(err)
	}
	return clone
}

func engineeringModuleWithoutCohesion() map[string]any {
	return map[string]any{
		"repository": "github.com/faustbrian/example", "directory": ".", "module_path": "github.com/faustbrian/example", "go_version": "1.27.0",
		"kind": "public library", "purpose": "", "lifecycle": "", "releasable": true, "version": "", "tag_prefix": "", "gates": map[string]any{},
		"test_tags": []any{}, "build_tags": []any{}, "required_services": []any{}, "external_runtime_dependencies": []any{}, "interoperability_tools": []any{},
		"conformance_corpora": []any{}, "specifications": []any{}, "owned_dependencies": []any{}, "reverse_owned_dependencies": []any{}, "packages": []any{},
		"family": "", "family_label": "", "family_description": "", "family_order": 0, "goal_status": "", "goal_files": []any{}, "goal_evidence": []any{},
		"provenance": []any{}, "cohesion": nil,
	}
}
