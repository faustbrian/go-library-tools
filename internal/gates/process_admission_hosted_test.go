//go:build darwin || linux

package gates

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"testing"
	"time"
)

type admissionProcessOutput struct {
	bytes.Buffer
	registrations int
}

func (output *admissionProcessOutput) setOverflowCallback(func()) {
	output.registrations++
}

func TestHostedBoundedProcessPreCanceledAdmission(t *testing.T) {
	if os.Getenv("CI") != "true" {
		t.Skip("bounded process controls run only in hosted CI")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	prepares := 0
	output := &admissionProcessOutput{}
	executor := &processExecutor{prepareBoundedProcess: func(*exec.Cmd) error {
		prepares++
		return nil
	}}
	err := executor.Run(ctx, Command{boundedScanner: true, Name: "", Stdout: output, Stderr: output})
	if !errors.Is(err, context.Canceled) || errors.Unwrap(err) != nil || prepares != 0 || output.Len() != 0 || output.registrations != 0 {
		t.Fatalf("pre-canceled admission = %v, prepares=%d output=%d registrations=%d", err, prepares, output.Len(), output.registrations)
	}
}

func TestHostedBoundedProcessEmptyCommandStartFailure(t *testing.T) {
	if os.Getenv("CI") != "true" {
		t.Skip("bounded process controls run only in hosted CI")
	}
	var prepared *exec.Cmd
	output := &admissionProcessOutput{}
	executor := &processExecutor{prepareBoundedProcess: func(command *exec.Cmd) error {
		prepared = command
		return nil
	}}
	err := executor.Run(t.Context(), Command{boundedScanner: true, Name: "", Stdout: output})
	if err == nil || err.Error() != "run : exec: no command" || errors.Unwrap(err) == nil || prepared == nil || prepared.Process != nil || output.Len() != 0 || output.registrations != 1 {
		t.Fatalf("empty command = %v, prepared=%v output=%d registrations=%d", err, prepared != nil, output.Len(), output.registrations)
	}
}

func TestHostedBoundedProcessShortSuccess(t *testing.T) {
	if os.Getenv("CI") != "true" {
		t.Skip("bounded process controls run only in hosted CI")
	}
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	var stdout, stderr bytes.Buffer
	executor := &processExecutor{}
	err := executor.Run(ctx, Command{
		boundedScanner: true,
		Name:           os.Args[0], Args: []string{"-test.run=^TestProcessHelper$", "--"},
		Env: map[string]string{
			"GO_WANT_HELPER": "1", "HELPER_OUTPUT": "short-owned-output", "HELPER_ERROR": "",
			"HELPER_FAIL": "", "HELPER_STREAM": "", "HELPER_WAIT": "", "HELPER_SPAWN_DESCENDANT": "", "HELPER_READ_STDIN": "",
		},
		Stdout: &stdout, Stderr: &stderr,
	})
	if err != nil || stdout.String() != "short-owned-output" || stderr.Len() != 0 {
		t.Fatalf("short command = %v, stdout=%q stderr=%q", err, stdout.String(), stderr.String())
	}
}
