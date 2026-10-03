package cli_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/faustbrian/go-library-tools/v2/internal/cli"
)

func TestSecurityCommandHonorsCanceledCallerBeforeSuccess(t *testing.T) {
	root := fixture(t)
	directory := t.TempDir()
	write(t, filepath.Join(directory, "risk-register.json"), `{"schema_version":1,"risks":[]}`)
	write(t, filepath.Join(directory, "security-matrix.json"), `{"schema_version":1,"modules":[]}`)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var stdout, stderr bytes.Buffer
	code := cli.ExecuteContext(ctx, []string{"security", "validate", "--directory", directory}, root, &stdout, &stderr)
	if code != 1 || stdout.Len() != 0 || stderr.Len() == 0 {
		t.Fatalf("canceled security command = code%d, success output%t, error output%t", code, stdout.Len() != 0, stderr.Len() != 0)
	}
}

func TestArchiveCommandRegularControlAndDirectoryAdmission(t *testing.T) {
	root := fixture(t)
	var content bytes.Buffer
	gz := gzip.NewWriter(&content)
	tarWriter := tar.NewWriter(gz)
	if err := tarWriter.WriteHeader(&tar.Header{Name: "ordinary.txt", Mode: 0o600, Size: 3}); err != nil {
		t.Fatal(err)
	}
	if _, err := tarWriter.Write([]byte("abc")); err != nil {
		t.Fatal(err)
	}
	if err := tarWriter.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "ordinary.tgz")
	if err := os.WriteFile(path, content.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		path string
		code int
	}{{path, 0}, {t.TempDir(), 1}} {
		var stdout, stderr bytes.Buffer
		code := cli.ExecuteContext(context.Background(), []string{"archive", "validate", "--file", test.path}, root, &stdout, &stderr)
		if code != test.code || (code == 0 && (stdout.String() != "bootstrap archive valid\n" || stderr.Len() != 0)) || (code != 0 && stdout.Len() != 0) {
			t.Fatalf("archive code%d, success output%t, error output%t", code, stdout.Len() != 0, stderr.Len() != 0)
		}
	}
}
