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
	if !errors.Is(err, context.Canceled) || !reflect.DeepEqual(err, context.Canceled) || visits != 1 {
		t.Fatalf("walk = %v, visits=%d; want original cancellation and one visit", err, visits)
	}
}
