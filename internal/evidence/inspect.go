package evidence

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type inspectFileSystem interface {
	Lstat(string) (os.FileInfo, error)
	WalkDir(string, fs.WalkDirFunc) error
	Rel(string, string) (string, error)
	Open(string) (io.ReadCloser, error)
}

type operatingInspectFileSystem struct{}

func (operatingInspectFileSystem) Lstat(path string) (os.FileInfo, error) { return os.Lstat(path) }
func (operatingInspectFileSystem) WalkDir(root string, visit fs.WalkDirFunc) error {
	return filepath.WalkDir(root, visit)
}
func (operatingInspectFileSystem) Rel(base, target string) (string, error) {
	return filepath.Rel(base, target)
}
func (operatingInspectFileSystem) Open(path string) (io.ReadCloser, error) { return os.Open(path) }

// Inspect validates every content-addressed record under an evidence root.
func Inspect(root, evidenceRoot, repository string, modules []string) ([]Record, error) {
	return inspect(operatingInspectFileSystem{}, root, evidenceRoot, repository, modules)
}

func inspect(files inspectFileSystem, root, evidenceRoot, repository string, modules []string) ([]Record, error) {
	directory := filepath.Join(root, evidenceRoot)
	info, err := files.Lstat(directory)
	if errors.Is(err, os.ErrNotExist) {
		return []Record{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("inspect evidence root: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%w: evidence root is not a real directory", ErrInvalid)
	}
	knownModules := make(map[string]struct{}, len(modules))
	for _, module := range modules {
		knownModules[module] = struct{}{}
	}
	records := make([]Record, 0)
	identities := make(map[string]string)
	err = files.WalkDir(directory, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("%w: symlink in evidence tree: %s", ErrInvalid, path)
		}
		if entry.IsDir() || filepath.Ext(path) != ".json" {
			return nil
		}
		relative, err := files.Rel(directory, path)
		if err != nil {
			return err
		}
		parts := strings.Split(filepath.ToSlash(relative), "/")
		byInput := slices.Index(parts, "by-input")
		if byInput < 0 {
			return nil
		}
		if len(parts) != byInput+3 {
			return fmt.Errorf("%w: malformed evidence path: %s", ErrInvalid, relative)
		}
		file, err := files.Open(path)
		if err != nil {
			return err
		}
		record, parseErr := Parse(file)
		closeErr := file.Close()
		if parseErr != nil {
			return fmt.Errorf("%s: %w", relative, parseErr)
		}
		if closeErr != nil {
			return fmt.Errorf("close evidence %s: %w", relative, closeErr)
		}
		expectedName := strings.TrimPrefix(record.InputDigest, "sha256:") + ".json"
		if parts[byInput+1] != record.Gate || parts[byInput+2] != expectedName {
			return fmt.Errorf("%w: evidence path does not match record: %s", ErrInvalid, relative)
		}
		if record.Repository != repository {
			return fmt.Errorf("%w: evidence repository mismatch: %s", ErrInvalid, relative)
		}
		if _, exists := knownModules[record.Module]; !exists {
			return fmt.Errorf("%w: evidence references unknown module %q", ErrInvalid, record.Module)
		}
		identity := record.Module + "\x00" + record.Package + "\x00" + record.Gate + "\x00" + record.InputDigest
		if previous, exists := identities[identity]; exists {
			return fmt.Errorf("%w: duplicate evidence identity in %s and %s", ErrInvalid, previous, relative)
		}
		identities[identity] = relative
		records = append(records, record)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("inspect evidence: %w", err)
	}
	slices.SortFunc(records, func(left, right Record) int {
		leftKey := left.Module + "\x00" + left.Package + "\x00" + left.Gate + "\x00" + left.InputDigest
		rightKey := right.Module + "\x00" + right.Package + "\x00" + right.Gate + "\x00" + right.InputDigest
		return strings.Compare(leftKey, rightKey)
	})
	return records, nil
}
