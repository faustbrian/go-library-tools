package cli

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/faustbrian/go-library-tools/v2/internal/archivecheck"
	"github.com/faustbrian/go-library-tools/v2/internal/repositoryfile"
	"github.com/faustbrian/go-library-tools/v2/internal/securitydocs"
)

func TestValidationPublicationAfterCompletedSecurity(t *testing.T) {
	directory := t.TempDir()
	for name, content := range map[string]string{
		"risk-register.json":   `{"schema_version":1,"risks":[]}`,
		"security-matrix.json": `{"schema_version":1,"modules":[]}`,
	} {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	for _, canceled := range []bool{false, true} {
		ctx, cancel := context.WithCancel(context.Background())
		if err := securitydocs.ValidateContext(ctx, directory); err != nil {
			cancel()
			t.Fatal(err)
		}
		assertValidationPublication(t, ctx, cancel, canceled, "ecosystem security records valid\n")
	}
}

func TestValidationPublicationAfterClosedArchive(t *testing.T) {
	var archive bytes.Buffer
	gzipWriter := gzip.NewWriter(&archive)
	if err := tar.NewWriter(gzipWriter).Close(); err != nil {
		t.Fatal(err)
	}
	if err := gzipWriter.Close(); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "ordinary.tgz")
	if err := os.WriteFile(path, archive.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, canceled := range []bool{false, true} {
		ctx, cancel := context.WithCancel(context.Background())
		file, err := repositoryfile.OpenContext(ctx, path, 256)
		if err != nil {
			cancel()
			t.Fatal(err)
		}
		validateErr := archivecheck.Validate(file, archivecheck.Limits{
			Entries: 1, Bytes: 2048, CompressedBytes: 256,
		})
		closeErr := file.Close()
		if validateErr != nil || closeErr != nil {
			cancel()
			t.Fatalf("ordinary validation/close failed: %v, %v", validateErr, closeErr)
		}
		assertValidationPublication(t, ctx, cancel, canceled, "bootstrap archive valid\n")
	}
}

func assertValidationPublication(t *testing.T, ctx context.Context, cancel context.CancelFunc, canceled bool, message string) {
	t.Helper()
	defer cancel()
	if err := ctx.Err(); err != nil {
		t.Fatal("validation must finish with the caller still active")
	}
	if canceled {
		cancel()
	}
	var stdout, stderr bytes.Buffer
	code := publishValidationSuccess(ctx, &stdout, &stderr, message)
	if canceled {
		if code != 1 || stdout.Len() != 0 || stderr.String() != "context canceled\n" {
			t.Fatalf("canceled publication: code%d, stdout bytes%d, stderr%q", code, stdout.Len(), stderr.String())
		}
	} else if code != 0 || stdout.String() != message || stderr.Len() != 0 {
		t.Fatalf("active publication: code%d, stdout%q, stderr bytes%d", code, stdout.String(), stderr.Len())
	}
}
