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
	for _, event := range []string{"pull_request", "push"} {
		t.Run(event, func(t *testing.T) {
			task := t.TempDir()
			log := filepath.Join(task, "commands")
			writeExecutable(t, filepath.Join(task, "golib"), `#!/usr/bin/env bash
set -euo pipefail
if test "$1" = --help; then
  printf '%s\n' 'golib check [--local] [--all|--module <directory>]' \
    'golib check --local --module <directory> [--module <directory> ...]' \
    'golib check --local --base-sha <commit> [--all|--module <directory> ...]'
else
  printf '%s\n' "$*" >>"$CHECK_LOG"
fi
`)
			base := strings.Repeat("a", 40)
			command := exec.CommandContext(t.Context(), "bash", "-c", step.Run)
			command.Env = append(os.Environ(), "PATH="+task+string(os.PathListSeparator)+os.Getenv("PATH"), "EVENT_NAME="+event, "BASE_SHA="+base, "MODULE_MATRIX=[{\"directory\":\".\"}]", "CHECK_LOG="+log)
			if output, err := command.CombinedOutput(); err != nil {
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
