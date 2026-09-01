// Package cli implements the stable golib command-line contract.
package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"slices"

	"github.com/faustbrian/go-library-tools/internal/buildinfo"
	"github.com/faustbrian/go-library-tools/internal/cohesion"
	"github.com/faustbrian/go-library-tools/internal/config"
	"github.com/faustbrian/go-library-tools/internal/consumers"
	"github.com/faustbrian/go-library-tools/internal/evidence"
	"github.com/faustbrian/go-library-tools/internal/gates"
	"github.com/faustbrian/go-library-tools/internal/inventory"
	"github.com/faustbrian/go-library-tools/internal/releasecheck"
	"github.com/faustbrian/go-library-tools/internal/repository"
	"github.com/faustbrian/go-library-tools/internal/specification"
	"github.com/faustbrian/go-library-tools/internal/upgrade"
)

const help = `golib validates and executes the Go library repository contract.

Usage:
  golib --version
  golib check [--all|--module <directory>]
  golib cohesion check [--json]
  golib cohesion catalog <consumer|engineering> [--json]
  golib cohesion aggregate <generate|check> --inputs <file> --output <directory>
  golib config validate
	golib config show --json
  golib inventory [--json]
	golib consumers validate [--json]
  golib repository check
  golib specification check [--online]
  golib workflows check
  golib coverage [--module <directory>]
  golib mutation [--module <directory>]
	golib mutation import --module <directory> --archive <path> --ledger <path>
  golib api check
  golib api update
  golib docs check [--module <directory>]
	golib services cycle [--module <directory>]
  golib release check
  golib release dry-run
  golib evidence inspect
  golib upgrade <plan|apply> --version <version> --workflow-sha <sha> --checksums-sha256 <digest> [--json]
`

// Execute runs one command and returns a stable process exit code.
func Execute(args []string, workingDirectory string, stdout, stderr io.Writer) int {
	return ExecuteContext(context.Background(), args, workingDirectory, stdout, stderr)
}

// ExecuteContext runs one command with caller-owned cancellation.
func ExecuteContext(ctx context.Context, args []string, workingDirectory string, stdout, stderr io.Writer) int {
	if len(args) == 1 && args[0] == "--version" {
		_, _ = fmt.Fprintln(stdout, buildinfo.Version)
		return 0
	}
	return executeContext(ctx, args, workingDirectory, stdout, stderr, gates.NewProcessExecutor)
}

type executorFactory func(string, io.Writer, io.Writer) (gates.Executor, func() error, error)

var renderCohesionCatalog = cohesion.RenderMarkdown

func execute(args []string, workingDirectory string, stdout, stderr io.Writer, createExecutor executorFactory) int {
	return executeContext(context.Background(), args, workingDirectory, stdout, stderr, createExecutor)
}

