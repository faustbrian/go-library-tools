//lint:file-ignore ST1005 Git diagnostic capitalization is part of the established contract.
package cohesion

import (
	"errors"
	"path/filepath"
)

type gitSourceRequest struct {
	path         string
	maximumBytes int64
}

// ResolveGitSourceFile proves the caller-selected checkout is the named clean
// repository revision and reads the blob from that immutable object tree.
func ResolveGitSourceFile(root, repository, revision, path string, maximumBytes int64) ([]byte, error) {
	if !resolutionMapIdentityPattern.MatchString(repository) || !resolutionMapGitSHAPattern.MatchString(revision) || !safeRelativePath(path) || len(path) > 4096 || maximumBytes < 0 {
		return nil, errors.New("Git source identity is invalid")
	}
	if !filepath.IsAbs(root) || filepath.Clean(root) != root {
		return nil, errors.New("Git source root is invalid")
	}
	contents, err := resolveGitSourceFiles(root, repository, revision, []gitSourceRequest{{path: path, maximumBytes: maximumBytes}})
	if err != nil {
		return nil, err
	}
	return contents[0], nil
}

// VerifyGitReleaseTag proves that an exact annotated tag object belongs to the
// named repository and peels to the exact commit without an external source.
func VerifyGitReleaseTag(root, repository, release, tagObject, peeledCommit string) error {
	if !resolutionMapIdentityPattern.MatchString(repository) || !resolutionMapReleasePattern.MatchString(release) ||
		!resolutionMapGitSHAPattern.MatchString(tagObject) || !resolutionMapGitSHAPattern.MatchString(peeledCommit) {
		return errors.New("release Git identity is invalid")
	}
	if !filepath.IsAbs(root) || filepath.Clean(root) != root {
		return errors.New("release Git root is invalid")
	}
	return verifyGitReleaseTag(root, repository, release, tagObject, peeledCommit)
}
