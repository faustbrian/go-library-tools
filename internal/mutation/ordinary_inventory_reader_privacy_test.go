package mutation_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/faustbrian/go-library-tools/v2/internal/mutation"
)

type ordinaryInventoryReaderFailure struct{}

func (ordinaryInventoryReaderFailure) Read([]byte) (int, error) {
	return 0, errors.New("ordinary reader detail")
}

func TestOrdinaryInventoryReaderFailuresAreCategorical(t *testing.T) {
	t.Run("zero", func(t *testing.T) {
		got, err := mutation.ParseZeroInventory(ordinaryInventoryReaderFailure{})
		if !errors.Is(err, mutation.ErrInvalid) || got.SchemaVersion != 0 || got.Packages != nil {
			t.Fatalf("zero inventory failure = %#v, %v; want zero and ErrInvalid", got, err)
		}
		if err.Error() != "invalid mutation evidence: read zero-mutant inventory" {
			t.Fatalf("default zero inventory error = %q; want category without reader detail", err.Error())
		}
	})
	t.Run("equivalent", func(t *testing.T) {
		got, err := mutation.ParseEquivalentInventory(ordinaryInventoryReaderFailure{})
		if !errors.Is(err, mutation.ErrInvalid) || got.SchemaVersion != 0 || got.Packages != nil {
			t.Fatalf("equivalent inventory failure = %#v, %v; want zero and ErrInvalid", got, err)
		}
		if err.Error() != "invalid mutation evidence: read equivalent-mutant inventory" {
			t.Fatalf("default equivalent inventory error = %q; want category without reader detail", err.Error())
		}
	})
}

func TestOrdinaryInventoryReadersAcceptEmptyReviewSets(t *testing.T) {
	const data = `{"schema_version":1,"packages":[]}`
	zero, err := mutation.ParseZeroInventory(strings.NewReader(data))
	if err != nil || zero.SchemaVersion != 1 || zero.Packages == nil || len(zero.Packages) != 0 {
		t.Fatalf("accepted zero inventory = %#v, %v", zero, err)
	}
	equivalent, err := mutation.ParseEquivalentInventory(strings.NewReader(data))
	if err != nil || equivalent.SchemaVersion != 1 || equivalent.Packages == nil || len(equivalent.Packages) != 0 {
		t.Fatalf("accepted equivalent inventory = %#v, %v", equivalent, err)
	}
}
