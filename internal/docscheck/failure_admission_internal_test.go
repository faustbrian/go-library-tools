package docscheck

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"os"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/faustbrian/go-library-tools/v2/internal/repositoryfile"
)

// Nil metadata/readers and partial open results characterize the private
// collaborator contract, not outcomes promised by the ordinary OS adapter.
func TestDocumentSourceFailuresPreserveCausesAndCleanup(t *testing.T) {
	// These numeric allowances belong only to the private helper contract;
	// public documentation checking supplies the valid fixed default.
	for _, maximum := range []int64{0, maximumDocumentSize + 1} {
		t.Run("invalid private allowance", func(t *testing.T) {
			inspections, opens := 0, 0
			stream := &failureDocumentStream{Reader: strings.NewReader("# R\n")}
			source := documentSource{
				inspect: func(string, string) (os.FileInfo, error) {
					inspections++
					return nil, nil
				},
				open: func(context.Context, string, int64) (io.ReadCloser, error) {
					opens++
					return stream, nil
				},
			}
			err := checkDocumentWithSource(t.Context(), "ordinary", "ordinary/README.md", maximum, source)
			if !errors.Is(err, repositoryfile.ErrTooLarge) || inspections != 0 || opens != 0 || stream.reads != 0 || stream.closes != 0 {
				t.Fatal("invalid private allowance dispatched acquisition or lost its refusal class")
			}
		})
	}
	info, err := fs.Stat(fstest.MapFS{"README.md": {Data: []byte("# R\n")}}, "README.md")
	if err != nil {
		t.Fatal(err)
	}
	failure := errors.New("ordinary acquisition failure")
	closeFailure := errors.New("ordinary close failure")
	for _, stage := range []string{"canceled", "inspect", "nil metadata", "open", "partial open", "nil reader", "read", "close"} {
		t.Run(stage, func(t *testing.T) {
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			if stage == "canceled" {
				cancel()
			}
			inspections, opens := 0, 0
			stream := &failureDocumentStream{Reader: strings.NewReader("# R\n")}
			want := failure
			switch stage {
			case "canceled":
				want = context.Canceled
			case "nil metadata", "nil reader":
				want = repositoryfile.ErrUnsafePath
			case "read":
				stream.Reader = documentErrorReader{err: failure}
				stream.closeErr = closeFailure
			case "partial open", "close":
				stream.closeErr = closeFailure
				if stage == "close" {
					want = closeFailure
				}
			}
			source := documentSource{
				inspect: func(string, string) (os.FileInfo, error) {
					inspections++
					switch stage {
					case "inspect":
						return nil, failure
					case "nil metadata":
						return nil, nil
					default:
						return info, nil
					}
				},
				open: func(context.Context, string, int64) (io.ReadCloser, error) {
					opens++
					switch stage {
					case "open":
						return nil, failure
					case "partial open":
						return stream, failure
					case "nil reader":
						return nil, nil
					default:
						return stream, nil
					}
				},
			}
			err := checkDocumentWithSource(ctx, "ordinary", "ordinary/README.md", 4, source)
			if !errors.Is(err, want) {
				t.Fatal("document refusal lost its cause or published success")
			}
			wantInspections, wantOpens, wantCloses := 1, 1, 0
			switch stage {
			case "canceled":
				wantInspections, wantOpens = 0, 0
			case "inspect", "nil metadata":
				wantOpens = 0
			case "partial open", "read", "close":
				wantCloses = 1
			}
			if inspections != wantInspections || opens != wantOpens || stream.closes != wantCloses {
				t.Fatal("document refusal dispatched later acquisition or lost owned cleanup")
			}
			if stage == "partial open" && !errors.Is(err, closeFailure) {
				t.Fatal("partial acquisition lost its owned cleanup cause")
			}
			if stage == "read" || stage == "close" {
				if stream.reads == 0 || !errors.Is(err, closeFailure) {
					t.Fatal("acquired reader failure was vacuous or lost the cleanup cause")
				}
			} else if stream.reads != 0 {
				t.Fatal("refused acquisition dispatched content reading")
			}
		})
	}
}

