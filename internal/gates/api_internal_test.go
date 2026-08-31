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

	"github.com/faustbrian/go-library-tools/internal/config"
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
		{"mkdir", true, &fakeAPIFiles{mkdirErr: failure, missingDirectory: true}, "create API baseline directory"},
		{"inspect directory", true, &fakeAPIFiles{lstatErr: failure}, "not a real directory"},
		{"nil directory", true, &fakeAPIFiles{nilInfo: true}, "not a real directory"},
		{"nil baseline", false, &fakeAPIFiles{nilInfo: true}, "missing API baseline"},
		{"create", false, &fakeAPIFiles{createErr: failure}, "create API snapshot"},
		{"close", false, &fakeAPIFiles{file: &fakeNamedFile{name: "snapshot", closeErr: failure}}, "close API snapshot"},
		{"nil snapshot", false, &fakeAPIFiles{nilSnapshotInfo: true}, "is empty"},
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

func TestAPIModulePinsDeclaredGoToolchain(t *testing.T) {
	files := &fakeAPIFiles{file: &fakeNamedFile{name: "snapshot"}}
	var commands []Command
	runner := Runner{
		Root: "/repo",
		Executor: workspaceExecutor{directory: "/task", run: func(_ context.Context, command Command) error {
			commands = append(commands, command)
			return nil
		}},
		apiFiles: files,
	}
	module := inventory.Module{Directory: ".", ModulePath: "example", GoVersion: "1.26.6"}
	if err := runner.apiModule(context.Background(), io.Discard, module, false); err != nil {
		t.Fatalf("apiModule() error = %v", err)
	}
	if len(commands) != 2 {
		t.Fatalf("API command count = %d, want 2", len(commands))
	}
	for index, command := range commands {
		if got := command.Env["GOTOOLCHAIN"]; got != "go1.26.6" {
			t.Errorf("command %d GOTOOLCHAIN = %q, want go1.26.6", index, got)
		}
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
		Policy: config.Config{API: config.API{Baselines: []config.APIBaseline{{
			Module: "z-enabled", Mode: "apidiff", Path: "api/baseline.txt",
		}}}},
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

func TestAPIBaselineDirectoryMustRemainBelowRoot(t *testing.T) {
	files := &fakeAPIFiles{}
	if err := ensureAPIBaselineDirectory(files, "/repo", ".", false); err != nil {
		t.Fatalf("ensureAPIBaselineDirectory(root) error = %v", err)
	}
	if err := ensureAPIBaselineDirectory(files, "/repo", "../outside", false); err == nil || !strings.Contains(err.Error(), "below repository root") {
		t.Fatalf("ensureAPIBaselineDirectory(outside) error = %v", err)
	}
	if err := ensureAPIBaselineDirectory(files, "/repo", "/absolute", false); err == nil || !strings.Contains(err.Error(), "below repository root") {
		t.Fatalf("ensureAPIBaselineDirectory(absolute) error = %v", err)
	}
	missing := &fakeAPIFiles{missingDirectory: true}
	if err := ensureAPIBaselineDirectory(missing, "/repo", "api", false); err == nil || missing.directoryCreated {
		t.Fatalf("ensureAPIBaselineDirectory(missing) error = %v, created = %v", err, missing.directoryCreated)
	}
}

func TestGoDocAPISnapshotReportsGenerationFailures(t *testing.T) {
	failure := errors.New("injected failure")
	tests := []struct {
		name     string
		executor Executor
		file     *fakeNamedFile
		want     string
	}{
		{"command", executorFunction(func(context.Context, Command) error { return failure }), &fakeNamedFile{name: "snapshot"}, "generate API documentation"},
		{"overflow", executorFunction(func(_ context.Context, command Command) error {
			_, _ = command.Stdout.Write(make([]byte, maximumAPIOutput+1))
			return nil
		}), &fakeNamedFile{name: "snapshot"}, "exceeded"},
		{"empty", executorFunction(func(context.Context, Command) error { return nil }), &fakeNamedFile{name: "snapshot"}, "is empty"},
		{"write", executorFunction(func(_ context.Context, command Command) error {
			_, _ = io.WriteString(command.Stdout, "API")
			return nil
		}), &fakeNamedFile{name: "snapshot", writeErr: failure}, "write API documentation"},
		{"short write", executorFunction(func(_ context.Context, command Command) error {
			_, _ = io.WriteString(command.Stdout, "API")
			return nil
		}), &fakeNamedFile{name: "snapshot", writeN: 1}, "short write"},
		{"close", executorFunction(func(_ context.Context, command Command) error {
			_, _ = io.WriteString(command.Stdout, "API")
			return nil
		}), &fakeNamedFile{name: "snapshot", closeErr: failure}, "close API snapshot"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			runner := Runner{Executor: test.executor}
			_, err := runner.generateAPISnapshot(context.Background(), "/repo", test.file, test.file.name,
				inventory.Module{Directory: ".", ModulePath: "example"},
				config.APIBaseline{Mode: "go-doc", Packages: []string{"."}})
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("generateAPISnapshot() error = %v", err)
			}
		})
	}
}

