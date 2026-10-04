package mutation_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/faustbrian/go-library-tools/v2/internal/mutation"
)

// This fixed-byte boundary fixture is executed by hosted CI, not local checks.
func TestEquivalentInventoryInclusiveByteAdmission(t *testing.T) {
	const maximum = 1 << 20
	const document = `{"schema_version":1,"packages":[]}`
	data := document + strings.Repeat(" ", maximum-len(document))

	got, err := mutation.ParseEquivalentInventory(strings.NewReader(data))
	if err != nil || got.SchemaVersion != 1 || got.Packages == nil || len(got.Packages) != 0 {
		t.Fatal("inclusive byte allowance must admit the empty review inventory")
	}

	got, err = mutation.ParseEquivalentInventory(strings.NewReader(data + " "))
	if !errors.Is(err, mutation.ErrInvalid) || got.SchemaVersion != 0 || got.Packages != nil {
		t.Fatal("one byte over the allowance must return ErrInvalid and no inventory")
	}
	if err.Error() != "invalid mutation evidence: equivalent-mutant inventory exceeds 1048576 bytes" {
		t.Fatal("byte refusal must expose only the fixed admission category")
	}
}
