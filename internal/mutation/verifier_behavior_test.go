//go:build verifierintegration

package mutation

import (
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"
)

func TestPatchedVerifierRenewsTimeoutForEachIntegrationPhase(t *testing.T) {
	workspace := t.TempDir()
	t.Cleanup(func() {
		_ = filepath.WalkDir(workspace, func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			info, err := entry.Info()
			if err == nil {
				_ = os.Chmod(path, info.Mode().Perm()|0o200)
			}
			return nil
		})
	})
	tool, err := BuildVerifier(context.Background(), filepath.Join(workspace, "verifier"), integrationProcess)
	if err != nil {
		t.Fatalf("BuildVerifier() error = %v", err)
	}

	root := filepath.Join(workspace, "fixture")
	if err := os.MkdirAll(root, 0o700); err != nil {
		t.Fatal(err)
	}
	for name, content := range map[string]string{
		"go.mod":       "module example.test/verifier\n\ngo 1.27.0\n",
		"sign.go":      "package verifier\n\nfunc Sign(value int) int {\n\tif value > 0 {\n\t\treturn 1\n\t}\n\treturn 0\n}\n",
		"sign_test.go": "package verifier\n\nimport (\n\t\"testing\"\n\t\"time\"\n)\n\nfunc TestSign(t *testing.T) {\n\ttime.Sleep(1800 * time.Millisecond)\n\tif Sign(1) != 1 || Sign(0) != 0 || Sign(-1) != 0 {\n\t\tt.Fatal(\"unexpected sign\")\n\t}\n}\n",
	} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	for _, arguments := range [][]string{{"init", "-q"}, {"add", "go.mod", "sign.go", "sign_test.go"}, {"-c", "user.name=golib-test", "-c", "user.email=golib-test@example.invalid", "commit", "-qm", "fixture"}} {
		if err := integrationProcess(context.Background(), "git", arguments, root, nil, io.Discard, io.Discard); err != nil {
			t.Fatalf("git %s: %v", strings.Join(arguments, " "), err)
		}
	}

	cache := filepath.Join(workspace, "execution")
	environment := map[string]string{
		"GOCACHE":    filepath.Join(cache, "build"),
		"GOMODCACHE": filepath.Join(cache, "mod"),
		"GOTMPDIR":   filepath.Join(cache, "tmp"),
		"GOWORK":     "off",
	}
	for _, directory := range []string{environment["GOCACHE"], environment["GOMODCACHE"], environment["GOTMPDIR"]} {
		if err := os.MkdirAll(directory, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	profile := filepath.Join(workspace, "coverage.out")
	if err := integrationProcess(context.Background(), "go", []string{"test", "-count=1", "-coverprofile=" + profile, "./..."}, root, environment, io.Discard, io.Discard); err != nil {
		t.Fatalf("coverage baseline: %v", err)
	}
	arguments, err := Arguments(".", filepath.Join(workspace, "report.json"), "", true, 1)
	if err != nil {
		t.Fatal(err)
	}
	environment["GOLIB_GREMLINS_COVERAGE_PROFILE"] = profile
	environment["GOLIB_GREMLINS_COVERAGE_ELAPSED"] = "3s"
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	if err := integrationProcess(ctx, tool.Path, arguments, root, environment, io.Discard, io.Discard); err != nil {
		t.Fatalf("patched verifier did not give each phase an independent timeout: %v", err)
	}
}

func integrationProcess(ctx context.Context, name string, arguments []string, directory string, environment map[string]string, stdout, stderr io.Writer) error {
	command := exec.CommandContext(ctx, name, arguments...)
	command.Dir = directory
	values := make(map[string]string)
	for _, value := range os.Environ() {
		name, item, ok := strings.Cut(value, "=")
		if ok {
			values[name] = item
		}
	}
	for name, value := range environment {
		values[name] = value
	}
	keys := make([]string, 0, len(values))
	for name := range values {
		keys = append(keys, name)
	}
	sort.Strings(keys)
	command.Env = make([]string, 0, len(keys))
	for _, name := range keys {
		command.Env = append(command.Env, name+"="+values[name])
	}
	command.Stdout = stdout
	command.Stderr = stderr
	return command.Run()
}
