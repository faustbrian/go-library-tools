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
			write(t, filepath.Join(root, "modules.json"), `{"schema_version":2,"repository":"github.com/faustbrian/example","modules":[]}`)
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
    "gates": {"tests": true},
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
