package mutation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDiagnosticReadFailureUsesOnlyBoundedNotice(t *testing.T) {
	var output bytes.Buffer
	reportFailedMutationCoordinates(&output, t.TempDir())
	if got := output.String(); got != "mutation diagnostic: native report unavailable or oversized\n" {
		t.Fatalf("unreadable report diagnostic = %q", got)
	}
}

func TestDiagnosticCoordinatesSkipKilledAndCapOutput(t *testing.T) {
	entries := []mutation{{Type: "comparison", Status: "KILLED", Line: 999, Column: 9}}
	for line := 1; line <= 17; line++ {
		entries = append(entries, mutation{Type: "comparison", Status: "LIVED", Line: line, Column: 1})
	}
	native := report{Files: []reportFile{{FileName: "ordinary.go", Mutations: entries}}}
	data, err := json.Marshal(native)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "ordinary.json")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	reportFailedMutationCoordinates(&output, path)
	var expected strings.Builder
	for line := 1; line <= 16; line++ {
		fmt.Fprintf(&expected, "mutation diagnostic: \"LIVED\" \"ordinary.go\" \"comparison\" %d:1\n", line)
	}
	expected.WriteString("mutation diagnostic: 1 additional non-killed coordinates omitted\n")
	if got := output.String(); got != expected.String() {
		t.Fatalf("bounded coordinate diagnostic = %q; want %q", got, expected.String())
	}
	after, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(after, data) {
		t.Fatalf("diagnostic modified its report: error %v", err)
	}
}

func TestDiagnosticCoordinateExactCountHasNoOmissionFooter(t *testing.T) {
	entries := make([]mutation, 16)
	var expected strings.Builder
	for index := range entries {
		entries[index] = mutation{Type: "comparison", Status: "LIVED", Line: index + 1, Column: 1}
		fmt.Fprintf(&expected, "mutation diagnostic: \"LIVED\" \"ordinary.go\" \"comparison\" %d:1\n", index+1)
	}
	data, err := json.Marshal(report{Files: []reportFile{{FileName: "ordinary.go", Mutations: entries}}})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "report.json")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	reportFailedMutationCoordinates(&output, path)
	if output.String() != expected.String() {
		t.Fatalf("exact coordinate limit = %q; want all coordinates without omission", output.String())
	}
}

func TestDiagnosticReportByteAdmissionAndCompleteDecode(t *testing.T) {
	const valid = `{"files":[{"file_name":"ordinary.go","mutations":[{"type":"comparison","status":"LIVED","line":1,"column":1}]}]}`
	for _, test := range []struct{ name, data, want string }{
		{"exact byte allowance", valid + strings.Repeat(" ", maximumCheckpointSize-len(valid)), "mutation diagnostic: \"LIVED\" \"ordinary.go\" \"comparison\" 1:1\n"},
		{"one extra byte", valid + strings.Repeat(" ", maximumCheckpointSize-len(valid)+1), "mutation diagnostic: native report unavailable or oversized\n"},
		{"malformed trailing content", valid + "!", "mutation diagnostic: native report malformed\n"},
		{"absent files", `{}`, "mutation diagnostic: native report malformed\n"},
	} {
		t.Run(test.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "report.json")
			if err := os.WriteFile(path, []byte(test.data), 0o600); err != nil {
				t.Fatal(err)
			}
			var output bytes.Buffer
			reportFailedMutationCoordinates(&output, path)
			if output.String() != test.want {
				t.Fatalf("complete bounded diagnostic = %q; want %q", output.String(), test.want)
			}
		})
	}
}
