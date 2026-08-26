package mutation

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestReadBootstrapRejectsInvalidArchiveStructure(t *testing.T) {
	t.Run("invalid zip", func(t *testing.T) {
		_, err := ReadBootstrap(bytes.NewReader([]byte("zip")), 3)
		assertInvalidContains(t, err, "open archive")
	})
	t.Run("empty zip", func(t *testing.T) {
		archive := archiveWithHeaders(t, nil)
		_, err := ReadBootstrap(bytes.NewReader(archive), int64(len(archive)))
		assertInvalidContains(t, err, "entry count")
	})
	t.Run("too many entries", func(t *testing.T) {
		headers := make([]zip.FileHeader, maximumCheckpointCount+1)
		for index := range headers {
			headers[index].Name = "mutation-checkpoints/" + strings.Repeat("x", index%10) + string(rune('a'+index%26)) + ".json"
		}
		archive := archiveWithHeaders(t, headers)
		_, err := ReadBootstrap(bytes.NewReader(archive), int64(len(archive)))
		assertInvalidContains(t, err, "entry count")
	})
}

func TestValidateArchiveEntryRejectsInvalidMetadata(t *testing.T) {
	tests := map[string]struct {
		header zip.FileHeader
		want   string
	}{
		"non canonical":   {header: zip.FileHeader{Name: "mutation-checkpoints/a/../b.json"}, want: "unsafe archive path"},
		"absolute":        {header: zip.FileHeader{Name: "/mutation-checkpoints/a.json"}, want: "unsafe archive path"},
		"current":         {header: zip.FileHeader{Name: "."}, want: "unsafe archive path"},
		"directory":       {header: directoryHeader("mutation-checkpoints/a.json"), want: "not a regular file"},
		"symlink":         {header: symlinkHeader("mutation-checkpoints/a.json"), want: "not a regular file"},
		"nested":          {header: zip.FileHeader{Name: "mutation-checkpoints/a/b.json"}, want: "unexpected archive entry"},
		"wrong suffix":    {header: zip.FileHeader{Name: "mutation-checkpoints/a.txt"}, want: "unexpected archive entry"},
		"empty":           {header: zip.FileHeader{Name: "mutation-checkpoints/a.json"}, want: "checkpoint size"},
		"oversized":       {header: sizedHeader(maximumCheckpointSize+1, 10), want: "checkpoint size"},
		"zero compressed": {header: sizedHeader(1, 0), want: "compression ratio"},
		"ratio":           {header: sizedHeader(201, 1), want: "compression ratio"},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			err := validateArchiveEntry(&zip.File{FileHeader: test.header})
			assertInvalidContains(t, err, test.want)
		})
	}
}

