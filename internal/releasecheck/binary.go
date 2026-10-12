package releasecheck

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"debug/buildinfo"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"runtime/debug"
	"slices"
	"strings"
	"time"

	"github.com/faustbrian/go-library-tools/v2/internal/archivecheck"
	"github.com/faustbrian/go-library-tools/v2/internal/repositoryfile"
)

//go:embed licenses.json licenses/*/*
var binaryLicenseFiles embed.FS

const maximumBinaryArchive = 256 << 20

type licenseFile struct {
	Path   string `json:"path"`
	Name   string `json:"name"`
	SHA256 string `json:"sha256"`
	Source string `json:"source"`
}

type licensedModule struct {
	Path    string        `json:"path"`
	Version string        `json:"version"`
	Sum     string        `json:"sum"`
	License string        `json:"license"`
	Files   []licenseFile `json:"files"`
}

type binaryLicenseManifest struct {
	SchemaVersion  int         `json:"schema_version"`
	BinaryPath     string      `json:"binary_path"`
	ProductLicense licenseFile `json:"product_license"`
	Runtime        struct {
		Version string        `json:"version"`
		License string        `json:"license"`
		Files   []licenseFile `json:"files"`
	} `json:"runtime"`
	Modules []licensedModule `json:"modules"`
}

// PackageBinary writes a new deterministic Tools distribution. License material
// is selected by the compiled binary's module identities, not go.mod extras.
// Unknown/replaced dependencies, toolchains and changed material fail before
// creating output. Existing archives are never overwritten.
func PackageBinary(ctx context.Context, binaryPath, licensePath, outputPath string) error {
	binary, err := readBinaryReleaseFile(ctx, binaryPath, 128<<20)
	if err != nil {
		return fmt.Errorf("read release binary: %w", err)
	}
	info, err := buildinfo.Read(bytes.NewReader(binary))
	if err != nil {
		return errors.New("release input has no Go binary build identity")
	}
	license, notices, err := binaryLicensePayload(info, binaryLicenseFiles)
	if err != nil {
		return err
	}
	productLicense, err := readBinaryReleaseFile(ctx, licensePath, 2<<20)
	if err != nil {
		return fmt.Errorf("read product license: %w", err)
	}
	if !bytes.Equal(productLicense, license) {
		return errors.New("product license differs from reviewed binary license mapping")
	}
	var archive bytes.Buffer
	compressed := gzip.NewWriter(&archive)
	writer := tar.NewWriter(compressed)
	for _, item := range []struct {
		name string
		mode int64
		body []byte
	}{{"LICENSE", 0o644, license}, {"THIRD_PARTY_NOTICES", 0o644, notices}, {"golib", 0o755, binary}} {
		if err := ctx.Err(); err != nil {
			return err
		}
		header := &tar.Header{Name: item.name, Mode: item.mode, Size: int64(len(item.body)), ModTime: time.Unix(0, 0), Format: tar.FormatUSTAR}
		if err := writer.WriteHeader(header); err != nil {
			return fmt.Errorf("write release archive header: %w", err)
		}
		if _, err := writer.Write(item.body); err != nil {
			return fmt.Errorf("write release archive content: %w", err)
		}
	}
	if err := errors.Join(writer.Close(), compressed.Close()); err != nil {
		return fmt.Errorf("finish release archive: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	directory, name := filepath.Split(outputPath)
	if directory == "" {
		directory = "."
	}
	root, err := os.OpenRoot(directory)
	if err != nil {
		return fmt.Errorf("open release archive directory: %w", err)
	}
	defer root.Close()
	output, err := root.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return fmt.Errorf("create new release archive: %w", err)
	}
	_, writeErr := output.Write(archive.Bytes())
	closeErr := output.Close()
	if err := errors.Join(writeErr, closeErr, ctx.Err()); err != nil {
		return errors.Join(err, root.Remove(name))
	}
	return nil
}

// VerifyBinaryArchive validates a bounded, unextracted Tools distribution and
// requires exact reviewed license/NOTICE bytes for its compiled dependency graph.
// It neither executes the binary nor accepts historical two-member archives.
func VerifyBinaryArchive(ctx context.Context, filename string) error {
	data, err := readBinaryReleaseFile(ctx, filename, maximumBinaryArchive)
	if err != nil {
		return fmt.Errorf("read binary distribution: %w", err)
	}
	return errors.Join(verifyBinaryArchive(data, binaryLicenseFiles), ctx.Err())
}

func readBinaryReleaseFile(ctx context.Context, filename string, maximum int64) ([]byte, error) {
	file, err := repositoryfile.OpenContext(ctx, filename, maximum)
	if err != nil {
		return nil, err
	}
	body, readErr := io.ReadAll(io.LimitReader(file, maximum+1))
	closeErr := file.Close()
	if err := errors.Join(readErr, closeErr); err != nil {
		return nil, err
	}
	if int64(len(body)) > maximum {
		return nil, repositoryfile.ErrTooLarge
	}
	return body, nil
}

func verifyBinaryArchive(data []byte, corpus fs.FS) error {
	if err := archivecheck.Validate(bytes.NewReader(data), archivecheck.Limits{Entries: 3, Bytes: maximumBinaryArchive, CompressedBytes: maximumBinaryArchive}); err != nil {
		return fmt.Errorf("validate binary distribution: %w", err)
	}
	compressed, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer compressed.Close()
	reader := tar.NewReader(compressed)
	members := make(map[string][]byte, 3)
	for {
		header, err := reader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}
		mode := int64(0o644)
		maximum := int64(2 << 20)
		switch header.Name {
		case "golib":
			mode, maximum = 0o755, 128<<20
		case "LICENSE", "THIRD_PARTY_NOTICES":
		default:
			return errors.New("unexpected binary distribution member")
		}
		if header.Typeflag != tar.TypeReg || header.Mode != mode || header.Size > maximum {
			return errors.New("invalid binary distribution member type, mode or size")
		}
		body, err := io.ReadAll(io.LimitReader(reader, maximum+1))
		if err != nil || int64(len(body)) > maximum {
			return errors.New("binary distribution member exceeds content bound")
		}
		members[header.Name] = body
	}
	if len(members) != 3 {
		return errors.New("binary distribution must retain binary, product license and third-party notices")
	}
	info, err := buildinfo.Read(bytes.NewReader(members["golib"]))
	if err != nil {
		return errors.New("binary distribution has no Go binary build identity")
	}
	license, notices, err := binaryLicensePayload(info, corpus)
	if err != nil {
		return err
	}
	if !bytes.Equal(members["LICENSE"], license) || !bytes.Equal(members["THIRD_PARTY_NOTICES"], notices) {
		return errors.New("binary distribution license or notices differ from reviewed source material")
	}
	return nil
}

