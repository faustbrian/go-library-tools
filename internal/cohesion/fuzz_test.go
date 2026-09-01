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
