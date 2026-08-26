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

	"github.com/faustbrian/go-library-tools/internal/gates"
)

func TestExecuteReportsExecutorCreationAndCleanupFailures(t *testing.T) {
	root := internalFixture(t)
	failure := errors.New("injected failure")
	tests := []struct {
		name    string
		factory executorFactory
	}{
		{"creation", func(string, io.Writer, io.Writer) (gates.Executor, func() error, error) {
			return nil, nil, failure
		}},
		{"cleanup", func(string, io.Writer, io.Writer) (gates.Executor, func() error, error) {
			return successfulExecutor{}, func() error { return failure }, nil
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := execute([]string{"check"}, root, &stdout, &stderr, test.factory)
			if code != 1 || !strings.Contains(stderr.String(), failure.Error()) {
				t.Fatalf("execute() code = %d, stderr = %q", code, stderr.String())
			}
		})
	}
}

type successfulExecutor struct{}

func (successfulExecutor) Run(context.Context, gates.Command) error { return nil }

func internalFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	files := map[string]string{
		".golib.yaml":   "schema_version: 1\ntool_version: v1.0.0\n",
		"go.mod":        "module example\n\ngo 1.27.0\n",
		"example.go":    "package example\n",
		"modules.json":  `{"schema_version":1,"repository":"example","go_version":"1.27.0","modules":[{"directory":".","module_path":"example","go_version":"1.27.0","kind":"public","releasable":true,"gates":{},"packages":[]}]}`,
		"packages.json": `{"schema_version":1,"repository":"example","packages":[]}`,
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(root, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return root
}
