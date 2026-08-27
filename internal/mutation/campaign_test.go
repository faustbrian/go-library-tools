package mutation

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/faustbrian/go-library-tools/internal/evidence"
)

func TestCampaignExecutesPersistsAndReusesPackageEvidence(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "source.go"), []byte("package example\n\nfunc Value() int { return 1 }\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, "adapter"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "adapter", "adapter.go"), []byte("package adapter\n\nfunc Value() int { return 2 }\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	source := t.TempDir()
	if err := os.WriteFile(filepath.Join(source, "go.mod"), []byte("module verifier\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	process := &campaignProcess{root: root, verifierSource: source}
	process.requireTags = true
	var output bytes.Buffer
	campaign := Campaign{
		Root: root, EvidenceRoot: filepath.Join(root, ".verification"),
		MutationRoot: filepath.Join(root, ".verification", "mutation"), Workspace: filepath.Join(root, ".task"),
		Policy: CampaignPolicy{
			Repository: "example", ModuleDirectory: ".", ModulePath: "example", GoVersion: "1.27.0",
			Packages: []string{"adapter", "."}, TestTags: []string{"integration"},
			ServiceIdentities: map[string]string{}, Workers: 2,
		},
		Environment:     map[string]string{"SECRET_SERVICE_PASSWORD": "must-not-persist"},
		RuntimeIdentity: RuntimeIdentity{GoVersion: "go1.27.0", GOOS: "linux", GOARCH: "amd64", CGOEnabled: "0"},
		Process:         process.run,
		Output:          &output, Now: func() time.Time { return time.Unix(10, 0) },
	}
	if err := campaign.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if process.mutations != 2 || !strings.Contains(output.String(), "killed 1/1") {
		t.Fatalf("first campaign mutations/output = %d, %q", process.mutations, output.String())
	}
	records, err := evidence.Inspect(root, ".verification", "example", []string{"."})
	if err != nil || len(records) != 2 {
		t.Fatalf("Inspect() = %#v, %v", records, err)
	}
	for _, record := range records {
		if _, leaked := record.Environment["SECRET_SERVICE_PASSWORD"]; leaked {
			t.Fatal("command secret persisted in mutation evidence")
		}
		if record.Environment["GOOS"] != "linux" || record.Environment["GOARCH"] != "amd64" {
			t.Fatalf("runtime identity = %#v", record.Environment)
		}
	}
	output.Reset()
	if err := campaign.Run(context.Background()); err != nil {
		t.Fatalf("Run(reuse) error = %v", err)
	}
	if process.mutations != 2 || strings.Count(output.String(), "reused content-identical") != 2 {
		t.Fatalf("reused campaign mutations/output = %d, %q", process.mutations, output.String())
	}
}

type campaignProcess struct {
	root           string
	verifierSource string
	mutations      int
	listCalls      int
	fail           string
	report         *string
	skipReport     bool
	mutateSource   bool
	afterMutation  func() error
	requireTags    bool
}

func (process *campaignProcess) run(_ context.Context, name string, args []string, _ string, environment map[string]string, stdout, _ io.Writer) error {
	switch {
	case name == "go" && len(args) > 1 && args[0] == "list":
		if containsArgument(args, "-tags=") {
			return errors.New("empty list tags")
		}
		if process.requireTags && !containsArgument(args, "-tags=integration") {
			return errors.New("missing list tags")
		}
		process.listCalls++
		if process.fail == "list" || process.fail == "second-list" && process.listCalls > 1 {
			return errors.New("list failed")
		}
		directory := process.root
		importPath := "example"
		goFiles := []string{"source.go"}
		if args[len(args)-1] == "./adapter" {
			directory = filepath.Join(process.root, "adapter")
			importPath = "example/adapter"
			goFiles = []string{"adapter.go"}
		}
		listing := listedPackage{
			Dir: directory, ImportPath: importPath, GoFiles: goFiles,
			Module: &listedModule{Path: "example", Main: true, GoVersion: "1.27.0"},
		}
		return json.NewEncoder(stdout).Encode(listing)
	case name == "go" && len(args) > 1 && args[0] == "mod":
		if process.fail == "download" {
			return errors.New("download failed")
		}
		_, err := io.WriteString(stdout, validDownloadMetadata(process.verifierSource))
		return err
	case name == "git":
		return nil
	case name == "go" && args[0] == "build":
		if process.fail == "build" {
			return errors.New("build failed")
		}
		return os.WriteFile(args[4], []byte("verifier"), 0o700)
	case name == "go" && args[0] == "test":
		if process.requireTags && !containsArgument(args, "-tags=integration") {
			return errors.New("missing test tags")
		}
		if process.fail == "coverage" {
			return errors.New("coverage failed")
		}
		for _, argument := range args {
			if argument == "-tags=" {
				return errors.New("empty tags argument")
			}
			if profile, found := strings.CutPrefix(argument, "-coverprofile="); found {
				return os.WriteFile(profile, []byte("mode: set\n"), 0o600)
			}
		}
	case strings.HasSuffix(name, "golib-gremlins"):
		process.mutations++
		if process.fail == "mutation" {
			return errors.New("mutation failed")
		}
		if environment["GOLIB_GREMLINS_COVERAGE_PROFILE"] == "" || environment["GOCACHE"] == "" {
			return errors.New("missing isolated mutation environment")
		}
		if !process.skipReport {
			value := `{"files":[{"file_name":"source.go","mutations":[{"type":"A","status":"KILLED","line":3,"column":1}]}]}`
			if process.report != nil {
				value = *process.report
			}
			for index, argument := range args {
				if argument == "--output" {
					if err := os.WriteFile(args[index+1], []byte(value), 0o600); err != nil {
						return err
					}
				}
			}
		}
		if process.mutateSource {
			if err := os.WriteFile(filepath.Join(process.root, "source.go"), []byte("package example\n\nfunc Value() int { return 9 }\n"), 0o600); err != nil {
				return err
			}
		}
		if process.afterMutation != nil {
			return process.afterMutation()
		}
	}
	return nil
}

func containsArgument(arguments []string, expected string) bool {
	return slices.Contains(arguments, expected)
}
