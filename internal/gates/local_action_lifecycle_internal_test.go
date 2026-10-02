package gates

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The scoped workflow I/O owner also governs local descriptors. Real handles
// establish cleanup; injected operation failures remain private categories.
func TestLocalActionDescriptorIOLifecycle(t *testing.T) {
	for _, stage := range []string{"stat", "root", "open", "read", "close", "accepted"} {
		t.Run(stage, func(t *testing.T) {
			rootPath := t.TempDir()
			const descriptor = "runs: {using: composite, steps: []}\n"
			workflowRefusalWrite(t, rootPath, "ordinary/action.yml", descriptor)
			const marker = "private-local-io-marker"
			cause := errors.New(marker)
			files := operatingWorkflowDescriptorFiles()
			var root *os.Root
			var file *os.File
			rootCalls, openCalls, closeCalls := 0, 0, 0
			files.stat = func(path string) (os.FileInfo, error) {
				if stage == "stat" {
					return nil, cause
				}
				return os.Stat(path)
			}
			files.openRoot = func(path string) (*os.Root, error) {
				rootCalls++
				if stage == "root" {
					return nil, cause
				}
				var err error
				root, err = os.OpenRoot(path)
				return root, err
			}
			files.open = func(owner *os.Root, name string) (*os.File, error) {
				openCalls++
				if stage == "open" {
					return nil, cause
				}
				var err error
				file, err = owner.Open(name)
				if err == nil && stage == "read" {
					err = file.Close()
				}
				return file, err
			}
			files.close = func(owned *os.File) error {
				closeCalls++
				err := owned.Close()
				if stage == "close" {
					return errors.Join(err, cause)
				}
				return err
			}
			t.Cleanup(func() {
				if file != nil {
					_ = file.Close()
				}
				if root != nil {
					_ = root.Close()
				}
			})
			inspection := localActionInspection{ctx: t.Context(), root: rootPath,
				active: map[string]bool{}, complete: map[string]bool{}, fileOperations: &files}
			err := inspection.local("./ordinary", 1)
			want := map[string]string{"stat": "local action path unavailable", "root": "local action root unavailable",
				"open": "local action descriptor unavailable", "read": "local action descriptor read failed or oversized",
				"close": "local action descriptor read failed or oversized"}[stage]
			if stage == "accepted" {
				if err != nil || inspection.files != 1 || inspection.bytes != len(descriptor) || !inspection.complete["ordinary"] {
					t.Fatalf("normal descriptor result: error=%v files=%d bytes=%d", err, inspection.files, inspection.bytes)
				}
			} else if err == nil || err.Error() != want || strings.Contains(err.Error(), marker) || inspection.bytes != 0 || len(inspection.complete) != 0 || len(inspection.active) != 0 {
				t.Fatalf("private descriptor refusal: error=%v bytes=%d complete=%v active=%v", err, inspection.bytes, inspection.complete, inspection.active)
			}
			if stage == "stat" && rootCalls != 0 || stage == "root" && openCalls != 0 || stage == "open" && closeCalls != 0 {
				t.Fatal("failed operation allowed later I/O")
			}
			if file != nil {
				if closeCalls != 1 {
					t.Fatalf("descriptor close calls=%d", closeCalls)
				}
				if _, err := file.Stat(); !errors.Is(err, os.ErrClosed) {
					t.Fatal("descriptor handle remained open")
				}
			}
			if root != nil {
				if _, err := root.Stat("ordinary/action.yml"); !errors.Is(err, os.ErrClosed) {
					t.Fatal("confined root remained open")
				}
			}
			data, readErr := os.ReadFile(filepath.Join(rootPath, "ordinary/action.yml"))
			if readErr != nil || string(data) != descriptor {
				t.Fatal("I/O inspection changed caller descriptor")
			}
		})
	}
}
