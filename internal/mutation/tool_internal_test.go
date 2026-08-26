package mutation

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestArgumentsRejectMalformedInputs(t *testing.T) {
	for _, input := range []struct {
		target  string
		output  string
		tags    string
		workers int
	}{
		{"../outside", "/tmp/report", "", 1},
		{".", "relative", "", 1},
		{".", "/tmp/report", "bad\ntag", 1},
		{".", "/tmp/report", "", 0},
		{".", "/tmp/report", "", 65},
	} {
		if _, err := Arguments(input.target, input.output, input.tags, false, input.workers); !errors.Is(err, ErrInvalid) {
			t.Fatalf("Arguments(%#v) error = %v", input, err)
		}
	}
}

func TestArgumentsAcceptExactWorkerAndTargetBoundaries(t *testing.T) {
	for _, input := range []struct {
		target  string
		workers int
	}{
		{".", 1},
		{"./adapter", 64},
	} {
		arguments, err := Arguments(input.target, "/tmp/report", "", false, input.workers)
		if err != nil || arguments[1] != input.target {
			t.Fatalf("Arguments(%#v) = %#v, %v", input, arguments, err)
		}
	}
}

func TestBuildVerifierRejectsInvalidWorkspaceAndProcess(t *testing.T) {
	if _, err := BuildVerifier(context.Background(), ".", func(context.Context, string, []string, string, map[string]string, io.Writer, io.Writer) error {
		return nil
	}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("BuildVerifier(relative) error = %v", err)
	}
	if _, err := BuildVerifier(context.Background(), t.TempDir(), nil); !errors.Is(err, ErrInvalid) {
		t.Fatalf("BuildVerifier(nil process) error = %v", err)
	}
	workspace := filepath.Join(t.TempDir(), "workspace")
	if err := os.WriteFile(workspace, []byte("not a directory"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := BuildVerifier(context.Background(), workspace, verifierProcess("", "", false)); err == nil {
		t.Fatal("BuildVerifier(file workspace) error = nil")
	}
}

func TestBuildVerifierReportsEveryConstructionFailure(t *testing.T) {
	source := t.TempDir()
	if err := os.WriteFile(filepath.Join(source, "go.mod"), []byte("module example\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	tests := map[string]struct {
		prepare  func(string)
		metadata string
		fail     string
		binary   bool
	}{
		"download": {metadata: validDownloadMetadata(source), fail: "download"},
		"metadata": {metadata: "{"},
		"trailing metadata": {
			metadata: validDownloadMetadata(source) + `{}`,
		},
		"checksum": {metadata: `{"Dir":"` + source + `","Sum":"bad","GoModSum":"bad"}`},
		"copy":     {metadata: validDownloadMetadata(filepath.Join(source, "missing"))},
		"write patch": {metadata: validDownloadMetadata(source), prepare: func(workspace string) {
			if err := os.Mkdir(filepath.Join(workspace, "gremlins-run-all-mutants.patch"), 0o700); err != nil {
				t.Fatal(err)
			}
		}},
		"apply":       {metadata: validDownloadMetadata(source), fail: "git"},
		"build":       {metadata: validDownloadMetadata(source), fail: "build"},
		"open binary": {metadata: validDownloadMetadata(source)},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			workspace := t.TempDir()
			if test.prepare != nil {
				test.prepare(workspace)
			}
			process := verifierProcess(test.metadata, test.fail, test.binary)
			if _, err := BuildVerifier(context.Background(), workspace, process); err == nil {
				t.Fatal("BuildVerifier() error = nil")
			}
		})
	}
}

func TestBuildVerifierRequiresEachPinnedChecksum(t *testing.T) {
	source := t.TempDir()
	if err := os.WriteFile(filepath.Join(source, "go.mod"), []byte("module example\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	for name, metadata := range map[string]string{
		"module": `{"Dir":"` + source + `","Sum":"bad","GoModSum":"` + gremlinsGoModSum + `"}`,
		"go.mod": `{"Dir":"` + source + `","Sum":"` + gremlinsSum + `","GoModSum":"bad"}`,
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := BuildVerifier(context.Background(), t.TempDir(), verifierProcess(metadata, "", false)); !errors.Is(err, ErrInvalid) {
				t.Fatalf("BuildVerifier() error = %v", err)
			}
		})
	}
}

