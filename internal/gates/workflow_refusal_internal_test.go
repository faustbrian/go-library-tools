package gates

import (
	"bytes"
	"context"
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
