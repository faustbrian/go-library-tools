package cli_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/faustbrian/go-library-tools/v2/internal/cli"
)

func TestBinaryReleaseRefusalsDoNotCreateOrOverwriteOutput(t *testing.T) {
	root := t.TempDir()
	binary := filepath.Join(root, "not-a-binary")
	license := filepath.Join(root, "LICENSE")
	output := filepath.Join(root, "release.tgz")
	for _, path := range []string{binary, license} {
		if err := os.WriteFile(path, []byte("input"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	args := []string{"release", "package", "--binary", binary, "--license", license, "--output", output}
	var stdout, stderr bytes.Buffer
	if code := cli.Execute(args, root, &stdout, &stderr); code != 1 || stdout.Len() != 0 {
		t.Fatalf("malformed binary result: %d %q %q", code, stdout.String(), stderr.String())
	}
	if _, err := os.Stat(output); !os.IsNotExist(err) {
		t.Fatal("refused packaging created output")
	}
	if err := os.WriteFile(output, []byte("preserve"), 0o600); err != nil {
		t.Fatal(err)
	}
	stdout.Reset()
	stderr.Reset()
	if code := cli.Execute([]string{"release", "archive-check", "--file", output}, root, &stdout, &stderr); code != 1 || stdout.Len() != 0 {
		t.Fatal("malformed archive accepted")
	}
	stored, err := os.ReadFile(output)
	if err != nil || string(stored) != "preserve" {
		t.Fatal("archive validation altered its input")
	}
}

func TestBinaryReleaseRejectsMalformedArgumentsWithoutRepositoryDiscovery(t *testing.T) {
	for _, args := range [][]string{
		{"release", "package"},
		{"release", "package", "--binary", "b", "--license", "l", "--output", ""},
		{"release", "package", "--wrong", "b", "--license", "l", "--output", "o"},
		{"release", "archive-check"},
		{"release", "archive-check", "--file", ""},
		{"release", "archive-check", "--file", "b", "extra"},
	} {
		var stdout, stderr bytes.Buffer
		if code := cli.Execute(args, t.TempDir(), &stdout, &stderr); code != 2 || stdout.Len() != 0 {
			t.Fatalf("malformed arguments result: %d %q %q", code, stdout.String(), stderr.String())
		}
	}
}
