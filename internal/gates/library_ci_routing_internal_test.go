package gates

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"go.yaml.in/yaml/v3"
)

func TestLibraryCIRoutesOrdinaryChecksByCLIContract(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("..", "..", ".github", "workflows", "library-ci.yml"))
	if err != nil {
		t.Fatal(err)
	}
	var workflow struct {
		Jobs map[string]struct {
			Steps []struct{ Name, Run string }
		}
	}
	if err := yaml.Unmarshal(content, &workflow); err != nil {
		t.Fatal(err)
	}
	var script string
	for _, step := range workflow.Jobs["quality"].Steps {
		if step.Name == "Run repository module contracts" {
			script = step.Run
		}
	}
	if script == "" {
		t.Fatal("ordinary module contract step missing")
	}
	for _, bootstrap := range []string{"true", "false"} {
		for _, local := range []string{"true", "false"} {
			t.Run("bootstrap="+bootstrap+"/local="+local, func(t *testing.T) {
				root := t.TempDir()
				binary := filepath.Join(root, "golib")
				stub := "#!/bin/sh\nif [ \"$1\" = --help ]; then\n" +
					" if [ \"$SUPPORTS_LOCAL\" = true ]; then printf '%s\\n' 'golib check [--local]'; else printf '%s\\n' 'golib check'; fi\n" +
					"else printf '%s\\n' \"$*\" > \"$COMMAND_LOG\"; exit \"${GOLIB_ROUTING_FIXTURE_STATUS:-0}\"; fi\n"
				if err := os.WriteFile(binary, []byte(stub), 0o700); err != nil {
					t.Fatal(err)
				}
				log := filepath.Join(root, "command")
				command := exec.CommandContext(t.Context(), "bash", "-euo", "pipefail", "-c", script)
				command.Env = append(os.Environ(), "PATH="+root+":"+os.Getenv("PATH"),
					"SOURCE_BOOTSTRAP="+bootstrap, "SUPPORTS_LOCAL="+local, "COMMAND_LOG="+log)
				if err := command.Run(); err != nil {
					t.Fatal("ordinary module routing failed")
				}
				got, err := os.ReadFile(log)
				if err != nil {
					t.Fatal(err)
				}
				want := "check --all"
				if local == "true" {
					want = "check --local --all"
				}
				if strings.TrimSpace(string(got)) != want {
					t.Fatalf("ordinary routing = %q; want %q", strings.TrimSpace(string(got)), want)
				}
				failed := exec.CommandContext(t.Context(), "bash", "-euo", "pipefail", "-c", script)
				failed.Env = append(command.Env, "GOLIB_ROUTING_FIXTURE_STATUS=7")
				failure, ok := failed.Run().(*exec.ExitError)
				if !ok || failure.ExitCode() != 7 {
					t.Fatal("ordinary routing did not preserve failed gate exit status")
				}
			})
		}
	}
}
