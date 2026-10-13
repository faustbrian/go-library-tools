package docscheck

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"html"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"unicode"

	"github.com/faustbrian/go-library-tools/v2/internal/repositoryfile"
	"github.com/yuin/goldmark/ast"
	goldmarkhtml "github.com/yuin/goldmark/renderer/html"
)

const maximumNavigationBytes = 32 << 20

var privateMachinePath = regexp.MustCompile(`(?:/Users/|/home/)[^\s/]+/|[A-Za-z]:\\Users\\[^\s\\]+\\`)

type navigationDocument struct {
	anchors map[string]bool
	links   []string
}

// CheckNavigationContext checks local Markdown fragments and consumer-page
// reachability from README.md in the declared module documentation trees.
// Referenced Markdown elsewhere inside the repository joins the same bounded
// graph; unrelated non-documentation fixtures are not invented as entry points.
// It borrows a stable trusted filesystem, rejects below-root symlinks, bounds
// documents and total input, and observes cooperative cancellation around reads
// and parsing. It neither executes examples nor qualifies prose or remote URLs.
// Errors may name document paths; CLI callers must sanitize untrusted details.
func CheckNavigationContext(ctx context.Context, root string, modules []string) error {
	canonical, _, err := rootResolver(filepath.EvalSymlinks).admit(ctx, root, root)
	if err != nil {
		return err
	}
	paths := make(map[string]bool)
	for _, module := range modules {
		if err := repositoryfile.ValidateDirectory(canonical, module); err != nil {
			return err
		}
		tree := filepath.Join(canonical, module)
		found, err := documentsFS(ctx, tree, os.DirFS(tree), defaultDocumentLimits())
		if err != nil {
			return err
		}
		for _, path := range found {
			paths[path] = true
			if len(paths) > maximumDocuments {
				return errors.New("navigation document limit exceeded")
			}
		}
	}
	entry := filepath.Join(canonical, "README.md")
	if !paths[entry] {
		return errors.New("navigation requires README.md")
	}
	documents := make(map[string]navigationDocument, len(paths))
	remaining := int64(maximumNavigationBytes)
	readQueue := make([]string, 0, len(paths))
	for path := range paths {
		readQueue = append(readQueue, path)
	}
	slices.Sort(readQueue)
	for index := 0; index < len(readQueue); index++ {
		path := readQueue[index]
		if err := ctx.Err(); err != nil {
			return err
		}
		relative, err := filepath.Rel(canonical, path)
		if err != nil {
			return err
		}
		data, err := repositoryfile.Read(canonical, relative, min(remaining, maximumDocumentSize))
		if err != nil {
			return err
		}
		remaining -= int64(len(data))
		if err := markdownParser(parseDocument).validate(ctx, canonical, path, data); err != nil {
			return err
		}
		if privateMachinePath.Match(data) {
			return fmt.Errorf("private machine path in documentation %s", relative)
		}
		document, err := parseNavigation(ctx, data)
		if err != nil {
			return err
		}
		documents[path] = document
		for _, link := range document.links {
			target, _, local, err := navigationTarget(ctx, canonical, path, link, documents)
			if err != nil {
				return err
			}
			if local && strings.EqualFold(filepath.Ext(target), ".md") && !paths[target] {
				paths[target] = true
				if len(paths) > maximumDocuments {
					return errors.New("navigation document limit exceeded")
				}
				readQueue = append(readQueue, target)
			}
		}
	}
	edges := make(map[string][]string, len(documents))
	for path, document := range documents {
		for _, link := range document.links {
			if err := ctx.Err(); err != nil {
				return err
			}
			target, fragment, local, err := navigationTarget(ctx, canonical, path, link, documents)
			if err != nil {
				return err
			}
			if !local {
				continue
			}
			if linked, exists := documents[target]; exists {
				if fragment != "" && !linked.anchors[fragment] {
					return fmt.Errorf("broken Markdown fragment in documentation %s", path)
				}
				edges[path] = append(edges[path], target)
			}
		}
	}
	reached := map[string]bool{entry: true}
	pending := []string{entry}
	for len(pending) != 0 {
		if err := ctx.Err(); err != nil {
			return err
		}
		current := pending[len(pending)-1]
		pending = pending[:len(pending)-1]
		for _, next := range edges[current] {
			if !reached[next] {
				reached[next] = true
				pending = append(pending, next)
			}
		}
	}
	for path := range documents {
		relative, err := filepath.Rel(canonical, path)
		if err != nil {
			return err
		}
		if !reached[path] && strings.Contains(string(filepath.Separator)+relative, string(filepath.Separator)+"docs"+string(filepath.Separator)) {
			return fmt.Errorf("unreachable documentation %s", path)
		}
	}
	return ctx.Err()
}

