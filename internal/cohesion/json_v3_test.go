//nolint:copyloopvar,intrange,modernize,unused // fixture loops and retained fields are intentional
package cohesion

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestDecodeTrustJSONRejectsNonIJSONAndAmbiguousDocuments(t *testing.T) {
	t.Parallel()

	tests := map[string][]byte{
		"byte order mark":       append([]byte{0xef, 0xbb, 0xbf}, []byte(`{"value":1}`)...),
		"invalid utf8":          {'{', '"', 'v', 'a', 'l', 'u', 'e', '"', ':', '"', 0xff, '"', '}'},
		"duplicate root key":    []byte(`{"value":1,"value":2}`),
		"duplicate decoded key": []byte(`{"value":1,"va\u006cue":2}`),
		"duplicate nested key":  []byte(`{"nested":{"value":1,"value":2}}`),
		"trailing value":        []byte(`{"value":1}{"value":2}`),
		"negative zero":         []byte(`{"value":-0}`),
		"fraction":              []byte(`{"value":1.0}`),
		"exponent":              []byte(`{"value":1e0}`),
		"lone high surrogate":   []byte(`{"value":"\ud800"}`),
		"lone low surrogate":    []byte(`{"value":"\udc00"}`),
		"escaped noncharacter":  []byte(`{"value":"\ufdd0"}`),
		"paired noncharacter":   []byte(`{"value":"\udbff\udfff"}`),
		"raw noncharacter":      []byte("{\"value\":\"\ufdd0\"}"),
	}
	for name, input := range tests {
		input := input
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			var target map[string]any
			if err := decodeTrustJSON(input, 1<<20, &target); err == nil {
				t.Fatalf("decodeTrustJSON(%q) error = nil", input)
			}
		})
	}
}

func TestDecodeTrustJSONEnforcesFrozenStructuralBounds(t *testing.T) {
	t.Parallel()

	for name, input := range map[string][]byte{
		"depth":    []byte(strings.Repeat(`[`, 65) + `0` + strings.Repeat(`]`, 65)),
		"string":   []byte(`{"value":"` + strings.Repeat("a", 1<<20+1) + `"}`),
		"members":  objectWithMembers(4097),
		"elements": []byte(`[` + strings.Repeat(`0,`, 4096) + `0]`),
	} {
		var target any
		if err := decodeTrustJSON(input, int64(len(input)), &target); err == nil {
			t.Fatalf("decodeTrustJSON(%s overflow) error = nil", name)
		}
	}

	for name, input := range map[string][]byte{
		"depth":          []byte(strings.Repeat(`[`, 64) + `0` + strings.Repeat(`]`, 64)),
		"string":         []byte(`{"value":"` + strings.Repeat("a", 1<<20) + `"}`),
		"members":        objectWithMembers(4096),
		"elements":       []byte(`[` + strings.Repeat(`0,`, 4095) + `0]`),
		"surrogate pair": []byte(`{"value":"\ud83d\ude00"}`),
		"escaped slash":  []byte(`{"value":"a\/b"}`),
	} {
		var boundary any
		if err := decodeTrustJSON(input, int64(len(input)), &boundary); err != nil {
			t.Fatalf("decodeTrustJSON(%s exact boundary) error = %v", name, err)
		}
	}

	var target map[string]any
	if err := decodeTrustJSON([]byte(`{"value":1,"nested":[true,false,null,"ok"]}`), 1<<20, &target); err != nil {
		t.Fatalf("decodeTrustJSON(valid) error = %v", err)
	}
	if target["value"] == nil || target["nested"] == nil {
		t.Fatalf("decoded target = %#v", target)
	}
}

func TestDecodeTrustJSONRejectsUnknownTypedFields(t *testing.T) {
	t.Parallel()

	var target struct {
		Value int `json:"value"`
	}
	if err := decodeTrustJSON([]byte(`{"value":1,"extra":2}`), 1<<20, &target); err == nil {
		t.Fatal("decodeTrustJSON(unknown field) error = nil")
	}
}

