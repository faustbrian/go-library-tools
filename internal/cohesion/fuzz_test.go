package cohesion

import "testing"

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
	f.Add([]byte(`{`))
	f.Fuzz(func(_ *testing.T, data []byte) {
		if len(data) > 1<<20 {
			return
		}
		_ = validateInputsSchema(data)
		_ = validateCatalogSchema(data)
	})
}
