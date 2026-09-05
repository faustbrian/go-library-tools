package cohesion

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/faustbrian/go-library-tools/internal/inventory"
)

func TestAggregateBuildsDeterministicEcosystemViewsFromRepositoryEngineeringProjections(t *testing.T) {
	root := t.TempDir()
	identity := Identity{
		DesignLanguageVersion: "1.0",
		DesignLanguageSHA256:  strings.Repeat("a", 64),
		SourceIdentity:        "v1.3.0",
		ToolingVersion:        "v1.3.0",
		PublicationStatus:     "published",
	}

	alpha := validAggregateModule()
	alpha.ModulePath = "github.com/faustbrian/go-alpha"
	alpha.Packages[0].ImportPath = alpha.ModulePath
	alpha.Cohesion.PrimaryEntryPackages = []string{alpha.ModulePath}
	alpha.Cohesion.Responsibility = "Provide alpha values."
	beta := validAggregateModule()
	beta.ModulePath = "github.com/faustbrian/go-beta"
	beta.Packages[0].ImportPath = beta.ModulePath
	beta.Cohesion.PrimaryEntryPackages = []string{beta.ModulePath}
	beta.Cohesion.Responsibility = "Provide beta values."

	projections := []struct {
		repository string
		module     inventory.Module
		filename   string
	}{
		{repository: "github.com/faustbrian/go-alpha", module: alpha, filename: "alpha.json"},
		{repository: "github.com/faustbrian/go-beta", module: beta, filename: "beta.json"},
	}
	inputs := aggregateInputManifest{
		SchemaVersion: 1,
		DesignLanguage: aggregateDesignLanguage{
			Version: identity.DesignLanguageVersion,
			SHA256:  identity.DesignLanguageSHA256,
		},
	}
	for _, projection := range projections {
		envelope, err := Project(inventory.Inventory{Repository: projection.repository, Modules: []inventory.Module{projection.module}}, "engineering", identity)
		if err != nil {
			t.Fatal(err)
		}
		data, err := json.Marshal(envelope)
		if err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(root, projection.filename)
		if err := os.WriteFile(path, data, 0o600); err != nil {
			t.Fatal(err)
		}
		digest := sha256.Sum256(data)
		inputs.Repositories = append(inputs.Repositories, aggregateRepositoryInput{
			Repository: projection.repository,
			Projection: projection.filename,
			SHA256:     hex.EncodeToString(digest[:]),
		})
	}
	manifestData, err := json.Marshal(inputs)
	if err != nil {
		t.Fatal(err)
	}
	manifestPath := filepath.Join(root, "inputs.json")
	if err := os.WriteFile(manifestPath, manifestData, 0o600); err != nil {
		t.Fatal(err)
	}

	artifacts, err := Aggregate(manifestPath, identity)
	if err != nil {
		t.Fatal(err)
	}
	for name, data := range map[string][]byte{
		"consumer JSON":        artifacts.ConsumerJSON,
		"consumer Markdown":    artifacts.ConsumerMarkdown,
		"engineering JSON":     artifacts.EngineeringJSON,
		"engineering Markdown": artifacts.EngineeringMarkdown,
	} {
		if len(data) == 0 {
			t.Fatalf("%s is empty", name)
		}
	}

	assertAggregateModules(t, artifacts.ConsumerJSON, "consumer", []string{alpha.ModulePath, beta.ModulePath})
	assertAggregateModules(t, artifacts.EngineeringJSON, "engineering", []string{alpha.ModulePath, beta.ModulePath})
	if got := string(artifacts.ConsumerMarkdown); !strings.Contains(got, "`github.com/faustbrian/go-alpha`") || !strings.Contains(got, "`github.com/faustbrian/go-beta`") {
		t.Fatalf("consumer Markdown = %q", got)
	}
	if got := string(artifacts.EngineeringMarkdown); !strings.Contains(got, "`github.com/faustbrian/go-alpha`") || !strings.Contains(got, "`github.com/faustbrian/go-beta`") {
		t.Fatalf("engineering Markdown = %q", got)
	}
}

func TestAggregateRejectsMissingDuplicateAndUnsortedRepositoryInputs(t *testing.T) {
	identity := Identity{
		DesignLanguageVersion: "1.0", DesignLanguageSHA256: strings.Repeat("a", 64),
		SourceIdentity: "v1.3.0", ToolingVersion: "v1.3.0", PublicationStatus: "published",
	}
	for _, test := range []struct {
		name         string
		repositories []aggregateRepositoryInput
		want         string
	}{
		{name: "missing", want: "input schema"},
		{name: "duplicate", repositories: []aggregateRepositoryInput{
			{Repository: "github.com/faustbrian/go-alpha", Projection: "alpha.json", SHA256: strings.Repeat("b", 64)},
			{Repository: "github.com/faustbrian/go-alpha", Projection: "alpha.json", SHA256: strings.Repeat("b", 64)},
		}, want: "input schema"},
		{name: "unsorted", repositories: []aggregateRepositoryInput{
			{Repository: "github.com/faustbrian/go-zeta", Projection: "zeta.json", SHA256: strings.Repeat("b", 64)},
			{Repository: "github.com/faustbrian/go-alpha", Projection: "alpha.json", SHA256: strings.Repeat("c", 64)},
		}, want: "unique and sorted"},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			inputs := aggregateInputManifest{
				SchemaVersion: 1,
				DesignLanguage: aggregateDesignLanguage{
					Version: identity.DesignLanguageVersion,
					SHA256:  identity.DesignLanguageSHA256,
				},
				Repositories: test.repositories,
			}
			data, err := json.Marshal(inputs)
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(root, "inputs.json")
			if err := os.WriteFile(path, data, 0o600); err != nil {
				t.Fatal(err)
			}
			if _, err := Aggregate(path, identity); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Aggregate() error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestAggregateRejectsRepositoryCountAboveEcosystemBound(t *testing.T) {
	identity := Identity{
		DesignLanguageVersion: "1.0", DesignLanguageSHA256: strings.Repeat("a", 64),
		SourceIdentity: "v1.3.0", ToolingVersion: "v1.3.0", PublicationStatus: "published",
	}
	inputs := aggregateInputManifest{
		SchemaVersion: 1,
		DesignLanguage: aggregateDesignLanguage{
			Version: identity.DesignLanguageVersion,
			SHA256:  identity.DesignLanguageSHA256,
		},
		Repositories: make([]aggregateRepositoryInput, maximumAggregateRepositories+1),
	}
	for index := range inputs.Repositories {
		inputs.Repositories[index] = aggregateRepositoryInput{
			Repository: fmt.Sprintf("github.com/faustbrian/go-example-%03d", index),
			Projection: fmt.Sprintf("repository-%03d.json", index),
			SHA256:     strings.Repeat("b", 64),
		}
	}
	data, err := json.Marshal(inputs)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "inputs.json")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := Aggregate(path, identity); err == nil || !strings.Contains(err.Error(), "repositories") {
		t.Fatalf("Aggregate() error = %v, want repository-count rejection", err)
	}
}

func TestAggregateRejectsRepositoryProjectionThatViolatesCatalogSchema(t *testing.T) {
	root := t.TempDir()
	identity := Identity{
		DesignLanguageVersion: "1.0", DesignLanguageSHA256: strings.Repeat("a", 64),
		SourceIdentity: "v1.3.0", ToolingVersion: "v1.3.0", PublicationStatus: "published",
	}
	repository := "github.com/faustbrian/go-example"
	envelope, err := Project(inventory.Inventory{Repository: repository, Modules: []inventory.Module{validAggregateModule()}}, "engineering", identity)
	if err != nil {
		t.Fatal(err)
	}
	projection := map[string]any{}
	data, err := json.Marshal(envelope)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &projection); err != nil {
		t.Fatal(err)
	}
	projection["unexpected"] = true
	data, err = json.Marshal(projection)
	if err != nil {
		t.Fatal(err)
	}
	projectionPath := filepath.Join(root, "repository.json")
	if err := os.WriteFile(projectionPath, data, 0o600); err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(data)
	inputs := aggregateInputManifest{
		SchemaVersion: 1,
		DesignLanguage: aggregateDesignLanguage{
			Version: identity.DesignLanguageVersion,
			SHA256:  identity.DesignLanguageSHA256,
		},
		Repositories: []aggregateRepositoryInput{{
			Repository: repository, Projection: filepath.Base(projectionPath), SHA256: hex.EncodeToString(digest[:]),
		}},
	}
	manifestData, err := json.Marshal(inputs)
	if err != nil {
		t.Fatal(err)
	}
	manifestPath := filepath.Join(root, "inputs.json")
	if err := os.WriteFile(manifestPath, manifestData, 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := Aggregate(manifestPath, identity); err == nil || !strings.Contains(err.Error(), "schema") {
		t.Fatalf("Aggregate() error = %v, want schema rejection", err)
	}
}

func TestAggregateRejectsInputManifestThatViolatesPublishedSchema(t *testing.T) {
	root := t.TempDir()
	identity := Identity{
		DesignLanguageVersion: "1.0", DesignLanguageSHA256: strings.Repeat("a", 64),
		SourceIdentity: "v1.3.0", ToolingVersion: "v1.3.0", PublicationStatus: "published",
	}
	manifest := map[string]any{
		"schema_version": 1,
		"design_language": map[string]any{
			"version": identity.DesignLanguageVersion,
			"sha256":  identity.DesignLanguageSHA256,
		},
		"repositories": []any{map[string]any{
			"repository": "github.com/faustbrian/go-example",
			"projection": "example.json",
			"sha256":     strings.Repeat("b", 64),
		}},
		"unexpected": true,
	}
	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "inputs.json")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := Aggregate(path, identity); err == nil || !strings.Contains(err.Error(), "input schema") {
		t.Fatalf("Aggregate() error = %v, want input schema rejection", err)
	}
}

