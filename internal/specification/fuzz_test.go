package specification

import "testing"

func FuzzSourceManifest(f *testing.F) {
	f.Add("manifest.tsv", "id\tversion\turl\tsha256\tstatus\n")
	f.Add("manifest.json", `{}`)
	f.Fuzz(func(_ *testing.T, path, content string) {
		_ = validateSourceManifest(path, []byte(content))
	})
}

func FuzzDecisionManifests(f *testing.F) {
	f.Add(`{"schema_version":1,"change_control":{},"decisions":[]}`, `{"schema_version":1,"decisions":[]}`)
	f.Fuzz(func(_ *testing.T, decisions, conformance string) {
		_, _ = loadDecisions([]byte(decisions))
		_, _ = loadConformance([]byte(conformance))
	})
}
