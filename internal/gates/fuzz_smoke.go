package gates

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/faustbrian/go-library-tools/v2/internal/config"
	"github.com/faustbrian/go-library-tools/v2/internal/inventory"
)

const (
	fuzzSmokeTotalBudget = 3 * time.Minute
	// Go's first active fuzz run compiles instrumented dependencies even when
	// ordinary tests are cached. Admission includes that cold build, not just
	// the two-second campaign; the total smoke deadline still takes precedence.
	fuzzSmokeStepBudget          = time.Minute
	fuzzSmokeMaximumCandidates   = 128
	fuzzSmokeMaximumTargets      = 32
	fuzzSmokeMaximumPackageBytes = 16 << 20
	fuzzSmokeMaximumPackages     = 10_000
)

type fuzzSmokeTarget struct {
	module inventory.Module
	step   config.Step
}

type fuzzPackage struct {
	ImportPath         string
	Dir                string
	Imports            []string
	TestImports        []string
	XTestImports       []string
	CFiles             []string
	CXXFiles           []string
	MFiles             []string
	HFiles             []string
	FFiles             []string
	SFiles             []string
	SwigFiles          []string
	SwigCXXFiles       []string
	SysoFiles          []string
	EmbedFiles         []string
	TestEmbedFiles     []string
	XTestEmbedFiles    []string
	EmbedPatterns      []string
	TestEmbedPatterns  []string
	XTestEmbedPatterns []string
	Module             *struct{ Path string }
	Error              json.RawMessage
	DepsErrors         []json.RawMessage
}

