package upgrade_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/faustbrian/go-library-tools/internal/upgrade"
)

const (
	oldSHA = "1111111111111111111111111111111111111111"
	newSHA = "2222222222222222222222222222222222222222"
	digest = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
)

func TestPlanAndApplyUpdateReleasePinsTogether(t *testing.T) {
	root := fixture(t, "schema_version: 1\ntool_version: v1.0.0\n", workflow(oldSHA, "v1.0.0"))
	request := upgrade.Request{Version: "v1.1.0", WorkflowSHA: newSHA, ChecksumsSHA256: digest}

	plan, err := upgrade.Plan(root, request)
	if err != nil {
		t.Fatal(err)
	}
	if !plan.Changed || len(plan.Files) != 2 || plan.Files[0].Path != ".github/workflows/ci.yml" || plan.Files[1].Path != ".golib.yaml" {
		t.Fatalf("Plan() = %#v", plan)
	}
	assertContains(t, read(t, filepath.Join(root, ".golib.yaml")), "tool_version: v1.0.0")

	applied, err := upgrade.Apply(root, request)
	if err != nil || !applied.Changed {
		t.Fatalf("Apply() = %#v, %v", applied, err)
	}
	configuration := read(t, filepath.Join(root, ".golib.yaml"))
	assertContains(t, configuration, "tool_version: v1.1.0")
	assertContains(t, configuration, "tool_checksums_sha256: "+digest)
	caller := read(t, filepath.Join(root, ".github", "workflows", "ci.yml"))
	assertContains(t, caller, "library-ci.yml@"+newSHA+" # v1.1.0")
	assertContains(t, caller, "tooling_sha: "+newSHA)

	current, err := upgrade.Plan(root, request)
	if err != nil || current.Changed {
		t.Fatalf("current Plan() = %#v, %v", current, err)
	}
}

func TestPlanReportsEitherPinFileChanging(t *testing.T) {
	request := upgrade.Request{Version: "v1.1.0", WorkflowSHA: newSHA, ChecksumsSHA256: digest}
	tests := []struct {
		name          string
		configuration string
		workflow      string
		changed       int
	}{
		{"workflow only", "tool_version: v1.1.0\ntool_checksums_sha256: " + digest + "\n", workflow(oldSHA, "v1.0.0"), 0},
		{"configuration only", "tool_version: v1.0.0\n", workflow(newSHA, "v1.1.0"), 1},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := upgrade.Plan(fixture(t, test.configuration, test.workflow), request)
			if err != nil || !result.Changed || !result.Files[test.changed].Changed || result.Files[1-test.changed].Changed {
				t.Fatalf("Plan() = %#v, %v", result, err)
			}
		})
	}
}

func TestApplyPreservesWorkflowContentOutsidePins(t *testing.T) {
	caller := workflow(oldSHA, "v1.0.0") + "# required deployment note\n"
	root := fixture(t, "tool_version: v1.0.0\n", caller)
	request := upgrade.Request{Version: "v1.1.0", WorkflowSHA: newSHA, ChecksumsSHA256: digest}
	if _, err := upgrade.Apply(root, request); err != nil {
		t.Fatal(err)
	}
	updated := read(t, filepath.Join(root, ".github", "workflows", "ci.yml"))
	assertContains(t, updated, "# required deployment note\n")
}

