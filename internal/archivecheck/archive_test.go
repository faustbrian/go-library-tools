package archivecheck_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
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
