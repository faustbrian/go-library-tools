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
	if got.Mutation.Root != ".verification/mutation" {
		t.Fatalf("Load() mutation root = %q", got.Mutation.Root)
	}
}

func TestLoadAcceptsReleaseChecksumPin(t *testing.T) {
	root := t.TempDir()
	digest := strings.Repeat("a", 64)
	write(t, filepath.Join(root, ".golib.yaml"), "schema_version: 1\ntool_version: v1.0.0\ntool_checksums_sha256: "+digest+"\n")
	got, err := config.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if got.ToolChecksumsSHA256 != digest {
		t.Fatalf("checksum digest = %q", got.ToolChecksumsSHA256)
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
	tests := map[string]struct {
		content string
		want    string
	}{
		"malformed":             {content: "schema_version: [\n", want: "decode .golib.yaml"},
		"multiple":              {content: "schema_version: 1\ntool_version: v1.0.0\n---\nsecond: document\n", want: "multiple YAML documents"},
		"bad trailing document": {content: "schema_version: 1\ntool_version: v1.0.0\n---\n[\n", want: "decode trailing YAML document"},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			write(t, filepath.Join(root, ".golib.yaml"), test.content)
			if _, err := config.Load(root); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Load() error = %v", err)
			}
		})
	}
}

func TestLoadRejectsUnsupportedSchemaAndToolVersions(t *testing.T) {
	tests := map[string]string{
		"schema":             "schema_version: 2\ntool_version: v1.0.0\n",
		"tool":               "schema_version: 1\ntool_version: latest\n",
		"checksum short":     "schema_version: 1\ntool_version: v1.0.0\ntool_checksums_sha256: abc\n",
		"checksum uppercase": "schema_version: 1\ntool_version: v1.0.0\ntool_checksums_sha256: " + strings.Repeat("A", 64) + "\n",
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

func TestLoadAcceptsExplicitMutationEvidenceRoot(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, ".golib.yaml"), "schema_version: 1\ntool_version: v1.0.0\nmutation:\n  root: evidence/mutations\n")
	got, err := config.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if got.Mutation.Root != "evidence/mutations" {
		t.Fatalf("mutation root = %q", got.Mutation.Root)
	}
}

func TestLoadAcceptsPinnedExternalRuntimes(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, ".golib.yaml"), `schema_version: 1
tool_version: v1.0.0
runtimes:
  deno: 2.9.4
  zsh: "5.9"
`)
	got, err := config.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if got.Runtimes.Deno != "2.9.4" || got.Runtimes.Zsh != "5.9" {
		t.Fatalf("runtimes = %#v", got.Runtimes)
	}
}

func TestLoadRejectsUnsupportedExternalRuntimes(t *testing.T) {
	tests := map[string]string{
		"floating deno":   "runtimes:\n  deno: latest\n",
		"unsupported zsh": "runtimes:\n  zsh: \"5.8\"\n",
	}
	for name, runtime := range tests {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			write(t, filepath.Join(root, ".golib.yaml"), "schema_version: 1\ntool_version: v1.0.0\n"+runtime)
			if _, err := config.Load(root); !errors.Is(err, config.ErrInvalid) {
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

func TestLoadAcceptsTypedRepositoryOperations(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, ".golib.yaml"), `schema_version: 1
tool_version: v1.0.0
operations:
  - module: .
    gate: test
    steps:
      - type: go-test
        packages: [./manual]
        run: ^TestStress$
        count: 20
        timeout: 10m
  - module: .
    gate: conformance
    steps:
      - type: go-test
        packages: [., ./...]
        run: ^TestConformance$
        count: 1
        timeout: 10m
      - type: make
        makefile: verification/package.mk
        target: generated
        timeout: 10m
`)
	got, err := config.Load(root)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(got.Operations) != 2 || len(got.Operations[0].Steps) != 1 {
		t.Fatalf("Load() operations = %#v", got.Operations)
	}
}

func TestLoadAcceptsTypedAPIBaselines(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, ".golib.yaml"), `schema_version: 1
tool_version: v1.0.0
api:
  baselines:
    - module: .
      mode: go-doc
      path: api/v1.txt
      packages: [., ./manual]
    - module: integration/reference
      mode: apidiff
      path: integration/reference/api/v1.export
`)
	got, err := config.Load(root)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(got.API.Baselines) != 2 || got.API.Baselines[0].Packages[1] != "./manual" {
		t.Fatalf("Load() API baselines = %#v", got.API.Baselines)
	}
}