func TestDecodeTrustJSONRejectsWholeFileOverflowBeforeDecoding(t *testing.T) {
	t.Parallel()

	target := map[string]any{"unchanged": true}
	if err := decodeTrustJSON([]byte(`{"value":1}`), 8, &target); err == nil {
		t.Fatal("decodeTrustJSON(oversize) error = nil")
	}
	if len(target) != 1 || target["unchanged"] != true {
		t.Fatalf("target changed on preflight rejection: %#v", target)
	}
}

func TestDecodeTrustJSONChargesThePerParseAllocationBudget(t *testing.T) {
	t.Parallel()

	input := []byte(`{"value":"allocated","items":["one","two"]}`)
	budget := newTrustJSONBudget(int64(len(input)))
	var target any
	if err := decodeTrustJSONWithBudget(input, int64(len(input)), &target, budget); err != nil {
		t.Fatalf("decodeTrustJSONWithBudget(valid) error = %v", err)
	}
	if budget.charged == 0 || budget.charged > budget.limit {
		t.Fatalf("budget = charged %d, limit %d", budget.charged, budget.limit)
	}
}

func TestDecodeTrustJSONRejectsMalformedScannerBoundaries(t *testing.T) {
	t.Parallel()

	invalid := [][]byte{
		{},
		[]byte(`+1`),
		[]byte(`01`),
		[]byte(`tru`),
		[]byte(`{"value":}`),
		[]byte(`{1:2}`),
		[]byte("{\"value\x01\":1}"),
		[]byte(`{"\x":1}`),
		[]byte(`{"\u12":1}`),
		[]byte(`{"\u12xz":1}`),
		[]byte(`{"\ud800\u12xz":1}`),
		[]byte(`{"\ud800\u0041":1}`),
		[]byte(`{"value" 1}`),
		[]byte(`{"value":1 "other":2}`),
		[]byte(`[0 1]`),
		[]byte(`"unterminated`),
		[]byte(`"unterminated\`),
	}
	for _, input := range invalid {
		var target any
		if err := decodeTrustJSON(input, int64(len(input)), &target); err == nil {
			t.Fatalf("decodeTrustJSON(%q) error = nil", input)
		}
	}

	for _, input := range [][]byte{[]byte(`{}`), []byte(`[]`), []byte(`-1`), []byte(`"\uABCD"`), []byte(" \t\r\nnull\n")} {
		var target any
		if err := decodeTrustJSON(input, int64(len(input)), &target); err != nil {
			t.Fatalf("decodeTrustJSON(%q) error = %v", input, err)
		}
	}
}

func TestTrustJSONBudgetAndDecodeFailureBoundaries(t *testing.T) {
	t.Parallel()

	var target any
	if err := decodeTrustJSONWithBudget([]byte(`null`), 4, &target, nil); err == nil {
		t.Fatal("decodeTrustJSONWithBudget(nil budget) error = nil")
	}
	if err := decodeTrustJSONWithBudget([]byte(`[0]`), 3, &target, &trustJSONBudget{}); err == nil {
		t.Fatal("decodeTrustJSONWithBudget(exhausted budget) error = nil")
	}
	var pointerTarget *int
	if err := assignTrustJSON(&pointerTarget, json.Number("1"), &trustJSONBudget{}); err == nil {
		t.Fatal("assignTrustJSON(pointer budget exhaustion) error = nil")
	}
	if err := decodeTrustJSONWithBudget([]byte(`{"a":0}`), 7, &target, &trustJSONBudget{limit: 1}); err == nil {
		t.Fatal("decodeTrustJSONWithBudget(object-key budget exhaustion) error = nil")
	}
	if err := decodeTrustJSONWithBudget([]byte(`"a"`), 3, &target, &trustJSONBudget{}); err == nil {
		t.Fatal("decodeTrustJSONWithBudget(string budget exhaustion) error = nil")
	}
	if err := decodeTrustJSON([]byte(`{"value":"wrong"}`), 1<<20, &struct {
		Value int `json:"value"`
	}{}); err == nil {
		t.Fatal("decodeTrustJSON(type mismatch) error = nil")
	}

	budget := &trustJSONBudget{limit: 1}
	if err := budget.charge(-1); err == nil {
		t.Fatal("charge(negative) error = nil")
	}

	objectDepthOverflow := []byte(strings.Repeat(`{"a":`, 65) + `0` + strings.Repeat(`}`, 65))
	if err := decodeTrustJSON(objectDepthOverflow, int64(len(objectDepthOverflow)), &target); err == nil {
		t.Fatal("decodeTrustJSON(object depth overflow) error = nil")
	}
	if err := decodeTrustJSON([]byte(`"\u12`), 5, &target); err == nil {
		t.Fatal("decodeTrustJSON(truncated Unicode escape) error = nil")
	}

	malformedStringScanner := trustJSONScanner{data: []byte(`x"`), budget: newTrustJSONBudget(2)}
	if _, err := malformedStringScanner.stringValue(); err == nil {
		t.Fatal("stringValue(malformed opening delimiter) error = nil")
	}
}

func TestDecodeTrustJSONChargesCollectionCapacityBeforeGrowth(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		input      []byte
		exactLimit int64
	}{
		{[]byte(`{"a":0}`), 2*trustJSONObjectEntryBytes + int64(reflect.TypeFor[trustJSONPathStep]().Size()) + 2},
		{[]byte(`[0]`), 41},
	} {
		var target any
		if err := decodeTrustJSONWithBudget(test.input, int64(len(test.input)), &target, &trustJSONBudget{limit: test.exactLimit}); err != nil {
			t.Fatalf("decodeTrustJSONWithBudget(%q, exact limit) error = %v", test.input, err)
		}
		if err := decodeTrustJSONWithBudget(test.input, int64(len(test.input)), &target, &trustJSONBudget{limit: test.exactLimit - 1}); err == nil {
			t.Fatalf("decodeTrustJSONWithBudget(%q, one below limit) error = nil", test.input)
		}
	}
}

