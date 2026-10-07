package gates

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Descriptor refusal must happen before the external validator can have effects.
func TestWorkflowRefusalBeforeExecutor(t *testing.T) {
	const pinned = "owner/action@0123456789012345678901234567890123456789"
	tests := []struct{ name, content, reason string }{
		{"malformed", "jobs: [\n", "decode workflow"},
		{"empty", "", "decode workflow"},
		{"multiple documents", "jobs: {}\n---\njobs: {}\n", "exactly one YAML document"},
		{"malformed trailing document", "jobs: {}\n---\n[\n", "decode workflow"},
		{"duplicate key", "jobs: {}\njobs: {}\n", "duplicate key"},
		{"unknown alias", "jobs: *missing\n", "decode workflow"},
		{"finite visit limit", "ordinary: [" + strings.Repeat("x,", 99_997) + "x]\n", "structure limit"},
		{"finite depth limit", "ordinary: " + strings.Repeat("[", 99) + "x" + strings.Repeat("]", 99) + "\n", "structure limit"},
		{"recursive mapping", "ordinary: &recursive {next: *recursive}\njobs: {}\n", "structure limit"},
		{"nested structure", "ordinary: " + strings.Repeat("[", 101) + "value" + strings.Repeat("]", 101) + "\njobs: {}\n", "structure limit"},
		{"non scalar action", "jobs: {test: {steps: [{uses: [ordinary]}]}}\n", "immutable SHA"},
		{"merged checkout persists", "settings: &settings {persist-credentials: true}\njobs: {test: {steps: [{uses: actions/checkout@0123456789012345678901234567890123456789, with: {<<: *settings}}]}}\n", "persist-credentials"},
		{"merged sequence checkout persists", "settings: &settings {persist-credentials: true}\njobs: {test: {steps: [{uses: actions/checkout@0123456789012345678901234567890123456789, with: {<<: [{ordinary: value}, *settings]}}]}}\n", "persist-credentials"},
		{"merged sequence checkout omission", "jobs: {test: {steps: [{uses: actions/checkout@0123456789012345678901234567890123456789, with: {<<: [{ordinary: value}]}}]}}\n", "persist-credentials"},
		{"merged step mutable action", "settings: &settings {uses: owner/action@main}\njobs: {test: {steps: [{<<: *settings}]}}\n", "immutable SHA"},
		{"sequence trigger", "on: [push, pull_request_target]\njobs: {}\n", "pull_request_target"},
		{"aliased sequence trigger", "events: &events [push, pull_request_target]\non: *events\njobs: {}\n", "pull_request_target"},
		{"aliased event element", "event: &event pull_request_target\non: [push, *event]\njobs: {}\n", "pull_request_target"},
		{"non scalar checkout settings", "jobs: {test: {steps: [{uses: actions/checkout@0123456789012345678901234567890123456789, with: ordinary}]}}\n", "persist-credentials"},
		{"non scalar image", "jobs: {test: {services: {db: {image: [ordinary]}}, steps: [{uses: " + pinned + "}]}}\n", "immutable digest"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			workflowRefusalWrite(t, root, ".github/workflows/ci.yml", test.content)
			var output bytes.Buffer
			calls := 0
			runner := Runner{Root: root, Output: &output, Executor: executorFunction(func(context.Context, Command) error { calls++; return nil })}
			err := runner.Workflows(t.Context())
			if err == nil || !strings.Contains(err.Error(), test.reason) {
				t.Fatalf("Workflows refusal = %v, want %q", err, test.reason)
			}
			if calls != 0 || output.Len() != 0 {
				t.Fatalf("refusal allowed effects: executor=%d output=%q", calls, output.String())
			}
		})
	}
}

