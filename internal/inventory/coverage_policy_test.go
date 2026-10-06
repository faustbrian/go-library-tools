package inventory_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/faustbrian/go-library-tools/v2/internal/config"
	"github.com/faustbrian/go-library-tools/v2/internal/inventory"
)

func TestLoadValidatesCoveragePolicyOwnership(t *testing.T) {
	root := fixture(t)
	write(t, filepath.Join(root, "modules.json"), `{"schema_version":1,"repository":"github.com/faustbrian/example","go_version":"1.27.0","modules":[{"directory":".","module_path":"github.com/faustbrian/example","go_version":"1.27.0","kind":"public","releasable":true,"gates":{"coverage":true},"packages":[]}]}`)
	policy := config.Config{Manifests: config.Manifests{Modules: "modules.json", Packages: "packages.json"}, Coverage: config.Coverage{Modules: []config.CoverageModule{{Module: ".", Mode: "evidence"}}}}
	if _, err := inventory.Load(root, policy); err != nil {
		t.Fatalf("enabled owner: %v", err)
	}
	policy.Coverage.Modules[0].Module = "missing"
	if _, err := inventory.Load(root, policy); err == nil || !strings.Contains(err.Error(), "references unknown module") {
		t.Fatalf("unknown owner: %v", err)
	}
	policy.Coverage.Modules[0].Module = "."
	write(t, filepath.Join(root, "modules.json"), `{"schema_version":1,"repository":"github.com/faustbrian/example","go_version":"1.27.0","modules":[{"directory":".","module_path":"github.com/faustbrian/example","go_version":"1.27.0","kind":"public","releasable":true,"gates":{},"packages":[]}]}`)
	if _, err := inventory.Load(root, policy); err == nil || !strings.Contains(err.Error(), "is not enabled") {
		t.Fatalf("disabled owner: %v", err)
	}
}
