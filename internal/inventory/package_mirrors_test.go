package inventory_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/faustbrian/go-library-tools/v2/internal/config"
	"github.com/faustbrian/go-library-tools/v2/internal/inventory"
)

func TestLoadDerivesOnlyOmittedPackageMirrors(t *testing.T) {
	for _, version := range []int{1, 2, 3} {
		for _, mode := range []string{"omitted", "mixed", "explicit"} {
			t.Run(fmt.Sprintf("v%d/%s", version, mode), func(t *testing.T) {
				root, document, modules, canonical := packageMirrorFixture(t, version)
				for index, module := range modules {
					if mode == "omitted" || mode == "mixed" && index != 0 {
						delete(module, "packages")
					}
				}
				raw := writePackageMirrorFixture(t, root, document, canonical)
				catalog, snapshot, err := inventory.LoadSnapshot(root, packageMirrorPolicy())
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(snapshot, raw) {
					t.Fatal("expanded inventory replaced original manifest snapshot")
				}
				for index, want := range [][]inventory.Package{canonical[:2], canonical[2:], {}} {
					if !reflect.DeepEqual(catalog.Modules[index].Packages, want) {
						t.Fatalf("module %d lost canonical package classification", index)
					}
				}
				encoded, err := json.Marshal(catalog)
				if err != nil {
					t.Fatal(err)
				}
				var expanded inventory.Inventory
				if err := json.Unmarshal(encoded, &expanded); err != nil || len(expanded.Modules) != 3 ||
					!reflect.DeepEqual(expanded.Modules[0].Packages, canonical[:2]) ||
					!reflect.DeepEqual(expanded.Modules[1].Packages, canonical[2:]) || expanded.Modules[2].Packages == nil {
					t.Fatal("public inventory JSON omitted expanded package records")
				}
			})
		}
	}
}

func TestLoadRefusesInvalidPackageMirrorsAndCanonicalOwners(t *testing.T) {
	for _, version := range []int{1, 2, 3} {
		for _, scenario := range []string{"empty", "null", "partial", "conflicting", "duplicate mirror", "duplicate canonical", "orphan canonical"} {
			t.Run(fmt.Sprintf("v%d/%s", version, scenario), func(t *testing.T) {
				root, document, modules, canonical := packageMirrorFixture(t, version)
				for _, module := range modules {
					delete(module, "packages")
				}
				switch scenario {
				case "empty":
					modules[0]["packages"] = []inventory.Package{}
				case "null":
					modules[0]["packages"] = nil
				case "partial":
					modules[0]["packages"] = canonical[:1]
				case "conflicting":
					changed := append([]inventory.Package{}, canonical[:2]...)
					changed[0].CoverageRequired = false
					modules[0]["packages"] = changed
				case "duplicate mirror":
					modules[0]["packages"] = []inventory.Package{canonical[0], canonical[0], canonical[1]}
				case "duplicate canonical":
					canonical = append(canonical, canonical[0])
				case "orphan canonical":
					canonical[0].ModuleDirectory = "orphan"
				}
				writePackageMirrorFixture(t, root, document, canonical)
				if _, err := inventory.Load(root, packageMirrorPolicy()); err == nil {
					t.Fatal("invalid mirror or canonical ownership was accepted")
				}
			})
		}
	}
}

func TestLoadRejectsIncompleteCanonicalRecordsBeforeDerivation(t *testing.T) {
	for _, version := range []int{1, 2, 3} {
		for _, field := range []string{"module_directory", "directory", "name", "import_path", "kind", "production", "executable", "coverage_required", "build_required", "build_tags"} {
			for _, mode := range []string{"missing", "null"} {
				t.Run(fmt.Sprintf("v%d/%s/%s", version, field, mode), func(t *testing.T) {
					root, document, modules, canonical := packageMirrorFixture(t, version)
					delete(modules[0], "packages")
					writePackageMirrorFixture(t, root, document, canonical)
					data, err := json.Marshal(canonical)
					if err != nil {
						t.Fatal(err)
					}
					var records []map[string]any
					if err := json.Unmarshal(data, &records); err != nil {
						t.Fatal(err)
					}
					if mode == "missing" {
						delete(records[0], field)
					} else {
						records[0][field] = nil
					}
					data, err = json.Marshal(map[string]any{"schema_version": 1, "repository": "github.com/faustbrian/example", "packages": records})
					if err != nil {
						t.Fatal(err)
					}
					write(t, filepath.Join(root, "packages.json"), string(data))
					if _, err := inventory.Load(root, packageMirrorPolicy()); err == nil {
						t.Fatal("derivation inferred an absent or null classification")
					}
				})
			}
		}
		for _, mode := range []string{"missing", "null"} {
			t.Run(fmt.Sprintf("v%d/packages/%s", version, mode), func(t *testing.T) {
				root, document, modules, canonical := packageMirrorFixture(t, version)
				for _, module := range modules {
					delete(module, "packages")
				}
				writePackageMirrorFixture(t, root, document, canonical)
				packageDocument := map[string]any{"schema_version": 1, "repository": "github.com/faustbrian/example"}
				if mode == "null" {
					packageDocument["packages"] = nil
				}
				data, err := json.Marshal(packageDocument)
				if err != nil {
					t.Fatal(err)
				}
				write(t, filepath.Join(root, "packages.json"), string(data))
				if _, err := inventory.Load(root, packageMirrorPolicy()); err == nil {
					t.Fatal("derivation inferred an absent or null canonical inventory")
				}
			})
		}
	}
}

