package gates

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/faustbrian/go-library-tools/v2/internal/config"
	"github.com/faustbrian/go-library-tools/v2/internal/inventory"
	"github.com/faustbrian/go-library-tools/v2/internal/mutation"
	"github.com/faustbrian/go-library-tools/v2/internal/repositoryfile"
)

func TestMutationEquivalentInventoryRefusalBeforeEffects(t *testing.T) {
	for _, directory := range []bool{true, false} {
		name := "malformed-json"
		if directory {
			name = "directory"
		}
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			writeValidMutationSetup(t, root)
			path := filepath.Join(root, ".verification", "mutation", "equivalent-inventory.json")
			if directory {
				if err := os.Mkdir(path, 0o700); err != nil {
					t.Fatal(err)
				}
				path = filepath.Join(path, "unrelated")
			}
			if err := os.WriteFile(path, []byte("{"), 0o600); err != nil {
				t.Fatal(err)
			}
			before := map[string][]byte{}
			for _, source := range []string{filepath.Join(root, "go.mod"), filepath.Join(root, ".verification", "mutation", "zero-inventory.json"), path} {
				data, err := os.ReadFile(source)
				if err != nil {
					t.Fatal(err)
				}
				before[source] = data
			}
			var output bytes.Buffer
			commands, campaigns := 0, 0
			runner := Runner{
				Root:   root,
				Policy: config.Config{Mutation: config.Mutation{Root: ".verification/mutation"}},
				Executor: workspaceExecutor{directory: t.TempDir(), run: func(context.Context, Command) error {
					commands++
					return errors.New("unexpected execution")
				}},
				mutationCampaign: func(context.Context, mutation.Campaign) error {
					campaigns++
					return errors.New("unexpected campaign")
				},
			}
			err := runner.runMutation(t.Context(), &output, inventory.Module{Directory: ".", ModulePath: "example", GoVersion: "1.27.0"})
			want := "parse equivalent-mutant inventory: "
			if directory {
				want = "read equivalent-mutant inventory: "
				if !errors.Is(err, repositoryfile.ErrNotRegular) {
					t.Fatal("directory refusal lost non-regular-file identity")
				}
			} else if !errors.Is(err, mutation.ErrInvalid) {
				t.Fatal("malformed inventory refusal lost invalid-inventory identity")
			}
			if err == nil || !strings.HasPrefix(err.Error(), want) || commands != 0 || campaigns != 0 || output.Len() != 0 {
				t.Fatalf("inventory refusal: error=%v commands=%d campaigns=%d output bytes=%d", err, commands, campaigns, output.Len())
			}
			for source, expected := range before {
				actual, readErr := os.ReadFile(source)
				if readErr != nil || !bytes.Equal(actual, expected) {
					t.Fatal("inventory refusal changed source or unrelated bytes")
				}
			}
		})
	}
}
