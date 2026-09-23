package gates

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestWorkflowsRunsPinnedActionlintWithBoundedDiagnostics(t *testing.T) {
	var command Command
	var output bytes.Buffer
	runner := Runner{Root: "/repo", Output: &output, Executor: executorFunction(func(_ context.Context, value Command) error {
		command = value
		return nil
	})}
	if err := runner.Workflows(context.Background()); err != nil {
		t.Fatalf("Workflows() error = %v", err)
	}
	if command.Name != "go" || command.Dir != "/repo" || command.Env["GOWORK"] != "off" {
		t.Fatalf("workflow command = %#v", command)
	}
	want := "run github.com/rhysd/actionlint/cmd/actionlint@v1.7.12 -no-color -oneline -shellcheck= -pyflakes="
	if strings.Join(command.Args, " ") != want || output.String() != "workflow contract passed\n" {
		t.Fatalf("workflow command/output = %q, %q", strings.Join(command.Args, " "), output.String())
	}
}

func TestWorkflowsRejectsSecurityPolicyViolationsThatActionlintAccepts(t *testing.T) {
	root := t.TempDir()
	workflowDirectory := filepath.Join(root, ".github", "workflows")
	if err := os.MkdirAll(workflowDirectory, 0o700); err != nil {
		t.Fatal(err)
	}
	workflow := "name: unsafe\non: pull_request_target\npermissions: write-all\njobs:\n  test:\n    runs-on: ubuntu-latest\n    steps:\n      - uses: actions/checkout@main\n"
	if err := os.WriteFile(filepath.Join(workflowDirectory, "unsafe.yml"), []byte(workflow), 0o600); err != nil {
		t.Fatal(err)
	}
	runner := Runner{Root: root, Executor: executorFunction(func(context.Context, Command) error { return nil })}
	err := runner.Workflows(context.Background())
	if err == nil || !strings.Contains(err.Error(), "workflow security policy") || !strings.Contains(err.Error(), "pull_request_target") {
		t.Fatalf("Workflows() error = %v", err)
	}
}

func TestWorkflowsRejectsQuotedAndFlowStyleSecurityViolations(t *testing.T) {
	tests := map[string]string{
		"quoted trigger":      "name: unsafe\n\"on\": {pull_request_target: {}}\njobs: {test: {runs-on: ubuntu-latest, steps: []}}\n",
		"quoted action":       "name: unsafe\non: push\njobs: {test: {runs-on: ubuntu-latest, steps: [{\"uses\": \"actions/checkout@main\"}]}}\n",
		"docker tag":          "name: unsafe\non: push\njobs: {test: {runs-on: ubuntu-latest, container: {image: alpine:latest}, steps: [{uses: \"docker://alpine:latest\"}]}}\n",
		"job container":       "name: unsafe\non: push\njobs: {test: {runs-on: ubuntu-latest, container: alpine:latest, steps: []}}\n",
		"service image":       "name: unsafe\non: push\njobs: {test: {runs-on: ubuntu-latest, services: {db: {image: postgres:18}}, steps: []}}\n",
		"quoted checkout":     "name: unsafe\non: push\njobs: {test: {runs-on: ubuntu-latest, steps: [{uses: \"actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1\", with: {\"persist-credentials\": true}}]}}\n",
		"checkout omission":   "name: unsafe\non: push\njobs: {test: {runs-on: ubuntu-latest, steps: [{uses: \"actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1\"}]}}\n",
		"checkout expression": "name: unsafe\non: push\njobs: {test: {runs-on: ubuntu-latest, steps: [{uses: \"ACTIONS/CHECKOUT@3d3c42e5aac5ba805825da76410c181273ba90b1\", with: {\"persist-credentials\": \"${{ false }}\"}}]}}\n",
	}
	for name, workflow := range tests {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			workflowDirectory := filepath.Join(root, ".github", "workflows")
			if err := os.MkdirAll(workflowDirectory, 0o700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(workflowDirectory, "unsafe.yml"), []byte(workflow), 0o600); err != nil {
				t.Fatal(err)
			}
			runner := Runner{Root: root, Executor: executorFunction(func(context.Context, Command) error { return nil })}
			if err := runner.Workflows(context.Background()); err == nil || !strings.Contains(err.Error(), "workflow security policy") {
				t.Fatalf("Workflows() error = %v", err)
			}
		})
	}
}

