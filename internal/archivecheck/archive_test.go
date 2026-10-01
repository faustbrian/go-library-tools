package archivecheck_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"math"
	"strings"
	"testing"

	"github.com/faustbrian/go-library-tools/v2/internal/archivecheck"
)

func TestValidateRejectsUnsafeAndOversizedBootstrapArchives(t *testing.T) {
	tests := []struct {
		name    string
		header  tar.Header
		payload string
		limits  archivecheck.Limits
		want    string
	}{
		{"parent traversal", tar.Header{Name: "../escape", Mode: 0o600, Typeflag: tar.TypeReg}, "x", archivecheck.Limits{Entries: 2, Bytes: 4 << 10}, "unsafe archive path"},
		{"absolute path", tar.Header{Name: "/escape", Mode: 0o600, Typeflag: tar.TypeReg}, "x", archivecheck.Limits{Entries: 2, Bytes: 4 << 10}, "unsafe archive path"},
		{"non canonical path", tar.Header{Name: "proxy/../escape", Mode: 0o600, Typeflag: tar.TypeReg}, "x", archivecheck.Limits{Entries: 2, Bytes: 4 << 10}, "non-canonical archive path"},
		{"symbolic link", tar.Header{Name: "proxy/link", Linkname: "../../escape", Mode: 0o777, Typeflag: tar.TypeSymlink}, "", archivecheck.Limits{Entries: 2, Bytes: 4 << 10}, "unsupported archive entry"},
		{"oversized entry", tar.Header{Name: "proxy/data", Mode: 0o600, Typeflag: tar.TypeReg}, "12345", archivecheck.Limits{Entries: 2, Bytes: 4}, "expanded byte limit"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := archivecheck.Validate(bytes.NewReader(archive(t, test.header, test.payload)), test.limits)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Validate() error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestValidateDoesNotDiscloseRejectedArchiveNames(t *testing.T) {
	secret := "credentials-AKIA1234567890-secret"
	tests := []struct {
		name   string
		header tar.Header
		want   string
	}{
		{"unsafe", tar.Header{Name: "../" + secret, Mode: 0o600, Typeflag: tar.TypeReg}, "unsafe archive path"},
		{"non canonical", tar.Header{Name: "proxy/../" + secret, Mode: 0o600, Typeflag: tar.TypeReg}, "non-canonical archive path"},
		{"unsupported", tar.Header{Name: secret, Mode: 0o600, Typeflag: tar.TypeSymlink}, "unsupported archive entry"},
		{"unsafe mode", tar.Header{Name: secret, Mode: 0o4600, Typeflag: tar.TypeReg}, "unsafe archive mode"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := archivecheck.Validate(bytes.NewReader(archive(t, test.header, "")), archivecheck.Limits{Entries: 4, Bytes: 4 << 10})
			if err == nil || err.Error() != test.want || strings.Contains(err.Error(), secret) {
				t.Fatalf("Validate() error = %q, want exact safe class %q", err, test.want)
			}
		})
	}

	value := archives(t,
		archiveEntry{header: tar.Header{Name: secret, Mode: 0o600, Typeflag: tar.TypeReg}},
		archiveEntry{header: tar.Header{Name: secret, Mode: 0o600, Typeflag: tar.TypeReg}},
	)
	err := archivecheck.Validate(bytes.NewReader(value), archivecheck.Limits{Entries: 4, Bytes: 4 << 10})
	if err == nil || err.Error() != "duplicate archive path" || strings.Contains(err.Error(), secret) {
		t.Fatalf("Validate() duplicate error = %q", err)
	}
}

func TestValidateRejectsDuplicateAndCompressedOversizedArchives(t *testing.T) {
	value := archives(t,
		archiveEntry{header: tar.Header{Name: "proxy/data", Mode: 0o600, Typeflag: tar.TypeReg}, payload: "one"},
		archiveEntry{header: tar.Header{Name: "proxy/data", Mode: 0o600, Typeflag: tar.TypeReg}, payload: "two"},
	)
	if err := archivecheck.Validate(bytes.NewReader(value), archivecheck.Limits{Entries: 4, Bytes: 4 << 10}); err == nil || !strings.Contains(err.Error(), "duplicate archive path") {
		t.Fatalf("Validate() duplicate error = %v", err)
	}
	if err := archivecheck.Validate(bytes.NewReader(value), archivecheck.Limits{Entries: 4, Bytes: 4 << 10, CompressedBytes: 8}); err == nil || !strings.Contains(err.Error(), "compressed byte limit") {
		t.Fatalf("Validate() compressed error = %v", err)
	}
}

func TestValidateAcceptsBoundedRegularBootstrapArchive(t *testing.T) {
	value := archive(t, tar.Header{Name: "proxy/cache/download/example/@v/v1.0.0.mod", Mode: 0o600, Typeflag: tar.TypeReg}, "module example\n")
	if err := archivecheck.Validate(bytes.NewReader(value), archivecheck.Limits{Entries: 4, Bytes: 4 << 10}); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestValidateRejectsInvalidGzipTrailer(t *testing.T) {
	value := archive(t, tar.Header{Name: "proxy/data", Mode: 0o600, Typeflag: tar.TypeReg}, "value")
	value[len(value)-1] ^= 0xff
	if err := archivecheck.Validate(bytes.NewReader(value), archivecheck.Limits{Entries: 4, Bytes: 4 << 10}); err == nil || !strings.Contains(err.Error(), "finish gzip archive") {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestValidateRejectsOversizedExtendedHeaderMetadata(t *testing.T) {
	value := archive(t, tar.Header{
		Name:     "proxy/data",
		Mode:     0o600,
		Typeflag: tar.TypeReg,
		PAXRecords: map[string]string{
			"comment": strings.Repeat("x", 8<<10),
		},
	}, "")

	err := archivecheck.Validate(bytes.NewReader(value), archivecheck.Limits{Entries: 4, Bytes: 2 << 10})
	if err == nil || !strings.Contains(err.Error(), "expanded byte limit") {
		t.Fatalf("Validate() error = %v, want expanded byte limit", err)
	}
}

func FuzzValidateNeverPanics(f *testing.F) {
	f.Add([]byte("not an archive"))
	f.Add(archive(f, tar.Header{Name: "proxy/data", Mode: 0o600, Typeflag: tar.TypeReg}, "value"))
	f.Fuzz(func(_ *testing.T, value []byte) {
		_ = archivecheck.Validate(bytes.NewReader(value), archivecheck.Limits{
			Entries:         32,
			Bytes:           1 << 20,
			CompressedBytes: 1 << 20,
		})
	})
}

func archive(t testing.TB, header tar.Header, payload string) []byte {
	return archives(t, archiveEntry{header: header, payload: payload})
}

type archiveEntry struct {
	header  tar.Header
	payload string
}

func archives(t testing.TB, entries ...archiveEntry) []byte {
	t.Helper()
	var value bytes.Buffer
	gzipWriter := gzip.NewWriter(&value)
	tarWriter := tar.NewWriter(gzipWriter)
	for _, entry := range entries {
		entry.header.Size = int64(len(entry.payload))
		if err := tarWriter.WriteHeader(&entry.header); err != nil {
			t.Fatal(err)
		}
		if _, err := tarWriter.Write([]byte(entry.payload)); err != nil {
			t.Fatal(err)
		}
	}
	if err := tarWriter.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gzipWriter.Close(); err != nil {
		t.Fatal(err)
	}
	return value.Bytes()
}

func TestValidateLimitBoundaries(t *testing.T) {
	value := archive(t, tar.Header{Name: "proxy/data", Mode: 0o600, Typeflag: tar.TypeReg}, "value")
	for _, limits := range []archivecheck.Limits{
		{Entries: 0, Bytes: 4096},
		{Entries: -1, Bytes: 4096},
		{Entries: 1, Bytes: 0},
		{Entries: 1, Bytes: -1},
	} {
		if err := archivecheck.Validate(bytes.NewReader(value), limits); err == nil || err.Error() != "archive limits must be positive" {
			t.Fatalf("Validate(%+v) = %v, want positive limits rejection", limits, err)
		}
	}
	for _, compressedLimit := range []int64{0, -1, math.MaxInt64} {
		if err := archivecheck.Validate(bytes.NewReader(value), archivecheck.Limits{Entries: 1, Bytes: math.MaxInt64, CompressedBytes: compressedLimit}); err != nil {
			t.Fatalf("Validate() with compressed limit %d = %v", compressedLimit, err)
		}
	}
}

func TestValidateEntryCountBoundaries(t *testing.T) {
	value := archives(t,
		archiveEntry{header: tar.Header{Name: "proxy/", Mode: 0o700, Typeflag: tar.TypeDir}},
		archiveEntry{header: tar.Header{Name: "proxy/data", Mode: 0o600, Typeflag: tar.TypeReg}, payload: "value"},
	)
	for _, limit := range []int{1, 2, 3} {
		err := archivecheck.Validate(bytes.NewReader(value), archivecheck.Limits{Entries: limit, Bytes: 8192})
		if limit == 1 {
			if err == nil || err.Error() != "archive entry limit exceeded: 1" {
				t.Fatalf("Validate() = %v, want entry-count rejection", err)
			}
		} else if err != nil {
			t.Fatalf("Validate() with entry bound %d = %v", limit, err)
		}
	}
}

func TestValidateExpandedStreamBoundariesAndTrailingData(t *testing.T) {
	value := archive(t, tar.Header{Name: "proxy/data", Mode: 0o600, Typeflag: tar.TypeReg}, "value")
	reader, err := gzip.NewReader(bytes.NewReader(value))
	if err != nil {
		t.Fatal(err)
	}
	raw, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	if err := reader.Close(); err != nil {
		t.Fatal(err)
	}
	for _, delta := range []int64{-1, 0, 1} {
		limit := int64(len(raw)) + delta
		err := archivecheck.Validate(bytes.NewReader(value), archivecheck.Limits{Entries: 1, Bytes: limit})
		if delta < 0 {
			if err == nil || !strings.HasPrefix(err.Error(), "expanded byte limit exceeded:") {
				t.Fatalf("Validate() below expanded length = %v, want limit rejection", err)
			}
		} else if err != nil {
			t.Fatalf("Validate() at expanded length %+d = %v", delta, err)
		}
	}
	for _, trailing := range [][]byte{
		gzipBytes(t, append(append([]byte(nil), raw...), 'x')),
		append(append([]byte(nil), value...), gzipBytes(t, []byte{'x'})...),
	} {
		err := archivecheck.Validate(bytes.NewReader(trailing), archivecheck.Limits{Entries: 1, Bytes: 8192})
		if err == nil || err.Error() != "gzip archive contains trailing decompressed data" {
			t.Fatalf("Validate() = %v, want trailing-content rejection", err)
		}
	}
}

func TestValidateCompressedStreamBoundaries(t *testing.T) {
	value := archive(t, tar.Header{Name: "proxy/data", Mode: 0o600, Typeflag: tar.TypeReg}, "value")
	for _, delta := range []int64{-1, 0, 1} {
		limit := int64(len(value)) + delta
		err := archivecheck.Validate(bytes.NewReader(value), archivecheck.Limits{Entries: 1, Bytes: 8192, CompressedBytes: limit})
		if delta < 0 {
			if err == nil || !strings.HasPrefix(err.Error(), "compressed byte limit exceeded:") {
				t.Fatalf("Validate() below compressed length = %v, want limit rejection", err)
			}
		} else if err != nil {
			t.Fatalf("Validate() at compressed length %+d = %v", delta, err)
		}
	}
	for _, limit := range []int64{10, int64(len(value)) - 2} {
		err := archivecheck.Validate(bytes.NewReader(value), archivecheck.Limits{Entries: 1, Bytes: 8192, CompressedBytes: limit})
		if err == nil || !strings.HasPrefix(err.Error(), "compressed byte limit exceeded:") {
			t.Fatalf("Validate() with truncated compressed stream %d = %v", limit, err)
		}
	}
}

func TestValidateRejectsOversizedDeclaredPayloadBeforeReadingIt(t *testing.T) {
	for _, test := range []struct {
		name         string
		firstPayload int
		declaredSize int64
		limit        int64
	}{
		{"single declaration", 0, 4096, 1024},
		{"cumulative declarations", 512, 1792, 2048},
	} {
		t.Run(test.name, func(t *testing.T) {
			var raw bytes.Buffer
			writer := tar.NewWriter(&raw)
			if test.firstPayload > 0 {
				if err := writer.WriteHeader(&tar.Header{Name: "proxy/first", Mode: 0o600, Typeflag: tar.TypeReg, Size: int64(test.firstPayload)}); err != nil {
					t.Fatal(err)
				}
				if _, err := writer.Write(bytes.Repeat([]byte{'a'}, test.firstPayload)); err != nil {
					t.Fatal(err)
				}
			}
			if err := writer.WriteHeader(&tar.Header{Name: "proxy/oversized", Mode: 0o600, Typeflag: tar.TypeReg, Size: test.declaredSize}); err != nil {
				t.Fatal(err)
			}
			// The complete headers fit; no declared oversized body is supplied.
			// Rejection must precede an attempt to consume that missing body.
			err := archivecheck.Validate(bytes.NewReader(gzipBytes(t, raw.Bytes())), archivecheck.Limits{Entries: 2, Bytes: test.limit})
			if err == nil || !strings.HasPrefix(err.Error(), "expanded byte limit exceeded:") {
				t.Fatalf("Validate() = %v, want declared-payload limit rejection", err)
			}
		})
	}
}

func TestValidatePreservesReaderFailuresAcrossArchiveStages(t *testing.T) {
	readerFailure := errors.New("fixture read failure")
	valid := archive(t, tar.Header{Name: "proxy/data", Mode: 0o600, Typeflag: tar.TypeReg}, "value")
	for _, test := range []struct {
		name  string
		value []byte
		want  string
	}{
		{"gzip header", nil, "open gzip archive:"},
		{"tar header", gzipBytes(t, []byte("not a tar header")), "read tar archive:"},
		{"gzip completion", valid, "finish gzip archive:"},
	} {
		t.Run(test.name, func(t *testing.T) {
			source := terminalErrorReader{source: bytes.NewReader(test.value), terminal: readerFailure}
			err := archivecheck.Validate(source, archivecheck.Limits{Entries: 2, Bytes: 8192})
			if err == nil || !strings.HasPrefix(err.Error(), test.want) || !errors.Is(err, readerFailure) {
				t.Fatalf("Validate() = %v, want %q preserving reader cause", err, test.want)
			}
		})
	}
}

func gzipBytes(t testing.TB, raw []byte) []byte {
	t.Helper()
	var value bytes.Buffer
	writer := gzip.NewWriter(&value)
	if _, err := writer.Write(raw); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return value.Bytes()
}

type terminalErrorReader struct {
	source   *bytes.Reader
	terminal error
}

func (r terminalErrorReader) Read(p []byte) (int, error) {
	n, err := r.source.Read(p)
	if errors.Is(err, io.EOF) {
		return n, r.terminal
	}
	return n, err
}
