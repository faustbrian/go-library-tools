package gates

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

type processExecutor struct {
	environment map[string]string
	stdout      io.Writer
	stderr      io.Writer
	task        string
}

type taskFileSystem interface {
	MkdirTemp(string, string) (string, error)
	MkdirAll(string, os.FileMode) error
	Walk(string, filepath.WalkFunc) error
	Chmod(string, os.FileMode) error
	RemoveAll(string) error
}

type operatingTaskFileSystem struct{}

func (operatingTaskFileSystem) MkdirTemp(directory, pattern string) (string, error) {
	return os.MkdirTemp(directory, pattern)
}

func (operatingTaskFileSystem) MkdirAll(path string, mode os.FileMode) error {
	return os.MkdirAll(path, mode)
}

func (operatingTaskFileSystem) Walk(root string, visit filepath.WalkFunc) error {
	return filepath.Walk(root, visit)
}

func (operatingTaskFileSystem) Chmod(path string, mode os.FileMode) error {
	return os.Chmod(path, mode)
}

func (operatingTaskFileSystem) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

// NewProcessExecutor creates an executor whose Go caches and temporary files
// are task-owned. The caller must invoke cleanup exactly once.
func NewProcessExecutor(repositoryRoot string, stdout, stderr io.Writer) (Executor, func() error, error) {
	return newProcessExecutor(repositoryRoot, stdout, stderr, operatingTaskFileSystem{})
}

func newProcessExecutor(repositoryRoot string, stdout, stderr io.Writer, files taskFileSystem) (*processExecutor, func() error, error) {
	slug := strings.NewReplacer("/", "-", "\\", "-").Replace(filepath.Base(repositoryRoot))
	task, err := files.MkdirTemp("", "golib-"+slug+"-")
	if err != nil {
		return nil, nil, fmt.Errorf("create task workspace: %w", err)
	}
	cleanup := func() error {
		if err := files.Walk(task, func(path string, info os.FileInfo, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			return files.Chmod(path, info.Mode()|0o200)
		}); err != nil {
			return fmt.Errorf("make task workspace removable: %w", err)
		}
		if err := files.RemoveAll(task); err != nil {
			return fmt.Errorf("remove task workspace: %w", err)
		}
		return nil
	}
	directories := map[string]string{
		"GOCACHE":    filepath.Join(task, "go-build"),
		"GOMODCACHE": filepath.Join(task, "go-mod"),
		"GOTMPDIR":   filepath.Join(task, "go-tmp"),
		"GOBIN":      filepath.Join(task, "bin"),
	}
	for _, directory := range directories {
		if err := files.MkdirAll(directory, 0o700); err != nil {
			_ = cleanup()
			return nil, nil, fmt.Errorf("create task directory: %w", err)
		}
	}
	return &processExecutor{environment: directories, stdout: stdout, stderr: stderr, task: task}, cleanup, nil
}

func (executor *processExecutor) TemporaryDirectory() string { return executor.task }

func (executor *processExecutor) Run(ctx context.Context, command Command) error {
	process := exec.CommandContext(ctx, command.Name, command.Args...)
	process.Dir = command.Dir
	process.Env = mergeEnvironment(os.Environ(), executor.environment, command.Env)
	process.Stdout = command.Stdout
	if process.Stdout == nil {
		process.Stdout = executor.stdout
	}
	process.Stderr = command.Stderr
	if process.Stderr == nil {
		process.Stderr = executor.stderr
	}
	if err := process.Run(); err != nil {
		return fmt.Errorf("run %s: %w", command.Name, err)
	}
	return nil
}

func mergeEnvironment(current []string, sets ...map[string]string) []string {
	values := make(map[string]string, len(current))
	for _, entry := range current {
		key, value, found := strings.Cut(entry, "=")
		if found {
			values[key] = value
		}
	}
	for _, set := range sets {
		for key, value := range set {
			values[key] = value
		}
	}
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	result := make([]string, 0, len(keys))
	for _, key := range keys {
		result = append(result, key+"="+values[key])
	}
	return result
}
