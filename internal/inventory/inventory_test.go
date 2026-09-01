package inventory_test

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/faustbrian/go-library-tools/internal/config"
	"github.com/faustbrian/go-library-tools/internal/inventory"
)

func TestLoadRequiresTypedOperationsForEnabledCustomGates(t *testing.T) {
	tests := []struct {
		name      string
		gates     string
		tools     string
		operation string
	}{
		{"fuzz", `{"fuzz":true}`, `[]`, "fuzz"},
		{"benchmark", `{"benchmarks":true}`, `[]`, "benchmark"},
		{"conformance", `{"conformance":true}`, `[]`, "conformance"},
		{"interoperability", `{}`, `["reference"]`, "interoperability"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := fixture(t)
			write(t, filepath.Join(root, "modules.json"), fmt.Sprintf(`{
  "schema_version":1,
  "repository":"github.com/faustbrian/example",
  "go_version":"1.27.0",
  "modules":[{
    "directory":".",
    "module_path":"github.com/faustbrian/example",
    "go_version":"1.27.0",
    "kind":"public",
    "releasable":true,
    "gates":%s,
    "interoperability_tools":%s,
    "packages":[]
  }]
}`, test.gates, test.tools))
			policy := config.Config{Manifests: config.Manifests{Modules: "modules.json", Packages: "packages.json"}}
			if _, err := inventory.Load(root, policy); err == nil || !strings.Contains(err.Error(), "requires a typed operation") {
				t.Fatalf("Load(missing %s operation) error = %v", test.operation, err)
			}
			policy.Operations = []config.Operation{{Module: ".", Gate: test.operation}}
			if _, err := inventory.Load(root, policy); err != nil {
				t.Fatalf("Load(%s operation) error = %v", test.operation, err)
			}
		})
	}
}

func TestLoadAllowsTypedTestOperationsOnlyForEnabledTests(t *testing.T) {
	root := fixture(t)
	policy := config.Config{
		Manifests:  config.Manifests{Modules: "modules.json", Packages: "packages.json"},
		Operations: []config.Operation{{Module: ".", Gate: "test"}},
	}
	if _, err := inventory.Load(root, policy); err != nil {
		t.Fatalf("Load(enabled test operation) error = %v", err)
	}

	write(t, filepath.Join(root, "modules.json"), `{"schema_version":1,"repository":"github.com/faustbrian/example","go_version":"1.27.0","modules":[{"directory":".","module_path":"github.com/faustbrian/example","go_version":"1.27.0","kind":"public","releasable":true,"gates":{},"packages":[]}]}`)
	if _, err := inventory.Load(root, policy); err == nil || !strings.Contains(err.Error(), `gate "test" is not enabled`) {
		t.Fatalf("Load(disabled test operation) error = %v", err)
	}
}

func TestLoadValidatesTypedAPIBaselineOwnership(t *testing.T) {
	root := fixture(t)
	policy := config.Config{
		Manifests: config.Manifests{Modules: "modules.json", Packages: "packages.json"},
		API:       config.API{Baselines: []config.APIBaseline{{Module: ".", Mode: "apidiff", Path: "api/v1.txt"}}},
	}
	if _, err := inventory.Load(root, policy); err != nil {
		t.Fatalf("Load(enabled API baseline) error = %v", err)
	}
	write(t, filepath.Join(root, "modules.json"), `{"schema_version":1,"repository":"github.com/faustbrian/example","go_version":"1.27.0","modules":[{"directory":".","module_path":"github.com/faustbrian/example","go_version":"1.27.0","kind":"public","releasable":true,"gates":{"tests":true},"packages":[]}]}`)
	if _, err := inventory.Load(root, policy); err == nil || !strings.Contains(err.Error(), "is not enabled") {
		t.Fatalf("Load(disabled API baseline) error = %v", err)
	}
	root = fixture(t)

	policy.API.Baselines[0].Module = "missing"
	if _, err := inventory.Load(root, policy); err == nil || !strings.Contains(err.Error(), "references unknown module") {
		t.Fatalf("Load(unknown API module) error = %v", err)
	}

	policy.API.Baselines[0].Module = "."
	policy.Operations = []config.Operation{{Module: ".", Gate: "api"}}
	if _, err := inventory.Load(root, policy); err == nil || !strings.Contains(err.Error(), "two API gate owners") {
		t.Fatalf("Load(duplicate API owner) error = %v", err)
	}
}

