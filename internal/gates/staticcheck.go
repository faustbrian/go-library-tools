package gates

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Staticcheck's released source needs the newer importer for Go 1.27 export
// data. Keep this tool-only graph separate from the module being analyzed.
const staticcheckBuildModule = `module golib.invalid/staticcheck-build

go 1.27.0

require (
 honnef.co/go/tools ` + staticcheckVersion + `
 golang.org/x/tools v0.50.0
)
`

func (runner Runner) staticcheck(ctx context.Context, output io.Writer, module, gate, directory string, args ...string) error {
	return announce(output, module, gate, func() (result error) {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("%s %s: %w", module, gate, err)
		}
		workspace, ok := runner.Executor.(taskWorkspace)
		if !ok || !filepath.IsAbs(workspace.TemporaryDirectory()) {
			return fmt.Errorf("%s %s: task-owned workspace required", module, gate)
		}
		root, err := os.MkdirTemp(workspace.TemporaryDirectory(), "staticcheck-")
		if err != nil {
			return fmt.Errorf("%s %s prepare: %w", module, gate, err)
		}
		defer func() {
			if err := os.RemoveAll(root); err != nil {
				result = errors.Join(result, fmt.Errorf("%s %s cleanup: %w", module, gate, err))
			}
		}()
		if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte(staticcheckBuildModule), 0o600); err != nil {
			return fmt.Errorf("%s %s prepare: %w", module, gate, err)
		}
		binary := filepath.Join(root, "staticcheck")
		if err := runner.Executor.Run(ctx, Command{
			Name: "go", Args: []string{"build", "-mod=mod", "-o", binary, "honnef.co/go/tools/cmd/staticcheck"},
			Dir: root, Env: map[string]string{"GOWORK": "off", "GOFLAGS": ""},
		}); err != nil {
			return fmt.Errorf("%s %s build: %w", module, gate, err)
		}
		if err := runner.Executor.Run(ctx, Command{
			Name: binary, Args: args, Dir: directory, Env: map[string]string{"GOWORK": "off"},
		}); err != nil {
			return fmt.Errorf("%s %s analysis: %w", module, gate, err)
		}
		return nil
	})
}
