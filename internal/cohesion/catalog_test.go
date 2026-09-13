package cohesion_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/faustbrian/go-library-tools/internal/cohesion"
	"github.com/faustbrian/go-library-tools/internal/inventory"
)

type projectedModule struct {
	ModulePath string `json:"module_path"`
}

func TestCatalogSeparatesConsumerAndEngineeringViewsDeterministically(t *testing.T) {
	catalog := inventory.Inventory{
		SchemaVersion: 2,
		Repository:    "github.com/faustbrian/example",
		GoVersion:     "1.27.0",
		Modules: []inventory.Module{
			module("github.com/faustbrian/example/tool", "public tool", true, "tooling", "active"),
			module("github.com/faustbrian/example/zeta", "public library", true, "tooling", "active"),
			module("github.com/faustbrian/example/alpha", "adapter", true, "foundations", "deprecated"),
			module("github.com/faustbrian/example/planned", "public library", true, "service-edge", "planned"),
			module("github.com/faustbrian/example/internal", "fixture", false, "", ""),
		},
	}
	identity := cohesion.Identity{DesignLanguageVersion: "1.0", DesignLanguageSHA256: strings.Repeat("a", 64), SourceIdentity: "unpublished", ToolingVersion: "dev", PublicationStatus: "unpublished"}

	consumer, err := cohesion.Project(catalog, "consumer", identity)
	if err != nil {
		t.Fatal(err)
	}
	consumerJSON, err := json.Marshal(consumer)
	if err != nil {
		t.Fatal(err)
	}
	var consumerResult struct {
		Modules []projectedModule `json:"modules"`
	}
	if err := json.Unmarshal(consumerJSON, &consumerResult); err != nil {
		t.Fatal(err)
	}
	if got := modulePaths(consumerResult.Modules); strings.Join(got, ",") != "github.com/faustbrian/example/alpha,github.com/faustbrian/example/zeta" {
		t.Fatalf("consumer modules = %v", got)
	}

	engineering, err := cohesion.Project(catalog, "engineering", identity)
	if err != nil {
		t.Fatal(err)
	}
	engineeringJSON, err := json.Marshal(engineering)
	if err != nil {
		t.Fatal(err)
	}
	var engineeringResult struct {
		Modules []projectedModule `json:"modules"`
	}
	if err := json.Unmarshal(engineeringJSON, &engineeringResult); err != nil {
		t.Fatal(err)
	}
	want := "github.com/faustbrian/example/alpha,github.com/faustbrian/example/planned,github.com/faustbrian/example/tool,github.com/faustbrian/example/zeta,github.com/faustbrian/example/internal"
	if got := modulePaths(engineeringResult.Modules); strings.Join(got, ",") != want {
		t.Fatalf("engineering modules = %v", got)
	}
}

func TestConsumerCatalogExcludesEachNonConsumerBoundary(t *testing.T) {
	visible := module("github.com/faustbrian/example/visible", "public library", true, "foundations", "active")
	nonReleasable := module("github.com/faustbrian/example/non-releasable", "public library", false, "foundations", "active")
	missingMetadata := module("github.com/faustbrian/example/missing-metadata", "public library", true, "", "")
	catalog := inventory.Inventory{Repository: "github.com/faustbrian/example", Modules: []inventory.Module{nonReleasable, missingMetadata, visible}}
	identity := cohesion.Identity{DesignLanguageVersion: "1.0", DesignLanguageSHA256: strings.Repeat("a", 64), SourceIdentity: "unpublished", ToolingVersion: "dev", PublicationStatus: "unpublished"}

	consumer, err := cohesion.Project(catalog, "consumer", identity)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(consumer)
	if err != nil {
		t.Fatal(err)
	}
	var result struct {
		Modules []projectedModule `json:"modules"`
	}
	if err := json.Unmarshal(encoded, &result); err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(modulePaths(result.Modules), ","); got != visible.ModulePath {
		t.Fatalf("consumer modules = %q", got)
	}
}

func TestConsumerCatalogMarkdownShowsLifecycleAndNavigation(t *testing.T) {
	readme := "adapters/legacy/README.md"
	adoption := "adapters/legacy/docs/migration.md"
	pkgGoDev := "https://pkg.go.dev/github.com/faustbrian/example/adapters/legacy"
	legacy := module("github.com/faustbrian/example/adapters/legacy", "adapter", true, "foundations", "deprecated")
	legacy.Directory = "adapters/legacy"
	legacy.Version = "1.2.3"
	legacy.Cohesion.Documentation = inventory.Documentation{
		README: &readme, Adoption: &adoption, PkgGoDev: &pkgGoDev,
	}
	catalog := inventory.Inventory{Repository: "github.com/faustbrian/example", Modules: []inventory.Module{legacy}}
	identity := cohesion.Identity{DesignLanguageVersion: "1.0", DesignLanguageSHA256: strings.Repeat("a", 64), SourceIdentity: "unpublished", ToolingVersion: "dev", PublicationStatus: "unpublished"}

	consumer, err := cohesion.Project(catalog, "consumer", identity)
	if err != nil {
		t.Fatal(err)
	}
	rendered, err := cohesion.RenderMarkdown(consumer)
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{
		"**deprecated**",
		"[github.com/faustbrian/example/adapters/legacy](https://pkg.go.dev/github.com/faustbrian/example/adapters/legacy@v1.2.3)",
		"[README](https://github.com/faustbrian/example/blob/adapters/legacy/v1.2.3/adapters/legacy/README.md)",
		"[migration](https://github.com/faustbrian/example/blob/adapters/legacy/v1.2.3/adapters/legacy/docs/migration.md)",
	} {
		if !strings.Contains(string(rendered), expected) {
			t.Fatalf("consumer Markdown lacks %q: %s", expected, rendered)
		}
	}
}

func TestConsumerCatalogMarkdownPreservesAbsoluteDocumentationURLs(t *testing.T) {
	readme := "https://docs.example.com/library/README.md"
	public := module("github.com/faustbrian/example/library", "public library", true, "foundations", "active")
	public.Directory = "."
	public.Version = "1.0.0"
	public.Cohesion.Documentation.README = &readme
	catalog := inventory.Inventory{Repository: "github.com/faustbrian/example", Modules: []inventory.Module{public}}
	identity := cohesion.Identity{DesignLanguageVersion: "1.0", DesignLanguageSHA256: strings.Repeat("a", 64), SourceIdentity: "unpublished", ToolingVersion: "dev", PublicationStatus: "unpublished"}

	projected, err := cohesion.Project(catalog, "consumer", identity)
	if err != nil {
		t.Fatal(err)
	}
	rendered, err := cohesion.RenderMarkdown(projected)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(rendered), "[README]("+readme+")") {
		t.Fatalf("RenderMarkdown() did not preserve absolute README URL: %s", rendered)
	}
}

func module(path, kind string, releasable bool, family, lifecycle string) inventory.Module {
	module := inventory.Module{Directory: path, ModulePath: path, GoVersion: "1.27.0", Kind: kind, Releasable: releasable}
	if family != "" {
		module.Cohesion = &inventory.Cohesion{Family: family, LifecycleStatus: lifecycle, Responsibility: path}
	}
	return module
}

func modulePaths(modules []projectedModule) []string {
	paths := make([]string, 0, len(modules))
	for _, module := range modules {
		paths = append(paths, module.ModulePath)
	}
	return paths
}
