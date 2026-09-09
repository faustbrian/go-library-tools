package cohesion

import (
	"encoding/json"
	"os"
	"testing"
)

func FuzzRequiredMetadataDiagnosticsNeverPanics(f *testing.F) {
	f.Add([]byte(`{"schema_version":2,"modules":[]}`))
	f.Add([]byte(`{"schema_version":2,"modules":[{"cohesion":{}}]}`))
	f.Add([]byte(`{"schema_version":1,"modules":[]}`))
	f.Add([]byte(`{`))
	f.Fuzz(func(_ *testing.T, data []byte) {
		_ = requiredMetadataDiagnosticsData(data)
	})
}

func FuzzAggregationSchemasNeverPanic(f *testing.F) {
	f.Add([]byte(`{"schema_version":1,"design_language":{"version":"1.0","sha256":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},"repositories":[]}`))
	f.Add([]byte(`{"schema_version":1,"view":"engineering","scope":"ecosystem","repository":null,"design_language":{"version":"1.0","sha256":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","source_identity":"unpublished"},"tooling":{"version":"dev","publication_status":"unpublished"},"modules":[]}`))
	f.Add([]byte(`{"schema_version":1,"repository_count":1,"repositories":[{"repository":"github.com/faustbrian/go-library-tools","source":{"kind":"release-source"}}]}`))
	f.Add([]byte(`{`))
	f.Fuzz(func(_ *testing.T, data []byte) {
		if len(data) > 1<<20 {
			return
		}
		_ = validateInputsSchema(data)
		_ = validateCatalogSchema(data)
		if validateSourcesSchema(data) == nil {
			var lock SourceLock
			if json.Unmarshal(data, &lock) == nil {
				_ = validateSourceLockOrder(lock)
			}
		}
	})
}

func FuzzSchemaV3TrustJSONNeverPanics(f *testing.F) {
	read := func(name string) []byte {
		f.Helper()
		data, err := os.ReadFile("../../release/" + name)
		if err != nil {
			f.Fatal(err)
		}
		return data
	}
	contractReview := read("cohesion-contract-review.json")
	contractFreeze := read("cohesion-contract-freeze.json")
	decisionReview := read("cohesion-schema-v3-decision-review.json")
	decisionFreeze := read("cohesion-schema-v3-decision-freeze.json")
	manifestV2, err := os.ReadFile("../../modules.json")
	if err != nil {
		f.Fatal(err)
	}
	manifestV3, err := makeManifestV3Fixture(manifestV2, "sha256:6e5ac948c2cf1ec439fdf90c53c0648befb74a65c3eca2eb595737ee8211891c")
	if err != nil {
		f.Fatal(err)
	}
	f.Add(decisionReview)
	f.Add(manifestV3)
	f.Add([]byte(`{"goal":{"id":"golib-cohesion-v1","requirements_sha256":"sha256:6e5ac948c2cf1ec439fdf90c53c0648befb74a65c3eca2eb595737ee8211891c","status":"in-progress"},"implementation":"in-progress","hardening":"in-progress","release":"in-progress","evidence":{"implementation":[],"hardening":[],"release":[]}}`))
	f.Add([]byte(`{"schema_version":3,"value":1}`))
	f.Add([]byte(`{"duplicate":1,"duplicate":2}`))
	f.Add([]byte(`{"nested":[{"value":"\ud83d\ude00"}]}`))
	f.Add([]byte{0xef, 0xbb, 0xbf, '{', '}'})
	f.Fuzz(func(_ *testing.T, data []byte) {
		if len(data) > 1<<20 {
			return
		}
		var decoded any
		_ = decodeTrustJSON(data, 1<<20, &decoded)
		_, _ = canonicalizeTrustJSON(data, 1<<20)
		_ = validateSchemaV3DecisionReview(data)
		_ = validateManifestDeliveryJSON(data, "sha256:6e5ac948c2cf1ec439fdf90c53c0648befb74a65c3eca2eb595737ee8211891c")
		_ = validateManifestV3(data, "sha256:6e5ac948c2cf1ec439fdf90c53c0648befb74a65c3eca2eb595737ee8211891c")
		_ = validateSchemaV3BootstrapControlSemantics(data, contractFreeze, decisionReview, decisionFreeze)
		_ = validateSchemaV3BootstrapControlSemantics(contractReview, data, decisionReview, decisionFreeze)
		_ = validateSchemaV3BootstrapControlSemantics(contractReview, contractFreeze, data, decisionFreeze)
		_ = validateSchemaV3BootstrapControlSemantics(contractReview, contractFreeze, decisionReview, data)
	})
}
