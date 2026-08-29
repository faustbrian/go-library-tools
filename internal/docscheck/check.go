// Package docscheck validates repository-owned Markdown navigation.
package docscheck

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

const (
	maximumDocumentSize = 4 << 20
	maximumDocuments    = 4096
)

// Check validates root documentation and local Markdown links without network
// access or following symlinks.
func Check(root string) error {
	if !filepath.IsAbs(root) {
		return errors.New("documentation root must be absolute")
	}
	canonicalRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return fmt.Errorf("resolve documentation root: %w", err)
	}
	root = canonicalRoot
	paths, err := documents(root)
	if err != nil {
		return err
	}
	for _, path := range paths {
		if err := checkDocument(root, path); err != nil {
			return err
		}
	}
	return nil
}

func documents(root string) ([]string, error) {
	paths := make([]string, 0, 32)
	readme := filepath.Join(root, "README.md")
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == root {
			return nil
		}
		relative := strings.TrimPrefix(path, root+string(filepath.Separator))
		inDocs := relative == "docs" || strings.HasPrefix(relative, "docs"+string(filepath.Separator))
		if entry.Type()&os.ModeSymlink != 0 {
			if inDocs || filepath.Ext(path) == ".md" {
				return fmt.Errorf("documentation symlink is not allowed: %s", path)
			}
			return nil
		}
		if entry.IsDir() && !inDocs {
			if path == readme {
				return errors.New("README.md must be a regular file")
			}
			return filepath.SkipDir
		}
		if !entry.IsDir() && filepath.Ext(path) == ".md" {
			paths = append(paths, path)
			if len(paths) > maximumDocuments {
				return errors.New("documentation file count exceeds limit")
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk documentation: %w", err)
	}
	if !slices.Contains(paths, readme) {
		return nil, errors.New("README.md is required")
	}
	return paths, nil
}

func checkDocument(root, document string) error {
	data, err := os.ReadFile(document)
	if err != nil {
		return fmt.Errorf("read documentation %s: %w", document, err)
	}
	if len(data) > maximumDocumentSize {
		return fmt.Errorf("documentation %s exceeds size limit", document)
	}
	for index, text := range strings.Split(string(data), "\n") {
		line := index + 1
		if strings.HasSuffix(text, " ") || strings.HasSuffix(text, "\t") {
			return fmt.Errorf("documentation %s:%d has trailing whitespace", document, line)
		}
	}
	parsed := goldmark.DefaultParser().Parse(text.NewReader(data))
	return ast.Walk(parsed, func(node ast.Node, _ bool) (ast.WalkStatus, error) {
		var destination []byte
		switch value := node.(type) {
		case *ast.Link:
			destination = value.Destination
		case *ast.Image:
			destination = value.Destination
		default:
			return ast.WalkContinue, nil
		}
		if err := checkLink(root, document, string(destination)); err != nil {
			position := min(max(node.Pos(), 0), len(data))
			line := bytes.Count(data[:position], []byte{'\n'}) + 1
			return ast.WalkStop, fmt.Errorf("documentation %s:%d: %w", document, line, err)
		}
		return ast.WalkContinue, nil
	})
}

func checkLink(root, document, target string) error {
	target = strings.TrimSpace(target)
	parsed, err := url.Parse(target)
	if err != nil {
		return fmt.Errorf("invalid link %q: %w", target, err)
	}
	if parsed.Scheme != "" {
		scheme := strings.ToLower(parsed.Scheme)
		if scheme == "https" || scheme == "http" || scheme == "mailto" {
			return nil
		}
		return fmt.Errorf("unsupported link scheme %q", parsed.Scheme)
	}
	if parsed.Host != "" || filepath.IsAbs(parsed.Path) {
		return fmt.Errorf("local link must be repository-relative: %q", target)
	}
	if parsed.Path == "" {
		return nil
	}
	resolved := filepath.Clean(filepath.Join(filepath.Dir(document), filepath.FromSlash(parsed.Path)))
	relative, err := filepath.Rel(root, resolved)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return fmt.Errorf("local link escapes repository: %q", target)
	}
	info, err := os.Lstat(resolved)
	if err != nil {
		return fmt.Errorf("broken local link %q: %w", target, err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("local link targets symlink %q", target)
	}
	return nil
}
