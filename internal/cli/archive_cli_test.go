package cli_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/faustbrian/go-library-tools/v2/internal/cli"
)

func TestExecuteAcceptsBootstrapArchiveWithoutExtraction(t *testing.T) {
	root := fixture(t)
	input := bootstrapArchive(t, "proxy/data")
	path := filepath.Join(root, "bootstrap.tgz")
	if err := os.WriteFile(path, input, 0o600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	code := cli.Execute([]string{"archive", "validate", "--file", path}, root, &stdout, &stderr)
	if code != 0 || stdout.String() != "bootstrap archive valid\n" || stderr.Len() != 0 {
		t.Fatalf("accepted archive result: code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
	stored, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(stored, input) {
		t.Fatal("validation altered its input")
	}
	if _, err := os.Stat(filepath.Join(root, "proxy")); !os.IsNotExist(err) {
		t.Fatal("validation extracted archive entries")
	}
}

func TestExecuteRejectsMissingBootstrapArchive(t *testing.T) {
	root := fixture(t)
	path := filepath.Join(root, "missing.tgz")
	var stdout, stderr bytes.Buffer
	code := cli.Execute([]string{"archive", "validate", "--file", path}, root, &stdout, &stderr)
	if code != 1 || stdout.Len() != 0 || !strings.Contains(stderr.String(), "open archive") {
		t.Fatalf("missing archive result: code=%d stdout_empty=%t stderr=%q", code, stdout.Len() == 0, stderr.String())
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("validation created a missing input")
	}
}

func TestExecuteRejectsMalformedArchiveArguments(t *testing.T) {
	root := fixture(t)
	for _, args := range [][]string{
		{"archive"}, {"archive", "validate"}, {"archive", "validate", "--file"},
		{"archive", "invalid", "--file", "missing.tgz"},
		{"archive", "validate", "--wrong", "missing.tgz"},
		{"archive", "validate", "--file", ""},
		{"archive", "validate", "--file", "missing.tgz", "extra"},
	} {
		var stdout, stderr bytes.Buffer
		code := cli.Execute(args, root, &stdout, &stderr)
		if code != 2 || stdout.Len() != 0 || stderr.String() != "usage: golib archive validate --file <path>\n" {
			t.Fatalf("malformed archive result: code=%d stdout_empty=%t stderr=%q", code, stdout.Len() == 0, stderr.String())
		}
	}
}

func TestExecuteRejectsUnsafeBootstrapArchiveWithoutDisclosure(t *testing.T) {
	root := fixture(t)
	name := "../rejected-entry"
	input := bootstrapArchive(t, name)
	path := filepath.Join(root, "unsafe.tgz")
	if err := os.WriteFile(path, input, 0o600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	code := cli.Execute([]string{"archive", "validate", "--file", path}, root, &stdout, &stderr)
	if code != 1 || stdout.Len() != 0 || stderr.Len() == 0 || strings.Contains(stderr.String(), name) {
		t.Fatal("unsafe archive must fail without success output or entry-name disclosure")
	}
	stored, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(stored, input) {
		t.Fatal("validation altered unsafe input")
	}
}

func bootstrapArchive(t *testing.T, name string) []byte {
	t.Helper()
	var data bytes.Buffer
	gzipWriter := gzip.NewWriter(&data)
	tarWriter := tar.NewWriter(gzipWriter)
	if name == "proxy/data" {
		if err := tarWriter.WriteHeader(&tar.Header{Name: "proxy/", Typeflag: tar.TypeDir, Mode: 0o700}); err != nil {
			t.Fatal(err)
		}
	}
	if err := tarWriter.WriteHeader(&tar.Header{Name: name, Typeflag: tar.TypeReg, Mode: 0o600, Size: 5}); err != nil {
		t.Fatal(err)
	}
	if _, err := tarWriter.Write([]byte("value")); err != nil {
		t.Fatal(err)
	}
	if err := tarWriter.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gzipWriter.Close(); err != nil {
		t.Fatal(err)
	}
	return data.Bytes()
}
