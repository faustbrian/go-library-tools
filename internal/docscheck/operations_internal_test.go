package docscheck

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/yuin/goldmark/ast"
)

func TestDocumentationRootResolutionOwnsCancellation(t *testing.T) {
	root := basic(t)
	tree := filepath.Join(root, "docs")
	if err := os.Mkdir(tree, 0o700); err != nil {
		t.Fatal(err)
	}
	wantRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}
	wantTree, err := filepath.EvalSymlinks(tree)
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name        string
		cancelAfter int
	}{
		{name: "ordinary"},
		{name: "repository resolution", cancelAfter: 1},
		{name: "tree resolution", cancelAfter: 2},
	} {
		t.Run(test.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			var resolved []string
			resolve := rootResolver(func(name string) (string, error) {
				canonical, err := filepath.EvalSymlinks(name)
				resolved = append(resolved, name)
				if len(resolved) == test.cancelAfter {
					cancel()
				}
				return canonical, err
			})
			gotRoot, gotTree, err := resolve.admit(ctx, root, tree)
			if test.cancelAfter == 0 {
				if err != nil || gotRoot != wantRoot || gotTree != wantTree || !slices.Equal(resolved, []string{root, tree}) {
					t.Fatal("ordinary root admission changed canonical paths or resolution order")
				}
			} else if !errors.Is(err, context.Canceled) || gotRoot != "" || gotTree != "" || len(resolved) != test.cancelAfter {
				t.Fatal("canceled resolution published roots or dispatched later resolution")
			}
		})
	}
}

func TestDocumentationRootResolutionPreservesOperationErrorOrder(t *testing.T) {
	root := basic(t)
	missing := filepath.Join(root, "missing")
	for _, repositoryMissing := range []bool{true, false} {
		ctx, cancel := context.WithCancel(t.Context())
		calls := 0
		resolve := rootResolver(func(name string) (string, error) {
			canonical, err := filepath.EvalSymlinks(name)
			calls++
			if err != nil {
				cancel()
			}
			return canonical, err
		})
		repository, tree, wantCalls := root, missing, 2
		if repositoryMissing {
			repository, tree, wantCalls = missing, root, 1
		}
		gotRoot, gotTree, err := resolve.admit(ctx, repository, tree)
		cancel()
		if !errors.Is(err, os.ErrNotExist) || errors.Is(err, context.Canceled) || gotRoot != "" || gotTree != "" || calls != wantCalls {
			t.Fatal("root operation error lost precedence or dispatched later resolution")
		}
	}
}

func TestAdmittedMarkdownPhasesOwnCancellation(t *testing.T) {
	root := basic(t)
	document := filepath.Join(root, "README.md")
	data := []byte("# R\n\n[policy](policy.md)\n")
	write(t, filepath.Join(root, "policy.md"), "# Policy\n")
	t.Run("before line validation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		cancel()
		parses := 0
		parse := markdownParser(func(data []byte) ast.Node {
			parses++
			return parseDocument(data)
		})
		if err := parse.validate(ctx, root, document, data); !errors.Is(err, context.Canceled) || parses != 0 {
			t.Fatal("canceled line validation dispatched parsing or published success")
		}
	})
	t.Run("after real parsing", func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()
		parses := 0
		parse := markdownParser(func(data []byte) ast.Node {
			parsed := parseDocument(data)
			parses++
			cancel()
			return parsed
		})
		if err := parse.validate(ctx, root, document, data); !errors.Is(err, context.Canceled) || parses != 1 {
			t.Fatal("real parser completion lost cancellation refusal")
		}
	})
	t.Run("owned real AST", func(t *testing.T) {
		parsed := parseDocument(data)
		ctx, cancel := context.WithCancel(t.Context())
		cancel()
		if err := validateDocumentAST(ctx, root, document, data, parsed); !errors.Is(err, context.Canceled) {
			t.Fatal("canceled AST validation published success")
		}
	})
	t.Run("ordinary", func(t *testing.T) {
		if err := markdownParser(parseDocument).validate(t.Context(), root, document, data); err != nil {
			t.Fatal("admitted ordinary Markdown phases changed validation")
		}
	})
}

func TestConfinedLinkComponentsOwnCancellation(t *testing.T) {
	root := basic(t)
	relative := filepath.Join("docs", "policy.md")
	write(t, filepath.Join(root, relative), "# Policy\n")
	for _, stage := range []string{"before components", "after real inspection", "ordinary"} {
		t.Run(stage, func(t *testing.T) {
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			if stage == "before components" {
				cancel()
			}
			var names []string
			inspect := linkInspector(func(name string) (os.FileInfo, error) {
				info, err := os.Lstat(name)
				names = append(names, name)
				if stage == "after real inspection" {
					cancel()
				}
				return info, err
			})
			err := inspect.validateComponents(ctx, root, relative, "docs/policy.md")
			switch stage {
			case "before components":
				if !errors.Is(err, context.Canceled) || len(names) != 0 {
					t.Fatal("canceled component admission dispatched metadata acquisition")
				}
			case "after real inspection":
				if !errors.Is(err, context.Canceled) || !slices.Equal(names, []string{filepath.Join(root, "docs")}) {
					t.Fatal("canceled real metadata inspection dispatched a later component")
				}
			case "ordinary":
				if err != nil || !slices.Equal(names, []string{filepath.Join(root, "docs"), filepath.Join(root, relative)}) {
					t.Fatal("ordinary confined component validation changed inspection order")
				}
			}
		})
	}
}

func TestConfinedLinkComponentsPreserveOperationErrorOrder(t *testing.T) {
	root := basic(t)
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	calls := 0
	inspect := linkInspector(func(name string) (os.FileInfo, error) {
		info, err := os.Lstat(name)
		calls++
		cancel()
		return info, err
	})
	err := inspect.validateComponents(ctx, root, filepath.Join("missing", "policy.md"), "missing/policy.md")
	if !errors.Is(err, os.ErrNotExist) || errors.Is(err, context.Canceled) || calls != 1 {
		t.Fatal("link metadata failure lost precedence or acquired later components")
	}
}

// The production caller admits absolute canonical paths. This tests only the
// private adapter's genuine Rel rejection, not public-path reachability.
func TestOrdinaryDocumentSourceRejectsIncomparablePrivatePaths(t *testing.T) {
	root := t.TempDir()
	document := filepath.Join(root, "README.md")
	_, want := filepath.Rel("ordinary", document)
	if want == nil {
		t.Fatal("private path precondition did not exercise a genuine Rel refusal")
	}
	info, err := ordinaryDocumentSource().inspect("ordinary", document)
	if info != nil || err == nil || err.Error() != want.Error() {
		t.Fatal("private source adapter lost Rel refusal before filesystem inspection")
	}
}
