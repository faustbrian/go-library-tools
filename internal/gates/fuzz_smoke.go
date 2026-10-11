package gates

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"slices"
	"strings"
	"time"

	"github.com/faustbrian/go-library-tools/v2/internal/config"
	"github.com/faustbrian/go-library-tools/v2/internal/inventory"
)

const (
	fuzzSmokePreparationBudget   = 5 * time.Minute
	fuzzSmokeTotalBudget         = 3 * time.Minute
	fuzzSmokeStepBudget          = 2 * time.Minute
	fuzzSmokeMaximumCandidates   = 128
	fuzzSmokeMaximumTargets      = 32
	fuzzSmokeMaximumPackageBytes = 16 << 20
	fuzzSmokeMaximumPackages     = 10_000
)

type fuzzSmokeTarget struct {
	module           inventory.Module
	step             config.Step
	packageDirectory string
	compilerBin      string
}

type preparedFuzzSmoke struct {
	binary, directory, cache, compilerBin string
	names                                 string
	listed                                bool
}

type fuzzPackage struct {
	ImportPath         string   `json:"ImportPath"`
	Dir                string   `json:"Dir"`
	Imports            []string `json:"Imports"`
	TestImports        []string `json:"TestImports"`
	XTestImports       []string `json:"XTestImports"`
	CgoFiles           []string `json:"CgoFiles"`
	CFiles             []string `json:"CFiles"`
	CXXFiles           []string `json:"CXXFiles"`
	MFiles             []string `json:"MFiles"`
	HFiles             []string `json:"HFiles"`
	FFiles             []string `json:"FFiles"`
	SFiles             []string `json:"SFiles"`
	SwigFiles          []string `json:"SwigFiles"`
	SwigCXXFiles       []string `json:"SwigCXXFiles"`
	SysoFiles          []string `json:"SysoFiles"`
	EmbedFiles         []string `json:"EmbedFiles"`
	TestEmbedFiles     []string `json:"TestEmbedFiles"`
	XTestEmbedFiles    []string `json:"XTestEmbedFiles"`
	EmbedPatterns      []string `json:"EmbedPatterns"`
	TestEmbedPatterns  []string `json:"TestEmbedPatterns"`
	XTestEmbedPatterns []string `json:"XTestEmbedPatterns"`
	Module             *struct {
		Path string `json:"Path"`
	} `json:"Module"`
	Error      json.RawMessage   `json:"Error"`
	DepsErrors []json.RawMessage `json:"DepsErrors"`
}

