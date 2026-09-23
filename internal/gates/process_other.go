//go:build !darwin && !linux

package gates

import (
	"errors"
	"os/exec"
)

func configureProcessGroup(*exec.Cmd) error {
	return errors.New("bounded process-tree termination is unsupported on this platform")
}

func terminateProcessTree(process *exec.Cmd) {
	if process.Process != nil {
		_ = process.Process.Kill()
	}
}
