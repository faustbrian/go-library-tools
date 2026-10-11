package cli

import (
	"bytes"
	"context"
	"errors"
	"io"
	"slices"
	"strings"
	"testing"

	"github.com/faustbrian/go-library-tools/v2/internal/gates"
)

func TestExecuteChangedSecretsRangeAndFailure(t *testing.T) {
	for _, fail := range []bool{false, true} {
		t.Run(map[bool]string{false: "success", true: "scanner failure"}[fail], func(t *testing.T) {
			root, workspace := internalFixture(t), t.TempDir()
			base, head := strings.Repeat("a", 40), strings.Repeat("b", 40)
			var scannedRange string
			executor := changedSecretsExecutor{directory: workspace, run: func(_ context.Context, command gates.Command) error {
				if command.Name == "git" {
					switch {
					case slices.Contains(command.Args, "rev-parse"):
						_, _ = io.WriteString(command.Stdout, head+"\n")
					case slices.Contains(command.Args, "for-each-ref"):
						_, _ = io.WriteString(command.Stdout, "refs/heads/main\n")
					case slices.Contains(command.Args, "--batch-all-objects"):
						_, _ = io.WriteString(command.Stdout, "1\n")
					case slices.Contains(command.Args, "diff"):
						_, _ = io.WriteString(command.Stdout, "README.md\x00")
					}
					return nil
				}
				if command.Name != "go" || !strings.Contains(strings.Join(command.Args, " "), "gitleaks/v8@") {
					return errors.New("unexpected runtime execution")
				}
				for _, argument := range command.Args {
					if value, ok := strings.CutPrefix(argument, "--log-opts="); ok {
						scannedRange = value
					}
				}
				if fail {
					_, _ = io.WriteString(command.Stderr, "private-marker")
					return errors.New("scanner failed")
				}
				return nil
			}}
			factory := func(string, io.Writer, io.Writer) (gates.Executor, func() error, error) {
				return executor, func() error { return nil }, nil
			}
			var stdout, stderr bytes.Buffer
			code := executeContext(t.Context(), []string{"secrets", "check", "--base-sha", base}, root, &stdout, &stderr, factory)
			if !strings.HasSuffix(scannedRange, base+".."+head) || (code != 0) != fail || strings.Contains(stdout.String()+stderr.String(), "private-marker") {
				t.Fatalf("changed secret command code=%d range=%q diagnostics=%q", code, scannedRange, stderr.String())
			}
		})
	}
}

type changedSecretsExecutor struct {
	directory string
	run       func(context.Context, gates.Command) error
}

func (executor changedSecretsExecutor) TemporaryDirectory() string { return executor.directory }
func (executor changedSecretsExecutor) Run(ctx context.Context, command gates.Command) error {
	return executor.run(ctx, command)
}

// Changed-local fixtures model the mandatory repository-owned source snapshot
// boundary even when the selected module disables broader security scanners.
func changedLocalCLIExecutor(t *testing.T, run func(context.Context, gates.Command) error) changedSecretsExecutor {
	t.Helper()
	return changedSecretsExecutor{directory: t.TempDir(), run: func(ctx context.Context, command gates.Command) error {
		if command.Name == "git" {
			switch {
			case slices.Contains(command.Args, "rev-parse"):
				_, err := io.WriteString(command.Stdout, strings.Repeat("b", 40)+"\n")
				return err
			case slices.Contains(command.Args, "for-each-ref"):
				_, err := io.WriteString(command.Stdout, "refs/heads/main\n")
				return err
			case slices.Contains(command.Args, "--batch-all-objects"):
				_, err := io.WriteString(command.Stdout, "1\n")
				return err
			}
		}
		return run(ctx, command)
	}}
}
