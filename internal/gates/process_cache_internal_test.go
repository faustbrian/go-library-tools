package gates

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/faustbrian/go-library-tools/v2/internal/config"
	"github.com/faustbrian/go-library-tools/v2/internal/inventory"
	"github.com/faustbrian/go-library-tools/v2/internal/mutation"
)

type processCacheProbe struct {
	Build       string `json:"build"`
	Modules     string `json:"modules"`
	Backend     string `json:"backend"`
	Session     string `json:"session"`
	Lint        string `json:"lint"`
	LintBackend string `json:"lint_backend"`
	Output      string `json:"output"`
	Temporary   string `json:"temporary"`
	Poisoned    bool   `json:"poisoned"`
}

// The finite child probes real process environments without compiling Go,
// downloading tools, or exercising process-group termination.
func TestProcessCacheProbe(_ *testing.T) {
	if os.Getenv("GOLIB_CACHE_PROBE") != "1" {
		return
	}
	probe := processCacheProbe{
		Build: os.Getenv("GOCACHE"), Modules: os.Getenv("GOMODCACHE"),
		Backend: os.Getenv("GOCACHEPROG"), Session: os.Getenv("GOLIB_TASK_CACHE_ROOT"),
		Lint: os.Getenv("GOLANGCI_LINT_CACHE"), LintBackend: os.Getenv("GOLANGCI_LINT_CACHEPROG"),
		Output:    os.Getenv("GOBIN"),
		Temporary: os.Getenv("GOTMPDIR"),
	}
	for _, directory := range []string{probe.Build, probe.Modules} {
		marker := filepath.Join(directory, "repository-poison")
		if _, err := os.Stat(marker); err == nil {
			probe.Poisoned = true
		} else if !errors.Is(err, os.ErrNotExist) {
			os.Exit(81)
		}
		if os.Getenv("GOLIB_CACHE_PROBE_RUNTIME") == "1" {
			if err := os.WriteFile(marker, []byte("repository-controlled"), 0o600); err != nil {
				os.Exit(82)
			}
		}
	}
	data, err := json.Marshal(probe)
	if err != nil || os.WriteFile(os.Getenv("GOLIB_CACHE_PROBE_REPORT"), data, 0o600) != nil {
		os.Exit(83)
	}
	if os.Getenv("GOLIB_CACHE_PROBE_FAIL") == "1" {
		os.Exit(23)
	}
	os.Exit(0)
}

func checkedProcessCacheExecutor(t *testing.T, created Executor) *processExecutor {
	t.Helper()
	executor, ok := created.(*processExecutor)
	if !ok {
		t.Fatal("cache fixture did not create a process executor")
	}
	return executor
}

