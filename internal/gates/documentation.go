package gates

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

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
	if module.Directory != "." {
		if operation, exists := runner.operation(module.Directory, "docs"); exists {
			return runner.runOperation(ctx, directory, module, operation)
		}
		if err := runner.Executor.Run(ctx, Command{
			Name: "go", Args: []string{"test", "./...", "-run=^Example", "-count=1", "-timeout=20m"},
			Dir: directory, Env: map[string]string{"GOWORK": "off"},
		}); err != nil {
			return fmt.Errorf("check documentation examples: %w", err)
		}
		return nil
	}
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
	links := runner.DocumentationLinks
	if links == nil {
		links = runner.runDocumentationLinks
	}
	if err := links(ctx, directory); err != nil {
		return err
	}
	if operation, exists := runner.operation(module.Directory, "docs"); exists {
		return runner.runOperation(ctx, directory, module, operation)
	}
	return nil
}

func (runner Runner) runDocumentationLinks(ctx context.Context, directory string) error {
	workspace, ok := runner.Executor.(taskWorkspace)
	if !ok || !filepath.IsAbs(workspace.TemporaryDirectory()) {
		return errors.New("documentation links require a task-owned workspace")
	}
	releaseFor := runner.documentationRelease
	if releaseFor == nil {
		releaseFor = docscheck.LycheeReleaseFor
	}
	release, err := releaseFor(runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return err
	}
	toolRoot := filepath.Join(workspace.TemporaryDirectory(), "documentation", "links")
	if err := os.MkdirAll(toolRoot, 0o700); err != nil {
		return fmt.Errorf("create link tool root: %w", err)
	}
	archive := filepath.Join(toolRoot, "lychee.tar.gz")
	if err := runner.Executor.Run(ctx, Command{
		Name: "curl",
		Args: []string{
			"--fail", "--silent", "--show-error", "--location", "--proto", "=https", "--tlsv1.2",
			release.URL, "--output", archive,
		},
		Dir: toolRoot,
	}); err != nil {
		return fmt.Errorf("download pinned link checker: %w", err)
	}
	extract := runner.documentationExtract
	if extract == nil {
		extract = docscheck.ExtractLychee
	}
	binary, err := extract(archive, release)
	if err != nil {
		return err
	}
	lychee := filepath.Join(toolRoot, "lychee")
	if err := os.WriteFile(lychee, binary, 0o700); err != nil {
		return fmt.Errorf("write link checker: %w", err)
	}
	if err := runner.Executor.Run(ctx, Command{
		Name: lychee,
		Args: []string{
			"--cache=false",
			"--exclude", `^https://pkg\.go\.dev/(badge/)?github\.com/faustbrian/go-`,
			"--exclude", `^https://doi\.org/10\.1145/190314\.190317$`,
			"--exclude", `^https://service\.unece\.org/trade/`,
			"--exclude-private", "--exclude-loopback",
			"--exclude", `^https://www\.iso\.org/standard/`,
			"--max-concurrency", "16", "--max-retries", "3", "--no-progress",
			"README.md", "docs/**/*.md",
		},
		Dir: directory,
	}); err != nil {
		return fmt.Errorf("check documentation links: %w", err)
	}
	return nil
}
