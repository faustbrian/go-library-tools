package mutation

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestInputDigestRejectsInvalidInputs(t *testing.T) {
	root := t.TempDir()
	valid := InputPolicy{ModuleDirectory: ".", PackageDirectory: ".", ModulePath: "example", GoVersion: "1.26.6"}
	listing := func(files string) string {
		return `{"Dir":"` + root + `","ImportPath":"example","GoFiles":` + files + `,"Module":{"Path":"example","Main":true}}`
	}
	tests := map[string]struct {
		root    string
		policy  InputPolicy
		listing string
	}{
		"relative root":  {root: ".", policy: valid, listing: listing(`[]`)},
		"bad policy":     {root: root, policy: InputPolicy{}, listing: listing(`[]`)},
		"malformed list": {root: root, policy: valid, listing: "{"},
		"missing target": {root: root, policy: valid, listing: `{"Dir":"/external","ImportPath":"external"}`},
		"empty target":   {root: root, policy: valid, listing: listing(`[]`)},
		"missing file":   {root: root, policy: valid, listing: listing(`["missing.go"]`)},
		"escaping file":  {root: root, policy: valid, listing: listing(`["../outside.go"]`)},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := InputDigest(test.root, test.policy, strings.NewReader(test.listing), nil); !errors.Is(err, ErrInvalid) && name != "missing file" {
				t.Fatalf("InputDigest() error = %v, want ErrInvalid", err)
			} else if err == nil {
				t.Fatal("InputDigest() error = nil")
			}
		})
	}
}

func TestInputDigestRejectsConflictingModuleIdentityAndDataSymlink(t *testing.T) {
	root := t.TempDir()
	writeMutationInput(t, root, "target.go", "package example\n")
	policy := InputPolicy{ModuleDirectory: ".", PackageDirectory: ".", ModulePath: "example", GoVersion: "1.26.6"}
	conflict := strings.Join([]string{
		`{"Dir":"` + root + `","ImportPath":"example","GoFiles":["target.go"],"Module":{"Path":"example","Main":true}}`,
		`{"Dir":"` + root + `","ImportPath":"example","GoFiles":["target.go"],"Module":{"Path":"example","Version":"v1.0.0"}}`,
	}, "\n")
	if _, err := InputDigest(root, policy, strings.NewReader(conflict), nil); !errors.Is(err, ErrInvalid) {
		t.Fatalf("InputDigest(conflict) error = %v", err)
	}
	if err := os.Symlink(root, filepath.Join(root, "testdata")); err != nil {
		t.Fatal(err)
	}
	valid := `{"Dir":"` + root + `","ImportPath":"example","GoFiles":["target.go"],"Module":{"Path":"example","Main":true}}`
	if _, err := InputDigest(root, policy, strings.NewReader(valid), nil); !errors.Is(err, ErrInvalid) {
		t.Fatalf("InputDigest(data symlink) error = %v", err)
	}
}

func TestInputDigestResolvesNestedTargetImport(t *testing.T) {
	root := t.TempDir()
	writeMutationInput(t, root, "nested/target.go", "package nested\n")
	policy := InputPolicy{ModuleDirectory: ".", PackageDirectory: "nested", ModulePath: "example", GoVersion: "1.26.6"}
	listing := `{"Dir":"` + filepath.Join(root, "nested") + `","ImportPath":"example/nested","GoFiles":["target.go"],"Module":{"Path":"example","Main":true}}`
	if _, err := InputDigest(root, policy, strings.NewReader(listing), nil); err != nil {
		t.Fatalf("InputDigest() error = %v", err)
	}
}

func TestInputDigestIgnoresGeneratedTestExecutable(t *testing.T) {
	root := t.TempDir()
	writeMutationInput(t, root, "target.go", "package example\n")
	policy := InputPolicy{ModuleDirectory: ".", PackageDirectory: ".", ModulePath: "example", GoVersion: "1.27.0"}
	listing := strings.Join([]string{
		`{"Dir":"` + root + `","ImportPath":"example","GoFiles":["target.go"],"Module":{"Path":"example","Main":true}}`,
		`{"Dir":"` + root + `","ImportPath":"example.test","GoFiles":["` + filepath.Join(t.TempDir(), "go-build", "generated") + `"],"Module":{"Path":"example","Main":true}}`,
	}, "\n")
	if _, err := InputDigest(root, policy, strings.NewReader(listing), nil); err != nil {
		t.Fatalf("InputDigest() error = %v", err)
	}
}