// LocalChanges scans changed secrets before ordinary local verification and
// adds bounded smoke for configured fuzz targets affected by committed
// base-to-HEAD inputs. All phases share this Runner's executor; dirty-work
// certification is not implied.
func (runner Runner) LocalChanges(ctx context.Context, selection []string, base string) error {
	secretContext, cancel := context.WithTimeout(ctx, changedSecretBudget)
	defer cancel()
	changes, err := runner.committedChanges(secretContext, base)
	if err != nil {
		return err
	}
	modules, err := runner.selectModules(selection)
	if err != nil {
		return err
	}
	if err := runner.prepareSecretChanges(secretContext, base, changes); err != nil {
		return err
	}
	if err := runner.Secrets(secretContext); err != nil {
		return err
	}
	cancel()
	runner.repositorySecretsScanned = true
	if err := runner.Local(ctx, selection); err != nil {
		return err
	}
	if len(changes) == 0 {
		return nil
	}
	output := runner.Output
	if output == nil {
		output = io.Discard
	}
	smokeContext, cancel := context.WithTimeout(ctx, fuzzSmokeTotalBudget)
	defer cancel()
	var targets []fuzzSmokeTarget
	selectedIdentities := make(map[string]bool)
	candidates := 0
	for _, module := range modules {
		operation, exists := runner.operation(module.Directory, "fuzz")
		if !exists {
			continue
		}
		var steps []config.Step
		var packages []string
		for _, step := range operation.Steps {
			if step.Type != "go-test" || step.Fuzz == "" || len(step.Packages) != 1 || strings.HasSuffix(step.Packages[0], "/...") {
				return errors.New("fuzz smoke requires configured single-package Go fuzz targets")
			}
			candidates++
			if candidates > fuzzSmokeMaximumCandidates {
				return errors.New("fuzz smoke candidate limit exceeded")
			}
			steps = append(steps, step)
			if !slices.Contains(packages, step.Packages[0]) {
				packages = append(packages, step.Packages[0])
			}
		}
		graph, err := runner.fuzzPackageGraph(smokeContext, module, packages)
		if err != nil {
			return err
		}
		for _, step := range steps {
			importPath := module.ModulePath
			if suffix := strings.TrimPrefix(step.Packages[0], "./"); suffix != "." {
				importPath += "/" + suffix
			}
			affected, err := runner.fuzzTargetAffected(smokeContext, module, importPath, graph, changes)
			if err != nil {
				return err
			}
			identity := module.Directory + "\x00" + step.Packages[0] + "\x00" + step.Fuzz
			if affected && !selectedIdentities[identity] {
				selectedIdentities[identity] = true
				targets = append(targets, fuzzSmokeTarget{module, step})
				if len(targets) > fuzzSmokeMaximumTargets {
					return errors.New("fuzz smoke target limit exceeded; review the affected target scope")
				}
			}
		}
	}
	for _, module := range modules {
		var selected []config.Step
		for _, target := range targets {
			if target.module.Directory == module.Directory {
				selected = append(selected, target.step)
			}
		}
		if len(selected) == 0 {
			continue
		}
		if err := runner.withModuleServices(smokeContext, module, func(scoped Runner) error {
			for _, step := range selected {
				if err := scoped.fuzzSmokeStep(smokeContext, module, step); err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			return err
		}
		_, _ = fmt.Fprintf(output, "[%s] fuzz-smoke: %d affected configured targets passed\n", module.Directory, len(selected))
	}
	return smokeContext.Err()
}

func (runner Runner) fuzzPackageGraph(ctx context.Context, module inventory.Module, packages []string) (map[string]fuzzPackage, error) {
	args := []string{"list", "-deps", "-test", "-json"}
	if len(module.TestTags) > 0 {
		args = append(args, "-tags="+strings.Join(module.TestTags, ","))
	}
	args = append(args, packages...)
	data, err := runner.fuzzCommand(ctx, Command{Name: "go", Args: args, Dir: filepath.Join(runner.Root, module.Directory), Env: map[string]string{"GOWORK": "off"}}, fuzzSmokeMaximumPackageBytes)
	if err != nil {
		return nil, err
	}
	graph := make(map[string]fuzzPackage)
	decoder := json.NewDecoder(strings.NewReader(data))
	for count := 0; ; count++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		var pkg fuzzPackage
		err := decoder.Decode(&pkg)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil || count >= fuzzSmokeMaximumPackages || pkg.ImportPath == "" || len(pkg.ImportPath) > 4096 ||
			!filepath.IsAbs(pkg.Dir) || filepath.Clean(pkg.Dir) != pkg.Dir ||
			(len(pkg.Error) > 0 && string(pkg.Error) != "null") || len(pkg.DepsErrors) != 0 {
			return nil, errors.New("fuzz smoke package discovery is invalid or incomplete")
		}
		if _, duplicate := graph[pkg.ImportPath]; duplicate {
			return nil, errors.New("fuzz smoke package discovery contains duplicate identities")
		}
		graph[pkg.ImportPath] = pkg
	}
	if len(graph) == 0 {
		return nil, errors.New("fuzz smoke package discovery is empty")
	}
	return graph, nil
}

func (runner Runner) fuzzTargetAffected(ctx context.Context, module inventory.Module, target string, graph map[string]fuzzPackage, changes []string) (bool, error) {
	// Go reports canonical directories, including macOS /var -> /private/var.
	canonicalRoot, err := filepath.EvalSymlinks(runner.Root)
	if err != nil {
		return false, errors.New("fuzz smoke repository root cannot be resolved")
	}
	affected := false
	for _, change := range changes {
		if change == ".golib.yaml" || change == ".go-version" ||
			change == path.Join(module.Directory, "go.mod") || change == path.Join(module.Directory, "go.sum") {
			affected = true
		}
	}
	seen := make(map[string]bool)
	queue := []string{target}
	for len(queue) > 0 {
		if err := ctx.Err(); err != nil {
			return false, err
		}
		current := queue[0]
		queue = queue[1:]
		if seen[current] {
			continue
		}
		seen[current] = true
		pkg, exists := graph[current]
		if !exists {
			return false, errors.New("fuzz smoke package dependency is missing")
		}
		if pkg.Module != nil {
			for _, owner := range runner.Catalog.Modules {
				if owner.ModulePath != pkg.Module.Path {
					continue
				}
				for _, change := range changes {
					if change == path.Join(owner.Directory, "go.mod") || change == path.Join(owner.Directory, "go.sum") {
						affected = true
					}
				}
			}
		}
		relative, err := filepath.Rel(runner.Root, pkg.Dir)
		if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			relative, err = filepath.Rel(canonicalRoot, pkg.Dir)
		}
		if err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			directory := filepath.ToSlash(relative)
			embedFiles, embedPatterns := pkg.EmbedFiles, pkg.EmbedPatterns
			if current == target {
				embedFiles = append(append(slices.Clone(embedFiles), pkg.TestEmbedFiles...), pkg.XTestEmbedFiles...)
				embedPatterns = append(append(slices.Clone(embedPatterns), pkg.TestEmbedPatterns...), pkg.XTestEmbedPatterns...)
			}
			nativeFiles := make(map[string]bool)
			for _, files := range [][]string{pkg.CFiles, pkg.CXXFiles, pkg.MFiles, pkg.HFiles, pkg.FFiles, pkg.SFiles, pkg.SwigFiles, pkg.SwigCXXFiles, pkg.SysoFiles} {
				for _, file := range files {
					nativeFiles[file] = true
				}
			}
			for _, change := range changes {
				if err := ctx.Err(); err != nil {
					return false, err
				}
				if (path.Dir(change) == directory && (fuzzSourceInput(change, current == target) || nativeFiles[path.Base(change)])) ||
					strings.HasPrefix(change, path.Join(directory, "testdata")+"/") {
					affected = true
				}
				for _, embedded := range embedFiles {
					if change == path.Join(directory, embedded) {
						affected = true
					}
				}
				for _, pattern := range embedPatterns {
					// A removed embedded input is absent from current EmbedFiles.
					// Match patterns and their directory descendants conservatively.
					pattern = strings.TrimPrefix(pattern, "all:")
					if _, err := path.Match(pattern, ""); err != nil {
						return false, errors.New("fuzz smoke embed pattern is invalid")
					}
					candidate := change
					if directory != "." {
						if !strings.HasPrefix(change, directory+"/") {
							continue
						}
						candidate = strings.TrimPrefix(change, directory+"/")
					}
					if candidate == "." {
						continue
					}
					for ; candidate != "."; candidate = path.Dir(candidate) {
						matches, _ := path.Match(pattern, candidate)
						if matches {
							affected = true
							break
						}
					}
				}
			}
		}
		queue = append(queue, pkg.Imports...)
		// Only the selected package's tests are built. Dependencies' own tests
		// appear in go list metadata but are not part of this fuzz binary.
		if current == target {
			queue = append(queue, pkg.TestImports...)
			queue = append(queue, pkg.XTestImports...)
		}
	}
	return affected, nil
}

