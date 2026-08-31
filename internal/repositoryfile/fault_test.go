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
		{"open", fakeFileSystem{info: info, openErr: failure}, "open file"},
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
}

type fakeFileSystem struct {
	info    os.FileInfo
	opened  file
	openErr error
}

func (fake fakeFileSystem) Lstat(string) (os.FileInfo, error) {
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
