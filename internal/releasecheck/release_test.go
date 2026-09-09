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

func TestValidateReturnsExactlyTheSelectedReleasableModule(t *testing.T) {
	invalidSibling := module("invalid", "invalid/v")
	invalidSibling.Version = "0.9.0"
	catalog := inventory.Inventory{Modules: []inventory.Module{
		invalidSibling,
		{Directory: "reference", Releasable: false},
		module(".", "v"),
	}}
	directories, err := Validate(catalog, stablePolicy(), ".")
	if err != nil || strings.Join(directories, ",") != "." {
		t.Fatalf("Validate(selected) = %#v, %v", directories, err)
	}
}

func TestValidateAllowsRiskSelectedGatesToBeDisabled(t *testing.T) {
	for _, gate := range []string{"coverage", "mutation", "race"} {
		t.Run(gate, func(t *testing.T) {
			selected := module(".", "v")
			selected.Gates[gate] = false

			directories, err := Validate(
				inventory.Inventory{Modules: []inventory.Module{selected}},
				stablePolicy(),
			)
			if err != nil || strings.Join(directories, ",") != "." {
				t.Fatalf("Validate() = %#v, %v", directories, err)
			}
		})
	}
}

func TestValidateRejectsInvalidReleaseSelections(t *testing.T) {
	catalog := inventory.Inventory{Modules: []inventory.Module{
		module(".", "v"),
		{Directory: "reference", Releasable: false},
	}}
	for _, test := range []struct {
		name      string
		selection []string
		want      string
	}{
		{name: "unknown", selection: []string{"missing"}, want: "unknown module: missing"},
		{name: "not releasable", selection: []string{"reference"}, want: "module reference is not releasable"},
		{name: "multiple", selection: []string{".", "reference"}, want: "at most one"},
	} {
		t.Run(test.name, func(t *testing.T) {
			if _, err := Validate(catalog, stablePolicy(), test.selection...); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Validate(%v) error = %v, want %q", test.selection, err, test.want)
			}
		})
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
		{"disabled gate", stablePolicy(), func() inventory.Module { value := module(".", "v"); value.Gates["tests"] = false; return value }(), "tests is disabled"},
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
