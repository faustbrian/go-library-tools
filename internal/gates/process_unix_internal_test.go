//go:build darwin || linux

package gates

import (
	"os"
	"os/exec"
	"testing"
)

func TestHostedProcessTerminationWithoutStartedProcess(t *testing.T) {
	if os.Getenv("CI") != "true" {
		t.Skip("bounded process controls run only in hosted CI")
	}
	process := &exec.Cmd{}
	terminateProcessGroup(process)
	if process.Process != nil {
		t.Fatal("nil-process termination created a process")
	}
}

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
