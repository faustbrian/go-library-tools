//go:build verifierintegration

package gates

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestTaskVerifierSessionReusesPinnedToolCompilation(t *testing.T) {
	root, session := t.TempDir(), processCacheSession(t)
	for _, name := range []string{"verifier-cache", "verifier-mod", "verifier-tmp"} {
		if err := os.Mkdir(filepath.Join(session, name), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("GOPROXY", "https://proxy.golang.org")
	t.Setenv("GOSUMDB", "sum.golang.org")
	t.Setenv("GOFLAGS", "-p=1")
	t.Setenv("GOMAXPROCS", "2")
	ctx, cancel := context.WithTimeout(t.Context(), 3*time.Minute)
	defer cancel()
	for invocation := range 2 {
		created, cleanup, err := NewProcessExecutorWithTaskCache(root, session, io.Discard, io.Discard)
		if err != nil {
			t.Fatal(err)
		}
		executor := checkedProcessCacheExecutor(t, created)
		t.Cleanup(func() {
			if _, err := os.Stat(executor.task); err == nil {
				if err := cleanup(); err != nil {
					t.Error(err)
				}
			}
		})
		var command Command
		runner := Runner{Root: root, Executor: executorFunction(func(_ context.Context, captured Command) error { command = captured; return nil })}
		if err := runner.Workflows(ctx); err != nil {
			t.Fatal(err)
		}
		// Keep the actual pinned workflow producer's classification and identity,
		// asking Go for build diagnostics and the finite executable for its version.
		command.Args = []string{"run", "-x", command.Args[1], "-version"}
		var stdout, diagnostics bytes.Buffer
		command.Stdout, command.Stderr = &stdout, &diagnostics
		started := time.Now()
		runErr := executor.Run(ctx, command)
		cleanupErr := cleanup()
		if runErr != nil || cleanupErr != nil {
			t.Fatalf("pinned tool failed: %v / %v\n%s", runErr, cleanupErr, diagnostics.String())
		}
		if !strings.Contains(stdout.String(), strings.TrimPrefix(actionlintVersion, "v")) {
			t.Fatalf("wrong pinned tool version: %q", stdout.String())
		}
		compiled := strings.Contains(diagnostics.String(), "compile -o ")
		if compiled != (invocation == 0) {
			t.Fatalf("cold/warm compilation=%v on invocation %d", compiled, invocation)
		}
		t.Logf("pinned tool invocation %d: %s; compiler execution=%v", invocation, time.Since(started).Round(time.Millisecond), compiled)
		if _, err := os.Stat(executor.task); !errors.Is(err, os.ErrNotExist) {
			t.Fatal("tool output scratch remains")
		}
	}
	entries, err := os.ReadDir(filepath.Join(session, "verifier-tmp"))
	if err != nil || len(entries) != 0 {
		t.Fatalf("pinned tool temporary files remain: %v / %v", entries, err)
	}
}
