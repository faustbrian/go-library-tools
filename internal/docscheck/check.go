// Package docscheck validates repository-owned Markdown navigation.
package docscheck

import (
	"errors"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	maximumDocumentSize = 4 << 20
	maximumDocuments    = 4096
)

var markdownLinkRE = regexp.MustCompile(`\[[^]]+\]\(([^)]*)\)`)

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
	foundReadme := false
	for _, path := range paths {
		if path == readme {
			foundReadme = true
			break
		}
	}
	if !foundReadme {
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
		for _, match := range markdownLinkRE.FindAllStringSubmatch(text, -1) {
			if err := checkLink(root, document, match[1]); err != nil {
				return fmt.Errorf("documentation %s:%d: %w", document, line, err)
			}
		}
	}
	return nil
}

func checkLink(root, document, target string) error {
	destination, err := linkDestination(target)
	if err != nil {
		return err
	}
	target = destination
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

func linkDestination(value string) (string, error) {
	value = strings.TrimSpace(value)
	if !strings.HasPrefix(value, "<") {
		if fields := strings.Fields(value); len(fields) > 0 {
			return fields[0], nil
		}
		return "", nil
	}
	closing := strings.IndexByte(value, '>')
	if closing < 0 {
		return "", errors.New("angle-bracket link destination is not closed")
	}
	return value[1:closing], nil
}
