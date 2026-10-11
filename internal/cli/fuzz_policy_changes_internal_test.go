package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/faustbrian/go-library-tools/v2/internal/gates"
)

// Real Git ranges exercise policy admission through the normal CLI. Only the
// external Go/scanner boundary is replaced; no application campaigns run here.
func TestExecuteChangedLocalFuzzPolicyUsesMaterialInputs(t *testing.T) {
	const header = "schema_version: 1\ntool_version: v1.0.0\n"
	const parserStep = "      - type: go-test\n        packages: [./parser]\n        fuzz: ^FuzzParse$\n        budget: 10m\n"
	const otherStep = "      - type: go-test\n        packages: [./other]\n        fuzz: ^FuzzOther$\n        budget: 10m\n"
	const fuzzOperation = "  - module: .\n    gate: fuzz\n    steps:\n"
	const localOperation = "  - module: .\n    gate: local-test\n    steps:\n      - type: go-test\n        packages: [. ]\n        run: ^TestContract$\n        count: 1\n        timeout: 1m\n"
	baseline := header + "operations:\n" + fuzzOperation + parserStep + otherStep
	for _, scenario := range []struct {
		name, before, after, source string
		want                        []string
		refuse                      bool
		catalogChanged              bool
	}{
		{"checksum and local risk only", baseline, header + "tool_checksums_sha256: " + strings.Repeat("a", 64) + "\noperations:\n" + localOperation + fuzzOperation + parserStep + otherStep, "", nil, false, false},
		{"policy plus causal source", baseline, header + "operations:\n" + localOperation + fuzzOperation + parserStep + otherStep, "helper/value.go", []string{"^FuzzParse$"}, false, false},
		{"full budget only", baseline, strings.ReplaceAll(baseline, "budget: 10m", "budget: 20m\n        timeout: 30m\n        count: 5"), "", nil, false, false},
		{"declaration order only", baseline, header + "operations:\n" + fuzzOperation + otherStep + parserStep, "", nil, false, false},
		{"new selector", baseline, strings.ReplaceAll(baseline, "^FuzzParse$", "^FuzzNew$"), "", []string{"^FuzzNew$"}, false, false},
		{"new target", header + "operations:\n" + fuzzOperation + otherStep, baseline, "", []string{"^FuzzParse$"}, false, false},
		{"legacy to typed", header + "operations:\n  - module: .\n    gate: fuzz\n    steps:\n      - type: make\n        makefile: verification/package.mk\n        target: fuzz\n", baseline, "", []string{"^FuzzParse$", "^FuzzOther$"}, false, false},
		{"catalog runtime input", baseline, header + "operations:\n" + localOperation + fuzzOperation + parserStep + otherStep, "", []string{"^FuzzParse$", "^FuzzOther$"}, false, true},
		{"catalog version only", baseline, header + "operations:\n" + localOperation + fuzzOperation + parserStep + otherStep, "", nil, false, true},
		{"catalog version only without policy edit", baseline, baseline, "", nil, false, true},
		{"runtime pin input", baseline, header + "runtimes:\n  deno: 2.6.5\noperations:\n" + fuzzOperation + parserStep + otherStep, "", []string{"^FuzzParse$", "^FuzzOther$"}, false, false},
		{"invalid historical catalog", baseline, baseline, "", nil, true, true},
		{"invalid historical policy", baseline + "unknown_fixture_field: true\n", baseline, "", nil, true, false},
		{"oversized historical policy", baseline + "#" + strings.Repeat("x", 1<<20) + "\n", baseline, "", nil, true, false},
		{"missing historical policy", "", baseline, "", nil, true, false},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			root := internalFixture(t)
			manifestPath := filepath.Join(root, "modules.json")
			manifest, err := os.ReadFile(manifestPath)
			if err != nil {
				t.Fatal(err)
			}
			manifest = bytes.Replace(manifest, []byte(`"gates":{}`), []byte(`"gates":{"fuzz":true}`), 1)
			manifest = bytes.Replace(manifest, []byte(`"releasable":true`), []byte(`"releasable":true,"version":"0.0.0"`), 1)
			initialManifest := manifest
			if scenario.name == "invalid historical catalog" {
				initialManifest = bytes.Replace(manifest, []byte(`"schema_version":1`), []byte(`"schema_version":1,"unknown_fixture_field":true`), 1)
			}
			if err := os.WriteFile(manifestPath, initialManifest, 0o600); err != nil {
				t.Fatal(err)
			}
			policyPath := filepath.Join(root, ".golib.yaml")
			if scenario.before == "" {
				if err := os.Remove(policyPath); err != nil {
					t.Fatal(err)
				}
			} else if err := os.WriteFile(policyPath, []byte(scenario.before), 0o600); err != nil {
				t.Fatal(err)
			}
			git := func(args ...string) string {
				t.Helper()
				command := exec.CommandContext(t.Context(), "git", append([]string{"-c", "core.hooksPath=/dev/null", "-c", "commit.gpgsign=false"}, args...)...)
				command.Dir = root
				command.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_NOSYSTEM=1")
				output, err := command.CombinedOutput()
				if err != nil {
					t.Fatalf("fixture Git %s failed: %v", args[0], err)
				}
				return strings.TrimSpace(string(output))
			}
			git("init", "--quiet")
			git("add", "--", ".go-version", "go.mod", "example.go", "modules.json", "packages.json")
			if scenario.before != "" {
				git("add", "--", ".golib.yaml")
			}
			git("-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "commit", "--quiet", "-m", "fixture")
			base := git("rev-parse", "HEAD")
			if err := os.WriteFile(policyPath, []byte(scenario.after), 0o600); err != nil {
				t.Fatal(err)
			}
			git("add", "--", ".golib.yaml")
			if scenario.catalogChanged {
				changed := bytes.Replace(manifest, []byte(`"gates":{"fuzz":true}`), []byte(`"gates":{"fuzz":true},"test_tags":["fixture"]`), 1)
				if strings.HasPrefix(scenario.name, "catalog version only") {
					changed = bytes.Replace(manifest, []byte(`"version":"0.0.0"`), []byte(`"version":"0.0.1"`), 1)
					if bytes.Equal(changed, manifest) {
						t.Fatal("version fixture did not change")
					}
				}
				if scenario.name == "invalid historical catalog" {
					changed = manifest
				}
				if err := os.WriteFile(manifestPath, changed, 0o600); err != nil {
					t.Fatal(err)
				}
				git("add", "--", "modules.json")
			}
			if scenario.source != "" {
				filename := filepath.Join(root, scenario.source)
				if err := os.MkdirAll(filepath.Dir(filename), 0o700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filename, []byte("package helper\nconst Value = 2\n"), 0o600); err != nil {
					t.Fatal(err)
				}
				git("add", "--", scenario.source)
			}
			git("-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "commit", "--quiet", "-m", "policy")
			var selected []string
			localChecks, secretScans, sessions, cleaned := 0, 0, 0, 0
			fuzz := cliFuzzFixture{root: root, workspace: t.TempDir()}
			factory := func(string, io.Writer, io.Writer) (gates.Executor, func() error, error) {
				sessions++
				return changedSecretsExecutor{directory: fuzz.workspace, run: func(ctx context.Context, command gates.Command) error {
					if handled, err := fuzz.run(ctx, t, command); handled {
						for _, argument := range command.Args {
							if selector, ok := strings.CutPrefix(argument, "-test.fuzz="); ok {
								selected = append(selected, selector)
							}
						}
						return err
					}
					if command.Name == "git" {
						process := exec.CommandContext(ctx, "git", command.Args...)
						process.Dir, process.Stdin, process.Stdout, process.Stderr = command.Dir, command.Stdin, command.Stdout, command.Stderr
						process.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_NOSYSTEM=1")
						return process.Run()
					}
					if command.Name == "gofmt" {
						localChecks++
					}
					if command.Name == "go" && strings.Contains(strings.Join(command.Args, " "), "gitleaks/v8@") {
						secretScans++
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
						return nil
					}
					if command.Name == "go" {
						if slices.Contains(command.Args, "-list=^TestContract$") {
							_, err := io.WriteString(command.Stdout, "TestContract\n")
							return err
						}
						if command.Stdout != nil && slices.Contains(command.Args, "-m") {
							_, err := io.WriteString(command.Stdout, "example\n")
							return err
						}
					}
					return nil
				}}, func() error { cleaned++; return nil }, nil
			}
			var stdout, stderr bytes.Buffer
			code := executeContext(t.Context(), []string{"check", "--local", "--base-sha", base}, root, &stdout, &stderr, factory)
			if (code != 0) != scenario.refuse {
				t.Fatalf("code=%d refuse=%v: %s", code, scenario.refuse, stderr.String())
			}
			if !slices.Equal(selected, scenario.want) {
				t.Fatalf("active targets=%v, want=%v", selected, scenario.want)
			}
			if len(fuzz.compiled) != len(scenario.want) || fuzz.listed != len(scenario.want) {
				t.Fatalf("compile/discovery contexts=%d/%d, want=%d", len(fuzz.compiled), fuzz.listed, len(scenario.want))
			}
			for _, command := range fuzz.compiled {
				if slices.Contains(command.Args, "-tags=fixture") != (scenario.name == "catalog runtime input") {
					t.Fatalf("compile lost material catalog tags: %v", command.Args)
				}
			}
			fuzz.assertCleaned(t)
			if sessions != 1 || cleaned != 1 || localChecks == 0 || secretScans == 0 {
				t.Fatalf("mandatory checks/single-session cleanup missing: local=%d secret=%d sessions=%d/%d", localChecks, secretScans, sessions, cleaned)
			}
		})
	}
}
