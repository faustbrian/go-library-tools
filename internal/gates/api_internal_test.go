package gates

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/faustbrian/go-library-tools/internal/inventory"
)

func TestAPIModuleReportsFilesystemFailures(t *testing.T) {
	failure := errors.New("injected failure")
	tests := []struct {
		name   string
		update bool
		files  *fakeAPIFiles
		want   string
	}{
		{"mkdir", true, &fakeAPIFiles{mkdirErr: failure}, "create API baseline directory"},
		{"inspect directory", true, &fakeAPIFiles{lstatErr: failure}, "not a real directory"},
		{"create", false, &fakeAPIFiles{createErr: failure}, "create API snapshot"},
		{"close", false, &fakeAPIFiles{file: &fakeNamedFile{name: "snapshot", closeErr: failure}}, "close API snapshot"},
		{"rename", true, &fakeAPIFiles{file: &fakeNamedFile{name: "snapshot"}, renameErr: failure}, "publish API baseline"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			runner := Runner{Root: "/repo", Executor: executorFunction(func(context.Context, Command) error { return nil }), apiFiles: test.files}
			err := runner.apiModule(context.Background(), io.Discard, inventory.Module{Directory: ".", ModulePath: "example"}, test.update)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("apiModule() error = %v", err)
			}
		})
	}
}

func TestAPIModuleUsesTaskWorkspaceAndReportsDiagnostics(t *testing.T) {
	files := &fakeAPIFiles{file: &fakeNamedFile{name: "snapshot"}}
	failure := errors.New("failed")
	executor := workspaceExecutor{directory: "/task", run: func(_ context.Context, command Command) error {
		if command.Stderr != nil {
			_, _ = io.WriteString(command.Stderr, "diagnostic")
			return failure
		}
		return nil
	}}
	runner := Runner{Root: "/repo", Executor: executor, apiFiles: files}
	err := runner.apiModule(context.Background(), io.Discard, inventory.Module{Directory: ".", ModulePath: "example"}, false)
	if err == nil || !strings.Contains(err.Error(), "diagnostic") || files.createdDirectory != "/task" {
		t.Fatalf("apiModule() = %v, directory %q", err, files.createdDirectory)
	}
}

func TestAPIModuleRejectsUnboundedVerifierOutput(t *testing.T) {
	files := &fakeAPIFiles{file: &fakeNamedFile{name: "snapshot"}}
	executor := workspaceExecutor{directory: "/task", run: func(_ context.Context, command Command) error {
		if command.Stdout != nil {
			_, _ = command.Stdout.Write(make([]byte, maximumAPIOutput+1))
		}
		return nil
	}}
	runner := Runner{Root: "/repo", Executor: executor, apiFiles: files}
	err := runner.apiModule(context.Background(), io.Discard, inventory.Module{Directory: ".", ModulePath: "example"}, false)
	if err == nil || !strings.Contains(err.Error(), "exceeded") {
		t.Fatalf("apiModule() error = %v", err)
	}
}

func TestBoundedBufferPreservesExactBoundaryAndTruncatesOverflow(t *testing.T) {
	var buffer boundedBuffer
	first := bytes.Repeat([]byte{'a'}, maximumAPIOutput-1)
	if written, err := buffer.Write(first); err != nil || written != len(first) || buffer.overflow {
		t.Fatalf("first write = %d, %v, overflow = %v", written, err, buffer.overflow)
	}
	if written, err := buffer.Write([]byte{'b'}); err != nil || written != 1 || buffer.overflow || buffer.Len() != maximumAPIOutput {
		t.Fatalf("boundary write = %d, %v, overflow = %v, length = %d", written, err, buffer.overflow, buffer.Len())
	}
	if written, err := buffer.Write([]byte("cd")); err != nil || written != 2 || !buffer.overflow || buffer.Len() != maximumAPIOutput {
		t.Fatalf("overflow write = %d, %v, overflow = %v, length = %d", written, err, buffer.overflow, buffer.Len())
	}
	if got := buffer.Bytes()[maximumAPIOutput-1]; got != 'b' {
		t.Fatalf("last retained byte = %q", got)
	}
}

func TestAPIDisabledModuleDoesNotStopLaterEnabledModule(t *testing.T) {
	files := &fakeAPIFiles{file: &fakeNamedFile{name: "snapshot"}}
	runs := 0
	runner := Runner{
		Root: "/repo",
		Catalog: inventory.Inventory{Modules: []inventory.Module{
			{Directory: "a-disabled"},
			{Directory: "z-enabled", ModulePath: "example", Gates: map[string]bool{"api_compatibility": true}},
		}},
		Executor: executorFunction(func(context.Context, Command) error {
			runs++
			return nil
		}),
		apiFiles: files,
	}
	if err := runner.API(context.Background(), []string{"a-disabled", "z-enabled"}, false); err != nil {
		t.Fatalf("API() error = %v", err)
	}
	if runs != 2 {
		t.Fatalf("enabled API command runs = %d, want 2", runs)
	}
}

func TestBoundedBufferAcceptsContentWithinLimit(t *testing.T) {
	var buffer boundedBuffer
	if written, err := buffer.Write([]byte("ok")); err != nil || written != 2 || buffer.String() != "ok" || buffer.overflow {
		t.Fatalf("Write() = %d, %v, %q, %v", written, err, buffer.String(), buffer.overflow)
	}
}

type fakeAPIFiles struct {
	file             namedWriteCloser
	mkdirErr         error
	createErr        error
	renameErr        error
	createdDirectory string
	lstatErr         error
}

func (files *fakeAPIFiles) Lstat(path string) (os.FileInfo, error) {
	if files.lstatErr != nil {
		return nil, files.lstatErr
	}
	return apiFileInfo{directory: strings.HasSuffix(path, "/api")}, nil
}
func (files *fakeAPIFiles) MkdirAll(string, os.FileMode) error { return files.mkdirErr }
func (files *fakeAPIFiles) CreateTemp(directory, _ string) (namedWriteCloser, error) {
	files.createdDirectory = directory
	if files.file == nil {
		files.file = &fakeNamedFile{name: "snapshot"}
	}
	return files.file, files.createErr
}
func (files *fakeAPIFiles) Rename(string, string) error { return files.renameErr }
func (*fakeAPIFiles) Remove(string) error               { return nil }

type apiFileInfo struct{ directory bool }

func (apiFileInfo) Name() string { return "snapshot" }
func (apiFileInfo) Size() int64  { return 1 }
func (info apiFileInfo) Mode() os.FileMode {
	if info.directory {
		return os.ModeDir | 0o700
	}
	return 0o600
}
func (apiFileInfo) ModTime() time.Time { return time.Time{} }
func (info apiFileInfo) IsDir() bool   { return info.directory }
func (apiFileInfo) Sys() any           { return nil }

type workspaceExecutor struct {
	directory string
	run       func(context.Context, Command) error
}

func (executor workspaceExecutor) Run(ctx context.Context, command Command) error {
	return executor.run(ctx, command)
}
func (executor workspaceExecutor) TemporaryDirectory() string { return executor.directory }
