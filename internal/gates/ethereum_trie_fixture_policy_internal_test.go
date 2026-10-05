package gates

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

// The input is public ethereum/tests@c67e485ff8b5be9abc8ad15345ec21aa22e290d9
// TrieTests/trietest.json, SHA256 0ce5e1151210958edf47911b332fe188696d741f9f44b1a471ee62bc666c1f0f.
// Split its synthetic value so this test source does not itself resemble a credential.
func ethereumTriePublicFixtureLine() string {
	value := strings.Join([]string{"01234567", "89012345", "67890123", "45678901", "23456789", "Very_Lon", "g"}, "")
	return "      [ \"key1\", \"" + value + "\"],"
}

func TestEthereumTrieFixturePolicyRequiresExactPublicLine(t *testing.T) {
	line := ethereumTriePublicFixtureLine()
	path := "testdata/ethereum-tests/TrieTests/trietest.json"
	for _, test := range []struct {
		name, path, rule, line string
		allowed                bool
	}{
		{"exact public line", path, "generic-api-key", line, true},
		{"scanner line framing", path, "generic-api-key", "\n" + line, true},
		{"different rule", path, "another-rule", line, false},
		{"different path", "testdata/other/trietest.json", "generic-api-key", line, false},
		{"path prefix", "nested/" + path, "generic-api-key", line, false},
		{"path suffix", path + ".other", "generic-api-key", line, false},
		{"changed key", path, "generic-api-key", strings.Replace(line, "key1", "key2", 1), false},
		{"changed value", path, "generic-api-key", strings.Replace(line, strings.Join([]string{"01234567", "89012345", "67890123", "45678901", "23456789", "Very_Lon", "g"}, ""), "ordinary-replacement-value", 1), false},
		{"changed whitespace", path, "generic-api-key", " " + line, false},
		{"line prefix", path, "generic-api-key", "prefix" + line, false},
		{"line suffix", path, "generic-api-key", line + "suffix", false},
		{"different assignment", path, "generic-api-key", "api_key = \"ordinary-replacement-value\"", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := ethereumTrieFixturePolicyAllows(t, test.path, test.rule, test.line); got != test.allowed {
				t.Fatalf("public fixture allowance = %v, want %v", got, test.allowed)
			}
		})
	}
}

// Evaluate the rendered expressions at Gitleaks' documented rule/path/line
// boundary without running a scanner. The hosted test separately covers parsing.
func ethereumTrieFixturePolicyAllows(t *testing.T, path, rule, line string) bool {
	t.Helper()
	for block := range strings.SplitSeq(gitleaksPolicy, "[[allowlists]]") {
		if !strings.Contains(block, `description = "Exact public Ethereum legacy trie fixture line."`) {
			continue
		}
		if !strings.Contains(block, `condition = "AND"`) ||
			!strings.Contains(block, `targetRules = ["generic-api-key"]`) ||
			!strings.Contains(block, `regexTarget = "line"`) {
			t.Fatal("fixture policy lacks its exact rule/path/line conjunction")
		}
		patterns := func(key string) string {
			t.Helper()
			match := regexp.MustCompile(key + ` = \['''(.*?)'''\]\n`).FindStringSubmatch(block)
			if len(match) != 2 {
				t.Fatal("fixture policy expression is missing")
			}
			return match[1]
		}
		return rule == "generic-api-key" &&
			regexp.MustCompile(patterns("paths")).MatchString(path) &&
			regexp.MustCompile(patterns("regexes")).MatchString(line)
	}
	return false
}

func TestEthereumTrieFixturePolicyHosted(t *testing.T) {
	if os.Getenv("GITHUB_ACTIONS") != "true" || os.Getenv("GOLIB_GITLEAKS_INTEGRATION") != "1" {
		t.Skip("pinned scanner parsing and finding retention run only in hosted CI")
	}
	line := ethereumTriePublicFixtureLine()
	path := "testdata/ethereum-tests/TrieTests/trietest.json"
	for _, test := range []struct {
		name, path, line string
		findings         bool
	}{
		{"exact public line", path, line, false},
		{"changed key", path, strings.Replace(line, "key1", "key2", 1), true},
		{"different path", "testdata/other/trietest.json", line, true},
		{"line suffix", path, line + "suffix", true},
	} {
		t.Run(test.name, func(t *testing.T) {
			root, configPath := gitleaksRepository(t, map[string]string{test.path: test.line + "\n"})
			for _, mode := range []string{"git", "dir"} {
				findings, _, err := runGitleaksIntegration(t, root, configPath, mode)
				if !test.findings {
					if err != nil || len(findings) != 0 {
						t.Fatalf("%s: exact public fixture remains a finding", mode)
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
					t.Fatalf("%s: changed tuple did not retain its finding", mode)
				}
			}
		})
	}
}