func TestDecodeTrustJSONReportsRequestedBufferCapacities(t *testing.T) {
	t.Parallel()

	var charges []int64
	budget := &trustJSONBudget{
		limit: 1 << 20,
		observe: func(capacity int64) {
			charges = append(charges, capacity)
		},
	}
	var target any
	input := []byte(`{"a":[0,1,2],"b":0,"c":0}`)
	if err := decodeTrustJSONWithBudget(input, int64(len(input)), &target, budget); err != nil {
		t.Fatal(err)
	}
	want := []int64{1, 24, 48, 1, 16, 1, 32, 1, 64, trustJSONObjectEntryBytes, 1, 1, 2 * trustJSONObjectEntryBytes, 1, 1, 4 * trustJSONObjectEntryBytes, 3 * trustJSONObjectEntryBytes}
	if !reflect.DeepEqual(charges, want) {
		t.Fatalf("buffer capacity charges = %v, want %v", charges, want)
	}
}

func TestAssignTrustJSONPopulatesTypedValuesWithoutReparsing(t *testing.T) {
	t.Parallel()

	type destination struct {
		Text   string         `json:"text,omitempty"`
		Bool   bool           `json:"bool"`
		Int    int8           `json:"int"`
		Uint   uint8          `json:"uint"`
		Number json.Number    `json:"number"`
		Ptr    *int           `json:"ptr"`
		Slice  []string       `json:"slice"`
		Map    map[string]int `json:"map"`
		hidden string
	}
	var got destination
	value := map[string]any{
		"text": "value", "bool": true, "int": json.Number("-2"), "uint": json.Number("2"),
		"number": json.Number("3"), "ptr": json.Number("4"),
		"slice": []any{"a", "b"}, "map": map[string]any{"a": json.Number("5")},
	}
	if err := assignTrustJSON(&got, value, &trustJSONBudget{limit: 1 << 20}); err != nil {
		t.Fatal(err)
	}
	if got.Text != "value" || !got.Bool || got.Int != -2 || got.Uint != 2 || got.Number != "3" || got.Ptr == nil || *got.Ptr != 4 || !reflect.DeepEqual(got.Slice, []string{"a", "b"}) || !reflect.DeepEqual(got.Map, map[string]int{"a": 5}) {
		t.Fatalf("assigned value = %#v", got)
	}
	got.Int = 7
	if err := assignTrustJSON(&got.Int, nil, &trustJSONBudget{}); err != nil || got.Int != 0 {
		t.Fatalf("assign null = %d, %v", got.Int, err)
	}
	var generic any
	if err := assignTrustJSON(&generic, "value", &trustJSONBudget{}); err != nil || generic != "value" {
		t.Fatalf("assign interface = %#v, %v", generic, err)
	}
	if _, exists := trustJSONStructField(reflect.ValueOf(&got).Elem(), "hidden"); exists {
		t.Fatal("unexported field was addressable by JSON name")
	}
}

