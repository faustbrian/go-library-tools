package cohesion

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckSourcesV2UsesStrictProductionPipelineAndNoFollowRead(t *testing.T) {
	root := t.TempDir()
	resolvedRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}
	root = resolvedRoot
	input := filepath.Join(root, "sources.json")
	if err := os.WriteFile(input, []byte(`{}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := CheckSourcesV2(input); err == nil {
		t.Fatal("CheckSourcesV2({}) error = nil")
	} else if diagnostic, ok := DiagnosticFromV3Error(err); !ok || diagnostic.Code != "schema-required-member" {
		t.Fatalf("CheckSourcesV2({}) error = %T %v, diagnostic = %#v", err, err, diagnostic)
	}
	link := filepath.Join(root, "sources-link.json")
	if err := os.Symlink(input, link); err != nil {
		t.Fatal(err)
	}
	if err := CheckSourcesV2(link); err == nil {
		t.Fatal("CheckSourcesV2(symlink) error = nil")
	}
}

func TestVerifySourcesV2RejectsUnreadableInputsAndInvocationMaps(t *testing.T) {
	t.Run("missing source-lock input", func(t *testing.T) {
		err := VerifySourcesV2(filepath.Join(t.TempDir(), "missing.json"), "github.com/faustbrian/example", filepath.Join(t.TempDir(), "map.json"))
		if err == nil || err.Error() != "resolve source-lock input" {
			t.Fatalf("VerifySourcesV2(missing input) error = %v", err)
		}
	})

	t.Run("invalid source-lock stops before invocation map", func(t *testing.T) {
		root := canonicalTempDir(t)
		input := filepath.Join(root, "sources.json")
		if err := os.WriteFile(input, []byte(`{}`), 0o600); err != nil {
			t.Fatal(err)
		}
		// The source lock is validated before the invocation map is read.  Keep
		// the map path absent to ensure this assertion protects that ordering.
		err := VerifySourcesV2(input, "github.com/faustbrian/example", filepath.Join(root, "missing-map.json"))
		if err == nil {
			t.Fatal("VerifySourcesV2(invalid source lock) error = nil")
		}
		if diagnostic, ok := DiagnosticFromV3Error(err); !ok || diagnostic.Code != "schema-required-member" {
			t.Fatalf("VerifySourcesV2(invalid source lock) error = %v, diagnostic = %#v", err, diagnostic)
		}
	})
}

func TestVerifyLockedCommitSourceResolvesPinnedManifestFromObjectTree(t *testing.T) {
	root := canonicalTempDir(t)
	runGitResolverTestCommand(t, root, "init", "-q")
	runGitResolverTestCommand(t, root, "config", "user.name", "Test")
	runGitResolverTestCommand(t, root, "config", "user.email", "test@example.invalid")
	runGitResolverTestCommand(t, root, "remote", "add", "origin", "https://github.com/faustbrian/example.git")
	manifest := []byte("{}\n")
	if err := os.WriteFile(filepath.Join(root, "modules.json"), manifest, 0o600); err != nil {
		t.Fatal(err)
	}
	runGitResolverTestCommand(t, root, "add", "modules.json")
	runGitResolverTestCommand(t, root, "commit", "-q", "-m", "fixture")
	revision := strings.TrimSpace(runGitResolverTestCommand(t, root, "rev-parse", "HEAD"))
	source := map[string]any{
		"repository":      "github.com/faustbrian/example",
		"source_kind":     "commit",
		"source_revision": revision,
		"manifest_path":   "modules.json",
		"manifest_sha256": exactBytesSHA256(manifest),
	}
	resolution := ResolutionMapV1{Sources: []ResolutionSourceV1{{Repository: "github.com/faustbrian/example", SourceRevision: revision, Root: root}}}

	if err := verifyLockedSourceManifest(source, resolution); err != nil {
		t.Fatalf("verifyLockedSourceManifest() error = %v", err)
	}
	source["manifest_sha256"] = "sha256:" + strings.Repeat("f", 64)
	if err := verifyLockedSourceManifest(source, resolution); err == nil {
		t.Fatal("verifyLockedSourceManifest(wrong digest) error = nil")
	}
}

func TestVerifyLockedReleaseSourceResolvesEveryPinnedSchema(t *testing.T) {
	root := canonicalTempDir(t)
	runGitResolverTestCommand(t, root, "init", "-q")
	runGitResolverTestCommand(t, root, "config", "user.name", "Test")
	runGitResolverTestCommand(t, root, "config", "user.email", "test@example.invalid")
	runGitResolverTestCommand(t, root, "remote", "add", "origin", "https://github.com/faustbrian/go-library-tools.git")
	manifest := []byte("{}\n")
	schema := []byte("{\"type\":\"object\"}\n")
	if err := os.Mkdir(filepath.Join(root, "schema"), 0o700); err != nil {
		t.Fatal(err)
	}
	for path, content := range map[string][]byte{"modules.json": manifest, "schema/example.schema.json": schema} {
		if err := os.WriteFile(filepath.Join(root, path), content, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	runGitResolverTestCommand(t, root, "add", "modules.json", "schema/example.schema.json")
	runGitResolverTestCommand(t, root, "commit", "-q", "-m", "fixture")
	revision := strings.TrimSpace(runGitResolverTestCommand(t, root, "rev-parse", "HEAD"))
	source := map[string]any{
		"repository":      "github.com/faustbrian/go-library-tools",
		"source_kind":     "release-source",
		"manifest_path":   "modules.json",
		"manifest_sha256": exactBytesSHA256(manifest),
		"tooling":         map[string]any{"peeled_commit": revision},
		"schemas": []any{map[string]any{
			"path": "schema/example.schema.json", "bytes_sha256": exactBytesSHA256(schema),
		}},
	}
	resolution := ResolutionMapV1{Sources: []ResolutionSourceV1{{Repository: "github.com/faustbrian/go-library-tools", SourceRevision: revision, Root: root}}}

	if err := verifyLockedSourceManifest(source, resolution); err != nil {
		t.Fatalf("verifyLockedSourceManifest(release source) error = %v", err)
	}
	source["schemas"].([]any)[0].(map[string]any)["bytes_sha256"] = "sha256:" + strings.Repeat("f", 64)
	if err := verifyLockedSourceManifest(source, resolution); err == nil {
		t.Fatal("verifyLockedSourceManifest(wrong schema digest) error = nil")
	}
}

func TestVerifySourceRosterRequiresExactCurrentMembership(t *testing.T) {
	root := canonicalTempDir(t)
	runGitResolverTestCommand(t, root, "init", "-q")
	runGitResolverTestCommand(t, root, "config", "user.name", "Test")
	runGitResolverTestCommand(t, root, "config", "user.email", "test@example.invalid")
	runGitResolverTestCommand(t, root, "remote", "add", "origin", "https://github.com/faustbrian/go-library-tools.git")
	roster := []byte(`{"schema_version":1,"repository_count":2,"repositories":[{"repository":"github.com/faustbrian/go-a","source":{"kind":"commit","commit":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},"tooling":{"version":"v1.5.0","checksums_sha256":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}},{"repository":"github.com/faustbrian/go-library-tools","source":{"kind":"release-source"}}]}`)
	if err := os.WriteFile(filepath.Join(root, "roster.json"), roster, 0o600); err != nil {
		t.Fatal(err)
	}
	runGitResolverTestCommand(t, root, "add", "roster.json")
	runGitResolverTestCommand(t, root, "commit", "-q", "-m", "fixture")
	revision := strings.TrimSpace(runGitResolverTestCommand(t, root, "rev-parse", "HEAD"))
	control := map[string]any{"repository": "github.com/faustbrian/go-library-tools", "path": "roster.json", "source_revision": revision, "bytes_sha256": exactBytesSHA256(roster)}
	resolution := ResolutionMapV1{Sources: []ResolutionSourceV1{{Repository: "github.com/faustbrian/go-library-tools", SourceRevision: revision, Root: root}}}
	current := []any{map[string]any{"repository": "github.com/faustbrian/go-a", "source_kind": "commit"}, map[string]any{"repository": "github.com/faustbrian/go-library-tools", "source_kind": "release-source"}}

	if err := verifySourceRoster(control, current, resolution); err != nil {
		t.Fatalf("verifySourceRoster() error = %v", err)
	}
	current[0].(map[string]any)["repository"] = "github.com/faustbrian/go-b"
	if err := verifySourceRoster(control, current, resolution); err == nil {
		t.Fatal("verifySourceRoster(substitution) error = nil")
	}
}
