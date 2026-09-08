//nolint:forcetypeassert,modernize // controlled fixture map copying
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
	source["manifest_sha256"] = exactBytesSHA256(manifest)
	source["manifest_path"] = "missing-modules.json"
	if err := verifyLockedSourceManifest(source, resolution); err == nil || err.Error() != "resolve locked source manifest" {
		t.Fatalf("verifyLockedSourceManifest(missing manifest) error = %v", err)
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

func TestVerifySourceRosterRejectsResolutionAndContentFailures(t *testing.T) {
	root := canonicalTempDir(t)
	runGitResolverTestCommand(t, root, "init", "-q")
	runGitResolverTestCommand(t, root, "config", "user.name", "Test")
	runGitResolverTestCommand(t, root, "config", "user.email", "test@example.invalid")
	runGitResolverTestCommand(t, root, "remote", "add", "origin", "https://github.com/faustbrian/go-library-tools.git")
	roster := []byte(`{"schema_version":1,"repository_count":1,"repositories":[{"repository":"github.com/faustbrian/go-a","source":{"kind":"commit","commit":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},"tooling":{"version":"v1.5.0","checksums_sha256":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}}]}`)
	if err := os.WriteFile(filepath.Join(root, "roster.json"), roster, 0o600); err != nil {
		t.Fatal(err)
	}
	runGitResolverTestCommand(t, root, "add", "roster.json")
	runGitResolverTestCommand(t, root, "commit", "-q", "-m", "fixture")
	revision := strings.TrimSpace(runGitResolverTestCommand(t, root, "rev-parse", "HEAD"))
	control := map[string]any{"repository": "github.com/faustbrian/go-library-tools", "path": "roster.json", "source_revision": revision, "bytes_sha256": exactBytesSHA256(roster)}
	current := []any{map[string]any{"repository": "github.com/faustbrian/go-a"}}

	cases := []struct {
		name       string
		resolution ResolutionMapV1
		control    map[string]any
		current    []any
		want       string
	}{
		{name: "missing identity", resolution: ResolutionMapV1{}, control: control, current: current, want: "invocation map is missing the source-roster identity"},
		{name: "missing bytes", resolution: ResolutionMapV1{Sources: []ResolutionSourceV1{{Repository: control["repository"].(string), SourceRevision: revision, Root: root}}}, control: func() map[string]any { c := cloneMap(control); c["path"] = "missing.json"; return c }(), current: current, want: "resolve source-roster bytes"},
		{name: "digest mismatch", resolution: ResolutionMapV1{Sources: []ResolutionSourceV1{{Repository: control["repository"].(string), SourceRevision: revision, Root: root}}}, control: func() map[string]any {
			c := cloneMap(control)
			c["bytes_sha256"] = "sha256:" + strings.Repeat("f", 64)
			return c
		}(), current: current, want: "source-roster digest does not match"},
		{name: "membership count", resolution: ResolutionMapV1{Sources: []ResolutionSourceV1{{Repository: control["repository"].(string), SourceRevision: revision, Root: root}}}, control: control, current: []any{}, want: "source-roster membership count does not match"},
		{name: "membership identity", resolution: ResolutionMapV1{Sources: []ResolutionSourceV1{{Repository: control["repository"].(string), SourceRevision: revision, Root: root}}}, control: control, current: []any{map[string]any{"repository": "github.com/faustbrian/go-b"}}, want: "source-roster membership does not match"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := verifySourceRoster(tc.control, tc.current, tc.resolution); err == nil || err.Error() != tc.want {
				t.Fatalf("verifySourceRoster() error = %v, want %q", err, tc.want)
			}
		})
	}
	if err := os.WriteFile(filepath.Join(root, "invalid.json"), []byte(`{}`), 0o600); err != nil {
		t.Fatal(err)
	}
	runGitResolverTestCommand(t, root, "add", "invalid.json")
	runGitResolverTestCommand(t, root, "commit", "-q", "-m", "invalid fixture")
	invalidRevision := strings.TrimSpace(runGitResolverTestCommand(t, root, "rev-parse", "HEAD"))
	invalidControl := map[string]any{"repository": control["repository"], "path": "invalid.json", "source_revision": invalidRevision, "bytes_sha256": exactBytesSHA256([]byte(`{}`))}
	invalidResolution := ResolutionMapV1{Sources: []ResolutionSourceV1{{Repository: control["repository"].(string), SourceRevision: invalidRevision, Root: root}}}
	if err := verifySourceRoster(invalidControl, current, invalidResolution); err == nil || err.Error() != "source-roster schema is invalid" {
		t.Fatalf("verifySourceRoster(invalid schema) error = %v", err)
	}
}

func TestLockedSourceResolutionRejectsUnknownIdentity(t *testing.T) {
	source := map[string]any{"repository": "github.com/faustbrian/example", "source_kind": "commit", "source_revision": strings.Repeat("a", 40)}
	if _, err := lockedSourceResolution(source, ResolutionMapV1{}); err == nil || err.Error() != "invocation map is missing a locked source identity" {
		t.Fatalf("lockedSourceResolution() error = %v", err)
	}
	if err := verifyLockedSourceManifest(source, ResolutionMapV1{}); err == nil || err.Error() != "invocation map is missing a locked source identity" {
		t.Fatalf("verifyLockedSourceManifest(unknown identity) error = %v", err)
	}
}

func cloneMap(input map[string]any) map[string]any {
	output := make(map[string]any, len(input))
	for key, value := range input {
		output[key] = value
	}
	return output
}
