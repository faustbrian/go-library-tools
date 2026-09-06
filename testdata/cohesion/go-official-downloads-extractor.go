//go:build ignore

// Command go-official-downloads-extractor creates the independently derived
// extraction oracle used to review the schema-v3 official Go download reader.
// It intentionally shares no code, schema bindings, fixtures, or normalization
// helpers with the production reader.
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	goversion "go/version"
	"io"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

const (
	maxIndexBytes  = 32 << 20
	maxStringBytes = 1 << 20
	maxCollection  = 4096
	maxDepth       = 64
)

var (
	digestPattern      = regexp.MustCompile(`^[0-9a-f]{64}$`)
	selected           = map[string]bool{"go1.26.6": true, "go1.27.0": true}
	historicalVersions = map[string]bool{"go1": true, "go1.9.2rc2": true}
)

type codedError struct {
	code string
	err  error
}

func (e *codedError) Error() string { return e.code + ": " + e.err.Error() }
func fail(code, format string, args ...any) error {
	return &codedError{code: code, err: fmt.Errorf(format, args...)}
}
func errorCode(err error) string {
	var coded *codedError
	if errors.As(err, &coded) {
		return coded.code
	}
	return "internal-error"
}

type extracted struct {
	Version            string `json:"version"`
	GOOS               string `json:"goos"`
	GOARCH             string `json:"goarch"`
	OfficialFilename   string `json:"official_filename"`
	DistributionSize   int64  `json:"distribution_size"`
	DistributionSHA256 string `json:"distribution_sha256"`
}

type acceptedCase struct {
	CaseID      string      `json:"case_id"`
	InputBase64 string      `json:"input_base64"`
	Expected    []extracted `json:"expected"`
}

type rejectedCase struct {
	CaseID      string `json:"case_id"`
	InputBase64 string `json:"input_base64"`
	ErrorCode   string `json:"error_code"`
}

type oracle struct {
	SchemaID            string         `json:"schema_id"`
	SchemaVersion       int            `json:"schema_version"`
	OfficialIndexSHA256 string         `json:"official_index_sha256"`
	AcceptedCount       int            `json:"accepted_count"`
	Accepted            []acceptedCase `json:"accepted"`
	RejectedCount       int            `json:"rejected_count"`
	Rejected            []rejectedCase `json:"rejected"`
}

