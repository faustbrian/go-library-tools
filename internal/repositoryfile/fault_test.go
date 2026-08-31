package repositoryfile

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestReadReportsFileSystemRacesAndFailures(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "file")
	if err := os.WriteFile(path, []byte("content"), 0o600); err != nil {
		t.Fatal(err)
	}
	info, err := os.Lstat(path)
	if err != nil {
		t.Fatal(err)
	}
	failure := errors.New("injected failure")

	tests := []struct {
		name  string
		files fileSystem
		want  string
	}{
		{"nil info", fakeFileSystem{}, "inspect file"},
		{"nil open file", fakeFileSystem{info: info}, "open file returned no file"},
		{"open", fakeFileSystem{info: info, openErr: failure}, "open file"},
		{"nil stat info", fakeFileSystem{info: info, opened: &fakeFile{}}, "inspect open file returned no metadata"},
		{"stat", fakeFileSystem{info: info, opened: &fakeFile{statErr: failure}}, "inspect open file"},
		{"changed", fakeFileSystem{info: info, opened: &fakeFile{info: fakeInfo{name: "other"}}}, "changed while opening"},
		{"nonregular", fakeFileSystem{info: info, opened: &fakeFile{info: fakeInfo{name: "other", mode: os.ModeDir}}}, "changed while opening"},
		{"read", fakeFileSystem{info: info, opened: &fakeFile{info: info, readErr: failure}}, "read file"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := read(root, "file", 100, test.files)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("read() error = %v", err)
			}
		})
	}
	for name, files := range map[string]fileSystem{
		"parent failure":  fakeFileSystem{lstat: func(string) (os.FileInfo, error) { return nil, failure }},
		"nil parent info": fakeFileSystem{lstat: func(string) (os.FileInfo, error) { return nil, nil }},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := read(root, "nested/file", 100, files); err == nil {
				t.Fatal("read() error = nil")
			}
		})
	}
}

type fakeFileSystem struct {
	lstat   func(string) (os.FileInfo, error)
	info    os.FileInfo
	opened  file
	openErr error
}

func (fake fakeFileSystem) Lstat(path string) (os.FileInfo, error) {
	if fake.lstat != nil {
		return fake.lstat(path)
	}
	return fake.info, nil
}

func (fake fakeFileSystem) Open(string) (file, error) {
	return fake.opened, fake.openErr
}

type fakeFile struct {
	info    os.FileInfo
	statErr error
	readErr error
}

func (fake *fakeFile) Read([]byte) (int, error) {
	return 0, fake.readErr
}

func (fake *fakeFile) Stat() (os.FileInfo, error) {
	return fake.info, fake.statErr
}

func (*fakeFile) Close() error {
	return nil
}

type fakeInfo struct {
	name string
	mode os.FileMode
}

func (fake fakeInfo) Name() string { return fake.name }
func (fakeInfo) Size() int64       { return 0 }
func (fake fakeInfo) Mode() os.FileMode {
	if fake.mode != 0 {
		return fake.mode
	}
	return 0o600
}
func (fakeInfo) ModTime() time.Time { return time.Time{} }
func (fakeInfo) IsDir() bool        { return false }
func (fakeInfo) Sys() any           { return nil }

var _ io.Reader = (*fakeFile)(nil)
