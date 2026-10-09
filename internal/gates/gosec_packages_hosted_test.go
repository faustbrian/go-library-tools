//go:build scannerintegration

package gates

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/faustbrian/go-library-tools/v2/internal/inventory"
)

func TestGosecSelectedPackagesHosted(t *testing.T) {
	if os.Getenv("GITHUB_ACTIONS") != "true" {
		t.Skip("pinned scanner integration runs only in hosted CI")
	}
	root := t.TempDir()
	for name, content := range map[string]string{
		"go.mod":                 "module example.invalid/selection\n\ngo 1.27.0\n",
		"root.go":                "package selection\nfunc Value() int { return 1 }\n",
		"newpackage/new.go":      "package newpackage\nfunc Value() int { return 2 }\n",
		"testsupport/support.go": "package testsupport\nfunc Value() int { return 3 }\n",
		// Deliberately incomplete fixture dependencies must not become scanner
		// targets merely because recursive filesystem discovery finds a Go file.
		"testdata/fixture/fixture.go": "package fixture\nimport _ \"example.invalid/selection/missing\"\n",
	} {
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	executor, cleanup, err := NewProcessExecutor(root, io.Discard, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := cleanup(); err != nil {
			t.Error(err)
		}
	})
	tool := "github.com/securego/gosec/v2/cmd/gosec@" + gosecVersion
	if err := executor.Run(t.Context(), Command{
		Name: "go", Args: []string{"run", tool, "-nosec-require-rules", "-nosec-require-justification", "./..."},
		Dir: root, Env: map[string]string{"GOWORK": "off"}, boundedScanner: true,
		Stdout: &boundedProcessOutput{limit: maximumSecurityProcessOutput}, Stderr: &boundedProcessOutput{limit: maximumSecurityProcessOutput},
	}); err == nil {
		t.Fatal("recursive scanner counterfactual unexpectedly admitted incomplete fixture dependencies")
	}
	stop := errors.New("stop after production gosec boundary")
	report := &goPackageOutput{}
	report.limit = maximumSecurityProcessOutput
	runner := Runner{Root: root, Executor: workspaceExecutor{directory: executorWorkspace(executor), run: func(ctx context.Context, command Command) error {
		if slices.Contains(command.Args, "list") || (len(command.Args) > 0 && command.Args[0] == "build" && strings.HasSuffix(command.Args[len(command.Args)-1], "/gosec")) {
			return executor.Run(ctx, command)
		}
		if filepath.Base(command.Name) == "gosec" {
			if !slices.Contains(command.Args, "-fmt=json") {
				t.Fatal("production Gosec does not request structured output")
			}
			command.Stdout = report
			return executor.Run(ctx, command)
		}
		if filepath.Base(command.Name) == "golib-analysis" || slices.Contains(command.Args, "github.com/faustbrian/go-analysis/cmd/golib-analysis") {
			return stop
		}
		return nil
	}}}
	err = runner.runSecurity(t.Context(), io.Discard, root, inventory.Module{Directory: "."})
	if !errors.Is(err, stop) || report.didOverflow() {
		t.Fatal("production gosec selection did not complete successfully")
	}
	var result struct {
		Stats struct {
			Files int `json:"files"`
		} `json:"Stats"`
	}
	if err := json.Unmarshal(report.data.Bytes(), &result); err != nil || result.Stats.Files != 3 {
		t.Fatal("pinned scanner did not scan root, newly added production and test-support files")
	}
}
