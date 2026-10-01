package securitydocs

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestAdmitRisksRejectsInvalidPolicyWithoutPartialRegistry(t *testing.T) {
	now := time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)
	valid := risk{ID: "SEC-1", Module: "github.com/acme/example", Severity: "low", Status: "open", Owner: "maintainers", Rationale: "risk", Mitigation: "bounded", ReviewCondition: "on change"}
	for _, test := range []struct {
		name, category string
		change         func(*risk)
	}{
		{"severity", "risk record has invalid severity or status", func(value *risk) { value.Severity = "urgent" }},
		{"status", "risk record has invalid severity or status", func(value *risk) { value.Status = "pending" }},
		{"owner", "risk record requires owner", func(value *risk) { value.Owner = "\u00a0" }},
		{"accepted evidence", "accepted risk requires evidence and expires_at", func(value *risk) {
			value.Status = "accepted"
			value.ExpiresAt = now.Add(time.Hour).Format(time.RFC3339)
		}},
		{"accepted expiry", "accepted risk requires evidence and expires_at", func(value *risk) { value.Status = "accepted"; value.Evidence = "review" }},
		{"invalid expiry", "accepted risk has invalid expires_at", func(value *risk) { value.Status = "accepted"; value.Evidence = "review"; value.ExpiresAt = "tomorrow" }},
		{"expiry equality", "accepted risk has expired", func(value *risk) {
			value.Status = "accepted"
			value.Evidence = "review"
			value.ExpiresAt = now.Format(time.RFC3339)
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			invalid := valid
			invalid.ID = "SEC-2"
			test.change(&invalid)
			registry, err := admitRisks([]risk{valid, invalid}, now)
			if registry != nil || err == nil || err.Error() != test.category {
				t.Fatalf("admitRisks() = %#v, %v, want no registry and %q", registry, err, test.category)
			}
		})
	}
	if registry, err := admitRisks([]risk{valid, valid}, now); registry != nil || err == nil || err.Error() != "duplicate risk id" {
		t.Fatalf("duplicate admission = %#v, %v", registry, err)
	}
	valid.Status, valid.Evidence, valid.ExpiresAt = "accepted", "review", now.Add(time.Second).Format(time.RFC3339)
	records := []risk{valid}
	registry, err := admitRisks(records, now)
	if err != nil || len(registry) != 1 || registry[valid.ID] != valid {
		t.Fatalf("accepted admission = %#v, %v", registry, err)
	}
	records[0].Owner = "changed by caller"
	if registry[valid.ID].Owner != "maintainers" {
		t.Fatal("caller changed admitted registry")
	}
	registry[valid.ID] = risk{}
	again, err := admitRisks([]risk{valid}, now)
	if err != nil || again[valid.ID] != valid {
		t.Fatalf("registry reused across admissions: %#v, %v", again, err)
	}
}

func TestAdmitModulesPreservesScannerVerdictAndResidualConsistency(t *testing.T) {
	var decoded matrix
	if err := json.Unmarshal(testPassingMatrix(), &decoded); err != nil {
		t.Fatal(err)
	}
	digest := strings.Repeat("a", 64)
	accepted := risk{ID: "SEC-1", Module: "github.com/acme/example", Severity: "medium", Status: "accepted", Evidence: "sha256:" + digest + ":security-matrix.json"}
	registry := map[string]risk{accepted.ID: accepted}
	for _, test := range []struct {
		name, category string
		change         func(*moduleRecord)
	}{
		{"valid pass", "", func(*moduleRecord) {}},
		{"failed scanner", "passing verdict has failed scanner", func(value *moduleRecord) { value.Scanners[0].Status = "failed" }},
		{"missing scanner", "passing verdict lacks a required scanner", func(value *moduleRecord) { value.Scanners = value.Scanners[1:] }},
		{"inapplicable scanner", "passing verdict requires passed scanners", func(value *moduleRecord) { value.Scanners[0].Status = "not-applicable" }},
		{"duplicate scanner", "module record has duplicate scanner", func(value *moduleRecord) { value.Scanners = append(value.Scanners, value.Scanners[0]) }},
		{"invalid scanner status", "module record has invalid scanner result", func(value *moduleRecord) { value.Scanners[0].Status = "pending" }},
		{"invalid scanner completion", "module record has invalid scanner completion", func(value *moduleRecord) { value.Scanners[0].CompletedAt = "tomorrow" }},
		{"invalid verdict", "module record has invalid release verdict", func(value *moduleRecord) { value.Verdict.Status = "ready" }},
		{"invalid verdict time", "module record has invalid release verdict time", func(value *moduleRecord) { value.Verdict.DecidedAt = "tomorrow" }},
		{"duplicate residual", "module record duplicates residual risk", func(value *moduleRecord) { value.Verdict.ResidualRisks = []string{"SEC-1", "SEC-1"} }},
		{"unknown residual", "module record references unknown residual risk", func(value *moduleRecord) { value.Verdict.ResidualRisks = []string{"SEC-2"} }},
		{"omitted accepted risk", "passing verdict omits an accepted risk", func(value *moduleRecord) { value.Verdict.ResidualRisks = nil }},
		{"blocked with failed scanner", "", func(value *moduleRecord) { value.Verdict.Status = "blocked"; value.Scanners[0].Status = "failed" }},
	} {
		t.Run(test.name, func(t *testing.T) {
			value := decoded.Modules[0]
			value.Scanners = append([]scanner(nil), value.Scanners...)
			value.Verdict.ResidualRisks = append([]string(nil), value.Verdict.ResidualRisks...)
			test.change(&value)
			before := registry[accepted.ID]
			err := admitModules([]moduleRecord{value}, registry, digest)
			if test.category == "" && err != nil || test.category != "" && (err == nil || err.Error() != test.category) {
				t.Fatalf("admitModules() = %v, want %q", err, test.category)
			}
			if registry[accepted.ID] != before || len(registry) != 1 {
				t.Fatal("module admission changed risk registry")
			}
		})
	}
	for _, test := range []struct{ name, severity, status, evidence, category string }{
		{"accepted high", "high", "accepted", accepted.Evidence, "passing verdict has unresolved critical or high risk"},
		{"open medium", "medium", "open", "", "residual risk must be accepted and module-owned"},
		{"different matrix", "medium", "accepted", "sha256:" + strings.Repeat("b", 64) + ":security-matrix.json", "resolved risk requires exact security matrix evidence"},
	} {
		t.Run(test.name, func(t *testing.T) {
			registered := accepted
			registered.Severity, registered.Status, registered.Evidence = test.severity, test.status, test.evidence
			if err := admitModules(decoded.Modules, map[string]risk{registered.ID: registered}, digest); err == nil || err.Error() != test.category {
				t.Fatalf("admitModules() = %v, want %q", err, test.category)
			}
		})
	}
}
