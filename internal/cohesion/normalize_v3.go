package cohesion

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

const maximumSchemaV3ArtifactBytes = 32 << 20

const (
	maximumDefaultSchemaV3ArtifactBytes        = 1 << 20
	maximumOfficialOracleSchemaV3ArtifactBytes = 128 << 20
)

const authorizationRecordV1SchemaIdentity = "https://github.com/faustbrian/go-library-tools/schema/cohesion-authorization-record-v1.schema.json"
const deliveryEvidenceV1SchemaIdentity = "https://github.com/faustbrian/go-library-tools/schema/cohesion-delivery-evidence-v1.schema.json"
const schemaV3GoalRequirementsSHA256 = "sha256:6e5ac948c2cf1ec439fdf90c53c0648befb74a65c3eca2eb595737ee8211891c"

type schemaV3NormalizedValue struct {
	Value            any
	Normalized       []byte
	NormalizedSHA256 string
}

type schemaV3SemanticValidator func(any) []schemaV3Failure

const (
	goOfficialDownloadsOracleV1SchemaIdentity       = "https://github.com/faustbrian/go-library-tools/schema/cohesion-go-official-downloads-oracle-v1.schema.json"
	goOfficialDownloadsOracleReviewV1SchemaIdentity = "https://github.com/faustbrian/go-library-tools/schema/cohesion-go-official-downloads-oracle-review-v1.schema.json"
	goToolchainsV1SchemaIdentity                    = "https://github.com/faustbrian/go-library-tools/schema/cohesion-go-toolchains-v1.schema.json"
	engineeringIdentitiesV1SchemaIdentity           = "https://github.com/faustbrian/go-library-tools/schema/cohesion-engineering-identities-v1.schema.json"
	authorizationRegistryV1SchemaIdentity           = "https://github.com/faustbrian/go-library-tools/schema/cohesion-authorization-registry-v1.schema.json"
	gatePolicyV1SchemaIdentity                      = "https://github.com/faustbrian/go-library-tools/schema/cohesion-gate-policy-v1.schema.json"
	residualInventoryV1SchemaIdentity               = "https://github.com/faustbrian/go-library-tools/schema/cohesion-residual-inventory-v1.schema.json"
	residualRegisterV1SchemaIdentity                = "https://github.com/faustbrian/go-library-tools/schema/cohesion-residual-register-v1.schema.json"
	schemaProvenanceV1SchemaIdentity                = "https://github.com/faustbrian/go-library-tools/schema/cohesion-schema-provenance-v1.schema.json"
	sourcesV2SchemaIdentity                         = "https://github.com/faustbrian/go-library-tools/schema/cohesion-sources-v2.schema.json"
	inputsV2SchemaIdentity                          = "https://github.com/faustbrian/go-library-tools/schema/cohesion-inputs-v2.schema.json"
	catalogV2SchemaIdentity                         = "https://github.com/faustbrian/go-library-tools/schema/cohesion-catalog-v2.schema.json"
)

var schemaV3SemanticValidators = map[string]schemaV3SemanticValidator{
	goOfficialDownloadsOracleV1SchemaIdentity:       validateOfficialDownloadsOracleV1Semantics,
	goOfficialDownloadsOracleReviewV1SchemaIdentity: validateOfficialDownloadsOracleReviewV1Semantics,
	goToolchainsV1SchemaIdentity:                    validateGoToolchainsV1Semantics,
	engineeringIdentitiesV1SchemaIdentity:           validateEngineeringIdentitiesV1Semantics,
	authorizationRegistryV1SchemaIdentity:           validateAuthorizationRegistryV1Semantics,
	gatePolicyV1SchemaIdentity:                      validateGatePolicyV1Semantics,
	residualInventoryV1SchemaIdentity:               validateResidualInventoryV1Semantics,
	residualRegisterV1SchemaIdentity:                validateResidualRegisterV1Semantics,
	schemaProvenanceV1SchemaIdentity:                validateSchemaProvenanceV1Semantics,
	sourcesV2SchemaIdentity:                         validateSourcesV2Semantics,
	inputsV2SchemaIdentity:                          validateInputsV2Semantics,
	catalogV2SchemaIdentity:                         validateCatalogV2Semantics,
	modulesV3SchemaIdentity:                         validateModulesV3Semantics,
	schemaV3DecisionReviewV1SchemaIdentity:          validateSchemaV3DecisionControlSemantics,
	schemaV3DecisionFreezeV1SchemaIdentity:          validateSchemaV3DecisionControlSemantics,
	contractReviewV1SchemaIdentity:                  validateContractReviewV1Semantics,
	contractFreezeV1SchemaIdentity:                  validateContractFreezeV1Semantics,
	deliveryEvidenceV1SchemaIdentity:                validateDeliveryEvidenceV1Semantics,
	authorizationRecordV1SchemaIdentity:             validateAuthorizationRecordV1Semantics,
}

