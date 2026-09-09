package inventory

import (
	"os"
	"testing"
)

func TestGeneratedSchemaV3MatchesPublishedSchema(t *testing.T) {
	want, err := os.ReadFile("../../schema/modules-v3.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	if modulesV3SchemaJSON != string(want) {
		t.Fatal("embedded schema v3 is stale; run go generate ./internal/inventory")
	}
}
