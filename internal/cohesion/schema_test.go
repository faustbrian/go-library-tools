package cohesion_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestPublishedCohesionSchemasAreClosedAndVersioned(t *testing.T) {
	tests := []struct {
		file       string
		definition string
	}{
		{"modules.schema.json", "cohesion"},
		{"cohesion-check.schema.json", "diagnostic"},
		{"cohesion-catalog.schema.json", "consumer_module"},
		{"cohesion-inputs.schema.json", "repository_input"},
	}
	for _, test := range tests {
		t.Run(test.file, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join("..", "..", "schema", test.file))
			if err != nil {
				t.Fatal(err)
			}
			var schema struct {
				Schema               string                     `json:"$schema"`
				AdditionalProperties bool                       `json:"additionalProperties"`
				Defs                 map[string]json.RawMessage `json:"$defs"`
			}
			if err := json.Unmarshal(data, &schema); err != nil {
				t.Fatal(err)
			}
			if schema.Schema == "" || schema.AdditionalProperties || schema.Defs[test.definition] == nil {
				t.Fatalf("schema contract = %#v", schema)
			}
		})
	}
}