func TestLoadValidatesMutationImportOwnership(t *testing.T) {
	root := fixture(t)
	write(t, filepath.Join(root, "modules.json"), `{"schema_version":1,"repository":"github.com/faustbrian/example","go_version":"1.27.0","modules":[{"directory":".","module_path":"github.com/faustbrian/example","go_version":"1.27.0","kind":"public","releasable":true,"gates":{"mutation":true},"packages":[]}]}`)
	policy := config.Config{
		Manifests: config.Manifests{Modules: "modules.json", Packages: "packages.json"},
		Mutation: config.Mutation{Imports: []config.MutationImport{{
			Module: ".", Archive: "checkpoint.zip", Ledger: "ledger.json",
		}}},
	}
	if _, err := inventory.Load(root, policy); err != nil {
		t.Fatalf("Load(enabled mutation import) error = %v", err)
	}

	policy.Mutation.Imports[0].Module = "missing"
	if _, err := inventory.Load(root, policy); err == nil || !strings.Contains(err.Error(), "references unknown module") {
		t.Fatalf("Load(unknown mutation module) error = %v", err)
	}

	policy.Mutation.Imports[0].Module = "."
	write(t, filepath.Join(root, "modules.json"), `{"schema_version":1,"repository":"github.com/faustbrian/example","go_version":"1.27.0","modules":[{"directory":".","module_path":"github.com/faustbrian/example","go_version":"1.27.0","kind":"public","releasable":true,"gates":{"tests":true},"packages":[]}]}`)
	if _, err := inventory.Load(root, policy); err == nil || !strings.Contains(err.Error(), "is not enabled") {
		t.Fatalf("Load(disabled mutation import) error = %v", err)
	}
}

func TestLoadValidatesCanonicalManifests(t *testing.T) {
	root := fixture(t)

	got, err := inventory.Load(root, config.Config{
		Manifests: config.Manifests{Modules: "modules.json", Packages: "packages.json"},
	})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if got.Repository != "github.com/faustbrian/example" || len(got.Modules) != 1 {
		t.Fatalf("Load() = %#v", got)
	}
	if got.Modules[0].Directory != "." || got.Modules[0].ModulePath != got.Repository {
		t.Fatalf("Load() module = %#v", got.Modules[0])
	}
}

func TestLoadSnapshotReturnsTheExactValidatedModuleManifest(t *testing.T) {
	root := fixture(t)
	policy := config.Config{Manifests: config.Manifests{Modules: "modules.json", Packages: "packages.json"}}
	want, err := os.ReadFile(filepath.Join(root, "modules.json"))
	if err != nil {
		t.Fatal(err)
	}

	catalog, snapshot, err := inventory.LoadSnapshot(root, policy)
	if err != nil {
		t.Fatal(err)
	}
	if catalog.Repository != "github.com/faustbrian/example" || string(snapshot) != string(want) {
		t.Fatalf("LoadSnapshot() = %#v, %q", catalog, snapshot)
	}
}

func TestLoadSnapshotPreservesDecodedModuleIdentityOnPackageFailure(t *testing.T) {
	root := fixture(t)
	policy := config.Config{Manifests: config.Manifests{Modules: "modules.json", Packages: "packages.json"}}
	write(t, filepath.Join(root, "packages.json"), `{"schema_version":1,"repository":"github.com/faustbrian/other","packages":[]}`)

	catalog, snapshot, err := inventory.LoadSnapshot(root, policy)
	if err == nil || catalog.Repository != "github.com/faustbrian/example" || catalog.SchemaVersion != 1 || len(catalog.Modules) != 1 || len(snapshot) == 0 {
		t.Fatalf("LoadSnapshot() = %#v, %q, %v", catalog, snapshot, err)
	}
	loaded, loadErr := inventory.Load(root, policy)
	if loadErr == nil || loaded.Repository != "" {
		t.Fatalf("Load() = %#v, %v", loaded, loadErr)
	}
}