func validateDeliveryEvidenceV1Semantics(value any) []schemaV3Failure {
	root := value.(map[string]any)
	entries := root["entries"].([]any)
	failures := make([]schemaV3Failure, 0)
	count, _ := oracleInteger(root["entry_count"])
	if count != len(entries) {
		failures = append(failures, schemaV3Failure{Code: "semantic-count-mismatch", Pointer: "/entry_count"})
	}
	seen := make(map[string]struct{}, len(entries))
	for index, raw := range entries {
		entry := raw.(map[string]any)
		pointer := fmt.Sprintf("/entries/%d", index)
		entryID := entry["entry_id"].(string)
		if _, duplicate := seen[entryID]; duplicate {
			failures = append(failures, schemaV3Failure{Code: "semantic-duplicate", Pointer: pointer + "/entry_id"})
		} else {
			seen[entryID] = struct{}{}
		}
		if entry["requirements_sha256"] != schemaV3GoalRequirementsSHA256 {
			failures = append(failures, schemaV3Failure{Code: "semantic-external-identity", Pointer: pointer + "/requirements_sha256"})
		}
		manifestBytes, err := canonicalMarshal(entry["input_manifest"], maximumDefaultSchemaV3ArtifactBytes)
		if err != nil {
			failures = append(failures, schemaV3Failure{Code: "semantic-unsupported", Pointer: pointer + "/input_manifest", Detail: err.Error()})
		} else if entry["input_manifest_sha256"] != exactBytesSHA256(manifestBytes) {
			failures = append(failures, schemaV3Failure{Code: "semantic-digest-mismatch", Pointer: pointer + "/input_manifest_sha256"})
		}
		if policy, exists := entry["policy"].(map[string]any); exists {
			digest, err := semanticObjectSHA256(policy, "policy_sha256", maximumDefaultSchemaV3ArtifactBytes)
			if err != nil {
				failures = append(failures, schemaV3Failure{Code: "semantic-unsupported", Pointer: pointer + "/policy/policy_sha256", Detail: err.Error()})
			} else if policy["policy_sha256"] != digest {
				failures = append(failures, schemaV3Failure{Code: "semantic-digest-mismatch", Pointer: pointer + "/policy/policy_sha256"})
			}
		}
	}
	return failures
}

func validateAuthorizationRecordV1Semantics(value any) []schemaV3Failure {
	root := value.(map[string]any)
	failures := make([]schemaV3Failure, 0)
	if !isValidWholeSecondUTC(root["created_at"].(string)) {
		failures = append(failures, schemaV3Failure{Code: "semantic-cross-field", Pointer: "/created_at"})
	}
	subject := root["subject"].(map[string]any)
	if subject["kind"] == "delivery" {
		path := subject["receipt_path"].(string)
		if len(path) > 4096 || !safeRelativePath(path) {
			failures = append(failures, schemaV3Failure{Code: "semantic-path", Pointer: "/subject/receipt_path"})
		}
	}
	return failures
}

func validateModulesV3Semantics(value any) []schemaV3Failure {
	encoded, err := canonicalMarshal(value, maximumManifestV3Bytes)
	if err != nil {
		return []schemaV3Failure{{Code: "semantic-unsupported", Detail: err.Error()}}
	}
	const requirementsSHA256 = "sha256:6e5ac948c2cf1ec439fdf90c53c0648befb74a65c3eca2eb595737ee8211891c"
	if err := validateManifestV3(encoded, requirementsSHA256); err != nil {
		return []schemaV3Failure{{Code: "semantic-cross-field", Detail: err.Error()}}
	}
	return nil
}

func validateCatalogV2Semantics(value any) []schemaV3Failure {
	root := value.(map[string]any)
	if root["scope"] == "repository" && root["publication_status"] == "final-input" {
		version, _ := oracleInteger(root["manifest_schema_version"])
		if version != 3 {
			return []schemaV3Failure{{Code: "semantic-cross-field", Pointer: "/publication_status"}}
		}
	}
	return nil
}

func validateInputsV2Semantics(value any) []schemaV3Failure {
	root := value.(map[string]any)
	projections := root["projections"].([]any)
	failures := make([]schemaV3Failure, 0)
	count, _ := oracleInteger(root["projection_count"])
	if count != len(projections) {
		failures = append(failures, schemaV3Failure{Code: "semantic-count-mismatch", Pointer: "/projection_count"})
	}
	seen := make(map[string]struct{}, len(projections))
	previous := ""
	for index, raw := range projections {
		repository := raw.(map[string]any)["repository"].(string)
		pointer := fmt.Sprintf("/projections/%d/repository", index)
		if _, duplicate := seen[repository]; duplicate {
			failures = append(failures, schemaV3Failure{Code: "semantic-duplicate", Pointer: pointer})
		} else {
			seen[repository] = struct{}{}
		}
		if previous != "" && repository < previous {
			failures = append(failures, schemaV3Failure{Code: "semantic-canonical-order", Pointer: pointer})
		}
		previous = repository
	}
	return failures
}

