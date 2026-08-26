package docscheck_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/faustbrian/go-library-tools/internal/docscheck"
)

func TestLycheeReleaseResolvesSupportedPlatforms(t *testing.T) {
	tests := []struct {
		goos, goarch, target, checksum string
	}{
		{"darwin", "arm64", "aarch64-apple-darwin", "c9d3740ea2d891854d37116c9fba840f37b6e7c89d330e7db84ac333631c4977"},
		{"darwin", "amd64", "x86_64-apple-darwin", "887503a9cff667d322b8d0892b40bf49976eb9507af8483220a3706cdad55978"},
		{"linux", "arm64", "aarch64-unknown-linux-gnu", "91a7bd65685da41b90ccb9bc867a3d649a7818042dae04ff405e55a25bddee4c"},
		{"linux", "amd64", "x86_64-unknown-linux-gnu", "1f4e0ef7f6554a6ed33dd7ac144fb2e1bbed98598e7af973042fc5cd43951c9a"},
	}
	for _, test := range tests {
		release, err := docscheck.LycheeReleaseFor(test.goos, test.goarch)
		if err != nil {
			t.Fatalf("LycheeReleaseFor(%s, %s) error = %v", test.goos, test.goarch, err)
		}
		if release.Version != "0.24.2" || release.Target != test.target || release.SHA256 != test.checksum ||
			!strings.HasSuffix(release.URL, "/lychee-v0.24.2/lychee-"+test.target+".tar.gz") {
			t.Fatalf("LycheeReleaseFor(%s, %s) = %#v", test.goos, test.goarch, release)
		}
	}
}

func TestLycheeReleaseRejectsUnsupportedPlatforms(t *testing.T) {
	if _, err := docscheck.LycheeReleaseFor("plan9", "amd64"); err == nil || !strings.Contains(err.Error(), "unsupported lychee platform") {
		t.Fatalf("LycheeReleaseFor() error = %v", err)
	}
}

func TestExtractLycheeReturnsOnlyVerifiedBinary(t *testing.T) {
	release, archive := lycheeArchive(t, []tarEntry{
		{name: "lychee-test/", kind: tar.TypeDir},
		{name: "lychee-test/README.md", body: []byte("documentation")},
		{name: "lychee-test/lychee", body: []byte("verified binary")},
	})
	binary, err := docscheck.ExtractLychee(archive, release)
	if err != nil {
		t.Fatalf("ExtractLychee() error = %v", err)
	}
	if string(binary) != "verified binary" {
		t.Fatalf("ExtractLychee() = %q", binary)
	}
}

func TestExtractLycheeRejectsUntrustedArchives(t *testing.T) {
	release, archive := lycheeArchive(t, []tarEntry{{name: "lychee-test/lychee", body: []byte("binary")}})
	tests := []struct {
		name    string
		archive string
		release docscheck.LycheeRelease
		want    string
	}{
		{name: "malformed checksum", archive: archive, release: withLycheeChecksum(release, "bad"), want: "malformed lychee checksum"},
		{name: "uppercase checksum", archive: archive, release: withLycheeChecksum(release, strings.ToUpper(release.SHA256)), want: "malformed lychee checksum"},
		{name: "checksum mismatch", archive: archive, release: withLycheeChecksum(release, strings.Repeat("0", 64)), want: "checksum mismatch"},
		{name: "missing archive", archive: filepath.Join(t.TempDir(), "missing"), release: release, want: "inspect lychee archive"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := docscheck.ExtractLychee(test.archive, test.release); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("ExtractLychee() error = %v", err)
			}
		})
	}

	link := filepath.Join(t.TempDir(), "archive.tar.gz")
	if err := os.Symlink(archive, link); err != nil {
		t.Fatal(err)
	}
	if _, err := docscheck.ExtractLychee(link, release); err == nil || !strings.Contains(err.Error(), "regular file") {
		t.Fatalf("ExtractLychee() symlink error = %v", err)
	}

	large := filepath.Join(t.TempDir(), "archive.tar.gz")
	if err := os.WriteFile(large, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Truncate(large, 129<<20); err != nil {
		t.Fatal(err)
	}
	if _, err := docscheck.ExtractLychee(large, release); err == nil || !strings.Contains(err.Error(), "too large") {
		t.Fatalf("ExtractLychee() large error = %v", err)
	}

	unreadable := filepath.Join(t.TempDir(), "archive.tar.gz")
	if err := os.WriteFile(unreadable, []byte("archive"), 0o000); err != nil {
		t.Fatal(err)
	}
	unreadableRelease := withLycheeChecksum(release, checksumOf([]byte("archive")))
	if _, err := docscheck.ExtractLychee(unreadable, unreadableRelease); err == nil || !strings.Contains(err.Error(), "read lychee archive") {
		t.Fatalf("ExtractLychee() unreadable error = %v", err)
	}
}

