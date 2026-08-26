// Package config loads and validates the repository-owned golib policy.
package config

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/faustbrian/go-library-tools/internal/repositoryfile"
	"go.yaml.in/yaml/v3"
)

const (
	// MaximumSize bounds untrusted configuration input.
	MaximumSize = 1 << 20
	fileName    = ".golib.yaml"
)

var (
	// ErrInvalid identifies a configuration that is syntactically valid YAML
	// but violates the supported contract.
	ErrInvalid = errors.New("invalid golib configuration")
	// ErrTooLarge identifies configuration input over MaximumSize.
	ErrTooLarge = repositoryfile.ErrTooLarge
	versionRE   = regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+$`)
)

// Config is the complete repository-specific policy. Facts already owned by
// canonical manifests are referenced here, not repeated.
type Config struct {
	SchemaVersion int       `yaml:"schema_version"`
	ToolVersion   string    `yaml:"tool_version"`
	Manifests     Manifests `yaml:"manifest,omitempty"`
	Evidence      Evidence  `yaml:"evidence,omitempty"`
}

// Manifests names the canonical repository inventory files.
type Manifests struct {
	Modules  string `yaml:"modules,omitempty"`
	Packages string `yaml:"packages,omitempty"`
}

// Evidence identifies the repository-owned verification evidence root.
type Evidence struct {
	Root string `yaml:"root,omitempty"`
}

// Load reads .golib.yaml from root, rejects unknown fields, applies stable
// defaults, and validates every repository-relative path.
func Load(root string) (Config, error) {
	data, err := repositoryfile.Read(root, fileName, MaximumSize)
	if err != nil {
		return Config{}, fmt.Errorf("load %s: %w", fileName, err)
	}

	var result Config
	decoder := yaml.NewDecoder(strings.NewReader(string(data)))
	decoder.KnownFields(true)
	if err := decoder.Decode(&result); err != nil {
		return Config{}, fmt.Errorf("decode %s: %w", fileName, err)
	}
	if err := ensureSingleDocument(decoder); err != nil {
		return Config{}, err
	}
	applyDefaults(&result)
	if err := result.validate(); err != nil {
		return Config{}, err
	}
	return result, nil
}

func ensureSingleDocument(decoder *yaml.Decoder) error {
	var extra any
	err := decoder.Decode(&extra)
	if errors.Is(err, io.EOF) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("decode trailing YAML document: %w", err)
	}
	return fmt.Errorf("%w: multiple YAML documents are not allowed", ErrInvalid)
}

func applyDefaults(value *Config) {
	if value.Manifests.Modules == "" {
		value.Manifests.Modules = "modules.json"
	}
	if value.Manifests.Packages == "" {
		value.Manifests.Packages = "packages.json"
	}
	if value.Evidence.Root == "" {
		value.Evidence.Root = ".verification"
	}
}

func (value Config) validate() error {
	if value.SchemaVersion != 1 {
		return fmt.Errorf("%w: schema_version must be 1", ErrInvalid)
	}
	if !versionRE.MatchString(value.ToolVersion) {
		return fmt.Errorf("%w: tool_version must be an exact vMAJOR.MINOR.PATCH release", ErrInvalid)
	}
	paths := []struct {
		name  string
		value string
	}{
		{"manifest.modules", value.Manifests.Modules},
		{"manifest.packages", value.Manifests.Packages},
		{"evidence.root", value.Evidence.Root},
	}
	for _, path := range paths {
		if err := validateRelativePath(path.value); err != nil {
			return fmt.Errorf("%w: %s: %v", ErrInvalid, path.name, err)
		}
	}
	return nil
}

func validateRelativePath(path string) error {
	if path == "" || filepath.IsAbs(path) {
		return errors.New("must be a non-empty repository-relative path")
	}
	clean := filepath.Clean(path)
	if clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return errors.New("must not escape the repository")
	}
	return nil
}
