package gates

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"
)

const publicPolicySourcePath = "internal/gates/runner.go"

func publicPolicySourceRecords(t *testing.T) []string {
	t.Helper()
	fetch := func(url string, size int64, digest string) []byte {
		t.Helper()
		request, err := http.NewRequestWithContext(t.Context(), http.MethodGet, url, nil)
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
		data, err := io.ReadAll(io.LimitReader(response.Body, size+1))
		if err != nil || int64(len(data)) != size || fmt.Sprintf("%x", sha256.Sum256(data)) != digest {
			t.Fatal("immutable public source integrity differs")
		}
		return data
	}
	// The oracle is the published vector source, not the candidate policy.
	vectors := fetch("https://raw.githubusercontent.com/faustbrian/go-webhook/2ceba720743bda41e9a0b0c5da89ef244d7f58d8/testdata/vectors/v1.json",
		2642, "fe59f929dce9965dd12c15501c9380e806bf66e908096906fb152f488bd5d5ba")
	var document struct {
		Vectors []struct {
			Canonical string `json:"canonical_base64url"`
		} `json:"vectors"`
	}
	if err := json.Unmarshal(vectors, &document); err != nil || len(document.Vectors) != 2 {
		t.Fatal("public vector shape differs")
	}
	source := fetch("https://raw.githubusercontent.com/faustbrian/go-library-tools/2f7820490a2e00effbaa389a356ca99e2f8b2881/internal/gates/runner.go",
		60767, "ffbee370b2633a33800e6d787d2f532c18f871448df2a5a20b7c5468c1689197")
	lines := strings.Split(string(source), "\n")
	if len(lines) < 59 {
		t.Fatal("public policy source is truncated")
	}
	records := []string{lines[57], lines[58]}
	digests := []string{
		"bc8c1d572c2388ca2e82bb6ef3040ecd65e1cc56e1b304221b07cde50197188a",
		"b45c71c6414fc3be23ded032f63dbbf0ff3a2dad46a1cf953dbebf2aecf92d7a",
	}
	for index, record := range records {
		if fmt.Sprintf("%x", sha256.Sum256([]byte(record))) != digests[index] {
			t.Fatal("failing source record identity differs")
		}
		pattern := strings.TrimSuffix(strings.TrimPrefix(record, "  '''"), "''',")
		matcher, err := regexp.Compile(pattern)
		if err != nil {
			t.Fatal("immutable public record matcher is invalid")
		}
		for vectorIndex, vector := range document.Vectors {
			canonical, err := base64.RawURLEncoding.DecodeString(vector.Canonical)
			if err != nil {
				t.Fatal("public canonical record is invalid")
			}
			decodedField := "      \"canonical_base64url\": \"" + string(canonical) + "\","
			if matcher.MatchString(decodedField) != (index == vectorIndex) {
				t.Fatal("source record is not the independently published canonical vector")
			}
		}
	}
	return records
}

