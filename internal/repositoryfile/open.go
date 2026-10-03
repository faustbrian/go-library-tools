package repositoryfile

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
)

// OpenContext admits a bounded regular input while preserving os.Stat/Open path
// and symlink semantics. The caller must retain its streaming byte bound.
// Stat/open is not atomic: callers must supply a trusted, stable filesystem;
// cancellation checkpoints cannot interrupt an in-flight filesystem operation.
func OpenContext(ctx context.Context, path string, maximum int64) (io.ReadCloser, error) {
	return openContext(ctx, path, maximum, os.Stat, operatingSystem{}.Open)
}

// OpenRootContext retains the supplied root's confinement and symlink semantics.
// It has the same stable-filesystem/cooperative-IO requirements as OpenContext.
func OpenRootContext(ctx context.Context, root *os.Root, path string, maximum int64) (io.ReadCloser, error) {
	return openContext(ctx, path, maximum, root.Stat, func(name string) (file, error) { return root.Open(name) })
}

func openContext(ctx context.Context, path string, maximum int64, stat func(string) (os.FileInfo, error), open func(string) (file, error)) (file, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if maximum <= 0 {
		return nil, ErrTooLarge
	}
	expected, err := stat(path)
	if err != nil {
		return nil, fmt.Errorf("inspect %s: %w", path, err)
	}
	return openAdmitted(ctx, path, maximum, expected, open)
}

func openAdmitted(ctx context.Context, path string, maximum int64, expected os.FileInfo, open func(string) (file, error)) (file, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if expected == nil {
		return nil, fmt.Errorf("inspect %s: %w", path, ErrUnsafePath)
	}
	if !expected.Mode().IsRegular() {
		return nil, fmt.Errorf("%w: %s", ErrNotRegular, path)
	}
	if expected.Size() > maximum {
		return nil, fmt.Errorf("%w: %s", ErrTooLarge, path)
	}
	openedFile, err := open(path)
	if err != nil {
		if openedFile != nil {
			_ = openedFile.Close()
		}
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	if openedFile == nil {
		return nil, fmt.Errorf("%w: open %s returned no file", ErrUnsafePath, path)
	}
	success := false
	defer func() {
		if !success {
			_ = openedFile.Close()
		}
	}()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	opened, err := openedFile.Stat()
	if err != nil {
		return nil, fmt.Errorf("inspect open %s: %w", path, err)
	}
	if opened == nil {
		return nil, fmt.Errorf("%w: inspect open %s returned no metadata", ErrUnsafePath, path)
	}
	if !opened.Mode().IsRegular() || !os.SameFile(expected, opened) {
		return nil, fmt.Errorf("%w: %s changed while opening", ErrUnsafePath, path)
	}
	if opened.Size() > maximum {
		return nil, fmt.Errorf("%w: %s", ErrTooLarge, path)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	success = true
	return &contextFile{file: openedFile, ctx: ctx}, nil
}

type contextFile struct {
	file
	ctx context.Context
}

func (f *contextFile) Read(buffer []byte) (int, error) {
	if err := f.ctx.Err(); err != nil {
		return 0, err
	}
	n, err := f.file.Read(buffer)
	if err == nil || errors.Is(err, io.EOF) {
		if stopped := f.ctx.Err(); stopped != nil {
			return 0, stopped
		}
	}
	return n, err
}
