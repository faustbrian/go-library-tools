// Package releasecheck validates immutable release metadata and required gates.
package releasecheck

import (
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/faustbrian/go-library-tools/internal/config"
	"github.com/faustbrian/go-library-tools/internal/inventory"
	"golang.org/x/mod/semver"
)

var requiredGates = []string{
	"coverage", "documentation", "lint", "mutation", "race", "security", "tests",
}

// Validate returns deterministic releasable module directories only when all
// release metadata and mandatory gates are valid.
func Validate(catalog inventory.Inventory, policy config.Config) ([]string, error) {
	if !semver.IsValid(policy.ToolVersion) || semver.Major(policy.ToolVersion) == "v0" {
		return nil, errors.New("tool policy must require a stable semantic version")
	}
	directories := make([]string, 0, len(catalog.Modules))
	for _, module := range catalog.Modules {
		if !module.Releasable {
			continue
		}
		version := "v" + strings.TrimPrefix(module.Version, "v")
		if !semver.IsValid(version) || semver.Major(version) == "v0" || module.Version != strings.TrimPrefix(version, "v") {
			return nil, fmt.Errorf("module %s must declare a stable version without v prefix", module.Directory)
		}
		expectedPrefix := "v"
		if module.Directory != "." {
			expectedPrefix = filepath.ToSlash(module.Directory) + "/v"
		}
		if module.TagPrefix != expectedPrefix {
			return nil, fmt.Errorf("module %s tag prefix must be %q", module.Directory, expectedPrefix)
		}
		for _, gate := range requiredGates {
			if !module.Gates[gate] {
				return nil, fmt.Errorf("module %s release gate %s is disabled", module.Directory, gate)
			}
		}
		directories = append(directories, module.Directory)
	}
	if len(directories) == 0 {
		return nil, errors.New("repository has no releasable modules")
	}
	sort.Strings(directories)
	return directories, nil
}