// LocalChanges scans changed secrets before ordinary local verification and
// adds bounded smoke for configured fuzz targets affected by committed
// base-to-HEAD inputs. All phases share this Runner's executor; dirty-work
// certification is not implied.
func (runner Runner) LocalChanges(ctx context.Context, selection []string, base string) (result error) {
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
	preparationStarted := time.Now()
	preparationComplete := false
	defer func() {
		if !preparationComplete {
			_, _ = fmt.Fprintf(output, "fuzz-smoke preparation: %s (5m ceiling; execution not started)\n", time.Since(preparationStarted).Round(time.Millisecond))
		}
	}()
	smokeContext, cancel := context.WithTimeout(ctx, fuzzSmokePreparationBudget)
	defer cancel()
	policyChanges, err := runner.changedFuzzPolicy(smokeContext, changes)
	if err != nil {
		return err
	}
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
			affected = affected || policyChanges[fuzzPolicyIdentity(module.Directory, step)]
			if affected && !selectedIdentities[identity] {
				selectedIdentities[identity] = true
				compilerBin := ""
				if nativeRuntime, exists := graph["runtime"]; exists {
					compilerBin = filepath.Join(filepath.Dir(filepath.Dir(nativeRuntime.Dir)), "bin")
				}
				targets = append(targets, fuzzSmokeTarget{module: module, step: step, packageDirectory: graph[importPath].Dir, compilerBin: compilerBin})
				if len(targets) > fuzzSmokeMaximumTargets {
					return errors.New("fuzz smoke target limit exceeded; review the affected target scope")
				}
			}
		}
	}
	if len(targets) == 0 {
		return smokeContext.Err()
	}
	workspace, ok := runner.Executor.(taskWorkspace)
	if !ok || !filepath.IsAbs(workspace.TemporaryDirectory()) {
		return errors.New("fuzz smoke requires an owned temporary workspace")
	}
	scratch, err := os.MkdirTemp(workspace.TemporaryDirectory(), "fuzz-smoke-")
	if err != nil {
		return errors.New("fuzz smoke temporary workspace cannot be created")
	}
	defer func() {
		if err := os.RemoveAll(scratch); err != nil {
			result = errors.Join(result, errors.New("fuzz smoke temporary workspace cleanup failed"))
		}
	}()
	prepared := make(map[string]*preparedFuzzSmoke)
	for _, target := range targets {
		identity := fuzzSmokeContextIdentity(target.module, target.step)
		if prepared[identity] != nil {
			continue
		}
		binary, err := runner.prepareFuzzSmoke(smokeContext, target, scratch)
		if err != nil {
			return err
		}
		prepared[identity] = binary
	}
	if err := smokeContext.Err(); err != nil {
		return err
	}
	cancel()
	preparationComplete = true
	_, _ = fmt.Fprintf(output, "fuzz-smoke preparation: %d instrumented contexts for %d selected targets in %s (5m ceiling)\n", len(prepared), len(targets), time.Since(preparationStarted).Round(time.Millisecond))
	smokeContext, cancel = context.WithTimeout(ctx, fuzzSmokeTotalBudget)
	defer cancel()
	executionStarted := time.Now()
	defer func() {
		_, _ = fmt.Fprintf(output, "fuzz-smoke execution and cleanup: %s (3m execution ceiling)\n", time.Since(executionStarted).Round(time.Millisecond))
	}()
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
				if err := scoped.fuzzSmokeStep(smokeContext, prepared[fuzzSmokeContextIdentity(module, step)], step); err != nil {
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

// changedFuzzPolicy compares only declarations that affect the smoke binary.
// Full-campaign budgets and non-fuzz gates remain validated by ordinary local
// checks but do not invalidate unchanged application fuzz targets.
func (runner Runner) changedFuzzPolicy(ctx context.Context, changes []string) (map[string]bool, error) {
	policyChanged := slices.Contains(changes, ".golib.yaml")
	catalogChanged := slices.Contains(changes, runner.Policy.Manifests.Modules) || slices.Contains(changes, runner.Policy.Manifests.Packages)
	if !policyChanged && !catalogChanged {
		return nil, nil
	}
	if runner.secretChanges == nil {
		return nil, errors.New("fuzz smoke policy requires an immutable Git range")
	}
	read := func(commit string) (config.Config, error) {
		data, err := runner.changeGit(ctx, "show", commit+":.golib.yaml")
		if err != nil {
			return config.Config{}, err
		}
		policy, err := config.Decode([]byte(data))
		if err != nil {
			return config.Config{}, errors.New("fuzz smoke historical policy is invalid or exceeds its size budget")
		}
		return policy, nil
	}
	before, after := runner.Policy, runner.Policy
	if policyChanged {
		var err error
		before, err = read(runner.secretChanges.base)
		if err != nil {
			return nil, err
		}
		after, err = read(runner.secretChanges.head)
		if err != nil {
			return nil, err
		}
	}
	if !reflect.DeepEqual(after, runner.Policy) {
		return nil, errors.New("fuzz smoke policy differs from immutable HEAD")
	}
	materialModules := make(map[string]bool)
	if catalogChanged || before.Manifests != after.Manifests {
		readCatalog := func(commit string, policy config.Config) (inventory.Inventory, error) {
			moduleData, err := runner.changeGit(ctx, "show", commit+":"+policy.Manifests.Modules)
			if err != nil {
				return inventory.Inventory{}, err
			}
			packageData, err := runner.changeGit(ctx, "show", commit+":"+policy.Manifests.Packages)
			if err != nil {
				return inventory.Inventory{}, err
			}
			catalog, err := inventory.Decode(policy, []byte(moduleData), []byte(packageData))
			if err != nil {
				return inventory.Inventory{}, errors.New("fuzz smoke historical inventory is invalid or exceeds its size budget")
			}
			return catalog, nil
		}
		oldCatalog, err := readCatalog(runner.secretChanges.base, before)
		if err != nil {
			return nil, err
		}
		newCatalog, err := readCatalog(runner.secretChanges.head, after)
		if err != nil {
			return nil, err
		}
		if !reflect.DeepEqual(newCatalog, runner.Catalog) {
			return nil, errors.New("fuzz smoke inventory differs from immutable HEAD")
		}
		oldModules := make(map[string]inventory.Module)
		for _, module := range oldCatalog.Modules {
			oldModules[module.Directory] = module
		}
		for _, module := range newCatalog.Modules {
			old, exists := oldModules[module.Directory]
			materialModules[module.Directory] = !exists || oldCatalog.GoVersion != newCatalog.GoVersion || !sameFuzzModuleInputs(old, module)
		}
	}
	previous := make(map[string]config.Step)
	for _, operation := range before.Operations {
		if operation.Gate != "fuzz" {
			continue
		}
		for _, step := range operation.Steps {
			if step.Type == "go-test" && step.Fuzz != "" && len(step.Packages) == 1 {
				previous[fuzzPolicyIdentity(operation.Module, step)] = normalizedFuzzPolicyStep(step)
			}
		}
	}
	selected := make(map[string]bool)
	for _, operation := range after.Operations {
		if operation.Gate != "fuzz" {
			continue
		}
		for _, step := range operation.Steps {
			if step.Type != "go-test" || step.Fuzz == "" || len(step.Packages) != 1 {
				continue
			}
			identity := fuzzPolicyIdentity(operation.Module, step)
			old, exists := previous[identity]
			selected[identity] = materialModules[operation.Module] || before.Runtimes != after.Runtimes || !exists || !reflect.DeepEqual(old, normalizedFuzzPolicyStep(step))
		}
	}
	return selected, ctx.Err()
}

// Versions, descriptions and unrelated gate policy do not change a fuzz binary.
// Compare only runtime, service and ownership declarations; package ordering is
// immaterial and package documentation/coverage classification is not execution.
func sameFuzzModuleInputs(before, after inventory.Module) bool {
	packageInputs := func(module inventory.Module) map[string]inventory.Package {
		result := make(map[string]inventory.Package)
		for _, pkg := range module.Packages {
			result[pkg.Directory] = inventory.Package{ModuleDirectory: pkg.ModuleDirectory, Directory: pkg.Directory, ImportPath: pkg.ImportPath, BuildTags: pkg.BuildTags}
		}
		return result
	}
	return before.ModulePath == after.ModulePath && before.GoVersion == after.GoVersion && before.Gates["fuzz"] == after.Gates["fuzz"] &&
		slices.Equal(before.TestTags, after.TestTags) && slices.Equal(before.BuildTags, after.BuildTags) &&
		slices.Equal(before.RequiredServices, after.RequiredServices) && slices.Equal(before.OwnedDependencies, after.OwnedDependencies) &&
		reflect.DeepEqual(packageInputs(before), packageInputs(after))
}

func fuzzPolicyIdentity(module string, step config.Step) string {
	return module + "\x00" + path.Clean(step.Packages[0]) + "\x00" + step.Fuzz
}

func normalizedFuzzPolicyStep(step config.Step) config.Step {
	step.Packages = []string{path.Clean(step.Packages[0])}
	step.Budget, step.Timeout, step.Count = "2s", "15s", 1
	return step
}

func (runner Runner) fuzzPackageGraph(ctx context.Context, module inventory.Module, packages []string) (map[string]fuzzPackage, error) {
	args := []string{"list", "-deps", "-test", "-json"}
	if len(module.TestTags) > 0 {
		args = append(args, "-tags="+strings.Join(module.TestTags, ","))
	}
	args = append(args, packages...)
	data, err := runner.fuzzCommand(ctx, Command{cacheClass: nativeRepositoryCache, Name: "go", Args: args, Dir: filepath.Join(runner.Root, module.Directory), Env: map[string]string{"GOWORK": "off"}}, fuzzSmokeMaximumPackageBytes)
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
		if change == ".go-version" ||
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
		for _, dependency := range pkg.Imports {
			// Cgo source imports the pseudo-package C, which go list does not
			// emit as a Go dependency. Do not relax real or test import checks.
			if dependency == "C" && len(pkg.CgoFiles) > 0 {
				continue
			}
			queue = append(queue, dependency)
		}
		// Only the selected package's tests are built. Dependencies' own tests
		// appear in go list metadata but are not part of this fuzz binary.
		if current == target {
			if test, exists := graph[target+".test"]; exists {
				// The test binary's resolved imports retain Go's exact rebuilt
				// dependency identities. Canonical test metadata can name a
				// dependency emitted only as a test-specific variant.
				queue = append(queue, test.Imports...)
			} else {
				queue = append(queue, pkg.TestImports...)
				queue = append(queue, pkg.XTestImports...)
			}
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

func fuzzSmokeContextIdentity(module inventory.Module, step config.Step) string {
	return module.Directory + "\x00" + path.Clean(step.Packages[0]) + "\x00" + strings.Join(module.TestTags, ",") + fmt.Sprintf("\x00%t", step.Race)
}

// Compile once per native package context, without running TestMain, seeds or
// fuzz functions. The selected graph is also the authority for the runtime cwd
// and compiler PATH that cmd/go normally supplies to its test subprocess.
func (runner Runner) prepareFuzzSmoke(ctx context.Context, target fuzzSmokeTarget, scratch string) (*preparedFuzzSmoke, error) {
	ctx, cancel := context.WithTimeout(ctx, fuzzSmokeStepBudget)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	directory, err := os.MkdirTemp(scratch, "context-")
	if err != nil {
		return nil, errors.New("fuzz smoke compilation workspace cannot be created")
	}
	prepared := &preparedFuzzSmoke{binary: filepath.Join(directory, "smoke.test.exe"), directory: target.packageDirectory, cache: filepath.Join(directory, "corpus"), compilerBin: target.compilerBin}
	if err := os.Mkdir(prepared.cache, 0o700); err != nil {
		return nil, errors.New("fuzz smoke corpus workspace cannot be created")
	}
	args := []string{"test", "-c", "-fuzz=^Fuzz", "-o", prepared.binary, "-p=1"}
	if target.step.Race {
		args = append(args, "-race")
	}
	if len(target.module.TestTags) > 0 {
		args = append(args, "-tags="+strings.Join(target.module.TestTags, ","))
	}
	args = append(args, target.step.Packages[0])
	_, err = runner.fuzzCommand(ctx, Command{cacheClass: nativeRepositoryCache, Name: "go", Args: args, Dir: filepath.Join(runner.Root, target.module.Directory), Env: map[string]string{"GOWORK": "off"}}, maximumSecurityProcessOutput)
	if err != nil {
		return nil, err
	}
	info, err := os.Lstat(prepared.binary)
	if err != nil || !info.Mode().IsRegular() || (runtime.GOOS != "windows" && info.Mode()&0o111 == 0) || !filepath.IsAbs(prepared.directory) {
		return nil, errors.New("fuzz smoke compiled binary or package directory is invalid")
	}
	return prepared, nil
}

func (prepared *preparedFuzzSmoke) command(args ...string) Command {
	env := map[string]string{"GOWORK": "off", "PWD": prepared.directory}
	if prepared.compilerBin != "" {
		env["PATH"] = prepared.compilerBin + string(os.PathListSeparator) + os.Getenv("PATH")
	}
	return Command{cacheClass: nativeRepositoryCache, Name: prepared.binary, Args: args, Dir: prepared.directory, Env: env}
}

func (runner Runner) fuzzSmokeStep(ctx context.Context, prepared *preparedFuzzSmoke, step config.Step) error {
	ctx, cancel := context.WithTimeout(ctx, fuzzSmokeStepBudget)
	defer cancel()
	if !prepared.listed {
		names, err := runner.fuzzCommand(ctx, prepared.command("-test.run=^$", "-test.list=^Fuzz", "-test.count=1", "-test.timeout=15s", "-test.paniconexit0"), maximumSecurityProcessOutput)
		if err != nil {
			return err
		}
		prepared.names, prepared.listed = names, true
	}
	selector, err := regexp.Compile(step.Fuzz)
	if err != nil {
		return errors.New("fuzz smoke target selector is invalid")
	}
	matches := 0
	for line := range strings.SplitSeq(prepared.names, "\n") {
		if strings.HasPrefix(line, "Fuzz") && selector.MatchString(line) {
			matches++
		}
	}
	if matches != 1 {
		return errors.New("fuzz smoke target must resolve to exactly one test")
	}
	_, err = runner.fuzzCommand(ctx, prepared.command("-test.run=^$", "-test.fuzz="+step.Fuzz, "-test.fuzztime=2s", "-test.timeout=15s", "-test.count=1", "-test.parallel=2", "-test.fuzzcachedir="+prepared.cache, "-test.paniconexit0"), maximumSecurityProcessOutput)
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
