package mutation

import (
	"fmt"
	"io"
	"strings"
	"time"
)

const maximumMigrationLedgerSize = 4 << 20

// MigrationLedger binds legacy verifier and input identities to explicitly
// reviewed semantic replacements. Git revisions are matching data for the old
// records only; approved evidence emitted by the new tool excludes them.
type MigrationLedger struct {
	SchemaVersion           int                     `json:"schema_version"`
	Reason                  string                  `json:"reason"`
	VerifierMigrationReview VerifierMigrationReview `json:"verifier_migration_review"`
	VerifierMigrations      []VerifierMigration     `json:"verifier_migrations"`
	Entries                 []InputMigration        `json:"entries"`
}

// VerifierMigrationReview records the human approval for one verifier.
type VerifierMigrationReview struct {
	GremlinsVerifierSHA256 string `json:"gremlins_verifier_sha256"`
	Reason                 string `json:"reason"`
	ReviewedAt             string `json:"reviewed_at"`
}

// VerifierMigration proves which legacy checkpoint used the reviewed verifier.
type VerifierMigration struct {
	ExecutionRevision      string `json:"execution_revision"`
	GateInputDigest        string `json:"gate_input_digest"`
	GremlinsVerifierSHA256 string `json:"gremlins_verifier_sha256"`
	GremlinsVersion        string `json:"gremlins_version"`
	Module                 string `json:"module"`
	Package                string `json:"package"`
	ReportSHA256           string `json:"report_sha256"`
}

// InputMigration approves one exact replacement input identity.
type InputMigration struct {
	ExecutionRevision      string `json:"execution_revision"`
	GateInputDigest        string `json:"gate_input_digest"`
	ReplacementInputDigest string `json:"replacement_gate_input_digest,omitempty"`
	MigrationReason        string `json:"migration_reason,omitempty"`
	GremlinsVerifierSHA256 string `json:"gremlins_verifier_sha256,omitempty"`
	GremlinsVersion        string `json:"gremlins_version"`
	Module                 string `json:"module"`
	Package                string `json:"package"`
	ReportSHA256           string `json:"report_sha256"`
}

// ParseMigrationLedger strictly parses one bounded approval ledger.
func ParseMigrationLedger(reader io.Reader) (MigrationLedger, error) {
	data, err := io.ReadAll(io.LimitReader(reader, maximumMigrationLedgerSize+1))
	if err != nil {
		return MigrationLedger{}, fmt.Errorf("%w: read migration ledger: %v", ErrInvalid, err)
	}
	if len(data) > maximumMigrationLedgerSize {
		return MigrationLedger{}, fmt.Errorf("%w: migration ledger exceeds %d bytes", ErrInvalid, maximumMigrationLedgerSize)
	}
	var ledger MigrationLedger
	if err := decodeStrict(data, &ledger); err != nil {
		return MigrationLedger{}, err
	}
	if err := ledger.validate(); err != nil {
		return MigrationLedger{}, err
	}
	return ledger, nil
}

func (ledger MigrationLedger) validate() error {
	if ledger.SchemaVersion != 3 {
		return fmt.Errorf("%w: migration schema_version must be 3", ErrInvalid)
	}
	if err := detailedReason("reason", ledger.Reason); err != nil {
		return err
	}
	review := ledger.VerifierMigrationReview
	if !digestRE.MatchString(review.GremlinsVerifierSHA256) {
		return fmt.Errorf("%w: verifier_migration_review digest is malformed", ErrInvalid)
	}
	if err := detailedReason("verifier_migration_review.reason", review.Reason); err != nil {
		return err
	}
	if _, err := time.Parse("2006-01-02", review.ReviewedAt); err != nil {
		return fmt.Errorf("%w: verifier_migration_review.reviewed_at is malformed", ErrInvalid)
	}
	seenVerifier := make(map[string]struct{}, len(ledger.VerifierMigrations))
	for index, migration := range ledger.VerifierMigrations {
		if err := migration.validate(review.GremlinsVerifierSHA256); err != nil {
			return fmt.Errorf("%w: verifier_migrations[%d]: %v", ErrInvalid, index, err)
		}
		identity := verifierMigrationIdentity(migration)
		if _, exists := seenVerifier[identity]; exists {
			return fmt.Errorf("%w: verifier_migrations[%d]: duplicate approval", ErrInvalid, index)
		}
		seenVerifier[identity] = struct{}{}
	}
	seenInput := make(map[string]struct{}, len(ledger.Entries))
	for index, migration := range ledger.Entries {
		if err := migration.validate(review.GremlinsVerifierSHA256); err != nil {
			return fmt.Errorf("%w: entries[%d]: %v", ErrInvalid, index, err)
		}
		identity := inputMigrationIdentity(migration)
		if _, exists := seenInput[identity]; exists {
			return fmt.Errorf("%w: entries[%d]: duplicate approval", ErrInvalid, index)
		}
		seenInput[identity] = struct{}{}
	}
	return nil
}

