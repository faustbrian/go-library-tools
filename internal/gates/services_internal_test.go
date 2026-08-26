package gates

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/faustbrian/go-library-tools/internal/config"
	"github.com/faustbrian/go-library-tools/internal/inventory"
	"github.com/faustbrian/go-library-tools/internal/mutation"
)

func TestWithModuleServicesScopesEnvironmentIdentityAndCleanup(t *testing.T) {
	lease := &fakeServiceLease{
		environment: map[string]string{"POSTGRES_URL": "postgres://fixture"},
		identities:  map[string]string{"postgresql": "postgres#sha256:identity"},
	}
	executor := &serviceRecordingExecutor{workspace: "/task"}
	var output bytes.Buffer
	runner := Runner{
		Executor: executor, Output: &output,
		startServices: func(_ context.Context, names []string) (serviceLease, error) {
			if !reflect.DeepEqual(names, []string{"postgresql"}) {
				t.Fatalf("service selection = %#v", names)
			}
			return lease, nil
		},
	}
	module := inventory.Module{Directory: ".", RequiredServices: []string{"postgresql"}}
	err := runner.withModuleServices(context.Background(), module, func(scoped Runner) error {
		if !reflect.DeepEqual(scoped.serviceIdentities, lease.identities) {
			t.Fatalf("service identities = %#v", scoped.serviceIdentities)
		}
		workspace, ok := scoped.Executor.(taskWorkspace)
		if !ok || workspace.TemporaryDirectory() != "/task" {
			t.Fatalf("scoped workspace = %T, %q", scoped.Executor, workspace.TemporaryDirectory())
		}
		return scoped.Executor.Run(context.Background(), Command{
			Name: "go", Env: map[string]string{"GOWORK": "off"}, Stdout: io.Discard,
		})
	})
	if err != nil {
		t.Fatalf("withModuleServices() error = %v", err)
	}
	if lease.closes != 1 || executor.commands[0].Env["POSTGRES_URL"] != "postgres://fixture" || executor.commands[0].Env["GOWORK"] != "off" {
		t.Fatalf("lease closes/command = %d, %#v", lease.closes, executor.commands)
	}
	if !strings.Contains(output.String(), "postgresql") {
		t.Fatalf("output = %q", output.String())
	}
}

func TestWithModuleServicesJoinsOperationAndCleanupFailures(t *testing.T) {
	operationFailure := errors.New("operation failed")
	cleanupFailure := errors.New("cleanup failed")
	lease := &fakeServiceLease{closeErr: cleanupFailure}
	runner := Runner{Executor: &serviceRecordingExecutor{}, startServices: func(context.Context, []string) (serviceLease, error) {
		return lease, nil
	}}
	err := runner.withModuleServices(context.Background(), inventory.Module{Directory: ".", RequiredServices: []string{"redis"}}, func(Runner) error {
		return operationFailure
	})
	if !errors.Is(err, operationFailure) || !errors.Is(err, cleanupFailure) || lease.closes != 1 {
		t.Fatalf("withModuleServices() error/closes = %v, %d", err, lease.closes)
	}
	startFailure := errors.New("start failed")
	runner.startServices = func(context.Context, []string) (serviceLease, error) { return nil, startFailure }
	if err := runner.withModuleServices(context.Background(), inventory.Module{RequiredServices: []string{"redis"}}, func(Runner) error { return nil }); !errors.Is(err, startFailure) {
		t.Fatalf("withModuleServices(start) error = %v", err)
	}
}

func TestWithModuleServicesSkipsEmptySelection(t *testing.T) {
	called := false
	runner := Runner{startServices: func(context.Context, []string) (serviceLease, error) {
		return nil, errors.New("must not start")
	}}
	if err := runner.withModuleServices(context.Background(), inventory.Module{}, func(scoped Runner) error {
		called = true
		return nil
	}); err != nil || !called {
		t.Fatalf("withModuleServices(empty) = %v, called = %v", err, called)
	}
}

