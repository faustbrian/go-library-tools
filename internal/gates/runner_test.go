package gates_test

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/faustbrian/go-library-tools/internal/gates"
	"github.com/faustbrian/go-library-tools/internal/inventory"
)

func TestCheckRunsStandardGatesInDeterministicOrder(t *testing.T) {
	root := fixture(t)
	executor := &recordingExecutor{}
	var output bytes.Buffer
	runner := gates.Runner{Root: root, Catalog: inventory.Inventory{Modules: []inventory.Module{{
		Directory: ".",
		Gates: map[string]bool{
			"lint": true, "tests": true, "race": true,
		},
	}}}, Executor: executor, Output: &output}

	if err := runner.Check(context.Background(), []string{"."}); err != nil {
		t.Fatalf("Check() error = %v", err)
	}

	want := []string{
		"go mod tidy -diff",
		"go vet ./...",
		"go test ./... -count=1 -timeout=20m",
		"go test -race ./... -count=1 -timeout=20m",
	}
	if !reflect.DeepEqual(executor.commands, want) {
		t.Fatalf("commands = %#v, want %#v", executor.commands, want)
	}
	for _, gate := range []string{"format-check", "tidy-check", "safety", "vet", "test", "race"} {
		if !strings.Contains(output.String(), "[.] "+gate+"\n") {
			t.Errorf("output does not include gate %q: %s", gate, output.String())
		}
	}
}

func TestCheckAppliesModuleTestTagsAndSkipsDisabledGates(t *testing.T) {
	root := fixture(t)
	executor := &recordingExecutor{}
	runner := gates.Runner{Root: root, Catalog: inventory.Inventory{Modules: []inventory.Module{{
		Directory: "nested", TestTags: []string{"integration", "postgres"},
		Gates: map[string]bool{"tests": true},
	}}}, Executor: executor}

	if err := runner.Check(context.Background(), []string{"nested"}); err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	want := []string{
		"go mod tidy -diff",
		"go test -tags=integration,postgres ./... -count=1 -timeout=20m",
	}
	if !reflect.DeepEqual(executor.commands, want) {
		t.Fatalf("commands = %#v, want %#v", executor.commands, want)
	}
}

func TestCheckRejectsUnknownSelectionAndStopsAtFailure(t *testing.T) {
	root := fixture(t)
	failure := errors.New("failed")
	executor := &recordingExecutor{failureAt: 1, failure: failure}
	runner := gates.Runner{Root: root, Catalog: inventory.Inventory{Modules: []inventory.Module{{Directory: ".", Gates: map[string]bool{"tests": true}}}}, Executor: executor}

	if err := runner.Check(context.Background(), []string{"missing"}); err == nil || !strings.Contains(err.Error(), "unknown module") {
		t.Fatalf("Check() unknown error = %v", err)
	}
	if err := runner.Check(context.Background(), []string{"."}); !errors.Is(err, failure) {
		t.Fatalf("Check() failure = %v", err)
	}
	if len(executor.commands) != 2 {
		t.Fatalf("executed %d commands after failure", len(executor.commands))
	}
}

func TestCheckStopsAtExternalGateFailures(t *testing.T) {
	tests := []struct {
		name      string
		gates     map[string]bool
		failureAt int
	}{
		{"tidy", map[string]bool{}, 0},
		{"vet", map[string]bool{"lint": true}, 1},
		{"race", map[string]bool{"tests": true, "race": true}, 2},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			failure := errors.New("failed")
			executor := &recordingExecutor{failureAt: test.failureAt, failure: failure}
			runner := gates.Runner{Root: fixture(t), Catalog: inventory.Inventory{Modules: []inventory.Module{{Directory: ".", Gates: test.gates}}}, Executor: executor}
			if err := runner.Check(context.Background(), []string{"."}); !errors.Is(err, failure) {
				t.Fatalf("Check() error = %v", err)
			}
		})
	}
}

func TestCheckRejectsUnformattedAndUnsafeSources(t *testing.T) {
	tests := map[string]string{
		"unformatted":   "package example\nfunc Value( )int{return 1}\n",
		"unsafe import": "package example\n\nimport \"unsafe\"\n\nvar _ = unsafe.Sizeof(0)\n",
		"cgo":           "package example\n\nimport \"C\"\n",
		"linkname":      "package example\n\n//go:linkname local target\nfunc local()\n",
	}
	for name, source := range tests {
		t.Run(name, func(t *testing.T) {
			root := fixture(t)
			write(t, filepath.Join(root, "unsafe.go"), source)
			runner := gates.Runner{Root: root, Catalog: inventory.Inventory{Modules: []inventory.Module{{Directory: "."}}}, Executor: &recordingExecutor{}}
			if err := runner.Check(context.Background(), []string{"."}); err == nil {
				t.Fatal("Check() error = nil")
			}
		})
	}
}

type recordingExecutor struct {
	commands  []string
	failureAt int
	failure   error
}

func (executor *recordingExecutor) Run(_ context.Context, command gates.Command) error {
	executor.commands = append(executor.commands, strings.Join(append([]string{command.Name}, command.Args...), " "))
	if executor.failure != nil && len(executor.commands)-1 == executor.failureAt {
		return executor.failure
	}
	return nil
}

func fixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	write(t, filepath.Join(root, "go.mod"), "module example\n\ngo 1.27.0\n")
	write(t, filepath.Join(root, "example.go"), "package example\n\nfunc Value() int { return 1 }\n")
	if err := os.MkdirAll(filepath.Join(root, "nested"), 0o700); err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(root, "nested", "go.mod"), "module example/nested\n\ngo 1.27.0\n")
	write(t, filepath.Join(root, "nested", "example.go"), "package nested\n\nfunc Value() int { return 1 }\n")
	return root
}

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}