func (migration VerifierMigration) validate(expectedVerifier string) error {
	if !revisionRE.MatchString(migration.ExecutionRevision) || !digestRE.MatchString(migration.GateInputDigest) ||
		!digestRE.MatchString(migration.ReportSHA256) || !versionRE.MatchString(migration.GremlinsVersion) ||
		migration.GremlinsVerifierSHA256 != expectedVerifier || !validRelative(migration.Module) || !validRelative(migration.Package) {
		return fmt.Errorf("verifier approval identity is malformed")
	}
	return nil
}

func (migration InputMigration) validate(expectedVerifier string) error {
	verifier := migration.GremlinsVerifierSHA256
	if verifier == "" {
		verifier = expectedVerifier
	}
	if !revisionRE.MatchString(migration.ExecutionRevision) || !digestRE.MatchString(migration.GateInputDigest) ||
		(migration.ReplacementInputDigest != "" && !digestRE.MatchString(migration.ReplacementInputDigest)) ||
		!digestRE.MatchString(migration.ReportSHA256) || !versionRE.MatchString(migration.GremlinsVersion) ||
		verifier != expectedVerifier || !validRelative(migration.Module) || !validRelative(migration.Package) {
		return fmt.Errorf("input approval identity is malformed")
	}
	if len(migration.MigrationReason) > 256 || strings.ContainsAny(migration.MigrationReason, "\x00\r\n") {
		return fmt.Errorf("input migration reason is malformed")
	}
	return nil
}

func detailedReason(name, reason string) error {
	trimmed := strings.TrimSpace(reason)
	if len(trimmed) < 40 || strings.ContainsAny(trimmed, "\x00\r\n") {
		return fmt.Errorf("%w: %s must be a detailed single-line explanation", ErrInvalid, name)
	}
	return nil
}

// Approve verifies the complete old checkpoint identity and, when necessary,
// its exact replacement input identity.
func (ledger MigrationLedger) Approve(checkpoint Checkpoint, currentInput, expectedVerifier string) error {
	if !digestRE.MatchString(currentInput) {
		return fmt.Errorf("%w: requested identity is malformed", ErrUnapproved)
	}
	if !digestRE.MatchString(expectedVerifier) {
		return fmt.Errorf("%w: requested identity is malformed", ErrUnapproved)
	}
	inputs := append([]string{checkpoint.InputDigest}, checkpoint.InputLineage...)
	verifierMatches := 0
	for _, migration := range ledger.VerifierMigrations {
		if migration.Module == checkpoint.Module && migration.Package == checkpoint.Package &&
			migration.ExecutionRevision == checkpoint.ExecutionRevision && migration.GremlinsVersion == checkpoint.Gremlins &&
			migration.GremlinsVerifierSHA256 == expectedVerifier && "sha256:"+migration.ReportSHA256 == checkpoint.ReportDigest &&
			contains(inputs, migration.GateInputDigest) {
			verifierMatches++
		}
	}
	if verifierMatches != 1 || (checkpoint.VerifierDigest != "" && checkpoint.VerifierDigest != expectedVerifier) {
		return fmt.Errorf("%w: legacy verifier identity is not uniquely approved", ErrUnapproved)
	}
	if checkpoint.InputDigest == currentInput {
		return nil
	}
	inputMatches := 0
	for _, migration := range ledger.Entries {
		verifier := migration.GremlinsVerifierSHA256
		if verifier == "" {
			verifier = expectedVerifier
		}
		if migration.Module == checkpoint.Module && migration.Package == checkpoint.Package &&
			migration.ExecutionRevision == checkpoint.ExecutionRevision && migration.GateInputDigest == checkpoint.InputDigest &&
			migration.ReplacementInputDigest == currentInput && migration.GremlinsVersion == checkpoint.Gremlins &&
			verifier == expectedVerifier && "sha256:"+migration.ReportSHA256 == checkpoint.ReportDigest {
			inputMatches++
		}
	}
	if inputMatches != 1 {
		return fmt.Errorf("%w: replacement input identity is not uniquely approved", ErrUnapproved)
	}
	return nil
}

func contains(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}

func verifierMigrationIdentity(value VerifierMigration) string {
	return strings.Join([]string{value.Module, value.Package, value.ExecutionRevision, value.GateInputDigest, value.ReportSHA256}, "\x00")
}

func inputMigrationIdentity(value InputMigration) string {
	return strings.Join([]string{value.Module, value.Package, value.ExecutionRevision, value.GateInputDigest, value.ReplacementInputDigest, value.ReportSHA256}, "\x00")
}
