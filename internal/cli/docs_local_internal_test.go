package cli

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/faustbrian/go-library-tools/v2/internal/gates"
)

func TestExecuteLocalDocumentationWithoutRuntimeCapabilities(t *testing.T) {
	for _, scenario := range []struct {
		name, content, diagnostic string
		cancel                    bool
	}{
		{"valid local link", "# Example\n[policy](.golib.yaml)\n", "", false},
		{"missing local file", "# Example\n[missing](missing.md)\n", "local documentation validation failed", false},
		{"network is not required", "# Example\n[upstream](https://unreachable.invalid/docs)\n", "", false},
		{"cancelled", "# Example\n", "context canceled", true},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			root := internalFixture(t)
			if err := os.WriteFile(filepath.Join(root, "README.md"), []byte(scenario.content), 0o600); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			if scenario.cancel {
				cancel()
			}
			factory := func(string, io.Writer, io.Writer) (gates.Executor, func() error, error) {
				t.Error("local documentation created a runtime execution capability")
				return nil, nil, errors.New("unexpected executor")
			}
			var stdout, stderr bytes.Buffer
			code := executeContext(ctx, []string{"docs", "check", "--local"}, root, &stdout, &stderr, factory)
			if scenario.diagnostic == "" {
				if code != 0 || stderr.Len() != 0 || !strings.Contains(stdout.String(), "local documentation valid") {
					t.Fatalf("local docs result = %d, stdout=%q, stderr=%q", code, stdout.String(), stderr.String())
				}
			} else if code != 1 || !strings.Contains(stderr.String(), scenario.diagnostic) || stdout.Len() != 0 {
				t.Fatalf("local docs failure = %d, stdout=%q, stderr=%q", code, stdout.String(), stderr.String())
			}
		})
	}
}

func TestExecuteLocalDocumentationDoesNotPublishLinkDestinations(t *testing.T) {
	root := internalFixture(t)
	const marker = "private-fixture-marker"
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("[missing](missing.md?private="+marker+")\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	code := executeContext(t.Context(), []string{"docs", "check", "--local"}, root, &stdout, &stderr, nil)
	if code != 1 || stdout.Len() != 0 {
		t.Fatalf("local documentation result = %d, stdout=%q", code, stdout.String())
	}
	if strings.Contains(stderr.String(), marker) || strings.Contains(stderr.String(), "missing.md") {
		t.Fatal("local documentation published a raw link destination")
	}
	if !strings.Contains(stderr.String(), "local documentation validation failed") {
		t.Fatal("local documentation omitted its sanitized failure diagnostic")
	}
}