func validateSourcesV2Semantics(value any) []schemaV3Failure {
	root := value.(map[string]any)
	sources := root["sources"].([]any)
	failures := make([]schemaV3Failure, 0)
	count, _ := oracleInteger(root["repository_count"])
	if count != len(sources) {
		failures = append(failures, schemaV3Failure{Code: "semantic-count-mismatch", Pointer: "/repository_count"})
	}
	seen := make(map[string]struct{}, len(sources))
	previous := ""
	for index, raw := range sources {
		repository := raw.(map[string]any)["repository"].(string)
		pointer := fmt.Sprintf("/sources/%d/repository", index)
		if _, duplicate := seen[repository]; duplicate {
			failures = append(failures, schemaV3Failure{Code: "semantic-duplicate", Pointer: pointer})
		} else {
			seen[repository] = struct{}{}
		}
		if previous != "" && repository < previous {
			failures = append(failures, schemaV3Failure{Code: "semantic-canonical-order", Pointer: pointer})
		}
		previous = repository
	}
	controls := root["controls"].(map[string]any)
	for _, name := range []string{"authorization_records", "registry_predecessors", "effective_policy_inputs"} {
		wrapper := controls[name].(map[string]any)
		declared, _ := oracleInteger(wrapper["count"])
		if declared != len(wrapper["entries"].([]any)) {
			failures = append(failures, schemaV3Failure{Code: "semantic-count-mismatch", Pointer: "/controls/" + name + "/count"})
		}
	}
	return failures
}

func validateGatePolicyV1Semantics(value any) []schemaV3Failure {
	root := value.(map[string]any)
	digest, err := semanticObjectSHA256(root, "policy_sha256", maximumSchemaV3ArtifactBytes)
	if err != nil {
		return []schemaV3Failure{{Code: "semantic-unsupported", Pointer: "/policy_sha256", Detail: err.Error()}}
	}
	if root["policy_sha256"] != digest {
		return []schemaV3Failure{{Code: "semantic-digest-mismatch", Pointer: "/policy_sha256"}}
	}
	return nil
}

func validateResidualInventoryV1Semantics(value any) []schemaV3Failure {
	root := value.(map[string]any)
	sources := root["sources"].([]any)
	items := root["items"].([]any)
	failures := make([]schemaV3Failure, 0)
	itemCount, _ := oracleInteger(root["item_count"])
	if itemCount != len(items) {
		failures = append(failures, schemaV3Failure{Code: "semantic-count-mismatch", Pointer: "/item_count"})
	}
	release := ""
	if len(sources) > 0 {
		release = sources[0].(map[string]any)["release"].(string)
	}
	for index, raw := range sources {
		if raw.(map[string]any)["release"] != release {
			return append(failures, schemaV3Failure{Code: "semantic-release-identity", Pointer: fmt.Sprintf("/sources/%d/release", index)})
		}
	}
	for index, raw := range sources {
		source := raw.(map[string]any)
		pointer := fmt.Sprintf("/sources/%d", index)
		sourceRelease := source["release"].(string)
		wantURL := "https://github.com/faustbrian/go-library-tools/releases/download/" + sourceRelease + "/" + source["asset"].(string)
		if source["release_url"] != wantURL {
			failures = append(failures, schemaV3Failure{Code: "semantic-cross-field", Pointer: pointer + "/release_url"})
		}
	}
	seen := make(map[string]struct{}, len(items))
	previous := ""
	for index, raw := range items {
		item := raw.(map[string]any)
		pointer := fmt.Sprintf("/items/%d", index)
		itemID := item["inventory_item_id"].(string)
		if _, duplicate := seen[itemID]; duplicate {
			failures = append(failures, schemaV3Failure{Code: "semantic-duplicate", Pointer: pointer + "/inventory_item_id"})
		} else {
			seen[itemID] = struct{}{}
		}
		tuple := strings.Join([]string{item["repository"].(string), item["module"].(string), itemID}, "\x00")
		if previous != "" && tuple < previous {
			failures = append(failures, schemaV3Failure{Code: "semantic-canonical-order", Pointer: pointer})
		}
		previous = tuple
	}
	return failures
}

