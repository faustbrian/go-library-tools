package gates

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/faustbrian/go-library-tools/v2/internal/inventory"
)

func inertDiscovery(t *testing.T, root string, directories []string) ([]string, error, error) {
	t.Helper()
	stop := errors.New("inert selected scanner boundary")
	var selected []string
	runner := Runner{Executor: workspaceExecutor{directory: t.TempDir(), emptyGoPackageInventory: true, run: func(_ context.Context, command Command) error {
		if slices.Contains(command.Args, "list") {
			for _, directory := range directories {
				if err := json.NewEncoder(command.Stdout).Encode(map[string]string{"Dir": directory}); err != nil {
					return err
				}
			}
		}
		if filepath.Base(command.Name) == "gosec" {
			if len(command.Args) < 3 {
				t.Fatal("structured scanner arguments missing")
			}
			selected = append([]string(nil), command.Args[3:]...)
			return stop
		}
		return nil
	}}}
	err := runner.runSecurity(t.Context(), io.Discard, root, inventory.Module{Directory: "."})
	return selected, err, stop
}

func TestGosecDiscoveryRejectsNoncanonicalAndControlPaths(t *testing.T) {
	for name, suffix := range map[string]string{
		"noncanonical": "/ordinary/../safe",
		"newline":      "/ordinary\n",
		"carriage":     "/ordinary\r",
		"nul":          "/ordinary\x00",
	} {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			selected, err, _ := inertDiscovery(t, root, []string{root + suffix})
			if err == nil || err.Error() != "gosec package discovery returned an invalid directory" || selected != nil {
				t.Fatalf("invalid discovery reached scanner = %t, error = %v", selected != nil, err)
			}
		})
	}
}

func TestGosecDiscoveryInclusiveArgumentBudgets(t *testing.T) {
	for _, test := range []struct {
		name          string
		count, length int
		oneLonger     bool
		valid         bool
	}{
		{"exact package count", 4096, 8, false, true},
		{"excess package count", 4097, 8, false, false},
		{"exact argument length", 1, 4096, false, true},
		{"excess argument length", 1, 4097, false, false},
		{"exact aggregate bytes", 32, 4095, false, true},
		{"one excess aggregate byte", 32, 4095, true, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			directories := make([]string, 0, test.count)
			want := make([]string, 0, test.count)
			for index := range test.count {
				length := test.length
				if test.oneLonger && index == 0 {
					length++
				}
				// No real long path is created. Two prefix and four suffix bytes
				// leave the stated argument length; discovery adds one separator.
				argument := "./" + strings.Repeat("x", length-6) + fmt.Sprintf("%04d", index)
				want = append(want, argument)
				directories = append(directories, root+"/"+argument[2:])
			}
			slices.Sort(want)
			selected, err, stop := inertDiscovery(t, root, directories)
			if test.valid {
				if !errors.Is(err, stop) || !slices.Equal(selected, want) {
					t.Fatal("inclusive discovery budget did not dispatch the complete sorted selection")
				}
			} else if err == nil || err.Error() != "gosec package discovery returned duplicate or excessive directories" || selected != nil {
				t.Fatalf("excessive discovery reached scanner = %t, error = %v", selected != nil, err)
			}
		})
	}
}
