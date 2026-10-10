package gates

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// compilerTool fixes only the tool's importer graph. Its released analyzer
// source and the application's module graph remain independent authorities.
type compilerTool struct {
	name    string
	module  string
	version string
	command string
}

var staticcheckCompilerTool = compilerTool{
	name: "staticcheck", module: "honnef.co/go/tools", version: staticcheckVersion,
	command: "honnef.co/go/tools/cmd/staticcheck",
}

func securityCompilerTool(identity string) (compilerTool, bool) {
	switch identity {
	case "github.com/securego/gosec/v2/cmd/gosec@" + gosecVersion:
		return compilerTool{name: "gosec", module: "github.com/securego/gosec/v2", version: gosecVersion, command: "github.com/securego/gosec/v2/cmd/gosec"}, true
	case "github.com/faustbrian/go-analysis/cmd/golib-analysis@" + goAnalysisVersion:
		return compilerTool{name: "golib-analysis", module: "github.com/faustbrian/go-analysis", version: goAnalysisVersion, command: "github.com/faustbrian/go-analysis/cmd/golib-analysis"}, true
	default:
		return compilerTool{}, false
	}
}

// withCompilerTool owns a single invocation's build graph and executable.
// The caller retains its existing analysis arguments, output policy and errors.
func (runner Runner) withCompilerTool(ctx context.Context, tool compilerTool, analyze func(string) error) (result error) {
	if err := ctx.Err(); err != nil {
		return err
	}
	workspace, ok := runner.Executor.(taskWorkspace)
	if !ok || !filepath.IsAbs(workspace.TemporaryDirectory()) {
		return errors.New("task-owned workspace required")
	}
	root, err := os.MkdirTemp(workspace.TemporaryDirectory(), tool.name+"-")
	if err != nil {
		return fmt.Errorf("prepare compiler tool: %w", err)
	}
	defer func() {
		if err := os.RemoveAll(root); err != nil {
			result = errors.Join(result, fmt.Errorf("compiler tool cleanup: %w", err))
		}
	}()
	manifest := "module golib.invalid/" + tool.name + "-build\n\ngo 1.27.0\n\nrequire (\n " + tool.module + " " + tool.version + "\n golang.org/x/tools v0.50.0\n)\n"
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte(manifest), 0o600); err != nil {
		return fmt.Errorf("prepare compiler tool: %w", err)
	}
	binary := filepath.Join(root, tool.name)
	// Build diagnostics are untrusted too. Never expose downloaded source or
	// compiler output through the public security gate's error channel.
	stdout := &boundedProcessOutput{limit: maximumSecurityProcessOutput}
	stderr := &boundedProcessOutput{limit: maximumSecurityProcessOutput}
	err = runner.Executor.Run(ctx, Command{
		Name: "go", Args: []string{"build", "-mod=mod", "-o", binary, tool.command},
		Dir: root, Env: map[string]string{"GOWORK": "off", "GOFLAGS": "", "GOTOOLCHAIN": "go1.27.2"},
		boundedScanner: true, Stdout: stdout, Stderr: stderr,
	})
	var overflow error
	if stdout.didOverflow() || stderr.didOverflow() {
		overflow = fmt.Errorf("compiler tool build output exceeded %d bytes", maximumSecurityProcessOutput)
	}
	if err != nil || overflow != nil {
		return fmt.Errorf("compiler tool build: %w", errors.Join(err, overflow))
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return analyze(binary)
}
