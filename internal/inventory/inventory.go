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
	GoalEvidence                []GoalEvidence  `json:"goal_evidence"`
	Provenance                  json.RawMessage `json:"provenance"`
	Cohesion                    *Cohesion       `json:"cohesion,omitempty"`
}

// Cohesion is the schema-v2 consumer and engineering catalog contract.
type Cohesion struct {
	Family                     string            `json:"family"`
	SecondaryCapabilities      []string          `json:"secondary_capabilities"`
	Responsibility             string            `json:"responsibility"`
	NonGoals                   []string          `json:"non_goals"`
	PublicPackageIdentifier    string            `json:"public_package_identifier"`
	PrimaryEntryPackages       []string          `json:"primary_entry_packages"`
	PackageSelection           map[string]string `json:"package_selection"`
	LifecycleStatus            string            `json:"lifecycle_status"`
	Maturity                   string            `json:"maturity"`
	ConstructionStyles         []string          `json:"construction_styles"`
	LifecycleStyles            []string          `json:"lifecycle_styles"`
	Ownership                  Ownership         `json:"ownership"`
	OptionalOwnedDependencies  []string          `json:"optional_owned_dependencies"`
	Adapters                   []string          `json:"adapters"`
	Companions                 []string          `json:"companions"`
	SupportedGo                SupportedGo       `json:"supported_go"`
	SupportedPlatforms         []string          `json:"supported_platforms"`
	SupportedBackends          []string          `json:"supported_backends"`
	SupportedProtocols         []string          `json:"supported_protocols"`
	Documentation              Documentation     `json:"documentation"`
	KnownGoodCompatibilitySets []string          `json:"known_good_compatibility_sets"`
	Delivery                   Delivery          `json:"delivery"`
}

// Ownership records consumer-visible configuration, alias, resource, and work ownership.
type Ownership struct {
	Configuration    string   `json:"configuration"`
	MutableInputs    []string `json:"mutable_inputs"`
	RuntimeResources string   `json:"runtime_resources"`
	BackgroundWork   string   `json:"background_work"`
}

// SupportedGo records the minimum version and exact versions exercised by CI.
type SupportedGo struct {
	Minimum string   `json:"minimum"`
	Tested  []string `json:"tested"`
}

// Documentation records entry points; nil means the entry point is inapplicable.
type Documentation struct {
	README         *string `json:"readme"`
	API            *string `json:"api"`
	Adoption       *string `json:"adoption"`
	Security       *string `json:"security"`
	Compatibility  *string `json:"compatibility"`
	Performance    *string `json:"performance"`
	Examples       *string `json:"examples"`
	FAQ            *string `json:"faq"`
	Changelog      *string `json:"changelog"`
	PkgGoDev       *string `json:"pkg_go_dev"`
	EcosystemIndex *string `json:"ecosystem_index"`
}

// Delivery keeps implementation, hardening, and release states independent.
type Delivery struct {
	Implementation string `json:"implementation"`
	Hardening      string `json:"hardening"`
	Release        string `json:"release"`
}

// GoalEvidence binds a goal document to its implementation and verification
// claims without making those claims part of gate selection.
type GoalEvidence struct {
	File                   string   `json:"file"`
	RequirementsSHA256     string   `json:"requirements_sha256"`
	ImplementationEvidence []string `json:"implementation_evidence"`
	VerificationGates      []string `json:"verification_gates"`
	ImplementationStatus   string   `json:"implementation_status"`
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
	catalog, err := load(root, policy, nil)
	if err != nil {
		return Inventory{}, err
	}
	return catalog, nil
}

// LoadSnapshot returns the validated inventory and the exact module-manifest
// bytes from which it was decoded so callers can perform additional checks
// without racing a second filesystem read.
func LoadSnapshot(root string, policy config.Config) (Inventory, []byte, error) {
	moduleManifest, err := repositoryfile.Read(root, policy.Manifests.Modules, maximumManifestSize)
	if err != nil {
		return Inventory{}, nil, fmt.Errorf("load module manifest: %w", err)
	}
	catalog, err := load(root, policy, moduleManifest)
	if err != nil {
		if catalog.SchemaVersion == 0 && catalog.Repository == "" && catalog.Modules == nil {
			return Inventory{}, nil, err
		}
		return catalog, moduleManifest, err
	}
	return catalog, moduleManifest, nil
}

