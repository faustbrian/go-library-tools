package gates

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// Cached DirEntry metadata cannot reliably reproduce a missing-file failure.
// The per-operation seam uses a real Lstat failure without permission or timing
// races, while the production wrapper continues to use DirEntry.Info.
func TestGitleaksAdmissionMetadataFailureIsCategorical(t *testing.T) {
	for _, missing := range []bool{false, true} {
		t.Run(map[bool]string{false: "available", true: "missing"}[missing], func(t *testing.T) {
			root := t.TempDir()
			path := filepath.Join(root, "private-source-marker")
			if err := os.WriteFile(path, []byte("a"), 0o600); err != nil {
				t.Fatal(err)
			}
			visits, metadataCalls := 0, 0
			err := inspectGitleaksCurrentTreeWithMetadata(t.Context(), root, securitySourceLimits{entries: 1, bytes: 1}, func(string, fs.DirEntry) error {
				visits++
				return nil
			}, func(relative string, _ fs.DirEntry) (fs.FileInfo, error) {
				metadataCalls++
				if relative != filepath.Base(path) {
					t.Fatal("metadata inspected an unexpected source")
				}
				if missing {
					if removeErr := os.Remove(path); removeErr != nil {
						t.Fatal(removeErr)
					}
				}
				return os.Lstat(path)
			})
			if metadataCalls != 1 {
				t.Fatal("source metadata was not inspected")
			}
			if missing {
				if err == nil || err.Error() != "gitleaks current-tree metadata failed" || visits != 0 {
					t.Fatalf("metadata refusal: error=%v visits=%d", err, visits)
				}
			} else if err != nil || visits != 1 {
				t.Fatalf("available metadata: error=%v visits=%d", err, visits)
			}
		})
	}
}

// Failed construction owns precisely its generated root, joins cleanup failure
// with the primary refusal, and never exposes a partially usable snapshot.
func TestGitleaksAdmissionFailurePreservesOwnedCleanup(t *testing.T) {
	for _, stage := range []string{"bundle", "current"} {
		t.Run(stage, func(t *testing.T) {
			root, workspace := t.TempDir(), t.TempDir()
			input := filepath.Join(root, "source")
			if err := os.WriteFile(input, []byte("ab"), 0o600); err != nil {
				t.Fatal(err)
			}
			sentinel := filepath.Join(workspace, "keep")
			if err := os.WriteFile(sentinel, []byte("owned sentinel"), 0o600); err != nil {
				t.Fatal(err)
			}
			commandCause := errors.New("private-command-marker")
			cleanupCause := errors.New("cleanup failure")
			var steps []string
			cleanupCalls := 0
			runner := Runner{Root: root, Executor: workspaceExecutor{directory: workspace, run: func(_ context.Context, command Command) error {
				step := command.Args[0]
				if step == "-C" {
					step = command.Args[2]
				}
				steps = append(steps, step)
				if stage == "bundle" && step == "bundle" {
					_, _ = io.WriteString(command.Stderr, commandCause.Error())
					return commandCause
				}
				if stage == "current" && step == "fetch" {
					// Deterministic external-boundary change after admission: source
					// now exceeds its tiny budget, without modifying existing bytes.
					return os.WriteFile(filepath.Join(root, "extra"), []byte("c"), 0o600)
				}
				return nil
			}}, gitleaksSourceCleanup: func(path string) error {
				cleanupCalls++
				if filepath.Dir(path) != workspace || !strings.HasPrefix(filepath.Base(path), "gitleaks-sources-") {
					t.Fatal("cleanup escaped its generated source root")
				}
				return errors.Join(os.RemoveAll(path), cleanupCause)
			}}
			limits := securitySourceLimits{refs: 1, objects: 1, entries: 2, bytes: 2}
			runner.securitySourceLimits = &limits
			sources, cleanup, err := runner.createGitleaksSources(t.Context())
			if err == nil || sources != (gitleaksSources{}) || cleanup != nil || cleanupCalls != 1 || !errors.Is(err, cleanupCause) {
				t.Fatalf("construction refusal lost ownership or cleanup: error=%v calls=%d", err, cleanupCalls)
			}
			wantSteps := []string{"for-each-ref", "cat-file", "bundle"}
			wantError := "bounded history bundle creation failed\ncleanup failure"
			if stage == "current" {
				wantSteps = append(wantSteps, "init", "fetch")
				wantError = "gitleaks current-tree byte limit exceeded\ncleanup failure"
			} else if !errors.Is(err, commandCause) {
				t.Fatal("bundle failure lost original command identity")
			}
			if err.Error() != wantError || !reflect.DeepEqual(steps, wantSteps) {
				t.Fatalf("failure was disclosed or later effects ran: error=%v steps=%v", err, steps)
			}
			entries, readErr := os.ReadDir(workspace)
			if readErr != nil || len(entries) != 1 || entries[0].Name() != "keep" {
				t.Fatal("cleanup left partial sources or removed unrelated workspace data")
			}
			for path, want := range map[string]string{input: "ab", sentinel: "owned sentinel"} {
				data, readErr := os.ReadFile(path)
				if readErr != nil || string(data) != want {
					t.Fatal("construction refusal changed pre-existing bytes")
				}
			}
		})
	}
}
