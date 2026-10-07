package mutation

import (
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestHostedVerifierUnitFirstExecution(t *testing.T) {
	if os.Getenv("CI") != "true" {
		t.Skip("actual verifier construction and execution are hosted-only")
	}
	workspace := t.TempDir()
	t.Cleanup(func() {
		if err := filepath.WalkDir(workspace, func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				return os.Chmod(path, 0o700)
			}
			return nil
		}); err != nil {
			t.Error("verifier workspace cleanup permission repair failed")
		}
	})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	run := func(ctx context.Context, name string, arguments []string, directory string, environment map[string]string, stdout, stderr io.Writer) error {
		command := exec.CommandContext(ctx, name, arguments...)
		command.Dir = directory
		command.Env = os.Environ()
		for key, value := range environment {
			command.Env = append(command.Env, key+"="+value)
		}
		command.Stdout, command.Stderr = stdout, stderr
		return command.Run()
	}
	verified := false
	process := func(ctx context.Context, name string, arguments []string, directory string, environment map[string]string, stdout, stderr io.Writer) error {
		if name == "go" && len(arguments) > 0 && arguments[0] == "build" {
			if err := run(ctx, "go", []string{"test", "-p=1", "./internal/engine", "-run", "^TestHostedUnitFirstFailureAndExampleFollowup$", "-count=1", "-timeout=2m"}, directory, environment, io.Discard, io.Discard); err != nil {
				return err
			}
			verified = true
		}
		return run(ctx, name, arguments, directory, environment, stdout, stderr)
	}
	if _, err := BuildVerifier(ctx, workspace, process); err != nil {
		t.Fatal("pinned verifier construction or finite executor contract failed")
	}
	if !verified {
		t.Fatal("pinned verifier build omitted the finite executor contract")
	}
}
