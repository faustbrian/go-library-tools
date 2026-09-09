package cohesion

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/bits"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"unicode/utf16"
	"unicode/utf8"
)

const (
	maximumTrustJSONDepth      = 64
	maximumTrustJSONCollection = 4096
	maximumTrustJSONString     = 1 << 20
	trustJSONBudgetOverhead    = 8 << 20
	maximumIJSONSafeInteger    = int64(1<<53 - 1)
)

var (
	trustJSONAnySlotBytes     = int64(reflect.TypeFor[any]().Size())
	trustJSONObjectEntryBytes = int64(reflect.TypeFor[trustJSONObjectEntry]().Size())
)

type surrogateKind uint8

const (
	notSurrogate surrogateKind = iota
	highSurrogate
	lowSurrogate
)

// decodeTrustJSON builds one capacity-accounted decoded representation before
// populating the caller's typed value. Legacy readers intentionally do not use
// this stricter boundary.
func decodeTrustJSON(data []byte, maximumBytes int64, target any) error {
	return decodeTrustJSONWithBudget(data, maximumBytes, target, newTrustJSONBudget(maximumBytes))
}

func decodeTrustJSONWithBudget(data []byte, maximumBytes int64, target any, budget *trustJSONBudget) error {
	value, err := parseTrustJSONWithBudget(data, maximumBytes, budget)
	if err != nil {
		return err
	}
	if err := assignTrustJSON(target, value, budget); err != nil {
		return fmt.Errorf("decode trust JSON: %w", err)
	}
	return nil
}

func parseTrustJSONWithBudget(data []byte, maximumBytes int64, budget *trustJSONBudget) (any, error) {
	return parseTrustJSONWithBudgetAndLimits(data, maximumBytes, budget, trustJSONParserLimits{})
}

type trustJSONParserLimits struct {
	maximumStringBytes func([]trustJSONPathStep) int
	maximumArrayItems  func([]trustJSONPathStep) int
}

func parseTrustJSONWithBudgetAndLimits(data []byte, maximumBytes int64, budget *trustJSONBudget, limits trustJSONParserLimits) (any, error) {
	if maximumBytes < 0 || int64(len(data)) > maximumBytes {
		return nil, errors.New("trust JSON exceeds its maximum byte length")
	}
	if budget == nil {
		return nil, errors.New("trust JSON allocation budget is required")
	}
	if len(data) >= 3 && bytes.Equal(data[:3], []byte{0xef, 0xbb, 0xbf}) {
		return nil, errors.New("trust JSON must not contain a byte-order mark")
	}
	if !utf8.Valid(data) {
		return nil, errors.New("trust JSON must be valid UTF-8")
	}

	scanner := trustJSONScanner{data: data, budget: budget, limits: limits}
	value, err := scanner.value(0)
	if err != nil {
		return nil, err
	}
	scanner.space()
	if scanner.offset != len(data) {
		return nil, scanner.failure("trailing data")
	}
	return value, nil
}

type trustJSONScanner struct {
	data   []byte
	offset int
	budget *trustJSONBudget
	limits trustJSONParserLimits
	path   []trustJSONPathStep
}

type trustJSONPathKind uint8

const (
	trustJSONObjectMember trustJSONPathKind = iota + 1
	trustJSONArrayElement
)

type trustJSONPathStep struct {
	kind trustJSONPathKind
	name string
}

type trustJSONObjectEntry struct {
	key       string
	keyOffset int
	value     any
}

type trustJSONBudget struct {
	charged int64
	limit   int64
	observe func(int64)
}

func newTrustJSONBudget(maximumBytes int64) *trustJSONBudget {
	limit := int64(trustJSONBudgetOverhead) + max(maximumBytes, 0)*2
	return &trustJSONBudget{limit: limit}
}

func (budget *trustJSONBudget) charge(capacity int64) error {
	if budget.observe != nil {
		budget.observe(capacity)
	}
	if capacity < 0 || budget.charged > budget.limit-capacity {
		return errors.New("trust JSON allocation budget exceeded")
	}
	budget.charged += capacity
	return nil
}

