// Package jcs implements deterministic JSON canonicalization and bounded
// decoding helpers.
package jcs

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"unicode/utf16"
	"unicode/utf8"
)

const (
	maximumArtifactBytes     = 32 << 20
	maximumDepth             = 64
	maximumHistoricalDepth   = 10000
	maximumStringBytes       = 1 << 20
	maximumCollectionEntries = 4096
	maximumAllocationBytes   = 2*maximumArtifactBytes + (8 << 20)
)

// Member is an object member used by the canonical JSON value model.
type Member struct {
	Name  string
	Value any
}

// Object is an ordered collection of JSON object members.
type Object []Member

// Number stores a canonical JSON number representation.
type Number string

type allocationBudget struct {
	charged        int
	limit          int
	chargeExternal func(int) error
}

func (budget *allocationBudget) charge(size int) error {
	limit := budget.limit
	if limit == 0 {
		limit = maximumAllocationBytes
	}
	if size < 0 || budget.charged > limit-size {
		return errors.New("JSON allocation budget exceeded")
	}
	if budget.chargeExternal != nil {
		if err := budget.chargeExternal(size); err != nil {
			return err
		}
	}
	budget.charged += size
	return nil
}

// Canonicalize parses strict I-JSON input and returns RFC 8785 canonical JSON.
func Canonicalize(data []byte) ([]byte, error) {
	return canonicalizeWithBudget(data, &allocationBudget{})
}

// CanonicalizeHistorical canonicalizes JSON already accepted and emitted by
// the released encoding/json readers. It preserves their numeric and nesting
// semantics without weakening Canonicalize's strict v3 input contract.
func CanonicalizeHistorical(data []byte) ([]byte, error) {
	return canonicalizeHistoricalWithBudget(data, &allocationBudget{})
}

// Decode parses strict I-JSON input into the package's JSON value model.
func Decode(data []byte) (any, error) {
	return decodeWithBudget(data, &allocationBudget{})
}

// DecodeWithCharge decodes JSON while charging every managed capacity request
// before the corresponding allocation is made.
func DecodeWithCharge(data []byte, charge func(int) error) (any, error) {
	if charge == nil {
		return nil, errors.New("allocation charge callback is required")
	}
	return decodeWithBudget(data, &allocationBudget{chargeExternal: charge})
}

// CanonicalizeValue renders a supported JSON value as RFC 8785 canonical JSON.
func CanonicalizeValue(value any) ([]byte, error) {
	output := managedBytes{budget: &allocationBudget{}, maximum: maximumArtifactBytes, maximumError: "canonical JSON exceeds 32 MiB"}
	if err := appendValue(&output, value); err != nil {
		return nil, err
	}
	return output.data, nil
}

// CanonicalizeValueWithCharge renders a decoded value while charging every
// output capacity request before the corresponding allocation is made.
func CanonicalizeValueWithCharge(value any, charge func(int) error) ([]byte, error) {
	if charge == nil {
		return nil, errors.New("allocation charge callback is required")
	}
	output := managedBytes{budget: &allocationBudget{chargeExternal: charge}, maximum: maximumArtifactBytes, maximumError: "canonical JSON exceeds 32 MiB"}
	if err := appendValue(&output, value); err != nil {
		return nil, err
	}
	return output.data, nil
}

// Marshal converts supported Go JSON values and renders them canonically while
// charging every managed conversion and output capacity before allocation.
func Marshal(value any) ([]byte, error) {
	return marshalWithBudget(value, &allocationBudget{})
}

// MarshalWithCharge is Marshal with an additional capacity-charge callback.
func MarshalWithCharge(value any, charge func(int) error) ([]byte, error) {
	if charge == nil {
		return nil, errors.New("allocation charge callback is required")
	}
	return marshalWithBudget(value, &allocationBudget{chargeExternal: charge})
}

func marshalWithBudget(value any, budget *allocationBudget) ([]byte, error) {
	converted, err := convertGoValue(reflect.ValueOf(value), budget, 1)
	if err != nil {
		return nil, err
	}
	output := managedBytes{budget: budget, maximum: maximumArtifactBytes, maximumError: "canonical JSON exceeds 32 MiB"}
	if err := appendValue(&output, converted); err != nil {
		return nil, err
	}
	return output.data, nil
}

