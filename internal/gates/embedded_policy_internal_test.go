package gates

import (
	"os"
	"strings"
	"testing"
)

// The owned policy is an input to both the scanner and its source scan. Moving
// its representation must not need a whole-file exemption or hide a credential
// next to the policy data.
func TestEmbeddedGitleaksPolicyPinnedScanner(t *testing.T) {
	if os.Getenv("GOLIB_GITLEAKS_INTEGRATION") != "1" {
		t.Skip("opt-in pinned scanner contract")
	}
	// Gitleaks' inherited default configuration ignores paths containing
	// gitleaks.toml. The data asset must remain an actually scanned source.
	const path = "internal/gates/policy/secrets.toml"
	credential := strings.Join([]string{"A1b2C3d4", "E5f6G7h8", "I9j0K1l2", "M3n4O5p6", "Q7r8S9t0", "U1v2W3x4", "Y5z6A7b8", "C9d0E1f2"}, "")
	for _, scenario := range []struct {
		name, content string
		findings      bool
	}{
		{"owned policy", gitleaksPolicy, false},
		{"adjacent credential", gitleaksPolicy + "\napi_key=" + credential + "\n", true},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			root, policy := gitleaksRepository(t, map[string]string{path: scenario.content})
			for _, mode := range []string{"git", "dir"} {
				findings, _, err := runGitleaksIntegration(t, root, policy, mode)
				if !scenario.findings {
					if err != nil || len(findings) != 0 {
						for _, finding := range findings {
							parts := strings.Split(finding.Fingerprint, ":")
							t.Logf("safe finding identity: %s %s line=%s", finding.RuleID, finding.File, parts[len(parts)-1])
						}
						t.Fatal("owned policy source is not scanner-clean")
					}
					continue
				}
				if err == nil || len(findings) != 1 || findings[0].RuleID != "generic-api-key" || findings[0].File != path {
					for _, finding := range findings {
						t.Logf("safe finding identity: %s %s", finding.RuleID, finding.File)
					}
					t.Logf("mode=%s scanner_failed=%t findings=%d", mode, err != nil, len(findings))
					t.Fatal("adjacent credential finding was lost or policy source introduced findings")
				}
			}
		})
	}
}
