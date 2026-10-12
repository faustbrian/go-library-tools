// Package config strictly decodes bounded repository-owned golib policy.
// Load admits .golib.yaml; Decode handles immutable blobs with the same defaults.
// Values own their parsed collections but remain mutable caller data: keep them
// stable while orchestration reads. This package selects typed recipes, not a
// shell executor, release discovery service or source of runtime readiness.
package config

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"regexp/syntax"
	"strconv"
	"strings"
	"time"

	"github.com/faustbrian/go-library-tools/v2/internal/repositoryfile"
	"go.yaml.in/yaml/v3"
)

const (
	// MaximumSize bounds untrusted configuration input.
	MaximumSize = 1 << 20
	fileName    = ".golib.yaml"
	// MaximumLocalTestBudget bounds declared fast risk checks per invocation.
	MaximumLocalTestBudget = 3 * time.Minute
)

var (
	// ErrInvalid identifies a configuration that is syntactically valid YAML
	// but violates the supported contract.
	ErrInvalid = errors.New("invalid golib configuration")
	// ErrTooLarge identifies configuration input over MaximumSize.
	ErrTooLarge  = repositoryfile.ErrTooLarge
	versionRE    = regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+$`)
	makeTargetRE = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:/-]*$`)
)

// Config is the complete repository-specific policy. Facts already owned by
// canonical manifests are referenced here, not repeated.
type Config struct {
	// SchemaVersion must be 1 for the supported strict policy format.
	SchemaVersion int `json:"schema_version" yaml:"schema_version"`
	// ToolVersion is the required exact CLI release; development source builds are
	// handled separately by buildinfo.
	ToolVersion string `json:"tool_version" yaml:"tool_version"`
	// ToolChecksumsSHA256 optionally pins the release checksum manifest digest;
	// decoding checks syntax, not downloaded authenticity. Omission preserves
	// older supported policies.
	ToolChecksumsSHA256 string `json:"tool_checksums_sha256,omitempty" yaml:"tool_checksums_sha256,omitempty"`
	// Manifests selects canonical repository-relative inventory owners, defaulting
	// to modules.json and packages.json.
	Manifests Manifests `json:"manifest" yaml:"manifest,omitempty"`
	// Evidence selects the evidence root, defaulting to .verification.
	Evidence Evidence `json:"evidence" yaml:"evidence,omitempty"`
	// Mutation selects report storage, worker admission and explicit historical
	// imports.
	Mutation Mutation `json:"mutation" yaml:"mutation,omitempty"`
	// Coverage overrides statement acceptance only for named enabled modules;
	// omission remains exact.
	Coverage Coverage `json:"coverage,omitzero" yaml:"coverage,omitempty"`
	// API selects explicit enabled module compatibility baselines without arbitrary
	// execution hooks.
	API API `json:"api,omitzero" yaml:"api,omitempty"`
	// Runtimes declares supported pinned non-Go test executables, not shell policy.
	Runtimes Runtimes `json:"runtimes" yaml:"runtimes,omitempty"`
	// Operations is an ordered set of unique module/gate recipes validated before
	// command execution.
	Operations []Operation `json:"operations,omitempty" yaml:"operations,omitempty"`
}

// Manifests names the canonical repository inventory files.
type Manifests struct {
	// Modules names the repository-relative module catalog; default is modules.json.
	Modules string `json:"modules" yaml:"modules,omitempty"`
	// Packages names the repository-relative package classifications; default is
	// packages.json.
	Packages string `json:"packages" yaml:"packages,omitempty"`
}

// Evidence identifies the repository-owned verification evidence root.
type Evidence struct {
	// Root is the repository-relative evidence directory; path admission is distinct
	// from report validity.
	Root string `json:"root" yaml:"root,omitempty"`
}

// Mutation identifies repository-owned reports, checkpoints, and reviews.
type Mutation struct {
	// Root is the repository-relative mutation evidence directory.
	Root string `json:"root" yaml:"root,omitempty"`
	// MaxWorkers admits 0–4 workers; zero delegates the execution default rather
	// than disabling mutation.
	MaxWorkers MutationWorkers `json:"max_workers,omitempty" yaml:"max_workers,omitempty"`
	// Imports declares unique module-specific legacy archives and ledgers for
	// explicit import.
	Imports []MutationImport `json:"imports,omitempty" yaml:"imports,omitempty"`
}