func convertGoValue(value reflect.Value, budget *allocationBudget, depth int) (any, error) {
	if !value.IsValid() {
		return nil, nil
	}
	for value.Kind() == reflect.Interface || value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return nil, nil
		}
		value = value.Elem()
	}
	switch value.Kind() {
	case reflect.Bool:
		return value.Bool(), nil
	case reflect.String:
		text := value.String()
		if len(text) > maximumStringBytes {
			return nil, errors.New("string exceeds 1 MiB")
		}
		if !utf8.ValidString(text) {
			return nil, errors.New("invalid UTF-8")
		}
		for _, runeValue := range text {
			if isNoncharacter(runeValue) {
				return nil, fmt.Errorf("unicode noncharacter U+%04X", runeValue)
			}
		}
		return text, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		integer := value.Int()
		const maximumSafeInteger = int64(1<<53 - 1)
		if integer < -maximumSafeInteger || integer > maximumSafeInteger {
			return nil, errors.New("integer outside I-JSON safe range")
		}
		var digits [64]byte
		formatted := strconv.AppendInt(digits[:0], integer, 10)
		if err := budget.charge(len(formatted)); err != nil {
			return nil, err
		}
		return Number(string(formatted)), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		integer := value.Uint()
		if integer > 1<<53-1 {
			return nil, errors.New("integer outside I-JSON safe range")
		}
		var digits [64]byte
		formatted := strconv.AppendUint(digits[:0], integer, 10)
		if err := budget.charge(len(formatted)); err != nil {
			return nil, err
		}
		return Number(string(formatted)), nil
	case reflect.Slice, reflect.Array:
		if depth > maximumDepth {
			return nil, fmt.Errorf("JSON nesting depth exceeds %d", maximumDepth)
		}
		if value.Kind() == reflect.Slice && value.IsNil() {
			return nil, nil
		}
		if value.Len() > maximumCollectionEntries {
			return nil, errors.New("array exceeds 4096 elements")
		}
		values := managedValues{budget: budget}
		for index := range value.Len() {
			converted, err := convertGoValue(value.Index(index), budget, depth+1)
			if err != nil {
				return nil, err
			}
			if err := values.append(converted); err != nil {
				return nil, err
			}
		}
		return values.data, nil
	case reflect.Map:
		if depth > maximumDepth {
			return nil, fmt.Errorf("JSON nesting depth exceeds %d", maximumDepth)
		}
		if value.IsNil() {
			return nil, nil
		}
		if value.Type().Key().Kind() != reflect.String {
			return nil, errors.New("JSON object map key is not a string")
		}
		if value.Len() > maximumCollectionEntries {
			return nil, errors.New("object exceeds 4096 members")
		}
		members := managedMembers{budget: budget}
		iterator := value.MapRange()
		for iterator.Next() {
			name := iterator.Key().String()
			converted, err := convertGoValue(iterator.Value(), budget, depth+1)
			if err != nil {
				return nil, err
			}
			if err := members.append(Member{Name: name, Value: converted}); err != nil {
				return nil, err
			}
		}
		return Object(members.data), nil
	case reflect.Struct:
		if depth > maximumDepth {
			return nil, fmt.Errorf("JSON nesting depth exceeds %d", maximumDepth)
		}
		members := managedMembers{budget: budget}
		typeValue := value.Type()
		for index := range value.NumField() {
			field := typeValue.Field(index)
			if field.PkgPath != "" {
				continue
			}
			name, omitEmpty, omitted := jsonField(field)
			if omitted || omitEmpty && value.Field(index).IsZero() {
				continue
			}
			converted, err := convertGoValue(value.Field(index), budget, depth+1)
			if err != nil {
				return nil, err
			}
			if err := members.append(Member{Name: name, Value: converted}); err != nil {
				return nil, err
			}
		}
		return Object(members.data), nil
	default:
		return nil, fmt.Errorf("unsupported JSON value type %s", value.Type())
	}
}

