//nolint:copyloopvar,errorlint,forcetypeassert,modernize // controlled fixture shapes and value errors
package cohesion

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
)

const diagnosticV3SchemaIdentity = "https://github.com/faustbrian/go-library-tools/schema/cohesion-diagnostic-v1.schema.json"

func TestValidateAndNormalizeSchemaV3AcceptsEquivalentEncoding(t *testing.T) {
	t.Parallel()

	canonical := []byte(`{"diagnostics":[{"code":"failure","message":"failed","path":""}],"schema_id":"urn:golib:cohesion:diagnostic:v1","schema_version":1,"status":"failed"}`)
	noncanonical := append([]byte(`{ `), canonical[1:]...)
	want, err := validateAndNormalizeSchemaV3(diagnosticV3SchemaIdentity, canonical)
	if err != nil {
		t.Fatal(err)
	}
	got, err := validateAndNormalizeSchemaV3(diagnosticV3SchemaIdentity, noncanonical)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got.Normalized, canonical) || got.NormalizedSHA256 != want.NormalizedSHA256 {
		t.Fatalf("noncanonical normalization = (%s, %s), want (%s, %s)", got.Normalized, got.NormalizedSHA256, canonical, want.NormalizedSHA256)
	}
}

func TestValidateAndNormalizeSchemaV3UsesTheIdentitySpecificArtifactLimit(t *testing.T) {
	t.Parallel()

	_, err := validateAndNormalizeSchemaV3(diagnosticV3SchemaIdentity, make([]byte, (1<<20)+1))
	failure, ok := err.(schemaV3Failure)
	if !ok || failure.Code != "limit-artifact-bytes" {
		t.Fatalf("diagnostic oversized error = %#v, want limit-artifact-bytes", err)
	}

	if got := schemaV3ArtifactByteLimit(goOfficialDownloadsOracleV1SchemaIdentity); got != 128<<20 {
		t.Fatalf("official oracle limit = %d, want %d", got, 128<<20)
	}
	if got := schemaV3ArtifactByteLimit("https://github.com/faustbrian/go-library-tools/schema/cohesion-catalog-v2.schema.json"); got != 32<<20 {
		t.Fatalf("catalog v2 limit = %d, want %d", got, 32<<20)
	}
}

func TestValidateSourcesV2SemanticsRejectsOrderDuplicatesAndControlCounts(t *testing.T) {
	t.Parallel()

	base := map[string]any{
		"repository_count": json.Number("2"),
		"sources": []any{
			map[string]any{"repository": "github.com/faustbrian/go-z"},
			map[string]any{"repository": "github.com/faustbrian/go-a"},
		},
		"controls": map[string]any{
			"authorization_records":   map[string]any{"count": json.Number("1"), "entries": []any{}},
			"registry_predecessors":   map[string]any{"count": json.Number("0"), "entries": []any{}},
			"effective_policy_inputs": map[string]any{"count": json.Number("0"), "entries": []any{}},
		},
	}
	failures := validateSourcesV2Semantics(base)
	assertSchemaV3FailurePresent(t, failures, "semantic-canonical-order", "/sources/1/repository")
	assertSchemaV3FailurePresent(t, failures, "semantic-count-mismatch", "/controls/authorization_records/count")

	base["sources"] = []any{
		map[string]any{"repository": "github.com/faustbrian/go-a"},
		map[string]any{"repository": "github.com/faustbrian/go-a"},
	}
	failures = validateSourcesV2Semantics(base)
	assertSchemaV3FailurePresent(t, failures, "semantic-duplicate", "/sources/1/repository")
}

func TestValidateDeliveryEvidenceV1SemanticsBindsCountIdentityAndInputManifestDigest(t *testing.T) {
	inputManifest := map[string]any{"ordinary": []any{}, "normalized_manifests": []any{}}
	manifestBytes, err := canonicalMarshal(inputManifest, maximumDefaultSchemaV3ArtifactBytes)
	if err != nil {
		t.Fatal(err)
	}
	digest := exactBytesSHA256(manifestBytes)
	base := map[string]any{
		"entry_count": json.Number("1"),
		"entries": []any{map[string]any{
			"entry_id": "receipt.1", "requirements_sha256": schemaV3GoalRequirementsSHA256,
			"input_manifest": inputManifest, "input_manifest_sha256": digest,
		}},
	}
	if failures := validateDeliveryEvidenceV1Semantics(base); len(failures) != 0 {
		t.Fatalf("validateDeliveryEvidenceV1Semantics(valid) = %#v", failures)
	}
	base["entry_count"] = json.Number("2")
	base["entries"].([]any)[0].(map[string]any)["input_manifest_sha256"] = "sha256:" + strings.Repeat("f", 64)
	failures := validateDeliveryEvidenceV1Semantics(base)
	assertContainsSchemaV3Failure(t, failures, "semantic-count-mismatch", "/entry_count")
	assertContainsSchemaV3Failure(t, failures, "semantic-digest-mismatch", "/entries/0/input_manifest_sha256")
}

