package mutation_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/faustbrian/go-library-tools/internal/mutation"
)

func TestInputDigestTracksOnlyObservedContentAndSemantics(t *testing.T) {
	root := t.TempDir()
	writeInput(t, root, "target.go", "package example\nfunc Value() int { return dependency() }\n")
	writeInput(t, root, "target_test.go", "package example\n")
	writeInput(t, root, "dep/dep.go", "package dep\n")
	writeInput(t, root, "unrelated/unrelated.go", "package unrelated\n")
	writeInput(t, root, "testdata/input.txt", "fixture")
	policy := mutation.InputPolicy{
		ModuleDirectory: ".", PackageDirectory: ".",
		ModulePath: "github.com/faustbrian/go-example", GoVersion: "1.26.6",
		TestTags: []string{"integration"}, RequiredServices: []string{"postgresql"},
		ServiceIdentities: map[string]string{"postgresql": "postgres:18.4-alpine"},
	}
	listing := inputListing(root)
	first, err := mutation.InputDigest(root, policy, strings.NewReader(listing), nil)
	if err != nil {
		t.Fatalf("InputDigest() error = %v", err)
	}
	if !strings.HasPrefix(first, "sha256:") {
		t.Fatalf("InputDigest() = %q", first)
	}

	writeInput(t, root, "unrelated/unrelated.go", "package unrelated\nvar Changed = true\n")
	unchanged, err := mutation.InputDigest(root, policy, strings.NewReader(listing), nil)
	if err != nil || unchanged != first {
		t.Fatalf("unrelated change digest = %q, %v; want %q", unchanged, err, first)
	}
	writeInput(t, root, "dep/dep.go", "package dep\nvar Changed = true\n")
	changed, err := mutation.InputDigest(root, policy, strings.NewReader(listing), nil)
	if err != nil {
		t.Fatal(err)
	}
	if changed == first {
		t.Fatal("dependency source change did not invalidate mutation input")
	}
}

func TestInputDigestIncludesExactZeroReview(t *testing.T) {
	root := t.TempDir()
	writeInput(t, root, "target.go", "package example\n")
	policy := mutation.InputPolicy{ModuleDirectory: ".", PackageDirectory: ".", ModulePath: "example", GoVersion: "1.26.6"}
	listing := `{"Dir":"` + root + `","ImportPath":"example","GoFiles":["target.go"],"Module":{"Path":"example","Main":true,"GoVersion":"1.26.6"}}`
	review := &mutation.ZeroReview{
		ModuleDirectory: ".", PackageDirectory: ".", SourceDigest: strings.Repeat("a", 64),
		GremlinsVersion: mutation.GremlinsVersion, GremlinsVerifierSHA256: mutation.LegacyVerifierDigest(),
		Reason: "The complete verifier finds no viable mutations in this declaration-only package.",
	}
	withReview, err := mutation.InputDigest(root, policy, strings.NewReader(listing), review)
	if err != nil {
		t.Fatal(err)
	}
	review.Reason += " Reviewed again."
	changed, err := mutation.InputDigest(root, policy, strings.NewReader(listing), review)
	if err != nil {
		t.Fatal(err)
	}
	if withReview == changed {
		t.Fatal("zero-mutant review change did not invalidate input")
	}
}

func inputListing(root string) string {
	return strings.Join([]string{
		`{"Dir":"` + root + `","ImportPath":"github.com/faustbrian/go-example","GoFiles":["target.go"],"TestGoFiles":["target_test.go"],"Module":{"Path":"github.com/faustbrian/go-example","Main":true,"GoVersion":"1.26.6"}}`,
		`{"Dir":"` + filepath.Join(root, "dep") + `","ImportPath":"github.com/faustbrian/go-example/dep","GoFiles":["dep.go"],"Module":{"Path":"github.com/faustbrian/go-example","Main":true,"GoVersion":"1.26.6"}}`,
		`{"Dir":"/module/cache/dependency","ImportPath":"example.com/dependency","GoFiles":["dependency.go"],"Module":{"Path":"example.com/dependency","Version":"v1.2.3","Sum":"h1:digest","GoVersion":"1.24"}}`,
	}, "\n")
}

func writeInput(t *testing.T, root, relative, content string) {
	t.Helper()
	path := filepath.Join(root, relative)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}
