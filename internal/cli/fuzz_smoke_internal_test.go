package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go/format"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/faustbrian/go-library-tools/v2/internal/gates"
)

func TestExecuteChangedLocalFuzzSelectsTransitiveTargetInOneSession(t *testing.T) {
	root := internalFixture(t)
	policy := `schema_version: 1
tool_version: v1.0.0
operations:
  - module: .
    gate: fuzz
    steps:
      - type: go-test
        packages: [./parser]
        fuzz: ^FuzzParse$
        budget: 10m
      - type: go-test
        packages: [./other]
        fuzz: ^FuzzOther$
        budget: 10m
      - type: go-test
        packages: [./parser]
        fuzz: ^FuzzParse$
        budget: 10m
`
	if err := os.WriteFile(filepath.Join(root, ".golib.yaml"), []byte(policy), 0o600); err != nil {
		t.Fatal(err)
	}
	manifest, err := os.ReadFile(filepath.Join(root, "modules.json"))
	if err != nil {
		t.Fatal(err)
	}
	manifest = bytes.Replace(manifest, []byte(`"gates":{}`), []byte(`"gates":{"fuzz":true}`), 1)
	if err := os.WriteFile(filepath.Join(root, "modules.json"), manifest, 0o600); err != nil {
		t.Fatal(err)
	}
	created, cleaned := 0, 0
	var fuzzCommands []gates.Command
	factory := func(string, io.Writer, io.Writer) (gates.Executor, func() error, error) {
		created++
		return changedLocalCLIExecutor(t, func(ctx context.Context, command gates.Command) error {
			if command.Name == "git" && slices.Contains(command.Args, "diff") {
				_, err := io.WriteString(command.Stdout, "helper/value.go\x00")
				return err
			}
			if command.Name == "go" && slices.Contains(command.Args, "list") {
				for _, record := range []map[string]any{
					{"ImportPath": "example/parser", "Dir": filepath.Join(root, "parser"), "Imports": []string{"example/helper"}},
					{"ImportPath": "example/helper", "Dir": filepath.Join(root, "helper")},
					{"ImportPath": "example/other", "Dir": filepath.Join(root, "other")},
				} {
					if err := json.NewEncoder(command.Stdout).Encode(record); err != nil {
						return err
					}
				}
			}
			if command.Name == "go" && slices.Contains(command.Args, "-list=^FuzzParse$") {
				_, err := fmt.Fprintln(command.Stdout, "FuzzParse")
				return err
			}
			if command.Name == "go" && slices.Contains(command.Args, "-fuzz=^FuzzParse$") {
				if _, ok := ctx.Deadline(); !ok {
					t.Error("fuzz execution has no wall deadline")
				}
				fuzzCommands = append(fuzzCommands, command)
			}
			return nil
		}), func() error { cleaned++; return nil }, nil
	}
	var stdout, stderr bytes.Buffer
	code := execute([]string{"check", "--local", "--base-sha", strings.Repeat("a", 40), "--module", "."}, root, &stdout, &stderr, factory)
	if code != 0 {
		t.Fatalf("changed local code = %d, diagnostic = %q", code, stderr.String())
	}
	if created != 1 || cleaned != 1 || len(fuzzCommands) != 1 {
		t.Fatalf("sessions = %d/%d, fuzz commands = %d", created, cleaned, len(fuzzCommands))
	}
	for _, argument := range []string{"./parser", "-fuzztime=2s", "-parallel=2", "-p=1", "-count=1"} {
		if !slices.Contains(fuzzCommands[0].Args, argument) {
			t.Errorf("missing bounded fuzz argument %q", argument)
		}
	}
}

