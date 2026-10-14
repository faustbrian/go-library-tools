package gates

import (
	"encoding/json"
	"os"
	"regexp"
	"strings"
	"testing"
)

// These are public source identities and synthetic protocol vocabulary from
// identity ef377dd and its control snapshot 93b4f70, not live credentials.
// Split source fragments retain exact bytes without making this fixture a
// second scanner-shaped copy of the same public metadata.
func TestIdentityPublicMetadataPinnedScanner(t *testing.T) {
	if os.Getenv("GOLIB_GITLEAKS_INTEGRATION") != "1" {
		t.Skip("opt-in pinned scanner contract")
	}
	data, err := os.ReadFile("testdata/identity_public_metadata.json")
	if err != nil {
		t.Fatal(err)
	}
	var records []struct {
		Path      string     `json:"path"`
		Fragments [][]string `json:"fragments"`
	}
	if err := json.Unmarshal(data, &records); err != nil || len(records) == 0 {
		t.Fatal("missing public metadata fixtures")
	}
	files := make(map[string]string)
	for _, record := range records {
		for _, fragments := range record.Fragments {
			files[record.Path] += strings.Join(fragments, "") + "\n"
		}
	}
	credential := strings.Join([]string{"A1b2C3d4", "E5f6G7h8", "I9j0K1l2", "M3n4O5p6", "Q7r8S9t0", "U1v2W3x4", "Y5z6A7b8", "C9d0E1f2"}, "")
	for _, scenario := range []struct {
		name   string
		mutate bool
		inline bool
	}{{"public records", false, false}, {"adjacent credentials", true, false}, {"same-line credentials", true, true}} {
		t.Run(scenario.name, func(t *testing.T) {
			input := make(map[string]string, len(files))
			for path, content := range files {
				if scenario.mutate {
					if scenario.inline {
						content = strings.TrimSuffix(content, "\n") + " api_key=" + credential + "\n"
					} else {
						content += "\n\napi_key=" + credential + "\n"
					}
				}
				input[path] = content
			}
			root, policy := gitleaksRepository(t, input)
			for _, mode := range []string{"git", "dir"} {
				findings, _, err := runGitleaksIntegration(t, root, policy, mode)
				if !scenario.mutate && (err != nil || len(findings) != 0) {
					t.Fatalf("%s rejected public metadata (findings=%d)", mode, len(findings))
				}
				if scenario.mutate {
					found := make(map[string]bool)
					for _, finding := range findings {
						if finding.RuleID == "generic-api-key" {
							found[finding.File] = true
						}
					}
					if err == nil || len(found) != len(files) {
						for path := range files {
							if !found[path] {
								t.Logf("missing credential path: %s", path)
							}
						}
						t.Fatalf("%s lost adjacent credentials (paths=%d, want %d)", mode, len(found), len(files))
					}
				}
			}
		})
	}
}

func TestIdentityMetadataPolicyRequiresExactTuple(t *testing.T) {
	data, err := os.ReadFile("testdata/identity_public_metadata.json")
	if err != nil {
		t.Fatal(err)
	}
	var records []struct {
		Path   string     `json:"path"`
		Rule   string     `json:"rule"`
		Values [][]string `json:"values"`
	}
	if err := json.Unmarshal(data, &records); err != nil {
		t.Fatal(err)
	}
	allows := func(path, rule, value string) bool {
		for block := range strings.SplitSeq(gitleaksPolicy, "[[allowlists]]") {
			if !strings.Contains(block, `description = "Exact public Identity metadata:`) {
				continue
			}
			if !strings.Contains(block, `condition = "AND"`) || !strings.Contains(block, `regexTarget = "secret"`) {
				t.Fatal("public metadata lost its exact tuple conjunction")
			}
			if !strings.Contains(block, `targetRules = ["`+rule+`"]`) {
				continue
			}
			matches := func(key, value string) bool {
				array := regexp.MustCompile(`(?s)` + key + ` = \[(.*?)\]\n`).FindStringSubmatch(block)
				if len(array) != 2 {
					t.Fatal("missing policy expression array")
				}
				for _, expression := range regexp.MustCompile(`'''(.*?)'''`).FindAllStringSubmatch(array[1], -1) {
					if regexp.MustCompile(expression[1]).MatchString(value) {
						return true
					}
				}
				return false
			}
			if matches("paths", path) && matches("regexes", value) {
				return true
			}
		}
		return false
	}
	for _, record := range records {
		for _, fragments := range record.Values {
			value := strings.Join(fragments, "")
			if !allows(record.Path, record.Rule, value) {
				t.Fatal("missing reviewed tuple")
			}
			for _, wrong := range []struct{ path, rule, value string }{
				{"other/" + record.Path, record.Rule, value},
				{record.Path + ".other", record.Rule, value},
				{record.Path, "other-rule", value},
				{record.Path, record.Rule, value + "!"},
				{record.Path, record.Rule, "!" + value},
				{record.Path, record.Rule, value[:len(value)-1] + "!"},
			} {
				if allows(wrong.path, wrong.rule, wrong.value) {
					t.Fatal("policy admitted a changed tuple")
				}
			}
		}
	}
}