func TestLoadAcceptsNestedModuleOperations(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, ".golib.yaml"), `schema_version: 1
tool_version: v1.0.0
operations:
  - module: integration/reference
    gate: fuzz
    steps:
      - type: go-test
        packages: [./...]
        fuzz: FuzzInput
`)
	got, err := config.Load(root)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got.Operations[0].Module != "integration/reference" {
		t.Fatalf("Load() module = %q", got.Operations[0].Module)
	}
}

func TestLoadAppliesBoundedOperationDefaults(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, ".golib.yaml"), `schema_version: 1
tool_version: v1.0.0
operations:
  - module: .
    gate: fuzz
    steps:
      - type: go-test
        fuzz: FuzzSpec
  - module: .
    gate: benchmark
    steps:
      - type: go-test
        benchmark: BenchmarkSpec
`)
	got, err := config.Load(root)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got.Operations[0].Steps[0].Budget != "10000x" {
		t.Fatalf("fuzz budget = %q", got.Operations[0].Steps[0].Budget)
	}
	if got.Operations[1].Steps[0].Budget != "100ms" {
		t.Fatalf("benchmark budget = %q", got.Operations[1].Steps[0].Budget)
	}
}

func TestLoadRejectsInvalidTypedOperations(t *testing.T) {
	tests := map[string]string{
		"unknown gate":        "module: .\n    gate: deploy\n    steps:\n      - type: go-test\n        packages: [./...]",
		"test benchmark":      "module: .\n    gate: test\n    steps:\n      - type: go-test\n        benchmark: BenchmarkInput",
		"test fuzz selector":  "module: .\n    gate: test\n    steps:\n      - type: go-test\n        fuzz: FuzzInput",
		"go test makefile":    "module: .\n    gate: conformance\n    steps:\n      - type: go-test\n        makefile: verification/package.mk",
		"go test make target": "module: .\n    gate: conformance\n    steps:\n      - type: go-test\n        target: generated",
		"missing steps":       "module: .\n    gate: docs",
		"unknown type":        "module: .\n    gate: docs\n    steps:\n      - type: shell",
		"invalid timeout":     "module: .\n    gate: docs\n    steps:\n      - type: go-test\n        timeout: forever",
		"zero timeout":        "module: .\n    gate: docs\n    steps:\n      - type: go-test\n        timeout: 0s",
		"negative timeout":    "module: .\n    gate: docs\n    steps:\n      - type: go-test\n        timeout: -1s",
		"invalid count":       "module: .\n    gate: docs\n    steps:\n      - type: go-test\n        count: -1",
		"missing module":      "gate: docs\n    steps:\n      - type: go-test",
		"escaping module":     "module: ../other\n    gate: docs\n    steps:\n      - type: go-test",
		"multiple selectors":  "module: .\n    gate: docs\n    steps:\n      - type: go-test\n        run: Test\n        benchmark: Benchmark",
		"empty package":       "module: .\n    gate: docs\n    steps:\n      - type: go-test\n        packages: ['']",
		"flag package":        "module: .\n    gate: docs\n    steps:\n      - type: go-test\n        packages: ['-run=Injected']",
		"parent package":      "module: .\n    gate: docs\n    steps:\n      - type: go-test\n        packages: ['./../other']",
		"embedded wildcard":   "module: .\n    gate: docs\n    steps:\n      - type: go-test\n        packages: ['./.../other']",
		"orphan budget":       "module: .\n    gate: docs\n    steps:\n      - type: go-test\n        budget: 1s",
		"zero duration":       "module: .\n    gate: fuzz\n    steps:\n      - type: go-test\n        fuzz: FuzzInput\n        budget: 0s",
		"zero iterations":     "module: .\n    gate: fuzz\n    steps:\n      - type: go-test\n        fuzz: FuzzInput\n        budget: 0x",
		"invalid budget":      "module: .\n    gate: benchmark\n    steps:\n      - type: go-test\n        benchmark: BenchmarkInput\n        budget: many",
		"make missing file":   "module: .\n    gate: conformance\n    steps:\n      - type: make\n        target: generated",
		"make escaping file":  "module: .\n    gate: conformance\n    steps:\n      - type: make\n        makefile: ../package.mk\n        target: generated",
		"make directory file": "module: .\n    gate: conformance\n    steps:\n      - type: make\n        makefile: .\n        target: generated",
		"make missing target": "module: .\n    gate: conformance\n    steps:\n      - type: make\n        makefile: verification/package.mk",
		"make flag target":    "module: .\n    gate: conformance\n    steps:\n      - type: make\n        makefile: verification/package.mk\n        target: --eval=bad",
		"make packages":       "module: .\n    gate: conformance\n    steps:\n      - type: make\n        makefile: verification/package.mk\n        target: generated\n        packages: [./...]",
		"make run":            "module: .\n    gate: conformance\n    steps:\n      - type: make\n        makefile: verification/package.mk\n        target: generated\n        run: Test",
		"make benchmark":      "module: .\n    gate: conformance\n    steps:\n      - type: make\n        makefile: verification/package.mk\n        target: generated\n        benchmark: Benchmark",
		"make fuzz":           "module: .\n    gate: conformance\n    steps:\n      - type: make\n        makefile: verification/package.mk\n        target: generated\n        fuzz: FuzzInput",
		"make budget":         "module: .\n    gate: conformance\n    steps:\n      - type: make\n        makefile: verification/package.mk\n        target: generated\n        budget: 1s",
		"make count":          "module: .\n    gate: conformance\n    steps:\n      - type: make\n        makefile: verification/package.mk\n        target: generated\n        count: 2",
	}
	for name, operation := range tests {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			write(t, filepath.Join(root, ".golib.yaml"), "schema_version: 1\ntool_version: v1.0.0\noperations:\n  - "+operation+"\n")
			if _, err := config.Load(root); !errors.Is(err, config.ErrInvalid) {
				t.Fatalf("Load() error = %v", err)
			}
		})
	}
}

