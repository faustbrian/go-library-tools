package gates

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// Public records independently recomputed from immutable OpenAPI decision
// registers 09771cf1e03bbdb684cca77785cbc7ee1ee2e9a1 and
// 92dcffd42c0dbd6d27c526e5cf33ea7f9b8591f1, not from policy expressions.
func openapiPublicDigestLines() []string {
	records := []struct {
		id     string
		digest []string
	}{
		{"OPENAPI-DEC-001", []string{"945adbbd", "c8784d07", "8c0c92f1", "8e25f093", "35b55c60", "f4dabc6c", "4cd892c4", "edf8e227"}},
		{"OPENAPI-DEC-002", []string{"7a6781bc", "ac02f255", "55164496", "9f0b5331", "0bddcb9c", "417eaaf7", "062ecb0e", "b421a193"}},
		{"OPENAPI-DEC-003", []string{"13253357", "9e14d96e", "ccffbf6c", "9456d1a9", "e36d895e", "ffb2b977", "d4deffdf", "a7e11e22"}},
		{"OPENAPI-DEC-004", []string{"3412d17f", "04d660ca", "df876062", "ff3f6775", "31a6e928", "d81ebf6e", "98926f74", "ef9a5525"}},
		{"OPENAPI-DEC-005", []string{"457129d9", "ab2d52e9", "0b4f2dff", "335b8cc3", "d0faeee8", "dc104f0e", "78820ab8", "b7d0b2da"}},
		{"OPENAPI-DEC-006", []string{"e0f1f7cd", "1e4806a3", "0399749f", "079f3700", "530c70f8", "ef8bf86e", "6ef92924", "57154b70"}},
		{"OPENAPI-DEC-007", []string{"514cb998", "9d61158b", "fb863b95", "37ae21a2", "f8a8682c", "e5baf327", "4b220ab5", "00a9fbed"}},
		{"OPENAPI-DEC-008", []string{"cc10809f", "406b1233", "6deb5353", "7ad9146b", "a71ac98c", "6962779c", "229dc203", "c501e018"}},
		{"OPENAPI-DEC-009", []string{"962eee67", "6c2080c6", "558e03d5", "3fe08911", "ab6cc2e9", "de353428", "f7f4cef6", "659bc300"}},
		{"OPENAPI-DEC-010", []string{"192a09d2", "7fcf47ca", "576b13c5", "37c30266", "880ca429", "8db996e4", "00c41f3f", "fbec96af"}},
		{"OPENAPI-DEC-001", []string{"5c40c0ec", "0ffe030e", "87f09461", "57b1670d", "d6e681ca", "fbd45833", "86df6698", "94d39882"}},
		{"OPENAPI-DEC-004", []string{"c38e2c7b", "778c8636", "f4ed8c9b", "7231a104", "9cfac298", "59a655b5", "c7011d44", "159bd1e9"}},
		{"OPENAPI-DEC-006", []string{"df83163c", "b3d52899", "abcdfcfc", "d95245e3", "ea9914b4", "1401ef55", "2eb6dc88", "12ded7bb"}},
		{"OPENAPI-DEC-010", []string{"cfd3ed8d", "4b897ba3", "9715ecd0", "61f875b6", "d5da62eb", "87aab1a0", "17de8e9f", "79fb37a7"}},
	}
	lines := make([]string, 0, len(records))
	for _, record := range records {
		lines = append(lines, "- "+record.id+" sha256:"+strings.Join(record.digest, ""))
	}
	return lines
}

