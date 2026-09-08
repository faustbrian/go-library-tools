package cohesion

import (
	"errors"
	"path/filepath"
	"strings"
)

const maximumResolvedReleaseAssetBytes = 512 << 20

// ResolveReleaseAsset reads one immutable release asset through the no-follow
// resolver and proves the caller-supplied exact-byte digest.
func ResolveReleaseAsset(root, asset, digest string, maximumBytes int64) ([]byte, error) {
	if !filepath.IsAbs(root) || filepath.Clean(root) != root || !safeReleaseAssetName(asset) || !validV3Digest(digest) || maximumBytes < 0 {
		return nil, errors.New("release asset identity is invalid")
	}
	content, err := readResolutionFile(filepath.Join(root, asset), maximumBytes)
	if err != nil {
		return nil, errors.New("resolve release asset")
	}
	if exactBytesSHA256(content) != digest {
		return nil, errors.New("release asset digest does not match")
	}
	return content, nil
}

func safeReleaseAssetName(asset string) bool {
	return asset != "" && asset != "." && asset != ".." && filepath.Base(asset) == asset && !strings.ContainsAny(asset, `/\\`)
}

func verifyReleasedToolingIdentity(tooling map[string]any, resolution ResolutionMapV1) error {
	repository := tooling["repository"].(string)
	release := tooling["release"].(string)
	tagObject := tooling["tag_object_sha"].(string)
	peeledCommit := tooling["peeled_commit"].(string)
	var releaseResolution *ResolutionReleaseV1
	for index := range resolution.Releases {
		candidate := &resolution.Releases[index]
		if candidate.Repository == repository && candidate.Release == release && candidate.TagObject == tagObject && candidate.PeeledCommit == peeledCommit {
			releaseResolution = candidate
			break
		}
	}
	if releaseResolution == nil {
		return errors.New("invocation map is missing a released-tool identity")
	}
	var sourceResolution *ResolutionSourceV1
	for index := range resolution.Sources {
		candidate := &resolution.Sources[index]
		if candidate.Repository == repository && candidate.SourceRevision == peeledCommit {
			sourceResolution = candidate
			break
		}
	}
	if sourceResolution == nil {
		return errors.New("invocation map is missing a released-tool source identity")
	}
	if err := VerifyGitReleaseTag(sourceResolution.Root, repository, release, tagObject, peeledCommit); err != nil {
		return errors.New("released-tool Git identity does not match")
	}
	for _, pair := range [][2]string{
		{tooling["executable_asset"].(string), tooling["platform_artifact_sha256"].(string)},
		{tooling["checksums_asset"].(string), tooling["checksums_sha256"].(string)},
		{tooling["release_manifest_asset"].(string), tooling["release_manifest_sha256"].(string)},
	} {
		if _, err := ResolveReleaseAsset(releaseResolution.Root, pair[0], pair[1], maximumResolvedReleaseAssetBytes); err != nil {
			return errors.New("released-tool asset does not match")
		}
	}
	return nil
}