func jsonField(field reflect.StructField) (name string, omitEmpty, omitted bool) {
	tag := field.Tag.Get("json")
	if tag == "-" {
		return "", false, true
	}
	name = field.Name
	if comma := strings.IndexByte(tag, ','); comma >= 0 {
		if comma != 0 {
			name = tag[:comma]
		}
		options := tag[comma+1:]
		omitEmpty = options == "omitempty" || strings.HasPrefix(options, "omitempty,") || strings.HasSuffix(options, ",omitempty") || strings.Contains(options, ",omitempty,")
	} else if tag != "" {
		name = tag
	}
	return name, omitEmpty, false
}

func canonicalizeWithBudget(data []byte, budget *allocationBudget) ([]byte, error) {
	value, err := decodeWithOptions(data, budget, parserOptions{
		maximumDepth:             maximumDepth,
		maximumStringBytes:       maximumStringBytes,
		maximumCollectionEntries: maximumCollectionEntries,
	})
	if err != nil {
		return nil, err
	}
	output := managedBytes{budget: budget, maximum: maximumArtifactBytes, maximumError: "canonical JSON exceeds 32 MiB"}
	if err := appendValue(&output, value); err != nil {
		return nil, err
	}
	return output.data, nil
}

func canonicalizeHistoricalWithBudget(data []byte, budget *allocationBudget) ([]byte, error) {
	value, err := decodeWithOptions(data, budget, parserOptions{
		maximumDepth:             maximumHistoricalDepth,
		maximumStringBytes:       maximumArtifactBytes,
		maximumCollectionEntries: maximumArtifactBytes,
		historicalNumbers:        true,
		allowNoncharacters:       true,
	})
	if err != nil {
		return nil, err
	}
	output := managedBytes{budget: budget, maximum: maximumArtifactBytes, maximumError: "canonical JSON exceeds 32 MiB"}
	if err := appendValue(&output, value); err != nil {
		return nil, err
	}
	return output.data, nil
}

func decodeWithBudget(data []byte, budget *allocationBudget) (any, error) {
	return decodeWithOptions(data, budget, parserOptions{
		maximumDepth:             maximumDepth,
		maximumStringBytes:       maximumStringBytes,
		maximumCollectionEntries: maximumCollectionEntries,
	})
}

type parserOptions struct {
	maximumDepth             int
	maximumStringBytes       int
	maximumCollectionEntries int
	historicalNumbers        bool
	allowNoncharacters       bool
}

func decodeWithOptions(data []byte, budget *allocationBudget, options parserOptions) (any, error) {
	if len(data) > maximumArtifactBytes {
		return nil, fmt.Errorf("JSON exceeds %d bytes", maximumArtifactBytes)
	}
	if !utf8.Valid(data) {
		return nil, errors.New("invalid UTF-8")
	}
	parser := parser{data: data, budget: budget, options: options}
	value, err := parser.value(1)
	if err != nil {
		return nil, err
	}
	parser.space()
	if parser.index != len(data) {
		return nil, errors.New("trailing JSON value")
	}
	return value, nil
}

type parser struct {
	data    []byte
	index   int
	budget  *allocationBudget
	options parserOptions
}

func (parser *parser) value(depth int) (any, error) {
	parser.space()
	if parser.index == len(parser.data) {
		return nil, errors.New("unexpected end of JSON")
	}
	switch parser.data[parser.index] {
	case '{':
		return parser.object(depth)
	case '[':
		return parser.array(depth)
	case '"':
		return parser.string()
	case 't':
		return parser.literal("true", true)
	case 'f':
		return parser.literal("false", false)
	case 'n':
		return parser.literal("null", nil)
	default:
		return parser.number()
	}
}