func TestValidateDeliveryEvidenceV1SemanticsChecksOptionalPolicyDigest(t *testing.T) {
	manifest := map[string]any{"ordinary": []any{}, "normalized_manifests": []any{}}
	policy := map[string]any{"mode": "strict", "policy_sha256": ""}
	digest, err := semanticObjectSHA256(policy, "policy_sha256", maximumDefaultSchemaV3ArtifactBytes)
	if err != nil {
		t.Fatal(err)
	}
	policy["policy_sha256"] = digest
	base := map[string]any{"entry_count": json.Number("1"), "entries": []any{map[string]any{
		"entry_id": "receipt.1", "requirements_sha256": schemaV3GoalRequirementsSHA256,
		"input_manifest": manifest, "input_manifest_sha256": mustCanonicalDigest(t, manifest), "policy": policy,
	}}}
	if failures := validateDeliveryEvidenceV1Semantics(base); len(failures) != 0 {
		t.Fatalf("valid policy = %#v", failures)
	}
	policy["policy_sha256"] = "sha256:" + strings.Repeat("f", 64)
	assertSchemaV3FailurePresent(t, validateDeliveryEvidenceV1Semantics(base), "semantic-digest-mismatch", "/entries/0/policy/policy_sha256")
}

func mustCanonicalDigest(t *testing.T, value any) string {
	t.Helper()
	data, err := canonicalMarshal(value, maximumDefaultSchemaV3ArtifactBytes)
	if err != nil {
		t.Fatal(err)
	}
	return exactBytesSHA256(data)
}

func TestValidateModulesV3SemanticsReportsMarshalAndManifestFailures(t *testing.T) {
	t.Parallel()
	if failures := validateModulesV3Semantics(map[string]any{"unsupported": func() {}}); len(failures) != 1 || failures[0].Code != "semantic-unsupported" {
		t.Fatalf("unsupported value failures = %#v", failures)
	}
	if failures := validateModulesV3Semantics(map[string]any{"modules": []any{}}); len(failures) != 1 || failures[0].Code != "semantic-cross-field" {
		t.Fatalf("invalid manifest failures = %#v", failures)
	}
}

func TestValidateGatePolicyV1SemanticsBindsPolicyDigest(t *testing.T) {
	policy := map[string]any{"rules": []any{"cohesion"}, "policy_sha256": ""}
	digest, err := semanticObjectSHA256(policy, "policy_sha256", maximumSchemaV3ArtifactBytes)
	if err != nil {
		t.Fatal(err)
	}
	policy["policy_sha256"] = digest
	if failures := validateGatePolicyV1Semantics(policy); len(failures) != 0 {
		t.Fatalf("valid policy failures = %#v", failures)
	}
	policy["policy_sha256"] = "sha256:" + strings.Repeat("0", 64)
	assertSchemaV3FailurePresent(t, validateGatePolicyV1Semantics(policy), "semantic-digest-mismatch", "/policy_sha256")
}

func TestValidateAuthorizationRecordV1SemanticsRejectsUnsafeReceiptPathAndImpossibleTimestamp(t *testing.T) {
	value := map[string]any{
		"created_at": "2026-02-30T00:00:00Z",
		"subject":    map[string]any{"kind": "delivery", "receipt_path": "../receipt.json"},
	}
	failures := validateAuthorizationRecordV1Semantics(value)
	assertContainsSchemaV3Failure(t, failures, "semantic-path", "/subject/receipt_path")
	assertContainsSchemaV3Failure(t, failures, "semantic-cross-field", "/created_at")
}

func TestValidateInputsV2SemanticsRejectsProjectionCountOrderAndDuplicates(t *testing.T) {
	t.Parallel()

	value := map[string]any{
		"projection_count": json.Number("3"),
		"projections": []any{
			map[string]any{"repository": "github.com/faustbrian/go-z"},
			map[string]any{"repository": "github.com/faustbrian/go-a"},
			map[string]any{"repository": "github.com/faustbrian/go-a"},
		},
	}
	failures := validateInputsV2Semantics(value)
	assertSchemaV3FailurePresent(t, failures, "semantic-canonical-order", "/projections/1/repository")
	assertSchemaV3FailurePresent(t, failures, "semantic-duplicate", "/projections/2/repository")

	value["projection_count"] = json.Number("2")
	failures = validateInputsV2Semantics(value)
	assertSchemaV3FailurePresent(t, failures, "semantic-count-mismatch", "/projection_count")
}

func TestValidateCatalogV2SemanticsRejectsLegacyFinalInput(t *testing.T) {
	t.Parallel()

	for _, version := range []json.Number{"1", "2"} {
		failures := validateCatalogV2Semantics(map[string]any{
			"scope":                   "repository",
			"manifest_schema_version": version,
			"publication_status":      "final-input",
		})
		assertSchemaV3FailurePresent(t, failures, "semantic-cross-field", "/publication_status")
	}
}

func assertSchemaV3FailurePresent(t *testing.T, failures []schemaV3Failure, code schemaV3ErrorCode, pointer string) {
	t.Helper()
	for _, failure := range failures {
		if failure.Code == code && failure.Pointer == pointer {
			return
		}
	}
	t.Fatalf("failures = %#v, missing %s at %s", failures, code, pointer)
}

