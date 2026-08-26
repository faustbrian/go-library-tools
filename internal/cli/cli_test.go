package cli_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/faustbrian/go-library-tools/internal/cli"
)

func TestExecuteShowsHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := cli.Execute([]string{"--help"}, t.TempDir(), &stdout, &stderr)
	if code != 0 || stderr.Len() != 0 {
		t.Fatalf("Execute() code = %d, stderr = %q", code, stderr.String())
	}
	for _, command := range []string{"check", "config validate", "config show --json", "inventory", "repository check", "release dry-run"} {
		if !strings.Contains(stdout.String(), command) {
			t.Errorf("help does not contain %q", command)
		}
	}
}

func TestExecuteShowsVersionWithoutRepository(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := cli.Execute([]string{"--version"}, t.TempDir(), &stdout, &stderr)
	if code != 0 || stderr.Len() != 0 || stdout.String() != "dev\n" {
		t.Fatalf("Execute() code = %d, stdout = %q, stderr = %q", code, stdout.String(), stderr.String())
	}
}

func TestExecuteValidatesConfigurationFromNestedDirectory(t *testing.T) {
	root := fixture(t)
	nested := filepath.Join(root, "nested", "directory")
	if err := os.MkdirAll(nested, 0o700); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	code := cli.Execute([]string{"config", "validate"}, nested, &stdout, &stderr)
	if code != 0 || stderr.Len() != 0 {
		t.Fatalf("Execute() code = %d, stderr = %q", code, stderr.String())
	}
	if got := stdout.String(); got != "configuration valid: github.com/faustbrian/example\n" {
		t.Fatalf("Execute() stdout = %q", got)
	}
}

func TestExecuteShowsMachineReadableConfiguration(t *testing.T) {
	root := fixture(t)
	write(t, filepath.Join(root, ".golib.yaml"), "schema_version: 1\ntool_version: v1.0.0\nruntimes:\n  deno: 2.9.4\n")
	var stdout, stderr bytes.Buffer
	code := cli.Execute([]string{"config", "show", "--json"}, root, &stdout, &stderr)
	if code != 0 || stderr.Len() != 0 {
		t.Fatalf("Execute() code = %d, stderr = %q", code, stderr.String())
	}
	var got map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	runtimes, ok := got["runtimes"].(map[string]any)
	if !ok || runtimes["deno"] != "2.9.4" {
		t.Fatalf("configuration = %#v", got)
	}
}

func TestExecutePrintsDeterministicJSONInventory(t *testing.T) {
	root := fixture(t)
	var stdout, stderr bytes.Buffer
	code := cli.Execute([]string{"inventory", "--json"}, root, &stdout, &stderr)
	if code != 0 || stderr.Len() != 0 {
		t.Fatalf("Execute() code = %d, stderr = %q", code, stderr.String())
	}
	var got map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("stdout is not JSON: %v", err)
	}
	if got["repository"] != "github.com/faustbrian/example" {
		t.Fatalf("inventory repository = %#v", got["repository"])
	}
}

func TestExecutePrintsHumanInventory(t *testing.T) {
	root := fixture(t)
	var stdout, stderr bytes.Buffer
	code := cli.Execute([]string{"inventory"}, root, &stdout, &stderr)
	if code != 0 || stderr.Len() != 0 || stdout.String() != "github.com/faustbrian/example: 1 module(s)\n" {
		t.Fatalf("Execute() code = %d, stdout = %q, stderr = %q", code, stdout.String(), stderr.String())
	}
}

func TestExecuteChecksStandaloneRepositoryContract(t *testing.T) {
	root := fixture(t)
	write(t, filepath.Join(root, "go.mod"), "module github.com/faustbrian/example\n\ngo 1.27.0\n")
	var stdout, stderr bytes.Buffer
	code := cli.Execute([]string{"repository", "check"}, root, &stdout, &stderr)
	if code != 0 || stderr.Len() != 0 || stdout.String() != "standalone repository contract passed\n" {
		t.Fatalf("Execute() = %d, %q, %q", code, stdout.String(), stderr.String())
	}
}

func TestExecuteInspectsEvidence(t *testing.T) {
	root := fixture(t)
	for _, args := range [][]string{{"evidence", "inspect"}, {"evidence", "inspect", "--json"}} {
		var stdout, stderr bytes.Buffer
		code := cli.Execute(args, root, &stdout, &stderr)
		if code != 0 || stderr.Len() != 0 || stdout.Len() == 0 {
			t.Fatalf("Execute(%v) = %d, %q, %q", args, code, stdout.String(), stderr.String())
		}
	}
}

