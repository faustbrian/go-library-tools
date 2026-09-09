//nolint:intrange,modernize // loop bounds intentionally exercise index semantics
package cohesion

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestCanonicalizeTrustJSONMatchesRFC8785PropertySortingVector(t *testing.T) {
	t.Parallel()

	input := []byte(`{
        "\u20ac": "Euro Sign",
        "\r": "Carriage Return",
        "\ufb33": "Hebrew Letter Dalet With Dagesh",
        "1": "One",
        "\ud83d\ude00": "Emoji: Grinning Face",
        "\u0080": "Control",
        "\u00f6": "Latin Small Letter O With Diaeresis"
      }`)
	want := []byte("{\"\\r\":\"Carriage Return\",\"1\":\"One\",\"\u0080\":\"Control\",\"ö\":\"Latin Small Letter O With Diaeresis\",\"€\":\"Euro Sign\",\"😀\":\"Emoji: Grinning Face\",\"דּ\":\"Hebrew Letter Dalet With Dagesh\"}")

	got, err := canonicalizeTrustJSON(input, 1<<20)
	if err != nil {
		t.Fatalf("canonicalizeTrustJSON() error = %v", err)
	}
	if string(got) != string(want) {
		t.Fatalf("canonicalizeTrustJSON() = %q, want %q", got, want)
	}
}

func TestCanonicalizeTrustJSONRecursesAndSerializesPrimitives(t *testing.T) {
	t.Parallel()

	input := []byte(" { \"z\" : [{\"b\":2,\"a\":1}], \"text\": \"<>&\\u2028\\u000f\\n\\\\\\\"\", \"bool\":true, \"nil\":null } ")
	want := []byte("{\"bool\":true,\"nil\":null,\"text\":\"<>&\u2028\\u000f\\n\\\\\\\"\",\"z\":[{\"a\":1,\"b\":2}]}")

	got, err := canonicalizeTrustJSON(input, 1<<20)
	if err != nil {
		t.Fatalf("canonicalizeTrustJSON() error = %v", err)
	}
	if string(got) != string(want) {
		t.Fatalf("canonicalizeTrustJSON() = %q, want %q", got, want)
	}
}

func TestCanonicalDigestDomainsUseFixedIndependentOracles(t *testing.T) {
	t.Parallel()

	exactInput := []byte("{ \"b\": 2, \"a\": 1 }\n")
	wantExact := sha256.Sum256(exactInput)
	if got := exactBytesSHA256(exactInput); got != "sha256:"+hex.EncodeToString(wantExact[:]) {
		t.Fatalf("exactBytesSHA256() = %q", got)
	}

	semantic, err := semanticObjectSHA256(map[string]any{
		"digest": "sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
		"b":      2,
		"a":      1,
	}, "digest", 1<<20)
	if err != nil {
		t.Fatalf("semanticObjectSHA256() error = %v", err)
	}
	const wantSemantic = "sha256:43258cff783fe7036d8a43033f830adfc60ec037382473548ac742b888292777"
	if semantic != wantSemantic {
		t.Fatalf("semanticObjectSHA256() = %q, want %q", semantic, wantSemantic)
	}
}

func TestDecodeCanonicalTrustJSONRejectsEquivalentNoncanonicalBytes(t *testing.T) {
	t.Parallel()

	var target map[string]any
	if err := decodeCanonicalTrustJSON([]byte(`{"a":1,"b":2}`), 1<<20, &target); err != nil {
		t.Fatalf("decodeCanonicalTrustJSON(canonical) error = %v", err)
	}
	if err := decodeCanonicalTrustJSON([]byte(`{"b":2,"a":1}`), 1<<20, &target); err == nil {
		t.Fatal("decodeCanonicalTrustJSON(noncanonical ordering) error = nil")
	}
	if err := decodeCanonicalTrustJSON([]byte("{\"a\":1,\"b\":2}\n"), 1<<20, &target); err == nil {
		t.Fatal("decodeCanonicalTrustJSON(trailing newline) error = nil")
	}
	if err := decodeCanonicalTrustJSON([]byte(`{`), 1<<20, &target); err == nil {
		t.Fatal("decodeCanonicalTrustJSON(invalid JSON) error = nil")
	}
}

func TestCanonicalDigestRejectsInvalidSemanticPreimages(t *testing.T) {
	t.Parallel()

	if _, err := canonicalMarshal(func() {}, 1<<20); err == nil {
		t.Fatal("canonicalMarshal(unsupported value) error = nil")
	}
	if _, err := semanticObjectSHA256(func() {}, "digest", 1<<20); err == nil {
		t.Fatal("semanticObjectSHA256(unsupported value) error = nil")
	}
	if _, err := semanticObjectSHA256([]any{1}, "digest", 1<<20); err == nil {
		t.Fatal("semanticObjectSHA256(array) error = nil")
	}
	if _, err := semanticObjectSHA256(map[string]any{"value": 1}, "digest", 1<<20); err == nil {
		t.Fatal("semanticObjectSHA256(missing omitted field) error = nil")
	}
}

func TestCanonicalWriterCoversEveryIntegerOnlyPrimitiveAndEscape(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	if err := writeCanonicalJSON(&output, []any{nil, false, true, json.Number("-1"), "\b\t\n\f\r\x01\"\\"}, newTrustJSONBudget(1<<20)); err != nil {
		t.Fatalf("writeCanonicalJSON() error = %v", err)
	}
	const want = `[null,false,true,-1,"\b\t\n\f\r\u0001\"\\"]`
	if output.String() != want {
		t.Fatalf("writeCanonicalJSON() = %q, want %q", output.String(), want)
	}

	for _, test := range []struct {
		left  string
		right string
		want  int
	}{
		{"a", "aa", -1},
		{"aa", "a", 1},
		{"a", "a", 0},
		{"b", "a", 1},
		{"a", "b", -1},
		{"😀", "דּ", -1},
	} {
		if got := compareUTF16(test.left, test.right); got != test.want {
			t.Fatalf("compareUTF16(%q, %q) = %d, want %d", test.left, test.right, got, test.want)
		}
	}
}

