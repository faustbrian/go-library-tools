package evidence_test

import (
	"encoding/json"
	"os"
	"testing"
)

func TestPublishedEvidenceSchemaIsStrictAndCurrent(t *testing.T) {
	data, err := os.ReadFile("../../schema/evidence.schema.json")
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
		t.Fatalf("published schema does not match evidence contract: %#v", schema)
	}
}
