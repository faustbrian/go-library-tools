package repository_test

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestReusableCodeQLBuildsAffectedPRModulesAndPreservesFullMainCoverage(t *testing.T) {
	// Execute the actual workflow body: compilation scope, canonical tags,
	// failure propagation and private-cache cleanup are this boundary's contract.
	step := reusableStep(t, "codeql", "", "Build selected modules")
	const selected = `[{"directory":"adapter"},{"directory":"consumer"}]`
	const affected = "adapter|build -tags=adapter,safe ./...|off\nconsumer|build ./...|off\n"
	const complete = ".|build ./...|off\nadapter|build -tags=adapter,safe ./...|off\nconsumer|build ./...|off\nother|build ./...|off\n"
	for _, scenario := range []struct {
		name, event, matrix, failedModule, inventoryStatus, commands string
		code                                                         int
	}{
		{"PR owners and reverse consumer", "pull_request", selected, "", "0", affected, 0},
		{"matrix cannot replace canonical tags", "pull_request", `[{"directory":"adapter","build_tags":["wrong"]},{"directory":"consumer"}]`, "", "0", affected, 0},
		{"main retains complete analysis", "push", selected, "", "0", complete, 0},
		{"schedule retains complete analysis", "schedule", selected, "", "0", complete, 0},
		{"manual assurance retains complete analysis", "workflow_dispatch", selected, "", "0", complete, 0},
		{"main does not consume partial matrix", "push", `{`, "", "0", complete, 0},
		{"empty selection", "pull_request", `[]`, "", "0", "", 1},
		{"invalid JSON", "pull_request", `{`, "", "0", "", 1},
		{"duplicate owner", "pull_request", `[{"directory":"adapter"},{"directory":"adapter"}]`, "", "0", "", 1},
		{"unknown owner", "pull_request", `[{"directory":"missing"}]`, "", "0", "", 1},
		{"partial lookup fails before any build", "pull_request", `[{"directory":"adapter"},{"directory":"missing"}]`, "", "0", "", 1},
		{"escaped owner", "pull_request", `[{"directory":"../adapter"}]`, "", "0", "", 1},
		{"newline owner", "pull_request", `[{"directory":"adapter\nother"}]`, "", "0", "", 1},
		{"non-string owner", "pull_request", `[{"directory":null}]`, "", "0", "", 1},
		{"missing owner", "pull_request", `[{"artifact":"adapter"}]`, "", "0", "", 1},
		{"selection cardinality", "pull_request", "[" + strings.Repeat(`{"directory":"adapter"},`, 64) + `{"directory":"consumer"}]`, "", "0", "", 1},
		{"selection bytes", "pull_request", strings.Repeat(" ", 65537) + selected, "", "0", "", 1},
		{"inventory failure", "pull_request", selected, "", "29", "", 29},
		{"consumer build failure", "pull_request", selected, "consumer", "0", affected, 19},
		{"failed build stops later owners", "pull_request", `[{"directory":"adapter"},{"directory":"consumer"},{"directory":"other"}]`, "consumer", "0", affected, 19},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			task := t.TempDir()
			root, bin, runner := filepath.Join(task, "source"), filepath.Join(task, "bin"), filepath.Join(task, "runner")
			for _, directory := range []string{root, bin, runner, filepath.Join(root, "adapter"), filepath.Join(root, "consumer"), filepath.Join(root, "other")} {
				if err := os.MkdirAll(directory, 0o700); err != nil {
					t.Fatal(err)
				}
			}
			inventory := `{"modules":[{"directory":"."},{"directory":"adapter","build_tags":["adapter","safe"]},{"directory":"consumer"},{"directory":"other"}]}`
			if err := os.WriteFile(filepath.Join(root, "modules.json"), []byte(inventory), 0o600); err != nil {
				t.Fatal(err)
			}
			log := filepath.Join(task, "builds")
			writeExecutable(t, filepath.Join(bin, "golib"), `#!/usr/bin/env bash
set -euo pipefail
test "$*" = 'inventory --json'
if test "$INVENTORY_STATUS" != 0; then exit "$INVENTORY_STATUS"; fi
printf '%s\n' "$INVENTORY_JSON"
`)
			writeExecutable(t, filepath.Join(bin, "go"), `#!/usr/bin/env bash
set -euo pipefail
for cache in "$GOCACHE" "$GOMODCACHE" "$GOTMPDIR"; do
  case "$cache" in "$RUNNER_TEMP"/golib-codeql.*/*) ;; *) exit 73 ;; esac
  test -d "$cache"
done
relative="${PWD#"$SOURCE_ROOT"}"
relative="${relative#/}"
relative="${relative:-.}"
printf '%s|%s|%s\n' "$relative" "$*" "${GOWORK:-unset}" >>"$BUILD_LOG"
if test "$relative" = "$FAILED_MODULE"; then exit 19; fi
`)
			command := exec.CommandContext(t.Context(), "bash", "-c", step.Run)
			command.Dir = root
			command.Env = append(os.Environ(), "PATH="+bin+string(os.PathListSeparator)+os.Getenv("PATH"),
				"RUNNER_TEMP="+runner, "MODULE_MATRIX="+scenario.matrix,
				"EVENT_NAME="+scenario.event, "RELEASE_DRY_RUN=false", "SOURCE_ROOT="+root,
				"BUILD_LOG="+log, "FAILED_MODULE="+scenario.failedModule,
				"INVENTORY_STATUS="+scenario.inventoryStatus, "INVENTORY_JSON="+inventory,
				"GOWORK="+filepath.Join(task, "unrelated.go.work"))
			output, err := command.CombinedOutput()
			code := 0
			if err != nil {
				var failure *exec.ExitError
				if !errors.As(err, &failure) {
					t.Fatal(err)
				}
				code = failure.ExitCode()
			}
			if code != scenario.code {
				t.Fatalf("CodeQL build code = %d, want %d: %s", code, scenario.code, output)
			}
			data, readErr := os.ReadFile(log)
			if readErr != nil && !os.IsNotExist(readErr) {
				t.Fatal(readErr)
			}
			if string(data) != scenario.commands {
				t.Fatalf("CodeQL compilation = %q, want %q", data, scenario.commands)
			}
			remaining, readErr := os.ReadDir(runner)
			if readErr != nil || len(remaining) != 0 {
				t.Fatalf("CodeQL cache cleanup: entries=%v error=%v", remaining, readErr)
			}
		})
	}
}
