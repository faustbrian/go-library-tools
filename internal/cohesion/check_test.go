package cohesion_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/faustbrian/go-library-tools/internal/cohesion"
	"github.com/faustbrian/go-library-tools/internal/config"
)

func TestCheckRequiresSchemaV2Adoption(t *testing.T) {
	root := fixture(t, 1, "")
	report := cohesion.Check(root, policy())

	if report.Valid || report.Repository == nil || *report.Repository != "github.com/faustbrian/example" || report.ManifestSchemaVersion == nil || *report.ManifestSchemaVersion != 1 {
		t.Fatalf("Check(schema v1) = %#v", report)
	}
	if report.Summary.TotalModules == nil || *report.Summary.TotalModules != 1 || report.Summary.ErrorCount != 1 {
		t.Fatalf("Check(schema v1) summary = %#v", report.Summary)
	}
	if len(report.Diagnostics) != 1 || report.Diagnostics[0].Code != "adoption-required" || report.Diagnostics[0].Path != "/schema_version" {
		t.Fatalf("Check(schema v1) diagnostics = %#v", report.Diagnostics)
	}
}

func TestCheckRequiresCohesionMetadataForEveryReleasableModule(t *testing.T) {
	root := fixture(t, 2, "")
	report := cohesion.Check(root, policy())

	if report.Valid || report.Summary.ClassifiedModules == nil || *report.Summary.ClassifiedModules != 0 {
		t.Fatalf("Check(missing cohesion) = %#v", report)
	}
	if len(report.Diagnostics) != 2 || !hasDiagnostic(report, "missing-metadata", "/modules/0/cohesion") || !hasDiagnostic(report, "invalid-manifest", "") {
		t.Fatalf("Check(missing cohesion) diagnostics = %#v", report.Diagnostics)
	}
}

func TestCheckCountsAValidReleasableClassifiedModule(t *testing.T) {
	root := fixture(t, 2, ",\n    \"cohesion\": "+validCohesion)
	report := cohesion.Check(root, policy())

	if !report.Valid || report.Summary.TotalModules == nil || *report.Summary.TotalModules != 1 ||
		report.Summary.ReleasableModules == nil || *report.Summary.ReleasableModules != 1 ||
		report.Summary.ClassifiedModules == nil || *report.Summary.ClassifiedModules != 1 ||
		report.Summary.ErrorCount != 0 || len(report.Diagnostics) != 0 {
		t.Fatalf("Check(valid) = %#v", report)
	}
}

func TestCheckRejectsInvalidCohesionMetadata(t *testing.T) {
	tests := []struct {
		name        string
		old         string
		replacement string
		code        string
		path        string
	}{
		{"family authority", `"family": "tooling"`, `"family": "foundations"`, "invalid-value", "/modules/0/cohesion/family"},
		{"empty responsibility", `"responsibility": "Validate standalone repositories."`, `"responsibility": ""`, "missing-metadata", "/modules/0/cohesion/responsibility"},
		{"invalid maturity", `"maturity": "stable"`, `"maturity": "mature"`, "invalid-value", "/modules/0/cohesion/maturity"},
		{"unsorted capabilities", `"secondary_capabilities": ["testing-and-conformance"]`, `"secondary_capabilities": ["testing-and-conformance", "configuration"]`, "nondeterministic-order", "/modules/0/cohesion/secondary_capabilities"},
		{"minimum Go mismatch", `"minimum": "1.27.0"`, `"minimum": "1.26.0"`, "invalid-value", "/modules/0/cohesion/supported_go/minimum"},
		{"unsafe README", `"readme": "README.md"`, `"readme": "../README.md"`, "unsafe-path", "/modules/0/cohesion/documentation/readme"},
		{"non-HTTPS API", `"api": "https://pkg.go.dev/github.com/faustbrian/example"`, `"api": "http://pkg.go.dev/github.com/faustbrian/example"`, "unsafe-path", "/modules/0/cohesion/documentation/api"},
		{"missing nullable document key", "    \"faq\": null,\n", "", "missing-metadata", "/modules/0/cohesion/documentation/faq"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			metadata := strings.Replace(validCohesion, test.old, test.replacement, 1)
			root := fixture(t, 2, ",\n    \"cohesion\": "+metadata)
			report := cohesion.Check(root, policy())
			if report.Valid || !hasDiagnostic(report, test.code, test.path) {
				t.Fatalf("Check(invalid %s) = %#v", test.name, report)
			}
		})
	}
}

func TestCheckRejectsSchemaV2ManifestThatViolatesNormativeSchema(t *testing.T) {
	for _, test := range []struct {
		name        string
		old         string
		replacement string
	}{
		{"missing canonical field", "\"test_tags\": [],", ""},
		{"invalid canonical value", "\"family_order\": 1", "\"family_order\": -1"},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := fixture(t, 2, ",\n    \"cohesion\": "+validCohesion)
			path := filepath.Join(root, "modules.json")
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			mutated := strings.Replace(string(data), test.old, test.replacement, 1)
			if mutated == string(data) {
				t.Fatalf("fixture does not contain %q", test.old)
			}
			write(t, path, mutated)
			report := cohesion.Check(root, policy())
			if report.Valid || !hasDiagnostic(report, "invalid-manifest", "") {
				t.Fatalf("Check(%s) = %#v", test.name, report)
			}
		})
	}
}

