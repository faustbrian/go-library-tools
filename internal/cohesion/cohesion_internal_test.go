package cohesion

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/faustbrian/go-library-tools/internal/config"
	"github.com/faustbrian/go-library-tools/internal/inventory"
)

func TestValidateModuleCoversContractBoundaries(t *testing.T) {
	mutations := []struct {
		name   string
		mutate func(*inventory.Module, string)
	}{
		{"family", func(module *inventory.Module, _ string) { module.Cohesion.Family = "unknown" }},
		{"capability", func(module *inventory.Module, _ string) { module.Cohesion.SecondaryCapabilities = []string{"unknown"} }},
		{"non-goals missing", func(module *inventory.Module, _ string) { module.Cohesion.NonGoals = nil }},
		{"non-goals blank", func(module *inventory.Module, _ string) { module.Cohesion.NonGoals = []string{" "} }},
		{"identifier", func(module *inventory.Module, _ string) { module.Cohesion.PublicPackageIdentifier = "bad-name" }},
		{"primary missing", func(module *inventory.Module, _ string) { module.Cohesion.PrimaryEntryPackages = nil }},
		{"primary unknown", func(module *inventory.Module, _ string) {
			module.Cohesion.PrimaryEntryPackages = []string{"example.com/missing"}
		}},
		{"rootless selection", func(module *inventory.Module, _ string) { module.Cohesion.PublicPackageIdentifier = "none" }},
		{"selection unknown", func(module *inventory.Module, _ string) {
			module.Cohesion.PackageSelection = map[string]string{"example.com/missing": "Use it."}
		}},
		{"selection blank", func(module *inventory.Module, _ string) {
			module.Cohesion.PackageSelection = map[string]string{module.ModulePath: " "}
		}},
		{"lifecycle", func(module *inventory.Module, _ string) { module.Cohesion.LifecycleStatus = "unknown" }},
		{"planned stable", func(module *inventory.Module, _ string) { module.Cohesion.LifecycleStatus = "planned" }},
		{"construction empty", func(module *inventory.Module, _ string) { module.Cohesion.ConstructionStyles = nil }},
		{"construction unknown", func(module *inventory.Module, _ string) { module.Cohesion.ConstructionStyles = []string{"unknown"} }},
		{"lifecycle style", func(module *inventory.Module, _ string) { module.Cohesion.LifecycleStyles = []string{"unknown"} }},
		{"configuration owner", func(module *inventory.Module, _ string) { module.Cohesion.Ownership.Configuration = "package" }},
		{"mutable blank", func(module *inventory.Module, _ string) { module.Cohesion.Ownership.MutableInputs = []string{""} }},
		{"mutable none conflict", func(module *inventory.Module, _ string) {
			module.Cohesion.Ownership.MutableInputs = []string{"copy", "none"}
		}},
		{"resource owner", func(module *inventory.Module, _ string) { module.Cohesion.Ownership.RuntimeResources = "unknown" }},
		{"work owner", func(module *inventory.Module, _ string) { module.Cohesion.Ownership.BackgroundWork = "unknown" }},
		{"optional order", func(module *inventory.Module, _ string) {
			module.Cohesion.OptionalOwnedDependencies = []string{"z", "a"}
		}},
		{"adapter duplicate", func(module *inventory.Module, _ string) { module.Cohesion.Adapters = []string{"a", "a"} }},
		{"dependency overlap", func(module *inventory.Module, _ string) {
			module.Cohesion.Companions = []string{"example.com/required"}
		}},
		{"minimum format", func(module *inventory.Module, _ string) { module.Cohesion.SupportedGo.Minimum = "1.27" }},
		{"tested empty", func(module *inventory.Module, _ string) { module.Cohesion.SupportedGo.Tested = nil }},
		{"tested below", func(module *inventory.Module, _ string) { module.Cohesion.SupportedGo.Tested = []string{"1.26.0"} }},
		{"platform empty", func(module *inventory.Module, _ string) { module.Cohesion.SupportedPlatforms = nil }},
		{"platform invalid", func(module *inventory.Module, _ string) { module.Cohesion.SupportedPlatforms = []string{"linux"} }},
		{"backend invalid", func(module *inventory.Module, _ string) { module.Cohesion.SupportedBackends = []string{"Postgres"} }},
		{"protocol invalid", func(module *inventory.Module, _ string) { module.Cohesion.SupportedProtocols = []string{"HTTP/1"} }},
		{"required document", func(module *inventory.Module, _ string) { module.Cohesion.Documentation.README = nil }},
		{"missing document", func(module *inventory.Module, _ string) { module.Cohesion.Documentation.README = new("missing.md") }},
		{"absolute document", func(module *inventory.Module, root string) {
			module.Cohesion.Documentation.README = new(filepath.Join(root, "README.md"))
		}},
		{"backslash document", func(module *inventory.Module, _ string) {
			module.Cohesion.Documentation.README = new(`docs\README.md`)
		}},
		{"unclean document", func(module *inventory.Module, _ string) {
			module.Cohesion.Documentation.README = new("docs/../README.md")
		}},
		{"symlink document", func(module *inventory.Module, root string) {
			if err := os.Symlink(filepath.Join(root, "README.md"), filepath.Join(root, "link.md")); err != nil {
				t.Fatal(err)
			}
			module.Cohesion.Documentation.README = new("link.md")
		}},
		{"symlinked parent document", func(module *inventory.Module, root string) {
			outside := t.TempDir()
			writeDocument(t, filepath.Join(outside, "README.md"))
			if err := os.Symlink(outside, filepath.Join(root, "docs")); err != nil {
				t.Fatal(err)
			}
			module.Cohesion.Documentation.README = new("docs/README.md")
		}},
		{"compatibility order", func(module *inventory.Module, _ string) {
			module.Cohesion.KnownGoodCompatibilitySets = []string{"z", "a"}
		}},
		{"delivery invalid", func(module *inventory.Module, _ string) { module.Cohesion.Delivery.Release = "unknown" }},
		{"verified without evidence", func(module *inventory.Module, _ string) { module.Cohesion.Delivery.Implementation = "verified" }},
	}
	for _, mutation := range mutations {
		t.Run(mutation.name, func(t *testing.T) {
			root := t.TempDir()
			writeDocument(t, filepath.Join(root, "README.md"))
			writeDocument(t, filepath.Join(root, "CHANGELOG.md"))
			module := validModule()
			mutation.mutate(&module, root)
			if diagnostics := validateModule(root, 0, module); len(diagnostics) == 0 {
				t.Fatalf("validateModule(%s) returned no diagnostics", mutation.name)
			}
		})
	}
}