func fuzzSourceInput(name string, includeTests bool) bool {
	if strings.HasSuffix(name, "_test.go") {
		return includeTests
	}
	// Native source removed from HEAD is absent from go list's current file
	// lists. Conservatively retain supported compiler inputs for deletions too.
	switch path.Ext(name) {
	case ".go", ".c", ".cc", ".cpp", ".cxx", ".m", ".h", ".hh", ".hpp", ".hxx", ".f", ".F", ".for", ".f90", ".s", ".S", ".sx", ".swig", ".swigcxx", ".syso":
		return true
	default:
		return false
	}
}

func (runner Runner) fuzzSmokeStep(ctx context.Context, module inventory.Module, step config.Step) error {
	ctx, cancel := context.WithTimeout(ctx, fuzzSmokeStepBudget)
	defer cancel()
	step.Budget, step.Timeout, step.Count = "2s", "15s", 1
	command, err := operationCommand(filepath.Join(runner.Root, module.Directory), module, step)
	if err != nil {
		return errors.New("fuzz smoke command is invalid")
	}
	list := command
	list.Args = []string{"test"}
	if len(module.TestTags) > 0 {
		list.Args = append(list.Args, "-tags="+strings.Join(module.TestTags, ","))
	}
	list.Args = append(list.Args, step.Packages[0], "-run=^$", "-list="+step.Fuzz, "-count=1", "-timeout=15s", "-p=1")
	names, err := runner.fuzzCommand(ctx, list, maximumSecurityProcessOutput)
	if err != nil {
		return err
	}
	selector, err := regexp.Compile(step.Fuzz)
	if err != nil {
		return errors.New("fuzz smoke target selector is invalid")
	}
	matches := 0
	for line := range strings.SplitSeq(names, "\n") {
		if strings.HasPrefix(line, "Fuzz") && selector.MatchString(line) {
			matches++
		}
	}
	if matches != 1 {
		return errors.New("fuzz smoke target must resolve to exactly one test")
	}
	command.Args = append(command.Args, "-parallel=2", "-p=1")
	_, err = runner.fuzzCommand(ctx, command, maximumSecurityProcessOutput)
	return err
}

func (runner Runner) fuzzCommand(ctx context.Context, command Command, limit int) (string, error) {
	return runner.capturedTestCommand(ctx, command, limit, "fuzz smoke command failed or exceeded its output budget")
}

func (runner Runner) capturedTestCommand(ctx context.Context, command Command, limit int, failureClass string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	stdout := &sourceInventoryOutput{limit: limit}
	stderr := &boundedProcessOutput{limit: maximumSecurityProcessOutput}
	command.Stdout, command.Stderr, command.boundedScanner = stdout, stderr, true
	if err := runner.Executor.Run(ctx, command); err != nil || stdout.didOverflow() || stderr.didOverflow() {
		return "", errors.Join(sourceCommandError{class: failureClass, cause: err}, ctx.Err())
	}
	return stdout.String(), ctx.Err()
}
