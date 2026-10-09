package services

import (
	"context"
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestGenericStartupRetainsBoundedDiagnostics(t *testing.T) {
	for _, service := range []string{"nats", "postgresql"} {
		for _, text := range []string{"", " \t\r\n", "docker: registry pull rejected"} {
			t.Run(service+"/"+text, func(t *testing.T) {
				cause := &diagnosticFailure{}
				manager := Manager{Token: fixedToken, Process: func(_ context.Context, _ string, args []string, _ map[string]string, stdout, stderr io.Writer) error {
					if args[0] != "run" {
						t.Fatal("failed startup continued to another operation")
					}
					_, _ = io.WriteString(stdout, "discarded stdout")
					_, _ = io.WriteString(stderr, text)
					return cause
				}}
				lease, err := manager.Start(context.Background(), []string{service})
				var typed *diagnosticFailure
				if lease != nil || !errors.Is(err, cause) || !errors.As(err, &typed) || typed != cause {
					t.Fatal("startup changed its failure identity or returned a lease")
				}
				want := "start " + service + ": docker exited 125"
				if strings.TrimSpace(text) != "" {
					want += ": " + text
				}
				if err.Error() != want {
					t.Fatalf("diagnostic = %q, want %q", err, want)
				}
			})
		}
	}
}

func TestGenericStartupSanitizesHostileDiagnosticTail(t *testing.T) {
	for _, chunkSize := range []int{3, 6000} {
		cause := errors.New("docker exited 125")
		text := strings.Repeat("x", 5000) + "fixture golib guest\x1b\x00\rdecisive failure"
		manager := Manager{Token: fixedToken, Process: func(_ context.Context, _ string, _ []string, _ map[string]string, _ io.Writer, stderr io.Writer) error {
			for remaining := text; remaining != ""; {
				size := min(chunkSize, len(remaining))
				written, err := io.WriteString(stderr, remaining[:size])
				if written != size || err != nil {
					t.Fatal("diagnostic capture rejected process output")
				}
				remaining = remaining[size:]
			}
			return cause
		}}
		_, err := manager.Start(context.Background(), []string{"postgresql"})
		_, diagnostic, found := strings.Cut(err.Error(), "docker exited 125: ")
		if !found || len(diagnostic) > 4096 || !utf8.ValidString(diagnostic) ||
			!strings.HasPrefix(diagnostic, "[diagnostic truncated]") ||
			!strings.HasSuffix(diagnostic, "fixture ***** *****??\ndecisive failure") ||
			strings.ContainsAny(diagnostic, "\x1b\x00\r") || strings.Contains(diagnostic, "golib") || strings.Contains(diagnostic, "guest") {
			t.Fatalf("unsafe or missing diagnostic tail (chunk %d)", chunkSize)
		}
	}
}

func TestGenericStartupIgnoresWhitespaceOnlyDiagnostics(t *testing.T) {
	for _, chunkSize := range []int{3, 6000} {
		cause := &diagnosticFailure{}
		manager := Manager{Token: fixedToken, Process: func(_ context.Context, _ string, _ []string, _ map[string]string, _ io.Writer, stderr io.Writer) error {
			for remaining := strings.Repeat(" \t\r\n", 1250); remaining != ""; {
				size := min(chunkSize, len(remaining))
				_, _ = io.WriteString(stderr, remaining[:size])
				remaining = remaining[size:]
			}
			return cause
		}}
		lease, err := manager.Start(context.Background(), []string{"nats"})
		var typed *diagnosticFailure
		if lease != nil || !errors.Is(err, cause) || !errors.As(err, &typed) || typed != cause || err.Error() != "start nats: docker exited 125" {
			t.Fatalf("whitespace-only diagnostic changed cause-only failure (chunk %d)", chunkSize)
		}
	}
}

func TestGenericDiagnosticFailurePreservesPartialCleanup(t *testing.T) {
	for _, cause := range []error{context.Canceled, context.DeadlineExceeded} {
		caller, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)
		backend := &fakeBackend{}
		cleanupFailure := errors.New("cleanup failed")
		manager := Manager{Token: fixedToken, Probe: successfulProbe, Wait: successfulWait, Process: func(ctx context.Context, name string, args []string, env map[string]string, stdout, stderr io.Writer) error {
			if args[0] == "run" && strings.Contains(strings.Join(args, " "), "golib-postgresql-") {
				cancel()
				_, _ = io.WriteString(stderr, "decisive failure")
				return cause
			}
			if args[0] == "rm" {
				if _, ok := ctx.Deadline(); !ok || ctx.Err() != nil {
					t.Fatal("cleanup lost its fresh bounded context")
				}
				_ = backend.run(ctx, name, args, env, stdout, stderr)
				return cleanupFailure
			}
			return backend.run(ctx, name, args, env, stdout, stderr)
		}}
		lease, err := manager.Start(caller, []string{"nats", "postgresql", "redis"})
		if lease != nil || !errors.Is(err, cause) || !errors.Is(err, cleanupFailure) || !strings.Contains(err.Error(), "decisive failure") ||
			!reflect.DeepEqual(backend.removed(), []string{"golib-nats-task"}) {
			t.Fatal("diagnostic failure changed partial cleanup or joined causes")
		}
	}
}

type diagnosticFailure struct{}

func (*diagnosticFailure) Error() string { return "docker exited 125" }