func TestValidateModuleAcceptsCompleteMetadata(t *testing.T) {
	root := t.TempDir()
	writeDocument(t, filepath.Join(root, "README.md"))
	writeDocument(t, filepath.Join(root, "CHANGELOG.md"))
	module := validModule()
	module.GoalStatus = "complete"
	module.GoalFiles = []string{".ai/GOAL.md"}
	module.GoalEvidence = []inventory.GoalEvidence{{
		File: ".ai/GOAL.md", RequirementsSHA256: strings.Repeat("a", 64),
		ImplementationEvidence: []string{"README.md"}, VerificationGates: []string{"test"},
		ImplementationStatus: "verified",
	}}
	module.Cohesion.Delivery = inventory.Delivery{Implementation: "verified", Hardening: "verified", Release: "verified"}
	if diagnostics := validateModule(root, 0, module); len(diagnostics) != 0 {
		t.Fatalf("validateModule(valid) = %#v", diagnostics)
	}
}

func TestValidateModuleRequiresAttributableVerifiedDeliveryEvidence(t *testing.T) {
	root := t.TempDir()
	writeDocument(t, filepath.Join(root, "README.md"))
	writeDocument(t, filepath.Join(root, "CHANGELOG.md"))
	complete := inventory.GoalEvidence{
		File: ".ai/GOAL.md", RequirementsSHA256: strings.Repeat("a", 64),
		ImplementationEvidence: []string{"README.md"}, VerificationGates: []string{"test"},
		ImplementationStatus: "verified",
	}
	for _, test := range []struct {
		name       string
		goalStatus string
		goalFiles  []string
		evidence   inventory.GoalEvidence
	}{
		{"incomplete goal", "in-progress", []string{".ai/GOAL.md"}, complete},
		{"unrelated goal file", "complete", []string{".ai/OTHER.md"}, complete},
		{"missing implementation evidence", "complete", []string{".ai/GOAL.md"}, func() inventory.GoalEvidence { value := complete; value.ImplementationEvidence = nil; return value }()},
		{"missing verification gates", "complete", []string{".ai/GOAL.md"}, func() inventory.GoalEvidence { value := complete; value.VerificationGates = nil; return value }()},
		{"stale verification status", "complete", []string{".ai/GOAL.md"}, func() inventory.GoalEvidence {
			value := complete
			value.ImplementationStatus = "implemented-requires-fresh-verification"
			return value
		}()},
	} {
		t.Run(test.name, func(t *testing.T) {
			module := validModule()
			module.GoalStatus = test.goalStatus
			module.GoalFiles = test.goalFiles
			module.GoalEvidence = []inventory.GoalEvidence{test.evidence}
			module.Cohesion.Delivery.Implementation = "verified"
			assertDiagnosticSet(t, validateModule(root, 0, module), []string{
				"invalid-reference:/modules/0/cohesion/delivery/implementation",
			})
		})
	}
}

