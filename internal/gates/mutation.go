package gates

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/faustbrian/go-library-tools/internal/config"
	"github.com/faustbrian/go-library-tools/internal/inventory"
	"github.com/faustbrian/go-library-tools/internal/mutation"
	"github.com/faustbrian/go-library-tools/internal/repositoryfile"
)

const maximumModuleFileSize = 16 << 20

type mutationCampaignRunner func(context.Context, mutation.Campaign) error
type mutationImportRunner func(context.Context, mutation.Campaign, []mutation.Checkpoint, mutation.MigrationLedger) error

type mutationFileSystem interface {
	MkdirAll(string, os.FileMode) error
	WriteFile(string, []byte, os.FileMode) error
}

type operatingMutationFiles struct{}

func (operatingMutationFiles) MkdirAll(path string, mode os.FileMode) error {
	return os.MkdirAll(path, mode)
}

func (operatingMutationFiles) WriteFile(path string, data []byte, mode os.FileMode) error {
	return os.WriteFile(path, data, mode)
}

// Mutation runs exact package-level mutation verification for selected modules.
func (runner Runner) Mutation(ctx context.Context, selection []string) error {
	modules, err := runner.selectModules(selection)
	if err != nil {
		return err
	}
	output := runner.Output
	if output == nil {
		output = io.Discard
	}
	for _, module := range modules {
		if !module.Gates["mutation"] {
			_, _ = fmt.Fprintf(output, "[%s] mutation: not applicable\n", module.Directory)
			continue
		}
		if err := runner.withModuleServices(ctx, module, func(scoped Runner) error {
			if migration, exists := scoped.configuredMutationImport(module.Directory); exists {
				if err := announce(output, module.Directory, "mutation-import", func() error {
					return scoped.runMutationImport(ctx, output, module, migration.Archive, migration.Ledger)
				}); err != nil {
					if !errors.Is(err, mutation.ErrInputChanged) {
						return err
					}
					_, _ = fmt.Fprintf(output, "[%s] mutation-import: changed packages require current verification\n", module.Directory)
				}
			}
			return announce(output, module.Directory, "mutation", func() error {
				return scoped.runMutation(ctx, output, module)
			})
		}); err != nil {
			return err
		}
	}
	return nil
}

func (runner Runner) configuredMutationImport(module string) (config.MutationImport, bool) {
	for _, migration := range runner.Policy.Mutation.Imports {
		if migration.Module == module {
			return migration, true
		}
	}
	return config.MutationImport{}, false
}

// MutationImport migrates one module's approved legacy checkpoint archive.
func (runner Runner) MutationImport(ctx context.Context, moduleDirectory, archivePath, ledgerPath string) error {
	module, err := selectedMutationModule(runner.Catalog.Modules, moduleDirectory)
	if err != nil {
		return err
	}
	if !module.Gates["mutation"] {
		return fmt.Errorf("module %s does not enable mutation verification", module.Directory)
	}
	output := runner.Output
	if output == nil {
		output = io.Discard
	}
	return runner.withModuleServices(ctx, module, func(scoped Runner) error {
		return announce(output, module.Directory, "mutation-import", func() error {
			return scoped.runMutationImport(ctx, output, module, archivePath, ledgerPath)
		})
	})
}

func selectedMutationModule(modules []inventory.Module, directory string) (inventory.Module, error) {
	for _, module := range modules {
		if module.Directory == directory {
			return module, nil
		}
	}
	return inventory.Module{}, fmt.Errorf("unknown module %q", directory)
}

func (runner Runner) runMutation(ctx context.Context, output io.Writer, module inventory.Module) error {
	campaign, err := runner.prepareMutationCampaign(ctx, output, module)
	if err != nil {
		return err
	}
	run := runner.mutationCampaign
	if run == nil {
		run = func(ctx context.Context, campaign mutation.Campaign) error { return campaign.Run(ctx) }
	}
	return run(ctx, campaign)
}

func (runner Runner) runMutationImport(ctx context.Context, output io.Writer, module inventory.Module, archivePath, ledgerPath string) error {
	archive, err := repositoryfile.Read(runner.Root, archivePath, mutation.MaximumArchiveSize)
	if err != nil {
		return fmt.Errorf("read mutation checkpoint archive: %w", err)
	}
	checkpoints, err := mutation.ReadBootstrap(bytes.NewReader(archive), int64(len(archive)))
	if err != nil {
		return fmt.Errorf("parse mutation checkpoint archive: %w", err)
	}
	ledgerData, err := repositoryfile.Read(runner.Root, ledgerPath, mutation.MaximumMigrationLedgerSize)
	if err != nil {
		return fmt.Errorf("read mutation migration ledger: %w", err)
	}
	ledger, err := mutation.ParseMigrationLedger(bytes.NewReader(ledgerData))
	if err != nil {
		return fmt.Errorf("parse mutation migration ledger: %w", err)
	}
	campaign, err := runner.prepareMutationCampaign(ctx, output, module)
	if err != nil {
		return err
	}
	run := runner.mutationImport
	if run == nil {
		run = func(ctx context.Context, campaign mutation.Campaign, checkpoints []mutation.Checkpoint, ledger mutation.MigrationLedger) error {
			return campaign.Import(ctx, checkpoints, ledger)
		}
	}
	return run(ctx, campaign, checkpoints, ledger)
}

