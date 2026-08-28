package mutation

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestParseMigrationLedgerRejectsInputFailures(t *testing.T) {
	if _, err := ParseMigrationLedger(failingReader{}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("ParseMigrationLedger(read failure) error = %v", err)
	}
	if _, err := ParseMigrationLedger(strings.NewReader(strings.Repeat("x", MaximumMigrationLedgerSize+1))); !errors.Is(err, ErrInvalid) {
		t.Fatalf("ParseMigrationLedger(oversized) error = %v", err)
	}
	if _, err := ParseMigrationLedger(strings.NewReader("{")); !errors.Is(err, ErrInvalid) {
		t.Fatalf("ParseMigrationLedger(malformed) error = %v", err)
	}
	if _, err := ParseMigrationLedger(strings.NewReader(`{"schema_version":2,"reason":"invalid","verifier_migration_review":{},"verifier_migrations":[],"entries":[]}`)); !errors.Is(err, ErrInvalid) {
		t.Fatalf("ParseMigrationLedger(invalid policy) error = %v", err)
	}
}

func TestParseMigrationLedgerAcceptsExactSizeLimit(t *testing.T) {
	ledger := validMigrationLedgerJSON(t)
	padded := ledger + strings.Repeat(" ", MaximumMigrationLedgerSize-len(ledger))
	if _, err := ParseMigrationLedger(strings.NewReader(padded)); err != nil {
		t.Fatalf("ParseMigrationLedger() error = %v", err)
	}
}

