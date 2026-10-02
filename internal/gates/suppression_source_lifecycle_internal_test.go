package gates

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestSuppressionSourceLifecycleCharacterization(t *testing.T) {
	for _, test := range []struct{ name, source, reason string }{
		{"reasoned directive", "package ordinary\n//#nosec G304 -- confined repository file\n", ""},
		{"broad directive", "package ordinary\n//#nosec -- broad\n", "exact rule IDs"},
		{"malformed source", "package\n", "expected 'IDENT'"},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			contents := map[string]string{"source.go": test.source, "unrelated.txt": "keep unrelated bytes"}
			for name, content := range contents {
				if err := os.WriteFile(filepath.Join(root, name), []byte(content), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			err := checkSecuritySuppressionsBounded(t.Context(), root, 2)
			if test.reason == "" {
				if err != nil {
					t.Fatal(err)
				}
			} else if err == nil || !strings.Contains(err.Error(), test.reason) {
				t.Fatalf("suppression source refusal: error=%v, want %q", err, test.reason)
			}
			for name, want := range contents {
				data, err := os.ReadFile(filepath.Join(root, name))
				if err != nil || string(data) != want {
					t.Fatal("suppression inspection changed source or unrelated bytes")
				}
			}
		})
	}
}

// Every fault is produced by a real confined filesystem operation. Collaborators
// are scoped to this read; no permission assumptions or deletion races are used.
func TestSuppressionSourceLifecycleFailures(t *testing.T) {
	for _, stage := range []string{"metadata", "metadata-size", "open", "read", "close", "read-close", "growth", "growth-close"} {
		t.Run(stage, func(t *testing.T) {
			rootPath := t.TempDir()
			const source = "package ordinary\n"
			const name = "source.go"
			path := filepath.Join(rootPath, name)
			unrelated := filepath.Join(rootPath, "unrelated.txt")
			for path, content := range map[string]string{path: source, unrelated: "keep unrelated bytes"} {
				if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			root, err := os.OpenRoot(rootPath)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = root.Close() })
			entries, err := os.ReadDir(rootPath)
			if err != nil {
				t.Fatal(err)
			}
			entry := entries[0] // source.go sorts before unrelated.txt.
			var input *os.File
			t.Cleanup(func() {
				if input != nil {
					_ = input.Close()
				}
			})
			var operationErr, closeErr error
			opens, closes := 0, 0
			var offset int64
			limit := int64(len(source))
			if stage == "metadata-size" {
				limit--
			}
			files := securitySuppressionSourceFiles{
				info: func(entry os.DirEntry) (os.FileInfo, error) {
					if stage == "metadata" {
						info, err := root.Stat("missing.go")
						operationErr = err
						return info, err
					}
					return entry.Info()
				},
				open: func(root *os.Root, relative string) (*os.File, error) {
					opens++
					if stage == "open" {
						file, err := root.Open("missing.go")
						operationErr = err
						return file, err
					}
					if strings.HasPrefix(stage, "growth") {
						// Synchronous fixture growth after metadata admission is not
						// an inspection effect and cannot race another operation.
						if err := os.WriteFile(path, []byte(source+"// added\n"), 0o600); err != nil {
							t.Fatal(err)
						}
					}
					var err error
					if stage == "read" {
						input, err = root.OpenFile(relative, os.O_WRONLY, 0)
					} else {
						input, err = root.Open(relative)
					}
					if err == nil && stage == "read-close" {
						if err := input.Close(); err != nil {
							t.Fatal(err)
						}
					}
					return input, err
				},
				close: func(file *os.File) error {
					closes++
					if stage != "read-close" {
						var err error
						offset, err = file.Seek(0, io.SeekCurrent)
						if err != nil {
							t.Fatal(err)
						}
					}
					closeErr = file.Close()
					if stage == "close" || stage == "growth-close" {
						// A second real close reports failure after releasing the
						// descriptor, without replacing a successful read result.
						closeErr = errors.Join(closeErr, file.Close())
					}
					return closeErr
				},
			}
			content, err := readSecuritySuppressionSource(root, name, entry, limit, files)
			if err == nil || content != nil {
				t.Fatalf("source lifecycle failure returned usable content: %q, %v", content, err)
			}
			wantOpens, wantCloses := 1, 1
			switch stage {
			case "metadata", "open":
				if _, ok := operationErr.(*os.PathError); !ok {
					t.Fatal("fixture did not produce a pointer-backed filesystem error")
				}
				if reflect.ValueOf(err) != reflect.ValueOf(operationErr) {
					t.Fatal("source operation error identity changed")
				}
				wantCloses = 0
				if stage == "metadata" {
					wantOpens = 0
				}
			case "metadata-size", "growth":
				if err.Error() != "security suppression source exceeds size limit: source.go" {
					t.Fatalf("source size refusal changed: %v", err)
				}
				if stage == "metadata-size" {
					wantOpens, wantCloses = 0, 0
				}
			case "read", "read-close":
				joined, ok := err.(interface{ Unwrap() []error })
				if !ok || len(joined.Unwrap()) == 0 {
					t.Fatal("source read error was not joined")
				}
				var readErr *os.PathError
				if !errors.As(joined.Unwrap()[0], &readErr) || readErr.Op != "read" || !errors.Is(err, readErr) {
					t.Fatal("source read error identity changed")
				}
				if stage == "read-close" {
					if _, ok := closeErr.(*os.PathError); !ok {
						t.Fatal("fixture did not produce a pointer-backed close error")
					}
					if len(joined.Unwrap()) != 2 || reflect.ValueOf(joined.Unwrap()[1]) != reflect.ValueOf(closeErr) {
						t.Fatal("simultaneous read and close errors were not retained in order")
					}
				}
			}
			if stage == "close" || stage == "read-close" || stage == "growth-close" {
				if closeErr == nil || !errors.Is(err, closeErr) || !errors.Is(err, os.ErrClosed) {
					t.Fatal("source close error identity or I/O precedence changed")
				}
			}
			if opens != wantOpens || closes != wantCloses {
				t.Fatalf("metadata/handle lifecycle changed: opens=%d closes=%d", opens, closes)
			}
			if input != nil {
				if _, err := input.Stat(); !errors.Is(err, os.ErrClosed) {
					t.Fatal("source failure retained an open handle")
				}
			}
			if strings.HasPrefix(stage, "growth") && offset != limit+1 {
				t.Fatalf("growth read exceeded or failed to reach bounded prefix: %d", offset)
			}
			wantSource := source
			if strings.HasPrefix(stage, "growth") {
				wantSource += "// added\n"
			}
			for path, want := range map[string]string{path: wantSource, unrelated: "keep unrelated bytes"} {
				data, err := os.ReadFile(path)
				if err != nil || string(data) != want {
					t.Fatal("source read changed original or unrelated bytes")
				}
			}
		})
	}
}

