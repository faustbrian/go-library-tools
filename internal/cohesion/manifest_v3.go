package cohesion

import (
	"errors"
	"fmt"
	"strings"
)

const maximumManifestV3Bytes = 1 << 20

func validateManifestV3(data []byte, requirementsSHA256 string) error {
	if !v3SHA256Pattern.MatchString(requirementsSHA256) {
		return errors.New("manifest-goal: frozen requirements digest is invalid")
	}
	budget := newTrustJSONBudget(maximumManifestV3Bytes)
	document, err := parseTrustJSONWithBudgetAndLimits(data, maximumManifestV3Bytes, budget, manifestV3TrustJSONLimits())
	if err != nil {
		return fmt.Errorf("json-manifest: %w", err)
	}
	if err := compiledSchemaV3Schemas[modulesV3SchemaIdentity].Validate(document); err != nil {
		return fmt.Errorf("schema-manifest: %w", err)
	}
	manifest := document.(map[string]any)
	if !canonicalIdentity(manifest["repository"].(string)) {
		return errors.New("manifest-repository: repository identity is not canonical")
	}
	modules := manifest["modules"].([]any)
	directories := make(map[string]struct{}, len(modules))
	selectedEvidence := make(map[string]string)
	for index, rawModule := range modules {
		module := rawModule.(map[string]any)
		directory := module["directory"].(string)
		if directory != "." && (!safeRelativePath(directory) || len(directory) > 4096) {
			return fmt.Errorf("manifest-module: module %d has an unsafe directory", index)
		}
		if _, exists := directories[directory]; exists {
			return fmt.Errorf("manifest-module: module %d repeats directory %q", index, directory)
		}
		directories[directory] = struct{}{}
		if err := validateManifestV3Provenance(index, module["provenance"].([]any)); err != nil {
			return err
		}
		cohesionValue, hasCohesion := module["cohesion"]
		if !hasCohesion {
			continue
		}
		cohesion := cohesionValue.(map[string]any)
		if err := validateManifestV3IntegrationRoles(index, cohesion["integration_roles"].([]any)); err != nil {
			return err
		}
		var delivery ManifestDelivery
		if err := assignTrustJSON(&delivery, cohesion["delivery"], budget); err != nil {
			return fmt.Errorf("json-manifest: module %d delivery: %w", index, err)
		}
		if err := validateManifestDeliverySelection(delivery, requirementsSHA256, selectedEvidence); err != nil {
			return fmt.Errorf("manifest-delivery: module %d: %w", index, err)
		}
	}
	return nil
}

func validateManifestV3Provenance(moduleIndex int, values []any) error {
	previous := ""
	for index, raw := range values {
		path := raw.(string)
		if len(path) > 4096 || !safeRelativePath(path) {
			return fmt.Errorf("manifest-provenance: module %d provenance %d is unsafe", moduleIndex, index)
		}
		if index != 0 && strings.Compare(previous, path) >= 0 {
			return fmt.Errorf("manifest-provenance: module %d provenance is not sorted and unique", moduleIndex)
		}
		previous = path
	}
	return nil
}

func validateManifestV3IntegrationRoles(moduleIndex int, values []any) error {
	previousTarget, previousRole := "", ""
	targets := make(map[string]struct{}, len(values))
	for index, raw := range values {
		role := raw.(map[string]any)
		target := role["target"].(string)
		name := role["role"].(string)
		if !canonicalIdentity(target) || !canonicalIdentity(role["rationale"].(string)) || !canonicalIdentity(role["authorization_id"].(string)) {
			return fmt.Errorf("manifest-integration-role: module %d role %d has a noncanonical identity", moduleIndex, index)
		}
		if _, exists := targets[target]; exists {
			return fmt.Errorf("manifest-integration-role: module %d repeats target %q", moduleIndex, target)
		}
		if index != 0 && (strings.Compare(previousTarget, target) > 0 || previousTarget == target && strings.Compare(previousRole, name) >= 0) {
			return fmt.Errorf("manifest-integration-role: module %d roles are not sorted and unique", moduleIndex)
		}
		targets[target] = struct{}{}
		previousTarget, previousRole = target, name
	}
	return nil
}

func manifestV3TrustJSONLimits() trustJSONParserLimits {
	return trustJSONParserLimits{maximumStringBytes: func(path []trustJSONPathStep) int {
		field := nearestManifestV3Field(path)
		switch field {
		case "schema_id":
			return len("urn:golib:cohesion:module-manifest:v3")
		case "requirements_sha256", "receipt_sha256":
			return len("sha256:") + 64
		case "id":
			return len(cohesionGoalID)
		case "directory", "module_directory", "module_path", "import_path", "receipt_path", "file", "package_selection",
			"primary_entry_packages", "optional_owned_dependencies", "owned_dependencies", "reverse_owned_dependencies", "goal_files", "implementation_evidence",
			"readme", "api", "adoption", "security", "compatibility", "performance", "examples", "faq", "changelog",
			"pkg_go_dev", "ecosystem_index":
			return 4096
		case "provenance":
			if manifestV3PathContainsMember(path, "integration_roles") {
				return len("authorized")
			}
			return 4096
		case "repository", "version", "tag_prefix", "goal_status", "public_package_identifier",
			"supported_platforms", "target", "rationale", "authorization_id", "entry_id":
			return 128
		case "kind":
			if manifestV3PathContainsMember(path, "packages") {
				return maximumTrustJSONString
			}
			return 128
		case "family":
			if !manifestV3PathContainsMember(path, "cohesion") {
				return 128
			}
			return len("integration-and-data-movement")
		case "lifecycle_status":
			return len("deprecated")
		case "maturity":
			return len("experimental")
		case "role":
			return len("domain-owned")
		case "secondary_capabilities":
			return len("scheduling-and-orchestration")
		case "construction_styles":
			return len("functional-options")
		case "lifecycle_styles":
			return len("stateless")
		case "configuration", "runtime_resources", "background_work":
			return len("package")
		case "mutable_inputs":
			return len("transfer")
		case "status":
			return len("not-applicable")
		case "implementation", "hardening", "release":
			return len("not-applicable")
		default:
			return maximumTrustJSONString
		}
	}}
}

func manifestV3PathContainsMember(path []trustJSONPathStep, name string) bool {
	for _, step := range path {
		if step.kind == trustJSONObjectMember && step.name == name {
			return true
		}
	}
	return false
}

func nearestManifestV3Field(path []trustJSONPathStep) string {
	for index := len(path) - 1; index >= 0; index-- {
		if path[index].kind == trustJSONObjectMember {
			return path[index].name
		}
	}
	return ""
}
