//go:build darwin || linux

package gates

import (
	"os/exec"
	"syscall"
)

func configureProcessGroup(process *exec.Cmd) error {
	process.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	return nil
}

// terminateProcessGroup terminates the original group. A collaborator which
// deliberately detaches is outside this mechanism's scope.
func terminateProcessGroup(process *exec.Cmd) {
	if process.Process == nil {
		return
	}
	_ = syscall.Kill(-process.Process.Pid, syscall.SIGKILL)
	_ = process.Process.Kill()
}
