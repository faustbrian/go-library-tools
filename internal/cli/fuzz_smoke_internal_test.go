package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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
	workspace := t.TempDir()
	fuzz := cliFuzzFixture{root: root, workspace: workspace}
	factory := func(string, io.Writer, io.Writer) (gates.Executor, func() error, error) {
		created++
		executor := changedLocalCLIExecutor(t, func(ctx context.Context, command gates.Command) error {
			if handled, err := fuzz.run(ctx, t, command); handled {
				return err
			}
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
			return nil
		})
		executor.directory = workspace
		return executor, func() error { cleaned++; return nil }, nil
	}
	var stdout, stderr bytes.Buffer
	code := execute([]string{"check", "--local", "--base-sha", strings.Repeat("a", 40), "--module", "."}, root, &stdout, &stderr, factory)
	if code != 0 {
		t.Fatalf("changed local code = %d, diagnostic = %q", code, stderr.String())
	}
	if created != 1 || cleaned != 1 || len(fuzz.active) != 1 {
		t.Fatalf("sessions = %d/%d, fuzz commands = %d", created, cleaned, len(fuzz.active))
	}
	if len(fuzz.compiled) != 1 || fuzz.listed != 1 || !slices.Contains(fuzz.compiled[0].Args, "./parser") || !slices.Contains(fuzz.active[0].Args, "-test.fuzz=^FuzzParse$") {
		t.Fatalf("wrong selected compile/discovery/campaign: compiled=%v listed=%d active=%v", fuzz.compiled, fuzz.listed, fuzz.active)
	}
	fuzz.assertCleaned(t)
}

// The fake compiler creates its requested output so production validation still
// checks a real, owned executable file. The fake binary then models native test
// discovery and exploration separately; no placeholder is executed.
type cliFuzzFixture struct {
	root, workspace  string
	compiled, active []gates.Command
	binaries         map[string]string
	listed           int
}

func (fixture *cliFuzzFixture) run(ctx context.Context, t *testing.T, command gates.Command) (bool, error) {
	t.Helper()
	if command.Name == "go" && slices.Contains(command.Args, "-c") {
		for _, argument := range []string{"test", "-fuzz=^Fuzz", "-p=1"} {
			if !slices.Contains(command.Args, argument) {
				t.Fatalf("compile missing %q: %v", argument, command.Args)
			}
		}
		index := slices.Index(command.Args, "-o")
		if index < 0 || index+1 >= len(command.Args) || command.Dir != fixture.root || command.Env["GOWORK"] != "off" {
			t.Fatalf("invalid compile output/module context: %+v", command)
		}
		binary := command.Args[index+1]
		fixture.assertOwned(t, binary)
		if err := os.WriteFile(binary, []byte("fixture executable\n"), 0o700); err != nil {
			t.Fatal(err)
		}
		if fixture.binaries == nil {
			fixture.binaries = make(map[string]string)
		}
		fixture.binaries[binary] = command.Args[len(command.Args)-1]
		fixture.compiled = append(fixture.compiled, command)
		return true, nil
	}
	packageName, exists := fixture.binaries[command.Name]
	if !exists {
		return false, nil
	}
	if _, ok := ctx.Deadline(); !ok {
		t.Fatal("native test execution has no wall deadline")
	}
	directory := filepath.Join(fixture.root, strings.TrimPrefix(packageName, "./"))
	if command.Dir != directory || command.Env["PWD"] != directory || command.Env["GOWORK"] != "off" {
		t.Fatalf("native binary lost package context: %+v", command)
	}
	for _, argument := range []string{"-test.run=^$", "-test.count=1", "-test.timeout=15s", "-test.paniconexit0"} {
		if !slices.Contains(command.Args, argument) {
			t.Fatalf("native test missing %q: %v", argument, command.Args)
		}
	}
	if slices.Contains(command.Args, "-test.list=^Fuzz") {
		fixture.listed++
		names := "FuzzParse\nFuzzNew\n"
		if packageName == "./other" {
			names = "FuzzOther\n"
		}
		_, err := io.WriteString(command.Stdout, names)
		return true, err
	}
	for _, argument := range []string{"-test.fuzztime=2s", "-test.parallel=2"} {
		if !slices.Contains(command.Args, argument) {
			t.Fatalf("active fuzz missing %q: %v", argument, command.Args)
		}
	}
	cacheFound, selectorFound := false, false
	for _, argument := range command.Args {
		if cache, ok := strings.CutPrefix(argument, "-test.fuzzcachedir="); ok {
			fixture.assertOwned(t, cache)
			if info, err := os.Stat(cache); err != nil || !info.IsDir() {
				t.Fatal("native fuzz corpus directory missing")
			}
			cacheFound = true
		}
		selectorFound = selectorFound || strings.HasPrefix(argument, "-test.fuzz=")
	}
	if !cacheFound || !selectorFound {
		t.Fatal("native fuzz lost selector or owned corpus")
	}
	fixture.active = append(fixture.active, command)
	return true, nil
}

func (fixture *cliFuzzFixture) assertOwned(t *testing.T, filename string) {
	t.Helper()
	relative, err := filepath.Rel(fixture.workspace, filename)
	if err != nil || !filepath.IsAbs(filename) || relative == "." || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		t.Fatalf("fuzz resource escaped owned workspace: %q", filename)
	}
}

func (fixture *cliFuzzFixture) assertCleaned(t *testing.T) {
	t.Helper()
	entries, err := os.ReadDir(fixture.workspace)
	if err != nil || len(entries) != 0 {
		t.Fatalf("fuzz preparation resources remain after session: %v, %v", entries, err)
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