func TestExecuteChangedLocalFuzzPreservesLegacyAndStopsOnLocalFailure(t *testing.T) {
	for _, scenario := range []struct {
		name    string
		args    []string
		failure bool
		want    int
	}{
		{"legacy local", []string{"check", "--local"}, false, 0},
		{"missing base", []string{"check", "--local", "--base-sha"}, false, 2},
		{"invalid base", []string{"check", "--local", "--base-sha", "invalid"}, false, 1},
		{"full contract rejects smoke flag", []string{"check", "--base-sha", strings.Repeat("a", 40)}, false, 2},
		{"local failure", []string{"check", "--local", "--base-sha", strings.Repeat("a", 40)}, true, 1},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			root := internalFixture(t)
			var stdout, stderr bytes.Buffer
			factory := func(string, io.Writer, io.Writer) (gates.Executor, func() error, error) {
				return changedLocalCLIExecutor(t, func(_ context.Context, command gates.Command) error {
					if command.Name == "git" && slices.Contains(command.Args, "diff") {
						_, _ = io.WriteString(command.Stdout, "example.go\x00")
					}
					if command.Name == "go" && slices.Contains(command.Args, "list") {
						t.Fatal("legacy, invalid input or failed local check started discovery")
					}
					if scenario.failure && command.Name == "gofmt" {
						return errors.New("fixture local failure")
					}
					return nil
				}), func() error { return nil }, nil
			}
			if code := execute(scenario.args, root, &stdout, &stderr, factory); code != scenario.want {
				t.Fatalf("code = %d, want %d, diagnostic = %q", code, scenario.want, stderr.String())
			}
		})
	}
}

// This integration executes only a small, successful, task-owned Go module.
// Process termination and failure-control cases use the fake executor above.
func TestExecuteChangedLocalFuzzDiscoversNativeTestDependency(t *testing.T) {
	root := internalFixture(t)
	manifest, err := os.ReadFile(filepath.Join(root, "modules.json"))
	if err != nil {
		t.Fatal(err)
	}
	manifest = bytes.Replace(manifest, []byte(`"gates":{}`), []byte(`"gates":{"tests":true,"fuzz":true}`), 1)
	if err := os.WriteFile(filepath.Join(root, "modules.json"), manifest, 0o600); err != nil {
		t.Fatal(err)
	}
	files := map[string]string{
		".golib.yaml": `schema_version: 1
tool_version: v1.0.0
operations:
  - module: .
    gate: fuzz
    steps:
      - type: go-test
        packages: [./parser]
        fuzz: ^FuzzParse$
      - type: go-test
        packages: [./other]
        fuzz: ^FuzzOther$
`,
		"helper/value.go":  "package helper\nconst Value = 1\n",
		"helper/value.s":   "// native fixture input\n",
		"parser/parser.go": "package parser\nfunc Parse(value string) string { return value }\n",
		"parser/parser_test.go": `package parser
import ("testing"; "example/helper")
func FuzzParse(f *testing.F) {
 f.Add("seed")
 f.Fuzz(func(t *testing.T, input string) { if helper.Value < 1 || Parse(input) != input { t.Fatal("wrong result") } })
}
`,
		"other/other.go": "package other\n",
		"other/other_test.go": `package other
import "testing"
func FuzzOther(f *testing.F) { f.Add("seed"); f.Fuzz(func(t *testing.T, input string) {}) }
`,
	}
	for name, content := range files {
		data := []byte(content)
		if strings.HasSuffix(name, ".go") {
			data, err = format.Source(data)
			if err != nil {
				t.Fatal(err)
			}
		}
		filename := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(filename), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filename, data, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	git := func(args ...string) string {
		command := exec.CommandContext(t.Context(), "git", append([]string{"-c", "core.hooksPath=/dev/null", "-c", "commit.gpgsign=false"}, args...)...)
		command.Dir = root
		command.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_NOSYSTEM=1")
		output, err := command.CombinedOutput()
		if err != nil {
			t.Fatalf("fixture git %s: %v", args[0], err)
		}
		return strings.TrimSpace(string(output))
	}
	git("init", "--quiet")
	git("add", "--", ".golib.yaml", ".go-version", "go.mod", "example.go", "modules.json", "packages.json", "helper", "parser", "other")
	git("-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "commit", "--quiet", "-m", "fixture")
	base := git("rev-parse", "HEAD")
	if err := os.WriteFile(filepath.Join(root, "helper/value.s"), []byte("// changed native fixture input\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	git("add", "--", "helper/value.s")
	git("-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "commit", "--quiet", "-m", "helper")
	var stdout, stderr bytes.Buffer
	code := ExecuteContext(t.Context(), []string{"check", "--local", "--base-sha", base}, root, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("native local fuzz code = %d, output = %q, error = %q", code, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "[.] fuzz-smoke: 1 affected configured targets passed") {
		t.Fatalf("native test dependency did not select exactly one target: %q", stdout.String())
	}
}
