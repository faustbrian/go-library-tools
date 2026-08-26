package docscheck_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"

	"github.com/faustbrian/go-library-tools/internal/docscheck"
)

func FuzzLycheeArchive(f *testing.F) {
	f.Add([]byte("not an archive"))
	f.Add(validLycheeArchive(f))
	f.Fuzz(func(t *testing.T, input []byte) {
		if len(input) > 2<<20 {
			return
		}
		archive := filepath.Join(t.TempDir(), "archive.tar.gz")
		if err := os.WriteFile(archive, input, 0o600); err != nil {
			t.Fatal(err)
		}
		digest := sha256.Sum256(input)
		release := docscheck.LycheeRelease{Target: "fuzz", SHA256: hex.EncodeToString(digest[:])}
		_, _ = docscheck.ExtractLychee(archive, release)
	})
}

func validLycheeArchive(t testing.TB) []byte {
	t.Helper()
	var compressed bytes.Buffer
	gzipWriter := gzip.NewWriter(&compressed)
	tarWriter := tar.NewWriter(gzipWriter)
	contents := []byte("binary")
	if err := tarWriter.WriteHeader(&tar.Header{Name: "lychee-fuzz/lychee", Typeflag: tar.TypeReg, Mode: 0o700, Size: int64(len(contents))}); err != nil {
		t.Fatal(err)
	}
	if _, err := tarWriter.Write(contents); err != nil {
		t.Fatal(err)
	}
	if err := tarWriter.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gzipWriter.Close(); err != nil {
		t.Fatal(err)
	}
	return compressed.Bytes()
}