func validateResidualRegisterV1Semantics(value any) []schemaV3Failure {
	root := value.(map[string]any)
	failures := make([]schemaV3Failure, 0)
	for _, collection := range []struct {
		count string
		rows  string
	}{{"release_entry_count", "release_entries"}, {"exception_entry_count", "exceptions"}, {"resolution_count", "resolutions"}} {
		count, _ := oracleInteger(root[collection.count])
		if count != len(root[collection.rows].([]any)) {
			failures = append(failures, schemaV3Failure{Code: "semantic-count-mismatch", Pointer: "/" + collection.count})
		}
	}
	exceptionItems := make(map[string]struct{}, len(root["exceptions"].([]any)))
	previousException := ""
	for index, raw := range root["exceptions"].([]any) {
		exception := raw.(map[string]any)
		itemID := exception["inventory_item_id"].(string)
		exceptionItems[itemID] = struct{}{}
		tuple := strings.Join([]string{exception["repository"].(string), exception["module"].(string), exception["exception_id"].(string)}, "\x00")
		if previousException != "" && tuple < previousException {
			failures = append(failures, schemaV3Failure{Code: "semantic-canonical-order", Pointer: fmt.Sprintf("/exceptions/%d", index)})
		}
		previousException = tuple
	}
	seenResolutions := make(map[string]struct{}, len(root["resolutions"].([]any)))
	previousResolution := ""
	for index, raw := range root["resolutions"].([]any) {
		resolution := raw.(map[string]any)
		pointer := fmt.Sprintf("/resolutions/%d", index)
		itemID := resolution["inventory_item_id"].(string)
		if _, overlap := exceptionItems[itemID]; overlap {
			failures = append(failures, schemaV3Failure{Code: "semantic-duplicate", Pointer: pointer + "/inventory_item_id"})
		}
		if _, duplicate := seenResolutions[itemID]; duplicate {
			failures = append(failures, schemaV3Failure{Code: "semantic-duplicate", Pointer: pointer + "/inventory_item_id"})
		} else {
			seenResolutions[itemID] = struct{}{}
		}
		if previousResolution != "" && itemID < previousResolution {
			failures = append(failures, schemaV3Failure{Code: "semantic-canonical-order", Pointer: pointer})
		}
		previousResolution = itemID
	}
	return failures
}

func validateSchemaProvenanceV1Semantics(value any) []schemaV3Failure {
	root := value.(map[string]any)
	rootFreeze := root["decision_freeze"].(map[string]any)
	rootFreezeBytes, err := canonicalMarshal(rootFreeze, maximumSchemaV3ArtifactBytes)
	if err != nil {
		return []schemaV3Failure{{Code: "semantic-unsupported", Pointer: "/decision_freeze", Detail: err.Error()}}
	}
	release := rootFreeze["release"].(string)
	entries := root["entries"].([]any)
	failures := make([]schemaV3Failure, 0)
	entryCount, _ := oracleInteger(root["entry_count"])
	if entryCount != len(entries) {
		failures = append(failures, schemaV3Failure{Code: "semantic-count-mismatch", Pointer: "/entry_count"})
	}
	previous := ""
	for index, raw := range entries {
		entry := raw.(map[string]any)
		pointer := fmt.Sprintf("/entries/%d", index)
		schemaPath := entry["schema_path"].(string)
		if previous != "" && schemaPath < previous {
			failures = append(failures, schemaV3Failure{Code: "semantic-canonical-order", Pointer: pointer + "/schema_path"})
		}
		previous = schemaPath
		if freeze, exists := entry["decision_freeze"].(map[string]any); exists {
			freezeBytes, marshalErr := canonicalMarshal(freeze, maximumSchemaV3ArtifactBytes)
			if marshalErr != nil {
				failures = append(failures, schemaV3Failure{Code: "semantic-unsupported", Pointer: pointer + "/decision_freeze", Detail: marshalErr.Error()})
			} else if !bytes.Equal(freezeBytes, rootFreezeBytes) {
				failures = append(failures, schemaV3Failure{Code: "semantic-external-identity", Pointer: pointer + "/decision_freeze"})
			}
		}
		for _, field := range []string{"decision_freeze", "historical_baseline", "accepted_corpus", "rejected_corpus", "forward_oracle", "resolved_reference_graph"} {
			locator, exists := entry[field].(map[string]any)
			if !exists {
				continue
			}
			if locatorRelease, exists := locator["release"].(string); exists && locatorRelease != release {
				failures = append(failures, schemaV3Failure{Code: "semantic-release-identity", Pointer: pointer + "/" + field + "/release"})
			}
		}
	}
	return failures
}

func validateOfficialDownloadsOracleReviewV1Semantics(value any) []schemaV3Failure {
	root := value.(map[string]any)
	oracle := root["oracle"].(map[string]any)
	extractor := root["extractor"].(map[string]any)
	failures := make([]schemaV3Failure, 0)
	if extractor["source_revision"] != oracle["source_revision"] {
		failures = append(failures, schemaV3Failure{Code: "semantic-cross-field", Pointer: "/extractor/source_revision"})
	}
	if !isValidWholeSecondUTC(root["created_at"].(string)) {
		failures = append(failures, schemaV3Failure{Code: "semantic-cross-field", Pointer: "/created_at"})
	}
	return failures
}

