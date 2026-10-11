package repository_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestReusableChangedLocalFuzzUsesPRAndPushBase(t *testing.T) {
	step := reusableStep(t, "quality", "", "Run repository module contracts")
	for _, scenario := range []struct {
		event          string
		changedSecrets bool
	}{
		{"pull_request", true}, {"push", true},
		{"pull_request", false}, {"push", false},
	} {
		t.Run(scenario.event+map[bool]string{true: " capable", false: " missing secrets"}[scenario.changedSecrets], func(t *testing.T) {
			task := t.TempDir()
			log := filepath.Join(task, "commands")
			writeExecutable(t, filepath.Join(task, "golib"), `#!/usr/bin/env bash
set -euo pipefail
if test "$1" = --help; then
  printf '%s\n' 'golib check [--local] [--all|--module <directory>]' \
    'golib check --local --module <directory> [--module <directory> ...]' \
    'golib check --local --base-sha <commit> [--all|--module <directory> ...]'
  if test "$CHANGED_SECRETS" = true; then
    printf '%s\n' 'golib secrets check [--base-sha <commit>]'
  fi
else
  printf '%s\n' "$*" >>"$CHECK_LOG"
fi
`)
			base := strings.Repeat("a", 40)
			command := exec.CommandContext(t.Context(), "bash", "-c", step.Run)
			capable := "false"
			if scenario.changedSecrets {
				capable = "true"
			}
			command.Env = append(os.Environ(), "PATH="+task+string(os.PathListSeparator)+os.Getenv("PATH"), "EVENT_NAME="+scenario.event, "BASE_SHA="+base, "MODULE_MATRIX=[{\"directory\":\".\"}]", "CHECK_LOG="+log, "CHANGED_SECRETS="+capable)
			output, err := command.CombinedOutput()
			if !scenario.changedSecrets {
				if err == nil {
					t.Fatal("fuzz-capable tooling without changed secrets was accepted")
				}
				if _, err := os.Stat(log); !os.IsNotExist(err) {
					t.Fatal("incapable tooling dispatched runtime checks")
				}
				return
			}
			if err != nil {
				t.Fatalf("quality script failed: %v: %s", err, output)
			}
			data, err := os.ReadFile(log)
			if err != nil {
				t.Fatal(err)
			}
			if string(data) != "check --local --base-sha "+base+" --module .\n" {
				t.Fatalf("changed quality invocation = %q", data)
			}
		})
	}
}
