// Package repository validates the standalone repository contract.
package repository

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/faustbrian/go-library-tools/internal/inventory"
	"github.com/faustbrian/go-library-tools/internal/repositoryfile"
	"golang.org/x/mod/modfile"
)

const maximumModuleFileSize = 1 << 20

// Check validates module identities, versions, workspace membership, and the
// absence of the copied legacy implementation.
func Check(root string, catalog inventory.Inventory) error {
	if !filepath.IsAbs(root) {
		return errors.New("repository root must be absolute")
	}
	version, err := repositoryfile.Read(root, ".go-version", maximumModuleFileSize)
	if err != nil {
		return fmt.Errorf("read .go-version: %w", err)
	}
	if strings.TrimSpace(string(version)) != catalog.GoVersion {
		return fmt.Errorf(".go-version does not match catalog Go version %q", catalog.GoVersion)
	}
	expectedWorkspace := make([]string, 0, len(catalog.Modules))
	seen := make(map[string]struct{}, len(catalog.Modules))
	for _, module := range catalog.Modules {
		if err := checkModule(root, catalog.Repository, module); err != nil {
			return err
		}
		if _, exists := seen[module.Directory]; exists {
			return fmt.Errorf("duplicate module directory: %s", module.Directory)
		}
		seen[module.Directory] = struct{}{}
		expectedWorkspace = append(expectedWorkspace, module.Directory)
	}
	slices.Sort(expectedWorkspace)
	if err := checkWorkspace(root, catalog.GoVersion, expectedWorkspace); err != nil {
		return err
	}
	if err := checkLegacy(root, os.Lstat); err != nil {
		return err
	}
	return nil
}

func checkLegacy(root string, lstat func(string) (os.FileInfo, error)) error {
	if _, err := lstat(filepath.Join(root, ".golib")); err == nil {
		return errors.New("legacy .golib implementation remains")
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("inspect legacy .golib path: %w", err)
	}
	return nil
}

func checkModule(root, repository string, module inventory.Module) error {
	validDirectory := [3]bool{module.Directory != "", !filepath.IsAbs(module.Directory), !strings.Contains(module.Directory, "\\")}
	if validDirectory != [3]bool{true, true, true} {
		return fmt.Errorf("invalid module directory: %q", module.Directory)
	}
	clean := filepath.Clean(module.Directory)
	if clean != module.Directory || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return fmt.Errorf("invalid module directory: %q", module.Directory)
	}
	inNamespace := module.ModulePath == repository
	if strings.HasPrefix(module.ModulePath, repository+"/") {
		inNamespace = true
	}
	if !inNamespace {
		return fmt.Errorf("module %s is outside repository namespace %s", module.ModulePath, repository)
	}
	relative := filepath.Join(module.Directory, "go.mod")
	data, err := repositoryfile.Read(root, relative, maximumModuleFileSize)
	if err != nil {
		return fmt.Errorf("read %s: %w", relative, err)
	}
	parsed, err := modfile.Parse(relative, data, nil)
	if err != nil {
		return fmt.Errorf("parse %s: %w", relative, err)
	}
	if parsed.Module == nil || parsed.Module.Mod.Path != module.ModulePath {
		return fmt.Errorf("%s module path does not match catalog", relative)
	}
	if parsed.Go == nil || parsed.Go.Version != module.GoVersion || module.GoVersion != strings.TrimSpace(module.GoVersion) {
		return fmt.Errorf("%s Go version does not match catalog", relative)
	}
	if len(parsed.Replace) > 0 {
		return fmt.Errorf("committed replace directive in %s", relative)
	}
	return nil
}

func checkWorkspace(root, goVersion string, expected []string) error {
	data, err := repositoryfile.Read(root, "go.work", maximumModuleFileSize)
	if errors.Is(err, os.ErrNotExist) {
		if len(expected) > 1 {
			return errors.New("multi-module repository requires go.work")
		}
		return nil
	}
	if err != nil {
		return fmt.Errorf("read go.work: %w", err)
	}
	parsed, err := modfile.ParseWork("go.work", data, nil)
	if err != nil {
		return fmt.Errorf("parse go.work: %w", err)
	}
	if parsed.Go == nil || parsed.Go.Version != goVersion {
		return fmt.Errorf("go.work Go version does not match catalog %q", goVersion)
	}
	actual := make([]string, 0, len(parsed.Use))
	for _, use := range parsed.Use {
		path := filepath.Clean(filepath.FromSlash(use.Path))
		if filepath.IsAbs(path) {
			return fmt.Errorf("go.work contains non-local module path: %s", use.Path)
		}
		if path == ".." {
			return fmt.Errorf("go.work contains non-local module path: %s", use.Path)
		}
		if strings.HasPrefix(path, ".."+string(filepath.Separator)) {
			return fmt.Errorf("go.work contains non-local module path: %s", use.Path)
		}
		actual = append(actual, path)
	}
	slices.Sort(actual)
	if len(actual) == 1 {
		if actual[0] == "." {
			return errors.New("root-only go.work is not allowed")
		}
	}
	if !slices.Equal(actual, expected) {
		return fmt.Errorf("go.work modules %v do not match catalog %v", actual, expected)
	}
	return nil
}
