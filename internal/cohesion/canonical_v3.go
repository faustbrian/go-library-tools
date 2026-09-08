package cohesion

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"unicode/utf8"
)

func canonicalizeTrustJSON(data []byte, maximumBytes int64) ([]byte, error) {
	return canonicalizeTrustJSONWithBudget(data, maximumBytes, newTrustJSONBudget(maximumBytes))
}

func canonicalizeTrustJSONWithBudget(data []byte, maximumBytes int64, budget *trustJSONBudget) ([]byte, error) {
	var value any
	if err := decodeTrustJSONWithBudget(data, maximumBytes, &value, budget); err != nil {
		return nil, err
	}
	return canonicalizeTrustJSONValue(value, len(data), budget)
}

func canonicalizeTrustJSONValue(value any, inputBytes int, budget *trustJSONBudget) ([]byte, error) {
	if err := budget.charge(int64(inputBytes)); err != nil {
		return nil, err
	}
	var output bytes.Buffer
	output.Grow(inputBytes)
	if err := writeCanonicalJSON(&output, value, budget); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}

func canonicalMarshal(value any, maximumBytes int64) ([]byte, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("marshal canonical JSON input: %w", err)
	}
	return canonicalizeTrustJSON(encoded, maximumBytes)
}

func decodeCanonicalTrustJSON(data []byte, maximumBytes int64, target any) error {
	canonical, err := canonicalizeTrustJSON(data, maximumBytes)
	if err != nil {
		return err
	}
	if !bytes.Equal(data, canonical) {
		return errors.New("trust JSON is not RFC 8785 canonical JSON")
	}
	return decodeTrustJSON(data, maximumBytes, target)
}

func exactBytesSHA256(data []byte) string {
	digest := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(digest[:])
}

func semanticObjectSHA256(value any, omittedField string, maximumBytes int64) (string, error) {
	encoded, err := canonicalMarshal(value, maximumBytes)
	if err != nil {
		return "", err
	}
	var object map[string]any
	if err := decodeTrustJSON(encoded, maximumBytes, &object); err != nil {
		return "", err
	}
	if _, exists := object[omittedField]; !exists {
		return "", fmt.Errorf("semantic digest omitted field %q is absent", omittedField)
	}
	delete(object, omittedField)
	preimage, err := canonicalMarshal(object, maximumBytes)
	if err != nil {
		return "", err
	}
	return exactBytesSHA256(preimage), nil
}

func writeCanonicalJSON(output *bytes.Buffer, value any, budget *trustJSONBudget) error {
	switch typed := value.(type) {
	case nil:
		output.WriteString("null")
	case bool:
		if typed {
			output.WriteString("true")
		} else {
			output.WriteString("false")
		}
	case json.Number:
		output.WriteString(typed.String())
	case string:
		writeCanonicalJSONString(output, typed)
	case []any:
		output.WriteByte('[')
		for index, item := range typed {
			if index != 0 {
				output.WriteByte(',')
			}
			if err := writeCanonicalJSON(output, item, budget); err != nil {
				return err
			}
		}
		output.WriteByte(']')
	case map[string]any:
		if err := budget.charge(int64(len(typed)) * int64(reflect.TypeFor[string]().Size())); err != nil {
			return err
		}
		keys := make([]string, 0, len(typed))
		for key := range typed {
			keys = append(keys, key)
		}
		sort.Slice(keys, func(left, right int) bool {
			return compareUTF16(keys[left], keys[right]) == -1
		})
		output.WriteByte('{')
		for index, key := range keys {
			if index != 0 {
				output.WriteByte(',')
			}
			writeCanonicalJSONString(output, key)
			output.WriteByte(':')
			if err := writeCanonicalJSON(output, typed[key], budget); err != nil {
				return err
			}
		}
		output.WriteByte('}')
	default:
		panic(fmt.Sprintf("unsupported decoded canonical JSON value %T", value))
	}
	return nil
}

func writeCanonicalJSONString(output *bytes.Buffer, value string) {
	output.WriteByte('"')
	for _, current := range value {
		switch current {
		case '\b':
			output.WriteString(`\b`)
		case '\t':
			output.WriteString(`\t`)
		case '\n':
			output.WriteString(`\n`)
		case '\f':
			output.WriteString(`\f`)
		case '\r':
			output.WriteString(`\r`)
		case '"':
			output.WriteString(`\"`)
		case '\\':
			output.WriteString(`\\`)
		default:
			if current < 0x20 {
				const hexadecimalDigits = "0123456789abcdef"
				output.WriteString(`\u00`)
				output.WriteByte(hexadecimalDigits[byte(current)>>4])
				output.WriteByte(hexadecimalDigits[byte(current)&0x0f])
			} else {
				output.WriteRune(current)
			}
		}
	}
	output.WriteByte('"')
}

func compareUTF16(left, right string) int {
	leftCursor := utf16Cursor{text: left}
	rightCursor := utf16Cursor{text: right}
	for {
		leftUnit, leftOK := leftCursor.next()
		rightUnit, rightOK := rightCursor.next()
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

type utf16Cursor struct {
	text       string
	offset     int
	low        uint16
	hasLowUnit bool
}

func (cursor *utf16Cursor) next() (uint16, bool) {
	if cursor.hasLowUnit {
		cursor.hasLowUnit = false
		return cursor.low, true
	}
	if cursor.offset == len(cursor.text) {
		return 0, false
	}
	value, width := utf8.DecodeRuneInString(cursor.text[cursor.offset:])
	cursor.offset += width
	if value <= 0xffff {
		return uint16(value), true
	}
	value -= 0x10000
	cursor.low = uint16(0xdc00 + value&0x3ff)
	cursor.hasLowUnit = true
	return uint16(0xd800 + value>>10), true
}
