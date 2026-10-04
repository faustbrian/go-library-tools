package gates

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/faustbrian/go-library-tools/v2/internal/inventory"
)

func TestNativeDocumentationReceivesRunnerCancellation(t *testing.T) {
	for _, operation := range []string{"docs", "local"} {
		t.Run(operation, func(t *testing.T) {
			root := localCoverageFixture(t, "# Ordinary\n")
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			laterChecks := 0
			runner := Runner{
				Root: root,
				Catalog: inventory.Inventory{Modules: []inventory.Module{{
					Directory: ".", Gates: map[string]bool{"documentation": true},
				}}},
				Executor: executorFunction(func(_ context.Context, command Command) error {
					// Ordinary local work completes, then the caller cancels
					// before the native documentation phase begins.
					if command.Name == "go" && slices.Equal(command.Args, []string{"mod", "tidy", "-diff"}) {
						cancel()
					}
					return nil
				}),
				DocumentationSpelling: func(context.Context, string) error { laterChecks++; return nil },
				DocumentationLinks:    func(context.Context, string) error { laterChecks++; return nil },
			}
			var err error
			switch operation {
			case "docs":
				cancel()
				err = runner.Docs(ctx, []string{"."})
			case "local":
				err = runner.Local(ctx, []string{"."})
			}
			if !errors.Is(err, context.Canceled) || laterChecks != 0 {
				t.Fatal("Runner lost native documentation cancellation or dispatched later checks")
			}
		})
	}
}
