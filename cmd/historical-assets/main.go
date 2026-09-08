package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/faustbrian/go-library-tools/internal/assets"
)

type generatorFunc func(assets.Generator) (map[string]string, error)
type commandFunc func(context.Context, string, []string, []string) ([]byte, error)

var (
	commandArguments               = func() []string { return os.Args[1:] }
	standardOutput   io.Writer     = os.Stdout
	standardError    io.Writer     = os.Stderr
	exitProcess                    = os.Exit
	absolutePath                   = filepath.Abs
	runCommand       commandFunc   = runExternalCommand
	relativePath                   = filepath.Rel
	generateAssets   generatorFunc = func(generator assets.Generator) (map[string]string, error) {
		return generator.Generate()
	}
)

func main() {
	exitProcess(run(commandArguments(), standardOutput, standardError, generateAssets, runCommand))
}

func run(args []string, stdout, stderr io.Writer, generate generatorFunc, execute commandFunc) int {
	options, err := parseOptions(args, stderr)
	if err != nil {
		return 2
	}
	for name, value := range map[string]string{
		"repository":       options.repository,
		"schema-root":      options.schemaRoot,
		"output":           options.output,
		"task-root":        options.taskRoot,
		"workspace-root":   options.workspaceRoot,
		"go":               options.goBinary,
		"git":              options.gitBinary,
		"module-proxy":     options.moduleProxy,
		"runner-templates": options.runnerTemplates,
	} {
		if value == "" {
			return reportError(stderr, "--%s is required", name)
		}
	}

	normalize := func(name, value string) (string, bool) {
		result, normalizeErr := absolutePath(value)
		if normalizeErr != nil {
			reportError(stderr, "%s: absolute path: %v", name, normalizeErr)
			return "", false
		}
		return filepath.Clean(result), true
	}
	var ok bool
	if options.repository, ok = normalize("repository", options.repository); !ok {
		return 1
	}
	if options.schemaRoot, ok = normalize("schema root", options.schemaRoot); !ok {
		return 1
	}
	if options.output, ok = normalize("output", options.output); !ok {
		return 1
	}
	if options.taskRoot, ok = normalize("task root", options.taskRoot); !ok {
		return 1
	}
	if options.workspaceRoot, ok = normalize("workspace root", options.workspaceRoot); !ok {
		return 1
	}
	if options.goBinary, ok = normalize("go", options.goBinary); !ok {
		return 1
	}
	if options.gitBinary, ok = normalize("git", options.gitBinary); !ok {
		return 1
	}
	if options.moduleProxy, ok = normalize("module proxy", options.moduleProxy); !ok {
		return 1
	}
	runnerTemplateDirs, splitErr := splitAbsolutePathListE(options.runnerTemplates)
	if splitErr != nil {
		return reportError(stderr, "%v", splitErr)
	}
	if err := within(options.workspaceRoot, options.output); err != nil {
		return reportError(stderr, "output: %v", err)
	}
	if err := within(options.workspaceRoot, options.taskRoot); err != nil {
		return reportError(stderr, "task root: %v", err)
	}
	if err := within(options.repository, options.schemaRoot); err != nil {
		return reportError(stderr, "schema root: %v", err)
	}
	if options.output == options.taskRoot || pathContains(options.output, options.taskRoot) || pathContains(options.taskRoot, options.output) {
		return reportError(stderr, "task root and output must not overlap")
	}

	rootContext, cancelRoot := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelRoot()
	rootOutput, err := execute(rootContext, options.gitBinary, []string{"-C", options.repository, "rev-parse", "--show-toplevel"}, []string{"PATH=" + filepath.Dir(options.gitBinary), "LANG=C.UTF-8", "LC_ALL=C.UTF-8"})
	if err != nil || filepath.Clean(strings.TrimSpace(string(rootOutput))) != options.repository {
		return reportError(stderr, "repository is not the exact Git checkout root")
	}
	if info, statErr := os.Stat(options.moduleProxy); statErr != nil || !info.IsDir() {
		return reportError(stderr, "module proxy is not a directory")
	}
	for _, runnerTemplateDir := range runnerTemplateDirs {
		if info, statErr := os.Stat(runnerTemplateDir); statErr != nil || !info.IsDir() {
			return reportError(stderr, "runner template root is not a directory: %s", runnerTemplateDir)
		}
	}
	versionContext, cancelVersion := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelVersion()
	versionOutput, err := execute(versionContext, options.goBinary, []string{"version"}, []string{"PATH=" + filepath.Dir(options.goBinary), "LANG=C.UTF-8", "LC_ALL=C.UTF-8", "GOTOOLCHAIN=local", "GOWORK=off", "GOENV=off"})
	if err != nil {
		return reportError(stderr, "Go version: %v", err)
	}
	fields := strings.Fields(string(versionOutput))
	if len(fields) < 3 || fields[2] != options.expectedGo {
		return reportError(stderr, "Go version = %q, want %s", strings.TrimSpace(string(versionOutput)), options.expectedGo)
	}
	if err := os.MkdirAll(options.taskRoot, 0o700); err != nil {
		return reportError(stderr, "task root: %v", err)
	}
	digests, err := generate(assets.Generator{Repository: options.repository, SchemaDir: options.schemaRoot, OutputDir: options.output, TaskRoot: options.taskRoot, GoBinary: options.goBinary, GitBinary: options.gitBinary, ModuleProxy: options.moduleProxy, WorkspaceRoot: options.workspaceRoot, RunnerTemplateDirs: runnerTemplateDirs})
	if err != nil {
		return reportError(stderr, "generate: %v", err)
	}
	names := make([]string, 0, len(digests))
	for name := range digests {
		names = append(names, name)
	}
	sort.Strings(names)
	ordered := make(map[string]string, len(names))
	for _, name := range names {
		ordered[name] = digests[name]
	}
	encoded, _ := json.Marshal(ordered)
	_, _ = fmt.Fprintln(stdout, string(encoded))
	return 0
}