func TestLoadSnapshotPreservesDecodedHeaderWhenValidationFails(t *testing.T) {
	root := fixture(t)
	policy := config.Config{Manifests: config.Manifests{Modules: "modules.json", Packages: "packages.json"}}
	write(t, filepath.Join(root, "modules.json"), `{"schema_version":0,"repository":"github.com/faustbrian/example","go_version":"1.27.0"}`)

	catalog, snapshot, err := inventory.LoadSnapshot(root, policy)
	if err == nil || catalog.Repository != "github.com/faustbrian/example" || catalog.SchemaVersion != 0 || catalog.Modules != nil || len(snapshot) == 0 {
		t.Fatalf("LoadSnapshot() = %#v, %q, %v", catalog, snapshot, err)
	}
}

func TestLoadSnapshotRejectsUnreadableOrInvalidModuleManifest(t *testing.T) {
	policy := config.Config{Manifests: config.Manifests{Modules: "modules.json", Packages: "packages.json"}}
	for _, test := range []struct {
		name string
		root func(*testing.T) string
	}{
		{"unreadable", func(t *testing.T) string { return t.TempDir() }},
		{"invalid", func(t *testing.T) string {
			root := fixture(t)
			write(t, filepath.Join(root, "modules.json"), "{")
			return root
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			catalog, snapshot, err := inventory.LoadSnapshot(test.root(t), policy)
			if err == nil || catalog.Repository != "" || snapshot != nil || !strings.Contains(err.Error(), "load module manifest") {
				t.Fatalf("LoadSnapshot() = %#v, %q, %v", catalog, snapshot, err)
			}
		})
	}
}

func TestLoadAcceptsAndPreservesSchemaV2CohesionMetadata(t *testing.T) {
	root := fixture(t)
	write(t, filepath.Join(root, "modules.json"), `{
  "schema_version": 2,
  "repository": "github.com/faustbrian/example",
  "go_version": "1.27.0",
  "modules": [{
    "directory": ".",
    "module_path": "github.com/faustbrian/example",
    "go_version": "1.27.0",
    "kind": "public tool",
    "purpose": "Example tooling.",
    "lifecycle": "stable",
    "releasable": true,
    "version": "1.0.0",
    "tag_prefix": "v",
    "gates": {},
    "packages": [],
    "family": "tooling",
    "family_label": "Tooling",
    "family_description": "Repository tooling.",
    "family_order": 1,
    "cohesion": {
      "family": "tooling",
      "secondary_capabilities": ["testing-and-conformance"],
      "responsibility": "Validate standalone repositories.",
      "non_goals": ["Own application runtime behavior."],
      "public_package_identifier": "none",
      "primary_entry_packages": ["github.com/faustbrian/example/cmd/example"],
      "package_selection": {"github.com/faustbrian/example/cmd/example": "Run repository checks."},
      "lifecycle_status": "active",
      "maturity": "stable",
      "construction_styles": ["plain-function"],
      "lifecycle_styles": ["stateless"],
      "ownership": {
        "configuration": "caller",
        "mutable_inputs": ["copy"],
        "runtime_resources": "none",
        "background_work": "none"
      },
      "optional_owned_dependencies": [],
      "adapters": [],
      "companions": [],
      "supported_go": {"minimum": "1.27.0", "tested": ["1.27.0"]},
      "supported_platforms": ["portable-go"],
      "supported_backends": [],
      "supported_protocols": [],
      "documentation": {
        "readme": "README.md",
        "api": "https://pkg.go.dev/github.com/faustbrian/example",
        "adoption": null,
        "security": null,
        "compatibility": null,
        "performance": null,
        "examples": null,
        "faq": null,
        "changelog": "CHANGELOG.md",
        "pkg_go_dev": "https://pkg.go.dev/github.com/faustbrian/example",
        "ecosystem_index": "https://example.com/ecosystem"
      },
      "known_good_compatibility_sets": [],
      "delivery": {"implementation": "verified", "hardening": "in-progress", "release": "in-progress"}
    }
  }]
}`)

	got, err := inventory.Load(root, config.Config{
		Manifests: config.Manifests{Modules: "modules.json", Packages: "packages.json"},
	})
	if err != nil {
		t.Fatalf("Load(schema v2) error = %v", err)
	}
	if got.SchemaVersion != 2 || got.Modules[0].Cohesion == nil || got.Modules[0].Cohesion.Family != "tooling" {
		t.Fatalf("Load(schema v2) = %#v", got)
	}
}