func TestInputDigestRejectsMismatchedZeroReview(t *testing.T) {
	root := t.TempDir()
	policy := InputPolicy{ModuleDirectory: ".", PackageDirectory: ".", ModulePath: "example", GoVersion: "1.26.6"}
	review := &ZeroReview{ModuleDirectory: ".", PackageDirectory: "other"}
	if _, err := InputDigest(root, policy, strings.NewReader(`{}`), review); !errors.Is(err, ErrInvalid) {
		t.Fatalf("InputDigest() error = %v", err)
	}
}

func TestInputPolicyValidationAndCanonicalization(t *testing.T) {
	valid := InputPolicy{
		ModuleDirectory: ".", PackageDirectory: "nested", ModulePath: "example", GoVersion: "1.26.6",
		TestTags: []string{"z", "a"}, BuildTags: []string{"b", "a"}, RequiredServices: []string{"valkey", "postgresql"},
		ServiceIdentities: map[string]string{"valkey": "valkey:9", "postgresql": "postgres:18"},
		OwnedModules:      []OwnedModule{{ModulePath: "example/z", Directory: "z"}, {ModulePath: "example/a", Directory: "a"}},
	}
	if err := valid.validate(); err != nil {
		t.Fatal(err)
	}
	canonical := valid.canonical()
	if strings.Join(canonical.TestTags, ",") != "a,z" || canonical.OwnedModules[0].ModulePath != "example/a" {
		t.Fatalf("canonical() = %#v", canonical)
	}
	for _, policy := range []InputPolicy{
		{ModuleDirectory: ".", PackageDirectory: ".", ModulePath: "example", GoVersion: "1.26.6", OwnedModules: []OwnedModule{{}}},
		{ModuleDirectory: ".", PackageDirectory: ".", ModulePath: "example", GoVersion: "1.26.6", OwnedModules: []OwnedModule{{ModulePath: "example", Directory: "nested"}}},
		{ModuleDirectory: ".", PackageDirectory: ".", ModulePath: "example", GoVersion: "1.26.6", RequiredServices: []string{"postgresql"}},
		{ModuleDirectory: ".", PackageDirectory: ".", ModulePath: "example", GoVersion: "1.26.6", ServiceIdentities: map[string]string{"postgresql": "postgres:18"}},
	} {
		if err := policy.validate(); !errors.Is(err, ErrInvalid) {
			t.Fatalf("validate() error = %v", err)
		}
	}
}

func TestParseListingBoundsAndFailures(t *testing.T) {
	if _, err := parseListing(failingReader{}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("parseListing(read) error = %v", err)
	}
	if _, err := parseListing(strings.NewReader(strings.Repeat("x", maximumListSize+1))); !errors.Is(err, ErrInvalid) {
		t.Fatalf("parseListing(size) error = %v", err)
	}
	if _, err := parseListing(strings.NewReader("")); !errors.Is(err, ErrInvalid) {
		t.Fatalf("parseListing(empty) error = %v", err)
	}
}

func TestProductionFilesIncludesEveryGoListClass(t *testing.T) {
	pkg := listedPackage{
		GoFiles: []string{"go"}, CgoFiles: []string{"cgo"}, CFiles: []string{"c"}, CXXFiles: []string{"cxx"},
		MFiles: []string{"m"}, HFiles: []string{"h"}, FFiles: []string{"f"}, SFiles: []string{"s"},
		SwigFiles: []string{"swig"}, SwigCXXFiles: []string{"swigcxx"}, SysoFiles: []string{"syso"}, EmbedFiles: []string{"embed"},
	}
	if got := productionFiles(pkg); len(got) != 12 {
		t.Fatalf("productionFiles() count = %d", len(got))
	}
}

