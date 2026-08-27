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

	"github.com/faustbrian/go-library-tools/internal/inventory"
	"github.com/faustbrian/go-library-tools/internal/repositoryfile"
	"golang.org/x/mod/module"
	modzip "golang.org/x/mod/zip"
)

type releaseFileSystem interface {
	MkdirTemp(string, string) (string, error)
	MkdirAll(string, os.FileMode) error
	WriteFile(string, []byte, os.FileMode) error
	Create(string) (io.WriteCloser, error)
}

type operatingReleaseFiles struct{}

func (operatingReleaseFiles) MkdirTemp(directory, pattern string) (string, error) {
	return os.MkdirTemp(directory, pattern)
}

func (operatingReleaseFiles) MkdirAll(path string, mode os.FileMode) error {
	return os.MkdirAll(path, mode)
}

func (operatingReleaseFiles) WriteFile(path string, data []byte, mode os.FileMode) error {
	return os.WriteFile(path, data, mode)
}

func (operatingReleaseFiles) Create(path string) (io.WriteCloser, error) {
	return os.Create(path)
}

// ReleaseDryRun rejects tag collisions, proves clean module consumption from a
// task-owned proxy, and then executes every release gate.
func (runner Runner) ReleaseDryRun(ctx context.Context, selection []string) error {
	modules, err := runner.selectModules(selection)
	if err != nil {
		return err
	}
	if err := runner.releaseRehearsal(ctx, modules); err != nil {
		return err
	}
	return runner.Check(ctx, selection)
}

func (runner Runner) releaseRehearsal(ctx context.Context, modules []inventory.Module) error {
	workspace, ok := runner.Executor.(taskWorkspace)
	if !ok || !filepath.IsAbs(workspace.TemporaryDirectory()) {
		return errors.New("release rehearsal requires a task-owned workspace")
	}
	for _, item := range modules {
		tag := item.TagPrefix + item.Version
		var output bytes.Buffer
		if err := runner.Executor.Run(ctx, Command{
			Name: "git", Args: []string{"tag", "--list", tag}, Dir: runner.Root, Stdout: &output,
		}); err != nil {
			return fmt.Errorf("inspect release tag %s: %w", tag, err)
		}
		if strings.TrimSpace(output.String()) != "" {
			return fmt.Errorf("release tag already exists: %s", tag)
		}
	}
	proxy, err := runner.buildReleaseProxy(workspace.TemporaryDirectory(), modules)
	if err != nil {
		return err
	}
	for _, item := range modules {
		version := "v" + item.Version
		if err := runner.Executor.Run(ctx, Command{
			Name: "go", Args: []string{"list", "-m", item.ModulePath + "@" + version},
			Dir:    workspace.TemporaryDirectory(),
			Env:    map[string]string{"GOPROXY": "file://" + proxy, "GOSUMDB": "off", "GOWORK": "off"},
			Stdout: io.Discard, Stderr: runner.Output,
		}); err != nil {
			return fmt.Errorf("consume release module %s@%s: %w", item.ModulePath, version, err)
		}
	}
	return nil
}

func (runner Runner) buildReleaseProxy(workspace string, modules []inventory.Module) (string, error) {
	files := runner.releaseFiles
	if files == nil {
		files = operatingReleaseFiles{}
	}
	proxy, err := files.MkdirTemp(workspace, "release-proxy-")
	if err != nil {
		return "", fmt.Errorf("create release proxy: %w", err)
	}
	for _, item := range modules {
		version := "v" + item.Version
		escapedPath, err := module.EscapePath(item.ModulePath)
		if err != nil {
			return "", fmt.Errorf("escape release module path %s: %w", item.ModulePath, err)
		}
		escapedVersion, err := module.EscapeVersion(version)
		if err != nil {
			return "", fmt.Errorf("escape release module version %s: %w", version, err)
		}
		directory := filepath.Join(proxy, filepath.FromSlash(escapedPath), "@v")
		if err := files.MkdirAll(directory, 0o700); err != nil {
			return "", fmt.Errorf("create release module proxy: %w", err)
		}
		moduleFile := filepath.Join(item.Directory, "go.mod")
		data, err := repositoryfile.Read(runner.Root, moduleFile, maximumModuleFileSize)
		if err != nil {
			return "", fmt.Errorf("read release %s: %w", moduleFile, err)
		}
		prefix := filepath.Join(directory, escapedVersion)
		if err := files.WriteFile(prefix+".mod", data, 0o600); err != nil {
			return "", fmt.Errorf("write release module metadata: %w", err)
		}
		information := fmt.Appendf(nil, "{\"Version\":%q,\"Time\":\"2000-01-01T00:00:00Z\"}\n", version)
		if err := files.WriteFile(prefix+".info", information, 0o600); err != nil {
			return "", fmt.Errorf("write release module information: %w", err)
		}
		if err := files.WriteFile(filepath.Join(directory, "list"), []byte(version+"\n"), 0o600); err != nil {
			return "", fmt.Errorf("write release module version list: %w", err)
		}
		archive, err := files.Create(prefix + ".zip")
		if err != nil {
			return "", fmt.Errorf("create release module archive: %w", err)
		}
		source := filepath.Join(runner.Root, filepath.FromSlash(item.Directory))
		create := runner.releaseArchive
		if create == nil {
			create = modzip.CreateFromDir
		}
		createErr := create(archive, module.Version{Path: item.ModulePath, Version: version}, source)
		closeErr := archive.Close()
		if createErr != nil {
			return "", fmt.Errorf("archive release module %s: %w", item.ModulePath, createErr)
		}
		if closeErr != nil {
			return "", fmt.Errorf("close release module archive: %w", closeErr)
		}
	}
	return proxy, nil
}