func (scanner *trustJSONScanner) value(depth int) (any, error) {
	scanner.space()
	if scanner.offset == len(scanner.data) {
		return nil, scanner.failure("missing value")
	}
	switch scanner.data[scanner.offset] {
	case '{':
		return scanner.object(depth + 1)
	case '[':
		return scanner.array(depth + 1)
	case '"':
		return scanner.stringValue()
	case 't':
		return scanner.literal("true", true)
	case 'f':
		return scanner.literal("false", false)
	case 'n':
		return scanner.literal("null", nil)
	default:
		return scanner.integer()
	}
}

func (scanner *trustJSONScanner) object(depth int) (any, error) {
	if depth > maximumTrustJSONDepth {
		return nil, scanner.failure("nesting depth exceeds 64")
	}
	scanner.offset++
	scanner.space()
	if scanner.take('}') {
		return map[string]any{}, nil
	}
	entries := make([]trustJSONObjectEntry, 0)
	for count := 1; ; count++ {
		if count > maximumTrustJSONCollection {
			return nil, scanner.failure("object has more than 4096 members")
		}
		scanner.space()
		if scanner.offset == len(scanner.data) || scanner.data[scanner.offset] != '"' {
			return nil, scanner.failure("object key is not a string")
		}
		keyOffset := scanner.offset
		key, err := scanner.stringValue()
		if err != nil {
			return nil, err
		}
		scanner.space()
		if !scanner.take(':') {
			return nil, scanner.failure("object key is missing a value separator")
		}
		pathLength := len(scanner.path)
		scanner.path, err = growTrustJSONPathSlice(scanner.path, pathLength+1, scanner.budget)
		if err != nil {
			return nil, err
		}
		scanner.path = append(scanner.path, trustJSONPathStep{kind: trustJSONObjectMember, name: key})
		value, valueErr := scanner.value(depth)
		scanner.path = scanner.path[:pathLength]
		if valueErr != nil {
			return nil, valueErr
		}
		entries, err = growTrustJSONObjectEntries(entries, len(entries)+1, scanner.budget)
		if err != nil {
			return nil, err
		}
		entries = append(entries, trustJSONObjectEntry{key: key, keyOffset: keyOffset, value: value})
		scanner.space()
		if scanner.take('}') {
			if err := scanner.budget.charge(int64(len(entries)) * trustJSONObjectEntryBytes); err != nil {
				return nil, err
			}
			slices.SortFunc(entries, func(left, right trustJSONObjectEntry) int {
				if compared := strings.Compare(left.key, right.key); compared != 0 {
					return compared
				}
				return left.keyOffset - right.keyOffset
			})
			previous := entries[0]
			for _, entry := range entries[1:] {
				if entry.key == previous.key {
					return nil, scanner.failureAt(entry.keyOffset, "object contains a duplicate decoded key")
				}
				previous = entry
			}
			result := make(map[string]any, len(entries))
			for _, entry := range entries {
				result[entry.key] = entry.value
			}
			return result, nil
		}
		if !scanner.take(',') {
			return nil, scanner.failure("object member is missing a separator")
		}
	}
}

func (scanner *trustJSONScanner) array(depth int) (any, error) {
	if depth > maximumTrustJSONDepth {
		return nil, scanner.failure("nesting depth exceeds 64")
	}
	scanner.offset++
	scanner.space()
	if scanner.take(']') {
		return []any{}, nil
	}
	elements := make([]any, 0)
	maximumItems := maximumTrustJSONCollection
	if scanner.limits.maximumArrayItems != nil {
		if configured := scanner.limits.maximumArrayItems(scanner.path); configured > 0 {
			maximumItems = configured
		}
	}
	for count := 1; ; count++ {
		if count > maximumItems {
			return nil, scanner.failure(fmt.Sprintf("array has more than %d elements", maximumItems))
		}
		pathLength := len(scanner.path)
		var err error
		scanner.path, err = growTrustJSONPathSlice(scanner.path, pathLength+1, scanner.budget)
		if err != nil {
			return nil, err
		}
		scanner.path = append(scanner.path, trustJSONPathStep{kind: trustJSONArrayElement})
		value, valueErr := scanner.value(depth)
		scanner.path = scanner.path[:pathLength]
		if valueErr != nil {
			return nil, valueErr
		}
		elements, err = growTrustJSONAnySlice(elements, len(elements)+1, scanner.budget)
		if err != nil {
			return nil, err
		}
		elements = append(elements, value)
		scanner.space()
		if scanner.take(']') {
			return elements, nil
		}
		if !scanner.take(',') {
			return nil, scanner.failure("array element is missing a separator")
		}
	}
}

