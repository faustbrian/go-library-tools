package gates

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/faustbrian/go-library-tools/internal/inventory"
	"github.com/faustbrian/go-library-tools/internal/services"
)

const serviceCleanupTimeout = 30 * time.Second

type serviceLease interface {
	Environment() map[string]string
	Identities() map[string]string
	Close(context.Context) error
}

type serviceStarter func(context.Context, []string) (serviceLease, error)

func (runner Runner) withModuleServices(ctx context.Context, module inventory.Module, operation func(Runner) error) error {
	if len(module.RequiredServices) == 0 {
		return operation(runner)
	}
	start := runner.startServices
	if start == nil {
		start = runner.defaultServiceStarter
	}
	lease, err := start(ctx, append([]string(nil), module.RequiredServices...))
	if err != nil {
		return fmt.Errorf("start services for %s: %w", module.Directory, err)
	}
	output := runner.Output
	if output == nil {
		output = io.Discard
	}
	_, _ = fmt.Fprintf(output, "[%s] services: %s\n", module.Directory, strings.Join(module.RequiredServices, ","))
	scoped := runner
	scoped.Executor = serviceEnvironmentExecutor{
		Executor: runner.Executor, environment: lease.Environment(), workspace: executorWorkspace(runner.Executor),
	}
	scoped.serviceIdentities = lease.Identities()
	operationErr := operation(scoped)
	cleanupContext, cancel := context.WithTimeout(context.Background(), serviceCleanupTimeout)
	cleanupErr := lease.Close(cleanupContext)
	cancel()
	return errors.Join(operationErr, cleanupErr)
}

func (runner Runner) defaultServiceStarter(ctx context.Context, names []string) (serviceLease, error) {
	manager := services.Manager{Process: func(ctx context.Context, name string, args []string, environment map[string]string, stdout, stderr io.Writer) error {
		return runner.Executor.Run(ctx, Command{Name: name, Args: args, Env: environment, Stdout: stdout, Stderr: stderr})
	}}
	return manager.Start(ctx, names)
}

type serviceEnvironmentExecutor struct {
	Executor
	environment map[string]string
	workspace   string
}

func (executor serviceEnvironmentExecutor) Run(ctx context.Context, command Command) error {
	command.Env = mergeMaps(executor.environment, command.Env)
	return executor.Executor.Run(ctx, command)
}

func (executor serviceEnvironmentExecutor) TemporaryDirectory() string { return executor.workspace }

func executorWorkspace(executor Executor) string {
	workspace, ok := executor.(taskWorkspace)
	if !ok {
		return ""
	}
	return workspace.TemporaryDirectory()
}

func mergeMaps(base, override map[string]string) map[string]string {
	result := make(map[string]string, len(base)+len(override))
	for key, value := range base {
		result[key] = value
	}
	for key, value := range override {
		result[key] = value
	}
	return result
}

func cloneServiceMap(source map[string]string) map[string]string {
	return mergeMaps(source, nil)
}
