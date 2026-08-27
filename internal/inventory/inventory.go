// Package inventory loads and cross-validates canonical module and package
// manifests without deriving policy from the working tree.
package inventory

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"

	"github.com/faustbrian/go-library-tools/internal/config"
	"github.com/faustbrian/go-library-tools/internal/repositoryfile"
)

const maximumManifestSize = 32 << 20

// Inventory is the validated repository catalog.
type Inventory struct {
	SchemaVersion int      `json:"schema_version"`
	Repository    string   `json:"repository"`
	GoVersion     string   `json:"go_version"`
	Modules       []Module `json:"modules"`
}

// Module contains canonical module policy used by orchestration.
type Module struct {
	Directory                   string          `json:"directory"`
	ModulePath                  string          `json:"module_path"`
	GoVersion                   string          `json:"go_version"`
	Kind                        string          `json:"kind"`
	Purpose                     string          `json:"purpose"`
	Lifecycle                   string          `json:"lifecycle"`
	Releasable                  bool            `json:"releasable"`
	Version                     string          `json:"version"`
	TagPrefix                   string          `json:"tag_prefix"`
	Gates                       map[string]bool `json:"gates"`
	TestTags                    []string        `json:"test_tags"`
	BuildTags                   []string        `json:"build_tags"`
	RequiredServices            []string        `json:"required_services"`
	ExternalRuntimeDependencies []string        `json:"external_runtime_dependencies"`
	InteroperabilityTools       []string        `json:"interoperability_tools"`
	ConformanceCorpora          []string        `json:"conformance_corpora"`
	Specifications              []string        `json:"specifications"`
	OwnedDependencies           []string        `json:"owned_dependencies"`
	ReverseOwnedDependencies    []string        `json:"reverse_owned_dependencies"`
	Packages                    []Package       `json:"packages"`
	Family                      string          `json:"family"`
	FamilyLabel                 string          `json:"family_label"`
	FamilyDescription           string          `json:"family_description"`
	FamilyOrder                 int             `json:"family_order"`
	GoalStatus                  string          `json:"goal_status"`
	GoalFiles                   []string        `json:"goal_files"`
	GoalEvidence                []string        `json:"goal_evidence"`
	Provenance                  json.RawMessage `json:"provenance"`
}

// Package is the canonical package classification shared by both manifests.
type Package struct {
	ModuleDirectory  string   `json:"module_directory"`
	Directory        string   `json:"directory"`
	Name             string   `json:"name"`
	ImportPath       string   `json:"import_path"`
	Kind             string   `json:"kind"`
	Production       bool     `json:"production"`
	Executable       bool     `json:"executable"`
	CoverageRequired bool     `json:"coverage_required"`
	BuildRequired    bool     `json:"build_required"`
	BuildTags        []string `json:"build_tags"`
}

type packageInventory struct {
	SchemaVersion int       `json:"schema_version"`
	Repository    string    `json:"repository"`
	Packages      []Package `json:"packages"`
}

