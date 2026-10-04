package docscheck

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/faustbrian/go-library-tools/v2/internal/repositoryfile"
)

func TestDocumentAdmissionPrecedesContentAcquisition(t *testing.T) {
	for _, test := range []struct {
		name    string
		mode    fs.FileMode
		maximum int64
		want    error
	}{
		{name: "inclusive", maximum: 4},
		{name: "one over", maximum: 3, want: repositoryfile.ErrTooLarge},
		{name: "nonregular metadata", mode: os.ModeNamedPipe, maximum: 4, want: repositoryfile.ErrNotRegular},
	} {
		t.Run(test.name, func(t *testing.T) {
			info, err := fs.Stat(fstest.MapFS{"README.md": {Data: []byte("# R\n"), Mode: test.mode}}, "README.md")
			if err != nil {
				t.Fatal(err)
			}
			inspections, opens := 0, 0
			stream := &documentStream{Reader: strings.NewReader("# R\n")}
			source := documentSource{
				inspect: func(string, string) (os.FileInfo, error) { inspections++; return info, nil },
				open:    func(context.Context, string, int64) (io.ReadCloser, error) { opens++; return stream, nil },
			}
			err = checkDocumentWithSource(t.Context(), "ordinary", "ordinary/README.md", test.maximum, source)
			if test.want != nil {
				if !errors.Is(err, test.want) || opens != 0 || stream.reads != 0 || stream.closes != 0 {
					t.Fatal("refused metadata dispatched content acquisition or lost its error class")
				}
			} else if err != nil || opens != 1 || stream.closes != 1 || stream.reads == 0 {
				t.Fatal("inclusive ordinary document was not validated and closed")
			}
			if inspections != 1 {
				t.Fatal("document metadata was not admitted before acquisition")
			}
		})
	}
}

func TestDocumentRetainedReadHasBoundedLookahead(t *testing.T) {
	info, err := fs.Stat(fstest.MapFS{"README.md": {Data: []byte("abc")}}, "README.md")
	if err != nil {
		t.Fatal(err)
	}
	reader := strings.NewReader("abcde")
	stream := &documentStream{Reader: reader}
	source := documentSource{
		inspect: func(string, string) (os.FileInfo, error) { return info, nil },
		open:    func(context.Context, string, int64) (io.ReadCloser, error) { return stream, nil },
	}
	err = checkDocumentWithSource(t.Context(), "ordinary", "ordinary/README.md", 3, source)
	if !errors.Is(err, repositoryfile.ErrTooLarge) || reader.Len() != 1 || stream.closes != 1 {
		t.Fatal("retained byte refusal did not preserve bounded lookahead and cleanup")
	}
}

func TestDocumentWalkBoundsAllEntriesAndBatches(t *testing.T) {
	for _, test := range []struct {
		name  string
		limit int
	}{{"inclusive", 2}, {"one over", 1}} {
		t.Run(test.name, func(t *testing.T) {
			limit := test.limit
			files := &observedDocumentFS{FS: fstest.MapFS{
				"README.md": {Data: []byte("# R\n")}, "note.txt": {Data: []byte("ordinary")},
			}}
			limits := documentLimits{bytes: 4, documents: 1, entries: limit, depth: 1}
			paths, err := documentsFS(t.Context(), "ordinary", files, limits)
			if limit == 2 {
				if err != nil || len(paths) != 1 || paths[0] != filepath.Join("ordinary", "README.md") {
					t.Fatal("inclusive traversal changed ordinary document selection")
				}
			} else if err == nil || !strings.Contains(err.Error(), "entry limit") || paths != nil {
				t.Fatal("non-Markdown entry escaped the shared traversal allowance")
			}
			for _, request := range files.requests {
				if request < 1 || request > 128 || request > limit+1 {
					t.Fatal("directory enumeration requested an unbounded or excessive listing")
				}
			}
			if len(files.requests) == 0 || files.opened != files.closed {
				t.Fatal("directory enumeration was vacuous or leaked a handle")
			}
		})
	}
}

