// Package inventory loads and cross-validates canonical module/package policy.
// Start with Load or LoadSnapshot; Decode consumes bounded immutable Git blobs.
// Explicit mirrors must match and omitted package lists derive from the canonical
// owner, not build-selected discovery. Parsed collections are caller-owned and
// mutable; keep them stable during concurrent readers. No sibling checkout,
// generator execution, gate process or release lifecycle is owned.
package inventory

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path"
	"reflect"
	"strings"
	"unicode"

	"github.com/faustbrian/go-library-tools/v2/internal/config"
	"github.com/faustbrian/go-library-tools/v2/internal/repositoryfile"
	"github.com/santhosh-tekuri/jsonschema/v6"
	"golang.org/x/mod/module"
)

const maximumManifestSize = 32 << 20

// MaximumModules bounds repository-wide orchestration while allowing the
// current multi-module families and backward-compatible manifest schemas.
const MaximumModules = 64
const modulesSchemaIdentity = "https://github.com/faustbrian/go-library-tools/schema/modules.schema.json"
const modulesV3SchemaIdentity = "https://github.com/faustbrian/go-library-tools/schema/modules-v3.schema.json"

//go:generate go run schema_generate_main.go

var compiledModulesV3Schema = compileModuleSchemas(map[string]string{
	modulesSchemaIdentity:   modulesSchemaJSON,
	modulesV3SchemaIdentity: modulesV3SchemaJSON,
}, modulesV3SchemaIdentity)

var compiledPackageSchema = compileModuleSchemas(map[string]string{
	modulesSchemaIdentity: modulesSchemaJSON,
}, modulesSchemaIdentity+"#/$defs/package")

// compileModuleSchemas owns fail-closed compilation of source-controlled schemas.
func compileModuleSchemas(resources map[string]string, target string) *jsonschema.Schema {
	compiler := jsonschema.NewCompiler()
	for identity, source := range resources {
		document, err := jsonschema.UnmarshalJSON(bytes.NewReader([]byte(source)))
		if err != nil {
			panic(err)
		}
		if err := compiler.AddResource(identity, document); err != nil {
			panic(err)
		}
	}
	compiled, err := compiler.Compile(target)
	if err != nil {
		panic(err)
	}
	return compiled
}

// Inventory is the validated repository catalog.
type Inventory struct {
	// SchemaID is optional schema-v3 authority identity; legacy versions reject its
	// presence.
	SchemaID string `json:"schema_id,omitempty"`
	// SchemaVersion selects supported versions 1–3 and their ownership/projection
	// rules.
	SchemaVersion int `json:"schema_version"`
	// Repository is the canonical identity that both manifests must share.
	Repository string `json:"repository"`
	// GoVersion records repository toolchain policy, not a discovered installation.
	GoVersion string `json:"go_version"`
	// Modules is the admitted module catalog; callers must not mutate shared slices
	// while readers use it.
	Modules []Module `json:"modules"`
}

// MarshalJSON keeps legacy output unchanged while ensuring optional schema-v3
// output contains only repository-owned manifest fields.
func (inventory Inventory) MarshalJSON() ([]byte, error) {
	type rawInventory Inventory
	encoded, err := json.Marshal(rawInventory(inventory))
	if err != nil || inventory.SchemaVersion != 3 {
		return encoded, err
	}
	return projectSchemaV3Inventory(encoded)
}

// projectSchemaV3Inventory filters only after the complete typed value has been
// marshaled, preserving RawMessage validation and JSON map output ordering.
func projectSchemaV3Inventory(encoded []byte) ([]byte, error) {
	var document map[string]any
	if err := json.Unmarshal(encoded, &document); err != nil {
		return nil, err
	}
	modules, ok := document["modules"].([]any)
	if !ok {
		return nil, errors.New("schema-v3 inventory modules must be an array")
	}
	for _, rawModule := range modules {
		module, ok := rawModule.(map[string]any)
		if !ok {
			return nil, errors.New("schema-v3 inventory module must be an object")
		}
		for _, field := range []string{"goal_status", "goal_files", "goal_evidence", "provenance"} {
			delete(module, field)
		}
		if metadata, ok := module["cohesion"].(map[string]any); ok {
			delete(metadata, "delivery")
		}
	}
	return json.Marshal(document)
}