func main() {
	input := flag.String("input", "", "raw go.dev download index")
	output := flag.String("output", "", "canonical oracle destination")
	flag.Parse()
	if *input == "" || *output == "" || flag.NArg() != 0 {
		fatal(errors.New("usage: go run go-official-downloads-extractor.go -input FILE -output FILE"))
	}
	raw, err := readBoundedFile(*input)
	if err != nil {
		fatal(err)
	}
	full, err := extract(raw)
	if err != nil {
		fatal(fmt.Errorf("official snapshot: %w", err))
	}

	validAMD64 := []byte(`[{"version":"go1.26.6","stable":true,"files":[{"filename":"go1.26.6.linux-amd64.tar.gz","os":"linux","arch":"amd64","version":"go1.26.6","sha256":"708effb774be8237570d0add163225abbdfaf4fca28b2611df167beba4feef89","size":66890545,"kind":"archive"}]}]`)
	validARM64 := []byte(`[{"version":"go1.27.0","stable":true,"files":[{"filename":"go1.27.0.linux-arm64.tar.gz","os":"linux","arch":"arm64","version":"go1.27.0","sha256":"51798d2c42d0e1c6ed7fd9f48728b4193abac9e8aad6dbac2fe96a81f5909bda","size":66988980,"kind":"archive"}]}]`)
	validFiltered := []byte(`[{"version":"go1.27.0","stable":false,"files":[{"filename":"go1.27.0.linux-amd64.tar.gz","os":"linux","arch":"amd64","version":"go1.27.0","sha256":"cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc","size":3,"kind":"archive"}]},{"version":"go1.27.1","stable":true,"files":[{"filename":"go1.27.1.linux-amd64.tar.gz","os":"linux","arch":"amd64","version":"go1.27.1","sha256":"dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd","size":4,"kind":"archive"}]}]`)
	validHistoricalAliases := []byte(`[{"version":"go1.4.2","stable":true,"files":[{"filename":"go1.4.2.darwin-amd64-osx10.6.tar.gz","os":"darwin","arch":"amd64","version":"go1.4.2","sha256":"eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee","size":5,"kind":"archive"},{"filename":"go1.4.2.darwin-amd64-osx10.8.tar.gz","os":"darwin","arch":"amd64","version":"go1.4.2","sha256":"eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee","size":5,"kind":"archive"}]}]`)
	validHistoricalEmptyHash := []byte(`[{"version":"go1.5.2","stable":true,"files":[{"filename":"go1.5.2.linux-amd64.tar.gz","os":"linux","arch":"amd64","version":"go1.5.2","sha256":"","size":1,"kind":"archive"}]}]`)
	validHistoricalZeroSize := []byte(`[{"version":"go1.4","stable":true,"files":[{"filename":"go1.4.src.tar.gz","os":"","arch":"","version":"go1.4","sha256":"","size":0,"kind":"source"}]}]`)
	validHistoricalGo1 := []byte(`[{"version":"go1","stable":true,"files":[{"filename":"go1.4-bootstrap-20170531.tar.gz","os":"","arch":"bootstrap","version":"go1","sha256":"5af68ba78661528835c5125e66b2ee1fc8df41a6aebdc849020ce4596e175910","size":50284377,"kind":"archive"}]}]`)
	validHistoricalPatchRC := []byte(`[{"version":"go1.9.2rc2","stable":true,"files":[{"filename":"go1.9.2rc2.linux-amd64.tar.gz","os":"linux","arch":"amd64","version":"go1.9.2rc2","sha256":"ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff","size":6,"kind":"archive"}]}]`)
	validLiteralReplacement := []byte(`[{"version":"go1.27.1","stable":true,"files":[{"filename":"go1.27.1-�.src.tar.gz","os":"","arch":"","version":"go1.27.1","sha256":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","size":7,"kind":"source"}]}]`)
	validEscapedReplacement := []byte(`[{"version":"go1.27.1","stable":true,"files":[{"filename":"go1.27.1-\ufffd.src.tar.gz","os":"","arch":"","version":"go1.27.1","sha256":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","size":7,"kind":"source"}]}]`)
	acceptedInputs := map[string][]byte{
		"full-snapshot":                      raw,
		"noncandidate-historical-aliases":    validHistoricalAliases,
		"noncandidate-historical-empty-hash": validHistoricalEmptyHash,
		"noncandidate-historical-go1":        validHistoricalGo1,
		"noncandidate-historical-patch-rc":   validHistoricalPatchRC,
		"noncandidate-historical-zero-size":  validHistoricalZeroSize,
		"noncandidate-literal-replacement":   validLiteralReplacement,
		"noncandidate-escaped-replacement":   validEscapedReplacement,
		"selected-linux-amd64":               validAMD64,
		"selected-linux-arm64":               validARM64,
		"unstable-or-unselected":             validFiltered,
	}

	rejectedInputs := map[string][]byte{
		"bom":                                   append([]byte{0xef, 0xbb, 0xbf}, validAMD64...),
		"invalid-json":                          []byte(`[`),
		"invalid-utf8":                          []byte{'[', '"', 0xff, '"', ']'},
		"trailing-value":                        append(append([]byte{}, validAMD64...), []byte(` []`)...),
		"duplicate-release-key":                 []byte(`[{"version":"go1.26.6","version":"go1.26.6","stable":true,"files":[]}]`),
		"duplicate-file-key":                    []byte(`[{"version":"go1.26.6","stable":true,"files":[{"filename":"a","filename":"b","os":"linux","arch":"amd64","version":"go1.26.6","sha256":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","size":1,"kind":"archive"}]}]`),
		"unknown-release-field":                 []byte(`[{"version":"go1.26.6","stable":true,"files":[],"extra":false}]`),
		"cross-shape-member":                    []byte(`[{"version":"go1.26.6","stable":true,"files":[{"filename":"a","os":"linux","arch":"amd64","version":"go1.26.6","sha256":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","size":1,"kind":"archive","stable":true}]}]`),
		"release-version-duplicate":             []byte(`[{"version":"go1.26.6","stable":true,"files":[]},{"version":"go1.26.6","stable":false,"files":[]}]`),
		"release-version-invalid":               []byte(`[{"version":"1.26.6","stable":true,"files":[]}]`),
		"historical-version-near-miss-go01":     []byte(`[{"version":"go01","stable":true,"files":[]}]`),
		"historical-version-near-miss-patch-rc": []byte(`[{"version":"go1.9.2rc3","stable":true,"files":[]}]`),
		"file-version-mismatch":                 replace(validAMD64, `"version":"go1.26.6","sha256"`, `"version":"go1.27.0","sha256"`),
		"filename-duplicate":                    []byte(`[{"version":"go1.26.6","stable":true,"files":[{"filename":"same","os":"linux","arch":"amd64","version":"go1.26.6","sha256":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","size":1,"kind":"archive"},{"filename":"same","os":"linux","arch":"arm64","version":"go1.26.6","sha256":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb","size":2,"kind":"archive"}]}]`),
		"duplicate-registry-candidate":          []byte(`[{"version":"go1.26.6","stable":true,"files":[{"filename":"a","os":"linux","arch":"amd64","version":"go1.26.6","sha256":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","size":1,"kind":"archive"},{"filename":"b","os":"linux","arch":"amd64","version":"go1.26.6","sha256":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb","size":2,"kind":"archive"}]}]`),
		"file-kind-invalid":                     replace(validAMD64, `"kind":"archive"`, `"kind":"binary"`),
		"candidate-empty-hash":                  replace(validAMD64, "708effb774be8237570d0add163225abbdfaf4fca28b2611df167beba4feef89", ""),
		"candidate-zero-size":                   replace(validAMD64, `"size":66890545`, `"size":0`),
		"file-size-fraction":                    replace(validAMD64, `"size":66890545`, `"size":1.5`),
		"file-size-negative-zero":               replace(validAMD64, `"size":66890545`, `"size":-0`),
		"file-size-overflow":                    replace(validAMD64, `"size":66890545`, `"size":9223372036854775808`),
		"file-sha256-uppercase":                 replace(validAMD64, "708effb774be8237570d0add163225abbdfaf4fca28b2611df167beba4feef89", strings.ToUpper("708effb774be8237570d0add163225abbdfaf4fca28b2611df167beba4feef89")),
		"file-sha256-short":                     replace(validAMD64, "708effb774be8237570d0add163225abbdfaf4fca28b2611df167beba4feef89", "aa"),
		"escaped-lone-surrogate":                replace(validAMD64, `"filename":"go1.26.6.linux-amd64.tar.gz"`, `"filename":"\ud800"`),
		"escaped-noncharacter":                  replace(validAMD64, `"filename":"go1.26.6.linux-amd64.tar.gz"`, `"filename":"\uffff"`),
	}

	result := oracle{
		SchemaID:            "urn:golib:cohesion:go-official-downloads-oracle:v1",
		SchemaVersion:       1,
		OfficialIndexSHA256: digest(raw),
	}
	for id, input := range acceptedInputs {
		rows, extractErr := extract(input)
		if extractErr != nil {
			fatal(fmt.Errorf("accepted case %s: %w", id, extractErr))
		}
		if id == "full-snapshot" && !equalRows(rows, full) {
			fatal(errors.New("full snapshot result changed during generation"))
		}
		result.Accepted = append(result.Accepted, acceptedCase{CaseID: id, InputBase64: base64.StdEncoding.EncodeToString(input), Expected: rows})
	}
	for id, input := range rejectedInputs {
		if _, extractErr := extract(input); extractErr == nil {
			fatal(fmt.Errorf("rejected case %s was accepted", id))
		} else {
			result.Rejected = append(result.Rejected, rejectedCase{CaseID: id, InputBase64: base64.StdEncoding.EncodeToString(input), ErrorCode: errorCode(extractErr)})
		}
	}
	sort.Slice(result.Accepted, func(i, j int) bool { return result.Accepted[i].CaseID < result.Accepted[j].CaseID })
	sort.Slice(result.Rejected, func(i, j int) bool { return result.Rejected[i].CaseID < result.Rejected[j].CaseID })
	result.AcceptedCount = len(result.Accepted)
	result.RejectedCount = len(result.Rejected)
	encoded, err := canonicalMarshal(result)
	if err != nil {
		fatal(err)
	}
	if err := os.WriteFile(*output, encoded, 0o644); err != nil {
		fatal(err)
	}
}