func executeContext(ctx context.Context, args []string, workingDirectory string, stdout, stderr io.Writer, createExecutor executorFactory) int {
	if len(args) == 1 && (args[0] == "--help" || args[0] == "-h") {
		_, _ = io.WriteString(stdout, help)
		return 0
	}
	if len(args) == 0 {
		return usage(stderr, "command is required")
	}
	if args[0] == "upgrade" {
		return executeUpgrade(args[1:], workingDirectory, stdout, stderr)
	}

	root, err := findRoot(workingDirectory)
	if err != nil {
		return failure(stderr, err)
	}
	policy, err := config.Load(root)
	if err != nil {
		return failure(stderr, err)
	}
	if err := buildinfo.ValidateRequired(policy.ToolVersion); err != nil {
		return failure(stderr, err)
	}
	if args[0] == "cohesion" {
		return executeCohesion(args[1:], root, policy, stdout, stderr)
	}
	catalog, err := inventory.Load(root, policy)
	if err != nil {
		return failure(stderr, err)
	}

	switch args[0] {
	case "check":
		selection, usageError := moduleSelection(args[1:], catalog.Modules)
		if usageError != nil {
			return usage(stderr, usageError.Error())
		}
		return withExecutor(root, stdout, stderr, createExecutor, func(executor gates.Executor) error {
			return (gates.Runner{Root: root, Catalog: catalog, Policy: policy, Executor: executor, Output: stdout}).Check(ctx, selection)
		})
	case "config":
		if len(args) == 2 && args[1] == "validate" {
			_, _ = fmt.Fprintf(stdout, "configuration valid: %s\n", catalog.Repository)
			return 0
		}
		if len(args) == 3 && args[1] == "show" && args[2] == "--json" {
			encoder := json.NewEncoder(stdout)
			encoder.SetIndent("", "  ")
			if err := encoder.Encode(policy); err != nil {
				return failure(stderr, fmt.Errorf("write configuration: %w", err))
			}
			return 0
		}
		return usage(stderr, "usage: golib config <validate|show --json>")
	case "repository":
		if len(args) != 2 || args[1] != "check" {
			return usage(stderr, "usage: golib repository check")
		}
		if err := repository.Check(root, catalog); err != nil {
			return failure(stderr, err)
		}
		_, _ = io.WriteString(stdout, "standalone repository contract passed\n")
		return 0
	case "specification":
		if len(args) < 2 || len(args) > 3 || args[1] != "check" || (len(args) == 3 && args[2] != "--online") {
			return usage(stderr, "usage: golib specification check [--online]")
		}
		var report specification.Report
		var checkError error
		if len(args) == 3 {
			report, checkError = specification.CheckOnline(ctx, root, catalog, &http.Client{})
		} else {
			report, checkError = specification.Check(root, catalog)
		}
		if checkError != nil {
			return failure(stderr, checkError)
		}
		_, _ = fmt.Fprintf(stdout, "specification decisions valid: %d module(s), %d decision(s)\n", report.Modules, report.Decisions)
		return 0
	case "workflows":
		if len(args) != 2 || args[1] != "check" {
			return usage(stderr, "usage: golib workflows check")
		}
		return withExecutor(root, stdout, stderr, createExecutor, func(executor gates.Executor) error {
			return (gates.Runner{Root: root, Executor: executor, Output: stdout}).Workflows(ctx)
		})
	case "inventory":
		if len(args) == 1 {
			_, _ = fmt.Fprintf(stdout, "%s: %d module(s)\n", catalog.Repository, len(catalog.Modules))
			return 0
		}
		if len(args) != 2 || args[1] != "--json" {
			return usage(stderr, "usage: golib inventory [--json]")
		}
		encoder := json.NewEncoder(stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(catalog); err != nil {
			return failure(stderr, fmt.Errorf("write inventory: %w", err))
		}
		return 0
	case "consumers":
		if len(args) < 2 || args[1] != "validate" || (len(args) == 3 && args[2] != "--json") || len(args) > 3 {
			return usage(stderr, "usage: golib consumers validate [--json]")
		}
		manifest, err := consumers.Load(root, "consumers.json")
		if err != nil {
			return failure(stderr, err)
		}
		summary := manifest.Summary()
		if len(args) == 3 {
			encoder := json.NewEncoder(stdout)
			encoder.SetIndent("", "  ")
			if err := encoder.Encode(summary); err != nil {
				return failure(stderr, fmt.Errorf("write consumer inventory: %w", err))
			}
			return 0
		}
		_, err = fmt.Fprintf(stdout, "consumer inventory valid: %d total, %d active, %d deferred, %d tooling\n", summary.Total, summary.Active, summary.Deferred, summary.Tooling)
		if err != nil {
			return failure(stderr, fmt.Errorf("write consumer inventory: %w", err))
		}
		return 0
	case "api":
		if len(args) < 2 || (args[1] != "check" && args[1] != "update") {
			return usage(stderr, "usage: golib api <check|update> [--module <directory>]")
		}
		selection, usageError := moduleSelection(args[2:], catalog.Modules)
		if usageError != nil {
			return usage(stderr, "usage: golib api <check|update> [--module <directory>]")
		}
		return withExecutor(root, stdout, stderr, createExecutor, func(executor gates.Executor) error {
			return (gates.Runner{Root: root, Catalog: catalog, Policy: policy, Executor: executor, Output: stdout}).API(ctx, selection, args[1] == "update")
		})
	case "coverage":
		selection, usageError := moduleSelection(args[1:], catalog.Modules)
		if usageError != nil {
			return usage(stderr, "usage: golib coverage [--module <directory>]")
		}
		return withExecutor(root, stdout, stderr, createExecutor, func(executor gates.Executor) error {
			return (gates.Runner{Root: root, Catalog: catalog, Policy: policy, Executor: executor, Output: stdout}).Coverage(ctx, selection)
		})
	case "mutation":
		if len(args) > 1 && args[1] == "import" {
			options, usageError := mutationImportArguments(args[2:])
			if usageError != nil {
				return usage(stderr, "usage: golib mutation import --module <directory> --archive <path> --ledger <path>")
			}
			return withExecutor(root, stdout, stderr, createExecutor, func(executor gates.Executor) error {
				return (gates.Runner{Root: root, Catalog: catalog, Policy: policy, Executor: executor, Output: stdout}).MutationImport(
					ctx, options.module, options.archive, options.ledger,
				)
			})
		}
		selection, usageError := moduleSelection(args[1:], catalog.Modules)
		if usageError != nil {
			return usage(stderr, "usage: golib mutation [--module <directory>]")
		}
		return withExecutor(root, stdout, stderr, createExecutor, func(executor gates.Executor) error {
			return (gates.Runner{Root: root, Catalog: catalog, Policy: policy, Executor: executor, Output: stdout}).Mutation(ctx, selection)
		})
	case "docs":
		if len(args) < 2 || args[1] != "check" {
			return usage(stderr, "usage: golib docs check [--module <directory>]")
		}
		selection, usageError := moduleSelection(args[2:], catalog.Modules)
		if usageError != nil {
			return usage(stderr, "usage: golib docs check [--module <directory>]")
		}
		return withExecutor(root, stdout, stderr, createExecutor, func(executor gates.Executor) error {
			return (gates.Runner{Root: root, Catalog: catalog, Policy: policy, Executor: executor, Output: stdout}).Docs(ctx, selection)
		})
	case "services":
		if len(args) < 2 || args[1] != "cycle" {
			return usage(stderr, "usage: golib services cycle [--module <directory>]")
		}
		selection, usageError := moduleSelection(args[2:], catalog.Modules)
		if usageError != nil {
			return usage(stderr, "usage: golib services cycle [--module <directory>]")
		}
		return withExecutor(root, stdout, stderr, createExecutor, func(executor gates.Executor) error {
			return (gates.Runner{Root: root, Catalog: catalog, Policy: policy, Executor: executor, Output: stdout}).ServiceCycle(ctx, selection)
		})
	case "release":
		if len(args) != 2 || (args[1] != "check" && args[1] != "dry-run") {
			return usage(stderr, "usage: golib release <check|dry-run>")
		}
		selection, validationError := releasecheck.Validate(catalog, policy)
		if validationError != nil {
			return failure(stderr, validationError)
		}
		if validationError := repository.Check(root, catalog); validationError != nil {
			return failure(stderr, validationError)
		}
		if args[1] == "check" {
			_, _ = fmt.Fprintf(stdout, "release contract passed for %d module(s)\n", len(selection))
			return 0
		}
		return withExecutor(root, stdout, stderr, createExecutor, func(executor gates.Executor) error {
			return (gates.Runner{Root: root, Catalog: catalog, Policy: policy, Executor: executor, Output: stdout}).ReleaseDryRun(ctx, selection)
		})
	case "evidence":
		if len(args) < 2 || args[1] != "inspect" || (len(args) == 3 && args[2] != "--json") || len(args) > 3 {
			return usage(stderr, "usage: golib evidence inspect [--json]")
		}
		modules := make([]string, 0, len(catalog.Modules))
		for _, module := range catalog.Modules {
			modules = append(modules, module.Directory)
		}
		records, inspectError := evidence.Inspect(root, policy.Evidence.Root, catalog.Repository, modules)
		if inspectError != nil {
			return failure(stderr, inspectError)
		}
		if len(args) == 3 {
			encoder := json.NewEncoder(stdout)
			encoder.SetIndent("", "  ")
			if err := encoder.Encode(records); err != nil {
				return failure(stderr, fmt.Errorf("write evidence inventory: %w", err))
			}
			return 0
		}
		_, _ = fmt.Fprintf(stdout, "%d valid evidence record(s)\n", len(records))
		return 0
	default:
		return usage(stderr, "unknown command: "+args[0])
	}
}

func executeCohesion(args []string, root string, policy config.Config, stdout, stderr io.Writer) int {
	if len(args) >= 1 && args[0] == "aggregate" {
		action, inputs, output, err := cohesionAggregateArguments(args[1:])
		if err != nil {
			return usage(stderr, "usage: golib cohesion aggregate <generate|check> --inputs <file> --output <directory>")
		}
		inputs = resolveCohesionPath(root, inputs)
		output = resolveCohesionPath(root, output)
		generateAggregate := cohesion.GenerateAggregate
		if action == "generate" && buildinfo.Version == "dev" {
			publicOutput, pathError := sameCohesionPath(output, filepath.Join(root, "docs", "ecosystem"))
			if pathError != nil {
				return failure(stderr, fmt.Errorf("resolve cohesion publication path: %w", pathError))
			}
			if publicOutput {
				return failure(stderr, errors.New("source build cannot publish cohesion catalogs; use a released checksummed binary"))
			}
			generateAggregate = func(inputsPath, outputDirectory string, identity cohesion.Identity) error {
				return cohesion.GenerateAggregateProtected(inputsPath, outputDirectory, filepath.Join(root, "docs", "ecosystem"), identity)
			}
		}
		identity := currentCohesionIdentity()
		if action == "generate" {
			err = generateAggregate(inputs, output, identity)
		} else {
			err = cohesion.CheckAggregate(inputs, output, identity)
		}
		if err != nil {
			return failure(stderr, err)
		}
		_, _ = fmt.Fprintf(stdout, "cohesion aggregate %s passed\n", action)
		return 0
	}
	if len(args) >= 1 && args[0] == "catalog" {
		if len(args) < 2 || len(args) > 3 || (args[1] != "consumer" && args[1] != "engineering") || (len(args) == 3 && args[2] != "--json") {
			return usage(stderr, "usage: golib cohesion catalog <consumer|engineering> [--json]")
		}
		catalog, report := cohesion.LoadAndCheck(root, policy)
		if !report.Valid {
			return failure(stderr, errors.New("cohesion metadata is invalid"))
		}
		envelope, err := cohesion.Project(catalog, args[1], currentCohesionIdentity())
		if err != nil {
			return failure(stderr, err)
		}
		if len(args) == 3 {
			encoder := json.NewEncoder(stdout)
			encoder.SetIndent("", "  ")
			if err := encoder.Encode(envelope); err != nil {
				return failure(stderr, fmt.Errorf("write cohesion catalog: %w", err))
			}
			return 0
		}
		markdown, err := renderCohesionCatalog(envelope)
		if err != nil {
			return failure(stderr, err)
		}
		if _, err := stdout.Write(markdown); err != nil {
			return failure(stderr, fmt.Errorf("write cohesion catalog: %w", err))
		}
		return 0
	}
	if len(args) < 1 || args[0] != "check" || len(args) > 2 || (len(args) == 2 && args[1] != "--json") {
		return usage(stderr, "usage: golib cohesion <check [--json]|catalog <consumer|engineering> [--json]|aggregate <generate|check> --inputs <file> --output <directory>>")
	}
	report := cohesion.Check(root, policy)
	if len(args) == 2 {
		encoder := json.NewEncoder(stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(report); err != nil {
			return failure(stderr, fmt.Errorf("write cohesion report: %w", err))
		}
		if report.Valid {
			return 0
		}
		return 1
	}
	if report.Valid {
		_, _ = fmt.Fprintf(stdout, "cohesion metadata valid: %d module(s)\n", *report.Summary.TotalModules)
		return 0
	}
	for _, diagnostic := range report.Diagnostics {
		_, _ = fmt.Fprintf(stdout, "%s: %s: %s\n", diagnostic.Code, diagnostic.Path, diagnostic.Message)
	}
	return 1
}

func cohesionAggregateArguments(args []string) (string, string, string, error) {
	if len(args) != 5 || (args[0] != "generate" && args[0] != "check") {
		return "", "", "", errors.New("invalid cohesion aggregate arguments")
	}
	values := map[string]string{}
	for index := 1; index < len(args); index += 2 {
		if (args[index] != "--inputs" && args[index] != "--output") || args[index+1] == "" {
			return "", "", "", errors.New("invalid cohesion aggregate arguments")
		}
		if _, exists := values[args[index]]; exists {
			return "", "", "", errors.New("duplicate cohesion aggregate argument")
		}
		values[args[index]] = args[index+1]
	}
	return args[0], values["--inputs"], values["--output"], nil
}

func resolveCohesionPath(root, value string) string {
	if filepath.IsAbs(value) {
		return filepath.Clean(value)
	}
	return filepath.Join(root, value)
}

func sameCohesionPath(left, right string) (bool, error) {
	return sameCohesionPathWithCanonicalizer(left, right, canonicalCohesionPath)
}

func sameCohesionPathWithCanonicalizer(left, right string, canonicalize func(string) (string, error)) (bool, error) {
	leftCanonical, err := canonicalize(left)
	if err != nil {
		return false, err
	}
	rightCanonical, err := canonicalize(right)
	if err != nil {
		return false, err
	}
	return leftCanonical == rightCanonical, nil
}

func canonicalCohesionPath(value string) (string, error) {
	return canonicalCohesionPathWithFunctions(value, filepath.Abs, filepath.EvalSymlinks)
}

func canonicalCohesionPathWithFunctions(value string, absolutePath func(string) (string, error), evaluateSymlinks func(string) (string, error)) (string, error) {
	absolute, err := absolutePath(value)
	if err != nil {
		return "", err
	}
	current := filepath.Clean(absolute)
	missing := make([]string, 0)
	for {
		resolved, err := evaluateSymlinks(current)
		if err == nil {
			for _, component := range slices.Backward(missing) {
				resolved = filepath.Join(resolved, component)
			}
			return filepath.Clean(resolved), nil
		}
		if !errors.Is(err, os.ErrNotExist) {
			return "", err
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", err
		}
		missing = append(missing, filepath.Base(current))
		current = parent
	}
}

func currentCohesionIdentity() cohesion.Identity {
	publicationStatus := "unpublished"
	if buildinfo.Version != "dev" {
		publicationStatus = "published"
	}
	return cohesion.Identity{
		DesignLanguageVersion: buildinfo.DesignLanguageVersion,
		DesignLanguageSHA256:  buildinfo.DesignLanguageSHA256,
		SourceIdentity:        buildinfo.DesignLanguageSourceIdentity,
		ToolingVersion:        buildinfo.Version,
		PublicationStatus:     publicationStatus,
	}
}

type upgradeOptions struct {
	action  string
	request upgrade.Request
	json    bool
}

func executeUpgrade(args []string, workingDirectory string, stdout, stderr io.Writer) int {
	options, err := parseUpgradeArguments(args)
	if err != nil {
		return usage(stderr, err.Error()+"\nusage: golib upgrade <plan|apply> --version <version> --workflow-sha <sha> --checksums-sha256 <digest> [--json]")
	}
	root, err := findRoot(workingDirectory)
	if err != nil {
		return failure(stderr, err)
	}
	var result upgrade.Result
	if options.action == "apply" {
		result, err = upgrade.Apply(root, options.request)
	} else {
		result, err = upgrade.Plan(root, options.request)
	}
	if err != nil {
		return failure(stderr, err)
	}
	if options.json {
		encoder := json.NewEncoder(stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(result); err != nil {
			return failure(stderr, fmt.Errorf("write upgrade result: %w", err))
		}
		return 0
	}
	if _, err := io.WriteString(stdout, result.Human(options.action == "apply")); err != nil {
		return failure(stderr, fmt.Errorf("write upgrade result: %w", err))
	}
	return 0
}

func parseUpgradeArguments(args []string) (upgradeOptions, error) {
	if len(args) == 0 {
		return upgradeOptions{}, errors.New("upgrade action is required")
	}
	if args[0] != "plan" && args[0] != "apply" {
		return upgradeOptions{}, errors.New("upgrade action must be plan or apply")
	}
	options := upgradeOptions{action: args[0]}
	values := make(map[string]string, 3)
	for index := 1; index < len(args); {
		if args[index] == "--json" {
			if options.json {
				return upgradeOptions{}, errors.New("upgrade JSON flag is duplicated")
			}
			options.json = true
			index++
		} else {
			if index+1 >= len(args) {
				return upgradeOptions{}, errors.New("upgrade flag has no value")
			}
			flag, value := args[index], args[index+1]
			if value == "" || (flag != "--version" && flag != "--workflow-sha" && flag != "--checksums-sha256") {
				return upgradeOptions{}, errors.New("upgrade arguments are malformed")
			}
			if _, duplicate := values[flag]; duplicate {
				return upgradeOptions{}, errors.New("upgrade argument is duplicated")
			}
			values[flag] = value
			index += 2
		}
	}
	if len(values) != 3 {
		return upgradeOptions{}, errors.New("upgrade release identity is incomplete")
	}
	options.request = upgrade.Request{
		Version: values["--version"], WorkflowSHA: values["--workflow-sha"], ChecksumsSHA256: values["--checksums-sha256"],
	}
	return options, nil
}

type mutationImportOptions struct {
	module  string
	archive string
	ledger  string
}

func mutationImportArguments(args []string) (mutationImportOptions, error) {
	if len(args) != 6 {
		return mutationImportOptions{}, errors.New("mutation import requires module, archive, and ledger paths")
	}
	values := make(map[string]string, 3)
	for index := 0; index < len(args); index += 2 {
		flag, value := args[index], args[index+1]
		if value == "" || (flag != "--module" && flag != "--archive" && flag != "--ledger") {
			return mutationImportOptions{}, errors.New("mutation import arguments are malformed")
		}
		if _, duplicate := values[flag]; duplicate {
			return mutationImportOptions{}, errors.New("mutation import arguments are duplicated")
		}
		values[flag] = value
	}
	return mutationImportOptions{module: values["--module"], archive: values["--archive"], ledger: values["--ledger"]}, nil
}

func withExecutor(root string, stdout, stderr io.Writer, create executorFactory, run func(gates.Executor) error) int {
	executor, cleanup, err := create(root, stdout, stderr)
	if err != nil {
		return failure(stderr, err)
	}
	runError := run(executor)
	cleanupError := cleanup()
	if runError != nil {
		return failure(stderr, runError)
	}
	if cleanupError != nil {
		return failure(stderr, cleanupError)
	}
	return 0
}

func moduleSelection(args []string, modules []inventory.Module) ([]string, error) {
	if len(args) == 0 || (len(args) == 1 && args[0] == "--all") {
		selection := make([]string, 0, len(modules))
		for _, module := range modules {
			selection = append(selection, module.Directory)
		}
		return selection, nil
	}
	if len(args) == 2 && args[0] == "--module" && args[1] != "" {
		return []string{args[1]}, nil
	}
	return nil, errors.New("usage: golib check [--all|--module <directory>]")
}

func findRoot(start string) (string, error) {
	if !filepath.IsAbs(start) {
		return "", errors.New("locate repository root: working directory must be absolute")
	}
	current := filepath.Clean(start)
	for {
		info, statErr := os.Stat(filepath.Join(current, ".golib.yaml"))
		if statErr == nil && !info.IsDir() {
			return current, nil
		}
		if statErr != nil && !errors.Is(statErr, os.ErrNotExist) {
			return "", fmt.Errorf("locate repository root: %w", statErr)
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", errors.New("locate repository root: .golib.yaml not found")
		}
		current = parent
	}
}

func usage(stderr io.Writer, message string) int {
	_, _ = fmt.Fprintln(stderr, message)
	return 2
}

func failure(stderr io.Writer, err error) int {
	_, _ = fmt.Fprintln(stderr, err)
	return 1
}
