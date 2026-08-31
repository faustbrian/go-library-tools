package cli_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/faustbrian/go-library-tools/internal/cli"
)

func TestExecuteShowsHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := cli.Execute([]string{"--help"}, t.TempDir(), &stdout, &stderr)
	if code != 0 || stderr.Len() != 0 {
		t.Fatalf("Execute() code = %d, stderr = %q", code, stderr.String())
	}
	for _, command := range []string{"check", "config validate", "config show --json", "inventory", "consumers validate", "repository check", "specification check [--online]", "workflows check", "services cycle", "release dry-run", "upgrade <plan|apply>"} {
		if !strings.Contains(stdout.String(), command) {
			t.Errorf("help does not contain %q", command)
		}
	}
	for _, unsupported := range []string{"services start", "services stop"} {
		if strings.Contains(stdout.String(), unsupported) {
			t.Errorf("help advertises unsupported command %q", unsupported)
		}
	}
}

func TestExecuteCyclesModuleServicesWithoutDetachedState(t *testing.T) {
	root := fixture(t)
	for _, args := range [][]string{
		{"services", "cycle"},
		{"services", "cycle", "--module", "."},
	} {
		var stdout, stderr bytes.Buffer
		code := cli.Execute(args, root, &stdout, &stderr)
		if code != 0 || stderr.Len() != 0 || !strings.Contains(stdout.String(), "services: not applicable") {
			t.Fatalf("Execute(%v) = %d, %q, %q", args, code, stdout.String(), stderr.String())
		}
	}
}

func TestExecuteShowsVersionWithoutRepository(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := cli.Execute([]string{"--version"}, t.TempDir(), &stdout, &stderr)
	if code != 0 || stderr.Len() != 0 || stdout.String() != "dev\n" {
		t.Fatalf("Execute() code = %d, stdout = %q, stderr = %q", code, stdout.String(), stderr.String())
	}
}

func TestExecuteValidatesConfigurationFromNestedDirectory(t *testing.T) {
	root := fixture(t)
	nested := filepath.Join(root, "nested", "directory")
	if err := os.MkdirAll(nested, 0o700); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	code := cli.Execute([]string{"config", "validate"}, nested, &stdout, &stderr)
	if code != 0 || stderr.Len() != 0 {
		t.Fatalf("Execute() code = %d, stderr = %q", code, stderr.String())
	}
	if got := stdout.String(); got != "configuration valid: github.com/faustbrian/example\n" {
		t.Fatalf("Execute() stdout = %q", got)
	}
}

func TestExecuteShowsMachineReadableConfiguration(t *testing.T) {
	root := fixture(t)
	write(t, filepath.Join(root, ".golib.yaml"), "schema_version: 1\ntool_version: v1.0.0\nruntimes:\n  deno: 2.9.4\n")
	var stdout, stderr bytes.Buffer
	code := cli.Execute([]string{"config", "show", "--json"}, root, &stdout, &stderr)
	if code != 0 || stderr.Len() != 0 {
		t.Fatalf("Execute() code = %d, stderr = %q", code, stderr.String())
	}
	var got map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	runtimes, ok := got["runtimes"].(map[string]any)
	if !ok || runtimes["deno"] != "2.9.4" {
		t.Fatalf("configuration = %#v", got)
	}
}

func TestExecutePrintsDeterministicJSONInventory(t *testing.T) {
	root := fixture(t)
	var stdout, stderr bytes.Buffer
	code := cli.Execute([]string{"inventory", "--json"}, root, &stdout, &stderr)
	if code != 0 || stderr.Len() != 0 {
		t.Fatalf("Execute() code = %d, stderr = %q", code, stderr.String())
	}
	var got map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("stdout is not JSON: %v", err)
	}
	if got["repository"] != "github.com/faustbrian/example" {
		t.Fatalf("inventory repository = %#v", got["repository"])
	}
}

func TestExecutePrintsHumanInventory(t *testing.T) {
	root := fixture(t)
	var stdout, stderr bytes.Buffer
	code := cli.Execute([]string{"inventory"}, root, &stdout, &stderr)
	if code != 0 || stderr.Len() != 0 || stdout.String() != "github.com/faustbrian/example: 1 module(s)\n" {
		t.Fatalf("Execute() code = %d, stdout = %q, stderr = %q", code, stdout.String(), stderr.String())
	}
}

