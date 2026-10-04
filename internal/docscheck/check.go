// Package docscheck validates repository-owned Markdown navigation.
package docscheck

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"

	"github.com/faustbrian/go-library-tools/v2/internal/repositoryfile"
)

const (
	maximumDocumentSize = 4 << 20
	maximumDocuments    = 4096
)

// Check validates root documentation and local Markdown links without network
// access or following symlinks.
func Check(root string) error {
	return CheckContext(context.Background(), root)
}

// CheckContext validates root documentation using the caller's context.
func CheckContext(ctx context.Context, root string) error {
	return CheckWithinContext(ctx, root, root)
}

// CheckWithin validates one documentation tree while allowing its local links
// to target files elsewhere in the containing repository.
func CheckWithin(repositoryRoot, documentationRoot string) error {
	return CheckWithinContext(context.Background(), repositoryRoot, documentationRoot)
}

// CheckWithinContext validates a documentation tree with cooperative cancellation.
// Roots are canonicalized as before. Below-root paths must be stable and trusted;
// filesystem operations and bounded Markdown parsing are not forcibly interrupted.
func CheckWithinContext(ctx context.Context, repositoryRoot, documentationRoot string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if !filepath.IsAbs(repositoryRoot) {
		return errors.New("documentation root must be absolute")
	}
	canonicalRoot, err := filepath.EvalSymlinks(repositoryRoot)
	if err != nil {
		return fmt.Errorf("resolve documentation root: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if !filepath.IsAbs(documentationRoot) {
		return errors.New("documentation tree must be absolute")
	}
	canonicalTree, err := filepath.EvalSymlinks(documentationRoot)
	if err != nil {
		return fmt.Errorf("resolve documentation tree: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	relative, err := filepath.Rel(canonicalRoot, canonicalTree)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return errors.New("documentation tree must be inside repository")
	}
	limits := defaultDocumentLimits()
	paths, err := documentsFS(ctx, canonicalTree, os.DirFS(canonicalTree), limits)
	if err != nil {
		return err
	}
	for _, path := range paths {
		if err := checkDocumentWithSource(ctx, canonicalRoot, path, limits.bytes, ordinaryDocumentSource()); err != nil {
			return err
		}
	}
	return ctx.Err()
}

func checkDocumentWithSource(ctx context.Context, root, document string, maximum int64, source documentSource) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if maximum <= 0 || maximum > maximumDocumentSize {
		return repositoryfile.ErrTooLarge
	}
	info, err := source.inspect(root, document)
	if err != nil {
		return fmt.Errorf("read documentation %s: %w", document, err)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if info == nil {
		return fmt.Errorf("read documentation %s: %w", document, repositoryfile.ErrUnsafePath)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("read documentation %s: %w", document, repositoryfile.ErrNotRegular)
	}
	if info.Size() > maximum {
		return fmt.Errorf("documentation %s exceeds size limit: %w", document, repositoryfile.ErrTooLarge)
	}
	reader, err := source.open(ctx, document, maximum)
	if err != nil {
		if reader != nil {
			err = errors.Join(err, reader.Close())
		}
		return fmt.Errorf("read documentation %s: %w", document, err)
	}
	if reader == nil {
		return fmt.Errorf("read documentation %s: %w", document, repositoryfile.ErrUnsafePath)
	}
	if err := ctx.Err(); err != nil {
		return errors.Join(err, reader.Close())
	}
	data, err := io.ReadAll(io.LimitReader(reader, maximum+1))
	err = errors.Join(err, reader.Close())
	if err != nil {
		return fmt.Errorf("read documentation %s: %w", document, err)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if int64(len(data)) > maximum {
		return fmt.Errorf("documentation %s exceeds size limit: %w", document, repositoryfile.ErrTooLarge)
	}
	for index, text := range strings.Split(string(data), "\n") {
		if err := ctx.Err(); err != nil {
			return err
		}
		line := index + 1
		if strings.HasSuffix(text, " ") || strings.HasSuffix(text, "\t") {
			return fmt.Errorf("documentation %s:%d has trailing whitespace", document, line)
		}
	}
	parsed := goldmark.DefaultParser().Parse(text.NewReader(data))
	if err := ctx.Err(); err != nil {
		return err
	}
	return ast.Walk(parsed, func(node ast.Node, _ bool) (ast.WalkStatus, error) {
		if err := ctx.Err(); err != nil {
			return ast.WalkStop, err
		}
		var destination []byte
		switch value := node.(type) {
		case *ast.Link:
			destination = value.Destination
		case *ast.Image:
			destination = value.Destination
		default:
			return ast.WalkContinue, nil
		}
		if err := checkLinkContext(ctx, root, document, string(destination)); err != nil {
			position := min(max(node.Pos(), 0), len(data))
			line := bytes.Count(data[:position], []byte{'\n'}) + 1
			return ast.WalkStop, fmt.Errorf("documentation %s:%d: %w", document, line, err)
		}
		return ast.WalkContinue, nil
	})
}

func checkLink(root, document, target string) error {
	return checkLinkContext(context.Background(), root, document, target)
}

func checkLinkContext(ctx context.Context, root, document, target string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
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
	current := root
	if relative == "." {
		relative = ""
	}
	for component := range strings.SplitSeq(relative, string(filepath.Separator)) {
		if err := ctx.Err(); err != nil {
			return err
		}
		if component == "" {
			continue
		}
		current = filepath.Join(current, component)
		info, statErr := os.Lstat(current)
		if statErr != nil {
			return fmt.Errorf("broken local link %q: %w", target, statErr)
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("local link targets symlink %q", target)
		}
	}
	return nil
}
