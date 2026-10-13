package mutation

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
)

// The process boundary supplies the same primary package shape as native
// go list. An inactive source produces a real LIVED report if the campaign
// includes it, instead of asserting only the spelling of command arguments.
func TestCampaignMutatesOnlyBuildSelectedProductionFiles(t *testing.T) {
	for _, platform := range []struct {
		name     string
		selected string
		inactive []string
	}{
		{"linux", "rename_linux.go", []string{"rename_windows.go", "rename_darwin.go"}},
		{"windows", "rename_windows.go", []string{"rename_linux.go", "rename_darwin.go"}},
		{"darwin", "rename_darwin.go", []string{"rename_linux.go", "rename_windows.go"}},
	} {
		t.Run(platform.name, func(t *testing.T) {
			campaign, process := campaignFixture(t)
			campaign.RuntimeIdentity.GOOS = platform.name
			for _, name := range []string{"rename_linux.go", "rename_windows.go", "rename_darwin.go", "rename_information.go"} {
				if err := os.WriteFile(filepath.Join(campaign.Root, name), []byte("package example\nfunc Value"+strings.ReplaceAll(name, ".", "_")+"() int { return 1 }\n"), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			campaign.Process = func(ctx context.Context, name string, args []string, directory string, environment map[string]string, stdout, stderr io.Writer) error {
				if name == "go" && len(args) > 0 && args[0] == "list" {
					return json.NewEncoder(stdout).Encode(listedPackage{
						Dir: campaign.Root, ImportPath: "example",
						GoFiles: []string{"source.go", "rename_information.go", platform.selected},
						Module:  &listedModule{Path: "example", Main: true, GoVersion: "1.27.0"},
					})
				}
				if strings.HasSuffix(name, "golib-gremlins") {
					excluded := func(filename string) bool {
						for index, argument := range args {
							if argument != "--exclude-files" || index+1 >= len(args) {
								continue
							}
							for expression := range strings.SplitSeq(args[index+1], ",") {
								matched, err := regexp.MatchString(expression, filename)
								if err != nil {
									t.Fatal(err)
								}
								if matched {
									return true
								}
							}
						}
						return false
					}
					for _, selected := range []string{"source.go", "rename_information.go", platform.selected} {
						if excluded(selected) {
							return fmt.Errorf("build-selected production source excluded: %s", selected)
						}
					}
					for _, inactive := range platform.inactive {
						if excluded(inactive) {
							continue
						}
						outputIndex := slices.Index(args, "--output")
						if outputIndex < 0 || outputIndex+1 == len(args) {
							t.Fatal("missing report output")
						}
						report := fmt.Sprintf(`{"files":[{"file_name":%q,"mutations":[{"type":"INVERT_LOGICAL","status":"LIVED","line":2,"column":1}]}]}`, inactive)
						if err := os.WriteFile(args[outputIndex+1], []byte(report), 0o600); err != nil {
							return err
						}
						return mutationExitCode(10)
					}
				}
				return process.run(ctx, name, args, directory, environment, stdout, stderr)
			}
			if err := campaign.Run(context.Background()); err != nil {
				t.Fatalf("inactive source reached native mutation campaign: %v", err)
			}
			if process.mutations != 1 {
				t.Fatalf("selected production campaign executions=%d, want 1", process.mutations)
			}
		})
	}
}
