package gates

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// QS1: literal inventories protect inclusive cumulative budgets and fail-closed
// parsing without manufacturing a large repository.
func TestGitleaksHistoryInventoryBoundaries(t *testing.T) {
	for _, test := range []struct {
		name, refs, objects string
		valid               bool
	}{
		{"below", "refs/heads/main\n", "1\n", true},
		{"exact", "refs/heads/main\nrefs/tags/release\n", "2\n3\n", true},
		{"extra-ref", "refs/heads/main\nrefs/tags/a\nrefs/tags/b\n", "1\n", false},
		{"empty-ref", "", "1\n", false},
		{"interior-empty-ref", "refs/heads/main\n\nrefs/tags/a\n", "1\n", false},
		{"malformed-ref", "HEAD\n", "1\n", false},
		{"extra-object", "refs/heads/main\n", "1\n1\n1\n", false},
		{"cumulative-bytes", "refs/heads/main\n", "3\n3\n", false},
		{"empty-object", "refs/heads/main\n", "", false},
		{"negative-object", "refs/heads/main\n", "-1\n", false},
		{"malformed-object", "refs/heads/main\n", "size\n", false},
		{"overflow-object", "refs/heads/main\n", "9223372036854775808\n", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			workspace := t.TempDir()
			runner := Runner{Root: t.TempDir(), Executor: workspaceExecutor{
				directory: workspace, emptySourceInventory: true,
				run: func(_ context.Context, command Command) error {
					value := test.refs
					if command.Args[2] == "cat-file" {
						value = test.objects
					}
					_, err := io.WriteString(command.Stdout, value)
					return err
				},
			}}
			err := runner.preflightGitleaksHistory(t.Context(), securitySourceLimits{refs: 2, objects: 2, bytes: 5})
			if (err == nil) != test.valid {
				t.Fatalf("inventory acceptance = %v, want %v", err == nil, test.valid)
			}
			entries, err := os.ReadDir(workspace)
			if err != nil || len(entries) != 0 {
				t.Fatalf("inventory produced artifacts: %v", err)
			}
		})
	}
}

// QS2/QS6: a command's hostile text stays out of operator errors, while its
// original cause and cancellation remain available to callers.
func TestGitleaksSourceCommandPreservesSafeCauses(t *testing.T) {
	const marker = "inert-private-output-marker"
	cause := errors.New(marker)
	for _, stage := range []string{"refs", "objects", "snapshot", "cancelled"} {
		t.Run(stage, func(t *testing.T) {
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			runner := Runner{Root: t.TempDir(), Executor: workspaceExecutor{
				directory: t.TempDir(), run: func(_ context.Context, command Command) error {
					if stage == "objects" && command.Args[2] == "for-each-ref" {
						return nil
					}
					_, _ = io.WriteString(command.Stderr, marker)
					if stage == "cancelled" {
						cancel()
					}
					return cause
				},
			}}
			var err error
			if stage == "refs" || stage == "objects" {
				err = runner.preflightGitleaksHistory(ctx, runner.sourceLimits())
			} else {
				err = runner.runGitleaksSourceCommand(ctx, Command{Name: "git"})
			}
			if !errors.Is(err, cause) || strings.Contains(err.Error(), marker) {
				t.Fatal("command cause was lost or disclosed")
			}
			if stage == "cancelled" && !errors.Is(err, context.Canceled) {
				t.Fatal("command cancellation was lost")
			}
		})
	}
}

// QS2: independent streams must each enforce their inclusive output limit.
func TestGitleaksSourceCommandOutputLimits(t *testing.T) {
	for _, stream := range []string{"stdout", "stderr"} {
		for _, extra := range []int{0, 1} {
			t.Run(stream+strings.Repeat("-over", extra), func(t *testing.T) {
				runner := Runner{Executor: workspaceExecutor{run: func(_ context.Context, command Command) error {
					writer := command.Stdout
					if stream == "stderr" {
						writer = command.Stderr
					}
					_, err := io.Copy(writer, strings.NewReader(strings.Repeat("x", maximumSecurityProcessOutput)))
					if err == nil && extra != 0 {
						_, err = io.WriteString(writer, "x")
					}
					return err
				}}}
				err := runner.runGitleaksSourceCommand(t.Context(), Command{Name: "git"})
				if (err == nil) != (extra == 0) {
					t.Fatalf("output boundary accepted = %v", err == nil)
				}
			})
		}
	}
}

// QS3: readback, not a Write return count, proves cumulative persisted bounds.
func TestBoundedSourceFileCumulativeWrites(t *testing.T) {
	for _, closed := range []bool{false, true} {
		t.Run(map[bool]string{false: "bounded", true: "closed"}[closed], func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "snapshot")
			file, err := os.Create(path)
			if err != nil {
				t.Fatal(err)
			}
			defer file.Close()
			writer := &boundedSourceFile{file: file, limit: 3}
			if _, err := io.WriteString(writer, "a"); err != nil {
				t.Fatal(err)
			}
			if closed {
				if err := file.Close(); err != nil {
					t.Fatal(err)
				}
				for _, value := range []string{"b", "bcd"} {
					if _, err := io.WriteString(writer, value); !errors.Is(err, os.ErrClosed) {
						t.Fatal("natural write failure was not retained")
					}
				}
			} else {
				if _, err := io.WriteString(writer, "bc"); err != nil {
					t.Fatal(err)
				}
				if _, err := io.WriteString(writer, "d"); err == nil || !writer.didOverflow() {
					t.Fatal("exhausted cumulative budget accepted another write")
				}
			}
			want := "abc"
			if closed {
				want = "a"
			}
			data, err := os.ReadFile(path)
			if err != nil || string(data) != want {
				t.Fatalf("persisted bytes differ from bounded prefix: %v", err)
			}
		})
	}
}

