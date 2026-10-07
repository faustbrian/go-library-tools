package gates

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func TestGitleaksHistoryEmptyAndThreeObjectBudgets(t *testing.T) {
	for _, test := range []struct {
		name, objects string
		valid         bool
	}{
		{"empty object", "0\n", true},
		{"exact with empty middle", "2\n0\n3\n", true},
		{"excess with empty middle", "2\n0\n4\n", false},
		{"three positive objects", "2\n2\n2\n", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			runner := Runner{Root: t.TempDir(), Executor: executorFunction(func(_ context.Context, command Command) error {
				value := "refs/heads/main\n"
				if command.Args[2] == "cat-file" {
					value = test.objects
				}
				_, err := io.WriteString(command.Stdout, value)
				return err
			})}
			err := runner.preflightGitleaksHistory(t.Context(), securitySourceLimits{refs: 1, objects: 3, bytes: 5})
			if test.valid {
				if err != nil {
					t.Fatalf("valid history inventory refused: %v", err)
				}
			} else if err == nil || err.Error() != "gitleaks history object or byte limit exceeded or malformed" {
				t.Fatalf("excess history inventory error = %v", err)
			}
		})
	}
}

func TestBoundedSourceFileOverflowPrefixRemainsCapped(t *testing.T) {
	path := filepath.Join(t.TempDir(), "snapshot")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	writer := &boundedSourceFile{file: file, limit: 3}
	if _, err := io.WriteString(writer, "a"); err != nil {
		t.Fatal(err)
	}
	for _, value := range []string{"bcd", "d"} {
		if _, err := io.WriteString(writer, value); err == nil || err.Error() != "history bundle byte limit exceeded" || !writer.didOverflow() {
			t.Fatalf("overflow state not preserved: %v", err)
		}
		data, err := os.ReadFile(path)
		if err != nil || string(data) != "abc" {
			t.Fatalf("overflow changed the persisted capped prefix: %v", err)
		}
	}
}

func TestGitleaksTreeEmptyMiddleKeepsCumulativeBudget(t *testing.T) {
	source := t.TempDir()
	for _, name := range []string{"a", "b", "c"} {
		if err := os.WriteFile(filepath.Join(source, name), nil, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	var visited []string
	metadataCalls := 0
	// Supply real metadata in a 1,0,1 sequence without assuming ReadDir sorts.
	err := inspectGitleaksCurrentTreeWithMetadata(t.Context(), source, securitySourceLimits{entries: 3, bytes: 1}, func(relative string, _ fs.DirEntry) error {
		visited = append(visited, relative)
		return nil
	}, func(relative string, _ fs.DirEntry) (fs.FileInfo, error) {
		data := []string{"x", "", "y"}[metadataCalls]
		metadataCalls++
		path := filepath.Join(source, relative)
		if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
			return nil, err
		}
		return os.Stat(path)
	})
	if err == nil || err.Error() != "gitleaks current-tree byte limit exceeded" || len(visited) != 2 || metadataCalls != 3 {
		t.Fatalf("tree admission error = %v, visited = %v", err, visited)
	}
	destination := filepath.Join(t.TempDir(), "snapshot")
	if err := copyGitleaksCurrentTreeBounded(t.Context(), source, destination, securitySourceLimits{entries: 3, bytes: 1}); err == nil {
		t.Fatal("excess tree produced a snapshot")
	}
	if _, err := os.Stat(destination); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("failed admission created a snapshot")
	}
	if err := inspectGitleaksCurrentTree(t.Context(), source, securitySourceLimits{entries: 3, bytes: 2}, nil); err != nil {
		t.Fatalf("inclusive tree allowance refused the empty entry: %v", err)
	}
}

func TestGitleaksCopyGrowthKeepsAggregatePersistedBudget(t *testing.T) {
	for _, growAt := range []int{1, 2} {
		t.Run(fmt.Sprintf("entry-%d", growAt+1), func(t *testing.T) {
			source := t.TempDir()
			for _, name := range []string{"a", "b", "c"} {
				if err := os.WriteFile(filepath.Join(source, name), nil, 0o600); err != nil {
					t.Fatal(err)
				}
			}
			destination := filepath.Join(t.TempDir(), "snapshot")
			var opened []string
			files := gitleaksCopyFiles{
				open: func(root *os.Root, relative string) (*os.File, error) {
					data := ""
					if len(opened) == 0 {
						data = "x"
					} else if len(opened) == growAt {
						data = "y"
					}
					opened = append(opened, relative)
					if data != "" {
						if err := os.WriteFile(filepath.Join(source, relative), []byte(data), 0o600); err != nil {
							return nil, err
						}
					}
					return root.Open(relative)
				},
				create: func(root *os.Root, relative string) (*os.File, error) {
					return root.OpenFile(relative, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
				},
				close: (*os.File).Close,
			}
			err := copyGitleaksCurrentTreeWithFiles(t.Context(), source, destination, securitySourceLimits{entries: 3, bytes: 1}, files)
			if err == nil || err.Error() != "create gitleaks current-tree snapshot: gitleaks current-tree copy failed" || len(opened) != growAt+1 {
				t.Fatalf("growth refusal error = %v", err)
			}
			for name, want := range map[string]string{opened[0]: "x", opened[growAt]: ""} {
				data, err := os.ReadFile(filepath.Join(destination, name))
				if err != nil || string(data) != want {
					t.Fatalf("persisted growth cap violated for %s: %v", name, err)
				}
			}
		})
	}
}