// Module contains canonical module policy used by orchestration.
type Module struct {
	// Directory is a canonical repository-relative module root; . denotes the root.
	Directory string `json:"directory"`
	// ModulePath is the validated import identity, including Go's required major
	// suffix.
	ModulePath string `json:"module_path"`
	// GoVersion is the declared minimum module Go version.
	GoVersion string `json:"go_version"`
	// Kind classifies public library, adapter, internal, command or other maintained
	// module purpose.
	Kind string `json:"kind"`
	// Purpose is maintained responsibility prose, not an inferred readiness claim.
	Purpose string `json:"purpose"`
	// Lifecycle is the legacy maintained lifecycle classification.
	Lifecycle string `json:"lifecycle"`
	// Releasable identifies modules eligible for release, independently of gate
	// success.
	Releasable bool `json:"releasable"`
	// Version is the catalog's release version; validation does not publish its tag.
	Version string `json:"version"`
	// TagPrefix selects tag naming for this module's release unit.
	TagPrefix string `json:"tag_prefix"`
	// Gates is the explicit applicability map; disabled work is not passing work.
	Gates map[string]bool `json:"gates"`
	// TestTags is the configured test build-tag set, separate from production
	// BuildTags.
	TestTags []string `json:"test_tags"`
	// BuildTags is the production build-tag set used by orchestration.
	BuildTags []string `json:"build_tags"`
	// RequiredServices is the explicit service dependency list, not permission to
	// start unowned services.
	RequiredServices []string `json:"required_services"`
	// ExternalRuntimeDependencies records infrastructure outside the Go module
	// graph.
	ExternalRuntimeDependencies []string `json:"external_runtime_dependencies"`
	// InteroperabilityTools names maintained peer/tool contracts requiring typed
	// operations.
	InteroperabilityTools []string `json:"interoperability_tools"`
	// ConformanceCorpora names reviewed conformance inputs, not an automatic
	// compliance claim.
	ConformanceCorpora []string `json:"conformance_corpora"`
	// Specifications names specification authorities whose decisions belong to the
	// module.
	Specifications []string `json:"specifications"`
	// OwnedDependencies records mandatory owned module edges for compatibility
	// planning.
	OwnedDependencies []string `json:"owned_dependencies"`
	// ReverseOwnedDependencies records known consumer edges without discovering
	// sibling checkouts.
	ReverseOwnedDependencies []string `json:"reverse_owned_dependencies"`
	// Packages is the canonical package classification; omission derives from
	// packages.json, explicit mirrors must match.
	Packages []Package `json:"packages"`
	// Family is the legacy family identifier that cohesion metadata must match.
	Family string `json:"family"`
	// FamilyLabel is the maintained human-readable family heading.
	FamilyLabel string `json:"family_label"`
	// FamilyDescription is the maintained family scope explanation.
	FamilyDescription string `json:"family_description"`
	// FamilyOrder is catalog grouping metadata, not dependency execution order.
	FamilyOrder int `json:"family_order"`
	// GoalStatus is legacy delivery metadata, not gate selection; schema-v3
	// projection omits it.
	GoalStatus string `json:"goal_status"`
	// GoalFiles names legacy goal evidence; schema-v3 projection omits it.
	GoalFiles []string `json:"goal_files"`
	// GoalEvidence records legacy requirement claims without making them gate
	// policy.
	GoalEvidence []GoalEvidence `json:"goal_evidence"`
	// Provenance preserves legacy raw JSON metadata; schema-v3 output omits it.
	Provenance json.RawMessage `json:"provenance"`
	// Cohesion is optional consumer/engineering metadata; releasable schema-v2+
	// modules require it.
	Cohesion *Cohesion `json:"cohesion,omitempty"`
}