func TestValidateAndNormalizeSchemaV3ReducesFrozenFailures(t *testing.T) {
	t.Parallel()

	valid := `{"diagnostics":[{"code":"failure","message":"failed","path":""}],"schema_id":"urn:golib:cohesion:diagnostic:v1","schema_version":1,"status":"failed"}`
	tests := []struct {
		name string
		data []byte
		code schemaV3ErrorCode
	}{
		{"artifact", make([]byte, maximumSchemaV3ArtifactBytes+1), "limit-artifact-bytes"},
		{"bom", append([]byte{0xef, 0xbb, 0xbf}, valid...), "json-bom"},
		{"invalid utf8", []byte{'{', '"', 0xff, '"', ':', '0', '}'}, "json-invalid-utf8"},
		{"trailing", []byte(valid + ` null`), "json-trailing-value"},
		{"duplicate", []byte(`{"diagnostics":[{"code":"failure","message":"failed","path":""}],"schema_id":"urn:golib:cohesion:diagnostic:v1","schema_id":"urn:golib:cohesion:diagnostic:v1","schema_version":1,"status":"failed"}`), "json-duplicate-key"},
		{"lone surrogate", []byte(`{"schema_id":"\ud800"}`), "json-lone-surrogate"},
		{"noncharacter", []byte(`{"schema_id":"\uffff"}`), "json-noncharacter"},
		{"negative zero", []byte(`{"schema_id":"x","value":-0}`), "json-negative-zero"},
		{"fraction", []byte(`{"schema_id":"x","value":0.5}`), "json-noninteger-number"},
		{"overflow", []byte(`{"schema_id":"x","value":9007199254740993}`), "json-integer-overflow"},
		{"missing", []byte(`{"diagnostics":[{"code":"failure","message":"failed","path":""}],"schema_version":1,"status":"failed"}`), "schema-required-member"},
		{"unknown", []byte(`{"diagnostics":[{"code":"failure","message":"failed","path":""}],"extra":null,"schema_id":"urn:golib:cohesion:diagnostic:v1","schema_version":1,"status":"failed"}`), "schema-unknown-member"},
		{"type", []byte(`{"diagnostics":[{"code":"failure","message":"failed","path":""}],"schema_id":null,"schema_version":1,"status":"failed"}`), "schema-type"},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_, err := validateAndNormalizeSchemaV3(diagnosticV3SchemaIdentity, test.data)
			failure, ok := err.(schemaV3Failure)
			if !ok || failure.Code != test.code {
				t.Fatalf("error = %#v, want schema-v3 code %q", err, test.code)
			}
		})
	}
}

func TestValidateAndNormalizeSchemaV3ReportsTheSecondDecodedKeyOffset(t *testing.T) {
	t.Parallel()

	input := []byte(`{"schema_id":"x","schema_id":"x"}`)
	_, err := validateAndNormalizeSchemaV3(diagnosticV3SchemaIdentity, input)
	failure, ok := err.(schemaV3Failure)
	if !ok {
		t.Fatalf("validateAndNormalizeSchemaV3() error = %T %v, want schemaV3Failure", err, err)
	}
	if failure.Code != "json-duplicate-key" || failure.ByteOffset != 17 {
		t.Fatalf("validateAndNormalizeSchemaV3() failure = %+v, want duplicate-key at byte 17", failure)
	}
}

func TestValidateAndNormalizeModulesV3UsesSchemaWrongStatusCode(t *testing.T) {
	t.Parallel()

	states := []DimensionState{DimensionBlocked, DimensionInProgress, DimensionNotApplicable, DimensionNotStarted, DimensionVerified}
	statuses := []GoalStatus{GoalBlocked, GoalComplete, GoalInProgress, GoalNotApplicable, GoalNotStarted}
	requirements := "sha256:6e5ac948c2cf1ec439fdf90c53c0648befb74a65c3eca2eb595737ee8211891c"
	fixture := manifestV3Fixture(t, requirements)
	mismatches := make(map[schemaV3ErrorCode]int)
	validFailures := 0
	cases := 0
	for _, implementation := range states {
		for _, hardening := range states {
			for _, release := range states {
				wantStatus := deriveGoalStatus(implementation, hardening)
				if _, err := validateAndNormalizeSchemaV3(modulesV3SchemaIdentity, modulesV3StateCase(t, fixture, implementation, hardening, release, wantStatus)); err != nil {
					validFailures++
				}
				for _, wrongStatus := range statuses {
					if wrongStatus == wantStatus {
						continue
					}
					cases++
					_, err := validateAndNormalizeSchemaV3(modulesV3SchemaIdentity, modulesV3StateCase(t, fixture, implementation, hardening, release, wrongStatus))
					failure, ok := err.(schemaV3Failure)
					if !ok || failure.Code != "schema-union" {
						if ok {
							mismatches[failure.Code]++
						} else {
							mismatches["semantic-unsupported"]++
						}
					}
				}
			}
		}
	}
	if validFailures != 0 || cases != 500 || len(mismatches) != 0 {
		t.Fatalf("valid failures = %d; wrong-status cases = %d with mismatches %v, want 125 accepted and 500 schema-union failures", validFailures, cases, mismatches)
	}
}