func TestValidationHelpersCoverOrderingVersionsAndSelection(t *testing.T) {
	for _, test := range []struct {
		left, right string
		want        int
	}{{"1.27.0", "1.27.0", 0}, {"1.28.0", "1.27.0", 1}, {"1.26.0", "1.27.0", -1}, {"bad", "1.27.0", -1}, {"x.27.0", "1.27.0", -1}, {"1.27.0", "x.27.0", -1}} {
		if got := compareGoVersions(test.left, test.right); got != test.want {
			t.Fatalf("compareGoVersions(%q, %q) = %d", test.left, test.right, got)
		}
	}
	if selectionCovers(map[string]string{"a": "Use A."}, []string{"a"}) != true || selectionCovers(map[string]string{}, []string{"a"}) {
		t.Fatal("selectionCovers returned an invalid result")
	}
	if !contains([]string{"a"}, "a") || contains([]string{"a"}, "b") || !containsBlank([]string{" "}) || containsBlank([]string{"a"}) {
		t.Fatal("list helper returned an invalid result")
	}
	for value, want := range map[string]bool{
		"https://example.com/path": true,
		"http://example.com/path":  false,
		"https:///path":            false,
		"https://user@example.com": false,
		"%":                        false,
	} {
		if got := isHTTPS(value); got != want {
			t.Fatalf("isHTTPS(%q) = %t", value, got)
		}
	}
}

