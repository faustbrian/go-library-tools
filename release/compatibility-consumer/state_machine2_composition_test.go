package compatibilityconsumer_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	sm "github.com/faustbrian/go-state-machine/v2"
	_ "github.com/faustbrian/go-state-machine/v2/diagram"
	_ "github.com/faustbrian/go-state-machine/v2/memory"
	_ "github.com/faustbrian/go-state-machine/v2/outbox"
	pg "github.com/faustbrian/go-state-machine/v2/postgres"
	"github.com/faustbrian/go-state-machine/v2/runner"
	_ "github.com/faustbrian/go-state-machine/v2/statemachinetest"
)

type stateMachineCollaborator struct{ handled, recorded int }

func (c *stateMachineCollaborator) Handle(context.Context, sm.Effect) error {
	c.handled++
	return nil
}

func (c *stateMachineCollaborator) Record(context.Context, runner.Record) error {
	c.recorded++
	return nil
}

func TestPublicStateMachineV2RunnerContract(t *testing.T) {
	c := &stateMachineCollaborator{}
	r, err := runner.New(c, runner.Options{Recorder: c, Limits: runner.Limits{
		MaxEffects: 2, MaxEffectPayloadBytes: 2, MaxTotalPayloadBytes: 3,
	}})
	if err != nil {
		t.Fatal(err)
	}
	for _, effects := range [][]sm.Effect{
		{{Kind: "a"}, {Kind: "b"}, {Kind: "c"}},
		{{Kind: "a", Payload: []byte("abc")}},
		{{Kind: "a", Payload: []byte("ab")}, {Kind: "b", Payload: []byte("cd")}},
	} {
		records, err := r.Execute(context.Background(), effects)
		if !errors.Is(err, sm.ErrLimitExceeded) || len(records) != 0 || c.handled != 0 || c.recorded != 0 {
			t.Fatalf("invalid plan reached stateMachineCollaborator: records=%d handled=%d recorded=%d err=%v", len(records), c.handled, c.recorded, err)
		}
	}
	effects := []sm.Effect{{Kind: "a", Payload: []byte("ab")}, {Kind: "b", Payload: []byte("c")}}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if records, err := r.Execute(ctx, effects); !errors.Is(err, context.Canceled) || len(records) != 0 || c.handled != 0 || c.recorded != 0 {
		t.Fatalf("canceled plan reached stateMachineCollaborator: %v", err)
	}
	records, err := r.Execute(context.Background(), effects)
	if err != nil || len(records) != 2 || c.handled != 2 || c.recorded != 2 || records[0].Outcome != runner.OutcomeSucceeded || records[1].Outcome != runner.OutcomeSucceeded {
		t.Fatalf("inclusive finite plan failed: records=%d handled=%d recorded=%d err=%v", len(records), c.handled, c.recorded, err)
	}
}

func TestPublicStateMachineV2CompileAndDiagnostics(t *testing.T) {
	definition := sm.Definition[string, string, struct{}]{
		Version: "one", Initial: "ready",
		States: []sm.StateDefinition[string]{{State: "ready"}, {State: "done", Terminal: true}},
		Transitions: []sm.TransitionDefinition[string, string, struct{}]{
			{ID: "finish", Sources: []string{"ready"}, Event: "finish", To: "done", Effects: []sm.Effect{{Kind: "notify", Payload: []byte("ab")}}},
		},
	}
	limits := sm.DefaultLimits()
	limits.MaxCompiledEffectPayloadBytes = 1
	if machine, err := sm.CompileWithLimits(definition, limits); machine != nil || err == nil {
		t.Fatalf("aggregate compile allowance was not enforced: %v", err)
	}
	limits.MaxCompiledEffectPayloadBytes = 2
	machine, err := sm.CompileWithLimits(definition, limits)
	if err != nil {
		t.Fatal(err)
	}
	result, err := machine.Transition(context.Background(), "ready", "finish", struct{}{}, sm.Metadata{})
	if err != nil || result.Next != "done" || len(result.Effects) != 1 || string(result.Effects[0].Payload) != "ab" {
		t.Fatalf("finite transition failed: result=%+v err=%v", result, err)
	}
	const marker = "application-value"
	guardErr := &sm.GuardRejectedError{TransitionID: sm.TransitionID(marker), Rejection: sm.Rejection{Code: marker, Message: marker}}
	if strings.Contains(guardErr.Error(), marker) || !errors.Is(guardErr, sm.ErrGuardRejected) {
		t.Fatal("default guard diagnostic changed")
	}
	if store, err := pg.New[string, string](pg.Options[string, string]{}); store != nil || !errors.Is(err, pg.ErrInvalidOptions) {
		t.Fatalf("missing caller-owned database dependencies accepted: %v", err)
	}
}