func readBoundedFile(path string) ([]byte, error) {
	source, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	raw, readErr := io.ReadAll(io.LimitReader(source, maxIndexBytes+1))
	closeErr := source.Close()
	if readErr != nil {
		return nil, readErr
	}
	if closeErr != nil {
		return nil, closeErr
	}
	if len(raw) > maxIndexBytes {
		return nil, fail("input-size-limit", "input exceeds %d bytes", maxIndexBytes)
	}
	return raw, nil
}

func replace(input []byte, old, replacement string) []byte {
	return bytes.ReplaceAll(input, []byte(old), []byte(replacement))
}

func digest(input []byte) string {
	sum := sha256.Sum256(input)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func equalRows(left, right []extracted) bool {
	a, _ := json.Marshal(left)
	b, _ := json.Marshal(right)
	return bytes.Equal(a, b)
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

func extract(input []byte) ([]extracted, error) {
	value, err := strictJSON(input)
	if err != nil {
		return nil, err
	}
	releases, ok := value.([]any)
	if !ok {
		return nil, fail("top-level-type", "expected array")
	}
	if len(releases) > maxCollection {
		return nil, fail("release-count-limit", "release count exceeds %d", maxCollection)
	}
	versions := make(map[string]bool, len(releases))
	filenames := make(map[string]bool)
	rows := make([]extracted, 0)
	for releaseIndex, rawRelease := range releases {
		release, ok := rawRelease.(map[string]any)
		if !ok {
			return nil, fail("release-type", "release %d is not an object", releaseIndex)
		}
		if err := exactKeys(release, "version", "stable", "files"); err != nil {
			return nil, err
		}
		version, ok := release["version"].(string)
		if !ok || !historicalVersions[version] && !goversion.IsValid(version) {
			return nil, fail("release-version-invalid", "release %d has invalid version", releaseIndex)
		}
		if versions[version] {
			return nil, fail("release-version-duplicate", "duplicate version %s", version)
		}
		versions[version] = true
		stable, ok := release["stable"].(bool)
		if !ok {
			return nil, fail("release-stable-invalid", "release %s has non-boolean stable", version)
		}
		files, ok := release["files"].([]any)
		if !ok {
			return nil, fail("files-type", "release %s files is not an array", version)
		}
		if len(files) > maxCollection {
			return nil, fail("file-count-limit", "release %s file count exceeds %d", version, maxCollection)
		}
		candidateTuples := make(map[string]bool, len(files))
		for fileIndex, rawFile := range files {
			file, ok := rawFile.(map[string]any)
			if !ok {
				return nil, fail("file-type", "release %s file %d is not an object", version, fileIndex)
			}
			if err := exactKeys(file, "filename", "os", "arch", "version", "sha256", "size", "kind"); err != nil {
				return nil, err
			}
			filename, filenameOK := file["filename"].(string)
			goos, osOK := file["os"].(string)
			goarch, archOK := file["arch"].(string)
			fileVersion, versionOK := file["version"].(string)
			checksum, checksumOK := file["sha256"].(string)
			kind, kindOK := file["kind"].(string)
			sizeNumber, sizeOK := file["size"].(json.Number)
			if !filenameOK || filename == "" || !osOK || !archOK || !versionOK || !checksumOK || !kindOK || !sizeOK {
				return nil, fail("file-field-invalid", "release %s file %d has invalid field type", version, fileIndex)
			}
			if fileVersion != version {
				return nil, fail("file-version-mismatch", "file %s version differs from release", filename)
			}
			if filenames[filename] {
				return nil, fail("filename-duplicate", "duplicate filename %s", filename)
			}
			filenames[filename] = true
			if kind != "archive" && kind != "installer" && kind != "source" {
				return nil, fail("file-kind-invalid", "file %s has invalid kind", filename)
			}
			size, parseErr := strconv.ParseInt(string(sizeNumber), 10, 64)
			if parseErr != nil || size < 0 {
				return nil, fail("file-size-invalid", "file %s has invalid size", filename)
			}
			if checksum != "" && !digestPattern.MatchString(checksum) {
				return nil, fail("file-sha256-invalid", "file %s has invalid sha256", filename)
			}
			candidate := stable && selected[version] && kind == "archive" && goos == "linux" && (goarch == "amd64" || goarch == "arm64")
			if candidate {
				if size == 0 {
					return nil, fail("candidate-size-invalid", "file %s has zero size", filename)
				}
				if !digestPattern.MatchString(checksum) {
					return nil, fail("candidate-sha256-invalid", "file %s has no usable sha256", filename)
				}
				tuple := version + "\x00" + goos + "\x00" + goarch + "\x00" + kind
				if candidateTuples[tuple] {
					return nil, fail("duplicate-registry-candidate", "release %s has duplicate selectable os/arch/kind", version)
				}
				candidateTuples[tuple] = true
				rows = append(rows, extracted{Version: version, GOOS: goos, GOARCH: goarch, OfficialFilename: filename, DistributionSize: size, DistributionSHA256: "sha256:" + checksum})
			}
		}
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Version != rows[j].Version {
			return rows[i].Version < rows[j].Version
		}
		if rows[i].GOOS != rows[j].GOOS {
			return rows[i].GOOS < rows[j].GOOS
		}
		if rows[i].GOARCH != rows[j].GOARCH {
			return rows[i].GOARCH < rows[j].GOARCH
		}
		return rows[i].OfficialFilename < rows[j].OfficialFilename
	})
	return rows, nil
}

func exactKeys(object map[string]any, expected ...string) error {
	if len(object) != len(expected) {
		return fail("unknown-or-missing-field", "object has wrong member count")
	}
	for _, key := range expected {
		if _, ok := object[key]; !ok {
			return fail("unknown-or-missing-field", "object is missing %s", key)
		}
	}
	return nil
}

func strictJSON(input []byte) (any, error) {
	if len(input) > maxIndexBytes {
		return nil, fail("input-size-limit", "input exceeds %d bytes", maxIndexBytes)
	}
	if bytes.HasPrefix(input, []byte{0xef, 0xbb, 0xbf}) {
		return nil, fail("bom", "byte-order mark is forbidden")
	}
	if !utf8.Valid(input) {
		return nil, fail("invalid-utf8", "input is not UTF-8")
	}
	if err := rejectInvalidSurrogateEscapes(input); err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(bytes.NewReader(input))
	decoder.UseNumber()
	value, err := parseValue(decoder, 1)
	if err != nil {
		return nil, err
	}
	if _, err := decoder.Token(); err != io.EOF {
		if err == nil {
			return nil, fail("trailing-value", "input contains a second value")
		}
		return nil, fail("invalid-json", "%v", err)
	}
	return value, nil
}

func parseValue(decoder *json.Decoder, depth int) (any, error) {
	if depth > maxDepth {
		return nil, fail("depth-limit", "nesting exceeds %d", maxDepth)
	}
	token, err := decoder.Token()
	if err != nil {
		return nil, fail("invalid-json", "%v", err)
	}
	switch value := token.(type) {
	case json.Delim:
		switch value {
		case '[':
			items := make([]any, 0)
			for decoder.More() {
				if len(items) == maxCollection {
					return nil, fail("collection-limit", "array exceeds %d elements", maxCollection)
				}
				item, itemErr := parseValue(decoder, depth+1)
				if itemErr != nil {
					return nil, itemErr
				}
				items = append(items, item)
			}
			if end, endErr := decoder.Token(); endErr != nil || end != json.Delim(']') {
				return nil, fail("invalid-json", "unterminated array")
			}
			return items, nil
		case '{':
			object := make(map[string]any)
			for decoder.More() {
				if len(object) == maxCollection {
					return nil, fail("collection-limit", "object exceeds %d members", maxCollection)
				}
				keyToken, keyErr := decoder.Token()
				if keyErr != nil {
					return nil, fail("invalid-json", "%v", keyErr)
				}
				key, ok := keyToken.(string)
				if !ok {
					return nil, fail("invalid-json", "object key is not a string")
				}
				if err := validString(key); err != nil {
					return nil, err
				}
				if _, exists := object[key]; exists {
					return nil, fail("duplicate-key", "duplicate object key %q", key)
				}
				item, itemErr := parseValue(decoder, depth+1)
				if itemErr != nil {
					return nil, itemErr
				}
				object[key] = item
			}
			if end, endErr := decoder.Token(); endErr != nil || end != json.Delim('}') {
				return nil, fail("invalid-json", "unterminated object")
			}
			return object, nil
		default:
			return nil, fail("invalid-json", "unexpected delimiter")
		}
	case string:
		if err := validString(value); err != nil {
			return nil, err
		}
		return value, nil
	case json.Number:
		number := string(value)
		if number == "-0" || strings.ContainsAny(number, ".eE") {
			return nil, fail("non-integer-number", "number %s is not a permitted integer", number)
		}
		if _, err := strconv.ParseInt(number, 10, 64); err != nil {
			return nil, fail("integer-range", "number is outside int64")
		}
		return value, nil
	default:
		return value, nil
	}
}

// canonicalMarshal implements the RFC 8785 rules needed by this oracle's
// deliberately ASCII-only, integer-only output domain. Keeping it here avoids
// sharing canonicalization code with the later production implementation.
func canonicalMarshal(value any) ([]byte, error) {
	ordinary, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(bytes.NewReader(ordinary))
	decoder.UseNumber()
	var decoded any
	if err := decoder.Decode(&decoded); err != nil {
		return nil, err
	}
	var output bytes.Buffer
	if err := writeCanonical(&output, decoded); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}

func writeCanonical(output *bytes.Buffer, value any) error {
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
		output.WriteString(string(typed))
	case string:
		for _, r := range typed {
			if r > 0x7f {
				return fmt.Errorf("canonical oracle output contains non-ASCII string")
			}
		}
		encoded, err := json.Marshal(typed)
		if err != nil {
			return err
		}
		output.Write(encoded)
	case []any:
		output.WriteByte('[')
		for index, item := range typed {
			if index > 0 {
				output.WriteByte(',')
			}
			if err := writeCanonical(output, item); err != nil {
				return err
			}
		}
		output.WriteByte(']')
	case map[string]any:
		keys := make([]string, 0, len(typed))
		for key := range typed {
			for _, r := range key {
				if r > 0x7f {
					return fmt.Errorf("canonical oracle output contains non-ASCII member name")
				}
			}
			keys = append(keys, key)
		}
		sort.Strings(keys)
		output.WriteByte('{')
		for index, key := range keys {
			if index > 0 {
				output.WriteByte(',')
			}
			encodedKey, err := json.Marshal(key)
			if err != nil {
				return err
			}
			output.Write(encodedKey)
			output.WriteByte(':')
			if err := writeCanonical(output, typed[key]); err != nil {
				return err
			}
		}
		output.WriteByte('}')
	default:
		return fmt.Errorf("unsupported canonical oracle value %T", value)
	}
	return nil
}

