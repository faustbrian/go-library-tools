package inventory

import (
	"os"
	"path/filepath"
	"testing"
)

func FuzzManifest(f *testing.F) {
	f.Add([]byte(`{"schema_version":1,"repository":"example","go_version":"1.27.0","modules":[]}`))
	f.Add([]byte(`{"schema_version":1,"unknown":true}`))
	f.Fuzz(func(t *testing.T, input []byte) {
		if len(input) > maximumManifestSize+1 {
			return
		}
		root := t.TempDir()
		if err := os.WriteFile(filepath.Join(root, "manifest.json"), input, 0o600); err != nil {
			t.Fatal(err)
		}
		var destination Inventory
		_ = decode(root, "manifest.json", &destination)
	})
}
