package evidence

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
	"os"
	"strings"
	"testing"
	"time"
)

func TestInspectReportsFilesystemFailures(t *testing.T) {
	failure := errors.New("injected failure")
	valid, err := Marshal(inspectInternalRecord())
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name  string
		files *fakeInspectFiles
		want  string
	}{
		{"lstat", &fakeInspectFiles{lstatErr: failure}, "inspect evidence root"},
		{"walk", &fakeInspectFiles{walkErr: failure}, "inspect evidence"},
		{"walk callback", &fakeInspectFiles{callbackErr: failure}, "inspect evidence"},
		{"relative", &fakeInspectFiles{entry: fakeDirEntry{}, relErr: failure}, "inspect evidence"},
		{"open", &fakeInspectFiles{entry: fakeDirEntry{}, relative: inspectInternalPath(), openErr: failure}, "inspect evidence"},
		{"close", &fakeInspectFiles{entry: fakeDirEntry{}, relative: inspectInternalPath(), reader: &closingReader{Reader: bytes.NewReader(valid), closeErr: failure}}, "close evidence"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := inspect(test.files, "/repo", ".verification", "example", []string{"."})
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("inspect() error = %v", err)
			}
		})
	}
}

type fakeInspectFiles struct {
	lstatErr    error
	walkErr     error
	callbackErr error
	entry       fs.DirEntry
	relative    string
	relErr      error
	reader      io.ReadCloser
	openErr     error
}

func (files *fakeInspectFiles) Lstat(string) (os.FileInfo, error) {
	return inspectDirInfo{}, files.lstatErr
}
func (files *fakeInspectFiles) WalkDir(root string, visit fs.WalkDirFunc) error {
	if files.walkErr != nil {
		return files.walkErr
	}
	if files.callbackErr != nil {
		return visit(root, nil, files.callbackErr)
	}
	return visit(root+"/record.json", files.entry, nil)
}
func (files *fakeInspectFiles) Rel(string, string) (string, error) {
	return files.relative, files.relErr
}
func (files *fakeInspectFiles) Open(string) (io.ReadCloser, error) {
	return files.reader, files.openErr
}

type inspectDirInfo struct{}

func (inspectDirInfo) Name() string       { return ".verification" }
func (inspectDirInfo) Size() int64        { return 0 }
func (inspectDirInfo) Mode() os.FileMode  { return os.ModeDir | 0o700 }
func (inspectDirInfo) ModTime() time.Time { return time.Time{} }
func (inspectDirInfo) IsDir() bool        { return true }
func (inspectDirInfo) Sys() any           { return nil }

type fakeDirEntry struct{}

func (fakeDirEntry) Name() string               { return "record.json" }
func (fakeDirEntry) IsDir() bool                { return false }
func (fakeDirEntry) Type() fs.FileMode          { return 0 }
func (fakeDirEntry) Info() (os.FileInfo, error) { return inspectDirInfo{}, nil }

type closingReader struct {
	io.Reader
	closeErr error
}

func (reader *closingReader) Close() error { return reader.closeErr }

func inspectInternalRecord() Record {
	digest := "sha256:" + strings.Repeat("a", 64)
	return Record{SchemaVersion: 1, Repository: "example", Module: ".", Gate: "test",
		InputDigest: digest, VerifierDigest: digest, Result: "passed", ReportDigest: digest,
		CompletedAt: time.Unix(1, 0).UTC()}
}
func inspectInternalPath() string { return "by-input/test/" + strings.Repeat("a", 64) + ".json" }
