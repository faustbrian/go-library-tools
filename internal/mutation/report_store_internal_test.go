package mutation

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var killedReport = []byte(`{"files":[{"file_name":"source.go","mutations":[{"type":"A","status":"KILLED","line":1,"column":1}]}]}`)

func TestStoreReportReportsAtomicFilesystemFailures(t *testing.T) {
	failure := errors.New("injected failure")
	tests := []struct {
		name  string
		files *fakeReportFiles
		want  string
	}{
		{"mkdir", &fakeReportFiles{mkdirErr: failure}, "prepare mutation report directory"},
		{"create", &fakeReportFiles{info: reportDirectoryInfo{}, createErr: failure}, "create temporary mutation report"},
		{"missing temporary", &fakeReportFiles{info: reportDirectoryInfo{}}, "create temporary mutation report returned no file"},
		{"write", &fakeReportFiles{info: reportDirectoryInfo{}, file: &fakeReportFile{writeErr: failure}}, "write mutation report"},
		{"sync", &fakeReportFiles{info: reportDirectoryInfo{}, file: &fakeReportFile{syncErr: failure}}, "write mutation report"},
		{"close", &fakeReportFiles{info: reportDirectoryInfo{}, file: &fakeReportFile{closeErr: failure}}, "write mutation report"},
		{"publish", &fakeReportFiles{info: reportDirectoryInfo{}, file: &fakeReportFile{}, linkErr: failure}, "publish mutation report"},
		{"read", &fakeReportFiles{info: reportDirectoryInfo{}, file: &fakeReportFile{}, linkErr: os.ErrExist, readErr: failure}, "read existing mutation report"},
		{"invalid existing", &fakeReportFiles{info: reportDirectoryInfo{}, file: &fakeReportFile{}, linkErr: os.ErrExist, data: []byte(`{}`)}, "different content"},
		{"different existing", &fakeReportFiles{info: reportDirectoryInfo{}, file: &fakeReportFile{}, linkErr: os.ErrExist, data: []byte(`{"files":[]}`)}, "different content"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, _, _, _, err := storeReport(test.files, t.TempDir(), "sha256:"+strings.Repeat("a", 64), killedReport)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("storeReport() error = %v", err)
			}
		})
	}
}

func TestReportStoreRejectsInvalidArgumentsAndReports(t *testing.T) {
	for _, input := range []struct {
		root   string
		digest string
		report []byte
	}{
		{"relative", "sha256:" + strings.Repeat("a", 64), killedReport},
		{t.TempDir(), strings.Repeat("a", 64), killedReport},
		{t.TempDir(), "sha256:bad", killedReport},
		{t.TempDir(), "sha256:" + strings.Repeat("a", 64), []byte(`{}`)},
	} {
		if _, _, _, err := StoreReport(input.root, input.digest, input.report); err == nil {
			t.Fatalf("StoreReport(%q, %q) error = nil", input.root, input.digest)
		}
	}
	if _, _, err := LoadReport("relative", "sha256:"+strings.Repeat("a", 64)); err == nil {
		t.Fatal("LoadReport(relative) error = nil")
	}
	if _, _, err := LoadReport(t.TempDir(), "sha256:bad"); err == nil {
		t.Fatal("LoadReport(bad digest) error = nil")
	}
	if _, _, err := LoadReport(t.TempDir(), "sha256:"+strings.Repeat("a", 64)); err == nil {
		t.Fatal("LoadReport(missing) error = nil")
	}
	root := t.TempDir()
	directory := filepath.Join(root, "reports")
	if err := os.MkdirAll(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, strings.Repeat("a", 64)+".json"), []byte(`{}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := LoadReport(root, "sha256:"+strings.Repeat("a", 64)); err == nil {
		t.Fatal("LoadReport(invalid report) error = nil")
	}
}

func TestPrepareReportDirectoriesRejectsInvalidPaths(t *testing.T) {
	failure := errors.New("injected failure")
	for _, test := range []struct {
		name  string
		files *fakeReportFiles
		want  string
	}{
		{"inspect", &fakeReportFiles{mkdirErr: os.ErrExist, lstatErr: failure}, "inspect mutation report directory"},
		{"missing metadata", &fakeReportFiles{mkdirErr: os.ErrExist}, "inspect mutation report directory"},
		{"not directory", &fakeReportFiles{mkdirErr: os.ErrExist, info: reportFileInfo{}}, "not a real directory"},
	} {
		t.Run(test.name, func(t *testing.T) {
			if err := prepareReportDirectories(test.files, "/mutation", "/mutation/reports"); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("prepareReportDirectories() error = %v", err)
			}
		})
	}
}

type fakeReportFiles struct {
	file      durableReportFile
	info      os.FileInfo
	mkdirErr  error
	lstatErr  error
	createErr error
	linkErr   error
	readErr   error
	data      []byte
}

func (files *fakeReportFiles) Mkdir(string, os.FileMode) error   { return files.mkdirErr }
func (files *fakeReportFiles) Lstat(string) (os.FileInfo, error) { return files.info, files.lstatErr }
func (files *fakeReportFiles) CreateTemp(string, string) (durableReportFile, error) {
	return files.file, files.createErr
}
func (files *fakeReportFiles) Link(string, string) error       { return files.linkErr }
func (files *fakeReportFiles) ReadFile(string) ([]byte, error) { return files.data, files.readErr }
func (*fakeReportFiles) Remove(string) error                   { return nil }

type fakeReportFile struct {
	writeErr error
	syncErr  error
	closeErr error
}

func (file *fakeReportFile) Write(value []byte) (int, error) { return len(value), file.writeErr }
func (file *fakeReportFile) Sync() error                     { return file.syncErr }
func (file *fakeReportFile) Close() error                    { return file.closeErr }
func (*fakeReportFile) Name() string                         { return "temporary" }

type reportFileInfo struct{}

func (reportFileInfo) Name() string       { return "report" }
func (reportFileInfo) Size() int64        { return 0 }
func (reportFileInfo) Mode() os.FileMode  { return 0 }
func (reportFileInfo) ModTime() time.Time { return time.Unix(1, 0) }
func (reportFileInfo) IsDir() bool        { return false }
func (reportFileInfo) Sys() any           { return nil }

type reportDirectoryInfo struct{ reportFileInfo }

func (reportDirectoryInfo) Mode() os.FileMode { return os.ModeDir }
func (reportDirectoryInfo) IsDir() bool       { return true }
