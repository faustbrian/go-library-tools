package cli

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/faustbrian/go-library-tools/v2/internal/gates"
)

func TestExecuteRepositorySecretsRejectsMalformedArguments(t *testing.T) {
	for _, test := range []struct {
		name string
		args []string
	}{
		{name: "missing subcommand", args: []string{"secrets"}},
		{name: "wrong subcommand", args: []string{"secrets", "invalid"}},
		{name: "empty subcommand", args: []string{"secrets", ""}},
		{name: "extra argument", args: []string{"secrets", "check", "extra"}},
		{name: "wrong subcommand and extra argument", args: []string{"secrets", "invalid", "extra"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := internalFixture(t)
			factoryCalled := false
			factory := func(string, io.Writer, io.Writer) (gates.Executor, func() error, error) {
				factoryCalled = true
				return nil, nil, errors.New("unexpected executor creation")
			}
			var stdout, stderr bytes.Buffer
			code := executeContext(t.Context(), test.args, root, &stdout, &stderr, factory)
			if code != 2 || stderr.String() != "usage: golib secrets check\n" || stdout.Len() != 0 {
				t.Fatalf("malformed command result = %d, stdout=%q, stderr=%q", code, stdout.String(), stderr.String())
			}
			if factoryCalled {
				t.Fatal("malformed command created an execution capability")
			}
		})
	}
}

func TestExecuteRepositorySecretsOnly(t *testing.T) {
	for _, stage := range []string{"success", "history", "current", "ignore", "cancelled"} {
		t.Run(stage, func(t *testing.T) {
			root := internalFixture(t)
			if stage == "ignore" {
				if err := os.WriteFile(filepath.Join(root, ".gitleaksignore"), []byte("inert\n"), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			if stage == "cancelled" {
				cancel()
			}
			executor := &secretsOnlyExecutor{directory: t.TempDir(), fail: stage}
			factory := func(string, io.Writer, io.Writer) (gates.Executor, func() error, error) {
				return executor, func() error { return nil }, nil
			}
			var stdout, stderr bytes.Buffer
			code := executeContext(ctx, []string{"secrets", "check"}, root, &stdout, &stderr, factory)
			wantCode := 1
			if stage == "success" {
				wantCode = 0
			}
			if code != wantCode || (stderr.Len() == 0) != (stage == "success") {
				t.Fatalf("secret-only result = %d, error present=%t", code, stderr.Len() != 0)
			}
			wantStages := []string{"git", "dir"}
			switch stage {
			case "history":
				wantStages = []string{"git"}
			case "ignore", "cancelled":
				wantStages = nil
			}
			if !slices.Equal(executor.stages, wantStages) {
				t.Fatalf("scan stages = %v, want %v", executor.stages, wantStages)
			}
			if strings.Contains(stdout.String()+stderr.String(), "inert scanner detail") {
				t.Fatal("scanner detail disclosed")
			}
		})
	}
}

type secretsOnlyExecutor struct {
	directory string
	fail      string
	stages    []string
}

func (executor *secretsOnlyExecutor) TemporaryDirectory() string { return executor.directory }

func (executor *secretsOnlyExecutor) Run(ctx context.Context, command gates.Command) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if command.Name == "git" {
		if slices.Contains(command.Args, "for-each-ref") {
			_, _ = io.WriteString(command.Stdout, "refs/heads/main\n")
		}
		if slices.Contains(command.Args, "cat-file") {
			_, _ = io.WriteString(command.Stdout, "1\n")
		}
		return nil
	}
	if command.Name != "go" || !strings.Contains(strings.Join(command.Args, " "), "github.com/zricethezav/gitleaks/v8@") {
		return errors.New("unexpected runtime command")
	}
	stage := "dir"
	if slices.Contains(command.Args, "git") {
		stage = "git"
	}
	executor.stages = append(executor.stages, stage)
	if executor.fail == "history" && stage == "git" || executor.fail == "current" && stage == "dir" {
		_, _ = io.WriteString(command.Stderr, "inert scanner detail")
		return errors.New("scanner failed")
	}
	return nil
}
