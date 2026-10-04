package docscheck

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/faustbrian/go-library-tools/v2/internal/repositoryfile"
)

type documentLimits struct {
	bytes     int64
	documents int
	entries   int
	depth     int
}

func defaultDocumentLimits() documentLimits {
	return documentLimits{bytes: maximumDocumentSize, documents: maximumDocuments, entries: 100_000, depth: 128}
}

type documentSource struct {
	inspect func(string, string) (os.FileInfo, error)
	open    func(context.Context, string, int64) (io.ReadCloser, error)
}

func ordinaryDocumentSource() documentSource {
	return documentSource{
		inspect: func(root, document string) (os.FileInfo, error) {
			relative, err := filepath.Rel(root, document)
			if err != nil {
				return nil, err
			}
			return repositoryfile.InspectRegularFile(root, relative)
		},
		open: repositoryfile.OpenContext,
	}
}

// The FS is rooted at the canonical documentation tree. Its caller supplies a
// stable, trusted filesystem; type checks do not promise an atomic snapshot.
func documentsFS(ctx context.Context, root string, files fs.FS, limits documentLimits) ([]string, error) {
	paths := make([]string, 0, min(32, limits.documents))
	entries := 0
	var walk func(string, int) error
	walk = func(directory string, depth int) (result error) {
		if err := ctx.Err(); err != nil {
			return err
		}
		if depth > limits.depth {
			return errors.New("documentation depth limit exceeded")
		}
		opened, err := files.Open(directory)
		if err != nil {
			return err
		}
		defer func() { result = errors.Join(result, opened.Close()) }()
		if err := ctx.Err(); err != nil {
			return err
		}
		info, err := opened.Stat()
		if err != nil {
			return err
		}
		if info == nil || !info.IsDir() {
			return repositoryfile.ErrNotRegular
		}
		reader, ok := opened.(fs.ReadDirFile)
		if !ok {
			return repositoryfile.ErrNotRegular
		}
		for {
			if err := ctx.Err(); err != nil {
				return err
			}
			// One lookahead entry distinguishes exact allowance from excess.
			// Compare before adding, avoiding overflow at an integer ceiling.
			request := 128
			if remaining := limits.entries - entries; remaining < request {
				request = remaining + 1
			}
			batch, readErr := reader.ReadDir(request)
			if err := ctx.Err(); err != nil {
				return err
			}
			if readErr != nil && !errors.Is(readErr, io.EOF) {
				return readErr
			}
			if len(batch) > request || (len(batch) == 0 && readErr == nil) {
				return errors.New("documentation directory enumeration violated its bound")
			}
			for _, entry := range batch {
				if err := ctx.Err(); err != nil {
					return err
				}
				if entries >= limits.entries {
					return errors.New("documentation entry limit exceeded")
				}
				entries++
				relative := path.Join(directory, entry.Name())
				name := filepath.Join(root, filepath.FromSlash(relative))
				inDocs := relative == "docs" || strings.HasPrefix(relative, "docs/")
				if entry.Type()&os.ModeSymlink != 0 {
					if inDocs || filepath.Ext(name) == ".md" {
						return fmt.Errorf("documentation symlink is not allowed: %s", name)
					}
					continue
				}
				if entry.IsDir() {
					if !inDocs {
						if relative == "README.md" {
							return errors.New("README.md must be a regular file")
						}
						continue
					}
					if err := walk(relative, depth+1); err != nil {
						return err
					}
					continue
				}
				if filepath.Ext(name) != ".md" {
					continue
				}
				info, err := entry.Info()
				if err != nil {
					return err
				}
				if info == nil || !info.Mode().IsRegular() {
					return fmt.Errorf("documentation %s: %w", name, repositoryfile.ErrNotRegular)
				}
				if info.Size() > limits.bytes {
					return fmt.Errorf("documentation %s exceeds size limit: %w", name, repositoryfile.ErrTooLarge)
				}
				if len(paths) >= limits.documents {
					return errors.New("documentation file count exceeds limit")
				}
				paths = append(paths, name)
			}
			if errors.Is(readErr, io.EOF) {
				return nil
			}
		}
	}
	if err := walk(".", 0); err != nil {
		return nil, fmt.Errorf("walk documentation: %w", err)
	}
	if !slices.Contains(paths, filepath.Join(root, "README.md")) {
		return nil, errors.New("README.md is required")
	}
	slices.Sort(paths)
	return paths, nil
}
