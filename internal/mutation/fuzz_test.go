package mutation_test

import (
	"bytes"
	"testing"

	"github.com/faustbrian/go-library-tools/internal/mutation"
)

func FuzzMutationRecords(f *testing.F) {
	f.Add([]byte(`{"schema_version":1,"packages":[]}`))
	f.Add([]byte("go_version='go1.27.0'\ngoos='linux'\ngoarch='amd64'\n"))
	f.Fuzz(func(t *testing.T, input []byte) {
		if len(input) > 2<<20 {
			return
		}
		_, _ = mutation.ValidateReport(bytes.NewReader(input))
		_, _ = mutation.ParseMigrationLedger(bytes.NewReader(input))
		_, _ = mutation.ParseZeroInventory(bytes.NewReader(input))
		_, _ = mutation.ParseRuntimeIdentity(bytes.NewReader(input))
	})
}
