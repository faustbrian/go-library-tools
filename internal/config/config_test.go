package config_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/faustbrian/go-library-tools/internal/config"
	"github.com/faustbrian/go-library-tools/internal/repositoryfile"
)

func TestLoadAppliesStableDefaults(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, ".golib.yaml"), "schema_version: 1\ntool_version: v1.0.0\n")

	got, err := config.Load(root)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if got.SchemaVersion != 1 || got.ToolVersion != "v1.0.0" {
		t.Fatalf("Load() identity = %#v", got)
	}
	if got.Manifests.Modules != "modules.json" || got.Manifests.Packages != "packages.json" {
		t.Fatalf("Load() manifests = %#v", got.Manifests)
	}
	if got.Evidence.Root != ".verification" {
		t.Fatalf("Load() evidence root = %q", got.Evidence.Root)
	}
}

func TestLoadRejectsUnknownFields(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, ".golib.yaml"), "schema_version: 1\ntool_version: v1.0.0\nsecret: value\n")

	_, err := config.Load(root)
	if err == nil || !strings.Contains(err.Error(), "field secret not found") {
		t.Fatalf("Load() error = %v", err)
	}
}

func TestLoadRejectsMalformedAndMultipleDocuments(t *testing.T) {
	tests := map[string]string{
		"malformed":             "schema_version: [\n",
		"multiple":              "schema_version: 1\ntool_version: v1.0.0\n---\nsecond: document\n",
		"bad trailing document": "schema_version: 1\ntool_version: v1.0.0\n---\n[\n",
	}
	for name, content := range tests {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			write(t, filepath.Join(root, ".golib.yaml"), content)
			if _, err := config.Load(root); err == nil {
				t.Fatal("Load() error = nil")
			}
		})
	}
}

func TestLoadRejectsUnsupportedSchemaAndToolVersions(t *testing.T) {
	tests := map[string]string{
		"schema": "schema_version: 2\ntool_version: v1.0.0\n",
		"tool":   "schema_version: 1\ntool_version: latest\n",
	}

	for name, content := range tests {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			write(t, filepath.Join(root, ".golib.yaml"), content)
			if _, err := config.Load(root); !errors.Is(err, config.ErrInvalid) {
				t.Fatalf("Load() error = %v", err)
			}
		})
	}
}

func TestLoadRejectsPathsOutsideRepository(t *testing.T) {
	tests := map[string]string{
		"parent":   "../modules.json",
		"absolute": "/tmp/modules.json",
	}
	for name, path := range tests {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			write(t, filepath.Join(root, ".golib.yaml"), "schema_version: 1\ntool_version: v1.0.0\nmanifest:\n  modules: "+path+"\n")
			_, err := config.Load(root)
			if err == nil || !strings.Contains(err.Error(), "manifest.modules") {
				t.Fatalf("Load() error = %v", err)
			}
		})
	}
}

func TestLoadReportsReadFailure(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, ".golib.yaml"), 0o700); err != nil {
		t.Fatal(err)
	}
	if _, err := config.Load(root); err == nil || !errors.Is(err, repositoryfile.ErrNotRegular) {
		t.Fatalf("Load() error = %v", err)
	}
}

func TestLoadReportsMissingAndOversizedConfiguration(t *testing.T) {
	root := t.TempDir()
	if _, err := config.Load(root); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("Load() missing error = %v", err)
	}

	write(t, filepath.Join(root, ".golib.yaml"), strings.Repeat("#", config.MaximumSize+1))
	if _, err := config.Load(root); !errors.Is(err, config.ErrTooLarge) {
		t.Fatalf("Load() oversized error = %v", err)
	}
}

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}
