// Package config loads and validates the repository-owned golib policy.
package config

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

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
	SchemaVersion int         `json:"schema_version" yaml:"schema_version"`
	ToolVersion   string      `json:"tool_version" yaml:"tool_version"`
	Manifests     Manifests   `json:"manifest" yaml:"manifest,omitempty"`
	Evidence      Evidence    `json:"evidence" yaml:"evidence,omitempty"`
	Mutation      Mutation    `json:"mutation" yaml:"mutation,omitempty"`
	Runtimes      Runtimes    `json:"runtimes" yaml:"runtimes,omitempty"`
	Operations    []Operation `json:"operations,omitempty" yaml:"operations,omitempty"`
}

// Manifests names the canonical repository inventory files.
type Manifests struct {
	Modules  string `json:"modules" yaml:"modules,omitempty"`
	Packages string `json:"packages" yaml:"packages,omitempty"`
}

// Evidence identifies the repository-owned verification evidence root.
type Evidence struct {
	Root string `json:"root" yaml:"root,omitempty"`
}

// Mutation identifies repository-owned reports, checkpoints, and reviews.
type Mutation struct {
	Root string `json:"root" yaml:"root,omitempty"`
}

// Runtimes declares exact non-Go executables required by repository tests.
// Node is a tooling-owned documentation runtime and is not repository policy.
type Runtimes struct {
	Deno string `json:"deno,omitempty" yaml:"deno,omitempty"`
	Zsh  string `json:"zsh,omitempty" yaml:"zsh,omitempty"`
}

// Operation replaces one package-specific legacy Make target with typed steps.
type Operation struct {
	Module string `json:"module" yaml:"module"`
	Gate   string `json:"gate" yaml:"gate"`
	Steps  []Step `json:"steps" yaml:"steps"`
}

// Step is a constrained operation with no shell or arbitrary executable hook.
type Step struct {
	Type      string   `json:"type" yaml:"type"`
	Packages  []string `json:"packages,omitempty" yaml:"packages,omitempty"`
	Run       string   `json:"run,omitempty" yaml:"run,omitempty"`
	Benchmark string   `json:"benchmark,omitempty" yaml:"benchmark,omitempty"`
	Fuzz      string   `json:"fuzz,omitempty" yaml:"fuzz,omitempty"`
	Budget    string   `json:"budget,omitempty" yaml:"budget,omitempty"`
	Count     int      `json:"count,omitempty" yaml:"count,omitempty"`
	Timeout   string   `json:"timeout" yaml:"timeout,omitempty"`
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
	if value.Mutation.Root == "" {
		value.Mutation.Root = ".verification/mutation"
	}
	for operationIndex := range value.Operations {
		for stepIndex := range value.Operations[operationIndex].Steps {
			step := &value.Operations[operationIndex].Steps[stepIndex]
			if step.Timeout == "" {
				step.Timeout = "20m"
			}
			if step.Type == "go-test" {
				if len(step.Packages) == 0 {
					step.Packages = []string{"./..."}
				}
				if step.Count == 0 {
					step.Count = 1
				}
				if step.Budget == "" && step.Fuzz != "" {
					step.Budget = "10000x"
				}
				if step.Budget == "" && step.Benchmark != "" {
					step.Budget = "100ms"
				}
			}
		}
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
		{"mutation.root", value.Mutation.Root},
	}
	for _, path := range paths {
		if err := validateRelativePath(path.value); err != nil {
			return fmt.Errorf("%w: %s: %v", ErrInvalid, path.name, err)
		}
	}
	if value.Runtimes.Deno != "" && !versionRE.MatchString("v"+value.Runtimes.Deno) {
		return fmt.Errorf("%w: runtimes.deno must be an exact MAJOR.MINOR.PATCH version", ErrInvalid)
	}
	if value.Runtimes.Zsh != "" && value.Runtimes.Zsh != "5.9" {
		return fmt.Errorf("%w: runtimes.zsh supports only 5.9", ErrInvalid)
	}
	seen := make(map[string]struct{}, len(value.Operations))
	for index, operation := range value.Operations {
		if err := operation.validate(); err != nil {
			return fmt.Errorf("%w: operations[%d]: %v", ErrInvalid, index, err)
		}
		identity := operation.Module + "\x00" + operation.Gate
		if _, exists := seen[identity]; exists {
			return fmt.Errorf("%w: operations[%d]: duplicate module and gate", ErrInvalid, index)
		}
		seen[identity] = struct{}{}
	}
	return nil
}

func (operation Operation) validate() error {
	if operation.Module == "" {
		return errors.New("module is required")
	}
	if err := validateRelativePath(operation.Module); err != nil && operation.Module != "." {
		return fmt.Errorf("module: %v", err)
	}
	allowedGates := map[string]bool{
		"api": true, "benchmark": true, "conformance": true, "docs": true,
		"fuzz": true, "interoperability": true,
	}
	if !allowedGates[operation.Gate] {
		return fmt.Errorf("gate %q does not allow custom operations", operation.Gate)
	}
	if len(operation.Steps) == 0 {
		return errors.New("steps must not be empty")
	}
	for index, step := range operation.Steps {
		if err := step.validate(); err != nil {
			return fmt.Errorf("steps[%d]: %v", index, err)
		}
	}
	return nil
}

func (step Step) validate() error {
	timeout, err := time.ParseDuration(step.Timeout)
	if err != nil {
		return fmt.Errorf("timeout: %v", err)
	}
	if timeout <= 0 {
		return errors.New("timeout must be positive")
	}
	if step.Count < 0 {
		return errors.New("count must not be negative")
	}
	switch step.Type {
	case "go-test":
		selectors := 0
		for _, selector := range []string{step.Run, step.Benchmark, step.Fuzz} {
			if selector != "" {
				selectors++
			}
		}
		if selectors > 1 {
			return errors.New("go-test accepts only one run, benchmark, or fuzz selector")
		}
		if step.Budget != "" {
			if step.Benchmark == "" && step.Fuzz == "" {
				return errors.New("budget requires a benchmark or fuzz selector")
			}
			if err := validateBudget(step.Budget); err != nil {
				return fmt.Errorf("budget: %v", err)
			}
		}
		for _, packagePattern := range step.Packages {
			if err := validatePackagePattern(packagePattern); err != nil {
				return fmt.Errorf("go-test package %q: %v", packagePattern, err)
			}
		}
	default:
		return fmt.Errorf("unsupported step type %q", step.Type)
	}
	return nil
}

func validatePackagePattern(value string) error {
	if value == "." {
		return nil
	}
	if !strings.HasPrefix(value, "./") || value == "./" || strings.ContainsAny(value, "\\\x00") {
		return errors.New("must be . or a repository-local ./ path")
	}
	segments := strings.Split(strings.TrimPrefix(value, "./"), "/")
	for index, segment := range segments {
		if segment == "" || segment == "." || segment == ".." {
			return errors.New("must not contain empty, current, or parent path segments")
		}
		if segment == "..." && index != len(segments)-1 {
			return errors.New("recursive wildcard must be the final path segment")
		}
	}
	return nil
}

func validateBudget(value string) error {
	if strings.HasSuffix(value, "x") {
		iterations, err := strconv.ParseUint(strings.TrimSuffix(value, "x"), 10, 64)
		if err != nil || iterations == 0 {
			return errors.New("must be a positive duration or iteration count ending in x")
		}
		return nil
	}
	duration, err := time.ParseDuration(value)
	if err != nil || duration <= 0 {
		return errors.New("must be a positive duration or iteration count ending in x")
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
