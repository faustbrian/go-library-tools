package gates

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

// Policy selection observes parser-owned workflow paths, not matching words
// anywhere in metadata. The inert validator proves dispatch, not its schema.
func TestWorkflowPolicyExecutionScope(t *testing.T) {
	const pinned = "owner/action@0123456789012345678901234567890123456789"
	const image = "ordinary@sha256:0123456789012345678901234567890123456789012345678901234567890123"
	for _, test := range []struct{ name, content, reason string }{
		{"root read permissions", "permissions: {contents: read}\njobs: {}", ""},
		{"job read permissions", "jobs: {test: {permissions: {contents: read}}}", ""},
		{"job write permissions", "jobs: {test: {permissions: write-all}}", "permissions write-all"},
		{"nested job permission metadata", "jobs: {test: {env: {permissions: write-all}}}", ""},
		{"other two-part permission metadata", "ordinary: {test: {permissions: write-all}}", ""},
		{"pinned reusable job", "jobs: {test: {uses: " + pinned + "}}", ""},
		{"mutable reusable job", "jobs: {test: {uses: owner/action@main}}", "immutable SHA"},
		{"other two-part action metadata", "ordinary: {test: {uses: owner/action@main}}", ""},
		{"nested job action metadata", "jobs: {test: {env: {uses: owner/action@main}}}", ""},
		{"pinned scalar container", "jobs: {test: {container: '" + image + "'}}", ""},
		{"mutable scalar container", "jobs: {test: {container: ordinary:latest}}", "job container images"},
		{"pinned mapping container", "jobs: {test: {container: {image: '" + image + "'}}}", ""},
		{"other two-part container metadata", "ordinary: {test: {container: ordinary:latest}}", ""},
		{"nested job container metadata", "jobs: {test: {env: {container: ordinary:latest}}}", ""},
		{"pinned service image", "jobs: {test: {services: {db: {image: '" + image + "'}}}}", ""},
		{"mutable mapping container", "jobs: {test: {container: {image: ordinary:latest}}}", "container images"},
		{"mutable service image", "jobs: {test: {services: {db: {image: ordinary:latest}}}}", "container images"},
		{"other three-part image metadata", "ordinary: {test: {container: {image: ordinary:latest}}}", ""},
		{"other four-part image metadata", "ordinary: {test: {services: {db: {image: ordinary:latest}}}}", ""},
		{"job environment image metadata", "jobs: {test: {env: {image: ordinary:latest}}}", ""},
		{"job other nested image metadata", "jobs: {test: {ordinary: {db: {image: ordinary:latest}}}}", ""},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			workflowRefusalWrite(t, root, ".github/workflows/ci.yml", test.content+"\n")
			var output bytes.Buffer
			calls := 0
			runner := Runner{Root: root, Output: &output, Executor: executorFunction(func(context.Context, Command) error { calls++; return nil })}
			err := runner.Workflows(t.Context())
			if test.reason != "" {
				if err == nil || !strings.Contains(err.Error(), test.reason) || calls != 0 || output.Len() != 0 {
					t.Fatalf("execution policy refusal = %v, calls=%d, output=%q; want %q without effects", err, calls, output.String(), test.reason)
				}
				return
			}
			if err != nil || calls != 1 || output.String() != "workflow contract passed\n" {
				t.Fatalf("policy admission = %v, calls=%d, output=%q", err, calls, output.String())
			}
		})
	}
}