// Cohesion is the schema-v2 consumer and engineering catalog contract.
type Cohesion struct {
	// Family is a frozen family identifier matching Module.Family.
	Family string `json:"family"`
	// SecondaryCapabilities is a sorted unique supported capability set.
	SecondaryCapabilities []string `json:"secondary_capabilities"`
	// Responsibility is reviewed responsibility prose, not inferred from exported
	// names.
	Responsibility string `json:"responsibility"`
	// NonGoals is a nonempty explicit scope-boundary list.
	NonGoals []string `json:"non_goals"`
	// PublicPackageIdentifier is a matching public Go identifier or none for
	// rootless suites.
	PublicPackageIdentifier string `json:"public_package_identifier"`
	// PrimaryEntryPackages is a sorted unique inventoried entry-package list.
	PrimaryEntryPackages []string `json:"primary_entry_packages"`
	// PackageSelection maps inventoried imports to maintained selection guidance.
	PackageSelection map[string]string `json:"package_selection"`
	// LifecycleStatus is planned, active, deprecated or retired, independently of
	// delivery proof.
	LifecycleStatus string `json:"lifecycle_status"`
	// Maturity is experimental, preview or stable; planned entries cannot be stable.
	Maturity string `json:"maturity"`
	// ConstructionStyles is a sorted unique supported construction vocabulary.
	ConstructionStyles []string `json:"construction_styles"`
	// LifecycleStyles is a sorted unique explicit lifecycle vocabulary.
	LifecycleStyles []string `json:"lifecycle_styles"`
	// Ownership records caller/package/shared ownership rather than deriving it from
	// Go types.
	Ownership Ownership `json:"ownership"`
	// OptionalOwnedDependencies is the sorted optional owned dependency set,
	// disjoint from mandatory edges.
	OptionalOwnedDependencies []string `json:"optional_owned_dependencies"`
	// Adapters names sorted owned adapter relationships.
	Adapters []string `json:"adapters"`
	// Companions names sorted companion relationships.
	Companions []string `json:"companions"`
	// SupportedGo records minimum and explicitly tested exact versions.
	SupportedGo SupportedGo `json:"supported_go"`
	// SupportedPlatforms is the sorted maintained platform set, not a test execution
	// result.
	SupportedPlatforms []string `json:"supported_platforms"`
	// SupportedBackends is the sorted declared backend set.
	SupportedBackends []string `json:"supported_backends"`
	// SupportedProtocols is the sorted declared protocol set.
	SupportedProtocols []string `json:"supported_protocols"`
	// Documentation records local or HTTPS guidance references, with nil marking
	// inapplicability.
	Documentation Documentation `json:"documentation"`
	// KnownGoodCompatibilitySets names reviewed compatibility sets; metadata alone
	// does not requalify them.
	KnownGoodCompatibilitySets []string `json:"known_good_compatibility_sets"`
	// Delivery retains legacy separate readiness states; schema-v3 projection omits
	// it.
	Delivery Delivery `json:"delivery"`
}

// Ownership records consumer-visible configuration, alias, resource, and work ownership.
type Ownership struct {
	// Configuration is caller or none for configuration ownership.
	Configuration string `json:"configuration"`
	// MutableInputs lists copy, borrow, transfer or none aliasing contracts; none
	// must stand alone.
	MutableInputs []string `json:"mutable_inputs"`
	// RuntimeResources identifies caller, package, shared or none resource
	// ownership.
	RuntimeResources string `json:"runtime_resources"`
	// BackgroundWork identifies caller, package, shared or none worker ownership.
	BackgroundWork string `json:"background_work"`
}

// SupportedGo records the minimum version and exact versions exercised by CI.
type SupportedGo struct {
	// Minimum must match Module.GoVersion and use an exact version.
	Minimum string `json:"minimum"`
	// Tested is a sorted unique exact-version set, each no lower than Minimum.
	Tested []string `json:"tested"`
}