func TestAssignTrustJSONRejectsIncompatibleDestinations(t *testing.T) {
	t.Parallel()

	var nilPointer *int
	for name, run := range map[string]func() error{
		"nil target":         func() error { return assignTrustJSON(nil, nil, &trustJSONBudget{}) },
		"non-pointer target": func() error { return assignTrustJSON(1, json.Number("1"), &trustJSONBudget{}) },
		"nil pointer target": func() error { return assignTrustJSON(nilPointer, json.Number("1"), &trustJSONBudget{}) },
		"number mismatch":    func() error { var value json.Number; return assignTrustJSON(&value, "1", &trustJSONBudget{}) },
		"pointer budget":     func() error { var value *int; return assignTrustJSON(&value, json.Number("1"), &trustJSONBudget{}) },
		"struct mismatch":    func() error { var value struct{}; return assignTrustJSON(&value, "x", &trustJSONBudget{}) },
		"unknown field": func() error {
			var value struct{}
			return assignTrustJSON(&value, map[string]any{"x": nil}, &trustJSONBudget{})
		},
		"struct field": func() error {
			var value struct {
				X int `json:"x"`
			}
			return assignTrustJSON(&value, map[string]any{"x": "bad"}, &trustJSONBudget{})
		},
		"slice mismatch": func() error { var value []int; return assignTrustJSON(&value, "x", &trustJSONBudget{}) },
		"slice budget": func() error {
			var value []int
			return assignTrustJSON(&value, []any{json.Number("1")}, &trustJSONBudget{})
		},
		"slice item": func() error {
			var value []int
			return assignTrustJSON(&value, []any{"bad"}, &trustJSONBudget{limit: 100})
		},
		"map mismatch": func() error { var value map[string]int; return assignTrustJSON(&value, "x", &trustJSONBudget{}) },
		"map key": func() error {
			var value map[int]int
			return assignTrustJSON(&value, map[string]any{}, &trustJSONBudget{})
		},
		"map budget": func() error {
			var value map[string]int
			return assignTrustJSON(&value, map[string]any{"x": json.Number("1")}, &trustJSONBudget{})
		},
		"map item": func() error {
			var value map[string]int
			return assignTrustJSON(&value, map[string]any{"x": "bad"}, &trustJSONBudget{limit: 100})
		},
		"string mismatch":   func() error { var value string; return assignTrustJSON(&value, true, &trustJSONBudget{}) },
		"boolean mismatch":  func() error { var value bool; return assignTrustJSON(&value, "true", &trustJSONBudget{}) },
		"integer mismatch":  func() error { var value int; return assignTrustJSON(&value, "1", &trustJSONBudget{}) },
		"integer overflow":  func() error { var value int8; return assignTrustJSON(&value, json.Number("128"), &trustJSONBudget{}) },
		"unsigned mismatch": func() error { var value uint; return assignTrustJSON(&value, "1", &trustJSONBudget{}) },
		"unsigned negative": func() error { var value uint; return assignTrustJSON(&value, json.Number("-1"), &trustJSONBudget{}) },
		"unsupported float": func() error { var value float64; return assignTrustJSON(&value, json.Number("1"), &trustJSONBudget{}) },
	} {
		if err := run(); err == nil {
			t.Errorf("%s error = nil", name)
		}
	}
}