func TestPublicPolicySourceHistoryPinnedScanner(t *testing.T) {
	if os.Getenv("GOLIB_GITLEAKS_INTEGRATION") != "1" {
		t.Skip("opt-in pinned scanner contract")
	}
	records := publicPolicySourceRecords(t)
	public := strings.Join(records, "\n") + "\n"
	credential := strings.Join([]string{"A1b2C3d4", "E5f6G7h8", "I9j0K1l2", "M3n4O5p6", "Q7r8S9t0", "U1v2W3x4", "Y5z6A7b8", "C9d0E1f2"}, "")
	identity := func(finding gitleaksFinding) string {
		parts := strings.Split(finding.Fingerprint, ":")
		return finding.RuleID + "\t" + finding.File + "\t" + parts[len(parts)-1]
	}
	expected := func(path string, lines ...int) []string {
		result := make([]string, 0, len(lines))
		for _, line := range lines {
			result = append(result, fmt.Sprintf("generic-api-key\t%s\t%d", path, line))
		}
		slices.Sort(result)
		return result
	}
	scan := func(t *testing.T, root, policy, mode string, want []string) {
		t.Helper()
		findings, _, err := runGitleaksIntegration(t, root, policy, mode)
		if (err != nil) != (len(want) != 0) {
			t.Fatal("pinned scanner returned an incorrect finding status")
		}
		got := make([]string, 0, len(findings))
		for _, finding := range findings {
			got = append(got, identity(finding))
		}
		slices.Sort(got)
		if !slices.Equal(got, want) {
			t.Fatalf("safe finding identities differ: got %v, want %v", got, want)
		}
	}
	changed := func(index int) string {
		copyRecords := slices.Clone(records)
		bounds := regexp.MustCompile(`[A-Za-z0-9]{20,}`).FindStringIndex(copyRecords[index])
		if bounds == nil {
			t.Fatal("public source has no detector-shaped counterfactual")
		}
		position := bounds[1] - 1
		replacement := byte('Z')
		if copyRecords[index][position] == replacement {
			replacement = 'Y'
		}
		copyRecords[index] = copyRecords[index][:position] + string(replacement) + copyRecords[index][position+1:]
		return strings.Join(copyRecords, "\n") + "\n"
	}
	for _, scenario := range []struct {
		name, path, content string
		baseline, candidate []int
	}{
		{"exact public source", publicPolicySourcePath, public, []int{1, 2}, nil},
		{"changed first record", publicPolicySourcePath, changed(0), []int{1, 2}, []int{1}},
		{"changed second record", publicPolicySourcePath, changed(1), []int{1, 2}, []int{2}},
		{"different path", "other/" + publicPolicySourcePath, public, []int{1, 2}, []int{1, 2}},
		{"different framing", publicPolicySourcePath, " " + records[0] + "\n " + records[1] + "\n", []int{1, 2}, []int{1, 2}},
		{"adjacent credential", publicPolicySourcePath, public + "api_key=" + credential + "\n", []int{1, 2, 3}, []int{3}},
		{"same line credential", publicPolicySourcePath, records[0] + " api_key=" + credential + "\n" + records[1] + "\n", []int{1, 1, 2}, []int{1, 1}},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			root, candidate := gitleaksRepository(t, map[string]string{scenario.path: scenario.content})
			baseline := filepath.Join(t.TempDir(), "default.toml")
			if err := os.WriteFile(baseline, []byte("[extend]\nuseDefault = true\n"), 0600); err != nil {
				t.Fatal(err)
			}
			for _, mode := range []string{"git", "dir"} {
				scan(t, root, baseline, mode, expected(scenario.path, scenario.baseline...))
				scan(t, root, candidate, mode, expected(scenario.path, scenario.candidate...))
			}
		})
	}
	t.Run("shared owner and foreign history hosted", func(t *testing.T) {
		if os.Getenv("GITHUB_ACTIONS") != "true" {
			t.Skip("shared bounded executor integration is hosted-only")
		}
		for _, variant := range []string{"public", "foreign history credential", "untracked credential"} {
			t.Run(variant, func(t *testing.T) {
				files := map[string]string{publicPolicySourcePath: public}
				if variant == "foreign history credential" {
					files["credential.txt"] = "api_key=" + credential + "\n"
				}
				root, _ := gitleaksRepository(t, files)
				git := func(arguments ...string) string {
					t.Helper()
					// #nosec G204 -- fixed Git fixture operations and task-owned paths.
					command := exec.CommandContext(t.Context(), "git", arguments...)
					command.Dir = root
					output, err := command.Output()
					if err != nil {
						t.Fatal("owned Git fixture operation failed")
					}
					return strings.TrimSpace(string(output))
				}
				original := git("rev-parse", "HEAD")
				if variant == "foreign history credential" {
					git("branch", "foreign-history")
					git("symbolic-ref", "HEAD", "refs/heads/clean-head")
					if err := os.Remove(filepath.Join(root, "credential.txt")); err != nil {
						t.Fatal(err)
					}
					git("add", "credential.txt")
					git("-c", "user.name=golib-test", "-c", "user.email=golib-test@example.invalid", "commit", "-qm", "test: create clean independent fixture head")
				}
				if variant == "untracked credential" {
					if err := os.WriteFile(filepath.Join(root, "credential.txt"), []byte("api_key="+credential+"\n"), 0600); err != nil {
						t.Fatal(err)
					}
				}
				before := git("show-ref") + git("status", "--porcelain")
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
				err = (Runner{Root: root, Executor: executor, Output: &output}).Secrets(t.Context())
				if (err != nil) != (variant != "public") {
					t.Fatal("shared secret gate returned incorrect status")
				}
				diagnostic := output.String()
				if err != nil {
					if !strings.Contains(err.Error(), "secret-findings") {
						t.Fatal("shared gate did not classify completed findings")
					}
					diagnostic += err.Error()
					commit := "-"
					if variant == "foreign history credential" {
						commit = original
					}
					location := fmt.Sprintf("secret-location %x %x 1 %s", sha256.Sum256([]byte("credential.txt")), sha256.Sum256([]byte("generic-api-key")), commit)
					if !strings.Contains(diagnostic, location) {
						t.Fatal("shared gate lost safe credential location identity")
					}
				}
				if strings.Contains(diagnostic, credential) || before != git("show-ref")+git("status", "--porcelain") {
					t.Fatal("shared gate leaked credentials or mutated its source")
				}
			})
		}
	})
}