// Documentation records entry points; nil means the entry point is inapplicable.
type Documentation struct {
	// README is the repository introduction reference, nil when inapplicable.
	README *string `json:"readme"`
	// API is the API guidance reference, nil when inapplicable.
	API *string `json:"api"`
	// Adoption is the adoption/migration guidance reference, nil when inapplicable.
	Adoption *string `json:"adoption"`
	// Security is the maintained security guidance reference, nil when inapplicable.
	Security *string `json:"security"`
	// Compatibility is the compatibility guidance reference, nil when inapplicable.
	Compatibility *string `json:"compatibility"`
	// Performance is the workload/evidence guidance reference, nil when
	// inapplicable.
	Performance *string `json:"performance"`
	// Examples is the usage example reference, nil when inapplicable.
	Examples *string `json:"examples"`
	// FAQ is the frequently asked questions reference, nil when inapplicable.
	FAQ *string `json:"faq"`
	// Changelog is the release history reference, nil when inapplicable.
	Changelog *string `json:"changelog"`
	// PkgGoDev is the public package documentation reference, nil when inapplicable.
	PkgGoDev *string `json:"pkg_go_dev"`
	// EcosystemIndex is the catalog entry reference, nil when inapplicable.
	EcosystemIndex *string `json:"ecosystem_index"`
}

// Delivery keeps implementation, hardening, and release states independent.
type Delivery struct {
	// Implementation records the legacy implementation state, distinct from
	// hardening/release.
	Implementation string `json:"implementation"`
	// Hardening records the legacy assurance state, not automatically implied by
	// implementation.
	Hardening string `json:"hardening"`
	// Release records the legacy publication state, not automatically implied by
	// local checks.
	Release string `json:"release"`
}

// GoalEvidence binds a goal document to its implementation and verification
// claims without making those claims part of gate selection.
type GoalEvidence struct {
	// File is the legacy repository-relative requirements document.
	File string `json:"file"`
	// RequirementsSHA256 binds the recorded requirements bytes, not current gate
	// success.
	RequirementsSHA256 string `json:"requirements_sha256"`
	// ImplementationEvidence names maintained implementation claims.
	ImplementationEvidence []string `json:"implementation_evidence"`
	// VerificationGates names claimed verification boundaries without enabling
	// gates.
	VerificationGates []string `json:"verification_gates"`
	// ImplementationStatus records the legacy implementation disposition.
	ImplementationStatus string `json:"implementation_status"`
}

