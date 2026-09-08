//go:build !darwin && !linux

package cohesion

import "errors"

func publishDirectoryNoReplace(string, string) error {
	return errors.New("atomic no-replace directory publication is unsupported on this platform")
}
