//go:build darwin || linux

package gates

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"syscall"
	"testing"
)

// Create/type-inspect only: the FIFO is never opened, read, written or listened
// on. Real unsupported filesystem entries must stop admission before effects.
func TestGitleaksFIFOAdmissionRefusesBeforeEffects(t *testing.T) {
	source, workspace := t.TempDir(), t.TempDir()
	path := filepath.Join(source, "private-fifo-marker")
	if err := syscall.Mkfifo(path, 0o600); err != nil {
		t.Fatal(err)
	}
	limits := securitySourceLimits{refs: 1, objects: 1, entries: 1, bytes: 1}
	visited := 0
	err := inspectGitleaksCurrentTree(t.Context(), source, limits, func(string, fs.DirEntry) error {
		visited++
		return nil
	})
	if err == nil || err.Error() != "gitleaks current-tree unsupported entry type" || visited != 0 {
		t.Fatalf("unsupported entry reached copying: error=%v visits=%d", err, visited)
	}
	destination := filepath.Join(workspace, "snapshot")
	err = copyGitleaksCurrentTreeBounded(t.Context(), source, destination, limits)
	if err == nil || err.Error() != "gitleaks current-tree unsupported entry type" {
		t.Fatalf("copy admitted unsupported entry: %v", err)
	}
	if _, err := os.Lstat(destination); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("unsupported entry created a snapshot")
	}
	var commands []string
	runner := Runner{Root: source, securitySourceLimits: &limits, Executor: workspaceExecutor{directory: workspace, run: func(_ context.Context, command Command) error {
		commands = append(commands, command.Args[2])
		return nil
	}}}
	sources, cleanup, err := runner.createGitleaksSources(t.Context())
	if cleanup != nil {
		_ = cleanup()
	}
	if err == nil || err.Error() != "gitleaks current-tree unsupported entry type" || cleanup != nil || sources != (gitleaksSources{}) {
		t.Fatalf("unsupported construction exposed usable sources: %v", err)
	}
	if !reflect.DeepEqual(commands, []string{"for-each-ref", "cat-file"}) {
		t.Fatalf("unsupported entry reached artifact execution: %v", commands)
	}
	entries, err := os.ReadDir(workspace)
	if err != nil || len(entries) != 0 {
		t.Fatal("unsupported admission changed the task workspace")
	}
	info, err := os.Lstat(path)
	if err != nil || info.Mode()&os.ModeNamedPipe == 0 {
		t.Fatal("unsupported admission changed the original FIFO")
	}
}
