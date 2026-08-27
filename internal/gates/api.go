package gates

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/faustbrian/go-library-tools/internal/inventory"
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
	baseline := filepath.Join(directory, "api", "baseline.txt")
	if !update {
		info, err := files.Lstat(baseline)
		if err != nil || !info.Mode().IsRegular() || info.Size() == 0 {
			return fmt.Errorf("missing API baseline: %s", baseline)
		}
	}
	temporaryDirectory := filepath.Dir(baseline)
	if !update {
		temporaryDirectory = ""
		if workspace, ok := runner.Executor.(taskWorkspace); ok {
			temporaryDirectory = workspace.TemporaryDirectory()
		}
	} else if err := files.MkdirAll(temporaryDirectory, 0o700); err != nil {
		return fmt.Errorf("create API baseline directory: %w", err)
	} else if info, err := files.Lstat(temporaryDirectory); err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("API baseline directory is not a real directory: %s", temporaryDirectory)
	}
	temporary, err := files.CreateTemp(temporaryDirectory, ".golib-api-*")
	if err != nil {
		return fmt.Errorf("create API snapshot: %w", err)
	}
	temporaryPath := temporary.Name()
	if err := temporary.Close(); err != nil {
		_ = files.Remove(temporaryPath)
		return fmt.Errorf("close API snapshot: %w", err)
	}
	defer func() { _ = files.Remove(temporaryPath) }()
	tool := "golang.org/x/exp/cmd/apidiff@" + compatibilityToolVersion
	if err := runner.Executor.Run(ctx, Command{
		Name: "go", Dir: directory, Env: map[string]string{"GOWORK": "off"},
		Args: []string{"run", tool, "-m", "-w", temporaryPath, module.ModulePath},
	}); err != nil {
		return fmt.Errorf("generate API snapshot for %s: %w", module.Directory, err)
	}
	info, err := files.Lstat(temporaryPath)
	if err != nil || info.Size() == 0 {
		return fmt.Errorf("API snapshot for %s is empty", module.Directory)
	}
	if update {
		if err := files.Rename(temporaryPath, baseline); err != nil {
			return fmt.Errorf("publish API baseline for %s: %w", module.Directory, err)
		}
		_, _ = fmt.Fprintf(output, "[%s] API baseline updated\n", module.Directory)
		return nil
	}
	var report, diagnostics boundedBuffer
	err = runner.Executor.Run(ctx, Command{
		Name: "go", Dir: directory, Env: map[string]string{"GOWORK": "off"},
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
