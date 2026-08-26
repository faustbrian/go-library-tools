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
	"testing"

	"github.com/faustbrian/go-library-tools/internal/config"
	"github.com/faustbrian/go-library-tools/internal/inventory"
	"github.com/faustbrian/go-library-tools/internal/mutation"
)

func TestMutationBuildsCanonicalCampaignAndRunsEnabledModules(t *testing.T) {
	root := t.TempDir()
	writeMutationFixture(t, root, ".", "module example\n\nrequire github.com/testcontainers/testcontainers-go v0.0.0\n")
	writeMutationFixture(t, root, "nested", "module example/nested\n")
	mutationRoot := filepath.Join(root, ".verification", "mutation")
	if err := os.MkdirAll(mutationRoot, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(mutationRoot, "zero-inventory.json"), []byte("{\"schema_version\":1,\"packages\":[]}"), 0o600); err != nil {
		t.Fatal(err)
	}

	processCalled := false
	executor := workspaceExecutor{directory: filepath.Join(root, ".task"), run: func(_ context.Context, command Command) error {
		if command.Name == "go" && reflect.DeepEqual(command.Args, []string{"env", "-json", "GOVERSION", "GOOS", "GOARCH", "CGO_ENABLED"}) {
			_, _ = io.WriteString(command.Stdout, `{"GOVERSION":"go1.27.0","GOOS":"linux","GOARCH":"amd64","CGO_ENABLED":"0"}`)
		}
		if command.Name == "probe" {
			processCalled = true
		}
		return nil
	}}
	var observed mutation.Campaign
	var output bytes.Buffer
	runner := Runner{
		Root: root,
		Catalog: inventory.Inventory{Repository: "example", Modules: []inventory.Module{
			{
				Directory: ".", ModulePath: "example", GoVersion: "1.27.0",
				Gates: map[string]bool{"mutation": true}, TestTags: []string{"integration", "interoperability"},
				BuildTags: []string{"custom"}, OwnedDependencies: []string{"example", "example/zeta", "example/nested", "example/fixture"},
				Packages: []inventory.Package{
					{Directory: ".", CoverageRequired: true},
					{Directory: "testsupport", CoverageRequired: false},
				},
			},
			{Directory: "nested", ModulePath: "example/nested", Kind: "public"},
			{Directory: "fixture", ModulePath: "example/fixture", Kind: "fixture"},
			{Directory: "other", ModulePath: "example/other", Kind: "public"},
			{Directory: "zeta", ModulePath: "example/zeta", Kind: "public"},
		}},
		Policy:   config.Config{Evidence: config.Evidence{Root: ".verification"}, Mutation: config.Mutation{Root: ".verification/mutation"}},
		Executor: executor,
		Output:   &output,
		mutationCampaign: func(_ context.Context, campaign mutation.Campaign) error {
			observed = campaign
			return nil
		},
	}

	if err := runner.Mutation(context.Background(), []string{"."}); err != nil {
		t.Fatalf("Mutation() error = %v", err)
	}
	if observed.Policy.Repository != "example" || observed.Policy.ModulePath != "example" ||
		!reflect.DeepEqual(observed.Policy.Packages, []string{"."}) ||
		!reflect.DeepEqual(observed.Policy.TestTags, []string{"integration"}) ||
		!reflect.DeepEqual(observed.Policy.BuildTags, []string{"custom"}) || observed.Policy.Workers != 1 {
		t.Fatalf("campaign policy = %#v", observed.Policy)
	}
	if !reflect.DeepEqual(observed.Policy.OwnedModules, []mutation.OwnedModule{
		{ModulePath: "example/nested", Directory: "nested"},
		{ModulePath: "example/zeta", Directory: "zeta"},
	}) {
		t.Fatalf("owned modules = %#v", observed.Policy.OwnedModules)
	}
	if observed.RuntimeIdentity.GoVersion != "go1.27.0" || observed.Workspace != executor.directory || observed.Process == nil {
		t.Fatalf("campaign runtime/workspace/process = %#v, %q, %v", observed.RuntimeIdentity, observed.Workspace, observed.Process)
	}
	if err := observed.Process(context.Background(), "probe", []string{"arg"}, root, map[string]string{"A": "B"}, io.Discard, io.Discard); err != nil || !processCalled {
		t.Fatalf("campaign process = %v, called = %v", err, processCalled)
	}
	if !strings.Contains(output.String(), "mutation") {
		t.Fatalf("output = %q", output.String())
	}
}

