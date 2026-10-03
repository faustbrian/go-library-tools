package mutation

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

const maximumEquivalentInventorySize = 1 << 20

// EquivalentInventory describes exact, independently reviewed tool limitations.
// It does not change a native mutant's LIVED status or its report counters.
type EquivalentInventory struct {
	SchemaVersion int                `json:"schema_version"`
	Packages      []EquivalentReview `json:"packages"`
}

// EquivalentReview applies only to one source and verifier identity.
type EquivalentReview struct {
	ModuleDirectory        string               `json:"module_directory"`
	PackageDirectory       string               `json:"package_directory"`
	SourceDigest           string               `json:"source_digest"`
	GremlinsVersion        string               `json:"gremlins_version"`
	GremlinsVerifierSHA256 string               `json:"gremlins_verifier_sha256"`
	Mutations              []EquivalentMutation `json:"mutations"`
}

// EquivalentMutation identifies a survivor and the observable contract under
// which the native mutation is equivalent, not necessarily byte-identical.
type EquivalentMutation struct {
	FileName       string `json:"file_name"`
	Type           string `json:"type"`
	Line           int    `json:"line"`
	Column         int    `json:"column"`
	ContractDomain string `json:"contract_domain"`
	Reason         string `json:"reason"`
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
func (inventory EquivalentInventory) Review(module, pkg, source, version, verifier string) (*EquivalentReview, error) {
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
		selected.Mutations = append([]EquivalentMutation(nil), review.Mutations...)
		return &selected, nil
	}
	return nil, nil
}
