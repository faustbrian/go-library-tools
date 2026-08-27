package gates

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestWorkflowsRunsPinnedActionlintWithBoundedDiagnostics(t *testing.T) {
	var command Command
	var output bytes.Buffer
	runner := Runner{Root: "/repo", Output: &output, Executor: executorFunction(func(_ context.Context, value Command) error {
		command = value
		return nil
	})}
	if err := runner.Workflows(context.Background()); err != nil {
		t.Fatalf("Workflows() error = %v", err)
	}
	if command.Name != "go" || command.Dir != "/repo" || command.Env["GOWORK"] != "off" {
		t.Fatalf("workflow command = %#v", command)
	}
	want := "run github.com/rhysd/actionlint/cmd/actionlint@v1.7.12 -no-color -oneline -shellcheck= -pyflakes="
	if strings.Join(command.Args, " ") != want || output.String() != "workflow contract passed\n" {
		t.Fatalf("workflow command/output = %q, %q", strings.Join(command.Args, " "), output.String())
	}
}

func TestWorkflowsReportsBoundedActionlintFailure(t *testing.T) {
	failure := errors.New("failed")
	runner := Runner{Root: "/repo", Executor: executorFunction(func(_ context.Context, command Command) error {
		_, _ = io.WriteString(command.Stderr, "invalid workflow")
		return failure
	})}
	err := runner.Workflows(context.Background())
	if err == nil || !errors.Is(err, failure) || !strings.Contains(err.Error(), "invalid workflow") {
		t.Fatalf("Workflows() error = %v", err)
	}
	runner.Executor = executorFunction(func(context.Context, Command) error { return failure })
	err = runner.Workflows(context.Background())
	if err == nil || !errors.Is(err, failure) {
		t.Fatalf("Workflows() plain tool error = %v", err)
	}
	if strings.Contains(err.Error(), "invalid workflow") {
		t.Fatalf("Workflows() plain tool diagnostics = %v", err)
	}

	runner.Executor = executorFunction(func(_ context.Context, command Command) error {
		_, _ = command.Stdout.Write(make([]byte, maximumWorkflowOutput+1))
		return nil
	})
	if err := runner.Workflows(context.Background()); err == nil || !strings.Contains(err.Error(), "exceeded") {
		t.Fatalf("Workflows() overflow error = %v", err)
	}
}

func TestBoundedWorkflowBufferEnforcesCumulativeLimit(t *testing.T) {
	var buffer boundedWorkflowBuffer
	first := make([]byte, maximumWorkflowOutput-1)
	if written, err := buffer.Write(first); err != nil || written != len(first) || buffer.overflow {
		t.Fatalf("first Write() = %d, %v, overflow %v", written, err, buffer.overflow)
	}
	if written, err := buffer.Write([]byte{'a'}); err != nil || written != 1 || buffer.overflow || buffer.data.Len() != maximumWorkflowOutput {
		t.Fatalf("boundary Write() = %d, %v, overflow %v, length %d", written, err, buffer.overflow, buffer.data.Len())
	}
	if written, err := buffer.Write([]byte("bc")); err != nil || written != 2 || !buffer.overflow || buffer.data.Len() != maximumWorkflowOutput {
		t.Fatalf("overflow Write() = %d, %v, overflow %v, length %d", written, err, buffer.overflow, buffer.data.Len())
	}
}
