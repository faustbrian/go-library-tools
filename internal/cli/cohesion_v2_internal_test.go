package cli

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/faustbrian/go-library-tools/internal/cohesion"
)

func TestParseCohesionV2InvocationAcceptsEveryCommandShape(t *testing.T) {
	tests := []struct {
		args   []string
		action string
		view   string
	}{
		{[]string{"catalog", "project", "consumer", "--schema-version", "2", "--mode", "preview", "--repository", "github.com/faustbrian/example", "--inputs", "sources.json", "--resolution-map", "map.json", "--output", "catalog.json"}, "project", "consumer"},
		{[]string{"catalog", "project", "recover", "engineering", "--schema-version", "2", "--mode", "final", "--repository", "github.com/faustbrian/example", "--inputs", "sources.json", "--resolution-map", "map.json", "--output", "catalog.json"}, "project-recover", "engineering"},
		{[]string{"aggregate", "generate", "--schema-version", "2", "--mode", "preview", "--inputs", "inputs.json", "--resolution-map", "map.json", "--output", "bundle"}, "aggregate-generate", ""},
		{[]string{"aggregate", "check", "--schema-version", "2", "--mode", "final", "--inputs", "inputs.json", "--resolution-map", "map.json", "--output", "bundle"}, "aggregate-check", ""},
		{[]string{"aggregate", "recover", "--schema-version", "2", "--mode", "final", "--inputs", "inputs.json", "--resolution-map", "map.json", "--output", "bundle"}, "aggregate-recover", ""},
		{[]string{"sources", "check", "--schema-version", "2", "--inputs", "sources.json"}, "sources-check", ""},
		{[]string{"sources", "verify", "--schema-version", "2", "--inputs", "sources.json", "--repository", "github.com/faustbrian/example", "--resolution-map", "map.json"}, "sources-verify", ""},
	}
	for _, test := range tests {
		got, err := parseCohesionV2Invocation(test.args)
		if err != nil {
			t.Fatalf("parseCohesionV2Invocation(%v) error = %v", test.args, err)
		}
		if got.action != test.action || got.view != test.view || got.schemaVersion != "2" {
			t.Fatalf("parseCohesionV2Invocation(%v) = %#v", test.args, got)
		}
	}
}

func TestParseCohesionV2InvocationRejectsIncompleteOrAmbiguousFlags(t *testing.T) {
	valid := []string{"catalog", "project", "consumer", "--schema-version", "2", "--mode", "preview", "--repository", "github.com/faustbrian/example", "--inputs", "sources.json", "--resolution-map", "map.json", "--output", "catalog.json"}
	tests := map[string][]string{
		"missing output":            valid[:len(valid)-2],
		"duplicate mode":            append(append([]string{}, valid...), "--mode", "final"),
		"unknown flag":              append(append([]string{}, valid...), "--other", "value"),
		"wrong schema":              append(append([]string{}, valid[:4]...), append([]string{"3"}, valid[5:]...)...),
		"wrong mode":                append(append([]string{}, valid[:6]...), append([]string{"other"}, valid[7:]...)...),
		"source check resolver":     []string{"sources", "check", "--schema-version", "2", "--inputs", "sources.json", "--resolution-map", "map.json"},
		"source verify no resolver": []string{"sources", "verify", "--schema-version", "2", "--inputs", "sources.json", "--repository", "github.com/faustbrian/example"},
	}
	for name, args := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := parseCohesionV2Invocation(args); err == nil {
				t.Fatalf("parseCohesionV2Invocation(%v) error = nil", args)
			}
		})
	}
	if _, err := parseCohesionV2Invocation([]string{"not-a-command"}); err == nil {
		t.Fatal("parseCohesionV2Invocation(unknown) error = nil")
	}
}

func TestParseCohesionV2InvocationMatchesCommandBoundariesAndViews(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"project boundary", []string{"catalog", "project", "consumer"}, "required flag"},
		{"recover boundary", []string{"catalog", "project", "recover", "engineering"}, "required flag"},
		{"aggregate boundary", []string{"aggregate", "generate"}, "required flag"},
		{"sources check boundary", []string{"sources", "check"}, "required flag"},
		{"sources verify boundary", []string{"sources", "verify"}, "required flag"},
		{"project unknown view", []string{"catalog", "project", "other"}, "unknown cohesion schema-v2 command"},
		{"recover unknown view", []string{"catalog", "project", "recover", "other"}, "unknown cohesion schema-v2 command"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := parseCohesionV2Invocation(test.args); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("parseCohesionV2Invocation(%v) error = %v, want %q", test.args, err, test.want)
			}
		})
	}
}