func TestTrustJSONCapacityHelpersRejectGrowthBeforeAllocation(t *testing.T) {
	t.Parallel()

	entries := make([]trustJSONObjectEntry, 0, 1)
	if _, err := growTrustJSONObjectEntries(entries, 1, &trustJSONBudget{}); err != nil {
		t.Fatal(err)
	}
	if _, err := growTrustJSONObjectEntries(entries, 2, &trustJSONBudget{}); err == nil {
		t.Fatal("growTrustJSONObjectEntries(exhausted) error = nil")
	}
	if _, err := appendTrustJSONBytes(nil, []byte("a"), maximumTrustJSONString, &trustJSONBudget{}); err == nil {
		t.Fatal("appendTrustJSONBytes(exhausted) error = nil")
	}
	if _, err := appendTrustJSONBytes(make([]byte, maximumTrustJSONString), []byte("a"), maximumTrustJSONString, &trustJSONBudget{limit: 1 << 30}); err == nil {
		t.Fatal("appendTrustJSONBytes(oversized) error = nil")
	}
	if _, err := growTrustJSONPathSlice(nil, 1, &trustJSONBudget{}); err == nil {
		t.Fatal("growTrustJSONPathSlice(exhausted) error = nil")
	}
	if _, err := appendTrustJSONBytes(make([]byte, 786432), []byte("a"), maximumTrustJSONString, &trustJSONBudget{limit: 1 << 20}); err != nil {
		t.Fatalf("appendTrustJSONBytes(clamped capacity) error = %v", err)
	}
	for name, limit := range map[string]int64{"path buffer": 2, "entry buffer": 19} {
		var target any
		if err := decodeTrustJSONWithBudget([]byte(`{"a":0}`), 7, &target, &trustJSONBudget{limit: limit}); err == nil {
			t.Errorf("decodeTrustJSONWithBudget(%s exhaustion) error = nil", name)
		}
	}
	var empty any
	if err := decodeTrustJSON([]byte(`""`), 2, &empty); err != nil || empty != "" {
		t.Fatalf("decodeTrustJSON(empty string) = %#v, %v", empty, err)
	}
}

func TestValidatedStringDecoderPanicsOnlyWhenItsPreconditionIsViolated(t *testing.T) {
	t.Parallel()

	defer func() {
		if recover() == nil {
			t.Fatal("writeValidatedEscape(invalid precondition) did not panic")
		}
	}()
	scanner := trustJSONScanner{data: []byte(`\x`)}
	scanner.writeValidatedEscape(&strings.Builder{})
}

func TestDecodeTrustJSONAcceptsCompositeAtWholeFileBudgetBoundary(t *testing.T) {
	const maximum = 32 << 20
	for _, test := range []struct {
		name         string
		bytes        int
		stringsCount int
	}{
		{name: "exact boundary", bytes: maximum, stringsCount: 32},
		{name: "adversarial distribution", bytes: maximum - 1, stringsCount: 63},
	} {
		document := compositeStringArray(test.bytes, test.stringsCount)
		var target any
		if err := decodeTrustJSON(document, maximum, &target); err != nil {
			t.Fatalf("decodeTrustJSON(%s) error = %v", test.name, err)
		}
	}
}

func compositeStringArray(totalBytes, stringsCount int) []byte {
	payloadBytes := totalBytes - (2 + stringsCount*2 + stringsCount - 1)
	base := payloadBytes / stringsCount
	remainder := payloadBytes % stringsCount
	document := make([]byte, 0, totalBytes)
	document = append(document, '[')
	for index := 0; index < stringsCount; index++ {
		if index > 0 {
			document = append(document, ',')
		}
		size := base
		if index < remainder {
			size++
		}
		document = append(document, '"')
		document = append(document, bytes.Repeat([]byte{'a'}, size)...)
		document = append(document, '"')
	}
	document = append(document, ']')
	if len(document) != totalBytes {
		panic("composite string array length mismatch")
	}
	return document
}

