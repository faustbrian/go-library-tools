package repository_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestCopiedToolingLifecycleScopesHiddenState(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	harness := filepath.Join(root, "lifecycle.sh")
	content := `#!/usr/bin/env bash
set -euo pipefail

source "$1"
root="$2"
source_root="${root}/source"
copied_tooling="${root}/copied"
mkdir -p "${source_root}/.golib"
printf 'original\n' >"${source_root}/.golib/marker"

inventory() {
    [[ ! -e "${source_root}/.golib" ]]
}
consumer_gate() {
    [[ "$(cat "${source_root}/.golib/marker")" == original ]]
    [[ ! -e "${copied_tooling}" ]]
}

hide_copied_tooling shared "${source_root}" "${copied_tooling}"
inventory
restore_copied_tooling shared "${source_root}" "${copied_tooling}"
consumer_gate

failure_source="${root}/failure-source"
failure_copy="${root}/failure-copy"
mkdir -p "${failure_source}/.golib"
printf 'failure\n' >"${failure_source}/.golib/marker"
status=0
(
    cleanup() {
        restore_copied_tooling shared "${failure_source}" "${failure_copy}"
    }
    trap cleanup EXIT
    hide_copied_tooling shared "${failure_source}" "${failure_copy}"
    [[ ! -e "${failure_source}/.golib" ]]
    exit 23
) || status=$?
[[ "${status}" -eq 23 ]]
[[ "$(cat "${failure_source}/.golib/marker")" == failure ]]
[[ ! -e "${failure_copy}" ]]

hide_copied_tooling legacy "${source_root}" "${copied_tooling}"
consumer_gate
restore_copied_tooling legacy "${source_root}" "${copied_tooling}"
consumer_gate

signal_source="${root}/signal-source"
signal_copy="${root}/signal-copy"
mkdir -p "${signal_source}/.golib"
printf 'signal\n' >"${signal_source}/.golib/marker"
status=0
(
    cleanup() {
        restore_copied_tooling shared "${signal_source}" "${signal_copy}"
    }
    trap cleanup EXIT
    trap 'cleanup; trap - EXIT TERM; exit 143' TERM
    hide_copied_tooling shared "${signal_source}" "${signal_copy}"
    kill -TERM "${BASHPID}"
) || status=$?
[[ "${status}" -ne 0 ]]
[[ "$(cat "${signal_source}/.golib/marker")" == signal ]]
[[ ! -e "${signal_copy}" ]]
`
	if err := os.WriteFile(harness, []byte(content), 0o700); err != nil {
		t.Fatal(err)
	}
	command := exec.CommandContext(
		t.Context(), "bash", harness,
		filepath.Join(projectRoot(t), "rehearsals", "copied-tooling-lifecycle.sh"), root,
	)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("copied tooling lifecycle: %v\n%s", err, output)
	}
}

func TestPerformanceComparisonAcceptsEmptyDisposableRuntime(t *testing.T) {
	t.Parallel()

	root, manifest := performanceComparisonFixture(t)
	output := filepath.Join(root, "aggregate.json")
	command := performanceComparisonCommand(t, root, output, "1", "1", "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee", "faustbrian/go-library-tools", manifest)
	if result, err := command.CombinedOutput(); err != nil {
		t.Fatalf("compare empty disposable runtime: %v\n%s", err, result)
	}
}

