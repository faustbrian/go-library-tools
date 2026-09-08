//go:build !darwin && !linux

package cohesion

import "errors"

func validateResolutionMetadata(string, bool) error {
	return errors.New("resolution metadata verification is unsupported on this platform")
}
