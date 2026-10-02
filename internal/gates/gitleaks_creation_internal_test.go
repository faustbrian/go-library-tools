package gates

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestGitleaksCreationFailurePrivacyAndCleanup(t *testing.T) {
	for _, stage := range []string{"ignore", "bundle"} {
		for _, cleanupFails := range []bool{false, true} {
			t.Run(stage+map[bool]string{false: "/cleanup-success", true: "/cleanup-failure"}[cleanupFails], func(t *testing.T) {
				root, workspace := t.TempDir(), t.TempDir()
				input, keep := filepath.Join(root, "source"), filepath.Join(workspace, "keep")
				for path, value := range map[string]string{input: "a", keep: "unchanged"} {
					if err := os.WriteFile(path, []byte(value), 0o600); err != nil {
						t.Fatal(err)
					}
				}
				const marker = "private-creation-marker"
				operationCause := errors.New(marker)
				operationError := &os.PathError{Op: "create", Path: marker, Err: operationCause}
				cleanupCause := errors.New("cleanup failure")
				var commands []string
				var freshRoot string
				cleanupCalls, opens := 0, 0
				runner := Runner{Root: root, Executor: workspaceExecutor{directory: workspace, run: func(_ context.Context, command Command) error {
					commands = append(commands, command.Args[2])
					return nil
				}}, gitleaksSourceCleanup: func(path string) error {
					cleanupCalls++
					if path != freshRoot || filepath.Dir(path) != workspace || !strings.HasPrefix(filepath.Base(path), "gitleaks-sources-") {
						t.Fatal("cleanup escaped its exact freshly created source root")
					}
					err := os.RemoveAll(path)
					if cleanupFails {
						return errors.Join(err, cleanupCause)
					}
					return err
				}}
				files := gitleaksCreationFiles{mkdir: func(path string, mode os.FileMode) error {
					freshRoot = filepath.Dir(path)
					if filepath.Base(path) != "ignore" || mode != 0o700 {
						t.Fatal("ignore creation changed ownership or permissions")
					}
					if stage == "ignore" {
						return operationError
					}
					return os.Mkdir(path, mode)
				}, openFile: func(path string, flags int, mode os.FileMode) (*os.File, error) {
					opens++
					if filepath.Dir(path) != freshRoot || filepath.Base(path) != "history.bundle" || flags != os.O_CREATE|os.O_EXCL|os.O_WRONLY || mode != 0o600 {
						t.Fatal("bundle creation escaped its owner or exclusive private permissions")
					}
					return nil, operationError
				}}
				sources, cleanup, err := runner.createGitleaksSourcesWithFiles(t.Context(), files)
				if err == nil || sources != (gitleaksSources{}) || cleanup != nil || cleanupCalls != 1 {
					t.Fatalf("failed construction exposed partial sources: error=%v cleanup calls=%d", err, cleanupCalls)
				}
				want := "create bounded history bundle"
				if stage == "ignore" {
					want = "create gitleaks ignore root"
				}
				if cleanupFails {
					want += "\ncleanup failure"
				}
				if err.Error() != want || strings.Contains(err.Error(), marker) {
					t.Errorf("creation failure leaked private cause or changed category: %v", err)
				}
				if stage == "ignore" {
					var original *os.PathError
					if opens != 0 || !errors.Is(err, operationCause) || !errors.As(err, &original) || original != operationError {
						t.Fatal("ignore failure lost trusted cause or attempted bundle creation")
					}
				} else if opens != 1 || errors.Is(err, operationCause) {
					t.Fatal("bundle failure changed deliberate cause discard")
				}
				if errors.Is(err, cleanupCause) != cleanupFails || !reflect.DeepEqual(commands, []string{"for-each-ref", "cat-file"}) {
					t.Fatal("creation failure lost cleanup identity or executed later commands")
				}
				entries, readErr := os.ReadDir(workspace)
				if readErr != nil || len(entries) != 1 || entries[0].Name() != "keep" {
					t.Fatal("failed construction leaked or removed unrelated resources")
				}
				for path, want := range map[string]string{input: "a", keep: "unchanged"} {
					data, readErr := os.ReadFile(path)
					if readErr != nil || string(data) != want {
						t.Fatal("construction changed caller bytes")
					}
				}
			})
		}
	}
}

func TestGitleaksCreationRealHandlesAndCleanup(t *testing.T) {
	root, workspace := t.TempDir(), t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "source"), []byte("a"), 0o600); err != nil {
		t.Fatal(err)
	}
	var bundle *os.File
	var commands []string
	cleanupCalls := 0
	runner := Runner{Root: root, Executor: workspaceExecutor{directory: workspace, run: func(_ context.Context, command Command) error {
		step := command.Args[0]
		if step == "-C" {
			step = command.Args[2]
		}
		commands = append(commands, step)
		if step == "bundle" {
			_, err := io.WriteString(command.Stdout, "b")
			return err
		}
		return nil
	}}, gitleaksSourceCleanup: func(path string) error { cleanupCalls++; return os.RemoveAll(path) }}
	sources, cleanup, err := runner.createGitleaksSourcesWithFiles(t.Context(), gitleaksCreationFiles{
		mkdir: os.Mkdir, openFile: func(path string, flags int, mode os.FileMode) (*os.File, error) {
			var err error
			bundle, err = os.OpenFile(path, flags, mode)
			return bundle, err
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if cleanup == nil {
		t.Fatal("successful construction lost cleanup ownership")
	}
	t.Cleanup(func() { _ = os.RemoveAll(sources.root) })
	if _, err := bundle.Stat(); !errors.Is(err, os.ErrClosed) {
		t.Fatal("successful construction retained writable bundle handle")
	}
	if !reflect.DeepEqual(commands, []string{"for-each-ref", "cat-file", "bundle", "init", "fetch"}) {
		t.Fatal("normal composition changed command order")
	}
	data, err := os.ReadFile(filepath.Join(sources.current, "source"))
	if err != nil || string(data) != "a" {
		t.Fatal("current snapshot did not preserve source bytes")
	}
	if err := cleanup(); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(workspace)
	if err != nil || len(entries) != 0 || cleanupCalls != 1 {
		t.Fatal("successful construction did not remove exactly its owned sources")
	}
}
