package evidence

import (
	"errors"
	"os"
	"strings"
	"testing"
	"time"
)

func TestStoreReportsAtomicFilesystemFailures(t *testing.T) {
	failure := errors.New("injected failure")
	tests := []struct {
		name  string
		files *fakeStoreFiles
		want  string
	}{
		{"mkdir", &fakeStoreFiles{mkdirErr: failure}, "create evidence directory"},
		{"create", &fakeStoreFiles{createErr: failure}, "create temporary evidence"},
		{"write", &fakeStoreFiles{file: &fakeDurableFile{writeErr: failure}}, "write evidence"},
		{"sync", &fakeStoreFiles{file: &fakeDurableFile{syncErr: failure}}, "write evidence"},
		{"close", &fakeStoreFiles{file: &fakeDurableFile{closeErr: failure}}, "write evidence"},
		{"publish", &fakeStoreFiles{file: &fakeDurableFile{}, linkErr: failure}, "publish evidence"},
		{"read", &fakeStoreFiles{file: &fakeDurableFile{}, linkErr: os.ErrExist, readErr: failure}, "read existing evidence"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, _, err := store(test.files, t.TempDir(), validInternalRecord())
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("store() error = %v", err)
			}
		})
	}
	if _, _, err := store(&fakeStoreFiles{}, t.TempDir(), Record{}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("store() validation error = %v", err)
	}
}

func TestPrepareDirectoriesReportsFilesystemFailures(t *testing.T) {
	failure := errors.New("injected failure")
	tests := []struct {
		name  string
		files *fakeStoreFiles
		want  string
	}{
		{"mkdir", &fakeStoreFiles{mkdirErr: failure}, "prepare evidence directory"},
		{"inspect", &fakeStoreFiles{mkdirErr: os.ErrExist, lstatErr: failure}, "inspect evidence directory"},
		{"not directory", &fakeStoreFiles{mkdirErr: os.ErrExist, info: fakeInfo{}}, "not a real directory"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := prepareDirectories(test.files, "/evidence", "test"); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("prepareDirectories() error = %v", err)
			}
		})
	}
}

type fakeStoreFiles struct {
	file      durableFile
	mkdirErr  error
	createErr error
	linkErr   error
	readErr   error
	lstatErr  error
	info      os.FileInfo
}

func (files *fakeStoreFiles) Mkdir(string, os.FileMode) error { return files.mkdirErr }
func (files *fakeStoreFiles) Lstat(string) (os.FileInfo, error) {
	return files.info, files.lstatErr
}
func (files *fakeStoreFiles) MkdirAll(string, os.FileMode) error { return files.mkdirErr }
func (files *fakeStoreFiles) CreateTemp(string, string) (durableFile, error) {
	return files.file, files.createErr
}
func (files *fakeStoreFiles) Link(string, string) error       { return files.linkErr }
func (files *fakeStoreFiles) ReadFile(string) ([]byte, error) { return nil, files.readErr }
func (*fakeStoreFiles) Remove(string) error                   { return nil }

type fakeDurableFile struct {
	writeErr error
	syncErr  error
	closeErr error
}

func (file *fakeDurableFile) Write(value []byte) (int, error) { return len(value), file.writeErr }
func (file *fakeDurableFile) Sync() error                     { return file.syncErr }
func (file *fakeDurableFile) Close() error                    { return file.closeErr }
func (*fakeDurableFile) Name() string                         { return "temporary" }

func validInternalRecord() Record {
	digest := "sha256:" + strings.Repeat("a", 64)
	return Record{
		SchemaVersion: SchemaVersion, Repository: "example", Module: ".", Gate: "test",
		InputDigest: digest, VerifierDigest: digest, Result: "passed", ReportDigest: digest,
		CompletedAt: timeForTest,
	}
}

var timeForTest = func() (value time.Time) { return time.Unix(1, 0).UTC() }()

type fakeInfo struct{}

func (fakeInfo) Name() string       { return "evidence" }
func (fakeInfo) Size() int64        { return 0 }
func (fakeInfo) Mode() os.FileMode  { return 0 }
func (fakeInfo) ModTime() time.Time { return timeForTest }
func (fakeInfo) IsDir() bool        { return false }
func (fakeInfo) Sys() any           { return nil }
