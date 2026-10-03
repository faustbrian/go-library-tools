package mutation

import (
	"encoding/json"
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"
)

func TestOrdinaryBootstrapReportRejectsWithoutPayloadDiagnostics(t *testing.T) {
	for _, test := range []struct {
		name, want string
		files      []reportFile
	}{
		{"unreviewed lived", "non-killed mutant lacks an exact equivalent review", []reportFile{{FileName: "ordinary.go", Mutations: []mutation{{Type: "NEGATION", Status: "LIVED", Line: 1, Column: 1}}}}},
		{"not covered", "non-killed mutant", []reportFile{{FileName: "ordinary.go", Mutations: []mutation{{Type: "NEGATION", Status: "NOT_COVERED", Line: 1, Column: 1}}}}},
		{"duplicate file", "duplicate mutation file", []reportFile{{FileName: "ordinary.go"}, {FileName: "ordinary.go"}}},
		{"duplicate identity", "duplicate mutation identity", []reportFile{{FileName: "ordinary.go", Mutations: []mutation{
			{Type: "NEGATION", Status: "KILLED", Line: 1, Column: 1},
			{Type: "NEGATION", Status: "KILLED", Line: 1, Column: 1},
		}}}},
	} {
		t.Run(test.name, func(t *testing.T) {
			data, err := json.Marshal(report{Files: test.files})
			if err != nil {
				t.Fatal(err)
			}
			got, err := ValidateReport(strings.NewReader(string(data)))
			if !errors.Is(err, ErrInvalid) || got != (ReportResult{}) {
				t.Fatal("rejected report must retain ErrInvalid and zero accounting")
			}
			if err.Error() != "invalid mutation evidence: "+test.want {
				t.Fatalf("default report diagnostic = %q; want categorical failure", err.Error())
			}
		})
	}
}

type ordinaryBootstrapReaderFailure struct{ closed bool }

func (*ordinaryBootstrapReaderFailure) Read([]byte) (int, error) {
	return 0, errors.New("ordinary reader diagnostic")
}

func (reader *ordinaryBootstrapReaderFailure) Close() error {
	reader.closed = true
	return nil
}

func TestOrdinaryBootstrapReaderErrorsAreCategorical(t *testing.T) {
	reader := &ordinaryBootstrapReaderFailure{}
	got, err := ValidateReport(reader)
	if !errors.Is(err, ErrInvalid) || got != (ReportResult{}) || err.Error() != "invalid mutation evidence: read mutation report" {
		t.Fatalf("report reader failure = %#v, %v; want zero and category", got, err)
	}
	checkpoint, err := readCheckpoint(1, func() (io.ReadCloser, error) { return reader, nil })
	if !errors.Is(err, ErrInvalid) || !reflect.DeepEqual(checkpoint, Checkpoint{}) || !reader.closed || err.Error() != "invalid mutation evidence: read checkpoint" {
		t.Fatalf("checkpoint reader failure = %#v, %v, closed %v; want zero and category", checkpoint, err, reader.closed)
	}
	checkpoint, err = readCheckpoint(1, func() (io.ReadCloser, error) { return nil, errors.New("ordinary open diagnostic") })
	if !errors.Is(err, ErrInvalid) || !reflect.DeepEqual(checkpoint, Checkpoint{}) || err.Error() != "invalid mutation evidence: open checkpoint" {
		t.Fatalf("checkpoint open failure = %#v, %v; want zero and category", checkpoint, err)
	}
}

func TestOrdinaryBootstrapKilledReportPreservesAccounting(t *testing.T) {
	const data = `{"files":[{"file_name":"ordinary.go","mutations":[{"type":"NEGATION","status":"KILLED","line":1,"column":1}]}]}`
	got, err := ValidateReport(strings.NewReader(data))
	if err != nil || got.Mutants != 1 || got.Killed != 1 || got.Equivalent != 0 || got.Digest != "sha256:42ab769afeeb17feef7af1c15681b94816fb65f1e1397f3086d83099a48d165b" {
		t.Fatalf("ordinary killed report = %#v, %v; want complete accepted accounting", got, err)
	}
}
