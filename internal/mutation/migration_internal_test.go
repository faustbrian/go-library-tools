package mutation

import (
	"errors"
	"strings"
	"testing"
)

func TestParseMigrationLedgerRejectsInputFailures(t *testing.T) {
	if _, err := ParseMigrationLedger(failingReader{}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("ParseMigrationLedger(read failure) error = %v", err)
	}
	if _, err := ParseMigrationLedger(strings.NewReader(strings.Repeat("x", maximumMigrationLedgerSize+1))); !errors.Is(err, ErrInvalid) {
		t.Fatalf("ParseMigrationLedger(oversized) error = %v", err)
	}
	if _, err := ParseMigrationLedger(strings.NewReader("{")); !errors.Is(err, ErrInvalid) {
		t.Fatalf("ParseMigrationLedger(malformed) error = %v", err)
	}
	if _, err := ParseMigrationLedger(strings.NewReader(`{"schema_version":2,"reason":"invalid","verifier_migration_review":{},"verifier_migrations":[],"entries":[]}`)); !errors.Is(err, ErrInvalid) {
		t.Fatalf("ParseMigrationLedger(invalid policy) error = %v", err)
	}
}

func TestMigrationLedgerValidationRejectsMalformedPolicy(t *testing.T) {
	tests := map[string]func(*MigrationLedger){
		"schema":         func(value *MigrationLedger) { value.SchemaVersion = 2 },
		"ledger reason":  func(value *MigrationLedger) { value.Reason = "short" },
		"review digest":  func(value *MigrationLedger) { value.VerifierMigrationReview.GremlinsVerifierSHA256 = "bad" },
		"review reason":  func(value *MigrationLedger) { value.VerifierMigrationReview.Reason = "short" },
		"review date":    func(value *MigrationLedger) { value.VerifierMigrationReview.ReviewedAt = "today" },
		"verifier entry": func(value *MigrationLedger) { value.VerifierMigrations[0].ExecutionRevision = "bad" },
		"input entry":    func(value *MigrationLedger) { value.Entries[0].GateInputDigest = "bad" },
		"duplicate verifier": func(value *MigrationLedger) {
			value.VerifierMigrations = append(value.VerifierMigrations, value.VerifierMigrations[0])
		},
		"duplicate input": func(value *MigrationLedger) { value.Entries = append(value.Entries, value.Entries[0]) },
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			ledger := validMigrationLedger()
			mutate(&ledger)
			if err := ledger.validate(); !errors.Is(err, ErrInvalid) {
				t.Fatalf("validate() error = %v, want ErrInvalid", err)
			}
		})
	}
}

func TestMigrationEntryValidation(t *testing.T) {
	verifier := strings.Repeat("c", 64)
	validVerifier := validMigrationLedger().VerifierMigrations[0]
	if err := validVerifier.validate(verifier); err != nil {
		t.Fatal(err)
	}
	validVerifier.Package = "../outside"
	if err := validVerifier.validate(verifier); err == nil {
		t.Fatal("VerifierMigration.validate() accepted unsafe package")
	}
	validInput := validMigrationLedger().Entries[0]
	validInput.GremlinsVerifierSHA256 = ""
	if err := validInput.validate(verifier); err != nil {
		t.Fatal(err)
	}
	validInput.ReplacementInputDigest = "bad"
	if err := validInput.validate(verifier); err == nil {
		t.Fatal("InputMigration.validate() accepted malformed replacement")
	}
	validInput.ReplacementInputDigest = strings.Repeat("b", 64)
	validInput.MigrationReason = "invalid\nreason"
	if err := validInput.validate(verifier); err == nil {
		t.Fatal("InputMigration.validate() accepted malformed reason")
	}
	if err := detailedReason("reason", strings.Repeat("valid ", 10)); err != nil {
		t.Fatal(err)
	}
}

