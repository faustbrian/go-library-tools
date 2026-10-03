package mutation

import (
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/faustbrian/go-library-tools/v2/internal/repositoryfile"
)

type sourceReadLimits struct {
	entries     int
	file, total int64
}

func mutationSourceLimits() sourceReadLimits {
	return sourceReadLimits{entries: maximumInputEntries, file: maximumInputFile, total: maximumInputTotal}
}

var (
	errSourceEntryLimit = errors.New("mutation source exceeds entry bound")
	errSourceTotalLimit = errors.New("mutation source exceeds byte bound")
)

func readSourceEntries(directory string, maximum int) ([]os.DirEntry, error) {
	if maximum < 0 {
		return nil, errSourceEntryLimit
	}
	file, err := os.Open(directory) // #nosec G304 -- callers admit a task-owned source directory before enumeration
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var entries []os.DirEntry
	for {
		remaining := maximum - len(entries)
		// At most one excess entry is observed, never retained. Small fixed
		// batches avoid whole-directory allocation before admission.
		batch, err := file.ReadDir(min(127, remaining) + 1)
		if err != nil && !errors.Is(err, io.EOF) {
			return nil, err
		}
		if len(batch) > remaining {
			return nil, errSourceEntryLimit
		}
		entries = append(entries, batch...)
		if errors.Is(err, io.EOF) {
			slices.SortFunc(entries, func(left, right os.DirEntry) int { return strings.Compare(left.Name(), right.Name()) })
			return entries, nil
		}
	}
}

func readSourceFile(root, relative string, perFile, remaining int64) ([]byte, error) {
	info, err := repositoryfile.InspectRegularFile(root, relative)
	if err != nil {
		return nil, err
	}
	// Keep historical per-file admission ahead of aggregate admission.
	if info.Size() > perFile {
		return nil, fmt.Errorf("%w: %s", repositoryfile.ErrTooLarge, relative)
	}
	if info.Size() > remaining {
		return nil, errSourceTotalLimit
	}
	// Read also rechecks confinement, regularity and opened-file identity. Its
	// positive limit permits an empty file when the remaining allowance is zero;
	// any bytes observed in that case are still rejected below.
	data, err := repositoryfile.Read(root, relative, max(1, min(perFile, remaining)))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > perFile {
		return nil, fmt.Errorf("%w: %s", repositoryfile.ErrTooLarge, relative)
	}
	if int64(len(data)) > remaining {
		return nil, errSourceTotalLimit
	}
	return data, nil
}
