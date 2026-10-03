package mutation

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"

	"github.com/faustbrian/go-library-tools/v2/internal/repositoryfile"
)

const legacyPackagePathCompatibility = "legacy_package_path_compatibility"

// Historical execution is applicable only when every direct production file
// selected by the campaign resolves to the same target with the old algorithm.
// Root packages are unchanged even when their declaration differs from the
// module name; subpackages must not fall back to an ancestor or module root.
func legacyPackagePathsUnchanged(root, moduleDirectory, packageDirectory, modulePath, expectedSource string) error {
	return legacyPackagePathsUnchangedWithLimits(root, moduleDirectory, packageDirectory, modulePath, expectedSource, mutationSourceLimits())
}

func legacyPackagePathsUnchangedWithLimits(root, moduleDirectory, packageDirectory, modulePath, expectedSource string, limits sourceReadLimits) error {
	relativeDirectory := filepath.Join(moduleDirectory, packageDirectory)
	if err := repositoryfile.ValidateDirectory(root, relativeDirectory); err != nil {
		return fmt.Errorf("%w: inspect historical package directory: %w", ErrInputChanged, err)
	}
	entries, err := readSourceEntries(filepath.Join(root, relativeDirectory), limits.entries)
	if err != nil {
		if errors.Is(err, errSourceEntryLimit) {
			return fmt.Errorf("%w: historical package source exceeds file bound", ErrInputChanged)
		}
		return fmt.Errorf("%w: inspect historical package source: %w", ErrInputChanged, err)
	}
	target := modulePath
	if packageDirectory != "." {
		target += "/" + filepath.ToSlash(packageDirectory)
	}
	files := 0
	remaining := limits.total
	manifest := sha256.New()
	for _, entry := range entries {
		name := entry.Name()
		if filepath.Ext(name) != ".go" || strings.HasSuffix(name, "_test.go") {
			continue
		}
		data, err := readSourceFile(root, filepath.Join(relativeDirectory, name), limits.file, remaining)
		if err != nil {
			if errors.Is(err, errSourceTotalLimit) {
				return fmt.Errorf("%w: historical package source exceeds byte bound", ErrInputChanged)
			}
			return fmt.Errorf("%w: read historical package declaration: %w", ErrInputChanged, err)
		}
		remaining -= int64(len(data))
		// Use the very bytes parsed below to reproduce SourceDigest. A
		// second file read cannot bind the declaration proof against an ABA
		// source change between input collection and compatibility checking.
		digest := sha256.Sum256(data)
		_, _ = fmt.Fprintf(manifest, "%s  %s\n", hex.EncodeToString(digest[:]), filepath.ToSlash(filepath.Join(relativeDirectory, name)))
		file, err := parser.ParseFile(token.NewFileSet(), name, data, parser.PackageClauseOnly)
		if err != nil {
			return fmt.Errorf("%w: historical package declaration cannot be proven", ErrInputChanged)
		}
		// Reproduce Gremlins v0.6.0 pkgName with the module-relative
		// CallingDir and direct source filename, including suffix matching.
		directory := filepath.Dir(fmt.Sprintf("%s/%s", packageDirectory, name))
		legacy := modulePath
		for {
			if strings.HasSuffix(directory, file.Name.Name) {
				legacy = modulePath + "/" + filepath.ToSlash(directory)
				break
			}
			parent := filepath.Dir(directory)
			if parent == directory {
				break
			}
			directory = parent
		}
		if legacy != target {
			return fmt.Errorf("%w: historical verifier selected a different package; current execution required", ErrInputChanged)
		}
		files++
	}
	if files == 0 {
		return fmt.Errorf("%w: historical package has no provable production files", ErrInputChanged)
	}
	if hex.EncodeToString(manifest.Sum(nil)) != expectedSource {
		return fmt.Errorf("%w: historical declaration proof does not match mutation source", ErrInputChanged)
	}
	return nil
}
