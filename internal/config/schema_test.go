package config_test

import (
	"encoding/json"
	"os"
	"testing"
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
