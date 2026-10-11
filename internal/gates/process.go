package gates

import (
	"context"
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

type processExecutor struct {
	environment           map[string]string
	runtimeCaches         map[string]string
	borrowedCaches        map[string]string
	stdout                io.Writer
	stderr                io.Writer
	task                  string
	prepareBoundedProcess func(*exec.Cmd) error
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

// NewProcessExecutor creates an executor whose Go and lint caches and temporary
// files are task-owned. The caller must invoke cleanup exactly once.
func NewProcessExecutor(repositoryRoot string, stdout, stderr io.Writer) (Executor, func() error, error) {
	return newProcessExecutor(repositoryRoot, stdout, stderr, operatingTaskFileSystem{})
}

// NewProcessExecutorWithTaskCache borrows only compilation caches from a
// caller-owned bounded session. Its cleanup removes only invocation scratch;
// the caller must remove the session after use. Only explicitly classified
// native repository commands borrow it; verifier caches remain invocation-owned.
func NewProcessExecutorWithTaskCache(repositoryRoot, cacheRoot string, stdout, stderr io.Writer) (Executor, func() error, error) {
	directories, err := taskCacheDirectories(repositoryRoot, cacheRoot)
	if err != nil {
		return nil, nil, err
	}
	executor, cleanup, err := newProcessExecutor(repositoryRoot, stdout, stderr, operatingTaskFileSystem{})
	if err != nil {
		return nil, nil, err
	}
	executor.borrowedCaches = directories
	return executor, cleanup, nil
}

func taskCacheDirectories(repositoryRoot, root string) (map[string]string, error) {
	invalid := errors.New("task cache requires a private, canonical temporary session with existing private cache directories")
	if len(root) > 4096 || !filepath.IsAbs(root) || filepath.Clean(root) != root || !strings.HasPrefix(filepath.Base(root), "golib-cache-session.") {
		return nil, invalid
	}
	canonical, err := filepath.EvalSymlinks(root)
	if err != nil || canonical != root {
		return nil, invalid
	}
	temporary, err := filepath.EvalSymlinks(os.TempDir())
	if err != nil || filepath.Dir(root) != temporary {
		return nil, invalid
	}
	repository, err := filepath.EvalSymlinks(repositoryRoot)
	if err != nil {
		return nil, invalid
	}
	for _, pair := range [][2]string{{repository, root}, {root, repository}} {
		relative, err := filepath.Rel(pair[0], pair[1])
		if err != nil || relative == "." || relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			return nil, invalid
		}
	}
	directories := map[string]string{"GOCACHE": filepath.Join(root, "cache"), "GOMODCACHE": filepath.Join(root, "mod")}
	for _, path := range []string{root, directories["GOCACHE"], directories["GOMODCACHE"]} {
		info, err := os.Lstat(path)
		if err != nil || !info.IsDir() || info.Mode().Perm() != 0o700 {
			return nil, invalid
		}
	}
	return directories, nil
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
		"GOCACHE":             filepath.Join(task, "go-build"),
		"GOMODCACHE":          filepath.Join(task, "go-mod"),
		"GOTMPDIR":            filepath.Join(task, "go-tmp"),
		"GOBIN":               filepath.Join(task, "bin"),
		"GOLANGCI_LINT_CACHE": filepath.Join(task, "golangci-lint"),
	}
	runtimeCaches := map[string]string{
		"GOCACHE":             filepath.Join(task, "runtime-go-build"),
		"GOMODCACHE":          filepath.Join(task, "runtime-go-mod"),
		"GOLANGCI_LINT_CACHE": filepath.Join(task, "runtime-lint"),
	}
	for _, directory := range appendCacheDirectories(directories, runtimeCaches) {
		if err := files.MkdirAll(directory, 0o700); err != nil {
			_ = cleanup()
			return nil, nil, fmt.Errorf("create task directory: %w", err)
		}
	}
	// An external lint-cache backend can retain results outside owned scratch.
	directories["GOLANGCI_LINT_CACHEPROG"] = ""
	directories["GOCACHEPROG"] = ""
	directories["GOLIB_TASK_CACHE_ROOT"] = ""
	return &processExecutor{environment: directories, runtimeCaches: runtimeCaches, stdout: stdout, stderr: stderr, task: task}, cleanup, nil
}