func TestParseMigrationLedgerRequiresFinalByteAtSizeLimit(t *testing.T) {
	ledger := validMigrationLedgerJSON(t)
	prefix := ledger[:len(ledger)-1]
	input := prefix + strings.Repeat(" ", MaximumMigrationLedgerSize-len(prefix)-1) + `}`
	if _, err := ParseMigrationLedger(strings.NewReader(input)); err != nil {
		t.Fatal(err)
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
	for name, mutate := range map[string]func(*VerifierMigration){
		"revision": func(value *VerifierMigration) { value.ExecutionRevision = "bad" },
		"input":    func(value *VerifierMigration) { value.GateInputDigest = "bad" },
		"report":   func(value *VerifierMigration) { value.ReportSHA256 = "bad" },
		"version":  func(value *VerifierMigration) { value.GremlinsVersion = "latest" },
		"verifier": func(value *VerifierMigration) { value.GremlinsVerifierSHA256 = strings.Repeat("d", 64) },
		"module":   func(value *VerifierMigration) { value.Module = "../outside" },
		"package":  func(value *VerifierMigration) { value.Package = "../outside" },
	} {
		t.Run("verifier "+name, func(t *testing.T) {
			value := validMigrationLedger().VerifierMigrations[0]
			mutate(&value)
			if err := value.validate(verifier); err == nil {
				t.Fatal("VerifierMigration.validate() accepted malformed identity")
			}
		})
	}
	for name, mutate := range map[string]func(*InputMigration){
		"revision":    func(value *InputMigration) { value.ExecutionRevision = "bad" },
		"input":       func(value *InputMigration) { value.GateInputDigest = "bad" },
		"replacement": func(value *InputMigration) { value.ReplacementInputDigest = "bad" },
		"report":      func(value *InputMigration) { value.ReportSHA256 = "bad" },
		"version":     func(value *InputMigration) { value.GremlinsVersion = "latest" },
		"verifier":    func(value *InputMigration) { value.GremlinsVerifierSHA256 = strings.Repeat("d", 64) },
		"module":      func(value *InputMigration) { value.Module = "../outside" },
		"package":     func(value *InputMigration) { value.Package = "../outside" },
	} {
		t.Run("input "+name, func(t *testing.T) {
			value := validMigrationLedger().Entries[0]
			mutate(&value)
			if err := value.validate(verifier); err == nil {
				t.Fatal("InputMigration.validate() accepted malformed identity")
			}
		})
	}
	validInput := validMigrationLedger().Entries[0]
	validInput.GremlinsVerifierSHA256 = ""
	if err := validInput.validate(verifier); err != nil {
		t.Fatal(err)
	}
	validInput.ReplacementInputDigest = strings.Repeat("b", 64)
	validInput.MigrationReason = "invalid\nreason"
	if err := validInput.validate(verifier); err == nil {
		t.Fatal("InputMigration.validate() accepted malformed reason")
	}
	if err := detailedReason("reason", strings.Repeat("valid ", 10)); err != nil {
		t.Fatal(err)
	}
	if err := detailedReason("reason", strings.Repeat("r", 40)); err != nil {
		t.Fatalf("detailedReason(40) error = %v", err)
	}
	if err := detailedReason("reason", strings.Repeat("r", 39)); err == nil {
		t.Fatal("detailedReason(39) error = nil")
	}
}

func TestInputMigrationReasonAcceptsExactMaximum(t *testing.T) {
	value := validMigrationLedger().Entries[0]
	value.MigrationReason = strings.Repeat("r", 256)
	if err := value.validate(strings.Repeat("c", 64)); err != nil {
		t.Fatal(err)
	}
	value.MigrationReason += "r"
	if err := value.validate(strings.Repeat("c", 64)); err == nil {
		t.Fatal("accepted 257-byte reason")
	}
}

func TestApproveRejectsEachMalformedRequestedIdentity(t *testing.T) {
	ledger := validMigrationLedger()
	checkpoint := validMigrationCheckpoint()
	for _, values := range [][2]string{{"bad", strings.Repeat("c", 64)}, {strings.Repeat("a", 64), "bad"}} {
		if err := ledger.Approve(checkpoint, values[0], values[1]); !errors.Is(err, ErrUnapproved) {
			t.Fatalf("Approve() = %v", err)
		}
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

func TestMigrationApprovalRequiresEveryIdentityField(t *testing.T) {
	checkpoint := validMigrationCheckpoint()
	verifier := strings.Repeat("c", 64)
	newInput := strings.Repeat("b", 64)
	for name, mutate := range map[string]func(*VerifierMigration){
		"module":    func(value *VerifierMigration) { value.Module = "other" },
		"package":   func(value *VerifierMigration) { value.Package = "other" },
		"revision":  func(value *VerifierMigration) { value.ExecutionRevision = strings.Repeat("f", 40) },
		"version":   func(value *VerifierMigration) { value.GremlinsVersion = "v0.7.0" },
		"verifier":  func(value *VerifierMigration) { value.GremlinsVerifierSHA256 = strings.Repeat("f", 64) },
		"report":    func(value *VerifierMigration) { value.ReportSHA256 = strings.Repeat("f", 64) },
		"old input": func(value *VerifierMigration) { value.GateInputDigest = strings.Repeat("f", 64) },
	} {
		t.Run("verifier "+name, func(t *testing.T) {
			ledger := validMigrationLedger()
			mutate(&ledger.VerifierMigrations[0])
			if err := ledger.Approve(checkpoint, newInput, verifier); !errors.Is(err, ErrUnapproved) {
				t.Fatalf("Approve() error = %v", err)
			}
		})
	}
	for name, mutate := range map[string]func(*InputMigration){
		"module":      func(value *InputMigration) { value.Module = "other" },
		"package":     func(value *InputMigration) { value.Package = "other" },
		"revision":    func(value *InputMigration) { value.ExecutionRevision = strings.Repeat("f", 40) },
		"old input":   func(value *InputMigration) { value.GateInputDigest = strings.Repeat("f", 64) },
		"replacement": func(value *InputMigration) { value.ReplacementInputDigest = strings.Repeat("f", 64) },
		"version":     func(value *InputMigration) { value.GremlinsVersion = "v0.7.0" },
		"verifier":    func(value *InputMigration) { value.GremlinsVerifierSHA256 = strings.Repeat("f", 64) },
		"report":      func(value *InputMigration) { value.ReportSHA256 = strings.Repeat("f", 64) },
	} {
		t.Run("input "+name, func(t *testing.T) {
			ledger := validMigrationLedger()
			mutate(&ledger.Entries[0])
			if err := ledger.Approve(checkpoint, newInput, verifier); !errors.Is(err, ErrUnapproved) {
				t.Fatalf("Approve() error = %v", err)
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

func TestMigrationLedgerApprovesBuiltInInputIdentityTransition(t *testing.T) {
	ledger := validMigrationLedger()
	checkpoint := validMigrationCheckpoint()
	currentInput := strings.Repeat("f", 64)
	legacyInput := ledger.Entries[0].ReplacementInputDigest
	verifier := ledger.VerifierMigrationReview.GremlinsVerifierSHA256
	if err := ledger.approveTransition(checkpoint, currentInput, legacyInput, verifier); err != nil {
		t.Fatalf("approveTransition() error = %v", err)
	}
	if err := ledger.approveTransition(checkpoint, currentInput, "bad", verifier); !errors.Is(err, ErrUnapproved) {
		t.Fatalf("approveTransition(malformed) error = %v, want ErrUnapproved", err)
	}
	if err := ledger.approveTransition(checkpoint, currentInput, currentInput, verifier); !errors.Is(err, ErrUnapproved) {
		t.Fatalf("approveTransition(unapproved) error = %v, want ErrUnapproved", err)
	}
	checkpoint.InputLineage = []string{checkpoint.InputDigest}
	checkpoint.InputDigest = ""
	if err := ledger.approveTransition(checkpoint, currentInput, "", verifier); !errors.Is(err, ErrUnapproved) {
		t.Fatalf("approveTransition(missing legacy input) error = %v, want ErrUnapproved", err)
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

func validMigrationLedgerJSON(t *testing.T) string {
	t.Helper()
	value := validMigrationLedger()
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return string(encoded)
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