func TestCheckReportsSchemaAndCohesionFailuresTogether(t *testing.T) {
	root := fixture(t, 2, ",\n    \"cohesion\": "+validCohesion)
	path := filepath.Join(root, "modules.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	mutated := strings.Replace(string(data), "\"test_tags\": [],", "", 1)
	mutated = strings.Replace(mutated, `"supported_backends": []`, `"supported_backends": ["Postgres"]`, 1)
	write(t, path, mutated)

	report := cohesion.Check(root, policy())
	if report.Valid || !hasDiagnostic(report, "invalid-manifest", "") ||
		!hasDiagnostic(report, "invalid-value", "/modules/0/cohesion/supported_backends") {
		t.Fatalf("Check(schema and cohesion failures) = %#v", report)
	}
}

func TestCheckPreservesEstablishedModuleCountsWhenPackageManifestFails(t *testing.T) {
	root := fixture(t, 2, ",\n    \"cohesion\": "+validCohesion)
	write(t, filepath.Join(root, "packages.json"), `{"schema_version":1,"repository":"github.com/faustbrian/other","packages":[]}`)
	report := cohesion.Check(root, policy())

	if report.Valid || report.Repository == nil || *report.Repository != "github.com/faustbrian/example" ||
		report.ManifestSchemaVersion == nil || *report.ManifestSchemaVersion != 2 ||
		report.Summary.TotalModules == nil || *report.Summary.TotalModules != 1 ||
		report.Summary.ReleasableModules == nil || *report.Summary.ReleasableModules != 1 ||
		report.Summary.ClassifiedModules == nil || *report.Summary.ClassifiedModules != 1 ||
		report.Summary.ErrorCount != 1 || !hasDiagnostic(report, "invalid-manifest", "") {
		t.Fatalf("Check(package failure) = %#v", report)
	}
}

func TestCheckPreservesFactsEstablishedByDecodedInvalidManifest(t *testing.T) {
	for _, test := range []struct {
		name        string
		old         string
		replacement string
		wantRepo    bool
		wantSchema  *int
	}{
		{"explicit zero schema identity", `"schema_version": 2`, `"schema_version": 0`, true, new(0)},
		{"omitted schema identity", "  \"schema_version\": 2,\n", "", true, nil},
		{"missing repository identity", `"repository": "github.com/faustbrian/example"`, `"repository": ""`, false, new(2)},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := fixture(t, 2, ",\n    \"cohesion\": "+validCohesion)
			path := filepath.Join(root, "modules.json")
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			mutated := strings.Replace(string(data), test.old, test.replacement, 1)
			if mutated == string(data) {
				t.Fatalf("fixture does not contain %q", test.old)
			}
			write(t, path, mutated)
			report := cohesion.Check(root, policy())
			if report.Valid || (report.Repository != nil) != test.wantRepo ||
				(report.ManifestSchemaVersion == nil) != (test.wantSchema == nil) ||
				(report.ManifestSchemaVersion != nil && *report.ManifestSchemaVersion != *test.wantSchema) ||
				report.Summary.TotalModules == nil || *report.Summary.TotalModules != 1 ||
				report.Summary.ReleasableModules == nil || *report.Summary.ReleasableModules != 1 ||
				report.Summary.ClassifiedModules == nil || *report.Summary.ClassifiedModules != 1 ||
				!hasDiagnostic(report, "invalid-manifest", "") {
				t.Fatalf("Check(%s) = %#v", test.name, report)
			}
		})
	}
}

func hasDiagnostic(report cohesion.Report, code, path string) bool {
	for _, diagnostic := range report.Diagnostics {
		if diagnostic.Code == code && diagnostic.Path == path {
			return true
		}
	}
	return false
}

const validCohesion = `{
  "family": "tooling",
  "secondary_capabilities": ["testing-and-conformance"],
  "responsibility": "Validate standalone repositories.",
  "non_goals": ["Own application runtime behavior."],
  "public_package_identifier": "example",
  "primary_entry_packages": ["github.com/faustbrian/example"],
  "package_selection": {},
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
  "delivery": {"implementation": "in-progress", "hardening": "in-progress", "release": "in-progress"}
}`

func policy() config.Config {
	return config.Config{Manifests: config.Manifests{Modules: "modules.json", Packages: "packages.json"}}
}

func fixture(t *testing.T, schemaVersion int, cohesionObject string) string {
	t.Helper()
	root := t.TempDir()
	manifest := `{
  "schema_version": ` + string(rune('0'+schemaVersion)) + `,
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
	"test_tags": [],
	"build_tags": [],
	"required_services": [],
	"external_runtime_dependencies": [],
	"interoperability_tools": [],
	"conformance_corpora": [],
	"specifications": [],
	"owned_dependencies": [],
	"reverse_owned_dependencies": [],
    "packages": [{
      "module_directory": ".",
      "directory": ".",
      "name": "example",
      "import_path": "github.com/faustbrian/example",
      "kind": "public",
      "production": true,
      "executable": true,
      "coverage_required": true,
      "build_required": true,
      "build_tags": []
    }],
    "family": "tooling",
    "family_label": "Tooling",
    "family_description": "Repository tooling.",
	"family_order": 1,
	"goal_status": "active",
	"goal_files": [],
	"goal_evidence": [],
	"provenance": []` + cohesionObject + `
  }]
}`
	write(t, filepath.Join(root, "modules.json"), manifest)
	write(t, filepath.Join(root, "packages.json"), `{"schema_version":1,"repository":"github.com/faustbrian/example","packages":[{"module_directory":".","directory":".","name":"example","import_path":"github.com/faustbrian/example","kind":"public","production":true,"executable":true,"coverage_required":true,"build_required":true,"build_tags":[]}]}`)
	write(t, filepath.Join(root, "README.md"), "# Example\n")
	write(t, filepath.Join(root, "CHANGELOG.md"), "# Changelog\n")
	return root
}

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}
