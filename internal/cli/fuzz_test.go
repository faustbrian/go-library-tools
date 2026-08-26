package cli

import (
	"reflect"
	"strings"
	"testing"

	"github.com/faustbrian/go-library-tools/internal/inventory"
)

func FuzzModuleSelection(f *testing.F) {
	f.Add("")
	f.Add("--module .")
	f.Add("--all unexpected")
	f.Fuzz(func(t *testing.T, input string) {
		if len(input) > 4096 {
			return
		}
		args := strings.Fields(input)
		if len(args) > 32 {
			return
		}
		modules := []inventory.Module{{Directory: "."}, {Directory: "integration/reference"}}
		first, firstErr := moduleSelection(args, modules)
		second, secondErr := moduleSelection(args, modules)
		if !reflect.DeepEqual(first, second) || errorText(firstErr) != errorText(secondErr) {
			t.Fatal("module selection is not deterministic")
		}
	})
}

func errorText(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
