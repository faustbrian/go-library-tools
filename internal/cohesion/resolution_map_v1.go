package cohesion

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

const maximumResolutionMapV1Bytes = 256 << 20

var (
	resolutionMapIdentityPattern  = regexp.MustCompile(`^(?:[a-z0-9]|[a-z0-9][a-z0-9./:_-]{0,126}[a-z0-9])$`)
	resolutionMapGitSHAPattern    = regexp.MustCompile(`^[0-9a-f]{40}$`)
	resolutionMapDigestPattern    = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)
	resolutionMapReleasePattern   = regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+$`)
	resolutionMapGoVersionPattern = regexp.MustCompile(`^go(?:0|[1-9][0-9]*)\.(?:0|[1-9][0-9]*)(?:\.(?:0|[1-9][0-9]*))?$`)
	resolutionMapGOOSPattern      = regexp.MustCompile(`^linux$`)
	resolutionMapGOARCHPattern    = regexp.MustCompile(`^(?:amd64|arm64)$`)
)

// ResolutionMapV1 is caller-local location data. Its roots are resolver
// capabilities and must never be copied into generated artifacts or errors.
type ResolutionMapV1 struct {
	Sources     []ResolutionSourceV1
	Releases    []ResolutionReleaseV1
	Toolchains  []ResolutionToolchainV1
	ModuleProxy ResolutionModuleProxyV1
}

// ResolutionSourceV1 identifies a locked source repository and root.
type ResolutionSourceV1 struct {
	Repository     string
	SourceRevision string
	Root           string
}

// ResolutionReleaseV1 identifies a locked release and its source objects.
type ResolutionReleaseV1 struct {
	Repository   string
	Release      string
	TagObject    string
	PeeledCommit string
	Root         string
}

// ResolutionToolchainV1 identifies a locked Go toolchain artifact set.
type ResolutionToolchainV1 struct {
	Version            string
	GOOS               string
	GOARCH             string
	DistributionSHA256 string
	TreeSHA256         string
	BinarySHA256       string
	Archive            string
	Root               string
}

// ResolutionModuleProxyV1 identifies a locked module proxy tree.
type ResolutionModuleProxyV1 struct {
	TreeSHA256 *string
	Root       *string
}

// ParseResolutionMapV1 validates the closed invocation-only resolution-map
// contract without dereferencing any caller-owned path.
func ParseResolutionMapV1(input []byte) (ResolutionMapV1, error) {
	limits := trustJSONParserLimits{maximumArrayItems: resolutionMapV1ArrayLimit}
	value, err := parseTrustJSONWithBudgetAndLimits(input, maximumResolutionMapV1Bytes, newTrustJSONBudget(maximumResolutionMapV1Bytes), limits)
	if err != nil {
		return ResolutionMapV1{}, fmt.Errorf("parse resolution map: %w", err)
	}
	root, ok := value.(map[string]any)
	if !ok {
		return ResolutionMapV1{}, errors.New("resolution map must be an object")
	}
	if err := requireExactMembers(root, "module_proxy", "releases", "sources", "toolchains"); err != nil {
		return ResolutionMapV1{}, err
	}

	result := ResolutionMapV1{}
	if result.Sources, err = decodeResolutionSources(root["sources"]); err != nil {
		return ResolutionMapV1{}, err
	}
	if result.Releases, err = decodeResolutionReleases(root["releases"]); err != nil {
		return ResolutionMapV1{}, err
	}
	if result.Toolchains, err = decodeResolutionToolchains(root["toolchains"]); err != nil {
		return ResolutionMapV1{}, err
	}
	if result.ModuleProxy, err = decodeResolutionModuleProxy(root["module_proxy"]); err != nil {
		return ResolutionMapV1{}, err
	}
	return result, nil
}

// ParseResolutionMapV1File reads a map through the no-follow resolver before
// parsing its closed invocation-only form.
func ParseResolutionMapV1File(path string) (ResolutionMapV1, error) {
	input, err := readResolutionFile(path, maximumResolutionMapV1Bytes)
	if err != nil {
		return ResolutionMapV1{}, err
	}
	return ParseResolutionMapV1(input)
}

func resolutionMapV1ArrayLimit(path []trustJSONPathStep) int {
	if len(path) != 1 || path[0].kind != trustJSONObjectMember {
		return maximumTrustJSONCollection
	}
	switch path[0].name {
	case "sources", "releases":
		return 16384
	case "toolchains":
		return 256
	default:
		return maximumTrustJSONCollection
	}
}

func decodeResolutionSources(value any) ([]ResolutionSourceV1, error) {
	rows, ok := value.([]any)
	if !ok {
		return nil, errors.New("resolution map sources must be an array")
	}
	result := make([]ResolutionSourceV1, 0, len(rows))
	identities := make(map[string]struct{}, len(rows))
	previous := ""
	for index, raw := range rows {
		row, ok := raw.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("resolution map source %d must be an object", index)
		}
		if err := requireExactMembers(row, "repository", "root", "source_revision"); err != nil {
			return nil, fmt.Errorf("resolution map source %d: %w", index, err)
		}
		repository, err := requiredPatternString(row, "repository", resolutionMapIdentityPattern)
		if err != nil {
			return nil, fmt.Errorf("resolution map source %d: %w", index, err)
		}
		revision, err := requiredPatternString(row, "source_revision", resolutionMapGitSHAPattern)
		if err != nil {
			return nil, fmt.Errorf("resolution map source %d: %w", index, err)
		}
		root, err := requiredAbsolutePath(row, "root")
		if err != nil {
			return nil, fmt.Errorf("resolution map source %d: %w", index, err)
		}
		identity := repository + "\x00" + revision
		if _, exists := identities[identity]; exists {
			return nil, fmt.Errorf("resolution map source %d duplicates an identity", index)
		}
		identities[identity] = struct{}{}
		sortKey := identity + "\x00" + root
		if previous != "" && sortKey < previous {
			return nil, fmt.Errorf("resolution map sources are not canonically ordered at row %d", index)
		}
		previous = sortKey
		result = append(result, ResolutionSourceV1{Repository: repository, SourceRevision: revision, Root: root})
	}
	return result, nil
}

func decodeResolutionReleases(value any) ([]ResolutionReleaseV1, error) {
	rows, ok := value.([]any)
	if !ok {
		return nil, errors.New("resolution map releases must be an array")
	}
	result := make([]ResolutionReleaseV1, 0, len(rows))
	identities := make(map[string]struct{}, len(rows))
	previous := ""
	for index, raw := range rows {
		row, ok := raw.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("resolution map release %d must be an object", index)
		}
		if err := requireExactMembers(row, "peeled_commit", "release", "repository", "root", "tag_object_sha"); err != nil {
			return nil, fmt.Errorf("resolution map release %d: %w", index, err)
		}
		repository, err := requiredPatternString(row, "repository", resolutionMapIdentityPattern)
		if err != nil {
			return nil, fmt.Errorf("resolution map release %d: %w", index, err)
		}
		release, err := requiredPatternString(row, "release", resolutionMapReleasePattern)
		if err != nil {
			return nil, fmt.Errorf("resolution map release %d: %w", index, err)
		}
		tag, err := requiredPatternString(row, "tag_object_sha", resolutionMapGitSHAPattern)
		if err != nil {
			return nil, fmt.Errorf("resolution map release %d: %w", index, err)
		}
		peel, err := requiredPatternString(row, "peeled_commit", resolutionMapGitSHAPattern)
		if err != nil {
			return nil, fmt.Errorf("resolution map release %d: %w", index, err)
		}
		root, err := requiredAbsolutePath(row, "root")
		if err != nil {
			return nil, fmt.Errorf("resolution map release %d: %w", index, err)
		}
		identity := strings.Join([]string{repository, release, tag, peel}, "\x00")
		if _, exists := identities[identity]; exists {
			return nil, fmt.Errorf("resolution map release %d duplicates an identity", index)
		}
		identities[identity] = struct{}{}
		sortKey := identity + "\x00" + root
		if previous != "" && sortKey < previous {
			return nil, fmt.Errorf("resolution map releases are not canonically ordered at row %d", index)
		}
		previous = sortKey
		result = append(result, ResolutionReleaseV1{Repository: repository, Release: release, TagObject: tag, PeeledCommit: peel, Root: root})
	}
	return result, nil
}

func decodeResolutionToolchains(value any) ([]ResolutionToolchainV1, error) {
	rows, ok := value.([]any)
	if !ok {
		return nil, errors.New("resolution map toolchains must be an array")
	}
	result := make([]ResolutionToolchainV1, 0, len(rows))
	identities := make(map[string]struct{}, len(rows))
	previous := ""
	for index, raw := range rows {
		row, ok := raw.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("resolution map toolchain %d must be an object", index)
		}
		if err := requireExactMembers(row, "archive", "binary_sha256", "distribution_sha256", "goarch", "goos", "root", "tree_sha256", "version"); err != nil {
			return nil, fmt.Errorf("resolution map toolchain %d: %w", index, err)
		}
		version, err := requiredPatternString(row, "version", resolutionMapGoVersionPattern)
		if err != nil {
			return nil, fmt.Errorf("resolution map toolchain %d: %w", index, err)
		}
		goos, err := requiredPatternString(row, "goos", resolutionMapGOOSPattern)
		if err != nil {
			return nil, fmt.Errorf("resolution map toolchain %d: %w", index, err)
		}
		goarch, err := requiredPatternString(row, "goarch", resolutionMapGOARCHPattern)
		if err != nil {
			return nil, fmt.Errorf("resolution map toolchain %d: %w", index, err)
		}
		distribution, err := requiredPatternString(row, "distribution_sha256", resolutionMapDigestPattern)
		if err != nil {
			return nil, fmt.Errorf("resolution map toolchain %d: %w", index, err)
		}
		tree, err := requiredPatternString(row, "tree_sha256", resolutionMapDigestPattern)
		if err != nil {
			return nil, fmt.Errorf("resolution map toolchain %d: %w", index, err)
		}
		binary, err := requiredPatternString(row, "binary_sha256", resolutionMapDigestPattern)
		if err != nil {
			return nil, fmt.Errorf("resolution map toolchain %d: %w", index, err)
		}
		archive, err := requiredAbsolutePath(row, "archive")
		if err != nil {
			return nil, fmt.Errorf("resolution map toolchain %d: %w", index, err)
		}
		root, err := requiredAbsolutePath(row, "root")
		if err != nil {
			return nil, fmt.Errorf("resolution map toolchain %d: %w", index, err)
		}
		identity := strings.Join([]string{version, goos, goarch, distribution, tree, binary}, "\x00")
		if _, exists := identities[identity]; exists {
			return nil, fmt.Errorf("resolution map toolchain %d duplicates an identity", index)
		}
		identities[identity] = struct{}{}
		sortKey := identity + "\x00" + root
		if previous != "" && sortKey < previous {
			return nil, fmt.Errorf("resolution map toolchains are not canonically ordered at row %d", index)
		}
		previous = sortKey
		result = append(result, ResolutionToolchainV1{Version: version, GOOS: goos, GOARCH: goarch, DistributionSHA256: distribution, TreeSHA256: tree, BinarySHA256: binary, Archive: archive, Root: root})
	}
	return result, nil
}

func decodeResolutionModuleProxy(value any) (ResolutionModuleProxyV1, error) {
	row, ok := value.(map[string]any)
	if !ok {
		return ResolutionModuleProxyV1{}, errors.New("resolution map module_proxy must be an object")
	}
	if err := requireExactMembers(row, "root", "tree_sha256"); err != nil {
		return ResolutionModuleProxyV1{}, fmt.Errorf("resolution map module_proxy: %w", err)
	}
	if row["root"] == nil && row["tree_sha256"] == nil {
		return ResolutionModuleProxyV1{}, nil
	}
	if row["root"] == nil || row["tree_sha256"] == nil {
		return ResolutionModuleProxyV1{}, errors.New("resolution map module_proxy root and tree_sha256 must both be null or both be strings")
	}
	root, err := requiredAbsolutePath(row, "root")
	if err != nil {
		return ResolutionModuleProxyV1{}, fmt.Errorf("resolution map module_proxy: %w", err)
	}
	digest, err := requiredPatternString(row, "tree_sha256", resolutionMapDigestPattern)
	if err != nil {
		return ResolutionModuleProxyV1{}, fmt.Errorf("resolution map module_proxy: %w", err)
	}
	return ResolutionModuleProxyV1{TreeSHA256: &digest, Root: &root}, nil
}

func requireExactMembers(value map[string]any, names ...string) error {
	if len(value) != len(names) {
		return errors.New("object does not contain its exact required members")
	}
	want := slices.Clone(names)
	slices.Sort(want)
	got := make([]string, 0, len(value))
	for name := range value {
		got = append(got, name)
	}
	slices.Sort(got)
	if !slices.Equal(got, want) {
		return errors.New("object does not contain its exact required members")
	}
	return nil
}

func requiredPatternString(value map[string]any, field string, pattern *regexp.Regexp) (string, error) {
	text, ok := value[field].(string)
	if !ok || !pattern.MatchString(text) {
		return "", fmt.Errorf("%s is invalid", field)
	}
	return text, nil
}

func requiredAbsolutePath(value map[string]any, field string) (string, error) {
	path, ok := value[field].(string)
	if !ok || len(path) == 0 || len(path) > 4096 || !filepath.IsAbs(path) || filepath.Clean(path) != path {
		return "", fmt.Errorf("%s must be a canonical absolute path", field)
	}
	return path, nil
}
