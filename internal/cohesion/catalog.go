package cohesion

import (
	"bytes"
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/faustbrian/go-library-tools/internal/inventory"
)

// Identity binds catalog output to the design language and generator.
type Identity struct {
	DesignLanguageVersion string
	DesignLanguageSHA256  string
	SourceIdentity        string
	ToolingVersion        string
	PublicationStatus     string
}

// RenderMarkdown renders an envelope without adding timestamps or local state.
func RenderMarkdown(envelope Envelope) ([]byte, error) {
	var output bytes.Buffer
	title := "Consumer"
	if envelope.View == "engineering" {
		title = "Engineering"
	}
	_, _ = fmt.Fprintf(&output, "# Golib %s Catalog\n\n", title)
	_, _ = fmt.Fprintf(&output, "Design language `%s` (`%s`); tooling `%s`.\n", envelope.DesignLanguage.Version, envelope.DesignLanguage.SourceIdentity, envelope.Tooling.Version)
	lastFamily := ""
	switch modules := envelope.Modules.(type) {
	case []consumerModule:
		for _, module := range modules {
			if module.Cohesion.Family != lastFamily {
				lastFamily = module.Cohesion.Family
				_, _ = fmt.Fprintf(&output, "\n## %s\n", lastFamily)
			}
			_, _ = fmt.Fprintf(&output, "\n- `%s`: %s\n", module.ModulePath, module.Cohesion.Responsibility)
		}
	case []engineeringModule:
		for _, module := range modules {
			family := "unclassified-internal"
			responsibility := module.Purpose
			if module.Cohesion != nil {
				family = module.Cohesion.Family
				responsibility = module.Cohesion.Responsibility
			}
			if family != lastFamily {
				lastFamily = family
				_, _ = fmt.Fprintf(&output, "\n## %s\n", lastFamily)
			}
			_, _ = fmt.Fprintf(&output, "\n- `%s` (%s): %s\n", module.ModulePath, module.Kind, responsibility)
		}
	default:
		return nil, errors.New("catalog modules do not match the envelope view")
	}
	return output.Bytes(), nil
}

// Envelope is a deterministic repository or ecosystem catalog projection.
type Envelope struct {
	SchemaVersion  int                    `json:"schema_version"`
	View           string                 `json:"view"`
	Scope          string                 `json:"scope"`
	Repository     *string                `json:"repository"`
	DesignLanguage DesignLanguageIdentity `json:"design_language"`
	Tooling        ToolingIdentity        `json:"tooling"`
	Modules        any                    `json:"modules"`
}

// DesignLanguageIdentity binds the catalog to reviewed consumer guidance.
type DesignLanguageIdentity struct {
	Version        string `json:"version"`
	SHA256         string `json:"sha256"`
	SourceIdentity string `json:"source_identity"`
}

// ToolingIdentity records the exact generator identity and publication state.
type ToolingIdentity struct {
	Version           string `json:"version"`
	PublicationStatus string `json:"publication_status"`
}

type consumerModule struct {
	Repository        string              `json:"repository"`
	Directory         string              `json:"directory"`
	ModulePath        string              `json:"module_path"`
	GoVersion         string              `json:"go_version"`
	Kind              string              `json:"kind"`
	Releasable        bool                `json:"releasable"`
	Version           string              `json:"version"`
	Specifications    []string            `json:"specifications"`
	OwnedDependencies []string            `json:"owned_dependencies"`
	Cohesion          *inventory.Cohesion `json:"cohesion"`
}

type engineeringModule struct {
	Repository                  string                   `json:"repository"`
	Directory                   string                   `json:"directory"`
	ModulePath                  string                   `json:"module_path"`
	GoVersion                   string                   `json:"go_version"`
	Kind                        string                   `json:"kind"`
	Purpose                     string                   `json:"purpose"`
	Lifecycle                   string                   `json:"lifecycle"`
	Releasable                  bool                     `json:"releasable"`
	Version                     string                   `json:"version"`
	TagPrefix                   string                   `json:"tag_prefix"`
	Gates                       map[string]bool          `json:"gates"`
	TestTags                    []string                 `json:"test_tags"`
	BuildTags                   []string                 `json:"build_tags"`
	RequiredServices            []string                 `json:"required_services"`
	ExternalRuntimeDependencies []string                 `json:"external_runtime_dependencies"`
	InteroperabilityTools       []string                 `json:"interoperability_tools"`
	ConformanceCorpora          []string                 `json:"conformance_corpora"`
	Specifications              []string                 `json:"specifications"`
	OwnedDependencies           []string                 `json:"owned_dependencies"`
	ReverseOwnedDependencies    []string                 `json:"reverse_owned_dependencies"`
	Packages                    []inventory.Package      `json:"packages"`
	Family                      string                   `json:"family"`
	FamilyLabel                 string                   `json:"family_label"`
	FamilyDescription           string                   `json:"family_description"`
	FamilyOrder                 int                      `json:"family_order"`
	GoalStatus                  string                   `json:"goal_status"`
	GoalFiles                   []string                 `json:"goal_files"`
	GoalEvidence                []inventory.GoalEvidence `json:"goal_evidence"`
	Provenance                  json.RawMessage          `json:"provenance"`
	Cohesion                    *inventory.Cohesion      `json:"cohesion"`
}

