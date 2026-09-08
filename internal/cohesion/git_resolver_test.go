package cohesion

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveGitSourceFileReadsPinnedCleanObjectAndRejectsDirtyTree(t *testing.T) {
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	runGitResolverTestCommand(t, root, "init", "-q")
	runGitResolverTestCommand(t, root, "config", "user.name", "Test")
	runGitResolverTestCommand(t, root, "config", "user.email", "test@example.invalid")
	runGitResolverTestCommand(t, root, "remote", "add", "origin", "https://github.com/faustbrian/example.git")
	if err := os.WriteFile(filepath.Join(root, "input.json"), []byte("pinned"), 0o600); err != nil {
		t.Fatal(err)
	}
	runGitResolverTestCommand(t, root, "add", "input.json")
	runGitResolverTestCommand(t, root, "commit", "-q", "-m", "test")
	revision := strings.TrimSpace(runGitResolverTestCommand(t, root, "rev-parse", "HEAD"))

	got, err := ResolveGitSourceFile(root, "github.com/faustbrian/example", revision, "input.json", 6)
	if err != nil || string(got) != "pinned" {
		t.Fatalf("ResolveGitSourceFile(clean) = %q, %v", got, err)
	}
	alternates := filepath.Join(root, ".git", "objects", "info", "alternates")
	if err := os.WriteFile(alternates, []byte(t.TempDir()+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := ResolveGitSourceFile(root, "github.com/faustbrian/example", revision, "input.json", 6); err == nil {
		t.Fatal("ResolveGitSourceFile(alternates) error = nil")
	}
	if err := os.Remove(alternates); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(t.TempDir(), "missing"), alternates); err != nil {
		t.Fatal(err)
	}
	if _, err := ResolveGitSourceFile(root, "github.com/faustbrian/example", revision, "input.json", 6); err == nil {
		t.Fatal("ResolveGitSourceFile(symlinked alternates) error = nil")
	}
	if err := os.Remove(alternates); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "input.json"), []byte("dirty"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := ResolveGitSourceFile(root, "github.com/faustbrian/example", revision, "input.json", 6); err == nil {
		t.Fatal("ResolveGitSourceFile(dirty) error = nil")
	}
	if err := os.WriteFile(filepath.Join(root, "input.json"), []byte("pinned"), 0o600); err != nil {
		t.Fatal(err)
	}
	runGitResolverTestCommand(t, root, "gc", "--prune=now")
	got, err = ResolveGitSourceFile(root, "github.com/faustbrian/example", revision, "input.json", 6)
	if err != nil || string(got) != "pinned" {
		t.Fatalf("ResolveGitSourceFile(packed clean) = %q, %v", got, err)
	}
}

func runGitResolverTestCommand(t *testing.T, directory string, args ...string) string {
	t.Helper()
	command := exec.Command("/usr/bin/git", append([]string{"-C", directory}, args...)...)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v: %s", args, err, output)
	}
	return string(output)
}
