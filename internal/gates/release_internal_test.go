package gates

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/faustbrian/go-library-tools/v2/internal/config"
	"github.com/faustbrian/go-library-tools/v2/internal/inventory"
	"golang.org/x/mod/module"
)

func TestReleaseRehearsalRejectsExistingTag(t *testing.T) {
	root := releaseFixture(t)
	executor := workspaceExecutor{directory: t.TempDir(), run: func(_ context.Context, command Command) error {
		if command.Name != "git" || command.Stdout == nil {
			t.Fatalf("command = %#v", command)
		}
		_, _ = io.WriteString(command.Stdout, "v1.0.0\n")
		return nil
	}}
	runner := Runner{Root: root, Executor: executor}
	second := releaseModule()
	second.Directory = "nested"
	second.ModulePath += "/nested"
	second.TagPrefix = "nested/v"
	err := runner.releaseRehearsal(context.Background(), []inventory.Module{second, releaseModule()})
	if err == nil || !strings.Contains(err.Error(), "release tag already exists") {
		t.Fatalf("releaseRehearsal() error = %v", err)
	}
}

func TestReleaseRehearsalBuildsAndConsumesLocalProxy(t *testing.T) {
	root := releaseFixture(t)
	workspace := t.TempDir()
	var commands []Command
	executor := workspaceExecutor{directory: workspace, run: func(_ context.Context, command Command) error {
		commands = append(commands, command)
		if command.Name == "go" && len(command.Args) > 0 && command.Args[0] == "list" {
			proxy := strings.TrimPrefix(command.Env["GOPROXY"], "file://")
			archive := filepath.Join(proxy, "example.com", "library", "@v", "v1.0.0.zip")
			reader, err := zip.OpenReader(archive)
			if err != nil {
				return err
			}
			defer reader.Close()
			if len(reader.File) != 2 {
				return errors.New("release archive has unexpected file count")
			}
		}
		return nil
	}}
	runner := Runner{Root: root, Executor: executor}
	if err := runner.releaseRehearsal(context.Background(), []inventory.Module{releaseModule()}); err != nil {
		t.Fatalf("releaseRehearsal() error = %v", err)
	}
	if len(commands) != 2 || commands[0].Name != "git" || commands[1].Name != "go" {
		t.Fatalf("commands = %#v", commands)
	}
	if got := strings.Join(commands[1].Args, " "); got != "list -m example.com/library@v1.0.0" {
		t.Fatalf("go command = %q", got)
	}
	if commands[1].Env["GOWORK"] != "off" || commands[1].Env["GOSUMDB"] != "off" {
		t.Fatalf("go environment = %#v", commands[1].Env)
	}
}