func validateAuthorizationRegistryV1Semantics(value any) []schemaV3Failure {
	root := value.(map[string]any)
	entries := root["entries"].([]any)
	failures := make([]schemaV3Failure, 0)
	entryCount, _ := oracleInteger(root["entry_count"])
	if entryCount != len(entries) {
		failures = append(failures, schemaV3Failure{Code: "semantic-count-mismatch", Pointer: "/entry_count"})
	}
	seenIDs := make(map[string]struct{}, len(entries))
	seenSubjects := make(map[string]struct{}, len(entries))
	previousID := ""
	for index, raw := range entries {
		entry := raw.(map[string]any)
		pointer := fmt.Sprintf("/entries/%d", index)
		authorizationID := entry["authorization_id"].(string)
		if _, duplicate := seenIDs[authorizationID]; duplicate {
			failures = append(failures, schemaV3Failure{Code: "semantic-duplicate", Pointer: pointer + "/authorization_id"})
		} else {
			seenIDs[authorizationID] = struct{}{}
		}
		if previousID != "" && authorizationID < previousID {
			failures = append(failures, schemaV3Failure{Code: "semantic-canonical-order", Pointer: pointer})
		}
		previousID = authorizationID

		subject := entry["subject"].(map[string]any)
		subjectBytes, err := canonicalMarshal(subject, maximumSchemaV3ArtifactBytes)
		if err != nil {
			failures = append(failures, schemaV3Failure{Code: "semantic-unsupported", Pointer: pointer + "/subject", Detail: err.Error()})
		} else if subjectKey := string(subjectBytes); subjectKey != "" {
			if _, duplicate := seenSubjects[subjectKey]; duplicate {
				failures = append(failures, schemaV3Failure{Code: "semantic-duplicate", Pointer: pointer + "/subject"})
			} else {
				seenSubjects[subjectKey] = struct{}{}
			}
		}

		recordPath := entry["record"].(map[string]any)["path"].(string)
		if len(recordPath) > 4096 || !safeRelativePath(recordPath) {
			failures = append(failures, schemaV3Failure{Code: "semantic-path", Pointer: pointer + "/record/path"})
		}
		if receiptPath, exists := subject["receipt_path"].(string); exists && (len(receiptPath) > 4096 || !safeRelativePath(receiptPath)) {
			failures = append(failures, schemaV3Failure{Code: "semantic-path", Pointer: pointer + "/subject/receipt_path"})
		}
	}
	return failures
}

func validateAndNormalizeSchemaV3(identity string, input []byte) (schemaV3NormalizedValue, error) {
	maximumBytes := schemaV3ArtifactByteLimit(identity)
	if len(input) > maximumBytes {
		return schemaV3NormalizedValue{}, schemaV3Failure{Code: "limit-artifact-bytes"}
	}
	if len(input) >= 3 && bytes.Equal(input[:3], []byte{0xef, 0xbb, 0xbf}) {
		return schemaV3NormalizedValue{}, schemaV3Failure{Code: "json-bom"}
	}
	if !utf8.Valid(input) {
		return schemaV3NormalizedValue{}, schemaV3Failure{Code: "json-invalid-utf8"}
	}
	compiled, exists := compiledSchemaV3Schemas[identity]
	if !exists {
		return schemaV3NormalizedValue{}, fmt.Errorf("unknown schema-v3 identity %q", identity)
	}
	budget := newTrustJSONBudget(int64(maximumBytes))
	limits := trustJSONParserLimits{}
	if identity == goOfficialDownloadsOracleV1SchemaIdentity {
		limits = officialDownloadsOracleTrustJSONLimits()
	}
	value, err := parseTrustJSONWithBudgetAndLimits(input, int64(maximumBytes), budget, limits)
	if err != nil {
		if failure, ok := reduceTrustJSONError(err); ok {
			return schemaV3NormalizedValue{}, failure
		}
		return schemaV3NormalizedValue{}, err
	}
	if err := compiled.Validate(value); err != nil {
		var validationError *jsonschema.ValidationError
		if !errors.As(err, &validationError) {
			return schemaV3NormalizedValue{}, err
		}
		failure, reduceErr := reduceSchemaV3ValidationError(validationError)
		if reduceErr != nil {
			return schemaV3NormalizedValue{}, reduceErr
		}
		return schemaV3NormalizedValue{}, failure
	}
	if validate := schemaV3SemanticValidators[identity]; validate != nil {
		if failure, ok := selectSchemaV3Failure(validate(value)); ok {
			return schemaV3NormalizedValue{}, failure
		}
	}
	normalized, err := canonicalizeTrustJSONValue(value, len(input), budget)
	if err != nil {
		if failure, ok := reduceTrustJSONError(err); ok {
			return schemaV3NormalizedValue{}, failure
		}
		return schemaV3NormalizedValue{}, err
	}
	if identity == authorizationRecordV1SchemaIdentity && !bytes.Equal(input, normalized) {
		return schemaV3NormalizedValue{}, schemaV3Failure{Code: "canonical-encoding-required"}
	}
	return schemaV3NormalizedValue{Value: value, Normalized: normalized, NormalizedSHA256: exactBytesSHA256(normalized)}, nil
}

