package mutation

import (
	"errors"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/faustbrian/go-library-tools/v2/internal/repositoryfile"
)

func TestOrdinarySourceDirectoryAdmission(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"b.go", "a.go"} {
		if err := os.WriteFile(filepath.Join(root, name), nil, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	entries, err := readSourceEntries(root, 1)
	if !errors.Is(err, errSourceEntryLimit) || entries != nil {
		t.Fatalf("below directory allowance = %#v, %v; want refusal before retaining excess entries", entries, err)
	}
	entries, err = readSourceEntries(root, 2)
	if err != nil || len(entries) != 2 || entries[0].Name() != "a.go" || entries[1].Name() != "b.go" {
		t.Fatalf("exact directory allowance = %#v, %v; want sorted admitted entries", entries, err)
	}
}

func TestOrdinarySourceDirectoryNonpositiveAllowances(t *testing.T) {
	t.Run("negative allowance precedes opening", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "missing")
		entries, err := readSourceEntries(path, -1)
		if entries != nil || !errors.Is(err, errSourceEntryLimit) || errors.Is(err, os.ErrNotExist) {
			t.Fatalf("negative directory allowance = %#v, %v; want nil entries and entry-limit refusal before opening", entries, err)
		}
	})

	t.Run("zero allowance admits empty directory", func(t *testing.T) {
		entries, err := readSourceEntries(t.TempDir(), 0)
		if err != nil || len(entries) != 0 {
			t.Fatalf("empty directory at zero allowance = %#v, %v; want no entries and no error", entries, err)
		}
	})

	t.Run("zero allowance rejects one entry", func(t *testing.T) {
		root := t.TempDir()
		if err := os.WriteFile(filepath.Join(root, "a.go"), []byte("a"), 0o600); err != nil {
			t.Fatal(err)
		}
		entries, err := readSourceEntries(root, 0)
		if entries != nil || !errors.Is(err, errSourceEntryLimit) {
			t.Fatalf("one directory entry at zero allowance = %#v, %v; want nil entries and entry-limit refusal", entries, err)
		}
	})
}

func TestOrdinarySourceDirectoryEnumerationIOErrors(t *testing.T) {
	t.Run("missing directory", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "missing")
		entries, err := readSourceEntries(path, 1)
		if entries != nil || !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("missing directory enumeration = %#v, %v; want nil entries and os.ErrNotExist", entries, err)
		}
		var pathErr *os.PathError
		if !errors.As(err, &pathErr) || pathErr.Op != "open" || pathErr.Path != path {
			t.Fatalf("missing directory error = %v; want open error attributed to %q", err, path)
		}
	})

	t.Run("regular file", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "file")
		if err := os.WriteFile(path, []byte("file"), 0o600); err != nil {
			t.Fatal(err)
		}
		entries, err := readSourceEntries(path, 1)
		if entries != nil || err == nil {
			t.Fatalf("regular file enumeration = %#v, %v; want nil entries and an error", entries, err)
		}
		var pathErr *os.PathError
		if !errors.As(err, &pathErr) || pathErr.Op == "" || pathErr.Op == "open" || pathErr.Path != path {
			t.Fatalf("regular file error = %v; want directory-read error attributed to %q", err, path)
		}
	})
}

