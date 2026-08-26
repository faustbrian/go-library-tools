package mutation_test

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/faustbrian/go-library-tools/internal/mutation"
)

func TestArgumentsMatchPinnedCampaignContract(t *testing.T) {
	got, err := mutation.Arguments("./adapter", "/tmp/report.json", "integration,postgres", false, 4)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"unleash", "./adapter", "--integration", "--coverpkg", "./adapter",
		"--exclude-files", "^.+/", "--workers", "4", "--test-cpu", "1",
		"--timeout-coefficient", "10", "--threshold-efficacy", "100",
		"--threshold-mcover", "100", "--arithmetic-base", "--conditionals-boundary",
		"--conditionals-negation", "--invert-assignments", "--invert-bitwise",
		"--invert-bwassign", "--increment-decrement", "--invert-logical",
		"--invert-loopctrl", "--invert-negatives", "--remove-self-assignments",
		"--output-statuses", "lctvsr", "--output", "/tmp/report.json",
		"--tags", "integration,postgres",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Arguments() = %#v", got)
	}
	discovery, err := mutation.Arguments(".", "/tmp/report.json", "", true, 1)
	if err != nil || discovery[len(discovery)-1] != "--dry-run" {
		t.Fatalf("Arguments(discovery) = %#v, %v", discovery, err)
	}
}

func TestBuildVerifierUsesPinnedSourceAndAssets(t *testing.T) {
	workspace := t.TempDir()
	source := t.TempDir()
	if err := os.WriteFile(filepath.Join(source, "go.mod"), []byte("module example\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	process := func(_ context.Context, name string, args []string, directory string, environment map[string]string, stdout, _ io.Writer) error {
		if name == "go" {
			for _, variable := range []string{"GOCACHE", "GOMODCACHE", "GOTMPDIR", "GOBIN"} {
				relative, err := filepath.Rel(workspace, environment[variable])
				if err != nil || filepath.IsAbs(relative) || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
					t.Fatalf("%s = %q", variable, environment[variable])
				}
			}
			if environment["GOWORK"] != "off" {
				t.Fatalf("GOWORK = %q", environment["GOWORK"])
			}
		}
		switch {
		case name == "go" && len(args) > 1 && args[0] == "mod":
			_, _ = io.WriteString(stdout, `{"Dir":"`+source+`","Sum":"h1:3G2ROO0I3q4bb5bxElQIUITTuEbl1iOfVYFqunGwrJI=","GoModSum":"h1:LLbvJR33CWsu1sgvQ4qMzU2rqkwYJK3Qy/Al59eHKjA="}`)
		case name == "git":
			if directory != filepath.Join(workspace, "gremlins-source") {
				t.Fatalf("git directory = %q", directory)
			}
		case name == "go" && args[0] == "build":
			output := args[4]
			if err := os.WriteFile(output, []byte("verified binary"), 0o700); err != nil {
				return err
			}
		default:
			t.Fatalf("unexpected command: %s %v", name, args)
		}
		return nil
	}
	tool, err := mutation.BuildVerifier(context.Background(), workspace, process)
	if err != nil {
		t.Fatalf("BuildVerifier() error = %v", err)
	}
	if !filepath.IsAbs(tool.Path) || !strings.HasPrefix(tool.Digest, "sha256:") {
		t.Fatalf("BuildVerifier() = %#v", tool)
	}
}
