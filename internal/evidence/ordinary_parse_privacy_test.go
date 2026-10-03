package evidence_test

import (
	"bytes"
	"errors"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/faustbrian/go-library-tools/v2/internal/evidence"
)

type ordinaryEvidenceReaderFailure struct{ cause error }

func (reader ordinaryEvidenceReaderFailure) Read([]byte) (int, error) {
	return 0, reader.cause
}

func TestOrdinaryEvidenceParseRedactsReaderCause(t *testing.T) {
	got, err := evidence.Parse(ordinaryEvidenceReaderFailure{cause: errors.New("ordinary reader diagnostic")})
	if !errors.Is(err, evidence.ErrInvalid) || !reflect.DeepEqual(got, evidence.Record{}) {
		t.Fatal("reader failure must retain ErrInvalid and a zero record")
	}
	if err.Error() != "invalid evidence: read failure" {
		t.Fatalf("default reader diagnostic = %q; want categorical failure", err.Error())
	}
}

func TestOrdinaryEvidenceParsePreservesCompleteRecord(t *testing.T) {
	want := validRecord()
	data, err := evidence.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	got, err := evidence.Parse(bytes.NewReader(data))
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("ordinary record roundtrip = %#v, %v; want unchanged complete record", got, err)
	}
}

func TestOrdinaryEvidenceInspectRedactsUnknownModule(t *testing.T) {
	root := t.TempDir()
	want := validRecord()
	want.Module = "ordinary-module"
	if _, _, err := evidence.Store(filepath.Join(root, ".verification"), want); err != nil {
		t.Fatal(err)
	}
	got, err := evidence.Inspect(root, ".verification", want.Repository, []string{"."})
	if !errors.Is(err, evidence.ErrInvalid) || got != nil {
		t.Fatal("unknown module must retain ErrInvalid without publishing an inventory")
	}
	if err.Error() != "inspect evidence: invalid evidence: evidence references unknown module" {
		t.Fatalf("default module diagnostic = %q; want categorical failure", err.Error())
	}
	got, err = evidence.Inspect(root, ".verification", want.Repository, []string{want.Module})
	if err != nil || !reflect.DeepEqual(got, []evidence.Record{want}) {
		t.Fatalf("allowed module inventory = %#v, %v; want unchanged complete record", got, err)
	}
}