func TestOrdinarySourceReadAdmitsBeforeBytes(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "source.go"), []byte("ab"), 0o600); err != nil {
		t.Fatal(err)
	}
	data, err := readSourceFile(root, "source.go", 1, 0)
	if !errors.Is(err, repositoryfile.ErrTooLarge) || data != nil {
		t.Fatalf("file-first admission = %q, %v; want per-file refusal", data, err)
	}
	data, err = readSourceFile(root, "source.go", 2, 1)
	if !errors.Is(err, errSourceTotalLimit) || data != nil {
		t.Fatalf("remaining-byte admission = %q, %v; want aggregate refusal", data, err)
	}
	data, err = readSourceFile(root, "source.go", 2, 2)
	if err != nil || string(data) != "ab" {
		t.Fatalf("exact admitted bytes = %q, %v", data, err)
	}
	if err := os.WriteFile(filepath.Join(root, "empty.go"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	data, err = readSourceFile(root, "empty.go", 2, 0)
	if err != nil || len(data) != 0 {
		t.Fatalf("empty file at zero remaining = %q, %v", data, err)
	}
}

func TestOrdinarySourceReadZeroAllowances(t *testing.T) {
	root := t.TempDir()
	for name, content := range map[string][]byte{"empty.go": nil, "one.go": []byte("a")} {
		if err := os.WriteFile(filepath.Join(root, name), content, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	for _, test := range []struct {
		name               string
		perFile, remaining int64
		cause              error
	}{
		{"per-file zero", 0, 1, repositoryfile.ErrTooLarge},
		{"aggregate zero", 1, 0, errSourceTotalLimit},
		{"both zero", 0, 0, repositoryfile.ErrTooLarge},
	} {
		t.Run(test.name, func(t *testing.T) {
			data, err := readSourceFile(root, "empty.go", test.perFile, test.remaining)
			if err != nil || len(data) != 0 {
				t.Fatalf("empty zero-allowance source = %q, %v; want empty success", data, err)
			}
			data, err = readSourceFile(root, "one.go", test.perFile, test.remaining)
			if data != nil || !errors.Is(err, test.cause) {
				t.Fatalf("one-byte zero-allowance source = %q, %v; want nil and %v", data, err, test.cause)
			}
		})
	}
}

func TestOrdinarySourceReadErrorNormalization(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "source.go"), []byte("ab"), 0o600); err != nil {
		t.Fatal(err)
	}
	data, zeroErr := repositoryfile.Read(root, "source.go", 0)
	if data != nil || !errors.Is(zeroErr, repositoryfile.ErrTooLarge) {
		t.Fatalf("real zero-bound refusal = %q, %v", data, zeroErr)
	}
	for _, test := range []struct {
		perFile, remaining int64
		want               error
	}{
		{0, 1, zeroErr},
		{0, 0, zeroErr},
		{1, 0, errSourceTotalLimit},
	} {
		if got := sourceReadError(zeroErr, test.perFile, test.remaining); !errors.Is(got, test.want) || reflect.ValueOf(got).Kind() != reflect.Pointer || reflect.ValueOf(got) != reflect.ValueOf(test.want) {
			t.Fatalf("zero-bound refusal with (%d, %d) = %v; want %v", test.perFile, test.remaining, got, test.want)
		}
	}
	data, positiveErr := repositoryfile.Read(root, "source.go", 1)
	if data != nil || !errors.Is(positiveErr, repositoryfile.ErrTooLarge) {
		t.Fatalf("real positive-bound refusal = %q, %v", data, positiveErr)
	}
	for _, limits := range [][2]int64{{1, 2}, {2, 1}, {1, 1}} {
		if got := sourceReadError(positiveErr, limits[0], limits[1]); !errors.Is(got, positiveErr) || reflect.ValueOf(got).Kind() != reflect.Pointer || reflect.ValueOf(got) != reflect.ValueOf(positiveErr) {
			t.Fatalf("positive-bound error changed with %v: %v", limits, got)
		}
	}
	for _, test := range []struct {
		path  string
		cause error
	}{
		{"missing.go", os.ErrNotExist},
		{".", repositoryfile.ErrNotRegular},
		{"../source.go", repositoryfile.ErrUnsafePath},
	} {
		data, original := repositoryfile.Read(root, test.path, 0)
		if data != nil || !errors.Is(original, test.cause) {
			t.Fatalf("real zero-bound path refusal for %q = %q, %v", test.path, data, original)
		}
		if got := sourceReadError(original, 1, 0); !errors.Is(got, test.cause) || reflect.ValueOf(got).Kind() != reflect.Pointer || reflect.ValueOf(got) != reflect.ValueOf(original) {
			t.Fatalf("non-size error changed for %q: %v", test.path, got)
		}
	}
}

func TestOrdinarySourceReadPropagatesOverflowingLookAheadLimit(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "source.go"), []byte("abc"), 0o600); err != nil {
		t.Fatal(err)
	}
	data, err := readSourceFile(root, "source.go", math.MaxInt64, math.MaxInt64)
	if data != nil || !errors.Is(err, repositoryfile.ErrTooLarge) {
		t.Fatalf("overflowing source read = %q, %v; want nil bytes and ErrTooLarge", data, err)
	}
	for _, limits := range [][2]int64{{math.MaxInt64 - 1, math.MaxInt64 - 1}, {math.MaxInt64, 3}, {3, math.MaxInt64}} {
		data, err := readSourceFile(root, "source.go", limits[0], limits[1])
		if err != nil || string(data) != "abc" {
			t.Fatalf("source read with allowances %v = %q, %v; want exact fixture bytes", limits, data, err)
		}
	}
}