func TestMigrationApprovalRejectsMismatches(t *testing.T) {
	ledger := validMigrationLedger()
	checkpoint := validMigrationCheckpoint()
	verifier := strings.Repeat("c", 64)
	newInput := strings.Repeat("b", 64)
	tests := map[string]struct {
		ledger     MigrationLedger
		checkpoint Checkpoint
		input      string
		verifier   string
	}{
		"malformed request":            {ledger: ledger, checkpoint: checkpoint, input: "bad", verifier: verifier},
		"missing verifier approval":    {ledger: MigrationLedger{}, checkpoint: checkpoint, input: newInput, verifier: verifier},
		"checkpoint verifier mismatch": {ledger: ledger, checkpoint: withVerifier(checkpoint, strings.Repeat("f", 64)), input: newInput, verifier: verifier},
		"duplicate verifier approval":  {ledger: withDuplicateVerifier(ledger), checkpoint: checkpoint, input: newInput, verifier: verifier},
		"missing input approval":       {ledger: withoutInputEntries(ledger), checkpoint: checkpoint, input: newInput, verifier: verifier},
		"duplicate input approval":     {ledger: withDuplicateInput(ledger), checkpoint: checkpoint, input: newInput, verifier: verifier},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if err := test.ledger.Approve(test.checkpoint, test.input, test.verifier); !errors.Is(err, ErrUnapproved) {
				t.Fatalf("Approve() error = %v, want ErrUnapproved", err)
			}
		})
	}
}

func TestMigrationApprovalUsesInputLineageAndDefaultVerifier(t *testing.T) {
	ledger := validMigrationLedger()
	ledger.Entries[0].GremlinsVerifierSHA256 = ""
	checkpoint := validMigrationCheckpoint()
	checkpoint.InputDigest = strings.Repeat("f", 64)
	checkpoint.InputLineage = []string{strings.Repeat("a", 64)}
	ledger.Entries[0].GateInputDigest = checkpoint.InputDigest
	if err := ledger.Approve(checkpoint, strings.Repeat("b", 64), strings.Repeat("c", 64)); err != nil {
		t.Fatalf("Approve() error = %v", err)
	}
	if contains([]string{"a"}, "b") {
		t.Fatal("contains() = true for absent value")
	}
}

func validMigrationLedger() MigrationLedger {
	verifier := strings.Repeat("c", 64)
	return MigrationLedger{
		SchemaVersion: 3,
		Reason:        "Exact identities require an explicit and detailed semantic migration review.",
		VerifierMigrationReview: VerifierMigrationReview{
			GremlinsVerifierSHA256: verifier,
			Reason:                 "The exact complete verifier and operator contract was independently reviewed.",
			ReviewedAt:             "2026-08-20",
		},
		VerifierMigrations: []VerifierMigration{{
			ExecutionRevision: strings.Repeat("e", 40), GateInputDigest: strings.Repeat("a", 64),
			GremlinsVerifierSHA256: verifier, GremlinsVersion: "v0.6.0", Module: ".", Package: ".",
			ReportSHA256: strings.Repeat("d", 64),
		}},
		Entries: []InputMigration{{
			ExecutionRevision: strings.Repeat("e", 40), GateInputDigest: strings.Repeat("a", 64),
			ReplacementInputDigest: strings.Repeat("b", 64), GremlinsVerifierSHA256: verifier,
			GremlinsVersion: "v0.6.0", Module: ".", Package: ".", ReportSHA256: strings.Repeat("d", 64),
		}},
	}
}

func validMigrationCheckpoint() Checkpoint {
	return Checkpoint{
		Module: ".", Package: ".", ExecutionRevision: strings.Repeat("e", 40),
		InputDigest: strings.Repeat("a", 64), Gremlins: "v0.6.0",
		ReportDigest: "sha256:" + strings.Repeat("d", 64),
	}
}

func withVerifier(value Checkpoint, verifier string) Checkpoint {
	value.VerifierDigest = verifier
	return value
}
func withDuplicateVerifier(value MigrationLedger) MigrationLedger {
	value.VerifierMigrations = append(value.VerifierMigrations, value.VerifierMigrations[0])
	return value
}
func withoutInputEntries(value MigrationLedger) MigrationLedger { value.Entries = nil; return value }
func withDuplicateInput(value MigrationLedger) MigrationLedger {
	value.Entries = append(value.Entries, value.Entries[0])
	return value
}
