package gates

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/faustbrian/go-library-tools/internal/inventory"
)

func TestLocalPropagatesSelectionAndServiceErrors(t *testing.T) {
	runner := Runner{Catalog: inventory.Inventory{Modules: []inventory.Module{{Directory: "."}}}}
	if err := runner.Local(context.Background(), []string{"missing"}); err == nil || !strings.Contains(err.Error(), "unknown module") {
		t.Fatalf("Local(selection) error = %v", err)
	}

	root := localCoverageFixture(t, "# Example\n")
	runner = Runner{
		Root:     root,
		Catalog:  inventory.Inventory{Modules: []inventory.Module{{Directory: ".", RequiredServices: []string{"db"}}}},
		Executor: executorFunction(func(context.Context, Command) error { return nil }),
		startServices: func(context.Context, []string) (serviceLease, error) {
			return &fakeServiceLease{}, nil
		},
	}
	// A malformed source makes the first local safety check fail inside the
	// service-scoped callback, proving Local returns callback errors.
	if err := os.WriteFile(filepath.Join(root, "broken.go"), []byte("package ["), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := runner.Local(context.Background(), []string{"."}); err == nil || !strings.Contains(err.Error(), "parse") {
		t.Fatalf("Local(callback) error = %v", err)
	}
}

func TestCheckModuleLocalPropagatesSafetyDocumentationAndAPIErrors(t *testing.T) {
	tests := []struct {
		name   string
		module inventory.Module
		setup  func(string)
		want   string
	}{
		{name: "safety", module: inventory.Module{Directory: "."}, setup: func(root string) {
			if err := os.WriteFile(filepath.Join(root, "broken.go"), []byte("package ["), 0o600); err != nil {
				t.Fatalf("write malformed source: %v", err)
			}
		}, want: "parse"},
		{name: "documentation", module: inventory.Module{Directory: ".", Gates: map[string]bool{"documentation": true}}, setup: func(root string) {
			if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("[broken](missing.md)\n"), 0o600); err != nil {
				t.Fatalf("write broken documentation: %v", err)
			}
		}, want: "broken local link"},
		{name: "api", module: inventory.Module{Directory: ".", Gates: map[string]bool{"api_compatibility": true}}, setup: func(string) {}, want: "missing API baseline"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := localCoverageFixture(t, "# Example\n")
			test.setup(root)
			runner := Runner{Root: root, Catalog: inventory.Inventory{Modules: []inventory.Module{test.module}}, Executor: executorFunction(func(context.Context, Command) error { return nil })}
			err := runner.Local(context.Background(), []string{"."})
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Local() error = %v, want %q", err, test.want)
			}
		})
	}
}

func localCoverageFixture(t *testing.T, readme string) string {
	t.Helper()
	root := t.TempDir()
	for name, content := range map[string]string{
		"go.mod":     "module example\n\ngo 1.27.0\n",
		"example.go": "package example\n",
		"README.md":  readme,
	} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return root
}
