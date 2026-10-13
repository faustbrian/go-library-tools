package mutation

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"
)

const maximumEquivalentInventorySize = 1 << 20

// EquivalentInventory describes exact, independently reviewed tool limitations.
// It does not change a native mutant's LIVED status or its report counters.
type EquivalentInventory struct {
	// SchemaVersion must be 1 for the closed review format.
	SchemaVersion int `json:"schema_version"`
	// Packages is an explicit unique module/package review list; empty admits no
	// exceptions.
	Packages []EquivalentReview `json:"packages"`
}

// EquivalentReview applies only to one source and verifier identity.
type EquivalentReview struct {
	architecture string
	// ModuleDirectory selects the exact repository-relative module.
	ModuleDirectory string `json:"module_directory"`
	// PackageDirectory selects the exact package within that module.
	PackageDirectory string `json:"package_directory"`
	// SourceDigest is the lowercase source-only SHA-256 requiring exact current
	// equality.
	SourceDigest string `json:"source_digest"`
	// GremlinsVersion is the exact engine release reviewed for these coordinates.
	GremlinsVersion string `json:"gremlins_version"`
	// GremlinsVerifierSHA256 is the exact semantic verifier digest without a prefix.
	GremlinsVerifierSHA256 string `json:"gremlins_verifier_sha256"`
	// Mutations is a nonempty unique coordinate set; Review returns its independent
	// slice copy.
	Mutations []EquivalentMutation `json:"mutations"`
}

// EquivalentMutation identifies a survivor and the observable contract under
// which the native mutation is equivalent, not necessarily byte-identical.
type EquivalentMutation struct {
	// FileName is the canonical relative Go source filename.
	FileName string `json:"file_name"`
	// Type is the exact native mutation operator identity.
	Type string `json:"type"`
	// Line is the positive native one-based source line.
	Line int `json:"line"`
	// Column is the positive native one-based source column.
	Column int `json:"column"`
	// ContractDomain names the reviewed observable domain in detailed single-line
	// prose.
	ContractDomain string `json:"contract_domain"`
	// Reason states the independently reviewed equivalence, not a blanket score
	// waiver.
	Reason string `json:"reason"`
	// Goarchs optionally limits this disposition to sorted unique supported
	// Go 1.27 target architectures. Absence preserves historical global scope.
	Goarchs targetArchitectures `json:"goarchs,omitempty"`
}

// targetArchitectures distinguishes absent legacy scope from explicit null.
type targetArchitectures []string

// UnmarshalJSON requires an array; null must not silently broaden scope.
func (architectures *targetArchitectures) UnmarshalJSON(data []byte) error {
	if strings.TrimSpace(string(data)) == "null" {
		return errors.New("architecture scope must not be null")
	}
	var values []string
	if err := json.Unmarshal(data, &values); err != nil {
		return err
	}
	*architectures = values
	return nil
}

// ParseEquivalentInventory strictly bounds and parses an optional review file.
func ParseEquivalentInventory(reader io.Reader) (EquivalentInventory, error) {
	data, err := io.ReadAll(io.LimitReader(reader, maximumEquivalentInventorySize+1))
	if err != nil {
		return EquivalentInventory{}, fmt.Errorf("%w: read equivalent-mutant inventory", ErrInvalid)
	}
	if len(data) > maximumEquivalentInventorySize {
		return EquivalentInventory{}, fmt.Errorf("%w: equivalent-mutant inventory exceeds %d bytes", ErrInvalid, maximumEquivalentInventorySize)
	}
	var inventory EquivalentInventory
	if err := decodeStrict(data, &inventory); err != nil {
		return EquivalentInventory{}, err
	}
	if inventory.SchemaVersion != 1 || inventory.Packages == nil {
		return EquivalentInventory{}, fmt.Errorf("%w: equivalent-mutant inventory requires schema_version 1 and packages", ErrInvalid)
	}
	seenPackages := make(map[string]struct{}, len(inventory.Packages))
	for index, review := range inventory.Packages {
		if err := review.validate(); err != nil {
			return EquivalentInventory{}, fmt.Errorf("%w: packages[%d]: %s", ErrInvalid, index, err.Error())
		}
		identity := review.ModuleDirectory + "\x00" + review.PackageDirectory
		if _, exists := seenPackages[identity]; exists {
			return EquivalentInventory{}, fmt.Errorf("%w: duplicate equivalent-mutant package", ErrInvalid)
		}
		seenPackages[identity] = struct{}{}
	}
	return inventory, nil
}

