package cohesion

import (
	"encoding/json"
	"fmt"
)

var requiredCohesionFields = []string{
	"family", "secondary_capabilities", "responsibility", "non_goals",
	"public_package_identifier", "primary_entry_packages", "package_selection",
	"lifecycle_status", "maturity", "construction_styles", "lifecycle_styles",
	"ownership", "optional_owned_dependencies", "adapters", "companions",
	"supported_go", "supported_platforms", "supported_backends",
	"supported_protocols", "documentation", "known_good_compatibility_sets", "delivery",
}

func requiredMetadataDiagnosticsData(data []byte) []Diagnostic {
	var manifest struct {
		SchemaVersion int               `json:"schema_version"`
		Modules       []json.RawMessage `json:"modules"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil || manifest.SchemaVersion != 2 {
		return nil
	}
	diagnostics := make([]Diagnostic, 0)
	for index, rawModule := range manifest.Modules {
		var module map[string]json.RawMessage
		if json.Unmarshal(rawModule, &module) != nil {
			continue
		}
		rawMetadata, exists := module["cohesion"]
		if !exists || string(rawMetadata) == "null" {
			continue
		}
		var metadata map[string]json.RawMessage
		if json.Unmarshal(rawMetadata, &metadata) != nil {
			continue
		}
		base := fmt.Sprintf("/modules/%d/cohesion", index)
		appendMissingFields(&diagnostics, metadata, requiredCohesionFields, base)
		appendMissingObjectFields(&diagnostics, metadata, "ownership", []string{"configuration", "mutable_inputs", "runtime_resources", "background_work"}, base)
		appendMissingObjectFields(&diagnostics, metadata, "supported_go", []string{"minimum", "tested"}, base)
		appendMissingObjectFields(&diagnostics, metadata, "documentation", []string{"readme", "api", "adoption", "security", "compatibility", "performance", "examples", "faq", "changelog", "pkg_go_dev", "ecosystem_index"}, base)
		appendMissingObjectFields(&diagnostics, metadata, "delivery", []string{"implementation", "hardening", "release"}, base)
	}
	return diagnostics
}

func appendMissingObjectFields(diagnostics *[]Diagnostic, parent map[string]json.RawMessage, name string, fields []string, base string) {
	raw, exists := parent[name]
	if !exists {
		return
	}
	var object map[string]json.RawMessage
	if json.Unmarshal(raw, &object) != nil {
		return
	}
	appendMissingFields(diagnostics, object, fields, base+"/"+name)
}

func appendMissingFields(diagnostics *[]Diagnostic, object map[string]json.RawMessage, fields []string, base string) {
	for _, field := range fields {
		if _, exists := object[field]; !exists {
			*diagnostics = append(*diagnostics, Diagnostic{
				Code: "missing-metadata", Path: base + "/" + field,
				Message: "required cohesion metadata field is missing",
			})
		}
	}
}