func TestValidateModulePreservesExactBoundaryDiagnostics(t *testing.T) {
	root := t.TempDir()
	writeDocument(t, filepath.Join(root, "README.md"))
	writeDocument(t, filepath.Join(root, "CHANGELOG.md"))
	tests := []struct {
		name   string
		mutate func(*inventory.Module)
		code   string
		path   string
		count  int
	}{
		{"malformed identifier", func(module *inventory.Module) { module.Cohesion.PublicPackageIdentifier = "bad-name" }, "invalid-value", "/modules/0/cohesion/public_package_identifier", 2},
		{"two invalid selections fail fast", func(module *inventory.Module) {
			module.Cohesion.PackageSelection = map[string]string{"example.com/missing-a": "Use A.", "example.com/missing-b": "Use B."}
		}, "invalid-reference", "/modules/0/cohesion/package_selection", 1},
		{"two invalid tested versions fail fast", func(module *inventory.Module) { module.Cohesion.SupportedGo.Tested = []string{"1.25.0", "1.26.0"} }, "invalid-value", "/modules/0/cohesion/supported_go/tested", 1},
		{"backslash remains an unsafe path", func(module *inventory.Module) { module.Cohesion.Documentation.README = new(`docs\README.md`) }, "unsafe-path", "/modules/0/cohesion/documentation/readme", 1},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			module := validModule()
			test.mutate(&module)
			diagnostics := validateModule(root, 0, module)
			if len(diagnostics) != test.count || !diagnosticExists(diagnostics, test.code, test.path) {
				t.Fatalf("validateModule() = %#v", diagnostics)
			}
		})
	}
}

func TestValidateModuleAcceptsEveryValidBoundaryAlternative(t *testing.T) {
	root := t.TempDir()
	writeDocument(t, filepath.Join(root, "README.md"))
	writeDocument(t, filepath.Join(root, "CHANGELOG.md"))
	tests := []struct {
		name   string
		mutate func(*inventory.Module)
	}{
		{"rootless selection", func(module *inventory.Module) {
			module.Packages = []inventory.Package{{
				ModuleDirectory: ".", Directory: "cmd/tool", Name: "main", ImportPath: module.ModulePath + "/cmd/tool", Kind: "command",
			}}
			module.Cohesion.PublicPackageIdentifier = "none"
			module.Cohesion.PrimaryEntryPackages = []string{module.ModulePath + "/cmd/tool"}
			module.Cohesion.PackageSelection = map[string]string{module.ModulePath + "/cmd/tool": "Install the command."}
		}},
		{"root command is not a public package", func(module *inventory.Module) {
			module.Packages = []inventory.Package{{
				ModuleDirectory: ".", Directory: ".", Name: "main", ImportPath: module.ModulePath, Kind: "command",
			}}
			module.Cohesion.PublicPackageIdentifier = "none"
			module.Cohesion.PackageSelection = map[string]string{module.ModulePath: "Install the command."}
		}},
		{"none mutable inputs", func(module *inventory.Module) { module.Cohesion.Ownership.MutableInputs = []string{"none"} }},
		{"multiple sorted mutable inputs", func(module *inventory.Module) { module.Cohesion.Ownership.MutableInputs = []string{"borrow", "copy"} }},
		{"one-hyphen platform", func(module *inventory.Module) { module.Cohesion.SupportedPlatforms = []string{"linux-amd64"} }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			module := validModule()
			test.mutate(&module)
			if diagnostics := validateModule(root, 0, module); len(diagnostics) != 0 {
				t.Fatalf("validateModule() = %#v", diagnostics)
			}
		})
	}
}

func TestValidateModuleRejectsRootlessIdentifierWithRootPublicPackage(t *testing.T) {
	root := t.TempDir()
	writeDocument(t, filepath.Join(root, "README.md"))
	writeDocument(t, filepath.Join(root, "CHANGELOG.md"))
	module := validModule()
	module.Cohesion.PublicPackageIdentifier = "none"
	module.Cohesion.PackageSelection = map[string]string{module.ModulePath: "Use the primary package."}

	assertDiagnosticSet(t, validateModule(root, 0, module), []string{
		"invalid-reference:/modules/0/cohesion/public_package_identifier",
	})
}

