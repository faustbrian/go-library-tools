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

	"github.com/faustbrian/go-library-tools/internal/inventory"
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
