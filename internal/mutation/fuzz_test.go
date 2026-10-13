package mutation_test

import (
	"bytes"
	"testing"

	"github.com/faustbrian/go-library-tools/v2/internal/mutation"
)

// FuzzMutationRecords caps input at 2 MiB and exercises report, migration,
// zero/equivalent-review and runtime parsers; parser success is not passing campaign
// evidence.
func FuzzMutationRecords(f *testing.F) {
	f.Add([]byte(`{"schema_version":1,"packages":[]}`))
	f.Add([]byte("go_version='go1.27.0'\ngoos='linux'\ngoarch='amd64'\n"))
	f.Add([]byte(`{"schema_version":1,"packages":[{"mutations":[{"goarchs":["amd64","arm64"]}]}]}`))
	f.Add([]byte(`{"schema_version":1,"packages":[{"mutations":[{"goarchs":null}]}]}`))
	f.Fuzz(func(_ *testing.T, input []byte) {
		if len(input) > 2<<20 {
			return
		}
		_, _ = mutation.ValidateReport(bytes.NewReader(input))
		_, _ = mutation.ParseMigrationLedger(bytes.NewReader(input))
		_, _ = mutation.ParseZeroInventory(bytes.NewReader(input))
		_, _ = mutation.ParseEquivalentInventory(bytes.NewReader(input))
		_, _ = mutation.ParseRuntimeIdentity(bytes.NewReader(input))
	})
}