func TestExecuteValidatesConsumerInventory(t *testing.T) {
	root := fixture(t)
	writeConsumerInventory(t, root)
	for _, test := range []struct {
		args []string
		want string
	}{
		{[]string{"consumers", "validate"}, "2 total, 1 active, 0 deferred, 1 tooling"},
		{[]string{"consumers", "validate", "--json"}, `"active": 1`},
	} {
		var stdout, stderr bytes.Buffer
		code := cli.Execute(test.args, root, &stdout, &stderr)
		if code != 0 || stderr.Len() != 0 || !strings.Contains(stdout.String(), test.want) {
			t.Fatalf("Execute(%v) = %d, %q, %q", test.args, code, stdout.String(), stderr.String())
		}
	}
}

func TestExecuteChecksSpecificationDecisions(t *testing.T) {
	root := fixture(t)
	for _, args := range [][]string{{"specification", "check"}, {"specification", "check", "--online"}} {
		var stdout, stderr bytes.Buffer
		if code := cli.Execute(args, root, &stdout, &stderr); code != 0 || stderr.Len() != 0 || !strings.Contains(stdout.String(), "0 module(s), 0 decision(s)") {
			t.Fatalf("Execute(%v) = %d, %q, %q", args, code, stdout.String(), stderr.String())
		}
	}

	writeSpecificationFixture(t, root, "resolved")
	var stdout, stderr bytes.Buffer
	if code := cli.Execute([]string{"specification", "check"}, root, &stdout, &stderr); code != 0 || stderr.Len() != 0 || !strings.Contains(stdout.String(), "1 module(s), 1 decision(s)") {
		t.Fatalf("Execute(specification check) = %d, %q, %q", code, stdout.String(), stderr.String())
	}
	writeSpecificationFixture(t, root, "unresolved")
	stdout.Reset()
	if code := cli.Execute([]string{"specification", "check"}, root, &stdout, &stderr); code != 1 || !strings.Contains(stderr.String(), "1 unresolved") {
		t.Fatalf("Execute(unresolved specification) = %d, %q", code, stderr.String())
	}
	for _, args := range [][]string{{"specification"}, {"specification", "lint"}, {"specification", "check", "--offline"}, {"specification", "check", "--online", "extra"}} {
		stderr.Reset()
		if code := cli.Execute(args, root, &stdout, &stderr); code != 2 || !strings.Contains(stderr.String(), "usage: golib specification check") {
			t.Fatalf("Execute(%v) = %d, %q", args, code, stderr.String())
		}
	}
}

