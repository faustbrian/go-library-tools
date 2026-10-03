package repositoryfile

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

type countedOpenFiles struct {
	fakeFileSystem
	calls int
}

type cooperativeFile struct {
	file
	cancel   context.CancelFunc
	closed   bool
	terminal error
}

func (f *cooperativeFile) Read(buffer []byte) (int, error) {
	n := copy(buffer, "abc")
	f.cancel()
	if f.terminal != nil {
		return n, f.terminal
	}
	return n, io.EOF
}

func TestOwnedReadClassifiesWrappedEOFWithoutReplacingTerminalFailure(t *testing.T) {
	for _, terminal := range []error{fmt.Errorf("ordinary completion: %w", io.EOF), errors.New("ordinary read failure")} {
		for _, canceled := range []bool{false, true} {
			ctx, cancel := context.WithCancel(context.Background())
			readerCancel := func() {}
			if canceled {
				readerCancel = cancel
			}
			reader := &contextFile{ctx: ctx, file: &cooperativeFile{cancel: readerCancel, terminal: terminal}}
			buffer := make([]byte, 3)
			n, err := reader.Read(buffer)
			cancel()
			if canceled && errors.Is(terminal, io.EOF) {
				if n != 0 || !errors.Is(err, context.Canceled) {
					t.Fatalf("canceled wrapped EOF published %d bytes, %v", n, err)
				}
				continue
			}
			// Classification alone would also accept a replacement wrapper.
			// These pointer-valued fixtures must retain the original error.
			if n != 3 || string(buffer[:n]) != "abc" || !errors.Is(err, terminal) || !reflect.ValueOf(err).Equal(reflect.ValueOf(terminal)) {
				t.Fatalf("ordinary bytes/terminal cause changed: n%d, %v", n, err)
			}
		}
	}
}

func (f *cooperativeFile) Close() error { f.closed = true; return nil }

func TestOwnedReadRetainsRealCancellationAfterReturnedBytes(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ordinary")
	if err := os.WriteFile(path, []byte("abc"), 0o600); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cooperative := &cooperativeFile{file: &fakeFile{info: info}, cancel: cancel}
	opened, err := openContext(ctx, path, 3, func(string) (os.FileInfo, error) { return info, nil }, func(string) (file, error) { return cooperative, nil })
	if err != nil {
		t.Fatal(err)
	}
	defer opened.Close()
	if n, err := opened.Read(make([]byte, 3)); n != 0 || !errors.Is(err, context.Canceled) {
		t.Fatalf("cooperative read = %d, %v", n, err)
	}
	if err := opened.Close(); err != nil || !cooperative.closed {
		t.Fatal("owned descriptor not closed")
	}
}

func TestOpenAdmissionRejectsNonregularWithoutDispatch(t *testing.T) {
	info, err := os.Stat(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	calls := 0
	opened, err := openContext(context.Background(), "ordinary", 10,
		func(string) (os.FileInfo, error) { return info, nil },
		func(string) (file, error) { calls++; return nil, errors.New("inert opener invoked") })
	if opened != nil || !errors.Is(err, ErrNotRegular) || calls != 0 {
		t.Fatalf("nonregular admission: %v, calls%d", err, calls)
	}
}

func TestOpenAdmissionChecksPostOpenGrowthAndCleanup(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ordinary")
	if err := os.WriteFile(path, []byte("abc"), 0o600); err != nil {
		t.Fatal(err)
	}
	var opened *os.File
	got, err := openContext(context.Background(), path, 3, os.Stat, func(string) (file, error) {
		if err := os.WriteFile(path, []byte("abcd"), 0o600); err != nil {
			return nil, err
		}
		var err error
		opened, err = os.Open(path)
		return opened, err
	})
	if got != nil || !errors.Is(err, ErrTooLarge) {
		t.Fatalf("postopen growth = %v", err)
	}
	if _, err := opened.Stat(); !errors.Is(err, os.ErrClosed) {
		t.Fatal("failed admission leaked descriptor")
	}
}

func TestOpenContextPreservesRegularSymlinkAndCancellation(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "ordinary")
	if err := os.WriteFile(path, []byte("abc"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("ordinary", filepath.Join(root, "link")); err != nil {
		t.Fatal(err)
	}
	confined, err := os.OpenRoot(root)
	if err != nil {
		t.Fatal(err)
	}
	defer confined.Close()
	for _, open := range []func(context.Context) (io.ReadCloser, error){
		func(ctx context.Context) (io.ReadCloser, error) {
			return OpenContext(ctx, filepath.Join(root, "link"), 3)
		},
		func(ctx context.Context) (io.ReadCloser, error) { return OpenRootContext(ctx, confined, "link", 3) },
	} {
		ctx, cancel := context.WithCancel(context.Background())
		f, err := open(ctx)
		if err != nil {
			cancel()
			t.Fatal(err)
		}
		data, err := io.ReadAll(f)
		if err != nil || string(data) != "abc" {
			t.Fatalf("ordinary read = %q, %v", data, err)
		}
		cancel()
		if _, err := f.Read(make([]byte, 1)); !errors.Is(err, context.Canceled) {
			t.Fatal("owned read ignored cancellation")
		}
		if err := f.Close(); err != nil {
			t.Fatal(err)
		}
		if f, err := open(ctx); f != nil || !errors.Is(err, context.Canceled) {
			t.Fatal("canceled open dispatched")
		}
	}
}

func (f *countedOpenFiles) Open(name string) (file, error) {
	f.calls++
	return f.fakeFileSystem.Open(name)
}

func TestReadAdmitsMetadataSizeBeforeOpenDispatch(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "ordinary")
	if err := os.WriteFile(path, []byte("abc"), 0o600); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	files := &countedOpenFiles{info: info, openErr: errors.New("inert opener invoked")}
	_, err = read(root, "ordinary", 2, files)
	if !errors.Is(err, ErrTooLarge) || files.calls != 0 {
		t.Fatalf("oversized admission = %v, open calls%d; want ErrTooLarge and no dispatch", err, files.calls)
	}
}