// MutationWorkers preserves integer-only worker admission in YAML as in JSON.
type MutationWorkers int

// UnmarshalYAML rejects floating-point and string coercion before admission.
func (workers *MutationWorkers) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.ScalarNode || node.Tag != "!!int" {
		return fmt.Errorf("%w: mutation.max_workers must be an integer", ErrInvalid)
	}
	var value int
	if err := node.Decode(&value); err != nil {
		return err
	}
	*workers = MutationWorkers(value)
	return nil
}

// MutationImport identifies one repository-owned approved legacy checkpoint
// set that is materialized into content-addressed evidence before verification.
type MutationImport struct {
	// Module is a canonical module directory, with . naming the root.
	Module string `json:"module" yaml:"module"`
	// Archive is the repository-relative historical report archive, not an arbitrary
	// download.
	Archive string `json:"archive" yaml:"archive"`
	// Ledger is the repository-relative reviewed import ledger.
	Ledger string `json:"ledger" yaml:"ledger"`
}

// Coverage selects explicit module-level statement acceptance. Omission keeps
// exact coverage; evidence mode collects counts without certifying risk coverage.
type Coverage struct {
	// Modules is a unique module-specific acceptance list, not an ecosystem coverage
	// score.
	Modules []CoverageModule `json:"modules,omitempty" yaml:"modules,omitempty"`
}

// CoverageModule changes acceptance only for its named canonical module.
type CoverageModule struct {
	// Module names an inventoried module with coverage enabled.
	Module string `json:"module" yaml:"module"`
	// Mode is exact or evidence; evidence records execution without claiming full
	// coverage.
	Mode string `json:"mode" yaml:"mode"`
}

// API identifies module-specific exported API baseline formats and locations.
type API struct {
	// Baselines is a unique module-specific baseline list; it cannot share ownership
	// with a custom API operation.
	Baselines []APIBaseline `json:"baselines,omitempty" yaml:"baselines,omitempty"`
}

// APIBaseline defines one module's repository-owned compatibility baseline.
type APIBaseline struct {
	// Module names an inventoried module with API compatibility enabled.
	Module string `json:"module" yaml:"module"`
	// Mode is apidiff export data or go-doc complete rendered prose.
	Mode string `json:"mode" yaml:"mode"`
	// Path is the repository-relative baseline file; publication uses the API gate's
	// admitted directory.
	Path string `json:"path" yaml:"path"`
	// Packages is a nonempty ordered nonrecursive local package list for go-doc;
	// forbidden for apidiff.
	Packages []string `json:"packages,omitempty" yaml:"packages,omitempty"`
}

// Runtimes declares exact non-Go executables required by repository tests.
// Node is a tooling-owned documentation runtime and is not repository policy.
type Runtimes struct {
	// Deno optionally pins an exact supported MAJOR.MINOR.PATCH runtime.
	Deno string `json:"deno,omitempty" yaml:"deno,omitempty"`
	// Zsh optionally selects the supported 5.9 shell; it does not allow arbitrary
	// versions.
	Zsh string `json:"zsh,omitempty" yaml:"zsh,omitempty"`
}

// Operation replaces one package-specific legacy Make target with typed steps.
type Operation struct {
	// Module is the canonical module directory whose gate owns the recipe.
	Module string `json:"module" yaml:"module"`
	// Gate is one supported opt-in gate; local-test has stricter bounded selector
	// rules.
	Gate string `json:"gate" yaml:"gate"`
	// Steps is an ordered nonempty command recipe, not shell text.
	Steps []Step `json:"steps" yaml:"steps"`
}