func TestValidateAndNormalizeSchemaV3AppliesIntrinsicCollectionSemantics(t *testing.T) {
	t.Parallel()

	digest := "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	commit := "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	tests := []struct {
		name     string
		identity string
		value    map[string]any
		code     schemaV3ErrorCode
	}{
		{
			name:     "official oracle count",
			identity: "https://github.com/faustbrian/go-library-tools/schema/cohesion-go-official-downloads-oracle-v1.schema.json",
			value: map[string]any{
				"schema_id": "urn:golib:cohesion:go-official-downloads-oracle:v1", "schema_version": 1,
				"official_index_sha256": digest, "accepted_count": 2,
				"accepted":       []any{map[string]any{"case_id": "accepted", "input_base64": "e30=", "expected": []any{}}},
				"rejected_count": 1, "rejected": []any{map[string]any{"case_id": "rejected", "input_base64": "e30=", "error_code": "invalid"}},
			},
			code: "semantic-count-mismatch",
		},
		{
			name:     "toolchain filename URL",
			identity: "https://github.com/faustbrian/go-library-tools/schema/cohesion-go-toolchains-v1.schema.json",
			value: map[string]any{
				"schema_id": "urn:golib:cohesion:go-toolchains:v1", "schema_version": 1,
				"official_index": map[string]any{"source_url": "https://go.dev/dl/?mode=json&include=all", "retrieved_at": "2026-09-08T00:00:00Z", "asset": "cohesion-go-official-downloads.json", "bytes_sha256": digest},
				"entry_count":    1,
				"entries":        []any{map[string]any{"version": "go1.26.0", "goos": "linux", "goarch": "amd64", "official_filename": "go1.26.0.linux-amd64.tar.gz", "distribution_url": "https://go.dev/dl/other.tar.gz", "distribution_sha256": digest, "distribution_size": 1, "tree_sha256": digest, "binary_sha256": digest}},
			},
			code: "semantic-cross-field",
		},
		{
			name:     "planned repository identity",
			identity: "https://github.com/faustbrian/go-library-tools/schema/cohesion-engineering-identities-v1.schema.json",
			value: map[string]any{
				"schema_id": "urn:golib:cohesion:engineering-identities:v1", "schema_version": 1, "identity_count": 1,
				"identities": []any{map[string]any{"identity": "github.com/faustbrian/go-wrong", "repository": "github.com/faustbrian/go-right", "source_revision": commit, "path": ".", "kind": "repository", "status": "planned", "non_installable": true, "owner": "owner", "goal_path": "GOAL.md", "goal_sha256": digest}},
			},
			code: "semantic-cross-field",
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			data, err := json.Marshal(test.value)
			if err != nil {
				t.Fatal(err)
			}
			_, err = validateAndNormalizeSchemaV3(test.identity, data)
			failure, ok := err.(schemaV3Failure)
			if !ok || failure.Code != test.code {
				t.Fatalf("error = %#v, want schema-v3 code %q", err, test.code)
			}
		})
	}
}

func TestValidateAndNormalizeSchemaV3AllowsLargeOfficialOracleInputsOnlyAtTheSchemaField(t *testing.T) {
	t.Parallel()

	largeBase64 := strings.Repeat("AAAA", maximumTrustJSONString/4+1)
	value := map[string]any{
		"schema_id":             "urn:golib:cohesion:go-official-downloads-oracle:v1",
		"schema_version":        1,
		"official_index_sha256": "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"accepted_count":        1,
		"accepted": []any{map[string]any{
			"case_id": "accepted", "input_base64": largeBase64, "expected": []any{},
		}},
		"rejected_count": 1,
		"rejected": []any{map[string]any{
			"case_id": "rejected", "input_base64": "e30=", "error_code": "invalid",
		}},
	}
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := validateAndNormalizeSchemaV3(goOfficialDownloadsOracleV1SchemaIdentity, data); err != nil {
		t.Fatalf("validateAndNormalizeSchemaV3(large official input_base64) error = %v", err)
	}
}

func TestValidateAndNormalizeSchemaV3RejectsWrongGatePolicyDigest(t *testing.T) {
	t.Parallel()

	value := gatePolicyV1Fixture(t)
	value["policy_sha256"] = "sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
	assertSchemaV3Failure(t, "https://github.com/faustbrian/go-library-tools/schema/cohesion-gate-policy-v1.schema.json", value, "semantic-digest-mismatch", "/policy_sha256")
}