// Package is the canonical package classification shared by both manifests.
type Package struct {
	// ModuleDirectory is the owning canonical module root.
	ModuleDirectory string `json:"module_directory"`
	// Directory is the repository-relative package source directory.
	Directory string `json:"directory"`
	// Name is the Go package identifier, not its import identity.
	Name string `json:"name"`
	// ImportPath is unique across the canonical package inventory.
	ImportPath string `json:"import_path"`
	// Kind distinguishes public/internal/command/test classification.
	Kind string `json:"kind"`
	// Production identifies production source for selected scans and accounting.
	Production bool `json:"production"`
	// Executable identifies executable statements rather than metadata-only source.
	Executable bool `json:"executable"`
	// CoverageRequired selects expected coverage accounting for this package.
	CoverageRequired bool `json:"coverage_required"`
	// BuildRequired selects build qualification for this package.
	BuildRequired bool `json:"build_required"`
	// BuildTags is the explicit package variant selection.
	BuildTags []string `json:"build_tags"`
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
// without racing a second filesystem read. On failure a partially decoded
// catalog and admitted module bytes may still be returned for diagnostics;
// callers must check err before using them as validated execution policy.
// This single-read binding is not an atomic snapshot of both manifest files.
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

// Decode validates immutable manifest bytes with the same admission rules as
// Load. It does not read the checkout while comparing historical inputs.
// Errors may accompany a partial diagnostic catalog; no returned policy is
// qualified until validation succeeds. Input bytes remain borrowed for the call.
func Decode(policy config.Config, moduleManifest, packageManifest []byte) (Inventory, error) {
	if len(moduleManifest) == 0 || len(packageManifest) == 0 || len(moduleManifest) > maximumManifestSize || len(packageManifest) > maximumManifestSize {
		return Inventory{}, errors.New("manifest bytes are missing or exceed their size budget")
	}
	return loadManifests("", policy, moduleManifest, packageManifest)
}

func load(root string, policy config.Config, moduleManifest []byte) (Inventory, error) {
	return loadManifests(root, policy, moduleManifest, nil)
}

func loadManifests(root string, policy config.Config, moduleManifest, packageManifest []byte) (Inventory, error) {
	if moduleManifest == nil {
		var err error
		moduleManifest, err = repositoryfile.Read(root, policy.Manifests.Modules, maximumManifestSize)
		if err != nil {
			return Inventory{}, fmt.Errorf("load module manifest: %w", err)
		}
	}
	var header struct {
		SchemaVersion int `json:"schema_version"`
		Modules       []struct {
			Packages json.RawMessage `json:"packages"`
		} `json:"modules"`
	}
	if err := json.Unmarshal(moduleManifest, &header); err != nil {
		return Inventory{}, fmt.Errorf("load module manifest: %w", err)
	}
	if err := validateModuleIdentities(root, moduleManifest); err != nil {
		return Inventory{}, err
	}
	if header.SchemaVersion == 3 {
		if err := validateSchemaV3Manifest(moduleManifest); err != nil {
			return Inventory{}, err
		}
	}
	var modules Inventory
	if err := decodeData(moduleManifest, &modules); err != nil {
		return Inventory{}, fmt.Errorf("load module manifest: %w", err)
	}
	var packages packageInventory
	if packageManifest == nil {
		var err error
		packageManifest, err = repositoryfile.Read(root, policy.Manifests.Packages, maximumManifestSize)
		if err != nil {
			return modules, fmt.Errorf("load package manifest: %w", err)
		}
	}
	if err := decodeData(packageManifest, &packages); err != nil {
		return modules, fmt.Errorf("load package manifest: %w", err)
	}
	if (modules.SchemaVersion != 1 && modules.SchemaVersion != 2 && modules.SchemaVersion != 3) || packages.SchemaVersion != 1 {
		return modules, errors.New("module manifest schema_version must be 1, 2, or 3 and package manifest schema_version must be 1")
	}
	if modules.SchemaVersion < 3 && modules.SchemaID != "" {
		return modules, errors.New("module manifest schema_version 1 or 2 must not contain schema_id")
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
	if err := validateDecodedModuleIdentities(modules); err != nil {
		return Inventory{}, err
	}
	if len(modules.Modules) == 0 {
		return modules, errors.New("module manifest contains no modules")
	}
	byDirectory := make(map[string]Module, len(modules.Modules))
	derived := make(map[string]bool, len(modules.Modules))
	for index, module := range modules.Modules {
		if _, exists := byDirectory[module.Directory]; exists {
			return modules, fmt.Errorf("duplicate module directory %q", module.Directory)
		}
		byDirectory[module.Directory] = module
		if header.Modules[index].Packages == nil {
			derived[module.Directory] = true
		}
	}
	if err := validateDerivedPackageRecords(packageManifest, packages.Packages, derived); err != nil {
		return modules, err
	}
	canonicalByModule := make(map[string][]Package, len(modules.Modules))
	for _, packagePolicy := range packages.Packages {
		if _, exists := byDirectory[packagePolicy.ModuleDirectory]; !exists {
			return modules, errors.New("canonical package references unknown module")
		}
		canonicalByModule[packagePolicy.ModuleDirectory] = append(canonicalByModule[packagePolicy.ModuleDirectory], packagePolicy)
	}
	for index, module := range modules.Modules {
		// Absence opts into the canonical owner. An explicit empty array or null
		// remains an explicit mirror and must not silently acquire records.
		if header.Modules[index].Packages != nil {
			continue
		}
		module.Packages = canonicalByModule[module.Directory]
		if module.Packages == nil {
			module.Packages = []Package{}
		}
		modules.Modules[index] = module
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
	for index, acceptance := range policy.Coverage.Modules {
		module, exists := byDirectory[acceptance.Module]
		if !exists {
			return modules, fmt.Errorf("coverage.modules[%d] references unknown module %q", index, acceptance.Module)
		}
		if !module.Gates["coverage"] {
			return modules, fmt.Errorf("coverage.modules[%d] is not enabled for module %q", index, acceptance.Module)
		}
	}
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
		// A local-test declaration is its own opt-in to a focused regression;
		// it does not enable the module's legacy whole-suite test campaigns.
		enabled := operation.Gate == "local-test" || (operation.Gate == "interoperability" && len(module.InteroperabilityTools) > 0)
		if operation.Gate != "interoperability" && operation.Gate != "local-test" {
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

// validateSchemaV3Manifest owns parsing and schema admission. Public loading
// still admits the header and module identities before reaching this owner.
func validateSchemaV3Manifest(manifest []byte) error {
	document, err := jsonschema.UnmarshalJSON(bytes.NewReader(manifest))
	if err != nil {
		return fmt.Errorf("load module manifest: %w", err)
	}
	if err := compiledModulesV3Schema.Validate(document); err != nil {
		return fmt.Errorf("load module manifest: schema v3: %w", err)
	}
	return nil
}

func validateModuleIdentities(_ string, manifest []byte) error {
	var document struct {
		Repository string                       `json:"repository"`
		Modules    []map[string]json.RawMessage `json:"modules"`
	}
	if err := json.Unmarshal(manifest, &document); err != nil {
		return errors.New("invalid module manifest")
	}
	if len(document.Modules) > MaximumModules {
		return errors.New("module cardinality limit exceeded")
	}
	for _, candidate := range document.Modules {
		directory, valid := identityString(candidate, "directory")
		if !valid || !validModuleDirectory(directory) {
			return errors.New("invalid module directory")
		}
		modulePath, valid := identityString(candidate, "module_path")
		if !valid || module.CheckImportPath(modulePath) != nil || !validModulePathCharacters(modulePath) {
			return errors.New("invalid module path")
		}
		if document.Repository != "" && directory == "." && !rootModulePathMatchesRepository(modulePath, document.Repository) {
			return errors.New("invalid module path")
		}
	}
	return nil
}

func identityString(candidate map[string]json.RawMessage, field string) (string, bool) {
	raw, exists := candidate[field]
	if !exists {
		return "", false
	}
	var value string
	if err := json.Unmarshal(raw, &value); err != nil || value == "" {
		return "", false
	}
	for _, character := range value {
		if unicode.IsControl(character) {
			return "", false
		}
	}
	return value, true
}

func validModuleDirectory(directory string) bool {
	if directory == "." {
		return true
	}
	if strings.Contains(directory, "\\") || path.IsAbs(directory) || path.Clean(directory) != directory || !fs.ValidPath(directory) {
		return false
	}
	for _, character := range directory {
		if unicode.IsSpace(character) || !strings.ContainsRune("abcdefghijklmnopqrstuvwxyz0123456789-./", character) {
			return false
		}
	}
	return true
}

func validModulePathCharacters(modulePath string) bool {
	for _, character := range modulePath {
		if !strings.ContainsRune("abcdefghijklmnopqrstuvwxyz0123456789-./", character) {
			return false
		}
	}
	return true
}

func rootModulePathMatchesRepository(modulePath, repository string) bool {
	if modulePath == repository {
		return true
	}
	prefix, version, found := module.SplitPathVersion(modulePath)
	return found && prefix == repository && version != ""
}

func validateDecodedModuleIdentities(catalog Inventory) error {
	for _, candidate := range catalog.Modules {
		if strings.HasPrefix(candidate.Directory, "testdata/") || rootModulePathMatchesRepository(candidate.ModulePath, catalog.Repository) || strings.HasPrefix(candidate.ModulePath, catalog.Repository+"/") {
			continue
		}
		return errors.New("invalid module path")
	}
	return nil
}

// Omission must not turn absent/null raw classification fields into Go zero
// values. Reuse the package definition rather than maintaining a second schema;
// explicit legacy mirrors retain their existing admission behavior.
func validateDerivedPackageRecords(manifest []byte, canonical []Package, derived map[string]bool) error {
	if len(derived) == 0 {
		return nil
	}
	var document struct {
		Packages []json.RawMessage `json:"packages"`
	}
	if err := json.Unmarshal(manifest, &document); err != nil || document.Packages == nil {
		return errors.New("canonical package records must be an explicit array for derivation")
	}
	for index, packagePolicy := range canonical {
		if !derived[packagePolicy.ModuleDirectory] {
			continue
		}
		value, err := jsonschema.UnmarshalJSON(bytes.NewReader(document.Packages[index]))
		if err != nil || compiledPackageSchema.Validate(value) != nil {
			return fmt.Errorf("canonical package record %d violates required classification", index)
		}
	}
	return nil
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
