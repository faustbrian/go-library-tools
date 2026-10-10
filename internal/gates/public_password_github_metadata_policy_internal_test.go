package gates

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

// Public decision records recomputed at Password 481f9392 and 828d9241,
// and GitHub descriptions-next 417c4fb (artifact SHA256 9d85f3a842c02157).
// Fragment literals so regression data does not resemble new credentials.
func TestPublicPasswordAndGitHubMetadataPinnedScanner(t *testing.T) {
	if os.Getenv("GOLIB_GITLEAKS_INTEGRATION") != "1" {
		t.Skip("opt-in pinned scanner contract")
	}
	// Deliberately synthetic high-entropy control; the repeated eight-character
	// fixture used elsewhere falls below this detector's entropy threshold.
	credential := strings.Join([]string{"A1b2C3d4", "E5f6G7h8", "I9j0K1l2", "M3n4O5p6", "Q7r8S9t0", "U1v2W3x4", "Y5z6A7b8", "C9d0E1f2"}, "")
	password := []string{
		strings.Join([]string{"  PASSWORD-DEC-0", "06 sha256:2aa164", "80aefc694ed42f78", "109e6347c3fd6610", "5fee868362aef58c", "087368bf83;"}, ""),
		strings.Join([]string{"  PASSWORD-DEC-0", "07 sha256:23b884", "6b5751879f08ba17", "02f0fa3625b5d5cb", "b6e9d9fe46536f4b", "bb25270d42."}, ""),
		strings.Join([]string{"  PASSWORD-DEC-0", "01 sha256:81d693", "f05f84c603efdb62", "9d9249e13e134def", "4e645afffd12e50b", "1e7233823c;"}, ""),
		strings.Join([]string{"  PASSWORD-DEC-0", "02 sha256:d0aff4", "5bb407811912bc21", "be738605732adbd6", "04decc339b449bd2", "5fdf125ddf;"}, ""),
		strings.Join([]string{"  PASSWORD-DEC-0", "03 sha256:abb7d2", "b09fe08461078fbe", "be88637d317ef52b", "5ffc035fa3072950", "717978f24f;"}, ""),
		strings.Join([]string{"  PASSWORD-DEC-0", "04 sha256:35d8b7", "5df3f64f7a8d45e3", "8c763ebc7e110953", "b0542d3c03cd230c", "2d1194e83e;"}, ""),
		strings.Join([]string{"  PASSWORD-DEC-0", "05 sha256:ed3cb3", "293b84a79a7cd068", "92599874eaf56b4f", "91071e8537df71fc", "93794a7104;"}, ""),
		strings.Join([]string{"  PASSWORD-DEC-0", "06 sha256:d02071", "7a693732de00655a", "0a0c4359626feacc", "6f71d130acbd2180", "0e3979a748;"}, ""),
		strings.Join([]string{"  PASSWORD-DEC-0", "07 sha256:e3dcde", "e13b077e6effd51a", "4b6dd9d00c0b3a37", "3826aaa0ee305495", "e4088a5def."}, ""),
		strings.Join([]string{"  PASSWORD-DEC-0", "01 sha256:ba44cc", "58be46506940728e", "9d6edfcfc165f84c", "201ac01efe709b49", "49a211cd92;"}, ""),
		strings.Join([]string{"  PASSWORD-DEC-0", "02 sha256:9d9202", "7d805ca587dcce98", "75bb0af09d2d9913", "ad883bee56e17340", "7d7a6b6362;"}, ""),
		strings.Join([]string{"  PASSWORD-DEC-0", "03 sha256:ea20f1", "c77cbbb8f8b5ca43", "a23c217ae77e142e", "ed42ee4db1f636be", "1fc949ab45;"}, ""),
		strings.Join([]string{"  PASSWORD-DEC-0", "04 sha256:3526e9", "549c855d048fec5c", "31220d1c3085331e", "ef128437c1d3bc1a", "1e682daf9d;"}, ""),
		strings.Join([]string{"  PASSWORD-DEC-0", "05 sha256:695d18", "94768e03f3a3d599", "4e432d9eac0248b9", "81bc3493c0931ce9", "ba19996d84;"}, ""),
	}
	examples := []string{
		strings.Join([]string{"              \"t", "emp_clone_token\"", ": \"ABTLWHOULUVAX", "GTRYU7OC2876QJ2O", "\","}, ""),
		strings.Join([]string{"                ", "\"temp_clone_toke", "n\": \"ABTLWHOULUV", "AXGTRYU7OC2876QJ", "2O\","}, ""),
		strings.Join([]string{"            \"tem", "p_clone_token\": ", "\"ABTLWHOULUVAXGT", "RYU7OC2876QJ2O\","}, ""),
		strings.Join([]string{"          \"temp_", "clone_token\": \"A", "BTLWHOULUVAXGTRY", "U7OC2876QJ2O\","}, ""),
	}
	for _, fixture := range []struct{ name, path, content string }{
		{"password decisions", "CHANGELOG.md", strings.Join(password, "\n")},
		{"public tool version", "tools/versions.env", "APIDIFF_VERSION=" + strings.Join([]string{"v0.0.0", "20260718201538", "764159d718ef"}, "-")},
		{"upstream GitHub examples", "specification/independent/github-rest-api/api.github.com.2022-11-28.json", strings.Join(examples, "\n")},
	} {
		t.Run(fixture.name, func(t *testing.T) {
			changed := regexp.MustCompile(`[A-Za-z0-9_-]{20,}`).ReplaceAllStringFunc(fixture.content, func(value string) string {
				replacement := byte('a')
				if value[len(value)-1] == replacement {
					replacement = 'b'
				}
				return value[:len(value)-1] + string(replacement)
			})
			if changed == fixture.content {
				t.Fatal("counterfactual did not change public metadata")
			}
			for _, scenario := range []struct {
				name, path, content string
				wantFindings        bool
			}{
				{"exact public records", fixture.path, fixture.content, false},
				{"different path", "other/" + fixture.path, fixture.content, true},
				{"changed bytes", fixture.path, changed, true},
				{"credential control", fixture.path, "api_key=" + credential, true},
				{"adjacent credential", fixture.path, fixture.content + "\napi_key=" + credential, true},
				{"same line credential", fixture.path, fixture.content + " api_key=" + credential, true},
			} {
				t.Run(scenario.name, func(t *testing.T) {
					root, policy := gitleaksRepository(t, map[string]string{scenario.path: scenario.content + "\n"})
					for _, mode := range []string{"git", "dir"} {
						findings, _, err := runGitleaksIntegration(t, root, policy, mode)
						if scenario.wantFindings {
							if err == nil || len(findings) == 0 {
								t.Fatalf("%s lost non-allowlisted findings", mode)
							}
						} else if err != nil || len(findings) != 0 {
							t.Fatalf("%s rejected immutable public records (findings=%d)", mode, len(findings))
						}
					}
				})
			}
		})
	}
}