func TestDocumentWalkFailuresPreserveCausesAndCleanup(t *testing.T) {
	failure := errors.New("ordinary directory failure")
	closeFailure := errors.New("ordinary directory close failure")
	for _, stage := range []string{"open", "stat", "nil metadata", "not directory", "not enumerable", "enumeration", "empty progress", "excess batch", "entry metadata"} {
		t.Run(stage, func(t *testing.T) {
			files := newFailureDocumentFS(t)
			files.closeErr = closeFailure
			want := failure
			switch stage {
			case "open":
				files.openErr = failure
			case "stat":
				files.statErr = failure
			case "nil metadata":
				files.nilInfo = true
				want = repositoryfile.ErrNotRegular
			case "not directory":
				info, err := fs.Stat(fstest.MapFS{"ordinary": {Data: []byte("a")}}, "ordinary")
				if err != nil {
					t.Fatal(err)
				}
				files.info = info
				want = repositoryfile.ErrNotRegular
			case "not enumerable":
				files.noReadDir = true
				want = repositoryfile.ErrNotRegular
			case "enumeration":
				files.readDir = func(int) ([]fs.DirEntry, error) { return nil, failure }
			case "empty progress":
				// These two responses intentionally violate fs.ReadDirFile's
				// contract; they only prove the existing defensive refusal.
				files.readDir = func(int) ([]fs.DirEntry, error) { return nil, nil }
			case "excess batch":
				files.readDir = func(n int) ([]fs.DirEntry, error) {
					return make([]fs.DirEntry, n+1), io.EOF
				}
			case "entry metadata":
				files.readDir = func(int) ([]fs.DirEntry, error) {
					return []fs.DirEntry{failureDocumentEntry{DirEntry: files.entry, infoErr: failure}}, io.EOF
				}
			}
			paths, err := documentsFS(t.Context(), "ordinary", files, documentLimits{bytes: 4, documents: 1, entries: 1, depth: 1})
			if paths != nil || err == nil {
				t.Fatal("refused directory published document paths")
			}
			switch stage {
			case "empty progress", "excess batch":
				if !strings.Contains(err.Error(), "enumeration violated its bound") {
					t.Fatal("invalid directory collaborator lost its bounded-enumeration refusal")
				}
			default:
				if !errors.Is(err, want) {
					t.Fatal("directory refusal lost its error class or cause")
				}
			}
			wantCloses := 1
			if stage == "open" {
				wantCloses = 0
			} else if !errors.Is(err, closeFailure) {
				t.Fatal("directory refusal lost the owned cleanup cause")
			}
			if files.opens != 1 || files.closes != wantCloses || files.contentReads != 0 {
				t.Fatal("directory refusal leaked an acquired handle or read content")
			}
			if stage == "open" || stage == "stat" || stage == "nil metadata" || stage == "not directory" || stage == "not enumerable" {
				if files.listings != 0 {
					t.Fatal("refused directory admission dispatched enumeration")
				}
			} else if files.listings != 1 {
				t.Fatal("enumeration failure was vacuous or continued listing")
			}
		})
	}
}

