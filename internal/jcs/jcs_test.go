package jcs

import (
	"errors"
	"strconv"
	"strings"
	"testing"
)

func TestCanonicalizeOrdersObjectMembersAndPreservesArrayOrder(t *testing.T) {
	t.Parallel()

	got, err := Canonicalize([]byte(`{"z":1,"a":[3,2,1],"nested":{"b":true,"a":null}}`))
	if err != nil {
		t.Fatal(err)
	}
	want := `{"a":[3,2,1],"nested":{"a":null,"b":true},"z":1}`
	if string(got) != want {
		t.Fatalf("Canonicalize() = %s, want %s", got, want)
	}
}

func TestCanonicalizeRejectsDepthStringAndCollectionBounds(t *testing.T) {
	t.Parallel()
	tooDeep := strings.Repeat("[", maximumDepth+1) + "0" + strings.Repeat("]", maximumDepth+1)
	tooLong := `{"s":"` + strings.Repeat("x", maximumStringBytes+1) + `"}`
	tooMany := `[` + strings.Repeat("0,", maximumCollectionEntries) + `0]`
	for _, input := range []string{tooDeep, tooLong, tooMany} {
		if _, err := Canonicalize([]byte(input)); err == nil {
			t.Fatal("bounded input accepted")
		}
	}
}

func TestCanonicalizeExactDepthBoundary(t *testing.T) {
	t.Parallel()
	for depth := maximumDepth - 1; depth <= maximumDepth; depth++ {
		input := strings.Repeat("[", depth) + "0" + strings.Repeat("]", depth)
		if _, err := Canonicalize([]byte(input)); err != nil {
			t.Fatalf("depth %d rejected: %v", depth, err)
		}
	}
	input := strings.Repeat("[", maximumDepth+1) + "0" + strings.Repeat("]", maximumDepth+1)
	if _, err := Canonicalize([]byte(input)); err == nil {
		t.Fatal("above maximum depth accepted")
	}
}

func TestCanonicalizeHistoricalUsesReleasedDepthBoundary(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name     string
		depth    int
		accepted bool
	}{
		{name: "below", depth: maximumHistoricalDepth - 1, accepted: true},
		{name: "exact", depth: maximumHistoricalDepth, accepted: true},
		{name: "above", depth: maximumHistoricalDepth + 1, accepted: false},
	} {
		t.Run(test.name, func(t *testing.T) {
			input := []byte(strings.Repeat("[", test.depth) + "0" + strings.Repeat("]", test.depth))
			_, err := CanonicalizeHistorical(input)
			if test.accepted != (err == nil) {
				t.Fatalf("depth=%d accepted=%v error=%v", test.depth, test.accepted, err)
			}
		})
	}
}

func TestCanonicalizeHistoricalNormalizesRFC8785Numbers(t *testing.T) {
	t.Parallel()
	for input, want := range map[string]string{
		`-0`:                   `0`,
		`1.5`:                  `1.5`,
		`1e2`:                  `100`,
		`1e-7`:                 `1e-7`,
		`1e21`:                 `1e+21`,
		`0.000001`:             `0.000001`,
		`333333333.33333329`:   `333333333.3333333`,
		`4.50`:                 `4.5`,
		`2e-3`:                 `0.002`,
		`1.000000000000000005`: `1`,
	} {
		got, err := CanonicalizeHistorical([]byte(input))
		if err != nil {
			t.Fatalf("CanonicalizeHistorical(%s): %v", input, err)
		}
		if string(got) != want {
			t.Fatalf("CanonicalizeHistorical(%s) = %s, want %s", input, got, want)
		}
	}
}

func TestCanonicalizeHistoricalPreservesReleasedNoncharacterValues(t *testing.T) {
	t.Parallel()

	want := []byte{0x22, 0xef, 0xb7, 0x90, 0xef, 0xbf, 0xbf, 0x22}
	for _, test := range []struct {
		name  string
		input []byte
	}{
		{name: "escaped", input: []byte(`"\ufdd0\uffff"`)},
		{name: "raw", input: want},
	} {
		t.Run(test.name, func(t *testing.T) {
			got, err := CanonicalizeHistorical(test.input)
			if err != nil {
				t.Fatalf("CanonicalizeHistorical(noncharacters) error = %v", err)
			}
			if string(got) != string(want) {
				t.Fatalf("CanonicalizeHistorical(noncharacters) = %q, want %q", got, want)
			}
			if _, err := Canonicalize(test.input); err == nil {
				t.Fatal("strict Canonicalize accepted Unicode noncharacters")
			}
		})
	}
}