func (scanner *trustJSONScanner) stringValue() (string, error) {
	probe := *scanner
	decodedLength, err := probe.scanStringLength()
	if err != nil {
		return "", err
	}
	if err := scanner.budget.charge(int64(decodedLength)); err != nil {
		return "", err
	}
	scanner.offset++
	var decoded strings.Builder
	decoded.Grow(decodedLength)
	end := probe.offset - 1
	for scanner.offset < end {
		current := scanner.data[scanner.offset]
		if current == '\\' {
			scanner.writeValidatedEscape(&decoded)
		} else {
			_, width := utf8.DecodeRune(scanner.data[scanner.offset:])
			_, _ = decoded.Write(scanner.data[scanner.offset : scanner.offset+width])
			scanner.offset += width
		}
	}
	scanner.offset = probe.offset
	return decoded.String(), nil
}

func (scanner *trustJSONScanner) scanStringLength() (int, error) {
	if scanner.offset == len(scanner.data) || scanner.data[scanner.offset] != '"' {
		return 0, scanner.failure("string does not start with a quote")
	}
	scanner.offset++
	decodedBytes := 0
	for scanner.offset < len(scanner.data) {
		current := scanner.data[scanner.offset]
		if current == '"' {
			scanner.offset++
			return decodedBytes, nil
		}
		if current < 0x20 {
			return 0, scanner.failure("string contains an unescaped control character")
		}
		var added int
		if current == '\\' {
			width, err := scanner.escapeWidth()
			if err != nil {
				return 0, err
			}
			added = width
		} else {
			runeValue, width := utf8.DecodeRune(scanner.data[scanner.offset:])
			if noncharacter(runeValue) {
				return 0, scanner.failure("string contains a Unicode noncharacter")
			}
			scanner.offset += width
			added = width
		}
		decodedBytes += added
		if decodedBytes > scanner.maximumStringBytes() {
			return 0, scanner.failure("string exceeds its maximum byte length")
		}
	}
	return 0, scanner.failure("unterminated string")
}

func growTrustJSONObjectEntries(values []trustJSONObjectEntry, required int, budget *trustJSONBudget) ([]trustJSONObjectEntry, error) {
	if required <= cap(values) {
		return values, nil
	}
	capacity := nextTrustJSONCapacity(cap(values), required)
	switch err := budget.charge(int64(capacity) * trustJSONObjectEntryBytes); err {
	case nil:
	default:
		return nil, err
	}
	grown := make([]trustJSONObjectEntry, len(values), capacity)
	copy(grown, values)
	return grown, nil
}

func growTrustJSONPathSlice(values []trustJSONPathStep, required int, budget *trustJSONBudget) ([]trustJSONPathStep, error) {
	if required <= cap(values) {
		return values, nil
	}
	capacity := nextTrustJSONCapacity(cap(values), required)
	switch err := budget.charge(int64(capacity) * int64(reflect.TypeFor[trustJSONPathStep]().Size())); err {
	case nil:
	default:
		return nil, err
	}
	grown := make([]trustJSONPathStep, len(values), capacity)
	copy(grown, values)
	return grown, nil
}

func growTrustJSONAnySlice(values []any, required int, budget *trustJSONBudget) ([]any, error) {
	if required <= cap(values) {
		return values, nil
	}
	capacity := nextTrustJSONCapacity(cap(values), required)
	switch err := budget.charge(int64(capacity) * trustJSONAnySlotBytes); err {
	case nil:
	default:
		return nil, err
	}
	grown := make([]any, len(values), capacity)
	copy(grown, values)
	return grown, nil
}

func appendTrustJSONBytes(values []byte, added []byte, maximum int, budget *trustJSONBudget) ([]byte, error) {
	required := len(values) + len(added)
	if required > maximum {
		return nil, errors.New("trust JSON string exceeds its maximum byte length")
	}
	if required > cap(values) {
		capacity := nextTrustJSONCapacity(cap(values), required)
		capacity = min(capacity, maximum)
		if err := budget.charge(int64(capacity)); err != nil {
			return nil, err
		}
		grown := make([]byte, len(values), capacity)
		copy(grown, values)
		values = grown
	}
	return append(values, added...), nil
}

