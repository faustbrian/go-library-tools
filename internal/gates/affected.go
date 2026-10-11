package gates

import (
	"context"
	"errors"
	"io/fs"
	"path"
	"slices"
	"strings"

	"github.com/faustbrian/go-library-tools/v2/internal/inventory"
	"github.com/faustbrian/go-library-tools/v2/internal/repositoryfile"
	"golang.org/x/mod/modfile"
)

// ChangedModules selects the module owners and repository-local consumers of
// committed changes between an immutable base and HEAD. Unknown shared inputs
// and an empty diff select every module; missing Git evidence is an error.
// It does not certify uncommitted or untracked changes.
func (runner Runner) ChangedModules(ctx context.Context, base string) ([]inventory.Module, error) {
	if len(base) != 40 || strings.IndexFunc(base, func(c rune) bool { return !strings.ContainsRune("0123456789abcdef", c) }) >= 0 {
		return nil, errors.New("affected inventory requires a full lowercase Git commit SHA")
	}
	if len(runner.Catalog.Modules) == 0 || len(runner.Catalog.Modules) > inventory.MaximumModules {
		return nil, errors.New("affected inventory module cardinality is invalid")
	}
	if _, err := runner.changeGit(ctx, "cat-file", "-e", base+"^{commit}"); err != nil {
		return nil, err
	}
	diff, err := runner.changeGit(ctx, "diff", "--no-ext-diff", "--no-textconv", "--no-renames", "--name-only", "-z", base, "HEAD", "--")
	if err != nil {
		return nil, err
	}
	selected := make(map[string]bool, len(runner.Catalog.Modules))
	all := diff == ""
	if diff != "" && !strings.HasSuffix(diff, "\x00") {
		return nil, errors.New("affected inventory Git paths are malformed")
	}
	for name := range strings.SplitSeq(strings.TrimSuffix(diff, "\x00"), "\x00") {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if all {
			break
		}
		if !fs.ValidPath(name) || path.Clean(name) != name || name == "." {
			return nil, errors.New("affected inventory Git paths are malformed")
		}
		owner := ""
		for _, module := range runner.Catalog.Modules {
			if module.Directory != "." && strings.HasPrefix(name, module.Directory+"/") && len(module.Directory) > len(owner) {
				owner = module.Directory
			}
		}
		if owner == "" {
			// Only explicit root source/package ownership permits narrowing. Root
			// policy, workflows, manifests, tools and other unknown inputs widen.
			for _, module := range runner.Catalog.Modules {
				if module.Directory != "." {
					continue
				}
				if !strings.Contains(name, "/") && strings.HasSuffix(name, ".go") {
					owner = "."
				}
				for _, pkg := range module.Packages {
					if pkg.Directory != "." && strings.HasPrefix(name, pkg.Directory+"/") {
						owner = "."
					}
				}
			}
		}
		if owner == "" {
			all = true
		} else {
			selected[owner] = true
		}
	}
	if all {
		return slices.Clone(runner.Catalog.Modules), ctx.Err()
	}

	// Read actual owned module requirements as well as declared relationships:
	// a stale optional manifest edge must not hide a direct Go consumer.
	dependencies := make(map[string][]string, len(runner.Catalog.Modules))
	for _, module := range runner.Catalog.Modules {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		data, err := repositoryfile.Read(runner.Root, path.Join(module.Directory, "go.mod"), 1<<20)
		if err != nil {
			return nil, errors.New("affected inventory cannot read owned module dependencies")
		}
		file, err := modfile.Parse("go.mod", data, nil)
		if err != nil || file.Module == nil || file.Module.Mod.Path != module.ModulePath {
			return nil, errors.New("affected inventory cannot parse owned module dependencies")
		}
		dependencies[module.Directory] = slices.Clone(module.OwnedDependencies)
		for _, requirement := range file.Require {
			dependencies[module.Directory] = append(dependencies[module.Directory], requirement.Mod.Path)
		}
		for _, replacement := range file.Replace {
			if replacement.New.Version != "" {
				continue
			}
			target := path.Clean(path.Join(module.Directory, replacement.New.Path))
			for _, owner := range runner.Catalog.Modules {
				if target == owner.Directory {
					dependencies[module.Directory] = append(dependencies[module.Directory], owner.ModulePath)
				}
			}
		}
	}
	for added := true; added; {
		added = false
		for _, module := range runner.Catalog.Modules {
			if selected[module.Directory] {
				continue
			}
			for _, dependency := range runner.Catalog.Modules {
				if selected[dependency.Directory] && slices.Contains(dependencies[module.Directory], dependency.ModulePath) {
					selected[module.Directory], added = true, true
					break
				}
			}
		}
	}
	modules := make([]inventory.Module, 0, len(selected))
	for _, module := range runner.Catalog.Modules {
		if selected[module.Directory] {
			modules = append(modules, module)
		}
	}
	return modules, ctx.Err()
}

func (runner Runner) changeGit(ctx context.Context, args ...string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	stdout := &sourceInventoryOutput{limit: maximumSecurityProcessOutput}
	stderr := &boundedProcessOutput{limit: maximumSecurityProcessOutput}
	command := Command{Name: "git", Dir: runner.Root,
		Args:   append([]string{"--no-replace-objects", "-c", "core.fsmonitor=false"}, args...),
		Env:    map[string]string{"GIT_CONFIG_GLOBAL": "/dev/null", "GIT_CONFIG_NOSYSTEM": "1"},
		Stdout: stdout, Stderr: stderr, boundedScanner: true}
	if err := runner.Executor.Run(ctx, command); err != nil || stdout.didOverflow() || stderr.didOverflow() {
		return "", errors.Join(sourceCommandError{class: "affected inventory Git comparison failed or exceeded output limit", cause: err}, ctx.Err())
	}
	return stdout.String(), ctx.Err()
}
