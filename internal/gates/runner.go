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

	"github.com/faustbrian/go-library-tools/internal/inventory"
)

// Command is one external process invocation without shell interpretation.
type Command struct {
	Name string
	Args []string
	Dir  string
	Env  map[string]string
}

// Executor runs one external command.
type Executor interface {
	Run(context.Context, Command) error
}

// Runner executes gates for modules in one validated repository.
type Runner struct {
	Root     string
	Catalog  inventory.Inventory
	Executor Executor
	Output   io.Writer
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
		if err := runner.checkModule(ctx, output, module); err != nil {
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
	if err := runner.command(ctx, output, module.Directory, "tidy-check", directory, "go", "mod", "tidy", "-diff"); err != nil {
		return err
	}
	if err := announce(output, module.Directory, "safety", func() error {
		return checkSafety(directory)
	}); err != nil {
		return err
	}
	if module.Gates["lint"] {
		if err := runner.command(ctx, output, module.Directory, "vet", directory, "go", "vet", "./..."); err != nil {
			return err
		}
	}
	if module.Gates["tests"] {
		args := testArguments(module.TestTags, false)
		if err := runner.command(ctx, output, module.Directory, "test", directory, "go", args...); err != nil {
			return err
		}
	}
	if module.Gates["race"] {
		args := testArguments(module.TestTags, true)
		if err := runner.command(ctx, output, module.Directory, "race", directory, "go", args...); err != nil {
			return err
		}
	}
	return nil
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

func (runner Runner) command(ctx context.Context, output io.Writer, module, gate, directory, name string, args ...string) error {
	return announce(output, module, gate, func() error {
		if err := runner.Executor.Run(ctx, Command{Name: name, Args: args, Dir: directory, Env: map[string]string{"GOWORK": "off"}}); err != nil {
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