func (scanner *trustJSONScanner) maximumStringBytes() int {
	if scanner.limits.maximumStringBytes != nil {
		if maximum := scanner.limits.maximumStringBytes(scanner.path); maximum > 0 {
			return maximum
		}
	}
	return maximumTrustJSONString
}

func officialDownloadsOracleTrustJSONLimits() trustJSONParserLimits {
	return trustJSONParserLimits{maximumStringBytes: func(path []trustJSONPathStep) int {
		if len(path) == 3 && path[0].kind == trustJSONObjectMember &&
			(path[0].name == "accepted" || path[0].name == "rejected") &&
			path[1].kind == trustJSONArrayElement && path[2] == (trustJSONPathStep{kind: trustJSONObjectMember, name: "input_base64"}) {
			return 44739244
		}
		return maximumTrustJSONString
	}}
}

func forwardOracleTrustJSONLimits() trustJSONParserLimits {
	return trustJSONParserLimits{maximumStringBytes: func(path []trustJSONPathStep) int {
		if len(path) == 3 && path[0] == (trustJSONPathStep{kind: trustJSONObjectMember, name: "fixtures"}) &&
			path[1].kind == trustJSONArrayElement && path[2] == (trustJSONPathStep{kind: trustJSONObjectMember, name: "bytes_base64"}) {
			return 5592408
		}
		return maximumTrustJSONString
	}}
}

func nextTrustJSONCapacity(current, required int) int {
	minimum := 1 << bits.Len(uint(max(required-1, 0)))
	return max(current, minimum)
}

func (scanner *trustJSONScanner) writeValidatedEscape(decoded *strings.Builder) {
	scanner.advanceOne()
	switch scanner.data[scanner.offset] {
	case '"', '\\', '/':
		value := scanner.data[scanner.offset]
		scanner.advanceOne()
		decoded.WriteByte(value)
	case 'b', 'f', 'n', 'r', 't':
		value := map[byte]byte{'b': '\b', 'f': '\f', 'n': '\n', 'r': '\r', 't': '\t'}[scanner.data[scanner.offset]]
		scanner.advanceOne()
		decoded.WriteByte(value)
	case 'u':
		first, _ := scanner.unicodeEscape()
		if surrogateClass(first) == highSurrogate {
			scanner.advanceOne()
			second, _ := scanner.unicodeEscape()
			value := utf16.DecodeRune(rune(first), rune(second))
			decoded.WriteRune(value)
			return
		}
		decoded.WriteRune(rune(first))
	default:
		panic("validated JSON string contains an invalid escape")
	}
}

func (scanner *trustJSONScanner) escapeWidth() (int, error) {
	scanner.offset++
	if scanner.offset == len(scanner.data) {
		return 0, scanner.failure("unterminated escape")
	}
	switch scanner.data[scanner.offset] {
	case '"', '\\', '/', 'b', 'f', 'n', 'r', 't':
		scanner.offset++
		return 1, nil
	case 'u':
		first, err := scanner.unicodeEscape()
		if err != nil {
			return 0, err
		}
		if surrogateClass(first) == highSurrogate {
			if !bytes.HasPrefix(scanner.data[scanner.offset:], []byte(`\u`)) {
				return 0, scanner.failure("string contains a lone high surrogate")
			}
			scanner.advanceOne()
			second, secondErr := scanner.unicodeEscape()
			if secondErr != nil {
				return 0, secondErr
			}
			if surrogateClass(second) != lowSurrogate {
				return 0, scanner.failure("string contains a lone high surrogate")
			}
			value := utf16.DecodeRune(rune(first), rune(second))
			if noncharacter(value) {
				return 0, scanner.failure("string contains a Unicode noncharacter")
			}
			return utf8.RuneLen(value), nil
		}
		if surrogateClass(first) == lowSurrogate {
			return 0, scanner.failure("string contains a lone low surrogate")
		}
		value := rune(first)
		if noncharacter(value) {
			return 0, scanner.failure("string contains a Unicode noncharacter")
		}
		return utf8.RuneLen(value), nil
	default:
		return 0, scanner.failure("string contains an invalid escape")
	}
}

func (scanner *trustJSONScanner) unicodeEscape() (int, error) {
	scanner.offset++
	if scanner.offset+4 > len(scanner.data) {
		return 0, scanner.failure("truncated Unicode escape")
	}
	value := 0
	for range 4 {
		digit := hexadecimal(scanner.data[scanner.offset])
		if digit < 0 {
			return 0, scanner.failure("invalid Unicode escape")
		}
		value = value*16 + digit
		scanner.offset++
	}
	return value, nil
}

