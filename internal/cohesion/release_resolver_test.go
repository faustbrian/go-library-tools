package cohesion

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveReleaseAssetVerifiesExactBytesWithoutFollowingLinks(t *testing.T) {
	root := canonicalTempDir(t)
	asset := filepath.Join(root, "asset.json")
	content := []byte("release asset\n")
	if err := os.WriteFile(asset, content, 0o600); err != nil {
		t.Fatal(err)
	}
	digest := exactBytesSHA256(content)

	got, err := ResolveReleaseAsset(root, "asset.json", digest, 1024)
	if err != nil {
		t.Fatalf("ResolveReleaseAsset() error = %v", err)
	}
	if string(got) != string(content) {
		t.Fatalf("ResolveReleaseAsset() = %q", got)
	}
	if _, err := ResolveReleaseAsset(root, "asset.json", "sha256:"+strings.Repeat("f", 64), 1024); err == nil {
		t.Fatal("ResolveReleaseAsset(wrong digest) error = nil")
	}

	link := filepath.Join(root, "linked.json")
	if err := os.Symlink(asset, link); err != nil {
		t.Fatal(err)
	}
	if _, err := ResolveReleaseAsset(root, "linked.json", digest, 1024); err == nil {
		t.Fatal("ResolveReleaseAsset(symlink) error = nil")
	}
}

func TestVerifyGitReleaseTagRequiresExactAnnotatedTagAndPeel(t *testing.T) {
	repository := canonicalTempDir(t)
	runGitResolverTestCommand(t, repository, "init")
	runGitResolverTestCommand(t, repository, "config", "user.email", "test@example.com")
	runGitResolverTestCommand(t, repository, "config", "user.name", "Test")
	runGitResolverTestCommand(t, repository, "remote", "add", "origin", "https://github.com/faustbrian/example.git")
	if err := os.WriteFile(filepath.Join(repository, "modules.json"), []byte("{}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	runGitResolverTestCommand(t, repository, "add", "modules.json")
	runGitResolverTestCommand(t, repository, "commit", "-m", "fixture")
	runGitResolverTestCommand(t, repository, "tag", "-a", "v1.2.3", "-m", "release")
	commit := strings.TrimSpace(runGitResolverTestCommand(t, repository, "rev-parse", "HEAD"))
	tag := strings.TrimSpace(runGitResolverTestCommand(t, repository, "rev-parse", "v1.2.3^{tag}"))

	if err := VerifyGitReleaseTag(repository, "github.com/faustbrian/example", "v1.2.3", tag, commit); err != nil {
		t.Fatalf("VerifyGitReleaseTag() error = %v", err)
	}
	if err := VerifyGitReleaseTag(repository, "github.com/faustbrian/example", "v1.2.4", tag, commit); err == nil {
		t.Fatal("VerifyGitReleaseTag(wrong tag name) error = nil")
	}
	if err := VerifyGitReleaseTag(repository, "github.com/faustbrian/example", "v1.2.3", commit, commit); err == nil {
		t.Fatal("VerifyGitReleaseTag(lightweight tag) error = nil")
	}
}

func TestVerifyReleasedToolingIdentityBindsTagAndAllReleaseAssets(t *testing.T) {
	repository := canonicalTempDir(t)
	runGitResolverTestCommand(t, repository, "init", "-q")
	runGitResolverTestCommand(t, repository, "config", "user.email", "test@example.com")
	runGitResolverTestCommand(t, repository, "config", "user.name", "Test")
	runGitResolverTestCommand(t, repository, "remote", "add", "origin", "https://github.com/faustbrian/go-library-tools.git")
	if err := os.WriteFile(filepath.Join(repository, "modules.json"), []byte("{}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	runGitResolverTestCommand(t, repository, "add", "modules.json")
	runGitResolverTestCommand(t, repository, "commit", "-q", "-m", "fixture")
	runGitResolverTestCommand(t, repository, "tag", "-a", "v1.6.0", "-m", "release")
	commit := strings.TrimSpace(runGitResolverTestCommand(t, repository, "rev-parse", "HEAD"))
	tag := strings.TrimSpace(runGitResolverTestCommand(t, repository, "rev-parse", "v1.6.0^{tag}"))
	releaseRoot := canonicalTempDir(t)
	assets := map[string][]byte{"golib-darwin-arm64.tar.gz": []byte("binary"), "checksums.txt": []byte("checksums"), "release-manifest.json": []byte("manifest")}
	for name, content := range assets {
		if err := os.WriteFile(filepath.Join(releaseRoot, name), content, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	tooling := map[string]any{
		"repository": "github.com/faustbrian/go-library-tools", "release": "v1.6.0", "tag_object_sha": tag, "peeled_commit": commit,
		"executable_asset": "golib-darwin-arm64.tar.gz", "platform_artifact_sha256": exactBytesSHA256(assets["golib-darwin-arm64.tar.gz"]),
		"checksums_asset": "checksums.txt", "checksums_sha256": exactBytesSHA256(assets["checksums.txt"]),
		"release_manifest_asset": "release-manifest.json", "release_manifest_sha256": exactBytesSHA256(assets["release-manifest.json"]),
	}
	resolution := ResolutionMapV1{
		Sources:  []ResolutionSourceV1{{Repository: "github.com/faustbrian/go-library-tools", SourceRevision: commit, Root: repository}},
		Releases: []ResolutionReleaseV1{{Repository: "github.com/faustbrian/go-library-tools", Release: "v1.6.0", TagObject: tag, PeeledCommit: commit, Root: releaseRoot}},
	}
	if err := verifyReleasedToolingIdentity(tooling, resolution); err != nil {
		t.Fatalf("verifyReleasedToolingIdentity() error = %v", err)
	}
	tooling["checksums_sha256"] = "sha256:" + strings.Repeat("f", 64)
	if err := verifyReleasedToolingIdentity(tooling, resolution); err == nil {
		t.Fatal("verifyReleasedToolingIdentity(wrong digest) error = nil")
	}
}

func canonicalTempDir(t *testing.T) string {
	t.Helper()
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return root
}