func TestWorkflowStructureExactLimits(t *testing.T) {
	for _, test := range []struct {
		name, content string
	}{
		{"visits", "ordinary: [" + strings.Repeat("x,", 99_996) + "x]\n"},
		{"depth", "ordinary: " + strings.Repeat("[", 98) + "x" + strings.Repeat("]", 98) + "\n"},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			workflowRefusalWrite(t, root, ".github/workflows/ci.yml", test.content)
			calls := 0
			var output bytes.Buffer
			runner := Runner{Root: root, Output: &output, Executor: executorFunction(func(context.Context, Command) error {
				calls++
				return nil
			})}
			if err := runner.Workflows(t.Context()); err != nil {
				t.Fatalf("inclusive structure limit refused: %v", err)
			}
			if calls != 1 || output.String() != "workflow contract passed\n" {
				t.Fatalf("admission effects = %d/%q", calls, output.String())
			}
		})
	}
}

// Safe merge precedence and ordinary Unicode metadata still permit validation.
func TestWorkflowRefusalPreservesPinnedMergeControls(t *testing.T) {
	for _, settings := range []string{
		"{<<: *disabled}",
		"{<<: [{ordinary: value}, *disabled]}",
		"{<<: [{ordinary: value}], persist-credentials: false}",
		"{<<: {persist-credentials: true}, persist-credentials: false}",
	} {
		t.Run(settings, func(t *testing.T) {
			root := t.TempDir()
			workflowRefusalWrite(t, root, ".github/workflows/nested/ci.yaml", "name: Café\non: [push, pull_request]\ndisabled: &disabled {persist-credentials: false}\njobs: {test: {steps: [{uses: actions/checkout@0123456789012345678901234567890123456789, with: "+settings+"}]}}\n")
			workflowRefusalWrite(t, root, ".github/workflows/ordinary.txt", "not YAML: [")
			var output bytes.Buffer
			calls := 0
			runner := Runner{Root: root, Output: &output, Executor: executorFunction(func(_ context.Context, command Command) error {
				calls++
				if command.Name != "go" || command.Dir != root || command.Env["GOWORK"] != "off" || !strings.Contains(strings.Join(command.Args, " "), "actionlint@"+actionlintVersion) {
					t.Fatalf("unexpected validator command: %#v", command)
				}
				return nil
			})}
			if err := runner.Workflows(t.Context()); err != nil {
				t.Fatal(err)
			}
			if calls != 1 || output.String() != "workflow contract passed\n" {
				t.Fatalf("executor/output = %d/%q", calls, output.String())
			}
		})
	}
}

func TestWorkflowCheckoutAliasesPreserveAdmissionAndRefusal(t *testing.T) {
	for _, test := range []struct{ name, setting, reason string }{
		{"literal false alias", "false", ""},
		{"literal true alias", "true", "persist-credentials"},
		{"sequence alias", "[false]", "persist-credentials"},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			const path = ".github/workflows/ci.yml"
			content := "checkout: &checkout actions/checkout@0123456789012345678901234567890123456789\nsetting: &setting " + test.setting + "\nsteps: &steps [{uses: *checkout, with: {persist-credentials: *setting}}]\njobs: {test: {steps: *steps}}\n"
			workflowRefusalWrite(t, root, path, content)
			var output bytes.Buffer
			calls := 0
			runner := Runner{Root: root, Output: &output, Executor: executorFunction(func(context.Context, Command) error {
				calls++
				return nil
			})}
			err := runner.Workflows(t.Context())
			if test.reason != "" {
				if err == nil || !strings.Contains(err.Error(), test.reason) || calls != 0 || output.Len() != 0 {
					t.Fatalf("aliased checkout refusal: error=%v calls=%d output=%q", err, calls, output.String())
				}
			} else if err != nil || calls != 1 || output.String() != "workflow contract passed\n" {
				t.Fatalf("aliased checkout admission: error=%v calls=%d output=%q", err, calls, output.String())
			}
			data, err := os.ReadFile(filepath.Join(root, path))
			if err != nil || string(data) != content {
				t.Fatal("workflow inspection changed source bytes")
			}
		})
	}
}