func TestLoadPreservesExplicitLegacyCanonicalDefaults(t *testing.T) {
	for _, version := range []int{1, 2, 3} {
		for _, mode := range []string{"missing", "null"} {
			for _, unrelatedOmission := range []bool{false, true} {
				t.Run(fmt.Sprintf("v%d/%s/unrelated-omission=%t", version, mode, unrelatedOmission), func(t *testing.T) {
					root, document, modules, canonical := packageMirrorFixture(t, version)
					canonical[0].CoverageRequired = false
					if unrelatedOmission {
						delete(modules[2], "packages")
					}
					writePackageMirrorFixture(t, root, document, canonical)
					data, err := json.Marshal(canonical)
					if err != nil {
						t.Fatal(err)
					}
					var records []map[string]any
					if err := json.Unmarshal(data, &records); err != nil {
						t.Fatal(err)
					}
					if mode == "missing" {
						delete(records[0], "coverage_required")
					} else {
						records[0]["coverage_required"] = nil
					}
					data, err = json.Marshal(map[string]any{"schema_version": 1, "repository": "github.com/faustbrian/example", "packages": records})
					if err != nil {
						t.Fatal(err)
					}
					write(t, filepath.Join(root, "packages.json"), string(data))
					catalog, err := inventory.Load(root, packageMirrorPolicy())
					if err != nil || len(catalog.Modules[0].Packages) != 2 || catalog.Modules[0].Packages[0].CoverageRequired {
						t.Fatal("explicit legacy mirror admission changed")
					}
				})
			}
		}
	}
}

func packageMirrorFixture(t *testing.T, version int) (string, map[string]any, []map[string]any, []inventory.Package) {
	t.Helper()
	canonical := []inventory.Package{
		{ModuleDirectory: ".", Directory: ".", Name: "example", ImportPath: "github.com/faustbrian/example", Kind: "public", Production: true, Executable: true, CoverageRequired: true, BuildRequired: true, BuildTags: []string{}},
		{ModuleDirectory: ".", Directory: "internal/helper", Name: "helper", ImportPath: "github.com/faustbrian/example/internal/helper", Kind: "internal", Production: true, Executable: true, CoverageRequired: true, BuildRequired: true, BuildTags: []string{}},
		{ModuleDirectory: "adapter", Directory: "adapter", Name: "adapter", ImportPath: "github.com/faustbrian/example/adapter", Kind: "public", Production: true, Executable: true, CoverageRequired: true, BuildRequired: true, BuildTags: []string{}},
	}
	var document map[string]any
	data := identityManifest(version,
		identityModule(version, ".", "github.com/faustbrian/example"),
		identityModule(version, "adapter", "github.com/faustbrian/example/adapter"),
		identityModule(version, "empty", "github.com/faustbrian/example/empty"))
	if err := json.Unmarshal([]byte(data), &document); err != nil {
		t.Fatal(err)
	}
	modules := make([]map[string]any, 0, 3)
	for index, raw := range document["modules"].([]any) {
		module := raw.(map[string]any)
		module["packages"] = []inventory.Package{}
		switch index {
		case 0:
			module["packages"] = canonical[:2]
		case 1:
			module["packages"] = canonical[2:]
		}
		modules = append(modules, module)
	}
	return t.TempDir(), document, modules, canonical
}

func writePackageMirrorFixture(t *testing.T, root string, document map[string]any, canonical []inventory.Package) []byte {
	t.Helper()
	moduleData, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	packageData, err := json.Marshal(map[string]any{"schema_version": 1, "repository": "github.com/faustbrian/example", "packages": canonical})
	if err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(root, "modules.json"), string(moduleData))
	write(t, filepath.Join(root, "packages.json"), string(packageData))
	return moduleData
}

func packageMirrorPolicy() config.Config {
	return config.Config{Manifests: config.Manifests{Modules: "modules.json", Packages: "packages.json"}}
}

// Retain the legacy distinction for package-free schema1 owners: null was
// already accepted there, but may not be used to derive nonempty packages.
func TestLoadPreservesLegacyPackageFreeNullMirror(t *testing.T) {
	root := fixture(t)
	write(t, filepath.Join(root, "modules.json"), strings.Replace(identityManifest(1, identityModule(1, ".", "github.com/faustbrian/example")), `"packages":[]`, `"packages":null`, 1))
	if _, err := inventory.Load(root, packageMirrorPolicy()); err != nil {
		t.Fatal(err)
	}
}
