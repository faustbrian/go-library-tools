package cli

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/faustbrian/go-library-tools/v2/internal/gates"
)

func TestDocumentationNavigationExercisesRealMarkdownWithoutExecutors(t *testing.T) {
	for _, test := range []struct {
		name   string
		readme string
		guide  string
		code   int
	}{
		{"reachable fragment", "# Project\n\n[Guide](docs/guide.md#run-a-reference)\n", "# Guide\n\n[Parent](../README.md)\n\n## Run a reference\n", 0},
		{"missing fragment", "# Project\n\n[Guide](docs/guide.md#missing-private-anchor)\n", "# Guide\n\n[Parent](../README.md)\n", 1},
		{"unreachable page", "# Project\n", "# Guide\n\n[Parent](../README.md)\n", 1},
		{"machine path", "# Project\n\n[Guide](docs/guide.md)\n", "# Guide\n\nPrivate checkout: /Users/private-person/application\n\n[Parent](../README.md)\n", 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := internalFixture(t)
			if err := os.MkdirAll(filepath.Join(root, "docs"), 0o700); err != nil {
				t.Fatal(err)
			}
			for path, body := range map[string]string{"README.md": test.readme, "docs/guide.md": test.guide} {
				if err := os.WriteFile(filepath.Join(root, path), []byte(body), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			factory := func(string, io.Writer, io.Writer) (gates.Executor, func() error, error) {
				t.Fatal("navigation must not acquire a process executor")
				return nil, nil, nil
			}
			var stdout, stderr bytes.Buffer
			code := execute([]string{"docs", "check", "--navigation"}, root, &stdout, &stderr, factory)
			if code != test.code {
				t.Fatalf("exit = %d, want %d; stdout=%q stderr=%q", code, test.code, stdout.String(), stderr.String())
			}
			if test.code == 0 && stdout.String() != "documentation navigation valid\n" {
				t.Fatalf("success output = %q", stdout.String())
			}
			if strings.Contains(stderr.String(), "private-person") || strings.Contains(stderr.String(), "missing-private-anchor") {
				t.Fatal("untrusted document content escaped through CLI diagnostics")
			}
		})
	}
}

func TestDocumentationNavigationPreservesLocalAndUsageBoundaries(t *testing.T) {
	root := internalFixture(t)
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("# Project\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "docs"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "docs", "orphan.md"), []byte("# Orphan\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"docs", "check", "--navigation", "--module", "."}, {"docs", "check", "--navigation", "--local"}} {
		if code := execute(args, root, &bytes.Buffer{}, &bytes.Buffer{}, nil); code != 2 {
			t.Fatalf("mixed args %v = %d", args, code)
		}
	}
	if code := execute([]string{"docs", "check", "--local"}, root, &bytes.Buffer{}, &bytes.Buffer{}, nil); code != 0 {
		t.Fatalf("existing local behavior changed: exit %d", code)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if code := executeContext(ctx, []string{"docs", "check", "--navigation"}, root, &bytes.Buffer{}, &bytes.Buffer{}, nil); code != 1 {
		t.Fatalf("cancelled navigation = %d", code)
	}
}

func TestDocumentationNavigationDoesNotInventFixtureDocumentationRequirements(t *testing.T) {
	root := internalFixture(t)
	for path, body := range map[string]string{
		"README.md":                  "# Root\n",
		"integration/fixture/go.mod": "module example/fixture\n\ngo 1.27.0\n",
		"modules.json":               `{"schema_version":1,"repository":"example","go_version":"1.27.0","modules":[{"directory":".","module_path":"example","go_version":"1.27.0","kind":"public","releasable":true,"gates":{},"packages":[]},{"directory":"integration/fixture","module_path":"example/fixture","go_version":"1.27.0","kind":"integration fixture","releasable":false,"gates":{},"packages":[]}]}`,
	} {
		if err := os.MkdirAll(filepath.Dir(filepath.Join(root, path)), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, path), []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	var stdout, stderr bytes.Buffer
	if code := execute([]string{"docs", "check", "--navigation"}, root, &stdout, &stderr, nil); code != 0 {
		t.Fatalf("undocumented non-releasable fixture should not require README: %d, %s", code, stderr.String())
	}
}
