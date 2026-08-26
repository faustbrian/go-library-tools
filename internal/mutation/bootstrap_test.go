package mutation_test

import (
	"archive/zip"
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/faustbrian/go-library-tools/internal/mutation"
)

func TestReadBootstrapAcceptsKilledCheckpoints(t *testing.T) {
	archive := bootstrapArchive(t, map[string]string{
		"mutation-checkpoints/root.json": checkpointJSON("KILLED"),
	})

	checkpoints, err := mutation.ReadBootstrap(bytes.NewReader(archive), int64(len(archive)))
	if err != nil {
		t.Fatalf("ReadBootstrap() error = %v", err)
	}
	if len(checkpoints) != 1 {
		t.Fatalf("ReadBootstrap() count = %d, want 1", len(checkpoints))
	}
	checkpoint := checkpoints[0]
	if checkpoint.Module != "." || checkpoint.Package != "." || checkpoint.Mutants != 1 || !strings.HasPrefix(checkpoint.ReportDigest, "sha256:") {
		t.Fatalf("ReadBootstrap() checkpoint = %#v", checkpoint)
	}
}

func TestReadBootstrapRejectsUnsafeOrUnprovenEvidence(t *testing.T) {
	tests := map[string]struct {
		files map[string]string
		want  string
	}{
		"traversal": {
			files: map[string]string{"../checkpoint.json": checkpointJSON("KILLED")},
			want:  "unsafe archive path",
		},
		"unexpected entry": {
			files: map[string]string{"notes.txt": "not evidence"},
			want:  "unexpected archive entry",
		},
		"surviving mutant": {
			files: map[string]string{"mutation-checkpoints/root.json": checkpointJSON("LIVED")},
			want:  "non-killed mutant",
		},
		"duplicate package": {
			files: map[string]string{
				"mutation-checkpoints/one.json": checkpointJSON("KILLED"),
				"mutation-checkpoints/two.json": checkpointJSON("KILLED"),
			},
			want: "duplicate checkpoint identity",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			archive := bootstrapArchive(t, test.files)
			_, err := mutation.ReadBootstrap(bytes.NewReader(archive), int64(len(archive)))
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("ReadBootstrap() error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestReadBootstrapRejectsOversizedArchive(t *testing.T) {
	_, err := mutation.ReadBootstrap(bytes.NewReader(nil), mutation.MaximumArchiveSize+1)
	if !errors.Is(err, mutation.ErrInvalid) {
		t.Fatalf("ReadBootstrap() error = %v, want ErrInvalid", err)
	}
}

func TestValidateReportRequiresCompleteMutationKills(t *testing.T) {
	result, err := mutation.ValidateReport(strings.NewReader(`{"files":[{"file_name":"example.go","mutations":[{"type":"NEGATION","status":"KILLED","line":1,"column":1}]}]}`))
	if err != nil {
		t.Fatalf("ValidateReport() error = %v", err)
	}
	if result.Mutants != 1 || !strings.HasPrefix(result.Digest, "sha256:") {
		t.Fatalf("ValidateReport() = %#v", result)
	}
	if _, err := mutation.ValidateReport(strings.NewReader(`{"files":[{"file_name":"example.go","mutations":[{"type":"NEGATION","status":"LIVED","line":1,"column":1}]}]}`)); !errors.Is(err, mutation.ErrInvalid) {
		t.Fatalf("ValidateReport(survivor) error = %v", err)
	}
}

func bootstrapArchive(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var output bytes.Buffer
	writer := zip.NewWriter(&output)
	for name, body := range files {
		entry, err := writer.Create(name)
		if err != nil {
			t.Fatalf("Create(%q): %v", name, err)
		}
		if _, err := entry.Write([]byte(body)); err != nil {
			t.Fatalf("Write(%q): %v", name, err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Close(): %v", err)
	}
	return output.Bytes()
}

func checkpointJSON(status string) string {
	return `{
  "schema_version": 3,
  "module": ".",
  "package": ".",
  "gate_input_digest": "f91204781a2bcaf807c34aa49cf8228cf0b21919a9b20da968b45f76cd866df7",
  "gremlins_version": "v0.6.0",
  "environment": {"GOVERSION":"go1.26.6","GOOS":"darwin","GOARCH":"arm64","CGO_ENABLED":"1"},
  "report": {
    "go_module": "github.com/faustbrian/golib/pkg/example",
    "files": [{"file_name":"example.go","mutations":[{"type":"CONDITIONALS_NEGATION","status":"` + status + `","line":7,"column":3}]}]
  }
}`
}
