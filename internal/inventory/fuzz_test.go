package inventory

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/faustbrian/go-library-tools/internal/config"
)

func FuzzManifest(f *testing.F) {
	f.Add([]byte(`{"schema_version":1,"repository":"example","go_version":"1.27.0","modules":[]}`))
	f.Add([]byte(`{"schema_version":1,"unknown":true}`))
	f.Add([]byte(`{"schema_id":"urn:golib:cohesion:module-manifest:v3","schema_version":3,"repository":"github.com/faustbrian/example","go_version":"1.27.0","modules":[{"directory":".","module_path":"github.com/faustbrian/example","go_version":"1.27.0","kind":"fixture","purpose":"","lifecycle":"internal","releasable":false,"version":"","tag_prefix":"","gates":{},"test_tags":[],"build_tags":[],"required_services":[],"external_runtime_dependencies":[],"interoperability_tools":[],"conformance_corpora":[],"specifications":[],"owned_dependencies":[],"reverse_owned_dependencies":[],"packages":[],"family":"","family_label":"","family_description":"","family_order":0}]}`))
	f.Fuzz(func(t *testing.T, input []byte) {
		if len(input) > maximumManifestSize+1 {
			return
		}
		root := t.TempDir()
		if err := os.WriteFile(filepath.Join(root, "modules.json"), input, 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, "packages.json"), []byte(`{"schema_version":1,"repository":"github.com/faustbrian/example","packages":[]}`), 0o600); err != nil {
			t.Fatal(err)
		}
		var destination Inventory
		_ = decode(root, "modules.json", &destination)
		_, _ = load(root, config.Config{Manifests: config.Manifests{Modules: "modules.json", Packages: "packages.json"}}, nil)
	})
}