func (review EquivalentReview) validate() error {
	if !validRelative(review.ModuleDirectory) || !validRelative(review.PackageDirectory) ||
		!digestRE.MatchString(review.SourceDigest) || !digestRE.MatchString(review.GremlinsVerifierSHA256) ||
		!versionRE.MatchString(review.GremlinsVersion) || len(review.Mutations) == 0 {
		return errors.New("equivalent-mutant package identity or mutation selection is malformed")
	}
	seen := make(map[string]struct{}, len(review.Mutations))
	for _, candidate := range review.Mutations {
		if candidate.Goarchs != nil {
			if len(candidate.Goarchs) == 0 {
				return errors.New("equivalent-mutant architecture scope is empty")
			}
			for index, arch := range candidate.Goarchs {
				if !supportedArchitecture(arch) || index > 0 && candidate.Goarchs[index-1] >= arch {
					return errors.New("equivalent-mutant architectures must be canonical unique supported targets")
				}
			}
		}
		if !validRelative(candidate.FileName) || !strings.HasSuffix(candidate.FileName, ".go") ||
			candidate.Type == "" || strings.ContainsAny(candidate.Type, "\x00\r\n") ||
			candidate.Line <= 0 || candidate.Column <= 0 ||
			len(strings.TrimSpace(candidate.ContractDomain)) < 20 || strings.ContainsAny(candidate.ContractDomain, "\x00\r\n") ||
			len(strings.TrimSpace(candidate.Reason)) < 40 || strings.ContainsAny(candidate.Reason, "\x00\r\n") {
			return errors.New("equivalent-mutant coordinate, contract domain or reason is malformed")
		}
		identity := mutationIdentity(candidate.FileName, candidate.Type, candidate.Line, candidate.Column)
		if _, exists := seen[identity]; exists {
			return errors.New("duplicate equivalent-mutant coordinate")
		}
		seen[identity] = struct{}{}
	}
	return nil
}

func mutationIdentity(file, kind string, line, column int) string {
	return fmt.Sprintf("%s\x00%s\x00%d\x00%d", file, kind, line, column)
}

// Review fails closed when a selected package's source or tool identity drifts.
// A nil review means that the package follows the original all-KILLED policy.
func (inventory EquivalentInventory) Review(module, pkg, source, version, verifier string, architecture ...string) (*EquivalentReview, error) {
	if len(architecture) > 1 {
		return nil, fmt.Errorf("%w: one native target architecture is required", ErrInvalid)
	}
	arch := ""
	if len(architecture) == 1 {
		arch = architecture[0]
	}
	for _, review := range inventory.Packages {
		if review.ModuleDirectory != module || review.PackageDirectory != pkg {
			continue
		}
		if err := review.validate(); err != nil {
			return nil, fmt.Errorf("%w: invalid equivalent-mutant review: %s", ErrInvalid, err.Error())
		}
		if review.SourceDigest != source || review.GremlinsVersion != version || review.GremlinsVerifierSHA256 != verifier {
			return nil, fmt.Errorf("%w: equivalent-mutant source or verifier identity changed", ErrInvalid)
		}
		selected := review
		selected.architecture = arch
		selected.Mutations = nil
		for _, candidate := range review.Mutations {
			if candidate.Goarchs != nil {
				if !supportedArchitecture(arch) {
					return nil, fmt.Errorf("%w: scoped equivalence requires known native GOARCH", ErrInvalid)
				}
				if !slices.Contains(candidate.Goarchs, arch) {
					continue
				}
				candidate.Goarchs = slices.Clone(candidate.Goarchs)
			}
			selected.Mutations = append(selected.Mutations, candidate)
		}
		if len(selected.Mutations) == 0 {
			return nil, nil
		}
		return &selected, nil
	}
	return nil, nil
}

// supportedArchitecture lists actual Go 1.27 build targets, not historical
// filename suffixes for removed architectures.
func supportedArchitecture(arch string) bool {
	switch arch {
	case "386", "amd64", "arm", "arm64", "loong64", "mips", "mips64", "mips64le", "mipsle", "ppc64", "ppc64le", "riscv64", "s390x", "wasm":
		return true
	default:
		return false
	}
}

// validateExecution refuses a parsed but not runtime-admitted scoped review.
func (review EquivalentReview) validateExecution() error {
	if err := review.validate(); err != nil {
		return err
	}
	for _, candidate := range review.Mutations {
		if candidate.Goarchs != nil && (!supportedArchitecture(review.architecture) || !slices.Contains(candidate.Goarchs, review.architecture)) {
			return errors.New("scoped equivalence has no admitted native target architecture")
		}
	}
	return nil
}
