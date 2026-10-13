package mutation

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestProductionSelectionUsesNativeBuildContext(t *testing.T) {
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	for name, data := range map[string]string{
		"go.mod":            "module example\ngo 1.27.0\n",
		"shared.go":         "package example\nconst Shared = 1\n",
		"shared_test.go":    "package example\nimport \"testing\"\nfunc TestShared(t *testing.T) {}\n",
		"native_linux.go":   "package example\nconst Native = 1\n",
		"native_windows.go": "package example\nconst Native = 2\n",
		"tagged.go":         "//go:build selected\n\npackage example\nconst Tagged = 1\n",
	} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(data), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	for _, platform := range []struct{ os, arch, native string }{
		{"linux", "amd64", "native_linux.go"},
		{"windows", "amd64", "native_windows.go"},
		{"windows", "386", "native_windows.go"},
	} {
		t.Run(platform.os+"/"+platform.arch, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
			defer cancel()
			command := exec.CommandContext(ctx, "go", "list", "-deps", "-test", "-json", "-tags=selected", ".")
			command.Dir = root
			command.Env = append(os.Environ(), "GOWORK=off", "GOOS="+platform.os, "GOARCH="+platform.arch, "CGO_ENABLED=0")
			output, err := command.Output()
			if err != nil {
				t.Fatalf("native package admission: %v", err)
			}
			selection, err := selectProduction(root, ".", ".", "example", strings.NewReader(string(output)))
			if err != nil {
				t.Fatal(err)
			}
			want := []string{platform.native, "shared.go", "tagged.go"}
			if !reflect.DeepEqual(selection.Files, want) {
				t.Fatalf("selected=%v want=%v", selection.Files, want)
			}
			command = exec.CommandContext(ctx, "go", "list", "-deps", "-test", "-json", ".")
			command.Dir = root
			command.Env = append(os.Environ(), "GOWORK=off", "GOOS="+platform.os, "GOARCH="+platform.arch, "CGO_ENABLED=0")
			output, err = command.Output()
			if err != nil {
				t.Fatal(err)
			}
			selection, err = selectProduction(root, ".", ".", "example", strings.NewReader(string(output)))
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(selection.Files, want[:2]) {
				t.Fatalf("inactive tag selected=%v", selection.Files)
			}
		})
	}
}

func TestProductionSelectionRefusesAmbiguousOrUnsafeTarget(t *testing.T) {
	campaign, _ := campaignFixture(t)
	root, err := filepath.EvalSymlinks(campaign.Root)
	if err != nil {
		t.Fatal(err)
	}
	valid := listedPackage{Dir: root, ImportPath: "example", GoFiles: []string{"source.go"}, Module: &listedModule{Path: "example", Main: true}}
	for _, test := range []struct {
		name      string
		change    func(*listedPackage)
		duplicate bool
	}{
		{"missing target", func(p *listedPackage) { p.ImportPath = "other" }, false},
		{"test variant only", func(p *listedPackage) { p.ForTest = "example" }, false},
		{"missing module", func(p *listedPackage) { p.Module = nil }, false},
		{"foreign module", func(p *listedPackage) { p.Module = &listedModule{Path: "other"} }, false},
		{"foreign directory", func(p *listedPackage) { p.Dir = filepath.Dir(root) }, false},
		{"empty production", func(p *listedPackage) { p.GoFiles = nil }, false},
		{"parent filename", func(p *listedPackage) { p.GoFiles = []string{"../source.go"} }, false},
		{"missing filename", func(p *listedPackage) { p.GoFiles = []string{"absent.go"} }, false},
		{"duplicate filename", func(p *listedPackage) { p.GoFiles = []string{"source.go", "source.go"} }, false},
		{"duplicate target", func(*listedPackage) {}, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			listed := valid
			test.change(&listed)
			data, err := json.Marshal(listed)
			if err != nil {
				t.Fatal(err)
			}
			if test.duplicate {
				data = append(append(data, '\n'), data...)
			}
			if _, err := selectProduction(root, ".", ".", "example", strings.NewReader(string(data))); !errors.Is(err, ErrInvalid) {
				t.Fatalf("unsafe selection accepted: %v", err)
			}
		})
	}
}

func TestProductionSelectionIncludesCgoAndExactRegexNames(t *testing.T) {
	campaign, _ := campaignFixture(t)
	root, err := filepath.EvalSymlinks(campaign.Root)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"native_cgo.go", `inactive,quoted"[x].go`} {
		if err := os.WriteFile(filepath.Join(root, name), []byte("package example\n"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	listed := listedPackage{Dir: root, ImportPath: "example", GoFiles: []string{"source.go"}, CgoFiles: []string{"native_cgo.go"}, Module: &listedModule{Path: "example", Main: true}}
	data, err := json.Marshal(listed)
	if err != nil {
		t.Fatal(err)
	}
	selection, err := selectProduction(root, ".", ".", "example", strings.NewReader(string(data)))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(selection.Files, []string{"native_cgo.go", "source.go"}) {
		t.Fatalf("cgo source omitted: %v", selection.Files)
	}
	expression := selection.exclusion()
	if strings.Contains(expression, ",") || strings.Contains(expression, `"`) {
		t.Fatal("filename breaks StringSlice CSV argument")
	}
	pattern, err := regexp.Compile(expression)
	if err != nil {
		t.Fatal(err)
	}
	if !pattern.MatchString(`inactive,quoted"[x].go`) || pattern.MatchString(`inactive,quoted"[x].go.extra`) || pattern.MatchString("source.go") || pattern.MatchString("native_cgo.go") {
		t.Fatal("exclusion changed selected/other filename domain")
	}
}

func TestNativeSelectionChangesInvalidateEvidence(t *testing.T) {
	selection := productionSelection{Files: []string{"shared.go"}, Excluded: []string{"native_windows.go"}}
	runtime := RuntimeIdentity{GoVersion: "go1.27.2", GOOS: "linux", GOARCH: "amd64", CGOEnabled: "0"}
	original := selectedInputDigest("input", selection, runtime)
	other := runtime
	other.GOOS = "windows"
	if original == selectedInputDigest("input", selection, other) {
		t.Fatal("OS change reused evidence")
	}
	other = runtime
	other.GOARCH = "386"
	if original == selectedInputDigest("input", selection, other) {
		t.Fatal("architecture change reused evidence")
	}
	changed := productionSelection{Files: []string{"native_windows.go", "shared.go"}, Excluded: []string{}}
	if original == selectedInputDigest("input", changed, runtime) || original == selectedInputDigest("changed source", selection, runtime) {
		t.Fatal("source/selection change reused evidence")
	}
}

func TestLegacyCheckpointCannotAdoptAmbiguousPlatform(t *testing.T) {
	for _, field := range []string{"GOOS", "GOARCH"} {
		for _, value := range []string{"", "different"} {
			t.Run(field+"/"+value, func(t *testing.T) {
				campaign, _ := campaignFixture(t)
				_, input, err := campaign.packageInput(t.Context(), ".")
				if err != nil {
					t.Fatal(err)
				}
				checkpoint, ledger := approvedImportFixture(t, input)
				checkpoint.Environment[field] = value
				if err := campaign.Import(t.Context(), []Checkpoint{checkpoint}, ledger); !errors.Is(err, ErrInputChanged) {
					t.Fatalf("ambiguous platform checkpoint accepted: %v", err)
				}
			})
		}
	}
}
