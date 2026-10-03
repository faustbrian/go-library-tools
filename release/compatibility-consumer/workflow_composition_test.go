package compatibilityconsumer_test

import (
	"bytes"
	"context"
	"testing"
	"time"

	workflow "github.com/faustbrian/go-workflow/v2"
	_ "github.com/faustbrian/go-workflow/v2/postgres"
)

func TestWorkflowV2PublishedDefinitionActivityComposition(t *testing.T) {
	steps := []workflow.StepSpec{{
		Name: "echo", Kind: workflow.StepActivity, Target: "echo",
		Timeout: time.Second, InputLimit: 8, ResultLimit: 8,
		Retry: workflow.RetryPolicy{MaxAttempts: 1, InitialDelay: time.Millisecond, MaxDelay: time.Millisecond},
	}}
	definition, err := workflow.NewDefinition(workflow.DefinitionSpec{
		Name: "consumer", Version: "1", Mode: workflow.Orchestration, Steps: steps,
	})
	if err != nil {
		t.Fatal(err)
	}
	steps[0].Name = "changed"
	copySteps := definition.Steps()
	if copySteps[0].Name != "echo" {
		t.Fatal("definition retained caller-owned steps")
	}
	copySteps[0].Name = "changed"
	if definition.Steps()[0].Name != "echo" {
		t.Fatal("definition exposed its owned steps")
	}
	registry, err := workflow.CompileDefinitions(definition)
	if err != nil {
		t.Fatal(err)
	}
	resolved, err := registry.Resolve("consumer", "1")
	if err != nil || resolved.Reference() != definition.Reference() {
		t.Fatalf("Resolve() lost the pinned definition: %v", err)
	}

	started := time.Now().UTC()
	input := []byte("ok")
	request, err := workflow.NewActivityRequest(workflow.ActivityRequestSpec{
		InstanceID: "instance", Definition: resolved.Reference(), StepName: "echo",
		Attempt: 1, MaxAttempts: 1, IdempotencyKey: "attempt",
		StartedAt: started, Deadline: started.Add(time.Minute),
		Input: input, InputLimit: 8, ResultLimit: 8,
	})
	if err != nil {
		t.Fatal(err)
	}
	input[0] = 'x'
	copyInput := request.Input()
	if !bytes.Equal(copyInput, []byte("ok")) {
		t.Fatal("request retained caller-owned input")
	}
	copyInput[0] = 'x'
	if !bytes.Equal(request.Input(), []byte("ok")) {
		t.Fatal("request exposed its owned input")
	}

	result := []byte("done")
	outcome, err := workflow.NewActivityOutcome(workflow.ActivityOutcomeSpec{
		Kind: workflow.ActivitySucceeded, Data: result,
	})
	if err != nil {
		t.Fatal(err)
	}
	result[0] = 'x'
	calls := 0
	activity, err := workflow.NewActivity("echo", func(ctx context.Context, got workflow.ActivityRequest) workflow.ActivityOutcome {
		calls++
		deadline, ok := ctx.Deadline()
		if ctx.Err() != nil || !ok || !deadline.Equal(request.Deadline()) ||
			got.Definition() != definition.Reference() || !bytes.Equal(got.Input(), []byte("ok")) {
			t.Fatal("activity lost the request identity, input, or deadline")
		}
		return outcome
	})
	if err != nil {
		t.Fatal(err)
	}
	activities, err := workflow.CompileActivities(activity)
	if err != nil {
		t.Fatal(err)
	}
	selected, err := activities.Resolve("echo")
	if err != nil || selected.Name() != "echo" {
		t.Fatalf("Resolve() lost the registered activity: %v", err)
	}
	got, err := selected.Execute(context.Background(), request)
	if err != nil || calls != 1 || got.Kind() != workflow.ActivitySucceeded ||
		got.Retryable() || !bytes.Equal(got.Data(), []byte("done")) {
		t.Fatalf("Execute() lost the explicit outcome: calls=%d, error=%v", calls, err)
	}
	copyResult := got.Data()
	copyResult[0] = 'x'
	if !bytes.Equal(got.Data(), []byte("done")) {
		t.Fatal("outcome exposed its owned result")
	}
}