func TestWorkflowLocalDescriptorRefusalBeforeExecutor(t *testing.T) {
	const marker = "application-private-detail"
	for _, test := range []struct{ name, descriptor, want string }{
		{"malformed", "runs: [\n# " + marker, "workflow security policy: invalid local action descriptor"},
		{"unsafe action", "name: " + marker + "\nruns: {using: composite, steps: [{uses: owner/action@main}]}", "workflow security policy: local action security policy failed"},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			workflowRefusalWrite(t, root, ".github/workflows/ci.yml", "jobs: {test: {steps: [{uses: ./ordinary}]}}")
			workflowRefusalWrite(t, root, "ordinary/action.yml", test.descriptor)
			var output bytes.Buffer
			calls := 0
			runner := Runner{Root: root, Output: &output, Executor: executorFunction(func(context.Context, Command) error { calls++; return nil })}
			err := runner.Workflows(t.Context())
			if err == nil || err.Error() != test.want || strings.Contains(err.Error(), marker) {
				t.Fatalf("local descriptor refusal = %v, want fixed category %q", err, test.want)
			}
			if calls != 0 || output.Len() != 0 {
				t.Fatalf("local descriptor refusal allowed effects: calls=%d output=%q", calls, output.String())
			}
		})
	}
}

func TestWorkflowLocalYAMLFallbackPreservesValidatorEffects(t *testing.T) {
	root := t.TempDir()
	workflowRefusalWrite(t, root, ".github/workflows/ci.yml", "jobs: {test: {steps: [{uses: ./ordinary}]}}")
	workflowRefusalWrite(t, root, "ordinary/action.yaml", "runs: {using: composite, steps: [{uses: owner/action@0123456789012345678901234567890123456789}]}")
	var output bytes.Buffer
	calls := 0
	runner := Runner{Root: root, Output: &output, Executor: executorFunction(func(_ context.Context, command Command) error {
		calls++
		if command.Name != "go" || !strings.Contains(strings.Join(command.Args, " "), "actionlint@"+actionlintVersion) {
			t.Fatalf("unexpected validator command: %#v", command)
		}
		return nil
	})}
	if err := runner.Workflows(t.Context()); err != nil {
		t.Fatal(err)
	}
	if calls != 1 || output.String() != "workflow contract passed\n" {
		t.Fatalf("fallback validator effects = %d/%q", calls, output.String())
	}
}

func TestLocalDescriptorInspectionPreservesCancellationBeforeDecode(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	inspection := localActionInspection{ctx: ctx}
	if err := inspection.inspect([]byte("runs: ["), 1, true); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled descriptor inspection = %v, want context cancellation", err)
	}
}

func TestWorkflowSymbolicDescriptorRefusedBeforeExecutor(t *testing.T) {
	root := t.TempDir()
	workflowRefusalWrite(t, root, "ordinary.yml", "jobs: {}\n")
	directory := filepath.Join(root, ".github", "workflows")
	if err := os.MkdirAll(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(root, "ordinary.yml"), filepath.Join(directory, "ci.yml")); err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	calls := 0
	runner := Runner{Root: root, Output: &output, Executor: executorFunction(func(context.Context, Command) error { calls++; return nil })}
	err := runner.Workflows(t.Context())
	const want = "workflow security policy: workflow security policy: symbolic link ci.yml"
	if err == nil || err.Error() != want || calls != 0 || output.Len() != 0 {
		t.Fatalf("symbolic descriptor = %v, executor calls=%d, output=%q; want %q without effects", err, calls, output.String(), want)
	}
}

func TestWorkflowMissingOrDirectoryDescriptorRefusedBeforeExecutor(t *testing.T) {
	for _, descriptorDirectory := range []bool{false, true} {
		name := "empty action directory"
		if descriptorDirectory {
			name = "descriptor is a directory"
		}
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			workflowRefusalWrite(t, root, ".github/workflows/ci.yml", "jobs: {test: {steps: [{uses: ./ordinary}]}}")
			directory := filepath.Join(root, "ordinary")
			if descriptorDirectory {
				directory = filepath.Join(directory, "action.yml")
			}
			if err := os.MkdirAll(directory, 0o700); err != nil {
				t.Fatal(err)
			}
			var output bytes.Buffer
			calls := 0
			runner := Runner{Root: root, Output: &output, Executor: executorFunction(func(context.Context, Command) error { calls++; return nil })}
			err := runner.Workflows(t.Context())
			const want = "workflow security policy: local action descriptor missing, symbolic, or oversized"
			if err == nil || err.Error() != want || calls != 0 || output.Len() != 0 {
				t.Fatalf("invalid descriptor = %v, executor calls=%d, output=%q; want %q without effects", err, calls, output.String(), want)
			}
		})
	}
}