func (parser *parser) object(depth int) (any, error) {
	if depth > parser.options.maximumDepth {
		return nil, fmt.Errorf("JSON nesting depth exceeds %d", parser.options.maximumDepth)
	}
	parser.index++
	parser.space()
	members := managedMembers{budget: parser.budget}
	if parser.take('}') {
		return Object{}, nil
	}
	for {
		parser.space()
		if parser.index == len(parser.data) || parser.data[parser.index] != '"' {
			return nil, errors.New("object member name is not a string")
		}
		name, err := parser.string()
		if err != nil {
			return nil, err
		}
		for _, member := range members.data {
			if member.Name == name {
				return nil, fmt.Errorf("duplicate object member %q", name)
			}
		}
		if len(members.data) == parser.options.maximumCollectionEntries {
			return nil, fmt.Errorf("object exceeds %d members", parser.options.maximumCollectionEntries)
		}
		parser.space()
		if !parser.take(':') {
			return nil, errors.New("missing object member colon")
		}
		value, err := parser.value(depth + 1)
		if err != nil {
			return nil, err
		}
		members.maximum = parser.options.maximumCollectionEntries
		if err := members.append(Member{Name: name, Value: value}); err != nil {
			return nil, err
		}
		parser.space()
		if parser.take('}') {
			return Object(members.data), nil
		}
		if !parser.take(',') {
			return nil, errors.New("missing object member separator")
		}
	}
}

func (parser *parser) array(depth int) (any, error) {
	if depth > parser.options.maximumDepth {
		return nil, fmt.Errorf("JSON nesting depth exceeds %d", parser.options.maximumDepth)
	}
	parser.index++
	parser.space()
	values := managedValues{budget: parser.budget}
	if parser.take(']') {
		return []any{}, nil
	}
	for {
		if len(values.data) == parser.options.maximumCollectionEntries {
			return nil, fmt.Errorf("array exceeds %d elements", parser.options.maximumCollectionEntries)
		}
		value, err := parser.value(depth + 1)
		if err != nil {
			return nil, err
		}
		values.maximum = parser.options.maximumCollectionEntries
		if err := values.append(value); err != nil {
			return nil, err
		}
		parser.space()
		if parser.take(']') {
			return values.data, nil
		}
		if !parser.take(',') {
			return nil, errors.New("missing array element separator")
		}
	}
}

func (parser *parser) string() (string, error) {
	parser.index++
	decoded := managedBytes{budget: parser.budget, maximum: parser.options.maximumStringBytes, maximumError: fmt.Sprintf("string exceeds %d bytes", parser.options.maximumStringBytes)}
	for parser.index < len(parser.data) {
		current := parser.data[parser.index]
		if current == '"' {
			parser.index++
			if err := parser.budget.charge(len(decoded.data)); err != nil {
				return "", err
			}
			return string(decoded.data), nil
		}
		if current < 0x20 {
			return "", errors.New("unescaped control character in string")
		}
		if current != '\\' {
			runeValue, size := utf8.DecodeRune(parser.data[parser.index:])
			if !parser.options.allowNoncharacters && isNoncharacter(runeValue) {
				return "", fmt.Errorf("unicode noncharacter U+%04X", runeValue)
			}
			if err := decoded.append(parser.data[parser.index : parser.index+size]...); err != nil {
				return "", err
			}
			parser.index += size
			continue
		}
		parser.index++
		if parser.index == len(parser.data) {
			return "", errors.New("unterminated string escape")
		}
		escape := parser.data[parser.index]
		parser.index++
		switch escape {
		case '"', '\\', '/':
			if err := decoded.append(escape); err != nil {
				return "", err
			}
		case 'b':
			if err := decoded.append('\b'); err != nil {
				return "", err
			}
		case 'f':
			if err := decoded.append('\f'); err != nil {
				return "", err
			}
		case 'n':
			if err := decoded.append('\n'); err != nil {
				return "", err
			}
		case 'r':
			if err := decoded.append('\r'); err != nil {
				return "", err
			}
		case 't':
			if err := decoded.append('\t'); err != nil {
				return "", err
			}
		case 'u':
			runeValue, err := parser.unicodeEscape()
			if err != nil {
				return "", err
			}
			if !parser.options.allowNoncharacters && isNoncharacter(runeValue) {
				return "", fmt.Errorf("unicode noncharacter U+%04X", runeValue)
			}
			var encoded [utf8.UTFMax]byte
			size := utf8.EncodeRune(encoded[:], runeValue)
			if err := decoded.append(encoded[:size]...); err != nil {
				return "", err
			}
		default:
			return "", fmt.Errorf("invalid string escape %q", escape)
		}
	}
	return "", errors.New("unterminated string")
}