func TestCanonicalizeHistoricalAllocationBoundaryIsPrecharged(t *testing.T) {
	t.Parallel()
	input := []byte(`{"value":[-0,1.5,1e2]}`)
	measurement := &allocationBudget{limit: maximumAllocationBytes}
	want, err := canonicalizeHistoricalWithBudget(input, measurement)
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name     string
		limit    int
		accepted bool
	}{
		{name: "below", limit: measurement.charged - 1, accepted: false},
		{name: "exact", limit: measurement.charged, accepted: true},
		{name: "above", limit: measurement.charged + 1, accepted: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			got, err := canonicalizeHistoricalWithBudget(input, &allocationBudget{limit: test.limit})
			if test.accepted != (err == nil) {
				t.Fatalf("limit=%d accepted=%v error=%v", test.limit, test.accepted, err)
			}
			if err == nil && string(got) != string(want) {
				t.Fatalf("canonical output = %s, want %s", got, want)
			}
		})
	}
}

func TestCanonicalizeExactStringAndArrayBoundaries(t *testing.T) {
	t.Parallel()
	for _, size := range []int{maximumStringBytes - 1, maximumStringBytes} {
		if _, err := Canonicalize([]byte(`"` + strings.Repeat("x", size) + `"`)); err != nil {
			t.Fatalf("string size %d rejected: %v", size, err)
		}
	}
	if _, err := Canonicalize([]byte(`"` + strings.Repeat("x", maximumStringBytes+1) + `"`)); err == nil {
		t.Fatal("above maximum string accepted")
	}
	for _, size := range []int{maximumCollectionEntries - 1, maximumCollectionEntries} {
		input := `[` + strings.TrimSuffix(strings.Repeat("0,", size), ",") + `]`
		if _, err := Canonicalize([]byte(input)); err != nil {
			t.Fatalf("array size %d rejected: %v", size, err)
		}
	}
	tooMany := `[` + strings.TrimSuffix(strings.Repeat("0,", maximumCollectionEntries+1), ",") + `]`
	if _, err := Canonicalize([]byte(tooMany)); err == nil {
		t.Fatal("above maximum array accepted")
	}
}

func TestCanonicalizeExactObjectAndFileBoundaries(t *testing.T) {
	t.Parallel()
	object := func(size int) string {
		var value strings.Builder
		value.WriteByte('{')
		for index := 0; index < size; index++ {
			if index != 0 {
				value.WriteByte(',')
			}
			value.WriteString(`"k`)
			value.WriteString(strconv.Itoa(index))
			value.WriteString(`":0`)
		}
		value.WriteByte('}')
		return value.String()
	}
	for _, size := range []int{maximumCollectionEntries - 1, maximumCollectionEntries} {
		if _, err := Canonicalize([]byte(object(size))); err != nil {
			t.Fatalf("object size %d rejected: %v", size, err)
		}
	}
	if _, err := Canonicalize([]byte(object(maximumCollectionEntries + 1))); err == nil {
		t.Fatal("above maximum object accepted")
	}
	for _, size := range []int{maximumArtifactBytes - 1, maximumArtifactBytes} {
		input := append([]byte(`{}`), []byte(strings.Repeat(" ", size-2))...)
		if _, err := Canonicalize(input); err != nil {
			t.Fatalf("file size %d rejected: %v", size, err)
		}
	}
	if _, err := Canonicalize(make([]byte, maximumArtifactBytes+1)); err == nil {
		t.Fatal("above maximum file accepted")
	}
}

func TestAllocationBudgetExactBoundary(t *testing.T) {
	t.Parallel()
	for _, size := range []int{maximumAllocationBytes - 1, maximumAllocationBytes} {
		budget := &allocationBudget{}
		if err := budget.charge(size); err != nil {
			t.Fatalf("budget %d rejected: %v", size, err)
		}
	}
	budget := &allocationBudget{}
	if err := budget.charge(maximumAllocationBytes + 1); err == nil {
		t.Fatal("above maximum allocation budget accepted")
	}
}

func TestCanonicalizeAllocationBoundaryIsEnforcedAcrossParseAndRender(t *testing.T) {
	t.Parallel()
	input := []byte(`{"value":["alpha","beta",{"nested":true}]}`)
	measurement := &allocationBudget{limit: maximumAllocationBytes}
	want, err := canonicalizeWithBudget(input, measurement)
	if err != nil {
		t.Fatal(err)
	}
	if measurement.charged == 0 {
		t.Fatal("canonicalization charged no capacity")
	}
	for _, test := range []struct {
		name     string
		limit    int
		accepted bool
	}{
		{name: "below", limit: measurement.charged - 1, accepted: false},
		{name: "exact", limit: measurement.charged, accepted: true},
		{name: "above", limit: measurement.charged + 1, accepted: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			got, err := canonicalizeWithBudget(input, &allocationBudget{limit: test.limit})
			if test.accepted != (err == nil) {
				t.Fatalf("limit=%d accepted=%v error=%v", test.limit, test.accepted, err)
			}
			if err == nil && string(got) != string(want) {
				t.Fatalf("canonical output = %s, want %s", got, want)
			}
		})
	}
}

