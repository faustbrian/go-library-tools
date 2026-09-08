//nolint:copyloopvar,forcetypeassert,modernize,revive // controlled fixture shapes and local test names
package cohesion

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
)

func TestVerifyForwardOracleV2ReconstructsFixturesAndSplices(t *testing.T) {
	t.Parallel()

	fixture := []byte(`{"diagnostics":[{"code":"failure","message":"failed","path":""}],"schema_id":"urn:golib:cohesion:diagnostic:v1","schema_version":1,"status":"failed"}`)
	digest := exactBytesSHA256(fixture)
	oracle := map[string]any{
		"format":        "golib-forward-oracle-v2",
		"fixture_count": 2,
		"fixtures": []any{
			map[string]any{"fixture_id": "base.canonical-minimum", "bytes_base64": base64.StdEncoding.EncodeToString(fixture), "bytes_sha256": digest},
			map[string]any{"fixture_id": "base.canonical-rich", "bytes_base64": base64.StdEncoding.EncodeToString(fixture), "bytes_sha256": digest},
		},
		"case_count": 3,
		"cases": []any{
			map[string]any{"case_id": "base.canonical-minimum", "input": map[string]any{"kind": "fixture", "fixture_id": "base.canonical-minimum"}, "outcome": "accepted", "normalized_value_sha256": digest, "error_code": nil},
			map[string]any{"case_id": "base.canonical-rich", "input": map[string]any{"kind": "fixture", "fixture_id": "base.canonical-rich"}, "outcome": "accepted", "normalized_value_sha256": digest, "error_code": nil},
			map[string]any{"case_id": "schema.unknown-member", "input": map[string]any{"kind": "splice", "fixture_id": "base.canonical-rich", "splices": []any{map[string]any{"offset": 1, "delete_count": 0, "insert_base64": base64.StdEncoding.EncodeToString([]byte(`"extra":null,`))}}}, "outcome": "rejected", "normalized_value_sha256": nil, "error_code": "schema-unknown-member"},
		},
	}
	data, err := canonicalMarshal(oracle, 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	expectation := forwardOracleV2Expectation{SchemaIdentity: diagnosticV3SchemaIdentity, CaseCount: 3, MaximumBytes: 1 << 20, CaseIDs: []string{"base.canonical-minimum", "base.canonical-rich", "schema.unknown-member"}}
	if err := verifyForwardOracleV2(data, expectation); err != nil {
		t.Fatal(err)
	}
}

func TestDecodeForwardOracleV2AllowsLargeFixtureBase64OnlyAtTheFixtureField(t *testing.T) {
	t.Parallel()

	largeFixture := []byte(strings.Repeat("a", maximumTrustJSONString))
	largeBase64 := base64.StdEncoding.EncodeToString(largeFixture)
	smallFixture := []byte(`{}`)
	largeDigest := exactBytesSHA256(largeFixture)
	smallDigest := exactBytesSHA256(smallFixture)
	oracle := map[string]any{
		"format": "golib-forward-oracle-v2", "fixture_count": 2,
		"fixtures": []any{
			map[string]any{"fixture_id": "base.canonical-minimum", "bytes_base64": base64.StdEncoding.EncodeToString(smallFixture), "bytes_sha256": smallDigest},
			map[string]any{"fixture_id": "base.canonical-rich", "bytes_base64": largeBase64, "bytes_sha256": largeDigest},
		},
		"case_count": 2,
		"cases": []any{
			map[string]any{"case_id": "base.canonical-minimum", "input": map[string]any{"kind": "fixture", "fixture_id": "base.canonical-minimum"}, "outcome": "accepted", "normalized_value_sha256": smallDigest, "error_code": nil},
			map[string]any{"case_id": "base.canonical-rich", "input": map[string]any{"kind": "fixture", "fixture_id": "base.canonical-rich"}, "outcome": "accepted", "normalized_value_sha256": largeDigest, "error_code": nil},
		},
	}
	data, err := json.Marshal(oracle)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := decodeForwardOracleV2(data, int64(len(data))); err != nil {
		t.Fatalf("decodeForwardOracleV2(large fixture bytes_base64) error = %v", err)
	}
}

func TestDecodeForwardOracleV2RejectsStructuralDrift(t *testing.T) {
	t.Parallel()

	fixture := []byte(`{}`)
	digest := exactBytesSHA256(fixture)
	valid := map[string]any{
		"format": "golib-forward-oracle-v2", "fixture_count": 2,
		"fixtures": []any{
			map[string]any{"fixture_id": "base.canonical-minimum", "bytes_base64": "e30=", "bytes_sha256": digest},
			map[string]any{"fixture_id": "base.canonical-rich", "bytes_base64": "e30=", "bytes_sha256": digest},
		},
		"case_count": 2,
		"cases": []any{
			map[string]any{"case_id": "base.canonical-minimum", "input": map[string]any{"kind": "fixture", "fixture_id": "base.canonical-minimum"}, "outcome": "accepted", "normalized_value_sha256": digest, "error_code": nil},
			map[string]any{"case_id": "base.canonical-rich", "input": map[string]any{"kind": "fixture", "fixture_id": "base.canonical-rich"}, "outcome": "accepted", "normalized_value_sha256": digest, "error_code": nil},
		},
	}
	mutations := map[string]func(map[string]any){
		"fixture count": func(value map[string]any) { value["fixture_count"] = 1 },
		"bad base64":    func(value map[string]any) { value["fixtures"].([]any)[0].(map[string]any)["bytes_base64"] = "e30" },
		"digest": func(value map[string]any) {
			value["fixtures"].([]any)[0].(map[string]any)["bytes_sha256"] = "sha256:" + string(make([]byte, 64))
		},
		"case order": func(value map[string]any) { cases := value["cases"].([]any); cases[0], cases[1] = cases[1], cases[0] },
		"minimum uses rich fixture": func(value map[string]any) {
			value["cases"].([]any)[0].(map[string]any)["input"].(map[string]any)["fixture_id"] = "base.canonical-rich"
		},
		"non-base uses fixture variant": func(value map[string]any) {
			value["case_count"] = 3
			value["cases"] = append(value["cases"].([]any), map[string]any{
				"case_id": "schema.unknown-member", "input": map[string]any{"kind": "fixture", "fixture_id": "base.canonical-rich"},
				"outcome": "rejected", "normalized_value_sha256": nil, "error_code": "schema-unknown-member",
			})
		},
		"unknown": func(value map[string]any) { value["unknown"] = true },
	}
	for name, mutate := range mutations {
		name, mutate := name, mutate
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			copy := cloneOracleTestValue(t, valid)
			mutate(copy)
			data, err := canonicalMarshal(copy, 1<<20)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := decodeForwardOracleV2(data, 1<<20); err == nil {
				t.Fatal("invalid forward oracle accepted")
			}
		})
	}
}

