package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestWithinRejectsEscapesAndAcceptsDescendants(t *testing.T) {
	t.Parallel()
	if err := within("/workspace", "/workspace/assets"); err != nil {
		t.Fatal(err)
	}
	if err := within("/workspace", "/outside"); err == nil {
		t.Fatal("escape accepted")
	}
}

func TestPathContainsRequiresActualDescendant(t *testing.T) {
	t.Parallel()
	if !pathContains("/workspace", "/workspace/task") {
		t.Fatal("descendant not detected")
	}
	if pathContains("/workspace/task", "/workspace") {
		t.Fatal("ancestor detected as descendant")
	}
}

func TestSplitAbsolutePathListNormalizesAndSkipsEmptyEntries(t *testing.T) {
	paths := splitAbsolutePathList(strings.Join([]string{"relative", "", "nested"}, string(filepath.ListSeparator)))
	if len(paths) != 2 || !filepath.IsAbs(paths[0]) || !filepath.IsAbs(paths[1]) {
		t.Fatalf("splitAbsolutePathList() = %#v", paths)
	}
}

func TestSplitAbsolutePathListAndFatalfUseInjectableExit(t *testing.T) {
	old := exitProcess
	defer func() { exitProcess = old }()
	called := 0
	exitProcess = func(code int) { called = code }
	if got := splitAbsolutePathList(""); len(got) != 0 {
		t.Fatalf("splitAbsolutePathList(empty) = %#v", got)
	}
	fatalf("failure %s", "case")
	if called != 1 {
		t.Fatalf("exit code = %d, want 1", called)
	}
}