func TestDefaultServiceStarterUsesExecutorWithoutDockerInTests(t *testing.T) {
	executor := &serviceRecordingExecutor{workspace: "/task", respond: true}
	runner := Runner{Executor: executor}
	err := runner.withModuleServices(context.Background(), inventory.Module{Directory: ".", RequiredServices: []string{"postgresql"}}, func(scoped Runner) error {
		if scoped.serviceIdentities["postgresql"] == "" {
			t.Fatalf("service identities = %#v", scoped.serviceIdentities)
		}
		return scoped.Executor.Run(context.Background(), Command{Name: "go"})
	})
	if err != nil {
		t.Fatalf("withModuleServices(default) error = %v", err)
	}
	if !strings.Contains(strings.Join(executor.names(), "\n"), "docker run") {
		t.Fatalf("commands = %#v", executor.names())
	}
	if workspace := executorWorkspace(executorFunction(func(context.Context, Command) error { return nil })); workspace != "" {
		t.Fatalf("executorWorkspace() = %q", workspace)
	}
}

func TestDefaultServiceStarterLoadsModuleOwnedOpenSearchPolicy(t *testing.T) {
	root := t.TempDir()
	moduleRoot := filepath.Join(root, "adapter")
	if err := os.MkdirAll(filepath.Join(moduleRoot, "scripts"), 0o700); err != nil {
		t.Fatal(err)
	}
	lock := `opensearch_image_repository='opensearchproject/opensearch'
opensearch_old_version='2.19.6'
opensearch_old_digest='sha256:8690b204fe914c60ca76d451ac73bc0481e034d32d3779944c8caca56a2b003f'
opensearch_new_version='3.8.0'
opensearch_new_digest='sha256:bcc1797519726ceb6d651d4a3e60b7c30da91793914a8dfe75fd441d4f641509'
`
	if err := os.WriteFile(filepath.Join(moduleRoot, "scripts", "opensearch-images.env"), []byte(lock), 0o600); err != nil {
		t.Fatal(err)
	}
	executor := &serviceRecordingExecutor{workspace: filepath.Join(root, ".task"), respond: true}
	var readinessURL string
	runner := Runner{
		Root: root, Executor: executor,
		serviceHTTPProbe: func(_ context.Context, value string) error {
			readinessURL = value
			return nil
		},
	}
	err := runner.withModuleServices(context.Background(), inventory.Module{
		Directory: "adapter", RequiredServices: []string{"opensearch"},
	}, func(scoped Runner) error {
		if !strings.Contains(scoped.serviceIdentities["opensearch"], "opensearchproject/opensearch@sha256:bcc179") {
			t.Fatalf("service identities = %#v", scoped.serviceIdentities)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("withModuleServices(OpenSearch) error = %v", err)
	}
	if readinessURL != "http://127.0.0.1:49152/" {
		t.Fatalf("readiness URL = %q", readinessURL)
	}
	if commands := strings.Join(executor.names(), "\n"); !strings.Contains(commands, "opensearchproject/opensearch@sha256:bcc179") {
		t.Fatalf("commands = %s", commands)
	}
}

func TestDefaultServiceStarterRejectsMissingAndMalformedOpenSearchPolicy(t *testing.T) {
	root := t.TempDir()
	runner := Runner{Root: root, Executor: &serviceRecordingExecutor{respond: true}}
	module := inventory.Module{Directory: ".", RequiredServices: []string{"opensearch"}}
	if err := runner.withModuleServices(context.Background(), module, func(Runner) error { return nil }); err == nil || !strings.Contains(err.Error(), "read OpenSearch image policy") {
		t.Fatalf("withModuleServices(missing policy) error = %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "scripts"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "scripts", "opensearch-images.env"), []byte("malformed\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := runner.withModuleServices(context.Background(), module, func(Runner) error { return nil }); err == nil || !strings.Contains(err.Error(), "parse OpenSearch image policy") {
		t.Fatalf("withModuleServices(malformed policy) error = %v", err)
	}
}

func TestCheckCoverageAndMutationUseModuleServiceScope(t *testing.T) {
	root := t.TempDir()
	writeValidMutationSetup(t, root)
	if err := os.WriteFile(filepath.Join(root, "source.go"), []byte("package example\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	module := inventory.Module{
		Directory: ".", ModulePath: "example", GoVersion: "1.27.0", RequiredServices: []string{"postgresql"},
		Gates:    map[string]bool{"coverage": true, "mutation": true},
		Packages: []inventory.Package{{Directory: ".", ImportPath: "example", CoverageRequired: true}},
	}
	for _, operation := range []struct {
		name string
		run  func(Runner) error
	}{
		{"check", func(runner Runner) error { return runner.Check(context.Background(), []string{"."}) }},
		{"coverage", func(runner Runner) error { return runner.Coverage(context.Background(), []string{"."}) }},
		{"mutation", func(runner Runner) error { return runner.Mutation(context.Background(), []string{"."}) }},
	} {
		t.Run(operation.name, func(t *testing.T) {
			if err := os.MkdirAll(filepath.Join(root, ".task"), 0o700); err != nil {
				t.Fatal(err)
			}
			lease := &fakeServiceLease{environment: map[string]string{"POSTGRES_URL": "postgres://fixture"}, identities: map[string]string{"postgresql": "postgres#sha256:identity"}}
			executor := &serviceRecordingExecutor{workspace: filepath.Join(root, ".task"), coverage: true}
			mutationObserved := false
			runner := Runner{
				Root: root, Catalog: inventory.Inventory{Repository: "example", Modules: []inventory.Module{module}},
				Policy:        config.Config{Evidence: config.Evidence{Root: ".verification"}, Mutation: config.Mutation{Root: ".verification/mutation"}},
				Executor:      executor,
				startServices: func(context.Context, []string) (serviceLease, error) { return lease, nil },
				mutationCampaign: func(_ context.Context, campaign mutation.Campaign) error {
					mutationObserved = true
					if !reflect.DeepEqual(campaign.Policy.RequiredServices, []string{"postgresql"}) || campaign.Policy.ServiceIdentities["postgresql"] == "" {
						t.Fatalf("mutation service policy = %#v", campaign.Policy)
					}
					return nil
				},
			}
			if err := operation.run(runner); err != nil {
				t.Fatalf("%s error = %v", operation.name, err)
			}
			if lease.closes != 1 {
				t.Fatalf("%s closes = %d", operation.name, lease.closes)
			}
			if operation.name != "coverage" && !mutationObserved {
				t.Fatalf("%s mutation campaign was not observed", operation.name)
			}
		})
	}
}

type fakeServiceLease struct {
	mu          sync.Mutex
	environment map[string]string
	identities  map[string]string
	closes      int
	closeErr    error
}

func (lease *fakeServiceLease) Environment() map[string]string {
	return cloneServiceMap(lease.environment)
}
func (lease *fakeServiceLease) Identities() map[string]string {
	return cloneServiceMap(lease.identities)
}
func (lease *fakeServiceLease) Close(context.Context) error {
	lease.mu.Lock()
	defer lease.mu.Unlock()
	lease.closes++
	return lease.closeErr
}

type serviceRecordingExecutor struct {
	mu        sync.Mutex
	commands  []Command
	workspace string
	respond   bool
	coverage  bool
}

func (executor *serviceRecordingExecutor) Run(_ context.Context, command Command) error {
	executor.mu.Lock()
	defer executor.mu.Unlock()
	executor.commands = append(executor.commands, command)
	if executor.respond && command.Name == "docker" && len(command.Args) > 0 {
		switch command.Args[0] {
		case "inspect":
			_, _ = io.WriteString(command.Stdout, "sha256:"+strings.Repeat("a", 64))
		case "port":
			_, _ = io.WriteString(command.Stdout, "127.0.0.1:49152\n")
		}
	}
	if command.Name == "go" && reflect.DeepEqual(command.Args, []string{"env", "-json", "GOVERSION", "GOOS", "GOARCH", "CGO_ENABLED"}) {
		_, _ = io.WriteString(command.Stdout, `{"GOVERSION":"go1.27.0","GOOS":"linux","GOARCH":"amd64","CGO_ENABLED":"0"}`)
	}
	if executor.coverage {
		for _, argument := range command.Args {
			if strings.HasPrefix(argument, "-coverprofile=") {
				return os.WriteFile(strings.TrimPrefix(argument, "-coverprofile="), []byte("mode: atomic\nexample/file.go:1.1,2.1 1 1\n"), 0o600)
			}
		}
	}
	return nil
}

func (executor *serviceRecordingExecutor) TemporaryDirectory() string { return executor.workspace }

func (executor *serviceRecordingExecutor) names() []string {
	executor.mu.Lock()
	defer executor.mu.Unlock()
	result := make([]string, len(executor.commands))
	for index, command := range executor.commands {
		result[index] = command.Name + " " + strings.Join(command.Args, " ")
	}
	return result
}