func TestManagedCollectionsRejectBeforeCapacityGrowth(t *testing.T) {
	t.Parallel()
	byteValues := &managedBytes{data: []byte{'a'}, budget: &allocationBudget{limit: 1}}
	bytePointer := &byteValues.data[0]
	if err := byteValues.append('b'); err == nil {
		t.Fatal("byte capacity growth above budget accepted")
	}
	if len(byteValues.data) != 1 || &byteValues.data[0] != bytePointer {
		t.Fatal("byte buffer changed after rejected capacity request")
	}

	arrayValues := &managedValues{data: make([]any, 1, 1), budget: &allocationBudget{limit: 1}}
	arrayPointer := &arrayValues.data[0]
	if err := arrayValues.append(true); err == nil {
		t.Fatal("array capacity growth above budget accepted")
	}
	if len(arrayValues.data) != 1 || &arrayValues.data[0] != arrayPointer {
		t.Fatal("array buffer changed after rejected capacity request")
	}

	objectValues := &managedMembers{data: make([]Member, 1, 1), budget: &allocationBudget{limit: 1}}
	objectPointer := &objectValues.data[0]
	if err := objectValues.append(Member{Name: "next"}); err == nil {
		t.Fatal("object capacity growth above budget accepted")
	}
	if len(objectValues.data) != 1 || &objectValues.data[0] != objectPointer {
		t.Fatal("object buffer changed after rejected capacity request")
	}
}

func TestExternalChargeRejectsDecodeAndRenderBeforeCapacityGrowth(t *testing.T) {
	t.Parallel()
	input := []byte(`{"value":["alpha","beta",{"nested":true}]}`)
	decodeCharge := 0
	value, err := DecodeWithCharge(input, func(size int) error {
		decodeCharge += size
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if decodeCharge == 0 {
		t.Fatal("decode reported no capacity requests")
	}
	renderCharge := 0
	if _, err := CanonicalizeValueWithCharge(value, func(size int) error {
		if renderCharge+size > 1 {
			return errors.New("injected render budget")
		}
		renderCharge += size
		return nil
	}); err == nil || !strings.Contains(err.Error(), "injected render budget") {
		t.Fatalf("render error = %v", err)
	}
	if renderCharge != 0 {
		t.Fatalf("rejected render capacity was charged: %d", renderCharge)
	}
}

func TestCanonicalizeRejectsDuplicateKeysTrailingValuesAndNonIntegerNumbers(t *testing.T) {
	t.Parallel()

	for _, input := range []string{`{"a":1,"a":2}`, `{} {}`, `{"n":1.5}`, `{"n":-0}`, `{"n":9007199254740992}`} {
		if _, err := Canonicalize([]byte(input)); err == nil {
			t.Fatalf("Canonicalize(%q) error = nil", input)
		}
	}
}

func TestCanonicalizeUsesJCSStringEscaping(t *testing.T) {
	t.Parallel()

	got, err := Canonicalize([]byte(`{"s":"<>&\u2028\u2029\b\t\n\f\r"}`))
	if err != nil {
		t.Fatal(err)
	}
	want := "{\"s\":\"<>&\u2028\u2029\\b\\t\\n\\f\\r\"}"
	if string(got) != want {
		t.Fatalf("Canonicalize() = %q, want %q", got, want)
	}
}

func TestCanonicalizeEscapesQuotesBackslashesAndOtherControls(t *testing.T) {
	t.Parallel()
	input := []byte(`{"s":"\"\\\u0000\u001f"}`)
	got, err := Canonicalize(input)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(input) {
		t.Fatalf("Canonicalize() = %q, want %q", got, input)
	}
}

func TestCanonicalizeRejectsLoneSurrogatesAndNoncharacters(t *testing.T) {
	t.Parallel()
	for _, input := range []string{`{"s":"\ud800"}`, `{"s":"\udc00"}`, `{"s":"\ud800x"}`, `{"s":"\ufdd0"}`, "{\"s\":\"\U0001ffff\"}"} {
		if _, err := Canonicalize([]byte(input)); err == nil {
			t.Fatalf("Canonicalize(%q) error = nil", input)
		}
	}
	if _, err := Canonicalize([]byte(`{"s":"\ud83d\ude00"}`)); err != nil {
		t.Fatalf("paired surrogate rejected: %v", err)
	}
}