func TestLoadRejectsCohesionMetadataInSchemaV1(t *testing.T) {
	root := fixture(t)
	write(t, filepath.Join(root, "modules.json"), `{"schema_version":1,"repository":"github.com/faustbrian/example","go_version":"1.27.0","modules":[{"directory":".","module_path":"github.com/faustbrian/example","go_version":"1.27.0","kind":"public","releasable":true,"gates":{},"packages":[],"cohesion":{}}]}`)
	write(t, filepath.Join(root, "packages.json"), `{"schema_version":1,"repository":"github.com/faustbrian/example","packages":[]}`)
	_, err := inventory.Load(root, config.Config{Manifests: config.Manifests{Modules: "modules.json", Packages: "packages.json"}})
	if err == nil || !strings.Contains(err.Error(), "schema_version 1 must not contain cohesion") {
		t.Fatalf("Load(schema v1 cohesion) error = %v", err)
	}
}

func TestLoadAcceptsStructuredGoalEvidence(t *testing.T) {
	root := fixture(t)
	manifest := `{
  "schema_version": 1,
  "repository": "github.com/faustbrian/example",
  "go_version": "1.27.0",
  "modules": [{
    "directory": ".",
    "module_path": "github.com/faustbrian/example",
    "go_version": "1.27.0",
    "kind": "public",
    "releasable": true,
    "gates": {"api_compatibility": true, "tests": true},
    "goal_evidence": [{
      "file": ".ai/GOAL.md",
      "requirements_sha256": "8943c11690449de8975138c956cd69f95bbb0eec6671cd6109e0b4d49387cfa2",
      "implementation_evidence": ["README.md", "docs/security.md"],
      "verification_gates": ["test", "coverage", "mutation"],
      "implementation_status": "implemented-requires-fresh-verification"
    }],
    "packages": []
  }]
}`
	write(t, filepath.Join(root, "modules.json"), manifest)

	got, err := inventory.Load(root, config.Config{
		Manifests: config.Manifests{Modules: "modules.json", Packages: "packages.json"},
	})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	evidence := got.Modules[0].GoalEvidence
	if len(evidence) != 1 || evidence[0].File != ".ai/GOAL.md" ||
		evidence[0].ImplementationStatus != "implemented-requires-fresh-verification" {
		t.Fatalf("Load() goal evidence = %#v", evidence)
	}

	write(t, filepath.Join(root, "modules.json"), strings.Replace(
		manifest,
		`"implementation_status": "implemented-requires-fresh-verification"`,
		`"implementation_status": "implemented-requires-fresh-verification", "unknown": true`,
		1,
	))
	if _, err := inventory.Load(root, config.Config{
		Manifests: config.Manifests{Modules: "modules.json", Packages: "packages.json"},
	}); err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("Load(unknown goal evidence field) error = %v", err)
	}
}

func TestLoadRejectsManifestMismatch(t *testing.T) {
	root := fixture(t)
	write(t, filepath.Join(root, "packages.json"), `{"schema_version":1,"repository":"github.com/faustbrian/other","packages":[]}`)

	_, err := inventory.Load(root, config.Config{
		Manifests: config.Manifests{Modules: "modules.json", Packages: "packages.json"},
	})
	if err == nil || !strings.Contains(err.Error(), "repository identities differ") {
		t.Fatalf("Load() error = %v", err)
	}
}

