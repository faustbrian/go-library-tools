// Package cohesion validates repository metadata and projects deterministic
// consumer/engineering catalogs. Start with LoadAndCheck or Project, then use
// explicit aggregation for locked inputs. Validation is not runtime or release
// qualification; projections may borrow nested catalog data. Publication owns
// confined staged files, not a crash-atomic transaction across all artifacts.
package cohesion

import (
	"cmp"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/faustbrian/go-library-tools/v2/internal/config"
	"github.com/faustbrian/go-library-tools/v2/internal/inventory"
)

// Report is the stable machine-readable cohesion-check result.
type Report struct {
	// SchemaVersion is 1 for the stable diagnostic shape.
	SchemaVersion int `json:"schema_version"`
	// Repository is nil when identity could not be established.
	Repository *string `json:"repository"`
	// ManifestSchemaVersion is nil when no source schema version could be
	// established.
	ManifestSchemaVersion *int `json:"manifest_schema_version"`
	// Valid is true only when no cohesion diagnostics remain, not when downstream
	// gates pass.
	Valid bool `json:"valid"`
	// Summary contains established counts, with nil preserving unavailable facts.
	Summary Summary `json:"summary"`
	// Diagnostics is sorted by path, code and message; parse failures remain
	// visible.
	Diagnostics []Diagnostic `json:"diagnostics"`
}

// Summary counts the modules that could be established from the manifest.
type Summary struct {
	// TotalModules is nil for unavailable inventory, otherwise the decoded count.
	TotalModules *int `json:"total_modules"`
	// ReleasableModules counts declared release units, not published releases; nil
	// means unavailable.
	ReleasableModules *int `json:"releasable_modules"`
	// ClassifiedModules counts modules with cohesion metadata; nil means
	// unavailable.
	ClassifiedModules *int `json:"classified_modules"`
	// ErrorCount counts reported cohesion violations, independently of partial
	// summary counts.
	ErrorCount int `json:"error_count"`
}

// Diagnostic identifies one deterministic cohesion violation.
type Diagnostic struct {
	// Code is the stable machine classification of the violation.
	Code string `json:"code"`
	// Path is the JSON-style manifest location, empty for document-wide failure.
	Path string `json:"path"`
	// Message is bounded maintained diagnostic prose rather than untrusted source
	// payload.
	Message string `json:"message"`
}

// Check validates adoption and required metadata without inferring repository siblings.
func Check(root string, policy config.Config) Report {
	_, report := LoadAndCheck(root, policy)
	return report
}

// LoadAndCheck binds diagnostics to one admitted module-manifest read and the
// corresponding decoded inventory. Report.Valid must be true before catalog
// projection; an invalid report may retain partial inventory/counts solely for
// diagnostics. Neither success nor one read is an atomic two-file snapshot.
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
		if catalog.SchemaVersion >= 2 && module.Releasable && module.Cohesion == nil {
			diagnostics = append(diagnostics, Diagnostic{
				Code:    "missing-metadata",
				Path:    fmt.Sprintf("/modules/%d/cohesion", index),
				Message: "releasable module requires cohesion metadata",
			})
		}
		if catalog.SchemaVersion >= 2 && module.Cohesion != nil {
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
