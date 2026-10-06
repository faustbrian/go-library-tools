package config_test

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"testing"

	"github.com/faustbrian/go-library-tools/v2/internal/config"
)

func TestLoadRejectsInvalidCoveragePolicies(t *testing.T) {
	for name, entries := range map[string]string{
		"unknown mode":     "    - module: .\n      mode: lower\n",
		"missing mode":     "    - module: .\n",
		"missing module":   "    - mode: evidence\n",
		"parent path":      "    - module: ../other\n      mode: evidence\n",
		"duplicate module": "    - module: .\n      mode: exact\n    - module: .\n      mode: evidence\n",
	} {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			write(t, filepath.Join(root, ".golib.yaml"), "schema_version: 1\ntool_version: v1.0.0\ncoverage:\n  modules:\n"+entries)
			if _, err := config.Load(root); !errors.Is(err, config.ErrInvalid) {
				t.Fatalf("invalid coverage acceptance = %v", err)
			}
		})
	}
}

func TestLoadRetainsExplicitCoverageEvidencePolicy(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, ".golib.yaml"), `schema_version: 1
tool_version: v1.0.0
coverage:
  modules:
    - module: .
      mode: evidence
`)
	policy, err := config.Load(root)
	if err != nil {
		t.Fatalf("explicit coverage policy refused: %v", err)
	}
	encoded, err := json.Marshal(policy)
	if err != nil {
		t.Fatal(err)
	}
	var retained struct {
		Coverage struct {
			Modules []struct {
				Module string `json:"module"`
				Mode   string `json:"mode"`
			} `json:"modules"`
		} `json:"coverage"`
	}
	if err := json.Unmarshal(encoded, &retained); err != nil {
		t.Fatal(err)
	}
	if len(retained.Coverage.Modules) != 1 || retained.Coverage.Modules[0].Module != "." ||
		retained.Coverage.Modules[0].Mode != "evidence" {
		t.Fatal("explicit coverage policy was not retained for the gate owner")
	}
}

func TestLoadRetainsNestedModuleCoveragePolicies(t *testing.T) {
	for _, mode := range []string{"exact", "evidence"} {
		t.Run(mode, func(t *testing.T) {
			root := t.TempDir()
			write(t, filepath.Join(root, ".golib.yaml"), "schema_version: 1\ntool_version: v1.0.0\ncoverage:\n  modules:\n    - module: adapters/http\n      mode: "+mode+"\n")
			policy, err := config.Load(root)
			if err != nil {
				t.Fatalf("nested coverage policy refused: %v", err)
			}
			want := config.CoverageModule{Module: "adapters/http", Mode: mode}
			if len(policy.Coverage.Modules) != 1 || policy.Coverage.Modules[0] != want {
				t.Fatalf("nested coverage policy = %#v, want %#v", policy.Coverage.Modules, want)
			}
		})
	}
}
