//go:build !darwin && !linux

package cohesion

import "errors"

func resolveGitSourceFiles(string, string, string, []gitSourceRequest) ([][]byte, error) {
	return nil, errors.New("Git object resolution is unsupported on this platform")
}

func verifyGitReleaseTag(string, string, string, string, string) error {
	return errors.New("Git object resolution is unsupported on this platform")
}
