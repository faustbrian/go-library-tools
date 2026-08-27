// Package gates orchestrates repository checks from canonical module policy.
package gates

import (
	"bytes"
	"context"
	"fmt"
	"go/format"
	"go/parser"
	"go/token"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/faustbrian/go-library-tools/internal/config"
	"github.com/faustbrian/go-library-tools/internal/coverage"
	"github.com/faustbrian/go-library-tools/internal/docscheck"
	"github.com/faustbrian/go-library-tools/internal/inventory"
	"github.com/faustbrian/go-library-tools/internal/repositoryfile"
	"github.com/faustbrian/go-library-tools/internal/services"
	"golang.org/x/mod/module"
)

const maximumMakefileSize = 4 << 20

const (
	golangCILintVersion = "v2.12.2"
	staticcheckVersion  = "v0.8.1"
	nilAwayVersion      = "v0.0.0-20260720194628-9fd1b8d7bac8"
	govulncheckVersion  = "v1.6.0"
	gitleaksVersion     = "v8.30.1"
	goLicensesVersion   = "v2.0.1"
	cycloneDXVersion    = "v1.10.0"
)

// Command is one external process invocation without shell interpretation.
type Command struct {
	Name   string
	Args   []string
	Dir    string
	Env    map[string]string
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

// Executor runs one external command.
type Executor interface {
	Run(context.Context, Command) error
}

type taskWorkspace interface {
	TemporaryDirectory() string
}

// Runner executes gates for modules in one validated repository.
type Runner struct {
	Root              string
	Catalog           inventory.Inventory
	Policy            config.Config
	Executor          Executor
	Output            io.Writer
	coverageFiles     coverageFileSystem
	apiFiles          apiFileSystem
	apiReadBaseline   func(string, string, int64) ([]byte, error)
	releaseFiles      releaseFileSystem
	releaseArchive    func(io.Writer, module.Version, string) error
	mutationFiles     mutationFileSystem
	mutationCampaign  mutationCampaignRunner
	mutationImport    mutationImportRunner
	startServices     serviceStarter
	serviceHTTPProbe  services.HTTPProbe
	serviceIdentities map[string]string
	// DocumentationSpelling is an isolated test boundary. Production callers
	// leave it nil and use the pinned task-owned implementation.
	DocumentationSpelling func(context.Context, string) error
	// DocumentationLinks is an isolated test boundary. Production callers
	// leave it nil and use the checksum-pinned task-owned implementation.
	DocumentationLinks   func(context.Context, string) error
	documentationRelease func(string, string) (docscheck.LycheeRelease, error)
	documentationExtract func(string, docscheck.LycheeRelease) ([]byte, error)
}

type namedWriteCloser interface {
	io.WriteCloser
	Name() string
}

type coverageFileSystem interface {
	CreateTemp(string) (namedWriteCloser, error)
	Open(string) (io.ReadCloser, error)
	Remove(string) error
}

type operatingCoverageFiles struct{}

func (operatingCoverageFiles) CreateTemp(directory string) (namedWriteCloser, error) {
	return os.CreateTemp(directory, "golib-coverage-*.out")
}

func (operatingCoverageFiles) Open(path string) (io.ReadCloser, error) {
	return os.Open(path)
}

func (operatingCoverageFiles) Remove(path string) error {
	return os.Remove(path)
}

// Check runs the standard contract for each explicitly selected module.
func (runner Runner) Check(ctx context.Context, selection []string) error {
	modules, err := runner.selectModules(selection)
	if err != nil {
		return err
	}
	output := runner.Output
	if output == nil {
		output = io.Discard
	}
	for _, module := range modules {
		if err := runner.withModuleServices(ctx, module, func(scoped Runner) error {
			return scoped.checkModule(ctx, output, module)
		}); err != nil {
			return err
		}
	}
	return nil
}

// Coverage runs only exact production-package coverage for selected modules.
func (runner Runner) Coverage(ctx context.Context, selection []string) error {
	modules, err := runner.selectModules(selection)
	if err != nil {
		return err
	}
	output := runner.Output
	if output == nil {
		output = io.Discard
	}
	for _, module := range modules {
		if !module.Gates["coverage"] {
			_, _ = fmt.Fprintf(output, "[%s] coverage: not applicable\n", module.Directory)
			continue
		}
		if err := runner.withModuleServices(ctx, module, func(scoped Runner) error {
			directory := filepath.Join(scoped.Root, module.Directory)
			return announce(output, module.Directory, "coverage", func() error {
				return scoped.runCoverage(ctx, output, directory, module)
			})
		}); err != nil {
			return err
		}
	}
	return nil
}

// Docs validates Markdown navigation and any additional typed documentation
// operation for selected modules.
func (runner Runner) Docs(ctx context.Context, selection []string) error {
	modules, err := runner.selectModules(selection)
	if err != nil {
		return err
	}
	output := runner.Output
	if output == nil {
		output = io.Discard
	}
	for _, module := range modules {
		if !module.Gates["documentation"] {
			_, _ = fmt.Fprintf(output, "[%s] docs: not applicable\n", module.Directory)
			continue
		}
		directory := filepath.Join(runner.Root, module.Directory)
		if err := announce(output, module.Directory, "docs", func() error {
			return runner.checkDocumentation(ctx, directory, module)
		}); err != nil {
			return err
		}
	}
	return nil
}

func (runner Runner) selectModules(selection []string) ([]inventory.Module, error) {
	available := make(map[string]inventory.Module, len(runner.Catalog.Modules))
	for _, module := range runner.Catalog.Modules {
		available[module.Directory] = module
	}
	unique := make(map[string]struct{}, len(selection))
	for _, directory := range selection {
		if _, ok := available[directory]; !ok {
			return nil, fmt.Errorf("unknown module: %s", directory)
		}
		unique[directory] = struct{}{}
	}
	directories := make([]string, 0, len(unique))
	for directory := range unique {
		directories = append(directories, directory)
	}
	sort.Strings(directories)
	modules := make([]inventory.Module, 0, len(directories))
	for _, directory := range directories {
		modules = append(modules, available[directory])
	}
	return modules, nil
}

func (runner Runner) checkModule(ctx context.Context, output io.Writer, module inventory.Module) error {
	directory := filepath.Join(runner.Root, module.Directory)
	if err := announce(output, module.Directory, "format-check", func() error {
		return checkFormatting(directory)
	}); err != nil {
		return err
	}
	if err := runner.command(ctx, output, module.Directory, "tidy-check", directory, "mod", "tidy", "-diff"); err != nil {
		return err
	}
	if err := announce(output, module.Directory, "safety", func() error {
		return checkSafety(directory)
	}); err != nil {
		return err
	}
	if module.Gates["lint"] {
		if err := runner.command(ctx, output, module.Directory, "vet", directory, "vet", "./..."); err != nil {
			return err
		}
	}
	if module.Gates["tests"] {
		args := testArguments(module.TestTags, false)
		if err := runner.command(ctx, output, module.Directory, "test", directory, args...); err != nil {
			return err
		}
		if operation, exists := runner.operation(module.Directory, "test"); exists {
			if err := runner.runOperation(ctx, directory, module, operation); err != nil {
				return err
			}
		}
	}
	if module.Gates["race"] {
		args := testArguments(module.TestTags, true)
		if err := runner.command(ctx, output, module.Directory, "race", directory, args...); err != nil {
			return err
		}
	}
	if module.Gates["coverage"] {
		if err := announce(output, module.Directory, "coverage", func() error {
			return runner.runCoverage(ctx, output, directory, module)
		}); err != nil {
			return err
		}
	}
	if module.Gates["mutation"] {
		if err := announce(output, module.Directory, "mutation", func() error {
			return runner.runMutation(ctx, output, module)
		}); err != nil {
			return err
		}
	}
	if module.Gates["lint"] {
		if err := runner.goTool(ctx, output, module.Directory, "lint", directory,
			"github.com/golangci/golangci-lint/v2/cmd/golangci-lint@"+golangCILintVersion,
			"run", "--allow-parallel-runners", "--timeout=10m", "./..."); err != nil {
			return err
		}
		if err := runner.goTool(ctx, output, module.Directory, "staticcheck", directory,
			"honnef.co/go/tools/cmd/staticcheck@"+staticcheckVersion, "./..."); err != nil {
			return err
		}
	}
	if module.Gates["security"] {
		if err := runner.goTool(ctx, output, module.Directory, "vulnerability", directory,
			"golang.org/x/vuln/cmd/govulncheck@"+govulncheckVersion, "./..."); err != nil {
			return err
		}
		if err := runner.goTool(ctx, output, module.Directory, "secrets", directory,
			"github.com/zricethezav/gitleaks/v8@"+gitleaksVersion,
			"dir", ".", "--config", filepath.Join(runner.Root, ".gitleaks.toml"), "--no-banner", "--redact"); err != nil {
			return err
		}
		if err := runner.goTool(ctx, output, module.Directory, "licenses", directory,
			"github.com/google/go-licenses/v2@"+goLicensesVersion,
			"check", "./...", "--ignore", module.ModulePath); err != nil {
			return err
		}
		if err := announce(output, module.Directory, "SBOM", func() error {
			return runner.runSBOM(ctx, directory, module)
		}); err != nil {
			return err
		}
	}
	if operation, exists := runner.operation(module.Directory, "fuzz"); exists {
		if err := announce(output, module.Directory, "fuzz", func() error {
			return runner.runOperation(ctx, directory, module, operation)
		}); err != nil {
			return err
		}
	}
	if module.Gates["documentation"] {
		if err := announce(output, module.Directory, "docs", func() error {
			return runner.checkDocumentation(ctx, directory, module)
		}); err != nil {
			return err
		}
	}
	if module.Gates["api_compatibility"] {
		if operation, exists := runner.operation(module.Directory, "api"); exists {
			if err := announce(output, module.Directory, "api", func() error {
				return runner.runOperation(ctx, directory, module, operation)
			}); err != nil {
				return err
			}
		} else if err := announce(output, module.Directory, "api", func() error {
			return runner.apiModule(ctx, output, module, false)
		}); err != nil {
			return err
		}
	}
	if module.Gates["lint"] {
		_, _ = fmt.Fprintf(output, "[%s] nilaway\n", module.Directory)
		err := runner.Executor.Run(ctx, Command{
			Name: "go", Dir: directory, Env: map[string]string{"GOWORK": "off"},
			Args: []string{"run", "go.uber.org/nilaway/cmd/nilaway@" + nilAwayVersion,
				"-include-pkgs=" + module.ModulePath, "./..."},
		})
		if err != nil {
			_, _ = fmt.Fprintf(output, "[%s] NilAway advisory: %v\n", module.Directory, err)
		}
	}
	for _, gate := range []string{"conformance", "interoperability", "benchmark"} {
		operation, exists := runner.operation(module.Directory, gate)
		if !exists {
			continue
		}
		if err := announce(output, module.Directory, gate, func() error {
			return runner.runOperation(ctx, directory, module, operation)
		}); err != nil {
			return err
		}
	}
	return nil
}

func (runner Runner) goTool(ctx context.Context, output io.Writer, module, gate, directory, tool string, args ...string) error {
	arguments := append([]string{"run", tool}, args...)
	return runner.command(ctx, output, module, gate, directory, arguments...)
}

func (runner Runner) runCoverage(ctx context.Context, output io.Writer, directory string, module inventory.Module) error {
	files := runner.coverageFiles
	if files == nil {
		files = operatingCoverageFiles{}
	}
	temporaryDirectory := ""
	if workspace, ok := runner.Executor.(taskWorkspace); ok {
		temporaryDirectory = workspace.TemporaryDirectory()
	}
	profile, err := files.CreateTemp(temporaryDirectory)
	if err != nil {
		return fmt.Errorf("create coverage profile: %w", err)
	}
	profilePath := profile.Name()
	if err := profile.Close(); err != nil {
		_ = files.Remove(profilePath)
		return fmt.Errorf("close coverage profile: %w", err)
	}
	defer func() { _ = files.Remove(profilePath) }()
	args := []string{"test"}
	if len(module.TestTags) > 0 {
		args = append(args, "-tags="+strings.Join(module.TestTags, ","))
	}
	args = append(args, "./...", "-count=1", "-timeout=20m", "-covermode=atomic", "-coverpkg=./...", "-coverprofile="+profilePath)
	if err := runner.Executor.Run(ctx, Command{Name: "go", Args: args, Dir: directory, Env: map[string]string{"GOWORK": "off"}}); err != nil {
		return err
	}
	opened, err := files.Open(profilePath)
	if err != nil {
		return fmt.Errorf("open coverage profile: %w", err)
	}
	defer opened.Close()
	expected := make([]string, 0, len(module.Packages))
	for _, packagePolicy := range module.Packages {
		if packagePolicy.CoverageRequired {
			expected = append(expected, packagePolicy.ImportPath)
		}
	}
	report, err := coverage.Verify(opened, expected)
	if err != nil {
		return err
	}
	_, _ = io.WriteString(output, report)
	_, _ = io.WriteString(output, "all production packages have exact 100% statement coverage\n")
	return nil
}

func (runner Runner) operation(module, gate string) (config.Operation, bool) {
	for _, operation := range runner.Policy.Operations {
		if operation.Module == module && operation.Gate == gate {
			return operation, true
		}
	}
	return config.Operation{}, false
}

func (runner Runner) runOperation(ctx context.Context, directory string, module inventory.Module, operation config.Operation) error {
	for index, step := range operation.Steps {
		timeout, err := time.ParseDuration(step.Timeout)
		if err != nil {
			return fmt.Errorf("%s %s step %d timeout: %w", module.Directory, operation.Gate, index, err)
		}
		stepContext, cancel := context.WithTimeout(ctx, timeout)
		command, err := operationCommand(directory, module, step)
		if err == nil {
			err = runner.Executor.Run(stepContext, command)
		}
		cancel()
		if err != nil {
			return fmt.Errorf("%s %s step %d: %w", module.Directory, operation.Gate, index, err)
		}
	}
	return nil
}

func operationCommand(directory string, module inventory.Module, step config.Step) (Command, error) {
	switch step.Type {
	case "go-test":
		args := []string{"test"}
		if len(module.TestTags) > 0 {
			args = append(args, "-tags="+strings.Join(module.TestTags, ","))
		}
		args = append(args, step.Packages...)
		args = append(args, fmt.Sprintf("-count=%d", step.Count), "-timeout="+step.Timeout)
		if step.Run != "" {
			args = append(args, "-run="+step.Run)
		}
		if step.Benchmark != "" {
			args = append(args, "-run=^$", "-bench="+step.Benchmark, "-benchmem", "-benchtime="+step.Budget)
		}
		if step.Fuzz != "" {
			args = append(args, "-run=^$", "-fuzz="+step.Fuzz, "-fuzztime="+step.Budget)
		}
		return Command{Name: "go", Args: args, Dir: directory, Env: map[string]string{"GOWORK": "off"}}, nil
	case "make":
		makefile, err := repositoryfile.Read(directory, step.Makefile, maximumMakefileSize)
		if err != nil {
			return Command{}, fmt.Errorf("read makefile: %w", err)
		}
		return Command{
			Name:  "make",
			Args:  []string{"--no-print-directory", "-f", "-", step.Target},
			Dir:   directory,
			Env:   map[string]string{"GOWORK": "off"},
			Stdin: bytes.NewReader(makefile),
		}, nil
	default:
		return Command{}, fmt.Errorf("unsupported operation type: %s", step.Type)
	}
}

func testArguments(tags []string, race bool) []string {
	args := []string{"test"}
	if race {
		args = append(args, "-race")
	}
	if len(tags) > 0 {
		args = append(args, "-tags="+strings.Join(tags, ","))
	}
	return append(args, "./...", "-count=1", "-timeout=20m")
}

func (runner Runner) command(ctx context.Context, output io.Writer, module, gate, directory string, args ...string) error {
	return announce(output, module, gate, func() error {
		if err := runner.Executor.Run(ctx, Command{Name: "go", Args: args, Dir: directory, Env: map[string]string{"GOWORK": "off"}}); err != nil {
			return fmt.Errorf("%s %s: %w", module, gate, err)
		}
		return nil
	})
}

func announce(output io.Writer, module, gate string, operation func() error) error {
	_, _ = fmt.Fprintf(output, "[%s] %s\n", module, gate)
	return operation()
}

func checkFormatting(root string) error {
	return walkModuleFiles(root, func(path string, _ fs.DirEntry) error {
		source, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		formatted, err := format.Source(source)
		if err != nil {
			return fmt.Errorf("format %s: %w", path, err)
		}
		if !bytes.Equal(source, formatted) {
			return fmt.Errorf("unformatted Go file: %s", path)
		}
		return nil
	})
}

func checkSafety(root string) error {
	return walkModuleFiles(root, func(path string, _ fs.DirEntry) error {
		if strings.HasSuffix(path, "_test.go") {
			return nil
		}
		parsed, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly|parser.ParseComments)
		if err != nil {
			return fmt.Errorf("parse %s: %w", path, err)
		}
		for _, imported := range parsed.Imports {
			if imported.Path.Value == `"unsafe"` || imported.Path.Value == `"C"` {
				return fmt.Errorf("forbidden production import %s in %s", imported.Path.Value, path)
			}
		}
		for _, group := range parsed.Comments {
			for _, comment := range group.List {
				if strings.HasPrefix(comment.Text, "//go:linkname") {
					return fmt.Errorf("forbidden go:linkname directive in %s", path)
				}
			}
		}
		return nil
	})
}

func walkModuleFiles(root string, visit func(string, fs.DirEntry) error) error {
	return filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if path != root {
				if entry.Name() == ".git" || entry.Name() == "vendor" {
					return filepath.SkipDir
				}
				if _, statErr := os.Stat(filepath.Join(path, "go.mod")); statErr == nil {
					return filepath.SkipDir
				} else if !os.IsNotExist(statErr) {
					return statErr
				}
			}
			return nil
		}
		if filepath.Ext(path) != ".go" {
			return nil
		}
		return visit(path, entry)
	})
}
