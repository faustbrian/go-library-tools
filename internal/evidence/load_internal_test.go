package evidence

import (
	"errors"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

func TestLoadReportsOpenCloseAndIdentityFailures(t *testing.T) {
	digest := "sha256:" + strings.Repeat("a", 64)
	record := validInternalRecord()
	data, err := Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name  string
		files loadFileSystem
		want  string
	}{
		{"open", fakeLoadFiles{info: loadFileInfo{}, openErr: errors.New("open failed")}, "open evidence"},
		{"parse", fakeLoadFiles{info: loadFileInfo{}, file: &loadReadCloser{Reader: strings.NewReader(`{}`)}}, "schema_version"},
		{"close", fakeLoadFiles{info: loadFileInfo{}, file: &loadReadCloser{Reader: strings.NewReader(string(data)), closeErr: errors.New("close failed")}}, "close evidence"},
	} {
		t.Run(test.name, func(t *testing.T) {
			if _, err := load(test.files, "/evidence", "test", digest); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("load() error = %v", err)
			}
		})
	}
	record.Gate = "other"
	data, err = Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	files := fakeLoadFiles{info: loadFileInfo{}, file: &loadReadCloser{Reader: strings.NewReader(string(data))}}
	if _, err := load(files, "/evidence", "test", digest); !errors.Is(err, ErrInvalid) {
		t.Fatalf("load(identity mismatch) error = %v", err)
	}
}

type fakeLoadFiles struct {
	info    os.FileInfo
	file    io.ReadCloser
	openErr error
}

func (files fakeLoadFiles) Lstat(string) (os.FileInfo, error) { return files.info, nil }
func (files fakeLoadFiles) Open(string) (io.ReadCloser, error) {
	return files.file, files.openErr
}

type loadReadCloser struct {
	io.Reader
	closeErr error
}

func (file *loadReadCloser) Close() error { return file.closeErr }

type loadFileInfo struct{}

func (loadFileInfo) Name() string       { return "evidence.json" }
func (loadFileInfo) Size() int64        { return 1 }
func (loadFileInfo) Mode() os.FileMode  { return 0o600 }
func (loadFileInfo) ModTime() time.Time { return time.Unix(1, 0) }
func (loadFileInfo) IsDir() bool        { return false }
func (loadFileInfo) Sys() any           { return nil }
