package cli

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestSourceDocumentationInventoryAndStrictCheck(t *testing.T) {
	root := internalFixture(t)
	if err := os.WriteFile(filepath.Join(root, "missing.go"), []byte("package example\nfunc Missing() {}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"init", "-q"}, {"add", "missing.go", "go.mod"}} {
		command := exec.CommandContext(t.Context(), "git", args...)
		command.Dir = root
		if err := command.Run(); err != nil {
			t.Fatal(err)
		}
	}
	for _, mode := range []string{"inventory", "check"} {
		var stdout, stderr bytes.Buffer
		code := executeContext(t.Context(), []string{"source-docs", mode, "--json"}, root, &stdout, &stderr, nil)
		want := 0
		if mode == "check" {
			want = 1
		}
		if code != want || stderr.Len() != 0 || !strings.Contains(stdout.String(), `"code": "exported-comment"`) || !strings.Contains(stdout.String(), `"code": "package-comment"`) {
			t.Fatalf("%s: code=%d, stdout=%q, stderr=%q", mode, code, stdout.String(), stderr.String())
		}
	}
}

func TestSourceDocumentationCLIRejectsInvalidArgumentsCancellationAndOutputFailure(t *testing.T) {
	root := internalFixture(t)
	for _, args := range [][]string{{}, {"other"}, {"check", "--online"}, {"check", "--json", "extra"}} {
		var stdout, stderr bytes.Buffer
		if code := executeSourceDocumentation(t.Context(), args, root, &stdout, &stderr); code != 2 || stdout.Len() != 0 {
			t.Fatalf("invalid source-docs arguments = %d", code)
		}
	}
	command := exec.CommandContext(t.Context(), "git", "init", "-q")
	command.Dir = root
	if err := command.Run(); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"inventory"}, {"inventory", "--json"}} {
		var stderr bytes.Buffer
		if code := executeSourceDocumentation(t.Context(), args, root, errorWriter{}, &stderr); code != 1 || !strings.Contains(stderr.String(), "write source documentation inventory failed") {
			t.Fatalf("source-docs writer refusal = %d, %q", code, stderr.String())
		}
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	var stdout, stderr bytes.Buffer
	if code := executeSourceDocumentation(ctx, []string{"check", "--json"}, root, &stdout, &stderr); code != 1 || stdout.Len() != 0 {
		t.Fatalf("canceled source-docs = %d", code)
	}
}
