package gates

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestSBOMRejectsInvalidOutput(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   string
	}{
		{"empty", "", "empty"},
		{"malformed", "{", "valid JSON"},
		{"wrong format", `{"bomFormat":"SPDX","specVersion":"1.6"}`, "bomFormat"},
		{"missing version", `{"bomFormat":"CycloneDX"}`, "specVersion"},
		{"unsupported version", `{"bomFormat":"CycloneDX","specVersion":"1.5"}`, "unsupported specVersion"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			runner := Runner{Executor: executorFunction(func(_ context.Context, command Command) error {
				_, _ = io.WriteString(command.Stdout, test.output)
				return nil
			})}
			err := runner.runSBOM(context.Background(), "/repo")
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("runSBOM() error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestSBOMDiagnosticsDoNotDiscloseToolControlledValues(t *testing.T) {
	secret := "credentials-AKIA1234567890-secret"
	tests := []string{
		`{"bomFormat":"` + secret + `","specVersion":"1.6"}`,
		`{"bomFormat":"CycloneDX","specVersion":"` + secret + `"}`,
		`{"bomFormat":"CycloneDX","specVersion":"1.6","payload":"` + secret + `"`,
	}
	for _, output := range tests {
		runner := Runner{Executor: executorFunction(func(_ context.Context, command Command) error {
			_, _ = io.WriteString(command.Stdout, output)
			_, _ = io.WriteString(command.Stderr, secret)
			return nil
		})}
		err := runner.runSBOM(context.Background(), "/repo")
		if err == nil || strings.Contains(err.Error(), secret) {
			t.Fatalf("runSBOM() error disclosed tool output: %v", err)
		}
	}
}

func TestSBOMOverflowPreservesBothOverflowAndToolFailure(t *testing.T) {
	failure := errors.New("tool failed")
	runner := Runner{Executor: executorFunction(func(_ context.Context, command Command) error {
		_, _ = command.Stderr.Write(make([]byte, maximumSBOMOutput+1))
		return failure
	})}
	err := runner.runSBOM(context.Background(), "/repo")
	if err == nil || !strings.Contains(err.Error(), "output exceeded") || !errors.Is(err, failure) {
		t.Fatalf("runSBOM() error = %v", err)
	}
}

func TestBoundedSBOMBufferAcceptsItsExactLimit(t *testing.T) {
	var buffer boundedSBOMBuffer
	value := make([]byte, maximumSBOMOutput)
	written, err := buffer.Write(value)
	if err != nil || written != len(value) || buffer.overflow || buffer.data.Len() != maximumSBOMOutput {
		t.Fatalf("Write() = %d, %v, overflow %v, length %d", written, err, buffer.overflow, buffer.data.Len())
	}
}

func TestBoundedSBOMBufferLimitsCumulativeWrites(t *testing.T) {
	var buffer boundedSBOMBuffer
	first := make([]byte, maximumSBOMOutput-1)
	if written, err := buffer.Write(first); err != nil || written != len(first) || buffer.overflow {
		t.Fatalf("first Write() = %d, %v, overflow %v", written, err, buffer.overflow)
	}
	if written, err := buffer.Write([]byte("ab")); err != nil || written != 2 || !buffer.overflow || buffer.data.Len() != maximumSBOMOutput {
		t.Fatalf("second Write() = %d, %v, overflow %v, length %d", written, err, buffer.overflow, buffer.data.Len())
	}
}

func TestSBOMRejectsExcessiveOutputAndPreservesToolFailure(t *testing.T) {
	runner := Runner{Executor: executorFunction(func(_ context.Context, command Command) error {
		_, _ = command.Stdout.Write(make([]byte, maximumSBOMOutput+1))
		return nil
	})}
	err := runner.runSBOM(context.Background(), "/repo")
	if err == nil || !strings.Contains(err.Error(), "exceeded") {
		t.Fatalf("runSBOM() overflow error = %v", err)
	}

	failure := errors.New("failed")
	runner.Executor = executorFunction(func(context.Context, Command) error { return failure })
	if err := runner.runSBOM(context.Background(), "/repo"); !errors.Is(err, failure) {
		t.Fatalf("runSBOM() tool error = %v", err)
	}
}