func load(root string, policy config.Config, moduleManifest []byte) (Inventory, error) {
	var modules Inventory
	var err error
	if moduleManifest == nil {
		err = decode(root, policy.Manifests.Modules, &modules)
	} else {
		err = decodeData(moduleManifest, &modules)
	}
	if err != nil {
		return Inventory{}, fmt.Errorf("load module manifest: %w", err)
	}
	var packages packageInventory
	if err := decode(root, policy.Manifests.Packages, &packages); err != nil {
		return modules, fmt.Errorf("load package manifest: %w", err)
	}
	if (modules.SchemaVersion != 1 && modules.SchemaVersion != 2) || packages.SchemaVersion != 1 {
		return modules, errors.New("module manifest schema_version must be 1 or 2 and package manifest schema_version must be 1")
	}
	if modules.SchemaVersion == 1 {
		for _, module := range modules.Modules {
			if module.Cohesion != nil {
				return modules, errors.New("module manifest schema_version 1 must not contain cohesion metadata")
			}
		}
	}
	if modules.Repository == "" || modules.Repository != packages.Repository {
		return modules, errors.New("module and package manifest repository identities differ")
	}
	if len(modules.Modules) == 0 {
		return modules, errors.New("module manifest contains no modules")
	}
	byDirectory := make(map[string]Module, len(modules.Modules))
	for _, module := range modules.Modules {
		if _, exists := byDirectory[module.Directory]; exists {
			return modules, fmt.Errorf("duplicate module directory %q", module.Directory)
		}
		byDirectory[module.Directory] = module
	}
	if err := validatePackages(modules.Modules, packages.Packages); err != nil {
		return modules, err
	}
	gateKeys := map[string]string{
		"api": "api_compatibility", "benchmark": "benchmarks",
		"conformance": "conformance", "docs": "documentation", "fuzz": "fuzz",
		"test": "tests",
	}
	apiOwners := make(map[string]struct{}, len(policy.API.Baselines))
	for index, migration := range policy.Mutation.Imports {
		module, exists := byDirectory[migration.Module]
		if !exists {
			return modules, fmt.Errorf("mutation.imports[%d] references unknown module %q", index, migration.Module)
		}
		if !module.Gates["mutation"] {
			return modules, fmt.Errorf("mutation.imports[%d] is not enabled for module %q", index, migration.Module)
		}
	}
	for index, baseline := range policy.API.Baselines {
		module, exists := byDirectory[baseline.Module]
		if !exists {
			return modules, fmt.Errorf("api.baselines[%d] references unknown module %q", index, baseline.Module)
		}
		if !module.Gates["api_compatibility"] {
			return modules, fmt.Errorf("api.baselines[%d] is not enabled for module %q", index, baseline.Module)
		}
		apiOwners[baseline.Module] = struct{}{}
	}
	declaredOperations := make(map[string]struct{}, len(policy.Operations))
	for index, operation := range policy.Operations {
		module, exists := byDirectory[operation.Module]
		if !exists {
			return modules, fmt.Errorf("operations[%d] references unknown module %q", index, operation.Module)
		}
		enabled := operation.Gate == "interoperability" && len(module.InteroperabilityTools) > 0
		if operation.Gate != "interoperability" {
			enabled = module.Gates[gateKeys[operation.Gate]]
		}
		if !enabled {
			return modules, fmt.Errorf("operations[%d] gate %q is not enabled for module %q", index, operation.Gate, operation.Module)
		}
		if operation.Gate == "api" {
			if _, exists := apiOwners[operation.Module]; exists {
				return modules, fmt.Errorf("module %q has two API gate owners", operation.Module)
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
				return modules, fmt.Errorf("module %q enables %s and requires a typed operation", module.Directory, gate)
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
	return decodeData(data, destination)
}

func decodeData(data []byte, destination any) error {
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
