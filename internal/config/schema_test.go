package config_test

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

func TestPublishedSchemaIsStrictAndCurrent(t *testing.T) {
	data, err := os.ReadFile("../../schema/golib.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	var schema struct {
		AdditionalProperties bool `json:"additionalProperties"`
		Properties           map[string]struct {
			Const int `json:"const"`
		} `json:"properties"`
	}
	if err := json.Unmarshal(data, &schema); err != nil {
		t.Fatal(err)
	}
	if schema.AdditionalProperties || schema.Properties["schema_version"].Const != 1 {
		t.Fatalf("published schema does not match decoder contract: %#v", schema)
	}
}

func TestPublishedSchemaAdmitsBoundedRiskChecks(t *testing.T) {
	data, err := os.ReadFile("../../schema/golib.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	document, err := jsonschema.UnmarshalJSON(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	compiler := jsonschema.NewCompiler()
	const identity = "https://example.invalid/golib.schema.json"
	if err := compiler.AddResource(identity, document); err != nil {
		t.Fatal(err)
	}
	schema, err := compiler.Compile(identity)
	if err != nil {
		t.Fatal(err)
	}
	for _, scenario := range []struct {
		name    string
		change  func(map[string]any)
		invalid bool
	}{
		{"focused race", func(map[string]any) {}, false},
		{"explicit nested package", func(step map[string]any) { step["packages"] = []any{"./adapter"} }, false},
		{"default count", func(step map[string]any) { step["count"] = 0 }, false},
		{"recursive packages", func(step map[string]any) { step["packages"] = []any{"./..."} }, true},
		{"embedded wildcard", func(step map[string]any) { step["packages"] = []any{"./integration..."} }, true},
		{"several packages", func(step map[string]any) { step["packages"] = []any{".", "./other"} }, true},
		{"unanchored selector", func(step map[string]any) { step["run"] = "TestContract" }, true},
		{"grouped alternatives", func(step map[string]any) { step["run"] = "^Test(Commit|Rollback)$" }, false},
		{"escaping alternative", func(step map[string]any) { step["run"] = "^TestContract$|^FuzzOther$" }, true},
		{"partially anchored alternative", func(step map[string]any) { step["run"] = "^TestContract|FuzzOther$" }, true},
		{"subtest selector", func(step map[string]any) { step["run"] = "^TestContract/case$" }, true},
		{"missing timeout", func(step map[string]any) { delete(step, "timeout") }, true},
		{"stress count", func(step map[string]any) { step["count"] = 20 }, true},
		{"fuzz campaign", func(step map[string]any) { step["fuzz"] = "^FuzzContract$" }, true},
		{"make campaign", func(step map[string]any) {
			step["type"] = "make"
			step["makefile"] = "verification/package.mk"
			step["target"] = "stress"
		}, true},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			step := map[string]any{"type": "go-test", "packages": []any{"."}, "run": "^TestContract$", "timeout": "1m", "count": 1, "race": true}
			scenario.change(step)
			policy := map[string]any{"schema_version": 1, "tool_version": "v2.0.0", "operations": []any{map[string]any{"module": ".", "gate": "local-test", "steps": []any{step}}}}
			if err := schema.Validate(policy); (err != nil) != scenario.invalid {
				t.Fatalf("schema admission: %v", err)
			}
		})
	}
}