func (runner Runner) prepareMutationCampaign(ctx context.Context, output io.Writer, module inventory.Module) (mutation.Campaign, error) {
	workspace, ok := runner.Executor.(taskWorkspace)
	if !ok || !filepath.IsAbs(workspace.TemporaryDirectory()) {
		return mutation.Campaign{}, errors.New("mutation requires a task-owned workspace")
	}
	if len(module.RequiredServices) != len(runner.serviceIdentities) {
		return mutation.Campaign{}, fmt.Errorf("mutation service identities are incomplete for %s", module.Directory)
	}
	for _, service := range module.RequiredServices {
		if runner.serviceIdentities[service] == "" {
			return mutation.Campaign{}, fmt.Errorf("mutation service identity for %s is missing", service)
		}
	}
	mutationRoot := filepath.Join(runner.Root, filepath.FromSlash(runner.Policy.Mutation.Root))
	reviews, err := loadZeroInventory(runner.Root, filepath.Join(runner.Policy.Mutation.Root, "zero-inventory.json"))
	if err != nil {
		return mutation.Campaign{}, err
	}
	runtimeIdentity, err := runner.runtimeIdentity(ctx, module.Directory)
	if err != nil {
		return mutation.Campaign{}, err
	}
	owned := runner.localOwnedModules(module)
	campaignWorkspace := filepath.Join(
		workspace.TemporaryDirectory(), "mutation-campaigns", mutationModuleSlug(module.ModulePath),
	)
	environment, err := runner.mutationEnvironment(ctx, campaignWorkspace, module, owned)
	if err != nil {
		return mutation.Campaign{}, err
	}
	packages := make([]string, 0, len(module.Packages))
	for _, pkg := range module.Packages {
		if pkg.CoverageRequired {
			packages = append(packages, pkg.Directory)
		}
	}
	sort.Strings(packages)
	testTags := make([]string, 0, len(module.TestTags))
	for _, tag := range module.TestTags {
		if tag != "interoperability" {
			testTags = append(testTags, tag)
		}
	}
	workers, err := runner.mutationWorkers(module)
	if err != nil {
		return mutation.Campaign{}, err
	}
	return mutation.Campaign{
		Root:         runner.Root,
		EvidenceRoot: filepath.Join(runner.Root, filepath.FromSlash(runner.Policy.Evidence.Root)),
		MutationRoot: mutationRoot,
		Workspace:    campaignWorkspace,
		Policy: mutation.CampaignPolicy{
			Repository: runner.Catalog.Repository, ModuleDirectory: module.Directory,
			ModulePath: module.ModulePath, GoVersion: module.GoVersion, Packages: packages,
			TestTags: testTags, BuildTags: append([]string(nil), module.BuildTags...),
			RequiredServices:  append([]string(nil), module.RequiredServices...),
			ServiceIdentities: cloneServiceMap(runner.serviceIdentities), OwnedModules: owned, Workers: workers,
		},
		ZeroReviews: reviews, Environment: environment, RuntimeIdentity: runtimeIdentity,
		Process: runner.mutationProcess(), Output: output,
	}, nil
}

func loadZeroInventory(root, path string) (mutation.ZeroInventory, error) {
	data, err := repositoryfile.Read(root, path, maximumModuleFileSize)
	if err != nil {
		return mutation.ZeroInventory{}, fmt.Errorf("read zero-mutant inventory: %w", err)
	}
	reviews, err := mutation.ParseZeroInventory(bytes.NewReader(data))
	if err != nil {
		return mutation.ZeroInventory{}, fmt.Errorf("parse zero-mutant inventory: %w", err)
	}
	return reviews, nil
}

