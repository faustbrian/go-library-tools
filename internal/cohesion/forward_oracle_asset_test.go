package cohesion

import (
	"os"
	"testing"
)

func TestDiagnosticForwardOracleAssetMatchesFrozenMatrix(t *testing.T) {
	data, err := os.ReadFile("../../testdata/cohesion/forward-oracles/cohesion-diagnostic-v1-forward-oracle.json")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"base.canonical-minimum", "base.canonical-rich",
		"diagnostic.invalid.code", "diagnostic.invalid.empty",
		"diagnostic.invalid.empty-causes", "diagnostic.invalid.message",
		"diagnostic.invalid.path", "diagnostic.invalid.unknown-cause",
	}
	if err := verifyForwardOracleV2(data, forwardOracleV2Expectation{
		SchemaIdentity: diagnosticV3SchemaIdentity,
		CaseCount:      len(want),
		MaximumBytes:   1 << 20,
		CaseIDs:        want,
		OraclePath:     "testdata/cohesion/forward-oracles/cohesion-diagnostic-v1-forward-oracle.json",
	}); err != nil {
		t.Fatal(err)
	}
}

func TestContractReviewForwardOracleAssetMatchesFrozenMatrix(t *testing.T) {
	data, err := os.ReadFile("../../testdata/cohesion/forward-oracles/cohesion-contract-review-v1-forward-oracle.json")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"base.canonical-minimum", "base.canonical-rich", "contract-review.invalid.accepted-findings", "contract-review.invalid.contract", "contract-review.invalid.goal", "contract-review.invalid.rejected-findings", "contract-review.invalid.release-url", "contract-review.invalid.reviewed-commit", "contract-review.invalid.workflow-id", "contract-review.invalid.workflow-outcome", "contract-review.valid.accepted", "contract-review.valid.rejected", "json.bom", "json.depth-65", "json.duplicate-key", "json.exponent", "json.fraction", "json.integer-overflow", "json.invalid-utf8", "json.lone-surrogate", "json.negative-zero", "json.noncanonical", "json.noncharacter", "json.trailing-value", "schema.missing-required", "schema.unknown-member", "schema.wrong-type"}
	if err := verifyForwardOracleV2(data, forwardOracleV2Expectation{SchemaIdentity: contractReviewV1SchemaIdentity, CaseCount: len(want), MaximumBytes: 1 << 20, CaseIDs: want, OraclePath: "testdata/cohesion/forward-oracles/cohesion-contract-review-v1-forward-oracle.json"}); err != nil {
		t.Fatal(err)
	}
}

func TestContractFreezeForwardOracleAssetMatchesFrozenMatrix(t *testing.T) {
	data, err := os.ReadFile("../../testdata/cohesion/forward-oracles/cohesion-contract-freeze-v1-forward-oracle.json")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"base.canonical-minimum", "base.canonical-rich", "contract-freeze.invalid.contract", "contract-freeze.invalid.goal", "contract-freeze.invalid.missing-review", "contract-freeze.invalid.review-path", "contract-freeze.invalid.schema", "contract-freeze.invalid.timestamp", "contract-freeze.invalid.unknown-member", "contract-freeze.valid"}
	if err := verifyForwardOracleV2(data, forwardOracleV2Expectation{SchemaIdentity: contractFreezeV1SchemaIdentity, CaseCount: len(want), MaximumBytes: 1 << 20, CaseIDs: want, OraclePath: "testdata/cohesion/forward-oracles/cohesion-contract-freeze-v1-forward-oracle.json"}); err != nil {
		t.Fatal(err)
	}
}

func TestCatalogV2ForwardOracleAssetMatchesFrozenMatrix(t *testing.T) {
	data, err := os.ReadFile("../../testdata/cohesion/forward-oracles/cohesion-catalog-v2-forward-oracle.json")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"base.canonical-minimum", "base.canonical-rich", "catalog-v2.invalid.final-input-manifest-version", "catalog-v2.invalid.missing-modules", "catalog-v2.invalid.unknown-member"}
	if err := verifyForwardOracleV2(data, forwardOracleV2Expectation{SchemaIdentity: "https://github.com/faustbrian/go-library-tools/schema/cohesion-catalog-v2.schema.json", CaseCount: len(want), MaximumBytes: 4 << 20, CaseIDs: want, OraclePath: "testdata/cohesion/forward-oracles/cohesion-catalog-v2-forward-oracle.json"}); err != nil {
		t.Fatal(err)
	}
}