func TestBuildVerifierBoundsDownloadMetadata(t *testing.T) {
	process := func(_ context.Context, _ string, _ []string, _ string, _ map[string]string, stdout, _ io.Writer) error {
		_, err := io.WriteString(stdout, strings.Repeat("x", maximumDownloadMetadataSize+1))
		return err
	}
	if _, err := BuildVerifier(context.Background(), t.TempDir(), process); err == nil {
		t.Fatal("BuildVerifier() accepted oversized metadata")
	}
}

func TestBuildVerifierReportsBinaryReadFailure(t *testing.T) {
	source := t.TempDir()
	if err := os.WriteFile(filepath.Join(source, "go.mod"), []byte("module example\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	process := func(_ context.Context, name string, args []string, _ string, _ map[string]string, stdout, _ io.Writer) error {
		switch {
		case name == "go" && len(args) > 1 && args[0] == "mod":
			_, _ = io.WriteString(stdout, validDownloadMetadata(source))
		case name == "git":
			return nil
		case name == "go" && args[0] == "build":
			return os.Mkdir(args[4], 0o700)
		}
		return nil
	}
	if _, err := BuildVerifier(context.Background(), t.TempDir(), process); err == nil {
		t.Fatal("BuildVerifier() binary read error = nil")
	}
}

func TestHashVerifierAndBoundedBufferFailures(t *testing.T) {
	if _, err := hashVerifier(failingReader{}); err == nil {
		t.Fatal("hashVerifier() error = nil")
	}
	buffer := boundedBuffer{maximum: 3}
	if count, err := buffer.Write([]byte("abc")); err != nil || count != 3 || buffer.String() != "abc" {
		t.Fatalf("boundedBuffer.Write() = %d, %v, %q", count, err, buffer.String())
	}
	if _, err := buffer.Write([]byte("d")); err == nil {
		t.Fatal("boundedBuffer accepted oversized output")
	}
}

func TestRequireJSONEndDistinguishesMultipleValues(t *testing.T) {
	decoder := json.NewDecoder(strings.NewReader(`{} {}`))
	var first any
	_ = decoder.Decode(&first)
	err := requireJSONEnd(decoder)
	if err == nil || !strings.Contains(err.Error(), "multiple JSON values") {
		t.Fatalf("requireJSONEnd() = %v", err)
	}
	decoder = json.NewDecoder(strings.NewReader(`{} {`))
	_ = decoder.Decode(&first)
	err = requireJSONEnd(decoder)
	if err == nil || strings.Contains(err.Error(), "multiple JSON values") {
		t.Fatalf("requireJSONEnd(trailing) = %v", err)
	}
}

func verifierProcess(metadata, fail string, writeBinary bool) Process {
	return func(_ context.Context, name string, args []string, _ string, _ map[string]string, stdout, _ io.Writer) error {
		if name == "go" && len(args) > 1 && args[0] == "mod" {
			if fail == "download" {
				return errors.New("download failed")
			}
			_, _ = io.WriteString(stdout, metadata)
			return nil
		}
		if name == "git" {
			if fail == "git" {
				return errors.New("apply failed")
			}
			return nil
		}
		if name == "go" && args[0] == "build" {
			if fail == "build" {
				return errors.New("build failed")
			}
			if writeBinary {
				return os.WriteFile(args[4], []byte("binary"), 0o700)
			}
			return nil
		}
		return errors.New("unexpected process")
	}
}

func validDownloadMetadata(source string) string {
	return `{"Dir":"` + source + `","Sum":"` + gremlinsSum + `","GoModSum":"` + gremlinsGoModSum + `"}`
}
