package mutation

import (
	"fmt"
	"io"
	"strings"
)

const maximumZeroInventorySize = 1 << 20

// ZeroInventory contains source- and verifier-bound reviews for packages where
// the complete mutation operator set finds no viable mutations.
type ZeroInventory struct {
	SchemaVersion int          `json:"schema_version"`
	Packages      []ZeroReview `json:"packages"`
}

// ZeroReview explains one exact zero-mutant result.
type ZeroReview struct {
	ModuleDirectory        string `json:"module_directory"`
	PackageDirectory       string `json:"package_directory"`
	SourceDigest           string `json:"source_digest"`
	GremlinsVersion        string `json:"gremlins_version"`
	GremlinsVerifierSHA256 string `json:"gremlins_verifier_sha256"`
	Reason                 string `json:"reason"`
}

// ParseZeroInventory strictly parses a bounded zero-mutant review inventory.
func ParseZeroInventory(reader io.Reader) (ZeroInventory, error) {
	data, err := io.ReadAll(io.LimitReader(reader, maximumZeroInventorySize+1))
	if err != nil {
		return ZeroInventory{}, fmt.Errorf("%w: read zero-mutant inventory: %v", ErrInvalid, err)
	}
	if len(data) > maximumZeroInventorySize {
		return ZeroInventory{}, fmt.Errorf("%w: zero-mutant inventory exceeds %d bytes", ErrInvalid, maximumZeroInventorySize)
	}
	var inventory ZeroInventory
	if err := decodeStrict(data, &inventory); err != nil {
		return ZeroInventory{}, err
	}
	if inventory.SchemaVersion != 1 {
		return ZeroInventory{}, fmt.Errorf("%w: zero-mutant schema_version must be 1", ErrInvalid)
	}
	seen := make(map[string]struct{}, len(inventory.Packages))
	for index, review := range inventory.Packages {
		if err := review.validate(); err != nil {
			return ZeroInventory{}, fmt.Errorf("%w: packages[%d]: %v", ErrInvalid, index, err)
		}
		identity := review.ModuleDirectory + "\x00" + review.PackageDirectory
		if _, exists := seen[identity]; exists {
			return ZeroInventory{}, fmt.Errorf("%w: packages[%d]: duplicate package identity", ErrInvalid, index)
		}
		seen[identity] = struct{}{}
	}
	return inventory, nil
}

func (review ZeroReview) validate() error {
	if !validRelative(review.ModuleDirectory) || !validRelative(review.PackageDirectory) {
		return fmt.Errorf("module_directory and package_directory must be safe relative paths")
	}
	if !digestRE.MatchString(review.SourceDigest) {
		return fmt.Errorf("source_digest is malformed")
	}
	if !digestRE.MatchString(review.GremlinsVerifierSHA256) {
		return fmt.Errorf("gremlins_verifier_sha256 is malformed")
	}
	if !versionRE.MatchString(review.GremlinsVersion) {
		return fmt.Errorf("gremlins_version is malformed")
	}
	reason := strings.TrimSpace(review.Reason)
	if len(reason) < 40 || strings.ContainsAny(reason, "\x00\r\n") {
		return fmt.Errorf("reason must be a detailed single-line explanation")
	}
	return nil
}

// Reviewed reports whether an exact source and verifier identity has a human
// zero-mutant review. Near matches deliberately fail closed.
func (inventory ZeroInventory) Reviewed(module, pkg, source, version, verifier string) bool {
	for _, review := range inventory.Packages {
		if review.ModuleDirectory == module && review.PackageDirectory == pkg &&
			review.SourceDigest == source && review.GremlinsVersion == version &&
			review.GremlinsVerifierSHA256 == verifier {
			return true
		}
	}
	return false
}
