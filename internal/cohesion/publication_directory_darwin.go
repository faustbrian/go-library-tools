//go:build darwin

package cohesion

import (
	"errors"
	"os"

	"golang.org/x/sys/unix"
)

func publishDirectoryNoReplace(stage, target string) error {
	if err := unix.RenameatxNp(unix.AT_FDCWD, stage, unix.AT_FDCWD, target, unix.RENAME_EXCL); err != nil {
		if errors.Is(err, unix.EEXIST) {
			return ErrV2TargetExists
		}
		return os.NewSyscallError("renameatx_np", err)
	}
	return nil
}