func TestGoDocAPISnapshotPinsDeclaredGoToolchain(t *testing.T) {
	file := &fakeNamedFile{name: "snapshot"}
	runner := Runner{Executor: executorFunction(func(_ context.Context, command Command) error {
		if got := command.Env["GOTOOLCHAIN"]; got != "go1.26.6" {
			t.Fatalf("GOTOOLCHAIN = %q, want go1.26.6", got)
		}
		_, _ = io.WriteString(command.Stdout, "API")
		return nil
	})}
	module := inventory.Module{Directory: ".", ModulePath: "example", GoVersion: "1.26.6"}
	if _, err := runner.generateAPISnapshot(context.Background(), "/repo", file, file.name, module,
		config.APIBaseline{Mode: "go-doc", Packages: []string{"."}}); err != nil {
		t.Fatalf("generateAPISnapshot() error = %v", err)
	}
}

func TestGoDocAPIModuleRejectsUnreadableAndDifferentBaselines(t *testing.T) {
	failure := errors.New("injected failure")
	for name, test := range map[string]struct {
		read func(string, string, int64) ([]byte, error)
		want string
	}{
		"read": {read: func(string, string, int64) ([]byte, error) { return nil, failure }, want: "read API baseline"},
		"different": {read: func(string, string, int64) ([]byte, error) {
			return []byte("different\n"), nil
		}, want: "documentation differs"},
	} {
		t.Run(name, func(t *testing.T) {
			runner := Runner{
				Root: "/repo",
				Policy: config.Config{API: config.API{Baselines: []config.APIBaseline{{
					Module: ".", Mode: "go-doc", Path: "api/v1.txt", Packages: []string{"."},
				}}}},
				Executor: executorFunction(func(_ context.Context, command Command) error {
					_, _ = io.WriteString(command.Stdout, "API")
					return nil
				}),
				apiFiles: &fakeAPIFiles{}, apiReadBaseline: test.read,
			}
			err := runner.apiModule(context.Background(), io.Discard, inventory.Module{Directory: ".", ModulePath: "example"}, false)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("apiModule() error = %v", err)
			}
		})
	}
}

type fakeAPIFiles struct {
	file             namedWriteCloser
	mkdirErr         error
	createErr        error
	renameErr        error
	createdDirectory string
	lstatErr         error
	missingDirectory bool
	directoryCreated bool
	nilInfo          bool
	nilSnapshotInfo  bool
}

func (files *fakeAPIFiles) Lstat(path string) (os.FileInfo, error) {
	if files.lstatErr != nil {
		return nil, files.lstatErr
	}
	if files.nilInfo {
		return nil, nil
	}
	if files.nilSnapshotInfo && path == "snapshot" {
		return nil, nil
	}
	if files.missingDirectory && strings.HasSuffix(path, "/api") && !files.directoryCreated {
		return nil, os.ErrNotExist
	}
	return apiFileInfo{directory: strings.HasSuffix(path, "/api")}, nil
}
func (files *fakeAPIFiles) MkdirAll(string, os.FileMode) error {
	if files.mkdirErr == nil {
		files.directoryCreated = true
	}
	return files.mkdirErr
}
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
