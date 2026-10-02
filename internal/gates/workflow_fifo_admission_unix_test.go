//go:build darwin || linux

package gates

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

// FIFO fixtures are created and type-inspected only, never opened or consumed.
// Both supported descriptor suffixes must refuse them before validator effects.
func TestWorkflowFIFOAdmissionBeforeValidator(t *testing.T) {
	for _, name := range []string{"ci.yml", "ci.yaml"} {
		for _, kind := range []string{"fifo", "ordinary"} {
			t.Run(name+"/"+kind, func(t *testing.T) {
				root := t.TempDir()
				directory := filepath.Join(root, ".github", "workflows")
				if err := os.MkdirAll(directory, 0o700); err != nil {
					t.Fatal(err)
				}
				path := filepath.Join(directory, name)
				const content = "jobs: {test: {steps: [{uses: owner/action@0123456789012345678901234567890123456789}]}}\n"
				if kind == "fifo" {
					if err := syscall.Mkfifo(path, 0o600); err != nil {
						t.Fatal(err)
					}
				} else if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
					t.Fatal(err)
				}
				sentinel := filepath.Join(root, "unrelated")
				if err := os.WriteFile(sentinel, []byte("keep"), 0o600); err != nil {
					t.Fatal(err)
				}
				before, err := os.Lstat(path)
				if err != nil {
					t.Fatal(err)
				}
				if kind == "fifo" && before.Mode()&os.ModeNamedPipe == 0 {
					t.Fatal("fixture is not a FIFO")
				}
				calls := 0
				var output bytes.Buffer
				runner := Runner{Root: root, Output: &output, Executor: executorFunction(func(context.Context, Command) error {
					calls++
					return nil
				})}
				err = runner.Workflows(t.Context())
				if kind == "fifo" {
					if err == nil || err.Error() != "workflow security policy: workflow security policy: unsupported descriptor file type" || calls != 0 || output.Len() != 0 {
						t.Fatalf("FIFO admission: error=%v calls=%d output bytes=%d", err, calls, output.Len())
					}
				} else {
					if err != nil || calls != 1 || output.String() != "workflow contract passed\n" {
						t.Fatalf("ordinary admission: error=%v calls=%d output=%q", err, calls, output.String())
					}
					data, readErr := os.ReadFile(path)
					if readErr != nil || string(data) != content {
						t.Fatal("admission changed ordinary workflow bytes")
					}
				}
				after, err := os.Lstat(path)
				if err != nil || !os.SameFile(before, after) || before.Mode() != after.Mode() {
					t.Fatal("admission changed descriptor identity or type")
				}
				data, err := os.ReadFile(sentinel)
				if err != nil || string(data) != "keep" {
					t.Fatal("admission changed unrelated source bytes")
				}
			})
		}
	}
}
