package cohesion

import (
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/faustbrian/go-library-tools/internal/inventory"
	"github.com/faustbrian/go-library-tools/internal/repositoryfile"
)

var (
	goIdentifier = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)
	goVersion    = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$`)
	kebab        = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	sha256Hex    = regexp.MustCompile(`^[0-9a-f]{64}$`)
)

var families = stringSet(
	"foundations", "service-edge", "protocols-and-descriptions",
	"persistence-and-durability", "resilience", "observability",
	"integration-and-data-movement", "domain-utilities", "tooling",
)

var capabilities = stringSet(
	"administration-and-control", "configuration", "cryptography-and-secrets",
	"data-encoding", "database-and-storage", "distributed-coordination",
	"eventing-and-messaging", "http-and-service-edge", "identity-and-access",
	"observability", "protocols-and-schemas", "resilience-and-admission",
	"scheduling-and-orchestration", "testing-and-conformance", "transport-and-networking",
)

var constructionStyles = stringSet(
	"plain-function", "validated-value", "new-config", "new-options",
	"functional-options", "builder-compile", "builder-build", "open",
	"connect", "load", "init", "must-helper",
)

var lifecycleStyles = stringSet("stateless", "run", "start", "drain", "shutdown", "close", "wait")
var lifecycleStatuses = stringSet("planned", "active", "deprecated", "retired")
var maturities = stringSet("experimental", "preview", "stable")
var configurationOwners = stringSet("caller", "none")
var resourceOwners = stringSet("caller", "package", "shared", "none")
var mutableInputStyles = stringSet("copy", "borrow", "transfer", "none")
var deliveryStatuses = stringSet("not-started", "in-progress", "verified", "blocked", "not-applicable")

func validateModule(root string, index int, module inventory.Module) []Diagnostic {
	metadata := module.Cohesion
	base := fmt.Sprintf("/modules/%d/cohesion", index)
	diagnostics := make([]Diagnostic, 0)
	add := func(code, suffix, message string) {
		diagnostics = append(diagnostics, Diagnostic{Code: code, Path: base + suffix, Message: message})
	}

	if !families[metadata.Family] || module.Family != metadata.Family {
		add("invalid-value", "/family", "cohesion family must be a frozen family and match the legacy family")
	}
	validateStringSet(metadata.SecondaryCapabilities, capabilities, false, base+"/secondary_capabilities", &diagnostics)
	if strings.TrimSpace(metadata.Responsibility) == "" {
		add("missing-metadata", "/responsibility", "responsibility must be non-empty reviewed prose")
	}
	if len(metadata.NonGoals) == 0 || containsBlank(metadata.NonGoals) {
		add("missing-metadata", "/non_goals", "non-goals must contain at least one non-empty boundary")
	}

	packages := make(map[string]inventory.Package, len(module.Packages))
	for _, packagePolicy := range module.Packages {
		packages[packagePolicy.ImportPath] = packagePolicy
	}
	if metadata.PublicPackageIdentifier != "none" && !goIdentifier.MatchString(metadata.PublicPackageIdentifier) {
		add("invalid-value", "/public_package_identifier", "public package identifier must be a Go identifier or none")
	}
	validateSortedUnique(metadata.PrimaryEntryPackages, true, base+"/primary_entry_packages", &diagnostics)
	for _, importPath := range metadata.PrimaryEntryPackages {
		if _, exists := packages[importPath]; !exists {
			add("invalid-reference", "/primary_entry_packages", "primary entry package must exist in the module package inventory")
		}
	}
	if metadata.PublicPackageIdentifier == "none" {
		if hasRootPublicPackage(module) {
			add("invalid-reference", "/public_package_identifier", "none is reserved for a module without a root public package")
		}
		if len(metadata.PrimaryEntryPackages) == 0 || !selectionCovers(metadata.PackageSelection, metadata.PrimaryEntryPackages) {
			add("invalid-value", "/package_selection", "a rootless suite requires selection prose for every primary entry package")
		}
	} else if !identifierMatches(metadata.PublicPackageIdentifier, module.ModulePath, metadata.PrimaryEntryPackages, packages) {
		add("invalid-reference", "/public_package_identifier", "public package identifier must match a primary entry package")
	}
	for packagePath, prose := range metadata.PackageSelection {
		if _, exists := packages[packagePath]; !exists || strings.TrimSpace(prose) == "" {
			add("invalid-reference", "/package_selection", "package selection must reference an inventoried package with non-empty prose")
			break
		}
	}

	if !lifecycleStatuses[metadata.LifecycleStatus] {
		add("invalid-value", "/lifecycle_status", "lifecycle status is not recognized")
	}
	if !maturities[metadata.Maturity] || (metadata.LifecycleStatus == "planned" && metadata.Maturity == "stable") {
		add("invalid-value", "/maturity", "maturity is not valid for the lifecycle status")
	}
	validateStringSet(metadata.ConstructionStyles, constructionStyles, true, base+"/construction_styles", &diagnostics)
	validateStringSet(metadata.LifecycleStyles, lifecycleStyles, true, base+"/lifecycle_styles", &diagnostics)

	if !configurationOwners[metadata.Ownership.Configuration] {
		add("invalid-value", "/ownership/configuration", "configuration ownership is not recognized")
	}
	validateStringSet(metadata.Ownership.MutableInputs, mutableInputStyles, true, base+"/ownership/mutable_inputs", &diagnostics)
	if contains(metadata.Ownership.MutableInputs, "none") && len(metadata.Ownership.MutableInputs) != 1 {
		add("invalid-value", "/ownership/mutable_inputs", "none must be the only mutable-input style")
	}
	if !resourceOwners[metadata.Ownership.RuntimeResources] {
		add("invalid-value", "/ownership/runtime_resources", "runtime resource ownership is not recognized")
	}
	if !resourceOwners[metadata.Ownership.BackgroundWork] {
		add("invalid-value", "/ownership/background_work", "background work ownership is not recognized")
	}

	validateSortedUnique(metadata.OptionalOwnedDependencies, false, base+"/optional_owned_dependencies", &diagnostics)
	validateSortedUnique(metadata.Adapters, false, base+"/adapters", &diagnostics)
	validateSortedUnique(metadata.Companions, false, base+"/companions", &diagnostics)
	validateDependencySets(module.OwnedDependencies, metadata, base, &diagnostics)

	if !goVersion.MatchString(metadata.SupportedGo.Minimum) || metadata.SupportedGo.Minimum != module.GoVersion {
		add("invalid-value", "/supported_go/minimum", "minimum Go version must match the module Go version")
	}
	validateSortedUnique(metadata.SupportedGo.Tested, true, base+"/supported_go/tested", &diagnostics)
	for _, tested := range metadata.SupportedGo.Tested {
		if !goVersion.MatchString(tested) || compareGoVersions(tested, metadata.SupportedGo.Minimum) < 0 {
			add("invalid-value", "/supported_go/tested", "tested Go versions must be exact and not below the minimum")
			break
		}
	}
	validateIdentifiers(metadata.SupportedPlatforms, true, true, base+"/supported_platforms", &diagnostics)
	validateIdentifiers(metadata.SupportedBackends, false, false, base+"/supported_backends", &diagnostics)
	validateIdentifiers(metadata.SupportedProtocols, false, false, base+"/supported_protocols", &diagnostics)

	validateDocumentation(root, metadata, base, &diagnostics)
	validateSortedUnique(metadata.KnownGoodCompatibilitySets, false, base+"/known_good_compatibility_sets", &diagnostics)
	validateDelivery(module, metadata, base, &diagnostics)
	return diagnostics
}

func validateStringSet(values []string, allowed map[string]bool, required bool, path string, diagnostics *[]Diagnostic) {
	validateSortedUnique(values, required, path, diagnostics)
	for _, value := range values {
		if !allowed[value] {
			*diagnostics = append(*diagnostics, Diagnostic{Code: "invalid-value", Path: path, Message: "set contains an unsupported value"})
			return
		}
	}
}

func validateSortedUnique(values []string, required bool, path string, diagnostics *[]Diagnostic) {
	if required && len(values) == 0 {
		*diagnostics = append(*diagnostics, Diagnostic{Code: "missing-metadata", Path: path, Message: "set must not be empty"})
		return
	}
	for index, value := range values {
		if strings.TrimSpace(value) == "" {
			*diagnostics = append(*diagnostics, Diagnostic{Code: "invalid-value", Path: path, Message: "set entries must be non-empty"})
			return
		}
		if index > 0 {
			if values[index-1] == value {
				*diagnostics = append(*diagnostics, Diagnostic{Code: "duplicate-entry", Path: path, Message: "set entries must be unique"})
				return
			}
			if strings.Compare(values[index-1], value) == 1 {
				*diagnostics = append(*diagnostics, Diagnostic{Code: "nondeterministic-order", Path: path, Message: "set entries must be sorted"})
				return
			}
		}
	}
}

func validateIdentifiers(values []string, required, platform bool, path string, diagnostics *[]Diagnostic) {
	validateSortedUnique(values, required, path, diagnostics)
	for _, value := range values {
		valid := kebab.MatchString(value)
		if platform {
			valid = value == "portable-go" || (valid && strings.Count(value, "-") >= 1)
		}
		if !valid {
			*diagnostics = append(*diagnostics, Diagnostic{Code: "invalid-value", Path: path, Message: "environment identifier is invalid"})
			return
		}
	}
}

func validateDependencySets(required []string, metadata *inventory.Cohesion, base string, diagnostics *[]Diagnostic) {
	seen := make(map[string]string)
	for _, set := range []struct {
		name   string
		values []string
	}{{"owned_dependencies", required}, {"optional_owned_dependencies", metadata.OptionalOwnedDependencies}, {"adapters", metadata.Adapters}, {"companions", metadata.Companions}} {
		for _, value := range set.values {
			if previous, exists := seen[value]; exists {
				*diagnostics = append(*diagnostics, Diagnostic{Code: "duplicate-entry", Path: base + "/" + set.name, Message: "dependency identity duplicates the " + previous + " set"})
				return
			}
			seen[value] = set.name
		}
	}
}

func validateDocumentation(root string, metadata *inventory.Cohesion, base string, diagnostics *[]Diagnostic) {
	values := []struct {
		name  string
		value *string
	}{{"readme", metadata.Documentation.README}, {"api", metadata.Documentation.API}, {"adoption", metadata.Documentation.Adoption}, {"security", metadata.Documentation.Security}, {"compatibility", metadata.Documentation.Compatibility}, {"performance", metadata.Documentation.Performance}, {"examples", metadata.Documentation.Examples}, {"faq", metadata.Documentation.FAQ}, {"changelog", metadata.Documentation.Changelog}, {"pkg_go_dev", metadata.Documentation.PkgGoDev}, {"ecosystem_index", metadata.Documentation.EcosystemIndex}}
	required := metadata.LifecycleStatus == "active" || metadata.LifecycleStatus == "deprecated"
	for _, entry := range values {
		path := base + "/documentation/" + entry.name
		if entry.value == nil {
			activeEntryPoint := required && contains([]string{"readme", "api", "changelog", "pkg_go_dev", "ecosystem_index"}, entry.name)
			plannedEntryPoint := metadata.LifecycleStatus == "planned" && entry.name == "readme"
			if activeEntryPoint || plannedEntryPoint {
				*diagnostics = append(*diagnostics, Diagnostic{Code: "missing-metadata", Path: path, Message: "lifecycle documentation entry point is required"})
			}
			continue
		}
		if isHTTPS(*entry.value) {
			continue
		}
		if !safeRelativePath(*entry.value) {
			*diagnostics = append(*diagnostics, Diagnostic{Code: "unsafe-path", Path: path, Message: "documentation path must remain inside the repository"})
			continue
		}
		if err := repositoryfile.ValidateRegularFile(root, filepath.FromSlash(*entry.value)); err != nil {
			*diagnostics = append(*diagnostics, Diagnostic{Code: "missing-document", Path: path, Message: "documentation path must be an existing regular non-symlink file"})
		}
	}
}

func validateDelivery(module inventory.Module, metadata *inventory.Cohesion, base string, diagnostics *[]Diagnostic) {
	for _, state := range []struct {
		name  string
		value string
	}{{"implementation", metadata.Delivery.Implementation}, {"hardening", metadata.Delivery.Hardening}, {"release", metadata.Delivery.Release}} {
		if !deliveryStatuses[state.value] {
			*diagnostics = append(*diagnostics, Diagnostic{Code: "invalid-value", Path: base + "/delivery/" + state.name, Message: "delivery status is not recognized"})
		}
		if state.value == "verified" && !hasAttributableGoalEvidence(module) {
			*diagnostics = append(*diagnostics, Diagnostic{Code: "invalid-reference", Path: base + "/delivery/" + state.name, Message: "verified delivery status requires complete attributable goal evidence"})
		}
	}
}

func hasRootPublicPackage(module inventory.Module) bool {
	for _, packagePolicy := range module.Packages {
		if packagePolicy.ImportPath == module.ModulePath && packagePolicy.Kind == "public" {
			return true
		}
	}
	return false
}

func hasAttributableGoalEvidence(module inventory.Module) bool {
	if module.GoalStatus != "complete" {
		return false
	}
	for _, evidence := range module.GoalEvidence {
		if contains(module.GoalFiles, evidence.File) && sha256Hex.MatchString(evidence.RequirementsSHA256) &&
			len(evidence.ImplementationEvidence) > 0 && !containsBlank(evidence.ImplementationEvidence) &&
			len(evidence.VerificationGates) > 0 && !containsBlank(evidence.VerificationGates) &&
			evidence.ImplementationStatus == "verified" {
			return true
		}
	}
	return false
}

func safeRelativePath(value string) bool {
	if filepath.IsAbs(value) || strings.Contains(value, "\\") {
		return false
	}
	clean := filepath.Clean(filepath.FromSlash(value))
	return clean != "." && clean != ".." && !strings.HasPrefix(clean, ".."+string(filepath.Separator)) && filepath.ToSlash(clean) == value
}

func isHTTPS(value string) bool {
	parsed, err := url.Parse(value)
	if err != nil {
		return false
	}
	return parsed != nil && parsed.Scheme == "https" && parsed.Host != "" && parsed.User == nil
}

func compareGoVersions(left, right string) int {
	leftParts := strings.Split(left, ".")
	rightParts := strings.Split(right, ".")
	if len(leftParts) != 3 || len(rightParts) != 3 {
		return -1
	}
	for index := range 3 {
		leftValue, leftError := strconv.Atoi(leftParts[index])
		rightValue, rightError := strconv.Atoi(rightParts[index])
		if leftError != nil || rightError != nil {
			return -1
		}
		if leftValue < rightValue {
			return -1
		}
		if leftValue > rightValue {
			return 1
		}
	}
	return 0
}

func identifierMatches(identifier, modulePath string, entries []string, packages map[string]inventory.Package) bool {
	return contains(entries, modulePath) && packages[modulePath].Name == identifier
}

func selectionCovers(selection map[string]string, entries []string) bool {
	for _, entry := range entries {
		if strings.TrimSpace(selection[entry]) == "" {
			return false
		}
	}
	return true
}

func contains(values []string, target string) bool {
	return slices.Contains(values, target)
}

func containsBlank(values []string) bool {
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			return true
		}
	}
	return false
}

func stringSet(values ...string) map[string]bool {
	set := make(map[string]bool, len(values))
	for _, value := range values {
		set[value] = true
	}
	return set
}