func (scanner *trustJSONScanner) integer() (any, error) {
	start := scanner.offset
	negative := scanner.take('-')
	if scanner.offset == len(scanner.data) {
		return nil, scanner.failure("value is not valid JSON")
	}
	if !isASCIIDigit(scanner.data[scanner.offset]) {
		return nil, scanner.failure("value is not valid JSON")
	}
	if scanner.data[scanner.offset] == '0' {
		scanner.offset++
		if negative {
			return nil, scanner.failure("negative zero is not permitted")
		}
		if scanner.offset < len(scanner.data) && isASCIIDigit(scanner.data[scanner.offset]) {
			return nil, scanner.failure("integer has a leading zero")
		}
	} else {
		for scanner.offset < len(scanner.data) && isASCIIDigit(scanner.data[scanner.offset]) {
			scanner.offset++
		}
	}
	if scanner.offset < len(scanner.data) {
		switch scanner.data[scanner.offset] {
		case '.', 'e', 'E':
			return nil, scanner.failure("only integer JSON numbers are permitted")
		}
	}
	length := scanner.offset - start
	if err := scanner.budget.charge(int64(length)); err != nil {
		return nil, err
	}
	text := string(scanner.data[start:scanner.offset])
	value, err := strconv.ParseInt(text, 10, 64)
	if err != nil || value < -maximumIJSONSafeInteger || value > maximumIJSONSafeInteger {
		return nil, scanner.failure("integer is outside the I-JSON interoperable range")
	}
	return json.Number(text), nil
}

func (scanner *trustJSONScanner) advanceOne() {
	remainingAfterCurrent := scanner.data[scanner.offset:][1:]
	scanner.offset = len(scanner.data) - len(remainingAfterCurrent)
}

func (scanner *trustJSONScanner) literal(text string, value any) (any, error) {
	if !bytes.HasPrefix(scanner.data[scanner.offset:], []byte(text)) {
		return nil, scanner.failure("invalid JSON literal")
	}
	scanner.offset += len(text)
	return value, nil
}

func assignTrustJSON(target any, value any, budget *trustJSONBudget) error {
	destination := reflect.ValueOf(target)
	if !destination.IsValid() || destination.Kind() != reflect.Pointer || destination.IsNil() {
		return errors.New("trust JSON decode target must be a non-nil pointer")
	}
	return assignTrustJSONValue(destination.Elem(), value, budget)
}

