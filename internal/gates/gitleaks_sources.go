package gates

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const maximumGitleaksSnapshotBytes int64 = 4 << 30

type securitySourceLimits struct {
	refs, objects, entries int
	bytes                  int64
}

type sourceCommandError struct {
	class string
	cause error
}

func (err sourceCommandError) Error() string { return err.class }
func (err sourceCommandError) Unwrap() error { return err.cause }

func (runner Runner) sourceLimits() securitySourceLimits {
	if runner.securitySourceLimits != nil {
		return *runner.securitySourceLimits
	}
	return securitySourceLimits{refs: 10_000, objects: 100_000, entries: maximumSecuritySourceFiles, bytes: maximumGitleaksSnapshotBytes}
}

type sourceInventoryOutput struct {
	data bytes.Buffer
	boundedProcessOutput
}

func (output *sourceInventoryOutput) String() string { return output.data.String() }
func (output *sourceInventoryOutput) Len() int       { return output.data.Len() }

func (output *sourceInventoryOutput) Write(value []byte) (int, error) {
	n, err := output.boundedProcessOutput.Write(value)
	if !output.didOverflow() {
		_, _ = output.data.Write(value)
	}
	return n, err
}

// Validate metadata without creating a bundle or copying a source entry.
func (runner Runner) preflightGitleaksHistory(ctx context.Context, limits securitySourceLimits) error {
	for _, objects := range []bool{false, true} {
		if err := ctx.Err(); err != nil {
			return err
		}
		args := []string{"-C", runner.Root, "for-each-ref", "--format=%(refname)"}
		if objects {
			args = []string{"-C", runner.Root, "cat-file", "--batch-all-objects", "--batch-check=%(objectsize)"}
		}
		stdout := &sourceInventoryOutput{limit: maximumSecurityProcessOutput}
		stderr := &boundedProcessOutput{limit: maximumSecurityProcessOutput}
		if err := runner.Executor.Run(ctx, Command{Name: "git", Args: args, Dir: runner.Root, Stdout: stdout, Stderr: stderr, boundedScanner: true}); err != nil || stdout.didOverflow() || stderr.didOverflow() {
			return errors.Join(sourceCommandError{class: "gitleaks history inventory failed or exceeded output limit", cause: err}, ctx.Err())
		}
		count := 0
		var total int64
		for line := range strings.SplitSeq(strings.TrimSuffix(stdout.String(), "\n"), "\n") {
			if err := ctx.Err(); err != nil {
				return err
			}
			if line == "" {
				return errors.New("gitleaks history inventory is empty or malformed")
			}
			count++
			if !objects {
				if !strings.HasPrefix(line, "refs/") || count > limits.refs {
					return errors.New("gitleaks history reference limit exceeded or malformed")
				}
			} else {
				size, err := strconv.ParseInt(line, 10, 64)
				if err != nil || size < 0 || count > limits.objects || size > limits.bytes-total {
					return errors.New("gitleaks history object or byte limit exceeded or malformed")
				}
				total += size
			}
		}
	}
	return nil
}

type gitleaksSources struct {
	root       string
	history    string
	current    string
	ignoreRoot string
}