func schemaV3ArtifactByteLimit(identity string) int {
	switch identity {
	case goOfficialDownloadsOracleV1SchemaIdentity:
		return maximumOfficialOracleSchemaV3ArtifactBytes
	case residualInventoryV1SchemaIdentity,
		"https://github.com/faustbrian/go-library-tools/schema/cohesion-catalog-v1.schema.json",
		"https://github.com/faustbrian/go-library-tools/schema/cohesion-catalog-v2.schema.json":
		return maximumSchemaV3ArtifactBytes
	default:
		return maximumDefaultSchemaV3ArtifactBytes
	}
}

func validateOfficialDownloadsOracleV1Semantics(value any) []schemaV3Failure {
	root := value.(map[string]any)
	accepted := root["accepted"].([]any)
	rejected := root["rejected"].([]any)
	failures := make([]schemaV3Failure, 0)
	acceptedCount, _ := oracleInteger(root["accepted_count"])
	if acceptedCount != len(accepted) {
		failures = append(failures, schemaV3Failure{Code: "semantic-count-mismatch", Pointer: "/accepted_count"})
	}
	rejectedCount, _ := oracleInteger(root["rejected_count"])
	if rejectedCount != len(rejected) {
		failures = append(failures, schemaV3Failure{Code: "semantic-count-mismatch", Pointer: "/rejected_count"})
	}
	caseIDs := make(map[string]string, len(accepted)+len(rejected))
	for _, collection := range []struct {
		name string
		rows []any
	}{{"accepted", accepted}, {"rejected", rejected}} {
		previous := ""
		for index, raw := range collection.rows {
			row := raw.(map[string]any)
			pointer := fmt.Sprintf("/%s/%d", collection.name, index)
			caseID := row["case_id"].(string)
			if previous != "" && caseID <= previous {
				code := schemaV3ErrorCode("semantic-canonical-order")
				if caseID == previous {
					code = "semantic-duplicate"
				}
				failures = append(failures, schemaV3Failure{Code: code, Pointer: pointer + "/case_id"})
			}
			previous = caseID
			if first, exists := caseIDs[caseID]; exists {
				failures = append(failures, schemaV3Failure{Code: "semantic-duplicate", Pointer: pointer + "/case_id", Detail: "duplicates " + first})
			} else {
				caseIDs[caseID] = pointer
			}
			if expected, exists := row["expected"].([]any); exists {
				previousTuple := ""
				for expectedIndex, rawExpected := range expected {
					expectedRow := rawExpected.(map[string]any)
					tuple := strings.Join([]string{expectedRow["version"].(string), expectedRow["goos"].(string), expectedRow["goarch"].(string), expectedRow["official_filename"].(string)}, "\x00")
					if previousTuple != "" && tuple <= previousTuple {
						failures = append(failures, schemaV3Failure{Code: "semantic-canonical-order", Pointer: fmt.Sprintf("%s/expected/%d", pointer, expectedIndex)})
					}
					previousTuple = tuple
				}
			}
		}
	}
	return failures
}

func validateGoToolchainsV1Semantics(value any) []schemaV3Failure {
	root := value.(map[string]any)
	entries := root["entries"].([]any)
	failures := make([]schemaV3Failure, 0)
	entryCount, _ := oracleInteger(root["entry_count"])
	if entryCount != len(entries) {
		failures = append(failures, schemaV3Failure{Code: "semantic-count-mismatch", Pointer: "/entry_count"})
	}
	if retrievedAt := root["official_index"].(map[string]any)["retrieved_at"].(string); !isValidWholeSecondUTC(retrievedAt) {
		failures = append(failures, schemaV3Failure{Code: "semantic-cross-field", Pointer: "/official_index/retrieved_at"})
	}
	previous := ""
	seen := make(map[string]struct{}, len(entries))
	for index, raw := range entries {
		entry := raw.(map[string]any)
		pointer := fmt.Sprintf("/entries/%d", index)
		tuple := strings.Join([]string{entry["version"].(string), entry["goos"].(string), entry["goarch"].(string)}, "\x00")
		if _, duplicate := seen[tuple]; duplicate {
			failures = append(failures, schemaV3Failure{Code: "semantic-duplicate", Pointer: pointer})
		} else {
			seen[tuple] = struct{}{}
		}
		if previous != "" && tuple < previous {
			failures = append(failures, schemaV3Failure{Code: "semantic-canonical-order", Pointer: pointer})
		}
		previous = tuple
		if entry["distribution_url"].(string) != "https://go.dev/dl/"+entry["official_filename"].(string) {
			failures = append(failures, schemaV3Failure{Code: "semantic-cross-field", Pointer: pointer + "/distribution_url"})
		}
	}
	return failures
}

