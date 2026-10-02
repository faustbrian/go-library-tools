package gates

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

// Exercise the production quota with tiny actual workflows, not a fabricated
// charged count. Refusal must precede any external validator effects.
func TestWorkflowDescriptorFileCountBeforeExecutor(t *testing.T) {
	for _, count := range []int{512, 513} {
		t.Run(strconv.Itoa(count), func(t *testing.T) {
			root := t.TempDir()
			directory := filepath.Join(root, ".github", "workflows")
			if err := os.MkdirAll(directory, 0o700); err != nil {
				t.Fatal(err)
			}
			const content = "jobs: {}\n"
			for index := range count {
				path := filepath.Join(directory, fmt.Sprintf("workflow-%03d.yml", index))
				if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			var output bytes.Buffer
			calls := 0
			runner := Runner{Root: root, Output: &output, Executor: executorFunction(func(context.Context, Command) error {
				calls++
				return nil
			})}
			err := runner.Workflows(t.Context())
			if count == 513 {
				if err == nil || err.Error() != "workflow security policy: workflow security policy: workflow file limit exceeded" || calls != 0 || output.Len() != 0 {
					t.Fatalf("workflow file quota refusal: error=%v calls=%d output=%q", err, calls, output.String())
				}
			} else if err != nil || calls != 1 || output.String() != "workflow contract passed\n" {
				t.Fatalf("inclusive workflow file quota: error=%v calls=%d output=%q", err, calls, output.String())
			}
			for index := range count {
				path := filepath.Join(directory, fmt.Sprintf("workflow-%03d.yml", index))
				data, err := os.ReadFile(path)
				if err != nil || string(data) != content {
					t.Fatal("workflow file quota inspection changed source bytes")
				}
			}
		})
	}
}

func workflowDescriptorRealFiles() workflowDescriptorFiles {
	return workflowDescriptorFiles{openRoot: os.OpenRoot, info: os.DirEntry.Info, open: (*os.Root).Open, close: (*os.File).Close}
}

// Real file operations produce every fault, without permission assumptions or
// asynchronous tree changes. The private read limit keeps growth fixtures tiny.
func TestWorkflowDescriptorReadLifecycle(t *testing.T) {
	for _, stage := range []string{"metadata", "metadata-size", "symbolic", "open", "read", "close", "read-close", "growth", "growth-close", "exact", "below"} {
		t.Run(stage, func(t *testing.T) {
			rootPath := t.TempDir()
			const source = "jobs: {}\n"
			name := "source.yml"
			path := filepath.Join(rootPath, name)
			unrelated := filepath.Join(rootPath, "unrelated.txt")
			for path, content := range map[string]string{path: source, unrelated: "keep unrelated bytes"} {
				if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			if stage == "symbolic" {
				name = "symbolic.yml"
				if err := os.Symlink("source.yml", filepath.Join(rootPath, name)); err != nil {
					t.Fatal(err)
				}
			}
			root, err := os.OpenRoot(rootPath)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = root.Close() })
			entries, err := os.ReadDir(rootPath)
			if err != nil {
				t.Fatal(err)
			}
			var entry os.DirEntry
			for _, candidate := range entries {
				if candidate.Name() == name {
					entry = candidate
				}
			}
			if entry == nil {
				t.Fatal("missing owned descriptor entry")
			}
			var input *os.File
			t.Cleanup(func() {
				if input != nil {
					_ = input.Close()
				}
			})
			var operationErr, closeErr error
			opens, closes := 0, 0
			var offset int64
			limit := int64(len(source))
			switch stage {
			case "metadata-size":
				limit--
			case "below":
				limit++
			}
			files := workflowDescriptorRealFiles()
			files.info = func(entry os.DirEntry) (os.FileInfo, error) {
				if stage == "metadata" {
					info, err := root.Stat("missing.yml")
					operationErr = err
					return info, err
				}
				return entry.Info()
			}
			files.open = func(root *os.Root, relative string) (*os.File, error) {
				opens++
				if stage == "open" {
					file, err := root.Open("missing.yml")
					operationErr = err
					return file, err
				}
				if strings.HasPrefix(stage, "growth") {
					if err := os.WriteFile(path, []byte(source+"# added\n"), 0o600); err != nil {
						t.Fatal(err)
					}
				}
				var err error
				if stage == "read" {
					input, err = root.OpenFile(relative, os.O_WRONLY, 0)
				} else {
					input, err = root.Open(relative)
				}
				if err == nil && stage == "read-close" {
					if err := input.Close(); err != nil {
						t.Fatal(err)
					}
				}
				return input, err
			}
			files.close = func(file *os.File) error {
				closes++
				if stage != "read-close" {
					var err error
					offset, err = file.Seek(0, io.SeekCurrent)
					if err != nil {
						t.Fatal(err)
					}
				}
				closeErr = file.Close()
				if stage == "close" || stage == "growth-close" {
					closeErr = errors.Join(closeErr, file.Close())
				}
				return closeErr
			}
			content, err := readWorkflowDescriptor(root, name, entry, limit, files)
			if stage == "exact" || stage == "below" {
				if err != nil || string(content) != source {
					t.Fatalf("inclusive descriptor limit changed: content=%q error=%v", content, err)
				}
			} else if err == nil || content != nil {
				t.Fatalf("descriptor failure returned usable content: %q, %v", content, err)
			}
			wantOpens, wantCloses := 1, 1
			switch stage {
			case "metadata", "open":
				if reflect.TypeOf(operationErr) != reflect.TypeFor[*os.PathError]() ||
					reflect.TypeOf(err) != reflect.TypeOf(operationErr) ||
					reflect.ValueOf(err).Pointer() != reflect.ValueOf(operationErr).Pointer() {
					t.Fatal("descriptor operation error identity changed")
				}
				wantCloses = 0
				if stage == "metadata" {
					wantOpens = 0
				}
			case "symbolic":
				if err.Error() != "workflow security policy: unsupported descriptor file type" {
					t.Fatalf("descriptor type refusal changed: %v", err)
				}
				wantOpens, wantCloses = 0, 0
			case "metadata-size", "growth":
				if err.Error() != "workflow security policy: source.yml exceeds size limit" {
					t.Fatalf("descriptor size refusal changed: %v", err)
				}
				if stage == "metadata-size" {
					wantOpens, wantCloses = 0, 0
				}
			case "read", "read-close":
				joined, ok := err.(interface{ Unwrap() []error })
				if !ok || len(joined.Unwrap()) == 0 {
					t.Fatal("descriptor read failure was not joined")
				}
				var readErr *os.PathError
				if !errors.As(joined.Unwrap()[0], &readErr) || readErr.Op != "read" {
					t.Fatal("descriptor read failure lost its original cause")
				}
				if stage == "read-close" {
					if reflect.TypeOf(closeErr) != reflect.TypeFor[*os.PathError]() || len(joined.Unwrap()) != 2 ||
						reflect.TypeOf(joined.Unwrap()[1]) != reflect.TypeOf(closeErr) ||
						reflect.ValueOf(joined.Unwrap()[1]).Pointer() != reflect.ValueOf(closeErr).Pointer() {
						t.Fatal("descriptor read/close causes were not retained in order")
					}
				}
			}
			if stage == "close" || stage == "read-close" || stage == "growth-close" {
				if closeErr == nil || !errors.Is(err, closeErr) || !errors.Is(err, os.ErrClosed) {
					t.Fatal("descriptor close error identity or I/O precedence changed")
				}
			}
			if opens != wantOpens || closes != wantCloses {
				t.Fatalf("descriptor admission/closure changed: opens=%d closes=%d", opens, closes)
			}
			if input != nil {
				if _, err := input.Stat(); !errors.Is(err, os.ErrClosed) {
					t.Fatal("descriptor read retained an open file handle")
				}
			}
			if strings.HasPrefix(stage, "growth") && offset != limit+1 {
				t.Fatalf("descriptor growth read exceeded its bounded prefix: %d", offset)
			}
			wantSource := source
			if strings.HasPrefix(stage, "growth") {
				wantSource += "# added\n"
			}
			for path, want := range map[string]string{path: wantSource, unrelated: "keep unrelated bytes"} {
				data, err := os.ReadFile(path)
				if err != nil || string(data) != want {
					t.Fatal("descriptor lifecycle changed source or unrelated bytes")
				}
			}
			if stage == "symbolic" {
				if target, err := os.Readlink(filepath.Join(rootPath, name)); err != nil || target != "source.yml" {
					t.Fatal("descriptor refusal changed symbolic entry identity")
				}
			}
		})
	}
}