// QS5/QS6: failures at existing command boundaries remove partial sources and
// never return a usable source or cleanup handle.
func TestGitleaksSourceConstructionFailureCleanup(t *testing.T) {
	for _, stage := range []string{"bundle", "init", "fetch", "current", "cancel-init", "cancel-fetch"} {
		t.Run(stage, func(t *testing.T) {
			root, workspace := t.TempDir(), t.TempDir()
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			cause := errors.New("inert-source-failure")
			runner := Runner{Root: root, Executor: workspaceExecutor{directory: workspace, run: func(_ context.Context, command Command) error {
				name := command.Args[0]
				if name == "-C" {
					name = command.Args[2]
				}
				if stage == name {
					return cause
				}
				if stage == "cancel-"+name {
					cancel()
				}
				if stage == "current" && name == "fetch" {
					return os.RemoveAll(root)
				}
				return nil
			}}}
			sources, cleanup, err := runner.createGitleaksSources(ctx)
			if err == nil || cleanup != nil || sources != (gitleaksSources{}) {
				t.Fatal("construction failure returned usable sources")
			}
			if stage == "bundle" || stage == "init" || stage == "fetch" {
				if !errors.Is(err, cause) {
					t.Fatal("construction lost command cause")
				}
			}
			if strings.HasPrefix(stage, "cancel-") && !errors.Is(err, context.Canceled) {
				t.Fatal("construction lost cancellation")
			}
			entries, err := os.ReadDir(workspace)
			if err != nil || len(entries) != 0 {
				t.Fatalf("partial sources remain: %v", err)
			}
		})
	}
}

// QS5: invalid workspace paths cannot produce source artifacts.
func TestGitleaksSourcesRequireUsableAbsoluteWorkspace(t *testing.T) {
	owned := t.TempDir()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	relative, err := filepath.Rel(cwd, owned)
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{relative, filepath.Join(t.TempDir(), "absent")} {
		runner := Runner{Root: t.TempDir(), Executor: workspaceExecutor{directory: path, run: func(_ context.Context, _ Command) error { return nil }}}
		sources, cleanup, err := runner.createGitleaksSources(t.Context())
		if err == nil || cleanup != nil || sources != (gitleaksSources{}) {
			if cleanup != nil {
				_ = cleanup()
			}
			t.Fatal("invalid task workspace produced sources")
		}
		if path == relative {
			entries, err := os.ReadDir(owned)
			if err != nil || len(entries) != 0 {
				t.Fatal("relative workspace acquired artifacts")
			}
		} else if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
			t.Fatal("absent workspace was created")
		}
	}
}

// QS4: metadata exclusions do not hide a legitimate suppression-named
// directory, and both aggregate entry and byte limits are inclusive.
func TestGitleaksCurrentTreeInclusiveBudgets(t *testing.T) {
	source := t.TempDir()
	if err := os.Mkdir(filepath.Join(source, ".gitleaksignore"), 0o700); err != nil {
		t.Fatal(err)
	}
	for path, content := range map[string]string{".git": "ignored metadata", ".gitleaksignore/visible": "a", "payload": "bcd"} {
		if err := os.WriteFile(filepath.Join(source, path), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	for _, test := range []struct {
		name    string
		entries int
		bytes   int64
		valid   bool
	}{{"below", 4, 5, true}, {"exact", 3, 4, true}, {"entry-over", 2, 4, false}, {"byte-over", 3, 3, false}} {
		t.Run(test.name, func(t *testing.T) {
			destination := filepath.Join(t.TempDir(), "snapshot")
			err := copyGitleaksCurrentTreeBounded(t.Context(), source, destination, securitySourceLimits{entries: test.entries, bytes: test.bytes})
			if (err == nil) != test.valid {
				t.Fatalf("current-tree acceptance = %v, want %v", err == nil, test.valid)
			}
			if !test.valid {
				if _, err := os.Stat(destination); !errors.Is(err, os.ErrNotExist) {
					t.Fatal("rejected tree produced a snapshot")
				}
				return
			}
			for path, want := range map[string]string{".gitleaksignore/visible": "a", "payload": "bcd"} {
				data, err := os.ReadFile(filepath.Join(destination, path))
				if err != nil || string(data) != want {
					t.Fatalf("eligible tree contents not preserved: %v", err)
				}
			}
			if _, err := os.Lstat(filepath.Join(destination, ".git")); !errors.Is(err, os.ErrNotExist) {
				t.Fatal("Git worktree metadata was copied")
			}
		})
	}
}