func TestReadBootstrapRejectsDuplicateArchiveEntry(t *testing.T) {
	var output bytes.Buffer
	writer := zip.NewWriter(&output)
	for range 2 {
		entry, err := writer.Create("mutation-checkpoints/root.json")
		if err != nil {
			t.Fatal(err)
		}
		_, _ = io.WriteString(entry, validCheckpointJSON())
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	_, err := ReadBootstrap(bytes.NewReader(output.Bytes()), int64(output.Len()))
	assertInvalidContains(t, err, "duplicate archive entry")
}

func TestExpandedArchiveAccountingIsBounded(t *testing.T) {
	if total, err := addExpanded(1, 2); err != nil || total != 3 {
		t.Fatalf("addExpanded() = %d, %v", total, err)
	}
	for _, values := range [][2]uint64{{maximumExpandedSize, 1}, {0, maximumExpandedSize + 1}} {
		if _, err := addExpanded(values[0], values[1]); !errors.Is(err, ErrInvalid) {
			t.Fatalf("addExpanded(%d, %d) error = %v", values[0], values[1], err)
		}
	}
}

func TestReadBootstrapRejectsExcessiveExpandedArchive(t *testing.T) {
	var output bytes.Buffer
	writer := zip.NewWriter(&output)
	body := strings.Repeat("x", maximumCheckpointSize)
	for index := 0; index <= maximumExpandedSize/maximumCheckpointSize; index++ {
		entry, err := writer.Create("mutation-checkpoints/" + strings.Repeat("x", index/26) + string(rune('a'+index%26)) + ".json")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := io.WriteString(entry, body); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	_, err := ReadBootstrap(bytes.NewReader(output.Bytes()), int64(output.Len()))
	assertInvalidContains(t, err, "expanded archive exceeds")
}

func TestReadCheckpointRejectsOpenReadAndSizeFailures(t *testing.T) {
	openError := errors.New("open failed")
	_, err := readCheckpoint(1, func() (io.ReadCloser, error) { return nil, openError })
	assertInvalidContains(t, err, "open checkpoint")

	_, err = readCheckpoint(1, func() (io.ReadCloser, error) {
		return io.NopCloser(failingReader{}), nil
	})
	assertInvalidContains(t, err, "read checkpoint")

	_, err = readCheckpoint(2, func() (io.ReadCloser, error) {
		return io.NopCloser(strings.NewReader("x")), nil
	})
	assertInvalidContains(t, err, "size mismatch")
}

func TestValidateReportRejectsReadAndSizeFailures(t *testing.T) {
	if _, err := ValidateReport(failingReader{}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("ValidateReport(read failure) error = %v", err)
	}
	if _, err := ValidateReport(strings.NewReader(strings.Repeat("x", maximumCheckpointSize+1))); !errors.Is(err, ErrInvalid) {
		t.Fatalf("ValidateReport(oversized) error = %v", err)
	}
}

func TestParseCheckpointRejectsMalformedContracts(t *testing.T) {
	valid := checkpointMap(t)
	tests := map[string]struct {
		mutate func(map[string]any)
		want   string
	}{
		"unknown checkpoint field": {mutate: func(value map[string]any) { value["unknown"] = true }, want: "unknown field"},
		"wrong schema":             {mutate: func(value map[string]any) { value["schema_version"] = 2 }, want: "schema_version"},
		"unsafe module":            {mutate: func(value map[string]any) { value["module"] = "../outside" }, want: "safe relative paths"},
		"empty package":            {mutate: func(value map[string]any) { value["package"] = "" }, want: "safe relative paths"},
		"bad input digest":         {mutate: func(value map[string]any) { value["gate_input_digest"] = "bad" }, want: "identity is malformed"},
		"bad version":              {mutate: func(value map[string]any) { value["gremlins_version"] = "latest" }, want: "identity is malformed"},
		"bad verifier digest":      {mutate: func(value map[string]any) { value["gremlins_verifier_sha256"] = "bad" }, want: "gremlins_verifier_sha256"},
		"bad binary digest":        {mutate: func(value map[string]any) { value["gremlins_binary_sha256"] = "bad" }, want: "gremlins_binary_sha256"},
		"bad source":               {mutate: func(value map[string]any) { value["verifier_identity_source"] = "claimed" }, want: "verifier_identity_source"},
		"missing environment":      {mutate: func(value map[string]any) { value["environment"] = map[string]string{} }, want: "environment is required"},
		"unknown report field":     {mutate: func(value map[string]any) { reportOf(value)["unknown"] = true }, want: "unknown field"},
		"missing file name":        {mutate: func(value map[string]any) { fileOf(value)["file_name"] = "" }, want: "file name is required"},
		"missing mutation type":    {mutate: func(value map[string]any) { mutationOf(value)["type"] = "" }, want: "location is malformed"},
		"invalid line":             {mutate: func(value map[string]any) { mutationOf(value)["line"] = 0 }, want: "location is malformed"},
		"invalid column":           {mutate: func(value map[string]any) { mutationOf(value)["column"] = 0 }, want: "location is malformed"},
		"survivor":                 {mutate: func(value map[string]any) { mutationOf(value)["status"] = "SURVIVED" }, want: "non-killed mutant"},
		"duplicate file": {mutate: func(value map[string]any) {
			files := reportOf(value)["files"].([]any)
			reportOf(value)["files"] = append(files, files[0])
		}, want: "duplicate mutation file"},
		"duplicate mutation": {mutate: func(value map[string]any) {
			mutations := fileOf(value)["mutations"].([]any)
			fileOf(value)["mutations"] = append(mutations, mutations[0])
		}, want: "duplicate mutation identity"},
		"incomplete counters": {mutate: func(value map[string]any) { reportOf(value)["mutants_total"] = 2 }, want: "incomplete aggregate counters"},
		"inconsistent counters": {mutate: func(value map[string]any) {
			report := reportOf(value)
			report["mutants_total"] = 1
			report["mutants_killed"] = 0
			report["mutants_lived"] = 1
			report["mutants_not_covered"] = 0
			report["mutants_not_viable"] = 0
			report["mutations_coverage"] = 0
			report["test_efficacy"] = 0
		}, want: "aggregate counters do not prove"},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			value := cloneMap(t, valid)
			test.mutate(value)
			_, err := parseCheckpoint(marshalMap(t, value))
			assertInvalidContains(t, err, test.want)
		})
	}
}

