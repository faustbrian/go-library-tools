package cli

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/faustbrian/go-library-tools/internal/buildinfo"
	"github.com/faustbrian/go-library-tools/internal/cohesion"
	"github.com/faustbrian/go-library-tools/internal/config"
	"github.com/faustbrian/go-library-tools/internal/inventory"
)

// This runtime characterization is deliberately separate from the historical
// JSON provenance corpus. RenderMarkdown and the non-JSON catalog command have
// a raw UTF-8 output contract, not a parser/emitter semantic-object domain.
func TestHistoricalCatalogV1MarkdownRuntimeAndCLI(t *testing.T) {
	identity := cohesion.Identity{
		DesignLanguageVersion: "1.0", DesignLanguageSHA256: strings.Repeat("a", 64),
		SourceIdentity: "unpublished", ToolingVersion: "dev", PublicationStatus: "unpublished",
	}
	module := func(path, responsibility string) inventory.Module {
		return inventory.Module{
			Directory: path, ModulePath: path, GoVersion: "1.27.0", Kind: "public library", Releasable: true,
			Cohesion: &inventory.Cohesion{Family: "foundations", LifecycleStatus: "active", Responsibility: responsibility},
		}
	}
	catalog := inventory.Inventory{Repository: "example.com/repository", Modules: []inventory.Module{
		module("example.com/zeta", "Provide zeta values."),
		module("example.com/library", "Provide example values."),
	}}
	envelope, err := cohesion.Project(catalog, "consumer", identity)
	if err != nil {
		t.Fatal(err)
	}
	markdown, err := cohesion.RenderMarkdown(envelope)
	want := "# Golib Consumer Catalog\n\nDesign language `1.0` (`unpublished`); tooling `dev`.\n\n## foundations\n\n- `example.com/library`: Provide example values.\n\n- `example.com/zeta`: Provide zeta values.\n"
	if err != nil || string(markdown) != want {
		t.Fatalf("RenderMarkdown() = %q, %v", markdown, err)
	}
	if _, err := cohesion.RenderMarkdown(cohesion.Envelope{Modules: "invalid"}); err == nil {
		t.Fatal("RenderMarkdown(invalid) error = nil")
	}

	repositoryRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	policy, err := config.Load(repositoryRoot)
	if err != nil {
		t.Fatal(err)
	}
	repositoryCatalog, report := cohesion.LoadAndCheck(repositoryRoot, policy)
	if !report.Valid {
		t.Fatalf("frozen repository cohesion report is invalid: %#v", report.Diagnostics)
	}
	publicationStatus := "unpublished"
	if buildinfo.Version != "dev" {
		publicationStatus = "published"
	}
	repositoryEnvelope, err := cohesion.Project(repositoryCatalog, "engineering", cohesion.Identity{
		DesignLanguageVersion: buildinfo.DesignLanguageVersion,
		DesignLanguageSHA256:  buildinfo.DesignLanguageSHA256,
		SourceIdentity:        buildinfo.DesignLanguageSourceIdentity,
		ToolingVersion:        buildinfo.Version,
		PublicationStatus:     publicationStatus,
	})
	if err != nil {
		t.Fatal(err)
	}
	wantCLI, err := cohesion.RenderMarkdown(repositoryEnvelope)
	if err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if code := executeCohesion([]string{"catalog", "engineering"}, repositoryRoot, policy, &stdout, &stderr); code != 0 || stderr.Len() != 0 {
		t.Fatalf("executeCohesion() = %d, stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
	if !bytes.Equal(stdout.Bytes(), wantCLI) {
		t.Fatalf("CLI Markdown differs from direct renderer\nCLI: %q\nAPI: %q", stdout.Bytes(), wantCLI)
	}
}
