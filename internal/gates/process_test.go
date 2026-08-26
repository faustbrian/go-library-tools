package gates

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestProcessExecutorUsesIsolatedTaskEnvironmentAndCleansIt(t *testing.T) {
	var stdout, stderr bytes.Buffer
	created, cleanup, err := NewProcessExecutor(t.TempDir(), &stdout, &stderr)
	if err != nil {
		t.Fatal(err)
	}
	executor := created.(*processExecutor)
	task := filepath.Dir(executor.environment["GOCACHE"])
	if executor.TemporaryDirectory() != task {
		t.Fatalf("TemporaryDirectory() = %q, want %q", executor.TemporaryDirectory(), task)
	}
	for _, name := range []string{"GOCACHE", "GOMODCACHE", "GOTMPDIR", "GOBIN"} {
		if !strings.HasPrefix(executor.environment[name], task+string(filepath.Separator)) {
			t.Fatalf("%s is not task-owned: %s", name, executor.environment[name])
		}
	}

	err = executor.Run(context.Background(), Command{
		Name: os.Args[0],
		Args: []string{"-test.run=TestProcessHelper", "--"},
		Env:  map[string]string{"GO_WANT_HELPER": "1", "HELPER_OUTPUT": "expected"},
	})
	if err != nil || stdout.String() != "expected" || stderr.Len() != 0 {
		t.Fatalf("Run() error = %v, stdout = %q, stderr = %q", err, stdout.String(), stderr.String())
	}
	if err := cleanup(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(task); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("task workspace remains: %v", err)
	}
}

func TestProcessExecutorReportsCommandFailure(t *testing.T) {
	created, cleanup, err := NewProcessExecutor(t.TempDir(), &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	executor := created.(*processExecutor)
	err = executor.Run(context.Background(), Command{
		Name: os.Args[0],
		Args: []string{"-test.run=TestProcessHelper", "--"},
		Env:  map[string]string{"GO_WANT_HELPER": "1", "HELPER_FAIL": "1"},
	})
	if err == nil || !strings.Contains(err.Error(), "run ") {
		t.Fatalf("Run() error = %v", err)
	}
}

func TestProcessExecutorHonorsCommandOutputOverrides(t *testing.T) {
	created, cleanup, err := NewProcessExecutor(t.TempDir(), io.Discard, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	var stdout, stderr bytes.Buffer
	err = created.Run(context.Background(), Command{
		Name: os.Args[0], Args: []string{"-test.run=TestProcessHelper", "--"},
		Env:    map[string]string{"GO_WANT_HELPER": "1", "HELPER_OUTPUT": "expected"},
		Stdout: &stdout, Stderr: &stderr,
	})
	if err != nil || stdout.String() != "expected" || stderr.Len() != 0 {
		t.Fatalf("Run() = %v, %q, %q", err, stdout.String(), stderr.String())
	}
}

func TestMergeEnvironmentIsSortedAndLastWriterWins(t *testing.T) {
	got := mergeEnvironment([]string{"B=old", "INVALID", "A=one"}, map[string]string{"B": "new", "C": "three"})
	want := []string{"A=one", "B=new", "C=three"}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("mergeEnvironment() = %#v", got)
	}
}

func TestNewProcessExecutorReportsTaskFileSystemFailures(t *testing.T) {
	failure := errors.New("injected failure")
	tests := []struct {
		name  string
		files *fakeTaskFileSystem
		want  string
	}{
		{"temporary root", &fakeTaskFileSystem{mkdirTempErr: failure}, "create task workspace"},
		{"cache directory", &fakeTaskFileSystem{mkdirAllErr: failure}, "create task directory"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, _, err := newProcessExecutor("/repo", &bytes.Buffer{}, &bytes.Buffer{}, test.files)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("newProcessExecutor() error = %v", err)
			}
		})
	}
}

func TestProcessExecutorCleanupReportsWalkAndRemoveFailures(t *testing.T) {
	failure := errors.New("injected failure")
	tests := []struct {
		name  string
		files *fakeTaskFileSystem
		want  string
	}{
		{"walk callback", &fakeTaskFileSystem{walkCallbackErr: failure}, "make task workspace removable"},
		{"walk", &fakeTaskFileSystem{walkErr: failure}, "make task workspace removable"},
		{"remove", &fakeTaskFileSystem{removeErr: failure}, "remove task workspace"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, cleanup, err := newProcessExecutor("/repo", &bytes.Buffer{}, &bytes.Buffer{}, test.files)
			if err != nil {
				t.Fatal(err)
			}
			if err := cleanup(); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("cleanup() error = %v", err)
			}
		})
	}
}

func TestProcessHelper(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER") != "1" {
		return
	}
	if os.Getenv("HELPER_FAIL") == "1" {
		os.Exit(23)
	}
	_, _ = os.Stdout.WriteString(os.Getenv("HELPER_OUTPUT"))
	os.Exit(0)
}

type fakeTaskFileSystem struct {
	mkdirTempErr    error
	mkdirAllErr     error
	walkCallbackErr error
	walkErr         error
	removeErr       error
}

func (fake *fakeTaskFileSystem) MkdirTemp(string, string) (string, error) {
	return "/task", fake.mkdirTempErr
}

func (fake *fakeTaskFileSystem) MkdirAll(string, os.FileMode) error {
	return fake.mkdirAllErr
}

func (fake *fakeTaskFileSystem) Walk(root string, visit filepath.WalkFunc) error {
	if fake.walkErr != nil {
		return fake.walkErr
	}
	if fake.walkCallbackErr != nil {
		return visit(root, nil, fake.walkCallbackErr)
	}
	return visit(root, fakeInfoForTask{}, nil)
}

func (*fakeTaskFileSystem) Chmod(string, os.FileMode) error { return nil }
func (fake *fakeTaskFileSystem) RemoveAll(string) error     { return fake.removeErr }

type fakeInfoForTask struct{}

func (fakeInfoForTask) Name() string       { return "task" }
func (fakeInfoForTask) Size() int64        { return 0 }
func (fakeInfoForTask) Mode() os.FileMode  { return os.ModeDir | 0o700 }
func (fakeInfoForTask) ModTime() time.Time { return time.Time{} }
func (fakeInfoForTask) IsDir() bool        { return true }
func (fakeInfoForTask) Sys() any           { return nil }