func TestSuppressionSourceLifecycleInclusiveLimit(t *testing.T) {
	const source = "package ordinary\n"
	rootPath := t.TempDir()
	if err := os.WriteFile(filepath.Join(rootPath, "source.go"), []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	root, err := os.OpenRoot(rootPath)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	entries, err := os.ReadDir(rootPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, limit := range []int64{int64(len(source)), int64(len(source) + 1)} {
		var input *os.File
		files := securitySuppressionSourceFiles{
			info: os.DirEntry.Info,
			open: func(root *os.Root, name string) (*os.File, error) {
				file, err := root.Open(name)
				input = file
				return file, err
			},
			close: (*os.File).Close,
		}
		content, err := readSecuritySuppressionSource(root, "source.go", entries[0], limit, files)
		if err != nil || string(content) != source {
			t.Fatalf("inclusive source limit changed: limit=%d content=%q error=%v", limit, content, err)
		}
		if _, err := input.Stat(); !errors.Is(err, os.ErrClosed) {
			t.Fatal("successful source read retained an open handle")
		}
	}
}

func TestSuppressionPreflightRefusesSymbolicGoSource(t *testing.T) {
	root := t.TempDir()
	const source = "package ordinary\n"
	target := filepath.Join(root, "ordinary.txt")
	if err := os.WriteFile(target, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "link.go")
	if err := os.Symlink("ordinary.txt", link); err != nil {
		t.Fatal(err)
	}
	err := checkSecuritySuppressionsBounded(t.Context(), root, 2)
	if err == nil || err.Error() != "security suppression source is not a regular file: link.go" {
		t.Fatalf("symbolic Go source was not refused: %v", err)
	}
	data, err := os.ReadFile(target)
	if err != nil || string(data) != source {
		t.Fatal("source admission changed the regular target")
	}
	if target, err := os.Readlink(link); err != nil || target != "ordinary.txt" {
		t.Fatal("source admission changed symbolic source identity")
	}
}
