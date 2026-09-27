//go:build !darwin && !linux

package gates

import (
	"errors"
	"os/exec"
)

func configureProcessGroup(*exec.Cmd) error {
	return errors.New("bounded scanner process-group termination is unsupported on this platform")
}

func terminateProcessGroup(process *exec.Cmd) {
	if process.Process != nil {
		_ = process.Process.Kill()
	}
}