func TestValidateModuleRequiresSelectionForRootlessPrimaryEntries(t *testing.T) {
	root := t.TempDir()
	writeDocument(t, filepath.Join(root, "README.md"))
	writeDocument(t, filepath.Join(root, "CHANGELOG.md"))
	module := validModule()
	module.Packages = []inventory.Package{{
		ModuleDirectory: ".", Directory: "cmd/tool", Name: "main", ImportPath: module.ModulePath + "/cmd/tool", Kind: "command",
	}}
	module.Cohesion.PublicPackageIdentifier = "none"
	module.Cohesion.PrimaryEntryPackages = []string{module.ModulePath + "/cmd/tool"}
	module.Cohesion.PackageSelection = map[string]string{}

	assertDiagnosticSet(t, validateModule(root, 0, module), []string{
		"invalid-value:/modules/0/cohesion/package_selection",
	})
}

func TestValidateDocumentationTraversesEveryEntry(t *testing.T) {
	root := t.TempDir()
	writeDocument(t, filepath.Join(root, "README.md"))
	writeDocument(t, filepath.Join(root, "CHANGELOG.md"))

	deprecated := validModule()
	deprecated.Cohesion.LifecycleStatus = "deprecated"
	deprecated.Cohesion.Documentation.Changelog = nil
	assertDiagnosticSet(t, validateModule(root, 0, deprecated), []string{"missing-metadata:/modules/0/cohesion/documentation/changelog"})

	afterOptionalNil := validModule()
	afterOptionalNil.Cohesion.Documentation.Adoption = nil
	afterOptionalNil.Cohesion.Documentation.Changelog = nil
	assertDiagnosticSet(t, validateModule(root, 0, afterOptionalNil), []string{"missing-metadata:/modules/0/cohesion/documentation/changelog"})

	afterHTTPS := validModule()
	afterHTTPS.Cohesion.Documentation.Changelog = nil
	assertDiagnosticSet(t, validateModule(root, 0, afterHTTPS), []string{"missing-metadata:/modules/0/cohesion/documentation/changelog"})

	afterUnsafe := validModule()
	afterUnsafe.Cohesion.Documentation.README = new("../README.md")
	afterUnsafe.Cohesion.Documentation.Changelog = nil
	assertDiagnosticSet(t, validateModule(root, 0, afterUnsafe), []string{
		"missing-metadata:/modules/0/cohesion/documentation/changelog",
		"unsafe-path:/modules/0/cohesion/documentation/readme",
	})
}

func TestValidateModuleRequiresPlannedDocumentationEntryPoint(t *testing.T) {
	root := t.TempDir()
	module := validModule()
	module.Cohesion.LifecycleStatus = "planned"
	module.Cohesion.Maturity = "preview"
	module.Cohesion.Documentation = inventory.Documentation{}

	assertDiagnosticSet(t, validateModule(root, 0, module), []string{
		"missing-metadata:/modules/0/cohesion/documentation/readme",
	})
}

func TestValidateModuleRequiresPublicIdentifierAtModuleRoot(t *testing.T) {
	root := t.TempDir()
	writeDocument(t, filepath.Join(root, "README.md"))
	writeDocument(t, filepath.Join(root, "CHANGELOG.md"))
	module := validModule()
	module.Packages = append(module.Packages, inventory.Package{
		ModuleDirectory: ".", Directory: "sub", Name: "sub", ImportPath: "example.com/library/sub", Kind: "public",
	})
	module.Cohesion.PublicPackageIdentifier = "sub"
	module.Cohesion.PrimaryEntryPackages = []string{"example.com/library/sub"}

	assertDiagnosticSet(t, validateModule(root, 0, module), []string{
		"invalid-reference:/modules/0/cohesion/public_package_identifier",
	})
}

func TestCheckReportsUnreadableManifestWithoutInventedCounts(t *testing.T) {
	report := Check(t.TempDir(), config.Config{Manifests: config.Manifests{Modules: "modules.json", Packages: "packages.json"}})
	if report.Valid || report.Repository != nil || report.Summary.TotalModules != nil || report.Summary.ErrorCount != 1 || report.Diagnostics[0].Code != "invalid-manifest" {
		t.Fatalf("Check(unreadable) = %#v", report)
	}
}

