package releasecheck

import (
	"strings"
	"testing"

	"github.com/faustbrian/go-library-tools/internal/config"
	"github.com/faustbrian/go-library-tools/internal/inventory"
)

func TestValidateReturnsStableReleasableModules(t *testing.T) {
	catalog := inventory.Inventory{Modules: []inventory.Module{
		module("nested", "nested/v"),
		{Directory: "reference", Releasable: false},
		module(".", "v"),
	}}
	directories, err := Validate(catalog, config.Config{ToolVersion: "v1.0.0"})
	if err != nil || strings.Join(directories, ",") != ".,nested" {
		t.Fatalf("Validate() = %#v, %v", directories, err)
	}
}

func TestValidateRejectsInvalidReleaseContracts(t *testing.T) {
	tests := []struct {
		name   string
		policy config.Config
		module inventory.Module
		want   string
	}{
		{"development tool", config.Config{ToolVersion: "v0.9.0"}, module(".", "v"), "stable semantic"},
		{"invalid tool", config.Config{ToolVersion: "dev"}, module(".", "v"), "stable semantic"},
		{"invalid module version", stablePolicy(), func() inventory.Module { value := module(".", "v"); value.Version = "v1.0.0"; return value }(), "without v prefix"},
		{"prerelease module", stablePolicy(), func() inventory.Module { value := module(".", "v"); value.Version = "0.9.0"; return value }(), "stable version"},
		{"root prefix", stablePolicy(), module(".", "root/v"), "tag prefix"},
		{"nested prefix", stablePolicy(), module("nested", "v"), "tag prefix"},
		{"disabled gate", stablePolicy(), func() inventory.Module { value := module(".", "v"); value.Gates["mutation"] = false; return value }(), "mutation is disabled"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := Validate(inventory.Inventory{Modules: []inventory.Module{test.module}}, test.policy)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Validate() error = %v, want %q", err, test.want)
			}
		})
	}
	if _, err := Validate(inventory.Inventory{Modules: []inventory.Module{{Releasable: false}}}, stablePolicy()); err == nil || !strings.Contains(err.Error(), "no releasable") {
		t.Fatalf("Validate(no modules) error = %v", err)
	}
}

func module(directory, prefix string) inventory.Module {
	gates := make(map[string]bool, len(requiredGates))
	for _, gate := range requiredGates {
		gates[gate] = true
	}
	return inventory.Module{Directory: directory, Releasable: true, Version: "1.0.0", TagPrefix: prefix, Gates: gates}
}

func stablePolicy() config.Config { return config.Config{ToolVersion: "v1.0.0"} }
