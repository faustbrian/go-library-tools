package gates

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLocalActionPathRefusalBeforeValidator(t *testing.T) {
	for _, test := range []struct{ name, reference, reason string }{
		{"nonlocal", "./../outside", "local action path is not repository-contained and canonical"},
		{"noncanonical", "./ordinary//child", "local action path is not repository-contained and canonical"},
		{"symbolic ancestor", "./symbolic", "local action path missing or symbolic"},
	} {
		t.Run(test.name, func(t *testing.T) {
			parent := t.TempDir()
			root := filepath.Join(parent, "source")
			workflowRefusalWrite(t, root, ".github/workflows/ci.yml", "jobs: {test: {steps: [{uses: '"+test.reference+"'}]}}")
			const descriptor = "runs: {using: composite, steps: []}\n"
			workflowRefusalWrite(t, parent, "outside/action.yml", descriptor)
			workflowRefusalWrite(t, root, "ordinary/child/action.yml", descriptor)
			if err := os.Symlink(filepath.Join(root, "ordinary", "child"), filepath.Join(root, "symbolic")); err != nil {
				t.Fatal(err)
			}
			calls := 0
			var output bytes.Buffer
			runner := Runner{Root: root, Output: &output, Executor: executorFunction(func(context.Context, Command) error { calls++; return nil })}
			err := runner.Workflows(t.Context())
			want := "workflow security policy: " + test.reason
			if err == nil || err.Error() != want || calls != 0 || output.Len() != 0 {
				t.Fatalf("path refusal = %v, calls=%d, output=%q; want %q without effects", err, calls, output.String(), want)
			}
		})
	}
}

// Precharged budgets represent preceding descriptors in this same graph.
// Real confined handles and YAML parsing still own the next descriptor.
func TestLocalActionInclusiveGraphLimits(t *testing.T) {
	const descriptor = "runs: {using: composite, steps: []}\n"
	for _, test := range []struct {
		name         string
		depth, files int
		bytes        int
		want         string
	}{
		{name: "exact depth", depth: maximumLocalActionDepth},
		{name: "excess depth", depth: maximumLocalActionDepth + 1, want: "local action depth limit exceeded"},
		{name: "exact files", depth: 1, files: maximumWorkflowFiles - 1},
		{name: "excess files", depth: 1, files: maximumWorkflowFiles, want: "local action descriptor count limit exceeded"},
		{name: "exact bytes", depth: 1, bytes: maximumWorkflowBytes - len(descriptor)},
		{name: "excess bytes", depth: 1, bytes: maximumWorkflowBytes - len(descriptor) + 1, want: "workflow descriptor total byte limit exceeded"},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			workflowRefusalWrite(t, root, "ordinary/action.yml", descriptor)
			inspection := localActionInspection{ctx: t.Context(), root: root, active: map[string]bool{}, complete: map[string]bool{}, files: test.files, bytes: test.bytes}
			err := inspection.local("./ordinary", test.depth)
			if test.want != "" {
				if err == nil || err.Error() != test.want || len(inspection.active) != 0 || len(inspection.complete) != 0 {
					t.Fatalf("graph refusal = %v, active=%v, complete=%v; want %q", err, inspection.active, inspection.complete, test.want)
				}
				return
			}
			if err != nil || inspection.files != test.files+1 || inspection.bytes != test.bytes+len(descriptor) || len(inspection.active) != 0 || !inspection.complete["ordinary"] {
				t.Fatalf("inclusive graph admission = %v, files=%d, bytes=%d, active=%v, complete=%v", err, inspection.files, inspection.bytes, inspection.active, inspection.complete)
			}
		})
	}
}

func TestLocalActionInclusiveDescriptorSize(t *testing.T) {
	const prefix = "runs: {using: composite, steps: []}\n#"
	const suffix = "\nname: \"x\""
	for _, extra := range []int{0, 1} {
		name := "exact bytes"
		if extra != 0 {
			name = "excess bytes"
		}
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			descriptor := prefix + strings.Repeat("a", maximumWorkflowOutput-len(prefix)-len(suffix)+extra) + suffix
			workflowRefusalWrite(t, root, "ordinary/action.yml", descriptor)
			inspection := localActionInspection{ctx: t.Context(), root: root, active: map[string]bool{}, complete: map[string]bool{}}
			err := inspection.local("./ordinary", 1)
			if extra != 0 {
				if err == nil || err.Error() != "local action descriptor missing, symbolic, or oversized" || inspection.files != 0 || inspection.bytes != 0 || len(inspection.complete) != 0 {
					t.Fatalf("oversized descriptor = %v, files=%d, bytes=%d, complete=%v", err, inspection.files, inspection.bytes, inspection.complete)
				}
				return
			}
			if err != nil || inspection.files != 1 || inspection.bytes != maximumWorkflowOutput || !inspection.complete["ordinary"] {
				t.Fatalf("inclusive descriptor admission = %v, files=%d, bytes=%d, complete=%v", err, inspection.files, inspection.bytes, inspection.complete)
			}
		})
	}
}
