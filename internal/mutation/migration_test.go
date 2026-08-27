package mutation_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/faustbrian/go-library-tools/internal/mutation"
)

func TestMigrationLedgerApprovesExactSemanticReplacement(t *testing.T) {
	oldInput := strings.Repeat("a", 64)
	newInput := strings.Repeat("b", 64)
	verifier := strings.Repeat("c", 64)
	report := strings.Repeat("d", 64)
	ledger, err := mutation.ParseMigrationLedger(strings.NewReader(migrationLedgerJSON(oldInput, newInput, verifier, report)))
	if err != nil {
		t.Fatalf("ParseMigrationLedger() error = %v", err)
	}
	checkpoint := mutation.Checkpoint{
		Module: ".", Package: ".", InputDigest: oldInput,
		ExecutionRevision: strings.Repeat("e", 40), Gremlins: "v0.6.0",
		ReportDigest: "sha256:" + report,
	}
	if err := ledger.Approve(checkpoint, newInput, verifier); err != nil {
		t.Fatalf("Approve() error = %v", err)
	}
	err = ledger.Approve(checkpoint, strings.Repeat("f", 64), verifier)
	if !errors.Is(err, mutation.ErrUnapproved) {
		t.Fatalf("Approve(changed input) error = %v, want ErrUnapproved", err)
	}
	if err == nil || !strings.Contains(err.Error(), "requested replacement "+strings.Repeat("f", 64)) {
		t.Fatalf("Approve(changed input) error = %v, want requested replacement digest", err)
	}
}

func TestMigrationLedgerApprovesUnchangedInputAfterVerifierReview(t *testing.T) {
	input := strings.Repeat("a", 64)
	verifier := strings.Repeat("c", 64)
	report := strings.Repeat("d", 64)
	ledger, err := mutation.ParseMigrationLedger(strings.NewReader(migrationLedgerJSON(input, input, verifier, report)))
	if err != nil {
		t.Fatal(err)
	}
	checkpoint := mutation.Checkpoint{
		Module: ".", Package: ".", InputDigest: input,
		ExecutionRevision: strings.Repeat("e", 40), Gremlins: "v0.6.0",
		ReportDigest: "sha256:" + report,
	}
	if err := ledger.Approve(checkpoint, input, verifier); err != nil {
		t.Fatalf("Approve() error = %v", err)
	}
}

func migrationLedgerJSON(oldInput, newInput, verifier, report string) string {
	revision := strings.Repeat("e", 40)
	return `{
  "schema_version": 3,
  "reason": "Exact content identities are retained only through explicitly reviewed semantic migrations.",
  "verifier_migration_review": {
    "gremlins_verifier_sha256": "` + verifier + `",
    "reason": "The listed checkpoints used the complete pinned mutation verifier and operator contract.",
    "reviewed_at": "2026-08-20"
  },
  "verifier_migrations": [{
    "execution_revision": "` + revision + `",
    "gate_input_digest": "` + oldInput + `",
    "gremlins_verifier_sha256": "` + verifier + `",
    "gremlins_version": "v0.6.0",
    "module": ".",
    "package": ".",
    "report_sha256": "` + report + `"
  }],
  "entries": [{
    "execution_revision": "` + revision + `",
    "gate_input_digest": "` + oldInput + `",
    "replacement_gate_input_digest": "` + newInput + `",
    "gremlins_version": "v0.6.0",
    "gremlins_verifier_sha256": "` + verifier + `",
    "module": ".",
    "package": ".",
    "report_sha256": "` + report + `"
  }]
}`
}
