package cohesion

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestVerifySchemaV3BatchReviewRequiresCompleteManifestExpectation(t *testing.T) {
	if err := verifySchemaV3BatchReview([]byte(`{}`), "sha256:"+string(make([]byte, 64)), nil); err == nil {
		t.Fatal("incomplete batch expectation accepted")
	}
}

func TestVerifySchemaV3BatchReviewValidatesAllTwentyEntries(t *testing.T) {
	paths := make([]schemaV3BatchOracleExpectation, 20)
	manifest := make([]any, 20)
	digest := "sha256:" + strings.Repeat("a", 64)
	for index := range paths {
		base := "schema-" + string(rune('a'+index))
		paths[index] = schemaV3BatchOracleExpectation{OraclePath: "testdata/cohesion/forward-oracles/" + base + ".json", ReviewPath: "testdata/cohesion/forward-oracle-reviews/" + base + ".json"}
		manifest[index] = map[string]any{"oracle_path": paths[index].OraclePath, "source_revision": strings.Repeat("b", 40), "oracle_sha256": digest, "case_count": 2, "case_roster_sha256": digest, "expected_results_sha256": digest, "serialized_ceiling": 1024, "review_path": paths[index].ReviewPath, "review_sha256": digest}
	}
	value := map[string]any{
		"schema_id": "urn:golib:cohesion:schema-v3-decision-review:v1", "schema_version": 1,
		"decision": map[string]any{"path": ".ai/cohesion/phase3/SCHEMA_V3_DECISIONS.md", "bytes_sha256": digest},
		"contract": map[string]any{}, "oracle_manifest": manifest,
		"reviews": []any{
			map[string]any{"reviewer_id": "contract-reviewer", "dimension": "contract", "reviewed_sha256": digest, "evidence_sha256": digest, "coverage": []any{"decision"}, "outcome": "accepted-no-findings"},
			map[string]any{"reviewer_id": "adversary-reviewer", "dimension": "execution-adversary", "reviewed_sha256": digest, "evidence_sha256": digest, "coverage": []any{"all-high-risk-cases"}, "outcome": "accepted-no-findings"},
			map[string]any{"reviewer_id": "verification-reviewer", "dimension": "verification", "reviewed_sha256": digest, "evidence_sha256": digest, "coverage": []any{"all-assets"}, "checker_source_sha256": digest, "checker_version": "v1", "outcome": "accepted-no-findings"},
		},
		"created_at": "2026-09-08T00:00:00Z",
	}
	data, err := canonicalMarshal(value, 4<<20)
	if err != nil {
		t.Fatal(err)
	}
	if err := verifySchemaV3BatchReview(data, digest, paths); err != nil {
		t.Fatalf("valid batch review rejected: %v", err)
	}
	manifest[1].(map[string]any)["oracle_path"] = manifest[0].(map[string]any)["oracle_path"]
	mutated, err := canonicalMarshal(value, 4<<20)
	if err != nil {
		t.Fatal(err)
	}
	if err := verifySchemaV3BatchReview(mutated, digest, paths); err == nil {
		t.Fatal("duplicate manifest path accepted")
	}
}

func TestVerifySchemaV3BatchReviewRejectsEachStructuralInvariant(t *testing.T) {
	t.Parallel()
	digest := "sha256:" + strings.Repeat("a", 64)
	paths := make([]schemaV3BatchOracleExpectation, 20)
	manifest := make([]any, 20)
	for index := range paths {
		name := fmt.Sprintf("schema-%02d", index)
		paths[index] = schemaV3BatchOracleExpectation{OraclePath: "oracle/" + name + ".json", ReviewPath: "review/" + name + ".json"}
		manifest[index] = map[string]any{"oracle_path": paths[index].OraclePath, "source_revision": strings.Repeat("b", 40), "oracle_sha256": digest, "case_count": 1, "case_roster_sha256": digest, "expected_results_sha256": digest, "serialized_ceiling": 1, "review_path": paths[index].ReviewPath, "review_sha256": digest}
	}
	value := map[string]any{
		"schema_id": "urn:golib:cohesion:schema-v3-decision-review:v1", "schema_version": 1,
		"decision": map[string]any{"path": "decision.json", "bytes_sha256": digest}, "contract": map[string]any{}, "oracle_manifest": manifest,
		"reviews": []any{
			map[string]any{"reviewer_id": "contract", "dimension": "contract", "reviewed_sha256": digest, "evidence_sha256": digest, "coverage": []any{}, "outcome": "accepted-no-findings"},
			map[string]any{"reviewer_id": "adversary", "dimension": "execution-adversary", "reviewed_sha256": digest, "evidence_sha256": digest, "coverage": []any{}, "outcome": "accepted-no-findings"},
			map[string]any{"reviewer_id": "verification", "dimension": "verification", "reviewed_sha256": digest, "evidence_sha256": digest, "coverage": []any{}, "checker_source_sha256": digest, "checker_version": "v1", "outcome": "accepted-no-findings"},
		}, "created_at": "2026-09-08T00:00:00Z",
	}
	canonical := func() []byte { data, err := canonicalMarshal(value, 4<<20); if err != nil { t.Fatal(err) }; return data }
	if err := verifySchemaV3BatchReview(canonical(), digest, paths); err != nil { t.Fatalf("baseline rejected: %v", err) }
	base := value
	mutations := []func(map[string]any){
		func(value map[string]any) { value["schema_id"] = "wrong" },
		func(value map[string]any) { value["schema_version"] = 2 },
		func(value map[string]any) { value["decision"].(map[string]any)["bytes_sha256"] = "wrong" },
		func(value map[string]any) { value["oracle_manifest"].([]any)[0].(map[string]any)["oracle_sha256"] = "wrong" },
		func(value map[string]any) { value["oracle_manifest"].([]any)[0].(map[string]any)["case_count"] = 0 },
		func(value map[string]any) { value["oracle_manifest"].([]any)[0].(map[string]any)["source_revision"] = "short" },
		func(value map[string]any) { value["reviews"] = []any{} },
	}
	for index, mutate := range mutations {
		value = cloneBatchReviewValue(base)
		mutate(value)
		if err := verifySchemaV3BatchReview(canonical(), digest, paths); err == nil { t.Fatalf("mutation %d accepted", index) }
	}
}

func cloneBatchReviewValue(value map[string]any) map[string]any {
	data, _ := json.Marshal(value)
	var clone map[string]any
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.UseNumber()
	_ = decoder.Decode(&clone)
	return clone
}