func TestMutationReportsDisabledModulesWithoutExecution(t *testing.T) {
	var output bytes.Buffer
	runner := Runner{
		Root: t.TempDir(), Catalog: inventory.Inventory{Modules: []inventory.Module{{Directory: ".", Gates: map[string]bool{}}}},
		Executor: executorFunction(func(context.Context, Command) error { return errors.New("must not execute") }), Output: &output,
	}
	if err := runner.Mutation(context.Background(), []string{"."}); err != nil {
		t.Fatalf("Mutation() error = %v", err)
	}
	if !strings.Contains(output.String(), "not applicable") {
		t.Fatalf("output = %q", output.String())
	}
	runner.Output = nil
	if err := runner.Mutation(context.Background(), []string{"."}); err != nil {
		t.Fatalf("Mutation() discarded output error = %v", err)
	}
	if err := runner.Mutation(context.Background(), []string{"missing"}); err == nil {
		t.Fatal("Mutation() unknown module error = nil")
	}
}

func TestMutationFailsClosedForInvalidSetup(t *testing.T) {
	failure := errors.New("campaign failed")
	tests := []struct {
		name     string
		prepare  func(*testing.T, string)
		executor Executor
		run      mutationCampaignRunner
		want     string
	}{
		{"workspace", writeValidMutationSetup, executorFunction(func(context.Context, Command) error { return nil }), nil, "task-owned workspace"},
		{"missing inventory", func(t *testing.T, root string) { writeMutationFixture(t, root, ".", "module example\n") }, workspaceExecutor{directory: "/task", run: func(context.Context, Command) error { return nil }}, nil, "zero-mutant inventory"},
		{"runtime", writeValidMutationSetup, workspaceExecutor{directory: "/task", run: func(_ context.Context, command Command) error {
			_, _ = io.WriteString(command.Stdout, "not-json")
			return nil
		}}, nil, "runtime identity"},
		{"runtime command", writeValidMutationSetup, workspaceExecutor{directory: "/task", run: func(context.Context, Command) error { return failure }}, nil, "read Go runtime identity"},
		{"services", writeValidMutationSetup, runtimeExecutor("/task"), nil, "service fixture"},
		{"campaign", writeValidMutationSetup, runtimeExecutor("/task"), func(context.Context, mutation.Campaign) error { return failure }, "campaign failed"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			test.prepare(t, root)
			module := inventory.Module{Directory: ".", ModulePath: "example", GoVersion: "1.27.0", Gates: map[string]bool{"mutation": true}, Packages: []inventory.Package{{Directory: ".", CoverageRequired: true}}}
			if test.name == "services" {
				module.RequiredServices = []string{"postgresql"}
			}
			runner := Runner{
				Root: root, Catalog: inventory.Inventory{Repository: "example", Modules: []inventory.Module{module}},
				Policy:   config.Config{Evidence: config.Evidence{Root: ".verification"}, Mutation: config.Mutation{Root: ".verification/mutation"}},
				Executor: test.executor, mutationCampaign: test.run,
			}
			err := runner.Mutation(context.Background(), []string{"."})
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Mutation() error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestMutationRejectsMalformedAndSymlinkedReviewInventories(t *testing.T) {
	root := t.TempDir()
	directory := filepath.Join(root, ".verification", "mutation")
	if err := os.MkdirAll(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(directory, "zero-inventory.json")
	if err := os.WriteFile(path, []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := loadZeroInventory(root, ".verification/mutation/zero-inventory.json"); err == nil || !strings.Contains(err.Error(), "parse zero-mutant inventory") {
		t.Fatalf("loadZeroInventory(malformed) error = %v", err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(root, "missing"), path); err != nil {
		t.Fatal(err)
	}
	if _, err := loadZeroInventory(root, ".verification/mutation/zero-inventory.json"); err == nil || !strings.Contains(err.Error(), "read zero-mutant inventory") {
		t.Fatalf("loadZeroInventory(symlink) error = %v", err)
	}
}

func TestMutationEnvironmentUsesIsolatedModfileAndHandlesFailures(t *testing.T) {
	root := t.TempDir()
	writeMutationFixture(t, root, ".", "module example\n")
	if err := os.WriteFile(filepath.Join(root, "go.sum"), []byte("sum\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	owned := []mutation.OwnedModule{{ModulePath: "example/nested", Directory: "nested"}}
	commands := 0
	runner := Runner{Root: root, Executor: executorFunction(func(_ context.Context, command Command) error {
		commands++
		if !strings.Contains(strings.Join(command.Args, " "), "-replace=example/nested=") {
			t.Fatalf("mod edit command = %#v", command.Args)
		}
		return nil
	})}
	environment, err := runner.mutationEnvironment(context.Background(), t.TempDir(), inventory.Module{Directory: ".", ModulePath: "example"}, owned)
	if err != nil || commands != 1 || !strings.Contains(environment["GOFLAGS"], "-modfile=") {
		t.Fatalf("mutationEnvironment() = %#v, commands = %d, error = %v", environment, commands, err)
	}

	failure := errors.New("injected failure")
	tests := []struct {
		name  string
		files mutationFileSystem
		exec  Executor
		root  func(*testing.T) string
		want  string
	}{
		{"module read", operatingMutationFiles{}, runner.Executor, func(t *testing.T) string { return t.TempDir() }, "read mutation go.mod"},
		{"mkdir", &failingMutationFiles{mkdirErr: failure}, runner.Executor, func(*testing.T) string { return root }, "create mutation modfile directory"},
		{"modfile write", &failingMutationFiles{writeErrAt: 1, failure: failure}, runner.Executor, func(*testing.T) string { return root }, "write mutation modfile"},
		{"sumfile write", &failingMutationFiles{writeErrAt: 2, failure: failure}, runner.Executor, func(*testing.T) string { return root }, "write mutation sumfile"},
		{"mod edit", operatingMutationFiles{}, executorFunction(func(context.Context, Command) error { return failure }), func(*testing.T) string { return root }, "configure mutation owned module"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			candidate := Runner{Root: test.root(t), Executor: test.exec, mutationFiles: test.files}
			_, err := candidate.mutationEnvironment(context.Background(), t.TempDir(), inventory.Module{Directory: ".", ModulePath: "example"}, owned)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("mutationEnvironment() error = %v, want %q", err, test.want)
			}
		})
	}
	unsafeRoot := t.TempDir()
	writeMutationFixture(t, unsafeRoot, ".", "module example\n")
	if err := os.Symlink(filepath.Join(unsafeRoot, "missing"), filepath.Join(unsafeRoot, "go.sum")); err != nil {
		t.Fatal(err)
	}
	unsafe := Runner{Root: unsafeRoot, Executor: runner.Executor}
	if _, err := unsafe.mutationEnvironment(context.Background(), t.TempDir(), inventory.Module{Directory: ".", ModulePath: "example"}, owned); err == nil || !strings.Contains(err.Error(), "read mutation go.sum") {
		t.Fatalf("mutationEnvironment() unsafe go.sum error = %v", err)
	}
}

func TestRunMutationPropagatesPreparationFailures(t *testing.T) {
	for _, test := range []struct {
		name    string
		catalog inventory.Inventory
		want    string
	}{
		{
			name: "environment",
			catalog: inventory.Inventory{Repository: "example", Modules: []inventory.Module{
				{Directory: ".", ModulePath: "example", GoVersion: "1.27.0", OwnedDependencies: []string{"example/nested"}, Packages: []inventory.Package{{Directory: ".", CoverageRequired: true}}},
				{Directory: "nested", ModulePath: "example/nested"},
			}},
			want: "read mutation go.mod",
		},
		{
			name: "workers",
			catalog: inventory.Inventory{Repository: "example", Modules: []inventory.Module{
				{Directory: ".", ModulePath: "example", GoVersion: "1.27.0", Packages: []inventory.Package{{Directory: ".", CoverageRequired: true}}},
			}},
			want: "read mutation worker policy",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			directory := filepath.Join(root, ".verification", "mutation")
			if err := os.MkdirAll(directory, 0o700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(directory, "zero-inventory.json"), []byte("{\"schema_version\":1,\"packages\":[]}"), 0o600); err != nil {
				t.Fatal(err)
			}
			runner := Runner{
				Root: root, Catalog: test.catalog,
				Policy:   config.Config{Evidence: config.Evidence{Root: ".verification"}, Mutation: config.Mutation{Root: ".verification/mutation"}},
				Executor: runtimeExecutor(filepath.Join(root, ".task")),
			}
			err := runner.runMutation(context.Background(), io.Discard, test.catalog.Modules[0])
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("runMutation() error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestMutationWorkerPolicyAndDefaultCampaign(t *testing.T) {
	root := t.TempDir()
	writeValidMutationSetup(t, root)
	if err := os.WriteFile(filepath.Join(root, "source.go"), []byte("package example\n\nfunc Value() int { return 1 }\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	runner := Runner{Root: root}
	workers, err := runner.mutationWorkers(inventory.Module{Directory: "."})
	if err != nil || workers != 4 {
		t.Fatalf("mutationWorkers() = %d, %v", workers, err)
	}
	if _, err := (Runner{Root: t.TempDir()}).mutationWorkers(inventory.Module{Directory: "."}); err == nil {
		t.Fatal("mutationWorkers() missing go.mod error = nil")
	}

	task := filepath.Join(root, ".task")
	executor := workspaceExecutor{directory: task, run: func(_ context.Context, command Command) error {
		if reflect.DeepEqual(command.Args, []string{"env", "-json", "GOVERSION", "GOOS", "GOARCH", "CGO_ENABLED"}) {
			_, _ = io.WriteString(command.Stdout, `{"GOVERSION":"go1.27.0","GOOS":"linux","GOARCH":"amd64","CGO_ENABLED":"0"}`)
			return nil
		}
		if len(command.Args) > 0 && command.Args[0] == "list" {
			return errors.New("stop campaign")
		}
		return nil
	}}
	runner = Runner{
		Root: root, Catalog: inventory.Inventory{Repository: "example", Modules: []inventory.Module{{
			Directory: ".", ModulePath: "example", GoVersion: "1.27.0", Gates: map[string]bool{"mutation": true},
			Packages: []inventory.Package{{Directory: ".", CoverageRequired: true}},
		}}},
		Policy: config.Config{Evidence: config.Evidence{Root: ".verification"}, Mutation: config.Mutation{Root: ".verification/mutation"}}, Executor: executor,
	}
	if err := runner.Mutation(context.Background(), []string{"."}); err == nil || !strings.Contains(err.Error(), "stop campaign") {
		t.Fatalf("default campaign error = %v", err)
	}
}

func TestCheckRunsAndStopsAtMutationGate(t *testing.T) {
	root := t.TempDir()
	writeValidMutationSetup(t, root)
	if err := os.WriteFile(filepath.Join(root, "source.go"), []byte("package example\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	failure := errors.New("mutation failed")
	for _, test := range []struct {
		name string
		run  mutationCampaignRunner
		want error
	}{
		{"pass", func(context.Context, mutation.Campaign) error { return nil }, nil},
		{"fail", func(context.Context, mutation.Campaign) error { return failure }, failure},
	} {
		t.Run(test.name, func(t *testing.T) {
			runner := Runner{
				Root: root, Catalog: inventory.Inventory{Repository: "example", Modules: []inventory.Module{{
					Directory: ".", ModulePath: "example", GoVersion: "1.27.0", Gates: map[string]bool{"mutation": true},
					Packages: []inventory.Package{{Directory: ".", CoverageRequired: true}},
				}}},
				Policy:   config.Config{Evidence: config.Evidence{Root: ".verification"}, Mutation: config.Mutation{Root: ".verification/mutation"}},
				Executor: runtimeExecutor(filepath.Join(root, ".task")), mutationCampaign: test.run,
			}
			err := runner.Check(context.Background(), []string{"."})
			if !errors.Is(err, test.want) {
				t.Fatalf("Check() error = %v, want %v", err, test.want)
			}
		})
	}
}

type failingMutationFiles struct {
	mkdirErr   error
	writeErrAt int
	writes     int
	failure    error
}

func (files *failingMutationFiles) MkdirAll(string, os.FileMode) error { return files.mkdirErr }

func (files *failingMutationFiles) WriteFile(string, []byte, os.FileMode) error {
	files.writes++
	if files.writes == files.writeErrAt {
		return files.failure
	}
	return nil
}

func writeValidMutationSetup(t *testing.T, root string) {
	t.Helper()
	writeMutationFixture(t, root, ".", "module example\n")
	directory := filepath.Join(root, ".verification", "mutation")
	if err := os.MkdirAll(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "zero-inventory.json"), []byte("{\"schema_version\":1,\"packages\":[]}"), 0o600); err != nil {
		t.Fatal(err)
	}
}

func writeMutationFixture(t *testing.T, root, directory, mod string) {
	t.Helper()
	path := filepath.Join(root, directory)
	if err := os.MkdirAll(path, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(path, "go.mod"), []byte(mod), 0o600); err != nil {
		t.Fatal(err)
	}
}

func runtimeExecutor(directory string) Executor {
	return workspaceExecutor{directory: directory, run: func(_ context.Context, command Command) error {
		if reflect.DeepEqual(command.Args, []string{"env", "-json", "GOVERSION", "GOOS", "GOARCH", "CGO_ENABLED"}) {
			_, _ = io.WriteString(command.Stdout, `{"GOVERSION":"go1.27.0","GOOS":"linux","GOARCH":"amd64","CGO_ENABLED":"0"}`)
		}
		return nil
	}}
}
