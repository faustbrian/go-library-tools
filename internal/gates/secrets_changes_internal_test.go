package gates

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/faustbrian/go-library-tools/v2/internal/inventory"
)

func changedSecretFixture(t *testing.T) (string, func(...string) string, func(string, string), func() string) {
	t.Helper()
	root := t.TempDir()
	git := func(args ...string) string {
		t.Helper()
		command := exec.CommandContext(t.Context(), "git", args...)
		command.Dir = root
		command.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_NOSYSTEM=1")
		output, err := command.Output()
		if err != nil {
			t.Fatal("owned Git fixture failed")
		}
		return strings.TrimSpace(string(output))
	}
	write := func(name, value string) {
		t.Helper()
		name = filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(name), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(name, []byte(value), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	commit := func() string {
		t.Helper()
		git("add", "--all") // The fixture is exclusively task-owned.
		git("-c", "user.name=fixture", "-c", "user.email=fixture@example.invalid", "-c", "commit.gpgsign=false", "-c", "core.hooksPath=/dev/null", "commit", "--allow-empty", "-qm", "fixture")
		return git("rev-parse", "HEAD")
	}
	git("init", "-q")
	return root, git, write, commit
}

func TestChangedSecretsSnapshotKeepsIntermediateHistoryAndOnlyChangedTree(t *testing.T) {
	for _, scenario := range []string{"changed tree", "empty net diff", "detached head"} {
		t.Run(scenario, func(t *testing.T) {
			root, git, write, commit := changedSecretFixture(t)
			write("old.txt", "baseline-marker\n")
			write("nested/file\n.md", "old contents\n")
			base := commit()
			write("temporary.txt", "intermediate-marker\n")
			commit()
			if err := os.Remove(filepath.Join(root, "temporary.txt")); err != nil {
				t.Fatal(err)
			}
			if scenario != "empty net diff" {
				write("nested/file\n.md", "changed-marker\n")
			}
			head := commit()
			if scenario == "detached head" {
				git("checkout", "--detach", "-q", head)
			}
			before := git("show-ref") + git("status", "--porcelain")
			workspace := t.TempDir()
			var modes []string
			runner := Runner{Root: root, Executor: workspaceExecutor{directory: workspace, emptySourceInventory: true, run: func(ctx context.Context, command Command) error {
				if command.Name == "git" {
					process := exec.CommandContext(ctx, command.Name, command.Args...)
					process.Dir, process.Stdout, process.Stderr = command.Dir, command.Stdout, command.Stderr
					process.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_NOSYSTEM=1")
					return process.Run()
				}
				if !strings.Contains(strings.Join(command.Args, " "), "gitleaks/v8@") {
					return errors.New("unexpected runtime command")
				}
				if slices.Contains(command.Args, "git") {
					modes = append(modes, "history")
					var options string
					for _, argument := range command.Args {
						if strings.HasPrefix(argument, "--log-opts=") {
							options = strings.TrimPrefix(argument, "--log-opts=")
						}
					}
					args := append([]string{"log", "-p", "-U0"}, strings.Fields(options)...)
					process := exec.CommandContext(ctx, "git", args...)
					process.Dir = command.Dir
					data, err := process.Output()
					if err != nil || !bytes.Contains(data, []byte("intermediate-marker")) || bytes.Contains(data, []byte("baseline-marker")) {
						t.Fatal("snapshot range lost new history or scanned old additions")
					}
				} else {
					modes = append(modes, "current")
					var files []string
					err := filepath.WalkDir(command.Dir, func(name string, entry os.DirEntry, err error) error {
						if err != nil {
							return err
						}
						if !entry.IsDir() {
							relative, err := filepath.Rel(command.Dir, name)
							if err != nil {
								return err
							}
							files = append(files, filepath.ToSlash(relative))
						}
						return nil
					})
					want := []string{"nested/file\n.md"}
					if scenario == "empty net diff" {
						want = nil
					}
					if err != nil || !slices.Equal(files, want) {
						t.Fatalf("current snapshot files=%q", files)
					}
				}
				return nil
			}}}
			if err := runner.SecretsChanges(t.Context(), base); err != nil {
				t.Fatal(err)
			}
			if !slices.Equal(modes, []string{"history", "current"}) || before != git("show-ref")+git("status", "--porcelain") {
				t.Fatal("incomplete scan or source mutation")
			}
			if files, err := os.ReadDir(workspace); err != nil || len(files) != 0 {
				t.Fatal("snapshot cleanup failed")
			}
		})
	}
}

func TestChangedSecretsRejectsUnavailableRangeBeforeSnapshot(t *testing.T) {
	root, git, write, commit := changedSecretFixture(t)
	write("file.txt", "base\n")
	base := commit()
	git("checkout", "--orphan", "unrelated", "-q")
	write("file.txt", "other\n")
	commit()
	for _, candidate := range []string{"invalid", strings.Repeat("f", 40), base} {
		workspace := t.TempDir()
		runner := Runner{Root: root, Executor: workspaceExecutor{directory: workspace, emptySourceInventory: true, run: func(ctx context.Context, command Command) error {
			if command.Name != "git" {
				t.Fatal("invalid range dispatched scanner")
			}
			process := exec.CommandContext(ctx, command.Name, command.Args...)
			process.Dir, process.Stdout, process.Stderr = command.Dir, command.Stdout, command.Stderr
			return process.Run()
		}}}
		if err := runner.SecretsChanges(t.Context(), candidate); err == nil {
			t.Fatal("invalid range accepted")
		}
		if files, err := os.ReadDir(workspace); err != nil || len(files) != 0 {
			t.Fatal("invalid range created snapshot")
		}
	}
}

func TestChangedSecretsPinnedScannerRetainsFindings(t *testing.T) {
	if os.Getenv("GOLIB_GITLEAKS_INTEGRATION") != "1" {
		t.Skip("opt-in pinned scanner integration")
	}
	root, git, write, commit := changedSecretFixture(t)
	credential := strings.Join([]string{"A1b2C3d4", "E5f6G7h8", "I9j0K1l2", "M3n4O5p6", "Q7r8S9t0", "U1v2W3x4", "Y5z6A7b8", "C9d0E1f2"}, "")
	write("old.txt", "api_key="+credential+"\n")
	write("README.md", "old documentation\n")
	base := commit()
	write("README.md", "new documentation\n")
	commit()
	var output bytes.Buffer
	executor, cleanup, err := NewProcessExecutor(root, io.Discard, io.Discard)
	if err != nil {
		t.Fatal("executor setup failed")
	}
	t.Cleanup(func() {
		if err := cleanup(); err != nil {
			t.Error("executor cleanup failed")
		}
	})
	runner := Runner{Root: root, Executor: executor, Output: &output}
	if err := runner.SecretsChanges(t.Context(), base); err != nil {
		t.Fatal("unchanged old credential was rescanned")
	}
	write("README.md", "api_key="+credential+"\n")
	if err := runner.SecretsChanges(t.Context(), base); err == nil || !strings.Contains(err.Error(), "secrets-current-tree") || !strings.Contains(err.Error(), "secret-findings") {
		t.Fatal("changed documentation contents escaped the current-file scan")
	}
	write("README.md", "new documentation\n")
	write("temporary.txt", "api_key="+credential+"\n")
	commit()
	if err := os.Remove(filepath.Join(root, "temporary.txt")); err != nil {
		t.Fatal(err)
	}
	commit()
	err = runner.SecretsChanges(t.Context(), base)
	if err == nil || !strings.Contains(err.Error(), "secret-findings") || strings.Contains(output.String()+err.Error(), credential) {
		t.Fatal("intermediate credential was missed or disclosed")
	}
	mergeBase := commit()
	mainBranch := git("branch", "--show-current")
	git("checkout", "-q", "-b", "side", mergeBase)
	commit()
	git("checkout", "-q", mainBranch)
	commit()
	git("-c", "user.name=fixture", "-c", "user.email=fixture@example.invalid", "-c", "commit.gpgsign=false", "-c", "core.hooksPath=/dev/null", "merge", "--no-commit", "--no-ff", "side")
	write("merge-only.txt", "api_key="+credential+"\n")
	commit()
	if err := os.Remove(filepath.Join(root, "merge-only.txt")); err != nil {
		t.Fatal(err)
	}
	commit()
	err = runner.SecretsChanges(t.Context(), mergeBase)
	if err == nil || !strings.Contains(err.Error(), "secret-findings") || strings.Contains(output.String()+err.Error(), credential) {
		t.Fatal("merge-only removed credential was missed or disclosed")
	}
}

func TestChangedSecretsDeadlineAndCancellationPrecedeSources(t *testing.T) {
	for _, cancelled := range []bool{false, true} {
		ctx, cancel := context.WithCancel(t.Context())
		if cancelled {
			cancel()
		}
		called := false
		runner := Runner{Executor: executorFunction(func(ctx context.Context, command Command) error {
			called = true
			if deadline, ok := ctx.Deadline(); !ok || time.Until(deadline) > changedSecretBudget {
				t.Fatal("range selection has no finite budget")
			}
			return errors.New("inert failure")
		})}
		err := runner.SecretsChanges(ctx, strings.Repeat("a", 40))
		cancel()
		if err == nil || called == cancelled || (cancelled && !errors.Is(err, context.Canceled)) {
			t.Fatal("range budget/cancellation admission failed")
		}
	}
}

func TestChangedSecretsLocalWithSecurityDisabledScansBeforeRuntime(t *testing.T) {
	for _, scenario := range []string{"changed local", "scanner failure", "legacy local"} {
		t.Run(scenario, func(t *testing.T) {
			root, _, write, commit := changedSecretFixture(t)
			write("go.mod", "module example\n\ngo 1.27.0\n")
			write("example.go", "package example\n")
			base := commit()
			write("example.go", "package example\n\nconst Value = 1\n")
			commit()
			var scans, tests, runtimeCommands int
			runner := Runner{Root: root, Catalog: inventory.Inventory{Modules: []inventory.Module{
				{Directory: ".", ModulePath: "example", Gates: map[string]bool{"tests": true}},
			}}, Executor: workspaceExecutor{directory: t.TempDir(), emptySourceInventory: true, run: func(ctx context.Context, command Command) error {
				if command.Name == "git" {
					process := exec.CommandContext(ctx, command.Name, command.Args...)
					process.Dir, process.Stdout, process.Stderr = command.Dir, command.Stdout, command.Stderr
					return process.Run()
				}
				if strings.Contains(strings.Join(command.Args, " "), "gitleaks/v8@") {
					scans++
					if runtimeCommands != 0 {
						t.Fatal("secret scan started after repository-controlled execution")
					}
					if scenario == "scanner failure" {
						return errors.New("scanner failed")
					}
					return nil
				}
				runtimeCommands++
				if command.Name == "go" && slices.Contains(command.Args, "test") {
					tests++
				}
				if scenario != "legacy local" && scans != 2 {
					return errors.New("runtime command preceded changed secret coverage")
				}
				return nil
			}}}
			var err error
			if scenario == "legacy local" {
				err = runner.Local(t.Context(), []string{"."})
			} else {
				err = runner.LocalChanges(t.Context(), []string{"."}, base)
			}
			if scenario == "scanner failure" {
				if err == nil || scans != 1 || runtimeCommands != 0 {
					t.Fatal("secret finding did not fail before runtime")
				}
			} else if err != nil || tests != 1 || (scenario == "legacy local" && scans != 0) || (scenario == "changed local" && scans != 2) {
				t.Fatalf("local coverage: scans=%d tests=%d err=%v", scans, tests, err)
			}
		})
	}
}

func TestChangedSecretsLocalRejectsNonAncestorWithSecurityDisabled(t *testing.T) {
	root, git, write, commit := changedSecretFixture(t)
	write("go.mod", "module example\n\ngo 1.27.0\n")
	write("example.go", "package example\n")
	base := commit()
	git("checkout", "--orphan", "unrelated", "-q")
	write("example.go", "package example\n\nconst Unrelated = true\n")
	commit()
	runner := Runner{Root: root, Catalog: inventory.Inventory{Modules: []inventory.Module{
		{Directory: ".", ModulePath: "example", Gates: map[string]bool{}},
	}}, Executor: workspaceExecutor{directory: t.TempDir(), emptySourceInventory: true, run: func(ctx context.Context, command Command) error {
		if command.Name != "git" {
			t.Fatal("nonancestor base reached scanner or runtime")
		}
		process := exec.CommandContext(ctx, command.Name, command.Args...)
		process.Dir, process.Stdout, process.Stderr = command.Dir, command.Stdout, command.Stderr
		return process.Run()
	}}}
	if err := runner.LocalChanges(t.Context(), []string{"."}, base); err == nil {
		t.Fatal("nonancestor base accepted with security disabled")
	}
}
