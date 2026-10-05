package gates

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/faustbrian/go-library-tools/v2/internal/inventory"
)

func TestGosecReceivesGoSelectedDirectories(t *testing.T) {
	root := t.TempDir()
	stop := errors.New("inert gosec boundary")
	var selected []string
	runner := Runner{Root: root, Executor: executorFunction(func(_ context.Context, command Command) error {
		if slices.Contains(command.Args, "list") {
			if !slices.Equal(command.Args, []string{"list", "-json=Dir", "./..."}) || command.Env["GOWORK"] != "off" || !command.boundedScanner || command.Stdout == nil || command.Stderr == nil {
				t.Fatal("discovery is not the bounded, workspace-independent ordinary Go selection")
			}
			for _, directory := range []string{"testsupport", ".", "newpackage"} {
				if err := json.NewEncoder(command.Stdout).Encode(map[string]string{"Dir": filepath.Join(root, directory)}); err != nil {
					return err
				}
			}
		}
		if slices.Contains(command.Args, "github.com/securego/gosec/v2/cmd/gosec@"+gosecVersion) {
			if !slices.Equal(command.Args[:5], []string{"run", "github.com/securego/gosec/v2/cmd/gosec@" + gosecVersion, "-fmt=json", "-nosec-require-rules", "-nosec-require-justification"}) {
				t.Fatal("structured reporting changed strict scanner flags")
			}
			selected = command.Args[5:]
			return stop
		}
		return nil
	})}
	// No package manifest is supplied: newly added and test-support source must
	// come from ordinary Go selection, not a stale production inventory.
	err := runner.runSecurity(t.Context(), io.Discard, root, inventory.Module{Directory: "."})
	if !errors.Is(err, stop) || !slices.Equal(selected, []string{"./", "./newpackage", "./testsupport"}) {
		t.Fatalf("selected = %q, error = %v", selected, err)
	}
}

func TestGosecDiscoveryFailsClosed(t *testing.T) {
	for _, name := range []string{"command failure", "empty", "malformed", "missing directory", "outside", "relative", "recursive suffix", "duplicate", "stdout overflow", "stderr overflow", "canceled"} {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			var scanner bool
			runner := Runner{Executor: executorFunction(func(_ context.Context, command Command) error {
				if slices.Contains(command.Args, "list") {
					text := `{"Dir":` + string(mustSelectionJSON(t, root)) + `}`
					switch name {
					case "command failure":
						return errors.New("private discovery detail")
					case "empty":
						text = ""
					case "malformed":
						text = "private discovery detail"
					case "missing directory":
						text = `{}`
					case "outside":
						text = `{"Dir":` + string(mustSelectionJSON(t, filepath.Dir(root))) + `}`
					case "relative":
						text = `{"Dir":"relative"}`
					case "recursive suffix":
						text = `{"Dir":` + string(mustSelectionJSON(t, filepath.Join(root, "recursive..."))) + `}`
					case "duplicate":
						text += text
					case "stdout overflow":
						text = strings.Repeat("x", maximumSecurityProcessOutput+1)
					case "stderr overflow":
						_, _ = io.WriteString(command.Stderr, strings.Repeat("x", maximumSecurityProcessOutput+1))
					case "canceled":
						cancel()
					}
					_, _ = io.WriteString(command.Stdout, text)
				}
				if slices.Contains(command.Args, "github.com/securego/gosec/v2/cmd/gosec@"+gosecVersion) {
					scanner = true
					return errors.New("inert scanner")
				}
				return nil
			})}
			err := runner.runSecurity(ctx, io.Discard, root, inventory.Module{Directory: "."})
			if err == nil || scanner || strings.Contains(err.Error(), "private discovery detail") {
				t.Fatalf("failed closed = %v, error = %v", !scanner, err)
			}
		})
	}
}

func mustSelectionJSON(t *testing.T, value string) []byte {
	t.Helper()
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return encoded
}

func TestGosecOrdinaryGoSelectionIncludesNewAndSupportPackages(t *testing.T) {
	root := t.TempDir()
	for name, content := range map[string]string{
		"go.mod":                      "module example.invalid/selection\n\ngo 1.27.0\n",
		"root.go":                     "package selection\n",
		"newpackage/new.go":           "package newpackage\n",
		"testsupport/support.go":      "package testsupport\n",
		"testdata/fixture/fixture.go": "package fixture\n",
	} {
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	stop := errors.New("inert gosec boundary")
	var selected []string
	runner := Runner{Executor: executorFunction(func(ctx context.Context, command Command) error {
		if slices.Contains(command.Args, "list") {
			// #nosec G204 -- executable and arguments are the ordinary Go selection from the production gate under test
			process := exec.CommandContext(ctx, command.Name, command.Args...)
			process.Dir, process.Stdout, process.Stderr = command.Dir, command.Stdout, command.Stderr
			process.Env = append(os.Environ(), "GOWORK=off")
			return process.Run()
		}
		if slices.Contains(command.Args, "github.com/securego/gosec/v2/cmd/gosec@"+gosecVersion) {
			selected = command.Args[5:]
			return stop
		}
		return nil
	})}
	err := runner.runSecurity(t.Context(), io.Discard, root, inventory.Module{Directory: "."})
	if !errors.Is(err, stop) || !slices.Equal(selected, []string{"./", "./newpackage", "./testsupport"}) {
		t.Fatalf("selected = %q, error = %v", selected, err)
	}
}
