package gates

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/faustbrian/go-library-tools/internal/docscheck"
	"github.com/faustbrian/go-library-tools/internal/inventory"
	"github.com/faustbrian/go-library-tools/internal/repositoryfile"
)

const maximumSpellingConfigSize = 1 << 20

func (runner Runner) runDocumentationSpelling(ctx context.Context, directory string) error {
	workspace, ok := runner.Executor.(taskWorkspace)
	if !ok || !filepath.IsAbs(workspace.TemporaryDirectory()) {
		return errors.New("documentation spelling requires a task-owned workspace")
	}
	if _, err := repositoryfile.Read(directory, "cspell.json", maximumSpellingConfigSize); err != nil {
		return fmt.Errorf("read cspell configuration: %w", err)
	}
	toolRoot := filepath.Join(workspace.TemporaryDirectory(), "documentation", "spelling")
	cache := filepath.Join(workspace.TemporaryDirectory(), "documentation", "npm-cache")
	if err := os.MkdirAll(toolRoot, 0o700); err != nil {
		return fmt.Errorf("create spelling tool root: %w", err)
	}
	packageJSON, packageLock := docscheck.SpellingAssets()
	assets := []struct {
		name string
		data []byte
	}{
		{name: "package.json", data: packageJSON},
		{name: "package-lock.json", data: packageLock},
	}
	for _, asset := range assets {
		if err := os.WriteFile(filepath.Join(toolRoot, asset.name), asset.data, 0o600); err != nil {
			return fmt.Errorf("write spelling tool %s: %w", asset.name, err)
		}
	}
	environment := map[string]string{
		"NPM_CONFIG_CACHE": cache,
		"NPM_CONFIG_AUDIT": "false",
		"NPM_CONFIG_FUND":  "false",
	}
	if err := runner.Executor.Run(ctx, Command{
		Name: "npm", Args: []string{"ci", "--ignore-scripts", "--no-audit", "--no-fund", "--silent"},
		Dir: toolRoot, Env: environment,
	}); err != nil {
		return fmt.Errorf("install pinned spelling tool: %w", err)
	}
	cspell := filepath.Join(toolRoot, "node_modules", ".bin", "cspell")
	if err := runner.Executor.Run(ctx, Command{
		Name: cspell,
		Args: []string{
			"lint", "--config", filepath.Join(directory, "cspell.json"), "--no-config-search",
			"--validate-directives", "--no-progress", "--no-summary", "README.md", "docs/**/*.md",
		},
		Dir: directory,
	}); err != nil {
		return fmt.Errorf("check documentation spelling: %w", err)
	}
	return nil
}

func (runner Runner) checkDocumentation(ctx context.Context, directory string, module inventory.Module) error {
	if err := docscheck.Check(directory); err != nil {
		return err
	}
	spelling := runner.DocumentationSpelling
	if spelling == nil {
		spelling = runner.runDocumentationSpelling
	}
	if err := spelling(ctx, directory); err != nil {
		return err
	}
	if operation, exists := runner.operation(module.Directory, "docs"); exists {
		return runner.runOperation(ctx, directory, module, operation)
	}
	return nil
}
