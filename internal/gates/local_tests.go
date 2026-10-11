package gates

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/faustbrian/go-library-tools/v2/internal/config"
	"github.com/faustbrian/go-library-tools/v2/internal/inventory"
)

func (runner Runner) preflightLocalTests(modules []inventory.Module) error {
	var budget time.Duration
	for _, module := range modules {
		operation, exists := runner.operation(module.Directory, "local-test")
		if !exists {
			continue
		}
		duration, err := operation.LocalTestBudget()
		if err != nil {
			return err
		}
		if duration > config.MaximumLocalTestBudget-budget {
			return errors.New("local-test budget across selected modules exceeds three minutes")
		}
		budget += duration
	}
	return nil
}

func (runner Runner) runLocalTests(ctx context.Context, output io.Writer, module inventory.Module) error {
	operation, exists := runner.operation(module.Directory, "local-test")
	if !exists {
		return nil
	}
	if _, err := operation.LocalTestBudget(); err != nil {
		return err
	}
	return announce(output, module.Directory, "local-test", func() error {
		for index, step := range operation.Steps {
			timeout, _ := time.ParseDuration(step.Timeout)
			stepContext, cancel := context.WithTimeout(ctx, timeout)
			err := runner.runLocalTest(stepContext, module, step)
			cancel()
			if err != nil {
				return fmt.Errorf("%s local-test step %d: %w", module.Directory, index, err)
			}
		}
		return nil
	})
}

func (runner Runner) runLocalTest(ctx context.Context, module inventory.Module, step config.Step) error {
	command, err := operationCommand(filepath.Join(runner.Root, module.Directory), module, step)
	if err != nil {
		return err
	}
	list := command
	list.Args = append(slices.Clone(command.Args), "-list="+step.Run)
	names, err := runner.capturedTestCommand(ctx, list, 64<<10, "local-test discovery failed or exceeded its output budget")
	if err != nil {
		return err
	}
	selector, err := regexp.Compile(step.Run)
	if err != nil {
		return errors.New("local-test selector is invalid")
	}
	matched := false
	for name := range strings.SplitSeq(names, "\n") {
		if strings.HasPrefix(name, "Test") && selector.MatchString(name) {
			matched = true
			break
		}
	}
	if !matched {
		return errors.New("local-test selector has no matching tests")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	_, err = runner.capturedTestCommand(ctx, command, maximumSecurityProcessOutput, "local-test execution failed or exceeded its output budget")
	return err
}
