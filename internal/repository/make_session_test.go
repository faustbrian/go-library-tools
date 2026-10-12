package repository_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestMakeReusesDisposableSessionAndCleansFailures exercises shared cache
// identity, fail-closed documentation admission before runtime checks, and
// cleanup on both downstream and early verifier failures.
func TestMakeReusesDisposableSessionAndCleansFailures(t *testing.T) {
	for _, test := range []struct {
		name, target, failure   string
		want                    []string
		changeEvent, changeBase string
	}{
		{"package tests and vet", "check-packages", "", []string{"go test ./internal/example", "go vet ./internal/example"}, "", ""},
		{"package test failure", "check-packages", "test ./internal/example", []string{"go test ./internal/example"}, "", ""},
		{"package vet failure", "check-packages", "vet ./internal/example", []string{"go test ./internal/example", "go vet ./internal/example"}, "", ""},
		{"local checks", "local-check", "", []string{"go build", "golib source-docs check --json", "golib config validate", "golib repository check", "golib cohesion check", "golib workflows check"}, "", ""},
		{"local documentation failure", "local-check", "source-docs check --json", []string{"go build", "golib source-docs check --json"}, "", ""},
		{"local CI documentation failure", "local-ci", "source-docs check --json", []string{"go build", "golib source-docs check --json"}, "", ""},
		{"metadata documentation failure", "metadata-ci", "source-docs check --json", []string{"go build", "golib source-docs check --json"}, "pull_request", strings.Repeat("a", 40)},
		{"CLI build failure", "local-ci", "build", []string{"go build"}, "", ""},
		{"local CI", "local-ci", "", []string{"go build", "golib source-docs check --json", "golib config validate", "golib repository check", "golib cohesion check", "golib workflows check", "golib security validate --directory docs/ecosystem/security", "golib check --local"}, "", ""},
		{"local CI failure", "local-ci", "cohesion check", []string{"go build", "golib source-docs check --json", "golib config validate", "golib repository check", "golib cohesion check"}, "", ""},
		{"PR fuzz base", "local-ci", "", []string{"go build", "golib source-docs check --json", "golib config validate", "golib repository check", "golib cohesion check", "golib workflows check", "golib security validate --directory docs/ecosystem/security", "golib check --local --base-sha " + strings.Repeat("a", 40)}, "pull_request", strings.Repeat("a", 40)},
		{"push fuzz base", "local-ci", "", []string{"go build", "golib source-docs check --json", "golib config validate", "golib repository check", "golib cohesion check", "golib workflows check", "golib security validate --directory docs/ecosystem/security", "golib check --local --base-sha " + strings.Repeat("b", 40)}, "push", strings.Repeat("b", 40)},
		{"metadata CI", "metadata-ci", "", []string{"go build", "golib source-docs check --json", "golib config validate", "golib repository check", "golib cohesion check", "golib workflows check", "golib security validate --directory docs/ecosystem/security", "golib docs check --local", "golib secrets check --base-sha " + strings.Repeat("a", 40)}, "pull_request", strings.Repeat("a", 40)},
		{"metadata secret failure", "metadata-ci", "secrets check --base-sha " + strings.Repeat("a", 40), []string{"go build", "golib source-docs check --json", "golib config validate", "golib repository check", "golib cohesion check", "golib workflows check", "golib security validate --directory docs/ecosystem/security", "golib docs check --local", "golib secrets check --base-sha " + strings.Repeat("a", 40)}, "pull_request", strings.Repeat("a", 40)},
		{"metadata doc failure", "metadata-ci", "docs check --local", []string{"go build", "golib source-docs check --json", "golib config validate", "golib repository check", "golib cohesion check", "golib workflows check", "golib security validate --directory docs/ecosystem/security", "golib docs check --local"}, "push", strings.Repeat("b", 40)},
		{"metadata missing base", "metadata-ci", "missing base", []string{"go build"}, "pull_request", ""},
	} {
		t.Run(test.name, func(t *testing.T) {
			repository := t.TempDir()
			writeRehearsalFile(t, filepath.Join(repository, "Makefile"), readProjectFile(t, "Makefile"))
			bin := t.TempDir()
			log := filepath.Join(t.TempDir(), "commands")
			cli := filepath.Join(bin, "golib-fixture")
			writeExecutable(t, cli, `#!/usr/bin/env bash
set -euo pipefail
printf 'golib %s\t%s\t%s\t%s\n' "$*" "$GOCACHE" "$GOMODCACHE" "$GOTMPDIR" >>"$GOLIB_MAKE_FIXTURE_LOG"
test "$GOLIB_TASK_CACHE_ROOT" = "$(dirname "$GOCACHE")"
if test "$*" = "$GOLIB_MAKE_FIXTURE_FAILURE"; then exit 23; fi
# Application tests can modify writable session outputs. No verifier may run
# afterward; replace this fixture executable to expose any such invocation.
if test "$1" = check; then printf '#!/bin/sh\nexit 99\n' > "$0"; fi
`)
			writeExecutable(t, filepath.Join(bin, "go"), `#!/usr/bin/env bash
set -euo pipefail
if test "$1" = build; then
  printf 'go build\t%s\t%s\t%s\n' "$GOCACHE" "$GOMODCACHE" "$GOTMPDIR" >>"$GOLIB_MAKE_FIXTURE_LOG"
  if test "$GOLIB_MAKE_FIXTURE_FAILURE" = build; then exit 23; fi
  test "$2" = -o
  cp "$GOLIB_MAKE_FIXTURE_BIN" "$3"
else
  printf 'go %s\t%s\t%s\t%s\n' "$*" "$GOCACHE" "$GOMODCACHE" "$GOTMPDIR" >>"$GOLIB_MAKE_FIXTURE_LOG"
  if test "$*" = "$GOLIB_MAKE_FIXTURE_FAILURE"; then exit 23; fi
fi
`)
			command := exec.CommandContext(t.Context(), "make", test.target, "PACKAGES=./internal/example")
			command.Dir = repository
			command.Env = append(os.Environ(), "PATH="+bin+string(os.PathListSeparator)+os.Getenv("PATH"),
				"GOLIB_MAKE_FIXTURE_LOG="+log, "GOLIB_MAKE_FIXTURE_BIN="+cli,
				"GOLIB_CHANGE_EVENT="+test.changeEvent, "GOLIB_BASE_SHA="+test.changeBase,
				"GOLIB_MAKE_FIXTURE_FAILURE="+test.failure)
			output, err := command.CombinedOutput()
			if (err != nil) != (test.failure != "") {
				t.Fatalf("make error = %v, output = %s", err, output)
			}
			contents, err := os.ReadFile(log)
			if err != nil {
				t.Fatalf("command log: %v; make output = %s", err, output)
			}
			lines := strings.Split(strings.TrimSpace(string(contents)), "\n")
			if len(lines) != len(test.want) {
				t.Fatalf("commands = %q, want %q", lines, test.want)
			}
			var session []string
			for index, line := range lines {
				fields := strings.Split(line, "\t")
				if len(fields) != 4 || fields[0] != test.want[index] {
					t.Fatalf("command = %q, want %q", line, test.want[index])
				}
				if session == nil {
					session = fields[1:]
				}
				for cacheIndex, directory := range fields[1:] {
					if directory != session[cacheIndex] {
						t.Fatalf("adjacent checks did not reuse their session: %q != %q", directory, session[cacheIndex])
					}
					if _, err := os.Stat(directory); !os.IsNotExist(err) {
						t.Fatalf("disposable cache remains: %q, error = %v", directory, err)
					}
				}
			}
		})
	}
}
