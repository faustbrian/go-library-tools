package mutation

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"io"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"
)

//go:embed testdata/native_cli_diagnostic.go.txt
var nativeCLIDiagnostic []byte

func TestNativeCLIDiagnosticInjectionRequiresUniqueRunBoundary(t *testing.T) {
	for _, source := range []string{"", "rel, err := run(cmd)\nrel, err := run(cmd)"} {
		if _, err := injectNativeCLIProbe([]byte(source)); err == nil {
			t.Fatal("ambiguous native execution boundary accepted")
		}
	}
	got, err := injectNativeCLIProbe([]byte("before\nrel, err := run(cmd)\nafter"))
	if err != nil || string(got) != "before\nrel, err := diagnosticObservedRun(cmd, ctx)\nafter" {
		t.Fatal("diagnostic changed more than the selected run boundary")
	}
}

func injectNativeCLIProbe(source []byte) ([]byte, error) {
	const anchor = "rel, err := run(cmd)"
	if bytes.Count(source, []byte(anchor)) != 1 {
		return nil, ErrInvalid
	}
	return bytes.Replace(source, []byte(anchor), []byte("rel, err := diagnosticObservedRun(cmd, ctx)"), 1), nil
}

func TestHostedNativeCLISingleMutationDiagnostic(t *testing.T) {
	if os.Getenv("CI") != "true" || os.Getenv("GOLIB_NATIVE_CLI_DIAGNOSTIC") != "1" {
		t.Skip("native CLI subprocess diagnostic is explicitly hosted-only")
	}
	// Neither the shared baseline nor the mutant full-suite followup may invoke
	// this opt-in diagnostic recursively. Ordinary test selection is unchanged.
	t.Setenv("GOLIB_NATIVE_CLI_DIAGNOSTIC", "")
	ctx, cancel := context.WithTimeout(t.Context(), 25*time.Minute)
	defer cancel()
	defer func() { t.Logf("native-cli outer_deadline_exceeded=%t", ctx.Err() == context.DeadlineExceeded) }()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal("resolve diagnostic source root")
	}
	identity := exec.CommandContext(ctx, "git", "rev-parse", "HEAD")
	identity.Dir = root
	commit, err := identity.Output()
	if err != nil {
		t.Fatal("resolve immutable diagnostic source identity")
	}
	t.Logf("native-cli source=%s verifier=%s historical_cache_equivalence=false", strings.TrimSpace(string(commit)), LegacyVerifierDigest())
	clean := exec.CommandContext(ctx, "git", "diff", "--quiet", "HEAD", "--")
	clean.Dir = root
	if err := clean.Run(); err != nil {
		t.Fatal("diagnostic requires unchanged committed source inputs")
	}
	manifest, err := os.ReadFile(filepath.Join(root, "modules.json"))
	var modules struct {
		Modules []struct {
			Directory string   `json:"directory"`
			TestTags  []string `json:"test_tags"`
		} `json:"modules"`
	}
	if err != nil || json.Unmarshal(manifest, &modules) != nil || len(modules.Modules) != 1 ||
		modules.Modules[0].Directory != "." || len(modules.Modules[0].TestTags) != 0 {
		t.Fatal("diagnostic root-module/tag policy changed")
	}
	workspace := t.TempDir()
	t.Cleanup(func() {
		if err := filepath.WalkDir(workspace, func(path string, entry os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() {
				return os.Chmod(path, 0o700)
			}
			return nil
		}); err != nil {
			t.Error("diagnostic workspace cleanup permission repair failed")
		}
	})
	execution := map[string]string{"GOWORK": "off", "GOFLAGS": os.Getenv("GOFLAGS")}
	for _, key := range []string{"GOCACHE", "GOMODCACHE", "GOTMPDIR"} {
		if value := os.Getenv(key); value != "" && filepath.IsAbs(value) {
			execution[key] = value
		} else {
			t.Fatal("diagnostic requires explicit task-owned Go directories")
		}
	}
	campaign := Campaign{Root: root, Workspace: workspace, Environment: execution,
		Policy: CampaignPolicy{ModuleDirectory: "."}, Process: nativeDiagnosticProcess}
	execution = campaign.commandEnvironment()
	state := campaignState{toolBuilt: true}
	if err := campaign.prepareExecution(ctx, &state); err != nil {
		t.Fatal("native diagnostic shared coverage baseline failed")
	}
	t.Logf("native-cli shared_phase_budget=%s", state.coverageElapsed)
	// Match runPackage's cache separation: baseline compilation must not warm
	// the mutant subprocess cache. This does not recreate campaign history.
	execution = maps.Clone(execution)
	execution["GOCACHE"] = filepath.Join(workspace, "mutation-cache", packageSlug("internal/cli"))
	if err := os.MkdirAll(execution["GOCACHE"], 0o700); err != nil {
		t.Fatal("create native diagnostic package mutation cache")
	}
	t.Log("native-cli cache_mode=separate-package-cache campaign_history_equivalence=false")
	verified := false
	process := func(ctx context.Context, name string, args []string, directory string, environment map[string]string, stdout, stderr io.Writer) error {
		if name == "go" && len(args) > 0 && args[0] == "build" {
			path := filepath.Join(directory, "internal", "engine", "executor.go")
			source, readErr := os.ReadFile(path)
			if readErr != nil {
				return ErrInvalid
			}
			instrumented, injectErr := injectNativeCLIProbe(source)
			if injectErr != nil {
				return injectErr
			}
			if err := os.WriteFile(path, instrumented, 0o600); err != nil {
				return err
			}
			if err := os.WriteFile(filepath.Join(directory, "internal", "engine", "native_cli_diagnostic_test.go"), nativeCLIDiagnostic, 0o600); err != nil {
				return err
			}
			for key, value := range execution {
				environment["GOLIB_NATIVE_EXEC_"+key] = value
			}
			environment["GOLIB_NATIVE_SOURCE_ROOT"] = root
			environment["GOLIB_NATIVE_PHASE_BUDGET"] = state.coverageElapsed
			var output boundedMutationOutput
			err := nativeDiagnosticProcess(ctx, "go", []string{"test", "-p=1", "./internal/engine", "-run", "^TestNativeCLISingleMutation$", "-count=1", "-timeout=24m", "-v"}, directory, environment, &output, &output)
			t.Log(output.buffer.String())
			if err != nil || output.overflow {
				return ErrInvalid
			}
			verified = true
			// Restore the copied verifier before its normal pinned binary build.
			if err := os.WriteFile(path, source, 0o600); err != nil {
				return err
			}
		}
		return nativeDiagnosticProcess(ctx, name, args, directory, environment, stdout, stderr)
	}
	if _, err := BuildVerifier(ctx, filepath.Join(workspace, "verifier"), process); err != nil || !verified {
		t.Fatal("native CLI diagnostic did not complete; no qualification inferred")
	}
}

func nativeDiagnosticProcess(ctx context.Context, name string, args []string, directory string, environment map[string]string, stdout, stderr io.Writer) error {
	command := exec.CommandContext(ctx, name, args...)
	command.Dir = directory
	values := make(map[string]string)
	for _, entry := range os.Environ() {
		key, value, found := strings.Cut(entry, "=")
		if found {
			values[key] = value
		}
	}
	maps.Copy(values, environment)
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		command.Env = append(command.Env, key+"="+values[key])
	}
	command.Stdout, command.Stderr = stdout, stderr
	return command.Run()
}
