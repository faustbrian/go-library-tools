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

func TestSourceDocumentationAuditsGitOnlyRepositoriesFromNestedDirectories(t *testing.T) {
	root := t.TempDir()
	module := filepath.Join(root, "scenarios", "consumer")
	if err := os.MkdirAll(module, 0o700); err != nil {
		t.Fatal(err)
	}
	files := map[string]string{
		"go.mod": "module example.org/consumer\n\ngo 1.27.0\n",
		"doc.go": "// Package consumer holds a standalone scenario without framework metadata.\npackage consumer\n",
	}
	for name, body := range files {
		if err := os.WriteFile(filepath.Join(module, name), []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	for _, args := range [][]string{{"init", "-q"}, {"add", "scenarios/consumer/go.mod", "scenarios/consumer/doc.go"}} {
		command := exec.CommandContext(t.Context(), "git", args...)
		command.Dir = root
		if err := command.Run(); err != nil {
			t.Fatal(err)
		}
	}
	for _, cwd := range []string{root, module} {
		for _, mode := range []string{"inventory", "check"} {
			var stdout, stderr bytes.Buffer
			code := executeContext(t.Context(), []string{"source-docs", mode, "--json"}, cwd, &stdout, &stderr, nil)
			if code != 0 || stderr.Len() != 0 || !strings.Contains(stdout.String(), `"directory": "scenarios/consumer"`) || !strings.Contains(stdout.String(), `"path": "scenarios/consumer/doc.go"`) {
				t.Fatalf("Git-only %s from %s: code=%d stdout=%q stderr=%q", mode, cwd, code, stdout.String(), stderr.String())
			}
		}
	}
	var stdout, stderr bytes.Buffer
	if code := executeContext(t.Context(), []string{"config", "validate"}, root, &stdout, &stderr, nil); code != 1 || !strings.Contains(stderr.String(), ".golib.yaml not found") {
		t.Fatalf("policy validation unexpectedly bypassed: code=%d stderr=%q", code, stderr.String())
	}
}

func TestSourceDocumentationRejectsForeignConfiguredWorktree(t *testing.T) {
	root, foreign := t.TempDir(), t.TempDir()
	for _, directory := range []string{root, foreign} {
		command := exec.CommandContext(t.Context(), "git", "init", "-q")
		command.Dir = directory
		if err := command.Run(); err != nil {
			t.Fatal(err)
		}
	}
	command := exec.CommandContext(t.Context(), "git", "config", "core.worktree", foreign)
	command.Dir = root
	if err := command.Run(); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	code := executeContext(t.Context(), []string{"source-docs", "inventory", "--json"}, root, &stdout, &stderr, nil)
	if code != 1 || stdout.Len() != 0 || !strings.Contains(stderr.String(), "does not contain the working directory") {
		t.Fatalf("foreign worktree: code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
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
