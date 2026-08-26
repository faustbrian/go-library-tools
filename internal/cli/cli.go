// Package cli implements the stable golib command-line contract.
package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/faustbrian/go-library-tools/internal/buildinfo"
	"github.com/faustbrian/go-library-tools/internal/config"
	"github.com/faustbrian/go-library-tools/internal/evidence"
	"github.com/faustbrian/go-library-tools/internal/gates"
	"github.com/faustbrian/go-library-tools/internal/inventory"
	"github.com/faustbrian/go-library-tools/internal/repository"
)

const help = `golib validates and executes the Go library repository contract.

Usage:
  golib --version
  golib check [--all|--module <directory>]
  golib config validate
  golib inventory [--json]
  golib repository check
  golib coverage [--module <directory>]
  golib mutation [--module <directory>]
  golib api check
  golib api update
	golib docs check [--module <directory>]
  golib services start <fixture>
  golib services stop <fixture>
  golib release check
  golib release dry-run
  golib evidence inspect
`

// Execute runs one command and returns a stable process exit code.
func Execute(args []string, workingDirectory string, stdout, stderr io.Writer) int {
	if len(args) == 1 && args[0] == "--version" {
		_, _ = fmt.Fprintln(stdout, buildinfo.Version)
		return 0
	}
	return execute(args, workingDirectory, stdout, stderr, gates.NewProcessExecutor)
}

type executorFactory func(string, io.Writer, io.Writer) (gates.Executor, func() error, error)

func execute(args []string, workingDirectory string, stdout, stderr io.Writer, createExecutor executorFactory) int {
	if len(args) == 1 && (args[0] == "--help" || args[0] == "-h") {
		_, _ = io.WriteString(stdout, help)
		return 0
	}
	if len(args) == 0 {
		return usage(stderr, "command is required")
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
			return (gates.Runner{Root: root, Catalog: catalog, Policy: policy, Executor: executor, Output: stdout}).Check(context.Background(), selection)
		})
	case "config":
		if len(args) != 2 || args[1] != "validate" {
			return usage(stderr, "usage: golib config validate")
		}
		_, _ = fmt.Fprintf(stdout, "configuration valid: %s\n", catalog.Repository)
		return 0
	case "repository":
		if len(args) != 2 || args[1] != "check" {
			return usage(stderr, "usage: golib repository check")
		}
		if err := repository.Check(root, catalog); err != nil {
			return failure(stderr, err)
		}
		_, _ = io.WriteString(stdout, "standalone repository contract passed\n")
		return 0
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
	case "api":
		if len(args) < 2 || (args[1] != "check" && args[1] != "update") {
			return usage(stderr, "usage: golib api <check|update> [--module <directory>]")
		}
		selection, usageError := moduleSelection(args[2:], catalog.Modules)
		if usageError != nil {
			return usage(stderr, "usage: golib api <check|update> [--module <directory>]")
		}
		return withExecutor(root, stdout, stderr, createExecutor, func(executor gates.Executor) error {
			return (gates.Runner{Root: root, Catalog: catalog, Policy: policy, Executor: executor, Output: stdout}).API(context.Background(), selection, args[1] == "update")
		})
	case "coverage":
		selection, usageError := moduleSelection(args[1:], catalog.Modules)
		if usageError != nil {
			return usage(stderr, "usage: golib coverage [--module <directory>]")
		}
		return withExecutor(root, stdout, stderr, createExecutor, func(executor gates.Executor) error {
			return (gates.Runner{Root: root, Catalog: catalog, Policy: policy, Executor: executor, Output: stdout}).Coverage(context.Background(), selection)
		})
	case "mutation":
		selection, usageError := moduleSelection(args[1:], catalog.Modules)
		if usageError != nil {
			return usage(stderr, "usage: golib mutation [--module <directory>]")
		}
		return withExecutor(root, stdout, stderr, createExecutor, func(executor gates.Executor) error {
			return (gates.Runner{Root: root, Catalog: catalog, Policy: policy, Executor: executor, Output: stdout}).Mutation(context.Background(), selection)
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
			return (gates.Runner{Root: root, Catalog: catalog, Policy: policy, Executor: executor, Output: stdout}).Docs(context.Background(), selection)
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
		return usage(stderr, fmt.Sprintf("unknown command: %s", args[0]))
	}
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