func TestExtractLycheeRejectsMalformedContents(t *testing.T) {
	tests := []struct {
		name    string
		entries []tarEntry
		want    string
	}{
		{name: "unsafe path", entries: []tarEntry{{name: "../lychee", body: []byte("binary")}}, want: "unsafe archive path"},
		{name: "backslash path", entries: []tarEntry{{name: `lychee-test\\lychee`, body: []byte("binary")}}, want: "unsafe archive path"},
		{name: "link", entries: []tarEntry{{name: "lychee-test/lychee", kind: tar.TypeSymlink}}, want: "unsupported archive entry"},
		{name: "binary directory", entries: []tarEntry{{name: "lychee-test/lychee", kind: tar.TypeDir}}, want: "binary is not a regular file"},
		{name: "missing binary", entries: []tarEntry{{name: "lychee-test/README.md", body: []byte("docs")}}, want: "does not contain"},
		{name: "duplicate binary", entries: []tarEntry{{name: "lychee-test/lychee", body: []byte("one")}, {name: "lychee-test/lychee", body: []byte("two")}}, want: "duplicate lychee binary"},
		{name: "binary too large", entries: []tarEntry{{name: "lychee-test/lychee", size: 65 << 20}}, want: "binary is too large"},
		{name: "archive too large", entries: []tarEntry{{name: "lychee-test/README.md", size: 257 << 20}}, want: "expanded contents are too large"},
		{name: "truncated binary", entries: []tarEntry{{name: "lychee-test/lychee", size: 10, body: []byte("short")}}, want: "read lychee binary"},
		{name: "truncated file", entries: []tarEntry{{name: "lychee-test/README.md", size: 10, body: []byte("short")}}, want: "read archive entry"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			release, archive := lycheeArchive(t, test.entries)
			if _, err := docscheck.ExtractLychee(archive, release); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("ExtractLychee() error = %v", err)
			}
		})
	}
}

func TestExtractLycheeBoundsEntryCountAndCompression(t *testing.T) {
	entries := make([]tarEntry, 257)
	for index := range entries {
		entries[index] = tarEntry{name: "lychee-test/file-" + string(rune('a'+index%26)) + strings.Repeat("x", index/26)}
	}
	release, archive := lycheeArchive(t, entries)
	if _, err := docscheck.ExtractLychee(archive, release); err == nil || !strings.Contains(err.Error(), "too many entries") {
		t.Fatalf("ExtractLychee() entries error = %v", err)
	}

	invalid := filepath.Join(t.TempDir(), "invalid.tar.gz")
	if err := os.WriteFile(invalid, []byte("not gzip"), 0o600); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(invalid)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(data)
	badRelease := docscheck.LycheeRelease{Target: "test", SHA256: hex.EncodeToString(digest[:])}
	if _, err := docscheck.ExtractLychee(invalid, badRelease); err == nil || !strings.Contains(err.Error(), "decompress lychee archive") {
		t.Fatalf("ExtractLychee() gzip error = %v", err)
	}

	malformed := writeCompressedArchive(t, []byte("not a tar archive"))
	malformedData, err := os.ReadFile(malformed)
	if err != nil {
		t.Fatal(err)
	}
	malformedRelease := docscheck.LycheeRelease{Target: "test", SHA256: checksumOf(malformedData)}
	if _, err := docscheck.ExtractLychee(malformed, malformedRelease); err == nil || !strings.Contains(err.Error(), "read lychee archive header") {
		t.Fatalf("ExtractLychee() tar error = %v", err)
	}
}

type tarEntry struct {
	name string
	kind byte
	size int64
	body []byte
}

func lycheeArchive(t *testing.T, entries []tarEntry) (docscheck.LycheeRelease, string) {
	t.Helper()
	var compressed bytes.Buffer
	gzipWriter := gzip.NewWriter(&compressed)
	tarWriter := tar.NewWriter(gzipWriter)
	for _, entry := range entries {
		kind := entry.kind
		if kind == 0 {
			kind = tar.TypeReg
		}
		size := entry.size
		if size == 0 {
			size = int64(len(entry.body))
		}
		if err := tarWriter.WriteHeader(&tar.Header{Name: entry.name, Typeflag: kind, Mode: 0o700, Size: size}); err != nil {
			t.Fatal(err)
		}
		_, _ = tarWriter.Write(entry.body)
	}
	_ = tarWriter.Close()
	if err := gzipWriter.Close(); err != nil {
		t.Fatal(err)
	}
	archive := filepath.Join(t.TempDir(), "lychee.tar.gz")
	if err := os.WriteFile(archive, compressed.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(compressed.Bytes())
	return docscheck.LycheeRelease{Target: "test", SHA256: hex.EncodeToString(digest[:])}, archive
}

func withLycheeChecksum(release docscheck.LycheeRelease, checksum string) docscheck.LycheeRelease {
	release.SHA256 = checksum
	return release
}

func writeCompressedArchive(t *testing.T, contents []byte) string {
	t.Helper()
	var compressed bytes.Buffer
	writer := gzip.NewWriter(&compressed)
	if _, err := writer.Write(contents); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	archive := filepath.Join(t.TempDir(), "archive.tar.gz")
	if err := os.WriteFile(archive, compressed.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}
	return archive
}

func checksumOf(contents []byte) string {
	digest := sha256.Sum256(contents)
	return hex.EncodeToString(digest[:])
}