func TestAggregateRejectsUnsafeProjectionPaths(t *testing.T) {
	identity := Identity{
		DesignLanguageVersion: "1.0", DesignLanguageSHA256: strings.Repeat("a", 64),
		SourceIdentity: "v1.3.0", ToolingVersion: "v1.3.0", PublicationStatus: "published",
	}
	for _, projection := range []string{"../outside.json", "/outside.json", `.\\outside.json`, "."} {
		t.Run(projection, func(t *testing.T) {
			root := t.TempDir()
			inputs := aggregateInputManifest{
				SchemaVersion: 1,
				DesignLanguage: aggregateDesignLanguage{
					Version: identity.DesignLanguageVersion,
					SHA256:  identity.DesignLanguageSHA256,
				},
				Repositories: []aggregateRepositoryInput{{
					Repository: "github.com/faustbrian/go-example",
					Projection: projection,
					SHA256:     strings.Repeat("b", 64),
				}},
			}
			data, err := json.Marshal(inputs)
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(root, "inputs.json")
			if err := os.WriteFile(path, data, 0o600); err != nil {
				t.Fatal(err)
			}

			if _, err := Aggregate(path, identity); err == nil {
				t.Fatal("Aggregate() error = nil, want unsafe-path rejection")
			}
		})
	}
}

func TestAggregateRejectsProjectionThroughIntermediateSymlinkOutsideManifestRoot(t *testing.T) {
	root, inputsPath, identity := writeValidAggregateInputs(t)
	outside := t.TempDir()
	projectionName := "repository.json"
	if err := os.Rename(filepath.Join(root, projectionName), filepath.Join(outside, projectionName)); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "projections")); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(inputsPath)
	if err != nil {
		t.Fatal(err)
	}
	var inputs aggregateInputManifest
	if err := json.Unmarshal(data, &inputs); err != nil {
		t.Fatal(err)
	}
	inputs.Repositories[0].Projection = "projections/" + projectionName
	data, err = json.Marshal(inputs)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(inputsPath, data, 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := Aggregate(inputsPath, identity); err == nil {
		t.Fatal("Aggregate(intermediate symlink escape) error = nil")
	}
}

