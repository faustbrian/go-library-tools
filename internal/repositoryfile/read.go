// Package repositoryfile safely reads bounded repository-owned files.
package repositoryfile

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var (
	// ErrUnsafePath identifies an absolute, escaping, or symlinked path.
	ErrUnsafePath = errors.New("unsafe repository path")
	// ErrNotRegular identifies a file type that is not a regular file.
	ErrNotRegular = errors.New("repository file is not regular")
	// ErrTooLarge identifies input that exceeds its declared bound.
	ErrTooLarge = errors.New("repository file exceeds maximum size")
)

// Read returns at most maximum bytes from a regular file below root. Every
// existing path component must be a real directory or file, never a symlink.
func Read(root, relative string, maximum int64) ([]byte, error) {
	return read(root, relative, maximum, operatingSystem{})
}

// ValidateDirectory verifies that relative identifies a real directory below
// root and that no existing path component is a symlink.
func ValidateDirectory(root, relative string) error {
	_, info, err := inspectPath(root, relative, operatingSystem{})
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%w: %s", ErrNotRegular, relative)
	}
	return nil
}

type fileSystem interface {
	Lstat(string) (os.FileInfo, error)
	Open(string) (file, error)
}

type file interface {
	io.Reader
	Stat() (os.FileInfo, error)
	Close() error
}

type operatingSystem struct{}

func (operatingSystem) Lstat(name string) (os.FileInfo, error) {
	return os.Lstat(name)
}

func (operatingSystem) Open(name string) (file, error) {
	return os.Open(name)
}

func read(root, relative string, maximum int64, files fileSystem) ([]byte, error) {
	if maximum <= 0 {
		return nil, ErrTooLarge
	}
	current, expected, err := inspectPath(root, relative, files)
	if err != nil {
		return nil, err
	}
	if !expected.Mode().IsRegular() {
		return nil, fmt.Errorf("%w: %s", ErrNotRegular, relative)
	}

	file, err := files.Open(current)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", relative, err)
	}
	if file == nil {
		return nil, fmt.Errorf("%w: open %s returned no file", ErrUnsafePath, relative)
	}
	defer file.Close()
	opened, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("inspect open %s: %w", relative, err)
	}
	if opened == nil {
		return nil, fmt.Errorf("%w: inspect open %s returned no metadata", ErrUnsafePath, relative)
	}
	if !opened.Mode().IsRegular() {
		return nil, fmt.Errorf("%w: %s changed while opening", ErrUnsafePath, relative)
	}
	if !os.SameFile(expected, opened) {
		return nil, fmt.Errorf("%w: %s changed while opening", ErrUnsafePath, relative)
	}

	data, err := io.ReadAll(io.LimitReader(file, maximum+1))
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", relative, err)
	}
	if int64(len(data)) > maximum {
		return nil, fmt.Errorf("%w: %s", ErrTooLarge, relative)
	}
	return data, nil
}

func inspectPath(root, relative string, files interface {
	Lstat(string) (os.FileInfo, error)
}) (string, os.FileInfo, error) {
	clean := filepath.Clean(relative)
	if relative == "" || filepath.IsAbs(relative) || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", nil, ErrUnsafePath
	}

	current := root
	parts := strings.Split(clean, string(filepath.Separator))
	for _, part := range parts[:len(parts)-1] {
		current = filepath.Join(current, part)
		info, err := files.Lstat(current)
		if err != nil {
			return "", nil, fmt.Errorf("inspect %s: %w", relative, err)
		}
		if info == nil {
			return "", nil, fmt.Errorf("inspect %s: %w", relative, ErrUnsafePath)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return "", nil, fmt.Errorf("%w: %s", ErrUnsafePath, relative)
		}
		if !info.IsDir() {
			return "", nil, fmt.Errorf("%w: path component %s", ErrNotRegular, part)
		}
	}
	current = filepath.Join(current, parts[len(parts)-1])
	expected, err := files.Lstat(current)
	if err != nil {
		return "", nil, fmt.Errorf("inspect %s: %w", relative, err)
	}
	if expected == nil {
		return "", nil, fmt.Errorf("inspect %s: %w", relative, ErrUnsafePath)
	}
	if expected.Mode()&os.ModeSymlink != 0 {
		return "", nil, fmt.Errorf("%w: %s", ErrUnsafePath, relative)
	}
	return current, expected, nil
}