func TestWorkflowRepeatedLocalDescriptorChargesOnce(t *testing.T) {
	root := t.TempDir()
	const workflow = "jobs: {test: {steps: [{uses: ./ordinary}, {uses: ./ordinary}]}}"
	const descriptor = "runs: {using: composite, steps: []}"
	workflowRefusalWrite(t, root, ".github/workflows/ci.yml", workflow)
	workflowRefusalWrite(t, root, "ordinary/action.yml", descriptor)
	inspection := localActionInspection{ctx: t.Context(), root: root, active: map[string]bool{}, complete: map[string]bool{}}
	if err := inspection.inspect([]byte(workflow), 0, false); err != nil {
		t.Fatal(err)
	}
	if inspection.files != 1 || inspection.bytes != len(descriptor) || len(inspection.active) != 0 || !inspection.complete["ordinary"] {
		t.Fatalf("repeated descriptor admission: files=%d, bytes=%d, active=%v, complete=%v", inspection.files, inspection.bytes, inspection.active, inspection.complete)
	}
	var output bytes.Buffer
	calls := 0
	runner := Runner{Root: root, Output: &output, Executor: executorFunction(func(context.Context, Command) error { calls++; return nil })}
	if err := runner.Workflows(t.Context()); err != nil {
		t.Fatal(err)
	}
	if calls != 1 || output.String() != "workflow contract passed\n" {
		t.Fatalf("repeated descriptor public control: executor calls=%d, output=%q", calls, output.String())
	}
}

// Shared parser-owned ASTs must preserve scalar aliases across both workflow
// selection and recursive local descriptors, without bypassing child policy.
func TestWorkflowLocalAliasesPreserveAdmissionAndRefusal(t *testing.T) {
	for _, mutable := range []bool{false, true} {
		name := "pinned child"
		if mutable {
			name = "mutable child"
		}
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			ref := "owner/action@0123456789012345678901234567890123456789"
			if mutable {
				ref = "owner/action@main"
			}
			contents := map[string]string{
				".github/workflows/ci.yml": "local: &local ./ordinary\njobs: {test: {steps: [{uses: *local}]}}\n",
				"ordinary/action.yml":      "child: &child ./nested\nruns: {using: composite, steps: [{uses: *child}]}\n",
				"nested/action.yml":        "runs: {using: composite, steps: [{uses: " + ref + "}]}\n",
			}
			for path, value := range contents {
				workflowRefusalWrite(t, root, path, value)
			}
			var output bytes.Buffer
			calls := 0
			runner := Runner{Root: root, Output: &output, Executor: executorFunction(func(context.Context, Command) error {
				calls++
				return nil
			})}
			err := runner.Workflows(t.Context())
			if mutable {
				if err == nil || err.Error() != "workflow security policy: local action security policy failed" || calls != 0 || output.Len() != 0 {
					t.Fatalf("aliased unsafe child: error=%v calls=%d output=%q", err, calls, output.String())
				}
			} else if err != nil || calls != 1 || output.String() != "workflow contract passed\n" {
				t.Fatalf("aliased pinned child: error=%v calls=%d output=%q", err, calls, output.String())
			}
			for path, want := range contents {
				data, err := os.ReadFile(filepath.Join(root, path))
				if err != nil || string(data) != want {
					t.Fatal("workflow inspection changed descriptor bytes")
				}
			}
		})
	}
}

func workflowRefusalWrite(t *testing.T, root, name, content string) {
	t.Helper()
	path := filepath.Join(root, name)
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
}
