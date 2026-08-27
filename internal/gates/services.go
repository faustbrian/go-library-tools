package gates

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"maps"
	"path/filepath"
	"slices"
	"strings"

	"github.com/faustbrian/go-library-tools/internal/inventory"
	"github.com/faustbrian/go-library-tools/internal/repositoryfile"
	"github.com/faustbrian/go-library-tools/internal/services"
)

const (
	serviceCleanupTimeout           = 30_000_000_000
	openSearchImageLock             = "scripts/opensearch-images.env"
	maximumOpenSearchImageLockBytes = 16 << 10
)

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
		start = func(ctx context.Context, names []string) (serviceLease, error) {
			return runner.defaultServiceStarter(ctx, module, names)
		}
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

func (runner Runner) defaultServiceStarter(ctx context.Context, module inventory.Module, names []string) (serviceLease, error) {
	manager := services.Manager{Process: func(ctx context.Context, name string, args []string, environment map[string]string, stdout, stderr io.Writer) error {
		return runner.Executor.Run(ctx, Command{Name: name, Args: args, Env: environment, Stdout: stdout, Stderr: stderr})
	}, HTTPProbe: runner.serviceHTTPProbe, Workspace: executorWorkspace(runner.Executor)}
	if containsService(names, "opensearch") {
		images, err := runner.loadOpenSearchImages(module)
		if err != nil {
			return nil, err
		}
		manager.OpenSearch = &images
	}
	return manager.Start(ctx, names)
}

func (runner Runner) loadOpenSearchImages(module inventory.Module) (services.OpenSearchImages, error) {
	relative := filepath.Join(module.Directory, openSearchImageLock)
	data, err := repositoryfile.Read(runner.Root, relative, maximumOpenSearchImageLockBytes)
	if err != nil {
		return services.OpenSearchImages{}, fmt.Errorf("read OpenSearch image policy for %s: %w", module.Directory, err)
	}
	images, err := services.ParseOpenSearchImages(bytes.NewReader(data))
	if err != nil {
		return services.OpenSearchImages{}, fmt.Errorf("parse OpenSearch image policy for %s: %w", module.Directory, err)
	}
	return images, nil
}

func containsService(names []string, target string) bool {
	return slices.Contains(names, target)
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
	result := make(map[string]string)
	maps.Copy(result, base)
	maps.Copy(result, override)
	return result
}

func cloneServiceMap(source map[string]string) map[string]string {
	return mergeMaps(source, nil)
}
