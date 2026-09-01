// Package cohesion validates and projects the ecosystem cohesion contract.
package cohesion

import (
	"cmp"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/faustbrian/go-library-tools/internal/config"
	"github.com/faustbrian/go-library-tools/internal/inventory"
)

// Report is the stable machine-readable cohesion-check result.
type Report struct {
	SchemaVersion         int          `json:"schema_version"`
	Repository            *string      `json:"repository"`
	ManifestSchemaVersion *int         `json:"manifest_schema_version"`
	Valid                 bool         `json:"valid"`
	Summary               Summary      `json:"summary"`
	Diagnostics           []Diagnostic `json:"diagnostics"`
}

// Summary counts the modules that could be established from the manifest.
type Summary struct {
	TotalModules      *int `json:"total_modules"`
	ReleasableModules *int `json:"releasable_modules"`
	ClassifiedModules *int `json:"classified_modules"`
	ErrorCount        int  `json:"error_count"`
}

// Diagnostic identifies one deterministic cohesion violation.
type Diagnostic struct {
	Code    string `json:"code"`
	Path    string `json:"path"`
	Message string `json:"message"`
}

// Check validates adoption and required metadata without inferring repository siblings.
func Check(root string, policy config.Config) Report {
	_, report := LoadAndCheck(root, policy)
	return report
}

// LoadAndCheck returns the single validated inventory snapshot used by catalog projection.
func LoadAndCheck(root string, policy config.Config) (inventory.Inventory, Report) {
	catalog, moduleManifest, err := inventory.LoadSnapshot(root, policy)
	if err != nil && len(moduleManifest) == 0 {
		return inventory.Inventory{}, invalidManifest()
	}

	total := len(catalog.Modules)
	releasable := 0
	classified := 0
	diagnostics := make([]Diagnostic, 0)
	schemaInvalid := false
	if err != nil {
		diagnostics = append(diagnostics, Diagnostic{
			Code: "invalid-manifest", Path: "", Message: "cannot load canonical manifests",
		})
	}
	if catalog.SchemaVersion == 2 {
		if err := validateModulesSchema(moduleManifest); err != nil {
			schemaInvalid = true
		}
	}
	for index, module := range catalog.Modules {
		if module.Releasable {
			releasable++
		}
		if module.Cohesion != nil {
			classified++
		}
		if catalog.SchemaVersion == 2 && module.Releasable && module.Cohesion == nil {
			diagnostics = append(diagnostics, Diagnostic{
				Code:    "missing-metadata",
				Path:    fmt.Sprintf("/modules/%d/cohesion", index),
				Message: "releasable module requires cohesion metadata",
			})
		}
		if catalog.SchemaVersion == 2 && module.Cohesion != nil {
			diagnostics = append(diagnostics, validateModule(root, index, module)...)
		}
	}
	if catalog.SchemaVersion == 1 {
		diagnostics = append(diagnostics, Diagnostic{
			Code:    "adoption-required",
			Path:    "/schema_version",
			Message: "cohesion validation requires modules.json schema_version 2",
		})
	}
	diagnostics = append(diagnostics, requiredMetadataDiagnosticsData(moduleManifest)...)
	if schemaInvalid {
		diagnostics = append(diagnostics, Diagnostic{
			Code: "invalid-manifest", Path: "", Message: "modules manifest violates the normative schema",
		})
	}
	sortDiagnostics(diagnostics)
	var repository *string
	if catalog.Repository != "" {
		repository = &catalog.Repository
	}
	var manifestHeader struct {
		SchemaVersion *int `json:"schema_version"`
	}
	_ = json.Unmarshal(moduleManifest, &manifestHeader)
	return catalog, Report{
		SchemaVersion:         1,
		Repository:            repository,
		ManifestSchemaVersion: manifestHeader.SchemaVersion,
		Valid:                 len(diagnostics) == 0,
		Summary: Summary{
			TotalModules:      &total,
			ReleasableModules: &releasable,
			ClassifiedModules: &classified,
			ErrorCount:        len(diagnostics),
		},
		Diagnostics: diagnostics,
	}
}

func invalidManifest() Report {
	return Report{
		SchemaVersion: 1,
		Valid:         false,
		Summary:       Summary{ErrorCount: 1},
		Diagnostics: []Diagnostic{{
			Code:    "invalid-manifest",
			Path:    "",
			Message: "cannot load canonical manifests",
		}},
	}
}

func sortDiagnostics(diagnostics []Diagnostic) {
	slices.SortFunc(diagnostics, func(left, right Diagnostic) int {
		return cmp.Or(
			strings.Compare(left.Path, right.Path),
			strings.Compare(left.Code, right.Code),
			strings.Compare(left.Message, right.Message),
		)
	})
}
