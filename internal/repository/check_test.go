package repository_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/faustbrian/go-library-tools/internal/inventory"
	"github.com/faustbrian/go-library-tools/internal/repository"
)

func TestCheckAcceptsStandaloneAndMultiModuleRepositories(t *testing.T) {
	root, catalog := fixture(t)
	if err := repository.Check(root, catalog); err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	nested := inventory.Module{Directory: "nested", ModulePath: "example/nested", GoVersion: "1.27.0"}
	write(t, filepath.Join(root, "nested", "go.mod"), "module example/nested\n\ngo 1.27.0\n")
	write(t, filepath.Join(root, "go.work"), "go 1.27.0\n\nuse (\n\t.\n\t./nested\n)\n")
	catalog.Modules = append(catalog.Modules, nested)
	if err := repository.Check(root, catalog); err != nil {
		t.Fatalf("Check() multi-module error = %v", err)
	}
}

func TestCheckRejectsRepositoryContractViolations(t *testing.T) {
	tests := map[string]func(*testing.T, string, *inventory.Inventory){
		"relative root": func(_ *testing.T, _ string, catalog *inventory.Inventory) { catalog.Repository = "relative-root" },
		"version mismatch": func(t *testing.T, root string, _ *inventory.Inventory) {
			write(t, filepath.Join(root, ".go-version"), "1.26.0\n")
		},
		"missing version": func(t *testing.T, root string, _ *inventory.Inventory) { remove(t, filepath.Join(root, ".go-version")) },
		"duplicate module": func(_ *testing.T, _ string, catalog *inventory.Inventory) {
			catalog.Modules = append(catalog.Modules, catalog.Modules[0])
		},
		"legacy tooling": func(t *testing.T, root string, _ *inventory.Inventory) {
			if err := os.Mkdir(filepath.Join(root, ".golib"), 0o700); err != nil {
				t.Fatal(err)
			}
		},
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			root, catalog := fixture(t)
			mutate(t, root, &catalog)
			checkRoot := root
			if name == "relative root" {
				checkRoot = "."
			}
			if err := repository.Check(checkRoot, catalog); err == nil {
				t.Fatal("Check() error = nil")
			}
		})
	}
}

func TestCheckRejectsInvalidModules(t *testing.T) {
	tests := map[string]func(*testing.T, string, *inventory.Module){
		"empty directory":     func(_ *testing.T, _ string, module *inventory.Module) { module.Directory = "" },
		"absolute directory":  func(_ *testing.T, _ string, module *inventory.Module) { module.Directory = "/tmp" },
		"backslash directory": func(_ *testing.T, _ string, module *inventory.Module) { module.Directory = `nested\module` },
		"unclean directory":   func(_ *testing.T, _ string, module *inventory.Module) { module.Directory = "nested/.." },
		"outside namespace":   func(_ *testing.T, _ string, module *inventory.Module) { module.ModulePath = "other/module" },
		"missing go.mod":      func(t *testing.T, root string, _ *inventory.Module) { remove(t, filepath.Join(root, "go.mod")) },
		"malformed go.mod": func(t *testing.T, root string, _ *inventory.Module) {
			write(t, filepath.Join(root, "go.mod"), "module\n")
		},
		"module mismatch": func(t *testing.T, root string, _ *inventory.Module) {
			write(t, filepath.Join(root, "go.mod"), "module example/other\n\ngo 1.27.0\n")
		},
		"missing go version": func(t *testing.T, root string, _ *inventory.Module) {
			write(t, filepath.Join(root, "go.mod"), "module example\n")
		},
		"go version mismatch": func(_ *testing.T, _ string, module *inventory.Module) { module.GoVersion = "1.26.0" },
		"replace directive": func(t *testing.T, root string, _ *inventory.Module) {
			write(t, filepath.Join(root, "go.mod"), "module example\n\ngo 1.27.0\nreplace example/old => ./old\n")
		},
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			root, catalog := fixture(t)
			mutate(t, root, &catalog.Modules[0])
			if err := repository.Check(root, catalog); err == nil {
				t.Fatal("Check() error = nil")
			}
		})
	}
}

func TestCheckRejectsInvalidWorkspaces(t *testing.T) {
	tests := map[string]string{
		"missing":          "",
		"malformed":        "go [",
		"version mismatch": "go 1.26.0\nuse (\n.\n./nested\n)\n",
		"non-local":        "go 1.27.0\nuse ../other\n",
		"root only":        "go 1.27.0\nuse .\n",
		"mismatch":         "go 1.27.0\nuse ./other\n",
	}
	for name, workspace := range tests {
		t.Run(name, func(t *testing.T) {
			root, catalog := fixture(t)
			catalog.Modules = append(catalog.Modules, inventory.Module{Directory: "nested", ModulePath: "example/nested", GoVersion: "1.27.0"})
			write(t, filepath.Join(root, "nested", "go.mod"), "module example/nested\n\ngo 1.27.0\n")
			if workspace != "" {
				write(t, filepath.Join(root, "go.work"), workspace)
			}
			if err := repository.Check(root, catalog); err == nil {
				t.Fatal("Check() error = nil")
			}
		})
	}
}

func TestCheckRejectsUnreadableWorkspace(t *testing.T) {
	root, catalog := fixture(t)
	if err := os.Symlink("missing", filepath.Join(root, "go.work")); err != nil {
		t.Fatal(err)
	}
	if err := repository.Check(root, catalog); err == nil || !strings.Contains(err.Error(), "read go.work") {
		t.Fatalf("Check() error = %v", err)
	}
}

func fixture(t *testing.T) (string, inventory.Inventory) {
	t.Helper()
	root := t.TempDir()
	write(t, filepath.Join(root, ".go-version"), "1.27.0\n")
	write(t, filepath.Join(root, "go.mod"), "module example\n\ngo 1.27.0\n")
	return root, inventory.Inventory{Repository: "example", GoVersion: "1.27.0", Modules: []inventory.Module{{Directory: ".", ModulePath: "example", GoVersion: "1.27.0"}}}
}

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func remove(t *testing.T, path string) {
	t.Helper()
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		t.Fatal(err)
	}
}