func rejectInvalidSurrogateEscapes(input []byte) error {
	inString := false
	for index := 0; index < len(input); index++ {
		switch input[index] {
		case '"':
			inString = !inString
		case '\\':
			if !inString || index+1 >= len(input) {
				continue
			}
			index++
			if input[index] != 'u' || index+4 >= len(input) {
				continue
			}
			codePoint, ok := decodeHex4(input[index+1 : index+5])
			if !ok {
				continue
			}
			if codePoint >= 0xdc00 && codePoint <= 0xdfff {
				return fail("invalid-unicode", "string contains lone low surrogate")
			}
			if codePoint >= 0xd800 && codePoint <= 0xdbff {
				if index+10 >= len(input) || input[index+5] != '\\' || input[index+6] != 'u' {
					return fail("invalid-unicode", "string contains lone high surrogate")
				}
				low, lowOK := decodeHex4(input[index+7 : index+11])
				if !lowOK || low < 0xdc00 || low > 0xdfff {
					return fail("invalid-unicode", "string contains invalid surrogate pair")
				}
				index += 10
				continue
			}
			index += 4
		}
	}
	return nil
}

func decodeHex4(input []byte) (rune, bool) {
	if len(input) != 4 {
		return 0, false
	}
	var value rune
	for _, current := range input {
		value <<= 4
		switch {
		case current >= '0' && current <= '9':
			value += rune(current - '0')
		case current >= 'a' && current <= 'f':
			value += rune(current-'a') + 10
		case current >= 'A' && current <= 'F':
			value += rune(current-'A') + 10
		default:
			return 0, false
		}
	}
	return value, true
}

func validString(value string) error {
	if len(value) > maxStringBytes {
		return fail("string-limit", "string exceeds %d bytes", maxStringBytes)
	}
	for _, r := range value {
		if r == 0xfffe || r == 0xffff || r >= 0xfdd0 && r <= 0xfdef || r&0xffff == 0xfffe || r&0xffff == 0xffff {
			return fail("invalid-unicode", "string contains forbidden code point")
		}
	}
	return nil
}