var familyOrder = map[string]int{
	"foundations":                   0,
	"service-edge":                  1,
	"protocols-and-descriptions":    2,
	"persistence-and-durability":    3,
	"resilience":                    4,
	"observability":                 5,
	"integration-and-data-movement": 6,
	"domain-utilities":              7,
	"tooling":                       8,
}

// Project creates a repository-scoped consumer or engineering catalog.
func Project(catalog inventory.Inventory, view string, identity Identity) (Envelope, error) {
	if view != "consumer" && view != "engineering" {
		return Envelope{}, errors.New("catalog view must be consumer or engineering")
	}
	if err := validateIdentity(identity); err != nil {
		return Envelope{}, err
	}
	modules := append([]inventory.Module(nil), catalog.Modules...)
	slices.SortFunc(modules, func(left, right inventory.Module) int {
		leftOrder := moduleFamilyOrder(left)
		rightOrder := moduleFamilyOrder(right)
		if leftOrder != rightOrder {
			return cmp.Compare(leftOrder, rightOrder)
		}
		return strings.Compare(left.ModulePath, right.ModulePath)
	})

	repository := catalog.Repository
	envelope := Envelope{
		SchemaVersion: 1,
		View:          view,
		Scope:         "repository",
		Repository:    &repository,
		DesignLanguage: DesignLanguageIdentity{
			Version:        identity.DesignLanguageVersion,
			SHA256:         identity.DesignLanguageSHA256,
			SourceIdentity: identity.SourceIdentity,
		},
		Tooling: ToolingIdentity{Version: identity.ToolingVersion, PublicationStatus: identity.PublicationStatus},
	}
	if view == "consumer" {
		projected := make([]consumerModule, 0, len(modules))
		for _, module := range modules {
			if !consumerVisible(module) {
				continue
			}
			projected = append(projected, consumerModule{
				Repository: catalog.Repository, Directory: module.Directory,
				ModulePath: module.ModulePath, GoVersion: module.GoVersion, Kind: module.Kind,
				Releasable: module.Releasable, Version: module.Version,
				Specifications: module.Specifications, OwnedDependencies: module.OwnedDependencies,
				Cohesion: module.Cohesion,
			})
		}
		envelope.Modules = projected
		return envelope, nil
	}

	projected := make([]engineeringModule, 0, len(modules))
	for _, module := range modules {
		projected = append(projected, engineeringProjection(catalog.Repository, module))
	}
	envelope.Modules = projected
	return envelope, nil
}

func consumerVisible(module inventory.Module) bool {
	if !module.Releasable || module.Cohesion == nil {
		return false
	}
	if module.Kind != "public library" && module.Kind != "adapter" {
		return false
	}
	return module.Cohesion.LifecycleStatus == "active" || module.Cohesion.LifecycleStatus == "deprecated"
}

func moduleFamilyOrder(module inventory.Module) int {
	if module.Cohesion == nil {
		return len(familyOrder)
	}
	if order, exists := familyOrder[module.Cohesion.Family]; exists {
		return order
	}
	return len(familyOrder)
}

func validateIdentity(identity Identity) error {
	if identity.DesignLanguageVersion != "1.0" || len(identity.DesignLanguageSHA256) != 64 || strings.Trim(identity.DesignLanguageSHA256, "0123456789abcdef") != "" {
		return errors.New("design-language identity is invalid")
	}
	if identity.PublicationStatus == "unpublished" && identity.SourceIdentity == "unpublished" && identity.ToolingVersion == "dev" {
		return nil
	}
	if identity.PublicationStatus == "published" && semverTag(identity.SourceIdentity) && semverTag(identity.ToolingVersion) {
		return nil
	}
	return errors.New("tooling publication identity is invalid")
}

func semverTag(value string) bool {
	if len(value) < 6 || value[0] != 'v' {
		return false
	}
	parts := strings.Split(value[1:], ".")
	if len(parts) != 3 {
		return false
	}
	for _, part := range parts {
		if part == "" || strings.Trim(part, "0123456789") != "" || (len(part) > 1 && part[0] == '0') {
			return false
		}
	}
	return true
}

func engineeringProjection(repository string, module inventory.Module) engineeringModule {
	return engineeringModule{
		Repository: repository, Directory: module.Directory, ModulePath: module.ModulePath,
		GoVersion: module.GoVersion, Kind: module.Kind, Purpose: module.Purpose,
		Lifecycle: module.Lifecycle, Releasable: module.Releasable, Version: module.Version,
		TagPrefix: module.TagPrefix, Gates: module.Gates, TestTags: module.TestTags,
		BuildTags: module.BuildTags, RequiredServices: module.RequiredServices,
		ExternalRuntimeDependencies: module.ExternalRuntimeDependencies,
		InteroperabilityTools:       module.InteroperabilityTools, ConformanceCorpora: module.ConformanceCorpora,
		Specifications: module.Specifications, OwnedDependencies: module.OwnedDependencies,
		ReverseOwnedDependencies: module.ReverseOwnedDependencies, Packages: module.Packages,
		Family: module.Family, FamilyLabel: module.FamilyLabel,
		FamilyDescription: module.FamilyDescription, FamilyOrder: module.FamilyOrder,
		GoalStatus: module.GoalStatus, GoalFiles: module.GoalFiles, GoalEvidence: module.GoalEvidence,
		Provenance: module.Provenance, Cohesion: module.Cohesion,
	}
}