func TestWorkflowDescriptorRootOpeningPreservesAbsenceAndErrors(t *testing.T) {
	for _, missing := range []bool{false, true} {
		t.Run(strconv.FormatBool(missing), func(t *testing.T) {
			root := t.TempDir()
			workflowRefusalWrite(t, root, ".github/workflows/ci.yml", "jobs: {}\n")
			invalidRoot := filepath.Join(root, ".github", "workflows", "ci.yml")
			if missing {
				invalidRoot = filepath.Join(root, "missing")
			}
			var openingErr error
			files := workflowDescriptorRealFiles()
			files.openRoot = func(string) (*os.Root, error) {
				opened, err := os.OpenRoot(invalidRoot)
				openingErr = err
				return opened, err
			}
			err := checkWorkflowSecurityWithFiles(t.Context(), root, workflowDescriptorLimits{descriptorBytes: 64, totalBytes: 64}, files)
			if openingErr == nil {
				t.Fatal("fixture did not produce a real directory opening error")
			}
			if missing {
				if !errors.Is(openingErr, os.ErrNotExist) || err != nil {
					t.Fatalf("workflow directory absence admission changed: %v", err)
				}
			} else if !errors.Is(err, openingErr) || err.Error() != "workflow security policy: "+openingErr.Error() {
				t.Fatalf("workflow root opening cause or diagnostic changed: %v", err)
			}
			data, err := os.ReadFile(filepath.Join(root, ".github", "workflows", "ci.yml"))
			if err != nil || string(data) != "jobs: {}\n" {
				t.Fatal("root admission changed workflow bytes")
			}
		})
	}
}

