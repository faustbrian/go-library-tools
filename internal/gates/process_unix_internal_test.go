//go:build darwin || linux

package gates

import (
	"os/exec"
	"testing"
)

func TestScannerPreparationOwnsOnlyOriginalProcessGroup(t *testing.T) {
	process := exec.CommandContext(t.Context(), "inert-command-that-is-never-started")
	if err := configureProcessGroup(process); err != nil {
		t.Fatal(err)
	}
	if process.SysProcAttr == nil || !process.SysProcAttr.Setpgid || process.SysProcAttr.Pgid != 0 {
		t.Fatal("scanner does not create its original process group")
	}
	if process.Process != nil {
		t.Fatal("configuration started a process")
	}
}
