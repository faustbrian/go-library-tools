//go:build darwin || linux

package cohesion

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/unix"
)

func readResolutionFile(path string, maximumBytes int64) ([]byte, error) {
	if maximumBytes < 0 || !filepath.IsAbs(path) || filepath.Clean(path) != path {
		return nil, errors.New("resolution file path must be canonical and absolute")
	}
	components := strings.Split(strings.TrimPrefix(path, string(filepath.Separator)), string(filepath.Separator))
	if len(components) == 0 || components[0] == "" {
		return nil, errors.New("resolution file path has no file component")
	}

	current, err := unix.Open(string(filepath.Separator), unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, errors.New("open resolution root")
	}
	defer unix.Close(current)
	for index, component := range components {
		if component == "" || component == "." || component == ".." {
			return nil, errors.New("resolution file path contains an unsafe component")
		}
		flags := unix.O_RDONLY | unix.O_CLOEXEC | unix.O_NOFOLLOW
		if index != len(components)-1 {
			flags |= unix.O_DIRECTORY
		}
		next, openErr := unix.Openat(current, component, flags, 0)
		if openErr != nil {
			return nil, fmt.Errorf("open resolution component %d", index)
		}
		if index != 0 {
			_ = unix.Close(current)
		}
		current = next
	}

	var stat unix.Stat_t
	if err := unix.Fstat(current, &stat); err != nil {
		return nil, errors.New("stat resolution file")
	}
	if stat.Mode&unix.S_IFMT != unix.S_IFREG {
		return nil, errors.New("resolution path is not a regular file")
	}
	if stat.Size < 0 || stat.Size > maximumBytes {
		return nil, errors.New("resolution file exceeds its maximum byte length")
	}
	file := os.NewFile(uintptr(current), "resolution-file")
	if file == nil {
		return nil, errors.New("adopt resolution file descriptor")
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, maximumBytes+1))
	if err != nil {
		return nil, errors.New("read resolution file")
	}
	if int64(len(data)) > maximumBytes {
		return nil, errors.New("resolution file exceeds its maximum byte length")
	}
	return data, nil
}
