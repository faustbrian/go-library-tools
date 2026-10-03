package compatibilityconsumer_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	moneyobjective "github.com/faustbrian/go-knapsack/objective/money/v3"
	"github.com/faustbrian/go-knapsack/v2"
	log "github.com/faustbrian/go-log/v2"
	"github.com/faustbrian/go-money/v2"
	"github.com/faustbrian/go-money/v2/moneytest"
	openrpc "github.com/faustbrian/go-openrpc/v2"
	"github.com/faustbrian/go-openrpc/v2/builder"
	"github.com/faustbrian/go-openrpc/v2/validate"
)

func TestOpenRPCV2BuilderValidationAndCanonicalComposition(t *testing.T) {
	version, err := openrpc.ParseVersion("1.4.1")
	if err != nil {
		t.Fatal(err)
	}
	info, err := openrpc.NewInfo(openrpc.InfoInput{Title: "Calculator", Version: "1.0.0"})
	if err != nil {
		t.Fatal(err)
	}
	method, err := openrpc.NewMethod(openrpc.MethodInput{Name: "add", Params: []openrpc.ContentDescriptorOrReference{}})
	if err != nil {
		t.Fatal(err)
	}
	documentBuilder, err := builder.NewDocument(version, info)
	if err != nil {
		t.Fatal(err)
	}
	documentBuilder, err = documentBuilder.WithMethod(method)
	if err != nil {
		t.Fatal(err)
	}
	document, err := documentBuilder.Build()
	if err != nil {
		t.Fatal(err)
	}
	report := validate.Document(context.Background(), document, validate.Options{MaxMethods: 1, MaxDiagnostics: 1})
	if !report.Valid() || report.Truncated() || len(report.Diagnostics()) != 0 {
		t.Fatalf("ordinary document validation = %#v", report)
	}
	encoded, err := openrpc.MarshalCanonical(document)
	if err != nil {
		t.Fatal(err)
	}
	const expected = `{"info":{"title":"Calculator","version":"1.0.0"},"methods":[{"name":"add","params":[]}],"openrpc":"1.4.1"}`
	if string(encoded) != expected {
		t.Fatalf("canonical output = %s, want %s", encoded, expected)
	}
}

func TestLogV2DefaultAndTrustedComposition(t *testing.T) {
	for _, test := range []struct {
		name    string
		trusted bool
		message string
	}{
		{name: "default", message: "[REDACTED]"},
		{name: "trusted", trusted: true, message: "application.ready"},
	} {
		t.Run(test.name, func(t *testing.T) {
			var output bytes.Buffer
			construct := log.New
			if test.trusted {
				construct = log.TrustedNew
			}
			logger, err := construct(slog.NewJSONHandler(&output, nil))
			if err != nil {
				t.Fatal(err)
			}
			logger.Info("application.ready", slog.String("component", "worker"))
			var record map[string]any
			if err := json.Unmarshal(output.Bytes(), &record); err != nil {
				t.Fatal(err)
			}
			if record["msg"] != test.message || record["level"] != "INFO" {
				t.Fatalf("record message/level = %v/%v, want %s/INFO", record["msg"], record["level"], test.message)
			}
			component, present := record["component"]
			if present != test.trusted || (present && component != "worker") {
				t.Fatalf("component = %v (present=%t), trusted=%t", component, present, test.trusted)
			}
		})
	}
}

func TestMoneyV2CanonicalObjectiveV3Composition(t *testing.T) {
	fixture := moneytest.CurrencyFixtures()[0]
	euro, moneyContext := fixture.Code, fixture.Context
	unitCost, err := money.Parse("0.60", euro, moneyContext)
	if err != nil {
		t.Fatal(err)
	}
	input := map[string]money.Money{"box": unitCost}
	costs, err := moneyobjective.New(input)
	if err != nil {
		t.Fatal(err)
	}
	clear(input)
	plan, err := knapsack.NewPlan(knapsack.PlanSpec{
		Containers: []knapsack.ContainerInstance{
			{ID: "box-1", TypeID: "box"},
			{ID: "box-2", TypeID: "box"},
		},
		Status:      knapsack.StatusFeasible,
		Termination: knapsack.TerminationCompleted,
	})
	if err != nil {
		t.Fatal(err)
	}
	var total money.Money
	total, err = costs.Total(plan)
	if err != nil {
		t.Fatal(err)
	}
	if total.String() != "1.20 EUR" || total.Currency() != euro || total.Context() != moneyContext {
		t.Fatalf("Total after caller map mutation = %s, currency=%v, context=%v", total.String(), total.Currency(), total.Context())
	}
}