func TestParseCohesionV2InvocationRejectsCrossBoundaryMatches(t *testing.T) {
	validProjectFlags := []string{"--schema-version", "2", "--mode", "preview", "--repository", "r", "--inputs", "i", "--resolution-map", "m", "--output", "o"}
	validAggregateFlags := []string{"--schema-version", "2", "--mode", "preview", "--inputs", "i", "--resolution-map", "m", "--output", "o"}
	validSourcesCheckFlags := []string{"--schema-version", "2", "--inputs", "i"}
	validSourcesVerifyFlags := []string{"--schema-version", "2", "--inputs", "i", "--repository", "r", "--resolution-map", "m"}
	tests := [][]string{
		append([]string{"catalog", "wrong", "consumer"}, validProjectFlags...),
		append([]string{"catalog", "project", "wrong", "consumer"}, validProjectFlags...),
		append([]string{"catalog", "project", "recover", "wrong"}, validProjectFlags...),
		append([]string{"aggregate", "wrong"}, validAggregateFlags...),
		append([]string{"sources", "wrong"}, validSourcesCheckFlags...),
		append([]string{"sources", "wrong"}, validSourcesVerifyFlags...),
	}
	for _, args := range tests {
		if _, err := parseCohesionV2Invocation(args); err == nil || !strings.Contains(err.Error(), "unknown cohesion schema-v2 command") {
			t.Fatalf("parseCohesionV2Invocation(%v) error = %v", args, err)
		}
	}
}

func TestIsCohesionV2InvocationRecognizesOnlyV2Markers(t *testing.T) {
	for _, test := range []struct {
		args []string
		want bool
	}{
		{[]string{"catalog", "project"}, true},
		{[]string{"cohesion", "sources", "--schema-version", "2"}, true},
		{[]string{"catalog", "list"}, false},
		{nil, false},
	} {
		if got := isCohesionV2Invocation(test.args); got != test.want {
			t.Errorf("isCohesionV2Invocation(%v) = %t, want %t", test.args, got, test.want)
		}
	}
}

func TestExecuteCohesionV2ReportsResolutionAndUnsupportedOperations(t *testing.T) {
	for _, test := range []struct {
		name           string
		args           []string
		wantCode       int
		wantDiagnostic string
	}{
		{"unsupported", []string{"aggregate", "generate", "--schema-version", "2", "--mode", "preview", "--inputs", "missing", "--resolution-map", "missing", "--output", "out"}, 1, "unsupported"},
		{"resolution failure", []string{"sources", "check", "--schema-version", "2", "--inputs", "missing"}, 1, "resolution-failed"},
	} {
		t.Run(test.name, func(t *testing.T) {
			var stderr bytes.Buffer
			if got := executeCohesionV2(test.args, t.TempDir(), &bytes.Buffer{}, &stderr); got != test.wantCode || !strings.Contains(stderr.String(), test.wantDiagnostic) {
				t.Fatalf("executeCohesionV2() = %d, stderr %q", got, stderr.String())
			}
		})
	}
}

func TestExecuteCohesionV2DispatchesSuccessfulSourceOperations(t *testing.T) {
	oldCheck, oldVerify := checkSourcesV2, verifySourcesV2
	t.Cleanup(func() { checkSourcesV2, verifySourcesV2 = oldCheck, oldVerify })
	var checked, verified bool
	checkSourcesV2 = func(path string) error { checked = path != ""; return nil }
	verifySourcesV2 = func(inputs, repository, resolution string) error {
		verified = inputs != "" && repository != "" && resolution != ""
		return nil
	}
	root := t.TempDir()
	if got := executeCohesionV2([]string{"sources", "check", "--schema-version", "2", "--inputs", filepath.Join(root, "inputs.json")}, root, &bytes.Buffer{}, &bytes.Buffer{}); got != 0 || !checked {
		t.Fatalf("successful sources check = %d, checked=%t", got, checked)
	}
	if got := executeCohesionV2([]string{"sources", "verify", "--schema-version", "2", "--inputs", filepath.Join(root, "inputs.json"), "--repository", "example", "--resolution-map", filepath.Join(root, "map.json")}, root, &bytes.Buffer{}, &bytes.Buffer{}); got != 0 || !verified {
		t.Fatalf("successful sources verify = %d, verified=%t", got, verified)
	}
}

