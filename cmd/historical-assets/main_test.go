package main

import "testing"

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
