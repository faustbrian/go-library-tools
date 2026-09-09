//nolint:forcetypeassert // controlled fixture shapes
package cohesion

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestValidateManifestV3RunsStrictSchemaAndManifestSemantics(t *testing.T) {
	t.Parallel()

	requirements := "sha256:6e5ac948c2cf1ec439fdf90c53c0648befb74a65c3eca2eb595737ee8211891c"
	valid := manifestV3Fixture(t, requirements)
	if err := validateManifestV3(valid, requirements); err != nil {
		t.Fatalf("validateManifestV3(valid) error = %v", err)
	}
	for name, repository := range map[string]string{
		"128 ASCII bytes":     strings.Repeat("a", 128),
		"128 multibyte bytes": strings.Repeat("ä", 64),
	} {
		candidate := cloneManifestV3(t, valid)
		candidate["repository"] = repository
		data, err := json.Marshal(candidate)
		if err != nil {
			t.Fatal(err)
		}
		if err := validateManifestV3(data, requirements); err != nil {
			t.Fatalf("validateManifestV3(%s repository) error = %v", name, err)
		}
	}
	for name, mutate := range map[string]func(map[string]any){
		"128-byte legacy family": func(value map[string]any) {
			manifestV3Modules(value)[0]["family"] = strings.Repeat("a", 128)
		},
		"legacy package kind above module-kind bound": func(value map[string]any) {
			manifestV3Modules(value)[0]["packages"].([]any)[0].(map[string]any)["kind"] = strings.Repeat("a", 129)
		},
		"legacy backend above identity bound": func(value map[string]any) {
			manifestV3Cohesion(manifestV3Modules(value)[0])["supported_backends"] = []any{strings.Repeat("a", 129)}
		},
		"legacy protocol above identity bound": func(value map[string]any) {
			manifestV3Cohesion(manifestV3Modules(value)[0])["supported_protocols"] = []any{strings.Repeat("a", 129)}
		},
	} {
		candidate := cloneManifestV3(t, valid)
		mutate(candidate)
		data, err := json.Marshal(candidate)
		if err != nil {
			t.Fatal(err)
		}
		if err := validateManifestV3(data, requirements); err != nil {
			t.Fatalf("validateManifestV3(%s) error = %v", name, err)
		}
	}

	for name, mutate := range map[string]func(map[string]any){
		"unsafe module directory": func(value map[string]any) {
			manifestV3Modules(value)[0]["directory"] = "../module"
		},
		"unsorted provenance": func(value map[string]any) {
			manifestV3Modules(value)[0]["provenance"] = []any{"z.json", "a.json"}
		},
		"duplicate provenance": func(value map[string]any) {
			manifestV3Modules(value)[0]["provenance"] = []any{"a.json", "a.json"}
		},
		"duplicate module directory": func(value map[string]any) {
			modules := value["modules"].([]any)
			value["modules"] = append(modules, cloneManifestV3Module(t, modules[0].(map[string]any)))
		},
		"duplicate integration target": func(value map[string]any) {
			manifestV3Cohesion(manifestV3Modules(value)[0])["integration_roles"] = []any{
				map[string]any{"target": "github.com/example/target", "role": "adapter", "provenance": "authorized", "rationale": "first", "authorization_id": "auth:first"},
				map[string]any{"target": "github.com/example/target", "role": "companion", "provenance": "authorized", "rationale": "second", "authorization_id": "auth:second"},
			}
		},
		"unsorted integration roles": func(value map[string]any) {
			manifestV3Cohesion(manifestV3Modules(value)[0])["integration_roles"] = []any{
				map[string]any{"target": "github.com/example/z", "role": "adapter", "provenance": "authorized", "rationale": "first", "authorization_id": "auth:first"},
				map[string]any{"target": "github.com/example/a", "role": "adapter", "provenance": "authorized", "rationale": "second", "authorization_id": "auth:second"},
			}
		},
		"cross-module duplicate evidence": func(value map[string]any) {
			modules := value["modules"].([]any)
			first := modules[0].(map[string]any)
			second := cloneManifestV3Module(t, first)
			second["directory"] = "second"
			for _, module := range []map[string]any{first, second} {
				delivery := manifestV3Delivery(module)
				delivery["implementation"] = "verified"
				delivery["goal"].(map[string]any)["status"] = "in-progress"
				delivery["evidence"].(map[string]any)["implementation"] = []any{map[string]any{
					"receipt_path":   "evidence/shared.json",
					"receipt_sha256": "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					"entry_id":       "shared",
				}}
			}
			value["modules"] = []any{first, second}
		},
		"wrong frozen requirements": func(value map[string]any) {
			manifestV3Delivery(manifestV3Modules(value)[0])["goal"].(map[string]any)["requirements_sha256"] = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
		},
		"129 ASCII byte repository": func(value map[string]any) {
			value["repository"] = strings.Repeat("a", 129)
		},
		"130 multibyte byte repository": func(value map[string]any) {
			value["repository"] = strings.Repeat("ä", 65)
		},
		"oversized fixed digest token": func(value map[string]any) {
			manifestV3Delivery(manifestV3Modules(value)[0])["goal"].(map[string]any)["requirements_sha256"] = strings.Repeat("a", 72)
		},
	} {
		candidate := cloneManifestV3(t, valid)
		mutate(candidate)
		data, err := json.Marshal(candidate)
		if err != nil {
			t.Fatal(err)
		}
		if err := validateManifestV3(data, requirements); err == nil {
			t.Fatalf("validateManifestV3(%s) error = nil", name)
		}
	}
}

func manifestV3Fixture(t *testing.T, requirements string) []byte {
	t.Helper()
	data, err := os.ReadFile("../../modules.json")
	if err != nil {
		t.Fatal(err)
	}
	result, err := makeManifestV3Fixture(data, requirements)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func makeManifestV3Fixture(data []byte, requirements string) ([]byte, error) {
	var manifest map[string]any
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, err
	}
	manifest["schema_id"] = "urn:golib:cohesion:module-manifest:v3"
	manifest["schema_version"] = 3
	for _, module := range manifestV3Modules(manifest) {
		cohesion := manifestV3Cohesion(module)
		delete(cohesion, "adapters")
		delete(cohesion, "companions")
		cohesion["integration_roles"] = []any{}
		cohesion["delivery"] = map[string]any{
			"goal": map[string]any{
				"id": cohesionGoalID, "requirements_sha256": requirements, "status": "in-progress",
			},
			"implementation": "in-progress", "hardening": "in-progress", "release": "in-progress",
			"evidence": map[string]any{"implementation": []any{}, "hardening": []any{}, "release": []any{}},
		}
	}
	result, err := json.Marshal(manifest)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func cloneManifestV3(t *testing.T, data []byte) map[string]any {
	t.Helper()
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatal(err)
	}
	return result
}

func manifestV3Modules(value map[string]any) []map[string]any {
	modules := value["modules"].([]any)
	result := make([]map[string]any, len(modules))
	for index, module := range modules {
		result[index] = module.(map[string]any)
	}
	return result
}

func manifestV3Cohesion(module map[string]any) map[string]any {
	return module["cohesion"].(map[string]any)
}

func manifestV3Delivery(module map[string]any) map[string]any {
	return manifestV3Cohesion(module)["delivery"].(map[string]any)
}

func cloneManifestV3Module(t *testing.T, module map[string]any) map[string]any {
	t.Helper()
	data, err := json.Marshal(module)
	if err != nil {
		t.Fatal(err)
	}
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatal(err)
	}
	return result
}