func TestOrdinarySourceReadPreservesValidationErrors(t *testing.T) {
	root := t.TempDir()
	for _, test := range []struct {
		path string
		is   error
	}{
		{"missing.go", os.ErrNotExist},
		{".", repositoryfile.ErrNotRegular},
	} {
		t.Run(test.path, func(t *testing.T) {
			data, err := readSourceFile(root, test.path, 0, 0)
			if data != nil || !errors.Is(err, test.is) {
				t.Fatalf("source validation before zero allowances = %q, %v; want nil bytes and %v", data, err, test.is)
			}
			validationErr := repositoryfile.ValidateRegularFile(root, test.path)
			if validationErr == nil || err.Error() != validationErr.Error() {
				t.Fatalf("source validation diagnostic = %v; want unchanged %v", err, validationErr)
			}
		})
	}
}

func TestOrdinarySourceDigestFiniteAllowances(t *testing.T) {
	root := t.TempDir()
	const content = "package example\n"
	if err := os.WriteFile(filepath.Join(root, "a.go"), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "b.go"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	exact := sourceReadLimits{entries: 2, file: int64(len(content)), total: int64(len(content))}
	want, err := SourceDigest(root, ".", ".")
	if err != nil {
		t.Fatal(err)
	}
	got, err := sourceDigestWithLimits(operatingSourceFiles{}, root, ".", ".", exact)
	if err != nil || got != want {
		t.Fatalf("exact source digest = %s, %v; want unchanged admitted identity", got, err)
	}
	for _, below := range []sourceReadLimits{
		{entries: 1, file: exact.file, total: exact.total},
		{entries: 2, file: exact.file - 1, total: exact.total},
		{entries: 2, file: exact.file, total: exact.total - 1},
	} {
		got, err := sourceDigestWithLimits(operatingSourceFiles{}, root, ".", ".", below)
		if !errors.Is(err, ErrInvalid) || got != "" {
			t.Fatalf("below source allowance = %s, %v; want no identity and ErrInvalid", got, err)
		}
	}
}

func TestOrdinaryHistoricalSourceFiniteAllowances(t *testing.T) {
	root := t.TempDir()
	const content = "package example\n"
	for _, name := range []string{"b.go", "a.go"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	source, err := SourceDigest(root, ".", ".")
	if err != nil {
		t.Fatal(err)
	}
	exact := sourceReadLimits{entries: 2, file: int64(len(content)), total: 2 * int64(len(content))}
	if err := legacyPackagePathsUnchangedWithLimits(root, ".", ".", "example", source, exact); err != nil {
		t.Fatalf("exact historical admission = %v", err)
	}
	for _, below := range []sourceReadLimits{
		{entries: 1, file: exact.file, total: exact.total},
		{entries: 2, file: exact.file - 1, total: exact.total},
		{entries: 2, file: exact.file, total: exact.total - 1},
	} {
		if err := legacyPackagePathsUnchangedWithLimits(root, ".", ".", "example", source, below); !errors.Is(err, ErrInputChanged) {
			t.Fatalf("below historical allowance = %v; want ErrInputChanged", err)
		}
	}
	fileFirst := sourceReadLimits{entries: 2, file: exact.file - 1, total: 0}
	err = legacyPackagePathsUnchangedWithLimits(root, ".", ".", "example", source, fileFirst)
	if !errors.Is(err, ErrInputChanged) || !errors.Is(err, repositoryfile.ErrTooLarge) {
		t.Fatalf("historical file-first admission = %v; want per-file refusal preserved", err)
	}
	if err := os.WriteFile(filepath.Join(root, "a.go"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	source, err = SourceDigest(root, ".", ".")
	if err != nil {
		t.Fatal(err)
	}
	if err := legacyPackagePathsUnchangedWithLimits(root, ".", ".", "example", source, exact); !errors.Is(err, ErrInputChanged) {
		t.Fatalf("empty historical declaration = %v; want existing declaration refusal", err)
	}
}
