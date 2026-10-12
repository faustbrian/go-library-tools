package specification

import "testing"

// FuzzSourceManifest varies format-name/content input without an explicit
// harness cap, exercising declarative parsing without downloads or source
// refresh.
func FuzzSourceManifest(f *testing.F) {
	f.Add("manifest.tsv", "id\tversion\turl\tsha256\tstatus\n")
	f.Add("manifest.json", `{}`)
	f.Fuzz(func(_ *testing.T, path, content string) {
		_ = validateSourceManifest(path, []byte(content))
	})
}

// FuzzDecisionManifests varies decision/conformance JSON without an explicit
// harness cap; parsing cannot execute evidence or certify its outcome.
func FuzzDecisionManifests(f *testing.F) {
	f.Add(`{"schema_version":1,"change_control":{},"decisions":[]}`, `{"schema_version":1,"decisions":[]}`)
	f.Fuzz(func(_ *testing.T, decisions, conformance string) {
		_, _ = loadDecisions([]byte(decisions))
		_, _ = loadConformance([]byte(conformance))
	})
}
