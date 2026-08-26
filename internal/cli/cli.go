// Package cli implements the stable golib command-line contract.
package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/faustbrian/go-library-tools/internal/config"
	"github.com/faustbrian/go-library-tools/internal/inventory"
)

const help = `golib validates and executes the Go library repository contract.

Usage:
  golib check [--all|--module <directory>]
  golib config validate
  golib inventory [--json]
  golib repository check
  golib coverage [--module <directory>]
  golib mutation [--module <directory>]
  golib api check
  golib api update
  golib docs check
  golib services start <fixture>
  golib services stop <fixture>
  golib release check
  golib release dry-run
  golib evidence inspect
`

// Execute runs one command and returns a stable process exit code.
func Execute(args []string, workingDirectory string, stdout, stderr io.Writer) int {
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
	catalog, err := inventory.Load(root, policy)
	if err != nil {
		return failure(stderr, err)
	}

	switch args[0] {
	case "config":
		if len(args) != 2 || args[1] != "validate" {
			return usage(stderr, "usage: golib config validate")
		}
		_, _ = fmt.Fprintf(stdout, "configuration valid: %s\n", catalog.Repository)
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
	default:
		return usage(stderr, fmt.Sprintf("unknown command: %s", args[0]))
	}
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
