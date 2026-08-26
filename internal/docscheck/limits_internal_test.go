package docscheck

import "testing"

func TestWithinMaximumIncludesExactBoundary(t *testing.T) {
	if !withinMaximum(10, 10) || withinMaximum(11, 10) {
		t.Fatal("withinMaximum does not enforce the inclusive resource boundary")
	}
}