func TestValidateAndNormalizeSchemaV3AppliesResidualInventorySemantics(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		mutate  func(map[string]any)
		code    schemaV3ErrorCode
		pointer string
	}{
		{"item count", func(value map[string]any) { value["item_count"] = 3 }, "semantic-count-mismatch", "/item_count"},
		{"duplicate item", func(value map[string]any) {
			items := value["items"].([]any)
			items[1].(map[string]any)["inventory_item_id"] = items[0].(map[string]any)["inventory_item_id"]
		}, "semantic-duplicate", "/items/1/inventory_item_id"},
		{"item order", func(value map[string]any) {
			items := value["items"].([]any)
			items[0], items[1] = items[1], items[0]
		}, "semantic-canonical-order", "/items/1"},
		{"source release", func(value map[string]any) {
			value["sources"].([]any)[1].(map[string]any)["release"] = "v1.6.1"
		}, "semantic-release-identity", "/sources/1/release"},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			assertSchemaV3Failure(t, "https://github.com/faustbrian/go-library-tools/schema/cohesion-residual-inventory-v1.schema.json", mutateFixture(t, residualInventoryV1Fixture(), test.mutate), test.code, test.pointer)
		})
	}
}

func TestValidateAndNormalizeSchemaV3AppliesResidualRegisterSemantics(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		mutate  func(map[string]any)
		code    schemaV3ErrorCode
		pointer string
	}{
		{"partition overlap", func(value map[string]any) {
			value["resolutions"].([]any)[0].(map[string]any)["inventory_item_id"] = "item-01"
		}, "semantic-duplicate", "/resolutions/0/inventory_item_id"},
		{"resolution order", func(value map[string]any) {
			resolutions := value["resolutions"].([]any)
			resolutions[0], resolutions[1] = resolutions[1], resolutions[0]
		}, "semantic-canonical-order", "/resolutions/1"},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			assertSchemaV3Failure(t, "https://github.com/faustbrian/go-library-tools/schema/cohesion-residual-register-v1.schema.json", mutateFixture(t, residualRegisterV1Fixture(), test.mutate), test.code, test.pointer)
		})
	}
}

func TestSchemaProvenanceV1SemanticValidatorRejectsDecisionFreezeOrderAndReleaseDrift(t *testing.T) {
	t.Parallel()

	digest := "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	freeze := map[string]any{"release": "v1.6.0", "bytes_sha256": digest}
	entries := []any{
		map[string]any{"schema_path": "schema/a.schema.json", "decision_freeze": map[string]any{"release": "v1.6.0", "bytes_sha256": digest}, "forward_oracle": map[string]any{"release": "v1.6.0"}},
		map[string]any{"schema_path": "schema/b.schema.json", "decision_freeze": map[string]any{"release": "v1.6.0", "bytes_sha256": digest}, "resolved_reference_graph": map[string]any{"release": "v1.6.0"}},
	}
	base := map[string]any{"decision_freeze": freeze, "entry_count": 2, "entries": entries}
	tests := []struct {
		name    string
		mutate  func(map[string]any)
		code    schemaV3ErrorCode
		pointer string
	}{
		{"decision freeze", func(value map[string]any) {
			value["entries"].([]any)[0].(map[string]any)["decision_freeze"].(map[string]any)["bytes_sha256"] = "sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
		}, "semantic-external-identity", "/entries/0/decision_freeze"},
		{"order", func(value map[string]any) {
			rows := value["entries"].([]any)
			rows[0], rows[1] = rows[1], rows[0]
		}, "semantic-canonical-order", "/entries/1/schema_path"},
		{"release", func(value map[string]any) {
			value["entries"].([]any)[1].(map[string]any)["resolved_reference_graph"].(map[string]any)["release"] = "v1.6.1"
		}, "semantic-release-identity", "/entries/1/resolved_reference_graph/release"},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			value := mutateFixture(t, base, test.mutate)
			validate := schemaV3SemanticValidators["https://github.com/faustbrian/go-library-tools/schema/cohesion-schema-provenance-v1.schema.json"]
			if validate == nil {
				t.Fatal("schema provenance semantic validator is not registered")
			}
			failure, ok := selectSchemaV3Failure(validate(value))
			if !ok || failure.Code != test.code || failure.Pointer != test.pointer {
				t.Fatalf("failure = %#v, want %s at %s", failure, test.code, test.pointer)
			}
		})
	}
}