func TestDocumentWalkDepthRefusesBeforeChildEnumeration(t *testing.T) {
	for _, depth := range []int{2, 1} {
		files := &observedDocumentFS{FS: fstest.MapFS{
			"README.md": {Data: []byte("# R\n")}, "docs/nested/guide.md": {Data: []byte("# G\n")},
		}}
		paths, err := documentsFS(t.Context(), "ordinary", files, documentLimits{bytes: 4, documents: 2, entries: 4, depth: depth})
		if depth == 2 {
			if err != nil || len(paths) != 2 {
				t.Fatal("inclusive depth refused ordinary documentation")
			}
		} else if err == nil || !strings.Contains(err.Error(), "depth limit") || paths != nil || slices.Contains(files.names, "docs/nested") {
			t.Fatal("over-depth directory was enumerated or lost its refusal")
		}
		if files.opened != files.closed {
			t.Fatal("depth admission leaked a directory handle")
		}
	}
}

func TestDocumentWalkRejectsSyntheticNonregularMarkdown(t *testing.T) {
	files := &observedDocumentFS{FS: fstest.MapFS{"README.md": {Data: []byte("# R\n"), Mode: os.ModeNamedPipe}}}
	paths, err := documentsFS(t.Context(), "ordinary", files, documentLimits{bytes: 4, documents: 1, entries: 1, depth: 1})
	if paths != nil || !errors.Is(err, repositoryfile.ErrNotRegular) || slices.Contains(files.names, "README.md") {
		t.Fatal("synthetic nonregular Markdown was admitted or opened")
	}
}

func TestDocumentCancellationStopsEnumerationAndPublicChecks(t *testing.T) {
	t.Run("between batches", func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()
		files := &observedDocumentFS{FS: fstest.MapFS{"README.md": {Data: []byte("# R\n")}}, onRead: cancel}
		paths, err := documentsFS(ctx, "ordinary", files, documentLimits{bytes: 4, documents: 1, entries: 1, depth: 1})
		if paths != nil || !errors.Is(err, context.Canceled) || len(files.requests) != 1 || files.opened != files.closed {
			t.Fatal("canceled directory acquisition continued or leaked its handle")
		}
	})
	t.Run("public entrypoints", func(t *testing.T) {
		root := basic(t)
		ctx, cancel := context.WithCancel(t.Context())
		cancel()
		for _, check := range []func() error{
			func() error { return CheckContext(ctx, root) },
			func() error { return CheckWithinContext(ctx, root, root) },
		} {
			if err := check(); !errors.Is(err, context.Canceled) {
				t.Fatal("public documentation validation lost cancellation")
			}
		}
	})
}

func TestDocumentOrdinaryOSAdapterAdmitsBeforeOpen(t *testing.T) {
	root := basic(t)
	path := filepath.Join(root, "README.md")
	write(t, path, "# R\n")
	for _, maximum := range []int64{4, 3} {
		source := ordinaryDocumentSource()
		open := source.open
		opens := 0
		source.open = func(ctx context.Context, name string, limit int64) (io.ReadCloser, error) {
			opens++
			return open(ctx, name, limit)
		}
		err := checkDocumentWithSource(t.Context(), root, path, maximum, source)
		if maximum == 4 {
			if err != nil || opens != 1 {
				t.Fatal("ordinary OS adapter did not validate inclusive Markdown")
			}
		} else if !errors.Is(err, repositoryfile.ErrTooLarge) || opens != 0 {
			t.Fatal("ordinary OS metadata did not refuse size before open")
		}
	}
}

