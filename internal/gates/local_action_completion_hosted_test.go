package gates

import (
	"context"
	"os"
	"os/exec"
	"testing"
	"time"

	"go.yaml.in/yaml/v3"
)

// A traversal regression can spin without consulting the context. Keep this
// check in a disposable hosted child, with a deadline and mandatory wait.
func TestLocalActionMappingTraversalCompletesHosted(t *testing.T) {
	if os.Getenv("GITHUB_ACTIONS") != "true" {
		t.Skip("process-isolated traversal check runs only in hosted CI")
	}
	const helper = "GOLIB_LOCAL_ACTION_COMPLETION_HELPER"
	if os.Getenv(helper) == "1" {
		var document yaml.Node
		if err := yaml.Unmarshal([]byte("first: one\nsecond: two\n"), &document); err != nil {
			t.Fatal(err)
		}
		inspection := localActionInspection{ctx: t.Context()}
		if err := inspection.inspectValidated(&document, 0, false); err != nil {
			t.Fatal("finite parser-owned mapping traversal failed")
		}
		return
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, executable, "-test.run=^TestLocalActionMappingTraversalCompletesHosted$", "-test.count=1")
	command.Env = append(os.Environ(), helper+"=1")
	// Run always waits, including when CommandContext kills the owned child.
	if err := command.Run(); err != nil {
		t.Fatal("finite mapping traversal did not complete in its owned child")
	}
}
