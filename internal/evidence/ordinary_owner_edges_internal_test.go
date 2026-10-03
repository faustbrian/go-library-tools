package evidence

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestOrdinaryParseAbsentRecordIsCategorical(t *testing.T) {
	got, err := Parse(strings.NewReader(""))
	if !errors.Is(err, ErrInvalid) || !reflect.DeepEqual(got, Record{}) {
		t.Fatalf("absent record = %#v, %v; want zero and ErrInvalid", got, err)
	}
	if err.Error() != "invalid evidence: decode failure" {
		t.Fatalf("absent record diagnostic = %q; want categorical decode failure", err.Error())
	}
}

func TestOrdinaryInspectMissingDirectoryIsEmpty(t *testing.T) {
	got, err := inspect(operatingInspectFileSystem{}, t.TempDir(), "missing", "ordinary", nil)
	if err != nil || got == nil || len(got) != 0 {
		t.Fatalf("missing directory = %#v, %v; want nonnil empty inventory", got, err)
	}
}

func TestOrdinaryInspectRegularFileIsNotDirectory(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "evidence"), []byte("ordinary"), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := inspect(operatingInspectFileSystem{}, root, "evidence", "ordinary", nil)
	if !errors.Is(err, ErrInvalid) || got != nil {
		t.Fatalf("regular file root = %#v, %v; want nil and ErrInvalid", got, err)
	}
	if err.Error() != "invalid evidence: evidence root is not a real directory" {
		t.Fatalf("regular file diagnostic = %q; want categorical directory rejection", err.Error())
	}
}