func TestDocumentCountAndOrderAreRetained(t *testing.T) {
	files := fstest.MapFS{
		"README.md":       {Data: []byte("# R\n")},
		"docs/z.md":       {Data: []byte("# Z\n")},
		"docs/a.md":       {Data: []byte("# A\n")},
		"ignored/file.md": {Data: []byte("# I\n")},
	}
	want := []string{filepath.Join("ordinary", "README.md"), filepath.Join("ordinary", "docs", "a.md"), filepath.Join("ordinary", "docs", "z.md")}
	paths, err := documentsFS(t.Context(), "ordinary", files, documentLimits{bytes: 4, documents: 3, entries: 5, depth: 1})
	if err != nil || !slices.Equal(paths, want) {
		t.Fatal("inclusive document count, sorting or ignored-directory policy changed")
	}
	paths, err = documentsFS(t.Context(), "ordinary", files, documentLimits{bytes: 4, documents: 2, entries: 5, depth: 1})
	if paths != nil || err == nil || !strings.Contains(err.Error(), "file count") {
		t.Fatal("document count allowance admitted an extra Markdown file")
	}
}

func TestDocumentCancellationStopsReadAndParsePhases(t *testing.T) {
	info, err := fs.Stat(fstest.MapFS{"README.md": {Data: []byte("# R\n")}}, "README.md")
	if err != nil {
		t.Fatal(err)
	}
	for _, stage := range []string{"metadata", "open", "read", "close"} {
		t.Run(stage, func(t *testing.T) {
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			stream := &documentStream{Reader: strings.NewReader("# R\n")}
			switch stage {
			case "read":
				stream.onRead = cancel
			case "close":
				stream.onClose = cancel
			}
			opens := 0
			source := documentSource{
				inspect: func(string, string) (os.FileInfo, error) {
					if stage == "metadata" {
						cancel()
					}
					return info, nil
				},
				open: func(context.Context, string, int64) (io.ReadCloser, error) {
					opens++
					if stage == "open" {
						cancel()
					}
					return stream, nil
				},
			}
			err := checkDocumentWithSource(ctx, "ordinary", "ordinary/README.md", 4, source)
			if !errors.Is(err, context.Canceled) {
				t.Fatal("read or parsing phase published success after cancellation")
			}
			if stage == "metadata" {
				if opens != 0 || stream.closes != 0 || stream.reads != 0 {
					t.Fatal("canceled metadata dispatched content IO")
				}
			} else if opens != 1 || stream.closes != 1 || (stage == "open" && stream.reads != 0) {
				t.Fatal("canceled content acquisition leaked its handle or continued reading")
			}
		})
	}
}

func TestDocumentCanceledLocalLinkDoesNotPublishSuccess(t *testing.T) {
	root := basic(t)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if err := checkLinkContext(ctx, root, filepath.Join(root, "README.md"), "#ordinary"); !errors.Is(err, context.Canceled) {
		t.Fatal("local-link phase published success after cancellation")
	}
}

type documentStream struct {
	io.Reader
	reads, closes   int
	onRead, onClose func()
}

func (s *documentStream) Read(buffer []byte) (int, error) {
	s.reads++
	if s.onRead != nil {
		s.onRead()
	}
	return s.Reader.Read(buffer)
}

func (s *documentStream) Close() error {
	s.closes++
	if s.onClose != nil {
		s.onClose()
	}
	return nil
}

type observedDocumentFS struct {
	fs.FS
	requests       []int
	names          []string
	opened, closed int
	onRead         func()
}

func (s *observedDocumentFS) Open(name string) (fs.File, error) {
	f, err := s.FS.Open(name)
	if err != nil {
		return nil, err
	}
	s.names = append(s.names, name)
	if directory, ok := f.(fs.ReadDirFile); ok {
		s.opened++
		return &observedDocumentDirectory{ReadDirFile: directory, source: s}, nil
	}
	return f, nil
}

type observedDocumentDirectory struct {
	fs.ReadDirFile
	source *observedDocumentFS
}

func (d *observedDocumentDirectory) ReadDir(n int) ([]fs.DirEntry, error) {
	d.source.requests = append(d.source.requests, n)
	if d.source.onRead != nil {
		d.source.onRead()
	}
	return d.ReadDirFile.ReadDir(n)
}

func (d *observedDocumentDirectory) Close() error {
	d.source.closed++
	return d.ReadDirFile.Close()
}