type options struct {
	repository, schemaRoot, output, taskRoot, workspaceRoot string
	goBinary, gitBinary, moduleProxy, runnerTemplates       string
	expectedGo                                              string
}

func parseOptions(args []string, stderr io.Writer) (options, error) {
	var result options
	flags := flag.NewFlagSet("historical-assets", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.StringVar(&result.repository, "repository", "", "read-only go-library-tools Git repository")
	flags.StringVar(&result.schemaRoot, "schema-root", "", "read-only root containing schema candidates")
	flags.StringVar(&result.output, "output", "", "asset output directory")
	flags.StringVar(&result.taskRoot, "task-root", "", "disposable checkout and cache directory")
	flags.StringVar(&result.workspaceRoot, "workspace-root", "", "root confining output and task paths")
	flags.StringVar(&result.goBinary, "go", "", "absolute Go executable path")
	flags.StringVar(&result.gitBinary, "git", "", "absolute Git executable path")
	flags.StringVar(&result.moduleProxy, "module-proxy", "", "read-only local Go module proxy root")
	flags.StringVar(&result.runnerTemplates, "runner-templates", "", "read-only directory containing historical observation runners")
	flags.StringVar(&result.expectedGo, "expected-go-version", "go1.27.0", "required exact Go version token")
	if err := flags.Parse(args); err != nil {
		return options{}, err
	}
	return result, nil
}

func runExternalCommand(ctx context.Context, name string, args []string, env []string) ([]byte, error) {
	command := exec.CommandContext(ctx, name, args...)
	command.Env = env
	return command.Output()
}

func reportError(stderr io.Writer, format string, args ...interface{}) int {
	_, _ = fmt.Fprintf(stderr, format+"\n", args...)
	return 1
}

func absolute(path string) string {
	result, err := absolutePath(path)
	if err != nil {
		fatalf("absolute path: %v", err)
	}
	return filepath.Clean(result)
}

func splitAbsolutePathList(value string) []string {
	result, err := splitAbsolutePathListE(value)
	if err != nil {
		fatalf("%v", err)
	}
	return result
}

func splitAbsolutePathListE(value string) ([]string, error) {
	parts := filepath.SplitList(value)
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if part == "" {
			continue
		}
		path, err := absolutePath(part)
		if err != nil {
			return nil, fmt.Errorf("absolute path: %w", err)
		}
		result = append(result, filepath.Clean(path))
	}
	if len(result) == 0 {
		return nil, errors.New("--runner-templates must contain at least one directory")
	}
	return result, nil
}

func within(root, target string) error {
	relative, err := relativePath(root, target)
	if err != nil {
		return err
	}
	if relative == ".." || filepath.IsAbs(relative) || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return fmt.Errorf("%s escapes %s", target, root)
	}
	return nil
}

func pathContains(parent, child string) bool {
	relative, err := filepath.Rel(parent, child)
	return err == nil && relative != ".." && !filepath.IsAbs(relative) && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
}

func fatalf(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(os.Stderr, format+"\n", args...)
	exitProcess(1)
}
