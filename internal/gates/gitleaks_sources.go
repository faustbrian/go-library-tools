package gates

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

const maximumGitleaksSnapshotBytes int64 = 4 << 30

type gitleaksSources struct {
	root       string
	history    string
	current    string
	ignoreRoot string
}

func (runner Runner) createGitleaksSources(ctx context.Context) (gitleaksSources, func() error, error) {
	workspace, ok := runner.Executor.(taskWorkspace)
	if !ok || !filepath.IsAbs(workspace.TemporaryDirectory()) {
		return gitleaksSources{}, nil, errors.New("gitleaks sources require an absolute task-owned temporary directory")
	}
	root, err := os.MkdirTemp(workspace.TemporaryDirectory(), "gitleaks-sources-")
	if err != nil {
		return gitleaksSources{}, nil, fmt.Errorf("create gitleaks source workspace: %w", err)
	}
	cleanup := func() error {
		remove := runner.gitleaksSourceCleanup
		if remove == nil {
			remove = os.RemoveAll
		}
		return remove(root)
	}
	sources := gitleaksSources{
		root: root, history: filepath.Join(root, "history"),
		current: filepath.Join(root, "current"), ignoreRoot: filepath.Join(root, "ignore"),
	}
	historyBundle := filepath.Join(root, "history.bundle")
	if err := os.Mkdir(sources.ignoreRoot, 0o700); err != nil {
		return gitleaksSources{}, nil, errors.Join(fmt.Errorf("create gitleaks ignore root: %w", err), cleanup())
	}
	commands := []Command{
		{Name: "git", Args: []string{"-C", runner.Root, "bundle", "create", historyBundle, "--all"}, Dir: workspace.TemporaryDirectory()},
		{Name: "git", Args: []string{"init", "--quiet", "--", sources.history}, Dir: workspace.TemporaryDirectory()},
		{Name: "git", Args: []string{
			"-C", sources.history, "fetch", "--quiet", "--force", "--no-recurse-submodules",
			historyBundle, "+refs/*:refs/golib-source/*",
		}, Dir: workspace.TemporaryDirectory()},
	}
	for _, command := range commands {
		if err := runner.runGitleaksSourceCommand(ctx, command); err != nil {
			return gitleaksSources{}, nil, errors.Join(fmt.Errorf("create gitleaks history snapshot: %w", err), cleanup())
		}
	}
	if err := copyGitleaksCurrentTree(runner.Root, sources.current); err != nil {
		return gitleaksSources{}, nil, errors.Join(err, cleanup())
	}
	return sources, cleanup, nil
}

func (runner Runner) runGitleaksSourceCommand(ctx context.Context, command Command) error {
	stdout := &boundedProcessOutput{limit: maximumSecurityProcessOutput}
	stderr := &boundedProcessOutput{limit: maximumSecurityProcessOutput}
	command.Stdout = stdout
	command.Stderr = stderr
	err := runner.Executor.Run(ctx, command)
	if stdout.didOverflow() || stderr.didOverflow() {
		err = errors.Join(err, fmt.Errorf("gitleaks history snapshot output exceeded %d bytes", maximumSecurityProcessOutput))
	}
	return err
}

func copyGitleaksCurrentTree(source, destination string) error {
	if err := os.Mkdir(destination, 0o700); err != nil {
		return fmt.Errorf("create gitleaks current-tree snapshot: %w", err)
	}
	root, err := os.OpenRoot(source)
	if err != nil {
		return fmt.Errorf("open gitleaks current-tree source: %w", err)
	}
	defer root.Close()
	targetRoot, err := os.OpenRoot(destination)
	if err != nil {
		return fmt.Errorf("open gitleaks current-tree snapshot: %w", err)
	}
	defer targetRoot.Close()
	files := 0
	var copied int64
	err = filepath.WalkDir(source, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, relErr := filepath.Rel(source, path)
		if relErr != nil {
			return relErr
		}
		if relative == "." {
			return nil
		}
		if relative == ".git" {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if relative == ".gitleaksignore" && !entry.IsDir() {
			return nil
		}
		if entry.IsDir() {
			return targetRoot.Mkdir(relative, 0o700)
		}
		files++
		if files > maximumSecuritySourceFiles {
			return fmt.Errorf("gitleaks current-tree snapshot exceeds %d files", maximumSecuritySourceFiles)
		}
		if entry.Type()&os.ModeSymlink != 0 {
			// filepath.WalkDir and Gitleaks both treat symlinks as non-regular
			// entries. Omitting them preserves the scanner's no-follow behavior.
			return nil
		}
		if !entry.Type().IsRegular() {
			return fmt.Errorf("gitleaks current-tree source contains unsupported file type: %s", filepath.ToSlash(relative))
		}
		remaining := maximumGitleaksSnapshotBytes - copied
		if remaining < 0 {
			return fmt.Errorf("gitleaks current-tree snapshot exceeds %d bytes", maximumGitleaksSnapshotBytes)
		}
		input, openErr := root.Open(relative)
		if openErr != nil {
			return openErr
		}
		output, createErr := targetRoot.OpenFile(relative, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
		if createErr != nil {
			_ = input.Close()
			return createErr
		}
		written, copyErr := io.Copy(output, io.LimitReader(input, remaining+1))
		closeErr := errors.Join(input.Close(), output.Close())
		if copyErr != nil || closeErr != nil {
			return errors.Join(copyErr, closeErr)
		}
		if written > remaining {
			return fmt.Errorf("gitleaks current-tree snapshot exceeds %d bytes", maximumGitleaksSnapshotBytes)
		}
		copied += written
		return nil
	})
	if err != nil {
		return fmt.Errorf("create gitleaks current-tree snapshot: %w", err)
	}
	return nil
}
