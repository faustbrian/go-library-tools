package repositoryfile

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStreamingOpenRefusesNonpositiveBudgetBeforeAcquisition(t *testing.T) {
	for _, maximum := range []int64{0, -1} {
		statCalls, openCalls := 0, 0
		opened, err := openContext(context.Background(), "ordinary", maximum,
			func(string) (os.FileInfo, error) { statCalls++; return nil, nil },
			func(string) (file, error) { openCalls++; return nil, nil })
		if opened != nil || !errors.Is(err, ErrTooLarge) || statCalls != 0 || openCalls != 0 {
			t.Fatalf("streaming allowance %d: reader published or refusal dispatched acquisition", maximum)
		}
	}
}

func TestStreamingOpenCancellationCheckpointsReleaseOwnership(t *testing.T) {
	info := ordinaryLifecycleMetadata(t)
	for _, stage := range []string{"metadata", "open", "opened metadata"} {
		t.Run(stage, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			owned := &lifecycleFile{reader: strings.NewReader("abc"), info: info}
			if stage == "opened metadata" {
				owned.onStat = cancel
			}
			openCalls := 0
			opened, err := openContext(ctx, "ordinary", 3,
				func(string) (os.FileInfo, error) {
					if stage == "metadata" {
						cancel()
					}
					return info, nil
				},
				func(string) (file, error) {
					openCalls++
					if stage == "open" {
						cancel()
					}
					return owned, nil
				})
			if opened != nil || !errors.Is(err, context.Canceled) {
				t.Fatal("canceled acquisition published a reader or replaced cancellation")
			}
			wantOpen, wantStat, wantClose := 1, 0, 1
			switch stage {
			case "metadata":
				wantOpen, wantClose = 0, 0
			case "opened metadata":
				wantStat = 1
			}
			if openCalls != wantOpen || owned.statCalls != wantStat || owned.closeCalls != wantClose || owned.readCalls != 0 {
				t.Fatal("cancellation dispatched later IO or leaked/double-closed an acquired handle")
			}
		})
	}
}

// Nil metadata and a partial open result are defensive injected-owner
// contracts; these cases do not assert that os.Stat or os.Open produces them.
func TestStreamingOpenDefensiveReturnsPreserveRefusalAndCleanup(t *testing.T) {
	t.Run("missing metadata", func(t *testing.T) {
		openCalls := 0
		opened, err := openContext(context.Background(), "ordinary", 3,
			func(string) (os.FileInfo, error) { return nil, nil },
			func(string) (file, error) { openCalls++; return nil, nil })
		if opened != nil || !errors.Is(err, ErrUnsafePath) || err.Error() != "inspect ordinary: unsafe repository path" || openCalls != 0 {
			t.Fatal("missing metadata did not retain unsafe-path refusal before open")
		}
	})
	t.Run("partial open", func(t *testing.T) {
		info := ordinaryLifecycleMetadata(t)
		failure := errors.New("ordinary open failure")
		owned := &lifecycleFile{reader: strings.NewReader("abc"), info: info}
		opened, err := openContext(context.Background(), "ordinary", 3,
			func(string) (os.FileInfo, error) { return info, nil },
			func(string) (file, error) { return owned, failure })
		if opened != nil || !errors.Is(err, failure) || err.Error() != "open ordinary: ordinary open failure" {
			t.Fatal("partial open published a reader or lost the original failure")
		}
		if owned.closeCalls != 1 || owned.statCalls != 0 || owned.readCalls != 0 {
			t.Fatal("partial open leaked/double-closed its handle or dispatched later IO")
		}
	})
}

func TestBoundedReadRetainsStreamLimitAfterMetadataAdmission(t *testing.T) {
	info := ordinaryLifecycleMetadata(t)
	for _, test := range []struct {
		name, content string
		tooLarge      bool
		remaining     int
	}{
		{name: "inclusive", content: "abc"},
		{name: "one over", content: "abcd", tooLarge: true},
		{name: "bounded lookahead", content: "abcde", tooLarge: true, remaining: 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			stream := strings.NewReader(test.content)
			owned := &lifecycleFile{reader: stream, info: info}
			// A stable ordinary metadata snapshot admits three bytes. The
			// injected stream isolates the independent retained-byte bound;
			// it does not perform a real filesystem growth/race experiment.
			data, err := read("ordinary-root", "ordinary", 3, fakeFileSystem{info: info, opened: owned})
			if test.tooLarge {
				if data != nil || !errors.Is(err, ErrTooLarge) || err.Error() != "repository file exceeds maximum size: ordinary" {
					t.Fatal("over-budget stream published bytes or lost size refusal")
				}
			} else if err != nil || string(data) != "abc" {
				t.Fatal("inclusive stream changed admitted bytes")
			}
			if owned.statCalls != 1 || owned.readCalls == 0 || owned.closeCalls != 1 || stream.Len() != test.remaining {
				t.Fatal("stream admission skipped IO, leaked/double-closed its handle, or exceeded bounded lookahead")
			}
		})
	}
}

func ordinaryLifecycleMetadata(t *testing.T) os.FileInfo {
	t.Helper()
	path := filepath.Join(t.TempDir(), "ordinary")
	if err := os.WriteFile(path, []byte("abc"), 0o600); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	return info
}

type lifecycleFile struct {
	reader                           io.Reader
	info                             os.FileInfo
	onStat                           func()
	statCalls, readCalls, closeCalls int
}

func (f *lifecycleFile) Stat() (os.FileInfo, error) {
	f.statCalls++
	if f.onStat != nil {
		f.onStat()
	}
	return f.info, nil
}

func (f *lifecycleFile) Read(buffer []byte) (int, error) {
	f.readCalls++
	return f.reader.Read(buffer)
}

func (f *lifecycleFile) Close() error {
	f.closeCalls++
	return nil
}
