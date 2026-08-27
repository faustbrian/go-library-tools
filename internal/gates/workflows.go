package gates

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
)

const (
	actionlintVersion     = "v1.7.12"
	maximumWorkflowOutput = 4 << 20
)

type boundedWorkflowBuffer struct {
	data     bytes.Buffer
	overflow bool
}

func (buffer *boundedWorkflowBuffer) Write(value []byte) (int, error) {
	remaining := maximumWorkflowOutput - buffer.data.Len()
	if len(value) <= remaining {
		return buffer.data.Write(value)
	}
	buffer.overflow = true
	_, _ = buffer.data.Write(value[:remaining])
	return len(value), nil
}

// Workflows validates every GitHub Actions workflow with the centrally pinned
// Actionlint release.
func (runner Runner) Workflows(ctx context.Context) error {
	output := runner.Output
	if output == nil {
		output = io.Discard
	}
	var standardOutput, diagnostics boundedWorkflowBuffer
	err := runner.Executor.Run(ctx, Command{
		Name: "go", Dir: runner.Root, Env: map[string]string{"GOWORK": "off"},
		Args: []string{
			"run", "github.com/rhysd/actionlint/cmd/actionlint@" + actionlintVersion,
			"-no-color", "-oneline", "-shellcheck=", "-pyflakes=",
		},
		Stdout: &standardOutput,
		Stderr: &diagnostics,
	})
	if standardOutput.overflow || diagnostics.overflow {
		return fmt.Errorf("actionlint output exceeded %d bytes", maximumWorkflowOutput)
	}
	if err != nil {
		message := strings.TrimSpace(standardOutput.data.String() + "\n" + diagnostics.data.String())
		if message != "" {
			return fmt.Errorf("actionlint failed: %w: %s", err, message)
		}
		return fmt.Errorf("actionlint failed: %w", err)
	}
	_, _ = io.WriteString(output, "workflow contract passed\n")
	return nil
}
