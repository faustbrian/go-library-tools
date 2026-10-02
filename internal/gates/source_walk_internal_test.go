package gates

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestSecuritySourceWalkPropagatesChangedChildFailure(t *testing.T) {
	for _, test := range []struct {
		name    string
		replace bool
		message string
	}{
		{"removed directory", false, "source directory unavailable"},
		{"directory replaced with file", true, "source directory read failed"},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			child := filepath.Join(root, "application-private-child")
			if err := os.Mkdir(child, 0o700); err != nil {
				t.Fatal(err)
			}
			visits := 0
			err := walkSecuritySource(context.Background(), root, 2, func(relative string, entry fs.DirEntry) error {
				visits++
				if relative != "application-private-child" || !entry.IsDir() {
					t.Fatalf("unexpected visit: %q, directory=%v", relative, entry.IsDir())
				}
				if err := os.Remove(child); err != nil {
					t.Fatal(err)
				}
				if test.replace {
					if err := os.WriteFile(child, []byte("ordinary"), 0o600); err != nil {
						t.Fatal(err)
					}
				}
				return nil
			})
			if err == nil || err.Error() != test.message || visits != 1 {
				t.Fatalf("walk = %v, visits=%d; want %q and one visit", err, visits, test.message)
			}
		})
	}
}

func TestSecuritySourceWalkCancellationBetweenEntries(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"first", "second"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte("ordinary"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	visits := 0
	err := walkSecuritySource(ctx, root, 2, func(_ string, _ fs.DirEntry) error {
		visits++
		cancel()
		return nil
	})
	if !errors.Is(err, context.Canceled) || reflect.TypeOf(err) != reflect.TypeOf(context.Canceled) || visits != 1 {
		t.Fatalf("walk = %v, visits=%d; want original cancellation and one visit", err, visits)
	}
}

// Depth counts directory traversal from root=0, not relative-path components.
// The blocked directory can be visited, but its contents must remain unvisited.
func TestSecuritySourceWalkDepthAdmissionBoundary(t *testing.T) {
	for _, test := range []struct {
		name                     string
		depth, limit, wantVisits int
		wantError                string
	}{
		{"depth128-admitted", 128, 129, 129, ""},
		{"depth129-refused", 129, 130, 129, "source traversal depth limit exceeded"},
		{"entry-budget-refused-first", 129, 128, 128, "source traversal entry limit exceeded"},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			relative := ""
			expected := make([]string, 0, test.depth+1)
			for range test.depth {
				relative = filepath.Join(relative, "d")
				if err := os.Mkdir(filepath.Join(root, relative), 0o700); err != nil {
					t.Fatal(err)
				}
				expected = append(expected, relative)
			}
			leaf := filepath.Join(relative, "private-source")
			path := filepath.Join(root, leaf)
			if err := os.WriteFile(path, []byte("keep"), 0o600); err != nil {
				t.Fatal(err)
			}
			expected = append(expected, leaf)
			var visited []string
			leafVisited := false
			err := walkSecuritySource(t.Context(), root, test.limit, func(relative string, entry fs.DirEntry) error {
				visited = append(visited, relative)
				if relative == leaf {
					leafVisited = true
					if !entry.Type().IsRegular() {
						t.Fatal("admitted leaf was not an ordinary file")
					}
				} else if !entry.IsDir() {
					t.Fatal("walk visited an unexpected non-directory")
				}
				return nil
			})
			if test.wantError == "" {
				if err != nil || !leafVisited {
					t.Fatalf("depth128 admission: error=%v leaf visited=%v", err, leafVisited)
				}
			} else if err == nil || err.Error() != test.wantError || leafVisited {
				t.Fatalf("bounded refusal: error=%v leaf visited=%v", err, leafVisited)
			}
			if !reflect.DeepEqual(visited, expected[:test.wantVisits]) {
				t.Fatalf("walk did not stop at the expected visit boundary: visits=%d want=%d", len(visited), test.wantVisits)
			}
			data, err := os.ReadFile(path)
			if err != nil || string(data) != "keep" {
				t.Fatal("bounded traversal changed source bytes")
			}
		})
	}
}
