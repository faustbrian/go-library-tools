//go:build releaseintegration

package releasecheck_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/faustbrian/go-library-tools/v2/internal/releasecheck"
)

// This release-boundary check reads a real distribution without executing it.
func TestBinaryArchiveRetainsDependencyAndRuntimeAttribution(t *testing.T) {
	filename := os.Getenv("GOLIB_LICENSE_ARCHIVE")
	if filename == "" {
		t.Fatal("GOLIB_LICENSE_ARCHIVE must name the candidate release archive")
	}
	file, err := os.Open(filename)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	compressed, err := gzip.NewReader(file)
	if err != nil {
		t.Fatal(err)
	}
	defer compressed.Close()
	reader := tar.NewReader(io.LimitReader(compressed, 256<<20))
	var notices string
	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if header.Name == "THIRD_PARTY_NOTICES" {
			body, err := io.ReadAll(io.LimitReader(reader, 2<<20))
			if err != nil {
				t.Fatal(err)
			}
			notices = string(body)
		}
	}
	for _, required := range []string{
		"github.com/santhosh-tekuri/jsonschema/v6@v6.0.3",
		"github.com/yuin/goldmark@v1.8.6",
		"go.yaml.in/yaml/v3@v3.0.5",
		"golang.org/x/mod@v0.41.0",
		"golang.org/x/text@v0.14.0",
		"go1.27.2", "Yusuke Inuzuka", "Canonical Ltd.", "The Go Authors", "Apache License",
	} {
		if !strings.Contains(notices, required) {
			t.Fatalf("binary distribution omits required license/attribution %q", required)
		}
	}
}

func TestBinaryArchivePackagingIsDeterministicAndPreservesExistingOutput(t *testing.T) {
	entries := distributionEntries(t)
	root := t.TempDir()
	binary := filepath.Join(root, "golib")
	license := filepath.Join(root, "LICENSE")
	if err := os.WriteFile(binary, entries["golib"], 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(license, entries["LICENSE"], 0o600); err != nil {
		t.Fatal(err)
	}
	first, second := filepath.Join(root, "first.tgz"), filepath.Join(root, "second.tgz")
	for _, output := range []string{first, second} {
		if err := releasecheck.PackageBinary(context.Background(), binary, license, output); err != nil {
			t.Fatal(err)
		}
		if err := releasecheck.VerifyBinaryArchive(context.Background(), output); err != nil {
			t.Fatal(err)
		}
	}
	a, _ := os.ReadFile(first)
	b, _ := os.ReadFile(second)
	if !bytes.Equal(a, b) {
		t.Fatal("same compiled binary produced different archive bytes")
	}
	if err := releasecheck.PackageBinary(context.Background(), binary, license, first); err == nil {
		t.Fatal("packaging overwrote an existing archive")
	}
	after, err := os.ReadFile(first)
	if err != nil || !bytes.Equal(after, a) {
		t.Fatal("existing archive changed on refusal")
	}
	if err := os.WriteFile(license, []byte("changed"), 0o600); err != nil {
		t.Fatal(err)
	}
	refused := filepath.Join(root, "refused.tgz")
	if err := releasecheck.PackageBinary(context.Background(), binary, license, refused); err == nil {
		t.Fatal("changed product license accepted")
	}
	if _, err := os.Stat(refused); !os.IsNotExist(err) {
		t.Fatal("changed license left an output artifact")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := releasecheck.PackageBinary(ctx, binary, license, refused); err == nil {
		t.Fatal("canceled packaging accepted")
	}
	if _, err := os.Stat(refused); !os.IsNotExist(err) {
		t.Fatal("canceled packaging left an output artifact")
	}
}

func TestBinaryArchiveRefusesAlteredMissingExtraAndLinkedMaterial(t *testing.T) {
	entries := distributionEntries(t)
	for _, change := range []string{"notice bytes", "product license", "missing notices", "extra member", "symlink"} {
		t.Run(change, func(t *testing.T) {
			var data bytes.Buffer
			compressed := gzip.NewWriter(&data)
			writer := tar.NewWriter(compressed)
			for _, name := range []string{"LICENSE", "THIRD_PARTY_NOTICES", "golib", "unexpected"} {
				body, exists := entries[name]
				if name == "unexpected" {
					exists = change == "extra member"
				}
				if !exists || (name == "THIRD_PARTY_NOTICES" && change == "missing notices") {
					continue
				}
				if (name == "THIRD_PARTY_NOTICES" && change == "notice bytes") || (name == "LICENSE" && change == "product license") {
					body = []byte("changed attribution")
				}
				mode := int64(0o644)
				if name == "golib" {
					mode = 0o755
				}
				header := &tar.Header{Name: name, Mode: mode, Size: int64(len(body))}
				if name == "THIRD_PARTY_NOTICES" && change == "symlink" {
					header.Typeflag, header.Linkname, header.Size = tar.TypeSymlink, "LICENSE", 0
					body = nil
				}
				if err := writer.WriteHeader(header); err != nil {
					t.Fatal(err)
				}
				if _, err := writer.Write(body); err != nil {
					t.Fatal(err)
				}
			}
			if err := writer.Close(); err != nil {
				t.Fatal(err)
			}
			if err := compressed.Close(); err != nil {
				t.Fatal(err)
			}
			filename := filepath.Join(t.TempDir(), "tampered.tgz")
			if err := os.WriteFile(filename, data.Bytes(), 0o600); err != nil {
				t.Fatal(err)
			}
			if err := releasecheck.VerifyBinaryArchive(context.Background(), filename); err == nil {
				t.Fatal("incomplete or altered distribution accepted")
			}
		})
	}
}

func distributionEntries(t *testing.T) map[string][]byte {
	t.Helper()
	filename := os.Getenv("GOLIB_LICENSE_ARCHIVE")
	if filename == "" {
		t.Fatal("GOLIB_LICENSE_ARCHIVE must name the candidate release archive")
	}
	if err := releasecheck.VerifyBinaryArchive(context.Background(), filename); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	compressed, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	defer compressed.Close()
	reader := tar.NewReader(compressed)
	entries := map[string][]byte{}
	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		entries[header.Name], err = io.ReadAll(reader)
		if err != nil {
			t.Fatal(err)
		}
	}
	return entries
}
