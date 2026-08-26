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

func TestExecuteRoutesAPICheckAndUpdate(t *testing.T) {
	root := internalFixture(t)
	if err := os.Mkdir(filepath.Join(root, "api"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "api", "baseline.txt"), []byte("baseline"), 0o600); err != nil {
		t.Fatal(err)
	}
	manifest := `{"schema_version":1,"repository":"example","go_version":"1.27.0","modules":[{"directory":".","module_path":"example","go_version":"1.27.0","kind":"public","releasable":true,"gates":{"api_compatibility":true},"packages":[]}]}`
	if err := os.WriteFile(filepath.Join(root, "modules.json"), []byte(manifest), 0o600); err != nil {
		t.Fatal(err)
	}
	factory := func(string, io.Writer, io.Writer) (gates.Executor, func() error, error) {
		return apiExecutor{}, func() error { return nil }, nil
	}
	for _, args := range [][]string{{"api", "check"}, {"api", "update", "--module", "."}} {
		var stdout, stderr bytes.Buffer
		if code := execute(args, root, &stdout, &stderr, factory); code != 0 || stderr.Len() != 0 {
			t.Fatalf("execute(%v) = %d, %q", args, code, stderr.String())
		}
	}
}

func TestExecuteRejectsInvalidAPIArguments(t *testing.T) {
	for _, args := range [][]string{{"api"}, {"api", "remove"}, {"api", "check", "--all", "extra"}} {
		var stdout, stderr bytes.Buffer
		if code := execute(args, internalFixture(t), &stdout, &stderr, nil); code != 2 || !strings.Contains(stderr.String(), "usage: golib api") {
			t.Fatalf("execute(%v) = %d, %q", args, code, stderr.String())
		}
	}
}

type successfulExecutor struct{}

func (successfulExecutor) Run(context.Context, gates.Command) error { return nil }

type apiExecutor struct{}

func (apiExecutor) Run(_ context.Context, command gates.Command) error {
	for index, argument := range command.Args {
		if argument == "-w" {
			return os.WriteFile(command.Args[index+1], []byte("snapshot"), 0o600)
		}
	}
	return nil
}

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
