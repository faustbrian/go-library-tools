package docscheck

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// These per-invocation owners preserve cooperative cancellation around real,
// indivisible operations. They do not preempt filesystem calls or parsing.
type rootResolver func(string) (string, error)

func (resolve rootResolver) admit(ctx context.Context, repositoryRoot, documentationRoot string) (string, string, error) {
	if err := ctx.Err(); err != nil {
		return "", "", err
	}
	if !filepath.IsAbs(repositoryRoot) {
		return "", "", errors.New("documentation root must be absolute")
	}
	canonicalRoot, err := resolve(repositoryRoot)
	if err != nil {
		return "", "", fmt.Errorf("resolve documentation root: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return "", "", err
	}
	if !filepath.IsAbs(documentationRoot) {
		return "", "", errors.New("documentation tree must be absolute")
	}
	canonicalTree, err := resolve(documentationRoot)
	if err != nil {
		return "", "", fmt.Errorf("resolve documentation tree: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return "", "", err
	}
	relative, err := filepath.Rel(canonicalRoot, canonicalTree)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", "", errors.New("documentation tree must be inside repository")
	}
	return canonicalRoot, canonicalTree, nil
}

type markdownParser func([]byte) ast.Node

func parseDocument(data []byte) ast.Node {
	return goldmark.DefaultParser().Parse(text.NewReader(data))
}

// validate consumes bytes that the document acquisition owner has admitted.
func (parse markdownParser) validate(ctx context.Context, root, document string, data []byte) error {
	if err := validateDocumentLines(ctx, document, data); err != nil {
		return err
	}
	parsed := parse(data)
	if err := ctx.Err(); err != nil {
		return err
	}
	return validateDocumentAST(ctx, root, document, data, parsed)
}

func validateDocumentLines(ctx context.Context, document string, data []byte) error {
	for index, text := range strings.Split(string(data), "\n") {
		if err := ctx.Err(); err != nil {
			return err
		}
		line := index + 1
		if strings.HasSuffix(text, " ") || strings.HasSuffix(text, "\t") {
			return fmt.Errorf("documentation %s:%d has trailing whitespace", document, line)
		}
	}
	return nil
}

func validateDocumentAST(ctx context.Context, root, document string, data []byte, parsed ast.Node) error {
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

type linkInspector func(string) (os.FileInfo, error)

// validateComponents consumes a relative path admitted by checkLinkContext.
func (inspect linkInspector) validateComponents(ctx context.Context, root, relative, target string) error {
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
		info, statErr := inspect(current)
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
