package docscheck

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/faustbrian/go-library-tools/v2/internal/repositoryfile"
)

func TestNavigationRenderedHeadingText(t *testing.T) {
	for _, test := range []struct {
		name, heading, valid, invalid string
	}{
		{"named entity", "A &amp; B", "a--b", "a-amp-b"},
		{"numeric entity", "&#x55;ber", "uber", "x55ber"},
		{"inline HTML", "A <em>B</em>", "a-b", "a-embem"},
		{"literal code", "A `&amp;` B", "a-amp-b", "a--b"},
		{"escaped entity", `A \&amp; B`, "a-amp-b", "a--b"},
		{"single entity decoding", "&amp;#65;", "65", "a"},
		{"escaped underscore", `A \_ B`, "a-_-b", "a--b"},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			for _, target := range []string{test.valid, test.invalid} {
				write(t, filepath.Join(root, "README.md"), "# "+test.heading+"\n\n[Self](#"+target+")\n")
				err := CheckNavigationContext(t.Context(), root, []string{"."})
				if target == test.valid && err != nil {
					t.Fatalf("rendered heading rejected: %v", err)
				}
				if target == test.invalid && (err == nil || !strings.Contains(err.Error(), "fragment")) {
					t.Fatalf("non-rendered anchor accepted: %v", err)
				}
			}
		})
	}
}

func TestNavigationRootMetadataDoesNotDependOnCheckoutAncestors(t *testing.T) {
	for _, placement := range []string{"plain", filepath.Join("docs", "repo")} {
		root := filepath.Join(t.TempDir(), placement)
		write(t, filepath.Join(root, "README.md"), "# Project\n")
		write(t, filepath.Join(root, "POLICY.md"), "# Policy\n")
		if err := CheckNavigationContext(t.Context(), root, []string{"."}); err != nil {
			t.Fatalf("root metadata under %s: %v", placement, err)
		}
		write(t, filepath.Join(root, "docs", "orphan.md"), "# Orphan\n")
		if err := CheckNavigationContext(t.Context(), root, []string{"."}); err == nil || !strings.Contains(err.Error(), "unreachable") {
			t.Fatalf("real detailed orphan under %s accepted: %v", placement, err)
		}
	}
}

func TestNavigationRepeatedHeadingsRespectBoundedWork(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "README.md"), strings.Repeat("# Same\n", 32000)+"\n[Last](#same-31999)\n")
	ctx, cancel := context.WithTimeout(t.Context(), 3*time.Second)
	defer cancel()
	if err := CheckNavigationContext(ctx, root, []string{"."}); err != nil {
		t.Fatalf("bounded repeated headings: %v", err)
	}
}

func TestNavigationChecksRenderedAnchorsAndReferenceLinks(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "README.md"), "# Project\n\n[Guide][g]\n\n[g]: docs/guide.md#über-new--a-b\n")
	write(t, filepath.Join(root, "docs", "guide.md"), "# Guide\n\n[Parent](../README.md)\n\n## Über `New()` & A B\n\n## Same\n\n## Same-1\n\n## Same\n\n[Duplicate](#same-2)\n\n## Cafe\u0301\n\n[Combining accent](#cafe\u0301)\n\n```go\n[Not a link](#absent)\n```\n")
	if err := CheckNavigationContext(context.Background(), root, []string{"."}); err != nil {
		t.Fatalf("rendered anchors/reference links: %v", err)
	}
	write(t, filepath.Join(root, "docs", "guide.md"), "# Guide\n\n[Parent](../README.md)\n\n## Über `New()` & A B\n\n[Absent duplicate](#same-1)\n")
	if err := CheckNavigationContext(context.Background(), root, []string{"."}); err == nil || !strings.Contains(err.Error(), "fragment") {
		t.Fatalf("missing actual heading accepted: %v", err)
	}
}

