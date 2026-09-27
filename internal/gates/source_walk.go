package gates

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// Unlike filepath.WalkDir, this does not allocate a hostile directory's entire
// entry list before its budget can be checked. Root.Open confines every read.
func walkSecuritySource(ctx context.Context, source string, limit int, visit func(string, fs.DirEntry) error) error {
	root, err := os.OpenRoot(source)
	if err != nil {
		return errors.New("source traversal root unavailable")
	}
	defer root.Close()
	entries := 0
	var walk func(string, int) error
	walk = func(relative string, depth int) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		if depth > 128 {
			return errors.New("source traversal depth limit exceeded")
		}
		directory, err := root.Open(relative)
		if err != nil {
			return errors.New("source directory unavailable")
		}
		defer directory.Close()
		for {
			batch, readErr := directory.ReadDir(128)
			if readErr != nil && !errors.Is(readErr, io.EOF) {
				return errors.New("source directory read failed")
			}
			for _, entry := range batch {
				if err := ctx.Err(); err != nil {
					return err
				}
				entries++
				if entries > limit {
					return errors.New("source traversal entry limit exceeded")
				}
				name := filepath.Join(relative, entry.Name())
				if err := visit(name, entry); errors.Is(err, filepath.SkipDir) {
					continue
				} else if err != nil {
					return err
				}
				if entry.IsDir() {
					if err := walk(name, depth+1); err != nil {
						return err
					}
				}
			}
			if errors.Is(readErr, io.EOF) {
				return nil
			}
		}
	}
	return walk(".", 0)
}
