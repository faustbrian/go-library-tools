package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/faustbrian/go-library-tools/internal/assets"
)

func main() {
	var repository, schemaRoot, output, taskRoot, workspaceRoot, goBinary, gitBinary, expectedGo, moduleProxy, runnerTemplates string
	flag.StringVar(&repository, "repository", "", "read-only go-library-tools Git repository")
	flag.StringVar(&schemaRoot, "schema-root", "", "read-only root containing schema candidates")
	flag.StringVar(&output, "output", "", "asset output directory")
	flag.StringVar(&taskRoot, "task-root", "", "disposable checkout and cache directory")
	flag.StringVar(&workspaceRoot, "workspace-root", "", "root confining output and task paths")
	flag.StringVar(&goBinary, "go", "", "absolute Go executable path")
	flag.StringVar(&gitBinary, "git", "", "absolute Git executable path")
	flag.StringVar(&moduleProxy, "module-proxy", "", "read-only local Go module proxy root")
	flag.StringVar(&runnerTemplates, "runner-templates", "", "read-only directory containing historical observation runners")
	flag.StringVar(&expectedGo, "expected-go-version", "go1.27.0", "required exact Go version token")
	flag.Parse()
	for name, value := range map[string]string{"repository": repository, "schema-root": schemaRoot, "output": output, "task-root": taskRoot, "workspace-root": workspaceRoot, "go": goBinary, "git": gitBinary, "module-proxy": moduleProxy, "runner-templates": runnerTemplates} {
		if value == "" {
			fatalf("--%s is required", name)
		}
	}
	repository, schemaRoot, output = absolute(repository), absolute(schemaRoot), absolute(output)
	taskRoot, workspaceRoot, goBinary, gitBinary, moduleProxy = absolute(taskRoot), absolute(workspaceRoot), absolute(goBinary), absolute(gitBinary), absolute(moduleProxy)
	runnerTemplateDirs := splitAbsolutePathList(runnerTemplates)
	if err := within(workspaceRoot, output); err != nil {
		fatalf("output: %v", err)
	}
	if err := within(workspaceRoot, taskRoot); err != nil {
		fatalf("task root: %v", err)
	}
	if err := within(repository, schemaRoot); err != nil {
		fatalf("schema root: %v", err)
	}
	if output == taskRoot || pathContains(output, taskRoot) || pathContains(taskRoot, output) {
		fatalf("task root and output must not overlap")
	}
	rootContext, cancelRoot := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelRoot()
	rootCommand := exec.CommandContext(rootContext, gitBinary, "-C", repository, "rev-parse", "--show-toplevel")
	rootCommand.Env = []string{"PATH=" + filepath.Dir(gitBinary), "LANG=C.UTF-8", "LC_ALL=C.UTF-8"}
	rootOutput, err := rootCommand.Output()
	if err != nil || filepath.Clean(strings.TrimSpace(string(rootOutput))) != repository {
		fatalf("repository is not the exact Git checkout root")
	}
	if info, err := os.Stat(moduleProxy); err != nil || !info.IsDir() {
		fatalf("module proxy is not a directory")
	}
	for _, runnerTemplateDir := range runnerTemplateDirs {
		if info, err := os.Stat(runnerTemplateDir); err != nil || !info.IsDir() {
			fatalf("runner template root is not a directory: %s", runnerTemplateDir)
		}
	}
	versionContext, cancelVersion := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelVersion()
	versionCommand := exec.CommandContext(versionContext, goBinary, "version")
	versionCommand.Env = []string{"PATH=" + filepath.Dir(goBinary), "LANG=C.UTF-8", "LC_ALL=C.UTF-8", "GOTOOLCHAIN=local", "GOWORK=off", "GOENV=off"}
	versionOutput, err := versionCommand.Output()
	if err != nil {
		fatalf("Go version: %v", err)
	}
	fields := strings.Fields(string(versionOutput))
	if len(fields) < 3 || fields[2] != expectedGo {
		fatalf("Go version = %q, want %s", strings.TrimSpace(string(versionOutput)), expectedGo)
	}
	if err := os.MkdirAll(taskRoot, 0o700); err != nil {
		fatalf("task root: %v", err)
	}
	digests, err := (assets.Generator{Repository: repository, SchemaDir: schemaRoot, OutputDir: output, TaskRoot: taskRoot, GoBinary: goBinary, GitBinary: gitBinary, ModuleProxy: moduleProxy, WorkspaceRoot: workspaceRoot, RunnerTemplateDirs: runnerTemplateDirs}).Generate()
	if err != nil {
		fatalf("generate: %v", err)
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
	fmt.Println(string(encoded))
}

func absolute(path string) string {
	result, err := filepath.Abs(path)
	if err != nil {
		fatalf("absolute path: %v", err)
	}
	return filepath.Clean(result)
}
func splitAbsolutePathList(value string) []string {
	parts := filepath.SplitList(value)
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			result = append(result, absolute(part))
		}
	}
	if len(result) == 0 {
		fatalf("--runner-templates must contain at least one directory")
	}
	return result
}
func within(root, target string) error {
	relative, err := filepath.Rel(root, target)
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

var exitProcess = os.Exit