func TestNavigationRequiresAnEntryRatherThanAnIsolatedCycle(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "README.md"), "# Project\n")
	write(t, filepath.Join(root, "docs", "a.md"), "# A\n\n[B](b.md)\n")
	write(t, filepath.Join(root, "docs", "b.md"), "# B\n\n[A](a.md)\n")
	if err := CheckNavigationContext(context.Background(), root, []string{"."}); err == nil || !strings.Contains(err.Error(), "unreachable") {
		t.Fatalf("isolated documentation cycle accepted: %v", err)
	}
	write(t, filepath.Join(root, "README.md"), "# Project\n\n[Start](docs/a.md)\n")
	if err := CheckNavigationContext(context.Background(), root, []string{"."}); err != nil {
		t.Fatal(err)
	}
}

func TestNavigationIncludesDeclaredNestedModuleDocuments(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "README.md"), "# Project\n\n[Adapter](adapters/target/)\n")
	write(t, filepath.Join(root, "adapters", "target", "README.md"), "# Adapter\n\n[Guide](docs/guide.md#install)\n")
	write(t, filepath.Join(root, "adapters", "target", "docs", "guide.md"), "# Guide\n\n## Install\n\n[Parent](../README.md)\n")
	if err := CheckNavigationContext(context.Background(), root, []string{".", "adapters/target"}); err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(root, "adapters", "target", "docs", "extra.md"), "# Missing navigation\n")
	if err := CheckNavigationContext(context.Background(), root, []string{".", "adapters/target"}); err == nil || !strings.Contains(err.Error(), "unreachable") {
		t.Fatalf("nested orphan accepted: %v", err)
	}
}

func TestNavigationFollowsReferencedMarkdownElsewhereInTheRepository(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "README.md"), "# Project\n\n[API](api/README.md#contract)\n")
	write(t, filepath.Join(root, "api", "README.md"), "# API\n\n## Contract\n\n[Details](details.md#limits)\n")
	write(t, filepath.Join(root, "api", "details.md"), "# Details\n\n## Limits\n\n[Root](../README.md)\n")
	if err := CheckNavigationContext(context.Background(), root, []string{"."}); err != nil {
		t.Fatalf("valid referenced contract: %v", err)
	}
	write(t, filepath.Join(root, "api", "details.md"), "# Details\n\n[Root](../README.md)\n")
	if err := CheckNavigationContext(context.Background(), root, []string{"."}); err == nil || !strings.Contains(err.Error(), "fragment") {
		t.Fatalf("referenced missing fragment accepted: %v", err)
	}
}

func TestNavigationRefusesUnsafePathsAndCancelledWork(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "README.md"), "# Project\n")
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(root, "linked")); err != nil {
		t.Fatal(err)
	}
	for _, module := range []string{"../", outside, "linked"} {
		if err := CheckNavigationContext(context.Background(), root, []string{".", module}); err == nil {
			t.Fatalf("unsafe module %q accepted", module)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := CheckNavigationContext(ctx, root, []string{"."}); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled navigation = %v", err)
	}
}

func TestNavigationRefusesPersonalPathsAndAggregateInputExcess(t *testing.T) {
	for _, path := range []string{"/Users/alice/project", "/home/alice/project", `C:\Users\alice\project`} {
		root := t.TempDir()
		write(t, filepath.Join(root, "README.md"), "# Project\n\n"+path+"\n")
		if err := CheckNavigationContext(context.Background(), root, []string{"."}); err == nil || !strings.Contains(err.Error(), "private machine") {
			t.Fatalf("personal path accepted: %v", err)
		}
	}
	root := t.TempDir()
	write(t, filepath.Join(root, "README.md"), "# Project\n")
	for index := range 9 {
		name := filepath.Join(root, "docs", string(rune('a'+index))+".md")
		write(t, name, strings.Repeat("x", maximumDocumentSize))
	}
	if err := CheckNavigationContext(context.Background(), root, []string{"."}); !errors.Is(err, repositoryfile.ErrTooLarge) {
		t.Fatalf("aggregate byte excess accepted: %v", err)
	}
}

func FuzzNavigationMarkdown(f *testing.F) {
	f.Add([]byte("# Start\n\n[Self](#start)\n"))
	f.Add([]byte("# Über `New()`\n\n~~~go\n[ignored](missing)\n~~~\n"))
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 64<<10 {
			t.Skip()
		}
		if _, err := parseNavigation(context.Background(), data); err != nil {
			t.Fatal(err)
		}
	})
}
