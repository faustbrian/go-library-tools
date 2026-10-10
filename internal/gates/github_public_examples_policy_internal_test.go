package gates

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"
)

const githubPublicSpecPath = "specification/independent/github-rest-api/api.github.com.2022-11-28.json"

// This oracle is the immutable official source, not candidate policy data.
// Its exact bytes are also imported by go-openapi@3ef6f63e06f9383739d3371dee1c8a245d1ba67d.
func githubPublicSpec(t *testing.T) string {
	t.Helper()
	request, err := http.NewRequestWithContext(t.Context(), http.MethodGet,
		"https://raw.githubusercontent.com/github/rest-api-description/417c4fb368fc6a7162ce5f3eeeddce1a9a217747/descriptions-next/api.github.com/api.github.com.2022-11-28.json", nil)
	if err != nil {
		t.Fatal("public source request failed")
	}
	client := &http.Client{Timeout: 45 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		t.Fatal("immutable public source unavailable")
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatal("immutable public source status differs")
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, 12834429))
	if err != nil || len(data) != 12834428 || fmt.Sprintf("%x", sha256.Sum256(data)) != "9d85f3a842c0215768f30f83ac7d1595430236fc51ce9c84e344b991a9f6b3da" {
		t.Fatal("immutable public source integrity differs")
	}
	return string(data)
}

type githubSpecLocation struct {
	rule string
	line int
}

func githubSpecOracle() []githubSpecLocation {
	return []githubSpecLocation{
		{"generic-api-key", 4613},
		{"generic-api-key", 4672},
		{"generic-api-key", 4744},
		{"generic-api-key", 4813},
		{"generic-api-key", 4912},
		{"generic-api-key", 81540},
		{"generic-api-key", 83925},
		{"generic-api-key", 312956},
		{"generic-api-key", 312957},
		{"generic-api-key", 313315},
		{"generic-api-key", 313347},
		{"generic-api-key", 313388},
		{"generic-api-key", 317150},
		{"generic-api-key", 317417},
		{"generic-api-key", 318610},
		{"generic-api-key", 318957},
		{"generic-api-key", 318965},
		{"generic-api-key", 319117},
		{"generic-api-key", 319123},
		{"generic-api-key", 319238},
		{"generic-api-key", 320478},
		{"generic-api-key", 320817},
		{"generic-api-key", 321824},
		{"generic-api-key", 321986},
		{"generic-api-key", 322146},
		{"generic-api-key", 322604},
		{"generic-api-key", 324279},
		{"generic-api-key", 324287},
		{"generic-api-key", 324419},
		{"generic-api-key", 324541},
		{"generic-api-key", 325890},
		{"generic-api-key", 326236},
		{"generic-api-key", 326244},
		{"generic-api-key", 326382},
		{"generic-api-key", 326503},
		{"generic-api-key", 329080},
		{"generic-api-key", 329185},
		{"generic-api-key", 329438},
		{"generic-api-key", 331686},
		{"generic-api-key", 331834},
		{"generic-api-key", 332148},
		{"generic-api-key", 334279},
		{"generic-api-key", 337037},
		{"generic-api-key", 337179},
		{"generic-api-key", 337918},
		{"generic-api-key", 338066},
		{"generic-api-key", 338439},
		{"generic-api-key", 338587},
		{"generic-api-key", 340945},
		{"generic-api-key", 341735},
		{"generic-api-key", 341892},
		{"generic-api-key", 342187},
		{"generic-api-key", 342195},
		{"generic-api-key", 342326},
		{"generic-api-key", 342448},
		{"generic-api-key", 342808},
		{"generic-api-key", 342817},
		{"generic-api-key", 342829},
		{"generic-api-key", 343197},
		{"generic-api-key", 343359},
		{"generic-api-key", 343520},
		{"generic-api-key", 343807},
		{"generic-api-key", 343840},
		{"generic-api-key", 343846},
		{"generic-api-key", 343855},
		{"generic-api-key", 343969},
		{"github-app-token", 5465},
		{"github-app-token", 313203},
		{"github-app-token", 313345},
		{"github-app-token", 313386},
		{"github-oauth", 5464},
		{"github-pat", 5462},
		{"github-refresh-token", 5466},
		{"jwt", 68325},
		{"private-key", 312958},
	}
}

func githubSpecIdentities(t *testing.T, findings []gitleaksFinding) []string {
	t.Helper()
	result := make([]string, 0, len(findings))
	for _, f := range findings {
		parts := strings.Split(f.Fingerprint, ":")
		line, err := strconv.Atoi(parts[len(parts)-1])
		if err != nil || line < 1 {
			t.Fatal("scanner location lacks line identity")
		}
		result = append(result, fmt.Sprintf("%s:%s:%d", f.RuleID, filepath.ToSlash(f.File), line))
	}
	slices.Sort(result)
	return result
}