func TestLoadRejectsInvalidAPIBaselines(t *testing.T) {
	tests := map[string]string{
		"duplicate":            "- module: .\n      mode: apidiff\n      path: api/a.txt\n    - module: .\n      mode: apidiff\n      path: api/b.txt",
		"escaping module":      "- module: ../other\n      mode: apidiff\n      path: api/v1.txt",
		"escaping path":        "- module: .\n      mode: apidiff\n      path: ../api.txt",
		"directory path":       "- module: .\n      mode: apidiff\n      path: .",
		"missing module":       "- mode: apidiff\n      path: api/v1.txt",
		"missing doc packages": "- module: .\n      mode: go-doc\n      path: api/v1.txt",
		"apidiff packages":     "- module: .\n      mode: apidiff\n      path: api/v1.txt\n      packages: [.]",
		"invalid doc package":  "- module: .\n      mode: go-doc\n      path: api/v1.txt\n      packages: ['-all']",
		"unknown mode":         "- module: .\n      mode: shell\n      path: api/v1.txt",
		"wildcard package":     "- module: .\n      mode: go-doc\n      path: api/v1.txt\n      packages: [./...]",
	}
	for name, baselines := range tests {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			write(t, filepath.Join(root, ".golib.yaml"), "schema_version: 1\ntool_version: v1.0.0\napi:\n  baselines:\n    "+baselines+"\n")
			if _, err := config.Load(root); !errors.Is(err, config.ErrInvalid) {
				t.Fatalf("Load() error = %v", err)
			}
		})
	}
}

func TestLoadRejectsDuplicateOperationOwnership(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, ".golib.yaml"), `schema_version: 1
tool_version: v1.0.0
operations:
  - module: .
    gate: docs
    steps: [{type: go-test}]
  - module: .
    gate: docs
    steps: [{type: go-test}]
`)
	if _, err := config.Load(root); !errors.Is(err, config.ErrInvalid) {
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