func TestRequiredMetadataInspectionFailsClosedWithoutInventingDiagnostics(t *testing.T) {
	for _, content := range []string{
		"{",
		`{"schema_version":1,"modules":[]}`,
		`{"schema_version":1,"modules":[{"cohesion":{}}]}`,
		`{"schema_version":2,"modules":[1]}`,
		`{"schema_version":2,"modules":[{}]}`,
		`{"schema_version":2,"modules":[{"cohesion":null}]}`,
		`{"schema_version":2,"modules":[{"cohesion":1}]}`,
	} {
		if diagnostics := requiredMetadataDiagnosticsData([]byte(content)); len(diagnostics) != 0 {
			t.Fatalf("requiredMetadataDiagnosticsData(%q) = %#v", content, diagnostics)
		}
	}
	if diagnostics := requiredMetadataDiagnosticsData([]byte(`{"schema_version":2,"modules":[{"cohesion":{"ownership":"invalid","supported_go":"invalid","documentation":"invalid","delivery":"invalid"}}]}`)); len(diagnostics) == 0 {
		t.Fatal("invalid nested objects produced no missing-field diagnostics")
	}
	var diagnostics []Diagnostic
	appendMissingObjectFields(&diagnostics, map[string]json.RawMessage{}, "nested", []string{"field"}, "")
	appendMissingObjectFields(&diagnostics, map[string]json.RawMessage{"nested": json.RawMessage(`1`)}, "nested", []string{"field"}, "")
	appendMissingObjectFields(&diagnostics, map[string]json.RawMessage{"nested": json.RawMessage(`{}`)}, "nested", []string{"field"}, "")
	if len(diagnostics) != 1 {
		t.Fatalf("appendMissingObjectFields diagnostics = %#v", diagnostics)
	}
}

func TestRequiredMetadataInspectionContinuesPastNonMetadataEntries(t *testing.T) {
	for _, data := range []string{
		`{"schema_version":2,"modules":[1,{"cohesion":{}}]}`,
		`{"schema_version":2,"modules":[{},{"cohesion":{}}]}`,
		`{"schema_version":2,"modules":[{"cohesion":null},{"cohesion":{}}]}`,
		`{"schema_version":2,"modules":[{"cohesion":1},{"cohesion":{}}]}`,
	} {
		diagnostics := requiredMetadataDiagnosticsData([]byte(data))
		if len(diagnostics) != len(requiredCohesionFields) {
			t.Fatalf("requiredMetadataDiagnosticsData(%s) returned %d diagnostics", data, len(diagnostics))
		}
		for _, diagnostic := range diagnostics {
			if !strings.HasPrefix(diagnostic.Path, "/modules/1/cohesion/") {
				t.Fatalf("requiredMetadataDiagnosticsData(%s) = %#v", data, diagnostics)
			}
		}
	}
}

func TestDiagnosticOrderingUsesPathCodeAndMessage(t *testing.T) {
	diagnostics := []Diagnostic{
		{Path: "/b", Code: "b", Message: "b"},
		{Path: "/a", Code: "b", Message: "b"},
		{Path: "/a", Code: "a", Message: "z"},
		{Path: "/a", Code: "a", Message: "a"},
	}
	sortDiagnostics(diagnostics)
	got := make([]string, 0, len(diagnostics))
	for _, diagnostic := range diagnostics {
		got = append(got, diagnostic.Path+diagnostic.Code+diagnostic.Message)
	}
	if strings.Join(got, ",") != "/aaa,/aaz,/abb,/bbb" {
		t.Fatalf("sortDiagnostics() = %v", got)
	}
}