// Step is a constrained operation with no shell or arbitrary executable hook.
type Step struct {
	// Type is go-test or make; mixed command-specific fields are refused.
	Type string `json:"type" yaml:"type"`
	// Makefile is a repository-local file required only for a make step.
	Makefile string `json:"makefile,omitempty" yaml:"makefile,omitempty"`
	// Target is a validated non-flag make target, not a shell expression.
	Target string `json:"target,omitempty" yaml:"target,omitempty"`
	// Packages is a local Go package pattern list; omitted go-test defaults to
	// ./....
	Packages []string `json:"packages,omitempty" yaml:"packages,omitempty"`
	// Run is an optional Go test selector, mutually exclusive with Benchmark and
	// Fuzz.
	Run string `json:"run,omitempty" yaml:"run,omitempty"`
	// Benchmark is an optional benchmark selector, mutually exclusive with Run and
	// Fuzz.
	Benchmark string `json:"benchmark,omitempty" yaml:"benchmark,omitempty"`
	// Fuzz is an optional fuzz target selector, mutually exclusive with Run and
	// Benchmark.
	Fuzz string `json:"fuzz,omitempty" yaml:"fuzz,omitempty"`
	// Budget is positive duration or iteration count for benchmark/fuzz; defaults
	// depend on selector.
	Budget string `json:"budget,omitempty" yaml:"budget,omitempty"`
	// Count is the go-test repetition count; omitted zero defaults to one before
	// validation.
	Count int `json:"count,omitempty" yaml:"count,omitempty"`
	// Race selects the Go race detector only for test runs, never
	// benchmark/fuzz/make.
	Race bool `json:"race,omitempty" yaml:"race,omitempty"`
	// Timeout is a positive parsed duration; omission defaults to 20m, while
	// local-test totals are capped.
	Timeout string `json:"timeout" yaml:"timeout,omitempty"`
}

// Load reads .golib.yaml from root, rejects unknown fields, applies stable
// defaults, and validates every repository-relative path.
func Load(root string) (Config, error) {
	data, err := repositoryfile.Read(root, fileName, MaximumSize)
	if err != nil {
		return Config{}, fmt.Errorf("load %s: %w", fileName, err)
	}
	return Decode(data)
}

// Decode applies the same bounded, strict policy contract to immutable Git
// blobs as Load applies to repository-owned files.
func Decode(data []byte) (Config, error) {
	if len(data) > MaximumSize {
		return Config{}, ErrTooLarge
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
	if value.ToolChecksumsSHA256 != "" && !isLowerHexDigest(value.ToolChecksumsSHA256) {
		return fmt.Errorf("%w: tool_checksums_sha256 must be exactly 64 lowercase hexadecimal characters", ErrInvalid)
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
			return fmt.Errorf("%w: %s: %s", ErrInvalid, path.name, err.Error())
		}
	}
	if value.Runtimes.Deno != "" && !versionRE.MatchString("v"+value.Runtimes.Deno) {
		return fmt.Errorf("%w: runtimes.deno must be an exact MAJOR.MINOR.PATCH version", ErrInvalid)
	}
	if value.Runtimes.Zsh != "" && value.Runtimes.Zsh != "5.9" {
		return fmt.Errorf("%w: runtimes.zsh supports only 5.9", ErrInvalid)
	}
	if value.Mutation.MaxWorkers < 0 || value.Mutation.MaxWorkers > 4 {
		return fmt.Errorf("%w: mutation.max_workers must be between 0 and 4", ErrInvalid)
	}
	seen := make(map[string]struct{}, len(value.Operations))
	mutationImports := make(map[string]struct{}, len(value.Mutation.Imports))
	for index, migration := range value.Mutation.Imports {
		if migration.Module == "" || (migration.Module != "." && validateRelativePath(migration.Module) != nil) {
			return fmt.Errorf("%w: mutation.imports[%d].module is invalid", ErrInvalid, index)
		}
		for _, path := range []struct{ name, value string }{{"archive", migration.Archive}, {"ledger", migration.Ledger}} {
			if err := validateRelativePath(path.value); err != nil || filepath.Clean(path.value) == "." {
				return fmt.Errorf("%w: mutation.imports[%d].%s is invalid", ErrInvalid, index, path.name)
			}
		}
		if _, exists := mutationImports[migration.Module]; exists {
			return fmt.Errorf("%w: mutation.imports[%d]: duplicate module", ErrInvalid, index)
		}
		mutationImports[migration.Module] = struct{}{}
	}
	apiModules := make(map[string]struct{}, len(value.API.Baselines))
	coverageModules := make(map[string]struct{}, len(value.Coverage.Modules))
	for index, acceptance := range value.Coverage.Modules {
		if acceptance.Module == "" || (acceptance.Module != "." && validateRelativePath(acceptance.Module) != nil) {
			return fmt.Errorf("%w: coverage.modules[%d].module is invalid", ErrInvalid, index)
		}
		if acceptance.Mode != "exact" && acceptance.Mode != "evidence" {
			return fmt.Errorf("%w: coverage.modules[%d].mode must be exact or evidence", ErrInvalid, index)
		}
		if _, exists := coverageModules[acceptance.Module]; exists {
			return fmt.Errorf("%w: coverage.modules[%d]: duplicate module", ErrInvalid, index)
		}
		coverageModules[acceptance.Module] = struct{}{}
	}
	for index, baseline := range value.API.Baselines {
		if err := baseline.validate(); err != nil {
			return fmt.Errorf("%w: api.baselines[%d]: %s", ErrInvalid, index, err.Error())
		}
		if _, exists := apiModules[baseline.Module]; exists {
			return fmt.Errorf("%w: api.baselines[%d]: duplicate module", ErrInvalid, index)
		}
		apiModules[baseline.Module] = struct{}{}
	}
	for index, operation := range value.Operations {
		if err := operation.validate(); err != nil {
			return fmt.Errorf("%w: operations[%d]: %s", ErrInvalid, index, err.Error())
		}
		identity := operation.Module + "\x00" + operation.Gate
		if _, exists := seen[identity]; exists {
			return fmt.Errorf("%w: operations[%d]: duplicate module and gate", ErrInvalid, index)
		}
		seen[identity] = struct{}{}
	}
	return nil
}

