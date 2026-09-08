package cohesion

import (
	"encoding/json"
	"errors"
)

// CheckSourcesV2 performs the context-free source-lock validation contract.
// It intentionally performs no filesystem, Git, release, or network
// resolution beyond the explicitly supplied input file.
func CheckSourcesV2(path string) error {
	_, err := ValidateAndNormalizeV3File(sourcesV2SchemaIdentity, path)
	return err
}

// VerifySourcesV2 validates the source lock and invocation map, requires the
// complete locked source closure, and resolves the selected repository's
// manifest from its immutable Git object tree.
func VerifySourcesV2(inputsPath, repository, resolutionMapPath string) error {
	input, err := readResolutionFile(inputsPath, maximumDefaultSchemaV3ArtifactBytes)
	if err != nil {
		return errors.New("resolve source-lock input")
	}
	validated, err := validateAndNormalizeSchemaV3(sourcesV2SchemaIdentity, input)
	if err != nil {
		return err
	}
	resolution, err := ParseResolutionMapV1File(resolutionMapPath)
	if err != nil {
		return errors.New("resolve invocation map")
	}
	root := validated.Value.(map[string]any)
	sources := root["sources"].([]any)
	controls := root["controls"].(map[string]any)
	if err := verifySourceRoster(controls["source_roster"].(map[string]any), sources, resolution); err != nil {
		return err
	}
	var selected map[string]any
	for _, raw := range sources {
		source := raw.(map[string]any)
		if _, err := lockedSourceResolution(source, resolution); err != nil {
			return err
		}
		if source["repository"] == repository {
			selected = source
		}
	}
	if selected == nil {
		return errors.New("requested repository is not present in the source lock")
	}
	if err := verifyLockedSourceManifest(selected, resolution); err != nil {
		return err
	}
	return verifyReleasedToolingIdentity(selected["tooling"].(map[string]any), resolution)
}

func verifySourceRoster(control map[string]any, current []any, resolution ResolutionMapV1) error {
	repository := control["repository"].(string)
	revision := control["source_revision"].(string)
	path := control["path"].(string)
	wantDigest := control["bytes_sha256"].(string)
	var selected *ResolutionSourceV1
	for index := range resolution.Sources {
		candidate := &resolution.Sources[index]
		if candidate.Repository == repository && candidate.SourceRevision == revision {
			selected = candidate
			break
		}
	}
	if selected == nil {
		return errors.New("invocation map is missing the source-roster identity")
	}
	content, err := ResolveGitSourceFile(selected.Root, repository, revision, path, maximumSchemaV3ArtifactBytes)
	if err != nil {
		return errors.New("resolve source-roster bytes")
	}
	if exactBytesSHA256(content) != wantDigest {
		return errors.New("source-roster digest does not match")
	}
	if err := validateSourcesSchema(content); err != nil {
		return errors.New("source-roster schema is invalid")
	}
	var roster struct {
		RepositoryCount int `json:"repository_count"`
		Repositories    []struct {
			Repository string `json:"repository"`
		} `json:"repositories"`
	}
	if err := json.Unmarshal(content, &roster); err != nil || roster.RepositoryCount != len(roster.Repositories) || len(roster.Repositories) != len(current) {
		return errors.New("source-roster membership count does not match")
	}
	previous := ""
	for index, row := range roster.Repositories {
		if (previous != "" && row.Repository <= previous) || row.Repository != current[index].(map[string]any)["repository"] {
			return errors.New("source-roster membership does not match")
		}
		previous = row.Repository
	}
	return nil
}

func verifyLockedSourceManifest(source map[string]any, resolution ResolutionMapV1) error {
	selected, err := lockedSourceResolution(source, resolution)
	if err != nil {
		return err
	}
	repository := source["repository"].(string)
	manifestPath := source["manifest_path"].(string)
	wantDigest := source["manifest_sha256"].(string)
	requests := []gitSourceRequest{{path: manifestPath, maximumBytes: maximumSchemaV3ArtifactBytes}}
	digests := []string{wantDigest}
	if source["source_kind"] == "release-source" {
		for _, raw := range source["schemas"].([]any) {
			schema := raw.(map[string]any)
			requests = append(requests, gitSourceRequest{path: schema["path"].(string), maximumBytes: maximumSchemaV3ArtifactBytes})
			digests = append(digests, schema["bytes_sha256"].(string))
		}
	}
	contents, err := resolveGitSourceFiles(selected.Root, repository, selected.SourceRevision, requests)
	if err != nil {
		return errors.New("resolve locked source manifest")
	}
	for index, content := range contents {
		if exactBytesSHA256(content) != digests[index] {
			if index == 0 {
				return errors.New("locked source manifest digest does not match")
			}
			return errors.New("locked release-source schema digest does not match")
		}
	}
	return nil
}

func lockedSourceResolution(source map[string]any, resolution ResolutionMapV1) (ResolutionSourceV1, error) {
	repository := source["repository"].(string)
	revision, _ := source["source_revision"].(string)
	if source["source_kind"] == "release-source" {
		tooling := source["tooling"].(map[string]any)
		revision = tooling["peeled_commit"].(string)
	}
	for _, candidate := range resolution.Sources {
		if candidate.Repository == repository && candidate.SourceRevision == revision {
			return candidate, nil
		}
	}
	return ResolutionSourceV1{}, errors.New("invocation map is missing a locked source identity")
}