func TestOfficialDownloadsOracleStringExceptionIsPathScoped(t *testing.T) {
	payload := strings.Repeat("a", maximumTrustJSONString+1)
	allowed := []byte(`{"accepted":[{"input_base64":"` + payload + `"}]}`)
	budget := newTrustJSONBudget(int64(len(allowed)))
	if _, err := parseTrustJSONWithBudgetAndLimits(allowed, int64(len(allowed)), budget, officialDownloadsOracleTrustJSONLimits()); err != nil {
		t.Fatalf("parseTrustJSONWithBudgetAndLimits(official oracle) error = %v", err)
	}
	if _, err := parseTrustJSONWithBudget(allowed, int64(len(allowed)), newTrustJSONBudget(int64(len(allowed)))); err == nil {
		t.Fatal("parseTrustJSONWithBudget(generic oversized string) error = nil")
	}
	disallowed := []byte(`{"other":[{"input_base64":"` + payload + `"}]}`)
	if _, err := parseTrustJSONWithBudgetAndLimits(disallowed, int64(len(disallowed)), newTrustJSONBudget(int64(len(disallowed))), officialDownloadsOracleTrustJSONLimits()); err == nil {
		t.Fatal("parseTrustJSONWithBudgetAndLimits(non-oracle path) error = nil")
	}
	limits := officialDownloadsOracleTrustJSONLimits()
	const officialMaximum = 44739244
	if got := limits.maximumStringBytes([]trustJSONPathStep{
		{kind: trustJSONObjectMember, name: "rejected"},
		{kind: trustJSONArrayElement},
		{kind: trustJSONObjectMember, name: "input_base64"},
	}); got != officialMaximum {
		t.Fatalf("official input_base64 limit = %d", got)
	}
	const oracleArtifactMaximum = 128 << 20
	for name, size := range map[string]int{"one below": officialMaximum - 1, "exact": officialMaximum} {
		document := officialOracleInputDocument(size)
		if _, err := parseTrustJSONWithBudgetAndLimits(document, oracleArtifactMaximum, newTrustJSONBudget(oracleArtifactMaximum), limits); err != nil {
			t.Fatalf("parseTrustJSONWithBudgetAndLimits(%s official maximum) error = %v", name, err)
		}
	}
	above := officialOracleInputDocument(officialMaximum + 1)
	if _, err := parseTrustJSONWithBudgetAndLimits(above, oracleArtifactMaximum, newTrustJSONBudget(oracleArtifactMaximum), limits); err == nil {
		t.Fatal("parseTrustJSONWithBudgetAndLimits(above official maximum) error = nil")
	}
}

func TestOfficialDownloadsOracleStringExceptionRequiresAnExactArrayRow(t *testing.T) {
	t.Parallel()

	oversized := strings.Repeat("a", maximumTrustJSONString+1)
	for name, input := range map[string][]byte{
		"direct object": []byte(`{"accepted":{"input_base64":"` + oversized + `"}}`),
		"extra array":   []byte(`{"rejected":[[{"input_base64":"` + oversized + `"}]]}`),
	} {
		if _, err := parseTrustJSONWithBudgetAndLimits(input, int64(len(input)), newTrustJSONBudget(int64(len(input))), officialDownloadsOracleTrustJSONLimits()); err == nil {
			t.Fatalf("parseTrustJSONWithBudgetAndLimits(%s malformed oracle) error = nil", name)
		}
	}

	valid := []byte(`{"accepted":[{"input_base64":"` + oversized + `"}]}`)
	if _, err := parseTrustJSONWithBudgetAndLimits(valid, int64(len(valid)), newTrustJSONBudget(int64(len(valid))), officialDownloadsOracleTrustJSONLimits()); err != nil {
		t.Fatalf("parseTrustJSONWithBudgetAndLimits(valid oracle row) error = %v", err)
	}
}

type trustJSONStringer interface {
	String() string
}

func TestDecodeTrustJSONRejectsValuesThatDoNotImplementDestinationInterface(t *testing.T) {
	t.Parallel()

	var target trustJSONStringer
	if err := decodeTrustJSON([]byte(`"value"`), 7, &target); err == nil {
		t.Fatal("decodeTrustJSON(non-implementing interface value) error = nil")
	}
}

func officialOracleInputDocument(size int) []byte {
	document := make([]byte, 0, size+64)
	document = append(document, `{"rejected":[{"input_base64":"`...)
	document = append(document, bytes.Repeat([]byte{'a'}, size)...)
	return append(document, `"}]}`...)
}

