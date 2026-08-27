package docscheck

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
)

const (
	lycheeVersion             = "0.24.2"
	maximumLycheeArchiveSize  = 128 << 20
	maximumLycheeExpandedSize = 256 << 20
	maximumLycheeBinarySize   = 64 << 20
	maximumLycheeEntries      = 256
)

// LycheeRelease identifies one checksum-pinned official binary archive.
type LycheeRelease struct {
	Version string
	Target  string
	URL     string
	SHA256  string
}

// LycheeReleaseFor returns the exact release artifact for a supported Go
// operating system and architecture pair.
func LycheeReleaseFor(goos, goarch string) (LycheeRelease, error) {
	platforms := map[string]struct {
		target   string
		checksum string
	}{
		"darwin/arm64": {"aarch64-apple-darwin", "c9d3740ea2d891854d37116c9fba840f37b6e7c89d330e7db84ac333631c4977"},
		"darwin/amd64": {"x86_64-apple-darwin", "887503a9cff667d322b8d0892b40bf49976eb9507af8483220a3706cdad55978"},
		"linux/arm64":  {"aarch64-unknown-linux-gnu", "91a7bd65685da41b90ccb9bc867a3d649a7818042dae04ff405e55a25bddee4c"},
		"linux/amd64":  {"x86_64-unknown-linux-gnu", "1f4e0ef7f6554a6ed33dd7ac144fb2e1bbed98598e7af973042fc5cd43951c9a"},
	}
	platform, ok := platforms[goos+"/"+goarch]
	if !ok {
		return LycheeRelease{}, fmt.Errorf("unsupported lychee platform: %s/%s", goos, goarch)
	}
	archive := "lychee-" + platform.target + ".tar.gz"
	return LycheeRelease{
		Version: lycheeVersion,
		Target:  platform.target,
		URL:     "https://github.com/lycheeverse/lychee/releases/download/lychee-v" + lycheeVersion + "/" + archive,
		SHA256:  platform.checksum,
	}, nil
}

// ExtractLychee verifies and narrowly reads the expected binary from an
// official archive. It never extracts repository-controlled archive paths.
func ExtractLychee(archive string, release LycheeRelease) ([]byte, error) {
	if len(release.SHA256) != sha256.Size*2 {
		return nil, errors.New("malformed lychee checksum")
	}
	if _, err := hex.DecodeString(release.SHA256); err != nil || strings.ToLower(release.SHA256) != release.SHA256 {
		return nil, errors.New("malformed lychee checksum")
	}
	info, err := os.Lstat(archive)
	if err != nil {
		return nil, fmt.Errorf("inspect lychee archive: %w", err)
	}
	if !info.Mode().IsRegular() {
		return nil, errors.New("lychee archive is not a regular file")
	}
	if !withinMaximum(info.Size(), maximumLycheeArchiveSize) {
		return nil, errors.New("lychee archive is too large")
	}
	compressed, err := os.ReadFile(archive)
	if err != nil {
		return nil, fmt.Errorf("read lychee archive: %w", err)
	}
	digest := sha256.Sum256(compressed)
	if hex.EncodeToString(digest[:]) != release.SHA256 {
		return nil, errors.New("lychee archive checksum mismatch")
	}
	reader, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		return nil, fmt.Errorf("decompress lychee archive: %w", err)
	}
	defer reader.Close()

	expected := "lychee-" + release.Target + "/lychee"
	tarReader := tar.NewReader(reader)
	var binary []byte
	var total int64
	for entries := 1; ; entries++ {
		header, err := tarReader.Next()
		if errors.Is(err, io.EOF) {
			if binary == nil {
				return nil, errors.New("lychee archive does not contain the expected binary")
			}
			return binary, nil
		}
		if err != nil {
			return nil, fmt.Errorf("read lychee archive header: %w", err)
		}
		if !withinMaximum(int64(entries), maximumLycheeEntries) {
			return nil, errors.New("lychee archive contains too many entries")
		}
		clean := path.Clean(header.Name)
		if path.IsAbs(header.Name) || clean == ".." || strings.HasPrefix(clean, "../") || strings.Contains(header.Name, `\`) {
			return nil, fmt.Errorf("unsafe archive path: %s", header.Name)
		}
		if header.Typeflag != tar.TypeDir && header.Typeflag != tar.TypeReg {
			return nil, fmt.Errorf("unsupported archive entry: %s", header.Name)
		}
		if header.Size < 0 || !withinMaximum(header.Size, maximumLycheeExpandedSize-total) {
			return nil, errors.New("lychee archive expanded contents are too large")
		}
		total += header.Size
		if clean == expected {
			if binary != nil {
				return nil, errors.New("duplicate lychee binary")
			}
			if header.Typeflag != tar.TypeReg {
				return nil, errors.New("lychee binary is not a regular file")
			}
			if !withinMaximum(header.Size, maximumLycheeBinarySize) {
				return nil, errors.New("lychee binary is too large")
			}
			binary, err = io.ReadAll(io.LimitReader(tarReader, header.Size+1))
			if err != nil {
				return nil, fmt.Errorf("read lychee binary: %w", io.ErrUnexpectedEOF)
			}
			continue
		}
		if _, err := io.CopyN(io.Discard, tarReader, header.Size); err != nil {
			return nil, fmt.Errorf("read archive entry %s: %w", header.Name, err)
		}
	}
}

func withinMaximum(value, maximum int64) bool { return value <= maximum }