func TestTaskCacheKeepsVerifiersPrivateAfterRepositoryExecution(t *testing.T) {
	for _, borrowed := range []bool{false, true} {
		mode := "default"
		if borrowed {
			mode = "borrowed"
		}
		t.Run(mode, func(t *testing.T) {
			for _, failure := range []bool{false, true} {
				name := "success"
				if failure {
					name = "failure"
				}
				t.Run(name, func(t *testing.T) {
					repository, session := t.TempDir(), processCacheSession(t)
					bin := t.TempDir()
					if err := os.WriteFile(filepath.Join(bin, "go"), []byte("#!/bin/sh\nexec \"$GOLIB_CACHE_PROBE_BINARY\" -test.run='^TestProcessCacheProbe$'\n"), 0o700); err != nil {
						t.Fatal(err)
					}
					t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
					t.Setenv("GOLIB_CACHE_PROBE_BINARY", os.Args[0])
					t.Setenv("GOLIB_TASK_CACHE_ROOT", session)
					t.Setenv("GOCACHEPROG", "unused-external-cache-backend")
					t.Setenv("GOBIN", filepath.Join(repository, "ambient-bin"))
					var created Executor
					var cleanup func() error
					var err error
					if borrowed {
						created, cleanup, err = NewProcessExecutorWithTaskCache(repository, session, io.Discard, io.Discard)
					} else {
						created, cleanup, err = NewProcessExecutor(repository, io.Discard, io.Discard)
					}
					if err != nil {
						t.Fatal(err)
					}
					cleaned := false
					t.Cleanup(func() {
						if !cleaned {
							if err := cleanup(); err != nil {
								t.Error(err)
							}
						}
					})
					executor := checkedProcessCacheExecutor(t, created)
					binaryOutput := filepath.Join(executor.TemporaryDirectory(), "selected-bin")
					if err := os.Mkdir(binaryOutput, 0o700); err != nil {
						t.Fatal(err)
					}
					// A successful finite capture needs no process-control effects.
					executor.prepareBoundedProcess = func(*exec.Cmd) error { return nil }
					for _, bounded := range []bool{false, true} {
						runtime, err := operationCommand(repository, inventory.Module{}, config.Step{
							Type: "go-test", Packages: []string{"."}, Run: "^TestProbe$", Count: 1, Timeout: "15s",
						})
						if err != nil {
							t.Fatal(err)
						}
						runtime.boundedScanner = bounded
						runtimeProbe := runProcessCacheProbe(t, executor, runtime, true, failure)
						if runtimeProbe.Output != executor.environment["GOBIN"] {
							t.Error("unrequested binary output escaped owned scratch")
						}
						if borrowed && (runtimeProbe.Build != filepath.Join(session, "cache") || runtimeProbe.Modules != filepath.Join(session, "mod")) {
							t.Fatal("native repository test did not borrow the caller session")
						}
						verifier := processCacheVerifierCommand(t)
						verifier.boundedScanner = bounded
						// Command-level cache overrides must not cross the trust boundary.
						verifier.Env = map[string]string{"GOCACHE": runtimeProbe.Build, "GOMODCACHE": runtimeProbe.Modules, "GOCACHEPROG": "command-controlled-backend", "GOLIB_TASK_CACHE_ROOT": session, "GOLANGCI_LINT_CACHE": runtimeProbe.Lint, "GOLANGCI_LINT_CACHEPROG": "command-controlled-lint-backend"}
						verifier.Env["GOBIN"] = binaryOutput
						probe := runProcessCacheProbe(t, executor, verifier, false, failure)
						if probe.Output != binaryOutput {
							t.Error("verifier ignored its caller-owned binary output directory")
						}
						if probe.Backend != "" || probe.Session != "" || probe.LintBackend != "" {
							t.Error("verifier inherited an external cache backend or nested cache opt-in")
						}
						if probe.Poisoned || probe.Build == runtimeProbe.Build || probe.Modules == runtimeProbe.Modules || probe.Lint == runtimeProbe.Lint {
							t.Error("verifier consumed caches modified by repository execution")
						}
						for _, directory := range []string{probe.Build, probe.Modules} {
							if filepath.Dir(directory) != executor.TemporaryDirectory() {
								t.Error("verifier cache is not invocation-private")
							}
						}
					}
					if err := cleanup(); err != nil {
						t.Fatal(err)
					}
					cleaned = true
					if _, err := os.Stat(executor.TemporaryDirectory()); !errors.Is(err, os.ErrNotExist) {
						t.Fatal("verifier scratch survived cleanup")
					}
					if borrowed {
						for _, directory := range []string{"cache", "mod"} {
							if _, err := os.Stat(filepath.Join(session, directory, "repository-poison")); err != nil {
								t.Fatal("executor removed caller-owned runtime cache")
							}
						}
					}
				})
			}
		})
	}
}