// Load reads both manifests and rejects inconsistent repository identities.
func Load(root string, policy config.Config) (Inventory, error) {
	var modules Inventory
	if err := decode(root, policy.Manifests.Modules, &modules); err != nil {
		return Inventory{}, fmt.Errorf("load module manifest: %w", err)
	}
	var packages packageInventory
	if err := decode(root, policy.Manifests.Packages, &packages); err != nil {
		return Inventory{}, fmt.Errorf("load package manifest: %w", err)
	}
	if modules.SchemaVersion != 1 || packages.SchemaVersion != 1 {
		return Inventory{}, errors.New("manifest schema_version must be 1")
	}
	if modules.Repository == "" || modules.Repository != packages.Repository {
		return Inventory{}, errors.New("module and package manifest repository identities differ")
	}
	if len(modules.Modules) == 0 {
		return Inventory{}, errors.New("module manifest contains no modules")
	}
	byDirectory := make(map[string]Module, len(modules.Modules))
	for _, module := range modules.Modules {
		if _, exists := byDirectory[module.Directory]; exists {
			return Inventory{}, fmt.Errorf("duplicate module directory %q", module.Directory)
		}
		byDirectory[module.Directory] = module
	}
	if err := validatePackages(modules.Modules, packages.Packages); err != nil {
		return Inventory{}, err
	}
	gateKeys := map[string]string{
		"api": "api_compatibility", "benchmark": "benchmarks",
		"conformance": "conformance", "docs": "documentation", "fuzz": "fuzz",
		"test": "tests",
	}
	apiOwners := make(map[string]struct{}, len(policy.API.Baselines))
	for index, baseline := range policy.API.Baselines {
		module, exists := byDirectory[baseline.Module]
		if !exists {
			return Inventory{}, fmt.Errorf("api.baselines[%d] references unknown module %q", index, baseline.Module)
		}
		if !module.Gates["api_compatibility"] {
			return Inventory{}, fmt.Errorf("api.baselines[%d] is not enabled for module %q", index, baseline.Module)
		}
		apiOwners[baseline.Module] = struct{}{}
	}
	declaredOperations := make(map[string]struct{}, len(policy.Operations))
	for index, operation := range policy.Operations {
		module, exists := byDirectory[operation.Module]
		if !exists {
			return Inventory{}, fmt.Errorf("operations[%d] references unknown module %q", index, operation.Module)
		}
		enabled := operation.Gate == "interoperability" && len(module.InteroperabilityTools) > 0
		if operation.Gate != "interoperability" {
			enabled = module.Gates[gateKeys[operation.Gate]]
		}
		if !enabled {
			return Inventory{}, fmt.Errorf("operations[%d] gate %q is not enabled for module %q", index, operation.Gate, operation.Module)
		}
		if operation.Gate == "api" {
			if _, exists := apiOwners[operation.Module]; exists {
				return Inventory{}, fmt.Errorf("module %q has two API gate owners", operation.Module)
			}
		}
		declaredOperations[operation.Module+"\x00"+operation.Gate] = struct{}{}
	}
	for _, module := range modules.Modules {
		required := make([]string, 0, 4)
		for _, gate := range []struct{ operation, manifest string }{
			{"benchmark", "benchmarks"}, {"conformance", "conformance"}, {"fuzz", "fuzz"},
		} {
			if module.Gates[gate.manifest] {
				required = append(required, gate.operation)
			}
		}
		if len(module.InteroperabilityTools) > 0 {
			required = append(required, "interoperability")
		}
		for _, gate := range required {
			if _, exists := declaredOperations[module.Directory+"\x00"+gate]; !exists {
				return Inventory{}, fmt.Errorf("module %q enables %s and requires a typed operation", module.Directory, gate)
			}
		}
	}
	return modules, nil
}

func validatePackages(modules []Module, canonical []Package) error {
	byImportPath := make(map[string]Package, len(canonical))
	for _, packagePolicy := range canonical {
		if _, exists := byImportPath[packagePolicy.ImportPath]; exists {
			return fmt.Errorf("duplicate package import path %q", packagePolicy.ImportPath)
		}
		byImportPath[packagePolicy.ImportPath] = packagePolicy
	}
	seen := make(map[string]struct{}, len(canonical))
	for _, module := range modules {
		for _, packagePolicy := range module.Packages {
			if packagePolicy.ModuleDirectory != module.Directory {
				return fmt.Errorf("package %q module directory does not match %q", packagePolicy.ImportPath, module.Directory)
			}
			if _, exists := seen[packagePolicy.ImportPath]; exists {
				return fmt.Errorf("duplicate module package import path %q", packagePolicy.ImportPath)
			}
			seen[packagePolicy.ImportPath] = struct{}{}
			stored, exists := byImportPath[packagePolicy.ImportPath]
			if !exists {
				return fmt.Errorf("package %q is missing from packages manifest", packagePolicy.ImportPath)
			}
			if !reflect.DeepEqual(stored, packagePolicy) {
				return fmt.Errorf("package %q differs between canonical manifests", packagePolicy.ImportPath)
			}
		}
	}
	for importPath := range byImportPath {
		if _, exists := seen[importPath]; !exists {
			return fmt.Errorf("package %q is missing from module manifest", importPath)
		}
	}
	return nil
}

func decode(root, path string, destination any) error {
	data, err := repositoryfile.Read(root, path, maximumManifestSize)
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("manifest contains multiple JSON values")
		}
		return err
	}
	return nil
}
