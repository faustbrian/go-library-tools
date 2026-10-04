package mutation

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
)

func TestSourceDigestAdmitsOrdinaryAncestorMetadata(t *testing.T) {
	files := ordinaryAncestorSource(t)
	got, err := sourceDigest(files, files.root, "module", "ordinary")
	if err != nil || got != "253d68920c6ab77fd6b980851370b12b2b5aa45249e958db69caaa6052720632" {
		t.Fatal("ordinary nested source must retain its literal content identity")
	}
}

func TestSourceDigestRefusesAncestorMetadataBeforeContent(t *testing.T) {
	for _, test := range []struct {
		name, module, pkg string
		mode              os.FileMode
	}{
		{"module symlink", "module", "ordinary", os.ModeSymlink},
		{"package ancestor symlink", ".", "module/ordinary", os.ModeSymlink},
		{"module nondirectory", "module", "ordinary", 0},
		{"package ancestor nondirectory", ".", "module/ordinary", 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			files := ordinaryAncestorSource(t)
			// The in-memory filesystem models ordinary metadata and bytes only;
			// no real filesystem link or permission operation is performed.
			files.content["module"] = &fstest.MapFile{Mode: test.mode, Data: []byte("backing")}
			files.content["backing/ordinary/source.go"] = files.content["module/ordinary/source.go"]
			got, err := sourceDigest(files, files.root, test.module, test.pkg)
			if got != "" || !errors.Is(err, ErrInvalid) {
				t.Fatal("inadmissible ancestor must return no digest and ErrInvalid")
			}
			if err.Error() != "invalid mutation evidence: mutation source path is not a real directory" {
				t.Fatal("ancestor refusal must retain the existing directory category")
			}
			if files.enumerated != 0 || files.read != 0 {
				t.Fatal("ancestor admission must precede enumeration and content reads")
			}
		})
	}
}

func TestSourceDigestAncestorMetadataFailurePreservesCause(t *testing.T) {
	files := ordinaryAncestorSource(t)
	failure := errors.New("ordinary metadata failure")
	files.metadataFailure = failure
	got, err := sourceDigest(files, files.root, "module", "ordinary")
	if got != "" || !errors.Is(err, failure) || errors.Is(err, ErrInvalid) {
		t.Fatal("metadata refusal must return no digest and preserve the actual cause")
	}
	if err.Error() != "inspect mutation source directory: ordinary metadata failure" {
		t.Fatal("metadata refusal must retain the existing error context")
	}
	if files.enumerated != 0 || files.read != 0 {
		t.Fatal("failed ancestor metadata must precede enumeration and content reads")
	}
}

type ancestorSourceFiles struct {
	root            string
	content         fstest.MapFS
	metadataFailure error
	enumerated      int
	read            int
}

func ordinaryAncestorSource(t *testing.T) *ancestorSourceFiles {
	t.Helper()
	root, err := filepath.Abs("ordinary-source-root")
	if err != nil {
		t.Fatal("derive ordinary absolute metadata root")
	}
	return &ancestorSourceFiles{root: root, content: fstest.MapFS{
		"module":                    {Mode: os.ModeDir},
		"module/ordinary":           {Mode: os.ModeDir},
		"module/ordinary/source.go": {Data: []byte("package ordinary\n")},
	}}
}

func (files *ancestorSourceFiles) Lstat(name string) (os.FileInfo, error) {
	relative, err := filepath.Rel(files.root, name)
	if err != nil {
		return nil, err
	}
	if relative == "module" && files.metadataFailure != nil {
		return nil, files.metadataFailure
	}
	return files.content.Lstat(filepath.ToSlash(relative))
}

func (files *ancestorSourceFiles) ReadDir(name string, _ int) ([]os.DirEntry, error) {
	files.enumerated++
	relative, err := filepath.Rel(files.root, name)
	if err != nil {
		return nil, err
	}
	return files.content.ReadDir(filepath.ToSlash(relative))
}

func (files *ancestorSourceFiles) ReadFile(name string, _, _ int64) ([]byte, error) {
	files.read++
	relative, err := filepath.Rel(files.root, name)
	if err != nil {
		return nil, err
	}
	return files.content.ReadFile(filepath.ToSlash(relative))
}

func (*ancestorSourceFiles) Rel(base, target string) (string, error) {
	return filepath.Rel(base, target)
}
