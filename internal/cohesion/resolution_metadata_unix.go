//go:build darwin || linux

package cohesion

import (
	"errors"

	"golang.org/x/sys/unix"
)

func validateResolutionMetadata(path string, directory bool) error {
	var stat unix.Stat_t
	if err := unix.Lstat(path, &stat); err != nil {
		return errors.New("stat resolution metadata")
	}
	if directory && stat.Mode&unix.S_IFMT != unix.S_IFDIR {
		return errors.New("resolution tree directory identity is invalid")
	}
	if !directory && stat.Mode&unix.S_IFMT != unix.S_IFREG {
		return errors.New("resolution tree file identity is invalid")
	}
	if !directory && stat.Nlink != 1 {
		return errors.New("resolution tree contains a hard-link alias")
	}
	count, err := unix.Llistxattr(path, nil)
	if err != nil && !errors.Is(err, unix.ENOTSUP) {
		return errors.New("inspect resolution extended attributes")
	}
	if count != 0 {
		return errors.New("resolution tree contains extended attributes")
	}
	return nil
}