func TestExecutePlansAndAppliesToolingUpgrade(t *testing.T) {
	root := fixture(t)
	oldSHA := strings.Repeat("1", 40)
	newSHA := strings.Repeat("2", 40)
	digest := strings.Repeat("a", 64)
	writeUpgradeWorkflow(t, root, oldSHA)
	arguments := []string{"--version", "v1.1.0", "--workflow-sha", newSHA, "--checksums-sha256", digest}

	var stdout, stderr bytes.Buffer
	code := cli.Execute(append([]string{"upgrade", "plan"}, arguments...), root, &stdout, &stderr)
	if code != 0 || stderr.Len() != 0 || !strings.Contains(stdout.String(), "tooling pins planned") {
		t.Fatalf("plan = %d, %q, %q", code, stdout.String(), stderr.String())
	}
	stdout.Reset()
	code = cli.Execute(append([]string{"upgrade", "plan", "--json"}, arguments...), root, &stdout, &stderr)
	if code != 0 || stderr.Len() != 0 || !json.Valid(stdout.Bytes()) {
		t.Fatalf("JSON plan = %d, %q, %q", code, stdout.String(), stderr.String())
	}
	stdout.Reset()
	code = cli.Execute(append(append([]string{"upgrade", "apply"}, arguments...), "--json"), root, &stdout, &stderr)
	if code != 0 || stderr.Len() != 0 {
		t.Fatalf("apply = %d, %q, %q", code, stdout.String(), stderr.String())
	}
	var result struct {
		Changed bool `json:"changed"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil || !result.Changed {
		t.Fatalf("upgrade JSON = %#v, %v", result, err)
	}
	assertFileContains(t, filepath.Join(root, ".golib.yaml"), "tool_checksums_sha256: "+digest)
}

func TestExecuteRejectsMalformedUpgradeArguments(t *testing.T) {
	valid := []string{"--version", "v1.1.0", "--workflow-sha", strings.Repeat("2", 40), "--checksums-sha256", strings.Repeat("a", 64)}
	tests := []struct {
		args []string
		want string
	}{
		{[]string{"upgrade"}, "action is required"},
		{[]string{"upgrade", "replace"}, "action must be plan or apply"},
		{[]string{"upgrade", "plan", "--json", "--json"}, "JSON flag is duplicated"},
		{[]string{"upgrade", "plan", "--version"}, "flag has no value"},
		{append([]string{"upgrade", "plan", "--unknown", "value"}, valid...), "arguments are malformed"},
		{append([]string{"upgrade", "plan", "--version", ""}, valid[2:]...), "arguments are malformed"},
		{append([]string{"upgrade", "plan", "--version", "v1.0.0", "--version", "v1.1.0"}, valid[2:]...), "argument is duplicated"},
	}
	for _, test := range tests {
		var stdout, stderr bytes.Buffer
		if code := cli.Execute(test.args, fixture(t), &stdout, &stderr); code != 2 || !strings.Contains(stderr.String(), "usage:") || !strings.Contains(stderr.String(), test.want) {
			t.Fatalf("Execute(%v) = %d, %q", test.args, code, stderr.String())
		}
	}
}

func TestExecuteReportsUpgradeFailures(t *testing.T) {
	oldSHA := strings.Repeat("1", 40)
	arguments := make([]string, 0, 9)
	arguments = append(arguments, "upgrade", "plan", "--version", "v1.1.0", "--workflow-sha", strings.Repeat("2", 40), "--checksums-sha256", strings.Repeat("a", 64))

	var stdout, stderr bytes.Buffer
	if code := cli.Execute(arguments, t.TempDir(), &stdout, &stderr); code != 1 || !strings.Contains(stderr.String(), "locate repository root") {
		t.Fatalf("missing root = %d, %q", code, stderr.String())
	}

	root := fixture(t)
	stderr.Reset()
	if code := cli.Execute(arguments, root, &stdout, &stderr); code != 1 || !strings.Contains(stderr.String(), "read CI workflow") {
		t.Fatalf("invalid repository = %d, %q", code, stderr.String())
	}

	writeUpgradeWorkflow(t, root, oldSHA)
	stderr.Reset()
	if code := cli.Execute(arguments, root, failingWriter{}, &stderr); code != 1 || !strings.Contains(stderr.String(), "write upgrade result") {
		t.Fatalf("human output = %d, %q", code, stderr.String())
	}
	stderr.Reset()
	if code := cli.Execute(append(arguments, "--json"), root, failingWriter{}, &stderr); code != 1 || !strings.Contains(stderr.String(), "write upgrade result") {
		t.Fatalf("JSON output = %d, %q", code, stderr.String())
	}
}

func TestExecuteChecksStandaloneRepositoryContract(t *testing.T) {
	root := fixture(t)
	write(t, filepath.Join(root, "go.mod"), "module github.com/faustbrian/example\n\ngo 1.27.0\n")
	var stdout, stderr bytes.Buffer
	code := cli.Execute([]string{"repository", "check"}, root, &stdout, &stderr)
	if code != 0 || stderr.Len() != 0 || stdout.String() != "standalone repository contract passed\n" {
		t.Fatalf("Execute() = %d, %q, %q", code, stdout.String(), stderr.String())
	}
}

func TestExecuteInspectsEvidence(t *testing.T) {
	root := fixture(t)
	for _, args := range [][]string{{"evidence", "inspect"}, {"evidence", "inspect", "--json"}} {
		var stdout, stderr bytes.Buffer
		code := cli.Execute(args, root, &stdout, &stderr)
		if code != 0 || stderr.Len() != 0 || stdout.Len() == 0 {
			t.Fatalf("Execute(%v) = %d, %q, %q", args, code, stdout.String(), stderr.String())
		}
	}
}

func TestExecuteRunsCheckWithDisposableGoEnvironment(t *testing.T) {
	root := fixture(t)
	write(t, filepath.Join(root, "go.mod"), "module github.com/faustbrian/example\n\ngo 1.27.0\n")
	write(t, filepath.Join(root, "example.go"), "package example\n\nfunc Value() int { return 1 }\n")
	write(t, filepath.Join(root, "modules.json"), `{"schema_version":1,"repository":"github.com/faustbrian/example","go_version":"1.27.0","modules":[{"directory":".","module_path":"github.com/faustbrian/example","go_version":"1.27.0","kind":"public","releasable":true,"gates":{},"packages":[]}]}`)
	var stdout, stderr bytes.Buffer
	code := cli.Execute([]string{"check", "--module", "."}, root, &stdout, &stderr)
	if code != 0 || stderr.Len() != 0 {
		t.Fatalf("Execute() code = %d, stdout = %q, stderr = %q", code, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "[.] safety\n") {
		t.Fatalf("check output = %q", stdout.String())
	}
}

func TestExecuteCheckDefaultsToAllModules(t *testing.T) {
	root := fixture(t)
	write(t, filepath.Join(root, "go.mod"), "module github.com/faustbrian/example\n\ngo 1.27.0\n")
	write(t, filepath.Join(root, "example.go"), "package example\n")
	write(t, filepath.Join(root, "modules.json"), `{"schema_version":1,"repository":"github.com/faustbrian/example","go_version":"1.27.0","modules":[{"directory":".","module_path":"github.com/faustbrian/example","go_version":"1.27.0","kind":"public","releasable":true,"gates":{},"packages":[]}]}`)
	var stdout, stderr bytes.Buffer
	code := cli.Execute([]string{"check"}, root, &stdout, &stderr)
	if code != 0 || stderr.Len() != 0 {
		t.Fatalf("Execute() code = %d, stderr = %q", code, stderr.String())
	}
}

func TestExecuteReportsOutputFailure(t *testing.T) {
	root := fixture(t)
	var stderr bytes.Buffer
	code := cli.Execute([]string{"config", "show", "--json"}, root, failingWriter{}, &stderr)
	if code != 1 || !strings.Contains(stderr.String(), "write configuration") {
		t.Fatalf("Execute() config code = %d, stderr = %q", code, stderr.String())
	}
	stderr.Reset()
	code = cli.Execute([]string{"inventory", "--json"}, root, failingWriter{}, &stderr)
	if code != 1 || !strings.Contains(stderr.String(), "write inventory") {
		t.Fatalf("Execute() code = %d, stderr = %q", code, stderr.String())
	}
	stderr.Reset()
	code = cli.Execute([]string{"evidence", "inspect", "--json"}, root, failingWriter{}, &stderr)
	if code != 1 || !strings.Contains(stderr.String(), "write evidence inventory") {
		t.Fatalf("Execute() evidence code = %d, stderr = %q", code, stderr.String())
	}
	writeConsumerInventory(t, root)
	stderr.Reset()
	code = cli.Execute([]string{"consumers", "validate", "--json"}, root, failingWriter{}, &stderr)
	if code != 1 || !strings.Contains(stderr.String(), "write consumer inventory") {
		t.Fatalf("Execute() consumer JSON code = %d, stderr = %q", code, stderr.String())
	}
	stderr.Reset()
	code = cli.Execute([]string{"consumers", "validate"}, root, failingWriter{}, &stderr)
	if code != 1 || !strings.Contains(stderr.String(), "write consumer inventory") {
		t.Fatalf("Execute() consumer code = %d, stderr = %q", code, stderr.String())
	}
}

func TestExecuteReportsUsageAndRepositoryErrors(t *testing.T) {
	tests := []struct {
		name string
		args []string
		root func(*testing.T) string
		code int
		want string
	}{
		{"missing command", nil, fixture, 2, "command is required"},
		{"unknown command", []string{"wat"}, fixture, 2, "unknown command"},
		{"invalid config arguments", []string{"config"}, fixture, 2, "usage:"},
		{"invalid config show arguments", []string{"config", "show"}, fixture, 2, "usage:"},
		{"invalid inventory arguments", []string{"inventory", "--yaml"}, fixture, 2, "usage:"},
		{"missing consumer action", []string{"consumers"}, fixture, 2, "usage:"},
		{"unknown consumer action", []string{"consumers", "show"}, fixture, 2, "usage:"},
		{"invalid consumer arguments", []string{"consumers", "validate", "--yaml"}, fixture, 2, "usage:"},
		{"extra consumer arguments", []string{"consumers", "validate", "--json", "extra"}, fixture, 2, "usage:"},
		{"missing consumer inventory", []string{"consumers", "validate"}, fixture, 1, "read consumer manifest"},
		{"invalid upgrade arguments", []string{"upgrade", "plan", "--version", "v1.0.0"}, fixture, 2, "usage:"},
		{"invalid repository arguments", []string{"repository"}, fixture, 2, "usage:"},
		{"missing services action", []string{"services"}, fixture, 2, "usage:"},
		{"invalid services arguments", []string{"services", "start"}, fixture, 2, "usage:"},
		{"invalid services module", []string{"services", "cycle", "--module"}, fixture, 2, "usage:"},
		{"missing evidence action", []string{"evidence"}, fixture, 2, "usage:"},
		{"invalid evidence arguments", []string{"evidence", "inspect", "--yaml"}, fixture, 2, "usage:"},
		{"unsafe evidence", []string{"evidence", "inspect"}, func(t *testing.T) string {
			root := fixture(t)
			target := filepath.Join(root, "target")
			if err := os.Mkdir(target, 0o700); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(target, filepath.Join(root, ".verification")); err != nil {
				t.Fatal(err)
			}
			return root
		}, 1, "evidence root"},
		{"invalid repository contract", []string{"repository", "check"}, func(t *testing.T) string {
			root := fixture(t)
			write(t, filepath.Join(root, ".go-version"), "1.26.0\n")
			return root
		}, 1, ".go-version"},
		{"invalid check arguments", []string{"check", "--module"}, fixture, 2, "usage:"},
		{"unknown check module", []string{"check", "--module", "missing"}, fixture, 1, "unknown module"},
		{"missing root", []string{"inventory"}, func(t *testing.T) string { return t.TempDir() }, 1, "locate repository root"},
		{"relative root", []string{"inventory"}, func(*testing.T) string { return "." }, 1, "must be absolute"},
		{"invalid root path", []string{"inventory"}, func(t *testing.T) string {
			root := t.TempDir()
			file := filepath.Join(root, "file")
			write(t, file, "not a directory")
			return file
		}, 1, "locate repository root"},
		{"unreadable root marker", []string{"inventory"}, func(t *testing.T) string {
			root := t.TempDir()
			marker := filepath.Join(root, ".golib.yaml")
			if err := os.Symlink(marker, marker); err != nil {
				t.Fatal(err)
			}
			return root
		}, 1, "locate repository root"},
		{"directory root marker", []string{"inventory"}, func(t *testing.T) string {
			root := t.TempDir()
			if err := os.Mkdir(filepath.Join(root, ".golib.yaml"), 0o700); err != nil {
				t.Fatal(err)
			}
			return root
		}, 1, ".golib.yaml not found"},
		{"invalid config", []string{"inventory"}, func(t *testing.T) string {
			root := fixture(t)
			write(t, filepath.Join(root, ".golib.yaml"), "schema_version: 2\ntool_version: v1.0.0\n")
			return root
		}, 1, "schema_version must be 1"},
		{"invalid manifest", []string{"inventory"}, func(t *testing.T) string {
			root := fixture(t)
			write(t, filepath.Join(root, "modules.json"), "{")
			return root
		}, 1, "load module manifest"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := cli.Execute(test.args, test.root(t), &stdout, &stderr)
			if code != test.code || !strings.Contains(stderr.String(), test.want) {
				t.Fatalf("Execute() code = %d, stderr = %q", code, stderr.String())
			}
		})
	}
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) {
	return 0, errors.New("write failed")
}

func fixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	write(t, filepath.Join(root, ".golib.yaml"), "schema_version: 1\ntool_version: v1.0.0\n")
	write(t, filepath.Join(root, ".go-version"), "1.27.0\n")
	write(t, filepath.Join(root, "modules.json"), `{"schema_version":1,"repository":"github.com/faustbrian/example","go_version":"1.27.0","modules":[{"directory":".","module_path":"github.com/faustbrian/example","go_version":"1.27.0","kind":"public","releasable":true,"gates":{"tests":true},"packages":[]}]}`)
	write(t, filepath.Join(root, "packages.json"), `{"schema_version":1,"repository":"github.com/faustbrian/example","packages":[]}`)
	return root
}

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func assertFileContains(t *testing.T, path, fragment string) {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), fragment) {
		t.Fatalf("%s does not contain %q", path, fragment)
	}
}

func writeUpgradeWorkflow(t *testing.T, root, sha string) {
	t.Helper()
	write(t, filepath.Join(root, ".github", "workflows", "ci.yml"), "jobs:\n  ci:\n    uses: faustbrian/go-library-tools/.github/workflows/library-ci.yml@"+sha+" # v1.0.0\n    with:\n      tooling_sha: "+sha+"\n")
}

func writeConsumerInventory(t *testing.T, root string) {
	t.Helper()
	write(t, filepath.Join(root, "consumers.json"), `{"schema_version":1,"owner":"faustbrian","repositories":[{"name":"go-library","classification":"active","default_branch":"main"},{"name":"go-tooling","classification":"tooling","default_branch":"main","reason":"tooling"}]}`)
}

func writeSpecificationFixture(t *testing.T, root, status string) {
	t.Helper()
	write(t, filepath.Join(root, "modules.json"), `{"schema_version":1,"repository":"github.com/faustbrian/example","go_version":"1.27.0","modules":[{"directory":".","module_path":"github.com/faustbrian/example","go_version":"1.27.0","kind":"public","releasable":true,"gates":{"tests":true},"specifications":["RFC 9110"],"provenance":["specification/manifest.tsv"],"packages":[]}]}`)
	write(t, filepath.Join(root, "specification/manifest.tsv"), "id\tversion\turl\tsha256\tstatus\nrfc9110\tRFC-9110\thttps://www.rfc-editor.org/rfc/rfc9110.txt\t21c1cdce6ab0e5509b04d84a28000836c7a087cf786efe6f04877ebfff47232a\tpinned\n")
	write(t, filepath.Join(root, "specification/monitoring.json"), `{"schema_version":1,"reviewed_at":"`+time.Now().UTC().Format("2006-01-02")+`","review_interval_days":90,"authorities":[{"id":"source","kind":"specification","version":"RFC 9110","url":"https://example.com/specification","sha256":"21c1cdce6ab0e5509b04d84a28000836c7a087cf786efe6f04877ebfff47232a","specifications":["RFC 9110"]},{"id":"errata","kind":"errata","url":"https://example.com/errata","sha256":"21c1cdce6ab0e5509b04d84a28000836c7a087cf786efe6f04877ebfff47232a","specifications":["RFC 9110"]}]}`)
	write(t, filepath.Join(root, "specification/README.md"), "# Matrix\n\nEXAMPLE-DEC-001 `TestContract`\n")
	write(t, filepath.Join(root, "docs/specification-decisions.md"), "# Decisions\n\n## EXAMPLE-DEC-001: Contract\n\n"+
		"status "+status+" owner classification source and issue interpretations and peer behavior selected behavior security resource compatibility wire evidence public surface upstream reconsider "+
		"[RFC 9110 section 7.1](https://www.rfc-editor.org/rfc/rfc9110.html#section-7.1) `TestContract`\n")
	write(t, filepath.Join(root, "specification/decisions.json"), `{"schema_version":1,"decisions":[{"id":"EXAMPLE-DEC-001","title":"Contract","status":"`+status+`","owner":"maintainers","classification":"omission","decision_scope":"defensive","specification":"RFC 9110","version":"RFC 9110","source_authority":"source","section":"7.1","requirement_strength":"not specified","issue":"The contract is omitted.","interpretations":["Accept","Reject"],"peer_behavior":"Peers disagree.","selected_behavior":"Accept.","rationale":"Compatibility.","security_consequences":"Bounded.","resource_consequences":"Bounded.","compatibility_consequences":"Stable.","wire_consequences":"Stable.","executable_evidence":["TestContract"],"fixture_evidence":["testdata/contract.json"],"fuzz_evidence":["FuzzContract"],"interoperability_evidence":["testdata/peer.tsv"],"public_apis":["Contract"],"documentation":["docs/specification-decisions.md"],"upstream_status":"No issue.","reconsider_when":"The source changes."}]}`)
	writeSpecificationChangeControl(t, root)
	write(t, filepath.Join(root, "specification/conformance.json"), `{"schema_version":1,"decisions":[{"id":"EXAMPLE-DEC-001","authoritative_sources":["source"],"executable_evidence":["TestContract"],"fixtures":["testdata/contract.json"],"fuzz":["FuzzContract"],"differential_evidence":["testdata/peer.tsv"],"differential_classification":"specification ambiguity","public_behavior":["Contract behavior."]}]}`)
	write(t, filepath.Join(root, "testdata/contract.json"), "{}\n")
	write(t, filepath.Join(root, "testdata/peer.tsv"), "peer\tbehavior\nexample\taccept\n")
	write(t, filepath.Join(root, "contract_test.go"), "package example\n\nimport \"testing\"\n\nfunc TestContract(t *testing.T) {}\nfunc FuzzContract(f *testing.F) {}\n")
	finalizeSpecificationFixture(t, root, "https://example.com/specification")
}

func finalizeSpecificationFixture(t *testing.T, root, authorityURL string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, "specification/decisions.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Decisions []map[string]any `json:"decisions"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	registerPath := filepath.Join(root, "docs/specification-decisions.md")
	register, err := os.ReadFile(registerPath)
	if err != nil {
		t.Fatal(err)
	}
	write(t, registerPath, string(register)+"\n"+string(data)+"\n"+authorityURL+"\n")
	var changelog strings.Builder
	changelog.WriteString("# Changelog\n\n")
	history := map[string]any{"schema_version": 1, "entries": []any{}}
	for _, item := range manifest.Decisions {
		encoded, _ := json.Marshal(item)
		sum := sha256.Sum256(encoded)
		_, _ = fmt.Fprintf(&changelog, "- %s sha256:%x\n", item["id"], sum)
		entries, ok := history["entries"].([]any)
		if !ok {
			t.Fatal("decision history entries are not an array")
		}
		history["entries"] = append(entries, map[string]any{"id": item["id"], "digests": []string{hex.EncodeToString(sum[:])}})
	}
	write(t, filepath.Join(root, "CHANGELOG.md"), changelog.String()+"\n[Decisions](docs/specification-decisions.md)\n")
	historyData, _ := json.Marshal(history)
	write(t, filepath.Join(root, "specification/decision-history.json"), string(historyData))
}

func writeSpecificationChangeControl(t *testing.T, root string) {
	t.Helper()
	decisionPath := filepath.Join(root, "specification/decisions.json")
	data, err := os.ReadFile(decisionPath)
	if err != nil {
		t.Fatal(err)
	}
	control := `"change_control":{"readme":"README.md","conformance":"specification/README.md","compatibility":"COMPATIBILITY.md","contribution":"CONTRIBUTING.md","changelog":"CHANGELOG.md","pull_request_template":".github/pull_request_template.md"},`
	write(t, decisionPath, strings.Replace(string(data), `"decisions":`, control+`"decisions":`, 1))
	for path, link := range map[string]string{
		"README.md":        "[Specification decisions](docs/specification-decisions.md)\n",
		"COMPATIBILITY.md": "[Specification decisions](docs/specification-decisions.md)\n",
		"CONTRIBUTING.md":  "[Specification decisions](docs/specification-decisions.md)\n",
		"CHANGELOG.md":     "## Specification Decisions\n\nEXAMPLE-DEC-001: [Decision register](docs/specification-decisions.md)\n",
	} {
		existing, _ := os.ReadFile(filepath.Join(root, path))
		write(t, filepath.Join(root, path), string(existing)+"\n"+link)
	}
	matrixPath := filepath.Join(root, "specification/README.md")
	matrix, err := os.ReadFile(matrixPath)
	if err != nil {
		t.Fatal(err)
	}
	write(t, matrixPath, string(matrix)+"\n[Specification decisions](../docs/specification-decisions.md)\n")
	write(t, filepath.Join(root, ".github/pull_request_template.md"), "## Specification Decisions\n\n- Decision identifier\n- Compatibility impact\n- Changelog entry\n- Superseded identifiers remain in the register\n")
}
