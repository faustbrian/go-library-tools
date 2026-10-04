package securitydocs

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestSecurityMatrixReadCancellationPreservesTerminalRefusal(t *testing.T) {
	directory := t.TempDir()
	for name, data := range map[string]string{
		"risk-register.json":   `{"schema_version":1,"risks":[]}`,
		"security-matrix.json": `{"schema_version":1,"modules":[]}`,
	} {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(data), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	var reads []string
	var ownedRoot *os.Root
	err := validateAtContext(ctx, directory, time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC), func(root *os.Root, path string) ([]byte, error) {
		ownedRoot = root
		data, err := readSecurityDocumentContext(ctx, root, path)
		reads = append(reads, path)
		if err == nil && path == "security-matrix.json" {
			cancel()
		}
		return data, err
	})
	if !errors.Is(err, context.Canceled) || !slices.Equal(reads, []string{"risk-register.json", "security-matrix.json"}) {
		t.Fatal("completed matrix read lost cancellation refusal or changed the snapshot read sequence")
	}
	assertSecurityRootClosed(t, ownedRoot)
	if err := ValidateContext(t.Context(), directory); err != nil {
		t.Fatal("canceled validation changed ordinary documents or prevented later validation")
	}
}

func TestSecurityReaderFailurePreservesCauseCleanupAndRecovery(t *testing.T) {
	for _, failedPath := range []string{"risk-register.json", "security-matrix.json"} {
		t.Run(failedPath, func(t *testing.T) {
			directory := t.TempDir()
			failure := errors.New("ordinary reader failure")
			cause := &os.PathError{Op: "read", Path: "ordinary", Err: failure}
			var reads []string
			var ownedRoot *os.Root
			refuse := true
			read := securityDocumentReader(func(root *os.Root, path string) ([]byte, error) {
				ownedRoot = root
				reads = append(reads, path)
				if _, err := root.Stat("."); err != nil {
					t.Fatal("reader was dispatched without its admitted live root")
				}
				if refuse && path == failedPath {
					return nil, cause
				}
				if path == "risk-register.json" {
					return []byte(`{"schema_version":1,"risks":[]}`), nil
				}
				return []byte(`{"schema_version":1,"modules":[]}`), nil
			})
			now := time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)
			err := validateAtContext(t.Context(), directory, now, read)
			var retained *os.PathError
			wantReads := []string{"risk-register.json"}
			if failedPath == "security-matrix.json" {
				wantReads = append(wantReads, "security-matrix.json")
			}
			if !errors.Is(err, failure) || !errors.Is(err, cause) || !errors.As(err, &retained) || retained.Op != "read" || retained.Path != "ordinary" || !slices.Equal(reads, wantReads) {
				t.Fatal("reader failure lost its typed cause or dispatched later document acquisition")
			}
			assertSecurityRootClosed(t, ownedRoot)
			refuse, reads = false, nil
			if err := validateAtContext(t.Context(), directory, now, read); err != nil || !slices.Equal(reads, []string{"risk-register.json", "security-matrix.json"}) {
				t.Fatal("ordinary reader recovery reused refused validation state")
			}
			assertSecurityRootClosed(t, ownedRoot)
		})
	}
}

// This is the private typed-decoding contract. The public validators supply
// compatible fixed schemas and targets; this does not assert public reachability.
func TestSecurityDecodePrivateTargetRefusalRemainsSafe(t *testing.T) {
	const schema = `{"type":"object","properties":{"ordinary":{"type":"string"}},"additionalProperties":false}`
	data := []byte(`{"ordinary":"value"}`)
	value := 17
	if err := decodeDocument(data, schema, &value); err == nil || err.Error() != "security document cannot be decoded" || value != 17 {
		t.Fatal("incompatible private target lost its safe refusal or changed retained scalar state")
	}
	var admitted map[string]string
	if err := decodeDocument(data, schema, &admitted); err != nil || !reflect.DeepEqual(admitted, map[string]string{"ordinary": "value"}) {
		t.Fatal("compatible private target changed admitted ordinary data")
	}
}

// This syntactically valid JSON contains invalid private schema configuration.
// It has no references or external resources; public callers use fixed schemas.
func TestSecurityDecodePrivateSchemaConfigurationRetainsCause(t *testing.T) {
	value := 17
	err := decodeDocument([]byte(`{}`), `{"type":"ordinary"}`, &value)
	if err == nil || !strings.HasPrefix(err.Error(), "compile published security schema: ") || errors.Unwrap(err) == nil || value != 17 {
		t.Fatal("invalid private schema configuration lost its compile cause or dispatched typed decoding")
	}
}

func assertSecurityRootClosed(t *testing.T, root *os.Root) {
	t.Helper()
	if root == nil {
		t.Fatal("validation did not acquire its ordinary root")
	}
	if _, err := root.Stat("."); !errors.Is(err, os.ErrClosed) {
		t.Fatal("validation retained its acquired root after terminal return")
	}
}