func TestGitHubPublicExamplesPolicyIntegration(t *testing.T) {
	if os.Getenv("GOLIB_GITLEAKS_INTEGRATION") != "1" {
		t.Skip("set GOLIB_GITLEAKS_INTEGRATION=1 for pinned scanner verification")
	}
	source := githubPublicSpec(t)
	lines := strings.Split(strings.TrimSuffix(source, "\n"), "\n")
	// Full authentic blob proves the observed 75 findings. Compact independent
	// source lines cover hostile variants without repeatedly scanning unrelated
	// megabytes; their content comes only from the verified oracle above.
	var compactLines []string
	var compactOracle []githubSpecLocation
	positions := map[string]int{}
	seenTuples := map[string]bool{}
	for _, item := range githubSpecOracle() {
		line := lines[item.line-1]
		n := positions[line]
		if n == 0 {
			compactLines = append(compactLines, line)
			n = len(compactLines)
			positions[line] = n
		}
		key := fmt.Sprintf("%s:%d", item.rule, n)
		if !seenTuples[key] {
			compactOracle = append(compactOracle, githubSpecLocation{item.rule, n})
			seenTuples[key] = true
		}
	}
	compact := strings.Join(compactLines, "\n") + "\n"
	defaultConfig := filepath.Join(t.TempDir(), "default.toml")
	if err := os.WriteFile(defaultConfig, []byte("[extend]\nuseDefault = true\n"), 0600); err != nil {
		t.Fatal(err)
	}
	scan := func(t *testing.T, root, config, mode string, want []string) []gitleaksFinding {
		t.Helper()
		findings, _, err := runGitleaksIntegration(t, root, config, mode)
		if (err != nil) != (len(want) > 0) || !slices.Equal(githubSpecIdentities(t, findings), want) {
			t.Fatalf("scanner identities/status differ: got %v, want %v, failed=%v", githubSpecIdentities(t, findings), want, err != nil)
		}
		return findings
	}
	identities := func(path string, locations []githubSpecLocation) []string {
		result := make([]string, 0, len(locations))
		for _, x := range locations {
			result = append(result, fmt.Sprintf("%s:%s:%d", x.rule, path, x.line))
		}
		slices.Sort(result)
		return result
	}
	t.Run("authentic complete public source", func(t *testing.T) {
		root, config := gitleaksRepository(t, map[string]string{githubPublicSpecPath: source})
		for _, mode := range []string{"git", "dir"} {
			t.Run(mode, func(t *testing.T) {
				scan(t, root, defaultConfig, mode, identities(githubPublicSpecPath, githubSpecOracle()))
				scan(t, root, config, mode, nil)
			})
		}
	})
	t.Run("changed content fields framing and adjacency", func(t *testing.T) {
		var hostile []string
		var changed []githubSpecLocation
		seen := map[string]bool{}
		for _, item := range githubSpecOracle() {
			if seen[item.rule] {
				continue
			}
			seen[item.rule] = true
			line := lines[item.line-1]
			// Preserve each detector's token shape while changing a payload byte.
			offset := strings.Index(line, ": ") + 3
			if offset < 3 {
				offset = strings.Index(line, "\"") + 1
			}
			if item.rule == "private-key" {
				offset = strings.Index(line, `\n`) + 2
			}
			offset += 8
			if offset >= len(line) {
				t.Fatal("source representative lacks payload")
			}
			replacement := byte('A')
			if line[offset] == replacement {
				replacement = 'B'
			}
			hostile = append(hostile, line[:offset]+string(replacement)+line[offset+1:])
			changed = append(changed, githubSpecLocation{item.rule, len(compactLines) + len(hostile)})
			// Exact public bytes are not exempt in a different named field/assignment.
			if colon := strings.Index(line, ": "); item.rule == "generic-api-key" && colon >= 0 {
				payload := line[colon+2:]
				hostile = append(hostile, `  "new_api_key": `+payload, `api_key = `+payload)
			}
			if item.rule == "generic-api-key" {
				hostile = append(hostile, "prefix "+line, line+" suffix")
			}
		}
		token := strings.Join([]string{"gh", "p_", "A1b2C3d4E5f6", "G7h8I9j0K1l2", "M3n4O5p6Q7r8"}, "")
		hostile = append(hostile, lines[5462-1]+" "+token, token)
		root, config := gitleaksRepository(t, map[string]string{githubPublicSpecPath: compact + strings.Join(hostile, "\n") + "\n"})
		for _, mode := range []string{"git", "dir"} {
			t.Run(mode, func(t *testing.T) {
				baseline, _, err := runGitleaksIntegration(t, root, defaultConfig, mode)
				if err == nil {
					t.Fatal("negative fixture did not exercise scanner")
				}
				all := githubSpecIdentities(t, baseline)
				for _, id := range identities(githubPublicSpecPath, changed) {
					if !slices.Contains(all, id) {
						t.Fatal("changed representative no longer exercises original detector")
					}
				}
				var retained []string
				public := identities(githubPublicSpecPath, compactOracle)
				for _, id := range all {
					if !slices.Contains(public, id) {
						retained = append(retained, id)
					}
				}
				// Every appended hostile line must produce a default finding. This prevents
				// an unrecognized key/token mutation from masquerading as retention proof.
				for n := range hostile {
					suffix := fmt.Sprintf(":%d", len(compactLines)+n+1)
					found := false
					for _, id := range retained {
						if strings.HasSuffix(id, suffix) {
							found = true
						}
					}
					if !found {
						t.Fatalf("negative line %d lacks default detector proof", n+1)
					}
				}
				scan(t, root, config, mode, retained)
			})
		}
	})
	t.Run("path near misses", func(t *testing.T) {
		files := map[string]string{"nested/" + githubPublicSpecPath: compact, githubPublicSpecPath + ".bak": compact}
		root, config := gitleaksRepository(t, files)
		var want []string
		for path := range files {
			want = append(want, identities(path, compactOracle)...)
		}
		slices.Sort(want)
		for _, mode := range []string{"git", "dir"} {
			scan(t, root, defaultConfig, mode, want)
			scan(t, root, config, mode, want)
		}
	})
	t.Run("deleted credential stays historical", func(t *testing.T) {
		token := strings.Join([]string{"gh", "p_", "A1b2C3d4E5f6", "G7h8I9j0K1l2", "M3n4O5p6Q7r8"}, "")
		root, config := gitleaksRepository(t, map[string]string{githubPublicSpecPath: compact + token + "\n"})
		// #nosec G204 -- fixed Git command in an ordinary owned fixture.
		before := exec.CommandContext(t.Context(), "git", "rev-parse", "HEAD")
		before.Dir = root
		commit, err := before.Output()
		if err != nil {
			t.Fatal("fixture identity failed")
		}
		if err := os.WriteFile(filepath.Join(root, githubPublicSpecPath), []byte(compact), 0600); err != nil {
			t.Fatal(err)
		}
		for _, args := range [][]string{{"add", githubPublicSpecPath}, {"-c", "user.name=golib-test", "-c", "user.email=golib-test@example.invalid", "commit", "-qm", "remove fixture credential"}} {
			// #nosec G204 -- fixed Git operations in an ordinary owned fixture.
			command := exec.CommandContext(t.Context(), "git", args...)
			command.Dir = root
			if err := command.Run(); err != nil {
				t.Fatal("fixture commit failed")
			}
		}
		findings := scan(t, root, config, "git", identities(githubPublicSpecPath, []githubSpecLocation{{"github-pat", len(compactLines) + 1}}))
		if !strings.HasPrefix(findings[0].Fingerprint, strings.TrimSpace(string(commit))+":") {
			t.Fatal("deleted finding lost original commit identity")
		}
		scan(t, root, config, "dir", nil)
	})
	t.Run("shared secrets owner hosted", func(t *testing.T) {
		if os.Getenv("GITHUB_ACTIONS") != "true" {
			t.Skip("shared bounded executor integration is hosted-only")
		}
		token := strings.Join([]string{"gh", "p_", "A1b2C3d4E5f6", "G7h8I9j0K1l2", "M3n4O5p6Q7r8"}, "")
		for _, hostile := range []bool{false, true} {
			t.Run(fmt.Sprintf("hostile=%v", hostile), func(t *testing.T) {
				content := compact
				if hostile {
					content += token + "\n"
				}
				root, _ := gitleaksRepository(t, map[string]string{githubPublicSpecPath: content})
				var output bytes.Buffer
				executor, cleanup, err := NewProcessExecutor(root, io.Discard, io.Discard)
				if err != nil {
					t.Fatal("owned executor setup failed")
				}
				defer func() {
					if err := cleanup(); err != nil {
						t.Fatal("owned executor cleanup failed")
					}
				}()
				runner := Runner{Root: root, Executor: executor, Output: &output}
				err = runner.Secrets(t.Context())
				if (err != nil) != hostile || (hostile && !strings.Contains(output.String(), "secret-findings")) || strings.Contains(output.String(), token) {
					t.Fatal("shared secret gate status or private diagnostic differs")
				}
			})
		}
	})

}