// Obtain the verifier command through the actual pinned-tool producer, rather
// than assigning the desired cache classification in the test itself.
func processCacheVerifierCommand(t *testing.T) Command {
	t.Helper()
	var command Command
	runner := Runner{Executor: executorFunction(func(_ context.Context, captured Command) error { command = captured; return nil })}
	if err := runner.goTool(t.Context(), io.Discard, ".", "lint", ".", "example.invalid/pinned-verifier@v1.0.0"); err != nil {
		t.Fatal(err)
	}
	return command
}

func TestRepositoryScriptsCannotUseVerifierOrBorrowedCaches(t *testing.T) {
	for _, borrowed := range []bool{false, true} {
		mode := "default"
		if borrowed {
			mode = "borrowed"
		}
		t.Run(mode, func(t *testing.T) {
			root, session := t.TempDir(), processCacheSession(t)
			if err := os.WriteFile(filepath.Join(root, "Makefile"), []byte("test:\n\tgo test .\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			bin := t.TempDir()
			for _, name := range []string{"go", "make", "repository-script"} {
				if err := os.WriteFile(filepath.Join(bin, name), []byte("#!/bin/sh\nexec \"$GOLIB_CACHE_PROBE_BINARY\" -test.run='^TestProcessCacheProbe$'\n"), 0o700); err != nil {
					t.Fatal(err)
				}
			}
			t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
			t.Setenv("GOLIB_CACHE_PROBE_BINARY", os.Args[0])
			var created Executor
			var cleanup func() error
			var err error
			if borrowed {
				created, cleanup, err = NewProcessExecutorWithTaskCache(root, session, io.Discard, io.Discard)
			} else {
				created, cleanup, err = NewProcessExecutor(root, io.Discard, io.Discard)
			}
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := cleanup(); err != nil {
					t.Error(err)
				}
			})
			executor := checkedProcessCacheExecutor(t, created)
			executor.prepareBoundedProcess = func(*exec.Cmd) error { return nil }
			verifier := runProcessCacheProbe(t, executor, processCacheVerifierCommand(t), true, false)
			for _, bounded := range []bool{false, true} {
				makeCommand, err := operationCommand(root, inventory.Module{}, config.Step{Type: "make", Makefile: "Makefile", Target: "test"})
				if err != nil {
					t.Fatal(err)
				}
				for _, command := range []Command{makeCommand, {Name: "repository-script"}} {
					command.boundedScanner = bounded
					command.Env = map[string]string{"GOCACHE": verifier.Build, "GOMODCACHE": verifier.Modules, "GOCACHEPROG": "repository-backend", "GOLIB_TASK_CACHE_ROOT": session, "GOLANGCI_LINT_CACHE": executor.environment["GOLANGCI_LINT_CACHE"], "GOLANGCI_LINT_CACHEPROG": "repository-lint-backend"}
					probe := runProcessCacheProbe(t, executor, command, false, true)
					if probe.Poisoned || probe.Build == verifier.Build || probe.Modules == verifier.Modules || probe.Lint == verifier.Lint || probe.LintBackend != "" || probe.Build == filepath.Join(session, "cache") || probe.Modules == filepath.Join(session, "mod") || probe.Backend != "" || probe.Session != "" {
						t.Error("repository script received verifier identity, borrowed caches, or an external backend")
					}
				}
				native, err := operationCommand(root, inventory.Module{}, config.Step{Type: "go-test", Packages: []string{"."}, Count: 1, Timeout: "15s"})
				if err != nil {
					t.Fatal(err)
				}
				native.boundedScanner = bounded
				native.Env = map[string]string{"GOCACHE": verifier.Build, "GOMODCACHE": verifier.Modules, "GOCACHEPROG": "repository-backend", "GOLIB_TASK_CACHE_ROOT": session}
				probe := runProcessCacheProbe(t, executor, native, false, false)
				if probe.Poisoned || probe.Build == verifier.Build || probe.Modules == verifier.Modules || probe.Backend != "" || probe.Session != "" {
					t.Error("native repository execution inherited verifier identity")
				}
				if borrowed && (probe.Build != filepath.Join(session, "cache") || probe.Modules != filepath.Join(session, "mod")) {
					t.Error("explicit native producer lost session borrowing")
				}
			}
		})
	}
}