func (parser *parser) unicodeEscape() (rune, error) {
	first, err := parser.hexRune()
	if err != nil {
		return 0, err
	}
	if first >= 0xdc00 && first <= 0xdfff {
		return 0, errors.New("unpaired low surrogate escape")
	}
	if first < 0xd800 || first > 0xdbff {
		return first, nil
	}
	if parser.index+2 > len(parser.data) || parser.data[parser.index] != '\\' || parser.data[parser.index+1] != 'u' {
		return 0, errors.New("unpaired high surrogate escape")
	}
	parser.index += 2
	second, err := parser.hexRune()
	if err != nil || second < 0xdc00 || second > 0xdfff {
		return 0, errors.New("unpaired high surrogate escape")
	}
	return utf16.DecodeRune(first, second), nil
}

func (parser *parser) hexRune() (rune, error) {
	if parser.index+4 > len(parser.data) {
		return 0, errors.New("short Unicode escape")
	}
	value := rune(0)
	for _, digit := range parser.data[parser.index : parser.index+4] {
		value <<= 4
		switch {
		case digit >= '0' && digit <= '9':
			value += rune(digit - '0')
		case digit >= 'a' && digit <= 'f':
			value += rune(digit-'a') + 10
		case digit >= 'A' && digit <= 'F':
			value += rune(digit-'A') + 10
		default:
			return 0, errors.New("invalid Unicode escape")
		}
	}
	parser.index += 4
	return value, nil
}

func (parser *parser) number() (any, error) {
	start := parser.index
	if parser.take('-') && parser.index == len(parser.data) {
		return nil, errors.New("invalid number")
	}
	if parser.take('0') {
		if parser.index < len(parser.data) && parser.data[parser.index] >= '0' && parser.data[parser.index] <= '9' {
			return nil, errors.New("invalid leading zero")
		}
	} else {
		if parser.index == len(parser.data) || parser.data[parser.index] < '1' || parser.data[parser.index] > '9' {
			return nil, errors.New("invalid number")
		}
		for parser.index < len(parser.data) && parser.data[parser.index] >= '0' && parser.data[parser.index] <= '9' {
			parser.index++
		}
	}
	if parser.index < len(parser.data) && parser.data[parser.index] == '.' {
		parser.index++
		if parser.index == len(parser.data) || parser.data[parser.index] < '0' || parser.data[parser.index] > '9' {
			return nil, errors.New("invalid number fraction")
		}
		for parser.index < len(parser.data) && parser.data[parser.index] >= '0' && parser.data[parser.index] <= '9' {
			parser.index++
		}
	}
	if parser.index < len(parser.data) && (parser.data[parser.index] == 'e' || parser.data[parser.index] == 'E') {
		parser.index++
		if parser.index < len(parser.data) && (parser.data[parser.index] == '+' || parser.data[parser.index] == '-') {
			parser.index++
		}
		if parser.index == len(parser.data) || parser.data[parser.index] < '0' || parser.data[parser.index] > '9' {
			return nil, errors.New("invalid number exponent")
		}
		for parser.index < len(parser.data) && parser.data[parser.index] >= '0' && parser.data[parser.index] <= '9' {
			parser.index++
		}
	}
	textBytes := parser.data[start:parser.index]
	if parser.options.historicalNumbers {
		return parser.historicalNumber(textBytes)
	}
	if len(textBytes) == 2 && textBytes[0] == '-' && textBytes[1] == '0' || bytes.ContainsAny(textBytes, ".eE") {
		return nil, fmt.Errorf("non-integer number %q", textBytes)
	}
	if err := parser.budget.charge(len(textBytes)); err != nil {
		return nil, err
	}
	text := string(textBytes)
	integer, err := strconv.ParseInt(text, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("integer outside int64: %w", err)
	}
	const maximumSafeInteger = int64(1<<53 - 1)
	if integer < -maximumSafeInteger || integer > maximumSafeInteger {
		return nil, errors.New("integer outside I-JSON safe range")
	}
	return Number(text), nil
}