func TestAggregateRejectsModuleFromAnotherRepository(t *testing.T) {
	root, inputsPath, identity := writeValidAggregateInputs(t)
	projectionPath := filepath.Join(root, "repository.json")
	data, err := os.ReadFile(projectionPath)
	if err != nil {
		t.Fatal(err)
	}
	var projection aggregateEngineeringEnvelope
	if err := json.Unmarshal(data, &projection); err != nil {
		t.Fatal(err)
	}
	projection.Modules[0].Repository = "github.com/faustbrian/go-other"
	data, err = json.Marshal(projection)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(projectionPath, data, 0o600); err != nil {
		t.Fatal(err)
	}

	manifestData, err := os.ReadFile(inputsPath)
	if err != nil {
		t.Fatal(err)
	}
	var inputs aggregateInputManifest
	if err := json.Unmarshal(manifestData, &inputs); err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(data)
	inputs.Repositories[0].SHA256 = hex.EncodeToString(digest[:])
	manifestData, err = json.Marshal(inputs)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(inputsPath, manifestData, 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := Aggregate(inputsPath, identity); err == nil || !strings.Contains(err.Error(), "module repository identity") {
		t.Fatalf("Aggregate() error = %v, want module repository identity rejection", err)
	}
}

func TestAggregateRejectsNonReleasableModuleFromAnotherRepository(t *testing.T) {
	root, inputsPath, identity := writeValidAggregateInputs(t)
	projectionPath := filepath.Join(root, "repository.json")
	data, err := os.ReadFile(projectionPath)
	if err != nil {
		t.Fatal(err)
	}
	var projection aggregateEngineeringEnvelope
	if err := json.Unmarshal(data, &projection); err != nil {
		t.Fatal(err)
	}
	projection.Modules[0].Repository = "github.com/faustbrian/go-other"
	projection.Modules[0].ModulePath = "example.com/analysis-coverage"
	projection.Modules[0].Kind = "fixture"
	projection.Modules[0].Releasable = false
	projection.Modules[0].Cohesion = nil
	data, err = json.Marshal(projection)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(projectionPath, data, 0o600); err != nil {
		t.Fatal(err)
	}

	manifestData, err := os.ReadFile(inputsPath)
	if err != nil {
		t.Fatal(err)
	}
	var inputs aggregateInputManifest
	if err := json.Unmarshal(manifestData, &inputs); err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(data)
	inputs.Repositories[0].SHA256 = hex.EncodeToString(digest[:])
	manifestData, err = json.Marshal(inputs)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(inputsPath, manifestData, 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := Aggregate(inputsPath, identity); err == nil || !strings.Contains(err.Error(), "module repository identity") {
		t.Fatalf("Aggregate() error = %v, want non-releasable module repository identity rejection", err)
	}
}

func TestAggregateRejectsRepositoryProjectionWithoutModules(t *testing.T) {
	root, inputsPath, identity := writeValidAggregateInputs(t)
	projectionPath := filepath.Join(root, "repository.json")
	data, err := os.ReadFile(projectionPath)
	if err != nil {
		t.Fatal(err)
	}
	var projection aggregateEngineeringEnvelope
	if err := json.Unmarshal(data, &projection); err != nil {
		t.Fatal(err)
	}
	projection.Modules = []engineeringModule{}
	data, err = json.Marshal(projection)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(projectionPath, data, 0o600); err != nil {
		t.Fatal(err)
	}
	manifestData, err := os.ReadFile(inputsPath)
	if err != nil {
		t.Fatal(err)
	}
	var inputs aggregateInputManifest
	if err := json.Unmarshal(manifestData, &inputs); err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(data)
	inputs.Repositories[0].SHA256 = hex.EncodeToString(digest[:])
	manifestData, err = json.Marshal(inputs)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(inputsPath, manifestData, 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := Aggregate(inputsPath, identity); err == nil || !strings.Contains(err.Error(), "contains no modules") {
		t.Fatalf("Aggregate() error = %v, want missing-module rejection", err)
	}
}

func TestAggregateRejectsModulePathOutsideRepository(t *testing.T) {
	root, inputsPath, identity := writeValidAggregateInputs(t)
	projectionPath := filepath.Join(root, "repository.json")
	data, err := os.ReadFile(projectionPath)
	if err != nil {
		t.Fatal(err)
	}
	var projection aggregateEngineeringEnvelope
	if err := json.Unmarshal(data, &projection); err != nil {
		t.Fatal(err)
	}
	projection.Modules[0].ModulePath = "github.com/faustbrian/go-other"
	data, err = json.Marshal(projection)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(projectionPath, data, 0o600); err != nil {
		t.Fatal(err)
	}
	manifestData, err := os.ReadFile(inputsPath)
	if err != nil {
		t.Fatal(err)
	}
	var inputs aggregateInputManifest
	if err := json.Unmarshal(manifestData, &inputs); err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(data)
	inputs.Repositories[0].SHA256 = hex.EncodeToString(digest[:])
	manifestData, err = json.Marshal(inputs)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(inputsPath, manifestData, 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := Aggregate(inputsPath, identity); err == nil || !strings.Contains(err.Error(), "module path identity") {
		t.Fatalf("Aggregate() error = %v, want module-path identity rejection", err)
	}
}

func TestAggregateAllowsNonReleasableModulePathsOutsideRepository(t *testing.T) {
	root, inputsPath, identity := writeValidAggregateInputs(t)
	projectionPath := filepath.Join(root, "repository.json")
	data, err := os.ReadFile(projectionPath)
	if err != nil {
		t.Fatal(err)
	}
	var projection aggregateEngineeringEnvelope
	if err := json.Unmarshal(data, &projection); err != nil {
		t.Fatal(err)
	}
	variants := []struct {
		directory  string
		modulePath string
		kind       string
	}{
		{directory: "testdata/coverage", modulePath: "example.com/analysis-coverage", kind: "fixture"},
		{directory: "benchmarks/load", modulePath: "example.com/analysis-benchmark", kind: "benchmark harness"},
	}
	for _, variant := range variants {
		module := projection.Modules[0]
		module.Directory = variant.directory
		module.ModulePath = variant.modulePath
		module.Kind = variant.kind
		module.Releasable = false
		module.Cohesion = nil
		projection.Modules = append(projection.Modules, module)
	}
	data, err = json.Marshal(projection)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(projectionPath, data, 0o600); err != nil {
		t.Fatal(err)
	}
	manifestData, err := os.ReadFile(inputsPath)
	if err != nil {
		t.Fatal(err)
	}
	var inputs aggregateInputManifest
	if err := json.Unmarshal(manifestData, &inputs); err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(data)
	inputs.Repositories[0].SHA256 = hex.EncodeToString(digest[:])
	manifestData, err = json.Marshal(inputs)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(inputsPath, manifestData, 0o600); err != nil {
		t.Fatal(err)
	}

	artifacts, err := Aggregate(inputsPath, identity)
	if err != nil {
		t.Fatalf("Aggregate() error = %v, want fixture module accepted", err)
	}
	for _, variant := range variants {
		if !bytes.Contains(artifacts.EngineeringJSON, []byte(`"module_path": "`+variant.modulePath+`"`)) {
			t.Fatalf("engineering catalog omitted %s module outside repository namespace", variant.kind)
		}
		if bytes.Contains(artifacts.ConsumerJSON, []byte(variant.modulePath)) {
			t.Fatalf("consumer catalog included non-releasable %s module", variant.kind)
		}
	}
}

func TestAggregateRejectsOversizedInputsAndProjections(t *testing.T) {
	identity := Identity{
		DesignLanguageVersion: "1.0", DesignLanguageSHA256: strings.Repeat("a", 64),
		SourceIdentity: "v1.3.0", ToolingVersion: "v1.3.0", PublicationStatus: "published",
	}
	t.Run("inputs", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "inputs.json")
		file, err := os.Create(path)
		if err != nil {
			t.Fatal(err)
		}
		if err := file.Truncate((1 << 20) + 1); err != nil {
			_ = file.Close()
			t.Fatal(err)
		}
		if err := file.Close(); err != nil {
			t.Fatal(err)
		}
		if _, err := Aggregate(path, identity); err == nil || !strings.Contains(err.Error(), "exceed maximum size") {
			t.Fatalf("Aggregate() error = %v, want bounded-input rejection", err)
		}
	})
	t.Run("projection", func(t *testing.T) {
		root := t.TempDir()
		projectionName := "repository.json"
		projectionPath := filepath.Join(root, projectionName)
		file, err := os.Create(projectionPath)
		if err != nil {
			t.Fatal(err)
		}
		if err := file.Truncate((32 << 20) + 1); err != nil {
			_ = file.Close()
			t.Fatal(err)
		}
		if err := file.Close(); err != nil {
			t.Fatal(err)
		}
		inputs := aggregateInputManifest{
			SchemaVersion: 1,
			DesignLanguage: aggregateDesignLanguage{
				Version: identity.DesignLanguageVersion,
				SHA256:  identity.DesignLanguageSHA256,
			},
			Repositories: []aggregateRepositoryInput{{
				Repository: "github.com/faustbrian/go-example",
				Projection: projectionName,
				SHA256:     strings.Repeat("b", 64),
			}},
		}
		data, err := json.Marshal(inputs)
		if err != nil {
			t.Fatal(err)
		}
		inputsPath := filepath.Join(root, "inputs.json")
		if err := os.WriteFile(inputsPath, data, 0o600); err != nil {
			t.Fatal(err)
		}
		if _, err := Aggregate(inputsPath, identity); err == nil || !strings.Contains(err.Error(), "exceed maximum size") {
			t.Fatalf("Aggregate() error = %v, want bounded-projection rejection", err)
		}
	})
}

func TestReadBoundedAggregateFileReportsEveryBoundary(t *testing.T) {
	failure := errors.New("injected failure")
	for name, files := range map[string]aggregateInputFiles{
		"inspect": &fakeAggregateInputFiles{lstatErr: failure},
		"regular": &fakeAggregateInputFiles{info: fakeCatalogInfo{mode: os.ModeDir}},
		"size":    &fakeAggregateInputFiles{info: fakeCatalogInfo{mode: 0o600, size: 4}},
		"open":    &fakeAggregateInputFiles{info: fakeCatalogInfo{mode: 0o600, size: 1}, openErr: failure},
		"read": &fakeAggregateInputFiles{
			info: fakeCatalogInfo{mode: 0o600, size: 1}, file: &fakeAggregateReader{err: failure, info: fakeCatalogInfo{mode: 0o600, size: 1}},
		},
		"grew": &fakeAggregateInputFiles{
			info: fakeCatalogInfo{mode: 0o600, size: 1}, file: &fakeAggregateReader{reader: strings.NewReader("abcd"), info: fakeCatalogInfo{mode: 0o600, size: 1}},
		},
		"stat": &fakeAggregateInputFiles{
			info: fakeCatalogInfo{mode: 0o600, size: 1}, file: &fakeAggregateReader{statErr: failure},
		},
		"changed": &fakeAggregateInputFiles{
			info: fakeCatalogInfo{mode: 0o600, size: 1}, file: &fakeAggregateReader{info: fakeCatalogInfo{mode: os.ModeDir}},
		},
		"replaced": &fakeAggregateInputFiles{
			info: fakeCatalogInfo{mode: 0o600, size: 1}, file: &fakeAggregateReader{info: fakeCatalogInfo{mode: 0o600, size: 1}}, different: true,
		},
		"opened size": &fakeAggregateInputFiles{
			info: fakeCatalogInfo{mode: 0o600, size: 1}, file: &fakeAggregateReader{info: fakeCatalogInfo{mode: 0o600, size: 4}},
		},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := readBoundedAggregateFileWithFiles("input.json", 3, "input", files); err == nil {
				t.Fatal("readBoundedAggregateFileWithFiles() error = nil")
			}
		})
	}
	data, err := readBoundedAggregateFileWithFiles("input.json", 3, "input", &fakeAggregateInputFiles{
		info: fakeCatalogInfo{mode: 0o600, size: 3}, file: &fakeAggregateReader{reader: strings.NewReader("abc"), info: fakeCatalogInfo{mode: 0o600, size: 3}},
	})
	if err != nil || string(data) != "abc" {
		t.Fatalf("readBoundedAggregateFileWithFiles() = %q, %v", data, err)
	}
	if _, err := readBoundedAggregateFileInRoot(filepath.Join(t.TempDir(), "missing"), "input.json", 3, "input"); err == nil {
		t.Fatal("readBoundedAggregateFileInRoot(missing root) error = nil")
	}
}

func TestGenerateAggregateWritesTheFourCanonicalCatalogArtifacts(t *testing.T) {
	root, inputsPath, identity := writeValidAggregateInputs(t)
	output := filepath.Join(root, "output")
	want, err := Aggregate(inputsPath, identity)
	if err != nil {
		t.Fatal(err)
	}

	if err := GenerateAggregate(inputsPath, output, identity); err != nil {
		t.Fatal(err)
	}
	for name, expected := range map[string][]byte{
		"catalog-consumer.json":    want.ConsumerJSON,
		"catalog-consumer.md":      want.ConsumerMarkdown,
		"catalog-engineering.json": want.EngineeringJSON,
		"catalog-engineering.md":   want.EngineeringMarkdown,
	} {
		actual, err := os.ReadFile(filepath.Join(output, name))
		if err != nil {
			t.Fatal(err)
		}
		if string(actual) != string(expected) {
			t.Fatalf("%s is stale", name)
		}
	}
	entries, err := os.ReadDir(output)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 4 {
		t.Fatalf("generated entries = %d, want 4", len(entries))
	}
}