func TestCatalogIdentityAndMarkdownContracts(t *testing.T) {
	identity := Identity{DesignLanguageVersion: "1.0", DesignLanguageSHA256: strings.Repeat("a", 64), SourceIdentity: "unpublished", ToolingVersion: "dev", PublicationStatus: "unpublished"}
	module := validModule()
	second := validModule()
	second.ModulePath = "example.com/zeta"
	second.Packages[0].ImportPath = second.ModulePath
	second.Cohesion.PrimaryEntryPackages = []string{second.ModulePath}
	second.Cohesion.Responsibility = "Provide zeta values."
	catalog := inventory.Inventory{Repository: "example.com/repository", Modules: []inventory.Module{second, module}}
	consumer, err := Project(catalog, "consumer", identity)
	if err != nil {
		t.Fatal(err)
	}
	consumerMarkdown, err := RenderMarkdown(consumer)
	wantConsumer := "# Golib Consumer Catalog\n\nDesign language `1.0` (`unpublished`); tooling `dev`.\n\n## foundations\n\n- `example.com/library`: Provide example values.\n\n- `example.com/zeta`: Provide zeta values.\n"
	if err != nil || string(consumerMarkdown) != wantConsumer || strings.Count(string(consumerMarkdown), "## foundations") != 1 {
		t.Fatalf("RenderMarkdown(consumer) = %q, %v", consumerMarkdown, err)
	}

	module.Releasable = false
	module.Cohesion = nil
	catalog.Modules = []inventory.Module{validModule(), module}
	engineering, err := Project(catalog, "engineering", identity)
	if err != nil {
		t.Fatal(err)
	}
	engineeringMarkdown, err := RenderMarkdown(engineering)
	if err != nil || !strings.Contains(string(engineeringMarkdown), "unclassified-internal") {
		t.Fatalf("RenderMarkdown(engineering) = %q, %v", engineeringMarkdown, err)
	}
	if _, err := RenderMarkdown(Envelope{Modules: "invalid"}); err == nil {
		t.Fatal("RenderMarkdown(invalid) error = nil")
	}

	published := identity
	published.SourceIdentity = "v1.2.3"
	published.ToolingVersion = "v1.2.3"
	published.PublicationStatus = "published"
	if _, err := Project(catalog, "engineering", published); err != nil {
		t.Fatalf("Project(published) error = %v", err)
	}
	for _, invalid := range []Identity{
		{},
		{DesignLanguageVersion: "2.0", DesignLanguageSHA256: strings.Repeat("a", 64)},
		{DesignLanguageVersion: "1.0", DesignLanguageSHA256: strings.Repeat("a", 63), SourceIdentity: "unpublished", ToolingVersion: "dev", PublicationStatus: "unpublished"},
		{DesignLanguageVersion: "1.0", DesignLanguageSHA256: strings.Repeat("A", 64)},
		{DesignLanguageVersion: "1.0", DesignLanguageSHA256: strings.Repeat("a", 64), SourceIdentity: "unpublished", ToolingVersion: "dev", PublicationStatus: "published"},
		{DesignLanguageVersion: "1.0", DesignLanguageSHA256: strings.Repeat("a", 64), SourceIdentity: "v1.2.3", ToolingVersion: "dev", PublicationStatus: "unpublished"},
		{DesignLanguageVersion: "1.0", DesignLanguageSHA256: strings.Repeat("a", 64), SourceIdentity: "unpublished", ToolingVersion: "v1.2.3", PublicationStatus: "unpublished"},
		{DesignLanguageVersion: "1.0", DesignLanguageSHA256: strings.Repeat("a", 64), SourceIdentity: "v1", ToolingVersion: "v1.2.3", PublicationStatus: "published"},
	} {
		if _, err := Project(catalog, "engineering", invalid); err == nil {
			t.Fatalf("Project(invalid identity %#v) error = nil", invalid)
		}
	}
	if _, err := Project(catalog, "unknown", identity); err == nil {
		t.Fatal("Project(unknown view) error = nil")
	}
}

