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