func validateEngineeringIdentitiesV1Semantics(value any) []schemaV3Failure {
	root := value.(map[string]any)
	identities := root["identities"].([]any)
	failures := make([]schemaV3Failure, 0)
	identityCount, _ := oracleInteger(root["identity_count"])
	if identityCount != len(identities) {
		failures = append(failures, schemaV3Failure{Code: "semantic-count-mismatch", Pointer: "/identity_count"})
	}
	previous := ""
	seenIdentities := make(map[string]struct{}, len(identities))
	seenPaths := make(map[string]struct{}, len(identities))
	for index, raw := range identities {
		entry := raw.(map[string]any)
		pointer := fmt.Sprintf("/identities/%d", index)
		identity := entry["identity"].(string)
		repository := entry["repository"].(string)
		path := entry["path"].(string)
		kind := entry["kind"].(string)
		tuple := strings.Join([]string{identity, kind, path}, "\x00")
		if previous != "" && tuple < previous {
			failures = append(failures, schemaV3Failure{Code: "semantic-canonical-order", Pointer: pointer})
		}
		previous = tuple
		if _, duplicate := seenIdentities[identity]; duplicate {
			failures = append(failures, schemaV3Failure{Code: "semantic-duplicate", Pointer: pointer + "/identity"})
		} else {
			seenIdentities[identity] = struct{}{}
		}
		repositoryPath := repository + "\x00" + path
		if _, duplicate := seenPaths[repositoryPath]; duplicate {
			failures = append(failures, schemaV3Failure{Code: "semantic-duplicate", Pointer: pointer + "/path"})
		} else {
			seenPaths[repositoryPath] = struct{}{}
		}
		if len(identity) > 128 {
			failures = append(failures, schemaV3Failure{Code: "semantic-cross-field", Pointer: pointer + "/identity"})
		}
		if path != "." && (len(path) > 4096 || !safeRelativePath(path)) {
			failures = append(failures, schemaV3Failure{Code: "semantic-path", Pointer: pointer + "/path"})
		}
		goalPath := entry["goal_path"].(string)
		if len(goalPath) > 4096 || !safeRelativePath(goalPath) {
			failures = append(failures, schemaV3Failure{Code: "semantic-path", Pointer: pointer + "/goal_path"})
		}
		switch kind {
		case "repository":
			if identity != repository || path != "." {
				failures = append(failures, schemaV3Failure{Code: "semantic-cross-field", Pointer: pointer + "/identity"})
			}
		case "module", "package", "adapter":
			if path == "." || identity != repository+"/"+path {
				failures = append(failures, schemaV3Failure{Code: "semantic-cross-field", Pointer: pointer + "/identity"})
			}
		default:
			if identity != repository+"#"+path {
				failures = append(failures, schemaV3Failure{Code: "semantic-cross-field", Pointer: pointer + "/identity"})
			}
		}
	}
	return failures
}

func isValidWholeSecondUTC(value string) bool {
	parsed, err := time.Parse("2006-01-02T15:04:05Z", value)
	return err == nil && parsed.UTC().Format("2006-01-02T15:04:05Z") == value
}

func validateSchemaV3DecisionControlSemantics(value any) []schemaV3Failure {
	root := value.(map[string]any)
	failures := validateFrozenContractIdentity(root["contract"].(map[string]any), "/contract")
	reviews := root["reviews"].([]any)
	reviewCount, _ := oracleInteger(root["review_count"])
	if reviewCount != len(reviews) {
		failures = append(failures, schemaV3Failure{Code: "semantic-count-mismatch", Pointer: "/review_count"})
	}
	decisionDigest := root["decision"].(map[string]any)["bytes_sha256"].(string)
	seenReviewers := make(map[string]struct{}, len(reviews))
	for index, raw := range reviews {
		review := raw.(map[string]any)
		pointer := fmt.Sprintf("/reviews/%d", index)
		reviewerID := review["reviewer_id"].(string)
		if _, duplicate := seenReviewers[reviewerID]; duplicate {
			failures = append(failures, schemaV3Failure{Code: "semantic-duplicate", Pointer: pointer + "/reviewer_id"})
		} else {
			seenReviewers[reviewerID] = struct{}{}
		}
		if review["reviewed_sha256"].(string) != decisionDigest {
			failures = append(failures, schemaV3Failure{Code: "semantic-digest-mismatch", Pointer: pointer + "/reviewed_sha256"})
		}
	}
	if !isValidWholeSecondUTC(root["created_at"].(string)) {
		failures = append(failures, schemaV3Failure{Code: "semantic-cross-field", Pointer: "/created_at"})
	}
	return failures
}