func TestExecuteCohesionV2PreservesValidatorDiagnostics(t *testing.T) {
	oldCheck, oldVerify := checkSourcesV2, verifySourcesV2
	t.Cleanup(func() { checkSourcesV2, verifySourcesV2 = oldCheck, oldVerify })
	checkSourcesV2 = func(path string) error { return cohesion.CheckSourcesV2(path) }
	verifySourcesV2 = func(string, string, string) error { return errors.New("generic verify failure") }
	input := filepath.Join(t.TempDir(), "invalid.json")
	if err := os.WriteFile(input, []byte(`{}`), 0o600); err != nil {
		t.Fatal(err)
	}
	var stderr bytes.Buffer
	if got := executeCohesionV2([]string{"sources", "check", "--schema-version", "2", "--inputs", input}, t.TempDir(), &bytes.Buffer{}, &stderr); got != 1 || !strings.Contains(stderr.String(), "schema-required-member") {
		t.Fatalf("diagnostic validator error = %d, %q", got, stderr.String())
	}
	verifySourcesV2 = func(inputs, _, _ string) error { return cohesion.CheckSourcesV2(inputs) }
	stderr.Reset()
	if got := executeCohesionV2([]string{"sources", "verify", "--schema-version", "2", "--inputs", input, "--repository", "r", "--resolution-map", "m"}, t.TempDir(), &bytes.Buffer{}, &stderr); got != 1 || !strings.Contains(stderr.String(), "schema-required-member") {
		t.Fatalf("diagnostic verify error = %d, %q", got, stderr.String())
	}
	verifySourcesV2 = func(string, string, string) error { return errors.New("generic verify failure") }
	stderr.Reset()
	if got := executeCohesionV2([]string{"sources", "verify", "--schema-version", "2", "--inputs", "x", "--repository", "r", "--resolution-map", "m"}, t.TempDir(), &bytes.Buffer{}, &stderr); got != 1 || !strings.Contains(stderr.String(), "resolution-failed") {
		t.Fatalf("generic validator error = %d, %q", got, stderr.String())
	}
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errors.New("write failed") }

func TestWriteCohesionV2DiagnosticHandlesWriterFailure(t *testing.T) {
	if got := writeCohesionV2Diagnostic(failingWriter{}, 7, structDiagnostic("x")); got != 7 {
		t.Fatalf("writeCohesionV2Diagnostic() = %d, want 7", got)
	}
}

func TestParseSingletonFlagsRejectsOddAndEmptyValues(t *testing.T) {
	allowed := flagSet("--known")
	if _, err := parseSingletonFlags([]string{"--known"}, allowed, nil); err == nil {
		t.Fatal("parseSingletonFlags(odd) error = nil")
	}
	if _, err := parseSingletonFlags([]string{"--known", ""}, allowed, nil); err == nil {
		t.Fatal("parseSingletonFlags(empty) error = nil")
	}
}

func TestFindRootRejectsRelativeAndMissingRepositories(t *testing.T) {
	if _, err := findRoot("relative"); err == nil {
		t.Fatal("findRoot(relative) error = nil")
	}
	missing := t.TempDir()
	if _, err := findRoot(missing); err == nil || !strings.Contains(err.Error(), ".golib.yaml not found") {
		t.Fatalf("findRoot(missing) error = %v", err)
	}
	if err := os.Mkdir(filepath.Join(missing, ".golib.yaml"), 0o700); err != nil {
		t.Fatal(err)
	}
	if _, err := findRoot(missing); err == nil || !strings.Contains(err.Error(), ".golib.yaml not found") {
		t.Fatalf("findRoot(directory marker) error = %v", err)
	}
}

func TestFindRootReportsCanonicalizationFailure(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".golib.yaml"), []byte("schema_version: 1\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	old := evalRootSymlinks
	evalRootSymlinks = func(string) (string, error) { return "", errors.New("symlink failure") }
	t.Cleanup(func() { evalRootSymlinks = old })
	if _, err := findRoot(root); err == nil || !strings.Contains(err.Error(), "canonicalize root") {
		t.Fatalf("findRoot(canonicalization failure) error = %v", err)
	}
}

func structDiagnostic(code string) (d cohesion.DiagnosticV3) {
	d.Code = code
	d.Message = "message"
	return
}
