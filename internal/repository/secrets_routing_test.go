package repository_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestReusableChangedMetadataSecretsRoutesEventAndRejectsOldTool(t *testing.T) {
	step := reusableStep(t, "repository-contract", "", "Scan metadata-only changes for secrets")
	for _, scenario := range []struct {
		event                string
		capable, wantSuccess bool
		want                 string
	}{
		{"pull_request", true, true, "secrets check --base-sha " + strings.Repeat("a", 40)},
		{"push", true, true, "secrets check --base-sha " + strings.Repeat("a", 40)},
		{"schedule", false, true, "secrets check"},
		{"workflow_dispatch", false, true, "secrets check"},
		{"pull_request", false, false, ""},
	} {
		t.Run(scenario.event+map[bool]string{true: " capable", false: " legacy"}[scenario.capable], func(t *testing.T) {
			task := t.TempDir()
			log := filepath.Join(task, "commands")
			writeExecutable(t, filepath.Join(task, "golib"), `#!/usr/bin/env bash
set -euo pipefail
if test "$1" = --help; then
  if test "$CAPABLE" = true; then printf '%s\n' 'golib secrets check [--base-sha <commit>]'; fi
else
  printf '%s\n' "$*" >>"$CHECK_LOG"
fi
`)
			command := exec.CommandContext(t.Context(), "bash", "-c", step.Run)
			capable := "false"
			if scenario.capable {
				capable = "true"
			}
			command.Env = append(os.Environ(), "PATH="+task+string(os.PathListSeparator)+os.Getenv("PATH"), "EVENT_NAME="+scenario.event, "BASE_SHA="+strings.Repeat("a", 40), "CAPABLE="+capable, "CHECK_LOG="+log)
			output, err := command.CombinedOutput()
			if (err == nil) != scenario.wantSuccess {
				t.Fatalf("script success=%v output=%s", err == nil, output)
			}
			data, readErr := os.ReadFile(log)
			if scenario.want == "" {
				if !os.IsNotExist(readErr) {
					t.Fatal("incapable tool ran a legacy full scan")
				}
			} else if readErr != nil || strings.TrimSpace(string(data)) != scenario.want {
				t.Fatalf("scan invocation=%q error=%v", data, readErr)
			}
		})
	}
}
