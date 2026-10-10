package gates

import (
	"os"
	"strings"
	"testing"
)

// Exact immutable public artifacts from APIQuery ae89ed7 and Capability
// 27f2eec, not application credentials. Keep literals fragmented so this
// scanner regression fixture does not resemble a newly supplied credential.
func publicConsumerMetadataFixtures() []struct{ path, rule, line string } {
	return []struct{ path, rule, line string }{
		{".verification/cohesion/api-query-v1-oracle/verify.go", "generic-api-key", strings.Join([]string{"\tvalidateRelease", "Artifact(ctx, ro", "ot, releaseCommi", "t, packet.Releas", "e.APIBaseline, \"", "ec28983aeaa00eed", "bfa61d2a6c5eeb67", "fc6c3185\", 25021", ", \"b8eac53159059", "9de50da2ed042591", "db7edf07c6a438e9", "00085aca53ed6354", "fc5\")"}, "")},
		{"testdata/v1-hmac.token", "jwt", strings.Join([]string{"cap1.eyJ2IjoxLCJ", "0eXAiOiJjYXBhYml", "saXR5IiwiYWxnIjo", "iaG1hYy1zaGEyNTY", "iLCJraWQiOiJpbnR", "lcm9wIn0.eyJ2Ijo", "xLCJpc3MiOiJpbnR", "lcm9wIiwiYXVkIjp", "bInNlcnZpY2UiXSw", "iYmVhcmVyIjp0cnV", "lLCJyZXNvdXJjZSI", "6Im9iamVjdHMvNDI", "iLCJvcGVyYXRpb24", "iOiJyZWFkIiwiaWF", "0IjoxNzg2Mjc2ODA", "wLCJuYmYiOjE3ODY", "yNzY4MDAsImV4cCI", "6MTc4NjI3Njg2MCw", "iaWQiOiJpbnRlcm9", "wLWNhcGFiaWxpdHk", "ifQ.Iwj0h9OnGkC0", "P1Hrw2L9dweW8Wvs", "C0md1z65DmNlXTQ"}, "")},
	}
}

func TestPublicConsumerMetadataPinnedScanner(t *testing.T) {
	if os.Getenv("GOLIB_GITLEAKS_INTEGRATION") != "1" {
		t.Skip("opt-in pinned scanner contract")
	}
	for _, fixture := range publicConsumerMetadataFixtures() {
		t.Run(fixture.rule, func(t *testing.T) {
			for _, scenario := range []struct {
				name, path, line string
				wantFindings     bool
			}{
				{"exact public artifact", fixture.path, fixture.line, false},
				{"different path", "other/" + fixture.path, fixture.line, true},
				{"changed bytes", fixture.path, strings.Replace(fixture.line, fixture.line[len(fixture.line)-8:], strings.Repeat("A", 8), 1), true},
				{"adjacent credential", fixture.path, fixture.line + " api_key=" + strings.Repeat("aB3cD4eF", 8), true},
			} {
				t.Run(scenario.name, func(t *testing.T) {
					root, policy := gitleaksRepository(t, map[string]string{scenario.path: scenario.line + "\n"})
					for _, mode := range []string{"git", "dir"} {
						findings, _, err := runGitleaksIntegration(t, root, policy, mode)
						if scenario.wantFindings {
							if err == nil || len(findings) == 0 {
								t.Fatalf("%s did not retain non-allowlisted findings", mode)
							}
						} else if err != nil || len(findings) != 0 {
							t.Fatalf("%s rejected exact public artifact (findings=%d)", mode, len(findings))
						}
					}
				})
			}
		})
	}
}