func TestInspectWorkflowAcceptsCheckoutWithLiteralCredentialDisable(t *testing.T) {
	safe := []byte("name: safe\non: push\njobs: {test: {runs-on: ubuntu-latest, steps: [{uses: \"actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1\", with: {\"persist-credentials\": false}}]}}\n")
	findings, err := inspectWorkflow(safe)
	if err != nil || len(findings) != 0 {
		t.Fatalf("inspectWorkflow(safe checkout) = %#v, %v", findings, err)
	}
}

func TestInspectWorkflowFollowsAliasesWithoutTreatingEnvironmentKeysAsActions(t *testing.T) {
	safe := []byte("name: safe\non: push\nenv: {uses: ordinary-value, persist-credentials: true}\njobs: {test: {runs-on: ubuntu-latest, steps: []}}\n")
	findings, err := inspectWorkflow(safe)
	if err != nil || len(findings) != 0 {
		t.Fatalf("inspectWorkflow(safe) = %#v, %v", findings, err)
	}
	unsafe := []byte("name: unsafe\non: push\nunsafe: &unsafe {uses: actions/checkout@main}\njobs: {test: {runs-on: ubuntu-latest, steps: [*unsafe]}}\n")
	findings, err = inspectWorkflow(unsafe)
	if err != nil || !slices.Contains(findings, "remote actions require immutable SHA references") {
		t.Fatalf("inspectWorkflow(alias) = %#v, %v", findings, err)
	}
}

func TestInspectWorkflowDereferencesSecuritySensitiveScalarAliases(t *testing.T) {
	workflow := []byte("name: unsafe\non: push\nwrite: &write write-all\nmutable: &mutable alpine:latest\npermissions: *write\njobs: {test: {runs-on: ubuntu-latest, container: *mutable, steps: []}}\n")
	findings, err := inspectWorkflow(workflow)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"permissions write-all is forbidden", "job container images require immutable digest references"} {
		if !slices.Contains(findings, want) {
			t.Fatalf("inspectWorkflow(alias) = %#v, missing %q", findings, want)
		}
	}
}

func FuzzInspectWorkflowNeverPanics(f *testing.F) {
	f.Add([]byte("name: safe\non: push\njobs: {}\n"))
	f.Add([]byte("name: unsafe\n'on': {pull_request_target: {}}\njobs: {}\n"))
	f.Fuzz(func(_ *testing.T, value []byte) {
		_, _ = inspectWorkflow(value)
	})
}

func TestWorkflowsReportsBoundedActionlintFailure(t *testing.T) {
	failure := errors.New("failed")
	runner := Runner{Root: "/repo", Executor: executorFunction(func(_ context.Context, command Command) error {
		_, _ = io.WriteString(command.Stderr, "invalid workflow")
		return failure
	})}
	err := runner.Workflows(context.Background())
	if err == nil || !errors.Is(err, failure) || !strings.Contains(err.Error(), "invalid workflow") {
		t.Fatalf("Workflows() error = %v", err)
	}
	runner.Executor = executorFunction(func(context.Context, Command) error { return failure })
	err = runner.Workflows(context.Background())
	if err == nil || !errors.Is(err, failure) {
		t.Fatalf("Workflows() plain tool error = %v", err)
	}
	if strings.Contains(err.Error(), "invalid workflow") {
		t.Fatalf("Workflows() plain tool diagnostics = %v", err)
	}

	runner.Executor = executorFunction(func(_ context.Context, command Command) error {
		_, _ = command.Stdout.Write(make([]byte, maximumWorkflowOutput+1))
		return nil
	})
	if err := runner.Workflows(context.Background()); err == nil || !strings.Contains(err.Error(), "exceeded") {
		t.Fatalf("Workflows() overflow error = %v", err)
	}
}

func TestBoundedWorkflowBufferEnforcesCumulativeLimit(t *testing.T) {
	var buffer boundedWorkflowBuffer
	first := make([]byte, maximumWorkflowOutput-1)
	if written, err := buffer.Write(first); err != nil || written != len(first) || buffer.overflow {
		t.Fatalf("first Write() = %d, %v, overflow %v", written, err, buffer.overflow)
	}
	if written, err := buffer.Write([]byte{'a'}); err != nil || written != 1 || buffer.overflow || buffer.data.Len() != maximumWorkflowOutput {
		t.Fatalf("boundary Write() = %d, %v, overflow %v, length %d", written, err, buffer.overflow, buffer.data.Len())
	}
	if written, err := buffer.Write([]byte("bc")); err != nil || written != 2 || !buffer.overflow || buffer.data.Len() != maximumWorkflowOutput {
		t.Fatalf("overflow Write() = %d, %v, overflow %v, length %d", written, err, buffer.overflow, buffer.data.Len())
	}
}
