package compatibilityconsumer_test

import (
	"testing"

	moneyobjective "github.com/faustbrian/go-knapsack/objective/money/v3"
	"github.com/faustbrian/go-knapsack/v2"
	"github.com/faustbrian/go-money/v2"
	"github.com/faustbrian/go-money/v2/moneytest"
)

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
