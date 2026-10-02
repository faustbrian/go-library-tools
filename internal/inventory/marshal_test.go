package inventory_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/faustbrian/go-library-tools/v2/internal/inventory"
)

func TestMarshalInventoryPreservesLegacyDeliveryMetadata(t *testing.T) {
	value := inventory.Inventory{SchemaVersion: 2, Modules: []inventory.Module{{
		Directory: ".", GoalStatus: "implemented", GoalFiles: []string{"goal.md"},
		Provenance: json.RawMessage(`{"source":"published"}`),
		Cohesion:   &inventory.Cohesion{Delivery: inventory.Delivery{Release: "published"}},
	}}}
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	var decoded inventory.Inventory
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.SchemaVersion != 2 || len(decoded.Modules) != 1 {
		t.Fatalf("legacy inventory = %#v", decoded)
	}
	module := decoded.Modules[0]
	if module.GoalStatus != "implemented" || len(module.GoalFiles) != 1 || module.GoalFiles[0] != "goal.md" ||
		string(module.Provenance) != `{"source":"published"}` || module.Cohesion == nil || module.Cohesion.Delivery.Release != "published" {
		t.Fatalf("legacy delivery metadata = %#v", module)
	}
}

func TestMarshalInventoryRejectsMissingSchemaV3Modules(t *testing.T) {
	encoded, err := json.Marshal(inventory.Inventory{SchemaVersion: 3})
	if err == nil || len(encoded) != 0 {
		t.Fatalf("missing modules = %q, %v", encoded, err)
	}
}

func TestMarshalInventoryRetainsInvalidProvenanceClassification(t *testing.T) {
	for _, version := range []int{2, 3} {
		value := inventory.Inventory{SchemaVersion: version, Modules: []inventory.Module{{
			Provenance: json.RawMessage(`{"source":}`),
		}}}
		encoded, err := json.Marshal(value)
		var marshalError *json.MarshalerError
		if !errors.As(err, &marshalError) || len(encoded) != 0 {
			t.Fatalf("schema %d malformed provenance = %q, %v", version, encoded, err)
		}
	}
}

func TestMarshalInventoryPreservesVersionedEmptyOutput(t *testing.T) {
	for _, test := range []struct {
		version int
		want    string
	}{
		{2, `{"schema_version":2,"repository":"","go_version":"","modules":[]}`},
		{3, `{"go_version":"","modules":[],"repository":"","schema_version":3}`},
	} {
		encoded, err := json.Marshal(inventory.Inventory{SchemaVersion: test.version, Modules: []inventory.Module{}})
		if err != nil || string(encoded) != test.want {
			t.Fatalf("schema %d output = %s, %v", test.version, encoded, err)
		}
	}
}
