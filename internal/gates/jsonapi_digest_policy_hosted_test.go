package gates

import (
	"os"
	"strings"
	"testing"
)

// These public decision hashes were independently recomputed from the two
// immutable JSONAPI decision registers identified in docs/security.md.
func TestJSONAPIDecisionDigestPolicyHosted(t *testing.T) {
	if os.Getenv("GITHUB_ACTIONS") != "true" {
		t.Skip("decision-digest scanner verification runs only in hosted CI")
	}
	if os.Getenv("GOLIB_GITLEAKS_INTEGRATION") != "1" {
		t.Skip("set GOLIB_GITLEAKS_INTEGRATION=1 to run the pinned scanner contract")
	}
	records := []struct {
		id     string
		digest []string
	}{
		{"JSONAPI-DEC-011", []string{"d234cd9a", "4e40ce09", "214a023d", "09c11af6", "e7077c1b", "a5a39f8c", "01a9dc6c", "bbb98095"}},
		{"JSONAPI-DEC-012", []string{"85161bb4", "7428b3ae", "ad1e8930", "b318ccd9", "a8146800", "1e190378", "cdb7b621", "eb3022ef"}},
		{"JSONAPI-DEC-013", []string{"9b156b1a", "af44e12e", "b6f10ee7", "4a737c98", "1aec3f89", "8b94fad1", "f547ac2e", "68c57e04"}},
		{"JSONAPI-DEC-001", []string{"7ab9e941", "c1eed52a", "bff8d37e", "d674e466", "d75682bf", "c9dc3427", "24208969", "2c7fadd4"}},
		{"JSONAPI-DEC-002", []string{"92b24e46", "b91b1722", "39d6c7ef", "6afeab30", "28f6d3d8", "83f6c66e", "641b9515", "74811cb3"}},
		{"JSONAPI-DEC-003", []string{"b894d768", "df3cdfa9", "6914f4b8", "805c3dfd", "fe3bf398", "8b7281e4", "8f73113c", "7abff5c1"}},
		{"JSONAPI-DEC-004", []string{"eefa54a5", "8ed63f71", "bbaa3b6d", "2d8be4ca", "bcaef92e", "e05d1d9c", "7e7014ae", "bde062fd"}},
		{"JSONAPI-DEC-005", []string{"8c5339c0", "c40d1ca9", "ecbc7db8", "fe8ecb13", "3d597b6d", "fa4d535f", "9140e7a8", "03751050"}},
		{"JSONAPI-DEC-006", []string{"91709eb6", "d8259986", "3d8e0375", "9d1cd9f7", "268d26e8", "cfb3b0dd", "652b7e27", "1873cb97"}},
		{"JSONAPI-DEC-007", []string{"17f841ab", "f7b95111", "69d91641", "2f03e7b5", "13bfde2b", "93b76422", "66a23d92", "e5ee99f0"}},
		{"JSONAPI-DEC-008", []string{"383a65c3", "c89cae68", "d2d61499", "3caab29d", "a026c9a6", "a2c057e5", "70b6bbf6", "52c9a7db"}},
		{"JSONAPI-DEC-009", []string{"ee353253", "3128439a", "7617e7da", "04fbbda1", "3f30b14b", "7eb30119", "0fd5712a", "1d2c0625"}},
		{"JSONAPI-DEC-010", []string{"6721f9d3", "ea09062d", "9665828d", "3e1f2127", "d3bd3d1d", "967d6214", "e01a164b", "c7b8b1f2"}},
		{"JSONAPI-DEC-001", []string{"1c992612", "f6fdf57e", "58587537", "47bcb574", "64f149b5", "17084813", "681e7029", "f31186f0"}},
		{"JSONAPI-DEC-002", []string{"7d6658ae", "3e8b8176", "dffdb809", "96a5439f", "55b30d8d", "3c1e5a2a", "3d1c9f28", "5f2d2338"}},
		{"JSONAPI-DEC-003", []string{"86ae12b7", "a1ba561c", "f2f69475", "6d980ed6", "eec544b3", "007f4705", "8e87f8c5", "1023115e"}},
		{"JSONAPI-DEC-004", []string{"be285995", "51762934", "5b407e0b", "93744ada", "5cc85d54", "b72092b3", "8310d9ff", "16b795f5"}},
		{"JSONAPI-DEC-005", []string{"db4077bb", "84f3aadc", "3f3b68c3", "07029f01", "c2f49ae4", "5a8f17d3", "4e0d4945", "a03f31d3"}},
		{"JSONAPI-DEC-006", []string{"526de3bc", "95e292ca", "970e560b", "d06783a9", "ce70a8ef", "656dfd8e", "7ef51151", "a00339c5"}},
		{"JSONAPI-DEC-007", []string{"ad587c6d", "3623a796", "4b3047f7", "42d2b9c1", "91b9b979", "98417185", "7c727709", "2d110501"}},
		{"JSONAPI-DEC-008", []string{"98b181e4", "0f339a4a", "11b389a6", "bfe2ed0c", "a28ac122", "5fabfdaa", "d62f5430", "bb5363b7"}},
		{"JSONAPI-DEC-009", []string{"47736a06", "49bf48e1", "9fed8cc7", "5bb8308a", "c1790699", "487a4e40", "55a8e2a4", "d486ec2b"}},
		{"JSONAPI-DEC-010", []string{"c7a9604e", "5ef3e66a", "7576826b", "8ab08b6e", "58d5de1b", "d4d95451", "84b4ff0c", "66d8c2fa"}},
	}
	lines := make([]string, 0, len(records))
	for _, record := range records {
		lines = append(lines, "- "+record.id+" sha256:"+strings.Join(record.digest, ""))
	}
	digest := strings.Join(records[0].digest, "")
	changedDigest := digest[:len(digest)-1] + "0"
	if changedDigest == digest {
		changedDigest = digest[:len(digest)-1] + "1"
	}
	for _, test := range []struct {
		name         string
		path         string
		content      string
		wantFindings bool
	}{
		{"exact public records", "CHANGELOG.md", strings.Join(lines, "\n") + "\n", false},
		{"changed digest", "CHANGELOG.md", "- " + records[0].id + " sha256:" + changedDigest + "\n", true},
		{"changed decision", "CHANGELOG.md", "- JSONAPI-DEC-999 sha256:" + digest + "\n", true},
		{"different path", "docs/CHANGELOG.md", lines[0] + "\n", true},
		{"additional text", "CHANGELOG.md", lines[0] + " api_key=" + changedDigest + "\n", true},
		{"credential assignment", "CHANGELOG.md", "api_key = \"" + digest + "\"\n", true},
	} {
		t.Run(test.name, func(t *testing.T) {
			root, configPath := gitleaksRepository(t, map[string]string{test.path: test.content})
			for _, mode := range []string{"git", "dir"} {
				findings, _, err := runGitleaksIntegration(t, root, configPath, mode)
				if !test.wantFindings {
					if err != nil || len(findings) != 0 {
						t.Fatalf("%s: exact public decision records remain scanner findings (count=%d)", mode, len(findings))
					}
					continue
				}
				matched := false
				for _, finding := range findings {
					if finding.RuleID == "generic-api-key" && finding.File == test.path {
						matched = true
					}
				}
				if err == nil || !matched {
					t.Fatalf("%s: non-allowlisted content was not retained as a finding", mode)
				}
			}
		})
	}
}