func TestLoadRejectsPackageManifestDivergence(t *testing.T) {
	packagePolicy := `{"module_directory":".","directory":".","name":"example","import_path":"github.com/faustbrian/example","kind":"public","production":true,"executable":true,"coverage_required":true,"build_required":true,"build_tags":[]}`
	tests := []struct {
		name     string
		module   string
		packages string
		want     string
	}{
		{"missing canonical", packagePolicy, "", "missing from packages manifest"},
		{"missing module", "", packagePolicy, "missing from module manifest"},
		{"different", packagePolicy, strings.Replace(packagePolicy, `"name":"example"`, `"name":"different"`, 1), "differs"},
		{"wrong module", strings.Replace(packagePolicy, `"module_directory":"."`, `"module_directory":"nested"`, 1), packagePolicy, "module directory"},
		{"duplicate module package", packagePolicy + "," + packagePolicy, packagePolicy, "duplicate module package"},
		{"duplicate canonical package", packagePolicy, packagePolicy + "," + packagePolicy, "duplicate package import"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := fixture(t)
			write(t, filepath.Join(root, "modules.json"), `{"schema_version":1,"repository":"github.com/faustbrian/example","go_version":"1.27.0","modules":[{"directory":".","module_path":"github.com/faustbrian/example","go_version":"1.27.0","kind":"public","releasable":true,"gates":{},"packages":[`+test.module+`]}]}`)
			write(t, filepath.Join(root, "packages.json"), `{"schema_version":1,"repository":"github.com/faustbrian/example","packages":[`+test.packages+`]}`)
			_, err := inventory.Load(root, config.Config{Manifests: config.Manifests{Modules: "modules.json", Packages: "packages.json"}})
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Load() error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestLoadRejectsDuplicateModuleDirectories(t *testing.T) {
	root := fixture(t)
	write(t, filepath.Join(root, "modules.json"), `{"schema_version":1,"repository":"github.com/faustbrian/example","go_version":"1.27.0","modules":[{"directory":".","module_path":"github.com/faustbrian/example","packages":[]},{"directory":".","module_path":"github.com/faustbrian/example","packages":[]}]}`)
	_, err := inventory.Load(root, config.Config{Manifests: config.Manifests{Modules: "modules.json", Packages: "packages.json"}})
	if err == nil || !strings.Contains(err.Error(), "duplicate module directory") {
		t.Fatalf("Load() error = %v", err)
	}
}

func TestLoadRejectsUnknownManifestFields(t *testing.T) {
	root := fixture(t)
	write(t, filepath.Join(root, "packages.json"), `{"schema_version":1,"repository":"github.com/faustbrian/example","packages":[],"extra":true}`)

	_, err := inventory.Load(root, config.Config{
		Manifests: config.Manifests{Modules: "modules.json", Packages: "packages.json"},
	})
	if err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("Load() error = %v", err)
	}
}

func TestLoadRejectsInvalidManifestContracts(t *testing.T) {
	tests := map[string]func(*testing.T, string){
		"unsupported module schema": func(t *testing.T, root string) {
			write(t, filepath.Join(root, "modules.json"), `{"schema_version":3,"repository":"github.com/faustbrian/example","modules":[]}`)
		},
		"unsupported package schema": func(t *testing.T, root string) {
			write(t, filepath.Join(root, "packages.json"), `{"schema_version":2,"repository":"github.com/faustbrian/example","packages":[]}`)
		},
		"empty repository": func(t *testing.T, root string) {
			write(t, filepath.Join(root, "packages.json"), `{"schema_version":1,"repository":"","packages":[]}`)
			write(t, filepath.Join(root, "modules.json"), `{"schema_version":1,"repository":"","modules":[]}`)
		},
		"no modules": func(t *testing.T, root string) {
			write(t, filepath.Join(root, "modules.json"), `{"schema_version":1,"repository":"github.com/faustbrian/example","modules":[]}`)
		},
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			root := fixture(t)
			mutate(t, root)
			_, err := inventory.Load(root, config.Config{Manifests: config.Manifests{Modules: "modules.json", Packages: "packages.json"}})
			if err == nil {
				t.Fatal("Load() error = nil")
			}
		})
	}
}

