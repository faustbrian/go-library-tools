package securitydocs

import (
	"encoding/json"
	"io"
	"strings"
	"testing"
)

// Keep this finite consumption assertion before excessive-depth fixtures.
// The native verifier stops on the first failure: a premature array return
// must not continue into fixtures that an ignored child error could loop on.
func TestJSONWalkConsumesEveryArrayMemberBeforeReturning(t *testing.T) {
	decoder := json.NewDecoder(strings.NewReader(`[{"ordinary":1},2]`))
	if err := walkJSONValue(decoder, 0); err != nil {
		t.Fatal(err)
	}
	if _, err := decoder.Token(); err != io.EOF {
		t.Fatalf("successful walk left array content unread: %v", err)
	}
}

func TestJSONWalkBoundsObjectNesting(t *testing.T) {
	for _, test := range []struct {
		depth int
		valid bool
	}{{128, true}, {129, false}} {
		value := strings.Repeat(`{"ordinary":`, test.depth) + "0" + strings.Repeat("}", test.depth)
		decoder := json.NewDecoder(strings.NewReader(value))
		if err := walkJSONValue(decoder, 0); (err == nil) != test.valid {
			t.Fatalf("object depth %d: error=%v, want valid=%t", test.depth, err, test.valid)
		}
	}
}
