package mutation

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"testing"
	"time"
)

// Run before direct source fixtures so non-cancelable traversal regressions
// fail within an owned hosted child rather than exhausting the campaign.
func TestSourceTraversalTerminationHosted(t *testing.T) {
	if os.Getenv("GITHUB_ACTIONS") != "true" {
		t.Skip("process-isolated traversal assertions run only in hosted CI")
	}
	for _, scenario := range []string{"empty directory", "nested matching declaration", "root alias"} {
		t.Run(scenario, func(t *testing.T) {
			if sourceTraversalOwnedChild(t) {
				return
			}
			root := t.TempDir()
			if scenario == "empty directory" {
				entries, err := readSourceEntries(root, 0)
				if err != nil || len(entries) != 0 {
					t.Fatalf("empty directory admission = %v, %v", entries, err)
				}
				return
			}
			packageDirectory, declaration := ".", "different"
			if scenario == "nested matching declaration" {
				packageDirectory, declaration = "adapters/wire", "wire"
			}
			directory := filepath.Join(root, packageDirectory)
			if err := os.MkdirAll(directory, 0o700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(directory, "source.go"), []byte("package "+declaration+"\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			source, err := SourceDigest(root, ".", packageDirectory)
			if err != nil {
				t.Fatal(err)
			}
			if err := legacyPackagePathsUnchanged(root, ".", packageDirectory, "example/v2", source); err != nil {
				t.Fatalf("finite historical path admission = %v", err)
			}
		})
	}
}

func sourceTraversalOwnedChild(t *testing.T) bool {
	t.Helper()
	const selector = "GOLIB_SOURCE_TRAVERSAL_CHILD"
	if os.Getenv(selector) == t.Name() {
		return false
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, executable, "-test.run=^"+regexp.QuoteMeta(t.Name())+"$", "-test.count=1")
	temporary := t.TempDir()
	command.Env = append(os.Environ(), selector+"="+t.Name(), "TMPDIR="+temporary, "TEMP="+temporary)
	var output bytes.Buffer
	command.Stdout, command.Stderr = &output, &output
	if err := command.Run(); err != nil {
		t.Fatalf("owned traversal assertion failed or exceeded deadline: %v; %s", err, output.String())
	}
	return true
}