func TestParseCheckpointAcceptsVerifierVariantsAndEmptyReports(t *testing.T) {
	for _, source := range []string{"", "executed", "approved-semantic-migration"} {
		value := checkpointMap(t)
		value["verifier_identity_source"] = source
		if source != "" {
			value["gremlins_verifier_sha256"] = strings.Repeat("b", 64)
			value["gremlins_binary_sha256"] = strings.Repeat("c", 64)
		}
		reportOf(value)["files"] = []any{}
		checkpoint, err := parseCheckpoint(marshalMap(t, value))
		if err != nil {
			t.Fatalf("parseCheckpoint(%q) error = %v", source, err)
		}
		if checkpoint.Mutants != 0 || len(checkpoint.Report) == 0 || len(checkpoint.Environment) == 0 {
			t.Fatalf("parseCheckpoint(%q) = %#v", source, checkpoint)
		}
	}
}

func TestDecodeStrictRejectsMultipleValuesAndTrailingSyntax(t *testing.T) {
	for _, data := range []string{
		validCheckpointJSON() + `{}`,
		validCheckpointJSON() + `{`,
		strings.Replace(validCheckpointJSON(), `"schema_version":3`, `"schema_version":3,"schema_version":3`, 1),
	} {
		_, err := parseCheckpoint([]byte(data))
		if !errors.Is(err, ErrInvalid) {
			t.Fatalf("parseCheckpoint() error = %v, want ErrInvalid", err)
		}
	}
}

func TestDuplicateKeyScannerRejectsMalformedStructures(t *testing.T) {
	for _, data := range []string{"", `{invalid`, `{"a":}`, `[}]`, `[{"a":}]`} {
		if err := rejectDuplicateKeys([]byte(data)); !errors.Is(err, ErrInvalid) {
			t.Fatalf("rejectDuplicateKeys(%q) error = %v", data, err)
		}
	}
}

func TestCanonicalReportDigestIgnoresJSONFormatting(t *testing.T) {
	left := canonicalReportDigest([]byte(`{"files":[],"go_module":"example"}`))
	right := canonicalReportDigest([]byte("{\n  \"go_module\": \"example\",\n  \"files\": []\n}"))
	if left != right || left != "sha256:2be4cb6094805afbabf362551ed21296406ed8695d68310b690c1a9495aea4a1" {
		t.Fatalf("canonicalReportDigest() = %q, %q", left, right)
	}
}