func TestLoadReportsManifestInputFailures(t *testing.T) {
	tests := map[string]func(*testing.T, string){
		"missing modules": func(t *testing.T, root string) {
			if err := os.Remove(filepath.Join(root, "modules.json")); err != nil {
				t.Fatal(err)
			}
		},
		"missing packages": func(t *testing.T, root string) {
			if err := os.Remove(filepath.Join(root, "packages.json")); err != nil {
				t.Fatal(err)
			}
		},
		"malformed modules": func(t *testing.T, root string) {
			write(t, filepath.Join(root, "modules.json"), "{")
		},
		"multiple package values": func(t *testing.T, root string) {
			write(t, filepath.Join(root, "packages.json"), `{"schema_version":1,"repository":"github.com/faustbrian/example","packages":[]} {}`)
		},
		"malformed trailing value": func(t *testing.T, root string) {
			write(t, filepath.Join(root, "packages.json"), `{"schema_version":1,"repository":"github.com/faustbrian/example","packages":[]} {`)
		},
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			root := fixture(t)
			mutate(t, root)
			_, err := inventory.Load(root, config.Config{Manifests: config.Manifests{Modules: "modules.json", Packages: "packages.json"}})
			if err == nil {
				t.Fatal("Load() error = nil")
			}
		})
	}
}

func TestLoadPropagatesReadFailure(t *testing.T) {
	root := fixture(t)
	if err := os.Remove(filepath.Join(root, "packages.json")); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, "packages.json"), 0o700); err != nil {
		t.Fatal(err)
	}
	_, err := inventory.Load(root, config.Config{Manifests: config.Manifests{Modules: "modules.json", Packages: "packages.json"}})
	if err == nil || errors.Is(err, os.ErrNotExist) {
		t.Fatalf("Load() error = %v", err)
	}
}

func TestLoadValidatesOperationModuleAndGateReferences(t *testing.T) {
	tests := []struct {
		name      string
		operation config.Operation
		want      string
	}{
		{"unknown module", config.Operation{Module: "missing", Gate: "docs"}, "unknown module"},
		{"disabled gate", config.Operation{Module: ".", Gate: "conformance"}, "not enabled"},
		{"interoperability without tools", config.Operation{Module: ".", Gate: "interoperability"}, "not enabled"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := fixture(t)
			_, err := inventory.Load(root, config.Config{
				Manifests:  config.Manifests{Modules: "modules.json", Packages: "packages.json"},
				Operations: []config.Operation{test.operation},
			})
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Load() error = %v", err)
			}
		})
	}
}

func TestLoadAcceptsEnabledInteroperabilityOperation(t *testing.T) {
	root := fixture(t)
	write(t, filepath.Join(root, "modules.json"), `{
  "schema_version": 1,
  "repository": "github.com/faustbrian/example",
  "go_version": "1.27.0",
  "modules": [{
    "directory": ".",
    "module_path": "github.com/faustbrian/example",
    "go_version": "1.27.0",
    "kind": "public",
    "releasable": true,
    "gates": {},
    "interoperability_tools": ["reference"],
    "packages": []
  }]
}`)
	_, err := inventory.Load(root, config.Config{
		Manifests:  config.Manifests{Modules: "modules.json", Packages: "packages.json"},
		Operations: []config.Operation{{Module: ".", Gate: "interoperability"}},
	})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
}

func fixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	write(t, filepath.Join(root, "go.mod"), "module github.com/faustbrian/example\n\ngo 1.27.0\n")
	write(t, filepath.Join(root, "modules.json"), `{
  "schema_version": 1,
  "repository": "github.com/faustbrian/example",
  "go_version": "1.27.0",
  "modules": [{
    "directory": ".",
    "module_path": "github.com/faustbrian/example",
    "go_version": "1.27.0",
    "kind": "public",
    "releasable": true,
    "gates": {"api_compatibility": true, "tests": true},
    "packages": []
  }]
}`)
	write(t, filepath.Join(root, "packages.json"), `{"schema_version":1,"repository":"github.com/faustbrian/example","packages":[]}`)
	return root
}

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}
