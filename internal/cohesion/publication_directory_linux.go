//go:build linux

package cohesion

import (
	"errors"
	"os"

	"golang.org/x/sys/unix"
)

func publishDirectoryNoReplace(stage, target string) error {
	if err := unix.Renameat2(unix.AT_FDCWD, stage, unix.AT_FDCWD, target, unix.RENAME_NOREPLACE); err != nil {
		if errors.Is(err, unix.EEXIST) {
			return ErrV2TargetExists
		}
		return os.NewSyscallError("renameat2", err)
	}
	return nil
}