func TestValidRelative(t *testing.T) {
	for value, want := range map[string]bool{".": true, "nested/package": true, "": false, "/absolute": false, "../outside": false, "a/../b": false, `a\b`: false, "a\x00b": false} {
		if got := validRelative(value); got != want {
			t.Errorf("validRelative(%q) = %t, want %t", value, got, want)
		}
	}
}

func TestParseZeroInventoryRejectsReadSizeAndDuplicateFailures(t *testing.T) {
	if _, err := ParseZeroInventory(failingReader{}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("ParseZeroInventory(read failure) error = %v", err)
	}
	if _, err := ParseZeroInventory(strings.NewReader(strings.Repeat("x", maximumZeroInventorySize+1))); !errors.Is(err, ErrInvalid) {
		t.Fatalf("ParseZeroInventory(oversized) error = %v", err)
	}
	digest := strings.Repeat("a", 64)
	verifier := strings.Repeat("b", 64)
	review := `{"module_directory":".","package_directory":".","source_digest":"` + digest + `","gremlins_version":"v0.6.0","gremlins_verifier_sha256":"` + verifier + `","reason":"This detailed review confirms no viable mutation exists."}`
	if _, err := ParseZeroInventory(strings.NewReader(`{"schema_version":1,"packages":[` + review + `,` + review + `]}`)); !errors.Is(err, ErrInvalid) {
		t.Fatalf("ParseZeroInventory(duplicate) error = %v", err)
	}
}

func archiveWithHeaders(t *testing.T, headers []zip.FileHeader) []byte {
	t.Helper()
	var output bytes.Buffer
	writer := zip.NewWriter(&output)
	for index := range headers {
		if _, err := writer.CreateHeader(&headers[index]); err != nil {
			t.Fatalf("CreateHeader(): %v", err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Close(): %v", err)
	}
	return output.Bytes()
}

func directoryHeader(name string) zip.FileHeader {
	header := zip.FileHeader{Name: name}
	header.SetMode(0o755 | 1<<31)
	return header
}

func symlinkHeader(name string) zip.FileHeader {
	header := zip.FileHeader{Name: name}
	header.SetMode(0o777 | 1<<27)
	return header
}

func sizedHeader(uncompressed, compressed uint64) zip.FileHeader {
	return zip.FileHeader{Name: "mutation-checkpoints/a.json", UncompressedSize64: uncompressed, CompressedSize64: compressed}
}

func checkpointMap(t *testing.T) map[string]any {
	t.Helper()
	var value map[string]any
	if err := json.Unmarshal([]byte(validCheckpointJSON()), &value); err != nil {
		t.Fatal(err)
	}
	return value
}

func validCheckpointJSON() string {
	return `{"schema_version":3,"module":".","package":".","gate_input_digest":"` + strings.Repeat("a", 64) + `","gremlins_version":"v0.6.0","environment":{"GOVERSION":"go1.26.6"},"report":{"files":[{"file_name":"example.go","mutations":[{"type":"NEGATION","status":"KILLED","line":1,"column":1}]}]}}`
}

func reportOf(value map[string]any) map[string]any { return value["report"].(map[string]any) }
func fileOf(value map[string]any) map[string]any {
	return reportOf(value)["files"].([]any)[0].(map[string]any)
}
func mutationOf(value map[string]any) map[string]any {
	return fileOf(value)["mutations"].([]any)[0].(map[string]any)
}

func cloneMap(t *testing.T, value map[string]any) map[string]any {
	t.Helper()
	var clone map[string]any
	if err := json.Unmarshal(marshalMap(t, value), &clone); err != nil {
		t.Fatal(err)
	}
	return clone
}

func marshalMap(t *testing.T, value map[string]any) []byte {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func assertInvalidContains(t *testing.T, err error, want string) {
	t.Helper()
	if !errors.Is(err, ErrInvalid) || !strings.Contains(err.Error(), want) {
		t.Fatalf("error = %v, want ErrInvalid containing %q", err, want)
	}
}

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) { return 0, errors.New("read failed") }
