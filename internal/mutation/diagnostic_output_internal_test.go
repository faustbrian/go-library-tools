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
