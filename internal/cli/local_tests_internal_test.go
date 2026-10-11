package cli

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

	"github.com/faustbrian/go-library-tools/v2/internal/gates"
)

func TestExecuteLocalRiskTestsThroughInventory(t *testing.T) {
	for _, scenario := range []struct {
		name, names string
		failure     error
		want        int
	}{
		{"matched", "TestContract\n", nil, 0},
		{"empty selector", "TestOther\n", nil, 1},
		{"contract failure", "TestContract\n", errors.New("contract failed"), 1},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			root := internalFixture(t)
			policy := "schema_version: 1\ntool_version: v2.0.0\noperations:\n  - module: .\n    gate: local-test\n    steps:\n      - type: go-test\n        packages: [.]\n        run: '^TestContract$'\n        race: true\n        timeout: 1m\n"
			if err := os.WriteFile(filepath.Join(root, ".golib.yaml"), []byte(policy), 0o600); err != nil {
				t.Fatal(err)
			}
			created, cleaned, discovered, executed := 0, 0, 0, 0
			factory := func(string, io.Writer, io.Writer) (gates.Executor, func() error, error) {
				created++
				return changedLocalCLIExecutor(t, func(ctx context.Context, command gates.Command) error {
					if command.Name != "go" || !slices.Contains(command.Args, "-run=^TestContract$") {
						return nil
					}
					if !slices.Contains(command.Args, "-race") || !slices.Contains(command.Args, "-count=1") {
						t.Fatal("focused race contract not retained")
					}
					if _, ok := ctx.Deadline(); !ok {
						t.Fatal("risk check has no deadline")
					}
					if slices.Contains(command.Args, "-list=^TestContract$") {
						discovered++
						_, err := io.WriteString(command.Stdout, scenario.names)
						return err
					}
					executed++
					return scenario.failure
				}), func() error { cleaned++; return nil }, nil
			}
			var stdout, stderr bytes.Buffer
			code := execute([]string{"check", "--local"}, root, &stdout, &stderr, factory)
			if code != scenario.want || created != 1 || cleaned != 1 || discovered != 1 {
				t.Fatalf("code=%d sessions=%d/%d discovery=%d: %s", code, created, cleaned, discovered, stderr.String())
			}
			if scenario.name == "empty selector" {
				if executed != 0 || !strings.Contains(stderr.String(), "no matching tests") {
					t.Fatalf("empty selector executed=%d: %s", executed, stderr.String())
				}
			} else if executed != 1 {
				t.Fatalf("focused executions=%d", executed)
			}
		})
	}
}