func binaryLicensePayload(info *debug.BuildInfo, corpus fs.FS) ([]byte, []byte, error) {
	manifestBytes, err := fs.ReadFile(corpus, "licenses.json")
	if err != nil || len(manifestBytes) > 2<<20 {
		return nil, nil, errors.New("binary license manifest missing or oversized")
	}
	var manifest binaryLicenseManifest
	decoder := json.NewDecoder(bytes.NewReader(manifestBytes))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&manifest); err != nil {
		return nil, nil, errors.New("invalid binary license manifest")
	}
	if decoder.Decode(new(any)) != io.EOF || manifest.SchemaVersion != 1 || manifest.BinaryPath == "" || info.Path != manifest.BinaryPath || info.GoVersion != manifest.Runtime.Version {
		return nil, nil, errors.New("unsupported binary or Go runtime license identity")
	}
	cgoSettings := 0
	for _, setting := range info.Settings {
		if setting.Key == "CGO_ENABLED" {
			if setting.Value != "0" {
				return nil, nil, errors.New("binary licensing requires the reviewed CGO-disabled build")
			}
			cgoSettings++
		}
		if setting.Key == "GOEXPERIMENT" && setting.Value != "" {
			return nil, nil, errors.New("unsupported experimental runtime license identity")
		}
	}
	if cgoSettings != 1 {
		return nil, nil, errors.New("binary licensing requires the reviewed CGO-disabled build")
	}
	license, err := readLicenseFile(corpus, manifest.ProductLicense)
	if err != nil {
		return nil, nil, err
	}
	var notices bytes.Buffer
	_, _ = fmt.Fprintf(&notices, "Third-party material for %s\n\nRuntime: %s\nLicense: %s\n", manifest.BinaryPath, manifest.Runtime.Version, manifest.Runtime.License)
	if err := appendLicenseFiles(&notices, corpus, manifest.Runtime.License, manifest.Runtime.Files); err != nil {
		return nil, nil, err
	}
	mappings := make(map[string]licensedModule, len(manifest.Modules))
	for _, module := range manifest.Modules {
		if module.Path == "" || module.Version == "" || !strings.HasPrefix(module.Sum, "h1:") {
			return nil, nil, errors.New("incomplete compiled dependency license mapping")
		}
		if _, exists := mappings[module.Path]; exists {
			return nil, nil, errors.New("duplicate compiled dependency license mapping")
		}
		mappings[module.Path] = module
	}
	dependencies := slices.Clone(info.Deps)
	for _, dependency := range dependencies {
		if dependency == nil {
			return nil, nil, errors.New("missing compiled dependency license identity")
		}
	}
	slices.SortFunc(dependencies, func(a, b *debug.Module) int { return strings.Compare(a.Path, b.Path) })
	seen := make(map[string]bool, len(dependencies))
	for _, dependency := range dependencies {
		mapping, exists := mappings[dependency.Path]
		if !exists || dependency.Replace != nil || dependency.Version != mapping.Version || dependency.Sum != mapping.Sum || seen[dependency.Path] {
			return nil, nil, errors.New("unsupported compiled dependency license identity")
		}
		seen[dependency.Path] = true
		_, _ = fmt.Fprintf(&notices, "\nModule: %s@%s\nModule checksum: %s\nLicense: %s\n", mapping.Path, mapping.Version, mapping.Sum, mapping.License)
		if err := appendLicenseFiles(&notices, corpus, mapping.License, mapping.Files); err != nil {
			return nil, nil, err
		}
	}
	return license, notices.Bytes(), nil
}

