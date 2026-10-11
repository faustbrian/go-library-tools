package cli

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/faustbrian/go-library-tools/v2/internal/gates"
)

func TestExecuteSelectedMutationRejectsUnavailablePackage(t *testing.T) {
	root := internalFixture(t)
	var stdout, stderr bytes.Buffer
	factory := func(string, io.Writer, io.Writer) (gates.Executor, func() error, error) {
		return coverageExecutor{}, func() error { return nil }, nil
	}
	code := execute([]string{"mutation", "--module", ".", "--package", "missing"}, root, &stdout, &stderr, factory)
	if code != 1 || !strings.Contains(stderr.String(), "does not enable mutation") {
		t.Fatalf("selected mutation = %d, stderr %q", code, stderr.String())
	}
}

func selectedMutationCLIFixture(t *testing.T, production bool) string {
	t.Helper()
	root := internalFixture(t)
	manifest, err := os.ReadFile(filepath.Join(root, "modules.json"))
	if err != nil {
		t.Fatal(err)
	}
	manifest = bytes.Replace(manifest, []byte(`"gates":{}`), []byte(`"gates":{"mutation":true}`), 1)
	packageRecord := `{"module_directory":".","directory":".","name":"example","import_path":"example","kind":"library","production":true,"coverage_required":true}`
	if !production {
		packageRecord = strings.Replace(packageRecord, `"production":true`, `"production":false`, 1)
	}
	manifest = bytes.Replace(manifest, []byte(`"packages":[]`), []byte(`"packages":[`+packageRecord+`]`), 1)
	if err := os.WriteFile(filepath.Join(root, "modules.json"), manifest, 0o600); err != nil {
		t.Fatal(err)
	}
	packages := `{"schema_version":1,"repository":"example","packages":[` + packageRecord + `]}`
	if err := os.WriteFile(filepath.Join(root, "packages.json"), []byte(packages), 0o600); err != nil {
		t.Fatal(err)
	}
	return root
}

func TestExecuteSelectedMutationRoutesOneScopeAndCleansFailedSetup(t *testing.T) {
	root := selectedMutationCLIFixture(t, true)
	created, cleaned := 0, 0
	factory := func(string, io.Writer, io.Writer) (gates.Executor, func() error, error) {
		created++
		return cliWorkspaceExecutor{directory: t.TempDir()}, func() error { cleaned++; return nil }, nil
	}
	var stdout, stderr bytes.Buffer
	code := execute([]string{"mutation", "--package", ".", "--module", "."}, root, &stdout, &stderr, factory)
	if code != 1 || created != 1 || cleaned != 1 || !strings.Contains(stderr.String(), "read zero-mutant inventory") || !strings.Contains(stdout.String(), "package . only") {
		t.Fatalf("selected setup = %d, sessions %d/%d, stdout %q, stderr %q", code, created, cleaned, stdout.String(), stderr.String())
	}
}

func TestExecuteSelectedMutationRejectsNonProductionCanonicalPackage(t *testing.T) {
	root := selectedMutationCLIFixture(t, false)
	var stdout, stderr bytes.Buffer
	factory := func(string, io.Writer, io.Writer) (gates.Executor, func() error, error) {
		return cliWorkspaceExecutor{directory: t.TempDir()}, func() error { return nil }, nil
	}
	code := execute([]string{"mutation", "--module", ".", "--package", "."}, root, &stdout, &stderr, factory)
	if code != 1 || !strings.Contains(stderr.String(), "not mutation-enabled") || stdout.Len() != 0 {
		t.Fatalf("non-production admission = %d, stdout %q, stderr %q", code, stdout.String(), stderr.String())
	}
}

func TestExecuteSelectedMutationMalformedArgumentsDoNotCreateExecutor(t *testing.T) {
	root := internalFixture(t)
	for _, args := range [][]string{
		{"mutation", "--package", "."},
		{"mutation", "--module", ".", "--package"},
		{"mutation", "--module", ".", "--package", ""},
		{"mutation", "--module", ".", "--package", ".", "--package", "other"},
		{"mutation", "--module", ".", "--module", ".", "--package", "."},
		{"mutation", "--module", ".", "--package", ".", "--all"},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			factory := func(string, io.Writer, io.Writer) (gates.Executor, func() error, error) {
				t.Fatal("malformed selection created an executor")
				return nil, nil, nil
			}
			if code := execute(args, root, &stdout, &stderr, factory); code != 2 || !strings.Contains(stderr.String(), "usage: golib mutation") {
				t.Fatalf("selected mutation = %d, stderr %q", code, stderr.String())
			}
		})
	}
}