func TestReleaseCandidateInstallsV2CommandFromLocalProxy(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	initialModuleFiles := make(map[string][]byte, 2)
	for _, name := range []string{"go.mod", "go.sum"} {
		initialModuleFiles[name], err = os.ReadFile(filepath.Join(root, name))
		if err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(func() {
		for name, initial := range initialModuleFiles {
			current, readErr := os.ReadFile(filepath.Join(root, name))
			if readErr != nil || !bytes.Equal(current, initial) {
				t.Errorf("release proxy test changed root %s: %v", name, readErr)
			}
		}
	})
	policy, err := config.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := inventory.Load(root, policy)
	if err != nil {
		t.Fatal(err)
	}
	if len(catalog.Modules) != 1 {
		t.Fatalf("module count = %d, want 1", len(catalog.Modules))
	}
	candidate := catalog.Modules[0]
	executor, cleanup, err := NewProcessExecutor(root, io.Discard, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := cleanup(); err != nil {
			t.Errorf("cleanup: %v", err)
		}
	})
	workspaceOwner, ok := executor.(taskWorkspace)
	if !ok {
		t.Fatalf("executor type = %T, want task workspace", executor)
	}
	workspace := workspaceOwner.TemporaryDirectory()
	seed := filepath.Join(workspace, "dependency-seed")
	if err := os.Mkdir(seed, 0o700); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"go.mod", "go.sum"} {
		if writeErr := os.WriteFile(filepath.Join(seed, name), initialModuleFiles[name], 0o600); writeErr != nil {
			t.Fatal(writeErr)
		}
	}
	if err := executor.Run(t.Context(), Command{Name: "go", Args: []string{"mod", "download", "all"}, Dir: seed}); err != nil {
		t.Fatalf("seed dependency proxy: %v", err)
	}
	proxy, err := (Runner{Root: root}).buildReleaseProxy(workspace, []inventory.Module{candidate})
	if err != nil {
		t.Fatalf("build release proxy: %v", err)
	}

	positiveBin := filepath.Join(workspace, "positive-bin")
	if err := os.Mkdir(positiveBin, 0o700); err != nil {
		t.Fatal(err)
	}
	positive := Command{
		Name: "go", Args: []string{"install", "github.com/faustbrian/go-library-tools/v2/cmd/golib@v2.0.0"},
		Dir: workspace, Env: map[string]string{
			"GOBIN": positiveBin, "GOPROXY": "file://" + proxy, "GOSUMDB": "off", "GOWORK": "off",
		},
	}
	if err := executor.Run(t.Context(), positive); err != nil {
		t.Fatalf("install v2 command: %v", err)
	}
	entries, err := os.ReadDir(positiveBin)
	if err != nil || len(entries) != 1 || entries[0].Name() != "golib" {
		t.Fatalf("installed commands = %#v, %v", entries, err)
	}

	negativeBin := filepath.Join(workspace, "negative-bin")
	if err := os.Mkdir(negativeBin, 0o700); err != nil {
		t.Fatal(err)
	}
	negative := positive
	negative.Args = []string{"install", "github.com/faustbrian/go-library-tools/cmd/golib@v2.0.0"}
	negative.Env = map[string]string{
		"GOBIN": negativeBin, "GOPROXY": "file://" + proxy, "GOSUMDB": "off", "GOWORK": "off",
	}
	if err := executor.Run(t.Context(), negative); err == nil {
		t.Fatal("installing the unsuffixed v2 command succeeded")
	}
	entries, err = os.ReadDir(negativeBin)
	if err != nil || len(entries) != 0 {
		t.Fatalf("unsuffixed install produced commands = %#v, %v", entries, err)
	}
}

func TestReleaseDryRunRoutesRehearsalBeforeGates(t *testing.T) {
	root := releaseFixture(t)
	module := releaseModule()
	runner := Runner{
		Root: root, Catalog: inventory.Inventory{Modules: []inventory.Module{module}},
		Executor: workspaceExecutor{directory: t.TempDir(), run: func(context.Context, Command) error { return nil }},
	}
	if err := runner.ReleaseDryRun(context.Background(), []string{"missing"}); err == nil {
		t.Fatal("ReleaseDryRun(unknown module) error = nil")
	}
	if err := runner.ReleaseDryRun(context.Background(), []string{"."}); err != nil {
		t.Fatalf("ReleaseDryRun() error = %v", err)
	}
	runner.Executor = executorFunction(func(context.Context, Command) error { return nil })
	if err := runner.ReleaseDryRun(context.Background(), []string{"."}); err == nil {
		t.Fatal("ReleaseDryRun(rehearsal failure) error = nil")
	}
}

func TestReleaseDryRunRejectsInvalidCompactSuppressionAfterRehearsal(t *testing.T) {
	root := releaseFixture(t)
	releaseWrite(t, filepath.Join(root, "library.go"), "package library\n/*#nosec*/\nfunc value() {}\n")
	module := releaseModule()
	module.Gates = map[string]bool{"security": true}
	commands := make([]Command, 0)
	runner := Runner{
		Root: root, Catalog: inventory.Inventory{Modules: []inventory.Module{module}},
		Executor: workspaceExecutor{directory: t.TempDir(), run: func(_ context.Context, command Command) error {
			commands = append(commands, command)
			return nil
		}},
	}
	err := runner.ReleaseDryRun(context.Background(), []string{"."})
	if err == nil || !strings.Contains(err.Error(), "library.go:2:") || !strings.Contains(err.Error(), "exact rule IDs") {
		t.Fatalf("ReleaseDryRun() error = %v", err)
	}
	if len(commands) < 2 || commands[0].Name != "git" || strings.Join(commands[0].Args, " ") != "tag --list v1.0.0" ||
		commands[1].Name != "go" || strings.Join(commands[1].Args, " ") != "list -m example.com/library@v1.0.0" {
		t.Fatalf("rehearsal command prefix = %#v", commands)
	}
	for _, command := range commands[2:] {
		joined := strings.Join(command.Args, " ")
		for _, tool := range []string{"govulncheck", "securego/gosec", "go-analysis", "gitleaks", "go-licenses", "cyclonedx-gomod"} {
			if strings.Contains(joined, tool) {
				t.Fatalf("security command ran after suppression failure: %#v", command)
			}
		}
	}
}

