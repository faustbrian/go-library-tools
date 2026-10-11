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
	"github.com/faustbrian/go-library-tools/v2/internal/inventory"
)

func affectedCLIFixture(t *testing.T) (string, string) {
	t.Helper()
	root := internalFixture(t)
	modules := []inventory.Module{
		{Directory: ".", ModulePath: "example", GoVersion: "1.27.0", Kind: "public", Gates: map[string]bool{"tests": true}},
		{Directory: "adapter", ModulePath: "example/adapter", GoVersion: "1.27.0", Kind: "public", Gates: map[string]bool{"tests": true}},
		{Directory: "other", ModulePath: "example/other", GoVersion: "1.27.0", Kind: "public", Gates: map[string]bool{"tests": true}},
		{Directory: "consumer", ModulePath: "example/consumer", GoVersion: "1.27.0", Kind: "public", Gates: map[string]bool{"tests": true}, OwnedDependencies: []string{"example/adapter"}},
	}
	for _, module := range modules[1:] {
		if err := os.MkdirAll(filepath.Join(root, module.Directory), 0o700); err != nil {
			t.Fatal(err)
		}
		for name, content := range map[string]string{"go.mod": "module " + module.ModulePath + "\n\ngo 1.27.0\n", "example.go": "package example\n"} {
			if err := os.WriteFile(filepath.Join(root, module.Directory, name), []byte(content), 0o600); err != nil {
				t.Fatal(err)
			}
		}
	}
	manifest, err := json.Marshal(inventory.Inventory{SchemaVersion: 1, Repository: "example", GoVersion: "1.27.0", Modules: modules})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "modules.json"), manifest, 0o600); err != nil {
		t.Fatal(err)
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
	git("add", "--", ".golib.yaml", ".go-version", "go.mod", "example.go", "modules.json", "packages.json", "adapter", "other", "consumer")
	git("-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "commit", "--quiet", "-m", "fixture")
	base := git("rev-parse", "HEAD")
	if err := os.WriteFile(filepath.Join(root, "adapter", "example.go"), []byte("package example\n\nconst Value = 1\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	git("add", "--", "adapter/example.go")
	git("-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "commit", "--quiet", "-m", "adapter")
	return root, base
}

func TestExecuteAffectedInventoryUsesCommittedModuleAndConsumers(t *testing.T) {
	root, base := affectedCLIFixture(t)
	var stdout, stderr bytes.Buffer
	if code := ExecuteContext(t.Context(), []string{"inventory", "--json", "--base-sha", base}, root, &stdout, &stderr); code != 0 {
		t.Fatalf("affected inventory code = %d, diagnostic = %q", code, stderr.String())
	}
	var catalog inventory.Inventory
	if err := json.Unmarshal(stdout.Bytes(), &catalog); err != nil {
		t.Fatal(err)
	}
	directories := make([]string, 0, len(catalog.Modules))
	for _, module := range catalog.Modules {
		directories = append(directories, module.Directory)
	}
	if !slices.Equal(directories, []string{"adapter", "consumer"}) {
		t.Fatalf("affected modules = %v", directories)
	}
}

func TestExecuteRepeatedLocalModulesRunOnceInOneSession(t *testing.T) {
	root, _ := affectedCLIFixture(t)
	var tested []string
	created, cleaned := 0, 0
	factory := func(string, io.Writer, io.Writer) (gates.Executor, func() error, error) {
		created++
		return cliExecutorFunction(func(_ context.Context, command gates.Command) error {
			if command.Name == "go" && slices.Contains(command.Args, "test") {
				directory, err := filepath.Rel(root, command.Dir)
				if err != nil {
					return err
				}
				tested = append(tested, directory)
			}
			return nil
		}), func() error { cleaned++; return nil }, nil
	}
	var stdout, stderr bytes.Buffer
	if code := execute([]string{"check", "--local", "--module", "adapter", "--module", "consumer", "--module", "adapter"}, root, &stdout, &stderr, factory); code != 0 {
		t.Fatalf("repeated local check code = %d, diagnostic = %q", code, stderr.String())
	}
	if created != 1 || cleaned != 1 || !slices.Equal(tested, []string{"adapter", "consumer"}) {
		t.Fatalf("sessions = %d/%d, tested modules = %v", created, cleaned, tested)
	}
}

func TestExecuteRepeatedLocalModulesRejectsMixedAndExcessiveArguments(t *testing.T) {
	root, _ := affectedCLIFixture(t)
	for _, arguments := range [][]string{
		{"--module", "adapter", "--all"},
		{"--module", "adapter", "--module", ""},
		slices.Repeat([]string{"--module", "adapter"}, inventory.MaximumModules+1),
	} {
		var stderr bytes.Buffer
		if code := execute(append([]string{"check", "--local"}, arguments...), root, io.Discard, &stderr, nil); code != 2 {
			t.Fatalf("invalid repeated selection code = %d", code)
		}
	}
}

func TestExecuteAffectedInventoryHandlesActualRenameAndMissingBase(t *testing.T) {
	root, _ := affectedCLIFixture(t)
	git := func(args ...string) string {
		command := exec.CommandContext(t.Context(), "git", append([]string{"-c", "core.hooksPath=/dev/null", "-c", "commit.gpgsign=false"}, args...)...)
		command.Dir = root
		command.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_NOSYSTEM=1")
		data, err := command.CombinedOutput()
		if err != nil {
			t.Fatalf("fixture Git failed: %v", err)
		}
		return strings.TrimSpace(string(data))
	}
	base := git("rev-parse", "HEAD")
	if err := os.Rename(filepath.Join(root, "adapter", "example.go"), filepath.Join(root, "other", "renamed\nfile.go")); err != nil {
		t.Fatal(err)
	}
	git("add", "--", "adapter/example.go", "other/renamed\nfile.go")
	git("-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "commit", "--quiet", "-m", "rename")
	var stdout, stderr bytes.Buffer
	if code := ExecuteContext(t.Context(), []string{"inventory", "--json", "--base-sha", base}, root, &stdout, &stderr); code != 0 {
		t.Fatalf("rename inventory code = %d, %q", code, stderr.String())
	}
	var catalog inventory.Inventory
	if err := json.Unmarshal(stdout.Bytes(), &catalog); err != nil {
		t.Fatal(err)
	}
	directories := make([]string, 0, len(catalog.Modules))
	for _, module := range catalog.Modules {
		directories = append(directories, module.Directory)
	}
	if !slices.Equal(directories, []string{"adapter", "other", "consumer"}) {
		t.Fatalf("rename owners = %v", directories)
	}
	stdout.Reset()
	stderr.Reset()
	if code := ExecuteContext(t.Context(), []string{"inventory", "--json", "--base-sha", strings.Repeat("f", 40)}, root, &stdout, &stderr); code != 1 || stdout.Len() != 0 {
		t.Fatalf("missing base code = %d, output = %q", code, stdout.String())
	}
}
