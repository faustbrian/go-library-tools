package mutation

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type reportFileSystem interface {
	Mkdir(string, os.FileMode) error
	Lstat(string) (os.FileInfo, error)
	CreateTemp(string, string) (durableReportFile, error)
	Link(string, string) error
	ReadFile(string) ([]byte, error)
	Remove(string) error
}

type durableReportFile interface {
	io.Writer
	Sync() error
	Close() error
	Name() string
}

type operatingReportFiles struct{}

func (operatingReportFiles) Mkdir(path string, mode os.FileMode) error {
	return os.Mkdir(path, mode)
}
func (operatingReportFiles) Lstat(path string) (os.FileInfo, error) { return os.Lstat(path) }
func (operatingReportFiles) CreateTemp(directory, pattern string) (durableReportFile, error) {
	return os.CreateTemp(directory, pattern)
}
func (operatingReportFiles) Link(oldPath, newPath string) error   { return os.Link(oldPath, newPath) }
func (operatingReportFiles) ReadFile(path string) ([]byte, error) { return os.ReadFile(path) }
func (operatingReportFiles) Remove(path string) error             { return os.Remove(path) }

// StoreReport validates and atomically publishes one immutable report by input.
// The bool reports whether byte-identical report content already existed.
func StoreReport(root, inputDigest string, report []byte) (string, bool, ReportResult, error) {
	return storeReport(operatingReportFiles{}, root, inputDigest, report)
}

func storeReport(files reportFileSystem, root, inputDigest string, report []byte) (string, bool, ReportResult, error) {
	if !filepath.IsAbs(root) || !strings.HasPrefix(inputDigest, "sha256:") || !digestRE.MatchString(strings.TrimPrefix(inputDigest, "sha256:")) {
		return "", false, ReportResult{}, fmt.Errorf("%w: report root or input digest is malformed", ErrInvalid)
	}
	result, err := ValidateReport(bytes.NewReader(report))
	if err != nil {
		return "", false, ReportResult{}, err
	}
	directory := filepath.Join(root, "reports")
	if err := prepareReportDirectories(files, root, directory); err != nil {
		return "", false, ReportResult{}, err
	}
	destination := filepath.Join(directory, strings.TrimPrefix(inputDigest, "sha256:")+".json")
	temporary, err := files.CreateTemp(directory, ".report-*")
	if err != nil {
		return "", false, ReportResult{}, fmt.Errorf("create temporary mutation report: %w", err)
	}
	temporaryPath := temporary.Name()
	defer files.Remove(temporaryPath)
	if _, err = temporary.Write(report); err == nil {
		err = temporary.Sync()
	}
	if closeErr := temporary.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return "", false, ReportResult{}, fmt.Errorf("write mutation report: %w", err)
	}
	if err := files.Link(temporaryPath, destination); err == nil {
		return destination, false, result, nil
	} else if !errors.Is(err, os.ErrExist) {
		return "", false, ReportResult{}, fmt.Errorf("publish mutation report: %w", err)
	}
	existing, err := files.ReadFile(destination)
	if err != nil {
		return "", false, ReportResult{}, fmt.Errorf("read existing mutation report: %w", err)
	}
	existingResult, err := ValidateReport(bytes.NewReader(existing))
	if err != nil || existingResult.Digest != result.Digest {
		return "", false, ReportResult{}, fmt.Errorf("%w: mutation report already exists with different content", ErrInvalid)
	}
	return destination, true, existingResult, nil
}

func prepareReportDirectories(files reportFileSystem, root, destination string) error {
	for _, directory := range []string{root, destination} {
		if err := files.Mkdir(directory, 0o700); err != nil && !errors.Is(err, os.ErrExist) {
			return fmt.Errorf("prepare mutation report directory: %w", err)
		}
		info, err := files.Lstat(directory)
		if err != nil {
			return fmt.Errorf("inspect mutation report directory: %w", err)
		}
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("%w: mutation report path is not a real directory", ErrInvalid)
		}
	}
	return nil
}

// LoadReport returns and validates the immutable report for one input digest.
func LoadReport(root, inputDigest string) ([]byte, ReportResult, error) {
	if !filepath.IsAbs(root) || !strings.HasPrefix(inputDigest, "sha256:") || !digestRE.MatchString(strings.TrimPrefix(inputDigest, "sha256:")) {
		return nil, ReportResult{}, fmt.Errorf("%w: report root or input digest is malformed", ErrInvalid)
	}
	path := filepath.Join(root, "reports", strings.TrimPrefix(inputDigest, "sha256:")+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, ReportResult{}, fmt.Errorf("read mutation report: %w", err)
	}
	result, err := ValidateReport(bytes.NewReader(data))
	if err != nil {
		return nil, ReportResult{}, err
	}
	return data, result, nil
}
