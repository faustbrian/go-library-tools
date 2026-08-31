package gates

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/faustbrian/go-library-tools/internal/config"
	"github.com/faustbrian/go-library-tools/internal/inventory"
	"github.com/faustbrian/go-library-tools/internal/repositoryfile"
)

const compatibilityToolVersion = "v0.0.0-20260718201538-764159d718ef"
const maximumAPIOutput = 4 << 20

type boundedBuffer struct {
	bytes.Buffer
	overflow bool
}

func (buffer *boundedBuffer) Write(value []byte) (int, error) {
	remaining := maximumAPIOutput - buffer.Len()
	if len(value) <= remaining {
		return buffer.Buffer.Write(value)
	}
	buffer.overflow = true
	_, _ = buffer.Buffer.Write(value[:remaining])
	return len(value), nil
}

type apiFileSystem interface {
	Lstat(string) (os.FileInfo, error)
	MkdirAll(string, os.FileMode) error
	CreateTemp(string, string) (namedWriteCloser, error)
	Rename(string, string) error
	Remove(string) error
}

type operatingAPIFiles struct{}

func (operatingAPIFiles) Lstat(path string) (os.FileInfo, error) { return os.Lstat(path) }
func (operatingAPIFiles) MkdirAll(path string, mode os.FileMode) error {
	return os.MkdirAll(path, mode)
}
func (operatingAPIFiles) CreateTemp(directory, pattern string) (namedWriteCloser, error) {
	return os.CreateTemp(directory, pattern)
}
func (operatingAPIFiles) Rename(oldPath, newPath string) error { return os.Rename(oldPath, newPath) }
func (operatingAPIFiles) Remove(path string) error             { return os.Remove(path) }

// API checks or updates exported API baselines for selected applicable modules.
func (runner Runner) API(ctx context.Context, selection []string, update bool) error {
	modules, err := runner.selectModules(selection)
	if err != nil {
		return err
	}
	output := runner.Output
	if output == nil {
		output = io.Discard
	}
	for _, module := range modules {
		if !module.Gates["api_compatibility"] {
			_, _ = fmt.Fprintf(output, "[%s] API compatibility: not applicable\n", module.Directory)
			continue
		}
		if err := runner.apiModule(ctx, output, module, update); err != nil {
			return err
		}
	}
	return nil
}

func (runner Runner) apiModule(ctx context.Context, output io.Writer, module inventory.Module, update bool) error {
	files := runner.apiFiles
	if files == nil {
		files = operatingAPIFiles{}
	}
	directory := filepath.Join(runner.Root, module.Directory)
	policy := runner.apiBaseline(module)
	baseline := filepath.Join(runner.Root, policy.Path)
	if !update {
		info, err := files.Lstat(baseline)
		if err != nil || info == nil || !info.Mode().IsRegular() || info.Size() == 0 {
			return fmt.Errorf("missing API baseline: %s", baseline)
		}
		if info.Size() > maximumAPIOutput {
			return fmt.Errorf("API baseline exceeds %d bytes: %s", maximumAPIOutput, baseline)
		}
	}
	temporaryDirectory := filepath.Dir(baseline)
	if err := ensureAPIBaselineDirectory(files, runner.Root, filepath.Dir(policy.Path), update); err != nil {
		return err
	}
	if !update {
		temporaryDirectory = ""
		if workspace, ok := runner.Executor.(taskWorkspace); ok {
			temporaryDirectory = workspace.TemporaryDirectory()
		}
	}
	temporary, err := files.CreateTemp(temporaryDirectory, ".golib-api-*")
	if err != nil {
		return fmt.Errorf("create API snapshot: %w", err)
	}
	temporaryPath := temporary.Name()
	defer func() { _ = files.Remove(temporaryPath) }()
	documentation, err := runner.generateAPISnapshot(ctx, directory, temporary, temporaryPath, module, policy)
	if err != nil {
		return err
	}
	info, err := files.Lstat(temporaryPath)
	if err != nil || info == nil || info.Size() == 0 {
		return fmt.Errorf("API snapshot for %s is empty", module.Directory)
	}
	if update {
		if err := files.Rename(temporaryPath, baseline); err != nil {
			return fmt.Errorf("publish API baseline for %s: %w", module.Directory, err)
		}
		_, _ = fmt.Fprintf(output, "[%s] API baseline updated\n", module.Directory)
		return nil
	}
	if policy.Mode == "go-doc" {
		readBaseline := runner.apiReadBaseline
		if readBaseline == nil {
			readBaseline = repositoryfile.Read
		}
		expected, readErr := readBaseline(runner.Root, policy.Path, maximumAPIOutput)
		if readErr != nil {
			return fmt.Errorf("read API baseline for %s: %w", module.Directory, readErr)
		}
		if !bytes.Equal(expected, documentation) {
			return fmt.Errorf("exported API documentation differs from %s", policy.Path)
		}
		_, _ = fmt.Fprintf(output, "[%s] API compatibility passed\n", module.Directory)
		return nil
	}
	tool := "golang.org/x/exp/cmd/apidiff@" + compatibilityToolVersion
	var report, diagnostics boundedBuffer
	err = runner.Executor.Run(ctx, Command{
		Name: "go", Dir: directory, Env: apiToolEnvironment(module),
		Args:   []string{"run", tool, "-m", "-incompatible", baseline, temporaryPath},
		Stdout: &report, Stderr: &diagnostics,
	})
	if report.overflow || diagnostics.overflow {
		return fmt.Errorf("API compatibility output exceeded %d bytes for %s", maximumAPIOutput, module.ModulePath)
	}
	if report.Len() > 0 {
		return fmt.Errorf("incompatible exported API changes in %s:\n%s", module.ModulePath, strings.TrimSpace(report.String()))
	}
	if err != nil {
		message := strings.TrimSpace(diagnostics.String())
		if message != "" {
			return fmt.Errorf("API compatibility tool failed for %s: %w: %s", module.ModulePath, err, message)
		}
		return fmt.Errorf("API compatibility tool failed for %s: %w", module.ModulePath, err)
	}
	_, _ = fmt.Fprintf(output, "[%s] API compatibility passed\n", module.Directory)
	return nil
}

