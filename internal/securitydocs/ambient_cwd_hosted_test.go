package securitydocs_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/faustbrian/go-library-tools/v2/internal/securitydocs"
)

// This filesystem counterpart is hosted-only. t.Chdir owns restoration of the
// previous working directory before the temporary-directory cleanups run.
func TestSecurityValidationDoesNotDependOnAmbientWorkingDirectory(t *testing.T) {
	if runtime.GOOS != "linux" || os.Getenv("GITHUB_ACTIONS") != "true" {
		t.Skip("requires the hosted Linux filesystem lane")
	}

	directory := t.TempDir()
	for name, data := range map[string]string{
		"risk-register.json":   `{"schema_version":1,"risks":[]}`,
		"security-matrix.json": `{"schema_version":1,"modules":[]}`,
	} {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(data), 0o600); err != nil {
			t.Fatal("cannot prepare ordinary security documents")
		}
	}
	if !filepath.IsAbs(directory) {
		t.Fatal("ordinary document directory must be absolute")
	}
	if err := securitydocs.ValidateContext(t.Context(), directory); err != nil {
		t.Fatal("ordinary security documents failed the baseline control")
	}

	workingDirectory := t.TempDir()
	t.Chdir(workingDirectory)
	// Remove only this empty, task-owned directory, never the document root.
	if err := os.Remove(workingDirectory); err != nil {
		t.Fatal("cannot remove the empty owned working-directory fixture")
	}
	if err := securitydocs.ValidateContext(t.Context(), directory); err != nil {
		t.Fatal("absolute security document validation depends on unrelated ambient working-directory availability")
	}
}