func TestGenerateAggregateDoesNotPartiallyPublishTheCatalogSet(t *testing.T) {
	root, inputsPath, identity := writeValidAggregateInputs(t)
	output := filepath.Join(root, "output")
	if err := os.Mkdir(output, 0o755); err != nil {
		t.Fatal(err)
	}
	old := []byte("previous catalog\n")
	for _, name := range []string{"catalog-consumer.json", "catalog-consumer.md"} {
		if err := os.WriteFile(filepath.Join(output, name), old, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Mkdir(filepath.Join(output, "catalog-engineering.json"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := GenerateAggregate(inputsPath, output, identity); err == nil {
		t.Fatal("GenerateAggregate() error = nil")
	}
	for _, name := range []string{"catalog-consumer.json", "catalog-consumer.md"} {
		actual, err := os.ReadFile(filepath.Join(output, name))
		if err != nil {
			t.Fatal(err)
		}
		if string(actual) != string(old) {
			t.Fatalf("%s was partially published", name)
		}
	}
}

func TestGenerateAggregateProtectedKeepsWritesBoundToOpenedOutputDirectory(t *testing.T) {
	root, inputsPath, identity := writeValidAggregateInputs(t)
	output := filepath.Join(root, "preview")
	protected := filepath.Join(root, "docs", "ecosystem")
	if err := os.MkdirAll(protected, 0o755); err != nil {
		t.Fatal(err)
	}
	protectedMarker := filepath.Join(protected, "marker")
	if err := os.WriteFile(protectedMarker, []byte("protected"), 0o600); err != nil {
		t.Fatal(err)
	}
	movedOutput := filepath.Join(root, "opened-preview")
	afterOpen := func() error {
		if err := os.Rename(output, movedOutput); err != nil {
			return err
		}
		return os.Symlink(protected, output)
	}

	if err := generateAggregateProtected(inputsPath, output, protected, identity, afterOpen); err != nil {
		t.Fatal(err)
	}
	for _, artifact := range aggregateArtifactNames {
		if _, err := os.Stat(filepath.Join(movedOutput, artifact.name)); err != nil {
			t.Fatalf("opened output missing %s: %v", artifact.name, err)
		}
		if _, err := os.Stat(filepath.Join(protected, artifact.name)); !os.IsNotExist(err) {
			t.Fatalf("protected output contains %s: %v", artifact.name, err)
		}
	}
	marker, err := os.ReadFile(protectedMarker)
	if err != nil || string(marker) != "protected" {
		t.Fatalf("protected marker = %q, %v", marker, err)
	}
}

func TestGenerateAggregateProtectedRejectsProtectedDirectoryIdentity(t *testing.T) {
	root, inputsPath, identity := writeValidAggregateInputs(t)
	protected := filepath.Join(root, "docs", "ecosystem")
	if err := GenerateAggregateProtected(inputsPath, protected, protected, identity); err == nil || !strings.Contains(err.Error(), "source build cannot publish") {
		t.Fatalf("GenerateAggregateProtected() error = %v", err)
	}
}

func TestGenerateAggregateProtectedReportsEveryBoundary(t *testing.T) {
	root, inputsPath, identity := writeValidAggregateInputs(t)
	protected := filepath.Join(root, "protected")
	if err := os.Mkdir(protected, 0o755); err != nil {
		t.Fatal(err)
	}
	failure := errors.New("injected failure")
	newOperations := func() protectedAggregateOperations {
		return protectedAggregateOperations{
			mkdirAll: os.MkdirAll,
			openRoot: os.OpenRoot,
			statRoot: func(root *os.Root) (os.FileInfo, error) { return root.Stat(".") },
			random:   rand.Reader,
		}
	}
	for name, mutate := range map[string]func(*protectedAggregateOperations){
		"create output": func(operations *protectedAggregateOperations) {
			operations.mkdirAll = func(string, os.FileMode) error { return failure }
		},
		"open output": func(operations *protectedAggregateOperations) {
			operations.openRoot = func(string) (*os.Root, error) { return nil, failure }
		},
		"after open": func(operations *protectedAggregateOperations) {
			operations.afterOpen = func() error { return failure }
		},
		"open protected": func(operations *protectedAggregateOperations) {
			calls := 0
			operations.openRoot = func(path string) (*os.Root, error) {
				calls++
				if calls == 2 {
					return nil, failure
				}
				return os.OpenRoot(path)
			}
		},
		"inspect output": func(operations *protectedAggregateOperations) {
			operations.statRoot = func(*os.Root) (os.FileInfo, error) { return nil, failure }
		},
		"inspect protected": func(operations *protectedAggregateOperations) {
			calls := 0
			operations.statRoot = func(root *os.Root) (os.FileInfo, error) {
				calls++
				if calls == 2 {
					return nil, failure
				}
				return root.Stat(".")
			}
		},
	} {
		t.Run(name, func(t *testing.T) {
			operations := newOperations()
			mutate(&operations)
			output := filepath.Join(root, "output-"+strings.ReplaceAll(name, " ", "-"))
			if err := generateAggregateProtectedWithOperations(inputsPath, output, protected, identity, operations); err == nil {
				t.Fatal("generateAggregateProtectedWithOperations() error = nil")
			}
		})
	}
	operations := newOperations()
	if err := generateAggregateProtectedWithOperations(filepath.Join(root, "missing-inputs.json"), filepath.Join(root, "invalid-input"), protected, identity, operations); err == nil {
		t.Fatal("generateAggregateProtectedWithOperations(invalid inputs) error = nil")
	}
	if err := generateAggregateProtectedWithOperations(inputsPath, filepath.Join(root, "missing-protected-output"), filepath.Join(root, "missing-protected"), identity, newOperations()); err != nil {
		t.Fatalf("generateAggregateProtectedWithOperations(missing protected) error = %v", err)
	}
}

func TestRootedCatalogFilesRejectEveryEscapingAndTemporaryBoundary(t *testing.T) {
	directory := t.TempDir()
	root, err := os.OpenRoot(directory)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	files := &rootedCatalogFiles{root: root, directory: directory, random: rand.Reader}
	outside := filepath.Join(filepath.Dir(directory), "outside")
	failure := errors.New("injected failure")
	if _, err := rootedCatalogRelative(directory, "path", func(string, string) (string, error) { return "", failure }); !errors.Is(err, failure) {
		t.Fatalf("rootedCatalogRelative() error = %v", err)
	}
	if _, err := files.relative(outside); err == nil {
		t.Fatal("relative(outside) error = nil")
	}
	if _, err := files.relative(directory); err == nil {
		t.Fatal("relative(output directory) error = nil")
	}
	if err := files.MkdirAll(outside, 0o755); err == nil {
		t.Fatal("MkdirAll(outside) error = nil")
	}
	if _, err := files.Lstat(outside); err == nil {
		t.Fatal("Lstat(outside) error = nil")
	}
	if _, err := files.CreateTemp(outside, ".catalog-*"); err == nil {
		t.Fatal("CreateTemp(outside) error = nil")
	}
	files.random = strings.NewReader("")
	if _, err := files.CreateTemp(directory, ".catalog-*"); err == nil {
		t.Fatal("CreateTemp(random failure) error = nil")
	}
	files.random = strings.NewReader(strings.Repeat("\x00", 16*100))
	collision := filepath.Join(directory, ".catalog-"+strings.Repeat("0", 32))
	if err := os.WriteFile(collision, []byte("collision"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := files.CreateTemp(directory, ".catalog-*"); err == nil || !strings.Contains(err.Error(), "unique") {
		t.Fatalf("CreateTemp(collisions) error = %v", err)
	}
	files.random = strings.NewReader(strings.Repeat("\x00", 16) + strings.Repeat("\x01", 16))
	created, err := files.CreateTemp(directory, ".catalog-*")
	if err != nil || !strings.Contains(created.Name(), strings.Repeat("01", 16)) {
		t.Fatalf("CreateTemp(collision then unique) = %v, %v", created, err)
	}
	if err := created.Close(); err != nil {
		t.Fatal(err)
	}
	if err := files.Remove(created.Name()); err != nil {
		t.Fatal(err)
	}
	files.random = strings.NewReader(strings.Repeat("\x01", 16))
	if _, err := files.CreateTemp(directory, "/absolute-*"); err == nil {
		t.Fatal("CreateTemp(absolute name) error = nil")
	}
	inside := filepath.Join(directory, "inside")
	if err := os.WriteFile(inside, []byte("inside"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := files.Rename(outside, inside); err == nil {
		t.Fatal("Rename(outside source) error = nil")
	}
	if err := files.Rename(inside, outside); err == nil {
		t.Fatal("Rename(outside target) error = nil")
	}
	if err := files.Remove(outside); err == nil || !strings.Contains(err.Error(), "escapes output directory") {
		t.Fatalf("Remove(outside) error = %v", err)
	}
}

func TestGenerateAggregateReportsEveryPublicationBoundary(t *testing.T) {
	_, inputsPath, identity := writeValidAggregateInputs(t)
	failure := errors.New("injected failure")
	for name, files := range map[string]*fakeCatalogFiles{
		"create output": {mkdirErr: failure},
		"non regular target": {
			lstatInfos: map[int]os.FileInfo{1: fakeCatalogInfo{mode: os.ModeDir}},
		},
		"inspect target": {
			lstatErrors: map[int]error{1: failure},
		},
		"stage later target": {
			createErrors: map[int]error{2: failure},
		},
		"publish set": {
			renameErrors: map[int]error{1: failure},
		},
	} {
		t.Run(name, func(t *testing.T) {
			if err := generateAggregateWithFiles(inputsPath, "/output", identity, files); err == nil {
				t.Fatal("generateAggregateWithFiles() error = nil")
			}
		})
	}
	if err := generateAggregateWithFiles(filepath.Join(t.TempDir(), "missing.json"), "/output", identity, &fakeCatalogFiles{}); err == nil {
		t.Fatal("generateAggregateWithFiles(invalid inputs) error = nil")
	}
}

func TestStageCatalogRejectsShortWrite(t *testing.T) {
	files := &fakeCatalogFiles{file: &fakeCatalogFile{writeN: 1}}
	if _, err := stageCatalog("/catalog.json", []byte("catalog"), false, files); !errors.Is(err, io.ErrShortWrite) {
		t.Fatalf("stageCatalog() error = %v, want short-write rejection", err)
	}
}

func TestStageCatalogReportsEveryFileBoundary(t *testing.T) {
	failure := errors.New("injected failure")
	for name, files := range map[string]*fakeCatalogFiles{
		"create": {createErr: failure},
		"chmod":  {file: &fakeCatalogFile{chmodErr: failure}},
		"write":  {file: &fakeCatalogFile{writeErr: failure}},
		"sync":   {file: &fakeCatalogFile{syncErr: failure}},
		"close":  {file: &fakeCatalogFile{closeErr: failure}},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := stageCatalog("/catalog.json", []byte("catalog"), true, files); err == nil {
				t.Fatal("stageCatalog() error = nil")
			}
		})
	}
	files := &fakeCatalogFiles{file: &fakeCatalogFile{}}
	entry, err := stageCatalog("/catalog.json", []byte("catalog"), true, files)
	if err != nil {
		t.Fatal(err)
	}
	if entry.target != "/catalog.json" || entry.temporary != "/temporary-catalog" || !entry.existed {
		t.Fatalf("stageCatalog() = %#v", entry)
	}
}

func TestPublishCatalogSetReportsEveryTransactionBoundary(t *testing.T) {
	failure := errors.New("injected failure")
	stagedExisting := []stagedCatalog{{target: "/catalog.json", temporary: "/catalog.tmp", existed: true}}
	for name, files := range map[string]*fakeCatalogFiles{
		"reserve backup": {createErrors: map[int]error{1: failure}},
		"close backup": {
			files: []*fakeCatalogFile{{name: "/backup", closeErr: failure}},
		},
		"prepare backup": {
			files:        []*fakeCatalogFile{{name: "/backup"}},
			removeErrors: map[int]error{1: failure},
		},
		"back up target": {
			files:        []*fakeCatalogFile{{name: "/backup"}},
			renameErrors: map[int]error{1: failure},
		},
		"publish staged": {
			files:        []*fakeCatalogFile{{name: "/backup"}},
			renameErrors: map[int]error{2: failure},
		},
		"remove backup": {
			files:        []*fakeCatalogFile{{name: "/backup"}},
			removeErrors: map[int]error{2: failure},
		},
	} {
		t.Run(name, func(t *testing.T) {
			candidate := append([]stagedCatalog(nil), stagedExisting...)
			if err := publishCatalogSet(candidate, files); err == nil {
				t.Fatal("publishCatalogSet() error = nil")
			}
		})
	}

	files := &fakeCatalogFiles{files: []*fakeCatalogFile{{name: "/backup"}}}
	if err := publishCatalogSet(append([]stagedCatalog(nil), stagedExisting...), files); err != nil {
		t.Fatal(err)
	}
	withoutExisting := []stagedCatalog{{target: "/new.json", temporary: "/new.tmp"}}
	if err := publishCatalogSet(withoutExisting, &fakeCatalogFiles{}); err != nil {
		t.Fatal(err)
	}
	mixed := []stagedCatalog{
		{target: "/new.json", temporary: "/new.tmp"},
		{target: "/existing.json", temporary: "/existing.tmp", existed: true},
	}
	mixedFiles := &fakeCatalogFiles{files: []*fakeCatalogFile{{name: "/backup"}}}
	if err := publishCatalogSet(mixed, mixedFiles); err != nil || mixedFiles.createCalls != 1 || mixedFiles.renameCalls != 3 {
		t.Fatalf("publishCatalogSet(mixed) calls = create %d, rename %d, error %v", mixedFiles.createCalls, mixedFiles.renameCalls, err)
	}
}

func TestRollbackCatalogSetReportsRemovalAndRestoreFailures(t *testing.T) {
	failure := errors.New("injected failure")
	staged := []stagedCatalog{{
		target: "/catalog.json", temporary: "/catalog.tmp", backup: "/catalog.backup", published: true,
	}}
	files := &fakeCatalogFiles{
		removeErrors: map[int]error{1: failure},
		renameErrors: map[int]error{1: failure},
	}
	err := rollbackCatalogSet(append([]stagedCatalog(nil), staged...), files, failure)
	if err == nil || !strings.Contains(err.Error(), "remove partially published") || !strings.Contains(err.Error(), "restore cohesion catalog") {
		t.Fatalf("rollbackCatalogSet() error = %v", err)
	}
	files = &fakeCatalogFiles{removeErrors: map[int]error{1: os.ErrNotExist}}
	if err := rollbackCatalogSet(append([]stagedCatalog(nil), staged...), files, failure); !errors.Is(err, failure) || strings.Contains(err.Error(), "remove partially published") {
		t.Fatalf("rollbackCatalogSet(not exist) error = %v", err)
	}
	cleanupFiles := &fakeCatalogFiles{}
	cleanupStagedCatalogs([]stagedCatalog{{temporary: "/temporary"}, {}}, cleanupFiles)
	if len(cleanupFiles.removed) != 1 || cleanupFiles.removed[0] != "/temporary" {
		t.Fatalf("cleanupStagedCatalogs() removed = %v", cleanupFiles.removed)
	}
}

func TestCheckAggregateByteComparesRenderedArtifactsWithoutWriting(t *testing.T) {
	root, inputsPath, identity := writeValidAggregateInputs(t)
	output := filepath.Join(root, "output")
	if err := GenerateAggregate(inputsPath, output, identity); err != nil {
		t.Fatal(err)
	}
	if err := CheckAggregate(inputsPath, output, identity); err != nil {
		t.Fatalf("CheckAggregate(current) error = %v", err)
	}
	stalePath := filepath.Join(output, "catalog-consumer.md")
	stale := []byte("stale\n")
	if err := os.WriteFile(stalePath, stale, 0o600); err != nil {
		t.Fatal(err)
	}

	if err := CheckAggregate(inputsPath, output, identity); err == nil || !strings.Contains(err.Error(), "catalog-consumer.md") {
		t.Fatalf("CheckAggregate(stale) error = %v", err)
	}
	actual, err := os.ReadFile(stalePath)
	if err != nil {
		t.Fatal(err)
	}
	if string(actual) != string(stale) {
		t.Fatalf("CheckAggregate rewrote stale artifact to %q", actual)
	}
	if err := CheckAggregate(filepath.Join(root, "missing-inputs.json"), output, identity); err == nil {
		t.Fatal("CheckAggregate(missing inputs) error = nil")
	}
}

func TestCheckAggregateAcceptsCombinedArtifactLargerThanOneProjection(t *testing.T) {
	artifacts := Artifacts{
		ConsumerJSON: []byte("consumer-json"), ConsumerMarkdown: []byte("consumer-markdown"),
		EngineeringJSON: []byte("engineering-json"), EngineeringMarkdown: []byte("engineering-markdown"),
	}
	reads := 0
	err := checkAggregateArtifacts("output", artifacts, int64(len(artifacts.EngineeringMarkdown)), func(path string, maximum int64, _ string) ([]byte, error) {
		reads++
		if maximum != int64(len(artifacts.EngineeringMarkdown)) {
			t.Fatalf("artifact read limit = %d, want %d", maximum, len(artifacts.EngineeringMarkdown))
		}
		for _, artifact := range aggregateArtifactNames {
			if filepath.Base(path) == artifact.name {
				return artifact.content(artifacts), nil
			}
		}
		return nil, os.ErrNotExist
	})
	if err != nil || reads != len(aggregateArtifactNames) {
		t.Fatalf("checkAggregateArtifacts() reads = %d, error = %v", reads, err)
	}
}

func TestCheckAggregateRejectsSymlinkedCatalogArtifact(t *testing.T) {
	root, inputsPath, identity := writeValidAggregateInputs(t)
	output := filepath.Join(root, "output")
	if err := GenerateAggregate(inputsPath, output, identity); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(output, "catalog-consumer.json")
	content, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	copyPath := filepath.Join(root, "catalog-copy.json")
	if err := os.WriteFile(copyPath, content, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(target); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(copyPath, target); err != nil {
		t.Fatal(err)
	}

	if err := CheckAggregate(inputsPath, output, identity); err == nil || !strings.Contains(err.Error(), "catalog-consumer.json") {
		t.Fatalf("CheckAggregate(symlink) error = %v", err)
	}
}

func TestAggregateHelperBoundaries(t *testing.T) {
	identity := Identity{
		DesignLanguageVersion: "1.0", DesignLanguageSHA256: strings.Repeat("a", 64),
		SourceIdentity: "v1.3.0", ToolingVersion: "v1.3.0", PublicationStatus: "published",
	}
	repository := "github.com/faustbrian/go-example"
	input := aggregateRepositoryInput{Repository: repository}
	validProjection := aggregateEngineeringEnvelope{
		SchemaVersion: 1, View: "engineering", Scope: "repository", Repository: &repository,
		DesignLanguage: DesignLanguageIdentity{
			Version: identity.DesignLanguageVersion, SHA256: identity.DesignLanguageSHA256,
			SourceIdentity: identity.SourceIdentity,
		},
		Tooling: ToolingIdentity{Version: identity.ToolingVersion, PublicationStatus: identity.PublicationStatus},
	}
	if err := validateAggregateProjection(input, validProjection, identity); err != nil {
		t.Fatal(err)
	}
	for name, mutate := range map[string]func(*aggregateEngineeringEnvelope){
		"schema":             func(value *aggregateEngineeringEnvelope) { value.SchemaVersion = 2 },
		"view":               func(value *aggregateEngineeringEnvelope) { value.View = "consumer" },
		"scope":              func(value *aggregateEngineeringEnvelope) { value.Scope = "ecosystem" },
		"missing repository": func(value *aggregateEngineeringEnvelope) { value.Repository = nil },
		"repository":         func(value *aggregateEngineeringEnvelope) { other := "other"; value.Repository = &other },
		"design version":     func(value *aggregateEngineeringEnvelope) { value.DesignLanguage.Version = "2.0" },
		"design digest":      func(value *aggregateEngineeringEnvelope) { value.DesignLanguage.SHA256 = strings.Repeat("b", 64) },
		"source identity":    func(value *aggregateEngineeringEnvelope) { value.DesignLanguage.SourceIdentity = "v9.9.9" },
		"tooling version":    func(value *aggregateEngineeringEnvelope) { value.Tooling.Version = "v9.9.9" },
		"publication status": func(value *aggregateEngineeringEnvelope) { value.Tooling.PublicationStatus = "unpublished" },
	} {
		t.Run(name, func(t *testing.T) {
			projection := validProjection
			mutate(&projection)
			if err := validateAggregateProjection(input, projection, identity); err == nil {
				t.Fatal("validateAggregateProjection() error = nil")
			}
		})
	}

	left := engineeringProjection(repository, validAggregateModule())
	right := engineeringProjection(repository, validAggregateModule())
	left.Cohesion.Family = "foundations"
	right.Cohesion.Family = "resilience"
	if compareEngineeringModules(left, right) >= 0 {
		t.Fatal("compareEngineeringModules() did not apply family order")
	}

	valid := engineeringProjection(repository, validAggregateModule())
	nonReleasable := valid
	nonReleasable.Releasable = false
	unclassified := valid
	unclassified.Cohesion = nil
	wrongKind := valid
	wrongKind.Kind = "fixture"
	planned := valid
	plannedCohesion := *valid.Cohesion
	planned.Cohesion = &plannedCohesion
	planned.Cohesion.LifecycleStatus = "planned"
	if got := aggregateConsumerModules([]engineeringModule{nonReleasable, unclassified, wrongKind, planned, valid}); len(got) != 1 {
		t.Fatalf("aggregateConsumerModules() length = %d, want 1", len(got))
	}
	if _, err := marshalCatalog(Envelope{Modules: func() {}}); err == nil {
		t.Fatal("marshalCatalog(unsupported value) error = nil")
	}
}

func TestAggregateWithOperationsReportsEveryBoundary(t *testing.T) {
	failure := errors.New("injected failure")
	identity := Identity{
		DesignLanguageVersion: "1.0", DesignLanguageSHA256: strings.Repeat("a", 64),
		SourceIdentity: "v1.3.0", ToolingVersion: "v1.3.0", PublicationStatus: "published",
	}
	repository := "github.com/faustbrian/go-example"
	projectionData := []byte("projection")
	digest := sha256.Sum256(projectionData)
	validInputs := aggregateInputManifest{
		SchemaVersion: 1,
		DesignLanguage: aggregateDesignLanguage{
			Version: identity.DesignLanguageVersion, SHA256: identity.DesignLanguageSHA256,
		},
		Repositories: []aggregateRepositoryInput{{
			Repository: repository, Projection: "repository.json", SHA256: hex.EncodeToString(digest[:]),
		}},
	}
	validProjection := aggregateEngineeringEnvelope{
		SchemaVersion: 1, View: "engineering", Scope: "repository", Repository: &repository,
		DesignLanguage: DesignLanguageIdentity{
			Version: identity.DesignLanguageVersion, SHA256: identity.DesignLanguageSHA256,
			SourceIdentity: identity.SourceIdentity,
		},
		Tooling: ToolingIdentity{Version: identity.ToolingVersion, PublicationStatus: identity.PublicationStatus},
		Modules: []engineeringModule{engineeringProjection(repository, aggregateModuleForRepository(repository))},
	}
	newOperations := func() aggregateOperations {
		return aggregateOperations{
			limits: defaultAggregateLimits(),
			read: func(string, int64, string) ([]byte, error) {
				return []byte("inputs"), nil
			},
			readProjection:     func(string, string, int64, string) ([]byte, error) { return projectionData, nil },
			validateInputs:     func([]byte) error { return nil },
			decodeInputs:       func([]byte, *aggregateInputManifest) error { return nil },
			validateCatalog:    func([]byte) error { return nil },
			decodeProjection:   func([]byte, *aggregateEngineeringEnvelope) error { return nil },
			validateProjection: validateAggregateProjection,
			marshal:            func(Envelope) ([]byte, error) { return []byte("json"), nil },
			render:             func(Envelope) ([]byte, error) { return []byte("markdown"), nil },
		}
	}
	run := func(t *testing.T, mutate func(*aggregateOperations, *aggregateInputManifest, *aggregateEngineeringEnvelope)) error {
		t.Helper()
		operations := newOperations()
		inputs := validInputs
		inputs.Repositories = append([]aggregateRepositoryInput(nil), validInputs.Repositories...)
		projection := validProjection
		projection.Modules = append([]engineeringModule(nil), validProjection.Modules...)
		mutate(&operations, &inputs, &projection)
		originalDecodeInputs := operations.decodeInputs
		operations.decodeInputs = func(data []byte, target *aggregateInputManifest) error {
			if err := originalDecodeInputs(data, target); err != nil {
				return err
			}
			*target = inputs
			return nil
		}
		originalDecodeProjection := operations.decodeProjection
		operations.decodeProjection = func(data []byte, target *aggregateEngineeringEnvelope) error {
			if err := originalDecodeProjection(data, target); err != nil {
				return err
			}
			*target = projection
			return nil
		}
		_, err := aggregateWithOperations("inputs.json", identity, operations)
		return err
	}

	cases := map[string]func(*aggregateOperations, *aggregateInputManifest, *aggregateEngineeringEnvelope){
		"read inputs": func(operations *aggregateOperations, _ *aggregateInputManifest, _ *aggregateEngineeringEnvelope) {
			operations.read = func(string, int64, string) ([]byte, error) { return nil, failure }
		},
		"read projection": func(operations *aggregateOperations, _ *aggregateInputManifest, _ *aggregateEngineeringEnvelope) {
			operations.readProjection = func(string, string, int64, string) ([]byte, error) { return nil, failure }
		},
		"validate inputs": func(operations *aggregateOperations, _ *aggregateInputManifest, _ *aggregateEngineeringEnvelope) {
			operations.validateInputs = func([]byte) error { return failure }
		},
		"decode inputs": func(operations *aggregateOperations, _ *aggregateInputManifest, _ *aggregateEngineeringEnvelope) {
			operations.decodeInputs = func([]byte, *aggregateInputManifest) error { return failure }
		},
		"schema version": func(_ *aggregateOperations, inputs *aggregateInputManifest, _ *aggregateEngineeringEnvelope) {
			inputs.SchemaVersion = 2
		},
		"design language": func(_ *aggregateOperations, inputs *aggregateInputManifest, _ *aggregateEngineeringEnvelope) {
			inputs.DesignLanguage.SHA256 = strings.Repeat("b", 64)
		},
		"missing repositories": func(_ *aggregateOperations, inputs *aggregateInputManifest, _ *aggregateEngineeringEnvelope) {
			inputs.Repositories = nil
		},
		"repository count": func(operations *aggregateOperations, _ *aggregateInputManifest, _ *aggregateEngineeringEnvelope) {
			operations.limits.maximumRepositories = 0
		},
		"total projection bytes": func(operations *aggregateOperations, _ *aggregateInputManifest, _ *aggregateEngineeringEnvelope) {
			operations.limits.maximumProjectionBytes = int64(len(projectionData) - 1)
		},
		"module count": func(operations *aggregateOperations, _ *aggregateInputManifest, _ *aggregateEngineeringEnvelope) {
			operations.limits.maximumModules = 0
		},
		"missing repository identity": func(_ *aggregateOperations, inputs *aggregateInputManifest, _ *aggregateEngineeringEnvelope) {
			inputs.Repositories[0].Repository = ""
		},
		"duplicate repositories": func(_ *aggregateOperations, inputs *aggregateInputManifest, _ *aggregateEngineeringEnvelope) {
			inputs.Repositories = append(inputs.Repositories, inputs.Repositories[0])
		},
		"unsafe projection": func(_ *aggregateOperations, inputs *aggregateInputManifest, _ *aggregateEngineeringEnvelope) {
			inputs.Repositories[0].Projection = "../repository.json"
		},
		"projection digest": func(_ *aggregateOperations, inputs *aggregateInputManifest, _ *aggregateEngineeringEnvelope) {
			inputs.Repositories[0].SHA256 = strings.Repeat("b", 64)
		},
		"validate projection schema": func(operations *aggregateOperations, _ *aggregateInputManifest, _ *aggregateEngineeringEnvelope) {
			operations.validateCatalog = func([]byte) error { return failure }
		},
		"decode projection": func(operations *aggregateOperations, _ *aggregateInputManifest, _ *aggregateEngineeringEnvelope) {
			operations.decodeProjection = func([]byte, *aggregateEngineeringEnvelope) error { return failure }
		},
		"validate projection identity": func(operations *aggregateOperations, _ *aggregateInputManifest, _ *aggregateEngineeringEnvelope) {
			operations.validateProjection = func(aggregateRepositoryInput, aggregateEngineeringEnvelope, Identity) error { return failure }
		},
		"duplicate module": func(_ *aggregateOperations, _ *aggregateInputManifest, projection *aggregateEngineeringEnvelope) {
			projection.Modules = append(projection.Modules, projection.Modules[0])
		},
		"marshal consumer": func(operations *aggregateOperations, _ *aggregateInputManifest, _ *aggregateEngineeringEnvelope) {
			operations.marshal = func(Envelope) ([]byte, error) { return nil, failure }
		},
		"marshal engineering": func(operations *aggregateOperations, _ *aggregateInputManifest, _ *aggregateEngineeringEnvelope) {
			calls := 0
			operations.marshal = func(Envelope) ([]byte, error) {
				calls++
				if calls == 2 {
					return nil, failure
				}
				return []byte("json"), nil
			}
		},
		"render consumer": func(operations *aggregateOperations, _ *aggregateInputManifest, _ *aggregateEngineeringEnvelope) {
			operations.render = func(Envelope) ([]byte, error) { return nil, failure }
		},
		"render engineering": func(operations *aggregateOperations, _ *aggregateInputManifest, _ *aggregateEngineeringEnvelope) {
			calls := 0
			operations.render = func(Envelope) ([]byte, error) {
				calls++
				if calls == 2 {
					return nil, failure
				}
				return []byte("markdown"), nil
			}
		},
		"artifact size": func(operations *aggregateOperations, _ *aggregateInputManifest, _ *aggregateEngineeringEnvelope) {
			operations.limits.maximumArtifactBytes = int64(len("markdown") - 1)
		},
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			if err := run(t, mutate); err == nil {
				t.Fatal("aggregateWithOperations() error = nil")
			}
		})
	}
	if err := run(t, func(_ *aggregateOperations, inputs *aggregateInputManifest, _ *aggregateEngineeringEnvelope) {
		inputs.Repositories = append(inputs.Repositories, inputs.Repositories[0])
	}); err == nil || !strings.Contains(err.Error(), "repositories must be unique and sorted") {
		t.Fatalf("aggregateWithOperations(duplicate repositories) error = %v", err)
	}
	if _, err := aggregateWithOperations("inputs.json", Identity{}, newOperations()); err == nil {
		t.Fatal("aggregateWithOperations(invalid identity) error = nil")
	}
	if err := run(t, func(*aggregateOperations, *aggregateInputManifest, *aggregateEngineeringEnvelope) {}); err != nil {
		t.Fatal(err)
	}
	if err := run(t, func(operations *aggregateOperations, _ *aggregateInputManifest, _ *aggregateEngineeringEnvelope) {
		operations.limits = aggregateLimits{
			maximumRepositories:    1,
			maximumProjectionBytes: int64(len(projectionData)),
			maximumModules:         1,
			maximumArtifactBytes:   int64(len("markdown")),
		}
	}); err != nil {
		t.Fatalf("aggregateWithOperations(exact limits) error = %v", err)
	}
}

func TestAggregateAccumulatesProjectionAndModuleBoundsAcrossRepositories(t *testing.T) {
	identity := Identity{
		DesignLanguageVersion: "1.0", DesignLanguageSHA256: strings.Repeat("a", 64),
		SourceIdentity: "v1.3.0", ToolingVersion: "v1.3.0", PublicationStatus: "published",
	}
	projectionData := [][]byte{[]byte("a"), []byte("bb"), []byte("ccc")}
	repositories := []string{"github.com/faustbrian/go-alpha", "github.com/faustbrian/go-beta", "github.com/faustbrian/go-gamma"}
	inputs := aggregateInputManifest{
		SchemaVersion:  1,
		DesignLanguage: aggregateDesignLanguage{Version: identity.DesignLanguageVersion, SHA256: identity.DesignLanguageSHA256},
	}
	projections := make([]aggregateEngineeringEnvelope, 0, len(repositories))
	for index, repository := range repositories {
		repositoryIdentity := repository
		digest := sha256.Sum256(projectionData[index])
		inputs.Repositories = append(inputs.Repositories, aggregateRepositoryInput{
			Repository: repository, Projection: filepath.Base(repository) + ".json", SHA256: hex.EncodeToString(digest[:]),
		})
		projections = append(projections, aggregateEngineeringEnvelope{
			SchemaVersion: 1, View: "engineering", Scope: "repository", Repository: &repositoryIdentity,
			DesignLanguage: DesignLanguageIdentity{
				Version: identity.DesignLanguageVersion, SHA256: identity.DesignLanguageSHA256, SourceIdentity: identity.SourceIdentity,
			},
			Tooling: ToolingIdentity{Version: identity.ToolingVersion, PublicationStatus: identity.PublicationStatus},
			Modules: []engineeringModule{engineeringProjection(repository, aggregateModuleForRepository(repository))},
		})
	}
	newOperations := func(limits aggregateLimits) aggregateOperations {
		readIndex := 0
		return aggregateOperations{
			limits: limits,
			read:   func(string, int64, string) ([]byte, error) { return []byte("inputs"), nil },
			readProjection: func(string, string, int64, string) ([]byte, error) {
				data := projectionData[readIndex]
				readIndex++
				return data, nil
			},
			validateInputs:     func([]byte) error { return nil },
			decodeInputs:       func([]byte, *aggregateInputManifest) error { return nil },
			validateCatalog:    func([]byte) error { return nil },
			decodeProjection:   func([]byte, *aggregateEngineeringEnvelope) error { return nil },
			validateProjection: validateAggregateProjection,
			marshal:            func(Envelope) ([]byte, error) { return []byte("json"), nil },
			render:             func(Envelope) ([]byte, error) { return []byte("markdown"), nil },
		}
	}
	run := func(limits aggregateLimits) error {
		operations := newOperations(limits)
		projectionIndex := 0
		operations.decodeInputs = func(_ []byte, target *aggregateInputManifest) error {
			*target = inputs
			return nil
		}
		operations.decodeProjection = func(_ []byte, target *aggregateEngineeringEnvelope) error {
			*target = projections[projectionIndex]
			projectionIndex++
			return nil
		}
		_, err := aggregateWithOperations("inputs.json", identity, operations)
		return err
	}
	exact := aggregateLimits{
		maximumRepositories: len(repositories), maximumProjectionBytes: int64(len(projectionData[0]) + len(projectionData[1]) + len(projectionData[2])),
		maximumModules: len(repositories), maximumArtifactBytes: int64(len("markdown")),
	}
	if err := run(exact); err != nil {
		t.Fatalf("aggregateWithOperations(exact accumulated bounds) error = %v", err)
	}
	projectionOverflow := exact
	projectionOverflow.maximumProjectionBytes--
	if err := run(projectionOverflow); err == nil || !strings.Contains(err.Error(), "maximum total size") {
		t.Fatalf("aggregateWithOperations(projection overflow) error = %v", err)
	}
	moduleOverflow := exact
	moduleOverflow.maximumModules--
	if err := run(moduleOverflow); err == nil || !strings.Contains(err.Error(), "maximum of 2 modules") {
		t.Fatalf("aggregateWithOperations(module overflow) error = %v", err)
	}
}

type aggregateEnvelopeObservation struct {
	View       string  `json:"view"`
	Scope      string  `json:"scope"`
	Repository *string `json:"repository"`
	Modules    []struct {
		ModulePath string `json:"module_path"`
	} `json:"modules"`
}

func assertAggregateModules(t *testing.T, data []byte, view string, want []string) {
	t.Helper()
	var observation aggregateEnvelopeObservation
	if err := json.Unmarshal(data, &observation); err != nil {
		t.Fatal(err)
	}
	if observation.View != view || observation.Scope != "ecosystem" || observation.Repository != nil {
		t.Fatalf("aggregate envelope = %#v", observation)
	}
	got := make([]string, 0, len(observation.Modules))
	for _, module := range observation.Modules {
		got = append(got, module.ModulePath)
	}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("aggregate modules = %v, want %v", got, want)
	}
}

func validAggregateModule() inventory.Module {
	module := validModule()
	module.TestTags = []string{}
	module.BuildTags = []string{}
	module.RequiredServices = []string{}
	module.ExternalRuntimeDependencies = []string{}
	module.InteroperabilityTools = []string{}
	module.ConformanceCorpora = []string{}
	module.Packages[0].BuildTags = []string{}
	return module
}

func aggregateModuleForRepository(repository string) inventory.Module {
	module := validAggregateModule()
	module.ModulePath = repository
	module.Packages[0].ImportPath = repository
	module.Cohesion.PrimaryEntryPackages = []string{repository}
	return module
}

func writeValidAggregateInputs(t *testing.T) (string, string, Identity) {
	t.Helper()
	root := t.TempDir()
	identity := Identity{
		DesignLanguageVersion: "1.0", DesignLanguageSHA256: strings.Repeat("a", 64),
		SourceIdentity: "v1.3.0", ToolingVersion: "v1.3.0", PublicationStatus: "published",
	}
	repository := "github.com/faustbrian/go-example"
	envelope, err := Project(inventory.Inventory{Repository: repository, Modules: []inventory.Module{aggregateModuleForRepository(repository)}}, "engineering", identity)
	if err != nil {
		t.Fatal(err)
	}
	projection, err := json.Marshal(envelope)
	if err != nil {
		t.Fatal(err)
	}
	projectionName := "repository.json"
	if err := os.WriteFile(filepath.Join(root, projectionName), projection, 0o600); err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(projection)
	inputs := aggregateInputManifest{
		SchemaVersion: 1,
		DesignLanguage: aggregateDesignLanguage{
			Version: identity.DesignLanguageVersion,
			SHA256:  identity.DesignLanguageSHA256,
		},
		Repositories: []aggregateRepositoryInput{{
			Repository: repository, Projection: projectionName, SHA256: hex.EncodeToString(digest[:]),
		}},
	}
	manifest, err := json.Marshal(inputs)
	if err != nil {
		t.Fatal(err)
	}
	inputsPath := filepath.Join(root, "inputs.json")
	if err := os.WriteFile(inputsPath, manifest, 0o600); err != nil {
		t.Fatal(err)
	}
	return root, inputsPath, identity
}

type fakeCatalogFiles struct {
	file         *fakeCatalogFile
	files        []*fakeCatalogFile
	mkdirErr     error
	createErr    error
	createErrors map[int]error
	lstatInfos   map[int]os.FileInfo
	lstatErrors  map[int]error
	renameErrors map[int]error
	removeErrors map[int]error
	createCalls  int
	lstatCalls   int
	renameCalls  int
	removeCalls  int
	removed      []string
}

func (files *fakeCatalogFiles) MkdirAll(string, os.FileMode) error { return files.mkdirErr }
func (files *fakeCatalogFiles) Lstat(string) (os.FileInfo, error) {
	files.lstatCalls++
	if err := files.lstatErrors[files.lstatCalls]; err != nil {
		return nil, err
	}
	if info := files.lstatInfos[files.lstatCalls]; info != nil {
		return info, nil
	}
	return nil, os.ErrNotExist
}
func (files *fakeCatalogFiles) CreateTemp(string, string) (catalogFile, error) {
	files.createCalls++
	if err := files.createErrors[files.createCalls]; err != nil {
		return nil, err
	}
	if files.createErr != nil {
		return nil, files.createErr
	}
	if files.createCalls <= len(files.files) {
		return files.files[files.createCalls-1], nil
	}
	if files.file == nil {
		files.file = &fakeCatalogFile{}
	}
	return files.file, nil
}
func (files *fakeCatalogFiles) Rename(string, string) error {
	files.renameCalls++
	return files.renameErrors[files.renameCalls]
}
func (files *fakeCatalogFiles) Remove(path string) error {
	files.removeCalls++
	files.removed = append(files.removed, path)
	return files.removeErrors[files.removeCalls]
}

type fakeCatalogFile struct {
	name     string
	writeN   int
	chmodErr error
	writeErr error
	syncErr  error
	closeErr error
}

func (file *fakeCatalogFile) Name() string {
	if file.name != "" {
		return file.name
	}
	return "/temporary-catalog"
}
func (file *fakeCatalogFile) Chmod(os.FileMode) error { return file.chmodErr }
func (file *fakeCatalogFile) Write(value []byte) (int, error) {
	if file.writeN != 0 {
		return file.writeN, file.writeErr
	}
	return len(value), file.writeErr
}
func (file *fakeCatalogFile) Sync() error  { return file.syncErr }
func (file *fakeCatalogFile) Close() error { return file.closeErr }

type fakeCatalogInfo struct {
	mode os.FileMode
	size int64
}

func (fakeCatalogInfo) Name() string { return "catalog" }
func (info fakeCatalogInfo) Size() int64 {
	if info.size != 0 {
		return info.size
	}
	return 1
}
func (info fakeCatalogInfo) Mode() os.FileMode { return info.mode }
func (fakeCatalogInfo) ModTime() time.Time     { return time.Time{} }
func (info fakeCatalogInfo) IsDir() bool       { return info.mode.IsDir() }
func (fakeCatalogInfo) Sys() any               { return nil }

type fakeAggregateInputFiles struct {
	info      os.FileInfo
	lstatErr  error
	file      aggregateReadCloser
	openErr   error
	different bool
}

func (files *fakeAggregateInputFiles) Lstat(string) (os.FileInfo, error) {
	return files.info, files.lstatErr
}
func (files *fakeAggregateInputFiles) Open(string) (aggregateReadCloser, error) {
	return files.file, files.openErr
}
func (files *fakeAggregateInputFiles) SameFile(os.FileInfo, os.FileInfo) bool {
	return !files.different
}

type fakeAggregateReader struct {
	reader  io.Reader
	err     error
	info    os.FileInfo
	statErr error
}

func (reader *fakeAggregateReader) Read(value []byte) (int, error) {
	if reader.err != nil {
		return 0, reader.err
	}
	return reader.reader.Read(value)
}
func (*fakeAggregateReader) Close() error { return nil }
func (reader *fakeAggregateReader) Stat() (os.FileInfo, error) {
	return reader.info, reader.statErr
}