func TestReleaseDryRunCompletesRehearsalBeforeBatchSuppressionPreflight(t *testing.T) {
	root := t.TempDir()
	for _, directory := range []string{"a-valid", "z-invalid"} {
		if err := os.MkdirAll(filepath.Join(root, directory), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	releaseWrite(t, filepath.Join(root, "a-valid", "go.mod"), "module example.com/library/a\n\ngo 1.27.0\n")
	releaseWrite(t, filepath.Join(root, "a-valid", "valid.go"), "package a\nfunc value() {}\n")
	releaseWrite(t, filepath.Join(root, "z-invalid", "go.mod"), "module example.com/library/z\n\ngo 1.27.0\n")
	releaseWrite(t, filepath.Join(root, "z-invalid", "invalid.go"), "package z\n//#nosec\nfunc value() {}\n")
	rootModule := releaseModule()
	rootModule.Directory = "a-valid"
	rootModule.ModulePath = "example.com/library/a"
	rootModule.TagPrefix = "a/v"
	rootModule.Gates = map[string]bool{"security": true}
	rootModule.RequiredServices = []string{"postgresql"}
	nestedModule := releaseModule()
	nestedModule.Directory = "z-invalid"
	nestedModule.ModulePath = "example.com/library/z"
	nestedModule.TagPrefix = "z/v"
	nestedModule.Gates = map[string]bool{"security": true}
	nestedModule.RequiredServices = []string{"valkey"}
	var commands []Command
	starts := 0
	var output bytes.Buffer
	runner := Runner{
		Root: root, Catalog: inventory.Inventory{Modules: []inventory.Module{nestedModule, rootModule}}, Output: &output,
		Executor: workspaceExecutor{directory: t.TempDir(), run: func(_ context.Context, command Command) error {
			commands = append(commands, command)
			return nil
		}},
		startServices: func(context.Context, []string) (serviceLease, error) {
			starts++
			return &fakeServiceLease{}, nil
		},
	}
	err := runner.ReleaseDryRun(context.Background(), []string{"z-invalid", "a-valid", "z-invalid"})
	if err == nil || !strings.Contains(err.Error(), "invalid.go:2:") || !strings.Contains(err.Error(), "exact rule IDs") {
		t.Fatalf("ReleaseDryRun() error = %v", err)
	}
	want := []string{
		"git tag --list a/v1.0.0",
		"git tag --list z/v1.0.0",
		"go list -m example.com/library/a@v1.0.0",
		"go list -m example.com/library/z@v1.0.0",
	}
	if len(commands) != len(want) {
		t.Fatalf("release commands = %#v, want rehearsal only", commands)
	}
	for index, command := range commands {
		got := command.Name + " " + strings.Join(command.Args, " ")
		if got != want[index] {
			t.Fatalf("release command %d = %q, want %q", index, got, want[index])
		}
	}
	if output.String() != "[a-valid] security-suppressions\n[z-invalid] security-suppressions\n" || starts != 0 {
		t.Fatalf("release preflight output/starts = %q/%d", output.String(), starts)
	}
}

func TestReleaseRehearsalFailsClosed(t *testing.T) {
	root := releaseFixture(t)
	failure := errors.New("injected failure")
	relative := Runner{Root: root, Executor: workspaceExecutor{
		directory: "relative", run: func(context.Context, Command) error { return nil },
	}}
	if err := relative.releaseRehearsal(context.Background(), []inventory.Module{releaseModule()}); err == nil ||
		!strings.Contains(err.Error(), "task-owned workspace") {
		t.Fatalf("releaseRehearsal(relative workspace) error = %v", err)
	}
	tests := map[string]Runner{
		"workspace": {Root: root, Executor: executorFunction(func(context.Context, Command) error { return nil })},
		"tag query": {Root: root, Executor: workspaceExecutor{directory: t.TempDir(), run: func(context.Context, Command) error {
			return failure
		}}},
		"proxy directory": {Root: root, Executor: workspaceExecutor{directory: filepath.Join(t.TempDir(), "missing", "task"), run: func(context.Context, Command) error {
			return nil
		}}},
	}
	for name, runner := range tests {
		t.Run(name, func(t *testing.T) {
			if err := runner.releaseRehearsal(context.Background(), []inventory.Module{releaseModule()}); err == nil {
				t.Fatalf("releaseRehearsal(%s) error = nil", name)
			}
		})
	}

	executor := workspaceExecutor{directory: t.TempDir(), run: func(_ context.Context, command Command) error {
		if command.Name == "go" {
			return failure
		}
		return nil
	}}
	if err := (Runner{Root: root, Executor: executor}).releaseRehearsal(context.Background(), []inventory.Module{releaseModule()}); !errors.Is(err, failure) {
		t.Fatalf("releaseRehearsal(go failure) error = %v", err)
	}

	badRoot := t.TempDir()
	releaseWrite(t, filepath.Join(badRoot, "go.mod"), "module example.com/library\n\ngo 1.27.0\n")
	if err := os.WriteFile(filepath.Join(badRoot, "bad"), bytes.Repeat([]byte{'x'}, 1), 0); err != nil {
		t.Fatal(err)
	}
	if err := (Runner{Root: badRoot, Executor: executor}).releaseRehearsal(context.Background(), []inventory.Module{releaseModule()}); err == nil {
		t.Fatal("releaseRehearsal(invalid source) error = nil")
	}
}

func TestBuildReleaseProxyFailsClosed(t *testing.T) {
	root := releaseFixture(t)
	failure := errors.New("injected failure")
	for _, stage := range []string{"temporary", "directory", "module", "information", "list", "archive", "close"} {
		t.Run(stage, func(t *testing.T) {
			runner := Runner{Root: root, releaseFiles: controlledReleaseFiles{stage: stage, failure: failure}}
			if _, err := runner.buildReleaseProxy(t.TempDir(), []inventory.Module{releaseModule()}); err == nil {
				t.Fatalf("buildReleaseProxy(%s) error = nil", stage)
			}
		})
	}
	for name, module := range map[string]inventory.Module{
		"path":    {Directory: ".", ModulePath: "bad path", Version: "1.0.0"},
		"version": {Directory: ".", ModulePath: "example.com/library", Version: "!"},
		"module":  {Directory: "missing", ModulePath: "example.com/library", Version: "1.0.0"},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := (Runner{Root: root}).buildReleaseProxy(t.TempDir(), []inventory.Module{module}); err == nil {
				t.Fatalf("buildReleaseProxy(%s) error = nil", name)
			}
		})
	}
	runner := Runner{
		Root: root,
		releaseArchive: func(io.Writer, module.Version, string) error {
			return failure
		},
	}
	if _, err := runner.buildReleaseProxy(t.TempDir(), []inventory.Module{releaseModule()}); !errors.Is(err, failure) {
		t.Fatalf("buildReleaseProxy(archive failure) error = %v", err)
	}
}

func releaseFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	releaseWrite(t, filepath.Join(root, "go.mod"), "module example.com/library\n\ngo 1.27.0\n")
	releaseWrite(t, filepath.Join(root, "library.go"), "package library\n")
	return root
}

func releaseWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func releaseModule() inventory.Module {
	return inventory.Module{
		Directory: ".", ModulePath: "example.com/library", Version: "1.0.0", TagPrefix: "v",
	}
}

type controlledReleaseFiles struct {
	stage   string
	failure error
}

func (files controlledReleaseFiles) MkdirTemp(directory, pattern string) (string, error) {
	if files.stage == "temporary" {
		return "", files.failure
	}
	return os.MkdirTemp(directory, pattern)
}

func (files controlledReleaseFiles) MkdirAll(path string, mode os.FileMode) error {
	if files.stage == "directory" {
		return files.failure
	}
	return os.MkdirAll(path, mode)
}

func (files controlledReleaseFiles) WriteFile(path string, data []byte, mode os.FileMode) error {
	suffix := map[string]string{"module": ".mod", "information": ".info", "list": string(filepath.Separator) + "list"}[files.stage]
	if suffix != "" && strings.HasSuffix(path, suffix) {
		return files.failure
	}
	return os.WriteFile(path, data, mode)
}

func (files controlledReleaseFiles) Create(path string) (io.WriteCloser, error) {
	if files.stage == "archive" {
		return nil, files.failure
	}
	file, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	if files.stage == "close" {
		return closeFailureFile{File: file, failure: files.failure}, nil
	}
	return file, nil
}

type closeFailureFile struct {
	*os.File
	failure error
}

func (file closeFailureFile) Close() error {
	_ = file.File.Close()
	return file.failure
}