func TestOwnedRootResolution(t *testing.T) {
	root := t.TempDir()
	owned := []OwnedModule{{ModulePath: "example", Directory: "."}, {ModulePath: "example/nested", Directory: "nested"}}
	if _, ok := resolveOwnedRoot(root, listedPackage{}, owned); ok {
		t.Fatal("resolveOwnedRoot() accepted package without module")
	}
	pkg := listedPackage{Dir: filepath.Join(root, "nested"), Module: &listedModule{Path: "example/nested"}}
	got, ok := resolveOwnedRoot(root, pkg, owned)
	if !ok || got.ModulePath != "example/nested" {
		t.Fatalf("resolveOwnedRoot() = %#v, %t", got, ok)
	}
	pkg.Dir = filepath.Join(root, "outside")
	if _, ok := resolveOwnedRoot(root, pkg, owned); ok {
		t.Fatal("resolveOwnedRoot() accepted directory outside module")
	}
}

func TestAddInputContentAndDataFailures(t *testing.T) {
	content := map[string][]byte{"existing": nil}
	total := 0
	if err := addContent(content, "existing", []byte("ignored"), &total); err != nil || total != 0 {
		t.Fatalf("addContent(existing) = %d, %v", total, err)
	}
	total = maximumInputTotal
	if err := addContent(content, "new", []byte("x"), &total); !errors.Is(err, ErrInvalid) {
		t.Fatalf("addContent(bound) error = %v", err)
	}
	root := t.TempDir()
	owned := OwnedModule{ModulePath: "example", Directory: "."}
	if err := addDataDirectory(root, owned, string([]byte{'x', 0}), content, &total); err == nil {
		t.Fatal("addDataDirectory() accepted invalid path")
	}
	file := filepath.Join(root, "testdata")
	writeMutationInput(t, root, "testdata", "not a directory")
	if err := addDataDirectory(root, owned, file, content, &total); !errors.Is(err, ErrInvalid) {
		t.Fatalf("addDataDirectory(file) error = %v", err)
	}
}

func TestVisitDataEntryBranches(t *testing.T) {
	root := t.TempDir()
	owned := OwnedModule{ModulePath: "example", Directory: "."}
	content := make(map[string][]byte)
	total := 0
	walkError := errors.New("walk failed")
	if err := visitDataEntry(root, owned, root, fakeDirEntry{}, walkError, content, &total); !errors.Is(err, walkError) {
		t.Fatalf("visitDataEntry(walk error) = %v", err)
	}
	if err := visitDataEntry(root, owned, root, fakeDirEntry{mode: os.ModeSymlink}, nil, content, &total); !errors.Is(err, ErrInvalid) {
		t.Fatalf("visitDataEntry(symlink) = %v", err)
	}
	if err := visitDataEntry(root, owned, root, fakeDirEntry{directory: true, mode: os.ModeDir}, nil, content, &total); err != nil {
		t.Fatalf("visitDataEntry(directory) = %v", err)
	}
	if err := visitDataEntry(root, owned, root, fakeDirEntry{mode: os.ModeNamedPipe}, nil, content, &total); !errors.Is(err, ErrInvalid) {
		t.Fatalf("visitDataEntry(nonregular) = %v", err)
	}
}

func writeMutationInput(t *testing.T, root, relative, content string) {
	t.Helper()
	path := filepath.Join(root, relative)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

type fakeDirEntry struct {
	directory bool
	mode      fs.FileMode
}

func (entry fakeDirEntry) Name() string               { return "entry" }
func (entry fakeDirEntry) IsDir() bool                { return entry.directory }
func (entry fakeDirEntry) Type() fs.FileMode          { return entry.mode }
func (entry fakeDirEntry) Info() (fs.FileInfo, error) { return fakeFileInfo{mode: entry.mode}, nil }

type fakeFileInfo struct{ mode fs.FileMode }

func (info fakeFileInfo) Name() string       { return "entry" }
func (info fakeFileInfo) Size() int64        { return 0 }
func (info fakeFileInfo) Mode() fs.FileMode  { return info.mode }
func (info fakeFileInfo) ModTime() time.Time { return time.Time{} }
func (info fakeFileInfo) IsDir() bool        { return info.mode.IsDir() }
func (info fakeFileInfo) Sys() any           { return nil }
