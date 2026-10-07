package gates

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"go.yaml.in/yaml/v3"
)

// Finite boundary checks precede recursive fixtures so a traversal regression
// can fail on a small admitted graph instead of exhausting a campaign timeout.
func TestWorkflowFiniteAliasAndActionPreflight(t *testing.T) {
	const checkout = "actions/checkout@0123456789012345678901234567890123456789"
	for _, test := range []struct{ name, content, reason string }{
		{"exact alias depth", "leaf: &leaf ordinary\nordinary: " + strings.Repeat("[", 97) + "*leaf" + strings.Repeat("]", 97) + "\njobs: {}\n", ""},
		{"excess alias depth", "leaf: &leaf ordinary\nordinary: " + strings.Repeat("[", 98) + "*leaf" + strings.Repeat("]", 98) + "\njobs: {}\n", "structure limit"},
		{"ordinary key before safe merge", "disabled: &disabled {persist-credentials: false}\njobs: {test: {steps: [{uses: " + checkout + ", with: {ordinary: value, <<: *disabled}}]}}\n", ""},
		{"empty action owner", "jobs: {test: {steps: [{uses: '@0123456789012345678901234567890123456789'}]}}\n", "immutable SHA"},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			workflowRefusalWrite(t, root, ".github/workflows/ci.yml", test.content)
			var output bytes.Buffer
			calls := 0
			runner := Runner{Root: root, Output: &output, Executor: executorFunction(func(context.Context, Command) error {
				calls++
				return nil
			})}
			err := runner.Workflows(t.Context())
			if test.reason != "" {
				if err == nil || !strings.Contains(err.Error(), test.reason) || calls != 0 || output.Len() != 0 {
					t.Fatalf("finite refusal = %v, calls=%d, output=%q; want %q without effects", err, calls, output.String(), test.reason)
				}
				return
			}
			if err != nil || calls != 1 || output.String() != "workflow contract passed\n" {
				t.Fatalf("finite admission = %v, calls=%d, output=%q", err, calls, output.String())
			}
		})
	}
}

// These are finite parser-owned merge graphs. The lookup is tested before
// whole-document inspection can mask its independent recursion allowance.
func TestWorkflowFiniteMergeLookupPreflight(t *testing.T) {
	for _, shape := range []struct{ name, prefix, suffix string }{
		{"mapping", "{<<: ", "}"},
		{"sequence", "{<<: [", "]}"},
	} {
		for _, merges := range []int{99, 100} {
			name := "exact depth"
			if merges == 100 {
				name = "excess depth"
			}
			t.Run(shape.name+"/"+name, func(t *testing.T) {
				content := "disabled: &disabled {persist-credentials: false}\nsettings: " +
					strings.Repeat(shape.prefix, merges) + "*disabled" + strings.Repeat(shape.suffix, merges) + "\n"
				var document yaml.Node
				if err := yaml.Unmarshal([]byte(content), &document); err != nil {
					t.Fatal(err)
				}
				settings := workflowMappingValue(document.Content[0], "settings", 0)
				if settings == nil {
					t.Fatal("decoded settings mapping is missing")
				}
				value := workflowMappingValue(settings, "persist-credentials", 0)
				if merges == 100 {
					if value != nil {
						t.Fatal("lookup crossed the recursion allowance")
					}
					return
				}
				if value == nil || value.Kind != yaml.ScalarNode || value.Value != "false" {
					t.Fatal("inclusive recursion allowance lost the merged setting")
				}
			})
		}
	}
}