func validateContractReviewV1Semantics(value any) []schemaV3Failure {
	root := value.(map[string]any)
	contract := root["contract"].(map[string]any)
	failures := validateFrozenContractIdentity(contract, "/contract")
	review := root["review"].(map[string]any)
	if review["reviewed_commit"] != contract["peeled_commit"] {
		failures = append(failures, schemaV3Failure{Code: "semantic-external-identity", Pointer: "/review/reviewed_commit"})
	}
	wantURL := "https://github.com/faustbrian/go-library-tools/releases/tag/" + contract["tag"].(string)
	if root["release"].(map[string]any)["url"] != wantURL {
		failures = append(failures, schemaV3Failure{Code: "semantic-external-identity", Pointer: "/release/url"})
	}
	if !isValidWholeSecondUTC(root["created_at"].(string)) {
		failures = append(failures, schemaV3Failure{Code: "semantic-cross-field", Pointer: "/created_at"})
	}
	return failures
}

func validateContractFreezeV1Semantics(value any) []schemaV3Failure {
	root := value.(map[string]any)
	contract := root["contract"].(map[string]any)
	failures := validateFrozenContractIdentity(contract, "/contract")
	review := root["review"].(map[string]any)
	if review["record_path"] != ".ai/cohesion/phase3/contract-releases/v1.5.5/CONTRACT_REVIEW.json" {
		failures = append(failures, schemaV3Failure{Code: "semantic-external-identity", Pointer: "/review/record_path"})
	}
	if review["reviewer_id"] != "/root/tooling_v154_delivery/v155_release_review" {
		failures = append(failures, schemaV3Failure{Code: "semantic-external-identity", Pointer: "/review/reviewer_id"})
	}
	if review["reviewed_commit"] != contract["peeled_commit"] {
		failures = append(failures, schemaV3Failure{Code: "semantic-external-identity", Pointer: "/review/reviewed_commit"})
	}
	if !isValidWholeSecondUTC(root["created_at"].(string)) {
		failures = append(failures, schemaV3Failure{Code: "semantic-cross-field", Pointer: "/created_at"})
	}
	return failures
}

func validateFrozenContractIdentity(contract map[string]any, pointer string) []schemaV3Failure {
	want := map[string]string{
		"repository":     "github.com/faustbrian/go-library-tools",
		"path":           "docs/ecosystem/goals/cohesion-v1.md",
		"tag":            "v1.5.5",
		"tag_object_sha": "6e54f464376865a60c618cce8f07fe70bcfabed0",
		"peeled_commit":  "66d2874dc98afe43dd7dad379d6f5e7d1b613111",
		"bytes_sha256":   "sha256:6e5ac948c2cf1ec439fdf90c53c0648befb74a65c3eca2eb595737ee8211891c",
	}
	failures := make([]schemaV3Failure, 0)
	for field, expected := range want {
		if contract[field] != expected {
			failures = append(failures, schemaV3Failure{Code: "semantic-external-identity", Pointer: pointer + "/" + field})
		}
	}
	return failures
}

func reduceTrustJSONError(err error) (schemaV3Failure, bool) {
	message := err.Error()
	offset := 0
	_, _ = fmt.Sscanf(message, "trust JSON at byte %d:", &offset)
	for fragment, code := range map[string]schemaV3ErrorCode{
		"byte-order mark":                        "json-bom",
		"valid UTF-8":                            "json-invalid-utf8",
		"lone high surrogate":                    "json-lone-surrogate",
		"lone low surrogate":                     "json-lone-surrogate",
		"Unicode noncharacter":                   "json-noncharacter",
		"negative zero":                          "json-negative-zero",
		"only integer JSON numbers":              "json-noninteger-number",
		"outside the I-JSON interoperable range": "json-integer-overflow",
		"duplicate decoded key":                  "json-duplicate-key",
		"trailing data":                          "json-trailing-value",
		"allocation budget exceeded":             "limit-allocation-capacity",
		"string exceeds its maximum byte length": "limit-string-bytes",
		"more than 4096 members":                 "limit-object-members",
		"more than 4096 elements":                "limit-array-elements",
		"nesting depth exceeds 64":               "limit-depth",
	} {
		if strings.Contains(message, fragment) {
			return schemaV3Failure{Code: code, ByteOffset: offset, Detail: message}, true
		}
	}
	return schemaV3Failure{}, false
}
