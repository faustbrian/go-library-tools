package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/faustbrian/go-library-tools/internal/assets"
)

type commandCall struct {
	name string
	args []string
	env  []string
}

func validArguments(t *testing.T) ([]string, string, string, string) {
	t.Helper()
	root := t.TempDir()
	repository := filepath.Join(root, "repository")
	schema := filepath.Join(repository, "schema")
	workspace := filepath.Join(root, "workspace")
	proxy := filepath.Join(workspace, "proxy")
	runner := filepath.Join(workspace, "runner")
	if err := os.MkdirAll(schema, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(proxy, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(runner, 0o755); err != nil {
		t.Fatal(err)
	}
	args := []string{
		"--repository", repository,
		"--schema-root", schema,
		"--workspace-root", workspace,
		"--task-root", filepath.Join(workspace, "task"),
		"--output", filepath.Join(workspace, "output"),
		"--go", filepath.Join(root, "go"),
		"--git", filepath.Join(root, "git"),
		"--module-proxy", proxy,
		"--runner-templates", runner,
	}
	return args, repository, filepath.Join(workspace, "task"), filepath.Join(workspace, "output")
}

func successfulCommand(repository string, calls *[]commandCall) commandFunc {
	return func(_ context.Context, name string, args []string, env []string) ([]byte, error) {
		*calls = append(*calls, commandCall{name: name, args: append([]string(nil), args...), env: append([]string(nil), env...)})
		if len(args) > 0 && args[0] == "version" {
			return []byte("go version go1.27.0 darwin/arm64\n"), nil
		}
		return []byte(repository + "\n"), nil
	}
}

func TestRunSuccessUsesInjectedBoundariesAndSortsJSON(t *testing.T) {
	args, repository, task, output := validArguments(t)
	var calls []commandCall
	var generated assets.Generator
	var stdout, stderr bytes.Buffer
	code := run(args, &stdout, &stderr, func(g assets.Generator) (map[string]string, error) {
		generated = g
		return map[string]string{"z": "last", "a": "first"}, nil
	}, successfulCommand(repository, &calls))
	if code != 0 || stderr.Len() != 0 || stdout.String() != `{"a":"first","z":"last"}`+"\n" {
		t.Fatalf("run() = %d, stdout=%q, stderr=%q", code, stdout.String(), stderr.String())
	}
	if generated.Repository != repository || generated.TaskRoot != task || generated.OutputDir != output || len(generated.RunnerTemplateDirs) != 1 {
		t.Fatalf("generator = %#v", generated)
	}
	if len(calls) != 2 || calls[0].args[len(calls[0].args)-1] != "--show-toplevel" || len(calls[1].args) != 1 || calls[1].args[0] != "version" {
		t.Fatalf("commands = %#v", calls)
	}
}

func TestRunValidationAndExecutionFailures(t *testing.T) {
	tests := []struct {
		name     string
		mutate   func([]string, string, string, string) []string
		command  func(string, *[]commandCall) commandFunc
		generate generatorFunc
		want     string
	}{
		{"unknown flag", func(args []string, _, _, _ string) []string { return append(args, "--unknown") }, nil, nil, "flag provided but not defined"},
		{"output escape", func(args []string, _, _, _ string) []string {
			replaceArgument(args, "--output", "/outside")
			return args
		}, nil, nil, "output:"},
		{"task escape", func(args []string, _, _, _ string) []string {
			replaceArgument(args, "--task-root", "/outside")
			return args
		}, nil, nil, "task root:"},
		{"schema escape", func(args []string, _, _, _ string) []string {
			replaceArgument(args, "--schema-root", "/outside")
			return args
		}, nil, nil, "schema root:"},
		{"overlap", func(args []string, _, _, _ string) []string {
			replaceArgument(args, "--task-root", filepath.Join("/tmp", "output", "task"))
			replaceArgument(args, "--output", "/tmp/output")
			replaceArgument(args, "--workspace-root", "/tmp")
			return args
		}, nil, nil, "must not overlap"},
		{"git failure", func(args []string, _, _, _ string) []string { return args }, func(_ string, _ *[]commandCall) commandFunc {
			return func(context.Context, string, []string, []string) ([]byte, error) {
				return nil, errors.New("git failed")
			}
		}, nil, "exact Git checkout root"},
		{"git wrong root", func(args []string, _, _, _ string) []string { return args }, func(_ string, _ *[]commandCall) commandFunc {
			return func(context.Context, string, []string, []string) ([]byte, error) { return []byte("other\n"), nil }
		}, nil, "exact Git checkout root"},
		{"proxy missing", func(args []string, _, _, workspace string) []string {
			replaceArgument(args, "--module-proxy", filepath.Join(workspace, "missing"))
			return args
		}, nil, nil, "module proxy"},
		{"runner missing", func(args []string, _, _, workspace string) []string {
			replaceArgument(args, "--runner-templates", filepath.Join(workspace, "missing"))
			return args
		}, nil, nil, "runner template root"},
		{"go failure", func(args []string, _, _, _ string) []string { return args }, func(repository string, calls *[]commandCall) commandFunc {
			base := successfulCommand(repository, calls)
			return func(ctx context.Context, name string, args, env []string) ([]byte, error) {
				if len(args) > 0 && args[0] == "version" {
					return nil, errors.New("go failed")
				}
				return base(ctx, name, args, env)
			}
		}, nil, "Go version"},
		{"go malformed", func(args []string, _, _, _ string) []string { return args }, func(repository string, calls *[]commandCall) commandFunc {
			base := successfulCommand(repository, calls)
			return func(ctx context.Context, name string, args, env []string) ([]byte, error) {
				if len(args) > 0 && args[0] == "version" {
					return []byte("go version\n"), nil
				}
				return base(ctx, name, args, env)
			}
		}, nil, "Go version ="},
		{"go mismatch", func(args []string, _, _, _ string) []string { return append(args, "--expected-go-version", "go9.9.9") }, nil, nil, "Go version ="},
		{"generation failure", func(args []string, _, _, _ string) []string { return args }, nil, func(assets.Generator) (map[string]string, error) { return nil, errors.New("generation failed") }, "generate:"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			args, repository, _, _ := validArguments(t)
			args = test.mutate(args, repository, "", filepath.Dir(filepath.Dir(repository)))
			var calls []commandCall
			execute := test.command
			if execute == nil {
				execute = successfulCommand
			}
			var stderr bytes.Buffer
			generate := test.generate
			if generate == nil {
				generate = func(assets.Generator) (map[string]string, error) { return map[string]string{}, nil }
			}
			if code := run(args, &bytes.Buffer{}, &stderr, generate, execute(repository, &calls)); code == 0 || !strings.Contains(stderr.String(), test.want) {
				t.Fatalf("run() = %d, stderr=%q", code, stderr.String())
			}
		})
	}
}

func replaceArgument(args []string, flagName, value string) {
	for i := range args {
		if args[i] == flagName && i+1 < len(args) {
			args[i+1] = value
			return
		}
	}
}

func TestRunReportsRequiredAndPathListFailures(t *testing.T) {
	var stderr bytes.Buffer
	if code := run(nil, &bytes.Buffer{}, &stderr, nil, nil); code != 1 || !strings.Contains(stderr.String(), "required") {
		t.Fatalf("missing args: code=%d stderr=%q", code, stderr.String())
	}
	args, _, _, _ := validArguments(t)
	replaceArgument(args, "--runner-templates", "")
	stderr.Reset()
	if code := run(args, &bytes.Buffer{}, &stderr, nil, nil); code != 1 || !strings.Contains(stderr.String(), "runner-templates") {
		t.Fatalf("empty runner list: code=%d stderr=%q", code, stderr.String())
	}
}

func TestRunNormalizesEveryPathAndReportsNormalizationFailures(t *testing.T) {
	fields := []string{"--repository", "--schema-root", "--output", "--task-root", "--workspace-root", "--go", "--git", "--module-proxy", "--runner-templates"}
	for fieldIndex, field := range fields {
		t.Run(field, func(t *testing.T) {
			args, _, _, _ := validArguments(t)
			original := absolutePath
			calls := 0
			absolutePath = func(path string) (string, error) {
				calls++
				if calls == fieldIndex+1 {
					return "", errors.New("normalization failed")
				}
				return filepath.Abs(path)
			}
			t.Cleanup(func() { absolutePath = original })
			var stderr bytes.Buffer
			if code := run(args, &bytes.Buffer{}, &stderr, nil, nil); code != 1 || !strings.Contains(stderr.String(), "absolute path") {
				t.Fatalf("run() = %d, stderr=%q", code, stderr.String())
			}
		})
	}
}

func TestMainUsesInjectedProcessBoundaries(t *testing.T) {
	args, repository, _, _ := validArguments(t)
	oldArgs, oldOut, oldErr, oldExit, oldGenerate, oldCommand := commandArguments, standardOutput, standardError, exitProcess, generateAssets, runCommand
	t.Cleanup(func() {
		commandArguments, standardOutput, standardError, exitProcess, generateAssets, runCommand = oldArgs, oldOut, oldErr, oldExit, oldGenerate, oldCommand
	})
	var stdout, stderr bytes.Buffer
	exitCode := -1
	commandArguments = func() []string { return args }
	standardOutput, standardError = &stdout, &stderr
	exitProcess = func(code int) { exitCode = code }
	generateAssets = func(g assets.Generator) (map[string]string, error) {
		if g.Repository != repository {
			t.Fatalf("repository = %q", g.Repository)
		}
		return map[string]string{"asset": "digest"}, nil
	}
	runCommand = successfulCommand(repository, new([]commandCall))
	main()
	if exitCode != 0 || stdout.String() != `{"asset":"digest"}`+"\n" || stderr.Len() != 0 {
		t.Fatalf("main() exit=%d stdout=%q stderr=%q", exitCode, stdout.String(), stderr.String())
	}
}

func TestRunReportsTaskDirectoryCreationFailure(t *testing.T) {
	args, repository, _, _ := validArguments(t)
	root := filepath.Dir(filepath.Dir(repository))
	file := filepath.Join(root, "task-file")
	if err := os.WriteFile(file, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	replaceArgument(args, "--workspace-root", root)
	replaceArgument(args, "--task-root", filepath.Join(file, "child"))
	var stderr bytes.Buffer
	if code := run(args, &bytes.Buffer{}, &stderr, func(assets.Generator) (map[string]string, error) { return nil, nil }, successfulCommand(repository, new([]commandCall))); code != 1 || !strings.Contains(stderr.String(), "task root") {
		t.Fatalf("mkdir failure: code=%d stderr=%q", code, stderr.String())
	}
}

func TestDefaultProcessSeamsRemainCallable(t *testing.T) {
	if len(commandArguments()) != len(os.Args)-1 {
		t.Fatal("commandArguments did not read os.Args")
	}
	if _, err := generateAssets(assets.Generator{}); err == nil {
		t.Fatal("default generator unexpectedly accepted empty paths")
	}
}

func TestRunExternalCommand(t *testing.T) {
	output, err := runExternalCommand(context.Background(), "/bin/echo", []string{"ok"}, []string{"PATH=/bin"})
	if err != nil || string(output) != "ok\n" {
		t.Fatalf("output=%q err=%v", output, err)
	}
	if _, err := runExternalCommand(context.Background(), "/does/not/exist", nil, nil); err == nil {
		t.Fatal("missing executable succeeded")
	}
}

func TestAbsoluteAndWithinReportInjectedPathErrors(t *testing.T) {
	oldAbsolute, oldExit, oldRelative := absolutePath, exitProcess, relativePath
	t.Cleanup(func() { absolutePath, exitProcess, relativePath = oldAbsolute, oldExit, oldRelative })
	absolutePath = func(string) (string, error) { return "", errors.New("absolute failed") }
	exitCode := 0
	exitProcess = func(code int) { exitCode = code }
	if got := absolute("relative"); got != "." || exitCode != 1 {
		t.Fatalf("absolute() = %q, exit=%d", got, exitCode)
	}
	absolutePath = filepath.Abs
	relativePath = func(string, string) (string, error) { return "", errors.New("relative failed") }
	if err := within("root", "target"); err == nil || err.Error() != "relative failed" {
		t.Fatalf("within() = %v", err)
	}
}

func TestSplitAbsolutePathListReportsInjectedPathErrors(t *testing.T) {
	oldAbsolute := absolutePath
	t.Cleanup(func() { absolutePath = oldAbsolute })
	absolutePath = func(string) (string, error) { return "", errors.New("absolute failed") }
	if _, err := splitAbsolutePathListE("entry"); err == nil || !strings.Contains(err.Error(), "absolute path") {
		t.Fatalf("split error = %v", err)
	}
}

func TestWithinRejectsEscapesAndAcceptsDescendants(t *testing.T) {
	t.Parallel()
	if err := within("/workspace", "/workspace/assets"); err != nil {
		t.Fatal(err)
	}
	if err := within("/workspace", "/outside"); err == nil {
		t.Fatal("escape accepted")
	}
}

func TestPathContainsRequiresActualDescendant(t *testing.T) {
	t.Parallel()
	if !pathContains("/workspace", "/workspace/task") {
		t.Fatal("descendant not detected")
	}
	if pathContains("/workspace/task", "/workspace") {
		t.Fatal("ancestor detected as descendant")
	}
}

func TestSplitAbsolutePathListNormalizesAndSkipsEmptyEntries(t *testing.T) {
	paths := splitAbsolutePathList(strings.Join([]string{"relative", "", "nested"}, string(filepath.ListSeparator)))
	if len(paths) != 2 || !filepath.IsAbs(paths[0]) || !filepath.IsAbs(paths[1]) {
		t.Fatalf("splitAbsolutePathList() = %#v", paths)
	}
}

func TestSplitAbsolutePathListAndFatalfUseInjectableExit(t *testing.T) {
	old := exitProcess
	defer func() { exitProcess = old }()
	called := 0
	exitProcess = func(code int) { called = code }
	if got := splitAbsolutePathList(""); len(got) != 0 {
		t.Fatalf("splitAbsolutePathList(empty) = %#v", got)
	}
	fatalf("failure %s", "case")
	if called != 1 {
		t.Fatalf("exit code = %d, want 1", called)
	}
}