func TestExecuteRunsCheckWithDisposableGoEnvironment(t *testing.T) {
	root := fixture(t)
	write(t, filepath.Join(root, "go.mod"), "module github.com/faustbrian/example\n\ngo 1.27.0\n")
	write(t, filepath.Join(root, "example.go"), "package example\n\nfunc Value() int { return 1 }\n")
	write(t, filepath.Join(root, "modules.json"), `{"schema_version":1,"repository":"github.com/faustbrian/example","go_version":"1.27.0","modules":[{"directory":".","module_path":"github.com/faustbrian/example","go_version":"1.27.0","kind":"public","releasable":true,"gates":{},"packages":[]}]}`)
	var stdout, stderr bytes.Buffer
	code := cli.Execute([]string{"check", "--module", "."}, root, &stdout, &stderr)
	if code != 0 || stderr.Len() != 0 {
		t.Fatalf("Execute() code = %d, stdout = %q, stderr = %q", code, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "[.] safety\n") {
		t.Fatalf("check output = %q", stdout.String())
	}
}

func TestExecuteCheckDefaultsToAllModules(t *testing.T) {
	root := fixture(t)
	write(t, filepath.Join(root, "go.mod"), "module github.com/faustbrian/example\n\ngo 1.27.0\n")
	write(t, filepath.Join(root, "example.go"), "package example\n")
	write(t, filepath.Join(root, "modules.json"), `{"schema_version":1,"repository":"github.com/faustbrian/example","go_version":"1.27.0","modules":[{"directory":".","module_path":"github.com/faustbrian/example","go_version":"1.27.0","kind":"public","releasable":true,"gates":{},"packages":[]}]}`)
	var stdout, stderr bytes.Buffer
	code := cli.Execute([]string{"check"}, root, &stdout, &stderr)
	if code != 0 || stderr.Len() != 0 {
		t.Fatalf("Execute() code = %d, stderr = %q", code, stderr.String())
	}
}

func TestExecuteReportsOutputFailure(t *testing.T) {
	root := fixture(t)
	var stderr bytes.Buffer
	code := cli.Execute([]string{"config", "show", "--json"}, root, failingWriter{}, &stderr)
	if code != 1 || !strings.Contains(stderr.String(), "write configuration") {
		t.Fatalf("Execute() config code = %d, stderr = %q", code, stderr.String())
	}
	stderr.Reset()
	code = cli.Execute([]string{"inventory", "--json"}, root, failingWriter{}, &stderr)
	if code != 1 || !strings.Contains(stderr.String(), "write inventory") {
		t.Fatalf("Execute() code = %d, stderr = %q", code, stderr.String())
	}
	stderr.Reset()
	code = cli.Execute([]string{"evidence", "inspect", "--json"}, root, failingWriter{}, &stderr)
	if code != 1 || !strings.Contains(stderr.String(), "write evidence inventory") {
		t.Fatalf("Execute() evidence code = %d, stderr = %q", code, stderr.String())
	}
}

func TestExecuteReportsUsageAndRepositoryErrors(t *testing.T) {
	tests := []struct {
		name string
		args []string
		root func(*testing.T) string
		code int
		want string
	}{
		{"missing command", nil, fixture, 2, "command is required"},
		{"unknown command", []string{"wat"}, fixture, 2, "unknown command"},
		{"invalid config arguments", []string{"config"}, fixture, 2, "usage:"},
		{"invalid config show arguments", []string{"config", "show"}, fixture, 2, "usage:"},
		{"invalid inventory arguments", []string{"inventory", "--yaml"}, fixture, 2, "usage:"},
		{"invalid repository arguments", []string{"repository"}, fixture, 2, "usage:"},
		{"invalid evidence arguments", []string{"evidence", "inspect", "--yaml"}, fixture, 2, "usage:"},
		{"unsafe evidence", []string{"evidence", "inspect"}, func(t *testing.T) string {
			root := fixture(t)
			target := filepath.Join(root, "target")
			if err := os.Mkdir(target, 0o700); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(target, filepath.Join(root, ".verification")); err != nil {
				t.Fatal(err)
			}
			return root
		}, 1, "evidence root"},
		{"invalid repository contract", []string{"repository", "check"}, func(t *testing.T) string {
			root := fixture(t)
			write(t, filepath.Join(root, ".go-version"), "1.26.0\n")
			return root
		}, 1, ".go-version"},
		{"invalid check arguments", []string{"check", "--module"}, fixture, 2, "usage:"},
		{"unknown check module", []string{"check", "--module", "missing"}, fixture, 1, "unknown module"},
		{"missing root", []string{"inventory"}, func(t *testing.T) string { return t.TempDir() }, 1, "locate repository root"},
		{"relative root", []string{"inventory"}, func(*testing.T) string { return "." }, 1, "must be absolute"},
		{"invalid root path", []string{"inventory"}, func(t *testing.T) string {
			root := t.TempDir()
			file := filepath.Join(root, "file")
			write(t, file, "not a directory")
			return file
		}, 1, "locate repository root"},
		{"invalid config", []string{"inventory"}, func(t *testing.T) string {
			root := fixture(t)
			write(t, filepath.Join(root, ".golib.yaml"), "schema_version: 2\ntool_version: v1.0.0\n")
			return root
		}, 1, "schema_version must be 1"},
		{"invalid manifest", []string{"inventory"}, func(t *testing.T) string {
			root := fixture(t)
			write(t, filepath.Join(root, "modules.json"), "{")
			return root
		}, 1, "load module manifest"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := cli.Execute(test.args, test.root(t), &stdout, &stderr)
			if code != test.code || !strings.Contains(stderr.String(), test.want) {
				t.Fatalf("Execute() code = %d, stderr = %q", code, stderr.String())
			}
		})
	}
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) {
	return 0, errors.New("write failed")
}

func fixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	write(t, filepath.Join(root, ".golib.yaml"), "schema_version: 1\ntool_version: v1.0.0\n")
	write(t, filepath.Join(root, ".go-version"), "1.27.0\n")
	write(t, filepath.Join(root, "modules.json"), `{"schema_version":1,"repository":"github.com/faustbrian/example","go_version":"1.27.0","modules":[{"directory":".","module_path":"github.com/faustbrian/example","go_version":"1.27.0","kind":"public","releasable":true,"gates":{"tests":true},"packages":[]}]}`)
	write(t, filepath.Join(root, "packages.json"), `{"schema_version":1,"repository":"github.com/faustbrian/example","packages":[]}`)
	return root
}

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}