func (runner Runner) createGitleaksSources(ctx context.Context) (gitleaksSources, func() error, error) {
	if err := ctx.Err(); err != nil {
		return gitleaksSources{}, nil, err
	}
	workspace, ok := runner.Executor.(taskWorkspace)
	if !ok || !filepath.IsAbs(workspace.TemporaryDirectory()) {
		return gitleaksSources{}, nil, errors.New("gitleaks sources require an absolute task-owned temporary directory")
	}
	limits := runner.sourceLimits()
	if err := runner.preflightGitleaksHistory(ctx, limits); err != nil {
		return gitleaksSources{}, nil, err
	}
	if err := inspectGitleaksCurrentTree(ctx, runner.Root, limits, nil); err != nil {
		return gitleaksSources{}, nil, err
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
	// #nosec G304 -- this exact bundle path is beneath the freshly created task-owned source root
	bundle, err := os.OpenFile(historyBundle, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return gitleaksSources{}, nil, errors.Join(errors.New("create bounded history bundle"), cleanup())
	}
	boundedBundle := &boundedSourceFile{file: bundle, limit: limits.bytes}
	stderr := &boundedProcessOutput{limit: maximumSecurityProcessOutput}
	err = runner.Executor.Run(ctx, Command{Name: "git", Args: []string{"-C", runner.Root, "bundle", "create", "-", "--all"}, Dir: workspace.TemporaryDirectory(), Stdout: boundedBundle, Stderr: stderr, boundedScanner: true})
	closeErr := bundle.Close()
	if err != nil || closeErr != nil || ctx.Err() != nil || boundedBundle.didOverflow() || stderr.didOverflow() {
		return gitleaksSources{}, nil, errors.Join(sourceCommandError{class: "bounded history bundle creation failed", cause: err}, ctx.Err(), cleanup())
	}
	commands := []Command{
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
	if err := copyGitleaksCurrentTreeBounded(ctx, runner.Root, sources.current, limits); err != nil {
		return gitleaksSources{}, nil, errors.Join(err, cleanup())
	}
	return sources, cleanup, nil
}

func (runner Runner) runGitleaksSourceCommand(ctx context.Context, command Command) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stdout := &boundedProcessOutput{limit: maximumSecurityProcessOutput}
	stderr := &boundedProcessOutput{limit: maximumSecurityProcessOutput}
	command.Stdout = stdout
	command.Stderr = stderr
	command.boundedScanner = true
	err := runner.Executor.Run(ctx, command)
	if stdout.didOverflow() || stderr.didOverflow() {
		err = errors.Join(err, fmt.Errorf("gitleaks history snapshot output exceeded %d bytes", maximumSecurityProcessOutput))
	}
	if err != nil {
		return errors.Join(sourceCommandError{class: "gitleaks history snapshot command failed", cause: err}, ctx.Err())
	}
	return ctx.Err()
}

func copyGitleaksCurrentTree(source, destination string) error {
	return copyGitleaksCurrentTreeBounded(context.Background(), source, destination, Runner{}.sourceLimits())
}

type boundedSourceFile struct {
	file           *os.File
	limit, written int64
	boundedProcessOutput
}

func (file *boundedSourceFile) Write(value []byte) (int, error) {
	remaining := file.limit - file.written
	if int64(len(value)) > remaining {
		if remaining > 0 {
			n, err := file.file.Write(value[:int(remaining)])
			file.written += int64(n)
			if err != nil {
				return n, err
			}
		}
		// This embedded writer only records overflow; the stable bundle error
		// below is authoritative rather than its deliberately ignored result.
		_, _ = file.boundedProcessOutput.Write(make([]byte, 1))
		return len(value), errors.New("history bundle byte limit exceeded")
	}
	n, err := file.file.Write(value)
	file.written += int64(n)
	return n, err
}

func inspectGitleaksCurrentTree(ctx context.Context, source string, limits securitySourceLimits, visit func(string, fs.DirEntry) error) error {
	return inspectGitleaksCurrentTreeWithMetadata(ctx, source, limits, visit, func(_ string, entry fs.DirEntry) (fs.FileInfo, error) {
		return entry.Info()
	})
}

// Keep metadata failure injection local to this operation: some filesystems
// cache DirEntry.Info during ReadDir, even if the source subsequently disappears.
func inspectGitleaksCurrentTreeWithMetadata(ctx context.Context, source string, limits securitySourceLimits, visit func(string, fs.DirEntry) error, metadata func(string, fs.DirEntry) (fs.FileInfo, error)) error {
	entries := 0
	var total int64
	return walkSecuritySource(ctx, source, limits.entries+2, func(relative string, entry fs.DirEntry) error {
		if err := ctx.Err(); err != nil {
			return err
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
		entries++
		if entries > limits.entries {
			return errors.New("gitleaks current-tree entry limit exceeded")
		}
		if entry.Type().IsRegular() {
			info, err := metadata(relative, entry)
			if err != nil {
				return errors.New("gitleaks current-tree metadata failed")
			}
			if info.Size() < 0 || info.Size() > limits.bytes-total {
				return errors.New("gitleaks current-tree byte limit exceeded")
			}
			total += info.Size()
		} else if !entry.IsDir() && entry.Type()&os.ModeSymlink == 0 {
			return errors.New("gitleaks current-tree unsupported entry type")
		}
		if visit != nil {
			return visit(relative, entry)
		}
		return nil
	})
}

func copyGitleaksCurrentTreeBounded(ctx context.Context, source, destination string, limits securitySourceLimits) error {
	return copyGitleaksCurrentTreeWithFiles(ctx, source, destination, limits, gitleaksCopyFiles{
		open: (*os.Root).Open,
		create: func(root *os.Root, relative string) (*os.File, error) {
			return root.OpenFile(relative, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
		},
		close: (*os.File).Close,
	})
}

// These operations stay local to a single copy, retaining real confined roots
// and file handles while allowing deterministic lifecycle failure checks.
type gitleaksCopyFiles struct {
	openRoot     func(string) (*os.Root, error)
	open, create func(*os.Root, string) (*os.File, error)
	close        func(*os.File) error
}

func copyGitleaksCurrentTreeWithFiles(ctx context.Context, source, destination string, limits securitySourceLimits, files gitleaksCopyFiles) error {
	openRoot := files.openRoot
	if openRoot == nil {
		openRoot = os.OpenRoot
	}
	if err := inspectGitleaksCurrentTree(ctx, source, limits, nil); err != nil {
		return err
	}
	if err := os.Mkdir(destination, 0o700); err != nil {
		return errors.New("create gitleaks current-tree snapshot failed")
	}
	root, err := openRoot(source)
	if err != nil {
		return errors.New("open gitleaks current-tree source failed")
	}
	defer root.Close()
	targetRoot, err := openRoot(destination)
	if err != nil {
		return errors.New("open gitleaks current-tree snapshot failed")
	}
	defer targetRoot.Close()
	var copied int64
	err = inspectGitleaksCurrentTree(ctx, source, limits, func(relative string, entry fs.DirEntry) error {
		if entry.IsDir() {
			return targetRoot.Mkdir(relative, 0o700)
		}
		if entry.Type()&os.ModeSymlink != 0 {
			// filepath.WalkDir and Gitleaks both treat symlinks as non-regular
			// entries. Omitting them preserves the scanner's no-follow behavior.
			return nil
		}
		remaining := limits.bytes - copied
		input, openErr := files.open(root, relative)
		if openErr != nil {
			return errors.New("gitleaks current-tree source open failed")
		}
		output, createErr := files.create(targetRoot, relative)
		if createErr != nil {
			_ = files.close(input)
			return errors.New("gitleaks current-tree destination open failed")
		}
		bounded := &boundedSourceFile{file: output, limit: remaining}
		written, copyErr := io.Copy(bounded, io.LimitReader(contextSourceReader{ctx: ctx, reader: input}, remaining+1))
		closeErr := errors.Join(files.close(input), files.close(output))
		if copyErr != nil || closeErr != nil {
			return errors.Join(errors.New("gitleaks current-tree copy failed"), ctx.Err())
		}
		copied += written
		return nil
	})
	if err != nil {
		return fmt.Errorf("create gitleaks current-tree snapshot: %w", err)
	}
	return nil
}

type contextSourceReader struct {
	ctx    context.Context
	reader io.Reader
}

func (reader contextSourceReader) Read(value []byte) (int, error) {
	if err := reader.ctx.Err(); err != nil {
		return 0, err
	}
	return reader.reader.Read(value)
}