func isLowerHexDigest(value string) bool {
	if len(value) != 64 {
		return false
	}
	for _, character := range value {
		if !strings.ContainsRune("0123456789abcdef", character) {
			return false
		}
	}
	return true
}

func (baseline APIBaseline) validate() error {
	if baseline.Module == "" {
		return errors.New("module is required")
	}
	if err := validateRelativePath(baseline.Module); err != nil && baseline.Module != "." {
		return fmt.Errorf("module: %w", err)
	}
	if err := validateRelativePath(baseline.Path); err != nil {
		return fmt.Errorf("path: %w", err)
	}
	if filepath.Clean(baseline.Path) == "." {
		return errors.New("path must identify a repository file")
	}
	switch baseline.Mode {
	case "apidiff":
		if len(baseline.Packages) != 0 {
			return errors.New("apidiff mode does not accept packages")
		}
	case "go-doc":
		if len(baseline.Packages) == 0 {
			return errors.New("go-doc mode requires packages")
		}
		for _, packagePattern := range baseline.Packages {
			if err := validatePackagePattern(packagePattern); err != nil {
				return fmt.Errorf("go-doc package %q: %w", packagePattern, err)
			}
			if strings.HasSuffix(packagePattern, "/...") {
				return fmt.Errorf("go-doc package %q: recursive wildcards are not supported", packagePattern)
			}
		}
	default:
		return fmt.Errorf("unsupported mode %q", baseline.Mode)
	}
	return nil
}

func (operation Operation) validate() error {
	if operation.Module == "" {
		return errors.New("module is required")
	}
	if err := validateRelativePath(operation.Module); err != nil && operation.Module != "." {
		return fmt.Errorf("module: %w", err)
	}
	allowedGates := map[string]bool{
		"api": true, "benchmark": true, "conformance": true, "docs": true,
		"fuzz": true, "interoperability": true, "test": true, "local-test": true,
	}
	if !allowedGates[operation.Gate] {
		return fmt.Errorf("gate %q does not allow custom operations", operation.Gate)
	}
	if len(operation.Steps) == 0 {
		return errors.New("steps must not be empty")
	}
	if operation.Gate == "local-test" {
		_, err := operation.LocalTestBudget()
		return err
	}
	for index, step := range operation.Steps {
		if err := step.validate(); err != nil {
			return fmt.Errorf("steps[%d]: %w", index, err)
		}
		if operation.Gate == "test" && (step.Benchmark != "" || step.Fuzz != "") {
			return fmt.Errorf("steps[%d]: test operations accept only run selectors", index)
		}
	}
	return nil
}

