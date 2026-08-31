package mutation

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestInputDigestRejectsInvalidInputs(t *testing.T) {
	root := t.TempDir()
	missingRoot := filepath.Join(t.TempDir(), "missing")
	writeMutationInput(t, root, "other.go", "package other\n")
	valid := InputPolicy{ModuleDirectory: ".", PackageDirectory: ".", ModulePath: "example", GoVersion: "1.26.6"}
	listing := func(files string) string {
		return `{"Dir":"` + root + `","ImportPath":"example","GoFiles":` + files + `,"Module":{"Path":"example","Main":true}}`
	}
	tests := map[string]struct {
		root    string
		policy  InputPolicy
		listing string
	}{
		"relative root":    {root: ".", policy: valid, listing: listing(`[]`)},
		"missing root":     {root: missingRoot, policy: valid, listing: listing(`[]`)},
		"bad policy":       {root: root, policy: InputPolicy{}, listing: listing(`[]`)},
		"malformed list":   {root: root, policy: valid, listing: "{"},
		"missing target":   {root: root, policy: valid, listing: `{"Dir":"/external","ImportPath":"external"}`},
		"local non-target": {root: root, policy: valid, listing: `{"Dir":"` + root + `","ImportPath":"example/other","GoFiles":["other.go"],"Module":{"Path":"example","Main":true}}`},
		"empty target":     {root: root, policy: valid, listing: listing(`[]`)},
		"missing file":     {root: root, policy: valid, listing: listing(`["missing.go"]`)},
		"escaping file":    {root: root, policy: valid, listing: listing(`["../outside.go"]`)},
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

func TestInputDigestIgnoresNestedDataSymlinksWithoutFollowingThem(t *testing.T) {
	root := t.TempDir()
	writeMutationInput(t, root, "target.go", "package example\n")
	writeMutationInput(t, root, "testdata/fixture.json", "{}\n")
	policy := InputPolicy{ModuleDirectory: ".", PackageDirectory: ".", ModulePath: "example", GoVersion: "1.26.6"}
	listing := `{"Dir":"` + root + `","ImportPath":"example","GoFiles":["target.go"],"Module":{"Path":"example","Main":true}}`

	want, err := InputDigest(root, policy, strings.NewReader(listing), nil)
	if err != nil {
		t.Fatalf("InputDigest(without symlink) error = %v", err)
	}
	outside := t.TempDir()
	writeMutationInput(t, outside, "external.json", "must not affect the digest\n")
	if err := os.Symlink(outside, filepath.Join(root, "testdata", "latest")); err != nil {
		t.Fatal(err)
	}

	got, err := InputDigest(root, policy, strings.NewReader(listing), nil)
	if err != nil {
		t.Fatalf("InputDigest(with nested symlink) error = %v", err)
	}
	if got != want {
		t.Fatalf("InputDigest(with nested symlink) = %s, want %s", got, want)
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

func TestInputDigestResolvesRepositoryRootSymlink(t *testing.T) {
	root := t.TempDir()
	writeMutationInput(t, root, "target.go", "package example\n")
	absoluteAlias := filepath.Join(t.TempDir(), "repository")
	if err := os.Symlink(root, absoluteAlias); err != nil {
		t.Fatal(err)
	}
	relativeParent := t.TempDir()
	relativeTarget, err := filepath.Rel(relativeParent, root)
	if err != nil {
		t.Fatal(err)
	}
	relativeAlias := filepath.Join(relativeParent, "repository")
	if err := os.Symlink(relativeTarget, relativeAlias); err != nil {
		t.Fatal(err)
	}
	policy := InputPolicy{ModuleDirectory: ".", PackageDirectory: ".", ModulePath: "example", GoVersion: "1.27.0"}
	canonicalListing := `{"Dir":"` + root + `","ImportPath":"example","GoFiles":["target.go"],"Module":{"Path":"example","Main":true}}`
	want, err := InputDigest(root, policy, strings.NewReader(canonicalListing), nil)
	if err != nil {
		t.Fatal(err)
	}
	for aliasName, alias := range map[string]string{"absolute target": absoluteAlias, "relative target": relativeAlias} {
		t.Run(aliasName, func(t *testing.T) {
			for listingName, directory := range map[string]string{"canonical listing": root, "alias listing": alias} {
				t.Run(listingName, func(t *testing.T) {
					listing := `{"Dir":"` + directory + `","ImportPath":"example","GoFiles":["target.go"],"Module":{"Path":"example","Main":true}}`
					got, digestErr := InputDigest(alias, policy, strings.NewReader(listing), nil)
					if digestErr != nil {
						t.Fatalf("InputDigest(symlink root) error = %v", digestErr)
					}
					if got != want {
						t.Fatalf("InputDigest(symlink root) = %s, want %s", got, want)
					}
				})
			}
		})
	}
}

func TestInputDigestRejectsListedFileSymlinks(t *testing.T) {
	root := t.TempDir()
	writeMutationInput(t, root, "target.go", "package example\n")
	if err := os.Symlink(filepath.Join(root, "target.go"), filepath.Join(root, "alias.go")); err != nil {
		t.Fatal(err)
	}
	policy := InputPolicy{ModuleDirectory: ".", PackageDirectory: ".", ModulePath: "example", GoVersion: "1.27.0"}
	for name, files := range map[string]string{
		"symlink only":       `["alias.go"]`,
		"symlink and target": `["target.go","alias.go"]`,
	} {
		t.Run(name, func(t *testing.T) {
			listing := `{"Dir":"` + root + `","ImportPath":"example","GoFiles":` + files + `,"Module":{"Path":"example","Main":true}}`
			if _, err := InputDigest(root, policy, strings.NewReader(listing), nil); err == nil {
				t.Fatal("InputDigest() accepted a listed symlink")
			}
		})
	}
}

func TestInputDigestRejectsListedPackageDirectorySymlinks(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "target")
	writeMutationInput(t, target, "target.go", "package example\n")
	other := filepath.Join(root, "other")
	writeMutationInput(t, other, "other.go", "package other\n")
	alias := filepath.Join(root, "alias")
	if err := os.Symlink(target, alias); err != nil {
		t.Fatal(err)
	}
	policy := InputPolicy{ModuleDirectory: ".", PackageDirectory: ".", ModulePath: "example", GoVersion: "1.27.0"}
	for name, listing := range map[string]string{
		"relative file":           `{"Dir":"` + alias + `","ImportPath":"example","GoFiles":["target.go"],"Module":{"Path":"example","Main":true}}`,
		"absolute canonical file": `{"Dir":"` + alias + `","ImportPath":"example","GoFiles":["` + filepath.Join(target, "target.go") + `"],"Module":{"Path":"example","Main":true}}`,
		"empty target": `{"Dir":"` + alias + `","ImportPath":"example","GoFiles":[],"Module":{"Path":"example","Main":true}}` + "\n" +
			`{"Dir":"` + other + `","ImportPath":"example/other","GoFiles":["other.go"],"Module":{"Path":"example","Main":true}}`,
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := InputDigest(root, policy, strings.NewReader(listing), nil); err == nil {
				t.Fatal("InputDigest() accepted a listed package directory symlink")
			}
		})
	}
}

func TestLegacyInputDigestV1RetainsUnobservedOwnedModules(t *testing.T) {
	root := t.TempDir()
	writeMutationInput(t, root, "target.go", "package example\n")
	listing := `{"Dir":"` + root + `","ImportPath":"example","GoFiles":["target.go"],"Module":{"Path":"example","Main":true}}`
	policy := InputPolicy{
		ModuleDirectory: ".", PackageDirectory: ".", ModulePath: "example", GoVersion: "1.27.0",
		OwnedModules: []OwnedModule{{ModulePath: "example/unobserved", Directory: "unobserved"}},
	}
	current, err := InputDigest(root, policy, strings.NewReader(listing), nil)
	if err != nil {
		t.Fatal(err)
	}
	legacy, err := legacyInputDigestV1(root, policy, strings.NewReader(listing), nil)
	if err != nil {
		t.Fatal(err)
	}
	if current == legacy {
		t.Fatal("legacy and current input identities unexpectedly match")
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

func TestInputDigestContinuesPastIrrelevantPackages(t *testing.T) {
	root := t.TempDir()
	writeMutationInput(t, root, "target.go", "package example\n")
	policy := InputPolicy{ModuleDirectory: ".", PackageDirectory: ".", ModulePath: "example", GoVersion: "1.27.0"}
	listing := strings.Join([]string{
		`{"Dir":"/external","ImportPath":"external","GoFiles":["external.go"],"Module":{"Path":"external"}}`,
		`{"Dir":"` + root + `","ImportPath":"example.test","GoFiles":["generated"],"Module":{"Path":"example","Main":true}}`,
		`{"Dir":"` + root + `","ImportPath":"example/empty","GoFiles":[],"Module":{"Path":"example","Main":true}}`,
		`{"Dir":"` + root + `","ImportPath":"example","GoFiles":["target.go"],"Module":{"Path":"example","Main":true}}`,
	}, "\n")
	if _, err := InputDigest(root, policy, strings.NewReader(listing), nil); err != nil {
		t.Fatalf("InputDigest() error = %v", err)
	}
}

func TestInputDigestIncludesExternalObserverTests(t *testing.T) {
	root := t.TempDir()
	writeMutationInput(t, root, "target.go", "package example\n")
	writeMutationInput(t, root, "external_test.go", "package example_test\n")
	policy := InputPolicy{ModuleDirectory: ".", PackageDirectory: ".", ModulePath: "example", GoVersion: "1.27.0"}
	listing := strings.Join([]string{
		`{"Dir":"` + root + `","ImportPath":"example","GoFiles":["target.go"],"Module":{"Path":"example","Main":true}}`,
		`{"Dir":"` + root + `","ImportPath":"example_test [example.test]","ForTest":"example","GoFiles":[],"XTestGoFiles":["external_test.go"],"Module":{"Path":"example","Main":true}}`,
	}, "\n")
	withTest, err := InputDigest(root, policy, strings.NewReader(listing), nil)
	if err != nil {
		t.Fatal(err)
	}
	writeMutationInput(t, root, "external_test.go", "package example_test\n// changed\n")
	changed, err := InputDigest(root, policy, strings.NewReader(listing), nil)
	if err != nil || withTest == changed {
		t.Fatalf("observer digest did not change: %v", err)
	}
}

func TestInputDigestIncludesTargetTests(t *testing.T) {
	root := t.TempDir()
	writeMutationInput(t, root, "target.go", "package example\n")
	writeMutationInput(t, root, "target_test.go", "package example\n")
	policy := InputPolicy{ModuleDirectory: ".", PackageDirectory: ".", ModulePath: "example", GoVersion: "1.27.0"}
	listing := `{"Dir":"` + root + `","ImportPath":"example","GoFiles":["target.go"],"TestGoFiles":["target_test.go"],"Module":{"Path":"example","Main":true}}`
	before, err := InputDigest(root, policy, strings.NewReader(listing), nil)
	if err != nil {
		t.Fatal(err)
	}
	writeMutationInput(t, root, "target_test.go", "package example\n// changed\n")
	after, err := InputDigest(root, policy, strings.NewReader(listing), nil)
	if err != nil || before == after {
		t.Fatalf("target test digest unchanged: %v", err)
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

func TestInputDigestRejectsEveryMismatchedZeroReviewField(t *testing.T) {
	root := t.TempDir()
	valid := ZeroReview{ModuleDirectory: ".", PackageDirectory: ".", SourceDigest: strings.Repeat("a", 64), GremlinsVersion: GremlinsVersion, GremlinsVerifierSHA256: LegacyVerifierDigest(), Reason: strings.Repeat("r", 40)}
	policy := InputPolicy{ModuleDirectory: ".", PackageDirectory: ".", ModulePath: "example", GoVersion: "1.26.6"}
	for name, mutate := range map[string]func(*ZeroReview){
		"invalid review": func(value *ZeroReview) { value.Reason = "short" },
		"module":         func(value *ZeroReview) { value.ModuleDirectory = "other" },
		"package":        func(value *ZeroReview) { value.PackageDirectory = "other" },
		"version":        func(value *ZeroReview) { value.GremlinsVersion = "v9.9.9" },
		"verifier":       func(value *ZeroReview) { value.GremlinsVerifierSHA256 = strings.Repeat("b", 64) },
	} {
		t.Run(name, func(t *testing.T) {
			review := valid
			mutate(&review)
			if _, err := InputDigest(root, policy, strings.NewReader(`{}`), &review); !errors.Is(err, ErrInvalid) {
				t.Fatalf("InputDigest() error = %v", err)
			}
		})
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
		{ModuleDirectory: ".", PackageDirectory: ".", ModulePath: "example", GoVersion: "1.26.6", OwnedModules: []OwnedModule{{ModulePath: "example/other", Directory: "../outside"}}},
		{ModuleDirectory: ".", PackageDirectory: ".", ModulePath: "example", GoVersion: "1.26.6", OwnedModules: []OwnedModule{{ModulePath: "example", Directory: "nested"}}},
		{ModuleDirectory: ".", PackageDirectory: ".", ModulePath: "example", GoVersion: "1.26.6", RequiredServices: []string{"postgresql"}},
		{ModuleDirectory: ".", PackageDirectory: ".", ModulePath: "example", GoVersion: "1.26.6", ServiceIdentities: map[string]string{"postgresql": "postgres:18"}},
	} {
		if err := policy.validate(); !errors.Is(err, ErrInvalid) {
			t.Fatalf("validate() error = %v", err)
		}
	}
}

func TestInputPolicyRejectsEachMalformedField(t *testing.T) {
	valid := InputPolicy{ModuleDirectory: ".", PackageDirectory: ".", ModulePath: "example", GoVersion: "1.26.6"}
	for name, mutate := range map[string]func(*InputPolicy){
		"module path":       func(value *InputPolicy) { value.ModulePath = "" },
		"go version":        func(value *InputPolicy) { value.GoVersion = "" },
		"module directory":  func(value *InputPolicy) { value.ModuleDirectory = "../x" },
		"package directory": func(value *InputPolicy) { value.PackageDirectory = "../x" },
	} {
		t.Run(name, func(t *testing.T) {
			policy := valid
			mutate(&policy)
			if err := policy.validate(); !errors.Is(err, ErrInvalid) {
				t.Fatalf("validate() error = %v", err)
			}
		})
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

func TestParseListingAcceptsExactMaximumSize(t *testing.T) {
	prefix := `{"Dir":"/tmp"`
	data := prefix + strings.Repeat(" ", maximumListSize-len(prefix)-1) + `}`
	packages, err := parseListing(strings.NewReader(data))
	if err != nil || len(packages) != 1 {
		t.Fatalf("parseListing(exact maximum) = %d, %v", len(packages), err)
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

func TestIsWithinRejectsInvalidAndParentPaths(t *testing.T) {
	root := t.TempDir()
	if isWithin(string([]byte{'x', 0}), root) {
		t.Fatal("isWithin accepted an invalid root")
	}
	for _, candidate := range []string{string([]byte{'x', 0}), filepath.Dir(root), filepath.Join(filepath.Dir(root), "sibling", "file")} {
		if isWithin(root, candidate) {
			t.Fatalf("isWithin(%q) = true", candidate)
		}
	}
	if !isWithin(root, root) || !isWithin(root, filepath.Join(root, "child")) {
		t.Fatal("isWithin rejected owned path")
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
	full := make(map[string][]byte, maximumInputFiles)
	for index := range maximumInputFiles {
		full[strconv.Itoa(index)] = nil
	}
	total = 0
	if err := addContent(full, "overflow", nil, &total); !errors.Is(err, ErrInvalid) {
		t.Fatalf("addContent(file count) = %v", err)
	}
	root := t.TempDir()
	owned := OwnedModule{ModulePath: "example", Directory: "."}
	entries := 0
	if err := addDataDirectory(root, owned, string([]byte{'x', 0}), &entries, maximumInputEntries, content, &total); err == nil {
		t.Fatal("addDataDirectory() accepted invalid path")
	}
	file := filepath.Join(root, "testdata")
	writeMutationInput(t, root, "testdata", "not a directory")
	if err := addDataDirectory(root, owned, file, &entries, maximumInputEntries, content, &total); !errors.Is(err, ErrInvalid) {
		t.Fatalf("addDataDirectory(file) error = %v", err)
	}
}

func TestAddDataDirectorySharesTraversalBoundAcrossRoots(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "corpus"), 0o700); err != nil {
		t.Fatal(err)
	}
	writeMutationInput(t, root, "fixtures/fixture.json", "{}\n")
	owned := OwnedModule{ModulePath: "example", Directory: "."}
	content := make(map[string][]byte)
	entries, total := 0, 0
	if err := addDataDirectory(root, owned, filepath.Join(root, "corpus"), &entries, 2, content, &total); err != nil {
		t.Fatalf("addDataDirectory(first root) error = %v", err)
	}
	if err := addDataDirectory(root, owned, filepath.Join(root, "fixtures"), &entries, 2, content, &total); !errors.Is(err, ErrInvalid) {
		t.Fatalf("addDataDirectory(aggregate bound) error = %v", err)
	}
}

func TestAddContentAcceptsExactBounds(t *testing.T) {
	content := make(map[string][]byte, maximumInputFiles)
	total := 0
	for index := range maximumInputFiles {
		data := []byte(nil)
		if index == maximumInputFiles-1 {
			data = make([]byte, maximumInputTotal)
		}
		if err := addContent(content, strconv.Itoa(index), data, &total); err != nil {
			t.Fatalf("addContent(%d) error = %v", index, err)
		}
	}
	if total != maximumInputTotal || len(content) != maximumInputFiles {
		t.Fatalf("bounds = files %d, bytes %d", len(content), total)
	}
}

func TestVisitDataEntryBranches(t *testing.T) {
	root := t.TempDir()
	owned := OwnedModule{ModulePath: "example", Directory: "."}
	content := make(map[string][]byte)
	total := 0
	entries := 0
	walkError := errors.New("walk failed")
	if err := visitDataEntry(root, owned, root, fakeDirEntry{}, walkError, &entries, maximumInputEntries, content, &total); !errors.Is(err, walkError) {
		t.Fatalf("visitDataEntry(walk error) = %v", err)
	}
	if err := visitDataEntry(root, owned, root, fakeDirEntry{mode: os.ModeSymlink}, nil, &entries, maximumInputEntries, content, &total); err != nil {
		t.Fatalf("visitDataEntry(symlink) = %v", err)
	}
	if err := visitDataEntry(root, owned, root, fakeDirEntry{directory: true, mode: os.ModeDir}, nil, &entries, maximumInputEntries, content, &total); err != nil {
		t.Fatalf("visitDataEntry(directory) = %v", err)
	}
	if err := visitDataEntry(root, owned, root, fakeDirEntry{mode: os.ModeNamedPipe}, nil, &entries, maximumInputEntries, content, &total); !errors.Is(err, ErrInvalid) {
		t.Fatalf("visitDataEntry(nonregular) = %v", err)
	}
	entries = maximumInputEntries - 1
	if err := visitDataEntry(root, owned, root, fakeDirEntry{mode: os.ModeSymlink}, nil, &entries, maximumInputEntries, content, &total); err != nil {
		t.Fatalf("visitDataEntry(exact entry bound) = %v", err)
	}
	if err := visitDataEntry(root, owned, root, fakeDirEntry{mode: os.ModeSymlink}, nil, &entries, maximumInputEntries, content, &total); !errors.Is(err, ErrInvalid) {
		t.Fatalf("visitDataEntry(entry bound) = %v", err)
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
