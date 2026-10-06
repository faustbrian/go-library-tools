package cohesion

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestSchemaGeneratorProducesCanonicalFilesAndReportsIOFailures(t *testing.T) {
	generator, err := os.ReadFile("schema_generate_main.go")
	if err != nil {
		t.Fatal(err)
	}
	schemas := []struct{ input, output string }{
		{"modules.schema.json", "modules_schema_generated.go"},
		{"cohesion-catalog.schema.json", "catalog_schema_generated.go"},
		{"cohesion-inputs.schema.json", "inputs_schema_generated.go"},
		{"cohesion-sources.schema.json", "sources_schema_generated.go"},
	}
	for _, scenario := range []string{"success", "missing input", "output is directory"} {
		t.Run(scenario, func(t *testing.T) {
			root := t.TempDir()
			working := filepath.Join(root, "internal", "cohesion")
			schemaDir := filepath.Join(root, "schema")
			for _, dir := range []string{working, schemaDir} {
				if err := os.MkdirAll(dir, 0o700); err != nil {
					t.Fatal(err)
				}
			}
			if err := os.WriteFile(filepath.Join(working, "schema_generate_main.go"), generator, 0o600); err != nil {
				t.Fatal(err)
			}
			for _, schema := range schemas {
				if scenario == "missing input" && schema.input == "modules.schema.json" {
					continue
				}
				data, err := os.ReadFile(filepath.Join("..", "..", "schema", schema.input))
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(schemaDir, schema.input), data, 0o600); err != nil {
					t.Fatal(err)
				}
			}
			if scenario == "output is directory" {
				if err := os.Mkdir(filepath.Join(working, schemas[0].output), 0o700); err != nil {
					t.Fatal(err)
				}
			}
			command := exec.CommandContext(t.Context(), "go", "run", "schema_generate_main.go")
			command.Dir = working
			output, runErr := command.CombinedOutput()
			if scenario != "success" {
				if runErr == nil || !bytes.Contains(output, []byte("panic:")) || !bytes.Contains(output, []byte("modules")) {
					t.Fatal("generator did not report its input/output failure")
				}
				for _, schema := range schemas[1:] {
					if _, err := os.Stat(filepath.Join(working, schema.output)); !os.IsNotExist(err) {
						t.Fatal("generator continued after its first IO failure")
					}
				}
				return
			}
			if runErr != nil {
				t.Fatalf("generator failed: %v", runErr)
			}
			if len(output) != 0 {
				t.Fatal("successful generator emitted unexpected diagnostics")
			}
			for _, schema := range schemas {
				want, err := os.ReadFile(schema.output)
				if err != nil {
					t.Fatal(err)
				}
				generated := filepath.Join(working, schema.output)
				got, err := os.ReadFile(generated)
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(got, want) {
					t.Fatalf("generator output differs from canonical %s", schema.output)
				}
				info, err := os.Stat(generated)
				if err != nil || info.Mode().Perm() != 0o600 {
					t.Fatal("generated file does not have private permissions")
				}
			}
		})
	}
}
