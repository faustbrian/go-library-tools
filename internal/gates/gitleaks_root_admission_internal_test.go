package gates

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// Faults are real OpenRoot failures at the requested path. Replacing that name
// synchronously and restoring it before returning avoids timing/permission
// races while preserving the original directory and its bytes.
func TestGitleaksRootAdmissionFailure(t *testing.T) {
	for _, stage := range []string{"source", "snapshot", "cancelled"} {
		t.Run(stage, func(t *testing.T) {
			source, workspace := t.TempDir(), t.TempDir()
			inputPath := filepath.Join(source, "source")
			sentinel := filepath.Join(workspace, "keep")
			for path, value := range map[string]string{inputPath: "ab", sentinel: "sentinel"} {
				if err := os.WriteFile(path, []byte(value), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			destination := filepath.Join(workspace, "snapshot")
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			if stage == "cancelled" {
				cancel()
			}
			rootCalls, fileCalls := 0, 0
			var openedRoots []*os.Root
			t.Cleanup(func() {
				for _, root := range openedRoots {
					_ = root.Close()
				}
			})
			files := gitleaksCopyFiles{
				openRoot: func(path string) (*os.Root, error) {
					rootCalls++
					want := source
					if rootCalls == 2 {
						want = destination
					}
					if path != want {
						t.Fatal("copy opened a root outside its selected source/snapshot")
					}
					if (stage == "source" && rootCalls == 1) || (stage == "snapshot" && rootCalls == 2) {
						held := filepath.Join(t.TempDir(), "held")
						if err := os.Rename(path, held); err != nil {
							t.Fatal(err)
						}
						defer func() {
							if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
								t.Fatal(err)
							}
							if err := os.Rename(held, path); err != nil {
								t.Fatal(err)
							}
						}()
						if err := os.WriteFile(path, []byte("private-root-marker"), 0o600); err != nil {
							t.Fatal(err)
						}
						return os.OpenRoot(path)
					}
					root, err := os.OpenRoot(path)
					if root != nil {
						openedRoots = append(openedRoots, root)
					}
					return root, err
				},
				open: func(root *os.Root, path string) (*os.File, error) {
					fileCalls++
					return root.Open(path)
				},
				create: func(root *os.Root, path string) (*os.File, error) {
					fileCalls++
					return root.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
				},
				close: (*os.File).Close,
			}
			err := copyGitleaksCurrentTreeWithFiles(ctx, source, destination, securitySourceLimits{entries: 1, bytes: 2}, files)
			wantCalls := 1
			wantError := "open gitleaks current-tree source failed"
			switch stage {
			case "snapshot":
				wantCalls = 2
				wantError = "open gitleaks current-tree snapshot failed"
			case "cancelled":
				wantCalls = 0
				wantError = "context canceled"
				if !errors.Is(err, context.Canceled) {
					t.Fatal("pre-admission cancellation lost its identity")
				}
			}
			if err == nil || err.Error() != wantError || rootCalls != wantCalls || fileCalls != 0 {
				t.Fatalf("root refusal: error=%v roots=%d files=%d", err, rootCalls, fileCalls)
			}
			for _, root := range openedRoots {
				file, err := root.Open(".")
				if file != nil {
					_ = file.Close()
				}
				if !errors.Is(err, os.ErrClosed) {
					t.Fatal("root refusal retained an open source root")
				}
			}
			for path, value := range map[string]string{inputPath: "ab", sentinel: "sentinel"} {
				data, err := os.ReadFile(path)
				if err != nil || string(data) != value {
					t.Fatal("root refusal changed original or unrelated bytes")
				}
			}
			if stage == "cancelled" {
				if _, err := os.Lstat(destination); !errors.Is(err, os.ErrNotExist) {
					t.Fatal("pre-cancelled admission created a snapshot")
				}
			} else {
				entries, err := os.ReadDir(destination)
				if err != nil || len(entries) != 0 {
					t.Fatal("root refusal partly populated its snapshot")
				}
				// The orchestration owner, not the copy operation, removes the
				// refused snapshot; unrelated workspace data remains intact.
				if err := os.Remove(destination); err != nil {
					t.Fatal(err)
				}
			}
			entries, err := os.ReadDir(workspace)
			if err != nil || len(entries) != 1 || entries[0].Name() != "keep" {
				t.Fatal("owned cleanup removed unrelated workspace data")
			}
		})
	}
}
