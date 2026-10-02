package gates

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// All faults below come from actual os.Root/os.File operations, not fabricated
// metadata or scheduling. Errors remain categorical, but cancellation survives.
func TestGitleaksCopyLifecycleFailures(t *testing.T) {
	for _, stage := range []string{"source-open", "destination-create", "read", "write", "input-close", "output-close", "both-close", "cancel"} {
		t.Run(stage, func(t *testing.T) {
			source, outside := t.TempDir(), t.TempDir()
			inputPath := filepath.Join(source, "private-input-marker")
			if err := os.WriteFile(inputPath, []byte("ab"), 0o600); err != nil {
				t.Fatal(err)
			}
			sentinel := filepath.Join(outside, "keep")
			if err := os.WriteFile(sentinel, []byte("sentinel"), 0o600); err != nil {
				t.Fatal(err)
			}
			destination := filepath.Join(outside, "snapshot")
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			var input, output *os.File
			t.Cleanup(func() {
				for _, file := range []*os.File{input, output} {
					if file != nil {
						_ = file.Close()
					}
				}
			})
			inputCloses, outputCloses := 0, 0
			files := gitleaksCopyFiles{
				open: func(root *os.Root, relative string) (*os.File, error) {
					if stage == "source-open" {
						if err := root.Close(); err != nil {
							t.Fatal(err)
						}
					}
					file, err := root.Open(relative)
					input = file
					if err == nil && stage == "read" {
						if err := file.Close(); err != nil {
							t.Fatal(err)
						}
					}
					return file, err
				},
				create: func(root *os.Root, relative string) (*os.File, error) {
					if stage == "destination-create" {
						// A real exclusive-create collision must not overwrite bytes.
						file, err := root.OpenFile(relative, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
						if err != nil {
							t.Fatal(err)
						}
						if _, err := file.WriteString("keep destination"); err != nil {
							t.Fatal(err)
						}
						if err := file.Close(); err != nil {
							t.Fatal(err)
						}
					}
					file, err := root.OpenFile(relative, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
					output = file
					if err == nil && stage == "write" {
						if err := file.Close(); err != nil {
							t.Fatal(err)
						}
					}
					if stage == "cancel" {
						cancel()
					}
					return file, err
				},
				close: func(file *os.File) error {
					isInput := file == input
					if isInput {
						inputCloses++
					} else if file == output {
						outputCloses++
					} else {
						t.Fatal("closed an unrelated file")
					}
					err := file.Close()
					if stage == "both-close" || (stage == "input-close" && isInput) || (stage == "output-close" && !isInput) {
						// A second real close fails after the owned descriptor was
						// released, permitting a deterministic close-only failure.
						return errors.Join(err, file.Close())
					}
					return err
				},
			}
			err := copyGitleaksCurrentTreeWithFiles(ctx, source, destination, securitySourceLimits{entries: 1, bytes: 2}, files)
			want := "create gitleaks current-tree snapshot: gitleaks current-tree copy failed"
			wantInputCloses, wantOutputCloses := 1, 1
			switch stage {
			case "source-open":
				want = "create gitleaks current-tree snapshot: gitleaks current-tree source open failed"
				wantInputCloses, wantOutputCloses = 0, 0
			case "destination-create":
				want = "create gitleaks current-tree snapshot: gitleaks current-tree destination open failed"
				wantOutputCloses = 0
			case "cancel":
				want += "\ncontext canceled"
				if !errors.Is(err, context.Canceled) {
					t.Fatal("copy lost cancellation identity")
				}
			}
			if err == nil || err.Error() != want || inputCloses != wantInputCloses || outputCloses != wantOutputCloses {
				t.Fatalf("copy failure: error=%v closes=%d/%d", err, inputCloses, outputCloses)
			}
			for _, file := range []*os.File{input, output} {
				if file != nil {
					if _, err := file.Stat(); !errors.Is(err, os.ErrClosed) {
						t.Fatal("copy failure retained an open owned descriptor")
					}
				}
			}
			for path, value := range map[string]string{inputPath: "ab", sentinel: "sentinel"} {
				data, err := os.ReadFile(path)
				if err != nil || string(data) != value {
					t.Fatal("copy failure changed source or unrelated bytes")
				}
			}
			wantBytes := ""
			if stage == "destination-create" {
				wantBytes = "keep destination"
			} else if stage == "input-close" || stage == "output-close" || stage == "both-close" {
				wantBytes = "ab"
			}
			if stage != "source-open" {
				data, err := os.ReadFile(filepath.Join(destination, filepath.Base(inputPath)))
				if err != nil || string(data) != wantBytes {
					t.Fatal("copy failure overwrote existing or misreported partial bytes")
				}
			}
		})
	}
}

// Metadata was admitted before open. Synchronous growth at that boundary must
// still persist only the admitted prefix and return a categorical failure.
func TestGitleaksCopyLifecycleGrowthStopsAtBudget(t *testing.T) {
	source := t.TempDir()
	path := filepath.Join(source, "source")
	if err := os.WriteFile(path, []byte("ab"), 0o600); err != nil {
		t.Fatal(err)
	}
	destination := filepath.Join(t.TempDir(), "snapshot")
	files := gitleaksCopyFiles{
		open: func(root *os.Root, relative string) (*os.File, error) {
			// This deterministic fixture mutation is not an effect of the copier.
			if err := os.WriteFile(path, []byte("abc"), 0o600); err != nil {
				t.Fatal(err)
			}
			return root.Open(relative)
		},
		create: func(root *os.Root, relative string) (*os.File, error) {
			return root.OpenFile(relative, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
		},
		close: (*os.File).Close,
	}
	err := copyGitleaksCurrentTreeWithFiles(t.Context(), source, destination, securitySourceLimits{entries: 1, bytes: 2}, files)
	if err == nil || err.Error() != "create gitleaks current-tree snapshot: gitleaks current-tree copy failed" {
		t.Fatalf("growth was not refused categorically: %v", err)
	}
	for path, want := range map[string]string{path: "abc", filepath.Join(destination, "source"): "ab"} {
		data, err := os.ReadFile(path)
		if err != nil || string(data) != want {
			t.Fatal("growth refusal exceeded its budget or changed fixture bytes")
		}
	}
}

// Real readback proves inclusive cumulative bounds, independent of directory
// enumeration order and writer return counts.
func TestGitleaksCopyLifecycleBudgets(t *testing.T) {
	for _, budget := range []int64{2, 3, 4} {
		t.Run(map[int64]string{2: "over", 3: "exact", 4: "below"}[budget], func(t *testing.T) {
			source := t.TempDir()
			contents := map[string]string{"first": "a", "second": "bc"}
			for name, value := range contents {
				if err := os.WriteFile(filepath.Join(source, name), []byte(value), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			destination := filepath.Join(t.TempDir(), "snapshot")
			err := copyGitleaksCurrentTreeBounded(t.Context(), source, destination, securitySourceLimits{entries: 2, bytes: budget})
			if budget == 2 {
				if err == nil || err.Error() != "gitleaks current-tree byte limit exceeded" {
					t.Fatalf("over-budget admission: %v", err)
				}
				if _, err := os.Lstat(destination); !errors.Is(err, os.ErrNotExist) {
					t.Fatal("refused tree acquired a snapshot")
				}
			} else if err != nil {
				t.Fatal(err)
			}
			var copied int64
			for name, value := range contents {
				original, err := os.ReadFile(filepath.Join(source, name))
				if err != nil || string(original) != value {
					t.Fatal("copy changed original source bytes")
				}
				if budget == 2 {
					continue
				}
				path := filepath.Join(destination, name)
				data, err := os.ReadFile(path)
				if err != nil || string(data) != value {
					t.Fatal("snapshot changed admitted bytes")
				}
				info, err := os.Stat(path)
				if err != nil || info.Mode().Perm() != 0o600 {
					t.Fatal("snapshot did not use private regular-file permissions")
				}
				copied += int64(len(data))
			}
			if budget >= 3 && copied != 3 {
				t.Fatal("cumulative snapshot bytes differ from admitted bytes")
			}
		})
	}
}
