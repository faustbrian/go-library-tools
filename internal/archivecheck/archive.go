// Package archivecheck validates untrusted gzip-compressed tar archives before extraction.
package archivecheck

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"math"
	"path"
	"strings"
)

// Limits bounds archive metadata and expanded content.
type Limits struct {
	Entries         int
	Bytes           int64
	CompressedBytes int64
}

// Validate rejects archive entries that could escape or change extraction semantics.
func Validate(source io.Reader, limits Limits) error {
	if limits.Entries <= 0 || limits.Bytes <= 0 {
		return errors.New("archive limits must be positive")
	}
	compressedSource := source
	var compressedLimit *io.LimitedReader
	if limits.CompressedBytes > 0 {
		compressedLimit = &io.LimitedReader{R: source, N: limitWithSentinel(limits.CompressedBytes)}
		compressedSource = compressedLimit
	}
	compressed, err := gzip.NewReader(compressedSource)
	if err != nil {
		if compressedLimit != nil && compressedLimit.N == 0 {
			return fmt.Errorf("compressed byte limit exceeded: %d", limits.CompressedBytes)
		}
		return fmt.Errorf("open gzip archive: %w", err)
	}
	defer compressed.Close()

	decompressed := &io.LimitedReader{R: compressed, N: limitWithSentinel(limits.Bytes)}
	reader := tar.NewReader(decompressed)
	seen := make(map[string]struct{})
	var entries int
	var expanded int64
	for {
		header, nextErr := reader.Next()
		if errors.Is(nextErr, io.EOF) {
			var trailing [1]byte
			count, err := decompressed.Read(trailing[:])
			if decompressed.N == 0 {
				return fmt.Errorf("expanded byte limit exceeded: %d", limits.Bytes)
			}
			if count > 0 {
				return errors.New("gzip archive contains trailing decompressed data")
			}
			if !errors.Is(err, io.EOF) {
				if compressedLimit != nil && compressedLimit.N == 0 {
					return fmt.Errorf("compressed byte limit exceeded: %d", limits.CompressedBytes)
				}
				return fmt.Errorf("finish gzip archive: %w", err)
			}
			if compressedLimit != nil && compressedLimit.N == 0 {
				return fmt.Errorf("compressed byte limit exceeded: %d", limits.CompressedBytes)
			}
			return nil
		}
		if nextErr != nil {
			if decompressed.N == 0 {
				return fmt.Errorf("expanded byte limit exceeded: %d", limits.Bytes)
			}
			if compressedLimit != nil && compressedLimit.N == 0 {
				return fmt.Errorf("compressed byte limit exceeded: %d", limits.CompressedBytes)
			}
			return fmt.Errorf("read tar archive: %w", nextErr)
		}
		entries++
		if entries > limits.Entries {
			return fmt.Errorf("archive entry limit exceeded: %d", limits.Entries)
		}
		normalized := strings.ReplaceAll(header.Name, "\\", "/")
		clean := path.Clean(normalized)
		if clean == "." || strings.HasPrefix(clean, "/") || clean == ".." || strings.HasPrefix(clean, "../") {
			return errors.New("unsafe archive path")
		}
		if strings.TrimSuffix(normalized, "/") != clean {
			return errors.New("non-canonical archive path")
		}
		if _, exists := seen[clean]; exists {
			return errors.New("duplicate archive path")
		}
		seen[clean] = struct{}{}
		if header.Typeflag != tar.TypeReg && header.Typeflag != 0 && header.Typeflag != tar.TypeDir {
			return errors.New("unsupported archive entry")
		}
		if header.Mode&0o7000 != 0 {
			return errors.New("unsafe archive mode")
		}
		if header.Size < 0 || header.Size > limits.Bytes-expanded {
			return fmt.Errorf("expanded byte limit exceeded: %d", limits.Bytes)
		}
		expanded += header.Size
	}
}

func limitWithSentinel(limit int64) int64 {
	if limit == math.MaxInt64 {
		return limit
	}
	return limit + 1
}