func TestSemverTagRejectsMalformedValues(t *testing.T) {
	for value, want := range map[string]bool{
		"v0.0.0": true, "v1.2.3": true, "1.2.3": false, "v1.2": false,
		"v01.2.3": false, "v1.02.3": false, "v1.2.03": false,
		"x1.2.3": false, "v1.2.3.4": false, "v1..3": false, "v1.x.3": false, "v": false,
	} {
		if got := semverTag(value); got != want {
			t.Fatalf("semverTag(%q) = %t", value, got)
		}
	}
	if got := moduleFamilyOrder(inventory.Module{Cohesion: &inventory.Cohesion{Family: "unknown"}}); got != len(familyOrder) {
		t.Fatalf("moduleFamilyOrder(unknown) = %d", got)
	}
}

func validModule() inventory.Module {
	readme := "README.md"
	changelog := "CHANGELOG.md"
	api := "https://pkg.go.dev/example.com/library"
	ecosystem := "https://example.com/ecosystem"
	return inventory.Module{
		Directory: ".", ModulePath: "example.com/library", GoVersion: "1.27.0",
		Kind: "public library", Purpose: "Example library.", Lifecycle: "stable",
		Releasable: true, Version: "1.0.0", TagPrefix: "v", Gates: map[string]bool{},
		Specifications: []string{}, OwnedDependencies: []string{"example.com/required"},
		ReverseOwnedDependencies: []string{}, Packages: []inventory.Package{{
			ModuleDirectory: ".", Directory: ".", Name: "library", ImportPath: "example.com/library", Kind: "public",
		}},
		Family: "foundations", GoalFiles: []string{}, GoalEvidence: []inventory.GoalEvidence{}, Provenance: []byte("[]"),
		Cohesion: &inventory.Cohesion{
			Family: "foundations", SecondaryCapabilities: []string{"testing-and-conformance"},
			Responsibility: "Provide example values.", NonGoals: []string{"Own application state."},
			PublicPackageIdentifier: "library", PrimaryEntryPackages: []string{"example.com/library"}, PackageSelection: map[string]string{},
			LifecycleStatus: "active", Maturity: "stable", ConstructionStyles: []string{"plain-function"}, LifecycleStyles: []string{"stateless"},
			Ownership:                 inventory.Ownership{Configuration: "caller", MutableInputs: []string{"copy"}, RuntimeResources: "none", BackgroundWork: "none"},
			OptionalOwnedDependencies: []string{}, Adapters: []string{}, Companions: []string{},
			SupportedGo:        inventory.SupportedGo{Minimum: "1.27.0", Tested: []string{"1.27.0"}},
			SupportedPlatforms: []string{"portable-go"}, SupportedBackends: []string{}, SupportedProtocols: []string{},
			Documentation:              inventory.Documentation{README: &readme, API: &api, Changelog: &changelog, PkgGoDev: &api, EcosystemIndex: &ecosystem},
			KnownGoodCompatibilitySets: []string{}, Delivery: inventory.Delivery{Implementation: "in-progress", Hardening: "in-progress", Release: "in-progress"},
		},
	}
}

func writeDocument(t *testing.T, path string) {
	t.Helper()
	if err := os.WriteFile(path, []byte("# Document\n"), 0o600); err != nil {
		t.Fatal(err)
	}
}

func diagnosticExists(diagnostics []Diagnostic, code, path string) bool {
	for _, diagnostic := range diagnostics {
		if diagnostic.Code == code && diagnostic.Path == path {
			return true
		}
	}
	return false
}

func assertDiagnosticSet(t *testing.T, diagnostics []Diagnostic, want []string) {
	t.Helper()
	got := make([]string, 0, len(diagnostics))
	for _, diagnostic := range diagnostics {
		got = append(got, diagnostic.Code+":"+diagnostic.Path)
	}
	sort.Strings(got)
	sort.Strings(want)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("diagnostics = %v, want %v", got, want)
	}
}