func TestDocumentWalkCancellationCheckpointsStopLaterWork(t *testing.T) {
	for _, stage := range []string{"before open", "after open", "after stat", "between entries"} {
		t.Run(stage, func(t *testing.T) {
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			files := newFailureDocumentFS(t)
			switch stage {
			case "before open":
				cancel()
			case "after open":
				files.onOpen = cancel
			case "after stat":
				files.onStat = cancel
			case "between entries":
				files.readDir = func(int) ([]fs.DirEntry, error) {
					return []fs.DirEntry{failureDocumentEntry{DirEntry: files.entry, onName: cancel}, files.entry}, io.EOF
				}
			}
			paths, err := documentsFS(ctx, "ordinary", files, documentLimits{bytes: 4, documents: 2, entries: 2, depth: 1})
			if paths != nil || !errors.Is(err, context.Canceled) || files.contentReads != 0 {
				t.Fatal("canceled traversal published success or read file content")
			}
			wantOpens, wantStats, wantListings := 1, 1, 0
			switch stage {
			case "before open":
				wantOpens, wantStats = 0, 0
			case "after open":
				wantStats = 0
			case "between entries":
				wantListings = 1
			}
			if files.opens != wantOpens || files.closes != wantOpens || files.stats != wantStats || files.listings != wantListings {
				t.Fatal("cancellation continued directory work or lost handle cleanup")
			}
		})
	}
}

type failureDocumentStream struct {
	io.Reader
	reads, closes int
	closeErr      error
}

func (s *failureDocumentStream) Read(p []byte) (int, error) {
	s.reads++
	return s.Reader.Read(p)
}

func (s *failureDocumentStream) Close() error { s.closes++; return s.closeErr }

type documentErrorReader struct{ err error }

func (r documentErrorReader) Read([]byte) (int, error) { return 0, r.err }

type failureDocumentFS struct {
	fs.FS
	info                                         fs.FileInfo
	entry                                        fs.DirEntry
	openErr, statErr, closeErr                   error
	nilInfo, noReadDir                           bool
	opens, stats, listings, closes, contentReads int
	onOpen, onStat                               func()
	readDir                                      func(int) ([]fs.DirEntry, error)
}

func newFailureDocumentFS(t *testing.T) *failureDocumentFS {
	t.Helper()
	files := fstest.MapFS{"README.md": {Data: []byte("# R\n")}}
	info, err := fs.Stat(files, ".")
	if err != nil {
		t.Fatal(err)
	}
	entries, err := fs.ReadDir(files, ".")
	if err != nil {
		t.Fatal(err)
	}
	return &failureDocumentFS{FS: files, info: info, entry: entries[0]}
}

func (f *failureDocumentFS) Open(name string) (fs.File, error) {
	f.opens++
	if f.onOpen != nil {
		f.onOpen()
	}
	if f.openErr != nil {
		return nil, f.openErr
	}
	opened, err := f.FS.Open(name)
	if err != nil {
		return nil, err
	}
	file := &failureDocumentFile{File: opened, source: f}
	if f.noReadDir {
		return file, nil
	}
	return &failureDocumentDirectory{failureDocumentFile: file, reader: opened.(fs.ReadDirFile)}, nil
}

type failureDocumentFile struct {
	fs.File
	source *failureDocumentFS
}

func (f *failureDocumentFile) Stat() (fs.FileInfo, error) {
	f.source.stats++
	if f.source.onStat != nil {
		f.source.onStat()
	}
	if f.source.nilInfo {
		return nil, nil
	}
	return f.source.info, f.source.statErr
}

func (f *failureDocumentFile) Read(p []byte) (int, error) {
	f.source.contentReads++
	return f.File.Read(p)
}

func (f *failureDocumentFile) Close() error {
	f.source.closes++
	return errors.Join(f.File.Close(), f.source.closeErr)
}

type failureDocumentDirectory struct {
	*failureDocumentFile
	reader fs.ReadDirFile
}

func (d *failureDocumentDirectory) ReadDir(n int) ([]fs.DirEntry, error) {
	d.source.listings++
	if d.source.readDir != nil {
		return d.source.readDir(n)
	}
	return d.reader.ReadDir(n)
}

type failureDocumentEntry struct {
	fs.DirEntry
	infoErr error
	onName  func()
}

func (e failureDocumentEntry) Name() string {
	if e.onName != nil {
		e.onName()
	}
	return e.DirEntry.Name()
}

func (e failureDocumentEntry) Info() (fs.FileInfo, error) {
	if e.infoErr != nil {
		return nil, e.infoErr
	}
	return e.DirEntry.Info()
}