// LocalTestBudget validates an explicitly selected fast test operation and
// returns its combined command deadlines. Legacy campaign gates are unchanged.
func (operation Operation) LocalTestBudget() (time.Duration, error) {
	if operation.Gate != "local-test" || len(operation.Steps) == 0 || len(operation.Steps) > 8 {
		return 0, errors.New("local-test requires one through eight focused steps")
	}
	var budget time.Duration
	for index, step := range operation.Steps {
		if err := step.validate(); err != nil {
			return 0, fmt.Errorf("steps[%d]: %w", index, err)
		}
		if step.Type != "go-test" || len(step.Packages) != 1 || strings.Contains(step.Packages[0], "...") || step.Count != 1 || step.Fuzz != "" || step.Benchmark != "" || step.Budget != "" {
			return 0, errors.New("local-test requires one explicit package and one test execution; campaigns are not allowed")
		}
		if len(step.Run) > 256 || !strings.HasPrefix(step.Run, "^Test") || !strings.HasSuffix(step.Run, "$") || strings.Contains(step.Run, "/") {
			return 0, errors.New("local-test requires an anchored top-level Test selector")
		}
		selector, err := syntax.Parse(step.Run, syntax.Perl)
		if err != nil {
			return 0, errors.New("local-test selector is invalid")
		}
		// Textual prefix/suffix checks alone permit a top-level alternative
		// to escape Test scope. Require anchors around the entire expression
		// and a mandatory literal Test prefix before any grouped alternatives.
		if selector.Op != syntax.OpConcat || len(selector.Sub) < 3 || selector.Sub[0].Op != syntax.OpBeginText || selector.Sub[len(selector.Sub)-1].Op != syntax.OpEndText || selector.Sub[1].Op != syntax.OpLiteral || !strings.HasPrefix(string(selector.Sub[1].Rune), "Test") {
			return 0, errors.New("local-test requires whole-expression anchors and a mandatory Test prefix; group alternatives after Test")
		}
		timeout, _ := time.ParseDuration(step.Timeout)
		if timeout > MaximumLocalTestBudget-budget {
			return 0, errors.New("local-test budget exceeds three minutes")
		}
		budget += timeout
	}
	return budget, nil
}

func (step Step) validate() error {
	timeout, err := time.ParseDuration(step.Timeout)
	if err != nil {
		return fmt.Errorf("timeout: %w", err)
	}
	if timeout <= 0 {
		return errors.New("timeout must be positive")
	}
	if step.Count < 0 {
		return errors.New("count must not be negative")
	}
	switch step.Type {
	case "go-test":
		if step.Race && (step.Benchmark != "" || step.Fuzz != "") {
			return errors.New("race accepts only test run selectors")
		}
		if step.Makefile != "" {
			return errors.New("go-test does not accept makefile")
		}
		if step.Target != "" {
			return errors.New("go-test does not accept target")
		}
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
				return fmt.Errorf("budget: %w", err)
			}
		}
		for _, packagePattern := range step.Packages {
			if err := validatePackagePattern(packagePattern); err != nil {
				return fmt.Errorf("go-test package %q: %w", packagePattern, err)
			}
		}
	case "make":
		if step.Race {
			return errors.New("make does not accept race")
		}
		if step.Makefile == "" {
			return errors.New("makefile is required")
		}
		if err := validateRelativePath(step.Makefile); err != nil {
			return errors.New("makefile must be repository-local")
		}
		if filepath.Clean(step.Makefile) == "." {
			return errors.New("makefile must identify a repository-local file")
		}
		if !makeTargetRE.MatchString(step.Target) {
			return errors.New("target must be a non-flag Make target")
		}
		if len(step.Packages) != 0 {
			return errors.New("make does not accept packages")
		}
		if step.Run != "" {
			return errors.New("make does not accept run")
		}
		if step.Benchmark != "" {
			return errors.New("make does not accept benchmark")
		}
		if step.Fuzz != "" {
			return errors.New("make does not accept fuzz")
		}
		if step.Budget != "" {
			return errors.New("make does not accept budget")
		}
		if step.Count != 0 {
			return errors.New("make does not accept count")
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
	if iterationsText, found := strings.CutSuffix(value, "x"); found {
		iterations, err := strconv.ParseUint(iterationsText, 10, 64)
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