func TestMutationBuildAndCampaignKeepDistinctCacheIdentities(t *testing.T) {
	for _, borrowed := range []bool{false, true} {
		mode := "default"
		if borrowed {
			mode = "borrowed"
		}
		t.Run(mode, func(t *testing.T) {
			root, session := t.TempDir(), processCacheSession(t)
			var created Executor
			var cleanup func() error
			var err error
			if borrowed {
				created, cleanup, err = NewProcessExecutorWithTaskCache(root, session, io.Discard, io.Discard)
			} else {
				created, cleanup, err = NewProcessExecutor(root, io.Discard, io.Discard)
			}
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := cleanup(); err != nil {
					t.Error(err)
				}
			})
			executor := checkedProcessCacheExecutor(t, created)
			workspace := filepath.Join(executor.TemporaryDirectory(), "mutation-campaigns", strings.Repeat("a", 64))
			cache := filepath.Join(workspace, "mutation-cache", "root")
			if err := os.MkdirAll(cache, 0o700); err != nil {
				t.Fatal(err)
			}
			if err := os.MkdirAll(filepath.Join(workspace, "gremlins-source"), 0o700); err != nil {
				t.Fatal(err)
			}
			bin := t.TempDir()
			program := []byte("#!/bin/sh\nexec \"$GOLIB_CACHE_PROBE_BINARY\" -test.run='^TestProcessCacheProbe$'\n")
			for _, path := range []string{filepath.Join(bin, "go"), filepath.Join(workspace, "golib-gremlins")} {
				if err := os.WriteFile(path, program, 0o700); err != nil {
					t.Fatal(err)
				}
			}
			t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
			t.Setenv("GOLIB_CACHE_PROBE_BINARY", os.Args[0])
			process := (Runner{Executor: executor}).mutationProcess(workspace)
			probe := func(name string, args []string, directory, build string, poison bool) processCacheProbe {
				t.Helper()
				report := filepath.Join(t.TempDir(), "probe.json")
				environment := map[string]string{"GOLIB_CACHE_PROBE": "1", "GOLIB_CACHE_PROBE_REPORT": report, "GOLIB_CACHE_PROBE_FAIL": "1", "GOCACHE": build, "GOMODCACHE": filepath.Join(session, "mod"), "GOCACHEPROG": "hostile-backend", "GOLIB_TASK_CACHE_ROOT": session}
				if poison {
					environment["GOLIB_CACHE_PROBE_RUNTIME"] = "1"
				}
				err := process(t.Context(), name, args, directory, environment, io.Discard, io.Discard)
				var status interface{ ExitCode() int }
				if !errors.As(err, &status) || status.ExitCode() != 23 {
					t.Fatalf("mutation process lost exit23: %v", err)
				}
				data, err := os.ReadFile(report)
				if err != nil {
					t.Fatal(err)
				}
				var result processCacheProbe
				if err := json.Unmarshal(data, &result); err != nil {
					t.Fatal(err)
				}
				return result
			}
			campaign := probe(filepath.Join(workspace, "golib-gremlins"), []string{"unleash", "."}, root, cache, true)
			if campaign.Build != cache || campaign.Backend != "" || campaign.Session != "" {
				t.Fatal("campaign lost its package cache or retained a cache backend")
			}
			for _, input := range []struct {
				args []string
				dir  string
			}{
				{[]string{"mod", "download", "-json", "github.com/go-gremlins/gremlins@" + mutation.GremlinsVersion}, workspace},
				{[]string{"build", "-trimpath", "-buildvcs=false", "-o", filepath.Join(workspace, "golib-gremlins"), "./cmd/gremlins"}, filepath.Join(workspace, "gremlins-source")},
			} {
				verifier := probe("go", input.args, input.dir, cache, false)
				if verifier.Poisoned || verifier.Build == campaign.Build || verifier.Modules == campaign.Modules || verifier.Backend != "" || verifier.Session != "" {
					t.Fatal("mutation verifier reused repository campaign caches")
				}
				if filepath.Dir(verifier.Build) != executor.TemporaryDirectory() || filepath.Dir(verifier.Modules) != executor.TemporaryDirectory() {
					t.Fatal("mutation build caches are not invocation-private")
				}
			}
			for _, unsafe := range []string{executor.environment["GOCACHE"], filepath.Join(session, "cache"), filepath.Join(workspace, "mutation-cache", "..", "go-build")} {
				if err := process(t.Context(), filepath.Join(workspace, "golib-gremlins"), []string{"unleash", "."}, root, map[string]string{"GOCACHE": unsafe}, io.Discard, io.Discard); err == nil || !strings.Contains(err.Error(), "owned private package cache") {
					t.Fatalf("unsafe package cache accepted: %v", err)
				}
			}
		})
	}
}