func appendLicenseFiles(output *bytes.Buffer, corpus fs.FS, license string, files []licenseFile) error {
	if license == "" || len(files) == 0 {
		return errors.New("missing reviewed license material")
	}
	if !slices.Contains([]string{"MIT", "Apache-2.0", "MIT AND Apache-2.0", "BSD-3-Clause"}, license) {
		return errors.New("unsupported reviewed license mapping")
	}
	seen := make(map[string]bool, len(files))
	for _, file := range files {
		if seen[file.Name] {
			return errors.New("duplicate reviewed license material")
		}
		seen[file.Name] = true
		body, err := readLicenseFile(corpus, file)
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintf(output, "\n--- %s ---\nSource: %s\nSHA-256: %s\n\n", file.Name, file.Source, file.SHA256)
		_, _ = output.Write(body)
		_, _ = output.WriteString("\n--- end ---\n")
	}
	if !seen["LICENSE"] || (license == "MIT AND Apache-2.0" && (!seen["NOTICE"] || !seen["APACHE-2.0"])) {
		return errors.New("reviewed license mapping omits required license or notice material")
	}
	return nil
}

func readLicenseFile(corpus fs.FS, file licenseFile) ([]byte, error) {
	if !fs.ValidPath(file.Path) || !strings.HasPrefix(file.Path, "licenses/") || path.Base(file.Path) != file.Name || !strings.HasPrefix(file.Source, "https://") {
		return nil, errors.New("invalid reviewed license source mapping")
	}
	body, err := fs.ReadFile(corpus, file.Path)
	if err != nil || len(body) == 0 || len(body) > 2<<20 {
		return nil, errors.New("reviewed license material missing or oversized")
	}
	digest := sha256.Sum256(body)
	if hex.EncodeToString(digest[:]) != file.SHA256 {
		return nil, errors.New("reviewed license material checksum changed")
	}
	return body, nil
}
