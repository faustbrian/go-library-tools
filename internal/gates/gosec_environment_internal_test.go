package gates

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Environment isolation does not require process-group or cancellation proof.
// Keep those hosted-only tests separate while using the production merge/spawn.
type environmentProbeExecutor struct{ *processExecutor }

func (executor environmentProbeExecutor) Run(ctx context.Context, command Command) error {
	command.boundedScanner = false
	return executor.processExecutor.Run(ctx, command)
}

func TestOwnedGosecDoesNotInheritAIConfiguration(t *testing.T) {
	parent := map[string]string{
		"GOSEC_AI_PROVIDER": "openai",
		"GOSEC_AI_API_KEY":  "inert-test-value",
		"GOSEC_AI_BASE_URL": "https://example.invalid/inert",
		"GOLIB_ENV_MARKER":  "preserved",
	}
	for key, value := range parent {
		t.Setenv(key, value)
	}
	for _, name := range []string{"gosec", "golib-analysis"} {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			capture := filepath.Join(root, "environment")
			t.Setenv("GOLIB_ENV_CAPTURE", capture)
			binary := filepath.Join(root, name)
			probe := "#!/bin/sh\nprintf '%s\\n' \"${GOSEC_AI_PROVIDER:-}\" \"${GOSEC_AI_API_KEY:-}\" \"${GOSEC_AI_BASE_URL:-}\" \"${GOLIB_ENV_MARKER:-}\" > \"${GOLIB_ENV_CAPTURE}\"\n"
			if err := os.WriteFile(binary, []byte(probe), 0o700); err != nil {
				t.Fatal(err)
			}
			identity := "github.com/securego/gosec/v2/cmd/gosec@" + gosecVersion
			if name == "golib-analysis" {
				identity = "github.com/faustbrian/go-analysis/cmd/golib-analysis@" + goAnalysisVersion
			}
			tool, ok := securityCompilerTool(identity)
			if !ok {
				t.Fatal("missing owned compiler identity")
			}
			executor := environmentProbeExecutor{&processExecutor{
				task: root, stdout: io.Discard, stderr: io.Discard,
				environment: map[string]string{},
			}}
			runner := Runner{Root: root, Executor: executor, compilerToolBinaries: map[compilerTool]string{tool: binary}}
			if err := runner.securityTool(t.Context(), io.Discard, ".", name, root, identity, "./"); err != nil {
				t.Fatal(err)
			}
			body, err := os.ReadFile(capture)
			if err != nil {
				t.Fatal(err)
			}
			values := strings.Split(strings.TrimSuffix(string(body), "\n"), "\n")
			if len(values) != 4 {
				t.Fatal("incomplete child environment observation")
			}
			for index, key := range []string{"GOSEC_AI_PROVIDER", "GOSEC_AI_API_KEY", "GOSEC_AI_BASE_URL"} {
				want := parent[key]
				if name == "gosec" {
					want = ""
				}
				if values[index] != want {
					t.Errorf("child %s isolation mismatch", key)
				}
			}
			if values[3] != parent["GOLIB_ENV_MARKER"] {
				t.Error("unrelated child environment changed")
			}
			for key, value := range parent {
				if os.Getenv(key) != value {
					t.Errorf("parent %s changed", key)
				}
			}
		})
	}
}