func TestDecodeTrustJSONEnforcesIJSONSafeIntegerRange(t *testing.T) {
	t.Parallel()

	for _, input := range [][]byte{[]byte(`9007199254740991`), []byte(`-9007199254740991`)} {
		var target any
		if err := decodeTrustJSON(input, int64(len(input)), &target); err != nil {
			t.Fatalf("decodeTrustJSON(%s) error = %v", input, err)
		}
	}
	for _, input := range [][]byte{[]byte(`9007199254740992`), []byte(`-9007199254740992`), []byte(`999999999999999999999999999999999999`)} {
		var target any
		if err := decodeTrustJSON(input, int64(len(input)), &target); err == nil {
			t.Fatalf("decodeTrustJSON(%s) error = nil", input)
		}
	}
}

func TestTrustJSONMutationSensitiveBoundaries(t *testing.T) {
	t.Parallel()

	var target any
	err := decodeTrustJSON([]byte{}, 0, &target)
	if err == nil || !strings.Contains(err.Error(), "missing value") {
		t.Fatalf("decodeTrustJSON(empty exact maximum) error = %v", err)
	}
	if err := decodeTrustJSON([]byte(`null`), 4, &target); err != nil {
		t.Fatalf("decodeTrustJSON(exact byte maximum) error = %v", err)
	}
	err = decodeTrustJSON([]byte{0xef, 0xbb, 0xbf}, 3, &target)
	if err == nil || !strings.Contains(err.Error(), "byte-order mark") {
		t.Fatalf("decodeTrustJSON(exact BOM) error = %v", err)
	}
	if got := newTrustJSONBudget(0).limit; got != trustJSONBudgetOverhead {
		t.Fatalf("newTrustJSONBudget(0).limit = %d", got)
	}
	observed := int64(-1)
	budget := &trustJSONBudget{observe: func(value int64) { observed = value }}
	if err := budget.charge(0); err != nil || observed != 0 || budget.charged != 0 {
		t.Fatalf("charge(0) = observed %d, charged %d, error %v", observed, budget.charged, err)
	}
	budget = &trustJSONBudget{limit: 7}
	if err := budget.charge(7); err != nil || budget.charged != 7 {
		t.Fatalf("charge(exact limit) = charged %d, error %v", budget.charged, err)
	}

	for depth, wantError := range map[int]bool{maximumTrustJSONDepth: false, maximumTrustJSONDepth + 1: true} {
		scanner := trustJSONScanner{data: []byte(`{}`), budget: newTrustJSONBudget(2)}
		_, err := scanner.object(depth)
		if (err != nil) != wantError {
			t.Fatalf("object(depth %d) error = %v", depth, err)
		}
	}
	values := make([]any, 0, 1)
	if _, err := growTrustJSONAnySlice(values, 1, &trustJSONBudget{}); err != nil {
		t.Fatalf("growTrustJSONAnySlice(exact capacity) error = %v", err)
	}
	bytesAtCapacity := make([]byte, 0, 1)
	if got, err := appendTrustJSONBytes(bytesAtCapacity, []byte("a"), 1, &trustJSONBudget{}); err != nil || string(got) != "a" {
		t.Fatalf("appendTrustJSONBytes(exact capacity) = %q, %v", got, err)
	}
	zeroLimitScanner := trustJSONScanner{limits: trustJSONParserLimits{maximumStringBytes: func([]trustJSONPathStep) int { return 0 }}}
	if got := zeroLimitScanner.maximumStringBytes(); got != maximumTrustJSONString {
		t.Fatalf("maximumStringBytes(zero override) = %d", got)
	}

	for _, test := range []struct{ current, required, want int }{
		{0, 1, 1}, {1, 1, 1}, {1, 2, 2}, {2, 3, 4}, {8, 3, 8},
	} {
		if got := nextTrustJSONCapacity(test.current, test.required); got != test.want {
			t.Fatalf("nextTrustJSONCapacity(%d, %d) = %d, want %d", test.current, test.required, got, test.want)
		}
	}

	for _, test := range []struct{ encoded, decoded string }{
		{`\"`, `"`}, {`\\`, `\`}, {`\/`, `/`}, {`\b`, "\b"}, {`\t`, "\t"},
		{`\u0041`, "A"}, {`\ud800\udc00`, "𐀀"}, {`\udbff\udfff`, "\U0010ffff"},
	} {
		scanner := trustJSONScanner{data: []byte(test.encoded)}
		var decoded strings.Builder
		scanner.writeValidatedEscape(&decoded)
		if decoded.String() != test.decoded || scanner.offset != len(test.encoded) {
			t.Fatalf("writeValidatedEscape(%q) = %q at %d", test.encoded, decoded.String(), scanner.offset)
		}
	}
	for _, encoded := range []string{`\ud7ff`, `\ue000`} {
		scanner := trustJSONScanner{data: []byte(encoded)}
		if _, err := scanner.escapeWidth(); err != nil || scanner.offset != len(encoded) {
			t.Fatalf("escapeWidth(%q) = offset %d, error %v", encoded, scanner.offset, err)
		}
	}
	for _, encoded := range []string{`\ud800`, `\udbff`, `\udc00`, `\udfff`, `\ud800\udbff`, `\ud800\ue000`} {
		scanner := trustJSONScanner{data: []byte(encoded)}
		if _, err := scanner.escapeWidth(); err == nil {
			t.Fatalf("escapeWidth(%q) error = nil", encoded)
		}
	}
	for _, encoded := range []string{`u0000`, `u0009`, `u000a`, `u000f`, `u00a0`, `u00af`, `u00A0`, `u00AF`} {
		scanner := trustJSONScanner{data: []byte(encoded)}
		if _, err := scanner.unicodeEscape(); err != nil || scanner.offset != len(encoded) {
			t.Fatalf("unicodeEscape(%q) = offset %d, error %v", encoded, scanner.offset, err)
		}
	}

	for _, input := range []string{"0", "1", "9", "10", "true", "false", "null"} {
		if err := decodeTrustJSON([]byte(input), int64(len(input)), &target); err != nil {
			t.Fatalf("decodeTrustJSON(%q) error = %v", input, err)
		}
	}
	for _, input := range []string{"-", "00", "09", "-0", "t", "truf", "fals", "nul"} {
		if err := decodeTrustJSON([]byte(input), int64(len(input)), &target); err == nil {
			t.Fatalf("decodeTrustJSON(%q) error = nil", input)
		}
	}

	var mapped map[string]int8
	if err := assignTrustJSON(&mapped, map[string]any{"a": json.Number("1")}, &trustJSONBudget{limit: 16}); err == nil {
		t.Fatal("assignTrustJSON(map one below entry size) error = nil")
	}
	if err := assignTrustJSON(&mapped, map[string]any{"a": json.Number("1")}, &trustJSONBudget{limit: 17}); err != nil {
		t.Fatalf("assignTrustJSON(map exact entry size) error = %v", err)
	}

	for value, want := range map[byte]int{'/': -1, '0': 0, '9': 9, ':': -1, '`': -1, 'a': 10, 'f': 15, 'g': -1, '@': -1, 'A': 10, 'F': 15, 'G': -1} {
		if got := hexadecimal(value); got != want {
			t.Fatalf("hexadecimal(%q) = %d, want %d", value, got, want)
		}
	}
	for value, want := range map[rune]bool{
		0xfdcf: false, 0xfdd0: true, 0xfdef: true, 0xfdf0: false,
		0xfffd: false, 0xfffe: true, 0xffff: true, 0x10000: false,
		0x10fffd: false, 0x10fffe: true, 0x10ffff: true,
	} {
		if got := noncharacter(value); got != want {
			t.Fatalf("noncharacter(%U) = %t, want %t", value, got, want)
		}
	}
}

func integerText(value int) string {
	if value == 0 {
		return "0"
	}
	digits := make([]byte, 0, 16)
	for value > 0 {
		digits = append(digits, byte('0'+value%10))
		value /= 10
	}
	for left, right := 0, len(digits)-1; left < right; left, right = left+1, right-1 {
		digits[left], digits[right] = digits[right], digits[left]
	}
	return string(digits)
}

func objectWithMembers(count int) []byte {
	var value strings.Builder
	value.WriteByte('{')
	for index := 0; index < count; index++ {
		if index != 0 {
			value.WriteByte(',')
		}
		value.WriteString(`"k`)
		value.WriteString(integerText(index))
		value.WriteString(`":0`)
	}
	value.WriteByte('}')
	return []byte(value.String())
}
