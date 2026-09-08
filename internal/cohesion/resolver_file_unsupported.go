//go:build !darwin && !linux

package cohesion

import "errors"

func readResolutionFile(string, int64) ([]byte, error) {
	return nil, errors.New("identity-preserving resolution is unsupported on this platform")
}
