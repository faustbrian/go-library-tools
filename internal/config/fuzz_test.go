package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/faustbrian/go-library-tools/internal/config"
)

func FuzzConfig(f *testing.F) {
	f.Add([]byte("schema_version: 1\ntool_version: v1.0.0\n"))
	f.Add([]byte("schema_version: 2\nunknown: true\n"))
	f.Fuzz(func(t *testing.T, input []byte) {
		if len(input) > config.MaximumSize+1 {
			return
		}
		root := t.TempDir()
		if err := os.WriteFile(filepath.Join(root, ".golib.yaml"), input, 0o600); err != nil {
			t.Fatal(err)
		}
		_, _ = config.Load(root)
	})
}