func assignTrustJSONValue(destination reflect.Value, value any, budget *trustJSONBudget) error {
	if value == nil {
		destination.SetZero()
		return nil
	}
	if destination.Type() == reflect.TypeFor[json.Number]() {
		number, ok := value.(json.Number)
		if !ok {
			return errors.New("trust JSON value is not a number")
		}
		destination.SetString(string(number))
		return nil
	}

	switch destination.Kind() {
	case reflect.Interface:
		decoded := reflect.ValueOf(value)
		if !decoded.Type().AssignableTo(destination.Type()) {
			return fmt.Errorf("trust JSON value %s does not implement destination %s", decoded.Type(), destination.Type())
		}
		destination.Set(decoded)
		return nil
	case reflect.Pointer:
		if err := budget.charge(int64(destination.Type().Elem().Size())); err != nil {
			return err
		}
		destination.Set(reflect.New(destination.Type().Elem()))
		return assignTrustJSONValue(destination.Elem(), value, budget)
	case reflect.Struct:
		object, ok := value.(map[string]any)
		if !ok {
			return errors.New("trust JSON value is not an object")
		}
		for key, item := range object {
			field, exists := trustJSONStructField(destination, key)
			if !exists {
				return fmt.Errorf("unknown field %q", key)
			}
			if err := assignTrustJSONValue(field, item, budget); err != nil {
				return fmt.Errorf("field %s: %w", key, err)
			}
		}
		return nil
	case reflect.Slice:
		items, ok := value.([]any)
		if !ok {
			return errors.New("trust JSON value is not an array")
		}
		if err := budget.charge(int64(len(items)) * int64(destination.Type().Elem().Size())); err != nil {
			return err
		}
		decoded := reflect.MakeSlice(destination.Type(), len(items), len(items))
		for index, item := range items {
			if err := assignTrustJSONValue(decoded.Index(index), item, budget); err != nil {
				return fmt.Errorf("array item %d: %w", index, err)
			}
		}
		destination.Set(decoded)
		return nil
	case reflect.Map:
		object, ok := value.(map[string]any)
		if !ok || destination.Type().Key().Kind() != reflect.String {
			return errors.New("trust JSON value is not a string-keyed object")
		}
		entryBytes := destination.Type().Key().Size() + destination.Type().Elem().Size()
		if err := budget.charge(int64(len(object)) * int64(entryBytes)); err != nil {
			return err
		}
		decoded := reflect.MakeMapWithSize(destination.Type(), len(object))
		for key, item := range object {
			entry := reflect.New(destination.Type().Elem()).Elem()
			if err := assignTrustJSONValue(entry, item, budget); err != nil {
				return fmt.Errorf("object member %s: %w", key, err)
			}
			decoded.SetMapIndex(reflect.ValueOf(key).Convert(destination.Type().Key()), entry)
		}
		destination.Set(decoded)
		return nil
	case reflect.String:
		text, ok := value.(string)
		if !ok {
			return errors.New("trust JSON value is not a string")
		}
		destination.SetString(text)
		return nil
	case reflect.Bool:
		boolean, ok := value.(bool)
		if !ok {
			return errors.New("trust JSON value is not a boolean")
		}
		destination.SetBool(boolean)
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		number, ok := value.(json.Number)
		if !ok {
			return errors.New("trust JSON value is not an integer")
		}
		integer, err := strconv.ParseInt(string(number), 10, destination.Type().Bits())
		if err != nil {
			return errors.New("trust JSON integer does not fit its destination")
		}
		destination.SetInt(integer)
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		number, ok := value.(json.Number)
		if !ok {
			return errors.New("trust JSON value is not an unsigned integer")
		}
		integer, err := strconv.ParseUint(string(number), 10, destination.Type().Bits())
		if err != nil {
			return errors.New("trust JSON unsigned integer does not fit its destination")
		}
		destination.SetUint(integer)
		return nil
	default:
		return fmt.Errorf("unsupported trust JSON decode destination %s", destination.Type())
	}
}

func trustJSONStructField(destination reflect.Value, name string) (reflect.Value, bool) {
	for index := range destination.NumField() {
		fieldType := destination.Type().Field(index)
		if fieldType.IsExported() {
			fieldName := fieldType.Name
			if tag := fieldType.Tag.Get("json"); tag != "" {
				for offset := 0; offset <= len(tag); offset++ {
					if offset == len(tag) || tag[offset] == ',' {
						fieldName = tag[:offset]
						break
					}
				}
			}
			if fieldName == name && fieldName != "-" {
				return destination.Field(index), true
			}
		}
	}
	return reflect.Value{}, false
}

func (scanner *trustJSONScanner) space() {
	for scanner.offset < len(scanner.data) {
		switch scanner.data[scanner.offset] {
		case ' ', '\t', '\r', '\n':
			scanner.offset++
		default:
			return
		}
	}
}

func (scanner *trustJSONScanner) take(value byte) bool {
	if scanner.offset == len(scanner.data) || scanner.data[scanner.offset] != value {
		return false
	}
	scanner.offset++
	return true
}

func (scanner *trustJSONScanner) failure(message string) error {
	return scanner.failureAt(scanner.offset, message)
}

func (scanner *trustJSONScanner) failureAt(offset int, message string) error {
	return fmt.Errorf("trust JSON at byte %d: %s", offset, message)
}

func hexadecimal(value byte) int {
	switch {
	case value >= '0' && value <= '9':
		return int(value - '0')
	case value >= 'a' && value <= 'f':
		return int(value-'a') + 10
	case value >= 'A' && value <= 'F':
		return int(value-'A') + 10
	default:
		return -1
	}
}

func isASCIIDigit(value byte) bool {
	switch value {
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return true
	default:
		return false
	}
}

func surrogateClass(value int) surrogateKind {
	switch value >> 10 {
	case 0x36:
		return highSurrogate
	case 0x37:
		return lowSurrogate
	default:
		return notSurrogate
	}
}

func noncharacter(value rune) bool {
	return value >= 0xfdd0 && value <= 0xfdef || value <= utf8.MaxRune && value&0xffff >= 0xfffe
}
