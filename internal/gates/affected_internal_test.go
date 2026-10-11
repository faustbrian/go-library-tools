package gates

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/faustbrian/go-library-tools/v2/internal/inventory"
)

func affectedRunner(t *testing.T, changes string) Runner {
	t.Helper()
	root := t.TempDir()
	modules := []inventory.Module{
		{Directory: ".", ModulePath: "example", Packages: []inventory.Package{{Directory: "internal/example"}}},
		{Directory: "adapter", ModulePath: "example/adapter"},
		{Directory: "adapter/nested", ModulePath: "example/nested"},
		{Directory: "consumer", ModulePath: "example/consumer"},
		{Directory: "transitive", ModulePath: "example/transitive"},
		{Directory: "other", ModulePath: "example/other"},
	}
	for _, module := range modules {
		if err := os.MkdirAll(filepath.Join(root, module.Directory), 0o700); err != nil {
			t.Fatal(err)
		}
		dependency := ""
		if module.Directory == "consumer" {
			dependency = "\nrequire example/adapter v1.0.0\n"
		}
		if module.Directory == "transitive" {
			dependency = "\nrequire example/consumer v1.0.0\n"
		}
		if err := os.WriteFile(filepath.Join(root, module.Directory, "go.mod"), []byte("module "+module.ModulePath+"\n\ngo 1.27.0\n"+dependency), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return Runner{Root: root, Catalog: inventory.Inventory{Modules: modules}, Executor: executorFunction(func(_ context.Context, command Command) error {
		if slices.Contains(command.Args, "diff") {
			_, err := io.WriteString(command.Stdout, changes)
			return err
		}
		return nil
	})}
}

func selectedDirectories(modules []inventory.Module) []string {
	var directories []string
	for _, module := range modules {
		directories = append(directories, module.Directory)
	}
	return directories
}

func TestChangedModulesOwnershipAndDependencyClosure(t *testing.T) {
	for _, scenario := range []struct {
		name, paths string
		want        []string
	}{
		{"nested owner", "adapter/nested/source.go\x00", []string{"adapter/nested"}},
		{"actual transitive consumers", "adapter/source.go\x00", []string{"adapter", "consumer", "transitive"}},
		{"root source", "example.go\x00", []string{"."}},
		{"known root package", "internal/example/data.json\x00", []string{"."}},
		{"path boundary", "adapter-other/source.go\x00", nil},
		{"unknown shared policy", ".golib.yaml\x00", nil},
		{"root dependency policy", "go.mod\x00", nil},
		{"empty diff", "", nil},
		{"unusual filename", "adapter/new\nfile.go\x00", []string{"adapter", "consumer", "transitive"}},
		{"both sides of rename", "adapter/old.go\x00other/new.go\x00", []string{"adapter", "consumer", "transitive", "other"}},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			runner := affectedRunner(t, scenario.paths)
			want := scenario.want
			if want == nil {
				want = selectedDirectories(runner.Catalog.Modules)
			}
			modules, err := runner.ChangedModules(t.Context(), strings.Repeat("a", 40))
			if err != nil || !slices.Equal(selectedDirectories(modules), want) {
				t.Fatalf("selection = %v, error = %v, want %v", selectedDirectories(modules), err, want)
			}
		})
	}
}

func TestChangedModulesIncludesDeclaredAndLocalReplacementConsumers(t *testing.T) {
	runner := affectedRunner(t, "adapter/nested/source.go\x00")
	runner.Catalog.Modules[1].OwnedDependencies = []string{"example/nested"}
	data := "module example/other\n\ngo 1.27.0\n\nrequire alias.example/adapter v1.0.0\nreplace alias.example/adapter => ../adapter\n"
	if err := os.WriteFile(filepath.Join(runner.Root, "other", "go.mod"), []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
	modules, err := runner.ChangedModules(t.Context(), strings.Repeat("a", 40))
	if err != nil || !slices.Equal(selectedDirectories(modules), []string{"adapter", "adapter/nested", "consumer", "transitive", "other"}) {
		t.Fatalf("selection = %v, error = %v", selectedDirectories(modules), err)
	}
}

func TestChangedModulesFailsClosedWithoutRawDiagnostics(t *testing.T) {
	for _, scenario := range []string{"missing Git base", "malformed paths", "oversized output", "invalid module", "cancellation", "invalid base"} {
		t.Run(scenario, func(t *testing.T) {
			runner := affectedRunner(t, "adapter/source.go\x00")
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			base := strings.Repeat("a", 40)
			calls := 0
			runner.Executor = executorFunction(func(_ context.Context, command Command) error {
				calls++
				if scenario == "missing Git base" {
					return errors.New("fixture-sensitive-marker")
				}
				if slices.Contains(command.Args, "diff") {
					data := "adapter/source.go\x00"
					if scenario == "malformed paths" {
						data = "adapter/source.go"
					}
					if scenario == "oversized output" {
						data = strings.Repeat("a", maximumSecurityProcessOutput+1)
					}
					_, err := io.WriteString(command.Stdout, data)
					return err
				}
				return nil
			})
			if scenario == "invalid module" {
				if err := os.WriteFile(filepath.Join(runner.Root, "other", "go.mod"), []byte("fixture-sensitive-marker"), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			if scenario == "cancellation" {
				cancel()
			}
			if scenario == "invalid base" {
				base = "HEAD"
			}
			modules, err := runner.ChangedModules(ctx, base)
			if err == nil || len(modules) != 0 || strings.Contains(err.Error(), "fixture-sensitive-marker") {
				t.Fatalf("unexpected selection/diagnostic: %v, %v", selectedDirectories(modules), err)
			}
			if (scenario == "cancellation" || scenario == "invalid base") && calls != 0 {
				t.Fatal("invalid request executed Git")
			}
		})
	}
}

func TestChangedModulesGitCommandsDisableContentExecution(t *testing.T) {
	runner := affectedRunner(t, "adapter/source.go\x00")
	runner.Executor = executorFunction(func(_ context.Context, command Command) error {
		if command.Name != "git" || !command.boundedScanner || !slices.Contains(command.Args, "--no-replace-objects") || !slices.Contains(command.Args, "core.fsmonitor=false") {
			return fmt.Errorf("Git trust options absent")
		}
		if slices.Contains(command.Args, "diff") {
			for _, option := range []string{"--no-ext-diff", "--no-textconv", "--no-renames", "-z"} {
				if !slices.Contains(command.Args, option) {
					return fmt.Errorf("Git diff option absent")
				}
			}
			_, err := io.WriteString(command.Stdout, "adapter/source.go\x00")
			return err
		}
		return nil
	})
	if _, err := runner.ChangedModules(t.Context(), strings.Repeat("a", 40)); err != nil {
		t.Fatal(err)
	}
}
