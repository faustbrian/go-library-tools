package mutation

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type sourceFileSystem interface {
	Lstat(string) (os.FileInfo, error)
	ReadDir(string) ([]os.DirEntry, error)
	ReadFile(string) ([]byte, error)
	Rel(string, string) (string, error)
}

type operatingSourceFiles struct{}

func (operatingSourceFiles) Lstat(path string) (os.FileInfo, error)     { return os.Lstat(path) }
func (operatingSourceFiles) ReadDir(path string) ([]os.DirEntry, error) { return os.ReadDir(path) }
func (operatingSourceFiles) ReadFile(path string) ([]byte, error)       { return os.ReadFile(path) }
func (operatingSourceFiles) Rel(base, target string) (string, error) {
	return filepath.Rel(base, target)
}

// SourceDigest reproduces the source-only identity used by zero-mutant
// reviews. It includes direct production Go files and their repository paths.
func SourceDigest(root, moduleDirectory, packageDirectory string) (string, error) {
	return sourceDigest(operatingSourceFiles{}, root, moduleDirectory, packageDirectory)
}

func sourceDigest(files sourceFileSystem, root, moduleDirectory, packageDirectory string) (string, error) {
	if !filepath.IsAbs(root) || !validRelative(moduleDirectory) || !validRelative(packageDirectory) {
		return "", fmt.Errorf("%w: source digest paths are malformed", ErrInvalid)
	}
	directory := filepath.Join(root, filepath.FromSlash(moduleDirectory), filepath.FromSlash(packageDirectory))
	info, err := files.Lstat(directory)
	if err != nil {
		return "", fmt.Errorf("inspect mutation source directory: %w", err)
	}
	if info == nil {
		return "", fmt.Errorf("%w: inspect mutation source directory returned no metadata", ErrInvalid)
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return "", fmt.Errorf("%w: mutation source path is not a real directory", ErrInvalid)
	}
	entries, err := files.ReadDir(directory)
	if err != nil {
		return "", fmt.Errorf("read mutation source directory: %w", err)
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry == nil {
			return "", fmt.Errorf("%w: mutation source directory contains an invalid entry", ErrInvalid)
		}
		name := entry.Name()
		if entry.Type()&os.ModeSymlink != 0 {
			return "", fmt.Errorf("%w: symlink in mutation source: %s", ErrInvalid, name)
		}
		if !entry.Type().IsRegular() || filepath.Ext(name) != ".go" || strings.HasSuffix(name, "_test.go") {
			continue
		}
		names = append(names, name)
	}
	if len(names) == 0 {
		return "", fmt.Errorf("%w: package has no production Go files", ErrInvalid)
	}
	sort.Strings(names)
	manifest := sha256.New()
	for _, name := range names {
		absolute := filepath.Join(directory, name)
		data, err := files.ReadFile(absolute)
		if err != nil {
			return "", fmt.Errorf("read mutation source %s: %w", name, err)
		}
		digest := sha256.Sum256(data)
		relative, err := files.Rel(root, absolute)
		if err != nil {
			return "", fmt.Errorf("resolve mutation source %s: %w", name, err)
		}
		_, _ = fmt.Fprintf(manifest, "%s  %s\n", hex.EncodeToString(digest[:]), filepath.ToSlash(relative))
	}
	return hex.EncodeToString(manifest.Sum(nil)), nil
}
