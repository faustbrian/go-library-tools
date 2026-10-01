package gates

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// Cancellation is a pre-effect refusal, not an executor error after dispatch.
func TestGitleaksRefusalBeforeCommandDispatch(t *testing.T) {
	for _, boundary := range []string{"history", "snapshot"} {
		t.Run(boundary, func(t *testing.T) {
			ctx, cancel := context.WithCancel(t.Context())
			cancel()
			calls := 0
			runner := Runner{Root: t.TempDir(), Executor: executorFunction(func(context.Context, Command) error {
				calls++
				return nil
			})}
			var err error
			if boundary == "history" {
				err = runner.preflightGitleaksHistory(ctx, securitySourceLimits{refs: 1, objects: 1, bytes: 1})
			} else {
				err = runner.runGitleaksSourceCommand(ctx, Command{Name: "git"})
			}
			if !errors.Is(err, context.Canceled) || calls != 0 {
				t.Fatalf("pre-canceled %s: cancellation=%v executor calls=%d", boundary, errors.Is(err, context.Canceled), calls)
			}
		})
	}
}

// Valid command output must not cause another inventory step after cancellation.
func TestGitleaksRefusalAfterInventoryCancellation(t *testing.T) {
	for _, cancelAfter := range []string{"for-each-ref", "cat-file"} {
		t.Run(cancelAfter, func(t *testing.T) {
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			var calls []string
			runner := Runner{Root: t.TempDir(), Executor: executorFunction(func(_ context.Context, command Command) error {
				step := command.Args[2]
				calls = append(calls, step)
				output := "refs/heads/main\n"
				if step == "cat-file" {
					output = "1\n"
				}
				if _, err := io.WriteString(command.Stdout, output); err != nil {
					return err
				}
				if step == cancelAfter {
					cancel()
				}
				return nil
			})}
			err := runner.preflightGitleaksHistory(ctx, securitySourceLimits{refs: 1, objects: 1, bytes: 1})
			want := []string{"for-each-ref"}
			if cancelAfter == "cat-file" {
				want = append(want, "cat-file")
			}
			if !errors.Is(err, context.Canceled) || !reflect.DeepEqual(calls, want) {
				t.Fatalf("inventory cancellation=%v calls=%v want=%v", errors.Is(err, context.Canceled), calls, want)
			}
		})
	}
}

// A valid history inventory does not authorize effects for an oversized tree.
func TestGitleaksRefusalCurrentTreeBeforeWorkspaceEffects(t *testing.T) {
	root, workspace := t.TempDir(), t.TempDir()
	path := filepath.Join(root, "source.txt")
	if err := os.WriteFile(path, []byte("ab"), 0o600); err != nil {
		t.Fatal(err)
	}
	limits := securitySourceLimits{refs: 1, objects: 1, entries: 1, bytes: 1}
	var calls []string
	artifactCommand := false
	runner := Runner{Root: root, securitySourceLimits: &limits, Executor: workspaceExecutor{
		directory: workspace, emptySourceInventory: true,
		run: func(_ context.Context, command Command) error {
			step := command.Args[0]
			if step == "-C" {
				step = command.Args[2]
			}
			calls = append(calls, step)
			switch step {
			case "for-each-ref":
				_, err := io.WriteString(command.Stdout, "refs/heads/main\n")
				return err
			case "cat-file":
				_, err := io.WriteString(command.Stdout, "1\n")
				return err
			default:
				artifactCommand = true
				return nil
			}
		},
	}}
	sources, cleanup, err := runner.createGitleaksSources(t.Context())
	if cleanup != nil {
		defer func() { _ = cleanup() }()
	}
	if err == nil || sources != (gitleaksSources{}) || cleanup != nil {
		t.Fatal("over-budget current tree returned usable sources or cleanup")
	}
	if artifactCommand || !reflect.DeepEqual(calls, []string{"for-each-ref", "cat-file"}) {
		t.Fatalf("preflight refusal allowed artifact command: calls=%v", calls)
	}
	entries, err := os.ReadDir(workspace)
	if err != nil || len(entries) != 0 {
		t.Fatalf("workspace changed on refusal: entries=%d error=%v", len(entries), err)
	}
	data, err := os.ReadFile(path)
	if err != nil || string(data) != "ab" {
		t.Fatal("source bytes changed on refusal")
	}
}

// Existing destinations are never overwritten or partly populated.
func TestGitleaksRefusalPreservesExistingDestination(t *testing.T) {
	source, destination := t.TempDir(), t.TempDir()
	input, sentinel := filepath.Join(source, "source.txt"), filepath.Join(destination, "sentinel.txt")
	if err := os.WriteFile(input, []byte("ab"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(sentinel, []byte("keep"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := copyGitleaksCurrentTreeBounded(t.Context(), source, destination, securitySourceLimits{entries: 1, bytes: 2}); err == nil {
		t.Fatal("existing destination was accepted")
	}
	for path, want := range map[string]string{input: "ab", sentinel: "keep"} {
		data, err := os.ReadFile(path)
		if err != nil || string(data) != want {
			t.Fatal("existing destination refusal changed owned bytes")
		}
	}
	entries, err := os.ReadDir(destination)
	if err != nil || len(entries) != 1 || entries[0].Name() != "sentinel.txt" {
		t.Fatal("existing destination was partly populated")
	}
}
