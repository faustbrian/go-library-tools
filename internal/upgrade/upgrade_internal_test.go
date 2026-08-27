package upgrade

import (
	"errors"
	"os"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestApplyReportsPlanAndWriteFailuresWithRollback(t *testing.T) {
	failure := errors.New("injected failure")
	if _, err := Apply(t.TempDir(), Request{}); err == nil {
		t.Fatal("Apply() plan error = nil")
	}
	result := Result{
		Changed:       true,
		Files:         []File{{Path: ".github/workflows/ci.yml", Changed: true}, {Path: ".golib.yaml", Changed: true}},
		configuration: []byte("new config"), workflow: []byte("new workflow"),
		original: map[string][]byte{".golib.yaml": []byte("old config")},
	}
	tests := map[string]struct {
		failures map[int]error
		files    []File
		writes   []string
	}{
		"configuration":     {failures: map[int]error{1: failure}, files: result.Files, writes: []string{".golib.yaml:new config"}},
		"workflow rollback": {failures: map[int]error{2: failure}, files: result.Files, writes: []string{".golib.yaml:new config", ".github/workflows/ci.yml:new workflow", ".golib.yaml:old config"}},
		"rollback":          {failures: map[int]error{2: failure, 3: failure}, files: result.Files, writes: []string{".golib.yaml:new config", ".github/workflows/ci.yml:new workflow", ".golib.yaml:old config"}},
		"workflow only":     {failures: map[int]error{1: failure}, files: []File{{Path: ".github/workflows/ci.yml", Changed: true}, {Path: ".golib.yaml"}}, writes: []string{".github/workflows/ci.yml:new workflow"}},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			value := result
			value.Files = test.files
			var writes []string
			_, err := apply(value, "/repo", func(_ string, path string, content []byte) error {
				writes = append(writes, path+":"+string(content))
				return test.failures[len(writes)]
			})
			if err == nil || !slices.Equal(writes, test.writes) {
				t.Fatalf("apply() = %v, writes %v", err, writes)
			}
		})
	}
}

func TestResultHumanReportsPlanApplyAndCurrentState(t *testing.T) {
	current := Result{}
	if got := current.Human(false); got != "tooling pins already current\n" {
		t.Fatalf("current Human() = %q", got)
	}
	changed := Result{Changed: true, Files: []File{{Path: "b", Changed: false}, {Path: "a", Changed: true}}}
	if got := changed.Human(false); got != "tooling pins planned: a\n" {
		t.Fatalf("planned Human() = %q", got)
	}
	if got := changed.Human(true); got != "tooling pins updated: a\n" {
		t.Fatalf("applied Human() = %q", got)
	}
}

func TestAtomicWriteReportsEveryFileBoundary(t *testing.T) {
	failure := errors.New("injected failure")
	tests := map[string]*fakeAtomicFiles{
		"inspect":     {lstatErr: failure},
		"directory":   {info: fakeAtomicInfo{mode: os.ModeDir}},
		"symlink":     {info: fakeAtomicInfo{mode: os.ModeSymlink}},
		"create":      {info: fakeAtomicInfo{mode: 0o640}, createErr: failure},
		"chmod":       {info: fakeAtomicInfo{mode: 0o640}, file: &fakeAtomicFile{chmodErr: failure}},
		"write":       {info: fakeAtomicInfo{mode: 0o640}, file: &fakeAtomicFile{writeErr: failure}},
		"short write": {info: fakeAtomicInfo{mode: 0o640}, file: &fakeAtomicFile{writeN: 1}},
		"sync":        {info: fakeAtomicInfo{mode: 0o640}, file: &fakeAtomicFile{syncErr: failure}},
		"close":       {info: fakeAtomicInfo{mode: 0o640}, file: &fakeAtomicFile{closeErr: failure}},
		"rename":      {info: fakeAtomicInfo{mode: 0o640}, file: &fakeAtomicFile{}, renameErr: failure},
	}
	for name, files := range tests {
		t.Run(name, func(t *testing.T) {
			if err := atomicWriteWithFiles("/repo", ".golib.yaml", []byte("value"), files); err == nil {
				t.Fatal("atomicWriteWithFiles() error = nil")
			}
		})
	}
	files := &fakeAtomicFiles{info: fakeAtomicInfo{mode: 0o640}, file: &fakeAtomicFile{}}
	if err := atomicWriteWithFiles("/repo", ".golib.yaml", []byte("value"), files); err != nil {
		t.Fatal(err)
	}
	if files.removed != "/repo/.golib-upgrade" || files.renamed != "/repo/.golib.yaml" {
		t.Fatalf("atomic operations = %#v", files)
	}
}

type fakeAtomicFiles struct {
	info      os.FileInfo
	lstatErr  error
	file      *fakeAtomicFile
	createErr error
	renameErr error
	removed   string
	renamed   string
}

func (files *fakeAtomicFiles) Lstat(string) (os.FileInfo, error) { return files.info, files.lstatErr }
func (files *fakeAtomicFiles) CreateTemp(string, string) (atomicFile, error) {
	if files.file == nil {
		files.file = &fakeAtomicFile{}
	}
	return files.file, files.createErr
}
func (files *fakeAtomicFiles) Rename(_, target string) error {
	files.renamed = target
	return files.renameErr
}
func (files *fakeAtomicFiles) Remove(path string) error {
	files.removed = path
	return nil
}

type fakeAtomicFile struct {
	chmodErr error
	writeErr error
	writeN   int
	syncErr  error
	closeErr error
}

func (*fakeAtomicFile) Name() string                 { return "/repo/.golib-upgrade" }
func (file *fakeAtomicFile) Chmod(os.FileMode) error { return file.chmodErr }
func (file *fakeAtomicFile) Write(value []byte) (int, error) {
	if file.writeN > 0 {
		return file.writeN, file.writeErr
	}
	return len(value), file.writeErr
}
func (file *fakeAtomicFile) Sync() error  { return file.syncErr }
func (file *fakeAtomicFile) Close() error { return file.closeErr }

type fakeAtomicInfo struct{ mode os.FileMode }

func (fakeAtomicInfo) Name() string           { return "file" }
func (fakeAtomicInfo) Size() int64            { return 1 }
func (info fakeAtomicInfo) Mode() os.FileMode { return info.mode }
func (fakeAtomicInfo) ModTime() time.Time     { return time.Time{} }
func (info fakeAtomicInfo) IsDir() bool       { return info.mode.IsDir() }
func (fakeAtomicInfo) Sys() any               { return nil }

func TestBuildPlanReportsInjectedReadFailures(t *testing.T) {
	failure := errors.New("injected failure")
	request := Request{Version: "v1.1.0", WorkflowSHA: strings.Repeat("a", 40), ChecksumsSHA256: strings.Repeat("b", 64)}
	for failAt := 1; failAt <= 2; failAt++ {
		calls := 0
		_, err := buildPlan("/repo", request, func(string, string, int64) ([]byte, error) {
			calls++
			if calls == failAt {
				return nil, failure
			}
			return []byte("tool_version: v1.0.0\n"), nil
		})
		if err == nil {
			t.Fatalf("buildPlan() read %d error = nil", failAt)
		}
	}
}