func (runner Runner) runtimeIdentity(ctx context.Context, module string) (mutation.RuntimeIdentity, error) {
	var output boundedBuffer
	directory := filepath.Join(runner.Root, filepath.FromSlash(module))
	err := runner.Executor.Run(ctx, Command{
		Name: "go", Args: []string{"env", "-json", "GOVERSION", "GOOS", "GOARCH", "CGO_ENABLED"},
		Dir: directory, Env: map[string]string{"GOWORK": "off"}, Stdout: &output, Stderr: io.Discard,
	})
	if err != nil {
		return mutation.RuntimeIdentity{}, fmt.Errorf("read Go runtime identity: %w", err)
	}
	identity, err := mutation.ParseRuntimeIdentity(bytes.NewReader(output.Bytes()))
	if err != nil {
		return mutation.RuntimeIdentity{}, fmt.Errorf("parse Go runtime identity: %w", err)
	}
	return identity, nil
}

func (runner Runner) localOwnedModules(module inventory.Module) []mutation.OwnedModule {
	declared := make(map[string]struct{}, len(module.OwnedDependencies))
	for _, dependency := range module.OwnedDependencies {
		declared[dependency] = struct{}{}
	}
	owned := make([]mutation.OwnedModule, 0, len(declared))
	for _, candidate := range runner.Catalog.Modules {
		if _, exists := declared[candidate.ModulePath]; !exists || candidate.ModulePath == module.ModulePath || candidate.Kind == "fixture" {
			continue
		}
		owned = append(owned, mutation.OwnedModule{ModulePath: candidate.ModulePath, Directory: candidate.Directory})
	}
	slices.SortFunc(owned, func(left, right mutation.OwnedModule) int {
		return strings.Compare(left.ModulePath, right.ModulePath)
	})
	return owned
}

func (runner Runner) mutationEnvironment(ctx context.Context, workspace string, module inventory.Module, owned []mutation.OwnedModule) (map[string]string, error) {
	environment := map[string]string{}
	if len(owned) == 0 {
		return environment, nil
	}
	moduleFile, err := repositoryfile.Read(runner.Root, filepath.Join(module.Directory, "go.mod"), maximumModuleFileSize)
	if err != nil {
		return nil, fmt.Errorf("read mutation go.mod: %w", err)
	}
	files := runner.mutationFiles
	if files == nil {
		files = operatingMutationFiles{}
	}
	directory := filepath.Join(workspace, "mutation-modfiles")
	if err := files.MkdirAll(directory, 0o700); err != nil {
		return nil, fmt.Errorf("create mutation modfile directory: %w", err)
	}
	modfile := filepath.Join(directory, mutationModuleSlug(module.ModulePath)+".mod")
	if err := files.WriteFile(modfile, moduleFile, 0o600); err != nil {
		return nil, fmt.Errorf("write mutation modfile: %w", err)
	}
	sum, err := repositoryfile.Read(runner.Root, filepath.Join(module.Directory, "go.sum"), maximumModuleFileSize)
	if err == nil {
		if err := files.WriteFile(strings.TrimSuffix(modfile, ".mod")+".sum", sum, 0o600); err != nil {
			return nil, fmt.Errorf("write mutation sumfile: %w", err)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("read mutation go.sum: %w", err)
	}
	for _, dependency := range owned {
		replacement := dependency.ModulePath + "=" + filepath.Join(runner.Root, filepath.FromSlash(dependency.Directory))
		if err := runner.Executor.Run(ctx, Command{
			Name: "go", Args: []string{"mod", "edit", "-modfile=" + modfile, "-replace=" + replacement},
			Dir: filepath.Join(runner.Root, filepath.FromSlash(module.Directory)), Env: map[string]string{"GOWORK": "off"},
		}); err != nil {
			return nil, fmt.Errorf("configure mutation owned module %s: %w", dependency.ModulePath, err)
		}
	}
	environment["GOFLAGS"] = "-modfile=" + modfile + " -mod=mod"
	return environment, nil
}

func (runner Runner) mutationWorkers(module inventory.Module) (int, error) {
	moduleFile, err := repositoryfile.Read(runner.Root, filepath.Join(module.Directory, "go.mod"), maximumModuleFileSize)
	if err != nil {
		return 0, fmt.Errorf("read mutation worker policy: %w", err)
	}
	if bytes.Contains(moduleFile, []byte("github.com/testcontainers/testcontainers-go")) {
		return 1, nil
	}
	return 4, nil
}

func (runner Runner) mutationProcess() mutation.Process {
	return func(ctx context.Context, name string, args []string, directory string, environment map[string]string, stdout, stderr io.Writer) error {
		return runner.Executor.Run(ctx, Command{Name: name, Args: args, Dir: directory, Env: environment, Stdout: stdout, Stderr: stderr})
	}
}

func mutationModuleSlug(modulePath string) string {
	digest := sha256.Sum256([]byte(modulePath))
	return hex.EncodeToString(digest[:])
}