func TestPerformanceComparisonRejectsDetachedAttribution(t *testing.T) {
	t.Parallel()

	root, manifest := performanceComparisonFixture(t)
	badManifest := filepath.Join(root, "bad-repositories.json")
	if err := os.WriteFile(badManifest, []byte(`{"schema_version":1,"repositories":[{"name":"go-authorization","revision":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","performance_profile":"service"},{"name":"go-knapsack","revision":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","performance_profile":"core"}]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name       string
		runID      string
		attempt    string
		tooling    string
		repository string
		manifest   string
	}{
		{name: "run", runID: "2", attempt: "1", tooling: "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee", repository: "faustbrian/go-library-tools", manifest: manifest},
		{name: "attempt", runID: "1", attempt: "2", tooling: "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee", repository: "faustbrian/go-library-tools", manifest: manifest},
		{name: "tooling", runID: "1", attempt: "1", tooling: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", repository: "faustbrian/go-library-tools", manifest: manifest},
		{name: "repository", runID: "1", attempt: "1", tooling: "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee", repository: "faustbrian/other", manifest: manifest},
		{name: "manifest", runID: "1", attempt: "1", tooling: "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee", repository: "faustbrian/go-library-tools", manifest: badManifest},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			output := filepath.Join(root, "aggregate-"+test.name+".json")
			command := performanceComparisonCommand(t, root, output, test.runID, test.attempt, test.tooling, test.repository, test.manifest)
			if result, err := command.CombinedOutput(); err == nil {
				t.Fatalf("detached attribution accepted:\n%s", result)
			}
		})
	}
}

func performanceComparisonFixture(t *testing.T) (string, string) {
	t.Helper()

	root := t.TempDir()
	for _, profile := range []string{"core", "service"} {
		for _, implementation := range []string{"legacy", "shared"} {
			directory := filepath.Join(root, profile+"-"+implementation)
			if err := os.MkdirAll(directory, 0o700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(directory, "performance-services.status"), []byte("0\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			data, err := json.Marshal(performanceReportFixture(profile, implementation))
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(directory, "performance.json"), data, 0o600); err != nil {
				t.Fatal(err)
			}
		}
	}
	manifest := filepath.Join(root, "repositories.json")
	if err := os.WriteFile(manifest, []byte(`{"schema_version":1,"repositories":[{"name":"go-authorization","revision":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","performance_profile":"service"},{"name":"go-knapsack","revision":"cccccccccccccccccccccccccccccccccccccccc","performance_profile":"core"}]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	return root, manifest
}

func performanceComparisonCommand(t *testing.T, root, output, runID, attempt, tooling, repository, manifest string) *exec.Cmd {
	t.Helper()
	return exec.CommandContext(t.Context(), "bash", filepath.Join(projectRoot(t), "rehearsals", "performance-compare.sh"), root, output, runID, attempt, tooling, repository, manifest)
}

func performanceReportFixture(profile, implementation string) map[string]any {
	repository := "go-authorization"
	revision := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	content := "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	samples := performanceSamples(25, "startup-diagnostic", 2, nil)
	samples = append(samples, performanceSamples(10, "repository-inventory", 0, nil)...)
	mutationPackages := any(nil)
	if profile == "core" {
		repository = "go-knapsack"
		revision = "cccccccccccccccccccccccccccccccccccccccc"
		content = "dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd"
		mutationPackages = 8
		samples = append(samples, performanceSamples(1, "checkpoint-warmup", 0, nil)...)
		samples = append(samples, performanceSamples(3, "checkpoint-reuse", 0, 8)...)
		samples = append(samples, performanceSamples(3, "module-scaling-sequential", 0, nil)...)
		samples = append(samples, performanceSamples(3, "module-scaling-concurrent", 0, nil)...)
	} else {
		samples = append(samples, performanceSamples(1, "service-warmup", 0, nil)...)
		samples = append(samples, performanceSamples(5, "service-lifecycle", 0, nil)...)
	}
	return map[string]any{
		"schema_version": 1, "repository": repository, "revision": revision,
		"content_sha256": content, "tooling_revision": "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
		"implementation": implementation, "profile": profile,
		"workflow":            map[string]any{"repository": "faustbrian/go-library-tools", "run_id": 1, "attempt": 1},
		"golib_sha256":        "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
		"artifact_size_bytes": 1, "mutation_package_count": mutationPackages,
		"isolated_cache_residue": map[string]any{"entries": 0, "bytes": 0}, "samples": samples,
	}
}

func performanceSamples(count int, metric string, exitCode int, reuse any) []map[string]any {
	samples := make([]map[string]any, count)
	for index := range count {
		sample := map[string]any{
			"implementation": "fixture", "metric": metric, "sample": index + 1,
			"wall_ns": 1, "peak_rss_kib": 1, "exit_code": exitCode, "reuse_count": reuse,
		}
		if metric == "module-scaling-sequential" || metric == "module-scaling-concurrent" {
			sample["peak_rss_kib"] = nil
			sample["module_count"] = 3
		}
		samples[index] = sample
	}
	return samples
}