func TestValidateAndNormalizeSchemaV3AppliesBootstrapControlSemantics(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		identity string
		path     string
		mutate   func(map[string]any)
		code     schemaV3ErrorCode
	}{
		{
			name: "decision review duplicate reviewer", identity: schemaV3DecisionReviewV1SchemaIdentity,
			path: "../../release/cohesion-schema-v3-decision-review.json",
			mutate: func(value map[string]any) {
				reviews := value["reviews"].([]any)
				reviews[2].(map[string]any)["reviewer_id"] = reviews[0].(map[string]any)["reviewer_id"]
			},
			code: "semantic-duplicate",
		},
		{
			name: "decision freeze impossible timestamp", identity: schemaV3DecisionFreezeV1SchemaIdentity,
			path: "../../release/cohesion-schema-v3-decision-freeze.json",
			mutate: func(value map[string]any) {
				value["created_at"] = "2026-02-31T04:26:59Z"
			},
			code: "semantic-cross-field",
		},
		{
			name: "contract review commit mismatch", identity: contractReviewV1SchemaIdentity,
			path: "../../release/cohesion-contract-review.json",
			mutate: func(value map[string]any) {
				value["review"].(map[string]any)["reviewed_commit"] = "ffffffffffffffffffffffffffffffffffffffff"
			},
			code: "semantic-external-identity",
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			data, err := os.ReadFile(test.path)
			if err != nil {
				t.Fatal(err)
			}
			var value map[string]any
			if err := json.Unmarshal(data, &value); err != nil {
				t.Fatal(err)
			}
			test.mutate(value)
			data, err = json.Marshal(value)
			if err != nil {
				t.Fatal(err)
			}
			_, err = validateAndNormalizeSchemaV3(test.identity, data)
			failure, ok := err.(schemaV3Failure)
			if !ok || failure.Code != test.code {
				t.Fatalf("error = %#v, want schema-v3 code %q", err, test.code)
			}
		})
	}
}

func TestValidateAndNormalizeSchemaV3AppliesOfficialReviewSemantics(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("../../testdata/cohesion/go-official-downloads-oracle-review.json")
	if err != nil {
		t.Fatal(err)
	}
	var value map[string]any
	if err := json.Unmarshal(data, &value); err != nil {
		t.Fatal(err)
	}
	value["extractor"].(map[string]any)["source_revision"] = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	data, err = json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	_, err = validateAndNormalizeSchemaV3("https://github.com/faustbrian/go-library-tools/schema/cohesion-go-official-downloads-oracle-review-v1.schema.json", data)
	failure, ok := err.(schemaV3Failure)
	if !ok || failure.Code != "semantic-cross-field" || failure.Pointer != "/extractor/source_revision" {
		t.Fatalf("error = %#v, want semantic-cross-field at /extractor/source_revision", err)
	}
}

func TestValidateAndNormalizeSchemaV3AppliesAuthorizationRegistrySemantics(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		mutate  func(map[string]any)
		code    schemaV3ErrorCode
		pointer string
	}{
		{
			name: "entry count",
			mutate: func(value map[string]any) {
				value["entry_count"] = 3
			},
			code: "semantic-count-mismatch", pointer: "/entry_count",
		},
		{
			name: "duplicate authorization ID",
			mutate: func(value map[string]any) {
				entries := value["entries"].([]any)
				entries[1].(map[string]any)["authorization_id"] = entries[0].(map[string]any)["authorization_id"]
			},
			code: "semantic-duplicate", pointer: "/entries/1/authorization_id",
		},
		{
			name: "duplicate subject",
			mutate: func(value map[string]any) {
				entries := value["entries"].([]any)
				first := entries[0].(map[string]any)
				duplicate := cloneJSONMap(t, first)
				duplicate["authorization_id"] = "auth-decision-review-implementation-source-2"
				duplicate["record"].(map[string]any)["path"] = "release/cohesion-authorization-records/auth-decision-review-implementation-source-2.json"
				entries[1] = duplicate
			},
			code: "semantic-duplicate", pointer: "/entries/1/subject",
		},
		{
			name: "unsorted",
			mutate: func(value map[string]any) {
				entries := value["entries"].([]any)
				entries[0], entries[1] = entries[1], entries[0]
			},
			code: "semantic-canonical-order", pointer: "/entries/1",
		},
		{
			name: "unsafe record path",
			mutate: func(value map[string]any) {
				value["entries"].([]any)[0].(map[string]any)["record"].(map[string]any)["path"] = "../record.json"
			},
			code: "semantic-path", pointer: "/entries/0/record/path",
		},
		{
			name: "unsafe receipt path",
			mutate: func(value map[string]any) {
				value["entries"].([]any)[0].(map[string]any)["subject"].(map[string]any)["receipt_path"] = "../receipt.json"
			},
			code: "semantic-path", pointer: "/entries/0/subject/receipt_path",
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			value := authorizationRegistryV1Fixture(t)
			test.mutate(value)
			data, err := json.Marshal(value)
			if err != nil {
				t.Fatal(err)
			}
			_, err = validateAndNormalizeSchemaV3("https://github.com/faustbrian/go-library-tools/schema/cohesion-authorization-registry-v1.schema.json", data)
			failure, ok := err.(schemaV3Failure)
			if !ok || failure.Code != test.code || failure.Pointer != test.pointer {
				t.Fatalf("error = %#v, want %s at %s", err, test.code, test.pointer)
			}
		})
	}
}