// Real repository tests can modify extracted module-cache source. The pinned
// compiler-tool producer must rebuild from the canonical proxy, not that cache.
func TestCompilerToolRejectsRepositoryPoisonedModuleCache(t *testing.T) {
	for _, mode := range []string{"default", "borrowed", "verifier session"} {
		t.Run(mode, func(t *testing.T) {
			root, proxy := t.TempDir(), t.TempDir()
			const module = "example.invalid/cache-verifier"
			processCacheProxyModule(t, proxy, module, "v1.0.0", map[string]string{
				"go.mod":               "module " + module + "\n\ngo 1.27.0\n",
				"verdict.go":           "package verdict\nfunc Verdict() string { return \"canonical-verifier\" }\n",
				"cmd/verifier/main.go": "package main\nimport (\"fmt\"; verdict \"" + module + "\")\nfunc main() { fmt.Print(verdict.Verdict()) }\n",
			})
			processCacheProxyModule(t, proxy, "golang.org/x/tools", "v0.50.0", map[string]string{"go.mod": "module golang.org/x/tools\n\ngo 1.27.0\n"})
			write := func(name, value string) {
				t.Helper()
				if err := os.WriteFile(filepath.Join(root, name), []byte(value), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			write("go.mod", "module example.invalid/repository\n\ngo 1.27.0\nrequire "+module+" v1.0.0\n")
			write("repository_test.go", `package repository
import ("os"; "path/filepath"; "testing"; verdict "example.invalid/cache-verifier")
func TestRepositoryPoisonsModuleCache(t *testing.T) {
 if verdict.Verdict() != "canonical-verifier" { t.Fatal("fixture was not canonical before execution") }
 source := filepath.Join(os.Getenv("GOMODCACHE"), "example.invalid/cache-verifier@v1.0.0", "verdict.go")
 if err := os.Chmod(source, 0600); err != nil { t.Fatal(err) }
 if err := os.WriteFile(source, []byte("package verdict\nfunc Verdict() string { return \"repository-poison\" }\n"), 0600); err != nil { t.Fatal(err) }
}`)
			t.Setenv("GOPROXY", "file://"+filepath.ToSlash(proxy))
			t.Setenv("GOSUMDB", "off")
			t.Setenv("GOCACHEPROG", "")
			t.Setenv("GOLIB_TASK_CACHE_ROOT", "")
			var created Executor
			var cleanup func() error
			var err error
			if mode != "default" {
				session := processCacheSession(t)
				if mode == "verifier session" {
					for _, name := range []string{"verifier-cache", "verifier-mod", "verifier-tmp"} {
						if err := os.Mkdir(filepath.Join(session, name), 0o700); err != nil {
							t.Fatal(err)
						}
					}
				}
				created, cleanup, err = NewProcessExecutorWithTaskCache(root, session, io.Discard, io.Discard)
			} else {
				created, cleanup, err = NewProcessExecutor(root, io.Discard, io.Discard)
			}
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := cleanup(); err != nil {
					t.Error(err)
				}
			})
			executor := checkedProcessCacheExecutor(t, created)
			executor.prepareBoundedProcess = func(*exec.Cmd) error { return nil }
			runtime, err := operationCommand(root, inventory.Module{}, config.Step{Type: "go-test", Packages: []string{"."}, Run: "^TestRepositoryPoisonsModuleCache$", Count: 1, Timeout: "30s"})
			if err != nil {
				t.Fatal(err)
			}
			runtime.Env = map[string]string{"GOFLAGS": "-mod=mod", "GOWORK": "off"}
			var diagnostics bytes.Buffer
			runtime.Stdout, runtime.Stderr = &diagnostics, &diagnostics
			if err := executor.Run(t.Context(), runtime); err != nil {
				t.Fatalf("native fixture: %v\n%s", err, diagnostics.String())
			}
			tool := compilerTool{name: "cache-verifier", module: module, version: "v1.0.0", command: module + "/cmd/verifier"}
			var verdict bytes.Buffer
			err = (Runner{Executor: executor}).withCompilerTool(t.Context(), tool, func(binary string) error {
				command := exec.CommandContext(t.Context(), binary)
				command.Stdout = &verdict
				return command.Run()
			})
			if err != nil {
				t.Fatal(err)
			}
			if verdict.String() != "canonical-verifier" {
				t.Fatalf("pinned compiler executed %q after repository cache poisoning", verdict.String())
			}
		})
	}
}

