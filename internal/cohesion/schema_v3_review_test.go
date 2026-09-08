package cohesion

import (
	"os"
	"strings"
	"testing"
)

func TestValidateSchemaV3DecisionReviewRunsStrictSchemaAndSemanticValidation(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("../../release/cohesion-schema-v3-decision-review.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := validateSchemaV3DecisionReview(data); err != nil {
		t.Fatalf("validateSchemaV3DecisionReview(valid) error = %v", err)
	}

	for name, replacement := range map[string]string{
		"duplicate reviewer":  "/root/schema_v6_contract_review",
		"mismatched decision": "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"wrong schema":        "urn:golib:cohesion:schema-v3-decision-review:v2",
	} {
		candidate := append([]byte(nil), data...)
		switch name {
		case "duplicate reviewer":
			candidate = replaceJSONOccurrence(candidate, "/root/schema_v6_execution_review", replacement)
		case "mismatched decision":
			candidate = replaceJSONOccurrence(candidate, "sha256:cfa83030b1b05535148292c4db063edaa35aca861e957ecd5579008090059e2f", replacement)
		case "wrong schema":
			candidate = replaceJSONOccurrence(candidate, "urn:golib:cohesion:schema-v3-decision-review:v1", replacement)
		}
		if err := validateSchemaV3DecisionReview(candidate); err == nil {
			t.Fatalf("validateSchemaV3DecisionReview(%s) error = nil", name)
		}
	}

	wrongType := replaceJSONOccurrence(data, `"review_count": 3`, `"review_count": "3"`)
	if err := validateSchemaV3DecisionReview(wrongType); err == nil {
		t.Fatal("validateSchemaV3DecisionReview(wrong type) error = nil")
	}
	if err := validateSchemaV3DecisionReview([]byte(`{"schema_id"`)); err == nil {
		t.Fatal("validateSchemaV3DecisionReview(malformed) error = nil")
	}
	tooManyReviewerBytes := replaceJSONOccurrence(data, "/root/schema_v6_contract_review", strings.Repeat("😀", 128))
	if err := validateSchemaV3DecisionReview(tooManyReviewerBytes); err == nil {
		t.Fatal("validateSchemaV3DecisionReview(128 four-byte reviewer runes) error = nil")
	}
	invalidDate := replaceJSONOccurrence(data, "2026-09-08T06:39:09Z", "2026-02-31T06:39:09Z")
	if err := validateSchemaV3DecisionReview(invalidDate); err == nil {
		t.Fatal("validateSchemaV3DecisionReview(impossible date) error = nil")
	}
	if err := validateSchemaV3DecisionReviewSemantics(schemaV3DecisionReview{
		CreatedAt: "2026-09-08T06:39:09Z",
		Reviews:   []schemaV3DecisionReviewEntry{{ReviewerID: strings.Repeat("😀", 128)}},
	}); err == nil {
		t.Fatal("validateSchemaV3DecisionReviewSemantics(oversized reviewer bytes) error = nil")
	}
	const digest = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	if err := validateSchemaV3DecisionReviewSemantics(schemaV3DecisionReview{
		CreatedAt: "2026-09-08T06:39:09Z",
		Decision:  schemaV3ReviewedDecision{BytesSHA256: digest},
		Reviews: []schemaV3DecisionReviewEntry{{
			ReviewerID: strings.Repeat("a", 128), ReviewedSHA256: digest,
		}},
	}); err != nil {
		t.Fatalf("validateSchemaV3DecisionReviewSemantics(exact reviewer boundary) error = %v", err)
	}
}

func TestCompiledSchemaV3SetContainsEveryVersionedContract(t *testing.T) {
	t.Parallel()

	identities := []string{
		"modules-v1.schema.json", "modules-v2.schema.json", "modules-v3.schema.json",
		"cohesion-catalog-v1.schema.json", "cohesion-catalog-v2.schema.json",
		"cohesion-inputs-v1.schema.json", "cohesion-inputs-v2.schema.json",
		"cohesion-sources-v1.schema.json", "cohesion-sources-v2.schema.json",
		"cohesion-delivery-evidence-v1.schema.json",
		"cohesion-authorization-registry-v1.schema.json",
		"cohesion-authorization-record-v1.schema.json",
		"cohesion-residual-inventory-v1.schema.json",
		"cohesion-residual-register-v1.schema.json",
		"cohesion-gate-policy-v1.schema.json",
		"cohesion-go-toolchains-v1.schema.json",
		"cohesion-go-official-downloads-oracle-v1.schema.json",
		"cohesion-go-official-downloads-oracle-review-v1.schema.json",
		"cohesion-engineering-identities-v1.schema.json",
		"cohesion-schema-provenance-v1.schema.json",
		"cohesion-schema-v3-decision-review-v1.schema.json",
		"cohesion-schema-v3-decision-freeze-v1.schema.json",
		"cohesion-contract-freeze-v1.schema.json",
		"cohesion-contract-review-v1.schema.json",
		"cohesion-diagnostic-v1.schema.json",
	}
	if len(compiledSchemaV3Schemas) != len(identities) {
		t.Fatalf("compiled schema-v3 count = %d, want %d", len(compiledSchemaV3Schemas), len(identities))
	}
	for _, file := range identities {
		identity := "https://github.com/faustbrian/go-library-tools/schema/" + file
		if compiledSchemaV3Schemas[identity] == nil {
			t.Fatalf("compiled schema-v3 set is missing %s", identity)
		}
	}
}