func TestOpenAPIDecisionDigestPolicyIntegration(t *testing.T) {
	if os.Getenv("GOLIB_GITLEAKS_INTEGRATION") != "1" {
		t.Skip("set GOLIB_GITLEAKS_INTEGRATION=1 for pinned scanner verification")
	}
	lines := openapiPublicDigestLines()
	digest := strings.Split(lines[0], "sha256:")[1]
	changed := digest[:63] + "0"
	if changed == digest {
		changed = digest[:63] + "1"
	}
	githubPAT := strings.Join([]string{"gh", "p_", "A1b2C3d4E5f6", "G7h8I9j0K1l2", "M3n4O5p6Q7r8"}, "")
	generic := "generic-api-key"
	type location struct {
		rule, path string
		line       int
	}
	type corpus struct {
		name     string
		files    map[string]string
		rejected []location
	}
	cases := []corpus{
		{"all public records", map[string]string{"CHANGELOG.md": strings.Join(lines, "\n") + "\n"}, nil},
		{"hostile tuples and adjacent credential", map[string]string{
			"CHANGELOG.md": strings.Join([]string{
				lines[0],
				strings.Replace(lines[0], digest, changed, 1),
				"- OPENAPI-DEC-999 sha256:" + digest,
				"- OPENAPI-DEC-001 sha256:" + strings.Split(lines[1], "sha256:")[1],
				"prefix " + lines[0],
				lines[0] + " suffix",
				"api_key = \"" + digest + "\"",
				"api_key = \"" + changed + "\"",
			}, "\n") + "\n",
			"docs/CHANGELOG.md": lines[0] + "\n",
			"CHANGELOG.md.bak":  lines[0] + "\n",
		}, []location{{generic, "CHANGELOG.md", 2}, {generic, "CHANGELOG.md", 3}, {generic, "CHANGELOG.md", 4}, {generic, "CHANGELOG.md", 5}, {generic, "CHANGELOG.md", 6}, {generic, "CHANGELOG.md", 7}, {generic, "CHANGELOG.md", 8}, {generic, "docs/CHANGELOG.md", 1}, {generic, "CHANGELOG.md.bak", 1}}},
		{"other detector", map[string]string{"CHANGELOG.md": lines[0] + "\n" + githubPAT + "\n"}, []location{{"github-pat", "CHANGELOG.md", 2}}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			root, config := gitleaksRepository(t, test.files)
			defaultConfig := filepath.Join(t.TempDir(), "default.toml")
			if err := os.WriteFile(defaultConfig, []byte("[extend]\nuseDefault = true\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			for _, mode := range []string{"git", "dir"} {
				t.Run(mode, func(t *testing.T) {
					// Full default-policy identity set proves each fixture exercises
					// its detector before testing the narrower central classification.
					wantDefault := slices.Clone(test.rejected)
					if test.name == "all public records" {
						for i := range lines {
							wantDefault = append(wantDefault, location{generic, "CHANGELOG.md", i + 1})
						}
					} else {
						wantDefault = append(wantDefault, location{generic, "CHANGELOG.md", 1})
					}
					scan := func(config string, want []location) {
						t.Helper()
						findings, _, err := runGitleaksIntegration(t, root, config, mode)
						if len(want) == 0 {
							if err != nil || len(findings) != 0 {
								t.Fatalf("exact public metadata remains findings: count=%d, scannerFailed=%v", len(findings), err != nil)
							}
							return
						}
						if err == nil {
							t.Fatal("scanner accepted hostile fixture")
						}
						got := make([]string, 0, len(findings))
						expected := make([]string, 0, len(want))
						for _, finding := range findings {
							parts := strings.Split(finding.Fingerprint, ":")
							n, parseErr := strconv.Atoi(parts[len(parts)-1])
							if parseErr != nil || n < 1 {
								t.Fatal("scanner finding lacks line identity")
							}
							got = append(got, fmt.Sprintf("%s:%s:%d", finding.RuleID, filepath.ToSlash(finding.File), n))
						}
						for _, item := range want {
							expected = append(expected, fmt.Sprintf("%s:%s:%d", item.rule, item.path, item.line))
						}
						slices.Sort(got)
						slices.Sort(expected)
						if !slices.Equal(got, expected) {
							t.Fatalf("finding identities differ: got %v, want %v", got, expected)
						}
					}
					scan(defaultConfig, wantDefault)
					scan(config, test.rejected)
				})
			}
		})
	}
	t.Run("removed credential remains historical finding", func(t *testing.T) {
		root, config := gitleaksRepository(t, map[string]string{"CHANGELOG.md": lines[0] + "\n" + githubPAT + "\n"})
		if err := os.WriteFile(filepath.Join(root, "CHANGELOG.md"), []byte(lines[0]+"\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		for _, args := range [][]string{{"add", "CHANGELOG.md"}, {"-c", "user.name=golib-test", "-c", "user.email=golib-test@example.invalid", "commit", "-qm", "remove fixture credential"}} {
			// #nosec G204 -- fixed Git operations in an ordinary test-owned fixture.
			command := exec.CommandContext(t.Context(), "git", args...)
			command.Dir = root
			if err := command.Run(); err != nil {
				t.Fatal("fixture commit failed")
			}
		}
		findings, _, err := runGitleaksIntegration(t, root, config, "git")
		if err == nil || len(findings) != 1 || findings[0].RuleID != "github-pat" || findings[0].File != "CHANGELOG.md" || !strings.HasSuffix(findings[0].Fingerprint, ":2") {
			t.Fatal("historical credential identity was not retained")
		}
		findings, _, err = runGitleaksIntegration(t, root, config, "dir")
		if err != nil || len(findings) != 0 {
			t.Fatal("current metadata-only tree did not pass")
		}
	})
}