func ensureAPIBaselineDirectory(files apiFileSystem, root, relative string, create bool) error {
	relative = filepath.Clean(relative)
	if filepath.IsAbs(relative) || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return fmt.Errorf("API baseline directory is not below repository root: %s", relative)
	}
	if relative == "." {
		return nil
	}
	current := root
	for component := range strings.SplitSeq(relative, string(filepath.Separator)) {
		current = filepath.Join(current, component)
		info, inspectErr := files.Lstat(current)
		if errors.Is(inspectErr, os.ErrNotExist) && create {
			if mkdirErr := files.MkdirAll(current, 0o700); mkdirErr != nil {
				return fmt.Errorf("create API baseline directory: %w", mkdirErr)
			}
			info, inspectErr = files.Lstat(current)
		}
		if inspectErr != nil || info == nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("API baseline directory is not a real directory: %s", current)
		}
	}
	return nil
}

func (runner Runner) apiBaseline(module inventory.Module) config.APIBaseline {
	for _, baseline := range runner.Policy.API.Baselines {
		if baseline.Module == module.Directory {
			return baseline
		}
	}
	return config.APIBaseline{
		Module: module.Directory,
		Mode:   "apidiff",
		Path:   filepath.Join(module.Directory, "api", "baseline.txt"),
	}
}

func (runner Runner) generateAPISnapshot(
	ctx context.Context,
	directory string,
	temporary namedWriteCloser,
	temporaryPath string,
	module inventory.Module,
	policy config.APIBaseline,
) ([]byte, error) {
	if policy.Mode == "apidiff" {
		if err := temporary.Close(); err != nil {
			return nil, fmt.Errorf("close API snapshot: %w", err)
		}
		tool := "golang.org/x/exp/cmd/apidiff@" + compatibilityToolVersion
		if err := runner.Executor.Run(ctx, Command{
			Name: "go", Dir: directory, Env: apiToolEnvironment(module),
			Args: []string{"run", tool, "-m", "-w", temporaryPath, module.ModulePath},
		}); err != nil {
			return nil, fmt.Errorf("generate API snapshot for %s: %w", module.Directory, err)
		}
		return nil, nil
	}

	var snapshot boundedBuffer
	for _, packagePattern := range policy.Packages {
		if err := runner.Executor.Run(ctx, Command{
			Name: "go", Dir: directory, Env: apiToolEnvironment(module),
			Args: []string{"doc", "-all", packagePattern}, Stdout: &snapshot,
		}); err != nil {
			_ = temporary.Close()
			return nil, fmt.Errorf("generate API documentation for %s: %w", packagePattern, err)
		}
		if snapshot.overflow {
			_ = temporary.Close()
			return nil, fmt.Errorf("API documentation exceeded %d bytes for %s", maximumAPIOutput, module.ModulePath)
		}
	}
	if snapshot.Len() == 0 {
		_ = temporary.Close()
		return nil, fmt.Errorf("API documentation for %s is empty", module.ModulePath)
	}
	documentation := append(bytes.TrimRight(snapshot.Bytes(), "\n"), '\n')
	written, err := temporary.Write(documentation)
	if err != nil {
		_ = temporary.Close()
		return nil, fmt.Errorf("write API documentation snapshot: %w", err)
	}
	if written != len(documentation) {
		_ = temporary.Close()
		return nil, fmt.Errorf("short write for API documentation snapshot: %d of %d bytes", written, len(documentation))
	}
	if err := temporary.Close(); err != nil {
		return nil, fmt.Errorf("close API snapshot: %w", err)
	}
	return documentation, nil
}

func apiToolEnvironment(module inventory.Module) map[string]string {
	return map[string]string{
		"GOTOOLCHAIN": "go" + module.GoVersion,
		"GOWORK":      "off",
	}
}