func parseNavigation(ctx context.Context, data []byte) (navigationDocument, error) {
	document := navigationDocument{anchors: make(map[string]bool)}
	nextSuffix := make(map[string]int)
	err := ast.Walk(parseDocument(data), func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if err := ctx.Err(); err != nil {
			return ast.WalkStop, err
		}
		if !entering {
			return ast.WalkContinue, nil
		}
		switch value := node.(type) {
		case *ast.Heading:
			title, err := navigationHeading(ctx, value, data)
			if err != nil {
				return ast.WalkStop, err
			}
			base := navigationSlug(title)
			anchor := base
			count := nextSuffix[base]
			if count != 0 {
				anchor = fmt.Sprintf("%s-%d", base, count)
			}
			for document.anchors[anchor] {
				if err := ctx.Err(); err != nil {
					return ast.WalkStop, err
				}
				count++
				anchor = fmt.Sprintf("%s-%d", base, count)
			}
			nextSuffix[base] = count + 1
			document.anchors[anchor] = true
		case *ast.Link:
			document.links = append(document.links, string(value.Destination))
		}
		return ast.WalkContinue, nil
	})
	return document, err
}

// navigationHeading keeps rendered text, excluding inline HTML markup. Normal
// text uses Goldmark's entity/escape semantics; raw code text remains literal.
func navigationHeading(ctx context.Context, heading *ast.Heading, data []byte) (string, error) {
	var title strings.Builder
	var escaped strings.Builder
	buffer := bufio.NewWriterSize(&escaped, 64)
	writer := goldmarkhtml.NewWriter()
	err := ast.Walk(heading, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if err := ctx.Err(); err != nil {
			return ast.WalkStop, err
		}
		if entering {
			switch value := node.(type) {
			case *ast.Text:
				if value.IsRaw() {
					title.Write(value.Value(data))
				} else {
					writer.Write(buffer, value.Value(data))
					if err := buffer.Flush(); err != nil {
						return ast.WalkStop, err
					}
					title.WriteString(html.UnescapeString(escaped.String()))
					escaped.Reset()
				}
				if value.SoftLineBreak() || value.HardLineBreak() {
					title.WriteByte(' ')
				}
			case *ast.AutoLink:
				title.Write(value.Label(data))
			}
		}
		return ast.WalkContinue, nil
	})
	return title.String(), err
}

func navigationSlug(title string) string {
	return strings.Map(func(value rune) rune {
		switch {
		case unicode.IsSpace(value):
			return '-'
		case unicode.IsLetter(value), unicode.IsNumber(value), unicode.IsMark(value), value == '-', value == '_':
			return unicode.ToLower(value)
		default:
			return -1
		}
	}, title)
}

func navigationTarget(ctx context.Context, root, source, destination string, documents map[string]navigationDocument) (string, string, bool, error) {
	parsed, err := url.Parse(destination)
	if err != nil {
		return "", "", false, err
	}
	if parsed.Scheme != "" || parsed.Host != "" {
		return "", "", false, nil
	}
	if err := checkLinkContext(ctx, root, source, destination); err != nil {
		return "", "", false, err
	}
	target := source
	if parsed.Path != "" {
		target = filepath.Clean(filepath.Join(filepath.Dir(source), filepath.FromSlash(parsed.Path)))
	}
	if _, exists := documents[target]; !exists {
		for _, name := range []string{"README.md", "index.md"} {
			candidate := filepath.Join(target, name)
			relative, err := filepath.Rel(root, candidate)
			if err != nil {
				return "", "", false, err
			}
			if _, err := repositoryfile.InspectRegularFile(root, relative); err == nil {
				target = candidate
				break
			}
		}
	}
	return target, parsed.Fragment, true, nil
}