func processCacheProxyModule(t *testing.T, proxy, module, version string, files map[string]string) {
	t.Helper()
	root := filepath.Join(proxy, filepath.FromSlash(module), "@v")
	if err := os.MkdirAll(root, 0o700); err != nil {
		t.Fatal(err)
	}
	var archive bytes.Buffer
	writer := zip.NewWriter(&archive)
	for name, contents := range files {
		entry, err := writer.Create(module + "@" + version + "/" + name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := io.Copy(entry, strings.NewReader(contents)); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	for suffix, data := range map[string][]byte{".zip": archive.Bytes(), ".mod": []byte(files["go.mod"]), ".info": []byte(`{"Version":"` + version + `","Time":"2026-01-01T00:00:00Z"}`)} {
		if err := os.WriteFile(filepath.Join(root, version+suffix), data, 0o600); err != nil {
			t.Fatal(err)
		}
	}
}

func runProcessCacheProbe(t *testing.T, executor *processExecutor, command Command, runtime, failure bool) processCacheProbe {
	t.Helper()
	if command.Env == nil {
		command.Env = make(map[string]string)
	}
	report := filepath.Join(t.TempDir(), "cache-probe.json")
	command.Env["GOLIB_CACHE_PROBE"] = "1"
	command.Env["GOLIB_CACHE_PROBE_REPORT"] = report
	if runtime {
		command.Env["GOLIB_CACHE_PROBE_RUNTIME"] = "1"
	}
	if failure {
		command.Env["GOLIB_CACHE_PROBE_FAIL"] = "1"
	}
	err := executor.Run(t.Context(), command)
	if failure {
		var status interface{ ExitCode() int }
		if !errors.As(err, &status) || status.ExitCode() != 23 {
			t.Fatalf("verifier failure was not preserved: %v", err)
		}
	} else if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(report)
	if err != nil {
		t.Fatal(err)
	}
	var probe processCacheProbe
	if err := json.Unmarshal(data, &probe); err != nil {
		t.Fatal(err)
	}
	return probe
}

func processCacheSession(t *testing.T) string {
	t.Helper()
	root, err := os.MkdirTemp("", "golib-cache-session.")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			return os.Chmod(path, info.Mode()|0o200)
		}); err != nil {
			t.Error(err)
		}
		if err := os.RemoveAll(root); err != nil {
			t.Error(err)
		}
	})
	root, err = filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"cache", "mod"} {
		if err := os.Mkdir(filepath.Join(root, name), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func TestTaskVerifierSessionReusesOnlyTrustedCachesAcrossExecutors(t *testing.T) {
	for _, failure := range []bool{false, true} {
		t.Run(map[bool]string{false: "success", true: "failure"}[failure], func(t *testing.T) {
			repository, session := t.TempDir(), processCacheSession(t)
			for _, name := range []string{"verifier-cache", "verifier-mod", "verifier-tmp"} {
				if err := os.Mkdir(filepath.Join(session, name), 0o700); err != nil {
					t.Fatal(err)
				}
			}
			bin := t.TempDir()
			if err := os.WriteFile(filepath.Join(bin, "go"), []byte("#!/bin/sh\nexec \"$GOLIB_CACHE_PROBE_BINARY\" -test.run='^TestProcessCacheProbe$'\n"), 0o700); err != nil {
				t.Fatal(err)
			}
			t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
			t.Setenv("GOLIB_CACHE_PROBE_BINARY", os.Args[0])
			t.Setenv("GOCACHEPROG", "ambient-backend")
			t.Setenv("GOLANGCI_LINT_CACHEPROG", "ambient-lint-backend")
			t.Setenv("GOLIB_TASK_CACHE_ROOT", session)
			var previous processCacheProbe
			for invocation := range 2 {
				created, cleanup, err := NewProcessExecutorWithTaskCache(repository, session, io.Discard, io.Discard)
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() {
					if err := cleanup(); err != nil && !errors.Is(err, os.ErrNotExist) {
						t.Error(err)
					}
				})
				executor := checkedProcessCacheExecutor(t, created)
				executor.prepareBoundedProcess = func(*exec.Cmd) error { return nil }
				for _, bounded := range []bool{false, true} {
					runtime := Command{Name: "go", cacheClass: nativeRepositoryCache, boundedScanner: bounded}
					runtime.Env = map[string]string{"GOCACHE": filepath.Join(session, "verifier-cache"), "GOMODCACHE": filepath.Join(session, "verifier-mod"), "GOTMPDIR": filepath.Join(session, "verifier-tmp")}
					poison := runProcessCacheProbe(t, executor, runtime, true, failure)
					unknown := runProcessCacheProbe(t, executor, Command{Name: "go", Env: runtime.Env, boundedScanner: bounded}, true, failure)
					verifier := processCacheVerifierCommand(t)
					verifier.boundedScanner = bounded
					verifier.Env = map[string]string{"GOCACHE": poison.Build, "GOMODCACHE": unknown.Modules, "GOTMPDIR": poison.Temporary, "GOCACHEPROG": "hostile", "GOLANGCI_LINT_CACHEPROG": "hostile", "GOLIB_TASK_CACHE_ROOT": session}
					verifier.Env["GOBIN"] = filepath.Join(executor.task, "selected-output")
					probe := runProcessCacheProbe(t, executor, verifier, false, failure)
					if probe.Build != filepath.Join(session, "verifier-cache") || probe.Modules != filepath.Join(session, "verifier-mod") || probe.Temporary != filepath.Join(session, "verifier-tmp") {
						t.Error("verifier did not reuse its private caller-session compartment")
					}
					if probe.Poisoned || probe.Backend != "" || probe.LintBackend != "" || probe.Session != "" || probe.Build == poison.Build || probe.Modules == unknown.Modules || probe.Temporary == poison.Temporary {
						t.Error("repository caches or hostile overrides contaminated verifier")
					}
					if filepath.Dir(probe.Lint) != executor.task || filepath.Dir(probe.Output) != executor.task {
						t.Error("lint results or output binaries escaped invocation ownership")
					}
					if probe.Output != verifier.Env["GOBIN"] {
						t.Error("legitimate install-output override was sealed")
					}
					if invocation > 0 && !bounded && (probe.Build != previous.Build || probe.Modules != previous.Modules || probe.Temporary != previous.Temporary || probe.Lint == previous.Lint || probe.Output == previous.Output) {
						t.Error("adjacent executor did not retain only compilation inputs")
					}
					previous = probe
				}
				if err := cleanup(); err != nil {
					t.Fatal(err)
				}
				if _, err := os.Stat(executor.task); !errors.Is(err, os.ErrNotExist) {
					t.Fatal("invocation scratch remains")
				}
				for _, name := range []string{"verifier-cache", "verifier-mod", "verifier-tmp"} {
					if _, err := os.Stat(filepath.Join(session, name)); err != nil {
						t.Fatal("caller compartment removed")
					}
				}
			}
		})
	}
}

func TestTaskVerifierSessionRejectsPartialOrUnsafeDirectories(t *testing.T) {
	for _, scenario := range []string{"partial", "shared", "symlink"} {
		t.Run(scenario, func(t *testing.T) {
			root, session := t.TempDir(), processCacheSession(t)
			for _, name := range []string{"verifier-cache", "verifier-mod", "verifier-tmp"} {
				if scenario == "partial" && name == "verifier-tmp" {
					continue
				}
				if err := os.Mkdir(filepath.Join(session, name), 0o700); err != nil {
					t.Fatal(err)
				}
			}
			if scenario == "shared" {
				if err := os.Chmod(filepath.Join(session, "verifier-mod"), 0o755); err != nil {
					t.Fatal(err)
				}
			}
			if scenario == "symlink" {
				if err := os.Remove(filepath.Join(session, "verifier-tmp")); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(t.TempDir(), filepath.Join(session, "verifier-tmp")); err != nil {
					t.Fatal(err)
				}
			}
			_, cleanup, err := NewProcessExecutorWithTaskCache(root, session, io.Discard, io.Discard)
			if cleanup != nil {
				if err := cleanup(); err != nil {
					t.Fatal(err)
				}
			}
			if err == nil {
				t.Fatal("unsafe verifier session admitted")
			}
		})
	}
}

func TestTaskVerifierSessionsRemainDistinctAndCanceledCallsCleanScratch(t *testing.T) {
	root := t.TempDir()
	var previous map[string]string
	for range 2 {
		session := processCacheSession(t)
		for _, name := range []string{"verifier-cache", "verifier-mod", "verifier-tmp"} {
			if err := os.Mkdir(filepath.Join(session, name), 0o700); err != nil {
				t.Fatal(err)
			}
		}
		created, cleanup, err := NewProcessExecutorWithTaskCache(root, session, io.Discard, io.Discard)
		if err != nil {
			t.Fatal(err)
		}
		executor := checkedProcessCacheExecutor(t, created)
		t.Cleanup(func() {
			if _, err := os.Stat(executor.task); err == nil {
				if err := cleanup(); err != nil {
					t.Error(err)
				}
			}
		})
		command := processCacheVerifierCommand(t)
		actual := make(map[string]string)
		for _, entry := range executor.commandEnvironment(command) {
			name, value, ok := strings.Cut(entry, "=")
			if ok {
				actual[name] = value
			}
		}
		for _, name := range []string{"GOCACHE", "GOMODCACHE", "GOTMPDIR"} {
			if previous != nil && actual[name] == previous[name] {
				t.Fatal("different sessions shared verifier identity")
			}
		}
		ctx, cancel := context.WithCancel(t.Context())
		cancel()
		command.boundedScanner = true
		if err := executor.Run(ctx, command); !errors.Is(err, context.Canceled) {
			t.Fatalf("pre-canceled verifier executed: %v", err)
		}
		if err := cleanup(); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(executor.task); !errors.Is(err, os.ErrNotExist) {
			t.Fatal("canceled scratch remains")
		}
		previous = actual
	}
}