func authorizationRegistryV1Fixture(t *testing.T) map[string]any {
	t.Helper()
	const fixture = `{"bootstrap_freeze_sha256":"sha256:0000000000000000000000000000000000000000000000000000000000000000","entries":[{"authorization_id":"auth-decision-review-implementation-source","authorization_kind":"decision-review","issuer_id":"cohesion-coordinator","record":{"bytes_sha256":"sha256:0000000000000000000000000000000000000000000000000000000000000000","path":"release/cohesion-authorization-records/auth-decision-review-implementation-source.json","repository":"github.com/faustbrian/go-library-tools","schema_id":"urn:golib:cohesion:authorization-record:v1","source_revision":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},"reviewer_id":"independent-reviewer","subject":{"applicable_input_manifest_sha256":"sha256:0000000000000000000000000000000000000000000000000000000000000000","dimension":"implementation","evidence_kind":"source-acceptance","goal_id":"golib-cohesion-v1","kind":"delivery","module":"github.com/faustbrian/go-oracle","outcome":"accepted","receipt_entry_id":"implementation-source","receipt_path":".verification/implementation-source.json","receipt_sha256":"sha256:0000000000000000000000000000000000000000000000000000000000000000","repository":"github.com/faustbrian/go-oracle","requirements_sha256":"sha256:6e5ac948c2cf1ec439fdf90c53c0648befb74a65c3eca2eb595737ee8211891c","source_revision":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}},{"authorization_id":"auth-release-authority-release-attestation","authorization_kind":"release-authority","issuer_id":"cohesion-coordinator","record":{"bytes_sha256":"sha256:0000000000000000000000000000000000000000000000000000000000000000","path":"release/cohesion-authorization-records/auth-release-authority-release-attestation.json","repository":"github.com/faustbrian/go-library-tools","schema_id":"urn:golib:cohesion:authorization-record:v1","source_revision":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},"release_verifier_id":"release-verifier","subject":{"applicable_input_manifest_sha256":"sha256:0000000000000000000000000000000000000000000000000000000000000000","dimension":"release","evidence_kind":"release-attestation","goal_id":"golib-cohesion-v1","kind":"delivery","module":"github.com/faustbrian/go-oracle","outcome":"accepted","receipt_entry_id":"release-attestation","receipt_path":".verification/release-attestation.json","receipt_sha256":"sha256:0000000000000000000000000000000000000000000000000000000000000000","repository":"github.com/faustbrian/go-oracle","requirements_sha256":"sha256:6e5ac948c2cf1ec439fdf90c53c0648befb74a65c3eca2eb595737ee8211891c","source_revision":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}}],"entry_count":2,"goal_id":"golib-cohesion-v1","predecessor_sha256":null,"registry_revision":1,"requirements_sha256":"sha256:6e5ac948c2cf1ec439fdf90c53c0648befb74a65c3eca2eb595737ee8211891c","schema_id":"urn:golib:cohesion:authorization-registry:v1","schema_version":1}`
	var value map[string]any
	if err := json.Unmarshal([]byte(fixture), &value); err != nil {
		t.Fatal(err)
	}
	return value
}

func assertSchemaV3Failure(t *testing.T, identity string, value map[string]any, code schemaV3ErrorCode, pointer string) {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	_, err = validateAndNormalizeSchemaV3(identity, data)
	failure, ok := err.(schemaV3Failure)
	if !ok || failure.Code != code || failure.Pointer != pointer {
		t.Fatalf("error = %#v, want %s at %s", err, code, pointer)
	}
}

func mutateFixture(t *testing.T, value map[string]any, mutate func(map[string]any)) map[string]any {
	t.Helper()
	value = cloneJSONMap(t, value)
	mutate(value)
	return value
}

func gatePolicyV1Fixture(t *testing.T) map[string]any {
	t.Helper()
	digest := "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	value := map[string]any{
		"rule_id": "golib-gate-policy-v1", "schema_id": "urn:golib:cohesion:gate-policy:v1", "schema_version": 1,
		"sources": []any{
			map[string]any{"kind": "embedded-defaults", "locator": "embedded://golib-gate-policy-v1", "present": true, "release": "v1.6.0", "tag_object_sha": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "peeled_commit": "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", "checksums_sha256": digest, "semantic_sha256": digest},
			map[string]any{"kind": "repository-config", "path": ".golib.yaml", "present": true, "bytes_sha256": digest},
			map[string]any{"kind": "module-manifest", "path": "modules.json", "present": false},
			map[string]any{"kind": "package-manifest", "path": "packages.json", "present": false},
		},
		"config": map[string]any{
			"schema_version": 1, "tool_version": "v1.6.0",
			"manifest": map[string]any{"modules": "modules.json", "packages": "packages.json"},
			"evidence": map[string]any{"root": ".verification"}, "mutation": map[string]any{"root": ".verification/mutation"},
			"runtimes": map[string]any{},
		},
		"modules": nil, "packages": nil, "policy_sha256": digest,
	}
	policyDigest, err := semanticObjectSHA256(value, "policy_sha256", maximumSchemaV3ArtifactBytes)
	if err != nil {
		t.Fatal(err)
	}
	value["policy_sha256"] = policyDigest
	return value
}