func TestVerifyForwardOracleV2ReviewBindsExactOracleAndThreeDimensions(t *testing.T) {
	t.Parallel()

	oracle := []byte(`{"cases":[]}`)
	oracleDigest := exactBytesSHA256(oracle)
	expectation := forwardOracleV2Expectation{
		MaximumBytes: 1024,
		OraclePath:   "testdata/cohesion/forward-oracles/cohesion-diagnostic-v1-forward-oracle.json",
	}
	review := map[string]any{
		"decision_sha256":    frozenSchemaV3DecisionSHA256,
		"oracle":             map[string]any{"path": expectation.OraclePath, "source_revision": strings.Repeat("a", 40), "bytes_sha256": oracleDigest},
		"serialized_bytes":   len(oracle),
		"serialized_ceiling": 1024,
		"review_count":       3,
		"reviews": []any{
			map[string]any{"reviewer_id": "oracle-contract-review", "dimension": "contract", "reviewed_sha256": oracleDigest, "outcome": "accepted-no-findings"},
			map[string]any{"reviewer_id": "oracle-execution-review", "dimension": "execution-adversary", "reviewed_sha256": oracleDigest, "outcome": "accepted-no-findings"},
			map[string]any{"reviewer_id": "oracle-verification-review", "dimension": "verification", "reviewed_sha256": oracleDigest, "outcome": "accepted-no-findings"},
		},
		"outcome": "accepted-no-findings",
	}
	encoded, err := canonicalMarshal(review, 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	if err := verifyForwardOracleV2Review(oracle, encoded, expectation); err != nil {
		t.Fatal(err)
	}

	for name, mutate := range map[string]func(map[string]any){
		"serialized bytes": func(value map[string]any) { value["serialized_bytes"] = len(oracle) + 1 },
		"ceiling":          func(value map[string]any) { value["serialized_ceiling"] = 1025 },
		"oracle digest": func(value map[string]any) {
			value["oracle"].(map[string]any)["bytes_sha256"] = "sha256:" + strings.Repeat("0", 64)
		},
		"review order": func(value map[string]any) {
			reviews := value["reviews"].([]any)
			reviews[0], reviews[1] = reviews[1], reviews[0]
		},
		"duplicate reviewer": func(value map[string]any) {
			reviews := value["reviews"].([]any)
			reviews[1].(map[string]any)["reviewer_id"] = reviews[0].(map[string]any)["reviewer_id"]
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			candidate := cloneOracleTestValue(t, review)
			mutate(candidate)
			data, err := canonicalMarshal(candidate, 1<<20)
			if err != nil {
				t.Fatal(err)
			}
			if err := verifyForwardOracleV2Review(oracle, data, expectation); err == nil {
				t.Fatal("invalid forward-oracle review accepted")
			}
		})
	}
}

func cloneOracleTestValue(t *testing.T, value map[string]any) map[string]any {
	t.Helper()
	data, err := canonicalMarshal(value, 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := parseTrustJSONWithBudget(data, int64(len(data)), newTrustJSONBudget(int64(len(data))))
	if err != nil {
		t.Fatal(err)
	}
	return parsed.(map[string]any)
}