func (parser *parser) historicalNumber(textBytes []byte) (any, error) {
	if err := parser.budget.charge(len(textBytes)); err != nil {
		return nil, err
	}
	number, err := strconv.ParseFloat(string(textBytes), 64)
	if err != nil {
		return nil, fmt.Errorf("number outside IEEE 754 binary64: %w", err)
	}
	if number == 0 {
		return Number("0"), nil
	}
	format, precision := byte('f'), -1
	absolute := number
	if absolute < 0 {
		absolute = -absolute
	}
	if absolute >= 1e21 || absolute < 1e-6 {
		format = 'e'
	}
	var formattedStorage [64]byte
	formatted := strconv.AppendFloat(formattedStorage[:0], number, format, precision, 64)
	if format == 'e' {
		formatted = normalizeExponent(formatted)
	}
	if err := parser.budget.charge(len(formatted)); err != nil {
		return nil, err
	}
	return Number(string(formatted)), nil
}

func normalizeExponent(value []byte) []byte {
	exponent := bytes.IndexByte(value, 'e')
	if exponent < 0 || exponent+2 >= len(value) {
		return value
	}
	digits := exponent + 1
	if value[digits] == '+' || value[digits] == '-' {
		digits++
	}
	for digits+1 < len(value) && value[digits] == '0' {
		copy(value[digits:], value[digits+1:])
		value = value[:len(value)-1]
	}
	return value
}

func (parser *parser) literal(text string, value any) (any, error) {
	if len(parser.data)-parser.index < len(text) {
		return nil, fmt.Errorf("invalid literal at byte %d", parser.index)
	}
	for index := range len(text) {
		if parser.data[parser.index+index] != text[index] {
			return nil, fmt.Errorf("invalid literal at byte %d", parser.index)
		}
	}
	parser.index += len(text)
	return value, nil
}

func (parser *parser) space() {
	for parser.index < len(parser.data) && strings.ContainsRune(" \t\r\n", rune(parser.data[parser.index])) {
		parser.index++
	}
}

func (parser *parser) take(value byte) bool {
	if parser.index < len(parser.data) && parser.data[parser.index] == value {
		parser.index++
		return true
	}
	return false
}

func isNoncharacter(value rune) bool {
	return value >= 0xfdd0 && value <= 0xfdef || value <= utf8.MaxRune && value&0xffff >= 0xfffe
}

type managedBytes struct {
	data         []byte
	budget       *allocationBudget
	maximum      int
	maximumError string
}

func (values *managedBytes) append(addition ...byte) error {
	needed := len(values.data) + len(addition)
	if values.maximum != 0 && needed > values.maximum {
		return errors.New(values.maximumError)
	}
	if needed > cap(values.data) {
		capacity := nextCapacity(cap(values.data), needed)
		if values.maximum != 0 && capacity > values.maximum {
			capacity = values.maximum
		}
		if err := values.budget.charge(capacity); err != nil {
			return err
		}
		grown := make([]byte, len(values.data), capacity)
		copy(grown, values.data)
		values.data = grown
	}
	values.data = append(values.data, addition...)
	return nil
}

type managedValues struct {
	data    []any
	budget  *allocationBudget
	maximum int
}

func (values *managedValues) append(value any) error {
	needed := len(values.data) + 1
	if needed > cap(values.data) {
		capacity := nextCapacity(cap(values.data), needed)
		maximum := values.maximum
		if maximum == 0 {
			maximum = maximumCollectionEntries
		}
		if capacity > maximum {
			capacity = maximum
		}
		if err := values.budget.charge(capacity * int(reflect.TypeFor[any]().Size())); err != nil {
			return err
		}
		grown := make([]any, len(values.data), capacity)
		copy(grown, values.data)
		values.data = grown
	}
	values.data = append(values.data, value)
	return nil
}

type managedMembers struct {
	data    []Member
	budget  *allocationBudget
	maximum int
}

func (members *managedMembers) append(member Member) error {
	needed := len(members.data) + 1
	if needed > cap(members.data) {
		capacity := nextCapacity(cap(members.data), needed)
		maximum := members.maximum
		if maximum == 0 {
			maximum = maximumCollectionEntries
		}
		if capacity > maximum {
			capacity = maximum
		}
		if err := members.budget.charge(capacity * int(reflect.TypeFor[Member]().Size())); err != nil {
			return err
		}
		grown := make([]Member, len(members.data), capacity)
		copy(grown, members.data)
		members.data = grown
	}
	members.data = append(members.data, member)
	return nil
}

