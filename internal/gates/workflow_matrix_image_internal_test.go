package gates

import (
	"context"
	"strings"
	"testing"
)

func TestWorkflowStaticMatrixImages(t *testing.T) {
	const image = "postgres:18@sha256:0123456789012345678901234567890123456789012345678901234567890123"
	for _, test := range []struct {
		name, strategy, use string
		admit               bool
	}{
		{"include service", "matrix: {include: [{version: '14', image: '" + image + "'}, {version: '18', image: '" + image + "'}]}", "services: {db: {image: '${{ matrix.image }}'}}", true},
		{"include mapping container", "matrix: {include: [{image: '" + image + "'}]}", "container: {image: '${{ matrix.image }}'}", true},
		{"include scalar container", "matrix: {include: [{image: '" + image + "'}]}", "container: '${{ matrix.image }}'", true},
		{"mutable included image", "matrix: {include: [{image: '" + image + "'}, {image: 'postgres:latest'}]}", "services: {db: {image: '${{ matrix.image }}'}}", false},
		{"missing included image", "matrix: {include: [{image: '" + image + "'}, {version: '18'}]}", "services: {db: {image: '${{ matrix.image }}'}}", false},
		{"merged-only included image", "matrix: {include: [{'<<': {image: '" + image + "'}}]}", "services: {db: {image: '${{ matrix.image }}'}}", false},
		{"empty include", "matrix: {include: []}", "services: {db: {image: '${{ matrix.image }}'}}", false},
		{"dynamic matrix", "matrix: '${{ fromJSON(needs.prepare.outputs.matrix) }}'", "services: {db: {image: '${{ matrix.image }}'}}", false},
		{"partial include over axes", "matrix: {version: ['14', '18'], include: [{version: '14', image: '" + image + "'}]}", "services: {db: {image: '${{ matrix.image }}'}}", false},
		{"nested expression", "matrix: {include: [{image: '${{ vars.IMAGE }}'}]}", "services: {db: {image: '${{ matrix.image }}'}}", false},
		{"expression with digest suffix", "matrix: {include: [{image: '${{vars.IMAGE}}@sha256:0123456789012345678901234567890123456789012345678901234567890123'}]}", "services: {db: {image: '${{ matrix.image }}'}}", false},
		{"unknown field", "matrix: {include: [{image: '" + image + "'}]}", "services: {db: {image: '${{ matrix.other }}'}}", false},
		{"expression fallback", "matrix: {include: [{image: '" + image + "'}]}", "services: {db: {image: '${{ matrix.image || vars.IMAGE }}'}}", false},
		{"non-string image", "matrix: {include: [{image: 18}]}", "services: {db: {image: '${{ matrix.image }}'}}", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			workflowRefusalWrite(t, root, ".github/workflows/ci.yml", "jobs: {test: {strategy: {"+test.strategy+"}, "+test.use+"}}\n")
			calls := 0
			runner := Runner{Root: root, Executor: executorFunction(func(context.Context, Command) error { calls++; return nil })}
			err := runner.Workflows(t.Context())
			if test.admit {
				if err != nil || calls != 1 {
					t.Fatalf("pinned matrix admission = %v, executor calls=%d", err, calls)
				}
			} else if err == nil || !strings.Contains(err.Error(), "immutable digest") || calls != 0 {
				t.Fatalf("matrix refusal = %v, executor calls=%d", err, calls)
			}
		})
	}
}

func TestWorkflowMatrixImageEntryLimit(t *testing.T) {
	const image = "postgres:18@sha256:0123456789012345678901234567890123456789012345678901234567890123"
	for _, count := range []int{256, 257} {
		root := t.TempDir()
		rows := strings.Repeat("{image: '"+image+"'},", count-1) + "{image: '" + image + "'}"
		workflowRefusalWrite(t, root, ".github/workflows/ci.yml", "jobs: {test: {strategy: {matrix: {include: ["+rows+"]}}, services: {db: {image: '${{ matrix.image }}'}}}}\n")
		calls := 0
		runner := Runner{Root: root, Executor: executorFunction(func(context.Context, Command) error { calls++; return nil })}
		err := runner.Workflows(t.Context())
		if count == 256 {
			if err != nil || calls != 1 {
				t.Fatalf("inclusive entry limit = %v, executor calls=%d", err, calls)
			}
		} else if err == nil || !strings.Contains(err.Error(), "immutable digest") || calls != 0 {
			t.Fatalf("excess entry refusal = %v, executor calls=%d", err, calls)
		}
	}
}

func TestWorkflowMatrixImagesStayInOwningJob(t *testing.T) {
	const image = "postgres:18@sha256:0123456789012345678901234567890123456789012345678901234567890123"
	root := t.TempDir()
	workflowRefusalWrite(t, root, ".github/workflows/ci.yml", "jobs: {producer: {strategy: {matrix: {include: [{image: '"+image+"'}]}}}, consumer: {services: {db: {image: '${{ matrix.image }}'}}}}\n")
	calls := 0
	runner := Runner{Root: root, Executor: executorFunction(func(context.Context, Command) error { calls++; return nil })}
	if err := runner.Workflows(t.Context()); err == nil || !strings.Contains(err.Error(), "immutable digest") || calls != 0 {
		t.Fatalf("foreign job domain refusal = %v, executor calls=%d", err, calls)
	}
}

func TestWorkflowMatrixImageLookupDoesNotExpandRecursiveMerges(t *testing.T) {
	root := t.TempDir()
	workflowRefusalWrite(t, root, ".github/workflows/ci.yml", "jobs: {test: {strategy: {matrix: {include: [&row {'<<': [*row, *row]}]}}, services: {db: {image: '${{ matrix.image }}'}}}}\n")
	calls := 0
	runner := Runner{Root: root, Executor: executorFunction(func(context.Context, Command) error { calls++; return nil })}
	if err := runner.Workflows(t.Context()); err == nil || !strings.Contains(err.Error(), "workflow YAML structure limit exceeded") || calls != 0 {
		t.Fatalf("structural quota refusal = %v, executor calls=%d", err, calls)
	}
}
