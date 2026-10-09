package gates

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"
)

// This assertion must precede larger scanner fixtures: an EOF-loop regression
// still consults its context and fails here without hanging a native campaign.
func TestGosecDiscoveryCompletesAtEOF(t *testing.T) {
	root := t.TempDir()
	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()
	runner := Runner{Executor: executorFunction(func(_ context.Context, command Command) error {
		return json.NewEncoder(command.Stdout).Encode(map[string]string{"Dir": root})
	})}
	selected, err := runner.gosecPackages(ctx, root)
	if err != nil || !slices.Equal(selected, []string{"./"}) {
		t.Fatalf("finite discovery did not complete successfully: selected=%q error=%v", selected, err)
	}
}

// Native traversal variants can loop without consulting context. Keep the
// earliest observable scenarios in hosted, owned children before direct tests.
func TestWorkflowSecurityCompletionPreflightHosted(t *testing.T) {
	if os.Getenv("GITHUB_ACTIONS") != "true" {
		t.Skip("process-isolated traversal counterfactuals run only in hosted CI")
	}
	const image = "postgres:18@sha256:0123456789012345678901234567890123456789012345678901234567890123"
	const checkout = "actions/checkout@0123456789012345678901234567890123456789"
	for _, test := range []struct{ name, content, reason string }{
		{"include before extra axis", "jobs: {test: {strategy: {matrix: {include: [{image: '" + image + "'}], version: ['18']}}, container: '${{ matrix.image }}'}}\n", "immutable digest"},
		{"third job strategy", "jobs: {test: {runs-on: ubuntu-24.04, timeout-minutes: 5, strategy: {matrix: {include: [{image: '" + image + "'}]}}, container: '${{ matrix.image }}'}}\n", ""},
		{"third explicit checkout setting", "jobs: {test: {steps: [{uses: '" + checkout + "', with: {fetch-depth: 1, ordinary: value, persist-credentials: false}}]}}\n", ""},
		{"third merged checkout setting", "disabled: &disabled {persist-credentials: false}\njobs: {test: {steps: [{uses: '" + checkout + "', with: {fetch-depth: 1, ordinary: value, <<: *disabled}}]}}\n", ""},
		{"third forbidden event", "on: {push: {}, schedule: [], pull_request_target: {}}\njobs: {}\n", "pull_request_target"},
		{"trailing scalar retains nested permission check", "jobs: {test: {permissions: write-all}}\nname: ordinary\n", "write-all"},
		{"recursive event refusal", "on: &events [*events]\njobs: {}\n", "structure limit"},
		{"third local image refusal", "jobs: {test: {steps: [{uses: './.github/actions/check'}]}}\n", "immutable digest"},
	} {
		t.Run(test.name, func(t *testing.T) {
			if workflowPreflightOwnedChild(t) {
				return
			}
			root := t.TempDir()
			workflowRefusalWrite(t, root, ".github/workflows/ci.yml", test.content)
			if test.name == "third local image refusal" {
				workflowRefusalWrite(t, root, ".github/actions/check/action.yml", "runs: {using: docker, ordinary: value, image: 'busybox:latest'}\n")
			}
			calls := 0
			var output bytes.Buffer
			runner := Runner{Root: root, Output: &output, Executor: executorFunction(func(context.Context, Command) error {
				calls++
				return nil
			})}
			err := runner.Workflows(t.Context())
			if test.reason != "" {
				if err == nil || !strings.Contains(err.Error(), test.reason) || calls != 0 || output.Len() != 0 {
					t.Fatalf("security refusal = %v, effects=%d, output=%q; want %q before effects", err, calls, output.String(), test.reason)
				}
			} else if err != nil || calls != 1 || output.String() != "workflow contract passed\n" {
				t.Fatalf("finite admission = %v, effects=%d, output=%q", err, calls, output.String())
			}
		})
	}
}

// Run always waits, including after its deadline kills the exact owned child.
// Ordinary local tests keep their existing no-child path; risky counterfactual
// process execution remains hosted-only.
func workflowPreflightOwnedChild(t *testing.T) bool {
	t.Helper()
	const helper = "GOLIB_WORKFLOW_PREFLIGHT_CHILD"
	if os.Getenv("GITHUB_ACTIONS") != "true" || os.Getenv(helper) == t.Name() {
		return false
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, executable, "-test.run=^"+regexp.QuoteMeta(t.Name())+"$", "-test.count=1")
	// The parent owns temporary storage too: killing the child must not leave
	// its fixture directories outside the parent's mandatory cleanup.
	temporary := t.TempDir()
	command.Env = append(os.Environ(), helper+"="+t.Name(), "TMPDIR="+temporary, "TEMP="+temporary)
	var output bytes.Buffer
	command.Stdout, command.Stderr = &output, &output
	if err := command.Run(); err != nil {
		t.Fatalf("owned preflight failed or exceeded its deadline: %v; %s", err, output.String())
	}
	return true
}
