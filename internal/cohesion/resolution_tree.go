//nolint:forcetypeassert,revive // validated resolution data has a fixed decoded shape
package cohesion

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type ResolutionTreeLimits struct {
	MaximumFiles       int
	MaximumDirectories int
	MaximumFileBytes   int64
	MaximumTotalBytes  int64
}

// DigestResolutionTree returns the semantic digest of the closed tagged-row
// tree preimage used by toolchain and module-proxy resolution.
func DigestResolutionTree(root string, limits ResolutionTreeLimits) (string, error) {
	if !filepath.IsAbs(root) || filepath.Clean(root) != root {
		return "", errors.New("resolution tree root must be canonical and absolute")
	}
	if limits.MaximumFiles < 0 || limits.MaximumDirectories < 0 || limits.MaximumFileBytes < 0 || limits.MaximumTotalBytes < 0 {
		return "", errors.New("resolution tree limits must be nonnegative")
	}
	if err := validateResolutionMetadata(root, true); err != nil {
		return "", err
	}
	rows := make([]map[string]any, 0)
	files, directories := 0, 0
	var totalBytes int64
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return errors.New("walk resolution tree")
		}
		if path == root {
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil || relative == "." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			return errors.New("resolution tree path escapes its root")
		}
		relative = filepath.ToSlash(relative)
		if !safeRelativePath(relative) {
			return errors.New("resolution tree contains an unsafe path")
		}
		info, err := entry.Info()
		if err != nil {
			return errors.New("stat resolution tree entry")
		}
		if info.Mode()&(os.ModeSymlink|os.ModeNamedPipe|os.ModeSocket|os.ModeDevice|os.ModeCharDevice|os.ModeSetuid|os.ModeSetgid|os.ModeSticky|os.ModeIrregular) != 0 {
			return errors.New("resolution tree contains an unsupported entry")
		}
		if entry.IsDir() {
			directories++
			if directories > limits.MaximumDirectories {
				return errors.New("resolution tree exceeds its directory limit")
			}
			if err := validateResolutionMetadata(path, true); err != nil {
				return err
			}
			rows = append(rows, map[string]any{"kind": "directory", "mode": int(info.Mode().Perm()), "path": relative})
			return nil
		}
		if !info.Mode().IsRegular() {
			return errors.New("resolution tree contains a non-regular file")
		}
		files++
		if files > limits.MaximumFiles {
			return errors.New("resolution tree exceeds its file limit")
		}
		if info.Size() < 0 || info.Size() > limits.MaximumFileBytes || totalBytes > limits.MaximumTotalBytes-info.Size() {
			return errors.New("resolution tree exceeds its byte limit")
		}
		if err := validateResolutionMetadata(path, false); err != nil {
			return err
		}
		content, err := readResolutionFile(path, limits.MaximumFileBytes)
		if err != nil {
			return err
		}
		totalBytes += int64(len(content))
		rows = append(rows, map[string]any{"bytes_sha256": exactBytesSHA256(content), "kind": "file", "mode": int(info.Mode().Perm()), "path": relative})
		return nil
	})
	if err != nil {
		return "", err
	}
	slices.SortFunc(rows, func(left, right map[string]any) int {
		return strings.Compare(left["path"].(string), right["path"].(string))
	})
	preimage, err := canonicalMarshal(rows, 512<<20)
	if err != nil {
		return "", fmt.Errorf("canonicalize resolution tree: %w", err)
	}
	return exactBytesSHA256(preimage), nil
}