func nextCapacity(current, needed int) int {
	capacity := current
	if capacity == 0 {
		capacity = 8
	}
	for capacity < needed {
		capacity *= 2
	}
	return capacity
}

func appendValue(output *managedBytes, value any) error {
	switch value := value.(type) {
	case Object:
		slices.SortFunc(value, func(left, right Member) int { return compareUTF16(left.Name, right.Name) })
		if err := output.append('{'); err != nil {
			return err
		}
		for index, member := range value {
			if index != 0 {
				if err := output.append(','); err != nil {
					return err
				}
			}
			if err := appendJSONString(output, member.Name); err != nil {
				return err
			}
			if err := output.append(':'); err != nil {
				return err
			}
			if err := appendValue(output, member.Value); err != nil {
				return err
			}
		}
		return output.append('}')
	case []any:
		if err := output.append('['); err != nil {
			return err
		}
		for index, element := range value {
			if index != 0 {
				if err := output.append(','); err != nil {
					return err
				}
			}
			if err := appendValue(output, element); err != nil {
				return err
			}
		}
		return output.append(']')
	case string:
		return appendJSONString(output, value)
	case Number:
		return output.append([]byte(value)...)
	case bool:
		if value {
			return output.append([]byte("true")...)
		}
		return output.append([]byte("false")...)
	case nil:
		return output.append([]byte("null")...)
	default:
		return fmt.Errorf("unsupported JSON value %T", value)
	}
}

func compareUTF16(left, right string) int {
	leftIterator, rightIterator := utf16Iterator{text: left}, utf16Iterator{text: right}
	for {
		leftUnit, leftOK := leftIterator.next()
		rightUnit, rightOK := rightIterator.next()
		if !leftOK || !rightOK {
			switch {
			case leftOK:
				return 1
			case rightOK:
				return -1
			default:
				return 0
			}
		}
		if leftUnit < rightUnit {
			return -1
		}
		if leftUnit > rightUnit {
			return 1
		}
	}
}

type utf16Iterator struct {
	text    string
	index   int
	pending uint16
}

func (iterator *utf16Iterator) next() (uint16, bool) {
	if iterator.pending != 0 {
		value := iterator.pending
		iterator.pending = 0
		return value, true
	}
	if iterator.index == len(iterator.text) {
		return 0, false
	}
	value, size := utf8.DecodeRuneInString(iterator.text[iterator.index:])
	iterator.index += size
	if value <= 0xffff {
		return uint16(value), true
	}
	value -= 0x10000
	iterator.pending = uint16(0xdc00 + (value & 0x3ff))
	return uint16(0xd800 + (value >> 10)), true
}

func appendJSONString(output *managedBytes, value string) error {
	if err := output.append('"'); err != nil {
		return err
	}
	const hexadecimal = "0123456789abcdef"
	for _, current := range value {
		switch current {
		case '"':
			if err := output.append('\\', '"'); err != nil {
				return err
			}
		case '\\':
			if err := output.append('\\', '\\'); err != nil {
				return err
			}
		case '\b':
			if err := output.append('\\', 'b'); err != nil {
				return err
			}
		case '\t':
			if err := output.append('\\', 't'); err != nil {
				return err
			}
		case '\n':
			if err := output.append('\\', 'n'); err != nil {
				return err
			}
		case '\f':
			if err := output.append('\\', 'f'); err != nil {
				return err
			}
		case '\r':
			if err := output.append('\\', 'r'); err != nil {
				return err
			}
		default:
			if current < 0x20 {
				if err := output.append('\\', 'u', '0', '0', hexadecimal[byte(current)>>4], hexadecimal[byte(current)&0xf]); err != nil {
					return err
				}
			} else {
				var encoded [utf8.UTFMax]byte
				size := utf8.EncodeRune(encoded[:], current)
				if err := output.append(encoded[:size]...); err != nil {
					return err
				}
			}
		}
	}
	return output.append('"')
}