func TestValidateSchemaV3BootstrapControlsRejectsCrossArtifactSubstitutions(t *testing.T) {
	t.Parallel()

	load := func(name string) []byte {
		t.Helper()
		data, err := os.ReadFile("../../release/" + name)
		if err != nil {
			t.Fatal(err)
		}
		return data
	}
	contractReview := load("cohesion-contract-review.json")
	contractFreeze := load("cohesion-contract-freeze.json")
	decisionReview := load("cohesion-schema-v3-decision-review.json")
	decisionFreeze := load("cohesion-schema-v3-decision-freeze.json")
	if err := validateSchemaV3BootstrapControls(contractReview, contractFreeze, decisionReview, decisionFreeze); err != nil {
		t.Fatalf("validateSchemaV3BootstrapControls(valid) error = %v", err)
	}
	if err := validateSchemaV3BootstrapControlSemantics(contractReview, contractFreeze, decisionReview, decisionFreeze); err != nil {
		t.Fatalf("validateSchemaV3BootstrapControlSemantics(valid) error = %v", err)
	}

	tests := map[string][4][]byte{
		"contract release URL and tag disagree": {
			replaceJSONOccurrence(contractReview, `"tag": "v1.5.5"`, `"tag": "v1.5.7"`),
			contractFreeze, decisionReview, decisionFreeze,
		},
		"contract freeze review digest changes": {
			contractReview,
			replaceJSONOccurrence(contractFreeze, "sha256:0793cd5175f84fdbafe4b13590d1cb48cd2450bcd2b0a4189986fcdf3e78df9a", "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"),
			decisionReview, decisionFreeze,
		},
		"decision review contract changes": {
			contractReview, contractFreeze,
			replaceJSONOccurrence(decisionReview, `"tag": "v1.5.5"`, `"tag": "v1.5.7"`),
			decisionFreeze,
		},
		"decision freeze review digest changes": {
			contractReview, contractFreeze, decisionReview,
			replaceJSONOccurrence(decisionFreeze, "sha256:c371d51fc8bf5b314d8c7cf51e8b96b7d4714397145cce09cc723f8370fa628d", "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"),
		},
		"decision freeze review set changes": {
			contractReview, contractFreeze, decisionReview,
			replaceJSONOccurrence(decisionFreeze, "/root/schema_v6_execution_review", "/root/schema_v6_execution_changed"),
		},
		"decision freeze timestamp changes": {
			contractReview, contractFreeze, decisionReview,
			replaceJSONOccurrence(decisionFreeze, "2026-09-08T06:39:09Z", "2026-09-08T06:39:10Z"),
		},
	}
	for name, documents := range tests {
		if err := validateSchemaV3BootstrapControlSemantics(documents[0], documents[1], documents[2], documents[3]); err == nil {
			t.Fatalf("validateSchemaV3BootstrapControlSemantics(%s) error = nil", name)
		}
	}
	whitespaceChanged := append(append([]byte(nil), contractReview...), '\n')
	if err := validateSchemaV3BootstrapControls(whitespaceChanged, contractFreeze, decisionReview, decisionFreeze); err == nil {
		t.Fatal("validateSchemaV3BootstrapControls(exact-byte drift) error = nil")
	}
}

func replaceJSONOccurrence(value []byte, old, replacement string) []byte {
	oldBytes := []byte(old)
	for index := 0; index+len(oldBytes) <= len(value); index++ {
		if string(value[index:index+len(oldBytes)]) == old {
			result := append([]byte(nil), value...)
			return append(append(result[:index], replacement...), result[index+len(oldBytes):]...)
		}
	}
	return value
}