func TestPlanRejectsInvalidRequestsAndRepositoryPins(t *testing.T) {
	valid := upgrade.Request{Version: "v1.1.0", WorkflowSHA: newSHA, ChecksumsSHA256: digest}
	tests := map[string]struct {
		request       upgrade.Request
		configuration string
		workflow      string
	}{
		"version":               {upgrade.Request{Version: "latest", WorkflowSHA: newSHA, ChecksumsSHA256: digest}, "schema_version: 1\ntool_version: v1.0.0\n", workflow(oldSHA, "v1.0.0")},
		"workflow sha":          {upgrade.Request{Version: "v1.1.0", WorkflowSHA: "abc", ChecksumsSHA256: digest}, "schema_version: 1\ntool_version: v1.0.0\n", workflow(oldSHA, "v1.0.0")},
		"checksum":              {upgrade.Request{Version: "v1.1.0", WorkflowSHA: newSHA, ChecksumsSHA256: "ABC"}, "schema_version: 1\ntool_version: v1.0.0\n", workflow(oldSHA, "v1.0.0")},
		"missing version":       {valid, "schema_version: 1\n", workflow(oldSHA, "v1.0.0")},
		"duplicate version":     {valid, "tool_version: v1.0.0\ntool_version: v1.0.1\n", workflow(oldSHA, "v1.0.0")},
		"duplicate checksum":    {valid, "tool_version: v1.0.0\ntool_checksums_sha256: " + digest + "\ntool_checksums_sha256: " + digest + "\n", workflow(oldSHA, "v1.0.0")},
		"missing workflow":      {valid, "tool_version: v1.0.0\n", "name: CI\n"},
		"duplicate workflow":    {valid, "tool_version: v1.0.0\n", workflow(oldSHA, "v1.0.0") + "  other:\n    uses: faustbrian/go-library-tools/.github/workflows/library-ci.yml@" + oldSHA + "\n"},
		"mismatched sha":        {valid, "tool_version: v1.0.0\n", strings.Replace(workflow(oldSHA, "v1.0.0"), "tooling_sha: "+oldSHA, "tooling_sha: "+newSHA, 1)},
		"duplicate tooling sha": {valid, "tool_version: v1.0.0\n", workflow(oldSHA, "v1.0.0") + "      tooling_sha: " + oldSHA + "\n"},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			root := fixture(t, test.configuration, test.workflow)
			if _, err := upgrade.Plan(root, test.request); err == nil {
				t.Fatal("Plan() error = nil")
			}
		})
	}
}

func TestPlanRejectsUnsafeOrMissingFiles(t *testing.T) {
	request := upgrade.Request{Version: "v1.1.0", WorkflowSHA: newSHA, ChecksumsSHA256: digest}
	t.Run("configuration symlink", func(t *testing.T) {
		root := fixture(t, "tool_version: v1.0.0\n", workflow(oldSHA, "v1.0.0"))
		if err := os.Remove(filepath.Join(root, ".golib.yaml")); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink("go.mod", filepath.Join(root, ".golib.yaml")); err != nil {
			t.Fatal(err)
		}
		if _, err := upgrade.Plan(root, request); err == nil {
			t.Fatal("Plan() error = nil")
		}
	})
	t.Run("workflow missing", func(t *testing.T) {
		root := fixture(t, "tool_version: v1.0.0\n", workflow(oldSHA, "v1.0.0"))
		if err := os.Remove(filepath.Join(root, ".github", "workflows", "ci.yml")); err != nil {
			t.Fatal(err)
		}
		if _, err := upgrade.Plan(root, request); err == nil {
			t.Fatal("Plan() error = nil")
		}
	})
}

func fixture(t *testing.T, configuration, caller string) string {
	t.Helper()
	root := t.TempDir()
	write(t, filepath.Join(root, ".golib.yaml"), configuration)
	write(t, filepath.Join(root, ".github", "workflows", "ci.yml"), caller)
	write(t, filepath.Join(root, "go.mod"), "module example\n\ngo 1.27.0\n")
	return root
}

func workflow(sha, version string) string {
	return "jobs:\n  ci:\n    uses: faustbrian/go-library-tools/.github/workflows/library-ci.yml@" + sha + " # " + version + "\n    with:\n      tooling_sha: " + sha + "\n"
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

func read(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(content)
}

func assertContains(t *testing.T, value, fragment string) {
	t.Helper()
	if !strings.Contains(value, fragment) {
		t.Fatalf("%q does not contain %q", value, fragment)
	}
}