func TestWorkflowDescriptorReadCancellationBeforeLocalTraversal(t *testing.T) {
	root := t.TempDir()
	workflowRefusalWrite(t, root, ".github/workflows/ci.yml", "jobs: {}\n")
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	var ownedRoot *os.Root
	var input *os.File
	files := workflowDescriptorRealFiles()
	files.openRoot = func(path string) (*os.Root, error) {
		root, err := os.OpenRoot(path)
		ownedRoot = root
		return root, err
	}
	files.open = func(root *os.Root, name string) (*os.File, error) {
		file, err := root.Open(name)
		input = file
		return file, err
	}
	files.close = func(file *os.File) error {
		err := file.Close()
		cancel()
		return err
	}
	err := checkWorkflowSecurityWithFiles(ctx, root, workflowDescriptorLimits{descriptorBytes: 64, totalBytes: 64}, files)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("post-read cancellation reached local traversal: %v", err)
	}
	if input == nil || ownedRoot == nil {
		t.Fatal("cancellation did not occur at the real descriptor read boundary")
	}
	if _, err := input.Stat(); !errors.Is(err, os.ErrClosed) {
		t.Fatal("cancelled inspection retained its descriptor handle")
	}
	if _, err := ownedRoot.Stat("ci.yml"); !errors.Is(err, os.ErrClosed) {
		t.Fatal("cancelled inspection retained its rooted directory handle")
	}
	data, err := os.ReadFile(filepath.Join(root, ".github", "workflows", "ci.yml"))
	if err != nil || string(data) != "jobs: {}\n" {
		t.Fatal("cancellation changed workflow bytes")
	}
}

func TestWorkflowDescriptorAggregateBytesIncludeSharedLocalActions(t *testing.T) {
	for _, local := range []bool{false, true} {
		for _, adjustment := range []int{-1, 0, 1} {
			t.Run(fmt.Sprintf("local-%v/adjustment-%d", local, adjustment), func(t *testing.T) {
				root := t.TempDir()
				workflow := "jobs: {}\n"
				contents := map[string]string{"unrelated.txt": "keep unrelated bytes"}
				if local {
					workflow = "jobs: {test: {steps: [{uses: ./ordinary}]}}\n"
					contents["ordinary/action.yml"] = "runs: {using: composite, steps: []}\n"
				}
				contents[".github/workflows/one.yml"] = workflow
				contents[".github/workflows/two.yml"] = workflow
				for path, content := range contents {
					workflowRefusalWrite(t, root, path, content)
				}
				// Both workflows reference the same local descriptor, charged
				// once by its existing owner regardless of enumeration order.
				total := 2*len(workflow) + len(contents["ordinary/action.yml"])
				err := checkWorkflowSecurityWithFiles(t.Context(), root, workflowDescriptorLimits{descriptorBytes: 64, totalBytes: total + adjustment}, workflowDescriptorRealFiles())
				if adjustment < 0 {
					if err == nil || err.Error() != "workflow security policy: workflow security policy: total descriptor byte limit exceeded" {
						t.Fatalf("aggregate descriptor bytes were not refused: %v", err)
					}
				} else if err != nil {
					t.Fatalf("inclusive aggregate descriptor bytes were refused: %v", err)
				}
				for path, want := range contents {
					data, err := os.ReadFile(filepath.Join(root, path))
					if err != nil || string(data) != want {
						t.Fatal("aggregate admission changed descriptor or unrelated bytes")
					}
				}
			})
		}
	}
}
