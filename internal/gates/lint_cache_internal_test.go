package gates

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestLintCacheWritesAreInvocationOwnedAndCleaned(t *testing.T) {
	parentCache := t.TempDir()
	t.Setenv("GOLANGCI_LINT_CACHE", parentCache)
	t.Setenv("GOLANGCI_LINT_CACHEPROG", "unused-parent-cache-program")
	parentMarker := filepath.Join(parentCache, "parent")
	if err := os.WriteFile(parentMarker, []byte("preserve"), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, borrowed := range []bool{false, true} {
		name := "private"
		if borrowed {
			name = "borrowed-go-cache"
		}
		t.Run(name, func(t *testing.T) {
			repository := t.TempDir()
			var session string
			if borrowed {
				created, err := os.MkdirTemp("", "golib-cache-session.")
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() {
					if err := os.RemoveAll(created); err != nil {
						t.Error(err)
					}
				})
				session, err = filepath.EvalSymlinks(created)
				if err != nil {
					t.Fatal(err)
				}
				for _, directory := range []string{"cache", "mod"} {
					if err := os.Mkdir(filepath.Join(session, directory), 0o700); err != nil {
						t.Fatal(err)
					}
				}
			}
			var previous string
			for range 2 {
				var executor Executor
				var cleanup func() error
				var err error
				if borrowed {
					executor, cleanup, err = NewProcessExecutorWithTaskCache(repository, session, io.Discard, io.Discard)
				} else {
					executor, cleanup, err = NewProcessExecutor(repository, io.Discard, io.Discard)
				}
				if err != nil {
					t.Fatal(err)
				}
				cleaned := false
				t.Cleanup(func() {
					if !cleaned {
						if err := cleanup(); err != nil {
							t.Error(err)
						}
					}
				})
				probe := filepath.Join(t.TempDir(), "cache-path")
				if err := executor.Run(context.Background(), Command{
					Name: os.Args[0], Args: []string{"-test.run=^TestLintCacheWriteHelper$"},
					Env: map[string]string{"GOLIB_LINT_CACHE_PROBE": probe},
				}); err != nil {
					t.Fatal(err)
				}
				data, err := os.ReadFile(probe)
				if err != nil {
					t.Fatal(err)
				}
				cache := string(data)
				workspace, ok := executor.(taskWorkspace)
				if !ok || filepath.Dir(cache) != workspace.TemporaryDirectory() || cache == previous {
					t.Fatal("native lint cache write escaped invocation ownership")
				}
				previous = cache
				if info, err := os.Stat(cache); err != nil || info.Mode().Perm() != 0o700 {
					t.Fatalf("private cache permissions: %v, %v", info, err)
				}
				if err := cleanup(); err != nil {
					t.Fatal(err)
				}
				cleaned = true
				if _, err := os.Stat(filepath.Join(cache, "entry")); !errors.Is(err, os.ErrNotExist) {
					t.Fatalf("lint cache entry remains after cleanup: %v", err)
				}
				if borrowed {
					for _, directory := range []string{"cache", "mod"} {
						if _, err := os.Stat(filepath.Join(session, directory)); err != nil {
							t.Fatalf("borrowed compilation cache was removed: %v", err)
						}
					}
				}
			}
		})
	}
	if data, err := os.ReadFile(parentMarker); err != nil || string(data) != "preserve" {
		t.Fatalf("parent cache changed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(parentCache, "entry")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("native helper wrote into parent cache: %v", err)
	}
}

// TestLintCacheWriteHelper performs one finite cache write in a child test
// binary. It never launches descendants or exercises process-control paths.
func TestLintCacheWriteHelper(t *testing.T) {
	probe := os.Getenv("GOLIB_LINT_CACHE_PROBE")
	if probe == "" {
		return
	}
	cache := os.Getenv("GOLANGCI_LINT_CACHE")
	if os.Getenv("GOLANGCI_LINT_CACHEPROG") != "" {
		t.Fatal("external lint-cache backend remained active")
	}
	if err := os.WriteFile(filepath.Join(cache, "entry"), []byte("lint-result"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(probe, []byte(cache), 0o600); err != nil {
		t.Fatal(err)
	}
}