func TestUTF16CursorEmitsPendingSurrogateAndEndOfInput(t *testing.T) {
	t.Parallel()

	cursor := utf16Cursor{text: "😀"}
	if _, ok := cursor.next(); !ok {
		t.Fatal("utf16Cursor.next() high surrogate reported end of input")
	}
	if _, ok := cursor.next(); !ok {
		t.Fatal("utf16Cursor.next() pending low surrogate reported end of input")
	}
	if _, ok := cursor.next(); ok {
		t.Fatal("utf16Cursor.next() after input reported a unit")
	}
}

func TestCanonicalWriterPanicsOnlyForAnImpossibleDecodedType(t *testing.T) {
	t.Parallel()

	defer func() {
		if recover() == nil {
			t.Fatal("writeCanonicalJSON(unsupported) did not panic")
		}
	}()
	_ = writeCanonicalJSON(&bytes.Buffer{}, 1, newTrustJSONBudget(1<<20))
}

func TestCanonicalizeTrustJSONRejectsIntegersOutsideIJSONRange(t *testing.T) {
	t.Parallel()

	if _, err := canonicalizeTrustJSON([]byte(`{"value":9007199254740992}`), 1<<20); err == nil {
		t.Fatal("canonicalizeTrustJSON(unsafe integer) error = nil")
	}
}

func TestCanonicalizeTrustJSONChargesOutputAndKeyOrderingAllocations(t *testing.T) {
	t.Parallel()

	input := []byte(`{"b":2,"a":1}`)
	parserBudget := newTrustJSONBudget(int64(len(input)))
	if _, err := parseTrustJSONWithBudget(input, int64(len(input)), parserBudget); err != nil {
		t.Fatalf("parseTrustJSONWithBudget() error = %v", err)
	}

	canonicalBudget := newTrustJSONBudget(int64(len(input)))
	got, err := canonicalizeTrustJSONWithBudget(input, int64(len(input)), canonicalBudget)
	if err != nil {
		t.Fatalf("canonicalizeTrustJSONWithBudget() error = %v", err)
	}
	if string(got) != `{"a":1,"b":2}` {
		t.Fatalf("canonicalizeTrustJSONWithBudget() = %q", got)
	}
	wantAdditional := int64(len(input)) + 2*int64(reflect.TypeFor[string]().Size())
	if gotAdditional := canonicalBudget.charged - parserBudget.charged; gotAdditional != wantAdditional {
		t.Fatalf("canonical allocation charge = %d, want %d", gotAdditional, wantAdditional)
	}
}

func TestCanonicalWriterPropagatesAllocationBudgetFailures(t *testing.T) {
	t.Parallel()

	if _, err := canonicalizeTrustJSONValue(nil, 1, &trustJSONBudget{limit: 0}); err == nil {
		t.Fatal("canonicalizeTrustJSONValue() budget failure = nil")
	}
	if _, err := canonicalizeTrustJSONValue(map[string]any{"key": "value"}, 0, &trustJSONBudget{limit: 0}); err == nil {
		t.Fatal("canonicalizeTrustJSONValue() map budget failure = nil")
	}

	var output bytes.Buffer
	if err := writeCanonicalJSON(&output, map[string]any{"key": "value"}, &trustJSONBudget{limit: 0}); err == nil {
		t.Fatal("writeCanonicalJSON() map budget failure = nil")
	}
	if err := writeCanonicalJSON(&output, []any{map[string]any{"nested": map[string]any{"key": "value"}}}, &trustJSONBudget{limit: 24}); err == nil {
		t.Fatal("writeCanonicalJSON() nested budget failure = nil")
	}
}

func TestCanonicalizeTrustJSONAcceptsTheExactLargeArtifactBoundary(t *testing.T) {
	t.Parallel()

	const maximum = 32 << 20
	const stringsCount = 32
	overhead := 2 + stringsCount*2 + stringsCount - 1
	payload := maximum - overhead
	var input strings.Builder
	input.Grow(maximum)
	input.WriteByte('[')
	for index := 0; index < stringsCount; index++ {
		if index != 0 {
			input.WriteByte(',')
		}
		input.WriteByte('"')
		length := payload / stringsCount
		if index < payload%stringsCount {
			length++
		}
		input.WriteString(strings.Repeat("a", length))
		input.WriteByte('"')
	}
	input.WriteByte(']')
	document := []byte(input.String())
	if len(document) != maximum {
		t.Fatalf("large canonical input length = %d, want %d", len(document), maximum)
	}
	budget := newTrustJSONBudget(maximum)
	got, err := canonicalizeTrustJSONWithBudget(document, maximum, budget)
	if err != nil {
		t.Fatalf("canonicalizeTrustJSONWithBudget(exact large boundary) error = %v", err)
	}
	if !bytes.Equal(got, document) {
		t.Fatal("canonicalizeTrustJSONWithBudget(exact large boundary) changed canonical bytes")
	}
	if budget.charged > budget.limit {
		t.Fatalf("canonical allocation charge = %d, limit %d", budget.charged, budget.limit)
	}
}
