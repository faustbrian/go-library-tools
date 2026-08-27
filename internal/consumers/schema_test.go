package consumers_test

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestPublishedConsumerSchemaIsStrictAndCurrent(t *testing.T) {
	data, err := os.ReadFile("../../schema/consumers.schema.json")
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
	content := string(data)
	for _, required := range []string{`"allOf"`, `"not"`, `"enum": ["deferred", "tooling"]`, `"pattern": "\\S"`, `@\\{`} {
		if !strings.Contains(content, required) {
			t.Errorf("published schema lacks decoder constraint %s", required)
		}
	}
}