func appendCacheDirectories(sets ...map[string]string) []string {
	var directories []string
	for _, set := range sets {
		for _, directory := range set {
			directories = append(directories, directory)
		}
	}
	return directories
}

// Seal owned cache identities after command overrides. Runtime tools can write
// their cache, so it must never become an input to a later verifier, even when
// no caller session was requested. Unknown producers cannot borrow a session.
func (executor *processExecutor) commandEnvironment(command Command) []string {
	protected := maps.Clone(executor.environment)
	// GOBIN selects an install output, not reusable compiler evidence. Keep the
	// owned default while allowing a caller to choose its owned output directory.
	delete(protected, "GOBIN")
	if command.cacheClass != verifierCache {
		maps.Copy(protected, executor.runtimeCaches)
		if command.cacheClass == nativeRepositoryCache {
			maps.Copy(protected, executor.borrowedCaches)
		}
		if command.cacheClass == nativeRepositoryCache && command.runtimeBuildCache != "" {
			protected["GOCACHE"] = command.runtimeBuildCache
		}
	}
	return mergeEnvironment(os.Environ(), executor.environment, command.Env, protected)
}

func (executor *processExecutor) TemporaryDirectory() string { return executor.task }

func (executor *processExecutor) Run(ctx context.Context, command Command) error {
	if command.boundedScanner {
		return executor.runBounded(ctx, command)
	}
	// #nosec G204 -- repository policy intentionally supplies argument-array gate commands without a shell
	process := exec.CommandContext(ctx, command.Name, command.Args...)
	process.Dir = command.Dir
	process.Env = executor.commandEnvironment(command)
	process.Stdin = command.Stdin
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

type overflowAwareOutput interface {
	io.Writer
	setOverflowCallback(func())
}

func (executor *processExecutor) runBounded(ctx context.Context, command Command) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	// #nosec G204 -- repository policy intentionally supplies argument-array gate commands without a shell
	//nolint:noctx // pinned scanners use owned cancellation of their original process group, not a sandbox
	process := exec.Command(command.Name, command.Args...)
	process.Dir = command.Dir
	process.Env = executor.commandEnvironment(command)
	process.Stdin = command.Stdin
	process.Stdout = command.Stdout
	process.Stderr = command.Stderr
	prepare := executor.prepareBoundedProcess
	if prepare == nil {
		prepare = configureProcessGroup
	}
	if err := prepare(process); err != nil {
		return fmt.Errorf("prepare bounded process: %w", err)
	}

	stop := make(chan struct{})
	var stopOnce sync.Once
	signalStop := func() { stopOnce.Do(func() { close(stop) }) }
	for _, stream := range []io.Writer{command.Stdout, command.Stderr} {
		if bounded, ok := stream.(overflowAwareOutput); ok {
			bounded.setOverflowCallback(signalStop)
		}
	}
	if err := process.Start(); err != nil {
		return fmt.Errorf("run %s: %w", command.Name, err)
	}
	done := make(chan struct{})
	monitorDone := make(chan struct{})
	go func() {
		defer close(monitorDone)
		select {
		case <-ctx.Done():
			terminateProcessGroup(process)
		case <-stop:
			terminateProcessGroup(process)
		case <-done:
		}
	}()
	err := process.Wait()
	close(done)
	<-monitorDone
	return boundedProcessResult(ctx, command.Name, err)
}

func boundedProcessResult(ctx context.Context, name string, err error) error {
	if contextError := ctx.Err(); contextError != nil {
		err = errors.Join(contextError, err)
	}
	if err != nil {
		return fmt.Errorf("run %s: %w", name, err)
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
		maps.Copy(values, set)
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