func residualInventoryV1Fixture() map[string]any {
	digest := "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	assets := []string{
		"cohesion-phase1-exception-register.md",
		"cohesion-phase1-inventory-cohort-a.md", "cohesion-phase1-inventory-cohort-b.md",
		"cohesion-phase1-inventory-cohort-c.md", "cohesion-phase1-inventory-cohort-d.md",
		"cohesion-phase1-inventory-cohort-e.md", "cohesion-phase1-inventory-cohort-f.md",
	}
	sources := make([]any, 0, len(assets))
	for _, asset := range assets {
		sources = append(sources, map[string]any{
			"repository": "github.com/faustbrian/go-library-tools", "release": "v1.6.0",
			"release_url": "https://github.com/faustbrian/go-library-tools/releases/download/v1.6.0/" + asset,
			"asset":       asset, "bytes_sha256": digest,
		})
	}
	item := func(id, asset, classification string, start int) map[string]any {
		return map[string]any{
			"inventory_item_id": id, "repository": "github.com/faustbrian/go-oracle", "module": "github.com/faustbrian/go-oracle",
			"difference": "difference " + id, "classification": classification,
			"source": map[string]any{"asset": asset, "bytes_sha256": digest, "section": "# Inventory", "start_byte": start, "end_byte": start + 1, "item_sha256": digest},
		}
	}
	return map[string]any{
		"schema_id": "urn:golib:cohesion:residual-inventory:v1", "schema_version": 1, "goal_id": "golib-cohesion-v1",
		"requirements_sha256": "sha256:6e5ac948c2cf1ec439fdf90c53c0648befb74a65c3eca2eb595737ee8211891c",
		"source_count":        7, "sources": sources, "item_count": 2,
		"items": []any{item("item-01", assets[0], "justified-domain-specific", 0), item("item-02", assets[1], "intentional-compatibility", 1)},
	}
}

func residualRegisterV1Fixture() map[string]any {
	digest := "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	locator := func() map[string]any {
		return map[string]any{"repository": "github.com/faustbrian/go-oracle", "path": "evidence.json", "source_revision": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "bytes_sha256": digest}
	}
	exception := func(id, item, disposition string) map[string]any {
		value := map[string]any{
			"exception_id": id, "inventory_item_id": item, "repository": "github.com/faustbrian/go-oracle", "module": "github.com/faustbrian/go-oracle",
			"difference": "difference " + item, "classification": "justified-domain-specific", "evidence": []any{locator()}, "owner": "cohesion-coordinator",
			"review_decision": map[string]any{"reviewer_id": "reviewer", "outcome": "accepted", "record": locator()}, "disposition": disposition,
		}
		if disposition == "temporary" {
			value["removal_condition"] = "implement"
		} else {
			value["permanence_reason"] = "domain-specific"
		}
		return value
	}
	return map[string]any{
		"schema_id": "urn:golib:cohesion:residual-register:v1", "schema_version": 1, "goal_id": "golib-cohesion-v1",
		"requirements_sha256": "sha256:6e5ac948c2cf1ec439fdf90c53c0648befb74a65c3eca2eb595737ee8211891c",
		"inventory":           map[string]any{"repository": "github.com/faustbrian/go-library-tools", "release": "v1.6.0", "release_url": "https://github.com/faustbrian/go-library-tools/releases/download/v1.6.0/cohesion-residual-inventory.json", "asset": "cohesion-residual-inventory.json", "bytes_sha256": digest},
		"release_entry_count": 0, "release_entries": []any{},
		"exception_entry_count": 2, "exceptions": []any{exception("exception-1", "item-01", "temporary"), exception("exception-2", "item-02", "permanent")},
		"resolution_count": 2, "resolutions": []any{
			map[string]any{"inventory_item_id": "item-03", "state": "resolved", "evidence": locator()},
			map[string]any{"inventory_item_id": "item-04", "state": "superseded", "evidence": locator(), "successor_item_id": "item-99"},
		},
	}
}

func modulesV3StateCase(t *testing.T, fixture []byte, implementation, hardening, release DimensionState, status GoalStatus) []byte {
	t.Helper()
	var document map[string]any
	if err := json.Unmarshal(fixture, &document); err != nil {
		t.Fatal(err)
	}
	delivery := manifestV3Delivery(manifestV3Modules(document)[0])
	delivery["implementation"] = string(implementation)
	delivery["hardening"] = string(hardening)
	delivery["release"] = string(release)
	delivery["goal"].(map[string]any)["status"] = string(status)
	evidence := delivery["evidence"].(map[string]any)
	for _, dimension := range []struct {
		name  string
		state DimensionState
	}{{"implementation", implementation}, {"hardening", hardening}, {"release", release}} {
		if dimension.state == DimensionNotStarted || dimension.state == DimensionInProgress {
			evidence[dimension.name] = []any{}
			continue
		}
		evidence[dimension.name] = []any{map[string]any{
			"receipt_path":   fmt.Sprintf("evidence/%s.json", dimension.name),
			"receipt_sha256": "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			"entry_id":       dimension.name,
		}}
	}
	data, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func assertContainsSchemaV3Failure(t *testing.T, failures []schemaV3Failure, code schemaV3ErrorCode, pointer string) {
	t.Helper()
	for _, failure := range failures {
		if failure.Code == code && failure.Pointer == pointer {
			return
		}
	}
	t.Fatalf("failures = %#v, want %s at %s", failures, code, pointer)
}
