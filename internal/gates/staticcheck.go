package gates

import (
	"context"
	"fmt"
	"io"
)

func (runner Runner) staticcheck(ctx context.Context, output io.Writer, module, gate, directory string, args ...string) error {
	return announce(output, module, gate, func() error {
		err := runner.withCompilerTool(ctx, staticcheckCompilerTool, func(binary string) error {
			if err := runner.Executor.Run(ctx, Command{Name: binary, Args: args, Dir: directory, Env: map[string]string{"GOWORK": "off"}}); err != nil {
				return fmt.Errorf("analysis: %w", err)
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("%s %s: %w", module, gate, err)
		}
		return nil
	})
}
