package mutation

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestSharedCoverageCalibratesWithNativeTestCPU(t *testing.T) {
	root := t.TempDir()
	for name, source := range map[string]string{
		"go.mod": "module example.test/cpu-calibration\n\ngo 1.27.0\n",
		"cpu_test.go": `package calibration
import ("runtime"; "testing")
func TestNativeCPU(t *testing.T) {
 if got := runtime.GOMAXPROCS(0); got != 1 {
  t.Fatalf("calibration ran with %d CPUs; native mutant phases use 1", got)
 }
}
`,
	} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(source), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	workspace := t.TempDir()
	environment := map[string]string{
		"GOENV": "off", "GOTOOLCHAIN": "local", "GOMAXPROCS": "2",
		"GOCACHE": t.TempDir(), "GOMODCACHE": t.TempDir(), "GOTMPDIR": t.TempDir(),
	}
	process := func(ctx context.Context, name string, args []string, directory string, env map[string]string, _, _ io.Writer) error {
		command := exec.CommandContext(ctx, name, args...)
		command.Dir = directory
		command.Env = os.Environ()
		for key, value := range env {
			command.Env = append(command.Env, key+"="+value)
		}
		output, err := command.CombinedOutput()
		if err != nil {
			return fmt.Errorf("calibration subprocess: %w: %s", err, output)
		}
		return nil
	}
	campaign := Campaign{
		Root: root, Workspace: workspace, Environment: environment,
		Policy: CampaignPolicy{ModuleDirectory: "."}, Process: process,
	}
	state := campaignState{toolBuilt: true}
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Minute)
	defer cancel()
	if err := campaign.prepareExecution(ctx, &state); err != nil {
		t.Fatal(err)
	}
	if profile, err := os.ReadFile(state.coverageProfile); err != nil || len(profile) == 0 {
		t.Fatalf("calibration did not produce coverage: %v", err)
	}
	if budget, err := time.ParseDuration(state.coverageElapsed); err != nil || budget < minimumMutationPhaseTimeout {
		t.Fatalf("calibration did not retain a bounded phase budget: %q, %v", state.coverageElapsed, err)
	}
}
